package generator

import (
	"context"
	"errors"
	"fmt"
	"path/filepath"
	"strings"

	"github.com/chancez/yamlforge/pkg/config"
)

var _ Generator = (*Pipeline)(nil)

type Pipeline struct {
	dir      string
	cfg      config.PipelineGenerator
	refStore *Store
	debug    bool
}

func NewPipeline(dir string, cfg config.PipelineGenerator, refStore *Store, debug bool) *Pipeline {
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
	if pipeline.cfg.Include != nil {
		valuesSet = append(valuesSet, "include")
	}
	if len(valuesSet) == 0 {
		return nil, errors.New("must configure a pipeline, generator, import or include")
	}
	if len(valuesSet) > 1 {
		return nil, fmt.Errorf("invalid configuration, cannot combine %s", strings.Join(valuesSet, ","))
	}

	if pipeline.cfg.Import != nil {
		return pipeline.executeImport(ctx)
	}

	if pipeline.cfg.Include != nil {
		return pipeline.executeInclude(ctx)
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
	data, err := pipeline.refStore.GetValueBytes(pipeline.dir, *pipeline.cfg.Import)
	if err != nil {
		return nil, fmt.Errorf("error getting value to import: %w", err)
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
		ref, err := pipeline.refStore.GetValueBytes(pipeline.dir, pipelineVar.Value)
		if err != nil {
			return nil, fmt.Errorf("variable %q: error getting pipeline variable reference: %w", pipelineVar.Name, err)
		}
		varName := pipelineVar.Name
		pipelineVars[varName] = ref
	}

	// FIXME: We need to know if the sub-pipeline being referenced is a file, and
	// if so, execute the sub-pipeline with the directory set to the directory of
	// the sub-pipeline, not the parent.
	// Ideally the pipeline wouldn't need to think about this, so GetReference
	// should potentially return the original path if the reference was to a
	// file.
	subPipelineDir := pipeline.dir
	if pipeline.cfg.Import.File != "" {
		if !filepath.IsAbs(pipeline.cfg.Import.File) {
			subPipelineDir = filepath.Dir(filepath.Join(pipeline.dir, pipeline.cfg.Import.File))
		} else {
			subPipelineDir = filepath.Dir(pipeline.cfg.Import.File)
		}
	}
	newStore := NewStore(pipelineVars)
	subPipeline := NewPipeline(subPipelineDir, subPipelineCfg.PipelineGenerator, newStore, pipeline.debug)
	return subPipeline.Generate(ctx)
}

func (pipeline *Pipeline) executeInclude(ctx context.Context) ([]byte, error) {
	data, err := pipeline.refStore.GetValueBytes(pipeline.dir, *pipeline.cfg.Include)
	if err != nil {
		return nil, fmt.Errorf("error getting value to import: %w", err)
	}

	subPipelineCfg, err := config.Parse(data)
	if err != nil {
		return nil, fmt.Errorf("error parsing pipeline: %w", err)
	}

	subPipeline := NewPipeline(pipeline.dir, subPipelineCfg.PipelineGenerator, pipeline.refStore, pipeline.debug)
	return subPipeline.Generate(ctx)

}

func (pipeline *Pipeline) executeGenerator(ctx context.Context, generatorCfg config.Generator) ([]byte, error) {
	kind, gen, err := pipeline.getGenerator(generatorCfg)
	if err != nil {
		return nil, fmt.Errorf("error getting generator: %w", err)
	}
	result, err := gen.Generate(ctx)
	if err != nil {
		return nil, fmt.Errorf("error executing %q generator: %w", kind, err)
	}
	if pipeline.debug {
		fmt.Printf("[DEBUG (generator: %q) - name: %q]:\n%s\n\n", kind, generatorCfg.Name, string(result))
	}
	return result, nil
}

func (pipeline *Pipeline) getGenerator(generatorCfg config.Generator) (string, Generator, error) {
	var (
		kind string
		gen  Generator
	)
	switch {
	case generatorCfg.File != nil:
		kind = "file"
		gen = NewFile(pipeline.dir, *generatorCfg.File)
	case generatorCfg.Value != nil:
		kind = "value"
		gen = NewValue(pipeline.dir, *generatorCfg.Value, pipeline.refStore)
	case generatorCfg.Exec != nil:
		kind = "exec"
		gen = NewExec(pipeline.dir, *generatorCfg.Exec, pipeline.refStore)
	case generatorCfg.Helm != nil:
		kind = "helm"
		gen = NewHelm(pipeline.dir, *generatorCfg.Helm, pipeline.refStore)
	case generatorCfg.Kustomize != nil:
		kind = "kustomize"
		gen = NewKustomize(pipeline.dir, *generatorCfg.Kustomize, pipeline.refStore)
	case generatorCfg.Merge != nil:
		kind = "merge"
		gen = NewMerge(pipeline.dir, *generatorCfg.Merge, pipeline.refStore)
	case generatorCfg.GoTemplate != nil:
		kind = "gotemplate"
		gen = NewGoTemplate(pipeline.dir, *generatorCfg.GoTemplate, pipeline.refStore)
	case generatorCfg.Pipeline != nil:
		kind = "pipeline"
		gen = NewPipeline(pipeline.dir, *generatorCfg.Pipeline, pipeline.refStore, pipeline.debug)
	case generatorCfg.JQ != nil:
		kind = "jq"
		gen = NewJQ(pipeline.dir, *generatorCfg.JQ, pipeline.refStore)
	case generatorCfg.CEL != nil:
		kind = "cel"
		gen = NewCEL(pipeline.dir, *generatorCfg.CEL, pipeline.refStore)
	case generatorCfg.JSONPatch != nil:
		kind = "jsonpatch"
		gen = NewJSONPatch(pipeline.dir, *generatorCfg.JSONPatch, pipeline.refStore)
	case generatorCfg.YAML != nil:
		kind = "yaml"
		gen = NewYAML(pipeline.dir, *generatorCfg.YAML, pipeline.refStore)
	case generatorCfg.JSON != nil:
		kind = "json"
		gen = NewJSON(pipeline.dir, *generatorCfg.JSON, pipeline.refStore)
	default:
		return "", nil, fmt.Errorf("generator not configured")
	}
	return kind, gen, nil
}
