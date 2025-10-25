package core

import (
	"encoding/json"
	"testing"
	"time"
)

// TestLIVDocumentIntegration tests the complete integration of all core types
func TestLIVDocumentIntegration(t *testing.T) {
	// Create a comprehensive LIV document that exercises all components
	document := createCompleteTestDocument()

	// Test 1: Document structure validation
	if document.Manifest == nil {
		t.Fatal("Manifest is nil")
	}

	if document.Content == nil {
		t.Fatal("Content is nil")
	}

	if document.Assets == nil {
		t.Fatal("Assets is nil")
	}

	if document.Signatures == nil {
		t.Fatal("Signatures is nil")
	}

	if document.WASMModules == nil {
		t.Fatal("WASMModules is nil")
	}

	// Test 2: Manifest completeness
	validateManifestCompleteness(t, document.Manifest)

	// Test 3: Content integrity
	validateContentIntegrity(t, document.Content)

	// Test 4: Asset bundle validation
	validateAssetBundle(t, document.Assets)

	// Test 5: WASM module validation
	validateWASMModules(t, document.WASMModules, document.Manifest.WASMConfig)

	// Test 6: Signature bundle validation
	validateSignatureBundle(t, document.Signatures)

	// Test 7: Cross-component consistency
	validateCrossComponentConsistency(t, document)

	// Test 8: JSON round-trip integrity
	validateJSONRoundTrip(t, document)

	// Test 9: Security policy enforcement
	validateSecurityPolicyEnforcement(t, document.Manifest.Security)

	// Test 10: Feature flag consistency
	validateFeatureFlagConsistency(t, document.Manifest.Features, document.Manifest.WASMConfig)
}

