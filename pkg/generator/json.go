package generator

import (
	"bytes"
	"context"
	"encoding/json"
	"fmt"
	"io"
	"strings"

	"github.com/chancez/yamlforge/pkg/config"
	"github.com/chancez/yamlforge/pkg/reference"
)

var _ Generator = (*JSON)(nil)

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

func (j *JSON) Generate(context.Context) ([]byte, error) {
	var out bytes.Buffer
	enc := json.NewEncoder(&out)
	if j.cfg.Indent != 0 {
		enc.SetIndent("", strings.Repeat(" ", j.cfg.Indent))
	}
	for _, input := range j.cfg.Input {
		ref, err := j.refStore.GetReference(j.dir, input)
		if err != nil {
			return nil, fmt.Errorf("error getting reference: %w", err)
		}
		dec := config.NewYAMLDecoder(bytes.NewBuffer(ref))
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
