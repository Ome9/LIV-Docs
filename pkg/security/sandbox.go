package security

import (
	"context"
	"fmt"
	"time"

	"github.com/liv-format/liv/pkg/core"
)

// Sandbox implements the core.Sandbox interface
type Sandbox struct {
	securityContext     *SecurityContext
	permissionValidator *PermissionValidator
	resourceMonitor     *ResourceMonitor
	loadedModules       map[string]core.WASMInstance
	logger              core.Logger
	metrics             core.MetricsCollector
	destroyed           bool
}

// Execute runs code within the sandbox
func (s *Sandbox) Execute(ctx context.Context, code string, permissions *core.WASMPermissions) (interface{}, error) {
	if s.destroyed {
		return nil, fmt.Errorf("sandbox has been destroyed")
	}

	// Validate permissions against security policy
	if !s.validateExecutionPermissions(permissions) {
		return nil, fmt.Errorf("execution permissions denied by security policy")
	}

	// For now, this is a placeholder implementation
	// In a real implementation, this would execute the code in a secure environment
	s.logger.Info("executing code in sandbox",
		"session_id", s.securityContext.SessionID,
		"code_length", len(code),
	)

	if s.metrics != nil {
		s.metrics.RecordSecurityEvent("code_execution", map[string]interface{}{
			"session_id":  s.securityContext.SessionID,
			"code_length": len(code),
		})
	}

	// Simulate execution result
	return map[string]interface{}{
		"result":     "execution_completed",
		"session_id": s.securityContext.SessionID,
		"timestamp":  time.Now(),
	}, nil
}

// LoadWASM loads a WASM module into the sandbox
func (s *Sandbox) LoadWASM(ctx context.Context, module []byte, config *core.WASMModule) (core.WASMInstance, error) {
	if s.destroyed {
		return nil, fmt.Errorf("sandbox has been destroyed")
	}

	if config == nil {
		return nil, fmt.Errorf("WASM module configuration cannot be nil")
	}

	// Validate the WASM module
	if err := s.validateWASMModule(module, config); err != nil {
		return nil, fmt.Errorf("WASM module validation failed: %w", err)
	}

	// Register module for resource monitoring
	constraints := s.securityContext.Constraints
	if config.Permissions != nil {
		constraints = &ResourceConstraints{
			MemoryLimit:     int64(config.Permissions.MemoryLimit),
			CPUTimeLimit:    time.Duration(config.Permissions.CPUTimeLimit) * time.Millisecond,
			AllowNetworking: config.Permissions.AllowNetworking,
			AllowFileSystem: config.Permissions.AllowFileSystem,
		}
	}

	err := s.resourceMonitor.RegisterModule(s.securityContext.SessionID, config.Name, constraints)
	if err != nil {
		return nil, fmt.Errorf("failed to register module for monitoring: %w", err)
	}

	// Create WASM instance
	instance := &WASMInstance{
		name:                config.Name,
		sessionID:           s.securityContext.SessionID,
		config:              config,
		constraints:         constraints,
		permissionValidator: s.permissionValidator,
		resourceMonitor:     s.resourceMonitor,
		logger:              s.logger,
		metrics:             s.metrics,
		startTime:           time.Now(),
	}

	if s.loadedModules == nil {
		s.loadedModules = make(map[string]core.WASMInstance)
	}
	s.loadedModules[config.Name] = instance

	s.logger.Info("WASM module loaded into sandbox",
		"session_id", s.securityContext.SessionID,
		"module_name", config.Name,
		"module_size", len(module),
	)

	if s.metrics != nil {
		s.metrics.RecordSecurityEvent("wasm_module_loaded", map[string]interface{}{
			"session_id":  s.securityContext.SessionID,
			"module_name": config.Name,
			"module_size": len(module),
		})
	}

	return instance, nil
}

// GetPermissions returns current sandbox permissions
func (s *Sandbox) GetPermissions() *core.SecurityPolicy {
	if s.destroyed {
		return nil
	}
	return s.securityContext.Policy
}

// UpdatePermissions updates sandbox permissions
func (s *Sandbox) UpdatePermissions(policy *core.SecurityPolicy) error {
	if s.destroyed {
		return fmt.Errorf("sandbox has been destroyed")
	}

	if policy == nil {
		return fmt.Errorf("security policy cannot be nil")
	}

	// Update the security context
	s.securityContext.Policy = policy
	s.securityContext.Constraints = &ResourceConstraints{
		MemoryLimit:     int64(policy.WASMPermissions.MemoryLimit),
		CPUTimeLimit:    time.Duration(policy.WASMPermissions.CPUTimeLimit) * time.Millisecond,
		AllowNetworking: policy.WASMPermissions.AllowNetworking,
		AllowFileSystem: policy.WASMPermissions.AllowFileSystem,
	}

	s.logger.Info("sandbox permissions updated",
		"session_id", s.securityContext.SessionID,
	)

	if s.metrics != nil {
		s.metrics.RecordSecurityEvent("permissions_updated", map[string]interface{}{
			"session_id": s.securityContext.SessionID,
		})
	}

	return nil
}

