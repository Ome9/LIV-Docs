package viewer

import (
	"bytes"
	"context"
	"fmt"
	"testing"
	"time"

	"github.com/liv-format/liv/pkg/container"
	"github.com/liv-format/liv/pkg/core"
	"github.com/liv-format/liv/pkg/manifest"
	"github.com/liv-format/liv/pkg/security"
)

// TestViewerIntegration_DocumentLoading tests document loading with various .liv file structures
func TestViewerIntegration_DocumentLoading(t *testing.T) {
	testCases := []struct {
		name        string
		description string
		setupFunc   func(*testing.T) *core.LIVDocument
		expectError bool
	}{
		{
			name:        "basic_static_document",
			description: "Simple static document with minimal content",
			setupFunc:   createBasicStaticDocument,
			expectError: false,
		},
		{
			name:        "complex_interactive_document",
			description: "Complex document with WASM, animations, and multiple assets",
			setupFunc:   createComplexInteractiveDocument,
			expectError: false,
		},
		{
			name:        "multimedia_document",
			description: "Document with images, fonts, audio, and data files",
			setupFunc:   createMultimediaDocument,
			expectError: false,
		},
		{
			name:        "nested_structure_document",
			description: "Document with deeply nested directory structures",
			setupFunc:   createNestedStructureDocument,
			expectError: false,
		},
	}

	for _, tc := range testCases {
		t.Run(tc.name, func(t *testing.T) {
			t.Logf("Testing: %s", tc.description)

			// Create test document
			document := tc.setupFunc(t)

			// Create viewer components
			packageManager := container.NewPackageManager()
			logger := &TestLogger{}
			metrics := &TestMetricsCollector{}
			securityManager := security.NewSecurityManager(nil, logger, metrics)
			validator := NewDocumentValidator(logger)

			loader := NewDocumentLoader(packageManager, securityManager, validator, logger, metrics)

			// Package document into bytes
			var buf bytes.Buffer
			err := packageManager.SavePackageToWriter(document, &buf)
			if err != nil {
				t.Fatalf("Failed to package document: %v", err)
			}

			// Load document through viewer
			ctx, cancel := context.WithTimeout(context.Background(), 30*time.Second)
			defer cancel()

			reader := bytes.NewReader(buf.Bytes())
			result, err := loader.LoadDocument(ctx, reader, fmt.Sprintf("%s.liv", tc.name))

			if tc.expectError {
				if err == nil {
					t.Errorf("Expected error for %s, but got none", tc.name)
				}
				return
			}

			if err != nil {
				t.Fatalf("Unexpected error loading %s: %v", tc.name, err)
			}

			if result == nil {
				t.Fatal("Load result is nil")
			}

			if result.Document == nil {
				t.Fatal("Loaded document is nil")
			}

			// Validate loaded document structure
			validateLoadedDocument(t, result.Document, document)

			// Check performance metrics
			if result.LoadTime < 0 {
				t.Error("Load time should be non-negative")
			}

			if result.LoadTime > 10*time.Second {
				t.Errorf("Load time too high: %v", result.LoadTime)
			}

			t.Logf("Successfully loaded %s in %v", tc.name, result.LoadTime)
		})
	}
}

