package security

import (
	"context"
	"testing"
	"time"

	"github.com/liv-format/liv/pkg/core"
)

// SimpleMockLogger implements core.Logger for testing
type SimpleMockLogger struct {
	logs []LogEntry
}

type LogEntry struct {
	Level   string
	Message string
	Fields  []interface{}
}

func (ml *SimpleMockLogger) Debug(msg string, fields ...interface{}) {
	ml.logs = append(ml.logs, LogEntry{"DEBUG", msg, fields})
}

func (ml *SimpleMockLogger) Info(msg string, fields ...interface{}) {
	ml.logs = append(ml.logs, LogEntry{"INFO", msg, fields})
}

func (ml *SimpleMockLogger) Warn(msg string, fields ...interface{}) {
	ml.logs = append(ml.logs, LogEntry{"WARN", msg, fields})
}

func (ml *SimpleMockLogger) Error(msg string, fields ...interface{}) {
	ml.logs = append(ml.logs, LogEntry{"ERROR", msg, fields})
}

func (ml *SimpleMockLogger) Fatal(msg string, fields ...interface{}) {
	ml.logs = append(ml.logs, LogEntry{"FATAL", msg, fields})
}

// SimpleMockMetricsCollector implements core.MetricsCollector for testing
type SimpleMockMetricsCollector struct {
	events []SecurityEventRecord
}

type SecurityEventRecord struct {
	Type    string
	Details map[string]interface{}
}

func (mmc *SimpleMockMetricsCollector) RecordDocumentLoad(size int64, duration int64) {}

func (mmc *SimpleMockMetricsCollector) RecordWASMExecution(module string, duration int64, memoryUsed uint64) {
}

func (mmc *SimpleMockMetricsCollector) RecordSecurityEvent(eventType string, details map[string]interface{}) {
	mmc.events = append(mmc.events, SecurityEventRecord{eventType, details})
}

func (mmc *SimpleMockMetricsCollector) GetMetrics() map[string]interface{} {
	return map[string]interface{}{
		"events": mmc.events,
	}
}

func TestNewPolicyEngine(t *testing.T) {
	logger := &SimpleMockLogger{}
	metrics := &SimpleMockMetricsCollector{}

	pe := NewPolicyEngine(logger, metrics)

	if pe == nil {
		t.Fatal("NewPolicyEngine returned nil")
	}

	if pe.defaultPolicy == nil {
		t.Error("default policy should not be nil")
	}

	if pe.logger != logger {
		t.Error("logger not set correctly")
	}

	if pe.metrics != metrics {
		t.Error("metrics collector not set correctly")
	}
}

func TestEvaluateWASMPermissions_ValidRequest(t *testing.T) {
	logger := &SimpleMockLogger{}
	metrics := &SimpleMockMetricsCollector{}
	pe := NewPolicyEngine(logger, metrics)

	policy := &core.SecurityPolicy{
		WASMPermissions: &core.WASMPermissions{
			MemoryLimit:     8 * 1024 * 1024, // 8MB
			CPUTimeLimit:    5000,            // 5 seconds
			AllowNetworking: true,
			AllowFileSystem: false,
			AllowedImports:  []string{"env.memory", "env.table"},
		},
	}

	requested := &core.WASMPermissions{
		MemoryLimit:     4 * 1024 * 1024, // 4MB
		CPUTimeLimit:    2000,            // 2 seconds
		AllowNetworking: true,
		AllowFileSystem: false,
		AllowedImports:  []string{"env.memory"},
	}

	result := pe.EvaluateWASMPermissions(requested, policy)

	if !result.Allowed {
		t.Errorf("expected permissions to be allowed, got denied: %v", result.Errors)
	}

	if len(result.Errors) > 0 {
		t.Errorf("expected no errors, got: %v", result.Errors)
	}

	// Should have warnings for networking
	if len(result.Warnings) == 0 {
		t.Error("expected warnings for networking access")
	}
}

