package world

import (
	"fmt"

	"github.com/JackDoan/alttpr-go/internal/helpers"
	"github.com/JackDoan/alttpr-go/internal/item"
)

// LocationCollection is an ordered set of Locations keyed by full name.
// Mirrors app/Support/LocationCollection.php.
type LocationCollection struct {
	order []*Location
	byKey map[string]*Location
}

// NewLocationCollection builds an empty collection.
func NewLocationCollection(locs ...*Location) *LocationCollection {
	c := &LocationCollection{byKey: map[string]*Location{}}
	for _, l := range locs {
		c.Add(l)
	}
	return c
}

// Add inserts a Location; duplicates by full name are ignored.
func (c *LocationCollection) Add(l *Location) *LocationCollection {
	key := l.FullName()
	if _, ok := c.byKey[key]; ok {
		return c
	}
	c.byKey[key] = l
	c.order = append(c.order, l)
	return c
}

// Remove deletes a Location by full name.
func (c *LocationCollection) Remove(fullName string) *LocationCollection {
	if _, ok := c.byKey[fullName]; !ok {
		return c
	}
	delete(c.byKey, fullName)
	for i, l := range c.order {
		if l.FullName() == fullName {
			c.order = append(c.order[:i], c.order[i+1:]...)
			break
		}
	}
	return c
}

// Get returns a Location by full name, or nil.
func (c *LocationCollection) Get(fullName string) *Location { return c.byKey[fullName] }

// All returns the underlying ordered slice.
func (c *LocationCollection) All() []*Location { return c.order }

// Count returns the number of locations.
func (c *LocationCollection) Count() int { return len(c.order) }

// Filter returns a new collection containing locations where keep(l) is true.
func (c *LocationCollection) Filter(keep func(*Location) bool) *LocationCollection {
	out := NewLocationCollection()
	for _, l := range c.order {
		if keep(l) {
			out.Add(l)
		}
	}
	return out
}

// Empty returns locations without an Item.
func (c *LocationCollection) Empty() *LocationCollection {
	return c.Filter(func(l *Location) bool { return !l.HasItem() })
}

// NonEmpty returns locations that already have an Item.
func (c *LocationCollection) NonEmpty() *LocationCollection {
	return c.Filter(func(l *Location) bool { return l.HasItem() })
}

// ItemInLocations reports whether `it` is placed at least `count` times
// across the named locations (by bare name, world inferred from this
// collection's checksForWorld). Mirrors PHP LocationCollection::itemInLocations.
func (c *LocationCollection) ItemInLocations(it *item.Item, worldID int, names []string, count int) bool {
	for _, name := range names {
		key := fmt.Sprintf("%s:%d", name, worldID)
		if l, ok := c.byKey[key]; ok && l.HasSpecificItem(it) {
			count--
		}
	}
	return count < 1
}

// LocationsWithItem returns locations whose placed item matches `it`.
// If `it` is nil, returns all locations with any item.
func (c *LocationCollection) LocationsWithItem(it *item.Item) *LocationCollection {
	return c.Filter(func(l *Location) bool {
		if !l.HasItem() {
			return false
		}
		if it == nil {
			return true
		}
		return l.HasSpecificItem(it)
	})
}

// CanAccess returns locations Link can reach with the given items.
func (c *LocationCollection) CanAccess(items *item.Collection) *LocationCollection {
	return c.Filter(func(l *Location) bool { return l.CanAccess(items, c) })
}

// Items collects the (non-empty) placed items as an ItemCollection.
func (c *LocationCollection) Items() *item.Collection {
	out := item.NewCollection()
	for _, l := range c.order {
		if l.HasItem() {
			out.Add(l.Item())
		}
	}
	return out
}

// Random returns a random sub-collection of up to n locations (no repeats).
// Mirrors PHP LocationCollection::randomCollection.
func (c *LocationCollection) Random(n int) (*LocationCollection, error) {
	if n >= c.Count() {
		shuffled, err := helpers.FyShuffle(c.order)
		if err != nil {
			return nil, err
		}
		return NewLocationCollection(shuffled...), nil
	}
	shuffled, err := helpers.FyShuffle(c.order)
	if err != nil {
		return nil, err
	}
	return NewLocationCollection(shuffled[:n]...), nil
}

// First returns the first location, or nil for empty.
func (c *LocationCollection) First() *Location {
	if len(c.order) == 0 {
		return nil
	}
	return c.order[0]
}

// Reverse returns a new collection in reverse order.
func (c *LocationCollection) Reverse() *LocationCollection {
	out := NewLocationCollection()
	for i := len(c.order) - 1; i >= 0; i-- {
		out.Add(c.order[i])
	}
	return out
}

// Merge returns a new collection combining this and `other`'s locations.
func (c *LocationCollection) Merge(other *LocationCollection) *LocationCollection {
	out := NewLocationCollection()
	for _, l := range c.order {
		out.Add(l)
	}
	if other != nil {
		for _, l := range other.order {
			out.Add(l)
		}
	}
	return out
}

// Copy returns a shallow copy (location pointers shared).
func (c *LocationCollection) Copy() *LocationCollection {
	out := NewLocationCollection()
	for _, l := range c.order {
		out.Add(l)
	}
	return out
}
