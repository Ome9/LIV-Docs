package main

import (
	"encoding/json"
	"fmt"
	"os"
	"strconv"
	"strings"

	"github.com/liv-format/liv/pkg/pdfops"
	"github.com/spf13/cobra"
)

func main() {
	// Initialize PDF operations
	pdfops.Init()

	rootCmd := &cobra.Command{
		Use:   "liv-pdf",
		Short: "LIV PDF Operations Tool",
		Long:  `Comprehensive PDF manipulation tool for LIV Format`,
	}

	rootCmd.AddCommand(extractTextCmd())
	rootCmd.AddCommand(mergeCmd())
	rootCmd.AddCommand(splitCmd())
	rootCmd.AddCommand(extractPagesCmd())
	rootCmd.AddCommand(rotateCmd())
	rootCmd.AddCommand(watermarkCmd())
	rootCmd.AddCommand(compressCmd())
	rootCmd.AddCommand(encryptCmd())
	rootCmd.AddCommand(infoCmd())
	rootCmd.AddCommand(setInfoCmd())
	rootCmd.AddCommand(convertToLIVCmd())

	if err := rootCmd.Execute(); err != nil {
		fmt.Fprintf(os.Stderr, "Error: %v\n", err)
		os.Exit(1)
	}
}

func extractTextCmd() *cobra.Command {
	return &cobra.Command{
		Use:   "extract-text [input.pdf]",
		Short: "Extract all text from PDF",
		Args:  cobra.ExactArgs(1),
		RunE: func(cmd *cobra.Command, args []string) error {
			ops, err := pdfops.New(args[0])
			if err != nil {
				return err
			}

			text, err := ops.ExtractText()
			if err != nil {
				return err
			}

			fmt.Println(text)
			return nil
		},
	}
}

func mergeCmd() *cobra.Command {
	var output string

	cmd := &cobra.Command{
		Use:   "merge [input1.pdf input2.pdf ...]",
		Short: "Merge multiple PDFs into one",
		Args:  cobra.MinimumNArgs(2),
		RunE: func(cmd *cobra.Command, args []string) error {
			if output == "" {
				output = "merged.pdf"
			}

			err := pdfops.MergePDFs(args, output)
			if err != nil {
				return err
			}

			fmt.Printf("Merged %d PDFs into %s\n", len(args), output)
			return nil
		},
	}

	cmd.Flags().StringVarP(&output, "output", "o", "", "Output file path")
	return cmd
}

func splitCmd() *cobra.Command {
	var outputDir string

	cmd := &cobra.Command{
		Use:   "split [input.pdf] [ranges]",
		Short: "Split PDF by page ranges (e.g., 1-3,4-6)",
		Args:  cobra.ExactArgs(2),
		RunE: func(cmd *cobra.Command, args []string) error {
			ops, err := pdfops.New(args[0])
			if err != nil {
				return err
			}

			if outputDir == "" {
				outputDir = "."
			}

			// Parse ranges
			ranges, err := parseRanges(args[1])
			if err != nil {
				return err
			}

			err = ops.SplitPDF(ranges, outputDir)
			if err != nil {
				return err
			}

			fmt.Printf("Split PDF into %d files in %s\n", len(ranges), outputDir)
			return nil
		},
	}

	cmd.Flags().StringVarP(&outputDir, "output-dir", "d", "", "Output directory")
	return cmd
}

func extractPagesCmd() *cobra.Command {
	var output string

	cmd := &cobra.Command{
		Use:   "extract-pages [input.pdf] [pages]",
		Short: "Extract specific pages (e.g., 1,3,5)",
		Args:  cobra.ExactArgs(2),
		RunE: func(cmd *cobra.Command, args []string) error {
			ops, err := pdfops.New(args[0])
			if err != nil {
				return err
			}

			if output == "" {
				output = "extracted.pdf"
			}

			// Parse page numbers
			pages, err := parsePageNumbers(args[1])
			if err != nil {
				return err
			}

			err = ops.ExtractPages(pages, output)
			if err != nil {
				return err
			}

			fmt.Printf("Extracted %d pages to %s\n", len(pages), output)
			return nil
		},
	}

	cmd.Flags().StringVarP(&output, "output", "o", "", "Output file path")
	return cmd
}

