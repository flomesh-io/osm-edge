#!/bin/bash
# shellcheck disable=SC1091
source .env

POD="$(kubectl get pods --selector app=bookbuyer -n "$BOOKBUYER_NAMESPACE" --no-headers | grep 'Running' | awk 'NR==1{print $1}')"

if [ -z "$POD" ]; then
    echo "Not found pod: bookbuyer"
    exit 1
fi

kubectl wait -n "$BOOKBUYER_NAMESPACE" --for=condition=ready pod --selector app=bookbuyer --timeout=900s

kubectl port-forward "$POD" -n "$BOOKBUYER_NAMESPACE" 15000:15000 --address 0.0.0.0

