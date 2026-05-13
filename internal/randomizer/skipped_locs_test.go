package randomizer

import (
	"sort"
	"testing"

	"github.com/JackDoan/alttpr-go/internal/boss"
	"github.com/JackDoan/alttpr-go/internal/item"
	"github.com/JackDoan/alttpr-go/internal/world"
)

// TestDiagnoseSkippedLocations lists exactly which locations get skipped
// by the end-to-end byte-comparison test, and why.
func TestDiagnoseSkippedLocations(t *testing.T) {
	ir := item.NewRegistry()
	br := boss.NewRegistry()
	w := world.NewStandard(world.DefaultStandardOptions(), ir, br)
	r, _ := New([]*world.World{w}, ir, br)
	_ = r.Randomize()

	type skipInfo struct {
		Name    string
		Region  string
		Item    string
		Reason  string
		Kind    int
	}
	var skipped []skipInfo

	for _, region := range w.AllRegions() {
		for _, l := range region.Locations.All() {
			if !l.HasItem() {
				continue
			}
			reason := ""
			switch {
			case len(l.Address) == 0:
				reason = "no addresses"
			case l.Address[0] < 0:
				reason = "first address is -1 sentinel"
			default:
				itemBytes := l.Item().GetBytes()
				if len(itemBytes) == 0 {
					reason = "item has no bytes"
				} else if itemBytes[0] < 0 {
					reason = "item's first byte is -1 sentinel"
				}
			}
			if reason == "" {
				continue
			}
			skipped = append(skipped, skipInfo{
				Name: l.Name, Region: region.Name,
				Item: l.Item().Name, Reason: reason, Kind: int(l.Kind),
			})
		}
	}
	sort.Slice(skipped, func(i, j int) bool { return skipped[i].Name < skipped[j].Name })

	t.Logf("=== %d skipped locations ===", len(skipped))
	for _, s := range skipped {
		t.Logf("  %-32s [%s] kind=%d item=%-20s — %s",
			s.Name, s.Region, s.Kind, s.Item, s.Reason)
	}
}
