package ui

import (
	"testing"

	"github.com/JackDoan/alttpr-go/cmd/alttpr-brick/input"
	"github.com/JackDoan/alttpr-go/internal/job"
)

// stubRenderer is a no-op Renderer; we don't assert on draw output here.
type stubRenderer struct{ w, h int }

func (s *stubRenderer) Bounds() (int, int)                       { return s.w, s.h }
func (s *stubRenderer) Clear(_ Color)                            {}
func (s *stubRenderer) FillRect(_, _, _, _ int, _ Color)         {}
func (s *stubRenderer) DrawText(_, _ int, _ string, _ Color)     {}

func TestMainCursorWraps(t *testing.T) {
	m := New(job.DefaultOptions())
	if m.mainCursor != 0 {
		t.Fatal("initial cursor not 0")
	}
	m.Step(input.BtnUp)
	if m.mainCursor != len(mainItems)-1 {
		t.Errorf("up from 0 should wrap to last; got %d", m.mainCursor)
	}
	m.Step(input.BtnDown)
	if m.mainCursor != 0 {
		t.Errorf("down from last should wrap to 0; got %d", m.mainCursor)
	}
}

func TestSelectGenerateReturnsAction(t *testing.T) {
	m := New(job.DefaultOptions())
	a := m.Step(input.BtnA)
	if a != ActionGenerate {
		t.Errorf("A on 'Generate Seed' should return ActionGenerate, got %v", a)
	}
	if m.Screen != ScreenGenerating {
		t.Errorf("screen should be ScreenGenerating, got %v", m.Screen)
	}
}

func TestSelectQuit(t *testing.T) {
	m := New(job.DefaultOptions())
	// "Quit" is the last entry.
	m.mainCursor = len(mainItems) - 1
	if mainItems[m.mainCursor] != "Quit" {
		t.Fatalf("expected last item to be Quit, got %q", mainItems[m.mainCursor])
	}
	if a := m.Step(input.BtnA); a != ActionQuit {
		t.Errorf("expected ActionQuit, got %v", a)
	}
}

func TestSelectViewSpoilersRequestsLoad(t *testing.T) {
	m := New(job.DefaultOptions())
	idx := indexOf(mainItems, "View Spoilers (debug)")
	if idx < 0 {
		t.Fatal("View Spoilers entry missing from mainItems")
	}
	m.mainCursor = idx
	a := m.Step(input.BtnA)
	if a != ActionLoadSpoilerList {
		t.Errorf("expected ActionLoadSpoilerList, got %v", a)
	}
	if m.Screen != ScreenSpoilerList {
		t.Errorf("expected ScreenSpoilerList, got %v", m.Screen)
	}
	if m.spoilerMode != modeRaw {
		t.Errorf("expected modeRaw for View Spoilers, got %v", m.spoilerMode)
	}
}

func TestSelectRevealSpoilerEntersRevealMode(t *testing.T) {
	m := New(job.DefaultOptions())
	idx := indexOf(mainItems, "Reveal Spoiler")
	if idx < 0 {
		t.Fatal("Reveal Spoiler entry missing")
	}
	m.mainCursor = idx
	a := m.Step(input.BtnA)
	if a != ActionLoadSpoilerList {
		t.Errorf("expected ActionLoadSpoilerList for reveal, got %v", a)
	}
	if m.spoilerMode != modeReveal {
		t.Errorf("expected modeReveal, got %v", m.spoilerMode)
	}
}

func TestSpoilerListPicksRevealAction(t *testing.T) {
	m := New(job.DefaultOptions())
	m.Screen = ScreenSpoilerList
	m.spoilerMode = modeReveal
	m.SetSpoilerList([]string{"a.json"})
	if a := m.Step(input.BtnA); a != ActionLoadReveal {
		t.Errorf("A in reveal mode should request ActionLoadReveal, got %v", a)
	}
}

func TestRevealToggleAndScroll(t *testing.T) {
	m := New(job.DefaultOptions())
	m.SetRevealEntries("test.json", []RevealEntry{
		{Item: "Bow", Location: "King Zora"},
		{Item: "Hookshot", Location: "Swamp Palace"},
		{Item: "PegasusBoots", Location: "Spiral Cave"},
	})
	if m.Screen != ScreenReveal {
		t.Fatalf("expected ScreenReveal, got %v", m.Screen)
	}
	if m.revealCategory != CatAll {
		t.Fatalf("expected to start in CatAll, got %v", m.revealCategory)
	}
	if m.revealEntries[0].Revealed {
		t.Fatal("entries should start hidden")
	}
	m.Step(input.BtnA) // reveal #0
	if !m.revealEntries[0].Revealed {
		t.Error("A should reveal current entry")
	}
	m.Step(input.BtnA) // toggle off
	if m.revealEntries[0].Revealed {
		t.Error("second A should un-reveal (toggle)")
	}
	m.Step(input.BtnDown)
	if m.revealCursor != 1 {
		t.Errorf("Down should move cursor to 1, got %d", m.revealCursor)
	}
	m.Step(input.BtnStart)
	for i, e := range m.revealEntries {
		if !e.Revealed {
			t.Errorf("Start should reveal all in current tab (entry %d still hidden)", i)
		}
	}
	m.Step(input.BtnB)
	if m.Screen != ScreenSpoilerList {
		t.Errorf("B should return to list, got %v", m.Screen)
	}
}

