// Permission Management Interface
// Creates UI for existing granular permission system and implements permission inheritance

package security

import (
	"context"
	"encoding/json"
	"fmt"
	"net/http"
	"time"

	"github.com/liv-format/liv/pkg/core"
)

// PermissionManager provides UI and API for managing granular permissions
type PermissionManager struct {
	policyManager   *PolicyManager
	securityManager core.SecurityManager
	cryptoProvider  core.CryptoProvider
	logger          core.Logger
	trustedSigners  map[string]*TrustedSigner
	permissionCache map[string]*PermissionEvaluation
}

// TrustedSigner represents a trusted certificate authority or signer
type TrustedSigner struct {
	ID          string    `json:"id"`
	Name        string    `json:"name"`
	PublicKey   []byte    `json:"public_key"`
	Certificate []byte    `json:"certificate"`
	ValidFrom   time.Time `json:"valid_from"`
	ValidUntil  time.Time `json:"valid_until"`
	Revoked     bool      `json:"revoked"`
	TrustLevel  string    `json:"trust_level"` // "system", "organization", "user"
}

// PermissionEvaluation represents the result of permission evaluation
type PermissionEvaluation struct {
	Granted       bool              `json:"granted"`
	InheritedFrom string            `json:"inherited_from,omitempty"`
	Restrictions  []string          `json:"restrictions"`
	Warnings      []SecurityWarning `json:"warnings"`
	TrustChain    []*TrustedSigner  `json:"trust_chain"`
	EvaluatedAt   time.Time         `json:"evaluated_at"`
	ExpiresAt     time.Time         `json:"expires_at"`
}

// PermissionTemplate defines reusable permission configurations
type PermissionTemplate struct {
	ID           string                `json:"id"`
	Name         string                `json:"name"`
	Description  string                `json:"description"`
	Category     string                `json:"category"`
	Permissions  *core.WASMPermissions `json:"permissions"`
	Restrictions []string              `json:"restrictions"`
	UseCase      string                `json:"use_case"`
}

// PermissionInheritanceRule defines how permissions are inherited
type PermissionInheritanceRule struct {
	ParentPolicy   string   `json:"parent_policy"`
	ChildPolicy    string   `json:"child_policy"`
	InheritedPerms []string `json:"inherited_permissions"`
	Overrides      []string `json:"overrides"`
	Restrictions   []string `json:"restrictions"`
}

// NewPermissionManager creates a new permission manager
func NewPermissionManager(pm *PolicyManager, sm core.SecurityManager, cp core.CryptoProvider, logger core.Logger) *PermissionManager {
	return &PermissionManager{
		policyManager:   pm,
		securityManager: sm,
		cryptoProvider:  cp,
		logger:          logger,
		trustedSigners:  make(map[string]*TrustedSigner),
		permissionCache: make(map[string]*PermissionEvaluation),
	}
}

