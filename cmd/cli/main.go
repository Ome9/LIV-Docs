package main

import (
	"fmt"
	"os"
	"os/exec"
	"path/filepath"
	"strings"
	"time"

	"github.com/liv-format/liv/pkg/container"
	"github.com/liv-format/liv/pkg/core"
	"github.com/liv-format/liv/pkg/integrity"
	"github.com/liv-format/liv/pkg/manifest"
	"github.com/spf13/cobra"
)

var (
	version = "0.1.0"
	commit  = "dev"
	date    = "unknown"
)

func main() {
	rootCmd := &cobra.Command{
		Use:   "liv",
		Short: "LIV Format CLI - Live Interactive Visual documents",
		Long: `LIV Format CLI provides tools for creating, viewing, and converting
Live Interactive Visual documents. LIV documents combine the portability
of PDF with modern web technologies for interactive content.`,
		Version: fmt.Sprintf("%s (commit: %s, built: %s)", version, commit, date),
	}

	// Add subcommands
	rootCmd.AddCommand(buildCmd())
	rootCmd.AddCommand(viewCmd())
	rootCmd.AddCommand(convertCmd())
	rootCmd.AddCommand(validateCmd())
	rootCmd.AddCommand(signCmd())

	// Execute the root command
	if err := rootCmd.Execute(); err != nil {
		fmt.Fprintf(os.Stderr, "Error: %v\n", err)
		os.Exit(1)
	}
}

func buildCmd() *cobra.Command {
	var (
		inputDir     string
		outputFile   string
		manifestFile string
		compress     bool
		sign         bool
		keyFile      string
	)

	cmd := &cobra.Command{
		Use:   "build",
		Short: "Build a LIV document from source files",
		Long: `Build creates a LIV document package from source files and assets.
It validates the content, generates a manifest, and optionally signs the document.`,
		Example: `  liv build --input ./my-doc --output document.liv
  liv build --input ./my-doc --output document.liv --sign --key private.pem`,
		RunE: func(cmd *cobra.Command, args []string) error {
			return runBuild(inputDir, outputFile, manifestFile, compress, sign, keyFile)
		},
	}

	cmd.Flags().StringVarP(&inputDir, "input", "i", "", "Input directory containing source files (required)")
	cmd.Flags().StringVarP(&outputFile, "output", "o", "", "Output LIV file path (required)")
	cmd.Flags().StringVarP(&manifestFile, "manifest", "m", "", "Custom manifest file (optional)")
	cmd.Flags().BoolVarP(&compress, "compress", "c", true, "Compress assets")
	cmd.Flags().BoolVarP(&sign, "sign", "s", false, "Sign the document")
	cmd.Flags().StringVarP(&keyFile, "key", "k", "", "Private key file for signing")

	cmd.MarkFlagRequired("input")
	cmd.MarkFlagRequired("output")

	return cmd
}

func viewCmd() *cobra.Command {
	var (
		port     int
		web      bool
		fallback bool
	)

	cmd := &cobra.Command{
		Use:   "view [file]",
		Short: "View a LIV document",
		Long: `View opens a LIV document in the viewer. Can run as a desktop application
or web server for browser-based viewing.`,
		Example: `  liv view document.liv
  liv view document.liv --web --port 8080
  liv view document.liv --fallback`,
		Args: cobra.ExactArgs(1),
		RunE: func(cmd *cobra.Command, args []string) error {
			return runView(args[0], port, web, fallback)
		},
	}

	cmd.Flags().IntVarP(&port, "port", "p", 8080, "Port for web server mode")
	cmd.Flags().BoolVarP(&web, "web", "w", false, "Run as web server")
	cmd.Flags().BoolVarP(&fallback, "fallback", "f", false, "Use static fallback mode")

	return cmd
}

func convertCmd() *cobra.Command {
	var (
		format     string
		outputFile string
		quality    int
	)

	cmd := &cobra.Command{
		Use:   "convert [input]",
		Short: "Convert between LIV and other formats",
		Long: `Convert transforms LIV documents to other formats (PDF, HTML, Markdown, EPUB)
or imports other formats into LIV documents.`,
		Example: `  liv convert document.liv --format pdf --output document.pdf
  liv convert document.html --format liv --output document.liv
  liv convert document.liv --format html --output document.html`,
		Args: cobra.ExactArgs(1),
		RunE: func(cmd *cobra.Command, args []string) error {
			return runConvert(args[0], format, outputFile, quality)
		},
	}

	cmd.Flags().StringVarP(&format, "format", "f", "", "Target format (pdf, html, markdown, epub, liv)")
	cmd.Flags().StringVarP(&outputFile, "output", "o", "", "Output file path")
	cmd.Flags().IntVarP(&quality, "quality", "q", 90, "Quality for lossy formats (1-100)")

	cmd.MarkFlagRequired("format")
	cmd.MarkFlagRequired("output")

	return cmd
}

func validateCmd() *cobra.Command {
	var (
		checkSignatures bool
		verbose         bool
	)

	cmd := &cobra.Command{
		Use:   "validate [file]",
		Short: "Validate a LIV document",
		Long: `Validate checks a LIV document for structural integrity, security compliance,
and content validity. Reports any errors or warnings found.`,
		Example: `  liv validate document.liv
  liv validate document.liv --signatures --verbose`,
		Args: cobra.ExactArgs(1),
		RunE: func(cmd *cobra.Command, args []string) error {
			return runValidate(args[0], checkSignatures, verbose)
		},
	}

	cmd.Flags().BoolVarP(&checkSignatures, "signatures", "s", true, "Verify digital signatures")
	cmd.Flags().BoolVarP(&verbose, "verbose", "v", false, "Verbose output")

	return cmd
}

func signCmd() *cobra.Command {
	var (
		keyFile    string
		outputFile string
	)

	cmd := &cobra.Command{
		Use:   "sign [file]",
		Short: "Sign a LIV document",
		Long: `Sign adds digital signatures to a LIV document for integrity verification
and authenticity validation.`,
		Example: `  liv sign document.liv --key private.pem
  liv sign document.liv --key private.pem --output signed-document.liv`,
		Args: cobra.ExactArgs(1),
		RunE: func(cmd *cobra.Command, args []string) error {
			return runSign(args[0], keyFile, outputFile)
		},
	}

	cmd.Flags().StringVarP(&keyFile, "key", "k", "", "Private key file for signing (required)")
	cmd.Flags().StringVarP(&outputFile, "output", "o", "", "Output file (default: overwrite input)")

	cmd.MarkFlagRequired("key")

	return cmd
}

