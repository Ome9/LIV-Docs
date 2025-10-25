package wasm

import (
	"context"
	"fmt"
	"sync"
	"time"

	"github.com/liv-format/liv/pkg/core"
)

// WASMRuntime orchestrates WASM module execution and lifecycle management
type WASMRuntime struct {
	loader          *WASMLoader
	securityManager core.SecurityManager
	activeSandboxes map[string]*RuntimeSandbox
	sandboxMutex    sync.RWMutex
	logger          core.Logger
	metrics         core.MetricsCollector
	config          *RuntimeConfiguration
	shutdownChan    chan struct{}
	monitoringWG    sync.WaitGroup
}

// RuntimeSandbox represents an active WASM execution sandbox
type RuntimeSandbox struct {
	ID              string
	Policy          *core.SecurityPolicy
	LoadedModules   map[string]core.WASMInstance
	CreatedAt       time.Time
	LastActivity    time.Time
	ExecutionCount  int64
	MemoryUsage     uint64
	CPUTime         time.Duration
	NetworkRequests int64
	FileOperations  int64
	Violations      []SecurityViolation
	Status          SandboxStatus
}

// SecurityViolation represents a security policy violation
type SecurityViolation struct {
	Type        string                 `json:"type"`
	Description string                 `json:"description"`
	ModuleName  string                 `json:"module_name"`
	Timestamp   time.Time              `json:"timestamp"`
	Severity    string                 `json:"severity"`
	Details     map[string]interface{} `json:"details"`
}

// SandboxStatus represents the current status of a sandbox
type SandboxStatus string

const (
	SandboxStatusActive     SandboxStatus = "active"
	SandboxStatusSuspended  SandboxStatus = "suspended"
	SandboxStatusTerminated SandboxStatus = "terminated"
	SandboxStatusError      SandboxStatus = "error"
)

// RuntimeConfiguration holds configuration for the WASM runtime
type RuntimeConfiguration struct {
	MaxSandboxes        int           `json:"max_sandboxes"`
	SandboxTimeout      time.Duration `json:"sandbox_timeout"`
	MonitoringInterval  time.Duration `json:"monitoring_interval"`
	MaxMemoryPerSandbox uint64        `json:"max_memory_per_sandbox"`
	MaxCPUTimePerSandbox time.Duration `json:"max_cpu_time_per_sandbox"`
	EnableResourceMonitoring bool      `json:"enable_resource_monitoring"`
	EnableSecurityAuditing   bool      `json:"enable_security_auditing"`
	AutoCleanupExpired       bool      `json:"auto_cleanup_expired"`
}

// NewWASMRuntime creates a new WASM runtime orchestrator
func NewWASMRuntime(loader *WASMLoader, securityManager core.SecurityManager, logger core.Logger, metrics core.MetricsCollector) *WASMRuntime {
	config := &RuntimeConfiguration{
		MaxSandboxes:             20,
		SandboxTimeout:           30 * time.Minute,
		MonitoringInterval:       5 * time.Second,
		MaxMemoryPerSandbox:      256 * 1024 * 1024, // 256MB
		MaxCPUTimePerSandbox:     60 * time.Second,   // 60 seconds
		EnableResourceMonitoring: true,
		EnableSecurityAuditing:   true,
		AutoCleanupExpired:       true,
	}

	return &WASMRuntime{
		loader:          loader,
		securityManager: securityManager,
		activeSandboxes: make(map[string]*RuntimeSandbox),
		logger:          logger,
		metrics:         metrics,
		config:          config,
		shutdownChan:    make(chan struct{}),
	}
}

// StartRuntime starts the WASM runtime and monitoring services
func (wr *WASMRuntime) StartRuntime(ctx context.Context) error {
	wr.logger.Info("starting WASM runtime")

	if wr.config.EnableResourceMonitoring {
		wr.monitoringWG.Add(1)
		go wr.resourceMonitoringLoop(ctx)
	}

	if wr.config.AutoCleanupExpired {
		wr.monitoringWG.Add(1)
		go wr.cleanupLoop(ctx)
	}

	wr.logger.Info("WASM runtime started successfully")
	return nil
}

// StopRuntime stops the WASM runtime and cleans up resources
func (wr *WASMRuntime) StopRuntime() error {
	wr.logger.Info("stopping WASM runtime")

	close(wr.shutdownChan)
	wr.monitoringWG.Wait()

	// Terminate all active sandboxes
	wr.sandboxMutex.Lock()
	defer wr.sandboxMutex.Unlock()

	for id, sandbox := range wr.activeSandboxes {
		if err := wr.terminateSandbox(sandbox); err != nil {
			wr.logger.Warn("failed to terminate sandbox during shutdown", "sandbox_id", id, "error", err)
		}
	}

	wr.activeSandboxes = make(map[string]*RuntimeSandbox)
	wr.logger.Info("WASM runtime stopped")
	return nil
}

