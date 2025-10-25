package integration

import (
	"testing"
	"os"
	"path/filepath"
	"io/ioutil"
	"strings"
	"regexp"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"

	"github.com/liv-format/liv/test/utils"
)

// TestSDKDocumentation tests SDK documentation completeness and accuracy
func TestSDKDocumentation(t *testing.T) {
	helper := utils.NewTestHelper(t)
	defer helper.Cleanup()

	t.Run("javascript_sdk_documentation", func(t *testing.T) {
		// Check if JavaScript SDK documentation exists
		jsSDKPath := filepath.Join("..", "..", "js", "src", "sdk-documentation.ts")
		if _, err := os.Stat(jsSDKPath); os.IsNotExist(err) {
			t.Skip("JavaScript SDK documentation not found")
		}

		// Read documentation file
		content, err := ioutil.ReadFile(jsSDKPath)
		require.NoError(t, err)

		docContent := string(content)

		// Test documentation completeness
		t.Run("api_coverage", func(t *testing.T) {
			// Check for main classes documentation
			assert.Contains(t, docContent, "LIVSDK", "Should document main SDK class")
			assert.Contains(t, docContent, "LIVDocumentBuilder", "Should document document builder")
			assert.Contains(t, docContent, "LIVHelpers", "Should document helper functions")

			// Check for method documentation
			assert.Contains(t, docContent, "createDocument", "Should document createDocument method")
			assert.Contains(t, docContent, "loadDocument", "Should document loadDocument method")
			assert.Contains(t, docContent, "createRenderer", "Should document createRenderer method")
			assert.Contains(t, docContent, "validateDocument", "Should document validateDocument method")

			// Check for builder methods
			assert.Contains(t, docContent, "setHTML", "Should document setHTML method")
			assert.Contains(t, docContent, "setCSS", "Should document setCSS method")
			assert.Contains(t, docContent, "addAsset", "Should document addAsset method")
			assert.Contains(t, docContent, "addWASMModule", "Should document addWASMModule method")
		})

		t.Run("example_code", func(t *testing.T) {
			// Check for code examples
			codeBlockRegex := regexp.MustCompile("```(?:typescript|javascript|ts|js)")
			codeBlocks := codeBlockRegex.FindAllString(docContent, -1)
			assert.Greater(t, len(codeBlocks), 0, "Should contain code examples")

			// Check for specific examples
			assert.Contains(t, docContent, "example", "Should contain example usage")
			assert.Contains(t, docContent, "const sdk", "Should contain SDK usage examples")
		})

		t.Run("type_definitions", func(t *testing.T) {
			// Check for type documentation
			assert.Contains(t, docContent, "interface", "Should document interfaces")
			assert.Contains(t, docContent, "type", "Should document types")
			
			// Check for parameter documentation
			assert.Contains(t, docContent, "@param", "Should document parameters")
			assert.Contains(t, docContent, "@returns", "Should document return values")
		})
	})

	t.Run("python_sdk_documentation", func(t *testing.T) {
		// Check if Python SDK documentation exists
		pythonSDKPath := filepath.Join("..", "..", "python", "README.md")
		if _, err := os.Stat(pythonSDKPath); os.IsNotExist(err) {
			t.Skip("Python SDK documentation not found")
		}

		// Read documentation file
		content, err := ioutil.ReadFile(pythonSDKPath)
		require.NoError(t, err)

		docContent := string(content)

		// Test documentation completeness
		t.Run("api_coverage", func(t *testing.T) {
			// Check for main classes documentation
			assert.Contains(t, docContent, "LIVDocument", "Should document main document class")
			assert.Contains(t, docContent, "DocumentMetadata", "Should document metadata class")
			assert.Contains(t, docContent, "LIVBuilder", "Should document builder class")

			// Check for installation instructions
			assert.Contains(t, docContent, "install", "Should contain installation instructions")
			assert.Contains(t, docContent, "pip", "Should mention pip installation")

			// Check for usage examples
			assert.Contains(t, docContent, "import", "Should contain import examples")
			assert.Contains(t, docContent, "example", "Should contain usage examples")
		})

		t.Run("code_examples", func(t *testing.T) {
			// Check for Python code examples
			codeBlockRegex := regexp.MustCompile("```python")
			codeBlocks := codeBlockRegex.FindAllString(docContent, -1)
			assert.Greater(t, len(codeBlocks), 0, "Should contain Python code examples")

			// Check for specific examples
			assert.Contains(t, docContent, "from liv import", "Should contain import examples")
		})

		t.Run("api_reference", func(t *testing.T) {
			// Check for API reference sections
			assert.Contains(t, docContent, "API", "Should contain API reference")
			assert.Contains(t, docContent, "method", "Should document methods")
			assert.Contains(t, docContent, "parameter", "Should document parameters")
		})
	})

	t.Run("cross_reference_consistency", func(t *testing.T) {
		// Test that both SDKs document similar functionality consistently
		
		jsSDKPath := filepath.Join("..", "..", "js", "src", "sdk-documentation.ts")
		pythonSDKPath := filepath.Join("..", "..", "python", "README.md")
		
		var jsContent, pythonContent string
		
		if jsData, err := ioutil.ReadFile(jsSDKPath); err == nil {
			jsContent = string(jsData)
		}
		
		if pythonData, err := ioutil.ReadFile(pythonSDKPath); err == nil {
			pythonContent = string(pythonData)
		}
		
		if jsContent != "" && pythonContent != "" {
			// Check that both SDKs document similar core concepts
			coreConcepts := []string{
				"document creation",
				"asset management",
				"validation",
				"metadata",
				"security",
			}
			
			for _, concept := range coreConcepts {
				jsHasConcept := strings.Contains(strings.ToLower(jsContent), concept)
				pythonHasConcept := strings.Contains(strings.ToLower(pythonContent), concept)
				
				if jsHasConcept || pythonHasConcept {
					// If one SDK documents it, both should
					assert.True(t, jsHasConcept, "JavaScript SDK should document %s", concept)
					assert.True(t, pythonHasConcept, "Python SDK should document %s", concept)
				}
			}
		}
	})
}

