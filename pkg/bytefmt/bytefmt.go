package bytefmt

import (
	"fmt"
	"strings"
)

const (
	Byte     = 1.0
	Kilobyte = 1024 * Byte
	Megabyte = 1024 * Kilobyte
	Gigabyte = 1024 * Megabyte
	Terabyte = 1024 * Gigabyte
)

// HumanReadableByte returns a human-readable byte string of the form 10M, 12.5K, and so forth.
//
// The following units are available:
//	T: Terabyte
//	G: Gigabyte
//	M: Megabyte
//	K: Kilobyte
//	B: Byte
// The unit that results in the smallest number greater than or equal to 1 is always chosen.
func HumanReadableByte(bytes int64) string {
	var unit string
	value := float32(bytes)

	switch {
	case bytes >= Terabyte:
		unit = "T"
		value = value / Terabyte
	case bytes >= Gigabyte:
		unit = "G"
		value = value / Gigabyte
	case bytes >= Megabyte:
		unit = "M"
		value = value / Megabyte
	case bytes >= Kilobyte:
		unit = "K"
		value = value / Kilobyte
	case bytes >= Byte:
		unit = "B"
	case bytes == 0:
		return "0"
	}

	v := fmt.Sprintf("%.1f", value)
	v = strings.TrimSuffix(v, ".0")
	return v + unit
}
