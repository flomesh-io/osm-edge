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

docker pull flomesh/httpbin:latest
docker tag flomesh/httpbin:latest localhost:5000/flomesh/httpbin:latest
docker push localhost:5000/flomesh/httpbin:latest
find "${OSM_HOME}"/tests -type f -exec sed -i 's#"flomesh/httpbin:latest"#"localhost:5000/flomesh/httpbin:latest"#g' {} +

docker pull flomesh/httpbin:ken
docker tag flomesh/httpbin:ken localhost:5000/flomesh/httpbin:ken
docker push localhost:5000/flomesh/httpbin:ken
find "${OSM_HOME}"/tests -type f -exec sed -i 's#"flomesh/httpbin:ken"#"localhost:5000/flomesh/httpbin:ken"#g' {} +

docker pull busybox:latest
docker tag busybox:latest localhost:5000/busybox:latest
docker push localhost:5000/busybox:latest
find "${OSM_HOME}"/tests -type f -exec sed -i 's#"busybox"#"localhost:5000/busybox"#g' {} +

docker pull curlimages/curl:latest
docker tag curlimages/curl:latest localhost:5000/curlimages/curl:latest
docker push localhost:5000/curlimages/curl:latest
find "${OSM_HOME}"/tests -type f -exec sed -i 's#"curlimages/curl"#"localhost:5000/curlimages/curl"#g' {} +

docker pull flomesh/alpine-debug:latest
docker tag flomesh/alpine-debug:latest localhost:5000/flomesh/alpine-debug:latest
docker push localhost:5000/flomesh/alpine-debug:latest
find "${OSM_HOME}"/tests -type f -exec sed -i 's#"flomesh/alpine-debug"#"localhost:5000/flomesh/alpine-debug"#g' {} +

docker pull nginx:1.19-alpine
docker tag nginx:1.19-alpine localhost:5000/nginx:1.19-alpine
docker push localhost:5000/nginx:1.19-alpine
find "${OSM_HOME}"/tests -type f -exec sed -i 's#"nginx:1.19-alpine"#"localhost:5000/nginx:1.19-alpine"#g' {} +

docker pull flomesh/grpcurl:latest
docker tag flomesh/grpcurl:latest localhost:5000/flomesh/grpcurl:latest
docker push localhost:5000/flomesh/grpcurl:latest
find "${OSM_HOME}"/tests -type f -exec sed -i 's#"flomesh/grpcurl"#"localhost:5000/flomesh/grpcurl"#g' {} +

docker pull flomesh/grpcbin:latest
docker tag flomesh/grpcbin:latest localhost:5000/flomesh/grpcbin:latest
docker push localhost:5000/flomesh/grpcbin:latest
find "${OSM_HOME}"/tests -type f -exec sed -i 's#"flomesh/grpcbin"#"localhost:5000/flomesh/grpcbin"#g' {} +
