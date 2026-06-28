package immut

import (
	"cmp"
	"hash/maphash"
	"iter"
)

// Set is an immutable set.
// The set is based on the Compressed Hash-Array Mapped Prefix-tree (CHAMP) data structure.
// All operations are thread-safe.
type Set[V cmp.Ordered] Map[V, struct{}]

// MakeSet creates a [Set].
func MakeSet[V cmp.Ordered]() Set[V] {
	return Set[V]{
		hashSeed: maphash.MakeSeed(),
	}
}

// Has returns true if the set contains the value, and returns false if it does not.
func (s Set[V]) Has(value V) bool {
	_, exists := s.root.Get(value, maphash.Comparable(s.hashSeed, value), 1)
	return exists
}

// Insert inserts the value into the set.
// Returns a copy of the set without mutating the original even if an identical
// value already existed in the set.
func (m Set[V]) Insert(value V) Set[V] {
	newRoot, isNewKey := m.root.Insert(value, maphash.Comparable(m.hashSeed, value), struct{}{}, 1, m.hashSeed)

	result := Set[V]{
		hashSeed: m.hashSeed,
		size:     m.size,
		root:     newRoot,
	}

	if isNewKey {
		result.size++
	}

	return result
}

// Delete removes the value from the set.
// If the value exists, returns a copy of the set without mutating the original.
// If the value doesn't exist, returns this set without mutating.
func (m Set[V]) Delete(value V) Set[V] {
	newRoot, deleted := m.root.Delete(value, maphash.Comparable(m.hashSeed, value), 1)
	if !deleted {
		return m
	}

	return Set[V]{
		hashSeed: m.hashSeed,
		size:     m.size - 1,
		root:     newRoot,
	}
}

func (s Set[V]) Len() int {
	return s.size
}

// Values iterates the values in the set.
func (s Set[V]) Values() iter.Seq[V] {
	return func(yield func(V) bool) {
		s.root.TraverseKeys(yield)
	}
}
