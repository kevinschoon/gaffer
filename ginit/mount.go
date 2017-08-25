package ginit

import (
	"fmt"
	"golang.org/x/sys/unix"
	"os"
	"path/filepath"
)

// Bind performs a bind mount
func Bind(path string, readOnly bool) Syscall {
	var flags uintptr
	if readOnly {
		flags = unix.MS_BIND | unix.MS_RDONLY
	} else {
		flags = unix.MS_BIND
	}
	return Syscall{
		Type: MOUNT,
		mount: mount{
			Source: path,
			Target: path,
			Flags:  flags,
		},
	}
}

// TmpFS performs a tmpfs mount at the given path
// percentage must be between 0 and 100 or we will
// panic. If it is zero we do not specify any flags.
func TmpFS(path string, percentage int) Syscall {
	if percentage < 0 || percentage > 100 {
		panic("invalid tempfs percentage")
	}
	var data string
	if percentage > 0 {
		data = fmt.Sprintf("%d", percentage)
	}
	return Syscall{
		Type: MOUNT,
		Before: func() error {
			return os.MkdirAll(path, 0755)
		},
		mount: mount{
			Source: "rootfs", // TODO: Unsure if this has significance with tempfs
			Target: path,
			FSType: "tmpfs",
			Data:   data,
		}}
}

// Overlay mounts the provided path as
// overlayfs creating "work" and "upper"
// directories in the parent directory.
func Overlay(lower, target string) Syscall {
	upper := filepath.Join(filepath.Dir(lower), "upper")
	work := filepath.Join(filepath.Dir(lower), "work")
	return Syscall{
		Type: MOUNT,
		Before: func() (err error) {
			err = os.MkdirAll(upper, 0755)
			if err != nil {
				return err
			}
			err = os.MkdirAll(work, 0755)
			if err != nil {
				return err
			}
			return nil
		},
		mount: mount{
			Source: "overlay",
			Target: target,
			FSType: "overlay",
			Flags:  0,
			Data:   fmt.Sprintf("lowerdir=%s,upperdir=%s,workdir=%s", lower, upper, work),
		},
	}
}

func Unmount(path string) Syscall {
	return Syscall{
		Type: UNMOUNT,
		unmount: unmount{
			Target: path,
		},
	}
}
