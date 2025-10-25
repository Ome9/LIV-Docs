package manifest

import (
	"crypto/sha256"
	"encoding/hex"
	"encoding/json"
	"fmt"
	"io"
	"os"
	"path/filepath"
	"strings"
	"time"

	"github.com/liv-format/liv/pkg/core"
)

// ManifestBuilder helps create and populate manifest structures
type ManifestBuilder struct {
	manifest  *core.Manifest
	validator *ManifestValidator
}

// NewManifestBuilder creates a new manifest builder
func NewManifestBuilder() *ManifestBuilder {
	return &ManifestBuilder{
		manifest: &core.Manifest{
			Version:   "1.0",
			Resources: make(map[string]*core.Resource),
		},
		validator: NewManifestValidator(),
	}
}

// SetMetadata sets the document metadata
func (mb *ManifestBuilder) SetMetadata(metadata *core.DocumentMetadata) *ManifestBuilder {
	mb.manifest.Metadata = metadata
	return mb
}

// SetSecurityPolicy sets the security policy
func (mb *ManifestBuilder) SetSecurityPolicy(policy *core.SecurityPolicy) *ManifestBuilder {
	mb.manifest.Security = policy
	return mb
}

// SetWASMConfig sets the WASM configuration
func (mb *ManifestBuilder) SetWASMConfig(config *core.WASMConfiguration) *ManifestBuilder {
	mb.manifest.WASMConfig = config
	return mb
}

// SetFeatureFlags sets the feature flags
func (mb *ManifestBuilder) SetFeatureFlags(features *core.FeatureFlags) *ManifestBuilder {
	mb.manifest.Features = features
	return mb
}

// AddResource adds a resource to the manifest
func (mb *ManifestBuilder) AddResource(path string, resource *core.Resource) *ManifestBuilder {
	if mb.manifest.Resources == nil {
		mb.manifest.Resources = make(map[string]*core.Resource)
	}
	mb.manifest.Resources[path] = resource
	return mb
}

// AddResourceFromFile adds a resource by reading from a file
func (mb *ManifestBuilder) AddResourceFromFile(path, filePath string) error {
	file, err := os.Open(filePath)
	if err != nil {
		return fmt.Errorf("failed to open file %s: %v", filePath, err)
	}
	defer file.Close()

	// Get file info
	info, err := file.Stat()
	if err != nil {
		return fmt.Errorf("failed to get file info for %s: %v", filePath, err)
	}

	// Calculate hash
	hasher := sha256.New()
	if _, err := io.Copy(hasher, file); err != nil {
		return fmt.Errorf("failed to calculate hash for %s: %v", filePath, err)
	}
	hash := hex.EncodeToString(hasher.Sum(nil))

	// Determine MIME type
	mimeType := mb.getMimeType(filePath)

	// Create resource
	resource := &core.Resource{
		Hash: hash,
		Size: info.Size(),
		Type: mimeType,
		Path: path,
	}

	mb.AddResource(path, resource)
	return nil
}

// ScanDirectory scans a directory and adds all files as resources
func (mb *ManifestBuilder) ScanDirectory(baseDir string) error {
	return filepath.Walk(baseDir, func(filePath string, info os.FileInfo, err error) error {
		if err != nil {
			return err
		}

		// Skip directories
		if info.IsDir() {
			return nil
		}

		// Calculate relative path
		relPath, err := filepath.Rel(baseDir, filePath)
		if err != nil {
			return fmt.Errorf("failed to get relative path: %v", err)
		}

		// Normalize path separators for cross-platform compatibility
		relPath = filepath.ToSlash(relPath)

		// Add resource
		return mb.AddResourceFromFile(relPath, filePath)
	})
}

// CreateDefaultMetadata creates default metadata with current timestamp
func (mb *ManifestBuilder) CreateDefaultMetadata(title, author string) *ManifestBuilder {
	now := time.Now()
	metadata := &core.DocumentMetadata{
		Title:       title,
		Author:      author,
		Created:     now,
		Modified:    now,
		Description: "",
		Version:     "1.0.0",
		Language:    "en",
	}
	return mb.SetMetadata(metadata)
}

