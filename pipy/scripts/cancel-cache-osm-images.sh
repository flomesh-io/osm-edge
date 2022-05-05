#!/bin/bash

set -euo pipefail

if [ -z "$1" ]; then
  echo "Error: expected one argument OSM_HOME"
  exit 1
fi

if [ -z "$2" ]; then
  echo "Error: expected one argument OS_ARCH"
  exit 1
fi

OSM_HOME=$1
BUILD_ARCH=$2

if [[ "${BUILD_ARCH}" == "amd64" ]]; then
  find ${OSM_HOME}/dockerfiles -type f | xargs sed -i 's# localhost:5000/library/alpine:3$# alpine:3#g'
  find ${OSM_HOME}/dockerfiles -type f | xargs sed -i 's# localhost:5000/library/busybox:1.33# busybox:1.33#g'
  find ${OSM_HOME}/dockerfiles -type f | xargs sed -i 's# localhost:5000/library/golang:\$GO_VERSION # golang:$GO_VERSION #g'
  find ${OSM_HOME}/dockerfiles -type f | xargs sed -i 's# localhost:5000/distroless/static# gcr.io/distroless/static#g'
  find ${OSM_HOME}/dockerfiles -type f | xargs sed -i 's# localhost:5000/openservicemesh/proxy-wasm-cpp-sdk# openservicemesh/proxy-wasm-cpp-sdk#g'

  sed -i 's#sidecarImage: localhost:5000/envoyproxy/envoy-alpine#sidecarImage: envoyproxy/envoy-alpine#g' ${OSM_HOME}/charts/osm/values.yaml
  sed -i 's#sidecarImage: localhost:5000/flomesh/pipy-nightly#sidecarImage: flomesh/pipy-nightly#g' ${OSM_HOME}/charts/osm/values.yaml
  sed -i 's#curlImage: localhost:5000/curlimages/curl#curlImage: curlimages/curl#g' ${OSM_HOME}/charts/osm/values.yaml
  sed -i 's#localhost:5000#docker.io#g' ${OSM_HOME}/charts/osm/values.yaml
fi

if [[ "${BUILD_ARCH}" == "arm64" ]]; then
  find ${OSM_HOME}/dockerfiles -type f | xargs sed -i 's# localhost:5000/arm64v8/alpine:3.12$# arm64v8/alpine:3.12#g'
  find ${OSM_HOME}/dockerfiles -type f | xargs sed -i 's# localhost:5000/library/busybox:1.33# busybox:1.33#g'
  find ${OSM_HOME}/dockerfiles -type f | xargs sed -i 's# localhost:5000/library/golang:\$GO_VERSION # golang:$GO_VERSION #g'
  find ${OSM_HOME}/dockerfiles -type f | xargs sed -i 's# localhost:5000/distroless/static# gcr.io/distroless/static#g'
  find ${OSM_HOME}/dockerfiles -type f | xargs sed -i 's# localhost:5000/flomesh/proxy-wasm-cpp-sdk:v2 AS# flomesh/proxy-wasm-cpp-sdk:v2 AS#g'

  sed -i 's#sidecarImage: localhost:5000/envoyproxy/envoy#sidecarImage: envoyproxy/envoy#g' ${OSM_HOME}/charts/osm/values.yaml
  sed -i 's#sidecarImage: localhost:5000/flomesh/pipy-nightly#sidecarImage: flomesh/pipy-nightly#g' ${OSM_HOME}/charts/osm/values.yaml
  sed -i 's#curlImage: localhost:5000/curlimages/curl#curlImage: curlimages/curl#g' ${OSM_HOME}/charts/osm/values.yaml
  sed -i 's#docker.io#localhost:5000#g' ${OSM_HOME}/charts/osm/values.yaml
fi
