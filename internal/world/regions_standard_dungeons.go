package world

import (
	"github.com/JackDoan/alttpr-go/internal/boss"
	"github.com/JackDoan/alttpr-go/internal/item"
)

// Helpers shared across dungeons.
type dungeonHelpers struct {
	r *Region
	b *regionBuilder
}

// bossLocCommon installs the conventional fill/always rules on a boss-drop
// location (matches the repeated PHP block in every dungeon).
func (h dungeonHelpers) bossDropRules(loc *Location, compassName, mapName string) {
	w := h.b.w
	loc.FillRule = func(itm *item.Item, _ *LocationCollection, _ *item.Collection) bool {
		if !w.ConfigBool("region.bossNormalLocation", true) &&
			(itm.IsType(item.TypeKey) || itm.IsType(item.TypeBigKey) ||
				itm.IsType(item.TypeMap) || itm.IsType(item.TypeCompass)) {
			return false
		}
		return true
	}
	loc.Always = func(itm *item.Item, _ *item.Collection) bool {
		return w.ConfigBool("region.bossNormalLocation", true) &&
			(itm == h.b.item(compassName) || itm == h.b.item(mapName))
	}
}

// wildCompMap returns the inline `!wildCompasses || items.has(compass) || boss.has(compass)` and same for map.
func wildCompMap(w *World, boss *Location, items *item.Collection, b *regionBuilder, compassName, mapName string) bool {
	return (!w.ConfigBool("region.wildCompasses", false) || items.Has1(compassName) || boss.HasSpecificItem(b.item(compassName))) &&
		(!w.ConfigBool("region.wildMaps", false) || items.Has1(mapName) || boss.HasSpecificItem(b.item(mapName)))
}

// ----------------------------------------------------------------------
// Standard / DesertPalace
// PHP: app/Region/Standard/DesertPalace.php
// ----------------------------------------------------------------------
func newStandardDesertPalace(b *regionBuilder) *Region {
	w := b.w
	r := NewRegion("Desert Palace", w)
	r.MapReveal = 0x1000
	r.RegionItems = b.regionItems(
		"BigKey", "BigKeyP2", "Compass", "CompassP2",
		"Key", "KeyP2", "Map", "MapP2", "MapP2", "PendantOfPower")
	r.Boss = b.boss("Lanmolas")

	bigChest := loc(r, KindBigChest, "Desert Palace - Big Chest", []int{0xE98F})
	loc(r, KindChest, "Desert Palace - Map Chest", []int{0xE9B6})
	torch := loc(r, KindDash, "Desert Palace - Torch", []int{0x180160})
	bigKeyChest := loc(r, KindChest, "Desert Palace - Big Key Chest", []int{0xE9C2})
	compassChest := loc(r, KindChest, "Desert Palace - Compass Chest", []int{0xE9CB})
	bossDrop := loc(r, KindDrop, "Desert Palace - Boss", []int{0x180151})
	prize := loc(r, KindPrizePendant, "Desert Palace - Prize",
		[]int{-1, 0x1209E, 0x53E7A, 0x53E7B, 0x180053, 0x180072, 0xC6FF})
	prize.MusicAddresses = []int{0x1559B, 0x1559C, 0x1559D, 0x1559E}
	r.SetPrizeLocation(prize)

	bigChest.Requirement = func(_ *LocationCollection, it *item.Collection) bool { return it.Has1("BigKeyP2") }
	bigKeyChest.Requirement = func(_ *LocationCollection, it *item.Collection) bool {
		return it.Has1("KeyP2") && it.CanKillMostThings(w, 5)
	}
	compassChest.Requirement = func(_ *LocationCollection, it *item.Collection) bool { return it.Has1("KeyP2") }
	torch.Requirement = func(_ *LocationCollection, it *item.Collection) bool { return it.Has1("PegasusBoots") }

	bossDrop.Requirement = func(locs *LocationCollection, it *item.Collection) bool {
		mireOk := false
		if mire := w.Region("Mire"); mire != nil {
			mireOk = mire.CanEnter(locs, it)
		}
		return r.CanEnter(locs, it) &&
			(it.CanLiftRocks() ||
				(w.ConfigBool("canBootsClip", false) && it.Has1("PegasusBoots")) ||
				(w.ConfigBool("canSuperSpeed", false) && it.CanSpinSpeed()) ||
				w.ConfigBool("canOneFrameClipOW", false) ||
				(it.Has1("MagicMirror") && mireOk)) &&
			it.CanLightTorches() &&
			it.Has1("BigKeyP2") && it.Has1("KeyP2") &&
			r.Boss.CanBeat(it, locs) &&
			wildCompMap(w, bossDrop, it, b, "CompassP2", "MapP2")
	}
	h := dungeonHelpers{r: r, b: b}
	h.bossDropRules(bossDrop, "CompassP2", "MapP2")
	// Desert overrides FillRule to also disallow KeyP2/BigKeyP2 directly.
	prev := bossDrop.FillRule
	bossDrop.FillRule = func(itm *item.Item, locs *LocationCollection, items *item.Collection) bool {
		if !prev(itm, locs, items) {
			return false
		}
		if itm == b.item("KeyP2") || itm == b.item("BigKeyP2") {
			return false
		}
		return true
	}

	r.CanEnterFn = func(locs *LocationCollection, it *item.Collection) bool {
		mireOk := false
		if mire := w.Region("Mire"); mire != nil {
			mireOk = mire.CanEnter(locs, it)
		}
		return it.Has1("RescueZelda") &&
			(it.Has1("BookOfMudora") ||
				(w.ConfigBool("canBootsClip", false) && it.Has1("PegasusBoots")) ||
				w.ConfigBool("canOneFrameClipOW", false) ||
				(it.Has1("MagicMirror") && mireOk))
	}
	r.CanCompFn = func(locs *LocationCollection, it *item.Collection) bool {
		return bossDrop.CanAccess(it, locs)
	}
	prize.Requirement = func(locs *LocationCollection, it *item.Collection) bool { return r.CanComplete(locs, it) }
	return r
}

// ----------------------------------------------------------------------
// Standard / TowerOfHera
// PHP: app/Region/Standard/TowerOfHera.php
// ----------------------------------------------------------------------
func newStandardTowerOfHera(b *regionBuilder) *Region {
	w := b.w
	r := NewRegion("Tower of Hera", w)
	r.MapReveal = 0x0020
	r.RegionItems = b.regionItems(
		"BigKey", "BigKeyP3", "Compass", "CompassP3",
		"Key", "KeyP3", "Map", "MapP3", "PendantOfWisdom")
	r.Boss = b.boss("Moldorm")
	r.CanPlaceBossFn = func(bb *boss.Boss) bool {
		// Inline matching the PHP override.
		if w.ConfigString("mode.weapons", "") == "swordless" && bb.Name == "Kholdstare" {
			return false
		}
		switch bb.Name {
		case "Agahnim", "Agahnim2", "Armos Knights", "Arrghus", "Blind", "Ganon", "Lanmolas", "Trinexx":
			return false
		}
		return true
	}

	bigKeyChest := loc(r, KindChest, "Tower of Hera - Big Key Chest", []int{0xE9E6})
	loc(r, KindHeraBasement, "Tower of Hera - Basement Cage", []int{0x180162})
	loc(r, KindChest, "Tower of Hera - Map Chest", []int{0xE9AD})
	compassChest := loc(r, KindChest, "Tower of Hera - Compass Chest", []int{0xE9FB})
	bigChest := loc(r, KindBigChest, "Tower of Hera - Big Chest", []int{0xE9F8})
	bossDrop := loc(r, KindDrop, "Tower of Hera - Boss", []int{0x180152})
	prize := loc(r, KindPrizePendant, "Tower of Hera - Prize",
		[]int{-1, 0x120A5, 0x53E78, 0x53E79, 0x18005A, 0x180071, 0xC706})
	prize.MusicAddresses = []int{0x155C5, 0x1107A, 0x10B8C}
	r.SetPrizeLocation(prize)

	main := func(locs *LocationCollection, it *item.Collection) bool {
		westDM := false
		if rr := w.Region("West Death Mountain"); rr != nil {
			westDM = rr.CanEnter(locs, it)
		}
		return (it.Has1("PegasusBoots") && w.ConfigBool("canBootsClip", false)) ||
			(w.ConfigBool("canSuperSpeed", false) && it.CanSpinSpeed()) ||
			w.ConfigBool("canOneFrameClipOW", false) ||
			((it.Has1("MagicMirror") || (it.Has1("Hookshot") && it.Has1("Hammer"))) && westDM)
	}

	mire := func(locs *LocationCollection, it *item.Collection) bool {
		mmOk := false
		if rr := w.Region("Misery Mire"); rr != nil {
			mmOk = rr.CanEnter(locs, it)
		}
		bigKeyD6 := b.item("BigKeyD6")
		return w.ConfigBool("canOneFrameClipUW", false) &&
			((locs.ItemInLocations(bigKeyD6, w.ID(),
				[]string{"Misery Mire - Compass Chest", "Misery Mire - Big Key Chest"}, 1) &&
				it.Has("KeyD6", 2)) ||
				it.Has("KeyD6", 3)) && mmOk
	}

	bigKeyChest.Requirement = func(_ *LocationCollection, it *item.Collection) bool {
		return it.CanLightTorches() && it.Has1("KeyP3")
	}
	bigKeyChest.Always = func(itm *item.Item, _ *item.Collection) bool {
		return w.ConfigString("accessibility", "") != "locations" && itm == b.item("KeyP3")
	}
	bigKeyChest.FillRule = func(itm *item.Item, _ *LocationCollection, _ *item.Collection) bool {
		return w.ConfigString("accessibility", "") != "locations" || itm != b.item("KeyP3")
	}

	compassChest.Requirement = func(locs *LocationCollection, it *item.Collection) bool {
		return (main(locs, it) && it.Has1("BigKeyP3")) || mire(locs, it)
	}
	bigChest.Requirement = func(locs *LocationCollection, it *item.Collection) bool {
		return (main(locs, it) && it.Has1("BigKeyP3")) ||
			(mire(locs, it) && (it.Has1("BigKeyP3") || it.Has1("BigKeyD6")))
	}

	bossDrop.Requirement = func(locs *LocationCollection, it *item.Collection) bool {
		return main(locs, it) &&
			r.Boss.CanBeat(it, locs) &&
			(it.Has1("BigKeyP3") || (mire(locs, it) && it.Has1("BigKeyD6"))) &&
			wildCompMap(w, bossDrop, it, b, "CompassP3", "MapP3")
	}
	dh := dungeonHelpers{r: r, b: b}
	dh.bossDropRules(bossDrop, "CompassP3", "MapP3")

	r.CanEnterFn = func(locs *LocationCollection, it *item.Collection) bool {
		return it.Has1("RescueZelda") && (main(locs, it) || mire(locs, it))
	}
	r.CanCompFn = func(locs *LocationCollection, it *item.Collection) bool {
		return bossDrop.CanAccess(it, locs)
	}
	prize.Requirement = func(locs *LocationCollection, it *item.Collection) bool { return r.CanComplete(locs, it) }
	return r
}

