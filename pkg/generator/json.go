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
	Register(config.JSONGenerator{}, func(_ string, refStore *reference.Store, cfg any) Generator {
		return NewJSON(cfg.(config.JSONGenerator), refStore)
	})
}

type JSON struct {
	cfg      config.JSONGenerator
	refStore *reference.Store
}

func NewJSON(cfg config.JSONGenerator, refStore *reference.Store) *JSON {
	return &JSON{
		cfg:      cfg,
		refStore: refStore,
	}
}

func (y *JSON) Generate(context.Context) ([]byte, error) {
	var out bytes.Buffer
	enc := json.NewEncoder(&out)
	for _, input := range y.cfg.Input {
		ref, err := y.refStore.GetReference(input)
		if err != nil {
			return nil, fmt.Errorf("error getting reference: %w", err)
		}
		// assume that all refs are JSON for now
		dec := yaml.NewDecoder(bytes.NewBuffer(ref))
		var tmp map[string]any
		for {
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
