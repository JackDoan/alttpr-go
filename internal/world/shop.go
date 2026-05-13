package world

import "github.com/JackDoan/alttpr-go/internal/item"

// ShopKind distinguishes regular shops from take-any. Mirrors PHP class
// inheritance (Shop, Shop\TakeAny, Shop\UpgradeShop).
type ShopKind int

const (
	ShopRegular ShopKind = iota
	ShopTakeAny
	ShopUpgrade
)

// ShopInventory is one slot: an item, a price, and optional max + replacement.
// Mirrors PHP Shop::$inventory entries.
type ShopInventory struct {
	Item            *item.Item
	Price           int
	Max             int
	Replacement     *item.Item
	ReplacePrice    int
}

// Shop mirrors PHP `ALttP\Shop`. PHP constructor signature:
//
//	Shop(name, config, shopkeeper, room_id, door_id, region, writes=[])
//
// `Config` is a packed byte (`td----qq`): take-any flag + door-check + slot count.
type Shop struct {
	Name       string
	Kind       ShopKind
	Config     int
	Shopkeeper int
	RoomID     int
	DoorID     int
	Active     bool
	Inventory  []ShopInventory
	Region     *Region

	// Requirement, if set, gates whether the shop is accessible.
	Requirement RequirementFunc

	// Writes are extra (address → bytes) writes applied when this shop
	// activates (used by Take-any to hijack entrance tables).
	Writes map[int][]int
}

// AddInventory mirrors PHP Shop::addInventory.
func (s *Shop) AddInventory(slot int, it *item.Item, price int) *Shop {
	for len(s.Inventory) <= slot {
		s.Inventory = append(s.Inventory, ShopInventory{})
	}
	s.Inventory[slot] = ShopInventory{Item: it, Price: price}
	return s
}

// ClearInventory empties Inventory.
func (s *Shop) ClearInventory() *Shop { s.Inventory = nil; return s }

// firstByte returns the first byte of an item's bytes, or 0xFF.
func firstByte(it *item.Item) int {
	if it == nil {
		return 0xFF
	}
	b := it.GetBytes()
	if len(b) == 0 {
		return 0xFF
	}
	return b[0] & 0xFF
}

// Bytes returns the 7-byte shop entry for the 0x184800 table:
// `room_id(LE16), door_id, 0x00, (config & 0xFC) + count(inv), shopkeeper, sram_offset`.
// Mirrors PHP Shop::getBytes.
func (s *Shop) Bytes(sramOffset int) []byte {
	out := []byte{
		byte(s.RoomID & 0xFF), byte((s.RoomID >> 8) & 0xFF),
		byte(s.DoorID & 0xFF),
		0x00,
		byte((s.Config & 0xFC) + len(s.Inventory)),
		byte(s.Shopkeeper & 0xFF),
		byte(sramOffset & 0xFF),
	}
	return out
}

// ShopCollection is an ordered map of Shops by name.
type ShopCollection struct {
	order []*Shop
	byKey map[string]*Shop
}

func NewShopCollection(shops ...*Shop) *ShopCollection {
	c := &ShopCollection{byKey: map[string]*Shop{}}
	for _, s := range shops {
		c.Add(s)
	}
	return c
}

func (c *ShopCollection) Add(s *Shop) *ShopCollection {
	if _, ok := c.byKey[s.Name]; ok {
		return c
	}
	c.byKey[s.Name] = s
	c.order = append(c.order, s)
	return c
}

func (c *ShopCollection) Get(name string) *Shop { return c.byKey[name] }
func (c *ShopCollection) All() []*Shop          { return c.order }
func (c *ShopCollection) Count() int            { return len(c.order) }
