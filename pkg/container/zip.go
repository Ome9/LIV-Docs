package container

import (
	"archive/zip"
	"bytes"
	"compress/flate"
	"crypto/sha256"
	"encoding/hex"
	"fmt"
	"io"
	"os"
	"path/filepath"
	"strings"
	"time"

	"github.com/liv-format/liv/pkg/core"
)

// ZIPContainer handles ZIP-based .liv file operations
type ZIPContainer struct {
	compressionLevel int
	validateStructure bool
}

// NewZIPContainer creates a new ZIP container handler
func NewZIPContainer() *ZIPContainer {
	return &ZIPContainer{
		compressionLevel:  flate.DefaultCompression,
		validateStructure: true,
	}
}

// SetCompressionLevel sets the compression level (0-9, -1 for default)
func (zc *ZIPContainer) SetCompressionLevel(level int) *ZIPContainer {
	zc.compressionLevel = level
	return zc
}

// SetValidateStructure enables/disables structure validation
func (zc *ZIPContainer) SetValidateStructure(validate bool) *ZIPContainer {
	zc.validateStructure = validate
	return zc
}

// CreateFromDirectory creates a .liv file from a directory structure
func (zc *ZIPContainer) CreateFromDirectory(sourceDir, outputPath string) error {
	// Create output file
	outFile, err := os.Create(outputPath)
	if err != nil {
		return fmt.Errorf("failed to create output file: %v", err)
	}
	defer outFile.Close()

	// Create ZIP writer
	zipWriter := zip.NewWriter(outFile)
	defer zipWriter.Close()

	// Set compression level
	zipWriter.RegisterCompressor(zip.Deflate, func(out io.Writer) (io.WriteCloser, error) {
		return flate.NewWriter(out, zc.compressionLevel)
	})

	// Walk directory and add files
	return filepath.Walk(sourceDir, func(filePath string, info os.FileInfo, err error) error {
		if err != nil {
			return err
		}

		// Skip directories
		if info.IsDir() {
			return nil
		}

		// Calculate relative path
		relPath, err := filepath.Rel(sourceDir, filePath)
		if err != nil {
			return fmt.Errorf("failed to get relative path: %v", err)
		}

		// Normalize path separators for ZIP format
		relPath = filepath.ToSlash(relPath)

		// Add file to ZIP
		return zc.addFileToZip(zipWriter, filePath, relPath)
	})
}

// CreateFromFiles creates a .liv file from a map of file paths to content
func (zc *ZIPContainer) CreateFromFiles(files map[string][]byte, outputPath string) error {
	// Create output file
	outFile, err := os.Create(outputPath)
	if err != nil {
		return fmt.Errorf("failed to create output file: %v", err)
	}
	defer outFile.Close()

	return zc.CreateFromFilesToWriter(files, outFile)
}

// CreateFromFilesToWriter creates a .liv file and writes to an io.Writer
func (zc *ZIPContainer) CreateFromFilesToWriter(files map[string][]byte, writer io.Writer) error {
	// Create ZIP writer
	zipWriter := zip.NewWriter(writer)
	defer zipWriter.Close()

	// Set compression level
	zipWriter.RegisterCompressor(zip.Deflate, func(out io.Writer) (io.WriteCloser, error) {
		return flate.NewWriter(out, zc.compressionLevel)
	})

	// Validate structure if enabled
	if zc.validateStructure {
		if err := zc.validateFileStructure(files); err != nil {
			return fmt.Errorf("structure validation failed: %v", err)
		}
	}

	// Add files to ZIP in a consistent order
	orderedPaths := zc.getOrderedPaths(files)
	
	for _, path := range orderedPaths {
		content := files[path]
		
		// Create ZIP file header
		header := &zip.FileHeader{
			Name:     path,
			Method:   zip.Deflate,
			Modified: time.Now(),
		}
		
		// Set compression method based on file type
		if zc.shouldCompress(path) {
			header.Method = zip.Deflate
		} else {
			header.Method = zip.Store
		}

		// Create writer for this file
		fileWriter, err := zipWriter.CreateHeader(header)
		if err != nil {
			return fmt.Errorf("failed to create ZIP entry for %s: %v", path, err)
		}

		// Write content
		if _, err := fileWriter.Write(content); err != nil {
			return fmt.Errorf("failed to write content for %s: %v", path, err)
		}
	}

	return nil
}

// ExtractToDirectory extracts a .liv file to a directory
func (zc *ZIPContainer) ExtractToDirectory(livPath, targetDir string) error {
	// Open .liv file
	reader, err := zip.OpenReader(livPath)
	if err != nil {
		return fmt.Errorf("failed to open .liv file: %v", err)
	}
	defer reader.Close()

	return zc.extractZipToDirectory(&reader.Reader, targetDir)
}

