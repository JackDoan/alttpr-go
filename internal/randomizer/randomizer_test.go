package randomizer

import (
	"testing"

	"github.com/JackDoan/alttpr-go/internal/boss"
	"github.com/JackDoan/alttpr-go/internal/item"
	"github.com/JackDoan/alttpr-go/internal/world"
)

// TestRandomize_StandardEndToEnd runs the full Standard pipeline:
// build a Standard world, prepare it (medallions, fountains, prize fill),
// run the filler. This is the closest thing to an end-to-end smoke test of
// Phase 2 so far. We don't yet validate against PHP byte-for-byte (RNG is
// non-seedable), but we do check structural invariants.
func TestRandomize_StandardEndToEnd(t *testing.T) {
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

	// All medallion slots filled.
	for _, l := range w.Region("Medallions").Locations.All() {
		if !l.HasItem() {
			t.Errorf("medallion slot %s unfilled", l.Name)
		}
	}
	// Both fountains filled.
	for _, l := range w.Region("Fountains").Locations.All() {
		if !l.HasItem() {
			t.Errorf("fountain %s unfilled", l.Name)
		}
	}
	// Bosses placed in vanilla mode.
	if w.Region("Eastern Palace").Boss == nil || w.Region("Eastern Palace").Boss.Name != "Armos Knights" {
		t.Error("Eastern Palace boss not set to Armos Knights")
	}
	if w.Region("Turtle Rock").Boss == nil || w.Region("Turtle Rock").Boss.Name != "Trinexx" {
		t.Error("Turtle Rock boss not set to Trinexx")
	}
	// Pre-collected items seeded.
	pc := w.PreCollectedItems()
	if pc.Count() < 7 {
		t.Errorf("pre-collected count = %d, want >= 7", pc.Count())
	}

	// Verify a bunch of locations got filled.
	empty := w.EmptyLocations().Count()
	total := w.Locations().Count()
	filled := total - empty
	t.Logf("Randomized: %d/%d locations filled", filled, total)
	if filled < 100 {
		t.Errorf("expected 100+ filled locations, got %d", filled)
	}
}
