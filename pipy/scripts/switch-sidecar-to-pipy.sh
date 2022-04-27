#!/bin/bash

set -euo pipefail

if [ -z "$1" ]; then
  echo "Error: expected one argument OSM_HOME"
  exit 1
fi

OSM_HOME=$1

sed -i 's#envoyproxy/envoy-alpine:v1.19.3@sha256:874e699857e023d9234b10ffc5af39ccfc9011feab89638e56ac4042ecd4b0f3$#flomesh/pipy-nightly:latest#g' ${OSM_HOME}/charts/osm/values.yaml
sed -i 's#envoyproxy/envoy-alpine$#flomesh/pipy-nightly#g' ${OSM_HOME}/charts/osm/values.yaml
sed -i 's#tag: v1.19.3#tag: latest#g' ${OSM_HOME}/charts/osm/values.yaml
sed -i '/COPY --from=builder \/osm\/osm-controller \//aCOPY pipy\/repo \/repo' ${OSM_HOME}/dockerfiles/Dockerfile.osm-controller