// Command implementations (stubs for now)

func runBuild(inputDir, outputFile, manifestFile string, compress, sign bool, keyFile string) error {
	fmt.Printf("Building LIV document from %s to %s\n", inputDir, outputFile)

	// Find the builder executable
	builderPath, err := findBuilderExecutable()
	if err != nil {
		return fmt.Errorf("builder not found: %v", err)
	}

	// Prepare arguments
	args := []string{
		"--input", inputDir,
		"--output", outputFile,
	}

	if manifestFile != "" {
		args = append(args, "--manifest", manifestFile)
	}

	if !compress {
		args = append(args, "--compress=false")
	}

	if sign {
		args = append(args, "--sign")
		if keyFile != "" {
			args = append(args, "--key", keyFile)
		}
	}

	args = append(args, "--verbose")

	// Execute builder
	cmd := exec.Command(builderPath, args...)
	cmd.Stdout = os.Stdout
	cmd.Stderr = os.Stderr

	return cmd.Run()
}

func runView(file string, port int, web, fallback bool) error {
	// Check if file exists
	if _, err := os.Stat(file); os.IsNotExist(err) {
		return fmt.Errorf("file not found: %s", file)
	}

	if web {
		fmt.Printf("Starting web viewer for: %s\n", file)
		fmt.Printf("Server will be available at: http://localhost:%d\n", port)

		// Find the viewer executable
		viewerPath, err := findViewerExecutable()
		if err != nil {
			return fmt.Errorf("viewer not found: %v", err)
		}

		// Prepare arguments
		args := []string{
			"--web",
			"--port", fmt.Sprintf("%d", port),
		}

		if fallback {
			args = append(args, "--fallback")
		}

		args = append(args, file)

		// Execute viewer
		cmd := exec.Command(viewerPath, args...)
		cmd.Stdout = os.Stdout
		cmd.Stderr = os.Stderr

		return cmd.Run()
	} else {
		// Desktop mode - for now, just validate and show info
		fmt.Printf("Opening LIV document: %s\n", file)

		// Validate document first
		zipContainer := container.NewZIPContainer()
		files, err := zipContainer.ExtractToMemory(file)
		if err != nil {
			return fmt.Errorf("failed to read document: %v", err)
		}

		// Parse manifest for info
		manifestData, exists := files["manifest.json"]
		if !exists {
			return fmt.Errorf("invalid LIV document: manifest.json not found")
		}

		validator := manifest.NewManifestValidator()
		parsedManifest, result := validator.ValidateManifestJSON(manifestData)
		if !result.IsValid {
			return fmt.Errorf("invalid manifest: %v", result.Errors)
		}

		// Display document info
		fmt.Printf("\nDocument Information:\n")
		fmt.Printf("  Title: %s\n", parsedManifest.Metadata.Title)
		fmt.Printf("  Author: %s\n", parsedManifest.Metadata.Author)
		fmt.Printf("  Version: %s\n", parsedManifest.Metadata.Version)
		fmt.Printf("  Created: %s\n", parsedManifest.Metadata.Created.Format("2006-01-02 15:04:05"))
		fmt.Printf("  Modified: %s\n", parsedManifest.Metadata.Modified.Format("2006-01-02 15:04:05"))
		fmt.Printf("  Resources: %d files\n", len(parsedManifest.Resources))

		if parsedManifest.Features != nil {
			fmt.Printf("  Features: ")
			features := []string{}
			if parsedManifest.Features.Animations {
				features = append(features, "animations")
			}
			if parsedManifest.Features.Interactivity {
				features = append(features, "interactivity")
			}
			if parsedManifest.Features.Charts {
				features = append(features, "charts")
			}
			if parsedManifest.Features.WebAssembly {
				features = append(features, "wasm")
			}
			fmt.Printf("%s\n", strings.Join(features, ", "))
		}

		fmt.Printf("\nNote: Desktop viewer not yet implemented. Use --web flag for web viewer.\n")
		return nil
	}
}

func runConvert(input, format, output string, quality int) error {
	fmt.Printf("Converting %s to %s format\n", input, format)

	// Check if input file exists
	if _, err := os.Stat(input); os.IsNotExist(err) {
		return fmt.Errorf("input file not found: %s", input)
	}

	switch strings.ToLower(format) {
	case "html":
		return convertToHTML(input, output)
	case "pdf":
		return convertToPDF(input, output, quality)
	case "markdown", "md":
		return convertToMarkdown(input, output)
	case "epub":
		return convertToEPUB(input, output)
	case "liv":
		return convertToLIV(input, output)
	default:
		return fmt.Errorf("unsupported format: %s", format)
	}
}

func convertToHTML(livFile, outputFile string) error {
	fmt.Printf("Extracting HTML content from LIV document...\n")

	// Extract document
	zipContainer := container.NewZIPContainer()
	files, err := zipContainer.ExtractToMemory(livFile)
	if err != nil {
		return fmt.Errorf("failed to extract LIV document: %v", err)
	}

	// Get HTML content
	htmlContent, exists := files["content/index.html"]
	if !exists {
		return fmt.Errorf("no HTML content found in document")
	}

	// Get CSS content if available
	cssContent := getFileContentSafe(files, "content/styles/main.css")

	// Create standalone HTML with embedded CSS
	html := string(htmlContent)

	if cssContent != "" {
		// Inject CSS into HTML
		styleTag := fmt.Sprintf("<style>\n%s\n</style>", cssContent)

		// Try to insert before closing </head> tag
		if headEnd := strings.Index(strings.ToLower(html), "</head>"); headEnd != -1 {
			html = html[:headEnd] + styleTag + "\n" + html[headEnd:]
		} else {
			// If no </head> tag, prepend to body
			html = styleTag + "\n" + html
		}
	}

	// Write HTML file
	err = os.WriteFile(outputFile, []byte(html), 0644)
	if err != nil {
		return fmt.Errorf("failed to write HTML file: %v", err)
	}

	fmt.Printf("✓ HTML exported to: %s\n", outputFile)
	return nil
}

