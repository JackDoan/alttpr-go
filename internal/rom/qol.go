package rom

import "fmt"

// SetHeartBeepSpeed mirrors app/Rom.php:177-200.
func (r *ROM) SetHeartBeepSpeed(setting string) error {
	var b byte
	switch setting {
	case "off":
		b = 0x00
	case "half":
		b = 0x40
	case "quarter":
		b = 0x80
	case "double":
		b = 0x10
	case "normal", "":
		b = 0x20
	default:
		b = 0x20
	}
	return r.Write(0x180033, []byte{b}, true)
}

// SetHeartColors mirrors app/Rom.php:637-656.
func (r *ROM) SetHeartColors(color string) error {
	var b byte
	switch color {
	case "blue":
		b = 0x01
	case "green":
		b = 0x02
	case "yellow":
		b = 0x03
	case "red", "":
		b = 0x00
	default:
		b = 0x00
	}
	return r.Write(0x187020, []byte{b}, true)
}

// SetMenuSpeed mirrors app/Rom.php:721-746.
func (r *ROM) SetMenuSpeed(menuSpeed string) error {
	var speed byte
	fast := false
	switch menuSpeed {
	case "instant":
		speed = 0xE8
		fast = true
	case "fast":
		speed = 0x10
	case "slow":
		speed = 0x04
	case "normal", "":
		speed = 0x08
	default:
		speed = 0x08
	}
	if err := r.Write(0x180048, []byte{speed}, true); err != nil {
		return err
	}
	pickA := byte(0x11)
	pickB := byte(0x12)
	if fast {
		pickA = 0x20
		pickB = 0x20
	}
	if err := r.Write(0x6DD9A, []byte{pickA}, true); err != nil {
		return err
	}
	if err := r.Write(0x6DF2A, []byte{pickB}, true); err != nil {
		return err
	}
	return r.Write(0x6E0E9, []byte{pickB}, true)
}

// SetQuickSwap mirrors app/Rom.php:755-760.
func (r *ROM) SetQuickSwap(enable bool) error {
	b := byte(0x00)
	if enable {
		b = 0x01
	}
	return r.Write(0x18004B, []byte{b}, true)
}

// MuteMusic mirrors app/Rom.php:2375-2380.
func (r *ROM) MuteMusic(enable bool) error {
	b := byte(0x00)
	if enable {
		b = 0x01
	}
	return r.Write(0x18021A, []byte{b}, true)
}

// ParseBool accepts "true"/"false" (case-insensitive) and the literal Go-true values.
// Mirrors the PHP randomizer's `strtolower($v) === 'true'` style for the `--quickswap` flag.
func ParseBool(s string) (bool, error) {
	switch s {
	case "true", "TRUE", "True", "1":
		return true, nil
	case "false", "FALSE", "False", "0", "":
		return false, nil
	}
	return false, fmt.Errorf("invalid bool %q", s)
}
