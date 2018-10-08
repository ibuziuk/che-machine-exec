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

package kubernetes_infra

import (
	"errors"
	"github.com/eclipse/che-machine-exec/api/model"
	wsConnHandler "github.com/eclipse/che-machine-exec/exec/ws-conn"
	"github.com/eclipse/che-machine-exec/line-buffer"
	"github.com/gorilla/websocket"
	"k8s.io/api/core/v1"
	"k8s.io/client-go/kubernetes"
	"k8s.io/client-go/kubernetes/scheme"
	corev1 "k8s.io/client-go/kubernetes/typed/core/v1"
	"k8s.io/client-go/rest"
	"k8s.io/client-go/tools/remotecommand"
	"strconv"
	"sync"
	"sync/atomic"
)

type MachineExecs struct {
	mutex   *sync.Mutex
	execMap map[int]*model.MachineExec
}

type KubernetesExecManager struct {
	core      corev1.CoreV1Interface
	nameSpace string
}

var (
	config *rest.Config

	machineExecs = MachineExecs{
		mutex:   &sync.Mutex{},
		execMap: make(map[int]*model.MachineExec),
	}
	prevExecID uint64 = 0
)

const (
	Pods = "pods"
	Exec = "exec"
	Post = "POST"
)

/**
 * Create new instance of the kubernetes exec manager
 */
func New() KubernetesExecManager {
	return KubernetesExecManager{
		core:      createClient().CoreV1(),
		nameSpace: GetNameSpace(),
	}
}

func createClient() *kubernetes.Clientset {
	var err error

	// creates the in-cluster config
	config, err = rest.InClusterConfig()
	if err != nil {
		panic(err.Error())
	}

	// creates the clientset
	clientset, err := kubernetes.NewForConfig(config)
	if err != nil {
		panic(err.Error())
	}

	return clientset
}

func (manager KubernetesExecManager) Create(machineExec *model.MachineExec) (int, error) {
	containerInfo, err := findContainerInfo(manager.core, manager.nameSpace, &machineExec.Identifier)
	if err != nil {
		return -1, err
	}

	req := manager.core.RESTClient().Post().
		Resource(Pods).
		Name(containerInfo.podName).
		Namespace(manager.nameSpace).
		SubResource(Exec).
		// set up params
		VersionedParams(&v1.PodExecOptions{
			Container: containerInfo.name,
			Command:   machineExec.Cmd,
			Stdout:    true,
			Stderr:    true,
			Stdin:     true,
			TTY:       machineExec.Tty,
		}, scheme.ParameterCodec)

	executor, err := remotecommand.NewSPDYExecutor(config, Post, req.URL())
	if err != nil {
		return -1, err
	}

	defer machineExecs.mutex.Unlock()
	machineExecs.mutex.Lock()

	machineExec.Executor = executor
	machineExec.ID = int(atomic.AddUint64(&prevExecID, 1))
	machineExec.MsgChan = make(chan []byte)
	machineExec.WsConnsLock = &sync.Mutex{}
	machineExec.WsConns = make([]*websocket.Conn, 0)
	machineExec.SizeChan = make(chan remotecommand.TerminalSize)

	machineExecs.execMap[machineExec.ID] = machineExec

	return machineExec.ID, nil
}

func (KubernetesExecManager) Check(id int) (int, error) {
	machineExec := getById(id)
	if machineExec == nil {
		return -1, errors.New("Exec '" + strconv.Itoa(id) + "' was not found")
	}
	return machineExec.ID, nil
}

func (KubernetesExecManager) Attach(id int, conn *websocket.Conn) error {
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

	ptyHandler := PtyHandlerImpl{machineExec: machineExec}
	machineExec.Buffer = line_buffer.New()

	return machineExec.Executor.Stream(remotecommand.StreamOptions{
		Stdin:             ptyHandler,
		Stdout:            ptyHandler,
		Stderr:            ptyHandler,
		TerminalSizeQueue: ptyHandler,
		Tty:               machineExec.Tty,
	})
}

func (KubernetesExecManager) Resize(id int, cols uint, rows uint) error {
	machineExec := getById(id)
	if machineExec == nil {
		return errors.New("Exec to resize '" + strconv.Itoa(id) + "' was not found")
	}

	machineExec.SizeChan <- remotecommand.TerminalSize{Width: uint16(cols), Height: uint16(rows)}
	return nil
}

func getById(id int) *model.MachineExec {
	defer machineExecs.mutex.Unlock()

	machineExecs.mutex.Lock()
	return machineExecs.execMap[id]
}
