package generator

import (
	"bytes"
	"context"
	"encoding/json"
	"errors"
	"fmt"
	"io"
	"iter"
	"os"
	"path"

	"github.com/chancez/yamlforge/pkg/config"
)

type Store struct {
	// map from an generator.Name to it's results
	references map[string]*Result
	// map a variable name to it's value
	vars map[string]any
}

func NewStore(vars map[string]any) *Store {
	return &Store{
		references: make(map[string]*Result),
		vars:       vars,
	}
}

func (store *Store) AddReference(name string, result *Result) error {
	if _, exists := store.references[name]; exists {
		return fmt.Errorf("reference %q already exists", name)
	}
	store.references[name] = result
	return nil
}

func (store *Store) GetValueBytes(dir string, ref config.Value) ([]byte, error) {
	ret, err := store.GetValue(dir, ref)
	if err != nil {
		return nil, err
	}
	return ConvertToBytes(ret)
}

func (store *Store) GetAnyValue(dir string, val config.AnyOrValue) (*Result, error) {
	if val.Any != nil {
		return &Result{Output: *val.Any}, nil
	}
	if val.Value != nil {
		res, err := store.GetValue(dir, *val.Value)
		if err != nil {
			return nil, err
		}
		return res, nil
	}
	return nil, nil
}

func (store *Store) GetStringValue(dir string, val config.StringOrValue) (string, error) {
	if val.String != nil {
		return *val.String, nil
	}
	if val.Value != nil {
		v, err := store.GetValue(dir, *val.Value)
		if err != nil {
			return "", err
		}
		if s, ok := v.Output.(string); ok {
			return s, nil
		}
		data, err := ConvertToBytes(v)
		if err != nil {
			return "", err
		}
		return string(data), nil
	}
	return "", nil
}

func (store *Store) GetStringValueList(dir string, vals []config.StringOrValue) ([]string, error) {
	var ret []string
	if len(vals) != 0 {
		for _, val := range vals {
			sv, err := store.GetStringValue(dir, val)
			if err != nil {
				return nil, err
			}
			ret = append(ret, sv)
		}
	}
	return ret, nil
}

func (store *Store) GetBoolValue(dir string, val config.BoolOrValue) (bool, error) {
	if val.Bool != nil {
		return *val.Bool, nil
	}
	if val.Value != nil {
		v, err := store.GetValue(dir, *val.Value)
		if err != nil {
			return false, err
		}
		if b, ok := v.Output.(bool); ok {
			return b, nil
		}
		data, err := ConvertToBytes(v)
		if err != nil {
			return false, err
		}
		var b bool
		err = config.DecodeYAML(data, &b)
		if err != nil {
			return false, err
		}
		return b, nil
	}
	return false, nil
}

func (store *Store) GetMapValue(dir string, val config.MapOrValue) (map[string]any, error) {
	if val.Map != nil {
		return val.Map, nil
	}
	if val.Value != nil {
		v, err := store.GetValue(dir, *val.Value)
		if err != nil {
			return nil, err
		}
		if mapVal, ok := v.Output.(map[string]any); ok {
			return mapVal, nil
		}

		data, err := ConvertToBytes(v)
		if err != nil {
			return nil, err
		}
		var m map[string]any
		err = config.DecodeYAML(data, &m)
		if err != nil {
			return nil, err
		}
		return m, nil
	}
	return nil, nil
}

func (store *Store) GetValue(dir string, ref config.Value) (*Result, error) {
	if ref.Format != "" {
		items, err := store.GetParsedValues(dir, ref)
		if err != nil {
			return nil, err
		}
		var res []any
		for item, err := range items {
			if err != nil {
				return nil, err
			}
			res = append(res, item.Parsed())
		}
		if len(res) == 1 {
			return &Result{Output: res[0]}, nil
		}
		return &Result{Output: res}, nil
	}
	return store.getValue(dir, ref)
}

