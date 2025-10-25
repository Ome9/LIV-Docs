// Tests for security policy management system

package security

import (
	"context"
	"io/ioutil"
	"os"
	"path/filepath"
	"testing"
	"time"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
	
	"github.com/liv-format/liv/pkg/core"
)

func TestPolicyManager_CreatePolicy(t *testing.T) {
	tempDir, err := ioutil.TempDir("", "policy-test-*")
	require.NoError(t, err)
	defer os.RemoveAll(tempDir)
	
	eventLogger := NewFileSecurityEventLogger(filepath.Join(tempDir, "security.log"))
	auditLogger := NewFileAuditLogger(filepath.Join(tempDir, "audit.log"))
	
	config := &PolicyManagerConfig{
		DefaultPolicyID:         "default",
		EnablePolicyInheritance: true,
		MaxPolicyDepth:          5,
		EnableVersioning:        true,
		AuditLogPath:           filepath.Join(tempDir, "audit.log"),
		EventLogPath:           filepath.Join(tempDir, "security.log"),
	}
	
	pm := NewPolicyManager(config, eventLogger, auditLogger)
	
	// Test creating a new policy
	policy := createTestPolicy("test-policy-1", "Test Policy 1")
	
	err = pm.CreatePolicy(context.Background(), policy, "test-user")
	assert.NoError(t, err, "Should create policy successfully")
	
	// Test retrieving the policy
	retrievedPolicy, err := pm.GetPolicy(context.Background(), "test-policy-1")
	assert.NoError(t, err, "Should retrieve policy successfully")
	assert.Equal(t, policy.ID, retrievedPolicy.ID)
	assert.Equal(t, policy.Name, retrievedPolicy.Name)
	assert.Equal(t, "test-user", retrievedPolicy.CreatedBy)
}

func TestPolicyManager_PolicyInheritance(t *testing.T) {
	tempDir, err := ioutil.TempDir("", "policy-inheritance-test-*")
	require.NoError(t, err)
	defer os.RemoveAll(tempDir)
	
	eventLogger := NewFileSecurityEventLogger(filepath.Join(tempDir, "security.log"))
	auditLogger := NewFileAuditLogger(filepath.Join(tempDir, "audit.log"))
	
	config := &PolicyManagerConfig{
		DefaultPolicyID:         "default",
		EnablePolicyInheritance: true,
		MaxPolicyDepth:          5,
	}
	
	pm := NewPolicyManager(config, eventLogger, auditLogger)
	
	// Create parent policy
	parentPolicy := createTestPolicy("parent-policy", "Parent Policy")
	err = pm.CreatePolicy(context.Background(), parentPolicy, "admin")
	require.NoError(t, err)
	
	// Create child policy with inheritance
	childPolicy := createTestPolicy("child-policy", "Child Policy")
	childPolicy.ParentPolicy = "parent-policy"
	childPolicy.SecurityPolicy.WASMPermissions = nil // Will inherit from parent
	
	err = pm.CreatePolicy(context.Background(), childPolicy, "admin")
	assert.NoError(t, err, "Should create child policy with inheritance")
	
	// Retrieve child policy and verify inheritance
	retrievedChild, err := pm.GetPolicy(context.Background(), "child-policy")
	assert.NoError(t, err, "Should retrieve child policy")
	assert.NotNil(t, retrievedChild.SecurityPolicy.WASMPermissions, "Should inherit WASM permissions from parent")
	assert.Equal(t, parentPolicy.SecurityPolicy.WASMPermissions.MemoryLimit, 
		retrievedChild.SecurityPolicy.WASMPermissions.MemoryLimit)
}

