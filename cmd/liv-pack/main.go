package main

import (
	"fmt"
	"os"
	"path/filepath"

	"github.com/spf13/cobra"
	"github.com/liv-format/liv/pkg/container"
)

func main() {
	var (
		compressionLevel int
		verbose          bool
		validate         bool
	)

	rootCmd := &cobra.Command{
		Use:   "liv-pack",
		Short: "LIV Package Management Tool",
		Long: `LIV Pack provides low-level ZIP container operations for .liv files.
This tool handles the packaging, extraction, and validation of .liv file containers.`,
	}

	// Pack command
	packCmd := &cobra.Command{
		Use:   "pack [source-dir] [output.liv]",
		Short: "Pack a directory into a .liv file",
		Long:  "Pack creates a .liv file from a directory structure with proper compression and validation.",
		Args:  cobra.ExactArgs(2),
		RunE: func(cmd *cobra.Command, args []string) error {
			return packDirectory(args[0], args[1], compressionLevel, verbose, validate)
		},
	}

	packCmd.Flags().IntVarP(&compressionLevel, "compression", "c", -1, "Compression level (0-9, -1 for default)")
	packCmd.Flags().BoolVarP(&verbose, "verbose", "v", false, "Verbose output")
	packCmd.Flags().BoolVarP(&validate, "validate", "", true, "Validate structure")

	// Unpack command
	unpackCmd := &cobra.Command{
		Use:   "unpack [input.liv] [target-dir]",
		Short: "Unpack a .liv file to a directory",
		Long:  "Unpack extracts a .liv file to a directory structure for inspection or editing.",
		Args:  cobra.ExactArgs(2),
		RunE: func(cmd *cobra.Command, args []string) error {
			return unpackFile(args[0], args[1], verbose)
		},
	}

	unpackCmd.Flags().BoolVarP(&verbose, "verbose", "v", false, "Verbose output")

	// List command
	listCmd := &cobra.Command{
		Use:   "list [input.liv]",
		Short: "List contents of a .liv file",
		Long:  "List shows the files contained in a .liv archive with size and compression information.",
		Args:  cobra.ExactArgs(1),
		RunE: func(cmd *cobra.Command, args []string) error {
			return listContents(args[0], verbose)
		},
	}

	listCmd.Flags().BoolVarP(&verbose, "verbose", "v", false, "Show detailed information")

	// Validate command
	validateCmd := &cobra.Command{
		Use:   "validate [input.liv]",
		Short: "Validate a .liv file structure",
		Long:  "Validate checks the internal structure and security of a .liv file.",
		Args:  cobra.ExactArgs(1),
		RunE: func(cmd *cobra.Command, args []string) error {
			return validateFile(args[0], verbose)
		},
	}

	validateCmd.Flags().BoolVarP(&verbose, "verbose", "v", false, "Show detailed validation results")

	// Info command
	infoCmd := &cobra.Command{
		Use:   "info [input.liv]",
		Short: "Show information about a .liv file",
		Long:  "Info displays detailed information about a .liv file including compression statistics.",
		Args:  cobra.ExactArgs(1),
		RunE: func(cmd *cobra.Command, args []string) error {
			return showInfo(args[0])
		},
	}

	// Add subcommands
	rootCmd.AddCommand(packCmd)
	rootCmd.AddCommand(unpackCmd)
	rootCmd.AddCommand(listCmd)
	rootCmd.AddCommand(validateCmd)
	rootCmd.AddCommand(infoCmd)

	if err := rootCmd.Execute(); err != nil {
		fmt.Fprintf(os.Stderr, "Error: %v\n", err)
		os.Exit(1)
	}
}

