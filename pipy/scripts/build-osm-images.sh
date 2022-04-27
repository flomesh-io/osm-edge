#!/bin/bash

set -euo pipefail

if [ -z "$1" ]; then
  echo "Error: expected one argument OSM_HOME"
  exit 1
fi

OSM_HOME=$1

cd ${OSM_HOME}
make docker-build
docker pull docker.io/flomesh/pipy-nightly:latest
docker tag flomesh/pipy-nightly:latest localhost:5000/flomesh/pipy-nightly:latest
docker push localhost:5000/flomesh/pipy-nightly:latest
