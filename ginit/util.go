package ginit

import (
	"golang.org/x/sys/unix"
	"os"
	"syscall"
)

func IsRoot() bool { return os.Getuid() == 0 }

// Exec does an execve with provided arguments
// it appends the executable to the front of
// the arguments and copies the existing environment.
func Exec(exe string, args ...string) error {
	newArgs := []string{exe}
	for _, arg := range args {
		newArgs = append(newArgs, arg)
	}
	return syscall.Exec(exe, newArgs, os.Environ())
}

// Check if a filesystem is memory based i.e. tempfs or ramfs
func IsMemFS(path string) (bool, error) {
	var stat unix.Statfs_t
	err := unix.Statfs(path, &stat)
	if err != nil {
		return false, err
	}
	if stat.Type == ramfsMagic || stat.Type == tmpfsMagic {
		return true, nil
	}
	return false, nil
}

// Mkdev is used to build the value of linux devices (in /dev/) which specifies major
// and minor number of the newly created device special file.
// Linux device nodes are a bit weird due to backwards compat with 16 bit device nodes.
// They are, from low to high: the lower 8 bits of the minor, then 12 bits of the major,
// then the top 12 bits of the minor.
// From https://github.com/moby/moby/blob/master/pkg/system/mknod.go
func Mkdev(major int64, minor int64) uint32 {
	return uint32(((minor & 0xfff00) << 12) | ((major & 0xfff) << 8) | (minor & 0xff))
}
