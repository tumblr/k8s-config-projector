#!/bin/bash
set -ex
PROJECTOR_IMAGE="${PROJECTOR_IMAGE:-tumblr/k8s-config-projector:latest}"
GENERATED_DIRECTORY="${GENERATED_DIRECTORY:-.generated/}"
TARGET_CLUSTERS=(
  bf2-DEVEL
  bf2-PRODUCTION
)

rootpath="$(dirname $0)/../.."
pushd $rootpath

# clean up any pre-existing generated stuff, and 
rm -rf "${GENERATED_DIRECTORY}" || :
for target in "${TARGET_CLUSTERS[@]}" ; do
  az="${target%-*}"
  cluster="${target#*-}"
  mkdir -p ${GENERATED_DIRECTORY}/$az/$cluster
done

for target in "${TARGET_CLUSTERS[@]}" ; do
  echo "Projecting manifests for $az-$cluster into ${GENERATED_DIRECTORY}/$az/$cluster"
  az="${target%-*}"
  cluster="${target#*-}"
  # project any manifests for this az/cluster into the .generated/az/cluster directory
  docker run \
    --rm \
    -v "$(pwd)/${GENERATED_DIRECTORY}/$az/$cluster:/output" \
    -v "$(pwd)/projection-manifests/:/manifests:ro" \
    -v "$(pwd)/config:/config:ro" \
    "${PROJECTOR_IMAGE}"
done

