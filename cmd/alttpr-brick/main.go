// Command alttpr-brick is the Trimui Brick on-device harness for the
// alttpr randomizer. It draws a button-driven menu directly to /dev/fb0,
// reads gamepad events from /dev/input/event*, and runs the randomizer
// in-process (no subprocess) when the user picks "Generate Seed".
//
// A --host-test mode runs a single randomization with hardcoded defaults
// and exits — useful for smoke-testing the job package on the desktop
// without the framebuffer.
package main

import (
	"encoding/json"
	"flag"
	"fmt"
	"os"
	"path/filepath"
	"sort"
	"strings"
	"time"

	"github.com/JackDoan/alttpr-go/cmd/alttpr-brick/fb"
	"github.com/JackDoan/alttpr-go/cmd/alttpr-brick/input"
	"github.com/JackDoan/alttpr-go/cmd/alttpr-brick/ui"
	"github.com/JackDoan/alttpr-go/internal/job"
)

// parseRevealEntries reads a spoiler JSON and flattens the regions map
// into a sorted list of (item, location) pairs. Sorted by item name so
// duplicates (e.g. ten "TwentyRupees" chests) cluster together — the
// user can reveal them one at a time without losing track.
func parseRevealEntries(path string) ([]ui.RevealEntry, error) {
	data, err := os.ReadFile(path)
	if err != nil {
		return nil, err
	}
	// The spoiler also has top-level Bosses/Equipped/etc, but for the
	// reveal flow we only care about location→item; everything else is
	// noise. Use a partial schema.
	var spoiler struct {
		Regions map[string]map[string]string `json:"regions"`
	}
	if err := json.Unmarshal(data, &spoiler); err != nil {
		return nil, err
	}
	var entries []ui.RevealEntry
	for _, locs := range spoiler.Regions {
		for loc, item := range locs {
			entries = append(entries, ui.RevealEntry{Item: item, Location: loc})
		}
	}
	sort.SliceStable(entries, func(i, j int) bool {
		if entries[i].Item != entries[j].Item {
			return entries[i].Item < entries[j].Item
		}
		return entries[i].Location < entries[j].Location
	})
	return entries, nil
}

// listSpoilers returns the basenames of *.json files in dir, newest-first
// by mtime. Missing dir returns an empty slice (not an error).
func listSpoilers(dir string) []string {
	entries, err := os.ReadDir(dir)
	if err != nil {
		return nil
	}
	type entry struct {
		name  string
		mtime int64
	}
	out := make([]entry, 0, len(entries))
	for _, e := range entries {
		if e.IsDir() {
			continue
		}
		name := e.Name()
		if !strings.HasSuffix(name, ".json") {
			continue
		}
		info, err := e.Info()
		if err != nil {
			continue
		}
		out = append(out, entry{name: name, mtime: info.ModTime().UnixNano()})
	}
	sort.Slice(out, func(i, j int) bool { return out[i].mtime > out[j].mtime })
	names := make([]string, len(out))
	for i := range out {
		names[i] = out[i].name
	}
	return names
}

// readSpoiler reads the file and returns its lines (with trailing
// whitespace stripped). Capped so a corrupt or huge file can't blow up
// memory on the device.
func readSpoiler(path string) ([]string, error) {
	const maxBytes = 4 << 20 // 4 MB
	data, err := os.ReadFile(path)
	if err != nil {
		return nil, err
	}
	if len(data) > maxBytes {
		data = data[:maxBytes]
	}
	raw := strings.Split(string(data), "\n")
	lines := make([]string, len(raw))
	for i, ln := range raw {
		lines[i] = strings.TrimRight(ln, "\r ")
	}
	return lines, nil
}

func main() {
	var (
		hostTest bool
		baseROM  string
		outDir   string
	)
	flag.BoolVar(&hostTest, "host-test", false, "run one randomize with defaults and exit (no framebuffer/input)")
	flag.StringVar(&baseROM, "base-rom", "", "override base ROM path (also taken from settings.json)")
	flag.StringVar(&outDir, "out", "", "override output directory")
	flag.Parse()

	cfgPath := configPath()
	cfg, err := LoadConfig(cfgPath)
	if err != nil {
		fmt.Fprintf(os.Stderr, "config: %v (continuing with defaults)\n", err)
		cfg = defaultConfig()
	}
	if baseROM != "" {
		cfg.BaseROM = baseROM
	}
	if outDir != "" {
		cfg.OutputDir = outDir
	}

	if hostTest {
		runHostTest(cfg)
		return
	}

	runInteractive(cfg, cfgPath)
}

