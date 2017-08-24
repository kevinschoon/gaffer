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
			Target: path,
			Source: path,
			Flags:  flags,
		},
	}
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
			FSType: "overlay",
			Target: target,
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

/*

type TempFS struct {
	Path string
	Size int
}

func (t TempFS) Syscall() Syscall {
	return Syscall{
		Type: MOUNT,
		Mount: Mount{
			Source: "tmpfs",
			Target: t.Path,
			FSType: "tmpfs",
			Flags:  0,
			Data:   fmt.Sprintf("size=%d", t.Path),
		},
	}
}

type OverlayFS struct {
	Path  string
	Upper string
	Lower string
	Work  string
}

func (o OverlayFS) Syscall() Syscall {
	return Syscall{
		Type: MOUNT,
		Mount: Mount{
			Source: "overlay",
			Target: o.Path,
			FSType: "overlay",
			Flags:  0,
			Data:   fmt.Sprintf("lowerdir=%s,upperdir=%s,workdir=%s", o.Lower, o.Upper, o.Work),
		},
	}
}
func (o OverlayFS) Mount() (err error) {
	err = os.MkdirAll(o.Upper, MountDirPerms)
	if err != nil {
		return err
	}
	err = os.MkdirAll(o.Work, MountDirPerms)
	if err != nil {
		return err
	}
	return syscall.Mount(
		"overlay",
		o.Path,
		"overlay",
		0,
		fmt.Sprintf("lowerdir=%s,upperdir=%s,workdir=%s", o.Lower, o.Upper, o.Work),
	)
}
*/
