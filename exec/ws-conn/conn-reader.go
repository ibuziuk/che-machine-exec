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

package ws_conn

import (
	"github.com/gorilla/websocket"
	"github.com/ws-skeleton/che-machine-exec/api/model"
	"log"
	"net"
)

func ReadWebSocketData(machineExec *model.MachineExec, wsConn *websocket.Conn) {
	defer machineExec.RemoveWebSocket(wsConn)

	for {
		msgType, wsBytes, err := wsConn.ReadMessage()
		if err != nil && IsNormalWSError(err) {
			log.Printf("failed to read ws-conn message") // todo better handle ws-conn error
			return
		}

		if msgType != websocket.TextMessage {
			continue
		}

		machineExec.MsgChan <- wsBytes
	}
}

func IsNormalWSError(err error) bool {
	closeErr, ok := err.(*websocket.CloseError)
	if ok && (closeErr.Code == websocket.CloseGoingAway || closeErr.Code == websocket.CloseNormalClosure) {
		return true
	}
	_, ok = err.(*net.OpError)
	return ok
}
