package world

import (
	"github.com/JackDoan/alttpr-go/internal/item"
	"github.com/JackDoan/alttpr-go/internal/rom"
)

// writeEscapeFills writes escape-fills flags and per-spawn refill counts
// based on what Uncle has placed at his location. Mirrors PHP World::setEscapeFills.
func (w *World) writeEscapeFills(r *rom.ROM) error {
	uncle := w.locations.Get("Link's Uncle:" + itoaW(w.id))
	uncleItems := item.NewCollection()
	uncleItems.SetChecksForWorld(w.id)
	if uncle != nil && uncle.HasItem() {
		uncleItems.Add(uncle.Item())
	}

	// Temporarily disable ignoreCanKillEscapeThings for the check.
	saved := w.ConfigBool("ignoreCanKillEscapeThings", false)
	w.config.Bools["ignoreCanKillEscapeThings"] = false
	if !uncleItems.CanKillEscapeThings(w) {
		uncleItems = uncleItems.Merge(w.preCollected)
	}
	w.config.Bools["ignoreCanKillEscapeThings"] = saved

	defaultHealth := w.ConfigString("enemizer.enemyHealth", "default") == "default"

	switch {
	case uncleItems.HasSword(1) || uncleItems.Has1("Hammer"):
		_ = r.SetEscapeFills(0b00000000, w.ConfigInt("rom.EscapeRefills.StartingRupees", 300))
		_ = r.SetUncleSpawnRefills(0, 0, 0)
		_ = r.SetZeldaSpawnRefills(0, 0, 0)
		_ = r.SetMantleSpawnRefills(0, 0, 0)
	case uncleItems.Has1("FireRod") ||
		uncleItems.Has1("CaneOfSomaria") ||
		(uncleItems.Has1("CaneOfByrna") && defaultHealth):
		_ = r.SetEscapeFills(0b00000100, w.ConfigInt("rom.EscapeRefills.StartingRupees", 300))
		_ = r.SetUncleSpawnRefills(w.ConfigInt("rom.EscapeRefills.Uncle.Magic", 0x80), 0, 0)
		_ = r.SetZeldaSpawnRefills(w.ConfigInt("rom.EscapeRefills.Zelda.Magic", 0x20), 0, 0)
		_ = r.SetMantleSpawnRefills(w.ConfigInt("rom.EscapeRefills.Mantle.Magic", 0x20), 0, 0)
		if w.ConfigBool("rom.EscapeAssist", false) {
			_ = r.SetEscapeAssist(0b00000100)
		}
	case uncleItems.CanShootArrows(w, 1):
		_ = r.SetEscapeFills(0b00000001, w.ConfigInt("rom.EscapeRefills.StartingRupees", 300))
		_ = r.SetUncleSpawnRefills(0, 0, w.ConfigInt("rom.EscapeRefills.Uncle.Arrows", 70))
		_ = r.SetZeldaSpawnRefills(0, 0, w.ConfigInt("rom.EscapeRefills.Zelda.Arrows", 10))
		_ = r.SetMantleSpawnRefills(0, 0, w.ConfigInt("rom.EscapeRefills.Mantle.Arrows", 10))
		if w.ConfigBool("rom.EscapeAssist", false) {
			_ = r.SetEscapeAssist(0b00000001)
		}
	case uncleItems.Has1("TenBombs") || w.ConfigString("logic", "") != "NoLogic":
		_ = r.SetEscapeFills(0b00000010, w.ConfigInt("rom.EscapeRefills.StartingRupees", 300))
		_ = r.SetUncleSpawnRefills(0, w.ConfigInt("rom.EscapeRefills.Uncle.Bombs", 50), 0)
		_ = r.SetZeldaSpawnRefills(0, w.ConfigInt("rom.EscapeRefills.Zelda.Bombs", 3), 0)
		_ = r.SetMantleSpawnRefills(0, w.ConfigInt("rom.EscapeRefills.Mantle.Bombs", 3), 0)
		if w.ConfigBool("rom.EscapeAssist", false) {
			_ = r.SetEscapeAssist(0b00000010)
		}
	}
	return nil
}
