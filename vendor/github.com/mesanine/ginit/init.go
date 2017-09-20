package ginit

import (
	"golang.org/x/sys/unix"
	"os"
	"os/signal"
	"time"
)

// Handler responds to an os.Signal
// in some meaningful way.
type Handler interface {
	Handle(os.Signal) error
}

// Init launches an "origin" signal handler
// propagating any signal it receives to it's
// children. If any child returns an error
// it will be passed back to the caller. Assuming
// the calling program is running as PID 1, the
// caller can print an informative message and
// then exit causing a kernel panic. If we receive
// SIGUSR1, SIGUSR2, or SIGTERM we trigger a halt,
// poweroff, or restart command with a syscall to
// the kernel. If we recieve SIGINT (ctrl-c) we
// exit cleanly, (for running in a Docker container).
// Resources:
// https://github.com/mirror/busybox/tree/master/init
// https://github.com/torvalds/linux/blob/master/kernel/reboot.c
func Init(children ...Handler) error {
	var (
		cmd int
		err error
	)
	sigCh := make(chan os.Signal, 10)
	signal.Reset()
	signal.Notify(sigCh)
loop:
	for {
		sig := <-sigCh
		for _, child := range children {
			err := child.Handle(sig)
			if err != nil {
				return err
			}
		}
		switch sig {
		// ctrl-c
		case unix.SIGINT:
			return nil
		// Halt
		case unix.SIGUSR1:
			cmd = unix.LINUX_REBOOT_CMD_HALT
			break loop
		// Poweroff
		case unix.SIGUSR2:
			cmd = unix.LINUX_REBOOT_CMD_POWER_OFF
			break loop
		// Reboot
		case unix.SIGTERM:
			cmd = unix.LINUX_REBOOT_CMD_RESTART
			break loop
		}
	}
	// SIGTERM all processes except pid 1
	err = unix.Kill(-1, unix.SIGTERM)
	if err != nil {
		return err
	}
	time.Sleep(1)
	// SIGKILL all processes except pid 1
	err = unix.Kill(-1, unix.SIGKILL)
	if err != nil {
		return err
	}
	time.Sleep(1)
	// Call final sync
	unix.Sync()
	// Bye Bye!
	return unix.Reboot(cmd)
}

// Terminal checks if a signal is "terminal"
// and will cause the OS to shutdown.
func Terminal(sig os.Signal) bool {
	switch sig {
	case unix.SIGINT:
		return true
	case unix.SIGUSR1:
		return true
	case unix.SIGUSR2:
		return true
	case unix.SIGTERM:
		return true
	}
	return false
}
