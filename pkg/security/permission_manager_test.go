// Tests for permission management interface

package security

import (
	"bytes"
	"context"
	"encoding/json"
	"io/ioutil"
	"net/http"
	"net/http/httptest"
	"os"
	"path/filepath"
	"strings"
	"testing"
	"time"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/mock"
	"github.com/stretchr/testify/require"

	"github.com/liv-format/liv/pkg/core"
)

// Mock implementations for testing
type MockSecurityManager struct {
	mock.Mock
}

func (m *MockSecurityManager) ValidateSignature(content []byte, signature string, publicKey []byte) bool {
	args := m.Called(content, signature, publicKey)
	return args.Bool(0)
}

func (m *MockSecurityManager) CreateSignature(content []byte, privateKey []byte) (string, error) {
	args := m.Called(content, privateKey)
	return args.String(0), args.Error(1)
}

func (m *MockSecurityManager) ValidateWASMModule(module []byte, permissions *core.WASMPermissions) error {
	args := m.Called(module, permissions)
	return args.Error(0)
}

func (m *MockSecurityManager) CreateSandbox(policy *core.SecurityPolicy) (core.Sandbox, error) {
	args := m.Called(policy)
	return args.Get(0).(core.Sandbox), args.Error(1)
}

func (m *MockSecurityManager) EvaluatePermissions(requested *core.WASMPermissions, policy *core.SecurityPolicy) bool {
	args := m.Called(requested, policy)
	return args.Bool(0)
}

func (m *MockSecurityManager) GenerateSecurityReport(doc *core.LIVDocument) *core.SecurityReport {
	args := m.Called(doc)
	return args.Get(0).(*core.SecurityReport)
}

type MockCryptoProvider struct {
	mock.Mock
}

func (m *MockCryptoProvider) GenerateKeyPair() (publicKey, privateKey []byte, err error) {
	args := m.Called()
	return args.Get(0).([]byte), args.Get(1).([]byte), args.Error(2)
}

func (m *MockCryptoProvider) Sign(data []byte, privateKey []byte) ([]byte, error) {
	args := m.Called(data, privateKey)
	return args.Get(0).([]byte), args.Error(1)
}

func (m *MockCryptoProvider) Verify(data []byte, signature []byte, publicKey []byte) bool {
	args := m.Called(data, signature, publicKey)
	return args.Bool(0)
}

func (m *MockCryptoProvider) Hash(data []byte) []byte {
	args := m.Called(data)
	return args.Get(0).([]byte)
}

func (m *MockCryptoProvider) GenerateRandomBytes(length int) ([]byte, error) {
	args := m.Called(length)
	return args.Get(0).([]byte), args.Error(1)
}

type MockLogger struct {
	mock.Mock
}

func (m *MockLogger) Debug(msg string, fields ...interface{}) {
	m.Called(msg, fields)
}

func (m *MockLogger) Info(msg string, fields ...interface{}) {
	m.Called(msg, fields)
}

func (m *MockLogger) Warn(msg string, fields ...interface{}) {
	m.Called(msg, fields)
}

func (m *MockLogger) Error(msg string, fields ...interface{}) {
	m.Called(msg, fields)
}

func (m *MockLogger) Fatal(msg string, fields ...interface{}) {
	m.Called(msg, fields)
}

type MockMetricsCollector struct {
	mock.Mock
}

func (m *MockMetricsCollector) RecordDocumentLoad(size int64, duration int64) {
	m.Called(size, duration)
}

func (m *MockMetricsCollector) RecordWASMExecution(module string, duration int64, memoryUsed uint64) {
	m.Called(module, duration, memoryUsed)
}

func (m *MockMetricsCollector) RecordSecurityEvent(eventType string, details map[string]interface{}) {
	m.Called(eventType, details)
}

func (m *MockMetricsCollector) GetMetrics() map[string]interface{} {
	args := m.Called()
	return args.Get(0).(map[string]interface{})
}

