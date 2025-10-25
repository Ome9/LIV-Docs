package utils

import (
	"crypto/rand"
	"crypto/rsa"
	"crypto/x509"
	"encoding/pem"
	"fmt"
	"io/ioutil"
	"os"
	"path/filepath"
	"strings"
	"testing"
	"time"

	"github.com/stretchr/testify/require"

	"github.com/liv-format/liv/pkg/container"
	"github.com/liv-format/liv/pkg/core"
	"github.com/liv-format/liv/pkg/manifest"
)

// TestHelper provides utility functions for tests
type TestHelper struct {
	TempDir string
	t       *testing.T
}

// NewTestHelper creates a new test helper instance
func NewTestHelper(t *testing.T) *TestHelper {
	tempDir, err := ioutil.TempDir("", "liv_test_")
	require.NoError(t, err)

	return &TestHelper{
		TempDir: tempDir,
		t:       t,
	}
}

// Cleanup removes temporary files created during testing
func (h *TestHelper) Cleanup() {
	if h.TempDir != "" {
		os.RemoveAll(h.TempDir)
	}
}

// CreateTempFile creates a temporary file with the given content
func (h *TestHelper) CreateTempFile(name string, content []byte) string {
	filePath := filepath.Join(h.TempDir, name)

	// Create directory if it doesn't exist
	dir := filepath.Dir(filePath)
	err := os.MkdirAll(dir, 0755)
	require.NoError(h.t, err)

	err = ioutil.WriteFile(filePath, content, 0644)
	require.NoError(h.t, err)

	return filePath
}

// CreateTestDocument creates a test document with the given parameters
func (h *TestHelper) CreateTestDocument(title, author, content string) *core.Document {
	return core.NewDocument(
		core.DocumentMetadata{
			Title:       title,
			Author:      author,
			Description: fmt.Sprintf("Test document: %s", title),
			Version:     "1.0",
			Created:     time.Now(),
			Modified:    time.Now(),
			Language:    "en",
		},
		core.DocumentContent{
			HTML: content,
			CSS:  "body { font-family: Arial, sans-serif; }",
		},
	)
}

// CreateTestContainer creates a test container with basic content
func (h *TestHelper) CreateTestContainer(name string) (*container.Container, string) {
	containerPath := filepath.Join(h.TempDir, name)
	container := container.NewContainer(containerPath)

	// Add basic manifest
	manifestObj := &manifest.Manifest{
		Version: "1.0",
		Metadata: &manifest.DocumentMetadata{
			Title:    "Test Container",
			Author:   "Test Helper",
			Created:  time.Now(),
			Modified: time.Now(),
			Version:  "1.0",
			Language: "en",
		},
		Resources: make(map[string]*manifest.Resource),
		Security: &manifest.SecurityPolicy{
			WASMPermissions: &manifest.WASMPermissions{},
			JSPermissions:   &manifest.JSPermissions{},
			NetworkPolicy:   &manifest.NetworkPolicy{},
			StoragePolicy:   &manifest.StoragePolicy{},
		},
		Features: &manifest.FeatureFlags{},
	}

	manifestData, err := manifestObj.MarshalJSON()
	require.NoError(h.t, err)

	err = container.AddFile("manifest.json", manifestData)
	require.NoError(h.t, err)

	// Add basic HTML content
	htmlContent := "<html><head><title>Test</title></head><body><h1>Test Content</h1></body></html>"
	err = container.AddFile("content/index.html", []byte(htmlContent))
	require.NoError(h.t, err)

	return container, containerPath
}

