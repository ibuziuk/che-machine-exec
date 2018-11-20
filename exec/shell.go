package exec

import "github.com/ws-skeleton/che-machine-exec/api/model"

// ShellDetector uses to get information about preferable exec shell
// defined inside container for current user.
// Information about preferable shell we get from /etc/passwd file.
type ShellDetector interface {
	// detect preferable shell inside container for current user
	detectShell(identifier model.MachineIdentifier) (string, error)
}