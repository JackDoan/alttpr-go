package world

import (
	"github.com/JackDoan/alttpr-go/internal/helpers"
	"github.com/JackDoan/alttpr-go/internal/item"
	"github.com/JackDoan/alttpr-go/internal/rom"
)

// InitialSram is the 0x500-byte starting save block written to ROM at 0x183000.
// Mirrors app/Support/InitialSram.php.
type InitialSram struct {
	bytes [0x500]byte
}

const (
	sramRoomData     = 0x000
	sramOverworldDat = 0x280
)

// NewInitialSram returns the default-initialized SRAM block: pre-opens
// Kakariko bomb hut + brewery and sets default ability flags.
// Mirrors PHP InitialSram::__construct.
func NewInitialSram() *InitialSram {
	s := &InitialSram{}
	s.bytes[sramRoomData+0x20D] = 0xF0
	s.bytes[sramRoomData+0x20F] = 0xF0
	s.bytes[0x379] = 0b01101000
	s.bytes[0x401] = 0xFF
	s.bytes[0x402] = 0xFF
	return s
}

func (s *InitialSram) orValue(idx int, val byte) { s.bytes[idx] |= val }
func (s *InitialSram) add(idx int, val int, cap int) {
	s.bytes[idx] = byte(min(int(s.bytes[idx])+val, cap))
}

// PreOpenAgaCurtains.
func (s *InitialSram) PreOpenAgaCurtains() { s.orValue(sramRoomData+0x61, 0x80) }

// PreOpenSkullWoodsCurtains.
func (s *InitialSram) PreOpenSkullWoodsCurtains() { s.orValue(sramRoomData+0x93, 0x80) }

// PreOpenCastleGate.
func (s *InitialSram) PreOpenCastleGate() { s.orValue(sramOverworldDat+0x1B, 0x20) }

// PreOpenGanonsTower.
func (s *InitialSram) PreOpenGanonsTower() { s.orValue(sramOverworldDat+0x43, 0x20) }

// PreOpenPyramid.
func (s *InitialSram) PreOpenPyramid() { s.orValue(sramOverworldDat+0x5B, 0x20) }

// SetProgressIndicator.
func (s *InitialSram) SetProgressIndicator(v byte) { s.orValue(0x3C5, v) }

// SetProgressFlags.
func (s *InitialSram) SetProgressFlags(v byte) { s.orValue(0x3C6, v) }

// SetStartingEntrance.
func (s *InitialSram) SetStartingEntrance(v byte) { s.orValue(0x3C8, v) }

// SetStartingTimer writes seconds*60 as a 32-bit LE value at 0x454.
func (s *InitialSram) SetStartingTimer(seconds int) {
	v := uint32(seconds * 60)
	s.bytes[0x454] = byte(v)
	s.bytes[0x455] = byte(v >> 8)
	s.bytes[0x456] = byte(v >> 16)
	s.bytes[0x457] = byte(v >> 24)
}

// SetSwordlessCurtains opens Aga + Skull Woods curtains.
func (s *InitialSram) SetSwordlessCurtains() {
	s.orValue(sramRoomData+0x61, 0x80)
	s.orValue(sramRoomData+0x93, 0x80)
}

// SetInstantPostAga sets post-Agahnim world state per PHP setInstantPostAga.
func (s *InitialSram) SetInstantPostAga(state string) {
	switch state {
	case "standard":
		s.orValue(0x3C5, 0x80)
		s.orValue(sramOverworldDat+0x02, 0x20)
	default: // open/retro/inverted
		s.orValue(0x3C5, 0x03)
		s.orValue(sramOverworldDat+0x02, 0x20)
	}
}

