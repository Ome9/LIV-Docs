package viewer

import (
	"strings"
	"testing"
	"time"

	"github.com/liv-format/liv/pkg/core"
)

func TestNewDocumentValidator(t *testing.T) {
	logger := &MockLogger{}

	validator := NewDocumentValidator(logger)

	if validator == nil {
		t.Fatal("NewDocumentValidator returned nil")
	}

	if validator.logger != logger {
		t.Error("logger not set correctly")
	}

	if validator.config == nil {
		t.Error("configuration not initialized")
	}
}

func TestDocumentValidator_ValidateDocument_ValidDocument(t *testing.T) {
	logger := &MockLogger{}
	validator := NewDocumentValidator(logger)

	document := createTestDocument()
	result := validator.ValidateDocument(document)

	if result == nil {
		t.Fatal("ValidateDocument returned nil result")
	}

	if !result.IsValid {
		t.Errorf("document should be valid, errors: %v", result.Errors)
	}

	if len(result.Errors) > 0 {
		t.Errorf("valid document should have no errors: %v", result.Errors)
	}
}

func TestDocumentValidator_ValidateDocument_NilDocument(t *testing.T) {
	logger := &MockLogger{}
	validator := NewDocumentValidator(logger)

	result := validator.ValidateDocument(nil)

	if result == nil {
		t.Fatal("ValidateDocument returned nil result")
	}

	if result.IsValid {
		t.Error("nil document should be invalid")
	}

	if len(result.Errors) == 0 {
		t.Error("nil document should have errors")
	}
}

func TestDocumentValidator_ValidateManifest_ValidManifest(t *testing.T) {
	logger := &MockLogger{}
	validator := NewDocumentValidator(logger)

	manifest := createTestDocument().Manifest
	result := validator.ValidateManifest(manifest)

	if result == nil {
		t.Fatal("ValidateManifest returned nil result")
	}

	if !result.IsValid {
		t.Errorf("manifest should be valid, errors: %v", result.Errors)
	}
}

func TestDocumentValidator_ValidateManifest_NilManifest(t *testing.T) {
	logger := &MockLogger{}
	validator := NewDocumentValidator(logger)

	result := validator.ValidateManifest(nil)

	if result == nil {
		t.Fatal("ValidateManifest returned nil result")
	}

	if result.IsValid {
		t.Error("nil manifest should be invalid")
	}

	if len(result.Errors) == 0 {
		t.Error("nil manifest should have errors")
	}
}

func TestDocumentValidator_ValidateManifest_MissingVersion(t *testing.T) {
	logger := &MockLogger{}
	validator := NewDocumentValidator(logger)

	manifest := createTestDocument().Manifest
	manifest.Version = ""

	result := validator.ValidateManifest(manifest)

	if result.IsValid {
		t.Error("manifest without version should be invalid")
	}

	found := false
	for _, err := range result.Errors {
		if err == "manifest version is required" {
			found = true
			break
		}
	}
	if !found {
		t.Error("should have error about missing version")
	}
}

func TestDocumentValidator_ValidateManifest_UnsupportedVersion(t *testing.T) {
	logger := &MockLogger{}
	validator := NewDocumentValidator(logger)

	manifest := createTestDocument().Manifest
	manifest.Version = "2.0"

	result := validator.ValidateManifest(manifest)

	// Should be valid but with warnings
	if !result.IsValid {
		t.Error("manifest with unsupported version should still be valid")
	}

	if len(result.Warnings) == 0 {
		t.Error("manifest with unsupported version should have warnings")
	}
}

func TestDocumentValidator_ValidateContent_ValidContent(t *testing.T) {
	logger := &MockLogger{}
	validator := NewDocumentValidator(logger)

	content := createTestDocument().Content
	result := validator.ValidateContent(content)

	if result == nil {
		t.Fatal("ValidateContent returned nil result")
	}

	if !result.IsValid {
		t.Errorf("content should be valid, errors: %v", result.Errors)
	}
}

