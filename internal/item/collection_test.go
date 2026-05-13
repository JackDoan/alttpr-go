package item

import "testing"

type fakeWorld struct {
	id       int
	inverted bool
	strs     map[string]string
	ints     map[string]int
	bools    map[string]bool
}

func (f *fakeWorld) ID() int          { return f.id }
func (f *fakeWorld) IsInverted() bool { return f.inverted }
func (f *fakeWorld) ConfigString(k, def string) string {
	if v, ok := f.strs[k]; ok {
		return v
	}
	return def
}
func (f *fakeWorld) ConfigInt(k string, def int) int {
	if v, ok := f.ints[k]; ok {
		return v
	}
	return def
}
func (f *fakeWorld) ConfigBool(k string, def bool) bool {
	if v, ok := f.bools[k]; ok {
		return v
	}
	return def
}

func makeColl(t *testing.T, names ...string) *Collection {
	t.Helper()
	r := NewRegistry()
	c := NewCollection()
	c.SetChecksForWorld(0)
	for _, n := range names {
		it, err := r.Get(n, 0)
		if err != nil {
			t.Fatalf("Get %s: %v", n, err)
		}
		c.Add(it)
	}
	return c
}

func TestHas_BasicAndShopKey(t *testing.T) {
	c := makeColl(t, "Bow")
	if !c.Has1("Bow") {
		t.Error("Bow not seen")
	}
	if c.Has1("Hammer") {
		t.Error("Hammer falsely seen")
	}
	// ShopKey hack: any Key* satisfies once ShopKey is in pool.
	c2 := makeColl(t, "ShopKey")
	if !c2.Has1("KeyP1") {
		t.Error("ShopKey hack should make KeyP1 true")
	}
}

func TestHas_Threshold(t *testing.T) {
	c := makeColl(t, "ProgressiveSword", "ProgressiveSword", "ProgressiveSword")
	if !c.Has("ProgressiveSword", 3) {
		t.Error("3 ProgressiveSword should satisfy at_least=3")
	}
	if c.Has("ProgressiveSword", 4) {
		t.Error("3 should not satisfy at_least=4")
	}
}

func TestHasSword_Levels(t *testing.T) {
	c := makeColl(t, "ProgressiveSword", "ProgressiveSword")
	if !c.HasSword(2) {
		t.Error("2 ProgressiveSword should pass HasSword(2)")
	}
	if c.HasSword(3) {
		t.Error("2 ProgressiveSword should not pass HasSword(3)")
	}
}

func TestCanLightTorches(t *testing.T) {
	if !makeColl(t, "Lamp").CanLightTorches() {
		t.Error("Lamp should light torches")
	}
	if !makeColl(t, "FireRod").CanLightTorches() {
		t.Error("FireRod should light torches")
	}
	if makeColl(t).CanLightTorches() {
		t.Error("empty should not light torches")
	}
}

func TestBottleCount_HasABottle(t *testing.T) {
	c := makeColl(t, "Bottle", "BottleWithFairy")
	if c.BottleCount() != 2 {
		t.Errorf("BottleCount=%d, want 2", c.BottleCount())
	}
	if !c.HasABottle() {
		t.Error("HasABottle should be true")
	}
}

func TestCanShootArrows_RupeeBow(t *testing.T) {
	w := &fakeWorld{bools: map[string]bool{"rom.rupeeBow": true}}
	c := makeColl(t, "Bow")
	if c.CanShootArrows(w, 1) {
		t.Error("Bow alone with rupeeBow should not shoot")
	}
	c.Add(mustGet(t, "ShopArrow", 0))
	if !c.CanShootArrows(w, 1) {
		t.Error("Bow + ShopArrow with rupeeBow should shoot")
	}
}

func TestHeartCount(t *testing.T) {
	c := makeColl(t, "PieceOfHeart", "PieceOfHeart", "PieceOfHeart", "PieceOfHeart", "HeartContainer")
	got := c.HeartCount(3.0)
	want := 3.0 + 1.0 + 1.0
	if got != want {
		t.Errorf("HeartCount=%v, want %v", got, want)
	}
}

func TestFilterAndDiff(t *testing.T) {
	c := makeColl(t, "Bow", "Hammer", "Lamp")
	bows := c.Filter(func(i *Item) bool { return i.Name == "Bow" })
	if bows.Count() != 1 {
		t.Errorf("filter count=%d, want 1", bows.Count())
	}
	d := c.Diff(makeColl(t, "Hammer"))
	if d.Has1("Hammer") {
		t.Error("Hammer should be diffed out")
	}
	if !d.Has1("Bow") {
		t.Error("Bow should remain")
	}
}

func mustGet(t *testing.T, name string, worldID int) *Item {
	t.Helper()
	it, err := NewRegistry().Get(name, worldID)
	if err != nil {
		t.Fatalf("Get %s: %v", name, err)
	}
	return it
}
