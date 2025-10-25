// Comprehensive security and administration tests
// Tests security policy enforcement, permission management, event handling, and audit logging

package security

import (
	"context"
	"encoding/json"
	"fmt"
	"io/ioutil"
	"net/http"
	"net/http/httptest"
	"os"
	"path/filepath"
	"testing"
	"time"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
	"github.com/stretchr/testify/suite"

	"github.com/liv-format/liv/pkg/core"
)

// SecurityAdministrationTestSuite provides comprehensive security and administration testing
type SecurityAdministrationTestSuite struct {
	suite.Suite
	tempDir           string
	policyManager     *PolicyManager
	permissionManager *PermissionManager
	eventLogger       *FileSecurityEventLogger
	auditLogger       *FileAuditLogger
	mockSM            *MockSecurityManager
	mockCP            *MockCryptoProvider
	mockLogger        *MockLogger
}

// SetupSuite initializes the test suite
func (suite *SecurityAdministrationTestSuite) SetupSuite() {
	var err error
	suite.tempDir, err = ioutil.TempDir("", "security-admin-test-*")
	suite.Require().NoError(err)

	// Create loggers
	suite.eventLogger = NewFileSecurityEventLogger(filepath.Join(suite.tempDir, "security-events.log"))
	suite.auditLogger = NewFileAuditLogger(filepath.Join(suite.tempDir, "audit.log"))

	// Create mocks
	suite.mockSM = &MockSecurityManager{}
	suite.mockCP = &MockCryptoProvider{}
	suite.mockLogger = &MockLogger{}

	// Create policy manager
	config := &PolicyManagerConfig{
		DefaultPolicyID:         "default",
		EnablePolicyInheritance: true,
		MaxPolicyDepth:          5,
		EnableVersioning:        true,
		AuditLogPath:            filepath.Join(suite.tempDir, "audit.log"),
		EventLogPath:            filepath.Join(suite.tempDir, "security-events.log"),
	}
	suite.policyManager = NewPolicyManager(config, suite.eventLogger, suite.auditLogger)

	// Create permission manager
	suite.permissionManager = NewPermissionManager(suite.policyManager, suite.mockSM, suite.mockCP, suite.mockLogger)

	// Create test policies
	suite.createTestPolicies()
}

// TearDownSuite cleans up the test suite
func (suite *SecurityAdministrationTestSuite) TearDownSuite() {
	os.RemoveAll(suite.tempDir)
}

// SetupTest prepares each test
func (suite *SecurityAdministrationTestSuite) SetupTest() {
	// Reset mocks for each test
	suite.mockSM.ExpectedCalls = nil
	suite.mockCP.ExpectedCalls = nil
	suite.mockLogger.ExpectedCalls = nil
}

// createTestPolicies creates test security policies
func (suite *SecurityAdministrationTestSuite) createTestPolicies() {
	ctx := context.Background()

	// Create basic policy
	basicPolicy := createTestPolicy("basic-policy", "Basic Policy")
	err := suite.policyManager.CreatePolicy(ctx, basicPolicy, "admin")
	suite.Require().NoError(err)

	// Create strict policy
	strictPolicy := createTestPolicy("strict-policy", "Strict Policy")
	strictPolicy.SecurityPolicy.WASMPermissions.MemoryLimit = 2 * 1024 * 1024 // 2MB
	strictPolicy.SecurityPolicy.WASMPermissions.CPUTimeLimit = 1000           // 1 second
	strictPolicy.AdminControls.RequireSignature = true
	strictPolicy.AdminControls.EnforceQuarantine = true
	err = suite.policyManager.CreatePolicy(ctx, strictPolicy, "admin")
	suite.Require().NoError(err)

	// Create permissive policy
	permissivePolicy := createTestPolicy("permissive-policy", "Permissive Policy")
	permissivePolicy.SecurityPolicy.WASMPermissions.MemoryLimit = 64 * 1024 * 1024 // 64MB
	permissivePolicy.SecurityPolicy.WASMPermissions.CPUTimeLimit = 30000           // 30 seconds
	permissivePolicy.SecurityPolicy.WASMPermissions.AllowNetworking = true
	permissivePolicy.SecurityPolicy.WASMPermissions.AllowFileSystem = true
	err = suite.policyManager.CreatePolicy(ctx, permissivePolicy, "admin")
	suite.Require().NoError(err)

	// Create child policy with inheritance
	childPolicy := createTestPolicy("child-policy", "Child Policy")
	childPolicy.ParentPolicy = "basic-policy"
	err = suite.policyManager.CreatePolicy(ctx, childPolicy, "admin")
	suite.Require().NoError(err)
}

