package core

import (
	"context"
	"io"
)

// PackageManager handles ZIP packaging, manifest parsing, and asset management
type PackageManager interface {
	// CreatePackage creates a new .liv package from source files
	CreatePackage(ctx context.Context, sources map[string]io.Reader, manifest *Manifest) (*LIVDocument, error)
	
	// ExtractPackage extracts a .liv package from a ZIP file
	ExtractPackage(ctx context.Context, reader io.Reader) (*LIVDocument, error)
	
	// ValidateStructure validates the internal structure of a .liv package
	ValidateStructure(doc *LIVDocument) *ValidationResult
	
	// CompressAssets compresses and deduplicates assets
	CompressAssets(assets *AssetBundle) (*AssetBundle, error)
	
	// LoadWASMModule loads and validates a WASM module
	LoadWASMModule(name string, data []byte) (*WASMModule, error)
}

// SecurityManager orchestrates sandbox permissions and signature validation
type SecurityManager interface {
	// ValidateSignature verifies cryptographic signatures
	ValidateSignature(content []byte, signature string, publicKey []byte) bool
	
	// CreateSignature creates a cryptographic signature
	CreateSignature(content []byte, privateKey []byte) (string, error)
	
	// ValidateWASMModule validates a WASM module against security policies
	ValidateWASMModule(module []byte, permissions *WASMPermissions) error
	
	// CreateSandbox creates a secure execution environment
	CreateSandbox(policy *SecurityPolicy) (Sandbox, error)
	
	// EvaluatePermissions evaluates permission requests against policies
	EvaluatePermissions(requested *WASMPermissions, policy *SecurityPolicy) bool
	
	// GenerateSecurityReport creates a comprehensive security report
	GenerateSecurityReport(doc *LIVDocument) *SecurityReport
}

// WASMLoader loads and manages Rust WASM modules
type WASMLoader interface {
	// LoadModule loads a WASM module into memory
	LoadModule(ctx context.Context, name string, data []byte) (WASMInstance, error)
	
	// UnloadModule removes a WASM module from memory
	UnloadModule(name string) error
	
	// ListModules returns all loaded modules
	ListModules() []string
	
	// GetModuleInfo returns information about a loaded module
	GetModuleInfo(name string) (*WASMModule, error)
	
	// ValidateModule validates a WASM module before loading
	ValidateModule(data []byte) error
}

// WASMInstance represents a loaded WASM module instance
type WASMInstance interface {
	// Call invokes a WASM function
	Call(ctx context.Context, function string, args ...interface{}) (interface{}, error)
	
	// GetExports returns available exported functions
	GetExports() []string
	
	// GetMemoryUsage returns current memory usage
	GetMemoryUsage() uint64
	
	// SetMemoryLimit sets memory usage limit
	SetMemoryLimit(limit uint64) error
	
	// Terminate forcefully terminates the instance
	Terminate() error
}

// Sandbox represents a secure execution environment
type Sandbox interface {
	// Execute runs code within the sandbox
	Execute(ctx context.Context, code string, permissions *WASMPermissions) (interface{}, error)
	
	// LoadWASM loads a WASM module into the sandbox
	LoadWASM(ctx context.Context, module []byte, config *WASMModule) (WASMInstance, error)
	
	// GetPermissions returns current sandbox permissions
	GetPermissions() *SecurityPolicy
	
	// UpdatePermissions updates sandbox permissions
	UpdatePermissions(policy *SecurityPolicy) error
	
	// Destroy destroys the sandbox and cleans up resources
	Destroy() error
}

// DocumentValidator validates document structure and content
type DocumentValidator interface {
	// ValidateDocument performs comprehensive document validation
	ValidateDocument(doc *LIVDocument) *ValidationResult
	
	// ValidateManifest validates manifest structure and content
	ValidateManifest(manifest *Manifest) *ValidationResult
	
	// ValidateContent validates document content
	ValidateContent(content *DocumentContent) *ValidationResult
	
	// ValidateAssets validates asset bundle
	ValidateAssets(assets *AssetBundle) *ValidationResult
	
	// ValidateSignatures validates all signatures
	ValidateSignatures(doc *LIVDocument) *ValidationResult
}

// CryptoProvider provides cryptographic operations
type CryptoProvider interface {
	// GenerateKeyPair generates a new key pair
	GenerateKeyPair() (publicKey, privateKey []byte, err error)
	
	// Sign creates a digital signature
	Sign(data []byte, privateKey []byte) ([]byte, error)
	
	// Verify verifies a digital signature
	Verify(data []byte, signature []byte, publicKey []byte) bool
	
	// Hash computes SHA-256 hash
	Hash(data []byte) []byte
	
	// GenerateRandomBytes generates cryptographically secure random bytes
	GenerateRandomBytes(length int) ([]byte, error)
}

// Logger provides structured logging
type Logger interface {
	// Debug logs debug messages
	Debug(msg string, fields ...interface{})
	
	// Info logs info messages
	Info(msg string, fields ...interface{})
	
	// Warn logs warning messages
	Warn(msg string, fields ...interface{})
	
	// Error logs error messages
	Error(msg string, fields ...interface{})
	
	// Fatal logs fatal messages and exits
	Fatal(msg string, fields ...interface{})
}

// MetricsCollector collects performance and usage metrics
type MetricsCollector interface {
	// RecordDocumentLoad records document loading metrics
	RecordDocumentLoad(size int64, duration int64)
	
	// RecordWASMExecution records WASM execution metrics
	RecordWASMExecution(module string, duration int64, memoryUsed uint64)
	
	// RecordSecurityEvent records security-related events
	RecordSecurityEvent(eventType string, details map[string]interface{})
	
	// GetMetrics returns collected metrics
	GetMetrics() map[string]interface{}
}