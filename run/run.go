package run

import (
	"fmt"
	"sync"
	"time"

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

func RunStages(app *App) {
	var total int = 0
	var index int = 0

	for _, s := range app.Stages {
		total += s.Parallelism
	}

	ticker := time.NewTicker(time.Second)
	go func() {
		for _ = range ticker.C {
			fmt.Printf(".")
		}
	}()

	for _, stageGroups := range app.Graph {
		var wg sync.WaitGroup
		wg.Add(len(stageGroups))

		for i := range stageGroups {
			stage := stageGroups[i]

			go func(i int) {
				stage.Run(i, total)
				wg.Done()
			}(index)

			index += stage.Parallelism
		}

		wg.Wait()
	}

	ticker.Stop()
	fmt.Printf("\n")

	for _, stageGroups := range app.Graph {
		for _, stage := range stageGroups {
			for i := 0; i < stage.Parallelism; i++ {
				fmt.Printf("\n==== Completed %s.%d with status code %d =====\n", stage.ID, i, stage.StatusCode[i])
				fmt.Printf("%s\n", stage.StdOut[i].String())
				fmt.Printf("%s\n", stage.StdErr[i].String())
			}
		}
	}
}
