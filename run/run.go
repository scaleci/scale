package run

import (
	"fmt"
	"strings"
	"sync"
	"time"
)

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

func RunStages(app *App, scaleBinaryPath string) {
	var total int = 0
	var index int = 0
	var lastCompletedStageGroupIndex int = 0

	for _, s := range app.Stages {
		total += s.Parallelism
	}

	ticker := time.NewTicker(time.Second)
	go func() {
		for _ = range ticker.C {
			fmt.Printf(".")
		}
	}()

	for stageGroupIndex, stageGroups := range app.Graph {
		var wg sync.WaitGroup
		wg.Add(len(stageGroups))

		if len(stageGroups) == 1 {
			fmt.Printf("\n==== Running stage: %s ====\n", stageGroups[0].ID)
		} else {
			fmt.Printf("\n==== Running stages in parallel: %s ====\n", stageNameCollection(stageGroups))
		}

		for i := range stageGroups {
			stage := stageGroups[i]

			go func(i int) {
				stage.Run(i, total, scaleBinaryPath)
				wg.Done()
			}(index)

			index += stage.Parallelism
		}

		wg.Wait()

		status := 0
		// If any one of the stages returned with
		// a non-zero return code, status will be > 0
		for _, s := range stageGroups {
			for _, st := range s.StatusCode {
				status += st
			}
		}

		if status > 0 {
			lastCompletedStageGroupIndex = stageGroupIndex
			break
		}
	}

	ticker.Stop()
	for i, stageGroups := range app.Graph {
		if i <= lastCompletedStageGroupIndex {
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
}

func stageNameCollection(stages []*Stage) string {
	stageNames := []string{}
	for _, stage := range stages {
		stageNames = append(stageNames, stage.ID)
	}

	return strings.Join(stageNames, " and ")
}
