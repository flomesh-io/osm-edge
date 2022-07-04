#!/bin/bash

# shellcheck disable=SC1091
set -aueo pipefail

source .env

OSM_POD=$(kubectl get pods -n "$K8S_NAMESPACE" --no-headers  --selector app=osm-controller | awk 'NR==1{print $1}')

kubectl port-forward --address 0.0.0.0 -n "$K8S_NAMESPACE" "$OSM_POD" 6060:6060 --address 0.0.0.0
