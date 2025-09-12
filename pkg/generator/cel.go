package generator

import (
	"context"
	"fmt"
	"reflect"

	"github.com/chancez/yamlforge/pkg/config"
	"github.com/google/cel-go/cel"
)

var _ Generator = (*CEL)(nil)

var (
	goBoolType = reflect.TypeOf(false)
)

type CEL struct {
	dir      string
	cfg      config.CELGenerator
	refStore *Store
}

func NewCEL(dir string, cfg config.CELGenerator, refStore *Store) *CEL {
	return &CEL{
		dir:      dir,
		cfg:      cfg,
		refStore: refStore,
	}
}

func (c *CEL) Generate(ctx context.Context) (*Result, error) {
	expr, err := c.refStore.GetStringValue(c.dir, c.cfg.Expr)
	if err != nil {
		return nil, fmt.Errorf("error getting expression: %w", err)
	}
	filter, err := c.refStore.GetBoolValue(c.dir, c.cfg.Filter)
	if err != nil {
		return nil, fmt.Errorf("error getting filter: %w", err)
	}

	prg, err := c.newProgram(expr, filter)
	if err != nil {
		return nil, fmt.Errorf("error creating CEL program: %w", err)
	}

	var results []any
	if c.cfg.Input != nil {
		vals, err := c.refStore.GetParsedValues(c.dir, *c.cfg.Input)
		if err != nil {
			return nil, fmt.Errorf("error getting input: %w", err)
		}
		invertFilter, err := c.refStore.GetBoolValue(c.dir, c.cfg.InvertFilter)
		if err != nil {
			return nil, fmt.Errorf("error getting invertFilter: %w", err)
		}
		collect, err := c.refStore.GetBoolValue(c.dir, c.cfg.Collect)
		if err != nil {
			return nil, fmt.Errorf("error getting collect: %w", err)
		}

		if collect {
			var allVals []any
			for val, err := range vals {
				if err != nil {
					return nil, fmt.Errorf("error while processing input: %w", err)
				}
				allVals = append(allVals, val.Parsed())
			}
			vals = func(yield func(ParsedValue, error) bool) {
				yield(ParsedValue{parsed: allVals}, nil)
			}
		}

		for val, err := range vals {
			if err != nil {
				return nil, fmt.Errorf("error while processing input: %w", err)
			}
			result, skip, err := c.evalProgram(ctx, prg, val.Parsed(), filter, invertFilter)
			if err != nil {
				return nil, err
			}
			if filter && skip {
				continue
			}
			results = append(results, result)
		}
	} else {
		// No input, just evaluate once with no variables
		out, _, err := prg.ContextEval(ctx, map[string]any{})
		if err != nil {
			return nil, fmt.Errorf("error evaluating CEL program: %s", err)
		}
		result := out.Value()
		results = append(results, result)
	}

	var output any
	if len(results) == 1 {
		output = results[0]
	} else {
		output = results
	}
	return &Result{Output: output}, nil
}

func (c *CEL) newProgram(expr string, filter bool) (cel.Program, error) {
	env, err := cel.NewEnv(
		cel.Variable("val", cel.DynType),
		cel.OptionalTypes(),
	)
	if err != nil {
		return nil, fmt.Errorf("error creating CEL environment: %w", err)
	}

	ast, iss := env.Compile(expr)
	if iss != nil && iss.Err() != nil {
		return nil, fmt.Errorf("CEL compile error: %s", iss.Err())
	}
	checked, iss := env.Check(ast)
	if iss != nil && iss.Err() != nil {
		return nil, fmt.Errorf("CEL type-check error: %s", iss.Err())
	}

	if filter && checked.OutputType() != cel.BoolType {
		return nil, fmt.Errorf("CEL expression has invalid result type: got %q, wanted %q", checked.OutputType(), cel.BoolType)
	}

	prg, err := env.Program(ast)
	if err != nil {
		return nil, fmt.Errorf("CEL program construction error: %s", err)
	}
	return prg, nil
}

func (c *CEL) evalProgram(ctx context.Context, prg cel.Program, inputVal any, filter, invertFilter bool) (outVal any, matched bool, error error) {
	out, _, err := prg.ContextEval(ctx, map[string]any{
		"val": inputVal,
	})
	if err != nil {
		return nil, false, fmt.Errorf("error evaluating CEL program: %s", err)
	}

	outVal = out.Value()
	if filter {
		outVal = inputVal
		v, err := out.ConvertToNative(goBoolType)
		if err != nil {
			return nil, false, fmt.Errorf("error converting CEL result to boolean: %s", err)
		}
		matched = v.(bool)

		if matched == invertFilter {
			return nil, true, nil
		}
	}
	return outVal, false, nil

}
