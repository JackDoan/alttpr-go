package rom

import (
	_ "embed"
	"fmt"
)

//go:embed ww_enabled.bin
var wwEnabledBlob []byte

//go:embed ww_disabled.bin
var wwDisabledBlob []byte

// le16 packs a 16-bit value little-endian.
func le16(v int) []byte { return []byte{byte(v & 0xFF), byte((v >> 8) & 0xFF)} }

// SetGoalRequiredCount writes the goal-item required count.
// PHP: app/Rom.php setGoalRequiredCount.
func (r *ROM) SetGoalRequiredCount(goal int) error {
	return r.Write(0x180167, le16(goal), true)
}

// SetGoalIcon sets the goal-item HUD icon. PHP: setGoalIcon.
func (r *ROM) SetGoalIcon(icon string) error {
	val := 0x280D
	if icon == "triforce" {
		val = 0x280E
	}
	return r.Write(0x180165, le16(val), true)
}

// SetLimitProgressiveSword. PHP: setLimitProgressiveSword.
func (r *ROM) SetLimitProgressiveSword(limit, replacement int) error {
	return r.Write(0x180090, []byte{byte(limit), byte(replacement)}, true)
}
func (r *ROM) SetLimitProgressiveShield(limit, replacement int) error {
	return r.Write(0x180092, []byte{byte(limit), byte(replacement)}, true)
}
func (r *ROM) SetLimitProgressiveArmor(limit, replacement int) error {
	return r.Write(0x180094, []byte{byte(limit), byte(replacement)}, true)
}
func (r *ROM) SetLimitBottle(limit, replacement int) error {
	return r.Write(0x180096, []byte{byte(limit), byte(replacement)}, true)
}
func (r *ROM) SetLimitProgressiveBow(limit, replacement int) error {
	return r.Write(0x180098, []byte{byte(limit), byte(replacement)}, true)
}

// SetGanonInvincible. PHP: setGanonInvincible.
func (r *ROM) SetGanonInvincible(setting string) error {
	var b byte
	switch setting {
	case "crystals":
		b = 0x03
	case "dungeons":
		b = 0x02
	case "yes":
		b = 0x01
	case "crystals_only":
		b = 0x04
	case "triforce_pieces":
		b = 0x05
	case "lightspeed":
		b = 0x06
	case "crystals_bosses":
		b = 0x07
	case "bosses_only":
		b = 0x08
	case "dungeons_no_agahnim":
		b = 0x09
	case "completionist":
		b = 0x0B
	case "no", "":
		b = 0x00
	default:
		b = 0x00
	}
	return r.Write(0x1801A8, []byte{b}, true)
}

// SetTowerCrystalRequirement. PHP: setTowerCrystalRequirement.
func (r *ROM) SetTowerCrystalRequirement(crystals int) error {
	c := max(0, min(crystals, 7))
	return r.Write(0x18019A, []byte{byte(c)}, true)
}

// SetGanonCrystalRequirement. PHP: setGanonCrystalRequirement.
func (r *ROM) SetGanonCrystalRequirement(crystals int) error {
	c := max(0, min(crystals, 7))
	return r.Write(0x1801A6, []byte{byte(c)}, true)
}

// SetMapMode. PHP: setMapMode.
func (r *ROM) SetMapMode(requireMap bool) error {
	b := byte(0x00)
	if requireMap {
		b = 0x01
	}
	return r.Write(0x18003B, []byte{b}, true)
}

// SetCompassMode. PHP: setCompassMode (on/pickup/off).
func (r *ROM) SetCompassMode(setting string) error {
	var b byte
	switch setting {
	case "on":
		b = 0x02
	case "pickup":
		b = 0x01
	default:
		b = 0x00
	}
	return r.Write(0x18003C, []byte{b}, true)
}

// SetSwordlessMode. PHP: setSwordlessMode (also calls HammerTablet/HammerBarrier/SRAM).
// We write just the two main bytes; the HammerTablet/Barrier toggles and
// SRAM swordless curtains are handled separately as the InitialSram port lands.
func (r *ROM) SetSwordlessMode(enable bool) error {
	b := byte(0x00)
	if enable {
		b = 0x01
	}
	if err := r.Write(0x18003F, []byte{b}, true); err != nil { // Hammer Ganon
		return err
	}
	return r.Write(0x180041, []byte{b}, true) // Swordless Medallions
}

// SetSewersLampCone. PHP: setSewersLampCone.
func (r *ROM) SetSewersLampCone(enable bool) error {
	b := byte(0x00)
	if enable {
		b = 0x01
	}
	return r.Write(0x180038, []byte{b}, true)
}

// SetSubstitutions writes the substitution table at 0x184000.
// PHP: setSubstitutions — pads with [0xFF, 0xFF, 0xFF, 0xFF] sentinel.
func (r *ROM) SetSubstitutions(subs []byte) error {
	out := append([]byte(nil), subs...)
	out = append(out, 0xFF, 0xFF, 0xFF, 0xFF)
	return r.Write(0x184000, out, true)
}

