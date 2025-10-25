// Security policy evaluation functions

package security

import (
	"fmt"
	"strings"

	"github.com/liv-format/liv/pkg/core"
)

// evaluateAdminControls evaluates administrative controls against a document
func (pm *PolicyManager) evaluateAdminControls(doc *core.LIVDocument, controls *AdminControls, evaluation *SecurityEvaluation) error {
	if controls == nil {
		return nil
	}

	// Check document size
	if controls.MaxDocumentSize > 0 {
		docSize := pm.calculateDocumentSize(doc)
		if docSize > controls.MaxDocumentSize {
			evaluation.Violations = append(evaluation.Violations, SecurityViolation{
				Type:        "document_size_exceeded",
				Severity:    SeverityHigh,
				Description: fmt.Sprintf("Document size (%d bytes) exceeds maximum allowed size (%d bytes)", docSize, controls.MaxDocumentSize),
				Details: map[string]interface{}{
					"actual_size": docSize,
					"max_size":    controls.MaxDocumentSize,
				},
				Remediation: "Reduce document size by compressing assets or removing unnecessary content",
			})
		}
	}

	// Check WASM module count
	if controls.MaxWASMModules > 0 {
		wasmCount := len(doc.WASMModules)
		if wasmCount > controls.MaxWASMModules {
			evaluation.Violations = append(evaluation.Violations, SecurityViolation{
				Type:        "wasm_modules_exceeded",
				Severity:    SeverityMedium,
				Description: fmt.Sprintf("Document contains %d WASM modules, exceeding maximum of %d", wasmCount, controls.MaxWASMModules),
				Details: map[string]interface{}{
					"actual_count": wasmCount,
					"max_count":    controls.MaxWASMModules,
				},
				Remediation: "Reduce the number of WASM modules or combine modules",
			})
		}
	}

	// Check file types in assets
	if len(controls.AllowedFileTypes) > 0 {
		pm.checkAllowedFileTypes(doc, controls.AllowedFileTypes, evaluation)
	}

	// Check blocked domains
	if len(controls.BlockedDomains) > 0 {
		pm.checkBlockedDomains(doc, controls.BlockedDomains, evaluation)
	}

	// Check signature requirement
	if controls.RequireSignature {
		pm.checkSignatureRequirement(doc, controls.TrustedSigners, evaluation)
	}

	return nil
}

// evaluateResourceLimits evaluates resource limits against a document
func (pm *PolicyManager) evaluateResourceLimits(doc *core.LIVDocument, limits *ResourceLimits, evaluation *SecurityEvaluation) error {
	if limits == nil {
		return nil
	}

	// Check memory requirements
	if limits.MaxMemoryPerDocument > 0 {
		estimatedMemory := pm.estimateMemoryUsage(doc)
		if estimatedMemory > limits.MaxMemoryPerDocument {
			evaluation.Violations = append(evaluation.Violations, SecurityViolation{
				Type:        "memory_limit_exceeded",
				Severity:    SeverityHigh,
				Description: fmt.Sprintf("Estimated memory usage (%d bytes) exceeds limit (%d bytes)", estimatedMemory, limits.MaxMemoryPerDocument),
				Details: map[string]interface{}{
					"estimated_memory": estimatedMemory,
					"memory_limit":     limits.MaxMemoryPerDocument,
				},
				Remediation: "Optimize document assets or reduce WASM module memory requirements",
			})
		}
	}

	// Check CPU time requirements
	if limits.MaxCPUTimePerDocument > 0 {
		estimatedCPUTime := pm.estimateCPUTime(doc)
		if estimatedCPUTime > limits.MaxCPUTimePerDocument {
			evaluation.Warnings = append(evaluation.Warnings, SecurityWarning{
				Type:        "cpu_time_warning",
				Description: fmt.Sprintf("Estimated CPU time (%d ms) may exceed limit (%d ms)", estimatedCPUTime, limits.MaxCPUTimePerDocument),
				Details: map[string]interface{}{
					"estimated_cpu_time": estimatedCPUTime,
					"cpu_time_limit":     limits.MaxCPUTimePerDocument,
				},
				Recommendation: "Optimize WASM modules for better performance",
			})
		}
	}

	return nil
}

