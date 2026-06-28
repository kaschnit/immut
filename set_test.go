package immut_test

import (
	"slices"
	"testing"

	"github.com/kaschnit/immut"
	"github.com/stretchr/testify/assert"
)

func TestSet_BasicCRUDAndImmutability(t *testing.T) {
	s1 := immut.MakeSet[string]()

	t.Run("Empty set has no items", func(t *testing.T) {
		assert.False(t, s1.Has("foo"))
		assert.Equal(t, 0, s1.Len())
	})

	s2 := s1.Insert("foo")
	t.Run("Insert an item into empty set", func(t *testing.T) {
		assert.False(t, s1.Has("foo"))
		assert.Equal(t, 0, s1.Len())
		assert.True(t, s2.Has("foo"))
		assert.Equal(t, 1, s2.Len())
	})

	s3 := s2.Insert("foo")
	t.Run("Re-insert a value in a set", func(t *testing.T) {
		assert.False(t, s1.Has("foo"))
		assert.Equal(t, 0, s1.Len())

		assert.True(t, s2.Has("foo"))
		assert.Equal(t, 1, s2.Len())

		assert.True(t, s3.Has("foo"))
		assert.Equal(t, 1, s3.Len())
	})

	s4 := s3.Delete("foo")
	t.Run("Delete an existing value in a map", func(t *testing.T) {
		assert.False(t, s1.Has("foo"))
		assert.Equal(t, 0, s1.Len())

		assert.True(t, s2.Has("foo"))
		assert.Equal(t, 1, s2.Len())

		assert.True(t, s3.Has("foo"))
		assert.Equal(t, 1, s3.Len())

		assert.False(t, s4.Has("foo"))
		assert.Equal(t, 0, s4.Len())
	})

	// 5. Deleting non-existent key returns identicalset
	s5 := s4.Delete("non-existent")
	t.Run("Delete a non-existent value in an empty map", func(t *testing.T) {
		assert.False(t, s1.Has("foo"))
		assert.Equal(t, 0, s1.Len())

		assert.True(t, s2.Has("foo"))
		assert.Equal(t, 1, s2.Len())

		assert.True(t, s3.Has("foo"))
		assert.Equal(t, 1, s3.Len())

		assert.False(t, s4.Has("foo"))
		assert.Equal(t, 0, s4.Len())

		assert.False(t, s5.Has("foo"))
		assert.Equal(t, s4, s5)
		assert.Equal(t, 0, s5.Len())
	})

	s6 := s3.Delete("non-existent")
	t.Run("Delete a non-existent value in a non-empty map", func(t *testing.T) {
		assert.False(t, s1.Has("foo"))
		assert.Equal(t, 0, s1.Len())

		assert.True(t, s2.Has("foo"))
		assert.Equal(t, 1, s2.Len())

		assert.True(t, s3.Has("foo"))
		assert.Equal(t, 1, s3.Len())

		assert.False(t, s4.Has("foo"))
		assert.Equal(t, 0, s4.Len())

		assert.Equal(t, s4, s5)
		assert.Equal(t, 0, s5.Len())

		assert.False(t, s5.Has("foo"))
		assert.Equal(t, s4, s5)
		assert.Equal(t, 0, s5.Len())

		assert.True(t, s6.Has("foo"))
		assert.Equal(t, 1, s6.Len())
	})
}

func TestSet_Iterators(t *testing.T) {
	s := immut.MakeSet[string]().
		Insert("A").
		Insert("B").
		Insert("C")
	assert.Equal(t, 3, s.Len())

	t.Run("Values iterates the values", func(t *testing.T) {
		values := slices.Sorted(s.Values())
		assert.Len(t, values, 3)

		expectedValues := []string{"A", "B", "C"}
		for i := range values {
			assert.Equal(t, expectedValues[i], values[i])
		}
	})

	t.Run("Early termination from Values iterator", func(t *testing.T) {
		breakCount := 0
		for range s.Values() {
			breakCount++
			break
		}
		assert.Equal(t, 1, breakCount, "Iterator yield logic ignored early-termination signal")
	})
}
