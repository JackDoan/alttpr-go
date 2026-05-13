package world

import (
	"bytes"
	"encoding/json"
	"os"
	"testing"

	"github.com/JackDoan/alttpr-go/internal/boss"
	"github.com/JackDoan/alttpr-go/internal/item"
)

// TestParity_DialogEncoder verifies our Go Dialog encoder produces the exact
// same bytes as PHP's `Dialog->convertDialogCompressed` for a battery of
// representative input strings. Reference dump is at /tmp/parity_dialog.json
// (created by dump_parity.php).
func TestParity_DialogEncoder(t *testing.T) {
	type entry struct {
		S string `json:"s"`
		B []int  `json:"b"`
	}
	var ref []entry
	for _, path := range []string{"/tmp/parity_dialog.json", "/tmp/parity_dialog2.json"} {
		raw, err := os.ReadFile(path)
		if err != nil {
			t.Logf("skip %s: %v", path, err)
			continue
		}
		var part []entry
		if err := json.Unmarshal(raw, &part); err != nil {
			t.Fatalf("decode %s: %v", path, err)
		}
		ref = append(ref, part...)
	}
	if len(ref) == 0 {
		t.Skip("no PHP dialog reference")
	}
	for _, e := range ref {
		want := make([]byte, len(e.B))
		for i, v := range e.B {
			want[i] = byte(v)
		}
		got := ConvertDialogCompressed(e.S, true, 2046, 19)
		if !bytes.Equal(got, want) {
			diff := -1
			for i := 0; i < len(got) && i < len(want); i++ {
				if got[i] != want[i] {
					diff = i
					break
				}
			}
			t.Errorf("Dialog(%q): %d-byte Go vs %d-byte PHP, first diff at index %d", e.S, len(got), len(want), diff)
		}
	}
	t.Logf("verified %d dialog strings byte-identical to PHP", len(ref))
}

// TestParity_InitialSramDefault verifies that, given the Randomizer's default
// pre-collected items (3x BossHeartContainer, BombUpgrade10, 3x ArrowUpgrade10),
// our Go SRAM matches PHP byte-for-byte.
func TestParity_InitialSramDefault(t *testing.T) {
	wantBytes, err := os.ReadFile("/tmp/parity_sram_default.bin")
	if err != nil {
		t.Skipf("no PHP SRAM reference: %v", err)
	}

	ir := item.NewRegistry()
	br := boss.NewRegistry()
	w := NewStandard(DefaultStandardOptions(), ir, br)

	pre := item.NewCollection()
	pre.SetChecksForWorld(w.ID())
	for _, name := range []string{
		"BossHeartContainer", "BossHeartContainer", "BossHeartContainer",
		"BombUpgrade10", "ArrowUpgrade10", "ArrowUpgrade10", "ArrowUpgrade10",
	} {
		it, _ := ir.Get(name, w.ID())
		pre.Add(it)
	}

	sram := NewInitialSram()
	sram.SetStartingEquipment(pre, "randomized", false)
	got := sram.Bytes()
	if !bytes.Equal(got, wantBytes) {
		firstDiff := -1
		for i := 0; i < len(got) && i < len(wantBytes); i++ {
			if got[i] != wantBytes[i] {
				firstDiff = i
				break
			}
		}
		t.Errorf("SRAM default mismatch: %d Go bytes vs %d PHP bytes, first diff at 0x%X",
			len(got), len(wantBytes), firstDiff)
		t.Logf("around first diff (Go) : % X", got[firstDiff:min(firstDiff+16, len(got))])
		t.Logf("around first diff (PHP): % X", wantBytes[firstDiff:min(firstDiff+16, len(wantBytes))])
	}
}

// TestParity_InitialSramRich verifies a richer pre-collected items collection
// (swords, shields, mail, bottles, pendants, crystals, capacity upgrades,
// hearts, rupees) produces byte-identical SRAM.
func TestParity_InitialSramRich(t *testing.T) {
	wantBytes, err := os.ReadFile("/tmp/parity_sram_rich.bin")
	if err != nil {
		t.Skipf("no PHP SRAM reference: %v", err)
	}

	ir := item.NewRegistry()
	br := boss.NewRegistry()
	w := NewStandard(DefaultStandardOptions(), ir, br)

	pre := item.NewCollection()
	pre.SetChecksForWorld(w.ID())
	for _, name := range []string{
		"L2Sword", "RedShield", "BlueMail", "PegasusBoots", "MoonPearl",
		"Hammer", "Hookshot", "FireRod", "Bottle", "BottleWithFairy",
		"PendantOfCourage", "PendantOfWisdom",
		"Crystal1", "Crystal2", "Crystal5",
		"OneHundredRupees", "FiftyRupees",
		"BombUpgrade5", "ArrowUpgrade5", "TenBombs", "TenArrows",
		"PieceOfHeart", "PieceOfHeart", "PieceOfHeart", "PieceOfHeart",
		"HeartContainer",
	} {
		it, _ := ir.Get(name, w.ID())
		pre.Add(it)
	}

	sram := NewInitialSram()
	sram.SetStartingEquipment(pre, "randomized", false)
	got := sram.Bytes()
	if !bytes.Equal(got, wantBytes) {
		firstDiff := -1
		for i := 0; i < len(got) && i < len(wantBytes); i++ {
			if got[i] != wantBytes[i] {
				firstDiff = i
				break
			}
		}
		// Find the differing offsets across the whole buffer.
		diffs := []int{}
		for i := 0; i < len(got) && i < len(wantBytes); i++ {
			if got[i] != wantBytes[i] {
				diffs = append(diffs, i)
			}
		}
		t.Errorf("SRAM rich mismatch: %d Go bytes vs %d PHP bytes, %d differing bytes; first diff at 0x%X (Go=0x%02X PHP=0x%02X)",
			len(got), len(wantBytes), len(diffs), firstDiff, got[firstDiff], wantBytes[firstDiff])
		if len(diffs) <= 30 {
			for _, off := range diffs {
				t.Logf("  diff at 0x%03X: Go=0x%02X PHP=0x%02X", off, got[off], wantBytes[off])
			}
		}
	}
}
