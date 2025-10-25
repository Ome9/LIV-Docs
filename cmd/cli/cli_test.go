package main

import (
	"os"
	"path/filepath"
	"strings"
	"testing"

	"github.com/liv-format/liv/pkg/container"
	"github.com/liv-format/liv/pkg/core"
	"github.com/liv-format/liv/pkg/integrity"
	"github.com/liv-format/liv/pkg/manifest"
)

// TestCLIFunctions tests the CLI functions directly
func TestCLIFunctions(t *testing.T) {
	// Setup test environment
	testDir := setupTestDir(t)
	defer os.RemoveAll(testDir)

	t.Run("ValidateFunction", func(t *testing.T) {
		testValidateFunction(t, testDir)
	})

	t.Run("ConvertFunction", func(t *testing.T) {
		testConvertFunction(t, testDir)
	})

	t.Run("SignFunction", func(t *testing.T) {
		testSignFunction(t, testDir)
	})

	t.Run("ViewFunction", func(t *testing.T) {
		testViewFunction(t, testDir)
	})
}

func setupTestDir(t *testing.T) string {
	// Create temporary directory
	testDir, err := os.MkdirTemp("", "liv-cli-func-test-*")
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
    <title>CLI Function Test</title>
</head>
<body>
    <h1>CLI Function Test</h1>
    <p>Test document for CLI function testing.</p>
</body>
</html>`

	if err := os.WriteFile(filepath.Join(contentDir, "index.html"), []byte(htmlContent), 0644); err != nil {
		t.Fatalf("Failed to create test HTML: %v", err)
	}

	// Create a valid LIV document using the container package
	zipContainer := container.NewZIPContainer()
	
	// Create manifest
	builder := manifest.NewManifestBuilder()
	builder.CreateDefaultMetadata("CLI Function Test", "Test Author")
	builder.CreateDefaultSecurityPolicy()
	builder.CreateDefaultFeatureFlags()
	
	// Add the HTML resource
	builder.AddResource("content/index.html", &core.Resource{
		Hash: "test-hash",
		Size: int64(len(htmlContent)),
		Type: "text/html",
		Path: "content/index.html",
	})
	
	// Save manifest
	manifestPath := filepath.Join(testDir, "manifest.json")
	if err := builder.SaveToFile(manifestPath); err != nil {
		t.Fatalf("Failed to save manifest: %v", err)
	}

	// Create LIV file
	livFile := filepath.Join(testDir, "test.liv")
	if err := zipContainer.CreateFromDirectory(testDir, livFile); err != nil {
		t.Fatalf("Failed to create test LIV file: %v", err)
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

func testValidateFunction(t *testing.T, testDir string) {
	livFile := filepath.Join(testDir, "test.liv")
	
	// Test validation function
	err := runValidate(livFile, false, false)
	if err != nil {
		t.Errorf("Validate function failed: %v", err)
	}

	// Test with signatures check
	err = runValidate(livFile, true, true)
	if err != nil {
		t.Errorf("Validate function with signatures failed: %v", err)
	}
}

func testConvertFunction(t *testing.T, testDir string) {
	livFile := filepath.Join(testDir, "test.liv")
	htmlOutput := filepath.Join(testDir, "converted.html")
	
	// Test HTML conversion
	err := runConvert(livFile, "html", htmlOutput, 90)
	if err != nil {
		t.Errorf("Convert function failed: %v", err)
	}

	// Verify HTML file was created
	if _, err := os.Stat(htmlOutput); os.IsNotExist(err) {
		t.Errorf("Converted HTML file was not created")
	}

	// Check HTML content
	htmlContent, err := os.ReadFile(htmlOutput)
	if err != nil {
		t.Errorf("Failed to read converted HTML: %v", err)
	}

	htmlStr := string(htmlContent)
	if !strings.Contains(htmlStr, "CLI Function Test") {
		t.Errorf("Converted HTML does not contain expected title")
	}

	// Test unsupported format
	err = runConvert(livFile, "unsupported", "test.out", 90)
	if err == nil {
		t.Errorf("Expected error for unsupported format, but conversion succeeded")
	}
}

func testSignFunction(t *testing.T, testDir string) {
	livFile := filepath.Join(testDir, "test.liv")
	keyPath := filepath.Join(testDir, "test-key.pem")
	signedFile := filepath.Join(testDir, "signed.liv")
	
	// Test signing function
	err := runSign(livFile, keyPath, signedFile)
	if err != nil {
		t.Errorf("Sign function failed: %v", err)
	}

	// Verify signed file was created
	if _, err := os.Stat(signedFile); os.IsNotExist(err) {
		t.Errorf("Signed file was not created")
	}

	// Test with nonexistent key file
	err = runSign(livFile, "nonexistent.pem", "test.liv")
	if err == nil {
		t.Errorf("Expected error for nonexistent key file, but signing succeeded")
	}
}

func testViewFunction(t *testing.T, testDir string) {
	livFile := filepath.Join(testDir, "test.liv")
	
	// Test view function (desktop mode)
	err := runView(livFile, 8080, false, false)
	if err != nil {
		t.Errorf("View function failed: %v", err)
	}

	// Test with nonexistent file
	err = runView("nonexistent.liv", 8080, false, false)
	if err == nil {
		t.Errorf("Expected error for nonexistent file, but view succeeded")
	}
}

// TestHelperFunctions tests utility functions
func TestHelperFunctions(t *testing.T) {
	t.Run("FindExecutables", func(t *testing.T) {
		// Test finding builder executable
		_, err := findBuilderExecutable()
		// This might fail in test environment, which is expected
		if err != nil {
			t.Logf("Builder executable not found (expected in test): %v", err)
		}

		// Test finding viewer executable
		_, err = findViewerExecutable()
		// This might fail in test environment, which is expected
		if err != nil {
			t.Logf("Viewer executable not found (expected in test): %v", err)
		}
	})

	t.Run("GetFileContentSafe", func(t *testing.T) {
		files := map[string][]byte{
			"test.txt": []byte("test content"),
		}

		// Test existing file
		content := getFileContentSafe(files, "test.txt")
		if content != "test content" {
			t.Errorf("Expected 'test content', got '%s'", content)
		}

		// Test nonexistent file
		content = getFileContentSafe(files, "nonexistent.txt")
		if content != "" {
			t.Errorf("Expected empty string for nonexistent file, got '%s'", content)
		}
	})
}

// TestCLIErrorCases tests error handling
func TestCLIErrorCases(t *testing.T) {
	t.Run("NonexistentFiles", func(t *testing.T) {
		// Test validate with nonexistent file
		err := runValidate("nonexistent.liv", false, false)
		if err == nil {
			t.Error("Expected error for nonexistent file in validate")
		}

		// Test convert with nonexistent file
		err = runConvert("nonexistent.liv", "html", "output.html", 90)
		if err == nil {
			t.Error("Expected error for nonexistent file in convert")
		}

		// Test sign with nonexistent file
		err = runSign("nonexistent.liv", "key.pem", "output.liv")
		if err == nil {
			t.Error("Expected error for nonexistent file in sign")
		}

		// Test view with nonexistent file
		err = runView("nonexistent.liv", 8080, false, false)
		if err == nil {
			t.Error("Expected error for nonexistent file in view")
		}
	})

	t.Run("InvalidFormats", func(t *testing.T) {
		// Create a temporary valid file for testing
		testDir := setupTestDir(t)
		defer os.RemoveAll(testDir)
		
		livFile := filepath.Join(testDir, "test.liv")

		// Test convert with invalid format
		err := runConvert(livFile, "invalid-format", "output.txt", 90)
		if err == nil {
			t.Error("Expected error for invalid format in convert")
		}
	})
}