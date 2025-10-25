package viewer

import (
	"context"
	"fmt"
	"io"
	"path/filepath"
	"strings"
	"sync"
	"time"

	"github.com/liv-format/liv/pkg/core"
)

// DocumentLoader handles loading and parsing of .liv documents
type DocumentLoader struct {
	packageManager  core.PackageManager
	securityManager core.SecurityManager
	validator       core.DocumentValidator
	cache           *DocumentCache
	logger          core.Logger
	metrics         core.MetricsCollector
	config          *LoaderConfiguration
}

// LoaderConfiguration holds configuration for the document loader
type LoaderConfiguration struct {
	EnableCaching       bool          `json:"enable_caching"`
	CacheSize          int           `json:"cache_size"`
	CacheExpiry        time.Duration `json:"cache_expiry"`
	MaxDocumentSize    int64         `json:"max_document_size"`
	ValidateSignatures bool          `json:"validate_signatures"`
	StrictValidation   bool          `json:"strict_validation"`
	LoadTimeout        time.Duration `json:"load_timeout"`
}

// DocumentCache provides caching for loaded documents
type DocumentCache struct {
	documents map[string]*CachedDocument
	mutex     sync.RWMutex
	maxSize   int
	expiry    time.Duration
}

// CachedDocument represents a cached document with metadata
type CachedDocument struct {
	Document   *core.LIVDocument
	LoadTime   time.Time
	AccessTime time.Time
	Size       int64
	Hash       string
}

// LoadResult represents the result of a document loading operation
type LoadResult struct {
	Document     *core.LIVDocument
	LoadTime     time.Duration
	FromCache    bool
	Warnings     []string
	SecurityInfo *core.SecurityReport
}

// LoadError represents an error that occurred during document loading
type LoadError struct {
	Type    LoadErrorType
	Message string
	Cause   error
}

// LoadErrorType defines types of loading errors
type LoadErrorType string

const (
	LoadErrorTypeInvalidFile     LoadErrorType = "invalid_file"
	LoadErrorTypeCorrupted       LoadErrorType = "corrupted"
	LoadErrorTypeUnsupported     LoadErrorType = "unsupported"
	LoadErrorTypeSecurity        LoadErrorType = "security"
	LoadErrorTypeTimeout         LoadErrorType = "timeout"
	LoadErrorTypeResourceLimit   LoadErrorType = "resource_limit"
)

func (le *LoadError) Error() string {
	if le.Cause != nil {
		return fmt.Sprintf("%s: %s (caused by: %v)", le.Type, le.Message, le.Cause)
	}
	return fmt.Sprintf("%s: %s", le.Type, le.Message)
}

// NewDocumentLoader creates a new document loader
func NewDocumentLoader(
	packageManager core.PackageManager,
	securityManager core.SecurityManager,
	validator core.DocumentValidator,
	logger core.Logger,
	metrics core.MetricsCollector,
) *DocumentLoader {
	config := &LoaderConfiguration{
		EnableCaching:       true,
		CacheSize:          50,
		CacheExpiry:        30 * time.Minute,
		MaxDocumentSize:    100 * 1024 * 1024, // 100MB
		ValidateSignatures: true,
		StrictValidation:   true,
		LoadTimeout:        30 * time.Second,
	}

	cache := &DocumentCache{
		documents: make(map[string]*CachedDocument),
		maxSize:   config.CacheSize,
		expiry:    config.CacheExpiry,
	}

	return &DocumentLoader{
		packageManager:  packageManager,
		securityManager: securityManager,
		validator:       validator,
		cache:          cache,
		logger:         logger,
		metrics:        metrics,
		config:         config,
	}
}

