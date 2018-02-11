package couchbase

import (
	"bytes"
	"fmt"
	"io"
	"io/ioutil"
	"os"
	"os/exec"
	"path"
	"path/filepath"
	"strings"

	"github.com/pkg/errors"
	"github.com/stairlin/kargo/context"
	"github.com/stairlin/kargo/plugin/source"
)

const name = "couchbase"

var (
	execCbbackup, _  = exec.LookPath("cbbackup")
	execCbrestore, _ = exec.LookPath("cbrestore")
	execTar, _       = exec.LookPath("tar")
)

func init() {
	source.Add(name, func() source.Source {
		return &Source{
			Host: "127.0.0.1",
			Port: "8091",
		}
	})
}

// Source is a Couchbase source
type Source struct {
	Host       string `toml:"host"`
	Port       string `toml:"port"`
	User       string `toml:"user"`
	Password   string `toml:"password"`
	DB         string `toml:"db"`
	SingleNode bool   `toml:"single_node"`
	Rehash     bool   `toml:"rehash"`
}

func (s *Source) Name() string {
	return name
}

func (s *Source) Init() error {
	// Ensure cbbackup/tar are present
	if _, err := exec.LookPath("cbbackup"); err != nil {
		return errors.Wrap(err, "cbbackup not found. Install it or check your $PATH")
	}
	if _, err := exec.LookPath("cbrestore"); err != nil {
		return errors.Wrap(err, "cbrestore not found. Install it or check your $PATH")
	}
	if _, err := exec.LookPath("tar"); err != nil {
		return errors.Wrap(err, "tar not found. Install it or check your $PATH")
	}
	if s.Host == "" {
		return errors.New("couchbase: missing host key")
	}
	return nil
}

func (s *Source) Backup(ctx *context.Context) (io.ReadCloser, error) {
	// cbbackup http://HOST:8091 /backup-42 -u Administrator -p password
	node := fmt.Sprintf("http://%s:%s", s.Host, s.Port)

	// Create temp path
	dest := ctx.TempPath()
	os.MkdirAll(dest, 0770)

	// Start backup
	args := []string{node, dest, "-u", s.User, "-p", s.Password}
	if s.SingleNode {
		args = append(args, "--single-node")
	}
	cmd := exec.Command(execCbbackup, args...)
	var out bytes.Buffer
	var stderr bytes.Buffer
	cmd.Stdout = &out
	cmd.Stderr = &stderr
	if err := cmd.Run(); err != nil {
		fmt.Println(fmt.Sprint(err) + ": " + stderr.String())
		return nil, errors.Wrap(s.parseError(err), "cbbackup error")
	}

	// Get backup directory
	files, err := ioutil.ReadDir(dest)
	if err != nil {
		return nil, errors.Wrap(err, "error listing backup folder")
	}
	if len(files) == 0 {
		return nil, errors.New("backup file not found")
	}
	if len(files) > 1 {
		return nil, errors.New("multiple (potential) backups found")
	}
	backupRootDir := files[0].Name()
	files, err = ioutil.ReadDir(path.Join(dest, backupRootDir))
	if err != nil {
		return nil, errors.Wrap(err, "error listing backup folder")
	}
	if len(files) == 0 {
		return nil, errors.New("backup file not found")
	}
	if len(files) > 1 {
		return nil, errors.New("multiple (potential) backups found")
	}
	backupSnapshot := files[0].Name()

	// Create tarball
	dir := path.Join(dest, backupRootDir, backupSnapshot)
	tarball := path.Join(dest, backupSnapshot+".tar")
	cmd = exec.Command(execTar, "-cvf", tarball, "-C", dir, ".")
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
		fmt.Println(fmt.Sprint(err) + ": " + stderr.String())
		return errors.Wrap(s.parseError(err), "tar error")
	}

	// Build params
	// cbrestore [options] [backup-dir] [destination]
	db := fmt.Sprintf("http://%s:%s", s.Host, s.Port)

	// Get all bucket names
	buckets, err := ioutil.ReadDir(dest)
	if err != nil {
		return errors.Wrap(err, "error listing buckets folder")
	}
	if len(buckets) == 0 {
		return errors.New("buckets folder not found")
	}

	fmt.Println("Buckets to restore:")
	for _, bucket := range buckets {
		fmt.Println("\t- ", bucket.Name())
	}

	for _, bucket := range buckets {
		b := bucket.Name()
		if !strings.HasPrefix(b, "bucket-") {
			continue
		}
		b = strings.TrimPrefix(b, "bucket-")
		args := []string{dest, db, "-u", s.User, "-p", s.Password, "-b", b}
		if s.Rehash {
			args = append(args, "-x", "rehash=1")
		}

		// Output command
		fmt.Println(execCbrestore, args)

		// Start restore
		cmd = exec.Command(execCbrestore, args...)
		cmd.Stdout = &out
		cmd.Stderr = &stderr
		if err := cmd.Run(); err != nil {
			fmt.Println(fmt.Sprint(err) + ": " + stderr.String())
			return errors.Wrap(s.parseError(err), "cbrestore error")
		}
	}

	return nil
}

func (s *Source) parseError(err error) error {
	return errors.Wrap(
		err,
		fmt.Sprintf("operation error"),
	)
}
