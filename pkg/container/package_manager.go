package container

import (
	"context"
	"fmt"
	"io"
	"os"
	"path/filepath"

	"github.com/liv-format/liv/pkg/core"
	"github.com/liv-format/liv/pkg/manifest"
)

// Container represents a .liv container for tests
type Container struct {
	Path    string
	manager *PackageManagerImpl
	files   map[string][]byte
}

// NewContainer creates a new container at the specified path
func NewContainer(path string) *Container {
	return &Container{
		Path:    path,
		manager: NewPackageManager(),
		files:   make(map[string][]byte),
	}
}

// AddFile adds a file to the container
func (c *Container) AddFile(path string, data []byte) error {
	c.files[path] = data
	return nil
}

// Save saves the container to disk
func (c *Container) Save() error {
	// Create ZIP archive at the container path
	return c.manager.zipContainer.CreateFromFiles(c.files, c.Path)
}

// OpenContainer opens an existing container from disk
func OpenContainer(path string) (*Container, error) {
	container := NewContainer(path)
	files, err := container.manager.zipContainer.ExtractToMemory(path)
	if err != nil {
		return nil, err
	}
	container.files = files
	return container, nil
}

// ReadFile reads a file from the container
func (c *Container) ReadFile(path string) ([]byte, error) {
	data, ok := c.files[path]
	if !ok {
		return nil, fmt.Errorf("file not found: %s", path)
	}
	return data, nil
}

// ListFiles returns a list of all files in the container
func (c *Container) ListFiles() ([]string, error) {
	files := make([]string, 0, len(c.files))
	for path := range c.files {
		files = append(files, path)
	}
	return files, nil
}

// PackageManagerImpl implements the core.PackageManager interface
type PackageManagerImpl struct {
	zipContainer *ZIPContainer
	validator    *manifest.ManifestValidator
	parser       *manifest.ManifestParser
}

// NewPackageManager creates a new package manager
func NewPackageManager() *PackageManagerImpl {
	return &PackageManagerImpl{
		zipContainer: NewZIPContainer(),
		validator:    manifest.NewManifestValidator(),
		parser:       manifest.NewManifestParser(),
	}
}

// CreatePackage creates a new .liv package from source files
func (pm *PackageManagerImpl) CreatePackage(ctx context.Context, sources map[string]io.Reader, manifest *core.Manifest) (*core.LIVDocument, error) {
	// Convert readers to byte arrays
	files := make(map[string][]byte)

	for path, reader := range sources {
		content, err := io.ReadAll(reader)
		if err != nil {
			return nil, fmt.Errorf("failed to read source file %s: %v", path, err)
		}
		files[path] = content
	}

	// Validate manifest
	result := pm.validator.ValidateManifest(manifest)
	if !result.IsValid {
		return nil, fmt.Errorf("manifest validation failed: %v", result.Errors)
	}

	// Create document structure
	document := &core.LIVDocument{
		Manifest:    manifest,
		Content:     &core.DocumentContent{},
		Assets:      &core.AssetBundle{},
		Signatures:  &core.SignatureBundle{},
		WASMModules: make(map[string][]byte),
	}

	// Extract content files
	if err := pm.extractContent(files, document); err != nil {
		return nil, fmt.Errorf("failed to extract content: %v", err)
	}

	// Extract assets
	if err := pm.extractAssets(files, document); err != nil {
		return nil, fmt.Errorf("failed to extract assets: %v", err)
	}

	// Extract WASM modules
	if err := pm.extractWASMModules(files, document); err != nil {
		return nil, fmt.Errorf("failed to extract WASM modules: %v", err)
	}

	// Extract signatures
	if err := pm.extractSignatures(files, document); err != nil {
		return nil, fmt.Errorf("failed to extract signatures: %v", err)
	}

	return document, nil
}