func convertToMarkdown(livFile, outputFile string) error {
	fmt.Printf("Converting LIV document to Markdown...\n")

	// Extract document
	zipContainer := container.NewZIPContainer()
	files, err := zipContainer.ExtractToMemory(livFile)
	if err != nil {
		return fmt.Errorf("failed to extract LIV document: %v", err)
	}

	// Get content - prefer static fallback for Markdown conversion
	var htmlContent string
	if staticFallback := getFileContentSafe(files, "content/static/fallback.html"); staticFallback != "" {
		htmlContent = staticFallback
	} else if mainHTML, exists := files["content/index.html"]; exists {
		htmlContent = string(mainHTML)
	} else {
		return fmt.Errorf("no HTML content found in document")
	}

	// Convert HTML to Markdown
	markdown := convertHTMLToMarkdown(htmlContent)

	// Write Markdown file
	err = os.WriteFile(outputFile, []byte(markdown), 0644)
	if err != nil {
		return fmt.Errorf("failed to write Markdown file: %v", err)
	}

	fmt.Printf("✓ Markdown exported to: %s\n", outputFile)
	return nil
}

func convertToEPUB(livFile, outputFile string) error {
	fmt.Printf("Converting LIV document to EPUB...\n")

	// Extract document
	zipContainer := container.NewZIPContainer()
	files, err := zipContainer.ExtractToMemory(livFile)
	if err != nil {
		return fmt.Errorf("failed to extract LIV document: %v", err)
	}

	// Parse manifest
	manifestData, exists := files["manifest.json"]
	if !exists {
		return fmt.Errorf("no manifest found in document")
	}

	manifestParser := manifest.NewManifestParser()
	doc, err := manifestParser.ParseFromBytes(manifestData)
	if err != nil {
		return fmt.Errorf("failed to parse manifest: %v", err)
	}

	// Get content - prefer static fallback for EPUB conversion
	var htmlContent string
	if staticFallback := getFileContentSafe(files, "content/static/fallback.html"); staticFallback != "" {
		htmlContent = staticFallback
	} else if mainHTML, exists := files["content/index.html"]; exists {
		htmlContent = string(mainHTML)
	} else {
		return fmt.Errorf("no HTML content found in document")
	}

	// Get CSS content
	cssContent := getFileContentSafe(files, "content/styles/main.css")

	// Create EPUB structure
	epubFiles := make(map[string][]byte)

	// Add mimetype (must be first and uncompressed)
	epubFiles["mimetype"] = []byte("application/epub+zip")

	// Add META-INF/container.xml
	containerXML := `<?xml version="1.0" encoding="UTF-8"?>
<container version="1.0" xmlns="urn:oasis:names:tc:opendocument:xmlns:container">
    <rootfiles>
        <rootfile full-path="OEBPS/content.opf" media-type="application/oebps-package+xml"/>
    </rootfiles>
</container>`
	epubFiles["META-INF/container.xml"] = []byte(containerXML)

	// Generate UUID for EPUB
	uuid := generateUUID()

	// Add content.opf (package document)
	contentOPF := fmt.Sprintf(`<?xml version="1.0" encoding="UTF-8"?>
<package version="3.0" xmlns="http://www.idpf.org/2007/opf" unique-identifier="uid">
    <metadata xmlns:dc="http://purl.org/dc/elements/1.1/">
        <dc:identifier id="uid">urn:uuid:%s</dc:identifier>
        <dc:title>%s</dc:title>
        <dc:creator>%s</dc:creator>
        <dc:language>%s</dc:language>
        <dc:date>%s</dc:date>
        <meta property="dcterms:modified">%s</meta>
    </metadata>
    <manifest>
        <item id="nav" href="nav.xhtml" media-type="application/xhtml+xml" properties="nav"/>
        <item id="ncx" href="toc.ncx" media-type="application/x-dtbncx+xml"/>
        <item id="content" href="content.xhtml" media-type="application/xhtml+xml"/>
        <item id="style" href="styles/main.css" media-type="text/css"/>
    </manifest>
    <spine toc="ncx">
        <itemref idref="content"/>
    </spine>
</package>`,
		uuid,
		escapeXML(doc.Metadata.Title),
		escapeXML(doc.Metadata.Author),
		doc.Metadata.Language,
		doc.Metadata.Created.Format("2006-01-02T15:04:05Z"),
		time.Now().Format("2006-01-02T15:04:05Z"))

	epubFiles["OEBPS/content.opf"] = []byte(contentOPF)

	// Add toc.ncx (EPUB 2 navigation)
	tocNCX := fmt.Sprintf(`<?xml version="1.0" encoding="UTF-8"?>
<ncx version="2005-1" xmlns="http://www.daisy.org/z3986/2005/ncx/">
    <head>
        <meta name="dtb:uid" content="urn:uuid:%s"/>
        <meta name="dtb:depth" content="1"/>
        <meta name="dtb:totalPageCount" content="0"/>
        <meta name="dtb:maxPageNumber" content="0"/>
    </head>
    <docTitle>
        <text>%s</text>
    </docTitle>
    <navMap>
        <navPoint id="navpoint-1" playOrder="1">
            <navLabel>
                <text>Content</text>
            </navLabel>
            <content src="content.xhtml"/>
        </navPoint>
    </navMap>
</ncx>`,
		uuid,
		escapeXML(doc.Metadata.Title))

	epubFiles["OEBPS/toc.ncx"] = []byte(tocNCX)

	// Add nav.xhtml (EPUB 3 navigation)
	navXHTML := `<?xml version="1.0" encoding="UTF-8"?>
<!DOCTYPE html>
<html xmlns="http://www.w3.org/1999/xhtml" xmlns:epub="http://www.idpf.org/2007/ops">
<head>
    <title>Navigation</title>
</head>
<body>
    <nav epub:type="toc" id="toc">
        <h1>Table of Contents</h1>
        <ol>
            <li><a href="content.xhtml">Content</a></li>
        </ol>
    </nav>
</body>
</html>`

	epubFiles["OEBPS/nav.xhtml"] = []byte(navXHTML)

	// Add main content file
	contentXHTML := fmt.Sprintf(`<?xml version="1.0" encoding="UTF-8"?>
<!DOCTYPE html>
<html xmlns="http://www.w3.org/1999/xhtml">
<head>
    <title>%s</title>
    <link rel="stylesheet" type="text/css" href="styles/main.css"/>
</head>
<body>
    %s
</body>
</html>`,
		escapeXML(doc.Metadata.Title),
		htmlContent)

	epubFiles["OEBPS/content.xhtml"] = []byte(contentXHTML)

	// Add CSS file
	if cssContent != "" {
		epubFiles["OEBPS/styles/main.css"] = []byte(cssContent)
	} else {
		// Add default EPUB CSS
		defaultCSS := `body {
    font-family: Georgia, serif;
    line-height: 1.6;
    margin: 1em;
}
h1, h2, h3, h4, h5, h6 {
    font-family: Arial, sans-serif;
    margin-top: 1.5em;
    margin-bottom: 0.5em;
}
p {
    margin-bottom: 1em;
    text-indent: 1.5em;
}
p:first-child, h1 + p, h2 + p, h3 + p {
    text-indent: 0;
}`
		epubFiles["OEBPS/styles/main.css"] = []byte(defaultCSS)
	}

	// Create EPUB file (ZIP format)
	err = zipContainer.CreateFromFiles(epubFiles, outputFile)
	if err != nil {
		return fmt.Errorf("failed to create EPUB file: %v", err)
	}

	fmt.Printf("✓ EPUB exported to: %s\n", outputFile)
	return nil
}

