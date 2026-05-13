// Package input reads gamepad button presses from /dev/input/event*
// devices and translates evdev keycodes into a small high-level Button
// enum that the UI consumes. Pure Go (syscall.Read on a non-blocking fd
// via goroutines), no CGO.
package input

import (
	"encoding/binary"
	"fmt"
	"io"
	"os"
	"path/filepath"
	"strings"
	"syscall"
	"time"
)

// Button is the abstract D-pad/button the UI deals in.
type Button int

const (
	BtnNone Button = iota
	BtnUp
	BtnDown
	BtnLeft
	BtnRight
	BtnA      // confirm
	BtnB      // back
	BtnStart  // open menu / Generate shortcut
	BtnSelect // toggle screen
	BtnMenu   // exit / power
	BtnL
	BtnR
)

func (b Button) String() string {
	switch b {
	case BtnUp:
		return "Up"
	case BtnDown:
		return "Down"
	case BtnLeft:
		return "Left"
	case BtnRight:
		return "Right"
	case BtnA:
		return "A"
	case BtnB:
		return "B"
	case BtnStart:
		return "Start"
	case BtnSelect:
		return "Select"
	case BtnMenu:
		return "Menu"
	case BtnL:
		return "L"
	case BtnR:
		return "R"
	}
	return "(none)"
}

// Event is a single button press (we only surface presses, not releases).
type Event struct {
	Button  Button
	RawCode uint16
}

// Linux input_event (24 bytes on 64-bit, 16 bytes on 32-bit, but 24 on
// modern aarch64 with the legacy timeval format). We use the aarch64
// layout: struct { struct timeval time; __u16 type; __u16 code; __s32 value; }
// where timeval is 16 bytes (sec int64 + usec int64). Total 24 bytes.
const eventSize = 24

// evdev type codes (linux/input-event-codes.h).
const (
	evSyn = 0x00
	evKey = 0x01
	evAbs = 0x03
)

// Absolute-axis codes for D-pad hats (linux/input-event-codes.h).
// Many gamepads report the D-pad on EV_ABS rather than EV_KEY. Each axis
// reports -1/0/+1; we synthesize directional button presses on the edge
// where the axis becomes non-zero.
const (
	absHat0X = 0x10 // left/right hat
	absHat0Y = 0x11 // up/down hat
)

// Standard evdev key codes used by the Trimui Brick's gamepad. These are
// the *default* mappings; the on-device button layout may differ. When in
// doubt, set INPUT_DEBUG=1 and look at the logged raw codes.
const (
	keyUp    = 103
	keyDown  = 108
	keyLeft  = 105
	keyRight = 106

	// Gamepad buttons. On the Trimui Brick the physical labels are wired
	// up differently from the BTN_SOUTH/EAST defaults — measured by
	// dumping INPUT_DEBUG output, the codes are: A=305, B=304, X=308, Y=307.
	btnA     = 305 // physical "A" (confirm)
	btnB     = 304 // physical "B" (back)
	btnX     = 308 // physical "X"
	btnY     = 307 // physical "Y"
	btnTL    = 310 // BTN_TL    (left shoulder)
	btnTR    = 311 // BTN_TR    (right shoulder)
	btnSel   = 314 // BTN_SELECT
	btnStart = 315 // BTN_START
	btnMode  = 316 // BTN_MODE  (typically home/menu)

	// Some Trimui builds expose Enter/Escape on keyboard codes.
	keyEnter = 28  // KEY_ENTER
	keyEsc   = 1   // KEY_ESC
	keyPower = 116 // KEY_POWER
)

// Reader fans events from every /dev/input/event* device that has a
// readable open() into a single channel.
type Reader struct {
	Events chan Event
	debug  *os.File
	closes []io.Closer

	// Last seen value per (device-pointer, hat-axis) so we can synthesize
	// edge events only when an axis transitions in/out of a non-zero state.
	// Keyed by *os.File pointer because some firmwares present a separate
	// hat device from the rest of the gamepad.
	hatState map[*os.File]hatAxes
}

type hatAxes struct{ x, y int32 }