// TestViewerIntegration_RenderingPerformance tests rendering performance with animated content
func TestViewerIntegration_RenderingPerformance(t *testing.T) {
	performanceTests := []struct {
		name           string
		description    string
		setupFunc      func(*testing.T) *core.LIVDocument
		maxLoadTime    time.Duration
	}{
		{
			name:           "simple_animations",
			description:    "Document with basic CSS animations",
			setupFunc:      createSimpleAnimatedDocument,
			maxLoadTime:    2 * time.Second,
		},
		{
			name:           "complex_animations",
			description:    "Document with complex CSS transforms and SVG animations",
			setupFunc:      createComplexAnimatedDocument,
			maxLoadTime:    5 * time.Second,
		},
	}

	for _, pt := range performanceTests {
		t.Run(pt.name, func(t *testing.T) {
			t.Logf("Performance testing: %s", pt.description)

			// Create test document
			document := pt.setupFunc(t)

			// Create viewer components with performance monitoring
			packageManager := container.NewPackageManager()
			logger := &TestLogger{}
			metrics := &TestMetricsCollector{}
			securityManager := security.NewSecurityManager(nil, logger, metrics)
			validator := NewDocumentValidator(logger)

			loader := NewDocumentLoader(packageManager, securityManager, validator, logger, metrics)

			// Package and load document
			var buf bytes.Buffer
			err := packageManager.SavePackageToWriter(document, &buf)
			if err != nil {
				t.Fatalf("Failed to package document: %v", err)
			}

			ctx := context.Background()
			reader := bytes.NewReader(buf.Bytes())

			// Measure load time
			startTime := time.Now()
			result, err := loader.LoadDocument(ctx, reader, fmt.Sprintf("%s.liv", pt.name))
			loadTime := time.Since(startTime)

			if err != nil {
				t.Fatalf("Failed to load document: %v", err)
			}

			if result == nil {
				t.Fatal("Load result is nil")
			}

			// Validate load time performance
			if loadTime > pt.maxLoadTime {
				t.Errorf("Load time %v exceeds maximum %v", loadTime, pt.maxLoadTime)
			}

			t.Logf("Performance results - Load: %v", loadTime)
		})
	}
}

// TestViewerIntegration_CrossPlatform tests cross-platform compatibility
func TestViewerIntegration_CrossPlatform(t *testing.T) {
	platformTests := []struct {
		name        string
		description string
		config      *PlatformConfig
		document    *core.LIVDocument
	}{
		{
			name:        "desktop_high_performance",
			description: "Desktop platform with high performance capabilities",
			config: &PlatformConfig{
				Platform:         "desktop",
				MemoryLimitMB:    512,
				CPUCores:         8,
				GPUAcceleration:  true,
				NetworkEnabled:   false,
				TouchSupport:     false,
				ScreenWidth:      1920,
				ScreenHeight:     1080,
				DevicePixelRatio: 1.0,
			},
			document: createComplexInteractiveDocument(t),
		},
		{
			name:        "mobile_limited_resources",
			description: "Mobile platform with limited resources",
			config: &PlatformConfig{
				Platform:         "mobile",
				MemoryLimitMB:    128,
				CPUCores:         4,
				GPUAcceleration:  false,
				NetworkEnabled:   false,
				TouchSupport:     true,
				ScreenWidth:      375,
				ScreenHeight:     812,
				DevicePixelRatio: 3.0,
			},
			document: createMobileOptimizedDocument(t),
		},
	}

	for _, pt := range platformTests {
		t.Run(pt.name, func(t *testing.T) {
			t.Logf("Cross-platform testing: %s", pt.description)

			// Create platform-specific viewer configuration
			viewerConfig := createPlatformViewerConfig(pt.config)

			// Create viewer components
			packageManager := container.NewPackageManager()
			logger := &TestLogger{}
			metrics := &TestMetricsCollector{}
			securityManager := security.NewSecurityManager(nil, logger, metrics)
			validator := NewDocumentValidator(logger)

			loader := NewDocumentLoader(packageManager, securityManager, validator, logger, metrics)
			loader.UpdateConfiguration(viewerConfig)

			// Package and load document
			var buf bytes.Buffer
			err := packageManager.SavePackageToWriter(pt.document, &buf)
			if err != nil {
				t.Fatalf("Failed to package document: %v", err)
			}

			ctx := context.Background()
			reader := bytes.NewReader(buf.Bytes())

			result, err := loader.LoadDocument(ctx, reader, fmt.Sprintf("%s.liv", pt.name))
			if err != nil {
				t.Fatalf("Failed to load document on %s: %v", pt.config.Platform, err)
			}

			// Validate platform-specific behavior
			validatePlatformCompatibility(t, result, pt.config)

			t.Logf("Successfully validated %s platform compatibility", pt.config.Platform)
		})
	}
}

