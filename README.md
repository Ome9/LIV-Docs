# LIV Document Format - Live Interactive Visual Documents

A modern document editor with Microsoft Word-like interface for creating professional documents. The LIV Editor provides a clean, professional editing experience with PDF import/export, rich formatting, and intuitive toolbar.

## 📝 LIV Editor - Professional Word Processor

The LIV Editor is a **professional document editor** with a clean, Microsoft Word-inspired interface. Create beautifully formatted documents with proper text flow, PDF conversion, and real-time statistics.

### ✨ Key Features - NEW Clean Interface

- **📄 Word-Like Canvas**: ContentEditable document with proper cursor positioning and left-aligned text flow
- **🎨 Professional Toolbar**: Microsoft Office-style toolbar without curves or decorative elements  
- **📥 PDF Import/Export**: Upload PDF files and convert to editable documents
- **✍️ Rich Text Editing**: Bold, italic, underline, strikethrough with live formatting
- **🔤 Font Controls**: Font family dropdown and size input (8-72pt)
- **📐 Text Alignment**: Left, center, right, justify alignment options
- **📝 Lists & Indentation**: Bulleted and numbered lists with indent/outdent
- **🔗 Content Insertion**: Links, images, and tables
- **⌨️ Keyboard Shortcuts**: Ctrl+S (save), Ctrl+Z/Y (undo/redo), Ctrl+B/I/U (formatting)
- **💾 Auto-Save**: Automatically saves every 30 seconds
- **📊 Live Statistics**: Real-time word and character count in status bar
- **🔄 Undo/Redo**: 50-step history for all document changes
- **📄 Page Layout**: Standard 8.5" x 11" pages with 1-inch margins
- **🖥️ Cross-Platform**: Electron-based desktop app for Windows, macOS, Linux

## 🏗️ Editor Interface

### **Header Bar**
- Document title (top-left)
- Save and Export buttons (top-right)

### **Professional Toolbar** (Flat, No Curves)
1. **File Group**: New | Open | Save | PDF Upload
2. **Edit Group**: Undo | Redo | Cut | Copy | Paste  
3. **Font Group**: Font Family Dropdown | Font Size Input
4. **Format Group**: Bold | Italic | Underline | Strikethrough
5. **Align Group**: Left | Center | Right | Justify
6. **List Group**: Bullets | Numbers | Indent | Outdent
7. **Insert Group**: Link | Image | Table
8. **PDF Tools**: Complete PDF manipulation suite

## 🔧 PDF Tools Suite

The LIV Editor includes a comprehensive PDF tools panel with 24 operations, **all powered by Go + UniPDF backend**:

### **How It Works**

1. **UI Layer** (Electron/JavaScript): Clean, professional interface
2. **Bridge Layer** (go-backend.js): Spawns Go binaries and captures output
3. **Processing Layer** (Go/UniPDF): Actual PDF manipulation using `pkg/pdfops`
4. **CLI Tool** (`liv-pdf`): Standalone binary for all PDF operations

### **Available Operations**

#### **Document Operations** (Go Backend)
- **Merge PDFs**: `liv-pdf merge input1.pdf input2.pdf -o merged.pdf`
- **Split PDF**: `liv-pdf split input.pdf 1-3,4-6 -d output/`
- **Compress PDF**: `liv-pdf compress input.pdf -o compressed.pdf`
- **Optimize PDF**: Uses UniPDF optimizer for web/print

#### **Page Operations** (Go Backend)
- **Rotate Pages**: `liv-pdf rotate input.pdf 1,3,5 -a 90 -o rotated.pdf`
- **Reorder Pages**: Extract and reassemble in new order
- **Delete Pages**: Remove specific pages
- **Extract Pages**: `liv-pdf extract-pages input.pdf 1,3,5 -o extracted.pdf`

#### **Content Operations** (Go Backend)
- **Add Watermark**: `liv-pdf watermark input.pdf -t "CONFIDENTIAL" -o watermarked.pdf`
- **Annotate**: Add comments and highlights
- **Redact**: Remove sensitive information
- **Sign PDF**: Digital signature support

