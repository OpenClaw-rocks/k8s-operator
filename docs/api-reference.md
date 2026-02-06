# API Reference

## OpenClawInstance (v1alpha1)

**Group**: `openclaw.openclaw.io`
**Version**: `v1alpha1`
**Kind**: `OpenClawInstance`
**Scope**: Namespaced

An `OpenClawInstance` represents a single deployment of the OpenClaw AI assistant in a Kubernetes cluster. The operator watches these resources and reconciles a full stack of dependent objects (Deployment, Service, RBAC, NetworkPolicy, storage, and more).

### Print Columns

When listing resources with `kubectl get openclawinstances`, the following columns are displayed:

| Column    | JSON Path                                          |
|-----------|----------------------------------------------------|
| Phase     | `.status.phase`                                    |
| Ready     | `.status.conditions[?(@.type=='Ready')].status`    |
| Gateway   | `.status.gatewayEndpoint`                          |
| Age       | `.metadata.creationTimestamp`                      |

---

## Spec Fields

### spec.image

Container image configuration for the main OpenClaw workload.

| Field          | Type                         | Default                        | Description                                                       |
|----------------|------------------------------|--------------------------------|-------------------------------------------------------------------|
| `repository`   | `string`                     | `ghcr.io/openclaw/openclaw`    | Container image repository.                                       |
| `tag`          | `string`                     | `latest`                       | Container image tag.                                              |
| `digest`       | `string`                     | --                             | Image digest (overrides `tag` if set). Format: `sha256:abc...`.   |
| `pullPolicy`   | `string`                     | `IfNotPresent`                 | Image pull policy. One of: `Always`, `IfNotPresent`, `Never`.     |
| `pullSecrets`  | `[]LocalObjectReference`     | --                             | List of Secrets for pulling from private registries.              |

### spec.config

Configuration for the OpenClaw application (`openclaw.json`).

| Field          | Type                  | Default | Description                                                                |
|----------------|-----------------------|---------|----------------------------------------------------------------------------|
| `configMapRef` | `ConfigMapKeySelector`| --      | Reference to an external ConfigMap. If set, `raw` is ignored.              |
| `raw`          | `RawConfig`           | --      | Inline JSON configuration. The operator creates a managed ConfigMap.       |

**ConfigMapKeySelector:**

| Field  | Type     | Default          | Description                            |
|--------|----------|------------------|----------------------------------------|
| `name` | `string` | (required)       | Name of the ConfigMap.                 |
| `key`  | `string` | `openclaw.json`  | Key within the ConfigMap to mount.     |

### spec.envFrom

| Field     | Type                  | Default | Description                                                                       |
|-----------|-----------------------|---------|-----------------------------------------------------------------------------------|
| `envFrom` | `[]EnvFromSource`     | --      | Sources to populate environment variables (e.g., Secrets for `ANTHROPIC_API_KEY`). |

Standard Kubernetes `EnvFromSource`. Commonly used to inject API keys from a Secret:

```yaml
spec:
  envFrom:
    - secretRef:
        name: openclaw-api-keys
```

### spec.env

| Field | Type           | Default | Description                                         |
|-------|----------------|---------|-----------------------------------------------------|
| `env` | `[]EnvVar`     | --      | Individual environment variables to set.             |

Standard Kubernetes `EnvVar`. Example:

```yaml
spec:
  env:
    - name: LOG_LEVEL
      value: "debug"
```

### spec.resources

Compute resource requirements for the main OpenClaw container.

| Field                | Type     | Default  | Description                          |
|----------------------|----------|----------|--------------------------------------|
| `requests.cpu`       | `string` | `500m`   | Minimum CPU (e.g., `500m`, `2`).     |
| `requests.memory`    | `string` | `1Gi`    | Minimum memory (e.g., `512Mi`).      |
| `limits.cpu`         | `string` | `2000m`  | Maximum CPU.                         |
| `limits.memory`      | `string` | `4Gi`    | Maximum memory.                      |

### spec.security

Security-related configuration for the instance.

#### spec.security.podSecurityContext

