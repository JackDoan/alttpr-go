package main

import (
	"encoding/json"
	"errors"
	"fmt"
	"io/fs"
	"os"
	"path/filepath"

	"github.com/JackDoan/alttpr-go/internal/job"
)

const (
	defaultBaseROM   = "/mnt/SDCARD/Apps/alttpr/base.sfc"
	defaultOutputDir = "/mnt/SDCARD/Roms/SFC"
	configFile       = "settings.json"
)

// Config is the on-disk state for the harness: where to find the base ROM,
// where to drop output ROMs, and the user's last menu choices.
type Config struct {
	BaseROM     string      `json:"base_rom"`
	OutputDir   string      `json:"output_dir"`
	LastOptions job.Options `json:"last_options"`
}

func defaultConfig() Config {
	c := Config{
		BaseROM:     defaultBaseROM,
		OutputDir:   defaultOutputDir,
		LastOptions: job.DefaultOptions(),
	}
	return c
}

// configPath returns where settings.json lives. We look (in order) at:
//  1. the directory containing the running binary (typical install)
//  2. the default Trimui app path
//
// The first writable directory wins.
func configPath() string {
	if exe, err := os.Executable(); err == nil {
		dir := filepath.Dir(exe)
		if isWritableDir(dir) {
			return filepath.Join(dir, configFile)
		}
	}
	return filepath.Join("/mnt/SDCARD/Apps/alttpr", configFile)
}

func isWritableDir(dir string) bool {
	info, err := os.Stat(dir)
	if err != nil || !info.IsDir() {
		return false
	}
	// Probe by creating a temp file. Avoids checking permission bits which
	// don't always reflect actual writability on FAT/exFAT SD cards.
	f, err := os.CreateTemp(dir, ".alttpr-probe-*")
	if err != nil {
		return false
	}
	name := f.Name()
	f.Close()
	os.Remove(name)
	return true
}

// LoadConfig reads settings.json, falling back to defaults if missing.
// A missing file is not an error; a malformed one is.
func LoadConfig(path string) (Config, error) {
	c := defaultConfig()
	data, err := os.ReadFile(path)
	if err != nil {
		if errors.Is(err, fs.ErrNotExist) {
			return c, nil
		}
		return c, fmt.Errorf("read %s: %w", path, err)
	}
	if err := json.Unmarshal(data, &c); err != nil {
		return c, fmt.Errorf("parse %s: %w", path, err)
	}
	// Fill any zero-value fields from defaults so a partial config still works.
	d := defaultConfig()
	if c.BaseROM == "" {
		c.BaseROM = d.BaseROM
	}
	if c.OutputDir == "" {
		c.OutputDir = d.OutputDir
	}
	c.LastOptions = mergeOptions(d.LastOptions, c.LastOptions)
	return c, nil
}

// SaveConfig writes the config back to disk, atomically when possible.
func SaveConfig(path string, c Config) error {
	data, err := json.MarshalIndent(c, "", "  ")
	if err != nil {
		return err
	}
	tmp := path + ".tmp"
	if err := os.WriteFile(tmp, data, 0o644); err != nil {
		return err
	}
	return os.Rename(tmp, path)
}

// mergeOptions returns user-over-default with empties treated as "use default".
func mergeOptions(d, u job.Options) job.Options {
	out := d
	if u.Goal != "" {
		out.Goal = u.Goal
	}
	if u.State != "" {
		out.State = u.State
	}
	if u.Weapons != "" {
		out.Weapons = u.Weapons
	}
	if u.Glitches != "" {
		out.Glitches = u.Glitches
	}
	if u.ItemPlacement != "" {
		out.ItemPlacement = u.ItemPlacement
	}
	if u.DungeonItems != "" {
		out.DungeonItems = u.DungeonItems
	}
	if u.Accessibility != "" {
		out.Accessibility = u.Accessibility
	}
	if u.ItemPool != "" {
		out.ItemPool = u.ItemPool
	}
	if u.ItemFunc != "" {
		out.ItemFunc = u.ItemFunc
	}
	if u.Hints != "" {
		out.Hints = u.Hints
	}
	if u.CrystalsGanon != 0 {
		out.CrystalsGanon = u.CrystalsGanon
	}
	if u.CrystalsTower != 0 {
		out.CrystalsTower = u.CrystalsTower
	}
	if u.TriforcePieces != 0 {
		out.TriforcePieces = u.TriforcePieces
	}
	if u.TriforceGoal != 0 {
		out.TriforceGoal = u.TriforceGoal
	}
	if u.HeartColor != "" {
		out.HeartColor = u.HeartColor
	}
	if u.HeartBeep != "" {
		out.HeartBeep = u.HeartBeep
	}
	if u.MenuSpeed != "" {
		out.MenuSpeed = u.MenuSpeed
	}
	if u.Quickswap != "" {
		out.Quickswap = u.Quickswap
	}
	out.NoMusic = u.NoMusic
	out.Tournament = u.Tournament
	return out
}
