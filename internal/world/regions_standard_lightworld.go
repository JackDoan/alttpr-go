package world

import (
	"github.com/JackDoan/alttpr-go/internal/item"
)

// ----------------------------------------------------------------------
// Standard / LightWorld/DeathMountain/West
// PHP: app/Region/Standard/LightWorld/DeathMountain/West.php
// ----------------------------------------------------------------------
func newStandardLWWestDM(b *regionBuilder) *Region {
	w := b.w
	r := NewRegion("West Death Mountain", w)

	oldMan := loc(r, KindNpc, "Old Man", []int{0xF69FA})
	loc(r, KindStanding, "Spectacle Rock Cave", []int{0x180002})
	ether := loc(r, KindEtherTablet, "Ether Tablet", []int{0x180016})
	specRock := loc(r, KindStanding, "Spectacle Rock", []int{0x180140})

	oldMan.Requirement = func(_ *LocationCollection, it *item.Collection) bool {
		return it.Has("Lamp", w.ConfigInt("item.require.Lamp", 1))
	}
	ether.Requirement = func(locs *LocationCollection, it *item.Collection) bool {
		hera := false
		if rr := w.Region("Tower of Hera"); rr != nil {
			hera = rr.CanEnter(locs, it)
		}
		return it.Has1("BookOfMudora") &&
			(it.HasSword(2) || (w.ConfigString("mode.weapons", "") == "swordless" && it.Has1("Hammer"))) &&
			hera
	}
	specRock.Requirement = func(_ *LocationCollection, it *item.Collection) bool {
		return it.Has1("MagicMirror") ||
			(w.ConfigBool("canBootsClip", false) && it.Has1("PegasusBoots")) ||
			w.ConfigBool("canOneFrameClipOW", false)
	}
	r.CanEnterFn = func(_ *LocationCollection, it *item.Collection) bool {
		return it.Has1("RescueZelda") &&
			(it.CanFly(w) ||
				w.ConfigBool("canOneFrameClipOW", false) ||
				(w.ConfigBool("canOWYBA", false) && it.HasABottle()) ||
				(w.ConfigBool("canBootsClip", false) && it.Has1("PegasusBoots")) ||
				(w.ConfigBool("canSuperSpeed", false) && it.CanSpinSpeed()) ||
				(it.CanLiftRocks() && it.Has("Lamp", w.ConfigInt("item.require.Lamp", 1))))
	}
	return r
}

