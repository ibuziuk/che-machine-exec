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
	"log"
	"time"
)

const PingPeriod = 30 * time.Second

func SendPingMessage(wsConn *websocket.Conn) {
	ticker := time.NewTicker(PingPeriod)
	defer ticker.Stop()

	for range ticker.C {
		if err := wsConn.WriteMessage(websocket.PingMessage, []byte{}); err != nil {
			log.Printf("Error occurs on sending ping message to ws-conn. %v", err)
			return
		}
	}
}
