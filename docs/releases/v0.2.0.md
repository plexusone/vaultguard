# VaultGuard v0.2.0 Release Notes

**Release Date:** December 29, 2025

## Overview

VaultGuard v0.2.0 introduces **file-based policy configuration** with support for **enterprise policies**. Organizations can now deploy system-wide security policies with locked fields that users cannot override, enabling centralized security governance while still allowing individual customization within defined boundaries.

## New Features

### File-Based Policy Configuration

Policies can now be loaded from JSON configuration files, separating security policy from application code.

```go
// Load policy from configuration files
policy, err := vaultguard.LoadPolicy()
if err != nil {
    log.Fatal(err)
}

sv, err := vaultguard.New(&vaultguard.Config{
    Policy: policy,
})
```

### Configuration Hierarchy

Policies are loaded in order of precedence (highest first):

| Priority | Source | Path |
|----------|--------|------|
| 1 | Environment Variable | `AGENTPLEXUS_POLICY_FILE` |
| 2 | User Config | `~/.plexusone/policy.json` |
| 3 | System Config | `/etc/plexusone/policy.json` (Linux/macOS) |
| 3 | System Config | `%ProgramData%\plexusone\policy.json` (Windows) |

### Enterprise Policies with Locked Fields

System administrators can deploy organization-wide policies with fields that users cannot override:

```json
{
  "version": 1,
  "local": {
    "require_encryption": true,
    "min_security_score": 50
  },
  "cloud": {
    "require_iam": true
  },
  "allow_insecure": false,
  "locked": [
    "local.require_encryption",
    "cloud.require_iam",
    "allow_insecure"
  ]
}
```

When both system and user configs exist, they are merged with locked fields always using the enterprise value.

### Policy Merging

The new `mergeFilePolicies()` function intelligently combines enterprise and user policies:

- Non-locked fields: User values override enterprise values
- Locked fields: Enterprise values always win
- `allow_insecure`: Can only become more restrictive (true → false), never more permissive

### New API Functions

| Function | Description |
|----------|-------------|
| `LoadPolicy()` | Load merged policy from configuration hierarchy |
| `LoadPolicyFromFile(path)` | Load policy from a specific file |
| `SavePolicy(policy)` | Save policy to user config directory |
| `EnsureConfigDir()` | Create user config directory if needed |
| `GetConfigPaths()` | Get all configuration file paths |

### FilePolicy Type

New `FilePolicy` struct extends `Policy` with file-specific features:

```go
type FilePolicy struct {
    Version int      // Policy file format version
    Policy           // Embedded base policy
    Locked  []string // Field paths that cannot be overridden
}
```

## Documentation

### MkDocs Documentation Site

A comprehensive documentation site has been added using MkDocs with the Material theme:

```
docsrc/
├── index.md                    # Home page
├── getting-started.md          # Installation & basic usage
├── policies/
│   ├── overview.md             # How policies work
│   ├── local-policy.md         # Workstation security
│   ├── cloud-policy.md         # AWS/GCP/Azure requirements
│   ├── kubernetes-policy.md    # K8s-specific settings
│   └── enterprise.md           # Locked fields, deployment
├── configuration/
│   ├── file-locations.md       # Platform-specific paths
│   └── examples.md             # Ready-to-use JSON configs
└── reference/
    └── json-schema.md          # Complete field reference
```

Build the documentation:

```bash
pip install mkdocs-material
mkdocs build     # Output to docs/
mkdocs serve     # Local preview
```

### Updated README.md

- Added File-Based Configuration section
- Added Documentation section with links to MkDocs site

### Updated PRESENTATION.md

- Added 4 new slides covering file-based configuration and enterprise policies
- Updated architecture diagram to include `config.go`
- Updated Best Practices and Summary slides

## Lockable Field Paths

The following field paths can be locked by enterprise policies:

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

## Example Configurations

### User Configuration

`~/.plexusone/policy.json`:

```json
{
  "version": 1,
  "local": {
    "min_security_score": 60,
    "require_encryption": true
  },
  "provider_map": {
    "local": "keyring"
  }
}
```

### Enterprise Configuration

`/etc/plexusone/policy.json`:

```json
{
  "version": 1,
  "local": {
    "min_security_score": 50,
    "require_encryption": true,
    "require_tpm": true
  },
  "cloud": {
    "require_iam": true,
    "aws": {
      "require_irsa": true,
      "allowed_account_ids": ["123456789012"]
    }
  },
  "kubernetes": {
    "denied_namespaces": ["default", "kube-system", "kube-public"]
  },
  "allow_insecure": false,
  "locked": [
    "local.require_encryption",
    "local.require_tpm",
    "cloud.require_iam",
    "allow_insecure"
  ]
}
```

## Breaking Changes

None. This release is fully backwards compatible with v0.1.0.

## Migration Guide

No migration required. Existing code continues to work unchanged. To adopt file-based configuration:

1. Create `~/.plexusone/policy.json` (user) or `/etc/plexusone/policy.json` (enterprise)
2. Replace inline `Policy` with `LoadPolicy()`:

```go
// Before (v0.1.0)
sv, err := vaultguard.New(&vaultguard.Config{
    Policy: vaultguard.DefaultPolicy(),
})

// After (v0.2.0) - using file-based config
policy, _ := vaultguard.LoadPolicy()
sv, err := vaultguard.New(&vaultguard.Config{
    Policy: policy,
})
```

## Dependencies

- `github.com/plexusone/posture v0.2.0` - Security posture assessment
- `github.com/plexusone/omnivault v0.1.0` - Secret management abstraction

## What's Next

Planned for future releases:

- Policy validation CLI tool
- JSON Schema file for IDE autocompletion
- Policy dry-run mode
- Audit logging for policy violations
- HashiCorp Vault provider support

## Contributors

- AgentPlexus team

---

For issues and feature requests, please visit: https://github.com/plexusone/vaultguard/issues
