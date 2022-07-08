#!/bin/bash

ENABLE_DEBUG_SERVER="${ENABLE_DEBUG_SERVER:-false}"
DEPLOY_GRAFANA="${DEPLOY_GRAFANA:-false}"
DEPLOY_JAEGER="${DEPLOY_JAEGER:-false}"
DEPLOY_PROMETHEUS="${DEPLOY_PROMETHEUS:-false}"

./scripts/port-forward-bookbuyer-ui.sh &
./scripts/port-forward-bookstore-ui.sh &
./scripts/port-forward-bookstore-ui-v2.sh &
./scripts/port-forward-bookstore-ui-v1.sh &
./scripts/port-forward-bookthief-ui.sh &

if [ "$ENABLE_DEBUG_SERVER" = true ]; then
./scripts/port-forward-osm-debug.sh &
fi
if [ "$DEPLOY_GRAFANA" = true ]; then
./scripts/port-forward-grafana.sh &
fi
if [ "$DEPLOY_JAEGER" = true ]; then
./scripts/port-forward-jaeger.sh &
fi
if [ "$DEPLOY_PROMETHEUS" = true ]; then
./scripts/port-forward-prometheus.sh &
fi

wait

