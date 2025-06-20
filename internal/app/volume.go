package app

import (
	"context"
	"docker-tui/internal/domain"
)

type VolumeUseCase struct {
	service domain.VolumeService
}

func NewVolumeUseCase(service domain.VolumeService) *VolumeUseCases {
	useCase := &VolumeUseCase{service: service}
	return &VolumeUseCases{
		Query:   useCase,
		Control: useCase,
	}
}

func (u *VolumeUseCase) ListAllVolumes(ctx context.Context) ([]domain.Volume, error) {
	return u.service.ListAllVolumes(ctx)
}

func (u *VolumeUseCase) RemoveVolume(ctx context.Context, id string) error {
	return u.service.RemoveVolume(ctx, id)
}

func (u *VolumeUseCase) VolumeInspect(ctx context.Context, name string) (string, error) {
	return u.service.VolumeInspect(ctx, name)
}

type VolumeUseCases struct {
	Query   VolumeQueryUseCase
	Control VolumeControlUseCase
}

type VolumeQueryUseCase interface {
	ListAllVolumes(ctx context.Context) ([]domain.Volume, error)
	VolumeInspect(ctx context.Context, name string) (string, error)
}

type VolumeControlUseCase interface {
	RemoveVolume(ctx context.Context, id string) error
}
