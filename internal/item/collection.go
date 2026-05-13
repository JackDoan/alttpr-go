package item

import (
	"maps"
	"strings"

	"github.com/JackDoan/alttpr-go/internal/logic"
)

// Collection mirrors app/Support/ItemCollection.php.
// PHP stores items by their per-world full name (e.g. "Bow:0") with a
// separate counts map. We keep that exact shape so PHP semantics (filter,
// merge, diff, has-with-threshold) port directly.
type Collection struct {
	// items holds one representative per unique full name, in first-seen order.
	items []*Item
	// counts maps full name -> multiplicity.
	counts map[string]int
	// checksForWorld is the implicit world ID appended to bare names in Has().
	checksForWorld int
}

// NewCollection builds a Collection seeded with `items`. Items with the same
// full name are deduplicated in the order slice but their counts accumulate.
func NewCollection(items ...*Item) *Collection {
	c := &Collection{counts: map[string]int{}}
	for _, it := range items {
		c.Add(it)
	}
	return c
}

// SetChecksForWorld mirrors PHP setChecksForWorld(): the world ID that
// Has(name) implicitly appends.
func (c *Collection) SetChecksForWorld(worldID int) { c.checksForWorld = worldID }

// Add inserts an item. The first occurrence pins the slice order; later
// occurrences only bump the count. Mirrors PHP addItem.
func (c *Collection) Add(it *Item) *Collection {
	name := it.FullName()
	if _, ok := c.counts[name]; !ok {
		c.items = append(c.items, it)
		c.counts[name] = 0
	}
	c.counts[name]++
	return c
}

// Remove decrements the count for `name`, removing the entry entirely once it
// hits zero. `name` is the full name (with ":<worldID>"). Mirrors removeItem.
func (c *Collection) Remove(name string) *Collection {
	if _, ok := c.counts[name]; !ok {
		return c
	}
	c.counts[name]--
	if c.counts[name] == 0 {
		delete(c.counts, name)
		for i, it := range c.items {
			if it.FullName() == name {
				c.items = append(c.items[:i], c.items[i+1:]...)
				break
			}
		}
	}
	return c
}

// Count returns the total multiplicity across all entries.
func (c *Collection) Count() int {
	n := 0
	for _, v := range c.counts {
		n += v
	}
	return n
}

// CountByFullName returns the count for a specific full name.
func (c *Collection) CountByFullName(fullName string) int { return c.counts[fullName] }

// Values returns the (deduplicated, first-seen-order) slice of items.
func (c *Collection) Values() []*Item { return c.items }

// Has reports whether the collection holds at least `atLeast` copies of the
// item with raw name `name` (in the current checksForWorld). Mirrors PHP has.
func (c *Collection) Has(name string, atLeast int) bool {
	if atLeast == 0 {
		return true
	}
	key := name + ":" + itoa(c.checksForWorld)

	// PHP's ShopKey hack: if any ShopKey is in the pool, any "Key*" satisfies.
	if c.counts["ShopKey:"+itoa(c.checksForWorld)] > 0 && strings.HasPrefix(name, "Key") {
		return true
	}

	return c.counts[key] >= atLeast
}

// Has1 is shorthand for Has(name, 1).
func (c *Collection) Has1(name string) bool { return c.Has(name, 1) }

// Filter returns a new Collection containing only items where keep returns true.
func (c *Collection) Filter(keep func(*Item) bool) *Collection {
	out := &Collection{counts: map[string]int{}, checksForWorld: c.checksForWorld}
	for _, it := range c.items {
		if !keep(it) {
			continue
		}
		out.items = append(out.items, it)
		out.counts[it.FullName()] = c.counts[it.FullName()]
	}
	return out
}

// Each invokes fn once per (count) occurrence of each item.
func (c *Collection) Each(fn func(*Item)) {
	for _, it := range c.items {
		n := c.counts[it.FullName()]
		for range n {
			fn(it)
		}
	}
}

// Copy returns a shallow clone (items pointers shared).
func (c *Collection) Copy() *Collection {
	out := &Collection{
		items:          append([]*Item(nil), c.items...),
		counts:         make(map[string]int, len(c.counts)),
		checksForWorld: c.checksForWorld,
	}
	maps.Copy(out.counts, c.counts)
	return out
}

// Merge returns a fresh Collection with both this and `other`'s items.
func (c *Collection) Merge(other *Collection) *Collection {
	merged := c.Copy()
	other.Each(func(it *Item) { merged.Add(it) })
	return merged
}

// Diff returns the items present here but not in `other`, using the same
// per-name count-difference semantics as PHP ItemCollection::diff.
func (c *Collection) Diff(other *Collection) *Collection {
	out := c.Copy()
	for name, oc := range other.counts {
		if cc, ok := out.counts[name]; ok {
			if oc < cc {
				out.counts[name] = cc - oc
			} else {
				delete(out.counts, name)
				for i, it := range out.items {
					if it.FullName() == name {
						out.items = append(out.items[:i], out.items[i+1:]...)
						break
					}
				}
			}
		}
	}
	return out
}

