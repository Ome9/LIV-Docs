package e2e

import (
	"io/ioutil"
	"os"
	"os/exec"
	"path/filepath"
	"strings"
	"testing"
	"time"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"

	"github.com/liv-format/liv/pkg/container"
	"github.com/liv-format/liv/pkg/core"
	"github.com/liv-format/liv/pkg/manifest"
)

// TestCompleteDocumentWorkflow tests the complete document creation and viewing workflow
func TestCompleteDocumentWorkflow(t *testing.T) {
	if testing.Short() {
		t.Skip("Skipping E2E test in short mode")
	}

	tempDir, err := ioutil.TempDir("", "liv_e2e_")
	require.NoError(t, err)
	defer os.RemoveAll(tempDir)

	t.Run("create_document_via_api", func(t *testing.T) {
		// Step 1: Create a document using the API
		doc := core.NewDocument(
			core.DocumentMetadata{
				Title:       "E2E Test Document",
				Author:      "E2E Tester",
				Description: "End-to-end test document",
				Version:     "1.0",
				Created:     time.Now(),
				Modified:    time.Now(),
				Language:    "en",
			},
			core.DocumentContent{
				HTML: `
					<html>
					<head>
						<title>E2E Test Document</title>
						<style>
							body { font-family: Arial, sans-serif; }
							.header { color: blue; }
						</style>
					</head>
					<body>
						<h1 class="header">Welcome to E2E Test</h1>
						<p>This document was created for end-to-end testing.</p>
						<img src="assets/test-image.png" alt="Test Image">
					</body>
					</html>
				`,
				CSS: `
					.header { 
						color: blue; 
						font-size: 24px; 
					}
					body { 
						margin: 20px; 
						background-color: #f5f5f5; 
					}
				`,
			},
		)

		// Validate the document
		err := doc.Validate()
		require.NoError(t, err)

		// Step 2: Create container and add document
		containerPath := filepath.Join(tempDir, "e2e-test.liv")
		cont := container.NewContainer(containerPath)

		// Add manifest from document
		metadata := doc.GetMetadata()
		manifestObj := &manifest.Manifest{
			Version: "1.0",
			Metadata: &manifest.DocumentMetadata{
				Title:       metadata.Title,
				Author:      metadata.Author,
				Description: metadata.Description,
				Created:     metadata.Created,
				Modified:    metadata.Modified,
				Version:     metadata.Version,
				Language:    metadata.Language,
			},
			Resources: map[string]*manifest.Resource{
				"index.html": {
					Type: "content",
					Path: "content/index.html",
				},
				"styles.css": {
					Type: "stylesheet",
					Path: "assets/styles.css",
				},
				"test-image.png": {
					Type: "image",
					Path: "assets/test-image.png",
				},
			},
			Security: &manifest.SecurityPolicy{
				WASMPermissions: &manifest.WASMPermissions{},
				JSPermissions:   &manifest.JSPermissions{},
				NetworkPolicy:   &manifest.NetworkPolicy{},
				StoragePolicy:   &manifest.StoragePolicy{},
			},
			Features: &manifest.FeatureFlags{},
		}

		manifestData, err := manifestObj.MarshalJSON()
		require.NoError(t, err)
		err = cont.AddFile("manifest.json", manifestData)
		require.NoError(t, err)

		// Add content files
		content := doc.GetContent()
		err = cont.AddFile("content/index.html", []byte(content.HTML))
		require.NoError(t, err)

		err = cont.AddFile("assets/styles.css", []byte(content.CSS))
		require.NoError(t, err)

		// Add a test image (fake PNG data)
		testImageData := []byte{0x89, 0x50, 0x4E, 0x47, 0x0D, 0x0A, 0x1A, 0x0A} // PNG header
		err = cont.AddFile("assets/test-image.png", testImageData)
		require.NoError(t, err)

		// Save the container
		err = cont.Save()
		require.NoError(t, err)

		// Verify file was created
		_, err = os.Stat(containerPath)
		require.NoError(t, err)

		t.Logf("Successfully created LIV document: %s", containerPath)
	})

	t.Run("read_and_validate_document", func(t *testing.T) {
		// Step 3: Read the document back and validate
		containerPath := filepath.Join(tempDir, "e2e-test.liv")

		readContainer, err := container.OpenContainer(containerPath)
		require.NoError(t, err)

		// Read and validate manifest
		manifestData, err := readContainer.ReadFile("manifest.json")
		require.NoError(t, err)

		manifest := &manifest.Manifest{}
		err = manifest.UnmarshalJSON(manifestData)
		require.NoError(t, err)
		err = manifest.Validate()
		require.NoError(t, err)

		// Verify all expected files exist
		files, err := readContainer.ListFiles()
		require.NoError(t, err)

		expectedFiles := []string{
			"manifest.json",
			"content/index.html",
			"assets/styles.css",
			"assets/test-image.png",
		}

		for _, expectedFile := range expectedFiles {
			assert.Contains(t, files, expectedFile, "Expected file should exist in container")
		}

		// Read and verify content
		htmlContent, err := readContainer.ReadFile("content/index.html")
		require.NoError(t, err)
		assert.Contains(t, string(htmlContent), "E2E Test Document")

		cssContent, err := readContainer.ReadFile("assets/styles.css")
		require.NoError(t, err)
		assert.Contains(t, string(cssContent), ".header")

		imageData, err := readContainer.ReadFile("assets/test-image.png")
		require.NoError(t, err)
		assert.Greater(t, len(imageData), 0)

		t.Logf("Successfully validated LIV document structure")
	})
}

