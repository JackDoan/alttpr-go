package world

import (
	"github.com/JackDoan/alttpr-go/internal/boss"
	"github.com/JackDoan/alttpr-go/internal/item"
)

// StandardOptions configures a Standard-mode World.
type StandardOptions struct {
	ID            int
	Goal          string // ganon, fast_ganon, dungeons, ganonhunt, triforce-hunt, pedestal, completionist
	Logic         string // NoGlitches, ...
	Glitches      string // none, ...
	ItemPlacement string // basic, advanced
	DungeonItems  string // standard, ...
	Accessibility string // item, locations, none
	Weapons       string // randomized, swordless, ...
	CrystalsGanon int    // 0..7
	CrystalsTower int    // 0..7
	ItemPool      string // normal
	ItemFunc      string // normal
	Hints         string // on/off
	Tournament    bool

	// TriforcePieces is the count placed in the advancement pool for
	// goal=triforce-hunt or goal=ganonhunt. Ignored for other goals.
	TriforcePieces int
	// TriforceGoal is the count required to win. Must be <= TriforcePieces.
	TriforceGoal int

	// ShufflePrizes = prize.crossWorld (mix pendants/crystals across slot types).
	// ShuffleCrystals = prize.shuffleCrystals (randomize among the 7 crystal slots).
	// ShufflePendants = prize.shufflePendants (randomize among the 3 pendant slots).
	// All three default to true (matches the PHP UI defaults).
	ShufflePrizes   bool
	ShuffleCrystals bool
	ShufflePendants bool
}

// DefaultStandardOptions returns the PHP-equivalent default-config Standard world.
// Mirrors the call-site in app/Console/Commands/Randomize.php.
func DefaultStandardOptions() StandardOptions {
	return StandardOptions{
		ID: 0,
		Goal: "ganon", Logic: "NoGlitches", Glitches: "none",
		ItemPlacement: "basic", DungeonItems: "standard", Accessibility: "item",
		Weapons: "randomized", CrystalsGanon: 7, CrystalsTower: 7,
		ItemPool: "normal", ItemFunc: "normal", Hints: "on",
		ShufflePrizes:   true,
		ShuffleCrystals: true,
		ShufflePendants: true,
	}
}

// NewOpen wires up an Open-mode World. Differs from Standard only in the
// Escape region (Zelda already saved) and pre-opened castle gate (handled
// via SRAM in WriteToRom).
func NewOpen(opts StandardOptions, ir *item.Registry, br *boss.Registry) *World {
	w := newStandardWorldBase(opts, "open", ir, br)
	return w
}

// NewStandard wires up a Standard-mode World with all 26 ported regions.
func NewStandard(opts StandardOptions, ir *item.Registry, br *boss.Registry) *World {
	return newStandardWorldBase(opts, "standard", ir, br)
}

// newStandardWorldBase constructs a Standard-layout world, optionally with
// the Open-mode Escape region overridden. `state` is "standard" or "open".
func newStandardWorldBase(opts StandardOptions, state string, ir *item.Registry, br *boss.Registry) *World {
	cfg := NewConfig()
	cfg.Strings["mode.state"] = state
	cfg.Strings["world.variant"] = "standard"
	cfg.Strings["goal"] = opts.Goal
	cfg.Strings["logic"] = opts.Logic
	cfg.Strings["glitches"] = opts.Glitches
	cfg.Strings["itemPlacement"] = opts.ItemPlacement
	cfg.Strings["dungeonItems"] = opts.DungeonItems
	cfg.Strings["accessibility"] = opts.Accessibility
	cfg.Strings["mode.weapons"] = opts.Weapons
	cfg.Strings["item.pool"] = opts.ItemPool
	cfg.Strings["item.functionality"] = opts.ItemFunc
	cfg.Strings["spoil.Hints"] = opts.Hints
	cfg.Ints["crystals.ganon"] = opts.CrystalsGanon
	cfg.Ints["crystals.tower"] = opts.CrystalsTower
	cfg.Ints["item.require.Lamp"] = 1
	cfg.Ints["rom.BottleFill.Magic"] = 0x80

	// Triforce-hunt / Ganon-hunt: add TriforcePieces to the pool and set
	// the goal count. PHP wires these via `item.count.TriforcePiece` and
	// `item.Goal.Required` config keys.
	if opts.Goal == "triforce-hunt" || opts.Goal == "ganonhunt" {
		if opts.TriforcePieces > 0 {
			cfg.Ints["item.count.TriforcePiece"] = opts.TriforcePieces
		}
		if opts.TriforceGoal > 0 {
			cfg.Ints["item.Goal.Required"] = opts.TriforceGoal
		}
	}
	cfg.Bools["region.swordsInPool"] = true
	cfg.Bools["region.bossNormalLocation"] = true
	cfg.Bools["rom.CatchableFairies"] = true
	cfg.Bools["tournament"] = opts.Tournament
	cfg.Bools["prize.crossWorld"] = opts.ShufflePrizes
	cfg.Bools["prize.shuffleCrystals"] = opts.ShuffleCrystals
	cfg.Bools["prize.shufflePendants"] = opts.ShufflePendants

	w := NewWorld(opts.ID, cfg)

	b := newBuilder(w, ir, br)

	// Pick the Escape variant per game state.
	escape := newStandardHyruleCastleEscape(b)
	if state == "open" {
		escape = newOpenHyruleCastleEscape(b)
	}

	// Construct regions in dependency-friendly order (a region's closures
	// look up other regions lazily, so any order is fine for construction,
	// but we add them to the world here in a stable PHP-matching order).
	regions := []*Region{
		newStandardFountains(b),
		newStandardMedallions(b),
		escape,
		newStandardHyruleCastleTower(b),
		newStandardEasternPalace(b),
		newStandardDesertPalace(b),
		newStandardTowerOfHera(b),
		newStandardPalaceOfDarkness(b),
		newStandardSwampPalace(b),
		newStandardSkullWoods(b),
		newStandardThievesTown(b),
		newStandardIcePalace(b),
		newStandardMiseryMire(b),
		newStandardTurtleRock(b),
		newStandardGanonsTower(b),
		newStandardLWWestDM(b),
		newStandardLWEastDM(b),
		newStandardLWNorthEast(b),
		newStandardLWNorthWest(b),
		newStandardLWSouth(b),
		newStandardDWWestDM(b),
		newStandardDWEastDM(b),
		newStandardDWMire(b),
		newStandardDWSouth(b),
		newStandardDWNorthWest(b),
		newStandardDWNorthEast(b),
	}
	for _, r := range regions {
		w.AddRegion(r)
	}

	// Win condition. Mirrors PHP World::$win_condition: a collection
	// either holds Triforce (set when Ganon is defeated or pedestal/Murahdahla
	// is reached), or — in triforce-hunt — North East Light World is
	// reachable AND enough TriforcePieces are collected.
	w.WinCondition = func(items *item.Collection) bool {
		items.SetChecksForWorld(w.ID())
		if items.Has1("DefeatGanon") || items.Has1("Triforce") {
			return true
		}
		if w.ConfigString("goal", "") == "triforce-hunt" {
			need := w.ConfigInt("item.Goal.Required", 0)
			ne := w.Region("North East Light World")
			if need > 0 && ne != nil &&
				ne.CanEnter(w.Locations(), items) &&
				items.Has("TriforcePiece", need) {
				return true
			}
		}
		return false
	}

	// Ganon's Tower junk-fill range mirrors PHP defaults.
	w.GanonsTowerJunkFillLow = 0
	w.GanonsTowerJunkFillHigh = 15

	return w
}
