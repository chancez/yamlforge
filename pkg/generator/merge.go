package generator

import (
	"context"
	"fmt"

	"github.com/chancez/yamlforge/pkg/config"
	"github.com/chancez/yamlforge/pkg/mapmerge"
)

var _ Generator = (*Merge)(nil)

type Merge struct {
	dir      string
	cfg      config.MergeGenerator
	refStore *Store
}

func NewMerge(dir string, cfg config.MergeGenerator, refStore *Store) *Merge {
	return &Merge{
		dir:      dir,
		cfg:      cfg,
		refStore: refStore,
	}
}

func (m *Merge) Generate(_ context.Context) ([]byte, error) {
	merged := make(map[string]any)
	for _, input := range m.cfg.Input {
		val, err := m.refStore.GetMapValue(m.dir, input)
		if err != nil {
			return nil, fmt.Errorf("error getting value: %w", err)
		}
		merged = mapmerge.Merge(merged, val)
	}
	out, err := config.EncodeYAML(merged)
	if err != nil {
		return nil, fmt.Errorf("error while marshaling merged results to YAML: %w", err)
	}
	return out, nil
}