// TestSecurityPolicyEnforcement tests security policy enforcement
func (suite *SecurityAdministrationTestSuite) TestSecurityPolicyEnforcement() {
	ctx := context.Background()

	// Test strict policy enforcement
	strictRequest := &PermissionRequest{
		DocumentID: "test-doc-strict",
		RequestedPerms: &core.WASMPermissions{
			MemoryLimit:     8 * 1024 * 1024, // Exceeds strict policy limit
			CPUTimeLimit:    5000,            // Exceeds strict policy limit
			AllowNetworking: false,
			AllowFileSystem: false,
			AllowedImports:  []string{"console"},
		},
		PolicyID:      "strict-policy",
		UserContext:   &UserContext{UserID: "test-user"},
		Justification: "Testing strict enforcement",
		RequestedAt:   time.Now(),
	}

	suite.mockSM.On("EvaluatePermissions", strictRequest.RequestedPerms, suite.getPolicy("strict-policy").SecurityPolicy).Return(false)
	suite.mockLogger.On("Info", "Permission evaluation completed",
		"document_id", "test-doc-strict",
		"policy_id", "strict-policy",
		"granted", false,
		"warnings", 0,
	).Return()

	evaluation, err := suite.permissionManager.EvaluatePermissionRequest(ctx, strictRequest)
	suite.NoError(err)
	suite.False(evaluation.Granted, "Strict policy should deny excessive permissions")
	suite.Empty(evaluation.InheritedFrom, "Should not inherit when policy denies")

	// Test permissive policy enforcement
	permissiveRequest := &PermissionRequest{
		DocumentID: "test-doc-permissive",
		RequestedPerms: &core.WASMPermissions{
			MemoryLimit:     32 * 1024 * 1024, // Within permissive policy limit
			CPUTimeLimit:    15000,            // Within permissive policy limit
			AllowNetworking: true,             // Allowed by permissive policy
			AllowFileSystem: true,             // Allowed by permissive policy
			AllowedImports:  []string{"console", "dom"},
		},
		PolicyID:      "permissive-policy",
		UserContext:   &UserContext{UserID: "test-user"},
		Justification: "Testing permissive enforcement",
		RequestedAt:   time.Now(),
	}

	suite.mockSM.On("EvaluatePermissions", permissiveRequest.RequestedPerms, suite.getPolicy("permissive-policy").SecurityPolicy).Return(true)
	suite.mockLogger.On("Info", "Permission evaluation completed",
		"document_id", "test-doc-permissive",
		"policy_id", "permissive-policy",
		"granted", true,
		"warnings", 4, // Should have warnings for high memory, long CPU, network, filesystem
	).Return()

	evaluation, err = suite.permissionManager.EvaluatePermissionRequest(ctx, permissiveRequest)
	suite.NoError(err)
	suite.True(evaluation.Granted, "Permissive policy should grant reasonable permissions")
	suite.Greater(len(evaluation.Warnings), 0, "Should generate security warnings for risky permissions")

	suite.mockSM.AssertExpectations(suite.T())
	suite.mockLogger.AssertExpectations(suite.T())
}

