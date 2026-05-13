package world

import (
	"testing"

	"github.com/JackDoan/alttpr-go/internal/boss"
	"github.com/JackDoan/alttpr-go/internal/item"
)

func TestNewStandard_FullWorld(t *testing.T) {
	w := NewStandard(DefaultStandardOptions(), item.NewRegistry(), boss.NewRegistry())

	wantRegions := []string{
		"North East Light World", "North West Light World", "South Light World",
		"West Death Mountain", "East Death Mountain",
		"Escape", "Hyrule Castle Tower",
		"Eastern Palace", "Desert Palace", "Tower of Hera",
		"East Dark World Death Mountain", "West Dark World Death Mountain",
		"North East Dark World", "North West Dark World", "South Dark World",
		"Mire",
		"Palace of Darkness", "Swamp Palace", "Skull Woods",
		"Thieves Town", "Ice Palace", "Misery Mire", "Turtle Rock", "Ganons Tower",
		"Medallions", "Fountains",
	}
	for _, name := range wantRegions {
		if w.Region(name) == nil {
			t.Errorf("region %q missing", name)
		}
	}
	if got := len(w.AllRegions()); got != len(wantRegions) {
		t.Errorf("region count = %d, want %d", got, len(wantRegions))
	}

	totalLocs := w.Locations().Count()
	if totalLocs < 200 {
		t.Errorf("world has %d locations, expected 200+", totalLocs)
	}
	t.Logf("Standard world built: %d regions, %d locations", len(w.AllRegions()), totalLocs)

	// Quick logic sanity: with no items, win condition should be false.
	if w.WinCondition(item.NewCollection()) {
		t.Error("empty inventory should not win")
	}
	defeat, err := item.NewRegistry().Get("DefeatGanon", 0)
	if err != nil {
		t.Fatalf("get DefeatGanon: %v", err)
	}
	withWin := item.NewCollection(defeat)
	withWin.SetChecksForWorld(0)
	if !w.WinCondition(withWin) {
		t.Error("having DefeatGanon should satisfy win condition")
	}
}
