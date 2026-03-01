# File Locations

VaultGuard loads policy configuration from files in a specific order, allowing both enterprise-wide and user-specific settings.

## Configuration Hierarchy

Policies are loaded in this order (highest precedence first):

1. **Environment Variable** - `AGENTPLEXUS_POLICY_FILE`
2. **User Config** - `~/.plexusone/policy.json`
3. **System Config** - Platform-specific (see below)
4. **No Config** - Permissive mode (no policy enforcement)

## Platform-Specific Paths

### Linux

| Type | Path |
|------|------|
| System (Enterprise) | `/etc/plexusone/policy.json` |
| User | `~/.plexusone/policy.json` |

### macOS

| Type | Path |
|------|------|
| System (Enterprise) | `/etc/plexusone/policy.json` |
| User | `~/.plexusone/policy.json` |

### Windows

| Type | Path |
|------|------|
| System (Enterprise) | `%ProgramData%\plexusone\policy.json` |
| User | `%USERPROFILE%\.plexusone\policy.json` |

Typical resolved paths on Windows:

- System: `C:\ProgramData\plexusone\policy.json`
- User: `C:\Users\<username>\.plexusone\policy.json`

## Environment Variable Override

Set `AGENTPLEXUS_POLICY_FILE` to use a specific policy file:

=== "Linux/macOS"

    ```bash
    export AGENTPLEXUS_POLICY_FILE=/path/to/custom-policy.json
    ```

=== "Windows (PowerShell)"

    ```powershell
    $env:AGENTPLEXUS_POLICY_FILE = "C:\path\to\custom-policy.json"
    ```

=== "Windows (CMD)"

    ```cmd
    set AGENTPLEXUS_POLICY_FILE=C:\path\to\custom-policy.json
    ```

When set, this overrides both system and user configuration files.

## Directory Structure

```
# Linux/macOS
/etc/plexusone/
└── policy.json          # Enterprise policy

~/.plexusone/
└── policy.json          # User policy

# Windows
C:\ProgramData\plexusone\
└── policy.json          # Enterprise policy

C:\Users\<username>\.plexusone\
└── policy.json          # User policy
```

## Creating the User Config Directory

VaultGuard provides a helper function to create the user config directory:

```go
configDir, err := vaultguard.EnsureConfigDir()
if err != nil {
    log.Fatal(err)
}
fmt.Printf("Config directory: %s\n", configDir)
// Output: Config directory: /home/user/.plexusone
```

The directory is created with mode `0700` (owner read/write/execute only).

## Saving a Policy

Save a policy to the user config directory:

```go
policy := &vaultguard.Policy{
    Local: &vaultguard.LocalPolicy{
        MinSecurityScore:  60,
        RequireEncryption: true,
    },
}

if err := vaultguard.SavePolicy(policy); err != nil {
    log.Fatal(err)
}
```

This creates `~/.plexusone/policy.json` with mode `0600` (owner read/write only).

## Checking Active Paths

Get the paths VaultGuard will check for configuration:

```go
paths := vaultguard.GetConfigPaths()
fmt.Printf("System: %s\n", paths["system"])
fmt.Printf("User: %s\n", paths["user"])
fmt.Printf("Env: %s\n", paths["env"])
```

Example output on Linux:

```
System: /etc/plexusone/policy.json
User: /home/user/.plexusone/policy.json
Env:
```

## Loading Policy

Load the merged policy from all sources:

```go
policy, err := vaultguard.LoadPolicy()
if err != nil {
    log.Fatal(err)
}

if policy == nil {
    fmt.Println("No policy files found - running in permissive mode")
} else {
    fmt.Printf("Policy loaded successfully\n")
}
```

Load from a specific file:

```go
filePolicy, err := vaultguard.LoadPolicyFromFile("/path/to/policy.json")
if err != nil {
    log.Fatal(err)
}

// Access the embedded Policy
policy := &filePolicy.Policy

// Check locked fields
fmt.Printf("Locked fields: %v\n", filePolicy.Locked)
```

## File Permissions

### Recommended Permissions

| File | Linux/macOS | Windows |
|------|-------------|---------|
| System policy | `644` (rw-r--r--) | Administrators: Full Control |
| User policy | `600` (rw-------) | User: Full Control |
| User directory | `700` (rwx------) | User: Full Control |

### Setting Permissions

=== "Linux/macOS"

    ```bash
    # System policy (readable by all, writable by root)
    sudo chmod 644 /etc/plexusone/policy.json

    # User policy (readable/writable by owner only)
    chmod 600 ~/.plexusone/policy.json
    chmod 700 ~/.plexusone
    ```

=== "Windows (PowerShell)"

    ```powershell
    # User policy - restrict to current user
    $acl = Get-Acl "$env:USERPROFILE\.plexusone\policy.json"
    $acl.SetAccessRuleProtection($true, $false)
    $rule = New-Object System.Security.AccessControl.FileSystemAccessRule(
        $env:USERNAME, "FullControl", "Allow")
    $acl.AddAccessRule($rule)
    Set-Acl "$env:USERPROFILE\.plexusone\policy.json" $acl
    ```

## Merge Behavior

When both system and user configs exist:

1. System config provides the base settings
2. User config values override system values
3. **Except** for fields listed in `locked` array
4. `allow_insecure` can only become more restrictive

See [Enterprise Policies](../policies/enterprise.md) for details on merge behavior.

## Troubleshooting

### Policy Not Loading

Check if files exist and are readable:

```bash
# Linux/macOS
ls -la /etc/plexusone/policy.json
ls -la ~/.plexusone/policy.json

# Check JSON validity
cat ~/.plexusone/policy.json | jq .
```

### Wrong Policy Applied

Check the precedence:

```go
paths := vaultguard.GetConfigPaths()
if paths["env"] != "" {
    fmt.Println("Using env override:", paths["env"])
}
```

### Permission Denied

Ensure proper ownership and permissions:

```bash
# Fix user directory ownership
sudo chown -R $USER:$USER ~/.plexusone

# Fix permissions
chmod 700 ~/.plexusone
chmod 600 ~/.plexusone/policy.json
```

## Next Steps

- [Example Configs](examples.md) - Ready-to-use policy files
- [Enterprise Policies](../policies/enterprise.md) - Lock fields for organization
- [JSON Schema](../reference/json-schema.md) - Complete field reference
