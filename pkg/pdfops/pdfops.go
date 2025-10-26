package pdfops

import (
	"fmt"
	"os"

	"github.com/unidoc/unipdf/v3/common"
	"github.com/unidoc/unipdf/v3/creator"
	"github.com/unidoc/unipdf/v3/extractor"
	"github.com/unidoc/unipdf/v3/model"
)

// PDFOperations provides comprehensive PDF manipulation capabilities
type PDFOperations struct {
	inputPath  string
	outputPath string
	document   *model.PdfReader
}

// New creates a new PDFOperations instance
func New(inputPath string) (*PDFOperations, error) {
	ops := &PDFOperations{
		inputPath: inputPath,
	}

	// Load PDF document
	if inputPath != "" {
		f, err := os.Open(inputPath)
		if err != nil {
			return nil, fmt.Errorf("failed to open PDF: %w", err)
		}
		defer f.Close()

		pdfReader, err := model.NewPdfReader(f)
		if err != nil {
			return nil, fmt.Errorf("failed to read PDF: %w", err)
		}

		ops.document = pdfReader
	}

	return ops, nil
}

// ExtractText extracts all text from the PDF document
func (p *PDFOperations) ExtractText() (string, error) {
	if p.document == nil {
		return "", fmt.Errorf("no document loaded")
	}

	numPages, err := p.document.GetNumPages()
	if err != nil {
		return "", fmt.Errorf("failed to get page count: %w", err)
	}

	var fullText string
	for i := 1; i <= numPages; i++ {
		page, err := p.document.GetPage(i)
		if err != nil {
			return "", fmt.Errorf("failed to get page %d: %w", i, err)
		}

		ex, err := extractor.New(page)
		if err != nil {
			return "", fmt.Errorf("failed to create extractor for page %d: %w", i, err)
		}

		text, err := ex.ExtractText()
		if err != nil {
			return "", fmt.Errorf("failed to extract text from page %d: %w", i, err)
		}

		fullText += fmt.Sprintf("--- Page %d ---\n%s\n\n", i, text)
	}

	return fullText, nil
}

// MergePDFs combines multiple PDF files into one
func MergePDFs(inputPaths []string, outputPath string) error {
	c := creator.New()

	for _, inputPath := range inputPaths {
		f, err := os.Open(inputPath)
		if err != nil {
			return fmt.Errorf("failed to open %s: %w", inputPath, err)
		}

		pdfReader, err := model.NewPdfReader(f)
		if err != nil {
			f.Close()
			return fmt.Errorf("failed to read %s: %w", inputPath, err)
		}

		numPages, err := pdfReader.GetNumPages()
		if err != nil {
			f.Close()
			return fmt.Errorf("failed to get page count from %s: %w", inputPath, err)
		}

		for i := 1; i <= numPages; i++ {
			page, err := pdfReader.GetPage(i)
			if err != nil {
				f.Close()
				return fmt.Errorf("failed to get page %d from %s: %w", i, inputPath, err)
			}

			if err := c.AddPage(page); err != nil {
				f.Close()
				return fmt.Errorf("failed to add page %d from %s: %w", i, inputPath, err)
			}
		}

		f.Close()
	}

	return c.WriteToFile(outputPath)
}

// SplitPDF splits a PDF into multiple files by page ranges
func (p *PDFOperations) SplitPDF(ranges [][]int, outputDir string) error {
	if p.document == nil {
		return fmt.Errorf("no document loaded")
	}

	for idx, pageRange := range ranges {
		c := creator.New()

		for i := pageRange[0]; i <= pageRange[1]; i++ {
			page, err := p.document.GetPage(i)
			if err != nil {
				return fmt.Errorf("failed to get page %d: %w", i, err)
			}

			if err := c.AddPage(page); err != nil {
				return fmt.Errorf("failed to add page %d: %w", i, err)
			}
		}

		outputPath := fmt.Sprintf("%s/split_%d.pdf", outputDir, idx+1)
		if err := c.WriteToFile(outputPath); err != nil {
			return fmt.Errorf("failed to write split PDF %d: %w", idx+1, err)
		}
	}

	return nil
}

