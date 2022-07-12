#!/bin/bash
# shellcheck disable=SC1091
source .env

POD="$(kubectl get pods --selector app="osm-prometheus" -n "$K8S_NAMESPACE" --no-headers | grep 'Running' | awk 'NR==1{print $1}')"

if [ -z "$POD" ]; then
    echo "Not found pod: osm-prometheus"
    exit 1
fi

kubectl wait -n "$K8S_NAMESPACE" --for=condition=ready pod --selector app=osm-prometheus --timeout=900s

kubectl port-forward "$POD" -n "$K8S_NAMESPACE" 7070:7070 --address 0.0.0.0

