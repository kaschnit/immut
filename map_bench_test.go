package immut_test

import (
	"fmt"
	"math/rand/v2"
	"testing"

	"github.com/kaschnit/immut"
)

// Helper to generate a predictable permutation of unique string keys
func generateKeys(count int) []string {
	keys := make([]string, count)
	for i := range count {
		keys[i] = fmt.Sprintf("key-structured-prefix-vector-%06d", i)
	}
	return keys
}

func BenchmarkMap_Set(b *testing.B) {
	sizes := []int{10, 100, 1_000, 10_000, 100_000, 500_000}

	for _, size := range sizes {
		keys := generateKeys(size)

		b.Run(fmt.Sprintf("immut.Map/Size-%d", size), func(b *testing.B) {
			b.ReportAllocs()
			b.ResetTimer()
			for i := 0; i < b.N; i++ {
				m := immut.MakeMap[string, int]()
				for j, key := range keys {
					m = m.Set(key, j)
				}
			}
		})
	}
}

func BenchmarkMap_Get(b *testing.B) {
	sizes := []int{10, 100, 1_000, 10_000, 100_000, 500_000}

	for _, size := range sizes {
		keys := generateKeys(size)

		// Pre-populate immut.Map map
		myMap := immut.MakeMap[string, int]()
		for j, key := range keys {
			myMap = myMap.Set(key, j)
		}

		b.Run(fmt.Sprintf("immut.Map/Size-%d", size), func(b *testing.B) {
			b.ReportAllocs()
			b.ResetTimer()
			for i := 0; i < b.N; i++ {
				key := keys[rand.IntN(size)]
				_, _ = myMap.Get(key)
			}
		})
	}
}