// LoadDocument loads a .liv document from a reader
func (dl *DocumentLoader) LoadDocument(ctx context.Context, reader io.Reader, filename string) (*LoadResult, error) {
	startTime := time.Now()

	// Create context with timeout
	loadCtx, cancel := context.WithTimeout(ctx, dl.config.LoadTimeout)
	defer cancel()

	// Validate file extension
	if !dl.isValidLIVFile(filename) {
		return nil, &LoadError{
			Type:    LoadErrorTypeInvalidFile,
			Message: fmt.Sprintf("invalid file extension: %s", filepath.Ext(filename)),
		}
	}

	// Check cache first
	if dl.config.EnableCaching {
		if cached := dl.getCachedDocument(filename); cached != nil {
			dl.logger.Debug("document loaded from cache", "filename", filename)
			return &LoadResult{
				Document:  cached.Document,
				LoadTime:  time.Since(startTime),
				FromCache: true,
			}, nil
		}
	}

	// Load document from reader
	document, err := dl.loadDocumentFromReader(loadCtx, reader, filename)
	if err != nil {
		return nil, err
	}

	// Validate document
	validationResult, securityReport, err := dl.validateDocument(loadCtx, document)
	if err != nil {
		return nil, err
	}

	// Cache document if enabled
	if dl.config.EnableCaching {
		dl.cacheDocument(filename, document)
	}

	loadTime := time.Since(startTime)

	// Record metrics
	if dl.metrics != nil {
		dl.metrics.RecordDocumentLoad(int64(len(document.WASMModules)), loadTime.Milliseconds())
	}

	dl.logger.Info("document loaded successfully",
		"filename", filename,
		"load_time", loadTime,
		"from_cache", false,
		"warnings", len(validationResult.Warnings),
	)

	return &LoadResult{
		Document:     document,
		LoadTime:     loadTime,
		FromCache:    false,
		Warnings:     validationResult.Warnings,
		SecurityInfo: securityReport,
	}, nil
}

// LoadDocumentFromFile loads a .liv document from a file path
func (dl *DocumentLoader) LoadDocumentFromFile(ctx context.Context, filePath string) (*LoadResult, error) {
	// This would typically open a file and call LoadDocument
	// For now, we'll return an error indicating this needs file system access
	return nil, &LoadError{
		Type:    LoadErrorTypeUnsupported,
		Message: "file system access not implemented in this version",
	}
}

// ValidateDocument validates a loaded document
func (dl *DocumentLoader) ValidateDocument(document *core.LIVDocument) (*core.ValidationResult, error) {
	if dl.validator == nil {
		return &core.ValidationResult{
			IsValid:  true,
			Errors:   []string{},
			Warnings: []string{"validation skipped: no validator configured"},
		}, nil
	}

	return dl.validator.ValidateDocument(document), nil
}

// GetCacheStats returns statistics about the document cache
func (dl *DocumentLoader) GetCacheStats() map[string]interface{} {
	dl.cache.mutex.RLock()
	defer dl.cache.mutex.RUnlock()

	totalSize := int64(0)
	oldestAccess := time.Now()
	newestAccess := time.Time{}

	for _, cached := range dl.cache.documents {
		totalSize += cached.Size
		if cached.AccessTime.Before(oldestAccess) {
			oldestAccess = cached.AccessTime
		}
		if cached.AccessTime.After(newestAccess) {
			newestAccess = cached.AccessTime
		}
	}

	return map[string]interface{}{
		"cached_documents": len(dl.cache.documents),
		"total_size":       totalSize,
		"max_size":         dl.cache.maxSize,
		"oldest_access":    oldestAccess,
		"newest_access":    newestAccess,
		"cache_enabled":    dl.config.EnableCaching,
	}
}

// ClearCache clears all cached documents
func (dl *DocumentLoader) ClearCache() {
	dl.cache.mutex.Lock()
	defer dl.cache.mutex.Unlock()

	dl.cache.documents = make(map[string]*CachedDocument)
	dl.logger.Info("document cache cleared")
}

// UpdateConfiguration updates the loader configuration
func (dl *DocumentLoader) UpdateConfiguration(config *LoaderConfiguration) error {
	if config == nil {
		return fmt.Errorf("configuration cannot be nil")
	}

	dl.config = config
	dl.cache.maxSize = config.CacheSize
	dl.cache.expiry = config.CacheExpiry

	dl.logger.Info("document loader configuration updated")
	return nil
}

// GetConfiguration returns the current loader configuration
func (dl *DocumentLoader) GetConfiguration() *LoaderConfiguration {
	return dl.config
}

// Helper methods

func (dl *DocumentLoader) isValidLIVFile(filename string) bool {
	ext := strings.ToLower(filepath.Ext(filename))
	return ext == ".liv"
}

func (dl *DocumentLoader) loadDocumentFromReader(ctx context.Context, reader io.Reader, filename string) (*core.LIVDocument, error) {
	// Check document size limit
	limitedReader := &limitedReader{
		reader: reader,
		limit:  dl.config.MaxDocumentSize,
	}

	// Extract document using package manager
	document, err := dl.packageManager.ExtractPackage(ctx, limitedReader)
	if err != nil {
		return nil, &LoadError{
			Type:    LoadErrorTypeCorrupted,
			Message: "failed to extract document package",
			Cause:   err,
		}
	}

	return document, nil
}

