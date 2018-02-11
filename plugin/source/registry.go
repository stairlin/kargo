package source

import (
	"io"

	"github.com/stairlin/kargo/context"
)

type Source interface {
	Name() string
	Init() error
	Backup(*context.Context) (io.ReadCloser, error)
	Restore(*context.Context, io.Reader) error
}

type Creator func() Source

var Sources = map[string]Creator{}

func Add(name string, creator Creator) {
	Sources[name] = creator
}
