package main

import (
	"context"
	"encoding/json"
	"fmt"
	"os"
	"strings"

	"github.com/liv-format/liv/pkg/container"
	"github.com/liv-format/liv/pkg/core"
	"github.com/liv-format/liv/pkg/integrity"
	"github.com/spf13/cobra"
)

func main() {
	var (
		verbose    bool
		keySize    int
		outputFile string
	)

	rootCmd := &cobra.Command{
		Use:   "liv-integrity",
		Short: "LIV Document Integrity and Signature Tool",
		Long: `LIV Integrity provides tools for verifying document integrity, 
generating and verifying digital signatures, and managing cryptographic keys 
for LIV documents.`,
	}

	// Hash command
	hashCmd := &cobra.Command{
		Use:   "hash [file-or-directory]",
		Short: "Calculate hashes for files or directories",
		Long:  "Calculate SHA-256 hashes for individual files or all files in a directory.",
		Args:  cobra.ExactArgs(1),
		RunE: func(cmd *cobra.Command, args []string) error {
			return calculateHashes(args[0], verbose)
		},
	}

	hashCmd.Flags().BoolVarP(&verbose, "verbose", "v", false, "Verbose output")

	// Verify command
	verifyCmd := &cobra.Command{
		Use:   "verify [liv-file]",
		Short: "Verify integrity of a LIV document",
		Long:  "Verify the integrity of all resources in a LIV document against their manifest hashes.",
		Args:  cobra.ExactArgs(1),
		RunE: func(cmd *cobra.Command, args []string) error {
			return verifyIntegrity(args[0], verbose)
		},
	}

	verifyCmd.Flags().BoolVarP(&verbose, "verbose", "v", false, "Detailed verification output")

	// Generate keys command
	generateKeysCmd := &cobra.Command{
		Use:   "generate-keys [key-name]",
		Short: "Generate RSA key pair for signing",
		Long:  "Generate a new RSA key pair for signing LIV documents.",
		Args:  cobra.ExactArgs(1),
		RunE: func(cmd *cobra.Command, args []string) error {
			return generateKeys(args[0], keySize, verbose)
		},
	}

	generateKeysCmd.Flags().IntVarP(&keySize, "key-size", "s", 2048, "RSA key size in bits")
	generateKeysCmd.Flags().BoolVarP(&verbose, "verbose", "v", false, "Verbose output")

	// Sign command
	signCmd := &cobra.Command{
		Use:   "sign [liv-file] [private-key]",
		Short: "Sign a LIV document",
		Long:  "Add digital signatures to a LIV document using a private key.",
		Args:  cobra.ExactArgs(2),
		RunE: func(cmd *cobra.Command, args []string) error {
			return signDocument(args[0], args[1], outputFile, verbose)
		},
	}

	signCmd.Flags().StringVarP(&outputFile, "output", "o", "", "Output file (default: overwrite input)")
	signCmd.Flags().BoolVarP(&verbose, "verbose", "v", false, "Verbose output")

	// Verify signature command
	verifySignatureCmd := &cobra.Command{
		Use:   "verify-signature [liv-file] [public-key]",
		Short: "Verify signatures in a LIV document",
		Long:  "Verify all digital signatures in a LIV document using a public key.",
		Args:  cobra.ExactArgs(2),
		RunE: func(cmd *cobra.Command, args []string) error {
			return verifySignatures(args[0], args[1], verbose)
		},
	}

	verifySignatureCmd.Flags().BoolVarP(&verbose, "verbose", "v", false, "Detailed verification output")

	// Report command
	reportCmd := &cobra.Command{
		Use:   "report [liv-file]",
		Short: "Generate integrity report for a LIV document",
		Long:  "Generate a comprehensive integrity report including hash verification and signature status.",
		Args:  cobra.ExactArgs(1),
		RunE: func(cmd *cobra.Command, args []string) error {
			return generateReport(args[0], verbose)
		},
	}

	reportCmd.Flags().BoolVarP(&verbose, "verbose", "v", false, "Include detailed information")

	// Add subcommands
	rootCmd.AddCommand(hashCmd)
	rootCmd.AddCommand(verifyCmd)
	rootCmd.AddCommand(generateKeysCmd)
	rootCmd.AddCommand(signCmd)
	rootCmd.AddCommand(verifySignatureCmd)
	rootCmd.AddCommand(reportCmd)

	if err := rootCmd.Execute(); err != nil {
		fmt.Fprintf(os.Stderr, "Error: %v\n", err)
		os.Exit(1)
	}
}

