# PDF to LIV Conversion - Integration Guide

## Overview

The LIV Editor now includes full PDF to LIV conversion functionality using the Go-based `liv-converter` tool. This allows users to upload PDF files and convert them to the LIV format with proper structure preservation.

## Architecture

```
┌─────────────────┐
│  Electron UI    │
│  (JavaScript)   │
└────────┬────────┘
         │ IPC
         ▼
┌─────────────────┐
│   Main Process  │
│  (main.js)      │
└────────┬────────┘
         │ spawn
         ▼
┌─────────────────┐
│  liv-converter  │
│  (Go Binary)    │
└────────┬────────┘
         │
         ▼
┌─────────────────┐
│   PDF Parser    │
│   (UniPDF)      │
└─────────────────┘
```

## Components

### 1. Go Converter (`liv-converter`)
- **Location**: `cmd/liv-converter/main.go`
- **Purpose**: CLI tool for PDF to LIV conversion
- **Dependencies**: UniPDF for PDF parsing
- **Commands**:
  - `convert`: Convert PDF to LIV format
  - `inspect`: Inspect LIV file structure
  - `validate`: Validate LIV file integrity

### 2. Converter Package (`internal/converter`)
- **converter.go**: Main orchestration logic
- **types**: Data structure definitions (PDFData, LIVDocument, etc.)
- **pdf/parser.go**: PDF parsing with UniPDF
- **liv/builder.go**: LIV document construction
- **liv/manifest.go**: Manifest generation
- **liv/packager.go**: ZIP packaging
- **assets.go**: Asset extraction and optimization

### 3. Electron Integration
- **go-backend.js**: Bridge to Go binaries
- **main.js**: IPC handlers for file operations
- **preload.js**: API exposure to renderer
- **liv-editor-clean.js**: UI implementation

## Conversion Flow

1. **User Uploads PDF**
   - User clicks "Import PDF" button
   - File picker opens, user selects PDF
   - PDF stored in `this.selectedPDF`

2. **Temporary File Creation**
   - PDF buffer sent to main process via IPC
   - Saved to temporary directory
   - Path: `%TEMP%/liv-editor/temp-{timestamp}.pdf`

3. **Go Converter Execution**
   - Main process spawns `liv-converter.exe`
   - Command: `liv-converter convert input.pdf --output output.liv`
   - Progress streamed back to UI

4. **Conversion Steps** (Go Backend)
   - **Step 1**: Parse PDF with UniPDF
     - Extract metadata (title, author, etc.)
     - Parse each page (text, images, graphics)
     - Preserve positioning information
   
   - **Step 2**: Build LIV Structure
     - Convert PDF pages to LIV pages
     - Transform text blocks to LIV elements
     - Map images to asset references
     - Generate element IDs
   
   - **Step 3**: Generate Manifest
     - Create manifest.json with metadata
     - Set permissions (scripts disabled by default)
     - List assets
     - Record source information
   
   - **Step 4**: Extract Assets
     - Extract images from PDF
     - Optimize based on quality setting
     - Re-encode with compression
   
   - **Step 5**: Package LIV
     - Create ZIP archive
     - Add document.json
     - Add manifest.json
     - Add assets/ directory
     - Save as .liv file

5. **Load Converted Document**
   - Read .liv file using JSZip
   - Parse document.json
   - Extract text content from elements
   - Display in editor

## File Structure

### PDF Data Structure
```javascript
{
  metadata: {
    title, author, subject, keywords,
    creator, producer, createdAt, modifiedAt
  },
  pages: [
    {
      number, width, height, rotation,
      textBlocks: [{ text, x, y, width, height, fontSize, color }],
      images: [{ id, x, y, width, height, data, format }],
      graphics: [{ type, x, y, width, height, color, path }]
    }
  ]
}
```

### LIV Document Structure
```javascript
{
  version: "1.0",
  format: "liv",
  pages: [
    {
      id: "page-1",
      number: 1,
      width: 612,
      height: 792,
      elements: [
        {
          id: "text-1",
          type: "text",
          content: "...",
          position: { x, y, width, height },
          style: { fontFamily, fontSize, color }
        }
      ]
    }
  ],
  styles: { /* CSS-like styles */ },
  scripts: []
}
```

