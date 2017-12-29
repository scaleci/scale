package run

import (
	"bufio"
	"fmt"
	"os/exec"
	"strings"
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
		err := startService(s)
		if err != nil {
			return err
		}
	}

	return nil
}

func StopServices(app *App) error {
	for _, s := range app.Services {
		err := stopService(s)
		if err != nil {
			return err
		}
	}

	return nil
}

func startService(s *Service) error {
	cmdName := "docker"
	cmdArgs := []string{
		"run",
		"-d",
	}

	cmdArgs = append(cmdArgs, s.PortsAsArgs()...)
	cmdArgs = append(cmdArgs, s.Image)
	cmd := exec.Command(cmdName, cmdArgs...)
	out, err := cmd.Output()
	if err != nil {
		return err
	}
	s.ContainerID = strings.Trim(string(out), "\n")

	fmt.Printf("[%s] Started with container ID %s\n", s.ID, s.ContainerID)
	return nil
}

func stopService(s *Service) error {
	if s.ContainerID != "" {
		cmd := exec.Command("docker", "stop", s.ContainerID, "-t", 5)
		_, err := cmd.Output()
		if err != nil {
			return err
		}

		fmt.Printf("[%s] Stopped with container ID %s\n", s.ID, s.ContainerID)
		return nil
	}

	return nil
}
