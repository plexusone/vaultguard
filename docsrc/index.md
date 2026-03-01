# VaultGuard

**Security-gated credential access for Go applications**

VaultGuard combines [Posture](https://github.com/plexusone/posture) (security posture assessment) with [OmniVault](https://github.com/plexusone/omnivault) (secret management) to provide environment-aware secure credential handling.

## Key Features

- **Automatic Environment Detection** - Detects local workstations, AWS EKS, GCP GKE, Azure AKS, Lambda, Cloud Run, and more
- **Security Policy Enforcement** - Define security requirements that must be met before credentials are accessed
- **Provider Auto-Selection** - Automatically uses the right secret provider for each environment
- **Enterprise Configuration** - System-wide policies with lockable fields for organizational control
- **Cross-Platform** - Works on macOS, Windows, Linux, and all major cloud platforms

## How It Works

```
┌─────────────────────────────────────────────────────────────────────┐
│                           VaultGuard                                │
│                                                                     │
│  1. Environment Detection                                           │
│     DetectEnvironment() → local | eks | gke | aks | lambda | ...    │
│                                                                     │
│  2. Security Checks (based on environment)                          │
│     ┌─────────────────────┐   ┌────────────────────────────────┐    │
│     │ Local (Posture)     │   │ Cloud                          │    │
│     │ • Secure Enclave    │   │ • IRSA/Workload Identity       │    │
│     │ • Disk Encryption   │   │ • Role/Account validation      │    │
│     │ • Secure Boot       │   │ • Namespace restrictions       │    │
│     └─────────────────────┘   └────────────────────────────────┘    │
│                                                                     │
│  3. Provider Auto-Selection                                         │
│     local → keyring | eks → aws-sm | gke → gcp-sm | ...             │
│                                                                     │
│  4. Credential Access (via OmniVault)                               │
│     sv.GetValue(ctx, "API_KEY") → secret value                      │
└─────────────────────────────────────────────────────────────────────┘
```

## Quick Example

```go
package main

import (
    "context"
    "log"

    "github.com/plexusone/vaultguard"
)

func main() {
    ctx := context.Background()

    // Create a secure vault with default settings
    sv, err := vaultguard.Quick()
    if err != nil {
        log.Fatalf("Security check failed: %v", err)
    }
    defer sv.Close()

    // Get credentials - only succeeds if security policy passes
    apiKey, err := sv.GetValue(ctx, "API_KEY")
    if err != nil {
        log.Fatal(err)
    }

    log.Printf("Environment: %s, Provider: %s", sv.Environment(), sv.Provider())
}
```

## Environment Support

| Environment | Detection | Security Checks | Default Provider |
|-------------|-----------|-----------------|------------------|
| Local (macOS/Windows/Linux) | Automatic | Posture (TPM, encryption, Secure Boot) | Keyring |
| AWS EKS | IRSA env vars | IRSA validation, role ARN checks | AWS Secrets Manager |
| AWS Lambda | Lambda env vars | IAM validation | AWS Secrets Manager |
| GCP GKE | Workload Identity | Service account validation | GCP Secret Manager |
| GCP Cloud Run | K_SERVICE env | IAM validation | GCP Secret Manager |
| Azure AKS | Workload Identity | Client/tenant validation | Azure Key Vault |
| Kubernetes | Service account mount | Namespace, SA validation | K8s Secrets |
| Container | Docker/cgroup detection | Basic checks | Environment vars |

## Installation

```bash
go get github.com/plexusone/vaultguard
```

## Next Steps

- [Getting Started](getting-started.md) - Installation and first steps
- [Policies Overview](policies/overview.md) - Learn how security policies work
- [Configuration](configuration/file-locations.md) - Set up user and enterprise policies
- [Example Configs](configuration/examples.md) - Copy-paste ready policy files
