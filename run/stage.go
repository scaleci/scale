package run

import (
	"bytes"
	"context"
	"fmt"
	"strings"
	"sync"
	"time"

	"github.com/docker/docker/api/types"
	"github.com/docker/docker/api/types/container"
	"github.com/docker/docker/api/types/mount"
	"github.com/docker/docker/api/types/strslice"
	"github.com/moby/moby/client"
	"github.com/moby/moby/pkg/stdcopy"
)

type Stage struct {
	ID          string
	Command     string
	DependsOn   []string `toml:"depends_on"`
	Parallelism int

	// Array by parallelism
	StdOut      []bytes.Buffer
	StdErr      []bytes.Buffer
	StatusCode  []int
	ContainerID []string

	ParentApp *App
}

func StopStages(app *App) error {
	ctx := context.Background()
	cli, err := client.NewEnvClient()
	if err != nil {
		return err
	}

	for _, stageGroup := range app.Graph {
		for _, stage := range stageGroup {
			for _, containerID := range stage.ContainerID {
				if containerID != "" {
					if err = cli.ContainerStop(ctx, containerID, nil); err != nil {
						return err
					}

				}
			}
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

// Call each invocation in a go routine
func (s *Stage) Run(currentIndex int, totalContainers int, scaleBinaryPath string) {
	var wg sync.WaitGroup
	wg.Add(s.Parallelism)

	s.StdOut = make([]bytes.Buffer, int(s.Parallelism))
	s.StdErr = make([]bytes.Buffer, int(s.Parallelism))
	s.StatusCode = make([]int, int(s.Parallelism))
	s.ContainerID = make([]string, int(s.Parallelism))

	for index := int(0); index < s.Parallelism; index++ {
		go func(i int, ci int, t int) {
			s.RunIndividual(i, ci, t, scaleBinaryPath)
			wg.Done()
		}(index, currentIndex, totalContainers)
	}

	wg.Wait()
}

func (s *Stage) RunIndividual(parallelismIndex int, currentIndex int, totalContainers int, scaleBinaryPath string) error {
	ctx := context.Background()
	cli, err := client.NewEnvClient()
	if err != nil {
		return err
	}

	resp, err := cli.ContainerCreate(ctx,
		&container.Config{
			Image: s.ParentApp.ImageName(),
			Env:   envForStage(s, parallelismIndex, currentIndex, totalContainers),
			Cmd: strslice.StrSlice{
				"/bin/bash",
				"-c",
				s.Command,
			},
		},
		&container.HostConfig{
			Mounts: []mount.Mount{
				mount.Mount{
					Source: scaleBinaryPath,
					Target: "/bin/scale",
					Type:   "bind",
				},
			},
		}, nil, "")

	if err != nil {
		return err
	}
	s.ContainerID[currentIndex] = resp.ID
	if err := cli.ContainerStart(ctx, resp.ID, types.ContainerStartOptions{}); err != nil {
		return err
	}

	status, err := cli.ContainerWait(ctx, resp.ID)
	if err != nil {
		return err
	}
	s.StatusCode[currentIndex] = int(status)

	reader, err := cli.ContainerLogs(ctx, resp.ID, types.ContainerLogsOptions{ShowStdout: true, ShowStderr: true})
	if err != nil {
		return err
	}

	_, err = stdcopy.StdCopy(&s.StdOut[parallelismIndex],
		&s.StdErr[parallelismIndex],
		reader)
	if err != nil {
		return err
	}

	return nil
}

func envForStage(s *Stage, parallelismIndex int, currentIndex int, totalContainers int) []string {
	envArray := []string{}

	for configKey, configVal := range s.ParentApp.GlobalConfig.Env {
		envArray = append(envArray, fmt.Sprintf("%s=%s", configKey, configVal))
	}

	// Inject "running state" into the container
	// MAX is the total number of containers (across all stages) that are part of this run
	envArray = append(envArray, fmt.Sprintf("SCALE_CI_MAX=%d", totalContainers))
	// TOTAL is the parallelism of the current stage
	envArray = append(envArray, fmt.Sprintf("SCALE_CI_TOTAL=%d", s.Parallelism))
	// INDEX is the index of the current running container
	envArray = append(envArray, fmt.Sprintf("SCALE_CI_INDEX=%d", parallelismIndex))

	for _, service := range s.ParentApp.Services {
		parallelIndex := int(-1)
		if s.Parallelism > int(1) {
			parallelIndex = currentIndex + parallelismIndex
		}

		for serviceKey, serviceVal := range service.Env(parallelIndex) {
			envArray = append(envArray, fmt.Sprintf("%s=%s", serviceKey, serviceVal))
		}
	}

	return envArray
}
