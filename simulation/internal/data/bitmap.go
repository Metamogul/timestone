package data

// bitmap implements a bit map.
// IMPORTANT: Note that this implementation assumes that
// bitmaps are  always trimmed to not contain trailing
// zero chunks.
type bitmap []uint64

func newBitmap(index int) bitmap {
	countSlices := index/64 + 1
	positionLastSlice := index % 64

	result := make(bitmap, countSlices)
	result[len(result)-1] = 1 << positionLastSlice

	return result
}

func (b *bitmap) or(bb bitmap) {
	if len(*b) < len(bb) {
		*b = append(*b, make([]uint64, len(bb)-len(*b))...)
	}

	for chunkIndex := range len(*b) {
		if chunkIndex >= len(bb) {
			break
		}

		(*b)[chunkIndex] = (*b)[chunkIndex] | bb[chunkIndex]
	}
}

func (b *bitmap) contains(target bitmap) bool {
	if len(target) > len(*b) {
		return false
	}

	for chunkIndex := range target {
		if (*b)[chunkIndex]&target[chunkIndex] != target[chunkIndex] {
			return false
		}
	}

	return true
}

func (b *bitmap) containedIn(target bitmap) bool {
	if len(target) < len(*b) {
		return false
	}

	for chunkIndex := range *b {
		if target[chunkIndex]&(*b)[chunkIndex] != (*b)[chunkIndex] {
			return false
		}
	}

	return true
}

func (b *bitmap) equal(target bitmap) bool {
	if len(target) != len(*b) {
		return false
	}

	for chunkIndex := range target {
		if (*b)[chunkIndex] != target[chunkIndex] {
			return false
		}
	}

	return true
}
