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

docker pull alpine:3
docker pull busybox:latest
docker pull curlimages/curl:latest
docker pull curlimages/curl:latest
docker pull devilbox/mysql:mysql-8.0
docker pull envoyproxy/envoy:v1.19.3
docker pull flomesh/alpine-debug:latest
docker pull grafana/grafana:8.2.2
docker pull grafana/grafana-image-renderer:3.2.1
docker pull jaegertracing/all-in-one
docker pull library/busybox:1.33
docker pull library/golang:1.17
docker pull nginx:1.19-alpine
docker pull projectcontour/contour:v1.18.0
docker pull prom/prometheus:v2.18.1

docker pull flomesh/pipy:latest
docker pull flomesh/grpcurl:latest
docker pull flomesh/grpcbin:latest
docker pull flomesh/httpbin:latest
docker pull flomesh/httpbin:ken
docker pull flomesh/proxy-wasm-cpp-sdk:v2

docker pull quay.io/jetstack/cert-manager-controller:v1.3.1
docker pull quay.io/jetstack/cert-manager-cainjector:v1.3.1
docker pull quay.io/jetstack/cert-manager-webhook:v1.3.1

docker pull gcr.io/distroless/base:latest
docker pull gcr.io/distroless/static:latest