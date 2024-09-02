#!/bin/bash

set -e
set -x
helm_version=3.15.2

function install_helm() {
  local target_arch="${1:?}"
  local dest_bin="${2:?}"
  local install_os=linux
  local target_file="helm-v${helm_version}-${install_os}-${target_arch}.tar.gz"

  if ! [ -e "${target_file}" ]; then
    curl -sLf --retry 3 -o "${target_file}" "https://get.helm.sh/${target_file}"
  fi
  mkdir -p /tmp/helm && tar -C /tmp/helm -xf "${target_file}"
  install -m 0755 "/tmp/helm/${install_os}-${target_arch}/helm" "${dest_bin}/helm"
  helm version --client
}

install_helm "$@"
