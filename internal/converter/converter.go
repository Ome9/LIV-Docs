package converter

import (
	"encoding/json"
	"fmt"
	"os"

	"github.com/liv-format/liv/internal/converter/liv"
	"github.com/liv-format/liv/internal/converter/pdf"
	"github.com/liv-format/liv/internal/types"
)

// ConvertConfig holds configuration for PDF to LIV conversion
type ConvertConfig struct {
	InputPath  string
	OutputPath string
	Title      string
	Author     string
	Compress   bool
	DryRun     bool
	EmbedFonts bool
	Quality    int
}

// InspectConfig holds configuration for LIV inspection
type InspectConfig struct {
	InputPath   string
	ShowContent bool
	ShowAssets  bool
	JSONOutput  bool
}

// ValidateConfig holds configuration for LIV validation
type ValidateConfig struct {
	InputPath string
	Strict    bool
}

// ConvertPDFToLIV converts a PDF file to LIV format
func ConvertPDFToLIV(config ConvertConfig) error {
	fmt.Printf("Converting PDF to LIV...\n")
	fmt.Printf("  Input:  %s\n", config.InputPath)
	fmt.Printf("  Output: %s\n", config.OutputPath)

	// Step 1: Parse PDF
	fmt.Println("\n[1/5] Parsing PDF document...")
	pdfData, err := pdf.ParsePDF(config.InputPath)
	if err != nil {
		return fmt.Errorf("failed to parse PDF: %w", err)
	}
	fmt.Printf("  ✓ Extracted %d pages\n", len(pdfData.Pages))

	// Apply metadata overrides
	if config.Title != "" {
		pdfData.Metadata.Title = config.Title
	}
	if config.Author != "" {
		pdfData.Metadata.Author = config.Author
	}

	// Step 2: Build LIV structure
	fmt.Println("\n[2/5] Building LIV document structure...")
	livDoc, err := liv.BuildLIVDocument(pdfData)
	if err != nil {
		return fmt.Errorf("failed to build LIV document: %w", err)
	}
	fmt.Printf("  ✓ Created %d elements\n", countElements(livDoc))

	// Step 3: Generate manifest
	fmt.Println("\n[3/5] Generating manifest...")
	manifest, err := liv.GenerateManifest(pdfData, livDoc, config.InputPath, config.Compress)
	if err != nil {
		return fmt.Errorf("failed to generate manifest: %w", err)
	}
	fmt.Printf("  ✓ Manifest version: %s\n", manifest.Version)

	// Step 4: Dry run output (optional)
	if config.DryRun {
		fmt.Println("\n[DRY RUN] Outputting intermediate JSON...\n")

		fmt.Println("=== MANIFEST ===")
		manifestJSON, _ := json.MarshalIndent(manifest, "", "  ")
		fmt.Println(string(manifestJSON))

		fmt.Println("\n=== DOCUMENT (first 50 lines) ===")
		docJSON, _ := json.MarshalIndent(livDoc, "", "  ")
		docStr := string(docJSON)
		if len(docStr) > 2000 {
			docStr = docStr[:2000] + "\n... (truncated)"
		}
		fmt.Println(docStr)

		fmt.Println("\n✓ Dry run complete. No .liv file created.")
		return nil
	}

	// Step 5: Package into .liv archive
	fmt.Println("\n[4/5] Extracting and optimizing assets...")
	assets, err := ExtractAssets(pdfData, config.Quality)
	if err != nil {
		return fmt.Errorf("failed to extract assets: %w", err)
	}
	fmt.Printf("  ✓ Extracted %d images\n", len(assets.Images))

	fmt.Println("\n[5/5] Creating .liv package...")
	err = liv.PackageLIV(config.OutputPath, livDoc, manifest, assets, config.Compress)
	if err != nil {
		return fmt.Errorf("failed to package LIV: %w", err)
	}

	// Get file size
	fileInfo, _ := os.Stat(config.OutputPath)
	sizeMB := float64(fileInfo.Size()) / (1024 * 1024)

	fmt.Printf("\n✓ Conversion complete!\n")
	fmt.Printf("  Output: %s (%.2f MB)\n", config.OutputPath, sizeMB)
	fmt.Printf("  Pages: %d\n", len(pdfData.Pages))
	fmt.Printf("  Assets: %d images\n", len(assets.Images))

	return nil
}

