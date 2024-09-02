package generator

import (
	"bytes"
	"context"
	"fmt"
	"os"
	"os/exec"
	"path"

	"github.com/chancez/yamlforge/pkg/config"
	"github.com/chancez/yamlforge/pkg/reference"
)

var _ Generator = (*Helm)(nil)

func init() {
	Register("helm", config.HelmGenerator{}, func(dir string, cfg any, refStore *reference.Store) Generator {
		return NewHelm(dir, cfg.(config.HelmGenerator), refStore)
	})
}

type Helm struct {
	dir      string
	cfg      config.HelmGenerator
	refStore *reference.Store
}

func NewHelm(dir string, cfg config.HelmGenerator, refStore *reference.Store) *Helm {
	return &Helm{
		dir:      dir,
		cfg:      cfg,
		refStore: refStore,
	}
}

func (h *Helm) Generate(context.Context) ([]byte, error) {
	var buf bytes.Buffer
	templateArgs := []string{
		"template",
		h.cfg.ReleaseName,
		h.cfg.Chart,
	}
	if h.cfg.Version != "" {
		templateArgs = append(templateArgs, "--version", h.cfg.Version)
	}
	if h.cfg.Repo != "" {
		templateArgs = append(templateArgs, "--repo", h.cfg.Repo)
	}
	if h.cfg.Namespace != "" {
		templateArgs = append(templateArgs, "--namespace", h.cfg.Namespace)
	}
	if len(h.cfg.APIVersions) != 0 {
		for _, apiVersion := range h.cfg.APIVersions {
			templateArgs = append(templateArgs, "--api-versions", apiVersion)
		}
	}
	var refs [][]byte
	for _, input := range h.cfg.Values {
		ref, err := h.refStore.GetReference(h.dir, input)
		if err != nil {
			return nil, fmt.Errorf("error getting reference: %w", err)
		}
		refs = append(refs, ref)
	}

	tmpDir, err := os.MkdirTemp(os.TempDir(), "yfg-helm-generator-")
	if err != nil {
		return nil, fmt.Errorf("error creating temporary directory: %w", err)
	}
	defer func() {
		os.RemoveAll(tmpDir)
	}()

	for i, ref := range refs {
		refPath := path.Join(tmpDir, fmt.Sprintf("ref-%d-values.yaml", i))
		err = os.WriteFile(refPath, ref, 0400)
		if err != nil {
			return nil, fmt.Errorf("error writing helm values to %q: %w", refPath, err)
		}
		templateArgs = append(templateArgs, "--values", refPath)
	}

	cmd := exec.Command("helm", templateArgs...)
	cmd.Stderr = os.Stderr
	cmd.Stdout = &buf
	err = cmd.Run()
	if err != nil {
		return nil, err
	}
	return buf.Bytes(), nil
}
