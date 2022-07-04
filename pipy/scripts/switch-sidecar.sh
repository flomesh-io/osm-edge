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
  echo "Error: expected one argument SIDECAR"
  exit 1
fi

OSM_HOME=$1
BUILD_ARCH=$2
SIDECAR=$3


"${OSM_HOME}"/pipy/scripts/change_metrics_labels.sh "${OSM_HOME}" "${BUILD_ARCH}" "${SIDECAR}"
