package wasm

import (
	"context"
	"fmt"
	"sync"
	"time"

	"github.com/liv-format/liv/pkg/core"
)

// WASMLoader implements the core.WASMLoader interface
type WASMLoader struct {
	loadedModules map[string]*LoadedModule
	modulesMutex  sync.RWMutex
	securityMgr   core.SecurityManager
	logger        core.Logger
	metrics       core.MetricsCollector
	config        *LoaderConfiguration
}

// LoadedModule represents a loaded WASM module with its metadata
type LoadedModule struct {
	Name        string
	Data        []byte
	Config      *core.WASMModule
	Instance    core.WASMInstance
	LoadTime    time.Time
	LastAccess  time.Time
	AccessCount int64
}

// LoaderConfiguration holds configuration for the WASM loader
type LoaderConfiguration struct {
	MaxModules          int           `json:"max_modules"`
	MaxModuleSize       int64         `json:"max_module_size"`
	ModuleTimeout       time.Duration `json:"module_timeout"`
	MemoryLimit         uint64        `json:"memory_limit"`
	EnableCaching       bool          `json:"enable_caching"`
	CacheExpiryTime     time.Duration `json:"cache_expiry_time"`
	AllowUnsafeModules  bool          `json:"allow_unsafe_modules"`
	StrictValidation    bool          `json:"strict_validation"`
}

// NewWASMLoader creates a new WASM loader with the given configuration
func NewWASMLoader(securityMgr core.SecurityManager, logger core.Logger, metrics core.MetricsCollector) *WASMLoader {
	config := &LoaderConfiguration{
		MaxModules:          10,
		MaxModuleSize:       16 * 1024 * 1024, // 16MB
		ModuleTimeout:       30 * time.Second,
		MemoryLimit:         128 * 1024 * 1024, // 128MB
		EnableCaching:       true,
		CacheExpiryTime:     1 * time.Hour,
		AllowUnsafeModules:  false,
		StrictValidation:    true,
	}

	return &WASMLoader{
		loadedModules: make(map[string]*LoadedModule),
		securityMgr:   securityMgr,
		logger:        logger,
		metrics:       metrics,
		config:        config,
	}
}

// LoadModule loads a WASM module into memory
func (wl *WASMLoader) LoadModule(ctx context.Context, name string, data []byte) (core.WASMInstance, error) {
	if name == "" {
		return nil, fmt.Errorf("module name cannot be empty")
	}

	if len(data) == 0 {
		return nil, fmt.Errorf("module data cannot be empty")
	}

	if int64(len(data)) > wl.config.MaxModuleSize {
		return nil, fmt.Errorf("module size %d exceeds maximum allowed size %d", len(data), wl.config.MaxModuleSize)
	}

	wl.modulesMutex.Lock()
	defer wl.modulesMutex.Unlock()

	// Check if module is already loaded
	if existing, exists := wl.loadedModules[name]; exists {
		existing.LastAccess = time.Now()
		existing.AccessCount++
		wl.logger.Info("reusing existing WASM module", "name", name)
		return existing.Instance, nil
	}

	// Check module limit
	if len(wl.loadedModules) >= wl.config.MaxModules {
		// Try to evict least recently used module
		if err := wl.evictLRUModule(); err != nil {
			return nil, fmt.Errorf("cannot load module: maximum modules reached and eviction failed: %w", err)
		}
	}

	// Validate the WASM module
	if err := wl.ValidateModule(data); err != nil {
		return nil, fmt.Errorf("module validation failed: %w", err)
	}

	// Create WASM instance
	instance, err := wl.createWASMInstance(ctx, name, data)
	if err != nil {
		return nil, fmt.Errorf("failed to create WASM instance: %w", err)
	}

	// Create module configuration
	config := &core.WASMModule{
		Name:       name,
		Version:    "1.0.0", // Default version
		EntryPoint: "main",
		Exports:    instance.GetExports(),
		Imports:    []string{}, // Will be populated during validation
		Permissions: &core.WASMPermissions{
			MemoryLimit:     wl.config.MemoryLimit,
			CPUTimeLimit:    30000, // 30 seconds
			AllowNetworking: false,
			AllowFileSystem: false,
			AllowedImports:  []string{},
		},
	}

	// Store loaded module
	loadedModule := &LoadedModule{
		Name:        name,
		Data:        data,
		Config:      config,
		Instance:    instance,
		LoadTime:    time.Now(),
		LastAccess:  time.Now(),
		AccessCount: 1,
	}

	wl.loadedModules[name] = loadedModule

	wl.logger.Info("WASM module loaded successfully",
		"name", name,
		"size", len(data),
		"exports", len(instance.GetExports()),
	)

	if wl.metrics != nil {
		wl.metrics.RecordSecurityEvent("wasm_module_loaded", map[string]interface{}{
			"module_name": name,
			"module_size": len(data),
			"exports":     len(instance.GetExports()),
		})
	}

	return instance, nil
}

