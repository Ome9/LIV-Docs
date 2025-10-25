package viewer

import (
	"context"
	"testing"
	"time"

	"github.com/liv-format/liv/pkg/core"
)

func TestNewResourceManager(t *testing.T) {
	logger := &MockLogger{}
	metrics := &MockMetricsCollector{}

	rm := NewResourceManager(logger, metrics)

	if rm == nil {
		t.Fatal("NewResourceManager returned nil")
	}

	if rm.logger != logger {
		t.Error("logger not set correctly")
	}

	if rm.metrics != metrics {
		t.Error("metrics collector not set correctly")
	}

	if rm.config == nil {
		t.Error("configuration not initialized")
	}

	if rm.cache == nil {
		t.Error("cache not initialized")
	}
}

func TestResourceManager_LoadResource_ValidResource(t *testing.T) {
	logger := &MockLogger{}
	metrics := &MockMetricsCollector{}

	rm := NewResourceManager(logger, metrics)
	document := createTestDocument()

	ctx := context.Background()
	result, err := rm.LoadResource(ctx, document, "content/index.html")

	if err != nil {
		t.Fatalf("LoadResource failed: %v", err)
	}

	if result == nil {
		t.Fatal("LoadResource returned nil result")
	}

	if result.Data == nil {
		t.Error("result should contain data")
	}

	if result.Info == nil {
		t.Error("result should contain resource info")
	}

	if result.Info.Path != "content/index.html" {
		t.Errorf("expected path 'content/index.html', got '%s'", result.Info.Path)
	}

	if result.Info.MimeType != "text/html" {
		t.Errorf("expected MIME type 'text/html', got '%s'", result.Info.MimeType)
	}
}

func TestResourceManager_LoadResource_InvalidResource(t *testing.T) {
	logger := &MockLogger{}
	metrics := &MockMetricsCollector{}

	rm := NewResourceManager(logger, metrics)
	document := createTestDocument()

	ctx := context.Background()
	_, err := rm.LoadResource(ctx, document, "nonexistent/resource.txt")

	if err == nil {
		t.Error("LoadResource should fail for nonexistent resource")
	}
}

func TestResourceManager_LoadResource_Caching(t *testing.T) {
	logger := &MockLogger{}
	metrics := &MockMetricsCollector{}

	rm := NewResourceManager(logger, metrics)
	document := createTestDocument()

	ctx := context.Background()

	// First load
	result1, err := rm.LoadResource(ctx, document, "content/index.html")
	if err != nil {
		t.Fatalf("first LoadResource failed: %v", err)
	}

	if result1.Info.FromCache {
		t.Error("first load should not be from cache")
	}

	// Second load (should be from cache)
	result2, err := rm.LoadResource(ctx, document, "content/index.html")
	if err != nil {
		t.Fatalf("second LoadResource failed: %v", err)
	}

	if !result2.Info.FromCache {
		t.Error("second load should be from cache")
	}

	// Data should be the same
	if string(result1.Data) != string(result2.Data) {
		t.Error("cached data should match original data")
	}
}

func TestResourceManager_PreloadResources(t *testing.T) {
	logger := &MockLogger{}
	metrics := &MockMetricsCollector{}

	rm := NewResourceManager(logger, metrics)
	document := createTestDocument()

	ctx := context.Background()
	err := rm.PreloadResources(ctx, document)

	if err != nil {
		t.Errorf("PreloadResources failed: %v", err)
	}

	// Check that resources were cached
	stats := rm.GetCacheStats()
	cachedCount := stats["cached_resources"].(int)
	if cachedCount == 0 {
		t.Error("preloading should cache some resources")
	}
}

func TestResourceManager_GetResourceList(t *testing.T) {
	logger := &MockLogger{}
	metrics := &MockMetricsCollector{}

	rm := NewResourceManager(logger, metrics)
	document := createTestDocument()

	resources := rm.GetResourceList(document)

	if len(resources) == 0 {
		t.Error("should return list of resources")
	}

	// Check for expected resource
	found := false
	for _, resource := range resources {
		if resource == "content/index.html" {
			found = true
			break
		}
	}

	if !found {
		t.Error("expected resource 'content/index.html' not found in list")
	}
}

