// Package ui is the on-device menu state machine. It owns no I/O — the
// caller drives it with input.Button events, calls Render once per frame
// with a Renderer that abstracts the framebuffer, and inspects Action()
// to decide when to invoke the randomizer.
//
// Layout of all screens is a vertical list of rows, with one row
// highlighted. Up/Down moves the cursor; Left/Right cycles the value of
// the focused row; A enters a screen or invokes an action; B goes back.
package ui

import (
	"fmt"

	"github.com/JackDoan/alttpr-go/cmd/alttpr-brick/input"
	"github.com/JackDoan/alttpr-go/internal/job"
)

// Screen identifies the active screen of the state machine.
type Screen int

const (
	ScreenMain Screen = iota
	ScreenGameplay
	ScreenCosmetic
	ScreenGenerating
	ScreenResult
	ScreenSpoilerList
	ScreenSpoilerView
	ScreenReveal
)

// spoilerMode tracks whether the file picker was entered from "View
// Spoilers" (raw JSON dump) or "Reveal Spoiler" (per-entry reveal).
type spoilerMode int

const (
	modeRaw spoilerMode = iota
	modeReveal
)

// Action is what the caller should do after Step returns. ActionNone means
// "no transition; just keep polling input and re-rendering".
type Action int

const (
	ActionNone Action = iota
	ActionGenerate
	ActionQuit
	// ActionLoadSpoilerList: the host should glob the output directory and
	// call SetSpoilerList with the resulting basenames (newest first).
	ActionLoadSpoilerList
	// ActionLoadSpoiler: the host should read the file at the model's
	// SelectedSpoiler() path and call SetSpoilerContent with its lines.
	ActionLoadSpoiler
	// ActionLoadReveal: the host should parse the JSON file at
	// SelectedSpoiler(), build the per-item entries, and call
	// SetRevealEntries.
	ActionLoadReveal
)

// Color is a renderer-agnostic RGB triple. The fb package's Color has the
// same shape; main.go converts when calling.
type Color struct{ R, G, B uint8 }

// Renderer is the minimal surface the menu draws against. The framebuffer
// implements this; tests stub it.
type Renderer interface {
	Bounds() (w, h int)
	Clear(c Color)
	FillRect(x, y, w, h int, c Color)
	DrawText(x, y int, text string, fg Color)
}

// Theme is the small palette the menu uses.
type Theme struct {
	BG, FG, FGDim, Highlight, HighlightFG, Accent, Err Color
}

func DefaultTheme() Theme {
	return Theme{
		BG:          Color{0x10, 0x14, 0x1A},
		FG:          Color{0xE0, 0xE6, 0xEE},
		FGDim:       Color{0x90, 0x96, 0x9E},
		Highlight:   Color{0x3F, 0x96, 0x59},
		HighlightFG: Color{0xFF, 0xFF, 0xFF},
		Accent:      Color{0x70, 0xC0, 0xA0},
		Err:         Color{0xE0, 0x60, 0x60},
	}
}

// Model is the full state of the menu.
type Model struct {
	Theme   Theme
	Screen  Screen
	Options job.Options

	// Per-screen cursor positions.
	mainCursor     int
	gameplayCursor int
	cosmeticCursor int

	// Status copy for ScreenGenerating / ScreenResult.
	statusTitle string
	statusLines []string
	statusOK    bool

	// Spoiler browser state. spoilerFiles holds basenames (newest first);
	// host code populates it via SetSpoilerList. spoilerLines is the
	// currently-open file's contents, populated via SetSpoilerContent.
	spoilerFiles      []string
	spoilerListCursor int
	spoilerListScroll int
	spoilerMode       spoilerMode
	spoilerTitle      string
	spoilerLines      []string
	spoilerScroll     int

	// Reveal-mode state. Each entry shows an item name; A flips Revealed
	// from false → true (one-shot — re-pressing A toggles it back off so
	// you can hide it again if you change your mind). revealCursor /
	// revealScroll are indices into the *visible* (filtered) slice; the
	// entries themselves stay in their original sort.
	revealEntries  []RevealEntry
	revealCursor   int
	revealScroll   int
	revealTitle    string
	revealCategory RevealCategory
}

