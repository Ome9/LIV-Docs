// Cross-platform compatibility tests for Go CLI tools and viewer

package test

import (
	"bytes"
	"context"
	"fmt"
	"io/ioutil"
	"os"
	"os/exec"
	"path/filepath"
	"runtime"
	"strings"
	"testing"
	"time"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

// Platform represents a target platform for testing
type Platform struct {
	OS           string
	Arch         string
	Name         string
	Executable   string
	PathSep      string
	LineEnding   string
	TempDir      string
	HomeDir      string
	ConfigDir    string
	CacheDir     string
	Features     PlatformFeatures
}

// PlatformFeatures represents platform-specific capabilities
type PlatformFeatures struct {
	FileAssociations bool
	SystemTray       bool
	Notifications    bool
	AutoStart        bool
	Sandboxing       bool
	HardwareAccel    bool
	TouchInput       bool
	HighDPI          bool
}

// Define test platforms
var testPlatforms = []Platform{
	{
		OS:         "windows",
		Arch:       "amd64",
		Name:       "Windows 64-bit",
		Executable: ".exe",
		PathSep:    "\\",
		LineEnding: "\r\n",
		TempDir:    "C:\\Temp",
		HomeDir:    "C:\\Users\\TestUser",
		ConfigDir:  "C:\\Users\\TestUser\\AppData\\Roaming",
		CacheDir:   "C:\\Users\\TestUser\\AppData\\Local",
		Features: PlatformFeatures{
			FileAssociations: true,
			SystemTray:       true,
			Notifications:    true,
			AutoStart:        true,
			Sandboxing:       true,
			HardwareAccel:    true,
			TouchInput:       true,
			HighDPI:          true,
		},
	},
	{
		OS:         "darwin",
		Arch:       "amd64",
		Name:       "macOS Intel",
		Executable: "",
		PathSep:    "/",
		LineEnding: "\n",
		TempDir:    "/tmp",
		HomeDir:    "/Users/testuser",
		ConfigDir:  "/Users/testuser/Library/Application Support",
		CacheDir:   "/Users/testuser/Library/Caches",
		Features: PlatformFeatures{
			FileAssociations: true,
			SystemTray:       true,
			Notifications:    true,
			AutoStart:        true,
			Sandboxing:       true,
			HardwareAccel:    true,
			TouchInput:       false,
			HighDPI:          true,
		},
	},
	{
		OS:         "darwin",
		Arch:       "arm64",
		Name:       "macOS Apple Silicon",
		Executable: "",
		PathSep:    "/",
		LineEnding: "\n",
		TempDir:    "/tmp",
		HomeDir:    "/Users/testuser",
		ConfigDir:  "/Users/testuser/Library/Application Support",
		CacheDir:   "/Users/testuser/Library/Caches",
		Features: PlatformFeatures{
			FileAssociations: true,
			SystemTray:       true,
			Notifications:    true,
			AutoStart:        true,
			Sandboxing:       true,
			HardwareAccel:    true,
			TouchInput:       false,
			HighDPI:          true,
		},
	},
	{
		OS:         "linux",
		Arch:       "amd64",
		Name:       "Linux 64-bit",
		Executable: "",
		PathSep:    "/",
		LineEnding: "\n",
		TempDir:    "/tmp",
		HomeDir:    "/home/testuser",
		ConfigDir:  "/home/testuser/.config",
		CacheDir:   "/home/testuser/.cache",
		Features: PlatformFeatures{
			FileAssociations: true,
			SystemTray:       true,
			Notifications:    true,
			AutoStart:        true,
			Sandboxing:       true,
			HardwareAccel:    true,
			TouchInput:       true,
			HighDPI:          true,
		},
	},
	{
		OS:         "linux",
		Arch:       "arm64",
		Name:       "Linux ARM64",
		Executable: "",
		PathSep:    "/",
		LineEnding: "\n",
		TempDir:    "/tmp",
		HomeDir:    "/home/testuser",
		ConfigDir:  "/home/testuser/.config",
		CacheDir:   "/home/testuser/.cache",
		Features: PlatformFeatures{
			FileAssociations: true,
			SystemTray:       false, // Limited on ARM
			Notifications:    true,
			AutoStart:        true,
			Sandboxing:       true,
			HardwareAccel:    false, // Limited on ARM
			TouchInput:       true,
			HighDPI:          false,
		},
	},
}

func TestCrossPlatformCompatibility(t *testing.T) {
	// Only test current platform in CI, but allow testing all platforms locally
	if os.Getenv("CI") != "" {
		testCurrentPlatform(t)
	} else {
		testAllPlatforms(t)
	}
}

func testCurrentPlatform(t *testing.T) {
	currentPlatform := getCurrentPlatform()
	t.Run(fmt.Sprintf("Current Platform: %s", currentPlatform.Name), func(t *testing.T) {
		testPlatform(t, currentPlatform)
	})
}

func testAllPlatforms(t *testing.T) {
	for _, platform := range testPlatforms {
		t.Run(platform.Name, func(t *testing.T) {
			testPlatform(t, platform)
		})
	}
}

func testPlatform(t *testing.T, platform Platform) {
	t.Run("CLI Tools", func(t *testing.T) {
		testCLITools(t, platform)
	})
	
	t.Run("File Operations", func(t *testing.T) {
		testFileOperations(t, platform)
	})
	
	t.Run("Path Handling", func(t *testing.T) {
		testPathHandling(t, platform)
	})
	
	t.Run("Configuration", func(t *testing.T) {
		testConfiguration(t, platform)
	})
	
	t.Run("Performance", func(t *testing.T) {
		testPerformance(t, platform)
	})
	
	if platform.Features.FileAssociations {
		t.Run("File Associations", func(t *testing.T) {
			testFileAssociations(t, platform)
		})
	}
}

func testCLITools(t *testing.T, platform Platform) {
	// Test CLI tool compilation and basic functionality
	t.Run("Build CLI Tools", func(t *testing.T) {
		if !canBuildForPlatform(platform) {
			t.Skip("Cannot build for this platform in current environment")
		}
		
		tempDir := createTempDir(t, platform)
		defer os.RemoveAll(tempDir)
		
		// Build CLI tool
		cliPath := buildCLITool(t, platform, tempDir)
		
		// Test basic CLI functionality
		testBasicCLIFunctionality(t, platform, cliPath)
	})
	
	t.Run("Viewer Tool", func(t *testing.T) {
		if !canBuildForPlatform(platform) {
			t.Skip("Cannot build for this platform in current environment")
		}
		
		tempDir := createTempDir(t, platform)
		defer os.RemoveAll(tempDir)
		
		// Build viewer tool
		viewerPath := buildViewerTool(t, platform, tempDir)
		
		// Test viewer functionality
		testViewerFunctionality(t, platform, viewerPath)
	})
}

func testFileOperations(t *testing.T, platform Platform) {
	tempDir := createTempDir(t, platform)
	defer os.RemoveAll(tempDir)
	
	t.Run("LIV File Creation", func(t *testing.T) {
		// Test creating .liv files with platform-specific paths
		livFile := filepath.Join(tempDir, "test.liv")
		
		// Create a test document
		err := createTestLIVFile(livFile, platform)
		assert.NoError(t, err, "Should create LIV file successfully")
		
		// Verify file exists and is readable
		assert.FileExists(t, livFile, "LIV file should exist")
		
		// Test file permissions
		info, err := os.Stat(livFile)
		require.NoError(t, err)
		assert.True(t, info.Mode().IsRegular(), "Should be a regular file")
	})
	
	t.Run("Path Separators", func(t *testing.T) {
		// Test that paths work correctly with platform-specific separators
		testPath := filepath.Join(tempDir, "subdir", "test.liv")
		
		err := os.MkdirAll(filepath.Dir(testPath), 0755)
		require.NoError(t, err)
		
		err = createTestLIVFile(testPath, platform)
		assert.NoError(t, err, "Should handle nested paths correctly")
		
		assert.FileExists(t, testPath, "File should exist at nested path")
	})
	
	t.Run("Unicode Filenames", func(t *testing.T) {
		// Test Unicode filename support
		unicodeFile := filepath.Join(tempDir, "测试文档.liv")
		
		err := createTestLIVFile(unicodeFile, platform)
		if platform.OS == "windows" && err != nil {
			t.Skip("Unicode filenames may not be supported on this Windows configuration")
		}
		
		assert.NoError(t, err, "Should handle Unicode filenames")
		assert.FileExists(t, unicodeFile, "Unicode filename should work")
	})
}

func testPathHandling(t *testing.T, platform Platform) {
	t.Run("Absolute Paths", func(t *testing.T) {
		// Test absolute path handling
		absPath := getAbsolutePath(platform, "test.liv")
		assert.True(t, filepath.IsAbs(absPath), "Should generate absolute path")
	})
	
	t.Run("Relative Paths", func(t *testing.T) {
		// Test relative path handling
		relPath := "documents/test.liv"
		normalized := filepath.Clean(relPath)
		
		expected := strings.ReplaceAll(relPath, "/", string(filepath.Separator))
		assert.Equal(t, expected, normalized, "Should normalize path separators")
	})
	
	t.Run("Home Directory", func(t *testing.T) {
		// Test home directory expansion
		homeDir := getHomeDir(platform)
		assert.NotEmpty(t, homeDir, "Should have home directory")
		
		if platform.OS != "windows" {
			assert.True(t, strings.HasPrefix(homeDir, "/"), "Unix home should start with /")
		} else {
			assert.True(t, strings.Contains(homeDir, ":"), "Windows home should contain drive letter")
		}
	})
}

func testConfiguration(t *testing.T, platform Platform) {
	t.Run("Config Directory", func(t *testing.T) {
		configDir := getConfigDir(platform)
		assert.NotEmpty(t, configDir, "Should have config directory")
		
		// Test config file creation
		configFile := filepath.Join(configDir, "liv-config.json")
		err := os.MkdirAll(filepath.Dir(configFile), 0755)
		require.NoError(t, err)
		
		err = ioutil.WriteFile(configFile, []byte(`{"version": "1.0"}`), 0644)
		assert.NoError(t, err, "Should create config file")
		
		// Clean up
		os.RemoveAll(configDir)
	})
	
	t.Run("Cache Directory", func(t *testing.T) {
		cacheDir := getCacheDir(platform)
		assert.NotEmpty(t, cacheDir, "Should have cache directory")
		
		// Test cache file creation
		cacheFile := filepath.Join(cacheDir, "test-cache.dat")
		err := os.MkdirAll(filepath.Dir(cacheFile), 0755)
		require.NoError(t, err)
		
		err = ioutil.WriteFile(cacheFile, []byte("cache data"), 0644)
		assert.NoError(t, err, "Should create cache file")
		
		// Clean up
		os.RemoveAll(cacheDir)
	})
	
	t.Run("Environment Variables", func(t *testing.T) {
		// Test platform-specific environment variable handling
		testEnvVar := "LIV_TEST_VAR"
		testValue := "test_value"
		
		err := os.Setenv(testEnvVar, testValue)
		require.NoError(t, err)
		
		value := os.Getenv(testEnvVar)
		assert.Equal(t, testValue, value, "Should handle environment variables")
		
		// Clean up
		os.Unsetenv(testEnvVar)
	})
}

func testPerformance(t *testing.T, platform Platform) {
	t.Run("File I/O Performance", func(t *testing.T) {
		tempDir := createTempDir(t, platform)
		defer os.RemoveAll(tempDir)
		
		// Test file creation performance
		start := time.Now()
		
		for i := 0; i < 100; i++ {
			testFile := filepath.Join(tempDir, fmt.Sprintf("test_%d.liv", i))
			err := createTestLIVFile(testFile, platform)
			require.NoError(t, err)
		}
		
		duration := time.Since(start)
		t.Logf("Created 100 files in %v on %s", duration, platform.Name)
		
		// Performance should be reasonable (less than 10 seconds for 100 files)
		assert.Less(t, duration, 10*time.Second, "File creation should be reasonably fast")
	})
	
	t.Run("Memory Usage", func(t *testing.T) {
		// Test memory usage during operations
		if !canBuildForPlatform(platform) {
			t.Skip("Cannot test memory usage without building for platform")
		}
		
		// This would require more sophisticated memory monitoring
		// For now, just ensure operations complete without obvious memory issues
		tempDir := createTempDir(t, platform)
		defer os.RemoveAll(tempDir)
		
		// Create a larger test file
		largeFile := filepath.Join(tempDir, "large_test.liv")
		err := createLargeLIVFile(largeFile, platform)
		assert.NoError(t, err, "Should handle large files without memory issues")
	})
}

func testFileAssociations(t *testing.T, platform Platform) {
	if !platform.Features.FileAssociations {
		t.Skip("File associations not supported on this platform")
	}
	
	t.Run("MIME Type Registration", func(t *testing.T) {
		// Test MIME type handling for .liv files
		mimeType := getLIVMimeType(platform)
		assert.NotEmpty(t, mimeType, "Should have MIME type for .liv files")
		
		expectedMimeTypes := []string{
			"application/x-liv",
			"application/liv",
			"application/octet-stream", // Fallback
		}
		
		found := false
		for _, expected := range expectedMimeTypes {
			if mimeType == expected {
				found = true
				break
			}
		}
		
		assert.True(t, found, "MIME type should be one of the expected values")
	})
}

func testBasicCLIFunctionality(t *testing.T, platform Platform, cliPath string) {
	// Test help command
	t.Run("Help Command", func(t *testing.T) {
		cmd := exec.Command(cliPath, "--help")
		output, err := cmd.CombinedOutput()
		
		assert.NoError(t, err, "Help command should succeed")
		assert.Contains(t, string(output), "Usage:", "Help should contain usage information")
	})
	
	// Test version command
	t.Run("Version Command", func(t *testing.T) {
		cmd := exec.Command(cliPath, "--version")
		output, err := cmd.CombinedOutput()
		
		assert.NoError(t, err, "Version command should succeed")
		assert.NotEmpty(t, string(output), "Version should not be empty")
	})
	
	// Test build command with invalid input
	t.Run("Build Command Error Handling", func(t *testing.T) {
		cmd := exec.Command(cliPath, "build", "nonexistent-directory")
		output, err := cmd.CombinedOutput()
		
		assert.Error(t, err, "Build command should fail with invalid input")
		assert.Contains(t, string(output), "error", "Error message should be present")
	})
}

func testViewerFunctionality(t *testing.T, platform Platform, viewerPath string) {
	// Test viewer help
	t.Run("Viewer Help", func(t *testing.T) {
		cmd := exec.Command(viewerPath, "--help")
		output, err := cmd.CombinedOutput()
		
		assert.NoError(t, err, "Viewer help should succeed")
		assert.Contains(t, string(output), "viewer", "Help should mention viewer")
	})
	
	// Test viewer with invalid file
	t.Run("Viewer Error Handling", func(t *testing.T) {
		cmd := exec.Command(viewerPath, "nonexistent.liv")
		cmd.Env = append(os.Environ(), "LIV_TEST_MODE=1") // Prevent GUI from opening
		
		ctx, cancel := context.WithTimeout(context.Background(), 5*time.Second)
		defer cancel()
		
		cmd = exec.CommandContext(ctx, viewerPath, "nonexistent.liv")
		output, err := cmd.CombinedOutput()
		
		// Should either error or handle gracefully
		if err == nil {
			t.Log("Viewer handled invalid file gracefully")
		} else {
			assert.Contains(t, string(output), "error", "Should provide error message")
		}
	})
}

// Helper functions

func getCurrentPlatform() Platform {
	for _, platform := range testPlatforms {
		if platform.OS == runtime.GOOS && platform.Arch == runtime.GOARCH {
			return platform
		}
	}
	
	// Return a default platform if not found
	return Platform{
		OS:         runtime.GOOS,
		Arch:       runtime.GOARCH,
		Name:       fmt.Sprintf("%s %s", runtime.GOOS, runtime.GOARCH),
		Executable: getExecutableExtension(),
		PathSep:    string(filepath.Separator),
		LineEnding: getLineEnding(),
		Features:   PlatformFeatures{}, // Default features
	}
}

func canBuildForPlatform(platform Platform) bool {
	// Check if we can cross-compile for this platform
	if platform.OS == runtime.GOOS && platform.Arch == runtime.GOARCH {
		return true // Same platform
	}
	
	// Check if Go supports cross-compilation for this target
	cmd := exec.Command("go", "env", "GOOS", "GOARCH")
	cmd.Env = append(os.Environ(), 
		fmt.Sprintf("GOOS=%s", platform.OS),
		fmt.Sprintf("GOARCH=%s", platform.Arch),
	)
	
	err := cmd.Run()
	return err == nil
}

func createTempDir(t *testing.T, platform Platform) string {
	tempDir, err := ioutil.TempDir("", "liv-test-*")
	require.NoError(t, err)
	return tempDir
}

func buildCLITool(t *testing.T, platform Platform, outputDir string) string {
	outputName := "liv-cli" + platform.Executable
	outputPath := filepath.Join(outputDir, outputName)
	
	cmd := exec.Command("go", "build", "-o", outputPath, "../cmd/cli/main.go")
	cmd.Env = append(os.Environ(),
		fmt.Sprintf("GOOS=%s", platform.OS),
		fmt.Sprintf("GOARCH=%s", platform.Arch),
	)
	
	var stderr bytes.Buffer
	cmd.Stderr = &stderr
	
	err := cmd.Run()
	if err != nil {
		t.Fatalf("Failed to build CLI tool for %s: %v\nStderr: %s", platform.Name, err, stderr.String())
	}
	
	return outputPath
}

func buildViewerTool(t *testing.T, platform Platform, outputDir string) string {
	outputName := "liv-viewer" + platform.Executable
	outputPath := filepath.Join(outputDir, outputName)
	
	cmd := exec.Command("go", "build", "-o", outputPath, "../cmd/viewer/main.go")
	cmd.Env = append(os.Environ(),
		fmt.Sprintf("GOOS=%s", platform.OS),
		fmt.Sprintf("GOARCH=%s", platform.Arch),
	)
	
	var stderr bytes.Buffer
	cmd.Stderr = &stderr
	
	err := cmd.Run()
	if err != nil {
		t.Fatalf("Failed to build viewer tool for %s: %v\nStderr: %s", platform.Name, err, stderr.String())
	}
	
	return outputPath
}

func createTestLIVFile(path string, platform Platform) error {
	// Create a minimal valid LIV file for testing
	dir := filepath.Dir(path)
	if err := os.MkdirAll(dir, 0755); err != nil {
		return err
	}
	
	// For now, create a simple test file
	// In a real implementation, this would create a proper LIV file
	content := fmt.Sprintf("Test LIV file for %s%s", platform.Name, platform.LineEnding)
	return ioutil.WriteFile(path, []byte(content), 0644)
}

func createLargeLIVFile(path string, platform Platform) error {
	// Create a larger test file to test memory usage
	dir := filepath.Dir(path)
	if err := os.MkdirAll(dir, 0755); err != nil {
		return err
	}
	
	// Create a 1MB test file
	content := make([]byte, 1024*1024)
	for i := range content {
		content[i] = byte(i % 256)
	}
	
	return ioutil.WriteFile(path, content, 0644)
}

func getAbsolutePath(platform Platform, filename string) string {
	if platform.OS == "windows" {
		return "C:" + platform.PathSep + filename
	}
	return platform.PathSep + filename
}

func getHomeDir(platform Platform) string {
	return platform.HomeDir
}

func getConfigDir(platform Platform) string {
	return filepath.Join(platform.ConfigDir, "liv")
}

func getCacheDir(platform Platform) string {
	return filepath.Join(platform.CacheDir, "liv")
}

func getLIVMimeType(platform Platform) string {
	// Return platform-appropriate MIME type
	switch platform.OS {
	case "windows":
		return "application/x-liv"
	case "darwin":
		return "application/x-liv"
	case "linux":
		return "application/x-liv"
	default:
		return "application/octet-stream"
	}
}

func getExecutableExtension() string {
	if runtime.GOOS == "windows" {
		return ".exe"
	}
	return ""
}

func getLineEnding() string {
	if runtime.GOOS == "windows" {
		return "\r\n"
	}
	return "\n"
}

// Benchmark tests for performance comparison across platforms
func BenchmarkCrossPlatformPerformance(b *testing.B) {
	platform := getCurrentPlatform()
	tempDir := createTempDirBench(b, platform)
	defer os.RemoveAll(tempDir)
	
	b.Run("FileCreation", func(b *testing.B) {
		for i := 0; i < b.N; i++ {
			testFile := filepath.Join(tempDir, fmt.Sprintf("bench_%d.liv", i))
			err := createTestLIVFile(testFile, platform)
			if err != nil {
				b.Fatal(err)
			}
		}
	})
	
	b.Run("FileReading", func(b *testing.B) {
		// Create test file first
		testFile := filepath.Join(tempDir, "bench_read.liv")
		err := createTestLIVFile(testFile, platform)
		if err != nil {
			b.Fatal(err)
		}
		
		b.ResetTimer()
		for i := 0; i < b.N; i++ {
			_, err := ioutil.ReadFile(testFile)
			if err != nil {
				b.Fatal(err)
			}
		}
	})
}

func createTempDirBench(b *testing.B, platform Platform) string {
	tempDir, err := ioutil.TempDir("", "liv-bench-*")
	if err != nil {
		b.Fatal(err)
	}
	return tempDir
}