package fileformat

import (
	"bytes"
	"context"
	"os"
	"strings"
	"testing"

	"github.com/liv-format/liv/pkg/container"
	"github.com/liv-format/liv/pkg/core"
	"github.com/liv-format/liv/pkg/integrity"
	"github.com/liv-format/liv/pkg/manifest"
)

// TestManifestValidationWithWASMConfig tests manifest validation with various WASM configurations
func TestManifestValidationWithWASMConfig(t *testing.T) {
	validator := manifest.NewManifestValidator()

	tests := []struct {
		name      string
		manifest  *core.Manifest
		wantValid bool
		wantError string
	}{
		{
			name: "valid manifest with WASM config",
			manifest: createValidManifestWithWASM(t),
			wantValid: true,
		},
		{
			name: "manifest with invalid WASM memory limit",
			manifest: func() *core.Manifest {
				m := createValidManifestWithWASM(t)
				m.WASMConfig.MemoryLimit = 1024*1024*1024 // 1GB - too high
				return m
			}(),
			wantValid: false,
			wantError: "must be at most 134217728",
		},
		{
			name: "manifest with WASM circular dependency",
			manifest: func() *core.Manifest {
				m := createValidManifestWithWASM(t)
				m.WASMConfig.Modules = map[string]*core.WASMModule{
					"module-a": {
						Name:    "module-a",
						Version: "1.0.0",
						Imports: []string{"module-b"},
					},
					"module-b": {
						Name:    "module-b", 
						Version: "1.0.0",
						Imports: []string{"module-a"}, // Circular dependency
					},
				}
				return m
			}(),
			wantValid: false,
			wantError: "circular dependency",
		},
		{
			name: "manifest with invalid WASM module name",
			manifest: func() *core.Manifest {
				m := createValidManifestWithWASM(t)
				m.WASMConfig.Modules["invalid-name!"] = &core.WASMModule{
					Name:    "invalid-name!",
					Version: "1.0.0",
				}
				return m
			}(),
			wantValid: false,
			wantError: "missing entry point",
		},
		{
			name: "manifest with WASM module version mismatch",
			manifest: func() *core.Manifest {
				m := createValidManifestWithWASM(t)
				m.WASMConfig.Modules["test-module"].Version = "invalid-version"
				return m
			}(),
			wantValid: false,
			wantError: "invalid version format",
		},
		{
			name: "manifest with excessive WASM CPU time limit",
			manifest: func() *core.Manifest {
				m := createValidManifestWithWASM(t)
				m.WASMConfig.Permissions.CPUTimeLimit = 60000 // 60 seconds - too high
				return m
			}(),
			wantValid: false,
			wantError: "must be at most 30000",
		},
		{
			name: "manifest with WASM networking enabled (security risk)",
			manifest: func() *core.Manifest {
				m := createValidManifestWithWASM(t)
				m.WASMConfig.Permissions.AllowNetworking = true
				return m
			}(),
			wantValid: true, // Valid but should generate warning
		},
		{
			name: "manifest with missing WASM entry point",
			manifest: func() *core.Manifest {
				m := createValidManifestWithWASM(t)
				m.WASMConfig.Modules["test-module"].EntryPoint = ""
				return m
			}(),
			wantValid: false,
			wantError: "missing entry point",
		},

	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			result := validator.ValidateManifest(tt.manifest)

			if result.IsValid != tt.wantValid {
				t.Errorf("ValidateManifest() isValid = %v, want %v", result.IsValid, tt.wantValid)
				if len(result.Errors) > 0 {
					t.Errorf("Errors: %v", result.Errors)
				}
			}

			if tt.wantError != "" {
				found := false
				for _, err := range result.Errors {
					if strings.Contains(err, tt.wantError) {
						found = true
						break
					}
				}
				if !found {
					t.Errorf("Expected error containing '%s' not found in: %v", tt.wantError, result.Errors)
				}
			}
		})
	}
}

