package liv

import (
	"archive/zip"
	"encoding/json"
	"fmt"
	"io"
	"os"
	"path/filepath"

	"github.com/liv-format/liv/internal/types"
)

// PackageLIV creates a .liv file from the document, manifest, and assets
func PackageLIV(outputPath string, doc *types.LIVDocument, manifest *types.LIVManifest, assets *types.ExtractedAssets, compress bool) error {
	// Create output file
	outFile, err := os.Create(outputPath)
	if err != nil {
		return fmt.Errorf("failed to create output file: %w", err)
	}
	defer outFile.Close()

	// Create ZIP writer
	zipWriter := zip.NewWriter(outFile)
	defer zipWriter.Close()

	// Set compression level
	if !compress {
		zipWriter.RegisterCompressor(zip.Deflate, func(out io.Writer) (io.WriteCloser, error) {
			return &noCompressionWriter{w: out}, nil
		})
	}

	// Write document.json
	if err := writeJSON(zipWriter, "document.json", doc); err != nil {
		return fmt.Errorf("failed to write document.json: %w", err)
	}

	// Write manifest.json
	if err := writeJSON(zipWriter, "manifest.json", manifest); err != nil {
		return fmt.Errorf("failed to write manifest.json: %w", err)
	}

	// Write image assets
	if assets != nil && len(assets.Images) > 0 {
		for _, img := range assets.Images {
			assetPath := filepath.Join("assets", "images", img.Filename)
			if err := writeAsset(zipWriter, assetPath, img.Data); err != nil {
				return fmt.Errorf("failed to write image asset %s: %w", img.Filename, err)
			}
		}
	}

	// Write font assets
	if assets != nil && len(assets.Fonts) > 0 {
		for _, font := range assets.Fonts {
			assetPath := filepath.Join("assets", "fonts", font.Filename)
			if err := writeAsset(zipWriter, assetPath, font.Data); err != nil {
				return fmt.Errorf("failed to write font asset %s: %w", font.Filename, err)
			}
		}
	}

	// TODO: Write style assets (CSS files)
	// TODO: Write script assets (JS files) if permissions allow
	// TODO: Implement CBOR format as alternative to JSON
	// TODO: Add digital signature support
	// TODO: Add encryption support

	return nil
}

// writeJSON writes a JSON file to the ZIP archive
func writeJSON(zipWriter *zip.Writer, filename string, data interface{}) error {
	// Create file in ZIP
	writer, err := zipWriter.Create(filename)
	if err != nil {
		return fmt.Errorf("failed to create file in ZIP: %w", err)
	}

	// Encode JSON with pretty formatting
	encoder := json.NewEncoder(writer)
	encoder.SetIndent("", "  ")
	if err := encoder.Encode(data); err != nil {
		return fmt.Errorf("failed to encode JSON: %w", err)
	}

	return nil
}

// writeAsset writes a binary asset to the ZIP archive
func writeAsset(zipWriter *zip.Writer, filename string, data []byte) error {
	writer, err := zipWriter.Create(filename)
	if err != nil {
		return fmt.Errorf("failed to create file in ZIP: %w", err)
	}

	if _, err := writer.Write(data); err != nil {
		return fmt.Errorf("failed to write asset data: %w", err)
	}

	return nil
}

// noCompressionWriter is a writer that doesn't compress (for --compress=false)
type noCompressionWriter struct {
	w io.Writer
}

func (w *noCompressionWriter) Write(p []byte) (int, error) {
	return w.w.Write(p)
}

func (w *noCompressionWriter) Close() error {
	return nil
}

// UnpackageLIV extracts a .liv file for inspection
func UnpackageLIV(livPath, outputDir string) error {
	// Open LIV file
	reader, err := zip.OpenReader(livPath)
	if err != nil {
		return fmt.Errorf("failed to open LIV file: %w", err)
	}
	defer reader.Close()

	// Create output directory
	if err := os.MkdirAll(outputDir, 0755); err != nil {
		return fmt.Errorf("failed to create output directory: %w", err)
	}

	// Extract all files
	for _, file := range reader.File {
		if err := extractFile(file, outputDir); err != nil {
			return fmt.Errorf("failed to extract %s: %w", file.Name, err)
		}
	}

	return nil
}

// extractFile extracts a single file from the ZIP archive
func extractFile(file *zip.File, destDir string) error {
	// Build destination path
	destPath := filepath.Join(destDir, file.Name)

	// Create parent directories
	if file.FileInfo().IsDir() {
		return os.MkdirAll(destPath, file.Mode())
	}

	if err := os.MkdirAll(filepath.Dir(destPath), 0755); err != nil {
		return err
	}

	// Open source file
	srcFile, err := file.Open()
	if err != nil {
		return err
	}
	defer srcFile.Close()

	// Create destination file
	destFile, err := os.Create(destPath)
	if err != nil {
		return err
	}
	defer destFile.Close()

	// Copy data
	if _, err := io.Copy(destFile, srcFile); err != nil {
		return err
	}

	return nil
}

// ReadLIVDocument reads and parses document.json from a .liv file
func ReadLIVDocument(livPath string) (*types.LIVDocument, error) {
	reader, err := zip.OpenReader(livPath)
	if err != nil {
		return nil, fmt.Errorf("failed to open LIV file: %w", err)
	}
	defer reader.Close()

	// Find document.json
	for _, file := range reader.File {
		if file.Name == "document.json" {
			rc, err := file.Open()
			if err != nil {
				return nil, fmt.Errorf("failed to open document.json: %w", err)
			}
			defer rc.Close()

			var doc types.LIVDocument
			if err := json.NewDecoder(rc).Decode(&doc); err != nil {
				return nil, fmt.Errorf("failed to decode document.json: %w", err)
			}

			return &doc, nil
		}
	}

	return nil, fmt.Errorf("document.json not found in LIV file")
}

// ReadLIVManifest reads and parses manifest.json from a .liv file
func ReadLIVManifest(livPath string) (*types.LIVManifest, error) {
	reader, err := zip.OpenReader(livPath)
	if err != nil {
		return nil, fmt.Errorf("failed to open LIV file: %w", err)
	}
	defer reader.Close()

	// Find manifest.json
	for _, file := range reader.File {
		if file.Name == "manifest.json" {
			rc, err := file.Open()
			if err != nil {
				return nil, fmt.Errorf("failed to open manifest.json: %w", err)
			}
			defer rc.Close()

			var manifest types.LIVManifest
			if err := json.NewDecoder(rc).Decode(&manifest); err != nil {
				return nil, fmt.Errorf("failed to decode manifest.json: %w", err)
			}

			return &manifest, nil
		}
	}

	return nil, fmt.Errorf("manifest.json not found in LIV file")
}
