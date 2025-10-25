package integration

import (
	"bytes"
	"context"
	"os"
	"path/filepath"
	"strings"
	"testing"

	"github.com/liv-format/liv/pkg/container"
	"github.com/liv-format/liv/pkg/core"
	"github.com/liv-format/liv/pkg/integrity"
	"github.com/liv-format/liv/pkg/manifest"
)

// TestCompleteWorkflow tests the entire LIV format workflow from creation to verification
func TestCompleteWorkflow(t *testing.T) {
	// Create temporary directory for test files
	tempDir, err := os.MkdirTemp("", "liv-integration-test-*")
	if err != nil {
		t.Fatalf("Failed to create temp directory: %v", err)
	}
	defer os.RemoveAll(tempDir)

	// Step 1: Create source files
	sourceDir := filepath.Join(tempDir, "source")
	createSourceFiles(t, sourceDir)

	// Step 2: Build manifest
	manifestPath := filepath.Join(sourceDir, "manifest.json")
	createManifestFile(t, manifestPath)

	// Step 3: Package into .liv file
	livFile := filepath.Join(tempDir, "test-document.liv")
	packageSourceFiles(t, sourceDir, livFile)

	// Step 4: Validate package structure
	validatePackageStructure(t, livFile)

	// Step 5: Extract and verify content
	extractDir := filepath.Join(tempDir, "extracted")
	extractAndVerify(t, livFile, extractDir, sourceDir)

	// Step 6: Test integrity verification
	verifyIntegrity(t, livFile)

	// Step 7: Test with signatures
	testWithSignatures(t, livFile, tempDir)
}

func createSourceFiles(t *testing.T, sourceDir string) {
	// Create directory structure
	dirs := []string{
		"content",
		"content/styles",
		"content/scripts",
		"content/static",
		"assets/images",
		"assets/fonts",
		"assets/data",
		"wasm",
		"signatures",
	}

	for _, dir := range dirs {
		fullDir := filepath.Join(sourceDir, dir)
		if err := os.MkdirAll(fullDir, 0755); err != nil {
			t.Fatalf("Failed to create directory %s: %v", dir, err)
		}
	}

	// Create content files
	files := map[string]string{
		"content/index.html": `<!DOCTYPE html>
<html lang="en">
<head>
    <meta charset="UTF-8">
    <title>Integration Test Document</title>
    <link rel="stylesheet" href="styles/main.css">
</head>
<body>
    <h1>LIV Integration Test</h1>
    <div id="app">
        <p>This is an integration test document.</p>
        <button onclick="testInteraction()">Test Interaction</button>
    </div>
    <script src="scripts/main.js"></script>
</body>
</html>`,
		"content/styles/main.css": `body {
    font-family: Arial, sans-serif;
    margin: 0;
    padding: 20px;
    background: #f5f5f5;
}

h1 {
    color: #333;
    text-align: center;
}

#app {
    max-width: 800px;
    margin: 0 auto;
    background: white;
    padding: 20px;
    border-radius: 8px;
    box-shadow: 0 2px 4px rgba(0,0,0,0.1);
}

button {
    background: #007bff;
    color: white;
    border: none;
    padding: 10px 20px;
    border-radius: 4px;
    cursor: pointer;
}

button:hover {
    background: #0056b3;
}`,
		"content/scripts/main.js": `function testInteraction() {
    console.log('Interaction test triggered');
    const button = event.target;
    button.textContent = 'Clicked!';
    
    setTimeout(() => {
        button.textContent = 'Test Interaction';
    }, 1000);
}

document.addEventListener('DOMContentLoaded', function() {
    console.log('Integration test document loaded');
});`,
		"content/static/fallback.html": `<!DOCTYPE html>
<html>
<head>
    <title>Integration Test - Static Fallback</title>
</head>
<body>
    <h1>LIV Integration Test (Static)</h1>
    <p>This is the static fallback version.</p>
</body>
</html>`,
		"assets/data/test-data.json": `{
    "test": true,
    "data": [1, 2, 3, 4, 5],
    "metadata": {
        "created": "2024-01-15T10:00:00Z",
        "type": "integration-test"
    }
}`,
	}

	for path, content := range files {
		fullPath := filepath.Join(sourceDir, path)
		if err := os.WriteFile(fullPath, []byte(content), 0644); err != nil {
			t.Fatalf("Failed to write file %s: %v", path, err)
		}
	}

	// Create a minimal WASM module
	wasmData := []byte{
		0x00, 0x61, 0x73, 0x6D, // Magic number
		0x01, 0x00, 0x00, 0x00, // Version
		// Minimal type section
		0x01, 0x04, 0x01, 0x60, 0x00, 0x00,
		// Function section
		0x03, 0x02, 0x01, 0x00,
		// Code section
		0x0A, 0x04, 0x01, 0x02, 0x00, 0x0B,
	}
	
	wasmPath := filepath.Join(sourceDir, "wasm/test-engine.wasm")
	if err := os.WriteFile(wasmPath, wasmData, 0644); err != nil {
		t.Fatalf("Failed to write WASM file: %v", err)
	}
}

