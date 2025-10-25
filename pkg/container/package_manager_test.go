package container

import (
	"bytes"
	"context"
	"io"
	"strings"
	"testing"
	"time"

	"github.com/liv-format/liv/pkg/core"
)

func TestPackageManagerImpl_CreatePackage(t *testing.T) {
	pm := NewPackageManager()

	// Create test manifest
	manifest := &core.Manifest{
		Version: "1.0",
		Metadata: &core.DocumentMetadata{
			Title:       "Test Document",
			Author:      "Test Author",
			Created:     time.Now().Add(-time.Hour),
			Modified:    time.Now(),
			Description: "A test document",
			Version:     "1.0.0",
			Language:    "en",
		},
		Security: &core.SecurityPolicy{
			WASMPermissions: &core.WASMPermissions{
				MemoryLimit:     64 * 1024 * 1024,
				AllowedImports:  []string{},
				CPUTimeLimit:    5000,
				AllowNetworking: false,
				AllowFileSystem: false,
			},
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
			ContentSecurityPolicy: "default-src 'self'",
			TrustedDomains:        []string{},
		},
		Resources: map[string]*core.Resource{
			"content/index.html": {
				Hash: "e3b0c44298fc1c149afbf4c8996fb92427ae41e4649b934ca495991b7852b855",
				Size: 1024,
				Type: "text/html",
				Path: "content/index.html",
			},
		},
		Features: &core.FeatureFlags{
			Animations:    true,
			Interactivity: true,
			Charts:        false,
			Forms:         false,
			Audio:         false,
			Video:         false,
			WebGL:         false,
			WebAssembly:   false,
		},
	}

	// Create test sources
	sources := map[string]io.Reader{
		"content/index.html":         strings.NewReader("<html><body><h1>Test</h1></body></html>"),
		"content/styles/main.css":    strings.NewReader("body { font-family: Arial; }"),
		"assets/images/logo.png":     strings.NewReader("fake-png-data"),
		"assets/data/sample.json":    strings.NewReader(`{"test": true}`),
		"signatures/content.sig":     strings.NewReader("fake-signature"),
		"signatures/manifest.sig":    strings.NewReader("fake-manifest-signature"),
	}

	// Create package
	ctx := context.Background()
	document, err := pm.CreatePackage(ctx, sources, manifest)
	if err != nil {
		t.Fatalf("Failed to create package: %v", err)
	}

	// Verify document structure
	if document.Manifest != manifest {
		t.Error("Manifest not set correctly")
	}

	if document.Content == nil {
		t.Fatal("Content is nil")
	}

	if document.Content.HTML == "" {
		t.Error("HTML content not extracted")
	}

	if document.Content.CSS == "" {
		t.Error("CSS content not extracted")
	}

	if document.Assets == nil {
		t.Fatal("Assets is nil")
	}

	if len(document.Assets.Images) == 0 {
		t.Error("Images not extracted")
	}

	if len(document.Assets.Data) == 0 {
		t.Error("Data files not extracted")
	}

	if document.Signatures == nil {
		t.Fatal("Signatures is nil")
	}

	if document.Signatures.ContentSignature == "" {
		t.Error("Content signature not extracted")
	}
}