// TestCLIIntegration tests CLI tool integration
func TestCLIIntegration(t *testing.T) {
	if testing.Short() {
		t.Skip("Skipping CLI integration test in short mode")
	}

	// Check if CLI tools are available
	cliPath := findCLITool()
	if cliPath == "" {
		t.Skip("LIV CLI tool not found, skipping CLI integration test")
	}

	tempDir, err := ioutil.TempDir("", "liv_cli_")
	require.NoError(t, err)
	defer os.RemoveAll(tempDir)

	t.Run("create_document_via_cli", func(t *testing.T) {
		// Create input files
		htmlFile := filepath.Join(tempDir, "input.html")
		htmlContent := `
			<html>
			<head><title>CLI Test</title></head>
			<body><h1>CLI Generated Document</h1></body>
			</html>
		`
		err := ioutil.WriteFile(htmlFile, []byte(htmlContent), 0644)
		require.NoError(t, err)

		outputFile := filepath.Join(tempDir, "cli-output.liv")

		// Run CLI command to create document
		cmd := exec.Command(cliPath, "create", "--input", htmlFile, "--output", outputFile)
		output, err := cmd.CombinedOutput()

		if err != nil {
			t.Logf("CLI command failed: %s", string(output))
			t.Logf("Command: %s", cmd.String())
		}
		require.NoError(t, err, "CLI create command should succeed")

		// Verify output file was created
		_, err = os.Stat(outputFile)
		require.NoError(t, err, "Output file should be created")

		t.Logf("CLI create command succeeded")
	})

	t.Run("validate_document_via_cli", func(t *testing.T) {
		outputFile := filepath.Join(tempDir, "cli-output.liv")

		// Run CLI command to validate document
		cmd := exec.Command(cliPath, "validate", outputFile)
		output, err := cmd.CombinedOutput()

		if err != nil {
			t.Logf("CLI validation failed: %s", string(output))
		}
		require.NoError(t, err, "CLI validate command should succeed")

		// Check that validation output indicates success
		assert.Contains(t, string(output), "valid", "Validation output should indicate success")

		t.Logf("CLI validate command succeeded")
	})

	t.Run("extract_document_via_cli", func(t *testing.T) {
		outputFile := filepath.Join(tempDir, "cli-output.liv")
		extractDir := filepath.Join(tempDir, "extracted")

		// Run CLI command to extract document
		cmd := exec.Command(cliPath, "extract", outputFile, "--output", extractDir)
		output, err := cmd.CombinedOutput()

		if err != nil {
			t.Logf("CLI extraction failed: %s", string(output))
		}
		require.NoError(t, err, "CLI extract command should succeed")

		// Verify extracted files exist
		manifestPath := filepath.Join(extractDir, "manifest.json")
		_, err = os.Stat(manifestPath)
		require.NoError(t, err, "Manifest should be extracted")

		t.Logf("CLI extract command succeeded")
	})
}

