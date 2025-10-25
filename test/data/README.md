# Test Data

This directory contains test documents and sample files used across the test suite.

## Structure

- `documents/` - Sample LIV documents for testing
- `assets/` - Test assets (images, fonts, etc.)
- `manifests/` - Sample manifest files
- `html/` - HTML content samples
- `css/` - CSS stylesheet samples
- `wasm/` - WASM module samples for testing
- `invalid/` - Invalid files for negative testing

## Usage

Test files in this directory are used by:
- Unit tests for validation
- Integration tests for real-world scenarios
- Performance tests for benchmarking
- Security tests for vulnerability testing

## File Naming Convention

- `valid-*` - Files that should pass validation
- `invalid-*` - Files that should fail validation
- `large-*` - Files for performance testing
- `malicious-*` - Files for security testing
- `sample-*` - General purpose sample files

## Maintenance

When adding new test files:
1. Follow the naming convention
2. Add a brief description in this README
3. Ensure files are referenced in appropriate tests
4. Keep file sizes reasonable (use generators for large files)

## Test Files

### Documents
- `valid-simple.liv` - Basic valid document
- `valid-complex.liv` - Complex document with multiple assets
- `invalid-corrupted.liv` - Corrupted ZIP structure
- `invalid-missing-manifest.liv` - Missing manifest file

### Assets
- `test-image.png` - Small test image (1x1 pixel)
- `test-font.ttf` - Minimal test font
- `large-image.jpg` - Large image for performance testing

### HTML Content
- `simple.html` - Basic HTML structure
- `complex.html` - Complex HTML with various elements
- `malicious.html` - HTML with potential security issues

### CSS Stylesheets
- `basic.css` - Simple stylesheet
- `complex.css` - Complex stylesheet with animations
- `malicious.css` - CSS with potential security issues

### WASM Modules
- `simple.wasm` - Basic WASM module
- `chart.wasm` - Chart rendering module
- `invalid.wasm` - Invalid WASM binary

### Manifests
- `valid-manifest.json` - Valid manifest structure
- `invalid-manifest.json` - Invalid manifest structure
- `complex-manifest.json` - Complex manifest with many resources