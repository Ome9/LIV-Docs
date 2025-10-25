package main

import (
	"os"
	"path/filepath"
	"strings"
	"testing"

	"github.com/liv-format/liv/pkg/container"
	"github.com/liv-format/liv/pkg/integrity"
	"github.com/liv-format/liv/pkg/manifest"
)

// TestBuilderFunctions tests the builder functions directly
func TestBuilderFunctions(t *testing.T) {
	// Setup test environment
	testDir := setupBuilderTestDir(t)
	defer os.RemoveAll(testDir)

	t.Run("ScanSourceFiles", func(t *testing.T) {
		testScanSourceFiles(t, testDir)
	})

	t.Run("ProcessAssets", func(t *testing.T) {
		testProcessAssets(t, testDir)
	})

	t.Run("GenerateManifest", func(t *testing.T) {
		testGenerateManifest(t, testDir)
	})

	t.Run("CreatePackage", func(t *testing.T) {
		testCreatePackage(t, testDir)
	})

	t.Run("SignDocument", func(t *testing.T) {
		testSignDocument(t, testDir)
	})

	t.Run("CompleteWorkflow", func(t *testing.T) {
		testCompleteBuilderWorkflow(t, testDir)
	})
}

func setupBuilderTestDir(t *testing.T) string {
	// Create temporary directory
	testDir, err := os.MkdirTemp("", "liv-builder-test-*")
	if err != nil {
		t.Fatalf("Failed to create test directory: %v", err)
	}

	// Create test document structure
	contentDir := filepath.Join(testDir, "content")
	if err := os.MkdirAll(contentDir, 0755); err != nil {
		t.Fatalf("Failed to create content directory: %v", err)
	}

	// Create test HTML file
	htmlContent := `<!DOCTYPE html>
<html lang="en">
<head>
    <meta charset="UTF-8">
    <title>Builder Test Document</title>
    <link rel="stylesheet" href="styles/main.css">
</head>
<body>
    <div class="container">
        <h1>Builder Test Document</h1>
        <p>This is a test document for builder testing.</p>
        <canvas id="test-canvas"></canvas>
    </div>
    <script src="scripts/main.js"></script>
</body>
</html>`

	if err := os.WriteFile(filepath.Join(contentDir, "index.html"), []byte(htmlContent), 0644); err != nil {
		t.Fatalf("Failed to create test HTML: %v", err)
	}

	// Create styles directory and CSS file
	stylesDir := filepath.Join(contentDir, "styles")
	if err := os.MkdirAll(stylesDir, 0755); err != nil {
		t.Fatalf("Failed to create styles directory: %v", err)
	}

	cssContent := `body {
    font-family: Arial, sans-serif;
    margin: 0;
    padding: 20px;
    background: #f0f0f0;
}

.container {
    max-width: 800px;
    margin: 0 auto;
    background: white;
    padding: 20px;
    border-radius: 8px;
}

#test-canvas {
    border: 1px solid #ccc;
    width: 100%;
    height: 200px;
}`

	if err := os.WriteFile(filepath.Join(stylesDir, "main.css"), []byte(cssContent), 0644); err != nil {
		t.Fatalf("Failed to create test CSS: %v", err)
	}

	// Create scripts directory and JS file with canvas usage
	scriptsDir := filepath.Join(contentDir, "scripts")
	if err := os.MkdirAll(scriptsDir, 0755); err != nil {
		t.Fatalf("Failed to create scripts directory: %v", err)
	}

	jsContent := `document.addEventListener('DOMContentLoaded', function() {
    const canvas = document.getElementById('test-canvas');
    if (canvas) {
        const ctx = canvas.getContext('2d');
        ctx.fillStyle = '#4CAF50';
        ctx.fillRect(10, 10, 100, 50);
        ctx.fillStyle = '#333';
        ctx.font = '16px Arial';
        ctx.fillText('Builder Test', 20, 35);
    }
    console.log('Builder test document loaded');
});`

	if err := os.WriteFile(filepath.Join(scriptsDir, "main.js"), []byte(jsContent), 0644); err != nil {
		t.Fatalf("Failed to create test JS: %v", err)
	}

	// Create WASM directory and dummy WASM file
	wasmDir := filepath.Join(testDir, "wasm")
	if err := os.MkdirAll(wasmDir, 0755); err != nil {
		t.Fatalf("Failed to create wasm directory: %v", err)
	}

	// Simple WASM binary (minimal valid WASM)
	wasmContent := []byte{0x00, 0x61, 0x73, 0x6d, 0x01, 0x00, 0x00, 0x00}
	if err := os.WriteFile(filepath.Join(wasmDir, "graphics.wasm"), wasmContent, 0644); err != nil {
		t.Fatalf("Failed to create test WASM: %v", err)
	}

	// Generate test key pair
	sigManager := integrity.NewSignatureManager()
	keyPair, err := sigManager.GenerateKeyPair(2048)
	if err != nil {
		t.Fatalf("Failed to generate key pair: %v", err)
	}

	keyPath := filepath.Join(testDir, "test-key.pem")
	if err := sigManager.SavePrivateKeyPEM(keyPair, keyPath); err != nil {
		t.Fatalf("Failed to save private key: %v", err)
	}

	return testDir
}