// ManyKeys bumps every Key* count to 10 (used by the test harness; mirrors PHP).
func (c *Collection) ManyKeys() *Collection {
	for k := range c.counts {
		if strings.HasPrefix(k, "Key") {
			c.counts[k] = 10
		}
	}
	return c
}

// ---- Domain helpers (port of ItemCollection's logic methods) ----

func (c *Collection) HeartCount(initial float64) float64 {
	out := initial
	for _, it := range c.items {
		if !it.IsType(TypeUpgradeHealth) {
			continue
		}
		n := c.counts[it.FullName()]
		for range n {
			if it.Name == "PieceOfHeart" {
				out += 0.25
			} else {
				out += 1
			}
		}
	}
	return out
}

func (c *Collection) HasHealth(minimum float64) bool {
	sum := 0.0
	for _, it := range c.items {
		if !it.IsType(TypeUpgradeHealth) {
			continue
		}
		power := it.Power
		if it.Type == TypeAlias && it.Target != nil {
			power = it.Target.Power
		}
		sum += power * float64(c.counts[it.FullName()])
	}
	return sum >= minimum
}

func (c *Collection) CanLiftRocks() bool {
	return c.Has1("PowerGlove") || c.Has1("ProgressiveGlove") || c.Has1("TitansMitt")
}

func (c *Collection) CanLiftDarkRocks() bool {
	return c.Has1("TitansMitt") || c.Has("ProgressiveGlove", 2)
}

func (c *Collection) CanLightTorches() bool { return c.Has1("FireRod") || c.Has1("Lamp") }

func (c *Collection) CanBlockLasers() bool {
	return c.Has1("MirrorShield") || c.Has("ProgressiveShield", 3)
}

func (c *Collection) CanBombThings() bool { return true }

func (c *Collection) CanSpinSpeed() bool {
	return c.Has1("PegasusBoots") && (c.HasSword(1) || c.Has1("Hookshot"))
}

func (c *Collection) HasSword(minLevel int) bool {
	switch minLevel {
	case 4:
		return c.Has("ProgressiveSword", 4) ||
			(c.Has1("UncleSword") && c.Has("ProgressiveSword", 3)) ||
			c.Has1("L4Sword")
	case 3:
		return c.Has("ProgressiveSword", 3) ||
			(c.Has1("UncleSword") && c.Has("ProgressiveSword", 2)) ||
			c.Has1("L3Sword") || c.Has1("L4Sword")
	case 2:
		return c.Has("ProgressiveSword", 2) ||
			(c.Has1("UncleSword") && c.Has1("ProgressiveSword")) ||
			c.Has1("L2Sword") || c.Has1("MasterSword") ||
			c.Has1("L3Sword") || c.Has1("L4Sword")
	default:
		return c.Has1("ProgressiveSword") || c.Has1("UncleSword") ||
			c.Has1("L1Sword") || c.Has1("L1SwordAndShield") ||
			c.Has1("L2Sword") || c.Has1("MasterSword") ||
			c.Has1("L3Sword") || c.Has1("L4Sword")
	}
}

func (c *Collection) HasArmor(minLevel int) bool {
	switch minLevel {
	case 2:
		return c.Has("ProgressiveArmor", 2) || c.Has1("RedMail")
	default:
		return c.Has1("ProgressiveArmor") || c.Has1("BlueMail") || c.Has1("RedMail")
	}
}

// HasABottle returns true if any bottle variant (named) is present.
func (c *Collection) HasABottle() bool {
	return c.Has1("BottleWithBee") || c.Has1("BottleWithFairy") ||
		c.Has1("BottleWithRedPotion") || c.Has1("BottleWithGreenPotion") ||
		c.Has1("BottleWithBluePotion") || c.Has1("Bottle") ||
		c.Has1("BottleWithGoldBee")
}

// BottleCount counts all bottle-type items.
func (c *Collection) BottleCount() int {
	n := 0
	for _, it := range c.items {
		if it.IsType(TypeBottle) {
			n += c.counts[it.FullName()]
		}
	}
	return n
}

func (c *Collection) HasBottle(atLeast int) bool { return c.BottleCount() >= atLeast }

func (c *Collection) GlitchedLinkInDarkWorld() bool { return c.Has1("MoonPearl") || c.HasABottle() }

func (c *Collection) CanGetGoodBee() bool {
	return c.Has1("BugCatchingNet") && c.HasABottle() &&
		(c.Has1("PegasusBoots") || (c.HasSword(1) && c.Has1("Quake")))
}

// CanAcquireFairy mirrors PHP — returns false only if the world's
// rom.CatchableFairies is explicitly disabled.
func (c *Collection) CanAcquireFairy(w logic.World) bool {
	if w == nil {
		return true
	}
	return w.ConfigBool("rom.CatchableFairies", true)
}

