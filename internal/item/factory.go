package item

import "fmt"

// Registry holds the per-world cache of items. Mirrors PHP Item::$items and
// Item::$worlds (which are static class properties).
type Registry struct {
	byWorld map[int]map[string]*Item // worldID -> name -> item
	all     map[int][]*Item          // worldID -> ordered list (mirrors ItemCollection iteration)
}

// NewRegistry returns an empty Registry.
func NewRegistry() *Registry {
	return &Registry{
		byWorld: make(map[int]map[string]*Item),
		all:     make(map[int][]*Item),
	}
}

// Get fetches an item by raw name in the given world, populating the cache
// on first access. Mirrors PHP Item::get + Item::all.
func (r *Registry) Get(name string, worldID int) (*Item, error) {
	r.populate(worldID)
	if it, ok := r.byWorld[worldID][name]; ok {
		return it, nil
	}
	// PHP falls through to getNice() which matches by nice_name; we mirror
	// that by checking the i18n key form.
	niceKey := "item." + name
	for _, it := range r.all[worldID] {
		if it.NiceName == niceKey {
			return it, nil
		}
	}
	return nil, fmt.Errorf("unknown item: %s", name)
}

// All returns the ordered list of items for worldID.
func (r *Registry) All(worldID int) []*Item {
	r.populate(worldID)
	return r.all[worldID]
}

// ClearCache wipes all per-world data. Mirrors PHP Item::clearCache().
func (r *Registry) ClearCache() {
	r.byWorld = make(map[int]map[string]*Item)
	r.all = make(map[int][]*Item)
}

func (r *Registry) populate(worldID int) {
	if _, ok := r.byWorld[worldID]; ok {
		return
	}
	r.byWorld[worldID] = make(map[string]*Item)
	r.all[worldID] = nil
	for _, def := range itemDefs(worldID) {
		r.add(worldID, def)
	}
	// Logical aliases (matches Item.php:269-271).
	r.addAlias(worldID, "UncleSword", "ProgressiveSword")
	r.addAlias(worldID, "ShopKey", "KeyGK")
	r.addAlias(worldID, "ShopArrow", "Arrow")
}

func (r *Registry) add(worldID int, it *Item) {
	r.byWorld[worldID][it.Name] = it
	r.all[worldID] = append(r.all[worldID], it)
}

func (r *Registry) addAlias(worldID int, name, target string) {
	t := r.byWorld[worldID][target]
	a := &Item{
		Name:       name,
		NiceName:   "item." + name,
		Type:       TypeAlias,
		WorldID:    worldID,
		TargetName: target,
		Target:     t,
	}
	r.add(worldID, a)
}

