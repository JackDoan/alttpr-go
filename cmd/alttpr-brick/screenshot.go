package main

import (
	"fmt"
	"image"
	"image/color"
	"image/png"
	"os"
	"path/filepath"

	"github.com/JackDoan/alttpr-go/cmd/alttpr-brick/font"
	"github.com/JackDoan/alttpr-go/cmd/alttpr-brick/input"
	"github.com/JackDoan/alttpr-go/cmd/alttpr-brick/ui"
	"github.com/JackDoan/alttpr-go/internal/job"
)

// pngRenderer implements ui.Renderer by writing into an image.RGBA. Used
// by the -screenshot flow so the harness can produce documentation
// images on a desktop without /dev/fb0 or evdev.
type pngRenderer struct{ img *image.RGBA }

func newPNGRenderer(w, h int) *pngRenderer {
	return &pngRenderer{img: image.NewRGBA(image.Rect(0, 0, w, h))}
}

func (r *pngRenderer) Bounds() (int, int) {
	b := r.img.Bounds()
	return b.Dx(), b.Dy()
}

func (r *pngRenderer) Clear(c ui.Color) {
	w, h := r.Bounds()
	r.FillRect(0, 0, w, h, c)
}

func (r *pngRenderer) FillRect(x, y, w, h int, c ui.Color) {
	col := color.RGBA{R: c.R, G: c.G, B: c.B, A: 0xFF}
	x0, y0 := clamp(x, 0, r.img.Bounds().Dx()), clamp(y, 0, r.img.Bounds().Dy())
	x1, y1 := clamp(x+w, 0, r.img.Bounds().Dx()), clamp(y+h, 0, r.img.Bounds().Dy())
	for py := y0; py < y1; py++ {
		for px := x0; px < x1; px++ {
			r.img.SetRGBA(px, py, col)
		}
	}
}

func (r *pngRenderer) DrawText(x, y int, text string, c ui.Color) {
	col := color.RGBA{R: c.R, G: c.G, B: c.B, A: 0xFF}
	cx := x
	w, h := r.Bounds()
	for _, rn := range text {
		for py := 0; py < font.CellHeight; py++ {
			yy := y + py
			if yy < 0 || yy >= h {
				continue
			}
			for px := 0; px < font.CellWidth; px++ {
				if font.Pixel(rn, px, py) == 0 {
					continue
				}
				xx := cx + px
				if xx < 0 || xx >= w {
					continue
				}
				r.img.SetRGBA(xx, yy, col)
			}
		}
		cx += font.CellWidth
	}
}

func (r *pngRenderer) save(path string) error {
	f, err := os.Create(path)
	if err != nil {
		return err
	}
	defer f.Close()
	return png.Encode(f, r.img)
}

func clamp(v, lo, hi int) int {
	if v < lo {
		return lo
	}
	if v > hi {
		return hi
	}
	return v
}

