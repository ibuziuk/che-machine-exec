package kubernetes_infra

import (
	"github.com/pkg/errors"
	"github.com/ws-skeleton/che-machine-exec/shell"
	"github.com/ws-skeleton/che-machine-exec/utils"
	"k8s.io/client-go/kubernetes/typed/core/v1"
	"k8s.io/client-go/rest"
	"strings"
)

type KubernetesShellDetector struct {
	shell.ContainerShellDetector

	core   v1.CoreV1Interface
	config *rest.Config

	namespace     string
	containerInfo *KubernetesContainerInfo
}

func NewKubernetesShellDetector(core v1.CoreV1Interface, config *rest.Config, namespace string, containerInfo *KubernetesContainerInfo) *KubernetesShellDetector {
	return &KubernetesShellDetector{
		core:          core,
		config:        config,
		namespace:     namespace,
		containerInfo: containerInfo,
	}
}

func (detector *KubernetesShellDetector) DetectShell() (shell string, err error) {
	userName, err := detector.getContainerUserId()
	if err != nil {
		return "", err
	}

	etcPassWdContent, err := detector.getEtcPasswdContent()
	if err != nil {
		return "", err
	}

	return utils.ParseEtcPassWd(etcPassWdContent, userName)
}

// Get default user name for current linux container
func (detector *KubernetesShellDetector) getContainerUserId() (userId string, err error) {
	userIdExec := NewIntelligenceExec(
		[]string{"id", "-u"},
		detector.containerInfo,
		detector.namespace,
		detector.core,
		detector.config)

	if err := userIdExec.Start(); err != nil {
		return "Unable to create exec", err
	}

	if errInfo := userIdExec.GetErrorOutPut(); errInfo != "" {
		return "", errors.New("Unable to get userName to find shell path" + errInfo)
	}

	// user name exec returns userName with new line symbol, so remove it
	userName := strings.Replace(userIdExec.GetOutPut(), "\n", "", 1)
	return userName, nil
}

// Get /etc/passwd file content. This file stores login shell information.
func (detector KubernetesShellDetector) getEtcPasswdContent() (etcPasswdContent string, err error) {
	etcPasswdContentExec := NewIntelligenceExec(
		[]string{"cat", "/etc/passwd"},
		detector.containerInfo,
		detector.namespace,
		detector.core,
		detector.config)

	if err := etcPasswdContentExec.Start(); err != nil {
		return "", err
	}

	if errInfo := etcPasswdContentExec.GetErrorOutPut(); errInfo != "" {
		return "", errors.New("Unable to get content with default shell for current user " + errInfo)
	}

	return etcPasswdContentExec.GetOutPut(), nil
}