func TestClassifyItem(t *testing.T) {
	cases := map[string]RevealCategory{
		"Hookshot":           CatItems,
		"PegasusBoots":       CatItems,
		"Bottle":             CatItems,
		"ProgressiveBow":     CatItems,
		"KeyD3":              CatDungeon,
		"BigKeyP2":           CatDungeon,
		"MapD1":              CatDungeon,
		"CompassA2":          CatDungeon,
		"Crystal3":           CatPrizes,
		"PendantOfCourage":   CatPrizes,
		"PieceOfHeart":       CatHearts,
		"HeartContainer":     CatHearts,
		"BossHeartContainer": CatHearts,
		"FiftyRupees":        CatJunk,
		"ThreeBombs":         CatJunk,
		"TenArrows":          CatJunk,
	}
	for item, want := range cases {
		if got := ClassifyItem(item); got != want {
			t.Errorf("ClassifyItem(%q) = %v, want %v", item, got, want)
		}
	}
}

func TestRevealTabFiltersAndCycles(t *testing.T) {
	m := New(job.DefaultOptions())
	m.SetRevealEntries("test.json", []RevealEntry{
		{Item: "Hookshot", Location: "Swamp Palace"},        // CatItems
		{Item: "BigKeyP2", Location: "Eastern Palace"},      // CatDungeon
		{Item: "KeyD3", Location: "Skull Woods"},            // CatDungeon
		{Item: "PieceOfHeart", Location: "Pyramid Fairy"},   // CatHearts
		{Item: "FiftyRupees", Location: "Mire Shed"},        // CatJunk
		{Item: "Crystal3", Location: "Palace Of Darkness"},  // CatPrizes
	})
	if got := len(m.visibleReveal()); got != 6 {
		t.Errorf("CatAll should show all 6 entries; got %d", got)
	}
	// Right → Items
	m.Step(input.BtnRight)
	if m.revealCategory != CatItems {
		t.Errorf("Right should advance to CatItems, got %v", m.revealCategory)
	}
	if got := len(m.visibleReveal()); got != 1 {
		t.Errorf("CatItems should show 1 entry; got %d", got)
	}
	// Right → Dungeon (2 entries)
	m.Step(input.BtnRight)
	if got := len(m.visibleReveal()); got != 2 {
		t.Errorf("CatDungeon should show 2 entries; got %d", got)
	}
	// Reveal-all-in-tab only affects this tab.
	m.Step(input.BtnStart)
	dungeonRevealed := 0
	otherRevealed := 0
	for _, e := range m.revealEntries {
		if e.Category == CatDungeon {
			if e.Revealed {
				dungeonRevealed++
			}
		} else if e.Revealed {
			otherRevealed++
		}
	}
	if dungeonRevealed != 2 {
		t.Errorf("Start in Dungeon tab should reveal both dungeon entries; got %d", dungeonRevealed)
	}
	if otherRevealed != 0 {
		t.Errorf("Start in Dungeon tab leaked into other tabs (%d revealed)", otherRevealed)
	}
	// Left wraps back through tabs.
	m.Step(input.BtnLeft)
	m.Step(input.BtnLeft)
	if m.revealCategory != CatAll {
		t.Errorf("two Lefts from Dungeon should return to CatAll, got %v", m.revealCategory)
	}
}

func indexOf(ss []string, target string) int {
	for i, s := range ss {
		if s == target {
			return i
		}
	}
	return -1
}

func TestSpoilerListEmptyAllowsBack(t *testing.T) {
	m := New(job.DefaultOptions())
	m.Screen = ScreenSpoilerList
	if a := m.Step(input.BtnA); a != ActionNone {
		t.Errorf("A on empty list should be a no-op, got %v", a)
	}
	if a := m.Step(input.BtnB); a != ActionNone {
		t.Errorf("B should be a no-op return, got %v", a)
	}
	if m.Screen != ScreenMain {
		t.Errorf("expected ScreenMain after B, got %v", m.Screen)
	}
}

func TestSpoilerListNavigationAndOpen(t *testing.T) {
	m := New(job.DefaultOptions())
	m.Screen = ScreenSpoilerList
	m.SetSpoilerList([]string{"a.json", "b.json", "c.json"})
	if m.SelectedSpoiler() != "a.json" {
		t.Fatalf("initial selection should be a.json, got %q", m.SelectedSpoiler())
	}
	m.Step(input.BtnDown)
	if m.SelectedSpoiler() != "b.json" {
		t.Errorf("Down should move to b.json, got %q", m.SelectedSpoiler())
	}
	if a := m.Step(input.BtnA); a != ActionLoadSpoiler {
		t.Errorf("A should request ActionLoadSpoiler, got %v", a)
	}
}

