package config

import (
	"bytes"
	"errors"
	"fmt"
	"os"

	"gopkg.in/yaml.v3"
)

func ParseFile(forgeFile string) (Config, error) {
	data, err := os.ReadFile(forgeFile)
	if err != nil {
		return Config{}, fmt.Errorf("error parsing config: %w", err)
	}
	return Parse(data)
}

func Parse(data []byte) (Config, error) {
	var cfg Config
	dec := yaml.NewDecoder(bytes.NewBuffer(data))
	dec.KnownFields(true)
	err := dec.Decode(&cfg)
	if err != nil {
		return Config{}, fmt.Errorf("error parsing config: %w", err)
	}

	if cfg.Generator != nil && len(cfg.Pipeline) != 0 {
		return Config{}, errors.New("validation error: cannot set both 'pipeline' and 'generator' options")
	}

	if cfg.Generator != nil {
		err := validateGenerators(*cfg.Generator)
		if err != nil {
			return Config{}, fmt.Errorf("validation error: %w", err)
		}
	}

	generatorPositions := make(map[string]int)
	for pos, gen := range cfg.Pipeline {
		if gen.Name == "" {
			return Config{}, fmt.Errorf("validation error: pipeline[%d], generator is missing a name", pos)
		}
		if existingPos, exists := generatorPositions[gen.Name]; exists {
			return Config{}, fmt.Errorf("validation error: pipeline[%d], generator %q already exists at pipeline[%d]", pos, gen.Name, existingPos)
		}
		err := validateGenerators(gen)
		if err != nil {
			return Config{}, fmt.Errorf("validation error: pipeline[%d] %s: %w", pos, gen.Name, err)
		}
		generatorPositions[gen.Name] = pos
	}

	return cfg, nil
}

func validateGenerators(generatorCfg Generator) error {
	count := 0
	if generatorCfg.File != nil {
		count++
	}
	if generatorCfg.Value != nil {
		count++
	}
	if generatorCfg.Exec != nil {
		count++
	}
	if generatorCfg.Helm != nil {
		count++
	}
	if generatorCfg.Kustomize != nil {
		count++
	}
	if generatorCfg.Merge != nil {
		count++
	}
	if generatorCfg.GoTemplate != nil {
		count++
	}
	if generatorCfg.Pipeline != nil {
		count++
	}
	if generatorCfg.JQ != nil {
		count++
	}
	if generatorCfg.CELFilter != nil {
		count++
	}
	if generatorCfg.YAML != nil {
		count++
	}
	if generatorCfg.JSON != nil {
		count++
	}
	if count == 0 {
		return fmt.Errorf("generator not configured")
	}
	if count > 1 {
		return fmt.Errorf("invalid configuration, cannot specify multiple generators in the same generator config")
	}
	return nil
}