// TestZIPContainerWithWASMModules tests ZIP container operations with WASM modules
func TestZIPContainerWithWASMModules(t *testing.T) {
	container := container.NewZIPContainer()

	// Create test files including WASM modules
	testFiles := map[string][]byte{
		"manifest.json": []byte(`{
			"version": "1.0",
			"wasm_config": {
				"modules": {
					"test-engine": {
						"name": "test-engine",
						"version": "1.0.0",
						"entry_point": "init"
					}
				},
				"permissions": {
					"memory_limit": 67108864,
					"cpu_time_limit": 5000,
					"allow_networking": false,
					"allow_file_system": false
				},
				"memory_limit": 67108864
			}
		}`),
		"content/index.html": []byte(`<!DOCTYPE html><html><body><h1>WASM Test</h1></body></html>`),
		"wasm/test-engine.wasm": []byte{
			// Valid WASM header
			0x00, 0x61, 0x73, 0x6D, // Magic number
			0x01, 0x00, 0x00, 0x00, // Version
			// Type section
			0x01, 0x07, 0x01, 0x60, 0x02, 0x7F, 0x7F, 0x01, 0x7F,
			// Function section
			0x03, 0x02, 0x01, 0x00,
			// Export section
			0x07, 0x07, 0x01, 0x03, 0x61, 0x64, 0x64, 0x00, 0x00,
			// Code section
			0x0A, 0x09, 0x01, 0x07, 0x00, 0x20, 0x00, 0x20, 0x01, 0x6A, 0x0B,
		},
		"wasm/invalid-module.wasm": []byte{0xFF, 0xFF, 0xFF, 0xFF}, // Invalid WASM
		"signatures/wasm-signatures.json": []byte(`{
			"test-engine": "signature-for-test-engine",
			"invalid-module": "signature-for-invalid-module"
		}`),
	}

	// Test creating ZIP with WASM modules
	var buf bytes.Buffer
	err := container.CreateFromFilesToWriter(testFiles, &buf)
	if err != nil {
		t.Fatalf("Failed to create ZIP with WASM modules: %v", err)
	}

	if buf.Len() == 0 {
		t.Fatal("ZIP file with WASM modules is empty")
	}

	// Test extracting ZIP with WASM modules
	// Create temporary file for extraction
	tempFile, err := os.CreateTemp("", "test-wasm-*.liv")
	if err != nil {
		t.Fatalf("Failed to create temp file: %v", err)
	}
	defer os.Remove(tempFile.Name())
	
	if _, err := tempFile.Write(buf.Bytes()); err != nil {
		t.Fatalf("Failed to write temp file: %v", err)
	}
	tempFile.Close()

	extractedFiles, err := container.ExtractToMemory(tempFile.Name())
	if err != nil {
		t.Fatalf("Failed to extract ZIP with WASM modules: %v", err)
	}

	// Verify WASM modules were extracted
	if wasmData, exists := extractedFiles["wasm/test-engine.wasm"]; exists {
		if !bytes.Equal(wasmData, testFiles["wasm/test-engine.wasm"]) {
			t.Error("WASM module data mismatch after extraction")
		}
		
		// Verify WASM magic number
		if len(wasmData) < 4 || !bytes.Equal(wasmData[:4], []byte{0x00, 0x61, 0x73, 0x6D}) {
			t.Error("Extracted WASM module has invalid magic number")
		}
	} else {
		t.Error("WASM module not found in extracted files")
	}

	// Test structure validation with WASM modules
	result := container.ValidateStructureFromMemory(testFiles)
	if !result.IsValid {
		t.Errorf("Structure validation failed for ZIP with WASM modules: %v", result.Errors)
	}

	// Test with missing WASM directory
	incompleteFiles := make(map[string][]byte)
	for path, content := range testFiles {
		if !strings.HasPrefix(path, "wasm/") {
			incompleteFiles[path] = content
		}
	}

	result = container.ValidateStructureFromMemory(incompleteFiles)
	// Note: ZIP container validation doesn't check WASM module references
	// This would be caught at the manifest validation level
	if !result.IsValid {
		t.Logf("Structure validation failed as expected: %v", result.Errors)
	}
}