func calculateHashes(path string, verbose bool) error {
	hasher := integrity.NewResourceHasher(integrity.SHA256)

	info, err := os.Stat(path)
	if err != nil {
		return fmt.Errorf("failed to access path: %v", err)
	}

	if info.IsDir() {
		// Hash directory
		if verbose {
			fmt.Printf("Calculating hashes for directory: %s\n", path)
		}

		hashes, err := hasher.HashDirectory(path)
		if err != nil {
			return fmt.Errorf("failed to hash directory: %v", err)
		}

		fmt.Printf("Directory: %s (%d files)\n", path, len(hashes))
		fmt.Printf("%-50s %s\n", "File", "SHA-256 Hash")
		fmt.Printf("%s\n", strings.Repeat("-", 114))

		for filePath, hash := range hashes {
			fmt.Printf("%-50s %s\n", truncatePath(filePath, 50), hash)
		}

	} else {
		// Hash single file
		if verbose {
			fmt.Printf("Calculating hash for file: %s\n", path)
		}

		hash, err := hasher.HashFile(path)
		if err != nil {
			return fmt.Errorf("failed to hash file: %v", err)
		}

		fmt.Printf("File: %s\n", path)
		fmt.Printf("SHA-256: %s\n", hash)

		if verbose {
			fmt.Printf("Size: %d bytes\n", info.Size())
		}
	}

	return nil
}

func verifyIntegrity(livFile string, verbose bool) error {
	if verbose {
		fmt.Printf("Verifying integrity of: %s\n", livFile)
	}

	// Extract LIV file
	packageManager := container.NewPackageManager()
	file, err := os.Open(livFile)
	if err != nil {
		return fmt.Errorf("failed to open LIV file: %v", err)
	}
	defer file.Close()

	document, err := packageManager.ExtractPackage(context.TODO(), file)
	if err != nil {
		return fmt.Errorf("failed to extract LIV document: %v", err)
	}

	// Create integrity validator
	validator := integrity.NewIntegrityValidator()

	// Convert document back to files for validation
	files, err := documentToFiles(document)
	if err != nil {
		return fmt.Errorf("failed to convert document to files: %v", err)
	}

	// Validate resources
	result := validator.ValidateResources(document.Manifest.Resources, files)

	fmt.Printf("Integrity Verification Results\n")
	fmt.Printf("==============================\n\n")

	if result.IsValid {
		fmt.Printf("✓ Status: VALID\n")
	} else {
		fmt.Printf("✗ Status: INVALID\n")
	}

	fmt.Printf("Resources: %d\n", len(document.Manifest.Resources))

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

	// Validate WASM modules if present
	if len(document.WASMModules) > 0 {
		wasmResult := validator.ValidateWASMModules(document.Manifest.WASMConfig, document.WASMModules)
		fmt.Printf("\nWASM Module Validation:\n")
		if wasmResult.IsValid {
			fmt.Printf("✓ WASM modules valid (%d modules)\n", len(document.WASMModules))
		} else {
			fmt.Printf("✗ WASM module validation failed\n")
			for _, err := range wasmResult.Errors {
				fmt.Printf("  - %s\n", err)
			}
		}
	}

	if !result.IsValid {
		return fmt.Errorf("integrity verification failed")
	}

	return nil
}

func generateKeys(keyName string, keySize int, verbose bool) error {
	if verbose {
		fmt.Printf("Generating %d-bit RSA key pair: %s\n", keySize, keyName)
	}

	sm := integrity.NewSignatureManager()

	keyPair, err := sm.GenerateKeyPair(keySize)
	if err != nil {
		return fmt.Errorf("failed to generate key pair: %v", err)
	}

	// Save private key
	privateKeyFile := keyName + "-private.pem"
	if err := sm.SavePrivateKeyPEM(keyPair, privateKeyFile); err != nil {
		return fmt.Errorf("failed to save private key: %v", err)
	}

	// Save public key
	publicKeyFile := keyName + "-public.pem"
	if err := sm.SavePublicKeyPEM(keyPair, publicKeyFile); err != nil {
		return fmt.Errorf("failed to save public key: %v", err)
	}

	fmt.Printf("✓ Generated key pair:\n")
	fmt.Printf("  Private key: %s\n", privateKeyFile)
	fmt.Printf("  Public key:  %s\n", publicKeyFile)

	if verbose {
		info := sm.GetSignatureInfo(keyPair.PublicKey)
		fmt.Printf("\nKey Information:\n")
		fmt.Printf("  Algorithm: %s\n", info.Algorithm)
		fmt.Printf("  Key size:  %d bits\n", info.KeySize)
		fmt.Printf("  Fingerprint: %s\n", info.Fingerprint)
	}

	return nil
}