// RevealEntry is one row in reveal mode: the item name (shown) and the
// location where it was placed (hidden until Revealed = true).
type RevealEntry struct {
	Item     string
	Location string
	Revealed bool
	Category RevealCategory
}

// RevealCategory groups items into the tabs the user can scroll through
// in reveal mode. CatAll is a virtual category that matches every entry.
type RevealCategory int

const (
	CatAll RevealCategory = iota
	CatItems
	CatDungeon
	CatPrizes
	CatHearts
	CatJunk
)

// revealTabs is the ordered list of tabs to render. CatAll is the default
// landing tab so the user sees everything before filtering.
var revealTabs = []struct {
	label string
	cat   RevealCategory
}{
	{"All", CatAll},
	{"Items", CatItems},
	{"Dungeon", CatDungeon},
	{"Prizes", CatPrizes},
	{"Hearts", CatHearts},
	{"Junk", CatJunk},
}

// New builds a model with the given last-used options.
func New(opts job.Options) *Model {
	return &Model{
		Theme:   DefaultTheme(),
		Screen:  ScreenMain,
		Options: opts,
	}
}

// Step consumes one button event and returns an Action describing what
// the caller should do next.
func (m *Model) Step(btn input.Button) Action {
	switch m.Screen {
	case ScreenMain:
		return m.stepMain(btn)
	case ScreenGameplay:
		return m.stepGameplay(btn)
	case ScreenCosmetic:
		return m.stepCosmetic(btn)
	case ScreenGenerating:
		// No input handled while generating; the job is async. Caller calls
		// SetResult when done, which transitions us to ScreenResult.
		return ActionNone
	case ScreenResult:
		switch btn {
		case input.BtnA, input.BtnB, input.BtnStart:
			m.Screen = ScreenMain
		}
		return ActionNone
	case ScreenSpoilerList:
		return m.stepSpoilerList(btn)
	case ScreenSpoilerView:
		return m.stepSpoilerView(btn)
	case ScreenReveal:
		return m.stepReveal(btn)
	}
	return ActionNone
}

// --- main screen ---------------------------------------------------------

var mainItems = []string{
	"Generate Seed",
	"Gameplay Settings",
	"Cosmetic Settings",
	"Reveal Spoiler",
	"View Spoilers (debug)",
	"Quit",
}

func (m *Model) stepMain(b input.Button) Action {
	switch b {
	case input.BtnUp:
		m.mainCursor = wrap(m.mainCursor-1, len(mainItems))
	case input.BtnDown:
		m.mainCursor = wrap(m.mainCursor+1, len(mainItems))
	case input.BtnA, input.BtnStart:
		switch m.mainCursor {
		case 0:
			m.Screen = ScreenGenerating
			m.statusTitle = "Generating seed..."
			m.statusLines = nil
			return ActionGenerate
		case 1:
			m.Screen = ScreenGameplay
		case 2:
			m.Screen = ScreenCosmetic
		case 3:
			m.Screen = ScreenSpoilerList
			m.spoilerMode = modeReveal
			m.spoilerListCursor = 0
			m.spoilerListScroll = 0
			return ActionLoadSpoilerList
		case 4:
			m.Screen = ScreenSpoilerList
			m.spoilerMode = modeRaw
			m.spoilerListCursor = 0
			m.spoilerListScroll = 0
			return ActionLoadSpoilerList
		case 5:
			return ActionQuit
		}
	}
	return ActionNone
}

// --- gameplay screen -----------------------------------------------------

type cycler struct {
	label  string
	values []string
	get    func(o *job.Options) string
	set    func(o *job.Options, v string)
}

type stepper struct {
	label string
	min   int
	max   int
	get   func(o *job.Options) int
	set   func(o *job.Options, v int)
}

type gameplayRow struct {
	cycler  *cycler
	stepper *stepper
}

