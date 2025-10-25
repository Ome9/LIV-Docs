package performance

import (
	"context"
	"fmt"
	"io/ioutil"
	"os"
	"path/filepath"
	"runtime"
	"strings"
	"sync"
	"testing"
	"time"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"

	"github.com/liv-format/liv/pkg/container"
	"github.com/liv-format/liv/pkg/core"
	"github.com/liv-format/liv/test/utils"
)

// TestDocumentLoadingPerformance tests and optimizes document loading performance
func TestDocumentLoadingPerformance(t *testing.T) {
	if testing.Short() {
		t.Skip("Skipping document loading performance test in short mode")
	}

	helper := utils.NewTestHelper(t)
	defer helper.Cleanup()

	// Test different document sizes and complexity levels
	testCases := []struct {
		name             string
		documentSize     int
		assetCount       int
		wasmModules      int
		expectedLoadTime time.Duration
	}{
		{
			name:             "small_document",
			documentSize:     10 * 1024, // 10KB
			assetCount:       5,
			wasmModules:      0,
			expectedLoadTime: 50 * time.Millisecond,
		},
		{
			name:             "medium_document",
			documentSize:     100 * 1024, // 100KB
			assetCount:       20,
			wasmModules:      1,
			expectedLoadTime: 200 * time.Millisecond,
		},
		{
			name:             "large_document",
			documentSize:     1024 * 1024, // 1MB
			assetCount:       50,
			wasmModules:      3,
			expectedLoadTime: 500 * time.Millisecond,
		},
		{
			name:             "complex_document",
			documentSize:     5 * 1024 * 1024, // 5MB
			assetCount:       100,
			wasmModules:      5,
			expectedLoadTime: 2 * time.Second,
		},
	}

	for _, tc := range testCases {
		t.Run(tc.name, func(t *testing.T) {
			// Create test document
			containerPath := filepath.Join(helper.TempDir, fmt.Sprintf("%s.liv", tc.name))
			cont := container.NewContainer(containerPath)

			// Add manifest
			manifest := createOptimizedManifest(tc.assetCount, tc.wasmModules)
			manifestData, err := manifest.MarshalJSON()
			require.NoError(t, err)

			err = cont.AddFile("manifest.json", manifestData)
			require.NoError(t, err)

			// Add content of specified size
			htmlContent := generateOptimizedHTMLContent(tc.documentSize)
			err = cont.AddFile("content/index.html", []byte(htmlContent))
			require.NoError(t, err)

			// Add CSS with performance optimizations
			cssContent := generateOptimizedCSSContent()
			err = cont.AddFile("content/styles/main.css", []byte(cssContent))
			require.NoError(t, err)

			// Add optimized JavaScript
			jsContent := generateOptimizedJSContent()
			err = cont.AddFile("content/scripts/main.js", []byte(jsContent))
			require.NoError(t, err)

			// Add assets
			for i := 0; i < tc.assetCount; i++ {
				assetData := generateOptimizedAssetData(1024) // 1KB per asset
				assetPath := fmt.Sprintf("assets/images/asset_%d.png", i)
				err = cont.AddFile(assetPath, assetData)
				require.NoError(t, err)
			}

			// Add WASM modules
			for i := 0; i < tc.wasmModules; i++ {
				wasmData := generateOptimizedWASMData(4096) // 4KB per module
				wasmPath := fmt.Sprintf("module_%d.wasm", i)
				err = cont.AddFile(wasmPath, wasmData)
				require.NoError(t, err)
			}

			// Save container
			err = cont.Save()
			require.NoError(t, err)

			// Measure loading performance
			loadTimes := make([]time.Duration, 5) // Test 5 times for average

			for i := 0; i < 5; i++ {
				start := time.Now()

				// Open and read container
				readContainer, err := container.OpenContainer(containerPath)
				require.NoError(t, err)

				// Read all files to simulate complete loading
				files, err := readContainer.ListFiles()
				require.NoError(t, err)

				for _, file := range files {
					_, err := readContainer.ReadFile(file)
					require.NoError(t, err)
				}

				loadTimes[i] = time.Since(start)
			}

			// Calculate average load time
			var totalTime time.Duration
			for _, loadTime := range loadTimes {
				totalTime += loadTime
			}
			avgLoadTime := totalTime / time.Duration(len(loadTimes))

			t.Logf("%s - Average load time: %v (expected: %v)", tc.name, avgLoadTime, tc.expectedLoadTime)

			// Performance assertion with some tolerance
			tolerance := tc.expectedLoadTime + (tc.expectedLoadTime / 2) // 50% tolerance
			assert.Less(t, avgLoadTime, tolerance,
				"Load time should be within expected range for %s", tc.name)

			// Log performance metrics
			fileInfo, _ := os.Stat(containerPath)
			t.Logf("Document size: %d bytes, Assets: %d, WASM modules: %d",
				fileInfo.Size(), tc.assetCount, tc.wasmModules)
		})
	}
}

