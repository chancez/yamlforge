package config

type Config struct {
	PipelineGenerator `yaml:",inline" json:",inline"`
}

type Stage struct {
	Name       string `yaml:"name" json:"name"`
	*Generator `yaml:",inline" json:",inline"`
}

type Generator struct {
	File       *FileGenerator       `yaml:"file,omitempty" json:"file,omitempty"`
	Exec       *ExecGenerator       `yaml:"exec,omitempty" json:"exec,omitempty"`
	Helm       *HelmGenerator       `yaml:"helm,omitempty" json:"helm,omitempty"`
	Kustomize  *KustomizeGenerator  `yaml:"kustomize,omitempty" json:"kustomize,omitempty"`
	Merge      *MergeGenerator      `yaml:"merge,omitempty" json:"merge,omitempty"`
	GoTemplate *GoTemplateGenerator `yaml:"gotemplate,omitempty" json:"gotemplate,omitempty"`
	Import     *ImportGenerator     `yaml:"import,omitempty" json:"import,omitempty"`
	JQ         *JQGenerator         `yaml:"jq,omitempty" json:"jq,omitempty"`
	YAML       *YAMLGenerator       `yaml:"yaml,omitempty" json:"yaml,omitempty"`
	JSON       *JSONGenerator       `yaml:"json,omitempty" json:"json,omitempty"`
}

type FileGenerator struct {
	Path string `yaml:"path,omitempty" json:"path,omitempty"`
}

type ExecGenerator struct {
	Command string   `yaml:"command,omitempty" json:"command,omitempty"`
	Args    []string `yaml:"args,omitempty" json:"args,omitempty"`
}

type HelmGenerator struct {
	ReleaseName string   `yaml:"releaseName,omitempty" json:"releaseName,omitempty"`
	Chart       string   `yaml:"chart,omitempty" json:"chart,omitempty"`
	Version     string   `yaml:"version,omitempty" json:"version,omitempty"`
	Repo        string   `yaml:"repo,omitempty" json:"repo,omitempty"`
	Namespace   string   `yaml:"namespace,omitempty" json:"namespace,omitempty"`
	APIVersions []string `yaml:"apiVersions,omitempty" json:"apiVersions,omitempty"`
	Values      []Value  `yaml:"values,omitempty" json:"values,omitempty"`
}

type KustomizeGenerator struct {
	Dir        string `yaml:"dir,omitempty" json:"dir,omitempty"`
	URL        string `yaml:"url,omitempty" json:"url,omitempty"`
	EnableHelm bool   `yaml:"enableHelm,omitempty" json:"enableHelm,omitempty"`
}

type MergeGenerator struct {
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
	Name  string `yaml:"name,omitempty" json:"name,omitempty"`
	Value `yaml:",inline" json:",inline"`
}

type JQGenerator struct {
	Expr     string `yaml:"expr,omitempty" json:"expr,omitempty"`
	ExprFile string `yaml:"exprFile,omitempty" json:"exprFile,omitempty"`
	Input    *Value `yaml:"input,omitempty" json:"input,omitempty"`
	Slurp    bool   `yaml:"slurp,omitempty" json:"slurp,omitempty"`
}

type YAMLGenerator struct {
	Input []Value `yaml:"input,omitempty" json:"input,omitempty"`
}

type JSONGenerator struct {
	Input []Value `yaml:"input,omitempty" json:"input,omitempty"`
}

type PipelineGenerator struct {
	Pipeline  []*Stage   `yaml:"pipeline,omitempty" json:"pipeline,omitempty"`
	Generator *Generator `yaml:"generator,omitempty" json:"generator,omitempty"`
}

type Value struct {
	Var   *string `yaml:"var,omitempty" json:"var,omitempty"`
	Ref   *string `yaml:"ref,omitempty" json:"ref,omitempty"`
	File  *string `yaml:"file,omitempty" json:"file,omitempty"`
	Value *any    `yaml:"value,omitempty" json:"value,omitempty"`
}
