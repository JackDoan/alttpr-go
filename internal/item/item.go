// Package item is the Go port of app/Item.php and its subclasses.
//
// PHP uses class inheritance for item kinds (Item\Sword, Item\Bottle, etc.).
// Most subclasses are empty marker classes used only for `instanceof` checks.
// We collapse the hierarchy to a single Item struct with a Type field;
// the special-case Health upgrade carries an extra Power float.
package item

import "fmt"

// Type discriminates item subclasses (PHP Item\Sword, Item\Bottle, ...).
type Type int

const (
	TypeGeneric Type = iota
	TypeSword
	TypeShield
	TypeBow
	TypeBottle
	TypeBottleContents
	TypeMedallion
	TypePendant
	TypeCrystal
	TypeArmor
	TypeKey
	TypeBigKey
	TypeMap
	TypeCompass
	TypeArrow
	TypeUpgradeArrow
	TypeUpgradeBomb
	TypeUpgradeHealth
	TypeUpgradeMagic
	TypeProgrammable
	TypeEvent
	TypeAlias // ItemAlias
)

// Item is one collectable.
type Item struct {
	Name      string
	NiceName  string
	Bytes     []int    // may contain -1 for PHP `null` slots; -1 sentinel preserved for transcription fidelity
	NamedBytes map[string]int // medallion-specific named keys t0/t1/t2/m0/m1/m2
	WorldID   int
	Type      Type

	// Type-specific fields:
	Power  float64 // TypeUpgradeHealth: hearts contributed (.25 for PieceOfHeart, 1 for HeartContainer, etc.)

	// Alias-only fields (TypeAlias):
	TargetName string
	Target     *Item
}

// NewItem constructs a generic Item.
func NewItem(name string, bytes []int, worldID int, t Type) *Item {
	return &Item{
		Name:     name,
		NiceName: "item." + name,
		Bytes:    bytes,
		WorldID:  worldID,
		Type:     t,
	}
}

// FullName mirrors PHP getName() which appends ":<worldID>".
func (i *Item) FullName() string {
	return fmt.Sprintf("%s:%d", i.Name, i.WorldID)
}

// RawName mirrors PHP getRawName().
func (i *Item) RawName() string {
	return i.Name
}

// GetBytes returns the bytes to write. For aliases this proxies to the target.
func (i *Item) GetBytes() []int {
	if i.Type == TypeAlias && i.Target != nil {
		return i.Target.GetBytes()
	}
	return i.Bytes
}

// GetNamedBytes returns medallion-style named bytes; nil for non-medallions.
func (i *Item) GetNamedBytes() map[string]int {
	if i.Type == TypeAlias && i.Target != nil {
		return i.Target.NamedBytes
	}
	return i.NamedBytes
}

// IsType reports whether this item is an instance of t. Aliases match their
// target's type as well as TypeAlias (matching PHP `instanceof` semantics
// where ItemAlias extends Item).
func (i *Item) IsType(t Type) bool {
	if i.Type == t {
		return true
	}
	if i.Type == TypeAlias && i.Target != nil && i.Target.Type == t {
		return true
	}
	return false
}

// String mirrors PHP __toString(): name plus a stable encoding of bytes.
func (i *Item) String() string {
	return fmt.Sprintf("%s%v", i.Name, i.GetBytes())
}
