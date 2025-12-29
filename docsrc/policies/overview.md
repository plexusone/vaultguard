# Policies Overview

VaultGuard uses policies to define security requirements that must be met before credentials can be accessed. This security-first approach ensures that sensitive data is only available in properly secured environments.

## What is a Policy?

A policy is a set of rules that define:

1. **Security requirements** - What security features must be present (encryption, TPM, IAM roles, etc.)
2. **Provider mappings** - Which secret provider to use in each environment
3. **Access restrictions** - Allowed accounts, regions, namespaces, etc.

## Policy Structure

```go
type Policy struct {
    // Security requirements for local workstations
    Local *LocalPolicy

    // Security requirements for cloud environments
    Cloud *CloudPolicy

    // Security requirements for Kubernetes
    Kubernetes *KubernetesPolicy

    // Map environments to secret providers
    ProviderMap map[Environment]Provider

    // Fallback provider when preferred is unavailable
    FallbackProvider Provider

    // Allow access even if security checks fail (dev only!)
    AllowInsecure bool
    InsecureReason string
}
```

## Built-in Policies

VaultGuard provides three built-in policies:

### Default Policy

Balanced security suitable for most production environments:

```go
sv, _ := vaultguard.Quick() // Uses DefaultPolicy()
```

- Local: Requires 50+ security score, disk encryption
- Cloud: Requires IAM (IRSA, Workload Identity)
- Kubernetes: Denies default, kube-system, kube-public namespaces
- Providers: Keyring for local, cloud secret managers for cloud

### Development Policy

Permissive settings for local development:

```go
sv, _ := vaultguard.QuickDev() // Uses DevelopmentPolicy()
```

- No security score minimum
- IAM not required
- All providers default to environment variables
- `AllowInsecure: true`

!!! danger
    Never use the development policy in production. It bypasses security checks entirely.

### Strict Policy

High-security requirements for sensitive environments:

```go
sv, _ := vaultguard.QuickStrict() // Uses StrictPolicy()
```

- Local: Requires 75+ score, encryption, TPM, Secure Boot
- Cloud: Requires IAM, IMDSv2 (AWS), Workload Identity
- Kubernetes: Requires service account, non-root, read-only root filesystem
- `AllowInsecure: false` (cannot be overridden)

## Policy Evaluation Flow

When you request credentials, VaultGuard:

```
1. Detect Environment
   └─→ local | eks | gke | aks | lambda | cloudrun | ...

2. Load Applicable Policy Section
   └─→ Local environment? Use policy.Local
   └─→ Cloud environment? Use policy.Cloud
   └─→ Kubernetes? Also apply policy.Kubernetes

3. Run Security Checks
   └─→ Local: Posture assessment (encryption, TPM, score)
   └─→ Cloud: IAM validation, account/region checks
   └─→ K8s: Namespace, service account validation

4. Evaluate Results
   └─→ All checks pass? → Allow credential access
   └─→ Checks fail + AllowInsecure? → Allow with warning
   └─→ Checks fail? → Deny access, return error
```

## Policy Sources

Policies can come from multiple sources:

### 1. Code (Inline)

```go
sv, err := vaultguard.New(&vaultguard.Config{
    Policy: &vaultguard.Policy{
        Local: &vaultguard.LocalPolicy{
            RequireEncryption: true,
        },
    },
})
```

### 2. Built-in Functions

```go
policy := vaultguard.DefaultPolicy()
policy := vaultguard.DevelopmentPolicy()
policy := vaultguard.StrictPolicy()
```

### 3. Configuration Files

```go
policy, err := vaultguard.LoadPolicy()
```

Loads from (in order of precedence):

1. `AGENTPLEXUS_POLICY_FILE` environment variable
2. `~/.agentplexus/policy.json` (user config)
3. `/etc/agentplexus/policy.json` (system/enterprise config)

See [File Locations](../configuration/file-locations.md) for platform-specific paths.

## Policy Merging

When both system (enterprise) and user configuration files exist, they are merged:

- System policy provides the base configuration
- User policy can customize non-locked fields
- **Locked fields** in system policy cannot be overridden by users
- `AllowInsecure` can only become *more* restrictive, never more permissive

```json
// /etc/agentplexus/policy.json (system)
{
  "version": 1,
  "local": {
    "require_encryption": true,
    "min_security_score": 50
  },
  "locked": ["local.require_encryption"]
}
```

```json
// ~/.agentplexus/policy.json (user)
{
  "version": 1,
  "local": {
    "require_encryption": false,  // IGNORED - field is locked
    "min_security_score": 75      // Applied - raises the minimum
  }
}
```

Result: `require_encryption: true`, `min_security_score: 75`

See [Enterprise Policies](enterprise.md) for details on locked fields.

## Next Steps

- [Local Policy](local-policy.md) - Workstation security requirements
- [Cloud Policy](cloud-policy.md) - AWS, GCP, Azure requirements
- [Kubernetes Policy](kubernetes-policy.md) - Kubernetes-specific settings
- [Enterprise Policies](enterprise.md) - Organization-wide enforcement
