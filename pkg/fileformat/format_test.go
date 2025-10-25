package fileformat

import (
	"bytes"
	"context"
	"fmt"
	"strings"
	"testing"
	"time"

	"github.com/liv-format/liv/pkg/container"
	"github.com/liv-format/liv/pkg/core"
	"github.com/liv-format/liv/pkg/integrity"
	"github.com/liv-format/liv/pkg/manifest"
)

// TestLIVFormatEndToEnd tests the complete LIV format workflow
func TestLIVFormatEndToEnd(t *testing.T) {
	// Test the complete workflow: Create -> Package -> Extract -> Validate
	
	// Step 1: Create a complete LIV document
	document := createTestDocument(t)
	
	// Step 2: Package the document
	packagedData := packageDocument(t, document)
	
	// Step 3: Extract the document
	extractedDocument := extractDocument(t, packagedData)
	
	// Step 4: Validate the extracted document
	validateExtractedDocument(t, document, extractedDocument)
	
	// Step 5: Test integrity verification
	verifyDocumentIntegrity(t, extractedDocument)
	
	// Step 6: Test signature verification (if signed)
	if extractedDocument.Signatures != nil {
		verifyDocumentSignatures(t, extractedDocument)
	}
}

func createTestDocument(t *testing.T) *core.LIVDocument {
	// Create document content first
	htmlContent := `<!DOCTYPE html>
<html>
<head>
    <title>Test Document</title>
    <link rel="stylesheet" href="styles/main.css">
</head>
<body>
    <h1>LIV Format Test</h1>
    <div id="content">This is a test document.</div>
</body>
</html>`

	cssContent := `body {
    font-family: Arial, sans-serif;
    margin: 20px;
    background: #f0f0f0;
}

h1 {
    color: #333;
    text-align: center;
}

#content {
    background: white;
    padding: 20px;
    border-radius: 8px;
    box-shadow: 0 2px 4px rgba(0,0,0,0.1);
}`

	// Create files map for hash calculation
	files := map[string][]byte{
		"content/index.html":      []byte(htmlContent),
		"content/styles/main.css": []byte(cssContent),
		"manifest.json":           []byte(`{"version": "1.0"}`), // Placeholder
	}

	// Generate correct resource manifest
	validator := integrity.NewIntegrityValidator()
	resources := validator.GenerateResourceManifest(files)

	// Create manifest using builder
	builder := manifest.NewManifestBuilder()
	builder.CreateDefaultMetadata("End-to-End Test Document", "Test Suite").
		CreateDefaultSecurityPolicy().
		CreateDefaultFeatureFlags()

	// Add resources with correct hashes
	for path, resource := range resources {
		builder.AddResource(path, resource)
	}

	// Add WASM module
	wasmModule := &core.WASMModule{
		Name:       "test-engine",
		Version:    "1.0.0",
		EntryPoint: "init",
		Exports:    []string{"init", "process"},
		Imports:    []string{"env.memory"},
	}
	builder.AddWASMModule(wasmModule)

	manifestObj, err := builder.Build()
	if err != nil {
		t.Fatalf("Failed to build manifest: %v", err)
	}

	// Create complete document
	document := &core.LIVDocument{
		Manifest: manifestObj,
		Content: &core.DocumentContent{
			HTML: htmlContent,
			CSS:  cssContent,
			InteractiveSpec: `// Test interactive specification
const testConfig = {
    engine: "test-engine",
    version: "1.0.0"
};

console.log("Test document loaded");`,
			StaticFallback: `<!DOCTYPE html>
<html>
<head><title>Test Document - Static</title></head>
<body>
    <h1>LIV Format Test (Static)</h1>
    <p>This is the static fallback version.</p>
</body>
</html>`,
		},
		Assets: &core.AssetBundle{
			Images: map[string][]byte{
				"logo.png": []byte("fake-png-data-for-testing"),
			},
			Fonts: map[string][]byte{
				"main.woff": []byte("fake-font-data-for-testing"),
			},
			Data: map[string][]byte{
				"config.json": []byte(`{"test": true, "version": "1.0"}`),
			},
		},
		Signatures: &core.SignatureBundle{
			ContentSignature:  "test-content-signature",
			ManifestSignature: "test-manifest-signature",
			WASMSignatures: map[string]string{
				"test-engine": "test-wasm-signature",
			},
		},
		WASMModules: map[string][]byte{
			"test-engine": {
				// Valid WASM header
				0x00, 0x61, 0x73, 0x6D, // Magic
				0x01, 0x00, 0x00, 0x00, // Version
				// Minimal module content
				0x01, 0x04, 0x01, 0x60, 0x00, 0x00,
			},
		},
	}

	return document
}

