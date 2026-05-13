package rom

import (
	"crypto/rand"
	"fmt"
	"math/big"
)

// SetRupoorValue. PHP: setRupoorValue — 16-bit LE at 0x180036.
func (r *ROM) SetRupoorValue(value int) error {
	return r.Write(0x180036, le16(value), true)
}

// SetByrnaCaveSpikeDamage. PHP: setByrnaCaveSpikeDamage.
func (r *ROM) SetByrnaCaveSpikeDamage(dmg int) error {
	return r.Write(0x180195, []byte{byte(dmg)}, true)
}

// SetCaneOfByrnaSpikeCaveUsage. PHP: setCaneOfByrnaSpikeCaveUsage.
func (r *ROM) SetCaneOfByrnaSpikeCaveUsage(normal, half, quarter int) error {
	return r.Write(0x18016B, []byte{byte(normal), byte(half), byte(quarter)}, true)
}

// SetCapeSpikeCaveUsage. PHP: setCapeSpikeCaveUsage.
func (r *ROM) SetCapeSpikeCaveUsage(normal, half, quarter int) error {
	return r.Write(0x18016E, []byte{byte(normal), byte(half), byte(quarter)}, true)
}

// SetCaneOfByrnaInvulnerability. PHP: setCaneOfByrnaInvulnerability.
func (r *ROM) SetCaneOfByrnaInvulnerability(enable bool) error {
	b := byte(0x00)
	if enable {
		b = 0x01
	}
	return r.Write(0x18004F, []byte{b}, true)
}

// SetCapeRegularMagicUsage. PHP: setCapeRegularMagicUsage.
func (r *ROM) SetCapeRegularMagicUsage(normal, half, quarter int) error {
	return r.Write(0x180174, []byte{byte(normal), byte(half), byte(quarter)}, true)
}

// SetPowderedSpriteFairyPrize. PHP: setPowderedSpriteFairyPrize.
func (r *ROM) SetPowderedSpriteFairyPrize(b int) error {
	return r.Write(0x36DD0, []byte{byte(b)}, true)
}

// SetCatchableBees. PHP: setCatchableBees.
func (r *ROM) SetCatchableBees(enable bool) error {
	b := byte(0xF0)
	if !enable {
		b = 0x80
	}
	return r.Write(0xF5D73, []byte{b}, true)
}

// SetStunItems. PHP: setStunItems — flag byte at 0x180180.
func (r *ROM) SetStunItems(flags int) error {
	return r.Write(0x180180, []byte{byte(flags)}, true)
}

// SetBlueClock / SetRedClock / SetGreenClock. PHP: setBlueClock/RedClock/GreenClock —
// 32-bit LE seconds*60 at fixed addresses.
func (r *ROM) SetBlueClock(seconds int) error {
	return r.Write(0x180200, le32(seconds*60), true)
}
func (r *ROM) SetRedClock(seconds int) error {
	return r.Write(0x180204, le32(seconds*60), true)
}
func (r *ROM) SetGreenClock(seconds int) error {
	return r.Write(0x180208, le32(seconds*60), true)
}

// SetEscapeFills. PHP: setEscapeFills.
func (r *ROM) SetEscapeFills(flags, rupees int) error {
	if err := r.Write(0x18004E, []byte{byte(flags)}, true); err != nil {
		return err
	}
	return r.Write(0x180183, le16(rupees), true)
}

// SetCapacityUpgradeFills. PHP: setCapacityUpgradeFills — 4 bytes at 0x180080.
func (r *ROM) SetCapacityUpgradeFills(fills [4]int) error {
	return r.Write(0x180080, []byte{byte(fills[0]), byte(fills[1]), byte(fills[2]), byte(fills[3])}, true)
}

// SetBallNChainDungeon. PHP: setBallNChainDungeon.
func (r *ROM) SetBallNChainDungeon(id int) error {
	return r.Write(0x18020A, []byte{byte(id)}, true)
}

