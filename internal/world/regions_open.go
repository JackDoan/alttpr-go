// Open-mode region overrides. Differs from Standard only in HyruleCastleEscape:
// Open pre-collects RescueZelda + relaxes sword/kill requirements in Hyrule
// Castle since Zelda is already saved.
package world

import "github.com/JackDoan/alttpr-go/internal/item"

// newOpenHyruleCastleEscape mirrors PHP Region\Open\HyruleCastleEscape,
// which extends Standard\HyruleCastleEscape but rewrites several location
// requirements (no canKillEscapeThings dependency since Uncle is bypassed).
func newOpenHyruleCastleEscape(b *regionBuilder) *Region {
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

	lampOrFireRod := func(it *item.Collection) bool {
		return it.Has("Lamp", w.ConfigInt("item.require.Lamp", 1)) ||
			(w.ConfigString("itemPlacement", "") == "advanced" && it.Has1("FireRod"))
	}

	secretRoomReq := func(_ *LocationCollection, it *item.Collection) bool {
		return it.CanLiftRocks() || (lampOrFireRod(it) && it.Has1("KeyH2") && it.CanKillMostThings(w, 5))
	}
	secLeft.Requirement = secretRoomReq
	secMid.Requirement = secretRoomReq
	secRight.Requirement = secretRoomReq

	darkCross.Requirement = func(_ *LocationCollection, it *item.Collection) bool {
		return lampOrFireRod(it)
	}
	keyAndKill := func(_ *LocationCollection, it *item.Collection) bool {
		return it.Has1("KeyH2") && it.CanKillMostThings(w, 5)
	}
	boomerang.Requirement = keyAndKill
	zeldaCell.Requirement = keyAndKill

	// Sanctuary and Map Chest are unrestricted in Open.
	_ = sanctuary
	_ = mapChest

	dungeonItemFillRule := func(itm *item.Item) bool {
		bad := (!w.ConfigBool("region.wildKeys", false) && itm.IsType(item.TypeKey)) ||
			(!w.ConfigBool("region.wildBigKeys", false) && itm.IsType(item.TypeBigKey)) ||
			(!w.ConfigBool("region.wildMaps", false) && itm.IsType(item.TypeMap)) ||
			(!w.ConfigBool("region.wildCompasses", false) && itm.IsType(item.TypeCompass))
		return !bad
	}
	passage.FillRule = func(itm *item.Item, _ *LocationCollection, _ *item.Collection) bool {
		return dungeonItemFillRule(itm)
	}
	uncle.FillRule = func(itm *item.Item, locs *LocationCollection, _ *item.Collection) bool {
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