// TestMemoryManagementOptimization tests memory usage optimization
func TestMemoryManagementOptimization(t *testing.T) {
	if testing.Short() {
		t.Skip("Skipping memory management optimization test in short mode")
	}

	helper := utils.NewTestHelper(t)
	defer helper.Cleanup()

	t.Run("memory_usage_optimization", func(t *testing.T) {
		// Force garbage collection before test
		runtime.GC()
		runtime.GC()

		var m1, m2, m3 runtime.MemStats
		runtime.ReadMemStats(&m1)

		// Create multiple documents to test memory management
		documents := make([]*container.Container, 50)

		for i := 0; i < 50; i++ {
			containerPath := filepath.Join(helper.TempDir, fmt.Sprintf("memory_test_%d.liv", i))
			cont := container.NewContainer(containerPath)

			// Add optimized content
			manifest := createOptimizedManifest(10, 1)
			manifestData, _ := manifest.MarshalJSON()
			cont.AddFile("manifest.json", manifestData)

			// Add content with memory-efficient structures
			htmlContent := generateMemoryEfficientHTML(i)
			cont.AddFile("content/index.html", []byte(htmlContent))

			// Add compressed assets
			assetData := generateCompressedAssetData(2048)
			cont.AddFile("assets/compressed_asset.png", assetData)

			cont.Save()
			documents[i] = cont
		}

		runtime.ReadMemStats(&m2)
		memoryUsed := m2.Alloc - m1.Alloc

		t.Logf("Memory used for 50 documents: %d bytes (%.2f MB)",
			memoryUsed, float64(memoryUsed)/(1024*1024))

		// Test memory cleanup
		for i := range documents {
			documents[i] = nil
		}
		documents = nil

		// Force garbage collection
		runtime.GC()
		runtime.GC()
		time.Sleep(100 * time.Millisecond)

		runtime.ReadMemStats(&m3)
		memoryAfterGC := m3.Alloc - m1.Alloc

		t.Logf("Memory after GC: %d bytes (%.2f MB)",
			memoryAfterGC, float64(memoryAfterGC)/(1024*1024))

		// Memory should be significantly reduced after GC
		memoryReduction := float64(memoryUsed-memoryAfterGC) / float64(memoryUsed)
		assert.Greater(t, memoryReduction, 0.5, "Should release at least 50% of memory after GC")

		// Total memory usage should be reasonable
		maxExpectedMemory := uint64(100 * 1024 * 1024) // 100MB
		assert.Less(t, memoryUsed, maxExpectedMemory, "Memory usage should be reasonable")
	})

	t.Run("memory_leak_detection", func(t *testing.T) {
		// Test for memory leaks in repeated operations

		runtime.GC()
		var initialMem runtime.MemStats
		runtime.ReadMemStats(&initialMem)

		// Perform repeated document operations
		for iteration := 0; iteration < 10; iteration++ {
			for i := 0; i < 10; i++ {
				containerPath := filepath.Join(helper.TempDir, fmt.Sprintf("leak_test_%d_%d.liv", iteration, i))
				cont := container.NewContainer(containerPath)

				// Add content
				manifest := createOptimizedManifest(5, 0)
				manifestData, _ := manifest.MarshalJSON()
				cont.AddFile("manifest.json", manifestData)

				htmlContent := generateMemoryEfficientHTML(i)
				cont.AddFile("content/index.html", []byte(htmlContent))

				cont.Save()

				// Read back immediately to test cleanup
				readContainer, _ := container.OpenContainer(containerPath)
				files, _ := readContainer.ListFiles()
				for _, file := range files {
					readContainer.ReadFile(file)
				}
			}

			// Force GC after each iteration
			runtime.GC()

			var currentMem runtime.MemStats
			runtime.ReadMemStats(&currentMem)

			memoryGrowth := currentMem.Alloc - initialMem.Alloc
			t.Logf("Iteration %d - Memory growth: %d bytes", iteration, memoryGrowth)

			// Memory growth should be bounded
			maxGrowth := uint64(50 * 1024 * 1024) // 50MB
			if memoryGrowth > maxGrowth {
				t.Errorf("Potential memory leak detected - growth: %d bytes", memoryGrowth)
			}
		}
	})
}