func createManifestFile(t *testing.T, manifestPath string) {
	builder := manifest.CreateInteractiveDocumentTemplate("Integration Test Document", "Integration Test Suite")

	// Add WASM module configuration
	wasmModule := &core.WASMModule{
		Name:       "test-engine",
		Version:    "1.0.0",
		EntryPoint: "init",
		Exports:    []string{"init", "process", "cleanup"},
		Imports:    []string{"env.memory", "env.console_log"},
		Permissions: &core.WASMPermissions{
			MemoryLimit:     32 * 1024 * 1024, // 32MB
			AllowedImports:  []string{"env"},
			CPUTimeLimit:    5000, // 5 seconds
			AllowNetworking: false,
			AllowFileSystem: false,
		},
	}
	builder.AddWASMModule(wasmModule)

	// Save manifest
	if err := builder.SaveToFile(manifestPath); err != nil {
		t.Fatalf("Failed to save manifest: %v", err)
	}
}

func packageSourceFiles(t *testing.T, sourceDir, livFile string) {
	zipContainer := container.NewZIPContainer()
	
	if err := zipContainer.CreateFromDirectory(sourceDir, livFile); err != nil {
		t.Fatalf("Failed to package source files: %v", err)
	}

	// Verify file was created
	if _, err := os.Stat(livFile); os.IsNotExist(err) {
		t.Fatal("LIV file was not created")
	}

	// Check file size
	info, err := os.Stat(livFile)
	if err != nil {
		t.Fatalf("Failed to get LIV file info: %v", err)
	}

	if info.Size() == 0 {
		t.Fatal("LIV file is empty")
	}

	t.Logf("Created LIV file: %s (%d bytes)", livFile, info.Size())
}

func validatePackageStructure(t *testing.T, livFile string) {
	zipContainer := container.NewZIPContainer()
	
	// Validate structure
	result := zipContainer.ValidateStructure(livFile)
	if !result.IsValid {
		t.Errorf("Package structure validation failed: %v", result.Errors)
	}

	if len(result.Warnings) > 0 {
		t.Logf("Package structure warnings: %v", result.Warnings)
	}

	// Get file list
	files, err := zipContainer.GetFileList(livFile)
	if err != nil {
		t.Fatalf("Failed to get file list: %v", err)
	}

	t.Logf("Package contains %d files", len(files))

	// Verify required files exist
	requiredFiles := []string{"manifest.json", "content/index.html"}
	for _, required := range requiredFiles {
		found := false
		for _, file := range files {
			if file == required {
				found = true
				break
			}
		}
		if !found {
			t.Errorf("Required file %s not found in package", required)
		}
	}
}

