// Package job is the shared orchestration layer between the desktop CLI
// (cmd/alttpr) and the on-device harness (cmd/alttpr-brick). It owns the
// "open base ROM → ensure it's patched → run randomizer → write to ROM → save"
// pipeline so both binaries call the same code.
package job

import (
	"encoding/json"
	"fmt"
	"os"
	"path/filepath"
	"time"

	"github.com/JackDoan/alttpr-go/internal/boss"
	"github.com/JackDoan/alttpr-go/internal/helpers"
	"github.com/JackDoan/alttpr-go/internal/item"
	"github.com/JackDoan/alttpr-go/internal/patch"
	"github.com/JackDoan/alttpr-go/internal/randomizer"
	"github.com/JackDoan/alttpr-go/internal/rom"
	"github.com/JackDoan/alttpr-go/internal/world"
)

// Options is the union of every knob the existing CLI exposes. Callers
// fill in only the fields they care about; the defaults below match
// world.DefaultStandardOptions + the Phase-1 QoL defaults.
type Options struct {
	// Required.
	BaseROMPath string
	OutputDir   string

	// Source-side.
	SkipMD5      bool
	BasePatch    string // override path; empty = embedded
	Unrandomized bool   // skip randomization, just apply QoL

	// Randomizer.
	Goal           string
	State          string // standard, open
	Weapons        string
	Glitches       string
	ItemPlacement  string
	DungeonItems   string
	Accessibility  string
	ItemPool       string
	ItemFunc       string
	Hints          string
	CrystalsGanon  int
	CrystalsTower  int
	Tournament     bool
	TriforcePieces int
	TriforceGoal   int

	// Prize shuffle options.
	// ShufflePrizes  = "Swap Pendants and Crystals Cross World" (prize.crossWorld).
	//                  When true, any prize item can go in any prize slot.
	//                  When false, crystals stay in crystal slots and pendants
	//                  stay in pendant slots.
	// ShuffleCrystals = prize.shuffleCrystals (vanilla placement when false).
	// ShufflePendants = prize.shufflePendants (vanilla placement when false).
	ShufflePrizes   bool
	ShuffleCrystals bool
	ShufflePendants bool

	// QoL / cosmetic.
	HeartColor string
	HeartBeep  string
	MenuSpeed  string
	NoMusic    bool
	Quickswap  string

	// Output.
	NoROM       bool // do not write a ROM file (useful with WantSpoiler)
	WantSpoiler bool // write spoiler JSON alongside the ROM
}

// DefaultOptions returns sane defaults matching cmd/alttpr's flag defaults.
func DefaultOptions() Options {
	return Options{
		Goal:           "ganon",
		State:          "standard",
		Weapons:        "randomized",
		Glitches:       "none",
		ItemPlacement:  "basic",
		DungeonItems:   "standard",
		Accessibility:  "item",
		ItemPool:       "normal",
		ItemFunc:       "normal",
		Hints:          "on",
		CrystalsGanon:  7,
		CrystalsTower:  7,
		TriforcePieces: 30,
		TriforceGoal:   20,
		HeartColor:      "red",
		HeartBeep:       "half",
		MenuSpeed:       "normal",
		Quickswap:       "false",
		ShufflePrizes:   true,
		ShuffleCrystals: true,
		ShufflePendants: true,
	}
}

// Result reports what files were written.
type Result struct {
	ROMPath     string
	SpoilerPath string
}

// Run executes the randomizer pipeline end-to-end and returns the paths of
// the files written. It is a refactor of cmd/alttpr/main.go:runRandomized;
// see that function's history for the per-step rationale.
func Run(o Options) (Result, error) {
	var res Result

	if err := o.validate(); err != nil {
		return res, err
	}
	if o.Unrandomized {
		return runUnrandomized(o)
	}

	ir := item.NewRegistry()
	br := boss.NewRegistry()
	wopts := world.StandardOptions{
		Goal: o.Goal, Logic: "NoGlitches", Glitches: o.Glitches,
		ItemPlacement: o.ItemPlacement, DungeonItems: o.DungeonItems, Accessibility: o.Accessibility,
		Weapons: o.Weapons, CrystalsGanon: o.CrystalsGanon, CrystalsTower: o.CrystalsTower,
		ItemPool: o.ItemPool, ItemFunc: o.ItemFunc, Hints: o.Hints,
		Tournament:      o.Tournament,
		TriforcePieces:  o.TriforcePieces,
		TriforceGoal:    o.TriforceGoal,
		ShufflePrizes:   o.ShufflePrizes,
		ShuffleCrystals: o.ShuffleCrystals,
		ShufflePendants: o.ShufflePendants,
	}
	if o.Goal == "triforce-hunt" || o.Goal == "ganonhunt" {
		if o.TriforceGoal > o.TriforcePieces {
			return res, fmt.Errorf("triforce-goal (%d) > triforce-pieces (%d): unwinnable",
				o.TriforceGoal, o.TriforcePieces)
		}
	}

	var w *world.World
	if o.State == "open" {
		w = world.NewOpen(wopts, ir, br)
	} else {
		w = world.NewStandard(wopts, ir, br)
	}
	rnd, err := randomizer.New([]*world.World{w}, ir, br)
	if err != nil {
		return res, err
	}
	if err := rnd.Randomize(); err != nil {
		return res, fmt.Errorf("randomize: %w", err)
	}

	hashSuffix := fmt.Sprintf("%d", time.Now().UnixNano()%100000000)
	stem := fmt.Sprintf("alttpr_%s_%s_%s_%s", o.Glitches, o.State, o.Goal, hashSuffix)

	if o.WantSpoiler {
		path := filepath.Join(o.OutputDir, stem+".json")
		data, err := json.MarshalIndent(w.GetSpoiler(), "", "  ")
		if err != nil {
			return res, err
		}
		if err := os.WriteFile(path, data, 0o644); err != nil {
			return res, err
		}
		res.SpoilerPath = path
	}

	if o.NoROM {
		return res, nil
	}

	romPath, err := buildROM(o, w, ir, stem)
	if err != nil {
		return res, err
	}
	res.ROMPath = romPath
	return res, nil
}

