package generator

import (
	"context"
	"fmt"
	"reflect"

	"github.com/chancez/yamlforge/pkg/config"
	"github.com/chancez/yamlforge/pkg/reference"
)

type Generator interface {
	Generate(context.Context) ([]byte, error)
}

type Registry struct {
	typeToFactory map[reflect.Type]GeneratorFactory
}

type NewGeneratorFunc func(dir string, refStore *reference.Store, cfg any) Generator

type GeneratorFactory struct {
	name string
	new  NewGeneratorFunc
}

func NewRegistry() *Registry {
	return &Registry{typeToFactory: make(map[reflect.Type]GeneratorFactory)}
}

func (reg *Registry) GetGenerator(dir string, refStore *reference.Store, generatorCfg config.Generator) (string, Generator, error) {
	v := reflect.ValueOf(generatorCfg)

	setFields := 0
	var cfg any
	for i := 0; i < v.NumField(); i++ {
		field := v.Field(i)
		if field.Kind() != reflect.Pointer {
			panic(fmt.Sprintf("expected config.Generator field %q to be a pointer, got: %q", field.Type().Name(), field.Kind().String()))
		}
		if !field.IsNil() {
			setFields++
			elem := field.Elem()
			cfg = elem.Interface()
		}
	}
	if setFields == 0 {
		return "", nil, fmt.Errorf("generator not configured")
	}
	if setFields > 1 {
		return "", nil, fmt.Errorf("multiple generators configured")
	}

	cfgType := reflect.TypeOf(cfg)
	factory, exists := reg.typeToFactory[cfgType]
	if !exists {
		panic(fmt.Sprintf("cannot find factory for %s", cfgType))
	}

	return factory.name, factory.new(dir, refStore, cfg), nil
}

func (reg *Registry) Register(name string, cfgType any, newGenerator NewGeneratorFunc) {
	ty := reflect.TypeOf(cfgType)
	if _, exists := reg.typeToFactory[ty]; exists {
		panic(fmt.Sprintf("duplicate generator registered for %q", ty.String()))
	}
	factory := GeneratorFactory{
		name: name,
		new:  newGenerator,
	}
	reg.typeToFactory[ty] = factory
}

var GlobalRegistry = NewRegistry()

func Register(name string, cfgType any, newGenerator NewGeneratorFunc) {
	GlobalRegistry.Register(name, cfgType, newGenerator)
}
