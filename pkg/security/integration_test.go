package security

import (
	"context"
	"testing"
	"time"

	"github.com/liv-format/liv/pkg/core"
)

// SimpleMockCryptoProvider implements core.CryptoProvider for testing
type SimpleMockCryptoProvider struct{}

func (mcp *SimpleMockCryptoProvider) GenerateKeyPair() (publicKey, privateKey []byte, err error) {
	return []byte("mock-public-key"), []byte("mock-private-key"), nil
}

func (mcp *SimpleMockCryptoProvider) Sign(data []byte, privateKey []byte) ([]byte, error) {
	return []byte("mock-signature"), nil
}

func (mcp *SimpleMockCryptoProvider) Verify(data []byte, signature []byte, publicKey []byte) bool {
	return string(signature) == "mock-signature"
}

func (mcp *SimpleMockCryptoProvider) Hash(data []byte) []byte {
	return []byte("mock-hash")
}

func (mcp *SimpleMockCryptoProvider) GenerateRandomBytes(length int) ([]byte, error) {
	bytes := make([]byte, length)
	for i := range bytes {
		bytes[i] = byte(i % 256)
	}
	return bytes, nil
}

func TestSecurityManagerIntegration(t *testing.T) {
	logger := &MockLogger{}
	metrics := &MockMetricsCollector{}
	crypto := &SimpleMockCryptoProvider{}

	// Create security manager
	sm := NewSecurityManager(crypto, logger, metrics)
	if sm == nil {
		t.Fatal("NewSecurityManager returned nil")
	}

	// Test signature operations
	content := []byte("test content")
	privateKey := []byte("mock-private-key")
	publicKey := []byte("mock-public-key")

	signature, err := sm.CreateSignature(content, privateKey)
	if err != nil {
		t.Fatalf("CreateSignature failed: %v", err)
	}

	valid := sm.ValidateSignature(content, signature, publicKey)
	if !valid {
		t.Error("ValidateSignature should return true for valid signature")
	}

	// Test WASM module validation
	wasmModule := []byte{0x00, 0x61, 0x73, 0x6d, 0x01, 0x00, 0x00, 0x00} // Valid WASM header
	permissions := &core.WASMPermissions{
		MemoryLimit:     4 * 1024 * 1024,
		CPUTimeLimit:    5000,
		AllowNetworking: false,
		AllowFileSystem: false,
		AllowedImports:  []string{},
	}

	err = sm.ValidateWASMModule(wasmModule, permissions)
	if err != nil {
		t.Errorf("ValidateWASMModule failed: %v", err)
	}

	// Test sandbox creation
	policy := &core.SecurityPolicy{
		WASMPermissions: permissions,
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
	}

	sandbox, err := sm.CreateSandbox(policy)
	if err != nil {
		t.Fatalf("CreateSandbox failed: %v", err)
	}

	// Test sandbox operations
	if sandbox.GetPermissions() != policy {
		t.Error("Sandbox permissions don't match expected policy")
	}

	// Test WASM loading in sandbox
	wasmConfig := &core.WASMModule{
		Name:        "test-module",
		Version:     "1.0.0",
		EntryPoint:  "main",
		Exports:     []string{"test_function"},
		Imports:     []string{},
		Permissions: permissions,
	}

	instance, err := sandbox.LoadWASM(context.Background(), wasmModule, wasmConfig)
	if err != nil {
		t.Fatalf("LoadWASM failed: %v", err)
	}

	// Test WASM instance operations
	exports := instance.GetExports()
	if len(exports) != 1 || exports[0] != "test_function" {
		t.Errorf("Expected exports [test_function], got %v", exports)
	}

	memUsage := instance.GetMemoryUsage()
	if memUsage == 0 {
		t.Error("Expected non-zero memory usage")
	}

	// Test function call
	result, err := instance.Call(context.Background(), "test_function", "arg1", "arg2")
	if err != nil {
		t.Errorf("Call failed: %v", err)
	}

	if result == nil {
		t.Error("Expected non-nil result from function call")
	}

	// Test permission evaluation
	requestedPerms := &core.WASMPermissions{
		MemoryLimit:     2 * 1024 * 1024, // Less than allowed
		CPUTimeLimit:    2000,            // Less than allowed
		AllowNetworking: false,
		AllowFileSystem: false,
		AllowedImports:  []string{},
	}

	allowed := sm.EvaluatePermissions(requestedPerms, policy)
	if !allowed {
		t.Error("Expected permissions to be allowed")
	}

	// Test excessive permission request
	excessivePerms := &core.WASMPermissions{
		MemoryLimit:     16 * 1024 * 1024, // More than allowed
		CPUTimeLimit:    2000,
		AllowNetworking: false,
		AllowFileSystem: false,
		AllowedImports:  []string{},
	}

	allowed = sm.EvaluatePermissions(excessivePerms, policy)
	if allowed {
		t.Error("Expected excessive permissions to be denied")
	}

	// Test security report generation
	doc := &core.LIVDocument{
		Manifest: &core.Manifest{
			Version: "1.0",
			Metadata: &core.DocumentMetadata{
				Title:    "Test Document",
				Author:   "Test Author",
				Created:  time.Now().Add(-time.Hour),
				Modified: time.Now(),
				Version:  "1.0.0",
				Language: "en",
			},
			Security: policy,
			Resources: map[string]*core.Resource{
				"content/index.html": {
					Hash: "test-hash",
					Size: 1024,
					Type: "text/html",
					Path: "content/index.html",
				},
			},
		},
		Signatures: &core.SignatureBundle{
			ContentSignature:  "test-content-sig",
			ManifestSignature: "test-manifest-sig",
		},
		WASMModules: map[string][]byte{
			"test-module": wasmModule,
		},
	}

	report := sm.GenerateSecurityReport(doc)
	if report == nil {
		t.Fatal("GenerateSecurityReport returned nil")
	}

	if !report.PermissionsValid {
		t.Error("Expected permissions to be valid")
	}

	// Cleanup
	err = instance.Terminate()
	if err != nil {
		t.Errorf("Terminate failed: %v", err)
	}

	err = sandbox.Destroy()
	if err != nil {
		t.Errorf("Destroy failed: %v", err)
	}
}

