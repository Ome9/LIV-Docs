package main

import (
	"encoding/json"
	"fmt"
	"os"

	"github.com/spf13/cobra"
	"github.com/liv-format/liv/pkg/core"
	"github.com/liv-format/liv/pkg/manifest"
)

func main() {
	var (
		verbose bool
		format  string
	)

	rootCmd := &cobra.Command{
		Use:   "manifest-validator [file]",
		Short: "Validate LIV document manifests",
		Long: `Manifest Validator checks LIV document manifests for structural integrity,
security compliance, and content validity. It can validate individual manifest
files or manifests within complete LIV documents.`,
		Args: cobra.ExactArgs(1),
		RunE: func(cmd *cobra.Command, args []string) error {
			return validateManifest(args[0], verbose, format)
		},
	}

	rootCmd.Flags().BoolVarP(&verbose, "verbose", "v", false, "Verbose output with warnings")
	rootCmd.Flags().StringVarP(&format, "format", "f", "text", "Output format (text, json)")

	if err := rootCmd.Execute(); err != nil {
		fmt.Fprintf(os.Stderr, "Error: %v\n", err)
		os.Exit(1)
	}
}

func validateManifest(filePath string, verbose bool, format string) error {
	// Read the manifest file
	data, err := os.ReadFile(filePath)
	if err != nil {
		return fmt.Errorf("failed to read manifest file: %v", err)
	}

	// Create validator and parse
	validator := manifest.NewManifestValidator()
	manifestObj, result := validator.ValidateManifestJSON(data)

	// Output results based on format
	switch format {
	case "json":
		return outputJSON(result, manifestObj)
	case "text":
		return outputText(result, manifestObj, verbose)
	default:
		return fmt.Errorf("unsupported output format: %s", format)
	}
}

func outputText(result *core.ValidationResult, manifestObj *core.Manifest, verbose bool) error {
	fmt.Printf("Manifest Validation Report\n")
	fmt.Printf("==========================\n\n")

	if manifestObj != nil {
		fmt.Printf("Document: %s\n", manifestObj.Metadata.Title)
		fmt.Printf("Author: %s\n", manifestObj.Metadata.Author)
		fmt.Printf("Version: %s\n", manifestObj.Metadata.Version)
		fmt.Printf("Format Version: %s\n\n", manifestObj.Version)
	}

	// Validation status
	if result.IsValid {
		fmt.Printf("✓ Status: VALID\n")
	} else {
		fmt.Printf("✗ Status: INVALID\n")
	}

	// Errors
	if len(result.Errors) > 0 {
		fmt.Printf("\nErrors (%d):\n", len(result.Errors))
		for i, err := range result.Errors {
			fmt.Printf("  %d. %s\n", i+1, err)
		}
	}

	// Warnings (if verbose or if there are errors)
	if len(result.Warnings) > 0 && (verbose || !result.IsValid) {
		fmt.Printf("\nWarnings (%d):\n", len(result.Warnings))
		for i, warning := range result.Warnings {
			fmt.Printf("  %d. %s\n", i+1, warning)
		}
	}

	// Summary information if verbose
	if verbose && manifestObj != nil {
		fmt.Printf("\nSummary:\n")
		fmt.Printf("  Resources: %d\n", len(manifestObj.Resources))
		
		if manifestObj.WASMConfig != nil {
			fmt.Printf("  WASM Modules: %d\n", len(manifestObj.WASMConfig.Modules))
			fmt.Printf("  WASM Memory Limit: %d MB\n", manifestObj.WASMConfig.MemoryLimit/(1024*1024))
		}
		
		if manifestObj.Security != nil {
			fmt.Printf("  JavaScript Mode: %s\n", manifestObj.Security.JSPermissions.ExecutionMode)
			fmt.Printf("  Network Access: %v\n", manifestObj.Security.NetworkPolicy.AllowOutbound)
		}
		
		if manifestObj.Features != nil {
			enabledFeatures := []string{}
			if manifestObj.Features.Animations {
				enabledFeatures = append(enabledFeatures, "animations")
			}
			if manifestObj.Features.Interactivity {
				enabledFeatures = append(enabledFeatures, "interactivity")
			}
			if manifestObj.Features.Charts {
				enabledFeatures = append(enabledFeatures, "charts")
			}
			if manifestObj.Features.WebAssembly {
				enabledFeatures = append(enabledFeatures, "webassembly")
			}
			fmt.Printf("  Enabled Features: %v\n", enabledFeatures)
		}
	}

	fmt.Println()

	if !result.IsValid {
		return fmt.Errorf("manifest validation failed")
	}

	return nil
}

func outputJSON(result *core.ValidationResult, manifestObj *core.Manifest) error {
	output := map[string]interface{}{
		"valid":    result.IsValid,
		"errors":   result.Errors,
		"warnings": result.Warnings,
	}

	if manifestObj != nil {
		output["manifest"] = map[string]interface{}{
			"title":          manifestObj.Metadata.Title,
			"author":         manifestObj.Metadata.Author,
			"version":        manifestObj.Metadata.Version,
			"format_version": manifestObj.Version,
			"resource_count": len(manifestObj.Resources),
		}

		if manifestObj.WASMConfig != nil {
			output["wasm_modules"] = len(manifestObj.WASMConfig.Modules)
		}
	}

	data, err := json.MarshalIndent(output, "", "  ")
	if err != nil {
		return fmt.Errorf("failed to marshal JSON output: %v", err)
	}

	fmt.Println(string(data))
	return nil
}

// Add a command to generate example manifests
func init() {
	generateCmd := &cobra.Command{
		Use:   "generate [type]",
		Short: "Generate example manifest files",
		Long:  "Generate example manifest files for different document types (interactive, static)",
		Args:  cobra.ExactArgs(1),
		RunE: func(cmd *cobra.Command, args []string) error {
			return generateExampleManifest(args[0])
		},
	}

	// Add generate command as subcommand
	// This would be added to the root command in a real implementation
	_ = generateCmd
}

func generateExampleManifest(manifestType string) error {
	var builder *manifest.ManifestBuilder

	switch manifestType {
	case "interactive":
		builder = manifest.CreateInteractiveDocumentTemplate("Interactive Document", "Example Author")
		
		// Add some example resources
		builder.AddResource("content/index.html", &core.Resource{
			Hash: "e3b0c44298fc1c149afbf4c8996fb92427ae41e4649b934ca495991b7852b855",
			Size: 4096,
			Type: "text/html",
			Path: "content/index.html",
		})
		
		// Add WASM module
		wasmModule := &core.WASMModule{
			Name:       "interactive-engine",
			Version:    "1.0.0",
			EntryPoint: "init_interactive_engine",
			Exports:    []string{"init_interactive_engine", "process_interaction", "render_frame"},
			Imports:    []string{"env.memory", "env.console_log"},
		}
		builder.AddWASMModule(wasmModule)

	case "static":
		builder = manifest.CreateStaticDocumentTemplate("Static Document", "Example Author")
		
		// Add basic resources
		builder.AddResource("content/index.html", &core.Resource{
			Hash: "e3b0c44298fc1c149afbf4c8996fb92427ae41e4649b934ca495991b7852b855",
			Size: 2048,
			Type: "text/html",
			Path: "content/index.html",
		})

	default:
		return fmt.Errorf("unknown manifest type: %s (supported: interactive, static)", manifestType)
	}

	// Build and output the manifest
	data, err := builder.BuildJSON()
	if err != nil {
		return fmt.Errorf("failed to build manifest: %v", err)
	}

	fmt.Println(string(data))
	return nil
}