func gameplayRows() []gameplayRow {
	return []gameplayRow{
		{cycler: &cycler{"Goal", []string{"ganon", "fast_ganon", "dungeons", "pedestal", "ganonhunt", "triforce-hunt", "completionist"},
			func(o *job.Options) string { return o.Goal },
			func(o *job.Options, v string) { o.Goal = v }}},
		{cycler: &cycler{"State", []string{"standard", "open"},
			func(o *job.Options) string { return o.State },
			func(o *job.Options, v string) { o.State = v }}},
		{cycler: &cycler{"Weapons", []string{"randomized", "swordless", "assured", "vanilla"},
			func(o *job.Options) string { return o.Weapons },
			func(o *job.Options, v string) { o.Weapons = v }}},
		{cycler: &cycler{"Item Pool", []string{"normal", "hard", "expert", "superexpert", "crowd_control"},
			func(o *job.Options) string { return o.ItemPool },
			func(o *job.Options, v string) { o.ItemPool = v }}},
		{cycler: &cycler{"Item Placement", []string{"basic", "advanced"},
			func(o *job.Options) string { return o.ItemPlacement },
			func(o *job.Options, v string) { o.ItemPlacement = v }}},
		{cycler: &cycler{"Accessibility", []string{"item", "locations", "none"},
			func(o *job.Options) string { return o.Accessibility },
			func(o *job.Options, v string) { o.Accessibility = v }}},
		{stepper: &stepper{"Crystals (Ganon)", 0, 7,
			func(o *job.Options) int { return o.CrystalsGanon },
			func(o *job.Options, v int) { o.CrystalsGanon = v }}},
		{stepper: &stepper{"Crystals (Tower)", 0, 7,
			func(o *job.Options) int { return o.CrystalsTower },
			func(o *job.Options, v int) { o.CrystalsTower = v }}},
	}
}

func (m *Model) stepGameplay(b input.Button) Action {
	rows := gameplayRows()
	switch b {
	case input.BtnUp:
		m.gameplayCursor = wrap(m.gameplayCursor-1, len(rows))
	case input.BtnDown:
		m.gameplayCursor = wrap(m.gameplayCursor+1, len(rows))
	case input.BtnLeft:
		cycleRow(&m.Options, rows[m.gameplayCursor], -1)
	case input.BtnRight:
		cycleRow(&m.Options, rows[m.gameplayCursor], +1)
	case input.BtnB:
		m.Screen = ScreenMain
	case input.BtnA:
		m.Screen = ScreenMain
	}
	return ActionNone
}

// --- cosmetic screen -----------------------------------------------------

func cosmeticRows() []gameplayRow {
	return []gameplayRow{
		{cycler: &cycler{"Heart Color", []string{"red", "blue", "green", "yellow", "random"},
			func(o *job.Options) string { return o.HeartColor },
			func(o *job.Options, v string) { o.HeartColor = v }}},
		{cycler: &cycler{"Heart Beep", []string{"off", "normal", "half", "quarter", "double"},
			func(o *job.Options) string { return o.HeartBeep },
			func(o *job.Options, v string) { o.HeartBeep = v }}},
		{cycler: &cycler{"Menu Speed", []string{"slow", "normal", "fast", "instant"},
			func(o *job.Options) string { return o.MenuSpeed },
			func(o *job.Options, v string) { o.MenuSpeed = v }}},
		{cycler: &cycler{"Quickswap", []string{"false", "true"},
			func(o *job.Options) string { return o.Quickswap },
			func(o *job.Options, v string) { o.Quickswap = v }}},
		{cycler: &cycler{"Mute Music", []string{"off", "on"},
			func(o *job.Options) string {
				if o.NoMusic {
					return "on"
				}
				return "off"
			},
			func(o *job.Options, v string) { o.NoMusic = v == "on" }}},
	}
}

