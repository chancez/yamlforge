package generator

import (
	"bytes"
	"context"
	"fmt"
	"os"
	"os/exec"

	"github.com/chancez/yamlforge/pkg/config"
	"github.com/chancez/yamlforge/pkg/reference"
)

var _ Generator = (*Exec)(nil)

type Exec struct {
	dir      string
	cfg      config.ExecGenerator
	refStore *reference.Store
}

// TODO: Use directory as PWD
func NewExec(dir string, cfg config.ExecGenerator, refStore *reference.Store) *Exec {
	return &Exec{
		dir:      dir,
		cfg:      cfg,
		refStore: refStore,
	}
}

func (e *Exec) Generate(context.Context) ([]byte, error) {
	var env []string
	for _, envVar := range e.cfg.Env {
		data, err := e.refStore.GetReference(e.dir, envVar.Value)
		if err != nil {
			return nil, fmt.Errorf("error getting reference: %w", err)
		}
		env = append(env, fmt.Sprintf("%s=%s", envVar.Name, string(data)))
	}
	var buf bytes.Buffer
	cmd := exec.Command(e.cfg.Command, e.cfg.Args...)
	cmd.Dir = e.dir
	cmd.Stderr = os.Stderr
	cmd.Stdout = &buf
	cmd.Env = env
	err := cmd.Run()
	if err != nil {
		return nil, err
	}
	return buf.Bytes(), nil
}
