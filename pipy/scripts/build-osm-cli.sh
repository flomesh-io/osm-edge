#!/bin/bash

set -euo pipefail

if [ -z "$1" ]; then
  echo "Error: expected one argument OSM_HOME"
  exit 1
fi

OSM_HOME=$1

cd "${OSM_HOME}"
rm -rf bin/osm
rm -rf cmd/cli/chart.tgz
make build-osm
