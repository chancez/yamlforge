package yamlforge

import (
	"bytes"
	"context"
	"encoding/json"
	"errors"
	"fmt"
	"html/template"
	"io"
	"os"
	"os/exec"
	"path"

	"gopkg.in/yaml.v3"
)

type pipelineState struct {
	forgeFile string
	// map from an stage.Name to it's results
	references map[string][]byte
	// map a variable name to it's value
	vars map[string][]byte
}

func Generate(ctx context.Context, forgeFile string, vars map[string]string) ([]byte, error) {
	cfg, err := Parse(forgeFile)
	if err != nil {
		return nil, fmt.Errorf("error parsing pipeline %s: %w", forgeFile, err)
	}

	state := pipelineState{
		forgeFile:  forgeFile,
		references: make(map[string][]byte),
		vars:       make(map[string][]byte),
	}

	for varName, varVal := range vars {
		state.vars[varName] = []byte(varVal)
	}

	return state.generate(ctx, cfg)
}

func (state *pipelineState) generate(ctx context.Context, cfg *Config) ([]byte, error) {
	var buf bytes.Buffer
	for _, stage := range cfg.Pipeline {
		switch {
		case stage.Generator != nil:
			result, err := state.handleGenerator(ctx, stage.Generator)
			if err != nil {
				return nil, fmt.Errorf("error running generator %q: %w", stage.Name, err)
			}
			state.references[stage.Name] = result
		case stage.Transformer != nil:
			result, err := state.handleTransformer(ctx, stage.Transformer)
			if err != nil {
				return nil, fmt.Errorf("error running transformer %q: %w", stage.Name, err)
			}
			state.references[stage.Name] = result
		case stage.Output != nil:
			err := state.handleOutput(stage.Output, &buf)
			if err != nil {
				return nil, fmt.Errorf("error running output %q: %w", stage.Name, err)
			}
		}
	}

	return buf.Bytes(), nil
}

func (state *pipelineState) readFile(filePath string) ([]byte, error) {
	return os.ReadFile(path.Join(path.Dir(state.forgeFile), filePath))
}

func (state *pipelineState) handleGenerator(ctx context.Context, generator *Generator) ([]byte, error) {
	switch {
	case generator.File != nil:
		return state.readFile(generator.File.Path)
	case generator.Exec != nil:
		var buf bytes.Buffer
		cmd := exec.Command(generator.Exec.Command, generator.Exec.Args...)
		cmd.Stderr = os.Stderr
		cmd.Stdout = &buf
		err := cmd.Run()
		if err != nil {
			return nil, err
		}
		return buf.Bytes(), nil
	case generator.Helm != nil:
		var buf bytes.Buffer
		templateArgs := []string{
			"template",
			generator.Helm.ReleaseName,
			generator.Helm.Chart,
		}
		if generator.Helm.Version != "" {
			templateArgs = append(templateArgs, "--version", generator.Helm.Version)
		}
		if generator.Helm.Repo != "" {
			templateArgs = append(templateArgs, "--repo", generator.Helm.Repo)
		}
		if generator.Helm.Namespace != "" {
			templateArgs = append(templateArgs, "--namespace", generator.Helm.Namespace)
		}
		if len(generator.Helm.APIVersions) != 0 {
			for _, apiVersion := range generator.Helm.APIVersions {
				templateArgs = append(templateArgs, "--api-versions", apiVersion)
			}
		}
		var refs [][]byte
		for _, input := range generator.Helm.Values {
			ref, err := state.getReference(input)
			if err != nil {
				return nil, fmt.Errorf("error getting reference: %w", err)
			}
			refs = append(refs, ref)
		}

		tmpDir, err := os.MkdirTemp(os.TempDir(), "yfg-helm-generator-")
		if err != nil {
			return nil, fmt.Errorf("error creating temporary directory: %w", err)
		}
		defer func() {
			os.RemoveAll(tmpDir)
		}()

		for i, ref := range refs {
			refPath := path.Join(tmpDir, fmt.Sprintf("ref-%d-values.yaml", i))
			err = os.WriteFile(refPath, ref, 0400)
			if err != nil {
				return nil, fmt.Errorf("error writing helm values to %q: %w", refPath, err)
			}
			templateArgs = append(templateArgs, "--values", refPath)
		}

		cmd := exec.Command("helm", templateArgs...)
		cmd.Stderr = os.Stderr
		cmd.Stdout = &buf
		err = cmd.Run()
		if err != nil {
			return nil, err
		}
		return buf.Bytes(), nil
	default:
		return nil, errors.New("invalid generator, no generator specified")
	}
}

