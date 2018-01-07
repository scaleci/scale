package run

import (
	"fmt"
	"sync"

	"github.com/scaleci/scale/exec"
)

type Stage struct {
	ID          string
	Command     string
	DependsOn   []string `toml:"depends_on"`
	Parallelism int64

	ParentApp *App
}

// Call each invocation in a go routine
func (s *Stage) Run(currentIndex int64, totalContainers int64) error {
	var wg sync.WaitGroup
	wg.Add(int(s.Parallelism))

	for index := int64(0); index < s.Parallelism; index++ {
		go func(i int64, ci int64, t int64) {
			s.RunIndividual(i, ci, t)
			wg.Done()
		}(index, currentIndex, totalContainers)
	}

	wg.Wait()

	return nil
}

func (s *Stage) RunIndividual(parallelismIndex int64, currentIndex int64, totalContainers int64) {
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
		parallelIndex := int64(-1)
		if s.Parallelism > int64(1) {
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

	fmt.Printf("%s %+v\n", cmdName, cmdArgs)

	// TODO: We need to track these errors and so far
	exec.Run(cmdName, cmdArgs, fmt.Sprintf("docker.run.%s", s.ID))
}
