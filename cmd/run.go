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
	"os"

	"github.com/scaleci/scale/run"
	"github.com/spf13/cobra"
)

// runCmd represents the run command
var runCmd = &cobra.Command{
	Use:   "run",
	Short: "Run tests for app",
	Long:  "Run tests for app",
	Run:   runTests,
}

func runTests(cmd *cobra.Command, args []string) {
	configFilePath := "./scale.toml"

	if _, err := os.Stat(configFilePath); err != nil {
		fmt.Printf("scale.toml file does not exist in the current directory\n")
		os.Exit(1)
	}

	app, err := run.Parse(configFilePath)
	if err != nil {
		fmt.Printf("Error parsing scale.toml file: %+v\n", err)
		os.Exit(1)
	}

	fmt.Printf("Running %s...\n", app.Name)
	err = run.Build(app)
	if err != nil {
		fmt.Printf("Error building app %s: %+v\n", app.Name, err)
		os.Exit(1)
	}
}

func init() {
	RootCmd.AddCommand(runCmd)
}
