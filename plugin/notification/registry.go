package notification

import (
	"time"

	"github.com/stairlin/kargo/context"
)

type Notifier interface {
	Name() string
	Init() error
	Send(*context.Context, Notification) error
}

type Notification struct {
	Type      Type
	Operation Operation

	StartTime time.Time
	EndTime   time.Time
	Body      string
	Error     error
}

// Type represents a type of notification
type Type int

const (
	Success Type = iota
	Failure Type = iota
	Timeout Type = iota
)

// Operation is the operation that triggers the notification
type Operation int

const (
	Backup  Operation = iota
	Restore Operation = iota
)

type Creator func() Notifier

var Notifiers = map[string]Creator{}

func Add(name string, creator Creator) {
	Notifiers[name] = creator
}
