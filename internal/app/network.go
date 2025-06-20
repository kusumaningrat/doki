package app

import (
	"context"
	"docker-tui/internal/domain"
)

type NetworkUseCase struct {
	service domain.NetworkService
}

func NewNetworkUseCase(service domain.NetworkService) *NetworkUseCases {
	useCase := &NetworkUseCase{service: service}
	return &NetworkUseCases{
		Query:   useCase,
		Control: useCase,
	}
}

func (u *NetworkUseCase) ListAllNetworks(ctx context.Context) ([]domain.Network, error) {
	return u.service.ListAllNetworks(ctx)
}

func (u *NetworkUseCase) RemoveNetwork(ctx context.Context, id string) error {
	return u.service.RemoveNetwork(ctx, id)
}

func (u *NetworkUseCase) InspectNetwork(ctx context.Context, name string) (string, error) {
	return u.service.InspectNetwork(ctx, name)
}

type NetworkUseCases struct {
	Query   NetworkQueryUseCase
	Control NetworkControlUseCase
}

type NetworkQueryUseCase interface {
	ListAllNetworks(ctx context.Context) ([]domain.Network, error)
	InspectNetwork(ctx context.Context, id string) (string, error)
}

type NetworkControlUseCase interface {
	RemoveNetwork(ctx context.Context, id string) error
}
