package foundationdb

import (
	"bytes"
	"fmt"
	"io"
	"io/ioutil"
	"os"
	"os/exec"
	"path"
	"path/filepath"
	"time"

	"github.com/pkg/errors"
	"github.com/stairlin/kargo/context"
	"github.com/stairlin/kargo/plugin/source"
)

const name = "foundationdb"

var (
	execFdbbackup, _   = exec.LookPath("fdbbackup")
	execFdbrestore, _  = exec.LookPath("fdbrestore")
	execBackupAgent, _ = exec.LookPath("backup_agent")
	execTar, _         = exec.LookPath("tar")
)

func init() {
	source.Add(name, func() source.Source {
		return &Source{}
	})
}

// Source is an InfluxDB source
type Source struct {
	Cluster     string `toml:"cluster"`
	Tag         string `toml:"tag"`
	FDBBackup   string `toml:"fdbbackup"`
	FDBRestore  string `toml:"fdbrestore"`
	BackupAgent string `toml:"backup_agent"`
}

func (s *Source) Name() string {
	return name
}

func (s *Source) Init() error {
	return nil
}

func (s *Source) Backup(ctx *context.Context) (io.ReadCloser, error) {
	backupAgent := s.BackupAgent
	if backupAgent == "" {
		if _, err := exec.LookPath("backup_agent"); err != nil {
			return nil, errors.Wrap(err, "backup_agent not found. Install it or check your $PATH")
		}
		backupAgent = execBackupAgent
	}
	fdbBackup := s.FDBBackup
	if fdbBackup == "" {
		if _, err := exec.LookPath("fdbbackup"); err != nil {
			return nil, errors.Wrap(err, "fdbbackup not found. Install it or check your $PATH")
		}
		fdbBackup = execFdbbackup
	}
	if _, err := exec.LookPath("tar"); err != nil {
		return nil, errors.Wrap(err, "tar not found. Install it or check your $PATH")
	}

	// Start backup agent
	args := []string{}
	if s.Cluster != "" {
		args = append(args, "-C", s.Cluster)
	}
	agent := agent{
		name: backupAgent,
		args: args,
	}
	if err := agent.Start(); err != nil {
		return nil, errors.Wrap(err, "backup agent error")
	}
	defer agent.Stop()

	// Start backup
	dest := ctx.TempPath()
	os.MkdirAll(dest, 0770)
	args = []string{"start", "-w"}
	if s.Cluster != "" {
		args = append(args, "-C", s.Cluster)
	}
	if s.Tag != "" {
		args = append(args, "-t", s.Tag)
	}
	args = append(args, "-d", dest)
	cmd := exec.Command(fdbBackup, args...)
	var out bytes.Buffer
	var stderr bytes.Buffer
	cmd.Stdout = &out
	cmd.Stderr = &stderr
	if err := cmd.Run(); err != nil {
		fmt.Println(fmt.Sprint(err) + ": " + stderr.String())
		return nil, errors.Wrap(s.parseError(err), "fdbbackup error")
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

	// Create tarball
	tarball := path.Join(dest, backupRootDir+".tar")
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
	backupAgent := s.BackupAgent
	if backupAgent == "" {
		if _, err := exec.LookPath("backup_agent"); err != nil {
			return errors.Wrap(err, "backup_agent not found. Install it or check your $PATH")
		}
		backupAgent = execBackupAgent
	}
	fdbRestore := s.FDBRestore
	if fdbRestore == "" {
		if _, err := exec.LookPath("fdbrestore"); err != nil {
			return errors.Wrap(err, "fdbrestore not found. Install it or check your $PATH")
		}
		fdbRestore = execFdbrestore
	}
	if _, err := exec.LookPath("tar"); err != nil {
		return errors.Wrap(err, "tar not found. Install it or check your $PATH")
	}

	// Start backup agent
	args := []string{}
	if s.Cluster != "" {
		args = append(args, "-C", s.Cluster)
	}
	agent := agent{
		name: backupAgent,
		args: args,
	}
	if err := agent.Start(); err != nil {
		return errors.Wrap(err, "backup agent error")
	}
	defer agent.Stop()

	// Untar file
	dest := ctx.TempPath()
	os.MkdirAll(dest, 0770)
	f, err := ctx.CreateTempFile(r)
	if err != nil {
		return err
	}
	backupPath, err := filepath.Abs(f.Name())
	if err != nil {
		return errors.Wrap(err, "cannot get backup absolute path")
	}
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

	// Start restoring
	args = []string{"start", "-w"}
	if s.Cluster != "" {
		args = append(args, "-C", s.Cluster)
	}
	if s.Tag != "" {
		args = append(args, "-t", s.Tag)
	}
	args = append(args, "-r", dest)
	cmd = exec.Command(fdbRestore, args...)
	cmd.Stdout = &out
	cmd.Stderr = &stderr
	if err := cmd.Run(); err != nil {
		fmt.Printf("%s - %s\n", fdbRestore, args)
		fmt.Println(fmt.Sprint(err) + ": " + stderr.String())
		return errors.Wrap(s.parseError(err), "fdbrestore error")
	}

	return nil
}

func (s *Source) parseError(err error) error {
	return errors.Wrap(
		err,
		fmt.Sprintf("operation error"),
	)
}

type agent struct {
	name string
	args []string

	cmd *exec.Cmd
}

func (a *agent) Start() error {
	a.cmd = exec.Command(a.name, a.args...)
	var out bytes.Buffer
	var stderr bytes.Buffer
	a.cmd.Stdout = &out
	a.cmd.Stderr = &stderr
	if err := a.cmd.Start(); err != nil {
		fmt.Println(fmt.Sprint(err) + ": " + stderr.String())
		return errors.Wrap(err, "backup_agent error")
	}
	time.Sleep(2 * time.Second)
	return nil
}

func (a *agent) Stop() error {
	if a.cmd == nil {
		return nil
	}
	return a.cmd.Process.Signal(os.Interrupt)
}
