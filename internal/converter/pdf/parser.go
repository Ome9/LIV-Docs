package pdf

import (
	"fmt"
	"os"
	"path/filepath"
	"strings"

	"github.com/liv-format/liv/internal/types"
	"rsc.io/pdf"
)

// ParsePDF parses a PDF file and extracts all content
func ParsePDF(pdfPath string) (*types.PDFData, error) {
	// Open the PDF file
	file, err := os.Open(pdfPath)
	if err != nil {
		return nil, fmt.Errorf("failed to open PDF: %w", err)
	}
	defer file.Close()

	// Get file info for size
	fileInfo, err := file.Stat()
	if err != nil {
		return nil, fmt.Errorf("failed to get file info: %w", err)
	}

	// Create PDF reader using rsc.io/pdf
	pdfReader, err := pdf.NewReader(file, fileInfo.Size())
	if err != nil {
		return nil, fmt.Errorf("failed to create PDF reader: %w", err)
	}

	// Extract metadata
	metadata := extractMetadata(pdfReader, pdfPath)

	// Extract pages
	numPages := pdfReader.NumPage()
	pages := make([]types.PDFPage, 0, numPages)

	for i := 1; i <= numPages; i++ {
		page, err := extractPage(pdfReader, i)
		if err != nil {
			return nil, fmt.Errorf("failed to extract page %d: %w", i, err)
		}
		pages = append(pages, *page)
	}

	return &types.PDFData{
		Metadata: *metadata,
		Pages:    pages,
	}, nil
}

// extractMetadata extracts PDF document metadata
func extractMetadata(reader *pdf.Reader, pdfPath string) *types.PDFMetadata {
	metadata := &types.PDFMetadata{}

	// Get info dictionary
	info := reader.Trailer().Key("Info")

	// Try to get title
	if info.Kind() == pdf.Dict {
		if title := info.Key("Title"); title.Kind() == pdf.String {
			metadata.Title = title.Text()
		}

		// Try to get author
		if author := info.Key("Author"); author.Kind() == pdf.String {
			metadata.Author = author.Text()
		}

		// Try to get subject
		if subject := info.Key("Subject"); subject.Kind() == pdf.String {
			metadata.Subject = subject.Text()
		}

		// Try to get keywords
		if keywords := info.Key("Keywords"); keywords.Kind() == pdf.String {
			metadata.Keywords = keywords.Text()
		}

		// Try to get creator
		if creator := info.Key("Creator"); creator.Kind() == pdf.String {
			metadata.Creator = creator.Text()
		}

		// Try to get producer
		if producer := info.Key("Producer"); producer.Kind() == pdf.String {
			metadata.Producer = producer.Text()
		}
	}

	// If no title, use filename
	if metadata.Title == "" {
		metadata.Title = strings.TrimSuffix(filepath.Base(pdfPath), ".pdf")
	}

	return metadata
}

// extractPage extracts content from a single PDF page
func extractPage(reader *pdf.Reader, pageNum int) (*types.PDFPage, error) {
	page := reader.Page(pageNum)
	if page.V.IsNull() {
		return nil, fmt.Errorf("page %d not found", pageNum)
	}

	// Get page dimensions
	mediaBox := page.V.Key("MediaBox")
	width := 612.0  // Default letter size width
	height := 792.0 // Default letter size height

	if mediaBox.Kind() == pdf.Array {
		if mediaBox.Len() >= 4 {
			width = mediaBox.Index(2).Float64()
			height = mediaBox.Index(3).Float64()
		}
	}

	// Get rotation
	rotation := 0
	if rotate := page.V.Key("Rotate"); rotate.Kind() == pdf.Integer {
		rotation = int(rotate.Int64())
	}

	// Extract text from the page
	content := page.Content()

	// Smart text extraction: use X position to detect word boundaries
	textBlocks := []types.PDFTextBlock{}

	if len(content.Text) > 0 {
		lastY := -1.0
		lastX := -1.0
		var lines []string
		var currentLine strings.Builder

		for _, textObj := range content.Text {
			// Detect new line by Y position change
			if lastY >= 0 && textObj.Y != lastY {
				// Y changed, new line
				if currentLine.Len() > 0 {
					lines = append(lines, currentLine.String())
					currentLine.Reset()
				}
				lastX = -1.0
			}

			// Check if we need a space based on X position gap
			if lastX >= 0 && lastY == textObj.Y {
				xGap := textObj.X - lastX
				// If gap is larger than typical character width (> 3 units), add space
				if xGap > 3 {
					currentLine.WriteString(" ")
				}
			}

			lastY = textObj.Y
			lastX = textObj.X + textObj.W // End position of this text

			// Add the text content, removing any existing spaces
			currentLine.WriteString(strings.TrimSpace(textObj.S))
		}

		// Add last line
		if currentLine.Len() > 0 {
			lines = append(lines, currentLine.String())
		}

		// Create text blocks from lines
		yPos := 50.0
		lineHeight := 20.0

		for _, line := range lines {
			line = strings.TrimSpace(line)

			if line != "" {
				textBlocks = append(textBlocks, types.PDFTextBlock{
					Text:     line,
					X:        50,
					Y:        yPos,
					Width:    width - 100,
					Height:   12,
					FontName: "Default",
					FontSize: 12,
					Color:    "#000000",
					Bold:     false,
					Italic:   false,
				})
				yPos += lineHeight
			}
		}
	}

	// No image extraction for now
	images := []types.PDFImage{}
	graphics := []types.PDFGraphic{}

	return &types.PDFPage{
		Number:     pageNum,
		Width:      width,
		Height:     height,
		Rotation:   rotation,
		TextBlocks: textBlocks,
		Images:     images,
		Graphics:   graphics,
	}, nil
}

// InspectPDF provides detailed information about a PDF file
func InspectPDF(pdfPath string) (map[string]interface{}, error) {
	file, err := os.Open(pdfPath)
	if err != nil {
		return nil, fmt.Errorf("failed to open PDF: %w", err)
	}
	defer file.Close()

	fileInfo, err := file.Stat()
	if err != nil {
		return nil, fmt.Errorf("failed to get file info: %w", err)
	}

	reader, err := pdf.NewReader(file, fileInfo.Size())
	if err != nil {
		return nil, fmt.Errorf("failed to create reader: %w", err)
	}

	numPages := reader.NumPage()

	info := map[string]interface{}{
		"filename":   filepath.Base(pdfPath),
		"pages":      numPages,
		"size_bytes": fileInfo.Size(),
	}

	// Get metadata
	metadata := extractMetadata(reader, pdfPath)
	if metadata.Title != "" {
		info["title"] = metadata.Title
	}
	if metadata.Author != "" {
		info["author"] = metadata.Author
	}

	return info, nil
}
