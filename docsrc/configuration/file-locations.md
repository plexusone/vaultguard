# File Locations

VaultGuard loads policy configuration from files in a specific order, allowing both enterprise-wide and user-specific settings.

## Configuration Hierarchy

Policies are loaded in this order (highest precedence first):

1. **Environment Variable** - `AGENTPLEXUS_POLICY_FILE`
2. **User Config** - `~/.agentplexus/policy.json`
3. **System Config** - Platform-specific (see below)
4. **No Config** - Permissive mode (no policy enforcement)

## Platform-Specific Paths

### Linux

| Type | Path |
|------|------|
| System (Enterprise) | `/etc/agentplexus/policy.json` |
| User | `~/.agentplexus/policy.json` |

### macOS

| Type | Path |
|------|------|
| System (Enterprise) | `/etc/agentplexus/policy.json` |
| User | `~/.agentplexus/policy.json` |

### Windows

| Type | Path |
|------|------|
| System (Enterprise) | `%ProgramData%\agentplexus\policy.json` |
| User | `%USERPROFILE%\.agentplexus\policy.json` |

Typical resolved paths on Windows:

- System: `C:\ProgramData\agentplexus\policy.json`
- User: `C:\Users\<username>\.agentplexus\policy.json`

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
/etc/agentplexus/
└── policy.json          # Enterprise policy

~/.agentplexus/
└── policy.json          # User policy

# Windows
C:\ProgramData\agentplexus\
└── policy.json          # Enterprise policy

C:\Users\<username>\.agentplexus\
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
// Output: Config directory: /home/user/.agentplexus
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

This creates `~/.agentplexus/policy.json` with mode `0600` (owner read/write only).

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
System: /etc/agentplexus/policy.json
User: /home/user/.agentplexus/policy.json
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
    sudo chmod 644 /etc/agentplexus/policy.json

    # User policy (readable/writable by owner only)
    chmod 600 ~/.agentplexus/policy.json
    chmod 700 ~/.agentplexus
    ```

=== "Windows (PowerShell)"

    ```powershell
    # User policy - restrict to current user
    $acl = Get-Acl "$env:USERPROFILE\.agentplexus\policy.json"
    $acl.SetAccessRuleProtection($true, $false)
    $rule = New-Object System.Security.AccessControl.FileSystemAccessRule(
        $env:USERNAME, "FullControl", "Allow")
    $acl.AddAccessRule($rule)
    Set-Acl "$env:USERPROFILE\.agentplexus\policy.json" $acl
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
ls -la /etc/agentplexus/policy.json
ls -la ~/.agentplexus/policy.json

# Check JSON validity
cat ~/.agentplexus/policy.json | jq .
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
sudo chown -R $USER:$USER ~/.agentplexus

# Fix permissions
chmod 700 ~/.agentplexus
chmod 600 ~/.agentplexus/policy.json
```

## Next Steps

- [Example Configs](examples.md) - Ready-to-use policy files
- [Enterprise Policies](../policies/enterprise.md) - Lock fields for organization
- [JSON Schema](../reference/json-schema.md) - Complete field reference
