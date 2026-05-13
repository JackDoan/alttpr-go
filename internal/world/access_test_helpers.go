package world

import (
	"github.com/JackDoan/alttpr-go/internal/item"
)

// AllItemsExcept builds an ItemCollection containing every known item for
// the given world (with manyKeys() applied), minus the named exclusions.
// Mirrors PHP TestCase::allItemsExcept.
//
// PHP-style group exclusion keys are honored:
//
//	"AnySword"      -> L1Sword, L1SwordAndShield, ProgressiveSword, UncleSword
//	"UpgradedSword" -> UncleSword, L2Sword, MasterSword, L3Sword, L4Sword
//	"AnyBottle"     -> all bottle variants
//	"AnyBow"        -> Bow, BowAndArrows, BowAndSilverArrows, ProgressiveBow
//	"Flute"         -> OcarinaActive, OcarinaInactive
//	"Gloves"        -> ProgressiveGlove, PowerGlove, TitansMitt
//
// Keys/BigKeys/ShopKey are always excluded (PHP behavior: not part of
// allItems()) regardless of the explicit list.
func AllItemsExcept(w *World, ir *item.Registry, exclude []string) *item.Collection {
	c := item.NewCollection()
	c.SetChecksForWorld(w.ID())
	// Seed with every item (single copy) — manyKeys() will bump Key* to 10.
	for _, it := range ir.All(w.ID()) {
		c.Add(it)
	}
	c.ManyKeys()

	groupExpand := map[string][]string{
		"AnySword":      {"L1Sword", "L1SwordAndShield", "ProgressiveSword", "UncleSword"},
		"UpgradedSword": {"UncleSword", "L2Sword", "MasterSword", "L3Sword", "L4Sword"},
		"AnyBottle": {
			"BottleWithBee", "BottleWithFairy", "BottleWithRedPotion",
			"BottleWithGreenPotion", "BottleWithBluePotion", "Bottle", "BottleWithGoldBee",
		},
		"AnyBow":  {"Bow", "BowAndArrows", "BowAndSilverArrows", "ProgressiveBow"},
		"Flute":   {"OcarinaActive", "OcarinaInactive"},
		"Gloves":  {"ProgressiveGlove", "PowerGlove", "TitansMitt"},
	}

	removeName := func(name string) {
		// Try to remove all copies (Has + Remove until gone). The
		// collection uses full names internally, so we look up the full name.
		full := name + ":"
		// Construct the full-key form. We use a small helper to render the world id.
		full += itoaW(w.ID())
		for c.CountByFullName(full) > 0 {
			c.Remove(full)
		}
	}

	all := append([]string{"BigKey", "Key", "ShopKey"}, exclude...)
	for _, name := range all {
		if expand, ok := groupExpand[name]; ok {
			for _, sub := range expand {
				removeName(sub)
			}
			// PHP also falls through "AnySword" -> still strips "UpgradedSword" group.
			if name == "AnySword" {
				for _, sub := range groupExpand["UpgradedSword"] {
					removeName(sub)
				}
			}
			continue
		}
		removeName(name)
	}
	return c
}

// AccessCase represents one parametrized test row from a PHP accessPool.
type AccessCase struct {
	Location string
	Access   bool
	Add      []string
	Except   []string
}

// RunAccessCases validates a slice of AccessCases against a world. Returns
// (passed, failures) where failures contain human-readable diffs.
func RunAccessCases(w *World, ir *item.Registry, cases []AccessCase) (passed int, failures []string) {
	for _, tc := range cases {
		var c *item.Collection
		if len(tc.Except) > 0 {
			c = AllItemsExcept(w, ir, tc.Except)
		} else {
			c = item.NewCollection()
			c.SetChecksForWorld(w.ID())
		}
		for _, name := range tc.Add {
			if it, err := ir.Get(name, w.ID()); err == nil {
				c.Add(it)
			}
		}
		// Add RescueZelda (PHP setUp does this for Standard tests).
		if zelda, err := ir.Get("RescueZelda", w.ID()); err == nil {
			c.Add(zelda)
		}

		loc := w.Locations().Get(tc.Location + ":" + itoaW(w.ID()))
		if loc == nil {
			failures = append(failures, "missing location: "+tc.Location)
			continue
		}
		got := loc.CanAccess(c, w.Locations())
		if got != tc.Access {
			failures = append(failures,
				tc.describe()+": got "+boolStr(got)+" want "+boolStr(tc.Access))
			continue
		}
		passed++
	}
	return
}

func boolStr(b bool) string {
	if b {
		return "true"
	}
	return "false"
}

func (tc AccessCase) describe() string {
	out := tc.Location
	if len(tc.Add) > 0 {
		out += " add=["
		for i, n := range tc.Add {
			if i > 0 {
				out += ","
			}
			out += n
		}
		out += "]"
	}
	if len(tc.Except) > 0 {
		out += " except=["
		for i, n := range tc.Except {
			if i > 0 {
				out += ","
			}
			out += n
		}
		out += "]"
	}
	return out
}
