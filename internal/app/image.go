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

type ImageUseCases struct {
	Query   ImageQueryUseCase
	Control ImageControlUseCase
}

type ImageQueryUseCase interface {
	ListAllImages(ctx context.Context) ([]domain.Image, error)
}

type ImageControlUseCase interface {
}