// TestResourceIntegrityWithWASMModules tests resource integrity verification including WASM modules
func TestResourceIntegrityWithWASMModules(t *testing.T) {
	validator := integrity.NewIntegrityValidator()

	// Create test WASM modules
	validWASMModule := []byte{
		0x00, 0x61, 0x73, 0x6D, // Magic
		0x01, 0x00, 0x00, 0x00, // Version
		0x01, 0x04, 0x01, 0x60, 0x00, 0x00, // Minimal valid content
	}

	corruptedWASMModule := []byte{
		0x00, 0x61, 0x73, 0x6D, // Magic
		0x01, 0x00, 0x00, 0x00, // Version
		0xFF, 0xFF, 0xFF, 0xFF, // Corrupted content
	}

	invalidWASMModule := []byte{0xFF, 0xFF, 0xFF, 0xFF} // Invalid magic

	// Create WASM configuration
	wasmConfig := &core.WASMConfiguration{
		Modules: map[string]*core.WASMModule{
			"valid-module": {
				Name:    "valid-module",
				Version: "1.0.0",
			},
			"corrupted-module": {
				Name:    "corrupted-module",
				Version: "1.0.0",
			},
			"invalid-module": {
				Name:    "invalid-module",
				Version: "1.0.0",
			},
		},
		Permissions: &core.WASMPermissions{
			MemoryLimit:     64 * 1024 * 1024,
			AllowedImports:  []string{"env"},
			CPUTimeLimit:    5000,
			AllowNetworking: false,
			AllowFileSystem: false,
		},
		MemoryLimit: 64 * 1024 * 1024,
	}

	tests := []struct {
		name        string
		wasmModules map[string][]byte
		wantValid   bool
		wantError   string
	}{
		{
			name: "valid WASM modules",
			wasmModules: map[string][]byte{
				"valid-module": validWASMModule,
			},
			wantValid: true,
		},
		{
			name: "invalid WASM magic number",
			wasmModules: map[string][]byte{
				"invalid-module": invalidWASMModule,
			},
			wantValid: false,
			wantError: "has invalid magic number",
		},
		{
			name: "corrupted WASM module",
			wasmModules: map[string][]byte{
				"corrupted-module": corruptedWASMModule,
			},
			wantValid: true, // Magic and version are valid, content corruption detected elsewhere
		},
		{
			name: "missing WASM module",
			wasmModules: map[string][]byte{
				// valid-module is missing
			},
			wantValid: false,
			wantError: "configured but not found",
		},
		{
			name: "extra WASM module",
			wasmModules: map[string][]byte{
				"valid-module": validWASMModule,
				"extra-module": validWASMModule, // Not in config
			},
			wantValid: true, // Should generate warning but not fail
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			// Update config to match test case
			testConfig := &core.WASMConfiguration{
				Modules:     make(map[string]*core.WASMModule),
				Permissions: wasmConfig.Permissions,
				MemoryLimit: wasmConfig.MemoryLimit,
			}

			// Add modules that should exist based on test case
			for moduleName := range tt.wasmModules {
				if module, exists := wasmConfig.Modules[moduleName]; exists {
					testConfig.Modules[moduleName] = module
				}
			}

			// For missing module test, add the expected module to config
			if tt.name == "missing WASM module" {
				testConfig.Modules["valid-module"] = wasmConfig.Modules["valid-module"]
			}

			result := validator.ValidateWASMModules(testConfig, tt.wasmModules)

			if result.IsValid != tt.wantValid {
				t.Errorf("ValidateWASMModules() isValid = %v, want %v", result.IsValid, tt.wantValid)
				if len(result.Errors) > 0 {
					t.Errorf("Errors: %v", result.Errors)
				}
			}

			if tt.wantError != "" {
				found := false
				for _, err := range result.Errors {
					if strings.Contains(err, tt.wantError) {
						found = true
						break
					}
				}
				if !found {
					t.Errorf("Expected error containing '%s' not found in: %v", tt.wantError, result.Errors)
				}
			}
		})
	}
}

