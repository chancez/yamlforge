package generator

import (
	"bytes"
	"context"
	"os"
	"os/exec"

	"github.com/chancez/yamlforge/pkg/config"
	"github.com/chancez/yamlforge/pkg/reference"
)

var _ Generator = (*Exec)(nil)

func init() {
	Register("exec", config.ExecGenerator{}, func(dir string, cfg any, refStore *reference.Store) Generator {
		return NewExec(cfg.(config.ExecGenerator))
	})
}

type Exec struct {
	cfg config.ExecGenerator
}

// TODO: Use directory as PWD
func NewExec(cfg config.ExecGenerator) *Exec {
	return &Exec{
		cfg: cfg,
	}
}

func (e *Exec) Generate(context.Context) ([]byte, error) {
	var buf bytes.Buffer
	cmd := exec.Command(e.cfg.Command, e.cfg.Args...)
	cmd.Stderr = os.Stderr
	cmd.Stdout = &buf
	err := cmd.Run()
	if err != nil {
		return nil, err
	}
	return buf.Bytes(), nil
}
