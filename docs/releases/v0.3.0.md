# VaultGuard v0.3.0 Release Notes

**Release Date:** March 2026

This release migrates the module path from the agentplexus organization to plexusone.

## Breaking Changes

### Module Path Changed

The Go module path has changed from `github.com/agentplexus/vaultguard` to `github.com/plexusone/vaultguard`.

**Before (v0.2.0):**
```go
import "github.com/agentplexus/vaultguard"
```

**After (v0.3.0):**
```go
import "github.com/plexusone/vaultguard"
```

## What's Changed

### Changed

- Module path migrated from `github.com/agentplexus/vaultguard` to `github.com/plexusone/vaultguard`
- Updated omnivault dependency to `github.com/plexusone/omnivault` v0.3.0
- Updated posture dependency to `github.com/plexusone/posture` v0.3.0
- All internal imports updated to use new module path
- Documentation URLs updated to plexusone organization
- CI workflows migrated to standard plexusone reusable workflows

### Fixed

- Suppress gosec G703 false positives for policy file loading and SDK token paths

## Installation

```bash
go get github.com/plexusone/vaultguard@v0.3.0
```

## Migration Guide

1. Update your `go.mod` imports:
   - Change `github.com/agentplexus/vaultguard` to `github.com/plexusone/vaultguard`
2. Run `go mod tidy`
3. Update any import statements in your code

## Dependencies

This release requires:

- `github.com/plexusone/omnivault` v0.3.0
- `github.com/plexusone/posture` v0.3.0