func TestPolicyManager_CircularInheritance(t *testing.T) {
	tempDir, err := ioutil.TempDir("", "circular-test-*")
	require.NoError(t, err)
	defer os.RemoveAll(tempDir)
	
	eventLogger := NewFileSecurityEventLogger(filepath.Join(tempDir, "security.log"))
	auditLogger := NewFileAuditLogger(filepath.Join(tempDir, "audit.log"))
	
	config := &PolicyManagerConfig{
		EnablePolicyInheritance: true,
		MaxPolicyDepth:          5,
	}
	
	pm := NewPolicyManager(config, eventLogger, auditLogger)
	
	// Create policies that would form a circular reference
	policy1 := createTestPolicy("policy-1", "Policy 1")
	policy2 := createTestPolicy("policy-2", "Policy 2")
	
	// Create first policy
	err = pm.CreatePolicy(context.Background(), policy1, "admin")
	require.NoError(t, err)
	
	// Create second policy with first as parent
	policy2.ParentPolicy = "policy-1"
	err = pm.CreatePolicy(context.Background(), policy2, "admin")
	require.NoError(t, err)
	
	// Try to update first policy to have second as parent (circular)
	policy1.ParentPolicy = "policy-2"
	err = pm.UpdatePolicy(context.Background(), "policy-1", policy1, "admin")
	assert.Error(t, err, "Should detect circular inheritance")
	assert.Contains(t, err.Error(), "circular")
}

func TestPolicyManager_EvaluateDocumentSecurity(t *testing.T) {
	tempDir, err := ioutil.TempDir("", "evaluation-test-*")
	require.NoError(t, err)
	defer os.RemoveAll(tempDir)
	
	eventLogger := NewFileSecurityEventLogger(filepath.Join(tempDir, "security.log"))
	auditLogger := NewFileAuditLogger(filepath.Join(tempDir, "audit.log"))
	
	config := &PolicyManagerConfig{
		DefaultPolicyID: "default",
	}
	
	pm := NewPolicyManager(config, eventLogger, auditLogger)
	
	// Create a test document
	doc := createTestDocument()
	
	// Create user context
	userContext := &UserContext{
		UserID:    "test-user",
		SessionID: "test-session",
		IPAddress: "127.0.0.1",
		Roles:     []string{"user"},
	}
	
	// Evaluate document security
	evaluation, err := pm.EvaluateDocumentSecurity(context.Background(), doc, "default", userContext)
	assert.NoError(t, err, "Should evaluate document security")
	assert.NotNil(t, evaluation, "Should return evaluation result")
	assert.Equal(t, userContext.UserID, evaluation.UserContext.UserID)
}

func TestPolicyManager_EnforceQuarantine(t *testing.T) {
	tempDir, err := ioutil.TempDir("", "quarantine-test-*")
	require.NoError(t, err)
	defer os.RemoveAll(tempDir)
	
	eventLogger := NewFileSecurityEventLogger(filepath.Join(tempDir, "security.log"))
	auditLogger := NewFileAuditLogger(filepath.Join(tempDir, "audit.log"))
	
	config := &PolicyManagerConfig{
		DefaultPolicyID: "default",
	}
	
	pm := NewPolicyManager(config, eventLogger, auditLogger)
	
	// Update default policy to enforce quarantine
	defaultPolicy, err := pm.GetPolicy(context.Background(), "default")
	require.NoError(t, err)
	
	defaultPolicy.AdminControls.EnforceQuarantine = true
	defaultPolicy.AdminControls.QuarantineDuration = 3600 // 1 hour
	
	err = pm.UpdatePolicy(context.Background(), "default", defaultPolicy, "admin")
	require.NoError(t, err)
	
	// Test quarantine enforcement
	doc := createTestDocument()
	err = pm.EnforceQuarantine(context.Background(), doc, "default", "Suspicious content detected")
	assert.NoError(t, err, "Should enforce quarantine successfully")
}

