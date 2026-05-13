package world

import (
	"github.com/JackDoan/alttpr-go/internal/helpers"
	"github.com/JackDoan/alttpr-go/internal/item"
)

// poolEntry is a single (item_name, count) tuple as defined by PHP config/item.php.
type poolEntry struct {
	Name  string
	Count int
}

// Pool tables — mirror PHP config/item.php:advancement/nice/junk/dungeon.
var (
	advancementPool = []poolEntry{
		{"L1Sword", 0}, {"MasterSword", 0}, {"ProgressiveSword", 4},
		{"BossHeartContainer", 10}, {"BottleWithRandom", 4},
		{"Bottle", 0}, {"BottleWithRedPotion", 0}, {"BottleWithGreenPotion", 0},
		{"BottleWithBluePotion", 0}, {"BottleWithBee", 0}, {"BottleWithGoldBee", 0},
		{"BottleWithFairy", 0},
		{"Bombos", 1}, {"BookOfMudora", 1}, {"Bow", 0}, {"BowAndArrows", 0},
		{"CaneOfSomaria", 1}, {"Cape", 1}, {"Ether", 1}, {"FireRod", 1},
		{"Flippers", 1}, {"Hammer", 1}, {"Hookshot", 1}, {"IceRod", 1},
		{"Lamp", 1}, {"MagicMirror", 1}, {"MoonPearl", 1}, {"Mushroom", 1},
		{"OcarinaInactive", 1}, {"OcarinaActive", 0}, {"PegasusBoots", 1},
		{"Powder", 1}, {"PowerGlove", 0}, {"Quake", 1}, {"Shovel", 1},
		{"TitansMitt", 0}, {"BowAndSilverArrows", 0}, {"SilverArrowUpgrade", 0},
		{"ProgressiveGlove", 2}, {"ProgressiveBow", 2},
		{"Triforce", 0}, {"TriforcePiece", 0}, {"PowerStar", 0},
		{"BugCatchingNet", 1}, {"MirrorShield", 0}, {"ProgressiveShield", 3},
		{"CaneOfByrna", 1}, {"TenBombs", 1}, {"HalfMagic", 1}, {"QuarterMagic", 0},
	}
	nicePool = []poolEntry{
		{"L3Sword", 0}, {"L4Sword", 0}, {"HeartContainer", 1},
		{"BlueShield", 0}, {"ProgressiveArmor", 2}, {"BlueMail", 0},
		{"Boomerang", 1}, {"RedBoomerang", 1}, {"RedShield", 0}, {"RedMail", 0},
		{"BlueClock", 0}, {"RedClock", 0}, {"GreenClock", 0},
		{"OneHundredRupees", 1}, {"ThreeHundredRupees", 5},
	}
	junkPool = []poolEntry{
		{"PieceOfHeart", 24},
		{"BombUpgrade5", 0}, {"BombUpgrade10", 0},
		{"ArrowUpgrade5", 0}, {"ArrowUpgrade10", 0},
		{"Arrow", 1}, {"TenArrows", 12},
		{"Bomb", 0}, {"ThreeBombs", 16},
		{"OneRupee", 2}, {"FiveRupees", 4}, {"TwentyRupees", 28},
		{"TwentyRupees2", 0}, {"FiftyRupees", 7},
		{"Heart", 0}, {"SmallMagic", 0}, {"Rupoor", 0},
	}
	dungeonPool = []poolEntry{
		{"BigKeyA2", 1}, {"BigKeyD1", 1}, {"BigKeyD2", 1}, {"BigKeyD3", 1},
		{"BigKeyD4", 1}, {"BigKeyD5", 1}, {"BigKeyD6", 1}, {"BigKeyD7", 1},
		{"BigKeyA1", 0}, {"BigKeyH2", 0},
		{"BigKeyP1", 1}, {"BigKeyP2", 1}, {"BigKeyP3", 1},
		{"KeyA2", 4}, {"KeyD1", 6}, {"KeyD2", 1}, {"KeyD3", 3},
		{"KeyD4", 1}, {"KeyD5", 2}, {"KeyD6", 3}, {"KeyD7", 4},
		{"KeyA1", 2}, {"KeyH2", 1},
		{"KeyP1", 0}, {"KeyP2", 1}, {"KeyP3", 1},
		{"MapA2", 1}, {"MapD1", 1}, {"MapD2", 1}, {"MapD3", 1},
		{"MapD4", 1}, {"MapD5", 1}, {"MapD6", 1}, {"MapD7", 1},
		{"MapA1", 0}, {"MapH2", 1},
		{"MapP1", 1}, {"MapP2", 1}, {"MapP3", 1},
		{"CompassA2", 1}, {"CompassD1", 1}, {"CompassD2", 1}, {"CompassD3", 1},
		{"CompassD4", 1}, {"CompassD5", 1}, {"CompassD6", 1}, {"CompassD7", 1},
		{"CompassA1", 0}, {"CompassH2", 0},
		{"CompassP1", 1}, {"CompassP2", 1}, {"CompassP3", 1},
	}
)

// GetBottle returns a random bottle item. Mirrors PHP World::getBottle.
func (w *World) GetBottle(ir *item.Registry) (*item.Item, error) {
	bottles := []string{
		"Bottle", "BottleWithRedPotion", "BottleWithGreenPotion",
		"BottleWithBluePotion", "BottleWithBee", "BottleWithGoldBee", "BottleWithFairy",
	}
	idx, err := helpers.GetRandomInt(0, len(bottles)-1)
	if err != nil {
		return nil, err
	}
	return ir.Get(bottles[idx], w.id)
}

// AdvancementItems returns the world's advancement-pool items, applying
// per-name count overrides from `item.count.<Name>` config keys (clamped to 216).
// Mirrors PHP World::getAdvancementItems.
func (w *World) AdvancementItems(ir *item.Registry) ([]*item.Item, error) {
	return w.expandPool(advancementPool, ir)
}

// NiceItems mirrors PHP World::getNiceItems.
func (w *World) NiceItems(ir *item.Registry) ([]*item.Item, error) {
	return w.expandPool(nicePool, ir)
}

// JunkItems mirrors PHP World::getItemPool (junk/trash).
func (w *World) JunkItems(ir *item.Registry) ([]*item.Item, error) {
	return w.expandPool(junkPool, ir)
}

// DungeonItems mirrors PHP World::getDungeonPool.
func (w *World) DungeonItems(ir *item.Registry) ([]*item.Item, error) {
	return w.expandPool(dungeonPool, ir)
}

func (w *World) expandPool(pool []poolEntry, ir *item.Registry) ([]*item.Item, error) {
	out := []*item.Item{}
	for _, e := range pool {
		n := min(w.ConfigInt("item.count."+e.Name, e.Count), 216)
		for range n {
			var it *item.Item
			var err error
			if e.Name == "BottleWithRandom" {
				it, err = w.GetBottle(ir)
			} else {
				it, err = ir.Get(e.Name, w.id)
			}
			if err != nil {
				return nil, err
			}
			out = append(out, it)
		}
	}
	return out, nil
}
