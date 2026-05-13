package world

import (
	"github.com/JackDoan/alttpr-go/internal/helpers"
	"github.com/JackDoan/alttpr-go/internal/item"
	"github.com/JackDoan/alttpr-go/internal/rom"
)

// WriteToRom writes this world's randomization (item placements + game state)
// to the given ROM. Mirrors a focused subset of PHP World::writeToRom that
// covers item placement and the critical game-state setters needed for the
// ROM to boot into a playable randomized state.
//
// What's currently written:
//   - Every filled location's item bytes (via Location.WriteItem)
//   - Empty locations cleared to "Nothing"
//   - Crystal requirements (tower + ganon)
//   - Goal icon + required count
//   - Game state (standard/open)
//   - Swordless mode
//   - Limit/substitution tables for sword/shield/armor/bottle/bow/etc.
//   - Map/compass modes
//   - Uncle sword/shield removal if the placement isn't a sword
//   - Pyramid Fairy / Smithy quick-give based on swordsInPool
//   - Other PHP-default ROM flags (fast ROM, lampless sewers cone, etc.)
//
// What's deferred:
//   - In-game text customization (PHP Text class — defaults remain)
//   - Credits text customization (Credits class — defaults remain)
//   - InitialSram (pre-collected starting equipment, timers — defaults remain)
//   - Map-on-pickup region reveals
//   - Pseudo-boots, tournament mode, override patches
//   - Multiworld
func (w *World) WriteToRom(r *rom.ROM, ir *item.Registry) error {
	// 1. Write all location items. Skip kinds that can't accept "Nothing".
	for _, region := range w.regions {
		for _, l := range region.Locations.All() {
			if l.HasItem() {
				if err := l.WriteItem(r, ir, nil); err != nil {
					return err
				}
				continue
			}
			switch l.Kind {
			case KindMedallion, KindPrize, KindPrizeEvent, KindPrizeCrystal, KindPrizePendant, KindTrade:
				continue
			}
			nothing, err := ir.Get("Nothing", w.id)
			if err != nil {
				return err
			}
			if err := l.SetItem(nothing); err != nil {
				continue
			}
			if err := l.WriteItem(r, ir, nil); err != nil {
				return err
			}
		}
	}

	// 2. Goal + crystal requirements + per-item magic-usage settings.
	if err := r.SetGoalRequiredCount(w.ConfigInt("item.Goal.Required", 0)); err != nil {
		return err
	}
	if err := r.SetGoalIcon(w.ConfigString("item.Goal.Icon", "triforce")); err != nil {
		return err
	}
	if err := r.SetTowerCrystalRequirement(w.ConfigInt("crystals.tower", 7)); err != nil {
		return err
	}
	if err := r.SetGanonCrystalRequirement(w.ConfigInt("crystals.ganon", 7)); err != nil {
		return err
	}

	// Spike-cave and cape magic usage with default values from PHP.
	if err := r.SetCaneOfByrnaSpikeCaveUsage(0x04, 0x02, 0x01); err != nil {
		return err
	}
	if err := r.SetCapeSpikeCaveUsage(0x04, 0x08, 0x10); err != nil {
		return err
	}
	if err := r.SetByrnaCaveSpikeDamage(0x08); err != nil {
		return err
	}
	if err := r.Write(0x45C42, []byte{0x04, 0x02, 0x01}, true); err != nil {
		return err
	}
	if err := r.SetCapeRegularMagicUsage(
		w.ConfigInt("rom.CapeMagicUsage.Normal", 0x04),
		w.ConfigInt("rom.CapeMagicUsage.Half", 0x08),
		w.ConfigInt("rom.CapeMagicUsage.Quarter", 0x10),
	); err != nil {
		return err
	}
	if err := r.SetCaneOfByrnaInvulnerability(w.ConfigBool("rom.CaneOfByrnaInvulnerability", true)); err != nil {
		return err
	}
	if err := r.SetPowderedSpriteFairyPrize(w.ConfigInt("rom.PowderedSpriteFairyPrize", 0xE3)); err != nil {
		return err
	}
	if err := r.SetStunItems(w.ConfigInt("rom.StunItems", 0x03)); err != nil {
		return err
	}
	if err := r.SetRupoorValue(w.ConfigInt("item.value.Rupoor", 0)); err != nil {
		return err
	}
	if err := r.SetGanonAgahnimRng(w.ConfigString("rom.GanonAgRNG", "table")); err != nil {
		return err
	}
	if err := r.SetCatchableBees(w.ConfigBool("rom.CatchableBees", true)); err != nil {
		return err
	}

	// 3. Ganon invincibility per goal. PHP switch in World::writeToRom.
	var inv string
	switch w.ConfigString("goal", "") {
	case "pedestal", "triforce-hunt":
		inv = "yes"
	case "dungeons":
		inv = "dungeons"
	case "ganonhunt":
		inv = "triforce_pieces"
	case "fast_ganon":
		inv = "crystals_only"
	case "completionist":
		inv = "completionist"
	default:
		inv = "crystals_only"
	}
	if err := r.SetGanonInvincible(inv); err != nil {
		return err
	}

	// 4. Item limits + substitutions.
	replacement := func(key, def string) byte {
		name := w.ConfigString(key, def)
		it, err := ir.Get(name, w.id)
		if err != nil {
			return 0x47 // TwentyRupees2 fallback byte.
		}
		bytes := it.GetBytes()
		if len(bytes) == 0 {
			return 0
		}
		return byte(bytes[0])
	}
	swordRepl := replacement("item.overflow.replacement.Sword", "TwentyRupees2")
	shieldRepl := replacement("item.overflow.replacement.Shield", "TwentyRupees2")
	armorRepl := replacement("item.overflow.replacement.Armor", "TwentyRupees2")
	bottleRepl := replacement("item.overflow.replacement.Bottle", "TwentyRupees2")
	bowRepl := replacement("item.overflow.replacement.Bow", "TwentyRupees2")
	bossHeartRepl := replacement("item.overflow.replacement.BossHeartContainer", "TwentyRupees2")
	pohRepl := replacement("item.overflow.replacement.PieceOfHeart", "TwentyRupees2")

	if err := r.SetLimitProgressiveSword(w.ConfigInt("item.overflow.count.Sword", 4), int(swordRepl)); err != nil {
		return err
	}
	if err := r.SetLimitProgressiveShield(w.ConfigInt("item.overflow.count.Shield", 3), int(shieldRepl)); err != nil {
		return err
	}
	if err := r.SetLimitProgressiveArmor(w.ConfigInt("item.overflow.count.Armor", 2), int(armorRepl)); err != nil {
		return err
	}
	if err := r.SetLimitBottle(w.ConfigInt("item.overflow.count.Bottle", 4), int(bottleRepl)); err != nil {
		return err
	}
	if err := r.SetLimitProgressiveBow(w.ConfigInt("item.overflow.count.Bow", 2), int(bowRepl)); err != nil {
		return err
	}

	// Substitutions block, PHP layout (see Rom.php setSubstitutions caller).
	silverArrowRepl := byte(0x43)
	if w.ConfigBool("rom.rupeeBow", false) {
		silverArrowRepl = 0x36
	}
	if err := r.SetSubstitutions([]byte{
		0x12, 0x01, 0x35, 0xFF, // lamp -> 5 rupees
		0x51, 0x06, 0x52, 0xFF, // 6x +5 bomb -> +10 bomb
		0x53, 0x06, 0x54, 0xFF, // 6x +5 arrow -> +10 arrow
		0x58, 0x01, silverArrowRepl, 0xFF, // silver arrows -> 1 arrow / 5 rupees
		0x3E, byte(w.ConfigInt("item.overflow.count.BossHeartContainer", 10)), bossHeartRepl, 0xFF,
		0x17, byte(w.ConfigInt("item.overflow.count.PieceOfHeart", 24)), pohRepl, 0xFF,
	}); err != nil {
		return err
	}

	// 5. Mode flags. setGameState in PHP also pokes SRAM (progress
	// indicator/flags/starting entrance), so we mirror that via Sram().
	state := w.ConfigString("mode.state", "standard")
	if err := r.SetGameState(state); err != nil {
		return err
	}
	sram := w.Sram()
	switch state {
	case "open", "retro":
		sram.PreOpenCastleGate()
		sram.SetProgressIndicator(0x02)
		sram.SetProgressFlags(0x14)
		sram.SetStartingEntrance(0x01)
	case "standard":
		sram.SetProgressIndicator(0x00)
		sram.SetProgressFlags(0x00)
		sram.SetStartingEntrance(0x00)
	}
	if w.ConfigString("mode.weapons", "") == "swordless" {
		sram.SetSwordlessCurtains()
	}
	if err := r.SetSwordlessMode(w.ConfigString("mode.weapons", "") == "swordless"); err != nil {
		return err
	}
	if err := r.SetMapMode(w.ConfigBool("rom.mapOnPickup", false)); err != nil {
		return err
	}
	if err := r.SetCompassMode(w.ConfigString("rom.dungeonCount", "off")); err != nil {
		return err
	}
	if err := r.SetGenericKeys(w.ConfigBool("rom.genericKeys", false)); err != nil {
		return err
	}
	if err := r.SetRupeeArrow(w.ConfigBool("rom.rupeeBow", false)); err != nil {
		return err
	}
	if err := r.SetCatchableFairies(w.ConfigBool("rom.CatchableFairies", true)); err != nil {
		return err
	}
	if err := r.SetWishingWellChests(true); err != nil {
		return err
	}
	if err := r.SetWishingWellUpgrade(false); err != nil {
		return err
	}
	if err := r.SetHyliaFairyShop(true); err != nil {
		return err
	}
	if err := r.SetRestrictFairyPonds(true); err != nil {
		return err
	}
	if err := r.SetSilversOnlyAtGanon(w.ConfigBool("rom.SilversOnlyAtGanon", false)); err != nil {
		return err
	}
	if err := r.SetPseudoBoots(w.ConfigBool("pseudoboots", false)); err != nil {
		return err
	}
	if err := r.EnableFastROM(w.ConfigBool("fastrom", true)); err != nil {
		return err
	}
	if err := r.SetGameType("item"); err != nil {
		return err
	}
	if err := r.SetPyramidFairyChests(w.ConfigBool("region.swordsInPool", true)); err != nil {
		return err
	}
	if err := r.SetSmithyQuickItemGive(w.ConfigBool("region.swordsInPool", true)); err != nil {
		return err
	}
	if err := r.SetBottleFills(
		w.ConfigInt("rom.BottleFill.Health", 0xA0),
		w.ConfigInt("rom.BottleFill.Magic", 0x80),
	); err != nil {
		return err
	}
	if err := r.SetClockMode(w.ConfigString("rom.timerMode", "off")); err != nil {
		return err
	}
	if err := w.WritePrizePacks(r); err != nil {
		return err
	}

	// 6. Logic-mode flag.
	if w.ConfigString("mode.state", "") != "inverted" {
		switch w.ConfigString("logic", "NoGlitches") {
		case "MajorGlitches", "HybridMajorGlitches", "NoLogic", "OverworldGlitches":
			if err := r.SetLockAgahnimDoorInEscape(false); err != nil {
				return err
			}
		default:
			if err := r.SetLockAgahnimDoorInEscape(true); err != nil {
				return err
			}
		}
	}

	// 7. Uncle sword/shield removal if placement isn't a sword.
	// PHP uses `instanceof Item\Sword` here, which is false for ItemAlias
	// (UncleSword is an alias, not a Sword subclass). So we check the direct
	// type, not the alias-resolved type.
	uncle := w.locations.Get("Link's Uncle:" + itoaW(w.id))
	if uncle != nil && uncle.HasItem() {
		uncleItem := uncle.Item()
		if uncleItem.Type != item.TypeSword {
			if err := r.RemoveUnclesSword(); err != nil {
				return err
			}
		}
		// Shield removal: PHP `!Shield || !L1SwordAndShield` — only KEEP shield
		// if item is BOTH a direct Shield instance AND named L1SwordAndShield
		// (which never matches since L1SwordAndShield is a Sword, not Shield),
		// so this effectively always runs.
		isL1ss := uncleItem.Name == "L1SwordAndShield"
		if uncleItem.Type != item.TypeShield || !isL1ss {
			if err := r.RemoveUnclesShield(); err != nil {
				return err
			}
		}
	} else {
		if err := r.RemoveUnclesSword(); err != nil {
			return err
		}
		if err := r.RemoveUnclesShield(); err != nil {
			return err
		}
	}

	// 7b. Shops, capacity-upgrade fills, RNG block, free-item text/menu, hud counter,
	// silvers equip, escape fills, ball-and-chain, compass totals, mystery masking,
	// digging-game RNG, full rupee-arrow bytes, logic mode toggles.
	if err := w.WriteShops(r); err != nil {
		return err
	}
	if err := r.SetCapacityUpgradeFills([4]int{
		w.ConfigInt("item.value.BombUpgrade5", 50),
		w.ConfigInt("item.value.BombUpgrade10", 50),
		w.ConfigInt("item.value.ArrowUpgrade5", 70),
		w.ConfigInt("item.value.ArrowUpgrade10", 70),
	}); err != nil {
		return err
	}
	if err := r.WriteRNGBlock(); err != nil {
		return err
	}
	if err := r.SetFreeItemTextMode(w.ConfigInt("rom.freeItemText", 0x00)); err != nil {
		return err
	}
	if err := r.SetFreeItemMenu(w.ConfigInt("rom.freeItemMenu", 0x00)); err != nil {
		return err
	}
	if err := r.SetSilversEquip("collection"); err != nil {
		return err
	}
	if err := r.SetBallNChainDungeon(0x02); err != nil {
		return err
	}
	if err := r.SetCompassCountTotals(nil); err != nil {
		return err
	}
	if err := r.SetMysteryMasking(w.ConfigString("spoilers", "on") == "mystery"); err != nil {
		return err
	}
	// Tournament mode toggle (writes 2 bytes at 0x180213).
	tournamentSetting := "none"
	if w.ConfigBool("tournament", false) {
		tournamentSetting = "standard"
	}
	if err := r.SetTournamentType(tournamentSetting); err != nil {
		return err
	}
	if err := r.RupeeArrowFull(w.ConfigBool("rom.rupeeBow", false)); err != nil {
		return err
	}
	digs, err := helpers.GetRandomInt(1, 30)
	if err != nil {
		return err
	}
	if err := r.SetDiggingGameRng(digs); err != nil {
		return err
	}

	// Triforce-hunt HUD counter (and goal-specific HUD toggle).
	triforceHud := w.ConfigString("goal", "") == "triforce-hunt" || w.ConfigString("goal", "") == "ganonhunt" ||
		w.ConfigInt("item.Goal.Required", 0) > 0
	hud := false
	if !triforceHud {
		hud = w.ConfigBool("rom.hudItemCounter", w.ConfigString("goal", "ganon") == "completionist")
	}
	if err := r.EnableHudItemCounter(hud); err != nil {
		return err
	}

	// Triforce-hunt turn-in.
	switch w.ConfigString("goal", "") {
	case "triforce-hunt":
		if err := r.EnableTriforceTurnIn(true); err != nil {
			return err
		}
	}

	// Logic-mode toggles per PHP World::writeToRom switch on rom.logicMode.
	logicMode := w.ConfigString("rom.logicMode", w.ConfigString("logic", "NoGlitches"))
	switch logicMode {
	case "MajorGlitches", "HybridMajorGlitches", "NoLogic":
		_ = r.SetSwampWaterLevel(false)
		_ = r.SetPreAgahnimDarkWorldDeathInDungeon(false)
		_ = r.SetSaveAndQuitFromBossRoom(true)
		_ = r.SetWorldOnAgahnimDeath(false)
		_ = r.SetRandomizerSeedType("MajorGlitches")
		_ = r.SetWarningFlags(0b01100000)
		_ = r.SetAllowAccidentalMajorGlitch(true)
		_ = r.SetSQEGFix(false)
		_ = r.SetZeldaMirrorFix(false)
	case "OverworldGlitches":
		_ = r.SetPreAgahnimDarkWorldDeathInDungeon(false)
		_ = r.SetSaveAndQuitFromBossRoom(true)
		_ = r.SetWorldOnAgahnimDeath(false)
		_ = r.SetRandomizerSeedType("OverworldGlitches")
		_ = r.SetWarningFlags(0b01000000)
		_ = r.SetAllowAccidentalMajorGlitch(true)
		_ = r.SetSQEGFix(false)
		_ = r.SetZeldaMirrorFix(false)
	default: // NoGlitches
		_ = r.SetSaveAndQuitFromBossRoom(true)
		_ = r.SetWorldOnAgahnimDeath(true)
		_ = r.SetAllowAccidentalMajorGlitch(false)
		_ = r.SetSQEGFix(true)
		_ = r.SetZeldaMirrorFix(true)
	}

	// 7c. Escape fills + spawn refills, based on Uncle's item.
	// Mirrors PHP World::setEscapeFills (only for standard mode).
	if w.ConfigString("mode.state", "") == "standard" {
		if err := w.writeEscapeFills(r); err != nil {
			return err
		}
	}

	// 8. Default text + credits + initial SRAM + total item count + checksum.
	if err := w.Text().WriteTo(r); err != nil {
		return err
	}
	if err := w.Credits().WriteTo(r); err != nil {
		return err
	}

	// SRAM: apply pre-collected starting equipment, then handle goal/mode
	// SRAM modifications (pre-open pyramid for fast_ganon/ganonhunt,
	// pre-open GT when crystals.tower=0), then starting timer.
	rupeeBow := w.ConfigBool("rom.rupeeBow", false)
	weaponsMode := w.ConfigString("mode.weapons", "")
	sram.SetStartingEquipment(w.preCollected, weaponsMode, rupeeBow)

	switch w.ConfigString("goal", "") {
	case "fast_ganon", "ganonhunt":
		sram.PreOpenPyramid()
	}
	if w.ConfigInt("crystals.tower", 7) == 0 {
		sram.PreOpenGanonsTower()
	}
	sram.SetStartingTimer(w.ConfigInt("rom.timerStart", 0))

	if err := sram.WriteTo(r); err != nil {
		return err
	}
	if err := r.SetTotalItemCount(w.totalItemCount()); err != nil {
		return err
	}
	return r.UpdateChecksum()
}

// totalItemCount counts collectables for the in-game total counter.
// PHP World::getTotalItemCount — counts items with "item get" animation.
// Simple approximation: count non-empty locations except prize/medallion/trade.
func (w *World) totalItemCount() int {
	count := 0
	for _, l := range w.locations.All() {
		if !l.HasItem() {
			continue
		}
		switch l.Kind {
		case KindPrize, KindPrizeEvent, KindPrizeCrystal, KindPrizePendant, KindMedallion, KindTrade:
			continue
		}
		count++
	}
	return count
}

// Avoid lint warning for unused helper if WriteToRom paths don't trigger it.
var _ = helpers.GetRandomInt
