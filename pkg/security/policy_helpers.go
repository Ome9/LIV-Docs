// Helper functions for security policy management

package security

import (
	"crypto/sha256"
	"encoding/hex"
	"fmt"
	"regexp"
	"time"

	"github.com/liv-format/liv/pkg/core"
)

// validatePolicy validates a system security policy
func (pm *PolicyManager) validatePolicy(policy *SystemSecurityPolicy) error {
	if policy == nil {
		return fmt.Errorf("policy cannot be nil")
	}

	// Validate basic fields
	if policy.ID == "" {
		return fmt.Errorf("policy ID cannot be empty")
	}

	if policy.Name == "" {
		return fmt.Errorf("policy name cannot be empty")
	}

	// Validate ID format (alphanumeric with hyphens and underscores)
	if !regexp.MustCompile(`^[a-zA-Z0-9_-]+$`).MatchString(policy.ID) {
		return fmt.Errorf("policy ID contains invalid characters")
	}

	// Validate core security policy
	if policy.SecurityPolicy == nil {
		return fmt.Errorf("core security policy cannot be nil")
	}

	if err := pm.validateCoreSecurityPolicy(policy.SecurityPolicy); err != nil {
		return fmt.Errorf("core security policy validation failed: %w", err)
	}

	// Validate administrative controls
	if policy.AdminControls != nil {
		if err := pm.validateAdminControls(policy.AdminControls); err != nil {
			return fmt.Errorf("admin controls validation failed: %w", err)
		}
	}

	// Validate resource limits
	if policy.ResourceLimits != nil {
		if err := pm.validateResourceLimits(policy.ResourceLimits); err != nil {
			return fmt.Errorf("resource limits validation failed: %w", err)
		}
	}

	// Validate compliance settings
	if policy.ComplianceSettings != nil {
		if err := pm.validateComplianceSettings(policy.ComplianceSettings); err != nil {
			return fmt.Errorf("compliance settings validation failed: %w", err)
		}
	}

	return nil
}

// validateCoreSecurityPolicy validates the core security policy
func (pm *PolicyManager) validateCoreSecurityPolicy(policy *core.SecurityPolicy) error {
	if policy.WASMPermissions == nil {
		return fmt.Errorf("WASM permissions cannot be nil")
	}

	// Validate WASM permissions
	if policy.WASMPermissions.MemoryLimit < 1024 || policy.WASMPermissions.MemoryLimit > 134217728 {
		return fmt.Errorf("WASM memory limit must be between 1KB and 128MB")
	}

	if policy.WASMPermissions.CPUTimeLimit < 100 || policy.WASMPermissions.CPUTimeLimit > 30000 {
		return fmt.Errorf("WASM CPU time limit must be between 100ms and 30s")
	}

	// Validate JS permissions
	if policy.JSPermissions == nil {
		return fmt.Errorf("JS permissions cannot be nil")
	}

	validExecutionModes := []string{"none", "sandboxed", "trusted"}
	if !contains(validExecutionModes, policy.JSPermissions.ExecutionMode) {
		return fmt.Errorf("invalid JS execution mode: %s", policy.JSPermissions.ExecutionMode)
	}

	validDOMAccess := []string{"none", "read", "write"}
	if !contains(validDOMAccess, policy.JSPermissions.DOMAccess) {
		return fmt.Errorf("invalid DOM access mode: %s", policy.JSPermissions.DOMAccess)
	}

	return nil
}

// validateAdminControls validates administrative controls
func (pm *PolicyManager) validateAdminControls(controls *AdminControls) error {
	if controls.MaxDocumentSize < 0 {
		return fmt.Errorf("max document size cannot be negative")
	}

	if controls.MaxWASMModules < 0 {
		return fmt.Errorf("max WASM modules cannot be negative")
	}

	if controls.QuarantineDuration < 0 {
		return fmt.Errorf("quarantine duration cannot be negative")
	}

	// Validate file types
	for _, fileType := range controls.AllowedFileTypes {
		if !regexp.MustCompile(`^[a-zA-Z0-9/.-]+$`).MatchString(fileType) {
			return fmt.Errorf("invalid file type format: %s", fileType)
		}
	}

	return nil
}

// validateResourceLimits validates resource limits
func (pm *PolicyManager) validateResourceLimits(limits *ResourceLimits) error {
	if limits.MaxConcurrentDocuments < 0 {
		return fmt.Errorf("max concurrent documents cannot be negative")
	}

	if limits.MaxMemoryPerDocument < 0 {
		return fmt.Errorf("max memory per document cannot be negative")
	}

	if limits.MaxCPUTimePerDocument < 0 {
		return fmt.Errorf("max CPU time per document cannot be negative")
	}

	if limits.DocumentTimeoutSeconds < 0 {
		return fmt.Errorf("document timeout cannot be negative")
	}

	return nil
}

