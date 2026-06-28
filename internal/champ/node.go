package champ

import (
	"cmp"
	"hash/maphash"
)

const (
	// maxDepth is the max depth of the CHAMP.
	// This is 13 because CHAMP uses 64 bit hash, and 5 bits are used per level.
	// The 13th level uses only 4 bits since the previous levels use a total of 60 bits,
	// leaving only 4 for the 13th level.
	// Nodes at maxDepth are collision nodes (first 60 or more of the 64 bits of the hash collide).
	maxDepth = 13
)

// KVNode is a CHAMP key-value node with branching factor of 32.
//
// Nodes at max depth are treated as collision nodes. Collision nodes use the same KVNode struct
// but don't use typical CHAMP techniques; entries in collision nodes are just maintained in sorted
// order and not tracked in bitmaps. Collision nodes have no children.
type KVNode[K cmp.Ordered, V any] struct {
	// entryMap is a bitmap indicating which of the 32 entries exist.
	// A collision node always has entryMap==0.
	entryMap uint32
	// childMap is a bitmap indicating which of the 32 children exist.
	// A collision node always has childMap==0.
	childMap uint32
	// entries are the key-value pair data items this node contains.
	// A collision node keeps entries sorted without tracking in entryMap and always
	// has len(entries)>0.
	entries []kvEntry[K, V]
	// children are the child nodes of this node.
	// A collision node always has len(children)==0.
	children []*KVNode[K, V]
}

// Get gets the value associated with the key.
// Returns the value if found, and bool indicating whether it was found.
func (node *KVNode[K, V]) Get(key K, hashSeed maphash.Seed) (V, bool) {
	return node.get(key, maphash.Comparable(hashSeed, key), 1)
}

func (node *KVNode[K, V]) get(key K, hash uint64, depth int) (V, bool) {
	if depth == maxDepth {
		// Max depth, must linear search to handle full hash collisions.
		// Iterate all entries and find matching key if there is any.
		for _, entry := range node.entries {
			if entry.key == key {
				return entry.value, true
			}
		}

		var zero V
		return zero, false
	}

	bitmapPos := getBitmapPos(hash, depth)

	// Check for entry.
	if node.entryMap&bitmapPos != 0 {
		// Get index of entry.
		index := getIndex(node.entryMap, bitmapPos)

		// Check if key matches to ensure no hash collision.
		if entry := node.entries[index]; entry.key == key {
			return entry.value, true
		}

		// Key did not match (hash collision).
		var zero V
		return zero, false
	}

	// Check for child.
	if node.childMap&bitmapPos != 0 {
		// Get index of child.
		index := getIndex(node.childMap, bitmapPos)
		// Recurse on child.
		return node.children[index].get(key, hash, depth+1)
	}

	// Key did not match (no entry or children).
	var zero V
	return zero, false
}

// Insert associates the value with the key.
// Returns the new node that replaces this node after insertion, and bool that is true if
// the key is new (key did not already exist in the CHAMP).
func (node *KVNode[K, V]) Insert(key K, value V, hashSeed maphash.Seed) (KVNode[K, V], bool) {
	return node.insert(key, value, maphash.Comparable(hashSeed, key), hashSeed, 1)
}

