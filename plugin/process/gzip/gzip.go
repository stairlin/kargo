package gzip

import (
	"compress/gzip"
	"io"

	"github.com/pkg/errors"
	"github.com/stairlin/kargo/context"
	"github.com/stairlin/kargo/log"
	"github.com/stairlin/kargo/plugin/process"
)

const (
	name         = "gzip"
	defaultLevel = gzip.BestSpeed
)

func init() {
	process.Add(name, func() process.Processor {
		return &Processor{}
	})
}

// Processor is a GZIP processor
type Processor struct{}

func (p *Processor) Name() string {
	return name
}

func (p *Processor) Init() error {
	return nil
}

// Encode compresses data from r
func (p *Processor) Encode(
	ctx *context.Context, r io.Reader,
) (io.ReadCloser, error) {
	out, in := io.Pipe()

	go func() {
		defer in.Close()

		gzip, err := gzip.NewWriterLevel(in, defaultLevel)
		if err != nil {
			ctx.Error(
				"Error creating gzip reader", log.String("proc", "gzip"), log.Error(err),
			)
			return
		}
		defer gzip.Close()
		if _, err := io.Copy(gzip, r); err != nil {
			ctx.Error(
				"Error compressing reader", log.String("proc", "gzip"), log.Error(err),
			)
			return
		}
	}()

	return out, nil
}

// Decode decompresses data from r
func (p *Processor) Decode(
	ctx *context.Context, r io.Reader,
) (io.ReadCloser, error) {
	out, err := gzip.NewReader(r)
	if err != nil {
		return nil, errors.Wrap(err, "gzip: error decompressing data")
	}
	return out, nil
}
