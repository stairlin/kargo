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
	"bytes"
	"fmt"
	"math"
	"os"
	"regexp"
	"text/tabwriter"
	"time"

	"github.com/spf13/cobra"
	"github.com/stairlin/kargo/agent"
	"github.com/stairlin/kargo/context"
	"github.com/stairlin/kargo/log"
	"github.com/stairlin/kargo/pkg/ago"
	"github.com/stairlin/kargo/pkg/bytefmt"
	"github.com/stairlin/kargo/plugin/storage"
)

const day = time.Hour * 24

var (
	// limit defines a limit of elements to display
	limit uint
)

// listCmd represents the list command
var listCmd = &cobra.Command{
	Use:   "list",
	Short: "List/Filter all available backups",
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

		// Build filters
		expr := cmd.Flag("pattern").Value.String()
		var pattern *regexp.Regexp
		if len(expr) > 0 {
			pattern, err = regexp.Compile(expr)
			if err != nil {
				fmt.Println(err)
				return
			}
		}
		prefix := cmd.Flag("prefix").Value.String()
		fromS := cmd.Flag("from").Value.String()
		var from int64
		if len(fromS) > 0 {
			t, err := time.Parse("2006-01-02", fromS)
			if err != nil {
				fmt.Println(err)
				return
			}
			from = beginningOfDay(t.UnixNano())
		}
		toS := cmd.Flag("to").Value.String()
		to := int64(math.MaxInt64)
		if len(toS) > 0 {
			t, err := time.Parse("2006-01-02", toS)
			if err != nil {
				fmt.Println(err)
				return
			}
			to = endOfDay(t.UnixNano())
		}
		filter := storage.WalkFilter{
			From:    from,
			To:      to,
			Prefix:  prefix,
			Pattern: pattern,
			Limit:   limit,
		}

		// Build output
		w := new(tabwriter.Writer)
		buf := bytes.NewBuffer([]byte{})
		w.Init(buf, 0, 8, 0, '\t', 0)
		fmt.Fprintln(w, "KEY\t SIZE\t LAST MODIFICATION\t FROM NOW")

		var totalItems int
		var totalSize int64
		walkFn := func(key string, f os.FileInfo, err error) error {
			if err != nil {
				return err
			}
			totalItems++
			totalSize += f.Size()

			fmt.Fprintf(w, "%s\t %s\t %s\t %s\n",
				key,
				bytefmt.HumanReadableByte(f.Size()),
				f.ModTime().String(),
				ago.Ago(f.ModTime()),
			)
			return nil
		}
		agent.Storage.Walk(ctx, &filter, walkFn)

		w.Flush()
		fmt.Println()
		fmt.Println(buf.String())
		fmt.Printf(
			"ITEMS %d (%s)\n",
			totalItems,
			bytefmt.HumanReadableByte(totalSize),
		)
	},
}

func init() {
	rootCmd.AddCommand(listCmd)

	// Here you will define your flags and configuration settings.

	// Cobra supports Persistent Flags which will work for this command
	// and all subcommands, e.g.:
	// listCmd.PersistentFlags().String("foo", "", "A help for foo")

	// Cobra supports local flags which will only run when this command
	// is called directly, e.g.:
	listCmd.Flags().StringP("pattern", "", "", "Filter keys with a pattern")
	listCmd.Flags().StringP("prefix", "", "", "Filter keys with a prefix")
	listCmd.Flags().StringP("from", "", "", "Keep keys created after time t")
	listCmd.Flags().StringP("to", "", "", "Keep keys created before time t")
	listCmd.Flags().UintVarP(&limit, "limit", "l", 30, "Limit the number of keys displayed")
}

func beginningOfDay(t int64) int64 {
	return floor(t, day)
}

func endOfDay(t int64) int64 {
	return floor(t, day) + int64(day-time.Nanosecond)
}

func floor(t int64, prec time.Duration) int64 {
	return t - abs(t)%int64(prec)
}

func abs(t int64) int64 {
	if t < 0 {
		return t * -1
	}
	return t
}
