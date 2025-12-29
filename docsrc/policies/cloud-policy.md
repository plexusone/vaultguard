# Cloud Policy

Cloud policies define security requirements for applications running in cloud environments like AWS, GCP, and Azure. These policies ensure proper IAM configuration and restrict access to authorized accounts, regions, and identities.

## Overview

Cloud policies are evaluated when `Environment.IsCloud()` returns true (EKS, GKE, AKS, Lambda, Cloud Run, Azure Functions).

```go
type CloudPolicy struct {
    RequireIAM bool        // Require cloud IAM to be configured
    AWS        *AWSPolicy  // AWS-specific requirements
    GCP        *GCPPolicy  // GCP-specific requirements
    Azure      *AzurePolicy // Azure-specific requirements
}
```

## Common Field: RequireIAM

Requires cloud-native IAM to be configured before allowing credential access:

```json
{
  "cloud": {
    "require_iam": true
  }
}
```

This ensures workloads are using:

- **AWS**: IRSA (IAM Roles for Service Accounts) or instance roles
- **GCP**: Workload Identity or service account keys
- **Azure**: Workload Identity or managed identity

!!! tip
    Always set `require_iam: true` in production. Environment variable credentials are a security risk in cloud environments.

## AWS Policy

### Fields

```go
type AWSPolicy struct {
    RequireIRSA       bool     // Require IRSA specifically
    AllowedRoleARNs   []string // Whitelist of IAM role ARNs
    AllowedAccountIDs []string // Whitelist of AWS account IDs
    AllowedRegions    []string // Whitelist of AWS regions
    RequireIMDSv2     bool     // Require IMDSv2 for EC2
}
```

### RequireIRSA

Requires IAM Roles for Service Accounts (IRSA) in EKS:

```json
{
  "cloud": {
    "aws": {
      "require_irsa": true
    }
  }
}
```

IRSA provides:

- Pod-level IAM permissions (not node-level)
- No long-lived credentials
- Automatic credential rotation

### AllowedRoleARNs

Restricts to specific IAM roles. Supports wildcards:

```json
{
  "cloud": {
    "aws": {
      "allowed_role_arns": [
        "arn:aws:iam::123456789012:role/my-app-prod",
        "arn:aws:iam::123456789012:role/my-app-*"
      ]
    }
  }
}
```

### AllowedAccountIDs

Restricts to specific AWS accounts:

```json
{
  "cloud": {
    "aws": {
      "allowed_account_ids": ["123456789012", "987654321098"]
    }
  }
}
```

### AllowedRegions

Restricts to specific AWS regions:

```json
{
  "cloud": {
    "aws": {
      "allowed_regions": ["us-east-1", "us-west-2", "eu-west-1"]
    }
  }
}
```

### RequireIMDSv2

Requires IMDSv2 (Instance Metadata Service v2) on EC2:

```json
{
  "cloud": {
    "aws": {
      "require_imdsv2": true
    }
  }
}
```

IMDSv2 protects against SSRF attacks that could steal instance credentials.

### Complete AWS Example

```json
{
  "cloud": {
    "require_iam": true,
    "aws": {
      "require_irsa": true,
      "require_imdsv2": true,
      "allowed_account_ids": ["123456789012"],
      "allowed_role_arns": [
        "arn:aws:iam::123456789012:role/production-*"
      ],
      "allowed_regions": ["us-east-1", "us-west-2"]
    }
  }
}
```

## GCP Policy

### Fields

```go
type GCPPolicy struct {
    RequireWorkloadIdentity  bool     // Require GKE Workload Identity
    AllowedServiceAccounts   []string // Whitelist of service accounts
    AllowedProjects          []string // Whitelist of GCP projects
    AllowedRegions           []string // Whitelist of GCP regions
}
```

### RequireWorkloadIdentity

Requires Workload Identity in GKE:

```json
{
  "cloud": {
    "gcp": {
      "require_workload_identity": true
    }
  }
}
```

Workload Identity provides:

- Kubernetes service account to GCP service account mapping
- No service account key files
- Automatic credential management

### AllowedServiceAccounts

Restricts to specific GCP service accounts:

```json
{
  "cloud": {
    "gcp": {
      "allowed_service_accounts": [
        "my-app@my-project.iam.gserviceaccount.com",
        "other-app@my-project.iam.gserviceaccount.com"
      ]
    }
  }
}
```