func rotateCmd() *cobra.Command {
	var output string
	var angle int

	cmd := &cobra.Command{
		Use:   "rotate [input.pdf] [pages]",
		Short: "Rotate specific pages (e.g., 1,3,5)",
		Args:  cobra.ExactArgs(2),
		RunE: func(cmd *cobra.Command, args []string) error {
			ops, err := pdfops.New(args[0])
			if err != nil {
				return err
			}

			if output == "" {
				output = "rotated.pdf"
			}

			// Parse page numbers
			pages, err := parsePageNumbers(args[1])
			if err != nil {
				return err
			}

			err = ops.RotatePages(pages, int64(angle), output)
			if err != nil {
				return err
			}

			fmt.Printf("Rotated %d pages by %d degrees to %s\n", len(pages), angle, output)
			return nil
		},
	}

	cmd.Flags().StringVarP(&output, "output", "o", "", "Output file path")
	cmd.Flags().IntVarP(&angle, "angle", "a", 90, "Rotation angle (90, 180, 270)")
	return cmd
}

func watermarkCmd() *cobra.Command {
	var output string
	var text string

	cmd := &cobra.Command{
		Use:   "watermark [input.pdf]",
		Short: "Add text watermark to PDF",
		Args:  cobra.ExactArgs(1),
		RunE: func(cmd *cobra.Command, args []string) error {
			ops, err := pdfops.New(args[0])
			if err != nil {
				return err
			}

			if output == "" {
				output = "watermarked.pdf"
			}

			if text == "" {
				return fmt.Errorf("watermark text is required")
			}

			err = ops.AddWatermark(text, output)
			if err != nil {
				return err
			}

			fmt.Printf("Added watermark to %s\n", output)
			return nil
		},
	}

	cmd.Flags().StringVarP(&output, "output", "o", "", "Output file path")
	cmd.Flags().StringVarP(&text, "text", "t", "", "Watermark text")
	cmd.MarkFlagRequired("text")
	return cmd
}

func compressCmd() *cobra.Command {
	var output string

	cmd := &cobra.Command{
		Use:   "compress [input.pdf]",
		Short: "Compress and optimize PDF",
		Args:  cobra.ExactArgs(1),
		RunE: func(cmd *cobra.Command, args []string) error {
			ops, err := pdfops.New(args[0])
			if err != nil {
				return err
			}

			if output == "" {
				output = "compressed.pdf"
			}

			err = ops.CompressPDF(output)
			if err != nil {
				return err
			}

			fmt.Printf("Compressed PDF to %s\n", output)
			return nil
		},
	}

	cmd.Flags().StringVarP(&output, "output", "o", "", "Output file path")
	return cmd
}

func encryptCmd() *cobra.Command {
	var output string
	var password string

	cmd := &cobra.Command{
		Use:   "encrypt [input.pdf]",
		Short: "Encrypt PDF with password",
		Args:  cobra.ExactArgs(1),
		RunE: func(cmd *cobra.Command, args []string) error {
			ops, err := pdfops.New(args[0])
			if err != nil {
				return err
			}

			if output == "" {
				output = "encrypted.pdf"
			}

			if password == "" {
				return fmt.Errorf("password is required")
			}

			err = ops.EncryptPDF(password, output)
			if err != nil {
				return err
			}

			fmt.Printf("Encrypted PDF to %s\n", output)
			return nil
		},
	}

	cmd.Flags().StringVarP(&output, "output", "o", "", "Output file path")
	cmd.Flags().StringVarP(&password, "password", "p", "", "Encryption password")
	cmd.MarkFlagRequired("password")
	return cmd
}