func TestPermissionManager_EvaluatePermissionRequest(t *testing.T) {
	tempDir, err := ioutil.TempDir("", "permission-test-*")
	require.NoError(t, err)
	defer os.RemoveAll(tempDir)

	// Create mocks
	mockSM := &MockSecurityManager{}
	mockCP := &MockCryptoProvider{}
	mockLogger := &MockLogger{}

	// Create policy manager
	eventLogger := NewFileSecurityEventLogger(filepath.Join(tempDir, "security.log"))
	auditLogger := NewFileAuditLogger(filepath.Join(tempDir, "audit.log"))
	config := &PolicyManagerConfig{
		DefaultPolicyID: "default",
	}
	policyManager := NewPolicyManager(config, eventLogger, auditLogger)

	// Create permission manager
	permManager := NewPermissionManager(policyManager, mockSM, mockCP, mockLogger)

	// Create test policy
	policy := createTestPolicy("test-policy", "Test Policy")
	err = policyManager.CreatePolicy(context.Background(), policy, "admin")
	require.NoError(t, err)

	// Create permission request
	request := &PermissionRequest{
		DocumentID: "test-doc-1",
		ModuleName: "test-module",
		RequestedPerms: &core.WASMPermissions{
			MemoryLimit:     8 * 1024 * 1024, // 8MB
			AllowedImports:  []string{"console"},
			CPUTimeLimit:    3000,
			AllowNetworking: false,
			AllowFileSystem: false,
		},
		PolicyID: "test-policy",
		UserContext: &UserContext{
			UserID:    "test-user",
			SessionID: "test-session",
			IPAddress: "127.0.0.1",
			Roles:     []string{"user"},
		},
		Justification: "Testing permission evaluation",
		RequestedAt:   time.Now(),
	}

	// Setup mock expectations
	mockSM.On("EvaluatePermissions", request.RequestedPerms, policy.SecurityPolicy).Return(true)
	mockLogger.On("Info", mock.AnythingOfType("string"), mock.Anything).Return()

	// Test permission evaluation
	evaluation, err := permManager.EvaluatePermissionRequest(context.Background(), request)
	assert.NoError(t, err, "Should evaluate permission request successfully")
	assert.NotNil(t, evaluation, "Should return evaluation result")
	assert.True(t, evaluation.Granted, "Should grant permissions")
	assert.Empty(t, evaluation.InheritedFrom, "Should not inherit permissions")

	// Verify mock calls
	mockSM.AssertExpectations(t)
	mockLogger.AssertExpectations(t)
}

func TestPermissionManager_PermissionInheritance(t *testing.T) {
	tempDir, err := ioutil.TempDir("", "inheritance-test-*")
	require.NoError(t, err)
	defer os.RemoveAll(tempDir)

	// Create mocks
	mockSM := &MockSecurityManager{}
	mockCP := &MockCryptoProvider{}
	mockLogger := &MockLogger{}

	// Create policy manager
	eventLogger := NewFileSecurityEventLogger(filepath.Join(tempDir, "security.log"))
	auditLogger := NewFileAuditLogger(filepath.Join(tempDir, "audit.log"))
	config := &PolicyManagerConfig{
		DefaultPolicyID:         "default",
		EnablePolicyInheritance: true,
		MaxPolicyDepth:          5,
	}
	policyManager := NewPolicyManager(config, eventLogger, auditLogger)

	// Create permission manager
	permManager := NewPermissionManager(policyManager, mockSM, mockCP, mockLogger)

	// Create parent policy
	parentPolicy := createTestPolicy("parent-policy", "Parent Policy")
	err = policyManager.CreatePolicy(context.Background(), parentPolicy, "admin")
	require.NoError(t, err)

	// Create child policy with inheritance
	childPolicy := createTestPolicy("child-policy", "Child Policy")
	childPolicy.ParentPolicy = "parent-policy"
	err = policyManager.CreatePolicy(context.Background(), childPolicy, "admin")
	require.NoError(t, err)

	// Create permission request for child policy
	request := &PermissionRequest{
		DocumentID: "test-doc-1",
		RequestedPerms: &core.WASMPermissions{
			MemoryLimit:     8 * 1024 * 1024,
			AllowedImports:  []string{"console"},
			CPUTimeLimit:    3000,
			AllowNetworking: false,
			AllowFileSystem: false,
		},
		PolicyID:      "child-policy",
		UserContext:   &UserContext{UserID: "test-user"},
		Justification: "Testing inheritance",
		RequestedAt:   time.Now(),
	}

	// Setup mock expectations - child policy denies, parent policy grants
	mockSM.On("EvaluatePermissions", request.RequestedPerms, childPolicy.SecurityPolicy).Return(false)
	mockSM.On("EvaluatePermissions", request.RequestedPerms, parentPolicy.SecurityPolicy).Return(true)
	mockLogger.On("Info", mock.AnythingOfType("string"), mock.Anything).Return()

	// Test permission evaluation with inheritance
	evaluation, err := permManager.EvaluatePermissionRequest(context.Background(), request)
	assert.NoError(t, err, "Should evaluate permission request successfully")
	assert.NotNil(t, evaluation, "Should return evaluation result")
	assert.True(t, evaluation.Granted, "Should grant permissions through inheritance")
	assert.Equal(t, "parent-policy", evaluation.InheritedFrom, "Should inherit from parent policy")
	assert.Greater(t, len(evaluation.Warnings), 0, "Should have inheritance warning")

	// Verify mock calls
	mockSM.AssertExpectations(t)
	mockLogger.AssertExpectations(t)
}

