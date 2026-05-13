package randomizer

import (
	_ "embed"
	"regexp"
	"strings"
)

//go:embed strings/uncle.txt
var stringsUncle string

//go:embed strings/tavern_man.txt
var stringsTavernMan string

//go:embed strings/blind.txt
var stringsBlind string

//go:embed strings/ganon_1.txt
var stringsGanon1 string

//go:embed strings/ganon_phase_3_no_silvers.txt
var stringsGanonNoSilvers string

//go:embed strings/triforce.txt
var stringsTriforce string

// stripCRLF + leading `-\n` mirrors PHP Randomizer::getTextArray.
var leadingDash = regexp.MustCompile(`^-\n`)

// parseStringPool splits a string pool file on `\n-\n` and filters empty entries.
// Mirrors PHP `Randomizer::getTextArray`.
func parseStringPool(content string) []string {
	content = strings.ReplaceAll(content, "\r\n", "\n")
	content = leadingDash.ReplaceAllString(content, "")
	raw := strings.Split(content, "\n-\n")
	out := make([]string, 0, len(raw))
	for _, s := range raw {
		if strings.TrimSpace(s) == "" {
			continue
		}
		out = append(out, strings.TrimRight(s, "\n"))
	}
	return out
}
