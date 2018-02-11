package pagerduty

import (
	"fmt"
	"time"

	pagerduty "github.com/PagerDuty/go-pagerduty"
	"github.com/pkg/errors"
	"github.com/stairlin/kargo/context"
	"github.com/stairlin/kargo/plugin/notification"
)

const name = "pagerduty"

func init() {
	notification.Add(name, func() notification.Notifier {
		return &Notifier{}
	})
}

type Notifier struct {
	Key string `toml:"key"`
}

func (n *Notifier) Name() string {
	return name
}

func (n *Notifier) Init() error {
	if n.Key == "" {
		return errors.New("pagerduty needs an API key")
	}
	return nil
}

func (n *Notifier) Send(
	ctx *context.Context, notif notification.Notification,
) error {
	if notif.Type == notification.Success {
		ctx.Info("pagerduty does not notify successes. Skip")
		return nil
	}
	ctx.Info("Sending Pagerduty notification...")

	// Round to the second
	duration := notif.EndTime.Sub(notif.StartTime)
	duration -= duration % time.Second

	var operation string
	switch notif.Operation {
	case notification.Backup:
		operation = "Backup"
	case notification.Restore:
		operation = "Restore"
	default:
		operation = "?"
	}

	var desc string
	switch notif.Type {
	case notification.Failure:
		desc = fmt.Sprintf(
			"%s failed after %s.\n\nReason: %s",
			operation,
			duration.String(),
			notif.Error,
		)
	case notification.Timeout:
		desc = fmt.Sprintf(
			"%s timed out after %s.\n\nReason: %s",
			operation,
			duration.String(),
			notif.Error,
		)
	default:
		desc = "unsupported notification type"
	}

	event := pagerduty.Event{
		Type:        "trigger",
		ServiceKey:  n.Key,
		Description: desc,
	}
	_, err := pagerduty.CreateEvent(event)
	if err != nil {
		return errors.Wrap(err, "cannot create Pagerduty event")
	}
	return nil
}