| Field          | Type     | Default | Description                                             |
|----------------|----------|---------|---------------------------------------------------------|
| `runAsUser`    | `*int64` | `1000`  | UID to run as. Setting to `0` is rejected by webhook.   |
| `runAsGroup`   | `*int64` | `1000`  | GID to run as.                                          |
| `fsGroup`      | `*int64` | `1000`  | Supplemental group for volume ownership.                |
| `runAsNonRoot` | `*bool`  | `true`  | Require non-root execution. Warns if set to `false`.    |

#### spec.security.containerSecurityContext

| Field                      | Type              | Default | Description                                                    |
|----------------------------|-------------------|---------|----------------------------------------------------------------|
| `allowPrivilegeEscalation` | `*bool`           | `false` | Allow privilege escalation. Warns if set to `true`.            |
| `readOnlyRootFilesystem`   | `*bool`           | `false` | Mount root filesystem as read-only. OpenClaw writes to `~/.openclaw/`. |
| `capabilities`             | `*Capabilities`   | Drop ALL | Linux capabilities to add or drop.                            |

#### spec.security.networkPolicy

| Field                      | Type       | Default | Description                                                  |
|----------------------------|------------|---------|--------------------------------------------------------------|
| `enabled`                  | `*bool`    | `true`  | Create a NetworkPolicy. Warns if disabled.                   |
| `allowedIngressCIDRs`      | `[]string` | --      | CIDRs allowed to reach the instance.                         |
| `allowedIngressNamespaces` | `[]string` | --      | Namespaces allowed to reach the instance.                    |
| `allowedEgressCIDRs`       | `[]string` | --      | CIDRs the instance can reach (in addition to HTTPS/DNS).     |
| `allowDNS`                 | `*bool`    | `true`  | Allow DNS resolution (UDP/TCP port 53).                      |

#### spec.security.rbac

| Field                  | Type         | Default | Description                                                  |
|------------------------|--------------|---------|--------------------------------------------------------------|
| `createServiceAccount` | `*bool`      | `true`  | Create a dedicated ServiceAccount for this instance.         |
| `serviceAccountName`   | `string`     | --      | Use an existing ServiceAccount (only when `createServiceAccount` is `false`). |
| `additionalRules`      | `[]RBACRule` | --      | Custom RBAC rules appended to the generated Role.            |

**RBACRule:**

| Field       | Type       | Description                                    |
|-------------|------------|------------------------------------------------|
| `apiGroups` | `[]string` | API groups (e.g., `[""]` for core, `["apps"]`).|
| `resources` | `[]string` | Resources (e.g., `["pods"]`).                  |
| `verbs`     | `[]string` | Verbs (e.g., `["get", "list"]`).               |

### spec.storage

Persistent storage configuration.

#### spec.storage.persistence

| Field           | Type                            | Default            | Description                                          |
|-----------------|---------------------------------|--------------------|------------------------------------------------------|
| `enabled`       | `*bool`                         | `true`             | Enable persistent storage via PVC.                   |
| `storageClass`  | `*string`                       | (cluster default)  | StorageClass name. Immutable after creation.         |
| `size`          | `string`                        | `10Gi`             | PVC size.                                            |
| `accessModes`   | `[]PersistentVolumeAccessMode`  | `[ReadWriteOnce]`  | PVC access modes.                                    |
| `existingClaim` | `string`                        | --                 | Name of an existing PVC to use instead of creating one. |

### spec.chromium

Optional Chromium sidecar for browser automation.

| Field                   | Type     | Default                          | Description                                         |
|-------------------------|----------|----------------------------------|-----------------------------------------------------|
| `enabled`               | `bool`   | `false`                          | Enable the Chromium sidecar container.              |
| `image.repository`      | `string` | `ghcr.io/browserless/chromium`   | Chromium container image repository.                |
| `image.tag`             | `string` | `latest`                         | Chromium image tag.                                 |
| `image.digest`          | `string` | --                               | Chromium image digest for supply chain security.    |
| `resources.requests.cpu`   | `string` | `250m`                        | Chromium minimum CPU.                               |
| `resources.requests.memory`| `string` | `512Mi`                       | Chromium minimum memory.                            |
| `resources.limits.cpu`     | `string` | `1000m`                       | Chromium maximum CPU.                               |
| `resources.limits.memory`  | `string` | `2Gi`                         | Chromium maximum memory.                            |

