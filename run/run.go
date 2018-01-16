package run

import (
	"fmt"
	"sync"

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
	var index int64 = 0

	for _, s := range app.Stages {
		total += s.Parallelism
	}

	for _, stageGroups := range app.Graph {
		var wg sync.WaitGroup
		wg.Add(len(stageGroups))

		for i := range stageGroups {
			stage := stageGroups[i]

			go func(i int64) {
				if err := stage.Run(i, total); err != nil {
					fmt.Errorf("error running stage: %+v for stage %s\n", err, stage.ID)
				}
				wg.Done()
			}(index)

			index += stage.Parallelism
		}

		wg.Wait()
	}

	return nil
}
