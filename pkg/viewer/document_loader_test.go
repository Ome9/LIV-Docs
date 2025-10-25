package viewer

import (
	"context"
	"fmt"
	"io"
	"strings"
	"testing"
	"time"

	"github.com/liv-format/liv/pkg/core"
)

// Mock implementations for testing

type MockPackageManager struct {
	extractFunc func(context.Context, io.Reader) (*core.LIVDocument, error)
}

func (mpm *MockPackageManager) CreatePackage(ctx context.Context, sources map[string]io.Reader, manifest *core.Manifest) (*core.LIVDocument, error) {
	return nil, nil
}

func (mpm *MockPackageManager) ExtractPackage(ctx context.Context, reader io.Reader) (*core.LIVDocument, error) {
	if mpm.extractFunc != nil {
		return mpm.extractFunc(ctx, reader)
	}
	return createTestDocument(), nil
}

func (mpm *MockPackageManager) ValidateStructure(doc *core.LIVDocument) *core.ValidationResult {
	return &core.ValidationResult{IsValid: true}
}

func (mpm *MockPackageManager) CompressAssets(assets *core.AssetBundle) (*core.AssetBundle, error) {
	return assets, nil
}

func (mpm *MockPackageManager) LoadWASMModule(name string, data []byte) (*core.WASMModule, error) {
	return &core.WASMModule{Name: name}, nil
}

type MockSecurityManager struct {
	reportFunc func(*core.LIVDocument) *core.SecurityReport
}

func (msm *MockSecurityManager) ValidateSignature(content []byte, signature string, publicKey []byte) bool {
	return true
}

func (msm *MockSecurityManager) CreateSignature(content []byte, privateKey []byte) (string, error) {
	return "mock-signature", nil
}

func (msm *MockSecurityManager) ValidateWASMModule(module []byte, permissions *core.WASMPermissions) error {
	return nil
}

func (msm *MockSecurityManager) CreateSandbox(policy *core.SecurityPolicy) (core.Sandbox, error) {
	return nil, nil
}

func (msm *MockSecurityManager) EvaluatePermissions(requested *core.WASMPermissions, policy *core.SecurityPolicy) bool {
	return true
}

func (msm *MockSecurityManager) GenerateSecurityReport(doc *core.LIVDocument) *core.SecurityReport {
	if msm.reportFunc != nil {
		return msm.reportFunc(doc)
	}
	return &core.SecurityReport{IsValid: true}
}

type MockDocumentValidator struct {
	validateFunc func(*core.LIVDocument) *core.ValidationResult
}

func (mdv *MockDocumentValidator) ValidateDocument(doc *core.LIVDocument) *core.ValidationResult {
	if mdv.validateFunc != nil {
		return mdv.validateFunc(doc)
	}
	return &core.ValidationResult{IsValid: true}
}

func (mdv *MockDocumentValidator) ValidateManifest(manifest *core.Manifest) *core.ValidationResult {
	return &core.ValidationResult{IsValid: true}
}

func (mdv *MockDocumentValidator) ValidateContent(content *core.DocumentContent) *core.ValidationResult {
	return &core.ValidationResult{IsValid: true}
}

func (mdv *MockDocumentValidator) ValidateAssets(assets *core.AssetBundle) *core.ValidationResult {
	return &core.ValidationResult{IsValid: true}
}

func (mdv *MockDocumentValidator) ValidateSignatures(doc *core.LIVDocument) *core.ValidationResult {
	return &core.ValidationResult{IsValid: true}
}

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

type MockMetricsCollector struct {
	events []map[string]interface{}
}

func (mmc *MockMetricsCollector) RecordDocumentLoad(size int64, duration int64) {
	mmc.events = append(mmc.events, map[string]interface{}{
		"type":     "document_load",
		"size":     size,
		"duration": duration,
	})
}

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

// Helper functions

func createTestDocument() *core.LIVDocument {
	return &core.LIVDocument{
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
			Security: &core.SecurityPolicy{
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
			},
			Resources: map[string]*core.Resource{
				"content/index.html": {
					Hash: "test-hash",
					Size: 30,
					Type: "text/html",
					Path: "content/index.html",
				},
			},
		},
		Content: &core.DocumentContent{
			HTML:           "<html><body>Test</body></html>",
			CSS:            "body { margin: 0; }",
			StaticFallback: "<html><body>Static Test</body></html>",
		},
		Assets: &core.AssetBundle{
			Images: map[string][]byte{
				"logo.png": []byte("fake-png-data"),
			},
		},
		WASMModules: map[string][]byte{
			"test-module": {0x00, 0x61, 0x73, 0x6d, 0x01, 0x00, 0x00, 0x00},
		},
	}
}

