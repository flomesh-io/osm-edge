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
  docker pull docker.io/library/alpine:3
  docker pull docker.io/library/busybox:1.33
  docker pull docker.io/openservicemesh/proxy-wasm-cpp-sdk:956f0d500c380cc1656a2d861b7ee12c2515a664
  docker pull docker.io/library/golang:1.17
  docker pull docker.io/envoyproxy/envoy-alpine:v1.19.3
  docker pull docker.io/projectcontour/contour:v1.18.0
  docker pull docker.io/flomesh/pipy-nightly:latest
  docker pull gcr.io/distroless/base:latest
  docker pull gcr.io/distroless/static:latest

  docker tag docker.io/library/alpine:3 localhost:5000/library/alpine:3
  docker tag docker.io/library/busybox:1.33 localhost:5000/library/busybox:1.33
  docker tag docker.io/openservicemesh/proxy-wasm-cpp-sdk:956f0d500c380cc1656a2d861b7ee12c2515a664 localhost:5000/openservicemesh/proxy-wasm-cpp-sdk:956f0d500c380cc1656a2d861b7ee12c2515a664
  docker tag docker.io/library/golang:1.17 localhost:5000/library/golang:1.17
  docker tag docker.io/envoyproxy/envoy-alpine:v1.19.3 localhost:5000/envoyproxy/envoy-alpine:v1.19.3
  docker tag docker.io/projectcontour/contour:v1.18.0 localhost:5000/projectcontour/contour:v1.18.0
  docker tag docker.io/flomesh/pipy-nightly:latest localhost:5000/flomesh/pipy-nightly:latest
  docker tag gcr.io/distroless/base:latest localhost:5000/distroless/base:latest
  docker tag gcr.io/distroless/static:latest localhost:5000/distroless/static:latest

  docker push localhost:5000/library/alpine:3
  docker push localhost:5000/library/busybox:1.33
  docker push localhost:5000/openservicemesh/proxy-wasm-cpp-sdk:956f0d500c380cc1656a2d861b7ee12c2515a664
  docker push localhost:5000/library/golang:1.17
  docker push localhost:5000/envoyproxy/envoy-alpine:v1.19.3
  docker push localhost:5000/projectcontour/contour:v1.18.0
  docker push localhost:5000/flomesh/pipy-nightly:latest
  docker push localhost:5000/distroless/base:latest
  docker push localhost:5000/distroless/static:latest

  find ${OSM_HOME}/dockerfiles -type f | xargs sed -i 's# alpine:3$# localhost:5000/library/alpine:3#g'
  find ${OSM_HOME}/dockerfiles -type f | xargs sed -i 's# busybox:1.33# localhost:5000/library/busybox:1.33#g'
  find ${OSM_HOME}/dockerfiles -type f | xargs sed -i 's# golang:\$GO_VERSION # localhost:5000/library/golang:$GO_VERSION #g'
  find ${OSM_HOME}/dockerfiles -type f | xargs sed -i 's# gcr.io/distroless/base# localhost:5000/distroless/base#g'
  find ${OSM_HOME}/dockerfiles -type f | xargs sed -i 's# gcr.io/distroless/static# localhost:5000/distroless/static#g'
  find ${OSM_HOME}/dockerfiles -type f | xargs sed -i 's# openservicemesh/proxy-wasm-cpp-sdk# localhost:5000/openservicemesh/proxy-wasm-cpp-sdk#g'

  sed -i 's#docker.io#localhost:5000#g' ${OSM_HOME}/charts/osm/values.yaml
  sed -i 's#sidecarImage: envoyproxy/envoy-alpine#sidecarImage: localhost:5000/envoyproxy/envoy-alpine#g' ${OSM_HOME}/charts/osm/values.yaml
  sed -i 's#sidecarImage: flomesh/pipy-nightly#sidecarImage: localhost:5000/flomesh/pipy-nightly#g' ${OSM_HOME}/charts/osm/values.yaml
  sed -i 's#curlImage: curlimages/curl#curlImage: localhost:5000/curlimages/curl#g' ${OSM_HOME}/charts/osm/values.yaml
