package rom

import (
	"os"
	"path/filepath"
	"testing"
)

func newEmpty(t *testing.T) *ROM {
	t.Helper()
	r, err := Open("")
	if err != nil {
		t.Fatalf("Open: %v", err)
	}
	t.Cleanup(func() { _ = r.Close() })
	if err := r.Resize(Size); err != nil {
		t.Fatalf("Resize: %v", err)
	}
	return r
}

func TestWriteReadByte(t *testing.T) {
	r := newEmpty(t)
	if err := r.Write(0x100, []byte{0xAB, 0xCD, 0xEF}, true); err != nil {
		t.Fatalf("Write: %v", err)
	}
	b, err := r.ReadByteAt(0x101)
	if err != nil {
		t.Fatalf("ReadByteAt: %v", err)
	}
	if b != 0xCD {
		t.Fatalf("got 0x%X, want 0xCD", b)
	}
}

func TestUpdateChecksum_AllZero(t *testing.T) {
	r := newEmpty(t)
	if err := r.UpdateChecksum(); err != nil {
		t.Fatalf("UpdateChecksum: %v", err)
	}
	// sum = 0x1FE (initial) + 0 for all bytes -> checksum=0x01FE, inverse=0xFE01.
	// Stored LE at 0x7FDC: inverse, then checksum -> 01 FE FE 01.
	want := []byte{0x01, 0xFE, 0xFE, 0x01}
	got, err := r.Read(0x7FDC, 4)
	if err != nil {
		t.Fatalf("Read: %v", err)
	}
	for i := range want {
		if got[i] != want[i] {
			t.Fatalf("checksum bytes mismatch: got %v want %v", got, want)
		}
	}
}

func TestSaveAndReopen(t *testing.T) {
	r := newEmpty(t)
	if err := r.Write(0x10, []byte{0x42}, true); err != nil {
		t.Fatalf("Write: %v", err)
	}
	dir := t.TempDir()
	out := filepath.Join(dir, "out.sfc")
	if err := r.Save(out); err != nil {
		t.Fatalf("Save: %v", err)
	}
	info, err := os.Stat(out)
	if err != nil {
		t.Fatalf("stat output: %v", err)
	}
	if info.Size() != int64(Size) {
		t.Fatalf("output size %d, want %d", info.Size(), Size)
	}
}
