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
  find ${OSM_HOME}/demo -type f | xargs sed -i 's/amd64/arm64/g'
  sed -i 's/kind create cluster --name/kind create cluster --image kindest\/node-arm64:v1.20.15 --name/g' ${OSM_HOME}/scripts/kind-with-registry.sh
  sed -i 's/image: mysql:5.6/image: devilbox\/mysql:mysql-8.0/g' ${OSM_HOME}/demo/deploy-mysql.sh
  sed -i 's/repository: envoyproxy\/envoy-alpine/repository: envoyproxy\/envoy/g' ${OSM_HOME}/charts/osm/values.yaml
  sed -i 's/envoy-alpine@sha256:6502a637c6c5fba4d03d0672d878d12da4bcc7a0d0fb3f1d506982dde0039abd/envoy@sha256:9bbd3140c7ba67e32ecdf1731c03f010e2de386ef84d215023327624fc2c37ae/g' ${OSM_HOME}/charts/osm/values.yaml
  sed -i 's/envoy-alpine@sha256:6502a637c6c5fba4d03d0672d878d12da4bcc7a0d0fb3f1d506982dde0039abd/envoy@sha256:9bbd3140c7ba67e32ecdf1731c03f010e2de386ef84d215023327624fc2c37ae/g' ${OSM_HOME}/charts/osm/values.schema.json
fi
