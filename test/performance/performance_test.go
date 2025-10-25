package performance

import (
	"fmt"
	"io/ioutil"
	"math/rand"
	"os"
	"path/filepath"
	"runtime"
	"strings"
	"testing"
	"time"

	"github.com/stretchr/testify/require"

	"github.com/liv-format/liv/pkg/container"
	"github.com/liv-format/liv/pkg/core"
)

// Performance test configuration
const (
	SmallDocumentSize  = 1024             // 1KB
	MediumDocumentSize = 1024 * 100       // 100KB
	LargeDocumentSize  = 1024 * 1024      // 1MB
	HugeDocumentSize   = 10 * 1024 * 1024 // 10MB
)

// TestDocumentCreationPerformance tests document creation performance with various sizes
func TestDocumentCreationPerformance(t *testing.T) {
	sizes := []struct {
		name string
		size int
	}{
		{"small", SmallDocumentSize},
		{"medium", MediumDocumentSize},
		{"large", LargeDocumentSize},
		{"huge", HugeDocumentSize},
	}

	for _, size := range sizes {
		t.Run(fmt.Sprintf("document_creation_%s", size.name), func(t *testing.T) {
			content := generateHTMLContent(size.size)

			start := time.Now()
			doc := core.NewDocument(
				core.DocumentMetadata{
					Title:    fmt.Sprintf("Performance Test %s", size.name),
					Author:   "Performance Tester",
					Created:  time.Now(),
					Modified: time.Now(),
					Version:  "1.0",
					Language: "en",
				},
				core.DocumentContent{
					HTML: content,
				},
			)
			err := doc.Validate()
			duration := time.Since(start)

			require.NoError(t, err)
			t.Logf("Document creation (%s): %v", size.name, duration)

			// Performance assertions
			switch size.name {
			case "small":
				require.Less(t, duration, 10*time.Millisecond, "Small document creation should be fast")
			case "medium":
				require.Less(t, duration, 50*time.Millisecond, "Medium document creation should be reasonable")
			case "large":
				require.Less(t, duration, 200*time.Millisecond, "Large document creation should complete quickly")
			case "huge":
				require.Less(t, duration, 1*time.Second, "Huge document creation should complete within 1 second")
			}
		})
	}
}

// TestContainerPerformance tests container operations performance
func TestContainerPerformance(t *testing.T) {
	tempDir, err := ioutil.TempDir("", "liv_perf_")
	require.NoError(t, err)
	defer os.RemoveAll(tempDir)

	fileCounts := []struct {
		name  string
		count int
	}{
		{"few_files", 10},
		{"many_files", 100},
		{"lots_of_files", 1000},
	}

	for _, fc := range fileCounts {
		t.Run(fmt.Sprintf("container_%s", fc.name), func(t *testing.T) {
			containerPath := filepath.Join(tempDir, fmt.Sprintf("%s.liv", fc.name))
			cont := container.NewContainer(containerPath)

			// Add files
			start := time.Now()
			for i := 0; i < fc.count; i++ {
				filename := fmt.Sprintf("file_%d.txt", i)
				content := fmt.Sprintf("Content for file %d", i)
				err := cont.AddFile(filename, []byte(content))
				require.NoError(t, err)
			}
			addDuration := time.Since(start)

			// Save container
			start = time.Now()
			err := cont.Save()
			saveDuration := time.Since(start)
			require.NoError(t, err)

			// Read container
			start = time.Now()
			readContainer, err := container.OpenContainer(containerPath)
			require.NoError(t, err)
			files, err := readContainer.ListFiles()
			require.NoError(t, err)
			readDuration := time.Since(start)

			require.Len(t, files, fc.count)

			t.Logf("Container %s - Add: %v, Save: %v, Read: %v", fc.name, addDuration, saveDuration, readDuration)

			// Performance assertions
			switch fc.name {
			case "few_files":
				require.Less(t, addDuration+saveDuration, 100*time.Millisecond)
			case "many_files":
				require.Less(t, addDuration+saveDuration, 500*time.Millisecond)
			case "lots_of_files":
				require.Less(t, addDuration+saveDuration, 2*time.Second)
			}
		})
	}
}

