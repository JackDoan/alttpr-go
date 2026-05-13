// Package randomizer — hints subsystem. Port of app/Services/HintService.php.
package randomizer

import (
	_ "embed"
	"encoding/json"
	"strings"

	"github.com/JackDoan/alttpr-go/internal/helpers"
	"github.com/JackDoan/alttpr-go/internal/item"
	"github.com/JackDoan/alttpr-go/internal/world"
)

//go:embed hints_en.json
var hintsEN []byte

//go:embed strings/hint.txt
var stringsHintJokes string

// hintTable is the parsed resources/lang/en/hint.php — `item` and `location`
// dictionaries whose values are either a single string or a list of options.
type hintTable struct {
	Item     map[string]any `json:"item"`
	Location map[string]any `json:"location"`
}

var parsedHints *hintTable

func loadHints() *hintTable {
	if parsedHints != nil {
		return parsedHints
	}
	var h hintTable
	if err := json.Unmarshal(hintsEN, &h); err != nil {
		panic("hints: cannot decode embedded dictionary: " + err.Error())
	}
	parsedHints = &h
	return parsedHints
}

// pickHintString resolves a hint table entry that may be a string or a list.
func pickHintString(v any) (string, error) {
	switch x := v.(type) {
	case string:
		return x, nil
	case []any:
		if len(x) == 0 {
			return "", nil
		}
		idx, err := helpers.GetRandomInt(0, len(x)-1)
		if err != nil {
			return "", err
		}
		if s, ok := x[idx].(string); ok {
			return s, nil
		}
	}
	return "", nil
}

// locationHint builds the per-location hint string ("foo at bar").
// Mirrors PHP Location::getHint.
func locationHint(l *world.Location) (string, error) {
	if !l.HasItem() {
		return "", nil
	}
	h := loadHints()
	it := l.Item()
	itemKey := it.Name
	if it.Type == item.TypeAlias && it.Target != nil {
		itemKey = it.Target.Name
	}
	itemPhrase, err := pickHintString(h.Item[itemKey])
	if err != nil {
		return "", err
	}
	if itemPhrase == "" {
		return "", nil
	}
	locPhrase, err := pickHintString(h.Location[l.Name])
	if err != nil {
		return "", err
	}
	if locPhrase == "" {
		return "", nil
	}
	combined := itemPhrase + " " + locPhrase
	// PHP `ucfirst`: only the first letter is uppercased.
	if len(combined) > 0 {
		combined = strings.ToUpper(combined[:1]) + combined[1:]
	}
	return combined, nil
}

// applyHints customizes telepathic-tile texts with placement hints.
// Mirrors PHP HintService::applyHints (deterministic subset; pool selections
// use Go's CSPRNG, which doesn't match PHP byte-for-byte).
func (r *Randomizer) applyHints(w *world.World) error {
	if w.ConfigString("spoil.Hints", "on") != "on" {
		// PHP writes a randomizer-version sign when hints are off.
		_ = w.SetText("sign_north_of_links_house", "Randomizer v31\n\n>    -veetorp")
		return nil
	}

	tiles := []string{
		"telepathic_tile_eastern_palace",
		"telepathic_tile_tower_of_hera_floor_4",
		"telepathic_tile_spectacle_rock",
		"telepathic_tile_swamp_entrance",
		"telepathic_tile_thieves_town_upstairs",
		"telepathic_tile_misery_mire",
		"telepathic_tile_palace_of_darkness",
		"telepathic_tile_desert_bonk_torch_room",
		"telepathic_tile_castle_tower",
		"telepathic_tile_ice_large_room",
		"telepathic_tile_turtle_rock",
		"telepathic_tile_ice_entrace",
		"telepathic_tile_ice_stalfos_knights_room",
		"telepathic_tile_tower_of_hera_entrance",
		"telepathic_tile_south_east_darkworld_cave",
	}
	tiles, err := helpers.FyShuffle(tiles)
	if err != nil {
		return err
	}

	pop := func() string {
		if len(tiles) == 0 {
			return ""
		}
		s := tiles[len(tiles)-1]
		tiles = tiles[:len(tiles)-1]
		return s
	}

	first := func(name string) *world.Location {
		it, err := r.itemRegistry.Get(name, w.ID())
		if err != nil {
			return nil
		}
		c := w.Locations().LocationsWithItem(it)
		if c.Count() == 0 {
			return nil
		}
		return c.First()
	}

	// Wild big keys → reveal Ganon's Tower BigKey location.
	if w.ConfigBool("region.wildBigKeys", false) {
		if l := first("BigKeyA2"); l != nil {
			if tile := pop(); tile != "" {
				if h, err := locationHint(l); err == nil && h != "" {
					_ = w.SetText(tile, h)
				}
			}
		}
	}

	// Pegasus Boots location hint.
	if l := first("PegasusBoots"); l != nil {
		if tile := pop(); tile != "" {
			if h, err := locationHint(l); err == nil && h != "" {
				_ = w.SetText(tile, h)
			}
		}
	}

	// Predefined location hints — pick 5 from the candidate set.
	locationCandidates := []string{
		"Sahasrahla", "Mimic Cave", "Catfish",
		"Graveyard Ledge", "Purple Chest",
		"Tower of Hera - Big Key Chest", "Swamp Palace - Big Chest",
		"Misery Mire - Big Key Chest", "Swamp Palace - Big Key Chest",
		"Pyramid Fairy - Left",
	}
	locationCandidates, err = helpers.FyShuffle(locationCandidates)
	if err != nil {
		return err
	}
	count := min(5, len(locationCandidates))
	for i := 0; i < count; i++ {
		l := w.Locations().Get(locationCandidates[i] + ":" + itoaW(w.ID()))
		if l == nil {
			continue
		}
		h, err := locationHint(l)
		if err != nil || h == "" {
			continue
		}
		tile := pop()
		if tile == "" {
			break
		}
		_ = w.SetText(tile, h)
	}

	// Item hints — pick 4 advancement items, hint their location.
	hintables := []*item.Item{}
	for _, it := range r.advancementItems {
		if it.IsType(item.TypeShield) || it.IsType(item.TypeKey) ||
			it.IsType(item.TypeMap) || it.IsType(item.TypeCompass) ||
			it.IsType(item.TypeBottle) || it.IsType(item.TypeSword) {
			continue
		}
		if !w.ConfigBool("region.wildBigKeys", false) && it.IsType(item.TypeBigKey) {
			continue
		}
		switch it.Name {
		case "TenBombs", "HalfMagic", "BugCatchingNet", "Powder", "Mushroom":
			continue
		}
		hintables = append(hintables, it)
	}
	hintables, err = helpers.FyShuffle(hintables)
	if err != nil {
		return err
	}
	hintCount := min(4, len(tiles))
	for i := 0; i < hintCount && i < len(hintables); i++ {
		l := first(hintables[i].Name)
		if l == nil {
			continue
		}
		switch l.Kind {
		case world.KindMedallion, world.KindFountain, world.KindPrize,
			world.KindPrizeEvent, world.KindPrizeCrystal, world.KindPrizePendant, world.KindTrade:
			continue
		}
		h, err := locationHint(l)
		if err != nil || h == "" {
			continue
		}
		tile := pop()
		if tile == "" {
			break
		}
		_ = w.SetText(tile, h)
	}

	// Remaining tiles get joke hints.
	jokes := parseStringPool(stringsHintJokes)
	for _, tile := range tiles {
		joke, err := pickFromPool(jokes)
		if err != nil {
			return err
		}
		if joke != "" {
			_ = w.SetText(tile, joke)
		}
	}
	return nil
}