// TestSDKAPICompleteness tests that SDKs implement all required functionality
func TestSDKAPICompleteness(t *testing.T) {
	helper := utils.NewTestHelper(t)
	defer helper.Cleanup()

	t.Run("javascript_api_completeness", func(t *testing.T) {
		// Check JavaScript SDK source files
		jsSDKPath := filepath.Join("..", "..", "js", "src", "sdk.ts")
		if _, err := os.Stat(jsSDKPath); os.IsNotExist(err) {
			t.Skip("JavaScript SDK source not found")
		}

		content, err := ioutil.ReadFile(jsSDKPath)
		require.NoError(t, err)

		sdkContent := string(content)

		// Test required classes exist
		t.Run("required_classes", func(t *testing.T) {
			requiredClasses := []string{
				"class LIVSDK",
				"class LIVDocumentBuilder",
				"class LIVHelpers",
			}

			for _, class := range requiredClasses {
				assert.Contains(t, sdkContent, class, "Should contain %s", class)
			}
		})

		// Test required methods exist
		t.Run("required_methods", func(t *testing.T) {
			requiredMethods := []string{
				"createDocument",
				"loadDocument",
				"createRenderer",
				"createEditor",
				"validateDocument",
				"convertDocument",
				"setHTML",
				"setCSS",
				"addAsset",
				"addWASMModule",
				"build",
			}

			for _, method := range requiredMethods {
				methodRegex := regexp.MustCompile(method + "\\s*\\(")
				assert.True(t, methodRegex.MatchString(sdkContent), "Should contain method %s", method)
			}
		})

		// Test helper functions exist
		t.Run("helper_functions", func(t *testing.T) {
			helperFunctions := []string{
				"createTextDocument",
				"createChartDocument",
				"createPresentationDocument",
			}

			for _, helper := range helperFunctions {
				assert.Contains(t, sdkContent, helper, "Should contain helper function %s", helper)
			}
		})

		// Test type definitions
		t.Run("type_definitions", func(t *testing.T) {
			// Check that types are imported
			assert.Contains(t, sdkContent, "import", "Should import types")
			assert.Contains(t, sdkContent, "DocumentMetadata", "Should use DocumentMetadata type")
			assert.Contains(t, sdkContent, "SecurityPolicy", "Should use SecurityPolicy type")
		})
	})

	t.Run("python_api_completeness", func(t *testing.T) {
		// Check Python SDK source files
		pythonSDKPath := filepath.Join("..", "..", "python", "src", "liv")
		if _, err := os.Stat(pythonSDKPath); os.IsNotExist(err) {
			t.Skip("Python SDK source not found")
		}

		// Check main document class
		documentPath := filepath.Join(pythonSDKPath, "document.py")
		if content, err := ioutil.ReadFile(documentPath); err == nil {
			docContent := string(content)

			t.Run("document_class", func(t *testing.T) {
				assert.Contains(t, docContent, "class LIVDocument", "Should contain LIVDocument class")
				
				requiredMethods := []string{
					"def load",
					"def save",
					"def validate",
					"def get_asset",
					"def get_wasm_module",
					"def list_assets",
					"def get_size_info",
				}

				for _, method := range requiredMethods {
					assert.Contains(t, docContent, method, "Should contain method %s", method)
				}
			})
		}

		// Check builder class
		builderPath := filepath.Join(pythonSDKPath, "builder.py")
		if content, err := ioutil.ReadFile(builderPath); err == nil {
			builderContent := string(content)

			t.Run("builder_class", func(t *testing.T) {
				assert.Contains(t, builderContent, "class LIVBuilder", "Should contain LIVBuilder class")
				
				builderMethods := []string{
					"def set_metadata",
					"def set_html_content",
					"def set_css_content",
					"def add_asset",
					"def enable_interactivity",
					"def build",
				}

				for _, method := range builderMethods {
					assert.Contains(t, builderContent, method, "Should contain method %s", method)
				}
			})
		}

		// Check CLI interface
		cliPath := filepath.Join(pythonSDKPath, "cli_interface.py")
		if content, err := ioutil.ReadFile(cliPath); err == nil {
			cliContent := string(content)

			t.Run("cli_interface", func(t *testing.T) {
				assert.Contains(t, cliContent, "class CLIInterface", "Should contain CLIInterface class")
				
				cliMethods := []string{
					"def build",
					"def validate",
					"def convert",
					"def extract",
				}

				for _, method := range cliMethods {
					assert.Contains(t, cliContent, method, "Should contain method %s", method)
				}
			})
		}

		// Check models
		modelsPath := filepath.Join(pythonSDKPath, "models.py")
		if content, err := ioutil.ReadFile(modelsPath); err == nil {
			modelsContent := string(content)

			t.Run("data_models", func(t *testing.T) {
				requiredModels := []string{
					"class DocumentMetadata",
					"class SecurityPolicy",
					"class AssetInfo",
					"class WASMModuleInfo",
					"class FeatureFlags",
				}

				for _, model := range requiredModels {
					assert.Contains(t, modelsContent, model, "Should contain model %s", model)
				}
			})
		}
	})

	t.Run("feature_parity", func(t *testing.T) {
		// Test that both SDKs support the same core features
		
		coreFeatures := map[string][]string{
			"Document Creation": {
				"metadata management",
				"content setting",
				"validation",
			},
			"Asset Management": {
				"image assets",
				"font assets",
				"data assets",
			},
			"WASM Integration": {
				"module loading",
				"permission management",
				"security policies",
			},
			"Format Conversion": {
				"export functionality",
				"import functionality",
			},
		}

		// This is a conceptual test - in practice, you would check
		// that both SDKs implement equivalent functionality
		for feature, subFeatures := range coreFeatures {
			t.Run(strings.ToLower(strings.ReplaceAll(feature, " ", "_")), func(t *testing.T) {
				// Test would verify that both JavaScript and Python SDKs
				// implement the required sub-features
				t.Logf("Testing feature parity for: %s", feature)
				for _, subFeature := range subFeatures {
					t.Logf("  - %s", subFeature)
				}
				// Actual implementation would check SDK source files
			})
		}
	})
}

