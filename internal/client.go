package docker

import (
	"bytes"
	"context" // Import context
	"docker-tui/internal/domain"
	"fmt"
	"strconv"
	"strings"
	"time"

	"github.com/docker/docker/api/types/container"
	"github.com/docker/docker/api/types/image"
	"github.com/docker/docker/client"
	"github.com/dustin/go-humanize"

	"docker-tui/internal/helper"
)

type DockerClient struct {
	cli *client.Client
}

func NewDockerClient() (*DockerClient, error) {
	cli, err := client.NewClientWithOpts(client.FromEnv, client.WithAPIVersionNegotiation())
	if err != nil {
		return nil, err
	}
	return &DockerClient{cli: cli}, nil
}

// Update all methods to accept and use ctx
func (d *DockerClient) ListAllContainers(ctx context.Context) ([]domain.Container, error) {
	containers, err := d.cli.ContainerList(ctx, container.ListOptions{All: true}) // Use ctx
	if err != nil {
		return nil, err
	}

	result := make([]domain.Container, 0, len(containers))
	for _, container := range containers {
		var ports []domain.Port
		for _, port := range container.Ports {
			ports = append(ports, domain.Port{
				IP:          port.IP,
				PrivatePort: port.PrivatePort,
				PublicPort:  port.PublicPort,
				Type:        port.Type,
			})
		}
		result = append(result, domain.Container{
			ID:      container.ID,
			Image:   container.Image,
			Command: container.Command,
			Created: humanize.Time(time.Unix(container.Created, 0)),
			Status:  container.Status,
			Ports:   ports,
			Name:    strings.TrimPrefix(container.Names[0], "/"),
		})
	}
	return result, nil
}

func (d *DockerClient) ListContainersByState(ctx context.Context, state string) ([]domain.Container, error) {
	containers, err := d.cli.ContainerList(ctx, container.ListOptions{All: true}) // Use ctx
	if err != nil {
		return nil, err
	}

	var result []domain.Container
	for _, c := range containers {
		// Note: Docker container.State is more granular than just "running" or "exited".
		// You might want to adjust this mapping depending on desired behavior.
		// For example, "running" is fine for active, but "exited" for inactive (stopped)
		// For "paused", "restarting" etc., you'd need more specific checks.
		if !strings.EqualFold(c.State, state) {
			continue
		}

		var ports []domain.Port
		for _, port := range c.Ports {
			ports = append(ports, domain.Port{
				IP:          port.IP,
				PrivatePort: port.PrivatePort,
				PublicPort:  port.PublicPort,
				Type:        port.Type,
			})
		}

		result = append(result, domain.Container{
			ID:      c.ID,
			Image:   c.Image,
			Command: c.Command,
			Created: humanize.Time(time.Unix(c.Created, 0)),
			Status:  c.Status,
			Ports:   ports,
			Name:    strings.TrimPrefix(c.Names[0], "/"),
		})
	}

	return result, nil
}

func (d *DockerClient) StartContainer(ctx context.Context, id string) error {
	return d.cli.ContainerStart(ctx, id, container.StartOptions{}) // Use ctx
}

func (d *DockerClient) ListRunningContainers(ctx context.Context) ([]domain.Container, error) {
	return d.ListContainersByState(ctx, "running")
}

func (d *DockerClient) ListStoppedContainers(ctx context.Context) ([]domain.Container, error) {
	return d.ListContainersByState(ctx, "exited")
}

func (d *DockerClient) StopContainer(ctx context.Context, id string) error {
	return d.cli.ContainerStop(ctx, id, container.StopOptions{}) // Use ctx
}

func (d *DockerClient) RestartContainer(ctx context.Context, id string) error {
	return d.cli.ContainerRestart(ctx, id, container.StopOptions{}) // Use ctx
}

func (d *DockerClient) RemoveContainer(ctx context.Context, id string, force bool) error {
	return d.cli.ContainerRemove(ctx, id, container.RemoveOptions{Force: force}) // Use ctx
}

func (d *DockerClient) GetContainerById(ctx context.Context, id string) (domain.Container, error) {
	container, err := d.cli.ContainerInspect(ctx, id) // Use ctx
	if err != nil {
		return domain.Container{}, err
	}

	var ports []domain.Port
	for port, bindings := range container.NetworkSettings.Ports {
		for _, binding := range bindings {
			hostPort, _ := strconv.ParseUint(binding.HostPort, 10, 16)
			ports = append(ports, domain.Port{
				IP:          binding.HostIP,
				PrivatePort: uint16(port.Int()),
				PublicPort:  uint16(hostPort),
				Type:        port.Proto(),
			})
		}
	}

	created, _ := strconv.ParseInt(container.Created, 10, 64)
	return domain.Container{
		ID:      container.ID,
		Image:   container.Config.Image,
		Command: strings.Join(container.Config.Cmd, " "),
		Created: humanize.Time(time.Unix(created, 0)),
		Status:  container.State.Status,
		Ports:   ports,
		Name:    strings.TrimPrefix(container.Name, "/"),
	}, nil
}

func (d *DockerClient) ContainerInspect(ctx context.Context, id string) (string, error) {
	_, jsonBytes, err := d.cli.ContainerInspectWithRaw(ctx, id, true)
	if err != nil {
		return "", fmt.Errorf("failed to inspect container %s: %w", id[:12], err)
	}

	return helper.PrettyJson(string(jsonBytes))
}

func (d *DockerClient) ListAllImages(ctx context.Context) ([]domain.Image, error) { // Added error return
	images, err := d.cli.ImageList(ctx, image.ListOptions{All: true})
	if err != nil {
		return nil, fmt.Errorf("failed to list docker images: %w", err)
	}

	var result []domain.Image

	for _, img := range images {
		imageID := img.ID[7:19] // Shorten image ID
		created := humanize.Time(time.Unix(img.Created, 0))
		size := humanize.Bytes(uint64(img.Size))

		if len(img.RepoTags) == 0 || (len(img.RepoTags) == 1 && img.RepoTags[0] == "<none>:<none>") {
			result = append(result, domain.Image{
				Repository: "<none>",
				Tag:        "<none>",
				ImageID:    imageID,
				Created:    created,
				Size:       size,
			})
		} else {
			for _, repoTag := range img.RepoTags {
				repository := "<none>"
				tag := "<none>"

				parts := strings.SplitN(repoTag, ":", 2)
				repository = parts[0]
				if len(parts) > 1 {
					tag = parts[1]
				}

				result = append(result, domain.Image{
					Repository: repository,
					Tag:        tag,
					ImageID:    imageID, // Same ImageID for all tags pointing to it
					Created:    created,
					Size:       size,
				})
			}
		}
	}
	return result, nil // Return nil for error
}

func (d *DockerClient) ImageInspect(ctx context.Context, id string) (string, error) {
	var buf bytes.Buffer
	_, err := d.cli.ImageInspect(ctx, id, client.ImageInspectWithRawResponse(&buf))
	if err != nil {
		return "", fmt.Errorf("failed to inspect image %s: %w", id[:12], err)
	}

	return helper.PrettyJson(buf.String())
}

func (d *DockerClient) RemoveImage(ctx context.Context, id string) error {
	_, err := d.cli.ImageRemove(ctx, id, image.RemoveOptions{})
	return err
}
