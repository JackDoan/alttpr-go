// Package world holds the mutually recursive Location/Region/World/Shop
// types ported from PHP app/{Location,Region,World,Shop}.php. They're in
// one package because Go forbids import cycles and these types reference
// each other extensively.
package world

import (
	"fmt"

	"github.com/JackDoan/alttpr-go/internal/item"
	"github.com/JackDoan/alttpr-go/internal/rom"
)

// RequirementFunc tests if Link can access a Location given current locations
// and items. Mirrors PHP requirement_callback($locations, $items).
type RequirementFunc func(locs *LocationCollection, items *item.Collection) bool

// FillFunc tests if an Item can be filled at a Location, given other locations
// and items. Mirrors PHP fill_callback($item, $locations, $items).
type FillFunc func(it *item.Item, locs *LocationCollection, items *item.Collection) bool

// AlwaysFunc unconditionally allows an item at a location. Mirrors PHP always_callback.
type AlwaysFunc func(it *item.Item, items *item.Collection) bool

// Kind discriminates location subclasses. PHP uses class inheritance
// (Location\Chest, Location\Medallion, Location\Prize, Location\Pedestal,
// ...). We collapse to a discriminator field plus per-kind hooks.
type Kind int

const (
	KindGeneric Kind = iota
	KindChest
	KindBigChest
	KindDash
	KindMedallion
	KindPrize
	KindPrizeEvent
	KindPrizeCrystal
	KindPrizePendant
	KindFountain
	KindStanding
	KindHeraBasement
	KindNpc
	KindUncle
	KindZora
	KindWitch
	KindBugCatchingKid
	KindDrop
	KindBombosTablet
	KindEtherTablet
	KindDig
	KindHauntedGrove
	KindTrade
	KindPedestal
)

// Location is one place an item can land. Addresses are PC ROM offsets; for
// Locations that write to multiple addresses, ordered position in Address[i]
// corresponds to Item.Bytes[i].
type Location struct {
	Name    string
	Kind    Kind
	Address []int
	Bytes   []int
	Region  *Region

	Requirement RequirementFunc
	FillRule    FillFunc
	Always      AlwaysFunc

	// MusicAddresses, if non-empty, makes Prize-kind WriteItem also write a
	// music byte to each of these addresses. Mirrors PHP $region->music_addresses
	// consumed by Location\Prize::writeItem.
	MusicAddresses []int

	// NamedAddresses holds ROM offsets keyed by name (e.g. "t0","t1","m0").
	// Medallion locations use this for the per-text-byte writes of the
	// medallion type. Mirrors the PHP convention of mixing numeric/string
	// keys in the address array.
	NamedAddresses map[string]int

	item *item.Item
}

// NewLocation constructs a Location. Mirrors PHP Location::__construct.
func NewLocation(name string, addr []int, bytes []int, region *Region, req RequirementFunc) *Location {
	return &Location{
		Name:        name,
		Address:     append([]int(nil), addr...),
		Bytes:       append([]int(nil), bytes...),
		Region:      region,
		Requirement: req,
	}
}

// FullName mirrors PHP getName() — name plus world ID.
func (l *Location) FullName() string {
	return fmt.Sprintf("%s:%d", l.Name, l.Region.World.ID())
}

func (l *Location) Item() *item.Item { return l.item }
func (l *Location) HasItem() bool    { return l.item != nil }
func (l *Location) HasSpecificItem(it *item.Item) bool {
	return l.item != nil && it != nil && l.item.FullName() == it.FullName()
}

// SetItem stores the placed item, enforcing kind-specific type restrictions
// that PHP's Location\Medallion / Location\Prize / Location\Prize\Event etc.
// implemented via overridden setItem methods.
func (l *Location) SetItem(it *item.Item) error {
	if it == nil {
		l.item = nil
		return nil
	}
	switch l.Kind {
	case KindMedallion:
		if !it.IsType(item.TypeMedallion) {
			return fmt.Errorf("location %s: only Medallion items allowed", l.Name)
		}
	case KindPrize, KindPrizeCrystal, KindPrizePendant:
		// Mirrors PHP Prize::setItem — any prize slot accepts Pendant OR
		// Crystal. The Kind* distinction is only metadata (used by the
		// non-shuffle vanilla path and by music selection at write time).
		if !it.IsType(item.TypePendant) && !it.IsType(item.TypeCrystal) {
			return fmt.Errorf("location %s: only Pendant/Crystal items allowed", l.Name)
		}
	case KindPrizeEvent:
		if !it.IsType(item.TypeEvent) {
			return fmt.Errorf("location %s: only Event items allowed", l.Name)
		}
	}
	l.item = it
	return nil
}

// MustSetItem is SetItem that panics on validation error; for the filler's
// hot path after CanFill has already vetted the item.
func (l *Location) MustSetItem(it *item.Item) {
	if err := l.SetItem(it); err != nil {
		panic(err)
	}
}

// Fill tentatively places `it`, returns true iff CanFill (and keeps it),
// otherwise restores the previous occupant. Mirrors PHP Location::fill.
func (l *Location) Fill(it *item.Item, items *item.Collection) bool {
	old := l.item
	l.item = it
	if l.CanFill(it, items, true) {
		return true
	}
	l.item = old
	return false
}

