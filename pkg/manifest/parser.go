package manifest

import (
	"encoding/json"
	"fmt"
	"io"
	"os"
	"strings"

	"github.com/liv-format/liv/pkg/core"
)

// ManifestParser handles parsing and serialization of manifest files
type ManifestParser struct {
	validator *ManifestValidator
}

// NewManifestParser creates a new manifest parser
func NewManifestParser() *ManifestParser {
	return &ManifestParser{
		validator: NewManifestValidator(),
	}
}

// ParseFromReader parses a manifest from an io.Reader
func (mp *ManifestParser) ParseFromReader(reader io.Reader) (*core.Manifest, error) {
	data, err := io.ReadAll(reader)
	if err != nil {
		return nil, fmt.Errorf("failed to read manifest data: %v", err)
	}

	return mp.ParseFromBytes(data)
}

// ParseFromBytes parses a manifest from byte data
func (mp *ManifestParser) ParseFromBytes(data []byte) (*core.Manifest, error) {
	// Validate JSON syntax first
	if !json.Valid(data) {
		return nil, fmt.Errorf("invalid JSON syntax")
	}

	// Parse and validate
	manifest, result := mp.validator.ValidateManifestJSON(data)
	if !result.IsValid {
		return nil, fmt.Errorf("manifest validation failed: %s", strings.Join(result.Errors, "; "))
	}

	return manifest, nil
}

// ParseFromFile parses a manifest from a file
func (mp *ManifestParser) ParseFromFile(filePath string) (*core.Manifest, error) {
	file, err := os.Open(filePath)
	if err != nil {
		return nil, fmt.Errorf("failed to open manifest file: %v", err)
	}
	defer file.Close()

	return mp.ParseFromReader(file)
}

// SerializeToBytes serializes a manifest to JSON bytes
func (mp *ManifestParser) SerializeToBytes(manifest *core.Manifest) ([]byte, error) {
	// Validate before serialization
	result := mp.validator.ValidateManifest(manifest)
	if !result.IsValid {
		return nil, fmt.Errorf("manifest validation failed: %s", strings.Join(result.Errors, "; "))
	}

	// Serialize with proper formatting
	data, err := json.MarshalIndent(manifest, "", "  ")
	if err != nil {
		return nil, fmt.Errorf("failed to serialize manifest: %v", err)
	}

	return data, nil
}

// SerializeToWriter serializes a manifest to an io.Writer
func (mp *ManifestParser) SerializeToWriter(manifest *core.Manifest, writer io.Writer) error {
	data, err := mp.SerializeToBytes(manifest)
	if err != nil {
		return err
	}

	_, err = writer.Write(data)
	if err != nil {
		return fmt.Errorf("failed to write manifest data: %v", err)
	}

	return nil
}

// SerializeToFile serializes a manifest to a file
func (mp *ManifestParser) SerializeToFile(manifest *core.Manifest, filePath string) error {
	file, err := os.Create(filePath)
	if err != nil {
		return fmt.Errorf("failed to create manifest file: %v", err)
	}
	defer file.Close()

	return mp.SerializeToWriter(manifest, file)
}

// ValidateAndParse combines validation and parsing in one step
func (mp *ManifestParser) ValidateAndParse(data []byte) (*core.Manifest, *core.ValidationResult, error) {
	manifest, result := mp.validator.ValidateManifestJSON(data)
	if !result.IsValid {
		return nil, result, fmt.Errorf("validation failed")
	}

	return manifest, result, nil
}

// ParseWithWarnings parses a manifest and returns warnings even if valid
func (mp *ManifestParser) ParseWithWarnings(data []byte) (*core.Manifest, []string, error) {
	manifest, result := mp.validator.ValidateManifestJSON(data)
	if !result.IsValid {
		return nil, nil, fmt.Errorf("manifest validation failed: %s", strings.Join(result.Errors, "; "))
	}

	return manifest, result.Warnings, nil
}

