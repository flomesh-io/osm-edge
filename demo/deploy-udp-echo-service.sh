#!/bin/bash

# This script deploys the resources corresponding to the udp-echo service.

set -aueo pipefail

# shellcheck disable=SC1091
source .env

CTR_TAG="${CTR_TAG:-latest}"

echo -e "Create udp-echo service"
kubectl apply -f - <<EOF
apiVersion: v1
kind: Service
metadata:
  name: udp-echo
  namespace: udp-demo
  labels:
    app: udp-echo
spec:
  ports:
  - name: udp
    port: 6000
    appProtocol: udp
    protocol: UDP
  selector:
    app: udp-echo
EOF

echo -e "Create udp-echo service account"
kubectl apply -f - <<EOF
apiVersion: v1
kind: ServiceAccount
metadata:
  name: udp-echo
  namespace: udp-demo
EOF

echo -e "Create udp-echo deployment"
kubectl apply -f - <<EOF
apiVersion: apps/v1
kind: Deployment
metadata:
  name: udp-echo-v1
  namespace: udp-demo
  labels:
    app: udp-echo
    version: v1
spec:
  replicas: 1
  selector:
    matchLabels:
      app: udp-echo
      version: v1
  template:
    metadata:
      labels:
        app: udp-echo
        version: v1
    spec:
      serviceAccountName: udp-echo
      containers:
      - name: udp-echo-server
        image: "${CTR_REGISTRY}/osm-edge-demo-udp-echo-server:${CTR_TAG}"
        imagePullPolicy: Always
        command: ["/udp-echo-server"]
        args: [ "--port", "6000" ]
        ports:
        - containerPort: 6000
          name: udp-echo-server
          protocol: UDP
      imagePullSecrets:
        - name: $CTR_REGISTRY_CREDS_NAME
EOF
