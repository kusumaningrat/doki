package app

import (
	"context"
	"docker-tui/internal/domain"
)

type ContainerUseCase struct {
	service domain.ContainerService
}

func NewContainerUseCase(service domain.ContainerService) *ContainerUseCase {
	return &ContainerUseCase{service: service}
}

func (u *ContainerUseCase) ListAllContainers(ctx context.Context) ([]domain.Container, error) {
	return u.service.ListAllContainers(ctx)
}

func (u *ContainerUseCase) ListContainersByState(ctx context.Context, state string) ([]domain.Container, error) {
	return u.service.ListContainersByState(ctx, state)
}

func (u *ContainerUseCase) GetContainerById(ctx context.Context, id string) (domain.Container, error) {
	return u.service.GetContainerById(ctx, id)
}
func (u *ContainerUseCase) StartContainer(ctx context.Context, id string) error {
	return u.service.StartContainer(ctx, id)
}

func (u *ContainerUseCase) StopContainer(ctx context.Context, id string) error {
	return u.service.StopContainer(ctx, id)
}

func (u *ContainerUseCase) RestartContainer(ctx context.Context, id string) error {
	return u.service.RestartContainer(ctx, id)
}

func (u *ContainerUseCase) RemoveContainer(ctx context.Context, id string) error {
	return u.service.RemoveContainer(ctx, id)
}
