package viewer

import (
	"fmt"
	"strings"
	"time"

	"github.com/liv-format/liv/pkg/core"
)

// DocumentValidator implements the core.DocumentValidator interface
type DocumentValidator struct {
	manifestValidator core.DocumentValidator // For manifest-specific validation
	logger            core.Logger
	config            *ValidatorConfiguration
}

// ValidatorConfiguration holds configuration for document validation
type ValidatorConfiguration struct {
	StrictMode         bool     `json:"strict_mode"`
	ValidateContent    bool     `json:"validate_content"`
	ValidateAssets     bool     `json:"validate_assets"`
	ValidateSignatures bool     `json:"validate_signatures"`
	ValidateWASM       bool     `json:"validate_wasm"`
	MaxContentSize     int      `json:"max_content_size"`
	MaxAssetSize       int      `json:"max_asset_size"`
	RequiredResources  []string `json:"required_resources"`
}

// NewDocumentValidator creates a new document validator
func NewDocumentValidator(logger core.Logger) *DocumentValidator {
	config := &ValidatorConfiguration{
		StrictMode:         false,
		ValidateContent:    true,
		ValidateAssets:     true,
		ValidateSignatures: true,
		ValidateWASM:       true,
		MaxContentSize:     10 * 1024 * 1024, // 10MB
		MaxAssetSize:       50 * 1024 * 1024, // 50MB
		RequiredResources: []string{
			"content/index.html",
			// Don't require manifest.json as a resource - it's the manifest itself
		},
	}

	return &DocumentValidator{
		logger: logger,
		config: config,
	}
}

// ValidateDocument performs comprehensive document validation
func (dv *DocumentValidator) ValidateDocument(doc *core.LIVDocument) *core.ValidationResult {
	result := &core.ValidationResult{
		IsValid:  true,
		Errors:   []string{},
		Warnings: []string{},
	}

	if doc == nil {
		result.IsValid = false
		result.Errors = append(result.Errors, "document is nil")
		return result
	}

	// Validate manifest
	if manifestResult := dv.ValidateManifest(doc.Manifest); manifestResult != nil {
		result.Errors = append(result.Errors, manifestResult.Errors...)
		result.Warnings = append(result.Warnings, manifestResult.Warnings...)
		if !manifestResult.IsValid {
			result.IsValid = false
		}
	}

	// Validate content
	if dv.config.ValidateContent {
		if contentResult := dv.ValidateContent(doc.Content); contentResult != nil {
			result.Errors = append(result.Errors, contentResult.Errors...)
			result.Warnings = append(result.Warnings, contentResult.Warnings...)
			if !contentResult.IsValid {
				result.IsValid = false
			}
		}
	}

	// Validate assets
	if dv.config.ValidateAssets {
		if assetResult := dv.ValidateAssets(doc.Assets); assetResult != nil {
			result.Errors = append(result.Errors, assetResult.Errors...)
			result.Warnings = append(result.Warnings, assetResult.Warnings...)
			if !assetResult.IsValid {
				result.IsValid = false
			}
		}
	}

	// Validate signatures
	if dv.config.ValidateSignatures {
		if sigResult := dv.ValidateSignatures(doc); sigResult != nil {
			result.Errors = append(result.Errors, sigResult.Errors...)
			result.Warnings = append(result.Warnings, sigResult.Warnings...)
			if !sigResult.IsValid {
				result.IsValid = false
			}
		}
	}

	// Validate WASM modules
	if dv.config.ValidateWASM && doc.WASMModules != nil {
		wasmErrors, wasmWarnings := dv.validateWASMModules(doc.WASMModules)
		result.Errors = append(result.Errors, wasmErrors...)
		result.Warnings = append(result.Warnings, wasmWarnings...)
		if len(wasmErrors) > 0 && dv.config.StrictMode {
			result.IsValid = false
		}
	}

	// Validate resource consistency
	resourceErrors, resourceWarnings := dv.validateResourceConsistency(doc)
	result.Errors = append(result.Errors, resourceErrors...)
	result.Warnings = append(result.Warnings, resourceWarnings...)
	if len(resourceErrors) > 0 {
		result.IsValid = false
	}

	dv.logger.Debug("document validation completed",
		"valid", result.IsValid,
		"errors", len(result.Errors),
		"warnings", len(result.Warnings),
	)

	return result
}