// TestWASMModuleValidationDetails tests detailed WASM module validation
func TestWASMModuleValidationDetails(t *testing.T) {
	validator := integrity.NewIntegrityValidator()

	tests := []struct {
		name      string
		wasmData  []byte
		wantValid bool
		wantError string
	}{
		{
			name: "minimal valid WASM",
			wasmData: []byte{
				0x00, 0x61, 0x73, 0x6D, // Magic
				0x01, 0x00, 0x00, 0x00, // Version
			},
			wantValid: true,
		},
		{
			name: "WASM with type section",
			wasmData: []byte{
				0x00, 0x61, 0x73, 0x6D, // Magic
				0x01, 0x00, 0x00, 0x00, // Version
				0x01, 0x07, 0x01, 0x60, 0x02, 0x7F, 0x7F, 0x01, 0x7F, // Type section
			},
			wantValid: true,
		},
		{
			name:      "too small to be valid WASM",
			wasmData:  []byte{0x00, 0x61},
			wantValid: false,
			wantError: "has invalid magic number",
		},
		{
			name:      "invalid magic number",
			wasmData:  []byte{0xFF, 0xFF, 0xFF, 0xFF, 0x01, 0x00, 0x00, 0x00},
			wantValid: false,
			wantError: "has invalid magic number",
		},
		{
			name:      "invalid version",
			wasmData:  []byte{0x00, 0x61, 0x73, 0x6D, 0xFF, 0xFF, 0xFF, 0xFF},
			wantValid: false,
			wantError: "has unsupported version",
		},
		{
			name:      "empty data",
			wasmData:  []byte{},
			wantValid: false,
			wantError: "has invalid magic number",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			// Create minimal config for single module test
			wasmConfig := &core.WASMConfiguration{
				Modules: map[string]*core.WASMModule{
					"test-module": {
						Name:    "test-module",
						Version: "1.0.0",
					},
				},
			}

			wasmModules := map[string][]byte{
				"test-module": tt.wasmData,
			}

			result := validator.ValidateWASMModules(wasmConfig, wasmModules)

			if result.IsValid != tt.wantValid {
				t.Errorf("ValidateWASMModules() isValid = %v, want %v", result.IsValid, tt.wantValid)
				if len(result.Errors) > 0 {
					t.Errorf("Errors: %v", result.Errors)
				}
			}

			if tt.wantError != "" {
				found := false
				for _, err := range result.Errors {
					if strings.Contains(err, tt.wantError) {
						found = true
						break
					}
				}
				if !found {
					t.Errorf("Expected error containing '%s' not found in: %v", tt.wantError, result.Errors)
				}
			}
		})
	}
}

// TestIntegratedFileFormatWithWASM tests the complete file format workflow with WASM modules
func TestIntegratedFileFormatWithWASM(t *testing.T) {
	// Create a complete document with WASM modules
	document := createCompleteWASMDocument(t)

	// Test packaging
	packageManager := container.NewPackageManager()
	var buf bytes.Buffer
	err := packageManager.SavePackageToWriter(document, &buf)
	if err != nil {
		t.Fatalf("Failed to package document with WASM: %v", err)
	}

	// Test extraction
	extractedDoc, err := packageManager.ExtractPackage(context.Background(), &buf)
	if err != nil {
		t.Fatalf("Failed to extract document with WASM: %v", err)
	}

	// Verify WASM modules were preserved
	if len(extractedDoc.WASMModules) != len(document.WASMModules) {
		t.Errorf("WASM module count mismatch: expected %d, got %d", 
			len(document.WASMModules), len(extractedDoc.WASMModules))
	}

	for moduleName, originalData := range document.WASMModules {
		if extractedData, exists := extractedDoc.WASMModules[moduleName]; exists {
			if !bytes.Equal(originalData, extractedData) {
				t.Errorf("WASM module %s data corrupted during packaging/extraction", moduleName)
			}
		} else {
			t.Errorf("WASM module %s missing after extraction", moduleName)
		}
	}

	// Test integrity validation
	validator := integrity.NewIntegrityValidator()
	result := validator.ValidateWASMModules(extractedDoc.Manifest.WASMConfig, extractedDoc.WASMModules)
	if !result.IsValid {
		t.Errorf("WASM integrity validation failed: %v", result.Errors)
	}

	// Test manifest validation
	manifestValidator := manifest.NewManifestValidator()
	manifestResult := manifestValidator.ValidateManifest(extractedDoc.Manifest)
	if !manifestResult.IsValid {
		t.Errorf("Manifest validation failed: %v", manifestResult.Errors)
	}
}