// ----------------------------------------------------------------------
// Standard / IcePalace
// PHP: app/Region/Standard/IcePalace.php
// ----------------------------------------------------------------------
func newStandardIcePalace(b *regionBuilder) *Region {
	w := b.w
	r := NewRegion("Ice Palace", w)
	r.MapReveal = 0x0040
	r.RegionItems = b.regionItems(
		"BigKey", "BigKeyD5", "Compass", "CompassD5",
		"Key", "KeyD5", "Map", "MapD5", "Crystal5")
	r.Boss = b.boss("Kholdstare")

	bigKeyChest := loc(r, KindChest, "Ice Palace - Big Key Chest", []int{0xE9A4})
	loc(r, KindChest, "Ice Palace - Compass Chest", []int{0xE9D4})
	mapChest := loc(r, KindChest, "Ice Palace - Map Chest", []int{0xE9DD})
	spikeRoom := loc(r, KindChest, "Ice Palace - Spike Room", []int{0xE9E0})
	freezor := loc(r, KindChest, "Ice Palace - Freezor Chest", []int{0xE995})
	loc(r, KindChest, "Ice Palace - Iced T Room", []int{0xE9E3})
	bigChest := loc(r, KindBigChest, "Ice Palace - Big Chest", []int{0xE9AA})
	bossDrop := loc(r, KindDrop, "Ice Palace - Boss", []int{0x180157})
	prize := loc(r, KindPrizeCrystal, "Ice Palace - Prize",
		[]int{-1, 0x120A4, 0x53E86, 0x53E87, 0x180059, 0x180078, 0xC705})
	prize.MusicAddresses = []int{0x155BF}
	r.SetPrizeLocation(prize)

	noDamage := func(it *item.Collection) bool {
		return !w.ConfigBool("region.cantTakeDamage", false) ||
			it.Has1("CaneOfByrna") || it.Has1("Cape") || it.Has1("Hookshot")
	}

	spikeRoom.Requirement = func(locs *LocationCollection, it *item.Collection) bool {
		bigKeyD5 := b.item("BigKeyD5")
		hasInListed := locs.ItemInLocations(bigKeyD5, w.ID(),
			[]string{"Ice Palace - Spike Room", "Ice Palace - Big Key Chest", "Ice Palace - Map Chest"}, 1)
		return noDamage(it) &&
			((it.Has1("Hookshot") || it.Has1("ShopKey")) ||
				(it.Has("KeyD5", 1) && hasInListed))
	}
	hammerLiftSpike := func(locs *LocationCollection, it *item.Collection) bool {
		return it.Has1("Hammer") && it.CanLiftRocks() && noDamage(it) &&
			spikeRoom.CanAccess(it, locs)
	}
	bigKeyChest.Requirement = hammerLiftSpike
	mapChest.Requirement = hammerLiftSpike
	freezor.Requirement = func(_ *LocationCollection, it *item.Collection) bool { return it.CanMeltThings(w) }
	bigChest.Requirement = func(_ *LocationCollection, it *item.Collection) bool { return it.Has1("BigKeyD5") }

	bossDrop.Requirement = func(locs *LocationCollection, it *item.Collection) bool {
		basic := w.ConfigString("itemPlacement", "") == "basic"
		keyReq := false
		if !basic {
			keyReq = (it.Has1("CaneOfSomaria") && it.Has1("KeyD5")) || it.Has("KeyD5", 2)
		} else {
			keyReq = it.Has("KeyD5", 2)
		}
		return r.CanEnter(locs, it) &&
			it.Has1("Hammer") && it.CanLiftRocks() &&
			r.Boss.CanBeat(it, locs) &&
			it.Has1("BigKeyD5") && keyReq &&
			wildCompMap(w, bossDrop, it, b, "CompassD5", "MapD5")
	}
	dungeonHelpers{r: r, b: b}.bossDropRules(bossDrop, "CompassD5", "MapD5")

	r.CanEnterFn = func(locs *LocationCollection, it *item.Collection) bool {
		southDW := false
		if rr := w.Region("South Dark World"); rr != nil {
			southDW = rr.CanEnter(locs, it)
		}
		basicGate := w.ConfigString("itemPlacement", "") != "basic" ||
			((w.ConfigString("mode.weapons", "") == "swordless" || it.HasSword(2)) &&
				it.HasHealth(12) && (it.HasBottle(2) || it.HasArmor(1)))
		meltGate := it.CanMeltThings(w) || w.ConfigBool("canOneFrameClipUW", false)

		directRoute := (it.Has1("MoonPearl") || w.ConfigBool("canDungeonRevive", false)) &&
			(it.Has1("Flippers") || w.ConfigBool("canFakeFlipper", false)) &&
			it.CanLiftDarkRocks()

		wrap := w.ConfigBool("canOneFrameClipOW", false) &&
			w.ConfigBool("canMirrorWrap", false) && it.Has1("MagicMirror")

		viaSouthDW := southDW && (((it.Has1("MoonPearl") ||
			(it.HasABottle() && w.ConfigBool("canOWYBA", false)) ||
			(w.ConfigBool("canBunnyRevive", false) && it.CanBunnyRevive(w))) &&
			((w.ConfigBool("canMirrorWrap", false) && it.Has1("MagicMirror") &&
				((w.ConfigBool("canBootsClip", false) && it.Has1("PegasusBoots")) ||
					(w.ConfigBool("canSuperSpeed", false) && it.CanSpinSpeed()))) ||
				(it.Has1("Flippers") && w.ConfigBool("canTransitionWrapped", false) &&
					((w.ConfigBool("canBootsClip", false) && it.Has1("PegasusBoots")) ||
						w.ConfigBool("canOneFrameClipOW", false))))) || wrap)

		return it.Has1("RescueZelda") && basicGate && meltGate && (directRoute || viaSouthDW)
	}
	r.CanCompFn = func(locs *LocationCollection, it *item.Collection) bool { return bossDrop.CanAccess(it, locs) }
	prize.Requirement = func(locs *LocationCollection, it *item.Collection) bool { return r.CanComplete(locs, it) }
	return r
}