func (m *Model) stepCosmetic(b input.Button) Action {
	rows := cosmeticRows()
	switch b {
	case input.BtnUp:
		m.cosmeticCursor = wrap(m.cosmeticCursor-1, len(rows))
	case input.BtnDown:
		m.cosmeticCursor = wrap(m.cosmeticCursor+1, len(rows))
	case input.BtnLeft:
		cycleRow(&m.Options, rows[m.cosmeticCursor], -1)
	case input.BtnRight:
		cycleRow(&m.Options, rows[m.cosmeticCursor], +1)
	case input.BtnA, input.BtnB:
		m.Screen = ScreenMain
	}
	return ActionNone
}

func cycleRow(o *job.Options, row gameplayRow, delta int) {
	if row.cycler != nil {
		cur := row.cycler.get(o)
		idx := 0
		for i, v := range row.cycler.values {
			if v == cur {
				idx = i
				break
			}
		}
		idx = wrap(idx+delta, len(row.cycler.values))
		row.cycler.set(o, row.cycler.values[idx])
		return
	}
	if row.stepper != nil {
		v := row.stepper.get(o) + delta
		if v < row.stepper.min {
			v = row.stepper.max
		}
		if v > row.stepper.max {
			v = row.stepper.min
		}
		row.stepper.set(o, v)
	}
}

// --- spoiler screens -----------------------------------------------------

// visibleListRows is the number of rows that fit on a list/text screen
// (screen height - header - footer) / rowH. Used for scroll math.
const visibleListRows = 24

// visibleTextRows is the number of rows that fit on the spoiler viewer.
// Slightly fewer than list rows since the viewer also shows the position
// indicator in the footer.
const visibleTextRows = 24

// maxLineChars caps how many characters of each spoiler line we render
// per row. Past this, the row is rendered with a trailing ellipsis so
// the viewer never has to deal with horizontal overflow.
const maxLineChars = 80

func (m *Model) stepSpoilerList(b input.Button) Action {
	n := len(m.spoilerFiles)
	switch b {
	case input.BtnUp:
		if n > 0 {
			m.spoilerListCursor = wrap(m.spoilerListCursor-1, n)
			m.adjustListScroll()
		}
	case input.BtnDown:
		if n > 0 {
			m.spoilerListCursor = wrap(m.spoilerListCursor+1, n)
			m.adjustListScroll()
		}
	case input.BtnLeft:
		m.spoilerListCursor -= visibleListRows
		if m.spoilerListCursor < 0 {
			m.spoilerListCursor = 0
		}
		m.adjustListScroll()
	case input.BtnRight:
		if n > 0 {
			m.spoilerListCursor += visibleListRows
			if m.spoilerListCursor >= n {
				m.spoilerListCursor = n - 1
			}
		}
		m.adjustListScroll()
	case input.BtnA, input.BtnStart:
		if n > 0 {
			if m.spoilerMode == modeReveal {
				return ActionLoadReveal
			}
			return ActionLoadSpoiler
		}
	case input.BtnB:
		m.Screen = ScreenMain
	}
	return ActionNone
}

func (m *Model) adjustListScroll() {
	if m.spoilerListCursor < m.spoilerListScroll {
		m.spoilerListScroll = m.spoilerListCursor
	}
	if m.spoilerListCursor >= m.spoilerListScroll+visibleListRows {
		m.spoilerListScroll = m.spoilerListCursor - visibleListRows + 1
	}
}

func (m *Model) stepSpoilerView(b input.Button) Action {
	n := len(m.spoilerLines)
	switch b {
	case input.BtnUp:
		if m.spoilerScroll > 0 {
			m.spoilerScroll--
		}
	case input.BtnDown:
		if m.spoilerScroll+visibleTextRows < n {
			m.spoilerScroll++
		}
	case input.BtnLeft:
		m.spoilerScroll -= visibleTextRows
		if m.spoilerScroll < 0 {
			m.spoilerScroll = 0
		}
	case input.BtnRight:
		m.spoilerScroll += visibleTextRows
		if m.spoilerScroll+visibleTextRows > n {
			m.spoilerScroll = n - visibleTextRows
		}
		if m.spoilerScroll < 0 {
			m.spoilerScroll = 0
		}
	case input.BtnB:
		m.Screen = ScreenSpoilerList
	}
	return ActionNone
}