// EvaluatePermissionRequest evaluates a permission request against policies
func (pm *PermissionManager) EvaluatePermissionRequest(ctx context.Context, request *PermissionRequest) (*PermissionEvaluation, error) {
	// Get the security policy
	policy, err := pm.policyManager.GetPolicy(ctx, request.PolicyID)
	if err != nil {
		return nil, fmt.Errorf("failed to get policy: %w", err)
	}

	evaluation := &PermissionEvaluation{
		Granted:     false,
		Warnings:    []SecurityWarning{},
		TrustChain:  []*TrustedSigner{},
		EvaluatedAt: time.Now(),
		ExpiresAt:   time.Now().Add(1 * time.Hour), // Default 1 hour expiry
	}

	// Check if permissions are granted by policy
	if perms, ok := request.RequestedPerms.(*core.WASMPermissions); ok && pm.securityManager.EvaluatePermissions(perms, policy.SecurityPolicy) {
		evaluation.Granted = true
	} else {
		// Check for inheritance
		inheritedEval, err := pm.checkPermissionInheritance(ctx, request, policy)
		if err != nil {
			pm.logger.Warn("Permission inheritance check failed", "error", err, "request_id", request.DocumentID)
		} else if inheritedEval != nil {
			evaluation = inheritedEval
		}
	}

	// Validate trust chain if signature verification is required
	if policy.AdminControls != nil && policy.AdminControls.RequireSignature {
		trustChain, err := pm.validateTrustChain(ctx, request.DocumentID)
		if err != nil {
			evaluation.Granted = false
			evaluation.Warnings = append(evaluation.Warnings, SecurityWarning{
				Type:           "trust_chain_validation_failed",
				Description:    fmt.Sprintf("Trust chain validation failed: %v", err),
				Details:        map[string]interface{}{"document_id": request.DocumentID},
				Recommendation: "Ensure document is signed by a trusted authority",
			})
		} else {
			evaluation.TrustChain = trustChain
		}
	}

	// Apply additional restrictions based on policy
	if perms, ok := request.RequestedPerms.(*core.WASMPermissions); ok {
		evaluation.Restrictions = pm.calculateRestrictions(perms, policy)

		// Generate security warnings
		warnings := pm.generateSecurityWarnings(perms, policy)
		evaluation.Warnings = append(evaluation.Warnings, warnings...)
	}

	// Cache the evaluation
	cacheKey := fmt.Sprintf("%s:%s:%s", request.DocumentID, request.PolicyID, request.ModuleName)
	pm.permissionCache[cacheKey] = evaluation

	// Log the evaluation
	pm.logger.Info("Permission evaluation completed",
		"document_id", request.DocumentID,
		"policy_id", request.PolicyID,
		"granted", evaluation.Granted,
		"warnings", len(evaluation.Warnings),
	)

	return evaluation, nil
}

// checkPermissionInheritance checks if permissions can be inherited from parent policies
func (pm *PermissionManager) checkPermissionInheritance(ctx context.Context, request *PermissionRequest, policy *SystemSecurityPolicy) (*PermissionEvaluation, error) {
	if policy.ParentPolicy == "" {
		return nil, nil // No inheritance possible
	}

	parentPolicy, err := pm.policyManager.GetPolicy(ctx, policy.ParentPolicy)
	if err != nil {
		return nil, fmt.Errorf("failed to get parent policy: %w", err)
	}

	// Check if parent policy grants the permissions
	perms, ok := request.RequestedPerms.(*core.WASMPermissions)
	if !ok {
		return nil, fmt.Errorf("invalid permission type")
	}

	if pm.securityManager.EvaluatePermissions(perms, parentPolicy.SecurityPolicy) {
		evaluation := &PermissionEvaluation{
			Granted:       true,
			InheritedFrom: policy.ParentPolicy,
			Warnings:      []SecurityWarning{},
			TrustChain:    []*TrustedSigner{},
			EvaluatedAt:   time.Now(),
			ExpiresAt:     time.Now().Add(30 * time.Minute), // Shorter expiry for inherited permissions
		}

		// Add inheritance warning
		evaluation.Warnings = append(evaluation.Warnings, SecurityWarning{
			Type:        "inherited_permissions",
			Description: fmt.Sprintf("Permissions inherited from parent policy: %s", policy.ParentPolicy),
			Details: map[string]interface{}{
				"parent_policy": policy.ParentPolicy,
				"child_policy":  policy.ID,
			},
			Recommendation: "Review inherited permissions for security implications",
		})

		return evaluation, nil
	}

	// Recursively check parent's parent
	return pm.checkPermissionInheritance(ctx, &PermissionRequest{
		DocumentID:     request.DocumentID,
		ModuleName:     request.ModuleName,
		RequestedPerms: request.RequestedPerms,
		PolicyID:       policy.ParentPolicy,
		UserContext:    request.UserContext,
		Justification:  request.Justification,
		RequestedAt:    request.RequestedAt,
	}, parentPolicy)
}

