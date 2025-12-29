// Package vaultguard provides security-gated credential access by combining
// Posture (security posture assessment) with OmniVault (secret management).
// It supports both local workstation deployments and cloud environments like
// AWS EKS, GCP GKE, and Azure AKS.
package vaultguard

import "time"

// Environment represents the deployment context.
type Environment string

const (
	// EnvUnknown is an undetected environment.
	EnvUnknown Environment = "unknown"

	// EnvLocal is a developer workstation (macOS, Windows, Linux desktop).
	EnvLocal Environment = "local"

	// EnvContainer is a generic container without cloud IAM.
	EnvContainer Environment = "container"

	// EnvKubernetes is Kubernetes without cloud-specific IAM.
	EnvKubernetes Environment = "kubernetes"

	// EnvEKS is AWS EKS with IRSA (IAM Roles for Service Accounts).
	EnvEKS Environment = "eks"

	// EnvGKE is GCP GKE with Workload Identity.
	EnvGKE Environment = "gke"

	// EnvAKS is Azure AKS with Workload Identity.
	EnvAKS Environment = "aks"

	// EnvLambda is AWS Lambda.
	EnvLambda Environment = "lambda"

	// EnvCloudRun is GCP Cloud Run.
	EnvCloudRun Environment = "cloudrun"

	// EnvAzureFunc is Azure Functions.
	EnvAzureFunc Environment = "azurefunc"
)

// String returns the string representation of the environment.
func (e Environment) String() string {
	return string(e)
}

// IsLocal returns true if this is a local workstation environment.
func (e Environment) IsLocal() bool {
	return e == EnvLocal
}

// IsCloud returns true if this is a cloud environment.
func (e Environment) IsCloud() bool {
	switch e {
	case EnvEKS, EnvGKE, EnvAKS, EnvLambda, EnvCloudRun, EnvAzureFunc:
		return true
	default:
		return false
	}
}

// IsKubernetes returns true if this is a Kubernetes environment.
func (e Environment) IsKubernetes() bool {
	switch e {
	case EnvKubernetes, EnvEKS, EnvGKE, EnvAKS:
		return true
	default:
		return false
	}
}

// IsAWS returns true if this is an AWS environment.
func (e Environment) IsAWS() bool {
	return e == EnvEKS || e == EnvLambda
}

// IsGCP returns true if this is a GCP environment.
func (e Environment) IsGCP() bool {
	return e == EnvGKE || e == EnvCloudRun
}

// IsAzure returns true if this is an Azure environment.
func (e Environment) IsAzure() bool {
	return e == EnvAKS || e == EnvAzureFunc
}

// Provider represents a secret provider type.
type Provider string

const (
	// ProviderEnv reads from environment variables.
	ProviderEnv Provider = "env"

	// ProviderFile reads from files.
	ProviderFile Provider = "file"

	// ProviderKeyring uses OS keyring (macOS Keychain, Windows Credential Manager).
	ProviderKeyring Provider = "keyring"

	// ProviderAWSSecretsManager uses AWS Secrets Manager.
	ProviderAWSSecretsManager Provider = "aws-sm"

	// ProviderAWSParameterStore uses AWS Systems Manager Parameter Store.
	ProviderAWSParameterStore Provider = "aws-ssm"

	// ProviderGCPSecretManager uses GCP Secret Manager.
	ProviderGCPSecretManager Provider = "gcp-sm"

	// ProviderAzureKeyVault uses Azure Key Vault.
	ProviderAzureKeyVault Provider = "azure-kv"

	// ProviderK8sSecret uses Kubernetes Secrets.
	ProviderK8sSecret Provider = "k8s"

	// ProviderVault uses HashiCorp Vault.
	ProviderVault Provider = "vault"
)

// String returns the string representation of the provider.
func (p Provider) String() string {
	return string(p)
}

// SecurityLevel represents the overall security assessment.
type SecurityLevel string