func TestDocumentValidator_ValidateContent_NilContent(t *testing.T) {
	logger := &MockLogger{}
	validator := NewDocumentValidator(logger)

	result := validator.ValidateContent(nil)

	if result == nil {
		t.Fatal("ValidateContent returned nil result")
	}

	if result.IsValid {
		t.Error("nil content should be invalid")
	}

	if len(result.Errors) == 0 {
		t.Error("nil content should have errors")
	}
}

func TestDocumentValidator_ValidateContent_EmptyHTML(t *testing.T) {
	logger := &MockLogger{}
	validator := NewDocumentValidator(logger)

	content := &core.DocumentContent{
		HTML:           "",
		CSS:            "body { margin: 0; }",
		StaticFallback: "<html><body>Static</body></html>",
	}

	result := validator.ValidateContent(content)

	if result.IsValid {
		t.Error("content without HTML should be invalid")
	}

	found := false
	for _, err := range result.Errors {
		if err == "HTML content is required" {
			found = true
			break
		}
	}
	if !found {
		t.Error("should have error about missing HTML content")
	}
}

func TestDocumentValidator_ValidateContent_LargeContent(t *testing.T) {
	logger := &MockLogger{}
	validator := NewDocumentValidator(logger)

	// Create content that exceeds the size limit
	largeHTML := make([]byte, validator.config.MaxContentSize+1)
	for i := range largeHTML {
		largeHTML[i] = 'a'
	}

	content := &core.DocumentContent{
		HTML:           string(largeHTML),
		CSS:            "body { margin: 0; }",
		StaticFallback: "<html><body>Static</body></html>",
	}

	result := validator.ValidateContent(content)

	if result.IsValid {
		t.Error("content exceeding size limit should be invalid")
	}

	found := false
	for _, err := range result.Errors {
		if strings.Contains(err, "HTML content size") && strings.Contains(err, "exceeds limit") {
			found = true
			break
		}
	}
	if !found {
		t.Error("should have error about HTML content size")
	}
}

func TestDocumentValidator_ValidateAssets_ValidAssets(t *testing.T) {
	logger := &MockLogger{}
	validator := NewDocumentValidator(logger)

	assets := createTestDocument().Assets
	result := validator.ValidateAssets(assets)

	if result == nil {
		t.Fatal("ValidateAssets returned nil result")
	}

	if !result.IsValid {
		t.Errorf("assets should be valid, errors: %v", result.Errors)
	}
}

func TestDocumentValidator_ValidateAssets_NilAssets(t *testing.T) {
	logger := &MockLogger{}
	validator := NewDocumentValidator(logger)

	result := validator.ValidateAssets(nil)

	if result == nil {
		t.Fatal("ValidateAssets returned nil result")
	}

	// Nil assets should be valid but with warnings
	if !result.IsValid {
		t.Error("nil assets should be valid")
	}

	if len(result.Warnings) == 0 {
		t.Error("nil assets should have warnings")
	}
}

func TestDocumentValidator_ValidateAssets_LargeAsset(t *testing.T) {
	logger := &MockLogger{}
	validator := NewDocumentValidator(logger)

	// Create asset that exceeds size limit
	largeAsset := make([]byte, validator.config.MaxAssetSize+1)

	assets := &core.AssetBundle{
		Images: map[string][]byte{
			"large.png": largeAsset,
		},
	}

	result := validator.ValidateAssets(assets)

	if result.IsValid {
		t.Error("assets with oversized asset should be invalid")
	}

	found := false
	for _, err := range result.Errors {
		if len(err) > 20 && err[:20] == "image asset 'large.p" {
			found = true
			break
		}
	}
	if !found {
		t.Error("should have error about oversized asset")
	}
}

func TestDocumentValidator_ValidateAssets_EmptyAsset(t *testing.T) {
	logger := &MockLogger{}
	validator := NewDocumentValidator(logger)

	assets := &core.AssetBundle{
		Images: map[string][]byte{
			"empty.png": {},
		},
	}

	result := validator.ValidateAssets(assets)

	// Should be valid but with warnings
	if !result.IsValid {
		t.Error("assets with empty asset should be valid")
	}

	if len(result.Warnings) == 0 {
		t.Error("assets with empty asset should have warnings")
	}
}