// ----------------------------------------------------------------------
// Standard / LightWorld/DeathMountain/East
// PHP: app/Region/Standard/LightWorld/DeathMountain/East.php
// ----------------------------------------------------------------------
func newStandardLWEastDM(b *regionBuilder) *Region {
	w := b.w
	r := NewRegion("East Death Mountain", w)

	loc(r, KindChest, "Spiral Cave", []int{0xE9BF})
	mimic := loc(r, KindChest, "Mimic Cave", []int{0xE9C5})
	loc(r, KindChest, "Paradox Cave Lower - Far Left", []int{0xEB2A})
	loc(r, KindChest, "Paradox Cave Lower - Left", []int{0xEB2D})
	loc(r, KindChest, "Paradox Cave Lower - Right", []int{0xEB30})
	loc(r, KindChest, "Paradox Cave Lower - Far Right", []int{0xEB33})
	loc(r, KindChest, "Paradox Cave Lower - Middle", []int{0xEB36})
	loc(r, KindChest, "Paradox Cave Upper - Left", []int{0xEB39})
	loc(r, KindChest, "Paradox Cave Upper - Right", []int{0xEB3C})
	floating := loc(r, KindStanding, "Floating Island", []int{0x180141})

	dmShop := &Shop{Name: "Light World Death Mountain Shop", Kind: ShopRegular, Config: 0x43, Shopkeeper: 0xA0, RoomID: 0x00FF, DoorID: 0x00, Region: r}
	dmShop.AddInventory(0, b.item("RedPotion"), 150)
	dmShop.AddInventory(1, b.item("Heart"), 10)
	dmShop.AddInventory(2, b.item("TenBombs"), 50)
	dmShop.Requirement = func(_ *LocationCollection, it *item.Collection) bool { return it.CanBombThings() }
	r.Shops.Add(dmShop)
	r.Shops.Add(&Shop{Name: "Hookshot Fairy", Kind: ShopTakeAny, Config: 0x83, Shopkeeper: 0xA0, RoomID: 0x0112, DoorID: 0x50, Region: r,
		Writes: map[int][]int{0xDBBC2: {0x58}}})

	mimic.Requirement = func(locs *LocationCollection, it *item.Collection) bool {
		eastDDM := false
		if rr := w.Region("East Dark World Death Mountain"); rr != nil {
			eastDDM = rr.CanEnter(locs, it)
		}
		tr := false
		if rr := w.Region("Turtle Rock"); rr != nil {
			tr = rr.CanEnter(locs, it)
		}
		return it.Has1("Hammer") && it.Has1("MagicMirror") &&
			(w.ConfigBool("canMirrorClip", false) ||
				(w.ConfigBool("canBootsClip", false) && it.Has1("PegasusBoots") &&
					(it.Has1("MoonPearl") || (w.ConfigBool("canOWYBA", false) && it.Has1("Bottle")))) ||
				(w.ConfigBool("canSuperSpeed", false) && it.CanSpinSpeed() && it.Has1("MoonPearl") && eastDDM) ||
				w.ConfigBool("canOneFrameClipOW", false) ||
				(it.Has("KeyD7", 2) && tr))
	}
	floating.Requirement = func(locs *LocationCollection, it *item.Collection) bool {
		eastDDM := false
		if rr := w.Region("East Dark World Death Mountain"); rr != nil {
			eastDDM = rr.CanEnter(locs, it)
		}
		return (w.ConfigBool("canBootsClip", false) && it.Has1("PegasusBoots")) ||
			(w.ConfigBool("canOWYBA", false) && it.Has1("Bottle")) ||
			w.ConfigBool("canOneFrameClipOW", false) ||
			(it.Has1("MagicMirror") &&
				((it.Has1("MoonPearl") && it.CanBombThings() && it.CanLiftRocks()) ||
					w.ConfigBool("canMirrorWrap", false)) && eastDDM)
	}
	r.CanEnterFn = func(locs *LocationCollection, it *item.Collection) bool {
		westDM := false
		if rr := w.Region("West Death Mountain"); rr != nil {
			westDM = rr.CanEnter(locs, it)
		}
		hera := false
		if rr := w.Region("Tower of Hera"); rr != nil {
			hera = rr.CanEnter(locs, it)
		}
		return it.Has1("RescueZelda") &&
			(w.ConfigBool("canOneFrameClipOW", false) ||
				(w.ConfigBool("canBootsClip", false) && it.Has1("PegasusBoots")) ||
				(w.ConfigBool("canSuperSpeed", false) && it.CanSpinSpeed()) ||
				((((w.ConfigBool("canMirrorClip", false) || w.ConfigBool("canMirrorWrap", false)) && it.Has1("MagicMirror")) ||
					it.Has1("Hookshot")) && westDM) ||
				(it.Has1("Hammer") && hera))
	}
	return r
}

