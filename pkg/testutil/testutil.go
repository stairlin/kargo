package testutil

import (
	"io/ioutil"
	"math/rand"
	"testing"
)

func GenRandBytes(t *testing.T, l int) []byte {
	b := make([]byte, l)
	if _, err := rand.Read(b); err != nil {
		t.Fatal(err)
	}
	return b
}

func Truncate(b []byte, n int) string {
	if len(b) > n {
		// It would be better to truncate based on runes, but that's just for
		// test output
		return string(append(b[:n], []byte("...")...))
	}
	return string(b)
}

func TempDir(t *testing.T, prefix string) string {
	dir, err := ioutil.TempDir(".", prefix)
	if err != nil {
		t.Fatal(err)
	}
	return dir
}
