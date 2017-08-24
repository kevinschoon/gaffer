package ginit

import (
	"fmt"
	"golang.org/x/sys/unix"
)

type Type int

const (
	_ Type = iota
	MOUNT
	UNMOUNT
)

type Option func(Syscall) Syscall

type Syscall struct {
	Type    Type
	mount   mount
	unmount unmount
	Before  func() error
	After   func() error
}

type mount struct {
	Source string
	Target string
	FSType string
	Flags  uintptr
	Data   string
}

type unmount struct {
	Target string
	Flags  int
}

func (s Syscall) CallWith(opts ...Option) error {
	for _, opt := range opts {
		s = opt(s)
	}
	return s.Call()
}

func (s Syscall) Call() (err error) {
	if s.Before != nil {
		err = s.Before()
		if err != nil {
			return err
		}
	}
	switch s.Type {
	case MOUNT:
		err = unix.Mount(
			s.mount.Source,
			s.mount.Target,
			s.mount.FSType,
			s.mount.Flags,
			s.mount.Data,
		)
		if err != nil {
			return err
		}
	case UNMOUNT:
		err = unix.Unmount(
			s.unmount.Target,
			s.unmount.Flags,
		)
		if err != nil {
			return err
		}
	default:
		return fmt.Errorf("unknown syscall type")
	}
	if s.After != nil {
		return s.After()
	}
	return nil
}

func copySyscall(s Syscall) Syscall {
	return Syscall{
		Type:    s.Type,
		mount:   s.mount,
		unmount: s.unmount,
		Before:  s.Before,
		After:   s.After,
	}
}