func testScanSourceFiles(t *testing.T, testDir string) {
	// Test scanning source files
	err := scanSourceFiles(testDir, true)
	if err != nil {
		t.Errorf("scanSourceFiles failed: %v", err)
	}

	// Test with missing required files
	emptyDir, err := os.MkdirTemp("", "empty-test-*")
	if err != nil {
		t.Fatalf("Failed to create empty test directory: %v", err)
	}
	defer os.RemoveAll(emptyDir)

	err = scanSourceFiles(emptyDir, false)
	if err == nil {
		t.Error("Expected error for missing required files, but scanSourceFiles succeeded")
	}
}

func testProcessAssets(t *testing.T, testDir string) {
	// Test processing assets
	err := processAssets(testDir, true, true)
	if err != nil {
		t.Errorf("processAssets failed: %v", err)
	}

	// Test without compression
	err = processAssets(testDir, false, false)
	if err != nil {
		t.Errorf("processAssets without compression failed: %v", err)
	}
}

func testGenerateManifest(t *testing.T, testDir string) {
	// Test generating manifest
	err := generateManifest(testDir, "", true)
	if err != nil {
		t.Errorf("generateManifest failed: %v", err)
	}

	// Check that manifest was created
	manifestPath := filepath.Join(testDir, "manifest.json")
	if _, err := os.Stat(manifestPath); os.IsNotExist(err) {
		t.Error("Manifest file was not created")
	}

	// Validate the generated manifest
	validator := manifest.NewManifestValidator()
	manifestData, err := os.ReadFile(manifestPath)
	if err != nil {
		t.Fatalf("Failed to read generated manifest: %v", err)
	}

	_, result := validator.ValidateManifestJSON(manifestData)
	if !result.IsValid {
		t.Errorf("Generated manifest is invalid: %v", result.Errors)
	}

	// Check that WASM configuration was detected
	manifestStr := string(manifestData)
	if !strings.Contains(manifestStr, "wasm_config") {
		t.Error("WASM configuration not found in generated manifest")
	}

	if !strings.Contains(manifestStr, "graphics") {
		t.Error("WASM module 'graphics' not found in generated manifest")
	}

	// Check that interactive security policy was applied
	if !strings.Contains(manifestStr, "wasm-unsafe-eval") {
		t.Error("Interactive security policy not applied (missing wasm-unsafe-eval in CSP)")
	}
}

