package integrity

import (
	"os"
	"testing"
	"time"

	"github.com/liv-format/liv/pkg/core"
)

func TestSignatureManager_GenerateKeyPair(t *testing.T) {
	sm := NewSignatureManager()

	keyPair, err := sm.GenerateKeyPair(2048)
	if err != nil {
		t.Fatalf("Failed to generate key pair: %v", err)
	}

	if keyPair.PrivateKey == nil {
		t.Error("Private key is nil")
	}

	if keyPair.PublicKey == nil {
		t.Error("Public key is nil")
	}

	// Verify key size
	expectedKeySize := 2048 / 8 // Convert bits to bytes
	if keyPair.PrivateKey.Size() != expectedKeySize {
		t.Errorf("Expected key size %d bytes, got %d", expectedKeySize, keyPair.PrivateKey.Size())
	}

	// Test minimum key size enforcement
	smallKeyPair, err := sm.GenerateKeyPair(1024)
	if err != nil {
		t.Fatalf("Failed to generate small key pair: %v", err)
	}

	// Should be upgraded to 2048
	if smallKeyPair.PrivateKey.Size() != expectedKeySize {
		t.Errorf("Small key should be upgraded to %d bytes, got %d", expectedKeySize, smallKeyPair.PrivateKey.Size())
	}
}

func TestSignatureManager_SaveAndLoadKeys(t *testing.T) {
	sm := NewSignatureManager()

	// Generate key pair
	keyPair, err := sm.GenerateKeyPair(2048)
	if err != nil {
		t.Fatalf("Failed to generate key pair: %v", err)
	}

	// Create temporary files
	privateKeyFile, err := os.CreateTemp("", "private-*.pem")
	if err != nil {
		t.Fatalf("Failed to create temp private key file: %v", err)
	}
	defer os.Remove(privateKeyFile.Name())
	privateKeyFile.Close()

	publicKeyFile, err := os.CreateTemp("", "public-*.pem")
	if err != nil {
		t.Fatalf("Failed to create temp public key file: %v", err)
	}
	defer os.Remove(publicKeyFile.Name())
	publicKeyFile.Close()

	// Save keys
	if err := sm.SavePrivateKeyPEM(keyPair, privateKeyFile.Name()); err != nil {
		t.Fatalf("Failed to save private key: %v", err)
	}

	if err := sm.SavePublicKeyPEM(keyPair, publicKeyFile.Name()); err != nil {
		t.Fatalf("Failed to save public key: %v", err)
	}

	// Load keys
	loadedPrivateKey, err := sm.LoadPrivateKeyPEM(privateKeyFile.Name())
	if err != nil {
		t.Fatalf("Failed to load private key: %v", err)
	}

	loadedPublicKey, err := sm.LoadPublicKeyPEM(publicKeyFile.Name())
	if err != nil {
		t.Fatalf("Failed to load public key: %v", err)
	}

	// Verify loaded keys match original
	if loadedPrivateKey.Size() != keyPair.PrivateKey.Size() {
		t.Error("Loaded private key size doesn't match original")
	}

	if loadedPublicKey.Size() != keyPair.PublicKey.Size() {
		t.Error("Loaded public key size doesn't match original")
	}

	// Test signing with loaded keys
	testData := []byte("Test data for loaded keys")
	signature, err := sm.SignData(testData, loadedPrivateKey)
	if err != nil {
		t.Fatalf("Failed to sign with loaded private key: %v", err)
	}

	valid, err := sm.VerifySignature(testData, signature, loadedPublicKey)
	if err != nil {
		t.Fatalf("Failed to verify with loaded public key: %v", err)
	}

	if !valid {
		t.Error("Signature verification failed with loaded keys")
	}
}

