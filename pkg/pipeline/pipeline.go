package pipeline

import (
	"context"
	"errors"
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

func (pipeline *Pipeline) readFile(filePath string) ([]byte, error) {
	return os.ReadFile(path.Join(path.Dir(pipeline.forgeFile), filePath))
}

func (pipeline *Pipeline) handleGenerator(ctx context.Context, generatorCfg config.Generator) ([]byte, error) {
	var gen generator.Generator
	switch {
	case generatorCfg.File != nil:
		gen = generator.NewFile(path.Dir(pipeline.forgeFile), *generatorCfg.File)
	case generatorCfg.Exec != nil:
		gen = generator.NewExec(*generatorCfg.Exec)
	case generatorCfg.Helm != nil:
		gen = generator.NewHelm(*generatorCfg.Helm, pipeline.referenceStore)
	case generatorCfg.Merge != nil:
		gen = generator.NewMerge(*generatorCfg.Merge, pipeline.referenceStore)
	case generatorCfg.GoTemplate != nil:
		gen = generator.NewGoTemplate(*generatorCfg.GoTemplate, pipeline.referenceStore)
	case generatorCfg.YAML != nil:
		gen = generator.NewYAML(*generatorCfg.YAML, pipeline.referenceStore)
	case generatorCfg.JSON != nil:
		gen = generator.NewJSON(*generatorCfg.JSON, pipeline.referenceStore)
	case generatorCfg.Import != nil:
		data, err := pipeline.readFile(generatorCfg.Import.Path)
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

		subPipeline := Pipeline{
			forgeFile:      generatorCfg.Import.Path,
			config:         subPipelineCfg,
			referenceStore: reference.NewStore(importVars),
		}
		result, err := subPipeline.Generate(ctx)
		if err != nil {
			return nil, fmt.Errorf("error executing pipeline: %w", err)
		}
		return result, nil
	default:
		return nil, errors.New("invalid generator, no generator specified")
	}
	return gen.Generate(ctx)
}
