# Kubernetes Policy

Kubernetes policies define additional security requirements for workloads running in Kubernetes clusters. These policies are applied alongside cloud policies (for EKS, GKE, AKS) or standalone for generic Kubernetes deployments.

## Overview

Kubernetes policies are evaluated when `Environment.IsKubernetes()` returns true (EKS, GKE, AKS, or generic Kubernetes).

```go
type KubernetesPolicy struct {
    RequireServiceAccount  bool     // Require a specific service account
    AllowedServiceAccounts []string // Whitelist of allowed service accounts
    AllowedNamespaces      []string // Whitelist of allowed namespaces
    DeniedNamespaces       []string // Blacklist of denied namespaces
    RequireNonRoot         bool     // Require non-root container
    RequireReadOnlyRoot    bool     // Require read-only root filesystem
}
```

## Fields

### RequireServiceAccount

Requires the pod to be running with a specific (non-default) service account:

```json
{
  "kubernetes": {
    "require_service_account": true
  }
}
```

This prevents workloads from running with the `default` service account, which often has no specific permissions configured.

### AllowedServiceAccounts

Restricts to specific Kubernetes service accounts:

```json
{
  "kubernetes": {
    "allowed_service_accounts": [
      "my-app-sa",
      "my-app-worker-sa"
    ]
  }
}
```

### AllowedNamespaces

Restricts to specific namespaces (whitelist):

```json
{
  "kubernetes": {
    "allowed_namespaces": [
      "production",
      "staging"
    ]
  }
}
```

!!! note
    When `allowed_namespaces` is set, only pods in those namespaces can access credentials. This takes precedence over `denied_namespaces`.

### DeniedNamespaces

Blocks specific namespaces (blacklist):

```json
{
  "kubernetes": {
    "denied_namespaces": [
      "default",
      "kube-system",
      "kube-public"
    ]
  }
}
```

Common namespaces to deny:

| Namespace | Reason |
|-----------|--------|
| `default` | Should not run production workloads |
| `kube-system` | Reserved for cluster components |
| `kube-public` | Publicly readable, not for secrets |
| `kube-node-lease` | System namespace for node heartbeats |

### RequireNonRoot

Requires the container to run as a non-root user:

```json
{
  "kubernetes": {
    "require_non_root": true
  }
}
```

This maps to the pod security context:

```yaml
securityContext:
  runAsNonRoot: true
```

### RequireReadOnlyRoot

Requires the container's root filesystem to be read-only:

```json
{
  "kubernetes": {
    "require_read_only_root": true
  }
}
```

This maps to the container security context:

```yaml
securityContext:
  readOnlyRootFilesystem: true
```

## Examples

### Minimal (Deny Default Namespaces)

```json
{
  "kubernetes": {
    "denied_namespaces": ["default", "kube-system", "kube-public"]
  }
}
```

### Standard Production

```json
{
  "kubernetes": {
    "require_service_account": true,
    "denied_namespaces": ["default", "kube-system", "kube-public"],
    "require_non_root": true
  }
}
```

### High Security

```json
{
  "kubernetes": {
    "require_service_account": true,
    "allowed_namespaces": ["production"],
    "allowed_service_accounts": ["my-app-sa"],
    "require_non_root": true,
    "require_read_only_root": true
  }
}
```

### Multi-Tenant Cluster

```json
{
  "kubernetes": {
    "require_service_account": true,
    "allowed_namespaces": [
      "team-a-prod",
      "team-a-staging",
      "team-b-prod",
      "team-b-staging"
    ],
    "denied_namespaces": ["default", "kube-system", "kube-public"]
  }
}
```

## Combined with Cloud Policy

Kubernetes policies work alongside cloud policies. For EKS, GKE, or AKS, both are evaluated:

```json
{
  "cloud": {
    "require_iam": true,
    "aws": {
      "require_irsa": true,
      "allowed_account_ids": ["123456789012"]
    }
  },
  "kubernetes": {
    "require_service_account": true,
    "denied_namespaces": ["default", "kube-system"],
    "require_non_root": true
  }
}
```

For an EKS workload, VaultGuard checks:

1. IRSA is configured with an allowed role
2. Running in an allowed AWS account
3. Not in a denied namespace
4. Running as non-root
5. Using a specific service account

## Pod Security Standards Alignment

VaultGuard's Kubernetes policy aligns with [Pod Security Standards](https://kubernetes.io/docs/concepts/security/pod-security-standards/):

| VaultGuard Setting | PSS Level | PSS Requirement |
|-------------------|-----------|-----------------|
| `require_non_root: true` | Baseline | `runAsNonRoot: true` |
| `require_read_only_root: true` | Restricted | `readOnlyRootFilesystem: true` |

!!! tip
    If your cluster enforces Pod Security Standards, ensure your VaultGuard policy requirements don't conflict with what the cluster allows.

## Checking Kubernetes Environment

```go
result, err := vaultguard.CheckSecurity(nil)
if err != nil {
    log.Fatal(err)
}

if result.Details.Cloud != nil && result.Details.Cloud.Kubernetes != nil {
    k8s := result.Details.Cloud.Kubernetes
    fmt.Printf("In Cluster: %v\n", k8s.InCluster)
    fmt.Printf("Namespace: %s\n", k8s.Namespace)
    fmt.Printf("Service Account: %s\n", k8s.ServiceAccount)
    fmt.Printf("Pod Name: %s\n", k8s.PodName)
}
```

## Default Provider

For generic Kubernetes (not EKS/GKE/AKS), the default provider is `k8s` (Kubernetes Secrets).

For managed Kubernetes with cloud IAM:

| Environment | Default Provider |
|-------------|-----------------|
| EKS | AWS Secrets Manager |
| GKE | GCP Secret Manager |
| AKS | Azure Key Vault |

Override with `provider_map`:

```json
{
  "provider_map": {
    "kubernetes": "vault",
    "eks": "aws-sm"
  }
}
```

## Next Steps

- [Enterprise Policies](enterprise.md) - Lock K8s settings organization-wide
- [Cloud Policy](cloud-policy.md) - Cloud-specific requirements
- [Example Configs](../configuration/examples.md) - Ready-to-use policy files