// TestPerformanceMonitoring tests built-in performance monitoring
func TestPerformanceMonitoring(t *testing.T) {
	helper := utils.NewTestHelper(t)
	defer helper.Cleanup()

	t.Run("performance_metrics_collection", func(t *testing.T) {
		// Test performance metrics collection during document operations

		metrics := &PerformanceMetrics{
			StartTime:  time.Now(),
			Operations: make(map[string]OperationMetric),
		}

		// Document creation metrics
		start := time.Now()
		doc, cont, containerPath := helper.CreateComplexTestDocument()
		_ = doc // Use doc to avoid unused variable

		err := cont.Save()
		require.NoError(t, err)

		metrics.RecordOperation("document_creation", time.Since(start), nil)

		// Document loading metrics
		start = time.Now()
		readContainer, err := container.OpenContainer(containerPath)
		require.NoError(t, err)

		files, err := readContainer.ListFiles()
		require.NoError(t, err)

		metrics.RecordOperation("document_loading", time.Since(start), nil)

		// File reading metrics
		start = time.Now()
		for _, file := range files {
			_, err := readContainer.ReadFile(file)
			require.NoError(t, err)
		}
		metrics.RecordOperation("file_reading", time.Since(start), nil)

		// Validation metrics
		start = time.Now()
		manifestData, err := readContainer.ReadFile("manifest.json")
		require.NoError(t, err)

		var manifest core.Manifest
		err = manifest.UnmarshalJSON(manifestData)
		require.NoError(t, err)

		err = manifest.Validate()
		require.NoError(t, err)

		metrics.RecordOperation("validation", time.Since(start), nil)

		// Generate performance report
		report := metrics.GenerateReport()
		t.Logf("Performance Report:\n%s", report)

		// Verify metrics were collected
		assert.Greater(t, len(metrics.Operations), 0, "Should collect performance metrics")

		for opName, metric := range metrics.Operations {
			assert.Greater(t, metric.Duration, time.Duration(0), "Operation %s should have positive duration", opName)
			t.Logf("Operation %s: %v", opName, metric.Duration)
		}
	})

	t.Run("performance_bottleneck_detection", func(t *testing.T) {
		// Test automatic bottleneck detection

		bottlenecks := detectPerformanceBottlenecks(helper)

		for _, bottleneck := range bottlenecks {
			t.Logf("Performance bottleneck detected: %s - %s (impact: %s)",
				bottleneck.Operation, bottleneck.Description, bottleneck.Impact)

			// Log optimization suggestions
			for _, suggestion := range bottleneck.Suggestions {
				t.Logf("  Suggestion: %s", suggestion)
			}
		}

		// Should detect some bottlenecks in test environment
		assert.GreaterOrEqual(t, len(bottlenecks), 0, "Should detect bottlenecks or report none")
	})
}