func (state *pipelineState) handleTransformer(ctx context.Context, transformer *Transformer) ([]byte, error) {
	switch {
	case transformer.Merge != nil:
		merged := make(map[string]any)
		for _, input := range transformer.Merge.Input {
			ref, err := state.getReference(input)
			if err != nil {
				return nil, fmt.Errorf("error getting reference: %w", err)
			}
			var m map[string]any
			err = yaml.Unmarshal(ref, &m)
			if err != nil {
				return nil, fmt.Errorf("error parsing reference as YAML: %w", err)
			}
			merged = Merge(merged, m)
		}
		out, err := yaml.Marshal(merged)
		if err != nil {
			return nil, fmt.Errorf("error while marshaling merged results to YAML: %w", err)
		}
		return out, nil
	case transformer.GoTemplate != nil:
		var buf bytes.Buffer
		tpl := template.New("go-template-transformer")
		res, err := state.getReference(transformer.GoTemplate.Input)
		if err != nil {
			return nil, fmt.Errorf("error getting reference: %w", err)
		}
		tpl, err = tpl.Parse(string(res))
		if err != nil {
			return nil, fmt.Errorf("error parsing template: %w", err)
		}
		err = tpl.Execute(&buf, transformer.GoTemplate.Vars)
		if err != nil {
			return nil, fmt.Errorf("error executing template: %w", err)
		}
		return buf.Bytes(), nil
	case transformer.Import != nil:
		data, err := state.readFile(transformer.Import.Path)
		if err != nil {
			return nil, fmt.Errorf("error importing transformer: %w", err)
		}
		transformerCfg, err := parse(data)
		if err != nil {
			return nil, fmt.Errorf("error parsing transformer: %w", err)
		}

		importVars := make(map[string][]byte)
		for i, importVar := range transformer.Import.Vars {
			if importVar.Name == "" {
				return nil, fmt.Errorf("vars[%d]: import variable name cannot be empty", i)
			}
			ref, err := state.getReference(importVar.Reference)
			if err != nil {
				return nil, fmt.Errorf("variable %q: error getting import variable reference: %w", importVar.Name, err)
			}
			varName := importVar.Name
			importVars[varName] = ref
		}

		transformerCtx := pipelineState{
			forgeFile:  transformer.Import.Path,
			vars:       importVars,
			references: make(map[string][]byte),
		}
		result, err := transformerCtx.generate(ctx, transformerCfg)
		if err != nil {
			return nil, fmt.Errorf("error executing transformer: %w", err)
		}
		return result, nil
	default:
		return nil, errors.New("invalid transformer, no transformer specified")
	}
}

func (state *pipelineState) handleOutput(outputConf *Output, out io.Writer) error {
	switch {
	case outputConf.YAML != nil:
		enc := yaml.NewEncoder(out)
		for _, input := range outputConf.YAML.Input {
			ref, err := state.getReference(input)
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
			ref, err := state.getReference(input)
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

func (state *pipelineState) getReference(ref Reference) ([]byte, error) {
	switch {
	case ref.Var != nil:
		varName := *ref.Var
		res, ok := state.vars[varName]
		if !ok {
			return nil, fmt.Errorf("could not find variable %q", varName)
		}
		return res, nil
	case ref.Ref != nil:
		refName := *ref.Ref
		res, ok := state.references[refName]
		if !ok {
			return nil, fmt.Errorf("could not find reference %q", refName)
		}
		return res, nil
	case ref.File != nil:
		return os.ReadFile(*ref.File)
	case ref.Literal != nil:
		return yaml.Marshal(ref.Literal)
	default:
		return nil, errors.New("invalid reference, must specify a reference type")
	}
}

func Parse(forgeFile string) (*Config, error) {
	data, err := os.ReadFile(forgeFile)
	if err != nil {
		return nil, fmt.Errorf("error parsing config: %w", err)
	}
	return parse(data)
}

func parse(data []byte) (*Config, error) {
	var cfg Config
	err := yaml.Unmarshal(data, &cfg)
	if err != nil {
		return nil, fmt.Errorf("error parsing config: %w", err)
	}

	stagePositions := make(map[string]int)
	for pos, stage := range cfg.Pipeline {
		if stage.Name == "" {
			return nil, fmt.Errorf("error parsing: pipeline[%d], stage is missing a name", pos)
		}
		if existingPos, exists := stagePositions[stage.Name]; exists {
			return nil, fmt.Errorf("error parsing: pipeline[%d], stage %q already exists at pipeline[%d]", pos, stage.Name, existingPos)
		}
		stagePositions[stage.Name] = pos
	}
	return &cfg, nil
}
