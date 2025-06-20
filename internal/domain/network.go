package domain

import "context"

type Network struct {
	NetworkID string
	Name      string
	Driver    string
	Scope     string
}

type NetworkService interface {
	ListAllNetworks(ctx context.Context) ([]Network, error)
	RemoveNetwork(ctx context.Context, id string) error
	InspectNetwork(ctx context.Context, name string) (string, error)
}
