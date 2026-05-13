package randomizer

import (
	"testing"

	"github.com/JackDoan/alttpr-go/internal/boss"
	"github.com/JackDoan/alttpr-go/internal/item"
	"github.com/JackDoan/alttpr-go/internal/rom"
	"github.com/JackDoan/alttpr-go/internal/world"
)

// TestEndToEnd_RandomizeAndWrite runs the full pipeline:
//   build Standard world -> randomize -> WriteToRom -> verify byte at each
//   filled, ROM-addressable location matches the placed item's first byte.
//
// This is the strongest local validation we can do: it proves the Go pipeline
// produces a self-consistent ROM (item placements end up at the right
// addresses) without depending on PHP.
func TestEndToEnd_RandomizeAndWrite(t *testing.T) {
	ir := item.NewRegistry()
	br := boss.NewRegistry()
	w := world.NewStandard(world.DefaultStandardOptions(), ir, br)

	r, err := New([]*world.World{w}, ir, br)
	if err != nil {
		t.Fatalf("New: %v", err)
	}
	if err := r.Randomize(); err != nil {
		t.Fatalf("Randomize: %v", err)
	}

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

	// For every filled location with a non-empty address list and bytes,
	// verify the ROM contains the placed item's first byte at the first address.
	checked := 0
	skipped := 0
	for _, region := range w.AllRegions() {
		for _, l := range region.Locations.All() {
			if !l.HasItem() {
				continue
			}
			if len(l.Address) == 0 || len(l.Bytes) > 0 && l.Bytes[0] < 0 {
				skipped++
				continue
			}
			// Skip locations whose first address is sentinel (-1) — those are
			// prize/medallion locations where the first byte slot has no
			// ROM address.
			if l.Address[0] < 0 {
				skipped++
				continue
			}
			itemBytes := l.Item().GetBytes()
			if len(itemBytes) == 0 || itemBytes[0] < 0 {
				skipped++
				continue
			}
			expected := byte(itemBytes[0])
			got, err := romFile.ReadByteAt(l.Address[0])
			if err != nil {
				t.Errorf("%s: ReadByteAt(0x%X): %v", l.Name, l.Address[0], err)
				continue
			}
			if got != expected {
				t.Errorf("%s @ 0x%X: got 0x%02X, want 0x%02X (item=%s)",
					l.Name, l.Address[0], got, expected, l.Item().Name)
			}
			checked++
		}
	}
	t.Logf("first-byte slot: verified %d / skipped %d locations", checked, skipped)
	if checked < 100 {
		t.Errorf("expected 100+ checked locations, got %d", checked)
	}

	// Second pass: for locations with multiple addresses (prizes, medallions),
	// verify every non-sentinel (address, byte) pair was correctly written.
	extraChecked, extraMismatches := 0, 0
	for _, region := range w.AllRegions() {
		for _, l := range region.Locations.All() {
			if !l.HasItem() {
				continue
			}
			itemBytes := l.Item().GetBytes()
			for i, addr := range l.Address {
				if i == 0 {
					continue // covered by first-byte pass above
				}
				if addr < 0 || i >= len(itemBytes) || itemBytes[i] < 0 {
					continue
				}
				got, err := romFile.ReadByteAt(addr)
				if err != nil {
					t.Errorf("%s addr[%d]@0x%X: %v", l.Name, i, addr, err)
					continue
				}
				want := byte(itemBytes[i])
				if got != want {
					t.Errorf("%s addr[%d]@0x%X: got 0x%02X want 0x%02X (item=%s)",
						l.Name, i, addr, got, want, l.Item().Name)
					extraMismatches++
				}
				extraChecked++
			}
		}
	}
	t.Logf("multi-slot pass: verified %d additional (addr, byte) pairs", extraChecked)

	// ROM-validity checks.
	if err := romFile.UpdateChecksum(); err != nil {
		t.Fatalf("UpdateChecksum: %v", err)
	}
	ok, err := romFile.CheckMD5()
	if err != nil {
		t.Fatalf("CheckMD5: %v", err)
	}
	// MD5 won't match the base patch hash because we've modified it; just
	// ensure the call succeeds. We don't expect ok=true here.
	_ = ok

	// Verify checksum stored is consistent.
	checksumBytes, _ := romFile.Read(0x7FDC, 4)
	inv := int(checksumBytes[0]) | (int(checksumBytes[1]) << 8)
	chk := int(checksumBytes[2]) | (int(checksumBytes[3]) << 8)
	if inv^chk != 0xFFFF {
		t.Errorf("checksum inconsistent: inverse 0x%04X ^ checksum 0x%04X = 0x%04X",
			inv, chk, inv^chk)
	}

	// Playthrough check: a successful seed should yield 1+ rounds and visit
	// >=100 regions (sanity bound).
	pt := w.GetPlaythrough()
	if pt == nil {
		t.Fatal("Playthrough nil after Randomize")
	}
	if pt.LongestItemChain < 1 {
		t.Errorf("LongestItemChain = %d, want >= 1", pt.LongestItemChain)
	}
	if pt.RegionsVisited < 100 {
		t.Errorf("RegionsVisited = %d, want >= 100", pt.RegionsVisited)
	}
	t.Logf("playthrough: %d rounds, %d regions visited", pt.LongestItemChain, pt.RegionsVisited)
}
