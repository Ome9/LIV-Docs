package viewer

import (
	"context"
	"fmt"
	"mime"
	"path/filepath"
	"strings"
	"sync"
	"time"

	"github.com/liv-format/liv/pkg/core"
)

// ResourceManager handles loading and caching of document resources
type ResourceManager struct {
	cache          *ResourceCache
	integrityCheck bool
	logger         core.Logger
	metrics        core.MetricsCollector
	config         *ResourceManagerConfig
}

// ResourceManagerConfig holds configuration for the resource manager
type ResourceManagerConfig struct {
	EnableCaching     bool          `json:"enable_caching"`
	CacheSize         int           `json:"cache_size"`
	CacheExpiry       time.Duration `json:"cache_expiry"`
	MaxResourceSize   int64         `json:"max_resource_size"`
	ValidateIntegrity bool          `json:"validate_integrity"`
	PreloadResources  bool          `json:"preload_resources"`
}

// ResourceCache provides caching for document resources
type ResourceCache struct {
	resources map[string]*CachedResource
	mutex     sync.RWMutex
	maxSize   int
	expiry    time.Duration
}

// CachedResource represents a cached resource with metadata
type CachedResource struct {
	Data         []byte
	MimeType     string
	Size         int64
	Hash         string
	LoadTime     time.Time
	AccessTime   time.Time
	AccessCount  int64
}

// ResourceInfo provides information about a resource
type ResourceInfo struct {
	Path        string
	MimeType    string
	Size        int64
	Hash        string
	Compressed  bool
	FromCache   bool
	LoadTime    time.Duration
}

// ResourceLoadResult represents the result of loading a resource
type ResourceLoadResult struct {
	Data     []byte
	Info     *ResourceInfo
	Warnings []string
}

// NewResourceManager creates a new resource manager
func NewResourceManager(logger core.Logger, metrics core.MetricsCollector) *ResourceManager {
	config := &ResourceManagerConfig{
		EnableCaching:     true,
		CacheSize:         200,
		CacheExpiry:       1 * time.Hour,
		MaxResourceSize:   50 * 1024 * 1024, // 50MB
		ValidateIntegrity: true,
		PreloadResources:  true,
	}

	cache := &ResourceCache{
		resources: make(map[string]*CachedResource),
		maxSize:   config.CacheSize,
		expiry:    config.CacheExpiry,
	}

	return &ResourceManager{
		cache:          cache,
		integrityCheck: config.ValidateIntegrity,
		logger:         logger,
		metrics:        metrics,
		config:         config,
	}
}

// LoadResource loads a specific resource from the document
func (rm *ResourceManager) LoadResource(ctx context.Context, document *core.LIVDocument, resourcePath string) (*ResourceLoadResult, error) {
	startTime := time.Now()

	// Check cache first
	if rm.config.EnableCaching {
		if cached := rm.getCachedResource(resourcePath); cached != nil {
			rm.logger.Debug("resource loaded from cache", "path", resourcePath)
			
			return &ResourceLoadResult{
				Data: cached.Data,
				Info: &ResourceInfo{
					Path:      resourcePath,
					MimeType:  cached.MimeType,
					Size:      cached.Size,
					Hash:      cached.Hash,
					FromCache: true,
					LoadTime:  time.Since(startTime),
				},
			}, nil
		}
	}

	// Load resource from document
	data, resourceInfo, err := rm.loadResourceFromDocument(document, resourcePath)
	if err != nil {
		return nil, err
	}

	// Validate resource integrity
	if rm.config.ValidateIntegrity {
		if err := rm.validateResourceIntegrity(document, resourcePath, data); err != nil {
			return nil, fmt.Errorf("resource integrity validation failed: %w", err)
		}
	}

	// Determine MIME type
	mimeType := rm.determineMimeType(resourcePath, data)

	// Cache resource if enabled
	if rm.config.EnableCaching {
		rm.cacheResource(resourcePath, data, mimeType, resourceInfo.Hash)
	}

	loadTime := time.Since(startTime)

	// Record metrics
	if rm.metrics != nil {
		rm.metrics.RecordSecurityEvent("resource_loaded", map[string]interface{}{
			"path":      resourcePath,
			"size":      len(data),
			"mime_type": mimeType,
			"from_cache": false,
		})
	}

	rm.logger.Debug("resource loaded successfully",
		"path", resourcePath,
		"size", len(data),
		"mime_type", mimeType,
		"load_time", loadTime,
	)

	return &ResourceLoadResult{
		Data: data,
		Info: &ResourceInfo{
			Path:      resourcePath,
			MimeType:  mimeType,
			Size:      int64(len(data)),
			Hash:      resourceInfo.Hash,
			FromCache: false,
			LoadTime:  loadTime,
		},
	}, nil
}