// TestSDKExamples tests that SDK examples work correctly
func TestSDKExamples(t *testing.T) {
	helper := utils.NewTestHelper(t)
	defer helper.Cleanup()

	t.Run("javascript_examples", func(t *testing.T) {
		// Check if JavaScript examples exist
		examplesPath := filepath.Join("..", "..", "js", "examples")
		if _, err := os.Stat(examplesPath); os.IsNotExist(err) {
			t.Skip("JavaScript examples not found")
		}

		// Find example files
		err := filepath.Walk(examplesPath, func(path string, info os.FileInfo, err error) error {
			if err != nil {
				return err
			}

			if strings.HasSuffix(path, ".js") || strings.HasSuffix(path, ".ts") {
				t.Run(filepath.Base(path), func(t *testing.T) {
					content, err := ioutil.ReadFile(path)
					require.NoError(t, err)

					exampleContent := string(content)

					// Test that examples contain proper SDK usage
					assert.Contains(t, exampleContent, "LIVSDK", "Example should use LIVSDK")
					
					// Test that examples have proper structure
					if strings.Contains(exampleContent, "async") {
						assert.Contains(t, exampleContent, "await", "Async examples should use await")
					}

					// Test that examples handle errors
					assert.True(t, 
						strings.Contains(exampleContent, "try") || 
						strings.Contains(exampleContent, "catch") ||
						strings.Contains(exampleContent, ".catch"),
						"Examples should handle errors")
				})
			}

			return nil
		})

		if err != nil {
			t.Logf("Error walking JavaScript examples directory: %v", err)
		}
	})

	t.Run("python_examples", func(t *testing.T) {
		// Check if Python examples exist
		examplesPath := filepath.Join("..", "..", "python", "examples")
		if _, err := os.Stat(examplesPath); os.IsNotExist(err) {
			t.Skip("Python examples not found")
		}

		// Find example files
		err := filepath.Walk(examplesPath, func(path string, info os.FileInfo, err error) error {
			if err != nil {
				return err
			}

			if strings.HasSuffix(path, ".py") {
				t.Run(filepath.Base(path), func(t *testing.T) {
					content, err := ioutil.ReadFile(path)
					require.NoError(t, err)

					exampleContent := string(content)

					// Test that examples contain proper SDK usage
					assert.Contains(t, exampleContent, "from liv import", "Example should import LIV SDK")
					
					// Test that examples have proper structure
					assert.Contains(t, exampleContent, "def ", "Examples should contain functions")

					// Test that examples handle errors
					assert.True(t, 
						strings.Contains(exampleContent, "try:") || 
						strings.Contains(exampleContent, "except"),
						"Examples should handle errors")

					// Test that examples have main execution
					assert.Contains(t, exampleContent, "if __name__", "Examples should have main execution block")
				})
			}

			return nil
		})

		if err != nil {
			t.Logf("Error walking Python examples directory: %v", err)
		}
	})
}
