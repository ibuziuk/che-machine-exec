package docker_infra

import (
	"bytes"
	"context"
	"github.com/docker/docker/api/types"
	"github.com/docker/docker/client"
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

	client *client.Client
}

func NewDockerIntelligenceExec(command []string, containerId string, client *client.Client) *DockerIntelligenceExec {
	var stdOut bytes.Buffer
	return &DockerIntelligenceExec{
		command:     command,
		containerId: containerId,
		stdOut:      &stdOut,
		client:client,
	}
}

// limit this command by time to prevent hangs?
func (exec *DockerIntelligenceExec) Start() (err error) {
	resp, err := exec.client.ContainerExecCreate(context.Background(), exec.containerId, types.ExecConfig{
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

	hjr, err := exec.client.ContainerExecAttach(context.Background(), resp.ID, types.ExecConfig{
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
