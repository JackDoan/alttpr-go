package world

import (
	_ "embed"
	"encoding/json"
	"fmt"

	"github.com/JackDoan/alttpr-go/internal/rom"
)

// defaultTextENJSON is the ordered English text dictionary dumped from
// PHP's `Text->removeUnwanted()` state via reflection. Each entry is
// `{"n": name, "b": [bytes...]}`. We preserve order so `getByteArray`
// emits identical concatenation to PHP.
//
//go:embed default_text_en.json
var defaultTextENJSON []byte

type textEntry struct {
	Name  string `json:"n"`
	Bytes []int  `json:"b"`
}

// Text is the per-world text table mirroring PHP `ALttP\Text`.
// Strings are addressed by name and emitted as a concatenated byte array
// at ROM 0xE0000.
type Text struct {
	order []string
	bytes map[string][]byte
}

// NewText loads the embedded default English table.
func NewText() *Text {
	var entries []textEntry
	if err := json.Unmarshal(defaultTextENJSON, &entries); err != nil {
		// Embedded data is build-time validated; panic if corrupted.
		panic(fmt.Sprintf("Text: cannot decode embedded dictionary: %v", err))
	}
	t := &Text{
		order: make([]string, 0, len(entries)),
		bytes: make(map[string][]byte, len(entries)),
	}
	for _, e := range entries {
		buf := make([]byte, len(e.Bytes))
		for i, v := range e.Bytes {
			buf[i] = byte(v)
		}
		t.order = append(t.order, e.Name)
		t.bytes[e.Name] = buf
	}
	return t
}

// SetString updates the bytes for `name` using the Dialog encoder.
// Mirrors PHP Text::setString. Pause/maxBytes/wrap defaults match PHP.
func (t *Text) SetString(name, value string) error {
	if _, ok := t.bytes[name]; !ok {
		return fmt.Errorf("text key %q does not exist", name)
	}
	t.bytes[name] = ConvertDialogCompressed(value, true, 2046, 19)
	return nil
}

// SetStringRaw updates the bytes for `name` directly (no encoding).
// Mirrors PHP Text::setStringRaw.
func (t *Text) SetStringRaw(name string, raw []byte) error {
	if _, ok := t.bytes[name]; !ok {
		return fmt.Errorf("text key %q does not exist", name)
	}
	t.bytes[name] = append([]byte(nil), raw...)
	return nil
}

// GetByteArray concatenates all values in original order, padding to
// 0x7355 with 0xFF when pad is true. Mirrors PHP Text::getByteArray.
func (t *Text) GetByteArray(pad bool) ([]byte, error) {
	out := make([]byte, 0, 0x7355)
	for _, name := range t.order {
		out = append(out, t.bytes[name]...)
	}
	if len(out) > 0x7355 {
		return nil, fmt.Errorf("text overflow: %d > 0x7355", len(out))
	}
	if pad {
		for len(out) < 0x7355 {
			out = append(out, 0xFF)
		}
	}
	return out, nil
}

// WriteTo writes the text region to ROM at 0xE0000. PHP Rom::writeText.
func (t *Text) WriteTo(r *rom.ROM) error {
	buf, err := t.GetByteArray(true)
	if err != nil {
		return err
	}
	return r.Write(0xE0000, buf, true)
}

// WriteDefaultText writes the default English text region (legacy entry point).
func WriteDefaultText(r *rom.ROM) error {
	return NewText().WriteTo(r)
}
