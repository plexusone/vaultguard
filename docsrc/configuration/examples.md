# Example Configs

This page provides ready-to-use policy configuration files for common scenarios.

## User Configurations

These go in `~/.agentplexus/policy.json` (Linux/macOS) or `%USERPROFILE%\.agentplexus\policy.json` (Windows).

### Minimal User Policy

Basic policy with encryption requirement:

```json
{
  "version": 1,
  "local": {
    "min_security_score": 50,
    "require_encryption": true
  }
}
```

### Developer Workstation

Balanced security for daily development:

```json
{
  "version": 1,
  "local": {
    "min_security_score": 40,
    "require_encryption": true
  },
  "provider_map": {
    "local": "keyring"
  },
  "fallback_provider": "env"
}
```

### Security-Conscious Developer

Higher standards for handling sensitive credentials:

```json
{
  "version": 1,
  "local": {
    "min_security_score": 70,
    "require_encryption": true,
    "require_tpm": true
  },
  "provider_map": {
    "local": "keyring"
  }
}
```

### macOS with Touch ID

Leveraging Apple Silicon security features:

```json
{
  "version": 1,
  "local": {
    "min_security_score": 60,
    "require_encryption": true,
    "require_tpm": true,
    "require_biometrics": true,
    "allowed_platforms": ["darwin"]
  },
  "provider_map": {
    "local": "keyring"
  }
}
```

### Environment Variables Only

For development containers or CI:

```json
{
  "version": 1,
  "local": {
    "min_security_score": 0
  },
  "provider_map": {
    "local": "env",
    "container": "env",
    "kubernetes": "env"
  },
  "fallback_provider": "env",
  "allow_insecure": true,
  "insecure_reason": "Development/CI environment"
}
```

!!! warning
    Only use `allow_insecure: true` in development. Never in production.

## Enterprise Configurations

These go in `/etc/agentplexus/policy.json` (Linux/macOS) or `%ProgramData%\agentplexus\policy.json` (Windows).

### Basic Enterprise Policy

Enforces encryption, lets users customize other settings:

```json
{
  "version": 1,
  "local": {
    "min_security_score": 50,
    "require_encryption": true
  },
  "locked": [
    "local.require_encryption"
  ]
}
```

### Standard Enterprise Policy

Balanced security with cloud IAM requirements:

```json
{
  "version": 1,
  "local": {
    "min_security_score": 50,
    "require_encryption": true
  },
  "cloud": {
    "require_iam": true
  },
  "kubernetes": {
    "denied_namespaces": ["default", "kube-system", "kube-public"]
  },
  "locked": [
    "local.require_encryption",
    "cloud.require_iam"
  ]
}
```

### Strict Enterprise Policy

High-security requirements for regulated industries:

```json
{
  "version": 1,
  "local": {
    "min_security_score": 75,
    "require_encryption": true,
    "require_tpm": true,
    "require_secure_boot": true
  },
  "cloud": {
    "require_iam": true,
    "aws": {
      "require_irsa": true,
      "require_imdsv2": true
    },
    "gcp": {
      "require_workload_identity": true
    },
    "azure": {
      "require_workload_identity": true
    }
  },
  "kubernetes": {
    "require_service_account": true,
    "denied_namespaces": ["default", "kube-system", "kube-public"],
    "require_non_root": true,
    "require_read_only_root": true
  },
  "allow_insecure": false,
  "locked": [
    "local.min_security_score",
    "local.require_encryption",
    "local.require_tpm",
    "local.require_secure_boot",
    "cloud.require_iam",
    "allow_insecure"
  ]
}
```

### AWS-Only Enterprise

For organizations standardized on AWS:

```json
{
  "version": 1,
  "local": {
    "min_security_score": 50,
    "require_encryption": true
  },
  "cloud": {
    "require_iam": true,
    "aws": {
      "require_irsa": true,
      "allowed_account_ids": ["123456789012", "987654321098"],
      "allowed_regions": ["us-east-1", "us-west-2", "eu-west-1"]
    }
  },
  "kubernetes": {
    "denied_namespaces": ["default", "kube-system", "kube-public"]
  },
  "provider_map": {
    "local": "keyring",
    "eks": "aws-sm",
    "lambda": "aws-sm"
  },
  "locked": [
    "local.require_encryption",
    "cloud.require_iam",
    "cloud.aws.require_irsa",
    "cloud.aws.allowed_account_ids",
    "provider_map.eks",
    "provider_map.lambda"
  ]
}
```

### GCP-Only Enterprise

For organizations standardized on GCP:

```json
{
  "version": 1,
  "local": {
    "min_security_score": 50,
    "require_encryption": true
  },
  "cloud": {
    "require_iam": true,
    "gcp": {
      "require_workload_identity": true,
      "allowed_projects": ["prod-project-123", "staging-project-456"],
      "allowed_regions": ["us-central1", "us-east1", "europe-west1"]
    }
  },
  "kubernetes": {
    "denied_namespaces": ["default", "kube-system", "kube-public"]
  },
  "provider_map": {
    "local": "keyring",
    "gke": "gcp-sm",
    "cloudrun": "gcp-sm"
  },
  "locked": [
    "local.require_encryption",
    "cloud.require_iam",
    "cloud.gcp.require_workload_identity",
    "cloud.gcp.allowed_projects",
    "provider_map.gke",
    "provider_map.cloudrun"
  ]
}
```