### AllowedProjects

Restricts to specific GCP projects:

```json
{
  "cloud": {
    "gcp": {
      "allowed_projects": ["my-prod-project", "my-staging-project"]
    }
  }
}
```

### AllowedRegions

Restricts to specific GCP regions:

```json
{
  "cloud": {
    "gcp": {
      "allowed_regions": ["us-central1", "us-east1", "europe-west1"]
    }
  }
}
```

### Complete GCP Example

```json
{
  "cloud": {
    "require_iam": true,
    "gcp": {
      "require_workload_identity": true,
      "allowed_projects": ["my-prod-project"],
      "allowed_service_accounts": [
        "my-app@my-prod-project.iam.gserviceaccount.com"
      ],
      "allowed_regions": ["us-central1", "us-east1"]
    }
  }
}
```

## Azure Policy

### Fields

```go
type AzurePolicy struct {
    RequireWorkloadIdentity bool     // Require AKS Workload Identity
    AllowedClientIDs        []string // Whitelist of Azure AD client IDs
    AllowedTenantIDs        []string // Whitelist of Azure AD tenant IDs
    AllowedSubscriptions    []string // Whitelist of Azure subscriptions
    AllowedRegions          []string // Whitelist of Azure regions
}
```

### RequireWorkloadIdentity

Requires Workload Identity in AKS:

```json
{
  "cloud": {
    "azure": {
      "require_workload_identity": true
    }
  }
}
```

### AllowedClientIDs

Restricts to specific Azure AD application client IDs:

```json
{
  "cloud": {
    "azure": {
      "allowed_client_ids": [
        "12345678-1234-1234-1234-123456789012"
      ]
    }
  }
}
```

### AllowedTenantIDs

Restricts to specific Azure AD tenants:

```json
{
  "cloud": {
    "azure": {
      "allowed_tenant_ids": [
        "87654321-4321-4321-4321-210987654321"
      ]
    }
  }
}
```

### AllowedSubscriptions

Restricts to specific Azure subscriptions:

```json
{
  "cloud": {
    "azure": {
      "allowed_subscriptions": [
        "aaaaaaaa-bbbb-cccc-dddd-eeeeeeeeeeee"
      ]
    }
  }
}
```

### AllowedRegions

Restricts to specific Azure regions:

```json
{
  "cloud": {
    "azure": {
      "allowed_regions": ["eastus", "westus2", "westeurope"]
    }
  }
}
```

### Complete Azure Example

```json
{
  "cloud": {
    "require_iam": true,
    "azure": {
      "require_workload_identity": true,
      "allowed_tenant_ids": ["87654321-4321-4321-4321-210987654321"],
      "allowed_subscriptions": ["aaaaaaaa-bbbb-cccc-dddd-eeeeeeeeeeee"],
      "allowed_regions": ["eastus", "westus2"]
    }
  }
}
```

## Multi-Cloud Policy

You can define requirements for all cloud providers in a single policy:

```json
{
  "cloud": {
    "require_iam": true,
    "aws": {
      "require_irsa": true,
      "allowed_account_ids": ["123456789012"]
    },
    "gcp": {
      "require_workload_identity": true,
      "allowed_projects": ["my-prod-project"]
    },
    "azure": {
      "require_workload_identity": true,
      "allowed_tenant_ids": ["87654321-4321-4321-4321-210987654321"]
    }
  }
}
```

VaultGuard applies the appropriate section based on the detected environment.

## Default Providers

| Environment | Default Provider |
|-------------|-----------------|
| EKS | AWS Secrets Manager (`aws-sm`) |
| Lambda | AWS Secrets Manager (`aws-sm`) |
| GKE | GCP Secret Manager (`gcp-sm`) |
| Cloud Run | GCP Secret Manager (`gcp-sm`) |
| AKS | Azure Key Vault (`azure-kv`) |

Override with `provider_map`:

```json
{
  "provider_map": {
    "eks": "aws-ssm",
    "gke": "env"
  }
}
```

## Next Steps

- [Kubernetes Policy](kubernetes-policy.md) - Additional K8s requirements
- [Enterprise Policies](enterprise.md) - Lock cloud settings organization-wide
- [Example Configs](../configuration/examples.md) - Ready-to-use policy files
