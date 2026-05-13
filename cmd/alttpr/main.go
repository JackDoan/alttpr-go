package main

import (
	"flag"
	"fmt"
	"os"

	"github.com/JackDoan/alttpr-go/internal/job"
)

func main() {
	fs := flag.NewFlagSet("alttpr", flag.ContinueOnError)
	fs.Usage = func() {
		fmt.Fprintf(os.Stderr, `alttpr — A Link to the Past ROM tool (Go port)

Usage:
  alttpr [flags] <input.sfc> <output_dir>

Modes:
  --unrandomized      Apply only the base patch + QoL options; no randomization.
  (default)           Run Standard-mode randomization in-memory; write spoiler JSON.

Flags:
`)
		fs.PrintDefaults()
	}

	// Phase 1 flags (apply to both modes).
	unrandomized := fs.Bool("unrandomized", false, "do not apply randomization to the ROM")
	skipMD5 := fs.Bool("skip-md5", false, "do not validate md5 of base ROM")
	heartColor := fs.String("heartcolor", "red", "set heart color (red, blue, green, yellow, random)")
	heartBeep := fs.String("heartbeep", "half", "set heart beep speed (off, normal, half, quarter, double)")
	menuSpeed := fs.String("menu-speed", "normal", "menu speed (slow, normal, fast, instant)")
	noMusic := fs.Bool("no-music", false, "mute all music")
	quickswap := fs.String("quickswap", "false", "set quickswap (true/false)")
	basePatch := fs.String("base-patch", "", "path to base patch JSON (default: embedded; falls back to storage/patches/<HASH>.json)")

	// Phase 2 randomizer flags.
	goal := fs.String("goal", "ganon", "set game goal (ganon, fast_ganon, dungeons, ganonhunt, triforce-hunt, pedestal, completionist)")
	state := fs.String("state", "standard", "set game state (standard, open)")
	weapons := fs.String("weapons", "randomized", "set weapons mode (randomized, swordless, assured, vanilla)")
	glitches := fs.String("glitches", "none", "set glitches (none, overworld_glitches, hybrid_major_glitches, major_glitches, no_logic)")
	itemPlacement := fs.String("item-placement", "basic", "set item placement (basic, advanced)")
	dungeonItems := fs.String("dungeon-items", "standard", "set dungeon-item placement")
	accessibility := fs.String("accessibility", "item", "set accessibility (item, locations, none)")
	itemPool := fs.String("item-pool", "normal", "set item pool (normal, hard, expert, superexpert, crowd_control)")
	itemFunc := fs.String("item-functionality", "normal", "set item functionality")
	hints := fs.String("hints", "on", "set hints (on, off)")
	crystalsGanon := fs.Int("crystals-ganon", 7, "set ganon crystal requirement (0-7)")
	crystalsTower := fs.Int("crystals-tower", 7, "set ganon tower crystal requirement (0-7)")
	tournament := fs.Bool("tournament", false, "enable tournament mode")
	spoiler := fs.Bool("spoiler", false, "emit a spoiler JSON alongside the output (always emitted in randomized mode)")
	noRom := fs.Bool("no-rom", false, "do not generate output ROM (useful with --spoiler)")

	triforcePieces := fs.Int("triforce-pieces", 30, "number of TriforcePieces placed in the pool")
	triforceGoal := fs.Int("triforce-goal", 20, "number of TriforcePieces required to win")

	shufflePrizes := fs.Bool("shuffle-prizes", true, "allow pendants and crystals to swap slot types (prize.crossWorld)")
	shuffleCrystals := fs.Bool("shuffle-crystals", true, "randomize the 7 crystals across crystal-type prize slots")
	shufflePendants := fs.Bool("shuffle-pendants", true, "randomize the 3 pendants across pendant-type prize slots")

	if err := fs.Parse(os.Args[1:]); err != nil {
		os.Exit(2)
	}
	if fs.NArg() != 2 {
		fs.Usage()
		os.Exit(2)
	}

	opts := job.Options{
		BaseROMPath:    fs.Arg(0),
		OutputDir:      fs.Arg(1),
		SkipMD5:        *skipMD5,
		BasePatch:      *basePatch,
		Unrandomized:   *unrandomized,
		Goal:           *goal,
		State:          *state,
		Weapons:        *weapons,
		Glitches:       *glitches,
		ItemPlacement:  *itemPlacement,
		DungeonItems:   *dungeonItems,
		Accessibility:  *accessibility,
		ItemPool:       *itemPool,
		ItemFunc:       *itemFunc,
		Hints:          *hints,
		CrystalsGanon:  *crystalsGanon,
		CrystalsTower:  *crystalsTower,
		Tournament:      *tournament,
		TriforcePieces:  *triforcePieces,
		TriforceGoal:    *triforceGoal,
		ShufflePrizes:   *shufflePrizes,
		ShuffleCrystals: *shuffleCrystals,
		ShufflePendants: *shufflePendants,
		HeartColor:     *heartColor,
		HeartBeep:      *heartBeep,
		MenuSpeed:      *menuSpeed,
		NoMusic:        *noMusic,
		Quickswap:      *quickswap,
		NoROM:          *noRom,
		// In randomized mode the spoiler is always written, matching prior behavior.
		WantSpoiler: *spoiler || !*unrandomized,
	}

	res, err := job.Run(opts)
	if err != nil {
		fmt.Fprintf(os.Stderr, "error: %v\n", err)
		os.Exit(1)
	}
	if res.SpoilerPath != "" {
		fmt.Printf("Spoiler Saved: %s\n", res.SpoilerPath)
	}
	if res.ROMPath != "" {
		fmt.Printf("ROM Saved: %s\n", res.ROMPath)
	}
}
