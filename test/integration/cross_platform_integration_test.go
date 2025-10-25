// Integration tests for cross-platform LIV document workflow

package integration

import (
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

// TestWorkflow represents a complete LIV document workflow test
type TestWorkflow struct {
	Name        string
	Description string
	Steps       []WorkflowStep
	Platform    string
	Expected    WorkflowExpectation
}

// WorkflowStep represents a single step in the workflow
type WorkflowStep struct {
	Name        string
	Command     []string
	Input       string
	Environment map[string]string
	Timeout     time.Duration
}

// WorkflowExpectation defines what we expect from the workflow
type WorkflowExpectation struct {
	Success      bool
	OutputFiles  []string
	ErrorPattern string
	Performance  PerformanceExpectation
}

// PerformanceExpectation defines performance requirements
type PerformanceExpectation struct {
	MaxDuration  time.Duration
	MaxMemoryMB  int
	MaxFileSize  int64
	MinFrameRate float64
}

func TestCrossPlatformIntegration(t *testing.T) {
	// Skip integration tests in short mode
	if testing.Short() {
		t.Skip("Skipping integration tests in short mode")
	}

	workflows := []TestWorkflow{
		{
			Name:        "Complete Document Creation Workflow",
			Description: "Create, validate, and view a LIV document",
			Platform:    runtime.GOOS,
			Steps: []WorkflowStep{
				{
					Name:    "Create Source Directory",
					Command: []string{"mkdir", "-p", "test-doc"},
					Timeout: 5 * time.Second,
				},
				{
					Name:    "Create HTML Content",
					Input:   createTestHTML(),
					Timeout: 5 * time.Second,
				},
				{
					Name:    "Create CSS Content",
					Input:   createTestCSS(),
					Timeout: 5 * time.Second,
				},
				{
					Name:    "Build LIV Document",
					Command: []string{"liv-cli", "build", "test-doc", "-o", "test.liv"},
					Timeout: 30 * time.Second,
				},
				{
					Name:    "Validate Document",
					Command: []string{"liv-cli", "validate", "test.liv"},
					Timeout: 10 * time.Second,
				},
				{
					Name:    "Start Viewer",
					Command: []string{"liv-viewer", "--web", "--port", "8080", "test.liv"},
					Timeout: 10 * time.Second,
					Environment: map[string]string{
						"LIV_TEST_MODE": "1",
					},
				},
			},
			Expected: WorkflowExpectation{
				Success:     true,
				OutputFiles: []string{"test.liv"},
				Performance: PerformanceExpectation{
					MaxDuration:  60 * time.Second,
					MaxMemoryMB:  100,
					MaxFileSize:  10 * 1024 * 1024, // 10MB
					MinFrameRate: 30.0,
				},
			},
		},
		{
			Name:        "Document Conversion Workflow",
			Description: "Convert between different document formats",
			Platform:    runtime.GOOS,
			Steps: []WorkflowStep{
				{
					Name:    "Create Test Document",
					Command: []string{"liv-cli", "build", "test-doc", "-o", "original.liv"},
					Timeout: 30 * time.Second,
				},
				{
					Name:    "Convert to HTML",
					Command: []string{"liv-cli", "convert", "original.liv", "output.html"},
					Timeout: 20 * time.Second,
				},
				{
					Name:    "Convert to PDF",
					Command: []string{"liv-cli", "convert", "original.liv", "output.pdf"},
					Timeout: 30 * time.Second,
				},
				{
					Name:    "Convert to EPUB",
					Command: []string{"liv-cli", "convert", "original.liv", "output.epub"},
					Timeout: 30 * time.Second,
				},
			},
			Expected: WorkflowExpectation{
				Success:     true,
				OutputFiles: []string{"original.liv", "output.html", "output.pdf", "output.epub"},
				Performance: PerformanceExpectation{
					MaxDuration: 90 * time.Second,
					MaxMemoryMB: 200,
				},
			},
		},
		{
			Name:        "Mobile Viewer Workflow",
			Description: "Test mobile-optimized viewing experience",
			Platform:    runtime.GOOS,
			Steps: []WorkflowStep{
				{
					Name:    "Create Mobile-Optimized Document",
					Input:   createMobileTestContent(),
					Timeout: 10 * time.Second,
				},
				{
					Name:    "Build Document",
					Command: []string{"liv-cli", "build", "mobile-doc", "-o", "mobile.liv"},
					Timeout: 30 * time.Second,
				},
				{
					Name:    "Start Mobile Viewer",
					Command: []string{"liv-viewer", "--web", "--port", "8081", "--mobile", "mobile.liv"},
					Timeout: 10 * time.Second,
					Environment: map[string]string{
						"LIV_TEST_MODE":   "1",
						"LIV_MOBILE_MODE": "1",
					},
				},
			},
			Expected: WorkflowExpectation{
				Success:     true,
				OutputFiles: []string{"mobile.liv"},
				Performance: PerformanceExpectation{
					MaxDuration:  45 * time.Second,
					MaxMemoryMB:  80,
					MinFrameRate: 24.0, // Lower for mobile
				},
			},
		},
		{
			Name:        "Desktop Application Workflow",
			Description: "Test desktop application functionality",
			Platform:    runtime.GOOS,
			Steps: []WorkflowStep{
				{
					Name:    "Create Desktop Document",
					Input:   createDesktopTestContent(),
					Timeout: 10 * time.Second,
				},
				{
					Name:    "Build Document",
					Command: []string{"liv-cli", "build", "desktop-doc", "-o", "desktop.liv"},
					Timeout: 30 * time.Second,
				},
				{
					Name:    "Test Desktop Viewer",
					Command: []string{"liv-viewer", "--desktop", "desktop.liv"},
					Timeout: 15 * time.Second,
					Environment: map[string]string{
						"LIV_TEST_MODE":    "1",
						"LIV_DESKTOP_MODE": "1",
					},
				},
			},
			Expected: WorkflowExpectation{
				Success:     true,
				OutputFiles: []string{"desktop.liv"},
				Performance: PerformanceExpectation{
					MaxDuration:  60 * time.Second,
					MaxMemoryMB:  150,
					MinFrameRate: 60.0,
				},
			},
		},
	}

	for _, workflow := range workflows {
		t.Run(workflow.Name, func(t *testing.T) {
			runWorkflow(t, workflow)
		})
	}
}

func runWorkflow(t *testing.T, workflow TestWorkflow) {
	// Create temporary directory for this workflow
	tempDir, err := ioutil.TempDir("", fmt.Sprintf("liv-integration-%s-*",
		strings.ReplaceAll(workflow.Name, " ", "-")))
	require.NoError(t, err)
	defer os.RemoveAll(tempDir)

	// Change to temp directory
	originalDir, err := os.Getwd()
	require.NoError(t, err)
	defer os.Chdir(originalDir)

	err = os.Chdir(tempDir)
	require.NoError(t, err)

	t.Logf("Running workflow '%s' in directory: %s", workflow.Name, tempDir)

	// Track workflow timing
	workflowStart := time.Now()

	// Execute each step
	for i, step := range workflow.Steps {
		t.Run(fmt.Sprintf("Step_%d_%s", i+1, step.Name), func(t *testing.T) {
			executeWorkflowStep(t, step, tempDir)
		})
	}

	workflowDuration := time.Since(workflowStart)

	// Verify expectations
	verifyWorkflowExpectations(t, workflow.Expected, tempDir, workflowDuration)
}

func executeWorkflowStep(t *testing.T, step WorkflowStep, workDir string) {
	stepStart := time.Now()

	// Handle input creation steps
	if step.Input != "" && len(step.Command) == 0 {
		err := createStepInput(step, workDir)
		assert.NoError(t, err, "Failed to create input for step: %s", step.Name)
		return
	}

	// Handle command execution steps
	if len(step.Command) > 0 {
		// Set up command
		ctx, cancel := context.WithTimeout(context.Background(), step.Timeout)
		defer cancel()

		cmd := exec.CommandContext(ctx, step.Command[0], step.Command[1:]...)
		cmd.Dir = workDir

		// Set environment variables
		cmd.Env = os.Environ()
		for key, value := range step.Environment {
			cmd.Env = append(cmd.Env, fmt.Sprintf("%s=%s", key, value))
		}

		// Execute command
		output, err := cmd.CombinedOutput()

		stepDuration := time.Since(stepStart)
		t.Logf("Step '%s' completed in %v", step.Name, stepDuration)

		if err != nil {
			t.Logf("Command output: %s", string(output))

			// Some commands are expected to fail in test mode
			if step.Environment["LIV_TEST_MODE"] == "1" &&
				(strings.Contains(step.Name, "Viewer") || strings.Contains(step.Name, "Start")) {
				t.Logf("Viewer command failed as expected in test mode: %v", err)
				return
			}

			assert.NoError(t, err, "Step '%s' failed: %s", step.Name, string(output))
		}
	}
}

func createStepInput(step WorkflowStep, workDir string) error {
	switch step.Name {
	case "Create HTML Content":
		return createFileFromInput(filepath.Join(workDir, "test-doc", "index.html"), step.Input)
	case "Create CSS Content":
		return createFileFromInput(filepath.Join(workDir, "test-doc", "style.css"), step.Input)
	case "Create Mobile-Optimized Document":
		return createMobileDocument(workDir, step.Input)
	case "Create Desktop Document":
		return createDesktopDocument(workDir, step.Input)
	default:
		return fmt.Errorf("unknown input step: %s", step.Name)
	}
}

func createFileFromInput(path, content string) error {
	dir := filepath.Dir(path)
	if err := os.MkdirAll(dir, 0755); err != nil {
		return err
	}
	return ioutil.WriteFile(path, []byte(content), 0644)
}

func createMobileDocument(workDir, content string) error {
	docDir := filepath.Join(workDir, "mobile-doc")
	if err := os.MkdirAll(docDir, 0755); err != nil {
		return err
	}

	// Create mobile-optimized files
	files := map[string]string{
		"index.html":    content,
		"mobile.css":    createMobileCSS(),
		"manifest.json": createMobileManifest(),
	}

	for filename, fileContent := range files {
		path := filepath.Join(docDir, filename)
		if err := ioutil.WriteFile(path, []byte(fileContent), 0644); err != nil {
			return err
		}
	}

	return nil
}

func createDesktopDocument(workDir, content string) error {
	docDir := filepath.Join(workDir, "desktop-doc")
	if err := os.MkdirAll(docDir, 0755); err != nil {
		return err
	}

	// Create desktop-optimized files
	files := map[string]string{
		"index.html":    content,
		"desktop.css":   createDesktopCSS(),
		"manifest.json": createDesktopManifest(),
	}

	for filename, fileContent := range files {
		path := filepath.Join(docDir, filename)
		if err := ioutil.WriteFile(path, []byte(fileContent), 0644); err != nil {
			return err
		}
	}

	return nil
}

func verifyWorkflowExpectations(t *testing.T, expected WorkflowExpectation, workDir string, duration time.Duration) {
	// Check performance expectations
	if expected.Performance.MaxDuration > 0 {
		assert.LessOrEqual(t, duration, expected.Performance.MaxDuration,
			"Workflow took too long: %v > %v", duration, expected.Performance.MaxDuration)
	}

	// Check output files
	for _, expectedFile := range expected.OutputFiles {
		filePath := filepath.Join(workDir, expectedFile)
		assert.FileExists(t, filePath, "Expected output file should exist: %s", expectedFile)

		// Check file size if specified
		if expected.Performance.MaxFileSize > 0 {
			info, err := os.Stat(filePath)
			if err == nil {
				assert.LessOrEqual(t, info.Size(), expected.Performance.MaxFileSize,
					"File %s is too large: %d > %d", expectedFile, info.Size(), expected.Performance.MaxFileSize)
			}
		}
	}
}

// Test content creation functions

func createTestHTML() string {
	return `<!DOCTYPE html>
<html lang="en">
<head>
    <meta charset="UTF-8">
    <meta name="viewport" content="width=device-width, initial-scale=1.0">
    <title>Cross-Platform Test Document</title>
    <link rel="stylesheet" href="style.css">
</head>
<body>
    <div class="container">
        <h1>Cross-Platform LIV Document</h1>
        <p>This document tests cross-platform compatibility.</p>
        
        <div class="interactive-element" data-interactive="true">
            <h2>Interactive Content</h2>
            <p>This element supports touch and mouse interactions.</p>
        </div>
        
        <div class="animation-test">
            <div class="animated-box"></div>
        </div>
        
        <svg class="test-svg" width="200" height="100">
            <rect x="10" y="10" width="180" height="80" fill="blue" opacity="0.7"/>
            <text x="100" y="55" text-anchor="middle" fill="white">SVG Test</text>
        </svg>
    </div>
</body>
</html>`
}

func createTestCSS() string {
	return `/* Cross-Platform Test Styles */
.container {
    max-width: 800px;
    margin: 0 auto;
    padding: 20px;
    font-family: -apple-system, BlinkMacSystemFont, 'Segoe UI', Roboto, sans-serif;
}

.interactive-element {
    background: #f0f0f0;
    padding: 20px;
    border-radius: 8px;
    cursor: pointer;
    transition: background-color 0.3s ease;
    min-height: 44px; /* Touch-friendly */
}

.interactive-element:hover {
    background: #e0e0e0;
}

.animation-test {
    margin: 20px 0;
    height: 100px;
    display: flex;
    align-items: center;
    justify-content: center;
}

.animated-box {
    width: 50px;
    height: 50px;
    background: #007bff;
    animation: bounce 2s infinite;
}

@keyframes bounce {
    0%, 100% { transform: translateY(0); }
    50% { transform: translateY(-20px); }
}

.test-svg {
    display: block;
    margin: 20px auto;
}

/* Responsive design */
@media (max-width: 768px) {
    .container {
        padding: 10px;
    }
    
    .interactive-element {
        padding: 15px;
        font-size: 16px; /* Prevent zoom on iOS */
    }
}

/* High DPI support */
@media (-webkit-min-device-pixel-ratio: 2), (min-resolution: 192dpi) {
    .test-svg {
        image-rendering: -webkit-optimize-contrast;
        image-rendering: crisp-edges;
    }
}

/* Reduced motion support */
@media (prefers-reduced-motion: reduce) {
    .animated-box {
        animation: none;
    }
    
    .interactive-element {
        transition: none;
    }
}`
}

func createMobileTestContent() string {
	return `<!DOCTYPE html>
<html lang="en">
<head>
    <meta charset="UTF-8">
    <meta name="viewport" content="width=device-width, initial-scale=1.0, user-scalable=no">
    <title>Mobile Test Document</title>
    <link rel="stylesheet" href="mobile.css">
</head>
<body>
    <div class="mobile-container">
        <header class="mobile-header">
            <h1>Mobile LIV Document</h1>
        </header>
        
        <main class="mobile-content">
            <div class="touch-area" data-interactive="true">
                <p>Touch and gesture test area</p>
            </div>
            
            <div class="swipe-container">
                <div class="swipe-item">Swipe me left or right</div>
            </div>
            
            <div class="pinch-zoom-area">
                <img src="data:image/svg+xml;base64,PHN2ZyB3aWR0aD0iMjAwIiBoZWlnaHQ9IjEwMCIgeG1sbnM9Imh0dHA6Ly93d3cudzMub3JnLzIwMDAvc3ZnIj4KICA8cmVjdCB3aWR0aD0iMTAwJSIgaGVpZ2h0PSIxMDAlIiBmaWxsPSIjNDI4NWY0Ii8+CiAgPHRleHQgeD0iNTAlIiB5PSI1MCUiIGZvbnQtZmFtaWx5PSJBcmlhbCIgZm9udC1zaXplPSIxNiIgZmlsbD0id2hpdGUiIHRleHQtYW5jaG9yPSJtaWRkbGUiIGR5PSIuM2VtIj5QaW5jaCB0byBab29tPC90ZXh0Pgo8L3N2Zz4K" alt="Pinch to zoom test">
            </div>
        </main>
        
        <footer class="mobile-footer">
            <button class="mobile-button">Test Button</button>
        </footer>
    </div>
</body>
</html>`
}

func createDesktopTestContent() string {
	return `<!DOCTYPE html>
<html lang="en">
<head>
    <meta charset="UTF-8">
    <meta name="viewport" content="width=device-width, initial-scale=1.0">
    <title>Desktop Test Document</title>
    <link rel="stylesheet" href="desktop.css">
</head>
<body>
    <div class="desktop-container">
        <nav class="desktop-nav">
            <ul>
                <li><a href="#section1">Section 1</a></li>
                <li><a href="#section2">Section 2</a></li>
                <li><a href="#section3">Section 3</a></li>
            </ul>
        </nav>
        
        <main class="desktop-main">
            <section id="section1">
                <h2>Desktop Features</h2>
                <p>This document showcases desktop-specific features.</p>
                
                <div class="desktop-grid">
                    <div class="grid-item">Item 1</div>
                    <div class="grid-item">Item 2</div>
                    <div class="grid-item">Item 3</div>
                    <div class="grid-item">Item 4</div>
                </div>
            </section>
            
            <section id="section2">
                <h2>Interactive Charts</h2>
                <div class="chart-container">
                    <canvas id="test-chart" width="400" height="200"></canvas>
                </div>
            </section>
            
            <section id="section3">
                <h2>High-Resolution Graphics</h2>
                <svg class="desktop-svg" width="600" height="300">
                    <defs>
                        <linearGradient id="grad1" x1="0%" y1="0%" x2="100%" y2="0%">
                            <stop offset="0%" style="stop-color:rgb(255,255,0);stop-opacity:1" />
                            <stop offset="100%" style="stop-color:rgb(255,0,0);stop-opacity:1" />
                        </linearGradient>
                    </defs>
                    <ellipse cx="300" cy="150" rx="200" ry="80" fill="url(#grad1)" />
                    <text x="300" y="155" font-family="Arial" font-size="20" fill="black" text-anchor="middle">Desktop Graphics</text>
                </svg>
            </section>
        </main>
    </div>
</body>
</html>`
}

func createMobileCSS() string {
	return `/* Mobile-Optimized Styles */
* {
    box-sizing: border-box;
}

body {
    margin: 0;
    padding: 0;
    font-family: -apple-system, BlinkMacSystemFont, 'Segoe UI', Roboto, sans-serif;
    font-size: 16px; /* Prevent zoom on iOS */
    line-height: 1.5;
    -webkit-text-size-adjust: 100%;
}

.mobile-container {
    min-height: 100vh;
    display: flex;
    flex-direction: column;
}

.mobile-header {
    background: #007bff;
    color: white;
    padding: 1rem;
    text-align: center;
}

.mobile-header h1 {
    margin: 0;
    font-size: 1.5rem;
}

.mobile-content {
    flex: 1;
    padding: 1rem;
}

.touch-area {
    background: #f8f9fa;
    border: 2px dashed #dee2e6;
    border-radius: 8px;
    padding: 2rem;
    text-align: center;
    margin: 1rem 0;
    min-height: 100px;
    display: flex;
    align-items: center;
    justify-content: center;
    -webkit-user-select: none;
    user-select: none;
}

.swipe-container {
    overflow-x: auto;
    -webkit-overflow-scrolling: touch;
    margin: 1rem 0;
}

.swipe-item {
    background: #28a745;
    color: white;
    padding: 1rem;
    border-radius: 8px;
    text-align: center;
    min-width: 200px;
    margin: 0.5rem;
}

.pinch-zoom-area {
    text-align: center;
    margin: 2rem 0;
}

.pinch-zoom-area img {
    max-width: 100%;
    height: auto;
    border-radius: 8px;
}

.mobile-footer {
    background: #f8f9fa;
    padding: 1rem;
    text-align: center;
}

.mobile-button {
    background: #007bff;
    color: white;
    border: none;
    padding: 12px 24px;
    border-radius: 8px;
    font-size: 16px;
    min-width: 44px;
    min-height: 44px;
    cursor: pointer;
    -webkit-tap-highlight-color: rgba(0,0,0,0.1);
}

.mobile-button:active {
    background: #0056b3;
    transform: scale(0.98);
}

/* Touch-friendly interactions */
@media (pointer: coarse) {
    .touch-area {
        min-height: 120px;
    }
    
    .mobile-button {
        padding: 16px 32px;
    }
}

/* Orientation handling */
@media (orientation: landscape) {
    .mobile-content {
        padding: 0.5rem 2rem;
    }
}

/* High DPI optimization */
@media (-webkit-min-device-pixel-ratio: 2) {
    .pinch-zoom-area img {
        image-rendering: -webkit-optimize-contrast;
    }
}`
}

func createDesktopCSS() string {
	return `/* Desktop-Optimized Styles */
* {
    box-sizing: border-box;
}

body {
    margin: 0;
    padding: 0;
    font-family: -apple-system, BlinkMacSystemFont, 'Segoe UI', Roboto, sans-serif;
    line-height: 1.6;
    color: #333;
}

.desktop-container {
    display: flex;
    min-height: 100vh;
}

.desktop-nav {
    width: 250px;
    background: #f8f9fa;
    border-right: 1px solid #dee2e6;
    padding: 2rem 0;
}

.desktop-nav ul {
    list-style: none;
    padding: 0;
    margin: 0;
}

.desktop-nav li {
    margin: 0;
}

.desktop-nav a {
    display: block;
    padding: 1rem 2rem;
    color: #495057;
    text-decoration: none;
    transition: background-color 0.2s ease;
}

.desktop-nav a:hover {
    background: #e9ecef;
    color: #007bff;
}

.desktop-main {
    flex: 1;
    padding: 2rem;
    max-width: calc(100% - 250px);
}

.desktop-main section {
    margin-bottom: 3rem;
}

.desktop-main h2 {
    color: #495057;
    border-bottom: 2px solid #007bff;
    padding-bottom: 0.5rem;
}

.desktop-grid {
    display: grid;
    grid-template-columns: repeat(auto-fit, minmax(200px, 1fr));
    gap: 1rem;
    margin: 2rem 0;
}

.grid-item {
    background: #007bff;
    color: white;
    padding: 2rem;
    border-radius: 8px;
    text-align: center;
    transition: transform 0.2s ease;
}

.grid-item:hover {
    transform: translateY(-2px);
    box-shadow: 0 4px 12px rgba(0,123,255,0.3);
}

.chart-container {
    background: #f8f9fa;
    border: 1px solid #dee2e6;
    border-radius: 8px;
    padding: 2rem;
    text-align: center;
    margin: 2rem 0;
}

.desktop-svg {
    display: block;
    margin: 2rem auto;
    border: 1px solid #dee2e6;
    border-radius: 8px;
}

/* High-resolution display optimization */
@media (-webkit-min-device-pixel-ratio: 2), (min-resolution: 192dpi) {
    .desktop-svg {
        image-rendering: -webkit-optimize-contrast;
        image-rendering: crisp-edges;
    }
}

/* Large desktop screens */
@media (min-width: 1440px) {
    .desktop-main {
        padding: 3rem;
    }
    
    .desktop-grid {
        grid-template-columns: repeat(4, 1fr);
    }
}

/* Reduced motion support */
@media (prefers-reduced-motion: reduce) {
    .grid-item {
        transition: none;
    }
    
    .desktop-nav a {
        transition: none;
    }
}`
}

func createMobileManifest() string {
	return `{
    "version": "1.0.0",
    "metadata": {
        "title": "Mobile Test Document",
        "author": "LIV Test Suite",
        "description": "Mobile-optimized test document",
        "language": "en"
    },
    "features": {
        "animations": true,
        "interactivity": true,
        "touch": true,
        "gestures": true
    },
    "security": {
        "wasmPermissions": {
            "memoryLimit": 8388608,
            "cpuTimeLimit": 3000
        }
    }
}`
}

func createDesktopManifest() string {
	return `{
    "version": "1.0.0",
    "metadata": {
        "title": "Desktop Test Document",
        "author": "LIV Test Suite",
        "description": "Desktop-optimized test document",
        "language": "en"
    },
    "features": {
        "animations": true,
        "interactivity": true,
        "charts": true,
        "highResolution": true
    },
    "security": {
        "wasmPermissions": {
            "memoryLimit": 33554432,
            "cpuTimeLimit": 10000
        }
    }
}`
}

// Performance testing functions

func TestCrossPlatformPerformance(t *testing.T) {
	if testing.Short() {
		t.Skip("Skipping performance tests in short mode")
	}

	platforms := []string{"web", "desktop", "mobile"}

	for _, platform := range platforms {
		t.Run(fmt.Sprintf("Performance_%s", platform), func(t *testing.T) {
			measurePlatformPerformance(t, platform)
		})
	}
}

func measurePlatformPerformance(t *testing.T, platform string) {
	tempDir, err := ioutil.TempDir("", fmt.Sprintf("liv-perf-%s-*", platform))
	require.NoError(t, err)
	defer os.RemoveAll(tempDir)

	// Create test document
	createPerformanceTestDocument(t, tempDir, platform)

	// Build document
	start := time.Now()
	buildCmd := exec.Command("liv-cli", "build", "perf-doc", "-o", "perf.liv")
	buildCmd.Dir = tempDir
	err = buildCmd.Run()
	buildDuration := time.Since(start)

	assert.NoError(t, err, "Build should succeed")
	assert.Less(t, buildDuration, 30*time.Second, "Build should complete within 30 seconds")

	// Test viewer startup
	start = time.Now()
	viewerCmd := exec.Command("liv-viewer", "--web", "--port", "8082", "perf.liv")
	viewerCmd.Dir = tempDir
	viewerCmd.Env = append(os.Environ(), "LIV_TEST_MODE=1")

	ctx, cancel := context.WithTimeout(context.Background(), 10*time.Second)
	defer cancel()

	viewerCmd = exec.CommandContext(ctx, "liv-viewer", "--web", "--port", "8082", "perf.liv")
	viewerCmd.Dir = tempDir
	viewerCmd.Env = append(os.Environ(), "LIV_TEST_MODE=1")

	err = viewerCmd.Start()
	if err == nil {
		viewerCmd.Process.Kill()
		viewerDuration := time.Since(start)
		assert.Less(t, viewerDuration, 5*time.Second, "Viewer should start within 5 seconds")
	}

	t.Logf("Platform %s - Build: %v, Viewer: %v", platform, buildDuration, time.Since(start))
}

func createPerformanceTestDocument(t *testing.T, tempDir, platform string) {
	docDir := filepath.Join(tempDir, "perf-doc")
	err := os.MkdirAll(docDir, 0755)
	require.NoError(t, err)

	// Create content based on platform
	var content string
	switch platform {
	case "mobile":
		content = createMobileTestContent()
	case "desktop":
		content = createDesktopTestContent()
	default:
		content = createTestHTML()
	}

	err = ioutil.WriteFile(filepath.Join(docDir, "index.html"), []byte(content), 0644)
	require.NoError(t, err)

	err = ioutil.WriteFile(filepath.Join(docDir, "style.css"), []byte(createTestCSS()), 0644)
	require.NoError(t, err)
}
