package world

import "github.com/JackDoan/alttpr-go/internal/item"

// bunny is the recurring "MoonPearl OR bottle-rev OR canBunnyRevive" check.
func bunny(w *World, it *item.Collection) bool {
	return it.Has1("MoonPearl") ||
		(w.ConfigBool("canOWYBA", false) && it.HasABottle()) ||
		(w.ConfigBool("canBunnyRevive", false) && it.CanBunnyRevive(w))
}

// ----------------------------------------------------------------------
// Standard / DarkWorld/DeathMountain/West
// PHP: app/Region/Standard/DarkWorld/DeathMountain/West.php
// ----------------------------------------------------------------------
func newStandardDWWestDM(b *regionBuilder) *Region {
	w := b.w
	r := NewRegion("West Dark World Death Mountain", w)

	spikeCave := loc(r, KindChest, "Spike Cave", []int{0xEA8B})
	fairy := &Shop{Name: "Dark Death Mountain Fairy", Kind: ShopTakeAny, Config: 0x83, Shopkeeper: 0xC1, RoomID: 0x0112, DoorID: 0x70, Region: r,
		Writes: map[int][]int{0xDBBE2: {0x58}}}
	r.Shops.Add(fairy)

	pearlOrAlt := func(it *item.Collection) bool {
		return it.Has1("MoonPearl") ||
			(w.ConfigBool("canOWYBA", false) && it.HasABottle() &&
				((it.Has1("PegasusBoots") && w.ConfigBool("canBootsClip", false)) ||
					w.ConfigBool("canOneFrameClipOW", false)))
	}
	fairy.Requirement = func(_ *LocationCollection, it *item.Collection) bool { return pearlOrAlt(it) }
	spikeCave.Requirement = func(_ *LocationCollection, it *item.Collection) bool {
		bunnyClause := it.Has1("MoonPearl") ||
			(w.ConfigBool("canOWYBA", false) && it.HasABottle() &&
				((it.Has1("PegasusBoots") && w.ConfigBool("canBootsClip", false)) ||
					w.ConfigBool("canOneFrameClipOW", false)) &&
				((it.Has1("Cape") && it.CanExtendMagic(nil, 3)) ||
					((!w.ConfigBool("region.cantTakeDamage", false) || it.CanExtendMagic(nil, 3)) && it.Has1("CaneOfByrna"))))
		survival := (it.CanExtendMagic(nil, 2) && it.Has1("Cape")) ||
			((!w.ConfigBool("region.cantTakeDamage", false) || it.CanExtendMagic(nil, 2)) && it.Has1("CaneOfByrna"))
		return bunnyClause && it.Has1("Hammer") && it.CanLiftRocks() && survival
	}

	r.CanEnterFn = func(locs *LocationCollection, it *item.Collection) bool {
		west := false
		if rr := w.Region("West Death Mountain"); rr != nil {
			west = rr.CanEnter(locs, it)
		}
		return it.Has1("RescueZelda") && west
	}
	return r
}

