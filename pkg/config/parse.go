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
		return Config{}, errors.New("cannot set both 'pipeline' and 'generator' options")
	}

	if cfg.Generator != nil {
		err := validateGenerators(*cfg.Generator)
		if err != nil {
			return Config{}, fmt.Errorf("error parsing: %w", err)
		}
	}

	generatorPositions := make(map[string]int)
	for pos, gen := range cfg.Pipeline {
		if gen.Name == "" {
			return Config{}, fmt.Errorf("error parsing: pipeline[%d], generator is missing a name", pos)
		}
		if existingPos, exists := generatorPositions[gen.Name]; exists {
			return Config{}, fmt.Errorf("error parsing: pipeline[%d], generator %q already exists at pipeline[%d]", pos, gen.Name, existingPos)
		}
		err := validateGenerators(gen)
		if err != nil {
			return Config{}, fmt.Errorf("error parsing: pipeline[%d] %s: %w", pos, gen.Name, err)
		}
		generatorPositions[gen.Name] = pos
	}

	return cfg, nil
}

func validateGenerators(generatorCfg Generator) error {
	count := 0
	switch {
	case generatorCfg.File != nil:
		count++
		fallthrough
	case generatorCfg.Exec != nil:
		count++
		fallthrough
	case generatorCfg.Helm != nil:
		count++
		fallthrough
	case generatorCfg.Kustomize != nil:
		count++
		fallthrough
	case generatorCfg.Merge != nil:
		count++
		fallthrough
	case generatorCfg.GoTemplate != nil:
		count++
		fallthrough
	case generatorCfg.Import != nil:
		count++
		fallthrough
	case generatorCfg.JQ != nil:
		count++
		fallthrough
	case generatorCfg.CELFilter != nil:
		count++
		fallthrough
	case generatorCfg.YAML != nil:
		count++
		fallthrough
	case generatorCfg.JSON != nil:
		count++
	}
	if count == 0 {
		return fmt.Errorf("generator not configured")
	}
	if count > 0 {
		return fmt.Errorf("invalid configuration, cannot specify multiple generators in the same generator config")
	}
	return nil
}