// ValidateManifest validates manifest structure and content
func (dv *DocumentValidator) ValidateManifest(manifest *core.Manifest) *core.ValidationResult {
	result := &core.ValidationResult{
		IsValid:  true,
		Errors:   []string{},
		Warnings: []string{},
	}

	if manifest == nil {
		result.IsValid = false
		result.Errors = append(result.Errors, "manifest is nil")
		return result
	}

	// Validate version
	if manifest.Version == "" {
		result.IsValid = false
		result.Errors = append(result.Errors, "manifest version is required")
	} else if manifest.Version != "1.0" {
		result.Warnings = append(result.Warnings, fmt.Sprintf("manifest version '%s' may not be fully supported", manifest.Version))
	}

	// Validate metadata
	if manifest.Metadata == nil {
		result.IsValid = false
		result.Errors = append(result.Errors, "manifest metadata is required")
	} else {
		metadataErrors, metadataWarnings := dv.validateMetadata(manifest.Metadata)
		result.Errors = append(result.Errors, metadataErrors...)
		result.Warnings = append(result.Warnings, metadataWarnings...)
		if len(metadataErrors) > 0 {
			result.IsValid = false
		}
	}

	// Validate security policy
	if manifest.Security == nil {
		result.IsValid = false
		result.Errors = append(result.Errors, "security policy is required")
	} else {
		securityErrors, securityWarnings := dv.validateSecurityPolicy(manifest.Security)
		result.Errors = append(result.Errors, securityErrors...)
		result.Warnings = append(result.Warnings, securityWarnings...)
		if len(securityErrors) > 0 {
			result.IsValid = false
		}
	}

	// Validate resources
	if manifest.Resources == nil {
		result.IsValid = false
		result.Errors = append(result.Errors, "resource manifest is required")
	} else {
		resourceErrors, resourceWarnings := dv.validateResourceManifest(manifest.Resources)
		result.Errors = append(result.Errors, resourceErrors...)
		result.Warnings = append(result.Warnings, resourceWarnings...)
		if len(resourceErrors) > 0 {
			result.IsValid = false
		}
	}

	return result
}

// ValidateContent validates document content
func (dv *DocumentValidator) ValidateContent(content *core.DocumentContent) *core.ValidationResult {
	result := &core.ValidationResult{
		IsValid:  true,
		Errors:   []string{},
		Warnings: []string{},
	}

	if content == nil {
		result.IsValid = false
		result.Errors = append(result.Errors, "document content is nil")
		return result
	}

	// Validate HTML content
	if content.HTML == "" {
		result.IsValid = false
		result.Errors = append(result.Errors, "HTML content is required")
	} else {
		if len(content.HTML) > dv.config.MaxContentSize {
			result.IsValid = false
			result.Errors = append(result.Errors, fmt.Sprintf("HTML content size %d exceeds limit %d", len(content.HTML), dv.config.MaxContentSize))
		}

		// Basic HTML validation
		if !strings.Contains(content.HTML, "<html") && !strings.Contains(content.HTML, "<!DOCTYPE") {
			result.Warnings = append(result.Warnings, "HTML content may be missing DOCTYPE or html tag")
		}
	}

	// Validate CSS content
	if content.CSS != "" {
		if len(content.CSS) > dv.config.MaxContentSize {
			result.Warnings = append(result.Warnings, fmt.Sprintf("CSS content size %d is large", len(content.CSS)))
		}
	}

	// Validate static fallback
	if content.StaticFallback == "" {
		result.Warnings = append(result.Warnings, "static fallback content is missing")
	} else {
		if len(content.StaticFallback) > dv.config.MaxContentSize {
			result.Warnings = append(result.Warnings, fmt.Sprintf("static fallback size %d is large", len(content.StaticFallback)))
		}
	}

	// Validate interactive spec
	if content.InteractiveSpec != "" {
		if len(content.InteractiveSpec) > dv.config.MaxContentSize {
			result.Warnings = append(result.Warnings, fmt.Sprintf("interactive spec size %d is large", len(content.InteractiveSpec)))
		}
	}

	return result
}