func testCreatePackage(t *testing.T, testDir string) {
	// First generate manifest
	err := generateManifest(testDir, "", false)
	if err != nil {
		t.Fatalf("Failed to generate manifest for package test: %v", err)
	}

	// Test creating package
	outputFile := filepath.Join(testDir, "test-package.liv")
	err = createPackage(testDir, outputFile, true)
	if err != nil {
		t.Errorf("createPackage failed: %v", err)
	}

	// Check that package was created
	if _, err := os.Stat(outputFile); os.IsNotExist(err) {
		t.Error("Package file was not created")
	}

	// Validate the package structure
	zipContainer := container.NewZIPContainer()
	files, err := zipContainer.ExtractToMemory(outputFile)
	if err != nil {
		t.Errorf("Failed to extract created package: %v", err)
	}

	// Check required files
	requiredFiles := []string{"manifest.json", "content/index.html"}
	for _, required := range requiredFiles {
		if _, exists := files[required]; !exists {
			t.Errorf("Required file missing from package: %s", required)
		}
	}

	// Check WASM file
	if _, exists := files["wasm/graphics.wasm"]; !exists {
		t.Error("WASM file missing from package")
	}
}

func testSignDocument(t *testing.T, testDir string) {
	// First create a document to sign
	err := generateManifest(testDir, "", false)
	if err != nil {
		t.Fatalf("Failed to generate manifest for sign test: %v", err)
	}

	outputFile := filepath.Join(testDir, "test-sign.liv")
	err = createPackage(testDir, outputFile, false)
	if err != nil {
		t.Fatalf("Failed to create package for sign test: %v", err)
	}

	// Test signing document
	keyPath := filepath.Join(testDir, "test-key.pem")
	err = signDocument(outputFile, keyPath, true)
	if err != nil {
		t.Errorf("signDocument failed: %v", err)
	}

	// Verify the document was signed
	zipContainer := container.NewZIPContainer()
	files, err := zipContainer.ExtractToMemory(outputFile)
	if err != nil {
		t.Errorf("Failed to extract signed document: %v", err)
	}

	// Check that manifest was updated
	manifestData, exists := files["manifest.json"]
	if !exists {
		t.Fatal("Manifest not found in signed document")
	}

	// Parse manifest to check for signature information
	validator := manifest.NewManifestValidator()
	parsedManifest, result := validator.ValidateManifestJSON(manifestData)
	if !result.IsValid {
		t.Errorf("Signed manifest is invalid: %v", result.Errors)
	}

	// Check that modification time was updated
	if parsedManifest.Metadata.Modified.IsZero() {
		t.Error("Modification time was not updated in signed document")
	}

	// Test signing with nonexistent key
	err = signDocument(outputFile, "nonexistent.pem", false)
	if err == nil {
		t.Error("Expected error for nonexistent key file, but signing succeeded")
	}
}

func testCompleteBuilderWorkflow(t *testing.T, testDir string) {
	t.Logf("Testing complete builder workflow")

	outputFile := filepath.Join(testDir, "workflow-test.liv")
	keyPath := filepath.Join(testDir, "test-key.pem")

	// Test complete workflow using runBuilder function
	err := runBuilder(testDir, outputFile, "", true, true, keyPath, true)
	if err != nil {
		t.Errorf("Complete builder workflow failed: %v", err)
	}

	// Verify the final document
	if _, err := os.Stat(outputFile); os.IsNotExist(err) {
		t.Error("Final document was not created")
	}

	// Validate the final document
	zipContainer := container.NewZIPContainer()
	structureResult := zipContainer.ValidateStructure(outputFile)
	if !structureResult.IsValid {
		t.Errorf("Final document structure is invalid: %v", structureResult.Errors)
	}

	// Extract and validate content
	files, err := zipContainer.ExtractToMemory(outputFile)
	if err != nil {
		t.Errorf("Failed to extract final document: %v", err)
	}

	// Check all expected files are present
	expectedFiles := []string{
		"manifest.json",
		"content/index.html",
		"content/styles/main.css",
		"content/scripts/main.js",
		"wasm/graphics.wasm",
	}

	for _, expected := range expectedFiles {
		if _, exists := files[expected]; !exists {
			t.Errorf("Expected file missing from final document: %s", expected)
		}
	}

	// Validate manifest
	manifestData := files["manifest.json"]
	validator := manifest.NewManifestValidator()
	parsedManifest, result := validator.ValidateManifestJSON(manifestData)
	if !result.IsValid {
		t.Errorf("Final manifest is invalid: %v", result.Errors)
	}

	// Check that document has correct features enabled
	if parsedManifest.Features == nil {
		t.Error("Features not set in final manifest")
	} else {
		if !parsedManifest.Features.WebAssembly {
			t.Error("WebAssembly feature not enabled despite WASM module presence")
		}
		if !parsedManifest.Features.Interactivity {
			t.Error("Interactivity feature not enabled despite interactive content")
		}
	}

	// Check WASM configuration
	if parsedManifest.WASMConfig == nil {
		t.Error("WASM configuration not set in final manifest")
	} else {
		if _, exists := parsedManifest.WASMConfig.Modules["graphics"]; !exists {
			t.Error("Graphics WASM module not configured in final manifest")
		}
	}

	t.Logf("Complete builder workflow test passed")
}

