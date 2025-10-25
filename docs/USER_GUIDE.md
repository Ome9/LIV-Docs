# LIV Document Format - User Guide

## Table of Contents

1. [Introduction](#introduction)
2. [Installation](#installation)
3. [Quick Start](#quick-start)
4. [CLI Tools](#cli-tools)
5. [JavaScript SDK](#javascript-sdk)
6. [Python SDK](#python-sdk)
7. [Desktop Application](#desktop-application)
8. [WYSIWYG Editor](#wysiwyg-editor)
9. [Security Features](#security-features)
10. [Performance Optimization](#performance-optimization)
11. [Troubleshooting](#troubleshooting)
12. [API Reference](#api-reference)

## Introduction

The LIV Document Format is a modern, secure, and interactive document format that combines the power of HTML, CSS, JavaScript, and WebAssembly (WASM) in a single, portable file. LIV documents are self-contained, digitally signed, and can include interactive content while maintaining strict security policies.

### Key Features

- **Interactive Content**: Support for JavaScript and WASM modules
- **Security-First**: Sandboxed execution with granular permissions
- **Cross-Platform**: Works on Windows, macOS, Linux, and web browsers
- **Self-Contained**: All assets embedded in a single .liv file
- **Digital Signatures**: Cryptographic verification of document integrity
- **Performance Optimized**: Built-in compression and caching
- **Multiple Interfaces**: CLI tools, SDKs, desktop app, and web viewer

## Installation

### Prerequisites

- **Go 1.19+** (for CLI tools and core functionality)
- **Node.js 16+** (for JavaScript SDK and web components)
- **Python 3.8+** (for Python SDK)
- **Rust 1.65+** (for WASM modules)

### Build from Source

```bash
# Clone the repository
git clone https://github.com/your-org/liv-document-format.git
cd liv-document-format

# Install dependencies and build all components
make install
make build

# Run tests to verify installation
make test
```

### Install CLI Tools

```bash
# Build CLI tools
make build-go

# Add to PATH (Linux/macOS)
export PATH=$PATH:$(pwd)/bin

# Add to PATH (Windows)
set PATH=%PATH%;%CD%\bin
```

### Install JavaScript SDK

```bash
cd js
npm install
npm run build

# For use in your project
npm install /path/to/liv-document-format/js
```

### Install Python SDK

```bash
cd python
pip install -e .

# Or install from PyPI (when published)
pip install liv-document-format
```

### Install Desktop Application

```bash
cd desktop
npm install
npm run build

# Platform-specific builds
npm run build:windows
npm run build:macos
npm run build:linux
```

## Quick Start

### Creating Your First LIV Document

#### Using CLI Tools

```bash
# Create a simple HTML file
echo '<html><body><h1>Hello LIV!</h1></body></html>' > hello.html

# Build a LIV document
liv-cli build --source . --output hello.liv

# View the document
liv-cli view hello.liv

# Validate the document
liv-cli validate hello.liv
```

#### Using JavaScript SDK

```javascript
import { LIVSDK } from 'liv-document-format';

async function createDocument() {
    const sdk = LIVSDK.getInstance();
    
    const builder = await sdk.createDocument({
        metadata: {
            title: 'My First LIV Document',
            author: 'Your Name',
            description: 'A simple LIV document'
        }
    });
    
    const document = await builder
        .setHTML('<html><body><h1>Hello from JavaScript!</h1></body></html>')
        .setCSS('body { font-family: Arial, sans-serif; }')
        .build();
    
    console.log('Document created successfully!');
}

createDocument();
```

#### Using Python SDK

```python
from liv import LIVDocument, DocumentMetadata

# Create document metadata
metadata = DocumentMetadata(
    title="My First LIV Document",
    author="Your Name",
    description="A simple LIV document"
)

# Create document
doc = LIVDocument()
doc.metadata = metadata
doc.html_content = '<html><body><h1>Hello from Python!</h1></body></html>'
doc.css_content = 'body { font-family: Arial, sans-serif; }'

# Save document
doc.save('hello.liv')
print('Document created successfully!')
```

## CLI Tools

The LIV CLI provides comprehensive command-line tools for document management.

### Available Commands

```bash
liv-cli --help
```

#### Build Command

Create LIV documents from source files:

```bash
# Basic build
liv-cli build --source ./my-document --output document.liv

# Build with signing
liv-cli build --source ./my-document --output document.liv --sign --key private.pem

# Build with custom manifest
liv-cli build --source ./my-document --output document.liv --manifest custom-manifest.json

# Build with compression
liv-cli build --source ./my-document --output document.liv --compress
```

#### View Command

Open and view LIV documents:

```bash
# View in default viewer
liv-cli view document.liv

# View in web browser
liv-cli view document.liv --browser

# View with specific renderer
liv-cli view document.liv --renderer canvas

# Headless viewing (for testing)
liv-cli view document.liv --headless
```

#### Validate Command

Validate LIV document structure and security:

```bash
# Basic validation
liv-cli validate document.liv

# Validate with security checks
liv-cli validate document.liv --security

# Validate signature
liv-cli validate document.liv --verify-signature --key public.pem

# Detailed validation report
liv-cli validate document.liv --verbose
```

#### Convert Command

Convert between different document formats:

```bash
# Convert to PDF
liv-cli convert document.liv --format pdf --output document.pdf

# Convert to HTML
liv-cli convert document.liv --format html --output document.html

# Convert to EPUB
liv-cli convert document.liv --format epub --output document.epub

# Convert from HTML to LIV
liv-cli convert document.html --format liv --output document.liv
```

#### Extract Command

Extract contents from LIV documents:

```bash
# Extract all contents
liv-cli extract document.liv --output ./extracted

# Extract specific files
liv-cli extract document.liv --files "*.html,*.css" --output ./extracted

# Extract manifest only
liv-cli extract document.liv --manifest-only --output manifest.json
```

#### Sign Command

Digitally sign LIV documents:

```bash
# Sign document
liv-cli sign document.liv --key private.pem --output signed-document.liv

# Verify signature
liv-cli verify signed-document.liv --key public.pem

# Generate key pair
liv-cli keygen --output-private private.pem --output-public public.pem
```

### Configuration

Create a configuration file at `~/.liv/config.yaml`:

```yaml
# Default settings
default_author: "Your Name"
default_license: "MIT"
compression_enabled: true
security_level: "strict"

# Signing settings
signing:
  default_key: "~/.liv/keys/private.pem"
  algorithm: "RSA-SHA256"

# Viewer settings
viewer:
  default_renderer: "webgl"
  enable_animations: true
  sandbox_mode: true

# Performance settings
performance:
  memory_limit: "512MB"
  cache_size: "100MB"
  parallel_processing: true
```

## JavaScript SDK

The JavaScript SDK provides a high-level API for creating and managing LIV documents in web applications and Node.js.

### Installation

```bash
npm install liv-document-format
```

### Basic Usage

```javascript
import { LIVSDK, LIVHelpers } from 'liv-document-format';

// Get SDK instance
const sdk = LIVSDK.getInstance();

// Create a document
const builder = await sdk.createDocument({
    metadata: {
        title: 'Interactive Document',
        author: 'Developer',
        description: 'A document with interactive content'
    },
    features: {
        animations: true,
        interactivity: true,
        charts: true
    }
});

// Add content
const document = await builder
    .setHTML(`
        <html>
        <head>
            <title>Interactive Document</title>
        </head>
        <body>
            <h1>Interactive Content</h1>
            <div id="chart-container"></div>
            <button onclick="updateChart()">Update Chart</button>
        </body>
        </html>
    `)
    .setCSS(`
        body { 
            font-family: -apple-system, BlinkMacSystemFont, sans-serif;
            margin: 20px;
        }
        #chart-container {
            width: 100%;
            height: 400px;
            border: 1px solid #ddd;
        }
    `)
    .setInteractiveSpec(`
        function updateChart() {
            // Chart update logic
            console.log('Chart updated!');
        }
    `)
    .build();

// Validate document
const validation = await sdk.validateDocument(document);
if (validation.isValid) {
    console.log('Document is valid!');
} else {
    console.error('Validation errors:', validation.errors);
}
```

### Adding Assets

```javascript
// Add image asset
const imageData = await fetch('logo.png').then(r => r.arrayBuffer());
builder.addAsset({
    type: 'image',
    name: 'logo.png',
    data: imageData,
    mimeType: 'image/png'
});

// Add font asset
const fontData = await fetch('custom-font.woff2').then(r => r.arrayBuffer());
builder.addAsset({
    type: 'font',
    name: 'custom-font.woff2',
    data: fontData,
    mimeType: 'font/woff2'
});
```

### Adding WASM Modules

```javascript
// Add WASM module
const wasmData = await fetch('chart-engine.wasm').then(r => r.arrayBuffer());
builder.addWASMModule({
    name: 'chart-engine',
    data: wasmData,
    version: '1.0',
    entryPoint: 'main',
    permissions: {
        memoryLimit: 32 * 1024 * 1024, // 32MB
        allowedImports: ['env'],
        allowNetworking: false,
        allowFileSystem: false
    }
});
```

### Helper Functions

```javascript
// Create text document
const textDoc = await LIVHelpers.createTextDocument(
    'My Article',
    'This is the content of my article...',
    'Author Name'
);

// Create chart document
const chartDoc = await LIVHelpers.createChartDocument(
    'Sales Report',
    {
        labels: ['Q1', 'Q2', 'Q3', 'Q4'],
        values: [100, 150, 120, 180]
    },
    'bar'
);

// Create presentation
const presentationDoc = await LIVHelpers.createPresentationDocument(
    'My Presentation',
    [
        { title: 'Introduction', content: 'Welcome to my presentation' },
        { title: 'Main Points', content: 'Here are the key points...' },
        { title: 'Conclusion', content: 'Thank you for your attention' }
    ]
);
```

### Rendering Documents

```javascript
// Create renderer
const container = document.getElementById('document-container');
const renderer = sdk.createRenderer(container, {
    enableInteractivity: true,
    enableAnimations: true,
    fallbackMode: false
});

// Load and render document
const document = await sdk.loadDocument('document.liv');
await renderer.render(document);
```

## Python SDK

The Python SDK provides tools for document automation, batch processing, and server-side document generation.

### Installation

```bash
pip install liv-document-format
```

### Basic Usage

```python
from liv import LIVDocument, DocumentMetadata, SecurityPolicy, FeatureFlags
from liv.builder import LIVBuilder

# Create document using builder pattern
builder = LIVBuilder()

# Set metadata
builder.set_metadata(
    title="Python Generated Document",
    author="Python Developer",
    description="A document created with the Python SDK"
)

# Set content
builder.set_html_content("""
<html>
<head>
    <title>Python Document</title>
</head>
<body>
    <h1>Generated with Python</h1>
    <p>This document was created using the LIV Python SDK.</p>
</body>
</html>
""")

builder.set_css_content("""
body {
    font-family: 'Segoe UI', Tahoma, Geneva, Verdana, sans-serif;
    margin: 40px;
    line-height: 1.6;
}
h1 {
    color: #2c3e50;
}
""")

# Enable features
builder.enable_interactivity()
builder.enable_animations()

# Build and save document
document = builder.build()
document.save('python-document.liv')
```

### Batch Processing

```python
from liv.batch_processor import BatchProcessor
import os

# Create batch processor
processor = BatchProcessor()

# Process directory of HTML files
input_dir = './html-documents'
output_dir = './liv-documents'

# Configure batch processing
processor.configure(
    input_pattern='*.html',
    output_format='liv',
    parallel_processing=True,
    max_workers=4
)

# Process all files
results = processor.process_directory(input_dir, output_dir)

for result in results:
    if result.success:
        print(f"✓ Converted {result.input_file} -> {result.output_file}")
    else:
        print(f"✗ Failed to convert {result.input_file}: {result.error}")
```

### Async Processing

```python
import asyncio
from liv.async_processor import AsyncProcessor

async def process_documents():
    processor = AsyncProcessor()
    
    # Process multiple documents concurrently
    tasks = [
        processor.process_async('doc1.html', 'doc1.liv'),
        processor.process_async('doc2.html', 'doc2.liv'),
        processor.process_async('doc3.html', 'doc3.liv')
    ]
    
    results = await asyncio.gather(*tasks)
    
    for result in results:
        print(f"Processed: {result.output_file}")

# Run async processing
asyncio.run(process_documents())
```

### CLI Integration

```python
from liv.cli_interface import CLIInterface

# Create CLI interface
cli = CLIInterface()

# Check if CLI tools are available
if cli.check_cli_available():
    # Build document using CLI
    result = cli.build(
        source_dir='./source',
        output_path='document.liv',
        sign=True,
        key_path='private.pem'
    )
    
    if result.success:
        print("Document built successfully!")
        
        # Validate the document
        validation = cli.validate('document.liv')
        print(f"Validation result: {validation.is_valid}")
    else:
        print(f"Build failed: {result.error}")
else:
    print("CLI tools not available, using Python-only processing")
```

## Desktop Application

The LIV Desktop Application provides a native viewing and editing experience for LIV documents.

### Installation

Download the appropriate installer for your platform:

- **Windows**: `LIV-Document-Viewer-Setup.exe`
- **macOS**: `LIV-Document-Viewer.dmg`
- **Linux**: `liv-document-viewer.AppImage`

### Features

- **Document Viewing**: Native rendering of LIV documents
- **File Association**: Double-click .liv files to open
- **Security Sandbox**: Safe execution of interactive content
- **Performance Optimization**: Hardware-accelerated rendering
- **Offline Support**: No internet connection required

### Usage

```bash
# Open document
liv-desktop document.liv

# Open with specific options
liv-desktop document.liv --fullscreen --no-sandbox

# Batch convert documents
liv-desktop --convert *.html --output-dir ./converted
```

### Configuration

The desktop app stores configuration in:

- **Windows**: `%APPDATA%/LIV Document Viewer/config.json`
- **macOS**: `~/Library/Application Support/LIV Document Viewer/config.json`
- **Linux**: `~/.config/liv-document-viewer/config.json`

```json
{
  "viewer": {
    "default_zoom": 1.0,
    "enable_animations": true,
    "hardware_acceleration": true,
    "sandbox_mode": "strict"
  },
  "security": {
    "allow_network_access": false,
    "allow_file_system": false,
    "memory_limit": "512MB"
  },
  "performance": {
    "cache_size": "100MB",
    "preload_assets": true,
    "lazy_loading": true
  }
}
```

## WYSIWYG Editor

The LIV WYSIWYG Editor provides visual editing capabilities for creating interactive documents.

### Features

- **Visual Editing**: Drag-and-drop interface
- **Code Editor**: Syntax-highlighted HTML/CSS/JS editing
- **Live Preview**: Real-time preview of changes
- **Asset Management**: Easy asset upload and management
- **WASM Integration**: Visual WASM module configuration
- **Security Configuration**: Visual security policy editor

### Usage

```bash
# Start editor
liv-editor

# Open existing document
liv-editor document.liv

# Create new document from template
liv-editor --template interactive-presentation
```

### Editor Interface

The editor provides multiple views:

1. **Design View**: Visual editing with drag-and-drop
2. **Code View**: Direct HTML/CSS/JavaScript editing
3. **Preview View**: Live preview of the document
4. **Assets View**: Asset management and upload
5. **Security View**: Security policy configuration

### Templates

Available document templates:

- `basic-document`: Simple HTML document
- `interactive-presentation`: Slideshow with animations
- `data-visualization`: Charts and graphs
- `interactive-story`: Narrative with interactive elements
- `technical-documentation`: Code documentation with examples

## Security Features

LIV documents implement comprehensive security measures to ensure safe execution of interactive content.

### Security Policies

Every LIV document includes a security policy that defines:

```json
{
  "wasmPermissions": {
    "memoryLimit": 67108864,
    "allowNetworking": false,
    "allowFileSystem": false,
    "allowedImports": ["env"]
  },
  "jsPermissions": {
    "executionMode": "sandboxed",
    "allowedAPIs": ["dom", "canvas"],
    "domAccess": "read-write"
  },
  "networkPolicy": {
    "allowOutbound": false,
    "allowedHosts": [],
    "allowedPorts": []
  },
  "storagePolicy": {
    "allowLocalStorage": false,
    "allowSessionStorage": false,
    "allowIndexedDB": false,
    "allowCookies": false
  }
}
```

### Digital Signatures

Sign documents for integrity verification:

```bash
# Generate key pair
liv-cli keygen --output-private private.pem --output-public public.pem

# Sign document
liv-cli sign document.liv --key private.pem --output signed-document.liv

# Verify signature
liv-cli verify signed-document.liv --key public.pem
```

### Sandboxing

All interactive content runs in a secure sandbox:

- **Memory Isolation**: Limited memory allocation
- **Network Restrictions**: Controlled network access
- **File System Protection**: No direct file system access
- **API Restrictions**: Limited browser API access

### Security Administration

Use the security administration tools:

```bash
# Start security admin server
security-admin --port 8080

# Configure global security policies
security-admin configure --policy strict

# Monitor security events
security-admin monitor --log-level info
```

## Performance Optimization

LIV includes built-in performance optimization features.

### Automatic Optimizations

- **Asset Compression**: Automatic compression of large assets
- **Memory Pooling**: Efficient memory management
- **Caching**: Multi-level caching system
- **Lazy Loading**: Load assets on demand
- **Parallel Processing**: Concurrent operations

### Performance Monitoring

Enable performance monitoring:

```javascript
import { EnableGlobalMonitoring, GenerateGlobalReport } from 'liv-document-format/performance';

// Enable monitoring
EnableGlobalMonitoring();

// Your application code here...

// Generate performance report
const report = GenerateGlobalReport();
console.log('Performance Report:', report);
```

### Optimization Configuration

Configure optimization settings:

```yaml
# .liv/performance.yaml
optimization:
  memory_pooling: true
  compression: true
  caching: true
  max_memory_usage: "512MB"
  compression_threshold: 1024
  cache_size: 1000
  gc_interval: "5m"
```

### Performance Testing

Run performance tests:

```bash
# Run all performance tests
make test-performance

# Run specific performance tests
make test-performance FILTER="memory"

# Generate performance report
make benchmark > performance-report.txt
```

## Troubleshooting

### Common Issues

#### Build Failures

**Issue**: `liv-cli build` fails with "manifest not found"

**Solution**:
```bash
# Ensure manifest.json exists in source directory
ls -la ./source/manifest.json

# Or let CLI generate manifest automatically
liv-cli build --source ./source --output document.liv --auto-manifest
```

**Issue**: WASM module compilation fails

**Solution**:
```bash
# Check Rust installation
rustc --version

# Install WASM target
rustup target add wasm32-unknown-unknown

# Install wasm-pack
cargo install wasm-pack
```

#### Runtime Errors

**Issue**: JavaScript execution blocked by security policy

**Solution**:
```json
// Update security policy in manifest.json
{
  "jsPermissions": {
    "executionMode": "sandboxed",
    "allowedAPIs": ["dom", "canvas", "console"]
  }
}
```

**Issue**: WASM module fails to load

**Solution**:
```json
// Check WASM permissions
{
  "wasmPermissions": {
    "memoryLimit": 67108864,
    "allowedImports": ["env", "your_module"]
  }
}
```

#### Performance Issues

**Issue**: Slow document loading

**Solution**:
```bash
# Enable compression
liv-cli build --source ./source --output document.liv --compress

# Optimize assets
liv-cli optimize --input document.liv --output optimized.liv
```

**Issue**: High memory usage

**Solution**:
```javascript
// Enable memory optimization
import { OptimizeMemoryUsage } from 'liv-document-format/performance';

// Periodically optimize memory
setInterval(() => {
    OptimizeMemoryUsage();
}, 60000); // Every minute
```

### Debug Mode

Enable debug mode for detailed logging:

```bash
# CLI debug mode
LIV_DEBUG=true liv-cli build --source ./source --output document.liv

# JavaScript debug mode
localStorage.setItem('LIV_DEBUG', 'true');

# Python debug mode
import os
os.environ['LIV_DEBUG'] = 'true'
```

### Log Files

Check log files for detailed error information:

- **CLI Logs**: `~/.liv/logs/cli.log`
- **Desktop App Logs**: `~/.liv/logs/desktop.log`
- **Security Logs**: `~/.liv/logs/security.log`

### Getting Help

- **Documentation**: [https://liv-format.org/docs](https://liv-format.org/docs)
- **GitHub Issues**: [https://github.com/your-org/liv-document-format/issues](https://github.com/your-org/liv-document-format/issues)
- **Community Forum**: [https://community.liv-format.org](https://community.liv-format.org)
- **Email Support**: support@liv-format.org

## API Reference

### Core Types

```typescript
interface DocumentMetadata {
    title: string;
    author: string;
    description?: string;
    version?: string;
    created?: string;
    modified?: string;
    language?: string;
    keywords?: string[];
    license?: string;
}

interface SecurityPolicy {
    wasmPermissions: WASMPermissions;
    jsPermissions: JSPermissions;
    networkPolicy: NetworkPolicy;
    storagePolicy: StoragePolicy;
}

interface WASMPermissions {
    memoryLimit: number;
    allowNetworking: boolean;
    allowFileSystem: boolean;
    allowedImports: string[];
    cpuTimeLimit?: number;
}

interface JSPermissions {
    executionMode: 'sandboxed' | 'restricted' | 'full';
    allowedAPIs: string[];
    domAccess: 'none' | 'read' | 'read-write';
}
```

### CLI Commands

```bash
# Build command
liv-cli build [options]
  --source <dir>        Source directory
  --output <file>       Output .liv file
  --manifest <file>     Custom manifest file
  --sign               Sign the document
  --key <file>         Private key for signing
  --compress           Enable compression
  --auto-manifest      Generate manifest automatically

# View command
liv-cli view <file> [options]
  --browser            Open in web browser
  --renderer <type>    Renderer type (webgl|canvas|svg)
  --headless           Headless mode
  --fullscreen         Fullscreen mode

# Validate command
liv-cli validate <file> [options]
  --security           Enable security validation
  --verify-signature   Verify digital signature
  --key <file>         Public key for verification
  --verbose            Detailed output

# Convert command
liv-cli convert <input> [options]
  --format <type>      Output format (pdf|html|epub|liv)
  --output <file>      Output file
  --quality <level>    Conversion quality (low|medium|high)
```

### JavaScript SDK API

```typescript
class LIVSDK {
    static getInstance(): LIVSDK;
    createDocument(options?: DocumentCreationOptions): Promise<LIVDocumentBuilder>;
    loadDocument(source: File | string | ArrayBuffer): Promise<LIVDocument>;
    createRenderer(container: HTMLElement, options?: RenderingOptions): LIVRenderer;
    validateDocument(document: LIVDocument): Promise<ValidationResult>;
}

class LIVDocumentBuilder {
    setMetadata(metadata: Partial<DocumentMetadata>): LIVDocumentBuilder;
    setHTML(html: string): LIVDocumentBuilder;
    setCSS(css: string): LIVDocumentBuilder;
    setInteractiveSpec(spec: string): LIVDocumentBuilder;
    addAsset(options: AssetOptions): LIVDocumentBuilder;
    addWASMModule(options: WASMModuleOptions): LIVDocumentBuilder;
    build(): Promise<LIVDocument>;
}
```

### Python SDK API

```python
class LIVDocument:
    def __init__(self, file_path: Optional[str] = None)
    def load(self, file_path: str) -> None
    def save(self, output_path: str, sign: bool = False) -> None
    def validate(self) -> bool
    def get_asset(self, name: str) -> Optional[AssetInfo]
    def list_assets(self, asset_type: Optional[str] = None) -> List[AssetInfo]

class LIVBuilder:
    def set_metadata(self, **kwargs) -> 'LIVBuilder'
    def set_html_content(self, html: str) -> 'LIVBuilder'
    def set_css_content(self, css: str) -> 'LIVBuilder'
    def add_asset(self, asset_type: str, name: str, data: bytes) -> 'LIVBuilder'
    def enable_interactivity(self) -> 'LIVBuilder'
    def build(self) -> LIVDocument
```

---

This completes the comprehensive user guide for the LIV Document Format. The system provides multiple interfaces (CLI, JavaScript SDK, Python SDK, Desktop App) for creating, viewing, and managing interactive documents with strong security and performance features.