func (node *KVNode[K, V]) insert(key K, value V, hash uint64, hashSeed maphash.Seed, depth int) (KVNode[K, V], bool) {
	if depth == maxDepth {
		// Max depth, do not use hashing here to avoid hash collision.
		// Check for existing entry to overwrite.
		for i := range node.entries {
			if node.entries[i].key == key {
				result := KVNode[K, V]{
					entryMap: node.entryMap,
					childMap: node.childMap,
					children: node.children,
					entries:  make([]kvEntry[K, V], len(node.entries)),
				}
				copy(result.entries, node.entries)
				result.entries[i].value = value
				return result, false
			}
		}

		// Add new entry.
		entryIndex := len(node.entries)
		for i := range node.entries {
			if node.entries[i].key > key {
				entryIndex = i
				break
			}
		}

		result := KVNode[K, V]{
			entryMap: node.entryMap,
			childMap: node.childMap,
			children: node.children,
			entries:  make([]kvEntry[K, V], len(node.entries)+1),
		}
		copy(result.entries[:entryIndex], node.entries[:entryIndex])
		result.entries[entryIndex] = kvEntry[K, V]{key: key, value: value}
		copy(result.entries[entryIndex+1:], node.entries[entryIndex:])

		return result, true
	}

	bitmapPos := getBitmapPos(hash, depth)

	// Check for entry.
	if node.entryMap&bitmapPos != 0 {
		// Entry exists.
		// Get index of entry.
		entryIndex := getIndex(node.entryMap, bitmapPos)

		// Check if key matches to decide whether to overwrite or handle
		// hash collision and propagate to child.
		existingEntry := node.entries[entryIndex]
		if existingEntry.key == key {
			// Overwrite entry (not a collision).
			result := KVNode[K, V]{
				entryMap: node.entryMap,
				childMap: node.childMap,
				children: node.children,
				entries:  make([]kvEntry[K, V], len(node.entries)),
			}
			copy(result.entries, node.entries)
			result.entries[entryIndex].value = value
			return result, false
		}

		// Split existing entry into a child branch
		child := KVNode[K, V]{}
		child, _ = child.insert(existingEntry.key, existingEntry.value, maphash.Comparable(hashSeed, existingEntry.key), hashSeed, depth+1)
		child, _ = child.insert(key, value, hash, hashSeed, depth+1)

		newChildMap := node.childMap | bitmapPos
		childIndex := getIndex(newChildMap, bitmapPos)

		result := KVNode[K, V]{
			entryMap: node.entryMap & ^bitmapPos,
			childMap: newChildMap,
			entries:  make([]kvEntry[K, V], len(node.entries)-1),
			children: make([]*KVNode[K, V], len(node.children)+1),
		}
		copy(result.entries[:entryIndex], node.entries[:entryIndex])
		copy(result.entries[entryIndex:], node.entries[entryIndex+1:])

		copy(result.children[:childIndex], node.children[:childIndex])
		copy(result.children[childIndex+1:], node.children[childIndex:])
		result.children[childIndex] = &child

		return result, true
	}

	// Check for child.
	// Either get existing one, or create a new one.
	if node.childMap&bitmapPos != 0 {
		// Child exists.
		childIndex := getIndex(node.childMap, bitmapPos)
		child, isNewKey := node.children[childIndex].insert(key, value, hash, hashSeed, depth+1)

		// Mutating an existing child path: structural sharing for entries,
		// and we only allocate a new children slice of the exact same size to update the pointer.
		result := KVNode[K, V]{
			entryMap: node.entryMap,
			childMap: node.childMap,
			entries:  node.entries,
			children: make([]*KVNode[K, V], len(node.children)),
		}
		copy(result.children, node.children)
		result.children[childIndex] = &child

		return result, isNewKey
	}

	// Child does not exist.
	// Create entry.
	// Slot is completely empty: entries length increases by 1, children is shared.
	entryIndex := getIndex(node.entryMap, bitmapPos)
	newEntryMap := node.entryMap | bitmapPos

	result := KVNode[K, V]{
		entryMap: newEntryMap,
		childMap: node.childMap,
		entries:  make([]kvEntry[K, V], len(node.entries)+1),
		children: node.children,
	}
	copy(result.entries[:entryIndex], node.entries[:entryIndex])
	result.entries[entryIndex] = kvEntry[K, V]{key: key, value: value}
	copy(result.entries[entryIndex+1:], node.entries[entryIndex:])

	return result, true
}

// Delete deletes the key and its associated value from the map.
// Returns the new node that replaces this node after insertion, and bool that is true if
// anything was actually deleted.
func (node *KVNode[K, V]) Delete(key K, hashSeed maphash.Seed) (KVNode[K, V], bool) {
	return node.delete(key, maphash.Comparable(hashSeed, key), 1)
}

