# Enterprise Policies

Enterprise policies allow organizations to enforce security requirements across all users and applications. System administrators can define baseline policies with locked fields that users cannot override.

## Overview

VaultGuard supports a two-tier configuration system:

| Layer | Location | Purpose |
|-------|----------|---------|
| **System/Enterprise** | `/etc/plexusone/policy.json` | Organization-wide defaults and enforced settings |
| **User** | `~/.plexusone/policy.json` | Personal preferences (within enterprise constraints) |

When both exist, they are merged with enterprise settings taking precedence on locked fields.

## File Policy Structure

Enterprise policies use the `FilePolicy` format, which extends the base `Policy`:

```go
type FilePolicy struct {
    Version int      // Policy file format version
    Policy           // Embedded base policy
    Locked  []string // Field paths that cannot be overridden
}
```

```json
{
  "version": 1,
  "local": { ... },
  "cloud": { ... },
  "kubernetes": { ... },
  "provider_map": { ... },
  "locked": [
    "local.require_encryption",
    "cloud.require_iam"
  ]
}
```

## Locked Fields

The `locked` array specifies field paths that users cannot override. When a field is locked, the enterprise value is always used regardless of user settings.

### Field Path Format

Field paths use dot notation:

| Path | Field |
|------|-------|
| `local.require_encryption` | LocalPolicy.RequireEncryption |
| `local.min_security_score` | LocalPolicy.MinSecurityScore |
| `local.require_tpm` | LocalPolicy.RequireTPM |
| `local.require_secure_boot` | LocalPolicy.RequireSecureBoot |
| `local.require_biometrics` | LocalPolicy.RequireBiometrics |
| `local.allowed_platforms` | LocalPolicy.AllowedPlatforms |
| `cloud.require_iam` | CloudPolicy.RequireIAM |
| `provider_map.{env}` | ProviderMap for specific environment |
| `fallback_provider` | FallbackProvider |
| `allow_insecure` | AllowInsecure (special handling) |

### Example: Locking Encryption Requirement

```json
{
  "version": 1,
  "local": {
    "require_encryption": true
  },
  "locked": ["local.require_encryption"]
}
```

Users cannot disable encryption requirement, even with:

```json
{
  "local": {
    "require_encryption": false
  }
}
```

## Merge Behavior

### Non-Locked Fields

Users can customize non-locked fields. Values are merged with user settings taking effect:

**Enterprise:**
```json
{
  "local": {
    "min_security_score": 50
  }
}
```

**User:**
```json
{
  "local": {
    "min_security_score": 75
  }
}
```

**Result:** `min_security_score: 75`

### Locked Fields

Locked fields always use the enterprise value:

**Enterprise:**
```json
{
  "local": {
    "min_security_score": 50
  },
  "locked": ["local.min_security_score"]
}
```

**User:**
```json
{
  "local": {
    "min_security_score": 25
  }
}
```

**Result:** `min_security_score: 50`

### AllowInsecure Special Handling

The `allow_insecure` field has special merge behavior: it can only become *more restrictive*, never more permissive.

| Enterprise | User | Result |
|------------|------|--------|
| `true` | `true` | `true` |
| `true` | `false` | `false` |
| `false` | `true` | `false` |
| `false` | `false` | `false` |

This prevents users from bypassing security checks.

## Deployment Scenarios

### Scenario 1: Developer Laptops

Enforce disk encryption while letting developers customize other settings:

```json
{
  "version": 1,
  "local": {
    "min_security_score": 50,
    "require_encryption": true
  },
  "locked": ["local.require_encryption"],
  "allow_insecure": false
}
```

Developers can adjust `min_security_score` for their workflow but cannot disable encryption.

### Scenario 2: Production Servers

Lock down all security settings:

```json
{
  "version": 1,
  "local": {
    "min_security_score": 75,
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
    "local.min_security_score",
    "local.require_encryption",
    "local.require_tpm",
    "cloud.require_iam",
    "allow_insecure"
  ]
}
```

### Scenario 3: Multi-Cloud with Provider Lock

Ensure consistent secret providers across environments:

```json
{
  "version": 1,
  "cloud": {
    "require_iam": true
  },
  "provider_map": {
    "local": "keyring",
    "eks": "aws-sm",
    "gke": "gcp-sm",
    "aks": "azure-kv"
  },
  "locked": [
    "provider_map.eks",
    "provider_map.gke",
    "provider_map.aks"
  ]
}
```

Users can change the local provider but not cloud providers.

## Installation

### Linux/macOS

```bash
# Create directory
sudo mkdir -p /etc/plexusone

# Create policy file
sudo cat > /etc/plexusone/policy.json << 'EOF'
{
  "version": 1,
  "local": {
    "require_encryption": true,
    "min_security_score": 50
  },
  "cloud": {
    "require_iam": true
  },
  "locked": ["local.require_encryption", "cloud.require_iam"]
}
EOF

# Set permissions
sudo chmod 644 /etc/plexusone/policy.json
```

### Windows

```powershell
# Create directory
New-Item -ItemType Directory -Force -Path "$env:ProgramData\plexusone"

# Create policy file
@"
{
  "version": 1,
  "local": {
    "require_encryption": true,
    "min_security_score": 50
  },
  "cloud": {
    "require_iam": true
  },
  "locked": ["local.require_encryption", "cloud.require_iam"]
}
"@ | Out-File -FilePath "$env:ProgramData\plexusone\policy.json" -Encoding UTF8
```

### Configuration Management

Deploy via your preferred configuration management tool:

**Ansible:**
```yaml
- name: Deploy VaultGuard enterprise policy
  copy:
    src: files/vaultguard-policy.json
    dest: /etc/plexusone/policy.json
    mode: '0644'
```

**Puppet:**
```puppet
file { '/etc/plexusone/policy.json':
  ensure  => file,
  source  => 'puppet:///modules/vaultguard/policy.json',
  mode    => '0644',
}
```

**Chef:**
```ruby
cookbook_file '/etc/plexusone/policy.json' do
  source 'policy.json'
  mode '0644'
end
```

## Verifying Configuration

Check which configuration files are active:

```go
paths := vaultguard.GetConfigPaths()
fmt.Printf("System config: %s\n", paths["system"])
fmt.Printf("User config: %s\n", paths["user"])
fmt.Printf("Env override: %s\n", paths["env"])
```

Load and inspect the merged policy:

```go
policy, err := vaultguard.LoadPolicy()
if err != nil {
    log.Fatal(err)
}

// policy now contains the merged result
fmt.Printf("Encryption required: %v\n", policy.Local.RequireEncryption)
```

## Best Practices

1. **Start minimal** - Lock only critical settings initially
2. **Document locked fields** - Communicate to users what they can/cannot change
3. **Test merging** - Verify user configs merge as expected
4. **Version control** - Keep enterprise policies in version control
5. **Audit regularly** - Review locked settings periodically

## Next Steps

- [File Locations](../configuration/file-locations.md) - Platform-specific paths
- [Example Configs](../configuration/examples.md) - Ready-to-use enterprise policies
- [JSON Schema](../reference/json-schema.md) - Complete field reference