// validateTrustChain validates the signature trust chain
func (pm *PermissionManager) validateTrustChain(ctx context.Context, documentID string) ([]*TrustedSigner, error) {
	// In a real implementation, this would:
	// 1. Extract signatures from the document
	// 2. Validate each signature in the chain
	// 3. Check certificate validity and revocation status
	// 4. Build the complete trust chain

	// For demonstration, return a mock trust chain
	trustChain := []*TrustedSigner{
		{
			ID:         "system-ca",
			Name:       "System Certificate Authority",
			TrustLevel: "system",
			ValidFrom:  time.Now().Add(-365 * 24 * time.Hour),
			ValidUntil: time.Now().Add(365 * 24 * time.Hour),
			Revoked:    false,
		},
	}

	return trustChain, nil
}

// calculateRestrictions calculates additional restrictions based on permissions and policy
func (pm *PermissionManager) calculateRestrictions(perms *core.WASMPermissions, policy *SystemSecurityPolicy) []string {
	restrictions := []string{}

	// Memory restrictions
	if perms.MemoryLimit > policy.SecurityPolicy.WASMPermissions.MemoryLimit {
		restrictions = append(restrictions, fmt.Sprintf("Memory limited to %d bytes (requested %d)",
			policy.SecurityPolicy.WASMPermissions.MemoryLimit, perms.MemoryLimit))
	}

	// CPU time restrictions
	if perms.CPUTimeLimit > policy.SecurityPolicy.WASMPermissions.CPUTimeLimit {
		restrictions = append(restrictions, fmt.Sprintf("CPU time limited to %d ms (requested %d)",
			policy.SecurityPolicy.WASMPermissions.CPUTimeLimit, perms.CPUTimeLimit))
	}

	// Network restrictions
	if perms.AllowNetworking && !policy.SecurityPolicy.WASMPermissions.AllowNetworking {
		restrictions = append(restrictions, "Network access denied by policy")
	}

	// File system restrictions
	if perms.AllowFileSystem && !policy.SecurityPolicy.WASMPermissions.AllowFileSystem {
		restrictions = append(restrictions, "File system access denied by policy")
	}

	// Import restrictions
	for _, requestedImport := range perms.AllowedImports {
		allowed := false
		for _, allowedImport := range policy.SecurityPolicy.WASMPermissions.AllowedImports {
			if requestedImport == allowedImport {
				allowed = true
				break
			}
		}
		if !allowed {
			restrictions = append(restrictions, fmt.Sprintf("Import '%s' not allowed by policy", requestedImport))
		}
	}

	return restrictions
}

// generateSecurityWarnings generates security warnings based on permissions
func (pm *PermissionManager) generateSecurityWarnings(perms *core.WASMPermissions, policy *SystemSecurityPolicy) []SecurityWarning {
	warnings := []SecurityWarning{}

	// High memory usage warning
	if perms.MemoryLimit > 32*1024*1024 { // > 32MB
		warnings = append(warnings, SecurityWarning{
			Type:        "high_memory_usage",
			Description: fmt.Sprintf("High memory usage requested: %d MB", perms.MemoryLimit/(1024*1024)),
			Details: map[string]interface{}{
				"requested_memory": perms.MemoryLimit,
				"threshold":        32 * 1024 * 1024,
			},
			Recommendation: "Consider optimizing memory usage or reviewing the necessity of high memory allocation",
		})
	}

	// Long CPU time warning
	if perms.CPUTimeLimit > 10000 { // > 10 seconds
		warnings = append(warnings, SecurityWarning{
			Type:        "long_cpu_time",
			Description: fmt.Sprintf("Long CPU time requested: %d seconds", perms.CPUTimeLimit/1000),
			Details: map[string]interface{}{
				"requested_cpu_time": perms.CPUTimeLimit,
				"threshold":          10000,
			},
			Recommendation: "Consider optimizing algorithms or breaking work into smaller chunks",
		})
	}

	// Network access warning
	if perms.AllowNetworking {
		warnings = append(warnings, SecurityWarning{
			Type:           "network_access_requested",
			Description:    "Network access requested - potential security risk",
			Details:        map[string]interface{}{"network_access": true},
			Recommendation: "Ensure network access is necessary and review data transmission security",
		})
	}

	// File system access warning
	if perms.AllowFileSystem {
		warnings = append(warnings, SecurityWarning{
			Type:           "filesystem_access_requested",
			Description:    "File system access requested - potential security risk",
			Details:        map[string]interface{}{"filesystem_access": true},
			Recommendation: "Ensure file system access is necessary and review file access patterns",
		})
	}

	return warnings
}

