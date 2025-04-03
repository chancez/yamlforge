package generator

import (
	"context"
	"encoding/json"
	"fmt"

	"github.com/chancez/yamlforge/pkg/config"
	"github.com/chancez/yamlforge/pkg/reference"
	jsonpatch "github.com/evanphx/json-patch/v5"
)

var _ Generator = (*JSONPatch)(nil)

type JSONPatch struct {
	dir      string
	cfg      config.JSONPatchGenerator
	refStore *reference.Store
}

func NewJSONPatch(dir string, cfg config.JSONPatchGenerator, refStore *reference.Store) *JSONPatch {
	return &JSONPatch{
		dir:      dir,
		cfg:      cfg,
		refStore: refStore,
	}
}

func (jp *JSONPatch) Generate(context.Context) ([]byte, error) {
	data, err := jp.refStore.GetReference(jp.dir, jp.cfg.Input)
	if err != nil {
		return nil, fmt.Errorf("error getting reference: %w", err)
	}

	configPatch := []byte(jp.cfg.Patch)

	// Try parsing the patch as YAML, and convert to JSON
	var parsedPatch any
	err = config.DecodeYAML([]byte(jp.cfg.Patch), &parsedPatch)
	if err == nil {
		configPatch, err = json.Marshal(parsedPatch)
		if err != nil {
			return nil, fmt.Errorf("unable to convert patch to JSON: %w", err)
		}
	}

	if jp.cfg.Merge {
		modified, err := jsonpatch.MergeMergePatches(data, configPatch)
		if err != nil {
			return nil, fmt.Errorf("error applying patch: %w", err)
		}
		return modified, nil
	}

	patch, err := jsonpatch.DecodePatch(configPatch)
	if err != nil {
		return nil, fmt.Errorf("error parsing JSON patch: %w", err)
	}

	modified, err := patch.ApplyWithOptions(data, &jsonpatch.ApplyOptions{
		EnsurePathExistsOnAdd:  true,
		SupportNegativeIndices: true,
	})
	if err != nil {
		return nil, fmt.Errorf("error applying patch: %w", err)
	}

	return modified, nil
}