func runHostTest(cfg Config) {
	opts := cfg.LastOptions
	opts.BaseROMPath = cfg.BaseROM
	opts.OutputDir = cfg.OutputDir
	// Default to /tmp if the user didn't set anything.
	if opts.OutputDir == "" || opts.OutputDir == defaultOutputDir {
		_ = os.MkdirAll("/tmp/alttpr-brick-test", 0o755)
		opts.OutputDir = "/tmp/alttpr-brick-test"
	}
	res, err := job.Run(opts)
	if err != nil {
		fmt.Fprintf(os.Stderr, "host-test: %v\n", err)
		os.Exit(1)
	}
	fmt.Printf("Spoiler: %s\nROM: %s\n", res.SpoilerPath, res.ROMPath)
}

func runInteractive(cfg Config, cfgPath string) {
	dev, err := fb.Open()
	if err != nil {
		fmt.Fprintf(os.Stderr, "framebuffer: %v\n", err)
		os.Exit(1)
	}
	defer dev.Close()

	in, err := input.Open()
	if err != nil {
		fmt.Fprintf(os.Stderr, "input: %v\n", err)
		os.Exit(1)
	}
	defer in.Close()

	r := &fbRenderer{dev: dev}
	model := ui.New(cfg.LastOptions)

	// Result channel from the background job.
	type jobResult struct {
		res job.Result
		err error
	}
	results := make(chan jobResult, 1)
	var jobRunning bool

	// First paint.
	model.Render(r)
	dev.Present()

	// Render at most ~30 fps, only when something changed.
	tick := time.NewTicker(33 * time.Millisecond)
	defer tick.Stop()

	for {
		dirty := false
		select {
		case ev := <-in.Events:
			action := model.Step(ev.Button)
			dirty = true
			switch action {
			case ui.ActionQuit:
				return
			case ui.ActionGenerate:
				if !jobRunning {
					jobRunning = true
					opts := model.Options
					opts.BaseROMPath = cfg.BaseROM
					opts.OutputDir = cfg.OutputDir
					// Always emit a spoiler in interactive mode.
					opts.WantSpoiler = true
					go func() {
						// Make sure the output dir exists; SD card might be a
						// fresh card.
						_ = os.MkdirAll(opts.OutputDir, 0o755)
						res, err := job.Run(opts)
						results <- jobResult{res, err}
					}()
				}
			case ui.ActionLoadSpoilerList:
				model.SetSpoilerList(listSpoilers(cfg.OutputDir))
			case ui.ActionLoadSpoiler:
				name := model.SelectedSpoiler()
				if name != "" {
					lines, err := readSpoiler(filepath.Join(cfg.OutputDir, name))
					if err != nil {
						model.SetSpoilerContent(name, []string{"Failed to read:", err.Error()})
					} else {
						model.SetSpoilerContent(name, lines)
					}
				}
			case ui.ActionLoadReveal:
				name := model.SelectedSpoiler()
				if name != "" {
					entries, err := parseRevealEntries(filepath.Join(cfg.OutputDir, name))
					if err != nil {
						model.SetSpoilerContent(name, []string{"Failed to parse:", err.Error()})
					} else {
						model.SetRevealEntries(name, entries)
					}
				}
			}
		case res := <-results:
			jobRunning = false
			if res.err != nil {
				model.SetResult(false, "Failed", []string{
					res.err.Error(),
					"Check base_rom in:",
					cfgPath,
				})
			} else {
				lines := []string{}
				if res.res.ROMPath != "" {
					lines = append(lines, "ROM: "+filepath.Base(res.res.ROMPath))
				}
				if res.res.SpoilerPath != "" {
					lines = append(lines, "Spoiler: "+filepath.Base(res.res.SpoilerPath))
				}
				lines = append(lines, "Dropped in: "+cfg.OutputDir)
				model.SetResult(true, "Done", lines)
				// Persist the choices the user just made.
				cfg.LastOptions = model.Options
				_ = SaveConfig(cfgPath, cfg)
			}
			dirty = true
		case <-tick.C:
			// Keep spinner-screens animated even when idle.
			if model.Screen == ui.ScreenGenerating {
				dirty = true
			}
		}
		if dirty {
			model.Render(r)
			dev.Present()
		}
	}
}

// fbRenderer bridges ui.Renderer to the fb package.
type fbRenderer struct{ dev *fb.FB }

func (r *fbRenderer) Bounds() (int, int) { return r.dev.Bounds() }
func (r *fbRenderer) Clear(c ui.Color)   { r.dev.Clear(fb.Color{R: c.R, G: c.G, B: c.B}) }
func (r *fbRenderer) FillRect(x, y, w, h int, c ui.Color) {
	r.dev.FillRect(x, y, w, h, fb.Color{R: c.R, G: c.G, B: c.B})
}
func (r *fbRenderer) DrawText(x, y int, s string, c ui.Color) {
	r.dev.DrawText(x, y, s, fb.Color{R: c.R, G: c.G, B: c.B})
}