When enabled, the sidecar:

- Exposes Chrome DevTools Protocol on port 9222.
- Runs as UID 1001 with a read-only root filesystem.
- Mounts a memory-backed emptyDir at `/dev/shm` (256Mi) for shared memory.
- Mounts an emptyDir at `/tmp` for scratch space.

### spec.networking

Network-related configuration for the instance.

#### spec.networking.service

| Field         | Type                | Default      | Description                                               |
|---------------|---------------------|--------------|-----------------------------------------------------------|
| `type`        | `string`            | `ClusterIP`  | Service type. One of: `ClusterIP`, `LoadBalancer`, `NodePort`. |
| `annotations` | `map[string]string` | --           | Annotations to add to the Service.                        |

The Service always exposes:

| Port Name   | Port   | Description                     |
|-------------|--------|---------------------------------|
| `gateway`   | 18789  | OpenClaw WebSocket gateway.     |
| `canvas`    | 18793  | OpenClaw Canvas HTTP server.    |
| `chromium`  | 9222   | Chrome DevTools Protocol (only if Chromium sidecar is enabled). |

#### spec.networking.ingress

| Field         | Type                | Default | Description                                         |
|---------------|---------------------|---------|-----------------------------------------------------|
| `enabled`     | `bool`              | `false` | Create an Ingress resource.                         |
| `className`   | `*string`           | --      | IngressClass to use (e.g., `nginx`, `traefik`).     |
| `annotations` | `map[string]string` | --      | Custom annotations added to the Ingress.            |
| `hosts`       | `[]IngressHost`     | --      | List of hosts to route traffic for.                 |
| `tls`         | `[]IngressTLS`      | --      | TLS termination configuration. Warns if empty.      |
| `security`    | `IngressSecuritySpec`| --     | Ingress security settings (HTTPS redirect, HSTS, rate limiting). |

**IngressHost:**

| Field   | Type            | Description                                 |
|---------|-----------------|---------------------------------------------|
| `host`  | `string`        | Fully qualified domain name.                |
| `paths` | `[]IngressPath` | Paths to route. Defaults to `[{path: "/"}]`.|

**IngressPath:**

| Field      | Type     | Default    | Description                                                              |
|------------|----------|------------|--------------------------------------------------------------------------|
| `path`     | `string` | `/`        | URL path.                                                                |
| `pathType` | `string` | `Prefix`   | Path matching. One of: `Prefix`, `Exact`, `ImplementationSpecific`.      |

**IngressTLS:**

| Field        | Type       | Description                                         |
|--------------|------------|-----------------------------------------------------|
| `hosts`      | `[]string` | Hostnames covered by the TLS certificate.           |
| `secretName` | `string`   | Secret containing the TLS key pair.                 |

**IngressSecuritySpec:**

| Field                       | Type     | Default | Description                                    |
|-----------------------------|----------|---------|------------------------------------------------|
| `forceHTTPS`                | `*bool`  | `true`  | Redirect HTTP to HTTPS.                        |
| `enableHSTS`                | `*bool`  | `true`  | Add HSTS headers.                              |
| `rateLimiting.enabled`      | `*bool`  | `true`  | Enable rate limiting.                          |
| `rateLimiting.requestsPerSecond` | `*int32` | `10` | Maximum requests per second.              |

The operator automatically adds WebSocket-related annotations for nginx-ingress (proxy timeouts, HTTP/1.1 upgrade).

### spec.probes

Health probe configuration for the main OpenClaw container. All probes use TCP socket checks against the gateway port (18789).

#### spec.probes.liveness

| Field                 | Type     | Default | Description                                           |
|-----------------------|----------|---------|-------------------------------------------------------|
| `enabled`             | `*bool`  | `true`  | Enable the liveness probe.                            |
| `initialDelaySeconds` | `*int32` | `30`    | Seconds to wait before the first check.               |
| `periodSeconds`       | `*int32` | `10`    | Seconds between checks.                              |
| `timeoutSeconds`      | `*int32` | `5`     | Seconds before the check times out.                  |
| `failureThreshold`    | `*int32` | `3`     | Consecutive failures before restarting the container. |

