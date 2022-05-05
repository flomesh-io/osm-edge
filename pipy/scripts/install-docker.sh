#!/bin/bash

sudo apt -y update
sudo apt -y remove docker docker-engine docker.io containerd runc

curl -fsSL https://download.docker.com/linux/ubuntu/gpg | sudo gpg --dearmor -o /usr/share/keyrings/docker-archive-keyring.gpg

echo \
  "deb [arch=$(dpkg --print-architecture) signed-by=/usr/share/keyrings/docker-archive-keyring.gpg] https://download.docker.com/linux/ubuntu \
  $(lsb_release -cs) stable" | sudo tee /etc/apt/sources.list.d/docker.list > /dev/null

sudo apt -y update
sudo apt -y install ca-certificates curl gnupg lsb-release
sudo apt -y install docker-ce docker-ce-cli containerd.io

sudo groupadd docker
sudo gpasswd -a $USER docker
sudo systemctl restart docker
sudo newgrp docker &

#Turn off swap
sudo swapoff -a
sudo sed -ri 's/.*swap.*/#&/' /etc/fstab

sudo tee /etc/modprobe.d/nf_conntrack.conf <<-'EOF'
#hashsize=nf_conntrack_max/8
options nf_conntrack hashsize=16384
EOF

#sudo sed -i '$anet.netfilter.nf_conntrack_max = 131072' /etc/sysctl.conf

#Iptables allowed bridge flow check
cat <<EOF | sudo tee /etc/modules-load.d/k8s.conf
br_netfilter
EOF

cat <<EOF | sudo tee /etc/sysctl.d/k8s.conf
net.bridge.bridge-nf-call-ip6tables = 1
net.bridge.bridge-nf-call-iptables = 1
EOF

sudo sysctl --system