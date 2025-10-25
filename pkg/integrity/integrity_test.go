package integrity

import (
	"bytes"
	"fmt"
	"os"
	"path/filepath"
	"strings"
	"testing"

	"github.com/liv-format/liv/pkg/core"
)

func TestResourceHasher_HashBytes(t *testing.T) {
	hasher := NewResourceHasher(SHA256)

	testData := []byte("Hello, World!")
	hash := hasher.HashBytes(testData)

	// SHA-256 of "Hello, World!" should be consistent
	expectedHash := "dffd6021bb2bd5b0af676290809ec3a53191dd81c7f70a4b28688a362182986f"
	if hash != expectedHash {
		t.Errorf("Expected hash %s, got %s", expectedHash, hash)
	}

	// Same data should produce same hash
	hash2 := hasher.HashBytes(testData)
	if hash != hash2 {
		t.Error("Same data produced different hashes")
	}

	// Different data should produce different hash
	differentData := []byte("Hello, World!!")
	differentHash := hasher.HashBytes(differentData)
	if hash == differentHash {
		t.Error("Different data produced same hash")
	}
}

func TestResourceHasher_HashFile(t *testing.T) {
	hasher := NewResourceHasher(SHA256)

	// Create temporary file
	tempFile, err := os.CreateTemp("", "test-*.txt")
	if err != nil {
		t.Fatalf("Failed to create temp file: %v", err)
	}
	defer os.Remove(tempFile.Name())

	testData := []byte("Test file content")
	if _, err := tempFile.Write(testData); err != nil {
		t.Fatalf("Failed to write temp file: %v", err)
	}
	tempFile.Close()

	// Hash file
	hash, err := hasher.HashFile(tempFile.Name())
	if err != nil {
		t.Fatalf("Failed to hash file: %v", err)
	}

	// Should match hash of the content
	expectedHash := hasher.HashBytes(testData)
	if hash != expectedHash {
		t.Errorf("File hash %s doesn't match content hash %s", hash, expectedHash)
	}

	// Test caching - second call should use cache
	hash2, err := hasher.HashFile(tempFile.Name())
	if err != nil {
		t.Fatalf("Failed to hash file second time: %v", err)
	}

	if hash != hash2 {
		t.Error("Cached hash doesn't match original")
	}

	// Verify cache is working
	if hasher.GetCacheSize() != 1 {
		t.Errorf("Expected cache size 1, got %d", hasher.GetCacheSize())
	}
}

func TestResourceHasher_VerifyBytes(t *testing.T) {
	hasher := NewResourceHasher(SHA256)

	testData := []byte("Test data for verification")
	correctHash := hasher.HashBytes(testData)
	incorrectHash := "incorrect_hash"

	// Test correct verification
	if !hasher.VerifyBytes(testData, correctHash) {
		t.Error("Verification failed for correct hash")
	}

	// Test incorrect verification
	if hasher.VerifyBytes(testData, incorrectHash) {
		t.Error("Verification succeeded for incorrect hash")
	}

	// Test case insensitive verification
	upperHash := strings.ToUpper(correctHash)
	if !hasher.VerifyBytes(testData, upperHash) {
		t.Error("Verification failed for uppercase hash")
	}
}