// CompareManifests compares two manifests and returns differences
func (mp *ManifestParser) CompareManifests(manifest1, manifest2 *core.Manifest) *ManifestDiff {
	diff := &ManifestDiff{
		Changes: make([]ManifestChange, 0),
	}

	// Compare versions
	if manifest1.Version != manifest2.Version {
		diff.Changes = append(diff.Changes, ManifestChange{
			Field:    "version",
			OldValue: manifest1.Version,
			NewValue: manifest2.Version,
			Type:     "modified",
		})
	}

	// Compare metadata
	if manifest1.Metadata != nil && manifest2.Metadata != nil {
		mp.compareMetadata(manifest1.Metadata, manifest2.Metadata, diff)
	} else if manifest1.Metadata != nil && manifest2.Metadata == nil {
		diff.Changes = append(diff.Changes, ManifestChange{
			Field: "metadata",
			Type:  "removed",
		})
	} else if manifest1.Metadata == nil && manifest2.Metadata != nil {
		diff.Changes = append(diff.Changes, ManifestChange{
			Field: "metadata",
			Type:  "added",
		})
	}

	// Compare security policies
	if manifest1.Security != nil && manifest2.Security != nil {
		mp.compareSecurityPolicies(manifest1.Security, manifest2.Security, diff)
	}

	// Compare resources
	mp.compareResources(manifest1.Resources, manifest2.Resources, diff)

	// Compare WASM configurations
	if manifest1.WASMConfig != nil && manifest2.WASMConfig != nil {
		mp.compareWASMConfigs(manifest1.WASMConfig, manifest2.WASMConfig, diff)
	}

	// Compare feature flags
	if manifest1.Features != nil && manifest2.Features != nil {
		mp.compareFeatureFlags(manifest1.Features, manifest2.Features, diff)
	}

	return diff
}

// ManifestDiff represents differences between two manifests
type ManifestDiff struct {
	Changes []ManifestChange `json:"changes"`
}

// ManifestChange represents a single change in a manifest
type ManifestChange struct {
	Field    string      `json:"field"`
	OldValue interface{} `json:"old_value,omitempty"`
	NewValue interface{} `json:"new_value,omitempty"`
	Type     string      `json:"type"` // "added", "removed", "modified"
}

// Helper methods for comparison

func (mp *ManifestParser) compareMetadata(meta1, meta2 *core.DocumentMetadata, diff *ManifestDiff) {
	if meta1.Title != meta2.Title {
		diff.Changes = append(diff.Changes, ManifestChange{
			Field:    "metadata.title",
			OldValue: meta1.Title,
			NewValue: meta2.Title,
			Type:     "modified",
		})
	}

	if meta1.Author != meta2.Author {
		diff.Changes = append(diff.Changes, ManifestChange{
			Field:    "metadata.author",
			OldValue: meta1.Author,
			NewValue: meta2.Author,
			Type:     "modified",
		})
	}

	if meta1.Description != meta2.Description {
		diff.Changes = append(diff.Changes, ManifestChange{
			Field:    "metadata.description",
			OldValue: meta1.Description,
			NewValue: meta2.Description,
			Type:     "modified",
		})
	}

	if meta1.Version != meta2.Version {
		diff.Changes = append(diff.Changes, ManifestChange{
			Field:    "metadata.version",
			OldValue: meta1.Version,
			NewValue: meta2.Version,
			Type:     "modified",
		})
	}

	if !meta1.Modified.Equal(meta2.Modified) {
		diff.Changes = append(diff.Changes, ManifestChange{
			Field:    "metadata.modified",
			OldValue: meta1.Modified.Format("2006-01-02T15:04:05Z07:00"),
			NewValue: meta2.Modified.Format("2006-01-02T15:04:05Z07:00"),
			Type:     "modified",
		})
	}
}

func (mp *ManifestParser) compareSecurityPolicies(policy1, policy2 *core.SecurityPolicy, diff *ManifestDiff) {
	// Compare WASM permissions
	if policy1.WASMPermissions != nil && policy2.WASMPermissions != nil {
		wasm1, wasm2 := policy1.WASMPermissions, policy2.WASMPermissions

		if wasm1.MemoryLimit != wasm2.MemoryLimit {
			diff.Changes = append(diff.Changes, ManifestChange{
				Field:    "security.wasm_permissions.memory_limit",
				OldValue: wasm1.MemoryLimit,
				NewValue: wasm2.MemoryLimit,
				Type:     "modified",
			})
		}

		if wasm1.CPUTimeLimit != wasm2.CPUTimeLimit {
			diff.Changes = append(diff.Changes, ManifestChange{
				Field:    "security.wasm_permissions.cpu_time_limit",
				OldValue: wasm1.CPUTimeLimit,
				NewValue: wasm2.CPUTimeLimit,
				Type:     "modified",
			})
		}

		if wasm1.AllowNetworking != wasm2.AllowNetworking {
			diff.Changes = append(diff.Changes, ManifestChange{
				Field:    "security.wasm_permissions.allow_networking",
				OldValue: wasm1.AllowNetworking,
				NewValue: wasm2.AllowNetworking,
				Type:     "modified",
			})
		}
	}

	// Compare JS permissions
	if policy1.JSPermissions != nil && policy2.JSPermissions != nil {
		js1, js2 := policy1.JSPermissions, policy2.JSPermissions

		if js1.ExecutionMode != js2.ExecutionMode {
			diff.Changes = append(diff.Changes, ManifestChange{
				Field:    "security.js_permissions.execution_mode",
				OldValue: js1.ExecutionMode,
				NewValue: js2.ExecutionMode,
				Type:     "modified",
			})
		}

		if js1.DOMAccess != js2.DOMAccess {
			diff.Changes = append(diff.Changes, ManifestChange{
				Field:    "security.js_permissions.dom_access",
				OldValue: js1.DOMAccess,
				NewValue: js2.DOMAccess,
				Type:     "modified",
			})
		}
	}

	// Compare CSP
	if policy1.ContentSecurityPolicy != policy2.ContentSecurityPolicy {
		diff.Changes = append(diff.Changes, ManifestChange{
			Field:    "security.content_security_policy",
			OldValue: policy1.ContentSecurityPolicy,
			NewValue: policy2.ContentSecurityPolicy,
			Type:     "modified",
		})
	}
}

