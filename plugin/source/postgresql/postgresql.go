package postgresql

import (
	"fmt"
	"io"
	"os"
	"os/exec"
	"path/filepath"
	"strings"

	"github.com/pkg/errors"

	"github.com/stairlin/kargo/context"
	"github.com/stairlin/kargo/plugin/source"
)

const name = "postgresql"

var (
	execPGDump, _    = exec.LookPath("pg_dump")
	execPGRestore, _ = exec.LookPath("pg_restore")
)

func init() {
	source.Add(name, func() source.Source {
		return &Source{}
	})
}

// Source is a PostgreSQL source
type Source struct {
	Host     string `toml:"host"`
	Port     string `toml:"port"`
	User     string `toml:"user"`
	Password string `toml:"password"`
	DB       string `toml:"db"`
}

func (s *Source) Name() string {
	return name
}

func (s *Source) Init() error {
	if _, err := exec.LookPath("pg_dump"); err != nil {
		return errors.Wrap(err, "pg_dump not found. Install it or check your $PATH")
	}
	if _, err := exec.LookPath("pg_restore"); err != nil {
		return errors.Wrap(err, "pg_restore not found. Install it or check your $PATH")
	}
	return nil
}

func (s *Source) Backup(ctx *context.Context) (io.ReadCloser, error) {
	// Build command line
	format := "-Fc"
	dbname := fmt.Sprintf("--dbname=postgresql://%s:%s@%s:%s/%s",
		s.User,
		s.Password,
		s.Host,
		s.Port,
		s.DB,
	)
	args := []string{format, dbname}

	// Start backup
	out, in := io.Pipe()
	ctx.AddCloser(in)
	cmd := exec.Command(execPGDump, args...)
	cmd.Stderr = os.Stdout
	cmd.Stdout = in
	if err := cmd.Run(); err != nil {
		return nil, s.parseError(err)
	}

	return out, nil
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

	// Build params
	host := fmt.Sprintf("-h%s", s.Host)
	port := fmt.Sprintf("-p%s", s.Port)
	user := fmt.Sprintf("-U%s", s.User)
	db := fmt.Sprintf("--dbname=%s", s.DB)
	noPwd := "--no-password"
	format := "-Fc"
	args := []string{host, port, user, db, noPwd, format, backupPath}

	// Output command
	fmt.Println(fmt.Sprintf("PGPASSWORD=%d", len(s.Password)), execPGRestore, args)

	// Start restore
	cmd := exec.Command(execPGRestore, args...)
	cmd.Env = append(cmd.Env, fmt.Sprintf("PGPASSWORD=%s", s.Password))
	cmd.Stdout = os.Stdout
	cmd.Stderr = os.Stdout
	if err := cmd.Run(); err != nil {
		return s.parseError(err)
	}

	return nil
}

func (s *Source) parseError(err error) error {
	if strings.Contains(err.Error(), "no such file or directory") {
		return errors.Wrap(
			err,
			fmt.Sprintf("cannot reach %s:%s. Ensure that the server is running", s.Host, s.Port),
		)
	} else if strings.Contains(err.Error(), "server version mismatch") {
		return errors.Wrap(
			err,
			fmt.Sprintf("postgres client/server version mismatch"),
		)
	}

	return errors.Wrap(
		err,
		fmt.Sprintf("operation error"),
	)
}