// ----------------------------------------------------------------------
// Standard / ThievesTown
// PHP: app/Region/Standard/ThievesTown.php
// ----------------------------------------------------------------------
func newStandardThievesTown(b *regionBuilder) *Region {
	w := b.w
	r := NewRegion("Thieves Town", w)
	r.MapReveal = 0x0010
	r.RegionItems = b.regionItems(
		"BigKey", "BigKeyD4", "Compass", "CompassD4",
		"Key", "KeyD4", "Map", "MapD4", "Crystal4")
	r.Boss = b.boss("Blind")

	attic := loc(r, KindChest, "Thieves' Town - Attic", []int{0xEA0D})
	loc(r, KindChest, "Thieves' Town - Big Key Chest", []int{0xEA04})
	loc(r, KindChest, "Thieves' Town - Map Chest", []int{0xEA01})
	loc(r, KindChest, "Thieves' Town - Compass Chest", []int{0xEA07})
	loc(r, KindChest, "Thieves' Town - Ambush Chest", []int{0xEA0A})
	bigChest := loc(r, KindBigChest, "Thieves' Town - Big Chest", []int{0xEA10})
	blindsCell := loc(r, KindChest, "Thieves' Town - Blind's Cell", []int{0xEA13})
	bossDrop := loc(r, KindDrop, "Thieves' Town - Boss", []int{0x180156})
	prize := loc(r, KindPrizeCrystal, "Thieves' Town - Prize",
		[]int{-1, 0x120A6, 0x53E82, 0x53E83, 0x18005B, 0x180076, 0xC707})
	prize.MusicAddresses = []int{0x155C6}
	r.SetPrizeLocation(prize)

	attic.Requirement = func(_ *LocationCollection, it *item.Collection) bool {
		return it.Has1("KeyD4") && it.Has1("BigKeyD4")
	}
	bigChest.Requirement = func(locs *LocationCollection, it *item.Collection) bool {
		if bigChest.HasSpecificItem(b.item("KeyD4")) {
			return it.Has1("Hammer") && it.Has1("BigKeyD4")
		}
		return it.Has1("Hammer") && it.Has1("KeyD4") && it.Has1("BigKeyD4")
	}
	bigChest.Always = func(itm *item.Item, items *item.Collection) bool {
		return w.ConfigString("accessibility", "") != "locations" && itm == b.item("KeyD4") && items.Has1("Hammer")
	}
	bigChest.FillRule = func(itm *item.Item, _ *LocationCollection, _ *item.Collection) bool {
		return w.ConfigString("accessibility", "") != "locations" || itm != b.item("KeyD4")
	}
	blindsCell.Requirement = func(_ *LocationCollection, it *item.Collection) bool { return it.Has1("BigKeyD4") }

	bossDrop.Requirement = func(locs *LocationCollection, it *item.Collection) bool {
		return r.CanEnter(locs, it) &&
			it.Has1("KeyD4") && it.Has1("BigKeyD4") &&
			r.Boss.CanBeat(it, locs) &&
			wildCompMap(w, bossDrop, it, b, "CompassD4", "MapD4")
	}
	dungeonHelpers{r: r, b: b}.bossDropRules(bossDrop, "CompassD4", "MapD4")

	r.CanEnterFn = func(locs *LocationCollection, it *item.Collection) bool {
		nwdw := false
		if rr := w.Region("North West Dark World"); rr != nil {
			nwdw = rr.CanEnter(locs, it)
		}
		basicGate := w.ConfigString("itemPlacement", "") != "basic" ||
			((w.ConfigString("mode.weapons", "") == "swordless" || it.HasSword(1)) && it.HasHealth(7) && it.HasABottle())
		bunnyOk := it.Has1("MoonPearl") ||
			(it.HasABottle() && w.ConfigBool("canOWYBA", false)) ||
			(w.ConfigBool("canBunnyRevive", false) && it.CanSpinSpeed())
		return it.Has1("RescueZelda") && basicGate && bunnyOk && nwdw
	}
	r.CanCompFn = func(locs *LocationCollection, it *item.Collection) bool { return bossDrop.CanAccess(it, locs) }
	prize.Requirement = func(locs *LocationCollection, it *item.Collection) bool { return r.CanComplete(locs, it) }
	return r
}

// ----------------------------------------------------------------------
// Standard / SkullWoods
// PHP: app/Region/Standard/SkullWoods.php
// ----------------------------------------------------------------------
func newStandardSkullWoods(b *regionBuilder) *Region {
	w := b.w
	r := NewRegion("Skull Woods", w)
	r.MapReveal = 0x0080
	r.RegionItems = b.regionItems(
		"BigKey", "BigKeyD3", "Compass", "CompassD3",
		"Key", "KeyD3", "Map", "MapD3", "Crystal3")
	r.Boss = b.boss("Mothula")
	r.CanPlaceBossFn = func(bb *boss.Boss) bool {
		if w.ConfigString("mode.weapons", "") == "swordless" && bb.Name == "Kholdstare" {
			return false
		}
		switch bb.Name {
		case "Agahnim", "Agahnim2", "Ganon", "Trinexx":
			return false
		}
		return true
	}

	bigChest := loc(r, KindBigChest, "Skull Woods - Big Chest", []int{0xE998})
	loc(r, KindChest, "Skull Woods - Big Key Chest", []int{0xE99E})
	loc(r, KindChest, "Skull Woods - Compass Chest", []int{0xE992})
	loc(r, KindChest, "Skull Woods - Map Chest", []int{0xE99B})
	bridgeRoom := loc(r, KindChest, "Skull Woods - Bridge Room", []int{0xE9FE})
	loc(r, KindChest, "Skull Woods - Pot Prison", []int{0xE9A1})
	pinball := loc(r, KindChest, "Skull Woods - Pinball Room", []int{0xE9C8})
	bossDrop := loc(r, KindDrop, "Skull Woods - Boss", []int{0x180155})
	prize := loc(r, KindPrizeCrystal, "Skull Woods - Prize",
		[]int{-1, 0x120A3, 0x53E7E, 0x53E7F, 0x180058, 0x180074, 0xC704})
	prize.MusicAddresses = []int{0x155BA, 0x155BB, 0x155BC, 0x155BD, 0x15608, 0x15609, 0x1560A, 0x1560B}
	r.SetPrizeLocation(prize)

	pinball.FillRule = func(itm *item.Item, _ *LocationCollection, _ *item.Collection) bool {
		return !w.ConfigBool("region.forceSkullWoodsKey", false) || itm == b.item("KeyD3")
	}

	bigChest.Requirement = func(_ *LocationCollection, it *item.Collection) bool { return it.Has1("BigKeyD3") }
	bigChest.Always = func(itm *item.Item, _ *item.Collection) bool {
		return w.ConfigString("accessibility", "") != "locations" && itm == b.item("BigKeyD3")
	}
	bigChest.FillRule = func(itm *item.Item, _ *LocationCollection, _ *item.Collection) bool {
		return w.ConfigString("accessibility", "") != "locations" || itm != b.item("BigKeyD3")
	}
	bridgeRoom.Requirement = func(_ *LocationCollection, it *item.Collection) bool {
		return it.Has1("FireRod") &&
			(it.Has1("MoonPearl") ||
				(w.ConfigBool("canDungeonRevive", false) && it.Has1("PegasusBoots") &&
					(it.Has1("MagicMirror") || it.HasABottle())))
	}

	bossDrop.Requirement = func(locs *LocationCollection, it *item.Collection) bool {
		return r.CanEnter(locs, it) &&
			it.Has1("FireRod") &&
			(it.Has1("MoonPearl") ||
				(w.ConfigBool("canDungeonRevive", false) && it.Has1("PegasusBoots") &&
					(it.Has1("MagicMirror") || it.HasABottle()))) &&
			(w.ConfigString("mode.weapons", "") == "swordless" || it.HasSword(1)) &&
			it.Has("KeyD3", 3) &&
			r.Boss.CanBeat(it, locs) &&
			wildCompMap(w, bossDrop, it, b, "CompassD3", "MapD3")
	}
	dungeonHelpers{r: r, b: b}.bossDropRules(bossDrop, "CompassD3", "MapD3")

	r.CanEnterFn = func(locs *LocationCollection, it *item.Collection) bool {
		nwdw := false
		if rr := w.Region("North West Dark World"); rr != nil {
			nwdw = rr.CanEnter(locs, it)
		}
		basicGate := w.ConfigString("itemPlacement", "") != "basic" ||
			((w.ConfigString("mode.weapons", "") == "swordless" || it.HasSword(1)) && it.HasHealth(7) && it.HasABottle())
		bunnyOk := w.ConfigBool("canDungeonRevive", false) || it.Has1("MoonPearl") ||
			(it.HasABottle() && w.ConfigBool("canOWYBA", false)) ||
			(w.ConfigBool("canBunnyRevive", false) && it.CanBunnyRevive(w))
		return it.Has1("RescueZelda") && basicGate && bunnyOk && nwdw
	}
	r.CanCompFn = func(locs *LocationCollection, it *item.Collection) bool { return bossDrop.CanAccess(it, locs) }
	prize.Requirement = func(locs *LocationCollection, it *item.Collection) bool { return r.CanComplete(locs, it) }
	return r
}

