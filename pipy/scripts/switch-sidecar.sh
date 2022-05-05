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

if [ -z "$3" ]; then
  echo "Error: expected one argument SIDECAR"
  exit 1
fi

OSM_HOME=$1
BUILD_ARCH=$2
SIDECAR=$3

if [[ "${SIDECAR}" == "pipy" ]]; then
  sed -i 's/"sidecar-type", "envoy"/"sidecar-type", "pipy"/g' ${OSM_HOME}/cmd/osm-injector/osm-injector.go
  sed -i 's/"sidecar-type", "envoy"/"sidecar-type", "pipy"/g' ${OSM_HOME}/cmd/osm-controller/osm-controller.go

  if [[ "${BUILD_ARCH}" == "amd64" ]]; then
    sed -i 's#envoyproxy/envoy-alpine:v1.19.3@sha256:.*b0f3"#flomesh/pipy-nightly:latest"#g' ${OSM_HOME}/charts/osm/values.schema.json
    sed -i 's#envoyproxy/envoy-alpine:v1.19.3@sha256:.*b0f3$#flomesh/pipy-nightly:latest#g' ${OSM_HOME}/charts/osm/values.yaml
    sed -i 's#envoyproxy/envoy-alpine$#flomesh/pipy-nightly#g' ${OSM_HOME}/charts/osm/values.yaml
  fi
  if [[ "${BUILD_ARCH}" == "arm64" ]]; then
    sed -i 's#envoyproxy/envoy:v1.19.3@sha256:.*37ae"#flomesh/pipy-nightly:latest"#g' ${OSM_HOME}/charts/osm/values.schema.json
    sed -i 's#envoyproxy/envoy:v1.19.3@sha256:.*37ae#flomesh/pipy-nightly:latest#g' ${OSM_HOME}/charts/osm/values.yaml
    sed -i 's#envoyproxy/envoy$#flomesh/pipy-nightly#g' ${OSM_HOME}/charts/osm/values.yaml
  fi
  sed -i 's#tag: v1.19.3#tag: latest#g' ${OSM_HOME}/charts/osm/values.yaml
  sed -i '/COPY --from=builder \/osm\/osm-controller \//aCOPY pipy\/repo \/repo' ${OSM_HOME}/dockerfiles/Dockerfile.osm-controller
fi

if [[ "${SIDECAR}" == "envoy" ]]; then
  sed -i 's/"sidecar-type", "pipy"/"sidecar-type", "envoy"/g' ${OSM_HOME}/cmd/osm-injector/osm-injector.go
  sed -i 's/"sidecar-type", "pipy"/"sidecar-type", "envoy"/g' ${OSM_HOME}/cmd/osm-controller/osm-controller.go

  if [[ "${BUILD_ARCH}" == "amd64" ]]; then
    sed -i 's#flomesh/pipy-nightly:latest"#envoyproxy/envoy-alpine:v1.19.3@sha256:874e699857e023d9234b10ffc5af39ccfc9011feab89638e56ac4042ecd4b0f3"#g' ${OSM_HOME}/charts/osm/values.schema.json
    sed -i 's#flomesh/pipy-nightly:latest$#envoyproxy/envoy-alpine:v1.19.3@sha256:874e699857e023d9234b10ffc5af39ccfc9011feab89638e56ac4042ecd4b0f3#g' ${OSM_HOME}/charts/osm/values.yaml
    sed -i 's#flomesh/pipy-nightly$#envoyproxy/envoy-alpine#g' ${OSM_HOME}/charts/osm/values.yaml
  fi
  if [[ "${BUILD_ARCH}" == "arm64" ]]; then
    sed -i 's#flomesh/pipy-nightly:latest"#envoyproxy/envoy:v1.19.3@sha256:9bbd3140c7ba67e32ecdf1731c03f010e2de386ef84d215023327624fc2c37ae"#g' ${OSM_HOME}/charts/osm/values.schema.json
    sed -i 's#flomesh/pipy-nightly:latest#envoyproxy/envoy:v1.19.3@sha256:9bbd3140c7ba67e32ecdf1731c03f010e2de386ef84d215023327624fc2c37ae#g' ${OSM_HOME}/charts/osm/values.yaml
    sed -i 's#flomesh/pipy-nightly$#envoyproxy/envoy#g' ${OSM_HOME}/charts/osm/values.yaml
  fi
  sed -i 's#tag: latest#tag: v1.19.3#g' ${OSM_HOME}/charts/osm/values.yaml
  sed -i '/^COPY pipy\/repo \/repo/d' ${OSM_HOME}/dockerfiles/Dockerfile.osm-controller
fi