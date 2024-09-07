# yamlforge

`yamlforge` is a versatile tool designed to simplify the generation of YAML (or JSON) through the use of pipelines.
It enables users to compose configurations dynamically by chaining together reusable building blocks called *generators*.

## Key Features

- **Pipeline-based generation**: Structure and combine multiple generators in a series to produce the final output.
- **Composability**: Seamlessly integrate existing tools and enhance them with `yamlforge`'s flexible generators.
- **Extensibility**: Use `yamlforge` to fill gaps in your Infrastructure-as-Code (IaC) workflows, particularly for Kubernetes, by automating the generation of complex configurations.

## How It Works

At its core, `yamlforge` operates around two main concepts:

- **Generators**: Components that produce output, typically YAML or JSON, based on inputs or templates.
  Generators can be anything from simple file readers to more advanced tools like templating engines or shell commands.
- **Pipelines**: A series of generators composed together, allowing for complex, layered configuration output.
  Each generator in the pipeline can either build on the previous one or introduce new data.

This approach helps avoid repetitive configurations by enabling dynamic composition and generation of YAML or JSON files based on conditions, templates, and external tools.

## Why yamlforge?

`yamlforge` is designed to solve the problem of repetitive, hard-to-manage configurations, especially in Kubernetes workflows.
Many tools lack composability, leading to duplication and rigidity in Infrastructure-as-Code setups.
`yamlforge` addresses this by allowing you to:

- Combine different tools in a meaningful way.
- Use conditional logic and templates to tailor configurations to different environments.
- Supplement missing features from existing tools with `yamlforge`’s own generators.

## Use Cases

`yamlforge` is ideal for various use cases where configuration complexity can be reduced:

- **Kubernetes Configurations**: Enhance tools like `kustomize` and Helm with `yamlforge` to produce richer, more dynamic Kubernetes configurations.
  See [kustomize.yfg.yaml](examples/kustomize.yfg.yaml) and [helm.yfg.yaml](examples/helm.yfg.yaml) for an example.

- **Dynamic Values with Helm**: Use `yamlforge` to generate dynamic `values.yaml` files for Helm charts.
  See how templating can help in [helm-templated-values.yfg.yaml](examples/advanced/helm-templated-values.yfg.yaml). For a more advanced use-case, see how to dynamically retrieve values in [helm-dynamically-retrieved-values.yfg.yaml](examples/advanced/helm-dynamically-retrieved-values.yfg.yaml).

- **Composable Transformers**: Build reusable transformers that can be applied to different configurations.
   Check out [reusable-transformer.yfg.yaml](examples/advanced/reusable-transformer.yfg.yaml) for a reusable transformer in action.

- **Dynamic Pipelines**: Create dynamic pipelines that change based on input.
  Take a look at [dynamic-pipeline.yfg.yaml](examples/advanced/dynamic-pipeline.yfg.yaml) to see this flexibility in action.

- **Integration with [CEL](https://cel.dev) (common expression language)**: Use `CEL` to extract relevant attributes or filter results.
  See [cel.yfg.yaml](examples/cel.yfg.yaml) and [cel-filter.yfg.yaml](examples/cel-filter.yfg.yaml) for an example.

- **Integration with [jq](https://jqlang.github.io/jq/)**: `jq` can be used to extract or transform data from other pipelines: [jq.yfg.yaml](examples/jq.yfg.yaml).


## Learn More

Explore the examples in the `examples/` directory to see `yamlforge` in action. Additionally, you can:

- Run `yfg explain` to explore the available configuration fields in detail.
- Use `yfg json-schema` to generate the [JSON Schema](https://json-schema.org) for validating your `yamlforge` configurations.
