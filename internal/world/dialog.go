package world

import "strings"

// dialogChars maps multi-byte ALttP-special characters (Hylian glyphs, hearts,
// face icons, etc.) to their byte encodings. Mirrors PHP Dialog::$characters
// for the non-Japanese/non-alphanumeric entries — those are computed.
var dialogChars = map[string][]byte{
	" ": {0xFF}, "…": {0x9F},
	"?": {0xC6}, "!": {0xC7}, ",": {0xC8}, "-": {0xC9},
	".": {0xCD}, "~": {0xCE}, ":": {0xEA},
	"'": {0x9D}, "≥": {0x99},
	"@": {0xFE, 0x6A},
	">": {0x9B, 0x9C},
	"%": {0xFD, 0x10}, "^": {0xFD, 0x11}, "=": {0xFD, 0x12},
	"¼": {0xE5, 0xE7}, "½": {0xE6, 0xE7},
	"¾": {0xE8, 0xE9}, "♥": {0xEA, 0xEB},
}

// charToHex returns the encoded bytes for a single rune.
// Mirrors PHP Dialog::charToHex (digits, A-Z, a-z, fallback table).
func charToHex(r rune) []byte {
	switch {
	case r >= '0' && r <= '9':
		return []byte{byte(int(r-'0') + 0xA0)}
	case r >= 'A' && r <= 'Z':
		return []byte{byte(int(r-'A') + 0xAA)}
	case r >= 'a' && r <= 'z':
		return []byte{byte(int(r) + 0x6F)}
	}
	if b, ok := dialogChars[string(r)]; ok {
		return b
	}
	return []byte{0xFF} // unknown -> space
}

// ConvertDialogCompressed encodes `s` for compressed dialog. Used by Text.setString.
// Mirrors PHP Dialog::convertDialogCompressed (single-pass, no {COMMAND} tokens).
// We support a minimal set: line breaks (\n -> row markers), wrap at 19, terminator 0x7F.
// Full PHP command grammar ({SPEED0}, {INTRO}, etc.) is deferred.
func ConvertDialogCompressed(s string, pause bool, maxBytes, wrap int) []byte {
	if maxBytes <= 0 {
		maxBytes = 2046
	}
	if wrap <= 0 {
		wrap = 19
	}
	// Prefix: 0xFB starts compressed dialog.
	out := []byte{0xFB}
	lines := strings.Split(s, "\n")
	i := 0
	for _, line := range lines {
		// Word-wrap individual lines to `wrap` columns.
		wrapped := wrapLine(line, wrap)
		for _, w := range wrapped {
			switch i {
			case 0:
				// First row: no leading marker.
			case 1:
				out = append(out, 0xF8) // row 2
			case 2:
				out = append(out, 0xF9) // row 3
			default:
				if i >= 3 {
					out = append(out, 0xF6) // scroll
				}
			}
			// PHP also caps each line at `wrap` chars via mb_substr after
			// wordwrap (handles the no-space, over-wrap case).
			capped := w
			if len(capped) > wrap {
				capped = capped[:wrap]
			}
			for _, r := range capped {
				out = append(out, charToHex(r)...)
			}
			i++
			if pause && i%3 == 0 && i < linesTotal(lines, wrap) {
				out = append(out, 0xFA)
			}
		}
	}
	// PHP `convertDialogCompressed` does NOT append a terminator; the
	// caller relies on length limits / next-string offsets.
	if len(out) > maxBytes {
		out = out[:maxBytes]
	}
	return out
}

// linesTotal returns the total wrapped-line count across `lines`.
func linesTotal(lines []string, wrap int) int {
	n := 0
	for _, l := range lines {
		n += len(wrapLine(l, wrap))
	}
	return n
}

// wrapLine splits `s` into chunks of at most `wrap` characters at whitespace.
// Falls back to hard-wrap for non-whitespace overflows.
func wrapLine(s string, wrap int) []string {
	if len(s) == 0 {
		return []string{""}
	}
	words := strings.Fields(s)
	if len(words) == 0 {
		return []string{""}
	}
	out := []string{}
	current := ""
	for _, w := range words {
		if current == "" {
			current = w
		} else if len(current)+1+len(w) <= wrap {
			current += " " + w
		} else {
			out = append(out, current)
			current = w
		}
	}
	if current != "" {
		out = append(out, current)
	}
	return out
}
