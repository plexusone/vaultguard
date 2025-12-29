package vaultguard

import (
	"context"
)

// Quick creates a SecureVault with sensible defaults for the detected environment.
// This is the simplest way to get started with OmniSafe.
//
// Example:
//
//	sv, err := omnisafe.Quick()
//	if err != nil {
//	    log.Fatal(err)
//	}
//	defer sv.Close()
//
//	apiKey, err := sv.GetValue(ctx, "API_KEY")
func Quick() (*SecureVault, error) {
	return New(nil)
}

// QuickDev creates a SecureVault with permissive settings for development.
// Security checks are relaxed and the environment provider is used.
//
// WARNING: Do not use in production!
func QuickDev() (*SecureVault, error) {
	return New(&Config{
		Policy: DevelopmentPolicy(),
	})
}

// QuickStrict creates a SecureVault with strict security requirements.
// Use this for high-security environments.
func QuickStrict() (*SecureVault, error) {
	return New(&Config{
		Policy: StrictPolicy(),
	})
}

// MustQuick is like Quick but panics on error.
func MustQuick() *SecureVault {
	sv, err := Quick()
	if err != nil {
		panic(err)
	}
	return sv
}

// MustQuickDev is like QuickDev but panics on error.
func MustQuickDev() *SecureVault {
	sv, err := QuickDev()
	if err != nil {
		panic(err)
	}
	return sv
}

// CheckSecurity performs security checks without creating a vault.
// Useful for pre-flight checks before starting an application.
//
// Example:
//
//	result, err := omnisafe.CheckSecurity(nil)
//	if err != nil {
//	    log.Fatalf("Security check failed: %v", err)
//	}
//	fmt.Printf("Security score: %d\n", result.Score)
func CheckSecurity(policy *Policy) (*SecurityResult, error) {
	if policy == nil {
		policy = DefaultPolicy()
	}

	env := DetectEnvironment()

	if env.IsLocal() {
		return CheckLocalSecurity(policy.Local)
	}

	result, err := CheckCloudSecurity(env, policy.Cloud)
	if err != nil {
		return result, err
	}

	if env.IsKubernetes() && policy.Kubernetes != nil {
		return CheckKubernetesSecurity(env, policy.Kubernetes)
	}

	return result, nil
}

// RequireSecurity checks security and returns an error if it fails.
// This is a convenience wrapper around CheckSecurity for use in init functions.
//
// Example:
//
//	func init() {
//	    if err := omnisafe.RequireSecurity(nil); err != nil {
//	        log.Fatal(err)
//	    }
//	}
func RequireSecurity(policy *Policy) error {
	_, err := CheckSecurity(policy)
	return err
}

// GetEnv is a convenience function that retrieves an environment variable
// through the secure vault. It performs security checks before access.
//
// Example:
//
//	apiKey, err := omnisafe.GetEnv(ctx, "API_KEY", nil)
func GetEnv(ctx context.Context, name string, policy *Policy) (string, error) {
	sv, err := New(&Config{Policy: policy})
	if err != nil {
		return "", err
	}
	defer func() { _ = sv.Close() }()

	return sv.GetValue(ctx, name)
}

// MustGetEnv is like GetEnv but panics on error.
func MustGetEnv(ctx context.Context, name string, policy *Policy) string {
	value, err := GetEnv(ctx, name, policy)
	if err != nil {
		panic(err)
	}
	return value
}

// LoadCredentials loads multiple credentials at once and returns them as a map.
// Missing credentials are returned as empty strings (no error).
//
// Example:
//
//	creds, err := omnisafe.LoadCredentials(ctx, nil,
//	    "GOOGLE_API_KEY",
//	    "ANTHROPIC_API_KEY",
//	    "OPENAI_API_KEY",
//	)
func LoadCredentials(ctx context.Context, policy *Policy, names ...string) (map[string]string, error) {
	sv, err := New(&Config{Policy: policy})
	if err != nil {
		return nil, err
	}
	defer func() { _ = sv.Close() }()

	creds := make(map[string]string, len(names))
	for _, name := range names {
		value, err := sv.GetValue(ctx, name)
		if err != nil {
			// Continue on missing credentials
			creds[name] = ""
		} else {
			creds[name] = value
		}
	}

	return creds, nil
}

// LoadRequiredCredentials loads credentials and returns an error if any are missing.
//
// Example:
//
//	creds, err := omnisafe.LoadRequiredCredentials(ctx, nil,
//	    "GOOGLE_API_KEY",
//	    "SERPER_API_KEY",
//	)
func LoadRequiredCredentials(ctx context.Context, policy *Policy, names ...string) (map[string]string, error) {
	sv, err := New(&Config{Policy: policy})
	if err != nil {
		return nil, err
	}
	defer func() { _ = sv.Close() }()

	creds := make(map[string]string, len(names))
	var missing []string

	for _, name := range names {
		value, err := sv.GetValue(ctx, name)
		if err != nil || value == "" {
			missing = append(missing, name)
		} else {
			creds[name] = value
		}
	}

	if len(missing) > 0 {
		return creds, &ProviderError{
			Err:       ErrSecretNotFound,
			Provider:  ProviderEnv,
			Operation: "get",
			Path:      missing[0],
		}
	}

	return creds, nil
}
