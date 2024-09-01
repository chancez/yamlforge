package generator

import (
	"context"
	"fmt"

	"github.com/chancez/yamlforge/pkg/config"
	"github.com/chancez/yamlforge/pkg/reference"
)

var _ Generator = (*Pipeline)(nil)

func init() {
	Register(config.PipelineGenerator{}, func(dir string, refStore *reference.Store, cfg any) Generator {
		return NewPipeline(dir, cfg.(config.PipelineGenerator), refStore)
	})
}

type Pipeline struct {
	dir      string
	cfg      config.PipelineGenerator
	refStore *reference.Store
}

func NewPipeline(dir string, cfg config.PipelineGenerator, refStore *reference.Store) *Pipeline {
	return &Pipeline{
		dir:      dir,
		cfg:      cfg,
		refStore: refStore,
	}
}

func (pipeline *Pipeline) Generate(ctx context.Context) ([]byte, error) {
	var output []byte
	for _, stage := range pipeline.cfg.Pipeline {
		if stage.Generator == nil {
			return nil, fmt.Errorf("error in stage %q: generator cannot be empty", stage.Name)
		}
		gen, err := GlobalRegistry.GetGenerator(pipeline.dir, pipeline.refStore, *stage.Generator)
		if err != nil {
			return nil, fmt.Errorf("error getting generator: %w", err)
		}
		result, err := gen.Generate(ctx)
		if err != nil {
			return nil, fmt.Errorf("error running stage %q: %w", stage.Name, err)
		}
		err = pipeline.refStore.AddReference(stage.Name, result)
		if err != nil {
			return nil, fmt.Errorf("error storing reference for stage %q: %w", stage.Name, err)
		}
		// The last stage is the output of a pipeline
		output = result
	}

	return output, nil
}
