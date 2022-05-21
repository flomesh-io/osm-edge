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

if [[ "${BUILD_ARCH}" == "arm64" ]]; then
  find ${OSM_HOME}/charts -type f | xargs sed -i 's/amd64/arm64/g'
  sed -i 's/repository: envoyproxy\/envoy-alpine/repository: envoyproxy\/envoy/g' ${OSM_HOME}/charts/osm/values.yaml
  sed -i 's/envoy-alpine:v1.19.3@sha256:.*b0f3/envoy:v1.19.3@sha256:9bbd3140c7ba67e32ecdf1731c03f010e2de386ef84d215023327624fc2c37ae/g' ${OSM_HOME}/charts/osm/README.md
  sed -i 's/envoy-alpine:v1.19.3@sha256:.*b0f3/envoy:v1.19.3@sha256:9bbd3140c7ba67e32ecdf1731c03f010e2de386ef84d215023327624fc2c37ae/g' ${OSM_HOME}/charts/osm/values.yaml
  sed -i 's/envoy-alpine:v1.19.3@sha256:.*b0f3/envoy:v1.19.3@sha256:9bbd3140c7ba67e32ecdf1731c03f010e2de386ef84d215023327624fc2c37ae/g' ${OSM_HOME}/charts/osm/values.schema.json
  sed -i 's/envoy-alpine:v1.19.3@sha256:.*b0f3/envoy:v1.19.3@sha256:9bbd3140c7ba67e32ecdf1731c03f010e2de386ef84d215023327624fc2c37ae/g' ${OSM_HOME}/cmd/osm-bootstrap/osm-bootstrap_test.go

  sed -i 's/kind create cluster --name/kind create cluster --image kindest\/node-arm64:v1.20.15 --name/g' ${OSM_HOME}/scripts/kind-with-registry.sh

  find ${OSM_HOME}/demo -type f | xargs sed -i 's/amd64/arm64/g'

  find ${OSM_HOME}/tests -type f | xargs sed -i 's#"simonkowallik/httpbin"#"flomesh/httpbin:latest"#g'
  find ${OSM_HOME}/tests -type f | xargs sed -i 's#"kennethreitz/httpbin"#"flomesh/httpbin:ken"#g'
  find ${OSM_HOME}/tests -type f | xargs sed -i 's#"songrgg/alpine-debug"#"flomesh/alpine-debug"#g'
  find ${OSM_HOME}/tests -type f | xargs sed -i 's#"networld/grpcurl"#"flomesh/grpcurl"#g'
  find ${OSM_HOME}/tests -type f | xargs sed -i 's#"moul/grpcbin"#"flomesh/grpcbin"#g'
fi

find ${OSM_HOME}/dockerfiles -type f | xargs sed -i 's#openservicemesh/proxy-wasm-cpp-sdk:.* AS#flomesh/proxy-wasm-cpp-sdk:v2 AS#g'
sed -i 's/image: mysql:5.6/image: devilbox\/mysql:mysql-8.0/g' ${OSM_HOME}/demo/deploy-mysql.sh