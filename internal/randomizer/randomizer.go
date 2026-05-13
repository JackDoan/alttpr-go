// Package randomizer is the Go port of app/Randomizer.php — the orchestrator
// that prepares each world (pre-collected items, goal, bosses, medallions,
// fountains, shops, prize placement) and then runs the Filler.
package randomizer

import (
	"fmt"

	"github.com/JackDoan/alttpr-go/internal/boss"
	"github.com/JackDoan/alttpr-go/internal/filler"
	"github.com/JackDoan/alttpr-go/internal/helpers"
	"github.com/JackDoan/alttpr-go/internal/item"
	"github.com/JackDoan/alttpr-go/internal/world"
)

// LOGIC version constant; matches PHP Randomizer::LOGIC.
const Logic = 31

// Randomizer mirrors PHP Randomizer.
type Randomizer struct {
	worlds       []*world.World
	itemRegistry *item.Registry
	bossRegistry *boss.Registry

	advancementItems []*item.Item
	trashItems       []*item.Item
	niceItems        []*item.Item
	dungeonItems     []*item.Item
}

// New constructs a Randomizer over the given worlds. Each world receives the
// default pre-collected items (PHP Randomizer::__construct).
func New(worlds []*world.World, ir *item.Registry, br *boss.Registry) (*Randomizer, error) {
	for _, w := range worlds {
		if w.PreCollectedItems().Count() == 0 {
			defaults := []string{
				"BossHeartContainer", "BossHeartContainer", "BossHeartContainer",
				"BombUpgrade10", "ArrowUpgrade10", "ArrowUpgrade10", "ArrowUpgrade10",
			}
			for _, name := range defaults {
				it, err := ir.Get(name, w.ID())
				if err != nil {
					return nil, fmt.Errorf("seed pre-collected %s: %w", name, err)
				}
				w.AddPreCollectedItem(it)
			}
		}
	}
	return &Randomizer{worlds: worlds, itemRegistry: ir, bossRegistry: br}, nil
}

// Randomize runs the orchestrated fill pipeline.
// Mirrors PHP Randomizer::randomize (minus HintService).
func (r *Randomizer) Randomize() error {
	for _, w := range r.worlds {
		if err := r.prepareWorld(w); err != nil {
			return err
		}
	}
	f := filler.NewRandomAssumed(r.worlds...)
	if err := f.Fill(r.dungeonItems, r.advancementItems, r.niceItems, r.trashItems); err != nil {
		return err
	}
	for _, w := range r.worlds {
		if err := r.setTexts(w); err != nil {
			return err
		}
		if err := r.applyHints(w); err != nil {
			return err
		}
		if err := r.randomizeCredits(w); err != nil {
			return err
		}
	}
	return nil
}

