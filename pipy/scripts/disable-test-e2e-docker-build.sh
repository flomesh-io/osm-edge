#!/bin/bash

set -euo pipefail

if [ -z "$1" ]; then
  echo "Error: expected one argument OSM_HOME"
  exit 1
fi

OSM_HOME=$1

sed -i 's/^test-e2e: docker-build-osm build-osm docker-build-tcp-echo-server$/test-e2e:/g' ${OSM_HOME}/Makefile
