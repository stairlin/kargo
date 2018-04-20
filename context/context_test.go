package context_test

import (
	"bytes"
	"io/ioutil"
	"testing"

	"github.com/stairlin/kargo/context"
)

func TestCreateTempFile(t *testing.T) {
	expect := []byte("foo")
	c := context.Background()
	defer c.Cleanup()

	f, err := c.CreateTempFile(bytes.NewReader(expect))
	if err != nil {
		t.Fatal(err)
	}
	got, err := ioutil.ReadAll(f)
	if err != nil {
		t.Fatal(err)
	}
	if string(expect) != string(got) {
		t.Errorf("expect %s, but got %s", expect, got)
	}
}