func TestPermissionManager_GetPermissionTemplates(t *testing.T) {
	// Create mocks
	mockSM := &MockSecurityManager{}
	mockCP := &MockCryptoProvider{}
	mockLogger := &MockLogger{}

	// Create permission manager
	permManager := NewPermissionManager(nil, mockSM, mockCP, mockLogger)

	// Get permission templates
	templates := permManager.GetPermissionTemplates()
	assert.NotEmpty(t, templates, "Should return permission templates")

	// Verify template structure
	for _, template := range templates {
		assert.NotEmpty(t, template.ID, "Template should have ID")
		assert.NotEmpty(t, template.Name, "Template should have name")
		assert.NotEmpty(t, template.Description, "Template should have description")
		assert.NotNil(t, template.Permissions, "Template should have permissions")
		assert.NotEmpty(t, template.UseCase, "Template should have use case")
	}

	// Check for specific templates
	templateIDs := make(map[string]bool)
	for _, template := range templates {
		templateIDs[template.ID] = true
	}

	assert.True(t, templateIDs["basic-document"], "Should have basic-document template")
	assert.True(t, templateIDs["interactive-content"], "Should have interactive-content template")
	assert.True(t, templateIDs["data-visualization"], "Should have data-visualization template")
	assert.True(t, templateIDs["network-enabled"], "Should have network-enabled template")
}

func TestPermissionManager_CreatePermissionTemplate(t *testing.T) {
	// Create mocks
	mockSM := &MockSecurityManager{}
	mockCP := &MockCryptoProvider{}
	mockLogger := &MockLogger{}

	// Create permission manager
	permManager := NewPermissionManager(nil, mockSM, mockCP, mockLogger)

	// Test valid template creation
	template := &PermissionTemplate{
		ID:          "custom-template",
		Name:        "Custom Template",
		Description: "A custom permission template",
		Category:    "custom",
		Permissions: &core.WASMPermissions{
			MemoryLimit:     4 * 1024 * 1024,
			AllowedImports:  []string{"console"},
			CPUTimeLimit:    2000,
			AllowNetworking: false,
			AllowFileSystem: false,
		},
		UseCase: "Custom use case",
	}

	mockLogger.On("Info", mock.AnythingOfType("string"), mock.Anything).Return()

	err := permManager.CreatePermissionTemplate(template)
	assert.NoError(t, err, "Should create template successfully")

	// Test invalid template creation
	invalidTemplate := &PermissionTemplate{
		Name: "Invalid Template",
		// Missing ID and permissions
	}

	err = permManager.CreatePermissionTemplate(invalidTemplate)
	assert.Error(t, err, "Should fail to create invalid template")
	assert.Contains(t, err.Error(), "ID", "Error should mention missing ID")

	mockLogger.AssertExpectations(t)
}

