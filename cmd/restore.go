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
	"io"
	"os"
	"time"

	"github.com/spf13/cobra"
	"github.com/stairlin/kargo/agent"
	"github.com/stairlin/kargo/context"
	"github.com/stairlin/kargo/log"
	"github.com/stairlin/kargo/plugin/notification"
	"github.com/tcnksm/go-input"
)

var (
	// local returns whether a command should use local a backup instead of the
	// backup from storage
	local bool
)

// restoreCmd represents the restore command
var restoreCmd = &cobra.Command{
	Use:   "restore",
	Short: "Restore the source from a backup",
	Long:  ``,
	Run: func(cmd *cobra.Command, args []string) {
		if len(args) == 0 {
			fmt.Println("missing key")
			return
		}
		key := args[0]

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

		// Request confirmation
		force := cmd.Flag("force").Value.String() == "true"
		if force {
			ctx.Warn("Force restore (no confirmation requested)")
		} else {
			ui := &input.UI{
				Writer: os.Stdout,
				Reader: os.Stdin,
			}
			query := "Warning! Restoring the source may cause data loss. Are you sure? (Y/n)"
			res, err := ui.Ask(query, &input.Options{
				Default:  "n",
				Required: true,
				Loop:     true,
			})
			if err != nil {
				return
			}
			if res != "Y" {
				return
			}
		}

		var notified bool
		defer func() {
			if notified {
				return
			}
			ctx.Warn("Sending unknown failure notification...")
			n := notification.Notification{
				Type:      notification.Failure,
				Operation: notification.Restore,
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

		// Pull data from store or local file
		var data io.ReadCloser
		var info os.FileInfo
		if local {
			ctx.Info("Loading file from local disk...", log.String("key", key))
			data, info, err = ctx.Load(key)
		} else {
			ctx.Info("Pulling file from storage...", log.String("key", key))
			data, info, err = agent.Storage.Pull(ctx, key)

			// Run processors backward
			if processBackup {
				for i := len(agent.Processors) - 1; i >= 0; i-- {
					proc := agent.Processors[i]
					data, err = proc.Decode(ctx, data)
					if err != nil {
						ctx.Error("Failed to decode data", log.Error(err))
						return
					}
					ctx.AddCloser(data)
				}
			}

			data = ctx.Progress("Pulling file", data, info.Size())
		}
		if err != nil {
			ctx.Error("Failed to pull/load file", log.Error(err))
			return
		}

		// Start restore
		if err := agent.Source.Restore(ctx, data); err != nil {
			ctx.Error("Failed to restore data", log.Error(err))
			return
		}

		n := &notification.Notification{
			Type:      notification.Success,
			Operation: notification.Restore,
			StartTime: ctx.StartTime,
			EndTime:   time.Now(),
			Body:      fmt.Sprintf("Key %s", key),
		}
		if err := agent.Notify(ctx, n); err != nil {
			return
		}
		notified = true

		ctx.Info("OK", log.String("duration", time.Now().Sub(ctx.StartTime).String()))
	},
}

func init() {
	rootCmd.AddCommand(restoreCmd)

	// Here you will define your flags and configuration settings.

	// Cobra supports Persistent Flags which will work for this command
	// and all subcommands, e.g.:
	// restoreCmd.PersistentFlags().String("foo", "", "A help for foo")

	// Cobra supports local flags which will only run when this command
	// is called directly, e.g.:
	restoreCmd.Flags().BoolP("force", "", false, "Bypass confirmation")
	restoreCmd.Flags().BoolVarP(&local, "local", "l", false, "Restore from a local file instead of the storage")
}