fi

if [[ "${BUILD_ARCH}" == "arm64" ]]; then
  docker pull docker.io/arm64v8/alpine:3.12
  docker pull docker.io/library/busybox:1.33
  docker pull docker.io/library/golang:1.17
  docker pull docker.io/envoyproxy/envoy:v1.19.3
  docker pull docker.io/projectcontour/contour:v1.18.0
  docker pull docker.io/flomesh/pipy-nightly:latest
  docker pull docker.io/flomesh/proxy-wasm-cpp-sdk:v2
  docker pull gcr.io/distroless/base:latest
  docker pull gcr.io/distroless/static:latest

  docker tag docker.io/arm64v8/alpine:3.12 localhost:5000/arm64v8/alpine:3.12
  docker tag docker.io/library/busybox:1.33 localhost:5000/library/busybox:1.33
  docker tag docker.io/library/golang:1.17 localhost:5000/library/golang:1.17
  docker tag docker.io/envoyproxy/envoy:v1.19.3 localhost:5000/envoyproxy/envoy:v1.19.3
  docker tag docker.io/projectcontour/contour:v1.18.0 localhost:5000/projectcontour/contour:v1.18.0
  docker tag docker.io/flomesh/pipy-nightly:latest localhost:5000/flomesh/pipy-nightly:latest
  docker tag gcr.io/distroless/base:latest localhost:5000/distroless/base:latest
  docker tag gcr.io/distroless/static:latest localhost:5000/distroless/static:latest
  docker tag docker.io/flomesh/proxy-wasm-cpp-sdk:v2 localhost:5000/flomesh/proxy-wasm-cpp-sdk:v2

  docker push localhost:5000/arm64v8/alpine:3.12
  docker push localhost:5000/library/busybox:1.33
  docker push localhost:5000/library/golang:1.17
  docker push localhost:5000/envoyproxy/envoy:v1.19.3
  docker push localhost:5000/projectcontour/contour:v1.18.0
  docker push localhost:5000/flomesh/pipy-nightly:latest
  docker push localhost:5000/distroless/base:latest
  docker push localhost:5000/distroless/static:latest
  docker push localhost:5000/flomesh/proxy-wasm-cpp-sdk:v2

  find ${OSM_HOME}/dockerfiles -type f | xargs sed -i 's# arm64v8/alpine:3.12$# localhost:5000/arm64v8/alpine:3.12#g'
  find ${OSM_HOME}/dockerfiles -type f | xargs sed -i 's# busybox:1.33# localhost:5000/library/busybox:1.33#g'
  find ${OSM_HOME}/dockerfiles -type f | xargs sed -i 's# golang:\$GO_VERSION # localhost:5000/library/golang:$GO_VERSION #g'
  find ${OSM_HOME}/dockerfiles -type f | xargs sed -i 's# gcr.io/distroless/base# localhost:5000/distroless/base#g'
  find ${OSM_HOME}/dockerfiles -type f | xargs sed -i 's# gcr.io/distroless/static# localhost:5000/distroless/static#g'
  find ${OSM_HOME}/dockerfiles -type f | xargs sed -i 's# flomesh/proxy-wasm-cpp-sdk:v2 AS# localhost:5000/flomesh/proxy-wasm-cpp-sdk:v2 AS#g'

  sed -i 's#docker.io#localhost:5000#g' ${OSM_HOME}/charts/osm/values.yaml
  sed -i 's#sidecarImage: envoyproxy/envoy#sidecarImage: localhost:5000/envoyproxy/envoy#g' ${OSM_HOME}/charts/osm/values.yaml
  sed -i 's#sidecarImage: flomesh/pipy-nightly#sidecarImage: localhost:5000/flomesh/pipy-nightly#g' ${OSM_HOME}/charts/osm/values.yaml
  sed -i 's#curlImage: curlimages/curl#curlImage: localhost:5000/curlimages/curl#g' ${OSM_HOME}/charts/osm/values.yaml
fi