// TestMemoryUsage tests memory usage during document operations
func TestMemoryUsage(t *testing.T) {
	if testing.Short() {
		t.Skip("Skipping memory test in short mode")
	}

	// Force garbage collection before test
	runtime.GC()
	runtime.GC()

	var m1, m2 runtime.MemStats
	runtime.ReadMemStats(&m1)

	// Create multiple large documents
	documents := make([]*core.Document, 100)
	for i := 0; i < 100; i++ {
		content := generateHTMLContent(MediumDocumentSize)
		documents[i] = core.NewDocument(
			core.DocumentMetadata{
				Title:    fmt.Sprintf("Memory Test Document %d", i),
				Author:   "Memory Tester",
				Created:  time.Now(),
				Modified: time.Now(),
				Version:  "1.0",
				Language: "en",
			},
			core.DocumentContent{
				HTML: content,
			},
		)
	}

	runtime.ReadMemStats(&m2)

	memoryUsed := m2.Alloc - m1.Alloc
	t.Logf("Memory used for 100 medium documents: %d bytes (%.2f MB)", memoryUsed, float64(memoryUsed)/(1024*1024))

	// Memory should be reasonable (less than 100MB for 100 medium documents)
	require.Less(t, memoryUsed, uint64(100*1024*1024), "Memory usage should be reasonable")

	// Clean up references to allow GC
	for i := range documents {
		documents[i] = nil
	}
	documents = nil

	// Force GC and check memory is released
	runtime.GC()
	runtime.GC()
	time.Sleep(100 * time.Millisecond) // Give GC time to work

	var m3 runtime.MemStats
	runtime.ReadMemStats(&m3)

	// Memory should be mostly released (allowing for some overhead)
	memoryAfterGC := m3.Alloc - m1.Alloc
	t.Logf("Memory after GC: %d bytes (%.2f MB)", memoryAfterGC, float64(memoryAfterGC)/(1024*1024))
	require.Less(t, memoryAfterGC, memoryUsed/2, "Memory should be mostly released after GC")
}

// TestConcurrentOperations tests performance under concurrent load
func TestConcurrentOperations(t *testing.T) {
	if testing.Short() {
		t.Skip("Skipping concurrent test in short mode")
	}

	tempDir, err := ioutil.TempDir("", "liv_concurrent_")
	require.NoError(t, err)
	defer os.RemoveAll(tempDir)

	concurrencyLevels := []int{1, 5, 10, 20}

	for _, concurrency := range concurrencyLevels {
		t.Run(fmt.Sprintf("concurrent_%d", concurrency), func(t *testing.T) {
			start := time.Now()
			done := make(chan bool, concurrency)

			// Start concurrent operations
			for i := 0; i < concurrency; i++ {
				go func(id int) {
					defer func() { done <- true }()

					// Create document
					content := generateHTMLContent(SmallDocumentSize)
					doc := core.NewDocument(
						core.DocumentMetadata{
							Title:    fmt.Sprintf("Concurrent Test %d", id),
							Author:   "Concurrent Tester",
							Created:  time.Now(),
							Modified: time.Now(),
							Version:  "1.0",
							Language: "en",
						},
						core.DocumentContent{
							HTML: content,
						},
					)

					// Validate document
					err := doc.Validate()
					require.NoError(t, err)

					// Create container
					containerPath := filepath.Join(tempDir, fmt.Sprintf("concurrent_%d.liv", id))
					cont := container.NewContainer(containerPath)
					err = cont.AddFile("test.html", []byte(content))
					require.NoError(t, err)
					err = cont.Save()
					require.NoError(t, err)
				}(i)
			}

			// Wait for all operations to complete
			for i := 0; i < concurrency; i++ {
				<-done
			}

			duration := time.Since(start)
			t.Logf("Concurrent operations (level %d): %v", concurrency, duration)

			// Performance should scale reasonably with concurrency
			require.Less(t, duration, 5*time.Second, "Concurrent operations should complete within reasonable time")
		})
	}
}

// TestLargeAssetHandling tests performance with large assets
func TestLargeAssetHandling(t *testing.T) {
	if testing.Short() {
		t.Skip("Skipping large asset test in short mode")
	}

	assetSizes := []struct {
		name string
		size int
	}{
		{"small_asset", 10 * 1024},       // 10KB
		{"medium_asset", 100 * 1024},     // 100KB
		{"large_asset", 1024 * 1024},     // 1MB
		{"huge_asset", 10 * 1024 * 1024}, // 10MB
	}

	tempDir, err := ioutil.TempDir("", "liv_assets_")
	require.NoError(t, err)
	defer os.RemoveAll(tempDir)

	for _, assetSize := range assetSizes {
		t.Run(assetSize.name, func(t *testing.T) {
			// Generate asset data
			assetData := make([]byte, assetSize.size)
			for i := range assetData {
				assetData[i] = byte(rand.Intn(256))
			}

			// Create container with asset
			containerPath := filepath.Join(tempDir, fmt.Sprintf("%s.liv", assetSize.name))
			cont := container.NewContainer(containerPath)

			start := time.Now()
			err := cont.AddFile("large_asset.bin", assetData)
			require.NoError(t, err)
			addDuration := time.Since(start)

			start = time.Now()
			err = cont.Save()
			require.NoError(t, err)
			saveDuration := time.Since(start)

			// Read back the asset
			start = time.Now()
			readContainer, err := container.OpenContainer(containerPath)
			require.NoError(t, err)
			readData, err := readContainer.ReadFile("large_asset.bin")
			require.NoError(t, err)
			readDuration := time.Since(start)

			require.Equal(t, len(assetData), len(readData))

			t.Logf("Asset %s - Add: %v, Save: %v, Read: %v", assetSize.name, addDuration, saveDuration, readDuration)

			// Performance assertions based on asset size
			switch assetSize.name {
			case "small_asset":
				require.Less(t, addDuration+saveDuration+readDuration, 100*time.Millisecond)
			case "medium_asset":
				require.Less(t, addDuration+saveDuration+readDuration, 500*time.Millisecond)
			case "large_asset":
				require.Less(t, addDuration+saveDuration+readDuration, 2*time.Second)
			case "huge_asset":
				require.Less(t, addDuration+saveDuration+readDuration, 10*time.Second)
			}
		})
	}
}

