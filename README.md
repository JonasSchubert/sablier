# Sablier

[![GitHub license](https://img.shields.io/github/license/sablierapp/sablier.svg)](https://github.com/sablierapp/sablier/blob/master/LICENSE)
[![GitHub contributors](https://img.shields.io/github/contributors/sablierapp/sablier.svg)](https://GitHub.com/sablierapp/sablier/graphs/contributors/)
[![GitHub issues](https://img.shields.io/github/issues/sablierapp/sablier.svg)](https://GitHub.com/sablierapp/sablier/issues/)
[![GitHub pull-requests](https://img.shields.io/github/issues-pr/sablierapp/sablier.svg)](https://GitHub.com/sablierapp/sablier/pulls/)
[![PRs Welcome](https://img.shields.io/badge/PRs-welcome-brightgreen.svg?style=flat-square)](http://makeapullrequest.com)

[![GoDoc](https://godoc.org/github.com/sablierapp/sablier?status.svg)](http://godoc.org/github.com/sablierapp/sablier)
![Latest Build](https://img.shields.io/github/actions/workflow/status/sablierapp/sablier/build.yml?style=flat-square&branch=main)
![Go Report](https://goreportcard.com/badge/github.com/sablierapp/sablier?style=flat-square)
![Go Version](https://img.shields.io/github/go-mod/go-version/sablierapp/sablier?style=flat-square)
![Latest Release](https://img.shields.io/github/v/release/sablierapp/sablier?style=flat-square&sort=semver)
![Latest PreRelease](https://img.shields.io/github/v/release/sablierapp/sablier?style=flat-square&include_prereleases&sort=semver)

An free and open-source software to start workloads on demand and stop them after a period of inactivity.

![Demo](./docs/assets/img/demo.gif)

Either because you don't want to overload your raspberry pi or because your QA environment gets used only once a week and wastes resources by keeping your workloads up and running, Sablier is a project that might interest you.

## 🎯 Features

- [Supports the following providers](https://sablierapp.dev/#/providers/overview)
  - Docker
  - Docker Swarm
  - Kubernetes
- [Supports multiple reverse proxies](https://sablierapp.dev/#/plugins/overview)
  - Apache APISIX
  - Caddy
  - Envoy
  - Istio
  - Nginx (NJS Module)
  - Nginx (WASM Module)
  - Traefik
- Scale up your workload automatically upon the first request
  - [with a themable waiting page](https://sablierapp.dev/#/themes)
  - [with a hanging request (hang until service is up)](https://sablierapp.dev/#/strategies?id=blocking-strategy)
- Scale your workload to zero automatically after a period of inactivity

## 📝 Documentation

[See the documentation here](https://sablierapp.dev)
