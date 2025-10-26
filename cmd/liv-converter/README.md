# LIV Converter - PDF to LIV Conversion Tool

A modular CLI tool for converting PDF documents into the LIV format with full document structure, assets, and metadata preservation.

## Overview

The LIV Converter (`liv-converter`) is a specialized tool that transforms PDF files into fully-structured LIV documents. It extracts text, images, layout information, and metadata from PDFs and packages everything into a `.liv` archive.

## Features

- **5-Step Conversion Process**:
  1. Parse PDF document structure
  2. Build LIV document elements
  3. Generate manifest with permissions
  4. Extract and optimize assets
  5. Package into `.liv` archive (ZIP format)

- **Commands**:
  - `convert` - Convert PDF to LIV
  - `inspect` - Examine LIV document structure
  - `validate` - Validate LIV format compliance

- **Smart Parsing**:
  - Text extraction with positioning
  - Image extraction and optimization
  - Metadata preservation
  - Layout analysis (TODO)

- **Flexible Output**:
  - Dry-run mode for JSON preview
  - Optional asset compression
  - Configurable image quality
  - Font embedding (TODO)

## Installation

```bash
# Build from source
cd cmd/liv-converter
go build

# Or use the project build script
./build.sh    # Linux/Mac
build.bat     # Windows
```

## Usage

### Basic Conversion

```bash
# Convert a PDF to LIV format
liv-converter convert document.pdf

# Specify output filename
liv-converter convert document.pdf --output=mydoc.liv
```

### Advanced Options

```bash
# Custom title and author
liv-converter convert document.pdf \
  --title="My Document" \
  --author="John Doe"

# Control image quality (1-100)
liv-converter convert document.pdf --quality=90

# Disable compression (faster, larger file)
liv-converter convert document.pdf --compress=false

# Dry run (output JSON without creating .liv)
liv-converter convert document.pdf --dry-run
```

### Inspection & Validation

```bash
# Inspect a LIV document
liv-converter inspect document.liv

# Show detailed content
liv-converter inspect document.liv --show-content

# Output as JSON
liv-converter inspect document.liv --json

# Validate a LIV document
liv-converter validate document.liv

# Strict validation
liv-converter validate document.liv --strict
```

## Architecture

```
internal/
├── types/              # Shared data structures
│   └── types.go        # PDFData, LIVDocument, LIVManifest, etc.
├── converter/          # Main conversion logic
│   ├── converter.go    # Orchestration (ConvertPDFToLIV, etc.)
│   ├── assets.go       # Asset extraction and optimization
│   ├── pdf/            # PDF parsing
│   │   └── parser.go   # UniPDF-based PDF parsing
│   └── liv/            # LIV building
│       ├── builder.go  # PDF→LIV transformation
│       ├── manifest.go # Manifest generation
│       └── packager.go # ZIP packaging & unpacking
```

## Data Flow

```
PDF File
    ↓
[1] Parse PDF (pdf.ParsePDF)
    → PDFData (pages, text, images, metadata)
    ↓
[2] Build LIV Document (liv.BuildLIVDocument)
    → LIVDocument (elements, pages, styles)
    ↓
[3] Generate Manifest (liv.GenerateManifest)
    → LIVManifest (metadata, permissions, assets)
    ↓
[4] Extract Assets (converter.ExtractAssets)
    → ExtractedAssets (optimized images, fonts)
    ↓
[5] Package (liv.PackageLIV)
    → .liv file (ZIP archive)
```

## File Structure

A `.liv` file is a ZIP archive containing:

```
document.liv (ZIP)
├── document.json      # LIV document structure
├── manifest.json      # Metadata & permissions
└── assets/
    ├── images/        # Extracted images
    │   ├── img-1.jpeg
    │   └── img-2.png
    └── fonts/         # Embedded fonts (TODO)
```

## Examples

### Example 1: Simple Conversion

```bash
$ liv-converter convert report.pdf
Converting PDF to LIV...
  Input:  report.pdf
  Output: report.liv

[1/5] Parsing PDF document...
  ✓ Extracted 5 pages

[2/5] Building LIV document structure...
  ✓ Created 127 elements

[3/5] Generating manifest...
  ✓ Manifest version: 1.0

[4/5] Extracting and optimizing assets...
  ✓ Extracted 3 images

[5/5] Creating .liv package...

✓ Conversion complete!
  Output: report.liv (2.34 MB)
  Pages: 5
  Assets: 3 images
```

### Example 2: Dry Run (Preview)

```bash
$ liv-converter convert document.pdf --dry-run

[DRY RUN] Outputting intermediate JSON...

=== MANIFEST ===
{
  "version": "1.0",
  "format": "liv",
  "metadata": {
    "title": "Sample Document",
    "author": "John Doe",
    ...
  },
  "permissions": {
    "allow_scripts": false,
    "allow_external_net": false,
    "allow_print": true,
    "allow_copy": true
  },
  ...
}

=== DOCUMENT (first 50 lines) ===
{
  "version": "1.0",
  "format": "liv",
  "pages": [
    {
      "id": "page-1",
      "number": 1,
      "width": 612,
      "height": 792,
      "elements": [...]
    }
  ]
}

✓ Dry run complete. No .liv file created.
```

### Example 3: Inspection

```bash
$ liv-converter inspect report.liv

Inspecting LIV document: report.liv

=== MANIFEST ===
Version: 1.0
Title: Annual Report
Author: Company Inc.
Pages: 5
Compression: true

=== ASSETS ===
Images: 3
Fonts: 0
Styles: 1

✓ Inspection complete
```

## Development Status

**Completed:**
- ✅ CLI skeleton with Cobra
- ✅ Type definitions (PDFData, LIVDocument, LIVManifest)
- ✅ PDF parser stub (UniPDF integration)
- ✅ LIV builder stub (element conversion)
- ✅ Manifest generator
- ✅ ZIP packager/unpackager
- ✅ Asset extraction stub
- ✅ Inspect & validate commands

**TODO (Future Expansion):**
- [ ] Advanced text positioning (TextMark analysis)
- [ ] Image extraction from PDF (UniPDF ImageExtractor)
- [ ] Vector graphics extraction (lines, shapes)
- [ ] Font embedding
- [ ] Layout detection (columns, tables)
- [ ] CSS/JS component injection
- [ ] CBOR format support
- [ ] Digital signatures
- [ ] Encryption support

## Dependencies

- `github.com/unidoc/unipdf/v3` - PDF parsing and manipulation
- `github.com/spf13/cobra` - CLI framework
- `archive/zip` - ZIP packaging
- `encoding/json` - JSON marshaling
- `image/*` - Image processing

## Integration with Electron

```javascript
// desktop/src/go-backend.js
const goBackend = require('./go-backend');

// Convert PDF to LIV
const result = await goBackend.execute('liv-converter', [
  'convert',
  inputPath,
  '--output=' + outputPath,
  '--quality=90'
]);

console.log('Conversion result:', result);
```

## License

See main project LICENSE file.

## Contributing

See main project CONTRIBUTING.md for contribution guidelines.

---

**Note**: This is a modular skeleton designed for extensibility. Many features are marked with `// TODO:` comments for future implementation. The core conversion flow is functional but may require enhancements for production use.