func (r *Randomizer) prepareWorld(w *world.World) error {
	// Pre-collect RescueZelda for non-standard modes. Mirrors PHP:
	// `if ($world->config('mode.state') != 'standard') addPreCollectedItem(RescueZelda)`.
	if w.ConfigString("mode.state", "") != "standard" {
		if it, err := r.itemRegistry.Get("RescueZelda", w.ID()); err == nil {
			w.AddPreCollectedItem(it)
		}
	}

	// 1. Goal item placement.
	switch w.ConfigString("goal", "") {
	case "pedestal":
		l := w.Locations().Get("Master Sword Pedestal:" + itoaW(w.ID()))
		if l != nil {
			triforce, _ := r.itemRegistry.Get("Triforce", w.ID())
			_ = l.SetItem(triforce)
		}
	case "ganon", "fast_ganon", "dungeons", "ganonhunt", "completionist", "":
		l := w.Locations().Get("Ganon:" + itoaW(w.ID()))
		if l != nil {
			triforce, _ := r.itemRegistry.Get("Triforce", w.ID())
			_ = l.SetItem(triforce)
		}
	}

	// 2. Item pools.
	advancement, err := w.AdvancementItems(r.itemRegistry)
	if err != nil {
		return err
	}
	nice, err := w.NiceItems(r.itemRegistry)
	if err != nil {
		return err
	}
	trash, err := w.JunkItems(r.itemRegistry)
	if err != nil {
		return err
	}
	dungeon, err := w.DungeonItems(r.itemRegistry)
	if err != nil {
		return err
	}

	// 3. World setup steps (shops/bosses/medallions/fountains/prizes).
	r.setShops(w)
	if err := r.setMedallions(w); err != nil {
		return err
	}
	r.placeBosses(w)
	if err := r.fillPrizes(w, 5); err != nil {
		return err
	}
	if err := r.setFountains(w); err != nil {
		return err
	}
	if err := w.ShufflePrizePacks(); err != nil {
		return err
	}

	// 4. Sword-handling per weapons mode. Default ("randomized") path only.
	weapons := w.ConfigString("mode.weapons", "randomized")
	swords := filterItems(advancement, func(it *item.Item) bool { return it.IsType(item.TypeSword) })
	advancement = filterItems(advancement, func(it *item.Item) bool { return !it.IsType(item.TypeSword) })
	bottles := filterItems(advancement, func(it *item.Item) bool { return it.IsType(item.TypeBottle) })
	advancement = filterItems(advancement, func(it *item.Item) bool { return !it.IsType(item.TypeBottle) })

	switch weapons {
	case "swordless":
		for range swords {
			rupees, _ := r.itemRegistry.Get("TwentyRupees2", w.ID())
			nice = append(nice, rupees)
		}
	default:
		// randomized: re-insert UncleSword (as alias to first sword) + master sword.
		if len(swords) > 0 {
			uncle, err := r.itemRegistry.Get("UncleSword", w.ID())
			if err == nil {
				// In PHP this is an ItemAlias targeting the first popped sword.
				uncle.Target = swords[len(swords)-1]
				swords = swords[:len(swords)-1]
				advancement = append(advancement, uncle)
			}
		}
		if len(swords) > 0 {
			advancement = append(advancement, swords[len(swords)-1])
			swords = swords[:len(swords)-1]
		}
		if w.ConfigBool("region.requireBetterSword", false) && len(swords) > 0 {
			advancement = append(advancement, swords[len(swords)-1])
			swords = swords[:len(swords)-1]
		}
		nice = append(nice, swords...)
	}
	// Put one bottle back into advancement; rest go to nice.
	if len(bottles) > 0 {
		advancement = append(advancement, bottles[len(bottles)-1])
		bottles = bottles[:len(bottles)-1]
	}
	nice = append(nice, bottles...)

	// 5. Wild* options promote dungeon items into advancement.
	keepDungeon := []*item.Item{}
	for _, it := range dungeon {
		switch {
		case it.IsType(item.TypeBigKey) && w.ConfigBool("region.wildBigKeys", false):
			advancement = append(advancement, it)
		case it.IsType(item.TypeKey) && w.ConfigBool("region.wildKeys", false):
			isSewerH2 := it.Name == "KeyH2" && w.ConfigString("mode.state", "") == "standard" && w.ConfigString("logic", "") != "NoLogic"
			if isSewerH2 {
				keepDungeon = append(keepDungeon, it)
			} else {
				advancement = append(advancement, it)
			}
		case it.IsType(item.TypeMap) && w.ConfigBool("region.wildMaps", false):
			advancement = append(advancement, it)
		case it.IsType(item.TypeCompass) && w.ConfigBool("region.wildCompasses", false):
			advancement = append(advancement, it)
		default:
			keepDungeon = append(keepDungeon, it)
		}
	}
	dungeon = keepDungeon

	// 6. Shuffle and accumulate per-world contributions.
	advancement, err = helpers.FyShuffle(advancement)
	if err != nil {
		return err
	}
	nice, err = helpers.FyShuffle(nice)
	if err != nil {
		return err
	}
	trash, err = helpers.FyShuffle(trash)
	if err != nil {
		return err
	}

	r.dungeonItems = append(r.dungeonItems, dungeon...)
	r.advancementItems = append(r.advancementItems, advancement...)
	r.niceItems = append(r.niceItems, nice...)
	r.trashItems = append(r.trashItems, trash...)
	return nil
}

// setShops activates non-take-any shops. Mirrors PHP Randomizer::setShops (basic path).
func (r *Randomizer) setShops(w *world.World) {
	for _, s := range w.Shops().All() {
		if s.Kind != world.ShopTakeAny {
			s.Active = true
		}
	}
}

// setMedallions randomly assigns Ether/Bombos/Quake to the two medallion slots.
// Mirrors PHP Randomizer::setMedallions.
func (r *Randomizer) setMedallions(w *world.World) error {
	medallions := []string{"Ether", "Bombos", "Quake"}
	region := w.Region("Medallions")
	if region == nil {
		return nil
	}
	for _, l := range region.Locations.All() {
		if l.HasItem() {
			continue
		}
		idx, err := helpers.GetRandomInt(0, 2)
		if err != nil {
			return err
		}
		it, err := r.itemRegistry.Get(medallions[idx], w.ID())
		if err != nil {
			return err
		}
		if err := l.SetItem(it); err != nil {
			return err
		}
	}
	return nil
}

// setFountains assigns random bottles to fountain locations.
// Mirrors PHP Randomizer::setFountains.
func (r *Randomizer) setFountains(w *world.World) error {
	region := w.Region("Fountains")
	if region == nil {
		return nil
	}
	for _, l := range region.Locations.All() {
		if l.HasItem() {
			continue
		}
		b, err := w.GetBottle(r.itemRegistry)
		if err != nil {
			return err
		}
		if err := l.SetItem(b); err != nil {
			return err
		}
	}
	return nil
}