const (
	// SecurityCritical means critical security issues detected.
	SecurityCritical SecurityLevel = "critical"

	// SecurityLow means security posture is low.
	SecurityLow SecurityLevel = "low"

	// SecurityMedium means security posture is acceptable.
	SecurityMedium SecurityLevel = "medium"

	// SecurityHigh means security posture is good.
	SecurityHigh SecurityLevel = "high"

	// SecurityExcellent means all security features are enabled.
	SecurityExcellent SecurityLevel = "excellent"
)

// SecurityResult contains the security assessment results.
type SecurityResult struct {
	// Environment detected.
	Environment Environment `json:"environment"`

	// Level is the overall security level.
	Level SecurityLevel `json:"level"`

	// Score is the numeric security score (0-100).
	Score int `json:"score"`

	// Passed indicates if the security policy passed.
	Passed bool `json:"passed"`

	// Message provides a human-readable summary.
	Message string `json:"message"`

	// Details contains environment-specific security details.
	Details SecurityDetails `json:"details"`

	// Recommendations for improving security.
	Recommendations []string `json:"recommendations,omitempty"`

	// Timestamp of the assessment.
	Timestamp time.Time `json:"timestamp"`
}

// SecurityDetails contains environment-specific security information.
type SecurityDetails struct {
	// Local security details (populated for local environments).
	Local *LocalSecurityDetails `json:"local,omitempty"`

	// Cloud security details (populated for cloud environments).
	Cloud *CloudSecurityDetails `json:"cloud,omitempty"`
}

// LocalSecurityDetails contains workstation security information.
type LocalSecurityDetails struct {
	// Platform (darwin, windows, linux).
	Platform string `json:"platform"`

	// TPMPresent indicates if TPM/Secure Enclave is available.
	TPMPresent bool `json:"tpm_present"`

	// TPMEnabled indicates if TPM/Secure Enclave is enabled.
	TPMEnabled bool `json:"tpm_enabled"`

	// TPMType is the type of security chip (tpm, secure_enclave).
	TPMType string `json:"tpm_type,omitempty"`

	// DiskEncrypted indicates if disk encryption is enabled.
	DiskEncrypted bool `json:"disk_encrypted"`

	// EncryptionType (filevault, bitlocker, luks).
	EncryptionType string `json:"encryption_type,omitempty"`

	// SecureBootEnabled indicates if Secure Boot is enabled.
	SecureBootEnabled bool `json:"secure_boot_enabled"`

	// BiometricsAvailable indicates if biometric auth is available.
	BiometricsAvailable bool `json:"biometrics_available"`

	// BiometricsConfigured indicates if biometrics are set up.
	BiometricsConfigured bool `json:"biometrics_configured"`
}

// CloudSecurityDetails contains cloud environment security information.
type CloudSecurityDetails struct {
	// CloudProvider (aws, gcp, azure).
	CloudProvider string `json:"cloud_provider"`

	// Region where the workload is running.
	Region string `json:"region,omitempty"`

	// IAM identity details.
	IAM *IAMDetails `json:"iam,omitempty"`

	// Kubernetes details (if applicable).
	Kubernetes *KubernetesDetails `json:"kubernetes,omitempty"`
}

// IAMDetails contains cloud IAM information.
type IAMDetails struct {
	// Configured indicates if cloud IAM is configured.
	Configured bool `json:"configured"`

	// RoleARN for AWS IRSA.
	RoleARN string `json:"role_arn,omitempty"`

	// ServiceAccountEmail for GCP Workload Identity.
	ServiceAccountEmail string `json:"service_account_email,omitempty"`

	// ClientID for Azure Workload Identity.
	ClientID string `json:"client_id,omitempty"`

	// TokenPath is the path to the projected token.
	TokenPath string `json:"token_path,omitempty"`
}

// KubernetesDetails contains Kubernetes environment information.
type KubernetesDetails struct {
	// InCluster indicates if running inside a Kubernetes cluster.
	InCluster bool `json:"in_cluster"`

	// Namespace is the current namespace.
	Namespace string `json:"namespace,omitempty"`

	// ServiceAccount is the service account name.
	ServiceAccount string `json:"service_account,omitempty"`

	// PodName is the current pod name.
	PodName string `json:"pod_name,omitempty"`
}
