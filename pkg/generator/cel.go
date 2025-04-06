package generator

import (
	"bytes"
	"context"
	"encoding/json"
	"fmt"
	"io"
	"reflect"

	"github.com/chancez/yamlforge/pkg/config"
	"github.com/chancez/yamlforge/pkg/reference"
	"github.com/google/cel-go/cel"
)

var _ Generator = (*CEL)(nil)

var (
	goBoolType = reflect.TypeOf(false)
)

type CEL struct {
	dir      string
	cfg      config.CELGenerator
	refStore *reference.Store
}

func NewCEL(dir string, cfg config.CELGenerator, refStore *reference.Store) *CEL {
	return &CEL{
		dir:      dir,
		cfg:      cfg,
		refStore: refStore,
	}
}

type encoder interface {
	Encode(any) error
}

func (c *CEL) Generate(ctx context.Context) ([]byte, error) {
	vals, err := c.refStore.GetParsedValues(c.dir, c.cfg.Input)
	if err != nil {
		return nil, err
	}

	var buf bytes.Buffer
	var enc encoder
	switch c.cfg.Format {
	case "yaml", "":
		enc = config.NewYAMLEncoder(&buf)
	case "json":
		enc = json.NewEncoder(&buf)
	default:
		return nil, fmt.Errorf("invalid output format specified: %q", c.cfg.Format)
	}

	env, err := cel.NewEnv(
		cel.Variable("val", cel.DynType),
	)
	if err != nil {
		return nil, fmt.Errorf("error creating CEL environment: %w", err)
	}

	ast, iss := env.Compile(c.cfg.Expr)
	if iss != nil && iss.Err() != nil {
		return nil, fmt.Errorf("CEL compile error: %s", iss.Err())
	}
	checked, iss := env.Check(ast)
	if iss != nil && iss.Err() != nil {
		return nil, fmt.Errorf("CEL type-check error: %s", iss.Err())
	}

	if c.cfg.Filter && checked.OutputType() != cel.BoolType {
		return nil, fmt.Errorf("CEL expression has invalid result type: got %q, wanted %q", checked.OutputType(), cel.BoolType)
	}

	prg, err := env.Program(ast)
	if err != nil {
		return nil, fmt.Errorf("CEL program construction error: %s", err)
	}

	for val, err := range vals {
		if err != nil {
			return nil, fmt.Errorf("error while processing input: %w", err)
		}
		out, _, err := prg.ContextEval(ctx, map[string]any{
			"val": val,
		})
		if err != nil {
			return nil, fmt.Errorf("error evaluating CEL program: %s", err)
		}

		toEncode := out.Value()
		if c.cfg.Filter {
			toEncode = val
			v, err := out.ConvertToNative(goBoolType)
			if err != nil {
				return nil, fmt.Errorf("error converting CEL result to boolean: %s", err)
			}
			matched := v.(bool)

			if matched == c.cfg.InvertFilter {
				continue
			}
		}

		err = enc.Encode(toEncode)
		if err != nil {
			return nil, fmt.Errorf("error encoding result at %s: %w", c.cfg.Input.Format, err)
		}
	}

	if closer, ok := enc.(io.Closer); ok {
		if err := closer.Close(); err != nil {
			return nil, fmt.Errorf("error closing encoder: %w", err)
		}
	}

	return buf.Bytes(), nil
}
