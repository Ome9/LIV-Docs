package integrity

import (
	"crypto/sha256"
	"encoding/hex"
	"fmt"
	"hash"
	"io"
	"os"
	"path/filepath"
	"strings"
	"sync"

	"github.com/liv-format/liv/pkg/core"
)

// HashAlgorithm represents supported hash algorithms
type HashAlgorithm string

const (
	SHA256 HashAlgorithm = "sha256"
	SHA512 HashAlgorithm = "sha512"
)

// ResourceHasher handles hashing and verification of resources
type ResourceHasher struct {
	algorithm HashAlgorithm
	mu        sync.RWMutex
	cache     map[string]string // Cache for computed hashes
}

// NewResourceHasher creates a new resource hasher
func NewResourceHasher(algorithm HashAlgorithm) *ResourceHasher {
	return &ResourceHasher{
		algorithm: algorithm,
		cache:     make(map[string]string),
	}
}

// HashBytes computes hash of byte data
func (rh *ResourceHasher) HashBytes(data []byte) string {
	var hasher hash.Hash
	
	switch rh.algorithm {
	case SHA256:
		hasher = sha256.New()
	case SHA512:
		// Could add SHA512 support here
		hasher = sha256.New()
	default:
		hasher = sha256.New()
	}
	
	hasher.Write(data)
	return hex.EncodeToString(hasher.Sum(nil))
}

// HashFile computes hash of a file
func (rh *ResourceHasher) HashFile(filePath string) (string, error) {
	// Check cache first
	rh.mu.RLock()
	if cached, exists := rh.cache[filePath]; exists {
		rh.mu.RUnlock()
		return cached, nil
	}
	rh.mu.RUnlock()

	// Open file
	file, err := os.Open(filePath)
	if err != nil {
		return "", fmt.Errorf("failed to open file %s: %v", filePath, err)
	}
	defer file.Close()

	// Create hasher
	var hasher hash.Hash
	switch rh.algorithm {
	case SHA256:
		hasher = sha256.New()
	case SHA512:
		// Could add SHA512 support here
		hasher = sha256.New()
	default:
		hasher = sha256.New()
	}

	// Hash file content
	if _, err := io.Copy(hasher, file); err != nil {
		return "", fmt.Errorf("failed to hash file %s: %v", filePath, err)
	}

	hashStr := hex.EncodeToString(hasher.Sum(nil))

	// Cache result
	rh.mu.Lock()
	rh.cache[filePath] = hashStr
	rh.mu.Unlock()

	return hashStr, nil
}

// HashReader computes hash from an io.Reader
func (rh *ResourceHasher) HashReader(reader io.Reader) (string, error) {
	var hasher hash.Hash
	
	switch rh.algorithm {
	case SHA256:
		hasher = sha256.New()
	case SHA512:
		// Could add SHA512 support here
		hasher = sha256.New()
	default:
		hasher = sha256.New()
	}

	if _, err := io.Copy(hasher, reader); err != nil {
		return "", fmt.Errorf("failed to hash reader: %v", err)
	}

	return hex.EncodeToString(hasher.Sum(nil)), nil
}

// VerifyBytes verifies that byte data matches expected hash
func (rh *ResourceHasher) VerifyBytes(data []byte, expectedHash string) bool {
	actualHash := rh.HashBytes(data)
	return strings.EqualFold(actualHash, expectedHash)
}

// VerifyFile verifies that a file matches expected hash
func (rh *ResourceHasher) VerifyFile(filePath, expectedHash string) (bool, error) {
	actualHash, err := rh.HashFile(filePath)
	if err != nil {
		return false, err
	}
	
	return strings.EqualFold(actualHash, expectedHash), nil
}

// VerifyReader verifies that reader content matches expected hash
func (rh *ResourceHasher) VerifyReader(reader io.Reader, expectedHash string) (bool, error) {
	actualHash, err := rh.HashReader(reader)
	if err != nil {
		return false, err
	}
	
	return strings.EqualFold(actualHash, expectedHash), nil
}

