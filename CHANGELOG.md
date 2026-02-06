# Changelog

All notable changes to this project will be documented in this file.

The format is based on [Keep a Changelog](https://keepachangelog.com/en/1.1.0/),
and this project adheres to [Semantic Versioning](https://semver.org/spec/v2.0.0.html).

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
