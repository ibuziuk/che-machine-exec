package exec

import (
	"github.com/ws-skeleton/che-machine-exec/api/events"
	"github.com/ws-skeleton/che-machine-exec/api/model"
	"log"
)

// Exec health watcher. This watcher cleans up exec resources
// and sends notification to the subscribed clients in case exec error or exit.
type HealthWatcher struct {
	execManager ExecManager
	exec        *model.MachineExec
}

// Create new exec health watcher
func NewHealthWatcher(exec *model.MachineExec) *HealthWatcher {
	return &HealthWatcher{
		exec:        exec,
		execManager: GetExecManager()}
}

// Look at the exec health and clean up application on exec exit/error,
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
	execExitEvent := &model.ExecExitEvent{ExecId: watcher.exec.ID}

	events.ExecEventBus.Pub(execExitEvent)
}

func (watcher *HealthWatcher) notifyClientsAboutError(err error) {
	execErrorEvent := &model.ExecErrorEvent{ExecId: watcher.exec.ID, Stack: err.Error()}

	events.ExecEventBus.Pub(execErrorEvent)
}
