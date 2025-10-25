package integrity

import (
	"crypto/rsa"
	"crypto/x509"
	"encoding/json"
	"fmt"
	"log"
	"os"
	"path/filepath"
	"time"

	"github.com/liv-format/liv/pkg/core"
)

// SignatureStorage handles persistent storage of signatures and certificates
type SignatureStorage struct {
	storageDir string
}

// NewSignatureStorage creates a new signature storage manager
func NewSignatureStorage(storageDir string) *SignatureStorage {
	return &SignatureStorage{
		storageDir: storageDir,
	}
}

// SaveSignatureBundle saves a signature bundle to storage
func (ss *SignatureStorage) SaveSignatureBundle(documentID string, bundle *core.SignatureBundle) error {
	if err := os.MkdirAll(ss.storageDir, 0755); err != nil {
		return fmt.Errorf("failed to create storage directory: %v", err)
	}

	filePath := filepath.Join(ss.storageDir, fmt.Sprintf("%s_signatures.json", documentID))
	
	data, err := json.MarshalIndent(bundle, "", "  ")
	if err != nil {
		return fmt.Errorf("failed to marshal signature bundle: %v", err)
	}

	if err := os.WriteFile(filePath, data, 0644); err != nil {
		return fmt.Errorf("failed to write signature bundle: %v", err)
	}

	return nil
}

// LoadSignatureBundle loads a signature bundle from storage
func (ss *SignatureStorage) LoadSignatureBundle(documentID string) (*core.SignatureBundle, error) {
	filePath := filepath.Join(ss.storageDir, fmt.Sprintf("%s_signatures.json", documentID))
	
	data, err := os.ReadFile(filePath)
	if err != nil {
		return nil, fmt.Errorf("failed to read signature bundle: %v", err)
	}

	var bundle core.SignatureBundle
	if err := json.Unmarshal(data, &bundle); err != nil {
		return nil, fmt.Errorf("failed to unmarshal signature bundle: %v", err)
	}

	return &bundle, nil
}

// CertificateManager handles X.509 certificate operations
type CertificateManager struct {
	signatureManager *SignatureManager
	trustStore       *TrustStore
}

// NewCertificateManager creates a new certificate manager
func NewCertificateManager(signatureManager *SignatureManager) *CertificateManager {
	return &CertificateManager{
		signatureManager: signatureManager,
		trustStore:       NewTrustStore(),
	}
}

// TrustStore manages trusted certificates and CAs
type TrustStore struct {
	rootCAs           []*x509.Certificate
	intermediateCAs   []*x509.Certificate
	trustedCerts      []*x509.Certificate
	revokedCerts      map[string]bool // Certificate serial numbers
}

// NewTrustStore creates a new trust store
func NewTrustStore() *TrustStore {
	return &TrustStore{
		rootCAs:          []*x509.Certificate{},
		intermediateCAs:  []*x509.Certificate{},
		trustedCerts:     []*x509.Certificate{},
		revokedCerts:     make(map[string]bool),
	}
}

// AddRootCA adds a root CA certificate to the trust store
func (ts *TrustStore) AddRootCA(cert *x509.Certificate) {
	ts.rootCAs = append(ts.rootCAs, cert)
}

// AddTrustedCertificate adds a trusted certificate
func (ts *TrustStore) AddTrustedCertificate(cert *x509.Certificate) {
	ts.trustedCerts = append(ts.trustedCerts, cert)
}

// RevokeCertificate marks a certificate as revoked
func (ts *TrustStore) RevokeCertificate(serialNumber string) {
	ts.revokedCerts[serialNumber] = true
}

// IsCertificateRevoked checks if a certificate is revoked
func (ts *TrustStore) IsCertificateRevoked(cert *x509.Certificate) bool {
	return ts.revokedCerts[cert.SerialNumber.String()]
}