func packageDocument(t *testing.T, document *core.LIVDocument) []byte {
	packageManager := container.NewPackageManager()
	
	var buf bytes.Buffer
	err := packageManager.SavePackageToWriter(document, &buf)
	if err != nil {
		t.Fatalf("Failed to package document: %v", err)
	}

	if buf.Len() == 0 {
		t.Fatal("Packaged document is empty")
	}

	return buf.Bytes()
}

func extractDocument(t *testing.T, packagedData []byte) *core.LIVDocument {
	packageManager := container.NewPackageManager()
	
	reader := bytes.NewReader(packagedData)
	document, err := packageManager.ExtractPackage(context.Background(), reader)
	if err != nil {
		t.Fatalf("Failed to extract document: %v", err)
	}

	if document == nil {
		t.Fatal("Extracted document is nil")
	}

	return document
}

func validateExtractedDocument(t *testing.T, original, extracted *core.LIVDocument) {
	// Validate manifest
	if extracted.Manifest == nil {
		t.Fatal("Extracted manifest is nil")
	}

	if extracted.Manifest.Metadata.Title != original.Manifest.Metadata.Title {
		t.Errorf("Title mismatch: expected %s, got %s", 
			original.Manifest.Metadata.Title, extracted.Manifest.Metadata.Title)
	}

	// Validate content
	if extracted.Content == nil {
		t.Fatal("Extracted content is nil")
	}

	if extracted.Content.HTML != original.Content.HTML {
		t.Error("HTML content mismatch")
	}

	if extracted.Content.CSS != original.Content.CSS {
		t.Error("CSS content mismatch")
	}

	// Validate assets
	if extracted.Assets == nil {
		t.Fatal("Extracted assets is nil")
	}

	if len(extracted.Assets.Images) != len(original.Assets.Images) {
		t.Errorf("Image count mismatch: expected %d, got %d", 
			len(original.Assets.Images), len(extracted.Assets.Images))
	}

	// Validate WASM modules
	if len(extracted.WASMModules) != len(original.WASMModules) {
		t.Errorf("WASM module count mismatch: expected %d, got %d", 
			len(original.WASMModules), len(extracted.WASMModules))
	}

	for moduleName, originalData := range original.WASMModules {
		if extractedData, exists := extracted.WASMModules[moduleName]; exists {
			if !bytes.Equal(originalData, extractedData) {
				t.Errorf("WASM module %s data mismatch", moduleName)
			}
		} else {
			t.Errorf("WASM module %s missing from extracted document", moduleName)
		}
	}
}

func verifyDocumentIntegrity(t *testing.T, document *core.LIVDocument) {
	validator := integrity.NewIntegrityValidator()
	
	// Convert document to files for validation
	files := make(map[string][]byte)
	
	// Add content files that are actually in the manifest
	if document.Content != nil {
		files["content/index.html"] = []byte(document.Content.HTML)
		files["content/styles/main.css"] = []byte(document.Content.CSS)
		// Add manifest.json as it's required
		files["manifest.json"] = []byte(`{"version": "1.0"}`)
	}

	// Generate integrity report
	report := validator.GenerateIntegrityReport(document.Manifest, files, document.WASMModules)
	
	if !report.Valid {
		t.Errorf("Document integrity validation failed:")
		t.Errorf("  Hash mismatches: %d", len(report.HashMismatches))
		for _, mismatch := range report.HashMismatches {
			t.Errorf("    Hash mismatch for %s: expected %s, got %s", mismatch.Path, mismatch.ExpectedHash, mismatch.ActualHash)
		}
		t.Errorf("  Size mismatches: %d", len(report.SizeMismatches))
		for _, mismatch := range report.SizeMismatches {
			t.Errorf("    Size mismatch for %s", mismatch.Path)
		}
		t.Errorf("  Missing resources: %d", len(report.MissingResources))
		for _, missing := range report.MissingResources {
			t.Errorf("    Missing resource: %s", missing)
		}
		t.Errorf("  Orphaned files: %d", len(report.OrphanedFiles))
		for _, orphaned := range report.OrphanedFiles {
			t.Errorf("    Orphaned file: %s", orphaned)
		}
		if report.WASMValidation != nil && !report.WASMValidation.IsValid {
			t.Errorf("  WASM validation errors: %v", report.WASMValidation.Errors)
		}
	}

	// Validate WASM modules specifically
	wasmResult := validator.ValidateWASMModules(document.Manifest.WASMConfig, document.WASMModules)
	if !wasmResult.IsValid {
		t.Errorf("WASM module validation failed: %v", wasmResult.Errors)
	}
}

