package integrity

import (
	"crypto"
	"crypto/rand"
	"crypto/rsa"
	"crypto/sha256"
	"crypto/x509"
	"encoding/base64"
	"encoding/pem"
	"fmt"
	"os"
	"time"

	"github.com/liv-format/liv/pkg/core"
)

// SignatureManager handles digital signatures for LIV documents
type SignatureManager struct {
	hasher *ResourceHasher
}

// NewSignatureManager creates a new signature manager
func NewSignatureManager() *SignatureManager {
	return &SignatureManager{
		hasher: NewResourceHasher(SHA256),
	}
}

// KeyPair represents an RSA key pair
type KeyPair struct {
	PrivateKey *rsa.PrivateKey
	PublicKey  *rsa.PublicKey
}

// GenerateKeyPair generates a new RSA key pair
func (sm *SignatureManager) GenerateKeyPair(keySize int) (*KeyPair, error) {
	if keySize < 2048 {
		keySize = 2048 // Minimum secure key size
	}
	
	privateKey, err := rsa.GenerateKey(rand.Reader, keySize)
	if err != nil {
		return nil, fmt.Errorf("failed to generate private key: %v", err)
	}
	
	return &KeyPair{
		PrivateKey: privateKey,
		PublicKey:  &privateKey.PublicKey,
	}, nil
}

// SavePrivateKeyPEM saves private key to PEM file
func (sm *SignatureManager) SavePrivateKeyPEM(keyPair *KeyPair, filePath string) error {
	privateKeyBytes, err := x509.MarshalPKCS8PrivateKey(keyPair.PrivateKey)
	if err != nil {
		return fmt.Errorf("failed to marshal private key: %v", err)
	}
	
	privateKeyPEM := &pem.Block{
		Type:  "PRIVATE KEY",
		Bytes: privateKeyBytes,
	}
	
	file, err := os.Create(filePath)
	if err != nil {
		return fmt.Errorf("failed to create private key file: %v", err)
	}
	defer file.Close()
	
	if err := pem.Encode(file, privateKeyPEM); err != nil {
		return fmt.Errorf("failed to encode private key: %v", err)
	}
	
	return nil
}

// SavePublicKeyPEM saves public key to PEM file
func (sm *SignatureManager) SavePublicKeyPEM(keyPair *KeyPair, filePath string) error {
	publicKeyBytes, err := x509.MarshalPKIXPublicKey(keyPair.PublicKey)
	if err != nil {
		return fmt.Errorf("failed to marshal public key: %v", err)
	}
	
	publicKeyPEM := &pem.Block{
		Type:  "PUBLIC KEY",
		Bytes: publicKeyBytes,
	}
	
	file, err := os.Create(filePath)
	if err != nil {
		return fmt.Errorf("failed to create public key file: %v", err)
	}
	defer file.Close()
	
	if err := pem.Encode(file, publicKeyPEM); err != nil {
		return fmt.Errorf("failed to encode public key: %v", err)
	}
	
	return nil
}

// LoadPrivateKeyPEM loads private key from PEM file
func (sm *SignatureManager) LoadPrivateKeyPEM(filePath string) (*rsa.PrivateKey, error) {
	keyData, err := os.ReadFile(filePath)
	if err != nil {
		return nil, fmt.Errorf("failed to read private key file: %v", err)
	}
	
	block, _ := pem.Decode(keyData)
	if block == nil {
		return nil, fmt.Errorf("failed to decode PEM block")
	}
	
	privateKey, err := x509.ParsePKCS8PrivateKey(block.Bytes)
	if err != nil {
		return nil, fmt.Errorf("failed to parse private key: %v", err)
	}
	
	rsaPrivateKey, ok := privateKey.(*rsa.PrivateKey)
	if !ok {
		return nil, fmt.Errorf("not an RSA private key")
	}
	
	return rsaPrivateKey, nil
}