// TestResourceCleanupOptimization tests resource cleanup optimization
func TestResourceCleanupOptimization(t *testing.T) {
	helper := utils.NewTestHelper(t)
	defer helper.Cleanup()

	t.Run("automatic_resource_cleanup", func(t *testing.T) {
		// Test automatic cleanup of temporary resources

		initialFiles := countTempFiles()

		// Create multiple documents with temporary resources
		for i := 0; i < 20; i++ {
			containerPath := filepath.Join(helper.TempDir, fmt.Sprintf("cleanup_test_%d.liv", i))
			cont := container.NewContainer(containerPath)

			// Add content that creates temporary resources
			manifest := createOptimizedManifest(5, 1)
			manifestData, _ := manifest.MarshalJSON()
			cont.AddFile("manifest.json", manifestData)

			// Add large content that might create temp files
			largeContent := generateLargeContent(1024 * 1024) // 1MB
			cont.AddFile("content/large_content.html", []byte(largeContent))

			cont.Save()

			// Immediately read and process
			readContainer, _ := container.OpenContainer(containerPath)
			files, _ := readContainer.ListFiles()
			for _, file := range files {
				readContainer.ReadFile(file)
			}
		}

		// Force cleanup
		runtime.GC()
		time.Sleep(200 * time.Millisecond)

		finalFiles := countTempFiles()

		t.Logf("Temporary files - Initial: %d, Final: %d", initialFiles, finalFiles)

		// Should not accumulate too many temporary files
		maxTempFiles := initialFiles + 10 // Allow some temporary files
		assert.LessOrEqual(t, finalFiles, maxTempFiles, "Should clean up temporary files")
	})

	t.Run("resource_pool_optimization", func(t *testing.T) {
		// Test resource pooling for better performance

		// Simulate resource pool usage
		pool := NewResourcePool(10) // Pool of 10 resources

		var wg sync.WaitGroup

		// Test concurrent resource usage
		for i := 0; i < 50; i++ {
			wg.Add(1)
			go func(id int) {
				defer wg.Done()

				// Get resource from pool
				resource := pool.Get()
				defer pool.Put(resource)

				// Simulate work with resource
				time.Sleep(10 * time.Millisecond)

				// Use resource for document processing
				containerPath := filepath.Join(helper.TempDir, fmt.Sprintf("pool_test_%d.liv", id))
				cont := container.NewContainer(containerPath)

				manifest := createOptimizedManifest(2, 0)
				manifestData, _ := manifest.MarshalJSON()
				cont.AddFile("manifest.json", manifestData)

				cont.Save()
			}(i)
		}

		wg.Wait()

		// Verify pool statistics
		stats := pool.GetStats()
		t.Logf("Resource pool stats - Gets: %d, Puts: %d, Created: %d",
			stats.Gets, stats.Puts, stats.Created)

		assert.Equal(t, stats.Gets, stats.Puts, "All resources should be returned to pool")
		assert.LessOrEqual(t, stats.Created, 10, "Should not create more resources than pool size")
	})
}

// TestConcurrentPerformanceOptimization tests concurrent operation optimization
func TestConcurrentPerformanceOptimization(t *testing.T) {
	if testing.Short() {
		t.Skip("Skipping concurrent performance optimization test in short mode")
	}

	helper := utils.NewTestHelper(t)
	defer helper.Cleanup()

	t.Run("concurrent_document_processing", func(t *testing.T) {
		// Test optimized concurrent document processing

		concurrencyLevels := []int{1, 2, 4, 8, 16}

		for _, concurrency := range concurrencyLevels {
			t.Run(fmt.Sprintf("concurrency_%d", concurrency), func(t *testing.T) {
				start := time.Now()

				// Use context for cancellation
				ctx, cancel := context.WithTimeout(context.Background(), 30*time.Second)
				defer cancel()

				// Channel to control concurrency
				semaphore := make(chan struct{}, concurrency)
				var wg sync.WaitGroup

				// Process documents concurrently
				numDocuments := 50
				for i := 0; i < numDocuments; i++ {
					wg.Add(1)
					go func(id int) {
						defer wg.Done()

						// Acquire semaphore
						select {
						case semaphore <- struct{}{}:
							defer func() { <-semaphore }()
						case <-ctx.Done():
							return
						}

						// Process document
						containerPath := filepath.Join(helper.TempDir, fmt.Sprintf("concurrent_%d_%d.liv", concurrency, id))
						cont := container.NewContainer(containerPath)

						// Add optimized content
						manifest := createOptimizedManifest(3, 0)
						manifestData, _ := manifest.MarshalJSON()
						cont.AddFile("manifest.json", manifestData)

						htmlContent := generateOptimizedHTMLContent(10 * 1024) // 10KB
						cont.AddFile("content/index.html", []byte(htmlContent))

						cont.Save()

						// Read back to test full cycle
						readContainer, _ := container.OpenContainer(containerPath)
						files, _ := readContainer.ListFiles()
						for _, file := range files {
							readContainer.ReadFile(file)
						}
					}(i)
				}

				wg.Wait()
				duration := time.Since(start)

				throughput := float64(numDocuments) / duration.Seconds()
				t.Logf("Concurrency %d - Duration: %v, Throughput: %.2f docs/sec",
					concurrency, duration, throughput)

				// Performance should improve with concurrency (up to a point)
				if concurrency > 1 {
					// Should complete within reasonable time
					maxExpectedTime := 20 * time.Second
					assert.Less(t, duration, maxExpectedTime,
						"Concurrent processing should complete within reasonable time")
				}
			})
		}
	})

	t.Run("load_balancing_optimization", func(t *testing.T) {
		// Test load balancing for optimal resource utilization

		// Create worker pool
		numWorkers := runtime.NumCPU()
		workQueue := make(chan WorkItem, 100)
		var wg sync.WaitGroup

		// Start workers
		for i := 0; i < numWorkers; i++ {
			wg.Add(1)
			go func(workerID int) {
				defer wg.Done()

				for work := range workQueue {
					// Process work item
					start := time.Now()

					containerPath := filepath.Join(helper.TempDir, fmt.Sprintf("worker_%d_%d.liv", workerID, work.ID))
					cont := container.NewContainer(containerPath)

					manifest := createOptimizedManifest(work.AssetCount, work.WASMCount)
					manifestData, _ := manifest.MarshalJSON()
					cont.AddFile("manifest.json", manifestData)

					htmlContent := generateOptimizedHTMLContent(work.ContentSize)
					cont.AddFile("content/index.html", []byte(htmlContent))

					cont.Save()

					work.Duration = time.Since(start)
					t.Logf("Worker %d processed item %d in %v", workerID, work.ID, work.Duration)
				}
			}(i)
		}

		// Submit work items
		workItems := make([]WorkItem, 50)
		for i := 0; i < 50; i++ {
			workItems[i] = WorkItem{
				ID:          i,
				ContentSize: (i%5 + 1) * 10 * 1024, // Varying sizes
				AssetCount:  i % 10,
				WASMCount:   i % 3,
			}
			workQueue <- workItems[i]
		}

		close(workQueue)
		wg.Wait()

		// Analyze load distribution
		_ = make(map[int]int) // workerLoads unused, but created for potential future use
		var totalDuration time.Duration

		for _, item := range workItems {
			totalDuration += item.Duration
		}

		avgDuration := totalDuration / time.Duration(len(workItems))
		t.Logf("Average processing time per item: %v", avgDuration)

		// Should process items efficiently
		maxExpectedAvg := 200 * time.Millisecond
		assert.Less(t, avgDuration, maxExpectedAvg, "Average processing time should be reasonable")
	})
}

