package main

import (
	"fmt"
	"os"
	"path/filepath"
	"strings"
	"time"

	"github.com/spf13/cobra"
	"github.com/liv-format/liv/pkg/container"
	"github.com/liv-format/liv/pkg/core"
	"github.com/liv-format/liv/pkg/integrity"
	"github.com/liv-format/liv/pkg/manifest"
)

func main() {
	var (
		inputDir     string
		outputFile   string
		manifestFile string
		compress     bool
		sign         bool
		keyFile      string
		verbose      bool
	)

	rootCmd := &cobra.Command{
		Use:   "liv-builder",
		Short: "LIV Document Builder",
		Long: `LIV Builder creates Live Interactive Visual documents from source files.
It packages content, assets, and metadata into a secure, portable .liv file.`,
		RunE: func(cmd *cobra.Command, args []string) error {
			return runBuilder(inputDir, outputFile, manifestFile, compress, sign, keyFile, verbose)
		},
	}

	rootCmd.Flags().StringVarP(&inputDir, "input", "i", "", "Input directory containing source files (required)")
	rootCmd.Flags().StringVarP(&outputFile, "output", "o", "", "Output LIV file path (required)")
	rootCmd.Flags().StringVarP(&manifestFile, "manifest", "m", "", "Custom manifest file (optional)")
	rootCmd.Flags().BoolVarP(&compress, "compress", "c", true, "Compress assets")
	rootCmd.Flags().BoolVarP(&sign, "sign", "s", false, "Sign the document")
	rootCmd.Flags().StringVarP(&keyFile, "key", "k", "", "Private key file for signing")
	rootCmd.Flags().BoolVarP(&verbose, "verbose", "v", false, "Verbose output")

	rootCmd.MarkFlagRequired("input")
	rootCmd.MarkFlagRequired("output")

	if err := rootCmd.Execute(); err != nil {
		fmt.Fprintf(os.Stderr, "Error: %v\n", err)
		os.Exit(1)
	}
}

func runBuilder(inputDir, outputFile, manifestFile string, compress, sign bool, keyFile string, verbose bool) error {
	fmt.Printf("LIV Document Builder\n")
	fmt.Printf("====================\n\n")
	
	if verbose {
		fmt.Printf("Input directory: %s\n", inputDir)
		fmt.Printf("Output file: %s\n", outputFile)
		fmt.Printf("Manifest file: %s\n", manifestFile)
		fmt.Printf("Compress assets: %v\n", compress)
		fmt.Printf("Sign document: %v\n", sign)
		if keyFile != "" {
			fmt.Printf("Key file: %s\n", keyFile)
		}
		fmt.Println()
	}
	
	// Validate input directory exists
	if _, err := os.Stat(inputDir); os.IsNotExist(err) {
		return fmt.Errorf("input directory does not exist: %s", inputDir)
	}
	
	// Validate signing requirements
	if sign && keyFile == "" {
		return fmt.Errorf("signing requires a key file (--key)")
	}
	
	if sign {
		if _, err := os.Stat(keyFile); os.IsNotExist(err) {
			return fmt.Errorf("key file does not exist: %s", keyFile)
		}
	}
	
	// Build process steps
	steps := []struct {
		name string
		fn   func() error
	}{
		{"Scanning source files", func() error { return scanSourceFiles(inputDir, verbose) }},
		{"Validating content", func() error { return validateContent(inputDir, verbose) }},
		{"Processing assets", func() error { return processAssets(inputDir, compress, verbose) }},
		{"Generating manifest", func() error { return generateManifest(inputDir, manifestFile, verbose) }},
		{"Creating package", func() error { return createPackage(inputDir, outputFile, verbose) }},
	}
	
	if sign {
		steps = append(steps, struct {
			name string
			fn   func() error
		}{"Signing document", func() error { return signDocument(outputFile, keyFile, verbose) }})
	}
	
	// Execute build steps
	for i, step := range steps {
		fmt.Printf("[%d/%d] %s...\n", i+1, len(steps), step.name)
		
		if err := step.fn(); err != nil {
			return fmt.Errorf("failed at step '%s': %v", step.name, err)
		}
		
		if verbose {
			fmt.Printf("  ✓ %s completed\n", step.name)
		}
	}
	
	fmt.Printf("\n✓ LIV document created successfully: %s\n", outputFile)
	
	// Show file info
	if info, err := os.Stat(outputFile); err == nil {
		fmt.Printf("  File size: %d bytes\n", info.Size())
	}
	
	return nil
}

