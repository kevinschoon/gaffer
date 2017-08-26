package subsystem

import "github.com/mesanine/gaffer/ginit/mount"

func ProcFS() Subsystem {
	return Subsystem{
		Mounts: []mount.MountArgs{mount.ProcFS()},
	}
}
