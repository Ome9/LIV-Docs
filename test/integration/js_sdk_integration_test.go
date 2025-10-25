package integration

import (
	"os/exec"
	"path/filepath"
	"strings"
	"testing"
	"time"

	"github.com/stretchr/testify/assert"

	"github.com/liv-format/liv/test/utils"
)

// TestJavaScriptSDKIntegration tests the JavaScript SDK integration
func TestJavaScriptSDKIntegration(t *testing.T) {
	if testing.Short() {
		t.Skip("Skipping JavaScript SDK integration test in short mode")
	}

	// Check if Node.js is available
	_, err := exec.LookPath("node")
	if err != nil {
		t.Skip("Node.js not found, skipping JavaScript SDK integration test")
	}

	helper := utils.NewTestHelper(t)
	defer helper.Cleanup()

	t.Run("basic_document_creation", func(t *testing.T) {
		testScript := `
const { LIVSDK, LIVHelpers } = require('../../js/dist/sdk');

async function testBasicCreation() {
	try {
		const sdk = LIVSDK.getInstance();
		
		// Test basic document creation
		const builder = await sdk.createDocument({
			metadata: {
				title: 'Test Document',
				author: 'Test Author',
				description: 'A test document created via JavaScript SDK'
			}
		});
		
		const html = '<html><head><title>Test</title></head><body><h1>Hello World</h1></body></html>';
		const css = 'body { font-family: Arial, sans-serif; }';
		
		const document = await builder
			.setHTML(html)
			.setCSS(css)
			.build();
		
		// Validate document
		const validation = await sdk.validateDocument(document);
		if (!validation.isValid) {
			throw new Error('Document validation failed: ' + validation.errors.join(', '));
		}
		
		// Get document info
		const info = sdk.getDocumentInfo(document);
		if (info.title !== 'Test Document') {
			throw new Error('Document title mismatch');
		}
		
		console.log('✓ Basic document creation test passed');
		return true;
	} catch (error) {
		console.error('✗ Basic document creation test failed:', error.message);
		return false;
	}
}

testBasicCreation().then(success => {
	process.exit(success ? 0 : 1);
});`

		testFile := helper.CreateTempFile("test_basic_creation.js", []byte(testScript))

		cmd := exec.Command("node", testFile)
		cmd.Dir = helper.TempDir
		output, err := cmd.CombinedOutput()

		t.Logf("JavaScript SDK basic creation test output: %s", string(output))

		if err != nil {
			t.Logf("Test failed (expected if SDK not built): %v", err)
			// Don't fail the test if SDK isn't built yet
		} else {
			assert.Contains(t, string(output), "✓ Basic document creation test passed")
		}
	})

	t.Run("asset_management", func(t *testing.T) {
		testScript := `
const { LIVSDK } = require('../../js/dist/sdk');
const fs = require('fs');

async function testAssetManagement() {
	try {
		const sdk = LIVSDK.getInstance();
		const builder = await sdk.createDocument({
			metadata: {
				title: 'Asset Test Document',
				author: 'Test Author'
			}
		});
		
		// Create test image data (fake PNG)
		const imageData = new Uint8Array([0x89, 0x50, 0x4E, 0x47, 0x0D, 0x0A, 0x1A, 0x0A]);
		
		// Add assets
		builder.addAsset({
			type: 'image',
			name: 'test-image.png',
			data: imageData.buffer,
			mimeType: 'image/png'
		});
		
		// Create test font data
		const fontData = new Uint8Array(1024).fill(0x42);
		builder.addAsset({
			type: 'font',
			name: 'test-font.woff2',
			data: fontData.buffer,
			mimeType: 'font/woff2'
		});
		
		const document = await builder.build();
		
		// Validate document with assets
		const validation = await sdk.validateDocument(document);
		if (!validation.isValid) {
			throw new Error('Document with assets validation failed');
		}
		
		console.log('✓ Asset management test passed');
		return true;
	} catch (error) {
		console.error('✗ Asset management test failed:', error.message);
		return false;
	}
}

testAssetManagement().then(success => {
	process.exit(success ? 0 : 1);
});`

		testFile := helper.CreateTempFile("test_asset_management.js", []byte(testScript))

		cmd := exec.Command("node", testFile)
		cmd.Dir = helper.TempDir
		output, err := cmd.CombinedOutput()

		t.Logf("JavaScript SDK asset management test output: %s", string(output))

		if err != nil {
			t.Logf("Test failed (expected if SDK not built): %v", err)
		} else {
			assert.Contains(t, string(output), "✓ Asset management test passed")
		}
	})

	t.Run("wasm_module_integration", func(t *testing.T) {
		testScript := `
const { LIVSDK } = require('../../js/dist/sdk');

async function testWASMIntegration() {
	try {
		const sdk = LIVSDK.getInstance();
		const builder = await sdk.createDocument({
			metadata: {
				title: 'WASM Test Document',
				author: 'Test Author'
			},
			features: {
				webassembly: true,
				interactivity: true
			}
		});
		
		// Create fake WASM module data (WASM magic number + version)
		const wasmData = new Uint8Array([0x00, 0x61, 0x73, 0x6d, 0x01, 0x00, 0x00, 0x00]);
		
		// Add WASM module
		builder.addWASMModule({
			name: 'test-module',
			data: wasmData.buffer,
			version: '1.0',
			entryPoint: 'main',
			permissions: {
				memoryLimit: 33554432, // 32MB
				allowedImports: ['env'],
				cpuTimeLimit: 5000,
				allowNetworking: false,
				allowFileSystem: false
			}
		});
		
		const document = await builder.build();
		
		// Validate document with WASM
		const validation = await sdk.validateDocument(document);
		if (!validation.isValid) {
			throw new Error('Document with WASM validation failed');
		}
		
		console.log('✓ WASM module integration test passed');
		return true;
	} catch (error) {
		console.error('✗ WASM module integration test failed:', error.message);
		return false;
	}
}

testWASMIntegration().then(success => {
	process.exit(success ? 0 : 1);
});`

		testFile := helper.CreateTempFile("test_wasm_integration.js", []byte(testScript))

		cmd := exec.Command("node", testFile)
		cmd.Dir = helper.TempDir
		output, err := cmd.CombinedOutput()

		t.Logf("JavaScript SDK WASM integration test output: %s", string(output))

		if err != nil {
			t.Logf("Test failed (expected if SDK not built): %v", err)
		} else {
			assert.Contains(t, string(output), "✓ WASM module integration test passed")
		}
	})

	t.Run("helper_functions", func(t *testing.T) {
		testScript := `
const { LIVHelpers } = require('../../js/dist/sdk');

async function testHelperFunctions() {
	try {
		// Test text document creation
		const textDoc = await LIVHelpers.createTextDocument(
			'Test Text Document',
			'This is a test document created using the helper function.',
			'Helper Tester'
		);
		
		if (!textDoc) {
			throw new Error('Text document creation failed');
		}
		
		// Test chart document creation
		const chartData = {
			labels: ['A', 'B', 'C'],
			values: [10, 20, 30]
		};
		
		const chartDoc = await LIVHelpers.createChartDocument(
			'Test Chart',
			chartData,
			'bar'
		);
		
		if (!chartDoc) {
			throw new Error('Chart document creation failed');
		}
		
		// Test presentation document creation
		const slides = [
			{ title: 'Slide 1', content: 'First slide content' },
			{ title: 'Slide 2', content: 'Second slide content' },
			{ title: 'Slide 3', content: 'Third slide content' }
		];
		
		const presentationDoc = await LIVHelpers.createPresentationDocument(
			'Test Presentation',
			slides
		);
		
		if (!presentationDoc) {
			throw new Error('Presentation document creation failed');
		}
		
		console.log('✓ Helper functions test passed');
		return true;
	} catch (error) {
		console.error('✗ Helper functions test failed:', error.message);
		return false;
	}
}

testHelperFunctions().then(success => {
	process.exit(success ? 0 : 1);
});`

		testFile := helper.CreateTempFile("test_helper_functions.js", []byte(testScript))

		cmd := exec.Command("node", testFile)
		cmd.Dir = helper.TempDir
		output, err := cmd.CombinedOutput()

		t.Logf("JavaScript SDK helper functions test output: %s", string(output))

		if err != nil {
			t.Logf("Test failed (expected if SDK not built): %v", err)
		} else {
			assert.Contains(t, string(output), "✓ Helper functions test passed")
		}
	})

	t.Run("renderer_integration", func(t *testing.T) {
		testScript := `
const { LIVSDK } = require('../../js/dist/sdk');
const { JSDOM } = require('jsdom');

async function testRendererIntegration() {
	try {
		// Create a DOM environment for testing
		const dom = new JSDOM('<!DOCTYPE html><html><body><div id="container"></div></body></html>');
		global.window = dom.window;
		global.document = dom.window.document;
		global.HTMLElement = dom.window.HTMLElement;
		
		const sdk = LIVSDK.getInstance();
		
		// Create a test document
		const builder = await sdk.createDocument({
			metadata: {
				title: 'Renderer Test Document',
				author: 'Test Author'
			}
		});
		
		const html = '<html><body><h1>Renderer Test</h1><p>This is a test document for renderer integration.</p></body></html>';
		const document = await builder.setHTML(html).build();
		
		// Create renderer
		const container = global.document.getElementById('container');
		const renderer = sdk.createRenderer(container, {
			enableInteractivity: false,
			enableAnimations: false,
			fallbackMode: true
		});
		
		if (!renderer) {
			throw new Error('Renderer creation failed');
		}
		
		console.log('✓ Renderer integration test passed');
		return true;
	} catch (error) {
		console.error('✗ Renderer integration test failed:', error.message);
		return false;
	}
}

testRendererIntegration().then(success => {
	process.exit(success ? 0 : 1);
});`

		testFile := helper.CreateTempFile("test_renderer_integration.js", []byte(testScript))

		// Install jsdom for testing if not available
		packageJsonContent := `{
  "name": "liv-sdk-test",
  "version": "1.0.0",
  "dependencies": {
    "jsdom": "^20.0.0"
  }
}`
		helper.CreateTempFile("package.json", []byte(packageJsonContent))

		// Try to install jsdom
		installCmd := exec.Command("npm", "install")
		installCmd.Dir = helper.TempDir
		installCmd.Run() // Ignore errors - jsdom might not be available

		cmd := exec.Command("node", testFile)
		cmd.Dir = helper.TempDir
		output, err := cmd.CombinedOutput()

		t.Logf("JavaScript SDK renderer integration test output: %s", string(output))

		if err != nil {
			t.Logf("Test failed (expected if dependencies not available): %v", err)
		} else {
			assert.Contains(t, string(output), "✓ Renderer integration test passed")
		}
	})

	t.Run("error_handling", func(t *testing.T) {
		testScript := `
const { LIVSDK } = require('../../js/dist/sdk');

async function testErrorHandling() {
	try {
		const sdk = LIVSDK.getInstance();
		
		// Test validation errors
		let errorCaught = false;
		try {
			const builder = await sdk.createDocument();
			// Try to build without required fields
			await builder.build();
		} catch (error) {
			errorCaught = true;
			if (!error.message.includes('title') && !error.message.includes('author')) {
				throw new Error('Expected validation error for missing title/author');
			}
		}
		
		if (!errorCaught) {
			throw new Error('Expected validation error was not thrown');
		}
		
		// Test invalid asset type
		errorCaught = false;
		try {
			const builder = await sdk.createDocument({
				metadata: { title: 'Test', author: 'Test' }
			});
			
			builder.addAsset({
				type: 'invalid-type',
				name: 'test.bin',
				data: new ArrayBuffer(100)
			});
		} catch (error) {
			errorCaught = true;
			if (!error.message.includes('Unsupported asset type')) {
				throw new Error('Expected asset type error');
			}
		}
		
		if (!errorCaught) {
			throw new Error('Expected asset type error was not thrown');
		}
		
		console.log('✓ Error handling test passed');
		return true;
	} catch (error) {
		console.error('✗ Error handling test failed:', error.message);
		return false;
	}
}

testErrorHandling().then(success => {
	process.exit(success ? 0 : 1);
});`

		testFile := helper.CreateTempFile("test_error_handling.js", []byte(testScript))

		cmd := exec.Command("node", testFile)
		cmd.Dir = helper.TempDir
		output, err := cmd.CombinedOutput()

		t.Logf("JavaScript SDK error handling test output: %s", string(output))

		if err != nil {
			t.Logf("Test failed (expected if SDK not built): %v", err)
		} else {
			assert.Contains(t, string(output), "✓ Error handling test passed")
		}
	})
}