// SetRupeeArrow. PHP: setRupeeArrow — fish bottle merchant.
func (r *ROM) SetRupeeArrow(enable bool) error {
	b := byte(0xE2)
	if enable {
		b = 0xDB
	}
	return r.Write(0x30052, []byte{b}, true)
}

// SetGenericKeys. PHP: setGenericKeys.
func (r *ROM) SetGenericKeys(enable bool) error {
	b := byte(0x00)
	if enable {
		b = 0x01
	}
	return r.Write(0x180172, []byte{b}, true)
}

// SetTotalItemCount. PHP: setTotalItemCount — 16-bit count of collectables.
func (r *ROM) SetTotalItemCount(count int) error {
	return r.Write(0x180196, le16(count), true)
}

// SetGameType. PHP: setGameType — bit-encoded mode byte.
func (r *ROM) SetGameType(gameType string) error {
	var b byte
	switch gameType {
	case "enemizer":
		b = 0b00000101
	case "entrance":
		b = 0b00000110
	case "room":
		b = 0b00001000
	default: // "item"
		b = 0b00000100
	}
	return r.Write(0x180211, []byte{b}, true)
}

// SetPyramidFairyChests. PHP: setPyramidFairyChests — writes a 6-byte block at 0x1FC16.
func (r *ROM) SetPyramidFairyChests(enable bool) error {
	if enable {
		return r.Write(0x1FC16, []byte{0xB1, 0xC6, 0xF9, 0xC9, 0xC6, 0xF9}, true)
	}
	return r.Write(0x1FC16, []byte{0xA8, 0xB8, 0x3D, 0xD0, 0xB8, 0x3D}, true)
}

// SetSmithyQuickItemGive. PHP: setSmithyQuickItemGive.
func (r *ROM) SetSmithyQuickItemGive(enable bool) error {
	b := byte(0x00)
	if enable {
		b = 0x01
	}
	return r.Write(0x180029, []byte{b}, true)
}

// SetPseudoBoots. PHP: setPseudoBoots.
func (r *ROM) SetPseudoBoots(enable bool) error {
	b := byte(0x00)
	if enable {
		b = 0x01
	}
	return r.Write(0x18008E, []byte{b}, true)
}

// SetFastROM. PHP: enableFastRom (already in qol; aliased).
func (r *ROM) EnableFastROM(enable bool) error {
	b := byte(0x00)
	if enable {
		b = 0x01
	}
	return r.Write(0x187032, []byte{b}, true)
}

// SetLockAgahnimDoorInEscape. PHP: setLockAgahnimDoorInEscape.
func (r *ROM) SetLockAgahnimDoorInEscape(enable bool) error {
	b := byte(0x00)
	if enable {
		b = 0x01
	}
	return r.Write(0x180169, []byte{b}, true)
}

// RemoveUnclesSword. PHP: removeUnclesSword — rewrites 10 6-byte tile records
// to remove the sword from Uncle's gear-up animation.
func (r *ROM) RemoveUnclesSword() error {
	records := []struct {
		addr int
		data []byte
	}{
		{0x6D263, []byte{0x00, 0x00, 0xF6, 0xFF, 0x00, 0x0E}},
		{0x6D26B, []byte{0x00, 0x00, 0xF6, 0xFF, 0x00, 0x0E}},
		{0x6D293, []byte{0x00, 0x00, 0xF6, 0xFF, 0x00, 0x0E}},
		{0x6D29B, []byte{0x00, 0x00, 0xF7, 0xFF, 0x00, 0x0E}},
		{0x6D2B3, []byte{0x00, 0x00, 0xF6, 0xFF, 0x02, 0x0E}},
		{0x6D2BB, []byte{0x00, 0x00, 0xF6, 0xFF, 0x02, 0x0E}},
		{0x6D2E3, []byte{0x00, 0x00, 0xF7, 0xFF, 0x02, 0x0E}},
		{0x6D2EB, []byte{0x00, 0x00, 0xF7, 0xFF, 0x02, 0x0E}},
		{0x6D31B, []byte{0x00, 0x00, 0xE4, 0xFF, 0x08, 0x0E}},
		{0x6D323, []byte{0x00, 0x00, 0xE4, 0xFF, 0x08, 0x0E}},
	}
	for _, rec := range records {
		if err := r.Write(rec.addr, rec.data, true); err != nil {
			return err
		}
	}
	return nil
}

// RemoveUnclesShield. PHP: removeUnclesShield — 7 records.
func (r *ROM) RemoveUnclesShield() error {
	records := []struct {
		addr int
		data []byte
	}{
		{0x6D253, []byte{0x00, 0x00, 0xF6, 0xFF, 0x00, 0x0E}},
		{0x6D25B, []byte{0x00, 0x00, 0xF6, 0xFF, 0x00, 0x0E}},
		{0x6D283, []byte{0x00, 0x00, 0xF6, 0xFF, 0x00, 0x0E}},
		{0x6D28B, []byte{0x00, 0x00, 0xF7, 0xFF, 0x00, 0x0E}},
		{0x6D2CB, []byte{0x00, 0x00, 0xF6, 0xFF, 0x02, 0x0E}},
		{0x6D2FB, []byte{0x00, 0x00, 0xF7, 0xFF, 0x02, 0x0E}},
		{0x6D313, []byte{0x00, 0x00, 0xE4, 0xFF, 0x08, 0x0E}},
	}
	for _, rec := range records {
		if err := r.Write(rec.addr, rec.data, true); err != nil {
			return err
		}
	}
	return nil
}