func packDirectory(sourceDir, outputPath string, compressionLevel int, verbose, validate bool) error {
	if verbose {
		fmt.Printf("Packing directory: %s\n", sourceDir)
		fmt.Printf("Output file: %s\n", outputPath)
		fmt.Printf("Compression level: %d\n", compressionLevel)
	}

	// Check if source directory exists
	if _, err := os.Stat(sourceDir); os.IsNotExist(err) {
		return fmt.Errorf("source directory does not exist: %s", sourceDir)
	}

	// Create ZIP container
	container := container.NewZIPContainer().
		SetCompressionLevel(compressionLevel).
		SetValidateStructure(validate)

	// Pack directory
	if err := container.CreateFromDirectory(sourceDir, outputPath); err != nil {
		return fmt.Errorf("failed to pack directory: %v", err)
	}

	// Show results
	if info, err := os.Stat(outputPath); err == nil {
		fmt.Printf("✓ Created %s (%d bytes)\n", outputPath, info.Size())
	}

	if verbose {
		// Show compression statistics
		if fileInfos, err := container.GetFileInfo(outputPath); err == nil {
			totalOriginal := int64(0)
			totalCompressed := int64(0)
			
			for _, info := range fileInfos {
				totalOriginal += info.Size
				totalCompressed += info.CompressedSize
			}
			
			if totalOriginal > 0 {
				ratio := float64(totalCompressed) / float64(totalOriginal)
				fmt.Printf("Compression: %d → %d bytes (%.1f%%)\n", 
					totalOriginal, totalCompressed, ratio*100)
			}
		}
	}

	return nil
}

func unpackFile(inputPath, targetDir string, verbose bool) error {
	if verbose {
		fmt.Printf("Unpacking file: %s\n", inputPath)
		fmt.Printf("Target directory: %s\n", targetDir)
	}

	// Check if input file exists
	if _, err := os.Stat(inputPath); os.IsNotExist(err) {
		return fmt.Errorf("input file does not exist: %s", inputPath)
	}

	// Create ZIP container
	container := container.NewZIPContainer()

	// Extract to directory
	if err := container.ExtractToDirectory(inputPath, targetDir); err != nil {
		return fmt.Errorf("failed to unpack file: %v", err)
	}

	fmt.Printf("✓ Extracted to %s\n", targetDir)

	if verbose {
		// Show extracted files
		if files, err := container.GetFileList(inputPath); err == nil {
			fmt.Printf("Extracted %d files:\n", len(files))
			for _, file := range files {
				fmt.Printf("  %s\n", file)
			}
		}
	}

	return nil
}

func listContents(inputPath string, verbose bool) error {
	// Check if input file exists
	if _, err := os.Stat(inputPath); os.IsNotExist(err) {
		return fmt.Errorf("input file does not exist: %s", inputPath)
	}

	// Create ZIP container
	container := container.NewZIPContainer()

	if verbose {
		// Show detailed file information
		fileInfos, err := container.GetFileInfo(inputPath)
		if err != nil {
			return fmt.Errorf("failed to get file info: %v", err)
		}

		fmt.Printf("Contents of %s:\n", inputPath)
		fmt.Printf("%-40s %10s %10s %8s %s\n", "Path", "Size", "Compressed", "Ratio", "Modified")
		fmt.Printf("%s\n", string(make([]byte, 80, 80)))

		totalSize := int64(0)
		totalCompressed := int64(0)

		for _, info := range fileInfos {
			ratio := info.CompressionRatio * 100
			fmt.Printf("%-40s %10d %10d %7.1f%% %s\n",
				truncatePath(info.Path, 40),
				info.Size,
				info.CompressedSize,
				ratio,
				info.Modified.Format("2006-01-02 15:04"))
			
			totalSize += info.Size
			totalCompressed += info.CompressedSize
		}

		fmt.Printf("%s\n", string(make([]byte, 80, 80)))
		overallRatio := float64(totalCompressed) / float64(totalSize) * 100
		fmt.Printf("%-40s %10d %10d %7.1f%%\n", 
			fmt.Sprintf("Total (%d files)", len(fileInfos)),
			totalSize, totalCompressed, overallRatio)

	} else {
		// Show simple file list
		files, err := container.GetFileList(inputPath)
		if err != nil {
			return fmt.Errorf("failed to get file list: %v", err)
		}

		fmt.Printf("Contents of %s (%d files):\n", inputPath, len(files))
		for _, file := range files {
			fmt.Printf("  %s\n", file)
		}
	}

	return nil
}