// placeBosses applies vanilla boss layout (PHP "none" shuffle case).
// Mirrors PHP Randomizer::placeBosses default branch.
func (r *Randomizer) placeBosses(w *world.World) {
	set := func(regionName, bossName string, level string) {
		region := w.Region(regionName)
		if region == nil {
			return
		}
		b, err := r.bossRegistry.Get(bossName, w)
		if err != nil {
			return
		}
		if level == "" {
			region.Boss = b
		} else {
			region.Bosses[level] = b
		}
	}
	set("Eastern Palace", "Armos Knights", "")
	set("Desert Palace", "Lanmolas", "")
	set("Tower of Hera", "Moldorm", "")
	set("Palace of Darkness", "Helmasaur King", "")
	set("Swamp Palace", "Arrghus", "")
	set("Skull Woods", "Mothula", "")
	set("Thieves Town", "Blind", "")
	set("Ice Palace", "Kholdstare", "")
	set("Misery Mire", "Vitreous", "")
	set("Turtle Rock", "Trinexx", "")
	set("Ganons Tower", "Armos Knights", "bottom")
	set("Ganons Tower", "Lanmolas", "middle")
	set("Ganons Tower", "Moldorm", "top")
	set("Hyrule Castle Tower", "Agahnim", "")
	set("Ganons Tower", "Agahnim2", "")
}

// fillPrizes places crystals/pendants. Faithful port of PHP
// Randomizer::fillPrizes: vanilla-fills if shuffle flags are off, then runs
// accessibility-aware placement honoring `prize.crossWorld` (when true,
// pendants and crystals may swap slot types).
//
// `attempts` is the remaining retry budget — on a failed placement the
// function clears the relevant locations and recurses with attempts-1.
func (r *Randomizer) fillPrizes(w *world.World, attempts int) error {
	prizeLocs := w.Locations().Filter(func(l *world.Location) bool {
		return l.Kind == world.KindPrize || l.Kind == world.KindPrizeCrystal || l.Kind == world.KindPrizePendant
	})
	crystalLocs := prizeLocs.Filter(func(l *world.Location) bool { return l.Kind == world.KindPrizeCrystal })
	pendantLocs := prizeLocs.Filter(func(l *world.Location) bool { return l.Kind == world.KindPrizePendant })

	// Vanilla fills when shuffle flags are off.
	if !w.ConfigBool("prize.shuffleCrystals", true) {
		mapping := map[string]string{
			"Palace of Darkness - Prize:": "Crystal1",
			"Swamp Palace - Prize:":       "Crystal2",
			"Skull Woods - Prize:":        "Crystal3",
			"Thieves' Town - Prize:":      "Crystal4",
			"Ice Palace - Prize:":         "Crystal5",
			"Misery Mire - Prize:":        "Crystal6",
			"Turtle Rock - Prize:":        "Crystal7",
		}
		for prefix, itemName := range mapping {
			l := w.Locations().Get(prefix + itoaW(w.ID()))
			if l == nil {
				continue
			}
			it, err := r.itemRegistry.Get(itemName, w.ID())
			if err != nil {
				return err
			}
			if err := l.SetItem(it); err != nil {
				return err
			}
		}
	}
	if !w.ConfigBool("prize.shufflePendants", true) {
		mapping := map[string]string{
			"Eastern Palace - Prize:": "PendantOfCourage",
			"Desert Palace - Prize:":  "PendantOfPower",
			"Tower of Hera - Prize:":  "PendantOfWisdom",
		}
		for prefix, itemName := range mapping {
			l := w.Locations().Get(prefix + itoaW(w.ID()))
			if l == nil {
				continue
			}
			it, err := r.itemRegistry.Get(itemName, w.ID())
			if err != nil {
				return err
			}
			if err := l.SetItem(it); err != nil {
				return err
			}
		}
	}

	// Build the full prize set; subtract already-placed (vanilla-filled) items.
	allPrizes := []*item.Item{}
	for _, name := range []string{
		"Crystal1", "Crystal2", "Crystal3", "Crystal4", "Crystal5", "Crystal6", "Crystal7",
		"PendantOfCourage", "PendantOfPower", "PendantOfWisdom",
	} {
		it, err := r.itemRegistry.Get(name, w.ID())
		if err != nil {
			return err
		}
		allPrizes = append(allPrizes, it)
	}
	placedSet := map[string]bool{}
	for _, l := range prizeLocs.NonEmpty().All() {
		placedSet[l.Item().FullName()] = true
	}
	remaining := []*item.Item{}
	for _, it := range allPrizes {
		if !placedSet[it.FullName()] {
			remaining = append(remaining, it)
		}
	}
	remaining, err := helpers.FyShuffle(remaining)
	if err != nil {
		return err
	}

	crossWorld := w.ConfigBool("prize.crossWorld", true)

	// --- Crystal-typed locations ---
	placePool := remaining
	if !crossWorld {
		placePool = filterItems(remaining, func(it *item.Item) bool { return it.IsType(item.TypeCrystal) })
	}
	leftover, err := r.fillPrizeSlots(w, crystalLocs.Empty().All(), placePool)
	if err != nil {
		if attempts > 0 {
			for _, l := range crystalLocs.Empty().All() {
				_ = l.SetItem(nil)
			}
			return r.fillPrizes(w, attempts-1)
		}
		return err
	}
	if crystalLocs.Empty().Count() > 0 {
		if attempts > 0 {
			for _, l := range crystalLocs.Empty().All() {
				_ = l.SetItem(nil)
			}
			return r.fillPrizes(w, attempts-1)
		}
		return fmt.Errorf("fillPrizes: could not fill crystal slots")
	}

	// --- Pendant-typed locations ---
	// crossWorld=true: continue with whatever's left after crystal placement.
	// crossWorld=false: re-derive from the original remaining pool (pendants only).
	placePool = leftover
	if !crossWorld {
		placePool = filterItems(remaining, func(it *item.Item) bool { return it.IsType(item.TypePendant) })
	}
	if _, err := r.fillPrizeSlots(w, pendantLocs.Empty().All(), placePool); err != nil {
		if attempts > 0 {
			for _, l := range pendantLocs.Empty().All() {
				_ = l.SetItem(nil)
			}
			return r.fillPrizes(w, attempts-1)
		}
		return err
	}
	if pendantLocs.Empty().Count() > 0 {
		if attempts > 0 {
			for _, l := range pendantLocs.Empty().All() {
				_ = l.SetItem(nil)
			}
			return r.fillPrizes(w, attempts-1)
		}
		return fmt.Errorf("fillPrizes: could not fill pendant slots")
	}
	return nil
}

