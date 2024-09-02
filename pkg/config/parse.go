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

	stagePositions := make(map[string]int)
	for pos, stage := range cfg.Pipeline {
		if stage.Name == "" {
			return Config{}, fmt.Errorf("error parsing: pipeline[%d], stage is missing a name", pos)
		}
		if existingPos, exists := stagePositions[stage.Name]; exists {
			return Config{}, fmt.Errorf("error parsing: pipeline[%d], stage %q already exists at pipeline[%d]", pos, stage.Name, existingPos)
		}
		stagePositions[stage.Name] = pos
	}
	return cfg, nil
}
