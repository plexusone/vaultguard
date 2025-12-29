---
marp: true
theme: agentplexus
paginate: true
---

# VaultGuard 🛡️

## Security-Gated Credential Access for Agents & Cloud-Native Apps

**🔐 Protect your secrets. Verify before access.**

---

# The Problem ⚠️

## Credentials Are Everywhere

- 🔑 API keys in environment variables
- 📄 Secrets in config files
- 🔄 Tokens passed between services
- 🤖 Agent systems accessing multiple providers

> **But how do you know your environment is secure enough to handle them?**

---

# The Threat Landscape 🚨

## Real-World Credential Exposure Risks

| Scenario | Risk |
|----------|------|
| 💻 Unencrypted laptop | Credentials stolen if device lost |
| 📦 Container without IAM | Exposed metadata service attacks |
| 🤖 Agent on shared infra | Cross-tenant credential leakage |
| 🔌 MCP server on insecure host | Tool credentials compromised |

**Traditional secret managers don't verify the security of the requesting environment.**

---

# Introducing VaultGuard 🛡️

## Security Posture + Secret Management

```
┌───────────────────────────────────────────┐
│          Your Application / Agent         │
└─────────────────────┬─────────────────────┘
                      │
                      ▼
┌───────────────────────────────────────────┐
│              VaultGuard                   │
│  ┌─────────────┐    ┌──────────────────┐  │
│  │  Security   │ +  │  Secret Access   │  │
│  │  Assessment │    │  (OmniVault)     │  │
│  └─────────────┘    └──────────────────┘  │
└───────────────────────────────────────────┘
```

**✅ Credentials only released when security requirements are met.**

---

# How It Works ⚙️

## Three-Step Security Gate

```go
// 1️⃣ Detect environment automatically
Environment: Local | AWS EKS | GCP GKE | Azure AKS | Lambda | Cloud Run

// 2️⃣ Assess security posture
Score: 75/100 | TPM: ✓ | Encryption: ✓ | Secure Boot: ✓

// 3️⃣ Enforce policy before credential access
if score >= policy.MinScore && requirements.Met() {
    return credentials  // ✅ Access granted
}
return SecurityError   // ❌ Access denied with recommendations
```

---

# Security Checks ✅

## What VaultGuard Verifies

<div class="columns">
<div>

### 💻 Local Environments
- 🔒 Disk encryption status
- 🔐 TPM presence & health
- 🛡️ Secure Boot enabled
- 👆 Biometric authentication
- 📊 Platform security score

</div>
<div>

### ☁️ Cloud Environments
- 🎭 IAM role configuration
- 🔑 IRSA / Workload Identity
- 📁 Namespace restrictions
- 🛡️ Pod security context
- 👤 Service account validation

</div>
</div>

---

# Policy System 📋

## Flexible Security Requirements

```go
// 🏭 Production - Strict security requirements
vaultguard.DefaultPolicy()

// 🛠️ Development - Permissive for local testing
vaultguard.DevelopmentPolicy()

// 🔒 High Security - Maximum protection
vaultguard.StrictPolicy()

// ⚙️ Custom - Your specific requirements
&vaultguard.Policy{
    Local: &LocalPolicy{MinSecurityScore: 60, RequireEncryption: true},
    Cloud: &CloudPolicy{RequireIAM: true},
}
```

---

# File-Based Configuration 📄

## Separate Policy from Code

```
Configuration Hierarchy (highest precedence first):

1. 🔧 AGENTPLEXUS_POLICY_FILE environment variable
2. 👤 User config:   ~/.agentplexus/policy.json
3. 🏢 System config: /etc/agentplexus/policy.json
```

```go
// Load policy from configuration files
policy, err := vaultguard.LoadPolicy()

sv, err := vaultguard.New(&vaultguard.Config{
    Policy: policy,  // 📋 From file, not hardcoded
})
```

---

# User Configuration 👤

## Personal Preferences

`~/.agentplexus/policy.json`:

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

**✅ Users can customize their security settings within enterprise constraints.**

---

# Enterprise Policies 🏢

## Organization-Wide Enforcement

