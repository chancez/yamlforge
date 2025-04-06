package generator

import (
	"bytes"
	"context"
	"fmt"
	"io"

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
		data, err := y.refStore.GetValueBytes(y.dir, input)
		if err != nil {
			return nil, fmt.Errorf("error getting value: %w", err)
		}
		dec := config.NewYAMLDecoder(bytes.NewBuffer(data))
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