#### **Conversion Tools** (Go Backend)
- **PDF to Images**: Convert each page to PNG/JPG
- **PDF to Text**: `liv-pdf extract-text input.pdf` - ✅ **Fully Functional**
- **OCR**: Text from scanned documents
- **Images to PDF**: Create from image files

#### **Security Tools** (Go Backend)
- **Encrypt PDF**: `liv-pdf encrypt input.pdf -p password -o encrypted.pdf`
- **Decrypt PDF**: Remove password protection
- **Set Permissions**: Control print/copy/edit rights
- **Verify Signature**: Check digital signatures

#### **Metadata & Info** (Go Backend)
- **Edit Metadata**: `liv-pdf set-info input.pdf --title "Doc" --author "John" -o updated.pdf`
- **Document Info**: `liv-pdf info input.pdf --json`
- **Font Info**: List embedded fonts
- **Bookmarks**: Table of contents

### **CLI Usage Examples**

```bash
# Extract all text from PDF
liv-pdf extract-text document.pdf

# Merge multiple PDFs
liv-pdf merge file1.pdf file2.pdf file3.pdf -o combined.pdf

# Split PDF into ranges
liv-pdf split document.pdf 1-10,11-20 -d ./split-output

# Extract specific pages
liv-pdf extract-pages document.pdf 1,5,10 -o selected-pages.pdf

# Rotate pages 90 degrees
liv-pdf rotate document.pdf 1,3,5 -a 90 -o rotated.pdf

# Add watermark
liv-pdf watermark document.pdf -t "DRAFT" -o watermarked.pdf

# Compress and optimize
liv-pdf compress large.pdf -o optimized.pdf

# Encrypt with password
liv-pdf encrypt document.pdf -p mypassword -o secure.pdf

# Get document information
liv-pdf info document.pdf --json

# Set metadata
liv-pdf set-info document.pdf --title "My Doc" --author "John Doe" -o updated.pdf

# Convert to LIV format
liv-pdf to-liv document.pdf -o document.liv
```

### **Integration with Main CLI**

```bash
# All PDF operations also available through main CLI
liv pdf extract-text document.pdf
liv pdf to-liv document.pdf -o document.liv
```

### **Document Canvas** (Word-Style)
- **Page Dimensions**: 816px × 1056px (8.5" × 11" at 96 DPI)
- **Margins**: 96px (1 inch) on all sides
- **Content Area**: 624px × 864px editable region
- **Background**: Clean white page on light gray canvas
- **Cursor**: Standard text cursor with proper click-to-type positioning
- **Text Flow**: Left-aligned, line-based like Microsoft Word

### **Status Bar**
- Word count and character count (left)
- Zoom level indicator (right)

## 🚀 Quick Start

### Build Go Binaries

```bash
# On Windows
build.bat

# On Linux/Mac
chmod +x build.sh
./build.sh
```

This builds all Go tools including the `liv-pdf` binary with full PDF operations.

### Install Desktop App

```bash
# Install Electron dependencies
cd desktop
npm install

# Run the editor
npm start
```

### Using the Editor

1. **Create New Document**: Click "New" or press Ctrl+N
2. **Open PDF**: Click "PDF" button to upload and convert PDF
3. **Type and Format**: Click in the document and start typing
4. **Access PDF Tools**: Click "PDF Tools" for 24 professional operations powered by Go backend
5. **Save Document**: Click "Save" or press Ctrl+S

### Keyboard Shortcuts

- `Ctrl+S` - Save document
- `Ctrl+Z` - Undo
- `Ctrl+Y` - Redo  
- `Ctrl+B` - Bold
- `Ctrl+I` - Italic
- `Ctrl+U` - Underline

## 🔧 Architecture

### Core Technologies

- **Go (Backend)**: PDF manipulation, document processing, WASM execution
- **Rust (WASM)**: Performance-critical interactive components
- **Electron**: Cross-platform desktop application
- **JavaScript/CSS**: UI components and live interactive elements

### File Structure

