package generator

import (
	"context"

	"github.com/chancez/yamlforge/pkg/config"
	"github.com/chancez/yamlforge/pkg/reference"
)

var _ Generator = (*Value)(nil)

type Value struct {
	dir      string
	val      config.AnyValue
	refStore *reference.Store
}

func NewValue(dir string, val config.AnyValue, refStore *reference.Store) *Value {
	return &Value{
		dir:      dir,
		val:      val,
		refStore: refStore,
	}
}

func (v *Value) Generate(context.Context) ([]byte, error) {
	val, err := v.refStore.GetAnyValue(v.dir, v.val)
	if err != nil {
		return nil, err
	}
	return reference.ConvertToBytes(val)
}