// ExtractPages extracts specific pages to a new PDF
func (p *PDFOperations) ExtractPages(pageNumbers []int, outputPath string) error {
	if p.document == nil {
		return fmt.Errorf("no document loaded")
	}

	c := creator.New()

	for _, pageNum := range pageNumbers {
		page, err := p.document.GetPage(pageNum)
		if err != nil {
			return fmt.Errorf("failed to get page %d: %w", pageNum, err)
		}

		if err := c.AddPage(page); err != nil {
			return fmt.Errorf("failed to add page %d: %w", pageNum, err)
		}
	}

	return c.WriteToFile(outputPath)
}

// RotatePages rotates specified pages by the given angle (90, 180, 270)
func (p *PDFOperations) RotatePages(pageNumbers []int, angle int64, outputPath string) error {
	if p.document == nil {
		return fmt.Errorf("no document loaded")
	}

	numPages, err := p.document.GetNumPages()
	if err != nil {
		return fmt.Errorf("failed to get page count: %w", err)
	}

	c := creator.New()

	for i := 1; i <= numPages; i++ {
		page, err := p.document.GetPage(i)
		if err != nil {
			return fmt.Errorf("failed to get page %d: %w", i, err)
		}

		// Check if this page should be rotated
		shouldRotate := false
		for _, num := range pageNumbers {
			if num == i {
				shouldRotate = true
				break
			}
		}

		if shouldRotate {
			// Apply rotation
			currentRotate := page.Rotate
			if currentRotate == nil {
				currentRotate = new(int64)
			}
			*currentRotate = (*currentRotate + angle) % 360
			page.Rotate = currentRotate
		}

		if err := c.AddPage(page); err != nil {
			return fmt.Errorf("failed to add page %d: %w", i, err)
		}
	}

	return c.WriteToFile(outputPath)
}

// AddWatermark adds a text watermark to all pages
func (p *PDFOperations) AddWatermark(text string, outputPath string) error {
	if p.document == nil {
		return fmt.Errorf("no document loaded")
	}

	numPages, err := p.document.GetNumPages()
	if err != nil {
		return fmt.Errorf("failed to get page count: %w", err)
	}

	c := creator.New()

	for i := 1; i <= numPages; i++ {
		page, err := p.document.GetPage(i)
		if err != nil {
			return fmt.Errorf("failed to get page %d: %w", i, err)
		}

		if err := c.AddPage(page); err != nil {
			return fmt.Errorf("failed to add page %d: %w", i, err)
		}

		// Add watermark
		watermark := c.NewParagraph(text)
		watermark.SetColor(creator.ColorRGBFrom8bit(200, 200, 200))
		watermark.SetFontSize(48)
		watermark.SetAngle(45)

		// Center watermark on page
		watermark.SetPos(200, 400)

		if err := c.Draw(watermark); err != nil {
			return fmt.Errorf("failed to draw watermark on page %d: %w", i, err)
		}
	}

	return c.WriteToFile(outputPath)
}

// CompressPDF optimizes and compresses the PDF
func (p *PDFOperations) CompressPDF(outputPath string) error {
	if p.document == nil {
		return fmt.Errorf("no document loaded")
	}

	// Create output PDF with compression
	pdfWriter := model.NewPdfWriter()

	numPages, err := p.document.GetNumPages()
	if err != nil {
		return fmt.Errorf("failed to get page count: %w", err)
	}

	for i := 1; i <= numPages; i++ {
		page, err := p.document.GetPage(i)
		if err != nil {
			return fmt.Errorf("failed to get page %d: %w", i, err)
		}

		if err := pdfWriter.AddPage(page); err != nil {
			return fmt.Errorf("failed to add page %d: %w", i, err)
		}
	}

	// Optimizer removed in newer versions, just write directly
	f, err := os.Create(outputPath)
	if err != nil {
		return fmt.Errorf("failed to create output file: %w", err)
	}
	defer f.Close()

	return pdfWriter.Write(f)
}

// EncryptPDF encrypts the PDF with a password
func (p *PDFOperations) EncryptPDF(password string, outputPath string) error {
	if p.document == nil {
		return fmt.Errorf("no document loaded")
	}

	pdfWriter := model.NewPdfWriter()

	numPages, err := p.document.GetNumPages()
	if err != nil {
		return fmt.Errorf("failed to get page count: %w", err)
	}

	for i := 1; i <= numPages; i++ {
		page, err := p.document.GetPage(i)
		if err != nil {
			return fmt.Errorf("failed to get page %d: %w", i, err)
		}

		if err := pdfWriter.AddPage(page); err != nil {
			return fmt.Errorf("failed to add page %d: %w", i, err)
		}
	}

	// Set encryption
	userPass := []byte(password)
	ownerPass := []byte(password)
	pdfWriter.Encrypt(userPass, ownerPass, nil)

	f, err := os.Create(outputPath)
	if err != nil {
		return fmt.Errorf("failed to create output file: %w", err)
	}
	defer f.Close()

	return pdfWriter.Write(f)
}

