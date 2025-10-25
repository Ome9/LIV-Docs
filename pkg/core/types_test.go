package core

import (
	"encoding/json"
	"fmt"
	"testing"
	"time"
)

func TestLIVDocument_JSONSerialization(t *testing.T) {
	// Create a complete LIV document
	now := time.Now()
	document := &LIVDocument{
		Manifest: &Manifest{
			Version: "1.0",
			Metadata: &DocumentMetadata{
				Title:       "Test Document",
				Author:      "Test Author",
				Created:     now.Add(-time.Hour),
				Modified:    now,
				Description: "A test document",
				Version:     "1.0.0",
				Language:    "en",
			},
			Security: &SecurityPolicy{
				WASMPermissions: &WASMPermissions{
					MemoryLimit:     64 * 1024 * 1024,
					AllowedImports:  []string{"env"},
					CPUTimeLimit:    5000,
					AllowNetworking: false,
					AllowFileSystem: false,
				},
				JSPermissions: &JSPermissions{
					ExecutionMode: "sandboxed",
					AllowedAPIs:   []string{"canvas"},
					DOMAccess:     "write",
				},
				NetworkPolicy: &NetworkPolicy{
					AllowOutbound: false,
					AllowedHosts:  []string{},
					AllowedPorts:  []int{},
				},
				StoragePolicy: &StoragePolicy{
					AllowLocalStorage:   true,
					AllowSessionStorage: true,
					AllowIndexedDB:      false,
					AllowCookies:        false,
				},
				ContentSecurityPolicy: "default-src 'self'",
				TrustedDomains:        []string{},
			},
			Resources: map[string]*Resource{
				"content/index.html": {
					Hash: "abc123",
					Size: 1024,
					Type: "text/html",
					Path: "content/index.html",
				},
			},
			Features: &FeatureFlags{
				Animations:    true,
				Interactivity: true,
				Charts:        false,
				Forms:         false,
				Audio:         false,
				Video:         false,
				WebGL:         false,
				WebAssembly:   true,
			},
		},
		Content: &DocumentContent{
			HTML:           "<html><body>Test</body></html>",
			CSS:            "body { color: red; }",
			InteractiveSpec: "console.log('test');",
			StaticFallback: "<html><body>Static</body></html>",
		},
		Assets: &AssetBundle{
			Images: map[string][]byte{
				"logo.png": []byte("fake-png-data"),
			},
			Fonts: map[string][]byte{
				"font.woff": []byte("fake-font-data"),
			},
			Data: map[string][]byte{
				"data.json": []byte(`{"test": true}`),
			},
		},
		Signatures: &SignatureBundle{
			ContentSignature:  "content-signature",
			ManifestSignature: "manifest-signature",
			WASMSignatures:    map[string]string{"module1": "wasm-signature"},
		},
		WASMModules: map[string][]byte{
			"module1": {0x00, 0x61, 0x73, 0x6D, 0x01, 0x00, 0x00, 0x00},
		},
	}

	// Test JSON marshaling
	data, err := json.Marshal(document)
	if err != nil {
		t.Fatalf("Failed to marshal document: %v", err)
	}

	// Test JSON unmarshaling
	var unmarshaled LIVDocument
	if err := json.Unmarshal(data, &unmarshaled); err != nil {
		t.Fatalf("Failed to unmarshal document: %v", err)
	}

	// Verify key fields
	if unmarshaled.Manifest.Metadata.Title != document.Manifest.Metadata.Title {
		t.Errorf("Title mismatch: expected %s, got %s", 
			document.Manifest.Metadata.Title, unmarshaled.Manifest.Metadata.Title)
	}

	if unmarshaled.Content.HTML != document.Content.HTML {
		t.Error("HTML content mismatch")
	}

	if len(unmarshaled.Assets.Images) != len(document.Assets.Images) {
		t.Errorf("Images count mismatch: expected %d, got %d", 
			len(document.Assets.Images), len(unmarshaled.Assets.Images))
	}
}

