package docker_infra

import (
	"github.com/ws-skeleton/che-machine-exec/shell"
	"github.com/ws-skeleton/che-machine-exec/utils"
)

type DockerShellDetector struct {
	shell.ContainerShellDetector
	ContainerId string
}

func NewDockerShellDetector(containerId string) *DockerShellDetector {
	return &DockerShellDetector{ContainerId:containerId}
}

func (shellDetector *DockerShellDetector) DetectShell() (shell string, err error) {
	userId, err := shellDetector.getContainerUserId()
	if err != nil {
		return "", err
	}

	etcPassWdContent, err := shellDetector.getEtcPasswdContent()
	if err != nil {
		return "", err
	}

	return utils.ParseEtcPassWd(etcPassWdContent, userId)
}

func (shellDetector *DockerShellDetector) getContainerUserId() (userId string, err error) {
	userIdExec := NewDockerIntelligenceExec([]string{"id", "-u"}, shellDetector.ContainerId)
	if err := userIdExec.Start(); err != nil {
		return "", err
	}

	userId, err = utils.ParseUID(userIdExec.GetOutPut())
	if err != nil {
		return "", err
	}

	return userId, nil
}

func (shellDetector *DockerShellDetector) getEtcPasswdContent() (etcPasswdContent string, err error) {
	etcPasswdContentExec := NewDockerIntelligenceExec([]string{"cat", "/etc/passwd"}, shellDetector.ContainerId)
	if err := etcPasswdContentExec.Start(); err != nil {
		return "", err
	}

	return string(etcPasswdContentExec.GetOutPut()), nil
}