// ExtractFromReader extracts a .liv file from an io.ReaderAt
func (zc *ZIPContainer) ExtractFromReader(reader io.ReaderAt, size int64, targetDir string) error {
	zipReader, err := zip.NewReader(reader, size)
	if err != nil {
		return fmt.Errorf("failed to create ZIP reader: %v", err)
	}

	return zc.extractZipToDirectory(zipReader, targetDir)
}

// ExtractToMemory extracts a .liv file to memory as a map of paths to content
func (zc *ZIPContainer) ExtractToMemory(livPath string) (map[string][]byte, error) {
	// Open .liv file
	reader, err := zip.OpenReader(livPath)
	if err != nil {
		return nil, fmt.Errorf("failed to open .liv file: %v", err)
	}
	defer reader.Close()

	return zc.extractZipToMemory(&reader.Reader)
}

// ExtractFromReaderToMemory extracts a .liv file from an io.ReaderAt to memory
func (zc *ZIPContainer) ExtractFromReaderToMemory(reader io.ReaderAt, size int64) (map[string][]byte, error) {
	zipReader, err := zip.NewReader(reader, size)
	if err != nil {
		return nil, fmt.Errorf("failed to create ZIP reader: %v", err)
	}

	return zc.extractZipToMemory(zipReader)
}

// ValidateStructure validates the structure of a .liv file
func (zc *ZIPContainer) ValidateStructure(livPath string) *core.ValidationResult {
	files, err := zc.ExtractToMemory(livPath)
	if err != nil {
		return &core.ValidationResult{
			IsValid: false,
			Errors:  []string{fmt.Sprintf("failed to extract file: %v", err)},
		}
	}

	return zc.validateExtractedStructure(files)
}

// ValidateStructureFromMemory validates the structure of extracted files
func (zc *ZIPContainer) ValidateStructureFromMemory(files map[string][]byte) *core.ValidationResult {
	return zc.validateExtractedStructure(files)
}

// GetFileList returns a list of files in a .liv archive
func (zc *ZIPContainer) GetFileList(livPath string) ([]string, error) {
	reader, err := zip.OpenReader(livPath)
	if err != nil {
		return nil, fmt.Errorf("failed to open .liv file: %v", err)
	}
	defer reader.Close()

	var files []string
	for _, file := range reader.File {
		if !file.FileInfo().IsDir() {
			files = append(files, file.Name)
		}
	}

	return files, nil
}

// GetFileInfo returns information about files in a .liv archive
func (zc *ZIPContainer) GetFileInfo(livPath string) (map[string]FileInfo, error) {
	reader, err := zip.OpenReader(livPath)
	if err != nil {
		return nil, fmt.Errorf("failed to open .liv file: %v", err)
	}
	defer reader.Close()

	fileInfos := make(map[string]FileInfo)
	
	for _, file := range reader.File {
		if !file.FileInfo().IsDir() {
			fileInfos[file.Name] = FileInfo{
				Path:             file.Name,
				Size:             int64(file.UncompressedSize64),
				CompressedSize:   int64(file.CompressedSize64),
				Modified:         file.Modified,
				CompressionRatio: float64(file.CompressedSize64) / float64(file.UncompressedSize64),
				Method:           file.Method,
			}
		}
	}

	return fileInfos, nil
}

// FileInfo contains information about a file in the archive
type FileInfo struct {
	Path             string    `json:"path"`
	Size             int64     `json:"size"`
	CompressedSize   int64     `json:"compressed_size"`
	Modified         time.Time `json:"modified"`
	CompressionRatio float64   `json:"compression_ratio"`
	Method           uint16    `json:"method"`
}

// Helper methods

func (zc *ZIPContainer) addFileToZip(zipWriter *zip.Writer, filePath, zipPath string) error {
	// Open source file
	file, err := os.Open(filePath)
	if err != nil {
		return fmt.Errorf("failed to open file %s: %v", filePath, err)
	}
	defer file.Close()

	// Get file info
	info, err := file.Stat()
	if err != nil {
		return fmt.Errorf("failed to get file info for %s: %v", filePath, err)
	}

	// Create ZIP file header
	header := &zip.FileHeader{
		Name:     zipPath,
		Modified: info.ModTime(),
	}

	// Set compression method
	if zc.shouldCompress(zipPath) {
		header.Method = zip.Deflate
	} else {
		header.Method = zip.Store
	}

	// Create writer for this file
	fileWriter, err := zipWriter.CreateHeader(header)
	if err != nil {
		return fmt.Errorf("failed to create ZIP entry for %s: %v", zipPath, err)
	}

	// Copy file content
	if _, err := io.Copy(fileWriter, file); err != nil {
		return fmt.Errorf("failed to write file %s to ZIP: %v", zipPath, err)
	}

	return nil
}

