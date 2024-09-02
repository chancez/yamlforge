package generator

import (
	"context"
	"errors"
	"fmt"

	"github.com/chancez/yamlforge/pkg/config"
	"github.com/chancez/yamlforge/pkg/reference"
)

var _ Generator = (*Pipeline)(nil)

func init() {
	Register("pipeline", config.PipelineGenerator{}, func(dir string, cfg any, refStore *reference.Store) Generator {
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
	if pipeline.cfg.Generator != nil && len(pipeline.cfg.Pipeline) != 0 {
		return nil, errors.New("cannot set both 'pipeline' and 'generator' options")
	}

	if pipeline.cfg.Generator != nil {
		return pipeline.executeGenerator(ctx, *pipeline.cfg.Generator)
	}

	var output []byte
	for _, stage := range pipeline.cfg.Pipeline {
		if stage.Generator == nil {
			return nil, fmt.Errorf("error in stage %q: generator cannot be empty", stage.Name)
		}
		result, err := pipeline.executeGenerator(ctx, *stage.Generator)
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

func (pipeline *Pipeline) executeGenerator(ctx context.Context, generatorCfg config.Generator) ([]byte, error) {
	name, gen, err := GlobalRegistry.GetGenerator(pipeline.dir, pipeline.refStore, generatorCfg)
	if err != nil {
		return nil, fmt.Errorf("error getting generator: %w", err)
	}
	result, err := gen.Generate(ctx)
	if err != nil {
		return nil, fmt.Errorf("error executing %q generator: %w", name, err)
	}
	return result, nil
}
