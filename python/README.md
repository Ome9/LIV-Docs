# LIV Document Format Python SDK

A Python SDK for automating LIV document creation, validation, and batch processing. This SDK provides a high-level Python interface to the LIV Document Format CLI tools and core functionality.

## Features

- **Document Creation**: Create LIV documents programmatically from Python
- **Batch Processing**: Process multiple documents efficiently
- **CLI Integration**: Seamless integration with existing Go CLI tools
- **Validation**: Comprehensive document validation and error reporting
- **Conversion**: Convert between LIV and other formats (PDF, HTML, Markdown, EPUB)
- **Security**: Digital signing and verification capabilities
- **Asset Management**: Handle images, fonts, and other assets
- **WASM Integration**: Support for interactive WASM modules

## Installation

### From PyPI (when available)

```bash
pip install liv-document-format
```

### From Source

```bash
git clone https://github.com/liv-document-format/liv-python
cd liv-python
pip install -e .
```

### Development Installation

```bash
pip install -e ".[dev,async,validation]"
```

## Quick Start

### Basic Document Creation

```python
from liv import LIVDocument, LIVBuilder

# Create a simple document
builder = LIVBuilder()
builder.set_metadata(
    title="My Document",
    author="John Doe",
    description="A sample LIV document"
)
builder.set_content(
    html="<h1>Hello World</h1><p>This is my first LIV document.</p>",
    css="h1 { color: blue; } p { font-size: 16px; }"
)

# Build and save the document
document = builder.build()
document.save("my_document.liv")
```

### Batch Processing

```python
from liv import LIVBatchProcessor
from pathlib import Path

# Process multiple HTML files into LIV documents
processor = LIVBatchProcessor()
html_files = Path("input").glob("*.html")

for html_file in html_files:
    result = processor.convert_html_to_liv(
        html_file,
        output_dir="output",
        metadata={
            "author": "Batch Processor",
            "title": html_file.stem
        }
    )
    print(f"Converted {html_file} -> {result.output_path}")
```

### Document Validation

```python
from liv import LIVValidator

validator = LIVValidator()
result = validator.validate_file("document.liv")

if result.is_valid:
    print("Document is valid!")
else:
    print("Validation errors:")
    for error in result.errors:
        print(f"  - {error}")
```

### Format Conversion

```python
from liv import LIVConverter

converter = LIVConverter()

# Convert LIV to PDF
converter.liv_to_pdf("document.liv", "output.pdf")

# Convert HTML to LIV
converter.html_to_liv("input.html", "output.liv")

# Convert LIV to EPUB
converter.liv_to_epub("document.liv", "output.epub")
```

## Advanced Usage

### Custom Security Policies

```python
from liv import LIVBuilder, SecurityPolicy, WASMPermissions

# Create document with custom security policy
builder = LIVBuilder()
security_policy = SecurityPolicy(
    wasm_permissions=WASMPermissions(
        memory_limit=64 * 1024 * 1024,  # 64MB
        allow_networking=False,
        allow_file_system=False
    ),
    js_execution_mode="sandboxed"
)

builder.set_security_policy(security_policy)
document = builder.build()
```

### Asset Management

```python
from liv import LIVBuilder

builder = LIVBuilder()

# Add various assets
builder.add_image("logo.png", "assets/logo.png")
builder.add_font("custom-font.woff2", "fonts/custom.woff2")
builder.add_data("config.json", {"theme": "dark", "version": "1.0"})

# Add WASM module
builder.add_wasm_module(
    name="chart-engine",
    path="modules/chart.wasm",
    permissions={"memory_limit": 32 * 1024 * 1024}
)

document = builder.build()
```

### Async Operations

```python
import asyncio
from liv import AsyncLIVProcessor

async def process_documents():
    processor = AsyncLIVProcessor()
    
    # Process multiple documents concurrently
    tasks = [
        processor.convert_html_to_liv_async("doc1.html", "doc1.liv"),
        processor.convert_html_to_liv_async("doc2.html", "doc2.liv"),
        processor.convert_html_to_liv_async("doc3.html", "doc3.liv"),
    ]
    
    results = await asyncio.gather(*tasks)
    return results

# Run async processing
results = asyncio.run(process_documents())
```

## CLI Integration

The Python SDK seamlessly integrates with the existing Go CLI tools:

```python
from liv import CLIInterface

cli = CLIInterface()

# Use CLI commands directly
result = cli.build(
    source_dir="src/",
    output="document.liv",
    manifest="manifest.json"
)

# Validate using CLI
validation = cli.validate("document.liv")

# Convert using CLI
cli.convert("document.liv", "output.pdf", format="pdf")
```

## Configuration

### Environment Variables

- `LIV_CLI_PATH`: Path to the LIV CLI executable (default: searches PATH)
- `LIV_TEMP_DIR`: Temporary directory for processing (default: system temp)
- `LIV_LOG_LEVEL`: Logging level (DEBUG, INFO, WARNING, ERROR)

### Configuration File

Create a `liv.config.json` file:

```json
{
  "cli_path": "/usr/local/bin/liv",
  "temp_dir": "/tmp/liv-processing",
  "default_security_policy": {
    "wasm_memory_limit": 67108864,
    "allow_networking": false
  },
  "batch_processing": {
    "max_concurrent": 4,
    "timeout": 300
  }
}
```

## API Reference

### Core Classes

- `LIVDocument`: Represents a LIV document
- `LIVBuilder`: Builder pattern for creating documents
- `LIVValidator`: Document validation functionality
- `LIVConverter`: Format conversion utilities
- `LIVBatchProcessor`: Batch processing operations

### Data Models

- `DocumentMetadata`: Document metadata structure
- `SecurityPolicy`: Security policy configuration
- `WASMPermissions`: WASM module permissions
- `ValidationResult`: Validation result data

### Utilities

- `CLIInterface`: Direct CLI tool integration
- `AssetManager`: Asset handling utilities
- `ConfigManager`: Configuration management

## Error Handling

```python
from liv import LIVError, ValidationError, ConversionError

try:
    document = LIVDocument.load("document.liv")
except ValidationError as e:
    print(f"Validation failed: {e}")
    for error in e.errors:
        print(f"  - {error}")
except LIVError as e:
    print(f"LIV error: {e}")
except Exception as e:
    print(f"Unexpected error: {e}")
```

## Testing

Run the test suite:

```bash
# Install development dependencies
pip install -e ".[dev]"

# Run tests
pytest

# Run tests with coverage
pytest --cov=liv --cov-report=html
```

## Contributing

1. Fork the repository
2. Create a feature branch
3. Make your changes
4. Add tests for new functionality
5. Run the test suite
6. Submit a pull request

## License

This project is licensed under the MIT License - see the LICENSE file for details.

## Support

- Documentation: https://liv-python.readthedocs.io/
- Issues: https://github.com/liv-document-format/liv-python/issues
- Discussions: https://github.com/liv-document-format/liv-python/discussions