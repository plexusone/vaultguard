package vaultguard

import (
	"fmt"
	"os"
	"slices"
	"strings"
	"time"
)

// CheckCloudSecurity performs security checks for cloud environments.
func CheckCloudSecurity(env Environment, policy *CloudPolicy) (*SecurityResult, error) {
	result := &SecurityResult{
		Environment: env,
		Timestamp:   time.Now(),
		Details: SecurityDetails{
			Cloud: &CloudSecurityDetails{},
		},
	}

	// Get environment details
	envDetails := GetEnvironmentDetails()
	cloud := result.Details.Cloud

	// Populate cloud details based on environment
	switch {
	case env.IsAWS():
		cloud.CloudProvider = "aws"
		if envDetails.AWS != nil {
			cloud.Region = envDetails.AWS.Region
			cloud.IAM = &IAMDetails{
				Configured: envDetails.AWS.RoleARN != "",
				RoleARN:    envDetails.AWS.RoleARN,
				TokenPath:  envDetails.AWS.TokenFile,
			}
		}
	case env.IsGCP():
		cloud.CloudProvider = "gcp"
		if envDetails.GCP != nil {
			cloud.Region = envDetails.GCP.Region
			cloud.IAM = &IAMDetails{
				Configured:          envDetails.GCP.ServiceAccountEmail != "",
				ServiceAccountEmail: envDetails.GCP.ServiceAccountEmail,
			}
		}
	case env.IsAzure():
		cloud.CloudProvider = "azure"
		if envDetails.Azure != nil {
			cloud.IAM = &IAMDetails{
				Configured: envDetails.Azure.ClientID != "",
				ClientID:   envDetails.Azure.ClientID,
				TokenPath:  envDetails.Azure.TokenFile,
			}
		}
	}

	// Add Kubernetes details if applicable
	if env.IsKubernetes() && envDetails.Kubernetes != nil {
		cloud.Kubernetes = envDetails.Kubernetes
	}

	// If no policy, return assessment without validation
	if policy == nil {
		result.Passed = true
		result.Score = calculateCloudScore(cloud)
		result.Level = scoreToLevel(result.Score)
		result.Message = fmt.Sprintf("Cloud environment: %s (%s)", env, cloud.CloudProvider)
		return result, nil
	}

	// Validate against policy
	var violations []string

	// Check IAM requirement
	if policy.RequireIAM {
		if cloud.IAM == nil || !cloud.IAM.Configured {
			violations = append(violations, "cloud IAM required but not configured")
			result.Recommendations = append(result.Recommendations,
				getIAMRecommendation(env))
		}
	}

	// AWS-specific checks
	if env.IsAWS() && policy.AWS != nil {
		awsViolations := checkAWSPolicy(envDetails.AWS, policy.AWS)
		violations = append(violations, awsViolations...)
	}

	// GCP-specific checks
	if env.IsGCP() && policy.GCP != nil {
		gcpViolations := checkGCPPolicy(envDetails.GCP, policy.GCP)
		violations = append(violations, gcpViolations...)
	}

	// Azure-specific checks
	if env.IsAzure() && policy.Azure != nil {
		azureViolations := checkAzurePolicy(envDetails.Azure, policy.Azure)
		violations = append(violations, azureViolations...)
	}

	// Set result
	result.Score = calculateCloudScore(cloud)
	result.Level = scoreToLevel(result.Score)

	if len(violations) > 0 {
		result.Passed = false
		result.Message = fmt.Sprintf("Cloud security policy violations: %v", violations)
		return result, &SecurityError{
			Err:             ErrSecurityCheckFailed,
			Check:           "cloud security policy",
			Environment:     env,
			Recommendations: result.Recommendations,
		}
	}

	result.Passed = true
	result.Message = fmt.Sprintf("Cloud security checks passed (%s)", cloud.CloudProvider)
	return result, nil
}

