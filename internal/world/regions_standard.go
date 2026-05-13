package world

import (
	"github.com/JackDoan/alttpr-go/internal/boss"
	"github.com/JackDoan/alttpr-go/internal/item"
)

// regionBuilder is the per-region construction context shared with closures.
// It centralizes lookups to the world's item registry and the world itself.
type regionBuilder struct {
	w  *World
	ir *item.Registry
	br *boss.Registry
}

func newBuilder(w *World, ir *item.Registry, br *boss.Registry) *regionBuilder {
	return &regionBuilder{w: w, ir: ir, br: br}
}

func (b *regionBuilder) item(name string) *item.Item {
	it, err := b.ir.Get(name, b.w.ID())
	if err != nil {
		panic(err)
	}
	return it
}

func (b *regionBuilder) boss(name string) *boss.Boss {
	bs, err := b.br.Get(name, b.w)
	if err != nil {
		panic(err)
	}
	return bs
}

// regionItems looks up items by name and returns them as a slice.
func (b *regionBuilder) regionItems(names ...string) []*item.Item {
	out := make([]*item.Item, len(names))
	for i, n := range names {
		out[i] = b.item(n)
	}
	return out
}

// loc constructs a Location of the given kind and attaches it to region r.
func loc(r *Region, kind Kind, name string, addr []int) *Location {
	l := NewLocation(name, addr, nil, r, nil)
	l.Kind = kind
	r.Locations.Add(l)
	return l
}

// medallionLoc constructs a Medallion location with named text-byte offsets.
func medallionLoc(r *Region, name string, baseAddr []int, named map[string]int) *Location {
	l := loc(r, KindMedallion, name, baseAddr)
	l.NamedAddresses = named
	return l
}

// ----------------------------------------------------------------------
// Standard / Fountains — fairy fountain prize locations.
// PHP: app/Region/Standard/Fountains.php
// ----------------------------------------------------------------------
func newStandardFountains(b *regionBuilder) *Region {
	r := NewRegion("Fountains", b.w)
	loc(r, KindFountain, "Waterfall Bottle", []int{0x348FF})
	loc(r, KindFountain, "Pyramid Bottle", []int{0x3493B})
	return r
}

// ----------------------------------------------------------------------
// Standard / Medallions — entry-medallion requirement slots for TR / MM.
// PHP: app/Region/Standard/Medallions.php
// ----------------------------------------------------------------------
func newStandardMedallions(b *regionBuilder) *Region {
	r := NewRegion("Medallions", b.w)
	// PHP arrays mix int and string keys; numeric prefix is [-1, 0x180023]
	// (the -1 mirrors PHP's null slot — i.e. byte index 0 has no address).
	medallionLoc(r, "Turtle Rock Medallion", []int{-1, 0x180023},
		map[string]int{"t0": 0x5020, "t1": 0x50FF, "t2": 0x51DE})
	medallionLoc(r, "Misery Mire Medallion", []int{-1, 0x180022},
		map[string]int{"m0": 0x4FF2, "m1": 0x50D1, "m2": 0x51B0})
	return r
}

// ----------------------------------------------------------------------
// Standard / HyruleCastleTower
// PHP: app/Region/Standard/HyruleCastleTower.php
// ----------------------------------------------------------------------
func newStandardHyruleCastleTower(b *regionBuilder) *Region {
	r := NewRegion("Hyrule Castle Tower", b.w)
	r.RegionItems = b.regionItems(
		"BigKey", "BigKeyA1", "Compass", "CompassA1",
		"Key", "KeyA1", "Map", "MapA1")

	loc(r, KindChest, "Castle Tower - Room 03", []int{0xEAB5})
	darkMaze := loc(r, KindChest, "Castle Tower - Dark Maze", []int{0xEAB2})
	agahnim := loc(r, KindPrizeEvent, "Agahnim", nil)
	_ = agahnim.SetItem(b.item("DefeatAgahnim"))
	r.SetPrizeLocation(agahnim)

	w := b.w
	lampReq := func(it *item.Collection) bool {
		return it.Has("Lamp", w.ConfigInt("item.require.Lamp", 1))
	}

	darkMaze.Requirement = func(_ *LocationCollection, it *item.Collection) bool {
		return lampReq(it) && it.Has1("KeyA1")
	}

	r.CanEnterFn = func(_ *LocationCollection, it *item.Collection) bool {
		return it.CanKillMostThings(w, 8) && it.Has1("RescueZelda") &&
			(it.Has1("Cape") || it.HasSword(2) ||
				(w.ConfigString("mode.weapons", "") == "swordless" && it.Has1("Hammer")))
	}

	r.CanCompFn = func(locs *LocationCollection, it *item.Collection) bool {
		return r.CanEnter(locs, it) && it.Has("KeyA1", 2) && lampReq(it) &&
			(it.HasSword(1) ||
				(w.ConfigString("mode.weapons", "") == "swordless" &&
					(it.Has1("Hammer") || it.Has1("BugCatchingNet"))))
	}
	// Prize location inherits can_complete as its requirement.
	agahnim.Requirement = func(locs *LocationCollection, it *item.Collection) bool {
		return r.CanComplete(locs, it)
	}
	return r
}