// SetSpoilerList is called by the host with the basenames of *.json files
// in the output directory, ordered newest-first.
func (m *Model) SetSpoilerList(files []string) {
	m.spoilerFiles = files
	if m.spoilerListCursor >= len(files) {
		m.spoilerListCursor = 0
	}
	m.spoilerListScroll = 0
}

// SelectedSpoiler returns the basename the user has the cursor on, or "".
func (m *Model) SelectedSpoiler() string {
	if m.spoilerListCursor < 0 || m.spoilerListCursor >= len(m.spoilerFiles) {
		return ""
	}
	return m.spoilerFiles[m.spoilerListCursor]
}

// SetSpoilerContent is called by the host after reading the selected file.
// lines should already be split by '\n'. Switches to the viewer screen.
func (m *Model) SetSpoilerContent(title string, lines []string) {
	m.spoilerTitle = title
	m.spoilerLines = lines
	m.spoilerScroll = 0
	m.Screen = ScreenSpoilerView
}

// SetRevealEntries is called by the host after parsing the spoiler JSON
// in reveal mode. entries should already be sorted in display order;
// each starts with Revealed=false. If an entry's Category is zero
// (CatAll), the model classifies it via ClassifyItem. Switches to
// ScreenReveal.
func (m *Model) SetRevealEntries(title string, entries []RevealEntry) {
	m.revealTitle = title
	m.revealEntries = entries
	for i := range m.revealEntries {
		if m.revealEntries[i].Category == CatAll {
			m.revealEntries[i].Category = ClassifyItem(m.revealEntries[i].Item)
		}
	}
	m.revealCursor = 0
	m.revealScroll = 0
	m.revealCategory = CatAll
	m.Screen = ScreenReveal
}

// visibleReveal returns the indices into revealEntries that belong to
// the currently-active tab.
func (m *Model) visibleReveal() []int {
	idxs := make([]int, 0, len(m.revealEntries))
	for i, e := range m.revealEntries {
		if m.revealCategory == CatAll || e.Category == m.revealCategory {
			idxs = append(idxs, i)
		}
	}
	return idxs
}

func (m *Model) stepReveal(b input.Button) Action {
	visible := m.visibleReveal()
	n := len(visible)
	switch b {
	case input.BtnUp:
		if n > 0 {
			m.revealCursor = wrap(m.revealCursor-1, n)
			m.adjustRevealScroll(n)
		}
	case input.BtnDown:
		if n > 0 {
			m.revealCursor = wrap(m.revealCursor+1, n)
			m.adjustRevealScroll(n)
		}
	case input.BtnLeft:
		m.cycleRevealTab(-1)
	case input.BtnRight:
		m.cycleRevealTab(+1)
	case input.BtnL:
		// Shoulder buttons: page up/down within the active tab.
		m.revealCursor -= visibleListRows
		if m.revealCursor < 0 {
			m.revealCursor = 0
		}
		m.adjustRevealScroll(n)
	case input.BtnR:
		if n > 0 {
			m.revealCursor += visibleListRows
			if m.revealCursor >= n {
				m.revealCursor = n - 1
			}
		}
		m.adjustRevealScroll(n)
	case input.BtnA:
		if m.revealCursor >= 0 && m.revealCursor < n {
			idx := visible[m.revealCursor]
			m.revealEntries[idx].Revealed = !m.revealEntries[idx].Revealed
		}
	case input.BtnStart:
		// Reveal everything in the current tab (the rest stays hidden).
		for _, i := range visible {
			m.revealEntries[i].Revealed = true
		}
	case input.BtnB:
		m.Screen = ScreenSpoilerList
	}
	return ActionNone
}

func (m *Model) cycleRevealTab(delta int) {
	idx := 0
	for i, t := range revealTabs {
		if t.cat == m.revealCategory {
			idx = i
			break
		}
	}
	idx = wrap(idx+delta, len(revealTabs))
	m.revealCategory = revealTabs[idx].cat
	m.revealCursor = 0
	m.revealScroll = 0
}

