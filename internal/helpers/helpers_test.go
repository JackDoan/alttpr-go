package helpers

import (
	"testing"
)

func TestPcSnesRoundTrip(t *testing.T) {
	cases := []int{0, 0x100, 0x7FFF, 0x8000, 0x10000, 0x18000, 0x1FFFFF}
	for _, pc := range cases {
		got := SnesToPc(PcToSnes(pc))
		if got != pc {
			t.Errorf("roundtrip pc=0x%X: snes=0x%X back=0x%X", pc, PcToSnes(pc), got)
		}
	}
}

func TestCountSetBits(t *testing.T) {
	for v, want := range map[int]int{0: 0, 1: 1, 2: 1, 3: 2, 7: 3, 0xFF: 8, 0x80000000: 1} {
		if got := CountSetBits(v); got != want {
			t.Errorf("CountSetBits(%d) = %d, want %d", v, got, want)
		}
	}
}

func TestHashArrayShape(t *testing.T) {
	got := HashArray(123456)
	for i, v := range got {
		if v < 0 || v > 0x1F {
			t.Errorf("HashArray[%d] = %d, out of 5-bit range", i, v)
		}
	}
}

func TestPatchMergeMinify_CoalescesAdjacent(t *testing.T) {
	in := []PatchWrite{
		{Offset: 10, Data: []byte{1}},
		{Offset: 11, Data: []byte{2}},
		{Offset: 12, Data: []byte{3}},
		{Offset: 20, Data: []byte{9}},
	}
	got := PatchMergeMinify(in, nil)
	if len(got) != 2 {
		t.Fatalf("len=%d, want 2: %+v", len(got), got)
	}
	if got[0].Offset != 10 || string(got[0].Data) != string([]byte{1, 2, 3}) {
		t.Errorf("entry 0 = %+v", got[0])
	}
	if got[1].Offset != 20 || string(got[1].Data) != string([]byte{9}) {
		t.Errorf("entry 1 = %+v", got[1])
	}
}

func TestPatchMergeMinify_RightOverwritesLeft(t *testing.T) {
	left := []PatchWrite{{Offset: 5, Data: []byte{1, 2, 3, 4}}}
	right := []PatchWrite{{Offset: 6, Data: []byte{0xFF, 0xFE}}}
	got := PatchMergeMinify(left, right)
	if len(got) != 1 {
		t.Fatalf("len=%d, want 1: %+v", len(got), got)
	}
	if got[0].Offset != 5 || string(got[0].Data) != string([]byte{1, 0xFF, 0xFE, 4}) {
		t.Errorf("got %+v", got[0])
	}
}

func TestFyShuffle_PreservesElements(t *testing.T) {
	in := []int{1, 2, 3, 4, 5, 6, 7, 8, 9, 10}
	out, err := FyShuffle(in)
	if err != nil {
		t.Fatalf("FyShuffle: %v", err)
	}
	if len(out) != len(in) {
		t.Fatalf("len changed")
	}
	sum := 0
	for _, v := range out {
		sum += v
	}
	if sum != 55 {
		t.Errorf("sum=%d, want 55", sum)
	}
}