func TestPackageManagerImpl_ExtractPackage(t *testing.T) {
	pm := NewPackageManager()

	// Create test ZIP data
	testFiles := map[string][]byte{
		"manifest.json": []byte(`{
			"version": "1.0",
			"metadata": {
				"title": "Test Document",
				"author": "Test Author",
				"created": "2024-01-01T00:00:00Z",
				"modified": "2024-01-01T01:00:00Z",
				"description": "A test document",
				"version": "1.0.0",
				"language": "en"
			},
			"security": {
				"wasm_permissions": {
					"memory_limit": 67108864,
					"allowed_imports": [],
					"cpu_time_limit": 5000,
					"allow_networking": false,
					"allow_file_system": false
				},
				"js_permissions": {
					"execution_mode": "sandboxed",
					"allowed_apis": [],
					"dom_access": "read"
				},
				"network_policy": {
					"allow_outbound": false,
					"allowed_hosts": [],
					"allowed_ports": []
				},
				"storage_policy": {
					"allow_local_storage": false,
					"allow_session_storage": false,
					"allow_indexed_db": false,
					"allow_cookies": false
				},
				"content_security_policy": "default-src 'self'",
				"trusted_domains": []
			},
			"resources": {
				"content/index.html": {
					"hash": "e3b0c44298fc1c149afbf4c8996fb92427ae41e4649b934ca495991b7852b855",
					"size": 42,
					"type": "text/html",
					"path": "content/index.html"
				}
			},
			"features": {
				"animations": true,
				"interactivity": true,
				"charts": false,
				"forms": false,
				"audio": false,
				"video": false,
				"webgl": false,
				"webassembly": false
			}
		}`),
		"content/index.html":      []byte("<html><body><h1>Test</h1></body></html>"),
		"content/styles/main.css": []byte("body { font-family: Arial; }"),
		"assets/images/logo.png":  []byte("fake-png-data"),
		"assets/data/sample.json": []byte(`{"test": true}`),
		"wasm/test-module.wasm":   []byte{0x00, 0x61, 0x73, 0x6D, 0x01, 0x00, 0x00, 0x00}, // Valid WASM header
		"signatures/content.sig":  []byte("fake-signature"),
	}

	// Create ZIP in memory
	zipContainer := NewZIPContainer()
	var buf bytes.Buffer
	err := zipContainer.CreateFromFilesToWriter(testFiles, &buf)
	if err != nil {
		t.Fatalf("Failed to create test ZIP: %v", err)
	}

	// Extract package
	ctx := context.Background()
	document, err := pm.ExtractPackage(ctx, &buf)
	if err != nil {
		t.Fatalf("Failed to extract package: %v", err)
	}

	// Verify extracted document
	if document.Manifest == nil {
		t.Fatal("Manifest is nil")
	}

	if document.Manifest.Metadata.Title != "Test Document" {
		t.Errorf("Expected title 'Test Document', got '%s'", document.Manifest.Metadata.Title)
	}

	if document.Content == nil {
		t.Fatal("Content is nil")
	}

	if !strings.Contains(document.Content.HTML, "<h1>Test</h1>") {
		t.Error("HTML content not extracted correctly")
	}

	if document.Assets == nil {
		t.Fatal("Assets is nil")
	}

	if len(document.Assets.Images) == 0 {
		t.Error("Images not extracted")
	}

	if len(document.WASMModules) == 0 {
		t.Error("WASM modules not extracted")
	}

	if document.Signatures == nil {
		t.Fatal("Signatures is nil")
	}

	if document.Signatures.ContentSignature == "" {
		t.Error("Content signature not extracted")
	}
}

