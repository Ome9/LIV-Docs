package manifest

import (
	"encoding/json"
	"fmt"
	"regexp"
	"time"

	"github.com/go-playground/validator/v10"
	"github.com/liv-format/liv/pkg/core"
)

// Type aliases for backward compatibility with tests
type (
	Manifest          = core.Manifest
	DocumentMetadata  = core.DocumentMetadata
	Resource          = core.Resource
	SecurityPolicy    = core.SecurityPolicy
	WASMPermissions   = core.WASMPermissions
	JSPermissions     = core.JSPermissions
	NetworkPolicy     = core.NetworkPolicy
	StoragePolicy     = core.StoragePolicy
	FeatureFlags      = core.FeatureFlags
	WASMConfiguration = core.WASMConfiguration
	WASMModule        = core.WASMModule
)

// ManifestValidator provides validation for LIV document manifests
type ManifestValidator struct {
	validator *validator.Validate
}

// NewManifestValidator creates a new manifest validator with custom validation rules
func NewManifestValidator() *ManifestValidator {
	v := validator.New()

	// Register custom validation functions
	v.RegisterValidation("semver", validateSemVer)
	v.RegisterValidation("iso8601", validateISO8601)
	v.RegisterValidation("sha256", validateSHA256)
	v.RegisterValidation("mimetype", validateMimeType)
	v.RegisterValidation("csp", validateCSP)
	v.RegisterValidation("domain", validateDomain)
	v.RegisterValidation("wasmmodule", validateWASMModuleName)

	return &ManifestValidator{
		validator: v,
	}
}

// ValidateManifest validates a complete manifest structure
func (mv *ManifestValidator) ValidateManifest(manifest *core.Manifest) *core.ValidationResult {
	if manifest == nil {
		return &core.ValidationResult{
			IsValid: false,
			Errors:  []string{"manifest cannot be nil"},
		}
	}

	var errors []string
	var warnings []string

	// Validate struct using tags
	if err := mv.validator.Struct(manifest); err != nil {
		if validationErrors, ok := err.(validator.ValidationErrors); ok {
			for _, fieldError := range validationErrors {
				errors = append(errors, mv.formatValidationError(fieldError))
			}
		} else {
			errors = append(errors, fmt.Sprintf("validation error: %v", err))
		}
	}

	// Additional semantic validation
	semanticErrors, semanticWarnings := mv.validateSemantics(manifest)
	errors = append(errors, semanticErrors...)
	warnings = append(warnings, semanticWarnings...)

	return &core.ValidationResult{
		IsValid:  len(errors) == 0,
		Errors:   errors,
		Warnings: warnings,
	}
}

// ValidateManifestJSON validates a manifest from JSON bytes
func (mv *ManifestValidator) ValidateManifestJSON(data []byte) (*core.Manifest, *core.ValidationResult) {
	var manifest core.Manifest

	// Parse JSON
	if err := json.Unmarshal(data, &manifest); err != nil {
		return nil, &core.ValidationResult{
			IsValid: false,
			Errors:  []string{fmt.Sprintf("invalid JSON: %v", err)},
		}
	}

	// Validate parsed manifest
	result := mv.ValidateManifest(&manifest)
	return &manifest, result
}

