package container

import (
	"bytes"
	"os"
	"path/filepath"
	"strings"
	"testing"
)

func TestZIPContainer_CreateAndExtract(t *testing.T) {
	container := NewZIPContainer()

	// Create test files
	testFiles := map[string][]byte{
		"manifest.json": []byte(`{"version": "1.0", "title": "Test Document"}`),
		"content/index.html": []byte(`<!DOCTYPE html>
<html>
<head><title>Test</title></head>
<body><h1>Hello World</h1></body>
</html>`),
		"content/styles/main.css": []byte(`body { font-family: Arial, sans-serif; }`),
		"assets/images/test.png": []byte("fake-png-data"),
		"assets/data/sample.json": []byte(`{"data": [1, 2, 3]}`),
	}

	// Test creating ZIP in memory
	var buf bytes.Buffer
	err := container.CreateFromFilesToWriter(testFiles, &buf)
	if err != nil {
		t.Fatalf("Failed to create ZIP: %v", err)
	}

	if buf.Len() == 0 {
		t.Fatal("ZIP file is empty")
	}

	// Create temporary file for testing
	tempFile, err := os.CreateTemp("", "test-*.liv")
	if err != nil {
		t.Fatalf("Failed to create temp file: %v", err)
	}
	defer os.Remove(tempFile.Name())

	// Write ZIP data to temp file
	if _, err := tempFile.Write(buf.Bytes()); err != nil {
		t.Fatalf("Failed to write temp file: %v", err)
	}
	tempFile.Close()

	// Test extracting to memory
	extractedFiles, err := container.ExtractToMemory(tempFile.Name())
	if err != nil {
		t.Fatalf("Failed to extract ZIP: %v", err)
	}

	// Verify extracted files
	if len(extractedFiles) != len(testFiles) {
		t.Errorf("Expected %d files, got %d", len(testFiles), len(extractedFiles))
	}

	for path, expectedContent := range testFiles {
		if extractedContent, exists := extractedFiles[path]; exists {
			if !bytes.Equal(extractedContent, expectedContent) {
				t.Errorf("Content mismatch for %s", path)
			}
		} else {
			t.Errorf("File %s not found in extracted files", path)
		}
	}
}

