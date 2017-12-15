// Copyright © 2017 NAME HERE <EMAIL ADDRESS>
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

	"github.com/scaleci/scale/tests"
	"github.com/spf13/cobra"
)

var globCmd = &cobra.Command{
	Use:   "glob",
	Short: "Glob files in a given path",
	Long:  "Glob files in a given path",
	Args:  cobra.ExactArgs(1),
	Run:   runGlob,
}

// testsCmd represents the tests command
var testsCmd = &cobra.Command{
	Use:   "tests",
	Short: "Utilities that help with tests",
	Long:  "Utilities that help with tests",
}

func runGlob(cmd *cobra.Command, args []string) {
	matches := tests.Glob(args[0])

	for i := 0; i < len(matches); i++ {
		fmt.Println(matches[i])
	}
}

func init() {
	RootCmd.AddCommand(testsCmd)
	testsCmd.AddCommand(globCmd)
}
