CTR_REGISTRY ?= flomesh
CTR_TAG      ?= latest
DOCKER_BUILDX_OUTPUT ?= type=registry

DOCKER_REGISTRY ?= docker.io/library
UBUNTU_VERSION ?= 20.04
KERNEL_VERSION ?= v5.4
PYTHON_VERSION ?= 3.8
DOCKER_BUILDX_PLATFORM ?= linux/amd64
LDFLAGS ?= "-s -w"

UBUNTU_TARGETS = ubuntu compiler base
DOCKER_UBUNTU_TARGETS = $(addprefix docker-build-interceptor-, $(UBUNTU_TARGETS))

.PHONY: buildx-context
buildx-context:
	@if ! docker buildx ls | grep -q "^osm "; then docker buildx create --name osm --driver-opt network=host; fi

.PHONY: docker-build-interceptor-ubuntu
docker-build-interceptor-ubuntu:
	docker buildx build --builder osm \
	--platform=$(DOCKER_BUILDX_PLATFORM) \
	-o $(DOCKER_BUILDX_OUTPUT) \
	-t $(CTR_REGISTRY)/osm-edge-interceptor:ubuntu$(UBUNTU_VERSION) \
	-f ./dockerfiles/Dockerfile.osm-edge-interceptor-ubuntu \
	--build-arg DOCKER_REGISTRY=$(DOCKER_REGISTRY) \
	--build-arg UBUNTU_VERSION=$(UBUNTU_VERSION) \
	.

.PHONY: docker-build-cross-interceptor-ubuntu
docker-build-cross-interceptor-ubuntu: DOCKER_BUILDX_PLATFORM=linux/amd64,linux/arm64
docker-build-cross-interceptor-ubuntu: docker-build-interceptor-ubuntu

.PHONY: docker-build-interceptor-compiler
docker-build-interceptor-compiler:
	docker buildx build --builder osm \
	--platform=$(DOCKER_BUILDX_PLATFORM) \
	-o $(DOCKER_BUILDX_OUTPUT) \
	-t $(CTR_REGISTRY)/osm-edge-interceptor:compiler$(UBUNTU_VERSION) \
	-f ./dockerfiles/Dockerfile.osm-edge-interceptor-compiler \
	--build-arg CTR_REGISTRY=$(CTR_REGISTRY) \
	--build-arg CTR_TAG=$(UBUNTU_VERSION) \
	--build-arg KERNEL_VERSION=$(KERNEL_VERSION) \
	.

.PHONY: docker-build-cross-interceptor-compiler
docker-build-cross-interceptor-compiler: DOCKER_BUILDX_PLATFORM=linux/amd64,linux/arm64
docker-build-cross-interceptor-compiler: docker-build-interceptor-compiler

.PHONY: docker-build-interceptor-base
docker-build-interceptor-base:
	docker buildx build --builder osm \
	--platform=$(DOCKER_BUILDX_PLATFORM) \
	-o $(DOCKER_BUILDX_OUTPUT) \
	-t $(CTR_REGISTRY)/osm-edge-interceptor:base$(UBUNTU_VERSION) \
	-f ./dockerfiles/Dockerfile.osm-edge-interceptor-base \
	--build-arg CTR_REGISTRY=$(CTR_REGISTRY) \
	--build-arg CTR_TAG=$(UBUNTU_VERSION) \
	--build-arg PYTHON_VERSION=$(PYTHON_VERSION) \
	.

.PHONY: docker-build-cross-interceptor-base
docker-build-cross-interceptor-base: DOCKER_BUILDX_PLATFORM=linux/amd64,linux/arm64
docker-build-cross-interceptor-base: docker-build-interceptor-base

.PHONY: docker-build-osm
docker-build-osm: buildx-context $(DOCKER_UBUNTU_TARGETS)

.PHONY: docker-build-cross-osm
docker-build-cross-osm: DOCKER_BUILDX_PLATFORM=linux/amd64,linux/arm64
docker-build-cross-osm: docker-build-osm

.PHONY: trivy-ci-setup
trivy-ci-setup:
	wget https://github.com/aquasecurity/trivy/releases/download/v0.23.0/trivy_0.23.0_Linux-64bit.tar.gz
	tar zxvf trivy_0.23.0_Linux-64bit.tar.gz
	echo $$(pwd) >> $(GITHUB_PATH)

# Show all vulnerabilities in logs
trivy-scan-ubuntu-verbose-%: NAME=$(@:trivy-scan-ubuntu-verbose-%=%)
trivy-scan-ubuntu-verbose-%:
	trivy image "$(CTR_REGISTRY)/osm-edge-interceptor:$(NAME)$(UBUNTU_VERSION)"

# Exit if vulnerability exists
trivy-scan-ubuntu-fail-%: NAME=$(@:trivy-scan-ubuntu-fail-%=%)
trivy-scan-ubuntu-fail-%:
	trivy image --exit-code 1 --ignore-unfixed --severity MEDIUM,HIGH,CRITICAL "$(CTR_REGISTRY)/osm-edge-interceptor:$(NAME)$(UBUNTU_VERSION)"

.PHONY: trivy-scan-images trivy-scan-images-fail trivy-scan-images-verbose
trivy-scan-images-verbose: $(addprefix trivy-scan-ubuntu-verbose-, $(UBUNTU_TARGETS))
trivy-scan-images-fail: $(addprefix trivy-scan-ubuntu-fail-, $(UBUNTU_TARGETS))
trivy-scan-images: trivy-scan-images-verbose trivy-scan-images-fail