// UnloadModule removes a WASM module from memory
func (wl *WASMLoader) UnloadModule(name string) error {
	wl.modulesMutex.Lock()
	defer wl.modulesMutex.Unlock()

	module, exists := wl.loadedModules[name]
	if !exists {
		return fmt.Errorf("module %s not found", name)
	}

	// Terminate the instance
	if err := module.Instance.Terminate(); err != nil {
		wl.logger.Warn("failed to terminate WASM instance", "name", name, "error", err)
	}

	delete(wl.loadedModules, name)

	wl.logger.Info("WASM module unloaded",
		"name", name,
		"runtime", time.Since(module.LoadTime),
		"access_count", module.AccessCount,
	)

	if wl.metrics != nil {
		wl.metrics.RecordWASMExecution(name,
			time.Since(module.LoadTime).Milliseconds(),
			module.Instance.GetMemoryUsage(),
		)
	}

	return nil
}

// ListModules returns all loaded modules
func (wl *WASMLoader) ListModules() []string {
	wl.modulesMutex.RLock()
	defer wl.modulesMutex.RUnlock()

	modules := make([]string, 0, len(wl.loadedModules))
	for name := range wl.loadedModules {
		modules = append(modules, name)
	}
	return modules
}

// GetModuleInfo returns information about a loaded module
func (wl *WASMLoader) GetModuleInfo(name string) (*core.WASMModule, error) {
	wl.modulesMutex.RLock()
	defer wl.modulesMutex.RUnlock()

	module, exists := wl.loadedModules[name]
	if !exists {
		return nil, fmt.Errorf("module %s not found", name)
	}

	module.LastAccess = time.Now()
	module.AccessCount++

	return module.Config, nil
}

// ValidateModule validates a WASM module before loading
func (wl *WASMLoader) ValidateModule(data []byte) error {
	if len(data) < 8 {
		return fmt.Errorf("WASM module too small: %d bytes", len(data))
	}

	// Check WASM magic number
	if string(data[:4]) != "\x00asm" {
		return fmt.Errorf("invalid WASM magic number")
	}

	// Check WASM version
	version := uint32(data[4]) | uint32(data[5])<<8 | uint32(data[6])<<16 | uint32(data[7])<<24
	if version != 1 {
		return fmt.Errorf("unsupported WASM version: %d", version)
	}

	// Additional validation using security manager
	if wl.securityMgr != nil {
		permissions := &core.WASMPermissions{
			MemoryLimit:     wl.config.MemoryLimit,
			CPUTimeLimit:    30000,
			AllowNetworking: false,
			AllowFileSystem: false,
			AllowedImports:  []string{},
		}

		if err := wl.securityMgr.ValidateWASMModule(data, permissions); err != nil {
			return fmt.Errorf("security validation failed: %w", err)
		}
	}

	return nil
}

// GetConfiguration returns the current loader configuration
func (wl *WASMLoader) GetConfiguration() *LoaderConfiguration {
	return wl.config
}

