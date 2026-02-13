# Changelog

All notable changes to this project will be documented in this file.

The format is based on [Keep a Changelog](https://keepachangelog.com/en/1.1.0/),
and this project adheres to [Semantic Versioning](https://semver.org/spec/v2.0.0.html).

## [0.4.0](https://github.com/openclaw-rocks/k8s-operator/compare/v0.3.0...v0.4.0) (2026-02-13)


### ⚠ BREAKING CHANGES

* CRD API group changed from `openclaw.openclaw.io` to `openclaw.openclaw.rocks`. Existing CRDs must be deleted and re-created. This is acceptable at v1alpha1 stability level.

### Features

* Add nautical banner image with crab captain at Kubernetes helm ([84d1854](https://github.com/openclaw-rocks/k8s-operator/commit/84d1854cb7200587d6745c9ae3dc29c049fc3b8e))
* update banner with real OpenClaw logo and Kubernetes logo ([#37](https://github.com/openclaw-rocks/k8s-operator/issues/37)) ([840adb9](https://github.com/openclaw-rocks/k8s-operator/commit/840adb9e285c14dffe1bfed17e9b9411b05f4538))


### Bug Fixes

* Apply same CreateOrUpdate pattern to ServiceMonitor reconciler ([#30](https://github.com/openclaw-rocks/k8s-operator/issues/30)) ([c1aaa36](https://github.com/openclaw-rocks/k8s-operator/commit/c1aaa3639b7cb82eb95696bbe06b98817d46a55e))
* change CRD API group domain from openclaw.io to openclaw.rocks ([#41](https://github.com/openclaw-rocks/k8s-operator/issues/41)) ([5bae852](https://github.com/openclaw-rocks/k8s-operator/commit/5bae852810d249358fad1453694503298851208c))
* Prevent endless Deployment reconciliation loop ([#29](https://github.com/openclaw-rocks/k8s-operator/issues/29)) ([db942b9](https://github.com/openclaw-rocks/k8s-operator/commit/db942b945586e6c3b49db7f0d3ca242f5fac7b44)), closes [#28](https://github.com/openclaw-rocks/k8s-operator/issues/28)
* update banner alt text ([#38](https://github.com/openclaw-rocks/k8s-operator/issues/38)) ([d00dc23](https://github.com/openclaw-rocks/k8s-operator/commit/d00dc23516d79e6a95b6aeca9d5d29c1a777e0e6))

## [Unreleased]

### Added
- Custom Prometheus metrics (reconciliation duration, instance phases, resource failures)
- ServiceMonitor resource creation for Prometheus Operator integration
- Defaulting webhook for setting sensible defaults
- Comprehensive resource builder unit tests
- Webhook validation unit tests
- `.golangci.yaml` linter configuration
- `.dockerignore` for optimized Docker builds
- Architecture documentation
- API reference documentation
- Troubleshooting guide
- Deployment guides for EKS, GKE, AKS
- Grafana dashboard example
- PrometheusRule alert examples

## [0.1.0] - 2024-01-01

### Added
- Initial release of OpenClaw Kubernetes Operator
- OpenClawInstance CRD (v1alpha1)
- Controller with full reconciliation lifecycle
- Security-first design (non-root, dropped capabilities, seccomp, NetworkPolicy)
- Validating webhook (blocks root, warns on insecure config)
- Managed resources: Deployment, Service, ServiceAccount, Role, RoleBinding, NetworkPolicy, PDB, ConfigMap, PVC, Ingress
- Chromium sidecar support for browser automation
- Helm chart for installation
- CI/CD with GitHub Actions (lint, test, security scan, multi-arch build)
- Container image signing with Cosign
- SBOM generation
- E2E test infrastructure
