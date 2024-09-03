package config

// Config defines a yamlforge configuration.
// It allows defining pipelines of generators that run in sequence and produce
// an (typically YAML) output.
type Config struct {
	PipelineGenerator `yaml:",inline" json:",inline"`
}

type Generator struct {
	Name       string               `yaml:"name" json:"name"`
	File       *FileGenerator       `yaml:"file,omitempty" json:"file,omitempty" jsonschema:"oneof_required=file"`
	Exec       *ExecGenerator       `yaml:"exec,omitempty" json:"exec,omitempty" jsonschema:"oneof_required=exec"`
	Helm       *HelmGenerator       `yaml:"helm,omitempty" json:"helm,omitempty" jsonschema:"oneof_required=helm"`
	Kustomize  *KustomizeGenerator  `yaml:"kustomize,omitempty" json:"kustomize,omitempty" jsonschema:"oneof_required=kustomize"`
	Merge      *MergeGenerator      `yaml:"merge,omitempty" json:"merge,omitempty" jsonschema:"oneof_required=merge"`
	GoTemplate *GoTemplateGenerator `yaml:"gotemplate,omitempty" json:"gotemplate,omitempty" jsonschema:"oneof_required=gotemplate"`
	Import     *ImportGenerator     `yaml:"import,omitempty" json:"import,omitempty" jsonschema:"oneof_required=import"`
	JQ         *JQGenerator         `yaml:"jq,omitempty" json:"jq,omitempty" jsonschema:"oneof_required=jq"`
	YAML       *YAMLGenerator       `yaml:"yaml,omitempty" json:"yaml,omitempty" jsonschema:"oneof_required=yaml"`
	JSON       *JSONGenerator       `yaml:"json,omitempty" json:"json,omitempty" jsonschema:"oneof_required=json"`
}

type FileGenerator struct {
	Path string `yaml:"path,omitempty" json:"path,omitempty"`
}

type ExecGenerator struct {
	Command string   `yaml:"command,omitempty" json:"command,omitempty"`
	Args    []string `yaml:"args,omitempty" json:"args,omitempty"`
}

type HelmGenerator struct {
	ReleaseName string   `yaml:"releaseName,omitempty" json:"releaseName"`
	Chart       string   `yaml:"chart,omitempty" json:"chart"`
	Version     string   `yaml:"version,omitempty" json:"version,omitempty"`
	Repo        string   `yaml:"repo,omitempty" json:"repo,omitempty"`
	Namespace   string   `yaml:"namespace,omitempty" json:"namespace,omitempty"`
	APIVersions []string `yaml:"apiVersions,omitempty" json:"apiVersions,omitempty"`
	Values      []Value  `yaml:"values,omitempty" json:"values,omitempty"`
}

type KustomizeGenerator struct {
	Dir        string `yaml:"dir,omitempty" json:"dir,omitempty" jsonschema:"oneof_required=dir"`
	URL        string `yaml:"url,omitempty" json:"url,omitempty"  jsonschema:"oneof_required=url"`
	EnableHelm bool   `yaml:"enableHelm,omitempty" json:"enableHelm,omitempty"`
}

type MergeGenerator struct {
	Name  string  `yaml:"name" json:"name"`
	Input []Value `yaml:"input,omitempty" json:"input,omitempty"`
}

type GoTemplateGenerator struct {
	Template *Value           `yaml:"template,omitempty" json:"template,omitempty"`
	Vars     map[string]any   `yaml:"vars,omitempty" json:"vars,omitempty"`
	RefVars  map[string]Value `yaml:"refVars,omitempty" json:"refVars,omitempty"`
}

type ImportGenerator struct {
	Path string           `yaml:"path,omitempty" json:"path,omitempty"`
	Vars []*NamedVariable `yaml:"vars,omitempty" json:"vars,omitempty"`
}

type NamedVariable struct {
	Name  string `yaml:"name,omitempty" json:"name"`
	Value `yaml:",inline" json:",inline"`
}

type JQGenerator struct {
	Expr     string `yaml:"expr,omitempty" json:"expr,omitempty" jsonschema:"oneof_required=expr"`
	ExprFile string `yaml:"exprFile,omitempty" json:"exprFile,omitempty" jsonschema:"oneof_required=exprFile"`
	Input    *Value `yaml:"input,omitempty" json:"input"`
	Slurp    bool   `yaml:"slurp,omitempty" json:"slurp,omitempty"`
}

type YAMLGenerator struct {
	Input []Value `yaml:"input,omitempty" json:"input"`
}

type JSONGenerator struct {
	Input []Value `yaml:"input,omitempty" json:"input"`
}

// PipelineGenerator is a generator which executes other generators.
// Currently it is only available through the top level configuration.
type PipelineGenerator struct {
	// Pipeline is a list of generators to run. Generators can reference the output of previous generators using their name in any Value refs.
	Pipeline []Generator `yaml:"pipeline,omitempty" json:"pipeline,omitempty" jsonschema:"oneof_required=pipeline"`
	// Generator is a single generator, for simple use-cases that do not require a full pipeline.
	Generator *Generator `yaml:"generator,omitempty" json:"generator,omitempty" jsonschema:"oneof_required=generator"`
}

type Value struct {
	Var   *string `yaml:"var,omitempty" json:"var,omitempty" jsonschema:"oneof_required=var"`
	Ref   *string `yaml:"ref,omitempty" json:"ref,omitempty" jsonschema:"oneof_required=ref"`
	File  *string `yaml:"file,omitempty" json:"file,omitempty" jsonschema:"oneof_required=file"`
	Value *any    `yaml:"value,omitempty" json:"value,omitempty" jsonschema:"oneof_required=value"`
}