func (zc *ZIPContainer) extractZipToDirectory(zipReader *zip.Reader, targetDir string) error {
	// Create target directory
	if err := os.MkdirAll(targetDir, 0755); err != nil {
		return fmt.Errorf("failed to create target directory: %v", err)
	}

	// Extract files
	for _, file := range zipReader.File {
		if err := zc.extractFile(file, targetDir); err != nil {
			return fmt.Errorf("failed to extract file %s: %v", file.Name, err)
		}
	}

	return nil
}

func (zc *ZIPContainer) extractZipToMemory(zipReader *zip.Reader) (map[string][]byte, error) {
	files := make(map[string][]byte)

	for _, file := range zipReader.File {
		// Skip directories
		if file.FileInfo().IsDir() {
			continue
		}

		// Open file in ZIP
		reader, err := file.Open()
		if err != nil {
			return nil, fmt.Errorf("failed to open file %s in ZIP: %v", file.Name, err)
		}

		// Read content
		content, err := io.ReadAll(reader)
		reader.Close()
		if err != nil {
			return nil, fmt.Errorf("failed to read file %s: %v", file.Name, err)
		}

		files[file.Name] = content
	}

	return files, nil
}

func (zc *ZIPContainer) extractFile(file *zip.File, targetDir string) error {
	// Skip directories
	if file.FileInfo().IsDir() {
		return nil
	}

	// Create full path
	fullPath := filepath.Join(targetDir, file.Name)

	// Security check: prevent directory traversal
	if !strings.HasPrefix(fullPath, filepath.Clean(targetDir)+string(os.PathSeparator)) {
		return fmt.Errorf("invalid file path: %s", file.Name)
	}

	// Create directory for file
	if err := os.MkdirAll(filepath.Dir(fullPath), 0755); err != nil {
		return fmt.Errorf("failed to create directory: %v", err)
	}

	// Open file in ZIP
	reader, err := file.Open()
	if err != nil {
		return fmt.Errorf("failed to open file in ZIP: %v", err)
	}
	defer reader.Close()

	// Create target file
	outFile, err := os.Create(fullPath)
	if err != nil {
		return fmt.Errorf("failed to create target file: %v", err)
	}
	defer outFile.Close()

	// Copy content
	if _, err := io.Copy(outFile, reader); err != nil {
		return fmt.Errorf("failed to copy file content: %v", err)
	}

	// Set file modification time
	if err := os.Chtimes(fullPath, file.Modified, file.Modified); err != nil {
		// Non-fatal error, just log it
		fmt.Printf("Warning: failed to set modification time for %s: %v\n", fullPath, err)
	}

	return nil
}

func (zc *ZIPContainer) validateFileStructure(files map[string][]byte) error {
	requiredFiles := []string{
		"manifest.json",
	}

	requiredDirs := []string{
		"content/",
	}

	// Check required files
	for _, required := range requiredFiles {
		if _, exists := files[required]; !exists {
			return fmt.Errorf("required file missing: %s", required)
		}
	}

	// Check required directories (at least one file in each)
	for _, requiredDir := range requiredDirs {
		found := false
		for path := range files {
			if strings.HasPrefix(path, requiredDir) {
				found = true
				break
			}
		}
		if !found {
			return fmt.Errorf("required directory missing or empty: %s", requiredDir)
		}
	}

	// Validate file paths
	for path := range files {
		if err := zc.validateFilePath(path); err != nil {
			return fmt.Errorf("invalid file path %s: %v", path, err)
		}
	}

	return nil
}

func (zc *ZIPContainer) validateExtractedStructure(files map[string][]byte) *core.ValidationResult {
	var errors []string
	var warnings []string

	// Check for required files
	requiredFiles := []string{
		"manifest.json",
		"content/index.html",
	}

	for _, required := range requiredFiles {
		if _, exists := files[required]; !exists {
			errors = append(errors, fmt.Sprintf("required file missing: %s", required))
		}
	}

	// Check for recommended files
	recommendedFiles := []string{
		"content/static/fallback.html",
	}

	for _, recommended := range recommendedFiles {
		if _, exists := files[recommended]; !exists {
			warnings = append(warnings, fmt.Sprintf("recommended file missing: %s", recommended))
		}
	}

	// Validate file paths
	for path := range files {
		if err := zc.validateFilePath(path); err != nil {
			errors = append(errors, fmt.Sprintf("invalid file path %s: %v", path, err))
		}
	}

	// Check for suspicious files
	suspiciousExtensions := []string{".exe", ".bat", ".sh", ".cmd", ".scr"}
	for path := range files {
		for _, ext := range suspiciousExtensions {
			if strings.HasSuffix(strings.ToLower(path), ext) {
				warnings = append(warnings, fmt.Sprintf("suspicious file type: %s", path))
			}
		}
	}

	// Check total size
	totalSize := int64(0)
	for _, content := range files {
		totalSize += int64(len(content))
	}

	if totalSize > 100*1024*1024 { // 100MB
		warnings = append(warnings, fmt.Sprintf("document is very large: %d bytes", totalSize))
	}

	return &core.ValidationResult{
		IsValid:  len(errors) == 0,
		Errors:   errors,
		Warnings: warnings,
	}
}

