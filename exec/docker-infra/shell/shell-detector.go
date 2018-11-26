package shell

import (
	"github.com/docker/docker/client"
	"github.com/ws-skeleton/che-machine-exec/exec/docker-infra"
	"github.com/ws-skeleton/che-machine-exec/shell"
	"github.com/ws-skeleton/che-machine-exec/utils"
)

type DockerShellDetector struct {
	shell.ContainerShellDetector

	client *client.Client
}

// Create new shell detector to get default shell for container on the docker infra.
func New(client *client.Client) *DockerShellDetector {
	return &DockerShellDetector{client:client}
}

// Detect default shell for current user inside container.
func (shellDetector *DockerShellDetector) DetectShell(containerInfo map[string]string) (shell string, err error) {
	userId, err := shellDetector.getContainerUID(containerInfo)
	if err != nil {
		return "", err
	}

	etcPassWdContent, err := shellDetector.getEtcPasswdContent(containerInfo)
	if err != nil {
		return "", err
	}

	return utils.ParseShellFromEtcPassWd(etcPassWdContent, userId)
}

// Get container user id
func (shellDetector *DockerShellDetector) getContainerUID(containerInfo map[string]string) (userId string, err error) {
	userIdExec := docker_infra.NewDockerIntelligenceExec([]string{"id", "-u"}, containerInfo[docker_infra.ContainerId], shellDetector.client)
	if err := userIdExec.Start(); err != nil {
		return "", err
	}

	userId, err = utils.ParseUID(userIdExec.GetOutPut())
	if err != nil {
		return "", err
	}

	return userId, nil
}

// Get content of the file /etc/passwd. This file contains information about defualt shell per user ID.
func (shellDetector *DockerShellDetector) getEtcPasswdContent(containerInfo map[string]string) (etcPasswdContent string, err error) {
	etcPasswdContentExec := docker_infra.NewDockerIntelligenceExec([]string{"cat", "/etc/passwd"}, containerInfo[docker_infra.ContainerId], shellDetector.client)
	if err := etcPasswdContentExec.Start(); err != nil {
		return "", err
	}

	return string(etcPasswdContentExec.GetOutPut()), nil
}