func (node *KVNode[K, V]) delete(key K, hash uint64, depth int) (KVNode[K, V], bool) {
	if depth == maxDepth {
		for i := range node.entries {
			if node.entries[i].key == key {
				// Exact size allocation for removing an element
				if len(node.entries) == 1 {
					return KVNode[K, V]{}, false // Terminal node is now completely empty
				}
				newEntries := make([]kvEntry[K, V], len(node.entries)-1)
				copy(newEntries[:i], node.entries[:i])
				copy(newEntries[i:], node.entries[i+1:])

				return KVNode[K, V]{
					entryMap: node.entryMap,
					childMap: node.childMap,
					entries:  newEntries,
					children: node.children,
				}, true
			}
		}

		return KVNode[K, V]{}, false
	}

	bitmapPos := getBitmapPos(hash, depth)

	if node.entryMap&bitmapPos != 0 {
		entryIndex := getIndex(node.entryMap, bitmapPos)
		if node.entries[entryIndex].key != key {
			return KVNode[K, V]{}, false // Hash collision but different key: item doesn't exist
		}

		// Exact size allocation: remove 1 entry from this node
		newEntries := make([]kvEntry[K, V], len(node.entries)-1)
		copy(newEntries[:entryIndex], node.entries[:entryIndex])
		copy(newEntries[entryIndex:], node.entries[entryIndex+1:])

		return KVNode[K, V]{
			entryMap: node.entryMap & ^bitmapPos,
			childMap: node.childMap,
			entries:  newEntries,
			children: node.children,
		}, true
	}

	if node.childMap&bitmapPos != 0 {
		childIndex := getIndex(node.childMap, bitmapPos)

		child, deleted := node.children[childIndex].delete(key, hash, depth+1)
		if !deleted {
			return KVNode[K, V]{}, false
		}

		if len(child.entries) == 1 && len(child.children) == 0 {
			// Child only has one item remaining, can be collapsed.
			// Child can never reach 0 entries, because it gets created at 2 and collapsed at 1.
			// Check the entries length, not the entryMap 1 count, to covers full hash collisions.
			entry := child.entries[0]

			// Remove the child pointer
			newChildren := make([]*KVNode[K, V], len(node.children)-1)
			copy(newChildren[:childIndex], node.children[:childIndex])
			copy(newChildren[childIndex:], node.children[childIndex+1:])

			// Add the collapsed leaf entry
			newEntryMap := node.entryMap | bitmapPos
			entryIndex := getIndex(newEntryMap, bitmapPos)

			newEntries := make([]kvEntry[K, V], len(node.entries)+1)
			copy(newEntries[:entryIndex], node.entries[:entryIndex])
			newEntries[entryIndex] = entry
			copy(newEntries[entryIndex+1:], node.entries[entryIndex:])

			return KVNode[K, V]{
				entryMap: newEntryMap,
				childMap: node.childMap & ^bitmapPos,
				entries:  newEntries,
				children: newChildren,
			}, true
		}

		// Standard child update: Structural sharing for entries array,
		// allocate an exact matching slice size only for the updated child pointer array.
		newChildren := make([]*KVNode[K, V], len(node.children))
		copy(newChildren, node.children)
		newChildren[childIndex] = &child

		return KVNode[K, V]{
			entryMap: node.entryMap,
			childMap: node.childMap,
			entries:  node.entries,
			children: newChildren,
		}, true
	}

	// Nothing to delete
	return KVNode[K, V]{}, false
}

// Traverse traverses all key/value pairs and calls the yield() callback on each pair.
// Exits early if yield(key, value) returns false.
func (node *KVNode[K, V]) Traverse(yield func(K, V) bool) bool {
	// If we are at a terminal collision node, ignore bitmaps and yield linearly
	if node.entryMap == 0 && node.childMap == 0 && len(node.entries) > 0 {
		for _, entry := range node.entries {
			if !yield(entry.key, entry.value) {
				return false
			}
		}
		return true
	}

	var entryIdx, childIdx int

	for i := range 64 {
		bit := uint32(1) << i
		if node.entryMap&bit != 0 {
			if !yield(node.entries[entryIdx].key, node.entries[entryIdx].value) {
				return false
			}
			entryIdx++
		}

		if node.childMap&bit != 0 {
			if !node.children[childIdx].Traverse(yield) {
				return false
			}
			childIdx++
		}
	}

	return true
}

// Traverse traverses all keys and calls the yield() callback on each key.
// Exits early if yield(key) returns false.
func (node *KVNode[K, V]) TraverseKeys(yield func(K) bool) bool {
	// If we are at a terminal collision node, ignore bitmaps and yield linearly
	if node.entryMap == 0 && node.childMap == 0 && len(node.entries) > 0 {
		for _, entry := range node.entries {
			if !yield(entry.key) {
				return false
			}
		}
		return true
	}

	var entryIdx, childIdx int

	for i := range 64 {
		bit := uint32(1) << i
		if node.entryMap&bit != 0 {
			if !yield(node.entries[entryIdx].key) {
				return false
			}
			entryIdx++
		}

		if node.childMap&bit != 0 {
			if !node.children[childIdx].TraverseKeys(yield) {
				return false
			}
			childIdx++
		}
	}

	return true
}

// TraverseValues traverses all values and calls the yield() callback on each value.
// Exits early if yield(value) returns false.
func (node *KVNode[K, V]) TraverseValues(yield func(V) bool) bool {
	// If we are at a terminal collision node, ignore bitmaps and yield linearly
	if node.entryMap == 0 && node.childMap == 0 && len(node.entries) > 0 {
		for _, entry := range node.entries {
			if !yield(entry.value) {
				return false
			}
		}
		return true
	}

	var entryIdx, childIdx int

	for i := range 64 {
		bit := uint32(1) << i
		if node.entryMap&bit != 0 {
			if !yield(node.entries[entryIdx].value) {
				return false
			}
			entryIdx++
		}

		if node.childMap&bit != 0 {
			if !node.children[childIdx].TraverseValues(yield) {
				return false
			}
			childIdx++
		}
	}

	return true
}
