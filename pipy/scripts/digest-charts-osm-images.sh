#!/bin/bash

set -euo pipefail

if [ -z "$1" ]; then
  echo "Error: expected one argument OSM_HOME"
  exit 1
fi

OSM_HOME=$1

OsmControllerDigest=`docker buildx imagetools inspect ${CTR_REGISTRY}/osm-controller:${CTR_TAG} --raw | sha256sum | awk '{print $1}'`
OsmInjectorDigest=`docker buildx imagetools inspect ${CTR_REGISTRY}/osm-injector:${CTR_TAG} --raw | sha256sum | awk '{print $1}'`
OsmInitDigest=`docker buildx imagetools inspect ${CTR_REGISTRY}/init:${CTR_TAG} --raw | sha256sum | awk '{print $1}'`
OsmCrdsDigest=`docker buildx imagetools inspect ${CTR_REGISTRY}/osm-crds:${CTR_TAG} --raw | sha256sum | awk '{print $1}'`
OsmBootstrapDigest=`docker buildx imagetools inspect ${CTR_REGISTRY}/osm-bootstrap:${CTR_TAG} --raw | sha256sum | awk '{print $1}'`
OsmPreinstallDigest=`docker buildx imagetools inspect ${CTR_REGISTRY}/osm-preinstall:${CTR_TAG} --raw | sha256sum | awk '{print $1}'`
OsmHealthcheckDigest=`docker buildx imagetools inspect ${CTR_REGISTRY}/osm-healthcheck:${CTR_TAG} --raw | sha256sum | awk '{print $1}'`

sed -i "s#osmController: \".*\"#osmController: \"sha256:${OsmControllerDigest}\"#g" ${OSM_HOME}/charts/osm/values.yaml
sed -i "s/osmInjector: \".*\"/osmInjector: \"sha256:${OsmInjectorDigest}\"/g" ${OSM_HOME}/charts/osm/values.yaml
sed -i "s/osmSidecarInit: \".*\"/osmSidecarInit: \"sha256:${OsmInitDigest}\"/g" ${OSM_HOME}/charts/osm/values.yaml
sed -i "s/osmCRDs: \".*\"/osmCRDs: \"sha256:${OsmCrdsDigest}\"/g" ${OSM_HOME}/charts/osm/values.yaml
sed -i "s/osmBootstrap: \".*\"/osmBootstrap: \"sha256:${OsmBootstrapDigest}\"/g" ${OSM_HOME}/charts/osm/values.yaml
sed -i "s/osmPreinstall: \".*\"/osmPreinstall: \"sha256:${OsmPreinstallDigest}\"/g" ${OSM_HOME}/charts/osm/values.yaml
sed -i "s/osmHealthcheck: \".*\"/osmHealthcheck: \"sha256:${OsmHealthcheckDigest}\"/g" ${OSM_HOME}/charts/osm/values.yaml
