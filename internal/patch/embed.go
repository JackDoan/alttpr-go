package patch

import (
	"bytes"
	_ "embed"
)

//go:embed all_patches_embed/edc01f3db798ae4dfe21101311598d44.json
var basePatchJSON []byte

// LoadEmbedded parses the base patch JSON that was compiled into the binary.
// Equivalent to Load(bytes.NewReader(...)) but spares callers the import.
func LoadEmbedded() ([]Entry, error) {
	return Load(bytes.NewReader(basePatchJSON))
}

// EmbeddedRaw returns the raw bytes for diagnostics; do not mutate.
func EmbeddedRaw() []byte { return basePatchJSON }