// Benchmark functions

// BenchmarkDocumentValidation benchmarks document validation
func BenchmarkDocumentValidation(b *testing.B) {
	doc := core.NewDocument(
		core.DocumentMetadata{
			Title:    "Benchmark Document",
			Author:   "Benchmark Author",
			Created:  time.Now(),
			Modified: time.Now(),
			Version:  "1.0",
			Language: "en",
		},
		core.DocumentContent{
			HTML: generateHTMLContent(MediumDocumentSize),
		},
	)

	b.ResetTimer()
	for i := 0; i < b.N; i++ {
		_ = doc.Validate()
	}
}

// BenchmarkContainerCreation benchmarks container creation
func BenchmarkContainerCreation(b *testing.B) {
	tempDir, _ := ioutil.TempDir("", "liv_bench_")
	defer os.RemoveAll(tempDir)

	content := generateHTMLContent(SmallDocumentSize)

	b.ResetTimer()
	for i := 0; i < b.N; i++ {
		containerPath := filepath.Join(tempDir, fmt.Sprintf("bench_%d.liv", i))
		cont := container.NewContainer(containerPath)
		_ = cont.AddFile("test.html", []byte(content))
		_ = cont.Save()
	}
}

// BenchmarkManifestValidation benchmarks manifest validation
func BenchmarkManifestValidation(b *testing.B) {
	m := &core.Manifest{
		Version: "1.0",
		Metadata: &core.DocumentMetadata{
			Title:    "Benchmark Manifest",
			Author:   "Benchmark Author",
			Created:  time.Now(),
			Modified: time.Now(),
			Version:  "1.0",
			Language: "en",
		},
		Security: &core.SecurityPolicy{
			WASMPermissions: &core.WASMPermissions{},
			JSPermissions:   &core.JSPermissions{},
			NetworkPolicy:   &core.NetworkPolicy{},
			StoragePolicy:   &core.StoragePolicy{},
		},
		Resources: make(map[string]*core.Resource),
	}

	// Add some resources
	for i := 0; i < 10; i++ {
		m.Resources[fmt.Sprintf("resource_%d", i)] = &core.Resource{
			Type: "asset",
			Path: fmt.Sprintf("assets/resource_%d.png", i),
			Hash: "0000000000000000000000000000000000000000000000000000000000000000",
			Size: 1024,
		}
	}

	b.ResetTimer()
	for i := 0; i < b.N; i++ {
		_ = m.Validate()
	}
}

// BenchmarkJSONSerialization benchmarks JSON serialization/deserialization
func BenchmarkJSONSerialization(b *testing.B) {
	doc := core.NewDocument(
		core.DocumentMetadata{
			Title:    "Benchmark Document",
			Author:   "Benchmark Author",
			Created:  time.Now(),
			Modified: time.Now(),
			Version:  "1.0",
			Language: "en",
		},
		core.DocumentContent{
			HTML: generateHTMLContent(MediumDocumentSize),
		},
	)

	b.Run("marshal", func(b *testing.B) {
		for i := 0; i < b.N; i++ {
			_, _ = doc.MarshalJSON()
		}
	})

	data, _ := doc.MarshalJSON()
	b.Run("unmarshal", func(b *testing.B) {
		for i := 0; i < b.N; i++ {
			newDoc := &core.Document{}
			_ = newDoc.UnmarshalJSON(data)
		}
	})
}

// Helper functions

// generateHTMLContent generates HTML content of specified size
func generateHTMLContent(size int) string {
	var builder strings.Builder
	builder.WriteString("<html><head><title>Generated Content</title></head><body>")

	// Generate content to reach desired size
	contentNeeded := size - builder.Len() - len("</body></html>")
	if contentNeeded > 0 {
		// Create paragraphs with repeated content
		paragraph := "<p>This is generated content for performance testing. "
		paragraph += "It contains various HTML elements and text to simulate real documents. "
		paragraph += "The content is repeated to reach the desired size for testing purposes.</p>"

		for builder.Len() < size-len("</body></html>") {
			if builder.Len()+len(paragraph) > size-len("</body></html>") {
				// Add partial content to reach exact size
				remaining := size - len("</body></html>") - builder.Len()
				builder.WriteString(paragraph[:remaining])
				break
			}
			builder.WriteString(paragraph)
		}
	}

	builder.WriteString("</body></html>")
	return builder.String()
}
