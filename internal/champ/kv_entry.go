package champ

// kvEntry is an entry in a CHAMP node.
type kvEntry[K comparable, V any] struct {
	// key is the key of the entry.
	key K
	// value is the value of the entry.
	value V
}
