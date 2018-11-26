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

package docker_infra

import (
	"errors"
	"fmt"
	"github.com/docker/docker/api/types"
	"github.com/docker/docker/client"
	"github.com/gorilla/websocket"
	"github.com/ws-skeleton/che-machine-exec/api/model"
	clientProvider "github.com/ws-skeleton/che-machine-exec/exec/docker-infra/client-provider"
	wsConnHandler "github.com/ws-skeleton/che-machine-exec/exec/ws-conn"
	"github.com/ws-skeleton/che-machine-exec/line-buffer"
	"github.com/ws-skeleton/che-machine-exec/utils"
	"golang.org/x/net/context"
	"strconv"
	"sync"
	"sync/atomic"
)

type DockerMachineExecManager struct {
	client *client.Client
}

type MachineExecs struct {
	mutex   *sync.Mutex
	execMap map[int]*model.MachineExec
}

var (
	machineExecs = MachineExecs{
		mutex:   &sync.Mutex{},
		execMap: make(map[int]*model.MachineExec),
	}
	prevExecID uint64 = 0
)

func New() DockerMachineExecManager {
	return DockerMachineExecManager{client: clientProvider.GetDockerClient()}
}

//  /etc/shells, echo $0, take a look /usr/sbin/nologin
func (manager DockerMachineExecManager) setUpExecShellPath(exec *model.MachineExec, containerId string) {
	if exec.Tty && len(exec.Cmd) == 0 {
		shellDetector := NewDockerShellDetector(containerId)
		if shell, err := shellDetector.DetectShell(); err == nil {
			exec.Cmd = []string{shell}
		} else {
			exec.Cmd = []string{utils.DefaultShell}
		}
	}
}

func (manager DockerMachineExecManager) Create(machineExec *model.MachineExec) (int, error) {
	container, err := FindMachineContainer(&machineExec.Identifier)
	if err != nil {
		return -1, err
	}

	manager.setUpExecShellPath(machineExec, container.ID)

	fmt.Println("found container for creation exec! id=", container.ID)

	resp, err := manager.client.ContainerExecCreate(context.Background(), container.ID, types.ExecConfig{
		Tty:          machineExec.Tty,
		AttachStdin:  true,
		AttachStdout: true,
		AttachStderr: true,
		Detach:       false,
		Cmd:          machineExec.Cmd,
	})
	if err != nil {
		return -1, err
	}

	defer machineExecs.mutex.Unlock()
	machineExecs.mutex.Lock()

	machineExec.ExecId = resp.ID
	machineExec.ID = int(atomic.AddUint64(&prevExecID, 1))
	machineExec.MsgChan = make(chan []byte)
	machineExec.WsConnsLock = &sync.Mutex{}
	machineExec.WsConns = make([]*websocket.Conn, 0)

	machineExecs.execMap[machineExec.ID] = machineExec

	fmt.Println("Create exec ", machineExec.ID, "execId", machineExec.ExecId)

	return machineExec.ID, nil
}

func (manager DockerMachineExecManager) Check(id int) (int, error) {
	machineExec := getById(id)
	if machineExec == nil {
		return -1, errors.New("Exec '" + strconv.Itoa(id) + "' was not found")
	}
	return machineExec.ID, nil
}

func (manager DockerMachineExecManager) Attach(id int, conn *websocket.Conn) error {
	machineExec := getById(id)
	if machineExec == nil {
		return errors.New("Exec '" + strconv.Itoa(id) + "' to attach was not found")
	}

	machineExec.AddWebSocket(conn)
	go wsConnHandler.ReadWebSocketData(machineExec, conn)
	go wsConnHandler.SendPingMessage(conn)

	if machineExec.Buffer != nil {
		// restore previous output.
		restoreContent := machineExec.Buffer.GetContent()
		return conn.WriteMessage(websocket.TextMessage, []byte(restoreContent))
	}

	hjr, err := manager.client.ContainerExecAttach(context.Background(), machineExec.ExecId, types.ExecConfig{
		Detach: false,
		Tty:    machineExec.Tty,
	})
	if err != nil {
		return errors.New("Failed to attach to exec " + err.Error())
	}

	machineExec.Hjr = &hjr
	machineExec.Buffer = line_buffer.New()

	machineExec.Start()

	return nil
}

func (manager DockerMachineExecManager) Resize(id int, cols uint, rows uint) error {
	machineExec := getById(id)
	if machineExec == nil {
		return errors.New("Exec to resize '" + strconv.Itoa(id) + "' was not found")
	}

	resizeParam := types.ResizeOptions{Height: rows, Width: cols}
	if err := manager.client.ContainerExecResize(context.Background(), machineExec.ExecId, resizeParam); err != nil {
		return err
	}

	return nil
}

func getById(id int) *model.MachineExec {
	defer machineExecs.mutex.Unlock()

	machineExecs.mutex.Lock()
	return machineExecs.execMap[id]
}