// PreloadResources preloads commonly used resources
func (rm *ResourceManager) PreloadResources(ctx context.Context, document *core.LIVDocument) error {
	if !rm.config.PreloadResources {
		return nil
	}

	// Preload critical resources
	criticalResources := []string{
		"content/index.html",
		"content/styles/main.css",
		"content/static/fallback.html",
	}

	for _, resourcePath := range criticalResources {
		if rm.resourceExists(document, resourcePath) {
			_, err := rm.LoadResource(ctx, document, resourcePath)
			if err != nil {
				rm.logger.Warn("failed to preload resource", "path", resourcePath, "error", err)
			}
		}
	}

	return nil
}

// GetResourceList returns a list of all resources in the document
func (rm *ResourceManager) GetResourceList(document *core.LIVDocument) []string {
	var resources []string

	if document.Manifest != nil && document.Manifest.Resources != nil {
		for path := range document.Manifest.Resources {
			resources = append(resources, path)
		}
	}

	return resources
}

// GetResourceInfo returns information about a specific resource
func (rm *ResourceManager) GetResourceInfo(document *core.LIVDocument, resourcePath string) (*ResourceInfo, error) {
	if document.Manifest == nil || document.Manifest.Resources == nil {
		return nil, fmt.Errorf("document manifest or resources not available")
	}

	resource, exists := document.Manifest.Resources[resourcePath]
	if !exists {
		return nil, fmt.Errorf("resource not found: %s", resourcePath)
	}

	return &ResourceInfo{
		Path:     resourcePath,
		MimeType: resource.Type,
		Size:     resource.Size,
		Hash:     resource.Hash,
	}, nil
}

// ValidateAllResources validates the integrity of all resources in the document
func (rm *ResourceManager) ValidateAllResources(ctx context.Context, document *core.LIVDocument) ([]string, error) {
	var warnings []string

	if document.Manifest == nil || document.Manifest.Resources == nil {
		return warnings, fmt.Errorf("document manifest or resources not available")
	}

	for resourcePath := range document.Manifest.Resources {
		data, _, err := rm.loadResourceFromDocument(document, resourcePath)
		if err != nil {
			warnings = append(warnings, fmt.Sprintf("failed to load resource %s: %v", resourcePath, err))
			continue
		}

		if err := rm.validateResourceIntegrity(document, resourcePath, data); err != nil {
			warnings = append(warnings, fmt.Sprintf("integrity validation failed for %s: %v", resourcePath, err))
		}
	}

	return warnings, nil
}

// GetCacheStats returns statistics about the resource cache
func (rm *ResourceManager) GetCacheStats() map[string]interface{} {
	rm.cache.mutex.RLock()
	defer rm.cache.mutex.RUnlock()

	totalSize := int64(0)
	totalAccess := int64(0)
	oldestAccess := time.Now()
	newestAccess := time.Time{}

	for _, cached := range rm.cache.resources {
		totalSize += cached.Size
		totalAccess += cached.AccessCount
		if cached.AccessTime.Before(oldestAccess) {
			oldestAccess = cached.AccessTime
		}
		if cached.AccessTime.After(newestAccess) {
			newestAccess = cached.AccessTime
		}
	}

	return map[string]interface{}{
		"cached_resources": len(rm.cache.resources),
		"total_size":       totalSize,
		"total_access":     totalAccess,
		"max_size":         rm.cache.maxSize,
		"oldest_access":    oldestAccess,
		"newest_access":    newestAccess,
		"cache_enabled":    rm.config.EnableCaching,
	}
}