func convertToLIV(inputFile, outputFile string) error {
	fmt.Printf("Converting %s to LIV format...\n", inputFile)

	// Read input file
	inputContent, err := os.ReadFile(inputFile)
	if err != nil {
		return fmt.Errorf("failed to read input file: %v", err)
	}

	// Determine input format based on file extension
	ext := strings.ToLower(filepath.Ext(inputFile))
	var htmlContent, title string

	switch ext {
	case ".html", ".htm":
		htmlContent = string(inputContent)
		// Extract title from HTML
		if titleStart := strings.Index(strings.ToLower(htmlContent), "<title>"); titleStart != -1 {
			titleStart += 7
			if titleEnd := strings.Index(strings.ToLower(htmlContent[titleStart:]), "</title>"); titleEnd != -1 {
				title = htmlContent[titleStart : titleStart+titleEnd]
			}
		}
		if title == "" {
			title = "Imported HTML Document"
		}
	case ".md", ".markdown":
		markdownContent := string(inputContent)
		htmlContent = convertMarkdownToHTML(markdownContent)
		// Extract title from first heading
		lines := strings.Split(markdownContent, "\n")
		for _, line := range lines {
			line = strings.TrimSpace(line)
			if strings.HasPrefix(line, "# ") {
				title = strings.TrimSpace(line[2:])
				break
			}
		}
		if title == "" {
			title = "Imported Markdown Document"
		}
	case ".epub":
		return fmt.Errorf("EPUB to LIV conversion not yet implemented")
	default:
		return fmt.Errorf("unsupported input format: %s (supported: .html, .htm, .md, .markdown)", ext)
	}

	// Create LIV document structure
	files := make(map[string][]byte)

	// Create manifest
	manifest := createImportManifest(title)
	manifestJSON, err := manifest.BuildJSON()
	if err != nil {
		return fmt.Errorf("failed to create manifest: %v", err)
	}
	files["manifest.json"] = manifestJSON

	// Create content files
	files["content/index.html"] = []byte(htmlContent)
	files["content/styles/main.css"] = []byte(generateDefaultCSS())
	files["content/static/fallback.html"] = []byte(stripInteractiveElements(htmlContent))

	// Create LIV file
	zipContainer := container.NewZIPContainer()
	err = zipContainer.CreateFromFiles(files, outputFile)
	if err != nil {
		return fmt.Errorf("failed to create LIV file: %v", err)
	}

	fmt.Printf("✓ LIV document created: %s\n", outputFile)
	return nil
}

func convertToPDF(livFile, outputFile string, quality int) error {
	fmt.Printf("Converting LIV document to PDF...\n")

	// Extract document
	zipContainer := container.NewZIPContainer()
	files, err := zipContainer.ExtractToMemory(livFile)
	if err != nil {
		return fmt.Errorf("failed to extract LIV document: %v", err)
	}

	// Parse manifest
	manifestData, exists := files["manifest.json"]
	if !exists {
		return fmt.Errorf("no manifest found in document")
	}

	manifestParser := manifest.NewManifestParser()
	doc, err := manifestParser.ParseFromBytes(manifestData)
	if err != nil {
		return fmt.Errorf("failed to parse manifest: %v", err)
	}

	// Get content
	htmlContent := getFileContentSafe(files, "content/index.html")
	cssContent := getFileContentSafe(files, "content/styles/main.css")
	staticFallback := getFileContentSafe(files, "content/static/fallback.html")

	// Use static fallback if available, otherwise use main HTML
	contentToConvert := staticFallback
	if contentToConvert == "" {
		contentToConvert = htmlContent
	}

	if contentToConvert == "" {
		return fmt.Errorf("no content found to convert")
	}

	// Create temporary HTML file with embedded CSS for PDF generation
	tempHTML := createPDFReadyHTML(contentToConvert, cssContent, doc.Metadata.Title)

	// Generate PDF using headless browser approach
	err = generatePDFFromHTML(tempHTML, outputFile, quality)
	if err != nil {
		return fmt.Errorf("failed to generate PDF: %v", err)
	}

	fmt.Printf("✓ PDF exported to: %s\n", outputFile)
	return nil
}

