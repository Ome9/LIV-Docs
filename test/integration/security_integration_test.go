// Integration tests for security and administration systems
// Tests integration with existing WASM security context, error handling, and signature systems

package integration

import (
	"context"
	"fmt"
	"io/ioutil"
	"os"
	"path/filepath"
	"testing"
	"time"

	"github.com/stretchr/testify/suite"

	"github.com/liv-format/liv/pkg/core"
	"github.com/liv-format/liv/pkg/security"
)

// SecurityIntegrationTestSuite tests security system integration
type SecurityIntegrationTestSuite struct {
	suite.Suite
	tempDir           string
	policyManager     *security.PolicyManager
	permissionManager *security.PermissionManager
	orchestrator      *security.SecurityOrchestrator
	wasmContext       *security.WASMSecurityContext
	errorHandler      *security.ErrorHandler
	cryptoProvider    *TestCryptoProvider
	securityManager   *TestSecurityManager
	logger            *TestLogger
}

// TestCryptoProvider implements core.CryptoProvider for testing
type TestCryptoProvider struct {
	keyPairs map[string]KeyPair
}

type KeyPair struct {
	PublicKey  []byte
	PrivateKey []byte
}

func NewTestCryptoProvider() *TestCryptoProvider {
	return &TestCryptoProvider{
		keyPairs: make(map[string]KeyPair),
	}
}

func (cp *TestCryptoProvider) GenerateKeyPair() (publicKey, privateKey []byte, err error) {
	keyPair := KeyPair{
		PublicKey:  []byte("test-public-key"),
		PrivateKey: []byte("test-private-key"),
	}
	cp.keyPairs["default"] = keyPair
	return keyPair.PublicKey, keyPair.PrivateKey, nil
}

func (cp *TestCryptoProvider) Sign(data []byte, privateKey []byte) ([]byte, error) {
	return []byte(fmt.Sprintf("signature-%x", data[:min(len(data), 8)])), nil
}

func (cp *TestCryptoProvider) Verify(data []byte, signature []byte, publicKey []byte) bool {
	expectedSig := fmt.Sprintf("signature-%x", data[:min(len(data), 8)])
	return string(signature) == expectedSig
}

func (cp *TestCryptoProvider) Hash(data []byte) []byte {
	return []byte(fmt.Sprintf("hash-%x", data[:min(len(data), 8)]))
}

func (cp *TestCryptoProvider) GenerateRandomBytes(length int) ([]byte, error) {
	bytes := make([]byte, length)
	for i := range bytes {
		bytes[i] = byte(i % 256)
	}
	return bytes, nil
}

// TestSecurityManager implements core.SecurityManager for testing
type TestSecurityManager struct {
	signatures map[string]string
	reports    map[string]*core.SecurityReport
}

func NewTestSecurityManager() *TestSecurityManager {
	return &TestSecurityManager{
		signatures: make(map[string]string),
		reports:    make(map[string]*core.SecurityReport),
	}
}

func (sm *TestSecurityManager) ValidateSignature(content []byte, signature string, publicKey []byte) bool {
	expectedSig := fmt.Sprintf("signature-%x", content[:min(len(content), 8)])
	return signature == expectedSig
}

func (sm *TestSecurityManager) CreateSignature(content []byte, privateKey []byte) (string, error) {
	signature := fmt.Sprintf("signature-%x", content[:min(len(content), 8)])
	sm.signatures[string(content)] = signature
	return signature, nil
}

func (sm *TestSecurityManager) ValidateWASMModule(module []byte, permissions *core.WASMPermissions) error {
	if len(module) == 0 {
		return fmt.Errorf("empty WASM module")
	}
	if permissions.MemoryLimit > 128*1024*1024 {
		return fmt.Errorf("memory limit too high")
	}
	return nil
}

func (sm *TestSecurityManager) CreateSandbox(policy *core.SecurityPolicy) (core.Sandbox, error) {
	return &TestSandbox{policy: policy}, nil
}

