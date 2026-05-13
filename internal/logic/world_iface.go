// Package logic holds interfaces that decouple the item/region/location
// layers from the (much larger) World type, so each can be ported and tested
// without first porting the whole World.
package logic

// World is the minimal contract the item/region/location layers need from a
// world. The full World type (forthcoming) will implement it.
type World interface {
	ID() int
	// ConfigString returns a config value as a string, with the given default.
	// Mirrors PHP World::config($key, $default) for string-valued keys.
	ConfigString(key, def string) string
	// ConfigInt for int-valued keys (e.g. rom.BottleFill.Magic = 0x80).
	ConfigInt(key string, def int) int
	// ConfigBool for boolean keys (e.g. rom.rupeeBow, rom.CatchableFairies).
	ConfigBool(key string, def bool) bool
	// IsInverted reports whether the world is the Inverted variant.
	// PHP code uses `instanceof World\Inverted` for the same check.
	IsInverted() bool
}