// SetOpenMode is the partial port — SRAM-modifying part is deferred.
// PHP: setOpenMode.
func (r *ROM) SetOpenMode(enable bool) error {
	if err := r.SetSewersLampCone(!enable); err != nil {
		return err
	}
	return nil
}

// SetStandardMode. PHP: setStandardMode (SRAM part deferred).
func (r *ROM) SetStandardMode() error {
	return r.SetSewersLampCone(true)
}

// SetGameState dispatches to the per-state setter. PHP: setGameState.
func (r *ROM) SetGameState(state string) error {
	switch state {
	case "open", "retro":
		return r.SetOpenMode(true)
	case "standard":
		return r.SetStandardMode()
	default:
		return fmt.Errorf("unsupported game state %q", state)
	}
}

// SetWishingWellChests. PHP: setWishingWellChests — writes item-table values
// at 0xE9AE/0xE9CF and a 205-byte base64-decoded blob at 0x1F714.
func (r *ROM) SetWishingWellChests(enable bool) error {
	if enable {
		if err := r.Write(0xE9AE, []byte{0x14, 0x01}, true); err != nil {
			return err
		}
		if err := r.Write(0xE9CF, []byte{0x14, 0x01}, true); err != nil {
			return err
		}
		return r.Write(0x1F714, wwEnabledBlob, true)
	}
	if err := r.Write(0xE9AE, []byte{0x05, 0x00}, true); err != nil {
		return err
	}
	if err := r.Write(0xE9CF, []byte{0x3D, 0x01}, true); err != nil {
		return err
	}
	return r.Write(0x1F714, wwDisabledBlob, true)
}

// SetWishingWellUpgrade. PHP: setWishingWellUpgrade.
func (r *ROM) SetWishingWellUpgrade(enable bool) error {
	a := byte(0x2A)
	bb := byte(0x05)
	if enable {
		a = 0x0C
		bb = 0x04
	}
	if err := r.Write(0x348DB, []byte{a}, true); err != nil {
		return err
	}
	return r.Write(0x348EB, []byte{bb}, true)
}

// SetHyliaFairyShop. PHP: setHyliaFairyShop — writes 6 bytes at 0x01F810.
func (r *ROM) SetHyliaFairyShop(enable bool) error {
	if enable {
		return r.Write(0x01F810, []byte{0x1A, 0x1E, 0x01, 0x1A, 0x1E, 0x01}, true)
	}
	return r.Write(0x01F810, []byte{0xFC, 0x94, 0xE4, 0xFD, 0x34, 0xE4}, true)
}

// SetRestrictFairyPonds. PHP: setRestrictFairyPonds.
func (r *ROM) SetRestrictFairyPonds(enable bool) error {
	b := byte(0x00)
	if enable {
		b = 0x01
	}
	return r.Write(0x18017E, []byte{b}, true)
}

// SetSilversOnlyAtGanon. PHP: setSilversOnlyAtGanon.
func (r *ROM) SetSilversOnlyAtGanon(enable bool) error {
	b := byte(0x00)
	if enable {
		b = 0x01
	}
	return r.Write(0x180181, []byte{b}, true)
}

// SetCatchableFairies. PHP: setCatchableFairies.
func (r *ROM) SetCatchableFairies(enable bool) error {
	b := byte(0x80)
	if enable {
		b = 0xF0
	}
	return r.Write(0x34FD6, []byte{b}, true)
}

// SetBottleFills. PHP: setBottleFills — health then magic.
func (r *ROM) SetBottleFills(health, magicBar int) error {
	if err := r.Write(0x180084, []byte{byte(health)}, true); err != nil {
		return err
	}
	return r.Write(0x180085, []byte{byte(magicBar)}, true)
}

// SetClockMode. PHP: setClockMode — writes 3 bytes at 0x180190 (mode hi, mode lo, restart).
// Caller is responsible for separately clearing compass mode when enabled.
func (r *ROM) SetClockMode(setting string) error {
	var bytes []byte
	restart := byte(0x00)
	switch setting {
	case "stopwatch":
		bytes = []byte{0x02, 0x01}
	case "countdown-ohko":
		bytes = []byte{0x01, 0x02}
		restart = 0x01
	case "countdown-continue":
		bytes = []byte{0x01, 0x01}
	case "countdown-stop":
		bytes = []byte{0x01, 0x00}
	case "countdown-end":
		bytes = []byte{0x01, 0x03}
	default: // "off"
		bytes = []byte{0x00, 0x00}
	}
	return r.Write(0x180190, append(bytes, restart), true)
}