// TestPermissionInheritanceEnforcement tests permission inheritance enforcement
func (suite *SecurityAdministrationTestSuite) TestPermissionInheritanceEnforcement() {
	ctx := context.Background()

	// Test inheritance when child policy denies but parent allows
	inheritanceRequest := &PermissionRequest{
		DocumentID: "test-doc-inheritance",
		RequestedPerms: &core.WASMPermissions{
			MemoryLimit:     8 * 1024 * 1024,
			CPUTimeLimit:    3000,
			AllowNetworking: false,
			AllowFileSystem: false,
			AllowedImports:  []string{"console"},
		},
		PolicyID:      "child-policy",
		UserContext:   &UserContext{UserID: "test-user"},
		Justification: "Testing inheritance",
		RequestedAt:   time.Now(),
	}

	// Child policy denies, parent policy allows
	suite.mockSM.On("EvaluatePermissions", inheritanceRequest.RequestedPerms, suite.getPolicy("child-policy").SecurityPolicy).Return(false)
	suite.mockSM.On("EvaluatePermissions", inheritanceRequest.RequestedPerms, suite.getPolicy("basic-policy").SecurityPolicy).Return(true)
	suite.mockLogger.On("Info", "Permission evaluation completed",
		"document_id", "test-doc-inheritance",
		"policy_id", "child-policy",
		"granted", true,
		"warnings", 1, // Should have inheritance warning
	).Return()

	evaluation, err := suite.permissionManager.EvaluatePermissionRequest(ctx, inheritanceRequest)
	suite.NoError(err)
	suite.True(evaluation.Granted, "Should grant permissions through inheritance")
	suite.Equal("basic-policy", evaluation.InheritedFrom, "Should inherit from parent policy")
	suite.Greater(len(evaluation.Warnings), 0, "Should have inheritance warning")

	// Verify inheritance warning
	hasInheritanceWarning := false
	for _, warning := range evaluation.Warnings {
		if warning.Type == "inherited_permissions" {
			hasInheritanceWarning = true
			break
		}
	}
	suite.True(hasInheritanceWarning, "Should have inheritance warning")

	suite.mockSM.AssertExpectations(suite.T())
	suite.mockLogger.AssertExpectations(suite.T())
}

// TestSecurityEventHandling tests security event handling and logging
func (suite *SecurityAdministrationTestSuite) TestSecurityEventHandling() {
	ctx := context.Background()

	// Test security event generation during policy violation
	violationRequest := &PermissionRequest{
		DocumentID: "test-doc-violation",
		RequestedPerms: &core.WASMPermissions{
			MemoryLimit:     128 * 1024 * 1024, // Excessive memory
			CPUTimeLimit:    60000,             // Excessive CPU time
			AllowNetworking: true,
			AllowFileSystem: true,
			AllowedImports:  []string{"console", "dom", "fetch", "filesystem"},
		},
		PolicyID:      "strict-policy",
		UserContext:   &UserContext{UserID: "test-user", IPAddress: "192.168.1.100"},
		Justification: "Testing violation handling",
		RequestedAt:   time.Now(),
	}

	suite.mockSM.On("EvaluatePermissions", violationRequest.RequestedPerms, suite.getPolicy("strict-policy").SecurityPolicy).Return(false)
	suite.mockLogger.On("Info", "Permission evaluation completed",
		"document_id", "test-doc-violation",
		"policy_id", "strict-policy",
		"granted", false,
		"warnings", 0,
	).Return()

	evaluation, err := suite.permissionManager.EvaluatePermissionRequest(ctx, violationRequest)
	suite.NoError(err)
	suite.False(evaluation.Granted, "Should deny excessive permissions")

	// Verify security events were logged
	events, err := suite.eventLogger.GetSecurityEvents(&EventFilter{
		EventTypes: []SecurityEventType{EventPolicyViolation},
	})
	suite.NoError(err)

	// Should have events from policy creation and evaluation
	suite.Greater(len(events), 0, "Should have logged security events")

	// Test resource monitoring violation
	resourceMetrics := &ResourceMetrics{
		MemoryUsage:         100 * 1024 * 1024, // 100MB - exceeds strict policy
		CPUTime:             30000,             // 30 seconds - exceeds strict policy
		ConcurrentDocuments: 10,                // Exceeds strict policy
	}

	report, err := suite.policyManager.MonitorResourceUsage(ctx, resourceMetrics)
	suite.NoError(err)
	suite.Greater(len(report.Violations), 0, "Should detect resource violations")
	suite.Equal("violations_detected", report.OverallStatus, "Should indicate violations")

	// Verify resource violation events were logged
	resourceEvents, err := suite.eventLogger.GetSecurityEvents(&EventFilter{
		EventTypes: []SecurityEventType{EventResourceExceeded},
	})
	suite.NoError(err)
	suite.Greater(len(resourceEvents), 0, "Should have logged resource violation events")

	suite.mockSM.AssertExpectations(suite.T())
	suite.mockLogger.AssertExpectations(suite.T())
}