// CreateDefaultSecurityPolicy creates a default security policy
func (mb *ManifestBuilder) CreateDefaultSecurityPolicy() *ManifestBuilder {
	policy := &core.SecurityPolicy{
		WASMPermissions: &core.WASMPermissions{
			MemoryLimit:     64 * 1024 * 1024, // 64MB
			AllowedImports:  []string{},
			CPUTimeLimit:    5000, // 5 seconds
			AllowNetworking: false,
			AllowFileSystem: false,
		},
		JSPermissions: &core.JSPermissions{
			ExecutionMode: "sandboxed",
			AllowedAPIs:   []string{},
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
		ContentSecurityPolicy: "default-src 'self'; script-src 'self' 'unsafe-inline'; style-src 'self' 'unsafe-inline';",
		TrustedDomains:        []string{},
	}
	return mb.SetSecurityPolicy(policy)
}

// CreateDefaultFeatureFlags creates default feature flags
func (mb *ManifestBuilder) CreateDefaultFeatureFlags() *ManifestBuilder {
	features := &core.FeatureFlags{
		Animations:    true,
		Interactivity: true,
		Charts:        true,
		Forms:         false,
		Audio:         false,
		Video:         false,
		WebGL:         false,
		WebAssembly:   true,
	}
	return mb.SetFeatureFlags(features)
}

// AddWASMModule adds a WASM module to the configuration
func (mb *ManifestBuilder) AddWASMModule(module *core.WASMModule) *ManifestBuilder {
	if mb.manifest.WASMConfig == nil {
		mb.manifest.WASMConfig = &core.WASMConfiguration{
			Modules: make(map[string]*core.WASMModule),
			Permissions: &core.WASMPermissions{
				MemoryLimit:     128 * 1024 * 1024, // 128MB
				AllowedImports:  []string{},
				CPUTimeLimit:    10000, // 10 seconds
				AllowNetworking: false,
				AllowFileSystem: false,
			},
			MemoryLimit: 128 * 1024 * 1024, // 128MB
		}
	}
	
	mb.manifest.WASMConfig.Modules[module.Name] = module
	return mb
}

// LoadFromFile loads an existing manifest from a JSON file
func (mb *ManifestBuilder) LoadFromFile(filePath string) error {
	data, err := os.ReadFile(filePath)
	if err != nil {
		return fmt.Errorf("failed to read manifest file: %v", err)
	}

	manifest, result := mb.validator.ValidateManifestJSON(data)
	if !result.IsValid {
		return fmt.Errorf("invalid manifest: %v", result.Errors)
	}

	mb.manifest = manifest
	return nil
}

// Build validates and returns the completed manifest
func (mb *ManifestBuilder) Build() (*core.Manifest, error) {
	// Validate the manifest
	result := mb.validator.ValidateManifest(mb.manifest)
	if !result.IsValid {
		return nil, fmt.Errorf("manifest validation failed: %v", result.Errors)
	}

	// Return a copy to prevent further modifications
	manifestCopy := *mb.manifest
	return &manifestCopy, nil
}

// BuildJSON validates and returns the manifest as JSON bytes
func (mb *ManifestBuilder) BuildJSON() ([]byte, error) {
	manifest, err := mb.Build()
	if err != nil {
		return nil, err
	}

	data, err := json.MarshalIndent(manifest, "", "  ")
	if err != nil {
		return nil, fmt.Errorf("failed to marshal manifest to JSON: %v", err)
	}

	return data, nil
}

// SaveToFile saves the manifest to a JSON file
func (mb *ManifestBuilder) SaveToFile(filePath string) error {
	// If the manifest directory contains files (content/, assets/, etc.), scan it and add resources
	dir := filepath.Dir(filePath)
	// Scan directory if it exists
	if _, err := os.Stat(dir); err == nil {
		_ = mb.ScanDirectory(dir) // best-effort: add any files found
	}

	// Validate the manifest before serialization
	if result := mb.validator.ValidateManifest(mb.manifest); !result.IsValid {
		return fmt.Errorf("manifest validation failed: %v", result.Errors)
	}

	// Marshal manifest to JSON
	manifest, err := mb.Build()
	if err != nil {
		return err
	}

	data, err := json.MarshalIndent(manifest, "", "  ")
	if err != nil {
		return fmt.Errorf("failed to marshal manifest to JSON: %v", err)
	}

	// Create directory if it doesn't exist
	if err := os.MkdirAll(dir, 0755); err != nil {
		return fmt.Errorf("failed to create directory: %v", err)
	}

	// Write manifest file
	if err := os.WriteFile(filePath, data, 0644); err != nil {
		return fmt.Errorf("failed to write manifest file: %v", err)
	}

	return nil
}

// GetManifest returns the current manifest (for inspection)
func (mb *ManifestBuilder) GetManifest() *core.Manifest {
	return mb.manifest
}

// Validate validates the current manifest
func (mb *ManifestBuilder) Validate() *core.ValidationResult {
	return mb.validator.ValidateManifest(mb.manifest)
}

// Helper methods

func (mb *ManifestBuilder) getMimeType(filePath string) string {
	ext := strings.ToLower(filepath.Ext(filePath))
	
	mimeTypes := map[string]string{
		".html": "text/html",
		".htm":  "text/html",
		".css":  "text/css",
		".js":   "application/javascript",
		".json": "application/json",
		".xml":  "application/xml",
		".txt":  "text/plain",
		".md":   "text/markdown",
		
		// Images
		".png":  "image/png",
		".jpg":  "image/jpeg",
		".jpeg": "image/jpeg",
		".gif":  "image/gif",
		".svg":  "image/svg+xml",
		".webp": "image/webp",
		".ico":  "image/x-icon",
		
		// Fonts
		".woff":  "font/woff",
		".woff2": "font/woff2",
		".ttf":   "font/ttf",
		".otf":   "font/otf",
		".eot":   "application/vnd.ms-fontobject",
		
		// Audio/Video
		".mp3":  "audio/mpeg",
		".wav":  "audio/wav",
		".ogg":  "audio/ogg",
		".mp4":  "video/mp4",
		".webm": "video/webm",
		
		// Data
		".csv":  "text/csv",
		".pdf":  "application/pdf",
		".zip":  "application/zip",
		
		// WASM
		".wasm": "application/wasm",
	}
	
	if mimeType, exists := mimeTypes[ext]; exists {
		return mimeType
	}
	
	return "application/octet-stream"
}

// Template methods for common manifest patterns

// CreateInteractiveDocumentTemplate creates a template for interactive documents
func CreateInteractiveDocumentTemplate(title, author string) *ManifestBuilder {
	mb := NewManifestBuilder()
	
	// Set metadata
	mb.CreateDefaultMetadata(title, author)
	
	// Set permissive security policy for interactive content
	policy := &core.SecurityPolicy{
		WASMPermissions: &core.WASMPermissions{
			MemoryLimit:     128 * 1024 * 1024, // 128MB
			AllowedImports:  []string{"env", "wasi_snapshot_preview1"},
			CPUTimeLimit:    15000, // 15 seconds
			AllowNetworking: false,
			AllowFileSystem: false,
		},
		JSPermissions: &core.JSPermissions{
			ExecutionMode: "sandboxed",
			AllowedAPIs:   []string{"canvas", "webgl", "audio"},
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
			AllowIndexedDB:      false,
			AllowCookies:        false,
		},
		ContentSecurityPolicy: "default-src 'self'; script-src 'self' 'unsafe-inline' 'wasm-unsafe-eval'; style-src 'self' 'unsafe-inline';",
		TrustedDomains:        []string{},
	}
	mb.SetSecurityPolicy(policy)
	
	// Enable interactive features
	features := &core.FeatureFlags{
		Animations:    true,
		Interactivity: true,
		Charts:        true,
		Forms:         true,
		Audio:         false,
		Video:         false,
		WebGL:         true,
		WebAssembly:   true,
	}
	mb.SetFeatureFlags(features)
	
	return mb
}

// CreateStaticDocumentTemplate creates a template for static documents
func CreateStaticDocumentTemplate(title, author string) *ManifestBuilder {
	mb := NewManifestBuilder()
	
	// Set metadata
	mb.CreateDefaultMetadata(title, author)
	
	// Set restrictive security policy for static content
	policy := &core.SecurityPolicy{
		WASMPermissions: &core.WASMPermissions{
			// Use minimum allowed values to satisfy validation while effectively disabling WASM
			MemoryLimit:     1024, // 1KB minimum
			AllowedImports:  []string{},
			CPUTimeLimit:    100, // 100ms minimum
			AllowNetworking: false,
			AllowFileSystem: false,
		},
		JSPermissions: &core.JSPermissions{
			ExecutionMode: "none",
			AllowedAPIs:   []string{},
			DOMAccess:     "none",
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
		ContentSecurityPolicy: "default-src 'none'; style-src 'self';",
		TrustedDomains:        []string{},
	}
	mb.SetSecurityPolicy(policy)
	
	// Disable interactive features
	features := &core.FeatureFlags{
		Animations:    false,
		Interactivity: false,
		Charts:        false,
		Forms:         false,
		Audio:         false,
		Video:         false,
		WebGL:         false,
		WebAssembly:   false,
	}
	mb.SetFeatureFlags(features)
	
	return mb
}