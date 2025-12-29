# VaultGuard

[![Build Status][build-status-svg]][build-status-url]
[![Lint Status][lint-status-svg]][lint-status-url]
[![Go Report Card][goreport-svg]][goreport-url]
[![Docs][docs-godoc-svg]][docs-godoc-url]
[![License][license-svg]][license-url]

Security-gated credential access for Go applications. Combines [Posture](https://github.com/agentplexus/posture) (security posture assessment) with [OmniVault](https://github.com/agentplexus/omnivault) (secret management) to provide environment-aware secure credential handling.

## Features

- **Automatic Environment Detection** - Detects local workstations, AWS EKS, GCP GKE, Azure AKS, Lambda, Cloud Run, and more
- **Security Policy Enforcement** - Define security requirements that must be met before credentials are accessed
- **Provider Auto-Selection** - Automatically uses the right secret provider for each environment
- **Cross-Platform** - Works on macOS, Windows, Linux, and all major cloud platforms

## Quick Start

```go
package main

import (
    "context"
    "log"

    "github.com/agentplexus/vaultguard"
)

func main() {
    ctx := context.Background()

    // Create a secure vault with default settings
    sv, err := vaultguard.Quick()
    if err != nil {
        log.Fatalf("Security check failed: %v", err)
    }
    defer sv.Close()

    // Get credentials
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

## Security Policies

### Default Policy
```go
sv, _ := vaultguard.Quick() // Uses DefaultPolicy()
```

### Development Policy (Permissive)
```go
sv, _ := vaultguard.QuickDev() // Relaxed for development
```

### Strict Policy
```go
sv, _ := vaultguard.QuickStrict() // High security requirements
```

### Custom Policy
```go
sv, err := vaultguard.New(&vaultguard.Config{
    Policy: &vaultguard.Policy{
        Local: &vaultguard.LocalPolicy{
            MinSecurityScore:  75,
            RequireEncryption: true,
            RequireTPM:        true,
        },
        Cloud: &vaultguard.CloudPolicy{
            RequireIAM: true,
            AWS: &vaultguard.AWSPolicy{
                RequireIRSA: true,
                AllowedRoleARNs: []string{
                    "arn:aws:iam::123456789:role/my-app-*",
                },
            },
        },
        Kubernetes: &vaultguard.KubernetesPolicy{
            DeniedNamespaces: []string{"default", "kube-system"},
        },
    },
})
```

## Convenience Functions

```go
// Pre-flight security check
result, err := vaultguard.CheckSecurity(nil)
fmt.Printf("Security score: %d\n", result.Score)

// Require security (for init functions)
if err := vaultguard.RequireSecurity(nil); err != nil {
    log.Fatal(err)
}

// Quick credential access
apiKey, err := vaultguard.GetEnv(ctx, "API_KEY", nil)

// Load multiple credentials
creds, err := vaultguard.LoadCredentials(ctx, nil,
    "GOOGLE_API_KEY",
    "ANTHROPIC_API_KEY",
    "OPENAI_API_KEY",
)

// Load required credentials (error if any missing)
creds, err := vaultguard.LoadRequiredCredentials(ctx, nil,
    "GOOGLE_API_KEY",
    "SERPER_API_KEY",
)
```

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

## Integration with stats-agent-team

```go
package main

import (
    "context"
    "log"

    "github.com/agentplexus/vaultguard"
)

func main() {
    ctx := context.Background()

    // Security-gated credential loading
    sv, err := vaultguard.New(&vaultguard.Config{
        Policy: &vaultguard.Policy{
            Local: &vaultguard.LocalPolicy{
                MinSecurityScore:  50,
                RequireEncryption: true,
            },
            Cloud: &vaultguard.CloudPolicy{
                RequireIAM: true,
            },
        },
    })
    if err != nil {
        log.Fatalf("Security requirements not met: %v", err)
    }
    defer sv.Close()

    // Load agent credentials
    creds, err := sv.LoadRequiredCredentials(ctx, nil,
        "LLM_PROVIDER",
        "GOOGLE_API_KEY",
        "SERPER_API_KEY",
    )
    if err != nil {
        log.Fatal(err)
    }

    // Start agents with credentials...
    log.Printf("Starting agents with provider: %s", creds["LLM_PROVIDER"])
}
```

## Installation

```bash
go get github.com/agentplexus/vaultguard
```

## Dependencies

- [Posture](https://github.com/agentplexus/posture) - Security posture assessment
- [OmniVault](https://github.com/agentplexus/omnivault) - Secret management

## License

MIT License

 [build-status-svg]: https://github.com/agentplexus/vaultguard/actions/workflows/ci.yaml/badge.svg?branch=main
 [build-status-url]: https://github.com/agentplexus/vaultguard/actions/workflows/ci.yaml
 [lint-status-svg]: https://github.com/agentplexus/vaultguard/actions/workflows/lint.yaml/badge.svg?branch=main
 [lint-status-url]: https://github.com/agentplexus/vaultguard/actions/workflows/lint.yaml
 [goreport-svg]: https://goreportcard.com/badge/github.com/agentplexus/vaultguard
 [goreport-url]: https://goreportcard.com/report/github.com/agentplexus/vaultguard
 [docs-godoc-svg]: https://pkg.go.dev/badge/github.com/agentplexus/vaultguard
 [docs-godoc-url]: https://pkg.go.dev/github.com/agentplexus/vaultguard
 [license-svg]: https://img.shields.io/badge/license-MIT-blue.svg
 [license-url]: https://github.com/agentplexus/vaultguard/blob/master/LICENSE
 [used-by-svg]: https://sourcegraph.com/github.com/agentplexus/vaultguard/-/badge.svg
 [used-by-url]: https://sourcegraph.com/github.com/agentplexus/vaultguard?badge
