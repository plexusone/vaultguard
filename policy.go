package vaultguard

// Policy defines security requirements for credential access.
// Different policies can be applied based on the detected environment.
type Policy struct {
	// Local contains security requirements for local workstations.
	// Used when Environment.IsLocal() returns true.
	Local *LocalPolicy `json:"local,omitempty"`

	// Cloud contains security requirements for cloud environments.
	// Used when Environment.IsCloud() returns true.
	Cloud *CloudPolicy `json:"cloud,omitempty"`

	// Kubernetes contains additional requirements for Kubernetes environments.
	// Used when Environment.IsKubernetes() returns true.
	Kubernetes *KubernetesPolicy `json:"kubernetes,omitempty"`

	// ProviderMap maps environments to preferred secret providers.
	// If not specified, providers are auto-selected based on environment.
	ProviderMap map[Environment]Provider `json:"provider_map,omitempty"`

	// FallbackProvider is used when the preferred provider is unavailable.
	FallbackProvider Provider `json:"fallback_provider,omitempty"`

	// AllowInsecure allows credential access even if security checks fail.
	// This should only be used for development/testing.
	AllowInsecure bool `json:"allow_insecure,omitempty"`

	// InsecureReason documents why AllowInsecure is set (for auditing).
	InsecureReason string `json:"insecure_reason,omitempty"`
}

// LocalPolicy defines security requirements for local workstations.
type LocalPolicy struct {
	// MinSecurityScore is the minimum Posture security score (0-100).
	// Default: 0 (no minimum).
	MinSecurityScore int `json:"min_security_score,omitempty"`

	// RequireEncryption requires disk encryption (FileVault/BitLocker/LUKS).
	RequireEncryption bool `json:"require_encryption,omitempty"`

	// RequireTPM requires TPM or Secure Enclave.
	RequireTPM bool `json:"require_tpm,omitempty"`

	// RequireSecureBoot requires Secure Boot to be enabled.
	RequireSecureBoot bool `json:"require_secure_boot,omitempty"`

	// RequireBiometrics requires biometric authentication to be configured.
	RequireBiometrics bool `json:"require_biometrics,omitempty"`

	// AllowedPlatforms restricts to specific platforms (darwin, windows, linux).
	// Empty means all platforms are allowed.
	AllowedPlatforms []string `json:"allowed_platforms,omitempty"`
}

// CloudPolicy defines security requirements for cloud environments.
type CloudPolicy struct {
	// RequireIAM requires cloud IAM (IRSA, Workload Identity) to be configured.
	RequireIAM bool `json:"require_iam,omitempty"`

	// AWS-specific requirements.
	AWS *AWSPolicy `json:"aws,omitempty"`

	// GCP-specific requirements.
	GCP *GCPPolicy `json:"gcp,omitempty"`

	// Azure-specific requirements.
	Azure *AzurePolicy `json:"azure,omitempty"`
}

// AWSPolicy defines AWS-specific security requirements.
type AWSPolicy struct {
	// RequireIRSA requires IRSA (IAM Roles for Service Accounts).
	RequireIRSA bool `json:"require_irsa,omitempty"`

	// AllowedRoleARNs is a whitelist of allowed IAM role ARNs.
	// Supports wildcards: "arn:aws:iam::123456789:role/my-app-*"
	AllowedRoleARNs []string `json:"allowed_role_arns,omitempty"`

	// AllowedAccountIDs is a whitelist of allowed AWS account IDs.
	AllowedAccountIDs []string `json:"allowed_account_ids,omitempty"`

	// AllowedRegions restricts to specific AWS regions.
	AllowedRegions []string `json:"allowed_regions,omitempty"`

	// RequireIMDSv2 requires IMDSv2 for EC2 metadata.
	RequireIMDSv2 bool `json:"require_imdsv2,omitempty"`
}

// GCPPolicy defines GCP-specific security requirements.
type GCPPolicy struct {
	// RequireWorkloadIdentity requires GKE Workload Identity.
	RequireWorkloadIdentity bool `json:"require_workload_identity,omitempty"`

	// AllowedServiceAccounts is a whitelist of allowed GCP service accounts.
	AllowedServiceAccounts []string `json:"allowed_service_accounts,omitempty"`

	// AllowedProjects is a whitelist of allowed GCP project IDs.
	AllowedProjects []string `json:"allowed_projects,omitempty"`

	// AllowedRegions restricts to specific GCP regions.
	AllowedRegions []string `json:"allowed_regions,omitempty"`
}

