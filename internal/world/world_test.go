package world

import (
	"testing"

	"github.com/JackDoan/alttpr-go/internal/item"
)

// Build a tiny two-location region/world to sanity-check the wiring.
func newToyWorld(t *testing.T) (*World, *Region, *Location, *Location) {
	t.Helper()
	cfg := NewConfig()
	cfg.Strings["mode.weapons"] = "randomized"
	w := NewWorld(0, cfg)
	r := NewRegion("Toy", w)

	// loc1: always accessible, never accepts Keys.
	loc1 := NewLocation("Loc1", []int{0x100}, []int{0x00}, r, nil)
	// loc2: requires Bow.
	loc2 := NewLocation("Loc2", []int{0x101}, []int{0x00}, r,
		func(_ *LocationCollection, items *item.Collection) bool { return items.Has1("Bow") })

	r.Locations.Add(loc1)
	r.Locations.Add(loc2)
	w.AddRegion(r)
	return w, r, loc1, loc2
}

func TestLocation_CanAccess(t *testing.T) {
	_, _, loc1, loc2 := newToyWorld(t)
	emptyItems := item.NewCollection()
	if !loc1.CanAccess(emptyItems, nil) {
		t.Error("loc1 should always be accessible")
	}
	if loc2.CanAccess(emptyItems, nil) {
		t.Error("loc2 without Bow should NOT be accessible")
	}

	reg := item.NewRegistry()
	bow, _ := reg.Get("Bow", 0)
	withBow := item.NewCollection(bow)
	if !loc2.CanAccess(withBow, nil) {
		t.Error("loc2 with Bow should be accessible")
	}
}

func TestLocation_FillRollbackOnFailure(t *testing.T) {
	_, _, _, loc2 := newToyWorld(t)
	itemReg := item.NewRegistry()
	hammer, _ := itemReg.Get("Hammer", 0)

	emptyItems := item.NewCollection()
	// Place Hammer at loc2. With no Bow, canAccess is false -> Fill returns false.
	if loc2.Fill(hammer, emptyItems) {
		t.Error("Fill should have failed (no Bow); did not roll back")
	}
	if loc2.HasItem() {
		t.Errorf("loc2 still has item after failed Fill: %v", loc2.Item())
	}
}

func TestWorld_CollectOtherItems(t *testing.T) {
	w, _, loc1, _ := newToyWorld(t)
	itemReg := item.NewRegistry()
	hammer, _ := itemReg.Get("Hammer", 0)
	loc1.SetItem(hammer)

	gained := w.CollectOtherItems(item.NewCollection())
	if !gained.Has1("Hammer") {
		t.Error("expected Hammer to be collected from loc1")
	}
	// Second pass should yield nothing new (cached).
	gained2 := w.CollectOtherItems(item.NewCollection())
	if gained2.Count() != 0 {
		t.Errorf("second pass yielded %d items, want 0", gained2.Count())
	}
}

func TestRegion_CanFill_KeyRestriction(t *testing.T) {
	_, r, _, _ := newToyWorld(t)
	itemReg := item.NewRegistry()
	keyD1, _ := itemReg.Get("KeyD1", 0)
	// Default: region.wildKeys=false, so Keys can't go into a region they
	// don't belong to.
	if r.CanFill(keyD1) {
		t.Error("KeyD1 should not fit Toy region (wildKeys off, not a region item)")
	}
	// Enabling wildKeys should permit it.
	cfg := r.World.Config()
	cfg.Bools["region.wildKeys"] = true
	if !r.CanFill(keyD1) {
		t.Error("KeyD1 should fit Toy region (wildKeys on)")
	}
}