// HashDirectory computes hashes for all files in a directory
func (rh *ResourceHasher) HashDirectory(dirPath string) (map[string]string, error) {
	hashes := make(map[string]string)
	
	err := filepath.Walk(dirPath, func(path string, info os.FileInfo, err error) error {
		if err != nil {
			return err
		}
		
		// Skip directories
		if info.IsDir() {
			return nil
		}
		
		// Calculate relative path
		relPath, err := filepath.Rel(dirPath, path)
		if err != nil {
			return fmt.Errorf("failed to get relative path: %v", err)
		}
		
		// Normalize path separators
		relPath = filepath.ToSlash(relPath)
		
		// Hash file
		hash, err := rh.HashFile(path)
		if err != nil {
			return fmt.Errorf("failed to hash file %s: %v", relPath, err)
		}
		
		hashes[relPath] = hash
		return nil
	})
	
	return hashes, err
}

// HashFiles computes hashes for a map of file paths to content
func (rh *ResourceHasher) HashFiles(files map[string][]byte) map[string]string {
	hashes := make(map[string]string)
	
	for path, content := range files {
		hashes[path] = rh.HashBytes(content)
	}
	
	return hashes
}

// ClearCache clears the hash cache
func (rh *ResourceHasher) ClearCache() {
	rh.mu.Lock()
	defer rh.mu.Unlock()
	rh.cache = make(map[string]string)
}

// GetCacheSize returns the number of cached hashes
func (rh *ResourceHasher) GetCacheSize() int {
	rh.mu.RLock()
	defer rh.mu.RUnlock()
	return len(rh.cache)
}

// BatchHasher handles batch hashing operations with concurrency
type BatchHasher struct {
	hasher      *ResourceHasher
	concurrency int
}

// NewBatchHasher creates a new batch hasher
func NewBatchHasher(algorithm HashAlgorithm, concurrency int) *BatchHasher {
	if concurrency <= 0 {
		concurrency = 4 // Default concurrency
	}
	
	return &BatchHasher{
		hasher:      NewResourceHasher(algorithm),
		concurrency: concurrency,
	}
}

// HashFilesParallel hashes multiple files in parallel
func (bh *BatchHasher) HashFilesParallel(filePaths []string) (map[string]string, error) {
	results := make(map[string]string)
	errors := make([]error, 0)
	
	// Create channels
	jobs := make(chan string, len(filePaths))
	results_chan := make(chan hashResult, len(filePaths))
	
	// Start workers
	var wg sync.WaitGroup
	for i := 0; i < bh.concurrency; i++ {
		wg.Add(1)
		go func() {
			defer wg.Done()
			for filePath := range jobs {
				hash, err := bh.hasher.HashFile(filePath)
				results_chan <- hashResult{
					path: filePath,
					hash: hash,
					err:  err,
				}
			}
		}()
	}
	
	// Send jobs
	go func() {
		for _, filePath := range filePaths {
			jobs <- filePath
		}
		close(jobs)
	}()
	
	// Wait for workers to finish
	go func() {
		wg.Wait()
		close(results_chan)
	}()
	
	// Collect results
	for result := range results_chan {
		if result.err != nil {
			errors = append(errors, fmt.Errorf("failed to hash %s: %v", result.path, result.err))
		} else {
			results[result.path] = result.hash
		}
	}
	
	// Return error if any hashing failed
	if len(errors) > 0 {
		return results, fmt.Errorf("hashing errors: %v", errors)
	}
	
	return results, nil
}

type hashResult struct {
	path string
	hash string
	err  error
}

// IntegrityValidator validates resource integrity
type IntegrityValidator struct {
	hasher *ResourceHasher
}

// NewIntegrityValidator creates a new integrity validator
func NewIntegrityValidator() *IntegrityValidator {
	return &IntegrityValidator{
		hasher: NewResourceHasher(SHA256),
	}
}

