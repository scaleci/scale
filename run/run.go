package run

import (
	"github.com/scaleci/scale/exec"
)

func Build(app *App) error {
	cmdName := "docker"
	cmdArgs := []string{
		"build",
		".",
		"-f", app.GlobalConfig.BuildWith,
		"-t", app.ImageName()}

	return exec.Run(cmdName, cmdArgs, "docker.build")
}

func StartServices(app *App) error {
	for _, s := range app.Services {
		err := s.Start()
		if err != nil {
			return err
		}
	}

	return nil
}

func StopServices(app *App) error {
	for _, s := range app.Services {
		err := s.Stop()
		if err != nil {
			return err
		}
	}

	return nil
}

func RunStages(app *App) error {
	var total int64 = 0

	for _, s := range app.Stages {
		total += s.Parallelism
	}

	for _, s := range app.Stages {
		if len(s.DependsOn) == 0 {
			err := s.Run(total)
			if err != nil {
				return err
			}
		}
	}

	return nil
}