// Open scans /dev/input/event* (overridable via TRIMUI_INPUT_GLOB) and
// starts one goroutine per device. The returned channel is buffered; the
// UI should drain it promptly to avoid losing fast input.
func Open() (*Reader, error) {
	pattern := envOr("TRIMUI_INPUT_GLOB", "/dev/input/event*")
	paths, err := filepath.Glob(pattern)
	if err != nil {
		return nil, fmt.Errorf("glob %s: %w", pattern, err)
	}
	if len(paths) == 0 {
		return nil, fmt.Errorf("no input devices match %s", pattern)
	}

	r := &Reader{
		Events:   make(chan Event, 32),
		hatState: make(map[*os.File]hatAxes),
	}

	if os.Getenv("INPUT_DEBUG") != "" {
		f, err := os.OpenFile("input-debug.log",
			os.O_CREATE|os.O_WRONLY|os.O_APPEND, 0o644)
		if err == nil {
			r.debug = f
			fmt.Fprintf(f, "\n--- input session %s ---\n", time.Now().Format(time.RFC3339))
		}
	}

	any := false
	for _, p := range paths {
		fd, err := syscall.Open(p, syscall.O_RDONLY|syscall.O_NONBLOCK, 0)
		if err != nil {
			continue
		}
		f := os.NewFile(uintptr(fd), p)
		r.closes = append(r.closes, f)
		go r.readDevice(f)
		any = true
	}
	if !any {
		return nil, fmt.Errorf("no input devices openable under %s", pattern)
	}
	return r, nil
}

func (r *Reader) Close() error {
	for _, c := range r.closes {
		_ = c.Close()
	}
	if r.debug != nil {
		_ = r.debug.Close()
	}
	return nil
}

// readDevice blocks (sort of — we use blocking reads via fd-as-file, the
// kernel buffers and wakes us on each event) and emits Events.
func (r *Reader) readDevice(f *os.File) {
	defer f.Close()
	buf := make([]byte, eventSize*16) // batch
	for {
		n, err := f.Read(buf)
		if err != nil {
			if err == io.EOF || strings.Contains(err.Error(), "file already closed") {
				return
			}
			// EAGAIN under non-blocking IO; brief backoff and retry.
			time.Sleep(5 * time.Millisecond)
			continue
		}
		for off := 0; off+eventSize <= n; off += eventSize {
			r.parse(f, buf[off:off+eventSize])
		}
	}
}

func (r *Reader) parse(dev *os.File, raw []byte) {
	// time (16 bytes), type u16, code u16, value s32
	evType := binary.LittleEndian.Uint16(raw[16:18])
	code := binary.LittleEndian.Uint16(raw[18:20])
	value := int32(binary.LittleEndian.Uint32(raw[20:24]))

	if evType == evSyn {
		return
	}

	if r.debug != nil {
		fmt.Fprintf(r.debug, "type=%d code=%d value=%d dev=%s\n", evType, code, value, dev.Name())
	}

	if evType == evAbs {
		r.handleAbs(dev, code, value)
		return
	}
	if evType != evKey {
		return
	}
	// value: 0 = release, 1 = press, 2 = repeat. Surface presses + repeats.
	if value == 0 {
		return
	}

	btn := decodeButton(code)
	if btn == BtnNone {
		return
	}
	select {
	case r.Events <- Event{Button: btn, RawCode: code}:
	default:
		// channel full; drop. Avoids head-of-line blocking the kernel.
	}
}

// handleAbs synthesizes BtnUp/Down/Left/Right from an EV_ABS hat axis.
// We only emit on the transition into a non-zero value (the press edge);
// the release (axis → 0) is not surfaced, matching how the menu handles
// EV_KEY events. Repeats are not synthesized either — the user can press
// again. Non-hat axes (sticks, triggers) are ignored.
func (r *Reader) handleAbs(dev *os.File, code uint16, value int32) {
	if code != absHat0X && code != absHat0Y {
		return
	}
	prev := r.hatState[dev]
	var btn Button
	switch code {
	case absHat0X:
		if value != 0 && prev.x == 0 {
			if value < 0 {
				btn = BtnLeft
			} else {
				btn = BtnRight
			}
		}
		prev.x = value
	case absHat0Y:
		if value != 0 && prev.y == 0 {
			if value < 0 {
				btn = BtnUp
			} else {
				btn = BtnDown
			}
		}
		prev.y = value
	}
	r.hatState[dev] = prev
	if btn == BtnNone {
		return
	}
	select {
	case r.Events <- Event{Button: btn, RawCode: code}:
	default:
	}
}

func decodeButton(code uint16) Button {
	switch int(code) {
	case keyUp:
		return BtnUp
	case keyDown:
		return BtnDown
	case keyLeft:
		return BtnLeft
	case keyRight:
		return BtnRight
	case btnA, keyEnter:
		return BtnA
	case btnB, keyEsc:
		return BtnB
	case btnStart:
		return BtnStart
	case btnSel:
		return BtnSelect
	case btnMode, keyPower:
		return BtnMenu
	case btnTL:
		return BtnL
	case btnTR:
		return BtnR
	case btnX:
		// Often used as a secondary confirm or a "details" key. Map to Select
		// for now so it's at least reachable.
		return BtnSelect
	case btnY:
		return BtnB
	}
	return BtnNone
}

func envOr(k, def string) string {
	if v := os.Getenv(k); v != "" {
		return v
	}
	return def
}
