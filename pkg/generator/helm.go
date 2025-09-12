package generator

import (
	"bytes"
	"context"
	"fmt"
	"os"
	"os/exec"
	"path"

	"github.com/chancez/yamlforge/pkg/config"
)

var _ Generator = (*Helm)(nil)

type Helm struct {
	dir      string
	cfg      config.HelmGenerator
	refStore *Store
}

func NewHelm(dir string, cfg config.HelmGenerator, refStore *Store) *Helm {
	return &Helm{
		dir:      dir,
		cfg:      cfg,
		refStore: refStore,
	}
}

func (h *Helm) Generate(context.Context) (*Result, error) {
	var buf bytes.Buffer
	releaseName, err := h.refStore.GetStringValue(h.dir, h.cfg.ReleaseName)
	if err != nil {
		return nil, err
	}
	chart, err := h.refStore.GetStringValue(h.dir, h.cfg.Chart)
	if err != nil {
		return nil, err
	}
	version, err := h.refStore.GetStringValue(h.dir, h.cfg.Version)
	if err != nil {
		return nil, err
	}
	repo, err := h.refStore.GetStringValue(h.dir, h.cfg.Repo)
	if err != nil {
		return nil, err
	}
	namespace, err := h.refStore.GetStringValue(h.dir, h.cfg.Namespace)
	if err != nil {
		return nil, err
	}
	includeCRDs, err := h.refStore.GetBoolValue(h.dir, h.cfg.IncludeCRDs)
	if err != nil {
		return nil, err
	}

	apiVersions, err := h.refStore.GetStringValueList(h.dir, h.cfg.APIVersions)
	if err != nil {
		return nil, err
	}

	templateArgs := []string{
		"template",
		releaseName,
		chart,
	}
	if version != "" {
		templateArgs = append(templateArgs, "--version", version)
	}
	if repo != "" {
		templateArgs = append(templateArgs, "--repo", repo)
	}
	if namespace != "" {
		templateArgs = append(templateArgs, "--namespace", namespace)
	}
	if includeCRDs {
		templateArgs = append(templateArgs, "--include-crds")
	}
	for _, apiVersion := range apiVersions {
		templateArgs = append(templateArgs, "--api-versions", apiVersion)
	}
	var refs []string
	for _, input := range h.cfg.Values {
		ref, err := h.refStore.GetStringValue(h.dir, input)
		if err != nil {
			return nil, fmt.Errorf("error getting value: %w", err)
		}
		refs = append(refs, ref)
	}

	tmpDir, err := os.MkdirTemp(os.TempDir(), "yfg-helm-generator-")
	if err != nil {
		return nil, fmt.Errorf("error creating temporary directory: %w", err)
	}
	defer func() {
		// nolint:errcheck
		os.RemoveAll(tmpDir)
	}()

	for i, ref := range refs {
		refPath := path.Join(tmpDir, fmt.Sprintf("ref-%d-values.yaml", i))
		err = os.WriteFile(refPath, []byte(ref), 0400)
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
	return &Result{Output: buf.Bytes(), Format: "yaml"}, nil
}
