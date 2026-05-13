package helpers

// PcToSnes converts a PC ROM address to SNES LoROM. Mirrors
// app/Helpers/number.php:pc_to_snes.
func PcToSnes(address int) int {
	return ((address << 1) & 0x7F0000) | (address & 0x7FFF) | 0x8000
}

// SnesToPc converts a SNES LoROM address to PC. Mirrors snes_to_pc.
func SnesToPc(address int) int {
	return ((address & 0x7F0000) >> 1) | (address & 0x7FFF)
}

// CountSetBits returns the number of 1-bits in value. Mirrors count_set_bits.
func CountSetBits(value int) int {
	n := 0
	for value != 0 {
		n += value & 1
		value >>= 1
	}
	return n
}

// HashArray reproduces the PHP `hash_array` function exactly, returning the
// 5-byte hash representation used by the seed name encoder.
// Mirrors app/Helpers/array.php:hash_array.
func HashArray(id int) [5]int {
	id = (id * 99371) % 33554431
	ret := 0
	for i := range 25 {
		bit := (id >> i) & 1
		shift := ((i%5)+1)*5 - (i / 5)
		ret += bit << shift
	}
	return [5]int{
		(ret >> 20) & 0x1F,
		(ret >> 15) & 0x1F,
		(ret >> 10) & 0x1F,
		(ret >> 5) & 0x1F,
		ret & 0x1F,
	}
}
