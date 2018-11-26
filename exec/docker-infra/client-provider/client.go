package client_provider

import (
	"github.com/docker/docker/client"
)

type DockerClientProvider struct {
	dockerClient *client.Client
}

func New() *DockerClientProvider  {
	return &DockerClientProvider{dockerClient:createDockerClient()}
}

func createDockerClient() *client.Client {
	dockerClient, err := client.NewEnvClient()

	// set up minimal docker version 1.13.0(api version 1.25).
	dockerClient.UpdateClientVersion("1.25")
	if err != nil {
		panic(err)
	}
	return dockerClient
}

func (clientProvider *DockerClientProvider) GetDockerClient() *client.Client {
	return clientProvider.dockerClient
}
