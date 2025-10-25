package wasm

import (
	"context"
	"testing"
	"time"

	"github.com/liv-format/liv/pkg/core"
)

// MockSecurityManager for testing
type MockSecurityManager struct {
	validateModuleFunc func([]byte, *core.WASMPermissions) error
}

func (msm *MockSecurityManager) ValidateSignature(content []byte, signature string, publicKey []byte) bool {
	return true
}

func (msm *MockSecurityManager) CreateSignature(content []byte, privateKey []byte) (string, error) {
	return "mock-signature", nil
}

func (msm *MockSecurityManager) ValidateWASMModule(module []byte, permissions *core.WASMPermissions) error {
	if msm.validateModuleFunc != nil {
		return msm.validateModuleFunc(module, permissions)
	}
	return nil
}

func (msm *MockSecurityManager) CreateSandbox(policy *core.SecurityPolicy) (core.Sandbox, error) {
	return nil, nil
}

func (msm *MockSecurityManager) EvaluatePermissions(requested *core.WASMPermissions, policy *core.SecurityPolicy) bool {
	return true
}

func (msm *MockSecurityManager) GenerateSecurityReport(doc *core.LIVDocument) *core.SecurityReport {
	return &core.SecurityReport{IsValid: true}
}

// MockLogger for testing
type MockLogger struct {
	logs []string
}

func (ml *MockLogger) Debug(msg string, fields ...interface{}) {
	ml.logs = append(ml.logs, "DEBUG: "+msg)
}

func (ml *MockLogger) Info(msg string, fields ...interface{}) {
	ml.logs = append(ml.logs, "INFO: "+msg)
}

func (ml *MockLogger) Warn(msg string, fields ...interface{}) {
	ml.logs = append(ml.logs, "WARN: "+msg)
}

func (ml *MockLogger) Error(msg string, fields ...interface{}) {
	ml.logs = append(ml.logs, "ERROR: "+msg)
}

func (ml *MockLogger) Fatal(msg string, fields ...interface{}) {
	ml.logs = append(ml.logs, "FATAL: "+msg)
}

// MockMetricsCollector for testing
type MockMetricsCollector struct {
	events []map[string]interface{}
}

func (mmc *MockMetricsCollector) RecordDocumentLoad(size int64, duration int64) {}

func (mmc *MockMetricsCollector) RecordWASMExecution(module string, duration int64, memoryUsed uint64) {}

func (mmc *MockMetricsCollector) RecordSecurityEvent(eventType string, details map[string]interface{}) {
	event := map[string]interface{}{
		"type": eventType,
	}
	for k, v := range details {
		event[k] = v
	}
	mmc.events = append(mmc.events, event)
}

func (mmc *MockMetricsCollector) GetMetrics() map[string]interface{} {
	return map[string]interface{}{
		"events": mmc.events,
	}
}

func TestNewWASMLoader(t *testing.T) {
	securityMgr := &MockSecurityManager{}
	logger := &MockLogger{}
	metrics := &MockMetricsCollector{}

	loader := NewWASMLoader(securityMgr, logger, metrics)

	if loader == nil {
		t.Fatal("NewWASMLoader returned nil")
	}

	if loader.securityMgr != securityMgr {
		t.Error("security manager not set correctly")
	}

	if loader.logger != logger {
		t.Error("logger not set correctly")
	}

	if loader.metrics != metrics {
		t.Error("metrics collector not set correctly")
	}

	if loader.config == nil {
		t.Error("configuration not initialized")
	}

	if len(loader.loadedModules) != 0 {
		t.Error("loaded modules should be empty initially")
	}
}