func extractAndVerify(t *testing.T, livFile, extractDir, originalSourceDir string) {
	zipContainer := container.NewZIPContainer()
	
	// Extract package
	if err := zipContainer.ExtractToDirectory(livFile, extractDir); err != nil {
		t.Fatalf("Failed to extract package: %v", err)
	}

	// Verify extracted files exist
	requiredFiles := []string{
		"manifest.json",
		"content/index.html",
		"content/styles/main.css",
		"content/scripts/main.js",
		"content/static/fallback.html",
		"assets/data/test-data.json",
		"wasm/test-engine.wasm",
	}

	for _, file := range requiredFiles {
		extractedPath := filepath.Join(extractDir, file)
		if _, err := os.Stat(extractedPath); os.IsNotExist(err) {
			t.Errorf("Extracted file %s does not exist", file)
		}
	}

	// Compare key files with originals
	compareFiles := []string{
		"content/index.html",
		"content/styles/main.css",
		"content/scripts/main.js",
	}

	for _, file := range compareFiles {
		originalPath := filepath.Join(originalSourceDir, file)
		extractedPath := filepath.Join(extractDir, file)

		originalContent, err := os.ReadFile(originalPath)
		if err != nil {
			t.Errorf("Failed to read original file %s: %v", file, err)
			continue
		}

		extractedContent, err := os.ReadFile(extractedPath)
		if err != nil {
			t.Errorf("Failed to read extracted file %s: %v", file, err)
			continue
		}

		if !bytes.Equal(originalContent, extractedContent) {
			t.Errorf("Content mismatch for file %s", file)
		}
	}
}

func verifyIntegrity(t *testing.T, livFile string) {
	// Extract document
	packageManager := container.NewPackageManager()
	file, err := os.Open(livFile)
	if err != nil {
		t.Fatalf("Failed to open LIV file: %v", err)
	}
	defer file.Close()

	document, err := packageManager.ExtractPackage(context.Background(), file)
	if err != nil {
		t.Fatalf("Failed to extract document: %v", err)
	}

	// Validate structure
	result := packageManager.ValidateStructure(document)
	if !result.IsValid {
		t.Errorf("Document structure validation failed: %v", result.Errors)
	}

	// Create integrity validator
	validator := integrity.NewIntegrityValidator()

	// Validate WASM modules
	wasmResult := validator.ValidateWASMModules(document.Manifest.WASMConfig, document.WASMModules)
	if !wasmResult.IsValid {
		t.Errorf("WASM module validation failed: %v", wasmResult.Errors)
	}

	t.Logf("Integrity verification completed successfully")
}

func testWithSignatures(t *testing.T, livFile, tempDir string) {
	// Generate key pair
	sm := integrity.NewSignatureManager()
	keyPair, err := sm.GenerateKeyPair(2048)
	if err != nil {
		t.Fatalf("Failed to generate key pair: %v", err)
	}

	// Save keys
	privateKeyFile := filepath.Join(tempDir, "private.pem")
	publicKeyFile := filepath.Join(tempDir, "public.pem")

	if err := sm.SavePrivateKeyPEM(keyPair, privateKeyFile); err != nil {
		t.Fatalf("Failed to save private key: %v", err)
	}

	if err := sm.SavePublicKeyPEM(keyPair, publicKeyFile); err != nil {
		t.Fatalf("Failed to save public key: %v", err)
	}

	// Extract document
	packageManager := container.NewPackageManager()
	file, err := os.Open(livFile)
	if err != nil {
		t.Fatalf("Failed to open LIV file: %v", err)
	}
	defer file.Close()

	document, err := packageManager.ExtractPackage(context.Background(), file)
	if err != nil {
		t.Fatalf("Failed to extract document: %v", err)
	}

	// Sign document
	signatures, err := sm.SignDocument(document, keyPair.PrivateKey)
	if err != nil {
		t.Fatalf("Failed to sign document: %v", err)
	}

	document.Signatures = signatures

	// Save signed document
	signedFile := filepath.Join(tempDir, "signed-document.liv")
	if err := packageManager.SavePackage(document, signedFile); err != nil {
		t.Fatalf("Failed to save signed document: %v", err)
	}

	// Verify signatures
	signedFileHandle, err := os.Open(signedFile)
	if err != nil {
		t.Fatalf("Failed to open signed file: %v", err)
	}
	defer signedFileHandle.Close()

	signedDocument, err := packageManager.ExtractPackage(context.Background(), signedFileHandle)
	if err != nil {
		t.Fatalf("Failed to extract signed document: %v", err)
	}

	// Verify signatures
	verificationResult := sm.VerifyDocument(signedDocument, keyPair.PublicKey)
	if !verificationResult.Valid {
		t.Errorf("Signature verification failed: %v", verificationResult.Errors)
	}

	if !verificationResult.ManifestValid {
		t.Error("Manifest signature verification failed")
	}

	if !verificationResult.ContentValid {
		t.Error("Content signature verification failed")
	}

	t.Logf("Signature verification completed successfully")
}