// SetStartingEquipment applies a pre-collected items pool to the SRAM.
// Mirrors PHP InitialSram::setStartingEquipment exactly (case-by-case).
func (s *InitialSram) SetStartingEquipment(items *item.Collection, weaponsMode string, rupeeBow bool) {
	startingRupees := 0
	startingArrowCapacity := 0
	startingBombCapacity := 0

	if items.HeartCount(0) < 1 {
		s.bytes[0x36C] = 0x18
		s.bytes[0x36D] = 0x18
	}

	items.Each(func(it *item.Item) {
		name := it.Name
		if it.Type == item.TypeAlias && it.Target != nil {
			name = it.Target.Name
		}
		switch name {
		case "L1Sword":
			s.bytes[0x359] = 0x01
			s.bytes[0x417] = 0x01
		case "L1SwordAndShield":
			s.bytes[0x359] = 0x01
			s.bytes[0x35A] = 0x01
			s.bytes[0x417] = 0x01
			s.bytes[0x422] = 0x01
		case "L2Sword", "MasterSword":
			s.bytes[0x359] = 0x02
			s.bytes[0x417] = 0x02
		case "L3Sword":
			s.bytes[0x359] = 0x03
			s.bytes[0x417] = 0x03
		case "L4Sword":
			s.bytes[0x359] = 0x04
			s.bytes[0x417] = 0x04
		case "BlueShield":
			s.bytes[0x35A] = 0x01
			s.bytes[0x422] = 0x01
		case "RedShield":
			s.bytes[0x35A] = 0x02
			s.bytes[0x422] = 0x02
		case "MirrorShield":
			s.bytes[0x35A] = 0x03
			s.bytes[0x422] = 0x03
		case "FireRod":
			s.bytes[0x345] = 0x01
		case "IceRod":
			s.bytes[0x346] = 0x01
		case "Hammer":
			s.bytes[0x34B] = 0x01
		case "Hookshot":
			s.bytes[0x342] = 0x01
		case "Bow":
			s.bytes[0x340] = 0x01
			if !rupeeBow {
				s.bytes[0x38E] |= 0b10000000
			}
		case "BowAndArrows":
			s.bytes[0x340] = 0x02
			s.bytes[0x38E] |= 0b10000000
			if rupeeBow {
				s.bytes[0x377] = 0x01
			}
		case "SilverArrowUpgrade":
			s.bytes[0x38E] |= 0b01000000
			if rupeeBow {
				s.bytes[0x377] = 0x01
			}
		case "BowAndSilverArrows":
			s.bytes[0x340] = 0x04
			s.bytes[0x38E] |= 0b01000000
			if rupeeBow {
				s.bytes[0x377] = 0x01
			} else {
				s.bytes[0x38E] |= 0b10000000
			}
		case "ProgressiveBow":
			// Note PHP `min()` is a no-op (missing assignment); we mirror that.
			if rupeeBow {
				s.bytes[0x377] = 0x01
			} else {
				s.bytes[0x38E] = 0b10000000
			}
		case "Boomerang":
			s.bytes[0x341] = 0x01
			s.bytes[0x38C] |= 0b10000000
		case "RedBoomerang":
			s.bytes[0x341] = 0x02
			s.bytes[0x38C] |= 0b01000000
		case "Mushroom":
			s.bytes[0x344] = 0x01
			s.bytes[0x38C] |= 0b00101000
		case "Powder":
			s.bytes[0x344] = 0x02
			s.bytes[0x38C] |= 0b00010000
		case "Bombos":
			s.bytes[0x347] = 0x01
		case "Ether":
			s.bytes[0x348] = 0x01
		case "Quake":
			s.bytes[0x349] = 0x01
		case "Lamp":
			s.bytes[0x34A] = 0x01
		case "Shovel":
			s.bytes[0x34C] = 0x01
			s.bytes[0x38C] |= 0b00000100
		case "OcarinaInactive":
			s.bytes[0x34C] = 0x02
			s.bytes[0x38C] |= 0b00000010
		case "OcarinaActive":
			s.bytes[0x34C] = 0x03
			s.bytes[0x38C] |= 0b00000001
		case "CaneOfSomaria":
			s.bytes[0x350] = 0x01
		case "Bottle":
			s.bottleSlot(0x02)
		case "BottleWithRedPotion":
			s.bottleSlot(0x03)
		case "BottleWithGreenPotion":
			s.bottleSlot(0x04)
		case "BottleWithBluePotion":
			s.bottleSlot(0x05)
		case "BottleWithBee":
			s.bottleSlot(0x07)
		case "BottleWithFairy":
			s.bottleSlot(0x06)
		case "BottleWithGoldBee":
			s.bottleSlot(0x08)
		case "CaneOfByrna":
			s.bytes[0x351] = 0x01
		case "Cape":
			s.bytes[0x352] = 0x01
		case "MagicMirror":
			s.bytes[0x353] = 0x02
		case "PowerGlove":
			s.bytes[0x354] = 0x01
		case "TitansMitt":
			s.bytes[0x354] = 0x02
		case "BookOfMudora":
			s.bytes[0x34E] = 0x01
		case "Flippers":
			s.bytes[0x356] = 0x01
			s.bytes[0x379] |= 0b00000010
		case "MoonPearl":
			s.bytes[0x357] = 0x01
		case "BugCatchingNet":
			s.bytes[0x34D] = 0x01
		case "BlueMail":
			s.bytes[0x35B] = 0x01
			s.bytes[0x46E] = 0x01
		case "RedMail":
			s.bytes[0x35B] = 0x02
			s.bytes[0x46E] = 0x02
		case "Bomb":
			s.add(0x343, 1, 99)
			s.bytes[0x38D] |= 0b00000010
		case "ThreeBombs":
			s.add(0x343, 3, 99)
			s.bytes[0x38D] |= 0b00000010
		case "TenBombs":
			s.add(0x343, 10, 99)
			s.bytes[0x38D] |= 0b00000010
		case "OneRupee":
			startingRupees += 1
		case "FiveRupees":
			startingRupees += 5
		case "TwentyRupees", "TwentyRupees2":
			startingRupees += 20
		case "FiftyRupees":
			startingRupees += 50
		case "OneHundredRupees":
			startingRupees += 100
		case "PendantOfCourage":
			s.bytes[0x374] |= 0b00000100
			s.add(0x429, 1, 3)
		case "PendantOfWisdom":
			s.bytes[0x374] |= 0b00000001
			s.add(0x429, 1, 3)
		case "PendantOfPower":
			s.bytes[0x374] |= 0b00000010
			s.add(0x429, 1, 3)
		case "HeartContainerNoAnimation", "BossHeartContainer", "HeartContainer":
			s.add(0x36C, 0x08, 0xA0)
			s.add(0x36D, 0x08, 0xA0)
		case "PieceOfHeart":
			s.bytes[0x36B] += 1
			if s.bytes[0x36B] >= 4 {
				inc := int(s.bytes[0x36B]/4) * 0x08
				s.add(0x36C, inc, 0xA0)
				s.bytes[0x36B] = s.bytes[0x36B] % 4
			}
		case "Heart":
			s.add(0x36D, 0x08, 0xA0)
		case "Arrow":
			s.add(0x377, 1, 99)
		case "TenArrows":
			s.add(0x377, 10, 99)
		case "SmallMagic":
			s.add(0x36E, 0x10, 0x80)
		case "ThreeHundredRupees":
			startingRupees += 300
		case "PegasusBoots":
			s.bytes[0x355] = 0x01
			s.bytes[0x379] |= 0b00000100
		case "BombUpgrade5":
			startingBombCapacity += 5
		case "BombUpgrade10":
			startingBombCapacity += 10
		case "ArrowUpgrade5":
			startingArrowCapacity += 5
		case "ArrowUpgrade10":
			startingArrowCapacity += 10
		case "HalfMagic":
			s.bytes[0x37B] = 0x01
		case "QuarterMagic":
			s.bytes[0x37B] = 0x02
		case "ProgressiveSword":
			s.add(0x359, 1, 4)
		case "ProgressiveShield":
			s.add(0x35A, 1, 3)
		case "ProgressiveArmor":
			s.add(0x35B, 1, 2)
		case "ProgressiveGlove":
			s.add(0x354, 1, 2)
		case "MapLW":
			s.bytes[0x368] |= 0b00000001
		case "MapDW":
			s.bytes[0x368] |= 0b00000010
		case "MapA2":
			s.bytes[0x368] |= 0b00000100
		case "MapD7":
			s.bytes[0x368] |= 0b00001000
		case "MapD4":
			s.bytes[0x368] |= 0b00010000
		case "MapP3":
			s.bytes[0x368] |= 0b00100000
		case "MapD5":
			s.bytes[0x368] |= 0b01000000
		case "MapD3":
			s.bytes[0x368] |= 0b10000000
		case "MapD6":
			s.bytes[0x369] |= 0b00000001
		case "MapD1":
			s.bytes[0x369] |= 0b00000010
		case "MapD2":
			s.bytes[0x369] |= 0b00000100
		case "MapA1":
			s.bytes[0x369] |= 0b00001000
		case "MapP2":
			s.bytes[0x369] |= 0b00010000
		case "MapP1":
			s.bytes[0x369] |= 0b00100000
		case "MapH1", "MapH2":
			s.bytes[0x369] |= 0b11000000
		case "CompassA2":
			s.bytes[0x364] |= 0b00000100
		case "CompassD7":
			s.bytes[0x364] |= 0b00001000
		case "CompassD4":
			s.bytes[0x364] |= 0b00010000
		case "CompassP3":
			s.bytes[0x364] |= 0b00100000
		case "CompassD5":
			s.bytes[0x364] |= 0b01000000
		case "CompassD3":
			s.bytes[0x364] |= 0b10000000
		case "CompassD6":
			s.bytes[0x365] |= 0b00000001
		case "CompassD1":
			s.bytes[0x365] |= 0b00000010
		case "CompassD2":
			s.bytes[0x365] |= 0b00000100
		case "CompassA1":
			s.bytes[0x365] |= 0b00001000
		case "CompassP2":
			s.bytes[0x365] |= 0b00010000
		case "CompassP1":
			s.bytes[0x365] |= 0b00100000
		case "CompassH1", "CompassH2":
			s.bytes[0x365] |= 0b11000000
		case "BigKeyA2":
			s.bytes[0x366] |= 0b00000100
		case "BigKeyD7":
			s.bytes[0x366] |= 0b00001000
		case "BigKeyD4":
			s.bytes[0x366] |= 0b00010000
		case "BigKeyP3":
			s.bytes[0x366] |= 0b00100000
		case "BigKeyD5":
			s.bytes[0x366] |= 0b01000000
		case "BigKeyD3":
			s.bytes[0x366] |= 0b10000000
		case "BigKeyD6":
			s.bytes[0x367] |= 0b00000001
		case "BigKeyD1":
			s.bytes[0x367] |= 0b00000010
		case "BigKeyD2":
			s.bytes[0x367] |= 0b00000100
		case "BigKeyA1":
			s.bytes[0x367] |= 0b00001000
		case "BigKeyP2":
			s.bytes[0x367] |= 0b00010000
		case "BigKeyP1":
			s.bytes[0x367] |= 0b00100000
		case "BigKeyH1", "BigKeyH2":
			s.bytes[0x367] |= 0b11000000
		case "KeyH1", "KeyH2":
			s.bytes[0x37C] += 1
			s.bytes[0x37D] += 1
		case "KeyP1":
			s.bytes[0x37E] += 1
		case "KeyP2":
			s.bytes[0x37F] += 1
		case "KeyA1":
			s.bytes[0x380] += 1
		case "KeyD2":
			s.bytes[0x381] += 1
		case "KeyD1":
			s.bytes[0x382] += 1
		case "KeyD6":
			s.bytes[0x383] += 1
		case "KeyD3":
			s.bytes[0x384] += 1
		case "KeyD5":
			s.bytes[0x385] += 1
		case "KeyP3":
			s.bytes[0x386] += 1
		case "KeyD4":
			s.bytes[0x387] += 1
		case "KeyD7":
			s.bytes[0x388] += 1
		case "KeyA2":
			s.bytes[0x389] += 1
		case "Crystal1":
			s.bytes[0x37A] |= 0b00000010
		case "Crystal2":
			s.bytes[0x37A] |= 0b00010000
		case "Crystal3":
			s.bytes[0x37A] |= 0b01000000
		case "Crystal4":
			s.bytes[0x37A] |= 0b00100000
		case "Crystal5":
			s.bytes[0x37A] |= 0b00000100
		case "Crystal6":
			s.bytes[0x37A] |= 0b00000001
		case "Crystal7":
			s.bytes[0x37A] |= 0b00001000
		}
	})

	s.bytes[0x362] = byte(startingRupees & 0xFF)
	s.bytes[0x360] = s.bytes[0x362]
	s.bytes[0x363] = byte((startingRupees >> 8) & 0xFF)
	s.bytes[0x361] = s.bytes[0x363]

	// Counters + highest equipment.
	s.bytes[0x476] = byte(helpers.CountSetBits(int(s.bytes[0x37A])))
	s.bytes[0x429] = byte(helpers.CountSetBits(int(s.bytes[0x374])))
	s.bytes[0x417] = s.bytes[0x359]
	s.bytes[0x422] = s.bytes[0x35A]
	s.bytes[0x46E] = s.bytes[0x35B]

	s.orValue(0x370, byte(startingBombCapacity))
	s.orValue(0x371, byte(startingArrowCapacity))

	if weaponsMode == "swordless" {
		s.bytes[0x359] = 0xFF
		s.bytes[0x417] = 0x00
	}
}

// bottleSlot fills the next free bottle slot (max 4) with `kind`.
func (s *InitialSram) bottleSlot(kind byte) {
	if s.bytes[0x34F] >= 4 {
		return
	}
	s.bytes[0x35C+int(s.bytes[0x34F])] = kind
	s.bytes[0x34F] += 1
}

// Bytes returns the underlying SRAM buffer.
func (s *InitialSram) Bytes() []byte { return s.bytes[:] }

// WriteTo writes the SRAM block to ROM at 0x183000. Mirrors PHP Rom::writeInitialSram.
func (s *InitialSram) WriteTo(r *rom.ROM) error {
	return r.Write(0x183000, s.bytes[:], true)
}