`/etc/agentplexus/policy.json`:

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
  "locked": [
    "local.require_encryption",
    "cloud.require_iam",
    "allow_insecure"
  ]
}
```

**🔒 Locked fields cannot be overridden by user configuration.**

---

# Policy Merging 🔀

## Enterprise + User = Final Policy

| Field | Enterprise | User | Result |
|-------|-----------|------|--------|
| `require_encryption` | `true` 🔒 | `false` | `true` ✅ |
| `min_security_score` | `50` | `75` | `75` ✅ |
| `allow_insecure` | `false` 🔒 | `true` | `false` ✅ |

> **🔒 = Locked field** - Enterprise value always wins

```go
// User tried to disable encryption, but it's locked
// Result: encryption still required ✅
```

---

# Environment Auto-Detection 🔍

## Runs Anywhere, Adapts Automatically

| Environment | Detection Method | Secret Provider |
|-------------|-----------------|-----------------|
| 💻 **macOS/Windows/Linux** | Runtime OS detection | System Keyring |
| 🟠 **AWS EKS** | IRSA env vars | AWS Secrets Manager |
| 🟠 **AWS Lambda** | Lambda env vars | AWS Secrets Manager |
| 🔵 **GCP GKE** | Workload Identity token | GCP Secret Manager |
| 🔵 **GCP Cloud Run** | K_SERVICE env var | GCP Secret Manager |
| 🟣 **Azure AKS** | Azure identity env vars | Azure Key Vault |
| ☸️ **Kubernetes** | Service account token | K8s Secrets |

---

# Why This Matters for AI Agents 🤖

## Agents Have Unique Security Challenges

- 🔑 **Multiple API Keys**: LLM providers, search APIs, databases
- 🔄 **Autonomous Operation**: Run without human oversight
- 🌍 **Diverse Environments**: Local dev, cloud, edge deployment
- ⚡ **High Privilege**: Often need access to sensitive systems

> An agent with leaked credentials can cause **significant damage** before detection. 💥

---

# Agent Credential Security 🔐

## The VaultGuard Approach

```go
func initAgent() (*Agent, error) {
    // 🛡️ Security-gated credential access
    sv, err := vaultguard.Quick()
    if err != nil {
        // ❌ Environment doesn't meet security requirements
        return nil, fmt.Errorf("insecure environment: %w", err)
    }
    defer sv.Close()

    // ✅ Only accessible after security verification
    openaiKey, _ := sv.GetValue(ctx, "OPENAI_API_KEY")
    anthropicKey, _ := sv.GetValue(ctx, "ANTHROPIC_API_KEY")

    return NewAgent(openaiKey, anthropicKey), nil
}
```

---

# MCP Server Security 🔌

## Model Context Protocol Servers Need Protection

MCP servers expose tools to AI models:
- 📁 File system access
- 🗄️ Database connections
- 🌐 API integrations
- ⚡ System commands

**🔐 These tools require credentials that must be protected.**

---

# Securing MCP Servers 🔒

## VaultGuard Integration Pattern

```go
func NewSecureMCPServer() (*MCPServer, error) {
    // 🔍 Verify security before starting server
    result, err := vaultguard.CheckSecurity(nil)
    if err != nil || !result.Passed {
        return nil, fmt.Errorf("MCP server requires secure environment")
    }

    sv, err := vaultguard.Quick()
    if err != nil {
        return nil, err
    }

    // 🔑 Load tool credentials securely
    dbCreds, _ := sv.GetValue(ctx, "DATABASE_URL")
    apiKey, _ := sv.GetValue(ctx, "EXTERNAL_API_KEY")

    return &MCPServer{vault: sv, db: dbCreds, api: apiKey}, nil
}
```

---

# Multi-Provider Agent Example 🌐

## Real-World Configuration

```go
sv, err := vaultguard.New(&vaultguard.Config{
    Policy: &vaultguard.Policy{
        Local: &vaultguard.LocalPolicy{
            MinSecurityScore: 50,
            RequireEncryption: true,  // 🔒
        },
        Cloud: &vaultguard.CloudPolicy{
            RequireIAM: true,  // 🎭
            AWS: &vaultguard.AWSPolicy{RequireIRSA: true},
            GCP: &vaultguard.GCPPolicy{RequireWorkloadIdentity: true},
        },
    },
})
```

---

# Loading Agent Credentials 📦

## Convenient Helper Functions

```go
// 📥 Load multiple credentials at once
creds, err := sv.LoadRequiredCredentials(ctx, nil,
    "OPENAI_API_KEY",
    "ANTHROPIC_API_KEY",
    "SERPER_API_KEY",
    "DATABASE_URL",
)
if err != nil {
    log.Fatal("Missing required credentials:", err)
}