// validateSemantics performs additional semantic validation beyond struct tags
func (mv *ManifestValidator) validateSemantics(manifest *core.Manifest) ([]string, []string) {
	var errors []string
	var warnings []string

	// Validate version compatibility
	if manifest.Version != "1.0" {
		warnings = append(warnings, fmt.Sprintf("manifest version '%s' may not be fully supported", manifest.Version))
	}

	// Validate metadata consistency
	if manifest.Metadata != nil {
		if manifest.Metadata.Created.After(manifest.Metadata.Modified) {
			errors = append(errors, "created date cannot be after modified date")
		}

		if manifest.Metadata.Modified.After(time.Now().Add(time.Hour)) {
			warnings = append(warnings, "modified date is in the future")
		}
	}

	// Validate security policy consistency
	if manifest.Security != nil {
		secErrors, secWarnings := mv.validateSecurityPolicy(manifest.Security)
		errors = append(errors, secErrors...)
		warnings = append(warnings, secWarnings...)
	}

	// Validate WASM configuration
	if manifest.WASMConfig != nil {
		wasmErrors, wasmWarnings := mv.validateWASMConfig(manifest.WASMConfig)
		errors = append(errors, wasmErrors...)
		warnings = append(warnings, wasmWarnings...)
	}

	// Validate resource references
	if manifest.Resources != nil {
		resErrors, resWarnings := mv.validateResources(manifest.Resources)
		errors = append(errors, resErrors...)
		warnings = append(warnings, resWarnings...)
	}

	// Validate feature flags consistency
	if manifest.Features != nil {
		featWarnings := mv.validateFeatureFlags(manifest.Features, manifest.WASMConfig)
		warnings = append(warnings, featWarnings...)
	}

	return errors, warnings
}

// validateSecurityPolicy validates security policy consistency
func (mv *ManifestValidator) validateSecurityPolicy(policy *core.SecurityPolicy) ([]string, []string) {
	var errors []string
	var warnings []string

	if policy.WASMPermissions == nil {
		errors = append(errors, "WASM permissions must be defined")
		return errors, warnings
	}

	if policy.JSPermissions == nil {
		errors = append(errors, "JavaScript permissions must be defined")
		return errors, warnings
	}

	// Check for overly permissive settings
	if policy.WASMPermissions.AllowNetworking && policy.NetworkPolicy != nil && policy.NetworkPolicy.AllowOutbound {
		warnings = append(warnings, "document allows both WASM and general network access")
	}

	if policy.JSPermissions.ExecutionMode == "trusted" {
		warnings = append(warnings, "document requests trusted JavaScript execution")
	}

	// Validate memory limits
	if policy.WASMPermissions.MemoryLimit > 256*1024*1024 { // 256MB
		warnings = append(warnings, "WASM memory limit is very high (>256MB)")
	}

	if policy.WASMPermissions.CPUTimeLimit > 30000 { // 30 seconds
		warnings = append(warnings, "WASM CPU time limit is very high (>30s)")
	}

	// Validate CSP if present
	if policy.ContentSecurityPolicy != "" {
		if !mv.isValidCSP(policy.ContentSecurityPolicy) {
			errors = append(errors, "invalid Content Security Policy syntax")
		}
	}

	return errors, warnings
}

// validateWASMConfig validates WASM configuration
func (mv *ManifestValidator) validateWASMConfig(config *core.WASMConfiguration) ([]string, []string) {
	var errors []string
	var warnings []string

	if len(config.Modules) == 0 {
		warnings = append(warnings, "no WASM modules defined")
		return errors, warnings
	}

	// Validate each module
	for name, module := range config.Modules {
		if module.Name != name {
			errors = append(errors, fmt.Sprintf("module name mismatch: key '%s' vs name '%s'", name, module.Name))
		}

		if module.EntryPoint == "" {
			errors = append(errors, fmt.Sprintf("module '%s' missing entry point", name))
		}

		// Validate semantic versioning
		if !mv.isValidSemVer(module.Version) {
			errors = append(errors, fmt.Sprintf("module '%s' has invalid version format", name))
		}

		// Check for circular dependencies
		if mv.hasCircularDependency(name, module, config.Modules) {
			errors = append(errors, fmt.Sprintf("circular dependency detected for module '%s'", name))
		}
	}

	// Validate global memory limit
	if config.MemoryLimit == 0 {
		warnings = append(warnings, "no global WASM memory limit set")
	}

	return errors, warnings
}