func validateFile(inputPath string, verbose bool) error {
	// Check if input file exists
	if _, err := os.Stat(inputPath); os.IsNotExist(err) {
		return fmt.Errorf("input file does not exist: %s", inputPath)
	}

	// Create ZIP container
	container := container.NewZIPContainer()

	// Validate structure
	result := container.ValidateStructure(inputPath)

	fmt.Printf("Validation Results for %s:\n", inputPath)

	if result.IsValid {
		fmt.Printf("✓ Status: VALID\n")
	} else {
		fmt.Printf("✗ Status: INVALID\n")
	}

	if len(result.Errors) > 0 {
		fmt.Printf("\nErrors (%d):\n", len(result.Errors))
		for i, err := range result.Errors {
			fmt.Printf("  %d. %s\n", i+1, err)
		}
	}

	if len(result.Warnings) > 0 && (verbose || !result.IsValid) {
		fmt.Printf("\nWarnings (%d):\n", len(result.Warnings))
		for i, warning := range result.Warnings {
			fmt.Printf("  %d. %s\n", i+1, warning)
		}
	}

	if verbose && result.IsValid {
		// Show additional validation details
		fmt.Printf("\nStructure Analysis:\n")
		
		files, err := container.GetFileList(inputPath)
		if err == nil {
			requiredFiles := []string{"manifest.json"}
			recommendedFiles := []string{"content/index.html", "content/static/fallback.html"}
			
			fmt.Printf("  Required files:\n")
			for _, required := range requiredFiles {
				found := false
				for _, file := range files {
					if file == required {
						found = true
						break
					}
				}
				if found {
					fmt.Printf("    ✓ %s\n", required)
				} else {
					fmt.Printf("    ✗ %s (missing)\n", required)
				}
			}
			
			fmt.Printf("  Recommended files:\n")
			for _, recommended := range recommendedFiles {
				found := false
				for _, file := range files {
					if file == recommended {
						found = true
						break
					}
				}
				if found {
					fmt.Printf("    ✓ %s\n", recommended)
				} else {
					fmt.Printf("    - %s (optional)\n", recommended)
				}
			}
		}
	}

	if !result.IsValid {
		return fmt.Errorf("validation failed")
	}

	return nil
}

func showInfo(inputPath string) error {
	// Check if input file exists
	fileInfo, err := os.Stat(inputPath)
	if os.IsNotExist(err) {
		return fmt.Errorf("input file does not exist: %s", inputPath)
	}

	// Create ZIP container
	container := container.NewZIPContainer()

	fmt.Printf("LIV File Information\n")
	fmt.Printf("====================\n\n")

	fmt.Printf("File: %s\n", inputPath)
	fmt.Printf("Size: %d bytes\n", fileInfo.Size())
	fmt.Printf("Modified: %s\n\n", fileInfo.ModTime().Format("2006-01-02 15:04:05"))

	// Get detailed file information
	fileInfos, err := container.GetFileInfo(inputPath)
	if err != nil {
		return fmt.Errorf("failed to get file info: %v", err)
	}

	// Calculate statistics
	totalFiles := len(fileInfos)
	totalOriginalSize := int64(0)
	totalCompressedSize := int64(0)
	
	fileTypes := make(map[string]int)
	
	for _, info := range fileInfos {
		totalOriginalSize += info.Size
		totalCompressedSize += info.CompressedSize
		
		ext := filepath.Ext(info.Path)
		if ext == "" {
			ext = "(no extension)"
		}
		fileTypes[ext]++
	}

	fmt.Printf("Archive Statistics:\n")
	fmt.Printf("  Files: %d\n", totalFiles)
	fmt.Printf("  Original size: %d bytes\n", totalOriginalSize)
	fmt.Printf("  Compressed size: %d bytes\n", totalCompressedSize)
	
	if totalOriginalSize > 0 {
		ratio := float64(totalCompressedSize) / float64(totalOriginalSize)
		savings := (1.0 - ratio) * 100
		fmt.Printf("  Compression ratio: %.1f%%\n", ratio*100)
		fmt.Printf("  Space savings: %.1f%%\n", savings)
	}

	fmt.Printf("\nFile Types:\n")
	for ext, count := range fileTypes {
		fmt.Printf("  %s: %d files\n", ext, count)
	}

	// Validate structure
	result := container.ValidateStructure(inputPath)
	fmt.Printf("\nStructure Validation:\n")
	if result.IsValid {
		fmt.Printf("  ✓ Valid LIV structure\n")
	} else {
		fmt.Printf("  ✗ Invalid structure (%d errors)\n", len(result.Errors))
	}
	
	if len(result.Warnings) > 0 {
		fmt.Printf("  ⚠ %d warnings\n", len(result.Warnings))
	}

	return nil
}

func truncatePath(path string, maxLen int) string {
	if len(path) <= maxLen {
		return path
	}
	
	if maxLen <= 3 {
		return path[:maxLen]
	}
	
	return "..." + path[len(path)-(maxLen-3):]
}