func TestEvaluateWASMPermissions_ExceedsMemoryLimit(t *testing.T) {
	logger := &SimpleMockLogger{}
	metrics := &SimpleMockMetricsCollector{}
	pe := NewPolicyEngine(logger, metrics)

	policy := &core.SecurityPolicy{
		WASMPermissions: &core.WASMPermissions{
			MemoryLimit:     4 * 1024 * 1024, // 4MB
			CPUTimeLimit:    5000,
			AllowNetworking: false,
			AllowFileSystem: false,
			AllowedImports:  []string{},
		},
	}

	requested := &core.WASMPermissions{
		MemoryLimit:     8 * 1024 * 1024, // 8MB - exceeds limit
		CPUTimeLimit:    2000,
		AllowNetworking: false,
		AllowFileSystem: false,
		AllowedImports:  []string{},
	}

	result := pe.EvaluateWASMPermissions(requested, policy)

	if result.Allowed {
		t.Error("expected permissions to be denied due to memory limit")
	}

	if len(result.Errors) == 0 {
		t.Error("expected errors for memory limit violation")
	}

	// Check that the error message mentions memory limit
	found := false
	for _, err := range result.Errors {
		if len(err) > 0 && err[0:len("requested memory limit")] == "requested memory limit" {
			found = true
			break
		}
	}
	if !found {
		t.Errorf("expected memory limit error, got: %v", result.Errors)
	}
}

func TestEvaluateWASMPermissions_UnauthorizedImport(t *testing.T) {
	logger := &SimpleMockLogger{}
	metrics := &SimpleMockMetricsCollector{}
	pe := NewPolicyEngine(logger, metrics)

	policy := &core.SecurityPolicy{
		WASMPermissions: &core.WASMPermissions{
			MemoryLimit:     4 * 1024 * 1024,
			CPUTimeLimit:    5000,
			AllowNetworking: false,
			AllowFileSystem: false,
			AllowedImports:  []string{"env.memory"}, // Only allow env.memory
		},
	}

	requested := &core.WASMPermissions{
		MemoryLimit:     2 * 1024 * 1024,
		CPUTimeLimit:    2000,
		AllowNetworking: false,
		AllowFileSystem: false,
		AllowedImports:  []string{"env.memory", "env.filesystem"}, // Unauthorized import
	}

	result := pe.EvaluateWASMPermissions(requested, policy)

	if result.Allowed {
		t.Error("expected permissions to be denied due to unauthorized import")
	}

	if len(result.Errors) == 0 {
		t.Error("expected errors for unauthorized import")
	}
}

func TestEvaluateWASMPermissions_NilRequest(t *testing.T) {
	logger := &SimpleMockLogger{}
	metrics := &SimpleMockMetricsCollector{}
	pe := NewPolicyEngine(logger, metrics)

	policy := &core.SecurityPolicy{
		WASMPermissions: &core.WASMPermissions{
			MemoryLimit:     4 * 1024 * 1024,
			CPUTimeLimit:    5000,
			AllowNetworking: false,
			AllowFileSystem: false,
			AllowedImports:  []string{},
		},
	}

	result := pe.EvaluateWASMPermissions(nil, policy)

	if result.Allowed {
		t.Error("expected permissions to be denied for nil request")
	}

	if len(result.Errors) == 0 {
		t.Error("expected errors for nil request")
	}
}

func TestValidateWASMModule_ValidModule(t *testing.T) {
	logger := &SimpleMockLogger{}
	metrics := &SimpleMockMetricsCollector{}
	pe := NewPolicyEngine(logger, metrics)

	// Valid WASM magic number and version
	moduleData := []byte{0x00, 0x61, 0x73, 0x6d, 0x01, 0x00, 0x00, 0x00}

	permissions := &core.WASMPermissions{
		MemoryLimit: 16 * 1024 * 1024, // 16MB
	}

	result := pe.ValidateWASMModule(moduleData, permissions)

	if !result.IsValid {
		t.Errorf("expected module to be valid, got errors: %v", result.Errors)
	}
}