// validateComplianceSettings validates compliance settings
func (pm *PolicyManager) validateComplianceSettings(settings *ComplianceSettings) error {
	if settings.DataRetentionDays < 0 {
		return fmt.Errorf("data retention days cannot be negative")
	}

	validClassifications := []string{"public", "internal", "confidential", "restricted"}
	if settings.DataClassification != "" && !contains(validClassifications, settings.DataClassification) {
		return fmt.Errorf("invalid data classification: %s", settings.DataClassification)
	}

	return nil
}

// setupPolicyInheritance sets up policy inheritance relationships
func (pm *PolicyManager) setupPolicyInheritance(policy *SystemSecurityPolicy) error {
	if policy.ParentPolicy == "" {
		return nil
	}

	// Check if parent policy exists
	parentPolicy, exists := pm.policies[policy.ParentPolicy]
	if !exists {
		return fmt.Errorf("parent policy %s not found", policy.ParentPolicy)
	}

	// Check for circular inheritance
	if err := pm.checkCircularInheritance(policy.ID, policy.ParentPolicy); err != nil {
		return fmt.Errorf("circular inheritance detected: %w", err)
	}

	// Check inheritance depth
	depth := pm.calculateInheritanceDepth(policy.ParentPolicy)
	if depth >= pm.config.MaxPolicyDepth {
		return fmt.Errorf("inheritance depth exceeds maximum allowed depth of %d", pm.config.MaxPolicyDepth)
	}

	// Add to parent's children
	if parentPolicy.ChildPolicies == nil {
		parentPolicy.ChildPolicies = []string{}
	}
	parentPolicy.ChildPolicies = append(parentPolicy.ChildPolicies, policy.ID)

	return nil
}

// applyPolicyInheritance applies inheritance to a policy
func (pm *PolicyManager) applyPolicyInheritance(policy *SystemSecurityPolicy) (*SystemSecurityPolicy, error) {
	if policy.ParentPolicy == "" {
		return policy, nil
	}

	parentPolicy, exists := pm.policies[policy.ParentPolicy]
	if !exists {
		return policy, fmt.Errorf("parent policy %s not found", policy.ParentPolicy)
	}

	// Create a copy of the policy
	inheritedPolicy := *policy

	// Apply inheritance rules (child overrides parent)
	if policy.SecurityPolicy.WASMPermissions == nil && parentPolicy.SecurityPolicy.WASMPermissions != nil {
		inheritedPolicy.SecurityPolicy.WASMPermissions = parentPolicy.SecurityPolicy.WASMPermissions
	}

	if policy.SecurityPolicy.JSPermissions == nil && parentPolicy.SecurityPolicy.JSPermissions != nil {
		inheritedPolicy.SecurityPolicy.JSPermissions = parentPolicy.SecurityPolicy.JSPermissions
	}

	if policy.AdminControls == nil && parentPolicy.AdminControls != nil {
		inheritedPolicy.AdminControls = parentPolicy.AdminControls
	}

	if policy.ResourceLimits == nil && parentPolicy.ResourceLimits != nil {
		inheritedPolicy.ResourceLimits = parentPolicy.ResourceLimits
	}

	return &inheritedPolicy, nil
}

// checkCircularInheritance checks for circular inheritance
func (pm *PolicyManager) checkCircularInheritance(policyID, parentID string) error {
	visited := make(map[string]bool)
	current := parentID

	for current != "" {
		if visited[current] {
			return fmt.Errorf("circular inheritance detected")
		}

		if current == policyID {
			return fmt.Errorf("circular inheritance detected")
		}

		visited[current] = true

		if policy, exists := pm.policies[current]; exists {
			current = policy.ParentPolicy
		} else {
			break
		}
	}

	return nil
}

// calculateInheritanceDepth calculates the inheritance depth
func (pm *PolicyManager) calculateInheritanceDepth(policyID string) int {
	depth := 0
	current := policyID

	for current != "" {
		if policy, exists := pm.policies[current]; exists {
			current = policy.ParentPolicy
			depth++
		} else {
			break
		}
	}

	return depth
}

// removeFromParentPolicy removes a child policy from its parent
func (pm *PolicyManager) removeFromParentPolicy(parentID, childID string) {
	if parentPolicy, exists := pm.policies[parentID]; exists {
		for i, child := range parentPolicy.ChildPolicies {
			if child == childID {
				parentPolicy.ChildPolicies = append(parentPolicy.ChildPolicies[:i], parentPolicy.ChildPolicies[i+1:]...)
				break
			}
		}
	}
}

