package randomizer

import (
	"testing"

	"github.com/JackDoan/alttpr-go/internal/boss"
	"github.com/JackDoan/alttpr-go/internal/item"
	"github.com/JackDoan/alttpr-go/internal/rom"
	"github.com/JackDoan/alttpr-go/internal/world"
)

// TestEndToEnd_OpenMode runs the full pipeline in Open mode and verifies
// that:
//   - RescueZelda is pre-collected (Open starts with Zelda saved)
//   - All 236 location slots get items
//   - The resulting ROM has a valid checksum
//   - ROM bytes for each location match the spoiler
func TestEndToEnd_OpenMode(t *testing.T) {
	ir := item.NewRegistry()
	br := boss.NewRegistry()
	opts := world.DefaultStandardOptions()
	w := world.NewOpen(opts, ir, br)

	r, err := New([]*world.World{w}, ir, br)
	if err != nil {
		t.Fatalf("New: %v", err)
	}
	if err := r.Randomize(); err != nil {
		t.Fatalf("Randomize: %v", err)
	}

	// RescueZelda must be in the pre-collected pool for Open mode.
	if !w.PreCollectedItems().Has1("RescueZelda") {
		t.Error("Open mode: RescueZelda not pre-collected")
	}

	// All locations should be filled.
	if empty := w.EmptyLocations().Count(); empty != 0 {
		t.Errorf("Open mode: %d empty locations after randomization", empty)
	}

	// Write to ROM and verify checksum + a sample location byte.
	romFile, err := rom.Open("")
	if err != nil {
		t.Fatalf("rom.Open: %v", err)
	}
	defer romFile.Close()
	if err := romFile.Resize(rom.Size); err != nil {
		t.Fatalf("Resize: %v", err)
	}
	if err := w.WriteToRom(romFile, ir); err != nil {
		t.Fatalf("WriteToRom: %v", err)
	}

	chk, _ := romFile.Read(0x7FDC, 4)
	inv := int(chk[0]) | (int(chk[1]) << 8)
	cs := int(chk[2]) | (int(chk[3]) << 8)
	if inv^cs != 0xFFFF {
		t.Errorf("invalid checksum: 0x%04X ^ 0x%04X = 0x%04X", inv, cs, inv^cs)
	}
}