func TestSecurityEventLogger(t *testing.T) {
	tempDir, err := ioutil.TempDir("", "event-logger-test-*")
	require.NoError(t, err)
	defer os.RemoveAll(tempDir)
	
	logPath := filepath.Join(tempDir, "security.log")
	logger := NewFileSecurityEventLogger(logPath)
	
	// Test logging an event
	event := &SecurityEvent{
		ID:          "test-event-1",
		Timestamp:   time.Now(),
		EventType:   EventPolicyViolation,
		Severity:    SeverityMedium,
		Source:      "test",
		Description: "Test security event",
		Details:     map[string]interface{}{"test": "value"},
	}
	
	err = logger.LogSecurityEvent(event)
	assert.NoError(t, err, "Should log security event")
	
	// Test retrieving events
	events, err := logger.GetSecurityEvents(&EventFilter{})
	assert.NoError(t, err, "Should retrieve security events")
	assert.Len(t, events, 1, "Should have one event")
	assert.Equal(t, event.ID, events[0].ID)
}

func TestAuditLogger(t *testing.T) {
	tempDir, err := ioutil.TempDir("", "audit-logger-test-*")
	require.NoError(t, err)
	defer os.RemoveAll(tempDir)
	
	logPath := filepath.Join(tempDir, "audit.log")
	logger := NewFileAuditLogger(logPath)
	
	// Test logging an audit event
	event := &AuditEvent{
		ID:        "test-audit-1",
		Timestamp: time.Now(),
		Action:    "create_policy",
		Resource:  "test-policy",
		UserID:    "test-user",
		Success:   true,
		Details:   map[string]interface{}{"policy_name": "Test Policy"},
	}
	
	err = logger.LogAuditEvent(event)
	assert.NoError(t, err, "Should log audit event")
	
	// Test retrieving audit trail
	events, err := logger.GetAuditTrail(&AuditFilter{})
	assert.NoError(t, err, "Should retrieve audit events")
	assert.Len(t, events, 1, "Should have one event")
	assert.Equal(t, event.ID, events[0].ID)
	
	// Test CSV export
	timeRange := &TimeRange{
		Start: time.Now().Add(-1 * time.Hour),
		End:   time.Now().Add(1 * time.Hour),
	}
	
	csvData, err := logger.ExportAuditLog("csv", timeRange)
	assert.NoError(t, err, "Should export to CSV")
	assert.Contains(t, string(csvData), "timestamp,action,resource")
	assert.Contains(t, string(csvData), "create_policy")
}

// Helper functions for tests

