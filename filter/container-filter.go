package filter

import "github.com/ws-skeleton/che-machine-exec/api/model"

// Container filter to find container information if it's possible by unique
// machine identifier
type ContainerFilter interface {
	FindContainerInfo(identifier *model.MachineIdentifier) (containerInfo map[string]string, err error)
}
