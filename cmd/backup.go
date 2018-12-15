// Copyright © 2018 Stairlin ltd <it@stairlin.com>
//
// Licensed under the Apache License, Version 2.0 (the "License");
// you may not use this file except in compliance with the License.
// You may obtain a copy of the License at
//
//     http://www.apache.org/licenses/LICENSE-2.0
//
// Unless required by applicable law or agreed to in writing, software
// distributed under the License is distributed on an "AS IS" BASIS,
// WITHOUT WARRANTIES OR CONDITIONS OF ANY KIND, either express or implied.
// See the License for the specific language governing permissions and
// limitations under the License.

package cmd

import (
	"errors"
	"fmt"
	"time"

	"github.com/spf13/cobra"
	"github.com/stairlin/kargo/agent"
	"github.com/stairlin/kargo/context"
	"github.com/stairlin/kargo/log"
	"github.com/stairlin/kargo/plugin/notification"
)

var (
	// key returns a backup custom key (if any)
	key string
)

// backupCmd represents the backup command
var backupCmd = &cobra.Command{
	Use:   "backup",
	Short: "Create a backup of the source, process it, and then store it",
	Long:  ``,
	Run: func(cmd *cobra.Command, args []string) {
		// Build context
		ctx := context.Background()
		ctx.Workdir = workdir
		defer ctx.Cleanup()

		// Build agent
		agent, err := agent.Build(ctx, configPath)
		if err != nil {
			ctx.Error("Failed to build agent", log.Error(err))
			return
		}
		agent.Silent = silent

		var notified bool
		defer func() {
			if notified {
				return
			}
			ctx.Warn("Sending unknown failure notification...")
			n := notification.Notification{
				Type:      notification.Failure,
				Operation: notification.Backup,
				StartTime: ctx.StartTime,
				EndTime:   time.Now(),
				Error:     errors.New("unknown failure"),
			}
			for _, notifier := range agent.Notifiers {
				if err := notifier.Send(ctx, n); err != nil {
					ctx.Error("Failed to send notification", log.Error(err))
				}
			}
		}()

		// Backup data
		ctx.Info("Backing up data...")
		data, err := agent.Source.Backup(ctx)
		if err != nil {
			ctx.Error("Failed to backup data", log.Error(err))
			return
		}
		ctx.AddCloser(data)

		// Run processors
		for _, proc := range agent.Processors {
			data, err = proc.Encode(ctx, data)
			if err != nil {
				ctx.Error("Failed to encode data", log.Error(err))
				return
			}
			ctx.AddCloser(data)
		}

		// Store backup
		data = ctx.Progress("Pushing file", data, 0)
		if key == "" {
			key = fmt.Sprintf("%s-%d", agent.Source.Name(), time.Now().Unix())
		}
		if err := agent.Storage.Push(ctx, key, data); err != nil {
			ctx.Error("Failed to push data", log.Error(err))
			return
		}
		data.Close()

		n := &notification.Notification{
			Type:      notification.Success,
			Operation: notification.Backup,
			StartTime: ctx.StartTime,
			EndTime:   time.Now(),
			Body:      fmt.Sprintf("Key %s", key),
		}
		if err := agent.Notify(ctx, n); err != nil {
			return
		}
		notified = true

		ctx.Info("OK",
			log.String("key", key),
			log.String("duration", time.Now().Sub(ctx.StartTime).String()),
		)
	},
}

func init() {
	rootCmd.AddCommand(backupCmd)

	// Here you will define your flags and configuration settings.

	// Cobra supports Persistent Flags which will work for this command
	// and all subcommands, e.g.:
	// backupCmd.PersistentFlags().String("foo", "", "A help for foo")

	// Cobra supports local flags which will only run when this command
	// is called directly, e.g.:
	backupCmd.Flags().StringVarP(&key, "key", "k", "", "Override default backup key")
}
