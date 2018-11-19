package exec

import (
	"errors"
	"github.com/stretchr/testify/mock"
	"github.com/ws-skeleton/che-machine-exec/api/model"
	"github.com/ws-skeleton/che-machine-exec/mocks"
	"testing"
	"time"
)

const Exec1ID = 0

func TestShouldCleanUpExecOnExit(t *testing.T) {
	machineExec := &model.MachineExec{ID: Exec1ID, ErrorChan: make(chan error), ExitChan: make(chan bool)}
	execManagerMock := &mocks.ExecManager{}
	execEventBusMock := &mocks.ExecEventBus{}

	execManagerMock.On("Remove", Exec1ID).Return()
	execEventBusMock.On("Pub", mock.Anything).Return()

	healthWatcher := NewHealthWatcher(machineExec, execEventBusMock, execManagerMock)
	healthWatcher.CleanUpOnExitOrError()

	machineExec.ExitChan <- true
	time.Sleep(1000 * time.Millisecond)

	execManagerMock.AssertExpectations(t)
	execEventBusMock.AssertExpectations(t)
}

func TestShouldCleanUpExecOnError(t *testing.T) {
	machineExec := &model.MachineExec{ID: Exec1ID, ErrorChan: make(chan error), ExitChan: make(chan bool)}
	execManagerMock := &mocks.ExecManager{}
	execEventBusMock := &mocks.ExecEventBus{}

	execManagerMock.On("Remove", Exec1ID).Return()
	execEventBusMock.On("Pub", mock.Anything).Return()

	healthWatcher := NewHealthWatcher(machineExec, execEventBusMock, execManagerMock)
	healthWatcher.CleanUpOnExitOrError()

	machineExec.ErrorChan <- errors.New("unable to create exec")
	time.Sleep(1000 * time.Millisecond)

	execManagerMock.AssertExpectations(t)
	execEventBusMock.AssertExpectations(t)
}
