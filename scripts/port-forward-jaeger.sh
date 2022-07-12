#!/bin/bash

# shellcheck disable=SC1091
set -aueo pipefail

source .env

OSM_POD=$(kubectl get pods -n "$K8S_NAMESPACE" --no-headers  --selector app=jaeger | awk 'NR==1{print $1}')

if [ -z "$OSM_POD" ]; then
    echo "Not found pod: jaeger"
    exit 1
fi

kubectl wait -n "$K8S_NAMESPACE" --for=condition=ready pod --selector app=jaeger --timeout=900s

kubectl port-forward -n "$K8S_NAMESPACE" "$OSM_POD"  16686:16686 --address 0.0.0.0
