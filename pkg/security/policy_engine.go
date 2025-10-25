package security

import (
	"context"
	"fmt"
	"time"

	"github.com/liv-format/liv/pkg/core"
)

// PolicyEngine implements security policy evaluation and enforcement
type PolicyEngine struct {
	defaultPolicy *core.SecurityPolicy
	logger        core.Logger
	metrics       core.MetricsCollector
}

// NewPolicyEngine creates a new security policy engine
func NewPolicyEngine(logger core.Logger, metrics core.MetricsCollector) *PolicyEngine {
	return &PolicyEngine{
		defaultPolicy: createDefaultSecurityPolicy(),
		logger:        logger,
		metrics:       metrics,
	}
}

// EvaluateWASMPermissions evaluates WASM permission requests against security policy
func (pe *PolicyEngine) EvaluateWASMPermissions(requested *core.WASMPermissions, policy *core.SecurityPolicy) *PermissionEvaluationResult {
	result := &PermissionEvaluationResult{
		Allowed:   true,
		Warnings:  []string{},
		Errors:    []string{},
		Timestamp: time.Now(),
	}

	if requested == nil {
		result.Allowed = false
		result.Errors = append(result.Errors, "requested permissions cannot be nil")
		return result
	}

	if policy == nil || policy.WASMPermissions == nil {
		pe.logger.Warn("no security policy provided, using default restrictive policy")
		policy = pe.defaultPolicy
	}

	allowedPerms := policy.WASMPermissions

	// Evaluate memory limit
	if requested.MemoryLimit > allowedPerms.MemoryLimit {
		result.Allowed = false
		result.Errors = append(result.Errors,
			fmt.Sprintf("requested memory limit %d exceeds allowed limit %d",
				requested.MemoryLimit, allowedPerms.MemoryLimit))
	}

	// Evaluate CPU time limit
	if requested.CPUTimeLimit > allowedPerms.CPUTimeLimit {
		result.Allowed = false
		result.Errors = append(result.Errors,
			fmt.Sprintf("requested CPU time limit %d exceeds allowed limit %d",
				requested.CPUTimeLimit, allowedPerms.CPUTimeLimit))
	}

	// Evaluate networking permissions
	if requested.AllowNetworking && !allowedPerms.AllowNetworking {
		result.Allowed = false
		result.Errors = append(result.Errors, "networking access not permitted by policy")
	}

	// Evaluate file system permissions
	if requested.AllowFileSystem && !allowedPerms.AllowFileSystem {
		result.Allowed = false
		result.Errors = append(result.Errors, "file system access not permitted by policy")
	}

	// Evaluate allowed imports
	for _, requestedImport := range requested.AllowedImports {
		if !pe.isImportAllowed(requestedImport, allowedPerms.AllowedImports) {
			result.Allowed = false
			result.Errors = append(result.Errors,
				fmt.Sprintf("import '%s' not permitted by policy", requestedImport))
		}
	}

	// Add warnings for potentially risky permissions
	if requested.MemoryLimit > 64*1024*1024 { // 64MB
		result.Warnings = append(result.Warnings, "high memory limit requested")
	}

	if requested.CPUTimeLimit > 10000 { // 10 seconds
		result.Warnings = append(result.Warnings, "high CPU time limit requested")
	}

	if requested.AllowNetworking {
		result.Warnings = append(result.Warnings, "network access requested")
	}

	pe.recordEvaluationMetrics(result)
	return result
}

// ValidateWASMModule validates a WASM module against security constraints
func (pe *PolicyEngine) ValidateWASMModule(moduleData []byte, permissions *core.WASMPermissions) *ModuleValidationResult {
	result := &ModuleValidationResult{
		IsValid:   true,
		Errors:    []string{},
		Warnings:  []string{},
		Timestamp: time.Now(),
	}

	if len(moduleData) == 0 {
		result.IsValid = false
		result.Errors = append(result.Errors, "WASM module data is empty")
		return result
	}

	// Validate WASM magic number
	if len(moduleData) < 4 || string(moduleData[:4]) != "\x00asm" {
		result.IsValid = false
		result.Errors = append(result.Errors, "invalid WASM magic number")
		return result
	}

	// Check module size limits
	maxModuleSize := uint64(16 * 1024 * 1024) // 16MB default limit
	if permissions != nil && permissions.MemoryLimit > 0 {
		maxModuleSize = permissions.MemoryLimit
	}

	if uint64(len(moduleData)) > maxModuleSize {
		result.IsValid = false
		result.Errors = append(result.Errors,
			fmt.Sprintf("WASM module size %d exceeds limit %d", len(moduleData), maxModuleSize))
	}

	// Validate WASM version
	if len(moduleData) >= 8 {
		version := uint32(moduleData[4]) | uint32(moduleData[5])<<8 |
			uint32(moduleData[6])<<16 | uint32(moduleData[7])<<24
		if version != 1 {
			result.Warnings = append(result.Warnings,
				fmt.Sprintf("unsupported WASM version: %d", version))
		}
	}

	// Additional security checks would go here
	// - Import validation
	// - Export validation
	// - Memory section validation
	// - Code section validation

	pe.recordModuleValidationMetrics(result, len(moduleData))
	return result
}

// EnforceResourceLimits creates resource constraints for WASM execution
func (pe *PolicyEngine) EnforceResourceLimits(permissions *core.WASMPermissions) *ResourceConstraints {
	if permissions == nil {
		permissions = pe.defaultPolicy.WASMPermissions
	}

	return &ResourceConstraints{
		MemoryLimit:     int64(permissions.MemoryLimit),
		CPUTimeLimit:    time.Duration(permissions.CPUTimeLimit) * time.Millisecond,
		AllowNetworking: permissions.AllowNetworking,
		AllowFileSystem: permissions.AllowFileSystem,
	}
}

