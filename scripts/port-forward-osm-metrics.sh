#!/bin/bash
# shellcheck disable=SC1091
source .env

POD="$(kubectl get pods --selector app=osm-controller -n "$K8S_NAMESPACE" --no-headers | grep 'Running' | awk 'NR==1{print $1}')"

if [ -z "$POD" ]; then
    echo "Not found pod: osm-controller"
    exit 1
fi

kubectl wait -n "$K8S_NAMESPACE" --for=condition=ready pod --selector app=osm-controller --timeout=900s

kubectl port-forward "$POD" -n "$K8S_NAMESPACE" 9091:9091 --address 0.0.0.0