func createPDFReadyHTML(htmlContent, cssContent, title string) string {
	// Create complete HTML document optimized for PDF generation
	html := fmt.Sprintf(`<!DOCTYPE html>
<html lang="en">
<head>
    <meta charset="UTF-8">
    <meta name="viewport" content="width=device-width, initial-scale=1.0">
    <title>%s</title>
    <style>
        /* PDF-specific styles */
        @page {
            margin: 1in;
            size: A4;
        }
        
        body {
            font-family: Arial, sans-serif;
            font-size: 12pt;
            line-height: 1.4;
            color: #000;
            background: #fff;
            margin: 0;
            padding: 0;
        }
        
        /* Ensure interactive elements are visible in PDF */
        .interactive-element {
            border: 2px dashed #007bff;
            padding: 10px;
            background: #f8f9fa;
        }
        
        .interactive-element::before {
            content: "Interactive Element: ";
            font-weight: bold;
            color: #007bff;
        }
        
        .chart-container {
            border: 1px solid #ddd;
            padding: 10px;
            background: #f8f9fa;
        }
        
        .chart-container::before {
            content: "Chart: ";
            font-weight: bold;
            color: #28a745;
        }
        
        /* Hide elements that shouldn't appear in PDF */
        .no-print {
            display: none !important;
        }
        
        /* Page break handling */
        .page-break {
            page-break-before: always;
        }
        
        /* Image handling */
        img {
            max-width: 100%%;
            height: auto;
            page-break-inside: avoid;
        }
        
        /* Table handling */
        table {
            border-collapse: collapse;
            width: 100%%;
            page-break-inside: avoid;
        }
        
        /* Custom CSS from document */
        %s
    </style>
</head>
<body>
    %s
</body>
</html>`, title, cssContent, htmlContent)

	return html
}

func generatePDFFromHTML(htmlContent, outputFile string, quality int) error {
	// Try to use headless Chrome/Chromium for PDF generation
	chromePaths := []string{
		"google-chrome",
		"chromium",
		"chromium-browser",
		"chrome",
		"/Applications/Google Chrome.app/Contents/MacOS/Google Chrome",
		"C:\\Program Files\\Google\\Chrome\\Application\\chrome.exe",
		"C:\\Program Files (x86)\\Google\\Chrome\\Application\\chrome.exe",
	}

	var chromePath string
	for _, path := range chromePaths {
		if _, err := exec.LookPath(path); err == nil {
			chromePath = path
			break
		}
		// Check if file exists (for absolute paths)
		if _, err := os.Stat(path); err == nil {
			chromePath = path
			break
		}
	}

	if chromePath == "" {
		return fmt.Errorf("Chrome/Chromium not found. Please install Chrome or Chromium for PDF generation")
	}

	// Create temporary HTML file
	tempDir := os.TempDir()
	tempHTMLFile := filepath.Join(tempDir, fmt.Sprintf("liv-pdf-temp-%d.html", time.Now().Unix()))

	err := os.WriteFile(tempHTMLFile, []byte(htmlContent), 0644)
	if err != nil {
		return fmt.Errorf("failed to create temporary HTML file: %v", err)
	}
	defer os.Remove(tempHTMLFile)

	// Generate PDF using Chrome headless
	args := []string{
		"--headless",
		"--disable-gpu",
		"--no-sandbox",
		"--disable-dev-shm-usage",
		"--print-to-pdf=" + outputFile,
		"--virtual-time-budget=5000",
		"--run-all-compositor-stages-before-draw",
		"file://" + tempHTMLFile,
	}

	cmd := exec.Command(chromePath, args...)
	output, err := cmd.CombinedOutput()
	if err != nil {
		return fmt.Errorf("PDF generation failed: %v\nOutput: %s", err, string(output))
	}

	// Verify PDF was created
	if _, err := os.Stat(outputFile); os.IsNotExist(err) {
		return fmt.Errorf("PDF file was not created")
	}

	return nil
}

// Markdown to HTML conversion function
func convertMarkdownToHTML(markdownContent string) string {
	html := markdownContent

	// Convert headings
	html = strings.ReplaceAll(html, "\n# ", "\n<h1>")
	html = strings.ReplaceAll(html, "\n## ", "\n<h2>")
	html = strings.ReplaceAll(html, "\n### ", "\n<h3>")
	html = strings.ReplaceAll(html, "\n#### ", "\n<h4>")
	html = strings.ReplaceAll(html, "\n##### ", "\n<h5>")
	html = strings.ReplaceAll(html, "\n###### ", "\n<h6>")

	// Handle headings at start of document
	if strings.HasPrefix(html, "# ") {
		html = "<h1>" + html[2:]
	}
	if strings.HasPrefix(html, "## ") {
		html = "<h2>" + html[3:]
	}
	if strings.HasPrefix(html, "### ") {
		html = "<h3>" + html[4:]
	}
	if strings.HasPrefix(html, "#### ") {
		html = "<h4>" + html[5:]
	}
	if strings.HasPrefix(html, "##### ") {
		html = "<h5>" + html[6:]
	}
	if strings.HasPrefix(html, "###### ") {
		html = "<h6>" + html[7:]
	}

	// Close heading tags at line endings
	html = strings.ReplaceAll(html, "<h1>", "<h1>")
	html = strings.ReplaceAll(html, "<h2>", "<h2>")
	html = strings.ReplaceAll(html, "<h3>", "<h3>")
	html = strings.ReplaceAll(html, "<h4>", "<h4>")
	html = strings.ReplaceAll(html, "<h5>", "<h5>")
	html = strings.ReplaceAll(html, "<h6>", "<h6>")

	// Convert emphasis
	html = strings.ReplaceAll(html, "**", "<strong>")
	html = strings.ReplaceAll(html, "*", "<em>")

	// Convert code
	html = strings.ReplaceAll(html, "`", "<code>")
	html = strings.ReplaceAll(html, "```", "<pre>")

	// Convert horizontal rules
	html = strings.ReplaceAll(html, "\n---\n", "\n<hr>\n")
	html = strings.ReplaceAll(html, "\n***\n", "\n<hr>\n")

	// Convert line breaks to paragraphs (simple approach)
	lines := strings.Split(html, "\n")
	var processedLines []string
	inParagraph := false

	for _, line := range lines {
		line = strings.TrimSpace(line)
		if line == "" {
			if inParagraph {
				processedLines = append(processedLines, "</p>")
				inParagraph = false
			}
		} else if strings.HasPrefix(line, "<h") || strings.HasPrefix(line, "<hr") || strings.HasPrefix(line, "<pre") {
			if inParagraph {
				processedLines = append(processedLines, "</p>")
				inParagraph = false
			}
			processedLines = append(processedLines, line)
		} else {
			if !inParagraph {
				processedLines = append(processedLines, "<p>")
				inParagraph = true
			}
			processedLines = append(processedLines, line)
		}
	}

	if inParagraph {
		processedLines = append(processedLines, "</p>")
	}

	return strings.Join(processedLines, "\n")
}