func signDocument(livFile, privateKeyFile, outputFile string, verbose bool) error {
	if verbose {
		fmt.Printf("Signing document: %s\n", livFile)
		fmt.Printf("Private key: %s\n", privateKeyFile)
	}

	// Load private key
	sm := integrity.NewSignatureManager()
	privateKey, err := sm.LoadPrivateKeyPEM(privateKeyFile)
	if err != nil {
		return fmt.Errorf("failed to load private key: %v", err)
	}

	// Extract LIV document
	packageManager := container.NewPackageManager()
	file, err := os.Open(livFile)
	if err != nil {
		return fmt.Errorf("failed to open LIV file: %v", err)
	}
	defer file.Close()

	document, err := packageManager.ExtractPackage(context.TODO(), file)
	if err != nil {
		return fmt.Errorf("failed to extract LIV document: %v", err)
	}

	// Sign document
	signatures, err := sm.SignDocument(document, privateKey)
	if err != nil {
		return fmt.Errorf("failed to sign document: %v", err)
	}

	document.Signatures = signatures

	// Determine output file
	if outputFile == "" {
		outputFile = livFile
	}

	// Save signed document
	if err := packageManager.SavePackage(document, outputFile); err != nil {
		return fmt.Errorf("failed to save signed document: %v", err)
	}

	fmt.Printf("✓ Document signed successfully\n")
	fmt.Printf("Output: %s\n", outputFile)

	if verbose {
		fmt.Printf("\nSignatures added:\n")
		fmt.Printf("  Manifest: %s\n", signatures.ManifestSignature[:16]+"...")
		if signatures.ContentSignature != "" {
			fmt.Printf("  Content:  %s\n", signatures.ContentSignature[:16]+"...")
		}
		if len(signatures.WASMSignatures) > 0 {
			fmt.Printf("  WASM modules: %d\n", len(signatures.WASMSignatures))
		}
	}

	return nil
}

func verifySignatures(livFile, publicKeyFile string, verbose bool) error {
	if verbose {
		fmt.Printf("Verifying signatures in: %s\n", livFile)
		fmt.Printf("Public key: %s\n", publicKeyFile)
	}

	// Load public key
	sm := integrity.NewSignatureManager()
	publicKey, err := sm.LoadPublicKeyPEM(publicKeyFile)
	if err != nil {
		return fmt.Errorf("failed to load public key: %v", err)
	}

	// Extract LIV document
	packageManager := container.NewPackageManager()
	file, err := os.Open(livFile)
	if err != nil {
		return fmt.Errorf("failed to open LIV file: %v", err)
	}
	defer file.Close()

	document, err := packageManager.ExtractPackage(context.TODO(), file)
	if err != nil {
		return fmt.Errorf("failed to extract LIV document: %v", err)
	}

	// Verify signatures
	result := sm.VerifyDocument(document, publicKey)

	fmt.Printf("Signature Verification Results\n")
	fmt.Printf("==============================\n\n")

	if result.Valid {
		fmt.Printf("✓ Status: VALID\n")
	} else {
		fmt.Printf("✗ Status: INVALID\n")
	}

	fmt.Printf("Manifest signature: ")
	if result.ManifestValid {
		fmt.Printf("✓ Valid\n")
	} else {
		fmt.Printf("✗ Invalid\n")
	}

	fmt.Printf("Content signature:  ")
	if result.ContentValid {
		fmt.Printf("✓ Valid\n")
	} else {
		fmt.Printf("✗ Invalid\n")
	}

	if len(result.WASMModulesValid) > 0 {
		fmt.Printf("WASM modules:\n")
		for moduleName, valid := range result.WASMModulesValid {
			if valid {
				fmt.Printf("  ✓ %s\n", moduleName)
			} else {
				fmt.Printf("  ✗ %s\n", moduleName)
			}
		}
	}

	if len(result.Errors) > 0 {
		fmt.Printf("\nErrors:\n")
		for _, err := range result.Errors {
			fmt.Printf("  - %s\n", err)
		}
	}

	if verbose {
		info := sm.GetSignatureInfo(publicKey)
		fmt.Printf("\nKey Information:\n")
		fmt.Printf("  Algorithm: %s\n", info.Algorithm)
		fmt.Printf("  Key size:  %d bits\n", info.KeySize)
		fmt.Printf("  Fingerprint: %s\n", info.Fingerprint)
		fmt.Printf("  Verified at: %s\n", result.VerificationTime.Format("2006-01-02 15:04:05"))
	}

	if !result.Valid {
		return fmt.Errorf("signature verification failed")
	}

	return nil
}

