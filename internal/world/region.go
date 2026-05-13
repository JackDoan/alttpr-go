package world

import (
	"github.com/JackDoan/alttpr-go/internal/boss"
	"github.com/JackDoan/alttpr-go/internal/item"
)

// EnterFunc returns whether Link can enter a Region with current items.
// Mirrors PHP can_enter($locations, $items).
type EnterFunc func(locs *LocationCollection, items *item.Collection) bool

// CompleteFunc returns whether the Region's prize can be obtained.
// Mirrors PHP can_complete.
type CompleteFunc func(locs *LocationCollection, items *item.Collection) bool

// Region is a logical group of Locations with shared accessibility rules.
// Mirrors app/Region.php.
type Region struct {
	Name        string
	World       *World
	Locations   *LocationCollection
	Shops       *ShopCollection
	CanEnterFn  EnterFunc
	CanCompFn   CompleteFunc
	// CanPlaceBossFn, if set, overrides the default boss-placement test
	// for this region (e.g. Tower of Hera restricts Moldorm-equivalents).
	CanPlaceBossFn func(*boss.Boss) bool
	// CanPlaceBossLevelFn, if set, overrides the default for a specific boss
	// level (used by Ganon's Tower, which has top/middle/bottom slots).
	CanPlaceBossLevelFn func(b *boss.Boss, level string) bool
	RegionItems         []*item.Item // items "guaranteed" to spawn in this region (used by canFill)
	Boss                *boss.Boss
	// Bosses holds named boss slots (e.g. "top", "middle", "bottom") for
	// regions with multiple bosses (Ganon's Tower).
	Bosses    map[string]*boss.Boss
	MapReveal int

	prizeLocation *Location
}

// NewRegion constructs a bare Region; subclasses populate Locations/RegionItems.
func NewRegion(name string, w *World) *Region {
	return &Region{
		Name:      name,
		World:     w,
		Locations: NewLocationCollection(),
		Shops:     NewShopCollection(),
		Bosses:    map[string]*boss.Boss{},
	}
}

// BossAt returns the boss for a given level slot. Empty level returns the
// default Boss. Levels "top"/"middle"/"bottom" check the Bosses map.
func (r *Region) BossAt(level string) *boss.Boss {
	if level == "" {
		return r.Boss
	}
	return r.Bosses[level]
}

// CanEnter mirrors PHP Region::canEnter.
func (r *Region) CanEnter(locs *LocationCollection, items *item.Collection) bool {
	if r.CanEnterFn == nil {
		return true
	}
	return r.CanEnterFn(locs, items)
}

// CanComplete mirrors PHP Region::canComplete.
func (r *Region) CanComplete(locs *LocationCollection, items *item.Collection) bool {
	if r.CanCompFn == nil {
		return true
	}
	return r.CanCompFn(locs, items)
}

// CanFill enforces the wild* config flags: certain dungeon items must stay
// in their home region unless the corresponding wildKeys/etc. is enabled.
// Mirrors PHP Region::canFill.
func (r *Region) CanFill(it *item.Item) bool {
	fromWorld := r.World // items always carry their world; in single-world setups this equals r.World
	if it.WorldID != r.World.ID() {
		// In multi-world setups we'd need to look up the other world for its
		// config; for now we use the placing region's world as a proxy.
	}

	isDungeon := false
	switch {
	case it.IsType(item.TypeKey) && !fromWorld.ConfigBool("region.wildKeys", false):
		isDungeon = true
	case it.IsType(item.TypeBigKey) && !fromWorld.ConfigBool("region.wildBigKeys", false):
		isDungeon = true
	case it.IsType(item.TypeMap) && !fromWorld.ConfigBool("region.wildMaps", false):
		isDungeon = true
	case it.IsType(item.TypeCompass) && !fromWorld.ConfigBool("region.wildCompasses", false):
		isDungeon = true
	}

	// Sewers Key (KeyH2) cannot leave its home region in standard, non-NoLogic.
	if it.Name == "KeyH2" &&
		fromWorld.ConfigString("mode.state", "") == "standard" &&
		fromWorld.ConfigString("logic", "") != "NoLogic" {
		isDungeon = true
	}

	if isDungeon && !r.IsRegionItem(it) {
		return false
	}
	return true
}

// IsRegionItem reports whether `it` belongs to this region's RegionItems set.
func (r *Region) IsRegionItem(target *item.Item) bool {
	for _, ri := range r.RegionItems {
		if ri == target || ri.FullName() == target.FullName() {
			return true
		}
	}
	return false
}

// CanPlaceBoss mirrors PHP Region::canPlaceBoss.
func (r *Region) CanPlaceBoss(b *boss.Boss) bool {
	if r.CanPlaceBossFn != nil {
		return r.CanPlaceBossFn(b)
	}
	if r.Name != "Ice Palace" &&
		r.World.ConfigString("mode.weapons", "") == "swordless" &&
		b.Name == "Kholdstare" {
		return false
	}
	switch b.Name {
	case "Agahnim", "Agahnim2", "Ganon":
		return false
	}
	return true
}

// SetPrizeLocation wires a prize Location into the Region; CanComplete becomes
// the location's requirement if set.
func (r *Region) SetPrizeLocation(l *Location) *Region {
	r.prizeLocation = l
	l.Region = r
	if r.CanCompFn != nil {
		l.Requirement = RequirementFunc(r.CanCompFn)
	}
	return r
}

// PrizeLocation returns the prize Location, or nil.
func (r *Region) PrizeLocation() *Location { return r.prizeLocation }

// Prize returns the placed prize item, or nil.
func (r *Region) Prize() *item.Item {
	if r.prizeLocation == nil || !r.prizeLocation.HasItem() {
		return nil
	}
	return r.prizeLocation.Item()
}

// HasPrize reports whether the region's prize equals `it` (or whether any
// prize is placed when `it` is nil).
func (r *Region) HasPrize(it *item.Item) bool {
	if r.prizeLocation == nil || !r.prizeLocation.HasItem() {
		return false
	}
	if it == nil {
		return true
	}
	return r.prizeLocation.HasSpecificItem(it)
}

// EmptyLocations returns the Region's locations without an item.
func (r *Region) EmptyLocations() *LocationCollection { return r.Locations.Empty() }