// ----------------------------------------------------------------------
// Standard / LightWorld/NorthEast
// PHP: app/Region/Standard/LightWorld/NorthEast.php
// ----------------------------------------------------------------------
func newStandardLWNorthEast(b *regionBuilder) *Region {
	w := b.w
	r := NewRegion("North East Light World", w)

	loc(r, KindChest, "Sahasrahla's Hut - Left", []int{0xEA82})
	loc(r, KindChest, "Sahasrahla's Hut - Middle", []int{0xEA85})
	loc(r, KindChest, "Sahasrahla's Hut - Right", []int{0xEA88})
	saha := loc(r, KindNpc, "Sahasrahla", []int{0x2F1FC})
	zora := loc(r, KindZora, "King Zora", []int{0xEE1C3})
	potShop := loc(r, KindWitch, "Potion Shop", []int{0x180014})
	zoraLedge := loc(r, KindStanding, "Zora's Ledge", []int{0x180149})
	wfLeft := loc(r, KindChest, "Waterfall Fairy - Left", []int{0xE9B0})
	wfRight := loc(r, KindChest, "Waterfall Fairy - Right", []int{0xE9D1})

	r.Shops.Add(&Shop{Name: "Long Fairy Cave", Kind: ShopTakeAny, Config: 0x83, Shopkeeper: 0xA0, RoomID: 0x0112, DoorID: 0x55, Region: r,
		Writes: map[int][]int{0xDBBC7: {0x58}}})
	r.Shops.Add(&Shop{Name: "Lake Hylia Fairy", Kind: ShopTakeAny, Config: 0x83, Shopkeeper: 0xA0, RoomID: 0x0112, DoorID: 0x5E, Region: r,
		Writes: map[int][]int{0xDBBD0: {0x58}}})

	saha.Requirement = func(_ *LocationCollection, it *item.Collection) bool { return it.Has1("PendantOfCourage") }
	zora.Requirement = func(_ *LocationCollection, it *item.Collection) bool {
		return w.ConfigBool("canFakeFlipper", false) ||
			((w.ConfigBool("canWaterWalk", false) || w.ConfigBool("canBootsClip", false)) && it.Has1("PegasusBoots")) ||
			w.ConfigBool("canOneFrameClipOW", false) ||
			it.CanLiftRocks() || it.Has1("Flippers")
	}
	potShop.Requirement = func(_ *LocationCollection, it *item.Collection) bool { return it.Has1("Mushroom") }
	zoraLedge.Requirement = func(_ *LocationCollection, it *item.Collection) bool {
		return it.Has1("Flippers") ||
			(it.Has1("PegasusBoots") && w.ConfigBool("canWaterWalk", false) &&
				(w.ConfigBool("canFakeFlipper", false) ||
					(w.ConfigBool("canSuperSpeed", false) && it.CanSpinSpeed()) ||
					w.ConfigBool("canBootsClip", false) ||
					w.ConfigBool("canOneFrameClipOW", false)))
	}
	wfReq := func(_ *LocationCollection, it *item.Collection) bool {
		return it.Has1("Flippers") ||
			(w.ConfigBool("canWaterWalk", false) && (it.Has1("PegasusBoots") ||
				(it.Has1("MoonPearl") && w.ConfigBool("canFakeFlipper", false))))
	}
	wfLeft.Requirement = wfReq
	wfRight.Requirement = wfReq

	r.CanEnterFn = func(_ *LocationCollection, it *item.Collection) bool { return it.Has1("RescueZelda") }
	return r
}

