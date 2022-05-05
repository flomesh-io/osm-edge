#!/bin/bash

set -euo pipefail

if [ -z "$1" ]; then
  echo "Error: expected one argument OS_ARCH"
  exit 1
fi

if [ -z "$2" ]; then
  echo "Error: expected one argument OS"
  exit 1
fi

BUILD_ARCH=$1
BUILD_OS=$2

sudo apt install -y apt-transport-https

sudo tee /etc/apt/sources.list.d/kubernetes.list <<-'EOF'
deb https://mirrors.aliyun.com/kubernetes/apt/ kubernetes-xenial main
EOF

curl https://mirrors.aliyun.com/kubernetes/apt/doc/apt-key.gpg | sudo apt-key add -

sudo apt -y update

sudo apt install -y kubelet=1.23.4-00 kubeadm=1.23.4-00 kubectl=1.23.4-00

sudo curl -Lo /usr/local/sbin/kind https://kind.sigs.k8s.io/dl/latest/kind-${BUILD_OS}-${BUILD_ARCH}
sudo chmod a+x /usr/local/sbin/kind

sudo snap install helm --classic