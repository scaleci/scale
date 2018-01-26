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
	"strconv"
	"strings"

	"github.com/scaleci/scale/tests"
	"github.com/spf13/cobra"
)

var index int = -1
var total int = -1
var parallelizeOpts string

var globCmd = &cobra.Command{
	Use:   "glob",
	Short: "Glob files in a given path",
	Long:  "Glob files in a given path",
	Args:  cobra.ExactArgs(1),
	Run:   runGlob,
}

var splitCmd = &cobra.Command{
	Use:   "split",
	Short: "Split files provided to it according to its execution times",
	Long:  "Split files provided to it according to its execution times",
	Args:  cobra.MinimumNArgs(1),
	Run:   runSplit,
}

var parallelizeCmd = &cobra.Command{
	Use:   "parallelize",
	Short: "Prepares a datastore to be used in parallel during stage runs",
	Long:  "Prepares a datastore to be used in parallel during stage runs",
	Args:  cobra.MinimumNArgs(1),
	Run:   runParallelize,
}

// testsCmd represents the tests command
var testsCmd = &cobra.Command{
	Use:   "tests",
	Short: "Utilities that help with tests",
	Long:  "Utilities that help with tests",
}

func runParallelize(cmd *cobra.Command, args []string) {
	protocol := args[0]

	allowedProtocols := map[string]bool{
		"postgres": true,
	}

	if val, ok := allowedProtocols[protocol]; !ok || val == false {
		fmt.Printf("%s not supported\n", protocol)
		os.Exit(1)
	}

	total, err := strconv.Atoi(os.Getenv("SCALE_CI_MAX"))
	if err != nil || total < 1 {
		fmt.Printf("SCALE_CI_MAX must be set with a value greater than 0\n")
		os.Exit(1)
	}

	parallelizeOptsMap := map[string]string{}
	for _, opt := range strings.Split(parallelizeOpts, ",") {
		kvPair := strings.Split(opt, "=")
		if len(kvPair) == 2 {
			parallelizeOptsMap[kvPair[0]] = kvPair[1]
		}
	}

	if err = tests.Parallelize(protocol, total, parallelizeOptsMap); err != nil {
		fmt.Printf("Error parallelizing datastores: %+v\n", err)
		os.Exit(1)
	}
}

func runGlob(cmd *cobra.Command, args []string) {
	matches := tests.Glob(args[0])

	for i := 0; i < len(matches); i++ {
		fmt.Println(matches[i])
	}
}

func runSplit(cmd *cobra.Command, inputFiles []string) {
	envIndexStr := os.Getenv("SCALE_CI_INDEX")
	if envIndexStr != "" {
		envIndex, err := strconv.Atoi(envIndexStr)
		if err == nil {
			index = envIndex
		}
	}

	envTotalStr := os.Getenv("SCALE_CI_TOTAL")
	if envIndexStr != "" {
		envTotal, err := strconv.Atoi(envTotalStr)
		if err == nil {
			total = envTotal
		}
	}

	if index < 0 || total < 1 {
		fmt.Printf("tests split command needs index and total, set by flags -i and -t respectively\n")
		fmt.Printf("or by setting the env variables SCALE_CI_INDEX and SCALE_CI_TOTAL respectively\n")
		os.Exit(1)
	}

	splitFiles := tests.Split(inputFiles, int64(index), int64(total))
	for i := 0; i < len(splitFiles); i++ {
		fmt.Println(splitFiles[i])
	}
}

func init() {
	RootCmd.AddCommand(testsCmd)
	testsCmd.AddCommand(globCmd)
	testsCmd.AddCommand(splitCmd)
	testsCmd.AddCommand(parallelizeCmd)

	f := splitCmd.Flags()
	f.IntVarP(&index, "index", "i", -1, "index of the container within which tests are run")
	f.IntVarP(&total, "total", "t", -1, "total number of containers")

	f = parallelizeCmd.Flags()
	f.StringVarP(&parallelizeOpts, "options", "o", "", "options for parallelizing")
}
