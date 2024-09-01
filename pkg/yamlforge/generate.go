package yamlforge

import (
	"context"
	"fmt"

	"github.com/chancez/yamlforge/pkg/config"
	"github.com/chancez/yamlforge/pkg/pipeline"
	"github.com/chancez/yamlforge/pkg/reference"
)

func Generate(ctx context.Context, forgeFile string, vars map[string][]byte) ([]byte, error) {
	cfg, err := config.ParseFile(forgeFile)
	if err != nil {
		return nil, fmt.Errorf("error parsing pipeline %s: %w", forgeFile, err)
	}

	refStore := reference.NewStore(vars)
	state := pipeline.NewPipeline(forgeFile, cfg, refStore)
	return state.Generate(ctx)
}