// TestBuilderHelperFunctions tests utility functions
func TestBuilderHelperFunctions(t *testing.T) {
	t.Run("GetMimeType", func(t *testing.T) {
		tests := []struct {
			ext      string
			expected string
		}{
			{".html", "text/html"},
			{".css", "text/css"},
			{".js", "application/javascript"},
			{".wasm", "application/wasm"},
			{".png", "image/png"},
			{".unknown", "application/octet-stream"},
		}

		for _, test := range tests {
			result := getMimeType(test.ext)
			if result != test.expected {
				t.Errorf("getMimeType(%s) = %s, expected %s", test.ext, result, test.expected)
			}
		}
	})

	t.Run("FileExists", func(t *testing.T) {
		// Create a temporary file
		tempFile, err := os.CreateTemp("", "test-file-*")
		if err != nil {
			t.Fatalf("Failed to create temp file: %v", err)
		}
		tempFile.Close()
		defer os.Remove(tempFile.Name())

		// Test existing file
		if !fileExists(tempFile.Name()) {
			t.Error("fileExists returned false for existing file")
		}

		// Test nonexistent file
		if fileExists("nonexistent-file-12345") {
			t.Error("fileExists returned true for nonexistent file")
		}
	})

	t.Run("ExtractHTMLTitle", func(t *testing.T) {
		tests := []struct {
			html     string
			expected string
		}{
			{"<html><head><title>Test Title</title></head></html>", "Test Title"},
			{"<HTML><HEAD><TITLE>Uppercase Title</TITLE></HEAD></HTML>", "Uppercase Title"},
			{"<html><head></head></html>", ""},
			{"<title>No Head Tag</title>", "No Head Tag"},
			{"<html>No title tag</html>", ""},
		}

		for _, test := range tests {
			result := extractHTMLTitle(test.html)
			if result != test.expected {
				t.Errorf("extractHTMLTitle(%s) = %s, expected %s", test.html, result, test.expected)
			}
		}
	})
}

// TestBuilderErrorHandling tests error conditions
func TestBuilderErrorHandling(t *testing.T) {
	t.Run("InvalidInputDirectory", func(t *testing.T) {
		err := runBuilder("nonexistent-directory", "output.liv", "", false, false, "", false)
		if err == nil {
			t.Error("Expected error for nonexistent input directory")
		}
	})

	t.Run("SigningWithoutKey", func(t *testing.T) {
		testDir := setupBuilderTestDir(t)
		defer os.RemoveAll(testDir)

		err := runBuilder(testDir, "output.liv", "", false, true, "", false)
		if err == nil {
			t.Error("Expected error for signing without key file")
		}
	})

	t.Run("SigningWithNonexistentKey", func(t *testing.T) {
		testDir := setupBuilderTestDir(t)
		defer os.RemoveAll(testDir)

		err := runBuilder(testDir, "output.liv", "", false, true, "nonexistent.pem", false)
		if err == nil {
			t.Error("Expected error for signing with nonexistent key file")
		}
	})
}