// ValidateCertificateChain validates a certificate chain
func (ts *TrustStore) ValidateCertificateChain(cert *x509.Certificate) error {
	// Check if certificate is revoked
	if ts.IsCertificateRevoked(cert) {
		return fmt.Errorf("certificate is revoked")
	}

	// Check certificate validity period
	now := time.Now()
	if now.Before(cert.NotBefore) || now.After(cert.NotAfter) {
		return fmt.Errorf("certificate is not valid at current time")
	}

	// Create certificate pool with root CAs
	roots := x509.NewCertPool()
	for _, rootCA := range ts.rootCAs {
		roots.AddCert(rootCA)
	}

	// Create intermediate pool
	intermediates := x509.NewCertPool()
	for _, intermediate := range ts.intermediateCAs {
		intermediates.AddCert(intermediate)
	}

	// Verify certificate chain
	opts := x509.VerifyOptions{
		Roots:         roots,
		Intermediates: intermediates,
		CurrentTime:   now,
	}

	_, err := cert.Verify(opts)
	return err
}

// EnhancedSignatureManager extends the basic signature manager with certificate support
type EnhancedSignatureManager struct {
	*SignatureManager
	certificateManager *CertificateManager
	storage           *SignatureStorage
	auditLogger       *AuditLogger
}

// NewEnhancedSignatureManager creates a new enhanced signature manager
func NewEnhancedSignatureManager(storageDir string) *EnhancedSignatureManager {
	sm := NewSignatureManager()
	cm := NewCertificateManager(sm)
	storage := NewSignatureStorage(storageDir)
	auditLogger := NewAuditLogger()

	return &EnhancedSignatureManager{
		SignatureManager:   sm,
		certificateManager: cm,
		storage:           storage,
		auditLogger:       auditLogger,
	}
}

// SignDocumentWithCertificate signs a document using a certificate
func (esm *EnhancedSignatureManager) SignDocumentWithCertificate(document *core.LIVDocument, cert *x509.Certificate, privateKey interface{}) (*core.SignatureBundle, error) {
	// Validate certificate
	if err := esm.certificateManager.trustStore.ValidateCertificateChain(cert); err != nil {
		esm.auditLogger.LogSecurityEvent("certificate_validation_failed", map[string]interface{}{
			"error": err.Error(),
			"cert_subject": cert.Subject.String(),
		})
		return nil, fmt.Errorf("certificate validation failed: %v", err)
	}

	// Sign document (assuming RSA private key for now)
	rsaPrivateKey, ok := privateKey.(*rsa.PrivateKey)
	if !ok {
		return nil, fmt.Errorf("unsupported private key type")
	}

	signatures, err := esm.SignDocument(document, rsaPrivateKey)
	if err != nil {
		esm.auditLogger.LogSecurityEvent("document_signing_failed", map[string]interface{}{
			"error": err.Error(),
		})
		return nil, err
	}

	// Log successful signing
	esm.auditLogger.LogSecurityEvent("document_signed", map[string]interface{}{
		"cert_subject": cert.Subject.String(),
		"cert_serial":  cert.SerialNumber.String(),
		"signed_at":    time.Now(),
	})

	return signatures, nil
}

// VerifyDocumentWithCertificate verifies a document using certificate-based validation
func (esm *EnhancedSignatureManager) VerifyDocumentWithCertificate(document *core.LIVDocument, cert *x509.Certificate) *EnhancedSignatureVerificationResult {
	result := &EnhancedSignatureVerificationResult{
		SignatureVerificationResult: SignatureVerificationResult{
			Valid:              true,
			ManifestValid:      false,
			ContentValid:       false,
			WASMModulesValid:   make(map[string]bool),
			Errors:             []string{},
			VerificationTime:   time.Now(),
		},
		CertificateValid:    false,
		CertificateInfo:     nil,
		TrustChainValid:     false,
	}

	// Validate certificate chain
	if err := esm.certificateManager.trustStore.ValidateCertificateChain(cert); err != nil {
		result.Valid = false
		result.Errors = append(result.Errors, fmt.Sprintf("certificate validation failed: %v", err))
		esm.auditLogger.LogSecurityEvent("certificate_verification_failed", map[string]interface{}{
			"error": err.Error(),
			"cert_subject": cert.Subject.String(),
		})
	} else {
		result.CertificateValid = true
		result.TrustChainValid = true
	}

	// Extract public key from certificate
	publicKey, ok := cert.PublicKey.(*rsa.PublicKey)
	if !ok {
		result.Valid = false
		result.Errors = append(result.Errors, "certificate does not contain RSA public key")
		return result
	}

	// Verify document signatures
	basicResult := esm.VerifyDocument(document, publicKey)
	result.SignatureVerificationResult = *basicResult

	// Set certificate info
	result.CertificateInfo = &CertificateInfo{
		Subject:      cert.Subject.String(),
		Issuer:       cert.Issuer.String(),
		SerialNumber: cert.SerialNumber.String(),
		NotBefore:    cert.NotBefore,
		NotAfter:     cert.NotAfter,
		KeyUsage:     cert.KeyUsage,
	}

	// Log verification result
	esm.auditLogger.LogSecurityEvent("document_verification", map[string]interface{}{
		"valid":        result.Valid,
		"cert_subject": cert.Subject.String(),
		"cert_serial":  cert.SerialNumber.String(),
		"verified_at":  time.Now(),
	})

	return result
}