// fillPrizeSlots assigns items from `pool` into `locs` using the PHP
// "assumed items" model: for each empty location, walk through the pool
// (pop, accessibility-check the location with the rest of the pool plus
// dungeon/advancement items, place if accessible, else rotate). Returns
// the remaining unused prizes.
func (r *Randomizer) fillPrizeSlots(w *world.World, locs []*world.Location, pool []*item.Item) ([]*item.Item, error) {
	dungeon, err := w.DungeonItems(r.itemRegistry)
	if err != nil {
		return pool, err
	}
	advancement, err := w.AdvancementItems(r.itemRegistry)
	if err != nil {
		return pool, err
	}

	// Pool is mutated as we pop / rotate. Copy to avoid aliasing the caller.
	work := append([]*item.Item{}, pool...)

	// buildStarting collects dungeon + advancement + the current pool into a
	// starting collection. Called repeatedly because the pool mutates.
	buildStarting := func() *item.Collection {
		s := item.NewCollection()
		for _, it := range dungeon {
			s.Add(it)
		}
		for _, it := range advancement {
			s.Add(it)
		}
		for _, it := range work {
			s.Add(it)
		}
		return s
	}

	for _, loc := range locs {
		var placedPrize *item.Item
		total := len(work)
		for i := 0; i < total; i++ {
			// Pop from the back (PHP array_pop).
			candidate := work[len(work)-1]
			work = work[:len(work)-1]

			assumed := w.CollectItemsFrom(buildStarting())
			if loc.CanAccess(assumed, w.Locations()) {
				placedPrize = candidate
				break
			}
			// Rotate: put back to front of queue and try the next.
			work = append([]*item.Item{candidate}, work...)
		}
		if placedPrize == nil {
			continue
		}
		if err := loc.SetItem(placedPrize); err != nil {
			return work, err
		}
		// Recompute the closure with the placement in effect (the placed
		// location is now part of the walk, so its item gets re-collected).
		// Mirrors PHP $world->checkWinCondition($assumed_items) → which calls
		// collectItems($assumed_items) internally.
		final := w.CollectItemsFrom(buildStarting())
		if w.WinCondition != nil && !w.WinCondition(final) {
			return work, fmt.Errorf("fillPrizes: unwinnable after placing %s in %s", placedPrize.Name, loc.Name)
		}
	}
	return work, nil
}

// filterItems returns a new slice of items where keep returns true.
func filterItems(items []*item.Item, keep func(*item.Item) bool) []*item.Item {
	out := make([]*item.Item, 0, len(items))
	for _, it := range items {
		if keep(it) {
			out = append(out, it)
		}
	}
	return out
}

// itoaW: package-local stringer for full-name lookups.
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