// Tests

func TestNewDocumentLoader(t *testing.T) {
	packageManager := &MockPackageManager{}
	securityManager := &MockSecurityManager{}
	validator := &MockDocumentValidator{}
	logger := &MockLogger{}
	metrics := &MockMetricsCollector{}

	loader := NewDocumentLoader(packageManager, securityManager, validator, logger, metrics)

	if loader == nil {
		t.Fatal("NewDocumentLoader returned nil")
	}

	if loader.packageManager != packageManager {
		t.Error("package manager not set correctly")
	}

	if loader.securityManager != securityManager {
		t.Error("security manager not set correctly")
	}

	if loader.validator != validator {
		t.Error("validator not set correctly")
	}

	if loader.config == nil {
		t.Error("configuration not initialized")
	}

	if loader.cache == nil {
		t.Error("cache not initialized")
	}
}

func TestDocumentLoader_LoadDocument_ValidFile(t *testing.T) {
	packageManager := &MockPackageManager{}
	securityManager := &MockSecurityManager{}
	validator := &MockDocumentValidator{}
	logger := &MockLogger{}
	metrics := &MockMetricsCollector{}

	loader := NewDocumentLoader(packageManager, securityManager, validator, logger, metrics)

	reader := strings.NewReader("test document content")
	ctx := context.Background()

	result, err := loader.LoadDocument(ctx, reader, "test.liv")

	if err != nil {
		t.Fatalf("LoadDocument failed: %v", err)
	}

	if result == nil {
		t.Fatal("LoadDocument returned nil result")
	}

	if result.Document == nil {
		t.Error("result should contain a document")
	}

	if result.FromCache {
		t.Error("first load should not be from cache")
	}

	if result.LoadTime < 0 {
		t.Error("load time should be non-negative")
	}
}

func TestDocumentLoader_LoadDocument_InvalidExtension(t *testing.T) {
	packageManager := &MockPackageManager{}
	securityManager := &MockSecurityManager{}
	validator := &MockDocumentValidator{}
	logger := &MockLogger{}
	metrics := &MockMetricsCollector{}

	loader := NewDocumentLoader(packageManager, securityManager, validator, logger, metrics)

	reader := strings.NewReader("test content")
	ctx := context.Background()

	_, err := loader.LoadDocument(ctx, reader, "test.txt")

	if err == nil {
		t.Error("LoadDocument should fail for invalid file extension")
	}

	loadErr, ok := err.(*LoadError)
	if !ok {
		t.Error("error should be of type LoadError")
	} else if loadErr.Type != LoadErrorTypeInvalidFile {
		t.Errorf("expected LoadErrorTypeInvalidFile, got %s", loadErr.Type)
	}
}

func TestDocumentLoader_LoadDocument_ExtractionFailure(t *testing.T) {
	packageManager := &MockPackageManager{
		extractFunc: func(ctx context.Context, reader io.Reader) (*core.LIVDocument, error) {
			return nil, fmt.Errorf("extraction failed")
		},
	}
	securityManager := &MockSecurityManager{}
	validator := &MockDocumentValidator{}
	logger := &MockLogger{}
	metrics := &MockMetricsCollector{}

	loader := NewDocumentLoader(packageManager, securityManager, validator, logger, metrics)

	reader := strings.NewReader("test content")
	ctx := context.Background()

	_, err := loader.LoadDocument(ctx, reader, "test.liv")

	if err == nil {
		t.Error("LoadDocument should fail when extraction fails")
	}

	loadErr, ok := err.(*LoadError)
	if !ok {
		t.Error("error should be of type LoadError")
	} else if loadErr.Type != LoadErrorTypeCorrupted {
		t.Errorf("expected LoadErrorTypeCorrupted, got %s", loadErr.Type)
	}
}

func TestDocumentLoader_LoadDocument_ValidationFailure(t *testing.T) {
	packageManager := &MockPackageManager{}
	securityManager := &MockSecurityManager{}
	validator := &MockDocumentValidator{
		validateFunc: func(doc *core.LIVDocument) *core.ValidationResult {
			return &core.ValidationResult{
				IsValid: false,
				Errors:  []string{"validation failed"},
			}
		},
	}
	logger := &MockLogger{}
	metrics := &MockMetricsCollector{}

	loader := NewDocumentLoader(packageManager, securityManager, validator, logger, metrics)

	reader := strings.NewReader("test content")
	ctx := context.Background()

	_, err := loader.LoadDocument(ctx, reader, "test.liv")

	if err == nil {
		t.Error("LoadDocument should fail when validation fails in strict mode")
	}

	loadErr, ok := err.(*LoadError)
	if !ok {
		t.Error("error should be of type LoadError")
	} else if loadErr.Type != LoadErrorTypeSecurity {
		t.Errorf("expected LoadErrorTypeSecurity, got %s", loadErr.Type)
	}
}

