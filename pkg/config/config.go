package config

// Config defines a yamlforge configuration.
// It allows defining pipelines of generators that run in sequence and produce
// an (typically YAML) output.
type Config struct {
	PipelineGenerator `yaml:",inline" json:",inline"`
}

// Generators execute some logic and produce output. Only one type of generator can be specified.
type Generator struct {
	// Name is the name of this generator which other generators can reference this generator's output by.
	Name string `yaml:"name" json:"name"`
	// File is a generator which reads files at the specified path and returns their output.
	File *FileGenerator `yaml:"file,omitempty" json:"file,omitempty" jsonschema:"oneof_required=file"`
	// Exec is a generator which execs the command specified and returns the stdout of the program.
	Exec *ExecGenerator `yaml:"exec,omitempty" json:"exec,omitempty" jsonschema:"oneof_required=exec"`
	// Helm is a generator which runs 'helm template' to render a Helm chart and returns the output.
	Helm *HelmGenerator `yaml:"helm,omitempty" json:"helm,omitempty" jsonschema:"oneof_required=helm"`
	// Kustomize is a generator which runs 'kustomize build' to render a Kustomization and returns the output.
	Kustomize *KustomizeGenerator `yaml:"kustomize,omitempty" json:"kustomize,omitempty" jsonschema:"oneof_required=kustomize"`
	// Merge is a generator which takes multiple inputs containing object-like data and deeply merges them together and returns the merged output.
	Merge *MergeGenerator `yaml:"merge,omitempty" json:"merge,omitempty" jsonschema:"oneof_required=merge"`
	// GoTemplate is a generator which renders Go 'text/template' templates and returns the output.
	GoTemplate *GoTemplateGenerator `yaml:"gotemplate,omitempty" json:"gotemplate,omitempty" jsonschema:"oneof_required=gotemplate"`
	// Import is a generator which imports pipelines from another file and returns the output.
	Import *ImportGenerator `yaml:"import,omitempty" json:"import,omitempty" jsonschema:"oneof_required=import"`
	// JQ is a generator which executes 'jq' and returns the output.
	JQ *JQGenerator `yaml:"jq,omitempty" json:"jq,omitempty" jsonschema:"oneof_required=jq"`
	// YAML is a generator which returns it's inputs as YAML.
	YAML *YAMLGenerator `yaml:"yaml,omitempty" json:"yaml,omitempty" jsonschema:"oneof_required=yaml"`
	// JSON is a generator which returns it's inputs as JSON.
	JSON *JSONGenerator `yaml:"json,omitempty" json:"json,omitempty" jsonschema:"oneof_required=json"`
}

// FileGenerator reads files at the specified path and returns their output.
type FileGenerator struct {
	// Path is the path relative to this pipeline file to read.
	Path string `yaml:"path,omitempty" json:"path,omitempty"`
}

// ExecGenerator execs the command specified and returns the stdout of the program.
type ExecGenerator struct {
	// Command is the command to execute.
	Command string `yaml:"command,omitempty" json:"command,omitempty"`
	// Args are the arguments to the command.
	Args []string `yaml:"args,omitempty" json:"args,omitempty"`
}

// HelmGenerator runs 'helm template' to render a Helm chart and returns the output.
type HelmGenerator struct {
	// ReleaseName is the release name.
	ReleaseName string `yaml:"releaseName,omitempty" json:"releaseName"`
	// Chart is the Helm chart to install. Prefix with oci:// to use a chart stored in an OCI registry.
	Chart string `yaml:"chart,omitempty" json:"chart"`
	// Version is the version of the helm chart to install.
	Version string `yaml:"version,omitempty" json:"version,omitempty"`
	// Repo is the repository to install the Helm chart from.
	Repo string `yaml:"repo,omitempty" json:"repo,omitempty"`
	// Namespace is the Kubernetes namespace to use when rendering resources.
	Namespace string `yaml:"namespace,omitempty" json:"namespace,omitempty"`
	// APIVersions are Kubernetes api versions used for Capabilities.APIVersions.
	APIVersions []string `yaml:"apiVersions,omitempty" json:"apiVersions,omitempty"`
	// Values are the Helm values used as configuration for the Helm chart.
	Values []Value `yaml:"values,omitempty" json:"values,omitempty"`
}

