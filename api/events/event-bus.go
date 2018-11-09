package events

import (
	"github.com/eclipse/che/agents/go-agents/core/event"
	"github.com/eclipse/che/agents/go-agents/core/jsonrpc"
	"log"
)

const (
	// method names to send events with information about exec to the clients.
	OnExecExit  = "onExecExit"
	OnExecError = "onExecError"
)

// Event bus to send events with information about execs to the clients.
var ExecEventBus = event.NewBus()

// Exec Event consumer to send exec events to the clients with help json-rpc tunnel.
// INFO: Tunnel it's one of the active json-rpc connection.
type ExecEventConsumer struct {
	event.Consumer
	tunnel *jsonrpc.Tunnel
}

// Send event to the client with help json-rpc tunnel.
func (execConsumer *ExecEventConsumer) Accept(event event.E) {
	if err := execConsumer.tunnel.Notify(event.Type(), event); err != nil {
		log.Println("Unable to send event to the tunnel ", execConsumer.tunnel.ID(), "Cause: ", err.Error())
	}
}
