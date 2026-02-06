# OpenClaw Kubernetes Operator

[![License](https://img.shields.io/badge/License-Apache%202.0-blue.svg)](https://opensource.org/licenses/Apache-2.0)
[![Go Report Card](https://goreportcard.com/badge/github.com/openclawrocks/k8s-operator)](https://goreportcard.com/report/github.com/openclawrocks/k8s-operator)
[![CI](https://github.com/OpenClaw-rocks/k8s-operator/actions/workflows/ci.yaml/badge.svg)](https://github.com/OpenClaw-rocks/k8s-operator/actions/workflows/ci.yaml)

A production-grade Kubernetes operator for deploying and managing OpenClaw AI assistant instances.

## Features

- **Declarative Deployment**: Deploy OpenClaw instances using Kubernetes Custom Resources
- **Security-First**: Built-in NetworkPolicies, RBAC, and Pod Security Standards
- **Provider-Agnostic**: Support any AI provider via environment variables
- **High Availability**: PodDisruptionBudgets, health probes, and graceful upgrades
- **Observability**: Prometheus metrics, structured logging, and Kubernetes events
- **Browser Automation**: Optional Chromium sidecar for web automation capabilities

## Quick Start

### Prerequisites

- Kubernetes 1.28+
- kubectl configured to access your cluster
- Helm 3 (for Helm installation)

### Installation

#### Using Helm

```bash
# Add the Helm repository (coming soon)
# helm repo add openclaw https://openclawrocks.github.io/k8s-operator

# Install the operator
helm install openclaw-operator ./charts/openclaw-operator \
  --namespace openclaw-operator-system \
  --create-namespace
```

#### Using Kustomize

```bash
# Install CRDs
make install

# Deploy the operator
make deploy IMG=ghcr.io/openclawrocks/openclaw-operator:latest
```

### Deploy an OpenClaw Instance

1. Create a secret with your API keys:

```yaml
apiVersion: v1
kind: Secret
metadata:
  name: openclaw-api-keys
type: Opaque
stringData:
  ANTHROPIC_API_KEY: "your-anthropic-api-key"
  OPENAI_API_KEY: "your-openai-api-key"
```

2. Create an OpenClawInstance:

```yaml
apiVersion: openclaw.openclaw.io/v1alpha1
kind: OpenClawInstance
metadata:
  name: my-openclaw
spec:
  envFrom:
    - secretRef:
        name: openclaw-api-keys
  storage:
    persistence:
      enabled: true
      size: 10Gi
```

3. Apply the resources:

```bash
kubectl apply -f secret.yaml
kubectl apply -f openclawinstance.yaml
```

## Configuration

### Minimal Example

```yaml
apiVersion: openclaw.openclaw.io/v1alpha1
kind: OpenClawInstance
metadata:
  name: my-openclaw
spec:
  envFrom:
    - secretRef:
        name: openclaw-api-keys
```

### Full Example

See [config/samples/openclaw_v1alpha1_openclawinstance_full.yaml](config/samples/openclaw_v1alpha1_openclawinstance_full.yaml) for a complete example with all options.

### Key Configuration Options

| Field | Description | Default |
|-------|-------------|---------|
| `spec.image.repository` | Container image repository | `ghcr.io/openclaw/openclaw` |
| `spec.image.tag` | Container image tag | `latest` |
| `spec.config.configMapRef` | Reference to external openclaw.json ConfigMap | - |
| `spec.config.raw` | Inline openclaw.json configuration | - |
| `spec.envFrom` | Environment variable sources (for API keys) | `[]` |
| `spec.resources.requests.cpu` | CPU request | `500m` |
| `spec.resources.requests.memory` | Memory request | `1Gi` |
| `spec.resources.limits.cpu` | CPU limit | `2000m` |
| `spec.resources.limits.memory` | Memory limit | `4Gi` |
| `spec.storage.persistence.enabled` | Enable persistent storage | `true` |
| `spec.storage.persistence.size` | PVC size | `10Gi` |
| `spec.chromium.enabled` | Enable Chromium sidecar | `false` |
| `spec.security.networkPolicy.enabled` | Enable NetworkPolicy | `true` |
| `spec.networking.ingress.enabled` | Enable Ingress | `false` |

## Security

The operator implements comprehensive security controls:

### Default Security Posture

- **Non-root execution**: All containers run as non-root (UID 1000)
- **Dropped capabilities**: All Linux capabilities are dropped
- **Seccomp profiles**: RuntimeDefault seccomp profile enabled
- **NetworkPolicies**: Default-deny with selective allowlisting
- **RBAC**: Minimal permissions per instance

### NetworkPolicy

By default, the operator creates a NetworkPolicy that:
- Allows ingress from the same namespace
- Allows DNS resolution (port 53)
- Allows HTTPS egress (port 443) for AI API access
- Blocks all other traffic

### Validating Webhook

The operator includes a validating webhook that:
- Blocks running as root (UID 0)
- Warns when NetworkPolicy is disabled
- Warns when Ingress lacks TLS
- Warns when Chromium lacks digest pinning

## Development

### Prerequisites

- Go 1.22+
- Docker
- kubectl
- Kind (for local testing)

### Building

```bash
# Generate code and manifests
make generate manifests

# Build the operator
make build

# Run tests
make test

# Build Docker image
make docker-build IMG=my-operator:dev
```

### Running Locally

```bash
# Install CRDs
make install

# Run the operator locally
make run
```

### Testing in Kind

```bash
# Create a Kind cluster
kind create cluster

# Install CRDs and deploy operator
make install
make deploy IMG=my-operator:dev

# Check operator logs
kubectl logs -n openclaw-operator-system deployment/openclaw-operator-controller-manager
```

## Managed Resources

For each OpenClawInstance, the operator creates and manages:

1. **Deployment** - Single-replica with security contexts
2. **Service** - ClusterIP (ports 18789, 18793)
3. **ServiceAccount** - Pod identity
4. **Role/RoleBinding** - Minimal RBAC permissions
5. **NetworkPolicy** - Network isolation (if enabled)
6. **PodDisruptionBudget** - Stability during maintenance
7. **ConfigMap** - openclaw.json (if using raw config)
8. **PersistentVolumeClaim** - Storage (if enabled)
9. **Ingress** - HTTP routing (if enabled)

## Status and Conditions

The operator updates the status of each OpenClawInstance:

```yaml
status:
  phase: Running  # Pending, Provisioning, Running, Degraded, Failed, Terminating
  conditions:
    - type: Ready
      status: "True"
    - type: DeploymentReady
      status: "True"
    - type: NetworkPolicyReady
      status: "True"
  gatewayEndpoint: my-openclaw.default.svc:18789
  canvasEndpoint: my-openclaw.default.svc:18793
```

## Contributing

Contributions are welcome! Please feel free to submit a Pull Request. For major changes, please open an issue first to discuss what you would like to change.

## License

This project is licensed under the Apache License 2.0 - see the [LICENSE](LICENSE) file for details.

This is the same license used by Kubernetes, Prometheus, and most CNCF projects, ensuring maximum compatibility and enterprise adoption.