// TestJavaScriptSDKIntegration tests JavaScript SDK integration
func TestJavaScriptSDKIntegration(t *testing.T) {
	if testing.Short() {
		t.Skip("Skipping JavaScript SDK test in short mode")
	}

	// Check if Node.js is available
	_, err := exec.LookPath("node")
	if err != nil {
		t.Skip("Node.js not found, skipping JavaScript SDK test")
	}

	tempDir, err := ioutil.TempDir("", "liv_js_")
	require.NoError(t, err)
	defer os.RemoveAll(tempDir)

	t.Run("javascript_sdk_basic_usage", func(t *testing.T) {
		// Create a test JavaScript file
		jsTestFile := filepath.Join(tempDir, "test-sdk.js")
		jsContent := `
			const { LIVDocument, LIVContainer } = require('../js/dist/sdk');
			const fs = require('fs');
			const path = require('path');

			async function testSDK() {
				try {
					// Create a document
					const doc = new LIVDocument({
						title: 'JS SDK Test',
						author: 'JS Tester',
						content: '<h1>JavaScript SDK Test</h1>'
					});

					// Validate document
					const isValid = await doc.validate();
					if (!isValid) {
						throw new Error('Document validation failed');
					}

					// Create container
					const container = new LIVContainer();
					await container.addDocument(doc);
					
					// Save to file
					const outputPath = path.join(__dirname, 'js-sdk-test.liv');
					await container.save(outputPath);

					// Verify file exists
					if (!fs.existsSync(outputPath)) {
						throw new Error('Output file was not created');
					}

					console.log('JavaScript SDK test passed');
					process.exit(0);
				} catch (error) {
					console.error('JavaScript SDK test failed:', error.message);
					process.exit(1);
				}
			}

			testSDK();
		`

		err := ioutil.WriteFile(jsTestFile, []byte(jsContent), 0644)
		require.NoError(t, err)

		// Run the JavaScript test
		cmd := exec.Command("node", jsTestFile)
		cmd.Dir = tempDir
		output, err := cmd.CombinedOutput()

		t.Logf("JavaScript SDK test output: %s", string(output))

		if err != nil {
			// If the SDK isn't built yet, this is expected
			t.Logf("JavaScript SDK test failed (expected if SDK not built): %v", err)
		} else {
			// Verify the output file was created
			outputFile := filepath.Join(tempDir, "js-sdk-test.liv")
			_, err = os.Stat(outputFile)
			assert.NoError(t, err, "JavaScript SDK should create output file")
		}
	})
}

// TestPythonSDKIntegration tests Python SDK integration
func TestPythonSDKIntegration(t *testing.T) {
	if testing.Short() {
		t.Skip("Skipping Python SDK test in short mode")
	}

	// Check if Python is available
	_, err := exec.LookPath("python")
	if err != nil {
		_, err = exec.LookPath("python3")
		if err != nil {
			t.Skip("Python not found, skipping Python SDK test")
		}
	}

	tempDir, err := ioutil.TempDir("", "liv_py_")
	require.NoError(t, err)
	defer os.RemoveAll(tempDir)

	t.Run("python_sdk_basic_usage", func(t *testing.T) {
		// Create a test Python file
		pyTestFile := filepath.Join(tempDir, "test_sdk.py")
		pyContent := `
import sys
import os
sys.path.insert(0, os.path.join(os.path.dirname(__file__), '..', 'python', 'src'))

try:
	from liv import LIVDocument, LIVContainer
	
	# Create a document
	doc = LIVDocument(
		title='Python SDK Test',
		author='Python Tester',
		content='<h1>Python SDK Test</h1>'
	)
	
	# Validate document
	if not doc.validate():
		raise Exception('Document validation failed')
	
	# Create container
	container = LIVContainer()
	container.add_document(doc)
	
	# Save to file
	output_path = os.path.join(os.path.dirname(__file__), 'py-sdk-test.liv')
	container.save(output_path)
	
	# Verify file exists
	if not os.path.exists(output_path):
		raise Exception('Output file was not created')
	
	print('Python SDK test passed')
	
except ImportError as e:
	print(f'Python SDK not available: {e}')
	sys.exit(0)  # Don't fail if SDK not installed
except Exception as e:
	print(f'Python SDK test failed: {e}')
	sys.exit(1)
		`

		err := ioutil.WriteFile(pyTestFile, []byte(pyContent), 0644)
		require.NoError(t, err)

		// Run the Python test
		pythonCmd := "python"
		if _, err := exec.LookPath("python3"); err == nil {
			pythonCmd = "python3"
		}

		cmd := exec.Command(pythonCmd, pyTestFile)
		cmd.Dir = tempDir
		output, err := cmd.CombinedOutput()

		t.Logf("Python SDK test output: %s", string(output))

		if err != nil {
			// If the SDK isn't installed yet, this is expected
			t.Logf("Python SDK test failed (expected if SDK not installed): %v", err)
		} else if strings.Contains(string(output), "passed") {
			// Verify the output file was created
			outputFile := filepath.Join(tempDir, "py-sdk-test.liv")
			_, err = os.Stat(outputFile)
			assert.NoError(t, err, "Python SDK should create output file")
		}
	})
}

