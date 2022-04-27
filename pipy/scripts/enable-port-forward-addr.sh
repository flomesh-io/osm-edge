#!/bin/bash

set -euo pipefail

if [ -z "$1" ]; then
  echo "Error: expected one argument OSM_HOME"
  exit 1
fi

OSM_HOME=$1
find ${OSM_HOME}/scripts -type f -name "port-forward-*" | xargs sed -i 's/port-forward "/port-forward --address 0.0.0.0 "/g'