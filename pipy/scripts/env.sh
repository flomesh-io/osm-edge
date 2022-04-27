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

OSM_HOME=$1
BUILD_ARCH=$2
BUILD_OS=$3
TARGETS=${BUILD_OS}/${BUILD_ARCH}
DOCKER_BUILDX_PLATFORM=${BUILD_OS}/${BUILD_ARCH}

cd ${OSM_HOME}
make .env
sed -i 's/localhost:5000$/localhost:5000\/flomesh/g' .env
sed -i 's/^# export CTR_TAG=.*/export CTR_TAG=latest/g' .env
sed -i 's/^#export USE_PRIVATE_REGISTRY=.*/export USE_PRIVATE_REGISTRY=true/g' .env

sed -i '/^export TARGETS=/d' ~/.bashrc
sed -i '/^export DOCKER_BUILDX_PLATFORM=/d' ~/.bashrc
sed -i '/^export CTR_REGISTRY=/d' ~/.bashrc
sed -i '/^export CTR_TAG=/d' ~/.bashrc

echo export TARGETS=${TARGETS} >> ~/.bashrc
echo export DOCKER_BUILDX_PLATFORM=${DOCKER_BUILDX_PLATFORM} >> ~/.bashrc
echo export CTR_REGISTRY=localhost:5000/flomesh >> ~/.bashrc
echo export CTR_TAG=latest >> ~/.bashrc