package reference

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
	return store.getReference(dir, ref)
}

func (store *Store) GetAnyValue(dir string, val config.AnyValue) (any, error) {
	if val.Any != nil {
		return *val.Any, nil
	}
	if val.Value != nil {
		return store.getReference(dir, *val.Value)
	}
	panic("invalid AnyValue")
}

func (store *Store) GetStringValue(dir string, val config.StringValue) (string, error) {
	if val.String != nil {
		return *val.String, nil
	}
	if val.Value != nil {
		data, err := store.getReference(dir, *val.Value)
		if err != nil {
			return "", err
		}
		var s string
		err = config.DecodeYAML(data, &s)
		if err != nil {
			return "", err
		}
		return s, err
	}
	panic("invalid StringValue")
}

func (store *Store) GetBoolValue(dir string, val config.BoolValue) (bool, error) {
	if val.Bool != nil {
		return *val.Bool, nil
	}
	if val.Value != nil {
		data, err := store.getReference(dir, *val.Value)
		if err != nil {
			return false, err
		}
		var b bool
		err = config.DecodeYAML(data, &b)
		if err != nil {
			return false, err
		}
		return b, err
	}
	panic("invalid BoolValue")
}

func (store *Store) getReference(dir string, ref config.Value) ([]byte, error) {
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

func (store *Store) GetParsedValues(dir string, parsedVal config.ParsedValue) (iter.Seq2[any, error], error) {
	dec, err := store.getParsedValueDecoder(dir, parsedVal)
	if err != nil {
		return nil, err
	}
	return func(yield func(any, error) bool) {
		for {
			var val any
			err := dec.Decode(&val)
			if err == io.EOF {
				return
			}
			if !yield(val, err) {
				return
			}
		}
	}, nil
}

func (store *Store) getParsedValueDecoder(dir string, val config.ParsedValue) (Decoder, error) {
	format := val.Format
	if format == "" && val.File != "" {
		switch path.Ext(val.File) {
		case "yaml":
			format = "yaml"
		case "json":
			format = "json"
		}
	}
	if format == "" {
		format = "yaml"
	}
	data, err := store.getReference(dir, val.Value)
	if err != nil {
		return nil, fmt.Errorf("error getting reference: %w", err)
	}
	return NewDecoder(format, data)
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
