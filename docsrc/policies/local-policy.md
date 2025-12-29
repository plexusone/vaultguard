# Local Policy

Local policies define security requirements for developer workstations running macOS, Windows, or Linux. These checks are performed using [Posture](https://github.com/agentplexus/posture), which assesses the security configuration of the local machine.

## Overview

Local policies are evaluated when `Environment.IsLocal()` returns true, typically on developer machines outside of cloud/container environments.

```go
type LocalPolicy struct {
    MinSecurityScore  int      // Minimum Posture score (0-100)
    RequireEncryption bool     // Require disk encryption
    RequireTPM        bool     // Require TPM/Secure Enclave
    RequireSecureBoot bool     // Require Secure Boot
    RequireBiometrics bool     // Require biometric auth configured
    AllowedPlatforms  []string // Restrict to specific platforms
}
```

## Fields

### MinSecurityScore

The minimum [Posture](https://github.com/agentplexus/posture) security score required (0-100).

| Score Range | Security Level | Typical Configuration |
|-------------|----------------|----------------------|
| 0-24 | Critical | Missing basic protections |
| 25-49 | Low | Some protections, gaps exist |
| 50-74 | Medium | Good baseline security |
| 75-89 | High | Strong security posture |
| 90-100 | Excellent | All security features enabled |

```json
{
  "local": {
    "min_security_score": 50
  }
}
```

!!! tip
    Start with a score of 50 for production and increase gradually. Use `vaultguard.CheckSecurity()` to see current scores across your team.

### RequireEncryption

Requires full-disk encryption to be enabled:

- **macOS**: FileVault
- **Windows**: BitLocker
- **Linux**: LUKS

```json
{
  "local": {
    "require_encryption": true
  }
}
```

!!! warning
    This is one of the most important security requirements. Unencrypted disks expose all credentials if a device is lost or stolen.

### RequireTPM

Requires a Trusted Platform Module or equivalent:

- **macOS**: Secure Enclave (T2/Apple Silicon)
- **Windows**: TPM 2.0
- **Linux**: TPM 2.0

```json
{
  "local": {
    "require_tpm": true
  }
}
```

TPM/Secure Enclave provides:

- Hardware-backed key storage
- Measured boot verification
- Tamper resistance

### RequireSecureBoot

Requires Secure Boot to be enabled:

```json
{
  "local": {
    "require_secure_boot": true
  }
}
```

Secure Boot ensures:

- Only signed bootloaders can run
- Protection against boot-level malware
- Chain of trust from firmware to OS

!!! note
    macOS with Apple Silicon always has Secure Boot enabled. On Intel Macs with T2 chip, it's configurable.

### RequireBiometrics

Requires biometric authentication to be configured:

- **macOS**: Touch ID
- **Windows**: Windows Hello (fingerprint/face)
- **Linux**: Fingerprint (if supported)

```json
{
  "local": {
    "require_biometrics": true
  }
}
```

### AllowedPlatforms

Restricts credential access to specific operating systems:

```json
{
  "local": {
    "allowed_platforms": ["darwin", "linux"]
  }
}
```

Valid values: `darwin`, `windows`, `linux`

An empty list allows all platforms.

## Examples

### Minimal Security (Development)

```json
{
  "local": {
    "min_security_score": 0
  }
}
```

### Standard Production

```json
{
  "local": {
    "min_security_score": 50,
    "require_encryption": true
  }
}
```

### High Security

```json
{
  "local": {
    "min_security_score": 75,
    "require_encryption": true,
    "require_tpm": true,
    "require_secure_boot": true
  }
}
```

### macOS-Only with Biometrics

```json
{
  "local": {
    "min_security_score": 60,
    "require_encryption": true,
    "require_tpm": true,
    "require_biometrics": true,
    "allowed_platforms": ["darwin"]
  }
}
```

## Checking Local Security

You can check the current security posture programmatically:

```go
result, err := vaultguard.CheckSecurity(nil)
if err != nil {
    log.Fatal(err)
}

fmt.Printf("Score: %d\n", result.Score)
fmt.Printf("Platform: %s\n", result.Details.Local.Platform)
fmt.Printf("Disk Encrypted: %v\n", result.Details.Local.DiskEncrypted)
fmt.Printf("TPM Present: %v\n", result.Details.Local.TPMPresent)
fmt.Printf("Secure Boot: %v\n", result.Details.Local.SecureBootEnabled)

for _, rec := range result.Recommendations {
    fmt.Printf("Recommendation: %s\n", rec)
}
```

## Default Provider

When running locally, VaultGuard defaults to the `keyring` provider:

- **macOS**: Keychain
- **Windows**: Credential Manager
- **Linux**: Secret Service (GNOME Keyring, KWallet)

This can be overridden in the policy's `ProviderMap`:

```json
{
  "local": {
    "require_encryption": true
  },
  "provider_map": {
    "local": "env"
  }
}
```

## Next Steps

- [Cloud Policy](cloud-policy.md) - AWS, GCP, Azure requirements
- [Enterprise Policies](enterprise.md) - Lock settings organization-wide
- [Example Configs](../configuration/examples.md) - Ready-to-use policy files