// TestAuditLogging tests audit logging functionality
func (suite *SecurityAdministrationTestSuite) TestAuditLogging() {
	ctx := context.Background()

	// Test policy creation audit logging
	newPolicy := createTestPolicy("audit-test-policy", "Audit Test Policy")
	err := suite.policyManager.CreatePolicy(ctx, newPolicy, "audit-admin")
	suite.NoError(err)

	// Verify audit log entry
	auditEvents, err := suite.auditLogger.GetAuditTrail(&AuditFilter{
		Actions: []string{"create_policy"},
		UserID:  "audit-admin",
	})
	suite.NoError(err)
	suite.Greater(len(auditEvents), 0, "Should have audit log entries for policy creation")

	// Find the specific audit event
	var createEvent *AuditEvent
	for _, event := range auditEvents {
		if event.Action == "create_policy" && event.Resource == "audit-test-policy" {
			createEvent = event
			break
		}
	}
	suite.NotNil(createEvent, "Should have audit event for policy creation")
	suite.Equal("audit-admin", createEvent.UserID, "Should record correct user")
	suite.True(createEvent.Success, "Should record successful operation")

	// Test policy update audit logging
	newPolicy.Description = "Updated description for audit testing"
	err = suite.policyManager.UpdatePolicy(ctx, "audit-test-policy", newPolicy, "audit-admin")
	suite.NoError(err)

	// Verify update audit log entry
	updateEvents, err := suite.auditLogger.GetAuditTrail(&AuditFilter{
		Actions: []string{"update_policy"},
		UserID:  "audit-admin",
	})
	suite.NoError(err)
	suite.Greater(len(updateEvents), 0, "Should have audit log entries for policy update")

	// Test audit log export
	timeRange := &TimeRange{
		Start: time.Now().Add(-1 * time.Hour),
		End:   time.Now().Add(1 * time.Hour),
	}

	csvData, err := suite.auditLogger.ExportAuditLog("csv", timeRange)
	suite.NoError(err)
	suite.Contains(string(csvData), "create_policy", "CSV export should contain policy creation")
	suite.Contains(string(csvData), "update_policy", "CSV export should contain policy update")

	jsonData, err := suite.auditLogger.ExportAuditLog("json", timeRange)
	suite.NoError(err)

	var exportedEvents []*AuditEvent
	err = json.Unmarshal(jsonData, &exportedEvents)
	suite.NoError(err)
	suite.Greater(len(exportedEvents), 0, "JSON export should contain audit events")
}

