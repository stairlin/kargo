package slack

import (
	"bytes"
	"encoding/json"
	"fmt"
	"net/http"
	"time"

	"github.com/pkg/errors"
	"github.com/stairlin/kargo/context"
	"github.com/stairlin/kargo/plugin/notification"
)

const name = "slack"

func init() {
	notification.Add(name, func() notification.Notifier {
		return &Notifier{}
	})
}

// Notifier is a Slack notifier
type Notifier struct {
	URL      string `toml:"url"`
	Channel  string `toml:"channel"`
	Username string `toml:"username"`

	HTTP *http.Client
}

type payload struct {
	Channel   string `json:"channel,omitempty"`
	Username  string `json:"username,omitempty"`
	Text      string `json:"text"`
	IconEmoji string `json:"icon_emoji,omitempty"`
}

func (n *Notifier) Name() string {
	return name
}

func (n *Notifier) Init() error {
	if n.URL == "" {
		return errors.New("slack needs a URL")
	}
	n.HTTP = &http.Client{
		Timeout: time.Duration(5 * time.Second),
	}
	return nil
}

func (n *Notifier) Send(
	ctx *context.Context, notif notification.Notification,
) error {
	ctx.Info("Sending Slack notification...")

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

	var text string
	switch notif.Type {
	case notification.Success:
		text = fmt.Sprintf(
			"%s done in %s - %s", operation, duration.String(), notif.Body,
		)
	case notification.Failure:
		text = fmt.Sprintf(
			"ERROR: %s failed after %s.\n\nReason: %s",
			operation,
			duration.String(),
			notif.Error,
		)
	case notification.Timeout:
		text = fmt.Sprintf(
			"ERROR: %s timed out after %s.\n\nReason: %s",
			operation,
			duration.String(),
			notif.Error,
		)
	default:
		text = "unknown notification type"
	}

	payload, err := json.Marshal(&payload{
		Channel:   n.Channel,
		Username:  n.Username,
		Text:      text,
		IconEmoji: ":ship:",
	})
	if err != nil {
		return errors.Wrap(err, "cannot marshal notification to JSON")
	}

	req, err := http.NewRequest("POST", n.URL, bytes.NewBuffer(payload))
	if err != nil {
		return errors.Wrap(err, "cannot create HTTP request")
	}

	req.Header.Set("Content-Type", "application/json")
	req.Header.Set("User-Agent", "Kargo webhook")

	// Send request (with retry)
	var res *http.Response
	var attempts = 0
	for attempts = 0; attempts < 3; attempts++ {
		if res, err = n.HTTP.Do(req); err != nil {
			time.Sleep(2 * time.Second)
			continue
		}

		break
	}
	if err != nil {
		return errors.Wrapf(err, "cannot send http request to %s", n.URL)
	}

	// Log as a warning in case the endpoint returns a bad status code
	if res.StatusCode >= 300 {
		fmt.Printf("webhook bad status code %d\n", res.StatusCode)
	}

	return nil
}
