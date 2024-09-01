package generator

import (
	"bytes"
	"context"
	"os"
	"os/exec"

	"github.com/chancez/yamlforge/pkg/config"
)

var _ Generator = (*Exec)(nil)

type Exec struct {
	cfg config.ExecGenerator
}

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
