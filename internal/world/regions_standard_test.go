package world

import (
	"testing"

	"github.com/JackDoan/alttpr-go/internal/boss"
	"github.com/JackDoan/alttpr-go/internal/item"
)

// All ported Standard-mode region constructors.
var standardRegions = []struct {
	name string
	fn   func(*regionBuilder) *Region
}{
	{"Fountains", newStandardFountains},
	{"Medallions", newStandardMedallions},
	{"Hyrule Castle Tower", newStandardHyruleCastleTower},
	{"Escape", newStandardHyruleCastleEscape},
	{"Eastern Palace", newStandardEasternPalace},
	{"Desert Palace", newStandardDesertPalace},
	{"Tower of Hera", newStandardTowerOfHera},
	{"Ice Palace", newStandardIcePalace},
	{"Thieves Town", newStandardThievesTown},
	{"Skull Woods", newStandardSkullWoods},
	{"Swamp Palace", newStandardSwampPalace},
	{"Misery Mire", newStandardMiseryMire},
	{"Palace of Darkness", newStandardPalaceOfDarkness},
	{"Turtle Rock", newStandardTurtleRock},
	{"Ganons Tower", newStandardGanonsTower},
	{"West Death Mountain", newStandardLWWestDM},
	{"East Death Mountain", newStandardLWEastDM},
	{"North East Light World", newStandardLWNorthEast},
	{"North West Light World", newStandardLWNorthWest},
	{"South Light World", newStandardLWSouth},
	{"West Dark World Death Mountain", newStandardDWWestDM},
	{"East Dark World Death Mountain", newStandardDWEastDM},
	{"Mire", newStandardDWMire},
	{"South Dark World", newStandardDWSouth},
	{"North West Dark World", newStandardDWNorthWest},
	{"North East Dark World", newStandardDWNorthEast},
}

func TestStandardRegions_AllConstruct(t *testing.T) {
	cfg := NewConfig()
	cfg.Strings["mode.weapons"] = "randomized"
	w := NewWorld(0, cfg)
	b := newBuilder(w, item.NewRegistry(), boss.NewRegistry())

	for _, tc := range standardRegions {
		t.Run(tc.name, func(t *testing.T) {
			r := tc.fn(b)
			if r == nil {
				t.Fatal("constructor returned nil")
			}
			if r.Locations.Count() == 0 {
				t.Error("region has no locations")
			}
			// Sanity: every location should reference this region.
			for _, l := range r.Locations.All() {
				if l.Region != r {
					t.Errorf("location %s has wrong region pointer", l.Name)
				}
			}
		})
	}
}

func TestStandardRegions_LocationCount(t *testing.T) {
	cfg := NewConfig()
	cfg.Strings["mode.weapons"] = "randomized"
	w := NewWorld(0, cfg)
	b := newBuilder(w, item.NewRegistry(), boss.NewRegistry())

	total := 0
	for _, tc := range standardRegions {
		r := tc.fn(b)
		total += r.Locations.Count()
	}
	if total < 180 {
		t.Errorf("expected ~180+ locations, got %d", total)
	}
	t.Logf("%d Standard regions ported = %d total locations", len(standardRegions), total)
}
