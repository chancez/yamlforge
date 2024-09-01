package pipeline

import (
	"context"
	"fmt"
	"os"
	"path"

	"github.com/chancez/yamlforge/pkg/config"
	"github.com/chancez/yamlforge/pkg/generator"
	"github.com/chancez/yamlforge/pkg/reference"
)

// A pipeline is also a generator
var _ generator.Generator = (*Pipeline)(nil)

type Pipeline struct {
	forgeFile string
	config    config.Config

	referenceStore *reference.Store
}

func NewPipeline(forgeFile string, cfg config.Config, refStore *reference.Store) *Pipeline {
	return &Pipeline{
		forgeFile:      forgeFile,
		config:         cfg,
		referenceStore: refStore,
	}
}

func (pipeline *Pipeline) Generate(ctx context.Context) ([]byte, error) {
	var output []byte
	for _, stage := range pipeline.config.Pipeline {
		if stage.Generator == nil {
			return nil, fmt.Errorf("error in stage %q: generator cannot be empty", stage.Name)
		}
		result, err := pipeline.handleGenerator(ctx, *stage.Generator)
		if err != nil {
			return nil, fmt.Errorf("error running stage %q: %w", stage.Name, err)
		}
		err = pipeline.referenceStore.AddReference(stage.Name, result)
		if err != nil {
			return nil, fmt.Errorf("error storing reference for stage %q: %w", stage.Name, err)
		}
		// The last stage is the output of a pipeline
		output = result
	}

	return output, nil
}

func (pipeline *Pipeline) handleGenerator(ctx context.Context, generatorCfg config.Generator) ([]byte, error) {
	var gen generator.Generator
	// TODO: Move to generator package
	if generatorCfg.Import != nil {
		dir := path.Dir(pipeline.forgeFile)
		data, err := os.ReadFile(path.Join(dir, generatorCfg.Import.Path))
		if err != nil {
			return nil, fmt.Errorf("error importing pipeline: %w", err)
		}
		subPipelineCfg, err := config.Parse(data)
		if err != nil {
			return nil, fmt.Errorf("error parsing pipeline: %w", err)
		}

		importVars := make(map[string][]byte)
		for i, importVar := range generatorCfg.Import.Vars {
			if importVar.Name == "" {
				return nil, fmt.Errorf("vars[%d]: import variable name cannot be empty", i)
			}
			ref, err := pipeline.referenceStore.GetReference(importVar.Reference)
			if err != nil {
				return nil, fmt.Errorf("variable %q: error getting import variable reference: %w", importVar.Name, err)
			}
			varName := importVar.Name
			importVars[varName] = ref
		}

		gen = &Pipeline{
			forgeFile:      generatorCfg.Import.Path,
			config:         subPipelineCfg,
			referenceStore: reference.NewStore(importVars),
		}
	} else {
		dir := path.Dir(pipeline.forgeFile)
		var err error
		gen, err = generator.GlobalRegistry.GetGenerator(dir, pipeline.referenceStore, generatorCfg)
		if err != nil {
			return nil, fmt.Errorf("error getting generator: %w", err)
		}
	}
	return gen.Generate(ctx)
}
