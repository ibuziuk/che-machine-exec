package events

import (
	"github.com/eclipse/che/agents/go-agents/core/event"
	"github.com/eclipse/che/agents/go-agents/core/jsonrpc"
	"log"
	"time"
)

const (
	// method names to send events with information about exec to the clients.
	OnExecExit  = "onExecExit"
	OnExecError = "onExecError"

	CleanUpPeriod = 5
)

// Event bus to send events with information about execs to the clients.
var ExecEventBus = event.NewBus()

// Periodically clean up event bus from closed connections.
func PeriodicallyCleanUpBus() {
	go func() {
		ticker := time.NewTicker(CleanUpPeriod * time.Second)
		for range ticker.C {
			ExecEventBus.RmIf(func(c event.Consumer) bool {
				if execConsumer, ok := c.(*ExecEventConsumer); ok {
					return execConsumer.Tunnel.IsClosed()
				}
				return false
			})
		}
	}()
}

// Exec Event consumer to send exec events to the clients with help json-rpc tunnel.
// INFO: Tunnel it's one of the active json-rpc connection.
type ExecEventConsumer struct {
	event.Consumer
	Tunnel *jsonrpc.Tunnel
}

// Send event to the client with help json-rpc tunnel.
func (execConsumer *ExecEventConsumer) Accept(event event.E) {
	if !execConsumer.Tunnel.IsClosed() {
		if err := execConsumer.Tunnel.Notify(event.Type(), event); err != nil {
			log.Println("Unable to send event to the tunnel: ", execConsumer.Tunnel.ID(), "Cause: ", err.Error())
		}
	}
}