// 🔑 Access by name
openai := creds["OPENAI_API_KEY"]
anthropic := creds["ANTHROPIC_API_KEY"]
```

---

# Security Assessment API 📊

## Pre-Flight Checks

```go
// 🔍 Check security without accessing secrets
result, err := vaultguard.CheckSecurity(&vaultguard.Config{
    Policy: vaultguard.DefaultPolicy(),
})

fmt.Printf("🌍 Environment: %s\n", result.Environment)
fmt.Printf("📊 Security Score: %d/100\n", result.Score)
fmt.Printf("🏷️ Security Level: %s\n", result.Level)
fmt.Printf("✅ Passed: %v\n", result.Passed)

// 💡 Get actionable recommendations
for _, rec := range result.Recommendations {
    fmt.Printf("  → %s\n", rec)
}
```

---

# Security Levels 📈

## Clear Classification

| Score | Level | Description |
|-------|-------|-------------|
| 80+ | 🔴 **Critical** | Maximum security, all features enabled |
| 60-79 | 🟠 **High** | Strong security, most features enabled |
| 40-59 | 🟡 **Medium** | Basic security, some gaps |
| 20-39 | 🟣 **Low** | Weak security, significant risks |
| 0-19 | ⚫ **Minimal** | Insecure, immediate action needed |

---

# Error Handling ⚠️

## Actionable Security Errors

```go
sv, err := vaultguard.Quick()
if err != nil {
    var secErr *vaultguard.SecurityError
    if errors.As(err, &secErr) {
        fmt.Printf("❌ Security check failed!\n")
        fmt.Printf("📊 Score: %d (required: %d)\n",
            secErr.Score, secErr.Required)
        fmt.Printf("🔧 Fix these issues:\n")
        for _, rec := range secErr.Recommendations {
            fmt.Printf("  → %s\n", rec)
        }
    }
}
```

---

# Development Mode 🛠️

## Safe Local Testing

```go
// 🧪 For development only - relaxed security
sv, err := vaultguard.QuickDev()

// ⚙️ Or explicitly allow insecure environments
sv, err := vaultguard.New(&vaultguard.Config{
    Policy: &vaultguard.Policy{
        AllowInsecure: true,  // ⚠️ Bypass security checks
    },
})
```

> **⚠️ Warning**: Never use development mode in production!

---

# Kubernetes Deployment ☸️

## Pod Security Validation

```go
Policy: &vaultguard.Policy{
    Kubernetes: &vaultguard.KubernetesPolicy{
        // 📁 Namespace restrictions
        DeniedNamespaces: []string{"default", "kube-system"},
        AllowedNamespaces: []string{"agents", "production"},

        // 🛡️ Pod security requirements
        RequireServiceAccount: true,
        RequireNonRoot: true,
        RequireReadOnlyRootFS: true,
    },
}
```

---

# Cloud IAM Validation ☁️

## AWS Example

```go
Policy: &vaultguard.Policy{
    Cloud: &vaultguard.CloudPolicy{
        RequireIAM: true,
        AWS: &vaultguard.AWSPolicy{
            RequireIRSA: true,  // 🔑
            AllowedRoleARNs: []string{
                "arn:aws:iam::123456789:role/agent-*",
            },
            AllowedAccountIDs: []string{"123456789"},  // 🏢
            AllowedRegions: []string{"us-east-1", "us-west-2"},  // 🌍
        },
    },
}
```

---

# Architecture Overview 🏗️

```
┌────────────────────────────────────────────────────────────────┐
│                         VaultGuard                             │
├────────────────────────────────────────────────────────────────┤
│  config.go       │ 📄 File-based policy loading & merging      │
│  detect.go       │ 🔍 Environment detection (local/cloud/k8s)  │
│  policy.go       │ 📋 Security policy definitions & presets    │
│  local.go        │ 💻 Local security assessment (via Posture)  │
│  cloud.go        │ ☁️ Cloud & Kubernetes security validation   │
│  vault.go        │ 🔐 Main API, provider management            │
│  convenience.go  │ ⚡ Helper functions (Quick, GetEnv, etc.)   │
│  errors.go       │ ⚠️ Structured error types                   │
│  types.go        │ 📦 Data structures & enumerations           │
└────────────────────────────────────────────────────────────────┘
                              │
              ┌───────────────┴───────────────┐
              ▼                               ▼
    ┌──────────────────┐           ┌──────────────────┐
    │     Posture      │           │    OmniVault     │
    │ 🛡️ Security      │           │ 🔑 Secret        │
    │    Assessment    │           │    Management    │
    └──────────────────┘           └──────────────────┘
