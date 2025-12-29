package vaultguard

import (
	"os"
	"runtime"
	"strings"
)

// DetectEnvironment automatically detects the current deployment environment.
// It checks for cloud-specific environment variables and Kubernetes markers
// to determine where the code is running.
func DetectEnvironment() Environment {
	// Check for AWS EKS with IRSA first (most specific)
	if isEKS() {
		return EnvEKS
	}

	// Check for AWS Lambda
	if isLambda() {
		return EnvLambda
	}

	// Check for GCP GKE with Workload Identity
	if isGKE() {
		return EnvGKE
	}

	// Check for GCP Cloud Run
	if isCloudRun() {
		return EnvCloudRun
	}

	// Check for Azure AKS with Workload Identity
	if isAKS() {
		return EnvAKS
	}

	// Check for Azure Functions
	if isAzureFunc() {
		return EnvAzureFunc
	}

	// Check for generic Kubernetes
	if isKubernetes() {
		return EnvKubernetes
	}

	// Check for generic container
	if isContainer() {
		return EnvContainer
	}

	// Default to local workstation
	return EnvLocal
}

// isEKS checks if running on AWS EKS with IRSA.
func isEKS() bool {
	// IRSA sets AWS_ROLE_ARN and AWS_WEB_IDENTITY_TOKEN_FILE
	roleARN := os.Getenv("AWS_ROLE_ARN")
	tokenFile := os.Getenv("AWS_WEB_IDENTITY_TOKEN_FILE")

	if roleARN != "" && tokenFile != "" {
		// Verify token file exists
		if _, err := os.Stat(tokenFile); err == nil {
			return true
		}
	}

	return false
}

// isLambda checks if running on AWS Lambda.
func isLambda() bool {
	// Lambda sets AWS_LAMBDA_FUNCTION_NAME
	return os.Getenv("AWS_LAMBDA_FUNCTION_NAME") != "" ||
		os.Getenv("AWS_EXECUTION_ENV") != ""
}

// isGKE checks if running on GCP GKE with Workload Identity.
func isGKE() bool {
	// GKE Workload Identity sets these
	if os.Getenv("GOOGLE_APPLICATION_CREDENTIALS") != "" {
		// Check if it's the projected token path
		creds := os.Getenv("GOOGLE_APPLICATION_CREDENTIALS")
		if strings.Contains(creds, "/var/run/secrets/") {
			return true
		}
	}

	// Check for GKE metadata
	if os.Getenv("GKE_CLUSTER_NAME") != "" {
		return true
	}

	// Check for Workload Identity token
	// This is a standard GCP Workload Identity token path, not a credential
	tokenPath := "/var/run/secrets/tokens/gcp-token" //nolint:gosec // G101: false positive - this is a file path, not a credential
	if _, err := os.Stat(tokenPath); err == nil {
		return true
	}

	return false
}

// isCloudRun checks if running on GCP Cloud Run.
func isCloudRun() bool {
	// Cloud Run sets K_SERVICE and K_REVISION
	return os.Getenv("K_SERVICE") != "" && os.Getenv("K_REVISION") != ""
}

// isAKS checks if running on Azure AKS with Workload Identity.
func isAKS() bool {
	// Azure Workload Identity sets these
	clientID := os.Getenv("AZURE_CLIENT_ID")
	tenantID := os.Getenv("AZURE_TENANT_ID")
	tokenFile := os.Getenv("AZURE_FEDERATED_TOKEN_FILE")

	if clientID != "" && tenantID != "" && tokenFile != "" {
		// Verify token file exists
		if _, err := os.Stat(tokenFile); err == nil {
			return true
		}
	}

	return false
}

// isAzureFunc checks if running on Azure Functions.
func isAzureFunc() bool {
	return os.Getenv("FUNCTIONS_WORKER_RUNTIME") != "" ||
		os.Getenv("AZURE_FUNCTIONS_ENVIRONMENT") != ""
}

// isKubernetes checks if running inside a Kubernetes cluster.
func isKubernetes() bool {
	// Check for Kubernetes service account token
	// This is a standard Kubernetes service account token path, not a credential
	tokenPath := "/var/run/secrets/kubernetes.io/serviceaccount/token" //nolint:gosec // G101: false positive - this is a file path, not a credential
	if _, err := os.Stat(tokenPath); err == nil {
		return true
	}

	// Check for Kubernetes environment variables
	if os.Getenv("KUBERNETES_SERVICE_HOST") != "" {
		return true
	}

	return false
}

// isContainer checks if running inside a container.
func isContainer() bool {
	// Check for Docker
	if _, err := os.Stat("/.dockerenv"); err == nil {
		return true
	}

	// Check for containerd/CRI-O via cgroup
	if data, err := os.ReadFile("/proc/1/cgroup"); err == nil {
		content := string(data)
		if strings.Contains(content, "docker") ||
			strings.Contains(content, "kubepods") ||
			strings.Contains(content, "containerd") {
			return true
		}
	}

	// Check for container runtime environment variables
	if os.Getenv("CONTAINER") != "" || os.Getenv("container") != "" {
		return true
	}

	return false
}