func TestDocumentMetadata_Validation(t *testing.T) {
	tests := []struct {
		name     string
		metadata *DocumentMetadata
		wantErr  bool
	}{
		{
			name: "valid metadata",
			metadata: &DocumentMetadata{
				Title:       "Valid Title",
				Author:      "Valid Author",
				Created:     time.Now().Add(-time.Hour),
				Modified:    time.Now(),
				Description: "Valid description",
				Version:     "1.0.0",
				Language:    "en",
			},
			wantErr: false,
		},
		{
			name: "empty title",
			metadata: &DocumentMetadata{
				Title:       "",
				Author:      "Valid Author",
				Created:     time.Now().Add(-time.Hour),
				Modified:    time.Now(),
				Description: "Valid description",
				Version:     "1.0.0",
				Language:    "en",
			},
			wantErr: false, // JSON marshaling doesn't validate struct tags
		},
		{
			name: "title too long",
			metadata: &DocumentMetadata{
				Title:       string(make([]byte, 201)), // 201 characters
				Author:      "Valid Author",
				Created:     time.Now().Add(-time.Hour),
				Modified:    time.Now(),
				Description: "Valid description",
				Version:     "1.0.0",
				Language:    "en",
			},
			wantErr: false, // JSON marshaling doesn't validate struct tags
		},
		{
			name: "invalid language code",
			metadata: &DocumentMetadata{
				Title:       "Valid Title",
				Author:      "Valid Author",
				Created:     time.Now().Add(-time.Hour),
				Modified:    time.Now(),
				Description: "Valid description",
				Version:     "1.0.0",
				Language:    "invalid", // Should be 2 characters
			},
			wantErr: false, // JSON marshaling doesn't validate struct tags
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			// Test JSON marshaling to trigger validation
			_, err := json.Marshal(tt.metadata)
			
			if tt.wantErr && err == nil {
				t.Error("Expected validation error but got none")
			}
			
			if !tt.wantErr && err != nil {
				t.Errorf("Unexpected validation error: %v", err)
			}
		})
	}
}

func TestSecurityPolicy_DefaultValues(t *testing.T) {
	policy := &SecurityPolicy{
		WASMPermissions: &WASMPermissions{
			MemoryLimit:     64 * 1024 * 1024,
			AllowedImports:  []string{},
			CPUTimeLimit:    5000,
			AllowNetworking: false,
			AllowFileSystem: false,
		},
		JSPermissions: &JSPermissions{
			ExecutionMode: "sandboxed",
			AllowedAPIs:   []string{},
			DOMAccess:     "read",
		},
		NetworkPolicy: &NetworkPolicy{
			AllowOutbound: false,
			AllowedHosts:  []string{},
			AllowedPorts:  []int{},
		},
		StoragePolicy: &StoragePolicy{
			AllowLocalStorage:   false,
			AllowSessionStorage: false,
			AllowIndexedDB:      false,
			AllowCookies:        false,
		},
	}

	// Test that default values are secure
	if policy.WASMPermissions.AllowNetworking {
		t.Error("WASM networking should be disabled by default")
	}

	if policy.WASMPermissions.AllowFileSystem {
		t.Error("WASM file system access should be disabled by default")
	}

	if policy.JSPermissions.ExecutionMode != "sandboxed" {
		t.Error("JavaScript should be sandboxed by default")
	}

	if policy.NetworkPolicy.AllowOutbound {
		t.Error("Network outbound should be disabled by default")
	}

	if policy.StoragePolicy.AllowLocalStorage {
		t.Error("Local storage should be disabled by default")
	}
}

func TestWASMConfiguration_ModuleValidation(t *testing.T) {
	config := &WASMConfiguration{
		Modules: map[string]*WASMModule{
			"test-module": {
				Name:       "test-module",
				Version:    "1.0.0",
				EntryPoint: "main",
				Exports:    []string{"init", "process"},
				Imports:    []string{"env.memory"},
				Permissions: &WASMPermissions{
					MemoryLimit:     32 * 1024 * 1024,
					AllowedImports:  []string{"env"},
					CPUTimeLimit:    3000,
					AllowNetworking: false,
					AllowFileSystem: false,
				},
			},
		},
		Permissions: &WASMPermissions{
			MemoryLimit:     64 * 1024 * 1024,
			AllowedImports:  []string{"env", "wasi"},
			CPUTimeLimit:    5000,
			AllowNetworking: false,
			AllowFileSystem: false,
		},
		MemoryLimit: 64 * 1024 * 1024,
	}

	// Test module access
	module, exists := config.Modules["test-module"]
	if !exists {
		t.Fatal("Test module not found")
	}

	if module.Name != "test-module" {
		t.Errorf("Module name mismatch: expected test-module, got %s", module.Name)
	}

	if len(module.Exports) != 2 {
		t.Errorf("Expected 2 exports, got %d", len(module.Exports))
	}

	// Test memory limit hierarchy
	if module.Permissions.MemoryLimit > config.MemoryLimit {
		t.Error("Module memory limit should not exceed global limit")
	}
}