// createDefaultPolicy creates a default security policy
func (pm *PolicyManager) createDefaultPolicy(policyID string) *SystemSecurityPolicy {
	return &SystemSecurityPolicy{
		SecurityPolicy: &core.SecurityPolicy{
			WASMPermissions: &core.WASMPermissions{
				MemoryLimit:     16 * 1024 * 1024, // 16MB
				AllowedImports:  []string{"console"},
				CPUTimeLimit:    5000, // 5 seconds
				AllowNetworking: false,
				AllowFileSystem: false,
			},
			JSPermissions: &core.JSPermissions{
				ExecutionMode: "sandboxed",
				AllowedAPIs:   []string{"console"},
				DOMAccess:     "read",
			},
			NetworkPolicy: &core.NetworkPolicy{
				AllowOutbound: false,
				AllowedHosts:  []string{},
				AllowedPorts:  []int{},
			},
			StoragePolicy: &core.StoragePolicy{
				AllowLocalStorage:   false,
				AllowSessionStorage: false,
				AllowIndexedDB:      false,
				AllowCookies:        false,
			},
		},
		ID:          policyID,
		Name:        "Default Security Policy",
		Description: "Default security policy with conservative settings",
		Version:     "1.0.0",
		AdminControls: &AdminControls{
			RequireApproval:       false,
			AllowedAdministrators: []string{},
			MaxDocumentSize:       10 * 1024 * 1024, // 10MB
			MaxWASMModules:        5,
			AllowedFileTypes:      []string{"text/html", "text/css", "application/javascript"},
			BlockedDomains:        []string{},
			RequireSignature:      false,
			TrustedSigners:        []string{},
			EnforceQuarantine:     false,
			QuarantineDuration:    3600, // 1 hour
		},
		EventConfig: &SecurityEventConfig{
			LogLevel:             "info",
			EnableAuditLog:       true,
			LogRetentionDays:     90,
			AlertThresholds:      map[string]int{"violations": 10, "suspicious_activity": 5},
			NotificationEmails:   []string{},
			EnableRealTimeAlerts: false,
		},
		ResourceLimits: &ResourceLimits{
			MaxConcurrentDocuments: 10,
			MaxMemoryPerDocument:   64 * 1024 * 1024,  // 64MB
			MaxCPUTimePerDocument:  30000,             // 30 seconds
			MaxNetworkBandwidth:    1024 * 1024,       // 1MB/s
			MaxStorageUsage:        100 * 1024 * 1024, // 100MB
			DocumentTimeoutSeconds: 300,               // 5 minutes
		},
		ComplianceSettings: &ComplianceSettings{
			EnableGDPRCompliance:  false,
			EnableHIPAACompliance: false,
			DataRetentionDays:     30,
			RequireDataEncryption: false,
			AllowedRegions:        []string{},
			DataClassification:    "internal",
		},
	}
}

// generateDocumentID generates a unique ID for a document
func generateDocumentID(doc *core.LIVDocument) string {
	hasher := sha256.New()
	hasher.Write([]byte(doc.Manifest.Metadata.Title))
	hasher.Write([]byte(doc.Manifest.Metadata.Author))
	hasher.Write([]byte(doc.Manifest.Metadata.Created.String()))
	return hex.EncodeToString(hasher.Sum(nil))[:16]
}

// calculatePolicyChanges calculates changes between two policies
func (pm *PolicyManager) calculatePolicyChanges(old, new *SystemSecurityPolicy) map[string]interface{} {
	changes := make(map[string]interface{})

	if old.Name != new.Name {
		changes["name"] = map[string]string{"old": old.Name, "new": new.Name}
	}

	if old.Description != new.Description {
		changes["description"] = map[string]string{"old": old.Description, "new": new.Description}
	}

	// Add more detailed change tracking as needed

	return changes
}

// logSecurityEvent logs a security event
func (pm *PolicyManager) logSecurityEvent(eventType SecurityEventType, severity SecurityEventSeverity, description string, details map[string]interface{}) {
	event := &SecurityEvent{
		ID:          generateEventID(),
		Timestamp:   time.Now(),
		EventType:   eventType,
		Severity:    severity,
		Source:      "policy_manager",
		Description: description,
		Details:     details,
	}

	if pm.eventLogger != nil {
		pm.eventLogger.LogSecurityEvent(event)
	}
}

// logAuditEvent logs an audit event
func (pm *PolicyManager) logAuditEvent(action, resource, userID string, success bool, details map[string]interface{}) {
	event := &AuditEvent{
		ID:        generateEventID(),
		Timestamp: time.Now(),
		Action:    action,
		Resource:  resource,
		UserID:    userID,
		Success:   success,
		Details:   details,
	}

	if pm.auditLogger != nil {
		pm.auditLogger.LogAuditEvent(event)
	}
}