// ----------------------------------------------------------------------
// Standard / DarkWorld/DeathMountain/East
// PHP: app/Region/Standard/DarkWorld/DeathMountain/East.php
// ----------------------------------------------------------------------
func newStandardDWEastDM(b *regionBuilder) *Region {
	w := b.w
	r := NewRegion("East Dark World Death Mountain", w)

	sbTop := loc(r, KindChest, "Superbunny Cave - Top", []int{0xEA7C})
	sbBot := loc(r, KindChest, "Superbunny Cave - Bottom", []int{0xEA7F})
	hsTR := loc(r, KindChest, "Hookshot Cave - Top Right", []int{0xEB51})
	hsTL := loc(r, KindChest, "Hookshot Cave - Top Left", []int{0xEB54})
	hsBL := loc(r, KindChest, "Hookshot Cave - Bottom Left", []int{0xEB57})
	hsBR := loc(r, KindChest, "Hookshot Cave - Bottom Right", []int{0xEB5A})

	shop := &Shop{Name: "Dark World Death Mountain Shop", Kind: ShopRegular, Config: 0x03, Shopkeeper: 0xC1, RoomID: 0x0112, DoorID: 0x6E, Region: r}
	shop.AddInventory(0, b.item("RedPotion"), 150)
	shop.AddInventory(1, b.item("Heart"), 10)
	shop.AddInventory(2, b.item("TenBombs"), 50)
	r.Shops.Add(shop)

	sbReq := func(_ *LocationCollection, it *item.Collection) bool {
		return w.ConfigBool("canSuperBunny", false) || it.Has1("MoonPearl") ||
			((w.ConfigBool("canOWYBA", false) && it.HasABottle()) &&
				((w.ConfigBool("canBootsClip", false) && it.Has1("PegasusBoots")) ||
					w.ConfigBool("canOneFrameClipOW", false)))
	}
	sbTop.Requirement = sbReq
	sbBot.Requirement = sbReq

	hsReq := func(_ *LocationCollection, it *item.Collection) bool {
		return it.Has1("Hookshot") &&
			(it.Has1("MoonPearl") || (w.ConfigBool("canOWYBA", false) && it.HasABottle())) &&
			(it.CanLiftRocks() || w.ConfigBool("canOneFrameClipOW", false) ||
				(w.ConfigBool("canBootsClip", false) && it.Has1("PegasusBoots")))
	}
	hsTR.Requirement = hsReq
	hsTL.Requirement = hsReq
	hsBL.Requirement = hsReq
	hsBR.Requirement = func(_ *LocationCollection, it *item.Collection) bool {
		mover := it.Has1("Hookshot") || (w.ConfigString("itemPlacement", "") != "basic" && it.Has1("PegasusBoots"))
		return mover &&
			(it.Has1("MoonPearl") || (w.ConfigBool("canOWYBA", false) && it.HasABottle())) &&
			(it.CanLiftRocks() || w.ConfigBool("canOneFrameClipOW", false) ||
				(w.ConfigBool("canBootsClip", false) && it.Has1("PegasusBoots")))
	}

	r.CanEnterFn = func(locs *LocationCollection, it *item.Collection) bool {
		east := false
		if rr := w.Region("East Death Mountain"); rr != nil {
			east = rr.CanEnter(locs, it)
		}
		west := false
		if rr := w.Region("West Death Mountain"); rr != nil {
			west = rr.CanEnter(locs, it)
		}
		return it.Has1("RescueZelda") &&
			((it.CanLiftDarkRocks() && east) ||
				(w.ConfigBool("canBootsClip", false) && it.Has1("PegasusBoots") &&
					(it.Has1("MoonPearl") || it.Has1("Hammer") ||
						(w.ConfigBool("canOWYBA", false) && it.HasABottle()))) ||
				w.ConfigBool("canOneFrameClipOW", false) ||
				(west && (w.ConfigBool("canMirrorClip", false) || w.ConfigBool("canMirrorWrap", false)) && it.Has1("MagicMirror")))
	}
	return r
}

// ----------------------------------------------------------------------
// Standard / DarkWorld/Mire
// PHP: app/Region/Standard/DarkWorld/Mire.php
// ----------------------------------------------------------------------
func newStandardDWMire(b *regionBuilder) *Region {
	w := b.w
	r := NewRegion("Mire", w)

	left := loc(r, KindChest, "Mire Shed - Left", []int{0xEA73})
	right := loc(r, KindChest, "Mire Shed - Right", []int{0xEA76})

	for _, s := range []*Shop{
		{Name: "Dark Desert Fairy", Kind: ShopTakeAny, Config: 0x83, Shopkeeper: 0xC1, RoomID: 0x0112, DoorID: 0x56, Region: r, Writes: map[int][]int{0xDBBC8: {0x58}}},
		{Name: "Dark Desert Hint", Kind: ShopTakeAny, Config: 0x83, Shopkeeper: 0xC1, RoomID: 0x0112, DoorID: 0x62, Region: r, Writes: map[int][]int{0xDBBD4: {0x58}}},
	} {
		r.Shops.Add(s)
	}
	shopGate := func(_ *LocationCollection, it *item.Collection) bool {
		return it.Has1("MoonPearl") ||
			((w.ConfigBool("canOWYBA", false) && it.HasABottle()) &&
				(w.ConfigBool("canOneFrameClipOW", false) || it.HasBottle(2) ||
					(it.Has1("MagicMirror") && it.Has1("BugCatchingNet") && w.ConfigBool("canBunnyRevive", false)) ||
					(w.ConfigBool("canBootsClip", false) && it.Has1("PegasusBoots"))))
	}
	r.Shops.Get("Dark Desert Fairy").Requirement = shopGate
	r.Shops.Get("Dark Desert Hint").Requirement = shopGate

	shedReq := func(_ *LocationCollection, it *item.Collection) bool {
		return shopGate(nil, it) ||
			(w.ConfigBool("canSuperBunny", false) && it.Has1("MagicMirror"))
	}
	left.Requirement = shedReq
	right.Requirement = shedReq

	r.CanEnterFn = func(locs *LocationCollection, it *item.Collection) bool {
		south := false
		if rr := w.Region("South Dark World"); rr != nil {
			south = rr.CanEnter(locs, it)
		}
		return it.Has1("RescueZelda") &&
			((it.CanLiftDarkRocks() &&
				(it.CanFly(w) || (w.ConfigBool("canBootsClip", false) && it.Has1("PegasusBoots")))) ||
				w.ConfigBool("canOneFrameClipOW", false) ||
				((((it.Has1("MoonPearl") || (w.ConfigBool("canBunnyRevive", false) && it.CanBunnyRevive(w))) &&
					(w.ConfigBool("canBootsClip", false) && it.Has1("PegasusBoots"))) ||
					(w.ConfigBool("canMirrorWrap", false) && it.Has1("MagicMirror"))) && south) ||
				(w.ConfigBool("canOWYBA", false) && it.HasABottle()))
	}
	return r
}