// ----------------------------------------------------------------------
// Standard / HyruleCastleEscape — Sanctuary, Sewers, Uncle, Secret Passage.
// PHP: app/Region/Standard/HyruleCastleEscape.php
// ----------------------------------------------------------------------
func newStandardHyruleCastleEscape(b *regionBuilder) *Region {
	w := b.w
	r := NewRegion("Escape", w)
	r.RegionItems = b.regionItems(
		"BigKey", "BigKeyH2", "Compass", "CompassH2",
		"Key", "KeyH2", "Map", "MapH2")

	sanctuary := loc(r, KindChest, "Sanctuary", []int{0xEA79})
	secLeft := loc(r, KindChest, "Sewers - Secret Room - Left", []int{0xEB5D})
	secMid := loc(r, KindChest, "Sewers - Secret Room - Middle", []int{0xEB60})
	secRight := loc(r, KindChest, "Sewers - Secret Room - Right", []int{0xEB63})
	darkCross := loc(r, KindChest, "Sewers - Dark Cross", []int{0xE96E})
	boomerang := loc(r, KindChest, "Hyrule Castle - Boomerang Chest", []int{0xE974})
	mapChest := loc(r, KindChest, "Hyrule Castle - Map Chest", []int{0xEB0C})
	zeldaCell := loc(r, KindChest, "Hyrule Castle - Zelda's Cell", []int{0xEB09})
	uncle := loc(r, KindUncle, "Link's Uncle", []int{0x2DF45})
	passage := loc(r, KindChest, "Secret Passage", []int{0xE971})
	zelda := loc(r, KindPrizeEvent, "Zelda", nil)
	_ = zelda.SetItem(b.item("RescueZelda"))
	r.SetPrizeLocation(zelda)

	killAndKey := func(_ *LocationCollection, it *item.Collection) bool {
		return it.CanKillEscapeThings(w) && it.Has1("KeyH2")
	}
	killOnly := func(_ *LocationCollection, it *item.Collection) bool {
		return it.CanKillEscapeThings(w)
	}
	sanctuary.Requirement = killAndKey
	secLeft.Requirement = killAndKey
	secMid.Requirement = killAndKey
	secRight.Requirement = killAndKey
	darkCross.Requirement = killOnly
	boomerang.Requirement = killOnly
	mapChest.Requirement = killOnly
	zeldaCell.Requirement = killOnly

	dungeonItemFillRule := func(itm *item.Item) bool {
		bad := (!w.ConfigBool("region.wildKeys", false) && itm.IsType(item.TypeKey)) ||
			(!w.ConfigBool("region.wildBigKeys", false) && itm.IsType(item.TypeBigKey)) ||
			(!w.ConfigBool("region.wildMaps", false) && itm.IsType(item.TypeMap)) ||
			(!w.ConfigBool("region.wildCompasses", false) && itm.IsType(item.TypeCompass))
		return !bad
	}

	passage.Requirement = killOnly
	passage.FillRule = func(itm *item.Item, _ *LocationCollection, _ *item.Collection) bool {
		return dungeonItemFillRule(itm)
	}
	uncle.FillRule = func(itm *item.Item, locs *LocationCollection, _ *item.Collection) bool {
		// PHP uses world->collectItems() — equivalent here: reachable items
		// from the world's current location state.
		reachable := w.CollectItems()
		return sanctuary.CanAccess(reachable, locs) && dungeonItemFillRule(itm)
	}

	r.CanCompFn = func(locs *LocationCollection, it *item.Collection) bool {
		return sanctuary.CanAccess(it, locs)
	}
	zelda.Requirement = func(locs *LocationCollection, it *item.Collection) bool {
		return r.CanComplete(locs, it)
	}
	return r
}

