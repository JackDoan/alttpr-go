// Package fb is a tiny pure-Go framebuffer renderer suitable for the
// Trimui Brick (linux/arm64) running stock firmware. It reads display
// metrics from /sys/class/graphics/fb0/* (avoiding any ioctl plumbing),
// mmaps /dev/fb0, and exposes blit primitives the UI layer uses.
//
// Pixel format: 32 bpp BGRA (the format stock Trimui OS exposes). 16 bpp
// support is not implemented; if Open detects an unexpected depth it
// returns an error so the user can investigate rather than getting silently
// scrambled output.
package fb

import (
	"fmt"
	"os"
	"strconv"
	"strings"
	"syscall"

	"github.com/JackDoan/alttpr-go/cmd/alttpr-brick/font"
)

// Default sysfs/dev paths. Overridable via env for non-default fbcons.
const (
	defaultDev      = "/dev/fb0"
	defaultSysClass = "/sys/class/graphics/fb0"
)

// Color is a 24-bit RGB tuple. Alpha is implicit (always opaque).
type Color struct{ R, G, B uint8 }

// Common palette.
var (
	Black = Color{0x10, 0x10, 0x10}
	White = Color{0xF0, 0xF0, 0xF0}
	Gray  = Color{0x60, 0x60, 0x60}
	Blue  = Color{0x40, 0x70, 0xC0}
	Green = Color{0x60, 0xB0, 0x60}
	Red   = Color{0xC0, 0x50, 0x50}
)

// FB is a double-buffered mmap-backed framebuffer. All drawing writes to
// an in-memory back buffer; Present() copies the whole buffer to the
// mmap'd /dev/fb0 in a single pass, so the user never sees a partial frame.
type FB struct {
	fd     int
	mem    []byte // mmap'd /dev/fb0
	buf    []byte // back buffer, same shape as mem
	W, H   int    // visible dimensions in pixels
	BPP    int    // bits per pixel (must be 32)
	Stride int    // bytes per row
}

// Open the device, query metrics, mmap, and ready it for drawing.
func Open() (*FB, error) {
	dev := envOr("TRIMUI_FB", defaultDev)
	sysDir := envOr("TRIMUI_FB_SYS", defaultSysClass)

	w, h, err := readVirtualSize(sysDir + "/virtual_size")
	if err != nil {
		return nil, fmt.Errorf("read virtual_size: %w", err)
	}
	bpp, err := readSingleInt(sysDir + "/bits_per_pixel")
	if err != nil {
		return nil, fmt.Errorf("read bits_per_pixel: %w", err)
	}
	if bpp != 32 {
		return nil, fmt.Errorf("unsupported framebuffer depth: %d bpp (expected 32)", bpp)
	}
	stride, err := readSingleInt(sysDir + "/stride")
	if err != nil {
		// Fall back to assuming packed: 4 bytes/pixel * width.
		stride = w * 4
	}

	fd, err := syscall.Open(dev, syscall.O_RDWR, 0)
	if err != nil {
		return nil, fmt.Errorf("open %s: %w", dev, err)
	}
	mem, err := syscall.Mmap(fd, 0, stride*h,
		syscall.PROT_READ|syscall.PROT_WRITE, syscall.MAP_SHARED)
	if err != nil {
		syscall.Close(fd)
		return nil, fmt.Errorf("mmap %s: %w", dev, err)
	}
	return &FB{
		fd:     fd,
		mem:    mem,
		buf:    make([]byte, stride*h),
		W:      w,
		H:      h,
		BPP:    bpp,
		Stride: stride,
	}, nil
}

func (f *FB) Close() error {
	if f.mem != nil {
		_ = syscall.Munmap(f.mem)
		f.mem = nil
	}
	if f.fd != 0 {
		err := syscall.Close(f.fd)
		f.fd = 0
		return err
	}
	return nil
}

func (f *FB) Bounds() (int, int) { return f.W, f.H }

// FillRect fills the (x,y,w,h) rectangle (clipped to screen) with color.
func (f *FB) FillRect(x, y, w, h int, c Color) {
	x0, y0 := clamp(x, 0, f.W), clamp(y, 0, f.H)
	x1, y1 := clamp(x+w, 0, f.W), clamp(y+h, 0, f.H)
	if x1 <= x0 || y1 <= y0 {
		return
	}
	// Build a single-pixel template and replicate horizontally per-row.
	bgra := [4]byte{c.B, c.G, c.R, 0xFF}
	for py := y0; py < y1; py++ {
		off := py*f.Stride + x0*4
		for px := x0; px < x1; px++ {
			copy(f.buf[off:off+4], bgra[:])
			off += 4
		}
	}
}

// Clear fills the entire framebuffer with c.
func (f *FB) Clear(c Color) { f.FillRect(0, 0, f.W, f.H, c) }

// DrawText renders text starting at (x,y) (top-left of the first glyph)
// in fg. Text wraps... not at all — callers are responsible for layout.
func (f *FB) DrawText(x, y int, text string, fg Color) {
	bgra := [4]byte{fg.B, fg.G, fg.R, 0xFF}
	cx := x
	for _, r := range text {
		for py := 0; py < font.CellHeight; py++ {
			yy := y + py
			if yy < 0 || yy >= f.H {
				continue
			}
			rowBase := yy * f.Stride
			for px := 0; px < font.CellWidth; px++ {
				if font.Pixel(r, px, py) == 0 {
					continue
				}
				xx := cx + px
				if xx < 0 || xx >= f.W {
					continue
				}
				off := rowBase + xx*4
				copy(f.buf[off:off+4], bgra[:])
			}
		}
		cx += font.CellWidth
	}
}

// TextWidth returns the pixel width a string will occupy when drawn.
func TextWidth(s string) int { return len(s) * font.CellWidth }

// Present copies the back buffer to the mmap'd framebuffer in a single
// pass. Call this once per frame after all draw calls; the user only
// ever sees fully-rendered frames, never a half-cleared screen.
func (f *FB) Present() { copy(f.mem, f.buf) }

// --- helpers -------------------------------------------------------------

func envOr(k, def string) string {
	if v := os.Getenv(k); v != "" {
		return v
	}
	return def
}

func readVirtualSize(path string) (int, int, error) {
	b, err := os.ReadFile(path)
	if err != nil {
		return 0, 0, err
	}
	s := strings.TrimSpace(string(b))
	parts := strings.Split(s, ",")
	if len(parts) != 2 {
		return 0, 0, fmt.Errorf("unexpected virtual_size %q", s)
	}
	w, err := strconv.Atoi(strings.TrimSpace(parts[0]))
	if err != nil {
		return 0, 0, err
	}
	h, err := strconv.Atoi(strings.TrimSpace(parts[1]))
	if err != nil {
		return 0, 0, err
	}
	return w, h, nil
}

func readSingleInt(path string) (int, error) {
	b, err := os.ReadFile(path)
	if err != nil {
		return 0, err
	}
	return strconv.Atoi(strings.TrimSpace(string(b)))
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
