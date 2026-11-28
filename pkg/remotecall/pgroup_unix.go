//go:build unix

package remotecall

import (
	"errors"
	"os/exec"

	"golang.org/x/sys/unix"
)

// isETXTBSY reports whether err is a transient Unix ETXTBSY ("text file busy")
// from execve racing with another goroutine's fork while a script file was open
// for writing.
func isETXTBSY(err error) bool {
	return errors.Is(err, unix.ETXTBSY)
}

// setProcessGroup starts cmd in a new process group and makes a cancel
// kill the whole group. With detach, cmd starts a new session instead,
// which is also a new process group, with no controlling terminal.
func setProcessGroup(cmd *exec.Cmd, detach bool) {
	if detach {
		cmd.SysProcAttr = &unix.SysProcAttr{Setsid: true}
	} else {
		cmd.SysProcAttr = &unix.SysProcAttr{Setpgid: true}
	}
	cmd.Cancel = func() error {
		return unix.Kill(-cmd.Process.Pid, unix.SIGKILL)
	}
}