func TestZIPContainer_ValidateStructure(t *testing.T) {
	container := NewZIPContainer()

	tests := []struct {
		name      string
		files     map[string][]byte
		wantValid bool
		wantError string
	}{
		{
			name: "valid structure",
			files: map[string][]byte{
				"manifest.json":      []byte(`{"version": "1.0"}`),
				"content/index.html": []byte(`<html></html>`),
			},
			wantValid: true,
		},
		{
			name: "missing manifest",
			files: map[string][]byte{
				"content/index.html": []byte(`<html></html>`),
			},
			wantValid: false,
			wantError: "required file missing: manifest.json",
		},
		{
			name: "missing content directory",
			files: map[string][]byte{
				"manifest.json": []byte(`{"version": "1.0"}`),
			},
			wantValid: false,
			wantError: "required file missing: content/index.html",
		},
		{
			name: "invalid file path",
			files: map[string][]byte{
				"manifest.json":      []byte(`{"version": "1.0"}`),
				"content/index.html": []byte(`<html></html>`),
				"../evil.exe":        []byte("malicious"),
			},
			wantValid: false,
			wantError: "path contains directory traversal",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			result := container.ValidateStructureFromMemory(tt.files)

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

func TestZIPContainer_CompressionSettings(t *testing.T) {
	// Test different compression levels
	testFiles := map[string][]byte{
		"manifest.json": []byte(`{"version": "1.0", "title": "Test Document"}`),
		"content/index.html": []byte(strings.Repeat("Hello World! ", 1000)), // Compressible content
		"assets/test.png": []byte(strings.Repeat("PNG", 100)), // Simulate binary data
	}

	// Test no compression
	container1 := NewZIPContainer().SetCompressionLevel(0)
	var buf1 bytes.Buffer
	err := container1.CreateFromFilesToWriter(testFiles, &buf1)
	if err != nil {
		t.Fatalf("Failed to create ZIP with no compression: %v", err)
	}

	// Test best compression
	container2 := NewZIPContainer().SetCompressionLevel(9)
	var buf2 bytes.Buffer
	err = container2.CreateFromFilesToWriter(testFiles, &buf2)
	if err != nil {
		t.Fatalf("Failed to create ZIP with best compression: %v", err)
	}

	// Best compression should result in smaller file for compressible content
	if buf2.Len() >= buf1.Len() {
		t.Errorf("Expected best compression (%d bytes) to be smaller than no compression (%d bytes)", buf2.Len(), buf1.Len())
	}
}

func TestZIPContainer_FileInfo(t *testing.T) {
	container := NewZIPContainer()

	testFiles := map[string][]byte{
		"manifest.json":      []byte(`{"version": "1.0"}`),
		"content/index.html": []byte(`<html><body>Test</body></html>`),
		"assets/data.json":   []byte(`{"test": true}`),
	}

	// Create temporary file
	tempFile, err := os.CreateTemp("", "test-*.liv")
	if err != nil {
		t.Fatalf("Failed to create temp file: %v", err)
	}
	defer os.Remove(tempFile.Name())
	tempFile.Close()

	// Create ZIP file
	err = container.CreateFromFiles(testFiles, tempFile.Name())
	if err != nil {
		t.Fatalf("Failed to create ZIP file: %v", err)
	}

	// Get file info
	fileInfos, err := container.GetFileInfo(tempFile.Name())
	if err != nil {
		t.Fatalf("Failed to get file info: %v", err)
	}

	// Verify file info
	if len(fileInfos) != len(testFiles) {
		t.Errorf("Expected %d file infos, got %d", len(testFiles), len(fileInfos))
	}

	for path, expectedContent := range testFiles {
		if info, exists := fileInfos[path]; exists {
			if info.Size != int64(len(expectedContent)) {
				t.Errorf("File %s: expected size %d, got %d", path, len(expectedContent), info.Size)
			}
			if info.Path != path {
				t.Errorf("File %s: expected path %s, got %s", path, path, info.Path)
			}
		} else {
			t.Errorf("File info for %s not found", path)
		}
	}
}

func TestZIPContainer_DirectoryOperations(t *testing.T) {
	container := NewZIPContainer()

	// Create temporary directory structure
	tempDir, err := os.MkdirTemp("", "test-liv-*")
	if err != nil {
		t.Fatalf("Failed to create temp directory: %v", err)
	}
	defer os.RemoveAll(tempDir)

	// Create test files in directory
	testFiles := map[string]string{
		"manifest.json":      `{"version": "1.0"}`,
		"content/index.html": `<html><body>Test</body></html>`,
		"assets/data.json":   `{"test": true}`,
	}

	for path, content := range testFiles {
		fullPath := filepath.Join(tempDir, path)
		if err := os.MkdirAll(filepath.Dir(fullPath), 0755); err != nil {
			t.Fatalf("Failed to create directory: %v", err)
		}
		if err := os.WriteFile(fullPath, []byte(content), 0644); err != nil {
			t.Fatalf("Failed to write file: %v", err)
		}
	}

	// Create ZIP from directory
	tempZip, err := os.CreateTemp("", "test-*.liv")
	if err != nil {
		t.Fatalf("Failed to create temp ZIP file: %v", err)
	}
	defer os.Remove(tempZip.Name())
	tempZip.Close()

	err = container.CreateFromDirectory(tempDir, tempZip.Name())
	if err != nil {
		t.Fatalf("Failed to create ZIP from directory: %v", err)
	}

	// Extract to new directory
	extractDir, err := os.MkdirTemp("", "test-extract-*")
	if err != nil {
		t.Fatalf("Failed to create extract directory: %v", err)
	}
	defer os.RemoveAll(extractDir)

	err = container.ExtractToDirectory(tempZip.Name(), extractDir)
	if err != nil {
		t.Fatalf("Failed to extract ZIP to directory: %v", err)
	}

	// Verify extracted files
	for path, expectedContent := range testFiles {
		extractedPath := filepath.Join(extractDir, path)
		if _, err := os.Stat(extractedPath); os.IsNotExist(err) {
			t.Errorf("Extracted file %s does not exist", path)
			continue
		}

		content, err := os.ReadFile(extractedPath)
		if err != nil {
			t.Errorf("Failed to read extracted file %s: %v", path, err)
			continue
		}

		if string(content) != expectedContent {
			t.Errorf("Content mismatch for %s: expected %s, got %s", path, expectedContent, string(content))
		}
	}
}

func TestDeduplicateFiles(t *testing.T) {
	// Create test files with duplicates
	testFiles := map[string][]byte{
		"file1.txt":     []byte("Hello World"),
		"file2.txt":     []byte("Different content"),
		"duplicate.txt": []byte("Hello World"), // Same as file1.txt
		"unique.txt":    []byte("Unique content"),
	}

	deduplicated, duplicates := DeduplicateFiles(testFiles)

	// Should have 3 unique files
	if len(deduplicated) != 3 {
		t.Errorf("Expected 3 deduplicated files, got %d", len(deduplicated))
	}

	// Should have 1 duplicate
	if len(duplicates) != 1 {
		t.Errorf("Expected 1 duplicate, got %d", len(duplicates))
	}

	// Check that duplicate points to original (lexicographically smallest path is chosen as original)
	if originalPath, exists := duplicates["file1.txt"]; exists {
		if originalPath != "duplicate.txt" {
			t.Errorf("Expected file1.txt to point to duplicate.txt, got %s", originalPath)
		}
	} else {
		t.Error("file1.txt not found in duplicates map")
	}
}

func TestCalculateFileHash(t *testing.T) {
	content := []byte("Hello World")
	hash := CalculateFileHash(content)

	// SHA-256 of "Hello World" should be consistent
	expectedHash := "a591a6d40bf420404a011733cfb7b190d62c65bf0bcda32b57b277d9ad9f146e"
	if hash != expectedHash {
		t.Errorf("Expected hash %s, got %s", expectedHash, hash)
	}

	// Same content should produce same hash
	hash2 := CalculateFileHash(content)
	if hash != hash2 {
		t.Error("Same content produced different hashes")
	}

	// Different content should produce different hash
	differentContent := []byte("Hello World!")
	differentHash := CalculateFileHash(differentContent)
	if hash == differentHash {
		t.Error("Different content produced same hash")
	}
}

func TestZIPContainer_CompressionStats(t *testing.T) {
	container := NewZIPContainer()

	testFiles := map[string][]byte{
		"manifest.json":      []byte(`{"version": "1.0"}`),
		"content/index.html": []byte(`<html><body>Test</body></html>`),
		"compressible.txt":   []byte(strings.Repeat("Hello World! ", 1000)),
		"small.txt":          []byte("Hi"),
	}

	stats, err := container.CompressFiles(testFiles)
	if err != nil {
		t.Fatalf("Failed to get compression stats: %v", err)
	}

	if stats.FileCount != len(testFiles) {
		t.Errorf("Expected file count %d, got %d", len(testFiles), stats.FileCount)
	}

	if stats.OriginalSize <= 0 {
		t.Error("Original size should be positive")
	}

	if stats.CompressedSize <= 0 {
		t.Error("Compressed size should be positive")
	}

	if stats.CompressionRatio <= 0 || stats.CompressionRatio > 1 {
		t.Errorf("Compression ratio should be between 0 and 1, got %f", stats.CompressionRatio)
	}

	// For highly compressible content, compression ratio should be significantly less than 1
	if stats.CompressionRatio > 0.8 {
		t.Errorf("Expected better compression ratio for compressible content, got %f", stats.CompressionRatio)
	}
}

func TestZIPContainer_SecurityValidation(t *testing.T) {
	container := NewZIPContainer()

	// Test files with security issues
	maliciousFiles := map[string][]byte{
		"manifest.json":      []byte(`{"version": "1.0"}`),
		"content/index.html": []byte(`<html></html>`),
		"../../../etc/passwd": []byte("malicious"),
		"C:\\Windows\\System32\\evil.exe": []byte("malicious"),
		"file<with>invalid:chars.txt":     []byte("invalid"),
	}

	result := container.ValidateStructureFromMemory(maliciousFiles)

	if result.IsValid {
		t.Error("Expected validation to fail for malicious files")
	}

	// Should have multiple security-related errors
	securityErrors := 0
	for _, err := range result.Errors {
		if strings.Contains(err, "directory traversal") ||
			strings.Contains(err, "absolute paths") ||
			strings.Contains(err, "invalid character") {
			securityErrors++
		}
	}

	if securityErrors == 0 {
		t.Error("Expected security-related validation errors")
	}
}

func BenchmarkZIPContainer_CreateFromFiles(b *testing.B) {
	container := NewZIPContainer()

	// Create test files of various sizes
	testFiles := map[string][]byte{
		"manifest.json":    []byte(`{"version": "1.0", "title": "Benchmark Test"}`),
		"content/index.html": []byte(strings.Repeat("<p>Hello World!</p>", 1000)),
		"content/styles/main.css": []byte(strings.Repeat("body { color: red; }", 100)),
		"assets/large.txt": []byte(strings.Repeat("Large file content ", 10000)),
	}

	b.ResetTimer()

	for i := 0; i < b.N; i++ {
		var buf bytes.Buffer
		err := container.CreateFromFilesToWriter(testFiles, &buf)
		if err != nil {
			b.Fatalf("Failed to create ZIP: %v", err)
		}
	}
}

func BenchmarkZIPContainer_ExtractToMemory(b *testing.B) {
	container := NewZIPContainer()

	// Create test ZIP file
	testFiles := map[string][]byte{
		"manifest.json":    []byte(`{"version": "1.0"}`),
		"content/index.html": []byte(strings.Repeat("<p>Content</p>", 1000)),
		"assets/data.json": []byte(strings.Repeat(`{"key": "value"}`, 500)),
	}

	tempFile, err := os.CreateTemp("", "benchmark-*.liv")
	if err != nil {
		b.Fatalf("Failed to create temp file: %v", err)
	}
	defer os.Remove(tempFile.Name())
	tempFile.Close()

	err = container.CreateFromFiles(testFiles, tempFile.Name())
	if err != nil {
		b.Fatalf("Failed to create test ZIP: %v", err)
	}

	b.ResetTimer()

	for i := 0; i < b.N; i++ {
		_, err := container.ExtractToMemory(tempFile.Name())
		if err != nil {
			b.Fatalf("Failed to extract ZIP: %v", err)
		}
	}
}