func TestSpoilerViewScroll(t *testing.T) {
	m := New(job.DefaultOptions())
	lines := make([]string, 200)
	for i := range lines {
		lines[i] = "line"
	}
	m.SetSpoilerContent("test.json", lines)
	if m.Screen != ScreenSpoilerView {
		t.Fatalf("expected viewer screen, got %v", m.Screen)
	}
	if m.spoilerScroll != 0 {
		t.Errorf("initial scroll should be 0, got %d", m.spoilerScroll)
	}
	m.Step(input.BtnUp)
	if m.spoilerScroll != 0 {
		t.Errorf("Up at top should not underflow, got %d", m.spoilerScroll)
	}
	m.Step(input.BtnDown)
	if m.spoilerScroll != 1 {
		t.Errorf("Down should advance by 1, got %d", m.spoilerScroll)
	}
	m.Step(input.BtnRight)
	if m.spoilerScroll != 1+visibleTextRows {
		t.Errorf("Right should page down by %d, got %d", visibleTextRows, m.spoilerScroll)
	}
	m.Step(input.BtnB)
	if m.Screen != ScreenSpoilerList {
		t.Errorf("B should return to list, got %v", m.Screen)
	}
}

func TestEnterGameplayThenCycleGoal(t *testing.T) {
	m := New(job.DefaultOptions())
	m.mainCursor = 1
	m.Step(input.BtnA)
	if m.Screen != ScreenGameplay {
		t.Fatalf("expected ScreenGameplay, got %v", m.Screen)
	}
	// Cursor sits on "Goal". Cycle right twice (ganon -> fast_ganon -> dungeons).
	m.Step(input.BtnRight)
	m.Step(input.BtnRight)
	if m.Options.Goal != "dungeons" {
		t.Errorf("expected goal=dungeons after 2 right presses, got %q", m.Options.Goal)
	}
}

func TestCycleGoalLeftWrapsAcrossList(t *testing.T) {
	m := New(job.DefaultOptions())
	m.Screen = ScreenGameplay
	// Goal starts at index 0 ("ganon"); Left should wrap to the last entry.
	m.Step(input.BtnLeft)
	if m.Options.Goal != "completionist" {
		t.Errorf("expected wrap to completionist; got %q", m.Options.Goal)
	}
}

func TestCrystalsStepperClamps(t *testing.T) {
	m := New(job.DefaultOptions())
	m.Screen = ScreenGameplay
	// Move cursor down to "Crystals (Ganon)".
	rows := gameplayRows()
	for i, row := range rows {
		if row.stepper != nil && row.stepper.label == "Crystals (Ganon)" {
			m.gameplayCursor = i
			break
		}
	}
	// Default is 7; right wraps to 0.
	m.Step(input.BtnRight)
	if m.Options.CrystalsGanon != 0 {
		t.Errorf("expected wrap to 0; got %d", m.Options.CrystalsGanon)
	}
	// Left from 0 wraps back to 7.
	m.Step(input.BtnLeft)
	if m.Options.CrystalsGanon != 7 {
		t.Errorf("expected wrap to 7; got %d", m.Options.CrystalsGanon)
	}
}

func TestBackFromGameplayReturnsToMain(t *testing.T) {
	m := New(job.DefaultOptions())
	m.Screen = ScreenGameplay
	m.Step(input.BtnB)
	if m.Screen != ScreenMain {
		t.Errorf("expected ScreenMain after B, got %v", m.Screen)
	}
}

func TestSetResultMovesToScreenResult(t *testing.T) {
	m := New(job.DefaultOptions())
	m.Screen = ScreenGenerating
	m.SetResult(true, "Done", []string{"ROM Saved: /x.sfc"})
	if m.Screen != ScreenResult {
		t.Errorf("expected ScreenResult, got %v", m.Screen)
	}
	if m.statusTitle != "Done" || len(m.statusLines) != 1 {
		t.Errorf("status not stored: %+v", m)
	}
	// A returns to main.
	m.Step(input.BtnA)
	if m.Screen != ScreenMain {
		t.Errorf("expected ScreenMain after A on result, got %v", m.Screen)
	}
}

func TestRenderDoesNotPanicForEveryScreen(t *testing.T) {
	m := New(job.DefaultOptions())
	r := &stubRenderer{w: 1024, h: 768}
	screens := []Screen{
		ScreenMain, ScreenGameplay, ScreenCosmetic,
		ScreenGenerating, ScreenResult,
		ScreenSpoilerList, ScreenSpoilerView, ScreenReveal,
	}
	for _, s := range screens {
		m.Screen = s
		m.Render(r)
	}
}
