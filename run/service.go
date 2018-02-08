package run

import (
	"context"
	"fmt"
	"strings"

	"github.com/docker/docker/api/types"
	"github.com/docker/docker/api/types/container"
	"github.com/docker/go-connections/nat"
	"github.com/moby/moby/client"
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

func (s *Service) Start() error {
	ctx := context.Background()
	cli, err := client.NewEnvClient()
	if err != nil {
		return err
	}

	_, _, err = cli.ImageInspectWithRaw(context.Background(), scaleBinaryPath)

	if err != nil && client.IsErrImageNotFound(err) {
		imagePullBody, err := cli.ImagePull(ctx, s.Image, types.ImagePullOptions{})
		if err != nil {
			return err
		}

		if err := StreamDockerResponse(imagePullBody, "status", "error"); err != nil {
			return err
		}
	} else if err != nil {
		return err
	}

	resp, err := cli.ContainerCreate(ctx,
		&container.Config{
			Image: s.Image,
		},
		&container.HostConfig{
			PortBindings: nat.PortMap{nat.Port(s.Port): []nat.PortBinding{nat.PortBinding{HostPort: s.Port}}},
		}, nil, "")

	if err != nil {
		return err
	}

	s.ContainerID = resp.ID
	if err := cli.ContainerStart(ctx, s.ContainerID, types.ContainerStartOptions{}); err != nil {
		return err
	}

	err = s.SetSocket()
	if err != nil {
		return err
	}

	fmt.Printf("[%s] Started with container ID %s\n", s.ID, s.ContainerID)
	return nil
}

func (s *Service) Stop() error {
	if s.ContainerID != "" {
		cli, err := client.NewEnvClient()
		if err != nil {
			return err
		}

		if err = cli.ContainerStop(context.Background(), s.ContainerID, nil); err != nil {
			return err
		}

		fmt.Printf("[%s] Stopped with container ID %s\n", s.ID, s.ContainerID)
		return nil
	}

	return nil
}

func (s *Service) SetSocket() error {
	if s.ContainerID != "" {
		host, port := "", ""
		ctx := context.Background()
		cli, err := client.NewEnvClient()
		if err != nil {
			return err
		}

		containerJSON, err := cli.ContainerInspect(ctx, s.ContainerID)
		if err != nil {
			return err
		}

		if endpointSettings := containerJSON.NetworkSettings.Networks["bridge"]; endpointSettings != nil {
			host = endpointSettings.IPAddress

			portMap := containerJSON.NetworkSettings.Ports[nat.Port(s.Port)]
			if len(portMap) > 0 {
				port = portMap[0].HostPort
			}
		}

		if host == "" || port == "" {
			return fmt.Errorf("Could not find host/port for service: %s\n", s.ID)
		}

		s.Host = host
		s.Socket = fmt.Sprintf("%s:%s", host, port)

		return nil
	}

	return fmt.Errorf("ContainerID is not set")
}

func (s *Service) Env(index int) map[string]string {
	env := map[string]string{}
	protocol := s.ID

	if s.Protocol != "" {
		protocol = s.Protocol
	}

	urlKey := fmt.Sprintf("%s_URL", strings.ToUpper(s.ID))
	protocolUrlKey := fmt.Sprintf("%s_URL", strings.ToUpper(protocol))
	urlValue := fmt.Sprintf("%s://%s", protocol, s.Socket)
	if index > int(-1) {
		urlValue = fmt.Sprintf("%s://%s/%s", protocol, s.Socket, s.Database(index))
	}
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

func (s *Service) Database(index int) string {
	protocol := s.ID
	database := fmt.Sprintf("%d", index)

	if s.Protocol != "" {
		protocol = s.Protocol
	}

	if protocol == "postgres" {
		database = fmt.Sprintf("scale%d", index)
	}

	return database
}
