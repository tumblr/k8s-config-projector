#!/usr/bin/env bash
set -e

echo "Desired namespaces: $@"
for cluster_namespace in "$@" ; do
  target="${cluster_namespace%/*}"
  ns="${cluster_namespace##*/}"
  if [[ -r /etc/k8s-kubeconfig/kubeconfig-${target} ]] ; then
    export KUBECONFIG="/etc/k8s-kubeconfig/kubeconfig-${target}"
    echo "Setting KUBECONFIG=$KUBECONFIG"
  else
    echo "Changing to context $target"
    kubectl config use-context "$(echo "$target" | tr '[:upper:]' '[:lower:]')"
  fi

  echo -n "Ensuring namespace=$ns is present in ${target} with context "
  kubectl config current-context

  if !kubectl get namespace $ns &>/dev/null ; then
    echo -n "  Creating namespace $ns... "
    kubectl create namespace $ns
  fi
done

