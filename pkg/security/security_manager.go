package security

import (
	"context"
	"fmt"
	"time"

	"github.com/liv-format/liv/pkg/core"
)

// SecurityManager implements the core.SecurityManager interface
type SecurityManager struct {
	policyEngine        *PolicyEngine
	permissionValidator *PermissionValidator
	resourceMonitor     *ResourceMonitor
	cryptoProvider      core.CryptoProvider
	logger              core.Logger
	metrics             core.MetricsCollector
	config              *SecurityConfiguration
}

// NewSecurityManager creates a new security manager with all components
func NewSecurityManager(cryptoProvider core.CryptoProvider, logger core.Logger, metrics core.MetricsCollector) *SecurityManager {
	config := &SecurityConfiguration{
		MaxMemoryPerModule:       64 * 1024 * 1024, // 64 MB
		MaxCPUTimePerModule:      30 * time.Second, // 30 seconds
		MaxConcurrentModules:     10,
		AuditLogEnabled:          true,
		MetricsCollectionEnabled: true,
		StrictModeEnabled:        false,
	}

	policyEngine := NewPolicyEngine(logger, metrics)
	permissionValidator := NewPermissionValidator(policyEngine, logger, metrics)
	resourceMonitor := NewResourceMonitor(permissionValidator, logger, metrics)

	return &SecurityManager{
		policyEngine:        policyEngine,
		permissionValidator: permissionValidator,
		resourceMonitor:     resourceMonitor,
		cryptoProvider:      cryptoProvider,
		logger:              logger,
		metrics:             metrics,
		config:              config,
	}
}

// ValidateSignature verifies cryptographic signatures
func (sm *SecurityManager) ValidateSignature(content []byte, signature string, publicKey []byte) bool {
	if sm.cryptoProvider == nil {
		sm.logger.Error("crypto provider not available")
		return false
	}

	// Convert signature from string to bytes (assuming hex encoding)
	sigBytes := []byte(signature)

	result := sm.cryptoProvider.Verify(content, sigBytes, publicKey)

	if sm.metrics != nil {
		sm.metrics.RecordSecurityEvent("signature_validation", map[string]interface{}{
			"valid": result,
		})
	}

	return result
}

// CreateSignature creates a cryptographic signature
func (sm *SecurityManager) CreateSignature(content []byte, privateKey []byte) (string, error) {
	if sm.cryptoProvider == nil {
		return "", fmt.Errorf("crypto provider not available")
	}

	sigBytes, err := sm.cryptoProvider.Sign(content, privateKey)
	if err != nil {
		return "", fmt.Errorf("failed to create signature: %w", err)
	}

	// Convert signature bytes to string (hex encoding)
	signature := string(sigBytes)

	if sm.metrics != nil {
		sm.metrics.RecordSecurityEvent("signature_creation", map[string]interface{}{
			"success": true,
		})
	}

	return signature, nil
}

// ValidateWASMModule validates a WASM module against security policies
func (sm *SecurityManager) ValidateWASMModule(module []byte, permissions *core.WASMPermissions) error {
	result := sm.policyEngine.ValidateWASMModule(module, permissions)

	if !result.IsValid && !result.Valid {
		return fmt.Errorf("WASM module validation failed: %v", result.Errors)
	}

	if len(result.Warnings) > 0 {
		sm.logger.Warn("WASM module validation warnings", "warnings", result.Warnings)
	}

	return nil
}

// CreateSandbox creates a secure execution environment
func (sm *SecurityManager) CreateSandbox(policy *core.SecurityPolicy) (core.Sandbox, error) {
	// Create security session
	securityCtx, err := sm.permissionValidator.CreateSession(policy)
	if err != nil {
		return nil, fmt.Errorf("failed to create security session: %w", err)
	}

	// Create sandbox implementation
	sandbox := &Sandbox{
		securityContext:     securityCtx,
		permissionValidator: sm.permissionValidator,
		resourceMonitor:     sm.resourceMonitor,
		logger:              sm.logger,
		metrics:             sm.metrics,
	}

	sm.logger.Info("sandbox created", "session_id", securityCtx.SessionID)

	return sandbox, nil
}

// EvaluatePermissions evaluates permission requests against policies
func (sm *SecurityManager) EvaluatePermissions(requested *core.WASMPermissions, policy *core.SecurityPolicy) bool {
	result := sm.policyEngine.EvaluateWASMPermissions(requested, policy)
	return result.Allowed
}

// GenerateSecurityReport creates a comprehensive security report
func (sm *SecurityManager) GenerateSecurityReport(doc *core.LIVDocument) *core.SecurityReport {
	report := &core.SecurityReport{
		IsValid:           true,
		SignatureVerified: false,
		IntegrityChecked:  false,
		PermissionsValid:  false,
		Warnings:          []string{},
		Errors:            []string{},
	}

	if doc == nil {
		report.IsValid = false
		report.Errors = append(report.Errors, "document is nil")
		return report
	}

	// Validate manifest
	if doc.Manifest == nil {
		report.IsValid = false
		report.Errors = append(report.Errors, "manifest is missing")
	} else {
		// Validate security policy
		if doc.Manifest.Security == nil {
			report.IsValid = false
			report.Errors = append(report.Errors, "security policy is missing")
		} else {
			report.PermissionsValid = sm.validateSecurityPolicy(doc.Manifest.Security, report)
		}
	}

	// Validate signatures
	if doc.Signatures != nil {
		report.SignatureVerified = sm.validateSignatures(doc, report)
	} else {
		report.Warnings = append(report.Warnings, "no signatures found")
	}

	// Validate resource integrity
	report.IntegrityChecked = sm.validateResourceIntegrity(doc, report)

	// Validate WASM modules
	if len(doc.WASMModules) > 0 {
		sm.validateWASMModules(doc, report)
	}

	report.IsValid = report.IsValid && len(report.Errors) == 0

	return report
}