// CreateTestManifest creates a test manifest with the given parameters
func (h *TestHelper) CreateTestManifest(title, author string, resourceCount int) *manifest.Manifest {
	manifestObj := &manifest.Manifest{
		Version: "1.0",
		Metadata: &manifest.DocumentMetadata{
			Title:       title,
			Author:      author,
			Description: fmt.Sprintf("Test manifest: %s", title),
			Created:     time.Now(),
			Modified:    time.Now(),
			Version:     "1.0",
			Language:    "en",
		},
		Resources: make(map[string]*manifest.Resource),
		Security: &manifest.SecurityPolicy{
			WASMPermissions: &manifest.WASMPermissions{},
			JSPermissions:   &manifest.JSPermissions{},
			NetworkPolicy:   &manifest.NetworkPolicy{},
			StoragePolicy:   &manifest.StoragePolicy{},
		},
		Features: &manifest.FeatureFlags{},
	}

	// Add test resources
	for i := 0; i < resourceCount; i++ {
		resourceName := fmt.Sprintf("resource_%d", i)
		manifestObj.Resources[resourceName] = &manifest.Resource{
			Type: "asset",
			Path: fmt.Sprintf("assets/%s.png", resourceName),
			Size: 1024,
		}
	}

	return manifestObj
}

// GenerateTestKeyPair generates a test RSA key pair for signing tests
func (h *TestHelper) GenerateTestKeyPair() (*rsa.PrivateKey, *rsa.PublicKey) {
	privateKey, err := rsa.GenerateKey(rand.Reader, 2048)
	require.NoError(h.t, err)

	return privateKey, &privateKey.PublicKey
}

// SaveKeyPair saves a key pair to PEM files
func (h *TestHelper) SaveKeyPair(privateKey *rsa.PrivateKey, publicKey *rsa.PublicKey) (string, string) {
	// Save private key
	privateKeyBytes, err := x509.MarshalPKCS8PrivateKey(privateKey)
	require.NoError(h.t, err)

	privateKeyPEM := pem.EncodeToMemory(&pem.Block{
		Type:  "PRIVATE KEY",
		Bytes: privateKeyBytes,
	})

	privateKeyPath := h.CreateTempFile("private.pem", privateKeyPEM)

	// Save public key
	publicKeyBytes, err := x509.MarshalPKIXPublicKey(publicKey)
	require.NoError(h.t, err)

	publicKeyPEM := pem.EncodeToMemory(&pem.Block{
		Type:  "PUBLIC KEY",
		Bytes: publicKeyBytes,
	})

	publicKeyPath := h.CreateTempFile("public.pem", publicKeyPEM)

	return privateKeyPath, publicKeyPath
}

// GenerateRandomData generates random data of the specified size
func (h *TestHelper) GenerateRandomData(size int) []byte {
	data := make([]byte, size)
	_, err := rand.Read(data)
	require.NoError(h.t, err)
	return data
}

// GenerateHTMLContent generates HTML content of approximately the specified size
func (h *TestHelper) GenerateHTMLContent(size int) string {
	var builder strings.Builder
	builder.WriteString("<!DOCTYPE html><html><head><title>Generated Content</title></head><body>")

	paragraph := "<p>This is generated test content. It contains various HTML elements to simulate real documents. "
	paragraph += "The content is repeated to reach the desired size for testing purposes. "
	paragraph += "Lorem ipsum dolor sit amet, consectetur adipiscing elit.</p>"

	for builder.Len() < size-len("</body></html>") {
		if builder.Len()+len(paragraph) > size-len("</body></html>") {
			remaining := size - len("</body></html>") - builder.Len()
			if remaining > 0 {
				builder.WriteString(paragraph[:remaining])
			}
			break
		}
		builder.WriteString(paragraph)
	}

	builder.WriteString("</body></html>")
	return builder.String()
}

// CreateMaliciousHTML creates HTML content with potential security issues
func (h *TestHelper) CreateMaliciousHTML() string {
	return `
		<html>
		<head>
			<title>Malicious Test</title>
			<script>alert('XSS');</script>
		</head>
		<body>
			<div onclick="alert('Click XSS')">Click me</div>
			<a href="javascript:alert('Link XSS')">Malicious link</a>
			<img src="javascript:alert('Image XSS')" alt="Malicious image">
		</body>
		</html>
	`
}