// ----------------------------------------------------------------------
// Standard / DarkWorld/South
// PHP: app/Region/Standard/DarkWorld/South.php
// ----------------------------------------------------------------------
func newStandardDWSouth(b *regionBuilder) *Region {
	w := b.w
	r := NewRegion("South Dark World", w)

	hypeTop := loc(r, KindChest, "Hype Cave - Top", []int{0xEB1E})
	hypeMR := loc(r, KindChest, "Hype Cave - Middle Right", []int{0xEB21})
	hypeML := loc(r, KindChest, "Hype Cave - Middle Left", []int{0xEB24})
	hypeBot := loc(r, KindChest, "Hype Cave - Bottom", []int{0xEB27})
	stumpy := loc(r, KindNpc, "Stumpy", []int{0x330C7})
	hypeNpc := loc(r, KindNpc, "Hype Cave - NPC", []int{0x180011})
	dig := loc(r, KindDig, "Digging Game", []int{0x180148})

	shop := &Shop{Name: "Dark World Lake Hylia Shop", Kind: ShopRegular, Config: 0x03, Shopkeeper: 0xC1, RoomID: 0x010F, DoorID: 0x74, Region: r}
	shop.AddInventory(0, b.item("RedPotion"), 150)
	shop.AddInventory(1, b.item("BlueShield"), 50)
	shop.AddInventory(2, b.item("TenBombs"), 50)
	r.Shops.Add(shop)
	for _, s := range []*Shop{
		{Name: "Archery Game", Kind: ShopTakeAny, Config: 0x83, Shopkeeper: 0xC1, RoomID: 0x010F, DoorID: 0x59, Region: r, Writes: map[int][]int{0xDBBCB: {0x60}}},
		{Name: "Bonk Fairy (Dark)", Kind: ShopTakeAny, Config: 0x83, Shopkeeper: 0xC1, RoomID: 0x0112, DoorID: 0x78, Region: r, Writes: map[int][]int{0xDBBEA: {0x58}}},
		{Name: "Dark Lake Hylia Ledge Fairy", Kind: ShopTakeAny, Config: 0x83, Shopkeeper: 0xC1, RoomID: 0x0112, DoorID: 0x81, Region: r, Writes: map[int][]int{0xDBBF3: {0x58}}},
		{Name: "Dark Lake Hylia Ledge Hint", Kind: ShopTakeAny, Config: 0x83, Shopkeeper: 0xC1, RoomID: 0x0112, DoorID: 0x6A, Region: r, Writes: map[int][]int{0xDBBDC: {0x58}}},
		{Name: "Dark Lake Hylia Ledge Spike Cave", Kind: ShopTakeAny, Config: 0x83, Shopkeeper: 0xC1, RoomID: 0x0112, DoorID: 0x7C, Region: r, Writes: map[int][]int{0xDBBEE: {0x58}}},
	} {
		r.Shops.Add(s)
	}

	bn := func(_ *LocationCollection, it *item.Collection) bool { return bunny(w, it) }
	hypeTop.Requirement = bn
	hypeMR.Requirement = bn
	hypeML.Requirement = bn
	hypeBot.Requirement = bn
	hypeNpc.Requirement = bn
	stumpy.Requirement = func(_ *LocationCollection, it *item.Collection) bool {
		return bunny(w, it) || (it.Has1("MagicMirror") && w.ConfigBool("canMirrorWrap", false))
	}
	dig.Requirement = bn

	r.Shops.Get("Bonk Fairy (Dark)").Requirement = func(_ *LocationCollection, it *item.Collection) bool {
		return it.Has1("PegasusBoots") && bunny(w, it)
	}
	hyliaFlipper := func(locs *LocationCollection, it *item.Collection) bool {
		nwdw := false
		if rr := w.Region("North West Dark World"); rr != nil {
			nwdw = rr.CanEnter(locs, it)
		}
		return it.Has1("Flippers") ||
			(nwdw && w.ConfigBool("canFakeFlipper", false) && !w.ConfigBool("region.cantTakeDamage", false)) ||
			(w.ConfigBool("canBunnyRevive", false) && it.CanBunnyRevive(w))
	}
	r.Shops.Get("Dark Lake Hylia Ledge Fairy").Requirement = func(locs *LocationCollection, it *item.Collection) bool {
		return (it.CanBombThings() || (it.Has1("PegasusBoots") && w.ConfigBool("canBootsClip", false))) &&
			hyliaFlipper(locs, it) && bunny(w, it)
	}
	r.Shops.Get("Dark Lake Hylia Ledge Hint").Requirement = func(locs *LocationCollection, it *item.Collection) bool {
		return hyliaFlipper(locs, it) && bunny(w, it)
	}
	r.Shops.Get("Dark Lake Hylia Ledge Spike Cave").Requirement = func(locs *LocationCollection, it *item.Collection) bool {
		return it.CanLiftRocks() && hyliaFlipper(locs, it) && bunny(w, it)
	}

	r.CanEnterFn = func(locs *LocationCollection, it *item.Collection) bool {
		nedw := false
		if rr := w.Region("North East Dark World"); rr != nil {
			nedw = rr.CanEnter(locs, it)
		}
		nwdw := false
		if rr := w.Region("North West Dark World"); rr != nil {
			nwdw = rr.CanEnter(locs, it)
		}
		return it.Has1("RescueZelda") &&
			((w.ConfigBool("canOWYBA", false) && it.HasABottle()) ||
				((it.Has1("MoonPearl") || (w.ConfigBool("canBunnyRevive", false) && it.CanBunnyRevive(w))) &&
					(nedw &&
						(it.Has1("Hammer") ||
							(w.ConfigBool("canSuperSpeed", false) && it.CanSpinSpeed() &&
								(w.ConfigBool("canFakeFlipper", false) || it.Has1("Flippers")))))) ||
				nwdw)
	}
	return r
}

