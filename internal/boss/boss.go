// Package boss is the Go port of app/Boss.php — boss objects with
// per-boss "can beat" predicates that depend on the player's items
// and the world's config.
package boss

import (
	"fmt"

	"github.com/JackDoan/alttpr-go/internal/item"
	"github.com/JackDoan/alttpr-go/internal/logic"
)

// LocationView is the slice of LocationCollection used by boss predicates.
// Most predicates ignore it; we keep the parameter to match PHP semantics
// so future location-aware predicates can land cleanly.
type LocationView any

// BeatFunc is the per-boss predicate signature.
type BeatFunc func(loc LocationView, items *item.Collection) bool

// Boss is one boss with its enemizer name and beat predicate.
type Boss struct {
	Name         string
	EnemizerName string
	canBeat      BeatFunc
}

// CanBeat reports whether the player can beat this boss with `items`.
// A nil predicate means "always".
func (b *Boss) CanBeat(items *item.Collection, locs LocationView) bool {
	if b.canBeat == nil {
		return true
	}
	return b.canBeat(locs, items)
}

// Registry caches bosses per world.
type Registry struct {
	byWorld map[int]map[string]*Boss
}

func NewRegistry() *Registry { return &Registry{byWorld: map[int]map[string]*Boss{}} }

// ClearCache mirrors PHP Boss::clearCache().
func (r *Registry) ClearCache() { r.byWorld = map[int]map[string]*Boss{} }

// Get fetches a boss by name within the given world.
func (r *Registry) Get(name string, w logic.World) (*Boss, error) {
	all := r.All(w)
	b, ok := all[name]
	if !ok {
		return nil, fmt.Errorf("unknown boss: %s", name)
	}
	return b, nil
}

// All returns the per-world boss table, populating on first access.
func (r *Registry) All(w logic.World) map[string]*Boss {
	id := w.ID()
	if existing, ok := r.byWorld[id]; ok {
		return existing
	}
	r.byWorld[id] = bossesFor(w)
	return r.byWorld[id]
}

func bossesFor(w logic.World) map[string]*Boss {
	mk := func(name, ename string, fn BeatFunc) *Boss {
		return &Boss{Name: name, EnemizerName: ename, canBeat: fn}
	}

	notBasic := func() bool { return w.ConfigString("itemPlacement", "") != "basic" }
	swordless := func() bool { return w.ConfigString("mode.weapons", "") == "swordless" }

	return map[string]*Boss{
		"Armos Knights": mk("Armos Knights", "Armos", func(_ LocationView, it *item.Collection) bool {
			return it.HasSword(1) || it.Has1("Hammer") || it.CanShootArrows(w, 1) ||
				it.Has1("Boomerang") || it.Has1("RedBoomerang") ||
				(it.CanExtendMagic(w, 4) && (it.Has1("FireRod") || it.Has1("IceRod"))) ||
				(it.CanExtendMagic(w, 2) && (it.Has1("CaneOfByrna") || it.Has1("CaneOfSomaria")))
		}),
		"Lanmolas": mk("Lanmolas", "Lanmola", func(_ LocationView, it *item.Collection) bool {
			return it.HasSword(1) || it.Has1("Hammer") || it.CanShootArrows(w, 1) ||
				it.Has1("FireRod") || it.Has1("IceRod") ||
				it.Has1("CaneOfByrna") || it.Has1("CaneOfSomaria")
		}),
		"Moldorm": mk("Moldorm", "Moldorm", func(_ LocationView, it *item.Collection) bool {
			return it.HasSword(1) || it.Has1("Hammer")
		}),
		"Agahnim": mk("Agahnim", "Agahnim", func(_ LocationView, it *item.Collection) bool {
			return it.HasSword(1) || it.Has1("Hammer") || it.Has1("BugCatchingNet")
		}),
		"Helmasaur King": mk("Helmasaur King", "Helmasaur", func(_ LocationView, it *item.Collection) bool {
			return (it.CanBombThings() || it.Has1("Hammer")) &&
				(it.HasSword(2) || it.CanShootArrows(w, 1) ||
					(notBasic() && it.HasSword(1)))
		}),
		"Arrghus": mk("Arrghus", "Arrghus", func(_ LocationView, it *item.Collection) bool {
			pre := notBasic() || swordless() || it.HasSword(2)
			return pre && it.Has1("Hookshot") &&
				(it.Has1("Hammer") || it.HasSword(1) ||
					((it.CanExtendMagic(w, 2) || it.CanShootArrows(w, 1)) &&
						(it.Has1("FireRod") || it.Has1("IceRod"))))
		}),
		"Mothula": mk("Mothula", "Mothula", func(_ LocationView, it *item.Collection) bool {
			pre := notBasic() || it.HasSword(2) || (it.CanExtendMagic(w, 2) && it.Has1("FireRod"))
			return pre && (it.HasSword(1) || it.Has1("Hammer") ||
				(it.CanExtendMagic(w, 2) && (it.Has1("FireRod") || it.Has1("CaneOfSomaria") || it.Has1("CaneOfByrna"))) ||
				it.CanGetGoodBee())
		}),
		"Blind": mk("Blind", "Blind", func(_ LocationView, it *item.Collection) bool {
			pre := notBasic() || swordless() ||
				(it.HasSword(1) && (it.Has1("Cape") || it.Has1("CaneOfByrna")))
			return pre && (it.HasSword(1) || it.Has1("Hammer") ||
				it.Has1("CaneOfSomaria") || it.Has1("CaneOfByrna"))
		}),
		"Kholdstare": mk("Kholdstare", "Kholdstare", func(_ LocationView, it *item.Collection) bool {
			pre := notBasic() || it.HasSword(2) ||
				(it.CanExtendMagic(w, 3) && it.Has1("FireRod")) ||
				(it.Has1("Bombos") && (swordless() || it.HasSword(1)) && it.CanExtendMagic(w, 2) && it.Has1("FireRod"))
			return pre && it.CanMeltThings(w) && (it.Has1("Hammer") || it.HasSword(1) ||
				(it.CanExtendMagic(w, 3) && it.Has1("FireRod")) ||
				(it.CanExtendMagic(w, 2) && it.Has1("FireRod") && it.Has1("Bombos") && swordless()))
		}),
		"Vitreous": mk("Vitreous", "Vitreous", func(_ LocationView, it *item.Collection) bool {
			pre := notBasic() || it.HasSword(2) || it.CanShootArrows(w, 1)
			return pre && (it.Has1("Hammer") || it.HasSword(1) || it.CanShootArrows(w, 1))
		}),
		"Trinexx": mk("Trinexx", "Trinexx", func(_ LocationView, it *item.Collection) bool {
			pre := notBasic() || swordless() || it.HasSword(3) ||
				(it.CanExtendMagic(w, 2) && it.HasSword(2))
			return it.Has1("FireRod") && it.Has1("IceRod") && pre &&
				(it.HasSword(3) || it.Has1("Hammer") ||
					(it.CanExtendMagic(w, 2) && it.HasSword(2)) ||
					(it.CanExtendMagic(w, 4) && it.HasSword(1)))
		}),
		"Agahnim2": mk("Agahnim2", "Agahnim2", func(_ LocationView, it *item.Collection) bool {
			return it.HasSword(1) || it.Has1("Hammer") || it.Has1("BugCatchingNet")
		}),
	}
}