// ValidateAssets validates asset bundle
func (dv *DocumentValidator) ValidateAssets(assets *core.AssetBundle) *core.ValidationResult {
	result := &core.ValidationResult{
		IsValid:  true,
		Errors:   []string{},
		Warnings: []string{},
	}

	if assets == nil {
		result.Warnings = append(result.Warnings, "no assets found in document")
		return result
	}

	totalSize := int64(0)

	// Validate images
	for name, data := range assets.Images {
		size := int64(len(data))
		totalSize += size

		if size > int64(dv.config.MaxAssetSize) {
			result.Errors = append(result.Errors, fmt.Sprintf("image asset '%s' size %d exceeds limit %d", name, size, dv.config.MaxAssetSize))
			result.IsValid = false
		}

		if size == 0 {
			result.Warnings = append(result.Warnings, fmt.Sprintf("image asset '%s' is empty", name))
		}
	}

	// Validate fonts
	for name, data := range assets.Fonts {
		size := int64(len(data))
		totalSize += size

		if size > int64(dv.config.MaxAssetSize) {
			result.Errors = append(result.Errors, fmt.Sprintf("font asset '%s' size %d exceeds limit %d", name, size, dv.config.MaxAssetSize))
			result.IsValid = false
		}

		if size == 0 {
			result.Warnings = append(result.Warnings, fmt.Sprintf("font asset '%s' is empty", name))
		}
	}

	// Validate data files
	for name, data := range assets.Data {
		size := int64(len(data))
		totalSize += size

		if size > int64(dv.config.MaxAssetSize) {
			result.Errors = append(result.Errors, fmt.Sprintf("data asset '%s' size %d exceeds limit %d", name, size, dv.config.MaxAssetSize))
			result.IsValid = false
		}

		if size == 0 {
			result.Warnings = append(result.Warnings, fmt.Sprintf("data asset '%s' is empty", name))
		}
	}

	// Check total asset size
	if totalSize > 100*1024*1024 { // 100MB
		result.Warnings = append(result.Warnings, fmt.Sprintf("total asset size %d is very large", totalSize))
	}

	return result
}

// ValidateSignatures validates all signatures
func (dv *DocumentValidator) ValidateSignatures(doc *core.LIVDocument) *core.ValidationResult {
	result := &core.ValidationResult{
		IsValid:  true,
		Errors:   []string{},
		Warnings: []string{},
	}

	if doc.Signatures == nil {
		result.Warnings = append(result.Warnings, "no signatures found in document")
		return result
	}

	// Validate content signature
	if doc.Signatures.ContentSignature == "" {
		result.Warnings = append(result.Warnings, "content signature is missing")
	}

	// Validate manifest signature
	if doc.Signatures.ManifestSignature == "" {
		result.Warnings = append(result.Warnings, "manifest signature is missing")
	}

	// Validate WASM signatures
	if len(doc.WASMModules) > 0 {
		if len(doc.Signatures.WASMSignatures) == 0 {
			result.Warnings = append(result.Warnings, "WASM modules present but no WASM signatures found")
		} else {
			for moduleName := range doc.WASMModules {
				if _, exists := doc.Signatures.WASMSignatures[moduleName]; !exists {
					result.Warnings = append(result.Warnings, fmt.Sprintf("WASM module '%s' has no signature", moduleName))
				}
			}
		}
	}

	return result
}

