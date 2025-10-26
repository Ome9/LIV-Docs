package liv

import (
	"fmt"
	"time"

	"github.com/liv-format/liv/internal/types"
)

// GenerateManifest creates a LIV manifest from PDF data and conversion options
func GenerateManifest(pdfData *types.PDFData, doc *types.LIVDocument, sourcePath string, compress bool) (*types.LIVManifest, error) {
	if pdfData == nil {
		return nil, fmt.Errorf("PDF data is nil")
	}
	if doc == nil {
		return nil, fmt.Errorf("LIV document is nil")
	}

	manifest := &types.LIVManifest{
		Version: "1.0",
		Format:  "liv",
		Metadata: types.ManifestMetadata{
			Title:       pdfData.Metadata.Title,
			Author:      pdfData.Metadata.Author,
			Subject:     pdfData.Metadata.Subject,
			Keywords:    pdfData.Metadata.Keywords,
			Creator:     pdfData.Metadata.Creator,
			Producer:    "LIV Converter v1.0",
			CreatedAt:   pdfData.Metadata.CreatedAt,
			ModifiedAt:  pdfData.Metadata.ModifiedAt,
			GeneratedAt: time.Now().Format(time.RFC3339),
		},
		Permissions: types.ManifestPermissions{
			AllowScripts:     false, // Safe default - scripts disabled
			AllowExternalNet: false, // Safe default - no external network
			AllowPrint:       true,
			AllowCopy:        true,
			AllowModify:      false,
		},
		Pages:       len(doc.Pages),
		Compression: compress,
		Source: types.ManifestSource{
			Type:     "pdf",
			Original: sourcePath,
		},
	}

	// Collect asset references from document
	imageAssets := []string{}
	fontAssets := []string{}
	styleAssets := []string{}

	// Scan all pages for assets
	for _, page := range doc.Pages {
		for _, element := range page.Elements {
			if element.Type == "image" {
				if assetID, ok := element.Properties["asset_id"].(string); ok {
					imageAssets = append(imageAssets, assetID)
				}
			}
		}
	}

	// TODO: Scan for font assets when font embedding is implemented
	// TODO: Scan for style assets when CSS injection is implemented

	manifest.Assets = types.ManifestAssets{
		Images: imageAssets,
		Fonts:  fontAssets,
		Styles: styleAssets,
	}

	return manifest, nil
}

// ValidateManifest validates a LIV manifest structure
func ValidateManifest(manifest *types.LIVManifest) error {
	if manifest == nil {
		return fmt.Errorf("manifest is nil")
	}

	if manifest.Version == "" {
		return fmt.Errorf("missing version")
	}

	if manifest.Format != "liv" {
		return fmt.Errorf("invalid format: %s", manifest.Format)
	}

	if manifest.Pages <= 0 {
		return fmt.Errorf("invalid page count: %d", manifest.Pages)
	}

	// Validate metadata
	if manifest.Metadata.Title == "" {
		return fmt.Errorf("missing title")
	}

	if manifest.Metadata.GeneratedAt == "" {
		return fmt.Errorf("missing generation timestamp")
	}

	// Validate source info
	if manifest.Source.Type == "" {
		return fmt.Errorf("missing source type")
	}

	return nil
}

// UpdateManifestMetadata updates manifest metadata with provided values
func UpdateManifestMetadata(manifest *types.LIVManifest, title, author string) {
	if title != "" {
		manifest.Metadata.Title = title
	}
	if author != "" {
		manifest.Metadata.Author = author
	}
}