// ExtractPackage extracts a .liv package from a ZIP file
func (pm *PackageManagerImpl) ExtractPackage(ctx context.Context, reader io.Reader) (*core.LIVDocument, error) {
	// Read all data into memory
	data, err := io.ReadAll(reader)
	if err != nil {
		return nil, fmt.Errorf("failed to read package data: %v", err)
	}

	// Create a temporary file for ZIP operations
	tempFile, err := os.CreateTemp("", "liv-extract-*.zip")
	if err != nil {
		return nil, fmt.Errorf("failed to create temporary file: %v", err)
	}
	defer os.Remove(tempFile.Name())
	defer tempFile.Close()

	// Write data to temp file
	if _, err := tempFile.Write(data); err != nil {
		return nil, fmt.Errorf("failed to write temporary file: %v", err)
	}
	tempFile.Close()

	// Extract files to memory
	files, err := pm.zipContainer.ExtractToMemory(tempFile.Name())
	if err != nil {
		return nil, fmt.Errorf("failed to extract ZIP: %v", err)
	}

	// Validate structure
	structureResult := pm.zipContainer.ValidateStructureFromMemory(files)
	if !structureResult.IsValid {
		return nil, fmt.Errorf("invalid package structure: %v", structureResult.Errors)
	}

	// Parse manifest
	manifestData, exists := files["manifest.json"]
	if !exists {
		return nil, fmt.Errorf("manifest.json not found in package")
	}

	manifestObj, err := pm.parser.ParseFromBytes(manifestData)
	if err != nil {
		return nil, fmt.Errorf("failed to parse manifest: %v", err)
	}

	// Create document structure
	document := &core.LIVDocument{
		Manifest:    manifestObj,
		Content:     &core.DocumentContent{},
		Assets:      &core.AssetBundle{},
		Signatures:  &core.SignatureBundle{},
		WASMModules: make(map[string][]byte),
	}

	// Extract content files
	if err := pm.extractContent(files, document); err != nil {
		return nil, fmt.Errorf("failed to extract content: %v", err)
	}

	// Extract assets
	if err := pm.extractAssets(files, document); err != nil {
		return nil, fmt.Errorf("failed to extract assets: %v", err)
	}

	// Extract WASM modules
	if err := pm.extractWASMModules(files, document); err != nil {
		return nil, fmt.Errorf("failed to extract WASM modules: %v", err)
	}

	// Extract signatures
	if err := pm.extractSignatures(files, document); err != nil {
		return nil, fmt.Errorf("failed to extract signatures: %v", err)
	}

	return document, nil
}

// ValidateStructure validates the internal structure of a .liv package
func (pm *PackageManagerImpl) ValidateStructure(doc *core.LIVDocument) *core.ValidationResult {
	var errors []string
	var warnings []string

	// Validate manifest
	if doc.Manifest == nil {
		errors = append(errors, "document manifest is missing")
	} else {
		manifestResult := pm.validator.ValidateManifest(doc.Manifest)
		errors = append(errors, manifestResult.Errors...)
		warnings = append(warnings, manifestResult.Warnings...)
	}

	// Validate content structure
	if doc.Content == nil {
		errors = append(errors, "document content is missing")
	} else {
		if doc.Content.HTML == "" && doc.Content.StaticFallback == "" {
			errors = append(errors, "document must have either HTML content or static fallback")
		}
	}

	// Validate assets structure
	if doc.Assets == nil {
		warnings = append(warnings, "document has no assets")
	}

	// Validate WASM modules if configured
	if doc.Manifest != nil && doc.Manifest.WASMConfig != nil && len(doc.Manifest.WASMConfig.Modules) > 0 {
		for moduleName := range doc.Manifest.WASMConfig.Modules {
			if _, exists := doc.WASMModules[moduleName]; !exists {
				errors = append(errors, fmt.Sprintf("WASM module '%s' referenced in manifest but not found", moduleName))
			}
		}
	}

	// Check for orphaned WASM modules
	for moduleName := range doc.WASMModules {
		if doc.Manifest.WASMConfig == nil || doc.Manifest.WASMConfig.Modules[moduleName] == nil {
			warnings = append(warnings, fmt.Sprintf("WASM module '%s' found but not referenced in manifest", moduleName))
		}
	}

	return &core.ValidationResult{
		IsValid:  len(errors) == 0,
		Errors:   errors,
		Warnings: warnings,
	}
}