func TestSignatureManager_SignAndVerifyData(t *testing.T) {
	sm := NewSignatureManager()

	keyPair, err := sm.GenerateKeyPair(2048)
	if err != nil {
		t.Fatalf("Failed to generate key pair: %v", err)
	}

	testData := []byte("This is test data for signing")

	// Sign data
	signature, err := sm.SignData(testData, keyPair.PrivateKey)
	if err != nil {
		t.Fatalf("Failed to sign data: %v", err)
	}

	if signature == "" {
		t.Error("Signature is empty")
	}

	// Verify signature
	valid, err := sm.VerifySignature(testData, signature, keyPair.PublicKey)
	if err != nil {
		t.Fatalf("Failed to verify signature: %v", err)
	}

	if !valid {
		t.Error("Signature verification failed")
	}

	// Test with different data (should fail)
	differentData := []byte("Different test data")
	valid, err = sm.VerifySignature(differentData, signature, keyPair.PublicKey)
	if err != nil {
		t.Fatalf("Failed to verify signature with different data: %v", err)
	}

	if valid {
		t.Error("Signature verification should have failed with different data")
	}

	// Test with invalid signature
	invalidSignature := "invalid_signature"
	valid, err = sm.VerifySignature(testData, invalidSignature, keyPair.PublicKey)
	if err == nil {
		t.Error("Expected error for invalid signature format")
	}
}

func TestSignatureManager_SignAndVerifyManifest(t *testing.T) {
	sm := NewSignatureManager()

	keyPair, err := sm.GenerateKeyPair(2048)
	if err != nil {
		t.Fatalf("Failed to generate key pair: %v", err)
	}

	// Create test manifest
	manifest := &core.Manifest{
		Version: "1.0",
		Metadata: &core.DocumentMetadata{
			Title:       "Test Document",
			Author:      "Test Author",
			Created:     time.Now().Add(-time.Hour),
			Modified:    time.Now(),
			Description: "Test manifest for signing",
			Version:     "1.0.0",
			Language:    "en",
		},
	}

	// Sign manifest
	signature, err := sm.SignManifest(manifest, keyPair.PrivateKey)
	if err != nil {
		t.Fatalf("Failed to sign manifest: %v", err)
	}

	// Verify manifest signature
	valid, err := sm.VerifyManifestSignature(manifest, signature, keyPair.PublicKey)
	if err != nil {
		t.Fatalf("Failed to verify manifest signature: %v", err)
	}

	if !valid {
		t.Error("Manifest signature verification failed")
	}

	// Test with modified manifest (should fail)
	modifiedManifest := *manifest
	modifiedManifest.Metadata.Title = "Modified Title"

	valid, err = sm.VerifyManifestSignature(&modifiedManifest, signature, keyPair.PublicKey)
	if err != nil {
		t.Fatalf("Failed to verify modified manifest signature: %v", err)
	}

	if valid {
		t.Error("Modified manifest signature verification should have failed")
	}
}

func TestSignatureManager_SignAndVerifyContent(t *testing.T) {
	sm := NewSignatureManager()

	keyPair, err := sm.GenerateKeyPair(2048)
	if err != nil {
		t.Fatalf("Failed to generate key pair: %v", err)
	}

	// Create test content
	content := &core.DocumentContent{
		HTML:           "<html><body>Test</body></html>",
		CSS:            "body { color: red; }",
		InteractiveSpec: "console.log('test');",
		StaticFallback: "<html><body>Static</body></html>",
	}

	// Sign content
	signature, err := sm.SignContent(content, keyPair.PrivateKey)
	if err != nil {
		t.Fatalf("Failed to sign content: %v", err)
	}

	// Verify content signature
	valid, err := sm.VerifyContentSignature(content, signature, keyPair.PublicKey)
	if err != nil {
		t.Fatalf("Failed to verify content signature: %v", err)
	}

	if !valid {
		t.Error("Content signature verification failed")
	}

	// Test with modified content (should fail)
	modifiedContent := *content
	modifiedContent.HTML = "<html><body>Modified</body></html>"

	valid, err = sm.VerifyContentSignature(&modifiedContent, signature, keyPair.PublicKey)
	if err != nil {
		t.Fatalf("Failed to verify modified content signature: %v", err)
	}

	if valid {
		t.Error("Modified content signature verification should have failed")
	}
}

