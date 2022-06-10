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

find "${OSM_HOME}"/tests -type f -exec sed -i 's#"localhost:5000/flomesh/httpbin:latest"#"flomesh/httpbin:latest"#g' {} +
find "${OSM_HOME}"/tests -type f -exec sed -i 's#"localhost:5000/flomesh/httpbin:ken"#"flomesh/httpbin:ken"#g' {} +
find "${OSM_HOME}"/tests -type f -exec sed -i 's#"localhost:5000/busybox"#"busybox"#g' {} +
find "${OSM_HOME}"/tests -type f -exec sed -i 's#"localhost:5000/curlimages/curl"#"curlimages/curl"#g' {} +
find "${OSM_HOME}"/tests -type f -exec sed -i 's#"localhost:5000/flomesh/alpine-debug"#"flomesh/alpine-debug"#g' {} +
find "${OSM_HOME}"/tests -type f -exec sed -i 's#"localhost:5000/nginx:1.19-alpine"#"nginx:1.19-alpine"#g' {} +
find "${OSM_HOME}"/tests -type f -exec sed -i 's#"localhost:5000/flomesh/grpcurl"#"flomesh/grpcurl"#g' {} +
find "${OSM_HOME}"/tests -type f -exec sed -i 's#"localhost:5000/flomesh/grpcbin"#"flomesh/grpcbin"#g' {} +