// TestRealWorldScenarios tests realistic usage scenarios
func TestRealWorldScenarios(t *testing.T) {
	scenarios := []struct {
		name        string
		description string
		setupFunc   func(*testing.T) *core.LIVDocument
		testFunc    func(*testing.T, *core.LIVDocument)
	}{
		{
			name:        "static_document",
			description: "Static document with no interactive features",
			setupFunc:   createStaticDocument,
			testFunc:    testStaticDocument,
		},
		{
			name:        "interactive_document",
			description: "Interactive document with WASM and animations",
			setupFunc:   createInteractiveDocument,
			testFunc:    testInteractiveDocument,
		},
		{
			name:        "multimedia_document",
			description: "Document with images, fonts, and data files",
			setupFunc:   createMultimediaDocument,
			testFunc:    testMultimediaDocument,
		},
	}

	for _, scenario := range scenarios {
		t.Run(scenario.name, func(t *testing.T) {
			t.Logf("Testing scenario: %s", scenario.description)
			
			document := scenario.setupFunc(t)
			scenario.testFunc(t, document)
			
			// Test packaging and extraction for each scenario
			testPackageExtractCycle(t, document)
		})
	}
}

func createStaticDocument(t *testing.T) *core.LIVDocument {
	builder := manifest.CreateStaticDocumentTemplate("Static Test Document", "Test Suite")
	
	// Add required content/index.html resource
	htmlContent := `<!DOCTYPE html>
<html>
<head><title>Static Document</title></head>
<body>
    <h1>Static LIV Document</h1>
    <p>This document contains only static content.</p>
</body>
</html>`
	builder.AddResource("content/index.html", &core.Resource{
		Path: "content/index.html",
		Type: "text/html",
		Size: int64(len(htmlContent)),
		Hash: "test-hash",
	})
	
	manifestObj, err := builder.Build()
	if err != nil {
		t.Fatalf("Failed to build static manifest: %v", err)
	}

	return &core.LIVDocument{
		Manifest: manifestObj,
		Content: &core.DocumentContent{
			HTML: htmlContent,
			CSS: `body { font-family: serif; margin: 20px; }`,
			StaticFallback: `<!DOCTYPE html>
<html>
<head><title>Static Document</title></head>
<body>
    <h1>Static LIV Document</h1>
    <p>This document contains only static content.</p>
</body>
</html>`,
		},
		Assets: &core.AssetBundle{
			Images: map[string][]byte{
				"simple.png": []byte("simple-png-data"),
			},
			Fonts: map[string][]byte{},
			Data:  map[string][]byte{},
		},
		Signatures:  &core.SignatureBundle{},
		WASMModules: map[string][]byte{},
	}
}

