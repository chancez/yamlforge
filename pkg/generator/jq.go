package generator

import (
	"bytes"
	"context"
	"fmt"
	"os"
	"os/exec"
	"strings"

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
	expr, err := jq.refStore.GetStringValue(jq.dir, jq.cfg.Expr)
	if err != nil {
		return nil, fmt.Errorf("error getting expression: %w", err)
	}
	jqArgs := []string{
		expr,
	}

	slurp, err := jq.refStore.GetBoolValue(jq.dir, jq.cfg.Slurp)
	if err != nil {
		return nil, fmt.Errorf("error getting slurp: %w", err)
	}

	if slurp {
		jqArgs = append(jqArgs, "--slurp")
	}

	jqArgs = append(jqArgs,
		"--raw-output",
		"--monochrome-output",
	)

	data, err := jq.refStore.GetStringValue(jq.dir, jq.cfg.Input)
	if err != nil {
		return nil, fmt.Errorf("error getting value: %w", err)
	}

	var buf bytes.Buffer
	cmd := exec.Command("jq", jqArgs...)
	cmd.Stdin = strings.NewReader(data)
	cmd.Stderr = os.Stderr
	cmd.Stdout = &buf
	err = cmd.Run()
	if err != nil {
		return nil, err
	}
	return buf.Bytes(), nil
}