// ----------------------------------------------------------------------
// Standard / LightWorld/NorthWest
// PHP: app/Region/Standard/LightWorld/NorthWest.php
// ----------------------------------------------------------------------
func newStandardLWNorthWest(b *regionBuilder) *Region {
	w := b.w
	r := NewRegion("North West Light World", w)

	masterSword := loc(r, KindPedestal, "Master Sword Pedestal", []int{0x289B0})
	kingsTomb := loc(r, KindChest, "King's Tomb", []int{0xE97A})
	loc(r, KindChest, "Kakariko Tavern", []int{0xE9CE})
	loc(r, KindChest, "Chicken House", []int{0xE9E9})
	loc(r, KindChest, "Kakariko Well - Top", []int{0xEA8E})
	loc(r, KindChest, "Kakariko Well - Left", []int{0xEA91})
	loc(r, KindChest, "Kakariko Well - Middle", []int{0xEA94})
	loc(r, KindChest, "Kakariko Well - Right", []int{0xEA97})
	loc(r, KindChest, "Kakariko Well - Bottom", []int{0xEA9A})
	loc(r, KindChest, "Blind's Hideout - Top", []int{0xEB0F})
	loc(r, KindChest, "Blind's Hideout - Left", []int{0xEB12})
	loc(r, KindChest, "Blind's Hideout - Right", []int{0xEB15})
	loc(r, KindChest, "Blind's Hideout - Far Left", []int{0xEB18})
	loc(r, KindChest, "Blind's Hideout - Far Right", []int{0xEB1B})
	pegasusRocks := loc(r, KindChest, "Pegasus Rocks", []int{0xEB3F})
	loc(r, KindNpc, "Bottle Merchant", []int{0x2EB18})
	magicBat := loc(r, KindNpc, "Magic Bat", []int{0x180015})
	sickKid := loc(r, KindBugCatchingKid, "Sick Kid", []int{0x339CF})
	loc(r, KindStanding, "Lost Woods Hideout", []int{0x180000})
	lumberjack := loc(r, KindStanding, "Lumberjack Tree", []int{0x180001})
	graveyardLedge := loc(r, KindStanding, "Graveyard Ledge", []int{0x180004})
	loc(r, KindStanding, "Mushroom", []int{0x180013})

	kakaShop := &Shop{Name: "Light World Kakariko Shop", Kind: ShopRegular, Config: 0x03, Shopkeeper: 0xA0, RoomID: 0x011F, DoorID: 0x46, Region: r}
	kakaShop.AddInventory(0, b.item("RedPotion"), 150)
	kakaShop.AddInventory(1, b.item("Heart"), 10)
	kakaShop.AddInventory(2, b.item("TenBombs"), 50)
	r.Shops.Add(kakaShop)
	for _, s := range []*Shop{
		{Name: "Fortune Teller (Light)", Kind: ShopTakeAny, Config: 0x83, Shopkeeper: 0xA0, RoomID: 0x011F, DoorID: 0x65, Region: r, Writes: map[int][]int{0xDBBD7: {0x46}}},
		{Name: "Bush Covered House", Kind: ShopTakeAny, Config: 0x83, Shopkeeper: 0xA0, RoomID: 0x011F, DoorID: 0x44, Region: r, Writes: map[int][]int{0xDBBB6: {0x46}}},
		{Name: "Lost Woods Gamble", Kind: ShopTakeAny, Config: 0x83, Shopkeeper: 0xA0, RoomID: 0x0112, DoorID: 0x3C, Region: r, Writes: map[int][]int{0xDBBAE: {0x58}}},
		{Name: "Lumberjack House", Kind: ShopTakeAny, Config: 0x83, Shopkeeper: 0xA0, RoomID: 0x011F, DoorID: 0x76, Region: r, Writes: map[int][]int{0xDBBE8: {0x46}}},
		{Name: "Snitch Lady East", Kind: ShopTakeAny, Config: 0x83, Shopkeeper: 0xA0, RoomID: 0x011F, DoorID: 0x3E, Region: r, Writes: map[int][]int{0xDBBB0: {0x46}}},
		{Name: "Snitch Lady West", Kind: ShopTakeAny, Config: 0x83, Shopkeeper: 0xA0, RoomID: 0x011F, DoorID: 0x3F, Region: r, Writes: map[int][]int{0xDBBB1: {0x46}}},
		{Name: "Bomb Hut", Kind: ShopTakeAny, Config: 0x83, Shopkeeper: 0xA0, RoomID: 0x011F, DoorID: 0x4A, Region: r, Writes: map[int][]int{0xDBBBC: {0x46}}},
	} {
		r.Shops.Add(s)
	}
	r.Shops.Get("Bomb Hut").Requirement = func(_ *LocationCollection, it *item.Collection) bool { return it.CanBombThings() }

	masterSword.Requirement = func(_ *LocationCollection, it *item.Collection) bool {
		return (w.ConfigString("itemPlacement", "") != "basic" || it.Has1("BookOfMudora")) &&
			it.Has1("PendantOfPower") && it.Has1("PendantOfWisdom") && it.Has1("PendantOfCourage")
	}
	mirrorNWDW := func(locs *LocationCollection, it *item.Collection) bool {
		nwdw := false
		if rr := w.Region("North West Dark World"); rr != nil {
			nwdw = rr.CanEnter(locs, it)
		}
		return it.Has1("MagicMirror") && nwdw &&
			(it.Has1("MoonPearl") || (it.HasABottle() && w.ConfigBool("canOWYBA", false)) ||
				(w.ConfigBool("canBunnyRevive", false) && it.CanBunnyRevive(w)))
	}
	kingsTomb.Requirement = func(locs *LocationCollection, it *item.Collection) bool {
		return it.Has1("PegasusBoots") &&
			(w.ConfigBool("canBootsClip", false) || it.CanLiftDarkRocks() ||
				w.ConfigBool("canOneFrameClipOW", false) ||
				(w.ConfigBool("canSuperSpeed", false) && it.CanSpinSpeed()) ||
				mirrorNWDW(locs, it))
	}
	pegasusRocks.Requirement = func(_ *LocationCollection, it *item.Collection) bool { return it.Has1("PegasusBoots") }
	magicBat.Requirement = func(locs *LocationCollection, it *item.Collection) bool {
		nwdw := false
		if rr := w.Region("North West Dark World"); rr != nil {
			nwdw = rr.CanEnter(locs, it)
		}
		nedw := false
		if rr := w.Region("North East Dark World"); rr != nil {
			nedw = rr.CanEnter(locs, it)
		}
		alt := it.Has1("MagicMirror") &&
			((w.ConfigBool("canMirrorWrap", false) && nwdw) ||
				((it.Has1("MoonPearl") || (it.HasABottle() && w.ConfigBool("canOWYBA", false)) ||
					(w.ConfigBool("canBunnyRevive", false) && it.CanBunnyRevive(w))) &&
					((it.CanLiftDarkRocks() && nwdw) ||
						(w.ConfigBool("canSuperSpeed", false) && it.CanSpinSpeed() &&
							(it.Has1("Flippers") || w.ConfigBool("canFakeFlipper", false)) && nedw))))
		clip := ((w.ConfigBool("canBootsClip", false) && it.Has1("PegasusBoots")) ||
			w.ConfigBool("canOneFrameClipOW", false)) && w.ConfigBool("canTransitionWrapped", false) &&
			(it.Has1("Flippers") || w.ConfigBool("canFakeFlipper", false))
		return it.Has1("Powder") && (it.Has1("Hammer") || clip || alt)
	}
	sickKid.Requirement = func(_ *LocationCollection, it *item.Collection) bool { return it.HasABottle() }
	lumberjack.Requirement = func(_ *LocationCollection, it *item.Collection) bool {
		return it.Has1("DefeatAgahnim") && it.Has1("PegasusBoots")
	}
	graveyardLedge.Requirement = func(locs *LocationCollection, it *item.Collection) bool {
		return (w.ConfigBool("canBootsClip", false) && it.Has1("PegasusBoots")) ||
			(w.ConfigBool("canSuperSpeed", false) && it.CanSpinSpeed()) ||
			w.ConfigBool("canOneFrameClipOW", false) ||
			mirrorNWDW(locs, it)
	}

	r.CanEnterFn = func(_ *LocationCollection, it *item.Collection) bool { return it.Has1("RescueZelda") }
	return r
}

