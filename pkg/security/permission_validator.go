package security

import (
	"context"
	"fmt"
	"sync"
	"time"

	"github.com/liv-format/liv/pkg/core"
)

// PermissionValidator handles runtime permission validation and enforcement
type PermissionValidator struct {
	policyEngine   *PolicyEngine
	activeSessions map[string]*SecurityContext
	sessionMutex   sync.RWMutex
	logger         core.Logger
	metrics        core.MetricsCollector
}

// NewPermissionValidator creates a new permission validator
func NewPermissionValidator(policyEngine *PolicyEngine, logger core.Logger, metrics core.MetricsCollector) *PermissionValidator {
	return &PermissionValidator{
		policyEngine:   policyEngine,
		activeSessions: make(map[string]*SecurityContext),
		logger:         logger,
		metrics:        metrics,
	}
}

// CreateSession creates a new security session for WASM execution
func (pv *PermissionValidator) CreateSession(policy *core.SecurityPolicy) (*SecurityContext, error) {
	if policy == nil {
		return nil, fmt.Errorf("security policy cannot be nil")
	}

	ctx := pv.policyEngine.CreateSecurityContext(policy)

	pv.sessionMutex.Lock()
	pv.activeSessions[ctx.SessionID] = ctx
	pv.sessionMutex.Unlock()

	pv.logger.Info("security session created", "session_id", ctx.SessionID)

	if pv.metrics != nil {
		pv.metrics.RecordSecurityEvent("session_created", map[string]interface{}{
			"session_id": ctx.SessionID,
		})
	}

	return ctx, nil
}

// DestroySession destroys a security session and cleans up resources
func (pv *PermissionValidator) DestroySession(sessionID string) error {
	pv.sessionMutex.Lock()
	defer pv.sessionMutex.Unlock()

	ctx, exists := pv.activeSessions[sessionID]
	if !exists {
		return fmt.Errorf("session %s not found", sessionID)
	}

	delete(pv.activeSessions, sessionID)

	pv.logger.Info("security session destroyed", "session_id", sessionID)

	if pv.metrics != nil {
		pv.metrics.RecordSecurityEvent("session_destroyed", map[string]interface{}{
			"session_id": sessionID,
			"duration":   time.Since(ctx.CreatedAt).Milliseconds(),
		})
	}

	return nil
}

// ValidateMemoryAllocation validates a memory allocation request
func (pv *PermissionValidator) ValidateMemoryAllocation(sessionID string, requestedSize uint64) (*PermissionResponse, error) {
	ctx, err := pv.getSession(sessionID)
	if err != nil {
		return nil, err
	}

	request := &PermissionRequest{
		Type:      string(PermissionTypeMemory),
		Timestamp: time.Now(),
	}

	response := pv.policyEngine.ValidatePermissionRequest(context.Background(), request, ctx)

	if !response.Granted {
		pv.recordViolation(sessionID, "memory_allocation_denied", map[string]interface{}{
			"requested_size": requestedSize,
			"limit":          ctx.Constraints.MemoryLimit,
		}, SecuritySeverityWarning)
	}

	return response, nil
}

// ValidateNetworkAccess validates a network access request
func (pv *PermissionValidator) ValidateNetworkAccess(sessionID string, host string, port int) (*PermissionResponse, error) {
	ctx, err := pv.getSession(sessionID)
	if err != nil {
		return nil, err
	}

	request := &PermissionRequest{
		Type:      string(PermissionTypeNetwork),
		Timestamp: time.Now(),
	}

	response := pv.policyEngine.ValidatePermissionRequest(context.Background(), request, ctx)

	// Additional network-specific validation
	if response.Granted && ctx.Policy.NetworkPolicy != nil {
		if !pv.isHostAllowed(host, ctx.Policy.NetworkPolicy.AllowedHosts) {
			response.Granted = false
			response.Reason = fmt.Sprintf("host '%s' not in allowed hosts list", host)
		}

		if response.Granted && !pv.isPortAllowed(port, ctx.Policy.NetworkPolicy.AllowedPorts) {
			response.Granted = false
			response.Reason = fmt.Sprintf("port %d not in allowed ports list", port)
		}
	}

	if !response.Granted {
		pv.recordViolation(sessionID, "network_access_denied", map[string]interface{}{
			"host": host,
			"port": port,
		}, SecuritySeverityWarning)
	}

	return response, nil
}

