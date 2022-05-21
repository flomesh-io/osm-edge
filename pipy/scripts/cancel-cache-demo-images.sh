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

sed -i 's# localhost:5000/devilbox/mysql:mysql-8.0# devilbox/mysql:mysql-8.0#g' ${OSM_HOME}/demo/deploy-mysql.sh
sed -i 's# localhost:5000/curlimages/curl# curlimages/curl#g' ${OSM_HOME}/demo/multicluster-fault-injection.sh
