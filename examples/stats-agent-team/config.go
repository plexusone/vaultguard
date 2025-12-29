// Example: Integrating OmniSafe with stats-agent-team
//
// This shows how to modify pkg/config/config.go to use OmniSafe
// for security-gated credential loading.
package main

import (
	"context"
	"fmt"
	"log"
	"log/slog"
	"os"

	"github.com/agentplexus/vaultguard"
)

// Config represents the agent configuration (from stats-agent-team/pkg/config)
type Config struct {
	// LLM Configuration
	LLMProvider string
	LLMAPIKey   string
	LLMModel    string
	LLMBaseURL  string

	// Provider-specific API keys
	GeminiAPIKey string
	ClaudeAPIKey string
	OpenAIAPIKey string
	XAIAPIKey    string
	OllamaURL    string

	// Search Configuration
	SearchProvider string
	SerperAPIKey   string
	SerpAPIKey     string

	// Security metadata
	SecurityScore int
	Environment   string
	Provider      string
}

// SecurityConfig holds security policy settings (from Helm values)
type SecurityConfig struct {
	// MinSecurityScore is the minimum required security score (0-100)
	MinSecurityScore int `yaml:"minSecurityScore"`

	// RequireEncryption requires disk encryption on local workstations
	RequireEncryption bool `yaml:"requireEncryption"`

	// RequireTPM requires TPM/Secure Enclave on local workstations
	RequireTPM bool `yaml:"requireTPM"`

	// RequireIRSA requires IRSA on AWS EKS
	RequireIRSA bool `yaml:"requireIRSA"`

	// AllowedRoleARNs restricts to specific IAM roles (EKS)
	AllowedRoleARNs []string `yaml:"allowedRoleARNs"`

	// AllowedNamespaces restricts to specific Kubernetes namespaces
	AllowedNamespaces []string `yaml:"allowedNamespaces"`

	// DeniedNamespaces blocks specific namespaces
	DeniedNamespaces []string `yaml:"deniedNamespaces"`

	// AllowInsecure allows running without security checks (dev only)
	AllowInsecure bool `yaml:"allowInsecure"`
}

// LoadConfigWithSecurity loads configuration with security checks.
// This replaces the original LoadConfig() function.
func LoadConfigWithSecurity(ctx context.Context, secCfg *SecurityConfig) (*Config, error) {
	// Build OmniSafe policy from security config
	policy := buildPolicy(secCfg)

	// Create secure vault with optional logging
	var logger *slog.Logger
	if os.Getenv("OMNISAFE_DEBUG") == "true" {
		logger = slog.New(slog.NewTextHandler(os.Stderr, &slog.HandlerOptions{
			Level: slog.LevelDebug,
		}))
	}

	sv, err := vaultguard.New(&vaultguard.Config{
		Policy: policy,
		Logger: logger,
	})
	if err != nil {
		return nil, fmt.Errorf("security check failed: %w", err)
	}
	defer func() { _ = sv.Close() }()

	// Log security status
	result := sv.SecurityResult()
	log.Printf("Security: env=%s, provider=%s, score=%d, passed=%v",
		sv.Environment(), sv.Provider(), result.Score, result.Passed)

	// Load configuration using secure vault
	cfg := &Config{
		// Security metadata
		SecurityScore: result.Score,
		Environment:   sv.Environment().String(),
		Provider:      sv.Provider().String(),

		// LLM Configuration
		LLMProvider: getEnvWithDefault(ctx, sv, "LLM_PROVIDER", "gemini"),
		LLMModel:    getEnv(ctx, sv, "LLM_MODEL"),
		LLMBaseURL:  getEnv(ctx, sv, "LLM_BASE_URL"),

		// Search Configuration
		SearchProvider: getEnvWithDefault(ctx, sv, "SEARCH_PROVIDER", "serper"),
	}

	// Load provider-specific API keys with fallbacks
	cfg.GeminiAPIKey = getEnvWithFallback(ctx, sv, "GEMINI_API_KEY", "GOOGLE_API_KEY")
	cfg.ClaudeAPIKey = getEnvWithFallback(ctx, sv, "CLAUDE_API_KEY", "ANTHROPIC_API_KEY")
	cfg.OpenAIAPIKey = getEnv(ctx, sv, "OPENAI_API_KEY")
	cfg.XAIAPIKey = getEnv(ctx, sv, "XAI_API_KEY")
	cfg.OllamaURL = getEnvWithDefault(ctx, sv, "OLLAMA_URL", "http://localhost:11434")

	// Generic LLM_API_KEY overrides provider-specific keys
	if genericKey := getEnv(ctx, sv, "LLM_API_KEY"); genericKey != "" {
		cfg.LLMAPIKey = genericKey
	} else {
		// Use provider-specific key
		switch cfg.LLMProvider {
		case "gemini":
			cfg.LLMAPIKey = cfg.GeminiAPIKey
		case "claude":
			cfg.LLMAPIKey = cfg.ClaudeAPIKey
		case "openai":
			cfg.LLMAPIKey = cfg.OpenAIAPIKey
		case "xai":
			cfg.LLMAPIKey = cfg.XAIAPIKey
		}
	}

	// Search API keys
	cfg.SerperAPIKey = getEnv(ctx, sv, "SERPER_API_KEY")
	cfg.SerpAPIKey = getEnv(ctx, sv, "SERPAPI_API_KEY")

	return cfg, nil
}

