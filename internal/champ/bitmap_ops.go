package champ

import "math/bits"

const (
	// branchingFactorBits is the number of bits needed to represent a branching
	// factor of 32. It is 6 because we need 6 bits to represent 32 items.
	branchingFactorBits = 5
	// branchingFactorBitsMask is a mask to extract branch bits for CHAMP with branching
	// factor of 32. This is 5 bits because [champBranchingFactorBits] is 5. It is a uint64
	// because it's used as a mask against the uint64 hash.
	branchingFactorBitsMask uint64 = 0b11111
)

// getBitmapPos gets the position of the hash's item within
// a bitmap representing a node's items at the given depth.
func getBitmapPos(hash uint64, depth int) uint32 {
	// There are branchingFactorBits bits per level.
	// At depth 1, we want the first branchingFactorBits.
	// At depth 2, we want the next branchingFactorBits.
	// And so on. So multiply branchingFactorBits by depth.
	rShiftForDepth := depth * branchingFactorBits

	// Extract the bits for the given depth.
	// Shift right to get thsi depth's bits all the way at the right.
	// Apply mask to clear any bits to the left of them.
	bitsForDepth := (hash >> rShiftForDepth) & branchingFactorBitsMask

	// Get the position within a 32-bit bitmap for the number that these bits represent.
	return uint32(1) << bitsForDepth
}

// getIndex gets an index for the item in the bitmap at the given position.
func getIndex(bitmap, bitmapPos uint32) int {
	// bitmapPos is a single set bit, so subtracting 1 yields a mask of all 1 bits.
	// For example, if bitmapPos is 0b1000, bitmapPos-1 is 0b0111.
	// This can be used to mask all bits before the given position.
	mask := bitmapPos - 1

	// Extract bits lower than (to the right of) bitmapPos.
	bitsLowerThanPos := bitmap & mask

	// Count the number of lower bits.
	// This provides the index into the slice of items in the node.
	return bits.OnesCount32(bitsLowerThanPos)
}