// CreateMaliciousCSS creates CSS content with potential security issues
func (h *TestHelper) CreateMaliciousCSS() string {
	return `
		body {
			background: url('javascript:alert("CSS XSS")');
		}
		.malicious {
			behavior: url('malicious.htc');
			-moz-binding: url('malicious.xml#test');
		}
		@import url('javascript:alert("Import XSS")');
	`
}

// AssertFileExists asserts that a file exists at the given path
func (h *TestHelper) AssertFileExists(path string) {
	_, err := os.Stat(path)
	require.NoError(h.t, err, "File should exist: %s", path)
}

// AssertFileNotExists asserts that a file does not exist at the given path
func (h *TestHelper) AssertFileNotExists(path string) {
	_, err := os.Stat(path)
	require.Error(h.t, err, "File should not exist: %s", path)
	require.True(h.t, os.IsNotExist(err), "Error should be 'not exist': %s", path)
}

// AssertFileSize asserts that a file has the expected size
func (h *TestHelper) AssertFileSize(path string, expectedSize int64) {
	info, err := os.Stat(path)
	require.NoError(h.t, err)
	require.Equal(h.t, expectedSize, info.Size(), "File size should match")
}

// ReadTestDataFile reads a file from the test data directory
func (h *TestHelper) ReadTestDataFile(relativePath string) []byte {
	// Assuming test data is in test/data relative to the project root
	dataPath := filepath.Join("..", "data", relativePath)
	content, err := ioutil.ReadFile(dataPath)
	require.NoError(h.t, err, "Should be able to read test data file: %s", relativePath)
	return content
}

// CreateComplexTestDocument creates a complex test document with multiple assets
func (h *TestHelper) CreateComplexTestDocument() (*core.Document, *container.Container, string) {
	// Create document
	doc := core.NewDocument(
		core.DocumentMetadata{
			Title:       "Complex Test Document",
			Author:      "Test Helper",
			Description: "A complex document with multiple assets for testing",
			Version:     "1.0",
			Language:    "en",
			Created:     time.Now(),
			Modified:    time.Now(),
		},
		core.DocumentContent{
			HTML: `
				<html>
				<head>
					<title>Complex Test Document</title>
					<link rel="stylesheet" href="assets/styles.css">
				</head>
				<body>
					<h1>Complex Test Document</h1>
					<img src="assets/images/logo.png" alt="Logo">
					<p>This is a complex document with multiple assets.</p>
					<canvas id="chart"></canvas>
					<script src="modules/chart.wasm"></script>
				</body>
				</html>
			`,
			CSS: `
				body { 
					font-family: 'CustomFont', Arial, sans-serif; 
					background: #f5f5f5;
				}
				h1 { 
					color: #007acc; 
					text-align: center;
				}
				img { 
					max-width: 100%; 
					height: auto;
				}
				#chart { 
					width: 100%; 
					height: 400px; 
					border: 1px solid #ccc;
				}
			`,
		},
	)

	// Create container
	containerPath := filepath.Join(h.TempDir, "complex-test.liv")
	cont := container.NewContainer(containerPath)

	// Get metadata from document
	metadata := doc.GetMetadata()
	content := doc.GetContent()

	// Add manifest
	manifest := &core.Manifest{
		Version: "1.0",
		Metadata: &core.DocumentMetadata{
			Title:       metadata.Title,
			Author:      metadata.Author,
			Description: metadata.Description,
			Created:     metadata.Created,
			Modified:    metadata.Modified,
			Version:     metadata.Version,
			Language:    metadata.Language,
		},
		Security: &core.SecurityPolicy{
			WASMPermissions: &core.WASMPermissions{},
			JSPermissions:   &core.JSPermissions{},
			NetworkPolicy:   &core.NetworkPolicy{},
			StoragePolicy:   &core.StoragePolicy{},
		},
		Resources: map[string]*core.Resource{
			"index.html": {
				Type: "content",
				Path: "content/index.html",
				Size: int64(len(content.HTML)),
				Hash: "0000000000000000000000000000000000000000000000000000000000000000",
			},
			"styles.css": {
				Type: "stylesheet",
				Path: "assets/styles.css",
				Size: int64(len(content.CSS)),
				Hash: "0000000000000000000000000000000000000000000000000000000000000000",
			},
			"logo.png": {
				Type: "image",
				Path: "assets/images/logo.png",
				Size: 1024,
				Hash: "0000000000000000000000000000000000000000000000000000000000000000",
			},
			"font.woff2": {
				Type: "font",
				Path: "assets/fonts/font.woff2",
				Size: 2048,
				Hash: "0000000000000000000000000000000000000000000000000000000000000000",
			},
			"chart.wasm": {
				Type: "wasm",
				Path: "modules/chart.wasm",
				Size: 4096,
				Hash: "0000000000000000000000000000000000000000000000000000000000000000",
			},
		},
	}

	manifestData, err := manifest.MarshalJSON()
	require.NoError(h.t, err)

	// Add files to container
	err = cont.AddFile("manifest.json", manifestData)
	require.NoError(h.t, err)

	err = cont.AddFile("content/index.html", []byte(content.HTML))
	require.NoError(h.t, err)

	err = cont.AddFile("assets/styles.css", []byte(content.CSS))
	require.NoError(h.t, err)

	// Add fake assets
	err = cont.AddFile("assets/images/logo.png", h.GenerateRandomData(1024))
	require.NoError(h.t, err)

	err = cont.AddFile("assets/fonts/font.woff2", h.GenerateRandomData(2048))
	require.NoError(h.t, err)

	err = cont.AddFile("modules/chart.wasm", h.GenerateRandomData(4096))
	require.NoError(h.t, err)

	return doc, cont, containerPath
}

