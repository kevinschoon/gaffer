package runc

// Code is borrowed from Linuxkit's runc and containerd packages
// https://github.com/linuxkit/linuxkit/blob/5ea2eaead11d4882e4dc889b4a3d7cd49bd0f9f3/pkg/runc/cmd/onboot/prepare.go

import (
	"fmt"
	"os"
	"path/filepath"
	"syscall"
)

func prepareRO(path string) error {
	// make rootfs a mount point, as runc doesn't like it much otherwise
	rootfs := filepath.Join(path, "rootfs")
	if err := syscall.Mount(rootfs, rootfs, "", syscall.MS_BIND, ""); err != nil {
		return err
	}
	return nil
}

func prepareRW(path string) error {
	// mount a tmpfs on tmp for upper and workdirs
	// make it private as nothing else should be using this
	tmp := filepath.Join(path, "tmp")
	if err := syscall.Mount("tmpfs", tmp, "tmpfs", 0, "size=10%"); err != nil {
		return err
	}
	// make it private as nothing else should be using this
	if err := syscall.Mount("", tmp, "", syscall.MS_REMOUNT|syscall.MS_PRIVATE, ""); err != nil {
		return err
	}
	upper := filepath.Join(tmp, "upper")
	// make the mount points
	if err := os.Mkdir(upper, 0744); err != nil {
		return err
	}
	work := filepath.Join(tmp, "work")
	if err := os.Mkdir(work, 0744); err != nil {
		return err
	}
	lower := filepath.Join(path, "lower")
	rootfs := filepath.Join(path, "rootfs")
	opt := fmt.Sprintf("lowerdir=%s,upperdir=%s,workdir=%s", lower, upper, work)
	if err := syscall.Mount("overlay", rootfs, "overlay", 0, opt); err != nil {
		return err
	}
	return nil
}

func cleanupRO(path string) error {
	// remove the bind mount
	rootfs := filepath.Join(path, "rootfs")
	return syscall.Unmount(rootfs, 0)
}

func cleanupRW(path string) (err error) {
	// remove the overlay mount
	rootfs := filepath.Join(path, "rootfs")
	err = os.RemoveAll(rootfs)
	if err != nil {
		return err
	}
	err = syscall.Unmount(rootfs, 0)
	if err != nil {
		return err
	}
	// remove the tmpfs
	tmp := filepath.Join(path, "tmp")
	err = os.RemoveAll(tmp)
	if err != nil {
		return err
	}
	return syscall.Unmount(tmp, 0)
}