// ----------------------------------------------------------------------
// Standard / DarkWorld/NorthWest
// PHP: app/Region/Standard/DarkWorld/NorthWest.php
// ----------------------------------------------------------------------
func newStandardDWNorthWest(b *regionBuilder) *Region {
	w := b.w
	r := NewRegion("North West Dark World", w)

	brewery := loc(r, KindChest, "Brewery", []int{0xE9EC})
	cShape := loc(r, KindChest, "C-Shaped House", []int{0xE9EF})
	chestGame := loc(r, KindChest, "Chest Game", []int{0xEDA8})
	hammerPegs := loc(r, KindStanding, "Hammer Pegs", []int{0x180006})
	bumper := loc(r, KindStanding, "Bumper Cave", []int{0x180146})
	smithAddr := 0x3355C
	if w.ConfigBool("region.swordsInPool", true) {
		smithAddr = 0x18002A
	}
	smith := loc(r, KindNpc, "Blacksmith", []int{smithAddr})
	purple := loc(r, KindNpc, "Purple Chest", []int{0x33D68})

	for _, s := range []*Shop{
		{Name: "Dark World Forest Shop", Kind: ShopRegular, Config: 0x03, Shopkeeper: 0xC1, RoomID: 0x0110, DoorID: 0x75, Region: r},
		{Name: "Dark World Lumberjack Hut Shop", Kind: ShopRegular, Config: 0x03, Shopkeeper: 0xC1, RoomID: 0x010F, DoorID: 0x57, Region: r},
		{Name: "Dark World Outcasts Shop", Kind: ShopRegular, Config: 0x03, Shopkeeper: 0xC1, RoomID: 0x010F, DoorID: 0x60, Region: r},
		{Name: "Dark Sanctuary Hint", Kind: ShopTakeAny, Config: 0x83, Shopkeeper: 0xC1, RoomID: 0x0112, DoorID: 0x5A, Region: r, Writes: map[int][]int{0xDBBCC: {0x58}}},
		{Name: "Fortune Teller (Dark)", Kind: ShopTakeAny, Config: 0x83, Shopkeeper: 0xC1, RoomID: 0x010F, DoorID: 0x66, Region: r, Writes: map[int][]int{0xDBBD8: {0x60}}},
	} {
		r.Shops.Add(s)
	}
	r.Shops.Get("Dark World Forest Shop").AddInventory(0, b.item("RedShield"), 500)
	r.Shops.Get("Dark World Forest Shop").AddInventory(1, b.item("Bee"), 10)
	r.Shops.Get("Dark World Forest Shop").AddInventory(2, b.item("TenArrows"), 30)
	r.Shops.Get("Dark World Lumberjack Hut Shop").AddInventory(0, b.item("RedPotion"), 150)
	r.Shops.Get("Dark World Lumberjack Hut Shop").AddInventory(1, b.item("BlueShield"), 50)
	r.Shops.Get("Dark World Lumberjack Hut Shop").AddInventory(2, b.item("TenBombs"), 50)
	r.Shops.Get("Dark World Outcasts Shop").AddInventory(0, b.item("RedPotion"), 150)
	r.Shops.Get("Dark World Outcasts Shop").AddInventory(1, b.item("BlueShield"), 50)
	r.Shops.Get("Dark World Outcasts Shop").AddInventory(2, b.item("TenBombs"), 50)
	r.Shops.Get("Dark World Outcasts Shop").Requirement = func(_ *LocationCollection, it *item.Collection) bool {
		return it.Has1("Hammer") && bunny(w, it)
	}

	brewery.Requirement = func(_ *LocationCollection, it *item.Collection) bool { return it.CanBombThings() && bunny(w, it) }
	cShape.Requirement = func(_ *LocationCollection, it *item.Collection) bool {
		return it.Has1("MoonPearl") || w.ConfigBool("canSuperBunny", false) ||
			(w.ConfigBool("canOWYBA", false) && it.HasABottle()) ||
			(w.ConfigBool("canBunnyRevive", false) && it.CanBunnyRevive(w))
	}
	chestGame.Requirement = cShape.Requirement

	hammerPegs.Requirement = func(_ *LocationCollection, it *item.Collection) bool {
		return it.Has1("Hammer") && bunny(w, it) &&
			(it.CanLiftDarkRocks() ||
				(it.Has1("MagicMirror") && w.ConfigBool("canMirrorWrap", false)) ||
				((w.ConfigBool("canFakeFlipper", false) || it.Has1("Flippers")) &&
					(w.ConfigBool("canOneFrameClipOW", false) ||
						(w.ConfigBool("canBootsClip", false) && it.Has1("PegasusBoots")) ||
						(w.ConfigBool("canSuperSpeed", false) && it.CanSpinSpeed()))))
	}
	bumper.Requirement = func(_ *LocationCollection, it *item.Collection) bool {
		return w.ConfigBool("canOneFrameClipOW", false) ||
			(bunny(w, it) &&
				((w.ConfigBool("canBootsClip", false) && it.Has1("PegasusBoots")) ||
					((w.ConfigString("itemPlacement", "") != "basic" || it.Has1("Hookshot")) &&
						it.CanLiftRocks() && it.Has1("Cape"))))
	}
	smith.Requirement = func(_ *LocationCollection, it *item.Collection) bool {
		return (w.ConfigString("itemPlacement", "") != "basic" || it.Has1("MagicMirror")) &&
			((bunny(w, it) && it.CanLiftDarkRocks()) ||
				((w.ConfigBool("canOWYBA", false) && it.HasABottle()) &&
					(w.ConfigBool("canOneFrameClipOW", false) ||
						(w.ConfigBool("canBootsClip", false) && it.Has1("PegasusBoots") &&
							(it.Has1("MoonPearl") || it.HasBottle(2))))))
	}
	purple.Requirement = func(locs *LocationCollection, it *item.Collection) bool {
		return smith.CanAccess(it, locs) &&
			((it.Has1("MagicMirror") && w.ConfigBool("canMirrorWrap", false)) ||
				(bunny(w, it) &&
					(it.CanLiftDarkRocks() ||
						((w.ConfigBool("canFakeFlipper", false) || it.Has1("Flippers")) &&
							(w.ConfigBool("canOneFrameClipOW", false) ||
								(w.ConfigBool("canBootsClip", false) && it.Has1("PegasusBoots")) ||
								(w.ConfigBool("canSuperSpeed", false) && it.CanSpinSpeed()))))))
	}

	r.CanEnterFn = func(locs *LocationCollection, it *item.Collection) bool {
		west := false
		if rr := w.Region("West Death Mountain"); rr != nil {
			west = rr.CanEnter(locs, it)
		}
		nedw := false
		if rr := w.Region("North East Dark World"); rr != nil {
			nedw = rr.CanEnter(locs, it)
		}
		return it.Has1("RescueZelda") &&
			(w.ConfigBool("canOneFrameClipOW", false) ||
				(w.ConfigBool("canOWYBA", false) && it.HasABottle()) ||
				(west && it.Has1("MagicMirror") &&
					(w.ConfigBool("canMirrorClip", false) ||
						(w.ConfigBool("canMirrorWrap", false) &&
							((w.ConfigBool("canBootsClip", false) && it.Has1("PegasusBoots")) ||
								(w.ConfigBool("canSuperSpeed", false) && it.CanSpinSpeed()))))) ||
				(w.ConfigBool("canBunnySurf", false) && w.ConfigBool("canMirrorWrap", false) && it.Has1("MagicMirror") &&
					it.Has1("Flippers") && it.Has1("DefeatAgahnim")) ||
				(it.Has1("MoonPearl") &&
					((nedw &&
						((it.Has1("Hookshot") || (w.ConfigBool("canMirrorWrap", false) && it.Has1("MagicMirror"))) &&
							(it.CanLiftRocks() || it.Has1("Hammer") ||
								(it.Has1("Flippers") || (w.ConfigBool("canBunnyRevive", false) && it.CanBunnyRevive(w)))))) ||
						(it.Has1("Hammer") && it.CanLiftRocks()) ||
						it.CanLiftDarkRocks() ||
						(w.ConfigBool("canBootsClip", false) && it.Has1("PegasusBoots")) ||
						(w.ConfigBool("canSuperSpeed", false) && it.CanSpinSpeed()))))
	}
	return r
}

