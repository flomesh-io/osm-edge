#!/bin/bash

set -uo pipefail

if [ -z "$1" ]; then
  echo "Error: expected one argument OSM_HOME"
  exit 1
fi

OSM_HOME=$1

cd ${OSM_HOME}

allCases=(
"CertManagerSimpleClientServer"
"SimpleClientServer traffic test involving osm-controller restart: HTTP"
"DebugServer"
"DeploymentsClientServer"
"HTTP egress policy without route matches"
"HTTP egress policy with route match"
"HTTPS egress policy"
"TCP egress policy"
"Egress"
"Fluent Bit deployment"
"Fluent Bit output"
"Garbage Collection"
"gRPC insecure traffic origination over HTTP2 with SMI HTTP routes"
"gRPC secure traffic origination over HTTP2 with SMI TCP routes"
"HashivaultSimpleClientServer"
"Test health probes can succeed"
"Helm install using default values"
"Ignore Label"
"HTTP ingress with IngressBackend"
"When OSM is Installed"
"Test IP range exclusion"
"Version v1.22.8"
"Version v1.21.11"
"#Custom WASM metrics between one client pod and one server"
"Multiple service ports"
"Multiple services matching same pod"
"Becomes ready after being reinstalled"
"PermissiveToSmiSwitching"
"Permissive mode HTTP test with a Kubernetes Service for the Source"
"Permissive mode HTTP test without a Kubernetes Service for the Source"
"Test traffic flowing from client to server with a Kubernetes Service for the Source: HTTP"
"Test traffic flowing from client to server without a Kubernetes Service for the Source: HTTP"
"Test global port exclusion"
"Test pod level port exclusion"
"proxy resources"
"Enable Reconciler"
"SMI TrafficTarget is set up properly"
"SMI Traffic Target is not in the same namespace as the destination"
"SimpleClientServer TCP with SMI policies"
"SimpleClientServer TCP in permissive mode"
"SimpleClientServer egress TCP"
"TCP server-first traffic"
"HTTP recursive traffic splitting with SMI"
"TCP recursive traffic splitting with SMI"
"ClientServerTrafficSplitSameSA"
"HTTP traffic splitting - root service selector matches backends"
"HTTP traffic splitting with SMI"
"TCP traffic splitting with SMI"
"HTTP traffic splitting with Permissive mode"
"#Tests upgrading the control plane"
"With SMI Traffic Target validation enabled"
"With SMI validation disabled"
)

# shellcheck disable=SC2068
for item in "${allCases[@]}"; do
  echo -e "Testing $item ..."
  E2E_FLAGS="-ginkgo.focus='$item' --timeout=0" make test-e2e 2>/dev/null | grep 'Passed.*Failed.*Skipped'
done
