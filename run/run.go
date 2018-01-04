package run

import (
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

	for _, stageGroups := range app.Graph {
		var wg sync.WaitGroup
		wg.Add(len(stageGroups))
		errors := []error{}

		for i := range stageGroups {
			stage := stageGroups[i]

			go func() {
				if err := stage.Run(total); err != nil {
					errors = append(errors, err)
				}
				wg.Done()
			}()
		}

		wg.Wait()

		// TODO: Need a more sane way to return these errors
		if len(errors) > 0 {
			return errors[0]
		}
	}

	return nil
}
