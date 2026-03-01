package vaultguard

import (
	"context"
	"fmt"
	"log/slog"
	"sync"

	"github.com/plexusone/omnivault"
	"github.com/plexusone/omnivault/vault"
)

// SecureVault provides security-gated access to secrets.
// It combines Posture security checks with OmniVault secret management.
type SecureVault struct {
	mu sync.RWMutex

	// Configuration
	policy *Policy
	logger *slog.Logger

	// Environment and security state
	env            Environment
	securityResult *SecurityResult
	securityPassed bool

	// OmniVault components
	resolver *omnivault.Resolver
	client   *omnivault.Client
	provider Provider

	// State
	initialized bool
	closed      bool
}

// Config configures a SecureVault.
type Config struct {
	// Policy defines security requirements. If nil, DefaultPolicy() is used.
	Policy *Policy

	// Logger for structured logging. If nil, logging is disabled.
	Logger *slog.Logger

	// ForceEnvironment overrides automatic environment detection.
	// Use for testing or when detection fails.
	ForceEnvironment Environment

	// CustomVault allows injecting a custom vault.Vault implementation.
	// If set, this takes precedence over provider auto-selection.
	CustomVault vault.Vault

	// ProviderConfig is passed to the auto-selected provider.
	ProviderConfig any
}

// New creates a new SecureVault with the given configuration.
// It automatically detects the environment, performs security checks,
// and initializes the appropriate secret provider.
func New(cfg *Config) (*SecureVault, error) {
	if cfg == nil {
		cfg = &Config{}
	}

	sv := &SecureVault{
		policy: cfg.Policy,
		logger: cfg.Logger,
	}

	// Use default policy if none provided
	if sv.policy == nil {
		sv.policy = DefaultPolicy()
	}

	// Detect or use forced environment
	if cfg.ForceEnvironment != "" {
		sv.env = cfg.ForceEnvironment
	} else {
		sv.env = DetectEnvironment()
	}

	sv.log("Detected environment", "environment", sv.env)

	// Perform security checks
	result, err := sv.checkSecurity()
	sv.securityResult = result

	if err != nil {
		if sv.policy.AllowInsecure {
			sv.log("Security check failed but AllowInsecure is set",
				"error", err,
				"reason", sv.policy.InsecureReason)
			sv.securityPassed = true
		} else {
			return nil, fmt.Errorf("security check failed: %w", err)
		}
	} else {
		sv.securityPassed = true
	}

	// Initialize vault provider
	if cfg.CustomVault != nil {
		// Use custom vault
		sv.client, err = omnivault.NewClient(omnivault.Config{
			CustomVault: cfg.CustomVault,
		})
		if err != nil {
			return nil, fmt.Errorf("failed to create client with custom vault: %w", err)
		}
		sv.provider = Provider(cfg.CustomVault.Name())
	} else {
		// Auto-select provider based on environment and policy
		sv.provider = sv.selectProvider()
		sv.log("Selected provider", "provider", sv.provider)

		// Initialize provider
		if err := sv.initializeProvider(cfg.ProviderConfig); err != nil {
			return nil, fmt.Errorf("failed to initialize provider %s: %w", sv.provider, err)
		}
	}

	sv.initialized = true
	return sv, nil
}

// checkSecurity performs security checks based on the environment.
func (sv *SecureVault) checkSecurity() (*SecurityResult, error) {
	if sv.env.IsLocal() {
		return CheckLocalSecurity(sv.policy.Local)
	}

	// For cloud environments, check cloud security first
	result, err := CheckCloudSecurity(sv.env, sv.policy.Cloud)
	if err != nil {
		return result, err
	}

	// Then check Kubernetes policy if applicable
	if sv.env.IsKubernetes() && sv.policy.Kubernetes != nil {
		k8sResult, k8sErr := CheckKubernetesSecurity(sv.env, sv.policy.Kubernetes)
		if k8sErr != nil {
			return k8sResult, k8sErr
		}
	}

	return result, nil
}

// selectProvider chooses the appropriate provider based on environment and policy.
func (sv *SecureVault) selectProvider() Provider {
	// Check policy provider map first
	if sv.policy.ProviderMap != nil {
		if provider, ok := sv.policy.ProviderMap[sv.env]; ok {
			return provider
		}
	}

	// Default provider selection based on environment
	switch sv.env {
	case EnvLocal:
		return ProviderKeyring
	case EnvEKS, EnvLambda:
		return ProviderAWSSecretsManager
	case EnvGKE, EnvCloudRun:
		return ProviderGCPSecretManager
	case EnvAKS, EnvAzureFunc:
		return ProviderAzureKeyVault
	case EnvKubernetes:
		return ProviderK8sSecret
	case EnvContainer:
		return ProviderEnv
	default:
		if sv.policy.FallbackProvider != "" {
			return sv.policy.FallbackProvider
		}
		return ProviderEnv
	}
}

