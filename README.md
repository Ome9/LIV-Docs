# LIV Document Format - Live Interactive Visual Documents

A modern, secure, and interactive document format that combines HTML, CSS, JavaScript, and WebAssembly in a single portable file. Now featuring a comprehensive PDF editor with beautiful animations and modern UI.

## 🚀 Features

- **📄 Interactive Content**: Support for JavaScript and WebAssembly modules
- **🔒 Security-First Design**: Sandboxed execution with granular permissions  
- **🌐 Cross-Platform**: Works on Windows, macOS, Linux, and web browsers
- **📦 Self-Contained**: All assets embedded in a single .liv file
- **✍️ Digital Signatures**: Cryptographic verification of document integrity
- **⚡ Performance Optimized**: Built-in compression and caching
- **🛠️ Multiple SDKs**: JavaScript, Python, and CLI interfaces
- **🎨 WYSIWYG Editor**: Visual document creation tool
- **🖥️ Desktop Application**: Native cross-platform viewer
- **📝 PDF Editor**: Full-featured PDF editor with 25+ operations
- **🎭 Beautiful Animations**: Powered by Anime.js with 10+ animation types
- **🎨 Google Fonts**: 8 integrated Google Fonts with 12 total font options
- **🌈 Color Presets**: 42 curated colors from Material Design & Tailwind CSS

## Architecture

The LIV format uses a multi-layer architecture for optimal performance and security:

- **Go Core Layer**: Handles packaging, manifest management, signatures, security orchestration, and WASM module loading
- **Rust WASM Layer**: Runs memory-safe interactive logic, live graphs, animations, and vector rendering  
- **Minimal JS Layer**: Provides sandboxed DOM/CSS updates based on WASM render outputs

## Features

- 📦 **Single-file container** - All content, assets, and metadata in one portable file
- 🔒 **Secure execution** - Sandboxed JavaScript, signed manifests, permission controls
- ✏️ **Editable content** - WYSIWYG and source-level editing capabilities
- 🎬 **Live animations** - CSS animations, SVG vectors, interactive charts
- 📱 **Cross-platform** - Desktop, mobile, and web compatibility
- 🔄 **Format conversion** - Export/import to PDF, HTML, Markdown, EPUB

## Quick Start

### Prerequisites

- Go 1.21+
- Rust 1.70+
- Node.js 18+
- wasm-pack

### Installation

```bash
# Install dependencies
make install

# Build all components
make build

# Run tests
make test
```

### Desktop Application

The LIV desktop application includes a comprehensive PDF editor with modern UI and beautiful animations.

```bash
# Navigate to desktop folder
cd desktop

# Install dependencies
npm install

# Run the application
npm start
```

#### PDF Editor Features

**PDF Operations (25+ methods)**:
- Create, open, merge, split PDFs
- Rotate, compress, encrypt/decrypt
- Add watermarks, stamps, signatures
- Form filling and annotations
- Bookmarks and attachments
- QR codes and barcodes

**Editing Tools**:
- Text tool with Google Fonts (8 fonts integrated)
- Image tool with drag & drop
- Shape tools (rectangle, circle, line)
- Signature and stamp tools
- Component library with drag & drop

**UI/UX Features**:
- Modern dark theme design
- 60+ keyboard shortcuts
- 42 color presets (Material Design, Tailwind CSS, Grayscale)
- 10+ animation types powered by Anime.js
- Real-time preview and zoom controls
- Drag and drop components
- Toast notifications
- Loading animations
- Confetti celebrations

**Keyboard Shortcuts**:
- `Ctrl+Shift+N`: New PDF
- `Ctrl+O`: Open PDF
- `Ctrl+S`: Save PDF
- `Ctrl+Z`: Undo
- `Ctrl+Y`: Redo
- `Ctrl++`: Zoom In
- `Ctrl+-`: Zoom Out
- `Ctrl+/`: Show shortcuts guide
- And 50+ more shortcuts!

### Creating a LIV Document

```bash
# Using the CLI builder
./bin/liv-cli build --input ./examples/sample --output document.liv

# Using the Go API
go run examples/create-document/main.go
```

### Viewing a LIV Document

```bash
# Using the CLI viewer
./bin/liv-viewer document.liv

# Using the web viewer
./bin/liv-viewer --web --port 8080 document.liv
```

## Project Structure

```
├── pkg/core/           # Go core types and interfaces
├── cmd/                # Go CLI applications
│   ├── cli/           # Main CLI tool
│   ├── viewer/        # Document viewer
│   └── builder/       # Document builder
├── wasm/              # Rust WASM modules
│   ├── interactive-engine/  # Interactive content engine
│   └── editor-engine/      # Editor logic engine
├── js/                # JavaScript/TypeScript interfaces
│   ├── src/          # Source code
│   ├── wasm/         # Generated WASM bindings
│   └── dist/         # Built JavaScript
├── desktop/           # Electron desktop application
│   ├── src/          # Application source
│   │   ├── pdf-editor.html      # PDF editor UI
│   │   ├── pdf-editor.css       # Editor styles (1,200+ lines)
│   │   ├── pdf-editor.js        # Editor logic (1,500+ lines)
│   │   ├── pdf-operations.js    # PDF operations (25 methods)
│   │   └── keybindings-manager.js # Keyboard shortcuts (60+)
│   ├── main.js       # Electron main process (22 IPC handlers)
│   └── package.json  # Desktop dependencies
├── examples/          # Example documents and code
├── docs/             # Documentation
└── tests/            # Test files
```

