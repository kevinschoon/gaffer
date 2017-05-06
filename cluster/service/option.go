package service

import (
	"fmt"
	"os"
)

type Env struct {
	Name  string `json:"name"`
	Value string `json:"value"`
}

func (e Env) String() string { return fmt.Sprintf("%s=%s", e.Name, e.Value) }

type File struct {
	Path    string      `json:"path"`
	Content []byte      `json:"content"`
	Mode    os.FileMode `json:"mode"`
}