func generateReport(livFile string, verbose bool) error {
	// Extract LIV document
	packageManager := container.NewPackageManager()
	file, err := os.Open(livFile)
	if err != nil {
		return fmt.Errorf("failed to open LIV file: %v", err)
	}
	defer file.Close()

	document, err := packageManager.ExtractPackage(context.TODO(), file)
	if err != nil {
		return fmt.Errorf("failed to extract LIV document: %v", err)
	}

	// Create integrity validator
	validator := integrity.NewIntegrityValidator()

	// Generate comprehensive report
	report := validator.GenerateIntegrityReport(document.Manifest, nil, document.WASMModules)

	if verbose {
		// Output as JSON for detailed analysis
		reportJSON, err := json.MarshalIndent(report, "", "  ")
		if err != nil {
			return fmt.Errorf("failed to marshal report: %v", err)
		}
		fmt.Println(string(reportJSON))
	} else {
		// Human-readable summary
		fmt.Printf("LIV Document Integrity Report\n")
		fmt.Printf("=============================\n\n")

		fmt.Printf("File: %s\n", livFile)
		fmt.Printf("Overall Status: ")
		if report.Valid {
			fmt.Printf("✓ VALID\n")
		} else {
			fmt.Printf("✗ INVALID\n")
		}

		fmt.Printf("\nResource Summary:\n")
		fmt.Printf("  Total resources: %d\n", report.TotalResources)
		fmt.Printf("  Validated: %d\n", report.ValidatedResources)
		fmt.Printf("  Hash mismatches: %d\n", len(report.HashMismatches))
		fmt.Printf("  Size mismatches: %d\n", len(report.SizeMismatches))
		fmt.Printf("  Missing resources: %d\n", len(report.MissingResources))
		fmt.Printf("  Orphaned files: %d\n", len(report.OrphanedFiles))

		if len(report.HashMismatches) > 0 {
			fmt.Printf("\nHash Mismatches:\n")
			for _, mismatch := range report.HashMismatches {
				fmt.Printf("  %s: expected %s, got %s\n",
					mismatch.Path, mismatch.ExpectedHash[:16]+"...", mismatch.ActualHash[:16]+"...")
			}
		}

		if len(report.MissingResources) > 0 {
			fmt.Printf("\nMissing Resources:\n")
			for _, missing := range report.MissingResources {
				fmt.Printf("  %s\n", missing)
			}
		}

		if report.WASMValidation != nil {
			fmt.Printf("\nWASM Validation: ")
			if report.WASMValidation.IsValid {
				fmt.Printf("✓ Valid\n")
			} else {
				fmt.Printf("✗ Invalid (%d errors)\n", len(report.WASMValidation.Errors))
			}
		}
	}

	return nil
}

// Helper functions

func documentToFiles(document *core.LIVDocument) (map[string][]byte, error) {
	files := make(map[string][]byte)

	// Add content files
	if document.Content != nil {
		if document.Content.HTML != "" {
			files["content/index.html"] = []byte(document.Content.HTML)
		}
		if document.Content.CSS != "" {
			files["content/styles/main.css"] = []byte(document.Content.CSS)
		}
		if document.Content.InteractiveSpec != "" {
			files["content/scripts/main.js"] = []byte(document.Content.InteractiveSpec)
		}
		if document.Content.StaticFallback != "" {
			files["content/static/fallback.html"] = []byte(document.Content.StaticFallback)
		}
	}

	// Add assets
	if document.Assets != nil {
		for name, data := range document.Assets.Images {
			files["assets/images/"+name] = data
		}
		for name, data := range document.Assets.Fonts {
			files["assets/fonts/"+name] = data
		}
		for name, data := range document.Assets.Data {
			files["assets/data/"+name] = data
		}
	}

	return files, nil
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