// validateResources validates resource definitions
func (mv *ManifestValidator) validateResources(resources map[string]*core.Resource) ([]string, []string) {
	var errors []string
	var warnings []string

	// Only require content/index.html; manifest.json is the manifest itself and shouldn't be validated as a resource
	requiredPaths := []string{
		"content/index.html",
	}

	// Check for required resources
	for _, path := range requiredPaths {
		if _, exists := resources[path]; !exists {
			errors = append(errors, fmt.Sprintf("required resource missing: %s", path))
		}
	}

	// Validate each resource
	for path, resource := range resources {
		if resource.Path != path {
			errors = append(errors, fmt.Sprintf("resource path mismatch: key '%s' vs path '%s'", path, resource.Path))
		}

		if resource.Size < 0 {
			errors = append(errors, fmt.Sprintf("resource '%s' has negative size", path))
		}

		if resource.Hash == "" {
			errors = append(errors, fmt.Sprintf("resource '%s' missing integrity hash", path))
		}

		// Validate MIME type
		if !mv.isValidMimeType(resource.Type) {
			warnings = append(warnings, fmt.Sprintf("resource '%s' has unusual MIME type: %s", path, resource.Type))
		}

		// Check for large resources
		if resource.Size > 10*1024*1024 { // 10MB
			warnings = append(warnings, fmt.Sprintf("resource '%s' is very large (%d bytes)", path, resource.Size))
		}
	}

	return errors, warnings
}

// validateFeatureFlags validates feature flag consistency
func (mv *ManifestValidator) validateFeatureFlags(features *core.FeatureFlags, wasmConfig *core.WASMConfiguration) []string {
	var warnings []string

	// Check for enabled features without corresponding modules
	if features.WebAssembly && (wasmConfig == nil || len(wasmConfig.Modules) == 0) {
		warnings = append(warnings, "WebAssembly feature enabled but no WASM modules defined")
	}

	if features.Charts && !features.Interactivity {
		warnings = append(warnings, "charts feature enabled but interactivity disabled")
	}

	if features.WebGL && !features.Interactivity {
		warnings = append(warnings, "WebGL feature enabled but interactivity disabled")
	}

	// Check for potentially resource-intensive combinations
	if features.Video && features.Audio && features.WebGL {
		warnings = append(warnings, "multiple media features enabled may impact performance")
	}

	return warnings
}

// Helper validation functions

func (mv *ManifestValidator) formatValidationError(err validator.FieldError) string {
	switch err.Tag() {
	case "required":
		return fmt.Sprintf("field '%s' is required", err.Field())
	case "min":
		return fmt.Sprintf("field '%s' must be at least %s", err.Field(), err.Param())
	case "max":
		return fmt.Sprintf("field '%s' must be at most %s", err.Field(), err.Param())
	case "len":
		return fmt.Sprintf("field '%s' must be exactly %s characters", err.Field(), err.Param())
	case "oneof":
		return fmt.Sprintf("field '%s' must be one of: %s", err.Field(), err.Param())
	case "semver":
		return fmt.Sprintf("field '%s' must be a valid semantic version", err.Field())
	case "iso8601":
		return fmt.Sprintf("field '%s' must be a valid ISO 8601 timestamp", err.Field())
	case "sha256":
		return fmt.Sprintf("field '%s' must be a valid SHA-256 hash", err.Field())
	case "mimetype":
		return fmt.Sprintf("field '%s' must be a valid MIME type", err.Field())
	case "csp":
		return fmt.Sprintf("field '%s' must be a valid Content Security Policy", err.Field())
	case "domain":
		return fmt.Sprintf("field '%s' must be a valid domain name", err.Field())
	case "wasmmodule":
		return fmt.Sprintf("field '%s' must be a valid WASM module name", err.Field())
	default:
		return fmt.Sprintf("field '%s' validation failed: %s", err.Field(), err.Tag())
	}
}

func (mv *ManifestValidator) isValidSemVer(version string) bool {
	semverRegex := regexp.MustCompile(`^(0|[1-9]\d*)\.(0|[1-9]\d*)\.(0|[1-9]\d*)(?:-((?:0|[1-9]\d*|\d*[a-zA-Z-][0-9a-zA-Z-]*)(?:\.(?:0|[1-9]\d*|\d*[a-zA-Z-][0-9a-zA-Z-]*))*))?(?:\+([0-9a-zA-Z-]+(?:\.[0-9a-zA-Z-]+)*))?$`)
	return semverRegex.MatchString(version)
}

