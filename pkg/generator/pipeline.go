package generator

import (
	"context"
	"errors"
	"fmt"

	"github.com/chancez/yamlforge/pkg/config"
	"github.com/chancez/yamlforge/pkg/reference"
)

var _ Generator = (*Pipeline)(nil)

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
	for _, gen := range pipeline.cfg.Pipeline {
		result, err := pipeline.executeGenerator(ctx, gen)
		if err != nil {
			return nil, fmt.Errorf("error running stage %q: %w", gen.Name, err)
		}
		err = pipeline.refStore.AddReference(gen.Name, result)
		if err != nil {
			return nil, fmt.Errorf("error storing reference for stage %q: %w", gen.Name, err)
		}
		// The last stage is the output of a pipeline
		output = result
	}

	return output, nil
}

func (pipeline *Pipeline) executeGenerator(ctx context.Context, generatorCfg config.Generator) ([]byte, error) {
	name, gen, err := pipeline.getGenerator(generatorCfg)
	if err != nil {
		return nil, fmt.Errorf("error getting generator: %w", err)
	}
	result, err := gen.Generate(ctx)
	if err != nil {
		return nil, fmt.Errorf("error executing %q generator: %w", name, err)
	}
	return result, nil
}

func (pipeline *Pipeline) getGenerator(generatorCfg config.Generator) (string, Generator, error) {
	var (
		name string
		gen  Generator
	)
	switch {
	case generatorCfg.File != nil:
		name = "file"
		gen = NewFile(pipeline.dir, *generatorCfg.File)
	case generatorCfg.Exec != nil:
		name = "exec"
		gen = NewExec(pipeline.dir, *generatorCfg.Exec)
	case generatorCfg.Helm != nil:
		name = "helm"
		gen = NewHelm(pipeline.dir, *generatorCfg.Helm, pipeline.refStore)
	case generatorCfg.Kustomize != nil:
		name = "kustomize"
		gen = NewKustomize(pipeline.dir, *generatorCfg.Kustomize, pipeline.refStore)
	case generatorCfg.Merge != nil:
		name = "merge"
		gen = NewMerge(pipeline.dir, *generatorCfg.Merge, pipeline.refStore)
	case generatorCfg.GoTemplate != nil:
		name = "gotemplate"
		gen = NewGoTemplate(pipeline.dir, *generatorCfg.GoTemplate, pipeline.refStore)
	case generatorCfg.Import != nil:
		name = "import"
		gen = NewImport(pipeline.dir, *generatorCfg.Import, pipeline.refStore)
	case generatorCfg.JQ != nil:
		name = "jq"
		gen = NewJQ(pipeline.dir, *generatorCfg.JQ, pipeline.refStore)
	case generatorCfg.CELFilter != nil:
		name = "celfilter"
		gen = NewCELFilter(pipeline.dir, *generatorCfg.CELFilter, pipeline.refStore)
	case generatorCfg.YAML != nil:
		name = "yaml"
		gen = NewYAML(pipeline.dir, *generatorCfg.YAML, pipeline.refStore)
	case generatorCfg.JSON != nil:
		name = "json"
		gen = NewJSON(pipeline.dir, *generatorCfg.JSON, pipeline.refStore)
	default:
		return "", nil, fmt.Errorf("generator not configured")
	}
	return name, gen, nil
}