func (store *Store) getValue(dir string, ref config.Value) (*Result, error) {
	switch {
	case ref.Var != "":
		varName := ref.Var
		res, ok := store.vars[varName]
		if !ok {
			if ref.IgnoreMissing {
				return &Result{Output: ref.Default}, nil
			}
			return nil, fmt.Errorf("could not find variable %q", varName)
		}
		return &Result{Output: res}, nil
	case ref.Env != "":
		return &Result{Output: os.Getenv(ref.Env)}, nil
	case ref.Ref != "":
		refName := ref.Ref
		res, ok := store.references[refName]
		if !ok {
			if ref.IgnoreMissing {
				return &Result{Output: ref.Default}, nil
			}
			return nil, fmt.Errorf("could not find reference %q", refName)
		}
		return res, nil
	case ref.File != "":
		res, err := os.ReadFile(path.Join(dir, ref.File))
		if err != nil {
			if errors.Is(err, os.ErrNotExist) && ref.IgnoreMissing {
				return &Result{Output: ref.Default}, nil
			}
			return nil, fmt.Errorf("error opening file %q", ref.File)
		}
		format := formatFromFileName(ref.File)
		return &Result{Output: res, Format: format}, nil
	case ref.Value != nil:
		ret, err := store.GetAnyValue(dir, *ref.Value)
		if err != nil {
			return nil, fmt.Errorf("error getting value: %w", err)
		}
		return ret, nil
	case ref.Values != nil:
		var vals []any
		for _, v := range ref.Values {
			ret, err := store.GetAnyValue(dir, v)
			if err != nil {
				return nil, fmt.Errorf("error getting value: %w", err)
			}
			vals = append(vals, ret)
		}
		return &Result{Output: vals}, nil
	case ref.PipelineGenerator != nil:
		err := config.ValidatePipelineGenerators(*ref.PipelineGenerator)
		if err != nil {
			return nil, fmt.Errorf("error getting value: %w", err)
		}
		subPipeline := NewPipeline(dir, *ref.PipelineGenerator, store, false)
		res, err := subPipeline.Generate(context.TODO())
		if err != nil {
			return nil, fmt.Errorf("error getting value: %w", err)
		}
		return res, nil
	default:
		return nil, errors.New("invalid reference, must specify a reference type")
	}
}

// ParsedValue stores the parsed Value retrieved from the store, and the original data before parsing.
type ParsedValue struct {
	parsed any
	data   []byte
}

func (pv ParsedValue) Data() []byte {
	return pv.data
}

func (pv ParsedValue) Parsed() any {
	return pv.parsed
}

func (store *Store) getParsedValueDecoder(data []byte, format string) (Decoder, error) {
	dec, err := NewDecoder(format, data)
	if err != nil {
		return nil, fmt.Errorf("error creating decoder: %w", err)
	}
	return dec, nil
}

func (store *Store) GetParsedValues(dir string, val config.Value) (iter.Seq2[ParsedValue, error], error) {
	res, err := store.getValue(dir, val)
	if err != nil {
		return nil, err
	}

	var data []byte
	switch v := res.Output.(type) {
	case string:
		data = []byte(v)
	case []byte:
		data = v
	default:
		return store.convertToParsedValueIter(res)
	}

	format := res.Format
	if format == "" {
		format = val.Format
	}
	if format == "" {
		return nil, fmt.Errorf("cannot handle value of type %T without format set", res.Output)
	}
	dec, err := store.getParsedValueDecoder(data, format)
	if err != nil {
		return nil, err
	}
	return func(yield func(ParsedValue, error) bool) {
		for {
			pv := ParsedValue{
				data: data,
			}
			err := dec.Decode(&pv.parsed)
			if err == io.EOF {
				return
			}
			if !yield(pv, err) {
				return
			}
		}
	}, nil
}

func (store *Store) convertToParsedValueIter(val *Result) (iter.Seq2[ParsedValue, error], error) {
	return func(yield func(ParsedValue, error) bool) {
		var items []any
		if vals, ok := val.Output.([]any); ok {
			items = vals
		} else {
			items = []any{val.Output}
		}
		for _, val := range items {
			pv := ParsedValue{
				parsed: val,
			}
			if !yield(pv, nil) {
				return
			}
		}
	}, nil
}

type Decoder interface {
	Decode(any) error
}

func NewDecoder(format string, data []byte) (Decoder, error) {
	switch format {
	case "yaml":
		return config.NewYAMLDecoder(bytes.NewBuffer(data)), nil
	case "json":
		return json.NewDecoder(bytes.NewBuffer(data)), nil
	case "":
		return nil, errors.New("input.format is required")
	default:
		return nil, fmt.Errorf("invalid input format specified: %q", format)
	}
}

func ConvertToBytes(res *Result) ([]byte, error) {
	if res == nil || res.Output == nil {
		return nil, nil
	}
	switch val := res.Output.(type) {
	case string:
		return []byte(val), nil
	case []byte:
		return val, nil
	default:
		return config.EncodeYAML(res.Output)
	}
}