func TestWASMLoader_LoadModule_ValidModule(t *testing.T) {
	securityMgr := &MockSecurityManager{}
	logger := &MockLogger{}
	metrics := &MockMetricsCollector{}

	loader := NewWASMLoader(securityMgr, logger, metrics)

	// Valid WASM module data (magic number + version)
	moduleData := []byte{0x00, 0x61, 0x73, 0x6d, 0x01, 0x00, 0x00, 0x00}

	ctx := context.Background()
	instance, err := loader.LoadModule(ctx, "test-module", moduleData)

	if err != nil {
		t.Fatalf("LoadModule failed: %v", err)
	}

	if instance == nil {
		t.Fatal("LoadModule returned nil instance")
	}

	// Check that module is in loaded modules
	modules := loader.ListModules()
	if len(modules) != 1 || modules[0] != "test-module" {
		t.Errorf("expected 1 module named 'test-module', got %v", modules)
	}

	// Check module info
	info, err := loader.GetModuleInfo("test-module")
	if err != nil {
		t.Errorf("GetModuleInfo failed: %v", err)
	}

	if info.Name != "test-module" {
		t.Errorf("expected module name 'test-module', got %s", info.Name)
	}
}

func TestWASMLoader_LoadModule_InvalidModule(t *testing.T) {
	securityMgr := &MockSecurityManager{}
	logger := &MockLogger{}
	metrics := &MockMetricsCollector{}

	loader := NewWASMLoader(securityMgr, logger, metrics)

	// Invalid WASM module data (wrong magic number)
	moduleData := []byte{0x00, 0x00, 0x00, 0x00, 0x01, 0x00, 0x00, 0x00}

	ctx := context.Background()
	_, err := loader.LoadModule(ctx, "invalid-module", moduleData)

	if err == nil {
		t.Error("LoadModule should have failed for invalid module")
	}

	// Check that no modules are loaded
	modules := loader.ListModules()
	if len(modules) != 0 {
		t.Errorf("expected 0 modules, got %v", modules)
	}
}

func TestWASMLoader_LoadModule_EmptyName(t *testing.T) {
	securityMgr := &MockSecurityManager{}
	logger := &MockLogger{}
	metrics := &MockMetricsCollector{}

	loader := NewWASMLoader(securityMgr, logger, metrics)

	moduleData := []byte{0x00, 0x61, 0x73, 0x6d, 0x01, 0x00, 0x00, 0x00}

	ctx := context.Background()
	_, err := loader.LoadModule(ctx, "", moduleData)

	if err == nil {
		t.Error("LoadModule should have failed for empty module name")
	}
}

func TestWASMLoader_LoadModule_EmptyData(t *testing.T) {
	securityMgr := &MockSecurityManager{}
	logger := &MockLogger{}
	metrics := &MockMetricsCollector{}

	loader := NewWASMLoader(securityMgr, logger, metrics)

	ctx := context.Background()
	_, err := loader.LoadModule(ctx, "test-module", []byte{})

	if err == nil {
		t.Error("LoadModule should have failed for empty module data")
	}
}

func TestWASMLoader_LoadModule_ExceedsSize(t *testing.T) {
	securityMgr := &MockSecurityManager{}
	logger := &MockLogger{}
	metrics := &MockMetricsCollector{}

	loader := NewWASMLoader(securityMgr, logger, metrics)

	// Create module data that exceeds the size limit
	moduleData := make([]byte, loader.config.MaxModuleSize+1)
	copy(moduleData, []byte{0x00, 0x61, 0x73, 0x6d, 0x01, 0x00, 0x00, 0x00})

	ctx := context.Background()
	_, err := loader.LoadModule(ctx, "large-module", moduleData)

	if err == nil {
		t.Error("LoadModule should have failed for oversized module")
	}
}

func TestWASMLoader_LoadModule_ReuseExisting(t *testing.T) {
	securityMgr := &MockSecurityManager{}
	logger := &MockLogger{}
	metrics := &MockMetricsCollector{}

	loader := NewWASMLoader(securityMgr, logger, metrics)

	moduleData := []byte{0x00, 0x61, 0x73, 0x6d, 0x01, 0x00, 0x00, 0x00}

	ctx := context.Background()

	// Load module first time
	instance1, err := loader.LoadModule(ctx, "test-module", moduleData)
	if err != nil {
		t.Fatalf("first LoadModule failed: %v", err)
	}

	// Load same module again
	instance2, err := loader.LoadModule(ctx, "test-module", moduleData)
	if err != nil {
		t.Fatalf("second LoadModule failed: %v", err)
	}

	// Should return the same instance
	if instance1 != instance2 {
		t.Error("LoadModule should reuse existing instance")
	}

	// Should still have only one module
	modules := loader.ListModules()
	if len(modules) != 1 {
		t.Errorf("expected 1 module, got %d", len(modules))
	}
}

