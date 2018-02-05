package run

import (
	"bytes"
	"fmt"
	"strings"
	"sync"
	"time"

	"github.com/scaleci/scale/exec"
)

type Stage struct {
	ID          string
	Command     string
	DependsOn   []string `toml:"depends_on"`
	Parallelism int

	// Array by parallelism
	StdOut     []bytes.Buffer
	StdErr     []bytes.Buffer
	StatusCode []int

	ParentApp *App
}

func RunStages(app *App, scaleBinaryPath string) {
	var total int = 0
	var index int = 0
	var lastCompletedStageGroupIndex int = 0

	for _, s := range app.Stages {
		total += s.Parallelism
	}

	ticker := time.NewTicker(time.Second)
	go func() {
		for _ = range ticker.C {
			fmt.Printf(".")
		}
	}()

	for stageGroupIndex, stageGroups := range app.Graph {
		var wg sync.WaitGroup
		wg.Add(len(stageGroups))

		if len(stageGroups) == 1 {
			fmt.Printf("\n==== Running stage: %s ====\n", stageGroups[0].ID)
		} else {
			fmt.Printf("\n==== Running stages in parallel: %s ====\n", stageNameCollection(stageGroups))
		}

		for i := range stageGroups {
			stage := stageGroups[i]

			go func(i int) {
				stage.Run(i, total, scaleBinaryPath)
				wg.Done()
			}(index)

			index += stage.Parallelism
		}

		wg.Wait()

		status := 0
		// If any one of the stages returned with
		// a non-zero return code, status will be > 0
		for _, s := range stageGroups {
			for _, st := range s.StatusCode {
				status += st
			}
		}

		if status > 0 {
			lastCompletedStageGroupIndex = stageGroupIndex
			break
		}
	}

	ticker.Stop()
	for i, stageGroups := range app.Graph {
		if i <= lastCompletedStageGroupIndex {
			for _, stage := range stageGroups {
				for i := 0; i < stage.Parallelism; i++ {
					fmt.Printf("\n==== Completed %s.%d with status code %d =====\n", stage.ID, i, stage.StatusCode[i])
					if stdout := stage.StdOut[i].String(); stdout != "" {
						fmt.Printf("%s\n", stdout)
					}
					if stderr := stage.StdErr[i].String(); stderr != "" {
						fmt.Printf("%s\n", stderr)
					}
				}
			}
		}
	}
}

func stageNameCollection(stages []*Stage) string {
	stageNames := []string{}
	for _, stage := range stages {
		stageNames = append(stageNames, stage.ID)
	}

	return strings.Join(stageNames, " and ")
}

// Call each invocation in a go routine
func (s *Stage) Run(currentIndex int, totalContainers int, scaleBinaryPath string) {
	var wg sync.WaitGroup
	wg.Add(s.Parallelism)

	s.StdOut = make([]bytes.Buffer, int(s.Parallelism))
	s.StdErr = make([]bytes.Buffer, int(s.Parallelism))
	s.StatusCode = make([]int, int(s.Parallelism))

	for index := int(0); index < s.Parallelism; index++ {
		go func(i int, ci int, t int) {
			s.RunIndividual(i, ci, t, scaleBinaryPath)
			wg.Done()
		}(index, currentIndex, totalContainers)
	}

	wg.Wait()
}

func (s *Stage) RunIndividual(parallelismIndex int, currentIndex int, totalContainers int, scaleBinaryPath string) {
	cmdName := "docker"
	cmdArgs := []string{
		"run",
		"--mount",
		fmt.Sprintf("type=bind,source=%s,target=/bin/scale", scaleBinaryPath),
	}

	env := make(map[string]string)
	for configKey, configVal := range s.ParentApp.GlobalConfig.Env {
		env[configKey] = configVal
	}
	// Inject "running state" into the container
	// MAX is the total number of containers (across all stages) that are part of this run
	env["SCALE_CI_MAX"] = fmt.Sprintf("%d", totalContainers)
	// TOTAL is the parallelism of the current stage
	env["SCALE_CI_TOTAL"] = fmt.Sprintf("%d", s.Parallelism)
	// INDEX is the index of the current running container
	env["SCALE_CI_INDEX"] = fmt.Sprintf("%d", parallelismIndex)

	for _, service := range s.ParentApp.Services {
		parallelIndex := int(-1)
		if s.Parallelism > int(1) {
			parallelIndex = currentIndex + parallelismIndex
		}

		for serviceKey, serviceVal := range service.Env(parallelIndex) {
			env[serviceKey] = serviceVal
		}
	}

	for k, v := range env {
		cmdArgs = append(cmdArgs, "-e", fmt.Sprintf("%s=%s", k, v))
	}

	cmdArgs = append(cmdArgs, s.ParentApp.ImageName())
	cmdArgs = append(cmdArgs, "/bin/bash", "-c")
	cmdArgs = append(cmdArgs, s.Command)

	// TODO: We need to track these errors and so far
	s.StatusCode[parallelismIndex] = exec.RunAndCaptureOutput(cmdName,
		cmdArgs,
		&s.StdOut[parallelismIndex],
		&s.StdErr[parallelismIndex])
}
