package world

import "github.com/JackDoan/alttpr-go/internal/rom"

// WriteShops emits the active-shop bytes table at 0x184800 and the
// per-inventory-item table at 0x184900. Mirrors PHP Rom::setupCustomShops.
func (w *World) WriteShops(r *rom.ROM) error {
	active := []*Shop{}
	for _, s := range w.shops.All() {
		if s.Active {
			active = append(active, s)
		}
	}

	shopData := []byte{}
	itemsData := []byte{}
	shopID := byte(0x00)
	sramOffset := 0

	for i, s := range active {
		id := shopID
		if i == len(active)-1 {
			id = 0xFF // sentinel
		}

		// Apply this shop's writeExtraData (entrance-table hijacks for take-anys).
		if err := s.writeExtra(r); err != nil {
			return err
		}

		shopData = append(shopData, id)
		shopData = append(shopData, s.Bytes(sramOffset)...)

		if s.Kind == ShopTakeAny {
			sramOffset++
		} else {
			sramOffset += len(s.Inventory)
		}
		if sramOffset > 36 {
			return errOverflow("shop SRAM indexing")
		}

		for _, inv := range s.Inventory {
			itemsData = append(itemsData,
				id,
				byte(firstByte(inv.Item)),
				byte(inv.Price&0xFF), byte((inv.Price>>8)&0xFF),
				byte(inv.Max),
				byte(firstByte(inv.Replacement)),
				byte(inv.ReplacePrice&0xFF), byte((inv.ReplacePrice>>8)&0xFF),
			)
		}
		shopID++
	}

	if err := r.Write(0x184800, shopData, true); err != nil {
		return err
	}
	itemsData = append(itemsData, 0xFF, 0xFF, 0xFF, 0xFF, 0xFF, 0xFF, 0xFF, 0xFF)
	return r.Write(0x184900, itemsData, true)
}

// writeExtra applies the Writes map (entrance-table hijacks).
func (s *Shop) writeExtra(r *rom.ROM) error {
	for addr, bytes := range s.Writes {
		buf := make([]byte, len(bytes))
		for i, v := range bytes {
			buf[i] = byte(v)
		}
		if err := r.Write(addr, buf, true); err != nil {
			return err
		}
	}
	return nil
}

// errOverflow is a small helper for human-readable overflow errors.
type errOverflowKind string

func (e errOverflowKind) Error() string { return "overflow: " + string(e) }
func errOverflow(s string) error        { return errOverflowKind(s) }