// ValidateResources validates all resources in a manifest
func (iv *IntegrityValidator) ValidateResources(resources map[string]*core.Resource, files map[string][]byte) *core.ValidationResult {
	var errors []string
	var warnings []string
	
	// Check that all resources in manifest exist in files
	for path, resource := range resources {
		if fileData, exists := files[path]; exists {
			// Verify hash
			actualHash := iv.hasher.HashBytes(fileData)
			if !strings.EqualFold(actualHash, resource.Hash) {
				errors = append(errors, fmt.Sprintf("hash mismatch for %s: expected %s, got %s", 
					path, resource.Hash, actualHash))
			}
			
			// Verify size
			actualSize := int64(len(fileData))
			if actualSize != resource.Size {
				errors = append(errors, fmt.Sprintf("size mismatch for %s: expected %d, got %d", 
					path, resource.Size, actualSize))
			}
		} else {
			errors = append(errors, fmt.Sprintf("resource %s referenced in manifest but not found in files", path))
		}
	}
	
	// Check for files not in manifest
	for path := range files {
		if _, exists := resources[path]; !exists {
			warnings = append(warnings, fmt.Sprintf("file %s found but not referenced in manifest", path))
		}
	}
	
	return &core.ValidationResult{
		IsValid:  len(errors) == 0,
		Errors:   errors,
		Warnings: warnings,
	}
}

// ValidateWASMModules validates WASM module integrity
func (iv *IntegrityValidator) ValidateWASMModules(wasmConfig *core.WASMConfiguration, wasmModules map[string][]byte) *core.ValidationResult {
	var errors []string
	var warnings []string
	
	if wasmConfig == nil {
		return &core.ValidationResult{
			IsValid:  true,
			Errors:   errors,
			Warnings: warnings,
		}
	}
	
	// Check that all configured modules exist
	for moduleName, moduleConfig := range wasmConfig.Modules {
		if moduleData, exists := wasmModules[moduleName]; exists {
			// Validate WASM magic number
			if len(moduleData) < 4 || 
				moduleData[0] != 0x00 || moduleData[1] != 0x61 || 
				moduleData[2] != 0x73 || moduleData[3] != 0x6D {
				errors = append(errors, fmt.Sprintf("WASM module %s has invalid magic number", moduleName))
			}
			
			// Validate WASM version
			if len(moduleData) < 8 ||
				moduleData[4] != 0x01 || moduleData[5] != 0x00 ||
				moduleData[6] != 0x00 || moduleData[7] != 0x00 {
				errors = append(errors, fmt.Sprintf("WASM module %s has unsupported version", moduleName))
			}
			
			// Check module size limits
			if len(moduleData) > 10*1024*1024 { // 10MB limit
				warnings = append(warnings, fmt.Sprintf("WASM module %s is very large (%d bytes)", 
					moduleName, len(moduleData)))
			}
			
		} else {
			errors = append(errors, fmt.Sprintf("WASM module %s configured but not found", moduleName))
		}
		
		// Validate module configuration
		if moduleConfig.Name != moduleName {
			errors = append(errors, fmt.Sprintf("WASM module name mismatch: config says %s, key is %s", 
				moduleConfig.Name, moduleName))
		}
	}
	
	// Check for unconfigured modules
	for moduleName := range wasmModules {
		if _, exists := wasmConfig.Modules[moduleName]; !exists {
			warnings = append(warnings, fmt.Sprintf("WASM module %s found but not configured", moduleName))
		}
	}
	
	return &core.ValidationResult{
		IsValid:  len(errors) == 0,
		Errors:   errors,
		Warnings: warnings,
	}
}

// GenerateResourceManifest generates resource entries for a manifest
func (iv *IntegrityValidator) GenerateResourceManifest(files map[string][]byte) map[string]*core.Resource {
	resources := make(map[string]*core.Resource)
	
	for path, content := range files {
		hash := iv.hasher.HashBytes(content)
		mimeType := iv.detectMimeType(path)
		
		resources[path] = &core.Resource{
			Hash: hash,
			Size: int64(len(content)),
			Type: mimeType,
			Path: path,
		}
	}
	
	return resources
}

// UpdateResourceManifest updates existing resource manifest with new hashes
func (iv *IntegrityValidator) UpdateResourceManifest(resources map[string]*core.Resource, files map[string][]byte) {
	for path, content := range files {
		if resource, exists := resources[path]; exists {
			// Update hash and size
			resource.Hash = iv.hasher.HashBytes(content)
			resource.Size = int64(len(content))
			
			// Update MIME type if not set
			if resource.Type == "" {
				resource.Type = iv.detectMimeType(path)
			}
		}
	}
}

