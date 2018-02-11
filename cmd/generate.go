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
	"encoding/base64"
	"fmt"

	"github.com/spf13/cobra"
	"github.com/stairlin/kargo/pkg/sec"
)

// generateCmd represents the generate command
var generateCmd = &cobra.Command{
	Use:   "generate",
	Short: "Kargo generators",
	Long:  ``,
}

// generateCipherCmd represents the generate cipher command
var generateCipherCmd = &cobra.Command{
	Use:   "cipher",
	Short: "Generate a random cipher key",
	Long:  ``,
	Run: func(cmd *cobra.Command, args []string) {
		k, err := sec.GenerateKey()
		if err != nil {
			fmt.Println(err)
		}
		fmt.Println("Cipher key:", base64.StdEncoding.EncodeToString(k))
	},
}

func init() {
	rootCmd.AddCommand(generateCmd)

	generateCmd.AddCommand(generateCipherCmd)

	// Here you will define your flags and configuration settings.

	// Cobra supports Persistent Flags which will work for this command
	// and all subcommands, e.g.:
	// generateCmd.PersistentFlags().String("foo", "", "A help for foo")

	// Cobra supports local flags which will only run when this command
	// is called directly, e.g.:
	// generateCmd.Flags().BoolP("toggle", "t", false, "Help message for toggle")
}