#### spec.probes.readiness

| Field                 | Type     | Default | Description                                           |
|-----------------------|----------|---------|-------------------------------------------------------|
| `enabled`             | `*bool`  | `true`  | Enable the readiness probe.                           |
| `initialDelaySeconds` | `*int32` | `5`     | Seconds to wait before the first check.               |
| `periodSeconds`       | `*int32` | `5`     | Seconds between checks.                              |
| `timeoutSeconds`      | `*int32` | `3`     | Seconds before the check times out.                  |
| `failureThreshold`    | `*int32` | `3`     | Consecutive failures before removing from endpoints. |

#### spec.probes.startup

| Field                 | Type     | Default | Description                                           |
|-----------------------|----------|---------|-------------------------------------------------------|
| `enabled`             | `*bool`  | `true`  | Enable the startup probe.                             |
| `initialDelaySeconds` | `*int32` | `0`     | Seconds to wait before the first check.               |
| `periodSeconds`       | `*int32` | `5`     | Seconds between checks.                              |
| `timeoutSeconds`      | `*int32` | `3`     | Seconds before the check times out.                  |
| `failureThreshold`    | `*int32` | `30`    | Consecutive failures before killing the container. Allows up to 150s startup. |

### spec.observability

Metrics and logging configuration.

#### spec.observability.metrics

| Field                       | Type                | Default | Description                                   |
|-----------------------------|---------------------|---------|-----------------------------------------------|
| `enabled`                   | `*bool`             | `true`  | Enable the metrics endpoint on the managed instance. |
| `port`                      | `*int32`            | `9090`  | Metrics port.                                 |
| `serviceMonitor.enabled`    | `*bool`             | `false` | Create a Prometheus `ServiceMonitor`.         |
| `serviceMonitor.interval`   | `string`            | `30s`   | Prometheus scrape interval.                   |
| `serviceMonitor.labels`     | `map[string]string` | --      | Labels to add to the ServiceMonitor (for Prometheus selector matching). |

#### spec.observability.logging

| Field    | Type     | Default | Description                                              |
|----------|----------|---------|----------------------------------------------------------|
| `level`  | `string` | `info`  | Log level. One of: `debug`, `info`, `warn`, `error`.     |
| `format` | `string` | `json`  | Log format. One of: `json`, `text`.                      |

### spec.availability

High availability and scheduling configuration.

| Field                             | Type                | Default | Description                                              |
|-----------------------------------|---------------------|---------|----------------------------------------------------------|
| `podDisruptionBudget.enabled`     | `*bool`             | `true`  | Create a PodDisruptionBudget.                            |
| `podDisruptionBudget.maxUnavailable` | `*int32`         | `1`     | Maximum pods that can be unavailable during disruption.  |
| `nodeSelector`                    | `map[string]string` | --      | Node labels for pod scheduling.                          |
| `tolerations`                     | `[]Toleration`      | --      | Tolerations for pod scheduling.                          |
| `affinity`                        | `*Affinity`         | --      | Affinity and anti-affinity rules.                        |

---

## Status Fields

### status.phase

| Field   | Type     | Description                                                                    |
|---------|----------|--------------------------------------------------------------------------------|
| `phase` | `string` | Current lifecycle phase: `Pending`, `Provisioning`, `Running`, `Degraded`, `Failed`, `Terminating`. |

### status.conditions

Standard `metav1.Condition` array. Condition types:

| Type                | Description                                      |
|---------------------|--------------------------------------------------|
| `Ready`             | Overall readiness of the instance.               |
| `ConfigValid`       | Configuration is valid and loaded.               |
| `DeploymentReady`   | Deployment has ready replicas.                   |
| `ServiceReady`      | Service has been created.                        |
| `NetworkPolicyReady`| NetworkPolicy has been applied.                  |
| `RBACReady`         | RBAC resources are in place.                     |
| `StorageReady`      | PVC has been provisioned.                        |

### status.endpoints

| Field              | Type     | Description                                                  |
|--------------------|----------|--------------------------------------------------------------|
| `gatewayEndpoint`  | `string` | In-cluster endpoint for the gateway: `<name>.<ns>.svc:18789`.|
| `canvasEndpoint`   | `string` | In-cluster endpoint for canvas: `<name>.<ns>.svc:18793`.     |

