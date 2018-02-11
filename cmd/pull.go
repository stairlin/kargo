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
	"fmt"

	"github.com/spf13/cobra"
	"github.com/stairlin/kargo/agent"
	"github.com/stairlin/kargo/context"
	"github.com/stairlin/kargo/log"
)

var (
	processBackup bool
)

// pullCmd represents the pull command
var pullCmd = &cobra.Command{
	Use:   "pull",
	Short: "Pull a backup from storage, process it, and then persist it locally",
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

		// Pull data from store
		ctx.Info("Pulling file from storage...")
		data, info, err := agent.Storage.Pull(ctx, key)
		if err != nil {
			ctx.Error("Failed to pull file", log.Error(err))
			return
		}
		data = ctx.Progress("Pulling file", data, info.Size())
		ctx.AddCloser(data)

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

		// Persist it locally
		if err := ctx.Persist(key, data); err != nil {
			ctx.Error("Failed to persist data", log.Error(err))
			return
		}
	},
}

func init() {
	rootCmd.AddCommand(pullCmd)

	// Here you will define your flags and configuration settings.

	// Cobra supports Persistent Flags which will work for this command
	// and all subcommands, e.g.:
	// pullCmd.PersistentFlags().String("foo", "", "A help for foo")

	// Cobra supports local flags which will only run when this command
	// is called directly, e.g.:
	pullCmd.Flags().StringP("key", "k", "", "Backup key")
	pullCmd.Flags().BoolVarP(&processBackup, "process", "p", true, "Process backup")
}
