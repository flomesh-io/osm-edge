#!/bin/bash

# shellcheck disable=SC1091
set -aueo pipefail

source .env

OSM_POD=$(kubectl get pods -n "$K8S_NAMESPACE" --no-headers  --selector app=osm-controller | awk 'NR==1{print $1}')

if [ -z "$POD" ]; then
    echo "Not found pod: osm-controller"
    exit 1
fi

kubectl wait -n "$K8S_NAMESPACE" --for=condition=ready pod --selector app=osm-controller --timeout=900s

kubectl port-forward -n "$K8S_NAMESPACE" "$OSM_POD" 6060:6060 --address 0.0.0.0
