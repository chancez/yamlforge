package pipeline

import (
	"bytes"
	"context"
	"encoding/json"
	"errors"
	"fmt"
	"io"
	"os"
	"path"
	"text/template"

	"github.com/chancez/yamlforge/pkg/config"
	"github.com/chancez/yamlforge/pkg/generator"
	"github.com/chancez/yamlforge/pkg/mapmerge"
	"github.com/chancez/yamlforge/pkg/reference"
	"gopkg.in/yaml.v3"
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
	var buf bytes.Buffer
	for _, stage := range pipeline.config.Pipeline {
		switch {
		case stage.Generator != nil:
			result, err := pipeline.handleGenerator(ctx, *stage.Generator)
			if err != nil {
				return nil, fmt.Errorf("error running generator %q: %w", stage.Name, err)
			}
			err = pipeline.referenceStore.AddReference(stage.Name, result)
			if err != nil {
				return nil, fmt.Errorf("error storing reference for generator %q: %w", stage.Name, err)
			}
		case stage.Output != nil:
			err := pipeline.handleOutput(*stage.Output, &buf)
			if err != nil {
				return nil, fmt.Errorf("error running output %q: %w", stage.Name, err)
			}
		}
	}

	return buf.Bytes(), nil
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
		merged := make(map[string]any)
		for _, input := range generatorCfg.Merge.Input {
			ref, err := pipeline.referenceStore.GetReference(input)
			if err != nil {
				return nil, fmt.Errorf("error getting reference: %w", err)
			}
			var m map[string]any
			err = yaml.Unmarshal(ref, &m)
			if err != nil {
				return nil, fmt.Errorf("error parsing reference as YAML: %w", err)
			}
			merged = mapmerge.Merge(merged, m)
		}
		out, err := yaml.Marshal(merged)
		if err != nil {
			return nil, fmt.Errorf("error while marshaling merged results to YAML: %w", err)
		}
		return out, nil
	case generatorCfg.GoTemplate != nil:
		var buf bytes.Buffer
		tpl := template.New("go-template-generator")
		res, err := pipeline.referenceStore.GetReference(generatorCfg.GoTemplate.Input)
		if err != nil {
			return nil, fmt.Errorf("error getting reference: %w", err)
		}
		tpl, err = tpl.Parse(string(res))
		if err != nil {
			return nil, fmt.Errorf("error parsing template: %w", err)
		}
		err = tpl.Execute(&buf, generatorCfg.GoTemplate.Vars)
		if err != nil {
			return nil, fmt.Errorf("error executing template: %w", err)
		}
		return buf.Bytes(), nil
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

func (pipeline *Pipeline) handleOutput(outputConf config.Output, out io.Writer) error {
	switch {
	case outputConf.YAML != nil:
		enc := yaml.NewEncoder(out)
		for _, input := range outputConf.YAML.Input {
			ref, err := pipeline.referenceStore.GetReference(input)
			if err != nil {
				return fmt.Errorf("error getting reference: %w", err)
			}
			// assume that all refs are YAML for now
			dec := yaml.NewDecoder(bytes.NewBuffer(ref))
			var tmp map[string]any
			for {
				err = dec.Decode(&tmp)
				if err == io.EOF {
					break
				}
				if err != nil {
					return fmt.Errorf("error decoding reference output as YAML: %w", err)
				}
				err = enc.Encode(tmp)
				if err != nil {
					return fmt.Errorf("error writing YAML: %w", err)
				}
			}
		}
		err := enc.Close()
		if err != nil {
			return fmt.Errorf("error writing YAML: %w", err)
		}
	case outputConf.JSON != nil:
		enc := json.NewEncoder(out)
		for _, input := range outputConf.JSON.Input {
			ref, err := pipeline.referenceStore.GetReference(input)
			if err != nil {
				return fmt.Errorf("error getting reference: %w", err)
			}
			// assume that all refs are YAML for now
			dec := yaml.NewDecoder(bytes.NewBuffer(ref))
			var tmp map[string]any
			for {
				err = dec.Decode(&tmp)
				if err == io.EOF {
					break
				}
				if err != nil {
					return fmt.Errorf("error decoding reference output as YAML: %w", err)
				}
				err = enc.Encode(tmp)
				if err != nil {
					return fmt.Errorf("error writing JSON: %w", err)
				}
			}
		}
	default:
		return errors.New("invalid output, must configure an output type")
	}
	return nil
}
