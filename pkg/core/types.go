package core

import (
	"encoding/json"
	"fmt"
	"time"
)

// LIVDocument represents a complete .liv document with all components
type LIVDocument struct {
	Manifest    *Manifest         `json:"manifest"`
	Content     *DocumentContent  `json:"content"`
	Assets      *AssetBundle      `json:"assets"`
	Signatures  *SignatureBundle  `json:"signatures"`
	WASMModules map[string][]byte `json:"wasm_modules"`
}

// DocumentContent holds the main content of the document
type DocumentContent struct {
	HTML            string `json:"html"`
	CSS             string `json:"css"`
	InteractiveSpec string `json:"interactive_spec"` // WASM module configuration
	StaticFallback  string `json:"static_fallback"`
}

// AssetBundle contains all document assets
type AssetBundle struct {
	Images map[string][]byte `json:"images"`
	Fonts  map[string][]byte `json:"fonts"`
	Data   map[string][]byte `json:"data"`
}

// SignatureBundle contains cryptographic signatures
type SignatureBundle struct {
	ContentSignature  string            `json:"content_signature"`
	ManifestSignature string            `json:"manifest_signature"`
	WASMSignatures    map[string]string `json:"wasm_signatures"`
}

// Manifest contains document metadata and security configuration
type Manifest struct {
	Version    string               `json:"version" validate:"required"`
	Metadata   *DocumentMetadata    `json:"metadata" validate:"required"`
	Security   *SecurityPolicy      `json:"security" validate:"required"`
	Resources  map[string]*Resource `json:"resources" validate:"required"`
	WASMConfig *WASMConfiguration   `json:"wasm_config"`
	Features   *FeatureFlags        `json:"features"`
}

// DocumentMetadata contains basic document information
type DocumentMetadata struct {
	Title       string    `json:"title" validate:"required,max=200"`
	Author      string    `json:"author" validate:"required,max=100"`
	Created     time.Time `json:"created" validate:"required"`
	Modified    time.Time `json:"modified" validate:"required"`
	Description string    `json:"description" validate:"max=1000"`
	Version     string    `json:"version" validate:"required,semver"`
	Language    string    `json:"language" validate:"required,len=2"`
}

// SecurityPolicy defines security constraints and permissions
type SecurityPolicy struct {
	WASMPermissions       *WASMPermissions `json:"wasm_permissions" validate:"required"`
	JSPermissions         *JSPermissions   `json:"js_permissions" validate:"required"`
	NetworkPolicy         *NetworkPolicy   `json:"network_policy" validate:"required"`
	StoragePolicy         *StoragePolicy   `json:"storage_policy" validate:"required"`
	ContentSecurityPolicy string           `json:"content_security_policy" validate:"csp"`
	TrustedDomains        []string         `json:"trusted_domains" validate:"dive,domain"`
}

// WASMPermissions defines WASM module execution constraints
type WASMPermissions struct {
	MemoryLimit     uint64   `json:"memory_limit" validate:"min=1024,max=134217728"` // 1KB to 128MB
	AllowedImports  []string `json:"allowed_imports"`
	CPUTimeLimit    uint64   `json:"cpu_time_limit" validate:"min=100,max=30000"` // 100ms to 30s
	AllowNetworking bool     `json:"allow_networking"`
	AllowFileSystem bool     `json:"allow_file_system"`
}

// JSPermissions defines JavaScript execution permissions
type JSPermissions struct {
	ExecutionMode string   `json:"execution_mode" validate:"oneof=none sandboxed trusted"`
	AllowedAPIs   []string `json:"allowed_apis"`
	DOMAccess     string   `json:"dom_access" validate:"oneof=none read write"`
}

// NetworkPolicy defines network access permissions
type NetworkPolicy struct {
	AllowOutbound bool     `json:"allow_outbound"`
	AllowedHosts  []string `json:"allowed_hosts"`
	AllowedPorts  []int    `json:"allowed_ports"`
}

// StoragePolicy defines storage access permissions
type StoragePolicy struct {
	AllowLocalStorage   bool `json:"allow_local_storage"`
	AllowSessionStorage bool `json:"allow_session_storage"`
	AllowIndexedDB      bool `json:"allow_indexed_db"`
	AllowCookies        bool `json:"allow_cookies"`
}

// Resource represents a file resource within the document
type Resource struct {
	Hash string `json:"hash" validate:"required,sha256"`
	Size int64  `json:"size" validate:"min=0"`
	Type string `json:"type" validate:"required,mimetype"`
	Path string `json:"path" validate:"required"`
}

// WASMConfiguration defines WASM module configuration
type WASMConfiguration struct {
	Modules     map[string]*WASMModule `json:"modules"`
	Permissions *WASMPermissions       `json:"permissions" validate:"required"`
	MemoryLimit uint64                 `json:"memory_limit" validate:"min=1024,max=134217728"`
}