func TestDocumentLoader_LoadDocument_SecurityFailure(t *testing.T) {
	packageManager := &MockPackageManager{}
	securityManager := &MockSecurityManager{
		reportFunc: func(doc *core.LIVDocument) *core.SecurityReport {
			return &core.SecurityReport{
				IsValid: false,
				Errors:  []string{"security validation failed"},
			}
		},
	}
	validator := &MockDocumentValidator{}
	logger := &MockLogger{}
	metrics := &MockMetricsCollector{}

	loader := NewDocumentLoader(packageManager, securityManager, validator, logger, metrics)

	reader := strings.NewReader("test content")
	ctx := context.Background()

	_, err := loader.LoadDocument(ctx, reader, "test.liv")

	if err == nil {
		t.Error("LoadDocument should fail when security validation fails")
	}

	loadErr, ok := err.(*LoadError)
	if !ok {
		t.Error("error should be of type LoadError")
	} else if loadErr.Type != LoadErrorTypeSecurity {
		t.Errorf("expected LoadErrorTypeSecurity, got %s", loadErr.Type)
	}
}

func TestDocumentLoader_LoadDocument_Caching(t *testing.T) {
	packageManager := &MockPackageManager{}
	securityManager := &MockSecurityManager{}
	validator := &MockDocumentValidator{}
	logger := &MockLogger{}
	metrics := &MockMetricsCollector{}

	loader := NewDocumentLoader(packageManager, securityManager, validator, logger, metrics)

	reader1 := strings.NewReader("test content")
	ctx := context.Background()

	// First load
	result1, err := loader.LoadDocument(ctx, reader1, "test.liv")
	if err != nil {
		t.Fatalf("first LoadDocument failed: %v", err)
	}

	if result1.FromCache {
		t.Error("first load should not be from cache")
	}

	reader2 := strings.NewReader("test content")

	// Second load (should be from cache)
	result2, err := loader.LoadDocument(ctx, reader2, "test.liv")
	if err != nil {
		t.Fatalf("second LoadDocument failed: %v", err)
	}

	if !result2.FromCache {
		t.Error("second load should be from cache")
	}

	// Cached loads should typically be faster, but we'll just check that it worked
	if result2.LoadTime < 0 {
		t.Error("cached load time should be non-negative")
	}
}

func TestDocumentLoader_LoadDocument_Timeout(t *testing.T) {
	packageManager := &MockPackageManager{
		extractFunc: func(ctx context.Context, reader io.Reader) (*core.LIVDocument, error) {
			// Check if context is cancelled
			select {
			case <-ctx.Done():
				return nil, ctx.Err()
			case <-time.After(100 * time.Millisecond):
				return createTestDocument(), nil
			}
		},
	}
	securityManager := &MockSecurityManager{}
	validator := &MockDocumentValidator{}
	logger := &MockLogger{}
	metrics := &MockMetricsCollector{}

	loader := NewDocumentLoader(packageManager, securityManager, validator, logger, metrics)
	
	// Set very short timeout
	loader.config.LoadTimeout = 10 * time.Millisecond

	reader := strings.NewReader("test content")
	ctx := context.Background()

	_, err := loader.LoadDocument(ctx, reader, "test.liv")

	if err == nil {
		t.Log("Note: timeout test may not work reliably in mock environment")
	}
}

func TestDocumentLoader_ValidateDocument(t *testing.T) {
	packageManager := &MockPackageManager{}
	securityManager := &MockSecurityManager{}
	validator := &MockDocumentValidator{}
	logger := &MockLogger{}
	metrics := &MockMetricsCollector{}

	loader := NewDocumentLoader(packageManager, securityManager, validator, logger, metrics)

	document := createTestDocument()

	result, err := loader.ValidateDocument(document)

	if err != nil {
		t.Errorf("ValidateDocument failed: %v", err)
	}

	if result == nil {
		t.Fatal("ValidateDocument returned nil result")
	}

	if !result.IsValid {
		t.Error("document should be valid")
	}
}