### Multi-Cloud Enterprise

For organizations using multiple cloud providers:

```json
{
  "version": 1,
  "local": {
    "min_security_score": 60,
    "require_encryption": true,
    "require_tpm": true
  },
  "cloud": {
    "require_iam": true,
    "aws": {
      "require_irsa": true,
      "allowed_account_ids": ["111111111111", "222222222222"]
    },
    "gcp": {
      "require_workload_identity": true,
      "allowed_projects": ["company-prod", "company-staging"]
    },
    "azure": {
      "require_workload_identity": true,
      "allowed_tenant_ids": ["aaaaaaaa-bbbb-cccc-dddd-eeeeeeeeeeee"]
    }
  },
  "kubernetes": {
    "require_service_account": true,
    "denied_namespaces": ["default", "kube-system", "kube-public"]
  },
  "provider_map": {
    "local": "keyring",
    "eks": "aws-sm",
    "lambda": "aws-sm",
    "gke": "gcp-sm",
    "cloudrun": "gcp-sm",
    "aks": "azure-kv"
  },
  "allow_insecure": false,
  "locked": [
    "local.require_encryption",
    "cloud.require_iam",
    "allow_insecure",
    "provider_map.eks",
    "provider_map.gke",
    "provider_map.aks"
  ]
}
```

## Environment-Specific Configurations

### Production

Use with `AGENTPLEXUS_POLICY_FILE=/etc/agentplexus/production.json`:

```json
{
  "version": 1,
  "cloud": {
    "require_iam": true,
    "aws": {
      "require_irsa": true,
      "require_imdsv2": true,
      "allowed_account_ids": ["111111111111"],
      "allowed_regions": ["us-east-1", "us-west-2"]
    }
  },
  "kubernetes": {
    "require_service_account": true,
    "allowed_namespaces": ["production"],
    "require_non_root": true,
    "require_read_only_root": true
  },
  "provider_map": {
    "eks": "aws-sm"
  },
  "allow_insecure": false
}
```

### Staging

Use with `AGENTPLEXUS_POLICY_FILE=/etc/agentplexus/staging.json`:

```json
{
  "version": 1,
  "cloud": {
    "require_iam": true,
    "aws": {
      "require_irsa": true,
      "allowed_account_ids": ["222222222222"],
      "allowed_regions": ["us-east-1"]
    }
  },
  "kubernetes": {
    "require_service_account": true,
    "allowed_namespaces": ["staging"],
    "require_non_root": true
  },
  "provider_map": {
    "eks": "aws-sm"
  },
  "allow_insecure": false
}
```

### CI/CD Pipeline

Use with `AGENTPLEXUS_POLICY_FILE` set in CI environment:

```json
{
  "version": 1,
  "cloud": {
    "require_iam": true,
    "aws": {
      "allowed_role_arns": [
        "arn:aws:iam::333333333333:role/ci-runner-*"
      ]
    }
  },
  "kubernetes": {
    "allowed_namespaces": ["ci", "build"]
  },
  "provider_map": {
    "eks": "aws-sm",
    "container": "env"
  },
  "fallback_provider": "env"
}
```

## Installation Scripts

### Linux/macOS Enterprise Deployment

```bash
#!/bin/bash
set -e

# Create directory
sudo mkdir -p /etc/agentplexus

# Deploy policy
sudo tee /etc/agentplexus/policy.json > /dev/null << 'EOF'
{
  "version": 1,
  "local": {
    "min_security_score": 50,
    "require_encryption": true
  },
  "cloud": {
    "require_iam": true
  },
  "locked": ["local.require_encryption", "cloud.require_iam"]
}
EOF

# Set permissions
sudo chmod 644 /etc/agentplexus/policy.json

echo "Enterprise policy deployed successfully"
```

### Windows Enterprise Deployment

```powershell
# Create directory
$policyDir = "$env:ProgramData\agentplexus"
New-Item -ItemType Directory -Force -Path $policyDir | Out-Null

# Deploy policy
$policy = @"
{
  "version": 1,
  "local": {
    "min_security_score": 50,
    "require_encryption": true
  },
  "cloud": {
    "require_iam": true
  },
  "locked": ["local.require_encryption", "cloud.require_iam"]
}
"@

$policy | Out-File -FilePath "$policyDir\policy.json" -Encoding UTF8

Write-Host "Enterprise policy deployed successfully"
```

## Next Steps

- [Enterprise Policies](../policies/enterprise.md) - Understanding locked fields
- [File Locations](file-locations.md) - Platform-specific paths
- [JSON Schema](../reference/json-schema.md) - Complete field reference