```
liv-file/
├── cmd/                   # Go CLI tools
│   ├── cli/              # Main LIV CLI
│   ├── liv-pdf/          # PDF operations tool (NEW!)
│   ├── builder/          # Document builder
│   └── viewer/           # Document viewer
├── pkg/                   # Go packages
│   ├── core/             # Core LIV format
│   ├── pdfops/           # PDF operations (NEW!)
│   ├── container/        # ZIP container
│   └── manifest/         # Manifest handling
├── desktop/              # Electron app
│   └── src/
│       ├── main.js       # Electron main process
│       ├── go-backend.js # Go binary integration (NEW!)
│       └── liv-editor-clean.*  # Editor UI
└── wasm/                 # Rust WASM modules
```

## 🎯 Technical Architecture
- **� Smart Layout Tools**: Figma-like alignment guides, rulers, grid system, and snapping
- **� Modern UI Library**: Built-in components using React, Tailwind CSS, ShadCN, and Aceternity UI
- **� Rich Data Visualization**: Integrated Chart.js, D3.js, and Three.js for charts and 3D graphics
- **✨ Beautiful Animations**: Powered by Anime.js with live preview
- **🌓 Dark/Light Themes**: Elegant, minimal design with theme switching
- **� JSON-Based Format**: Documents saved as .liv files (JSON) for easy version control
- **🖥️ Cross-Platform**: Desktop application for Windows, macOS, and Linux
- **� Live Preview**: See changes in real-time as you build
- **📱 Responsive Design**: Components adapt to different screen sizes
- **🎨 Custom Styling**: Tailwind-inspired design system with modern gradients
- **🔧 Extensible**: Load custom component packs and libraries

## 🏗️ Editor Layout

The LIV Editor features a professional, Figma-inspired layout:

### **Left Panel: Canvas**
- Drag-and-drop document editing
- Live preview with zoom controls
- Smart alignment guides and rulers
- Grid system with snapping
- Multi-component selection
- Real-time rendering

### **Right Panel: Component Library**
- **UI Components**: Buttons, cards, modals, tabs, accordions, tooltips
- **Charts & Data**: Bar charts, line charts, pie charts, scatter plots (Chart.js, D3.js)
- **Animations**: Fade in, slide up, bounce, rotate effects (Anime.js)
- **Media**: Images, videos, galleries
- **Library Packs**: Chart.js Pack, D3.js Pack, Three.js Pack (extensible)
- Search and filter components

### **Bottom Panel: Properties & Code**
- **Properties Tab**: Edit component properties (position, size, colors, etc.)
- **Code Tab**: View and edit component source code
- **JSON Tab**: View document structure
- **Console Tab**: Debug and preview output

## 🎯 Architecture

The LIV Editor is built on a modular, component-based architecture:

- **Canvas Engine**: Handles document rendering, component placement, and visual editing
- **Component System**: Modular UI components with metadata (category, props, drag-drop support)
- **Layout Tools**: Alignment guides, rulers, snapping, grid system (Figma-like)
- **State Management**: History/undo system, clipboard, document serialization
- **Export System**: Generate standalone HTML, PDF, or JSON exports
- **Theme Engine**: Dark/light themes with modern gradients
- **Animation Framework**: Anime.js integration for smooth transitions

### Technology Stack

- **Frontend**: Vanilla JavaScript (no heavy frameworks for performance)
- **UI Libraries**: React components, Tailwind CSS utilities, ShadCN UI, Aceternity UI
- **Charting**: Chart.js, D3.js, Three.js (via library packs)
- **Animations**: Anime.js for smooth component transitions
- **Desktop**: Electron for cross-platform desktop app
- **File Format**: JSON-based .liv documents with embedded assets

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

### 🖥️ Desktop Application

The LIV Editor desktop application provides a powerful visual interface for creating interactive documents.

```bash
# Navigate to desktop folder
cd desktop

# Install dependencies
npm install

# Run the LIV Editor
npm start
```

#### LIV Editor Features

**Visual Editing**:
- Drag-and-drop component placement
- Multi-select and group editing
- Smart alignment guides (like Figma)
- Rulers with measurements
- Grid system with snap-to-grid
- Live zoom (10%-300%)
- Pan and navigate canvas

**Component Library**:
- **40+ Built-in Components**: UI elements, charts, animations, media
- **Library Packs**: Install Chart.js, D3.js, Three.js packs
- **Component Search**: Quick filtering
- **Categories**: UI, Charts, Animations, Media, Custom

