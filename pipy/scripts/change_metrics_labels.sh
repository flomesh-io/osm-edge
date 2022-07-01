#!/bin/bash

set -euo pipefail

if [ -z "$1" ]; then
  echo "Error: expected one argument OSM_HOME"
  exit 1
fi

if [ -z "$2" ]; then
  echo "Error: expected one argument OS_ARCH"
  exit 1
fi

if [ -z "$3" ]; then
  echo "Error: expected one argument SIDECAR"
  exit 1
fi

OSM_HOME=$1
BUILD_ARCH=$2
SIDECAR=$3


SRC=envoy
DST=sidecar

if [[ "${SIDECAR}" == "envoy" ]]; then
  SRC=sidecar
  DST=envoy
fi


find "${OSM_HOME}"/charts/osm/grafana/dashboards/ -type f -exec sed -i "s/${SRC}_cluster_external_upstream_rq/${DST}_cluster_external_upstream_rq/g" {} \;
find "${OSM_HOME}"/charts/osm/grafana/dashboards/ -type f -exec sed -i "s/${SRC}_cluster_external_upstream_rq_completed/${DST}_cluster_external_upstream_rq_completed/g" {} \;
find "${OSM_HOME}"/charts/osm/grafana/dashboards/ -type f -exec sed -i "s/${SRC}_cluster_external_upstream_rq_xx/${DST}_cluster_external_upstream_rq_xx/g" {} \;
find "${OSM_HOME}"/charts/osm/grafana/dashboards/ -type f -exec sed -i "s/${SRC}_cluster_name/${DST}_cluster_name/g" {} \;
find "${OSM_HOME}"/charts/osm/grafana/dashboards/ -type f -exec sed -i "s/${SRC}_cluster_upstream_cx_active/${DST}_cluster_upstream_cx_active/g" {} \;
find "${OSM_HOME}"/charts/osm/grafana/dashboards/ -type f -exec sed -i "s/${SRC}_cluster_upstream_cx_connect_timeout/${DST}_cluster_upstream_cx_connect_timeout/g" {} \;
find "${OSM_HOME}"/charts/osm/grafana/dashboards/ -type f -exec sed -i "s/${SRC}_cluster_upstream_cx_destroy_local_with_active_rq/${DST}_cluster_upstream_cx_destroy_local_with_active_rq/g" {} \;
find "${OSM_HOME}"/charts/osm/grafana/dashboards/ -type f -exec sed -i "s/${SRC}_cluster_upstream_cx_destroy_remote_with_active_rq/${DST}_cluster_upstream_cx_destroy_remote_with_active_rq/g" {} \;
find "${OSM_HOME}"/charts/osm/grafana/dashboards/ -type f -exec sed -i "s/${SRC}_cluster_upstream_cx_rx_bytes_total/${DST}_cluster_upstream_cx_rx_bytes_total/g" {} \;
find "${OSM_HOME}"/charts/osm/grafana/dashboards/ -type f -exec sed -i "s/${SRC}_cluster_upstream_cx_tx_bytes_total/${DST}_cluster_upstream_cx_tx_bytes_total/g" {} \;
find "${OSM_HOME}"/charts/osm/grafana/dashboards/ -type f -exec sed -i "s/${SRC}_cluster_upstream_rq_pending_failure_eject/${DST}_cluster_upstream_rq_pending_failure_eject/g" {} \;
find "${OSM_HOME}"/charts/osm/grafana/dashboards/ -type f -exec sed -i "s/${SRC}_cluster_upstream_rq_pending_overflow/${DST}_cluster_upstream_rq_pending_overflow/g" {} \;
find "${OSM_HOME}"/charts/osm/grafana/dashboards/ -type f -exec sed -i "s/${SRC}_cluster_upstream_rq_rx_reset/${DST}_cluster_upstream_rq_rx_reset/g" {} \;
find "${OSM_HOME}"/charts/osm/grafana/dashboards/ -type f -exec sed -i "s/${SRC}_cluster_upstream_rq_timeout/${DST}_cluster_upstream_rq_timeout/g" {} \;
find "${OSM_HOME}"/charts/osm/grafana/dashboards/ -type f -exec sed -i "s/${SRC}_cluster_upstream_rq_total/${DST}_cluster_upstream_rq_total/g" {} \;
find "${OSM_HOME}"/charts/osm/grafana/dashboards/ -type f -exec sed -i "s/${SRC}_cluster_upstream_rq_tx_reset/${DST}_cluster_upstream_rq_tx_reset/g" {} \;
find "${OSM_HOME}"/charts/osm/grafana/dashboards/ -type f -exec sed -i "s/${SRC}_cluster_upstream_rq_xx/${DST}_cluster_upstream_rq_xx/g" {} \;
find "${OSM_HOME}"/charts/osm/grafana/dashboards/ -type f -exec sed -i "s/${SRC}_response_code/${DST}_response_code/g" {} \;
find "${OSM_HOME}"/charts/osm/grafana/dashboards/ -type f -exec sed -i "s/${SRC}_response_code_class/${DST}_response_code_class/g" {} \;
find "${OSM_HOME}"/charts/osm/grafana/dashboards/ -type f -exec sed -i "s/${SRC}_server_live/${DST}_server_live/g" {} \;


