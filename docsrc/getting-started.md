# Getting Started

This guide covers installing VaultGuard and your first secure credential access.

## Installation

```bash
go get github.com/agentplexus/vaultguard
```

## Basic Usage

### Quick Start (Recommended)

The simplest way to use VaultGuard is with the `Quick()` function, which uses sensible defaults:

```go
package main

import (
    "context"
    "log"

    "github.com/agentplexus/vaultguard"
)

func main() {
    ctx := context.Background()

    // Create vault with default policy
    sv, err := vaultguard.Quick()
    if err != nil {
        log.Fatalf("Security check failed: %v", err)
    }
    defer sv.Close()

    // Access credentials
    apiKey, err := sv.GetValue(ctx, "API_KEY")
    if err != nil {
        log.Fatal(err)
    }

    log.Printf("Got API key from %s provider", sv.Provider())
}
```

### Development Mode

For local development where you want relaxed security checks:

```go
sv, err := vaultguard.QuickDev()
if err != nil {
    log.Fatal(err)
}
defer sv.Close()
```

!!! warning
    Never use `QuickDev()` in production. It sets `AllowInsecure: true` and relaxes all security requirements.

### Strict Mode

For high-security environments:

```go
sv, err := vaultguard.QuickStrict()
if err != nil {
    log.Fatal(err)
}
defer sv.Close()
```

## Loading Multiple Credentials

### Load All Available

```go
creds, err := sv.LoadCredentials(ctx, nil,
    "GOOGLE_API_KEY",
    "ANTHROPIC_API_KEY",
    "OPENAI_API_KEY",
)
// creds contains only the keys that were found
```

### Load Required (Fail if Missing)

```go
creds, err := sv.LoadRequiredCredentials(ctx, nil,
    "GOOGLE_API_KEY",
    "SERPER_API_KEY",
)
if err != nil {
    log.Fatal(err) // Fails if any key is missing
}
```

## Convenience Functions

### Pre-flight Security Check

Check security posture before starting your application:

```go
result, err := vaultguard.CheckSecurity(nil)
if err != nil {
    log.Fatal(err)
}

fmt.Printf("Security score: %d/100\n", result.Score)
fmt.Printf("Environment: %s\n", result.Environment)
fmt.Printf("Level: %s\n", result.Level)

for _, rec := range result.Recommendations {
    fmt.Printf("  - %s\n", rec)
}
```

### Require Security (for init)

Useful in `init()` functions to fail fast:

```go
func init() {
    if err := vaultguard.RequireSecurity(nil); err != nil {
        log.Fatalf("Security requirements not met: %v", err)
    }
}
```

### One-off Credential Access

When you just need a single value without keeping the vault open:

```go
apiKey, err := vaultguard.GetEnv(ctx, "API_KEY", nil)
if err != nil {
    log.Fatal(err)
}
```

## Custom Policy

For fine-grained control over security requirements:

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

## Using Configuration Files

VaultGuard can load policies from configuration files, allowing separation of policy from code:

```go
// Loads policy from:
// 1. AGENTPLEXUS_POLICY_FILE env var (if set)
// 2. ~/.agentplexus/policy.json (user config)
// 3. /etc/agentplexus/policy.json (system config)
policy, err := vaultguard.LoadPolicy()
if err != nil {
    log.Fatal(err)
}

sv, err := vaultguard.New(&vaultguard.Config{
    Policy: policy,
})
```

See [Configuration](configuration/file-locations.md) for details on policy file locations and [Example Configs](configuration/examples.md) for ready-to-use policy files.

## Next Steps

- [Policies Overview](policies/overview.md) - Understand how policies work
- [Local Policy](policies/local-policy.md) - Configure workstation security
- [Cloud Policy](policies/cloud-policy.md) - Configure cloud environment security
- [Enterprise Policies](policies/enterprise.md) - Set up organization-wide policies
