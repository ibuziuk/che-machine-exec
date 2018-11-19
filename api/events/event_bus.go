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

type ExecEventBus interface {
	// Periodically clean up event bus from closed connections.
	PeriodicallyCleanUpBus()
	// Send event `e` to the all registered consumers
	Pub(e event.E)
	// Define event consumers and event types
	SubAny(consumer event.Consumer, types ...string)
}

type ExecEventBusImpl struct {
	*event.Bus
}

// Event bus to send events with information about execs to the clients.
var EventBus = &ExecEventBusImpl{Bus: event.NewBus()}

// Periodically clean up event bus from closed connections.
func (eventBus *ExecEventBusImpl) PeriodicallyCleanUpBus() {
	go func() {
		ticker := time.NewTicker(CleanUpPeriod * time.Second)
		for range ticker.C {
			eventBus.RmIf(func(c event.Consumer) bool {
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