func (c *Collection) CanBunnyRevive(w logic.World) bool {
	base := c.HasABottle() && c.Has1("BugCatchingNet")
	if !base {
		return false
	}
	if w != nil {
		return c.CanAcquireFairy(w)
	}
	return true
}

// CanFly mirrors PHP canFly — depends on inverted/normal world.
func (c *Collection) CanFly(w logic.World) bool {
	if c.Has1("OcarinaActive") {
		return true
	}
	if !c.Has1("OcarinaInactive") {
		return false
	}
	return c.canActivateOcarina(w)
}

func (c *Collection) canActivateOcarina(w logic.World) bool {
	if w != nil && w.IsInverted() {
		return c.Has1("MoonPearl") &&
			(c.Has1("DefeatAgahnim") ||
				((c.Has1("Hammer") && c.CanLiftRocks()) || c.CanLiftDarkRocks()))
	}
	return true
}

// CanShootArrows mirrors PHP canShootArrows.
func (c *Collection) CanShootArrows(w logic.World, minLevel int) bool {
	rupeeBow := w != nil && w.ConfigBool("rom.rupeeBow", false)
	switch minLevel {
	case 2:
		return c.Has1("BowAndSilverArrows") ||
			(c.Has("ProgressiveBow", 2) && (!rupeeBow || c.Has1("ShopArrow"))) ||
			(c.Has1("SilverArrowUpgrade") &&
				(c.Has1("Bow") || c.Has1("BowAndArrows") || c.Has1("ProgressiveBow")))
	default:
		return ((c.Has1("Bow") || c.Has1("ProgressiveBow")) &&
			(!rupeeBow || c.Has1("ShopArrow") || c.Has1("SilverArrowUpgrade"))) ||
			c.Has1("BowAndArrows") || c.Has1("BowAndSilverArrows")
	}
}

// CanMeltThings mirrors PHP canMeltThings.
func (c *Collection) CanMeltThings(w logic.World) bool {
	if c.Has1("FireRod") {
		return true
	}
	if !c.Has1("Bombos") {
		return false
	}
	swordless := w != nil && w.ConfigString("mode.weapons", "") == "swordless"
	return swordless || c.HasSword(1)
}

// CanExtendMagic mirrors PHP canExtendMagic.
func (c *Collection) CanExtendMagic(w logic.World, bars float64) bool {
	magicMod := 1.0
	if w != nil {
		fill := w.ConfigInt("rom.BottleFill.Magic", 0x80)
		magicMod = float64(fill) / float64(0x80)
		if magicMod > 1 {
			magicMod = 1
		}
	}
	base := 1
	if c.Has1("QuarterMagic") {
		base = 4
	} else if c.Has1("HalfMagic") {
		base = 2
	}
	bottle := float64(base) * float64(c.BottleCount()) * magicMod
	return float64(base)+bottle >= bars
}

// CanKillEscapeThings mirrors PHP canKillEscapeThings.
func (c *Collection) CanKillEscapeThings(w logic.World) bool {
	defaultHealth := w == nil || w.ConfigString("enemizer.enemyHealth", "default") == "default"
	if c.Has1("UncleSword") || c.Has1("CaneOfSomaria") {
		return true
	}
	if defaultHealth && (c.Has1("TenBombs") || c.Has1("CaneOfByrna")) {
		return true
	}
	if c.CanShootArrows(w, 1) {
		return true
	}
	if c.Has1("Hammer") || c.Has1("FireRod") {
		return true
	}
	if w != nil && w.ConfigBool("ignoreCanKillEscapeThings", false) {
		return true
	}
	return false
}

// CanKillMostThings mirrors PHP canKillMostThings.
func (c *Collection) CanKillMostThings(w logic.World, enemies int) bool {
	if c.HasSword(1) || c.Has1("CaneOfSomaria") {
		return true
	}
	defaultHealth := w == nil || w.ConfigString("enemizer.enemyHealth", "default") == "default"
	if defaultHealth {
		if c.CanBombThings() && enemies < 6 {
			return true
		}
		if c.Has1("CaneOfByrna") && (enemies < 6 || c.CanExtendMagic(w, 2.0)) {
			return true
		}
	}
	if c.CanShootArrows(w, 1) {
		return true
	}
	return c.Has1("Hammer") || c.Has1("FireRod")
}

// itoa avoids strconv for the hot path.
func itoa(n int) string {
	if n == 0 {
		return "0"
	}
	neg := n < 0
	if neg {
		n = -n
	}
	var buf [20]byte
	i := len(buf)
	for n > 0 {
		i--
		buf[i] = byte('0' + n%10)
		n /= 10
	}
	if neg {
		i--
		buf[i] = '-'
	}
	return string(buf[i:])
}
