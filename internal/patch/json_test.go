package patch

import (
	"strings"
	"testing"
)

func TestLoad_BasicShape(t *testing.T) {
	in := `[{"100": [1, 2, 3]}, {"200": [255]}, {"5": [0, 0, 0, 0]}]`
	got, err := Load(strings.NewReader(in))
	if err != nil {
		t.Fatalf("Load: %v", err)
	}
	want := []Entry{
		{Offset: 100, Data: []byte{1, 2, 3}},
		{Offset: 200, Data: []byte{255}},
		{Offset: 5, Data: []byte{0, 0, 0, 0}},
	}
	if len(got) != len(want) {
		t.Fatalf("len=%d, want %d", len(got), len(want))
	}
	for i := range got {
		if got[i].Offset != want[i].Offset {
			t.Errorf("entry %d offset=%d, want %d", i, got[i].Offset, want[i].Offset)
		}
		if string(got[i].Data) != string(want[i].Data) {
			t.Errorf("entry %d data=%v, want %v", i, got[i].Data, want[i].Data)
		}
	}
}

func TestLoad_PreservesOrder(t *testing.T) {
	in := `[{"10": [170]}, {"10": [187]}]`
	got, err := Load(strings.NewReader(in))
	if err != nil {
		t.Fatalf("Load: %v", err)
	}
	if len(got) != 2 || got[0].Data[0] != 170 || got[1].Data[0] != 187 {
		t.Fatalf("order not preserved: %+v", got)
	}
}

func TestLoad_RejectsOutOfRange(t *testing.T) {
	in := `[{"1": [256]}]`
	if _, err := Load(strings.NewReader(in)); err == nil {
		t.Fatal("expected out-of-range error")
	}
}
