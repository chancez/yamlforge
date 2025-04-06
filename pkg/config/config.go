package config

// Config defines a yamlforge configuration.
type Config struct {
	PipelineGenerator `yaml:",inline" json:",inline"`
}

// Generators execute some logic and produce output.
// Only one type of generator can be specified.
type Generator struct {
	// Name is the name of this generator which other generators can reference this generator's output by.
	Name string `yaml:"name" json:"name"`
	// Value is a simple generator that takes a value and returns it unaltered.
	Value *AnyValue `yaml:"value,omitempty" json:"value,omitempty" jsonschema:"oneof_required=value"`
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
	// Pipeline executes other pipelines or generators and returns the output.
	Pipeline *PipelineGenerator `yaml:"pipeline,omitempty" json:"pipeline,omitempty" jsonschema:"oneof_required=pipeline"`
	// JQ is a generator which executes 'jq' and returns the output.
	JQ *JQGenerator `yaml:"jq,omitempty" json:"jq,omitempty" jsonschema:"oneof_required=jq"`
	// CEL is a generator which evaluates a CEL expression against the input.
	CEL *CELGenerator `yaml:"cel,omitempty" json:"cel,omitempty" jsonschema:"oneof_required=cel"`
	// JSONPatch is a generator which evaluates a JSONPatch against the input.
	JSONPatch *JSONPatchGenerator `yaml:"jsonpatch,omitempty" json:"jsonpatch,omitempty" jsonschema:"oneof_required=jsonpatch"`
	// YAML is a generator which returns it's inputs as YAML.
	YAML *YAMLGenerator `yaml:"yaml,omitempty" json:"yaml,omitempty" jsonschema:"oneof_required=yaml"`
	// JSON is a generator which returns it's inputs as JSON.
	JSON *JSONGenerator `yaml:"json,omitempty" json:"json,omitempty" jsonschema:"oneof_required=json"`
}

// FileGenerator reads files at the specified path and returns their output.
type FileGenerator struct {
	// Path is the path relative to this pipeline file to read.
	Path string `yaml:"path" json:"path"`
}

// ExecGenerator execs the command specified and returns the stdout of the program.
type ExecGenerator struct {
	// Command is the command to execute.
	Command string `yaml:"command" json:"command"`
	// Args are the arguments to the command.
	Args []string `yaml:"args,omitempty" json:"args,omitempty"`
	// Env is a list of environment variables to set for the command.
	Env []NamedValue `yaml:"env,omitempty" json:"env,omitempty"`
}

// HelmGenerator runs 'helm template' to render a Helm chart and returns the output.
type HelmGenerator struct {
	// ReleaseName is the release name.
	ReleaseName StringValue `yaml:"releaseName" json:"releaseName"`
	// Chart is the Helm chart to install. Prefix with oci:// to use a chart stored in an OCI registry.
	Chart StringValue `yaml:"chart" json:"chart"`
	// Version is the version of the helm chart to install.
	Version StringValue `yaml:"version,omitempty" json:"version,omitempty"`
	// Repo is the repository to install the Helm chart from.
	Repo StringValue `yaml:"repo,omitempty" json:"repo,omitempty"`
	// Namespace is the Kubernetes namespace to use when rendering resources.
	Namespace StringValue `yaml:"namespace,omitempty" json:"namespace,omitempty"`
	// IncludeCRDs specifies if CRDs are included in the templated output
	IncludeCRDs BoolValue `yaml:"includeCRDs,omitempty" json:"includeCRDs,omitempty"`
	// APIVersions are Kubernetes api versions used for Capabilities.APIVersions.
	APIVersions []StringValue `yaml:"apiVersions,omitempty" json:"apiVersions,omitempty"`
	// Values are the Helm values used as configuration for the Helm chart.
	Values []StringValue `yaml:"values,omitempty" json:"values,omitempty"`
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
	Input []ParsedValue `yaml:"input" json:"input"`
}

// GoTemplateGenerator renders Go 'text/template' templates and returns the output.
type GoTemplateGenerator struct {
	// Template is the template to render.
	Template Value `yaml:"template" json:"template"`
	// Vars are input variables to the template.
	Vars map[string]any `yaml:"vars,omitempty" json:"vars,omitempty"`
	// RefVars are input variables to the template allowing references to other generator outputs or files.
	RefVars map[string]ParsedValue `yaml:"refVars,omitempty" json:"refVars,omitempty"`
}