func (mp *ManifestParser) compareResources(resources1, resources2 map[string]*core.Resource, diff *ManifestDiff) {
	// Find added and modified resources
	for path, resource2 := range resources2 {
		if resource1, exists := resources1[path]; exists {
			// Resource exists in both, check for modifications
			if resource1.Hash != resource2.Hash {
				diff.Changes = append(diff.Changes, ManifestChange{
					Field:    fmt.Sprintf("resources.%s.hash", path),
					OldValue: resource1.Hash,
					NewValue: resource2.Hash,
					Type:     "modified",
				})
			}
			if resource1.Size != resource2.Size {
				diff.Changes = append(diff.Changes, ManifestChange{
					Field:    fmt.Sprintf("resources.%s.size", path),
					OldValue: resource1.Size,
					NewValue: resource2.Size,
					Type:     "modified",
				})
			}
		} else {
			// Resource added
			diff.Changes = append(diff.Changes, ManifestChange{
				Field: fmt.Sprintf("resources.%s", path),
				Type:  "added",
			})
		}
	}

	// Find removed resources
	for path := range resources1 {
		if _, exists := resources2[path]; !exists {
			diff.Changes = append(diff.Changes, ManifestChange{
				Field: fmt.Sprintf("resources.%s", path),
				Type:  "removed",
			})
		}
	}
}

func (mp *ManifestParser) compareWASMConfigs(config1, config2 *core.WASMConfiguration, diff *ManifestDiff) {
	if config1.MemoryLimit != config2.MemoryLimit {
		diff.Changes = append(diff.Changes, ManifestChange{
			Field:    "wasm_config.memory_limit",
			OldValue: config1.MemoryLimit,
			NewValue: config2.MemoryLimit,
			Type:     "modified",
		})
	}

	// Compare modules
	for name, module2 := range config2.Modules {
		if module1, exists := config1.Modules[name]; exists {
			if module1.Version != module2.Version {
				diff.Changes = append(diff.Changes, ManifestChange{
					Field:    fmt.Sprintf("wasm_config.modules.%s.version", name),
					OldValue: module1.Version,
					NewValue: module2.Version,
					Type:     "modified",
				})
			}
		} else {
			diff.Changes = append(diff.Changes, ManifestChange{
				Field: fmt.Sprintf("wasm_config.modules.%s", name),
				Type:  "added",
			})
		}
	}

	for name := range config1.Modules {
		if _, exists := config2.Modules[name]; !exists {
			diff.Changes = append(diff.Changes, ManifestChange{
				Field: fmt.Sprintf("wasm_config.modules.%s", name),
				Type:  "removed",
			})
		}
	}
}

func (mp *ManifestParser) compareFeatureFlags(features1, features2 *core.FeatureFlags, diff *ManifestDiff) {
	if features1.Animations != features2.Animations {
		diff.Changes = append(diff.Changes, ManifestChange{
			Field:    "features.animations",
			OldValue: features1.Animations,
			NewValue: features2.Animations,
			Type:     "modified",
		})
	}

	if features1.Interactivity != features2.Interactivity {
		diff.Changes = append(diff.Changes, ManifestChange{
			Field:    "features.interactivity",
			OldValue: features1.Interactivity,
			NewValue: features2.Interactivity,
			Type:     "modified",
		})
	}

	if features1.Charts != features2.Charts {
		diff.Changes = append(diff.Changes, ManifestChange{
			Field:    "features.charts",
			OldValue: features1.Charts,
			NewValue: features2.Charts,
			Type:     "modified",
		})
	}

	if features1.WebAssembly != features2.WebAssembly {
		diff.Changes = append(diff.Changes, ManifestChange{
			Field:    "features.webassembly",
			OldValue: features1.WebAssembly,
			NewValue: features2.WebAssembly,
			Type:     "modified",
		})
	}
}