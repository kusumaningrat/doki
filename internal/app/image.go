package app

import (
	"context"
	"docker-tui/internal/domain"
)

type ImageUseCase struct {
	service domain.ImageService
}

func NewImageUseCase(service domain.ImageService) *ImageUseCases {
	useCase := &ImageUseCase{service: service}
	return &ImageUseCases{
		Query:   useCase,
		Control: useCase,
	}
}

func (u *ImageUseCase) ListAllImages(ctx context.Context) ([]domain.Image, error) {
	return u.service.ListAllImages(ctx)
}

func (u *ImageUseCase) RemoveImage(ctx context.Context, id string) error {
	return u.service.RemoveImage(ctx, id)
}

func (u *ImageUseCase) ImageInspect(ctx context.Context, id string) (string, error) {
	return u.service.ImageInspect(ctx, id)
}

type ImageUseCases struct {
	Query   ImageQueryUseCase
	Control ImageControlUseCase
}

type ImageQueryUseCase interface {
	ListAllImages(ctx context.Context) ([]domain.Image, error)
	ImageInspect(ctx context.Context, id string) (string, error)
}

type ImageControlUseCase interface {
	RemoveImage(ctx context.Context, identifier string) error
}