// CreateSecurityContext creates a security context for WASM execution
func (pe *PolicyEngine) CreateSecurityContext(policy *core.SecurityPolicy) *SecurityContext {
	if policy == nil {
		policy = pe.defaultPolicy
	}

	return &SecurityContext{
		Policy:      policy,
		Constraints: pe.EnforceResourceLimits(policy.WASMPermissions),
		SessionID:   pe.generateSessionID(),
		CreatedAt:   time.Now(),
	}
}

// ValidatePermissionRequest validates a permission request against current policy
func (pe *PolicyEngine) ValidatePermissionRequest(ctx context.Context, request *PermissionRequest, securityCtx *SecurityContext) *PermissionResponse {
	response := &PermissionResponse{
		RequestID: request.DocumentID, // Use DocumentID as request ID
		Granted:   false,
		Reason:    "",
	}

	if securityCtx == nil {
		response.Reason = "no security context provided"
		return response
	}

	switch request.Type {
	case string(PermissionTypeMemory):
		response.Granted = pe.validateMemoryRequest(request, securityCtx)
		if !response.Granted {
			response.Reason = "memory allocation exceeds policy limits"
		}

	case string(PermissionTypeNetwork):
		response.Granted = pe.validateNetworkRequest(request, securityCtx)
		if !response.Granted {
			response.Reason = "network access not permitted by policy"
		}

	case string(PermissionTypeFileSystem):
		response.Granted = pe.validateFileSystemRequest(request, securityCtx)
		if !response.Granted {
			response.Reason = "file system access not permitted by policy"
		}

	case string(PermissionTypeImport):
		response.Granted = pe.validateImportRequest(request, securityCtx)
		if !response.Granted {
			response.Reason = "import not permitted by policy"
		}

	default:
		response.Reason = "unknown permission type"
	}

	pe.recordPermissionRequest(request, response, securityCtx)
	return response
}

// Helper methods

func (pe *PolicyEngine) isImportAllowed(requestedImport string, allowedImports []string) bool {
	if len(allowedImports) == 0 {
		return false // Default deny
	}

	for _, allowed := range allowedImports {
		if allowed == "*" || allowed == requestedImport {
			return true
		}
		// Support wildcard matching
		if pe.matchesWildcard(requestedImport, allowed) {
			return true
		}
	}
	return false
}

func (pe *PolicyEngine) matchesWildcard(input, pattern string) bool {
	// Simple wildcard matching - could be enhanced with regex
	if pattern == "*" {
		return true
	}
	// For now, just exact match or full wildcard
	return input == pattern
}

func (pe *PolicyEngine) validateMemoryRequest(request *PermissionRequest, ctx *SecurityContext) bool {
	// Simplified: just check if memory is allowed based on constraints
	return ctx.Constraints.MemoryLimit > 0
}

func (pe *PolicyEngine) validateNetworkRequest(request *PermissionRequest, ctx *SecurityContext) bool {
	return ctx.Constraints.AllowNetworking
}

func (pe *PolicyEngine) validateFileSystemRequest(request *PermissionRequest, ctx *SecurityContext) bool {
	return ctx.Constraints.AllowFileSystem
}

func (pe *PolicyEngine) validateImportRequest(request *PermissionRequest, ctx *SecurityContext) bool {
	// Simplified: allow imports if module name is specified
	return request.ModuleName != ""
}

func (pe *PolicyEngine) generateSessionID() string {
	// Simple session ID generation - could use crypto/rand for production
	return fmt.Sprintf("session_%d", time.Now().UnixNano())
}

func (pe *PolicyEngine) recordEvaluationMetrics(result *PermissionEvaluationResult) {
	if pe.metrics != nil {
		pe.metrics.RecordSecurityEvent("permission_evaluation", map[string]interface{}{
			"allowed":       result.Allowed,
			"error_count":   len(result.Errors),
			"warning_count": len(result.Warnings),
		})
	}
}

func (pe *PolicyEngine) recordModuleValidationMetrics(result *ModuleValidationResult, moduleSize int) {
	if pe.metrics != nil {
		pe.metrics.RecordSecurityEvent("module_validation", map[string]interface{}{
			"valid":       result.IsValid,
			"module_size": moduleSize,
			"error_count": len(result.Errors),
		})
	}
}

func (pe *PolicyEngine) recordPermissionRequest(request *PermissionRequest, response *PermissionResponse, ctx *SecurityContext) {
	if pe.metrics != nil {
		pe.metrics.RecordSecurityEvent("permission_request", map[string]interface{}{
			"type":       string(request.Type),
			"granted":    response.Granted,
			"session_id": ctx.SessionID,
		})
	}

	if pe.logger != nil {
		pe.logger.Info("permission request processed",
			"request_id", request.DocumentID,
			"type", request.Type,
			"granted", response.Granted,
			"reason", response.Reason,
			"session_id", ctx.SessionID,
		)
	}
}

func createDefaultSecurityPolicy() *core.SecurityPolicy {
	return &core.SecurityPolicy{
		WASMPermissions: &core.WASMPermissions{
			MemoryLimit:     4 * 1024 * 1024, // 4MB
			AllowedImports:  []string{},      // No imports allowed by default
			CPUTimeLimit:    1000,            // 1 second
			AllowNetworking: false,
			AllowFileSystem: false,
		},
		JSPermissions: &core.JSPermissions{
			ExecutionMode: "sandboxed",
			AllowedAPIs:   []string{},
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
		ContentSecurityPolicy: "default-src 'none'; script-src 'self'",
		TrustedDomains:        []string{},
	}
}
