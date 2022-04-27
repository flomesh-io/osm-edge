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
  docker pull kennethreitz/httpbin:latest
  docker tag kennethreitz/httpbin:latest localhost:5000/kennethreitz/httpbin:latest
  docker push localhost:5000/kennethreitz/httpbin:latest
  find ${OSM_HOME}/tests -type f | xargs sed -i 's#"kennethreitz/httpbin"#"localhost:5000/kennethreitz/httpbin:latest"#g'

  docker pull simonkowallik/httpbin:latest
  docker tag simonkowallik/httpbin:latest localhost:5000/simonkowallik/httpbin:latest
  docker push localhost:5000/simonkowallik/httpbin:latest
  find ${OSM_HOME}/tests -type f | xargs sed -i 's#"simonkowallik/httpbin"#"localhost:5000/simonkowallik/httpbin:latest"#g'

  docker pull busybox:latest
  docker tag busybox:latest localhost:5000/busybox:latest
  docker push localhost:5000/busybox:latest
  find ${OSM_HOME}/tests -type f | xargs sed -i 's#"busybox"#"localhost:5000/busybox:latest"#g'

  docker pull curlimages/curl:latest
  docker tag curlimages/curl:latest localhost:5000/curlimages/curl:latest
  docker push localhost:5000/curlimages/curl:latest
  find ${OSM_HOME}/tests -type f | xargs sed -i 's#"curlimages/curl"#"localhost:5000/curlimages/curl:latest"#g'

  docker pull songrgg/alpine-debug:latest
  docker tag songrgg/alpine-debug:latest localhost:5000/songrgg/alpine-debug:latest
  docker push localhost:5000/songrgg/alpine-debug:latest
  find ${OSM_HOME}/tests -type f | xargs sed -i 's#"songrgg/alpine-debug"#"localhost:5000/songrgg/alpine-debug:latest"#g'

  docker pull nginx:1.19-alpine
  docker tag nginx:1.19-alpine localhost:5000/nginx:1.19-alpine
  docker push localhost:5000/nginx:1.19-alpine
  find ${OSM_HOME}/tests -type f | xargs sed -i 's#"nginx:1.19-alpine"#"localhost:5000/nginx:1.19-alpine"#g'

  docker pull networld/grpcurl:latest
  docker tag networld/grpcurl:latest localhost:5000/networld/grpcurl:latest
  docker push localhost:5000/networld/grpcurl:latest
  find ${OSM_HOME}/tests -type f | xargs sed -i 's#"networld/grpcurl"#"localhost:5000/networld/grpcurl:latest"#g'

  docker pull moul/grpcbin:latest
  docker tag moul/grpcbin:latest localhost:5000/moul/grpcbin:latest
  docker push localhost:5000/moul/grpcbin:latest
  find ${OSM_HOME}/tests -type f | xargs sed -i 's#"moul/grpcbin"#"localhost:5000/moul/grpcbin:latest"#g'
fi

if [[ "${BUILD_ARCH}" == "arm64" ]]; then
  docker pull naqvis/httpbin:latest
  docker tag naqvis/httpbin:latest localhost:5000/naqvis/httpbin:latest
  docker push localhost:5000/naqvis/httpbin:latest
  find ${OSM_HOME}/tests -type f | xargs sed -i 's#"kennethreitz/httpbin"#"localhost:5000/naqvis/httpbin:latest"#g'
  find ${OSM_HOME}/tests -type f | xargs sed -i 's#"simonkowallik/httpbin"#"localhost:5000/naqvis/httpbin:latest"#g'

  docker pull busybox:latest
  docker tag busybox:latest localhost:5000/busybox:latest
  docker push localhost:5000/busybox:latest
  find ${OSM_HOME}/tests -type f | xargs sed -i 's#"busybox"#"localhost:5000/busybox:latest"#g'

  docker pull curlimages/curl:latest
  docker tag curlimages/curl:latest localhost:5000/curlimages/curl:latest
  docker push localhost:5000/curlimages/curl:latest
  find ${OSM_HOME}/tests -type f | xargs sed -i 's#"curlimages/curl"#"localhost:5000/curlimages/curl:latest"#g'

  docker pull naqvis/alpine-debug:latest
  docker tag naqvis/alpine-debug:latest localhost:5000/naqvis/alpine-debug:latest
  docker push localhost:5000/naqvis/alpine-debug:latest
  find ${OSM_HOME}/tests -type f | xargs sed -i 's#"songrgg/alpine-debug"#"localhost:5000/naqvis/alpine-debug:latest"#g'

  docker pull zmazay/alpine-debug
  docker tag zmazay/alpine-debug localhost:5000/zmazay/alpine-debug:latest
  docker push localhost:5000/zmazay/alpine-debug:latest
  find ${OSM_HOME}/tests -type f | xargs sed -i 's#naqvis/alpine-debug#zmazay/alpine-debug#g'
  find ${OSM_HOME}/tests -type f | xargs sed -i 's#"songrgg/alpine-debug"#"localhost:5000/zmazay/alpine-debug:latest"#g'

  docker pull nginx:1.19-alpine
  docker tag nginx:1.19-alpine localhost:5000/nginx:1.19-alpine
  docker push localhost:5000/nginx:1.19-alpine
  find ${OSM_HOME}/tests -type f | xargs sed -i 's#"nginx:1.19-alpine"#"localhost:5000/nginx:1.19-alpine"#g'

  docker pull naqvis/grpcurl:latest
  docker tag naqvis/grpcurl:latest localhost:5000/naqvis/grpcurl:latest
  docker push localhost:5000/naqvis/grpcurl:latest
  find ${OSM_HOME}/tests -type f | xargs sed -i 's#"networld/grpcurl"#"localhost:5000/naqvis/grpcurl:latest"#g'

  docker pull naqvis/grpcbin:latest
  docker tag naqvis/grpcbin:latest localhost:5000/naqvis/grpcbin:latest
  docker push localhost:5000/naqvis/grpcbin:latest
  find ${OSM_HOME}/tests -type f | xargs sed -i 's#"moul/grpcbin"#"localhost:5000/naqvis/grpcbin:latest"#g'
fi