func TestPackageManagerImpl_ValidateStructure(t *testing.T) {
	pm := NewPackageManager()

	tests := []struct {
		name      string
		document  *core.LIVDocument
		wantValid bool
		wantError string
	}{
		{
			name: "valid document",
			document: &core.LIVDocument{
				Manifest: &core.Manifest{
					Version: "1.0",
					Metadata: &core.DocumentMetadata{
						Title:       "Test",
						Author:      "Author",
						Created:     time.Now().Add(-time.Hour),
						Modified:    time.Now(),
						Description: "Test",
						Version:     "1.0.0",
						Language:    "en",
					},
					Security: &core.SecurityPolicy{
						WASMPermissions: &core.WASMPermissions{
							MemoryLimit:     64 * 1024 * 1024,
							AllowedImports:  []string{},
							CPUTimeLimit:    5000,
							AllowNetworking: false,
							AllowFileSystem: false,
						},
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
						ContentSecurityPolicy: "default-src 'self'",
						TrustedDomains:        []string{},
					},
					Resources: map[string]*core.Resource{
						"content/index.html": {
							Hash: "e3b0c44298fc1c149afbf4c8996fb92427ae41e4649b934ca495991b7852b855",
							Size: 1024,
							Type: "text/html",
							Path: "content/index.html",
						},
					},
					Features:  &core.FeatureFlags{},
				},
				Content: &core.DocumentContent{
					HTML: "<html></html>",
				},
				Assets:      &core.AssetBundle{},
				Signatures:  &core.SignatureBundle{},
				WASMModules: make(map[string][]byte),
			},
			wantValid: true,
		},
		{
			name: "missing content",
			document: &core.LIVDocument{
				Manifest: &core.Manifest{
					Version: "1.0",
					Metadata: &core.DocumentMetadata{
						Title:       "Test",
						Author:      "Author",
						Created:     time.Now().Add(-time.Hour),
						Modified:    time.Now(),
						Description: "Test",
						Version:     "1.0.0",
						Language:    "en",
					},
					Security: &core.SecurityPolicy{
						WASMPermissions: &core.WASMPermissions{
							MemoryLimit:     64 * 1024 * 1024,
							AllowedImports:  []string{},
							CPUTimeLimit:    5000,
							AllowNetworking: false,
							AllowFileSystem: false,
						},
						JSPermissions: &core.JSPermissions{
							ExecutionMode: "sandboxed",
							AllowedAPIs:   []string{},
							DOMAccess:     "read",
						},
						NetworkPolicy: &core.NetworkPolicy{},
						StoragePolicy: &core.StoragePolicy{},
					},
					Resources: map[string]*core.Resource{},
					Features:  &core.FeatureFlags{},
				},
				Content:     nil, // Missing content
				Assets:      &core.AssetBundle{},
				Signatures:  &core.SignatureBundle{},
				WASMModules: make(map[string][]byte),
			},
			wantValid: false,
			wantError: "document content is missing",
		},
		{
			name: "WASM module mismatch",
			document: &core.LIVDocument{
				Manifest: &core.Manifest{
					Version: "1.0",
					Metadata: &core.DocumentMetadata{
						Title:       "Test",
						Author:      "Author",
						Created:     time.Now().Add(-time.Hour),
						Modified:    time.Now(),
						Description: "Test",
						Version:     "1.0.0",
						Language:    "en",
					},
					Security: &core.SecurityPolicy{
						WASMPermissions: &core.WASMPermissions{
							MemoryLimit:     64 * 1024 * 1024,
							AllowedImports:  []string{},
							CPUTimeLimit:    5000,
							AllowNetworking: false,
							AllowFileSystem: false,
						},
						JSPermissions: &core.JSPermissions{
							ExecutionMode: "sandboxed",
							AllowedAPIs:   []string{},
							DOMAccess:     "read",
						},
						NetworkPolicy: &core.NetworkPolicy{},
						StoragePolicy: &core.StoragePolicy{},
					},
					Resources: map[string]*core.Resource{},
					Features:  &core.FeatureFlags{},
					WASMConfig: &core.WASMConfiguration{
						Modules: map[string]*core.WASMModule{
							"missing-module": {
								Name:    "missing-module",
								Version: "1.0.0",
							},
						},
						Permissions: &core.WASMPermissions{
							MemoryLimit:     64 * 1024 * 1024,
							AllowedImports:  []string{},
							CPUTimeLimit:    5000,
							AllowNetworking: false,
							AllowFileSystem: false,
						},
						MemoryLimit: 64 * 1024 * 1024,
					},
				},
				Content: &core.DocumentContent{
					HTML: "<html></html>",
				},
				Assets:      &core.AssetBundle{},
				Signatures:  &core.SignatureBundle{},
				WASMModules: make(map[string][]byte), // Empty, but manifest references modules
			},
			wantValid: false,
			wantError: "WASM module 'missing-module' referenced in manifest but not found",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			result := pm.ValidateStructure(tt.document)

			if result.IsValid != tt.wantValid {
				t.Errorf("ValidateStructure() isValid = %v, want %v", result.IsValid, tt.wantValid)
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

func TestPackageManagerImpl_CompressAssets(t *testing.T) {
	pm := NewPackageManager()

	// Create test assets with duplicates
	originalAssets := &core.AssetBundle{
		Images: map[string][]byte{
			"logo.png":      []byte("png-data-1"),
			"duplicate.png": []byte("png-data-1"), // Same as logo.png
			"unique.png":    []byte("png-data-2"),
		},
		Fonts: map[string][]byte{
			"font1.woff": []byte("font-data-1"),
			"font2.woff": []byte("font-data-2"),
		},
		Data: map[string][]byte{
			"data1.json": []byte(`{"key": "value1"}`),
			"data2.json": []byte(`{"key": "value2"}`),
		},
	}

	compressedAssets, err := pm.CompressAssets(originalAssets)
	if err != nil {
		t.Fatalf("Failed to compress assets: %v", err)
	}

	// Should have deduplicated the duplicate image
	if len(compressedAssets.Images) != 2 {
		t.Errorf("Expected 2 unique images after deduplication, got %d", len(compressedAssets.Images))
	}

	// Fonts and data should remain the same
	if len(compressedAssets.Fonts) != len(originalAssets.Fonts) {
		t.Errorf("Font count changed: expected %d, got %d", len(originalAssets.Fonts), len(compressedAssets.Fonts))
	}

	if len(compressedAssets.Data) != len(originalAssets.Data) {
		t.Errorf("Data count changed: expected %d, got %d", len(originalAssets.Data), len(compressedAssets.Data))
	}
}

func TestPackageManagerImpl_LoadWASMModule(t *testing.T) {
	pm := NewPackageManager()

	tests := []struct {
		name      string
		data      []byte
		wantError bool
	}{
		{
			name:      "valid WASM module",
			data:      []byte{0x00, 0x61, 0x73, 0x6D, 0x01, 0x00, 0x00, 0x00}, // Valid WASM header
			wantError: false,
		},
		{
			name:      "too small",
			data:      []byte{0x00, 0x61},
			wantError: true,
		},
		{
			name:      "invalid magic number",
			data:      []byte{0xFF, 0xFF, 0xFF, 0xFF, 0x01, 0x00, 0x00, 0x00},
			wantError: true,
		},
		{
			name:      "invalid version",
			data:      []byte{0x00, 0x61, 0x73, 0x6D, 0xFF, 0xFF, 0xFF, 0xFF},
			wantError: true,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			module, err := pm.LoadWASMModule("test-module", tt.data)

			if tt.wantError {
				if err == nil {
					t.Error("Expected error but got none")
				}
			} else {
				if err != nil {
					t.Errorf("Unexpected error: %v", err)
				}
				if module == nil {
					t.Error("Expected module but got nil")
				}
				if module != nil && module.Name != "test-module" {
					t.Errorf("Expected module name 'test-module', got '%s'", module.Name)
				}
			}
		})
	}
}

func TestPackageManagerImpl_SaveAndLoadRoundTrip(t *testing.T) {
	pm := NewPackageManager()

	// Create a complete test document
	originalDocument := &core.LIVDocument{
		Manifest: &core.Manifest{
			Version: "1.0",
			Metadata: &core.DocumentMetadata{
				Title:       "Round Trip Test",
				Author:      "Test Author",
				Created:     time.Now().Add(-time.Hour),
				Modified:    time.Now(),
				Description: "Testing save and load",
				Version:     "1.0.0",
				Language:    "en",
			},
			Security: &core.SecurityPolicy{
				WASMPermissions: &core.WASMPermissions{
					MemoryLimit:     64 * 1024 * 1024,
					AllowedImports:  []string{},
					CPUTimeLimit:    5000,
					AllowNetworking: false,
					AllowFileSystem: false,
				},
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
				ContentSecurityPolicy: "default-src 'self'",
				TrustedDomains:        []string{},
			},
			Resources: map[string]*core.Resource{
				"content/index.html": {
					Hash: "e3b0c44298fc1c149afbf4c8996fb92427ae41e4649b934ca495991b7852b855",
					Size: 1024,
					Type: "text/html",
					Path: "content/index.html",
				},
			},
			Features: &core.FeatureFlags{
				Animations:    true,
				Interactivity: true,
				Charts:        false,
				Forms:         false,
				Audio:         false,
				Video:         false,
				WebGL:         false,
				WebAssembly:   false,
			},
		},
		Content: &core.DocumentContent{
			HTML:           "<html><body><h1>Test</h1></body></html>",
			CSS:            "body { font-family: Arial; }",
			InteractiveSpec: "console.log('Hello World');",
			StaticFallback: "<html><body><h1>Static Test</h1></body></html>",
		},
		Assets: &core.AssetBundle{
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
		Signatures: &core.SignatureBundle{
			ContentSignature:  "fake-content-signature",
			ManifestSignature: "fake-manifest-signature",
			WASMSignatures:    map[string]string{},
		},
		WASMModules: map[string][]byte{
			"test-module": {0x00, 0x61, 0x73, 0x6D, 0x01, 0x00, 0x00, 0x00},
		},
	}

	// Save to buffer
	var buf bytes.Buffer
	err := pm.SavePackageToWriter(originalDocument, &buf)
	if err != nil {
		t.Fatalf("Failed to save package: %v", err)
	}

	// Load from buffer
	ctx := context.Background()
	loadedDocument, err := pm.ExtractPackage(ctx, &buf)
	if err != nil {
		t.Fatalf("Failed to load package: %v", err)
	}

	// Compare documents
	if loadedDocument.Manifest.Metadata.Title != originalDocument.Manifest.Metadata.Title {
		t.Errorf("Title mismatch: expected '%s', got '%s'",
			originalDocument.Manifest.Metadata.Title,
			loadedDocument.Manifest.Metadata.Title)
	}

	if loadedDocument.Content.HTML != originalDocument.Content.HTML {
		t.Error("HTML content mismatch")
	}

	if loadedDocument.Content.CSS != originalDocument.Content.CSS {
		t.Error("CSS content mismatch")
	}

	if len(loadedDocument.Assets.Images) != len(originalDocument.Assets.Images) {
		t.Errorf("Image count mismatch: expected %d, got %d",
			len(originalDocument.Assets.Images),
			len(loadedDocument.Assets.Images))
	}

	if len(loadedDocument.WASMModules) != len(originalDocument.WASMModules) {
		t.Errorf("WASM module count mismatch: expected %d, got %d",
			len(originalDocument.WASMModules),
			len(loadedDocument.WASMModules))
	}
}

func BenchmarkPackageManagerImpl_CreatePackage(b *testing.B) {
	pm := NewPackageManager()

	// Create test manifest
	manifest := &core.Manifest{
		Version: "1.0",
		Metadata: &core.DocumentMetadata{
			Title:       "Benchmark Test",
			Author:      "Test Author",
			Created:     time.Now().Add(-time.Hour),
			Modified:    time.Now(),
			Description: "Benchmark test document",
			Version:     "1.0.0",
			Language:    "en",
		},
		Security: &core.SecurityPolicy{
			WASMPermissions: &core.WASMPermissions{
				MemoryLimit:     64 * 1024 * 1024,
				AllowedImports:  []string{},
				CPUTimeLimit:    5000,
				AllowNetworking: false,
				AllowFileSystem: false,
			},
			JSPermissions: &core.JSPermissions{
				ExecutionMode: "sandboxed",
				AllowedAPIs:   []string{},
				DOMAccess:     "read",
			},
			NetworkPolicy: &core.NetworkPolicy{},
			StoragePolicy: &core.StoragePolicy{},
		},
		Resources: map[string]*core.Resource{},
		Features:  &core.FeatureFlags{},
	}

	// Create test sources
	sources := map[string]io.Reader{
		"content/index.html":      strings.NewReader(strings.Repeat("<p>Content</p>", 1000)),
		"content/styles/main.css": strings.NewReader(strings.Repeat("body { color: red; }", 100)),
		"assets/data/large.json":  strings.NewReader(strings.Repeat(`{"key": "value"}`, 1000)),
	}

	ctx := context.Background()
	b.ResetTimer()

	for i := 0; i < b.N; i++ {
		// Reset readers
		for path, reader := range sources {
			if seeker, ok := reader.(io.Seeker); ok {
				seeker.Seek(0, io.SeekStart)
			} else {
				// Recreate reader for benchmark
				switch path {
				case "content/index.html":
					sources[path] = strings.NewReader(strings.Repeat("<p>Content</p>", 1000))
				case "content/styles/main.css":
					sources[path] = strings.NewReader(strings.Repeat("body { color: red; }", 100))
				case "assets/data/large.json":
					sources[path] = strings.NewReader(strings.Repeat(`{"key": "value"}`, 1000))
				}
			}
		}

		_, err := pm.CreatePackage(ctx, sources, manifest)
		if err != nil {
			b.Fatalf("Failed to create package: %v", err)
		}
	}
}