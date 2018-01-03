package run

import (
	"fmt"
	"os/exec"
	"strings"
)

type Service struct {
	ID          string
	Image       string
	Host        string
	Port        string
	ContainerID string
	Protocol    string
	// Host:Port -> Set once the container is started
	Socket string
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
	err = s.SetSocket()
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

func (s *Service) SetSocket() error {
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
		s.Socket = strings.Trim(string(out), "\n")
		s.Host = strings.Split(s.Socket, ":")[0]

		return nil
	}

	return fmt.Errorf("ContainerID is not set")
}

func (s *Service) Env() map[string]string {
	env := map[string]string{}
	protocol := s.ID

	if s.Protocol != "" {
		protocol = s.Protocol
	}

	urlKey := fmt.Sprintf("%s_URL", strings.ToUpper(s.ID))
	protocolUrlKey := fmt.Sprintf("%s_URL", strings.ToUpper(protocol))
	urlValue := fmt.Sprintf("%s://%s", protocol, s.Socket)
	hostKey := fmt.Sprintf("%s_HOST", strings.ToUpper(s.ID))
	protocolHostKey := fmt.Sprintf("%s_HOST", strings.ToUpper(protocol))
	hostValue := s.Host
	portKey := fmt.Sprintf("%s_PORT", strings.ToUpper(s.ID))
	protocolPortKey := fmt.Sprintf("%s_PORT", strings.ToUpper(protocol))
	portValue := strings.Split(s.Port, "/")[0]

	env[urlKey] = urlValue
	env[protocolUrlKey] = urlValue
	env[hostKey] = hostValue
	env[protocolHostKey] = hostValue
	env[portKey] = portValue
	env[protocolPortKey] = portValue

	return env
}