func (m *Model) adjustRevealScroll(n int) {
	if n == 0 {
		m.revealScroll = 0
		return
	}
	if m.revealCursor < m.revealScroll {
		m.revealScroll = m.revealCursor
	}
	if m.revealCursor >= m.revealScroll+visibleListRows {
		m.revealScroll = m.revealCursor - visibleListRows + 1
	}
}

// ClassifyItem returns the tab a spoiler item belongs to. Exposed so the
// host (and tests) can pre-classify if they want; SetRevealEntries also
// calls it for any entry left as CatAll.
func ClassifyItem(item string) RevealCategory {
	switch item {
	case "PieceOfHeart", "HeartContainer", "BossHeartContainer":
		return CatHearts
	}
	switch {
	case startsWith(item, "Key"), startsWith(item, "BigKey"),
		startsWith(item, "Map"), startsWith(item, "Compass"):
		return CatDungeon
	case startsWith(item, "Crystal"), startsWith(item, "PendantOf"):
		return CatPrizes
	case isJunkItem(item):
		return CatJunk
	}
	return CatItems
}

func startsWith(s, prefix string) bool {
	return len(s) >= len(prefix) && s[:len(prefix)] == prefix
}

func isJunkItem(item string) bool {
	switch item {
	case "OneRupee", "FiveRupees", "TwentyRupees", "FiftyRupees",
		"OneHundredRupees", "ThreeHundredRupees",
		"Arrow", "TenArrows",
		"ThreeBombs", "TenBombs",
		"Rupoor":
		return true
	}
	return false
}

// SetResult is called when the background job finishes. ok=true on success.
func (m *Model) SetResult(ok bool, title string, lines []string) {
	m.Screen = ScreenResult
	m.statusTitle = title
	m.statusLines = lines
	m.statusOK = ok
}

// --- render --------------------------------------------------------------

const (
	pad      = 16
	rowH     = 24
	headerH  = 32
	footerH  = 24
	listGapY = 8
)

// Render draws the current screen.
func (m *Model) Render(r Renderer) {
	r.Clear(m.Theme.BG)
	switch m.Screen {
	case ScreenMain:
		m.renderList(r, "alttpr — Trimui Brick", mainItems, m.mainCursor, mainFooter)
	case ScreenGameplay:
		m.renderRowList(r, "Gameplay Settings", gameplayRows(), m.gameplayCursor, settingsFooter)
	case ScreenCosmetic:
		m.renderRowList(r, "Cosmetic Settings", cosmeticRows(), m.cosmeticCursor, settingsFooter)
	case ScreenGenerating:
		m.renderStatus(r, m.statusTitle, []string{"Please wait — this can take a few seconds."}, m.Theme.FG)
	case ScreenResult:
		fg := m.Theme.Accent
		if !m.statusOK {
			fg = m.Theme.Err
		}
		m.renderStatus(r, m.statusTitle, m.statusLines, fg)
	case ScreenSpoilerList:
		m.renderSpoilerList(r)
	case ScreenSpoilerView:
		m.renderSpoilerView(r)
	case ScreenReveal:
		m.renderReveal(r)
	}
}

