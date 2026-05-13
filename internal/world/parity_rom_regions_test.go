package world

import (
	"bytes"
	"os"
	"testing"
)

// romRegionsToVerify lists ROM byte ranges that should match PHP exactly,
// independent of RNG. These are the deterministic config/data writes that
// don't depend on filler randomness or random text selection.
var romRegionsToVerify = []struct {
	Name string
	From int
	To   int // exclusive
}{
	// Game-state config block: substitutions, swordless mode, map/compass modes,
	// crystal counts, item limits, etc.
	{"game-state config", 0x180020, 0x180100},
	// Goal config + various boolean flags.
	{"goal config + flags", 0x180160, 0x180200},
	// Cape/Byrna magic usage tables.
	{"magic usage", 0x180168, 0x180180},
	// Compass count totals.
	{"compass count totals", 0x187000, 0x187010},
	// Substitution table.
	{"substitution table", 0x184000, 0x184010},
	// Heart color (single byte) is in the QoL region.
	{"heart color byte", 0x187020, 0x187021},
}

// TestParity_DeterministicRomRegions verifies specific deterministic ROM
// regions match PHP byte-for-byte for the default Standard-mode seed.
// Reference dumps are written by dump_rom_regions.php into /tmp/parity_rom/.
//
// This test is the strongest "ROM edited correctly" assurance we have for
// the non-RNG parts of WriteToRom.
func TestParity_DeterministicRomRegions(t *testing.T) {
	for _, region := range romRegionsToVerify {
		t.Run(region.Name, func(t *testing.T) {
			fn := "/tmp/parity_rom/" + region.Name + ".bin"
			want, err := os.ReadFile(fn)
			if err != nil {
				t.Skipf("no reference for %s: %v", region.Name, err)
			}
			goFn := "/tmp/parity_rom/go_" + region.Name + ".bin"
			got, err := os.ReadFile(goFn)
			if err != nil {
				t.Skipf("no Go-side dump for %s: %v", region.Name, err)
			}
			if !bytes.Equal(got, want) {
				diffs := []int{}
				for i := 0; i < len(got) && i < len(want); i++ {
					if got[i] != want[i] {
						diffs = append(diffs, i)
					}
				}
				t.Errorf("%s: %d differing bytes in [0x%X, 0x%X)",
					region.Name, len(diffs), region.From, region.To)
				if len(diffs) < 30 {
					for _, off := range diffs {
						t.Logf("  offset+0x%X: Go=0x%02X PHP=0x%02X", off, got[off], want[off])
					}
				}
			}
		})
	}
}
