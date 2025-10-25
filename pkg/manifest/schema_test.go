package manifest

import (
	"encoding/json"
	"testing"
	"time"

	"github.com/liv-format/liv/pkg/core"
)

func TestManifestValidator_ValidateManifest(t *testing.T) {
	validator := NewManifestValidator()

	tests := []struct {
		name      string
		manifest  *core.Manifest
		wantValid bool
		wantError string
	}{
		{
			name: "valid manifest",
			manifest: &core.Manifest{
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
			},
			wantValid: true,
		},
		{
			name:      "nil manifest",
			manifest:  nil,
			wantValid: false,
			wantError: "manifest cannot be nil",
		},
		{
			name: "missing required fields",
			manifest: &core.Manifest{
				Version: "1.0",
				// Missing metadata, security, resources
			},
			wantValid: false,
		},
		{
			name: "invalid version format",
			manifest: &core.Manifest{
				Version: "1.0",
				Metadata: &core.DocumentMetadata{
					Title:       "Test Document",
					Author:      "Test Author",
					Created:     time.Now(),
					Modified:    time.Now(),
					Description: "A test document",
					Version:     "invalid-version", // Invalid semver
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
			},
			wantValid: false,
		},
		{
			name: "created after modified date",
			manifest: &core.Manifest{
				Version: "1.0",
				Metadata: &core.DocumentMetadata{
					Title:       "Test Document",
					Author:      "Test Author",
					Created:     time.Now(),
					Modified:    time.Now().Add(-time.Hour), // Modified before created
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
					NetworkPolicy: &core.NetworkPolicy{},
					StoragePolicy: &core.StoragePolicy{},
				},
				Resources: map[string]*core.Resource{},
			},
			wantValid: false,
			wantError: "created date cannot be after modified date",
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
					if err == tt.wantError {
						found = true
						break
					}
				}
				if !found {
					t.Errorf("Expected error '%s' not found in: %v", tt.wantError, result.Errors)
				}
			}
		})
	}
}

func TestManifestValidator_ValidateManifestJSON(t *testing.T) {
	validator := NewManifestValidator()

	tests := []struct {
		name      string
		json      string
		wantValid bool
	}{
		{
			name: "valid JSON manifest",
			json: `{
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
						"size": 1024,
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
			}`,
			wantValid: true,
		},
		{
			name:      "invalid JSON",
			json:      `{"version": "1.0", "invalid": }`,
			wantValid: false,
		},
		{
			name: "missing required fields",
			json: `{
				"version": "1.0"
			}`,
			wantValid: false,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			_, result := validator.ValidateManifestJSON([]byte(tt.json))

			if result.IsValid != tt.wantValid {
				t.Errorf("ValidateManifestJSON() isValid = %v, want %v", result.IsValid, tt.wantValid)
				if len(result.Errors) > 0 {
					t.Errorf("Errors: %v", result.Errors)
				}
			}
		})
	}
}

func TestCustomValidationFunctions(t *testing.T) {
	tests := []struct {
		name     string
		value    string
		function func(string) bool
		want     bool
	}{
		{
			name:     "valid semver",
			value:    "1.0.0",
			function: func(v string) bool { mv := &ManifestValidator{}; return mv.isValidSemVer(v) },
			want:     true,
		},
		{
			name:     "invalid semver",
			value:    "1.0",
			function: func(v string) bool { mv := &ManifestValidator{}; return mv.isValidSemVer(v) },
			want:     false,
		},
		{
			name:     "valid MIME type",
			value:    "text/html",
			function: func(v string) bool { mv := &ManifestValidator{}; return mv.isValidMimeType(v) },
			want:     true,
		},
		{
			name:     "invalid MIME type",
			value:    "invalid-mime",
			function: func(v string) bool { mv := &ManifestValidator{}; return mv.isValidMimeType(v) },
			want:     false,
		},
		{
			name:     "valid CSP",
			value:    "default-src 'self'; script-src 'self'",
			function: func(v string) bool { mv := &ManifestValidator{}; return mv.isValidCSP(v) },
			want:     true,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			got := tt.function(tt.value)
			if got != tt.want {
				t.Errorf("validation function = %v, want %v for value %s", got, tt.want, tt.value)
			}
		})
	}
}

func TestWASMModuleCircularDependency(t *testing.T) {
	validator := NewManifestValidator()

	modules := map[string]*core.WASMModule{
		"module-a": {
			Name:    "module-a",
			Version: "1.0.0",
			Imports: []string{"module-b"},
		},
		"module-b": {
			Name:    "module-b",
			Version: "1.0.0",
			Imports: []string{"module-c"},
		},
		"module-c": {
			Name:    "module-c",
			Version: "1.0.0",
			Imports: []string{"module-a"}, // Circular dependency
		},
	}

	hasCircular := validator.hasCircularDependency("module-a", modules["module-a"], modules)
	if !hasCircular {
		t.Error("Expected circular dependency to be detected")
	}

	// Test without circular dependency
	modules["module-c"].Imports = []string{} // Remove circular dependency
	hasCircular = validator.hasCircularDependency("module-a", modules["module-a"], modules)
	if hasCircular {
		t.Error("Expected no circular dependency")
	}
}

func TestManifestBuilder_Integration(t *testing.T) {
	// Test the integration between builder and validator
	builder := NewManifestBuilder()

	// Build a complete manifest
	builder.CreateDefaultMetadata("Test Document", "Test Author").
		CreateDefaultSecurityPolicy().
		CreateDefaultFeatureFlags()

	// Add a resource
	builder.AddResource("content/index.html", &core.Resource{
		Hash: "e3b0c44298fc1c149afbf4c8996fb92427ae41e4649b934ca495991b7852b855",
		Size: 1024,
		Type: "text/html",
		Path: "content/index.html",
	})

	// Build and validate
	manifest, err := builder.Build()
	if err != nil {
		t.Fatalf("Failed to build manifest: %v", err)
	}

	// Validate the built manifest
	validator := NewManifestValidator()
	result := validator.ValidateManifest(manifest)

	if !result.IsValid {
		t.Errorf("Built manifest is invalid: %v", result.Errors)
	}

	// Test JSON serialization
	data, err := json.MarshalIndent(manifest, "", "  ")
	if err != nil {
		t.Fatalf("Failed to marshal manifest: %v", err)
	}

	// Test JSON parsing
	parsedManifest, parseResult := validator.ValidateManifestJSON(data)
	if !parseResult.IsValid {
		t.Errorf("Parsed manifest is invalid: %v", parseResult.Errors)
	}

	if parsedManifest.Metadata.Title != "Test Document" {
		t.Errorf("Parsed manifest title = %s, want %s", parsedManifest.Metadata.Title, "Test Document")
	}
}