// SetCompassCountTotals. PHP: setCompassCountTotals — 16-byte block at 0x187000.
func (r *ROM) SetCompassCountTotals(totals []int) error {
	def := []int{0x08, 0x08, 0x06, 0x06, 0x02, 0x0A, 0x0E, 0x08, 0x08, 0x08, 0x06, 0x08, 0x0C, 0x1B, 0x00, 0x00}
	if len(totals) == 0 {
		totals = def
	}
	buf := make([]byte, len(totals))
	for i, v := range totals {
		buf[i] = byte(v)
	}
	return r.Write(0x187000, buf, true)
}

// SetMapRevealSahasrahla. PHP: setMapRevealSahasrahla — 16-bit LE at 0x18017C.
func (r *ROM) SetMapRevealSahasrahla(reveals int) error {
	return r.Write(0x18017C, le16(reveals), true)
}

// SetMapRevealBombShop. PHP: setMapRevealBombShop — 16-bit LE at 0x18017A.
func (r *ROM) SetMapRevealBombShop(reveals int) error {
	return r.Write(0x18017A, le16(reveals), true)
}

// SetMysteryMasking is a no-op at the ROM-byte level — PHP `setMysteryMasking`
// modifies the `intro_main` text string, which is handled via Text.SetString
// at the World level. Kept as a stub so WriteToRom can call it unconditionally.
func (r *ROM) SetMysteryMasking(enable bool) error { return nil }

// SetFreeItemTextMode. PHP: setFreeItemTextMode.
func (r *ROM) SetFreeItemTextMode(flags int) error {
	return r.Write(0x180184, []byte{byte(flags)}, true)
}

// SetFreeItemMenu. PHP: setFreeItemMenu.
func (r *ROM) SetFreeItemMenu(flags int) error {
	return r.Write(0x180185, []byte{byte(flags)}, true)
}

// SetGanonAgahnimRng. PHP: setGanonAgahnimRng — 'table' = 0x00, 'random' = 0x01.
func (r *ROM) SetGanonAgahnimRng(setting string) error {
	b := byte(0x00)
	if setting == "random" {
		b = 0x01
	}
	return r.Write(0x180086, []byte{b}, true)
}

// SetDiggingGameRng. PHP: setDiggingGameRng.
func (r *ROM) SetDiggingGameRng(digs int) error {
	if err := r.Write(0x180A6C, []byte{byte(digs)}, true); err != nil {
		return err
	}
	return r.Write(0xEFD95, []byte{byte(digs)}, true)
}

// SetSilversEquip. PHP: setSilversEquip — 'collection' = 0x00, 'on' = 0x01.
func (r *ROM) SetSilversEquip(setting string) error {
	b := byte(0x00)
	if setting == "on" {
		b = 0x01
	}
	return r.Write(0x180182, []byte{b}, true)
}

// EnableTriforceTurnIn. PHP: enableTriforceTurnIn.
func (r *ROM) EnableTriforceTurnIn(enable bool) error {
	b := byte(0x00)
	if enable {
		b = 0x01
	}
	return r.Write(0x180194, []byte{b}, true)
}

// EnableHudItemCounter. PHP: enableHudItemCounter.
func (r *ROM) EnableHudItemCounter(enable bool) error {
	b := byte(0x00)
	if enable {
		b = 0x01
	}
	return r.Write(0x180215, []byte{b}, true)
}

// SetSwampWaterLevel. PHP: setSwampWaterLevel.
func (r *ROM) SetSwampWaterLevel(enable bool) error {
	b := byte(0x00)
	if enable {
		b = 0x01
	}
	return r.Write(0x180066, []byte{b}, true)
}

// SetPreAgahnimDarkWorldDeathInDungeon. PHP: setPreAgahnimDarkWorldDeathInDungeon.
func (r *ROM) SetPreAgahnimDarkWorldDeathInDungeon(enable bool) error {
	b := byte(0x00)
	if enable {
		b = 0x01
	}
	return r.Write(0x180169, []byte{b}, true)
}