// WASMModule represents a single WASM module
type WASMModule struct {
	Name        string            `json:"name" validate:"required,wasmmodule"`
	Version     string            `json:"version" validate:"required,semver"`
	EntryPoint  string            `json:"entry_point" validate:"required"`
	Exports     []string          `json:"exports"`
	Imports     []string          `json:"imports"`
	Permissions *WASMPermissions  `json:"permissions"`
	Metadata    map[string]string `json:"metadata"`
}

// FeatureFlags defines enabled document features
type FeatureFlags struct {
	Animations    bool `json:"animations"`
	Interactivity bool `json:"interactivity"`
	Charts        bool `json:"charts"`
	Forms         bool `json:"forms"`
	Audio         bool `json:"audio"`
	Video         bool `json:"video"`
	WebGL         bool `json:"webgl"`
	WebAssembly   bool `json:"webassembly"`
}

// ValidationResult represents the result of document validation
type ValidationResult struct {
	IsValid  bool     `json:"is_valid"`
	Errors   []string `json:"errors"`
	Warnings []string `json:"warnings"`
}

// SecurityReport represents security validation results
type SecurityReport struct {
	IsValid           bool     `json:"is_valid"`
	SignatureVerified bool     `json:"signature_verified"`
	IntegrityChecked  bool     `json:"integrity_checked"`
	PermissionsValid  bool     `json:"permissions_valid"`
	Warnings          []string `json:"warnings"`
	Errors            []string `json:"errors"`
}

// Document is an alias for LIVDocument for backward compatibility
type Document = LIVDocument

// NewDocument creates a new document with simplified initialization for tests
func NewDocument(metadata DocumentMetadata, content DocumentContent) *Document {
	return &Document{
		Manifest: &Manifest{
			Version:  "1.0",
			Metadata: &metadata,
			Security: &SecurityPolicy{
				WASMPermissions: &WASMPermissions{},
				JSPermissions:   &JSPermissions{},
				NetworkPolicy:   &NetworkPolicy{},
				StoragePolicy:   &StoragePolicy{},
			},
			Resources: make(map[string]*Resource),
		},
		Content:     &content,
		Assets:      &AssetBundle{},
		Signatures:  &SignatureBundle{},
		WASMModules: make(map[string][]byte),
	}
}

// GetMetadata returns the document metadata
func (d *LIVDocument) GetMetadata() *DocumentMetadata {
	if d.Manifest == nil {
		return nil
	}
	return d.Manifest.Metadata
}

// GetContent returns the document content
func (d *LIVDocument) GetContent() *DocumentContent {
	return d.Content
}

// MarshalJSON marshals the document to JSON
func (d *LIVDocument) MarshalJSON() ([]byte, error) {
	type DocumentAlias LIVDocument
	return json.Marshal((*DocumentAlias)(d))
}

// UnmarshalJSON unmarshals JSON into the document
func (d *LIVDocument) UnmarshalJSON(data []byte) error {
	type DocumentAlias LIVDocument
	return json.Unmarshal(data, (*DocumentAlias)(d))
}

// Validate validates the document structure
func (d *LIVDocument) Validate() error {
	if d.Manifest == nil {
		return fmt.Errorf("document manifest is required")
	}
	if d.Content == nil {
		return fmt.Errorf("document content is required")
	}
	if d.Manifest.Metadata == nil {
		return fmt.Errorf("document metadata is required")
	}
	if d.Manifest.Metadata.Title == "" {
		return fmt.Errorf("document title is required")
	}
	if d.Manifest.Metadata.Author == "" {
		return fmt.Errorf("document author is required")
	}
	return nil
}

// MarshalJSON marshals the manifest to JSON
func (m *Manifest) MarshalJSON() ([]byte, error) {
	type ManifestAlias Manifest
	return json.Marshal((*ManifestAlias)(m))
}

// UnmarshalJSON unmarshals JSON into the manifest
func (m *Manifest) UnmarshalJSON(data []byte) error {
	type ManifestAlias Manifest
	return json.Unmarshal(data, (*ManifestAlias)(m))
}

// Validate validates the manifest structure
func (m *Manifest) Validate() error {
	if m.Version == "" {
		return fmt.Errorf("manifest version is required")
	}
	if m.Metadata == nil {
		return fmt.Errorf("manifest metadata is required")
	}
	if m.Metadata.Title == "" {
		return fmt.Errorf("manifest title is required")
	}
	if m.Metadata.Author == "" {
		return fmt.Errorf("manifest author is required")
	}
	if m.Security == nil {
		return fmt.Errorf("manifest security policy is required")
	}
	return nil
}
