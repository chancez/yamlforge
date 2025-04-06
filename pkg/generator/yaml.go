package generator

import (
	"bytes"
	"context"
	"fmt"

	"github.com/chancez/yamlforge/pkg/config"
	"github.com/chancez/yamlforge/pkg/reference"
)

var _ Generator = (*YAML)(nil)

type YAML struct {
	dir      string
	cfg      config.YAMLGenerator
	refStore *reference.Store
}

func NewYAML(dir string, cfg config.YAMLGenerator, refStore *reference.Store) *YAML {
	return &YAML{
		dir:      dir,
		cfg:      cfg,
		refStore: refStore,
	}
}

func (y *YAML) Generate(context.Context) ([]byte, error) {
	var out bytes.Buffer
	enc := config.NewYAMLEncoderWithIndent(&out, y.cfg.Indent)
	for _, input := range y.cfg.Input {
		vals, err := y.refStore.GetParsedValues(y.dir, input)
		if err != nil {
			return nil, fmt.Errorf("error getting value: %w", err)
		}
		for val, err := range vals {
			if err != nil {
				return nil, fmt.Errorf("error while processing input: %w", err)
			}
			err = enc.Encode(val.Parsed())
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
