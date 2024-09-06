package generator

import (
	"bytes"
	"context"
	"errors"
	"os"
	"os/exec"
	"path"

	"github.com/chancez/yamlforge/pkg/config"
	"github.com/chancez/yamlforge/pkg/reference"
)

var _ Generator = (*Kustomize)(nil)

type Kustomize struct {
	dir      string
	cfg      config.KustomizeGenerator
	refStore *reference.Store
}

func NewKustomize(dir string, cfg config.KustomizeGenerator, refStore *reference.Store) *Kustomize {
	return &Kustomize{
		dir:      dir,
		cfg:      cfg,
		refStore: refStore,
	}
}

func (h *Kustomize) Generate(context.Context) ([]byte, error) {
	var buf bytes.Buffer
	kustomizeArgs := []string{
		"build",
	}
	switch {
	case h.cfg.Dir != "" && h.cfg.URL != "":
		return nil, errors.New("cannot specify both dir and url")
	case h.cfg.Dir != "":
		kustomizeDir := path.Join(h.dir, h.cfg.Dir)
		kustomizeArgs = append(kustomizeArgs, kustomizeDir)
	case h.cfg.URL != "":
		kustomizeArgs = append(kustomizeArgs, h.cfg.URL)
	}
	if h.cfg.EnableHelm {
		kustomizeArgs = append(kustomizeArgs, "--enable-helm")
	}
	cmd := exec.Command("kustomize", kustomizeArgs...)
	cmd.Stderr = os.Stderr
	cmd.Stdout = &buf
	err := cmd.Run()
	if err != nil {
		return nil, err
	}
	return buf.Bytes(), nil
}