func TestValidateWASMModule_InvalidMagicNumber(t *testing.T) {
	logger := &SimpleMockLogger{}
	metrics := &SimpleMockMetricsCollector{}
	pe := NewPolicyEngine(logger, metrics)

	// Invalid magic number
	moduleData := []byte{0x00, 0x00, 0x00, 0x00, 0x01, 0x00, 0x00, 0x00}

	permissions := &core.WASMPermissions{
		MemoryLimit: 16 * 1024 * 1024,
	}

	result := pe.ValidateWASMModule(moduleData, permissions)

	if result.IsValid {
		t.Error("expected module to be invalid due to bad magic number")
	}

	if len(result.Errors) == 0 {
		t.Error("expected errors for invalid magic number")
	}
}

func TestValidateWASMModule_EmptyModule(t *testing.T) {
	logger := &SimpleMockLogger{}
	metrics := &SimpleMockMetricsCollector{}
	pe := NewPolicyEngine(logger, metrics)

	result := pe.ValidateWASMModule([]byte{}, nil)

	if result.IsValid {
		t.Error("expected empty module to be invalid")
	}

	if len(result.Errors) == 0 {
		t.Error("expected errors for empty module")
	}
}

func TestEnforceResourceLimits(t *testing.T) {
	logger := &SimpleMockLogger{}
	metrics := &SimpleMockMetricsCollector{}
	pe := NewPolicyEngine(logger, metrics)

	permissions := &core.WASMPermissions{
		MemoryLimit:     4 * 1024 * 1024, // 4MB
		CPUTimeLimit:    5000,            // 5 seconds
		AllowNetworking: true,
		AllowFileSystem: false,
		AllowedImports:  []string{"env.memory"},
	}

	constraints := pe.EnforceResourceLimits(permissions)

	if constraints == nil {
		t.Fatal("EnforceResourceLimits returned nil")
	}

	if constraints.MemoryLimit != int64(permissions.MemoryLimit) {
		t.Errorf("expected memory limit %d, got %d", permissions.MemoryLimit, constraints.MemoryLimit)
	}

	expectedCPULimit := time.Duration(permissions.CPUTimeLimit) * time.Millisecond
	if constraints.CPUTimeLimit != expectedCPULimit {
		t.Errorf("expected CPU limit %v, got %v", expectedCPULimit, constraints.CPUTimeLimit)
	}

	if constraints.AllowNetworking != permissions.AllowNetworking {
		t.Errorf("expected networking %v, got %v", permissions.AllowNetworking, constraints.AllowNetworking)
	}

	if constraints.AllowFileSystem != permissions.AllowFileSystem {
		t.Errorf("expected filesystem %v, got %v", permissions.AllowFileSystem, constraints.AllowFileSystem)
	}
}

func TestCreateSecurityContext(t *testing.T) {
	logger := &SimpleMockLogger{}
	metrics := &SimpleMockMetricsCollector{}
	pe := NewPolicyEngine(logger, metrics)

	policy := &core.SecurityPolicy{
		WASMPermissions: &core.WASMPermissions{
			MemoryLimit:     4 * 1024 * 1024,
			CPUTimeLimit:    5000,
			AllowNetworking: false,
			AllowFileSystem: false,
			AllowedImports:  []string{},
		},
	}

	ctx := pe.CreateSecurityContext(policy)

	if ctx == nil {
		t.Fatal("CreateSecurityContext returned nil")
	}

	if ctx.Policy != policy {
		t.Error("security context policy not set correctly")
	}

	if ctx.Constraints == nil {
		t.Error("security context constraints not set")
	}

	if ctx.SessionID == "" {
		t.Error("security context session ID not set")
	}

	if ctx.CreatedAt.IsZero() {
		t.Error("security context creation time not set")
	}
}

func TestValidatePermissionRequest_MemoryRequest(t *testing.T) {
	logger := &SimpleMockLogger{}
	metrics := &SimpleMockMetricsCollector{}
	pe := NewPolicyEngine(logger, metrics)

	policy := &core.SecurityPolicy{
		WASMPermissions: &core.WASMPermissions{
			MemoryLimit: 8 * 1024 * 1024, // 8MB
		},
	}

	securityCtx := pe.CreateSecurityContext(policy)

	request := &PermissionRequest{
		Type: string(PermissionTypeMemory),
		RequestedPerms: map[string]interface{}{
			"size": uint64(4 * 1024 * 1024), // 4MB - within limit
		},
		RequestedAt: time.Now(),
	}

	response := pe.ValidatePermissionRequest(context.Background(), request, securityCtx)

	if !response.Granted {
		t.Errorf("expected memory request to be granted, got denied: %s", response.Reason)
	}
}