// buildPolicy creates an OmniSafe policy from security config
func buildPolicy(secCfg *SecurityConfig) *vaultguard.Policy {
	if secCfg == nil {
		return vaultguard.DefaultPolicy()
	}

	policy := &vaultguard.Policy{
		AllowInsecure:  secCfg.AllowInsecure,
		InsecureReason: "Configured via Helm values",
	}

	// Local security settings
	if secCfg.MinSecurityScore > 0 || secCfg.RequireEncryption || secCfg.RequireTPM {
		policy.Local = &vaultguard.LocalPolicy{
			MinSecurityScore:  secCfg.MinSecurityScore,
			RequireEncryption: secCfg.RequireEncryption,
			RequireTPM:        secCfg.RequireTPM,
		}
	}

	// Cloud security settings
	if secCfg.RequireIRSA || len(secCfg.AllowedRoleARNs) > 0 {
		policy.Cloud = &vaultguard.CloudPolicy{
			RequireIAM: true,
			AWS: &vaultguard.AWSPolicy{
				RequireIRSA:     secCfg.RequireIRSA,
				AllowedRoleARNs: secCfg.AllowedRoleARNs,
			},
		}
	}

	// Kubernetes settings
	if len(secCfg.AllowedNamespaces) > 0 || len(secCfg.DeniedNamespaces) > 0 {
		policy.Kubernetes = &vaultguard.KubernetesPolicy{
			AllowedNamespaces: secCfg.AllowedNamespaces,
			DeniedNamespaces:  secCfg.DeniedNamespaces,
		}
	}

	// Provider mapping
	policy.ProviderMap = map[vaultguard.Environment]vaultguard.Provider{
		vaultguard.EnvLocal: vaultguard.ProviderEnv, // keyring when module available
		vaultguard.EnvEKS:   vaultguard.ProviderEnv, // aws-sm when module available
		vaultguard.EnvGKE:   vaultguard.ProviderEnv, // gcp-sm when module available
		vaultguard.EnvAKS:   vaultguard.ProviderEnv, // azure-kv when module available
	}
	policy.FallbackProvider = vaultguard.ProviderEnv

	return policy
}

// Helper functions for secure credential access

func getEnv(ctx context.Context, sv *vaultguard.SecureVault, key string) string {
	value, _ := sv.GetValue(ctx, key)
	return value
}

func getEnvWithDefault(ctx context.Context, sv *vaultguard.SecureVault, key, defaultValue string) string {
	value, err := sv.GetValue(ctx, key)
	if err != nil || value == "" {
		return defaultValue
	}
	return value
}

func getEnvWithFallback(ctx context.Context, sv *vaultguard.SecureVault, primary, fallback string) string {
	value := getEnv(ctx, sv, primary)
	if value == "" {
		value = getEnv(ctx, sv, fallback)
	}
	return value
}

// Example main function showing usage
func main() {
	ctx := context.Background()

	// Security config would come from Helm values in production
	secCfg := &SecurityConfig{
		MinSecurityScore:  50,
		RequireEncryption: true,
		RequireIRSA:       true,
		DeniedNamespaces:  []string{"default", "kube-system"},
		AllowInsecure:     os.Getenv("ALLOW_INSECURE") == "true",
	}

	cfg, err := LoadConfigWithSecurity(ctx, secCfg)
	if err != nil {
		log.Fatalf("Failed to load config: %v", err)
	}

	fmt.Printf("Configuration loaded successfully:\n")
	fmt.Printf("  Environment:    %s\n", cfg.Environment)
	fmt.Printf("  Security Score: %d\n", cfg.SecurityScore)
	fmt.Printf("  LLM Provider:   %s\n", cfg.LLMProvider)
	fmt.Printf("  Search Provider: %s\n", cfg.SearchProvider)
	fmt.Printf("  API Key Set:    %v\n", cfg.LLMAPIKey != "")
}
