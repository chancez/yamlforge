package generator

import (
	"context"

	"github.com/chancez/yamlforge/pkg/config"
	"github.com/chancez/yamlforge/pkg/reference"
)

var _ Generator = (*Value)(nil)

type Value struct {
	dir      string
	cfg      config.ValueGenerator
	refStore *reference.Store
}

func NewValue(dir string, cfg config.ValueGenerator, refStore *reference.Store) *Value {
	return &Value{
		dir:      dir,
		cfg:      cfg,
		refStore: refStore,
	}
}

func (v *Value) Generate(context.Context) ([]byte, error) {
	return v.refStore.GetReference(v.dir, v.cfg.Input)
}