// ValidateWASMModuleSignature validates WASM module signature with enhanced security
func (esm *EnhancedSignatureManager) ValidateWASMModuleSignature(moduleName string, moduleData []byte, signature string, cert *x509.Certificate) *WASMSignatureValidationResult {
	result := &WASMSignatureValidationResult{
		Valid:           false,
		ModuleName:      moduleName,
		SignatureValid:  false,
		CertificateValid: false,
		SecurityChecks:  make(map[string]bool),
		Errors:          []string{},
		ValidatedAt:     time.Now(),
	}

	// Validate certificate
	if err := esm.certificateManager.trustStore.ValidateCertificateChain(cert); err != nil {
		result.Errors = append(result.Errors, fmt.Sprintf("certificate validation failed: %v", err))
		esm.auditLogger.LogSecurityEvent("wasm_certificate_validation_failed", map[string]interface{}{
			"module_name": moduleName,
			"error":       err.Error(),
		})
		return result
	}
	result.CertificateValid = true

	// Extract public key
	publicKey, ok := cert.PublicKey.(*rsa.PublicKey)
	if !ok {
		result.Errors = append(result.Errors, "certificate does not contain RSA public key")
		return result
	}

	// Verify WASM module signature
	valid, err := esm.VerifyWASMModuleSignature(moduleData, signature, publicKey)
	if err != nil {
		result.Errors = append(result.Errors, fmt.Sprintf("signature verification failed: %v", err))
		return result
	}
	result.SignatureValid = valid

	// Perform additional security checks
	result.SecurityChecks["wasm_magic_valid"] = esm.validateWASMHeader(moduleData)
	result.SecurityChecks["module_size_acceptable"] = len(moduleData) <= 10*1024*1024 // 10MB limit
	result.SecurityChecks["certificate_not_revoked"] = !esm.certificateManager.trustStore.IsCertificateRevoked(cert)

	// Overall validity
	result.Valid = result.SignatureValid && result.CertificateValid && 
		result.SecurityChecks["wasm_magic_valid"] && 
		result.SecurityChecks["module_size_acceptable"] && 
		result.SecurityChecks["certificate_not_revoked"]

	// Log validation result
	esm.auditLogger.LogSecurityEvent("wasm_signature_validation", map[string]interface{}{
		"module_name": moduleName,
		"valid":       result.Valid,
		"cert_subject": cert.Subject.String(),
		"validated_at": time.Now(),
	})

	return result
}

// validateWASMHeader validates WASM module header
func (esm *EnhancedSignatureManager) validateWASMHeader(moduleData []byte) bool {
	if len(moduleData) < 8 {
		return false
	}
	
	// Check WASM magic number and version
	magic := []byte{0x00, 0x61, 0x73, 0x6D}
	version := []byte{0x01, 0x00, 0x00, 0x00}
	
	for i := 0; i < 4; i++ {
		if moduleData[i] != magic[i] {
			return false
		}
		if moduleData[i+4] != version[i] {
			return false
		}
	}
	
	return true
}

// AuditLogger handles security event logging
type AuditLogger struct {
	logFile *os.File
}

// NewAuditLogger creates a new audit logger
func NewAuditLogger() *AuditLogger {
	// For now, use standard logger. In production, this would write to a secure log file
	return &AuditLogger{}
}

