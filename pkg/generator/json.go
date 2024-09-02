package generator

import (
	"bytes"
	"context"
	"encoding/json"
	"fmt"
	"io"

	"github.com/chancez/yamlforge/pkg/config"
	"github.com/chancez/yamlforge/pkg/reference"
	"gopkg.in/yaml.v3"
)

var _ Generator = (*JSON)(nil)

func init() {
	Register("json", config.JSONGenerator{}, func(dir string, cfg any, refStore *reference.Store) Generator {
		return NewJSON(dir, cfg.(config.JSONGenerator), refStore)
	})
}

type JSON struct {
	dir      string
	cfg      config.JSONGenerator
	refStore *reference.Store
}

func NewJSON(dir string, cfg config.JSONGenerator, refStore *reference.Store) *JSON {
	return &JSON{
		dir:      dir,
		cfg:      cfg,
		refStore: refStore,
	}
}

func (y *JSON) Generate(context.Context) ([]byte, error) {
	var out bytes.Buffer
	enc := json.NewEncoder(&out)
	for _, input := range y.cfg.Input {
		ref, err := y.refStore.GetReference(y.dir, input)
		if err != nil {
			return nil, fmt.Errorf("error getting reference: %w", err)
		}
		dec := yaml.NewDecoder(bytes.NewBuffer(ref))
		for {
			var tmp any
			err = dec.Decode(&tmp)
			if err == io.EOF {
				break
			}
			if err != nil {
				return nil, fmt.Errorf("error decoding reference as YAML: %w", err)
			}
			err = enc.Encode(tmp)
			if err != nil {
				return nil, fmt.Errorf("error writing JSON: %w", err)
			}
		}
	}
	return out.Bytes(), nil
}
