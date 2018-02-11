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

	"github.com/stairlin/kargo/plugin/notification"
	"github.com/stairlin/kargo/plugin/process"
	"github.com/stairlin/kargo/plugin/source"
	"github.com/stairlin/kargo/plugin/storage"

	"github.com/spf13/cobra"
)

// pluginsCmd represents the plugins command
var pluginsCmd = &cobra.Command{
	Use:   "plugins",
	Short: "Show all available plugins",
	Long:  ``,
	Run: func(cmd *cobra.Command, args []string) {
		fmt.Println("Sources:")
		for name := range source.Sources {
			fmt.Println("\t -", name)
		}

		fmt.Println("Processors:")
		for name := range process.Processors {
			fmt.Println("\t -", name)
		}

		fmt.Println("Storages:")
		for name := range storage.Storages {
			fmt.Println("\t -", name)
		}

		fmt.Println("Notifiers:")
		for name := range notification.Notifiers {
			fmt.Println("\t -", name)
		}
	},
}

func init() {
	rootCmd.AddCommand(pluginsCmd)

	// Here you will define your flags and configuration settings.

	// Cobra supports Persistent Flags which will work for this command
	// and all subcommands, e.g.:
	// pluginsCmd.PersistentFlags().String("foo", "", "A help for foo")

	// Cobra supports local flags which will only run when this command
	// is called directly, e.g.:
	// pluginsCmd.Flags().BoolP("toggle", "t", false, "Help message for toggle")
}