// GetEnvironmentDetails returns detailed information about the current environment.
func GetEnvironmentDetails() *EnvironmentDetails {
	env := DetectEnvironment()

	details := &EnvironmentDetails{
		Environment: env,
		Platform:    runtime.GOOS,
		Arch:        runtime.GOARCH,
	}

	switch {
	case env.IsAWS():
		details.AWS = getAWSDetails()
	case env.IsGCP():
		details.GCP = getGCPDetails()
	case env.IsAzure():
		details.Azure = getAzureDetails()
	}

	if env.IsKubernetes() {
		details.Kubernetes = getKubernetesDetails()
	}

	return details
}

// EnvironmentDetails contains detailed environment information.
type EnvironmentDetails struct {
	Environment Environment        `json:"environment"`
	Platform    string             `json:"platform"`
	Arch        string             `json:"arch"`
	AWS         *AWSDetails        `json:"aws,omitempty"`
	GCP         *GCPDetails        `json:"gcp,omitempty"`
	Azure       *AzureDetails      `json:"azure,omitempty"`
	Kubernetes  *KubernetesDetails `json:"kubernetes,omitempty"`
}

// AWSDetails contains AWS-specific environment details.
type AWSDetails struct {
	RoleARN      string `json:"role_arn,omitempty"`
	Region       string `json:"region,omitempty"`
	AccountID    string `json:"account_id,omitempty"`
	TokenFile    string `json:"token_file,omitempty"`
	FunctionName string `json:"function_name,omitempty"`
	ExecutionEnv string `json:"execution_env,omitempty"`
}

// GCPDetails contains GCP-specific environment details.
type GCPDetails struct {
	ProjectID           string `json:"project_id,omitempty"`
	Region              string `json:"region,omitempty"`
	ServiceAccountEmail string `json:"service_account_email,omitempty"`
	ClusterName         string `json:"cluster_name,omitempty"`
	ServiceName         string `json:"service_name,omitempty"`
}

// AzureDetails contains Azure-specific environment details.
type AzureDetails struct {
	ClientID      string `json:"client_id,omitempty"`
	TenantID      string `json:"tenant_id,omitempty"`
	TokenFile     string `json:"token_file,omitempty"`
	Subscription  string `json:"subscription,omitempty"`
	ResourceGroup string `json:"resource_group,omitempty"`
}

func getAWSDetails() *AWSDetails {
	details := &AWSDetails{
		RoleARN:      os.Getenv("AWS_ROLE_ARN"),
		Region:       os.Getenv("AWS_REGION"),
		TokenFile:    os.Getenv("AWS_WEB_IDENTITY_TOKEN_FILE"),
		FunctionName: os.Getenv("AWS_LAMBDA_FUNCTION_NAME"),
		ExecutionEnv: os.Getenv("AWS_EXECUTION_ENV"),
	}

	// Try to get region from default region if AWS_REGION is not set
	if details.Region == "" {
		details.Region = os.Getenv("AWS_DEFAULT_REGION")
	}

	// Extract account ID from role ARN
	if details.RoleARN != "" {
		parts := strings.Split(details.RoleARN, ":")
		if len(parts) >= 5 {
			details.AccountID = parts[4]
		}
	}

	return details
}

func getGCPDetails() *GCPDetails {
	details := &GCPDetails{
		ProjectID:   os.Getenv("GOOGLE_CLOUD_PROJECT"),
		Region:      os.Getenv("GOOGLE_CLOUD_REGION"),
		ClusterName: os.Getenv("GKE_CLUSTER_NAME"),
		ServiceName: os.Getenv("K_SERVICE"),
	}

	// Alternative project ID env var
	if details.ProjectID == "" {
		details.ProjectID = os.Getenv("GCLOUD_PROJECT")
	}

	return details
}

func getAzureDetails() *AzureDetails {
	return &AzureDetails{
		ClientID:      os.Getenv("AZURE_CLIENT_ID"),
		TenantID:      os.Getenv("AZURE_TENANT_ID"),
		TokenFile:     os.Getenv("AZURE_FEDERATED_TOKEN_FILE"),
		Subscription:  os.Getenv("AZURE_SUBSCRIPTION_ID"),
		ResourceGroup: os.Getenv("AZURE_RESOURCE_GROUP"),
	}
}

func getKubernetesDetails() *KubernetesDetails {
	details := &KubernetesDetails{
		InCluster: true,
	}

	// Read namespace from mounted file
	if data, err := os.ReadFile("/var/run/secrets/kubernetes.io/serviceaccount/namespace"); err == nil {
		details.Namespace = strings.TrimSpace(string(data))
	}

	// Get pod name from environment (commonly set by downward API)
	details.PodName = os.Getenv("POD_NAME")
	if details.PodName == "" {
		details.PodName = os.Getenv("HOSTNAME")
	}

	// Get service account from environment (commonly set by downward API)
	details.ServiceAccount = os.Getenv("SERVICE_ACCOUNT")

	return details
}