// itemDefs returns the canonical Item table for one world, in PHP source
// order. Bytes preserve hex literal values; -1 is the PHP `null` placeholder
// used in Crystal/Event/Programmable byte arrays.
func itemDefs(worldID int) []*Item {
	g := func(name string, bytes ...int) *Item { return NewItem(name, bytes, worldID, TypeGeneric) }
	mk := func(name string, t Type, bytes ...int) *Item { return NewItem(name, bytes, worldID, t) }
	hp := func(name string, b int, power float64) *Item {
		it := NewItem(name, []int{b}, worldID, TypeUpgradeHealth)
		it.Power = power
		return it
	}
	medallion := func(name string, b0, b1, t0, t1, t2, m0, m1, m2 int) *Item {
		it := NewItem(name, []int{b0, b1}, worldID, TypeMedallion)
		it.NamedBytes = map[string]int{"t0": t0, "t1": t1, "t2": t2, "m0": m0, "m1": m1, "m2": m2}
		return it
	}

	return []*Item{
		g("Nothing", 0x5A),
		mk("L1Sword", TypeSword, 0x49),
		mk("L1SwordAndShield", TypeSword, 0x00),
		mk("L2Sword", TypeSword, 0x01),
		mk("MasterSword", TypeSword, 0x50),
		mk("L3Sword", TypeSword, 0x02),
		mk("L4Sword", TypeSword, 0x03),
		mk("BlueShield", TypeShield, 0x04),
		mk("RedShield", TypeShield, 0x05),
		mk("MirrorShield", TypeShield, 0x06),
		g("FireRod", 0x07),
		g("IceRod", 0x08),
		g("Hammer", 0x09),
		g("Hookshot", 0x0A),
		mk("Bow", TypeBow, 0x0B),
		g("Boomerang", 0x0C),
		g("Powder", 0x0D),
		mk("Bee", TypeBottleContents, 0x0E),
		medallion("Bombos", 0x0F, 0x00, 0x31, 0x90, 0x00, 0x31, 0x80, 0x00),
		medallion("Ether", 0x10, 0x01, 0x31, 0x98, 0x00, 0x13, 0x9F, 0xF1),
		medallion("Quake", 0x11, 0x02, 0x14, 0xEF, 0xC4, 0x31, 0x88, 0x00),
		g("Lamp", 0x12),
		g("Shovel", 0x13),
		g("OcarinaInactive", 0x14),
		g("CaneOfSomaria", 0x15),
		mk("Bottle", TypeBottle, 0x16),
		hp("PieceOfHeart", 0x17, 0.25),
		g("CaneOfByrna", 0x18),
		g("Cape", 0x19),
		g("MagicMirror", 0x1A),
		g("PowerGlove", 0x1B),
		g("TitansMitt", 0x1C),
		g("BookOfMudora", 0x1D),
		g("Flippers", 0x1E),
		g("MoonPearl", 0x1F),
		mk("Crystal", TypeCrystal, 0x20),
		g("BugCatchingNet", 0x21),
		mk("BlueMail", TypeArmor, 0x22),
		mk("RedMail", TypeArmor, 0x23),
		mk("Key", TypeKey, 0x24),
		mk("Compass", TypeCompass, 0x25),
		hp("HeartContainerNoAnimation", 0x26, 1),
		g("Bomb", 0x27),
		g("ThreeBombs", 0x28),
		g("Mushroom", 0x29),
		g("RedBoomerang", 0x2A),
		mk("BottleWithRedPotion", TypeBottle, 0x2B),
		mk("BottleWithGreenPotion", TypeBottle, 0x2C),
		mk("BottleWithBluePotion", TypeBottle, 0x2D),
		mk("RedPotion", TypeBottleContents, 0x2E),
		mk("GreenPotion", TypeBottleContents, 0x2F),
		mk("BluePotion", TypeBottleContents, 0x30),
		g("TenBombs", 0x31),
		mk("BigKey", TypeBigKey, 0x32),
		mk("Map", TypeMap, 0x33),
		g("OneRupee", 0x34),
		g("FiveRupees", 0x35),
		g("TwentyRupees", 0x36),
		mk("PendantOfCourage", TypePendant, 0x37, 0x04, 0x38, 0x62, 0x00, 0x69, 0x37),
		mk("PendantOfWisdom", TypePendant, 0x38, 0x01, 0x32, 0x60, 0x00, 0x69, 0x38),
		mk("PendantOfPower", TypePendant, 0x39, 0x02, 0x34, 0x60, 0x00, 0x69, 0x39),
		mk("BowAndArrows", TypeBow, 0x3A),
		mk("BowAndSilverArrows", TypeBow, 0x3B),
		mk("BottleWithBee", TypeBottle, 0x3C),
		mk("BottleWithFairy", TypeBottle, 0x3D),
		hp("BossHeartContainer", 0x3E, 1),
		hp("HeartContainer", 0x3F, 1),
		g("OneHundredRupees", 0x40),
		g("FiftyRupees", 0x41),
		g("Heart", 0x42),
		mk("Arrow", TypeArrow, 0x43),
		mk("TenArrows", TypeArrow, 0x44),
		g("SmallMagic", 0x45),
		g("ThreeHundredRupees", 0x46),
		g("TwentyRupees2", 0x47),
		mk("BottleWithGoldBee", TypeBottle, 0x48),
		g("OcarinaActive", 0x4A),
		g("PegasusBoots", 0x4B),
		mk("BombUpgrade5", TypeUpgradeBomb, 0x51),
		mk("BombUpgrade10", TypeUpgradeBomb, 0x52),
		mk("BombUpgrade50", TypeUpgradeBomb, 0x4C),
		mk("ArrowUpgrade5", TypeUpgradeArrow, 0x53),
		mk("ArrowUpgrade10", TypeUpgradeArrow, 0x54),
		mk("ArrowUpgrade70", TypeUpgradeArrow, 0x4D),
		mk("HalfMagic", TypeUpgradeMagic, 0x4E),
		mk("QuarterMagic", TypeUpgradeMagic, 0x4F),
		mk("Programmable1", TypeProgrammable, 0x55),
		mk("Programmable2", TypeProgrammable, 0x56),
		mk("Programmable3", TypeProgrammable, 0x57),
		g("SilverArrowUpgrade", 0x58),
		g("Rupoor", 0x59),
		g("RedClock", 0x5B),
		g("BlueClock", 0x5C),
		g("GreenClock", 0x5D),
		mk("ProgressiveSword", TypeSword, 0x5E),
		mk("ProgressiveShield", TypeShield, 0x5F),
		mk("ProgressiveArmor", TypeArmor, 0x60),
		g("ProgressiveGlove", 0x61),
		g("singleRNG", 0x62),
		g("multiRNG", 0x63),
		mk("ProgressiveBow", TypeBow, 0x64),
		mk("ProgressiveBowAlternate", TypeBow, 0x65),
		mk("Triforce", TypeEvent, 0x6A),
		g("PowerStar", 0x6B),
		g("TriforcePiece", 0x6C),
		mk("MapLW", TypeMap, 0x70),
		mk("MapDW", TypeMap, 0x71),
		mk("MapA2", TypeMap, 0x72),
		mk("MapD7", TypeMap, 0x73),
		mk("MapD4", TypeMap, 0x74),
		mk("MapP3", TypeMap, 0x75),
		mk("MapD5", TypeMap, 0x76),
		mk("MapD3", TypeMap, 0x77),
		mk("MapD6", TypeMap, 0x78),
		mk("MapD1", TypeMap, 0x79),
		mk("MapD2", TypeMap, 0x7A),
		mk("MapA1", TypeMap, 0x7B),
		mk("MapP2", TypeMap, 0x7C),
		mk("MapP1", TypeMap, 0x7D),
		mk("MapH1", TypeMap, 0x7E),
		mk("MapH2", TypeMap, 0x7F),
		mk("CompassA2", TypeCompass, 0x82),
		mk("CompassD7", TypeCompass, 0x83),
		mk("CompassD4", TypeCompass, 0x84),
		mk("CompassP3", TypeCompass, 0x85),
		mk("CompassD5", TypeCompass, 0x86),
		mk("CompassD3", TypeCompass, 0x87),
		mk("CompassD6", TypeCompass, 0x88),
		mk("CompassD1", TypeCompass, 0x89),
		mk("CompassD2", TypeCompass, 0x8A),
		mk("CompassA1", TypeCompass, 0x8B),
		mk("CompassP2", TypeCompass, 0x8C),
		mk("CompassP1", TypeCompass, 0x8D),
		mk("CompassH1", TypeCompass, 0x8E),
		mk("CompassH2", TypeCompass, 0x8F),
		mk("BigKeyA2", TypeBigKey, 0x92),
		mk("BigKeyD7", TypeBigKey, 0x93),
		mk("BigKeyD4", TypeBigKey, 0x94),
		mk("BigKeyP3", TypeBigKey, 0x95),
		mk("BigKeyD5", TypeBigKey, 0x96),
		mk("BigKeyD3", TypeBigKey, 0x97),
		mk("BigKeyD6", TypeBigKey, 0x98),
		mk("BigKeyD1", TypeBigKey, 0x99),
		mk("BigKeyD2", TypeBigKey, 0x9A),
		mk("BigKeyA1", TypeBigKey, 0x9B),
		mk("BigKeyP2", TypeBigKey, 0x9C),
		mk("BigKeyP1", TypeBigKey, 0x9D),
		mk("BigKeyH1", TypeBigKey, 0x9E),
		mk("BigKeyH2", TypeBigKey, 0x9F),
		mk("KeyH2", TypeKey, 0xA0),
		mk("KeyH1", TypeKey, 0xA1),
		mk("KeyP1", TypeKey, 0xA2),
		mk("KeyP2", TypeKey, 0xA3),
		mk("KeyA1", TypeKey, 0xA4),
		mk("KeyD2", TypeKey, 0xA5),
		mk("KeyD1", TypeKey, 0xA6),
		mk("KeyD6", TypeKey, 0xA7),
		mk("KeyD3", TypeKey, 0xA8),
		mk("KeyD5", TypeKey, 0xA9),
		mk("KeyP3", TypeKey, 0xAA),
		mk("KeyD4", TypeKey, 0xAB),
		mk("KeyD7", TypeKey, 0xAC),
		mk("KeyA2", TypeKey, 0xAD),
		mk("KeyGK", TypeKey, 0xAF),
		// Crystal byte arrays start with PHP `null`; we use -1 as sentinel.
		mk("Crystal1", TypeCrystal, -1, 0x02, 0x34, 0x64, 0x40, 0x7F, 0x20),
		mk("Crystal2", TypeCrystal, -1, 0x10, 0x34, 0x64, 0x40, 0x79, 0x20),
		mk("Crystal3", TypeCrystal, -1, 0x40, 0x34, 0x64, 0x40, 0x6C, 0x20),
		mk("Crystal4", TypeCrystal, -1, 0x20, 0x34, 0x64, 0x40, 0x6D, 0x20),
		mk("Crystal5", TypeCrystal, -1, 0x04, 0x32, 0x64, 0x40, 0x6E, 0x20),
		mk("Crystal6", TypeCrystal, -1, 0x01, 0x32, 0x64, 0x40, 0x6F, 0x20),
		mk("Crystal7", TypeCrystal, -1, 0x08, 0x34, 0x64, 0x40, 0x7C, 0x20),
		mk("RescueZelda", TypeEvent, -1),
		mk("DefeatAgahnim", TypeEvent, -1),
		mk("BigRedBomb", TypeEvent, -1),
		mk("DefeatAgahnim2", TypeEvent, -1),
		mk("DefeatGanon", TypeEvent, -1),
	}
}
