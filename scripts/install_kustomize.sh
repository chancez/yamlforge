#!/bin/bash

set -e
set -x
kustomize_version=5.4.3

function install_kustomize() {
  local target_arch="${1:?}"
  local dest_bin="${2:?}"
  local install_os=linux
  local target_file="kustomize_v${kustomize_version}_${install_os}_${target_arch}.tar.gz"

  if ! [ -e "${target_file}" ]; then
    curl -sLf --retry 3 -o "${target_file}" "https://github.com/kubernetes-sigs/kustomize/releases/download/kustomize%2Fv${kustomize_version}/${target_file}"
  fi
  mkdir -p /tmp/kustomize && tar -C /tmp/kustomize -xf "${target_file}"
  install -m 0755 "/tmp/kustomize/kustomize" "${dest_bin}/kustomize"
  kustomize version
}

install_kustomize "$@"