// GetPermissionTemplates returns available permission templates
func (pm *PermissionManager) GetPermissionTemplates() []*PermissionTemplate {
	return []*PermissionTemplate{
		{
			ID:          "basic-document",
			Name:        "Basic Document",
			Description: "Basic permissions for simple document rendering",
			Category:    "document",
			Permissions: &core.WASMPermissions{
				MemoryLimit:     4 * 1024 * 1024, // 4MB
				AllowedImports:  []string{"console"},
				CPUTimeLimit:    2000, // 2 seconds
				AllowNetworking: false,
				AllowFileSystem: false,
			},
			Restrictions: []string{"No network access", "No file system access"},
			UseCase:      "Static documents with minimal interactivity",
		},
		{
			ID:          "interactive-content",
			Name:        "Interactive Content",
			Description: "Permissions for interactive documents with user input",
			Category:    "interactive",
			Permissions: &core.WASMPermissions{
				MemoryLimit:     16 * 1024 * 1024, // 16MB
				AllowedImports:  []string{"console", "dom", "events"},
				CPUTimeLimit:    10000, // 10 seconds
				AllowNetworking: false,
				AllowFileSystem: false,
			},
			Restrictions: []string{"No network access", "No file system access"},
			UseCase:      "Interactive forms, games, and dynamic content",
		},
		{
			ID:          "data-visualization",
			Name:        "Data Visualization",
			Description: "Permissions for charts and data processing",
			Category:    "visualization",
			Permissions: &core.WASMPermissions{
				MemoryLimit:     32 * 1024 * 1024, // 32MB
				AllowedImports:  []string{"console", "dom", "canvas", "webgl"},
				CPUTimeLimit:    15000, // 15 seconds
				AllowNetworking: false,
				AllowFileSystem: false,
			},
			Restrictions: []string{"No network access", "No file system access"},
			UseCase:      "Charts, graphs, and complex data visualizations",
		},
		{
			ID:          "network-enabled",
			Name:        "Network Enabled",
			Description: "Permissions for documents that need network access",
			Category:    "network",
			Permissions: &core.WASMPermissions{
				MemoryLimit:     16 * 1024 * 1024, // 16MB
				AllowedImports:  []string{"console", "dom", "fetch"},
				CPUTimeLimit:    10000, // 10 seconds
				AllowNetworking: true,
				AllowFileSystem: false,
			},
			Restrictions: []string{"Network access to trusted domains only", "No file system access"},
			UseCase:      "Documents that fetch external data or communicate with APIs",
		},
	}
}

// CreatePermissionTemplate creates a new permission template
func (pm *PermissionManager) CreatePermissionTemplate(template *PermissionTemplate) error {
	// Validate template
	if template.ID == "" || template.Name == "" {
		return fmt.Errorf("template ID and name are required")
	}

	if template.Permissions == nil {
		return fmt.Errorf("template permissions are required")
	}

	// In a real implementation, this would save to persistent storage
	pm.logger.Info("Permission template created", "template_id", template.ID, "name", template.Name)

	return nil
}

// HTTP Handlers for Permission Management UI

// ServePermissionManagementUI serves the permission management web interface
func (pm *PermissionManager) ServePermissionManagementUI() http.Handler {
	mux := http.NewServeMux()

	// Serve static files
	mux.HandleFunc("/", pm.handlePermissionDashboard)
	mux.HandleFunc("/api/permissions/evaluate", pm.handleEvaluatePermission)
	mux.HandleFunc("/api/permissions/templates", pm.handlePermissionTemplates)
	mux.HandleFunc("/api/permissions/policies", pm.handlePolicies)
	mux.HandleFunc("/api/permissions/trust-chain", pm.handleTrustChain)

	return mux
}