// Helper methods

func (dv *DocumentValidator) validateMetadata(metadata *core.DocumentMetadata) ([]string, []string) {
	var errors []string
	var warnings []string

	// Validate required fields
	if metadata.Title == "" {
		errors = append(errors, "document title is required")
	}

	if metadata.Author == "" {
		errors = append(errors, "document author is required")
	}

	if metadata.Version == "" {
		errors = append(errors, "document version is required")
	}

	if metadata.Language == "" {
		errors = append(errors, "document language is required")
	}

	// Validate dates
	if metadata.Created.IsZero() {
		errors = append(errors, "document creation date is required")
	}

	if metadata.Modified.IsZero() {
		errors = append(errors, "document modification date is required")
	}

	if !metadata.Created.IsZero() && !metadata.Modified.IsZero() {
		if metadata.Created.After(metadata.Modified) {
			errors = append(errors, "creation date cannot be after modification date")
		}
	}

	if !metadata.Modified.IsZero() && metadata.Modified.After(time.Now().Add(time.Hour)) {
		warnings = append(warnings, "modification date is in the future")
	}

	// Validate field lengths
	if len(metadata.Title) > 200 {
		errors = append(errors, "document title is too long (max 200 characters)")
	}

	if len(metadata.Author) > 100 {
		errors = append(errors, "document author is too long (max 100 characters)")
	}

	if len(metadata.Description) > 1000 {
		warnings = append(warnings, "document description is very long (>1000 characters)")
	}

	return errors, warnings
}

func (dv *DocumentValidator) validateSecurityPolicy(policy *core.SecurityPolicy) ([]string, []string) {
	var errors []string
	var warnings []string

	// Validate WASM permissions
	if policy.WASMPermissions == nil {
		errors = append(errors, "WASM permissions are required")
	} else {
		if policy.WASMPermissions.MemoryLimit == 0 {
			warnings = append(warnings, "WASM memory limit is not set")
		} else if policy.WASMPermissions.MemoryLimit > 256*1024*1024 {
			warnings = append(warnings, "WASM memory limit is very high (>256MB)")
		}

		if policy.WASMPermissions.CPUTimeLimit > 30000 {
			warnings = append(warnings, "WASM CPU time limit is very high (>30s)")
		}

		if policy.WASMPermissions.AllowNetworking {
			warnings = append(warnings, "WASM networking is enabled")
		}

		if policy.WASMPermissions.AllowFileSystem {
			warnings = append(warnings, "WASM file system access is enabled")
		}
	}

	// Validate JS permissions
	if policy.JSPermissions == nil {
		errors = append(errors, "JavaScript permissions are required")
	} else {
		if policy.JSPermissions.ExecutionMode == "trusted" {
			warnings = append(warnings, "JavaScript execution mode is set to trusted")
		}
	}

	return errors, warnings
}

func (dv *DocumentValidator) validateResourceManifest(resources map[string]*core.Resource) ([]string, []string) {
	var errors []string
	var warnings []string

	// Check for required resources
	for _, required := range dv.config.RequiredResources {
		if _, exists := resources[required]; !exists {
			errors = append(errors, fmt.Sprintf("required resource missing: %s", required))
		}
	}

	// Validate each resource
	for path, resource := range resources {
		if resource.Hash == "" {
			errors = append(errors, fmt.Sprintf("resource '%s' missing integrity hash", path))
		}

		if resource.Size < 0 {
			errors = append(errors, fmt.Sprintf("resource '%s' has negative size", path))
		}

		if resource.Type == "" {
			warnings = append(warnings, fmt.Sprintf("resource '%s' missing MIME type", path))
		}

		if resource.Path != path {
			errors = append(errors, fmt.Sprintf("resource path mismatch: key '%s' vs path '%s'", path, resource.Path))
		}

		// Check for large resources
		if resource.Size > 10*1024*1024 { // 10MB
			warnings = append(warnings, fmt.Sprintf("resource '%s' is very large (%d bytes)", path, resource.Size))
		}
	}

	return errors, warnings
}