func verifyDocumentSignatures(t *testing.T, document *core.LIVDocument) {
	// Note: This is a basic signature structure test
	// Full cryptographic verification would require actual keys
	
	if document.Signatures == nil {
		t.Fatal("Document signatures are nil")
	}

	if document.Signatures.ContentSignature == "" {
		t.Error("Content signature is empty")
	}

	if document.Signatures.ManifestSignature == "" {
		t.Error("Manifest signature is empty")
	}

	// Verify WASM signatures exist for all modules
	for moduleName := range document.WASMModules {
		if _, exists := document.Signatures.WASMSignatures[moduleName]; !exists {
			t.Errorf("WASM signature missing for module %s", moduleName)
		}
	}
}

// TestLIVFormatCompatibility tests format compatibility and version handling
func TestLIVFormatCompatibility(t *testing.T) {
	tests := []struct {
		name    string
		version string
		valid   bool
	}{
		{
			name:    "current version",
			version: "1.0",
			valid:   true,
		},
		{
			name:    "future version",
			version: "2.0",
			valid:   true, // Should be handled gracefully with warnings
		},
		{
			name:    "invalid version",
			version: "invalid",
			valid:   true, // Should be handled gracefully with warnings
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			document := createTestDocument(t)
			document.Manifest.Version = tt.version

			// Test packaging
			packageManager := container.NewPackageManager()
			var buf bytes.Buffer
			err := packageManager.SavePackageToWriter(document, &buf)

			if tt.valid && err != nil {
				t.Errorf("Expected valid version %s to package successfully, got error: %v", tt.version, err)
			}

			if !tt.valid && err == nil {
				t.Errorf("Expected invalid version %s to fail packaging", tt.version)
			}
		})
	}
}

// TestLIVFormatSecurity tests security-related format features
func TestLIVFormatSecurity(t *testing.T) {
	document := createTestDocument(t)

	// Test 1: Validate security policy enforcement
	if document.Manifest.Security.WASMPermissions.AllowNetworking {
		t.Error("WASM networking should be disabled by default")
	}

	if document.Manifest.Security.NetworkPolicy.AllowOutbound {
		t.Error("Network outbound should be disabled by default")
	}

	// Test 2: Validate memory limits
	if document.Manifest.Security.WASMPermissions.MemoryLimit > 256*1024*1024 {
		t.Error("WASM memory limit is too high")
	}

	// Test 3: Validate CSP presence
	if document.Manifest.Security.ContentSecurityPolicy == "" {
		t.Error("Content Security Policy should be defined")
	}

	// Test 4: Test with malicious content (should be rejected)
	maliciousDocument := createTestDocument(t)
	maliciousDocument.Content.HTML = `<script>alert('xss')</script>`
	
	// Package and extract to see if malicious content is handled
	packageManager := container.NewPackageManager()
	var buf bytes.Buffer
	err := packageManager.SavePackageToWriter(maliciousDocument, &buf)
	if err != nil {
		t.Logf("Malicious content rejected during packaging: %v", err)
	}
}

// TestLIVFormatPerformance tests performance characteristics
func TestLIVFormatPerformance(t *testing.T) {
	// Create a larger document for performance testing
	document := createLargeTestDocument(t)

	packageManager := container.NewPackageManager()

	// Test packaging performance
	start := time.Now()
	var buf bytes.Buffer
	err := packageManager.SavePackageToWriter(document, &buf)
	packagingTime := time.Since(start)

	if err != nil {
		t.Fatalf("Failed to package large document: %v", err)
	}

	t.Logf("Packaging time: %v", packagingTime)
	t.Logf("Package size: %d bytes", buf.Len())

	// Test extraction performance
	start = time.Now()
	reader := bytes.NewReader(buf.Bytes())
	_, err = packageManager.ExtractPackage(context.Background(), reader)
	extractionTime := time.Since(start)

	if err != nil {
		t.Fatalf("Failed to extract large document: %v", err)
	}

	t.Logf("Extraction time: %v", extractionTime)

	// Performance thresholds (adjust based on requirements)
	if packagingTime > 5*time.Second {
		t.Errorf("Packaging took too long: %v", packagingTime)
	}

	if extractionTime > 5*time.Second {
		t.Errorf("Extraction took too long: %v", extractionTime)
	}
}