func createCompleteTestDocument() *LIVDocument {
	now := time.Now()
	
	return &LIVDocument{
		Manifest: &Manifest{
			Version: "1.0",
			Metadata: &DocumentMetadata{
				Title:       "Complete Integration Test Document",
				Author:      "Integration Test Suite",
				Created:     now.Add(-2 * time.Hour),
				Modified:    now,
				Description: "A comprehensive test document that exercises all LIV format features and components for integration testing.",
				Version:     "2.1.0",
				Language:    "en",
			},
			Security: &SecurityPolicy{
				WASMPermissions: &WASMPermissions{
					MemoryLimit:     128 * 1024 * 1024, // 128MB
					AllowedImports:  []string{"env", "wasi_snapshot_preview1"},
					CPUTimeLimit:    15000, // 15 seconds
					AllowNetworking: false,
					AllowFileSystem: false,
				},
				JSPermissions: &JSPermissions{
					ExecutionMode: "sandboxed",
					AllowedAPIs:   []string{"canvas", "webgl", "audio"},
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
				ContentSecurityPolicy: "default-src 'self'; script-src 'self' 'unsafe-inline' 'wasm-unsafe-eval'; style-src 'self' 'unsafe-inline'; img-src 'self' data:;",
				TrustedDomains:        []string{},
			},
			Resources: map[string]*Resource{
				"manifest.json": {
					Hash: "manifest-hash-abc123",
					Size: 2048,
					Type: "application/json",
					Path: "manifest.json",
				},
				"content/index.html": {
					Hash: "html-hash-def456",
					Size: 4096,
					Type: "text/html",
					Path: "content/index.html",
				},
				"content/styles/main.css": {
					Hash: "css-hash-ghi789",
					Size: 2048,
					Type: "text/css",
					Path: "content/styles/main.css",
				},
				"content/scripts/main.js": {
					Hash: "js-hash-jkl012",
					Size: 8192,
					Type: "application/javascript",
					Path: "content/scripts/main.js",
				},
				"assets/images/logo.svg": {
					Hash: "svg-hash-mno345",
					Size: 1024,
					Type: "image/svg+xml",
					Path: "assets/images/logo.svg",
				},
				"assets/fonts/main.woff2": {
					Hash: "font-hash-pqr678",
					Size: 32768,
					Type: "font/woff2",
					Path: "assets/fonts/main.woff2",
				},
				"wasm/interactive.wasm": {
					Hash: "wasm-hash-stu901",
					Size: 65536,
					Type: "application/wasm",
					Path: "wasm/interactive.wasm",
				},
			},
			WASMConfig: &WASMConfiguration{
				Modules: map[string]*WASMModule{
					"interactive-engine": {
						Name:       "interactive-engine",
						Version:    "1.2.0",
						EntryPoint: "init_interactive_engine",
						Exports:    []string{"init_interactive_engine", "process_interaction", "render_frame", "cleanup"},
						Imports:    []string{"env.memory", "env.console_log", "env.performance_now"},
						Permissions: &WASMPermissions{
							MemoryLimit:     64 * 1024 * 1024, // 64MB
							AllowedImports:  []string{"env"},
							CPUTimeLimit:    10000, // 10 seconds
							AllowNetworking: false,
							AllowFileSystem: false,
						},
						Metadata: map[string]string{
							"description": "Interactive content rendering engine",
							"author":      "LIV Format Team",
							"license":     "MIT",
						},
					},
				},
				Permissions: &WASMPermissions{
					MemoryLimit:     128 * 1024 * 1024,
					AllowedImports:  []string{"env", "wasi_snapshot_preview1"},
					CPUTimeLimit:    15000,
					AllowNetworking: false,
					AllowFileSystem: false,
				},
				MemoryLimit: 128 * 1024 * 1024,
			},
			Features: &FeatureFlags{
				Animations:    true,
				Interactivity: true,
				Charts:        true,
				Forms:         true,
				Audio:         true,
				Video:         false,
				WebGL:         true,
				WebAssembly:   true,
			},
		},
		Content: &DocumentContent{
			HTML: `<!DOCTYPE html>
<html lang="en">
<head>
    <meta charset="UTF-8">
    <title>Integration Test Document</title>
    <link rel="stylesheet" href="styles/main.css">
</head>
<body>
    <h1>LIV Integration Test</h1>
    <div id="interactive-content"></div>
    <script src="scripts/main.js"></script>
</body>
</html>`,
			CSS: `body {
    font-family: Arial, sans-serif;
    margin: 0;
    padding: 20px;
    background: linear-gradient(135deg, #667eea 0%, #764ba2 100%);
}

h1 {
    color: white;
    text-align: center;
}

#interactive-content {
    background: white;
    padding: 20px;
    border-radius: 8px;
    margin: 20px 0;
}`,
			InteractiveSpec: `// Interactive content specification
const interactiveConfig = {
    engine: "interactive-engine",
    version: "1.2.0",
    features: ["animations", "charts", "webgl"],
    initialization: {
        memoryLimit: 64 * 1024 * 1024,
        timeLimit: 10000
    }
};

// Initialize interactive content
document.addEventListener('DOMContentLoaded', function() {
    console.log('Initializing LIV interactive content');
    // WASM module would be loaded here
});`,
			StaticFallback: `<!DOCTYPE html>
<html lang="en">
<head>
    <meta charset="UTF-8">
    <title>Integration Test Document - Static Fallback</title>
</head>
<body>
    <h1>LIV Integration Test (Static Mode)</h1>
    <p>This is the static fallback version of the document.</p>
    <p>Interactive features are not available in this mode.</p>
</body>
</html>`,
		},
		Assets: &AssetBundle{
			Images: map[string][]byte{
				"logo.svg":        []byte(`<svg xmlns="http://www.w3.org/2000/svg" viewBox="0 0 100 100"><circle cx="50" cy="50" r="40" fill="#007bff"/></svg>`),
				"background.png":  []byte("fake-png-background-data"),
				"icon-chart.svg":  []byte(`<svg xmlns="http://www.w3.org/2000/svg" viewBox="0 0 24 24"><path d="M3 13h8V3H3v10zm0 8h8v-6H3v6zm10 0h8V11h-8v10zm0-18v6h8V3h-8z"/></svg>`),
			},
			Fonts: map[string][]byte{
				"main.woff2":      []byte("fake-woff2-main-font-data"),
				"bold.woff2":      []byte("fake-woff2-bold-font-data"),
				"icons.woff":      []byte("fake-woff-icon-font-data"),
			},
			Data: map[string][]byte{
				"chart-data.json": []byte(`{
					"datasets": [
						{"label": "Performance", "data": [85, 92, 78, 96, 88]},
						{"label": "Security", "data": [95, 89, 92, 94, 97]}
					],
					"labels": ["Q1", "Q2", "Q3", "Q4", "Q5"]
				}`),
				"config.json":     []byte(`{
					"theme": "default",
					"animations": true,
					"performance": {
						"maxFPS": 60,
						"memoryLimit": "64MB"
					}
				}`),
				"localization.csv": []byte("key,en,es,fr\nwelcome,Welcome,Bienvenido,Bienvenue\ngoodbye,Goodbye,Adiós,Au revoir"),
			},
		},
		Signatures: &SignatureBundle{
			ContentSignature:  "content-signature-xyz789abc123def456ghi789jkl012mno345pqr678stu901vwx234",
			ManifestSignature: "manifest-signature-abc123def456ghi789jkl012mno345pqr678stu901vwx234yz567",
			WASMSignatures: map[string]string{
				"interactive-engine": "wasm-signature-def456ghi789jkl012mno345pqr678stu901vwx234yz567abc123",
			},
		},
		WASMModules: map[string][]byte{
			"interactive-engine": {
				// Valid WASM magic number and version
				0x00, 0x61, 0x73, 0x6D, // Magic number
				0x01, 0x00, 0x00, 0x00, // Version
				// Minimal WASM module content
				0x01, 0x04, 0x01, 0x60, 0x00, 0x00, // Type section
				0x03, 0x02, 0x01, 0x00,             // Function section
				0x0A, 0x04, 0x01, 0x02, 0x00, 0x0B, // Code section
			},
		},
	}
}

func validateManifestCompleteness(t *testing.T, manifest *Manifest) {
	if manifest.Version == "" {
		t.Error("Manifest version is empty")
	}

	if manifest.Metadata == nil {
		t.Fatal("Manifest metadata is nil")
	}

	if manifest.Metadata.Title == "" {
		t.Error("Document title is empty")
	}

	if manifest.Metadata.Author == "" {
		t.Error("Document author is empty")
	}

	if manifest.Security == nil {
		t.Fatal("Security policy is nil")
	}

	if len(manifest.Resources) == 0 {
		t.Error("No resources defined in manifest")
	}

	// Validate required resources
	requiredResources := []string{"manifest.json", "content/index.html"}
	for _, required := range requiredResources {
		if _, exists := manifest.Resources[required]; !exists {
			t.Errorf("Required resource missing: %s", required)
		}
	}
}

func validateContentIntegrity(t *testing.T, content *DocumentContent) {
	if content.HTML == "" {
		t.Error("HTML content is empty")
	}

	if content.CSS == "" {
		t.Error("CSS content is empty")
	}

	if content.StaticFallback == "" {
		t.Error("Static fallback content is empty")
	}

	// Validate HTML structure
	if !contains(content.HTML, "<!DOCTYPE html>") {
		t.Error("HTML missing DOCTYPE declaration")
	}

	if !contains(content.HTML, "<html") {
		t.Error("HTML missing html tag")
	}

	// Validate CSS structure
	if !contains(content.CSS, "body") {
		t.Error("CSS missing body styles")
	}

	// Validate static fallback
	if !contains(content.StaticFallback, "Static") {
		t.Error("Static fallback doesn't indicate static mode")
	}
}

func validateAssetBundle(t *testing.T, assets *AssetBundle) {
	if len(assets.Images) == 0 {
		t.Error("No images in asset bundle")
	}

	if len(assets.Fonts) == 0 {
		t.Error("No fonts in asset bundle")
	}

	if len(assets.Data) == 0 {
		t.Error("No data files in asset bundle")
	}

	// Validate specific assets
	if _, exists := assets.Images["logo.svg"]; !exists {
		t.Error("Logo image missing from assets")
	}

	if _, exists := assets.Data["chart-data.json"]; !exists {
		t.Error("Chart data missing from assets")
	}

	// Validate asset content
	for name, data := range assets.Images {
		if len(data) == 0 {
			t.Errorf("Image asset %s is empty", name)
		}
	}

	for name, data := range assets.Fonts {
		if len(data) == 0 {
			t.Errorf("Font asset %s is empty", name)
		}
	}

	for name, data := range assets.Data {
		if len(data) == 0 {
			t.Errorf("Data asset %s is empty", name)
		}
	}
}

func validateWASMModules(t *testing.T, wasmModules map[string][]byte, wasmConfig *WASMConfiguration) {
	if len(wasmModules) == 0 {
		t.Error("No WASM modules present")
	}

	if wasmConfig == nil {
		t.Fatal("WASM configuration is nil")
	}

	// Validate module consistency
	for moduleName, moduleData := range wasmModules {
		if len(moduleData) < 8 {
			t.Errorf("WASM module %s is too small", moduleName)
			continue
		}

		// Check WASM magic number
		if moduleData[0] != 0x00 || moduleData[1] != 0x61 || 
		   moduleData[2] != 0x73 || moduleData[3] != 0x6D {
			t.Errorf("WASM module %s has invalid magic number", moduleName)
		}

		// Check WASM version
		if moduleData[4] != 0x01 || moduleData[5] != 0x00 ||
		   moduleData[6] != 0x00 || moduleData[7] != 0x00 {
			t.Errorf("WASM module %s has invalid version", moduleName)
		}

		// Check configuration exists
		if _, exists := wasmConfig.Modules[moduleName]; !exists {
			t.Errorf("WASM module %s not configured", moduleName)
		}
	}

	// Validate configuration completeness
	for moduleName, moduleConfig := range wasmConfig.Modules {
		if moduleConfig.Name != moduleName {
			t.Errorf("Module name mismatch: config says %s, key is %s", moduleConfig.Name, moduleName)
		}

		if moduleConfig.Version == "" {
			t.Errorf("Module %s missing version", moduleName)
		}

		if moduleConfig.EntryPoint == "" {
			t.Errorf("Module %s missing entry point", moduleName)
		}

		if len(moduleConfig.Exports) == 0 {
			t.Errorf("Module %s has no exports", moduleName)
		}
	}
}

func validateSignatureBundle(t *testing.T, signatures *SignatureBundle) {
	if signatures.ContentSignature == "" {
		t.Error("Content signature is empty")
	}

	if signatures.ManifestSignature == "" {
		t.Error("Manifest signature is empty")
	}

	if len(signatures.WASMSignatures) == 0 {
		t.Error("No WASM signatures present")
	}

	// Validate signature format (should be base64-like)
	signatures_to_check := []string{
		signatures.ContentSignature,
		signatures.ManifestSignature,
	}

	for _, sig := range signatures_to_check {
		if len(sig) < 32 {
			t.Error("Signature appears too short")
		}
	}

	for moduleName, signature := range signatures.WASMSignatures {
		if signature == "" {
			t.Errorf("WASM signature for %s is empty", moduleName)
		}
	}
}

func validateCrossComponentConsistency(t *testing.T, document *LIVDocument) {
	// Validate WASM module consistency
	for moduleName := range document.WASMModules {
		if _, exists := document.Manifest.WASMConfig.Modules[moduleName]; !exists {
			t.Errorf("WASM module %s exists but not configured", moduleName)
		}

		if _, exists := document.Signatures.WASMSignatures[moduleName]; !exists {
			t.Errorf("WASM module %s exists but not signed", moduleName)
		}
	}

	// Validate feature flag consistency
	if document.Manifest.Features.WebAssembly && len(document.WASMModules) == 0 {
		t.Error("WebAssembly feature enabled but no WASM modules present")
	}

	if document.Manifest.Features.Charts && !document.Manifest.Features.Interactivity {
		t.Error("Charts feature enabled but interactivity disabled")
	}

	// Validate security policy consistency
	if document.Manifest.Security.WASMPermissions.MemoryLimit < document.Manifest.WASMConfig.MemoryLimit {
		t.Error("WASM config memory limit exceeds security policy limit")
	}
}

func validateJSONRoundTrip(t *testing.T, document *LIVDocument) {
	// Marshal to JSON
	data, err := json.Marshal(document)
	if err != nil {
		t.Fatalf("Failed to marshal document to JSON: %v", err)
	}

	// Unmarshal from JSON
	var unmarshaled LIVDocument
	if err := json.Unmarshal(data, &unmarshaled); err != nil {
		t.Fatalf("Failed to unmarshal document from JSON: %v", err)
	}

	// Validate key fields survived round-trip
	if unmarshaled.Manifest.Metadata.Title != document.Manifest.Metadata.Title {
		t.Error("Title lost in JSON round-trip")
	}

	if len(unmarshaled.Assets.Images) != len(document.Assets.Images) {
		t.Error("Image count changed in JSON round-trip")
	}

	if len(unmarshaled.WASMModules) != len(document.WASMModules) {
		t.Error("WASM module count changed in JSON round-trip")
	}

	if unmarshaled.Manifest.Features.WebAssembly != document.Manifest.Features.WebAssembly {
		t.Error("WebAssembly feature flag lost in JSON round-trip")
	}
}

func validateSecurityPolicyEnforcement(t *testing.T, policy *SecurityPolicy) {
	// Validate secure defaults
	if policy.WASMPermissions.AllowNetworking {
		t.Error("WASM networking should be disabled by default")
	}

	if policy.WASMPermissions.AllowFileSystem {
		t.Error("WASM file system access should be disabled by default")
	}

	if policy.NetworkPolicy.AllowOutbound {
		t.Error("Network outbound should be disabled by default")
	}

	// Validate memory limits are reasonable
	if policy.WASMPermissions.MemoryLimit > 256*1024*1024 { // 256MB
		t.Error("WASM memory limit is very high")
	}

	if policy.WASMPermissions.CPUTimeLimit > 30000 { // 30 seconds
		t.Error("WASM CPU time limit is very high")
	}

	// Validate CSP is present
	if policy.ContentSecurityPolicy == "" {
		t.Error("Content Security Policy is empty")
	}
}

func validateFeatureFlagConsistency(t *testing.T, features *FeatureFlags, wasmConfig *WASMConfiguration) {
	// Validate WebAssembly consistency
	if features.WebAssembly && (wasmConfig == nil || len(wasmConfig.Modules) == 0) {
		t.Error("WebAssembly feature enabled but no WASM modules configured")
	}

	// Validate interactivity dependencies
	if features.Charts && !features.Interactivity {
		t.Error("Charts require interactivity to be enabled")
	}

	if features.WebGL && !features.Interactivity {
		t.Error("WebGL requires interactivity to be enabled")
	}

	// Validate multimedia combinations
	if features.Audio && features.Video {
		// This is fine, just noting the combination
		t.Logf("Document has both audio and video features enabled")
	}
}

// Helper function
func contains(s, substr string) bool {
	return len(s) >= len(substr) && (s == substr || 
		(len(s) > len(substr) && 
		 (s[:len(substr)] == substr || 
		  s[len(s)-len(substr):] == substr || 
		  containsSubstring(s, substr))))
}

func containsSubstring(s, substr string) bool {
	for i := 0; i <= len(s)-len(substr); i++ {
		if s[i:i+len(substr)] == substr {
			return true
		}
	}
	return false
}