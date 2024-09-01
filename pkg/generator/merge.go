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
	Register(config.MergeGenerator{}, func(_ string, refStore *reference.Store, cfg any) Generator {
		return NewMerge(cfg.(config.MergeGenerator), refStore)
	})
}

type Merge struct {
	cfg      config.MergeGenerator
	refStore *reference.Store
}

func NewMerge(cfg config.MergeGenerator, refStore *reference.Store) *Merge {
	return &Merge{
		cfg:      cfg,
		refStore: refStore,
	}
}

func (m *Merge) Generate(_ context.Context) ([]byte, error) {
	merged := make(map[string]any)
	for _, input := range m.cfg.Input {
		ref, err := m.refStore.GetReference(input)
		if err != nil {
			return nil, fmt.Errorf("error getting reference: %w", err)
		}
		var item map[string]any
		err = yaml.Unmarshal(ref, &item)
		if err != nil {
			return nil, fmt.Errorf("error parsing reference as YAML: %w", err)
		}
		merged = mapmerge.Merge(merged, item)
	}
	out, err := yaml.Marshal(merged)
	if err != nil {
		return nil, fmt.Errorf("error while marshaling merged results to YAML: %w", err)
	}
	return out, nil
}
