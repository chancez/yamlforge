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
	references map[string][]byte
	// map a variable name to it's value
	vars map[string][]byte
}

func NewStore(vars map[string][]byte) *Store {
	return &Store{
		references: make(map[string][]byte),
		vars:       vars,
	}
}

func (store *Store) AddReference(name string, data []byte) error {
	if _, exists := store.references[name]; exists {
		return fmt.Errorf("reference %q already exists", name)
	}
	store.references[name] = data
	return nil
}

func (store *Store) GetValueBytes(dir string, ref config.Value) ([]byte, error) {
	return store.getValue(dir, ref)
}

func (store *Store) GetAnyValue(dir string, val config.AnyOrValue) (any, error) {
	if val.Any != nil {
		return *val.Any, nil
	}
	if val.Value != nil {
		return store.getValue(dir, *val.Value)
	}
	return nil, nil
}

func (store *Store) GetStringValue(dir string, val config.StringOrValue) (string, error) {
	if val.String != nil {
		return *val.String, nil
	}
	if val.Value != nil {
		data, err := store.getValue(dir, *val.Value)
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
		for _, apiVersion := range vals {
			apiV, err := store.GetStringValue(dir, apiVersion)
			if err != nil {
				return nil, err
			}
			ret = append(ret, apiV)
		}
	}
	return ret, nil
}

func (store *Store) GetBoolValue(dir string, val config.BoolOrValue) (bool, error) {
	if val.Bool != nil {
		return *val.Bool, nil
	}
	if val.Value != nil {
		data, err := store.getValue(dir, *val.Value)
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
		data, err := store.getValue(dir, *val.Value)
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

func (store *Store) getValue(dir string, ref config.Value) ([]byte, error) {
	switch {
	case ref.Var != "":
		varName := ref.Var
		res, ok := store.vars[varName]
		if !ok {
			if ref.IgnoreMissing {
				return ConvertToBytes(ref.Default)
			}
			return nil, fmt.Errorf("could not find variable %q", varName)
		}
		return []byte(res), nil
	case ref.Env != "":
		return []byte(os.Getenv(ref.Env)), nil
	case ref.Ref != "":
		refName := ref.Ref
		res, ok := store.references[refName]
		if !ok {
			if ref.IgnoreMissing {
				return ConvertToBytes(ref.Default)
			}
			return nil, fmt.Errorf("could not find reference %q", refName)
		}
		return res, nil
	case ref.File != "":
		res, err := os.ReadFile(path.Join(dir, ref.File))
		if err != nil {
			if errors.Is(err, os.ErrNotExist) && ref.IgnoreMissing {
				return ConvertToBytes(ref.Default)
			}
			return nil, fmt.Errorf("error opening file %q", ref.File)
		}
		return res, nil
	case ref.Value != nil:
		return ConvertToBytes(ref.Value)
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

func (store *Store) getParsedValueDecoder(dir string, parsedVal config.ParsedValue) (Decoder, []byte, error) {
	data, err := store.getValue(dir, parsedVal.Value)
	if err != nil {
		return nil, nil, fmt.Errorf("error getting reference: %w", err)
	}
	format := parsedVal.Format
	if format == "" && parsedVal.File != "" {
		switch path.Ext(parsedVal.File) {
		case "yaml":
			format = "yaml"
		case "json":
			format = "json"
		}
	}
	if format == "" {
		format = "yaml"
	}
	dec, err := NewDecoder(format, data)
	if err != nil {
		return nil, nil, fmt.Errorf("error creating decoder: %w", err)
	}
	return dec, data, nil
}

func (store *Store) GetParsedValues(dir string, parsedVal config.ParsedValue) (iter.Seq2[ParsedValue, error], error) {
	dec, data, err := store.getParsedValueDecoder(dir, parsedVal)
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

func (store *Store) GetParsedValue(dir string, parsedVal config.ParsedValue) (ParsedValue, error) {
	dec, data, err := store.getParsedValueDecoder(dir, parsedVal)
	if err != nil {
		return ParsedValue{}, err
	}
	var vals []any
	pv := ParsedValue{
		data: data,
	}
	for {
		var tmp any
		err := dec.Decode(&tmp)
		if err == io.EOF {
			break
		}
		vals = append(vals, tmp)
	}
	if len(vals) == 1 {
		pv.parsed = vals[0]
	} else {
		pv.parsed = vals
	}
	return pv, nil
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