// CheckKubernetesSecurity performs Kubernetes-specific security checks.
func CheckKubernetesSecurity(env Environment, policy *KubernetesPolicy) (*SecurityResult, error) {
	result := &SecurityResult{
		Environment: env,
		Timestamp:   time.Now(),
	}

	envDetails := GetEnvironmentDetails()
	k8s := envDetails.Kubernetes

	if k8s == nil {
		result.Passed = false
		result.Message = "Not running in Kubernetes"
		return result, &SecurityError{
			Err:         ErrEnvironmentNotSupported,
			Check:       "kubernetes detection",
			Environment: env,
		}
	}

	// If no policy, return assessment
	if policy == nil {
		result.Passed = true
		result.Score = 75 // Default score for K8s
		result.Level = SecurityHigh
		result.Message = fmt.Sprintf("Kubernetes: namespace=%s, sa=%s",
			k8s.Namespace, k8s.ServiceAccount)
		return result, nil
	}

	// Validate against policy
	var violations []string

	// Check namespace restrictions
	if len(policy.AllowedNamespaces) > 0 {
		if !slices.Contains(policy.AllowedNamespaces, k8s.Namespace) {
			violations = append(violations, fmt.Sprintf(
				"namespace %q not in allowed list", k8s.Namespace,
			))
		}
	}

	if len(policy.DeniedNamespaces) > 0 {
		if slices.Contains(policy.DeniedNamespaces, k8s.Namespace) {
			violations = append(violations, fmt.Sprintf(
				"namespace %q is in denied list", k8s.Namespace,
			))
		}
	}

	// Check service account restrictions
	if policy.RequireServiceAccount && k8s.ServiceAccount == "" {
		violations = append(violations, "service account required but not detected")
	}

	if len(policy.AllowedServiceAccounts) > 0 && k8s.ServiceAccount != "" {
		if !slices.Contains(policy.AllowedServiceAccounts, k8s.ServiceAccount) {
			violations = append(violations, fmt.Sprintf(
				"service account %q not in allowed list", k8s.ServiceAccount,
			))
		}
	}

	// Check security context (best effort - may not be detectable from inside container)
	if policy.RequireNonRoot {
		if os.Getuid() == 0 {
			violations = append(violations, "running as root but non-root required")
		}
	}

	// Set result
	if len(violations) > 0 {
		result.Passed = false
		result.Score = 25
		result.Level = SecurityLow
		result.Message = fmt.Sprintf("Kubernetes policy violations: %v", violations)
		return result, &SecurityError{
			Err:         ErrSecurityCheckFailed,
			Check:       "kubernetes security policy",
			Environment: env,
		}
	}

	result.Passed = true
	result.Score = 75
	result.Level = SecurityHigh
	result.Message = "Kubernetes security checks passed"
	return result, nil
}

// checkAWSPolicy validates AWS-specific policy requirements.
func checkAWSPolicy(details *AWSDetails, policy *AWSPolicy) []string {
	var violations []string

	if details == nil {
		if policy.RequireIRSA {
			violations = append(violations, "IRSA required but AWS details not available")
		}
		return violations
	}

	// Check IRSA
	if policy.RequireIRSA {
		if details.RoleARN == "" || details.TokenFile == "" {
			violations = append(violations, "IRSA required but not configured")
		}
	}

	// Check allowed role ARNs
	if len(policy.AllowedRoleARNs) > 0 && details.RoleARN != "" {
		allowed := false
		for _, pattern := range policy.AllowedRoleARNs {
			if matchARNPattern(details.RoleARN, pattern) {
				allowed = true
				break
			}
		}
		if !allowed {
			violations = append(violations, fmt.Sprintf(
				"role ARN %q not in allowed list", details.RoleARN,
			))
		}
	}

	// Check allowed account IDs
	if len(policy.AllowedAccountIDs) > 0 && details.AccountID != "" {
		if !slices.Contains(policy.AllowedAccountIDs, details.AccountID) {
			violations = append(violations, fmt.Sprintf(
				"account ID %q not in allowed list", details.AccountID,
			))
		}
	}

	// Check allowed regions
	if len(policy.AllowedRegions) > 0 && details.Region != "" {
		if !slices.Contains(policy.AllowedRegions, details.Region) {
			violations = append(violations, fmt.Sprintf(
				"region %q not in allowed list", details.Region,
			))
		}
	}

	return violations
}