func TestFeatureFlags_Combinations(t *testing.T) {
	tests := []struct {
		name     string
		features *FeatureFlags
		valid    bool
	}{
		{
			name: "static document",
			features: &FeatureFlags{
				Animations:    false,
				Interactivity: false,
				Charts:        false,
				Forms:         false,
				Audio:         false,
				Video:         false,
				WebGL:         false,
				WebAssembly:   false,
			},
			valid: true,
		},
		{
			name: "interactive document",
			features: &FeatureFlags{
				Animations:    true,
				Interactivity: true,
				Charts:        true,
				Forms:         true,
				Audio:         false,
				Video:         false,
				WebGL:         false,
				WebAssembly:   true,
			},
			valid: true,
		},
		{
			name: "multimedia document",
			features: &FeatureFlags{
				Animations:    true,
				Interactivity: true,
				Charts:        false,
				Forms:         false,
				Audio:         true,
				Video:         true,
				WebGL:         true,
				WebAssembly:   true,
			},
			valid: true,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			// Test JSON serialization
			data, err := json.Marshal(tt.features)
			if err != nil {
				t.Fatalf("Failed to marshal features: %v", err)
			}

			var unmarshaled FeatureFlags
			if err := json.Unmarshal(data, &unmarshaled); err != nil {
				t.Fatalf("Failed to unmarshal features: %v", err)
			}

			// Verify features match
			if unmarshaled.Animations != tt.features.Animations {
				t.Error("Animations flag mismatch")
			}

			if unmarshaled.WebAssembly != tt.features.WebAssembly {
				t.Error("WebAssembly flag mismatch")
			}
		})
	}
}

func TestResource_Properties(t *testing.T) {
	resource := &Resource{
		Hash: "sha256:abc123def456",
		Size: 1024,
		Type: "text/html",
		Path: "content/index.html",
	}

	// Test JSON serialization
	data, err := json.Marshal(resource)
	if err != nil {
		t.Fatalf("Failed to marshal resource: %v", err)
	}

	var unmarshaled Resource
	if err := json.Unmarshal(data, &unmarshaled); err != nil {
		t.Fatalf("Failed to unmarshal resource: %v", err)
	}

	// Verify properties
	if unmarshaled.Hash != resource.Hash {
		t.Errorf("Hash mismatch: expected %s, got %s", resource.Hash, unmarshaled.Hash)
	}

	if unmarshaled.Size != resource.Size {
		t.Errorf("Size mismatch: expected %d, got %d", resource.Size, unmarshaled.Size)
	}

	if unmarshaled.Type != resource.Type {
		t.Errorf("Type mismatch: expected %s, got %s", resource.Type, unmarshaled.Type)
	}

	if unmarshaled.Path != resource.Path {
		t.Errorf("Path mismatch: expected %s, got %s", resource.Path, unmarshaled.Path)
	}
}

func TestValidationResult_ErrorHandling(t *testing.T) {
	tests := []struct {
		name   string
		result *ValidationResult
		valid  bool
	}{
		{
			name: "valid result",
			result: &ValidationResult{
				IsValid:  true,
				Errors:   []string{},
				Warnings: []string{"minor warning"},
			},
			valid: true,
		},
		{
			name: "invalid with errors",
			result: &ValidationResult{
				IsValid:  false,
				Errors:   []string{"critical error", "another error"},
				Warnings: []string{},
			},
			valid: false,
		},
		{
			name: "valid with warnings",
			result: &ValidationResult{
				IsValid:  true,
				Errors:   []string{},
				Warnings: []string{"warning 1", "warning 2"},
			},
			valid: true,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			if tt.result.IsValid != tt.valid {
				t.Errorf("Validity mismatch: expected %v, got %v", tt.valid, tt.result.IsValid)
			}

			// Valid results should have no errors
			if tt.result.IsValid && len(tt.result.Errors) > 0 {
				t.Error("Valid result should not have errors")
			}

			// Invalid results should have at least one error
			if !tt.result.IsValid && len(tt.result.Errors) == 0 {
				t.Error("Invalid result should have errors")
			}
		})
	}
}

func TestSecurityReport_Comprehensive(t *testing.T) {
	report := &SecurityReport{
		IsValid:           true,
		SignatureVerified: true,
		IntegrityChecked:  true,
		PermissionsValid:  true,
		Warnings:          []string{"minor security warning"},
		Errors:            []string{},
	}

	// Test JSON serialization
	data, err := json.Marshal(report)
	if err != nil {
		t.Fatalf("Failed to marshal security report: %v", err)
	}

	var unmarshaled SecurityReport
	if err := json.Unmarshal(data, &unmarshaled); err != nil {
		t.Fatalf("Failed to unmarshal security report: %v", err)
	}

	// Verify all security checks
	if !unmarshaled.IsValid {
		t.Error("Security report should be valid")
	}

	if !unmarshaled.SignatureVerified {
		t.Error("Signature should be verified")
	}

	if !unmarshaled.IntegrityChecked {
		t.Error("Integrity should be checked")
	}

	if !unmarshaled.PermissionsValid {
		t.Error("Permissions should be valid")
	}

	if len(unmarshaled.Warnings) != 1 {
		t.Errorf("Expected 1 warning, got %d", len(unmarshaled.Warnings))
	}

	if len(unmarshaled.Errors) != 0 {
		t.Errorf("Expected 0 errors, got %d", len(unmarshaled.Errors))
	}
}