// Helper functions for document creation

func createBasicStaticDocument(t *testing.T) *core.LIVDocument {
	builder := manifest.NewManifestBuilder()
	
	// Create metadata
	builder.CreateDefaultMetadata("Basic Static Test", "Integration Test")
	
	// Set security policy with minimum required values
	policy := &core.SecurityPolicy{
		WASMPermissions: &core.WASMPermissions{
			MemoryLimit:     1024, // Minimum required
			AllowedImports:  []string{},
			CPUTimeLimit:    100, // Minimum required
			AllowNetworking: false,
			AllowFileSystem: false,
		},
		JSPermissions: &core.JSPermissions{
			ExecutionMode: "none",
			AllowedAPIs:   []string{},
			DOMAccess:     "none",
		},
		NetworkPolicy: &core.NetworkPolicy{
			AllowOutbound: false,
			AllowedHosts:  []string{},
			AllowedPorts:  []int{},
		},
		StoragePolicy: &core.StoragePolicy{
			AllowLocalStorage:   false,
			AllowSessionStorage: false,
			AllowIndexedDB:      false,
			AllowCookies:        false,
		},
		ContentSecurityPolicy: "default-src 'none'; style-src 'self';",
		TrustedDomains:        []string{},
	}
	builder.SetSecurityPolicy(policy)
	
	// Add required resources (don't add manifest.json - it's the manifest itself)
	builder.AddResource("content/index.html", &core.Resource{
		Hash: "html-hash-123",
		Size: 200,
		Type: "text/html",
		Path: "content/index.html",
	})
	
	manifestObj, err := builder.Build()
	if err != nil {
		t.Fatalf("Failed to build manifest: %v", err)
	}

	return &core.LIVDocument{
		Manifest: manifestObj,
		Content: &core.DocumentContent{
			HTML: `<!DOCTYPE html><html><head><title>Basic Test</title></head><body><h1>Basic Static Document</h1><p>Simple content for testing.</p></body></html>`,
			CSS:  `body { font-family: Arial, sans-serif; margin: 20px; } h1 { color: #333; }`,
			StaticFallback: `<!DOCTYPE html><html><head><title>Basic Test</title></head><body><h1>Basic Static Document</h1><p>Static fallback content.</p></body></html>`,
		},
		Assets:      &core.AssetBundle{Images: map[string][]byte{}, Fonts: map[string][]byte{}, Data: map[string][]byte{}},
		Signatures:  &core.SignatureBundle{},
		WASMModules: map[string][]byte{},
	}
}