func createInteractiveDocument(t *testing.T) *core.LIVDocument {
	builder := manifest.CreateInteractiveDocumentTemplate("Interactive Test Document", "Test Suite")
	
	// Add WASM module
	wasmModule := &core.WASMModule{
		Name:       "interactive-engine",
		Version:    "1.0.0",
		EntryPoint: "init_engine",
		Exports:    []string{"init_engine", "render", "interact"},
		Imports:    []string{"env.memory"},
	}
	builder.AddWASMModule(wasmModule)

	// Add required content/index.html resource
	htmlContent := `<!DOCTYPE html>
<html>
<head>
    <title>Interactive Document</title>
    <style>
        .animated { animation: pulse 2s infinite; }
        @keyframes pulse { 0% { opacity: 1; } 50% { opacity: 0.5; } 100% { opacity: 1; } }
    </style>
</head>
<body>
    <h1 class="animated">Interactive LIV Document</h1>
    <canvas id="chart" width="400" height="300"></canvas>
    <button onclick="updateChart()">Update Chart</button>
</body>
</html>`
	builder.AddResource("content/index.html", &core.Resource{
		Path: "content/index.html",
		Type: "text/html",
		Size: int64(len(htmlContent)),
		Hash: "test-hash",
	})

	manifestObj, err := builder.Build()
	if err != nil {
		t.Fatalf("Failed to build interactive manifest: %v", err)
	}

	return &core.LIVDocument{
		Manifest: manifestObj,
		Content: &core.DocumentContent{
			HTML: htmlContent,
			CSS: `body { font-family: sans-serif; margin: 20px; }
.animated { color: #007bff; }
canvas { border: 1px solid #ccc; }`,
			InteractiveSpec: `function updateChart() {
    console.log('Updating chart via WASM engine');
    // WASM interaction would happen here
}`,
			StaticFallback: `<!DOCTYPE html>
<html>
<head><title>Interactive Document - Static</title></head>
<body>
    <h1>Interactive LIV Document (Static Mode)</h1>
    <p>Interactive features disabled.</p>
</body>
</html>`,
		},
		Assets: &core.AssetBundle{
			Images: map[string][]byte{
				"chart-bg.svg": []byte(`<svg xmlns="http://www.w3.org/2000/svg" viewBox="0 0 400 300"><rect width="400" height="300" fill="#f8f9fa"/></svg>`),
			},
			Fonts: map[string][]byte{},
			Data: map[string][]byte{
				"chart-data.json": []byte(`{"values": [10, 20, 30, 40, 50]}`),
			},
		},
		Signatures: &core.SignatureBundle{},
		WASMModules: map[string][]byte{
			"interactive-engine": {
				0x00, 0x61, 0x73, 0x6D, 0x01, 0x00, 0x00, 0x00,
				0x01, 0x04, 0x01, 0x60, 0x00, 0x00,
				0x03, 0x02, 0x01, 0x00,
				0x0A, 0x04, 0x01, 0x02, 0x00, 0x0B,
			},
		},
	}
}

func createMultimediaDocument(t *testing.T) *core.LIVDocument {
	builder := manifest.CreateInteractiveDocumentTemplate("Multimedia Test Document", "Test Suite")
	
	// Enable multimedia features
	features := &core.FeatureFlags{
		Animations:    true,
		Interactivity: true,
		Charts:        true,
		Forms:         true,
		Audio:         true,
		Video:         false, // Keep video disabled for test simplicity
		WebGL:         true,
		WebAssembly:   true,
	}
	builder.SetFeatureFlags(features)

	// Add required content/index.html resource
	htmlContent := `<!DOCTYPE html>
<html>
<head>
    <title>Multimedia Document</title>
</head>
<body>
    <h1>Multimedia LIV Document</h1>
    <img src="assets/images/hero.jpg" alt="Hero Image">
    <audio controls src="assets/audio/background.mp3"></audio>
    <canvas id="webgl-canvas" width="400" height="300"></canvas>
</body>
</html>`
	builder.AddResource("content/index.html", &core.Resource{
		Path: "content/index.html",
		Type: "text/html",
		Size: int64(len(htmlContent)),
		Hash: "test-hash",
	})

	manifestObj, err := builder.Build()
	if err != nil {
		t.Fatalf("Failed to build multimedia manifest: %v", err)
	}

	return &core.LIVDocument{
		Manifest: manifestObj,
		Content: &core.DocumentContent{
			HTML: htmlContent,
			CSS: `@font-face {
    font-family: 'CustomFont';
    src: url('assets/fonts/custom.woff2') format('woff2');
}
body { font-family: 'CustomFont', sans-serif; }`,
			InteractiveSpec: `// WebGL and audio integration
const canvas = document.getElementById('webgl-canvas');
const gl = canvas.getContext('webgl');
// WebGL setup would go here`,
		},
		Assets: &core.AssetBundle{
			Images: map[string][]byte{
				"hero.jpg":    []byte("fake-jpeg-hero-image-data"),
				"icon.svg":    []byte(`<svg xmlns="http://www.w3.org/2000/svg" viewBox="0 0 24 24"><circle cx="12" cy="12" r="10"/></svg>`),
				"texture.png": []byte("fake-png-texture-data"),
			},
			Fonts: map[string][]byte{
				"custom.woff2":    []byte("fake-woff2-custom-font-data"),
				"icons.woff":      []byte("fake-woff-icon-font-data"),
			},
			Data: map[string][]byte{
				"audio-config.json": []byte(`{"volume": 0.8, "autoplay": false}`),
				"webgl-shaders.json": []byte(`{"vertex": "...", "fragment": "..."}`),
				"large-dataset.csv": []byte("id,value,category\n1,100,A\n2,200,B\n3,300,C"),
			},
		},
		Signatures:  &core.SignatureBundle{},
		WASMModules: map[string][]byte{},
	}
}