func (mv *ManifestValidator) isValidCSP(csp string) bool {
	// Basic CSP validation - check for common directives and characters
	cspRegex := regexp.MustCompile(`^[a-zA-Z0-9\-\s'*.:/_; -]+$`)
	return cspRegex.MatchString(csp)
}

func (mv *ManifestValidator) isValidMimeType(mimeType string) bool {
	mimeRegex := regexp.MustCompile(`^[a-zA-Z0-9][a-zA-Z0-9!#$&\-\^_]*\/[a-zA-Z0-9][a-zA-Z0-9!#$&\-\^_.]*$`)
	return mimeRegex.MatchString(mimeType)
}

func (mv *ManifestValidator) hasCircularDependency(moduleName string, module *core.WASMModule, allModules map[string]*core.WASMModule) bool {
	visited := make(map[string]bool)
	return mv.checkCircularDependency(moduleName, module, allModules, visited)
}

func (mv *ManifestValidator) checkCircularDependency(moduleName string, module *core.WASMModule, allModules map[string]*core.WASMModule, visited map[string]bool) bool {
	if visited[moduleName] {
		return true // Circular dependency found
	}

	visited[moduleName] = true

	for _, dep := range module.Imports {
		if depModule, exists := allModules[dep]; exists {
			if mv.checkCircularDependency(dep, depModule, allModules, visited) {
				return true
			}
		}
	}

	visited[moduleName] = false
	return false
}

// Custom validation functions for validator tags

func validateSemVer(fl validator.FieldLevel) bool {
	semverRegex := regexp.MustCompile(`^(0|[1-9]\d*)\.(0|[1-9]\d*)\.(0|[1-9]\d*)(?:-((?:0|[1-9]\d*|\d*[a-zA-Z-][0-9a-zA-Z-]*)(?:\.(?:0|[1-9]\d*|\d*[a-zA-Z-][0-9a-zA-Z-]*))*))?(?:\+([0-9a-zA-Z-]+(?:\.[0-9a-zA-Z-]+)*))?$`)
	return semverRegex.MatchString(fl.Field().String())
}

func validateISO8601(fl validator.FieldLevel) bool {
	_, err := time.Parse(time.RFC3339, fl.Field().String())
	return err == nil
}

func validateSHA256(fl validator.FieldLevel) bool {
	sha256Regex := regexp.MustCompile(`^[a-fA-F0-9]{64}$`)
	return sha256Regex.MatchString(fl.Field().String())
}

func validateMimeType(fl validator.FieldLevel) bool {
	mimeRegex := regexp.MustCompile(`^[a-zA-Z0-9][a-zA-Z0-9!#$&\-\^_]*\/[a-zA-Z0-9][a-zA-Z0-9!#$&\-\^_.]*$`)
	return mimeRegex.MatchString(fl.Field().String())
}

func validateCSP(fl validator.FieldLevel) bool {
	cspRegex := regexp.MustCompile(`^[a-zA-Z0-9\-\s'*.:/_; ]+$`)
	return cspRegex.MatchString(fl.Field().String())
}

func validateDomain(fl validator.FieldLevel) bool {
	domainRegex := regexp.MustCompile(`^[a-zA-Z0-9]([a-zA-Z0-9\-]{0,61}[a-zA-Z0-9])?(\.[a-zA-Z0-9]([a-zA-Z0-9\-]{0,61}[a-zA-Z0-9])?)*$`)
	return domainRegex.MatchString(fl.Field().String())
}

func validateWASMModuleName(fl validator.FieldLevel) bool {
	moduleNameRegex := regexp.MustCompile(`^[a-zA-Z][a-zA-Z0-9_-]*$`)
	return moduleNameRegex.MatchString(fl.Field().String())
}