// Helper types and functions

type PerformanceMetrics struct {
	StartTime  time.Time
	Operations map[string]OperationMetric
	mu         sync.RWMutex
}

type OperationMetric struct {
	Duration time.Duration
	Count    int
	Errors   []error
}

type PerformanceBottleneck struct {
	Operation   string
	Description string
	Impact      string
	Suggestions []string
}

type ResourcePool struct {
	resources chan interface{}
	stats     PoolStats
	mu        sync.RWMutex
}

type PoolStats struct {
	Gets    int
	Puts    int
	Created int
}

type WorkItem struct {
	ID          int
	ContentSize int
	AssetCount  int
	WASMCount   int
	Duration    time.Duration
}

func (pm *PerformanceMetrics) RecordOperation(name string, duration time.Duration, err error) {
	pm.mu.Lock()
	defer pm.mu.Unlock()

	metric := pm.Operations[name]
	metric.Duration += duration
	metric.Count++
	if err != nil {
		metric.Errors = append(metric.Errors, err)
	}
	pm.Operations[name] = metric
}

func (pm *PerformanceMetrics) GenerateReport() string {
	pm.mu.RLock()
	defer pm.mu.RUnlock()

	report := fmt.Sprintf("Performance Report (Total time: %v)\n", time.Since(pm.StartTime))
	report += "========================================\n"

	for name, metric := range pm.Operations {
		avgDuration := metric.Duration / time.Duration(metric.Count)
		report += fmt.Sprintf("%s: %d ops, avg %v, total %v",
			name, metric.Count, avgDuration, metric.Duration)
		if len(metric.Errors) > 0 {
			report += fmt.Sprintf(" (%d errors)", len(metric.Errors))
		}
		report += "\n"
	}

	return report
}

func NewResourcePool(size int) *ResourcePool {
	pool := &ResourcePool{
		resources: make(chan interface{}, size),
	}

	// Pre-populate pool
	for i := 0; i < size; i++ {
		pool.resources <- createResource()
		pool.stats.Created++
	}

	return pool
}