// TestSystemValidationAndMetrics tests system validation and security metrics
func (suite *SecurityAdministrationTestSuite) TestSystemValidationAndMetrics() {
	ctx := context.Background()

	// Test system configuration validation
	report, err := suite.policyManager.ValidateSystemConfiguration(ctx)
	suite.NoError(err)
	suite.NotNil(report, "Should return validation report")
	suite.Greater(report.TotalPolicies, 0, "Should count existing policies")

	// Check for validation issues
	if len(report.Issues) > 0 {
		// Verify issue types are valid
		validIssueTypes := map[string]bool{
			"missing_default_policy":   true,
			"overly_permissive_policy": true,
			"missing_audit_logging":    true,
		}

		for _, issue := range report.Issues {
			suite.True(validIssueTypes[issue.Type], "Should have valid issue type: %s", issue.Type)
			suite.NotEmpty(issue.Description, "Issue should have description")
			suite.NotEmpty(issue.Recommendation, "Issue should have recommendation")
		}
	}

	// Test security metrics
	metrics, err := suite.policyManager.GetSecurityMetrics(ctx)
	suite.NoError(err)
	suite.NotNil(metrics, "Should return security metrics")
	suite.Greater(metrics.TotalPolicies, 0, "Should count policies")
	suite.GreaterOrEqual(metrics.ComplianceScore, 0.0, "Compliance score should be non-negative")
	suite.LessOrEqual(metrics.ComplianceScore, 100.0, "Compliance score should not exceed 100")
	suite.Contains([]string{"low", "medium", "high", "critical"}, metrics.ThreatLevel, "Should have valid threat level")

	// Test event statistics
	timeRange := &TimeRange{
		Start: time.Now().Add(-24 * time.Hour),
		End:   time.Now(),
	}

	stats, err := suite.eventLogger.GetEventStatistics(timeRange)
	suite.NoError(err)
	suite.NotNil(stats, "Should return event statistics")
	suite.GreaterOrEqual(stats.TotalEvents, 0, "Should have non-negative event count")
}

// TestPermissionManagementIntegration tests permission management integration
func (suite *SecurityAdministrationTestSuite) TestPermissionManagementIntegration() {
	// Test permission management HTTP interface
	handler := suite.permissionManager.ServePermissionManagementUI()

	// Test template endpoint
	req := httptest.NewRequest("GET", "/api/permissions/templates", nil)
	w := httptest.NewRecorder()
	handler.ServeHTTP(w, req)

	suite.Equal(http.StatusOK, w.Code, "Should return OK for templates")

	var templates []*PermissionTemplate
	err := json.Unmarshal(w.Body.Bytes(), &templates)
	suite.NoError(err, "Should parse template response")
	suite.Greater(len(templates), 0, "Should return templates")

	// Verify template structure
	for _, template := range templates {
		suite.NotEmpty(template.ID, "Template should have ID")
		suite.NotEmpty(template.Name, "Template should have name")
		suite.NotNil(template.Permissions, "Template should have permissions")
	}

	// Test policies endpoint
	req = httptest.NewRequest("GET", "/api/permissions/policies", nil)
	w = httptest.NewRecorder()
	handler.ServeHTTP(w, req)

	suite.Equal(http.StatusOK, w.Code, "Should return OK for policies")

	var policies []*SystemSecurityPolicy
	err = json.Unmarshal(w.Body.Bytes(), &policies)
	suite.NoError(err, "Should parse policies response")
	suite.Greater(len(policies), 0, "Should return policies")

	// Test trust chain endpoint
	req = httptest.NewRequest("GET", "/api/permissions/trust-chain?document_id=test-doc", nil)
	w = httptest.NewRecorder()
	handler.ServeHTTP(w, req)

	suite.Equal(http.StatusOK, w.Code, "Should return OK for trust chain")

	var trustChain []*TrustedSigner
	err = json.Unmarshal(w.Body.Bytes(), &trustChain)
	suite.NoError(err, "Should parse trust chain response")
}

