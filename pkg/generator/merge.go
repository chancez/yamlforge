package generator

import (
	"context"
	"fmt"

	"github.com/chancez/yamlforge/pkg/config"
	"github.com/chancez/yamlforge/pkg/mapmerge"
	"github.com/chancez/yamlforge/pkg/reference"
)

var _ Generator = (*Merge)(nil)

type Merge struct {
	dir      string
	cfg      config.MergeGenerator
	refStore *reference.Store
}

func NewMerge(dir string, cfg config.MergeGenerator, refStore *reference.Store) *Merge {
	return &Merge{
		dir:      dir,
		cfg:      cfg,
		refStore: refStore,
	}
}

func (m *Merge) Generate(_ context.Context) ([]byte, error) {
	merged := make(map[string]any)
	for _, input := range m.cfg.Input {
		vals, err := m.refStore.GetParsedValues(m.dir, input)
		if err != nil {
			return nil, fmt.Errorf("error getting value: %w", err)
		}
		for val, err := range vals {
			if err != nil {
				return nil, fmt.Errorf("error while processing input: %w", err)
			}
			itemMap, ok := val.Parsed().(map[string]any)
			if !ok {
				// Provide a snippet of the data being merged
				refStr := string(val.Data())
				if len(refStr) > 20 {
					refStr = refStr[0:20]
				}
				return nil, fmt.Errorf("unable to merge non-map values from %q", refStr)
			}
			merged = mapmerge.Merge(merged, itemMap)
		}
	}
	out, err := config.EncodeYAML(merged)
	if err != nil {
		return nil, fmt.Errorf("error while marshaling merged results to YAML: %w", err)
	}
	return out, nil
}
