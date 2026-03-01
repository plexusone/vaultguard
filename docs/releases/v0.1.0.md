# VaultGuard v0.1.0 Release Notes

**Release Date:** December 28, 2025

## Overview

VaultGuard v0.1.0 is the initial release of the security-gated credential access library for Go. VaultGuard combines security posture assessment with secret management to ensure credentials are only accessible when adequate security requirements are met.

## Features

### Core Functionality

- **Security-Gated Access**: Credentials are only released after security posture verification passes configured policy requirements
- **Automatic Environment Detection**: Automatically detects deployment environment and selects appropriate secret provider
- **Flexible Policy System**: Three preset policies (Default, Development, Strict) plus full custom policy support

### Environment Support

| Environment | Detection | Secret Provider |
|-------------|-----------|-----------------|
| Local (macOS/Windows/Linux) | Runtime OS detection | System Keyring |
| AWS EKS | IRSA environment variables | AWS Secrets Manager |
| AWS Lambda | Lambda environment variables | AWS Secrets Manager |
| GCP GKE | Workload Identity token | GCP Secret Manager |
| GCP Cloud Run | K_SERVICE environment variable | GCP Secret Manager |
| Azure AKS | Azure identity environment variables | Azure Key Vault |
| Generic Kubernetes | Service account token | Kubernetes Secrets |
| Container | Docker/cgroup detection | Environment Variables |

### Security Checks

**Local Environments:**
- Disk encryption status
- TPM presence and health
- Secure Boot enabled
- Biometric authentication availability
- Platform security score (0-100)

**Cloud Environments:**
- IAM role configuration validation
- IRSA / Workload Identity verification
- Role ARN pattern matching
- Account/Project/Subscription restrictions
- Region restrictions

**Kubernetes Environments:**
- Namespace allow/deny lists
- Service account requirements
- Pod security context validation (non-root, read-only filesystem)

### Policy Presets

- **DefaultPolicy()**: Production-ready with minimum 50/100 security score, disk encryption required, IAM required for cloud
- **DevelopmentPolicy()**: Permissive mode for local development and testing
- **StrictPolicy()**: High-security mode requiring 75/100 score, TPM, Secure Boot, and all IAM features

### Convenience Functions

- `Quick()` / `QuickDev()` / `QuickStrict()`: One-liner vault initialization
- `CheckSecurity()`: Pre-flight security assessment without credential access
- `RequireSecurity()`: Panic-based security check for use in `init()` functions
- `GetEnv()`: Security-gated environment variable access
- `LoadCredentials()` / `LoadRequiredCredentials()`: Batch credential loading

### Error Handling

- `SecurityError`: Detailed security check failures with recommendations
- `PolicyError`: Policy violation details
- `ProviderError`: Provider-specific error information
- Standard sentinel errors: `ErrSecurityCheckFailed`, `ErrPolicyViolation`, `ErrSecretNotFound`, etc.

## Installation

```bash
go get github.com/plexusone/vaultguard
```

## Quick Start

```go
package main

import (
    "context"
    "log"

    "github.com/plexusone/vaultguard"
)

func main() {
    ctx := context.Background()

    // Create security-gated vault with default policy
    sv, err := vaultguard.Quick()
    if err != nil {
        log.Fatal("Security requirements not met:", err)
    }
    defer sv.Close()

    // Access credentials (only after security verification)
    apiKey, err := sv.GetValue(ctx, "API_KEY")
    if err != nil {
        log.Fatal("Failed to get credential:", err)
    }

    // Use apiKey safely...
}
```

## Dependencies

- `github.com/plexusone/posture v0.2.0` - Security posture assessment
- `github.com/plexusone/omnivault v0.1.0` - Secret management abstraction

## Examples

The release includes example implementations:

- **agent-credentials**: Full policy configuration for AI agent credential management
- **stats-agent-team**: Real-world integration with Helm/YAML configuration

## Documentation

- `README.md`: Quick start guide and API overview
- `PRESENTATION.md`: Marp-compatible presentation explaining the project
- `docs/index.html`: API documentation

## Known Limitations

- Cloud provider secret backends require respective SDK authentication to be configured
- Local security assessment requires the Posture library which has platform-specific implementations
- Kubernetes namespace detection requires the downward API or standard service account mounts

## What's Next

Planned for future releases:
- HashiCorp Vault provider support
- Secret caching with configurable TTL
- Audit logging integration
- Additional cloud provider support (DigitalOcean, Alibaba Cloud)
- CLI tool for security assessment

## Contributors

- Initial development by AgentPlexus team

---

For issues and feature requests, please visit: https://github.com/plexusone/vaultguard/issues
