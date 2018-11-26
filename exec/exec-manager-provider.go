//
// Copyright (c) 2012-2018 Red Hat, Inc.
// This program and the accompanying materials are made
// available under the terms of the Eclipse Public License 2.0
// which is available at https://www.eclipse.org/legal/epl-2.0/
//
// SPDX-License-Identifier: EPL-2.0
//
// Contributors:
//   Red Hat, Inc. - initial API and implementation
//

package exec

import (
	"github.com/gorilla/websocket"
	"github.com/ws-skeleton/che-machine-exec/api/model"
	"github.com/ws-skeleton/che-machine-exec/exec/docker-infra"
	dockerContainerFilter "github.com/ws-skeleton/che-machine-exec/exec/docker-infra/filter"
	dockerShellDetector "github.com/ws-skeleton/che-machine-exec/exec/docker-infra/shell"
	dockerClientProvider "github.com/ws-skeleton/che-machine-exec/exec/docker-infra/client-provider"
	"github.com/ws-skeleton/che-machine-exec/exec/kubernetes-infra"
	kubernetesContainerFilter "github.com/ws-skeleton/che-machine-exec/exec/kubernetes-infra/filter"
	kubernetesShellDetector "github.com/ws-skeleton/che-machine-exec/exec/kubernetes-infra/shell"
	kubernetesClientProvider "github.com/ws-skeleton/che-machine-exec/exec/kubernetes-infra/client-provider"
	"github.com/ws-skeleton/che-machine-exec/exec/kubernetes-infra/namespace"
	"log"
	"os"
)

var execManager ExecManager

type ExecManager interface {
	Create(*model.MachineExec) (int, error)
	Check(id int) (int, error)
	Attach(id int, conn *websocket.Conn) error
	Resize(id int, cols uint, rows uint) error
}

func CreateExecManager() ExecManager {
	var manager ExecManager

	if isKubernetesInfra() {
		log.Println("Use kubernetes implementation")

		nameSpace := namespace.NewNameSpaceProvider().GetNameSpace()
		clientProvider := kubernetesClientProvider.New()
		client := clientProvider.GetKubernetesClient()
		config := clientProvider.GetKubernetesConfig()

		shellDetector := kubernetesShellDetector.New(client.CoreV1(), config, nameSpace)
		containerFilter := kubernetesContainerFilter.New(nameSpace, client.CoreV1())

		manager = kubernetes_infra.New(nameSpace, client.CoreV1(), config, containerFilter, shellDetector)
	} else if isDockerInfra() {
		log.Println("Use docker implementation")

		dockerClient := dockerClientProvider.New().GetDockerClient()

		containerFilter := dockerContainerFilter.New(dockerClient)
		shellDetector := dockerShellDetector.New(dockerClient)

		manager = docker_infra.New(dockerClient, containerFilter, shellDetector)
	}

	// todo what we should do in the case, when we have no implementation. Should we return stub, or only log error or throw panic...

	return manager
}

func GetExecManager() ExecManager {
	if execManager == nil {
		execManager = CreateExecManager()
	}
	return execManager
}

func isKubernetesInfra() bool {
	stat, err := os.Stat("/var/run/secrets/kubernetes.io/serviceaccount")
	if err == nil && stat.IsDir() {
		return true
	}

	return false
}

func isDockerInfra() bool {
	stat, err := os.Stat("/var/run/docker.sock")
	if err == nil && !stat.Mode().IsRegular() && !stat.IsDir() {
		return true
	}

	return false
}
