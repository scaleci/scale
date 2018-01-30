package run

import (
	"bytes"
	"fmt"
	"sync"

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
