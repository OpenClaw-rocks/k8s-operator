# OpenClaw Kubernetes Operator

[![License](https://img.shields.io/badge/License-Apache%202.0-blue.svg)](https://opensource.org/licenses/Apache-2.0)
[![Go Report Card](https://goreportcard.com/badge/github.com/openclawrocks/k8s-operator)](https://goreportcard.com/report/github.com/openclawrocks/k8s-operator)
[![CI](https://github.com/OpenClaw-rocks/k8s-operator/actions/workflows/ci.yaml/badge.svg)](https://github.com/OpenClaw-rocks/k8s-operator/actions/workflows/ci.yaml)
[![Kubernetes](https://img.shields.io/badge/Kubernetes-1.28%2B-326CE5?logo=kubernetes&logoColor=white)](https://kubernetes.io)
[![Go](https://img.shields.io/badge/Go-1.24-00ADD8?logo=go&logoColor=white)](https://go.dev)

**Self-host [OpenClaw](https://openclaw.rocks) AI agents on Kubernetes with production-grade security, observability, and lifecycle management.**

OpenClaw is an AI agent platform that acts on your behalf across Telegram, Discord, WhatsApp, and Signal — managing your inbox, calendar, smart home, and more through 50+ integrations. While [openclaw.rocks](https://openclaw.rocks) offers fully managed hosting, this operator lets you run OpenClaw on your own infrastructure with the same operational rigor.

---

## Why an Operator?

Deploying AI agents to Kubernetes involves more than a Deployment and a Service. You need network isolation, secret management, persistent storage, health monitoring, optional browser automation, and config rollouts — all wired correctly. This operator encodes those concerns into a single `OpenClawInstance` custom resource so you can go from zero to production in minutes:

```yaml
apiVersion: openclaw.openclaw.io/v1alpha1
kind: OpenClawInstance
metadata:
  name: my-agent
spec:
  envFrom:
    - secretRef:
        name: openclaw-api-keys
  storage:
    persistence:
      enabled: true
      size: 10Gi
```

The operator reconciles this into a fully managed stack of 9+ Kubernetes resources — secured, monitored, and self-healing.

## Features

| | Feature | Details |
|---|---|---|
| **Declarative** | Single CRD | One resource defines the entire deployment — Deployment, Service, RBAC, NetworkPolicy, PVC, PDB, Ingress, and more |
| **Secure** | Hardened by default | Non-root (UID 1000), all capabilities dropped, seccomp RuntimeDefault, default-deny NetworkPolicy, validating webhook |
| **Observable** | Built-in metrics | Prometheus metrics, ServiceMonitor integration, structured JSON logging, Kubernetes events |
| **Flexible** | Provider-agnostic config | Use any AI provider — Anthropic, OpenAI, or others — via environment variables and inline or external config |
| **Resilient** | Self-healing lifecycle | PodDisruptionBudgets, health probes, automatic config rollouts via content hashing, 5-minute drift detection |
| **Extensible** | Chromium sidecar | Optional headless browser for web automation, injected as a hardened sidecar with shared-memory tuning |

## Architecture

```
┌──────────────────────────────────────────────────────────────┐
│  OpenClawInstance CR                                         │
│  (your declarative config)                                   │
└──────────────┬───────────────────────────────────────────────┘
               │ watch
               ▼
┌──────────────────────────────────────────────────────────────┐
│  OpenClaw Operator                                           │
│  ┌────────────┐  ┌──────────────┐  ┌──────────────────────┐ │
│  │ Reconciler │  │   Webhooks   │  │  Prometheus Metrics  │ │
│  │            │  │  (validate   │  │  (reconcile count,   │ │
│  │  creates → │  │   & default) │  │   duration, phases)  │ │
│  └────────────┘  └──────────────┘  └──────────────────────┘ │
└──────────────┬───────────────────────────────────────────────┘
               │ manages
               ▼
┌──────────────────────────────────────────────────────────────┐
│  Managed Resources (per instance)                            │
│                                                              │
│  ServiceAccount ─► Role ─► RoleBinding    NetworkPolicy      │
│  ConfigMap        PVC      PDB            ServiceMonitor     │
│                                                              │
│  ┌────────────────────────────────────────────────────────┐  │
│  │  Deployment                                            │  │
│  │  ┌──────────────────────┐  ┌────────────────────────┐ │  │
│  │  │  OpenClaw Container  │  │  Chromium Sidecar      │ │  │
│  │  │  (AI agent runtime)  │  │  (optional, port 9222) │ │  │
│  │  └──────────────────────┘  └────────────────────────┘ │  │
│  └────────────────────────────────────────────────────────┘  │
│                                                              │
│  Service (ports 18789, 18793) ─► Ingress (optional)          │
└──────────────────────────────────────────────────────────────┘
```

## Quick Start

### Prerequisites

- Kubernetes 1.28+
- Helm 3

### 1. Install the operator

```bash
helm install openclaw-operator \
  oci://ghcr.io/openclawrocks/charts/openclaw-operator \
  --namespace openclaw-operator-system \
  --create-namespace
```

<details>
<summary>Alternative: install with Kustomize</summary>

```bash
# Install CRDs
make install

# Deploy the operator
make deploy IMG=ghcr.io/openclawrocks/openclaw-operator:latest
```

</details>

### 2. Create a secret with your API keys

```yaml
apiVersion: v1
kind: Secret
metadata:
  name: openclaw-api-keys
type: Opaque
stringData:
  ANTHROPIC_API_KEY: "sk-ant-..."
```

### 3. Deploy an OpenClaw instance

```yaml
apiVersion: openclaw.openclaw.io/v1alpha1
kind: OpenClawInstance
metadata:
  name: my-agent
spec:
  envFrom:
    - secretRef:
        name: openclaw-api-keys
  storage:
    persistence:
      enabled: true
      size: 10Gi
```

```bash
kubectl apply -f secret.yaml -f openclawinstance.yaml
```

### 4. Verify

```bash
kubectl get openclawinstances
# NAME       PHASE     AGE
# my-agent   Running   2m

kubectl get pods
# NAME                        READY   STATUS    AGE
# my-agent-7f8b9c6d4-x2k9p   1/1     Running   2m
```

## Configuration

### Inline config (openclaw.json)

```yaml
spec:
  config:
    raw:
      agents:
        defaults:
          model:
            primary: "anthropic/claude-sonnet-4-20250514"
          sandbox: true
      session:
        scope: "per-sender"
```

### External ConfigMap reference

```yaml
spec:
  config:
    configMapRef:
      name: my-openclaw-config
      key: openclaw.json
```

Config changes are detected via SHA-256 hashing and automatically trigger a rolling update — no manual restart needed.

### Chromium sidecar

Enable headless browser automation for web scraping, screenshots, and browser-based integrations:

```yaml
spec:
  chromium:
    enabled: true
    image:
      repository: ghcr.io/browserless/chromium
      tag: "v2.0.0"
    resources:
      requests:
        cpu: "250m"
        memory: "512Mi"
      limits:
        cpu: "1000m"
        memory: "2Gi"
```

When enabled, the operator automatically injects a `CHROMIUM_URL` environment variable into the main container and configures shared memory, security contexts, and health probes for the sidecar.

### All configuration options

| Field | Description | Default |
|-------|-------------|---------|
| `spec.image.repository` | Container image | `ghcr.io/openclaw/openclaw` |
| `spec.image.tag` | Image tag | `latest` |
| `spec.config.raw` | Inline openclaw.json | - |
| `spec.config.configMapRef` | External ConfigMap reference | - |
| `spec.envFrom` | Secret/ConfigMap env sources | `[]` |
| `spec.resources.requests.cpu` | CPU request | `500m` |
| `spec.resources.requests.memory` | Memory request | `1Gi` |
| `spec.resources.limits.cpu` | CPU limit | `2000m` |
| `spec.resources.limits.memory` | Memory limit | `4Gi` |
| `spec.storage.persistence.enabled` | Persistent storage | `true` |
| `spec.storage.persistence.size` | PVC size | `10Gi` |
| `spec.storage.persistence.storageClass` | StorageClass name | cluster default |
| `spec.chromium.enabled` | Chromium sidecar | `false` |
| `spec.security.networkPolicy.enabled` | NetworkPolicy | `true` |
| `spec.security.networkPolicy.additionalEgress` | Custom egress rules | `[]` |
| `spec.networking.service.type` | Service type | `ClusterIP` |
| `spec.networking.ingress.enabled` | Ingress | `false` |
| `spec.observability.metrics.enabled` | Prometheus metrics | `true` |
| `spec.observability.metrics.serviceMonitor.enabled` | ServiceMonitor | `false` |
| `spec.availability.podDisruptionBudget.enabled` | PDB | `true` |

See the [full example](config/samples/openclaw_v1alpha1_openclawinstance_full.yaml) for every available field, or the [API reference](docs/api-reference.md) for detailed documentation.

## Security

The operator follows a **secure-by-default** philosophy. Every instance ships with hardened settings out of the box — no extra configuration needed.

### Defaults

- **Non-root execution** — containers run as UID 1000; root (UID 0) is blocked by the validating webhook
- **All capabilities dropped** — no ambient Linux capabilities
- **Seccomp RuntimeDefault** — syscall filtering enabled
- **Default-deny NetworkPolicy** — only DNS (53) and HTTPS (443) egress allowed; ingress limited to same namespace
- **Minimal RBAC** — each instance gets its own ServiceAccount with read-only access to its own ConfigMap
- **Read-only root filesystem** — supported for the Chromium sidecar; scratch dirs via emptyDir

### Validating webhook

| Check | Severity | Behavior |
|-------|----------|----------|
| `runAsUser: 0` | Error | Blocked — root execution not allowed |
| NetworkPolicy disabled | Warning | Deployment proceeds with a warning |
| Ingress without TLS | Warning | Deployment proceeds with a warning |
| Chromium without digest pinning | Warning | Deployment proceeds with a warning |

### Custom network rules

Allow egress to internal services or specific CIDRs:

```yaml
spec:
  security:
    networkPolicy:
      enabled: true
      additionalEgress:
        - ports:
            - protocol: TCP
              port: 5432
          to:
            - ipBlock:
                cidr: 10.0.0.0/8
```

## Observability

### Prometheus metrics

| Metric | Type | Description |
|--------|------|-------------|
| `openclaw_reconcile_total` | Counter | Reconciliations by result (success/error) |
| `openclaw_reconcile_duration_seconds` | Histogram | Reconciliation latency |
| `openclaw_instance_phase` | Gauge | Current phase per instance |

### ServiceMonitor

```yaml
spec:
  observability:
    metrics:
      enabled: true
      serviceMonitor:
        enabled: true
        interval: 15s
        labels:
          release: prometheus
```

### Instance status

```bash
kubectl get openclawinstance my-agent -o yaml
```

```yaml
status:
  phase: Running
  conditions:
    - type: Ready
      status: "True"
    - type: DeploymentReady
      status: "True"
    - type: NetworkPolicyReady
      status: "True"
  gatewayEndpoint: my-agent.default.svc:18789
  canvasEndpoint: my-agent.default.svc:18793
```

Phases: `Pending` → `Provisioning` → `Running` | `Degraded` | `Failed` | `Terminating`

## Deployment Guides

Platform-specific deployment guides are available for:

- [AWS EKS](docs/deployment.md#aws-eks)
- [Google GKE](docs/deployment.md#google-gke)
- [Azure AKS](docs/deployment.md#azure-aks)
- [Kind (local development)](docs/deployment.md#kind)

## Development

```bash
# Clone and set up
git clone https://github.com/OpenClaw-rocks/k8s-operator.git
cd k8s-operator
go mod download

# Generate code and manifests
make generate manifests

# Run tests
make test

# Run linter
make lint

# Run locally against a Kind cluster
kind create cluster
make install
make run
```

See [CONTRIBUTING.md](CONTRIBUTING.md) for the full development guide.

## Roadmap

- **v0.2.0** — Multi-instance scaling, HPA integration, backup & restore
- **v0.3.0** — Multi-cluster support, GitOps-native config, cert-manager integration
- **v1.0.0** — API graduation to v1, conformance test suite, CNCF Artifact Hub listing

See the full [roadmap](ROADMAP.md) for details.

## Don't Want to Self-Host?

[OpenClaw](https://openclaw.rocks) offers fully managed hosting starting at **EUR 10/mo** — no Kubernetes cluster required. Setup, updates, and 24/7 uptime handled for you.

## Contributing

Contributions are welcome. Please open an issue to discuss significant changes before submitting a PR. See [CONTRIBUTING.md](CONTRIBUTING.md) for guidelines.

## License

Apache License 2.0 — the same license used by Kubernetes, Prometheus, and most CNCF projects. See [LICENSE](LICENSE) for details.