func (dv *DocumentValidator) validateWASMModules(modules map[string][]byte) ([]string, []string) {
	var errors []string
	var warnings []string

	for name, data := range modules {
		if len(data) == 0 {
			errors = append(errors, fmt.Sprintf("WASM module '%s' is empty", name))
			continue
		}

		// Basic WASM validation
		if len(data) < 8 {
			errors = append(errors, fmt.Sprintf("WASM module '%s' is too small", name))
			continue
		}

		// Check WASM magic number
		if string(data[:4]) != "\x00asm" {
			errors = append(errors, fmt.Sprintf("WASM module '%s' has invalid magic number", name))
			continue
		}

		// Check WASM version
		version := uint32(data[4]) | uint32(data[5])<<8 | uint32(data[6])<<16 | uint32(data[7])<<24
		if version != 1 {
			warnings = append(warnings, fmt.Sprintf("WASM module '%s' has unsupported version: %d", name, version))
		}

		// Check module size
		if len(data) > 16*1024*1024 { // 16MB
			warnings = append(warnings, fmt.Sprintf("WASM module '%s' is very large (%d bytes)", name, len(data)))
		}
	}

	return errors, warnings
}

func (dv *DocumentValidator) validateResourceConsistency(doc *core.LIVDocument) ([]string, []string) {
	var errors []string
	var warnings []string

	if doc.Manifest == nil || doc.Manifest.Resources == nil {
		return errors, warnings
	}

	// Check that all manifest resources exist in the document
	for resourcePath := range doc.Manifest.Resources {
		exists := false

		// Check content resources
		if strings.HasPrefix(resourcePath, "content/") {
			exists = dv.contentResourceExists(doc.Content, resourcePath)
		}

		// Check asset resources
		if strings.HasPrefix(resourcePath, "assets/") {
			exists = dv.assetResourceExists(doc.Assets, resourcePath)
		}

		if !exists {
			errors = append(errors, fmt.Sprintf("resource '%s' listed in manifest but not found in document", resourcePath))
		}
	}

	return errors, warnings
}

func (dv *DocumentValidator) contentResourceExists(content *core.DocumentContent, resourcePath string) bool {
	if content == nil {
		return false
	}

	switch resourcePath {
	case "content/index.html":
		return content.HTML != ""
	case "content/styles/main.css":
		return content.CSS != ""
	case "content/static/fallback.html":
		return content.StaticFallback != ""
	case "content/scripts/interactive.js":
		return content.InteractiveSpec != ""
	default:
		return false
	}
}

func (dv *DocumentValidator) assetResourceExists(assets *core.AssetBundle, resourcePath string) bool {
	if assets == nil {
		return false
	}

	parts := strings.Split(resourcePath, "/")
	if len(parts) < 3 {
		return false
	}

	assetType := parts[1]
	assetName := strings.Join(parts[2:], "/")

	switch assetType {
	case "images":
		_, exists := assets.Images[assetName]
		return exists
	case "fonts":
		_, exists := assets.Fonts[assetName]
		return exists
	case "data":
		_, exists := assets.Data[assetName]
		return exists
	default:
		return false
	}
}

// UpdateConfiguration updates the validator configuration
func (dv *DocumentValidator) UpdateConfiguration(config *ValidatorConfiguration) error {
	if config == nil {
		return fmt.Errorf("configuration cannot be nil")
	}

	dv.config = config
	dv.logger.Info("document validator configuration updated")
	return nil
}

// GetConfiguration returns the current validator configuration
func (dv *DocumentValidator) GetConfiguration() *ValidatorConfiguration {
	return dv.config
}
