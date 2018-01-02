package run

import (
	"bufio"
	"fmt"
	"os/exec"
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

	cmd := exec.Command(cmdName, cmdArgs...)
	cmdReader, err := cmd.StdoutPipe()
	if err != nil {
		return err
	}

	scanner := bufio.NewScanner(cmdReader)
	go func() {
		for scanner.Scan() {
			fmt.Printf("[docker.run.%s] %s\n", s.ID, scanner.Text())
		}
	}()

	err = cmd.Start()
	if err != nil {
		return err
	}

	err = cmd.Wait()
	if err != nil {
		return err
	}

	return nil
}
