package docker_infra

import (
	"bytes"
	"context"
	"github.com/docker/docker/api/types"
	clientProvider "github.com/ws-skeleton/che-machine-exec/exec/docker-infra/client-provider"
	"io"
)

// Exec to get some information from container.
// command for such exec should be
// "endless" and simple(For example: "whoami", "arch", "env").
// It should not be shell based command.
// This exec is always not "tty" and doesn't provide sending input to the command.
type DockerIntelligenceExec struct {
	// command with arguments
	command []string

	// unique docker container id
	containerId string

	// buffer to store exec output
	stdOut *bytes.Buffer
}

func NewDockerIntelligenceExec(command []string, containerId string) *DockerIntelligenceExec {
	var stdOut bytes.Buffer
	return &DockerIntelligenceExec{
		command:     command,
		containerId: containerId,
		stdOut:      &stdOut,
	}
}

// limit this command by time to prevent hangs?
func (exec *DockerIntelligenceExec) Start() (err error) {
	resp, err := clientProvider.GetDockerClient().ContainerExecCreate(context.Background(), exec.containerId, types.ExecConfig{
		Tty:          false,
		AttachStdin:  false,
		AttachStdout: true,
		AttachStderr: true,
		Detach:       false,
		Cmd:          exec.command,
	})
	if err != nil {
		return err
	}

	hjr, err := clientProvider.GetDockerClient().ContainerExecAttach(context.Background(), resp.ID, types.ExecConfig{
		Detach: false,
		Tty:    false,
	})
	if err != nil {
		return err
	}

	_, err = io.Copy(exec.stdOut, hjr.Reader)
	return err
}

// Get exec output content
func (exec *DockerIntelligenceExec) GetOutPut() []byte {
	return exec.stdOut.Bytes()
}