// ValidateFileSystemAccess validates a file system access request
func (pv *PermissionValidator) ValidateFileSystemAccess(sessionID string, path string, operation string) (*PermissionResponse, error) {
	ctx, err := pv.getSession(sessionID)
	if err != nil {
		return nil, err
	}

	request := &PermissionRequest{
		Type:      string(PermissionTypeFileSystem),
		Timestamp: time.Now(),
	}

	response := pv.policyEngine.ValidatePermissionRequest(context.Background(), request, ctx)

	if !response.Granted {
		pv.recordViolation(sessionID, "filesystem_access_denied", map[string]interface{}{
			"path":      path,
			"operation": operation,
		}, SecuritySeverityWarning)
	}

	return response, nil
}

// ValidateImport validates a WASM import request
func (pv *PermissionValidator) ValidateImport(sessionID string, importName string, importType string) (*PermissionResponse, error) {
	ctx, err := pv.getSession(sessionID)
	if err != nil {
		return nil, err
	}

	request := &PermissionRequest{
		Type:      string(PermissionTypeImport),
		Timestamp: time.Now(),
	}

	response := pv.policyEngine.ValidatePermissionRequest(context.Background(), request, ctx)

	if !response.Granted {
		pv.recordViolation(sessionID, "import_denied", map[string]interface{}{
			"import":      importName,
			"import_type": importType,
		}, SecuritySeverityError)
	}

	return response, nil
}

// EnforceResourceLimits enforces resource limits for a session
func (pv *PermissionValidator) EnforceResourceLimits(sessionID string, metrics *RuntimeMetrics) []PolicyViolation {
	ctx, err := pv.getSession(sessionID)
	if err != nil {
		return []PolicyViolation{{
			Type:        "session_error",
			Description: err.Error(),
			SessionID:   sessionID,
			Timestamp:   time.Now(),
			Severity:    SecuritySeverityError,
		}}
	}

	var violations []PolicyViolation

	// Check memory limits
	if metrics.Memory != nil && metrics.Memory.IsMemoryLimitExceeded() {
		violation := PolicyViolation{
			Type:        "memory_limit_exceeded",
			Description: fmt.Sprintf("Memory usage %d exceeds limit %d", metrics.Memory.Used, metrics.Memory.Limit),
			SessionID:   sessionID,
			ModuleName:  metrics.ModuleName,
			Details: map[string]interface{}{
				"used":  metrics.Memory.Used,
				"limit": metrics.Memory.Limit,
			},
			Timestamp: time.Now(),
			Severity:  SecuritySeverityError,
		}
		violations = append(violations, violation)
		pv.recordViolation(sessionID, violation.Type, violation.Details, violation.Severity)
	}

	// Check CPU limits
	if metrics.CPU != nil && metrics.CPU.IsCPULimitExceeded() {
		violation := PolicyViolation{
			Type:        "cpu_limit_exceeded",
			Description: fmt.Sprintf("CPU time %v exceeds limit %v", metrics.CPU.Used, metrics.CPU.Limit),
			SessionID:   sessionID,
			ModuleName:  metrics.ModuleName,
			Details: map[string]interface{}{
				"used":  metrics.CPU.Used.Milliseconds(),
				"limit": metrics.CPU.Limit.Milliseconds(),
			},
			Timestamp: time.Now(),
			Severity:  SecuritySeverityError,
		}
		violations = append(violations, violation)
		pv.recordViolation(sessionID, violation.Type, violation.Details, violation.Severity)
	}

	// Check network activity if networking is disabled
	if metrics.Network != nil && metrics.Network.RequestCount > 0 && !ctx.Constraints.AllowNetworking {
		violation := PolicyViolation{
			Type:        "unauthorized_network_access",
			Description: "Network access attempted when not permitted",
			SessionID:   sessionID,
			ModuleName:  metrics.ModuleName,
			Details: map[string]interface{}{
				"request_count": metrics.Network.RequestCount,
			},
			Timestamp: time.Now(),
			Severity:  SecuritySeverityCritical,
		}
		violations = append(violations, violation)
		pv.recordViolation(sessionID, violation.Type, violation.Details, violation.Severity)
	}

	// Check file system activity if file system access is disabled
	if metrics.FileSystem != nil && (metrics.FileSystem.ReadOperations > 0 || metrics.FileSystem.WriteOperations > 0) && !ctx.Constraints.AllowFileSystem {
		violation := PolicyViolation{
			Type:        "unauthorized_filesystem_access",
			Description: "File system access attempted when not permitted",
			SessionID:   sessionID,
			ModuleName:  metrics.ModuleName,
			Details: map[string]interface{}{
				"read_operations":  metrics.FileSystem.ReadOperations,
				"write_operations": metrics.FileSystem.WriteOperations,
			},
			Timestamp: time.Now(),
			Severity:  SecuritySeverityCritical,
		}
		violations = append(violations, violation)
		pv.recordViolation(sessionID, violation.Type, violation.Details, violation.Severity)
	}

	return violations
}