// handlePermissionDashboard serves the main permission management dashboard
func (pm *PermissionManager) handlePermissionDashboard(w http.ResponseWriter, r *http.Request) {
	if r.Method != http.MethodGet {
		http.Error(w, "Method not allowed", http.StatusMethodNotAllowed)
		return
	}

	// Serve the permission management UI HTML
	html := pm.generatePermissionDashboardHTML()
	w.Header().Set("Content-Type", "text/html")
	w.Write([]byte(html))
}

// handleEvaluatePermission handles permission evaluation requests
func (pm *PermissionManager) handleEvaluatePermission(w http.ResponseWriter, r *http.Request) {
	if r.Method != http.MethodPost {
		http.Error(w, "Method not allowed", http.StatusMethodNotAllowed)
		return
	}

	var request PermissionRequest
	if err := json.NewDecoder(r.Body).Decode(&request); err != nil {
		http.Error(w, "Invalid request body", http.StatusBadRequest)
		return
	}

	evaluation, err := pm.EvaluatePermissionRequest(r.Context(), &request)
	if err != nil {
		http.Error(w, fmt.Sprintf("Permission evaluation failed: %v", err), http.StatusInternalServerError)
		return
	}

	w.Header().Set("Content-Type", "application/json")
	json.NewEncoder(w).Encode(evaluation)
}

// handlePermissionTemplates handles permission template requests
func (pm *PermissionManager) handlePermissionTemplates(w http.ResponseWriter, r *http.Request) {
	switch r.Method {
	case http.MethodGet:
		templates := pm.GetPermissionTemplates()
		w.Header().Set("Content-Type", "application/json")
		json.NewEncoder(w).Encode(templates)
	case http.MethodPost:
		var template PermissionTemplate
		if err := json.NewDecoder(r.Body).Decode(&template); err != nil {
			http.Error(w, "Invalid request body", http.StatusBadRequest)
			return
		}

		if err := pm.CreatePermissionTemplate(&template); err != nil {
			http.Error(w, fmt.Sprintf("Failed to create template: %v", err), http.StatusBadRequest)
			return
		}

		w.WriteHeader(http.StatusCreated)
		json.NewEncoder(w).Encode(template)
	default:
		http.Error(w, "Method not allowed", http.StatusMethodNotAllowed)
	}
}

// handlePolicies handles security policy requests
func (pm *PermissionManager) handlePolicies(w http.ResponseWriter, r *http.Request) {
	if r.Method != http.MethodGet {
		http.Error(w, "Method not allowed", http.StatusMethodNotAllowed)
		return
	}

	policies, err := pm.policyManager.ListPolicies(r.Context())
	if err != nil {
		http.Error(w, fmt.Sprintf("Failed to list policies: %v", err), http.StatusInternalServerError)
		return
	}

	w.Header().Set("Content-Type", "application/json")
	json.NewEncoder(w).Encode(policies)
}

// handleTrustChain handles trust chain validation requests
func (pm *PermissionManager) handleTrustChain(w http.ResponseWriter, r *http.Request) {
	if r.Method != http.MethodGet {
		http.Error(w, "Method not allowed", http.StatusMethodNotAllowed)
		return
	}

	documentID := r.URL.Query().Get("document_id")
	if documentID == "" {
		http.Error(w, "document_id parameter required", http.StatusBadRequest)
		return
	}

	trustChain, err := pm.validateTrustChain(r.Context(), documentID)
	if err != nil {
		http.Error(w, fmt.Sprintf("Trust chain validation failed: %v", err), http.StatusInternalServerError)
		return
	}

	w.Header().Set("Content-Type", "application/json")
	json.NewEncoder(w).Encode(trustChain)
}