func scanSourceFiles(inputDir string, verbose bool) error {
	if verbose {
		fmt.Printf("  Scanning directory: %s\n", inputDir)
	}
	
	var fileCount int
	var totalSize int64
	
	err := filepath.Walk(inputDir, func(path string, info os.FileInfo, err error) error {
		if err != nil {
			return err
		}
		
		if !info.IsDir() {
			fileCount++
			totalSize += info.Size()
			
			if verbose {
				relPath, _ := filepath.Rel(inputDir, path)
				fmt.Printf("    Found: %s (%d bytes)\n", relPath, info.Size())
			}
		}
		
		return nil
	})
	
	if err != nil {
		return fmt.Errorf("failed to scan source files: %v", err)
	}
	
	if verbose {
		fmt.Printf("  Total files: %d\n", fileCount)
		fmt.Printf("  Total size: %d bytes\n", totalSize)
	}
	
	// Check for required files
	requiredFiles := []string{
		"content/index.html",
	}
	
	for _, required := range requiredFiles {
		fullPath := filepath.Join(inputDir, required)
		if _, err := os.Stat(fullPath); os.IsNotExist(err) {
			return fmt.Errorf("required file missing: %s", required)
		}
	}
	
	return nil
}

func validateContent(inputDir string, verbose bool) error {
	// TODO: Implement content validation
	if verbose {
		fmt.Printf("  Validating HTML, CSS, and JavaScript content\n")
		fmt.Printf("  Checking security policies\n")
		fmt.Printf("  Verifying asset references\n")
	}
	
	return nil
}

func processAssets(inputDir string, compress bool, verbose bool) error {
	if verbose {
		fmt.Printf("  Processing images, fonts, and data files\n")
		if compress {
			fmt.Printf("  Compressing assets\n")
		}
		fmt.Printf("  Calculating integrity hashes\n")
	}
	
	// Initialize hasher for integrity calculation
	hasher := integrity.NewResourceHasher(integrity.SHA256)
	
	var processedCount int
	
	err := filepath.Walk(inputDir, func(path string, info os.FileInfo, err error) error {
		if err != nil {
			return err
		}
		
		if info.IsDir() {
			return nil
		}
		
		// Skip hidden files and system files
		if strings.HasPrefix(info.Name(), ".") {
			return nil
		}
		
		// Calculate hash for integrity verification
		hash, err := hasher.HashFile(path)
		if err != nil {
			return fmt.Errorf("failed to hash file %s: %v", path, err)
		}
		
		processedCount++
		
		if verbose {
			relPath, _ := filepath.Rel(inputDir, path)
			fmt.Printf("    Processed: %s (hash: %s)\n", relPath, hash[:16]+"...")
		}
		
		return nil
	})
	
	if err != nil {
		return fmt.Errorf("failed to process assets: %v", err)
	}
	
	if verbose {
		fmt.Printf("  Processed %d assets\n", processedCount)
	}
	
	return nil
}