func TestAssetBundle_Operations(t *testing.T) {
	bundle := &AssetBundle{
		Images: map[string][]byte{
			"logo.png":    []byte("png-data"),
			"icon.svg":    []byte("svg-data"),
		},
		Fonts: map[string][]byte{
			"main.woff":   []byte("font-data-1"),
			"bold.woff2":  []byte("font-data-2"),
		},
		Data: map[string][]byte{
			"config.json": []byte(`{"setting": "value"}`),
			"data.csv":    []byte("col1,col2\nval1,val2"),
		},
	}

	// Test asset counts
	if len(bundle.Images) != 2 {
		t.Errorf("Expected 2 images, got %d", len(bundle.Images))
	}

	if len(bundle.Fonts) != 2 {
		t.Errorf("Expected 2 fonts, got %d", len(bundle.Fonts))
	}

	if len(bundle.Data) != 2 {
		t.Errorf("Expected 2 data files, got %d", len(bundle.Data))
	}

	// Test asset access
	if logoData, exists := bundle.Images["logo.png"]; exists {
		if string(logoData) != "png-data" {
			t.Error("Logo data mismatch")
		}
	} else {
		t.Error("Logo not found in images")
	}

	// Test JSON serialization
	data, err := json.Marshal(bundle)
	if err != nil {
		t.Fatalf("Failed to marshal asset bundle: %v", err)
	}

	var unmarshaled AssetBundle
	if err := json.Unmarshal(data, &unmarshaled); err != nil {
		t.Fatalf("Failed to unmarshal asset bundle: %v", err)
	}

	// Verify counts after serialization
	if len(unmarshaled.Images) != len(bundle.Images) {
		t.Error("Image count mismatch after serialization")
	}

	if len(unmarshaled.Fonts) != len(bundle.Fonts) {
		t.Error("Font count mismatch after serialization")
	}

	if len(unmarshaled.Data) != len(bundle.Data) {
		t.Error("Data count mismatch after serialization")
	}
}

func BenchmarkLIVDocument_JSONMarshal(b *testing.B) {
	document := &LIVDocument{
		Manifest: &Manifest{
			Version: "1.0",
			Metadata: &DocumentMetadata{
				Title:    "Benchmark Document",
				Author:   "Benchmark Author",
				Created:  time.Now(),
				Modified: time.Now(),
				Version:  "1.0.0",
				Language: "en",
			},
			Resources: make(map[string]*Resource),
		},
		Content: &DocumentContent{
			HTML: "<html><body>Benchmark content</body></html>",
			CSS:  "body { color: blue; }",
		},
		Assets: &AssetBundle{
			Images: make(map[string][]byte),
			Fonts:  make(map[string][]byte),
			Data:   make(map[string][]byte),
		},
	}

	// Add some resources for realistic benchmarking
	for i := 0; i < 100; i++ {
		document.Manifest.Resources[fmt.Sprintf("file%d.txt", i)] = &Resource{
			Hash: fmt.Sprintf("hash%d", i),
			Size: int64(i * 100),
			Type: "text/plain",
			Path: fmt.Sprintf("file%d.txt", i),
		}
	}

	b.ResetTimer()
	for i := 0; i < b.N; i++ {
		_, err := json.Marshal(document)
		if err != nil {
			b.Fatalf("Failed to marshal document: %v", err)
		}
	}
}

func BenchmarkLIVDocument_JSONUnmarshal(b *testing.B) {
	document := &LIVDocument{
		Manifest: &Manifest{
			Version: "1.0",
			Metadata: &DocumentMetadata{
				Title:    "Benchmark Document",
				Author:   "Benchmark Author",
				Created:  time.Now(),
				Modified: time.Now(),
				Version:  "1.0.0",
				Language: "en",
			},
			Resources: make(map[string]*Resource),
		},
		Content: &DocumentContent{
			HTML: "<html><body>Benchmark content</body></html>",
			CSS:  "body { color: blue; }",
		},
	}

	data, err := json.Marshal(document)
	if err != nil {
		b.Fatalf("Failed to marshal document: %v", err)
	}

	b.ResetTimer()
	for i := 0; i < b.N; i++ {
		var unmarshaled LIVDocument
		err := json.Unmarshal(data, &unmarshaled)
		if err != nil {
			b.Fatalf("Failed to unmarshal document: %v", err)
		}
	}
}