// HTML to Markdown conversion function
func convertHTMLToMarkdown(htmlContent string) string {
	markdown := htmlContent

	// Convert headings
	markdown = strings.ReplaceAll(markdown, "<h1>", "# ")
	markdown = strings.ReplaceAll(markdown, "</h1>", "\n\n")
	markdown = strings.ReplaceAll(markdown, "<h2>", "## ")
	markdown = strings.ReplaceAll(markdown, "</h2>", "\n\n")
	markdown = strings.ReplaceAll(markdown, "<h3>", "### ")
	markdown = strings.ReplaceAll(markdown, "</h3>", "\n\n")
	markdown = strings.ReplaceAll(markdown, "<h4>", "#### ")
	markdown = strings.ReplaceAll(markdown, "</h4>", "\n\n")
	markdown = strings.ReplaceAll(markdown, "<h5>", "##### ")
	markdown = strings.ReplaceAll(markdown, "</h5>", "\n\n")
	markdown = strings.ReplaceAll(markdown, "<h6>", "###### ")
	markdown = strings.ReplaceAll(markdown, "</h6>", "\n\n")

	// Convert paragraphs
	markdown = strings.ReplaceAll(markdown, "<p>", "")
	markdown = strings.ReplaceAll(markdown, "</p>", "\n\n")

	// Convert emphasis
	markdown = strings.ReplaceAll(markdown, "<strong>", "**")
	markdown = strings.ReplaceAll(markdown, "</strong>", "**")
	markdown = strings.ReplaceAll(markdown, "<b>", "**")
	markdown = strings.ReplaceAll(markdown, "</b>", "**")
	markdown = strings.ReplaceAll(markdown, "<em>", "*")
	markdown = strings.ReplaceAll(markdown, "</em>", "*")
	markdown = strings.ReplaceAll(markdown, "<i>", "*")
	markdown = strings.ReplaceAll(markdown, "</i>", "*")

	// Convert code
	markdown = strings.ReplaceAll(markdown, "<code>", "`")
	markdown = strings.ReplaceAll(markdown, "</code>", "`")
	markdown = strings.ReplaceAll(markdown, "<pre>", "```\n")
	markdown = strings.ReplaceAll(markdown, "</pre>", "\n```\n\n")

	// Convert horizontal rules
	markdown = strings.ReplaceAll(markdown, "<hr>", "---\n\n")
	markdown = strings.ReplaceAll(markdown, "<hr/>", "---\n\n")
	markdown = strings.ReplaceAll(markdown, "<hr />", "---\n\n")

	// Convert line breaks
	markdown = strings.ReplaceAll(markdown, "<br>", "\n")
	markdown = strings.ReplaceAll(markdown, "<br/>", "\n")
	markdown = strings.ReplaceAll(markdown, "<br />", "\n")

	// Remove remaining HTML tags (simple approach)
	// This is a basic implementation - for production use, consider using a proper HTML parser
	for strings.Contains(markdown, "<") && strings.Contains(markdown, ">") {
		start := strings.Index(markdown, "<")
		end := strings.Index(markdown[start:], ">")
		if end == -1 {
			break
		}
		markdown = markdown[:start] + markdown[start+end+1:]
	}

	// Clean up extra whitespace
	lines := strings.Split(markdown, "\n")
	var cleanLines []string
	for _, line := range lines {
		cleanLines = append(cleanLines, strings.TrimSpace(line))
	}
	markdown = strings.Join(cleanLines, "\n")

	// Remove excessive newlines
	for strings.Contains(markdown, "\n\n\n") {
		markdown = strings.ReplaceAll(markdown, "\n\n\n", "\n\n")
	}

	return strings.TrimSpace(markdown)
}

// Create manifest for imported documents
func createImportManifest(title string) *manifest.ManifestBuilder {
	builder := manifest.NewManifestBuilder()

	// Set metadata
	metadata := &core.DocumentMetadata{
		Title:       title,
		Author:      "LIV Converter",
		Created:     time.Now(),
		Modified:    time.Now(),
		Description: "Imported document",
		Version:     "1.0.0",
		Language:    "en",
	}
	builder.SetMetadata(metadata)

	// Set security policy (restrictive for imported content)
	security := &core.SecurityPolicy{
		WASMPermissions: &core.WASMPermissions{
			MemoryLimit:     64 * 1024 * 1024, // 64MB
			AllowedImports:  []string{"env"},
			CPUTimeLimit:    5000,
			AllowNetworking: false,
			AllowFileSystem: false,
		},
		JSPermissions: &core.JSPermissions{
			ExecutionMode: "sandboxed",
			AllowedAPIs:   []string{"dom"},
			DOMAccess:     "read",
		},
		NetworkPolicy: &core.NetworkPolicy{
			AllowOutbound: false,
			AllowedHosts:  []string{},
			AllowedPorts:  []int{},
		},
		StoragePolicy: &core.StoragePolicy{
			AllowLocalStorage:   false,
			AllowSessionStorage: false,
			AllowIndexedDB:      false,
			AllowCookies:        false,
		},
		ContentSecurityPolicy: "default-src 'self';",
		TrustedDomains:        []string{},
	}
	builder.SetSecurityPolicy(security)

	// Set feature flags (minimal for imported content)
	features := &core.FeatureFlags{
		Animations:    false,
		Interactivity: false,
		Charts:        false,
		Forms:         false,
		Audio:         false,
		Video:         false,
		WebGL:         false,
		WebAssembly:   false,
	}
	builder.SetFeatureFlags(features)

	// Add resources
	builder.AddResource("content/index.html", &core.Resource{
		Hash: "", // Will be calculated during packaging
		Size: 0,  // Will be calculated during packaging
		Type: "text/html",
	})
	builder.AddResource("content/styles/main.css", &core.Resource{
		Hash: "",
		Size: 0,
		Type: "text/css",
	})
	builder.AddResource("content/static/fallback.html", &core.Resource{
		Hash: "",
		Size: 0,
		Type: "text/html",
	})

	return builder
}