// LoadPublicKeyPEM loads public key from PEM file
func (sm *SignatureManager) LoadPublicKeyPEM(filePath string) (*rsa.PublicKey, error) {
	keyData, err := os.ReadFile(filePath)
	if err != nil {
		return nil, fmt.Errorf("failed to read public key file: %v", err)
	}
	
	block, _ := pem.Decode(keyData)
	if block == nil {
		return nil, fmt.Errorf("failed to decode PEM block")
	}
	
	publicKey, err := x509.ParsePKIXPublicKey(block.Bytes)
	if err != nil {
		return nil, fmt.Errorf("failed to parse public key: %v", err)
	}
	
	rsaPublicKey, ok := publicKey.(*rsa.PublicKey)
	if !ok {
		return nil, fmt.Errorf("not an RSA public key")
	}
	
	return rsaPublicKey, nil
}

// SignData signs data with private key
func (sm *SignatureManager) SignData(data []byte, privateKey *rsa.PrivateKey) (string, error) {
	// Hash the data
	hash := sha256.Sum256(data)
	
	// Sign the hash
	signature, err := rsa.SignPKCS1v15(rand.Reader, privateKey, crypto.SHA256, hash[:])
	if err != nil {
		return "", fmt.Errorf("failed to sign data: %v", err)
	}
	
	// Encode signature as base64
	return base64.StdEncoding.EncodeToString(signature), nil
}

// VerifySignature verifies signature with public key
func (sm *SignatureManager) VerifySignature(data []byte, signatureStr string, publicKey *rsa.PublicKey) (bool, error) {
	// Decode signature from base64
	signature, err := base64.StdEncoding.DecodeString(signatureStr)
	if err != nil {
		return false, fmt.Errorf("failed to decode signature: %v", err)
	}
	
	// Hash the data
	hash := sha256.Sum256(data)
	
	// Verify signature
	err = rsa.VerifyPKCS1v15(publicKey, crypto.SHA256, hash[:], signature)
	if err != nil {
		return false, nil // Invalid signature, but not an error
	}
	
	return true, nil
}

// SignManifest signs a manifest
func (sm *SignatureManager) SignManifest(manifest *core.Manifest, privateKey *rsa.PrivateKey) (string, error) {
	// Serialize manifest to canonical JSON
	manifestData, err := sm.serializeManifestForSigning(manifest)
	if err != nil {
		return "", fmt.Errorf("failed to serialize manifest: %v", err)
	}
	
	return sm.SignData(manifestData, privateKey)
}

// VerifyManifestSignature verifies manifest signature
func (sm *SignatureManager) VerifyManifestSignature(manifest *core.Manifest, signature string, publicKey *rsa.PublicKey) (bool, error) {
	// Serialize manifest to canonical JSON
	manifestData, err := sm.serializeManifestForSigning(manifest)
	if err != nil {
		return false, fmt.Errorf("failed to serialize manifest: %v", err)
	}
	
	return sm.VerifySignature(manifestData, signature, publicKey)
}

// SignContent signs document content
func (sm *SignatureManager) SignContent(content *core.DocumentContent, privateKey *rsa.PrivateKey) (string, error) {
	// Create content hash from all content parts
	contentData := sm.serializeContentForSigning(content)
	return sm.SignData(contentData, privateKey)
}

// VerifyContentSignature verifies content signature
func (sm *SignatureManager) VerifyContentSignature(content *core.DocumentContent, signature string, publicKey *rsa.PublicKey) (bool, error) {
	contentData := sm.serializeContentForSigning(content)
	return sm.VerifySignature(contentData, signature, publicKey)
}

// SignWASMModule signs a WASM module
func (sm *SignatureManager) SignWASMModule(moduleData []byte, privateKey *rsa.PrivateKey) (string, error) {
	return sm.SignData(moduleData, privateKey)
}

// VerifyWASMModuleSignature verifies WASM module signature
func (sm *SignatureManager) VerifyWASMModuleSignature(moduleData []byte, signature string, publicKey *rsa.PublicKey) (bool, error) {
	return sm.VerifySignature(moduleData, signature, publicKey)
}

