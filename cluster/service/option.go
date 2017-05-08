package service

import (
	"fmt"
	"io/ioutil"
	"os"
	"strings"
)

type Env struct {
	Name  string `json:"name"`
	Value string `json:"value"`
}

func (e Env) String() string { return fmt.Sprintf("%s=%s", e.Name, e.Value) }

type File struct {
	Path    string      `json:"path"`
	Content []string    `json:"content"`
	Mode    os.FileMode `json:"mode"`
	Dir     bool        `json:"dir"`
}

func (f *File) Write(cwd string) error {
	err := os.Chdir(cwd)
	if err != nil {
		return err
	}
	if f.Mode == 0 {
		if !f.Dir {
			f.Mode = os.FileMode(0644)
		}
	}
	split := strings.Split(f.Path, "/")
	if len(split) == 1 {
		if f.Dir {
			err = os.Mkdir(f.Path, 0755)
			if err != nil {
				return err
			}
			return nil
		}
		err = ioutil.WriteFile(f.Path, []byte(strings.Join(f.Content, "\n")), f.Mode)
		if err != nil {
			return err
		}
		return nil
	}
	err = os.MkdirAll(fmt.Sprintf("./%s", strings.Join(split[:len(split)-1], "/")), 0755)
	if err != nil {
		return err
	}
	return ioutil.WriteFile(f.Path, []byte(strings.Join(f.Content, "\n")), f.Mode)
}