// UpdateConfiguration updates the loader configuration
func (wl *WASMLoader) UpdateConfiguration(config *LoaderConfiguration) error {
	if config == nil {
		return fmt.Errorf("configuration cannot be nil")
	}

	wl.config = config
	wl.logger.Info("WASM loader configuration updated")
	return nil
}

// GetLoadedModuleStats returns statistics about loaded modules
func (wl *WASMLoader) GetLoadedModuleStats() map[string]interface{} {
	wl.modulesMutex.RLock()
	defer wl.modulesMutex.RUnlock()

	totalMemory := uint64(0)
	totalSize := int64(0)
	oldestModule := time.Now()
	newestModule := time.Time{}

	for _, module := range wl.loadedModules {
		totalMemory += module.Instance.GetMemoryUsage()
		totalSize += int64(len(module.Data))

		if module.LoadTime.Before(oldestModule) {
			oldestModule = module.LoadTime
		}
		if module.LoadTime.After(newestModule) {
			newestModule = module.LoadTime
		}
	}

	return map[string]interface{}{
		"loaded_modules":  len(wl.loadedModules),
		"total_memory":    totalMemory,
		"total_size":      totalSize,
		"oldest_module":   oldestModule,
		"newest_module":   newestModule,
		"max_modules":     wl.config.MaxModules,
		"memory_limit":    wl.config.MemoryLimit,
	}
}

// CleanupExpiredModules removes modules that haven't been accessed recently
func (wl *WASMLoader) CleanupExpiredModules() int {
	if !wl.config.EnableCaching {
		return 0
	}

	wl.modulesMutex.Lock()
	defer wl.modulesMutex.Unlock()

	expiredModules := []string{}
	cutoff := time.Now().Add(-wl.config.CacheExpiryTime)

	for name, module := range wl.loadedModules {
		if module.LastAccess.Before(cutoff) {
			expiredModules = append(expiredModules, name)
		}
	}

	for _, name := range expiredModules {
		module := wl.loadedModules[name]
		if err := module.Instance.Terminate(); err != nil {
			wl.logger.Warn("failed to terminate expired WASM instance", "name", name, "error", err)
		}
		delete(wl.loadedModules, name)
	}

	if len(expiredModules) > 0 {
		wl.logger.Info("cleaned up expired WASM modules", "count", len(expiredModules))
	}

	return len(expiredModules)
}

// Helper methods

func (wl *WASMLoader) createWASMInstance(ctx context.Context, name string, data []byte) (core.WASMInstance, error) {
	// Create a new WASM instance
	instance := &WASMInstance{
		name:         name,
		data:         data,
		memoryUsage:  0,
		memoryLimit:  wl.config.MemoryLimit,
		exports:      []string{}, // Will be populated during initialization
		terminated:   false,
		createdAt:    time.Now(),
		logger:       wl.logger,
		metrics:      wl.metrics,
	}

	// Initialize the instance (simulate WASM loading)
	if err := instance.initialize(ctx); err != nil {
		return nil, fmt.Errorf("failed to initialize WASM instance: %w", err)
	}

	return instance, nil
}

func (wl *WASMLoader) evictLRUModule() error {
	if len(wl.loadedModules) == 0 {
		return fmt.Errorf("no modules to evict")
	}

	// Find least recently used module
	var lruName string
	var lruTime time.Time
	first := true

	for name, module := range wl.loadedModules {
		if first || module.LastAccess.Before(lruTime) {
			lruTime = module.LastAccess
			lruName = name
			first = false
		}
	}

	if lruName == "" {
		return fmt.Errorf("could not find module to evict")
	}

	// Evict the LRU module
	module := wl.loadedModules[lruName]
	if err := module.Instance.Terminate(); err != nil {
		wl.logger.Warn("failed to terminate evicted WASM instance", "name", lruName, "error", err)
	}

	delete(wl.loadedModules, lruName)

	wl.logger.Info("evicted LRU WASM module", "name", lruName, "last_access", lruTime)

	return nil
}