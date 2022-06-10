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
  echo "Error: expected one argument OS"
  exit 1
fi

BUILD_ARCH=$2

wget -q https://registry.hub.docker.com/v1/repositories/kindest/node-"${BUILD_ARCH}"/tags -O - | jq -r '.[].name' | sort -u