// CreateSandbox creates a new WASM execution sandbox
func (wr *WASMRuntime) CreateSandbox(policy *core.SecurityPolicy) (string, error) {
	if policy == nil {
		return "", fmt.Errorf("security policy cannot be nil")
	}

	wr.sandboxMutex.Lock()
	defer wr.sandboxMutex.Unlock()

	// Check sandbox limit
	if len(wr.activeSandboxes) >= wr.config.MaxSandboxes {
		return "", fmt.Errorf("maximum number of sandboxes reached: %d", wr.config.MaxSandboxes)
	}

	// Generate unique sandbox ID with additional randomness
	sandboxID := fmt.Sprintf("sandbox_%d_%d", time.Now().UnixNano(), len(wr.activeSandboxes))

	// Create sandbox
	sandbox := &RuntimeSandbox{
		ID:            sandboxID,
		Policy:        policy,
		LoadedModules: make(map[string]core.WASMInstance),
		CreatedAt:     time.Now(),
		LastActivity:  time.Now(),
		Status:        SandboxStatusActive,
		Violations:    []SecurityViolation{},
	}

	wr.activeSandboxes[sandboxID] = sandbox

	wr.logger.Info("WASM sandbox created", "sandbox_id", sandboxID)

	if wr.metrics != nil {
		wr.metrics.RecordSecurityEvent("sandbox_created", map[string]interface{}{
			"sandbox_id": sandboxID,
		})
	}

	return sandboxID, nil
}

// LoadModuleInSandbox loads a WASM module into a specific sandbox
func (wr *WASMRuntime) LoadModuleInSandbox(ctx context.Context, sandboxID string, moduleName string, moduleData []byte, config *core.WASMModule) error {
	wr.sandboxMutex.Lock()
	defer wr.sandboxMutex.Unlock()

	sandbox, exists := wr.activeSandboxes[sandboxID]
	if !exists {
		return fmt.Errorf("sandbox %s not found", sandboxID)
	}

	if sandbox.Status != SandboxStatusActive {
		return fmt.Errorf("sandbox %s is not active (status: %s)", sandboxID, sandbox.Status)
	}

	// Validate module against sandbox policy
	if err := wr.validateModuleForSandbox(moduleData, config, sandbox); err != nil {
		return fmt.Errorf("module validation failed: %w", err)
	}

	// Load module using the loader
	instance, err := wr.loader.LoadModule(ctx, moduleName, moduleData)
	if err != nil {
		return fmt.Errorf("failed to load module: %w", err)
	}

	// Apply sandbox-specific memory limits
	if config.Permissions != nil && config.Permissions.MemoryLimit > 0 {
		if err := instance.SetMemoryLimit(config.Permissions.MemoryLimit); err != nil {
			instance.Terminate()
			return fmt.Errorf("failed to set memory limit: %w", err)
		}
	}

	sandbox.LoadedModules[moduleName] = instance
	sandbox.LastActivity = time.Now()

	wr.logger.Info("WASM module loaded in sandbox",
		"sandbox_id", sandboxID,
		"module_name", moduleName,
		"module_size", len(moduleData),
	)

	if wr.metrics != nil {
		wr.metrics.RecordSecurityEvent("module_loaded_in_sandbox", map[string]interface{}{
			"sandbox_id":  sandboxID,
			"module_name": moduleName,
			"module_size": len(moduleData),
		})
	}

	return nil
}

// ExecuteInSandbox executes a function in a WASM module within a sandbox
func (wr *WASMRuntime) ExecuteInSandbox(ctx context.Context, sandboxID string, moduleName string, functionName string, args ...interface{}) (interface{}, error) {
	wr.sandboxMutex.RLock()
	sandbox, exists := wr.activeSandboxes[sandboxID]
	wr.sandboxMutex.RUnlock()

	if !exists {
		return nil, fmt.Errorf("sandbox %s not found", sandboxID)
	}

	if sandbox.Status != SandboxStatusActive {
		return nil, fmt.Errorf("sandbox %s is not active (status: %s)", sandboxID, sandbox.Status)
	}

	instance, exists := sandbox.LoadedModules[moduleName]
	if !exists {
		return nil, fmt.Errorf("module %s not loaded in sandbox %s", moduleName, sandboxID)
	}

	// Create execution context with timeout
	execCtx, cancel := context.WithTimeout(ctx, wr.config.MaxCPUTimePerSandbox)
	defer cancel()

	startTime := time.Now()

	// Execute function
	result, err := instance.Call(execCtx, functionName, args...)
	
	executionTime := time.Since(startTime)

	// Update sandbox statistics
	wr.sandboxMutex.Lock()
	sandbox.LastActivity = time.Now()
	sandbox.ExecutionCount++
	sandbox.CPUTime += executionTime
	sandbox.MemoryUsage = instance.GetMemoryUsage()
	wr.sandboxMutex.Unlock()

	// Check for resource violations
	wr.checkResourceViolations(sandbox, executionTime)

	if err != nil {
		wr.logger.Warn("WASM function execution failed",
			"sandbox_id", sandboxID,
			"module_name", moduleName,
			"function", functionName,
			"error", err,
		)
		return nil, err
	}

	wr.logger.Debug("WASM function executed successfully",
		"sandbox_id", sandboxID,
		"module_name", moduleName,
		"function", functionName,
		"duration", executionTime,
	)

	return result, nil
}