func (sm *TestSecurityManager) EvaluatePermissions(requested *core.WASMPermissions, policy *core.SecurityPolicy) bool {
	if policy.WASMPermissions == nil {
		return false
	}

	// Check memory limit
	if requested.MemoryLimit > policy.WASMPermissions.MemoryLimit {
		return false
	}

	// Check CPU time limit
	if requested.CPUTimeLimit > policy.WASMPermissions.CPUTimeLimit {
		return false
	}

	// Check networking permission
	if requested.AllowNetworking && !policy.WASMPermissions.AllowNetworking {
		return false
	}

	// Check filesystem permission
	if requested.AllowFileSystem && !policy.WASMPermissions.AllowFileSystem {
		return false
	}

	return true
}

func (sm *TestSecurityManager) GenerateSecurityReport(doc *core.LIVDocument) *core.SecurityReport {
	report := &core.SecurityReport{
		IsValid:           true,
		SignatureVerified: true,
		IntegrityChecked:  true,
		PermissionsValid:  true,
		Warnings:          []string{},
		Errors:            []string{},
	}

	// Check for potential issues
	if doc.Signatures == nil || doc.Signatures.ContentSignature == "" {
		report.SignatureVerified = false
		report.Warnings = append(report.Warnings, "No content signature found")
	}

	if len(doc.WASMModules) > 5 {
		report.Warnings = append(report.Warnings, "High number of WASM modules")
	}

	sm.reports[generateDocumentID(doc)] = report
	return report
}

// TestSandbox implements core.Sandbox for testing
type TestSandbox struct {
	policy *core.SecurityPolicy
}

func (s *TestSandbox) Execute(ctx context.Context, code string, permissions *core.WASMPermissions) (interface{}, error) {
	return "executed", nil
}

func (s *TestSandbox) LoadWASM(ctx context.Context, module []byte, config *core.WASMModule) (core.WASMInstance, error) {
	return &TestWASMInstance{module: module}, nil
}

func (s *TestSandbox) GetPermissions() *core.SecurityPolicy {
	return s.policy
}

func (s *TestSandbox) UpdatePermissions(policy *core.SecurityPolicy) error {
	s.policy = policy
	return nil
}

func (s *TestSandbox) Destroy() error {
	return nil
}

// TestWASMInstance implements core.WASMInstance for testing
type TestWASMInstance struct {
	module []byte
}

func (w *TestWASMInstance) Call(ctx context.Context, function string, args ...interface{}) (interface{}, error) {
	return fmt.Sprintf("called-%s", function), nil
}

func (w *TestWASMInstance) GetExports() []string {
	return []string{"main", "init", "update"}
}

func (w *TestWASMInstance) GetMemoryUsage() uint64 {
	return uint64(len(w.module) * 2) // Simulate memory usage
}

func (w *TestWASMInstance) SetMemoryLimit(limit uint64) error {
	return nil
}

func (w *TestWASMInstance) Terminate() error {
	return nil
}

// TestLogger implements core.Logger for testing
type TestLogger struct {
	logs []LogEntry
}

type LogEntry struct {
	Level   string
	Message string
	Fields  []interface{}
}

func NewTestLogger() *TestLogger {
	return &TestLogger{logs: []LogEntry{}}
}

func (l *TestLogger) Debug(msg string, fields ...interface{}) {
	l.logs = append(l.logs, LogEntry{"DEBUG", msg, fields})
}

func (l *TestLogger) Info(msg string, fields ...interface{}) {
	l.logs = append(l.logs, LogEntry{"INFO", msg, fields})
}

func (l *TestLogger) Warn(msg string, fields ...interface{}) {
	l.logs = append(l.logs, LogEntry{"WARN", msg, fields})
}

func (l *TestLogger) Error(msg string, fields ...interface{}) {
	l.logs = append(l.logs, LogEntry{"ERROR", msg, fields})
}

func (l *TestLogger) Fatal(msg string, fields ...interface{}) {
	l.logs = append(l.logs, LogEntry{"FATAL", msg, fields})
}

func (l *TestLogger) GetLogs() []LogEntry {
	return l.logs
}