### LIV Manifest Structure
```javascript
{
  version: "1.0",
  format: "liv",
  metadata: { /* document metadata */ },
  permissions: {
    allowScripts: false,
    allowExternalNet: false,
    allowPrint: true,
    allowCopy: true,
    allowModify: false
  },
  pages: 5,
  assets: {
    images: ["img-1.jpeg", "img-2.png"],
    fonts: [],
    styles: []
  },
  compression: true,
  source: {
    type: "pdf",
    original: "document.pdf"
  }
}
```

## API Reference

### JavaScript (Renderer Process)

```javascript
// Convert PDF to LIV
const result = await window.electronAPI.goPDFToLIV({
  inputPath: '/path/to/input.pdf',
  outputPath: '/path/to/output.liv'
});

// Save temporary file
const tempPath = await window.electronAPI.saveTempFile({
  name: 'document.pdf',
  buffer: arrayBuffer
});

// Read LIV file
const livContent = await window.electronAPI.readLIVFile('/path/to/document.liv');
```

### Go (Command Line)

```bash
# Convert PDF to LIV
liv-converter convert input.pdf --output output.liv

# With options
liv-converter convert input.pdf \
  --output output.liv \
  --title "My Document" \
  --author "John Doe" \
  --compress \
  --quality 85

# Dry run (output JSON only)
liv-converter convert input.pdf --dry-run

# Inspect LIV file
liv-converter inspect document.liv --show-content

# Validate LIV file
liv-converter validate document.liv --strict
```

## Configuration Options

| Option | Type | Default | Description |
|--------|------|---------|-------------|
| `--output` | string | input+".liv" | Output file path |
| `--title` | string | from PDF | Document title override |
| `--author` | string | from PDF | Document author override |
| `--compress` | bool | true | Enable compression |
| `--quality` | int | 85 | Image quality (0-100) |
| `--embed-fonts` | bool | false | Embed fonts (not impl.) |
| `--dry-run` | bool | false | Output JSON only |

## Testing

### Manual Test
1. Run desktop app: `cd desktop && npm start`
2. Click "Import PDF" button
3. Select a PDF file
4. Click "Convert" button
5. Verify content appears in editor

### CLI Test
```bash
# Build converter
go build -o bin/liv-converter.exe ./cmd/liv-converter

# Convert test PDF
bin\liv-converter.exe convert test-document.pdf --output test-output.liv

# Inspect output
bin\liv-converter.exe inspect test-output.liv --show-content

# Validate
bin\liv-converter.exe validate test-output.liv
```

## Troubleshooting

### "liv-converter.exe not found"
- **Cause**: Binary not built or not in correct location
- **Solution**: Run `go build -o bin/liv-converter.exe ./cmd/liv-converter`
- **Check**: Verify `bin/liv-converter.exe` exists

### "Failed to parse PDF"
- **Cause**: PDF is encrypted, corrupted, or unsupported version
- **Solution**: Try opening PDF in another viewer, decrypt if needed
- **Check**: Run `liv-converter convert file.pdf` from command line

### "No content extracted"
- **Cause**: PDF contains only images (scanned document)
- **Solution**: Use OCR tool first to add text layer
- **Note**: Image extraction is basic, needs enhancement

### "Conversion successful but content not loading"
- **Cause**: LIV file structure issue or parsing error
- **Solution**: Run `liv-converter inspect output.liv` to check structure
- **Check**: Verify document.json exists in ZIP

## Future Enhancements

### High Priority
- [ ] Improve text positioning accuracy
- [ ] Extract embedded images
- [ ] Preserve text formatting (bold, italic)
- [ ] Detect tables and convert to structured elements
- [ ] Handle multi-column layouts

### Medium Priority
- [ ] Font embedding support
- [ ] Vector graphics extraction
- [ ] Form field conversion
- [ ] Hyperlink preservation
- [ ] Bookmarks/outline conversion

### Low Priority
- [ ] OCR integration for scanned PDFs
- [ ] CBOR format support (alternative to JSON)
- [ ] Digital signature preservation
- [ ] Annotation conversion
- [ ] 3D content support

## Related Files

- `cmd/liv-converter/main.go` - CLI entry point
- `internal/converter/converter.go` - Main logic
- `internal/converter/pdf/parser.go` - PDF parsing
- `internal/converter/liv/builder.go` - LIV construction
- `internal/converter/liv/packager.go` - ZIP packaging
- `desktop/src/go-backend.js` - Electron bridge
- `desktop/src/liv-editor-clean.js` - UI implementation

## Support

For issues or questions:
- Check logs in `%APPDATA%/liv-editor/logs/`
- Run converter with `--verbose` flag
- Open issue on GitHub with PDF sample
