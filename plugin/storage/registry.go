package storage

import (
	"errors"
	"io"
	"os"

	"github.com/stairlin/kargo/context"
)

type Storage interface {
	Name() string
	// Init initialises the storage plugin. It should be called only once
	Init() error
	// Info returns information about a file
	Info(ctx *context.Context, key string) (os.FileInfo, error)
	// Push pushes data from r to the storage
	Push(ctx *context.Context, key string, r io.Reader) error
	// Pull pulls data from the storage and returns a reader
	Pull(ctx *context.Context, key string) (io.ReadCloser, os.FileInfo, error)
	// Walk walks the file tree rooted at root, calling walkFn for each file or
	// directory in the tree, including root. All errors that arise visiting
	// files and directories are filtered by walkFn. The files are walked in
	// lexical order, which makes the output deterministic.
	Walk(ctx *context.Context, f func(key string, f os.FileInfo, err error) error)
}

type Creator func() Storage

var Storages = map[string]Creator{}

func Add(name string, creator Creator) {
	Storages[name] = creator
}

var (
	// ErrKeyNotFound when a requested key does not exist in the storage
	ErrKeyNotFound = errors.New("the requested key does not exist")
)