// initializeProvider sets up the selected provider.
func (sv *SecureVault) initializeProvider(providerConfig any) error {
	var err error

	switch sv.provider {
	case ProviderEnv:
		sv.client, err = omnivault.NewClient(omnivault.Config{
			Provider: omnivault.ProviderEnv,
		})

	case ProviderFile:
		sv.client, err = omnivault.NewClient(omnivault.Config{
			Provider:       omnivault.ProviderFile,
			ProviderConfig: providerConfig,
		})

	case ProviderKeyring, ProviderAWSSecretsManager, ProviderAWSParameterStore,
		ProviderGCPSecretManager, ProviderAzureKeyVault, ProviderK8sSecret, ProviderVault:
		// These providers require external modules (omnivault-keyring, omnivault-aws, etc.)
		// For now, fall back to env provider with a warning
		sv.log("Provider requires external module, falling back to env",
			"requested_provider", sv.provider,
			"fallback", ProviderEnv)

		// Try to use custom vault if provided, otherwise fall back to env
		sv.provider = ProviderEnv
		sv.client, err = omnivault.NewClient(omnivault.Config{
			Provider: omnivault.ProviderEnv,
		})

	default:
		return fmt.Errorf("unsupported provider: %s", sv.provider)
	}

	return err
}

// Get retrieves a secret by path.
func (sv *SecureVault) Get(ctx context.Context, path string) (*vault.Secret, error) {
	if err := sv.ensureReady(); err != nil {
		return nil, err
	}
	return sv.client.Get(ctx, path)
}

// GetValue retrieves the primary value of a secret.
func (sv *SecureVault) GetValue(ctx context.Context, path string) (string, error) {
	if err := sv.ensureReady(); err != nil {
		return "", err
	}
	return sv.client.GetValue(ctx, path)
}

// GetField retrieves a specific field from a multi-field secret.
func (sv *SecureVault) GetField(ctx context.Context, path, field string) (string, error) {
	if err := sv.ensureReady(); err != nil {
		return "", err
	}
	return sv.client.GetField(ctx, path, field)
}

// Set stores a secret.
func (sv *SecureVault) Set(ctx context.Context, path string, secret *vault.Secret) error {
	if err := sv.ensureReady(); err != nil {
		return err
	}
	return sv.client.Set(ctx, path, secret)
}

// SetValue stores a simple string secret.
func (sv *SecureVault) SetValue(ctx context.Context, path, value string) error {
	return sv.Set(ctx, path, &vault.Secret{Value: value})
}

// Delete removes a secret.
func (sv *SecureVault) Delete(ctx context.Context, path string) error {
	if err := sv.ensureReady(); err != nil {
		return err
	}
	return sv.client.Delete(ctx, path)
}

// Exists checks if a secret exists.
func (sv *SecureVault) Exists(ctx context.Context, path string) (bool, error) {
	if err := sv.ensureReady(); err != nil {
		return false, err
	}
	return sv.client.Exists(ctx, path)
}

// List returns secrets matching a prefix.
func (sv *SecureVault) List(ctx context.Context, prefix string) ([]string, error) {
	if err := sv.ensureReady(); err != nil {
		return nil, err
	}
	return sv.client.List(ctx, prefix)
}

// Close releases resources.
func (sv *SecureVault) Close() error {
	sv.mu.Lock()
	defer sv.mu.Unlock()

	if sv.closed {
		return nil
	}

	sv.closed = true

	if sv.client != nil {
		return sv.client.Close()
	}
	if sv.resolver != nil {
		return sv.resolver.Close()
	}

	return nil
}

// Environment returns the detected environment.
func (sv *SecureVault) Environment() Environment {
	return sv.env
}

// Provider returns the active provider.
func (sv *SecureVault) Provider() Provider {
	return sv.provider
}

// SecurityResult returns the security check results.
func (sv *SecureVault) SecurityResult() *SecurityResult {
	return sv.securityResult
}

// IsSecure returns true if security checks passed.
func (sv *SecureVault) IsSecure() bool {
	return sv.securityPassed && sv.securityResult != nil && sv.securityResult.Passed
}

// ensureReady checks that the vault is ready for operations.
func (sv *SecureVault) ensureReady() error {
	sv.mu.RLock()
	defer sv.mu.RUnlock()

	if !sv.initialized {
		return ErrNotInitialized
	}
	if sv.closed {
		return fmt.Errorf("vault is closed")
	}
	if !sv.securityPassed {
		return ErrSecurityCheckFailed
	}

	return nil
}

// log logs a message if a logger is configured.
func (sv *SecureVault) log(msg string, args ...any) {
	if sv.logger != nil {
		sv.logger.Info(msg, args...)
	}
}
