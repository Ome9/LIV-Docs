# API Reference

Complete API documentation for the LIV document format libraries and tools.

## 📚 Overview

The LIV ecosystem provides APIs for:

- **Go Library**: Core document processing and validation
- **JavaScript/TypeScript Library**: Web-based rendering and interaction
- **Rust Library**: WebAssembly interactive engine
- **CLI Tools**: Command-line interface for document operations
- **REST API**: Web service endpoints for document processing

## 🔧 Go Library API

### Package: `github.com/liv-format/liv/pkg/core`

#### Document

```go
type Document struct {
    Manifest    *Manifest
    Content     *Content
    Assets      map[string][]byte
    Signatures  *Signatures
    WasmModules map[string][]byte
}

// NewDocument creates a new LIV document
func NewDocument() *Document

// LoadFromFile loads a document from a .liv file
func LoadFromFile(path string) (*Document, error)

// LoadFromReader loads a document from an io.Reader
func LoadFromReader(r io.Reader) (*Document, error)

// Save saves the document to a .liv file
func (d *Document) Save(path string) error

// WriteToWriter writes the document to an io.Writer
func (d *Document) WriteToWriter(w io.Writer) error

// Validate validates the document structure and content
func (d *Document) Validate() (*ValidationResult, error)

// Sign signs the document with the provided key
func (d *Document) Sign(privateKey *rsa.PrivateKey) error

// Verify verifies the document signatures
func (d *Document) Verify() (*VerificationResult, error)

// GetSize returns the total size of the document
func (d *Document) GetSize() int64

// GetAsset retrieves an asset by path
func (d *Document) GetAsset(path string) ([]byte, error)

// AddAsset adds an asset to the document
func (d *Document) AddAsset(path string, data []byte) error

// RemoveAsset removes an asset from the document
func (d *Document) RemoveAsset(path string) error

// ListAssets returns a list of all asset paths
func (d *Document) ListAssets() []string
```

## 🌐 JavaScript/TypeScript Library API

### Package: `@liv-format/renderer`

#### LIVRenderer

```typescript
interface SecureRenderingOptions {
  container: HTMLElement;
  permissions: SecurityPolicy;
  enableFallback?: boolean;
  strictSecurity?: boolean;
  maxRenderTime?: number;
  enableAnimations?: boolean;
  targetFPS?: number;
  enableSVG?: boolean;
  enableResponsiveDesign?: boolean;
  errorHandler?: ErrorHandler;
}

class LIVRenderer {
  constructor(options: SecureRenderingOptions);
  
  // Document rendering
  async renderDocument(document: LIVDocument): Promise<void>;
  async renderFromURL(url: string): Promise<void>;
  async renderFromFile(file: File): Promise<void>;
  
  // Lifecycle management
  async initialize(): Promise<void>;
  destroy(): void;
  pause(): void;
  resume(): void;
  
  // State management
  getRenderingState(): RenderingState;
  getPerformanceMetrics(): PerformanceMetrics;
  getSecurityReport(): SecurityReport;
  
  // Event handling
  addEventListener(event: string, handler: EventHandler): void;
  removeEventListener(event: string, handler: EventHandler): void;
  
  // Configuration
  updateOptions(options: Partial<SecureRenderingOptions>): void;
  setSecurityPolicy(policy: SecurityPolicy): void;
  
  // Utility methods
  takeScreenshot(): Promise<Blob>;
  exportToPDF(): Promise<Blob>;
  print(): void;
}
```

## 🖥️ CLI Tools API

### liv-cli

```bash
# Document operations
liv-cli build <source> [options]     # Build a LIV document
liv-cli validate <document> [options] # Validate a document
liv-cli convert <input> <output> [options] # Convert between formats
liv-cli sign <document> <key> [options] # Sign a document
liv-cli verify <document> [options]  # Verify signatures

# Project operations
liv-cli init [directory] [options]   # Initialize new project
liv-cli serve [document] [options]   # Serve document for development
liv-cli watch [source] [options]     # Watch and rebuild on changes

# Asset operations
liv-cli optimize <assets> [options]  # Optimize assets
liv-cli extract <document> [options] # Extract assets from document
liv-cli info <document> [options]    # Show document information

# Utility operations
liv-cli version                      # Show version information
liv-cli help [command]               # Show help information
```

---

*This API reference is for LIV format version 1.0.0. For the latest updates, visit our [documentation site](https://docs.liv-format.org).*