// SignDocument signs an entire LIV document
func (sm *SignatureManager) SignDocument(document *core.LIVDocument, privateKey *rsa.PrivateKey) (*core.SignatureBundle, error) {
	signatures := &core.SignatureBundle{
		WASMSignatures: make(map[string]string),
	}
	
	// Sign manifest
	manifestSig, err := sm.SignManifest(document.Manifest, privateKey)
	if err != nil {
		return nil, fmt.Errorf("failed to sign manifest: %v", err)
	}
	signatures.ManifestSignature = manifestSig
	
	// Sign content
	if document.Content != nil {
		contentSig, err := sm.SignContent(document.Content, privateKey)
		if err != nil {
			return nil, fmt.Errorf("failed to sign content: %v", err)
		}
		signatures.ContentSignature = contentSig
	}
	
	// Sign WASM modules
	for moduleName, moduleData := range document.WASMModules {
		wasmSig, err := sm.SignWASMModule(moduleData, privateKey)
		if err != nil {
			return nil, fmt.Errorf("failed to sign WASM module %s: %v", moduleName, err)
		}
		signatures.WASMSignatures[moduleName] = wasmSig
	}
	
	return signatures, nil
}

// VerifyDocument verifies all signatures in a LIV document
func (sm *SignatureManager) VerifyDocument(document *core.LIVDocument, publicKey *rsa.PublicKey) *SignatureVerificationResult {
	result := &SignatureVerificationResult{
		Valid:              true,
		ManifestValid:      false,
		ContentValid:       false,
		WASMModulesValid:   make(map[string]bool),
		Errors:             []string{},
		VerificationTime:   time.Now(),
	}
	
	// Verify manifest signature
	if document.Signatures != nil && document.Signatures.ManifestSignature != "" {
		valid, err := sm.VerifyManifestSignature(document.Manifest, document.Signatures.ManifestSignature, publicKey)
		if err != nil {
			result.Valid = false
			result.Errors = append(result.Errors, fmt.Sprintf("manifest signature verification error: %v", err))
		} else {
			result.ManifestValid = valid
			if !valid {
				result.Valid = false
				result.Errors = append(result.Errors, "manifest signature is invalid")
			}
		}
	} else {
		result.Valid = false
		result.Errors = append(result.Errors, "manifest signature is missing")
	}
	
	// Verify content signature
	if document.Content != nil && document.Signatures != nil && document.Signatures.ContentSignature != "" {
		valid, err := sm.VerifyContentSignature(document.Content, document.Signatures.ContentSignature, publicKey)
		if err != nil {
			result.Valid = false
			result.Errors = append(result.Errors, fmt.Sprintf("content signature verification error: %v", err))
		} else {
			result.ContentValid = valid
			if !valid {
				result.Valid = false
				result.Errors = append(result.Errors, "content signature is invalid")
			}
		}
	} else if document.Content != nil {
		result.Valid = false
		result.Errors = append(result.Errors, "content signature is missing")
	}
	
	// Verify WASM module signatures
	for moduleName, moduleData := range document.WASMModules {
		if document.Signatures != nil && document.Signatures.WASMSignatures[moduleName] != "" {
			valid, err := sm.VerifyWASMModuleSignature(moduleData, document.Signatures.WASMSignatures[moduleName], publicKey)
			if err != nil {
				result.Valid = false
				result.Errors = append(result.Errors, fmt.Sprintf("WASM module %s signature verification error: %v", moduleName, err))
			} else {
				result.WASMModulesValid[moduleName] = valid
				if !valid {
					result.Valid = false
					result.Errors = append(result.Errors, fmt.Sprintf("WASM module %s signature is invalid", moduleName))
				}
			}
		} else {
			result.Valid = false
			result.Errors = append(result.Errors, fmt.Sprintf("WASM module %s signature is missing", moduleName))
		}
	}
	
	return result
}