func TestValidatePermissionRequest_ExcessiveMemoryRequest(t *testing.T) {
	logger := &SimpleMockLogger{}
	metrics := &SimpleMockMetricsCollector{}
	pe := NewPolicyEngine(logger, metrics)

	policy := &core.SecurityPolicy{
		WASMPermissions: &core.WASMPermissions{
			MemoryLimit: 4 * 1024 * 1024, // 4MB
		},
	}

	securityCtx := pe.CreateSecurityContext(policy)

	request := &PermissionRequest{
		Type: string(PermissionTypeMemory),
		RequestedPerms: map[string]interface{}{
			"size": uint64(8 * 1024 * 1024), // 8MB - exceeds limit
		},
		RequestedAt: time.Now(),
	}

	response := pe.ValidatePermissionRequest(context.Background(), request, securityCtx)

	if response.Granted {
		t.Error("expected excessive memory request to be denied")
	}

	if response.Reason == "" {
		t.Error("expected reason for denial")
	}
}

func TestValidatePermissionRequest_NetworkRequest(t *testing.T) {
	logger := &SimpleMockLogger{}
	metrics := &SimpleMockMetricsCollector{}
	pe := NewPolicyEngine(logger, metrics)

	policy := &core.SecurityPolicy{
		WASMPermissions: &core.WASMPermissions{
			AllowNetworking: true,
		},
	}

	securityCtx := pe.CreateSecurityContext(policy)

	request := &PermissionRequest{
		Type: string(PermissionTypeNetwork),
		RequestedPerms: map[string]interface{}{
			"host": "example.com",
			"port": 443,
		},
		RequestedAt: time.Now(),
	}

	response := pe.ValidatePermissionRequest(context.Background(), request, securityCtx)

	if !response.Granted {
		t.Errorf("expected network request to be granted, got denied: %s", response.Reason)
	}
}

func TestValidatePermissionRequest_DeniedNetworkRequest(t *testing.T) {
	logger := &SimpleMockLogger{}
	metrics := &SimpleMockMetricsCollector{}
	pe := NewPolicyEngine(logger, metrics)

	policy := &core.SecurityPolicy{
		WASMPermissions: &core.WASMPermissions{
			AllowNetworking: false, // Networking disabled
		},
	}

	securityCtx := pe.CreateSecurityContext(policy)

	request := &PermissionRequest{
		Type: string(PermissionTypeNetwork),
		RequestedPerms: map[string]interface{}{
			"host": "example.com",
			"port": 443,
		},
		RequestedAt: time.Now(),
	}

	response := pe.ValidatePermissionRequest(context.Background(), request, securityCtx)

	if response.Granted {
		t.Error("expected network request to be denied when networking is disabled")
	}
}

func TestMetricsRecording(t *testing.T) {
	logger := &SimpleMockLogger{}
	metrics := &SimpleMockMetricsCollector{}
	pe := NewPolicyEngine(logger, metrics)

	policy := &core.SecurityPolicy{
		WASMPermissions: &core.WASMPermissions{
			MemoryLimit: 4 * 1024 * 1024,
		},
	}

	requested := &core.WASMPermissions{
		MemoryLimit: 2 * 1024 * 1024,
	}

	// This should trigger metrics recording
	pe.EvaluateWASMPermissions(requested, policy)

	if len(metrics.events) == 0 {
		t.Error("expected metrics to be recorded")
	}

	// Check that the right event type was recorded
	found := false
	for _, event := range metrics.events {
		if event.Type == "permission_evaluation" {
			found = true
			break
		}
	}
	if !found {
		t.Error("expected permission_evaluation event to be recorded")
	}
}