// TerminateSandbox terminates a sandbox and all its modules
func (wr *WASMRuntime) TerminateSandbox(sandboxID string) error {
	wr.sandboxMutex.Lock()
	defer wr.sandboxMutex.Unlock()

	sandbox, exists := wr.activeSandboxes[sandboxID]
	if !exists {
		return fmt.Errorf("sandbox %s not found", sandboxID)
	}

	if err := wr.terminateSandbox(sandbox); err != nil {
		return err
	}

	delete(wr.activeSandboxes, sandboxID)

	wr.logger.Info("sandbox terminated",
		"sandbox_id", sandboxID,
		"runtime", time.Since(sandbox.CreatedAt),
		"executions", sandbox.ExecutionCount,
	)

	return nil
}

// GetSandboxInfo returns information about a specific sandbox
func (wr *WASMRuntime) GetSandboxInfo(sandboxID string) (*RuntimeSandbox, error) {
	wr.sandboxMutex.RLock()
	defer wr.sandboxMutex.RUnlock()

	sandbox, exists := wr.activeSandboxes[sandboxID]
	if !exists {
		return nil, fmt.Errorf("sandbox %s not found", sandboxID)
	}

	// Return a copy to prevent external modification
	sandboxCopy := &RuntimeSandbox{
		ID:              sandbox.ID,
		Policy:          sandbox.Policy,
		CreatedAt:       sandbox.CreatedAt,
		LastActivity:    sandbox.LastActivity,
		ExecutionCount:  sandbox.ExecutionCount,
		MemoryUsage:     sandbox.MemoryUsage,
		CPUTime:         sandbox.CPUTime,
		NetworkRequests: sandbox.NetworkRequests,
		FileOperations:  sandbox.FileOperations,
		Status:          sandbox.Status,
		Violations:      append([]SecurityViolation{}, sandbox.Violations...),
		LoadedModules:   make(map[string]core.WASMInstance),
	}

	// Copy module references
	for name, instance := range sandbox.LoadedModules {
		sandboxCopy.LoadedModules[name] = instance
	}

	return sandboxCopy, nil
}

// ListActiveSandboxes returns a list of all active sandbox IDs
func (wr *WASMRuntime) ListActiveSandboxes() []string {
	wr.sandboxMutex.RLock()
	defer wr.sandboxMutex.RUnlock()

	sandboxes := make([]string, 0, len(wr.activeSandboxes))
	for id := range wr.activeSandboxes {
		sandboxes = append(sandboxes, id)
	}
	return sandboxes
}

// GetRuntimeStats returns overall runtime statistics
func (wr *WASMRuntime) GetRuntimeStats() map[string]interface{} {
	wr.sandboxMutex.RLock()
	defer wr.sandboxMutex.RUnlock()

	totalMemory := uint64(0)
	totalCPUTime := time.Duration(0)
	totalExecutions := int64(0)
	totalViolations := 0

	for _, sandbox := range wr.activeSandboxes {
		totalMemory += sandbox.MemoryUsage
		totalCPUTime += sandbox.CPUTime
		totalExecutions += sandbox.ExecutionCount
		totalViolations += len(sandbox.Violations)
	}

	return map[string]interface{}{
		"active_sandboxes":   len(wr.activeSandboxes),
		"total_memory":       totalMemory,
		"total_cpu_time":     totalCPUTime.Milliseconds(),
		"total_executions":   totalExecutions,
		"total_violations":   totalViolations,
		"max_sandboxes":      wr.config.MaxSandboxes,
		"monitoring_enabled": wr.config.EnableResourceMonitoring,
	}
}

// Helper methods

func (wr *WASMRuntime) validateModuleForSandbox(moduleData []byte, config *core.WASMModule, sandbox *RuntimeSandbox) error {
	// Use security manager to validate module
	permissions := config.Permissions
	if permissions == nil && sandbox.Policy.WASMPermissions != nil {
		permissions = sandbox.Policy.WASMPermissions
	}

	if wr.securityManager != nil {
		return wr.securityManager.ValidateWASMModule(moduleData, permissions)
	}

	return nil
}