// TestSecurityIntegrationWithWASMContext tests integration with WASM security context
func (suite *SecurityAdministrationTestSuite) TestSecurityIntegrationWithWASMContext() {
	ctx := context.Background()

	// Create security orchestrator for integration testing
	wasmContext := &WASMSecurityContext{
		activeModules:   make(map[string]*WASMModuleContext),
		resourceMonitor: &ResourceMonitor{},
		permissionEngine: &PermissionEngine{
			activePermissions: make(map[string]*core.WASMPermissions),
		},
	}

	errorHandler := &ErrorHandler{
		errorLogger: func(err error, details map[string]interface{}) {
			suite.mockLogger.Error(err.Error(), details)
		},
	}

	orchestrator := NewSecurityOrchestrator(suite.policyManager, wasmContext, errorHandler)

	// Test document processing with security integration
	doc := createTestDocument()
	userContext := &UserContext{
		UserID:    "integration-user",
		SessionID: "integration-session",
		IPAddress: "127.0.0.1",
		Roles:     []string{"user"},
	}

	// Setup mock expectations for successful processing
	suite.mockSM.On("EvaluatePermissions", doc.Manifest.Security.WASMPermissions, suite.getPolicy("basic-policy").SecurityPolicy).Return(true)
	suite.mockLogger.On("Info", "Permission evaluation completed",
		"document_id", generateDocumentID(doc),
		"policy_id", "basic-policy",
		"granted", true,
		"warnings", 0,
	).Return()

	err := orchestrator.ProcessDocument(ctx, doc, "basic-policy", userContext)
	suite.NoError(err, "Should process document successfully with valid permissions")

	// Verify WASM context was set up
	suite.Greater(len(wasmContext.activeModules), 0, "Should have active WASM modules")
	suite.Greater(len(wasmContext.permissionEngine.activePermissions), 0, "Should have active permissions")

	// Test document processing with security violations
	violatingDoc := createTestDocument()
	violatingDoc.Manifest.Security.WASMPermissions.MemoryLimit = 128 * 1024 * 1024 // Excessive memory

	suite.mockSM.On("EvaluatePermissions", violatingDoc.Manifest.Security.WASMPermissions, suite.getPolicy("strict-policy").SecurityPolicy).Return(false)
	suite.mockLogger.On("Info", "Permission evaluation completed",
		"document_id", generateDocumentID(violatingDoc),
		"policy_id", "strict-policy",
		"granted", false,
		"warnings", 0,
	).Return()

	err = orchestrator.ProcessDocument(ctx, violatingDoc, "strict-policy", userContext)
	suite.Error(err, "Should fail to process document with security violations")
	suite.Contains(err.Error(), "quarantined", "Should mention quarantine in error")

	suite.mockSM.AssertExpectations(suite.T())
	suite.mockLogger.AssertExpectations(suite.T())
}

// TestErrorHandlingAndRecovery tests error handling and recovery mechanisms
func (suite *SecurityAdministrationTestSuite) TestErrorHandlingAndRecovery() {
	ctx := context.Background()

	// Test invalid policy creation
	invalidPolicy := &SystemSecurityPolicy{
		// Missing required fields
		SecurityPolicy: nil,
	}

	err := suite.policyManager.CreatePolicy(ctx, invalidPolicy, "test-user")
	suite.Error(err, "Should fail to create invalid policy")
	suite.Contains(err.Error(), "validation failed", "Should mention validation failure")

	// Test permission evaluation with missing policy
	missingPolicyRequest := &PermissionRequest{
		DocumentID:     "test-doc",
		RequestedPerms: &core.WASMPermissions{MemoryLimit: 1024},
		PolicyID:       "non-existent-policy",
		UserContext:    &UserContext{UserID: "test-user"},
	}

	evaluation, err := suite.permissionManager.EvaluatePermissionRequest(ctx, missingPolicyRequest)
	suite.Error(err, "Should fail with missing policy")
	suite.Nil(evaluation, "Should not return evaluation for missing policy")

	// Test resource monitoring with invalid metrics
	invalidMetrics := &ResourceMetrics{
		MemoryUsage: -1, // Invalid negative value
	}

	report, err := suite.policyManager.MonitorResourceUsage(ctx, invalidMetrics)
	suite.NoError(err, "Should handle invalid metrics gracefully")
	suite.NotNil(report, "Should return report even with invalid metrics")

	// Test audit log with corrupted data
	// This would test the resilience of the audit system
	corruptedLogPath := filepath.Join(suite.tempDir, "corrupted-audit.log")
	err = ioutil.WriteFile(corruptedLogPath, []byte("invalid json data\n{malformed"), 0644)
	suite.NoError(err)

	corruptedLogger := NewFileAuditLogger(corruptedLogPath)
	events, err := corruptedLogger.GetAuditTrail(&AuditFilter{})
	suite.NoError(err, "Should handle corrupted audit log gracefully")
	suite.Empty(events, "Should return empty events for corrupted log")
}