// SetSaveAndQuitFromBossRoom. PHP: setSaveAndQuitFromBossRoom.
func (r *ROM) SetSaveAndQuitFromBossRoom(enable bool) error {
	b := byte(0x00)
	if enable {
		b = 0x01
	}
	return r.Write(0x180042, []byte{b}, true)
}

// SetWorldOnAgahnimDeath. PHP: setWorldOnAgahnimDeath.
func (r *ROM) SetWorldOnAgahnimDeath(enable bool) error {
	b := byte(0x00)
	if enable {
		b = 0x01
	}
	return r.Write(0x1800A1, []byte{b}, true)
}

// SetAllowAccidentalMajorGlitch. PHP: setAllowAccidentalMajorGlitch.
func (r *ROM) SetAllowAccidentalMajorGlitch(enable bool) error {
	b := byte(0x00)
	if enable {
		b = 0x01
	}
	return r.Write(0x180209, []byte{b}, true)
}

// SetSQEGFix. PHP: setSQEGFix.
func (r *ROM) SetSQEGFix(enable bool) error {
	b := byte(0x00)
	if enable {
		b = 0x01
	}
	return r.Write(0x180169, []byte{b}, true)
}

// SetZeldaMirrorFix. PHP: setZeldaMirrorFix.
func (r *ROM) SetZeldaMirrorFix(enable bool) error {
	b := byte(0x00)
	if enable {
		b = 0x01
	}
	return r.Write(0x180248, []byte{b}, true)
}

// SetRandomizerSeedType. PHP: setRandomizerSeedType.
func (r *ROM) SetRandomizerSeedType(setting string) error {
	var b byte
	switch setting {
	case "OverworldGlitches":
		b = 0x02
	case "MajorGlitches":
		b = 0x01
	default: // NoGlitches
		b = 0x00
	}
	return r.Write(0x180210, []byte{b}, true)
}

// SetWarningFlags. PHP: setWarningFlags.
func (r *ROM) SetWarningFlags(flags int) error {
	return r.Write(0x180212, []byte{byte(flags)}, true)
}

// SetGameType replaces the older 4-bit version. Already in setters.go.
// (No duplicate here.)

// RupeeArrowFull writes the full set of PHP setRupeeArrow addresses.
// Replaces the simpler SetRupeeArrow which only writes the fish-merchant byte.
func (r *ROM) RupeeArrowFull(enable bool) error {
	fish := byte(0xE2)
	pot := byte(0xE1)
	chest := []byte{0x43, 0x44}
	thief := []byte{0xAF, 0x77, 0xF3, 0x7E}
	pikit := []byte{0xAF, 0x77, 0xF3, 0x7E}
	enableByte := byte(0x00)
	woodCost := []byte{0x00, 0x00}
	silverCost := []byte{0x00, 0x00}
	if enable {
		fish = 0xDB
		pot = 0xDA
		chest = []byte{0x35, 0x41}
		thief = []byte{0xA9, 0x00, 0xEA, 0xEA}
		pikit = []byte{0xA9, 0x00, 0xEA, 0xEA}
		enableByte = 0x01
		woodCost = []byte{0x0A, 0x00}
		silverCost = []byte{0x32, 0x00}
	}
	if err := r.Write(0x30052, []byte{fish}, true); err != nil {
		return err
	}
	if err := r.Write(0x301FC, []byte{pot}, true); err != nil {
		return err
	}
	if err := r.Write(0xECB4E, thief, true); err != nil {
		return err
	}
	if err := r.Write(0xF0D96, pikit, true); err != nil {
		return err
	}
	if err := r.Write(0x180175, []byte{enableByte}, true); err != nil {
		return err
	}
	if err := r.Write(0x180176, woodCost, true); err != nil {
		return err
	}
	if err := r.Write(0x180178, silverCost, true); err != nil {
		return err
	}
	return r.Write(0xEDA5, chest, true)
}