func (wr *WASMRuntime) terminateSandbox(sandbox *RuntimeSandbox) error {
	// Terminate all loaded modules
	for name, instance := range sandbox.LoadedModules {
		if err := instance.Terminate(); err != nil {
			wr.logger.Warn("failed to terminate module", "module", name, "error", err)
		}
	}

	sandbox.Status = SandboxStatusTerminated
	sandbox.LoadedModules = make(map[string]core.WASMInstance)

	return nil
}

func (wr *WASMRuntime) checkResourceViolations(sandbox *RuntimeSandbox, executionTime time.Duration) {
	violations := []SecurityViolation{}

	// Check memory violations
	if sandbox.MemoryUsage > wr.config.MaxMemoryPerSandbox {
		violations = append(violations, SecurityViolation{
			Type:        "memory_limit_exceeded",
			Description: fmt.Sprintf("Memory usage %d exceeds limit %d", sandbox.MemoryUsage, wr.config.MaxMemoryPerSandbox),
			Timestamp:   time.Now(),
			Severity:    "high",
			Details: map[string]interface{}{
				"current_usage": sandbox.MemoryUsage,
				"limit":         wr.config.MaxMemoryPerSandbox,
			},
		})
	}

	// Check CPU time violations
	if sandbox.CPUTime > wr.config.MaxCPUTimePerSandbox {
		violations = append(violations, SecurityViolation{
			Type:        "cpu_time_exceeded",
			Description: fmt.Sprintf("CPU time %v exceeds limit %v", sandbox.CPUTime, wr.config.MaxCPUTimePerSandbox),
			Timestamp:   time.Now(),
			Severity:    "high",
			Details: map[string]interface{}{
				"current_usage": sandbox.CPUTime.Milliseconds(),
				"limit":         wr.config.MaxCPUTimePerSandbox.Milliseconds(),
			},
		})
	}

	// Add violations to sandbox
	if len(violations) > 0 {
		wr.sandboxMutex.Lock()
		sandbox.Violations = append(sandbox.Violations, violations...)
		wr.sandboxMutex.Unlock()

		for _, violation := range violations {
			wr.logger.Warn("security violation detected",
				"sandbox_id", sandbox.ID,
				"type", violation.Type,
				"severity", violation.Severity,
			)

			if wr.metrics != nil {
				wr.metrics.RecordSecurityEvent("security_violation", map[string]interface{}{
					"sandbox_id": sandbox.ID,
					"type":       violation.Type,
					"severity":   violation.Severity,
				})
			}
		}
	}
}

func (wr *WASMRuntime) resourceMonitoringLoop(ctx context.Context) {
	defer wr.monitoringWG.Done()

	ticker := time.NewTicker(wr.config.MonitoringInterval)
	defer ticker.Stop()

	for {
		select {
		case <-ctx.Done():
			return
		case <-wr.shutdownChan:
			return
		case <-ticker.C:
			wr.performResourceMonitoring()
		}
	}
}

func (wr *WASMRuntime) performResourceMonitoring() {
	wr.sandboxMutex.RLock()
	defer wr.sandboxMutex.RUnlock()

	for _, sandbox := range wr.activeSandboxes {
		if sandbox.Status != SandboxStatusActive {
			continue
		}

		// Update memory usage from all modules
		totalMemory := uint64(0)
		for _, instance := range sandbox.LoadedModules {
			totalMemory += instance.GetMemoryUsage()
		}
		sandbox.MemoryUsage = totalMemory

		// Check for violations
		wr.checkResourceViolations(sandbox, 0)
	}
}

func (wr *WASMRuntime) cleanupLoop(ctx context.Context) {
	defer wr.monitoringWG.Done()

	ticker := time.NewTicker(1 * time.Minute) // Cleanup every minute
	defer ticker.Stop()

	for {
		select {
		case <-ctx.Done():
			return
		case <-wr.shutdownChan:
			return
		case <-ticker.C:
			wr.performCleanup()
		}
	}
}

func (wr *WASMRuntime) performCleanup() {
	wr.sandboxMutex.Lock()
	defer wr.sandboxMutex.Unlock()

	expiredSandboxes := []string{}
	cutoff := time.Now().Add(-wr.config.SandboxTimeout)

	for id, sandbox := range wr.activeSandboxes {
		if sandbox.LastActivity.Before(cutoff) {
			expiredSandboxes = append(expiredSandboxes, id)
		}
	}

	for _, id := range expiredSandboxes {
		sandbox := wr.activeSandboxes[id]
		if err := wr.terminateSandbox(sandbox); err != nil {
			wr.logger.Warn("failed to cleanup expired sandbox", "sandbox_id", id, "error", err)
		} else {
			delete(wr.activeSandboxes, id)
			wr.logger.Info("cleaned up expired sandbox", "sandbox_id", id)
		}
	}
}