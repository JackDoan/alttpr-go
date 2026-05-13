package world

import (
	"os"
	"path/filepath"
	"testing"

	"github.com/JackDoan/alttpr-go/internal/boss"
	"github.com/JackDoan/alttpr-go/internal/item"
	"github.com/JackDoan/alttpr-go/internal/rom"
)

// TestWriteToRom_Structural builds a Standard world, fills it (via direct
// item placement, no randomizer), writes to ROM, and verifies that the
// item bytes at each location's ROM address match the placed item.
func TestWriteToRom_Structural(t *testing.T) {
	ir := item.NewRegistry()
	br := boss.NewRegistry()
	w := NewStandard(DefaultStandardOptions(), ir, br)

	// Place a known item at a known location (Eastern Palace Big Chest).
	bigChest := w.Locations().Get("Eastern Palace - Big Chest:0")
	if bigChest == nil {
		t.Fatal("Eastern Palace - Big Chest not found")
	}
	hammer, err := ir.Get("Hammer", 0)
	if err != nil {
		t.Fatalf("get Hammer: %v", err)
	}
	if err := bigChest.SetItem(hammer); err != nil {
		t.Fatalf("set item: %v", err)
	}

	// Open empty ROM, resize, write.
	r, err := rom.Open("")
	if err != nil {
		t.Fatalf("rom.Open: %v", err)
	}
	defer r.Close()
	if err := r.Resize(rom.Size); err != nil {
		t.Fatalf("Resize: %v", err)
	}

	if err := w.WriteToRom(r, ir); err != nil {
		t.Fatalf("WriteToRom: %v", err)
	}

	// Verify Hammer byte (0x09) at Eastern Palace Big Chest address (0xE97D).
	b, err := r.ReadByteAt(0xE97D)
	if err != nil {
		t.Fatalf("read: %v", err)
	}
	if b != 0x09 {
		t.Errorf("Hammer byte at 0xE97D = 0x%02X, want 0x09", b)
	}

	// Verify checksum: store + recompute should match.
	if err := r.UpdateChecksum(); err != nil {
		t.Fatalf("UpdateChecksum: %v", err)
	}
	checksumBytes, err := r.Read(0x7FDC, 4)
	if err != nil {
		t.Fatalf("read checksum: %v", err)
	}
	inverse := int(checksumBytes[0]) | (int(checksumBytes[1]) << 8)
	checksum := int(checksumBytes[2]) | (int(checksumBytes[3]) << 8)
	if inverse^checksum != 0xFFFF {
		t.Errorf("inverse ^ checksum = 0x%04X, want 0xFFFF", inverse^checksum)
	}

	// Save to disk + reload to verify file integrity.
	dir := t.TempDir()
	out := filepath.Join(dir, "test.sfc")
	if err := r.Save(out); err != nil {
		t.Fatalf("Save: %v", err)
	}
	info, err := os.Stat(out)
	if err != nil {
		t.Fatalf("stat: %v", err)
	}
	if info.Size() != int64(rom.Size) {
		t.Errorf("ROM size = %d, want %d", info.Size(), rom.Size)
	}
	t.Logf("wrote valid SFC: %d bytes, checksum 0x%04X", info.Size(), checksum)
}
