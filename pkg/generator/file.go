package generator

import (
	"context"
	"fmt"
	"os"
	"path"

	"github.com/chancez/yamlforge/pkg/config"
	"github.com/chancez/yamlforge/pkg/reference"
)

var _ Generator = (*File)(nil)

func init() {
	Register("file", config.FileGenerator{}, func(dir string, cfg any, refStore *reference.Store) Generator {
		return NewFile(dir, cfg.(config.FileGenerator))
	})
}

type File struct {
	// Path lookups are relative to the dir specified.
	dir string
	cfg config.FileGenerator
}

func NewFile(dir string, cfg config.FileGenerator) *File {
	return &File{
		dir: dir,
		cfg: cfg,
	}
}

func (f *File) Generate(context.Context) ([]byte, error) {
	data, err := os.ReadFile(path.Join(f.dir, f.cfg.Path))
	if err != nil {
		return nil, fmt.Errorf("error reading %q: %w", f.cfg.Path, err)
	}
	return data, nil
}