// ----------------------------------------------------------------------
// Standard / SwampPalace
// PHP: app/Region/Standard/SwampPalace.php
// ----------------------------------------------------------------------
func newStandardSwampPalace(b *regionBuilder) *Region {
	w := b.w
	r := NewRegion("Swamp Palace", w)
	r.MapReveal = 0x0400
	r.RegionItems = b.regionItems(
		"BigKey", "BigKeyD2", "Compass", "CompassD2",
		"Key", "KeyD2", "Map", "MapD2", "Crystal2")
	r.Boss = b.boss("Arrghus")

	entrance := loc(r, KindChest, "Swamp Palace - Entrance", []int{0xEA9D})
	bigChest := loc(r, KindBigChest, "Swamp Palace - Big Chest", []int{0xE989})
	bigKeyChest := loc(r, KindChest, "Swamp Palace - Big Key Chest", []int{0xEAA6})
	mapChest := loc(r, KindChest, "Swamp Palace - Map Chest", []int{0xE986})
	westChest := loc(r, KindChest, "Swamp Palace - West Chest", []int{0xEAA3})
	compassChest := loc(r, KindChest, "Swamp Palace - Compass Chest", []int{0xEAA0})
	floodLeft := loc(r, KindChest, "Swamp Palace - Flooded Room - Left", []int{0xEAA9})
	floodRight := loc(r, KindChest, "Swamp Palace - Flooded Room - Right", []int{0xEAAC})
	waterfall := loc(r, KindChest, "Swamp Palace - Waterfall Room", []int{0xEAAF})
	bossDrop := loc(r, KindDrop, "Swamp Palace - Boss", []int{0x180154})
	prize := loc(r, KindPrizeCrystal, "Swamp Palace - Prize",
		[]int{-1, 0x120A0, 0x53E88, 0x53E89, 0x180055, 0x180079, 0xC701})
	prize.MusicAddresses = []int{0x155B7}
	r.SetPrizeLocation(prize)

	mire := func(locs *LocationCollection, it *item.Collection) bool {
		mm := false
		if rr := w.Region("Misery Mire"); rr != nil {
			mm = rr.CanEnter(locs, it)
		}
		return w.ConfigBool("canOneFrameClipUW", false) && it.Has("KeyD6", 3) && mm
	}
	hera := func(locs *LocationCollection, it *item.Collection) bool {
		th := false
		if rr := w.Region("Tower of Hera"); rr != nil {
			th = rr.CanEnter(locs, it)
		}
		return w.ConfigBool("canOneFrameClipUW", false) && th && it.Has1("BigKeyP3")
	}
	hasKeyOrAlt := func(locs *LocationCollection, it *item.Collection) bool {
		return it.Has1("KeyD2") || mire(locs, it)
	}
	hammerOrAlt := func(locs *LocationCollection, it *item.Collection) bool {
		return it.Has1("Hammer") || mire(locs, it) || hera(locs, it)
	}

	entrance.FillRule = func(itm *item.Item, _ *LocationCollection, _ *item.Collection) bool {
		return w.ConfigBool("canOneFrameClipUW", false) || w.ConfigBool("region.wildKeys", false) || itm == b.item("KeyD2")
	}
	bigChest.Requirement = func(locs *LocationCollection, it *item.Collection) bool {
		return hasKeyOrAlt(locs, it) && hammerOrAlt(locs, it) &&
			(it.Has1("BigKeyD2") || (mire(locs, it) && it.Has1("BigKeyD6")) || (hera(locs, it) && it.Has1("BigKeyP3")))
	}
	bigChest.Always = func(itm *item.Item, _ *item.Collection) bool {
		return w.ConfigString("accessibility", "") != "locations" && itm == b.item("BigKeyD2")
	}
	bigChest.FillRule = func(itm *item.Item, _ *LocationCollection, _ *item.Collection) bool {
		return w.ConfigString("accessibility", "") != "locations" || itm != b.item("BigKeyD2")
	}
	bigKeyChest.Requirement = func(locs *LocationCollection, it *item.Collection) bool {
		return hasKeyOrAlt(locs, it) && hammerOrAlt(locs, it)
	}
	mapChest.Requirement = func(locs *LocationCollection, it *item.Collection) bool {
		return it.CanBombThings() && hasKeyOrAlt(locs, it)
	}
	westChest.Requirement = bigKeyChest.Requirement
	compassChest.Requirement = bigKeyChest.Requirement
	hkReq := func(locs *LocationCollection, it *item.Collection) bool {
		return it.Has1("Hookshot") && hasKeyOrAlt(locs, it) && hammerOrAlt(locs, it)
	}
	floodLeft.Requirement = hkReq
	floodRight.Requirement = hkReq
	waterfall.Requirement = hkReq

	bossDrop.Requirement = func(locs *LocationCollection, it *item.Collection) bool {
		return it.Has1("Hookshot") && hasKeyOrAlt(locs, it) && hammerOrAlt(locs, it) &&
			r.Boss.CanBeat(it, locs) &&
			wildCompMap(w, bossDrop, it, b, "CompassD2", "MapD2")
	}
	dungeonHelpers{r: r, b: b}.bossDropRules(bossDrop, "CompassD2", "MapD2")

	r.CanEnterFn = func(locs *LocationCollection, it *item.Collection) bool {
		south := false
		if rr := w.Region("South Dark World"); rr != nil {
			south = rr.CanEnter(locs, it)
		}
		basicGate := w.ConfigString("itemPlacement", "") != "basic" ||
			((w.ConfigString("mode.weapons", "") == "swordless" || it.HasSword(1)) && it.HasHealth(7) && it.HasABottle())
		oldMan := false
		if ol := w.Locations().Get("Old Man:" + itoaW(w.ID())); ol != nil {
			oldMan = ol.CanAccess(it, locs)
		}
		mirrorPath := it.Has1("MoonPearl") && it.Has1("MagicMirror")
		altPath := w.ConfigBool("canOneFrameClipUW", false) &&
			(it.Has1("BigKeyP3") || it.Has1("BigKeyD6")) && mire(locs, it) && oldMan &&
			((it.Has1("PegasusBoots") && w.ConfigBool("canBootsClip", false)) ||
				(w.ConfigBool("canSuperSpeed", false) && it.CanSpinSpeed()) ||
				w.ConfigBool("canOneFrameClipOW", false))
		return it.Has1("RescueZelda") && basicGate && it.Has1("Flippers") && south && (mirrorPath || altPath)
	}
	r.CanCompFn = func(locs *LocationCollection, it *item.Collection) bool { return bossDrop.CanAccess(it, locs) }
	prize.Requirement = func(locs *LocationCollection, it *item.Collection) bool { return r.CanComplete(locs, it) }
	return r
}

// ----------------------------------------------------------------------
// Standard / MiseryMire
// PHP: app/Region/Standard/MiseryMire.php
// ----------------------------------------------------------------------
func newStandardMiseryMire(b *regionBuilder) *Region {
	w := b.w
	r := NewRegion("Misery Mire", w)
	r.MapReveal = 0x0100
	r.RegionItems = b.regionItems(
		"BigKey", "BigKeyD6", "Compass", "CompassD6",
		"Key", "KeyD6", "Map", "MapD6", "Crystal6")
	r.Boss = b.boss("Vitreous")

	bigChest := loc(r, KindBigChest, "Misery Mire - Big Chest", []int{0xEA67})
	mainLobby := loc(r, KindChest, "Misery Mire - Main Lobby", []int{0xEA5E})
	bigKeyChest := loc(r, KindChest, "Misery Mire - Big Key Chest", []int{0xEA6D})
	compassChest := loc(r, KindChest, "Misery Mire - Compass Chest", []int{0xEA64})
	loc(r, KindChest, "Misery Mire - Bridge Chest", []int{0xEA61})
	mapChest := loc(r, KindChest, "Misery Mire - Map Chest", []int{0xEA6A})
	spikeChest := loc(r, KindChest, "Misery Mire - Spike Chest", []int{0xE9DA})
	bossDrop := loc(r, KindDrop, "Misery Mire - Boss", []int{0x180158})
	prize := loc(r, KindPrizeCrystal, "Misery Mire - Prize",
		[]int{-1, 0x120A2, 0x53E84, 0x53E85, 0x180057, 0x180077, 0xC703})
	prize.MusicAddresses = []int{0x155B9}
	r.SetPrizeLocation(prize)

	bigChest.Requirement = func(_ *LocationCollection, it *item.Collection) bool { return it.Has1("BigKeyD6") }
	spikeChest.Requirement = func(_ *LocationCollection, it *item.Collection) bool {
		return !w.ConfigBool("region.cantTakeDamage", false) || it.Has1("CaneOfByrna") || it.Has1("Cape")
	}
	mainLobby.Requirement = func(_ *LocationCollection, it *item.Collection) bool {
		return it.Has1("KeyD6") || it.Has1("BigKeyD6")
	}
	mapChest.Requirement = mainLobby.Requirement

	bigKeyReq := func(locs *LocationCollection, it *item.Collection) bool {
		bigKeyD6 := b.item("BigKeyD6")
		hasBKInCompass := compassChest.HasSpecificItem(bigKeyD6) || bigKeyChest.HasSpecificItem(bigKeyD6)
		return it.CanLightTorches() &&
			((hasBKInCompass && it.Has("KeyD6", 2)) || it.Has("KeyD6", 3))
	}
	bigKeyChest.Requirement = bigKeyReq
	compassChest.Requirement = bigKeyReq

	bossDrop.Requirement = func(locs *LocationCollection, it *item.Collection) bool {
		return r.CanEnter(locs, it) &&
			it.Has1("CaneOfSomaria") && it.Has("Lamp", w.ConfigInt("item.require.Lamp", 1)) &&
			it.Has1("BigKeyD6") &&
			r.Boss.CanBeat(it, locs) &&
			wildCompMap(w, bossDrop, it, b, "CompassD6", "MapD6")
	}
	dungeonHelpers{r: r, b: b}.bossDropRules(bossDrop, "CompassD6", "MapD6")

	r.CanEnterFn = func(locs *LocationCollection, it *item.Collection) bool {
		mire := false
		if rr := w.Region("Mire"); rr != nil {
			mire = rr.CanEnter(locs, it)
		}
		basicGate := w.ConfigString("itemPlacement", "") != "basic" ||
			((w.ConfigString("mode.weapons", "") == "swordless" || it.HasSword(2)) && it.HasHealth(12) && (it.HasBottle(2) || it.HasArmor(1)))

		mmMed := w.Locations().Get("Misery Mire Medallion:" + itoaW(w.ID()))
		medOk := false
		if mmMed != nil {
			medOk = ((mmMed.HasSpecificItem(b.item("Bombos")) && it.Has1("Bombos")) ||
				(mmMed.HasSpecificItem(b.item("Ether")) && it.Has1("Ether")) ||
				(mmMed.HasSpecificItem(b.item("Quake")) && it.Has1("Quake"))) &&
				(w.ConfigString("mode.weapons", "") == "swordless" || it.HasSword(1))
		}
		bunnyOk := it.Has1("MoonPearl") ||
			(it.HasABottle() &&
				((it.Has1("BugCatchingNet") && w.ConfigBool("canBunnyRevive", false) &&
					((it.CanLiftDarkRocks() && (it.CanFly(w) || (w.ConfigBool("canBootsClip", false) && it.Has1("PegasusBoots")))) ||
						(w.ConfigBool("canOWYBA", false) && it.Has1("MagicMirror")) ||
						w.ConfigBool("canOneFrameClipOW", false))) ||
					(w.ConfigBool("canOWYBA", false) &&
						((w.ConfigBool("canBootsClip", false) && it.Has1("PegasusBoots")) ||
							w.ConfigBool("canOneFrameClipOW", false) ||
							it.HasBottle(2)))))
		bootsOrHook := (w.ConfigString("itemPlacement", "") != "basic" && it.Has1("PegasusBoots")) || it.Has1("Hookshot")
		return it.Has1("RescueZelda") && basicGate && medOk && bunnyOk && bootsOrHook && it.CanKillMostThings(w, 8) && mire
	}
	r.CanCompFn = func(locs *LocationCollection, it *item.Collection) bool { return bossDrop.CanAccess(it, locs) }
	prize.Requirement = func(locs *LocationCollection, it *item.Collection) bool { return r.CanComplete(locs, it) }
	return r
}