// CompressAssets compresses and deduplicates assets
func (pm *PackageManagerImpl) CompressAssets(assets *core.AssetBundle) (*core.AssetBundle, error) {
	if assets == nil {
		return &core.AssetBundle{
			Images: make(map[string][]byte),
			Fonts:  make(map[string][]byte),
			Data:   make(map[string][]byte),
		}, nil
	}

	// Combine all assets for deduplication
	allAssets := make(map[string][]byte)

	// Add images with prefix
	for name, data := range assets.Images {
		allAssets["images/"+name] = data
	}

	// Add fonts with prefix
	for name, data := range assets.Fonts {
		allAssets["fonts/"+name] = data
	}

	// Add data with prefix
	for name, data := range assets.Data {
		allAssets["data/"+name] = data
	}

	// Deduplicate
	deduplicated, duplicates := DeduplicateFiles(allAssets)

	// Log duplicates for information
	if len(duplicates) > 0 {
		fmt.Printf("Found %d duplicate files that were deduplicated\n", len(duplicates))
		for duplicate, original := range duplicates {
			fmt.Printf("  %s -> %s\n", duplicate, original)
		}
	}

	// Separate back into categories
	compressedAssets := &core.AssetBundle{
		Images: make(map[string][]byte),
		Fonts:  make(map[string][]byte),
		Data:   make(map[string][]byte),
	}

	for path, data := range deduplicated {
		if filepath.HasPrefix(path, "images/") {
			name := path[7:] // Remove "images/" prefix
			compressedAssets.Images[name] = data
		} else if filepath.HasPrefix(path, "fonts/") {
			name := path[6:] // Remove "fonts/" prefix
			compressedAssets.Fonts[name] = data
		} else if filepath.HasPrefix(path, "data/") {
			name := path[5:] // Remove "data/" prefix
			compressedAssets.Data[name] = data
		}
	}

	return compressedAssets, nil
}

// LoadWASMModule loads and validates a WASM module
func (pm *PackageManagerImpl) LoadWASMModule(name string, data []byte) (*core.WASMModule, error) {
	// Basic WASM validation - check for WASM magic number
	if len(data) < 4 {
		return nil, fmt.Errorf("WASM module too small")
	}

	// Check WASM magic number (0x00 0x61 0x73 0x6D)
	if data[0] != 0x00 || data[1] != 0x61 || data[2] != 0x73 || data[3] != 0x6D {
		return nil, fmt.Errorf("invalid WASM magic number")
	}

	// Check WASM version (0x01 0x00 0x00 0x00)
	if len(data) < 8 {
		return nil, fmt.Errorf("WASM module missing version")
	}

	if data[4] != 0x01 || data[5] != 0x00 || data[6] != 0x00 || data[7] != 0x00 {
		return nil, fmt.Errorf("unsupported WASM version")
	}

	// Create basic module info
	module := &core.WASMModule{
		Name:       name,
		Version:    "1.0.0", // Default version
		EntryPoint: "main",  // Default entry point
		Exports:    []string{},
		Imports:    []string{},
		Permissions: &core.WASMPermissions{
			MemoryLimit:     64 * 1024 * 1024, // 64MB default
			AllowedImports:  []string{},
			CPUTimeLimit:    5000, // 5 seconds default
			AllowNetworking: false,
			AllowFileSystem: false,
		},
		Metadata: make(map[string]string),
	}

	// TODO: Parse WASM module to extract actual exports/imports
	// This would require a full WASM parser, which is beyond the current scope
	// For now, we'll use defaults and rely on manifest configuration

	return module, nil
}

// Helper methods for extracting different types of content

func (pm *PackageManagerImpl) extractContent(files map[string][]byte, document *core.LIVDocument) error {
	content := &core.DocumentContent{}

	// Extract HTML content
	if htmlData, exists := files["content/index.html"]; exists {
		content.HTML = string(htmlData)
	}

	// Extract CSS content
	if cssData, exists := files["content/styles/main.css"]; exists {
		content.CSS = string(cssData)
	}

	// Extract JavaScript/interactive spec
	if jsData, exists := files["content/scripts/main.js"]; exists {
		content.InteractiveSpec = string(jsData)
	}

	// Extract static fallback
	if fallbackData, exists := files["content/static/fallback.html"]; exists {
		content.StaticFallback = string(fallbackData)
	}

	document.Content = content
	return nil
}

func (pm *PackageManagerImpl) extractAssets(files map[string][]byte, document *core.LIVDocument) error {
	assets := &core.AssetBundle{
		Images: make(map[string][]byte),
		Fonts:  make(map[string][]byte),
		Data:   make(map[string][]byte),
	}

	for path, data := range files {
		if filepath.HasPrefix(path, "assets/images/") {
			name := filepath.Base(path)
			assets.Images[name] = data
		} else if filepath.HasPrefix(path, "assets/fonts/") {
			name := filepath.Base(path)
			assets.Fonts[name] = data
		} else if filepath.HasPrefix(path, "assets/data/") {
			name := filepath.Base(path)
			assets.Data[name] = data
		}
	}

	document.Assets = assets
	return nil
}

