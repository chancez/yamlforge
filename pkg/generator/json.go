package generator

import (
	"bytes"
	"context"
	"encoding/json"
	"fmt"
	"strings"

	"github.com/chancez/yamlforge/pkg/config"
)

var _ Generator = (*JSON)(nil)

type JSON struct {
	dir      string
	cfg      config.JSONGenerator
	refStore *Store
}

func NewJSON(dir string, cfg config.JSONGenerator, refStore *Store) *JSON {
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
		vals, err := j.refStore.GetParsedValues(j.dir, input)
		if err != nil {
			return nil, fmt.Errorf("error getting value: %w", err)
		}
		for val, err := range vals {
			if err != nil {
				return nil, fmt.Errorf("error while processing input: %w", err)
			}
			err = enc.Encode(val.Parsed())
			if err != nil {
				return nil, fmt.Errorf("error writing JSON: %w", err)
			}
		}
	}
	return out.Bytes(), nil
}
