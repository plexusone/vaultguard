package vaultguard

import (
	"fmt"
	"slices"
	"time"

	"github.com/plexusone/posture/inspector"
)

// CheckLocalSecurity performs security checks for local workstation environments
// using Posture.
func CheckLocalSecurity(policy *LocalPolicy) (*SecurityResult, error) {
	result := &SecurityResult{
		Environment: EnvLocal,
		Timestamp:   time.Now(),
		Details: SecurityDetails{
			Local: &LocalSecurityDetails{},
		},
	}

	// Get security summary from Posture
	summary, err := inspector.GetSecuritySummary()
	if err != nil {
		result.Passed = false
		result.Level = SecurityCritical
		result.Message = fmt.Sprintf("Failed to get security summary: %v", err)
		return result, err
	}

	// Populate local security details
	local := result.Details.Local
	local.Platform = summary.Platform

	if summary.TPM != nil {
		local.TPMPresent = summary.TPM.Present
		local.TPMEnabled = summary.TPM.Enabled
		local.TPMType = summary.TPM.Type
	}

	if summary.Encryption != nil {
		local.DiskEncrypted = summary.Encryption.Enabled
		local.EncryptionType = summary.Encryption.Type
	}

	if summary.SecureBoot != nil {
		local.SecureBootEnabled = summary.SecureBoot.Enabled
	}

	if summary.Biometrics != nil {
		local.BiometricsAvailable = summary.Biometrics.Available
		local.BiometricsConfigured = summary.Biometrics.Configured
	}

	// Set score and level from Posture
	result.Score = summary.OverallScore
	result.Level = scoreToLevel(summary.OverallScore)
	result.Recommendations = summary.Recommendations

	// If no policy, just return the assessment
	if policy == nil {
		result.Passed = true
		result.Message = fmt.Sprintf("Security score: %d/100 (%s)", result.Score, result.Level)
		return result, nil
	}

	// Validate against policy
	var violations []string

	// Check platform
	if len(policy.AllowedPlatforms) > 0 {
		if !slices.Contains(policy.AllowedPlatforms, local.Platform) {
			violations = append(violations, fmt.Sprintf(
				"platform %s not in allowed list %v", local.Platform, policy.AllowedPlatforms,
			))
		}
	}

	// Check minimum security score
	if policy.MinSecurityScore > 0 && result.Score < policy.MinSecurityScore {
		violations = append(violations, fmt.Sprintf(
			"security score %d below minimum %d", result.Score, policy.MinSecurityScore,
		))
	}

	// Check disk encryption
	if policy.RequireEncryption && !local.DiskEncrypted {
		violations = append(violations, "disk encryption required but not enabled")
		result.Recommendations = append(result.Recommendations, getEncryptionRecommendation(local.Platform))
	}

	// Check TPM/Secure Enclave
	if policy.RequireTPM && (!local.TPMPresent || !local.TPMEnabled) {
		violations = append(violations, "TPM/Secure Enclave required but not available or enabled")
	}

	// Check Secure Boot
	if policy.RequireSecureBoot && !local.SecureBootEnabled {
		violations = append(violations, "Secure Boot required but not enabled")
	}

	// Check biometrics
	if policy.RequireBiometrics && (!local.BiometricsAvailable || !local.BiometricsConfigured) {
		violations = append(violations, "biometric authentication required but not configured")
	}

	// Set result based on violations
	if len(violations) > 0 {
		result.Passed = false
		result.Message = fmt.Sprintf("Security policy violations: %v", violations)
		return result, &SecurityError{
			Err:             ErrSecurityCheckFailed,
			Check:           "local security policy",
			Environment:     EnvLocal,
			Recommendations: result.Recommendations,
		}
	}

	result.Passed = true
	result.Message = fmt.Sprintf("Security checks passed (score: %d/100)", result.Score)
	return result, nil
}

// scoreToLevel converts a numeric score to a SecurityLevel.
func scoreToLevel(score int) SecurityLevel {
	switch {
	case score >= 90:
		return SecurityExcellent
	case score >= 70:
		return SecurityHigh
	case score >= 50:
		return SecurityMedium
	case score >= 25:
		return SecurityLow
	default:
		return SecurityCritical
	}
}

// getEncryptionRecommendation returns platform-specific encryption guidance.
func getEncryptionRecommendation(platform string) string {
	switch platform {
	case "darwin":
		return "Enable FileVault: System Settings > Privacy & Security > FileVault"
	case "windows":
		return "Enable BitLocker: Settings > Privacy & security > Device encryption"
	case "linux":
		return "Enable LUKS encryption for your disk partitions"
	default:
		return "Enable full disk encryption for your platform"
	}
}

// IsLocalSecuritySupported returns true if local security checks are available.
// This checks if Posture functionality is available on the current platform.
func IsLocalSecuritySupported() bool {
	// Posture supports darwin, windows, and linux
	env := DetectEnvironment()
	return env == EnvLocal
}

// GetLocalSecuritySummary returns a quick security summary for local environments.
func GetLocalSecuritySummary() (*inspector.SecuritySummary, error) {
	return inspector.GetSecuritySummary()
}