func TestWASMLoader_UnloadModule(t *testing.T) {
	securityMgr := &MockSecurityManager{}
	logger := &MockLogger{}
	metrics := &MockMetricsCollector{}

	loader := NewWASMLoader(securityMgr, logger, metrics)

	moduleData := []byte{0x00, 0x61, 0x73, 0x6d, 0x01, 0x00, 0x00, 0x00}

	ctx := context.Background()
	_, err := loader.LoadModule(ctx, "test-module", moduleData)
	if err != nil {
		t.Fatalf("LoadModule failed: %v", err)
	}

	// Unload the module
	err = loader.UnloadModule("test-module")
	if err != nil {
		t.Errorf("UnloadModule failed: %v", err)
	}

	// Check that module is no longer loaded
	modules := loader.ListModules()
	if len(modules) != 0 {
		t.Errorf("expected 0 modules after unload, got %v", modules)
	}

	// Try to get info for unloaded module
	_, err = loader.GetModuleInfo("test-module")
	if err == nil {
		t.Error("GetModuleInfo should fail for unloaded module")
	}
}

func TestWASMLoader_UnloadModule_NotFound(t *testing.T) {
	securityMgr := &MockSecurityManager{}
	logger := &MockLogger{}
	metrics := &MockMetricsCollector{}

	loader := NewWASMLoader(securityMgr, logger, metrics)

	err := loader.UnloadModule("nonexistent-module")
	if err == nil {
		t.Error("UnloadModule should fail for nonexistent module")
	}
}

func TestWASMLoader_ValidateModule(t *testing.T) {
	securityMgr := &MockSecurityManager{}
	logger := &MockLogger{}
	metrics := &MockMetricsCollector{}

	loader := NewWASMLoader(securityMgr, logger, metrics)

	tests := []struct {
		name        string
		data        []byte
		expectError bool
	}{
		{
			name:        "valid WASM module",
			data:        []byte{0x00, 0x61, 0x73, 0x6d, 0x01, 0x00, 0x00, 0x00},
			expectError: false,
		},
		{
			name:        "invalid magic number",
			data:        []byte{0x00, 0x00, 0x00, 0x00, 0x01, 0x00, 0x00, 0x00},
			expectError: true,
		},
		{
			name:        "invalid version",
			data:        []byte{0x00, 0x61, 0x73, 0x6d, 0x02, 0x00, 0x00, 0x00},
			expectError: true,
		},
		{
			name:        "too small",
			data:        []byte{0x00, 0x61, 0x73},
			expectError: true,
		},
		{
			name:        "empty data",
			data:        []byte{},
			expectError: true,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			err := loader.ValidateModule(tt.data)
			if tt.expectError && err == nil {
				t.Error("ValidateModule should have failed")
			}
			if !tt.expectError && err != nil {
				t.Errorf("ValidateModule should have succeeded: %v", err)
			}
		})
	}
}

func TestWASMLoader_GetLoadedModuleStats(t *testing.T) {
	securityMgr := &MockSecurityManager{}
	logger := &MockLogger{}
	metrics := &MockMetricsCollector{}

	loader := NewWASMLoader(securityMgr, logger, metrics)

	// Initially no modules
	stats := loader.GetLoadedModuleStats()
	if stats["loaded_modules"].(int) != 0 {
		t.Error("expected 0 loaded modules initially")
	}

	// Load a module
	moduleData := []byte{0x00, 0x61, 0x73, 0x6d, 0x01, 0x00, 0x00, 0x00}
	ctx := context.Background()
	_, err := loader.LoadModule(ctx, "test-module", moduleData)
	if err != nil {
		t.Fatalf("LoadModule failed: %v", err)
	}

	// Check stats after loading
	stats = loader.GetLoadedModuleStats()
	if stats["loaded_modules"].(int) != 1 {
		t.Error("expected 1 loaded module")
	}

	if stats["total_size"].(int64) != int64(len(moduleData)) {
		t.Errorf("expected total size %d, got %v", len(moduleData), stats["total_size"])
	}
}