func TestDocumentValidator_ValidateSignatures_ValidSignatures(t *testing.T) {
	logger := &MockLogger{}
	validator := NewDocumentValidator(logger)

	document := createTestDocument()
	document.Signatures = &core.SignatureBundle{
		ContentSignature:  "content-sig",
		ManifestSignature: "manifest-sig",
		WASMSignatures: map[string]string{
			"test-module": "wasm-sig",
		},
	}

	result := validator.ValidateSignatures(document)

	if result == nil {
		t.Fatal("ValidateSignatures returned nil result")
	}

	if !result.IsValid {
		t.Errorf("signatures should be valid, errors: %v", result.Errors)
	}
}

func TestDocumentValidator_ValidateSignatures_MissingSignatures(t *testing.T) {
	logger := &MockLogger{}
	validator := NewDocumentValidator(logger)

	document := createTestDocument()
	document.Signatures = &core.SignatureBundle{
		ContentSignature:  "",
		ManifestSignature: "",
		WASMSignatures:    map[string]string{},
	}

	result := validator.ValidateSignatures(document)

	// Should be valid but with warnings
	if !result.IsValid {
		t.Error("document with missing signatures should be valid")
	}

	if len(result.Warnings) == 0 {
		t.Error("document with missing signatures should have warnings")
	}
}

func TestDocumentValidator_ValidateMetadata(t *testing.T) {
	logger := &MockLogger{}
	validator := NewDocumentValidator(logger)

	// Test valid metadata
	metadata := createTestDocument().Manifest.Metadata
	errors, warnings := validator.validateMetadata(metadata)

	if len(errors) > 0 {
		t.Errorf("valid metadata should have no errors: %v", errors)
	}

	// Test missing required fields
	emptyMetadata := &core.DocumentMetadata{}
	errors, warnings = validator.validateMetadata(emptyMetadata)

	if len(errors) == 0 {
		t.Error("empty metadata should have errors")
	}

	// Should have errors for missing title, author, version, language
	expectedErrors := []string{"title", "author", "version", "language"}
	for _, expected := range expectedErrors {
		found := false
		for _, err := range errors {
			if len(err) > len(expected) && err[len(err)-len(expected)-12:len(err)-12] == expected {
				found = true
				break
			}
		}
		if !found {
			t.Errorf("should have error about missing %s", expected)
		}
	}

	// Test invalid dates
	invalidMetadata := &core.DocumentMetadata{
		Title:    "Test",
		Author:   "Test Author",
		Version:  "1.0.0",
		Language: "en",
		Created:  time.Now(),
		Modified: time.Now().Add(-time.Hour), // Modified before created
	}

	errors, warnings = validator.validateMetadata(invalidMetadata)
	if len(errors) == 0 {
		t.Error("metadata with invalid dates should have errors")
	}

	// Test future modification date
	futureMetadata := &core.DocumentMetadata{
		Title:    "Test",
		Author:   "Test Author",
		Version:  "1.0.0",
		Language: "en",
		Created:  time.Now().Add(-time.Hour),
		Modified: time.Now().Add(2 * time.Hour), // Future date
	}

	errors, warnings = validator.validateMetadata(futureMetadata)
	if len(warnings) == 0 {
		t.Error("metadata with future modification date should have warnings")
	}
}

