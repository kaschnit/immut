package immut

import (
	"cmp"
	"hash/maphash"
	"iter"

	"github.com/kaschnit/immut/internal/champ"
)

// Map is an immutable map.
// The map is based on the Compressed Hash-Array Mapped Prefix-tree (CHAMP) data structure.
// All operations are thread-safe.
type Map[K cmp.Ordered, V any] struct {
	hashSeed maphash.Seed
	size     int
	root     champ.KVNode[K, V]
}

// MakeMap creates a [Map].
func MakeMap[K cmp.Ordered, V any]() Map[K, V] {
	return Map[K, V]{
		hashSeed: maphash.MakeSeed(),
	}
}

// Get gets the value associated with key.
// If the key exists, returns the value and true.
// If the key doesn't exist, returns the zero value and false.
func (m Map[K, V]) Get(key K) (V, bool) {
	return m.root.Get(key, m.hashSeed)
}

// Set associates the value with the key.
// Returns a copy of the map without mutating the original even if an identical
// key/value pair already existed in the map.
func (m Map[K, V]) Set(key K, value V) Map[K, V] {
	newRoot, isNewKey := m.root.Insert(key, value, m.hashSeed)

	result := Map[K, V]{
		hashSeed: m.hashSeed,
		size:     m.size,
		root:     newRoot,
	}

	if isNewKey {
		result.size++
	}

	return result
}

// Delete removes the key and its associated value.
// If the key exists, returns a copy of the map without mutating the original.
// If the key doesn't exist, returns this map without mutating.
func (m Map[K, V]) Delete(key K) Map[K, V] {
	newRoot, deleted := m.root.Delete(key, m.hashSeed)
	if !deleted {
		return m
	}

	return Map[K, V]{
		hashSeed: m.hashSeed,
		size:     m.size - 1,
		root:     newRoot,
	}
}

// Len returns the number of items in this map.
func (m Map[K, V]) Len() int {
	return m.size
}

// All iterates the key/value pairs.
func (m Map[K, V]) All() iter.Seq2[K, V] {
	return func(yield func(K, V) bool) {
		m.root.Traverse(yield)
	}
}

// Keys iterates the keys.
func (m Map[K, V]) Keys() iter.Seq[K] {
	return func(yield func(K) bool) {
		m.root.TraverseKeys(yield)
	}
}

// Values iterates the values.
func (m Map[K, V]) Values() iter.Seq[V] {
	return func(yield func(V) bool) {
		m.root.TraverseValues(yield)
	}
}