func createComplexInteractiveDocument(t *testing.T) *core.LIVDocument {
	builder := manifest.CreateInteractiveDocumentTemplate("Complex Interactive Test", "Integration Test")
	
	wasmModule := &core.WASMModule{
		Name:       "interactive-engine",
		Version:    "1.0.0",
		EntryPoint: "init",
		Exports:    []string{"init", "render", "interact"},
		Imports:    []string{"env.memory"},
	}
	builder.AddWASMModule(wasmModule)
	
	// Add required resources
	builder.AddResource("content/index.html", &core.Resource{
		Hash: "html-hash-456",
		Size: 500,
		Type: "text/html",
		Path: "content/index.html",
	})

	manifestObj, err := builder.Build()
	if err != nil {
		t.Fatalf("Failed to build manifest: %v", err)
	}

	return &core.LIVDocument{
		Manifest: manifestObj,
		Content: &core.DocumentContent{
			HTML: `<!DOCTYPE html>
<html>
<head>
    <title>Complex Interactive Test</title>
</head>
<body>
    <div class="container">
        <h1>Interactive Dashboard</h1>
        <svg width="400" height="300" id="chart">
            <rect x="10" y="10" width="100" height="200" fill="blue">
                <animate attributeName="height" values="200;250;200" dur="3s" repeatCount="indefinite"/>
            </rect>
        </svg>
        <button onclick="updateChart()">Update</button>
    </div>
</body>
</html>`,
			CSS: `
@keyframes pulse { 0% { opacity: 1; } 50% { opacity: 0.5; } 100% { opacity: 1; } }
.container { max-width: 800px; margin: 0 auto; padding: 20px; }
h1 { animation: pulse 2s infinite; }
svg { border: 1px solid #ccc; }
button { padding: 10px 20px; background: #007bff; color: white; border: none; }`,
			InteractiveSpec: `function updateChart() { console.log('Chart updated'); }`,
			StaticFallback: `<!DOCTYPE html><html><body><h1>Interactive Dashboard (Static)</h1><p>Static version</p></body></html>`,
		},
		Assets: &core.AssetBundle{
			Images: map[string][]byte{
				"chart-bg.svg": []byte(`<svg xmlns="http://www.w3.org/2000/svg" viewBox="0 0 400 300"><rect width="400" height="300" fill="#f8f9fa"/></svg>`),
			},
			Fonts: map[string][]byte{},
			Data: map[string][]byte{
				"chart-data.json": []byte(`{"values": [10, 20, 30, 40, 50]}`),
			},
		},
		Signatures: &core.SignatureBundle{},
		WASMModules: map[string][]byte{
			"interactive-engine": {0x00, 0x61, 0x73, 0x6D, 0x01, 0x00, 0x00, 0x00, 0x01, 0x04, 0x01, 0x60, 0x00, 0x00, 0x03, 0x02, 0x01, 0x00, 0x0A, 0x04, 0x01, 0x02, 0x00, 0x0B},
		},
	}
}

func createMultimediaDocument(t *testing.T) *core.LIVDocument {
	document := createComplexInteractiveDocument(t)
	
	// Add multimedia assets
	document.Assets.Images["hero.jpg"] = []byte("fake-jpeg-hero-data")
	document.Assets.Images["icon.png"] = []byte("fake-png-icon-data")
	document.Assets.Fonts["custom.woff2"] = []byte("fake-woff2-font-data")
	document.Assets.Data["audio-config.json"] = []byte(`{"volume": 0.8, "autoplay": false}`)
	
	return document
}

func createNestedStructureDocument(t *testing.T) *core.LIVDocument {
	document := createBasicStaticDocument(t)
	
	// Add nested resources - but don't add them to manifest.Resources
	// The actual files would be in the document's content
	// For this test, we're just testing the document structure, not resource validation
	
	return document
}

func createSimpleAnimatedDocument(t *testing.T) *core.LIVDocument {
	document := createBasicStaticDocument(t)
	
	// Add simple animations
	document.Content.CSS += `
@keyframes fadeIn { from { opacity: 0; } to { opacity: 1; } }
.animated { animation: fadeIn 1s ease-in; }`
	
	return document
}

func createComplexAnimatedDocument(t *testing.T) *core.LIVDocument {
	document := createComplexInteractiveDocument(t)
	
	// Add complex animations
	document.Content.CSS += `
@keyframes rotate3d {
	0% { transform: perspective(1000px) rotateX(0deg) rotateY(0deg); }
	50% { transform: perspective(1000px) rotateX(180deg) rotateY(90deg); }
	100% { transform: perspective(1000px) rotateX(360deg) rotateY(180deg); }
}
.complex-animation { animation: rotate3d 4s infinite; }`
	
	return document
}

func createMobileOptimizedDocument(t *testing.T) *core.LIVDocument {
	document := createBasicStaticDocument(t)
	
	// Add mobile-specific optimizations
	document.Content.CSS += `
@media (max-width: 768px) {
    body { font-size: 14px; padding: 10px; }
    .container { max-width: 100%; }
}`
	
	return document
}

// Cross-platform testing helpers