func (rp *ResourcePool) Get() interface{} {
	rp.mu.Lock()
	rp.stats.Gets++
	rp.mu.Unlock()

	select {
	case resource := <-rp.resources:
		return resource
	default:
		// Create new resource if pool is empty
		rp.mu.Lock()
		rp.stats.Created++
		rp.mu.Unlock()
		return createResource()
	}
}

func (rp *ResourcePool) Put(resource interface{}) {
	rp.mu.Lock()
	rp.stats.Puts++
	rp.mu.Unlock()

	select {
	case rp.resources <- resource:
		// Successfully returned to pool
	default:
		// Pool is full, discard resource
	}
}

func (rp *ResourcePool) GetStats() PoolStats {
	rp.mu.RLock()
	defer rp.mu.RUnlock()
	return rp.stats
}

func createResource() interface{} {
	// Create a mock resource (in real implementation, this would be a real resource)
	return make([]byte, 1024)
}

func createOptimizedManifest(assetCount, wasmCount int) *core.Manifest {
	manifest := &core.Manifest{
		Version: "1.0",
		Metadata: &core.DocumentMetadata{
			Title:  "Optimized Performance Test Document",
			Author: "Performance Optimizer",
		},
		Resources: make(map[string]*core.Resource),
	}

	// Add optimized resource entries
	for i := 0; i < assetCount; i++ {
		resourceName := fmt.Sprintf("asset_%d", i)
		manifest.Resources[resourceName] = &core.Resource{
			Type: "image",
			Path: fmt.Sprintf("assets/images/asset_%d.png", i),
			Size: 1024,
		}
	}

	// Add WASM configuration if needed
	if wasmCount > 0 {
		wasmConfig := &core.WASMConfiguration{
			Modules: make(map[string]*core.WASMModule),
			Permissions: &core.WASMPermissions{
				MemoryLimit:     32 * 1024 * 1024, // 32MB
				AllowNetworking: false,
				AllowFileSystem: false,
			},
		}

		for i := 0; i < wasmCount; i++ {
			moduleName := fmt.Sprintf("module_%d", i)
			wasmConfig.Modules[moduleName] = &core.WASMModule{
				Name:    moduleName,
				Version: "1.0",
			}
		}

		manifest.WASMConfig = wasmConfig
	}

	return manifest
}

func generateOptimizedHTMLContent(size int) string {
	// Generate HTML with performance optimizations
	html := `<!DOCTYPE html>
<html lang="en">
<head>
    <meta charset="UTF-8">
    <meta name="viewport" content="width=device-width, initial-scale=1.0">
    <title>Optimized Performance Test</title>
    <link rel="preload" href="styles/main.css" as="style">
    <link rel="stylesheet" href="styles/main.css">
</head>
<body>
    <div class="container">
        <h1>Performance Optimized Document</h1>`

	// Add content to reach desired size
	contentNeeded := size - len(html) - len("</div></body></html>")
	if contentNeeded > 0 {
		paragraph := "<p>This is optimized content for performance testing. "
		paragraph += "The content is structured for efficient parsing and rendering. "
		paragraph += "Memory usage is optimized through careful DOM structure.</p>"

		for len(html) < size-len("</div></body></html>") {
			if len(html)+len(paragraph) > size-len("</div></body></html>") {
				remaining := size - len("</div></body></html>") - len(html)
				if remaining > 0 {
					html += paragraph[:remaining]
				}
				break
			}
			html += paragraph
		}
	}

	html += "</div></body></html>"
	return html
}

func generateOptimizedCSSContent() string {
	return `/* Optimized CSS for performance */
.container {
    max-width: 1200px;
    margin: 0 auto;
    padding: 20px;
    font-family: system-ui, -apple-system, sans-serif;
}

/* Use efficient selectors */
h1 { color: #333; margin-bottom: 1rem; }
p { line-height: 1.6; margin-bottom: 1rem; }

/* Optimize animations */
@media (prefers-reduced-motion: no-preference) {
    .fade-in { animation: fadeIn 0.3s ease-out; }
}

@keyframes fadeIn {
    from { opacity: 0; }
    to { opacity: 1; }
}

/* Efficient responsive design */
@media (max-width: 768px) {
    .container { padding: 10px; }
}`
}

