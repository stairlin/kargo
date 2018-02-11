package agent

import (
	"fmt"
	"os"
	"strings"

	toml "github.com/pelletier/go-toml"
	"github.com/pelletier/go-toml/query"
	"github.com/pkg/errors"
	"github.com/stairlin/kargo/context"
	"github.com/stairlin/kargo/log"
	"github.com/stairlin/kargo/plugin/notification"
	"github.com/stairlin/kargo/plugin/process"
	"github.com/stairlin/kargo/plugin/source"
	"github.com/stairlin/kargo/plugin/storage"

	// Load all plugins
	_ "github.com/stairlin/kargo/plugin/notification/all"
	_ "github.com/stairlin/kargo/plugin/process/all"
	_ "github.com/stairlin/kargo/plugin/source/all"
	_ "github.com/stairlin/kargo/plugin/storage/all"
)

type Agent struct {
	Debug  bool `toml:"debug"`
	Silent bool

	Source     source.Source
	Processors []process.Processor
	Storage    storage.Storage
	Notifiers  []notification.Notifier
}

// Build returns a new Agent with all plugins loaded
func Build(ctx *context.Context, configPath string) (*Agent, error) {
	ctx.Info("Init...", log.String("config", configPath))
	tree, err := toml.LoadFile(configPath)
	if err != nil {
		return nil, errors.Wrap(
			err, fmt.Sprintf("cannot open config file: %s", configPath),
		)
	}

	// Replace environment variables with their value
	q, _ := query.Compile("$..")
	results := q.Execute(tree)
	for _, item := range results.Values() {
		switch v := item.(type) {
		case *toml.Tree:
			for _, key := range v.Keys() {
				v.Set(key, valueOf(v.Get(key)))
			}
		case []*toml.Tree:
			for _, tree := range v {
				for _, key := range tree.Keys() {
					tree.Set(key, valueOf(tree.Get(key)))
				}
			}
		}
	}

	a := &Agent{}
	if conf, ok := tree.Get("agent").(*toml.Tree); ok {
		if err := conf.Unmarshal(a); err != nil {
			return nil, errors.Wrap(err, "cannot unmarshal agent config")
		}
	}

	// Source
	if conf, ok := tree.Get("source").(*toml.Tree); ok {
		for _, k := range conf.Keys() {
			sourceCreator, ok := source.Sources[k]
			if !ok {
				return nil, fmt.Errorf("source <%s> does not exist", k)
			}
			source := sourceCreator()
			if err := conf.Get(k).(*toml.Tree).Unmarshal(source); err != nil {
				return nil, fmt.Errorf("cannot unmarshal <%s> config", k)
			}
			if err := source.Init(); err != nil {
				return nil, fmt.Errorf("cannot init <%s> %s", k, err)
			}
			a.Source = source
		}
	}

	// Storage
	if conf, ok := tree.Get("storage").(*toml.Tree); ok {
		for _, k := range conf.Keys() {
			storageCreator, ok := storage.Storages[k]
			if !ok {
				return nil, fmt.Errorf("storage <%s> does not exist", k)
			}
			storage := storageCreator()
			if err := conf.Get(k).(*toml.Tree).Unmarshal(storage); err != nil {
				return nil, fmt.Errorf("cannot unmarshal <%s> config", k)
			}
			if err := storage.Init(); err != nil {
				return nil, fmt.Errorf("cannot init <%s> %s", k, err)
			}
			a.Storage = storage
		}
	}

	// Processors
	if conf, ok := tree.Get("processors").(*toml.Tree); ok {
		for _, k := range conf.Keys() {
			procCreator, ok := process.Processors[k]
			if !ok {
				return nil, fmt.Errorf("processor <%s> does not exist", k)
			}
			proc := procCreator()

			if conf, ok := conf.Get(k).(*toml.Tree); ok {
				if err := conf.Unmarshal(proc); err != nil {
					return nil, fmt.Errorf("cannot unmarshal <%s> %s", k, err)
				}
			}

			if err := proc.Init(); err != nil {
				return nil, fmt.Errorf("cannot init <%s> %s", k, err)
			}
			a.Processors = append(a.Processors, proc)
		}
	}

	// Notifiers
	if conf, ok := tree.Get("notifiers").(*toml.Tree); ok {
		for _, k := range conf.Keys() {
			trees, ok := conf.Get(k).([]*toml.Tree)
			if !ok {
				continue
			}

			for _, t := range trees {
				notifCreator, ok := notification.Notifiers[k]
				if !ok {
					return nil, fmt.Errorf("processor <%s> does not exist", k)
				}
				notifier := notifCreator()

				if err := t.Unmarshal(notifier); err != nil {
					return nil, fmt.Errorf("cannot unmarshal <%s> config", k)
				}

				if err := notifier.Init(); err != nil {
					return nil, fmt.Errorf("cannot init <%s>", k)
				}
				a.Notifiers = append(a.Notifiers, notifier)
			}
		}
	}
	return a, nil
}

// Notify sends the notification n to all notifiers
func (a *Agent) Notify(ctx *context.Context, n *notification.Notification) error {
	if a.Silent {
		return nil
	}

	for _, notifier := range a.Notifiers {
		if err := notifier.Send(ctx, *n); err != nil {
			ctx.Error(
				fmt.Sprintf("%s failed to send notification", notifier.Name()),
				log.Error(err),
			)
			return err
		}
	}
	return nil
}

const prefix = "$"

// valueOf extracts the environment variable(s) from v
//
// e.g. foo -> foo
//      $DATABASE_URL -> http://foo.bar:8083
func valueOf(v interface{}) interface{} {
	switch v := v.(type) {
	case string:
		if strings.HasPrefix(v, prefix) && len(v) > 1 {
			return os.Getenv(v[1:])
		}
	case []interface{}:
		r := make([]interface{}, len(v))
		for i := range v {
			r[i] = valueOf(v[i])
		}
		return r
	}
	return v
}
