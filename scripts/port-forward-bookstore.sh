#!/bin/bash
# shellcheck disable=SC1091
source .env

backend="$1"
thisScript="$(dirname "$0")/$(basename "$0")"

if [ -z "$backend" ]; then
    echo "Usage: $thisScript <backend-name>"
    exit 1
fi

POD="$(kubectl get pods --selector app="$backend" -n "$BOOKSTORE_NAMESPACE" --no-headers | grep 'Running' | awk 'NR==1{print $1}')"

if [ -z "$POD" ]; then
    echo "Not found pod: $backend"
    exit 1
fi

kubectl wait -n "$BOOKSTORE_NAMESPACE" --for=condition=ready pod --selector app="$backend" --timeout=900s

kubectl port-forward "$POD" -n "$BOOKSTORE_NAMESPACE" 15000:15000 --address 0.0.0.0