func TestPermissionManager_HTTPHandlers(t *testing.T) {
	tempDir, err := ioutil.TempDir("", "http-test-*")
	require.NoError(t, err)
	defer os.RemoveAll(tempDir)

	// Create mocks
	mockSM := &MockSecurityManager{}
	mockCP := &MockCryptoProvider{}
	mockLogger := &MockLogger{}

	// Create policy manager
	eventLogger := NewFileSecurityEventLogger(filepath.Join(tempDir, "security.log"))
	auditLogger := NewFileAuditLogger(filepath.Join(tempDir, "audit.log"))
	config := &PolicyManagerConfig{
		DefaultPolicyID: "default",
	}
	policyManager := NewPolicyManager(config, eventLogger, auditLogger)

	// Create permission manager
	permManager := NewPermissionManager(policyManager, mockSM, mockCP, mockLogger)

	// Create test policy
	policy := createTestPolicy("test-policy", "Test Policy")
	err = policyManager.CreatePolicy(context.Background(), policy, "admin")
	require.NoError(t, err)

	// Create HTTP handler
	handler := permManager.ServePermissionManagementUI()

	// Test dashboard endpoint
	req := httptest.NewRequest("GET", "/", nil)
	w := httptest.NewRecorder()
	handler.ServeHTTP(w, req)

	assert.Equal(t, http.StatusOK, w.Code, "Should return OK for dashboard")
	assert.Contains(t, w.Header().Get("Content-Type"), "text/html", "Should return HTML content")
	assert.Contains(t, w.Body.String(), "Permission Management", "Should contain dashboard content")

	// Test permission templates endpoint
	req = httptest.NewRequest("GET", "/api/permissions/templates", nil)
	w = httptest.NewRecorder()
	handler.ServeHTTP(w, req)

	assert.Equal(t, http.StatusOK, w.Code, "Should return OK for templates")
	assert.Contains(t, w.Header().Get("Content-Type"), "application/json", "Should return JSON content")

	var templates []*PermissionTemplate
	err = json.Unmarshal(w.Body.Bytes(), &templates)
	assert.NoError(t, err, "Should parse JSON response")
	assert.NotEmpty(t, templates, "Should return templates")

	// Test policies endpoint
	req = httptest.NewRequest("GET", "/api/permissions/policies", nil)
	w = httptest.NewRecorder()
	handler.ServeHTTP(w, req)

	assert.Equal(t, http.StatusOK, w.Code, "Should return OK for policies")
	assert.Contains(t, w.Header().Get("Content-Type"), "application/json", "Should return JSON content")

	var policies []*SystemSecurityPolicy
	err = json.Unmarshal(w.Body.Bytes(), &policies)
	assert.NoError(t, err, "Should parse JSON response")
	assert.NotEmpty(t, policies, "Should return policies")

	// Test permission evaluation endpoint
	request := &PermissionRequest{
		DocumentID: "test-doc-1",
		RequestedPerms: &core.WASMPermissions{
			MemoryLimit:     8 * 1024 * 1024,
			AllowedImports:  []string{"console"},
			CPUTimeLimit:    3000,
			AllowNetworking: false,
			AllowFileSystem: false,
		},
		PolicyID:      "test-policy",
		UserContext:   &UserContext{UserID: "test-user"},
		Justification: "Testing HTTP endpoint",
		RequestedAt:   time.Now(),
	}

	requestBody, err := json.Marshal(request)
	require.NoError(t, err)

	mockSM.On("EvaluatePermissions", request.RequestedPerms, policy.SecurityPolicy).Return(true)
	mockLogger.On("Info", mock.AnythingOfType("string"), mock.Anything).Return()

	req = httptest.NewRequest("POST", "/api/permissions/evaluate", bytes.NewReader(requestBody))
	req.Header.Set("Content-Type", "application/json")
	w = httptest.NewRecorder()
	handler.ServeHTTP(w, req)

	assert.Equal(t, http.StatusOK, w.Code, "Should return OK for evaluation")
	assert.Contains(t, w.Header().Get("Content-Type"), "application/json", "Should return JSON content")

	var evaluation *PermissionEvaluation
	err = json.Unmarshal(w.Body.Bytes(), &evaluation)
	assert.NoError(t, err, "Should parse JSON response")
	assert.NotNil(t, evaluation, "Should return evaluation")
	assert.True(t, evaluation.Granted, "Should grant permissions")

	// Verify mock calls
	mockSM.AssertExpectations(t)
	mockLogger.AssertExpectations(t)
}

