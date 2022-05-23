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

find ${OSM_HOME}/dockerfiles -type f | xargs sed -i 's# localhost:5000/alpine:3$# alpine:3#g'
find ${OSM_HOME}/dockerfiles -type f | xargs sed -i 's# localhost:5000/library/busybox:1.33# busybox:1.33#g'
find ${OSM_HOME}/dockerfiles -type f | xargs sed -i 's# localhost:5000/library/golang:\$GO_VERSION # golang:$GO_VERSION #g'
find ${OSM_HOME}/dockerfiles -type f | xargs sed -i 's# localhost:5000/distroless/static# gcr.io/distroless/static#g'
find ${OSM_HOME}/dockerfiles -type f | xargs sed -i 's# localhost:5000/flomesh/proxy-wasm-cpp-sdk:v2 AS# flomesh/proxy-wasm-cpp-sdk:v2 AS#g'

sed -i 's#sidecarImage: localhost:5000/envoyproxy/envoy#sidecarImage: envoyproxy/envoy#g' ${OSM_HOME}/charts/osm/values.yaml
sed -i 's#sidecarImage: localhost:5000/flomesh/pipy#sidecarImage: flomesh/pipy#g' ${OSM_HOME}/charts/osm/values.yaml
sed -i 's#curlImage: localhost:5000/curlimages/curl#curlImage: curlimages/curl#g' ${OSM_HOME}/charts/osm/values.yaml
sed -i 's#image: localhost:5000/prom/prometheus:v2.18.1#image: prom/prometheus:v2.18.1#g' ${OSM_HOME}/charts/osm/values.yaml
sed -i 's#image: localhost:5000/grafana/grafana:8.2.2#image: grafana/grafana:8.2.2#g' ${OSM_HOME}/charts/osm/values.yaml
sed -i 's#rendererImage: localhost:5000/grafana/grafana-image-renderer:3.2.1#rendererImage: grafana/grafana-image-renderer:3.2.1#g' ${OSM_HOME}/charts/osm/values.yaml
sed -i 's#image: localhost:5000/jaegertracing/all-in-one#image: jaegertracing/all-in-one#g' ${OSM_HOME}/charts/osm/values.yaml
sed -i 's#docker.io#localhost:5000#g' ${OSM_HOME}/charts/osm/values.yaml