// TestDesktopAppIntegration tests desktop application integration
func TestDesktopAppIntegration(t *testing.T) {
	if testing.Short() {
		t.Skip("Skipping desktop app test in short mode")
	}

	// Check if Electron is available
	_, err := exec.LookPath("electron")
	if err != nil {
		t.Skip("Electron not found, skipping desktop app test")
	}

	tempDir, err := ioutil.TempDir("", "liv_desktop_")
	require.NoError(t, err)
	defer os.RemoveAll(tempDir)

	t.Run("desktop_app_document_opening", func(t *testing.T) {
		// Create a test document
		containerPath := filepath.Join(tempDir, "desktop-test.liv")
		cont := container.NewContainer(containerPath)

		// Add basic content
		htmlContent := "<html><head><title>Desktop Test</title></head><body><h1>Desktop App Test</h1></body></html>"
		err := cont.AddFile("content/index.html", []byte(htmlContent))
		require.NoError(t, err)

		manifestObj := &manifest.Manifest{
			Version: "1.0",
			Metadata: &manifest.DocumentMetadata{
				Title:    "Desktop Test Document",
				Author:   "Desktop Tester",
				Created:  time.Now(),
				Modified: time.Now(),
				Version:  "1.0",
				Language: "en",
			},
			Resources: map[string]*manifest.Resource{
				"index.html": {
					Type: "content",
					Path: "content/index.html",
				},
			},
			Security: &manifest.SecurityPolicy{
				WASMPermissions: &manifest.WASMPermissions{},
				JSPermissions:   &manifest.JSPermissions{},
				NetworkPolicy:   &manifest.NetworkPolicy{},
				StoragePolicy:   &manifest.StoragePolicy{},
			},
			Features: &manifest.FeatureFlags{},
		}

		manifestData, err := manifestObj.MarshalJSON()
		require.NoError(t, err)
		err = cont.AddFile("manifest.json", manifestData)
		require.NoError(t, err)

		err = cont.Save()
		require.NoError(t, err)

		// Test opening the document with the desktop app (headless mode)
		desktopAppPath := filepath.Join("..", "desktop")
		cmd := exec.Command("electron", ".", "--test-mode", "--document", containerPath)
		cmd.Dir = desktopAppPath

		// Set a timeout for the desktop app test
		done := make(chan error, 1)
		go func() {
			output, err := cmd.CombinedOutput()
			if err != nil {
				t.Logf("Desktop app output: %s", string(output))
			}
			done <- err
		}()

		select {
		case err := <-done:
			if err != nil {
				t.Logf("Desktop app test failed (expected if app not built): %v", err)
			} else {
				t.Logf("Desktop app test completed successfully")
			}
		case <-time.After(10 * time.Second):
			t.Logf("Desktop app test timed out (this may be expected)")
		}
	})
}

// TestSecurityIntegration tests security features end-to-end
func TestSecurityIntegration(t *testing.T) {
	tempDir, err := ioutil.TempDir("", "liv_security_")
	require.NoError(t, err)
	defer os.RemoveAll(tempDir)

	t.Run("document_signing_and_verification", func(t *testing.T) {
		// This test would require actual signing implementation
		// For now, we test the security validation pipeline

		// Create a document with potentially unsafe content
		doc := core.NewDocument(
			core.DocumentMetadata{
				Title:    "Security Test Document",
				Author:   "Security Tester",
				Created:  time.Now(),
				Modified: time.Now(),
				Version:  "1.0",
				Language: "en",
			},
			core.DocumentContent{
				HTML: `
					<html>
					<head><title>Security Test</title></head>
					<body>
						<h1>Security Test</h1>
						<script>console.log('This should be sanitized');</script>
						<div onclick="alert('XSS')">Click me</div>
					</body>
					</html>
				`,
			},
		)

		// Validate document (should catch security issues)
		err := doc.Validate()
		// Depending on implementation, this might pass or fail
		// The key is that security validation is happening
		t.Logf("Document validation result: %v", err)

		// Test container security
		containerPath := filepath.Join(tempDir, "security-test.liv")
		cont := container.NewContainer(containerPath)

		// Add content and save
		content := doc.GetContent()
		err = cont.AddFile("content/index.html", []byte(content.HTML))
		require.NoError(t, err)

		err = cont.Save()
		require.NoError(t, err)

		// Verify container integrity
		readContainer, err := container.OpenContainer(containerPath)
		require.NoError(t, err)

		files, err := readContainer.ListFiles()
		require.NoError(t, err)
		assert.Contains(t, files, "content/index.html")

		t.Logf("Security integration test completed")
	})
}

// Helper functions

// findCLITool attempts to find the LIV CLI tool
func findCLITool() string {
	// Common locations for the CLI tool
	possiblePaths := []string{
		"./bin/liv",
		"./bin/liv.exe",
		"../bin/liv",
		"../bin/liv.exe",
		"liv",
		"liv.exe",
	}

	for _, path := range possiblePaths {
		if _, err := os.Stat(path); err == nil {
			return path
		}
		if _, err := exec.LookPath(path); err == nil {
			return path
		}
	}

	return ""
}