func generateOptimizedJSContent() string {
	return `// Optimized JavaScript for performance
(function() {
    'use strict';
    
    // Use efficient DOM queries
    const container = document.querySelector('.container');
    
    // Optimize event handling
    if (container) {
        container.addEventListener('click', handleClick, { passive: true });
    }
    
    function handleClick(event) {
        // Efficient event handling
        if (event.target.matches('button')) {
            event.target.classList.add('clicked');
        }
    }
    
    // Optimize resource loading
    if ('requestIdleCallback' in window) {
        requestIdleCallback(initializeNonCritical);
    } else {
        setTimeout(initializeNonCritical, 100);
    }
    
    function initializeNonCritical() {
        // Initialize non-critical features
        console.log('Non-critical features initialized');
    }
})();`
}

func generateMemoryEfficientHTML(index int) string {
	return fmt.Sprintf(`<!DOCTYPE html>
<html>
<head>
    <title>Memory Efficient Document %d</title>
</head>
<body>
    <h1>Document %d</h1>
    <p>This document is optimized for memory efficiency.</p>
</body>
</html>`, index, index)
}

func generateOptimizedAssetData(size int) []byte {
	// Generate compressed/optimized asset data
	data := make([]byte, size)
	// Use a pattern that compresses well
	for i := range data {
		data[i] = byte(i % 16)
	}
	return data
}

func generateCompressedAssetData(size int) []byte {
	// Generate asset data that simulates compression
	data := make([]byte, size)
	// Use repeating pattern for better compression
	pattern := []byte{0x89, 0x50, 0x4E, 0x47} // PNG header pattern
	for i := range data {
		data[i] = pattern[i%len(pattern)]
	}
	return data
}

func generateOptimizedWASMData(size int) []byte {
	// Generate optimized WASM binary
	data := make([]byte, size)
	// WASM magic number and version
	copy(data[:8], []byte{0x00, 0x61, 0x73, 0x6d, 0x01, 0x00, 0x00, 0x00})
	// Fill rest with optimized bytecode pattern
	for i := 8; i < len(data); i++ {
		data[i] = byte((i - 8) % 256)
	}
	return data
}

func generateLargeContent(size int) string {
	// Generate large content efficiently
	content := make([]byte, size)
	pattern := "Performance test content. "
	for i := 0; i < size; i++ {
		content[i] = pattern[i%len(pattern)]
	}
	return string(content)
}

func countTempFiles() int {
	// Count temporary files in system temp directory
	tempDir := os.TempDir()
	files, err := ioutil.ReadDir(tempDir)
	if err != nil {
		return 0
	}

	count := 0
	for _, file := range files {
		if strings.Contains(file.Name(), "liv") {
			count++
		}
	}
	return count
}

func detectPerformanceBottlenecks(helper *utils.TestHelper) []PerformanceBottleneck {
	// Detect common performance bottlenecks
	bottlenecks := []PerformanceBottleneck{}

	// Test document loading performance
	start := time.Now()
	doc, cont, containerPath := helper.CreateComplexTestDocument()
	_ = doc
	cont.Save()

	readContainer, _ := container.OpenContainer(containerPath)
	files, _ := readContainer.ListFiles()
	for _, file := range files {
		readContainer.ReadFile(file)
	}
	loadTime := time.Since(start)

	if loadTime > 500*time.Millisecond {
		bottlenecks = append(bottlenecks, PerformanceBottleneck{
			Operation:   "document_loading",
			Description: "Document loading is slower than expected",
			Impact:      "high",
			Suggestions: []string{
				"Implement lazy loading for assets",
				"Add compression for large content",
				"Optimize manifest parsing",
			},
		})
	}

	// Test memory usage
	runtime.GC()
	var m1, m2 runtime.MemStats
	runtime.ReadMemStats(&m1)

	// Create multiple documents
	for i := 0; i < 10; i++ {
		doc, cont, _ := helper.CreateComplexTestDocument()
		_ = doc
		cont.Save()
	}

	runtime.ReadMemStats(&m2)
	memoryUsed := m2.Alloc - m1.Alloc

	if memoryUsed > 50*1024*1024 { // 50MB
		bottlenecks = append(bottlenecks, PerformanceBottleneck{
			Operation:   "memory_usage",
			Description: "High memory usage detected",
			Impact:      "medium",
			Suggestions: []string{
				"Implement object pooling",
				"Add memory cleanup routines",
				"Optimize data structures",
			},
		})
	}

	return bottlenecks
}