// MeasureExecutionTime measures the execution time of a function
func (h *TestHelper) MeasureExecutionTime(name string, fn func()) time.Duration {
	start := time.Now()
	fn()
	duration := time.Since(start)
	h.t.Logf("%s execution time: %v", name, duration)
	return duration
}

// AssertExecutionTime asserts that a function executes within the expected time
func (h *TestHelper) AssertExecutionTime(name string, maxDuration time.Duration, fn func()) {
	duration := h.MeasureExecutionTime(name, fn)
	require.Less(h.t, duration, maxDuration, "%s should execute within %v", name, maxDuration)
}

// CreateBenchmarkData creates data for benchmark tests
func (h *TestHelper) CreateBenchmarkData(size int) []byte {
	return h.GenerateRandomData(size)
}

// ValidateTestDocument validates a test document and returns any errors
func (h *TestHelper) ValidateTestDocument(doc *core.Document) error {
	return doc.Validate()
}

// ValidateTestManifest validates a test manifest and returns any errors
func (h *TestHelper) ValidateTestManifest(manifest *manifest.Manifest) error {
	return manifest.Validate()
}

// CreateSecurityTestData creates data for security testing
func (h *TestHelper) CreateSecurityTestData() map[string]string {
	return map[string]string{
		"xss_script":     "<script>alert('XSS')</script>",
		"xss_onclick":    "<div onclick=\"alert('XSS')\">Click</div>",
		"xss_href":       "<a href=\"javascript:alert('XSS')\">Link</a>",
		"xss_src":        "<img src=\"javascript:alert('XSS')\">",
		"css_expression": "width: expression(alert('XSS'))",
		"css_javascript": "background: url('javascript:alert(\"XSS\")')",
		"css_import":     "@import url('javascript:alert(\"XSS\")')",
	}
}

// LogTestProgress logs test progress for long-running tests
func (h *TestHelper) LogTestProgress(message string, args ...interface{}) {
	h.t.Logf("[TEST PROGRESS] "+message, args...)
}
