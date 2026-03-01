package vaultguard

import (
	"encoding/json"
	"os"
	"path/filepath"
	"runtime"
)

const (
	// ConfigDirName is the directory name for agentplexus configuration.
	ConfigDirName = ".agentplexus"

	// PolicyFileName is the name of the policy configuration file.
	PolicyFileName = "policy.json"

	// EnvPolicyFile is the environment variable for explicit policy file path.
	EnvPolicyFile = "AGENTPLEXUS_POLICY_FILE"
)

// FilePolicy extends Policy with file-based configuration features.
type FilePolicy struct {
	// Version of the policy file format.
	Version int `json:"version"`

	// Extends the base Policy.
	Policy

	// Locked lists field paths that cannot be overridden by user config.
	// Used by enterprise/system configs to enforce settings.
	// Example: ["local.require_encryption", "local.min_security_score"]
	Locked []string `json:"locked,omitempty"`
}

// systemConfigPath returns the system-wide configuration path.
func systemConfigPath() string {
	switch runtime.GOOS {
	case "windows":
		return filepath.Join(os.Getenv("ProgramData"), "agentplexus", PolicyFileName)
	default:
		return filepath.Join("/etc", "agentplexus", PolicyFileName)
	}
}

// userConfigPath returns the user-specific configuration path.
func userConfigPath() string {
	home, err := os.UserHomeDir()
	if err != nil {
		return ""
	}
	return filepath.Join(home, ConfigDirName, PolicyFileName)
}

// LoadPolicyFromFile loads a policy from a specific file path.
func LoadPolicyFromFile(path string) (*FilePolicy, error) {
	// #nosec G703 -- path is intentionally caller-provided for policy file loading
	data, err := os.ReadFile(path)
	if err != nil {
		return nil, err
	}

	var fp FilePolicy
	if err := json.Unmarshal(data, &fp); err != nil {
		return nil, err
	}

	return &fp, nil
}

// LoadPolicy loads the policy from the configuration hierarchy.
// Precedence (highest to lowest):
//  1. AGENTPLEXUS_POLICY_FILE environment variable
//  2. User config (~/.agentplexus/policy.yaml)
//  3. System config (/etc/agentplexus/policy.yaml)
//  4. nil (no policy = permissive mode)
func LoadPolicy() (*Policy, error) {
	var systemPolicy, userPolicy *FilePolicy
	var err error

	// Check for env var override first
	if envPath := os.Getenv(EnvPolicyFile); envPath != "" {
		fp, err := LoadPolicyFromFile(envPath)
		if err != nil {
			return nil, err
		}
		return &fp.Policy, nil
	}

	// Load system config (enterprise)
	systemPath := systemConfigPath()
	if fileExists(systemPath) {
		systemPolicy, err = LoadPolicyFromFile(systemPath)
		if err != nil {
			return nil, err
		}
	}

	// Load user config
	userPath := userConfigPath()
	if fileExists(userPath) {
		userPolicy, err = LoadPolicyFromFile(userPath)
		if err != nil {
			return nil, err
		}
	}

	// No config files = permissive mode (return nil)
	if systemPolicy == nil && userPolicy == nil {
		return nil, nil
	}

	// Only system config
	if userPolicy == nil {
		return &systemPolicy.Policy, nil
	}

	// Only user config
	if systemPolicy == nil {
		return &userPolicy.Policy, nil
	}

	// Both exist - merge with system taking precedence on locked fields
	merged := mergeFilePolicies(systemPolicy, userPolicy)
	return &merged.Policy, nil
}

// mergeFilePolicies merges user policy into system policy,
// respecting locked fields from the system policy.
func mergeFilePolicies(system, user *FilePolicy) *FilePolicy {
	result := &FilePolicy{
		Version: system.Version,
		Policy:  system.Policy,
		Locked:  system.Locked,
	}

	// Build a set of locked fields
	locked := make(map[string]bool)
	for _, field := range system.Locked {
		locked[field] = true
	}

	// Merge Local policy
	if user.Local != nil {
		if result.Local == nil {
			result.Local = &LocalPolicy{}
		}
		if !locked["local.min_security_score"] && user.Local.MinSecurityScore > 0 {
			result.Local.MinSecurityScore = user.Local.MinSecurityScore
		}
		if !locked["local.require_encryption"] {
			result.Local.RequireEncryption = user.Local.RequireEncryption || result.Local.RequireEncryption
		}
		if !locked["local.require_tpm"] {
			result.Local.RequireTPM = user.Local.RequireTPM || result.Local.RequireTPM
		}
		if !locked["local.require_secure_boot"] {
			result.Local.RequireSecureBoot = user.Local.RequireSecureBoot || result.Local.RequireSecureBoot
		}
		if !locked["local.require_biometrics"] {
			result.Local.RequireBiometrics = user.Local.RequireBiometrics || result.Local.RequireBiometrics
		}
		if !locked["local.allowed_platforms"] && len(user.Local.AllowedPlatforms) > 0 {
			result.Local.AllowedPlatforms = user.Local.AllowedPlatforms
		}
	}

	// Merge ProviderMap (user can add, but not override locked providers)
	if user.ProviderMap != nil {
		if result.ProviderMap == nil {
			result.ProviderMap = make(map[Environment]Provider)
		}
		for env, provider := range user.ProviderMap {
			if !locked["provider_map."+string(env)] {
				result.ProviderMap[env] = provider
			}
		}
	}

	// Merge FallbackProvider
	if !locked["fallback_provider"] && user.FallbackProvider != "" {
		result.FallbackProvider = user.FallbackProvider
	}

	// AllowInsecure can only be set more restrictive, not more permissive
	// System says false -> user cannot override to true
	if result.AllowInsecure && !user.AllowInsecure {
		result.AllowInsecure = false
	}

	return result
}

// fileExists checks if a file exists and is not a directory.
func fileExists(path string) bool {
	info, err := os.Stat(path)
	if err != nil {
		return false
	}
	return !info.IsDir()
}

// EnsureConfigDir creates the user config directory if it doesn't exist.
func EnsureConfigDir() (string, error) {
	home, err := os.UserHomeDir()
	if err != nil {
		return "", err
	}

	configDir := filepath.Join(home, ConfigDirName)
	if err := os.MkdirAll(configDir, 0700); err != nil {
		return "", err
	}

	return configDir, nil
}

// SavePolicy saves a policy to the user config directory.
func SavePolicy(policy *Policy) error {
	configDir, err := EnsureConfigDir()
	if err != nil {
		return err
	}

	fp := &FilePolicy{
		Version: 1,
		Policy:  *policy,
	}

	data, err := json.MarshalIndent(fp, "", "  ")
	if err != nil {
		return err
	}

	path := filepath.Join(configDir, PolicyFileName)
	return os.WriteFile(path, data, 0600)
}

// GetConfigPaths returns the paths that will be checked for configuration.
func GetConfigPaths() map[string]string {
	return map[string]string{
		"system": systemConfigPath(),
		"user":   userConfigPath(),
		"env":    os.Getenv(EnvPolicyFile),
	}
}
