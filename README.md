# Open Service Mesh Edge (OSM-Edge)

[![build](https://github.com/flomesh-io/osm-edge/workflows/Go/badge.svg)](https://github.com/flomesh-io/osm-edge/actions?query=workflow%3AGo)
[![report](https://goreportcard.com/badge/github.com/flomesh-io/osm-edge)](https://goreportcard.com/report/github.com/flomesh-io/osm-edge)
[![codecov](https://codecov.io/gh/flomesh-io/osm-edge/branch/main/graph/badge.svg)](https://codecov.io/gh/flomesh-io/osm-edge)
[![License: MIT](https://img.shields.io/badge/License-MIT-yellow.svg)](https://github.com/flomesh-io/osm-edge/blob/main/LICENSE)
[![release](https://img.shields.io/github/release/flomesh-io/osm-edge/all.svg)](https://github.com/flomesh-io/osm-edge/releases)

Open Service Mesh Edge (OSM-Edge) fork from [Open Service Mesh](https://github.com/openservicemesh/osm) is a lightweight, extensible, Cloud Native [service mesh][1] built purposely for Edge computing. OSM edge uses lightweight, Cloud Native, programmable proxy [Pipy](https://flomesh.io) as sidecar proxy.


The OSM project builds on the ideas and implementations of many cloud native ecosystem projects including [Linkerd](https://github.com/linkerd/linkerd), [Istio](https://github.com/istio/istio), [Consul](https://github.com/hashicorp/consul), [Pipy](https://github.com/flomesh-io/pipy), [Envoy](https://github.com/envoyproxy/envoy), [Kuma](https://github.com/kumahq/kuma), [Helm](https://github.com/helm/helm), and the [SMI](https://github.com/servicemeshinterface/smi-spec) specification.

## Table of Contents
- [Open Service Mesh Edge (OSM-Edge)](#open-service-mesh-edge-osm-edge)
  - [Table of Contents](#table-of-contents)
  - [Overview](#overview)
    - [Core Principles](#core-principles)
    - [Documentation](#documentation)
    - [Features](#features)
    - [Project status](#project-status)
    - [Support](#support)
    - [SMI Specification support](#smi-specification-support)
  - [OSM Design](#osm-design)
  - [Install](#install)
    - [Prerequisites](#prerequisites)
    - [Get the OSM CLI](#get-the-osm-cli)
    - [Install OSM Edge](#install-osm-edge)
  - [Demonstration](#demonstration)
  - [Using OSM Edge](#using-osm-edge)
    - [Quick Start](#quick-start)
    - [OSM Usage Patterns](#osm-usage-patterns)
  - [Community](#community)
  - [Development Guide](#development-guide)
  - [Code of Conduct](#code-of-conduct)
  - [License](#license)


## Overview

OSM-Edge runs an Sidecar based control plane on Kubernetes, can be configured with SMI APIs, and works by injecting a [Pipy](https://flomesh.io) Sidecar proxy as a sidecar container next to each instance of your application. The proxy contains and executes rules around access control policies, implements routing configuration, and captures metrics. The control plane continually configures proxies to ensure policies and routing rules are up to date and ensures proxies are healthy.

### Core Principles
1. Simple to understand and contribute to
1. Effortless to install, maintain, and operate
1. Painless to troubleshoot
1. Easy to configure via [Service Mesh Interface (SMI)][2]

### Documentation
Documentation pertaining to the usage of Open Service Mesh Edge is made available at [osm-edge-docs.flomesh.io](https://osm-edge-docs.flomesh.io/).

Documentation pertaining to development, release workflows, and other repository specific documentation, can be found in the [docs folder](/docs).

### Features

1. Easily and transparently configure [traffic shifting][3] for deployments
1. Secure service to service communication by [enabling mTLS](https://osm-edge-docs.flomesh.io/docs/guides/certificates/)
1. Define and execute fine grained [access control][4] policies for services
1. [Observability](https://osm-edge-docs.flomesh.io/docs/troubleshooting/observability/) and insights into application metrics for debugging and monitoring services
1. Integrate with [external certificate management](https://osm-edge-docs.flomesh.io/docs/guides/certificates/) services/solutions with a pluggable interface
1. Onboard applications onto the mesh by enabling [automatic sidecar injection](https://osm-edge-docs.flomesh.io/docs/guides/app_onboarding/sidecar_injection/) of Sidecar proxy

### Project status

OSM-Edge is under active development and is ready for production workloads.

### Support

[Please search open issues on GitHub](https://github.com/flomesh-io/osm-edge/issues), and if your issue isn't already represented please [open a new one](https://github.com/flomesh-io/osm-edge/issues/new/choose). The OSM project maintainers will respond to the best of their abilities.

### SMI Specification support

|   Kind    | SMI Resource |         Supported Version          |          Comments          |
| :---------------------------- | - | :--------------------------------: |  :--------------------------------: |
| TrafficTarget  | traffictargets.access.smi-spec.io |  [v1alpha3](https://github.com/servicemeshinterface/smi-spec/blob/v0.6.0/apis/traffic-access/v1alpha3/traffic-access.md)  | |
| HTTPRouteGroup | httproutegroups.specs.smi-spec.io | [v1alpha4](https://github.com/servicemeshinterface/smi-spec/blob/v0.6.0/apis/traffic-specs/v1alpha4/traffic-specs.md#httproutegroup) | |
| TCPRoute | tcproutes.specs.smi-spec.io | [v1alpha4](https://github.com/servicemeshinterface/smi-spec/blob/v0.6.0/apis/traffic-specs/v1alpha4/traffic-specs.md#tcproute) | |
| UDPRoute | udproutes.specs.smi-spec.io | [v1alpha4](https://github.com/servicemeshinterface/smi-spec/blob/v0.6.0/apis/traffic-specs/v1alpha4/traffic-specs.md#udproute) | |
| TrafficSplit | trafficsplits.split.smi-spec.io | [v1alpha2](https://github.com/servicemeshinterface/smi-spec/blob/v0.6.0/apis/traffic-split/v1alpha2/traffic-split.md) | |
| TrafficMetrics  | \*.metrics.smi-spec.io | [v1alpha1](https://github.com/servicemeshinterface/smi-spec/blob/v0.6.0/apis/traffic-metrics/v1alpha1/traffic-metrics.md) | 🚧 **In Progress** 🚧 |

## OSM Design

Read more about [OSM's high level goals, design, and architecture](DESIGN.md).

## Install

### Prerequisites
- Kubernetes cluster running Kubernetes v1.20.0 or greater
- kubectl current context is configured for the target cluster install
  - ```kubectl config current-context```

### Get the OSM CLI

The simplest way of installing Open Service Mesh on a Kubernetes cluster is by using the `osm` CLI.

Download the `osm` binary from the [Releases page](https://github.com/flomesh-io/osm-edge/releases). Unpack the `osm` binary and add it to `$PATH` to get started.
```shell
sudo mv ./osm /usr/local/bin/osm
```

### Install OSM Edge
```shell
$ osm install
```
![OSM Install Demo](img/osm-install-demo-v0.9.2.gif "OSM Install Demo")

See the [installation guide](https://osm-edge-docs.flomesh.io/docs/guides/install/) for more detailed options.

## Demonstration

The OSM [Bookstore demo](https://osm-edge-docs.flomesh.io/docs/getting_started/) is a step-by-step walkthrough of how to install a bookbuyer and bookstore apps, and configure connectivity between these using SMI.

## Using OSM Edge

After installing OSM, [onboard a microservice application](https://osm-edge-docs.flomesh.io/docs/guides/app_onboarding/) to the service mesh.

### Quick Start

Refer to [Quick Start](https://osm-edge-docs.flomesh.io/docs/quickstart/) guide for step-by-step guide on how to start quickly.

### OSM Usage Patterns

1. [Traffic Management](https://osm-edge-docs.flomesh.io/docs/guides/traffic_management/)
1. [Observability](https://osm-edge-docs.flomesh.io/docs/troubleshooting/observability/)
1. [Certificates](https://osm-edge-docs.flomesh.io/docs/guides/certificates/)
1. [Sidecar Injection](https://osm-edge-docs.flomesh.io/docs/guides/app_onboarding/sidecar_injection/)


## Community

Connect with the Open Service Mesh community:

- GitHub [issues](https://github.com/flomesh-io/osm-edge/issues) and [pull requests](https://github.com/flomesh-io/osm-edge/pulls) in this repo
- OSM-Edge Slack: <a href="https://join.slack.com/t/flomesh-io/shared_invite/zt-16f4yv2hc-qvEgSrMATKn5LjmDAwzlbw">Join</a> the Flomesh-io Slack for related discussions

## Development Guide

If you would like to contribute to OSM, check out the [development guide](docs/development_guide/README.md).

## Code of Conduct

This project has adopted the [CNCF Code of Conduct](https://github.com/cncf/foundation/blob/master/code-of-conduct.md). See [CODE_OF_CONDUCT.md](CODE_OF_CONDUCT.md) for further details.

## License

This software is covered under the Apache 2.0 license. You can read the license [here](LICENSE).


[1]: https://en.wikipedia.org/wiki/Service_mesh
[2]: https://github.com/servicemeshinterface/smi-spec/blob/master/SPEC_LATEST_STABLE.md
[3]: https://github.com/servicemeshinterface/smi-spec/blob/v0.6.0/apis/traffic-split/v1alpha2/traffic-split.md
[4]: https://github.com/servicemeshinterface/smi-spec/blob/v0.6.0/apis/traffic-access/v1alpha3/traffic-access.md
