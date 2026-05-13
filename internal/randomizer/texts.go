package randomizer

import (
	"fmt"

	"github.com/JackDoan/alttpr-go/internal/helpers"
	"github.com/JackDoan/alttpr-go/internal/item"
	"github.com/JackDoan/alttpr-go/internal/world"
)

// pickFromPool returns a single random entry from a parsed string pool.
func pickFromPool(pool []string) (string, error) {
	if len(pool) == 0 {
		return "", nil
	}
	idx, err := helpers.GetRandomInt(0, len(pool)-1)
	if err != nil {
		return "", err
	}
	return pool[idx], nil
}

// setTexts customizes a small set of in-game dialogue strings based on the
// randomized state. Mirrors a deterministic-only subset of PHP
// Randomizer::setTexts — pool-shuffled strings (uncle/blind/tavern_man hints)
// keep their dumped defaults so we don't depend on PHP's RNG.
//
// Strings customized:
//   - sahasrahla_bring_courage: where the green pendant lives
//   - bomb_shop:                where Crystal 5 / Crystal 6 live
//   - sign_ganons_tower:        tower crystal requirement (when < 7)
//   - sign_ganon, ganon_fall_in_alt: goal-specific
func (r *Randomizer) setTexts(w *world.World) error {
	// Random hint pools from strings/*.txt.
	uncle, err := pickFromPool(parseStringPool(stringsUncle))
	if err != nil {
		return err
	}
	blind, err := pickFromPool(parseStringPool(stringsBlind))
	if err != nil {
		return err
	}
	tavern, err := pickFromPool(parseStringPool(stringsTavernMan))
	if err != nil {
		return err
	}
	ganonFallIn, err := pickFromPool(parseStringPool(stringsGanon1))
	if err != nil {
		return err
	}
	triforce, err := pickFromPool(parseStringPool(stringsTriforce))
	if err != nil {
		return err
	}
	fakeSilvers, err := pickFromPool(parseStringPool(stringsGanonNoSilvers))
	if err != nil {
		return err
	}

	if uncle != "" {
		_ = w.SetText("uncle_leaving_text", uncle)
	}
	if blind != "" {
		_ = w.SetText("blind_by_the_light", blind)
	}
	if tavern != "" {
		_ = w.SetText("kakariko_tavern_fisherman", tavern)
	}
	if ganonFallIn != "" {
		_ = w.SetText("ganon_fall_in", ganonFallIn)
	}
	if triforce != "" {
		_ = w.SetText("end_triforce", "{NOBORDER}\n"+triforce)
	}
	_ = w.SetText("ganon_phase_3_alt", "Got wax in\nyour ears?\nI cannot die!")
	_ = w.SetText("ganon_phase_3_no_silvers", fakeSilvers)
	_ = w.SetText("ganon_phase_3_no_silvers_alt", fakeSilvers)

	// Helper: find first location holding an item by name.
	first := func(name string) *world.Location {
		it, err := r.itemRegistry.Get(name, w.ID())
		if err != nil {
			return nil
		}
		locs := w.Locations().LocationsWithItem(it)
		if locs.Count() == 0 {
			return nil
		}
		return locs.First()
	}

	if l := first("PendantOfCourage"); l != nil && l.Region != nil {
		_ = w.SetText("sahasrahla_bring_courage",
			fmt.Sprintf("Want something\nfor free? Go\nearn the green\npendant in\n%s\nand I'll give\nyou something.",
				l.Region.Name))
	}

	c5 := first("Crystal5")
	c6 := first("Crystal6")
	if c5 != nil && c6 != nil && c5.Region != nil && c6.Region != nil {
		_ = w.SetText("bomb_shop",
			fmt.Sprintf("bring me the\ncrystals from\n%s\nand\n%s\nso I can make\na big bomb!",
				c5.Region.Name, c6.Region.Name))
	}

	tower := w.ConfigInt("crystals.tower", 7)
	if tower < 7 {
		fmtStr := "You need %d crystals to enter."
		if tower == 1 {
			fmtStr = "You need %d crystal to enter."
		}
		_ = w.SetText("sign_ganons_tower", fmt.Sprintf(fmtStr, tower))
	}

	goal := w.ConfigString("goal", "ganon")
	ganon := w.ConfigInt("crystals.ganon", 7)
	ganonPhase3FallIn := "You think you\nare ready to\nface me?\n\nI will not die\n\nunless you\ncomplete your\ngoals. Dingus!"

	var ganonSingular, ganonPlural string
	switch goal {
	case "ganon":
		ganonSingular = "To beat Ganon you must collect %d Crystal and defeat his minion at the top of his tower."
		ganonPlural = "To beat Ganon you must collect %d Crystals and defeat his minion at the top of his tower."
	default:
		ganonSingular = "You need %d Crystal to beat Ganon."
		ganonPlural = "You need %d Crystals to beat Ganon."
	}
	ganonFmt := ganonPlural
	if ganon == 1 {
		ganonFmt = ganonSingular
	}

	switch goal {
	case "ganon", "fast_ganon":
		_ = w.SetText("sign_ganon", fmt.Sprintf(ganonFmt, ganon))
		_ = w.SetText("ganon_fall_in_alt", ganonPhase3FallIn)
	case "ganonhunt":
		_ = w.SetText("sign_ganon", fmt.Sprintf("To beat Ganon you must collect %d Triforce Pieces.",
			w.ConfigInt("item.Goal.Required", 0)))
		_ = w.SetText("ganon_fall_in_alt", ganonPhase3FallIn)
	case "pedestal":
		_ = w.SetText("ganon_fall_in_alt",
			"You cannot\nkill me. You\nshould go for\nyour real goal\nIt's on the\npedestal.\n\nYou dingus!\n")
		_ = w.SetText("sign_ganon", "You need to get to the pedestal... Dingus!")
	case "triforce-hunt":
		_ = w.SetText("ganon_fall_in_alt",
			"So you thought\nyou could come\nhere and beat\nme? I have\nhidden the\nTriforce\npieces well.\nWithout them,\nyou can't win!")
		_ = w.SetText("sign_ganon", "Go find the Triforce pieces... Dingus!")
	case "dungeons":
		_ = w.SetText("sign_ganon", "You need to defeat all of Ganon's bosses.")
		_ = w.SetText("ganon_fall_in_alt", ganonPhase3FallIn)
	case "completionist":
		_ = w.SetText("sign_ganon", "You need to collect EVERY item and defeat EVERY boss.")
		_ = w.SetText("ganon_fall_in_alt", ganonPhase3FallIn)
	default:
		_ = w.SetText("ganon_fall_in_alt", ganonPhase3FallIn)
	}
	return nil
}

// Avoid lint of unused import while developing.
var _ = item.NewCollection