func (m *Model) renderReveal(r Renderer) {
	w, h := r.Bounds()

	// Title row (filename).
	r.DrawText(pad, pad, truncate(m.revealTitle, maxLineChars), m.Theme.FG)

	// Tab strip on the second row. Each tab is a fixed-width slot so the
	// active highlight band lines up regardless of label length.
	tabY := pad + rowH
	tabW := (w - 2*pad) / len(revealTabs)
	for i, t := range revealTabs {
		x := pad + i*tabW
		fg := m.Theme.FGDim
		if t.cat == m.revealCategory {
			r.FillRect(x, tabY-2, tabW, rowH, m.Theme.Highlight)
			fg = m.Theme.HighlightFG
		}
		// Center the label inside the slot.
		labelX := x + (tabW-TextWidthRunes(t.label))/2
		r.DrawText(labelX, tabY+4, t.label, fg)
	}
	dividerY := tabY + rowH + 2
	r.FillRect(pad, dividerY, w-2*pad, 2, m.Theme.FGDim)

	visible := m.visibleReveal()
	if len(visible) == 0 {
		r.DrawText(pad, dividerY+rowH, "(no entries in this tab)", m.Theme.FGDim)
		r.DrawText(pad, h-footerH, "Left/Right: switch tab   B: back", m.Theme.FGDim)
		return
	}

	listTop := dividerY + listGapY
	// Recompute how many rows actually fit between the divider and footer.
	maxRows := (h - footerH - listTop) / rowH
	if maxRows < 1 {
		maxRows = 1
	}
	end := m.revealScroll + maxRows
	if end > len(visible) {
		end = len(visible)
	}

	y := listTop
	for i := m.revealScroll; i < end; i++ {
		e := m.revealEntries[visible[i]]
		loc := "???"
		locColor := m.Theme.FGDim
		if e.Revealed {
			loc = e.Location
			locColor = m.Theme.Accent
		}
		label := truncate(e.Item, maxLineChars/2-2)
		valueX := pad + w/2
		if i == m.revealCursor {
			r.FillRect(pad-4, y-2, w-2*pad+8, rowH, m.Theme.Highlight)
			r.DrawText(pad, y+4, label, m.Theme.HighlightFG)
			r.DrawText(valueX, y+4, truncate(loc, maxLineChars/2-2), m.Theme.HighlightFG)
		} else {
			r.DrawText(pad, y+4, label, m.Theme.FG)
			r.DrawText(valueX, y+4, truncate(loc, maxLineChars/2-2), locColor)
		}
		y += rowH
	}

	// Footer: per-tab and total reveal counts.
	revealedTab, revealedAll := 0, 0
	for _, idx := range visible {
		if m.revealEntries[idx].Revealed {
			revealedTab++
		}
	}
	for _, e := range m.revealEntries {
		if e.Revealed {
			revealedAll++
		}
	}
	footer := fmt.Sprintf("A: reveal   L/R: page   Start: reveal tab   B: back   tab %d/%d   all %d/%d",
		revealedTab, len(visible), revealedAll, len(m.revealEntries))
	r.DrawText(pad, h-footerH, truncate(footer, maxLineChars), m.Theme.FGDim)
}

// TextWidthRunes returns the pixel width a string will occupy when drawn.
// Lives in the ui package so renderers don't have to import fb just to
// center-align labels.
func TextWidthRunes(s string) int { return len(s) * 8 }

func (m *Model) renderSpoilerList(r Renderer) {
	w, h := r.Bounds()
	r.DrawText(pad, pad, "Spoilers", m.Theme.FG)
	r.FillRect(pad, pad+headerH-listGapY-2, w-2*pad, 2, m.Theme.FGDim)

	if len(m.spoilerFiles) == 0 {
		r.DrawText(pad, pad+headerH+rowH, "(no spoilers found yet)", m.Theme.FGDim)
		r.DrawText(pad, h-footerH, "B: back", m.Theme.FGDim)
		return
	}

	y := pad + headerH
	end := m.spoilerListScroll + visibleListRows
	if end > len(m.spoilerFiles) {
		end = len(m.spoilerFiles)
	}
	for i := m.spoilerListScroll; i < end; i++ {
		label := m.spoilerFiles[i]
		if i == m.spoilerListCursor {
			r.FillRect(pad-4, y-2, w-2*pad+8, rowH, m.Theme.Highlight)
			r.DrawText(pad, y+4, truncate(label, maxLineChars), m.Theme.HighlightFG)
		} else {
			r.DrawText(pad, y+4, truncate(label, maxLineChars), m.Theme.FG)
		}
		y += rowH
	}

	footer := fmt.Sprintf("A: open   B: back   %d/%d", m.spoilerListCursor+1, len(m.spoilerFiles))
	r.DrawText(pad, h-footerH, footer, m.Theme.FGDim)
}

