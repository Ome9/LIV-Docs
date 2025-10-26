package main

import (
	"fmt"
	"os"

	"github.com/liv-format/liv/internal/converter"
	"github.com/spf13/cobra"
)

var (
	version = "0.1.0"
	commit  = "dev"
	date    = "unknown"
)

func main() {
	rootCmd := &cobra.Command{
		Use:   "liv-converter",
		Short: "LIV Converter - Convert PDF files to LIV format",
		Long: `LIV Converter is a tool for transforming PDF documents into the LIV format.
It extracts text, images, and layout information from PDFs and creates
fully-structured LIV documents with manifest, assets, and metadata.`,
		Version: fmt.Sprintf("%s (commit: %s, built: %s)", version, commit, date),
	}

	// Add subcommands
	rootCmd.AddCommand(convertCmd())
	rootCmd.AddCommand(inspectCmd())
	rootCmd.AddCommand(validateCmd())

	if err := rootCmd.Execute(); err != nil {
		fmt.Fprintf(os.Stderr, "Error: %v\n", err)
		os.Exit(1)
	}
}

func convertCmd() *cobra.Command {
	var (
		output     string
		title      string
		author     string
		compress   bool
		dryRun     bool
		embedFonts bool
		quality    int
	)

	cmd := &cobra.Command{
		Use:   "convert [input.pdf]",
		Short: "Convert PDF to LIV format",
		Long: `Convert a PDF document to LIV format by extracting text, images,
and layout information, then packaging it into a .liv archive.`,
		Example: `  liv-converter convert document.pdf
  liv-converter convert document.pdf --output=mydoc.liv
  liv-converter convert document.pdf --dry-run
  liv-converter convert document.pdf --title="My Document" --author="John Doe"`,
		Args: cobra.ExactArgs(1),
		RunE: func(cmd *cobra.Command, args []string) error {
			inputPath := args[0]

			if output == "" {
				// Default output name
				output = inputPath[:len(inputPath)-4] + ".liv"
			}

			config := converter.ConvertConfig{
				InputPath:  inputPath,
				OutputPath: output,
				Title:      title,
				Author:     author,
				Compress:   compress,
				DryRun:     dryRun,
				EmbedFonts: embedFonts,
				Quality:    quality,
			}

			return converter.ConvertPDFToLIV(config)
		},
	}

	cmd.Flags().StringVarP(&output, "output", "o", "", "Output .liv file path")
	cmd.Flags().StringVarP(&title, "title", "t", "", "Document title (default: from PDF metadata)")
	cmd.Flags().StringVarP(&author, "author", "a", "", "Document author (default: from PDF metadata)")
	cmd.Flags().BoolVarP(&compress, "compress", "c", true, "Compress assets in .liv archive")
	cmd.Flags().BoolVarP(&dryRun, "dry-run", "d", false, "Output intermediate JSON without creating .liv file")
	cmd.Flags().BoolVarP(&embedFonts, "embed-fonts", "f", false, "Embed fonts in LIV document")
	cmd.Flags().IntVarP(&quality, "quality", "q", 85, "Image quality (1-100) for optimization")

	return cmd
}

func inspectCmd() *cobra.Command {
	var (
		showContent bool
		showAssets  bool
		jsonOutput  bool
	)

	cmd := &cobra.Command{
		Use:   "inspect [input.liv]",
		Short: "Inspect a LIV document",
		Long: `Inspect the contents of a LIV document, displaying manifest,
document structure, and asset information.`,
		Example: `  liv-converter inspect document.liv
  liv-converter inspect document.liv --show-content
  liv-converter inspect document.liv --json`,
		Args: cobra.ExactArgs(1),
		RunE: func(cmd *cobra.Command, args []string) error {
			config := converter.InspectConfig{
				InputPath:   args[0],
				ShowContent: showContent,
				ShowAssets:  showAssets,
				JSONOutput:  jsonOutput,
			}

			return converter.InspectLIV(config)
		},
	}

	cmd.Flags().BoolVarP(&showContent, "show-content", "c", false, "Show document content")
	cmd.Flags().BoolVarP(&showAssets, "show-assets", "a", false, "Show asset details")
	cmd.Flags().BoolVarP(&jsonOutput, "json", "j", false, "Output as JSON")

	return cmd
}

func validateCmd() *cobra.Command {
	var strict bool

	cmd := &cobra.Command{
		Use:   "validate [input.liv]",
		Short: "Validate a LIV document",
		Long: `Validate that a LIV document conforms to the specification,
checking manifest structure, document schema, and asset integrity.`,
		Example: `  liv-converter validate document.liv
  liv-converter validate document.liv --strict`,
		Args: cobra.ExactArgs(1),
		RunE: func(cmd *cobra.Command, args []string) error {
			config := converter.ValidateConfig{
				InputPath: args[0],
				Strict:    strict,
			}

			return converter.ValidateLIV(config)
		},
	}

	cmd.Flags().BoolVarP(&strict, "strict", "s", false, "Enable strict validation mode")

	return cmd
}