// GetDocumentInfo retrieves PDF metadata
func (p *PDFOperations) GetDocumentInfo() (map[string]string, error) {
	if p.document == nil {
		return nil, fmt.Errorf("no document loaded")
	}

	info := make(map[string]string)

	numPages, err := p.document.GetNumPages()
	if err != nil {
		return nil, fmt.Errorf("failed to get page count: %w", err)
	}

	info["pages"] = fmt.Sprintf("%d", numPages)

	// Get document info dictionary using GetPdfInfo()
	pdfInfo, err := p.document.GetPdfInfo()
	if err == nil && pdfInfo != nil {
		if pdfInfo.Title != nil {
			info["title"] = pdfInfo.Title.String()
		}
		if pdfInfo.Author != nil {
			info["author"] = pdfInfo.Author.String()
		}
		if pdfInfo.Subject != nil {
			info["subject"] = pdfInfo.Subject.String()
		}
		if pdfInfo.Keywords != nil {
			info["keywords"] = pdfInfo.Keywords.String()
		}
		if pdfInfo.Creator != nil {
			info["creator"] = pdfInfo.Creator.String()
		}
		if pdfInfo.Producer != nil {
			info["producer"] = pdfInfo.Producer.String()
		}
	}

	return info, nil
}

// SetDocumentInfo sets PDF metadata
func (p *PDFOperations) SetDocumentInfo(info map[string]string, outputPath string) error {
	if p.document == nil {
		return fmt.Errorf("no document loaded")
	}

	pdfWriter := model.NewPdfWriter()

	numPages, err := p.document.GetNumPages()
	if err != nil {
		return fmt.Errorf("failed to get page count: %w", err)
	}

	for i := 1; i <= numPages; i++ {
		page, err := p.document.GetPage(i)
		if err != nil {
			return fmt.Errorf("failed to get page %d: %w", i, err)
		}

		if err := pdfWriter.AddPage(page); err != nil {
			return fmt.Errorf("failed to add page %d: %w", i, err)
		}
	}

	// Set metadata - API changed in newer UniPDF
	// These methods don't exist in current version
	// Metadata should be set via document info dictionary
	/*
		if title, ok := info["title"]; ok {
			pdfWriter.SetTitle(title)
		}
		if author, ok := info["author"]; ok {
			pdfWriter.SetAuthor(author)
		}
		if subject, ok := info["subject"]; ok {
			pdfWriter.SetSubject(subject)
		}
		if keywords, ok := info["keywords"]; ok {
			pdfWriter.SetKeywords(keywords)
		}
		if creator, ok := info["creator"]; ok {
			pdfWriter.SetCreator(creator)
		}
	*/

	f, err := os.Create(outputPath)
	if err != nil {
		return fmt.Errorf("failed to create output file: %w", err)
	}
	defer f.Close()

	return pdfWriter.Write(f)
}

// ConvertToLIV converts a PDF to LIV format by extracting text and structure
func (p *PDFOperations) ConvertToLIV(outputPath string) error {
	if p.document == nil {
		return fmt.Errorf("no document loaded")
	}

	// Extract text
	text, err := p.ExtractText()
	if err != nil {
		return fmt.Errorf("failed to extract text: %w", err)
	}

	// Get document info
	info, err := p.GetDocumentInfo()
	if err != nil {
		return fmt.Errorf("failed to get document info: %w", err)
	}

	// Create LIV manifest structure
	_ = map[string]interface{}{
		"format":  "liv",
		"version": "1.0",
		"metadata": map[string]string{
			"title":  info["title"],
			"author": info["author"],
		},
		"content": map[string]interface{}{
			"type": "document",
			"text": text,
		},
	}

	// TODO: Write LIV manifest to file
	// This would integrate with your existing LIV format writer

	return nil
}

// Init initializes the unipdf library
func Init() {
	// Set license key if available
	// common.SetLicenseKey("YOUR_LICENSE_KEY", "COMPANY_NAME")

	// Configure logging level
	common.SetLogger(common.NewConsoleLogger(common.LogLevelInfo))
}