## File Format

A .liv file is a ZIP container with the following structure:

```
document.liv
├── manifest.json          # Document metadata and security
├── content/
│   ├── index.html         # Main document structure
│   ├── styles/main.css    # Stylesheets
│   ├── scripts/main.js    # Interactive functionality
│   └── static/fallback.html  # Static fallback
├── assets/
│   ├── images/           # Image resources
│   ├── fonts/            # Font files
│   └── data/             # JSON/CSV data
└── signatures/
    ├── content.sig       # Content signature
    └── manifest.sig      # Manifest signature
```

## Security Model

LIV documents implement multiple security layers:

1. **Digital Signatures** - All content is cryptographically signed
2. **Permission System** - Granular controls for WASM and JavaScript execution
3. **Sandboxed Execution** - All interactive content runs in isolated environments
4. **Resource Integrity** - SHA-256 hashing ensures content hasn't been tampered with
5. **Static Fallback** - Non-interactive version available for security-conscious environments

## Development

### Building Components

```bash
# Build Go components
make build-go

# Build WASM modules  
make build-wasm

# Build JavaScript
make build-js

# Development mode with hot reload
make dev
```

### Running Tests

```bash
# All tests
make test

# Specific component tests
make test-go
make test-wasm
make test-js
```

### Code Quality

```bash
# Format code
make fmt

# Run linters
make lint

# Generate documentation
make docs
```

## API Examples

### Go API

```go
package main

import (
    "github.com/liv-format/liv/pkg/core"
    "github.com/liv-format/liv/pkg/builder"
)

func main() {
    // Create a new document
    doc := &core.LIVDocument{
        Manifest: &core.Manifest{
            Version: "1.0",
            Metadata: &core.DocumentMetadata{
                Title:  "My Document",
                Author: "John Doe",
            },
        },
    }
    
    // Build and save
    builder := builder.New()
    err := builder.Build(doc, "output.liv")
    if err != nil {
        panic(err)
    }
}
```

### JavaScript API

```javascript
import { LIVLoader, LIVRenderer } from 'liv-format';

// Load a document
const loader = new LIVLoader();
const document = await loader.loadFromFile(file);

// Render in browser
const renderer = new LIVRenderer({
    container: document.getElementById('viewer'),
    permissions: document.manifest.security
});

await renderer.loadWASMModule(interactiveEngine);
renderer.startRenderLoop();
```

### Rust WASM API

```rust
use wasm_bindgen::prelude::*;
use liv_interactive_engine::*;

#[wasm_bindgen]
pub fn create_chart(data: &str) -> String {
    let chart = InteractiveElement {
        id: "chart-1".to_string(),
        element_type: ElementType::Chart,
        // ... configuration
    };
    
    serde_json::to_string(&chart).unwrap()
}
```

## Contributing

1. Fork the repository
2. Create a feature branch (`git checkout -b feature/amazing-feature`)
3. Make your changes
4. Run tests (`make test`)
5. Commit your changes (`git commit -m 'Add amazing feature'`)
6. Push to the branch (`git push origin feature/amazing-feature`)
7. Open a Pull Request

## License

This project is licensed under the MIT License - see the [LICENSE](LICENSE) file for details.

## 🗺️ Roadmap

### ✅ Completed (v1.0)
- [x] Core document format and specification
- [x] Go backend with CLI tools (liv-cli, liv-viewer, liv-builder)
- [x] Rust WASM interactive engine  
- [x] JavaScript and Python SDKs
- [x] Desktop application (Electron-based)
- [x] **Comprehensive PDF Editor** with 25+ operations
- [x] **Google Fonts Integration** (8 fonts, 12 total options)
- [x] **Color Presets System** (42 curated colors)
- [x] **Beautiful Animations** (10+ types powered by Anime.js)
- [x] **Keyboard Shortcuts** (60+ customizable shortcuts)
- [x] **Modern UI/UX** (Dark theme, responsive design)
- [x] Security system with sandboxing
- [x] Performance optimization and monitoring
- [x] Comprehensive test suite (unit, integration, e2e, performance)
- [x] Cross-platform compatibility (Windows, macOS, Linux)
- [x] Digital signature support
- [x] WYSIWYG editor
- [x] Format conversion tools (PDF, HTML, EPUB, Markdown)

### 🚧 In Progress (v1.1)
- [ ] Enhanced mobile viewer applications
- [ ] Real-time collaboration features
- [ ] Advanced plugin system architecture
- [ ] Cloud-based document hosting
- [ ] PDF editor cloud sync
- [ ] Additional Google Fonts (user requests)
- [ ] Custom color palette creation

### 🔮 Future (v2.0+)
- [ ] AI-powered content generation
- [ ] Blockchain-based verification
- [ ] Extended WASM capabilities
- [ ] Advanced analytics and insights
- [ ] PDF OCR and text extraction
- [ ] Advanced animation presets
- [ ] Template library for common documents

## Support

- 📖 [Documentation](docs/)
- 🐛 [Issue Tracker](https://github.com/liv-format/liv/issues)
- 💬 [Discussions](https://github.com/liv-format/liv/discussions)
- 📧 [Email Support](mailto:support@liv-format.org)