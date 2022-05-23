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

docker pull docker.io/alpine:3
docker pull docker.io/library/busybox:1.33
docker pull docker.io/library/golang:1.17
docker pull docker.io/envoyproxy/envoy:v1.19.3
docker pull docker.io/projectcontour/contour:v1.18.0
docker pull docker.io/flomesh/pipy:latest
docker pull docker.io/flomesh/proxy-wasm-cpp-sdk:v2
docker pull docker.io/prom/prometheus:v2.18.1
docker pull docker.io/grafana/grafana:8.2.2
docker pull docker.io/grafana/grafana-image-renderer:3.2.1
docker pull docker.io/jaegertracing/all-in-one
docker pull gcr.io/distroless/base:latest
docker pull gcr.io/distroless/static:latest

docker tag docker.io/alpine:3 localhost:5000/alpine:3
docker tag docker.io/library/busybox:1.33 localhost:5000/library/busybox:1.33
docker tag docker.io/library/golang:1.17 localhost:5000/library/golang:1.17
docker tag docker.io/envoyproxy/envoy:v1.19.3 localhost:5000/envoyproxy/envoy:v1.19.3
docker tag docker.io/projectcontour/contour:v1.18.0 localhost:5000/projectcontour/contour:v1.18.0
docker tag docker.io/flomesh/pipy:latest localhost:5000/flomesh/pipy:latest
docker tag docker.io/flomesh/proxy-wasm-cpp-sdk:v2 localhost:5000/flomesh/proxy-wasm-cpp-sdk:v2
docker tag docker.io/prom/prometheus:v2.18.1 localhost:5000/prom/prometheus:v2.18.1
docker tag docker.io/grafana/grafana:8.2.2 localhost:5000/grafana/grafana:8.2.2
docker tag docker.io/grafana/grafana-image-renderer:3.2.1 localhost:5000/grafana/grafana-image-renderer:3.2.1
docker tag docker.io/jaegertracing/all-in-one localhost:5000/jaegertracing/all-in-one
docker tag gcr.io/distroless/base:latest localhost:5000/distroless/base:latest
docker tag gcr.io/distroless/static:latest localhost:5000/distroless/static:latest

docker push localhost:5000/alpine:3
docker push localhost:5000/library/busybox:1.33
docker push localhost:5000/library/golang:1.17
docker push localhost:5000/envoyproxy/envoy:v1.19.3
docker push localhost:5000/projectcontour/contour:v1.18.0
docker push localhost:5000/flomesh/pipy:latest
docker push localhost:5000/flomesh/proxy-wasm-cpp-sdk:v2
docker push localhost:5000/prom/prometheus:v2.18.1
docker push localhost:5000/grafana/grafana:8.2.2
docker push localhost:5000/grafana/grafana-image-renderer:3.2.1
docker push localhost:5000/jaegertracing/all-in-one
docker push localhost:5000/distroless/base:latest
docker push localhost:5000/distroless/static:latest

find ${OSM_HOME}/dockerfiles -type f | xargs sed -i 's# alpine:3$# localhost:5000/alpine:3#g'
find ${OSM_HOME}/dockerfiles -type f | xargs sed -i 's# busybox:1.33# localhost:5000/library/busybox:1.33#g'
find ${OSM_HOME}/dockerfiles -type f | xargs sed -i 's# golang:\$GO_VERSION # localhost:5000/library/golang:$GO_VERSION #g'
find ${OSM_HOME}/dockerfiles -type f | xargs sed -i 's# gcr.io/distroless/base# localhost:5000/distroless/base#g'
find ${OSM_HOME}/dockerfiles -type f | xargs sed -i 's# gcr.io/distroless/static# localhost:5000/distroless/static#g'
find ${OSM_HOME}/dockerfiles -type f | xargs sed -i 's# flomesh/proxy-wasm-cpp-sdk:v2 AS# localhost:5000/flomesh/proxy-wasm-cpp-sdk:v2 AS#g'

sed -i 's#docker.io#localhost:5000#g' ${OSM_HOME}/charts/osm/values.yaml
sed -i 's#sidecarImage: envoyproxy/envoy#sidecarImage: localhost:5000/envoyproxy/envoy#g' ${OSM_HOME}/charts/osm/values.yaml
sed -i 's#sidecarImage: flomesh/pipy#sidecarImage: localhost:5000/flomesh/pipy#g' ${OSM_HOME}/charts/osm/values.yaml
sed -i 's#curlImage: curlimages/curl#curlImage: localhost:5000/curlimages/curl#g' ${OSM_HOME}/charts/osm/values.yaml
sed -i 's#image: prom/prometheus:v2.18.1#image: localhost:5000/prom/prometheus:v2.18.1#g' ${OSM_HOME}/charts/osm/values.yaml
sed -i 's#image: grafana/grafana:8.2.2#image: localhost:5000/grafana/grafana:8.2.2#g' ${OSM_HOME}/charts/osm/values.yaml
sed -i 's#rendererImage: grafana/grafana-image-renderer:3.2.1#rendererImage: localhost:5000/grafana/grafana-image-renderer:3.2.1#g' ${OSM_HOME}/charts/osm/values.yaml
sed -i 's#image: jaegertracing/all-in-one#image: localhost:5000/jaegertracing/all-in-one#g' ${OSM_HOME}/charts/osm/values.yaml