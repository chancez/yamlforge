package yamlforge

type Config struct {
	Pipeline []*Stage `yaml:"pipeline" json:"pipeline"`
}

type Stage struct {
	Name        string       `yaml:"name,omitempty" json:"name,omitempty"`
	Generator   *Generator   `yaml:"generator,omitempty" json:"generator,omitempty"`
	Transformer *Transformer `yaml:"transformer,omitempty" json:"transformer,omitempty"`
	Output      *Output      `yaml:"output,omitempty" json:"output,omitempty"`
}

type Generator struct {
	File *FileGenerator `yaml:"file,omitempty" json:"file,omitempty"`
	Exec *ExecGenerator `yaml:"exec,omitempty" json:"exec,omitempty"`
	Helm *HelmGenerator `yaml:"helm,omitempty" json:"helm,omitempty"`
}

type FileGenerator struct {
	Path string `yaml:"path,omitempty" json:"path,omitempty"`
}

type ExecGenerator struct {
	Command string   `yaml:"command,omitempty" json:"command,omitempty"`
	Args    []string `yaml:"args,omitempty" json:"args,omitempty"`
}

type HelmGenerator struct {
	ReleaseName string      `yaml:"releaseName,omitempty" json:"releaseName,omitempty"`
	Chart       string      `yaml:"chart,omitempty" json:"chart,omitempty"`
	Version     string      `yaml:"version,omitempty" json:"version,omitempty"`
	Repo        string      `yaml:"repo,omitempty" json:"repo,omitempty"`
	Namespace   string      `yaml:"namespace,omitempty" json:"namespace,omitempty"`
	APIVersions []string    `yaml:"apiVersions,omitempty" json:"apiVersions,omitempty"`
	Values      []Reference `yaml:"values,omitempty" json:"values,omitempty"`
}

type Transformer struct {
	Merge      *MergeTransformer      `yaml:"merge,omitempty" json:"merge,omitempty"`
	GoTemplate *GoTemplateTransformer `yaml:"gotemplate,omitempty" json:"gotemplate,omitempty"`
	Import     *ImportTransformer     `yaml:"import,omitempty" json:"import,omitempty"`
}

type MergeTransformer struct {
	Input []Reference `yaml:"input,omitempty" json:"input,omitempty"`
}

type GoTemplateTransformer struct {
	Input Reference      `yaml:"input,omitempty" json:"input,omitempty"`
	Vars  map[string]any `yaml:"vars,omitempty" json:"vars,omitempty"`
}

type ImportTransformer struct {
	Path string           `yaml:"path,omitempty" json:"path,omitempty"`
	Vars []ImportVariable `yaml:"vars,omitempty" json:"vars,omitempty"`
}

type ImportVariable struct {
	Name      string `yaml:"name,omitempty" json:"name,omitempty"`
	Reference `yaml:",inline"`
}

type Output struct {
	YAML *YAMLOutput `yaml:"yaml,omitempty" json:"yaml,omitempty"`
	JSON *JSONOutput `yaml:"json,omitempty" json:"json,omitempty"`
}

type YAMLOutput struct {
	Input []Reference `yaml:"input,omitempty" json:"input,omitempty"`
}

type JSONOutput struct {
	Input []Reference `yaml:"input,omitempty" json:"input,omitempty"`
}

type Reference struct {
	Var     *string        `yaml:"var,omitempty" json:"var,omitempty"`
	Ref     *string        `yaml:"ref,omitempty" json:"ref,omitempty"`
	File    *string        `yaml:"file,omitempty" json:"file,omitempty"`
	Literal map[string]any `yaml:"literal,omitempty" json:"literal,omitempty"`
}
