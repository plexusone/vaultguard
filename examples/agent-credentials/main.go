// Example: Using OmniSafe for agent credentials
//
// This example demonstrates how to use OmniSafe to securely manage
// credentials for AI agents, with automatic environment detection
// and security policy enforcement.
//
// On macOS: Uses Posture security checks + Keychain storage
// On AWS EKS: Uses IRSA validation + AWS Secrets Manager
// On GCP GKE: Uses Workload Identity + GCP Secret Manager
package main

import (
	"context"
	"fmt"
	"log"
	"log/slog"
	"os"

	"github.com/plexusone/vaultguard"
)

func main() {
	ctx := context.Background()

	// Create a logger for visibility
	logger := slog.New(slog.NewTextHandler(os.Stdout, &slog.HandlerOptions{
		Level: slog.LevelInfo,
	}))

	// Define security policy
	policy := &vaultguard.Policy{
		// Local workstation requirements
		Local: &vaultguard.LocalPolicy{
			MinSecurityScore:  50,   // Require at least 50/100 security score
			RequireEncryption: true, // Require disk encryption
		},

		// Cloud requirements
		Cloud: &vaultguard.CloudPolicy{
			RequireIAM: true, // Require cloud IAM (IRSA, Workload Identity)
			AWS: &vaultguard.AWSPolicy{
				RequireIRSA: true,
				// Optionally restrict to specific roles
				// AllowedRoleARNs: []string{"arn:aws:iam::123456789:role/stats-agent-*"},
			},
		},

		// Kubernetes requirements
		Kubernetes: &vaultguard.KubernetesPolicy{
			DeniedNamespaces: []string{"default", "kube-system"},
		},

		// Provider selection per environment
		ProviderMap: map[vaultguard.Environment]vaultguard.Provider{
			vaultguard.EnvLocal: vaultguard.ProviderEnv, // Use env vars for now (keyring requires external module)
			vaultguard.EnvEKS:   vaultguard.ProviderAWSSecretsManager,
			vaultguard.EnvGKE:   vaultguard.ProviderGCPSecretManager,
			vaultguard.EnvAKS:   vaultguard.ProviderAzureKeyVault,
		},

		// Fallback for development
		FallbackProvider: vaultguard.ProviderEnv,
	}

	// Create SecureVault
	sv, err := vaultguard.New(&vaultguard.Config{
		Policy: policy,
		Logger: logger,
	})
	if err != nil {
		log.Fatalf("Failed to create SecureVault: %v", err)
	}
	defer func() { _ = sv.Close() }()

	// Print environment info
	fmt.Printf("\n=== OmniSafe Agent Credentials Example ===\n\n")
	fmt.Printf("Environment: %s\n", sv.Environment())
	fmt.Printf("Provider:    %s\n", sv.Provider())
	fmt.Printf("Secure:      %v\n", sv.IsSecure())

	// Print security result
	result := sv.SecurityResult()
	if result != nil {
		fmt.Printf("\nSecurity Assessment:\n")
		fmt.Printf("  Score:   %d/100\n", result.Score)
		fmt.Printf("  Level:   %s\n", result.Level)
		fmt.Printf("  Passed:  %v\n", result.Passed)
		fmt.Printf("  Message: %s\n", result.Message)

		if len(result.Recommendations) > 0 {
			fmt.Printf("\nRecommendations:\n")
			for _, rec := range result.Recommendations {
				fmt.Printf("  - %s\n", rec)
			}
		}
	}

	// Example: Load agent credentials
	fmt.Printf("\n=== Loading Agent Credentials ===\n\n")

	credentials := map[string]string{
		"LLM_PROVIDER":      "",
		"GOOGLE_API_KEY":    "",
		"ANTHROPIC_API_KEY": "",
		"SERPER_API_KEY":    "",
	}

	for key := range credentials {
		value, err := sv.GetValue(ctx, key)
		if err != nil {
			fmt.Printf("  %s: (not found)\n", key)
		} else if value != "" {
			// Mask the value for security
			masked := maskValue(value)
			fmt.Printf("  %s: %s\n", key, masked)
		} else {
			fmt.Printf("  %s: (empty)\n", key)
		}
	}

	fmt.Printf("\n=== Done ===\n")
}

// maskValue returns a masked version of a secret value for display.
func maskValue(value string) string {
	if len(value) <= 8 {
		return "****"
	}
	return value[:4] + "..." + value[len(value)-4:]
}
