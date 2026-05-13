package world

import (
	"github.com/JackDoan/alttpr-go/internal/item"
	"github.com/JackDoan/alttpr-go/internal/logic"
)

// World is the container for all regions, locations, shops, and config for
// one randomized world. Mirrors app/World.php as a skeleton; subclasses
// (Standard/Open/Inverted/Retro) will set Regions/Locations on construction.
//
// Implements logic.World.
type World struct {
	id     int
	config Config

	regions   map[string]*Region
	locations *LocationCollection
	shops     *ShopCollection

	// preCollected holds starting items (e.g. UncleSword). Mirrors PHP
	// World::$pre_collected_items.
	preCollected *item.Collection

	// collectedLocations is the per-fill cache of locations whose items
	// have been "collected" by the filler so far. Mirrors PHP
	// World::$collected_locations.
	collectedLocations map[string]bool

	// WinCondition is set by the Randomizer to the goal predicate
	// (e.g. "have Triforce", "defeated Ganon"). Used by WorldCollection.
	WinCondition WinCondition

	// GanonsTowerJunkFillLow/High control how much junk the filler dumps
	// into Ganon's Tower before placing required items. Mirrors PHP
	// World::getGanonsTowerJunkFillRange.
	GanonsTowerJunkFillLow  int
	GanonsTowerJunkFillHigh int

	// Text is the per-world dialogue table. Populated lazily.
	text *Text

	// Sram is the per-world starting SRAM block. Populated lazily.
	sram *InitialSram

	// credits is the per-world credit-scene table. Populated lazily.
	credits *Credits

	// prizes is the per-world prize-pack drop table. Populated lazily.
	prizes *drops
}

// Sram returns the per-world initial SRAM, creating it on first access.
func (w *World) Sram() *InitialSram {
	if w.sram == nil {
		w.sram = NewInitialSram()
	}
	return w.sram
}

// Credits returns the per-world credits scenes, creating on first access.
func (w *World) Credits() *Credits {
	if w.credits == nil {
		w.credits = NewCredits()
	}
	return w.credits
}

// SetCredit replaces a credit scene's first line with the given text
// (center-aligned). Mirrors PHP World::setCredit.
func (w *World) SetCredit(scene, text string) bool {
	return w.Credits().UpdateCreditLine(scene, 0, text, "center")
}

// Text returns the per-world text table, creating it on first access.
func (w *World) Text() *Text {
	if w.text == nil {
		w.text = NewText()
	}
	return w.text
}

// SetText updates a named dialog string. Mirrors PHP World::setText
// which delegates to Rom->setText -> Text->setString.
func (w *World) SetText(name, value string) error {
	return w.Text().SetString(name, value)
}

// Config is a typed config bag. PHP uses dotted keys ("rom.rupeeBow"); we
// preserve that via plain map lookups. Subtype methods know which key
// they're after.
type Config struct {
	Strings map[string]string
	Ints    map[string]int
	Bools   map[string]bool
}

// NewConfig returns an empty Config.
func NewConfig() Config {
	return Config{
		Strings: map[string]string{},
		Ints:    map[string]int{},
		Bools:   map[string]bool{},
	}
}

// NewWorld builds a bare World; subclass constructors populate regions/locations.
func NewWorld(id int, cfg Config) *World {
	return &World{
		id:                 id,
		config:             cfg,
		regions:            map[string]*Region{},
		locations:          NewLocationCollection(),
		shops:              NewShopCollection(),
		preCollected:       item.NewCollection(),
		collectedLocations: map[string]bool{},
	}
}

// --- logic.World implementation ---

func (w *World) ID() int          { return w.id }
func (w *World) IsInverted() bool { return w.config.Strings["world.variant"] == "inverted" }
func (w *World) ConfigString(key, def string) string {
	if v, ok := w.config.Strings[key]; ok {
		return v
	}
	return def
}
func (w *World) ConfigInt(key string, def int) int {
	if v, ok := w.config.Ints[key]; ok {
		return v
	}
	return def
}
func (w *World) ConfigBool(key string, def bool) bool {
	if v, ok := w.config.Bools[key]; ok {
		return v
	}
	return def
}

// compile-time check: *World satisfies logic.World.
var _ logic.World = (*World)(nil)

// --- World accessors ---