// evaluateCompliance evaluates compliance settings against a document
func (pm *PolicyManager) evaluateCompliance(doc *core.LIVDocument, settings *ComplianceSettings, evaluation *SecurityEvaluation) error {
	if settings == nil {
		return nil
	}

	// Check GDPR compliance
	if settings.EnableGDPRCompliance {
		pm.checkGDPRCompliance(doc, evaluation)
	}

	// Check HIPAA compliance
	if settings.EnableHIPAACompliance {
		pm.checkHIPAACompliance(doc, evaluation)
	}

	// Check data encryption requirement
	if settings.RequireDataEncryption {
		pm.checkDataEncryption(doc, evaluation)
	}

	// Check data classification
	if settings.DataClassification != "" {
		pm.checkDataClassification(doc, settings.DataClassification, evaluation)
	}

	return nil
}

// evaluateCoreSecurityPolicy evaluates the core security policy
func (pm *PolicyManager) evaluateCoreSecurityPolicy(doc *core.LIVDocument, policy *core.SecurityPolicy, evaluation *SecurityEvaluation) error {
	// Evaluate WASM permissions
	if policy.WASMPermissions != nil {
		pm.evaluateWASMPermissions(doc, policy.WASMPermissions, evaluation)
	}

	// Evaluate JS permissions
	if policy.JSPermissions != nil {
		pm.evaluateJSPermissions(doc, policy.JSPermissions, evaluation)
	}

	// Evaluate network policy
	if policy.NetworkPolicy != nil {
		pm.evaluateNetworkPolicy(doc, policy.NetworkPolicy, evaluation)
	}

	// Evaluate storage policy
	if policy.StoragePolicy != nil {
		pm.evaluateStoragePolicy(doc, policy.StoragePolicy, evaluation)
	}

	return nil
}

// Helper evaluation functions

func (pm *PolicyManager) calculateDocumentSize(doc *core.LIVDocument) int64 {
	size := int64(len(doc.Content.HTML) + len(doc.Content.CSS) + len(doc.Content.InteractiveSpec) + len(doc.Content.StaticFallback))

	// Add asset sizes
	for _, data := range doc.Assets.Images {
		size += int64(len(data))
	}
	for _, data := range doc.Assets.Fonts {
		size += int64(len(data))
	}
	for _, data := range doc.Assets.Data {
		size += int64(len(data))
	}

	// Add WASM module sizes
	for _, data := range doc.WASMModules {
		size += int64(len(data))
	}

	return size
}

func (pm *PolicyManager) estimateMemoryUsage(doc *core.LIVDocument) int64 {
	// Rough estimation: document size * 3 (for parsing and processing overhead)
	return pm.calculateDocumentSize(doc) * 3
}

func (pm *PolicyManager) estimateCPUTime(doc *core.LIVDocument) int64 {
	// Rough estimation based on WASM modules and content complexity
	baseTime := int64(100) // 100ms base

	// Add time for each WASM module
	baseTime += int64(len(doc.WASMModules)) * 500 // 500ms per module

	// Add time based on content size
	contentSize := int64(len(doc.Content.HTML) + len(doc.Content.CSS))
	baseTime += contentSize / 1024 // 1ms per KB of content

	return baseTime
}

func (pm *PolicyManager) checkAllowedFileTypes(doc *core.LIVDocument, allowedTypes []string, evaluation *SecurityEvaluation) {
	// Check resources in manifest
	for path, resource := range doc.Manifest.Resources {
		if !contains(allowedTypes, resource.Type) {
			evaluation.Violations = append(evaluation.Violations, SecurityViolation{
				Type:        "disallowed_file_type",
				Severity:    SeverityMedium,
				Description: fmt.Sprintf("File type '%s' not allowed for resource '%s'", resource.Type, path),
				Details: map[string]interface{}{
					"file_path":     path,
					"file_type":     resource.Type,
					"allowed_types": allowedTypes,
				},
				Remediation: "Remove the file or convert it to an allowed file type",
			})
		}
	}
}

