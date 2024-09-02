package generator

import (
	"context"
	"fmt"
	"os"
	"path"

	"github.com/chancez/yamlforge/pkg/config"
	"github.com/chancez/yamlforge/pkg/reference"
)

var _ Generator = (*Import)(nil)

func init() {
	Register("import", config.ImportGenerator{}, func(dir string, cfg any, refStore *reference.Store) Generator {
		return NewImport(dir, cfg.(config.ImportGenerator), refStore)
	})
}

type Import struct {
	// Path lookups are relative to the dir specified.
	dir      string
	cfg      config.ImportGenerator
	refStore *reference.Store
}

func NewImport(dir string, cfg config.ImportGenerator, refStore *reference.Store) *Import {
	return &Import{
		dir:      dir,
		cfg:      cfg,
		refStore: refStore,
	}
}

func (imp *Import) Generate(ctx context.Context) ([]byte, error) {
	data, err := os.ReadFile(path.Join(imp.dir, imp.cfg.Path))
	if err != nil {
		return nil, fmt.Errorf("error importing pipeline: %w", err)
	}
	subPipelineCfg, err := config.Parse(data)
	if err != nil {
		return nil, fmt.Errorf("error parsing pipeline: %w", err)
	}

	importVars := make(map[string][]byte)
	for i, importVar := range imp.cfg.Vars {
		if importVar == nil {
			return nil, fmt.Errorf("vars[%d]: import variable must be specified", i)
		}
		if importVar.Name == "" {
			return nil, fmt.Errorf("vars[%d]: import variable name cannot be empty", i)
		}
		ref, err := imp.refStore.GetReference(imp.dir, importVar.Value)
		if err != nil {
			return nil, fmt.Errorf("variable %q: error getting import variable reference: %w", importVar.Name, err)
		}
		varName := importVar.Name
		importVars[varName] = ref
	}

	newStore := reference.NewStore(importVars)
	subPipeline := NewPipeline(path.Dir(imp.cfg.Path), subPipelineCfg.PipelineGenerator, newStore)
	return subPipeline.Generate(ctx)
}