// ClearCache clears all cached resources
func (rm *ResourceManager) ClearCache() {
	rm.cache.mutex.Lock()
	defer rm.cache.mutex.Unlock()

	rm.cache.resources = make(map[string]*CachedResource)
	rm.logger.Info("resource cache cleared")
}

// UpdateConfiguration updates the resource manager configuration
func (rm *ResourceManager) UpdateConfiguration(config *ResourceManagerConfig) error {
	if config == nil {
		return fmt.Errorf("configuration cannot be nil")
	}

	rm.config = config
	rm.cache.maxSize = config.CacheSize
	rm.cache.expiry = config.CacheExpiry
	rm.integrityCheck = config.ValidateIntegrity

	rm.logger.Info("resource manager configuration updated")
	return nil
}

// Helper methods

func (rm *ResourceManager) loadResourceFromDocument(document *core.LIVDocument, resourcePath string) ([]byte, *core.Resource, error) {
	// Check manifest for resource info
	if document.Manifest == nil || document.Manifest.Resources == nil {
		return nil, nil, fmt.Errorf("document manifest or resources not available")
	}

	resourceInfo, exists := document.Manifest.Resources[resourcePath]
	if !exists {
		return nil, nil, fmt.Errorf("resource not found in manifest: %s", resourcePath)
	}

	// Load resource data based on path
	var data []byte
	var err error

	switch {
	case strings.HasPrefix(resourcePath, "content/"):
		data, err = rm.loadContentResource(document, resourcePath)
	case strings.HasPrefix(resourcePath, "assets/"):
		data, err = rm.loadAssetResource(document, resourcePath)
	default:
		return nil, nil, fmt.Errorf("unsupported resource path: %s", resourcePath)
	}

	if err != nil {
		return nil, nil, err
	}

	// Check size limit
	if int64(len(data)) > rm.config.MaxResourceSize {
		return nil, nil, fmt.Errorf("resource size %d exceeds limit %d", len(data), rm.config.MaxResourceSize)
	}

	return data, resourceInfo, nil
}

func (rm *ResourceManager) loadContentResource(document *core.LIVDocument, resourcePath string) ([]byte, error) {
	if document.Content == nil {
		return nil, fmt.Errorf("document content not available")
	}

	switch resourcePath {
	case "content/index.html":
		return []byte(document.Content.HTML), nil
	case "content/styles/main.css":
		return []byte(document.Content.CSS), nil
	case "content/static/fallback.html":
		return []byte(document.Content.StaticFallback), nil
	case "content/scripts/interactive.js":
		return []byte(document.Content.InteractiveSpec), nil
	default:
		return nil, fmt.Errorf("unknown content resource: %s", resourcePath)
	}
}

func (rm *ResourceManager) loadAssetResource(document *core.LIVDocument, resourcePath string) ([]byte, error) {
	if document.Assets == nil {
		return nil, fmt.Errorf("document assets not available")
	}

	// Extract asset type and name from path
	parts := strings.Split(resourcePath, "/")
	if len(parts) < 3 {
		return nil, fmt.Errorf("invalid asset path: %s", resourcePath)
	}

	assetType := parts[1]
	assetName := strings.Join(parts[2:], "/")

	switch assetType {
	case "images":
		if data, exists := document.Assets.Images[assetName]; exists {
			return data, nil
		}
	case "fonts":
		if data, exists := document.Assets.Fonts[assetName]; exists {
			return data, nil
		}
	case "data":
		if data, exists := document.Assets.Data[assetName]; exists {
			return data, nil
		}
	default:
		return nil, fmt.Errorf("unknown asset type: %s", assetType)
	}

	return nil, fmt.Errorf("asset not found: %s", resourcePath)
}