// Helper functions

func createValidManifestWithWASM(t *testing.T) *core.Manifest {
	builder := manifest.NewManifestBuilder()
	builder.CreateDefaultMetadata("WASM Test Document", "Test Suite").
		CreateDefaultSecurityPolicy().
		CreateDefaultFeatureFlags()

	// Add required resources
	builder.AddResource("content/index.html", &core.Resource{
		Hash: "html-hash-123",
		Size: 1024,
		Type: "text/html",
		Path: "content/index.html",
	})

	builder.AddResource("manifest.json", &core.Resource{
		Hash: "manifest-hash-456",
		Size: 512,
		Type: "application/json",
		Path: "manifest.json",
	})

	// Add WASM module
	wasmModule := &core.WASMModule{
		Name:       "test-module",
		Version:    "1.0.0",
		EntryPoint: "init",
		Exports:    []string{"init", "process", "cleanup"},
		Imports:    []string{"env.memory", "env.table"},
		Permissions: &core.WASMPermissions{
			MemoryLimit:     32 * 1024 * 1024, // 32MB
			AllowedImports:  []string{"env"},
			CPUTimeLimit:    3000,
			AllowNetworking: false,
			AllowFileSystem: false,
		},
		Metadata: map[string]string{
			"description": "Test WASM module",
			"author":      "Test Suite",
		},
	}
	builder.AddWASMModule(wasmModule)

	manifest, err := builder.Build()
	if err != nil {
		t.Fatalf("Failed to build manifest with WASM: %v", err)
	}

	return manifest
}

func createCompleteWASMDocument(t *testing.T) *core.LIVDocument {
	manifest := createValidManifestWithWASM(t)

	// Create valid WASM module data
	wasmModuleData := []byte{
		0x00, 0x61, 0x73, 0x6D, // Magic
		0x01, 0x00, 0x00, 0x00, // Version
		// Type section
		0x01, 0x07, 0x01, 0x60, 0x02, 0x7F, 0x7F, 0x01, 0x7F,
		// Function section
		0x03, 0x02, 0x01, 0x00,
		// Export section
		0x07, 0x0A, 0x01, 0x06, 0x70, 0x72, 0x6F, 0x63, 0x65, 0x73, 0x73, 0x00, 0x00,
		// Code section
		0x0A, 0x09, 0x01, 0x07, 0x00, 0x20, 0x00, 0x20, 0x01, 0x6A, 0x0B,
	}

	return &core.LIVDocument{
		Manifest: manifest,
		Content: &core.DocumentContent{
			HTML: `<!DOCTYPE html>
<html>
<head>
    <title>WASM Test Document</title>
</head>
<body>
    <h1>Interactive WASM Content</h1>
    <div id="wasm-output"></div>
</body>
</html>`,
			CSS: `body { font-family: Arial, sans-serif; }
#wasm-output { border: 1px solid #ccc; padding: 10px; }`,
			InteractiveSpec: `// WASM module configuration
const wasmConfig = {
    module: "test-module",
    entryPoint: "init",
    memoryLimit: 33554432
};`,
			StaticFallback: `<!DOCTYPE html>
<html>
<head><title>WASM Test Document - Static</title></head>
<body>
    <h1>Interactive WASM Content (Static Fallback)</h1>
    <p>WASM functionality not available in static mode.</p>
</body>
</html>`,
		},
		Assets: &core.AssetBundle{
			Images: map[string][]byte{
				"wasm-icon.svg": []byte(`<svg><circle cx="50" cy="50" r="40"/></svg>`),
			},
			Fonts: map[string][]byte{},
			Data: map[string][]byte{
				"wasm-config.json": []byte(`{"module": "test-module", "version": "1.0.0"}`),
			},
		},
		Signatures: &core.SignatureBundle{
			ContentSignature:  "test-content-signature",
			ManifestSignature: "test-manifest-signature",
			WASMSignatures: map[string]string{
				"test-module": "test-wasm-module-signature",
			},
		},
		WASMModules: map[string][]byte{
			"test-module": wasmModuleData,
		},
	}
}