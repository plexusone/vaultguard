# JSON Schema Reference

Complete reference for the VaultGuard policy JSON configuration file.

## File Structure

```json
{
  "version": 1,
  "local": { ... },
  "cloud": { ... },
  "kubernetes": { ... },
  "provider_map": { ... },
  "fallback_provider": "string",
  "allow_insecure": false,
  "insecure_reason": "string",
  "locked": [ ... ]
}
```

## Top-Level Fields

| Field | Type | Required | Description |
|-------|------|----------|-------------|
| `version` | integer | Yes | Policy file format version. Currently `1`. |
| `local` | object | No | Local workstation security requirements. |
| `cloud` | object | No | Cloud environment security requirements. |
| `kubernetes` | object | No | Kubernetes-specific security requirements. |
| `provider_map` | object | No | Maps environments to secret providers. |
| `fallback_provider` | string | No | Provider to use when preferred is unavailable. |
| `allow_insecure` | boolean | No | Allow access even if security checks fail. Default: `false`. |
| `insecure_reason` | string | No | Documentation for why `allow_insecure` is set. |
| `locked` | array | No | Field paths that cannot be overridden by user config. |

## Local Policy Object

Security requirements for local workstations (macOS, Windows, Linux).

```json
{
  "local": {
    "min_security_score": 50,
    "require_encryption": true,
    "require_tpm": false,
    "require_secure_boot": false,
    "require_biometrics": false,
    "allowed_platforms": ["darwin", "linux"]
  }
}
```

| Field | Type | Default | Description |
|-------|------|---------|-------------|
| `min_security_score` | integer | `0` | Minimum Posture security score (0-100). |
| `require_encryption` | boolean | `false` | Require disk encryption (FileVault/BitLocker/LUKS). |
| `require_tpm` | boolean | `false` | Require TPM or Secure Enclave. |
| `require_secure_boot` | boolean | `false` | Require Secure Boot to be enabled. |
| `require_biometrics` | boolean | `false` | Require biometric authentication configured. |
| `allowed_platforms` | array | `[]` | Restrict to specific platforms. Empty = all allowed. |

**Allowed platform values:** `darwin`, `windows`, `linux`

## Cloud Policy Object

Security requirements for cloud environments.

```json
{
  "cloud": {
    "require_iam": true,
    "aws": { ... },
    "gcp": { ... },
    "azure": { ... }
  }
}
```

| Field | Type | Default | Description |
|-------|------|---------|-------------|
| `require_iam` | boolean | `false` | Require cloud IAM to be configured. |
| `aws` | object | `null` | AWS-specific requirements. |
| `gcp` | object | `null` | GCP-specific requirements. |
| `azure` | object | `null` | Azure-specific requirements. |

### AWS Policy Object

```json
{
  "cloud": {
    "aws": {
      "require_irsa": true,
      "allowed_role_arns": ["arn:aws:iam::123456789012:role/my-app-*"],
      "allowed_account_ids": ["123456789012"],
      "allowed_regions": ["us-east-1", "us-west-2"],
      "require_imdsv2": true
    }
  }
}
```

| Field | Type | Default | Description |
|-------|------|---------|-------------|
| `require_irsa` | boolean | `false` | Require IRSA (IAM Roles for Service Accounts). |
| `allowed_role_arns` | array | `[]` | Whitelist of IAM role ARNs. Supports wildcards. |
| `allowed_account_ids` | array | `[]` | Whitelist of AWS account IDs. |
| `allowed_regions` | array | `[]` | Whitelist of AWS regions. |
| `require_imdsv2` | boolean | `false` | Require IMDSv2 for EC2 instances. |

### GCP Policy Object

```json
{
  "cloud": {
    "gcp": {
      "require_workload_identity": true,
      "allowed_service_accounts": ["app@project.iam.gserviceaccount.com"],
      "allowed_projects": ["my-project"],
      "allowed_regions": ["us-central1"]
    }
  }
}
```

| Field | Type | Default | Description |
|-------|------|---------|-------------|
| `require_workload_identity` | boolean | `false` | Require GKE Workload Identity. |
| `allowed_service_accounts` | array | `[]` | Whitelist of GCP service account emails. |
| `allowed_projects` | array | `[]` | Whitelist of GCP project IDs. |
| `allowed_regions` | array | `[]` | Whitelist of GCP regions. |

### Azure Policy Object

```json
{
  "cloud": {
    "azure": {
      "require_workload_identity": true,
      "allowed_client_ids": ["12345678-1234-1234-1234-123456789012"],
      "allowed_tenant_ids": ["87654321-4321-4321-4321-210987654321"],
      "allowed_subscriptions": ["aaaaaaaa-bbbb-cccc-dddd-eeeeeeeeeeee"],
      "allowed_regions": ["eastus", "westus2"]
    }
  }
}
```

| Field | Type | Default | Description |
|-------|------|---------|-------------|
| `require_workload_identity` | boolean | `false` | Require AKS Workload Identity. |
| `allowed_client_ids` | array | `[]` | Whitelist of Azure AD client IDs. |
| `allowed_tenant_ids` | array | `[]` | Whitelist of Azure AD tenant IDs. |
| `allowed_subscriptions` | array | `[]` | Whitelist of Azure subscription IDs. |
| `allowed_regions` | array | `[]` | Whitelist of Azure regions. |

## Kubernetes Policy Object

Additional requirements for Kubernetes environments.

