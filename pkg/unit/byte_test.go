package unit_test

import (
	"testing"

	"github.com/stairlin/kargo/pkg/unit"
)

func TestByteSizes(t *testing.T) {
	t.Parallel()

	table := []struct {
		res    unit.Byte
		expect float64
	}{
		{res: unit.KB * 3, expect: 3072},
		{res: unit.MB * 5, expect: 5242880},
		{res: unit.GB * 7, expect: 7516192768},
		{res: unit.TB * 9, expect: 9895604649984},
		{res: unit.TB * 0, expect: 0},
		{res: unit.TB * 1, expect: 1099511627776},
		{res: unit.TB * -1, expect: -1099511627776},
	}

	for _, test := range table {
		if float64(test.res) != test.expect {
			t.Errorf("expect to get %f, but got %f", test.expect, test.res)
		}
	}
}

func TestByteParse(t *testing.T) {
	t.Parallel()

	table := []struct {
		in     string
		expect unit.Byte
		err    error
	}{
		{in: "3 kB", expect: unit.KB * 3},
		{in: "3.3 kB", expect: unit.KB * 3.3},
		{in: "3kB", expect: unit.KB * 3},
		{in: "3.3kB", expect: unit.KB * 3.3},
		{in: "5 MB", expect: unit.MB * 5},
		{in: "7 GB", expect: unit.GB * 7},
		{in: "9 TB", expect: unit.TB * 9},
		{in: "0 B", expect: unit.TB * 0},
		{in: "1 TB", expect: unit.TB * 1},
		{in: "-1 TB", expect: unit.TB * -1},
	}

	for _, test := range table {
		res, err := unit.ParseByte(test.in)
		if err != test.err {
			t.Errorf("expect to get error %s, but got %s", test.err, err)
		}
		if err != nil {
			continue
		}

		if res != test.expect {
			t.Errorf("expect to get %f, but got %f", test.expect, res)
		}
	}
}

func TestByteStringFormat(t *testing.T) {
	t.Parallel()

	table := []struct {
		in     unit.Byte
		expect string
	}{
		{in: unit.KB * 3, expect: "3 kB"},
		{in: unit.KB * 3.3, expect: "3.3 kB"},
		{in: unit.MB * 5, expect: "5 MB"},
		{in: unit.GB * 7, expect: "7 GB"},
		{in: unit.TB * 9, expect: "9 TB"},
		{in: unit.TB * 0, expect: "0 B"},
		{in: unit.TB * 1, expect: "1 TB"},
		{in: unit.TB * -1, expect: "-1 TB"},
	}

	for _, test := range table {
		res := test.in.String()
		if res != test.expect {
			t.Errorf("expect to get %s, but got %s", test.expect, res)
		}
	}
}
