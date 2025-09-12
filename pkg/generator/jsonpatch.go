package generator

import (
	"context"
	"encoding/json"
	"fmt"

	"github.com/chancez/yamlforge/pkg/config"
	jsonpatch "github.com/evanphx/json-patch/v5"
)

var _ Generator = (*JSONPatch)(nil)

type JSONPatch struct {
	dir      string
	cfg      config.JSONPatchGenerator
	refStore *Store
}

func NewJSONPatch(dir string, cfg config.JSONPatchGenerator, refStore *Store) *JSONPatch {
	return &JSONPatch{
		dir:      dir,
		cfg:      cfg,
		refStore: refStore,
	}
}

func (jp *JSONPatch) Generate(context.Context) (*Result, error) {
	input, err := jp.refStore.GetStringValue(jp.dir, jp.cfg.Input)
	if err != nil {
		return nil, fmt.Errorf("error getting input: %w", err)
	}

	patch, err := jp.refStore.GetStringValue(jp.dir, jp.cfg.Patch)
	if err != nil {
		return nil, fmt.Errorf("error getting patch: %w", err)
	}

	merge, err := jp.refStore.GetBoolValue(jp.dir, jp.cfg.Merge)
	if err != nil {
		return nil, fmt.Errorf("error getting merge: %w", err)
	}

	configPatch := []byte(patch)

	// Try parsing the patch as YAML, and convert to JSON
	var parsedPatch any
	err = config.DecodeYAML([]byte(patch), &parsedPatch)
	if err == nil {
		configPatch, err = json.Marshal(parsedPatch)
		if err != nil {
			return nil, fmt.Errorf("unable to convert patch to JSON: %w", err)
		}
	}

	inputBytes := []byte(input)

	if merge {
		modified, err := jsonpatch.MergeMergePatches(inputBytes, configPatch)
		if err != nil {
			return nil, fmt.Errorf("error applying patch: %w", err)
		}
		return &Result{Output: modified, Format: "json"}, nil
	}

	decodedPatch, err := jsonpatch.DecodePatch(configPatch)
	if err != nil {
		return nil, fmt.Errorf("error parsing JSON patch: %w", err)
	}

	modified, err := decodedPatch.ApplyWithOptions(inputBytes, &jsonpatch.ApplyOptions{
		EnsurePathExistsOnAdd:  true,
		SupportNegativeIndices: true,
	})
	if err != nil {
		return nil, fmt.Errorf("error applying patch: %w", err)
	}

	return &Result{Output: modified, Format: "json"}, nil
}