func (pm *PackageManagerImpl) extractWASMModules(files map[string][]byte, document *core.LIVDocument) error {
	wasmModules := make(map[string][]byte)

	for path, data := range files {
		if filepath.Ext(path) == ".wasm" {
			name := filepath.Base(path)
			name = name[:len(name)-5] // Remove .wasm extension
			wasmModules[name] = data
		}
	}

	document.WASMModules = wasmModules
	return nil
}

func (pm *PackageManagerImpl) extractSignatures(files map[string][]byte, document *core.LIVDocument) error {
	signatures := &core.SignatureBundle{}

	if contentSig, exists := files["signatures/content.sig"]; exists {
		signatures.ContentSignature = string(contentSig)
	}

	if manifestSig, exists := files["signatures/manifest.sig"]; exists {
		signatures.ManifestSignature = string(manifestSig)
	}

	// Extract WASM signatures
	wasmSignatures := make(map[string]string)
	for path, data := range files {
		if filepath.HasPrefix(path, "signatures/") && filepath.Ext(path) == ".sig" {
			name := filepath.Base(path)
			name = name[:len(name)-4] // Remove .sig extension
			if name != "content" && name != "manifest" {
				wasmSignatures[name] = string(data)
			}
		}
	}
	signatures.WASMSignatures = wasmSignatures

	document.Signatures = signatures
	return nil
}

// SavePackage saves a LIV document to a file
func (pm *PackageManagerImpl) SavePackage(document *core.LIVDocument, outputPath string) error {
	// Validate document structure
	result := pm.ValidateStructure(document)
	if !result.IsValid {
		return fmt.Errorf("document validation failed: %v", result.Errors)
	}

	// Convert document to files map
	files, err := pm.documentToFiles(document)
	if err != nil {
		return fmt.Errorf("failed to convert document to files: %v", err)
	}

	// Create ZIP file
	return pm.zipContainer.CreateFromFiles(files, outputPath)
}

// SavePackageToWriter saves a LIV document to a writer
func (pm *PackageManagerImpl) SavePackageToWriter(document *core.LIVDocument, writer io.Writer) error {
	// Validate document structure
	result := pm.ValidateStructure(document)
	if !result.IsValid {
		return fmt.Errorf("document validation failed: %v", result.Errors)
	}

	// Convert document to files map
	files, err := pm.documentToFiles(document)
	if err != nil {
		return fmt.Errorf("failed to convert document to files: %v", err)
	}

	// Create ZIP to writer
	return pm.zipContainer.CreateFromFilesToWriter(files, writer)
}

func (pm *PackageManagerImpl) documentToFiles(document *core.LIVDocument) (map[string][]byte, error) {
	files := make(map[string][]byte)

	// Add manifest
	manifestData, err := pm.parser.SerializeToBytes(document.Manifest)
	if err != nil {
		return nil, fmt.Errorf("failed to serialize manifest: %v", err)
	}
	files["manifest.json"] = manifestData

	// Add content files
	if document.Content != nil {
		if document.Content.HTML != "" {
			files["content/index.html"] = []byte(document.Content.HTML)
		}
		if document.Content.CSS != "" {
			files["content/styles/main.css"] = []byte(document.Content.CSS)
		}
		if document.Content.InteractiveSpec != "" {
			files["content/scripts/main.js"] = []byte(document.Content.InteractiveSpec)
		}
		if document.Content.StaticFallback != "" {
			files["content/static/fallback.html"] = []byte(document.Content.StaticFallback)
		}
	}

	// Add assets
	if document.Assets != nil {
		for name, data := range document.Assets.Images {
			files["assets/images/"+name] = data
		}
		for name, data := range document.Assets.Fonts {
			files["assets/fonts/"+name] = data
		}
		for name, data := range document.Assets.Data {
			files["assets/data/"+name] = data
		}
	}

	// Add WASM modules
	for name, data := range document.WASMModules {
		files["wasm/"+name+".wasm"] = data
	}

	// Add signatures
	if document.Signatures != nil {
		if document.Signatures.ContentSignature != "" {
			files["signatures/content.sig"] = []byte(document.Signatures.ContentSignature)
		}
		if document.Signatures.ManifestSignature != "" {
			files["signatures/manifest.sig"] = []byte(document.Signatures.ManifestSignature)
		}
		for name, sig := range document.Signatures.WASMSignatures {
			files["signatures/"+name+".sig"] = []byte(sig)
		}
	}

	return files, nil
}
