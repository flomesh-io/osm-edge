#!/bin/bash

set -aueo pipefail


# shellcheck disable=SC1091
source .env

BOOKBUYER_NAMESPACE="${BOOKBUYER_NAMESPACE:-bookbuyer}"
BOOKSTORE_NAMESPACE="${BOOKSTORE_NAMESPACE:-bookstore}"
BOOKSTORE_SVC="bookstore-v1"
KUBERNETES_NODE_ARCH="${KUBERNETES_NODE_ARCH:-amd64}"
KUBERNETES_NODE_OS="${KUBERNETES_NODE_OS:-linux}"

action=$1

if [ "$action" != "apply" ] && [ "$action" != "delete" ]; then
  echo "Must pass in 'apply' or 'delete' as parameter.

Usage: multicluster-fault-inject.sh [apply|delete]

apply:  inject faults to bookstore service
delete: remove faults from bookstore service"
  exit 1
fi

kubectl "$action" -f - <<EOF
apiVersion: apps/v1
kind: Deployment
metadata:
  name: errors
  namespace: $BOOKSTORE_NAMESPACE
spec:
  replicas: 4
  selector:
    matchLabels:
      app: $BOOKSTORE_SVC
      purpose: fault
  template:
    metadata:
      labels:
        app: $BOOKSTORE_SVC
        purpose: fault
      annotations:
        openservicemesh.io/sidecar-injection: "false"
    spec:
      serviceAccountName: "$BOOKSTORE_SVC"
      nodeSelector:
        kubernetes.io/arch: ${KUBERNETES_NODE_ARCH}
        kubernetes.io/os: ${KUBERNETES_NODE_OS}
      containers:
        - image: curlimages/curl
          imagePullPolicy: IfNotPresent
          name: curl
          command: ["sleep", "365d"]
EOF
