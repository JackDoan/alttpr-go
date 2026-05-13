package world

import (
	"github.com/JackDoan/alttpr-go/internal/helpers"
	"github.com/JackDoan/alttpr-go/internal/rom"
)

// dropByte maps a sprite name to the byte the prize pack writes. Mirrors
// PHP `Sprite::get($name)->getBytes()[0]` for the Droppable sprite subset.
var dropByte = map[string]byte{
	"Heart":             0xD8,
	"RupeeGreen":        0xD9,
	"RupeeBlue":         0xDA,
	"RupeeRed":          0xDB,
	"BombRefill1":       0xDC,
	"BombRefill4":       0xDD,
	"BombRefill8":       0xDE,
	"MagicRefillSmall":  0xDF,
	"MagicRefillFull":   0xE0,
	"ArrowRefill5":      0xE1,
	"ArrowRefill10":     0xE2,
	"Fairy":             0xE3,
}

// defaultPrizePacks is the vanilla 7-pack drop layout (mirrors PHP
// Randomizer::shufflePrizePacks).
var defaultPrizePacks = [][]string{
	{"Heart", "Heart", "Heart", "Heart", "RupeeGreen", "Heart", "Heart", "RupeeGreen"},
	{"RupeeBlue", "RupeeGreen", "RupeeBlue", "RupeeRed", "RupeeBlue", "RupeeGreen", "RupeeBlue", "RupeeBlue"},
	{"MagicRefillFull", "MagicRefillSmall", "MagicRefillSmall", "RupeeBlue", "MagicRefillFull", "MagicRefillSmall", "Heart", "MagicRefillSmall"},
	{"BombRefill1", "BombRefill1", "BombRefill1", "BombRefill4", "BombRefill1", "BombRefill1", "BombRefill8", "BombRefill1"},
	{"ArrowRefill5", "Heart", "ArrowRefill5", "ArrowRefill10", "ArrowRefill5", "Heart", "ArrowRefill5", "ArrowRefill10"},
	{"MagicRefillSmall", "RupeeGreen", "Heart", "ArrowRefill5", "MagicRefillSmall", "BombRefill1", "RupeeGreen", "Heart"},
	{"Heart", "Fairy", "MagicRefillFull", "RupeeRed", "BombRefill8", "Heart", "RupeeRed", "ArrowRefill10"},
}

// Default tail values (PHP World::$drops static defaults) — these fill
// drop_bytes indexes 56-62 for trees/crab/stunned/fish writes.
var defaultDropTail = []string{
	"RupeeGreen", "RupeeBlue", "RupeeRed", // pull-tree prizes (3)
	"RupeeGreen", "RupeeRed", // rupee-crab prizes (2)
	"RupeeGreen",             // stunned sprite prize (1)
	"RupeeGreen",             // fish-save prize (1)
}

// drops is the per-world prize-pack assignment. Index 0..6 = 7 packs of 8.
// Indexes 56-62 = tree/crab/stunned/fish defaults.
type drops struct {
	packs    [][]string // 7 packs of 8
	tail     []string   // 7 tail entries
}

// PrizePacks returns the per-world prize-pack state, initialized to vanilla.
func (w *World) PrizePacks() *drops {
	if w.prizes == nil {
		w.prizes = &drops{
			packs: make([][]string, 7),
			tail:  append([]string(nil), defaultDropTail...),
		}
		for i, p := range defaultPrizePacks {
			w.prizes.packs[i] = append([]string(nil), p...)
		}
	}
	return w.prizes
}

// ShufflePrizePacks does PHP `Randomizer::shufflePrizePacks` — fy_shuffle the
// 7 vanilla packs and assign them to keys 0..6 in shuffle order.
func (w *World) ShufflePrizePacks() error {
	if w.ConfigBool("customPrizePacks", false) {
		return nil
	}
	pp := w.PrizePacks()
	indices := []int{0, 1, 2, 3, 4, 5, 6}
	shuffled, err := helpers.FyShuffle(indices)
	if err != nil {
		return err
	}
	original := append([][]string(nil), pp.packs...)
	for slot, src := range shuffled {
		pp.packs[slot] = append([]string(nil), original[src]...)
	}
	return nil
}

// WritePrizePacks writes the prize-pack table to ROM at 0x37A78 + tail values
// via dedicated Rom setters. Mirrors PHP World::writePrizePacksToRom.
func (w *World) WritePrizePacks(r *rom.ROM) error {
	pp := w.PrizePacks()

	// Build the 56-byte main packs.
	bytes56 := make([]byte, 0, 56)
	for _, pack := range pp.packs {
		for _, name := range pack {
			b, ok := dropByte[name]
			if !ok {
				b = 0xD8 // unknown -> heart
			}
			bytes56 = append(bytes56, b)
		}
	}

	// Build the 7 tail bytes.
	tail := make([]byte, 7)
	for i, name := range pp.tail {
		b, ok := dropByte[name]
		if !ok {
			b = 0xD8
		}
		tail[i] = b
	}

	// rom.NoFarieDrops + rom.rupeeBow rewrites.
	if w.ConfigBool("rom.NoFarieDrops", false) {
		for i, b := range bytes56 {
			if b == 0xE0 {
				bytes56[i] = 0xDF
			} else if b == 0xE3 {
				bytes56[i] = 0xD8
			}
		}
	}
	if w.ConfigBool("rom.rupeeBow", false) {
		for i, b := range bytes56 {
			if b == 0xE1 {
				bytes56[i] = 0xDA
			} else if b == 0xE2 {
				bytes56[i] = 0xDB
			}
		}
	}

	if err := r.Write(0x37A78, bytes56, true); err != nil {
		return err
	}
	if err := r.SetPullTreePrizes(int(tail[0]), int(tail[1]), int(tail[2])); err != nil {
		return err
	}
	if err := r.SetRupeeCrabPrizes(int(tail[3]), int(tail[4])); err != nil {
		return err
	}
	if err := r.SetStunnedSpritePrize(int(tail[5])); err != nil {
		return err
	}
	return r.SetFishSavePrize(int(tail[6]))
}
