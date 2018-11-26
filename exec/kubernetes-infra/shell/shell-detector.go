package shell

import (
	"github.com/pkg/errors"
	"github.com/ws-skeleton/che-machine-exec/exec/kubernetes-infra"
	"github.com/ws-skeleton/che-machine-exec/shell"
	"github.com/ws-skeleton/che-machine-exec/utils"
	"k8s.io/client-go/kubernetes/typed/core/v1"
	"k8s.io/client-go/rest"
	"strings"
)

type KubernetesShellDetector struct {
	shell.ContainerShellDetector

	api    v1.CoreV1Interface
	config *rest.Config

	namespace     string
}

func New(api v1.CoreV1Interface, config *rest.Config, namespace string) *KubernetesShellDetector {
	return &KubernetesShellDetector{
		api:       api,
		config:    config,
		namespace: namespace,
	}
}

func (detector *KubernetesShellDetector) DetectShell(containerInfo map[string]string) (shell string, err error) {
	userName, err := detector.getContainerUID(containerInfo)
	if err != nil {
		return "", err
	}

	etcPassWdContent, err := detector.getEtcPasswdContent(containerInfo)
	if err != nil {
		return "", err
	}

	return utils.ParseShellFromEtcPassWd(etcPassWdContent, userName)
}

// Get default user name for current linux container
func (detector *KubernetesShellDetector) getContainerUID(containerInfo map[string]string) (userId string, err error) {
	userIdExec := kubernetes_infra.NewIntelligenceExec(
		[]string{"id", "-u"},
		containerInfo,
		detector.namespace,
		detector.api,
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
func (detector KubernetesShellDetector) getEtcPasswdContent(containerInfo map[string]string) (etcPasswdContent string, err error) {
	etcPasswdContentExec := kubernetes_infra.NewIntelligenceExec(
		[]string{"cat", "/etc/passwd"},
		containerInfo,
		detector.namespace,
		detector.api,
		detector.config)

	if err := etcPasswdContentExec.Start(); err != nil {
		return "", err
	}

	if errInfo := etcPasswdContentExec.GetErrorOutPut(); errInfo != "" {
		return "", errors.New("Unable to get content with default shell for current user " + errInfo)
	}

	return etcPasswdContentExec.GetOutPut(), nil
}
