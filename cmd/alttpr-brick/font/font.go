// Package font wraps golang.org/x/image/font/basicfont.Face7x13 in a
// blit-only API that doesn't depend on image/draw at the call site. We
// pre-rasterize each printable ASCII glyph once at startup into a
// fixed-size mask, then the framebuffer renderer composites those masks
// directly onto its backing buffer.
package font

import (
	"image"
	"image/color"
	"image/draw"

	"golang.org/x/image/font"
	"golang.org/x/image/font/basicfont"
	"golang.org/x/image/math/fixed"
)

const (
	// CellWidth/CellHeight describe one monospace cell. The 7x13 face is
	// padded to 8x16 for ergonomic alignment on the framebuffer.
	CellWidth  = 8
	CellHeight = 16

	firstRune = 0x20 // space
	lastRune  = 0x7E // tilde
	numGlyphs = lastRune - firstRune + 1
)

// glyphMask holds a CellWidth*CellHeight grid of opacity values (0 = empty,
// 0xFF = filled). We don't antialias — the source face is already a 1-bit
// bitmap, so any non-zero pixel is treated as solid.
type glyphMask [CellWidth * CellHeight]byte

var glyphs [numGlyphs]glyphMask

func init() {
	face := basicfont.Face7x13
	tmp := image.NewAlpha(image.Rect(0, 0, CellWidth, CellHeight))
	white := image.NewUniform(color.Alpha{A: 0xFF})

	for r := rune(firstRune); r <= rune(lastRune); r++ {
		// Clear cell.
		draw.Draw(tmp, tmp.Bounds(), image.Transparent, image.Point{}, draw.Src)

		d := &font.Drawer{
			Dst:  tmp,
			Src:  white,
			Face: face,
			// Baseline lands a couple of pixels below the cell midline so
			// descenders ('g', 'p', ...) aren't clipped.
			Dot: fixed.Point26_6{
				X: fixed.I(0),
				Y: fixed.I(face.Ascent + 1),
			},
		}
		d.DrawString(string(r))

		idx := r - firstRune
		for y := 0; y < CellHeight; y++ {
			for x := 0; x < CellWidth; x++ {
				if tmp.AlphaAt(x, y).A != 0 {
					glyphs[idx][y*CellWidth+x] = 0xFF
				}
			}
		}
	}
}

// Glyph returns the rasterized mask for r. Non-printable runes return the
// space-glyph mask (i.e. blank).
func Glyph(r rune) *glyphMask {
	if r < firstRune || r > lastRune {
		r = ' '
	}
	return &glyphs[r-firstRune]
}

// Pixel returns the opacity of a glyph at (x,y). Out-of-range returns 0.
// Hot path; the framebuffer renderer pixel-tests every cell.
func Pixel(r rune, x, y int) byte {
	if x < 0 || x >= CellWidth || y < 0 || y >= CellHeight {
		return 0
	}
	return Glyph(r)[y*CellWidth+x]
}