// ----------------------------------------------------------------------
// Standard / EasternPalace
// PHP: app/Region/Standard/EasternPalace.php
// ----------------------------------------------------------------------
func newStandardEasternPalace(b *regionBuilder) *Region {
	w := b.w
	r := NewRegion("Eastern Palace", w)
	r.MapReveal = 0x2000
	r.RegionItems = b.regionItems(
		"BigKey", "BigKeyP1", "Compass", "CompassP1",
		"Key", "KeyP1", "Map", "MapP1", "PendantOfWisdom")
	r.Boss = b.boss("Armos Knights")

	loc(r, KindChest, "Eastern Palace - Compass Chest", []int{0xE977})
	bigChest := loc(r, KindBigChest, "Eastern Palace - Big Chest", []int{0xE97D})
	loc(r, KindChest, "Eastern Palace - Cannonball Chest", []int{0xE9B3})
	bigKeyChest := loc(r, KindChest, "Eastern Palace - Big Key Chest", []int{0xE9B9})
	loc(r, KindChest, "Eastern Palace - Map Chest", []int{0xE9F5})
	bossDrop := loc(r, KindDrop, "Eastern Palace - Boss", []int{0x180150})
	prize := loc(r, KindPrizePendant, "Eastern Palace - Prize",
		[]int{-1, 0x1209D, 0x53E76, 0x53E77, 0x180052, 0x180070, 0xC6FE})
	prize.MusicAddresses = []int{0x1559A}
	r.SetPrizeLocation(prize)

	bigChest.Requirement = func(_ *LocationCollection, it *item.Collection) bool {
		return it.Has1("BigKeyP1")
	}
	bigKeyChest.Requirement = func(_ *LocationCollection, it *item.Collection) bool {
		return it.Has("Lamp", w.ConfigInt("item.require.Lamp", 1))
	}

	bossDrop.Requirement = func(locs *LocationCollection, it *item.Collection) bool {
		return it.CanShootArrows(w, 1) &&
			(it.Has("Lamp", w.ConfigInt("item.require.Lamp", 1)) ||
				(w.ConfigString("itemPlacement", "") == "advanced" && it.Has1("FireRod"))) &&
			it.Has1("BigKeyP1") &&
			r.Boss.CanBeat(it, locs) &&
			(!w.ConfigBool("region.wildCompasses", false) || it.Has1("CompassP1") ||
				bossDrop.HasSpecificItem(b.item("CompassP1"))) &&
			(!w.ConfigBool("region.wildMaps", false) || it.Has1("MapP1") ||
				bossDrop.HasSpecificItem(b.item("MapP1")))
	}
	bossDrop.FillRule = func(itm *item.Item, _ *LocationCollection, _ *item.Collection) bool {
		if !w.ConfigBool("region.bossNormalLocation", true) &&
			(itm.IsType(item.TypeKey) || itm.IsType(item.TypeBigKey) ||
				itm.IsType(item.TypeMap) || itm.IsType(item.TypeCompass)) {
			return false
		}
		return true
	}
	bossDrop.Always = func(itm *item.Item, _ *item.Collection) bool {
		return w.ConfigBool("region.bossNormalLocation", true) &&
			(itm == b.item("CompassP1") || itm == b.item("MapP1"))
	}

	r.CanEnterFn = func(_ *LocationCollection, it *item.Collection) bool {
		return it.Has1("RescueZelda")
	}
	r.CanCompFn = func(locs *LocationCollection, it *item.Collection) bool {
		return bossDrop.CanAccess(it, locs)
	}
	prize.Requirement = func(locs *LocationCollection, it *item.Collection) bool {
		return r.CanComplete(locs, it)
	}
	return r
}