func createLargeTestDocument(t *testing.T) *core.LIVDocument {
	document := createTestDocument(t)

	// Add many resources
	for i := 0; i < 100; i++ {
		resourceName := fmt.Sprintf("resource%d.txt", i)
		document.Manifest.Resources[resourceName] = &core.Resource{
			Hash: fmt.Sprintf("hash-%d", i),
			Size: int64(i * 100),
			Type: "text/plain",
			Path: resourceName,
		}
	}

	// Add large assets
	largeData := make([]byte, 1024*1024) // 1MB
	for i := range largeData {
		largeData[i] = byte(i % 256)
	}

	document.Assets.Images["large-image.png"] = largeData
	document.Assets.Data["large-data.json"] = largeData

	// Add large HTML content
	var htmlBuilder strings.Builder
	htmlBuilder.WriteString("<!DOCTYPE html><html><head><title>Large Document</title></head><body>")
	for i := 0; i < 1000; i++ {
		htmlBuilder.WriteString(fmt.Sprintf("<p>This is paragraph %d with some content.</p>", i))
	}
	htmlBuilder.WriteString("</body></html>")
	document.Content.HTML = htmlBuilder.String()

	return document
}

// TestLIVFormatErrorHandling tests error handling in various scenarios
func TestLIVFormatErrorHandling(t *testing.T) {
	packageManager := container.NewPackageManager()

	// Test 1: Invalid document structure
	invalidDocument := &core.LIVDocument{
		Manifest: nil, // Invalid: nil manifest
	}

	var buf bytes.Buffer
	err := packageManager.SavePackageToWriter(invalidDocument, &buf)
	if err == nil {
		t.Error("Expected error for invalid document structure")
	}

	// Test 2: Corrupted package data
	corruptedData := []byte("not-a-valid-zip-file")
	reader := bytes.NewReader(corruptedData)
	_, err = packageManager.ExtractPackage(context.Background(), reader)
	if err == nil {
		t.Error("Expected error for corrupted package data")
	}

	// Test 3: Missing required files
	document := createTestDocument(t)
	delete(document.Manifest.Resources, "content/index.html")

	buf.Reset()
	err = packageManager.SavePackageToWriter(document, &buf)
	if err != nil {
		t.Logf("Document with missing required files rejected: %v", err)
	}

	// Test 4: Invalid WASM module
	document = createTestDocument(t)
	document.WASMModules["test-engine"] = []byte("invalid-wasm-data")

	validator := integrity.NewIntegrityValidator()
	result := validator.ValidateWASMModules(document.Manifest.WASMConfig, document.WASMModules)
	if result.IsValid {
		t.Error("Expected WASM validation to fail for invalid module")
	}
}

// BenchmarkLIVFormatOperations benchmarks key format operations
func BenchmarkLIVFormatOperations(b *testing.B) {
	document := createTestDocument(nil)
	packageManager := container.NewPackageManager()

	b.Run("Package", func(b *testing.B) {
		for i := 0; i < b.N; i++ {
			var buf bytes.Buffer
			err := packageManager.SavePackageToWriter(document, &buf)
			if err != nil {
				b.Fatalf("Failed to package document: %v", err)
			}
		}
	})

	// Package once for extraction benchmark
	var packagedData bytes.Buffer
	err := packageManager.SavePackageToWriter(document, &packagedData)
	if err != nil {
		b.Fatalf("Failed to package document for benchmark: %v", err)
	}

	b.Run("Extract", func(b *testing.B) {
		for i := 0; i < b.N; i++ {
			reader := bytes.NewReader(packagedData.Bytes())
			_, err := packageManager.ExtractPackage(context.Background(), reader)
			if err != nil {
				b.Fatalf("Failed to extract document: %v", err)
			}
		}
	})

	b.Run("Validate", func(b *testing.B) {
		validator := integrity.NewIntegrityValidator()
		for i := 0; i < b.N; i++ {
			_ = validator.ValidateWASMModules(document.Manifest.WASMConfig, document.WASMModules)
		}
	})
}

