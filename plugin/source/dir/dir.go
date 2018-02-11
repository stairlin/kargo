package dir

import (
	"bytes"
	"fmt"
	"io"
	"os"
	"os/exec"
	"path"
	"path/filepath"
	"strings"

	"github.com/pkg/errors"
	"github.com/stairlin/kargo/context"
	"github.com/stairlin/kargo/plugin/source"
)

const name = "dir"

var (
	execTar, _ = exec.LookPath("tar")
)

func init() {
	source.Add(name, func() source.Source {
		return &Source{}
	})
}

type Source struct {
	Path string `toml:"path"`
}

func (s *Source) Name() string {
	return name
}

func (s *Source) Init() error {
	_, err := exec.LookPath("tar")
	if err != nil {
		return errors.Wrap(err, "tar not found. Install it or check your $PATH")
	}

	s.Path = strings.TrimSpace(s.Path)
	if s.Path == "" {
		return errors.New("dir: missing path")
	}
	return nil
}

func (s *Source) Backup(ctx *context.Context) (io.ReadCloser, error) {
	// Clean source path
	src := s.Path
	if !strings.HasSuffix(src, "/") {
		src += "/"
	}

	// Create tarball
	tarball := path.Join(ctx.Workdir, "backup-"+ctx.UUID+".tar")
	var args []string
	args = append(args, "-cvf", tarball, "-C", src, ".")
	cmd := exec.Command(execTar, args...)
	var out bytes.Buffer
	var stderr bytes.Buffer
	cmd.Stdout = &out
	cmd.Stderr = &stderr
	if err := cmd.Run(); err != nil {
		ctx.Error(fmt.Sprint(err) + ": " + stderr.String())
		return nil, errors.Wrap(s.parseError(err), "tar error")
	}

	// Ensure correct chmod
	os.Chmod(tarball, os.ModePerm)

	// Open backup file
	bak, err := os.Open(tarball)
	if err != nil {
		return nil, errors.Wrap(err, "cannot open backup file")
	}
	ctx.AddFile(bak)
	return bak, nil
}

func (s *Source) Restore(ctx *context.Context, r io.Reader) error {
	f, err := ctx.CreateTempFile(r)
	if err != nil {
		return err
	}
	backupPath, err := filepath.Abs(f.Name())
	if err != nil {
		return errors.Wrap(err, "cannot get backup absolute path")
	}

	// Untar to path
	cmd := exec.Command(execTar, "-xvf", backupPath, "-C", s.Path)
	var out bytes.Buffer
	var stderr bytes.Buffer
	cmd.Stdout = &out
	cmd.Stderr = &stderr
	if err := cmd.Run(); err != nil {
		ctx.Error(fmt.Sprint(err) + ": " + stderr.String())
		return errors.Wrap(s.parseError(err), "tar error")
	}

	return nil
}

func (s *Source) parseError(err error) error {
	return errors.Wrap(
		err,
		fmt.Sprintf("operation error"),
	)
}
