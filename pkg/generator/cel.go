package generator

import (
	"bytes"
	"context"
	"encoding/json"
	"errors"
	"fmt"
	"io"
	"reflect"

	"github.com/chancez/yamlforge/pkg/config"
	"github.com/chancez/yamlforge/pkg/reference"
	"github.com/google/cel-go/cel"
	"gopkg.in/yaml.v3"
)

var _ Generator = (*CELFilter)(nil)

var (
	goBoolType = reflect.TypeOf(false)
)

func init() {
	Register("celfilter", config.CELFilterGenerator{}, func(dir string, cfg any, refStore *reference.Store) Generator {
		return NewCELFilter(dir, cfg.(config.CELFilterGenerator), refStore)
	})
}

type CELFilter struct {
	dir      string
	cfg      config.CELFilterGenerator
	refStore *reference.Store
}

func NewCELFilter(dir string, cfg config.CELFilterGenerator, refStore *reference.Store) *CELFilter {
	return &CELFilter{
		dir:      dir,
		cfg:      cfg,
		refStore: refStore,
	}
}

type decoder interface {
	Decode(any) error
}

type encoder interface {
	Encode(any) error
}

func (c *CELFilter) Generate(ctx context.Context) ([]byte, error) {
	ref, err := c.refStore.GetReference(c.dir, c.cfg.Input.Value)
	if err != nil {
		return nil, fmt.Errorf("error getting reference: %w", err)
	}

	var buf bytes.Buffer
	var dec decoder
	var enc encoder
	switch c.cfg.Input.Format {
	case "yaml":
		dec = yaml.NewDecoder(bytes.NewBuffer(ref))
		enc = yaml.NewEncoder(&buf)
	case "json":
		dec = json.NewDecoder(bytes.NewBuffer(ref))
		enc = json.NewEncoder(&buf)
	case "":
		return nil, errors.New("input.format is required")
	default:
		return nil, fmt.Errorf("invalid format specified: %q", c.cfg.Input.Format)
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
	if checked.OutputType() != cel.BoolType {
		return nil, fmt.Errorf("CEL expression has invalid result type: got %q, wanted %q", checked.OutputType(), cel.BoolType)
	}

	prg, err := env.Program(ast)
	if err != nil {
		return nil, fmt.Errorf("CEL program construction error: %s", err)
	}

	for {
		var val any
		err := dec.Decode(&val)
		if err == io.EOF {
			break
		}
		if err != nil {
			return nil, fmt.Errorf("error decoding input as %s: %w", c.cfg.Input.Format, err)
		}

		out, _, err := prg.ContextEval(ctx, map[string]any{
			"val": val,
		})
		if err != nil {
			return nil, fmt.Errorf("error evaluating CEL program: %s", err)
		}

		v, err := out.ConvertToNative(goBoolType)
		if err != nil {
			return nil, fmt.Errorf("error converting CEL result to boolean: %s", err)
		}
		matched := v.(bool)

		if matched == c.cfg.Discard {
			continue
		}

		err = enc.Encode(val)
		if err != nil {
			return nil, fmt.Errorf("error encoding result at %s: %w", c.cfg.Input.Format, err)
		}
	}

	return buf.Bytes(), nil
}