// Helper methods

func (suite *SecurityAdministrationTestSuite) getPolicy(policyID string) *SystemSecurityPolicy {
	policy, err := suite.policyManager.GetPolicy(context.Background(), policyID)
	suite.Require().NoError(err)
	return policy
}

// TestSecurityAdministrationSuite runs the complete security administration test suite
func TestSecurityAdministrationSuite(t *testing.T) {
	suite.Run(t, new(SecurityAdministrationTestSuite))
}

// Additional integration tests for specific security scenarios

func TestSecurityPolicyEnforcementScenarios(t *testing.T) {
	tempDir, err := ioutil.TempDir("", "security-scenarios-*")
	require.NoError(t, err)
	defer os.RemoveAll(tempDir)

	// Setup
	eventLogger := NewFileSecurityEventLogger(filepath.Join(tempDir, "security.log"))
	auditLogger := NewFileAuditLogger(filepath.Join(tempDir, "audit.log"))
	config := &PolicyManagerConfig{DefaultPolicyID: "default"}
	pm := NewPolicyManager(config, eventLogger, auditLogger)

	ctx := context.Background()

	// Test scenario: Escalating permission requests
	t.Run("EscalatingPermissionRequests", func(t *testing.T) {
		// Create a policy with moderate limits
		moderatePolicy := createTestPolicy("moderate-policy", "Moderate Policy")
		moderatePolicy.SecurityPolicy.WASMPermissions.MemoryLimit = 16 * 1024 * 1024 // 16MB
		err := pm.CreatePolicy(ctx, moderatePolicy, "admin")
		require.NoError(t, err)

		// Test requests with increasing resource demands
		requests := []*ResourceMetrics{
			{MemoryUsage: 8 * 1024 * 1024},  // 8MB - should pass
			{MemoryUsage: 16 * 1024 * 1024}, // 16MB - at limit
			{MemoryUsage: 32 * 1024 * 1024}, // 32MB - should violate
			{MemoryUsage: 64 * 1024 * 1024}, // 64MB - severe violation
		}

		violationCounts := []int{0, 0, 1, 1} // Expected violation counts

		for i, metrics := range requests {
			report, err := pm.MonitorResourceUsage(ctx, metrics)
			assert.NoError(t, err)
			assert.Len(t, report.Violations, violationCounts[i],
				"Request %d should have %d violations", i+1, violationCounts[i])
		}
	})

	// Test scenario: Policy inheritance chain
	t.Run("PolicyInheritanceChain", func(t *testing.T) {
		// Create inheritance chain: grandparent -> parent -> child
		grandparentPolicy := createTestPolicy("grandparent", "Grandparent Policy")
		grandparentPolicy.SecurityPolicy.WASMPermissions.AllowNetworking = true
		err := pm.CreatePolicy(ctx, grandparentPolicy, "admin")
		require.NoError(t, err)

		parentPolicy := createTestPolicy("parent", "Parent Policy")
		parentPolicy.ParentPolicy = "grandparent"
		parentPolicy.SecurityPolicy.WASMPermissions.AllowNetworking = false // Override
		err = pm.CreatePolicy(ctx, parentPolicy, "admin")
		require.NoError(t, err)

		childPolicy := createTestPolicy("child", "Child Policy")
		childPolicy.ParentPolicy = "parent"
		// Child doesn't specify networking - should inherit from parent (false)
		childPolicy.SecurityPolicy.WASMPermissions = &core.WASMPermissions{
			MemoryLimit:  4 * 1024 * 1024,
			CPUTimeLimit: 2000,
			// AllowNetworking not specified - should inherit
		}
		err = pm.CreatePolicy(ctx, childPolicy, "admin")
		require.NoError(t, err)

		// Verify inheritance works correctly
		resolvedChild, err := pm.GetPolicy(ctx, "child")
		assert.NoError(t, err)
		assert.False(t, resolvedChild.SecurityPolicy.WASMPermissions.AllowNetworking,
			"Child should inherit networking=false from parent, not grandparent")
	})

	// Test scenario: Concurrent policy modifications
	t.Run("ConcurrentPolicyModifications", func(t *testing.T) {
		concurrentPolicy := createTestPolicy("concurrent-policy", "Concurrent Policy")
		err := pm.CreatePolicy(ctx, concurrentPolicy, "admin")
		require.NoError(t, err)

		// Simulate concurrent updates
		done := make(chan bool, 2)

		go func() {
			for i := 0; i < 10; i++ {
				policy, _ := pm.GetPolicy(ctx, "concurrent-policy")
				policy.Description = fmt.Sprintf("Updated by goroutine 1 - %d", i)
				pm.UpdatePolicy(ctx, "concurrent-policy", policy, "user1")
				time.Sleep(1 * time.Millisecond)
			}
			done <- true
		}()

		go func() {
			for i := 0; i < 10; i++ {
				policy, _ := pm.GetPolicy(ctx, "concurrent-policy")
				policy.Description = fmt.Sprintf("Updated by goroutine 2 - %d", i)
				pm.UpdatePolicy(ctx, "concurrent-policy", policy, "user2")
				time.Sleep(1 * time.Millisecond)
			}
			done <- true
		}()

		// Wait for both goroutines to complete
		<-done
		<-done

		// Verify policy still exists and is valid
		finalPolicy, err := pm.GetPolicy(ctx, "concurrent-policy")
		assert.NoError(t, err)
		assert.NotNil(t, finalPolicy)
		assert.Contains(t, finalPolicy.Description, "Updated by goroutine")
	})
}

