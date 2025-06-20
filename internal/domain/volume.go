package domain

import "context"

type Volume struct {
	Name       string
	Driver     string
	Mountpoint string
	Labels     map[string]string
}

type VolumeService interface {
	ListAllVolumes(ctx context.Context) ([]Volume, error)
	RemoveVolume(ctx context.Context, id string) error
	VolumeInspect(ctx context.Context, name string) (string, error)
}