// ----------------------------------------------------------------------
// Standard / PalaceOfDarkness
// PHP: app/Region/Standard/PalaceOfDarkness.php
// ----------------------------------------------------------------------
func newStandardPalaceOfDarkness(b *regionBuilder) *Region {
	w := b.w
	r := NewRegion("Palace of Darkness", w)
	r.MapReveal = 0x0200
	r.RegionItems = b.regionItems(
		"BigKey", "BigKeyD1", "Compass", "CompassD1",
		"Key", "KeyD1", "Map", "MapD1", "Crystal1")
	r.Boss = b.boss("Helmasaur King")

	loc(r, KindChest, "Palace of Darkness - Shooter Room", []int{0xEA5B})
	bigKeyChest := loc(r, KindChest, "Palace of Darkness - Big Key Chest", []int{0xEA37})
	arenaLedge := loc(r, KindChest, "Palace of Darkness - The Arena - Ledge", []int{0xEA3A})
	arenaBridge := loc(r, KindChest, "Palace of Darkness - The Arena - Bridge", []int{0xEA3D})
	stalfos := loc(r, KindChest, "Palace of Darkness - Stalfos Basement", []int{0xEA49})
	mapChest := loc(r, KindChest, "Palace of Darkness - Map Chest", []int{0xEA52})
	bigChest := loc(r, KindBigChest, "Palace of Darkness - Big Chest", []int{0xEA40})
	compassChest := loc(r, KindChest, "Palace of Darkness - Compass Chest", []int{0xEA43})
	hellway := loc(r, KindChest, "Palace of Darkness - Harmless Hellway", []int{0xEA46})
	basementLeft := loc(r, KindChest, "Palace of Darkness - Dark Basement - Left", []int{0xEA4C})
	basementRight := loc(r, KindChest, "Palace of Darkness - Dark Basement - Right", []int{0xEA4F})
	mazeTop := loc(r, KindChest, "Palace of Darkness - Dark Maze - Top", []int{0xEA55})
	mazeBottom := loc(r, KindChest, "Palace of Darkness - Dark Maze - Bottom", []int{0xEA58})
	bossDrop := loc(r, KindDrop, "Palace of Darkness - Boss", []int{0x180153})
	prize := loc(r, KindPrizeCrystal, "Palace of Darkness - Prize",
		[]int{-1, 0x120A1, 0x53E7C, 0x53E7D, 0x180056, 0x180073, 0xC702})
	prize.MusicAddresses = []int{0x155B8}
	r.SetPrizeLocation(prize)

	// keyN reflects PHP's `hammer & arrows & lamp || wildKeys ? 6 : 5` pattern.
	keyHighLow := func(it *item.Collection, high, low int) bool {
		need := low
		if (it.Has1("Hammer") && it.CanShootArrows(w, 1) && it.Has("Lamp", w.ConfigInt("item.require.Lamp", 1))) || w.ConfigBool("region.wildKeys", false) {
			need = high
		}
		return it.Has("KeyD1", need)
	}

	arenaLedge.Requirement = func(_ *LocationCollection, it *item.Collection) bool { return it.CanShootArrows(w, 1) }
	bigKeyChest.Requirement = func(_ *LocationCollection, it *item.Collection) bool {
		if bigKeyChest.HasSpecificItem(b.item("KeyD1")) {
			return it.Has1("KeyD1")
		}
		return keyHighLow(it, 6, 5)
	}
	bigKeyChest.Always = func(itm *item.Item, items *item.Collection) bool {
		return w.ConfigString("accessibility", "") != "locations" && itm == b.item("KeyD1") && items.Has("KeyD1", 5)
	}
	bigKeyChest.FillRule = func(itm *item.Item, _ *LocationCollection, _ *item.Collection) bool {
		return w.ConfigString("accessibility", "") != "locations" || itm != b.item("KeyD1")
	}
	arenaBridge.Requirement = func(_ *LocationCollection, it *item.Collection) bool {
		return it.Has1("KeyD1") || (it.CanShootArrows(w, 1) && it.Has1("Hammer"))
	}
	bigChest.Requirement = func(_ *LocationCollection, it *item.Collection) bool {
		need := 5
		if (it.Has1("Hammer") && it.CanShootArrows(w, 1)) || w.ConfigBool("region.wildKeys", false) {
			need = 6
		}
		return it.Has("Lamp", w.ConfigInt("item.require.Lamp", 1)) && it.Has1("BigKeyD1") && it.Has("KeyD1", need)
	}
	bigChest.FillRule = func(itm *item.Item, _ *LocationCollection, _ *item.Collection) bool {
		return itm != b.item("KeyD1")
	}
	compassChest.Requirement = func(_ *LocationCollection, it *item.Collection) bool { return keyHighLow(it, 4, 3) }
	hellway.Requirement = func(locs *LocationCollection, it *item.Collection) bool {
		if hellway.HasSpecificItem(b.item("KeyD1")) {
			return keyHighLow(it, 4, 3)
		}
		return keyHighLow(it, 6, 5)
	}
	hellway.Always = func(itm *item.Item, items *item.Collection) bool {
		return w.ConfigString("accessibility", "") != "locations" && itm == b.item("KeyD1") && items.Has("KeyD1", 5)
	}
	hellway.FillRule = func(itm *item.Item, _ *LocationCollection, _ *item.Collection) bool {
		return w.ConfigString("accessibility", "") != "locations" || itm != b.item("KeyD1")
	}
	stalfos.Requirement = func(_ *LocationCollection, it *item.Collection) bool {
		return it.Has1("KeyD1") || (it.CanShootArrows(w, 1) && it.Has1("Hammer"))
	}
	basementReq := func(_ *LocationCollection, it *item.Collection) bool {
		return (it.Has("Lamp", w.ConfigInt("item.require.Lamp", 1)) ||
			(w.ConfigString("itemPlacement", "") == "advanced" && it.Has1("FireRod"))) &&
			keyHighLow(it, 4, 3)
	}
	basementLeft.Requirement = basementReq
	basementRight.Requirement = basementReq
	mapChest.Requirement = func(_ *LocationCollection, it *item.Collection) bool { return it.CanShootArrows(w, 1) }
	mazeReq := func(_ *LocationCollection, it *item.Collection) bool {
		need := 5
		if (it.Has1("Hammer") && it.CanShootArrows(w, 1)) || w.ConfigBool("region.wildKeys", false) {
			need = 6
		}
		return it.Has("Lamp", w.ConfigInt("item.require.Lamp", 1)) && it.Has("KeyD1", need)
	}
	mazeTop.Requirement = mazeReq
	mazeBottom.Requirement = mazeReq
	mazeTop.FillRule = func(itm *item.Item, _ *LocationCollection, _ *item.Collection) bool {
		return itm != b.item("KeyD1")
	}
	mazeBottom.FillRule = mazeTop.FillRule

	bossDrop.Requirement = func(locs *LocationCollection, it *item.Collection) bool {
		return r.CanEnter(locs, it) &&
			r.Boss.CanBeat(it, locs) &&
			it.Has1("Hammer") && it.Has("Lamp", w.ConfigInt("item.require.Lamp", 1)) && it.CanShootArrows(w, 1) &&
			it.Has1("BigKeyD1") && it.Has("KeyD1", 6) &&
			wildCompMap(w, bossDrop, it, b, "CompassD1", "MapD1")
	}
	dungeonHelpers{r: r, b: b}.bossDropRules(bossDrop, "CompassD1", "MapD1")

	r.CanEnterFn = func(locs *LocationCollection, it *item.Collection) bool {
		nedw := false
		if rr := w.Region("North East Dark World"); rr != nil {
			nedw = rr.CanEnter(locs, it)
		}
		westDM := false
		if rr := w.Region("West Death Mountain"); rr != nil {
			westDM = rr.CanEnter(locs, it)
		}
		basicGate := w.ConfigString("itemPlacement", "") != "basic" ||
			((w.ConfigString("mode.weapons", "") == "swordless" || it.HasSword(1)) && it.HasHealth(7) && it.HasABottle())
		bunnyOk := it.Has1("MoonPearl") ||
			(w.ConfigBool("canOWYBA", false) && it.HasABottle()) ||
			(w.ConfigBool("canBunnyRevive", false) && it.CanBunnyRevive(w))
		mainRoute := bunnyOk && nedw
		altRoute := w.ConfigBool("canOneFrameClipUW", false) && w.ConfigBool("canDungeonRevive", false) && westDM
		return it.Has1("RescueZelda") && basicGate && (mainRoute || altRoute)
	}
	r.CanCompFn = func(locs *LocationCollection, it *item.Collection) bool { return bossDrop.CanAccess(it, locs) }
	prize.Requirement = func(locs *LocationCollection, it *item.Collection) bool { return r.CanComplete(locs, it) }
	return r
}