// LogSecurityEvent logs a security-related event
func (al *AuditLogger) LogSecurityEvent(eventType string, details map[string]interface{}) {
	event := map[string]interface{}{
		"timestamp":  time.Now().Format(time.RFC3339),
		"event_type": eventType,
		"details":    details,
	}
	
	eventJSON, _ := json.Marshal(event)
	log.Printf("SECURITY_EVENT: %s", string(eventJSON))
}

// Enhanced result structures

// EnhancedSignatureVerificationResult extends basic verification with certificate info
type EnhancedSignatureVerificationResult struct {
	SignatureVerificationResult
	CertificateValid    bool             `json:"certificate_valid"`
	CertificateInfo     *CertificateInfo `json:"certificate_info,omitempty"`
	TrustChainValid     bool             `json:"trust_chain_valid"`
}

// CertificateInfo contains certificate details
type CertificateInfo struct {
	Subject      string    `json:"subject"`
	Issuer       string    `json:"issuer"`
	SerialNumber string    `json:"serial_number"`
	NotBefore    time.Time `json:"not_before"`
	NotAfter     time.Time `json:"not_after"`
	KeyUsage     x509.KeyUsage `json:"key_usage"`
}

// WASMSignatureValidationResult contains WASM-specific validation results
type WASMSignatureValidationResult struct {
	Valid            bool              `json:"valid"`
	ModuleName       string            `json:"module_name"`
	SignatureValid   bool              `json:"signature_valid"`
	CertificateValid bool              `json:"certificate_valid"`
	SecurityChecks   map[string]bool   `json:"security_checks"`
	Errors           []string          `json:"errors"`
	ValidatedAt      time.Time         `json:"validated_at"`
}

// SignaturePolicy defines signature validation policies
type SignaturePolicy struct {
	RequireCertificates     bool     `json:"require_certificates"`
	AllowSelfSigned        bool     `json:"allow_self_signed"`
	RequiredKeyUsages      []x509.KeyUsage `json:"required_key_usages"`
	MaxCertificateAge      time.Duration `json:"max_certificate_age"`
	RequiredSignatureAlgs  []string `json:"required_signature_algorithms"`
	TrustedIssuers         []string `json:"trusted_issuers"`
}

// DefaultSignaturePolicy returns a secure default signature policy
func DefaultSignaturePolicy() *SignaturePolicy {
	return &SignaturePolicy{
		RequireCertificates:    true,
		AllowSelfSigned:       false,
		RequiredKeyUsages:     []x509.KeyUsage{x509.KeyUsageDigitalSignature},
		MaxCertificateAge:     365 * 24 * time.Hour, // 1 year
		RequiredSignatureAlgs: []string{"RSA-SHA256"},
		TrustedIssuers:        []string{},
	}
}

// ValidateSignaturePolicy validates a signature against policy
func (esm *EnhancedSignatureManager) ValidateSignaturePolicy(cert *x509.Certificate, policy *SignaturePolicy) error {
	// Check if certificates are required
	if policy.RequireCertificates && cert == nil {
		return fmt.Errorf("certificate required by policy")
	}

	if cert == nil {
		return nil // No certificate to validate
	}

	// Check self-signed certificates
	if !policy.AllowSelfSigned && cert.Subject.String() == cert.Issuer.String() {
		return fmt.Errorf("self-signed certificates not allowed by policy")
	}

	// Check key usage
	if len(policy.RequiredKeyUsages) > 0 {
		hasRequiredUsage := false
		for _, requiredUsage := range policy.RequiredKeyUsages {
			if cert.KeyUsage&requiredUsage != 0 {
				hasRequiredUsage = true
				break
			}
		}
		if !hasRequiredUsage {
			return fmt.Errorf("certificate does not have required key usage")
		}
	}

	// Check certificate age
	if policy.MaxCertificateAge > 0 {
		age := time.Since(cert.NotBefore)
		if age > policy.MaxCertificateAge {
			return fmt.Errorf("certificate is too old (age: %v, max: %v)", age, policy.MaxCertificateAge)
		}
	}

	// Check trusted issuers
	if len(policy.TrustedIssuers) > 0 {
		issuerFound := false
		for _, trustedIssuer := range policy.TrustedIssuers {
			if cert.Issuer.String() == trustedIssuer {
				issuerFound = true
				break
			}
		}
		if !issuerFound {
			return fmt.Errorf("certificate issuer not in trusted list")
		}
	}

	return nil
}