```

---

# Quick Start 🚀

## Get Running in Minutes

```bash
# 📥 Install
go get github.com/agentplexus/vaultguard
```

```go
package main

import "github.com/agentplexus/vaultguard"

func main() {
    sv, err := vaultguard.Quick()  // 🛡️ Security-gated
    if err != nil {
        log.Fatal(err)
    }
    defer sv.Close()

    apiKey, err := sv.GetValue(ctx, "API_KEY")  // 🔑 Safe access
    // Use apiKey safely...
}
```

---

# Use Case: Secure Agent Team 👥

## Multi-Agent System with Shared Credentials

```go
func initAgentTeam() (*Team, error) {
    sv, _ := vaultguard.New(&Config{
        Policy: DefaultPolicy(),  // 🛡️
    })

    // 🎯 Each agent gets only the credentials it needs
    researcher, _ := NewResearcher(sv.GetValue(ctx, "SERPER_KEY"))
    writer, _ := NewWriter(sv.GetValue(ctx, "ANTHROPIC_KEY"))
    reviewer, _ := NewReviewer(sv.GetValue(ctx, "OPENAI_KEY"))

    return &Team{researcher, writer, reviewer}, nil
}
```

---

# Use Case: MCP Tool Server 🔧

## Secure Tool Execution Environment

```go
type SecureToolServer struct {
    vault *vaultguard.SecureVault  // 🛡️
}

func (s *SecureToolServer) ExecuteTool(name string, params map[string]any) (any, error) {
    // 🔑 Tools access credentials through the secure vault
    switch name {
    case "query_database":
        dbURL, _ := s.vault.GetValue(ctx, "DATABASE_URL")
        return queryDB(dbURL, params)  // 🗄️
    case "call_api":
        apiKey, _ := s.vault.GetValue(ctx, "EXTERNAL_API_KEY")
        return callAPI(apiKey, params)  // 🌐
    }
    return nil, fmt.Errorf("unknown tool: %s", name)
}
```

---

# Best Practices ✨

## Recommendations for Production

1. 🏭 **Always use DefaultPolicy() or StrictPolicy()** in production
2. 🚫 **Never commit AllowInsecure: true** to version control
3. 📄 **Use file-based configuration** to separate policy from code
4. 🏢 **Deploy enterprise policies** with locked fields for org-wide enforcement
5. 📝 **Log security assessments** for audit trails
6. 💡 **Handle SecurityError** gracefully with user guidance
7. 🔄 **Rotate credentials** regularly using your secret backend
8. 📁 **Restrict namespace access** in Kubernetes deployments

---

# Summary 📝

## VaultGuard Delivers

- 🛡️ **Security-First**: Credentials only accessible in verified environments
- ⚡ **Zero Configuration**: Auto-detects environment and selects providers
- ☁️ **Cloud-Native**: First-class AWS, GCP, Azure, Kubernetes support
- 🤖 **Agent-Ready**: Built for autonomous AI systems
- 🔌 **MCP Compatible**: Secure tool credential management
- 📋 **Flexible Policies**: From development to high-security production
- 🏢 **Enterprise Ready**: File-based config with locked fields for org control

---

# Get Started 🎯

## Resources

- 📦 **GitHub**: github.com/agentplexus/vaultguard
- 📚 **Docs**: agentplexus.github.io/vaultguard
- 🔗 **Dependencies**:
  - github.com/agentplexus/posture
  - github.com/agentplexus/omnivault

```go
go get github.com/agentplexus/vaultguard
```

---

<!-- _class: lead -->

# Questions❓

## Secure your agents. Protect your credentials. 🛡️

**VaultGuard** - Security-Gated Credential Access

