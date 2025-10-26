package liv

import (
	"fmt"
	"strings"

	"github.com/liv-format/liv/internal/types"
)

// BuildLIVDocument converts parsed PDF data into a LIV document structure
func BuildLIVDocument(pdfData *types.PDFData) (*types.LIVDocument, error) {
	if pdfData == nil {
		return nil, fmt.Errorf("PDF data is nil")
	}

	doc := &types.LIVDocument{
		Version: "1.0",
		Format:  "liv",
		Pages:   make([]types.LIVPage, 0, len(pdfData.Pages)),
		Styles:  make(map[string]any),
	}

	// Convert each PDF page to a LIV page
	for i, pdfPage := range pdfData.Pages {
		livPage, err := convertPage(&pdfPage, i+1)
		if err != nil {
			return nil, fmt.Errorf("failed to convert page %d: %w", i+1, err)
		}
		doc.Pages = append(doc.Pages, *livPage)
	}

	// Add default styles
	doc.Styles["body"] = map[string]any{
		"font-family": "Arial, sans-serif",
		"font-size":   "12pt",
		"color":       "#000000",
	}

	// TODO: Add CSS stylesheets for better formatting
	// TODO: Add JavaScript components for interactivity
	// TODO: Implement advanced layout algorithms for better text flow

	return doc, nil
}

// convertPage converts a PDF page to a LIV page
func convertPage(pdfPage *types.PDFPage, pageNum int) (*types.LIVPage, error) {
	livPage := &types.LIVPage{
		ID:       fmt.Sprintf("page-%d", pageNum),
		Number:   pageNum,
		Width:    pdfPage.Width,
		Height:   pdfPage.Height,
		Elements: make([]types.LIVElement, 0),
	}

	elementID := 0

	// Convert text blocks to LIV elements
	for _, textBlock := range pdfPage.TextBlocks {
		elementID++
		element := convertTextBlock(&textBlock, elementID)
		livPage.Elements = append(livPage.Elements, element)
	}

	// Convert images to LIV elements
	for _, image := range pdfPage.Images {
		elementID++
		element := convertImage(&image, elementID)
		livPage.Elements = append(livPage.Elements, element)
	}

	// Convert graphics to LIV elements
	for _, graphic := range pdfPage.Graphics {
		elementID++
		element := convertGraphic(&graphic, elementID)
		livPage.Elements = append(livPage.Elements, element)
	}

	// TODO: Implement smart layout detection (columns, tables, etc.)
	// TODO: Group related elements into containers
	// TODO: Detect and preserve reading order

	return livPage, nil
}

// convertTextBlock converts a PDF text block to a LIV element
func convertTextBlock(textBlock *types.PDFTextBlock, id int) types.LIVElement {
	element := types.LIVElement{
		ID:      fmt.Sprintf("text-%d", id),
		Type:    "text",
		Content: textBlock.Text,
		Position: types.ElementPos{
			X:      textBlock.X,
			Y:      textBlock.Y,
			Width:  textBlock.Width,
			Height: textBlock.Height,
		},
		Style: types.ElementStyle{
			FontFamily: normalizeFontFamily(textBlock.FontName),
			FontSize:   textBlock.FontSize,
			Color:      textBlock.Color,
		},
	}

	// Apply font weight
	if textBlock.Bold {
		element.Style.FontWeight = "bold"
	}

	// Apply font style
	if textBlock.Italic {
		element.Style.FontStyle = "italic"
	}

	return element
}

// convertImage converts a PDF image to a LIV element
func convertImage(image *types.PDFImage, id int) types.LIVElement {
	return types.LIVElement{
		ID:   fmt.Sprintf("image-%d", id),
		Type: "image",
		Position: types.ElementPos{
			X:      image.X,
			Y:      image.Y,
			Width:  image.Width,
			Height: image.Height,
		},
		Properties: map[string]any{
			"asset_id": image.ID,
			"format":   image.Format,
			"dpi":      image.DPI,
		},
	}
}

// convertGraphic converts a PDF graphic to a LIV element
func convertGraphic(graphic *types.PDFGraphic, id int) types.LIVElement {
	return types.LIVElement{
		ID:   fmt.Sprintf("shape-%d", id),
		Type: "shape",
		Position: types.ElementPos{
			X:      graphic.X,
			Y:      graphic.Y,
			Width:  graphic.Width,
			Height: graphic.Height,
		},
		Style: types.ElementStyle{
			Color: graphic.Color,
		},
		Properties: map[string]any{
			"shape_type":   graphic.Type,
			"stroke_width": graphic.StrokeWidth,
			"path":         graphic.Path,
		},
	}
}

// normalizeFontFamily converts PDF font names to standard web fonts
func normalizeFontFamily(pdfFont string) string {
	pdfFont = strings.ToLower(pdfFont)

	// Map common PDF fonts to web-safe fonts
	fontMap := map[string]string{
		"times":     "Times New Roman, serif",
		"helvetica": "Arial, sans-serif",
		"courier":   "Courier New, monospace",
		"arial":     "Arial, sans-serif",
	}

	for key, value := range fontMap {
		if strings.Contains(pdfFont, key) {
			return value
		}
	}

	// Default fallback
	return "Arial, sans-serif"
}

// ValidateLIVDocument validates a LIV document structure
func ValidateLIVDocument(doc *types.LIVDocument) error {
	if doc == nil {
		return fmt.Errorf("document is nil")
	}

	if doc.Version == "" {
		return fmt.Errorf("missing version")
	}

	if doc.Format != "liv" {
		return fmt.Errorf("invalid format: %s", doc.Format)
	}

	if len(doc.Pages) == 0 {
		return fmt.Errorf("document has no pages")
	}

	// Validate each page
	for i, page := range doc.Pages {
		if err := validatePage(&page, i+1); err != nil {
			return fmt.Errorf("invalid page %d: %w", i+1, err)
		}
	}

	return nil
}

// validatePage validates a LIV page
func validatePage(page *types.LIVPage, pageNum int) error {
	if page.ID == "" {
		return fmt.Errorf("missing page ID")
	}

	if page.Width <= 0 || page.Height <= 0 {
		return fmt.Errorf("invalid dimensions: %fx%f", page.Width, page.Height)
	}

	// Validate elements
	for i, element := range page.Elements {
		if element.ID == "" {
			return fmt.Errorf("element %d missing ID", i)
		}
		if element.Type == "" {
			return fmt.Errorf("element %s missing type", element.ID)
		}
	}

	return nil
}