func (pm *PolicyManager) checkBlockedDomains(doc *core.LIVDocument, blockedDomains []string, evaluation *SecurityEvaluation) {
	// Check for blocked domains in content
	content := doc.Content.HTML + doc.Content.CSS + doc.Content.InteractiveSpec

	for _, domain := range blockedDomains {
		if strings.Contains(content, domain) {
			evaluation.Violations = append(evaluation.Violations, SecurityViolation{
				Type:        "blocked_domain_reference",
				Severity:    SeverityHigh,
				Description: fmt.Sprintf("Reference to blocked domain '%s' found in document content", domain),
				Details: map[string]interface{}{
					"blocked_domain": domain,
				},
				Remediation: "Remove references to the blocked domain",
			})
		}
	}
}

func (pm *PolicyManager) checkSignatureRequirement(doc *core.LIVDocument, trustedSigners []string, evaluation *SecurityEvaluation) {
	if doc.Signatures == nil || doc.Signatures.ContentSignature == "" {
		evaluation.Violations = append(evaluation.Violations, SecurityViolation{
			Type:        "missing_signature",
			Severity:    SeverityCritical,
			Description: "Document signature is required but not present",
			Details:     map[string]interface{}{},
			Remediation: "Sign the document with a trusted certificate",
		})
		return
	}

	// Additional signature validation would be performed here
	// This is a simplified check
}

func (pm *PolicyManager) checkGDPRCompliance(doc *core.LIVDocument, evaluation *SecurityEvaluation) {
	// Check for potential PII in content
	content := strings.ToLower(doc.Content.HTML + doc.Content.CSS + doc.Content.InteractiveSpec)

	piiKeywords := []string{"email", "phone", "address", "ssn", "social security", "passport", "driver license"}

	for _, keyword := range piiKeywords {
		if strings.Contains(content, keyword) {
			evaluation.Warnings = append(evaluation.Warnings, SecurityWarning{
				Type:        "potential_pii_detected",
				Description: fmt.Sprintf("Potential PII keyword '%s' detected in document content", keyword),
				Details: map[string]interface{}{
					"keyword": keyword,
				},
				Recommendation: "Review content for personal data and ensure GDPR compliance",
			})
		}
	}
}

func (pm *PolicyManager) checkHIPAACompliance(doc *core.LIVDocument, evaluation *SecurityEvaluation) {
	// Check for potential PHI in content
	content := strings.ToLower(doc.Content.HTML + doc.Content.CSS + doc.Content.InteractiveSpec)

	phiKeywords := []string{"patient", "medical", "health", "diagnosis", "treatment", "medication", "doctor", "hospital"}

	for _, keyword := range phiKeywords {
		if strings.Contains(content, keyword) {
			evaluation.Warnings = append(evaluation.Warnings, SecurityWarning{
				Type:        "potential_phi_detected",
				Description: fmt.Sprintf("Potential PHI keyword '%s' detected in document content", keyword),
				Details: map[string]interface{}{
					"keyword": keyword,
				},
				Recommendation: "Review content for protected health information and ensure HIPAA compliance",
			})
		}
	}
}

func (pm *PolicyManager) checkDataEncryption(doc *core.LIVDocument, evaluation *SecurityEvaluation) {
	// Check if document appears to be encrypted
	// This is a simplified check - in practice, you'd check for encryption headers or metadata

	if doc.Signatures == nil || doc.Signatures.ContentSignature == "" {
		evaluation.Violations = append(evaluation.Violations, SecurityViolation{
			Type:        "encryption_required",
			Severity:    SeverityHigh,
			Description: "Data encryption is required but document does not appear to be encrypted",
			Details:     map[string]interface{}{},
			Remediation: "Encrypt the document content before processing",
		})
	}
}

func (pm *PolicyManager) checkDataClassification(doc *core.LIVDocument, requiredClassification string, evaluation *SecurityEvaluation) {
	// Check if document has appropriate classification metadata
	// This would typically be stored in the manifest metadata

	// For now, we'll add a warning if classification is not explicitly set
	evaluation.Warnings = append(evaluation.Warnings, SecurityWarning{
		Type:        "data_classification_check",
		Description: fmt.Sprintf("Verify document meets '%s' data classification requirements", requiredClassification),
		Details: map[string]interface{}{
			"required_classification": requiredClassification,
		},
		Recommendation: "Review document content and ensure it meets the required data classification level",
	})
}