// Generate default CSS for imported documents
func generateDefaultCSS() string {
	return `/* Default Import Styles */
body {
    font-family: -apple-system, BlinkMacSystemFont, 'Segoe UI', Roboto, sans-serif;
    line-height: 1.6;
    color: #333;
    max-width: 800px;
    margin: 0 auto;
    padding: 20px;
}

h1, h2, h3, h4, h5, h6 {
    margin-top: 0;
    margin-bottom: 16px;
    font-weight: 600;
    line-height: 1.25;
}

h1 { font-size: 2em; }
h2 { font-size: 1.5em; }
h3 { font-size: 1.25em; }

p {
    margin-bottom: 16px;
}

a {
    color: #0366d6;
    text-decoration: none;
}

a:hover {
    text-decoration: underline;
}

img {
    max-width: 100%;
    height: auto;
}

code {
    background-color: #f6f8fa;
    border-radius: 3px;
    font-size: 85%;
    margin: 0;
    padding: 0.2em 0.4em;
}

pre {
    background-color: #f6f8fa;
    border-radius: 6px;
    font-size: 85%;
    line-height: 1.45;
    overflow: auto;
    padding: 16px;
}

blockquote {
    border-left: 4px solid #dfe2e5;
    margin: 0;
    padding: 0 16px;
    color: #6a737d;
}

ul, ol {
    margin-bottom: 16px;
    padding-left: 2em;
}

li {
    margin-bottom: 0.25em;
}

hr {
    border: none;
    border-top: 1px solid #e1e4e8;
    margin: 24px 0;
}`
}

// Strip interactive elements for static fallback
func stripInteractiveElements(html string) string {
	// Remove script tags
	staticHTML := html

	// Simple approach to remove script tags
	for strings.Contains(staticHTML, "<script") {
		start := strings.Index(staticHTML, "<script")
		if start == -1 {
			break
		}
		end := strings.Index(staticHTML[start:], "</script>")
		if end == -1 {
			break
		}
		staticHTML = staticHTML[:start] + staticHTML[start+end+9:]
	}

	// Remove event handlers (basic approach)
	staticHTML = strings.ReplaceAll(staticHTML, " onclick=", " data-onclick=")
	staticHTML = strings.ReplaceAll(staticHTML, " onload=", " data-onload=")
	staticHTML = strings.ReplaceAll(staticHTML, " onchange=", " data-onchange=")

	// Convert form elements to static versions
	staticHTML = strings.ReplaceAll(staticHTML, "<input", "<span class=\"static-input\"")
	staticHTML = strings.ReplaceAll(staticHTML, "<button", "<span class=\"static-button\"")
	staticHTML = strings.ReplaceAll(staticHTML, "</button>", "</span>")

	return staticHTML
}

// Generate UUID for EPUB identifier
func generateUUID() string {
	// Simple UUID v4 generation
	b := make([]byte, 16)
	for i := range b {
		b[i] = byte(time.Now().UnixNano() % 256)
	}

	// Set version (4) and variant bits
	b[6] = (b[6] & 0x0f) | 0x40
	b[8] = (b[8] & 0x3f) | 0x80

	return fmt.Sprintf("%x-%x-%x-%x-%x", b[0:4], b[4:6], b[6:8], b[8:10], b[10:16])
}

// Escape XML special characters
func escapeXML(text string) string {
	text = strings.ReplaceAll(text, "&", "&amp;")
	text = strings.ReplaceAll(text, "<", "&lt;")
	text = strings.ReplaceAll(text, ">", "&gt;")
	text = strings.ReplaceAll(text, "\"", "&quot;")
	text = strings.ReplaceAll(text, "'", "&#39;")
	return text
}

// Helper functions

func findBuilderExecutable() (string, error) {
	// Look for builder in common locations
	candidates := []string{
		"./bin/liv-builder.exe",
		"./bin/liv-builder",
		"liv-builder.exe",
		"liv-builder",
	}

	for _, candidate := range candidates {
		if _, err := os.Stat(candidate); err == nil {
			return candidate, nil
		}
	}

	// Try to find in PATH
	if path, err := exec.LookPath("liv-builder"); err == nil {
		return path, nil
	}

	return "", fmt.Errorf("liv-builder executable not found")
}

func findViewerExecutable() (string, error) {
	// Look for viewer in common locations
	candidates := []string{
		"./bin/liv-viewer.exe",
		"./bin/liv-viewer",
		"liv-viewer.exe",
		"liv-viewer",
	}

	for _, candidate := range candidates {
		if _, err := os.Stat(candidate); err == nil {
			return candidate, nil
		}
	}

	// Try to find in PATH
	if path, err := exec.LookPath("liv-viewer"); err == nil {
		return path, nil
	}

	return "", fmt.Errorf("liv-viewer executable not found")
}

func getFileContentSafe(files map[string][]byte, path string) string {
	if content, exists := files[path]; exists {
		return string(content)
	}
	return ""
}

