package run

type Service struct {
	ID          string
	Image       string
	Ports       []string
	ContainerID string
}

func (s *Service) PortsAsArgs() []string {
	portsAsArgs := []string{}

	for _, p := range s.Ports {
		portsAsArgs = append(portsAsArgs, []string{"-p", p}...)
	}

	return portsAsArgs
}