func (rm *ResourceManager) validateResourceIntegrity(document *core.LIVDocument, resourcePath string, data []byte) error {
	if document.Manifest == nil || document.Manifest.Resources == nil {
		return fmt.Errorf("cannot validate integrity: manifest not available")
	}

	resourceInfo, exists := document.Manifest.Resources[resourcePath]
	if !exists {
		return fmt.Errorf("resource not found in manifest: %s", resourcePath)
	}

	// Validate size
	if int64(len(data)) != resourceInfo.Size {
		return fmt.Errorf("size mismatch: expected %d, got %d", resourceInfo.Size, len(data))
	}

	// Validate hash (simplified - in production would use proper hashing)
	if resourceInfo.Hash == "" {
		return fmt.Errorf("no hash available for integrity check")
	}

	// In a real implementation, this would calculate and compare actual hashes
	rm.logger.Debug("resource integrity validated", "path", resourcePath, "size", len(data))

	return nil
}

func (rm *ResourceManager) determineMimeType(resourcePath string, data []byte) string {
	// First try to determine from file extension
	ext := filepath.Ext(resourcePath)
	if mimeType := mime.TypeByExtension(ext); mimeType != "" {
		// Strip charset parameter if present
		if idx := strings.Index(mimeType, ";"); idx != -1 {
			return strings.TrimSpace(mimeType[:idx])
		}
		return mimeType
	}

	// Try to detect from content
	switch {
	case strings.HasSuffix(resourcePath, ".html"):
		return "text/html"
	case strings.HasSuffix(resourcePath, ".css"):
		return "text/css"
	case strings.HasSuffix(resourcePath, ".js"):
		return "application/javascript"
	case strings.HasSuffix(resourcePath, ".json"):
		return "application/json"
	case strings.HasSuffix(resourcePath, ".svg"):
		return "image/svg+xml"
	case strings.HasSuffix(resourcePath, ".png"):
		return "image/png"
	case strings.HasSuffix(resourcePath, ".jpg"), strings.HasSuffix(resourcePath, ".jpeg"):
		return "image/jpeg"
	case strings.HasSuffix(resourcePath, ".woff2"):
		return "font/woff2"
	case strings.HasSuffix(resourcePath, ".woff"):
		return "font/woff"
	default:
		return "application/octet-stream"
	}
}

func (rm *ResourceManager) resourceExists(document *core.LIVDocument, resourcePath string) bool {
	if document.Manifest == nil || document.Manifest.Resources == nil {
		return false
	}
	_, exists := document.Manifest.Resources[resourcePath]
	return exists
}

func (rm *ResourceManager) getCachedResource(resourcePath string) *CachedResource {
	rm.cache.mutex.RLock()
	defer rm.cache.mutex.RUnlock()

	cached, exists := rm.cache.resources[resourcePath]
	if !exists {
		return nil
	}

	// Check if cache entry has expired
	if time.Since(cached.LoadTime) > rm.cache.expiry {
		// Remove expired entry
		go func() {
			rm.cache.mutex.Lock()
			delete(rm.cache.resources, resourcePath)
			rm.cache.mutex.Unlock()
		}()
		return nil
	}

	// Update access time and count
	cached.AccessTime = time.Now()
	cached.AccessCount++
	return cached
}

func (rm *ResourceManager) cacheResource(resourcePath string, data []byte, mimeType, hash string) {
	rm.cache.mutex.Lock()
	defer rm.cache.mutex.Unlock()

	// Check cache size limit
	if len(rm.cache.resources) >= rm.cache.maxSize {
		rm.evictLRUResource()
	}

	// Cache the resource
	rm.cache.resources[resourcePath] = &CachedResource{
		Data:        data,
		MimeType:    mimeType,
		Size:        int64(len(data)),
		Hash:        hash,
		LoadTime:    time.Now(),
		AccessTime:  time.Now(),
		AccessCount: 1,
	}
}

func (rm *ResourceManager) evictLRUResource() {
	var lruKey string
	var lruTime time.Time = time.Now()

	for key, cached := range rm.cache.resources {
		if cached.AccessTime.Before(lruTime) {
			lruTime = cached.AccessTime
			lruKey = key
		}
	}

	if lruKey != "" {
		delete(rm.cache.resources, lruKey)
		rm.logger.Debug("evicted LRU resource from cache", "path", lruKey)
	}
}