// checkGCPPolicy validates GCP-specific policy requirements.
func checkGCPPolicy(details *GCPDetails, policy *GCPPolicy) []string {
	var violations []string

	if details == nil {
		if policy.RequireWorkloadIdentity {
			violations = append(violations, "Workload Identity required but GCP details not available")
		}
		return violations
	}

	// Check Workload Identity
	if policy.RequireWorkloadIdentity && details.ServiceAccountEmail == "" {
		violations = append(violations, "Workload Identity required but not configured")
	}

	// Check allowed service accounts
	if len(policy.AllowedServiceAccounts) > 0 && details.ServiceAccountEmail != "" {
		if !slices.Contains(policy.AllowedServiceAccounts, details.ServiceAccountEmail) {
			violations = append(violations, fmt.Sprintf(
				"service account %q not in allowed list", details.ServiceAccountEmail,
			))
		}
	}

	// Check allowed projects
	if len(policy.AllowedProjects) > 0 && details.ProjectID != "" {
		if !slices.Contains(policy.AllowedProjects, details.ProjectID) {
			violations = append(violations, fmt.Sprintf(
				"project %q not in allowed list", details.ProjectID,
			))
		}
	}

	// Check allowed regions
	if len(policy.AllowedRegions) > 0 && details.Region != "" {
		if !slices.Contains(policy.AllowedRegions, details.Region) {
			violations = append(violations, fmt.Sprintf(
				"region %q not in allowed list", details.Region,
			))
		}
	}

	return violations
}

// checkAzurePolicy validates Azure-specific policy requirements.
func checkAzurePolicy(details *AzureDetails, policy *AzurePolicy) []string {
	var violations []string

	if details == nil {
		if policy.RequireWorkloadIdentity {
			violations = append(violations, "Workload Identity required but Azure details not available")
		}
		return violations
	}

	// Check Workload Identity
	if policy.RequireWorkloadIdentity {
		if details.ClientID == "" || details.TenantID == "" || details.TokenFile == "" {
			violations = append(violations, "Workload Identity required but not fully configured")
		}
	}

	// Check allowed client IDs
	if len(policy.AllowedClientIDs) > 0 && details.ClientID != "" {
		if !slices.Contains(policy.AllowedClientIDs, details.ClientID) {
			violations = append(violations, fmt.Sprintf(
				"client ID %q not in allowed list", details.ClientID,
			))
		}
	}

	// Check allowed tenant IDs
	if len(policy.AllowedTenantIDs) > 0 && details.TenantID != "" {
		if !slices.Contains(policy.AllowedTenantIDs, details.TenantID) {
			violations = append(violations, fmt.Sprintf(
				"tenant ID %q not in allowed list", details.TenantID,
			))
		}
	}

	// Check allowed subscriptions
	if len(policy.AllowedSubscriptions) > 0 && details.Subscription != "" {
		if !slices.Contains(policy.AllowedSubscriptions, details.Subscription) {
			violations = append(violations, fmt.Sprintf(
				"subscription %q not in allowed list", details.Subscription,
			))
		}
	}

	return violations
}

// matchARNPattern checks if an ARN matches a pattern (supports * wildcard).
func matchARNPattern(arn, pattern string) bool {
	// Simple wildcard matching
	if pattern == "*" {
		return true
	}
	if strings.HasSuffix(pattern, "*") {
		prefix := strings.TrimSuffix(pattern, "*")
		return strings.HasPrefix(arn, prefix)
	}
	return arn == pattern
}

// calculateCloudScore calculates a security score for cloud environments.
func calculateCloudScore(cloud *CloudSecurityDetails) int {
	if cloud == nil {
		return 0
	}

	score := 50 // Base score for being in cloud

	// IAM configured
	if cloud.IAM != nil && cloud.IAM.Configured {
		score += 25
	}

	// Kubernetes with namespace isolation
	if cloud.Kubernetes != nil {
		score += 15
		if cloud.Kubernetes.Namespace != "default" {
			score += 10
		}
	}

	if score > 100 {
		score = 100
	}

	return score
}

// getIAMRecommendation returns cloud-specific IAM setup guidance.
func getIAMRecommendation(env Environment) string {
	switch env {
	case EnvEKS:
		return "Configure IRSA: https://docs.aws.amazon.com/eks/latest/userguide/iam-roles-for-service-accounts.html"
	case EnvGKE:
		return "Configure Workload Identity: https://cloud.google.com/kubernetes-engine/docs/how-to/workload-identity"
	case EnvAKS:
		return "Configure Workload Identity: https://learn.microsoft.com/en-us/azure/aks/workload-identity-overview"
	default:
		return "Configure cloud IAM for your environment"
	}
}
