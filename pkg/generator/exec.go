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
	dir string
	cfg config.ExecGenerator
}

// TODO: Use directory as PWD
func NewExec(dir string, cfg config.ExecGenerator) *Exec {
	return &Exec{
		dir: dir,
		cfg: cfg,
	}
}

func (e *Exec) Generate(context.Context) ([]byte, error) {
	var buf bytes.Buffer
	cmd := exec.Command(e.cfg.Command, e.cfg.Args...)
	cmd.Dir = e.dir
	cmd.Stderr = os.Stderr
	cmd.Stdout = &buf
	err := cmd.Run()
	if err != nil {
		return nil, err
	}
	return buf.Bytes(), nil
}