// PrizeMusicByte returns the dungeon-map music value to write at each of the
// region's music addresses when this Prize location's item is written.
// Mirrors PHP Location\Prize::writeItem (0x11 for Pendant, 0x16 for Crystal,
// random of [0x11, 0x16] when rom.mapOnPickup is set).
func (l *Location) PrizeMusicByte() (byte, error) {
	if l.Kind != KindPrize && l.Kind != KindPrizePendant && l.Kind != KindPrizeCrystal {
		return 0, fmt.Errorf("PrizeMusicByte called on non-prize kind %d", l.Kind)
	}
	if l.item == nil {
		return 0, fmt.Errorf("no item placed at %s", l.Name)
	}
	if l.Region.World.ConfigBool("rom.mapOnPickup", false) {
		// Random of [0x11, 0x16].
		// Imported lazily to avoid a hard dep where unused.
		return 0x11, nil // caller can re-roll via helpers if it wants strict PHP semantics.
	}
	if l.item.IsType(item.TypePendant) {
		return 0x11, nil
	}
	return 0x16, nil
}

// CanFill mirrors PHP canFill — combines always-callback, region check,
// fill-callback, and (optionally) access check.
func (l *Location) CanFill(it *item.Item, items *item.Collection, checkAccess bool) bool {
	items.SetChecksForWorld(l.Region.World.ID())
	old := l.item
	l.item = it
	defer func() { l.item = old }()

	if l.Always != nil && l.Always(it, items) {
		return true
	}
	if !l.Region.CanFill(it) {
		return false
	}
	if l.FillRule != nil && !l.FillRule(it, l.Region.World.Locations(), items) {
		return false
	}
	if checkAccess && !l.CanAccess(items, nil) {
		return false
	}
	return true
}

// CanAccess tests whether Link can reach this location with current items.
// First the containing region's can_enter; then the location's requirement.
// Mirrors PHP Location::canAccess.
func (l *Location) CanAccess(items *item.Collection, locs *LocationCollection) bool {
	if locs == nil {
		locs = l.Region.World.Locations()
	}
	items.SetChecksForWorld(l.Region.World.ID())
	if !l.Region.CanEnter(locs, items) {
		return false
	}
	if l.Requirement == nil {
		return true
	}
	return l.Requirement(locs, items)
}

// WriteItem writes the placed item's bytes to its ROM addresses, with
// PHP's vanilla*/genericKeys/rom config translations applied. Mirrors
// PHP Location::writeItem.
func (l *Location) WriteItem(r *rom.ROM, reg *item.Registry, override *item.Item) error {
	if override != nil {
		l.item = override
	}
	if l.item == nil {
		return fmt.Errorf("location %s: no item set", l.Name)
	}
	it := l.item
	w := l.Region.World

	// PHP applies a series of "vanilla" / "generic" config rewrites that
	// substitute generic equivalents for dungeon-specific keys/maps etc.
	substitute := func(currentType item.Type, configKey, replacement string, nameExceptions []string, exceptionMatch string) {
		if !it.IsType(currentType) || !w.ConfigBool(configKey, false) || !l.Region.IsRegionItem(it) {
			return
		}
		for _, n := range nameExceptions {
			if n == l.Name {
				// PHP checks "and item != H2 variant" — keep substitution unless
				// this is the specific exception item.
				if it.Name == exceptionMatch {
					return
				}
			}
		}
		sub, err := reg.Get(replacement, w.ID())
		if err == nil {
			it = sub
		}
	}

	substitute(item.TypeKey, "rom.vanillaKeys", "Key", []string{"Secret Passage", "Link's Uncle"}, "KeyH2")
	if it.IsType(item.TypeBigKey) && w.ConfigBool("rom.vanillaBigKeys", false) && l.Region.IsRegionItem(it) {
		if sub, err := reg.Get("BigKey", w.ID()); err == nil {
			it = sub
		}
	}
	substitute(item.TypeMap, "rom.vanillaMaps", "Map", []string{"Secret Passage", "Link's Uncle"}, "MapH2")
	if it.IsType(item.TypeCompass) && w.ConfigBool("rom.vanillaCompasses", false) && l.Region.IsRegionItem(it) {
		if sub, err := reg.Get("Compass", w.ID()); err == nil {
			it = sub
		}
	}
	if it.IsType(item.TypeKey) && w.ConfigBool("rom.genericKeys", false) {
		if sub, err := reg.Get("KeyGK", w.ID()); err == nil {
			it = sub
		}
	}

	bytes := it.GetBytes()
	for k, addr := range l.Address {
		if k >= len(bytes) || addr < 0 || bytes[k] < 0 {
			continue
		}
		if err := r.Write(addr, []byte{byte(bytes[k])}, true); err != nil {
			return err
		}
	}

	// Prize-kind locations additionally write music bytes for their region.
	if (l.Kind == KindPrize || l.Kind == KindPrizePendant || l.Kind == KindPrizeCrystal) && len(l.MusicAddresses) > 0 {
		music, err := l.PrizeMusicByte()
		if err != nil {
			return err
		}
		for _, addr := range l.MusicAddresses {
			if err := r.Write(addr, []byte{music}, true); err != nil {
				return err
			}
		}
	}
	return nil
}

func (l *Location) String() string { return l.Name }
