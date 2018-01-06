package run

import (
	"fmt"

	"github.com/scaleci/scale/exec"
)

type Stage struct {
	ID          string
	Command     string
	DependsOn   []string `toml:"depends_on"`
	Parallelism int64

	ParentApp *App
}

func (s *Stage) Run(totalContainers int64) error {
	cmdName := "docker"

	for index := 0; index < s.Parallelism; index++ {
		cmdArgs := []string{
			"run",
			"--mount",
			fmt.Sprintf("type=bind,source=%s,target=/bin/scale", scaleBinaryPath),
		}
		env := s.ParentApp.Env()
		// Inject "running state" into the container
		// MAX is the total number of containers (across all stages) that are part of this run
		env["SCALE_CI_MAX"] = fmt.Sprintf("%d", totalContainers)
		// TOTAL is the parallelism of the current stage
		env["SCALE_CI_TOTAL"] = fmt.Sprintf("%d", s.Parallelism)
		// INDEX is the index of the current running container
		env["SCALE_CI_INDEX"] = fmt.Sprintf("%d", index)

		for k, v := range env {
			cmdArgs = append(cmdArgs, "-e", fmt.Sprintf("%s=%s", k, v))
		}

		cmdArgs = append(cmdArgs, s.ParentApp.ImageName())
		cmdArgs = append(cmdArgs, "/bin/bash", "-c")
		cmdArgs = append(cmdArgs, s.Command)

		return exec.Run(cmdName, cmdArgs, fmt.Sprintf("docker.run.%s", s.ID))
	}
}
