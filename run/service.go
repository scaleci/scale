package run

import (
	"fmt"
	"os/exec"
	"strings"
)

type Service struct {
	ID          string
	Image       string
	Port        string
	ContainerID string
	Protocol    string
	// Set once the container is started
	HostAndPort string
}

func (s *Service) PortAsArgs() []string {
	port := strings.Split(s.Port, "/")[0]
	return []string{"-p", fmt.Sprintf("%s:%s", port, port)}
}

func (s *Service) Start() error {
	cmdName := "docker"
	cmdArgs := []string{
		"run",
		"-d",
	}

	cmdArgs = append(cmdArgs, s.PortAsArgs()...)
	cmdArgs = append(cmdArgs, s.Image)
	cmd := exec.Command(cmdName, cmdArgs...)
	out, err := cmd.Output()
	if err != nil {
		return err
	}
	s.ContainerID = strings.Trim(string(out), "\n")
	err = s.SetHostAndPort()
	if err != nil {
		return err
	}

	fmt.Printf("[%s] Started with container ID %s\n", s.ID, s.ContainerID)
	return nil
}

func (s *Service) Stop() error {
	if s.ContainerID != "" {
		cmd := exec.Command("docker", "stop", s.ContainerID, "-t", "5")
		_, err := cmd.Output()
		if err != nil {
			return err
		}

		fmt.Printf("[%s] Stopped with container ID %s\n", s.ID, s.ContainerID)
		return nil
	}

	return nil
}

func (s *Service) SetHostAndPort() error {
	if s.ContainerID != "" {
		cmdName := "docker"
		cmdArgs := []string{
			"inspect",
			fmt.Sprintf("--format={{range .NetworkSettings.Networks}}{{.IPAddress}}{{end}}:{{(index (index .NetworkSettings.Ports \"%s\") 0).HostPort}}", s.Port),
			s.ContainerID,
		}
		cmd := exec.Command(cmdName, cmdArgs...)
		out, err := cmd.Output()
		if err != nil {
			return err
		}
		s.HostAndPort = strings.Trim(string(out), "\n")

		return nil
	}

	return fmt.Errorf("ContainerID is not set")
}
