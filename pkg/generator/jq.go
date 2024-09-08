package generator

import (
	"bytes"
	"context"
	"errors"
	"fmt"
	"os"
	"os/exec"

	"github.com/chancez/yamlforge/pkg/config"
	"github.com/chancez/yamlforge/pkg/reference"
)

var _ Generator = (*JQ)(nil)

type JQ struct {
	dir      string
	cfg      config.JQGenerator
	refStore *reference.Store
}

func NewJQ(dir string, cfg config.JQGenerator, refStore *reference.Store) *JQ {
	return &JQ{
		dir:      dir,
		cfg:      cfg,
		refStore: refStore,
	}
}

func (jq *JQ) Generate(context.Context) ([]byte, error) {
	var jqArgs []string

	switch {
	case jq.cfg.Expr != "" && jq.cfg.ExprFile != "":
		return nil, errors.New("cannot specify both expr and exprFile")
	case jq.cfg.Expr != "":
		jqArgs = append(jqArgs, jq.cfg.Expr)
	case jq.cfg.ExprFile != "":
		jqArgs = append(jqArgs, "--from-file", jq.cfg.ExprFile)
	default:
		return nil, errors.New("expression is required")
	}

	if jq.cfg.Slurp {
		jqArgs = append(jqArgs, "--slurp")
	}

	jqArgs = append(jqArgs,
		"--raw-output",
		"--monochrome-output",
	)

	refVal, err := jq.refStore.GetReference(jq.dir, jq.cfg.Input)
	if err != nil {
		return nil, fmt.Errorf("error getting reference: %w", err)
	}

	var buf bytes.Buffer
	cmd := exec.Command("jq", jqArgs...)
	cmd.Stdin = bytes.NewBuffer(refVal)
	cmd.Stderr = os.Stderr
	cmd.Stdout = &buf
	err = cmd.Run()
	if err != nil {
		return nil, err
	}
	return buf.Bytes(), nil
}
