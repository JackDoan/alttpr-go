package main

import (
	"os"
	"path/filepath"
	"testing"
)

func TestLoadConfig_MissingFileReturnsDefaults(t *testing.T) {
	dir := t.TempDir()
	c, err := LoadConfig(filepath.Join(dir, "nope.json"))
	if err != nil {
		t.Fatalf("LoadConfig(missing): %v", err)
	}
	d := defaultConfig()
	if c.BaseROM != d.BaseROM || c.OutputDir != d.OutputDir {
		t.Errorf("defaults not applied: %+v", c)
	}
	if c.LastOptions.Goal != "ganon" {
		t.Errorf("LastOptions defaults missing: %+v", c.LastOptions)
	}
}

func TestSaveLoadRoundTrip(t *testing.T) {
	dir := t.TempDir()
	path := filepath.Join(dir, "settings.json")

	c := defaultConfig()
	c.BaseROM = "/somewhere/base.sfc"
	c.OutputDir = "/elsewhere/Roms"
	c.LastOptions.Goal = "pedestal"
	c.LastOptions.CrystalsGanon = 5

	if err := SaveConfig(path, c); err != nil {
		t.Fatalf("Save: %v", err)
	}
	got, err := LoadConfig(path)
	if err != nil {
		t.Fatalf("Load: %v", err)
	}
	if got.BaseROM != c.BaseROM || got.OutputDir != c.OutputDir {
		t.Errorf("paths not preserved: %+v", got)
	}
	if got.LastOptions.Goal != "pedestal" || got.LastOptions.CrystalsGanon != 5 {
		t.Errorf("LastOptions not preserved: %+v", got.LastOptions)
	}
}

func TestLoadConfig_PartialOverlayedOnDefaults(t *testing.T) {
	dir := t.TempDir()
	path := filepath.Join(dir, "settings.json")
	if err := os.WriteFile(path, []byte(`{"base_rom":"/x.sfc"}`), 0o644); err != nil {
		t.Fatal(err)
	}
	c, err := LoadConfig(path)
	if err != nil {
		t.Fatal(err)
	}
	if c.BaseROM != "/x.sfc" {
		t.Errorf("base_rom not honored: %q", c.BaseROM)
	}
	if c.OutputDir != defaultOutputDir {
		t.Errorf("output_dir default not applied: %q", c.OutputDir)
	}
	if c.LastOptions.Goal != "ganon" {
		t.Errorf("LastOptions defaults not applied: %+v", c.LastOptions)
	}
}
