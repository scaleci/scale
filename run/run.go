package run

import (
	"bufio"
	"fmt"
	"os/exec"
)

func Build(app *App) error {
	cmdName := "docker"
	cmdArgs := []string{
		"build",
		".",
		"-f", app.GlobalConfig.BuildWith,
		"-t", app.ImageName()}

	cmd := exec.Command(cmdName, cmdArgs...)
	cmdReader, err := cmd.StdoutPipe()
	if err != nil {
		return err
	}

	scanner := bufio.NewScanner(cmdReader)
	go func() {
		for scanner.Scan() {
			fmt.Printf("[docker build] %s\n", scanner.Text())
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
	for _, s := range app.Stages {
		if len(s.DependsOn) == 0 {
			err := s.Run()
			if err != nil {
				return err
			}
		}
	}

	return nil
}