// JQGenerator executes 'jq' and returns the output.
type JQGenerator struct {
	// Expr is the jq expression to evaluate. Cannot be specified in combination with ExprFile.
	Expr string `yaml:"expr,omitempty" json:"expr,omitempty" jsonschema:"oneof_required=expr"`
	// ExprFile is a path to a file relative to the manifest containing the jq expression to evaluate. Cannot be specified in combination with Expr.
	ExprFile string `yaml:"exprFile,omitempty" json:"exprFile,omitempty" jsonschema:"oneof_required=exprFile"`
	// Input is the JSON input for jq to execute the expression over.
	Input Value `yaml:"input" json:"input"`
	// Slurp configures jq to read all inputs into an array and use it as a single input value.
	Slurp bool `yaml:"slurp,omitempty" json:"slurp,omitempty"`
}

// CELGenerator evaluates a CEL expression and returns the result of the expression.
type CELGenerator struct {
	// Input values are parsed then evaluated against the configure CEL expression.
	Input ParsedValue `yaml:"input" json:"input"`
	// Expr is a CEL expression evaluated with the input set to the variable 'val'.
	Expr string `yaml:"expr" json:"expr"`
	// When filter is true, the CEL expression becomes a filter returning a boolean indicating if the input should be kept.
	Filter bool `yaml:"filter,omitempty" json:"filter,omitempty"`
	// If Filter and InvertFilter is true, instead of keeping the result, it will be discarded.
	InvertFilter bool `yaml:"invertFilter,omitempty" json:"invertFilter,omitempty"`
	// Format is the format the output should be returned as. If unspecified, it defaults to YAML.
	Format string `yaml:"format,omitempty" json:"format,omitempty" jsonschema:"enum=yaml,enum=json,default=yaml"`
}

// JSONPatchGenerator evaluates a JSONPatch against the input.
type JSONPatchGenerator struct {
	// Input is the value to apply the patch to. It must be JSON.
	Input Value `yaml:"input" json:"input"`
	// Patch is the JSON patch. If it is YAML, it will be automatically converted to JSON.
	Patch string `yaml:"patch" json:"patch"`
	// If merge is true, then patch is interpreted as a JSON merge patch.
	Merge bool `yaml:"merge,omitempty" json:"merge,omitempty"`
}

// YAMLGenerator returns it's inputs as YAML.
type YAMLGenerator struct {
	// Inputs are the inputs to convert to YAML. If a single input produces multiple objects or multiple inputs are provided, a stream of YAML documents is returned.
	Input []ParsedValue `yaml:"input" json:"input"`
	// Indent defines the indent level to use for the output.
	Indent int `yaml:"indent,omitempty" json:"indent,omitempty"`
}

// JSONGenerator returns it's inputs as JSON.
type JSONGenerator struct {
	// Inputs are the inputs to convert to JSON. If a single input produces multiple objects or multiple inputs are provided, a stream of YAML objects is returned.
	Input []ParsedValue `yaml:"input" json:"input"`
	// Indent defines the indent level to use for the output.
	Indent int `yaml:"indent,omitempty" json:"indent,omitempty"`
}

// PipelineGenerator executes other generators in a pipeline or singular context.
type PipelineGenerator struct {
	// Pipeline is a list of generators to run. Generators can reference the output of previous generators using their name in any Value refs.
	Pipeline []Generator `yaml:"pipeline,omitempty" json:"pipeline,omitempty" jsonschema:"oneof_required=pipeline"`
	// Generator is a single generator, for simple use-cases that do not require a full pipeline.
	Generator *Generator `yaml:"generator,omitempty" json:"generator,omitempty" jsonschema:"oneof_required=generator"`
	// Import is a value containing a pipeline to import. Imported pipelines
	// share no references or variables with their parent pipeline.
	Import *Value `yaml:"import,omitempty" json:"import,omitempty" jsonschema:"oneof_required=import"`
	// Include is a value containing a pipeline to include. Included pipelines
	// share the same context as their parent, meaning variables and references
	// in the parent pipeline are available within the included pipeline,
	// behaving as if the included pipeline was directly written in the parent
	// pipeline.
	// If a included pipeline includes a generator with the same name as it's
	// parent it will result in an error.
	Include *Value `yaml:"include,omitempty" json:"include,omitempty" jsonschema:"oneof_required=include"`
	// Vars defines variables that the pipeline is providing to the sub-pipeline.
	Vars []NamedValue `yaml:"vars,omitempty" json:"vars,omitempty"`
}
