package run

import (
	"fmt"
	"strings"
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

		if len(stageGroups) == 1 {
			fmt.Printf("\n==== Running stage: %s ====\n", stageGroups[0].ID)
		} else {
			fmt.Printf("\n==== Running stages: %s ====\n", stageNameCollection(stageGroups))
		}

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
	for _, stageGroups := range app.Graph {
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

func stageNameCollection(stages []*Stage) string {
	stageNames := []string{}
	for _, stage := range stages {
		stageNames = append(stageNames, stage.ID)
	}

	return strings.Join(stageNames, " and ")
}
