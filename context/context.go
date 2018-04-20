package context

import (
	"bufio"
	"context"
	"io"
	"io/ioutil"
	"os"
	"path/filepath"
	"strings"
	"sync"
	"time"

	"github.com/google/uuid"
	"github.com/pkg/errors"
	"github.com/stairlin/kargo/log"
)

const (
	// defaultChmod defines the default permissions applied to folders/files
	// created by a context
	defaultChmod = 0640
)

// Context is a standard context enhanced for Kargo. It contains useful
// functions and additional context.
type Context struct {
	mu      sync.RWMutex
	closers []io.Closer
	files   []*os.File
	dirs    []string
	logger  log.Logger

	// UUID is the unique context ID
	UUID string
	// Context is the standard go context
	Context context.Context
	// Workdir contains the path where files must be created
	Workdir string
	// StartTime is the moment in time where this context was created
	StartTime time.Time
}

// Background returns a non-nil, empty Context. It is never canceled,
// has no values, and has no deadline.
// It is typically used by the main function, initialization, and tests,
// and as the top-level Context for incoming requests.
func Background() *Context {
	return newContext(context.Background())
}

// WithDeadline returns a copy of the parent context with the deadline adjusted
// to be no later than d. If the parent's deadline is already earlier than d,
// WithDeadline(parent, d) is semantically equivalent to parent.
// The returned context's Done channel is closed when the deadline expires,
// when the returned cancel function is called, or when the parent context's
// Done channel is closed, whichever happens first.
func WithDeadline(parent *Context, deadline time.Time) *Context {
	c, _ := context.WithDeadline(parent.Context, deadline)
	ctx := newContext(c)
	ctx.Workdir = parent.Workdir
	return ctx
}

func newContext(c context.Context) *Context {
	ctx := &Context{
		StartTime: time.Now(),
		UUID:      uuid.New().String(),
		Context:   c,
	}
	ctx.logger = log.New(log.String("id", ctx.ID()))
	return ctx
}

func (c *Context) ID() string {
	return strings.Split(c.UUID, "-")[0]
}

// CreateTempFile creates a file and copy data from r. This file will be
// deleted at the end of this context
func (c *Context) CreateTempFile(r io.Reader) (*os.File, error) {
	f, err := ioutil.TempFile(c.Workdir, "tmp")
	if err != nil {
		return nil, errors.Wrap(err, "cannot create temporary file")
	}
	if err := f.Chmod(defaultChmod); err != nil {
		return nil, errors.Wrap(err, "cannot chmod temporary file")
	}
	c.AddFile(f)

	// Buffer writes to disk
	buf := bufio.NewWriter(f)

	_, err = io.Copy(buf, r)
	switch err {
	case nil, io.ErrUnexpectedEOF:
	default:
		return nil, errors.Wrap(err, "cannot copy temp file")
	}
	if err := buf.Flush(); err != nil {
		return nil, errors.Wrap(err, "cannot flush buffer to temp file")
	}
	if _, err := f.Seek(0, 0); err != nil {
		return nil, errors.Wrap(err, "cannot rewind temp file")
	}
	return f, nil
}

// Persist saves data from r in a file on the working directory.
func (c *Context) Persist(name string, r io.Reader) error {
	f, err := os.Create(filepath.Join(c.Workdir, name))
	if err != nil {
		return errors.Wrap(err, "cannot create file")
	}
	if err := f.Chmod(defaultChmod); err != nil {
		return errors.Wrap(err, "cannot chmod file")
	}

	// Buffer writes to disk
	buf := bufio.NewWriter(f)
	defer buf.Flush()

	_, err = io.Copy(buf, r)
	switch err {
	case nil, io.ErrUnexpectedEOF:
	default:
		return errors.Wrap(err, "cannot copy data to file")
	}
	if err := f.Sync(); err != nil {
		return errors.Wrap(err, "cannot sync data to the disk")
	}
	return nil
}

// Load loads data from local file
func (c *Context) Load(name string) (io.ReadCloser, os.FileInfo, error) {
	f, err := os.Open(filepath.Join(c.Workdir, name))
	if err != nil {
		return nil, nil, errors.Wrap(err, "cannot open file")
	}

	info, err := f.Stat()
	if err != nil {
		return nil, nil, errors.Wrap(err, "cannot load file stat")
	}
	return f, info, nil
}

// TempPath returns a random file name for a temporary file
// TODO: Remove
func (c *Context) TempPath() string {
	if err := os.MkdirAll(c.Workdir, os.ModeDir); err != nil {
		panic(err)
	}
	dir, err := ioutil.TempDir(c.Workdir, "kargo-tmp-"+c.UUID)
	if err != nil {
		panic(err)
	}

	c.mu.Lock()
	c.dirs = append(c.dirs, dir)
	c.mu.Unlock()

	return filepath.Join(dir, uuid.New().String())
}

// Cleanup closes all closers and removes all temporary files. This should be
// called at the end of the context.
func (c *Context) Cleanup() {
	c.mu.Lock()
	defer c.mu.Unlock()

	for _, closer := range c.closers {
		closer.Close()
	}
	c.closers = []io.Closer{}
	for _, file := range c.files {
		file.Close()
		os.Remove(file.Name())
	}
	c.files = []*os.File{}
	for _, dir := range c.dirs {
		os.RemoveAll(dir)
	}
	c.dirs = []string{}
}

// AddCloser registers a resource to be closed at the end of this context
func (c *Context) AddCloser(fn io.Closer) {
	c.mu.Lock()
	c.closers = append(c.closers, fn)
	c.mu.Unlock()
}

// AddFile registers a resource to be closed/removed at the end of this context
func (c *Context) AddFile(f *os.File) {
	c.mu.Lock()
	c.files = append(c.files, f)
	c.mu.Unlock()
}

// Info outputs an info log message
func (c *Context) Info(s string, fields ...log.Field) {
	c.logger.Info(s, fields...)
}

// Warn outputs a warning log message
func (c *Context) Warn(s string, fields ...log.Field) {
	c.logger.Warn(s, fields...)
}

// Error outputs an error log message
func (c *Context) Error(s string, fields ...log.Field) {
	c.logger.Error(s, fields...)
}

// Progress outputs and update a progress bar
func (c *Context) Progress(
	s string, r io.Reader, size int64,
) io.ReadCloser {
	proxy := c.logger.Progress(s, r, size)
	c.AddCloser(proxy)
	return proxy
}

// Deadline returns the time when work done on behalf of this context
// should be canceled. Deadline returns ok==false when no deadline is
// set. Successive calls to Deadline return the same results.
func (c *Context) Deadline() (deadline time.Time, ok bool) {
	return c.Context.Deadline()
}

// Done returns a channel that's closed when work done on behalf of this
// context should be canceled. Done may return nil if this context can
// never be canceled. Successive calls to Done return the same value.
func (c *Context) Done() <-chan struct{} {
	return c.Context.Done()
}

// If Done is not yet closed, Err returns nil.
// If Done is closed, Err returns a non-nil error explaining why:
// Canceled if the context was canceled
// or DeadlineExceeded if the context's deadline passed.
// After Err returns a non-nil error, successive calls to Err return the same error.
func (c *Context) Err() error {
	return c.Context.Err()
}

// Value returns the value associated with this context for key, or nil
// if no value is associated with key. Successive calls to Value with
// the same key returns the same result.
func (c *Context) Value(key interface{}) interface{} {
	return c.Context.Value(key)
}