func TestResourceMonitoringIntegration(t *testing.T) {
	logger := &MockLogger{}
	metrics := &MockMetricsCollector{}
	crypto := &SimpleMockCryptoProvider{}

	sm := NewSecurityManager(crypto, logger, metrics)

	// Start resource monitoring
	ctx, cancel := context.WithTimeout(context.Background(), 2*time.Second)
	defer cancel()

	err := sm.StartResourceMonitoring(ctx)
	if err != nil {
		t.Fatalf("StartResourceMonitoring failed: %v", err)
	}

	// Create a sandbox and load a module
	policy := &core.SecurityPolicy{
		WASMPermissions: &core.WASMPermissions{
			MemoryLimit:     4 * 1024 * 1024,
			CPUTimeLimit:    5000,
			AllowNetworking: false,
			AllowFileSystem: false,
			AllowedImports:  []string{},
		},
		JSPermissions: &core.JSPermissions{
			ExecutionMode: "sandboxed",
			AllowedAPIs:   []string{},
			DOMAccess:     "read",
		},
		NetworkPolicy: &core.NetworkPolicy{
			AllowOutbound: false,
		},
		StoragePolicy: &core.StoragePolicy{
			AllowLocalStorage: false,
		},
	}

	sandbox, err := sm.CreateSandbox(policy)
	if err != nil {
		t.Fatalf("CreateSandbox failed: %v", err)
	}

	wasmModule := []byte{0x00, 0x61, 0x73, 0x6d, 0x01, 0x00, 0x00, 0x00}
	wasmConfig := &core.WASMModule{
		Name:        "monitored-module",
		Version:     "1.0.0",
		EntryPoint:  "main",
		Exports:     []string{"test_function"},
		Permissions: policy.WASMPermissions,
	}

	instance, err := sandbox.LoadWASM(context.Background(), wasmModule, wasmConfig)
	if err != nil {
		t.Fatalf("LoadWASM failed: %v", err)
	}

	// Let monitoring run for a short time
	time.Sleep(100 * time.Millisecond)

	// Get security metrics
	securityMetrics := sm.GetSecurityMetrics()
	if securityMetrics == nil {
		t.Error("GetSecurityMetrics returned nil")
	}

	if activeSessions, ok := securityMetrics["active_sessions"].(int); !ok || activeSessions == 0 {
		t.Error("Expected at least one active session")
	}

	// Cleanup
	instance.Terminate()
	sandbox.Destroy()

	err = sm.StopResourceMonitoring()
	if err != nil {
		t.Errorf("StopResourceMonitoring failed: %v", err)
	}

	// Clean up expired sessions
	cleaned := sm.CleanupExpiredSessions(time.Nanosecond) // Very short age to clean all
	if cleaned < 0 {
		t.Error("CleanupExpiredSessions returned negative count")
	}
}

func TestSecurityConfigurationManagement(t *testing.T) {
	logger := &MockLogger{}
	metrics := &MockMetricsCollector{}
	crypto := &SimpleMockCryptoProvider{}

	sm := NewSecurityManager(crypto, logger, metrics)

	// Test getting default configuration
	config := sm.GetSecurityConfiguration()
	if config == nil {
		t.Fatal("GetSecurityConfiguration returned nil")
	}

	if config.MaxMemoryPerModule == 0 {
		t.Error("Expected non-zero max memory per module")
	}

	if config.MaxCPUTimePerModule == 0 {
		t.Error("Expected non-zero max CPU time per module")
	}

	// Test updating configuration
	newConfig := &SecurityConfiguration{
		MaxMemoryPerModule:       64 * 1024 * 1024, // 64MB
		MaxCPUTimePerModule:      15 * time.Second, // 15 seconds
		MaxConcurrentModules:     5,
		AuditLogEnabled:          true,
		MetricsCollectionEnabled: true,
		StrictModeEnabled:        true,
	}

	err := sm.UpdateSecurityConfiguration(newConfig)
	if err != nil {
		t.Fatalf("UpdateSecurityConfiguration failed: %v", err)
	}

	updatedConfig := sm.GetSecurityConfiguration()
	if updatedConfig.MaxMemoryPerModule != newConfig.MaxMemoryPerModule {
		t.Error("Configuration not updated correctly")
	}

	if updatedConfig.StrictModeEnabled != newConfig.StrictModeEnabled {
		t.Error("Strict mode not updated correctly")
	}

	// Test nil configuration
	err = sm.UpdateSecurityConfiguration(nil)
	if err == nil {
		t.Error("Expected error for nil configuration")
	}
}
