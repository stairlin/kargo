package unit

import (
	"fmt"
	"math"
	"strconv"
	"strings"

	"github.com/pkg/errors"
)

// Byte represents a byte size
type Byte float64

const (
	_ = iota
	// KB means kilobyte
	KB Byte = 1 << (10 * iota)
	// MB means megabyte
	MB
	// GB means gigabyte
	GB
	// TB means terabyte
	TB
)

// ParseByte parses the string and return a Byte struct
//
// e.g. 3.3 kB, 3kB, 7 GB, 9 TB, -1 TB
func ParseByte(s string) (Byte, error) {
	s = strings.TrimSpace(s)
	if len(s) < 3 {
		return 0, fmt.Errorf("invalid format")
	}

	// Either we have a space in between (or not)
	res := strings.Split(s, " ")
	var n, scale string
	if len(res) == 2 {
		n = res[0]
		scale = res[1]
	} else {
		n = s[0 : len(s)-2]
		scale = s[len(s)-2:]
	}

	var mag Byte
	switch scale {
	case "B":
		mag = 1
	case "kB":
		mag = KB
	case "MB":
		mag = MB
	case "GB":
		mag = GB
	case "TB":
		mag = TB
	default:
		return 0, fmt.Errorf("unknown scale %s", scale)
	}

	v, err := strconv.ParseFloat(n, 64)
	if err != nil {
		return 0, errors.Wrap(err, "invalid number")
	}

	return Byte(v) * mag, nil
}

func (b Byte) String() string {
	switch {
	case math.Abs(float64(b/TB)) >= 1:
		return strconv.FormatFloat(float64(b/TB), 'f', -1, 64) + " TB"
	case math.Abs(float64(b/GB)) >= 1:
		return strconv.FormatFloat(float64(b/GB), 'f', -1, 64) + " GB"
	case math.Abs(float64(b/MB)) >= 1:
		return strconv.FormatFloat(float64(b/MB), 'f', -1, 64) + " MB"
	case math.Abs(float64(b/KB)) >= 1:
		return strconv.FormatFloat(float64(b/KB), 'f', -1, 64) + " kB"
	}

	return strconv.FormatFloat(float64(b), 'f', -1, 64) + " B"
}

// Gt returns whether the value b is greater than v
func (b Byte) Gt(v int64) bool {
	return b > Byte(v)
}

// Lt returns whether the value b is less than v
func (b Byte) Lt(v int64) bool {
	return b < Byte(v)
}

// MarshalJSON implements the json.Marshaler interface.
func (b Byte) MarshalJSON() ([]byte, error) {
	return []byte(strings.Join([]string{"\"", b.String(), "\""}, "")), nil
}

// UnmarshalJSON implements the json.Unmarshaler interface.
func (b *Byte) UnmarshalJSON(data []byte) error {
	r, err := ParseByte(strings.Trim(string(data), "\""))
	if err != nil {
		return err
	}
	*b = r
	return nil
}
