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
	appUseCases := &handler.AppUseCases{
		Containers: containerUseCase,
	}

	handler.RunCLI(appUseCases)
}