// KustomizeGenerator runs 'kustomize build' to render a Kustomization and returns the output.
type KustomizeGenerator struct {
	// Dir is the path to a directory containing 'kustomization.yaml' relative to the manifest.
	Dir string `yaml:"dir,omitempty" json:"dir,omitempty" jsonschema:"oneof_required=dir"`
	// URL is a git repository URL with a path suffix containing a 'kustomization.yaml'.
	URL string `yaml:"url,omitempty" json:"url,omitempty"  jsonschema:"oneof_required=url"`
	// EnableHelm enables use of the Helm chart inflator generator.
	EnableHelm bool `yaml:"enableHelm,omitempty" json:"enableHelm,omitempty"`
}

// MergeGenerator takes multiple inputs containing object-like data and deeply merges them together and returns the merged output.
type MergeGenerator struct {
	// Inputs are the inputs to merge. Inputs specified later in the list take precedence, overwriting values in earlier inputs.
	Input []Value `yaml:"input,omitempty" json:"input,omitempty"`
}

// GoTemplateGenerator renders Go 'text/template' templates and returns the output.
type GoTemplateGenerator struct {
	// Template is the template to render.
	Template *Value `yaml:"template,omitempty" json:"template,omitempty"`
	// Vars are input variables to the template.
	Vars map[string]any `yaml:"vars,omitempty" json:"vars,omitempty"`
	// RefVars are input variables to the template allowing references to other generator outputs or files.
	RefVars map[string]Value `yaml:"refVars,omitempty" json:"refVars,omitempty"`
}

// ImportGenerator imports pipelines from another file and returns the output.
type ImportGenerator struct {
	// Path is the path to the manifest to import.
	Path string `yaml:"path" json:"path"`
	// Vars defines variables that the imported pipeline expects.
	Vars []*NamedVariable `yaml:"vars,omitempty" json:"vars,omitempty"`
}

// NamedVariable is named Value that can be referenced by imported pipelines using the var attribute of a value.
type NamedVariable struct {
	Name  string `yaml:"name,omitempty" json:"name"`
	Value `yaml:",inline" json:",inline"`
}

// JQGenerator executes 'jq' and returns the output.
type JQGenerator struct {
	// Expr is the jq expression to evaluate. Cannot be specified in combination with ExprFile.
	Expr string `yaml:"expr,omitempty" json:"expr,omitempty" jsonschema:"oneof_required=expr"`
	// ExprFile is a path to a file relative to the manifest containing the jq expression to evaluate. Cannot be specified in combination with Expr.
	ExprFile string `yaml:"exprFile,omitempty" json:"exprFile,omitempty" jsonschema:"oneof_required=exprFile"`
	// Input is the JSON input for jq to execute the expression over.
	Input *Value `yaml:"input" json:"input"`
	// Slurp configures jq to read all inputs into an array and use it as a single input value.
	Slurp bool `yaml:"slurp,omitempty" json:"slurp,omitempty"`
}

// YAMLGenerator returns it's inputs as YAML.
type YAMLGenerator struct {
	// Inputs are the inputs to convert to YAML. If a single input produces multiple objects or multiple inputs are provided, a stream of YAML documents is returned.
	Input []Value `yaml:"input,omitempty" json:"input"`
}

// JSONGenerator returns it's inputs as JSON.
type JSONGenerator struct {
	// Inputs are the inputs to convert to JSON. If a single input produces multiple objects or multiple inputs are provided, a stream of YAML objects is returned.
	Input []Value `yaml:"input,omitempty" json:"input"`
}

// PipelineGenerator executes other generators in a pipeline or singular context.
type PipelineGenerator struct {
	// Pipeline is a list of generators to run. Generators can reference the output of previous generators using their name in any Value refs.
	Pipeline []Generator `yaml:"pipeline,omitempty" json:"pipeline,omitempty" jsonschema:"oneof_required=pipeline"`
	// Generator is a single generator, for simple use-cases that do not require a full pipeline.
	Generator *Generator `yaml:"generator,omitempty" json:"generator,omitempty" jsonschema:"oneof_required=generator"`
}

// Value provides inputs to generators.
type Value struct {
	// Var allows defining variables that can be externally provided to a pipeline.
	Var *string `yaml:"var,omitempty" json:"var,omitempty" jsonschema:"oneof_required=var"`
	// Ref takes the name of a previous stage in the pipeline and returns the output of that stage.
	Ref *string `yaml:"ref,omitempty" json:"ref,omitempty" jsonschema:"oneof_required=ref"`
	// File takes a path relative to this pipeline file to read and returns the content of the file specified.
	File *string `yaml:"file,omitempty" json:"file,omitempty" jsonschema:"oneof_required=file"`
	// Value simply returns the value specified. It can be any valid YAML/JSON type ( string, boolean, number, array, object).
	Value *any `yaml:"value,omitempty" json:"value,omitempty" jsonschema:"oneof_required=value"`
}