func (pm *PolicyManager) evaluateWASMPermissions(doc *core.LIVDocument, permissions *core.WASMPermissions, evaluation *SecurityEvaluation) {
	// Check WASM modules against permissions
	for moduleName, moduleData := range doc.WASMModules {
		// Check module size against memory limit
		moduleSize := int64(len(moduleData))
		if moduleSize > int64(permissions.MemoryLimit) {
			evaluation.Violations = append(evaluation.Violations, SecurityViolation{
				Type:        "wasm_memory_exceeded",
				Severity:    SeverityHigh,
				Description: fmt.Sprintf("WASM module '%s' size (%d bytes) exceeds memory limit (%d bytes)", moduleName, moduleSize, permissions.MemoryLimit),
				Details: map[string]interface{}{
					"module_name":  moduleName,
					"module_size":  moduleSize,
					"memory_limit": permissions.MemoryLimit,
				},
				Remediation: "Optimize WASM module or increase memory limit",
			})
		}

		// Additional WASM validation would be performed here
		// This could include checking imports against allowed imports, etc.
	}
}

func (pm *PolicyManager) evaluateJSPermissions(doc *core.LIVDocument, permissions *core.JSPermissions, evaluation *SecurityEvaluation) {
	// Check JavaScript content against permissions
	jsContent := doc.Content.HTML + doc.Content.InteractiveSpec

	if permissions.ExecutionMode == "none" && strings.Contains(jsContent, "<script") {
		evaluation.Violations = append(evaluation.Violations, SecurityViolation{
			Type:        "js_execution_forbidden",
			Severity:    SeverityHigh,
			Description: "JavaScript execution is forbidden but script tags found in content",
			Details:     map[string]interface{}{},
			Remediation: "Remove JavaScript content or change execution mode",
		})
	}

	// Check for dangerous JavaScript patterns
	dangerousPatterns := []string{"eval(", "Function(", "setTimeout(", "setInterval("}
	for _, pattern := range dangerousPatterns {
		if strings.Contains(jsContent, pattern) {
			evaluation.Warnings = append(evaluation.Warnings, SecurityWarning{
				Type:        "dangerous_js_pattern",
				Description: fmt.Sprintf("Potentially dangerous JavaScript pattern '%s' detected", pattern),
				Details: map[string]interface{}{
					"pattern": pattern,
				},
				Recommendation: "Review JavaScript code for security implications",
			})
		}
	}
}

func (pm *PolicyManager) evaluateNetworkPolicy(doc *core.LIVDocument, policy *core.NetworkPolicy, evaluation *SecurityEvaluation) {
	// Check for network requests in content
	content := doc.Content.HTML + doc.Content.CSS + doc.Content.InteractiveSpec

	if !policy.AllowOutbound {
		// Check for external URLs
		urlPatterns := []string{"http://", "https://", "ftp://", "ws://", "wss://"}
		for _, pattern := range urlPatterns {
			if strings.Contains(content, pattern) {
				evaluation.Warnings = append(evaluation.Warnings, SecurityWarning{
					Type:        "external_url_detected",
					Description: fmt.Sprintf("External URL pattern '%s' detected but outbound connections not allowed", pattern),
					Details: map[string]interface{}{
						"url_pattern": pattern,
					},
					Recommendation: "Remove external URLs or enable outbound connections",
				})
			}
		}
	}
}

func (pm *PolicyManager) evaluateStoragePolicy(doc *core.LIVDocument, policy *core.StoragePolicy, evaluation *SecurityEvaluation) {
	// Check for storage API usage in content
	content := doc.Content.HTML + doc.Content.InteractiveSpec

	storageAPIs := map[string]bool{
		"localStorage":    policy.AllowLocalStorage,
		"sessionStorage":  policy.AllowSessionStorage,
		"indexedDB":       policy.AllowIndexedDB,
		"document.cookie": policy.AllowCookies,
	}

	for api, allowed := range storageAPIs {
		if !allowed && strings.Contains(content, api) {
			evaluation.Violations = append(evaluation.Violations, SecurityViolation{
				Type:        "storage_api_forbidden",
				Severity:    SeverityMedium,
				Description: fmt.Sprintf("Storage API '%s' usage detected but not allowed by policy", api),
				Details: map[string]interface{}{
					"storage_api": api,
				},
				Remediation: fmt.Sprintf("Remove usage of %s or enable it in the storage policy", api),
			})
		}
	}
}
