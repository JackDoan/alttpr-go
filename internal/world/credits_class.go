package world

import (
	"strings"

	"github.com/JackDoan/alttpr-go/internal/rom"
)

// CreditPart is one line of a credit scene: type (small/small_alt/large),
// x/y position, and text. Mirrors PHP Credits scene records.
type CreditPart struct {
	Type string // "small" | "small_alt" | "large"
	X    int
	Y    int
	Text string
}

// Credits manages the credit-sequence text table written to ROM 0x181500 +
// a 17-entry pointer table at 0x76CC0. Mirrors PHP Support\Credits.
type Credits struct {
	order  []string
	scenes map[string][]CreditPart
}

// NewCredits returns the default credit scenes (matches PHP `$scenes`).
func NewCredits() *Credits {
	c := &Credits{
		order:  []string{"castle", "sanctuary", "kakariko", "desert", "hera", "house", "zora", "witch", "lumberjacks", "grove", "well", "smithy", "kakariko2", "bridge", "woods", "pedestal"},
		scenes: map[string][]CreditPart{},
	}
	add := func(name string, parts ...CreditPart) { c.scenes[name] = parts }

	add("castle",
		CreditPart{"small", 5, 19, "The return of the King"},
		CreditPart{"large", 9, 23, "Hyrule Castle"})
	add("sanctuary",
		CreditPart{"small", 8, 19, "The loyal priest"},
		CreditPart{"large", 11, 23, "Sanctuary"})
	add("kakariko",
		CreditPart{"small", 4, 19, "Sahasralah's Homecoming"},
		CreditPart{"large", 9, 23, "Kakariko Town"})
	add("desert",
		CreditPart{"small", 4, 19, "vultures rule the desert"},
		CreditPart{"large", 9, 23, "Desert Palace"})
	add("hera",
		CreditPart{"small", 4, 19, "the bully makes a friend"},
		CreditPart{"large", 9, 23, "Mountain Tower"})
	add("house",
		CreditPart{"small", 6, 19, "your uncle recovers"},
		CreditPart{"large", 11, 23, "Your House"})
	add("zora",
		CreditPart{"small", 6, 19, "finger webs for sale"},
		CreditPart{"large", 8, 23, "Zora's Waterfall"})
	add("witch",
		CreditPart{"small", 4, 19, "the witch and assistant"},
		CreditPart{"large", 11, 23, "Magic Shop"})
	add("lumberjacks",
		CreditPart{"small", 8, 19, "twin lumberjacks"},
		CreditPart{"large", 9, 23, "Woodsmen's Hut"})
	add("grove",
		CreditPart{"small", 4, 19, "ocarina boy plays again"},
		CreditPart{"large", 9, 23, "Haunted Grove"})
	add("well",
		CreditPart{"small", 4, 19, "venus, queen of faeries"},
		CreditPart{"large", 10, 23, "Wishing Well"})
	add("smithy",
		CreditPart{"small", 4, 19, "the dwarven swordsmiths"},
		CreditPart{"large", 12, 23, "Smithery"})
	add("kakariko2",
		CreditPart{"small", 6, 19, "the bug-catching kid"},
		CreditPart{"large", 9, 23, "Kakariko Town"})
	add("bridge",
		CreditPart{"small", 8, 19, "the lost old man"},
		CreditPart{"large", 9, 23, "Death Mountain"})
	add("woods",
		CreditPart{"small", 8, 19, "the forest thief"},
		CreditPart{"large", 11, 23, "Lost Woods"})
	add("pedestal",
		CreditPart{"small", 6, 19, "and the master sword"},
		CreditPart{"small_alt", 8, 21, "sleeps again..."},
		CreditPart{"large", 12, 23, "Forever!"})
	return c
}

// UpdateCreditLine replaces a credit line for a scene + line index, with the
// given alignment ("center"/"left"/"right"). PHP truncates to 32 chars.
// Mirrors PHP Credits::updateCreditLine.
func (c *Credits) UpdateCreditLine(scene string, line int, text, align string) bool {
	parts, ok := c.scenes[scene]
	if !ok || line < 0 || line >= len(parts) {
		return false
	}
	if len(text) > 32 {
		text = text[:32]
	}
	parts[line].Text = text
	switch align {
	case "left":
		parts[line].X = 0
	case "right":
		parts[line].X = max(0, 32-len(text))
	default:
		parts[line].X = max(0, (32-len(text))/2)
	}
	c.scenes[scene] = parts
	return true
}

// charToCreditsHex maps a char into the small-credits font. Mirrors PHP.
func charToCreditsHex(ch byte) byte {
	c := ch
	if c >= 'A' && c <= 'Z' {
		c = c - 'A' + 'a' // strtolower
	}
	if c >= 'a' && c <= 'z' {
		return c - 0x47
	}
	switch c {
	case ' ':
		return 0x9F
	case ',':
		return 0x34
	case '.':
		return 0x37
	case '-':
		return 0x36
	case '\'':
		return 0x35
	}
	return 0x9F
}

func charToAltCreditsHex(ch byte) byte {
	c := ch
	if c >= 'A' && c <= 'Z' {
		c = c - 'A' + 'a'
	}
	if c >= 'a' && c <= 'z' {
		return c - 0x29
	}
	if c == '.' {
		return 0x52
	}
	return 0x9F
}