func createTestPolicy(id, name string) *SystemSecurityPolicy {
	return &SystemSecurityPolicy{
		SecurityPolicy: &core.SecurityPolicy{
			WASMPermissions: &core.WASMPermissions{
				MemoryLimit:     8 * 1024 * 1024, // 8MB
				AllowedImports:  []string{"console"},
				CPUTimeLimit:    3000, // 3 seconds
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
		ID:          id,
		Name:        name,
		Description: "Test security policy",
		Version:     "1.0.0",
		AdminControls: &AdminControls{
			RequireApproval:      false,
			MaxDocumentSize:      5 * 1024 * 1024, // 5MB
			MaxWASMModules:       3,
			AllowedFileTypes:     []string{"text/html", "text/css"},
			RequireSignature:     false,
			EnforceQuarantine:    false,
			QuarantineDuration:   1800, // 30 minutes
		},
		ResourceLimits: &ResourceLimits{
			MaxConcurrentDocuments: 5,
			MaxMemoryPerDocument:   32 * 1024 * 1024, // 32MB
			MaxCPUTimePerDocument:  10000,            // 10 seconds
			DocumentTimeoutSeconds: 120,              // 2 minutes
		},
		ComplianceSettings: &ComplianceSettings{
			EnableGDPRCompliance:  false,
			EnableHIPAACompliance: false,
			DataRetentionDays:     30,
			RequireDataEncryption: false,
			DataClassification:    "internal",
		},
	}
}

func TestPolicyManager_GetSecurityMetrics(t *testing.T) {
	tempDir, err := ioutil.TempDir("", "metrics-test-*")
	require.NoError(t, err)
	defer os.RemoveAll(tempDir)
	
	eventLogger := NewFileSecurityEventLogger(filepath.Join(tempDir, "security.log"))
	auditLogger := NewFileAuditLogger(filepath.Join(tempDir, "audit.log"))
	
	config := &PolicyManagerConfig{
		DefaultPolicyID: "default",
	}
	
	pm := NewPolicyManager(config, eventLogger, auditLogger)
	
	// Create some test policies
	policy1 := createTestPolicy("policy-1", "Policy 1")
	policy1.ComplianceSettings.EnableGDPRCompliance = true
	err = pm.CreatePolicy(context.Background(), policy1, "admin")
	require.NoError(t, err)
	
	policy2 := createTestPolicy("policy-2", "Policy 2")
	policy2.AdminControls.RequireSignature = true
	err = pm.CreatePolicy(context.Background(), policy2, "admin")
	require.NoError(t, err)
	
	// Get security metrics
	metrics, err := pm.GetSecurityMetrics(context.Background())
	assert.NoError(t, err, "Should get security metrics")
	assert.NotNil(t, metrics, "Should return metrics")
	assert.Equal(t, 3, metrics.TotalPolicies, "Should count all policies including default")
	assert.Equal(t, 1, metrics.PolicyDistribution["gdpr"], "Should count GDPR policies")
	assert.Equal(t, 1, metrics.PolicyDistribution["signed"], "Should count signed policies")
}

func TestPolicyManager_ValidateSystemConfiguration(t *testing.T) {
	tempDir, err := ioutil.TempDir("", "validation-test-*")
	require.NoError(t, err)
	defer os.RemoveAll(tempDir)
	
	eventLogger := NewFileSecurityEventLogger(filepath.Join(tempDir, "security.log"))
	auditLogger := NewFileAuditLogger(filepath.Join(tempDir, "audit.log"))
	
	config := &PolicyManagerConfig{
		DefaultPolicyID: "default",
	}
	
	pm := NewPolicyManager(config, eventLogger, auditLogger)
	
	// Create a policy with issues
	policy := createTestPolicy("problematic-policy", "Problematic Policy")
	policy.EventConfig.EnableAuditLog = false // This should trigger an issue
	policy.SecurityPolicy.WASMPermissions.MemoryLimit = 128 * 1024 * 1024 // Overly permissive
	policy.AdminControls.RequireSignature = false // Overly permissive
	
	err = pm.CreatePolicy(context.Background(), policy, "admin")
	require.NoError(t, err)
	
	// Validate system configuration
	report, err := pm.ValidateSystemConfiguration(context.Background())
	assert.NoError(t, err, "Should validate system configuration")
	assert.NotNil(t, report, "Should return validation report")
	assert.Greater(t, len(report.Issues), 0, "Should detect configuration issues")
	
	// Check for specific issues
	hasAuditIssue := false
	hasPermissiveIssue := false
	for _, issue := range report.Issues {
		if issue.Type == "missing_audit_logging" {
			hasAuditIssue = true
		}
		if issue.Type == "overly_permissive_policy" {
			hasPermissiveIssue = true
		}
	}
	assert.True(t, hasAuditIssue, "Should detect missing audit logging")
	assert.True(t, hasPermissiveIssue, "Should detect overly permissive policy")
}

func TestPolicyManager_MonitorResourceUsage(t *testing.T) {
	tempDir, err := ioutil.TempDir("", "resource-test-*")
	require.NoError(t, err)
	defer os.RemoveAll(tempDir)
	
	eventLogger := NewFileSecurityEventLogger(filepath.Join(tempDir, "security.log"))
	auditLogger := NewFileAuditLogger(filepath.Join(tempDir, "audit.log"))
	
	config := &PolicyManagerConfig{
		DefaultPolicyID: "default",
	}
	
	pm := NewPolicyManager(config, eventLogger, auditLogger)
	
	// Create resource metrics that exceed limits
	resourceMetrics := &ResourceMetrics{
		MemoryUsage:         100 * 1024 * 1024, // 100MB
		CPUTime:             60000,              // 60 seconds
		ConcurrentDocuments: 20,
		NetworkBandwidth:    2 * 1024 * 1024, // 2MB/s
		StorageUsage:        200 * 1024 * 1024, // 200MB
	}
	
	// Monitor resource usage
	report, err := pm.MonitorResourceUsage(context.Background(), resourceMetrics)
	assert.NoError(t, err, "Should monitor resource usage")
	assert.NotNil(t, report, "Should return monitoring report")
	
	// Should detect violations against default policy limits
	assert.Greater(t, len(report.Violations), 0, "Should detect resource violations")
	assert.Equal(t, "violations_detected", report.OverallStatus, "Should indicate violations detected")
}

func TestPolicyManager_CreatePolicyFromTemplate(t *testing.T) {
	tempDir, err := ioutil.TempDir("", "template-test-*")
	require.NoError(t, err)
	defer os.RemoveAll(tempDir)
	
	eventLogger := NewFileSecurityEventLogger(filepath.Join(tempDir, "security.log"))
	auditLogger := NewFileAuditLogger(filepath.Join(tempDir, "audit.log"))
	
	config := &PolicyManagerConfig{
		DefaultPolicyID: "default",
	}
	
	pm := NewPolicyManager(config, eventLogger, auditLogger)
	
	// Create policy from template
	variables := map[string]interface{}{
		"memory_limit":       int64(32 * 1024 * 1024), // 32MB
		"max_document_size":  int64(20 * 1024 * 1024), // 20MB
		"require_signature":  true,
	}
	
	err = pm.CreatePolicyFromTemplate(context.Background(), "high-security", "test-from-template", variables, "admin")
	assert.NoError(t, err, "Should create policy from template")
	
	// Verify the created policy
	policy, err := pm.GetPolicy(context.Background(), "test-from-template")
	assert.NoError(t, err, "Should retrieve created policy")
	assert.Equal(t, uint64(32*1024*1024), policy.SecurityPolicy.WASMPermissions.MemoryLimit, "Should apply memory limit variable")
	assert.Equal(t, int64(20*1024*1024), policy.AdminControls.MaxDocumentSize, "Should apply document size variable")
	assert.True(t, policy.AdminControls.RequireSignature, "Should apply signature requirement variable")
}

func createTestDocument() *core.LIVDocument {
	return &core.LIVDocument{
		Manifest: &core.Manifest{
			Version: "1.0.0",
			Metadata: &core.DocumentMetadata{
				Title:       "Test Document",
				Author:      "Test Author",
				Created:     time.Now(),
				Modified:    time.Now(),
				Description: "Test document for security evaluation",
				Version:     "1.0.0",
				Language:    "en",
			},
			Security: &core.SecurityPolicy{
				WASMPermissions: &core.WASMPermissions{
					MemoryLimit:     4 * 1024 * 1024,
					AllowedImports:  []string{"console"},
					CPUTimeLimit:    2000,
					AllowNetworking: false,
					AllowFileSystem: false,
				},
			},
			Resources: map[string]*core.Resource{
				"style.css": {
					Hash: "test-hash",
					Size: 1024,
					Type: "text/css",
					Path: "style.css",
				},
			},
		},
		Content: &core.DocumentContent{
			HTML:           "<html><body><h1>Test</h1></body></html>",
			CSS:            "body { font-family: Arial; }",
			InteractiveSpec: "{}",
			StaticFallback: "<html><body><h1>Static Test</h1></body></html>",
		},
		Assets: &core.AssetBundle{
			Images: map[string][]byte{},
			Fonts:  map[string][]byte{},
			Data:   map[string][]byte{},
		},
		Signatures: &core.SignatureBundle{
			ContentSignature:  "test-signature",
			ManifestSignature: "test-signature",
			WASMSignatures:    map[string]string{},
		},
		WASMModules: map[string][]byte{
			"test-module": make([]byte, 1024), // 1KB test module
		},
	}
}