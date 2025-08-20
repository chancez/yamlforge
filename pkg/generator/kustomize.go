package generator

import (
	"bytes"
	"context"
	"errors"
	"os"
	"os/exec"
	"path"

	"github.com/chancez/yamlforge/pkg/config"
)

var _ Generator = (*Kustomize)(nil)

type Kustomize struct {
	dir      string
	cfg      config.KustomizeGenerator
	refStore *Store
}

func NewKustomize(dir string, cfg config.KustomizeGenerator, refStore *Store) *Kustomize {
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
	dir, err := h.refStore.GetStringValue(h.dir, h.cfg.Dir)
	if err != nil {
		return nil, err
	}
	u, err := h.refStore.GetStringValue(h.dir, h.cfg.URL)
	if err != nil {
		return nil, err
	}
	enableHelm, err := h.refStore.GetBoolValue(h.dir, h.cfg.EnableHelm)
	if err != nil {
		return nil, err
	}

	switch {
	case dir != "" && u != "":
		return nil, errors.New("cannot specify both dir and url")
	case dir != "":
		kustomizeDir := path.Join(h.dir, dir)
		kustomizeArgs = append(kustomizeArgs, kustomizeDir)
	case u != "":
		kustomizeArgs = append(kustomizeArgs, u)
	}
	if enableHelm {
		kustomizeArgs = append(kustomizeArgs, "--enable-helm")
	}
	cmd := exec.Command("kustomize", kustomizeArgs...)
	cmd.Stderr = os.Stderr
	cmd.Stdout = &buf
	err = cmd.Run()
	if err != nil {
		return nil, err
	}
	return buf.Bytes(), nil
}