func (w *World) Locations() *LocationCollection { return w.locations }
func (w *World) Shops() *ShopCollection         { return w.shops }
func (w *World) Region(name string) *Region     { return w.regions[name] }
func (w *World) AllRegions() map[string]*Region { return w.regions }
func (w *World) Config() Config                 { return w.config }

// AddRegion registers a region with the world, merging its locations and
// shops into the world-wide collections.
func (w *World) AddRegion(r *Region) {
	w.regions[r.Name] = r
	for _, l := range r.Locations.All() {
		w.locations.Add(l)
	}
	for _, s := range r.Shops.All() {
		w.shops.Add(s)
	}
}

// PreCollectedItems returns the starting items collection.
func (w *World) PreCollectedItems() *item.Collection { return w.preCollected }

// AddPreCollectedItem adds an item to the pre-collected pool.
func (w *World) AddPreCollectedItem(it *item.Item) { w.preCollected.Add(it) }

// ResetCollectedLocations clears the per-fill collection cache.
func (w *World) ResetCollectedLocations() { w.collectedLocations = map[string]bool{} }

// CollectedLocationsCount returns how many locations the filler has
// already marked as collected.
func (w *World) CollectedLocationsCount() int { return len(w.collectedLocations) }

// CollectOtherItems walks through locations accessible with `items`, adds
// their items to a returned collection, and marks them as collected.
// Mirrors PHP World::collectOtherItems.
func (w *World) CollectOtherItems(items *item.Collection) *item.Collection {
	gained := item.NewCollection()
	for _, loc := range w.locations.All() {
		if !loc.HasItem() {
			continue
		}
		key := loc.FullName()
		if w.collectedLocations[key] {
			continue
		}
		if !loc.CanAccess(items, w.locations) {
			continue
		}
		w.collectedLocations[key] = true
		gained.Add(loc.Item())
	}
	return gained
}

// EmptyLocations returns the world-wide collection of locations without an item.
func (w *World) EmptyLocations() *LocationCollection { return w.locations.Empty() }

// CollectableLocations returns locations that hold items the player can pick up
// (excluding Medallion and Fountain slots and already-collected entries).
// Mirrors PHP World::getCollectableLocations.
func (w *World) CollectableLocations() *LocationCollection {
	return w.locations.Filter(func(l *Location) bool {
		if l.Kind == KindMedallion || l.Kind == KindFountain {
			return false
		}
		return !w.collectedLocations[l.FullName()]
	})
}

// LocationSpheres returns the list of location collections accessible at
// each "sphere" of collection (sphere 0 = pre-collected items). Mirrors PHP
// World::getLocationSpheres.
func (w *World) LocationSpheres() []*LocationCollection {
	// Sphere 0 = synthetic locations holding each pre-collected item.
	spheres := []*LocationCollection{NewLocationCollection()}
	myItems := w.preCollected.Copy()
	found := NewLocationCollection()

	for {
		sphere := w.CollectableLocations().Filter(func(l *Location) bool {
			if !l.HasItem() {
				return false
			}
			if found.Get(l.FullName()) != nil {
				return false
			}
			return l.CanAccess(myItems, w.locations)
		})
		if sphere.Count() == 0 {
			break
		}
		spheres = append(spheres, sphere)
		for _, l := range sphere.All() {
			found.Add(l)
			if l.HasItem() {
				myItems.Add(l.Item())
			}
		}
	}
	return spheres
}

// CollectItems runs a fixed-point sphere expansion: starting from
// pre-collected items, repeatedly walk all accessible locations and add
// their placed items to the collection until no new locations open up.
// Mirrors PHP World::collectItems / collectItemsAfterCheck.
func (w *World) CollectItems() *item.Collection {
	return w.CollectItemsFrom(nil)
}

// CollectItemsFrom mirrors PHP World::collectItems($collected): starts from
// `starting + preCollected` and expands by walking accessible locations
// until fixed point. Used by the prize filler to compute "assumed items".
func (w *World) CollectItemsFrom(starting *item.Collection) *item.Collection {
	w.ResetCollectedLocations()
	collected := item.NewCollection()
	if starting != nil {
		collected = collected.Merge(starting)
	}
	collected = collected.Merge(w.preCollected)
	collected.SetChecksForWorld(w.id)
	prev := -1
	for {
		gained := w.CollectOtherItems(collected)
		collected = collected.Merge(gained)
		curr := len(w.collectedLocations)
		if curr == prev {
			return collected
		}
		prev = curr
	}
}