```json
{
  "kubernetes": {
    "require_service_account": true,
    "allowed_service_accounts": ["my-app-sa"],
    "allowed_namespaces": ["production", "staging"],
    "denied_namespaces": ["default", "kube-system"],
    "require_non_root": true,
    "require_read_only_root": true
  }
}
```

| Field | Type | Default | Description |
|-------|------|---------|-------------|
| `require_service_account` | boolean | `false` | Require non-default service account. |
| `allowed_service_accounts` | array | `[]` | Whitelist of service accounts. |
| `allowed_namespaces` | array | `[]` | Whitelist of namespaces. |
| `denied_namespaces` | array | `[]` | Blacklist of namespaces. |
| `require_non_root` | boolean | `false` | Require container to run as non-root. |
| `require_read_only_root` | boolean | `false` | Require read-only root filesystem. |

!!! note
    When `allowed_namespaces` is set, it takes precedence over `denied_namespaces`.

## Provider Map Object

Maps environments to secret providers.

```json
{
  "provider_map": {
    "local": "keyring",
    "container": "env",
    "kubernetes": "k8s",
    "eks": "aws-sm",
    "lambda": "aws-sm",
    "gke": "gcp-sm",
    "cloudrun": "gcp-sm",
    "aks": "azure-kv"
  }
}
```

### Environment Keys

| Key | Description |
|-----|-------------|
| `local` | Local workstation |
| `container` | Generic container |
| `kubernetes` | Kubernetes without cloud IAM |
| `eks` | AWS EKS with IRSA |
| `lambda` | AWS Lambda |
| `gke` | GCP GKE with Workload Identity |
| `cloudrun` | GCP Cloud Run |
| `aks` | Azure AKS with Workload Identity |
| `azurefunc` | Azure Functions |

### Provider Values

| Value | Description |
|-------|-------------|
| `env` | Environment variables |
| `file` | File-based secrets |
| `keyring` | OS keyring (Keychain, Credential Manager) |
| `aws-sm` | AWS Secrets Manager |
| `aws-ssm` | AWS Systems Manager Parameter Store |
| `gcp-sm` | GCP Secret Manager |
| `azure-kv` | Azure Key Vault |
| `k8s` | Kubernetes Secrets |
| `vault` | HashiCorp Vault |

## Locked Fields Array

Field paths that cannot be overridden by user configuration (enterprise policies only).

```json
{
  "locked": [
    "local.require_encryption",
    "local.min_security_score",
    "cloud.require_iam",
    "provider_map.eks",
    "allow_insecure"
  ]
}
```

### Available Field Paths

**Local Policy:**

- `local.min_security_score`
- `local.require_encryption`
- `local.require_tpm`
- `local.require_secure_boot`
- `local.require_biometrics`
- `local.allowed_platforms`

**Cloud Policy:**

- `cloud.require_iam`

**Provider Map:**

- `provider_map.local`
- `provider_map.container`
- `provider_map.kubernetes`
- `provider_map.eks`
- `provider_map.lambda`
- `provider_map.gke`
- `provider_map.cloudrun`
- `provider_map.aks`
- `provider_map.azurefunc`

**Other:**

- `fallback_provider`
- `allow_insecure`

## Complete Example

```json
{
  "version": 1,
  "local": {
    "min_security_score": 60,
    "require_encryption": true,
    "require_tpm": true,
    "require_secure_boot": false,
    "require_biometrics": false,
    "allowed_platforms": []
  },
  "cloud": {
    "require_iam": true,
    "aws": {
      "require_irsa": true,
      "allowed_role_arns": [
        "arn:aws:iam::123456789012:role/production-*"
      ],
      "allowed_account_ids": ["123456789012"],
      "allowed_regions": ["us-east-1", "us-west-2"],
      "require_imdsv2": true
    },
    "gcp": {
      "require_workload_identity": true,
      "allowed_service_accounts": [
        "my-app@my-project.iam.gserviceaccount.com"
      ],
      "allowed_projects": ["my-project"],
      "allowed_regions": ["us-central1", "us-east1"]
    },
    "azure": {
      "require_workload_identity": true,
      "allowed_client_ids": ["12345678-1234-1234-1234-123456789012"],
      "allowed_tenant_ids": ["87654321-4321-4321-4321-210987654321"],
      "allowed_subscriptions": [],
      "allowed_regions": ["eastus", "westus2"]
    }
  },
  "kubernetes": {
    "require_service_account": true,
    "allowed_service_accounts": ["my-app-sa"],
    "allowed_namespaces": [],
    "denied_namespaces": ["default", "kube-system", "kube-public"],
    "require_non_root": true,
    "require_read_only_root": false
  },
  "provider_map": {
    "local": "keyring",
    "eks": "aws-sm",
    "lambda": "aws-sm",
    "gke": "gcp-sm",
    "cloudrun": "gcp-sm",
    "aks": "azure-kv"
  },
  "fallback_provider": "env",
  "allow_insecure": false,
  "insecure_reason": "",
  "locked": [
    "local.require_encryption",
    "cloud.require_iam",
    "allow_insecure"
  ]
}
```

## Validation

VaultGuard validates policy files when loading:

- JSON must be valid
- `version` must be a supported version (currently `1`)
- Field values must be valid types
- Provider and environment strings must be recognized values

Invalid policies result in an error from `LoadPolicy()` or `LoadPolicyFromFile()`.

## Next Steps

- [Example Configs](../configuration/examples.md) - Ready-to-use policy files
- [Enterprise Policies](../policies/enterprise.md) - Understanding locked fields
- [File Locations](../configuration/file-locations.md) - Platform-specific paths