find "${OSM_HOME}"/charts/osm/templates/ -name prometheus-configmap.yaml -type f -exec sed -i "s/${SRC}_cluster_external_upstream_rq/${DST}_cluster_external_upstream_rq/g" {} \;
find "${OSM_HOME}"/charts/osm/templates/ -name prometheus-configmap.yaml -type f -exec sed -i "s/${SRC}_cluster_external_upstream_rq_completed/${DST}_cluster_external_upstream_rq_completed/g" {} \;
find "${OSM_HOME}"/charts/osm/templates/ -name prometheus-configmap.yaml -type f -exec sed -i "s/${SRC}_cluster_external_upstream_rq_xx/${DST}_cluster_external_upstream_rq_xx/g" {} \;
find "${OSM_HOME}"/charts/osm/templates/ -name prometheus-configmap.yaml -type f -exec sed -i "s/${SRC}_cluster_name/${DST}_cluster_name/g" {} \;
find "${OSM_HOME}"/charts/osm/templates/ -name prometheus-configmap.yaml -type f -exec sed -i "s/${SRC}_cluster_upstream_cx_active/${DST}_cluster_upstream_cx_active/g" {} \;
find "${OSM_HOME}"/charts/osm/templates/ -name prometheus-configmap.yaml -type f -exec sed -i "s/${SRC}_cluster_upstream_cx_connect_timeout/${DST}_cluster_upstream_cx_connect_timeout/g" {} \;
find "${OSM_HOME}"/charts/osm/templates/ -name prometheus-configmap.yaml -type f -exec sed -i "s/${SRC}_cluster_upstream_cx_destroy_local_with_active_rq/${DST}_cluster_upstream_cx_destroy_local_with_active_rq/g" {} \;
find "${OSM_HOME}"/charts/osm/templates/ -name prometheus-configmap.yaml -type f -exec sed -i "s/${SRC}_cluster_upstream_cx_destroy_remote_with_active_rq/${DST}_cluster_upstream_cx_destroy_remote_with_active_rq/g" {} \;
find "${OSM_HOME}"/charts/osm/templates/ -name prometheus-configmap.yaml -type f -exec sed -i "s/${SRC}_cluster_upstream_cx_rx_bytes_total/${DST}_cluster_upstream_cx_rx_bytes_total/g" {} \;
find "${OSM_HOME}"/charts/osm/templates/ -name prometheus-configmap.yaml -type f -exec sed -i "s/${SRC}_cluster_upstream_cx_tx_bytes_total/${DST}_cluster_upstream_cx_tx_bytes_total/g" {} \;
find "${OSM_HOME}"/charts/osm/templates/ -name prometheus-configmap.yaml -type f -exec sed -i "s/${SRC}_cluster_upstream_rq_pending_failure_eject/${DST}_cluster_upstream_rq_pending_failure_eject/g" {} \;
find "${OSM_HOME}"/charts/osm/templates/ -name prometheus-configmap.yaml -type f -exec sed -i "s/${SRC}_cluster_upstream_rq_pending_overflow/${DST}_cluster_upstream_rq_pending_overflow/g" {} \;
find "${OSM_HOME}"/charts/osm/templates/ -name prometheus-configmap.yaml -type f -exec sed -i "s/${SRC}_cluster_upstream_rq_rx_reset/${DST}_cluster_upstream_rq_rx_reset/g" {} \;
find "${OSM_HOME}"/charts/osm/templates/ -name prometheus-configmap.yaml -type f -exec sed -i "s/${SRC}_cluster_upstream_rq_timeout/${DST}_cluster_upstream_rq_timeout/g" {} \;
find "${OSM_HOME}"/charts/osm/templates/ -name prometheus-configmap.yaml -type f -exec sed -i "s/${SRC}_cluster_upstream_rq_total/${DST}_cluster_upstream_rq_total/g" {} \;
find "${OSM_HOME}"/charts/osm/templates/ -name prometheus-configmap.yaml -type f -exec sed -i "s/${SRC}_cluster_upstream_rq_tx_reset/${DST}_cluster_upstream_rq_tx_reset/g" {} \;
find "${OSM_HOME}"/charts/osm/templates/ -name prometheus-configmap.yaml -type f -exec sed -i "s/${SRC}_cluster_upstream_rq_xx/${DST}_cluster_upstream_rq_xx/g" {} \;
find "${OSM_HOME}"/charts/osm/templates/ -name prometheus-configmap.yaml -type f -exec sed -i "s/${SRC}_response_code/${DST}_response_code/g" {} \;
find "${OSM_HOME}"/charts/osm/templates/ -name prometheus-configmap.yaml -type f -exec sed -i "s/${SRC}_response_code_class/${DST}_response_code_class/g" {} \;
find "${OSM_HOME}"/charts/osm/templates/ -name prometheus-configmap.yaml -type f -exec sed -i "s/${SRC}_server_live/${DST}_server_live/g" {} \;


sed -i 's/container=\\"'"$SRC"'\\"/container=\\"'"$DST"'\\"/g' "${OSM_HOME}"/charts/osm/grafana/dashboards/osm-data-plane-performance.json

if [ -f "${OSM_HOME}/charts/osm/grafana/dashboards/osm-mesh-envoy-details.json" ]; then
  mv "${OSM_HOME}"/charts/osm/grafana/dashboards/osm-mesh-envoy-details.json "${OSM_HOME}"/charts/osm/grafana/dashboards/osm-mesh-sidecar-details.json
fi


