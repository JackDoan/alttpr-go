package world

import "github.com/JackDoan/alttpr-go/internal/item"

// Spoiler is the JSON-serializable record of a randomized world.
// Mirrors PHP World::getSpoiler output shape.
type Spoiler struct {
	Regions     map[string]map[string]string `json:"regions"`
	Shops       []SpoilerShop                `json:"shops,omitempty"`
	Bosses      map[string]string            `json:"Bosses"`
	Equipped    map[string]string            `json:"Equipped,omitempty"`
	Meta        SpoilerMeta                  `json:"meta"`
	Playthrough *Playthrough                 `json:"playthrough,omitempty"`
}

// SpoilerShop describes one active shop in the spoiler.
type SpoilerShop struct {
	Location string         `json:"location"`
	Type     string         `json:"type"`
	Items    map[int]string `json:"items,omitempty"`
}

// SpoilerMeta carries metadata fields. Mirrors PHP $spoiler['meta'].
type SpoilerMeta struct {
	ItemPlacement     string `json:"item_placement"`
	ItemPool          string `json:"item_pool"`
	ItemFunctionality string `json:"item_functionality"`
	DungeonItems      string `json:"dungeon_items"`
	Logic             string `json:"logic"`
	Accessibility     string `json:"accessibility"`
	Goal              string `json:"goal"`
	Build             string `json:"build"`
	Mode              string `json:"mode"`
	Weapons           string `json:"weapons"`
	WorldID           int    `json:"world_id"`
	CrystalsGanon     int    `json:"crystals_ganon"`
	CrystalsTower     int    `json:"crystals_tower"`
	Tournament        bool   `json:"tournament"`
	Hints             string `json:"hints"`
}

// GetSpoiler builds the spoiler for this world.
// Mirrors a subset of PHP World::getSpoiler — enough to drive the CLI
// `--spoiler` flag. Excludes hints + playthrough (those are PHP Services).
func (w *World) GetSpoiler() Spoiler {
	s := Spoiler{
		Regions: map[string]map[string]string{},
		Bosses:  map[string]string{},
	}

	for _, r := range w.regions {
		entries := map[string]string{}
		for _, l := range r.Locations.All() {
			if l.Kind == KindPrizeEvent || l.Kind == KindTrade {
				continue
			}
			if l.HasItem() {
				it := l.Item()
				name := it.Name
				if w.ConfigBool("rom.genericKeys", false) && it.IsType(item.TypeKey) {
					name = "Key"
				} else if it.Type == item.TypeAlias && it.Target != nil {
					name = it.Target.Name
				}
				entries[l.Name] = name
			} else {
				entries[l.Name] = "Nothing"
			}
		}
		s.Regions[r.Name] = entries
	}

	for _, sh := range w.shops.All() {
		if !sh.Active {
			continue
		}
		sd := SpoilerShop{Location: sh.Name, Type: "Shop"}
		if sh.Kind == ShopTakeAny {
			sd.Type = "Take Any"
		}
		sd.Items = map[int]string{}
		for i, inv := range sh.Inventory {
			if inv.Item != nil {
				sd.Items[i] = inv.Item.Name
			}
		}
		s.Shops = append(s.Shops, sd)
	}

	bossName := func(regionName, level string) string {
		r := w.regions[regionName]
		if r == nil {
			return ""
		}
		b := r.BossAt(level)
		if b == nil {
			return ""
		}
		return b.Name
	}
	s.Bosses["Eastern Palace"] = bossName("Eastern Palace", "")
	s.Bosses["Desert Palace"] = bossName("Desert Palace", "")
	s.Bosses["Tower Of Hera"] = bossName("Tower of Hera", "")
	s.Bosses["Hyrule Castle"] = "Agahnim"
	s.Bosses["Palace Of Darkness"] = bossName("Palace of Darkness", "")
	s.Bosses["Swamp Palace"] = bossName("Swamp Palace", "")
	s.Bosses["Skull Woods"] = bossName("Skull Woods", "")
	s.Bosses["Thieves Town"] = bossName("Thieves Town", "")
	s.Bosses["Ice Palace"] = bossName("Ice Palace", "")
	s.Bosses["Misery Mire"] = bossName("Misery Mire", "")
	s.Bosses["Turtle Rock"] = bossName("Turtle Rock", "")
	s.Bosses["Ganons Tower Basement"] = bossName("Ganons Tower", "bottom")
	s.Bosses["Ganons Tower Middle"] = bossName("Ganons Tower", "middle")
	s.Bosses["Ganons Tower Top"] = bossName("Ganons Tower", "top")
	s.Bosses["Ganons Tower"] = "Agahnim 2"
	s.Bosses["Ganon"] = "Ganon"

	// Pre-collected ("Equipped") slot listing.
	if w.preCollected.Count() > 0 {
		s.Equipped = map[string]string{}
		i := 0
		for _, it := range w.preCollected.Values() {
			if it.IsType(item.TypeUpgradeArrow) || it.IsType(item.TypeUpgradeBomb) || it.IsType(item.TypeEvent) {
				continue
			}
			i++
			name := it.Name
			if it.Type == item.TypeAlias && it.Target != nil {
				name = it.Target.Name
			}
			s.Equipped[itoa2W(i, "Equipment Slot ")] = name
		}
	}

	s.Playthrough = w.GetPlaythrough()

	s.Meta = SpoilerMeta{
		ItemPlacement:     w.ConfigString("itemPlacement", ""),
		ItemPool:          w.ConfigString("item.pool", ""),
		ItemFunctionality: w.ConfigString("item.functionality", ""),
		DungeonItems:      w.ConfigString("dungeonItems", ""),
		Logic:             w.ConfigString("logic", ""),
		Accessibility:     w.ConfigString("accessibility", ""),
		Goal:              w.ConfigString("goal", ""),
		Build:             "2024-02-18", // matches rom.Build
		Mode:              w.ConfigString("mode.state", ""),
		Weapons:           w.ConfigString("mode.weapons", ""),
		WorldID:           w.id,
		CrystalsGanon:     w.ConfigInt("crystals.ganon", 7),
		CrystalsTower:     w.ConfigInt("crystals.tower", 7),
		Tournament:        w.ConfigBool("tournament", false),
		Hints:             w.ConfigString("spoil.Hints", ""),
	}
	return s
}

// itoa2W stringifies "prefix<n>" without strconv.
func itoa2W(n int, prefix string) string {
	if n == 0 {
		return prefix + "0"
	}
	var buf [20]byte
	i := len(buf)
	for n > 0 {
		i--
		buf[i] = byte('0' + n%10)
		n /= 10
	}
	return prefix + string(buf[i:])
}