func runValidate(file string, checkSignatures, verbose bool) error {
	if verbose {
		fmt.Printf("Validating LIV document: %s\n", file)
	}

	// Check if file exists
	if _, err := os.Stat(file); os.IsNotExist(err) {
		return fmt.Errorf("file not found: %s", file)
	}

	// Create ZIP container for validation
	zipContainer := container.NewZIPContainer()

	// Validate ZIP structure
	structureResult := zipContainer.ValidateStructure(file)

	if verbose {
		fmt.Printf("\nStructure Validation:\n")
	}

	if structureResult.IsValid {
		fmt.Printf("✓ Document structure is valid\n")
	} else {
		fmt.Printf("✗ Document structure is invalid\n")
		for _, err := range structureResult.Errors {
			fmt.Printf("  Error: %s\n", err)
		}
	}

	if len(structureResult.Warnings) > 0 {
		fmt.Printf("Warnings:\n")
		for _, warning := range structureResult.Warnings {
			fmt.Printf("  Warning: %s\n", warning)
		}
	}

	// Extract and validate manifest
	files, err := zipContainer.ExtractToMemory(file)
	if err != nil {
		return fmt.Errorf("failed to extract document: %v", err)
	}

	manifestData, exists := files["manifest.json"]
	if !exists {
		return fmt.Errorf("manifest.json not found in document")
	}

	// Validate manifest
	validator := manifest.NewManifestValidator()
	parsedManifest, manifestResult := validator.ValidateManifestJSON(manifestData)

	if verbose {
		fmt.Printf("\nManifest Validation:\n")
	}

	if manifestResult.IsValid {
		fmt.Printf("✓ Manifest is valid\n")
	} else {
		fmt.Printf("✗ Manifest is invalid\n")
		for _, err := range manifestResult.Errors {
			fmt.Printf("  Error: %s\n", err)
		}
	}

	if len(manifestResult.Warnings) > 0 {
		fmt.Printf("Manifest Warnings:\n")
		for _, warning := range manifestResult.Warnings {
			fmt.Printf("  Warning: %s\n", warning)
		}
	}

	// Check signatures if requested
	if checkSignatures && parsedManifest != nil {
		if verbose {
			fmt.Printf("\nSignature Validation:\n")
		}

		// Create document structure for signature verification
		document := &core.LIVDocument{
			Manifest: parsedManifest,
			Content: &core.DocumentContent{
				HTML:            string(files["content/index.html"]),
				CSS:             getFileContentSafe(files, "content/styles/main.css"),
				InteractiveSpec: getFileContentSafe(files, "content/interactive.json"),
				StaticFallback:  getFileContentSafe(files, "content/static/fallback.html"),
			},
			WASMModules: make(map[string][]byte),
		}

		// Add WASM modules
		for path, content := range files {
			if strings.HasSuffix(path, ".wasm") {
				moduleName := strings.TrimSuffix(filepath.Base(path), ".wasm")
				document.WASMModules[moduleName] = content
			}
		}

		// Check if document has signatures
		if document.Signatures == nil {
			fmt.Printf("⚠ Document is not signed\n")
		} else {
			fmt.Printf("✓ Document contains signatures\n")
			// Note: Full signature verification would require the public key
			fmt.Printf("  Manifest signature: %s...\n", document.Signatures.ManifestSignature[:16])
			fmt.Printf("  Content signature: %s...\n", document.Signatures.ContentSignature[:16])
			if len(document.Signatures.WASMSignatures) > 0 {
				fmt.Printf("  WASM signatures: %d modules\n", len(document.Signatures.WASMSignatures))
			}
		}
	}

	// Summary
	fmt.Printf("\nValidation Summary:\n")
	allValid := structureResult.IsValid && manifestResult.IsValid
	if allValid {
		fmt.Printf("✓ Document is valid\n")
		return nil
	} else {
		fmt.Printf("✗ Document has validation errors\n")
		return fmt.Errorf("validation failed")
	}
}

func runSign(file, keyFile, outputFile string) error {
	fmt.Printf("Signing LIV document: %s\n", file)

	// Check if files exist
	if _, err := os.Stat(file); os.IsNotExist(err) {
		return fmt.Errorf("input file not found: %s", file)
	}

	if _, err := os.Stat(keyFile); os.IsNotExist(err) {
		return fmt.Errorf("key file not found: %s", keyFile)
	}

	// Set output file if not specified
	if outputFile == "" {
		outputFile = file // Overwrite original
	}

	// Create signature manager
	sigManager := integrity.NewSignatureManager()

	// Load private key
	privateKey, err := sigManager.LoadPrivateKeyPEM(keyFile)
	if err != nil {
		return fmt.Errorf("failed to load private key: %v", err)
	}

	// Load document
	zipContainer := container.NewZIPContainer()
	files, err := zipContainer.ExtractToMemory(file)
	if err != nil {
		return fmt.Errorf("failed to extract document: %v", err)
	}

	// Parse manifest
	manifestData, exists := files["manifest.json"]
	if !exists {
		return fmt.Errorf("manifest.json not found in document")
	}

	validator := manifest.NewManifestValidator()
	parsedManifest, result := validator.ValidateManifestJSON(manifestData)
	if !result.IsValid {
		return fmt.Errorf("invalid manifest: %v", result.Errors)
	}

	// Create document structure for signing
	document := &core.LIVDocument{
		Manifest: parsedManifest,
		Content: &core.DocumentContent{
			HTML:            string(files["content/index.html"]),
			CSS:             getFileContentSafe(files, "content/styles/main.css"),
			InteractiveSpec: getFileContentSafe(files, "content/interactive.json"),
			StaticFallback:  getFileContentSafe(files, "content/static/fallback.html"),
		},
		WASMModules: make(map[string][]byte),
	}

	// Add WASM modules
	for path, content := range files {
		if strings.HasSuffix(path, ".wasm") {
			moduleName := strings.TrimSuffix(filepath.Base(path), ".wasm")
			document.WASMModules[moduleName] = content
		}
	}

	// Sign the document
	fmt.Printf("Generating signatures...\n")
	signatures, err := sigManager.SignDocument(document, privateKey)
	if err != nil {
		return fmt.Errorf("failed to sign document: %v", err)
	}

	// Update document with signatures
	document.Signatures = signatures

	// Update manifest with new modification time
	document.Manifest.Metadata.Modified = time.Now()

	// Re-serialize manifest
	manifestBuilder := manifest.NewManifestBuilder()
	manifestBuilder.SetMetadata(document.Manifest.Metadata)
	manifestBuilder.SetSecurityPolicy(document.Manifest.Security)
	if document.Manifest.WASMConfig != nil {
		manifestBuilder.SetWASMConfig(document.Manifest.WASMConfig)
	}
	if document.Manifest.Features != nil {
		manifestBuilder.SetFeatureFlags(document.Manifest.Features)
	}

	// Add resources back
	for path, resource := range document.Manifest.Resources {
		manifestBuilder.AddResource(path, resource)
	}

	updatedManifestData, err := manifestBuilder.BuildJSON()
	if err != nil {
		return fmt.Errorf("failed to build updated manifest: %v", err)
	}

	// Update files with new manifest
	files["manifest.json"] = updatedManifestData

	// Create signed document
	fmt.Printf("Creating signed document...\n")
	err = zipContainer.CreateFromFiles(files, outputFile)
	if err != nil {
		return fmt.Errorf("failed to create signed document: %v", err)
	}

	fmt.Printf("✓ Document signed successfully\n")
	fmt.Printf("  Manifest signature: %s...\n", signatures.ManifestSignature[:16])
	fmt.Printf("  Content signature: %s...\n", signatures.ContentSignature[:16])
	if len(signatures.WASMSignatures) > 0 {
		fmt.Printf("  WASM signatures: %d modules\n", len(signatures.WASMSignatures))
	}
	fmt.Printf("  Output: %s\n", outputFile)

	return nil
}