// ----------------------------------------------------------------------
// Standard / TurtleRock
// PHP: app/Region/Standard/TurtleRock.php
// ----------------------------------------------------------------------
func newStandardTurtleRock(b *regionBuilder) *Region {
	w := b.w
	r := NewRegion("Turtle Rock", w)
	r.MapReveal = 0x0008
	r.RegionItems = b.regionItems(
		"BigKey", "BigKeyD7", "Compass", "CompassD7",
		"Key", "KeyD7", "Map", "MapD7", "Crystal7")
	r.Boss = b.boss("Trinexx")

	chainChomps := loc(r, KindChest, "Turtle Rock - Chain Chomps", []int{0xEA16})
	compassChest := loc(r, KindChest, "Turtle Rock - Compass Chest", []int{0xEA22})
	rollerLeft := loc(r, KindChest, "Turtle Rock - Roller Room - Left", []int{0xEA1C})
	rollerRight := loc(r, KindChest, "Turtle Rock - Roller Room - Right", []int{0xEA1F})
	bigChest := loc(r, KindBigChest, "Turtle Rock - Big Chest", []int{0xEA19})
	bigKeyChest := loc(r, KindChest, "Turtle Rock - Big Key Chest", []int{0xEA25})
	crystaroller := loc(r, KindChest, "Turtle Rock - Crystaroller Room", []int{0xEA34})
	eyeBL := loc(r, KindChest, "Turtle Rock - Eye Bridge - Bottom Left", []int{0xEA31})
	eyeBR := loc(r, KindChest, "Turtle Rock - Eye Bridge - Bottom Right", []int{0xEA2E})
	eyeTL := loc(r, KindChest, "Turtle Rock - Eye Bridge - Top Left", []int{0xEA2B})
	eyeTR := loc(r, KindChest, "Turtle Rock - Eye Bridge - Top Right", []int{0xEA28})
	bossDrop := loc(r, KindDrop, "Turtle Rock - Boss", []int{0x180159})
	prize := loc(r, KindPrizeCrystal, "Turtle Rock - Prize",
		[]int{-1, 0x120A7, 0x53E80, 0x53E81, 0x18005C, 0x180075, 0xC708})
	prize.MusicAddresses = []int{0x155C7, 0x155A7, 0x155AA, 0x155AB}
	r.SetPrizeLocation(prize)

	upper := func(locs *LocationCollection, it *item.Collection) bool {
		eastDM := false
		if rr := w.Region("East Death Mountain"); rr != nil {
			eastDM = rr.CanEnter(locs, it)
		}
		mmMed := w.Locations().Get("Turtle Rock Medallion:" + itoaW(w.ID()))
		medOk := false
		if mmMed != nil {
			medOk = ((mmMed.HasSpecificItem(b.item("Bombos")) && it.Has1("Bombos")) ||
				(mmMed.HasSpecificItem(b.item("Ether")) && it.Has1("Ether")) ||
				(mmMed.HasSpecificItem(b.item("Quake")) && it.Has1("Quake"))) &&
				(w.ConfigString("mode.weapons", "") == "swordless" || it.HasSword(1))
		}
		bunnyOk := it.Has1("MoonPearl") ||
			(w.ConfigBool("canOWYBA", false) && it.HasABottle() &&
				((w.ConfigBool("canBootsClip", false) && it.Has1("PegasusBoots")) ||
					w.ConfigBool("canOneFrameClipOW", false)))
		return medOk && bunnyOk && it.Has1("CaneOfSomaria") &&
			((it.Has1("Hammer") && it.CanLiftDarkRocks() && eastDM) ||
				((w.ConfigBool("canBootsClip", false) && it.Has1("PegasusBoots")) ||
					w.ConfigBool("canOneFrameClipOW", false)))
	}
	middle := func(locs *LocationCollection, it *item.Collection) bool {
		eastDDM := false
		if rr := w.Region("East Dark World Death Mountain"); rr != nil {
			eastDDM = rr.CanEnter(locs, it)
		}
		gate := ((w.ConfigBool("canMirrorClip", false) && it.Has1("MagicMirror")) &&
			(it.Has1("MoonPearl") || w.ConfigBool("canDungeonRevive", false))) ||
			((w.ConfigBool("canBootsClip", false) && it.Has1("PegasusBoots")) &&
				((w.ConfigBool("canOWYBA", false) && it.HasABottle()) || it.Has1("MoonPearl"))) ||
			(w.ConfigBool("canSuperSpeed", false) && it.Has1("MoonPearl") && it.CanSpinSpeed()) ||
			(w.ConfigBool("canOneFrameClipOW", false) &&
				(w.ConfigBool("canDungeonRevive", false) || it.Has1("MoonPearl") ||
					(w.ConfigBool("canOWYBA", false) && it.HasABottle())))
		damageOk := it.Has1("PegasusBoots") || it.Has1("CaneOfSomaria") || it.Has1("Hookshot") ||
			!w.ConfigBool("region.cantTakeDamage", false) || it.Has1("Cape") || it.Has1("CaneOfByrna")
		return gate && damageOk && eastDDM
	}
	lower := func(locs *LocationCollection, it *item.Collection) bool {
		westDM := false
		if rr := w.Region("West Death Mountain"); rr != nil {
			westDM = rr.CanEnter(locs, it)
		}
		eastDDM := false
		if rr := w.Region("East Dark World Death Mountain"); rr != nil {
			eastDDM = rr.CanEnter(locs, it)
		}
		return w.ConfigBool("canMirrorWrap", false) && it.Has1("MagicMirror") &&
			(it.Has1("MoonPearl") || (w.ConfigBool("canOWYBA", false) && it.HasABottle())) &&
			((((w.ConfigBool("canBootsClip", false) && it.Has1("PegasusBoots")) ||
				w.ConfigBool("canOneFrameClipOW", false)) && westDM) ||
				((w.ConfigBool("canSuperSpeed", false) && it.CanSpinSpeed()) && eastDDM))
	}
	lampReq := func(it *item.Collection) bool { return it.Has("Lamp", w.ConfigInt("item.require.Lamp", 1)) }

	chainChomps.Requirement = func(locs *LocationCollection, it *item.Collection) bool {
		return (upper(locs, it) && it.Has1("KeyD7")) || middle(locs, it) ||
			(lower(locs, it) && lampReq(it) && it.Has1("CaneOfSomaria"))
	}
	rollerLR := func(otherName string) func(*LocationCollection, *item.Collection) bool {
		return func(locs *LocationCollection, it *item.Collection) bool {
			bigKeyInOther := locs.ItemInLocations(b.item("BigKeyD7"), w.ID(),
				[]string{otherName, "Turtle Rock - Compass Chest"}, 1)
			return it.Has1("FireRod") && it.Has1("CaneOfSomaria") &&
				(upper(locs, it) ||
					(middle(locs, it) && ((bigKeyInOther && it.Has("KeyD7", 2)) || it.Has("KeyD7", 4))) ||
					(lower(locs, it) && lampReq(it) && it.Has("KeyD7", 4)))
		}
	}
	rollerLeft.Requirement = rollerLR("Turtle Rock - Roller Room - Right")
	rollerRight.Requirement = rollerLR("Turtle Rock - Roller Room - Left")

	compassChest.Requirement = func(locs *LocationCollection, it *item.Collection) bool {
		bigKeyInRollers := locs.ItemInLocations(b.item("BigKeyD7"), w.ID(),
			[]string{"Turtle Rock - Roller Room - Left", "Turtle Rock - Roller Room - Right"}, 1)
		return it.Has1("CaneOfSomaria") &&
			(upper(locs, it) ||
				(middle(locs, it) && ((bigKeyInRollers && it.Has("KeyD7", 2)) || it.Has("KeyD7", 4))) ||
				(lower(locs, it) && lampReq(it) && it.Has("KeyD7", 4)))
	}
	bigChest.Requirement = func(locs *LocationCollection, it *item.Collection) bool {
		return it.Has1("BigKeyD7") &&
			((upper(locs, it) && it.Has("KeyD7", 2)) ||
				(middle(locs, it) && (it.Has1("Hookshot") || it.Has1("CaneOfSomaria"))) ||
				(lower(locs, it) && lampReq(it) && it.Has1("CaneOfSomaria")))
	}
	bigChest.FillRule = func(itm *item.Item, _ *LocationCollection, _ *item.Collection) bool {
		return itm != b.item("BigKeyD7")
	}

	bigKeyChest.Requirement = func(_ *LocationCollection, it *item.Collection) bool {
		if !bigKeyChest.HasSpecificItem(b.item("BigKeyD7")) && w.ConfigBool("region.wildKeys", false) {
			if bigKeyChest.HasSpecificItem(b.item("KeyD7")) {
				return it.Has("KeyD7", 3)
			}
			return it.Has("KeyD7", 4)
		}
		return it.Has("KeyD7", 2)
	}
	bigKeyChest.Always = func(itm *item.Item, items *item.Collection) bool {
		return w.ConfigString("accessibility", "") != "locations" && itm == b.item("KeyD7") && items.Has("KeyD7", 3)
	}
	bigKeyChest.FillRule = func(itm *item.Item, _ *LocationCollection, _ *item.Collection) bool {
		return w.ConfigString("accessibility", "") != "locations" || itm != b.item("KeyD7")
	}
	crystaroller.Requirement = func(locs *LocationCollection, it *item.Collection) bool {
		return (it.Has1("BigKeyD7") && ((upper(locs, it) && it.Has("KeyD7", 2)) || middle(locs, it))) ||
			(lower(locs, it) && lampReq(it) && it.Has1("CaneOfSomaria"))
	}
	eyeReq := func(locs *LocationCollection, it *item.Collection) bool {
		gate := lower(locs, it) ||
			((upper(locs, it) || middle(locs, it)) && lampReq(it) && it.Has1("CaneOfSomaria") &&
				it.Has1("BigKeyD7") && it.Has("KeyD7", 3))
		damageOk := w.ConfigString("itemPlacement", "") != "basic" || it.Has1("Cape") || it.Has1("CaneOfByrna") ||
			(w.ConfigInt("item.overflow.count.Shield", 3) >= 3 && it.CanBlockLasers())
		return gate && damageOk
	}
	eyeBL.Requirement = eyeReq
	eyeBR.Requirement = eyeReq
	eyeTL.Requirement = eyeReq
	eyeTR.Requirement = eyeReq

	bossDrop.Requirement = func(locs *LocationCollection, it *item.Collection) bool {
		westDM := false
		if rr := w.Region("West Death Mountain"); rr != nil {
			westDM = rr.CanEnter(locs, it)
		}
		eastDDM := false
		if rr := w.Region("East Dark World Death Mountain"); rr != nil {
			eastDDM = rr.CanEnter(locs, it)
		}
		wrapPath := w.ConfigBool("canMirrorWrap", false) && it.Has1("MagicMirror") &&
			(it.Has1("MoonPearl") || (w.ConfigBool("canOWYBA", false) && it.HasABottle())) &&
			((((w.ConfigBool("canBootsClip", false) && it.Has1("PegasusBoots")) ||
				w.ConfigBool("canOneFrameClipOW", false)) && westDM) ||
				((w.ConfigBool("canSuperSpeed", false) && it.CanSpinSpeed()) && eastDDM))
		return r.CanEnter(locs, it) &&
			it.Has("KeyD7", 4) &&
			(wrapPath || lampReq(it)) &&
			it.Has1("BigKeyD7") && it.Has1("CaneOfSomaria") &&
			r.Boss.CanBeat(it, locs) &&
			wildCompMap(w, bossDrop, it, b, "CompassD7", "MapD7")
	}
	dungeonHelpers{r: r, b: b}.bossDropRules(bossDrop, "CompassD7", "MapD7")

	r.CanEnterFn = func(locs *LocationCollection, it *item.Collection) bool {
		basicGate := w.ConfigString("itemPlacement", "") != "basic" ||
			((w.ConfigString("mode.weapons", "") == "swordless" || it.HasSword(2)) && it.HasHealth(12) && (it.HasBottle(2) || it.HasArmor(1)))
		return it.Has1("RescueZelda") && basicGate &&
			(lower(locs, it) || middle(locs, it) || upper(locs, it))
	}
	r.CanCompFn = func(locs *LocationCollection, it *item.Collection) bool { return bossDrop.CanAccess(it, locs) }
	prize.Requirement = func(locs *LocationCollection, it *item.Collection) bool { return r.CanComplete(locs, it) }
	return r
}

