package generator

import (
	"bytes"
	"context"
	"fmt"
	"text/template"

	"github.com/chancez/yamlforge/pkg/config"
	"github.com/chancez/yamlforge/pkg/reference"
)

var _ Generator = (*GoTemplate)(nil)

type GoTemplate struct {
	cfg      config.GoTemplateGenerator
	refStore *reference.Store
}

func NewGoTemplate(cfg config.GoTemplateGenerator, refStore *reference.Store) *GoTemplate {
	return &GoTemplate{
		cfg:      cfg,
		refStore: refStore,
	}
}

func (gt *GoTemplate) Generate(_ context.Context) ([]byte, error) {
	var buf bytes.Buffer
	tpl := template.New("go-template-generator")
	res, err := gt.refStore.GetReference(gt.cfg.Input)
	if err != nil {
		return nil, fmt.Errorf("error getting reference: %w", err)
	}
	tpl, err = tpl.Parse(string(res))
	if err != nil {
		return nil, fmt.Errorf("error parsing template: %w", err)
	}
	err = tpl.Execute(&buf, gt.cfg.Vars)
	if err != nil {
		return nil, fmt.Errorf("error executing template: %w", err)
	}
	return buf.Bytes(), nil
}