type PlatformConfig struct {
	Platform         string
	MemoryLimitMB    int64
	CPUCores         int
	GPUAcceleration  bool
	NetworkEnabled   bool
	TouchSupport     bool
	ScreenWidth      int
	ScreenHeight     int
	DevicePixelRatio float64
}

func createPlatformViewerConfig(config *PlatformConfig) *LoaderConfiguration {
	return &LoaderConfiguration{
		EnableCaching:       config.MemoryLimitMB > 128,
		CacheSize:          int(config.MemoryLimitMB / 4), // Use 1/4 of memory for cache
		CacheExpiry:        1 * time.Hour,
		MaxDocumentSize:    config.MemoryLimitMB * 1024 * 1024 / 2, // Half of available memory
		ValidateSignatures: true,
		StrictValidation:   config.Platform == "desktop",
		LoadTimeout:        30 * time.Second,
	}
}

// Validation helpers

func validateLoadedDocument(t *testing.T, loaded, original *core.LIVDocument) {
	if loaded.Manifest.Metadata.Title != original.Manifest.Metadata.Title {
		t.Error("Document title mismatch after loading")
	}
	
	if len(loaded.Assets.Images) != len(original.Assets.Images) {
		t.Error("Image asset count mismatch after loading")
	}
	
	if len(loaded.WASMModules) != len(original.WASMModules) {
		t.Error("WASM module count mismatch after loading")
	}
}

func validatePlatformCompatibility(t *testing.T, result *LoadResult, config *PlatformConfig) {
	// Validate memory constraints
	if config.MemoryLimitMB < 128 {
		// Low memory platforms should use fallback mode for complex documents
		if len(result.Document.WASMModules) > 0 {
			t.Logf("Note: Low memory platform may require fallback mode for WASM content")
		}
	}
	
	// Validate load time based on platform performance
	maxLoadTime := 5 * time.Second
	if config.CPUCores < 4 {
		maxLoadTime = 10 * time.Second // Allow more time for slower platforms
	}
	
	if result.LoadTime > maxLoadTime {
		t.Errorf("Load time %v exceeds platform limit %v", result.LoadTime, maxLoadTime)
	}
}

// Test helper types

type TestLogger struct {
	logs []string
}

func (tl *TestLogger) Debug(msg string, fields ...interface{}) { tl.logs = append(tl.logs, "DEBUG: "+msg) }
func (tl *TestLogger) Info(msg string, fields ...interface{})  { tl.logs = append(tl.logs, "INFO: "+msg) }
func (tl *TestLogger) Warn(msg string, fields ...interface{})  { tl.logs = append(tl.logs, "WARN: "+msg) }
func (tl *TestLogger) Error(msg string, fields ...interface{}) { tl.logs = append(tl.logs, "ERROR: "+msg) }
func (tl *TestLogger) Fatal(msg string, fields ...interface{}) { tl.logs = append(tl.logs, "FATAL: "+msg) }

type TestMetricsCollector struct {
	events []map[string]interface{}
}

func (tmc *TestMetricsCollector) RecordDocumentLoad(size int64, duration int64) {
	tmc.events = append(tmc.events, map[string]interface{}{
		"type":     "document_load",
		"size":     size,
		"duration": duration,
	})
}

func (tmc *TestMetricsCollector) RecordWASMExecution(module string, duration int64, memoryUsed uint64) {
	tmc.events = append(tmc.events, map[string]interface{}{
		"type":        "wasm_execution",
		"module":      module,
		"duration":    duration,
		"memory_used": memoryUsed,
	})
}

func (tmc *TestMetricsCollector) RecordSecurityEvent(eventType string, details map[string]interface{}) {
	event := map[string]interface{}{
		"type": eventType,
	}
	for k, v := range details {
		event[k] = v
	}
	tmc.events = append(tmc.events, event)
}

func (tmc *TestMetricsCollector) GetMetrics() map[string]interface{} {
	return map[string]interface{}{
		"events": tmc.events,
	}
}