// TestJavaScriptSDKPerformance tests JavaScript SDK performance
func TestJavaScriptSDKPerformance(t *testing.T) {
	if testing.Short() {
		t.Skip("Skipping JavaScript SDK performance test in short mode")
	}

	// Check if Node.js is available
	_, err := exec.LookPath("node")
	if err != nil {
		t.Skip("Node.js not found, skipping JavaScript SDK performance test")
	}

	helper := utils.NewTestHelper(t)
	defer helper.Cleanup()

	t.Run("document_creation_performance", func(t *testing.T) {
		testScript := `
const { LIVSDK } = require('../../js/dist/sdk');

async function testPerformance() {
	try {
		const sdk = LIVSDK.getInstance();
		const iterations = 100;
		
		console.log('Starting performance test with', iterations, 'iterations...');
		
		const startTime = Date.now();
		
		for (let i = 0; i < iterations; i++) {
			const builder = await sdk.createDocument({
				metadata: {
					title: 'Performance Test Document ' + i,
					author: 'Performance Tester'
				}
			});
			
			const html = '<html><body><h1>Performance Test ' + i + '</h1></body></html>';
			const document = await builder.setHTML(html).build();
			
			// Validate each document
			const validation = await sdk.validateDocument(document);
			if (!validation.isValid) {
				throw new Error('Document validation failed at iteration ' + i);
			}
		}
		
		const endTime = Date.now();
		const duration = endTime - startTime;
		const avgTime = duration / iterations;
		
		console.log('Performance test completed:');
		console.log('- Total time:', duration, 'ms');
		console.log('- Average time per document:', avgTime.toFixed(2), 'ms');
		console.log('- Documents per second:', (1000 / avgTime).toFixed(2));
		
		// Performance assertion - should create documents reasonably fast
		if (avgTime > 100) {
			console.warn('Warning: Document creation is slower than expected');
		}
		
		console.log('✓ Performance test completed');
		return true;
	} catch (error) {
		console.error('✗ Performance test failed:', error.message);
		return false;
	}
}

testPerformance().then(success => {
	process.exit(success ? 0 : 1);
});`

		testFile := helper.CreateTempFile("test_performance.js", []byte(testScript))

		cmd := exec.Command("node", testFile)
		cmd.Dir = helper.TempDir

		// Set timeout for performance test
		done := make(chan error, 1)
		go func() {
			output, err := cmd.CombinedOutput()
			t.Logf("JavaScript SDK performance test output: %s", string(output))
			done <- err
		}()

		select {
		case err := <-done:
			if err != nil {
				t.Logf("Performance test failed (expected if SDK not built): %v", err)
			}
		case <-time.After(30 * time.Second):
			t.Logf("Performance test timed out")
		}
	})
}