// generateEventID generates a unique event ID
func generateEventID() string {
	hasher := sha256.New()
	hasher.Write([]byte(time.Now().String()))
	return hex.EncodeToString(hasher.Sum(nil))[:16]
}

// contains checks if a slice contains a string
func contains(slice []string, item string) bool {
	for _, s := range slice {
		if s == item {
			return true
		}
	}
	return false
}

// storeQuarantineRecord stores a quarantine record (placeholder implementation)
func (pm *PolicyManager) storeQuarantineRecord(record *QuarantineRecord) error {
	// In a real implementation, this would store to a database
	// For now, we'll just log it
	pm.logAuditEvent("quarantine_document", record.DocumentID, "system", true, map[string]interface{}{
		"reason":     record.Reason,
		"expires_at": record.ExpiresAt,
	})
	return nil
}

// getPolicyTemplate retrieves a policy template by ID
func (pm *PolicyManager) getPolicyTemplate(templateID string) *PolicyTemplate {
	// In a real implementation, this would load from a template store
	// For now, return a basic template
	switch templateID {
	case "basic-security":
		return &PolicyTemplate{
			ID:          "basic-security",
			Name:        "Basic Security Policy",
			Description: "A basic security policy template with conservative settings",
			Category:    "security",
			Template:    pm.createDefaultPolicy("template"),
			Variables:   map[string]interface{}{},
		}
	case "high-security":
		policy := pm.createDefaultPolicy("template")
		policy.AdminControls.RequireSignature = true
		policy.AdminControls.EnforceQuarantine = true
		policy.SecurityPolicy.WASMPermissions.MemoryLimit = 8 * 1024 * 1024 // 8MB
		return &PolicyTemplate{
			ID:          "high-security",
			Name:        "High Security Policy",
			Description: "A high security policy template with strict settings",
			Category:    "security",
			Template:    policy,
			Variables:   map[string]interface{}{},
		}
	default:
		return nil
	}
}

// applyTemplateVariables applies variables to a policy template
func (pm *PolicyManager) applyTemplateVariables(policy *SystemSecurityPolicy, variables map[string]interface{}) error {
	// Apply template variables (simplified implementation)
	if memoryLimit, ok := variables["memory_limit"].(int64); ok {
		policy.SecurityPolicy.WASMPermissions.MemoryLimit = uint64(memoryLimit)
	}

	if maxDocSize, ok := variables["max_document_size"].(int64); ok {
		policy.AdminControls.MaxDocumentSize = maxDocSize
	}

	if requireSig, ok := variables["require_signature"].(bool); ok {
		policy.AdminControls.RequireSignature = requireSig
	}

	return nil
}

// isPolicyOverlyPermissive checks if a policy has overly permissive settings
func (pm *PolicyManager) isPolicyOverlyPermissive(policy *SystemSecurityPolicy) bool {
	// Check for overly permissive WASM settings
	if policy.SecurityPolicy.WASMPermissions != nil {
		if policy.SecurityPolicy.WASMPermissions.MemoryLimit > 64*1024*1024 { // > 64MB
			return true
		}
		if policy.SecurityPolicy.WASMPermissions.AllowNetworking {
			return true
		}
		if policy.SecurityPolicy.WASMPermissions.AllowFileSystem {
			return true
		}
	}

	// Check for overly permissive JS settings
	if policy.SecurityPolicy.JSPermissions != nil {
		if policy.SecurityPolicy.JSPermissions.ExecutionMode == "trusted" {
			return true
		}
		if policy.SecurityPolicy.JSPermissions.DOMAccess == "write" {
			return true
		}
	}

	// Check for overly permissive network settings
	if policy.SecurityPolicy.NetworkPolicy != nil {
		if policy.SecurityPolicy.NetworkPolicy.AllowOutbound {
			return true
		}
	}

	// Check for overly permissive admin controls
	if policy.AdminControls != nil {
		if policy.AdminControls.MaxDocumentSize > 100*1024*1024 { // > 100MB
			return true
		}
		if !policy.AdminControls.RequireSignature {
			return true
		}
	}

	return false
}

// calculateOverallStatus calculates overall system status from issues
func (pm *PolicyManager) calculateOverallStatus(issues []SystemValidationIssue) string {
	if len(issues) == 0 {
		return "healthy"
	}

	criticalCount := 0
	highCount := 0

	for _, issue := range issues {
		switch issue.Severity {
		case "critical":
			criticalCount++
		case "high":
			highCount++
		}
	}

	if criticalCount > 0 {
		return "critical"
	} else if highCount > 0 {
		return "warning"
	} else {
		return "minor_issues"
	}
}