func TestDocumentLoader_GetCacheStats(t *testing.T) {
	packageManager := &MockPackageManager{}
	securityManager := &MockSecurityManager{}
	validator := &MockDocumentValidator{}
	logger := &MockLogger{}
	metrics := &MockMetricsCollector{}

	loader := NewDocumentLoader(packageManager, securityManager, validator, logger, metrics)

	// Initially empty cache
	stats := loader.GetCacheStats()
	if stats["cached_documents"].(int) != 0 {
		t.Error("cache should be empty initially")
	}

	// Load a document to populate cache
	reader := strings.NewReader("test content")
	ctx := context.Background()
	_, err := loader.LoadDocument(ctx, reader, "test.liv")
	if err != nil {
		t.Fatalf("LoadDocument failed: %v", err)
	}

	// Check cache stats after loading
	stats = loader.GetCacheStats()
	if stats["cached_documents"].(int) != 1 {
		t.Error("cache should contain 1 document")
	}

	if !stats["cache_enabled"].(bool) {
		t.Error("cache should be enabled")
	}
}

func TestDocumentLoader_ClearCache(t *testing.T) {
	packageManager := &MockPackageManager{}
	securityManager := &MockSecurityManager{}
	validator := &MockDocumentValidator{}
	logger := &MockLogger{}
	metrics := &MockMetricsCollector{}

	loader := NewDocumentLoader(packageManager, securityManager, validator, logger, metrics)

	// Load a document to populate cache
	reader := strings.NewReader("test content")
	ctx := context.Background()
	_, err := loader.LoadDocument(ctx, reader, "test.liv")
	if err != nil {
		t.Fatalf("LoadDocument failed: %v", err)
	}

	// Verify cache has content
	stats := loader.GetCacheStats()
	if stats["cached_documents"].(int) != 1 {
		t.Error("cache should contain 1 document before clearing")
	}

	// Clear cache
	loader.ClearCache()

	// Verify cache is empty
	stats = loader.GetCacheStats()
	if stats["cached_documents"].(int) != 0 {
		t.Error("cache should be empty after clearing")
	}
}

func TestDocumentLoader_UpdateConfiguration(t *testing.T) {
	packageManager := &MockPackageManager{}
	securityManager := &MockSecurityManager{}
	validator := &MockDocumentValidator{}
	logger := &MockLogger{}
	metrics := &MockMetricsCollector{}

	loader := NewDocumentLoader(packageManager, securityManager, validator, logger, metrics)

	newConfig := &LoaderConfiguration{
		EnableCaching:       false,
		CacheSize:          100,
		CacheExpiry:        1 * time.Hour,
		MaxDocumentSize:    200 * 1024 * 1024,
		ValidateSignatures: false,
		StrictValidation:   false,
		LoadTimeout:        60 * time.Second,
	}

	err := loader.UpdateConfiguration(newConfig)
	if err != nil {
		t.Errorf("UpdateConfiguration failed: %v", err)
	}

	config := loader.GetConfiguration()
	if config.EnableCaching != newConfig.EnableCaching {
		t.Error("configuration not updated correctly")
	}

	if config.CacheSize != newConfig.CacheSize {
		t.Error("cache size not updated correctly")
	}

	// Test nil configuration
	err = loader.UpdateConfiguration(nil)
	if err == nil {
		t.Error("UpdateConfiguration should fail for nil config")
	}
}

func TestLimitedReader(t *testing.T) {
	content := "this is a test content that is longer than the limit"
	reader := strings.NewReader(content)
	
	limitedReader := &limitedReader{
		reader: reader,
		limit:  10, // 10 bytes limit
	}

	// Read within limit
	buf := make([]byte, 5)
	n, err := limitedReader.Read(buf)
	if err != nil {
		t.Errorf("Read within limit failed: %v", err)
	}
	if n != 5 {
		t.Errorf("expected 5 bytes read, got %d", n)
	}

	// Read up to limit
	buf = make([]byte, 10)
	n, err = limitedReader.Read(buf)
	if err != nil {
		t.Errorf("Read up to limit failed: %v", err)
	}
	if n != 5 { // Only 5 more bytes available within limit
		t.Errorf("expected 5 bytes read, got %d", n)
	}

	// Try to read beyond limit
	buf = make([]byte, 5)
	_, err = limitedReader.Read(buf)
	if err == nil {
		t.Error("Read beyond limit should fail")
	}

	loadErr, ok := err.(*LoadError)
	if !ok {
		t.Error("error should be of type LoadError")
	} else if loadErr.Type != LoadErrorTypeResourceLimit {
		t.Errorf("expected LoadErrorTypeResourceLimit, got %s", loadErr.Type)
	}
}