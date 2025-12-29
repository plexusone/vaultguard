package vaultguard

import (
	"errors"
	"fmt"
)

// Standard errors for omnisafe operations.
var (
	// ErrSecurityCheckFailed indicates the security policy was not met.
	ErrSecurityCheckFailed = errors.New("security check failed")

	// ErrEnvironmentNotSupported indicates the environment is not supported.
	ErrEnvironmentNotSupported = errors.New("environment not supported")

	// ErrProviderNotAvailable indicates the secret provider is not available.
	ErrProviderNotAvailable = errors.New("provider not available")

	// ErrPolicyViolation indicates a policy requirement was violated.
	ErrPolicyViolation = errors.New("policy violation")

	// ErrNotInitialized indicates the SecureVault was not properly initialized.
	ErrNotInitialized = errors.New("secure vault not initialized")

	// ErrSecretNotFound indicates the requested secret was not found.
	ErrSecretNotFound = errors.New("secret not found")

	// ErrAccessDenied indicates access to the secret was denied.
	ErrAccessDenied = errors.New("access denied")
)

// SecurityError represents a security check failure with details.
type SecurityError struct {
	// Err is the underlying error.
	Err error

	// Check is the name of the security check that failed.
	Check string

	// Required is what was required.
	Required string

	// Actual is what was found.
	Actual string

	// Environment where the check was performed.
	Environment Environment

	// Recommendations for fixing the issue.
	Recommendations []string
}

// Error implements the error interface.
func (e *SecurityError) Error() string {
	if e.Required != "" && e.Actual != "" {
		return fmt.Sprintf("%s: %s (required: %s, actual: %s)", e.Err, e.Check, e.Required, e.Actual)
	}
	return fmt.Sprintf("%s: %s", e.Err, e.Check)
}

// Unwrap returns the underlying error.
func (e *SecurityError) Unwrap() error {
	return e.Err
}

// NewSecurityError creates a new SecurityError.
func NewSecurityError(check, required, actual string, env Environment, recommendations ...string) *SecurityError {
	return &SecurityError{
		Err:             ErrSecurityCheckFailed,
		Check:           check,
		Required:        required,
		Actual:          actual,
		Environment:     env,
		Recommendations: recommendations,
	}
}

// PolicyError represents a policy violation with details.
type PolicyError struct {
	// Err is the underlying error.
	Err error

	// Policy is the name of the policy that was violated.
	Policy string

	// Requirement describes what was required.
	Requirement string

	// Details provides additional context.
	Details string
}

// Error implements the error interface.
func (e *PolicyError) Error() string {
	if e.Details != "" {
		return fmt.Sprintf("%s: %s - %s (%s)", e.Err, e.Policy, e.Requirement, e.Details)
	}
	return fmt.Sprintf("%s: %s - %s", e.Err, e.Policy, e.Requirement)
}

// Unwrap returns the underlying error.
func (e *PolicyError) Unwrap() error {
	return e.Err
}

// NewPolicyError creates a new PolicyError.
func NewPolicyError(policy, requirement, details string) *PolicyError {
	return &PolicyError{
		Err:         ErrPolicyViolation,
		Policy:      policy,
		Requirement: requirement,
		Details:     details,
	}
}

// ProviderError represents a provider-related error.
type ProviderError struct {
	// Err is the underlying error.
	Err error

	// Provider is the provider that encountered the error.
	Provider Provider

	// Operation is the operation that failed.
	Operation string

	// Path is the secret path (if applicable).
	Path string

	// Cause is the underlying cause.
	Cause error
}

// Error implements the error interface.
func (e *ProviderError) Error() string {
	if e.Cause != nil {
		return fmt.Sprintf("%s: provider %s failed %s on %s: %v", e.Err, e.Provider, e.Operation, e.Path, e.Cause)
	}
	return fmt.Sprintf("%s: provider %s failed %s on %s", e.Err, e.Provider, e.Operation, e.Path)
}

// Unwrap returns the underlying error.
func (e *ProviderError) Unwrap() error {
	if e.Cause != nil {
		return e.Cause
	}
	return e.Err
}

// NewProviderError creates a new ProviderError.
func NewProviderError(provider Provider, operation, path string, cause error) *ProviderError {
	return &ProviderError{
		Err:       ErrProviderNotAvailable,
		Provider:  provider,
		Operation: operation,
		Path:      path,
		Cause:     cause,
	}
}
