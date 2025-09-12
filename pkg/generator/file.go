package generator

import (
	"context"
	"fmt"
	"os"
	"path"
	"path/filepath"

	"github.com/chancez/yamlforge/pkg/config"
)

var _ Generator = (*File)(nil)

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

func (f *File) Generate(context.Context) (*Result, error) {
	data, err := os.ReadFile(path.Join(f.dir, f.cfg.Path))
	if err != nil {
		return nil, fmt.Errorf("error reading %q: %w", f.cfg.Path, err)
	}
	format := formatFromFileName(f.cfg.Path)
	return &Result{Output: data, Format: format}, nil
}

func formatFromFileName(f string) string {
	ext := filepath.Ext(f)
	switch ext {
	case ".yml", ".yaml":
		return "yaml"
	case ".json":
		return "json"
	default:
		return ""
	}
}