func testStaticDocument(t *testing.T, document *core.LIVDocument) {
	// Verify static document properties
	if document.Manifest.Features.Interactivity {
		t.Error("Static document should not have interactivity enabled")
	}

	if document.Manifest.Features.WebAssembly {
		t.Error("Static document should not have WebAssembly enabled")
	}

	if document.Manifest.Security.JSPermissions.ExecutionMode != "none" {
		t.Error("Static document should have JavaScript execution disabled")
	}

	if len(document.WASMModules) > 0 {
		t.Error("Static document should not have WASM modules")
	}
}

func testInteractiveDocument(t *testing.T, document *core.LIVDocument) {
	// Verify interactive document properties
	if !document.Manifest.Features.Interactivity {
		t.Error("Interactive document should have interactivity enabled")
	}

	if !document.Manifest.Features.WebAssembly {
		t.Error("Interactive document should have WebAssembly enabled")
	}

	if document.Manifest.Security.JSPermissions.ExecutionMode == "none" {
		t.Error("Interactive document should allow JavaScript execution")
	}

	if len(document.WASMModules) == 0 {
		t.Error("Interactive document should have WASM modules")
	}

	// Verify WASM module configuration
	if document.Manifest.WASMConfig == nil {
		t.Fatal("Interactive document should have WASM configuration")
	}

	if len(document.Manifest.WASMConfig.Modules) == 0 {
		t.Error("Interactive document should have configured WASM modules")
	}
}

func testMultimediaDocument(t *testing.T, document *core.LIVDocument) {
	// Verify multimedia features
	if !document.Manifest.Features.Audio {
		t.Error("Multimedia document should have audio enabled")
	}

	if !document.Manifest.Features.WebGL {
		t.Error("Multimedia document should have WebGL enabled")
	}

	// Verify multimedia assets
	if len(document.Assets.Images) == 0 {
		t.Error("Multimedia document should have images")
	}

	if len(document.Assets.Fonts) == 0 {
		t.Error("Multimedia document should have fonts")
	}

	// Verify content references multimedia assets
	if !contains(document.Content.HTML, "img src=") {
		t.Error("HTML should reference images")
	}

	if !contains(document.Content.HTML, "audio") {
		t.Error("HTML should include audio elements")
	}

	if !contains(document.Content.CSS, "@font-face") {
		t.Error("CSS should include font definitions")
	}
}

func testPackageExtractCycle(t *testing.T, document *core.LIVDocument) {
	packageManager := container.NewPackageManager()

	// Package document
	var buf bytes.Buffer
	err := packageManager.SavePackageToWriter(document, &buf)
	if err != nil {
		t.Fatalf("Failed to package document: %v", err)
	}

	// Extract document
	reader := bytes.NewReader(buf.Bytes())
	extractedDocument, err := packageManager.ExtractPackage(context.Background(), reader)
	if err != nil {
		t.Fatalf("Failed to extract document: %v", err)
	}

	// Validate extracted document
	result := packageManager.ValidateStructure(extractedDocument)
	if !result.IsValid {
		t.Errorf("Extracted document validation failed: %v", result.Errors)
	}

	// Compare key properties
	if extractedDocument.Manifest.Metadata.Title != document.Manifest.Metadata.Title {
		t.Error("Document title changed during package/extract cycle")
	}

	if len(extractedDocument.Assets.Images) != len(document.Assets.Images) {
		t.Error("Asset count changed during package/extract cycle")
	}
}

// Helper function
func contains(s, substr string) bool {
	return strings.Contains(s, substr)
}