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
	"github.com/ws-skeleton/che-machine-exec/exec/kubernetes-infra"
	"log"
	"os"
)

var execManager ExecManager

// ExecManager to manage exec life cycle.
type ExecManager interface {
	// Create new Exec defined by machine exec model object.
	Create(machineExec *model.MachineExec) (int, error)

	// Remove information about exec by ExecId.
	// It's can be useful in case exec error or exec exit.
	Remove(execId int) // todo rename it. f.e.: CleanUp

	// Check if exec with current id is exists
	Check(id int) (int, error)

	// Attach simple websocket connection to the exec stdIn/stdOut by unique exec id.
	Attach(id int, conn *websocket.Conn) error

	// Resize exec by unique id.
	Resize(id int, cols uint, rows uint) error
}

// Create and return new ExecManager for current infrastructure.
// Fail with panic if it is impossible.
func CreateExecManager() ExecManager {
	switch {
	case isKubernetesInfra():
		log.Println("Use kubernetes implementation")
		return kubernetes_infra.New()
	case isDockerInfra():
		log.Println("Use docker implementation")
		return docker_infra.New()
	default:
		log.Fatal("Unable to create manager for current infrastructure.")
	}

	return nil
}

// Get exec manager for current infrastructure
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
