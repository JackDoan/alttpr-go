// Package filler is the Go port of app/Filler.php and app/Filler/RandomAssumed.php.
package filler

import (
	"fmt"

	"github.com/JackDoan/alttpr-go/internal/helpers"
	"github.com/JackDoan/alttpr-go/internal/item"
	"github.com/JackDoan/alttpr-go/internal/world"
)

// Filler is the abstract base. PHP factory selects RandomAssumed.
type Filler interface {
	Fill(dungeon, required, nice, extra []*item.Item) error
}

// RandomAssumed implements the assumed-fill algorithm.
// Mirrors app/Filler/RandomAssumed.php:RandomAssumed.
type RandomAssumed struct {
	worlds []*world.World
}

// NewRandomAssumed builds an assumed filler over the given worlds.
func NewRandomAssumed(worlds ...*world.World) *RandomAssumed {
	return &RandomAssumed{worlds: worlds}
}

// Fill places dungeon items first (full assumption), then required, then
// nice, then extra. Mirrors PHP RandomAssumed::fill.
func (f *RandomAssumed) Fill(dungeon, required, nice, extra []*item.Item) error {
	allLocs := world.NewLocationCollection()
	for _, w := range f.worlds {
		allLocs = allLocs.Merge(w.EmptyLocations())
	}
	shuffled, err := shuffleLocations(allLocs)
	if err != nil {
		return err
	}

	// Dungeon items pre-placed against required+nice as assumed.
	reqPlusNice := make([]*item.Item, 0, len(required)+len(nice))
	reqPlusNice = append(reqPlusNice, required...)
	reqPlusNice = append(reqPlusNice, nice...)
	if err := f.fillItemsInLocations(dungeon, shuffled, reqPlusNice); err != nil {
		return err
	}

	// Junk fill for Ganon's Tower (per-world).
	for _, w := range f.worlds {
		gt := w.Region("Ganons Tower")
		if gt == nil {
			continue
		}
		lo, hi := w.GanonsTowerJunkFillLow, w.GanonsTowerJunkFillHigh
		if hi < lo {
			lo, hi = hi, lo
		}
		n, err := helpers.GetRandomInt(lo, hi)
		if err != nil {
			return err
		}
		gtLocs, err := gt.EmptyLocations().Random(n)
		if err != nil {
			return err
		}
		shuffledExtra, err := helpers.FyShuffle(extra)
		if err != nil {
			return err
		}
		cut := min(gtLocs.Count(), len(shuffledExtra))
		trash := shuffledExtra[:cut]
		extra = shuffledExtra[cut:]
		fastFill(trash, gtLocs)
	}

	// Required: fill in reverse order against remaining empty locations.
	emptyRev := shuffled.Empty().Reverse()
	shuffledReq, err := helpers.FyShuffle(required)
	if err != nil {
		return err
	}
	if err := f.fillItemsInLocations(shuffledReq, emptyRev, nil); err != nil {
		return err
	}

	// Nice + extra: fast-fill (no logic check) into remaining locations.
	shuffledNice, err := helpers.FyShuffle(nice)
	if err != nil {
		return err
	}
	remainingLocs, err := shuffleLocations(emptyRev.Empty())
	if err != nil {
		return err
	}
	fastFill(shuffledNice, remainingLocs)

	shuffledExtra, err := helpers.FyShuffle(extra)
	if err != nil {
		return err
	}
	fastFill(shuffledExtra, remainingLocs.Empty())
	return nil
}

// fillItemsInLocations places each item at the first valid location,
// computing reachable items by sphere expansion that *assumes* the
// remaining items are eventually placed (the assumed-fill technique).
// Mirrors PHP RandomAssumed::fillItemsInLocations.
func (f *RandomAssumed) fillItemsInLocations(items []*item.Item, locs *world.LocationCollection, baseAssumed []*item.Item) error {
	remaining := item.NewCollection(items...)
	if remaining.Count() > locs.Empty().Count() {
		return fmt.Errorf("trying to fill more items (%d) than locations (%d)",
			remaining.Count(), locs.Empty().Count())
	}
	wc := world.NewWorldCollection(f.worlds...)

	for _, it := range items {
		remaining.Remove(it.FullName())
		base := item.NewCollection(baseAssumed...)
		starting := remaining.Merge(base)

		for _, w := range f.worlds {
			w.ResetCollectedLocations()
		}

		assumed := starting.Copy()
		for _, w := range f.worlds {
			assumed = assumed.Merge(w.PreCollectedItems())
		}
		prev := -1
		for {
			curr := 0
			for _, w := range f.worlds {
				gained := w.CollectOtherItems(assumed)
				assumed = assumed.Merge(gained)
				curr += w.CollectedLocationsCount()
			}
			if curr == prev {
				break
			}
			prev = curr
		}

		// PHP item.getWorld() returns the World the item was created for.
		// We approximate by using the first world (single-world mode);
		// multi-world would route via worlds[item.WorldID].
		homeWorld := wc.Get(it.WorldID)
		if homeWorld == nil {
			homeWorld = f.worlds[0]
		}
		canSkipAccess := homeWorld.ConfigString("accessibility", "") == "none" &&
			(!it.IsType(item.TypeKey) || homeWorld.ConfigBool("region.wildKeys", false)) &&
			(!it.IsType(item.TypeBigKey) || homeWorld.ConfigBool("region.wildBigKeys", false)) &&
			(!it.IsType(item.TypeMap) || homeWorld.ConfigBool("region.wildMaps", false)) &&
			(!it.IsType(item.TypeCompass) || homeWorld.ConfigBool("region.wildCompasses", false)) &&
			wc.CheckWinCondition(assumed)

		fillable := locs.Filter(func(l *world.Location) bool {
			if l.HasItem() {
				return false
			}
			return l.CanFill(it, assumed, !canSkipAccess)
		})
		if fillable.Count() == 0 {
			return fmt.Errorf("no available locations for %s", it.FullName())
		}

		var chosen *world.Location
		if it.IsType(item.TypeMap) || it.IsType(item.TypeCompass) {
			n, err := helpers.GetRandomInt(0, fillable.Count()-1)
			if err != nil {
				return err
			}
			chosen = fillable.All()[n]
		} else {
			chosen = fillable.First()
		}
		if err := chosen.SetItem(it); err != nil {
			return fmt.Errorf("placing %s at %s: %w", it.FullName(), chosen.Name, err)
		}
	}
	return nil
}

// shuffleLocations returns a copy in random order.
func shuffleLocations(locs *world.LocationCollection) (*world.LocationCollection, error) {
	shuffled, err := helpers.FyShuffle(locs.All())
	if err != nil {
		return nil, err
	}
	return world.NewLocationCollection(shuffled...), nil
}

// fastFill places items in locations with no logic check, pop-order.
// Mirrors PHP Filler::fastFillItemsInLocations.
func fastFill(items []*item.Item, locs *world.LocationCollection) {
	idx := len(items) - 1
	for _, l := range locs.All() {
		if l.HasItem() {
			continue
		}
		if idx < 0 {
			break
		}
		_ = l.SetItem(items[idx])
		idx--
	}
}