// ----------------------------------------------------------------------
// Standard / GanonsTower — multi-boss tower (top/middle/bottom + final).
// PHP: app/Region/Standard/GanonsTower.php
// ----------------------------------------------------------------------
func newStandardGanonsTower(b *regionBuilder) *Region {
	w := b.w
	r := NewRegion("Ganons Tower", w)
	r.RegionItems = b.regionItems(
		"BigKey", "BigKeyA2", "Compass", "CompassA2",
		"Key", "KeyA2", "Map", "MapA2")
	r.Boss = b.boss("Agahnim2")
	r.Bosses["top"] = b.boss("Moldorm")
	r.Bosses["middle"] = b.boss("Lanmolas")
	r.Bosses["bottom"] = b.boss("Armos Knights")
	r.CanPlaceBossLevelFn = func(bb *boss.Boss, level string) bool {
		if w.ConfigString("mode.weapons", "") == "swordless" && bb.Name == "Kholdstare" {
			return false
		}
		switch level {
		case "top":
			switch bb.Name {
			case "Agahnim", "Agahnim2", "Armos Knights", "Arrghus", "Blind", "Ganon", "Lanmolas", "Trinexx":
				return false
			}
			return true
		case "middle":
			switch bb.Name {
			case "Agahnim", "Agahnim2", "Blind", "Ganon":
				return false
			}
			return true
		}
		switch bb.Name {
		case "Agahnim", "Agahnim2", "Ganon":
			return false
		}
		return true
	}

	bobsTorch := loc(r, KindDash, "Ganon's Tower - Bob's Torch", []int{0x180161})
	dmTL := loc(r, KindChest, "Ganon's Tower - DMs Room - Top Left", []int{0xEAB8})
	dmTR := loc(r, KindChest, "Ganon's Tower - DMs Room - Top Right", []int{0xEABB})
	dmBL := loc(r, KindChest, "Ganon's Tower - DMs Room - Bottom Left", []int{0xEABE})
	dmBR := loc(r, KindChest, "Ganon's Tower - DMs Room - Bottom Right", []int{0xEAC1})
	randTL := loc(r, KindChest, "Ganon's Tower - Randomizer Room - Top Left", []int{0xEAC4})
	randTR := loc(r, KindChest, "Ganon's Tower - Randomizer Room - Top Right", []int{0xEAC7})
	randBL := loc(r, KindChest, "Ganon's Tower - Randomizer Room - Bottom Left", []int{0xEACA})
	randBR := loc(r, KindChest, "Ganon's Tower - Randomizer Room - Bottom Right", []int{0xEACD})
	firesnake := loc(r, KindChest, "Ganon's Tower - Firesnake Room", []int{0xEAD0})
	mapChest := loc(r, KindChest, "Ganon's Tower - Map Chest", []int{0xEAD3})
	bigChest := loc(r, KindBigChest, "Ganon's Tower - Big Chest", []int{0xEAD6})
	loc(r, KindChest, "Ganon's Tower - Hope Room - Left", []int{0xEAD9})
	loc(r, KindChest, "Ganon's Tower - Hope Room - Right", []int{0xEADC})
	bobsChest := loc(r, KindChest, "Ganon's Tower - Bob's Chest", []int{0xEADF})
	tileRoom := loc(r, KindChest, "Ganon's Tower - Tile Room", []int{0xEAE2})
	compTL := loc(r, KindChest, "Ganon's Tower - Compass Room - Top Left", []int{0xEAE5})
	compTR := loc(r, KindChest, "Ganon's Tower - Compass Room - Top Right", []int{0xEAE8})
	compBL := loc(r, KindChest, "Ganon's Tower - Compass Room - Bottom Left", []int{0xEAEB})
	compBR := loc(r, KindChest, "Ganon's Tower - Compass Room - Bottom Right", []int{0xEAEE})
	bigKeyChest := loc(r, KindChest, "Ganon's Tower - Big Key Chest", []int{0xEAF1})
	bigKeyRoomL := loc(r, KindChest, "Ganon's Tower - Big Key Room - Left", []int{0xEAF4})
	bigKeyRoomR := loc(r, KindChest, "Ganon's Tower - Big Key Room - Right", []int{0xEAF7})
	miniHelmaL := loc(r, KindChest, "Ganon's Tower - Mini Helmasaur Room - Left", []int{0xEAFD})
	miniHelmaR := loc(r, KindChest, "Ganon's Tower - Mini Helmasaur Room - Right", []int{0xEB00})
	preMoldorm := loc(r, KindChest, "Ganon's Tower - Pre-Moldorm Chest", []int{0xEB03})
	moldormChest := loc(r, KindChest, "Ganon's Tower - Moldorm Chest", []int{0xEB06})
	prize := loc(r, KindPrizeEvent, "Agahnim 2", nil)
	_ = prize.SetItem(b.item("DefeatAgahnim2"))
	r.SetPrizeLocation(prize)

	hh := func(it *item.Collection) bool { return it.Has1("Hammer") && it.Has1("Hookshot") }

	bobsTorch.Requirement = func(_ *LocationCollection, it *item.Collection) bool { return it.Has1("PegasusBoots") }
	dmTL.Requirement = func(_ *LocationCollection, it *item.Collection) bool { return hh(it) }
	dmTR.Requirement = dmTL.Requirement
	dmBL.Requirement = dmTL.Requirement
	dmBR.Requirement = dmTL.Requirement

	randReq := func(others []string) RequirementFunc {
		return func(locs *LocationCollection, it *item.Collection) bool {
			has := locs.ItemInLocations(b.item("BigKeyA2"), w.ID(), others, 1)
			return hh(it) && ((has && it.Has("KeyA2", 3)) || it.Has("KeyA2", 4))
		}
	}
	randTL.Requirement = randReq([]string{"Ganon's Tower - Randomizer Room - Top Right", "Ganon's Tower - Randomizer Room - Bottom Left", "Ganon's Tower - Randomizer Room - Bottom Right"})
	randTR.Requirement = randReq([]string{"Ganon's Tower - Randomizer Room - Top Left", "Ganon's Tower - Randomizer Room - Bottom Left", "Ganon's Tower - Randomizer Room - Bottom Right"})
	randBL.Requirement = randReq([]string{"Ganon's Tower - Randomizer Room - Top Right", "Ganon's Tower - Randomizer Room - Top Left", "Ganon's Tower - Randomizer Room - Bottom Right"})
	randBR.Requirement = randReq([]string{"Ganon's Tower - Randomizer Room - Top Right", "Ganon's Tower - Randomizer Room - Top Left", "Ganon's Tower - Randomizer Room - Bottom Left"})

	firesnake.Requirement = func(locs *LocationCollection, it *item.Collection) bool {
		others := []string{
			"Ganon's Tower - Randomizer Room - Top Right",
			"Ganon's Tower - Randomizer Room - Top Left",
			"Ganon's Tower - Randomizer Room - Bottom Left",
			"Ganon's Tower - Randomizer Room - Bottom Right",
		}
		bkInOthers := locs.ItemInLocations(b.item("BigKeyA2"), w.ID(), others, 1)
		return hh(it) &&
			(((bkInOthers || firesnake.HasSpecificItem(b.item("KeyA2"))) && it.Has("KeyA2", 2)) ||
				it.Has("KeyA2", 3))
	}
	mapChest.Requirement = func(_ *LocationCollection, it *item.Collection) bool {
		mover := it.Has1("Hookshot") || (w.ConfigString("itemPlacement", "") != "basic" && it.Has1("PegasusBoots"))
		placedHere := mapChest.item
		special := placedHere != nil &&
			(placedHere.FullName() == b.item("BigKeyA2").FullName() ||
				placedHere.FullName() == b.item("KeyA2").FullName())
		need := 4
		if special {
			need = 3
		}
		return it.Has1("Hammer") && mover && it.Has("KeyA2", need)
	}
	mapChest.Always = func(itm *item.Item, items *item.Collection) bool {
		return w.ConfigString("accessibility", "") != "locations" && itm == b.item("KeyA2") && items.Has("KeyA2", 3)
	}
	mapChest.FillRule = func(itm *item.Item, _ *LocationCollection, _ *item.Collection) bool {
		return w.ConfigString("accessibility", "") != "locations" || itm != b.item("KeyA2")
	}

	hhOrFireRodSomaria := func(it *item.Collection) bool {
		return hh(it) || (it.Has1("FireRod") && it.Has1("CaneOfSomaria"))
	}

	bigChest.Requirement = func(_ *LocationCollection, it *item.Collection) bool {
		return it.Has1("BigKeyA2") && it.Has("KeyA2", 3) && hhOrFireRodSomaria(it)
	}
	bigChest.FillRule = func(itm *item.Item, _ *LocationCollection, _ *item.Collection) bool {
		return itm != b.item("BigKeyA2")
	}
	bobsChest.Requirement = func(_ *LocationCollection, it *item.Collection) bool {
		return hhOrFireRodSomaria(it) && it.Has("KeyA2", 3) &&
			(w.ConfigString("itemPlacement", "") != "basic" ||
				(it.Has1("FireRod") || (it.Has1("Ether") && it.HasSword(1))))
	}
	tileRoom.Requirement = func(_ *LocationCollection, it *item.Collection) bool { return it.Has1("CaneOfSomaria") }

	compReq := func(others []string) RequirementFunc {
		return func(locs *LocationCollection, it *item.Collection) bool {
			has := locs.ItemInLocations(b.item("BigKeyA2"), w.ID(), others, 1)
			return it.Has1("FireRod") && it.Has1("CaneOfSomaria") &&
				((has && it.Has("KeyA2", 3)) || it.Has("KeyA2", 4))
		}
	}
	compTL.Requirement = compReq([]string{"Ganon's Tower - Compass Room - Top Right", "Ganon's Tower - Compass Room - Bottom Left", "Ganon's Tower - Compass Room - Bottom Right"})
	compTR.Requirement = compReq([]string{"Ganon's Tower - Compass Room - Top Left", "Ganon's Tower - Compass Room - Bottom Left", "Ganon's Tower - Compass Room - Bottom Right"})
	compBL.Requirement = compReq([]string{"Ganon's Tower - Compass Room - Top Right", "Ganon's Tower - Compass Room - Top Left", "Ganon's Tower - Compass Room - Bottom Right"})
	compBR.Requirement = compReq([]string{"Ganon's Tower - Compass Room - Top Right", "Ganon's Tower - Compass Room - Top Left", "Ganon's Tower - Compass Room - Bottom Left"})

	bottomBossReq := func(locs *LocationCollection, it *item.Collection) bool {
		return hhOrFireRodSomaria(it) && it.Has("KeyA2", 3) && r.Bosses["bottom"].CanBeat(it, locs)
	}
	bigKeyChest.Requirement = bottomBossReq
	bigKeyRoomL.Requirement = bottomBossReq
	bigKeyRoomR.Requirement = bottomBossReq

	miniReq := func(locs *LocationCollection, it *item.Collection) bool {
		return it.CanShootArrows(w, 1) && it.CanLightTorches() &&
			it.Has1("BigKeyA2") && it.Has("KeyA2", 3) &&
			r.Bosses["middle"].CanBeat(it, locs)
	}
	miniHelmaL.Requirement = miniReq
	miniHelmaR.Requirement = miniReq
	preMoldorm.Requirement = miniReq
	for _, l := range []*Location{miniHelmaL, miniHelmaR, preMoldorm} {
		l.FillRule = func(itm *item.Item, _ *LocationCollection, _ *item.Collection) bool {
			return itm != b.item("BigKeyA2")
		}
	}
	moldormChest.Requirement = func(locs *LocationCollection, it *item.Collection) bool {
		return it.Has1("Hookshot") && it.CanShootArrows(w, 1) && it.CanLightTorches() &&
			it.Has1("BigKeyA2") && it.Has("KeyA2", 4) &&
			r.Bosses["middle"].CanBeat(it, locs) &&
			r.Bosses["top"].CanBeat(it, locs)
	}
	moldormChest.FillRule = func(itm *item.Item, _ *LocationCollection, _ *item.Collection) bool {
		return itm != b.item("KeyA2") && itm != b.item("BigKeyA2")
	}

	r.CanCompFn = func(locs *LocationCollection, it *item.Collection) bool {
		return r.CanEnter(locs, it) &&
			moldormChest.CanAccess(it, locs) &&
			r.Boss.CanBeat(it, locs)
	}
	prize.Requirement = func(locs *LocationCollection, it *item.Collection) bool { return r.CanComplete(locs, it) }

	r.CanEnterFn = func(locs *LocationCollection, it *item.Collection) bool {
		basicGate := w.ConfigString("itemPlacement", "") != "basic" ||
			((w.ConfigString("mode.weapons", "") == "swordless" || it.HasSword(2)) && it.HasHealth(12) && (it.HasBottle(2) || it.HasArmor(1)))
		bottleOrPearl := it.Has1("MoonPearl") || (w.ConfigBool("canOWYBA", false) && it.HasABottle())
		crystalCount := 0
		for i := 1; i <= 7; i++ {
			if it.Has1("Crystal" + itoaW(i)) {
				crystalCount++
			}
		}
		req := w.ConfigInt("crystals.tower", 7)
		eastDDM := false
		if rr := w.Region("East Dark World Death Mountain"); rr != nil {
			eastDDM = rr.CanEnter(locs, it)
		}
		westDDM := false
		if rr := w.Region("West Dark World Death Mountain"); rr != nil {
			westDDM = rr.CanEnter(locs, it)
		}
		mainPath := bottleOrPearl &&
			((crystalCount >= req && eastDDM) ||
				(((w.ConfigBool("canBootsClip", false) && it.Has1("PegasusBoots")) ||
					(w.ConfigBool("canSuperSpeed", false) && it.Has1("PegasusBoots") && it.Has1("Hookshot"))) && westDDM))
		altPath := w.ConfigBool("canOneFrameClipOW", false) &&
			(w.ConfigBool("canDungeonRevive", false) || it.Has1("MoonPearl") ||
				(w.ConfigBool("canOWYBA", false) && it.HasABottle()))
		return it.Has1("RescueZelda") && basicGate && (mainPath || altPath)
	}
	return r
}

// itoaW: package-local stringer for the locations lookup hack above.
// Use a tiny helper to avoid importing strconv.
func itoaW(n int) string {
	if n == 0 {
		return "0"
	}
	neg := n < 0
	if neg {
		n = -n
	}
	var buf [20]byte
	i := len(buf)
	for n > 0 {
		i--
		buf[i] = byte('0' + n%10)
		n /= 10
	}
	if neg {
		i--
		buf[i] = '-'
	}
	return string(buf[i:])
}