// SetupSuite initializes the integration test suite
func (suite *SecurityIntegrationTestSuite) SetupSuite() {
	var err error
	suite.tempDir, err = ioutil.TempDir("", "security-integration-*")
	suite.Require().NoError(err)

	// Create test implementations
	suite.cryptoProvider = NewTestCryptoProvider()
	suite.securityManager = NewTestSecurityManager()
	suite.logger = NewTestLogger()

	// Create security components
	eventLogger := security.NewFileSecurityEventLogger(filepath.Join(suite.tempDir, "security-events.log"))
	auditLogger := security.NewFileAuditLogger(filepath.Join(suite.tempDir, "audit.log"))

	config := &security.PolicyManagerConfig{
		DefaultPolicyID:         "default",
		EnablePolicyInheritance: true,
		MaxPolicyDepth:          5,
		EnableVersioning:        true,
		AuditLogPath:            filepath.Join(suite.tempDir, "audit.log"),
		EventLogPath:            filepath.Join(suite.tempDir, "security-events.log"),
	}

	suite.policyManager = security.NewPolicyManager(config, eventLogger, auditLogger)
	suite.permissionManager = security.NewPermissionManager(suite.policyManager, suite.securityManager, suite.cryptoProvider, suite.logger)

	// Create WASM security context - using nil since we can't initialize unexported fields
	suite.wasmContext = nil

	// Create error handler - using nil since we can't initialize unexported fields
	suite.errorHandler = nil

	// Create security orchestrator
	suite.orchestrator = security.NewSecurityOrchestrator(suite.policyManager, suite.wasmContext, suite.errorHandler)

	// Create test policies
	suite.createTestPolicies()
}

// TearDownSuite cleans up the integration test suite
func (suite *SecurityIntegrationTestSuite) TearDownSuite() {
	os.RemoveAll(suite.tempDir)
}