func TestSecurityEventCorrelation(t *testing.T) {
	tempDir, err := ioutil.TempDir("", "event-correlation-*")
	require.NoError(t, err)
	defer os.RemoveAll(tempDir)

	eventLogger := NewFileSecurityEventLogger(filepath.Join(tempDir, "security.log"))

	// Generate correlated security events
	baseTime := time.Now()
	events := []*SecurityEvent{
		{
			ID:        "event-1",
			Timestamp: baseTime,
			EventType: EventPolicyViolation,
			Severity:  SeverityMedium,
			Source:    "test-source",
			UserID:    "user-123",
			Details:   map[string]interface{}{"violation_type": "memory_exceeded"},
		},
		{
			ID:        "event-2",
			Timestamp: baseTime.Add(1 * time.Second),
			EventType: EventResourceExceeded,
			Severity:  SeverityHigh,
			Source:    "test-source",
			UserID:    "user-123",
			Details:   map[string]interface{}{"resource_type": "memory"},
		},
		{
			ID:        "event-3",
			Timestamp: baseTime.Add(2 * time.Second),
			EventType: EventSuspiciousActivity,
			Severity:  SeverityCritical,
			Source:    "test-source",
			UserID:    "user-123",
			Details:   map[string]interface{}{"activity_type": "repeated_violations"},
		},
	}

	// Log events
	for _, event := range events {
		err := eventLogger.LogSecurityEvent(event)
		require.NoError(t, err)
	}

	// Test event correlation by user
	userEvents, err := eventLogger.GetSecurityEvents(&EventFilter{
		UserID: "user-123",
	})
	assert.NoError(t, err)
	assert.Len(t, userEvents, 3, "Should find all events for user")

	// Test event correlation by severity escalation
	highSeverityEvents, err := eventLogger.GetSecurityEvents(&EventFilter{
		Severities: []SecurityEventSeverity{SeverityHigh, SeverityCritical},
	})
	assert.NoError(t, err)
	assert.Len(t, highSeverityEvents, 2, "Should find high and critical severity events")

	// Test temporal correlation
	timeWindow := &EventFilter{
		StartTime: &baseTime,
		EndTime:   &[]time.Time{baseTime.Add(3 * time.Second)}[0],
	}
	windowEvents, err := eventLogger.GetSecurityEvents(timeWindow)
	assert.NoError(t, err)
	assert.Len(t, windowEvents, 3, "Should find all events in time window")
}