// WriteRNGBlock writes 1024 random bytes to 0x178000. PHP writeRNGBlock.
func (r *ROM) WriteRNGBlock() error {
	buf := make([]byte, 1024)
	for i := range buf {
		n, err := rand.Int(rand.Reader, big.NewInt(256))
		if err != nil {
			return fmt.Errorf("rng block: %w", err)
		}
		buf[i] = byte(n.Int64())
	}
	return r.Write(0x178000, buf, true)
}

// SetUncleSpawnRefills. PHP: setUncleSpawnRefills — magic, bombs, arrows at 0x180186.
func (r *ROM) SetUncleSpawnRefills(magic, bombs, arrows int) error {
	return r.Write(0x180186, []byte{byte(magic), byte(bombs), byte(arrows)}, true)
}

// SetZeldaSpawnRefills. PHP: setZeldaSpawnRefills — at 0x180189.
func (r *ROM) SetZeldaSpawnRefills(magic, bombs, arrows int) error {
	return r.Write(0x180189, []byte{byte(magic), byte(bombs), byte(arrows)}, true)
}

// SetMantleSpawnRefills. PHP: setMantleSpawnRefills — at 0x18018C.
func (r *ROM) SetMantleSpawnRefills(magic, bombs, arrows int) error {
	return r.Write(0x18018C, []byte{byte(magic), byte(bombs), byte(arrows)}, true)
}

// SetTournamentType. PHP: setTournamentType — 2 bytes at 0x180213.
func (r *ROM) SetTournamentType(setting string) error {
	bytes := []byte{0x00, 0x01} // "none" default
	if setting == "standard" {
		bytes = []byte{0x01, 0x00}
	}
	return r.Write(0x180213, bytes, true)
}

// SetStartScreenHash. PHP: setStartScreenHash — 5 bytes at 0x180215.
func (r *ROM) SetStartScreenHash(bs [5]int) error {
	out := []byte{byte(bs[0]), byte(bs[1]), byte(bs[2]), byte(bs[3]), byte(bs[4])}
	return r.Write(0x180215, out, true)
}

// SetSeedString. PHP: setSeedString — writes a 21-byte ASCII string at 0x7FC0.
func (r *ROM) SetSeedString(s string) error {
	buf := make([]byte, 21)
	for i := range buf {
		buf[i] = ' '
	}
	for i := 0; i < len(s) && i < 21; i++ {
		buf[i] = s[i]
	}
	return r.Write(0x7FC0, buf, true)
}

// SetEscapeAssist. PHP: setEscapeAssist.
func (r *ROM) SetEscapeAssist(flags int) error {
	return r.Write(0x18004D, []byte{byte(flags)}, true)
}

// SetPullTreePrizes. PHP: setPullTreePrizes — 3 bytes at 0xEFBD4.
func (r *ROM) SetPullTreePrizes(low, mid, high int) error {
	return r.Write(0xEFBD4, []byte{byte(low), byte(mid), byte(high)}, true)
}

// SetRupeeCrabPrizes. PHP: setRupeeCrabPrizes.
func (r *ROM) SetRupeeCrabPrizes(main, final int) error {
	if err := r.Write(0x329C8, []byte{byte(main)}, true); err != nil {
		return err
	}
	return r.Write(0x329C4, []byte{byte(final)}, true)
}

// SetStunnedSpritePrize. PHP: setStunnedSpritePrize.
func (r *ROM) SetStunnedSpritePrize(sprite int) error {
	return r.Write(0x37993, []byte{byte(sprite)}, true)
}

// SetFishSavePrize. PHP: setFishSavePrize.
func (r *ROM) SetFishSavePrize(prize int) error {
	return r.Write(0xE82CC, []byte{byte(prize)}, true)
}

// le32 packs a 32-bit value little-endian.
func le32(v int) []byte {
	u := uint32(v)
	return []byte{byte(u), byte(u >> 8), byte(u >> 16), byte(u >> 24)}
}
