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

func init() {
	Register(config.GoTemplateGenerator{}, func(dir string, refStore *reference.Store, cfg any) Generator {
		return NewGoTemplate(dir, cfg.(config.GoTemplateGenerator), refStore)
	})
}

type GoTemplate struct {
	dir      string
	cfg      config.GoTemplateGenerator
	refStore *reference.Store
}

func NewGoTemplate(dir string, cfg config.GoTemplateGenerator, refStore *reference.Store) *GoTemplate {
	return &GoTemplate{
		dir:      dir,
		cfg:      cfg,
		refStore: refStore,
	}
}

func (gt *GoTemplate) Generate(_ context.Context) ([]byte, error) {
	var buf bytes.Buffer
	tpl := template.New("go-template-generator")
	res, err := gt.refStore.GetReference(gt.dir, gt.cfg.Input)
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