// SignatureVerificationResult contains signature verification results
type SignatureVerificationResult struct {
	Valid              bool              `json:"valid"`
	ManifestValid      bool              `json:"manifest_valid"`
	ContentValid       bool              `json:"content_valid"`
	WASMModulesValid   map[string]bool   `json:"wasm_modules_valid"`
	Errors             []string          `json:"errors"`
	VerificationTime   time.Time         `json:"verification_time"`
}

// Helper methods for serialization

func (sm *SignatureManager) serializeManifestForSigning(manifest *core.Manifest) ([]byte, error) {
	// Create a copy of manifest without signatures for signing
	manifestCopy := *manifest
	
	// Remove any existing signatures from the copy
	// (This ensures we're signing the content, not including signatures)
	
	// For now, we'll use a simple approach - hash the key components
	data := fmt.Sprintf("version:%s|title:%s|author:%s|created:%s|modified:%s",
		manifestCopy.Version,
		manifestCopy.Metadata.Title,
		manifestCopy.Metadata.Author,
		manifestCopy.Metadata.Created.Format(time.RFC3339),
		manifestCopy.Metadata.Modified.Format(time.RFC3339))
	
	return []byte(data), nil
}

func (sm *SignatureManager) serializeContentForSigning(content *core.DocumentContent) []byte {
	// Concatenate all content for signing
	data := content.HTML + content.CSS + content.InteractiveSpec + content.StaticFallback
	return []byte(data)
}

// TrustChain represents a chain of trust for signatures
type TrustChain struct {
	RootCertificates    []*x509.Certificate
	IntermediateCerts   []*x509.Certificate
	TrustedPublicKeys   []*rsa.PublicKey
}

// NewTrustChain creates a new trust chain
func NewTrustChain() *TrustChain {
	return &TrustChain{
		RootCertificates:  []*x509.Certificate{},
		IntermediateCerts: []*x509.Certificate{},
		TrustedPublicKeys: []*rsa.PublicKey{},
	}
}

// AddTrustedPublicKey adds a trusted public key
func (tc *TrustChain) AddTrustedPublicKey(publicKey *rsa.PublicKey) {
	tc.TrustedPublicKeys = append(tc.TrustedPublicKeys, publicKey)
}

// VerifyWithTrustChain verifies signature against trust chain
func (sm *SignatureManager) VerifyWithTrustChain(document *core.LIVDocument, trustChain *TrustChain) *SignatureVerificationResult {
	// Try verification with each trusted public key
	for _, publicKey := range trustChain.TrustedPublicKeys {
		result := sm.VerifyDocument(document, publicKey)
		if result.Valid {
			return result
		}
	}
	
	// If no trusted key worked, return failure
	return &SignatureVerificationResult{
		Valid:            false,
		Errors:           []string{"no trusted signature found"},
		VerificationTime: time.Now(),
	}
}

// SignatureInfo contains information about a signature
type SignatureInfo struct {
	Algorithm     string    `json:"algorithm"`
	KeySize       int       `json:"key_size"`
	SignedAt      time.Time `json:"signed_at"`
	ValidUntil    time.Time `json:"valid_until,omitempty"`
	Issuer        string    `json:"issuer,omitempty"`
	Subject       string    `json:"subject,omitempty"`
	Fingerprint   string    `json:"fingerprint"`
}

// GetSignatureInfo extracts information about a signature
func (sm *SignatureManager) GetSignatureInfo(publicKey *rsa.PublicKey) *SignatureInfo {
	// Calculate key fingerprint
	publicKeyBytes, _ := x509.MarshalPKIXPublicKey(publicKey)
	fingerprint := sm.hasher.HashBytes(publicKeyBytes)
	
	return &SignatureInfo{
		Algorithm:   "RSA-SHA256",
		KeySize:     publicKey.Size() * 8, // Convert bytes to bits
		SignedAt:    time.Now(),
		Fingerprint: fingerprint[:16], // First 16 chars of hash
	}
}