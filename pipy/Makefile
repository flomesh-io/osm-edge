#!/bin/bash

ARCH_MAP_x86_64 := amd64
ARCH_MAP_arm64 := arm64
ARCH_MAP_aarch64 := arm64

BUILDARCH := $(ARCH_MAP_$(shell uname -m))
BUILDOS := $(shell uname -s | tr '[:upper:]' '[:lower:]')

TARGETS := $(BUILDOS)/$(BUILDARCH)
DOCKER_BUILDX_PLATFORM := $(BUILDOS)/$(BUILDARCH)

OSM_HOME ?= $(dir $(abspath $(dir $$PWD)))

default: build

install-docker:
	scripts/install-docker.sh

k8s-tools: install-docker
	scripts/install-k8s-tools.sh ${BUILDARCH} ${BUILDOS}

golang:
	make -f scripts/Makefile.golang

.env:
	scripts/env.sh ${OSM_HOME} ${BUILDARCH} ${BUILDOS}

secret:
	scripts/secret.sh ${OSM_HOME}

goproxy:
	@sed -i 's/CH go/CH GO111MODULE=on GOPROXY=https:\/\/goproxy.cn go/g' ../dockerfiles/Dockerfile.demo
	@sed -i 's/CH go/CH GO111MODULE=on GOPROXY=https:\/\/goproxy.cn go/g' ../dockerfiles/Dockerfile.osm-bootstrap
	@sed -i 's/CH go/CH GO111MODULE=on GOPROXY=https:\/\/goproxy.cn go/g' ../dockerfiles/Dockerfile.osm-injector
	@sed -i 's/CH go/CH GO111MODULE=on GOPROXY=https:\/\/goproxy.cn go/g' ../dockerfiles/Dockerfile.osm-controller

disable-goproxy:
	@sed -i 's/CH GO111MODULE=on GOPROXY=https:\/\/goproxy.cn go/CH go/g' ../dockerfiles/Dockerfile.demo
	@sed -i 's/CH GO111MODULE=on GOPROXY=https:\/\/goproxy.cn go/CH go/g' ../dockerfiles/Dockerfile.osm-bootstrap
	@sed -i 's/CH GO111MODULE=on GOPROXY=https:\/\/goproxy.cn go/CH go/g' ../dockerfiles/Dockerfile.osm-injector
	@sed -i 's/CH GO111MODULE=on GOPROXY=https:\/\/goproxy.cn go/CH go/g' ../dockerfiles/Dockerfile.osm-controller

adapter-os-arch:
	scripts/adapter-os-arch.sh ${OSM_HOME} ${BUILDARCH}

disable-wasm-stats:
	scripts/disable-wasm-stats.sh ${OSM_HOME} ${BUILDARCH}

enable-wasm-stats:
	scripts/enable-wasm-stats.sh ${OSM_HOME} ${BUILDARCH}

cache-osm-images:
	scripts/cache-osm-images.sh ${OSM_HOME} ${BUILDARCH}

cancel-cache-osm-images:
	scripts/cancel-cache-osm-images.sh ${OSM_HOME} ${BUILDARCH}

cache-demo-images:
	scripts/cache-demo-images.sh ${OSM_HOME} ${BUILDARCH}

cancel-cache-demo-images:
	scripts/cancel-cache-demo-images.sh ${OSM_HOME} ${BUILDARCH}

cache-test-e2e-images:
	scripts/cache-test-e2e-images.sh ${OSM_HOME} ${BUILDARCH}

cancel-cache-test-e2e-images:
	scripts/cancel-cache-test-e2e-images.sh ${OSM_HOME} ${BUILDARCH}

switch-sidecar-to-pipy: disable-wasm-stats
	scripts/switch-sidecar.sh ${OSM_HOME} ${BUILDARCH} pipy

switch-sidecar-to-envoy: enable-wasm-stats
	scripts/switch-sidecar.sh ${OSM_HOME} ${BUILDARCH} envoy

build-osm-images:
	scripts/build-osm-images.sh ${OSM_HOME}

digest-charts-osm-images:
	scripts/digest-charts-osm-images.sh ${OSM_HOME}

build-osm-cli:
	scripts/build-osm-cli.sh ${OSM_HOME}

enable-port-forward-addr:
	scripts/enable-port-forward-addr.sh ${OSM_HOME}

disable-test-e2e-docker-build:
	scripts/disable-test-e2e-docker-build.sh ${OSM_HOME}

once: .env secret adapter-os-arch
	@echo
	@echo "Please execute \"\033[1;32;40msource ~/.bashrc\033[0m\""
	@echo

cache: cache-osm-images cache-demo-images cache-test-e2e-images

cancel-cache: cancel-cache-osm-images cancel-cache-demo-images cancel-cache-test-e2e-images

pipy: switch-sidecar-to-pipy

envoy: switch-sidecar-to-envoy

build: build-osm-images digest-charts-osm-images build-osm-cli