// AzurePolicy defines Azure-specific security requirements.
type AzurePolicy struct {
	// RequireWorkloadIdentity requires AKS Workload Identity.
	RequireWorkloadIdentity bool `json:"require_workload_identity,omitempty"`

	// AllowedClientIDs is a whitelist of allowed Azure AD client IDs.
	AllowedClientIDs []string `json:"allowed_client_ids,omitempty"`

	// AllowedTenantIDs is a whitelist of allowed Azure AD tenant IDs.
	AllowedTenantIDs []string `json:"allowed_tenant_ids,omitempty"`

	// AllowedSubscriptions restricts to specific Azure subscriptions.
	AllowedSubscriptions []string `json:"allowed_subscriptions,omitempty"`

	// AllowedRegions restricts to specific Azure regions.
	AllowedRegions []string `json:"allowed_regions,omitempty"`
}

// KubernetesPolicy defines Kubernetes-specific security requirements.
type KubernetesPolicy struct {
	// RequireServiceAccount requires a specific service account.
	RequireServiceAccount bool `json:"require_service_account,omitempty"`

	// AllowedServiceAccounts is a whitelist of allowed service accounts.
	AllowedServiceAccounts []string `json:"allowed_service_accounts,omitempty"`

	// AllowedNamespaces restricts to specific namespaces.
	AllowedNamespaces []string `json:"allowed_namespaces,omitempty"`

	// DeniedNamespaces is a blacklist of namespaces.
	DeniedNamespaces []string `json:"denied_namespaces,omitempty"`

	// RequireNonRoot requires the pod to run as non-root.
	RequireNonRoot bool `json:"require_non_root,omitempty"`

	// RequireReadOnlyRoot requires a read-only root filesystem.
	RequireReadOnlyRoot bool `json:"require_read_only_root,omitempty"`
}

// DefaultPolicy returns a sensible default policy for production use.
func DefaultPolicy() *Policy {
	return &Policy{
		Local: &LocalPolicy{
			MinSecurityScore:  50,
			RequireEncryption: true,
		},
		Cloud: &CloudPolicy{
			RequireIAM: true,
			AWS: &AWSPolicy{
				RequireIRSA: true,
			},
			GCP: &GCPPolicy{
				RequireWorkloadIdentity: true,
			},
			Azure: &AzurePolicy{
				RequireWorkloadIdentity: true,
			},
		},
		Kubernetes: &KubernetesPolicy{
			DeniedNamespaces: []string{"default", "kube-system", "kube-public"},
		},
		ProviderMap: map[Environment]Provider{
			EnvLocal:    ProviderKeyring,
			EnvEKS:      ProviderAWSSecretsManager,
			EnvLambda:   ProviderAWSSecretsManager,
			EnvGKE:      ProviderGCPSecretManager,
			EnvCloudRun: ProviderGCPSecretManager,
			EnvAKS:      ProviderAzureKeyVault,
		},
		FallbackProvider: ProviderEnv,
	}
}

// DevelopmentPolicy returns a permissive policy for development.
func DevelopmentPolicy() *Policy {
	return &Policy{
		Local: &LocalPolicy{
			MinSecurityScore: 0, // No minimum
		},
		Cloud: &CloudPolicy{
			RequireIAM: false, // Allow env vars in dev containers
		},
		ProviderMap: map[Environment]Provider{
			EnvLocal:      ProviderEnv,
			EnvContainer:  ProviderEnv,
			EnvKubernetes: ProviderEnv,
			EnvEKS:        ProviderEnv,
			EnvGKE:        ProviderEnv,
			EnvAKS:        ProviderEnv,
		},
		FallbackProvider: ProviderEnv,
		AllowInsecure:    true,
		InsecureReason:   "Development environment",
	}
}

// StrictPolicy returns a strict policy for high-security environments.
func StrictPolicy() *Policy {
	return &Policy{
		Local: &LocalPolicy{
			MinSecurityScore:  75,
			RequireEncryption: true,
			RequireTPM:        true,
			RequireSecureBoot: true,
		},
		Cloud: &CloudPolicy{
			RequireIAM: true,
			AWS: &AWSPolicy{
				RequireIRSA:   true,
				RequireIMDSv2: true,
			},
			GCP: &GCPPolicy{
				RequireWorkloadIdentity: true,
			},
			Azure: &AzurePolicy{
				RequireWorkloadIdentity: true,
			},
		},
		Kubernetes: &KubernetesPolicy{
			RequireServiceAccount: true,
			RequireNonRoot:        true,
			RequireReadOnlyRoot:   true,
			DeniedNamespaces:      []string{"default", "kube-system", "kube-public"},
		},
		ProviderMap: map[Environment]Provider{
			EnvLocal:    ProviderKeyring,
			EnvEKS:      ProviderAWSSecretsManager,
			EnvLambda:   ProviderAWSSecretsManager,
			EnvGKE:      ProviderGCPSecretManager,
			EnvCloudRun: ProviderGCPSecretManager,
			EnvAKS:      ProviderAzureKeyVault,
		},
		AllowInsecure: false,
	}
}
