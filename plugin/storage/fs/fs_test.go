package fs_test

import (
	"bytes"
	"io/ioutil"
	"math"
	"os"
	"strconv"
	"testing"

	"github.com/stairlin/kargo/context"
	"github.com/stairlin/kargo/pkg/testutil"
	"github.com/stairlin/kargo/pkg/unit"
	"github.com/stairlin/kargo/plugin/storage"
	"github.com/stairlin/kargo/plugin/storage/fs"
)

func TestVerbatim(t *testing.T) {
	dir := testutil.TempDir(t, "fs")
	defer os.RemoveAll(dir)

	store := &fs.Store{
		Path: dir,
	}
	if err := store.Init(); err != nil {
		t.Fatal(err)
	}

	expect := testutil.GenRandBytes(t, int(64*unit.MB))
	input := bytes.NewReader(expect)

	ctx := context.Background()

	// Ensure the key does not exist
	_, _, err := store.Pull(ctx, "foo")
	if storage.ErrKeyNotFound != err {
		t.Errorf("expect error %s, but got %s", storage.ErrKeyNotFound, err)
	}

	// Push/Pull
	if err := store.Push(ctx, "foo", input); err != nil {
		t.Fatal("Error pushing data to storage", err)
	}
	out, _, err := store.Pull(ctx, "foo")
	if err != nil {
		t.Fatal("Error pulling data from storage", err)
	}

	// Tests
	got, err := ioutil.ReadAll(out)
	if err != nil {
		t.Fatal(err)
	}
	if len(expect) != len(got) {
		t.Errorf("expect length of %d, but got %d", len(expect), len(got))
	}
	if string(expect) != string(got) {
		t.Errorf("expect text %s, but got %s",
			testutil.Truncate(expect, 140), testutil.Truncate(got, 140),
		)
	}

	// Close reader
	if err := out.Close(); err != nil {
		t.Error("output close err", err)
	}
}

func TestKeyNotFound(t *testing.T) {
	dir := testutil.TempDir(t, "fs")
	defer os.RemoveAll(dir)

	store := &fs.Store{
		Path: dir,
	}
	if err := store.Init(); err != nil {
		t.Fatal(err)
	}

	ctx := context.Background()
	_, _, err := store.Pull(ctx, "not_found")
	if storage.ErrKeyNotFound != err {
		t.Errorf("expect error %s, but got %s", storage.ErrKeyNotFound, err)
	}
}

func TestWalk(t *testing.T) {
	dir := testutil.TempDir(t, "fs")
	defer os.RemoveAll(dir)
	store := &fs.Store{
		Path: dir,
	}
	if err := store.Init(); err != nil {
		t.Fatal(err)
	}
	paths := map[string]bool{}

	ctx := context.Background()

	expect := 0
	count := 0
	filter := &storage.WalkFilter{
		To: int64(math.MaxInt64),
	}
	store.Walk(ctx, filter, func(path string, f os.FileInfo, err error) error {
		count++
		return nil
	})
	if expect != count {
		t.Errorf("expect walk to be called %d times, but got %d", expect, count)
	}

	// Push first key
	input := bytes.NewReader(testutil.GenRandBytes(t, int(24*unit.KB)))
	if err := store.Push(ctx, "foo", input); err != nil {
		t.Fatal("Error pushing data to storage", err)
	}
	paths["foo"] = false

	expect = 1
	count = 0
	filter = &storage.WalkFilter{
		To: int64(math.MaxInt64),
	}
	store.Walk(ctx, filter, func(path string, f os.FileInfo, err error) error {
		count++
		return nil
	})
	if expect != count {
		t.Errorf("expect walk to be called %d times, but got %d", expect, count)
	}

	// Push more keys
	for i := 0; i < 50; i++ {
		input := bytes.NewReader(testutil.GenRandBytes(t, (i+1)*int(unit.KB)))
		key := strconv.Itoa(i)
		paths[key] = false
		if err := store.Push(ctx, key, input); err != nil {
			t.Fatal("Error pushing data to storage", err)
		}
	}

	expect = 51
	count = 0
	filter = &storage.WalkFilter{
		To: int64(math.MaxInt64),
	}
	store.Walk(ctx, filter, func(path string, f os.FileInfo, err error) error {
		if checked, exists := paths[path]; exists {
			if checked {
				t.Errorf("key %s already checked", path)
			}
			paths[path] = true
		} else {
			t.Errorf("key %s should not exist", path)
		}
		count++
		return nil
	})
	if expect != count {
		t.Errorf("expect walk to be called %d times, but got %d", expect, count)
	}
}
