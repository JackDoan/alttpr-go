package patch

import (
	"encoding/json"
	"fmt"
	"io"
	"os"
	"strconv"

	"github.com/JackDoan/alttpr-go/internal/rom"
)

// Entry is one address->bytes write. The patch is an ordered list of these.
type Entry struct {
	Offset int
	Data   []byte
}

// LoadFile reads a JSON patch from disk in the PHP format:
//
//	[{"<addr>": [b0, b1, ...]}, {"<addr>": [...]}, ...]
//
// Each element has exactly one key; order is preserved.
// Mirrors app/Console/Commands/UpdateBuildRecord.php:148-204 output shape.
func LoadFile(path string) ([]Entry, error) {
	f, err := os.Open(path)
	if err != nil {
		return nil, fmt.Errorf("open patch %q: %w", path, err)
	}
	defer f.Close()
	return Load(f)
}

func Load(r io.Reader) ([]Entry, error) {
	dec := json.NewDecoder(r)
	dec.UseNumber()

	// Expect opening '['.
	tok, err := dec.Token()
	if err != nil {
		return nil, fmt.Errorf("read opening token: %w", err)
	}
	if delim, ok := tok.(json.Delim); !ok || delim != '[' {
		return nil, fmt.Errorf("expected array, got %v", tok)
	}

	var entries []Entry
	for dec.More() {
		// Each element is an object with exactly one entry.
		tok, err := dec.Token()
		if err != nil {
			return nil, fmt.Errorf("read element open: %w", err)
		}
		delim, ok := tok.(json.Delim)
		if !ok || delim != '{' {
			return nil, fmt.Errorf("expected object, got %v", tok)
		}

		// Address key.
		tok, err = dec.Token()
		if err != nil {
			return nil, fmt.Errorf("read address key: %w", err)
		}
		keyStr, ok := tok.(string)
		if !ok {
			return nil, fmt.Errorf("expected string key, got %v", tok)
		}
		addr, err := strconv.Atoi(keyStr)
		if err != nil {
			return nil, fmt.Errorf("address key %q is not an int: %w", keyStr, err)
		}

		// Byte array.
		var rawBytes []json.Number
		if err := dec.Decode(&rawBytes); err != nil {
			return nil, fmt.Errorf("decode bytes for addr %d: %w", addr, err)
		}
		data := make([]byte, len(rawBytes))
		for i, n := range rawBytes {
			v, err := n.Int64()
			if err != nil {
				return nil, fmt.Errorf("byte at addr %d index %d: %w", addr, i, err)
			}
			if v < 0 || v > 255 {
				return nil, fmt.Errorf("byte at addr %d index %d out of range: %d", addr, i, v)
			}
			data[i] = byte(v)
		}

		// Closing '}'.
		tok, err = dec.Token()
		if err != nil {
			return nil, fmt.Errorf("read element close: %w", err)
		}
		if delim, ok := tok.(json.Delim); !ok || delim != '}' {
			return nil, fmt.Errorf("expected object close, got %v", tok)
		}

		entries = append(entries, Entry{Offset: addr, Data: data})
	}

	// Closing ']'.
	if _, err := dec.Token(); err != nil {
		return nil, fmt.Errorf("read closing token: %w", err)
	}

	return entries, nil
}

// Apply writes each entry to the ROM in order. Mirrors app/Rom.php:2446-2455.
// The base patch is applied silently (no write log), matching PHP behavior
// where `applyPatch` does not pass the log flag (PHP defaults to true, but
// the base patch comes through `applyPatchFile` which also uses the default;
// for fidelity we follow whatever the user wants via the `log` parameter).
func Apply(r *rom.ROM, entries []Entry, log bool) error {
	for _, e := range entries {
		if err := r.Write(e.Offset, e.Data, log); err != nil {
			return fmt.Errorf("apply at 0x%X: %w", e.Offset, err)
		}
	}
	return nil
}
