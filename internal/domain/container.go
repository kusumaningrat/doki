package domain

import "context"

type Port struct {
	IP          string
	PrivatePort uint16
	PublicPort  uint16
	Type        string
}

type Container struct {
	ID      string
	Image   string
	Command string
	Created string
	Status  string
	Ports   []Port
	Name    string
}

type ContainerService interface {
	ListAllContainers(ctx context.Context) ([]Container, error)
	ListContainersByState(ctx context.Context, state string) ([]Container, error)
	GetContainerById(ctx context.Context, id string) (Container, error)
	StartContainer(ctx context.Context, id string) error
	StopContainer(ctx context.Context, id string) error
	RestartContainer(ctx context.Context, id string) error
	RemoveContainer(ctx context.Context, id string, force bool) error
	ContainerInspect(ctx context.Context, id string) (string, error)
}
