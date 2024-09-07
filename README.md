# yamlforge

yamlforge is a tool to produce YAML (or JSON) using pipelines.

## Usage

Take a look at our examples in the `examples/` directory for comprehensive examples.

Here's an example that renders a Helm chart using dynamically rendered `values.yaml`:

```
yfg generate ./examples/advanced/helm-templated-values.yaml
```

You use `yfg explain` to get details on the different fields available:

```
yfg explain config
```

Or you can use `yfg json-schema` to print the [JSON Schema](https://json-schema.org) for yamlforge configurations.

## Concepts

yamlforge is built on the concept of pipelines and generators.

A generator is something which generates or produces some text (typically YAML or JSON), it's pretty open-ended.

A pipeline in a series of generators that are composed together to produce your YAML output.

## Why

One of the most common problems I see when trying to define IIAC for Kubernetes is repetition and lack of composability across tools.

With yamlforge, you can use existing tools, and combine them where it makes sense, and supplement their missing features with the generators that yamlforge provides.

## Use Cases

The best way to understand yamlforge is to look at the examples.

A common pain point in Helm is that your `values.yaml` is often duplicated across many environments.
You can have multiple values files to help split it up, but sometimes you want
logic to determine if a value should be set, or you want to generate some of
the values based on other configuration options.

Using the `gotemplate` generator with the `helm` generator in yamlforge enables you do this.

For an example take a look at [`examples/advanced/helm-templated-values.yaml`](examples/advanced/helm-templated-values.yaml).
