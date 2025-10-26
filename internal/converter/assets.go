package converter

import (
	"bytes"
	"fmt"
	"image"
	"image/jpeg"
	"image/png"

	"github.com/liv-format/liv/internal/types"
)

// ExtractAssets extracts and processes assets from parsed PDF data
func ExtractAssets(pdfData *types.PDFData, quality int) (*types.ExtractedAssets, error) {
	if pdfData == nil {
		return nil, fmt.Errorf("PDF data is nil")
	}

	assets := &types.ExtractedAssets{
		Images: []types.AssetImage{},
		Fonts:  []types.AssetFont{},
	}

	// Extract images from all pages
	for _, page := range pdfData.Pages {
		for _, img := range page.Images {
			assetImg, err := processImage(&img, quality)
			if err != nil {
				// Log warning but continue processing
				fmt.Printf("Warning: failed to process image %s: %v\n", img.ID, err)
				continue
			}
			assets.Images = append(assets.Images, *assetImg)
		}
	}

	// TODO: Extract and embed fonts from PDF
	// This requires:
	// 1. Parsing font definitions from PDF
	// 2. Extracting font data (TTF/OTF)
	// 3. Converting to web-friendly formats
	// 4. Handling font licensing issues

	return assets, nil
}

// processImage processes a PDF image for inclusion in LIV document
func processImage(pdfImg *types.PDFImage, quality int) (*types.AssetImage, error) {
	if len(pdfImg.Data) == 0 {
		return nil, fmt.Errorf("image data is empty")
	}

	// Decode image data
	img, format, err := image.Decode(bytes.NewReader(pdfImg.Data))
	if err != nil {
		return nil, fmt.Errorf("failed to decode image: %w", err)
	}

	// Determine output format
	outputFormat := pdfImg.Format
	if outputFormat == "" {
		outputFormat = format
	}

	// Re-encode with quality settings
	var outputData bytes.Buffer
	switch outputFormat {
	case "jpeg", "jpg":
		err = jpeg.Encode(&outputData, img, &jpeg.Options{Quality: quality})
	case "png":
		err = png.Encode(&outputData, img)
	default:
		// Default to JPEG for unknown formats
		err = jpeg.Encode(&outputData, img, &jpeg.Options{Quality: quality})
		outputFormat = "jpeg"
	}

	if err != nil {
		return nil, fmt.Errorf("failed to encode image: %w", err)
	}

	// Get image dimensions
	bounds := img.Bounds()

	return &types.AssetImage{
		ID:       pdfImg.ID,
		Filename: fmt.Sprintf("%s.%s", pdfImg.ID, outputFormat),
		Data:     outputData.Bytes(),
		Format:   outputFormat,
		Width:    bounds.Dx(),
		Height:   bounds.Dy(),
	}, nil
}

// OptimizeImage applies additional optimization to an image
func OptimizeImage(img *types.AssetImage, maxWidth, maxHeight int) (*types.AssetImage, error) {
	// TODO: Implement image resizing/optimization
	// 1. Resize images that exceed max dimensions
	// 2. Apply additional compression
	// 3. Convert to WebP format for better compression
	// 4. Strip unnecessary metadata

	return img, nil
}

// ExtractFonts extracts font data from PDF
func ExtractFonts(pdfData *types.PDFData) ([]types.AssetFont, error) {
	// TODO: Implement font extraction
	// This is complex and requires:
	// 1. Parse PDF font dictionaries
	// 2. Extract embedded font streams
	// 3. Convert CIDFont to TrueType/OpenType
	// 4. Handle font subsetting
	// 5. Generate web font formats (WOFF/WOFF2)

	return []types.AssetFont{}, nil
}

// EmbedFont embeds a font file into the asset collection
func EmbedFont(fontPath string) (*types.AssetFont, error) {
	// TODO: Implement font embedding from external file
	// 1. Read font file
	// 2. Parse font metadata (family, style, weight)
	// 3. Convert to web format if needed
	// 4. Generate proper filename

	return nil, fmt.Errorf("font embedding not yet implemented")
}

// AssetStats provides statistics about extracted assets
type AssetStats struct {
	TotalImages    int
	TotalImageSize int64
	TotalFonts     int
	TotalFontSize  int64
	ImageFormats   map[string]int
	FontFamilies   map[string]int
}

// GetAssetStats calculates statistics for an asset collection
func GetAssetStats(assets *types.ExtractedAssets) *AssetStats {
	stats := &AssetStats{
		ImageFormats: make(map[string]int),
		FontFamilies: make(map[string]int),
	}

	if assets == nil {
		return stats
	}

	// Count images
	stats.TotalImages = len(assets.Images)
	for _, img := range assets.Images {
		stats.TotalImageSize += int64(len(img.Data))
		stats.ImageFormats[img.Format]++
	}

	// Count fonts
	stats.TotalFonts = len(assets.Fonts)
	for _, font := range assets.Fonts {
		stats.TotalFontSize += int64(len(font.Data))
		stats.FontFamilies[font.Family]++
	}

	return stats
}
