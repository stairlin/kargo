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

	"github.com/pkg/errors"
	"github.com/spf13/cobra"
	"github.com/stairlin/kargo/agent"
	"github.com/stairlin/kargo/context"
)

// infoCmd represents the info command
var infoCmd = &cobra.Command{
	Use:   "info",
	Short: "Show the current config",
	Long:  ``,
	Run: func(cmd *cobra.Command, args []string) {
		// Build context
		ctx := context.Background()
		ctx.Workdir = workdir
		defer ctx.Cleanup()

		agent, err := agent.Build(ctx, configPath)
		if err != nil {
			err = errors.Wrapf(err, "agent backup error")
			fmt.Println(err)
			return
		}

		fmt.Println("WORKFLOW:")
		fmt.Printf("\t - %s\n", agent.Source.Name())
		for _, p := range agent.Processors {
			fmt.Printf("\t - %s\n", p.Name())
		}
		fmt.Printf("\t - %s\n", agent.Storage.Name())

		fmt.Println("\nNOTIFIERS:")
		for _, n := range agent.Notifiers {
			fmt.Printf("\t - %s\n", n.Name())
		}
	},
}

func init() {
	rootCmd.AddCommand(infoCmd)

	// Here you will define your flags and configuration settings.

	// Cobra supports Persistent Flags which will work for this command
	// and all subcommands, e.g.:
	// infoCmd.PersistentFlags().String("foo", "", "A help for foo")

	// Cobra supports local flags which will only run when this command
	// is called directly, e.g.:
	// infoCmd.Flags().BoolP("toggle", "t", false, "Help message for toggle")
}