func infoCmd() *cobra.Command {
	var jsonOutput bool

	cmd := &cobra.Command{
		Use:   "info [input.pdf]",
		Short: "Get PDF document information",
		Args:  cobra.ExactArgs(1),
		RunE: func(cmd *cobra.Command, args []string) error {
			ops, err := pdfops.New(args[0])
			if err != nil {
				return err
			}

			info, err := ops.GetDocumentInfo()
			if err != nil {
				return err
			}

			if jsonOutput {
				data, err := json.MarshalIndent(info, "", "  ")
				if err != nil {
					return err
				}
				fmt.Println(string(data))
			} else {
				fmt.Println("PDF Document Information:")
				fmt.Println("-------------------------")
				for key, value := range info {
					fmt.Printf("%s: %s\n", key, value)
				}
			}

			return nil
		},
	}

	cmd.Flags().BoolVarP(&jsonOutput, "json", "j", false, "Output as JSON")
	return cmd
}

func setInfoCmd() *cobra.Command {
	var output string
	var title, author, subject, keywords string

	cmd := &cobra.Command{
		Use:   "set-info [input.pdf]",
		Short: "Set PDF metadata",
		Args:  cobra.ExactArgs(1),
		RunE: func(cmd *cobra.Command, args []string) error {
			ops, err := pdfops.New(args[0])
			if err != nil {
				return err
			}

			if output == "" {
				output = "updated.pdf"
			}

			info := make(map[string]string)
			if title != "" {
				info["title"] = title
			}
			if author != "" {
				info["author"] = author
			}
			if subject != "" {
				info["subject"] = subject
			}
			if keywords != "" {
				info["keywords"] = keywords
			}

			err = ops.SetDocumentInfo(info, output)
			if err != nil {
				return err
			}

			fmt.Printf("Updated PDF metadata in %s\n", output)
			return nil
		},
	}

	cmd.Flags().StringVarP(&output, "output", "o", "", "Output file path")
	cmd.Flags().StringVar(&title, "title", "", "Document title")
	cmd.Flags().StringVar(&author, "author", "", "Document author")
	cmd.Flags().StringVar(&subject, "subject", "", "Document subject")
	cmd.Flags().StringVar(&keywords, "keywords", "", "Document keywords")
	return cmd
}

func convertToLIVCmd() *cobra.Command {
	var output string

	cmd := &cobra.Command{
		Use:   "to-liv [input.pdf]",
		Short: "Convert PDF to LIV format",
		Args:  cobra.ExactArgs(1),
		RunE: func(cmd *cobra.Command, args []string) error {
			ops, err := pdfops.New(args[0])
			if err != nil {
				return err
			}

			if output == "" {
				output = strings.TrimSuffix(args[0], ".pdf") + ".liv"
			}

			err = ops.ConvertToLIV(output)
			if err != nil {
				return err
			}

			fmt.Printf("Converted PDF to LIV format: %s\n", output)
			return nil
		},
	}

	cmd.Flags().StringVarP(&output, "output", "o", "", "Output file path")
	return cmd
}

// Helper functions

func parseRanges(rangesStr string) ([][]int, error) {
	parts := strings.Split(rangesStr, ",")
	var ranges [][]int

	for _, part := range parts {
		rangeParts := strings.Split(strings.TrimSpace(part), "-")
		if len(rangeParts) != 2 {
			return nil, fmt.Errorf("invalid range format: %s", part)
		}

		start, err := strconv.Atoi(strings.TrimSpace(rangeParts[0]))
		if err != nil {
			return nil, fmt.Errorf("invalid start page: %s", rangeParts[0])
		}

		end, err := strconv.Atoi(strings.TrimSpace(rangeParts[1]))
		if err != nil {
			return nil, fmt.Errorf("invalid end page: %s", rangeParts[1])
		}

		ranges = append(ranges, []int{start, end})
	}

	return ranges, nil
}

func parsePageNumbers(pagesStr string) ([]int, error) {
	parts := strings.Split(pagesStr, ",")
	var pages []int

	for _, part := range parts {
		page, err := strconv.Atoi(strings.TrimSpace(part))
		if err != nil {
			return nil, fmt.Errorf("invalid page number: %s", part)
		}
		pages = append(pages, page)
	}

	return pages, nil
}