// Destroy destroys the sandbox and cleans up resources
func (s *Sandbox) Destroy() error {
	if s.destroyed {
		return fmt.Errorf("sandbox already destroyed")
	}

	// Terminate all loaded modules
	for name, instance := range s.loadedModules {
		if err := instance.Terminate(); err != nil {
			s.logger.Warn("failed to terminate WASM instance",
				"module_name", name,
				"error", err,
			)
		}
	}

	// Unregister modules from resource monitoring
	for name := range s.loadedModules {
		err := s.resourceMonitor.UnregisterModule(s.securityContext.SessionID, name)
		if err != nil {
			s.logger.Warn("failed to unregister module from monitoring",
				"module_name", name,
				"error", err,
			)
		}
	}

	// Destroy the security session
	err := s.permissionValidator.DestroySession(s.securityContext.SessionID)
	if err != nil {
		s.logger.Warn("failed to destroy security session",
			"session_id", s.securityContext.SessionID,
			"error", err,
		)
	}

	s.destroyed = true
	s.loadedModules = nil

	s.logger.Info("sandbox destroyed",
		"session_id", s.securityContext.SessionID,
	)

	if s.metrics != nil {
		s.metrics.RecordSecurityEvent("sandbox_destroyed", map[string]interface{}{
			"session_id": s.securityContext.SessionID,
		})
	}

	return nil
}

// Helper methods

func (s *Sandbox) validateExecutionPermissions(permissions *core.WASMPermissions) bool {
	if permissions == nil {
		return true // Use default permissions
	}

	// Check if requested permissions are allowed by the security policy
	result := s.permissionValidator.policyEngine.EvaluateWASMPermissions(permissions, s.securityContext.Policy)
	return result.Allowed
}

func (s *Sandbox) validateWASMModule(module []byte, config *core.WASMModule) error {
	// Validate module against security policy
	permissions := config.Permissions
	if permissions == nil && s.securityContext.Policy != nil {
		permissions = s.securityContext.Policy.WASMPermissions
	}

	result := s.permissionValidator.policyEngine.ValidateWASMModule(module, permissions)
	if !result.IsValid {
		return fmt.Errorf("WASM module validation failed: %v", result.Errors)
	}

	return nil
}

// WASMInstance implements the core.WASMInstance interface
type WASMInstance struct {
	name                string
	sessionID           string
	config              *core.WASMModule
	constraints         *ResourceConstraints
	permissionValidator *PermissionValidator
	resourceMonitor     *ResourceMonitor
	logger              core.Logger
	metrics             core.MetricsCollector
	startTime           time.Time
	terminated          bool
}

// Call invokes a WASM function
func (wi *WASMInstance) Call(ctx context.Context, function string, args ...interface{}) (interface{}, error) {
	if wi.terminated {
		return nil, fmt.Errorf("WASM instance has been terminated")
	}

	// Check if the function is in the allowed exports
	if !wi.isFunctionAllowed(function) {
		return nil, fmt.Errorf("function '%s' not in allowed exports", function)
	}

	// Record CPU usage (simulated)
	startTime := time.Now()
	defer func() {
		cpuTime := time.Since(startTime)
		wi.resourceMonitor.UpdateCPUUsage(wi.sessionID, wi.name, cpuTime)
	}()

	// Simulate function execution
	wi.logger.Info("WASM function called",
		"session_id", wi.sessionID,
		"module_name", wi.name,
		"function", function,
		"args_count", len(args),
	)

	if wi.metrics != nil {
		wi.metrics.RecordSecurityEvent("wasm_function_call", map[string]interface{}{
			"session_id":  wi.sessionID,
			"module_name": wi.name,
			"function":    function,
		})
	}

	// Return simulated result
	return map[string]interface{}{
		"result":      "function_executed",
		"function":    function,
		"module_name": wi.name,
		"timestamp":   time.Now(),
	}, nil
}

// GetExports returns available exported functions
func (wi *WASMInstance) GetExports() []string {
	if wi.terminated {
		return []string{}
	}
	return wi.config.Exports
}

// GetMemoryUsage returns current memory usage
func (wi *WASMInstance) GetMemoryUsage() uint64 {
	if wi.terminated {
		return 0
	}

	// In a real implementation, this would query the actual WASM runtime
	// For now, return a simulated value
	return 1024 * 1024 // 1MB
}

// SetMemoryLimit sets memory usage limit
func (wi *WASMInstance) SetMemoryLimit(limit uint64) error {
	if wi.terminated {
		return fmt.Errorf("WASM instance has been terminated")
	}

	if int64(limit) > wi.constraints.MemoryLimit {
		return fmt.Errorf("requested limit %d exceeds maximum allowed %d", limit, wi.constraints.MemoryLimit)
	}

	wi.constraints.MemoryLimit = int64(limit)

	wi.logger.Info("WASM memory limit updated",
		"session_id", wi.sessionID,
		"module_name", wi.name,
		"new_limit", limit,
	)

	return nil
}

// Terminate forcefully terminates the instance
func (wi *WASMInstance) Terminate() error {
	if wi.terminated {
		return fmt.Errorf("WASM instance already terminated")
	}

	wi.terminated = true

	wi.logger.Info("WASM instance terminated",
		"session_id", wi.sessionID,
		"module_name", wi.name,
		"runtime", time.Since(wi.startTime),
	)

	if wi.metrics != nil {
		wi.metrics.RecordWASMExecution(wi.name,
			time.Since(wi.startTime).Milliseconds(),
			wi.GetMemoryUsage(),
		)
	}

	return nil
}

// Helper methods for WASMInstance

func (wi *WASMInstance) isFunctionAllowed(function string) bool {
	for _, export := range wi.config.Exports {
		if export == function {
			return true
		}
	}
	return false
}