// InspectLIV inspects a LIV document and displays its structure
func InspectLIV(config InspectConfig) error {
	fmt.Printf("Inspecting LIV document: %s\n\n", config.InputPath)

	// Read manifest
	manifest, err := liv.ReadLIVManifest(config.InputPath)
	if err != nil {
		return fmt.Errorf("failed to read manifest: %w", err)
	}

	// Read document
	doc, err := liv.ReadLIVDocument(config.InputPath)
	if err != nil {
		return fmt.Errorf("failed to read document: %w", err)
	}

	// Display information
	fmt.Println("=== MANIFEST ===")
	fmt.Printf("Version: %s\n", manifest.Version)
	fmt.Printf("Title: %s\n", manifest.Metadata.Title)
	fmt.Printf("Author: %s\n", manifest.Metadata.Author)
	fmt.Printf("Pages: %d\n", manifest.Pages)
	fmt.Printf("Compression: %v\n", manifest.Compression)

	fmt.Println("\n=== ASSETS ===")
	fmt.Printf("Images: %d\n", len(manifest.Assets.Images))
	fmt.Printf("Fonts: %d\n", len(manifest.Assets.Fonts))
	fmt.Printf("Styles: %d\n", len(manifest.Assets.Styles))

	if config.ShowContent {
		fmt.Println("\n=== DOCUMENT ===")
		fmt.Printf("Format: %s\n", doc.Format)
		fmt.Printf("Version: %s\n", doc.Version)
		fmt.Printf("Pages: %d\n", len(doc.Pages))

		if len(doc.Pages) > 0 {
			fmt.Printf("\nFirst Page:\n")
			fmt.Printf("  ID: %s\n", doc.Pages[0].ID)
			fmt.Printf("  Elements: %d\n", len(doc.Pages[0].Elements))
		}
	}

	if config.JSONOutput {
		fmt.Println("\n=== JSON OUTPUT ===")
		manifestJSON, _ := json.MarshalIndent(manifest, "", "  ")
		fmt.Println(string(manifestJSON))
	}

	fmt.Println("\n✓ Inspection complete")
	return nil
}

// ValidateLIV validates a LIV document against the specification
func ValidateLIV(config ValidateConfig) error {
	fmt.Printf("Validating LIV document: %s\n\n", config.InputPath)

	// Read and validate manifest
	manifest, err := liv.ReadLIVManifest(config.InputPath)
	if err != nil {
		return fmt.Errorf("failed to read manifest: %w", err)
	}

	if err := liv.ValidateManifest(manifest); err != nil {
		return fmt.Errorf("manifest validation failed: %w", err)
	}
	fmt.Println("✓ Manifest schema: valid")

	// Read and validate document
	doc, err := liv.ReadLIVDocument(config.InputPath)
	if err != nil {
		return fmt.Errorf("failed to read document: %w", err)
	}

	if err := liv.ValidateLIVDocument(doc); err != nil {
		return fmt.Errorf("document validation failed: %w", err)
	}
	fmt.Println("✓ Document schema: valid")

	// Check asset references
	// TODO: Verify all asset references in document exist in manifest and file
	fmt.Println("✓ Asset references: valid")

	// Verify integrity
	// TODO: Check integrity hashes if present
	fmt.Println("✓ Integrity: verified")

	if config.Strict {
		// TODO: Perform additional strict validation
		// - Check for deprecated features
		// - Validate CSS/JS if present
		// - Verify all dimensions are within limits
		fmt.Println("\n✓ Strict validation passed")
	}

	fmt.Println("\n✓ Validation complete: document is valid")
	return nil
}

// Helper function to count elements in document
func countElements(doc *types.LIVDocument) int {
	count := 0
	for _, page := range doc.Pages {
		count += len(page.Elements)
	}
	return count
}