func TestResourceManager_GetResourceInfo(t *testing.T) {
	logger := &MockLogger{}
	metrics := &MockMetricsCollector{}

	rm := NewResourceManager(logger, metrics)
	document := createTestDocument()

	info, err := rm.GetResourceInfo(document, "content/index.html")

	if err != nil {
		t.Errorf("GetResourceInfo failed: %v", err)
	}

	if info == nil {
		t.Fatal("GetResourceInfo returned nil info")
	}

	if info.Path != "content/index.html" {
		t.Errorf("expected path 'content/index.html', got '%s'", info.Path)
	}

	if info.MimeType != "text/html" {
		t.Errorf("expected MIME type 'text/html', got '%s'", info.MimeType)
	}

	// Test nonexistent resource
	_, err = rm.GetResourceInfo(document, "nonexistent/resource.txt")
	if err == nil {
		t.Error("GetResourceInfo should fail for nonexistent resource")
	}
}

func TestResourceManager_ValidateAllResources(t *testing.T) {
	logger := &MockLogger{}
	metrics := &MockMetricsCollector{}

	rm := NewResourceManager(logger, metrics)
	document := createTestDocument()

	ctx := context.Background()
	warnings, err := rm.ValidateAllResources(ctx, document)

	if err != nil {
		t.Errorf("ValidateAllResources failed: %v", err)
	}

	// Should not have warnings for valid document
	if len(warnings) > 0 {
		t.Logf("Validation warnings: %v", warnings)
	}
}

func TestResourceManager_GetCacheStats(t *testing.T) {
	logger := &MockLogger{}
	metrics := &MockMetricsCollector{}

	rm := NewResourceManager(logger, metrics)

	// Initially empty cache
	stats := rm.GetCacheStats()
	if stats["cached_resources"].(int) != 0 {
		t.Error("cache should be empty initially")
	}

	// Load a resource to populate cache
	document := createTestDocument()
	ctx := context.Background()
	_, err := rm.LoadResource(ctx, document, "content/index.html")
	if err != nil {
		t.Fatalf("LoadResource failed: %v", err)
	}

	// Check cache stats after loading
	stats = rm.GetCacheStats()
	if stats["cached_resources"].(int) != 1 {
		t.Error("cache should contain 1 resource")
	}

	if !stats["cache_enabled"].(bool) {
		t.Error("cache should be enabled")
	}
}

func TestResourceManager_ClearCache(t *testing.T) {
	logger := &MockLogger{}
	metrics := &MockMetricsCollector{}

	rm := NewResourceManager(logger, metrics)

	// Load a resource to populate cache
	document := createTestDocument()
	ctx := context.Background()
	_, err := rm.LoadResource(ctx, document, "content/index.html")
	if err != nil {
		t.Fatalf("LoadResource failed: %v", err)
	}

	// Verify cache has content
	stats := rm.GetCacheStats()
	if stats["cached_resources"].(int) != 1 {
		t.Error("cache should contain 1 resource before clearing")
	}

	// Clear cache
	rm.ClearCache()

	// Verify cache is empty
	stats = rm.GetCacheStats()
	if stats["cached_resources"].(int) != 0 {
		t.Error("cache should be empty after clearing")
	}
}

func TestResourceManager_UpdateConfiguration(t *testing.T) {
	logger := &MockLogger{}
	metrics := &MockMetricsCollector{}

	rm := NewResourceManager(logger, metrics)

	newConfig := &ResourceManagerConfig{
		EnableCaching:     false,
		CacheSize:         500,
		CacheExpiry:       2 * time.Hour,
		MaxResourceSize:   100 * 1024 * 1024,
		ValidateIntegrity: false,
		PreloadResources:  false,
	}

	err := rm.UpdateConfiguration(newConfig)
	if err != nil {
		t.Errorf("UpdateConfiguration failed: %v", err)
	}

	if rm.config.EnableCaching != newConfig.EnableCaching {
		t.Error("configuration not updated correctly")
	}

	if rm.config.CacheSize != newConfig.CacheSize {
		t.Error("cache size not updated correctly")
	}

	// Test nil configuration
	err = rm.UpdateConfiguration(nil)
	if err == nil {
		t.Error("UpdateConfiguration should fail for nil config")
	}
}

