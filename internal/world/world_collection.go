package world

import (
	"github.com/JackDoan/alttpr-go/internal/item"
)

// WinCondition is the predicate (set per-world by the Randomizer) that
// reports whether the goal is reachable with the given items.
// Mirrors PHP $world->getWinCondition()(items).
type WinCondition func(*item.Collection) bool

// WorldCollection is a service over one or more Worlds. Mirrors
// app/Support/WorldCollection.php — used by the Filler to test win
// conditions across all worlds (multiworld) but reduces cleanly to one.
type WorldCollection struct {
	worlds []*World
}

// NewWorldCollection wraps the given worlds.
func NewWorldCollection(ws ...*World) *WorldCollection { return &WorldCollection{worlds: ws} }

// First returns the first world (panics if empty).
func (wc *WorldCollection) First() *World { return wc.worlds[0] }

// Get returns the world with the given ID, or nil.
func (wc *WorldCollection) Get(id int) *World {
	for _, w := range wc.worlds {
		if w.ID() == id {
			return w
		}
	}
	return nil
}

// All returns the wrapped worlds.
func (wc *WorldCollection) All() []*World { return wc.worlds }

// IsWinnable runs the sphere expansion across all worlds and reports
// whether every world's win condition is satisfied. Mirrors PHP
// WorldCollection::isWinnable.
func (wc *WorldCollection) IsWinnable() bool {
	for _, w := range wc.worlds {
		w.ResetCollectedLocations()
	}
	assumed := item.NewCollection()
	for _, w := range wc.worlds {
		assumed = assumed.Merge(w.PreCollectedItems())
	}
	prev := -1
	for {
		curr := 0
		for _, w := range wc.worlds {
			gained := w.CollectOtherItems(assumed)
			assumed = assumed.Merge(gained)
			curr += w.CollectedLocationsCount()
		}
		if curr == prev {
			break
		}
		prev = curr
	}
	for _, w := range wc.worlds {
		if w.WinCondition == nil {
			continue
		}
		if !w.WinCondition(assumed) {
			return false
		}
	}
	return true
}

// CheckWinCondition reports whether every world's win condition is
// satisfied with the given items (no sphere expansion).
// Mirrors PHP WorldCollection::checkWinCondition.
func (wc *WorldCollection) CheckWinCondition(items *item.Collection) bool {
	for _, w := range wc.worlds {
		if w.WinCondition == nil {
			continue
		}
		if !w.WinCondition(items) {
			return false
		}
	}
	return true
}
