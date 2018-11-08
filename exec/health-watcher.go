package exec

import (
	tunner "github.com/eclipse/che/agents/go-agents/core/jsonrpc"
	"github.com/ws-skeleton/che-machine-exec/api/jsonrpc"
	"github.com/ws-skeleton/che-machine-exec/api/model"
	"log"
)

// Exec health watcher. This watcher cleans up exec resources
// and sends notification to the subscribed clients in case exec error or exit.
type HealthWatcher struct {
	tunnel      *tunner.Tunnel
	execManager ExecManager
	exec        *model.MachineExec
}

// Create new exec health watcher
func NewHealthWatcher(exec *model.MachineExec, tunnel *tunner.Tunnel, execManager ExecManager) *HealthWatcher {
	return &HealthWatcher{
		exec:        exec,
		tunnel:      tunnel,
		execManager: execManager}
}

// Look at the exec health and clean up application on terminal exit/error,
// sent exit/error event to the subscribed clients
func (watcher *HealthWatcher) CleanUpOnExitOrError() {
	go func() {

		select {
		case <-watcher.exec.ExitChan:
			watcher.execManager.Remove(watcher.exec.ID)
			watcher.notifyClientsAboutExit()

		case err := <-watcher.exec.ErrorChan:
			watcher.execManager.Remove(watcher.exec.ID)
			watcher.notifyClientsAboutError(err)
		}
		log.Println("done")
	}()
}

func (watcher *HealthWatcher) notifyClientsAboutExit() {
	terminalExitEvent := &model.TerminalExitEvent{TerminalId: watcher.exec.ID}

	if err := watcher.tunnel.Notify(jsonrpc.OnTerminalExitChanged, terminalExitEvent); err != nil {
		log.Println("Unable to send close terminal message")
	}
}

func (watcher *HealthWatcher) notifyClientsAboutError(err error) {
	terminalError := &model.TerminalError{Stack: err.Error()}
	terminalErrorEvent := &model.TerminalErrorEvent{TerminalId: watcher.exec.ID, TerminalError: terminalError}

	if err := watcher.tunnel.Notify(jsonrpc.OnTerminalError, terminalErrorEvent); err != nil {
		log.Println("Unable to send error terminal message")
	}
}
