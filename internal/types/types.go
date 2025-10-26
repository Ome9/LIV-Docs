package types

// PDFData represents the parsed PDF content
type PDFData struct {
	Metadata PDFMetadata
	Pages    []PDFPage
}

// PDFMetadata contains PDF document metadata
type PDFMetadata struct {
	Title      string
	Author     string
	Subject    string
	Keywords   string
	Creator    string
	Producer   string
	CreatedAt  string
	ModifiedAt string
}

// PDFPage represents a single PDF page
type PDFPage struct {
	Number     int
	Width      float64
	Height     float64
	Rotation   int
	TextBlocks []PDFTextBlock
	Images     []PDFImage
	Graphics   []PDFGraphic
}

// PDFTextBlock represents a block of text with positioning
type PDFTextBlock struct {
	Text     string
	X        float64
	Y        float64
	Width    float64
	Height   float64
	FontName string
	FontSize float64
	Color    string
	Bold     bool
	Italic   bool
}

// PDFImage represents an embedded image
type PDFImage struct {
	ID     string
	X      float64
	Y      float64
	Width  float64
	Height float64
	Data   []byte
	Format string // "jpeg", "png", etc.
	DPI    int
}

// PDFGraphic represents vector graphics (lines, shapes)
type PDFGraphic struct {
	Type        string // "line", "rect", "path"
	X           float64
	Y           float64
	Width       float64
	Height      float64
	Color       string
	StrokeWidth float64
	Path        string // SVG path for complex shapes
}

// LIVDocument represents the complete LIV document structure
type LIVDocument struct {
	Version string         `json:"version"`
	Format  string         `json:"format"`
	Pages   []LIVPage      `json:"pages"`
	Styles  map[string]any `json:"styles,omitempty"`
	Scripts []string       `json:"scripts,omitempty"`
}

// LIVPage represents a single page in the LIV document
type LIVPage struct {
	ID       string       `json:"id"`
	Number   int          `json:"number"`
	Width    float64      `json:"width"`
	Height   float64      `json:"height"`
	Elements []LIVElement `json:"elements"`
}

// LIVElement represents a document element (text, image, shape, etc.)
type LIVElement struct {
	ID         string         `json:"id"`
	Type       string         `json:"type"` // "text", "image", "shape", "container"
	Content    string         `json:"content,omitempty"`
	Position   ElementPos     `json:"position"`
	Style      ElementStyle   `json:"style,omitempty"`
	Properties map[string]any `json:"properties,omitempty"`
}

// ElementPos defines element positioning
type ElementPos struct {
	X      float64 `json:"x"`
	Y      float64 `json:"y"`
	Width  float64 `json:"width"`
	Height float64 `json:"height"`
	ZIndex int     `json:"z_index,omitempty"`
}

// ElementStyle defines element styling
type ElementStyle struct {
	FontFamily string  `json:"font_family,omitempty"`
	FontSize   float64 `json:"font_size,omitempty"`
	FontWeight string  `json:"font_weight,omitempty"`
	FontStyle  string  `json:"font_style,omitempty"`
	Color      string  `json:"color,omitempty"`
	Background string  `json:"background,omitempty"`
	Border     string  `json:"border,omitempty"`
	Padding    string  `json:"padding,omitempty"`
	TextAlign  string  `json:"text_align,omitempty"`
	LineHeight float64 `json:"line_height,omitempty"`
}

// LIVManifest represents the manifest.json structure
type LIVManifest struct {
	Version     string              `json:"version"`
	Format      string              `json:"format"`
	Metadata    ManifestMetadata    `json:"metadata"`
	Permissions ManifestPermissions `json:"permissions"`
	Pages       int                 `json:"pages"`
	Assets      ManifestAssets      `json:"assets"`
	Compression bool                `json:"compression"`
	Source      ManifestSource      `json:"source"`
}

// ManifestMetadata contains document metadata
type ManifestMetadata struct {
	Title       string `json:"title"`
	Author      string `json:"author"`
	Subject     string `json:"subject"`
	Keywords    string `json:"keywords"`
	Creator     string `json:"creator"`
	Producer    string `json:"producer"`
	CreatedAt   string `json:"created_at"`
	ModifiedAt  string `json:"modified_at"`
	GeneratedAt string `json:"generated_at"`
}

// ManifestPermissions defines document permissions
type ManifestPermissions struct {
	AllowScripts     bool `json:"allow_scripts"`
	AllowExternalNet bool `json:"allow_external_net"`
	AllowPrint       bool `json:"allow_print"`
	AllowCopy        bool `json:"allow_copy"`
	AllowModify      bool `json:"allow_modify"`
}

// ManifestAssets lists document assets
type ManifestAssets struct {
	Images []string `json:"images"`
	Fonts  []string `json:"fonts"`
	Styles []string `json:"styles"`
}

// ManifestSource tracks the source of the LIV document
type ManifestSource struct {
	Type     string `json:"type"`
	Original string `json:"original"`
}

// ExtractedAssets contains all assets extracted from PDF
type ExtractedAssets struct {
	Images []AssetImage
	Fonts  []AssetFont
}

// AssetImage represents an extracted image asset
type AssetImage struct {
	ID       string
	Filename string
	Data     []byte
	Format   string
	Width    int
	Height   int
}

// AssetFont represents an extracted font asset
type AssetFont struct {
	ID       string
	Filename string
	Data     []byte
	Family   string
	Style    string
}