### status.observedGeneration

| Field                | Type    | Description                                              |
|----------------------|---------|----------------------------------------------------------|
| `observedGeneration` | `int64` | The `.metadata.generation` last processed by the controller. |

### status.lastReconcileTime

| Field               | Type          | Description                                     |
|---------------------|---------------|-------------------------------------------------|
| `lastReconcileTime` | `*metav1.Time`| Timestamp of the last successful reconciliation.|

### status.managedResources

| Field                | Type     | Description                           |
|----------------------|----------|---------------------------------------|
| `deployment`         | `string` | Name of the managed Deployment.       |
| `service`            | `string` | Name of the managed Service.          |
| `configMap`          | `string` | Name of the managed ConfigMap.        |
| `pvc`                | `string` | Name of the managed PVC.             |
| `networkPolicy`      | `string` | Name of the managed NetworkPolicy.    |
| `podDisruptionBudget`| `string` | Name of the managed PDB.             |
| `serviceAccount`     | `string` | Name of the managed ServiceAccount.   |
| `role`               | `string` | Name of the managed Role.            |
| `roleBinding`        | `string` | Name of the managed RoleBinding.      |

---

## Full Example

```yaml
apiVersion: openclaw.openclaw.io/v1alpha1
kind: OpenClawInstance
metadata:
  name: my-assistant
  namespace: openclaw
spec:
  image:
    repository: ghcr.io/openclaw/openclaw
    tag: "0.5.0"
    pullPolicy: IfNotPresent
    pullSecrets:
      - name: ghcr-secret

  config:
    raw:
      mcpServers:
        filesystem:
          command: npx
          args: ["-y", "@modelcontextprotocol/server-filesystem", "/data"]

  envFrom:
    - secretRef:
        name: openclaw-api-keys

  resources:
    requests:
      cpu: "1"
      memory: 2Gi
    limits:
      cpu: "4"
      memory: 8Gi

  security:
    podSecurityContext:
      runAsUser: 1000
      runAsGroup: 1000
      fsGroup: 1000
      runAsNonRoot: true
    containerSecurityContext:
      allowPrivilegeEscalation: false
    networkPolicy:
      enabled: true
      allowedIngressNamespaces:
        - monitoring
      allowDNS: true
    rbac:
      createServiceAccount: true

  storage:
    persistence:
      enabled: true
      storageClass: gp3
      size: 50Gi
      accessModes:
        - ReadWriteOnce

  chromium:
    enabled: true
    image:
      repository: ghcr.io/browserless/chromium
      tag: "v2.0.0"
    resources:
      requests:
        cpu: 500m
        memory: 1Gi
      limits:
        cpu: "2"
        memory: 4Gi

  networking:
    service:
      type: ClusterIP
    ingress:
      enabled: true
      className: nginx
      hosts:
        - host: openclaw.example.com
          paths:
            - path: /
              pathType: Prefix
      tls:
        - hosts:
            - openclaw.example.com
          secretName: openclaw-tls
      security:
        forceHTTPS: true
        enableHSTS: true
        rateLimiting:
          enabled: true
          requestsPerSecond: 20

  probes:
    liveness:
      enabled: true
      initialDelaySeconds: 60
      periodSeconds: 15
    readiness:
      enabled: true
      initialDelaySeconds: 10
    startup:
      enabled: true
      failureThreshold: 60

  observability:
    metrics:
      enabled: true
      serviceMonitor:
        enabled: true
        interval: 15s
        labels:
          release: prometheus
    logging:
      level: info
      format: json

  availability:
    podDisruptionBudget:
      enabled: true
      maxUnavailable: 1
    nodeSelector:
      node-type: compute
    tolerations:
      - key: dedicated
        operator: Equal
        value: openclaw
        effect: NoSchedule
    affinity:
      podAntiAffinity:
        preferredDuringSchedulingIgnoredDuringExecution:
          - weight: 100
            podAffinityTerm:
              topologyKey: kubernetes.io/hostname
              labelSelector:
                matchLabels:
                  app.kubernetes.io/name: openclaw
```
