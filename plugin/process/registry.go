package process

import (
	"io"

	"github.com/stairlin/kargo/context"
)

type Processor interface {
	Name() string
	Init() error
	Encode(ctx *context.Context, r io.Reader) (io.ReadCloser, error)
	Decode(ctx *context.Context, r io.Reader) (io.ReadCloser, error)
}

type Creator func() Processor

var Processors = map[string]Creator{}

func Add(name string, creator Creator) {
	Processors[name] = creator
}
