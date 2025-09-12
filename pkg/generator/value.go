package generator

import (
	"context"

	"github.com/chancez/yamlforge/pkg/config"
)

var _ Generator = (*Value)(nil)

type Value struct {
	dir      string
	val      config.AnyOrValue
	refStore *Store
}

func NewValue(dir string, val config.AnyOrValue, refStore *Store) *Value {
	return &Value{
		dir:      dir,
		val:      val,
		refStore: refStore,
	}
}

func (v *Value) Generate(context.Context) (*Result, error) {
	val, err := v.refStore.GetAnyValue(v.dir, v.val)
	if err != nil {
		return nil, err
	}
	return val, nil
}
