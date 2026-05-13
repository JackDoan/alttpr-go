package helpers

import (
	"sort"
)

// PatchWrite is one address->bytes record in the randomizer's patch format.
type PatchWrite struct {
	Offset int
	Data   []byte
}

// PatchMergeMinify takes two ordered lists of writes (left then right),
// decomposes them to per-byte writes (right overwriting left at any
// overlapping offset), then merges runs of consecutive offsets back into
// single multi-byte writes. Mirrors app/Helpers/array.php:patch_merge_minify.
//
// The resulting slice is sorted by offset.
func PatchMergeMinify(left, right []PatchWrite) []PatchWrite {
	bytesAt := make(map[int]byte)
	decompose := func(ws []PatchWrite) {
		for _, w := range ws {
			for i, b := range w.Data {
				bytesAt[w.Offset+i] = b
			}
		}
	}
	decompose(left)
	decompose(right)

	offsets := make([]int, 0, len(bytesAt))
	for off := range bytesAt {
		offsets = append(offsets, off)
	}
	sort.Ints(offsets)

	if len(offsets) == 0 {
		return nil
	}

	// Coalesce runs of consecutive offsets.
	out := []PatchWrite{}
	runStart := offsets[0]
	run := []byte{bytesAt[runStart]}
	for i := 1; i < len(offsets); i++ {
		if offsets[i] == offsets[i-1]+1 {
			run = append(run, bytesAt[offsets[i]])
			continue
		}
		out = append(out, PatchWrite{Offset: runStart, Data: run})
		runStart = offsets[i]
		run = []byte{bytesAt[runStart]}
	}
	out = append(out, PatchWrite{Offset: runStart, Data: run})
	return out
}