// runScreenshots drives the menu through a representative tour and
// writes one PNG per screen to outDir. No real framebuffer or evdev
// involved — this is for documentation captures on a desktop.
//
// Each named scene gets its own file. The on-device screen is 1024x768
// on the Trimui Brick; we use that here so previews match what users
// will actually see.
func runScreenshots(outDir string) error {
	if err := os.MkdirAll(outDir, 0o755); err != nil {
		return err
	}
	const w, h = 1024, 768

	type scene struct {
		name string
		fn   func(m *ui.Model)
	}

	// Sample reveal entries spanning every tab so the Reveal screen shows
	// real-looking data. Locations are made up but plausible.
	sampleReveal := []ui.RevealEntry{
		{Item: "Bow", Location: "Pyramid Fairy", Category: ui.CatItems},
		{Item: "Hookshot", Location: "Swamp Palace - Big Chest", Category: ui.CatItems},
		{Item: "PegasusBoots", Location: "Spiral Cave", Category: ui.CatItems},
		{Item: "Lamp", Location: "Kakariko Tavern", Category: ui.CatItems},
		{Item: "MoonPearl", Location: "Library", Category: ui.CatItems},
		{Item: "ProgressiveSword", Location: "Uncle", Category: ui.CatItems},
		{Item: "Bottle", Location: "Bottle Vendor", Category: ui.CatItems},
		{Item: "BigKeyD1", Location: "Skull Woods - Compass Chest", Category: ui.CatDungeon},
		{Item: "KeyD3", Location: "Thieves' Town - Big Key Chest", Category: ui.CatDungeon},
		{Item: "MapP2", Location: "Eastern Palace - Map Chest", Category: ui.CatDungeon},
		{Item: "CompassA2", Location: "Ganons Tower - Bob's Chest", Category: ui.CatDungeon},
		{Item: "Crystal3", Location: "Palace of Darkness", Category: ui.CatPrizes},
		{Item: "PendantOfCourage", Location: "Eastern Palace", Category: ui.CatPrizes},
		{Item: "PieceOfHeart", Location: "Hyrule Castle - Floor 2", Category: ui.CatHearts},
		{Item: "HeartContainer", Location: "Misery Mire", Category: ui.CatHearts},
		{Item: "BossHeartContainer", Location: "Tower of Hera", Category: ui.CatHearts},
		{Item: "FiftyRupees", Location: "Desert Palace - Big Chest", Category: ui.CatJunk},
		{Item: "ThreeBombs", Location: "Cave 45", Category: ui.CatJunk},
		{Item: "TenArrows", Location: "Dam Switch Room", Category: ui.CatJunk},
	}

	scenes := []scene{
		{"01-main.png", func(m *ui.Model) {}},
		{"02-gameplay.png", func(m *ui.Model) {
			m.Step(input.BtnDown) // Gameplay Settings
			m.Step(input.BtnA)
		}},
		{"03-cosmetic.png", func(m *ui.Model) {
			m.Step(input.BtnDown)
			m.Step(input.BtnDown) // Cosmetic Settings
			m.Step(input.BtnA)
		}},
		{"04-spoiler-list.png", func(m *ui.Model) {
			m.Step(input.BtnDown)
			m.Step(input.BtnDown)
			m.Step(input.BtnDown) // Reveal Spoiler
			m.Step(input.BtnA)
			m.SetSpoilerList([]string{
				"alttpr_none_standard_ganon_72866444.json",
				"alttpr_none_open_pedestal_71924011.json",
				"alttpr_none_standard_fast_ganon_71500001.json",
			})
		}},
		{"05-reveal-all.png", func(m *ui.Model) {
			m.SetRevealEntries("alttpr_none_standard_ganon_72866444.json", clone(sampleReveal))
			// Reveal a few entries so the screenshot shows both states.
			m.Step(input.BtnA)
			m.Step(input.BtnDown)
			m.Step(input.BtnA)
			m.Step(input.BtnDown)
			m.Step(input.BtnDown)
			m.Step(input.BtnA)
		}},
		{"06-reveal-items.png", func(m *ui.Model) {
			m.SetRevealEntries("alttpr_none_standard_ganon_72866444.json", clone(sampleReveal))
			m.Step(input.BtnRight) // Items tab
		}},
		{"07-reveal-dungeon.png", func(m *ui.Model) {
			m.SetRevealEntries("alttpr_none_standard_ganon_72866444.json", clone(sampleReveal))
			m.Step(input.BtnRight)
			m.Step(input.BtnRight) // Dungeon tab
			// Reveal them all to demo Start.
			m.Step(input.BtnStart)
		}},
		{"08-result.png", func(m *ui.Model) {
			m.SetResult(true, "Done", []string{
				"ROM: alttpr_none_standard_ganon_72866444.sfc",
				"Spoiler: alttpr_none_standard_ganon_72866444.json",
				"Dropped in: /mnt/SDCARD/Roms/SFC",
			})
		}},
	}

	for _, s := range scenes {
		m := ui.New(job.DefaultOptions())
		s.fn(m)
		r := newPNGRenderer(w, h)
		m.Render(r)
		path := filepath.Join(outDir, s.name)
		if err := r.save(path); err != nil {
			return fmt.Errorf("save %s: %w", path, err)
		}
		fmt.Println("wrote", path)
	}
	return nil
}

func clone(s []ui.RevealEntry) []ui.RevealEntry {
	out := make([]ui.RevealEntry, len(s))
	copy(out, s)
	return out
}
