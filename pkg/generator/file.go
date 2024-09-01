package generator

import (
	"context"
	"os"
	"path"

	"github.com/chancez/yamlforge/pkg/config"
	"github.com/chancez/yamlforge/pkg/reference"
)

var _ Generator = (*File)(nil)

func init() {
	Register(config.FileGenerator{}, func(dir string, refStore *reference.Store, cfg any) Generator {
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
	return os.ReadFile(path.Join(f.dir, f.cfg.Path))
}