func TestSignatureManager_SignAndVerifyWASMModule(t *testing.T) {
	sm := NewSignatureManager()

	keyPair, err := sm.GenerateKeyPair(2048)
	if err != nil {
		t.Fatalf("Failed to generate key pair: %v", err)
	}

	// Create test WASM module data
	wasmData := []byte{0x00, 0x61, 0x73, 0x6D, 0x01, 0x00, 0x00, 0x00, 0x01, 0x02, 0x03}

	// Sign WASM module
	signature, err := sm.SignWASMModule(wasmData, keyPair.PrivateKey)
	if err != nil {
		t.Fatalf("Failed to sign WASM module: %v", err)
	}

	// Verify WASM module signature
	valid, err := sm.VerifyWASMModuleSignature(wasmData, signature, keyPair.PublicKey)
	if err != nil {
		t.Fatalf("Failed to verify WASM module signature: %v", err)
	}

	if !valid {
		t.Error("WASM module signature verification failed")
	}

	// Test with modified WASM data (should fail)
	modifiedWASMData := []byte{0x00, 0x61, 0x73, 0x6D, 0x01, 0x00, 0x00, 0x00, 0xFF, 0xFF, 0xFF}

	valid, err = sm.VerifyWASMModuleSignature(modifiedWASMData, signature, keyPair.PublicKey)
	if err != nil {
		t.Fatalf("Failed to verify modified WASM module signature: %v", err)
	}

	if valid {
		t.Error("Modified WASM module signature verification should have failed")
	}
}

func TestSignatureManager_SignAndVerifyDocument(t *testing.T) {
	sm := NewSignatureManager()

	keyPair, err := sm.GenerateKeyPair(2048)
	if err != nil {
		t.Fatalf("Failed to generate key pair: %v", err)
	}

	// Create test document
	document := &core.LIVDocument{
		Manifest: &core.Manifest{
			Version: "1.0",
			Metadata: &core.DocumentMetadata{
				Title:       "Test Document",
				Author:      "Test Author",
				Created:     time.Now().Add(-time.Hour),
				Modified:    time.Now(),
				Description: "Test document for signing",
				Version:     "1.0.0",
				Language:    "en",
			},
		},
		Content: &core.DocumentContent{
			HTML: "<html><body>Test</body></html>",
			CSS:  "body { color: red; }",
		},
		WASMModules: map[string][]byte{
			"test-module": {0x00, 0x61, 0x73, 0x6D, 0x01, 0x00, 0x00, 0x00},
		},
	}

	// Sign document
	signatures, err := sm.SignDocument(document, keyPair.PrivateKey)
	if err != nil {
		t.Fatalf("Failed to sign document: %v", err)
	}

	// Set signatures in document
	document.Signatures = signatures

	// Verify document
	result := sm.VerifyDocument(document, keyPair.PublicKey)

	if !result.Valid {
		t.Errorf("Document verification failed: %v", result.Errors)
	}

	if !result.ManifestValid {
		t.Error("Manifest signature verification failed")
	}

	if !result.ContentValid {
		t.Error("Content signature verification failed")
	}

	if len(result.WASMModulesValid) != 1 {
		t.Errorf("Expected 1 WASM module verification result, got %d", len(result.WASMModulesValid))
	}

	if !result.WASMModulesValid["test-module"] {
		t.Error("WASM module signature verification failed")
	}

	// Test with corrupted signatures
	corruptedDocument := *document
	corruptedDocument.Signatures = &core.SignatureBundle{
		ManifestSignature: "corrupted_signature",
		ContentSignature:  signatures.ContentSignature,
		WASMSignatures:    signatures.WASMSignatures,
	}

	result = sm.VerifyDocument(&corruptedDocument, keyPair.PublicKey)
	if result.Valid {
		t.Error("Document verification should have failed with corrupted manifest signature")
	}
}