func TestPermissionManager_SecurityWarnings(t *testing.T) {
	// Create mocks
	mockSM := &MockSecurityManager{}
	mockCP := &MockCryptoProvider{}
	mockLogger := &MockLogger{}

	// Create permission manager
	permManager := NewPermissionManager(nil, mockSM, mockCP, mockLogger)

	// Create test policy
	policy := createTestPolicy("test-policy", "Test Policy")

	// Test high memory usage warning
	highMemoryPerms := &core.WASMPermissions{
		MemoryLimit:     64 * 1024 * 1024, // 64MB - should trigger warning
		AllowedImports:  []string{"console"},
		CPUTimeLimit:    3000,
		AllowNetworking: false,
		AllowFileSystem: false,
	}

	warnings := permManager.generateSecurityWarnings(highMemoryPerms, policy)
	assert.NotEmpty(t, warnings, "Should generate warnings for high memory usage")

	hasMemoryWarning := false
	for _, warning := range warnings {
		if warning.Type == "high_memory_usage" {
			hasMemoryWarning = true
			break
		}
	}
	assert.True(t, hasMemoryWarning, "Should have high memory usage warning")

	// Test network access warning
	networkPerms := &core.WASMPermissions{
		MemoryLimit:     8 * 1024 * 1024,
		AllowedImports:  []string{"console"},
		CPUTimeLimit:    3000,
		AllowNetworking: true, // Should trigger warning
		AllowFileSystem: false,
	}

	warnings = permManager.generateSecurityWarnings(networkPerms, policy)
	hasNetworkWarning := false
	for _, warning := range warnings {
		if warning.Type == "network_access_requested" {
			hasNetworkWarning = true
			break
		}
	}
	assert.True(t, hasNetworkWarning, "Should have network access warning")

	// Test file system access warning
	filesystemPerms := &core.WASMPermissions{
		MemoryLimit:     8 * 1024 * 1024,
		AllowedImports:  []string{"console"},
		CPUTimeLimit:    3000,
		AllowNetworking: false,
		AllowFileSystem: true, // Should trigger warning
	}

	warnings = permManager.generateSecurityWarnings(filesystemPerms, policy)
	hasFilesystemWarning := false
	for _, warning := range warnings {
		if warning.Type == "filesystem_access_requested" {
			hasFilesystemWarning = true
			break
		}
	}
	assert.True(t, hasFilesystemWarning, "Should have filesystem access warning")
}

func TestPermissionManager_CalculateRestrictions(t *testing.T) {
	// Create mocks
	mockSM := &MockSecurityManager{}
	mockCP := &MockCryptoProvider{}
	mockLogger := &MockLogger{}

	// Create permission manager
	permManager := NewPermissionManager(nil, mockSM, mockCP, mockLogger)

	// Create test policy with restrictive settings
	policy := createTestPolicy("restrictive-policy", "Restrictive Policy")
	policy.SecurityPolicy.WASMPermissions.MemoryLimit = 4 * 1024 * 1024 // 4MB limit
	policy.SecurityPolicy.WASMPermissions.CPUTimeLimit = 2000           // 2 second limit
	policy.SecurityPolicy.WASMPermissions.AllowNetworking = false
	policy.SecurityPolicy.WASMPermissions.AllowFileSystem = false
	policy.SecurityPolicy.WASMPermissions.AllowedImports = []string{"console"}

	// Test permissions that exceed policy limits
	requestedPerms := &core.WASMPermissions{
		MemoryLimit:     8 * 1024 * 1024,                     // Exceeds policy limit
		CPUTimeLimit:    5000,                                // Exceeds policy limit
		AllowNetworking: true,                                // Not allowed by policy
		AllowFileSystem: true,                                // Not allowed by policy
		AllowedImports:  []string{"console", "dom", "fetch"}, // Some not allowed
	}

	restrictions := permManager.calculateRestrictions(requestedPerms, policy)
	assert.NotEmpty(t, restrictions, "Should generate restrictions")

	// Check for specific restrictions
	restrictionText := strings.Join(restrictions, " ")
	assert.Contains(t, restrictionText, "Memory limited", "Should have memory restriction")
	assert.Contains(t, restrictionText, "CPU time limited", "Should have CPU time restriction")
	assert.Contains(t, restrictionText, "Network access denied", "Should have network restriction")
	assert.Contains(t, restrictionText, "File system access denied", "Should have filesystem restriction")
	assert.Contains(t, restrictionText, "Import 'dom' not allowed", "Should have import restriction for dom")
	assert.Contains(t, restrictionText, "Import 'fetch' not allowed", "Should have import restriction for fetch")
}