func (m *Model) renderSpoilerView(r Renderer) {
	w, h := r.Bounds()
	r.DrawText(pad, pad, truncate(m.spoilerTitle, maxLineChars), m.Theme.FG)
	r.FillRect(pad, pad+headerH-listGapY-2, w-2*pad, 2, m.Theme.FGDim)

	y := pad + headerH
	end := m.spoilerScroll + visibleTextRows
	if end > len(m.spoilerLines) {
		end = len(m.spoilerLines)
	}
	for i := m.spoilerScroll; i < end; i++ {
		r.DrawText(pad, y, truncate(m.spoilerLines[i], maxLineChars), m.Theme.FG)
		y += rowH
	}

	footer := "B: back   Up/Down: scroll   Left/Right: page"
	if len(m.spoilerLines) > 0 {
		last := m.spoilerScroll + visibleTextRows
		if last > len(m.spoilerLines) {
			last = len(m.spoilerLines)
		}
		footer = fmt.Sprintf("%s   %d-%d/%d", footer, m.spoilerScroll+1, last, len(m.spoilerLines))
	}
	r.DrawText(pad, h-footerH, footer, m.Theme.FGDim)
}

func truncate(s string, max int) string {
	if len(s) <= max {
		return s
	}
	if max < 4 {
		return s[:max]
	}
	return s[:max-3] + "..."
}

var mainFooter = "A: select   B: back   Start: quick-gen"
var settingsFooter = "Left/Right: cycle   B: back"

func (m *Model) renderList(r Renderer, title string, items []string, cursor int, footer string) {
	w, _ := r.Bounds()
	r.DrawText(pad, pad, title, m.Theme.FG)
	r.FillRect(pad, pad+headerH-listGapY-2, w-2*pad, 2, m.Theme.FGDim)

	y := pad + headerH
	for i, label := range items {
		if i == cursor {
			r.FillRect(pad-4, y-2, w-2*pad+8, rowH, m.Theme.Highlight)
			r.DrawText(pad, y+4, label, m.Theme.HighlightFG)
		} else {
			r.DrawText(pad, y+4, label, m.Theme.FG)
		}
		y += rowH
	}

	_, h := r.Bounds()
	r.DrawText(pad, h-footerH, footer, m.Theme.FGDim)
}

func (m *Model) renderRowList(r Renderer, title string, rows []gameplayRow, cursor int, footer string) {
	w, h := r.Bounds()
	r.DrawText(pad, pad, title, m.Theme.FG)
	r.FillRect(pad, pad+headerH-listGapY-2, w-2*pad, 2, m.Theme.FGDim)

	y := pad + headerH
	for i, row := range rows {
		label, value := rowText(&m.Options, row)
		if i == cursor {
			r.FillRect(pad-4, y-2, w-2*pad+8, rowH, m.Theme.Highlight)
			r.DrawText(pad, y+4, label, m.Theme.HighlightFG)
			r.DrawText(pad+w/2, y+4, "< "+value+" >", m.Theme.HighlightFG)
		} else {
			r.DrawText(pad, y+4, label, m.Theme.FG)
			r.DrawText(pad+w/2, y+4, value, m.Theme.FGDim)
		}
		y += rowH
	}

	r.DrawText(pad, h-footerH, footer, m.Theme.FGDim)
}

func (m *Model) renderStatus(r Renderer, title string, lines []string, fg Color) {
	r.DrawText(pad, pad, title, fg)
	y := pad + headerH
	for _, ln := range lines {
		r.DrawText(pad, y, ln, m.Theme.FG)
		y += rowH
	}
	_, h := r.Bounds()
	r.DrawText(pad, h-footerH, "A: continue", m.Theme.FGDim)
}

func rowText(o *job.Options, row gameplayRow) (string, string) {
	if row.cycler != nil {
		return row.cycler.label, row.cycler.get(o)
	}
	if row.stepper != nil {
		return row.stepper.label, fmt.Sprintf("%d", row.stepper.get(o))
	}
	return "", ""
}

func wrap(v, n int) int {
	if n <= 0 {
		return 0
	}
	v %= n
	if v < 0 {
		v += n
	}
	return v
}
