package generator

import (
	"bytes"
	"context"
	"fmt"
	"io"

	"github.com/chancez/yamlforge/pkg/config"
	"github.com/chancez/yamlforge/pkg/reference"
	"gopkg.in/yaml.v3"
)

var _ Generator = (*YAML)(nil)

func init() {
	Register(config.YAMLGenerator{}, func(_ string, refStore *reference.Store, cfg any) Generator {
		return NewYAML(cfg.(config.YAMLGenerator), refStore)
	})
}

type YAML struct {
	cfg      config.YAMLGenerator
	refStore *reference.Store
}

func NewYAML(cfg config.YAMLGenerator, refStore *reference.Store) *YAML {
	return &YAML{
		cfg:      cfg,
		refStore: refStore,
	}
}

func (y *YAML) Generate(context.Context) ([]byte, error) {
	var out bytes.Buffer
	enc := yaml.NewEncoder(&out)
	for _, input := range y.cfg.Input {
		ref, err := y.refStore.GetReference(input)
		if err != nil {
			return nil, fmt.Errorf("error getting reference: %w", err)
		}
		// assume that all refs are YAML for now
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
				return nil, fmt.Errorf("error writing YAML: %w", err)
			}
		}
	}
	err := enc.Close()
	if err != nil {
		return nil, fmt.Errorf("error writing YAML: %w", err)
	}
	return out.Bytes(), nil
}
