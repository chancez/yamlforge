package generator

import (
	"bytes"
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
	references map[string]any
	// map a variable name to it's value
	vars map[string]any
}

func NewStore(vars map[string]any) *Store {
	return &Store{
		references: make(map[string]any),
		vars:       vars,
	}
}

func (store *Store) AddReference(name string, result any) error {
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

func (store *Store) GetAnyValue(dir string, val config.AnyOrValue) (any, error) {
	if val.Any != nil {
		return *val.Any, nil
	}
	if val.Value != nil {
		return store.GetValue(dir, *val.Value)
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
		if s, ok := v.(string); ok {
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
		if b, ok := v.(bool); ok {
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
		if mapVal, ok := v.(map[string]any); ok {
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

func (store *Store) GetValue(dir string, ref config.Value) (any, error) {
	switch {
	case ref.Var != "":
		varName := ref.Var
		res, ok := store.vars[varName]
		if !ok {
			if ref.IgnoreMissing {
				return ref.Default, nil
			}
			return nil, fmt.Errorf("could not find variable %q", varName)
		}
		return res, nil
	case ref.Env != "":
		return os.Getenv(ref.Env), nil
	case ref.Ref != "":
		refName := ref.Ref
		res, ok := store.references[refName]
		if !ok {
			if ref.IgnoreMissing {
				return ref.Default, nil
			}
			return nil, fmt.Errorf("could not find reference %q", refName)
		}
		return res, nil
	case ref.File != "":
		res, err := os.ReadFile(path.Join(dir, ref.File))
		if err != nil {
			if errors.Is(err, os.ErrNotExist) && ref.IgnoreMissing {
				return ref.Default, nil
			}
			return nil, fmt.Errorf("error opening file %q", ref.File)
		}
		return res, nil
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
		return vals, nil
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
	if format == "" {
		format = "yaml"
	}
	dec, err := NewDecoder(format, data)
	if err != nil {
		return nil, fmt.Errorf("error creating decoder: %w", err)
	}
	return dec, nil
}

func (store *Store) GetParsedValues(dir string, parsedVal config.ParsedValue) (iter.Seq2[ParsedValue, error], error) {
	val, err := store.GetValue(dir, parsedVal.Value)
	if err != nil {
		return nil, err
	}
	if valBytes, ok := val.([]byte); ok {
		dec, err := store.getParsedValueDecoder(valBytes, parsedVal.Format)
		if err != nil {
			return nil, err
		}
		return func(yield func(ParsedValue, error) bool) {
			for {
				pv := ParsedValue{
					data: valBytes,
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
	} else {
		return func(yield func(ParsedValue, error) bool) {
			var items []any
			if vals, ok := val.([]any); ok {
				items = vals
			} else {
				items = []any{val}
			}
			for _, val := range items {
				pv := ParsedValue{
					parsed: val,
				}
				if !yield(pv, err) {
					return
				}
			}
		}, nil
	}
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

func ConvertToBytes(val any) ([]byte, error) {
	if val == nil {
		return nil, nil
	}
	switch val := val.(type) {
	case string:
		return []byte(val), nil
	case []byte:
		return val, nil
	}
	return config.EncodeYAML(val)
}