func TestSignatureManager_TrustChain(t *testing.T) {
	sm := NewSignatureManager()

	// Generate multiple key pairs
	keyPair1, err := sm.GenerateKeyPair(2048)
	if err != nil {
		t.Fatalf("Failed to generate key pair 1: %v", err)
	}

	keyPair2, err := sm.GenerateKeyPair(2048)
	if err != nil {
		t.Fatalf("Failed to generate key pair 2: %v", err)
	}

	// Create trust chain with first key
	trustChain := NewTrustChain()
	trustChain.AddTrustedPublicKey(keyPair1.PublicKey)

	// Create and sign document with first key
	document := &core.LIVDocument{
		Manifest: &core.Manifest{
			Version: "1.0",
			Metadata: &core.DocumentMetadata{
				Title:    "Trust Chain Test",
				Author:   "Test Author",
				Created:  time.Now().Add(-time.Hour),
				Modified: time.Now(),
				Version:  "1.0.0",
				Language: "en",
			},
		},
		Content: &core.DocumentContent{
			HTML: "<html><body>Test</body></html>",
		},
	}

	signatures, err := sm.SignDocument(document, keyPair1.PrivateKey)
	if err != nil {
		t.Fatalf("Failed to sign document: %v", err)
	}
	document.Signatures = signatures

	// Verify with trust chain (should succeed)
	result := sm.VerifyWithTrustChain(document, trustChain)
	if !result.Valid {
		t.Error("Trust chain verification should have succeeded")
	}

	// Sign document with second key (not in trust chain)
	signatures2, err := sm.SignDocument(document, keyPair2.PrivateKey)
	if err != nil {
		t.Fatalf("Failed to sign document with second key: %v", err)
	}
	document.Signatures = signatures2

	// Verify with trust chain (should fail)
	result = sm.VerifyWithTrustChain(document, trustChain)
	if result.Valid {
		t.Error("Trust chain verification should have failed for untrusted key")
	}

	// Add second key to trust chain
	trustChain.AddTrustedPublicKey(keyPair2.PublicKey)

	// Verify again (should succeed now)
	result = sm.VerifyWithTrustChain(document, trustChain)
	if !result.Valid {
		t.Error("Trust chain verification should have succeeded after adding key")
	}
}

func TestSignatureManager_GetSignatureInfo(t *testing.T) {
	sm := NewSignatureManager()

	keyPair, err := sm.GenerateKeyPair(2048)
	if err != nil {
		t.Fatalf("Failed to generate key pair: %v", err)
	}

	info := sm.GetSignatureInfo(keyPair.PublicKey)

	if info.Algorithm != "RSA-SHA256" {
		t.Errorf("Expected algorithm RSA-SHA256, got %s", info.Algorithm)
	}

	if info.KeySize != 2048 {
		t.Errorf("Expected key size 2048, got %d", info.KeySize)
	}

	if info.Fingerprint == "" {
		t.Error("Fingerprint should not be empty")
	}

	if len(info.Fingerprint) != 16 {
		t.Errorf("Expected fingerprint length 16, got %d", len(info.Fingerprint))
	}

	// Test that same key produces same fingerprint
	info2 := sm.GetSignatureInfo(keyPair.PublicKey)
	if info.Fingerprint != info2.Fingerprint {
		t.Error("Same key should produce same fingerprint")
	}

	// Test that different key produces different fingerprint
	keyPair2, err := sm.GenerateKeyPair(2048)
	if err != nil {
		t.Fatalf("Failed to generate second key pair: %v", err)
	}

	info3 := sm.GetSignatureInfo(keyPair2.PublicKey)
	if info.Fingerprint == info3.Fingerprint {
		t.Error("Different keys should produce different fingerprints")
	}
}

func BenchmarkSignatureManager_SignData(b *testing.B) {
	sm := NewSignatureManager()

	keyPair, err := sm.GenerateKeyPair(2048)
	if err != nil {
		b.Fatalf("Failed to generate key pair: %v", err)
	}

	testData := []byte("Benchmark data for signing performance test")

	b.ResetTimer()
	for i := 0; i < b.N; i++ {
		_, err := sm.SignData(testData, keyPair.PrivateKey)
		if err != nil {
			b.Fatalf("Failed to sign data: %v", err)
		}
	}
}

func BenchmarkSignatureManager_VerifySignature(b *testing.B) {
	sm := NewSignatureManager()

	keyPair, err := sm.GenerateKeyPair(2048)
	if err != nil {
		b.Fatalf("Failed to generate key pair: %v", err)
	}

	testData := []byte("Benchmark data for verification performance test")
	signature, err := sm.SignData(testData, keyPair.PrivateKey)
	if err != nil {
		b.Fatalf("Failed to sign data: %v", err)
	}

	b.ResetTimer()
	for i := 0; i < b.N; i++ {
		_, err := sm.VerifySignature(testData, signature, keyPair.PublicKey)
		if err != nil {
			b.Fatalf("Failed to verify signature: %v", err)
		}
	}
}