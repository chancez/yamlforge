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
		data, err := m.refStore.GetValueBytes(m.dir, input)
		if err != nil {
			return nil, fmt.Errorf("error getting value: %w", err)
		}
		var item any
		err = config.DecodeYAML(data, &item)
		if err != nil {
			return nil, fmt.Errorf("error parsing reference as YAML: %w", err)
		}
		itemMap, ok := item.(map[string]any)
		if !ok {
			// Provide a snippet of the data being merged
			refStr := string(data)
			if len(refStr) > 20 {
				refStr = refStr[0:20]
			}
			return nil, fmt.Errorf("unable to merge non-map values from %q", refStr)
		}
		merged = mapmerge.Merge(merged, itemMap)
	}
	out, err := config.EncodeYAML(merged)
	if err != nil {
		return nil, fmt.Errorf("error while marshaling merged results to YAML: %w", err)
	}
	return out, nil
}
