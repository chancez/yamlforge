package reference

import (
	"errors"
	"fmt"
	"os"

	"gopkg.in/yaml.v3"

	"github.com/chancez/yamlforge/pkg/config"
)

type Store struct {
	// map from an stage.Name to it's results
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

func (store *Store) GetReference(ref config.Reference) ([]byte, error) {
	switch {
	case ref.Var != nil:
		varName := *ref.Var
		res, ok := store.vars[varName]
		if !ok {
			return nil, fmt.Errorf("could not find variable %q", varName)
		}
		return []byte(res), nil
	case ref.Ref != nil:
		refName := *ref.Ref
		res, ok := store.references[refName]
		if !ok {
			return nil, fmt.Errorf("could not find reference %q", refName)
		}
		return res, nil
	case ref.File != nil:
		return os.ReadFile(*ref.File)
	case ref.Literal != nil:
		return yaml.Marshal(ref.Literal)
	default:
		return nil, errors.New("invalid reference, must specify a reference type")
	}
}