func TestDocumentValidator_ValidateSecurityPolicy(t *testing.T) {
	logger := &MockLogger{}
	validator := NewDocumentValidator(logger)

	// Test valid security policy
	policy := createTestDocument().Manifest.Security
	errors, warnings := validator.validateSecurityPolicy(policy)

	if len(errors) > 0 {
		t.Errorf("valid security policy should have no errors: %v", errors)
	}

	// Test missing WASM permissions
	invalidPolicy := &core.SecurityPolicy{
		WASMPermissions: nil,
		JSPermissions: &core.JSPermissions{
			ExecutionMode: "sandboxed",
		},
	}

	errors, warnings = validator.validateSecurityPolicy(invalidPolicy)
	if len(errors) == 0 {
		t.Error("policy without WASM permissions should have errors")
	}

	// Test missing JS permissions
	invalidPolicy2 := &core.SecurityPolicy{
		WASMPermissions: &core.WASMPermissions{
			MemoryLimit: 4 * 1024 * 1024,
		},
		JSPermissions: nil,
	}

	errors, warnings = validator.validateSecurityPolicy(invalidPolicy2)
	if len(errors) == 0 {
		t.Error("policy without JS permissions should have errors")
	}

	// Test high limits (should generate warnings)
	highLimitPolicy := &core.SecurityPolicy{
		WASMPermissions: &core.WASMPermissions{
			MemoryLimit:     512 * 1024 * 1024, // 512MB
			CPUTimeLimit:    60000,              // 60 seconds
			AllowNetworking: true,
			AllowFileSystem: true,
		},
		JSPermissions: &core.JSPermissions{
			ExecutionMode: "trusted",
		},
	}

	errors, warnings = validator.validateSecurityPolicy(highLimitPolicy)
	if len(warnings) == 0 {
		t.Error("policy with high limits should have warnings")
	}
}

func TestDocumentValidator_ValidateWASMModules(t *testing.T) {
	logger := &MockLogger{}
	validator := NewDocumentValidator(logger)

	// Test valid WASM modules
	modules := map[string][]byte{
		"valid-module": {0x00, 0x61, 0x73, 0x6d, 0x01, 0x00, 0x00, 0x00},
	}

	errors, warnings := validator.validateWASMModules(modules)
	if len(errors) > 0 {
		t.Errorf("valid WASM modules should have no errors: %v", errors)
	}

	// Test empty module
	emptyModules := map[string][]byte{
		"empty-module": {},
	}

	errors, warnings = validator.validateWASMModules(emptyModules)
	if len(errors) == 0 {
		t.Error("empty WASM module should have errors")
	}

	// Test invalid magic number
	invalidModules := map[string][]byte{
		"invalid-module": {0x00, 0x00, 0x00, 0x00, 0x01, 0x00, 0x00, 0x00},
	}

	errors, warnings = validator.validateWASMModules(invalidModules)
	if len(errors) == 0 {
		t.Error("WASM module with invalid magic number should have errors")
	}

	// Test unsupported version
	unsupportedVersionModules := map[string][]byte{
		"unsupported-module": {0x00, 0x61, 0x73, 0x6d, 0x02, 0x00, 0x00, 0x00},
	}

	errors, warnings = validator.validateWASMModules(unsupportedVersionModules)
	if len(warnings) == 0 {
		t.Error("WASM module with unsupported version should have warnings")
	}
}

func TestDocumentValidator_UpdateConfiguration(t *testing.T) {
	logger := &MockLogger{}
	validator := NewDocumentValidator(logger)

	newConfig := &ValidatorConfiguration{
		StrictMode:         true,
		ValidateContent:    false,
		ValidateAssets:     false,
		ValidateSignatures: false,
		ValidateWASM:       false,
		MaxContentSize:     20 * 1024 * 1024,
		MaxAssetSize:       100 * 1024 * 1024,
		RequiredResources:  []string{"content/index.html"},
	}

	err := validator.UpdateConfiguration(newConfig)
	if err != nil {
		t.Errorf("UpdateConfiguration failed: %v", err)
	}

	config := validator.GetConfiguration()
	if config.StrictMode != newConfig.StrictMode {
		t.Error("configuration not updated correctly")
	}

	if config.MaxContentSize != newConfig.MaxContentSize {
		t.Error("max content size not updated correctly")
	}

	// Test nil configuration
	err = validator.UpdateConfiguration(nil)
	if err == nil {
		t.Error("UpdateConfiguration should fail for nil config")
	}
}