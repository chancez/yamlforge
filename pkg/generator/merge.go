package generator

import (
	"context"
	"fmt"

	"gopkg.in/yaml.v3"

	"github.com/chancez/yamlforge/pkg/config"
	"github.com/chancez/yamlforge/pkg/mapmerge"
	"github.com/chancez/yamlforge/pkg/reference"
)

var _ Generator = (*Merge)(nil)

func init() {
	Register("merge", config.MergeGenerator{}, func(dir string, cfg any, refStore *reference.Store) Generator {
		return NewMerge(dir, cfg.(config.MergeGenerator), refStore)
	})
}

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
		ref, err := m.refStore.GetReference(m.dir, input)
		if err != nil {
			return nil, fmt.Errorf("error getting reference: %w", err)
		}
		var item any
		err = yaml.Unmarshal(ref, &item)
		if err != nil {
			return nil, fmt.Errorf("error parsing reference as YAML: %w", err)
		}
		itemMap, ok := item.(map[string]any)
		if !ok {
			// Provide a snippet of the data being merged
			refStr := string(ref)
			if len(refStr) > 20 {
				refStr = refStr[0:20]
			}
			return nil, fmt.Errorf("unable to merge non-map values from %q", refStr)
		}
		merged = mapmerge.Merge(merged, itemMap)
	}
	out, err := yaml.Marshal(merged)
	if err != nil {
		return nil, fmt.Errorf("error while marshaling merged results to YAML: %w", err)
	}
	return out, nil
}
