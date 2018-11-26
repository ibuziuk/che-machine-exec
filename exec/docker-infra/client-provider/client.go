package client_provider

import "github.com/docker/docker/client"

var dockerClient *client.Client

func createDockerClient() *client.Client {
	dockerClient, err := client.NewEnvClient()

	// set up minimal docker version 1.13.0(api version 1.25).
	dockerClient.UpdateClientVersion("1.25")
	if err != nil {
		panic(err)
	}
	return dockerClient
}

func GetDockerClient() *client.Client {
	if dockerClient == nil {
		dockerClient = createDockerClient()
	}
	return dockerClient
}
