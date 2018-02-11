package cipher

import (
	"encoding/base64"
	"io"

	"github.com/pkg/errors"
	"github.com/stairlin/kargo/context"
	"github.com/stairlin/kargo/pkg/sec"
	"github.com/stairlin/kargo/pkg/unit"
	"github.com/stairlin/kargo/plugin/process"
)

const (
	name = "cipher"
)

var (
	// contains the maximum amount of data that can be encrypted at once
	blockSize = unit.KB * 32
)

func init() {
	process.Add(name, func() process.Processor {
		return &Processor{}
	})
}

// Processor is a cipher processor
type Processor struct {
	Keys    []string `toml:"keys"`
	Default uint32   `toml:"default"`

	Rotator *sec.Rotator
}

func (p *Processor) Name() string {
	return name
}

func (p *Processor) Init() error {
	if len(p.Keys) < 1 {
		return errors.New("there must be at lest one encryption key")
	}
	if int(p.Default) >= len(p.Keys) {
		return errors.New("invalid default key index")
	}

	// Decode keys
	keys := map[uint32][]byte{}
	for i, key := range p.Keys {
		decodedKey, err := base64.StdEncoding.DecodeString(key)
		if err != nil {
			return errors.Wrapf(err, "cannot decode key #%d", i)
		}
		if len(decodedKey) != sec.KeySize {
			return errors.Errorf(
				"invalid encryption key length %d != %d", len(decodedKey), sec.KeySize,
			)
		}
		keys[uint32(i)] = decodedKey
	}

	p.Rotator = sec.NewRotator(keys, p.Default)
	return nil
}

// Encode encrypt data from r
func (p *Processor) Encode(
	ctx *context.Context, r io.Reader,
) (io.ReadCloser, error) {
	return p.Rotator.EncryptReader(r, int(blockSize))
}

// Decode decrypt data from r
func (p *Processor) Decode(
	ctx *context.Context, r io.Reader,
) (io.ReadCloser, error) {
	return p.Rotator.DecryptReader(r, int(blockSize))
}