// ----------------------------------------------------------------------
// Standard / LightWorld/South
// PHP: app/Region/Standard/LightWorld/South.php
// ----------------------------------------------------------------------
func newStandardLWSouth(b *regionBuilder) *Region {
	w := b.w
	r := NewRegion("South Light World", w)

	loc(r, KindChest, "Floodgate Chest", []int{0xE98C})
	loc(r, KindChest, "Link's House", []int{0xE9BC})
	aginah := loc(r, KindChest, "Aginah's Cave", []int{0xE9F2})
	mmFL := loc(r, KindChest, "Mini Moldorm Cave - Far Left", []int{0xEB42})
	mmL := loc(r, KindChest, "Mini Moldorm Cave - Left", []int{0xEB45})
	mmR := loc(r, KindChest, "Mini Moldorm Cave - Right", []int{0xEB48})
	mmFR := loc(r, KindChest, "Mini Moldorm Cave - Far Right", []int{0xEB4B})
	loc(r, KindChest, "Ice Rod Cave", []int{0xEB4E})
	hobo := loc(r, KindNpc, "Hobo", []int{0x33E7D})
	bombos := loc(r, KindBombosTablet, "Bombos Tablet", []int{0x180017})
	cave45 := loc(r, KindStanding, "Cave 45", []int{0x180003})
	checker := loc(r, KindStanding, "Checkerboard Cave", []int{0x180005})
	mmNpc := loc(r, KindNpc, "Mini Moldorm Cave - NPC", []int{0x180010})
	library := loc(r, KindDash, "Library", []int{0x180012})
	loc(r, KindStanding, "Maze Race", []int{0x180142})
	desertLedge := loc(r, KindStanding, "Desert Ledge", []int{0x180143})
	lakeIsland := loc(r, KindStanding, "Lake Hylia Island", []int{0x180144})
	loc(r, KindStanding, "Sunken Treasure", []int{0x180145})
	flute := loc(r, KindHauntedGrove, "Flute Spot", []int{0x18014A})

	lhShop := &Shop{Name: "Light World Lake Hylia Shop", Kind: ShopRegular, Config: 0x03, Shopkeeper: 0xA0, RoomID: 0x0112, DoorID: 0x58, Region: r}
	lhShop.AddInventory(0, b.item("RedPotion"), 150)
	lhShop.AddInventory(1, b.item("Heart"), 10)
	lhShop.AddInventory(2, b.item("TenBombs"), 50)
	r.Shops.Add(lhShop)

	cap := &Shop{Name: "Capacity Upgrade", Kind: ShopUpgrade, Config: 0x12, Shopkeeper: 0x04, RoomID: 0x0115, DoorID: 0x5D, Region: r}
	cap.AddInventory(0, b.item("BombUpgrade5"), 100)
	cap.Inventory[0].Max = 7
	cap.AddInventory(1, b.item("ArrowUpgrade5"), 100)
	cap.Inventory[1].Max = 7
	cap.Requirement = func(_ *LocationCollection, it *item.Collection) bool {
		return (w.ConfigBool("canWaterWalk", false) && it.Has1("PegasusBoots")) ||
			w.ConfigBool("canFakeFlipper", false) || it.Has1("Flippers")
	}
	r.Shops.Add(cap)
	for _, s := range []*Shop{
		{Name: "20 Rupee Cave", Kind: ShopTakeAny, Config: 0x83, Shopkeeper: 0xA0, RoomID: 0x0112, DoorID: 0x7B, Region: r, Writes: map[int][]int{0xDBBED: {0x58}}},
		{Name: "50 Rupee Cave", Kind: ShopTakeAny, Config: 0x83, Shopkeeper: 0xA0, RoomID: 0x0112, DoorID: 0x79, Region: r, Writes: map[int][]int{0xDBBEB: {0x58}}},
		{Name: "Bonk Fairy (Light)", Kind: ShopTakeAny, Config: 0x83, Shopkeeper: 0xA0, RoomID: 0x0112, DoorID: 0x77, Region: r, Writes: map[int][]int{0xDBBE9: {0x58}}},
		{Name: "Desert Fairy", Kind: ShopTakeAny, Config: 0x83, Shopkeeper: 0xA0, RoomID: 0x0112, DoorID: 0x72, Region: r, Writes: map[int][]int{0xDBBE4: {0x58}}},
		{Name: "Good Bee Cave", Kind: ShopTakeAny, Config: 0x83, Shopkeeper: 0xA0, RoomID: 0x0112, DoorID: 0x6B, Region: r, Writes: map[int][]int{0xDBBDD: {0x58}}},
		{Name: "Lake Hylia Fortune Teller", Kind: ShopTakeAny, Config: 0x83, Shopkeeper: 0xA0, RoomID: 0x011F, DoorID: 0x73, Region: r, Writes: map[int][]int{0xDBBE5: {0x46}}},
		{Name: "Light Hype Fairy", Kind: ShopTakeAny, Config: 0x83, Shopkeeper: 0xA0, RoomID: 0x0112, DoorID: 0x6C, Region: r, Writes: map[int][]int{0xDBBDE: {0x58}}},
		{Name: "Kakariko Gamble Game", Kind: ShopTakeAny, Config: 0x83, Shopkeeper: 0xA0, RoomID: 0x011F, DoorID: 0x67, Region: r, Writes: map[int][]int{0xDBBD9: {0x46}}},
	} {
		r.Shops.Add(s)
	}
	r.Shops.Get("20 Rupee Cave").Requirement = func(_ *LocationCollection, it *item.Collection) bool { return it.CanLiftRocks() }
	r.Shops.Get("50 Rupee Cave").Requirement = r.Shops.Get("20 Rupee Cave").Requirement
	r.Shops.Get("Bonk Fairy (Light)").Requirement = func(_ *LocationCollection, it *item.Collection) bool { return it.Has1("PegasusBoots") }
	r.Shops.Get("Light Hype Fairy").Requirement = func(_ *LocationCollection, it *item.Collection) bool { return it.CanBombThings() }

	aginah.Requirement = func(_ *LocationCollection, it *item.Collection) bool { return it.CanBombThings() }
	miniMold := func(_ *LocationCollection, it *item.Collection) bool {
		return it.CanBombThings() && it.CanKillMostThings(w, 5)
	}
	mmFL.Requirement = miniMold
	mmL.Requirement = miniMold
	mmR.Requirement = miniMold
	mmFR.Requirement = miniMold
	mmNpc.Requirement = miniMold

	hobo.Requirement = func(_ *LocationCollection, it *item.Collection) bool {
		return (w.ConfigBool("canWaterWalk", false) && it.Has1("PegasusBoots")) ||
			w.ConfigBool("canFakeFlipper", false) || it.Has1("Flippers")
	}
	bombos.Requirement = func(locs *LocationCollection, it *item.Collection) bool {
		south := false
		if rr := w.Region("South Dark World"); rr != nil {
			south = rr.CanEnter(locs, it)
		}
		return it.Has1("BookOfMudora") &&
			(it.HasSword(2) || (w.ConfigString("mode.weapons", "") == "swordless" && it.Has1("Hammer"))) &&
			(w.ConfigBool("canOneFrameClipOW", false) ||
				(w.ConfigBool("canBootsClip", false) && it.Has1("PegasusBoots")) ||
				(it.Has1("MagicMirror") && south))
	}
	cave45.Requirement = func(locs *LocationCollection, it *item.Collection) bool {
		south := false
		if rr := w.Region("South Dark World"); rr != nil {
			south = rr.CanEnter(locs, it)
		}
		return w.ConfigBool("canOneFrameClipOW", false) ||
			(w.ConfigBool("canBootsClip", false) && it.Has1("PegasusBoots")) ||
			(it.Has1("MagicMirror") && south)
	}
	checker.Requirement = func(locs *LocationCollection, it *item.Collection) bool {
		mire := false
		if rr := w.Region("Mire"); rr != nil {
			mire = rr.CanEnter(locs, it)
		}
		return it.CanLiftRocks() &&
			(w.ConfigBool("canOneFrameClipOW", false) ||
				(w.ConfigBool("canBootsClip", false) && it.Has1("PegasusBoots")) ||
				(it.Has1("MagicMirror") && mire))
	}
	library.Requirement = func(_ *LocationCollection, it *item.Collection) bool { return it.Has1("PegasusBoots") }
	desertLedge.Requirement = func(locs *LocationCollection, it *item.Collection) bool {
		if rr := w.Region("Desert Palace"); rr != nil {
			return rr.CanEnter(locs, it)
		}
		return false
	}
	lakeIsland.Requirement = func(locs *LocationCollection, it *item.Collection) bool {
		nwdw := false
		if rr := w.Region("North West Dark World"); rr != nil {
			nwdw = rr.CanEnter(locs, it)
		}
		nedw := false
		if rr := w.Region("North East Dark World"); rr != nil {
			nedw = rr.CanEnter(locs, it)
		}
		return w.ConfigBool("canOneFrameClipOW", false) ||
			(w.ConfigBool("canBootsClip", false) && it.Has1("PegasusBoots")) ||
			(it.Has1("MagicMirror") && w.ConfigBool("canWaterWalk", false) && it.Has1("PegasusBoots") &&
				(it.Has1("MoonPearl") || (w.ConfigBool("canOWYBA", false) && it.HasABottle())) && nwdw) ||
			(it.Has1("Flippers") && it.Has1("MagicMirror") &&
				(w.ConfigBool("canBunnySurf", false) || it.Has1("MoonPearl") ||
					(w.ConfigBool("canOWYBA", false) && it.HasABottle())) && nedw)
	}
	flute.Requirement = func(_ *LocationCollection, it *item.Collection) bool { return it.Has1("Shovel") }

	r.CanEnterFn = func(_ *LocationCollection, it *item.Collection) bool { return it.Has1("RescueZelda") }
	return r
}
