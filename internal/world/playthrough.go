package world

import (
	"fmt"

	"github.com/JackDoan/alttpr-go/internal/item"
)

// Playthrough is the spoiler-friendly chronological progression: each
// round groups locations whose items can be picked up given everything
// collected in previous rounds. Mirrors PHP PlaythroughService::getPlayThrough
// (forward-walk portion; the optimization pass that prunes non-required
// items is left for a future pass).
type Playthrough struct {
	LongestItemChain int                                `json:"longest_item_chain"`
	Equipped         map[string]string                  `json:"Equipped,omitempty"`
	Rounds           map[int]map[string]map[string]string `json:"rounds"`
	RegionsVisited   int                                `json:"regions_visited"`
}

// GetPlaythrough builds a playthrough record. Mirrors PHP forward-walk path.
func (w *World) GetPlaythrough() *Playthrough {
	pt := &Playthrough{Rounds: map[int]map[string]map[string]string{}}

	// Reset the filler's collected-locations cache so CollectableLocations
	// returns the whole world (the filler may have left it populated).
	w.ResetCollectedLocations()

	myItems := w.preCollected.Copy()
	myItems.SetChecksForWorld(w.id)

	visited := map[string]bool{}
	roundOrder := []map[*Location]bool{}
	longest := 1

	for {
		if len(roundOrder) > 0 && len(roundOrder[longest-1]) > 0 {
			longest++
		}
		if len(roundOrder) < longest {
			roundOrder = append(roundOrder, map[*Location]bool{})
		}

		available := w.CollectableLocations().Filter(func(l *Location) bool {
			if visited[l.FullName()] {
				return false
			}
			return l.CanAccess(myItems, w.locations)
		})

		if available.Count() == 0 {
			break
		}

		foundItems := []*item.Item{}
		for _, l := range available.All() {
			if visited[l.FullName()] {
				continue
			}
			if !l.HasItem() {
				visited[l.FullName()] = true
				continue
			}
			visited[l.FullName()] = true
			it := l.Item()
			foundItems = append(foundItems, it)

			// PHP excludes Keys (when not wild), Maps, Compasses, and RescueZelda
			// from the playthrough rounds — they're collected but not noteworthy.
			isExcluded := false
			if (w.ConfigBool("rom.genericKeys", false) || !w.ConfigBool("region.wildKeys", false)) && it.IsType(item.TypeKey) {
				isExcluded = true
			}
			if it.IsType(item.TypeMap) || it.IsType(item.TypeCompass) {
				isExcluded = true
			}
			if it.Name == "RescueZelda" {
				isExcluded = true
			}
			if l.Kind == KindTrade {
				isExcluded = true
			}
			if isExcluded {
				continue
			}
			roundOrder[longest-1][l] = true
		}
		if len(foundItems) == 0 {
			break
		}
		// Merge found items into myItems.
		for _, it := range foundItems {
			myItems.Add(it)
		}
	}

	// Build the round map: rounds[i][regionName][locationName] = itemName.
	for i, locs := range roundOrder {
		if len(locs) == 0 {
			continue
		}
		round := i + 1
		entry := map[string]map[string]string{}
		for l := range locs {
			rname := "(unknown)"
			if l.Region != nil {
				rname = l.Region.Name
			}
			if _, ok := entry[rname]; !ok {
				entry[rname] = map[string]string{}
			}
			itName := l.Item().Name
			if l.Item().Type == item.TypeAlias && l.Item().Target != nil {
				itName = l.Item().Target.Name
			}
			entry[rname][l.FullName()] = itName
		}
		pt.Rounds[round] = entry
	}
	pt.LongestItemChain = len(pt.Rounds)

	// Equipped: pre-collected items minus upgrades/events.
	if w.preCollected.Count() > 0 {
		pt.Equipped = map[string]string{}
		i := 0
		for _, it := range w.preCollected.Values() {
			if it.IsType(item.TypeUpgradeArrow) || it.IsType(item.TypeUpgradeBomb) ||
				it.IsType(item.TypeUpgradeHealth) || it.IsType(item.TypeEvent) {
				continue
			}
			i++
			pt.Equipped[fmt.Sprintf("Equipment Slot %d", i)] = it.Name
		}
	}

	// Count total visited regions (matches PHP's regions_visited sum).
	total := len(pt.Equipped)
	for _, r := range pt.Rounds {
		for _, locs := range r {
			total += len(locs)
		}
	}
	pt.RegionsVisited = total
	return pt
}