// ----------------------------------------------------------------------
// Standard / DarkWorld/NorthEast (includes Pyramid + win condition)
// PHP: app/Region/Standard/DarkWorld/NorthEast.php
// ----------------------------------------------------------------------
func newStandardDWNorthEast(b *regionBuilder) *Region {
	w := b.w
	r := NewRegion("North East Dark World", w)

	catfish := loc(r, KindStanding, "Catfish", []int{0xEE185})
	loc(r, KindStanding, "Pyramid", []int{0x180147})
	pfSword := loc(r, KindTrade, "Pyramid Fairy - Sword", []int{0x180028})
	_ = pfSword.SetItem(b.item("L1Sword"))
	pfBow := loc(r, KindTrade, "Pyramid Fairy - Bow", []int{0x34914})
	_ = pfBow.SetItem(b.item("Bow"))

	var pfLeft, pfRight *Location
	if w.ConfigBool("region.swordsInPool", true) {
		pfLeft = loc(r, KindChest, "Pyramid Fairy - Left", []int{0xE980})
		pfRight = loc(r, KindChest, "Pyramid Fairy - Right", []int{0xE983})
	}

	ganon := loc(r, KindPrizeEvent, "Ganon", nil)
	_ = ganon.SetItem(b.item("DefeatGanon"))
	r.SetPrizeLocation(ganon)

	pshop := &Shop{Name: "Dark World Potion Shop", Kind: ShopRegular, Config: 0x03, Shopkeeper: 0xC1, RoomID: 0x010F, DoorID: 0x6F, Region: r}
	pshop.AddInventory(0, b.item("RedPotion"), 150)
	pshop.AddInventory(1, b.item("BlueShield"), 50)
	pshop.AddInventory(2, b.item("TenBombs"), 50)
	r.Shops.Add(pshop)
	for _, s := range []*Shop{
		{Name: "Dark Lake Hylia Fairy", Kind: ShopTakeAny, Config: 0x83, Shopkeeper: 0xC1, RoomID: 0x0112, DoorID: 0x6D, Region: r, Writes: map[int][]int{0xDBBDF: {0x58}}},
		{Name: "East Dark World Hint", Kind: ShopTakeAny, Config: 0x83, Shopkeeper: 0xC1, RoomID: 0x0112, DoorID: 0x69, Region: r, Writes: map[int][]int{0xDBBDB: {0x58}}},
		{Name: "Palace of Darkness Hint", Kind: ShopTakeAny, Config: 0x83, Shopkeeper: 0xC1, RoomID: 0x010F, DoorID: 0x68, Region: r, Writes: map[int][]int{0xDBBDA: {0x60}}},
	} {
		r.Shops.Add(s)
	}
	pshop.Requirement = func(_ *LocationCollection, it *item.Collection) bool {
		return (w.ConfigBool("canBunnyRevive", false) && it.CanBunnyRevive(w)) ||
			((it.Has1("MoonPearl") || (w.ConfigBool("canOWYBA", false) && it.HasABottle())) &&
				(w.ConfigBool("canOneFrameClipOW", false) ||
					(w.ConfigBool("canBootsClip", false) && it.Has1("PegasusBoots")) ||
					it.Has1("Hammer") || it.Has1("Flippers") || it.CanLiftRocks()))
	}
	bn := func(_ *LocationCollection, it *item.Collection) bool { return bunny(w, it) }
	r.Shops.Get("East Dark World Hint").Requirement = bn
	r.Shops.Get("Palace of Darkness Hint").Requirement = bn

	catfish.Requirement = func(_ *LocationCollection, it *item.Collection) bool {
		return bunny(w, it) &&
			(w.ConfigBool("canOneFrameClipOW", false) ||
				(w.ConfigBool("canBootsClip", false) && it.Has1("PegasusBoots")) ||
				it.CanLiftRocks())
	}

	pfBaseAccess := func(locs *LocationCollection, it *item.Collection) bool {
		south := false
		if rr := w.Region("South Dark World"); rr != nil {
			south = rr.CanEnter(locs, it)
		}
		main := it.Has1("Crystal5") && it.Has1("Crystal6") && south &&
			((it.Has1("MoonPearl") && it.Has1("Hammer")) ||
				(w.ConfigBool("canOWYBA", false) && it.HasABottle()) ||
				(it.Has1("MagicMirror") && it.Has1("DefeatAgahnim")))
		clip := w.ConfigBool("canMirrorClip", false) && it.Has1("MagicMirror") &&
			((w.ConfigBool("canSuperSpeed", false) && it.CanSpinSpeed()) ||
				w.ConfigBool("canOneFrameClipOW", false) ||
				(w.ConfigBool("canBootsClip", false) && it.Has1("PegasusBoots")))
		return main || clip
	}
	pfSword.Requirement = func(locs *LocationCollection, it *item.Collection) bool {
		return it.HasSword(1) && pfBaseAccess(locs, it)
	}
	pfBow.Requirement = func(locs *LocationCollection, it *item.Collection) bool {
		return it.CanShootArrows(w, 1) && pfBaseAccess(locs, it)
	}
	if pfLeft != nil {
		pfLeft.Requirement = pfBaseAccess
		pfRight.Requirement = pfBaseAccess
	}

	r.CanEnterFn = func(locs *LocationCollection, it *item.Collection) bool {
		west := false
		if rr := w.Region("West Death Mountain"); rr != nil {
			west = rr.CanEnter(locs, it)
		}
		return it.Has1("RescueZelda") &&
			((w.ConfigBool("canOWYBA", false) && it.HasABottle()) ||
				w.ConfigBool("canOneFrameClipOW", false) ||
				(it.Has1("MoonPearl") && (w.ConfigBool("canBootsClip", false) && it.Has1("PegasusBoots"))) ||
				(it.Has1("MagicMirror") && w.ConfigBool("canMirrorClip", false) && west &&
					((w.ConfigBool("canBootsClip", false) && it.Has1("PegasusBoots")) ||
						(w.ConfigBool("canBunnyRevive", false) && it.CanBunnyRevive(w)) ||
						(w.ConfigBool("canSuperSpeed", false) && it.CanSpinSpeed()) ||
						(it.Has1("MoonPearl") && w.ConfigBool("canFakeFlipper", false)))) ||
				it.Has1("DefeatAgahnim") ||
				(it.Has1("MagicMirror") && w.ConfigBool("canMirrorWrap", false) &&
					w.ConfigBool("canBunnyRevive", false) && it.CanBunnyRevive(w) &&
					((w.ConfigBool("canSuperSpeed", false) && it.CanSpinSpeed()) ||
						(w.ConfigBool("canBootsClip", false) && it.Has1("PegasusBoots")) ||
						w.ConfigBool("canMirrorClip", false)) && west) ||
				(it.Has1("Hammer") && it.CanLiftRocks() && it.Has1("MoonPearl")) ||
				((it.CanLiftDarkRocks() ||
					(west && ((w.ConfigBool("canSuperSpeed", false) && it.CanSpinSpeed()) ||
						(it.Has1("MagicMirror") && w.ConfigBool("canMirrorClip", false))))) &&
					it.Has1("MoonPearl") &&
					(it.Has1("Hammer") || it.Has1("Flippers") ||
						(it.Has1("MagicMirror") && w.ConfigBool("canMirrorWrap", false) && it.CanLiftRocks()) ||
						(w.ConfigBool("canBunnyRevive", false) && it.CanBunnyRevive(w)) ||
						(w.ConfigBool("canWaterWalk", false) && it.Has1("PegasusBoots")) ||
						(w.ConfigBool("canFakeFlipper", false) && !w.ConfigBool("region.cantTakeDamage", false)))))
	}

	// Ganon prize requirement: full goal-aware win condition.
	ganon.Requirement = func(_ *LocationCollection, it *item.Collection) bool {
		goal := w.ConfigString("goal", "")
		if goal == "ganonhunt" && !it.Has("TriforcePiece", w.ConfigInt("item.Goal.Required", 0)) {
			return false
		}
		if goal == "dungeons" {
			if !it.Has1("PendantOfCourage") || !it.Has1("PendantOfWisdom") || !it.Has1("PendantOfPower") ||
				!it.Has1("DefeatAgahnim") || !it.Has1("Crystal1") || !it.Has1("Crystal2") ||
				!it.Has1("Crystal3") || !it.Has1("Crystal4") || !it.Has1("Crystal5") ||
				!it.Has1("Crystal6") || !it.Has1("Crystal7") || !it.Has1("DefeatAgahnim2") {
				return false
			}
		}
		if goal == "ganon" || goal == "fast_ganon" {
			n := 0
			for i := 1; i <= 7; i++ {
				if it.Has1("Crystal" + itoaW(i)) {
					n++
				}
			}
			if n < w.ConfigInt("crystals.ganon", 7) {
				return false
			}
		}
		if !(it.Has1("MoonPearl") || (w.ConfigBool("canOWYBA", false) && it.HasABottle())) {
			return false
		}
		fastOrHunt := goal == "fast_ganon" || goal == "ganonhunt"
		if !(it.Has1("DefeatAgahnim2") || fastOrHunt) {
			return false
		}
		if w.ConfigBool("region.requireBetterBow", false) && !it.CanShootArrows(w, 2) {
			return false
		}
		lampReq := it.Has("Lamp", w.ConfigInt("item.require.Lamp", 1))
		swordless := w.ConfigString("mode.weapons", "") == "swordless" && it.Has1("Hammer") &&
			(lampReq || (it.Has1("FireRod") && it.CanExtendMagic(w, 1)))
		sword2 := !w.ConfigBool("region.requireBetterSword", false) && it.HasSword(2) &&
			(lampReq || (it.Has1("FireRod") && it.CanExtendMagic(w, 3)))
		sword3 := it.HasSword(3) && (lampReq || (it.Has1("FireRod") && it.CanExtendMagic(w, 2)))
		return swordless || sword2 || sword3
	}
	return r
}