func TestResourceHasher_HashDirectory(t *testing.T) {
	hasher := NewResourceHasher(SHA256)

	// Create temporary directory structure
	tempDir, err := os.MkdirTemp("", "test-hash-dir-*")
	if err != nil {
		t.Fatalf("Failed to create temp directory: %v", err)
	}
	defer os.RemoveAll(tempDir)

	// Create test files
	testFiles := map[string]string{
		"file1.txt":     "Content of file 1",
		"file2.txt":     "Content of file 2",
		"sub/file3.txt": "Content of file 3",
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

	// Hash directory
	hashes, err := hasher.HashDirectory(tempDir)
	if err != nil {
		t.Fatalf("Failed to hash directory: %v", err)
	}

	// Verify all files were hashed
	if len(hashes) != len(testFiles) {
		t.Errorf("Expected %d hashes, got %d", len(testFiles), len(hashes))
	}

	// Verify hash correctness
	for path, content := range testFiles {
		expectedHash := hasher.HashBytes([]byte(content))
		if actualHash, exists := hashes[path]; exists {
			if actualHash != expectedHash {
				t.Errorf("Hash mismatch for %s: expected %s, got %s", path, expectedHash, actualHash)
			}
		} else {
			t.Errorf("Hash not found for file %s", path)
		}
	}
}

func TestBatchHasher_HashFilesParallel(t *testing.T) {
	batchHasher := NewBatchHasher(SHA256, 2)

	// Create temporary files
	tempDir, err := os.MkdirTemp("", "test-batch-*")
	if err != nil {
		t.Fatalf("Failed to create temp directory: %v", err)
	}
	defer os.RemoveAll(tempDir)

	testFiles := map[string]string{
		"file1.txt": "Content 1",
		"file2.txt": "Content 2",
		"file3.txt": "Content 3",
		"file4.txt": "Content 4",
	}

	var filePaths []string
	for filename, content := range testFiles {
		fullPath := filepath.Join(tempDir, filename)
		if err := os.WriteFile(fullPath, []byte(content), 0644); err != nil {
			t.Fatalf("Failed to write file: %v", err)
		}
		filePaths = append(filePaths, fullPath)
	}

	// Hash files in parallel
	hashes, err := batchHasher.HashFilesParallel(filePaths)
	if err != nil {
		t.Fatalf("Failed to hash files in parallel: %v", err)
	}

	// Verify results
	if len(hashes) != len(filePaths) {
		t.Errorf("Expected %d hashes, got %d", len(filePaths), len(hashes))
	}

	// Verify hash correctness
	for _, filePath := range filePaths {
		filename := filepath.Base(filePath)
		expectedContent := testFiles[filename]
		expectedHash := batchHasher.hasher.HashBytes([]byte(expectedContent))

		if actualHash, exists := hashes[filePath]; exists {
			if actualHash != expectedHash {
				t.Errorf("Hash mismatch for %s: expected %s, got %s", filePath, expectedHash, actualHash)
			}
		} else {
			t.Errorf("Hash not found for file %s", filePath)
		}
	}
}

func TestIntegrityValidator_ValidateResources(t *testing.T) {
	validator := NewIntegrityValidator()

	// Create test resources and files
	testFiles := map[string][]byte{
		"file1.txt": []byte("Content 1"),
		"file2.txt": []byte("Content 2"),
		"file3.txt": []byte("Content 3"),
	}

	resources := make(map[string]*core.Resource)
	for path, content := range testFiles {
		hash := validator.hasher.HashBytes(content)
		resources[path] = &core.Resource{
			Hash: hash,
			Size: int64(len(content)),
			Type: "text/plain",
			Path: path,
		}
	}

	// Test valid resources
	result := validator.ValidateResources(resources, testFiles)
	if !result.IsValid {
		t.Errorf("Validation failed for valid resources: %v", result.Errors)
	}

	// Test hash mismatch
	corruptedFiles := make(map[string][]byte)
	for path, content := range testFiles {
		corruptedFiles[path] = content
	}
	corruptedFiles["file1.txt"] = []byte("Corrupted content")

	result = validator.ValidateResources(resources, corruptedFiles)
	if result.IsValid {
		t.Error("Validation succeeded for corrupted files")
	}

	if len(result.Errors) == 0 {
		t.Error("Expected hash mismatch error")
	}

	// Test missing file
	incompleteFiles := map[string][]byte{
		"file1.txt": []byte("Content 1"),
		"file2.txt": []byte("Content 2"),
		// file3.txt is missing
	}

	result = validator.ValidateResources(resources, incompleteFiles)
	if result.IsValid {
		t.Error("Validation succeeded for missing files")
	}

	// Test orphaned file
	extraFiles := make(map[string][]byte)
	for path, content := range testFiles {
		extraFiles[path] = content
	}
	extraFiles["extra.txt"] = []byte("Extra content")

	result = validator.ValidateResources(resources, extraFiles)
	if !result.IsValid {
		t.Error("Validation failed for extra files (should only warn)")
	}

	if len(result.Warnings) == 0 {
		t.Error("Expected warning for orphaned file")
	}
}

func TestIntegrityValidator_ValidateWASMModules(t *testing.T) {
	validator := NewIntegrityValidator()

	// Valid WASM module data (magic number + version)
	validWASMData := []byte{0x00, 0x61, 0x73, 0x6D, 0x01, 0x00, 0x00, 0x00, 0x01, 0x02, 0x03}
	invalidWASMData := []byte{0xFF, 0xFF, 0xFF, 0xFF, 0x01, 0x00, 0x00, 0x00}

	wasmConfig := &core.WASMConfiguration{
		Modules: map[string]*core.WASMModule{
			"test-module": {
				Name:    "test-module",
				Version: "1.0.0",
			},
		},
	}

	wasmModules := map[string][]byte{
		"test-module": validWASMData,
	}

	// Test valid WASM module
	result := validator.ValidateWASMModules(wasmConfig, wasmModules)
	if !result.IsValid {
		t.Errorf("Validation failed for valid WASM module: %v", result.Errors)
	}

	// Test invalid WASM magic number
	wasmModules["test-module"] = invalidWASMData
	result = validator.ValidateWASMModules(wasmConfig, wasmModules)
	if result.IsValid {
		t.Error("Validation succeeded for invalid WASM magic number")
	}

	// Test missing WASM module
	emptyModules := map[string][]byte{}
	result = validator.ValidateWASMModules(wasmConfig, emptyModules)
	if result.IsValid {
		t.Error("Validation succeeded for missing WASM module")
	}

	// Test nil config (should be valid)
	result = validator.ValidateWASMModules(nil, wasmModules)
	if !result.IsValid {
		t.Error("Validation failed for nil WASM config")
	}
}

func TestIntegrityValidator_GenerateResourceManifest(t *testing.T) {
	validator := NewIntegrityValidator()

	testFiles := map[string][]byte{
		"index.html":     []byte("<html><body>Test</body></html>"),
		"styles/main.css": []byte("body { color: red; }"),
		"script.js":      []byte("console.log('test');"),
		"image.png":      []byte("fake-png-data"),
	}

	resources := validator.GenerateResourceManifest(testFiles)

	// Verify all files have resources
	if len(resources) != len(testFiles) {
		t.Errorf("Expected %d resources, got %d", len(testFiles), len(resources))
	}

	// Verify resource properties
	for path, content := range testFiles {
		if resource, exists := resources[path]; exists {
			// Check hash
			expectedHash := validator.hasher.HashBytes(content)
			if resource.Hash != expectedHash {
				t.Errorf("Hash mismatch for %s: expected %s, got %s", path, expectedHash, resource.Hash)
			}

			// Check size
			if resource.Size != int64(len(content)) {
				t.Errorf("Size mismatch for %s: expected %d, got %d", path, len(content), resource.Size)
			}

			// Check path
			if resource.Path != path {
				t.Errorf("Path mismatch for %s: expected %s, got %s", path, path, resource.Path)
			}

			// Check MIME type
			if resource.Type == "" {
				t.Errorf("MIME type not set for %s", path)
			}
		} else {
			t.Errorf("Resource not found for file %s", path)
		}
	}
}

func TestIntegrityValidator_GenerateIntegrityReport(t *testing.T) {
	validator := NewIntegrityValidator()

	// Create test manifest
	manifest := &core.Manifest{
		Resources: map[string]*core.Resource{
			"file1.txt": {
				Hash: validator.hasher.HashBytes([]byte("Content 1")),
				Size: 9,
				Type: "text/plain",
				Path: "file1.txt",
			},
			"file2.txt": {
				Hash: "incorrect_hash",
				Size: 9,
				Type: "text/plain",
				Path: "file2.txt",
			},
		},
	}

	files := map[string][]byte{
		"file1.txt": []byte("Content 1"),
		"file2.txt": []byte("Content 2"),
		"extra.txt": []byte("Extra content"),
	}

	wasmModules := map[string][]byte{}

	report := validator.GenerateIntegrityReport(manifest, files, wasmModules)

	// Should be invalid due to hash mismatch
	if report.Valid {
		t.Error("Report should be invalid due to hash mismatch")
	}

	// Check statistics
	if report.TotalResources != 2 {
		t.Errorf("Expected 2 total resources, got %d", report.TotalResources)
	}

	if report.ValidatedResources != 2 {
		t.Errorf("Expected 2 validated resources, got %d", report.ValidatedResources)
	}

	// Check hash mismatches
	if len(report.HashMismatches) != 1 {
		t.Errorf("Expected 1 hash mismatch, got %d", len(report.HashMismatches))
	}

	// Check orphaned files
	if len(report.OrphanedFiles) != 1 {
		t.Errorf("Expected 1 orphaned file, got %d", len(report.OrphanedFiles))
	}

	if report.OrphanedFiles[0] != "extra.txt" {
		t.Errorf("Expected orphaned file 'extra.txt', got '%s'", report.OrphanedFiles[0])
	}
}

func TestResourceHasher_HashReader(t *testing.T) {
	hasher := NewResourceHasher(SHA256)

	testData := []byte("Test data for reader")
	reader := bytes.NewReader(testData)

	hash, err := hasher.HashReader(reader)
	if err != nil {
		t.Fatalf("Failed to hash reader: %v", err)
	}

	// Should match hash of the data
	expectedHash := hasher.HashBytes(testData)
	if hash != expectedHash {
		t.Errorf("Reader hash %s doesn't match data hash %s", hash, expectedHash)
	}
}

func TestResourceHasher_VerifyFile(t *testing.T) {
	hasher := NewResourceHasher(SHA256)

	// Create temporary file
	tempFile, err := os.CreateTemp("", "test-verify-*.txt")
	if err != nil {
		t.Fatalf("Failed to create temp file: %v", err)
	}
	defer os.Remove(tempFile.Name())

	testData := []byte("Test file for verification")
	if _, err := tempFile.Write(testData); err != nil {
		t.Fatalf("Failed to write temp file: %v", err)
	}
	tempFile.Close()

	correctHash := hasher.HashBytes(testData)
	incorrectHash := "incorrect_hash"

	// Test correct verification
	valid, err := hasher.VerifyFile(tempFile.Name(), correctHash)
	if err != nil {
		t.Fatalf("Failed to verify file: %v", err)
	}
	if !valid {
		t.Error("Verification failed for correct hash")
	}

	// Test incorrect verification
	valid, err = hasher.VerifyFile(tempFile.Name(), incorrectHash)
	if err != nil {
		t.Fatalf("Failed to verify file: %v", err)
	}
	if valid {
		t.Error("Verification succeeded for incorrect hash")
	}
}

func TestResourceHasher_ClearCache(t *testing.T) {
	hasher := NewResourceHasher(SHA256)

	// Add something to cache
	testData := []byte("test")
	hasher.HashBytes(testData)

	// Create and hash a temp file to populate cache
	tempFile, err := os.CreateTemp("", "test-cache-*.txt")
	if err != nil {
		t.Fatalf("Failed to create temp file: %v", err)
	}
	defer os.Remove(tempFile.Name())

	if _, err := tempFile.Write(testData); err != nil {
		t.Fatalf("Failed to write temp file: %v", err)
	}
	tempFile.Close()

	hasher.HashFile(tempFile.Name())

	// Verify cache has entries
	if hasher.GetCacheSize() == 0 {
		t.Error("Cache should have entries")
	}

	// Clear cache
	hasher.ClearCache()

	// Verify cache is empty
	if hasher.GetCacheSize() != 0 {
		t.Errorf("Cache should be empty after clear, got size %d", hasher.GetCacheSize())
	}
}

func BenchmarkResourceHasher_HashBytes(b *testing.B) {
	hasher := NewResourceHasher(SHA256)
	testData := []byte(strings.Repeat("Hello World! ", 1000))

	b.ResetTimer()
	for i := 0; i < b.N; i++ {
		hasher.HashBytes(testData)
	}
}

func BenchmarkResourceHasher_HashFile(b *testing.B) {
	hasher := NewResourceHasher(SHA256)

	// Create test file
	tempFile, err := os.CreateTemp("", "benchmark-*.txt")
	if err != nil {
		b.Fatalf("Failed to create temp file: %v", err)
	}
	defer os.Remove(tempFile.Name())

	testData := []byte(strings.Repeat("Benchmark data ", 10000))
	if _, err := tempFile.Write(testData); err != nil {
		b.Fatalf("Failed to write temp file: %v", err)
	}
	tempFile.Close()

	b.ResetTimer()
	for i := 0; i < b.N; i++ {
		hasher.ClearCache() // Clear cache to avoid caching effects
		_, err := hasher.HashFile(tempFile.Name())
		if err != nil {
			b.Fatalf("Failed to hash file: %v", err)
		}
	}
}

func BenchmarkBatchHasher_HashFilesParallel(b *testing.B) {
	batchHasher := NewBatchHasher(SHA256, 4)

	// Create test files
	tempDir, err := os.MkdirTemp("", "benchmark-batch-*")
	if err != nil {
		b.Fatalf("Failed to create temp directory: %v", err)
	}
	defer os.RemoveAll(tempDir)

	var filePaths []string
	for i := 0; i < 10; i++ {
		filePath := filepath.Join(tempDir, fmt.Sprintf("file%d.txt", i))
		testData := []byte(strings.Repeat(fmt.Sprintf("File %d content ", i), 1000))
		if err := os.WriteFile(filePath, testData, 0644); err != nil {
			b.Fatalf("Failed to write test file: %v", err)
		}
		filePaths = append(filePaths, filePath)
	}

	b.ResetTimer()
	for i := 0; i < b.N; i++ {
		_, err := batchHasher.HashFilesParallel(filePaths)
		if err != nil {
			b.Fatalf("Failed to hash files in parallel: %v", err)
		}
	}
}