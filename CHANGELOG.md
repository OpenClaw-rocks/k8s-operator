# Changelog

All notable changes to this project will be documented in this file.

The format is based on [Keep a Changelog](https://keepachangelog.com/en/1.1.0/),
and this project adheres to [Semantic Versioning](https://semver.org/spec/v2.0.0.html).

## [0.6.0](https://github.com/openclaw-rocks/k8s-operator/compare/v0.5.0...v0.6.0) (2026-02-13)


### ⚠ BREAKING CHANGES

* CRD API group changed from `openclaw.openclaw.io` to `openclaw.openclaw.rocks`. Existing CRDs must be deleted and re-created. This is acceptable at v1alpha1 stability level.

### Features

* Add nautical banner image with crab captain at Kubernetes helm ([84d1854](https://github.com/openclaw-rocks/k8s-operator/commit/84d1854cb7200587d6745c9ae3dc29c049fc3b8e))
* Add observability, testing, docs, and dev tooling ([5af27e1](https://github.com/openclaw-rocks/k8s-operator/commit/5af27e1884285d385356ee548f32a62dc8608a95))
* Add observability, testing, docs, and dev tooling for production readiness ([a8c063b](https://github.com/openclaw-rocks/k8s-operator/commit/a8c063b1e06143e06eacf84419ecbc31a9435684))
* Add support for custom sidecar containers ([#27](https://github.com/openclaw-rocks/k8s-operator/issues/27)) ([f0071f7](https://github.com/openclaw-rocks/k8s-operator/commit/f0071f72dd92b248223f53b7af9b2f3eb9a184be)), closes [#24](https://github.com/openclaw-rocks/k8s-operator/issues/24)
* Initial OpenClaw Kubernetes Operator implementation ([6d873ff](https://github.com/openclaw-rocks/k8s-operator/commit/6d873ff855273c9e51337181e657ced2b82af711))
* Inject CHROMIUM_URL env var into main container when sidecar is enabled ([8553cc7](https://github.com/openclaw-rocks/k8s-operator/commit/8553cc7ec0b41675bd815a0f8febd0622bbf068c))
* Inject CHROMIUM_URL env var when sidecar is enabled ([b80dabe](https://github.com/openclaw-rocks/k8s-operator/commit/b80dabe32dbf2ff0f755c92f4e83e2467050bd89))
* Replace manual release steps with GoReleaser ([#2](https://github.com/openclaw-rocks/k8s-operator/issues/2)) ([fedc497](https://github.com/openclaw-rocks/k8s-operator/commit/fedc49766d414e0a2d96bc9c080abcb82b2b2202))
* Support custom egress rules in NetworkPolicy ([#15](https://github.com/openclaw-rocks/k8s-operator/issues/15)) ([#16](https://github.com/openclaw-rocks/k8s-operator/issues/16)) ([c62dc09](https://github.com/openclaw-rocks/k8s-operator/commit/c62dc09cec6112cf4625175423d033310ce50dc3))
* update banner with real OpenClaw logo and Kubernetes logo ([#37](https://github.com/openclaw-rocks/k8s-operator/issues/37)) ([840adb9](https://github.com/openclaw-rocks/k8s-operator/commit/840adb9e285c14dffe1bfed17e9b9411b05f4538))


### Bug Fixes

* Add leader election RBAC and E2E test infrastructure ([4eb6019](https://github.com/openclaw-rocks/k8s-operator/commit/4eb601994f3b87bde07f4c28a51f1dd010f045d4))
* Apply same CreateOrUpdate pattern to ServiceMonitor reconciler ([#30](https://github.com/openclaw-rocks/k8s-operator/issues/30)) ([c1aaa36](https://github.com/openclaw-rocks/k8s-operator/commit/c1aaa3639b7cb82eb95696bbe06b98817d46a55e))
* Bump chromium /dev/shm from 256Mi to 1Gi ([925d7b8](https://github.com/openclaw-rocks/k8s-operator/commit/925d7b84c0630b73f581fca32d8da68ca582cd19))
* Bump chromium /dev/shm sizeLimit from 256Mi to 1Gi ([a4b3fbb](https://github.com/openclaw-rocks/k8s-operator/commit/a4b3fbb005fd40a743a54dcb8ce42c0db9ef7b27))
* change CRD API group domain from openclaw.io to openclaw.rocks ([#41](https://github.com/openclaw-rocks/k8s-operator/issues/41)) ([5bae852](https://github.com/openclaw-rocks/k8s-operator/commit/5bae852810d249358fad1453694503298851208c))
* Chromium sidecar crash (UID mismatch + read-only rootfs) and lint ([#14](https://github.com/openclaw-rocks/k8s-operator/issues/14)) ([febe1d3](https://github.com/openclaw-rocks/k8s-operator/commit/febe1d323bb80e4ffc0e36721fa1cd75a949b023))
* Disable SBOM upload-release-assets to avoid race condition ([d909808](https://github.com/openclaw-rocks/k8s-operator/commit/d909808d5aecde367748691604be1fbc152a6619))
* Downgrade to Go 1.23 for golangci-lint compatibility ([dd33c08](https://github.com/openclaw-rocks/k8s-operator/commit/dd33c08d1ffe962f358c14a6ae0a7b3105799d44))
* Increase golangci-lint timeout and update to v1.63.4 ([1d982ca](https://github.com/openclaw-rocks/k8s-operator/commit/1d982ca114f3316ad656bf873ad0fc6694f71cc5))
* Link OpenClaw to openclaw.ai, not openclaw.rocks ([679b043](https://github.com/openclaw-rocks/k8s-operator/commit/679b04340afc6835284f2e9e13e7aef02e69ed14))
* Link OpenClaw to openclaw.ai, not openclaw.rocks ([#22](https://github.com/openclaw-rocks/k8s-operator/issues/22)) ([9bfdbd8](https://github.com/openclaw-rocks/k8s-operator/commit/9bfdbd86b29ed43f043a16bdbd126f6960df32a2))
* Polish README copy and diagram alignment ([#19](https://github.com/openclaw-rocks/k8s-operator/issues/19)) ([ef6a8d1](https://github.com/openclaw-rocks/k8s-operator/commit/ef6a8d1f0973f36d3c790d34529f86e1c1d7343c))
* Pre-enable channel modules in config to prevent EBUSY on startup ([#13](https://github.com/openclaw-rocks/k8s-operator/issues/13)) ([21ee585](https://github.com/openclaw-rocks/k8s-operator/commit/21ee585fb4bef4f77afed0b0a8d730c0975af0e3)), closes [#11](https://github.com/openclaw-rocks/k8s-operator/issues/11)
* Prevent endless Deployment reconciliation loop ([#29](https://github.com/openclaw-rocks/k8s-operator/issues/29)) ([db942b9](https://github.com/openclaw-rocks/k8s-operator/commit/db942b945586e6c3b49db7f0d3ca242f5fac7b44)), closes [#28](https://github.com/openclaw-rocks/k8s-operator/issues/28)
* Remove chart-releaser, keep OCI-only Helm distribution ([#3](https://github.com/openclaw-rocks/k8s-operator/issues/3)) ([99e3cdb](https://github.com/openclaw-rocks/k8s-operator/commit/99e3cdba598ac46e5cd3cab2dc7dad00b456a0e8))
* Replace config subPath mount with init container to avoid EBUSY ([#10](https://github.com/openclaw-rocks/k8s-operator/issues/10)) ([38b60d3](https://github.com/openclaw-rocks/k8s-operator/commit/38b60d39b835fb1eca459f6c4d612d0304873c3e)), closes [#9](https://github.com/openclaw-rocks/k8s-operator/issues/9)
* Resolve variable shadowing of err in PVC reconciliation ([a9ea85b](https://github.com/openclaw-rocks/k8s-operator/commit/a9ea85b05770ba4123b4c0325e551ba93e90d103))
* Set HOME env var to match config mount path ([#5](https://github.com/openclaw-rocks/k8s-operator/issues/5)) ([175ff92](https://github.com/openclaw-rocks/k8s-operator/commit/175ff9219b5af8d80e37579627f8ebccce71cb1b)), closes [#4](https://github.com/openclaw-rocks/k8s-operator/issues/4)
* skip GitHub release creation in release-please ([#47](https://github.com/openclaw-rocks/k8s-operator/issues/47)) ([96a5be9](https://github.com/openclaw-rocks/k8s-operator/commit/96a5be93020ef61d6a37f804e1a741b3a229e86d))
* Specify kind cluster name for image loading ([7ac2116](https://github.com/openclaw-rocks/k8s-operator/commit/7ac21166b36e79fb0d93c23eb24de8c420dd8446))
* update banner alt text ([#38](https://github.com/openclaw-rocks/k8s-operator/issues/38)) ([d00dc23](https://github.com/openclaw-rocks/k8s-operator/commit/d00dc23516d79e6a95b6aeca9d5d29c1a777e0e6))
* Update Chart.yaml version/appVersion to match latest release v0.2.4 ([#26](https://github.com/openclaw-rocks/k8s-operator/issues/26)) ([c475ffc](https://github.com/openclaw-rocks/k8s-operator/commit/c475ffcd145b5109befb784a1a92672377be8936))
* Update copyright to 2026 OpenClaw.rocks ([bca1f0f](https://github.com/openclaw-rocks/k8s-operator/commit/bca1f0f626ebd8c7c1285882d20b87483380fba4))
* Update copyright to 2026 OpenClaw.rocks ([4d462f9](https://github.com/openclaw-rocks/k8s-operator/commit/4d462f97799865f48c13525074702f4b1163e54f))
* Update Go version to 1.24 for CI compatibility ([5c13d06](https://github.com/openclaw-rocks/k8s-operator/commit/5c13d068f3cba6d19609a3ea78b237d4b0b311ff))
* Use correct GitHub org name (OpenClaw-rocks) in all references ([#20](https://github.com/openclaw-rocks/k8s-operator/issues/20)) ([c157899](https://github.com/openclaw-rocks/k8s-operator/commit/c1578998cb7bfb4824eae64ce390fba4e508e6ca))
* Use direct append instead of loop for image pull secrets ([187477c](https://github.com/openclaw-rocks/k8s-operator/commit/187477c3e205ddf332bd9f2831936d141b1e4088))
* Use Go 1.24 with goinstall mode for golangci-lint ([3dd40c8](https://github.com/openclaw-rocks/k8s-operator/commit/3dd40c8fc02e964c9edd98407ae44a7d94634b2b))
* Use govet enable list for shadow analyzer ([3f04b1b](https://github.com/openclaw-rocks/k8s-operator/commit/3f04b1bd705f33df5c438e3ae9b22fd5e2cd7975))
* Use lowercase image names for OCI registry compatibility ([2d3c87e](https://github.com/openclaw-rocks/k8s-operator/commit/2d3c87ea562800a98b5be6c63b71bafe1fbae1e8))
* Use lowercase owner name for Helm OCI registry ([67bc33e](https://github.com/openclaw-rocks/k8s-operator/commit/67bc33e503a6f1329908dd437731b3bf96be2ea5))
* use PAT for release-please to trigger downstream workflows ([#45](https://github.com/openclaw-rocks/k8s-operator/issues/45)) ([21191ed](https://github.com/openclaw-rocks/k8s-operator/commit/21191ed3e29ff2d4aeb4c0aa269eca1fb7f6284a))

## [0.5.0](https://github.com/openclaw-rocks/k8s-operator/compare/v0.4.0...v0.5.0) (2026-02-13)


### ⚠ BREAKING CHANGES

* CRD API group changed from `openclaw.openclaw.io` to `openclaw.openclaw.rocks`. Existing CRDs must be deleted and re-created. This is acceptable at v1alpha1 stability level.

### Features

* Add nautical banner image with crab captain at Kubernetes helm ([84d1854](https://github.com/openclaw-rocks/k8s-operator/commit/84d1854cb7200587d6745c9ae3dc29c049fc3b8e))
* Add observability, testing, docs, and dev tooling ([5af27e1](https://github.com/openclaw-rocks/k8s-operator/commit/5af27e1884285d385356ee548f32a62dc8608a95))
* Add observability, testing, docs, and dev tooling for production readiness ([a8c063b](https://github.com/openclaw-rocks/k8s-operator/commit/a8c063b1e06143e06eacf84419ecbc31a9435684))
* Add support for custom sidecar containers ([#27](https://github.com/openclaw-rocks/k8s-operator/issues/27)) ([f0071f7](https://github.com/openclaw-rocks/k8s-operator/commit/f0071f72dd92b248223f53b7af9b2f3eb9a184be)), closes [#24](https://github.com/openclaw-rocks/k8s-operator/issues/24)
* Initial OpenClaw Kubernetes Operator implementation ([6d873ff](https://github.com/openclaw-rocks/k8s-operator/commit/6d873ff855273c9e51337181e657ced2b82af711))
* Inject CHROMIUM_URL env var into main container when sidecar is enabled ([8553cc7](https://github.com/openclaw-rocks/k8s-operator/commit/8553cc7ec0b41675bd815a0f8febd0622bbf068c))
* Inject CHROMIUM_URL env var when sidecar is enabled ([b80dabe](https://github.com/openclaw-rocks/k8s-operator/commit/b80dabe32dbf2ff0f755c92f4e83e2467050bd89))
* Replace manual release steps with GoReleaser ([#2](https://github.com/openclaw-rocks/k8s-operator/issues/2)) ([fedc497](https://github.com/openclaw-rocks/k8s-operator/commit/fedc49766d414e0a2d96bc9c080abcb82b2b2202))
* Support custom egress rules in NetworkPolicy ([#15](https://github.com/openclaw-rocks/k8s-operator/issues/15)) ([#16](https://github.com/openclaw-rocks/k8s-operator/issues/16)) ([c62dc09](https://github.com/openclaw-rocks/k8s-operator/commit/c62dc09cec6112cf4625175423d033310ce50dc3))
* update banner with real OpenClaw logo and Kubernetes logo ([#37](https://github.com/openclaw-rocks/k8s-operator/issues/37)) ([840adb9](https://github.com/openclaw-rocks/k8s-operator/commit/840adb9e285c14dffe1bfed17e9b9411b05f4538))


### Bug Fixes

* Add leader election RBAC and E2E test infrastructure ([4eb6019](https://github.com/openclaw-rocks/k8s-operator/commit/4eb601994f3b87bde07f4c28a51f1dd010f045d4))
* Apply same CreateOrUpdate pattern to ServiceMonitor reconciler ([#30](https://github.com/openclaw-rocks/k8s-operator/issues/30)) ([c1aaa36](https://github.com/openclaw-rocks/k8s-operator/commit/c1aaa3639b7cb82eb95696bbe06b98817d46a55e))
* Bump chromium /dev/shm from 256Mi to 1Gi ([925d7b8](https://github.com/openclaw-rocks/k8s-operator/commit/925d7b84c0630b73f581fca32d8da68ca582cd19))
* Bump chromium /dev/shm sizeLimit from 256Mi to 1Gi ([a4b3fbb](https://github.com/openclaw-rocks/k8s-operator/commit/a4b3fbb005fd40a743a54dcb8ce42c0db9ef7b27))
* change CRD API group domain from openclaw.io to openclaw.rocks ([#41](https://github.com/openclaw-rocks/k8s-operator/issues/41)) ([5bae852](https://github.com/openclaw-rocks/k8s-operator/commit/5bae852810d249358fad1453694503298851208c))
* Chromium sidecar crash (UID mismatch + read-only rootfs) and lint ([#14](https://github.com/openclaw-rocks/k8s-operator/issues/14)) ([febe1d3](https://github.com/openclaw-rocks/k8s-operator/commit/febe1d323bb80e4ffc0e36721fa1cd75a949b023))
* Disable SBOM upload-release-assets to avoid race condition ([d909808](https://github.com/openclaw-rocks/k8s-operator/commit/d909808d5aecde367748691604be1fbc152a6619))
* Downgrade to Go 1.23 for golangci-lint compatibility ([dd33c08](https://github.com/openclaw-rocks/k8s-operator/commit/dd33c08d1ffe962f358c14a6ae0a7b3105799d44))
* Increase golangci-lint timeout and update to v1.63.4 ([1d982ca](https://github.com/openclaw-rocks/k8s-operator/commit/1d982ca114f3316ad656bf873ad0fc6694f71cc5))
* Link OpenClaw to openclaw.ai, not openclaw.rocks ([679b043](https://github.com/openclaw-rocks/k8s-operator/commit/679b04340afc6835284f2e9e13e7aef02e69ed14))
* Link OpenClaw to openclaw.ai, not openclaw.rocks ([#22](https://github.com/openclaw-rocks/k8s-operator/issues/22)) ([9bfdbd8](https://github.com/openclaw-rocks/k8s-operator/commit/9bfdbd86b29ed43f043a16bdbd126f6960df32a2))
* Polish README copy and diagram alignment ([#19](https://github.com/openclaw-rocks/k8s-operator/issues/19)) ([ef6a8d1](https://github.com/openclaw-rocks/k8s-operator/commit/ef6a8d1f0973f36d3c790d34529f86e1c1d7343c))
* Pre-enable channel modules in config to prevent EBUSY on startup ([#13](https://github.com/openclaw-rocks/k8s-operator/issues/13)) ([21ee585](https://github.com/openclaw-rocks/k8s-operator/commit/21ee585fb4bef4f77afed0b0a8d730c0975af0e3)), closes [#11](https://github.com/openclaw-rocks/k8s-operator/issues/11)
* Prevent endless Deployment reconciliation loop ([#29](https://github.com/openclaw-rocks/k8s-operator/issues/29)) ([db942b9](https://github.com/openclaw-rocks/k8s-operator/commit/db942b945586e6c3b49db7f0d3ca242f5fac7b44)), closes [#28](https://github.com/openclaw-rocks/k8s-operator/issues/28)
* Remove chart-releaser, keep OCI-only Helm distribution ([#3](https://github.com/openclaw-rocks/k8s-operator/issues/3)) ([99e3cdb](https://github.com/openclaw-rocks/k8s-operator/commit/99e3cdba598ac46e5cd3cab2dc7dad00b456a0e8))
* Replace config subPath mount with init container to avoid EBUSY ([#10](https://github.com/openclaw-rocks/k8s-operator/issues/10)) ([38b60d3](https://github.com/openclaw-rocks/k8s-operator/commit/38b60d39b835fb1eca459f6c4d612d0304873c3e)), closes [#9](https://github.com/openclaw-rocks/k8s-operator/issues/9)
* Resolve variable shadowing of err in PVC reconciliation ([a9ea85b](https://github.com/openclaw-rocks/k8s-operator/commit/a9ea85b05770ba4123b4c0325e551ba93e90d103))
* Set HOME env var to match config mount path ([#5](https://github.com/openclaw-rocks/k8s-operator/issues/5)) ([175ff92](https://github.com/openclaw-rocks/k8s-operator/commit/175ff9219b5af8d80e37579627f8ebccce71cb1b)), closes [#4](https://github.com/openclaw-rocks/k8s-operator/issues/4)
* Specify kind cluster name for image loading ([7ac2116](https://github.com/openclaw-rocks/k8s-operator/commit/7ac21166b36e79fb0d93c23eb24de8c420dd8446))
* update banner alt text ([#38](https://github.com/openclaw-rocks/k8s-operator/issues/38)) ([d00dc23](https://github.com/openclaw-rocks/k8s-operator/commit/d00dc23516d79e6a95b6aeca9d5d29c1a777e0e6))
* Update Chart.yaml version/appVersion to match latest release v0.2.4 ([#26](https://github.com/openclaw-rocks/k8s-operator/issues/26)) ([c475ffc](https://github.com/openclaw-rocks/k8s-operator/commit/c475ffcd145b5109befb784a1a92672377be8936))
* Update copyright to 2026 OpenClaw.rocks ([bca1f0f](https://github.com/openclaw-rocks/k8s-operator/commit/bca1f0f626ebd8c7c1285882d20b87483380fba4))
* Update copyright to 2026 OpenClaw.rocks ([4d462f9](https://github.com/openclaw-rocks/k8s-operator/commit/4d462f97799865f48c13525074702f4b1163e54f))
* Update Go version to 1.24 for CI compatibility ([5c13d06](https://github.com/openclaw-rocks/k8s-operator/commit/5c13d068f3cba6d19609a3ea78b237d4b0b311ff))
* Use correct GitHub org name (OpenClaw-rocks) in all references ([#20](https://github.com/openclaw-rocks/k8s-operator/issues/20)) ([c157899](https://github.com/openclaw-rocks/k8s-operator/commit/c1578998cb7bfb4824eae64ce390fba4e508e6ca))
* Use direct append instead of loop for image pull secrets ([187477c](https://github.com/openclaw-rocks/k8s-operator/commit/187477c3e205ddf332bd9f2831936d141b1e4088))
* Use Go 1.24 with goinstall mode for golangci-lint ([3dd40c8](https://github.com/openclaw-rocks/k8s-operator/commit/3dd40c8fc02e964c9edd98407ae44a7d94634b2b))
* Use govet enable list for shadow analyzer ([3f04b1b](https://github.com/openclaw-rocks/k8s-operator/commit/3f04b1bd705f33df5c438e3ae9b22fd5e2cd7975))
* Use lowercase image names for OCI registry compatibility ([2d3c87e](https://github.com/openclaw-rocks/k8s-operator/commit/2d3c87ea562800a98b5be6c63b71bafe1fbae1e8))
* Use lowercase owner name for Helm OCI registry ([67bc33e](https://github.com/openclaw-rocks/k8s-operator/commit/67bc33e503a6f1329908dd437731b3bf96be2ea5))
* use PAT for release-please to trigger downstream workflows ([#45](https://github.com/openclaw-rocks/k8s-operator/issues/45)) ([21191ed](https://github.com/openclaw-rocks/k8s-operator/commit/21191ed3e29ff2d4aeb4c0aa269eca1fb7f6284a))

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
