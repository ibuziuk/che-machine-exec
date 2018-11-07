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

package jsonrpc

import (
	"fmt"
	"github.com/ws-skeleton/che-machine-exec/api/model"

	"github.com/eclipse/che/agents/go-agents/core/jsonrpc"
	"github.com/ws-skeleton/che-machine-exec/exec"
	"log"
	"strconv"
)

type IdParam struct {
	Id int `json:"id"`
}

type OperationResult struct {
	Id   int    `json:"id"`
	Text string `json:"text"`
}

type ResizeParam struct {
	Id   int  `json:"id"`
	Cols uint `json:"cols"`
	Rows uint `json:"rows"`
}

var (
	execManager = exec.GetExecManager()
)

func jsonRpcCreateExec(tunnel *jsonrpc.Tunnel, params interface{}, t jsonrpc.RespTransmitter) {
	machineExec := params.(*model.MachineExec)

	id, err := execManager.Create(machineExec,
		func(done bool) {
			terminalExitEvent := &model.TerminalExitEvent{TerminalId: machineExec.ID}

			if err := tunnel.Notify("onTerminalExitChanged", terminalExitEvent); err != nil {
				fmt.Println("Unable to send close terminal message")
			}
		},
		func(err error) {
			terminalError := &model.TerminalError{Stack: err.Error()}
			terminalErrorEvent := &model.TerminalErrorEvent{TerminalId: machineExec.ID, TerminalError: terminalError}

			if err := tunnel.Notify("onTerminalError", terminalErrorEvent); err != nil {
				fmt.Println("Unable to send error terminal message")
			}
		})
	if err != nil {
		log.Println("Unable to create machine exec. Cause: ", err.Error()) // rework to terminal error too
		t.SendError(jsonrpc.NewArgsError(err))
	}

	t.Send(id)
}

func jsonRpcCheckExec(_ *jsonrpc.Tunnel, params interface{}, t jsonrpc.RespTransmitter) {
	idParam := params.(*IdParam)

	id, err := execManager.Check(idParam.Id)
	if err != nil {
		t.SendError(jsonrpc.NewArgsError(err))
	}

	t.Send(id)
}

func jsonRpcResizeExec(_ *jsonrpc.Tunnel, params interface{}) (interface{}, error) {
	resizeParam := params.(*ResizeParam)

	if err := execManager.Resize(resizeParam.Id, resizeParam.Cols, resizeParam.Rows); err != nil {
		return nil, jsonrpc.NewArgsError(err)
	}

	return &OperationResult{
		Id: resizeParam.Id, Text: "Exec with id " + strconv.Itoa(resizeParam.Id) + "  was successfully resized",
	}, nil
}