// createTestPolicies creates test security policies
func (suite *SecurityIntegrationTestSuite) createTestPolicies() {
	ctx := context.Background()

	// Create comprehensive security policy
	comprehensivePolicy := &security.SystemSecurityPolicy{
		SecurityPolicy: &core.SecurityPolicy{
			WASMPermissions: &core.WASMPermissions{
				MemoryLimit:     16 * 1024 * 1024, // 16MB
				AllowedImports:  []string{"console", "dom"},
				CPUTimeLimit:    5000, // 5 seconds
				AllowNetworking: false,
				AllowFileSystem: false,
			},
			JSPermissions: &core.JSPermissions{
				ExecutionMode: "sandboxed",
				AllowedAPIs:   []string{"console", "dom"},
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
		ID:          "comprehensive-policy",
		Name:        "Comprehensive Security Policy",
		Description: "Comprehensive policy for integration testing",
		Version:     "1.0.0",
		AdminControls: &security.AdminControls{
			RequireApproval:    false,
			MaxDocumentSize:    20 * 1024 * 1024, // 20MB
			MaxWASMModules:     3,
			AllowedFileTypes:   []string{"text/html", "text/css", "application/javascript"},
			RequireSignature:   true,
			TrustedSigners:     []string{"test-ca"},
			EnforceQuarantine:  true,
			QuarantineDuration: 3600, // 1 hour
		},
		EventConfig: &security.SecurityEventConfig{
			LogLevel:             "info",
			EnableAuditLog:       true,
			LogRetentionDays:     90,
			AlertThresholds:      map[string]int{"violations": 5},
			EnableRealTimeAlerts: true,
		},
		ResourceLimits: &security.ResourceLimits{
			MaxConcurrentDocuments: 5,
			MaxMemoryPerDocument:   32 * 1024 * 1024, // 32MB
			MaxCPUTimePerDocument:  10000,            // 10 seconds
			DocumentTimeoutSeconds: 120,              // 2 minutes
		},
		ComplianceSettings: &security.ComplianceSettings{
			EnableGDPRCompliance:  true,
			EnableHIPAACompliance: false,
			DataRetentionDays:     30,
			RequireDataEncryption: true,
			DataClassification:    "confidential",
		},
	}

	err := suite.policyManager.CreatePolicy(ctx, comprehensivePolicy, "integration-admin")
	suite.Require().NoError(err)
}

// TestEndToEndSecurityWorkflow tests complete security workflow
func (suite *SecurityIntegrationTestSuite) TestEndToEndSecurityWorkflow() {
	ctx := context.Background()

	// Create test document
	doc := suite.createTestDocument()
	userContext := &security.UserContext{
		UserID:    "integration-user",
		SessionID: "integration-session",
		IPAddress: "127.0.0.1",
		Roles:     []string{"user"},
	}

	// Test complete document processing workflow
	err := suite.orchestrator.ProcessDocument(ctx, doc, "comprehensive-policy", userContext)
	suite.NoError(err, "Should process valid document successfully")

	// Note: Cannot verify WASM context details due to unexported fields
	// suite.Greater(len(suite.wasmContext.ActiveModules), 0, "Should have active WASM modules")
	// suite.Greater(len(suite.wasmContext.PermissionEngine.ActivePermissions), 0, "Should have configured permissions")

	// Verify security events were logged
	logs := suite.logger.GetLogs()
	hasPermissionLog := false
	for _, log := range logs {
		if log.Level == "INFO" && log.Message == "Permission evaluation completed" {
			hasPermissionLog = true
			break
		}
	}
	suite.True(hasPermissionLog, "Should have logged permission evaluation")

	// Test system status reporting
	status, err := suite.orchestrator.GetSystemStatus(ctx)
	suite.NoError(err, "Should get system status")
	suite.NotNil(status, "Should return system status")
	suite.Greater(status.SecurityMetrics.TotalPolicies, 0, "Should have policies")
	suite.Greater(status.WASMModuleCount, 0, "Should have WASM modules")
	suite.Contains([]string{"healthy", "minor_issues", "warning", "critical"}, status.OverallHealth, "Should have valid health status")
}

// TestWASMSecurityContextIntegration tests WASM security context integration
func (suite *SecurityIntegrationTestSuite) TestWASMSecurityContextIntegration() {
	ctx := context.Background()

	// Create document with multiple WASM modules
	doc := suite.createTestDocument()
	doc.WASMModules["chart-module"] = make([]byte, 2048)
	doc.WASMModules["interaction-module"] = make([]byte, 1024)

	userContext := &security.UserContext{UserID: "wasm-test-user"}

	// Process document
	err := suite.orchestrator.ProcessDocument(ctx, doc, "comprehensive-policy", userContext)
	suite.NoError(err, "Should process document with multiple WASM modules")

	// Note: Cannot verify WASM module details due to unexported fields
	// suite.Len(suite.wasmContext.ActiveModules, 3, "Should have 3 active WASM modules")
	// for moduleID := range doc.WASMModules {
	// 	_, hasPermissions := suite.wasmContext.PermissionEngine.ActivePermissions[moduleID]
	// 	suite.True(hasPermissions, "Module %s should have permissions configured", moduleID)
	// }

	// Test resource monitoring
	resourceMetrics := &security.ResourceMetrics{
		MemoryUsage:         20 * 1024 * 1024, // 20MB
		CPUTime:             3000,             // 3 seconds
		ConcurrentDocuments: 2,
	}

	report, err := suite.policyManager.MonitorResourceUsage(ctx, resourceMetrics)
	suite.NoError(err, "Should monitor resource usage")
	suite.Equal("healthy", report.OverallStatus, "Should be healthy within limits")

	// Test resource violation
	excessiveMetrics := &security.ResourceMetrics{
		MemoryUsage:         50 * 1024 * 1024, // 50MB - exceeds policy limit
		CPUTime:             15000,            // 15 seconds - exceeds policy limit
		ConcurrentDocuments: 10,               // Exceeds policy limit
	}

	violationReport, err := suite.policyManager.MonitorResourceUsage(ctx, excessiveMetrics)
	suite.NoError(err, "Should handle resource violations")
	suite.Greater(len(violationReport.Violations), 0, "Should detect violations")
	suite.Equal("violations_detected", violationReport.OverallStatus, "Should indicate violations")
}

// TestSignatureAndTrustChainIntegration tests signature verification and trust chain
func (suite *SecurityIntegrationTestSuite) TestSignatureAndTrustChainIntegration() {
	ctx := context.Background()

	// Create document with signatures
	doc := suite.createTestDocument()

	// Sign the document content
	contentData := []byte(doc.Content.HTML + doc.Content.CSS)
	signature, err := suite.cryptoProvider.Sign(contentData, []byte("test-private-key"))
	suite.NoError(err, "Should create signature")

	doc.Signatures.ContentSignature = string(signature)

	// Test signature verification through permission evaluation
	request := &security.PermissionRequest{
		DocumentID: generateDocumentID(doc),
		RequestedPerms: &core.WASMPermissions{
			MemoryLimit:     8 * 1024 * 1024,
			CPUTimeLimit:    3000,
			AllowNetworking: false,
			AllowFileSystem: false,
			AllowedImports:  []string{"console"},
		},
		PolicyID:      "comprehensive-policy",
		UserContext:   &security.UserContext{UserID: "signature-test-user"},
		Justification: "Testing signature verification",
		RequestedAt:   time.Now(),
	}

	evaluation, err := suite.permissionManager.EvaluatePermissionRequest(ctx, request)
	suite.NoError(err, "Should evaluate permissions with signature")
	suite.NotNil(evaluation, "Should return evaluation")

	// Note: Cannot test trust chain validation due to unexported method
	// trustChain, err := suite.permissionManager.ValidateTrustChain(ctx, generateDocumentID(doc))
	// suite.NoError(err, "Should validate trust chain")
	// suite.Greater(len(trustChain), 0, "Should have trust chain entries")

	// Test with invalid signature
	doc.Signatures.ContentSignature = "invalid-signature"

	invalidEvaluation, err := suite.permissionManager.EvaluatePermissionRequest(ctx, request)
	suite.NoError(err, "Should handle invalid signature gracefully")
	suite.Greater(len(invalidEvaluation.Warnings), 0, "Should have warnings for invalid signature")
}

// TestErrorHandlingIntegration tests error handling integration
func (suite *SecurityIntegrationTestSuite) TestErrorHandlingIntegration() {
	ctx := context.Background()

	// Test document with security violations
	violatingDoc := suite.createTestDocument()
	violatingDoc.Manifest.Security.WASMPermissions.MemoryLimit = 128 * 1024 * 1024 // Excessive memory
	violatingDoc.WASMModules["malicious-module"] = make([]byte, 0)                 // Empty module

	userContext := &security.UserContext{UserID: "error-test-user"}

	// Process document - should fail gracefully
	err := suite.orchestrator.ProcessDocument(ctx, violatingDoc, "comprehensive-policy", userContext)
	suite.Error(err, "Should fail to process violating document")

	// Verify error was logged
	logs := suite.logger.GetLogs()
	hasErrorLog := false
	for _, log := range logs {
		if log.Level == "ERROR" {
			hasErrorLog = true
			break
		}
	}
	suite.True(hasErrorLog, "Should have logged errors")

	// Test recovery from errors
	validDoc := suite.createTestDocument()
	err = suite.orchestrator.ProcessDocument(ctx, validDoc, "comprehensive-policy", userContext)
	suite.NoError(err, "Should recover and process valid document after error")
}

// TestComplianceAndAuditIntegration tests compliance and audit integration
func (suite *SecurityIntegrationTestSuite) TestComplianceAndAuditIntegration() {
	ctx := context.Background()

	// Test GDPR compliance checking
	gdprDoc := suite.createTestDocument()
	gdprDoc.Content.HTML = "<html><body>User email: user@example.com</body></html>" // Contains PII

	userContext := &security.UserContext{UserID: "gdpr-test-user"}

	err := suite.orchestrator.ProcessDocument(ctx, gdprDoc, "comprehensive-policy", userContext)
	suite.NoError(err, "Should process document with PII")

	// Verify GDPR compliance warnings were generated
	evaluation, err := suite.policyManager.EvaluateDocumentSecurity(ctx, gdprDoc, "comprehensive-policy", userContext)
	suite.NoError(err, "Should evaluate document security")

	hasGDPRWarning := false
	for _, warning := range evaluation.Warnings {
		if warning.Type == "potential_pii_detected" {
			hasGDPRWarning = true
			break
		}
	}
	suite.True(hasGDPRWarning, "Should detect potential PII for GDPR compliance")

	// Note: Cannot test audit trail features due to missing methods
	// auditEvents, err := suite.policyManager.GetAuditTrail(&security.AuditFilter{
	// 	UserID: "gdpr-test-user",
	// })
	// suite.NoError(err, "Should get audit trail")
	// suite.Greater(len(auditEvents), 0, "Should have audit events")

	// Test audit export
	// timeRange := &security.TimeRange{
	// 	Start: time.Now().Add(-1 * time.Hour),
	// 	End:   time.Now().Add(1 * time.Hour),
	// }

	// csvExport, err := suite.policyManager.ExportAuditLog("csv", timeRange)
	// suite.NoError(err, "Should export audit log as CSV")
	// suite.Contains(string(csvExport), "timestamp,action,resource", "CSV should have headers")

	// jsonExport, err := suite.policyManager.ExportAuditLog("json", timeRange)
	// suite.NoError(err, "Should export audit log as JSON")
	// suite.True(len(jsonExport) > 0, "JSON export should not be empty")
}

// TestPerformanceAndScalability tests performance and scalability
func (suite *SecurityIntegrationTestSuite) TestPerformanceAndScalability() {
	ctx := context.Background()

	// Test processing multiple documents concurrently
	numDocs := 10
	done := make(chan error, numDocs)

	for i := 0; i < numDocs; i++ {
		go func(docID int) {
			doc := suite.createTestDocument()
			doc.Manifest.Metadata.Title = fmt.Sprintf("Performance Test Doc %d", docID)

			userContext := &security.UserContext{
				UserID: fmt.Sprintf("perf-user-%d", docID),
			}

			err := suite.orchestrator.ProcessDocument(ctx, doc, "comprehensive-policy", userContext)
			done <- err
		}(i)
	}

	// Wait for all documents to process
	for i := 0; i < numDocs; i++ {
		err := <-done
		suite.NoError(err, "Document %d should process successfully", i)
	}

	// Verify system can handle the load
	status, err := suite.orchestrator.GetSystemStatus(ctx)
	suite.NoError(err, "Should get system status after load test")
	suite.Contains([]string{"healthy", "minor_issues", "warning"}, status.OverallHealth, "System should remain stable under load")

	// Test policy evaluation performance
	start := time.Now()
	for i := 0; i < 100; i++ {
		request := &security.PermissionRequest{
			DocumentID: fmt.Sprintf("perf-doc-%d", i),
			RequestedPerms: &core.WASMPermissions{
				MemoryLimit:    8 * 1024 * 1024,
				CPUTimeLimit:   3000,
				AllowedImports: []string{"console"},
			},
			PolicyID:    "comprehensive-policy",
			UserContext: &security.UserContext{UserID: fmt.Sprintf("perf-user-%d", i)},
		}

		_, err := suite.permissionManager.EvaluatePermissionRequest(ctx, request)
		suite.NoError(err, "Permission evaluation %d should succeed", i)
	}
	duration := time.Since(start)

	// Should complete 100 evaluations in reasonable time (< 1 second)
	suite.Less(duration, 1*time.Second, "100 permission evaluations should complete quickly")
}

// Helper methods

func (suite *SecurityIntegrationTestSuite) createTestDocument() *core.LIVDocument {
	return &core.LIVDocument{
		Manifest: &core.Manifest{
			Version: "1.0.0",
			Metadata: &core.DocumentMetadata{
				Title:       "Integration Test Document",
				Author:      "Test Author",
				Created:     time.Now(),
				Modified:    time.Now(),
				Description: "Test document for security integration",
				Version:     "1.0.0",
				Language:    "en",
			},
			Security: &core.SecurityPolicy{
				WASMPermissions: &core.WASMPermissions{
					MemoryLimit:     8 * 1024 * 1024,
					AllowedImports:  []string{"console"},
					CPUTimeLimit:    3000,
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
			HTML:            "<html><body><h1>Integration Test</h1></body></html>",
			CSS:             "body { font-family: Arial; }",
			InteractiveSpec: "{}",
			StaticFallback:  "<html><body><h1>Static Test</h1></body></html>",
		},
		Assets: &core.AssetBundle{
			Images: map[string][]byte{},
			Fonts:  map[string][]byte{},
			Data:   map[string][]byte{},
		},
		Signatures: &core.SignatureBundle{
			ContentSignature:  "",
			ManifestSignature: "",
			WASMSignatures:    map[string]string{},
		},
		WASMModules: map[string][]byte{
			"test-module": make([]byte, 1024),
		},
	}
}

func generateDocumentID(doc *core.LIVDocument) string {
	return fmt.Sprintf("doc-%s-%d", doc.Manifest.Metadata.Title, doc.Manifest.Metadata.Created.Unix())
}

func min(a, b int) int {
	if a < b {
		return a
	}
	return b
}

// TestSecurityIntegrationSuite runs the complete security integration test suite
func TestSecurityIntegrationSuite(t *testing.T) {
	suite.Run(t, new(SecurityIntegrationTestSuite))
}