func (o *Options) validate() error {
	if info, err := os.Stat(o.BaseROMPath); err != nil || info.IsDir() {
		return fmt.Errorf("base ROM not readable: %s", o.BaseROMPath)
	}
	if info, err := os.Stat(o.OutputDir); err != nil || !info.IsDir() {
		return fmt.Errorf("output directory not writable: %s", o.OutputDir)
	}
	if !o.Unrandomized {
		if o.State != "standard" && o.State != "open" {
			return fmt.Errorf("supported states are standard, open (got %q)", o.State)
		}
		if o.Glitches != "none" {
			return fmt.Errorf("only --glitches=none is supported (got %q)", o.Glitches)
		}
	}
	return nil
}

// resolvePatchEntries returns the patch entry list, preferring a user-supplied
// JSON path over the embedded copy. Missing user path falls through to the
// embedded patch so the brick binary works out of the box.
func resolvePatchEntries(basePatch string) ([]patch.Entry, error) {
	if basePatch != "" {
		if _, err := os.Stat(basePatch); err == nil {
			entries, err := patch.LoadFile(basePatch)
			if err != nil {
				return nil, fmt.Errorf("load base patch: %w", err)
			}
			return entries, nil
		}
	}
	entries, err := patch.LoadEmbedded()
	if err != nil {
		return nil, fmt.Errorf("load embedded base patch: %w", err)
	}
	return entries, nil
}

// ensurePatched leaves the ROM in a state where r.CheckMD5() == true. If the
// source file already matches, no patch is applied; otherwise the patch is
// resolved (file → embedded) and applied.
func ensurePatched(r *rom.ROM, basePatch string, skipMD5 bool) error {
	if skipMD5 {
		return nil
	}
	ok, err := r.CheckMD5()
	if err != nil {
		return err
	}
	if !ok {
		if err := r.Resize(rom.Size); err != nil {
			return err
		}
		entries, err := resolvePatchEntries(basePatch)
		if err != nil {
			return err
		}
		if err := patch.Apply(r, entries, false); err != nil {
			return err
		}
	}
	ok, err = r.CheckMD5()
	if err != nil {
		return err
	}
	if !ok {
		return fmt.Errorf("MD5 check failed after base patch")
	}
	return nil
}

// resolveHeartColor handles the "random" → concrete color mapping.
func resolveHeartColor(hc string) (string, error) {
	if hc != "random" {
		return hc, nil
	}
	opts := []string{"blue", "green", "yellow", "red"}
	idx, err := helpers.GetRandomInt(0, 3)
	if err != nil {
		return "", err
	}
	return opts[idx], nil
}

func buildROM(o Options, w *world.World, ir *item.Registry, stem string) (string, error) {
	quickswap, err := rom.ParseBool(o.Quickswap)
	if err != nil {
		return "", err
	}
	heartColor, err := resolveHeartColor(o.HeartColor)
	if err != nil {
		return "", err
	}

	r, err := rom.Open(o.BaseROMPath)
	if err != nil {
		return "", err
	}
	defer r.Close()

	if err := ensurePatched(r, o.BasePatch, o.SkipMD5); err != nil {
		return "", err
	}

	if err := r.SetHeartColors(heartColor); err != nil {
		return "", err
	}
	if err := r.SetHeartBeepSpeed(o.HeartBeep); err != nil {
		return "", err
	}
	if err := r.SetQuickSwap(quickswap); err != nil {
		return "", err
	}
	if err := r.SetMenuSpeed(o.MenuSpeed); err != nil {
		return "", err
	}
	if err := r.MuteMusic(o.NoMusic); err != nil {
		return "", err
	}

	if err := w.WriteToRom(r, ir); err != nil {
		return "", fmt.Errorf("write randomization: %w", err)
	}

	out := filepath.Join(o.OutputDir, stem+".sfc")
	if err := r.Save(out); err != nil {
		return "", err
	}
	return out, nil
}

// runUnrandomized mirrors the Phase-1 path: apply base patch + QoL only.
func runUnrandomized(o Options) (Result, error) {
	var res Result
	quickswap, err := rom.ParseBool(o.Quickswap)
	if err != nil {
		return res, err
	}
	heartColor, err := resolveHeartColor(o.HeartColor)
	if err != nil {
		return res, err
	}

	r, err := rom.Open(o.BaseROMPath)
	if err != nil {
		return res, err
	}
	defer r.Close()

	if err := ensurePatched(r, o.BasePatch, o.SkipMD5); err != nil {
		return res, err
	}

	if err := r.SetHeartColors(heartColor); err != nil {
		return res, err
	}
	if err := r.SetHeartBeepSpeed(o.HeartBeep); err != nil {
		return res, err
	}
	if err := r.SetQuickSwap(quickswap); err != nil {
		return res, err
	}

	out := filepath.Join(o.OutputDir, fmt.Sprintf("alttp-%s.sfc", rom.Build))
	if err := r.Save(out); err != nil {
		return res, err
	}
	res.ROMPath = out
	return res, nil
}
