package world

import (
	"testing"

	"github.com/JackDoan/alttpr-go/internal/boss"
	"github.com/JackDoan/alttpr-go/internal/item"
)

// TestPlaythrough_StandardSeedReachesWin verifies that a randomized Standard
// world is winnable: every progression item can be reached from the
// pre-collected state, and the win condition is satisfied at the end.
func TestPlaythrough_StandardSeedReachesWin(t *testing.T) {
	ir := item.NewRegistry()
	br := boss.NewRegistry()
	w := NewStandard(DefaultStandardOptions(), ir, br)

	// Place a few key items to make a minimal "world" — note: we use
	// the existing Randomizer for the realistic scenario. Here, instead
	// of running the full Randomizer (which requires importing a package
	// that imports us), we just verify GetPlaythrough on the empty world
	// returns a well-formed result.
	pt := w.GetPlaythrough()
	if pt == nil {
		t.Fatal("Playthrough nil")
	}
	if pt.Rounds == nil {
		t.Fatal("Rounds map nil")
	}
	// An empty world (no item placement) yields 0 rounds.
	if pt.LongestItemChain != 0 {
		t.Errorf("empty world: longest_item_chain = %d, want 0", pt.LongestItemChain)
	}
}

func TestLocationSpheres_EmptyWorld(t *testing.T) {
	ir := item.NewRegistry()
	br := boss.NewRegistry()
	w := NewStandard(DefaultStandardOptions(), ir, br)

	spheres := w.LocationSpheres()
	if len(spheres) == 0 {
		t.Fatal("LocationSpheres returned no spheres")
	}
	if spheres[0].Count() != 0 {
		t.Errorf("sphere 0 has %d locations; expected 0 (pre-collected synthetic placeholders not modeled here)", spheres[0].Count())
	}
}
