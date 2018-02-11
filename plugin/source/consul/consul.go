package consul

import (
	"bytes"
	"fmt"
	"io"
	"net/url"
	"os"
	"os/exec"
	"path"
	"path/filepath"

	"github.com/pkg/errors"
	"github.com/stairlin/kargo/context"
	"github.com/stairlin/kargo/plugin/source"
)

const name = "consul"

var execConsul, _ = exec.LookPath("consul")

func init() {
	source.Add(name, func() source.Source {
		return &Source{}
	})
}

// Source is a Consul source
type Source struct {
	CAFile        string `toml:"ca_file"`
	CAPath        string `toml:"ca_path"`
	ClientCert    string `toml:"client_cert"`
	ClientKey     string `toml:"client_key"`
	TLSServerName string `toml:"tls_server_name"`
	Token         string `toml:"token"`
	Datacenter    string `toml:"datacenter"`
	Stale         bool   `toml:"stale"`
	HTTPAddr      string `toml:"http_addr"`
}

func (s *Source) Name() string {
	return name
}

func (s *Source) Init() error {
	// Ensure consul is present
	_, err := exec.LookPath("consul")
	if err != nil {
		return errors.Wrap(err, "consul not found. Install it or check your $PATH")
	}

	if s.HTTPAddr != "" {
		url, err := url.Parse(s.HTTPAddr)
		if err != nil {
			return errors.Wrapf(err, "cannot parse url '%s'", s.HTTPAddr)
		}
		if url.Port() == "" {
			return errors.New("http-addr must contain a port")
		}
		s.HTTPAddr = url.String()
	}
	return nil
}

func (s *Source) Backup(ctx *context.Context) (io.ReadCloser, error) {
	// consul snapshot save [options] FILE
	args := []string{"snapshot", "save"}
	args = append(args, s.buildArgs()...)

	// Prepare backup destination
	dest := ctx.TempPath()
	os.MkdirAll(dest, 0770)
	snapshot := path.Join(dest, "backup.snap")
	args = append(args, snapshot)

	// Start backup
	cmd := exec.Command(execConsul, args...)
	var out bytes.Buffer
	var stderr bytes.Buffer
	cmd.Stdout = &out
	cmd.Stderr = &stderr
	if err := cmd.Run(); err != nil {
		fmt.Println(fmt.Sprint(err) + ": " + stderr.String())
		return nil, errors.Wrap(s.parseError(err), "consul error")
	}

	// Open backup file
	backup, err := os.Open(snapshot)
	if err != nil {
		return nil, errors.Wrap(err, "cannot open backup file")
	}
	return backup, nil
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

	// consul snapshot restore [options] FILE
	args := []string{"snapshot", "restore"}
	args = append(args, s.buildArgs()...)
	args = append(args, backupPath)

	// Start restore
	cmd := exec.Command(execConsul, args...)
	var out bytes.Buffer
	var stderr bytes.Buffer
	cmd.Stdout = &out
	cmd.Stderr = &stderr
	if err := cmd.Run(); err != nil {
		fmt.Println(fmt.Sprint(err) + ": " + stderr.String())
		return errors.Wrap(s.parseError(err), "consul error")
	}

	return nil
}

func (s *Source) buildArgs() []string {
	var args []string
	if s.CAFile != "" {
		args = append(args, "-ca-file", s.CAFile)
	}
	if s.CAPath != "" {
		args = append(args, "-ca-path", s.CAPath)
	}
	if s.ClientCert != "" {
		args = append(args, "-client-cert", s.ClientCert)
	}
	if s.ClientKey != "" {
		args = append(args, "-client-cert", s.ClientKey)
	}
	if s.HTTPAddr != "" {
		args = append(args, "-http-addr", s.HTTPAddr)
	}
	if s.TLSServerName != "" {
		args = append(args, "-tls-server-name", s.TLSServerName)
	}
	if s.Token != "" {
		args = append(args, "-token", s.Token)
	}
	if s.Datacenter != "" {
		args = append(args, "-datacenter", s.Datacenter)
	}
	if s.Stale {
		args = append(args, "-stale")
	}
	return args
}

func (s *Source) parseError(err error) error {
	return errors.Wrap(
		err,
		fmt.Sprintf("operation error"),
	)
}