func (zc *ZIPContainer) validateFilePath(path string) error {
	// Check for directory traversal
	if strings.Contains(path, "..") {
		return fmt.Errorf("path contains directory traversal")
	}

	// Check for absolute paths
	if filepath.IsAbs(path) {
		return fmt.Errorf("absolute paths not allowed")
	}

	// Check for invalid characters
	invalidChars := []string{"<", ">", ":", "\"", "|", "?", "*"}
	for _, char := range invalidChars {
		if strings.Contains(path, char) {
			return fmt.Errorf("path contains invalid character: %s", char)
		}
	}

	// Check path length
	if len(path) > 260 {
		return fmt.Errorf("path too long: %d characters", len(path))
	}

	return nil
}

func (zc *ZIPContainer) shouldCompress(path string) bool {
	// Don't compress already compressed formats
	noCompressExtensions := []string{
		".png", ".jpg", ".jpeg", ".gif", ".webp",
		".mp3", ".mp4", ".webm", ".ogg",
		".woff", ".woff2", ".ttf",
		".zip", ".gz", ".bz2",
		".wasm", // WASM files are already optimized
	}

	ext := strings.ToLower(filepath.Ext(path))
	for _, noCompress := range noCompressExtensions {
		if ext == noCompress {
			return false
		}
	}

	return true
}

func (zc *ZIPContainer) getOrderedPaths(files map[string][]byte) []string {
	// Define priority order for files
	priorityFiles := []string{
		"manifest.json",
		"content/index.html",
		"content/static/fallback.html",
	}

	var orderedPaths []string
	used := make(map[string]bool)

	// Add priority files first
	for _, priority := range priorityFiles {
		if _, exists := files[priority]; exists {
			orderedPaths = append(orderedPaths, priority)
			used[priority] = true
		}
	}

	// Add remaining files in alphabetical order
	var remaining []string
	for path := range files {
		if !used[path] {
			remaining = append(remaining, path)
		}
	}

	// Sort remaining files
	for i := 0; i < len(remaining); i++ {
		for j := i + 1; j < len(remaining); j++ {
			if remaining[i] > remaining[j] {
				remaining[i], remaining[j] = remaining[j], remaining[i]
			}
		}
	}

	orderedPaths = append(orderedPaths, remaining...)
	return orderedPaths
}

// Utility functions for working with .liv files

// CalculateFileHash calculates SHA-256 hash of file content
func CalculateFileHash(content []byte) string {
	hash := sha256.Sum256(content)
	return hex.EncodeToString(hash[:])
}

// CompressFiles compresses a map of files and returns compression statistics
func (zc *ZIPContainer) CompressFiles(files map[string][]byte) (*CompressionStats, error) {
	var buf bytes.Buffer
	
	if err := zc.CreateFromFilesToWriter(files, &buf); err != nil {
		return nil, err
	}

	// Calculate statistics
	originalSize := int64(0)
	for _, content := range files {
		originalSize += int64(len(content))
	}

	compressedSize := int64(buf.Len())
	
	return &CompressionStats{
		OriginalSize:     originalSize,
		CompressedSize:   compressedSize,
		CompressionRatio: float64(compressedSize) / float64(originalSize),
		FileCount:        len(files),
	}, nil
}

// CompressionStats contains compression statistics
type CompressionStats struct {
	OriginalSize     int64   `json:"original_size"`
	CompressedSize   int64   `json:"compressed_size"`
	CompressionRatio float64 `json:"compression_ratio"`
	FileCount        int     `json:"file_count"`
}

// DeduplicateFiles removes duplicate files based on content hash
func DeduplicateFiles(files map[string][]byte) (map[string][]byte, map[string]string) {
	// Build a map from hash to list of paths to ensure deterministic selection
	hashToPaths := make(map[string][]string)
	for path, content := range files {
		hash := CalculateFileHash(content)
		hashToPaths[hash] = append(hashToPaths[hash], path)
	}

	deduplicated := make(map[string][]byte)
	duplicates := make(map[string]string)

	for _, paths := range hashToPaths {
		// Choose a deterministic original: lexicographically smallest path
		original := paths[0]
		for _, p := range paths[1:] {
			if p < original {
				original = p
			}
		}

		// Add original content
		deduplicated[original] = files[original]

		// Mark other paths as duplicates
		for _, p := range paths {
			if p == original {
				continue
			}
			duplicates[p] = original
		}
	}

	return deduplicated, duplicates
}