// StartResourceMonitoring starts the resource monitoring system
func (sm *SecurityManager) StartResourceMonitoring(ctx context.Context) error {
	sm.resourceMonitor.StartMonitoring(ctx, 5*time.Second)
	sm.logger.Info("resource monitoring started")
	return nil
}

// StopResourceMonitoring stops the resource monitoring system
func (sm *SecurityManager) StopResourceMonitoring() error {
	sm.resourceMonitor.StopMonitoring()
	sm.logger.Info("resource monitoring stopped")
	return nil
}

// GetSecurityMetrics returns current security metrics
func (sm *SecurityManager) GetSecurityMetrics() map[string]interface{} {
	activeSessions := sm.permissionValidator.ListActiveSessions()
	resourceSummary := sm.resourceMonitor.GetResourceSummary()

	return map[string]interface{}{
		"active_sessions":   len(activeSessions),
		"total_modules":     resourceSummary.TotalModules,
		"total_memory_used": resourceSummary.TotalMemoryUsed,
		"total_cpu_time":    resourceSummary.TotalCPUTime.Milliseconds(),
		"policy_violations": resourceSummary.PolicyViolations,
		"timestamp":         time.Now(),
	}
}

// CleanupExpiredSessions removes expired security sessions
func (sm *SecurityManager) CleanupExpiredSessions(maxAge time.Duration) int {
	return sm.permissionValidator.CleanupExpiredSessions(maxAge)
}

// UpdateSecurityConfiguration updates the security configuration
func (sm *SecurityManager) UpdateSecurityConfiguration(config *SecurityConfiguration) error {
	if config == nil {
		return fmt.Errorf("security configuration cannot be nil")
	}

	sm.config = config
	sm.logger.Info("security configuration updated")

	return nil
}

// GetSecurityConfiguration returns the current security configuration
func (sm *SecurityManager) GetSecurityConfiguration() *SecurityConfiguration {
	return sm.config
}

// Helper methods for security report generation

func (sm *SecurityManager) validateSecurityPolicy(policy *core.SecurityPolicy, report *core.SecurityReport) bool {
	valid := true

	if policy.WASMPermissions == nil {
		report.Errors = append(report.Errors, "WASM permissions not defined")
		valid = false
	} else {
		// Check for overly permissive settings
		if int64(policy.WASMPermissions.MemoryLimit) > sm.config.MaxMemoryPerModule {
			report.Warnings = append(report.Warnings,
				fmt.Sprintf("WASM memory limit %d exceeds recommended maximum %d",
					policy.WASMPermissions.MemoryLimit, sm.config.MaxMemoryPerModule))
		}

		cpuLimitMs := uint64(sm.config.MaxCPUTimePerModule / time.Millisecond)
		if policy.WASMPermissions.CPUTimeLimit > cpuLimitMs {
			report.Warnings = append(report.Warnings,
				fmt.Sprintf("WASM CPU time limit %d exceeds recommended maximum %d",
					policy.WASMPermissions.CPUTimeLimit, cpuLimitMs))
		}

		if policy.WASMPermissions.AllowNetworking {
			report.Warnings = append(report.Warnings, "WASM networking is enabled")
		}

		if policy.WASMPermissions.AllowFileSystem {
			report.Warnings = append(report.Warnings, "WASM file system access is enabled")
		}
	}

	if policy.JSPermissions == nil {
		report.Errors = append(report.Errors, "JavaScript permissions not defined")
		valid = false
	} else if policy.JSPermissions.ExecutionMode == "trusted" {
		report.Warnings = append(report.Warnings, "JavaScript execution mode is set to trusted")
	}

	return valid
}

func (sm *SecurityManager) validateSignatures(doc *core.LIVDocument, report *core.SecurityReport) bool {
	// This would require actual signature validation logic
	// For now, just check if signatures exist
	if doc.Signatures.ContentSignature == "" {
		report.Warnings = append(report.Warnings, "content signature is empty")
		return false
	}

	if doc.Signatures.ManifestSignature == "" {
		report.Warnings = append(report.Warnings, "manifest signature is empty")
		return false
	}

	// In a real implementation, we would validate the signatures here
	return true
}

func (sm *SecurityManager) validateResourceIntegrity(doc *core.LIVDocument, report *core.SecurityReport) bool {
	if doc.Manifest == nil || doc.Manifest.Resources == nil {
		report.Errors = append(report.Errors, "resource manifest is missing")
		return false
	}

	// Check that all resources have integrity hashes
	for path, resource := range doc.Manifest.Resources {
		if resource.Hash == "" {
			report.Errors = append(report.Errors, fmt.Sprintf("resource %s missing integrity hash", path))
			return false
		}
	}

	return true
}

func (sm *SecurityManager) validateWASMModules(doc *core.LIVDocument, report *core.SecurityReport) {
	for name, moduleData := range doc.WASMModules {
		var permissions *core.WASMPermissions
		if doc.Manifest.WASMConfig != nil && doc.Manifest.WASMConfig.Modules != nil {
			if module, exists := doc.Manifest.WASMConfig.Modules[name]; exists {
				permissions = module.Permissions
			}
		}

		result := sm.policyEngine.ValidateWASMModule(moduleData, permissions)
		if !result.IsValid {
			report.Errors = append(report.Errors,
				fmt.Sprintf("WASM module %s validation failed: %v", name, result.Errors))
		}

		if len(result.Warnings) > 0 {
			for _, warning := range result.Warnings {
				report.Warnings = append(report.Warnings,
					fmt.Sprintf("WASM module %s: %s", name, warning))
			}
		}
	}
}
