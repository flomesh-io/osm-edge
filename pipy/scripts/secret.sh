#!/bin/bash

set -euo pipefail

if [ -z "$1" ]; then
  echo "Error: expected one argument OSM_HOME"
  exit 1
fi

OSM_HOME=$1

cd ${OSM_HOME}
make kind-up
kubectl get namespace osm-system || kubectl create namespace osm-system
kubectl get secret acr-creds --namespace=osm-system || kubectl --namespace=osm-system \
create secret docker-registry acr-creds \
--docker-server=localhost:5000 \
--docker-username=flomesh \
--docker-password=flomesh
sed -i '/^export CTR_REGISTRY_USERNAME=/d' ${OSM_HOME}/.env
sed -i '/export CTR_REGISTRY_PASSWORD=/i\export CTR_REGISTRY_USERNAME=flomesh' ${OSM_HOME}/.env
sed -i 's/^export CTR_REGISTRY_PASSWORD=.*/export CTR_REGISTRY_PASSWORD=flomesh/' ${OSM_HOME}/.env
make kind-reset
