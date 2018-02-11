package influxdb

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

const name = "influxdb"

var (
	execInfluxd, _ = exec.LookPath("influxd")
	execTar, _     = exec.LookPath("tar")
)

func init() {
	source.Add(name, func() source.Source {
		return &Source{}
	})
}

// Source is an InfluxDB source
type Source struct {
	Host     string `toml:"host"`
	Port     string `toml:"port"`
	User     string `toml:"user"`
	Password string `toml:"password"`
	DB       string `toml:"db"`
	Metadir  string `toml:"metadir"`
	Datadir  string `toml:"datadir"`
}

func (s *Source) Name() string {
	return name
}

func (s *Source) Init() error {
	if _, err := exec.LookPath("influxd"); err != nil {
		return errors.Wrap(err, "influxd not found. Install it or check your $PATH")
	}
	if _, err := exec.LookPath("tar"); err != nil {
		return errors.Wrap(err, "tar not found. Install it or check your $PATH")
	}
	return nil
}

func (s *Source) Backup(ctx *context.Context) (io.ReadCloser, error) {
	// influxd backup -database mydatabase -host 10.0.0.0:8088 /tmp/mysnapshot
	args := []string{"backup"}
	if s.DB != "" {
		args = append(args, "-database", s.DB)
	}
	if s.Host != "" {
		host := fmt.Sprintf("%s:%s", s.Host, s.Port)
		args = append(args, "-host", host)
	}

	// Create temp path
	dest := ctx.TempPath()
	os.MkdirAll(dest, 0770)
	args = append(args, dest)

	// Start backup
	cmd := exec.Command(execInfluxd, args...)
	var out bytes.Buffer
	var stderr bytes.Buffer
	cmd.Stdout = &out
	cmd.Stderr = &stderr
	if err := cmd.Run(); err != nil {
		fmt.Println(fmt.Sprint(err) + ": " + stderr.String())
		return nil, errors.Wrap(s.parseError(err), "influxd error")
	}

	// Create tarball
	tarball := dest + ".tar"
	cmd = exec.Command(execTar, "-cvf", tarball, "-C", dest, ".")
	cmd.Stdout = &out
	cmd.Stderr = &stderr
	if err := cmd.Run(); err != nil {
		fmt.Println(fmt.Sprint(err) + ": " + stderr.String())
		return nil, errors.Wrap(s.parseError(err), "tar error")
	}

	// Open backup file
	bak, err := os.Open(tarball)
	if err != nil {
		return nil, errors.Wrap(err, "cannot open backup file")
	}
	return bak, nil
}

func (s *Source) Restore(ctx *context.Context, r io.Reader) error {
	if s.Metadir == "" {
		return errors.New("metadir is undefined")
	}
	if s.Datadir == "" {
		return errors.New("datadir is undefined")
	}
	if s.DB == "" {
		return errors.New("db is undefined")
	}

	f, err := ctx.CreateTempFile(r)
	if err != nil {
		return err
	}
	backupPath, err := filepath.Abs(f.Name())
	if err != nil {
		return errors.Wrap(err, "cannot get backup absolute path")
	}

	// Create temporary folder to untar backup
	dest := ctx.TempPath()
	os.MkdirAll(dest, 0770)

	// Untar file
	cmd := exec.Command(execTar, "-xvf", backupPath, "-C", dest)
	var out bytes.Buffer
	var stderr bytes.Buffer
	cmd.Stdout = &out
	cmd.Stderr = &stderr
	if err := cmd.Run(); err != nil {
		fmt.Printf("\tSrc: %s", backupPath)
		fmt.Printf("\tDest: %s", dest)
		fmt.Println(fmt.Sprint(err) + ": " + stderr.String())
		return errors.Wrap(s.parseError(err), "tar error")
	}

	// Start restoring metastore
	// influxd restore -metadir /var/lib/influxdb/meta /path/to/mysnapshot
	args := []string{
		"restore",
		"-metadir", s.Metadir,
		dest,
	}
	cmd = exec.Command(execInfluxd, args...)
	cmd.Stdout = &out
	cmd.Stderr = &stderr
	if err := cmd.Run(); err != nil {
		fmt.Printf("%s - %s\n", execInfluxd, args)
		fmt.Println(fmt.Sprint(err) + ": " + stderr.String())
		return errors.Wrap(s.parseError(err), "influxd restore metastore error")
	}

	// Then restore databases
	// influxd restore -database foo -datadir /path/to/mysnapshot
	args = []string{
		"restore",
		"-database", s.DB,
		"-datadir", s.Datadir,
		dest,
	}
	cmd = exec.Command(execInfluxd, args...)
	cmd.Stdout = &out
	cmd.Stderr = &stderr
	if err := cmd.Run(); err != nil {
		fmt.Printf("%s - %s\n", execInfluxd, args)
		fmt.Println(fmt.Sprint(err) + ": " + stderr.String())
		return errors.Wrap(s.parseError(err), "influxd restore database error")
	}

	fmt.Println("Note:")
	fmt.Println(" The permissions on the shards may no longer be accurate.")
	fmt.Println(" To ensure the file permissions are correct, please run:")
	fmt.Printf("\tsudo chown -R influxdb:influxdb %s\n\n", path.Dir(s.Metadir))

	return nil
}

func (s *Source) parseError(err error) error {
	es := err.Error()
	if strings.Contains(es, "no backup files for") {
		return errors.Wrap(
			err,
			fmt.Sprintf("there is no backup data for the given database"),
		)
	}
	return errors.Wrap(
		err,
		fmt.Sprintf("operation error"),
	)
}
