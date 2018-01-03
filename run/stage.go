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

func (s *Stage) Run() error {
	cmdName := "docker"
	cmdArgs := []string{
		"run",
		"--mount",
		fmt.Sprintf("type=bind,source=%s,target=/bin/scale", scaleBinaryPath),
	}

	for k, v := range s.ParentApp.Env() {
		cmdArgs = append(cmdArgs, "-e", fmt.Sprintf("%s=%s", k, v))
	}

	cmdArgs = append(cmdArgs, s.ParentApp.ImageName())
	cmdArgs = append(cmdArgs, "/bin/bash", "-c")
	cmdArgs = append(cmdArgs, s.Command)

	return exec.Run(cmdName, cmdArgs, fmt.Sprintf("docker.run.%s", s.ID))
}
