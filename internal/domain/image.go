package domain

import "context"

type Image struct {
	Repository string
	Tag        string
	ImageID    string
	Created    string
	Size       string
}

type ImageService interface {
	ListAllImages(ctx context.Context) ([]Image, error)
}