func convertCredits(s string) []byte {
	out := make([]byte, len(s))
	for i := range s {
		out[i] = charToCreditsHex(s[i])
	}
	return out
}

func convertAltCredits(s string) []byte {
	out := make([]byte, len(s))
	for i := range s {
		out[i] = charToAltCreditsHex(s[i])
	}
	return out
}

func convertLargeCreditsTop(s string) []byte {
	out := make([]byte, len(s))
	for i := range s {
		c := s[i]
		if c >= 'A' && c <= 'Z' {
			c = c - 'A' + 'a'
		}
		if c >= 'a' && c <= 'z' {
			out[i] = c - 0x4
			continue
		}
		switch c {
		case '\'':
			out[i] = 0xD9
		case '!':
			out[i] = 0xE5
		case '_':
			out[i] = 0xDE
		default:
			out[i] = 0x9F
		}
	}
	return out
}

func convertLargeCreditsBottom(s string) []byte {
	out := make([]byte, len(s))
	for i := range s {
		c := s[i]
		if c >= 'A' && c <= 'Z' {
			c = c - 'A' + 'a'
		}
		if c >= 'a' && c <= 'z' {
			out[i] = c + 0x22
			continue
		}
		switch c {
		case '\'':
			out[i] = 0xEC
		case '!':
			out[i] = 0xF8
		case '_':
			out[i] = 0xF1
		default:
			out[i] = 0x9F
		}
	}
	return out
}

// header builds the 4-byte header per credit text run.
// PHP: pack('N', (0x6000 | (y>>5<<11) | ((y&0x1F)<<5) | (x>>5<<10) | (x&0x1F)) << 16 | (length*2 - 1))
// Returns 4 bytes big-endian.
func creditsHeader(x, y, length int) []byte {
	hi := uint32(0x6000) |
		(uint32(y>>5) << 11) | (uint32(y&0x1F) << 5) |
		(uint32(x>>5) << 10) | uint32(x&0x1F)
	val := (hi << 16) | uint32(length*2-1)
	// PHP pack('N',...) is big-endian.
	return []byte{byte(val >> 24), byte(val >> 16), byte(val >> 8), byte(val)}
}

// smallConverted emits the small-text block for one record, including
// the special apostrophe + comma re-encoding passes from PHP.
func smallConverted(p CreditPart) []byte {
	out := []byte{}
	conv := convertCredits(p.Text)
	out = append(out, creditsHeader(p.X, p.Y, len(conv))...)
	out = append(out, conv...)

	// apostrophes: replace ' → , at original positions, blank others.
	apos := strings.Map(func(r rune) rune {
		if r == '\'' {
			return ','
		}
		return ' '
	}, p.Text)
	if trimmed := strings.TrimSpace(apos); trimmed != "" {
		conv = convertCredits(trimmed)
		idx := strings.Index(apos, ",")
		out = append(out, creditsHeader(p.X+idx, p.Y-1, len(conv))...)
		out = append(out, conv...)
	}

	// commas: replace , → ' at original positions, blank others.
	comm := strings.Map(func(r rune) rune {
		if r == ',' {
			return '\''
		}
		return ' '
	}, p.Text)
	if trimmed := strings.TrimSpace(comm); trimmed != "" {
		conv = convertCredits(trimmed)
		idx := strings.Index(comm, "'")
		out = append(out, creditsHeader(p.X+idx, p.Y+1, len(conv))...)
		out = append(out, conv...)
	}
	return out
}

func smallAltConverted(p CreditPart) []byte {
	conv := convertAltCredits(p.Text)
	out := append([]byte(nil), creditsHeader(p.X, p.Y, len(conv))...)
	return append(out, conv...)
}

func largeConverted(p CreditPart) []byte {
	out := []byte{}
	top := convertLargeCreditsTop(p.Text)
	out = append(out, creditsHeader(p.X, p.Y, len(top))...)
	out = append(out, top...)
	bot := convertLargeCreditsBottom(p.Text)
	out = append(out, creditsHeader(p.X, p.Y+1, len(bot))...)
	out = append(out, bot...)
	return out
}

// BinaryData returns the credits data + pointer table. Mirrors PHP getBinaryData.
// `pointers` contains N+1 16-bit positions: one per scene (start offset),
// plus a final "end" pointer.
func (c *Credits) BinaryData() (data []byte, pointers []uint16) {
	for _, scene := range c.order {
		pointers = append(pointers, uint16(len(data)))
		for _, p := range c.scenes[scene] {
			switch p.Type {
			case "small":
				data = append(data, smallConverted(p)...)
			case "small_alt":
				data = append(data, smallAltConverted(p)...)
			case "large":
				data = append(data, largeConverted(p)...)
			}
		}
	}
	pointers = append(pointers, uint16(len(data)))
	return data, pointers
}

// WriteTo writes credits data at 0x181500 + pointer table at 0x76CC0.
// Mirrors PHP Rom::writeCredits.
func (c *Credits) WriteTo(r *rom.ROM) error {
	data, ptrs := c.BinaryData()
	if err := r.Write(0x181500, data, true); err != nil {
		return err
	}
	pbuf := make([]byte, len(ptrs)*2)
	for i, p := range ptrs {
		pbuf[i*2] = byte(p)
		pbuf[i*2+1] = byte(p >> 8)
	}
	return r.Write(0x76CC0, pbuf, true)
}