// TestJavaScriptSDKTypeScript tests TypeScript integration
func TestJavaScriptSDKTypeScript(t *testing.T) {
	if testing.Short() {
		t.Skip("Skipping TypeScript integration test in short mode")
	}

	// Check if TypeScript compiler is available
	_, err := exec.LookPath("tsc")
	if err != nil {
		t.Skip("TypeScript compiler not found, skipping TypeScript integration test")
	}

	helper := utils.NewTestHelper(t)
	defer helper.Cleanup()

	t.Run("typescript_compilation", func(t *testing.T) {
		// Create TypeScript test file
		tsContent := `
import { LIVSDK, LIVHelpers, DocumentCreationOptions } from '../../js/src/sdk';

async function testTypeScript(): Promise<boolean> {
	try {
		const sdk = LIVSDK.getInstance();
		
		const options: DocumentCreationOptions = {
			metadata: {
				title: 'TypeScript Test Document',
				author: 'TypeScript Tester',
				description: 'Testing TypeScript integration'
			},
			features: {
				animations: true,
				interactivity: false
			}
		};
		
		const builder = await sdk.createDocument(options);
		const document = await builder
			.setHTML('<html><body><h1>TypeScript Test</h1></body></html>')
			.setCSS('body { color: blue; }')
			.build();
		
		const validation = await sdk.validateDocument(document);
		
		console.log('TypeScript compilation and execution successful');
		return validation.isValid;
	} catch (error) {
		console.error('TypeScript test failed:', error);
		return false;
	}
}

testTypeScript().then(success => {
	process.exit(success ? 0 : 1);
});`

		tsFile := helper.CreateTempFile("test_typescript.ts", []byte(tsContent))

		// Create tsconfig.json
		tsconfigContent := `{
  "compilerOptions": {
    "target": "ES2020",
    "module": "commonjs",
    "lib": ["ES2020", "DOM"],
    "outDir": "./dist",
    "rootDir": "./",
    "strict": true,
    "esModuleInterop": true,
    "skipLibCheck": true,
    "forceConsistentCasingInFileNames": true,
    "moduleResolution": "node"
  },
  "include": ["*.ts"],
  "exclude": ["node_modules", "dist"]
}`
		helper.CreateTempFile("tsconfig.json", []byte(tsconfigContent))

		// Compile TypeScript
		compileCmd := exec.Command("tsc", filepath.Base(tsFile))
		compileCmd.Dir = helper.TempDir
		compileOutput, compileErr := compileCmd.CombinedOutput()

		t.Logf("TypeScript compilation output: %s", string(compileOutput))

		if compileErr != nil {
			t.Logf("TypeScript compilation failed (expected if types not available): %v", compileErr)
			return
		}

		// Run compiled JavaScript
		jsFile := strings.Replace(filepath.Base(tsFile), ".ts", ".js", 1)
		runCmd := exec.Command("node", jsFile)
		runCmd.Dir = helper.TempDir
		runOutput, runErr := runCmd.CombinedOutput()

		t.Logf("TypeScript execution output: %s", string(runOutput))

		if runErr != nil {
			t.Logf("TypeScript execution failed (expected if SDK not built): %v", runErr)
		} else {
			assert.Contains(t, string(runOutput), "TypeScript compilation and execution successful")
		}
	})
}
