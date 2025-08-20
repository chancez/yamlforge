package generator

import (
	"bytes"
	"context"
	"fmt"
	"os"
	"os/exec"

	"github.com/chancez/yamlforge/pkg/config"
)

var _ Generator = (*Exec)(nil)

type Exec struct {
	dir      string
	cfg      config.ExecGenerator
	refStore *Store
}

// TODO: Use directory as PWD
func NewExec(dir string, cfg config.ExecGenerator, refStore *Store) *Exec {
	return &Exec{
		dir:      dir,
		cfg:      cfg,
		refStore: refStore,
	}
}

func (e *Exec) Generate(context.Context) ([]byte, error) {
	var env []string
	for _, envVar := range e.cfg.Env {
		data, err := e.refStore.GetValueBytes(e.dir, envVar.Value)
		if err != nil {
			return nil, fmt.Errorf("error getting value: %w", err)
		}
		env = append(env, fmt.Sprintf("%s=%s", envVar.Name, string(data)))
	}

	command, err := e.refStore.GetStringValue(e.dir, e.cfg.Command)
	if err != nil {
		return nil, err
	}
	args, err := e.refStore.GetStringValueList(e.dir, e.cfg.Args)
	if err != nil {
		return nil, err
	}

	var buf bytes.Buffer
	cmd := exec.Command(command, args...)
	cmd.Dir = e.dir
	cmd.Stderr = os.Stderr
	cmd.Stdout = &buf
	cmd.Env = append(os.Environ(), env...)
	err = cmd.Run()
	if err != nil {
		return nil, err
	}
	return buf.Bytes(), nil
}