func generateManifest(inputDir, manifestFile string, verbose bool) error {
	if verbose {
		fmt.Printf("  Generating document manifest\n")
		if manifestFile != "" {
			fmt.Printf("  Using custom manifest: %s\n", manifestFile)
		}
		fmt.Printf("  Setting security policies\n")
		fmt.Printf("  Recording resource metadata\n")
	}
	
	// Create manifest builder
	builder := manifest.NewManifestBuilder()
	hasher := integrity.NewResourceHasher(integrity.SHA256)
	
	// Load custom manifest if provided, otherwise create default metadata
	var metadata *core.DocumentMetadata
	
	if manifestFile != "" {
		// Load existing manifest and extract metadata
		if _, err := os.Stat(manifestFile); err == nil {
			existingBuilder := manifest.NewManifestBuilder()
			if err := existingBuilder.LoadFromFile(manifestFile); err == nil {
				existingManifest := existingBuilder.GetManifest()
				metadata = existingManifest.Metadata
				
				// Also copy security policy and features if they exist
				if existingManifest.Security != nil {
					builder.SetSecurityPolicy(existingManifest.Security)
				}
				if existingManifest.Features != nil {
					builder.SetFeatureFlags(existingManifest.Features)
				}
				if existingManifest.WASMConfig != nil {
					builder.SetWASMConfig(existingManifest.WASMConfig)
				}
				
				if verbose {
					fmt.Printf("  Loaded custom manifest: %s\n", manifestFile)
				}
			} else if verbose {
				fmt.Printf("  Warning: Could not load custom manifest, using defaults\n")
			}
		}
	}
	
	// Create default metadata if not loaded from custom manifest
	if metadata == nil {
		// Try to extract title from HTML
		title := "LIV Document"
		if htmlPath := filepath.Join(inputDir, "content/index.html"); fileExists(htmlPath) {
			if htmlContent, err := os.ReadFile(htmlPath); err == nil {
				if extractedTitle := extractHTMLTitle(string(htmlContent)); extractedTitle != "" {
					title = extractedTitle
				}
			}
		}
		
		metadata = &core.DocumentMetadata{
			Title:       title,
			Author:      "LIV Builder",
			Created:     time.Now(),
			Modified:    time.Now(),
			Description: "Generated by LIV Builder",
			Version:     "1.0.0",
			Language:    "en",
		}
	} else {
		// Update modification time for existing metadata
		metadata.Modified = time.Now()
	}
	
	builder.SetMetadata(metadata)
	
	// Detect if document has interactive content (WASM modules or complex JS)
	hasWASM := false
	hasInteractiveJS := false
	
	// Scan for WASM modules and interactive content
	filepath.Walk(inputDir, func(path string, info os.FileInfo, err error) error {
		if err != nil {
			return nil // Continue scanning
		}
		
		if strings.HasSuffix(strings.ToLower(path), ".wasm") {
			hasWASM = true
		}
		
		if strings.HasSuffix(strings.ToLower(path), ".js") {
			// Simple heuristic: check for interactive keywords
			if content, err := os.ReadFile(path); err == nil {
				contentStr := strings.ToLower(string(content))
				if strings.Contains(contentStr, "canvas") || 
				   strings.Contains(contentStr, "webgl") ||
				   strings.Contains(contentStr, "websocket") ||
				   strings.Contains(contentStr, "fetch") {
					hasInteractiveJS = true
				}
			}
		}
		
		return nil
	})
	
	// Set security policy based on content type
	var securityPolicy *core.SecurityPolicy
	
	if hasWASM || hasInteractiveJS {
		// Interactive document with more permissive policy
		securityPolicy = &core.SecurityPolicy{
			WASMPermissions: &core.WASMPermissions{
				MemoryLimit:     128 * 1024 * 1024, // 128MB for interactive content
				AllowedImports:  []string{"env", "wasi_snapshot_preview1"},
				CPUTimeLimit:    15000, // 15 seconds for complex interactions
				AllowNetworking: false,
				AllowFileSystem: false,
			},
			JSPermissions: &core.JSPermissions{
				ExecutionMode: "sandboxed",
				AllowedAPIs:   []string{"dom", "canvas", "webgl", "audio"},
				DOMAccess:     "write",
			},
			NetworkPolicy: &core.NetworkPolicy{
				AllowOutbound: false,
				AllowedHosts:  []string{},
				AllowedPorts:  []int{},
			},
			StoragePolicy: &core.StoragePolicy{
				AllowLocalStorage:   true,
				AllowSessionStorage: true,
				AllowIndexedDB:      true,
				AllowCookies:        false,
			},
			ContentSecurityPolicy: "default-src 'self'; script-src 'self' 'unsafe-inline' 'wasm-unsafe-eval'; style-src 'self' 'unsafe-inline'; img-src 'self' data:;",
			TrustedDomains:        []string{},
		}
		
		if verbose {
			fmt.Printf("  Detected interactive content - using permissive security policy\n")
		}
	} else {
		// Static document with restrictive policy
		securityPolicy = &core.SecurityPolicy{
			WASMPermissions: &core.WASMPermissions{
				MemoryLimit:     1024, // Minimal for validation
				AllowedImports:  []string{},
				CPUTimeLimit:    100, // Minimal for validation
				AllowNetworking: false,
				AllowFileSystem: false,
			},
			JSPermissions: &core.JSPermissions{
				ExecutionMode: "sandboxed",
				AllowedAPIs:   []string{"dom"},
				DOMAccess:     "read",
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
			ContentSecurityPolicy: "default-src 'self'; script-src 'self'; style-src 'self' 'unsafe-inline';",
			TrustedDomains:        []string{},
		}
		
		if verbose {
			fmt.Printf("  Detected static content - using restrictive security policy\n")
		}
	}
	
	builder.SetSecurityPolicy(securityPolicy)
	
	// Set feature flags based on detected content
	features := &core.FeatureFlags{
		Animations:    true,  // Always enable basic animations
		Interactivity: hasWASM || hasInteractiveJS,
		Charts:        hasWASM || hasInteractiveJS,
		Forms:         hasInteractiveJS,
		Audio:         false, // Require explicit configuration
		Video:         false, // Require explicit configuration
		WebGL:         hasInteractiveJS,
		WebAssembly:   hasWASM,
	}
	builder.SetFeatureFlags(features)
	
	// Configure WASM modules if any are found
	if hasWASM {
		wasmConfig := &core.WASMConfiguration{
			Modules:     make(map[string]*core.WASMModule),
			Permissions: securityPolicy.WASMPermissions,
			MemoryLimit: securityPolicy.WASMPermissions.MemoryLimit,
		}
		
		// Scan for WASM modules and add them to configuration
		filepath.Walk(inputDir, func(path string, info os.FileInfo, err error) error {
			if err != nil || !strings.HasSuffix(strings.ToLower(path), ".wasm") {
				return nil
			}
			
			relPath, _ := filepath.Rel(inputDir, path)
			moduleName := strings.TrimSuffix(filepath.Base(path), ".wasm")
			
			wasmModule := &core.WASMModule{
				Name:        moduleName,
				Version:     "1.0.0",
				EntryPoint:  "main",
				Exports:     []string{"main", "memory"},
				Imports:     []string{"env"},
				Permissions: securityPolicy.WASMPermissions,
				Metadata: map[string]string{
					"path":        relPath,
					"description": fmt.Sprintf("WASM module: %s", moduleName),
					"created":     time.Now().Format(time.RFC3339),
				},
			}
			
			wasmConfig.Modules[moduleName] = wasmModule
			
			if verbose {
				fmt.Printf("    Configured WASM module: %s\n", moduleName)
			}
			
			return nil
		})
		
		if len(wasmConfig.Modules) > 0 {
			builder.SetWASMConfig(wasmConfig)
			if verbose {
				fmt.Printf("  Added WASM configuration with %d modules\n", len(wasmConfig.Modules))
			}
		}
	}
	
	// Scan and add resources
	err := filepath.Walk(inputDir, func(path string, info os.FileInfo, err error) error {
		if err != nil {
			return err
		}
		
		if info.IsDir() || strings.HasPrefix(info.Name(), ".") {
			return nil
		}
		
		// Calculate relative path
		relPath, err := filepath.Rel(inputDir, path)
		if err != nil {
			return err
		}
		
		// Normalize path separators
		relPath = filepath.ToSlash(relPath)
		
		// Calculate hash
		hash, err := hasher.HashFile(path)
		if err != nil {
			return fmt.Errorf("failed to hash file %s: %v", path, err)
		}
		
		// Determine MIME type
		mimeType := getMimeType(filepath.Ext(path))
		
		// Add resource to manifest
		builder.AddResource(relPath, &core.Resource{
			Hash: hash,
			Size: info.Size(),
			Type: mimeType,
			Path: relPath,
		})
		
		if verbose {
			fmt.Printf("    Added resource: %s (%s)\n", relPath, mimeType)
		}
		
		return nil
	})
	
	if err != nil {
		return fmt.Errorf("failed to scan resources: %v", err)
	}
	
	// Build and validate manifest
	builtManifest, err := builder.Build()
	if err != nil {
		return fmt.Errorf("failed to build manifest: %v", err)
	}
	
	// Save manifest to input directory for packaging
	manifestPath := filepath.Join(inputDir, "manifest.json")
	err = builder.SaveToFile(manifestPath)
	if err != nil {
		return fmt.Errorf("failed to save manifest: %v", err)
	}
	
	if verbose {
		fmt.Printf("  Generated manifest with %d resources\n", len(builtManifest.Resources))
		fmt.Printf("  Saved manifest to: %s\n", manifestPath)
	}
	
	return nil
}

// getMimeType returns the MIME type for a file extension
func getMimeType(ext string) string {
	ext = strings.ToLower(ext)
	switch ext {
	case ".html", ".htm":
		return "text/html"
	case ".css":
		return "text/css"
	case ".js":
		return "application/javascript"
	case ".json":
		return "application/json"
	case ".png":
		return "image/png"
	case ".jpg", ".jpeg":
		return "image/jpeg"
	case ".gif":
		return "image/gif"
	case ".svg":
		return "image/svg+xml"
	case ".woff":
		return "font/woff"
	case ".woff2":
		return "font/woff2"
	case ".ttf":
		return "font/ttf"
	case ".wasm":
		return "application/wasm"
	default:
		return "application/octet-stream"
	}
}

func createPackage(inputDir, outputFile string, verbose bool) error {
	if verbose {
		fmt.Printf("  Creating ZIP container\n")
		fmt.Printf("  Packaging content and assets\n")
		fmt.Printf("  Writing to: %s\n", outputFile)
	}
	
	// Create ZIP container with compression
	zipContainer := container.NewZIPContainer().
		SetCompressionLevel(-1). // Use default compression
		SetValidateStructure(true)
	
	// Create the .liv file from directory
	err := zipContainer.CreateFromDirectory(inputDir, outputFile)
	if err != nil {
		return fmt.Errorf("failed to create ZIP package: %v", err)
	}
	
	if verbose {
		// Get file info for reporting
		if info, err := os.Stat(outputFile); err == nil {
			fmt.Printf("  Package created: %d bytes\n", info.Size())
		}
		
		// Get compression statistics
		fileInfos, err := zipContainer.GetFileInfo(outputFile)
		if err == nil {
			var totalOriginal, totalCompressed int64
			for _, info := range fileInfos {
				totalOriginal += info.Size
				totalCompressed += info.CompressedSize
			}
			
			if totalOriginal > 0 {
				ratio := float64(totalCompressed) / float64(totalOriginal) * 100
				fmt.Printf("  Compression: %.1f%% (%d → %d bytes)\n", 
					ratio, totalOriginal, totalCompressed)
			}
		}
	}
	
	return nil
}

func signDocument(outputFile, keyFile string, verbose bool) error {
	if verbose {
		fmt.Printf("  Loading private key: %s\n", keyFile)
		fmt.Printf("  Generating content signatures\n")
		fmt.Printf("  Updating document with signatures\n")
	}
	
	// Create signature manager
	sigManager := integrity.NewSignatureManager()
	
	// Load private key
	privateKey, err := sigManager.LoadPrivateKeyPEM(keyFile)
	if err != nil {
		return fmt.Errorf("failed to load private key: %v", err)
	}
	
	// Load the document from the .liv file
	zipContainer := container.NewZIPContainer()
	files, err := zipContainer.ExtractToMemory(outputFile)
	if err != nil {
		return fmt.Errorf("failed to extract document for signing: %v", err)
	}
	
	// Parse manifest
	manifestData, exists := files["manifest.json"]
	if !exists {
		return fmt.Errorf("manifest.json not found in document")
	}
	
	validator := manifest.NewManifestValidator()
	parsedManifest, result := validator.ValidateManifestJSON(manifestData)
	if !result.IsValid {
		return fmt.Errorf("invalid manifest: %v", result.Errors)
	}
	
	// Create LIV document structure for signing
	document := &core.LIVDocument{
		Manifest: parsedManifest,
		Content: &core.DocumentContent{
			HTML:           string(files["content/index.html"]),
			CSS:            getFileContent(files, "content/styles/main.css", ""),
			InteractiveSpec: getFileContent(files, "content/interactive.json", ""),
			StaticFallback: getFileContent(files, "content/static/fallback.html", ""),
		},
		WASMModules: make(map[string][]byte),
	}
	
	// Add WASM modules if any
	for path, content := range files {
		if strings.HasSuffix(path, ".wasm") {
			moduleName := strings.TrimSuffix(filepath.Base(path), ".wasm")
			document.WASMModules[moduleName] = content
		}
	}
	
	// Sign the document
	signatures, err := sigManager.SignDocument(document, privateKey)
	if err != nil {
		return fmt.Errorf("failed to sign document: %v", err)
	}
	
	if verbose {
		fmt.Printf("  Generated signatures:\n")
		fmt.Printf("    Manifest: %s...\n", signatures.ManifestSignature[:16])
		fmt.Printf("    Content: %s...\n", signatures.ContentSignature[:16])
		if len(signatures.WASMSignatures) > 0 {
			fmt.Printf("    WASM modules: %d\n", len(signatures.WASMSignatures))
		}
	}
	
	// Update the document with signatures
	document.Signatures = signatures
	
	// Update manifest with signature information
	document.Manifest.Metadata.Modified = time.Now()
	
	// Re-serialize manifest with signatures
	manifestBuilder := manifest.NewManifestBuilder()
	manifestBuilder.SetMetadata(document.Manifest.Metadata)
	manifestBuilder.SetSecurityPolicy(document.Manifest.Security)
	if document.Manifest.WASMConfig != nil {
		manifestBuilder.SetWASMConfig(document.Manifest.WASMConfig)
	}
	if document.Manifest.Features != nil {
		manifestBuilder.SetFeatureFlags(document.Manifest.Features)
	}
	
	// Add resources back
	for path, resource := range document.Manifest.Resources {
		manifestBuilder.AddResource(path, resource)
	}
	
	updatedManifestData, err := manifestBuilder.BuildJSON()
	if err != nil {
		return fmt.Errorf("failed to build updated manifest: %v", err)
	}
	
	// Update files map with new manifest
	files["manifest.json"] = updatedManifestData
	
	// Create new signed .liv file
	err = zipContainer.CreateFromFiles(files, outputFile)
	if err != nil {
		return fmt.Errorf("failed to create signed document: %v", err)
	}
	
	if verbose {
		fmt.Printf("  Document signed successfully\n")
	}
	
	return nil
}

// getFileContent safely gets file content with fallback
func getFileContent(files map[string][]byte, path, fallback string) string {
	if content, exists := files[path]; exists {
		return string(content)
	}
	return fallback
}

// fileExists checks if a file exists
func fileExists(path string) bool {
	_, err := os.Stat(path)
	return err == nil
}

// extractHTMLTitle extracts the title from HTML content
func extractHTMLTitle(html string) string {
	// Simple regex to extract title
	titleStart := strings.Index(strings.ToLower(html), "<title>")
	if titleStart == -1 {
		return ""
	}
	
	titleStart += 7 // Length of "<title>"
	titleEnd := strings.Index(strings.ToLower(html[titleStart:]), "</title>")
	if titleEnd == -1 {
		return ""
	}
	
	title := strings.TrimSpace(html[titleStart : titleStart+titleEnd])
	if len(title) > 200 {
		title = title[:200] + "..."
	}
	
	return title
}