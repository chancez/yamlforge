package reference

import (
	"errors"
	"fmt"
	"os"
	"path"

	"gopkg.in/yaml.v3"

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

func (store *Store) GetReference(dir string, ref config.Value) ([]byte, error) {
	return store.getReference(dir, ref)
}

func (store *Store) getReference(dir string, ref config.Value) ([]byte, error) {
	switch {
	case ref.Var != "":
		varName := ref.Var
		res, ok := store.vars[varName]
		if !ok {
			if ref.IgnoreMissing {
				return convertToBytes(ref.Default)
			}
			return nil, fmt.Errorf("could not find variable %q", varName)
		}
		return []byte(res), nil
	case ref.Ref != "":
		refName := ref.Ref
		res, ok := store.references[refName]
		if !ok {
			if ref.IgnoreMissing {
				return convertToBytes(ref.Default)
			}
			return nil, fmt.Errorf("could not find reference %q", refName)
		}
		return res, nil
	case ref.File != "":
		res, err := os.ReadFile(path.Join(dir, ref.File))
		if err != nil {
			if errors.Is(err, os.ErrNotExist) && ref.IgnoreMissing {
				return convertToBytes(ref.Default)
			}
			return nil, fmt.Errorf("error opening file %q", ref.File)
		}
		return res, nil
	case ref.Value != nil:
		return convertToBytes(ref.Value)
	default:
		return nil, errors.New("invalid reference, must specify a reference type")
	}
}

func convertToBytes(val any) ([]byte, error) {
	if val == nil {
		return nil, nil
	}
	switch val := val.(type) {
	case string:
		return []byte(val), nil
	case []byte:
		return val, nil
	}
	return yaml.Marshal(val)
}
