# Argo CD Config Management Plugin Example

This example will deploy Argo CD with `yamlforge` configured as a custom Config Management Plugin (CMP).

You will also use `yamlforge` to manage the configuration of Argo CD, and deploy an example `Application` that uses the `yamlforge` CMP.

## Setup

Create a Kubernetes cluster using [KIND](https://kind.sigs.k8s.io):

```
kind create cluster
```

## Deploy

Next, use `yamlforge` to render the Argo CD helm chart, and the example `Application`:

```
yfg generate examples/argocd-cm-plugin/forge.yaml | kubectl apply -f -
```

Within a few minutes, you should see an `nginx` pod in the default namespace:

```
kubectl get pods -n default
```

## Review

The [`forge.yaml`](./forge.yaml) used the `helm` generator to deploy ArgoCD and the `value` generator to configure an `Application` to deploy.

The example application can be found in the [`example-app`](./example-app) directory.