func (dl *DocumentLoader) validateDocument(ctx context.Context, document *core.LIVDocument) (*core.ValidationResult, *core.SecurityReport, error) {
	var validationResult *core.ValidationResult
	var securityReport *core.SecurityReport

	// Validate document structure
	if dl.validator != nil {
		validationResult = dl.validator.ValidateDocument(document)
		if !validationResult.IsValid && dl.config.StrictValidation {
			return nil, nil, &LoadError{
				Type:    LoadErrorTypeSecurity,
				Message: fmt.Sprintf("document validation failed: %v", validationResult.Errors),
			}
		}
	}

	// Validate security
	if dl.securityManager != nil {
		securityReport = dl.securityManager.GenerateSecurityReport(document)
		if !securityReport.IsValid && dl.config.ValidateSignatures {
			return nil, nil, &LoadError{
				Type:    LoadErrorTypeSecurity,
				Message: fmt.Sprintf("security validation failed: %v", securityReport.Errors),
			}
		}
	}

	return validationResult, securityReport, nil
}

func (dl *DocumentLoader) getCachedDocument(filename string) *CachedDocument {
	dl.cache.mutex.RLock()
	defer dl.cache.mutex.RUnlock()

	cached, exists := dl.cache.documents[filename]
	if !exists {
		return nil
	}

	// Check if cache entry has expired
	if time.Since(cached.LoadTime) > dl.cache.expiry {
		// Remove expired entry
		go func() {
			dl.cache.mutex.Lock()
			delete(dl.cache.documents, filename)
			dl.cache.mutex.Unlock()
		}()
		return nil
	}

	// Update access time
	cached.AccessTime = time.Now()
	return cached
}

func (dl *DocumentLoader) cacheDocument(filename string, document *core.LIVDocument) {
	dl.cache.mutex.Lock()
	defer dl.cache.mutex.Unlock()

	// Check cache size limit
	if len(dl.cache.documents) >= dl.cache.maxSize {
		dl.evictLRUDocument()
	}

	// Calculate document size (approximate)
	size := dl.calculateDocumentSize(document)

	// Cache the document
	dl.cache.documents[filename] = &CachedDocument{
		Document:   document,
		LoadTime:   time.Now(),
		AccessTime: time.Now(),
		Size:       size,
		Hash:       dl.calculateDocumentHash(document),
	}
}

func (dl *DocumentLoader) evictLRUDocument() {
	var lruKey string
	var lruTime time.Time = time.Now()

	for key, cached := range dl.cache.documents {
		if cached.AccessTime.Before(lruTime) {
			lruTime = cached.AccessTime
			lruKey = key
		}
	}

	if lruKey != "" {
		delete(dl.cache.documents, lruKey)
		dl.logger.Debug("evicted LRU document from cache", "filename", lruKey)
	}
}

func (dl *DocumentLoader) calculateDocumentSize(document *core.LIVDocument) int64 {
	size := int64(0)

	// Calculate content size
	if document.Content != nil {
		size += int64(len(document.Content.HTML))
		size += int64(len(document.Content.CSS))
		size += int64(len(document.Content.InteractiveSpec))
		size += int64(len(document.Content.StaticFallback))
	}

	// Calculate assets size
	if document.Assets != nil {
		for _, data := range document.Assets.Images {
			size += int64(len(data))
		}
		for _, data := range document.Assets.Fonts {
			size += int64(len(data))
		}
		for _, data := range document.Assets.Data {
			size += int64(len(data))
		}
	}

	// Calculate WASM modules size
	for _, data := range document.WASMModules {
		size += int64(len(data))
	}

	return size
}

func (dl *DocumentLoader) calculateDocumentHash(document *core.LIVDocument) string {
	// Simple hash calculation - in production this would use proper hashing
	return fmt.Sprintf("hash_%d", time.Now().UnixNano())
}

// limitedReader wraps an io.Reader to enforce size limits
type limitedReader struct {
	reader io.Reader
	limit  int64
	read   int64
}

func (lr *limitedReader) Read(p []byte) (n int, err error) {
	if lr.read >= lr.limit {
		return 0, &LoadError{
			Type:    LoadErrorTypeResourceLimit,
			Message: fmt.Sprintf("document size exceeds limit of %d bytes", lr.limit),
		}
	}

	maxRead := lr.limit - lr.read
	if int64(len(p)) > maxRead {
		p = p[:maxRead]
	}

	n, err = lr.reader.Read(p)
	lr.read += int64(n)
	return n, err
}