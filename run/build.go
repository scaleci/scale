package run

import (
	"fmt"

	"github.com/scaleci/scale/exec"
)

func Build(app *App) error {
	cmdName := "docker"
	cmdArgs := []string{
		"build",
		".",
		"-f", app.GlobalConfig.BuildWith,
		"-t", app.ImageName()}

	for k, v := range app.GlobalConfig.Env {
		cmdArgs = append(cmdArgs, "--build-arg", fmt.Sprintf("%s=%s", k, v))
	}

	return exec.Run(cmdName, cmdArgs, "docker.build")
}
