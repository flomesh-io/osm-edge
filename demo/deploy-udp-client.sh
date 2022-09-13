#!/bin/bash

# This script deploys the resources corresponding to the udp-client.

set -aueo pipefail

# shellcheck disable=SC1091
source .env

CTR_TAG="${CTR_TAG:-latest}"

echo -e "Create udp-client service account"
kubectl apply -f - <<EOF
apiVersion: v1
kind: ServiceAccount
metadata:
  name: udp-client
  namespace: udp-demo
EOF

echo -e "Create udp-client deployment"
kubectl apply -f - <<EOF
apiVersion: apps/v1
kind: Deployment
metadata:
  name: udp-client-v1
  namespace: udp-demo
  labels:
    app: udp-client
    version: v1
spec:
  replicas: 1
  selector:
    matchLabels:
      app: udp-client
      version: v1
  template:
    metadata:
      labels:
        app: udp-client
        version: v1
    spec:
      serviceAccountName: udp-client
      containers:
      - name: tcp-client
        image: "${CTR_REGISTRY}/osm-edge-demo-udp-client:${CTR_TAG}"
        imagePullPolicy: Always
        command: ["/udp-client"]
        args: [ "udp-echo", "6000", "hello world."]
      imagePullSecrets:
        - name: $CTR_REGISTRY_CREDS_NAME
EOF
