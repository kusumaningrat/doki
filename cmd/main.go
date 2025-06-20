package main

import (
	docker "docker-tui/internal"
	"docker-tui/internal/app"
	"docker-tui/internal/handler"
	"log"
)

func main() {
	dockerClient, err := docker.NewDockerClient()
	if err != nil {
		log.Fatalf("Failed to initialize docker client: %v", err)
	}

	containerUseCase := app.NewContainerUseCase(dockerClient)
	imageUseCase := app.NewImageUseCase(dockerClient)
	volumeUseCase := app.NewVolumeUseCase(dockerClient)
	networkUseCase := app.NewNetworkUseCase(dockerClient)

	appUseCases := &handler.AppUseCases{
		Containers: containerUseCase,
		Images:     imageUseCase,
		Volumes:    volumeUseCase,
		Networks:   networkUseCase,
	}

	handler.RunCLI(appUseCases)
}
