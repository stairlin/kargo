package log

import (
	"fmt"
	"io"
	"os"
	"sync"

	log "github.com/sirupsen/logrus"
	pb "gopkg.in/cheggaaa/pb.v1"
)

func init() {
	// Output to stdout instead of the default stderr
	log.SetOutput(os.Stdout)
}

type Logger interface {
	// Info outputs an info log message
	Info(s string, fields ...Field)
	// Warn outputs a warning log message
	Warn(s string, fields ...Field)
	// Error outputs an error log message
	Error(s string, fields ...Field)
	// Progress tracks progresses of r
	Progress(s string, r io.Reader, size ...int64) io.ReadCloser
}

// New creates a new standard logger
func New(fields ...Field) Logger {
	l := log.New()
	l.SetLevel(log.InfoLevel)
	return &ttyLogger{logger: l.WithFields(wrapFields(fields...))}
}

type ttyLogger struct {
	mu     sync.Mutex
	logger *log.Entry
	bar    *bar
}

// Info outputs an info log message
func (l *ttyLogger) Info(s string, fields ...Field) {
	l.logFunc(fields...).Info(s)
}

// Warn outputs a warning log message
func (l *ttyLogger) Warn(s string, fields ...Field) {
	l.logFunc(fields...).Warn(s)
}

// Error outputs an error log message
func (l *ttyLogger) Error(s string, fields ...Field) {
	l.logFunc(fields...).Error(s)
}

// Error outputs an error log message
func (l *ttyLogger) logFunc(fields ...Field) *log.Entry {
	l.mu.Lock()
	defer l.mu.Unlock()
	if l.bar != nil {
		l.bar.Close()
		l.bar = nil
	}
	return l.logger.WithFields(wrapFields(fields...))
}

// Progress outputs and update a progress bar until it is being closed
// or another inbound log line interupts it
func (l *ttyLogger) Progress(
	s string, r io.Reader, size ...int64,
) io.ReadCloser {
	l.mu.Lock()
	defer l.mu.Unlock()
	if l.bar != nil {
		l.bar.Close()
	}
	var total int64
	if len(size) > 0 {
		total = size[0]
	}
	b := pb.New64(total).SetUnits(pb.U_BYTES)
	b.ShowTimeLeft = true
	b.ShowPercent = false
	b.ShowSpeed = true
	b.SetMaxWidth(100)
	b.Prefix(fmt.Sprintf("\x1b[36mINFO\x1b[0m       %s", s))
	b.Start()
	l.bar = &bar{b, b.NewProxyReader(r)}
	return l.bar
}

func wrapFields(fields ...Field) log.Fields {
	lf := log.Fields{}
	for _, field := range fields {
		k, v := field.KV()
		lf[k] = v
	}
	return lf
}

type bar struct {
	*pb.ProgressBar
	proxy io.Reader
}

func (b *bar) Close() error {
	b.Finish()
	return nil
}

func (b *bar) Read(p []byte) (n int, err error) {
	return b.proxy.Read(p)
}
