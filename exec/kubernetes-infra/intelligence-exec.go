package kubernetes_infra

import (
	"bytes"
	corev1 "k8s.io/api/core/v1"
	"k8s.io/client-go/kubernetes/scheme"
	"k8s.io/client-go/kubernetes/typed/core/v1"
	"k8s.io/client-go/rest"
	"k8s.io/client-go/tools/remotecommand"
)

// Exec to get some information from container.
// command for such exec should be
// "endless" and simple(For example: "whoami", "arch", "env").
// It should not be shell based command.
// This exec is always not "tty" and doesn't provide sending input to the command.
type KubernetesIntelligenceExec struct {
	// command with arguments
	command []string

	// information to find container
	namespace     string
	containerInfo *KubernetesContainerInfo

	// stdOut/stdErr buffers
	stdOut *bytes.Buffer
	stdErr *bytes.Buffer

	// api to spawn exec
	core   v1.CoreV1Interface
	config *rest.Config
}

func NewIntelligenceExec(command []string, containerInfo *KubernetesContainerInfo, namespace string, core v1.CoreV1Interface, config *rest.Config) *KubernetesIntelligenceExec {
	var stdOut, stdErr bytes.Buffer
	return &KubernetesIntelligenceExec{
		command:       command,
		containerInfo: containerInfo,
		namespace:     namespace,
		stdOut:        &stdOut,
		stdErr:        &stdErr,
		core:          core,
		config:        config,
	}
}

// limit this command by time to prevent hangs?
func (exec *KubernetesIntelligenceExec) Start() (err error) {
	req := exec.core.RESTClient().
		Post().
		Namespace(exec.namespace).
		Resource(Pods).
		Name(exec.containerInfo.PodName).
		SubResource(Exec).
		// set up params
		VersionedParams(&corev1.PodExecOptions{
			Container: exec.containerInfo.Name,
			Command:   exec.command,
			Stdout:    true,
			Stderr:    true,
			// no input reader, spawns exec only to get some info from container
			Stdin: false,
			// no tty, exec should launch simple no terminal command
			TTY: false,
		}, scheme.ParameterCodec)

	executor, err := remotecommand.NewSPDYExecutor(exec.config, "POST", req.URL())
	if err != nil {
		return err
	}

	err = executor.Stream(remotecommand.StreamOptions{
		Stdout: exec.stdOut,
		Stderr: exec.stdErr,
		Tty:    false,
	})

	return err
}

// Get exec output content
func (exec *KubernetesIntelligenceExec) GetOutPut() string {
	return string(exec.stdOut.Bytes())
}

// Get exec error content
func (exec *KubernetesIntelligenceExec) GetErrorOutPut() string {
	return string(exec.stdErr.Bytes())
}
