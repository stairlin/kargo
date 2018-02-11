// Package fs stores backups on the filesystem
package fs

import (
	"bufio"
	"io"
	"os"
	"path"
	"path/filepath"

	"github.com/pkg/errors"
	"github.com/stairlin/kargo/context"
	"github.com/stairlin/kargo/plugin/storage"
)

const name = "fs"

func init() {
	storage.Add(name, func() storage.Storage {
		return &Store{}
	})
}

type Store struct {
	Path string `toml:"path"`
}

func (s *Store) Name() string {
	return name
}

func (s *Store) Init() error {
	path := filepath.Clean(s.Path)
	if err := os.MkdirAll(path, 0770); err != nil {
		return errors.Wrapf(err, "cannot create fs storage base path '%s'", path)
	}
	return nil
}

func (s *Store) Info(ctx *context.Context, key string) (os.FileInfo, error) {
	info, err := os.Stat(path.Join(s.Path, key))
	switch {
	case err == nil:
		return info, nil
	case os.IsNotExist(err):
		return nil, storage.ErrKeyNotFound
	}
	return nil, errors.Wrap(err, "cannot open file")
}

func (s *Store) Push(ctx *context.Context, key string, r io.Reader) error {
	f, err := os.Create(path.Join(s.Path, key))
	if err != nil {
		return errors.Wrap(err, "cannot create file")
	}
	defer f.Close()

	bufw := bufio.NewWriter(f)
	defer bufw.Flush()

	_, err = io.Copy(bufw, r)
	switch err {
	case nil, io.ErrUnexpectedEOF:
		return nil
	}
	return err
}

func (s *Store) Pull(
	ctx *context.Context, key string,
) (io.ReadCloser, os.FileInfo, error) {
	f, err := os.Open(path.Join(s.Path, key))
	switch {
	case err == nil:
	case os.IsNotExist(err):
		return nil, nil, storage.ErrKeyNotFound
	default:
		return nil, nil, errors.Wrap(err, "cannot open file")
	}
	info, err := f.Stat()
	if err != nil {
		return nil, nil, err
	}
	return f, info, nil
}

func (s *Store) Walk(
	ctx *context.Context, walkFn func(key string, f os.FileInfo, err error) error,
) {
	filepath.Walk(s.Path, func(path string, f os.FileInfo, err error) error {
		if f.IsDir() {
			return nil
		}
		key, err := filepath.Rel(s.Path, path)
		if err != nil {
			return err
		}
		return walkFn(key, f, err)
	})
}
