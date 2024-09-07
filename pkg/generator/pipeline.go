package generator

import (
	"context"
	"errors"
	"fmt"
	"path"
	"strings"

	"github.com/chancez/yamlforge/pkg/config"
	"github.com/chancez/yamlforge/pkg/reference"
)

var _ Generator = (*Pipeline)(nil)

type Pipeline struct {
	dir      string
	cfg      config.PipelineGenerator
	refStore *reference.Store
	debug    bool
}

func NewPipeline(dir string, cfg config.PipelineGenerator, refStore *reference.Store, debug bool) *Pipeline {
	return &Pipeline{
		dir:      dir,
		cfg:      cfg,
		refStore: refStore,
		debug:    debug,
	}
}

func (pipeline *Pipeline) Generate(ctx context.Context) ([]byte, error) {
	var valuesSet []string
	if pipeline.cfg.Generator != nil {
		valuesSet = append(valuesSet, "generator")
	}
	if len(pipeline.cfg.Pipeline) != 0 {
		valuesSet = append(valuesSet, "pipeline")
	}
	if pipeline.cfg.Import != nil {
		valuesSet = append(valuesSet, "import")
	}
	if len(valuesSet) == 0 {
		return nil, errors.New("must configure a pipeline, generator or import")
	}
	if len(valuesSet) > 1 {
		return nil, fmt.Errorf("invalid configuration, cannot combine %s", strings.Join(valuesSet, ","))
	}

	if pipeline.cfg.Import != nil {
		return pipeline.executeImport(ctx)
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

func (pipeline *Pipeline) executeImport(ctx context.Context) ([]byte, error) {
	dir := pipeline.dir
	data, err := pipeline.refStore.GetReference(pipeline.dir, *pipeline.cfg.Import)
	if err != nil {
		return nil, fmt.Errorf("error getting reference to import: %w", err)
	}
	if pipeline.cfg.Import.File != "" {
		dir = path.Dir(pipeline.cfg.Import.File)
	}

	subPipelineCfg, err := config.Parse(data)
	if err != nil {
		return nil, fmt.Errorf("error parsing pipeline: %w", err)
	}

	pipelineVars := make(map[string][]byte)
	for i, pipelineVar := range pipeline.cfg.Vars {
		if pipelineVar.Name == "" {
			return nil, fmt.Errorf("vars[%d]: pipeline variable name cannot be empty", i)
		}
		ref, err := pipeline.refStore.GetReference(pipeline.dir, pipelineVar.Value)
		if err != nil {
			return nil, fmt.Errorf("variable %q: error getting pipeline variable reference: %w", pipelineVar.Name, err)
		}
		varName := pipelineVar.Name
		pipelineVars[varName] = ref
	}

	newStore := reference.NewStore(pipelineVars)
	subPipeline := NewPipeline(dir, subPipelineCfg.PipelineGenerator, newStore, pipeline.debug)
	return subPipeline.Generate(ctx)
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
	if pipeline.debug {
		fmt.Printf("[DEBUG]:\n%s\n", string(result))
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
	case generatorCfg.Value != nil:
		name = "value"
		gen = NewValue(pipeline.dir, *generatorCfg.Value, pipeline.refStore)
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
	case generatorCfg.Pipeline != nil:
		name = "pipeline"
		gen = NewPipeline(pipeline.dir, *generatorCfg.Pipeline, pipeline.refStore, pipeline.debug)
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