func TestWASMLoader_CleanupExpiredModules(t *testing.T) {
	securityMgr := &MockSecurityManager{}
	logger := &MockLogger{}
	metrics := &MockMetricsCollector{}

	loader := NewWASMLoader(securityMgr, logger, metrics)

	// Set very short cache expiry time for testing
	loader.config.CacheExpiryTime = 1 * time.Millisecond
	loader.config.EnableCaching = true

	moduleData := []byte{0x00, 0x61, 0x73, 0x6d, 0x01, 0x00, 0x00, 0x00}

	ctx := context.Background()
	_, err := loader.LoadModule(ctx, "test-module", moduleData)
	if err != nil {
		t.Fatalf("LoadModule failed: %v", err)
	}

	// Wait for module to expire
	time.Sleep(10 * time.Millisecond)

	// Cleanup expired modules
	cleaned := loader.CleanupExpiredModules()
	if cleaned != 1 {
		t.Errorf("expected 1 module to be cleaned up, got %d", cleaned)
	}

	// Check that module is no longer loaded
	modules := loader.ListModules()
	if len(modules) != 0 {
		t.Errorf("expected 0 modules after cleanup, got %v", modules)
	}
}

func TestWASMLoader_MaxModulesLimit(t *testing.T) {
	securityMgr := &MockSecurityManager{}
	logger := &MockLogger{}
	metrics := &MockMetricsCollector{}

	loader := NewWASMLoader(securityMgr, logger, metrics)

	// Set low module limit for testing
	loader.config.MaxModules = 2

	moduleData := []byte{0x00, 0x61, 0x73, 0x6d, 0x01, 0x00, 0x00, 0x00}
	ctx := context.Background()

	// Load modules up to the limit
	_, err := loader.LoadModule(ctx, "module1", moduleData)
	if err != nil {
		t.Fatalf("LoadModule 1 failed: %v", err)
	}

	_, err = loader.LoadModule(ctx, "module2", moduleData)
	if err != nil {
		t.Fatalf("LoadModule 2 failed: %v", err)
	}

	// Third module should trigger eviction
	_, err = loader.LoadModule(ctx, "module3", moduleData)
	if err != nil {
		t.Fatalf("LoadModule 3 failed: %v", err)
	}

	// Should still have only 2 modules (one evicted)
	modules := loader.ListModules()
	if len(modules) != 2 {
		t.Errorf("expected 2 modules after eviction, got %d", len(modules))
	}
}

func TestWASMLoader_Configuration(t *testing.T) {
	securityMgr := &MockSecurityManager{}
	logger := &MockLogger{}
	metrics := &MockMetricsCollector{}

	loader := NewWASMLoader(securityMgr, logger, metrics)

	// Get initial configuration
	config := loader.GetConfiguration()
	if config == nil {
		t.Fatal("GetConfiguration returned nil")
	}

	// Update configuration
	newConfig := &LoaderConfiguration{
		MaxModules:          5,
		MaxModuleSize:       8 * 1024 * 1024,
		ModuleTimeout:       15 * time.Second,
		MemoryLimit:         64 * 1024 * 1024,
		EnableCaching:       false,
		CacheExpiryTime:     30 * time.Minute,
		AllowUnsafeModules:  true,
		StrictValidation:    false,
	}

	err := loader.UpdateConfiguration(newConfig)
	if err != nil {
		t.Errorf("UpdateConfiguration failed: %v", err)
	}

	// Verify configuration was updated
	updatedConfig := loader.GetConfiguration()
	if updatedConfig.MaxModules != newConfig.MaxModules {
		t.Error("configuration not updated correctly")
	}

	// Test nil configuration
	err = loader.UpdateConfiguration(nil)
	if err == nil {
		t.Error("UpdateConfiguration should fail for nil config")
	}
}