// GetSessionMetrics returns metrics for a specific session
func (pv *PermissionValidator) GetSessionMetrics(sessionID string) (*SecurityContext, error) {
	return pv.getSession(sessionID)
}

// ListActiveSessions returns a list of active session IDs
func (pv *PermissionValidator) ListActiveSessions() []string {
	pv.sessionMutex.RLock()
	defer pv.sessionMutex.RUnlock()

	sessions := make([]string, 0, len(pv.activeSessions))
	for sessionID := range pv.activeSessions {
		sessions = append(sessions, sessionID)
	}
	return sessions
}

// CleanupExpiredSessions removes expired sessions
func (pv *PermissionValidator) CleanupExpiredSessions(maxAge time.Duration) int {
	pv.sessionMutex.Lock()
	defer pv.sessionMutex.Unlock()

	var expiredSessions []string
	for sessionID, ctx := range pv.activeSessions {
		if ctx.IsExpired(maxAge) {
			expiredSessions = append(expiredSessions, sessionID)
		}
	}

	for _, sessionID := range expiredSessions {
		delete(pv.activeSessions, sessionID)
		pv.logger.Info("expired session cleaned up", "session_id", sessionID)
	}

	return len(expiredSessions)
}

// Helper methods

func (pv *PermissionValidator) getSession(sessionID string) (*SecurityContext, error) {
	pv.sessionMutex.RLock()
	defer pv.sessionMutex.RUnlock()

	ctx, exists := pv.activeSessions[sessionID]
	if !exists {
		return nil, fmt.Errorf("session %s not found", sessionID)
	}
	return ctx, nil
}

func (pv *PermissionValidator) isHostAllowed(host string, allowedHosts []string) bool {
	if len(allowedHosts) == 0 {
		return false
	}

	for _, allowed := range allowedHosts {
		if allowed == "*" || allowed == host {
			return true
		}
	}
	return false
}

func (pv *PermissionValidator) isPortAllowed(port int, allowedPorts []int) bool {
	if len(allowedPorts) == 0 {
		return false
	}

	for _, allowed := range allowedPorts {
		if allowed == port {
			return true
		}
	}
	return false
}

func (pv *PermissionValidator) recordViolation(sessionID string, violationType string, details map[string]interface{}, severity SecuritySeverity) {
	if pv.metrics != nil {
		pv.metrics.RecordSecurityEvent("policy_violation", map[string]interface{}{
			"session_id":     sessionID,
			"violation_type": violationType,
			"severity":       string(severity),
			"details":        details,
		})
	}

	if pv.logger != nil {
		pv.logger.Warn("security policy violation",
			"session_id", sessionID,
			"type", violationType,
			"severity", string(severity),
			"details", details,
		)
	}
}
