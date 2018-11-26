package shell

// ContainerShellDetector uses to get information about preferable exec shell
// defined inside container for current active user.
// Information about preferable shell we get from /etc/passwd file.
type ContainerShellDetector interface {
	// detect preferable shell inside container for current user
	DetectShell() (shell string, err error)
}
