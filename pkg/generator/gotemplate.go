package generator

import (
	"bytes"
	"context"
	"errors"
	"fmt"
	"text/template"

	"github.com/Masterminds/sprig/v3"
	"github.com/chancez/yamlforge/pkg/config"
)

var extraTemplateFuncs = template.FuncMap{
	"required": func(warn string, val any) (any, error) {
		if val == nil {
			return val, errors.New(warn)
		} else if _, ok := val.(string); ok {
			if val == "" {
				return val, errors.New(warn)
			}
		}
		return val, nil
	},
}

var _ Generator = (*GoTemplate)(nil)

type GoTemplate struct {
	dir      string
	cfg      config.GoTemplateGenerator
	refStore *Store
}

func NewGoTemplate(dir string, cfg config.GoTemplateGenerator, refStore *Store) *GoTemplate {
	return &GoTemplate{
		dir:      dir,
		cfg:      cfg,
		refStore: refStore,
	}
}

func (gt *GoTemplate) Generate(_ context.Context) (*Result, error) {
	var buf bytes.Buffer
	tpl := template.New("go-template-generator").Option("missingkey=error").Funcs(sprig.FuncMap()).Funcs(extraTemplateFuncs)
	val, err := gt.refStore.GetStringValue(gt.dir, gt.cfg.Template)
	if err != nil {
		return nil, fmt.Errorf("error getting value for 'template': %w", err)
	}
	tpl, err = tpl.Parse(val)
	if err != nil {
		return nil, fmt.Errorf("error parsing template: %w", err)
	}

	vars := make(map[string]any)
	for name, ref := range gt.cfg.Vars {
		if name == "" {
			return nil, fmt.Errorf("vars: variable name cannot be empty")
		}
		v, err := gt.refStore.GetAnyValue(gt.dir, ref)
		if err != nil {
			return nil, fmt.Errorf("variable %q: error getting value: %w", name, err)
		}
		varVal := v.Output
		// Convert bytes to string when using with templates.
		if bv, ok := varVal.([]byte); ok {
			varVal = string(bv)
		}
		vars[name] = varVal
	}

	err = tpl.Execute(&buf, vars)
	if err != nil {
		return nil, fmt.Errorf("error executing template: %w", err)
	}
	return &Result{Output: buf.Bytes()}, nil
}