func TestResourceManager_LoadContentResource(t *testing.T) {
	logger := &MockLogger{}
	metrics := &MockMetricsCollector{}

	rm := NewResourceManager(logger, metrics)
	document := createTestDocument()

	tests := []struct {
		path     string
		expected string
	}{
		{"content/index.html", document.Content.HTML},
		{"content/styles/main.css", document.Content.CSS},
		{"content/static/fallback.html", document.Content.StaticFallback},
	}

	for _, tt := range tests {
		t.Run(tt.path, func(t *testing.T) {
			data, err := rm.loadContentResource(document, tt.path)
			if err != nil {
				t.Errorf("loadContentResource failed for %s: %v", tt.path, err)
			}

			if string(data) != tt.expected {
				t.Errorf("expected content '%s', got '%s'", tt.expected, string(data))
			}
		})
	}

	// Test unknown content resource
	_, err := rm.loadContentResource(document, "content/unknown.txt")
	if err == nil {
		t.Error("loadContentResource should fail for unknown resource")
	}
}

func TestResourceManager_LoadAssetResource(t *testing.T) {
	logger := &MockLogger{}
	metrics := &MockMetricsCollector{}

	rm := NewResourceManager(logger, metrics)
	document := createTestDocument()

	// Test valid asset
	data, err := rm.loadAssetResource(document, "assets/images/logo.png")
	if err != nil {
		t.Errorf("loadAssetResource failed: %v", err)
	}

	expectedData := document.Assets.Images["logo.png"]
	if string(data) != string(expectedData) {
		t.Error("asset data doesn't match expected data")
	}

	// Test invalid asset path
	_, err = rm.loadAssetResource(document, "invalid/path")
	if err == nil {
		t.Error("loadAssetResource should fail for invalid path")
	}

	// Test unknown asset type
	_, err = rm.loadAssetResource(document, "assets/unknown/file.txt")
	if err == nil {
		t.Error("loadAssetResource should fail for unknown asset type")
	}

	// Test nonexistent asset
	_, err = rm.loadAssetResource(document, "assets/images/nonexistent.png")
	if err == nil {
		t.Error("loadAssetResource should fail for nonexistent asset")
	}
}

func TestResourceManager_DetermineMimeType(t *testing.T) {
	logger := &MockLogger{}
	metrics := &MockMetricsCollector{}

	rm := NewResourceManager(logger, metrics)

	tests := []struct {
		path     string
		expected string
	}{
		{"test.html", "text/html"},
		{"test.css", "text/css"},
		{"test.js", "application/javascript"},
		{"test.json", "application/json"},
		{"test.svg", "image/svg+xml"},
		{"test.png", "image/png"},
		{"test.jpg", "image/jpeg"},
		{"test.jpeg", "image/jpeg"},
		{"test.woff2", "font/woff2"},
		{"test.woff", "font/woff"},
		{"test.unknown", "application/octet-stream"},
	}

	for _, tt := range tests {
		t.Run(tt.path, func(t *testing.T) {
			mimeType := rm.determineMimeType(tt.path, []byte{})
			if mimeType != tt.expected {
				t.Errorf("expected MIME type '%s', got '%s'", tt.expected, mimeType)
			}
		})
	}
}

func TestResourceManager_ResourceExists(t *testing.T) {
	logger := &MockLogger{}
	metrics := &MockMetricsCollector{}

	rm := NewResourceManager(logger, metrics)
	document := createTestDocument()

	// Test existing resource
	if !rm.resourceExists(document, "content/index.html") {
		t.Error("resourceExists should return true for existing resource")
	}

	// Test nonexistent resource
	if rm.resourceExists(document, "nonexistent/resource.txt") {
		t.Error("resourceExists should return false for nonexistent resource")
	}

	// Test with nil manifest
	documentWithoutManifest := &core.LIVDocument{}
	if rm.resourceExists(documentWithoutManifest, "content/index.html") {
		t.Error("resourceExists should return false when manifest is nil")
	}
}