// detectMimeType detects MIME type from file extension
func (iv *IntegrityValidator) detectMimeType(filePath string) string {
	ext := strings.ToLower(filepath.Ext(filePath))
	
	mimeTypes := map[string]string{
		".html": "text/html",
		".htm":  "text/html",
		".css":  "text/css",
		".js":   "application/javascript",
		".json": "application/json",
		".xml":  "application/xml",
		".txt":  "text/plain",
		".md":   "text/markdown",
		
		// Images
		".png":  "image/png",
		".jpg":  "image/jpeg",
		".jpeg": "image/jpeg",
		".gif":  "image/gif",
		".svg":  "image/svg+xml",
		".webp": "image/webp",
		".ico":  "image/x-icon",
		
		// Fonts
		".woff":  "font/woff",
		".woff2": "font/woff2",
		".ttf":   "font/ttf",
		".otf":   "font/otf",
		".eot":   "application/vnd.ms-fontobject",
		
		// Audio/Video
		".mp3":  "audio/mpeg",
		".wav":  "audio/wav",
		".ogg":  "audio/ogg",
		".mp4":  "video/mp4",
		".webm": "video/webm",
		
		// Data
		".csv":  "text/csv",
		".pdf":  "application/pdf",
		".zip":  "application/zip",
		
		// WASM
		".wasm": "application/wasm",
	}
	
	if mimeType, exists := mimeTypes[ext]; exists {
		return mimeType
	}
	
	return "application/octet-stream"
}

// IntegrityReport contains integrity validation results
type IntegrityReport struct {
	Valid             bool                    `json:"valid"`
	TotalResources    int                     `json:"total_resources"`
	ValidatedResources int                    `json:"validated_resources"`
	HashMismatches    []HashMismatch          `json:"hash_mismatches"`
	SizeMismatches    []SizeMismatch          `json:"size_mismatches"`
	MissingResources  []string                `json:"missing_resources"`
	OrphanedFiles     []string                `json:"orphaned_files"`
	WASMValidation    *core.ValidationResult  `json:"wasm_validation"`
}

// HashMismatch represents a hash validation failure
type HashMismatch struct {
	Path         string `json:"path"`
	ExpectedHash string `json:"expected_hash"`
	ActualHash   string `json:"actual_hash"`
}

// SizeMismatch represents a size validation failure
type SizeMismatch struct {
	Path         string `json:"path"`
	ExpectedSize int64  `json:"expected_size"`
	ActualSize   int64  `json:"actual_size"`
}

// GenerateIntegrityReport generates a comprehensive integrity report
func (iv *IntegrityValidator) GenerateIntegrityReport(manifest *core.Manifest, files map[string][]byte, wasmModules map[string][]byte) *IntegrityReport {
	report := &IntegrityReport{
		Valid:             true,
		TotalResources:    len(manifest.Resources),
		ValidatedResources: 0,
		HashMismatches:    []HashMismatch{},
		SizeMismatches:    []SizeMismatch{},
		MissingResources:  []string{},
		OrphanedFiles:     []string{},
	}
	
	// Validate resources
	for path, resource := range manifest.Resources {
		if fileData, exists := files[path]; exists {
			report.ValidatedResources++
			
			// Check hash
			actualHash := iv.hasher.HashBytes(fileData)
			if !strings.EqualFold(actualHash, resource.Hash) {
				report.Valid = false
				report.HashMismatches = append(report.HashMismatches, HashMismatch{
					Path:         path,
					ExpectedHash: resource.Hash,
					ActualHash:   actualHash,
				})
			}
			
			// Check size
			actualSize := int64(len(fileData))
			if actualSize != resource.Size {
				report.Valid = false
				report.SizeMismatches = append(report.SizeMismatches, SizeMismatch{
					Path:         path,
					ExpectedSize: resource.Size,
					ActualSize:   actualSize,
				})
			}
		} else {
			report.Valid = false
			report.MissingResources = append(report.MissingResources, path)
		}
	}
	
	// Check for orphaned files
	for path := range files {
		if _, exists := manifest.Resources[path]; !exists {
			report.OrphanedFiles = append(report.OrphanedFiles, path)
		}
	}
	
	// Validate WASM modules
	report.WASMValidation = iv.ValidateWASMModules(manifest.WASMConfig, wasmModules)
	if !report.WASMValidation.IsValid {
		report.Valid = false
	}
	
	return report
}