**Document Operations**:
- Create new documents
- Open .liv and .json files
- Save as .liv format
- Export to HTML, PDF, JSON
- Live preview mode

**UI/UX Features**:
- Modern dark/light themes
- Tailwind-inspired design system
- 60+ keyboard shortcuts
- Real-time property editing
- Layers panel for organization
- Code editor with syntax highlighting
- Toast notifications with animations
- Smooth Anime.js transitions

**Keyboard Shortcuts**:
- `Ctrl+N`: New Document
- `Ctrl+O`: Open Document
- `Ctrl+S`: Save Document
- `Ctrl+Z`/`Ctrl+Y`: Undo/Redo
- `Ctrl+C`/`Ctrl+V`/`Ctrl+X`: Copy/Paste/Cut
- `Ctrl++`/`Ctrl+-`: Zoom In/Out
- `Ctrl+'`: Toggle Grid
- `Ctrl+R`: Toggle Rulers
- `Ctrl+;`: Toggle Smart Guides
- `Ctrl+1/2/3`: Switch views (Canvas/Code/Preview)
- `Delete`: Delete component

### Creating a LIV Document

#### Using the Visual Editor (Recommended)

1. Launch the LIV Editor desktop app
2. Drag components from the library onto the canvas
3. Position and customize components using the properties panel
4. Save as `.liv` file

#### Using the CLI Builder

```bash
# Using the CLI builder
./bin/liv-cli build --input ./examples/sample --output document.liv

# Using the Go API
go run examples/create-document/main.go
```

### Viewing a LIV Document

```bash
# Using the desktop app (double-click .liv file)
# Or use the CLI viewer
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
│   │   ├── liv-editor.html      # LIV Editor UI (dual-pane layout)
│   │   ├── liv-editor.css       # Editor styles (Tailwind-inspired)
│   │   ├── liv-editor.js        # Editor core logic (component system)
│   │   ├── components/          # Reusable component library
│   │   │   ├── ui/             # UI components (buttons, cards, etc.)
│   │   │   ├── charts/         # Chart components (Chart.js, D3.js)
│   │   │   ├── animations/     # Animation components (Anime.js)
│   │   │   └── registry.js     # Component metadata registry
│   │   ├── keybinding-manager.js # Keyboard shortcuts
│   │   └── preload.js          # Electron preload
│   ├── main.js       # Electron main process
│   └── package.json  # Desktop dependencies
├── examples/          # Example documents and code
├── docs/             # Documentation
└── tests/            # Test files
```

## File Format

A .liv file is a JSON document with the following structure:

```json
{
  "metadata": {
    "title": "My Interactive Document",
    "author": "Jane Doe",
    "created": "2024-01-01T00:00:00Z",
    "modified": "2024-01-02T00:00:00Z",
    "version": "1.0"
  },
  "components": [
    {
      "id": "comp_123",
      "type": "bar-chart",
      "category": "charts",
      "x": 100,
      "y": 200,
      "width": 400,
      "height": 300,
      "properties": {
        "data": [10, 20, 30, 40],
        "labels": ["Q1", "Q2", "Q3", "Q4"],
        "color": "#3b82f6"
      }
    },
    {
      "id": "comp_456",
      "type": "button",
      "category": "ui",
      "x": 100,
      "y": 550,
      "width": 200,
      "height": 50,
      "properties": {
        "text": "Click Me",
        "color": "#8b5cf6"
      }
    }
  ],
  "assets": {
    "images": {
      "logo.png": "data:image/png;base64,..."
    }
  },
  "libraries": ["chartjs-pack", "anime-pack"],
  "styles": {
    "theme": "dark",
    "primaryColor": "#3b82f6"
  }
}
```

### Component Structure

Each component follows this schema:

```typescript
interface Component {
  id: string;                    // Unique identifier
  type: string;                  // Component type (button, chart, etc.)
  category: string;              // Category (ui, charts, animations, media)
  x: number;                     // X position on canvas
  y: number;                     // Y position on canvas
  width: number;                 // Component width
  height: number;                // Component height
  rotation?: number;             // Rotation in degrees (optional)
  opacity?: number;              // Opacity 0-1 (optional)
  properties: Record<string, any>; // Type-specific properties
}
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