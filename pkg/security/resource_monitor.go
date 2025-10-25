package security

import (
	"context"
	"fmt"
	"sync"
	"time"

	"github.com/liv-format/liv/pkg/core"
)

// ResourceMonitor tracks and enforces resource usage for WASM modules
type ResourceMonitor struct {
	activeModules       map[string]*RuntimeMetrics
	modulesMutex        sync.RWMutex
	validator           *PermissionValidator
	logger              core.Logger
	metrics             core.MetricsCollector
	monitoringStop      chan struct{}
	monitoringWG        sync.WaitGroup
	totalMemoryUsage    int64
	totalCPUTime        time.Duration
	concurrentDocuments int
	networkBandwidth    int64
}

// NewResourceMonitor creates a new resource monitor
func NewResourceMonitor(validator *PermissionValidator, logger core.Logger, metrics core.MetricsCollector) *ResourceMonitor {
	return &ResourceMonitor{
		activeModules:  make(map[string]*RuntimeMetrics),
		validator:      validator,
		logger:         logger,
		metrics:        metrics,
		monitoringStop: make(chan struct{}),
	}
}

// StartMonitoring starts the resource monitoring loop
func (rm *ResourceMonitor) StartMonitoring(ctx context.Context, interval time.Duration) {
	rm.monitoringWG.Add(1)
	go rm.monitoringLoop(ctx, interval)
}

// StopMonitoring stops the resource monitoring loop
func (rm *ResourceMonitor) StopMonitoring() {
	close(rm.monitoringStop)
	rm.monitoringWG.Wait()
}

// RegisterModule registers a new WASM module for monitoring
func (rm *ResourceMonitor) RegisterModule(sessionID, moduleName string, constraints *ResourceConstraints) error {
	if sessionID == "" || moduleName == "" {
		return fmt.Errorf("session ID and module name cannot be empty")
	}

	if constraints == nil {
		return fmt.Errorf("resource constraints cannot be nil")
	}

	rm.modulesMutex.Lock()
	defer rm.modulesMutex.Unlock()

	moduleKey := fmt.Sprintf("%s:%s", sessionID, moduleName)

	if _, exists := rm.activeModules[moduleKey]; exists {
		return fmt.Errorf("module %s already registered for session %s", moduleName, sessionID)
	}

	metrics := &RuntimeMetrics{
		SessionID:  sessionID,
		ModuleName: moduleName,
		Memory: &MemoryUsage{
			Limit: constraints.MemoryLimit,
		},
		CPU: &CPUUsage{
			Limit: constraints.CPUTimeLimit,
		},
		Network:            &NetworkActivity{},
		FileSystem:         &FileSystemActivity{},
		PermissionRequests: []PermissionRequest{},
		PolicyViolations:   []PolicyViolation{},
		StartTime:          time.Now(),
	}

	rm.activeModules[moduleKey] = metrics

	rm.logger.Info("module registered for monitoring",
		"session_id", sessionID,
		"module_name", moduleName,
		"memory_limit", constraints.MemoryLimit,
		"cpu_limit", constraints.CPUTimeLimit,
	)

	if rm.metrics != nil {
		rm.metrics.RecordSecurityEvent("module_registered", map[string]interface{}{
			"session_id":   sessionID,
			"module_name":  moduleName,
			"memory_limit": constraints.MemoryLimit,
			"cpu_limit":    constraints.CPUTimeLimit.Milliseconds(),
		})
	}

	return nil
}

// UnregisterModule unregisters a WASM module from monitoring
func (rm *ResourceMonitor) UnregisterModule(sessionID, moduleName string) error {
	rm.modulesMutex.Lock()
	defer rm.modulesMutex.Unlock()

	moduleKey := fmt.Sprintf("%s:%s", sessionID, moduleName)

	metrics, exists := rm.activeModules[moduleKey]
	if !exists {
		return fmt.Errorf("module %s not found for session %s", moduleName, sessionID)
	}

	metrics.EndTime = time.Now()
	delete(rm.activeModules, moduleKey)

	rm.logger.Info("module unregistered from monitoring",
		"session_id", sessionID,
		"module_name", moduleName,
		"duration", metrics.GetDuration(),
	)

	if rm.metrics != nil {
		rm.metrics.RecordWASMExecution(moduleName,
			metrics.GetDuration().Milliseconds(),
			uint64(metrics.Memory.Peak),
		)
	}

	return nil
}

// UpdateMemoryUsage updates memory usage for a module
func (rm *ResourceMonitor) UpdateMemoryUsage(sessionID, moduleName string, allocated, used uint64) error {
	metrics, err := rm.getModuleMetrics(sessionID, moduleName)
	if err != nil {
		return err
	}

	rm.modulesMutex.Lock()
	defer rm.modulesMutex.Unlock()

	metrics.Memory.Allocated = int64(allocated)
	metrics.Memory.Used = int64(used)
	metrics.Memory.Timestamp = time.Now()

	if int64(used) > metrics.Memory.Peak {
		metrics.Memory.Peak = int64(used)
	}

	// Check for memory limit violations
	if metrics.Memory.IsMemoryLimitExceeded() {
		violation := PolicyViolation{
			Type:        "memory_limit_exceeded",
			Description: fmt.Sprintf("Memory usage %d exceeds limit %d", used, metrics.Memory.Limit),
			SessionID:   sessionID,
			ModuleName:  moduleName,
			Details: map[string]interface{}{
				"used":      used,
				"allocated": allocated,
				"limit":     metrics.Memory.Limit,
			},
			Timestamp: time.Now(),
			Severity:  SecuritySeverityError,
		}
		metrics.AddPolicyViolation(violation)
		rm.recordViolation(violation)
	}

	return nil
}

// UpdateCPUUsage updates CPU usage for a module
func (rm *ResourceMonitor) UpdateCPUUsage(sessionID, moduleName string, cpuTime time.Duration) error {
	metrics, err := rm.getModuleMetrics(sessionID, moduleName)
	if err != nil {
		return err
	}

	rm.modulesMutex.Lock()
	defer rm.modulesMutex.Unlock()

	metrics.CPU.Used = cpuTime
	metrics.CPU.Timestamp = time.Now()

	// Check for CPU limit violations
	if metrics.CPU.IsCPULimitExceeded() {
		violation := PolicyViolation{
			Type:        "cpu_limit_exceeded",
			Description: fmt.Sprintf("CPU time %v exceeds limit %v", cpuTime, metrics.CPU.Limit),
			SessionID:   sessionID,
			ModuleName:  moduleName,
			Details: map[string]interface{}{
				"used":  cpuTime.Milliseconds(),
				"limit": metrics.CPU.Limit.Milliseconds(),
			},
			Timestamp: time.Now(),
			Severity:  SecuritySeverityError,
		}
		metrics.AddPolicyViolation(violation)
		rm.recordViolation(violation)
	}

	return nil
}

// RecordNetworkActivity records network activity for a module
func (rm *ResourceMonitor) RecordNetworkActivity(sessionID, moduleName string, bytesSent, bytesReceived uint64) error {
	metrics, err := rm.getModuleMetrics(sessionID, moduleName)
	if err != nil {
		return err
	}

	rm.modulesMutex.Lock()
	defer rm.modulesMutex.Unlock()

	metrics.Network.RequestCount++
	metrics.Network.BytesSent += int64(bytesSent)
	metrics.Network.BytesReceived += int64(bytesReceived)
	metrics.Network.Timestamp = time.Now()

	return nil
}

// RecordFileSystemActivity records file system activity for a module
func (rm *ResourceMonitor) RecordFileSystemActivity(sessionID, moduleName string, operation string, bytesTransferred uint64) error {
	metrics, err := rm.getModuleMetrics(sessionID, moduleName)
	if err != nil {
		return err
	}

	rm.modulesMutex.Lock()
	defer rm.modulesMutex.Unlock()

	switch operation {
	case "read":
		metrics.FileSystem.ReadOperations++
		metrics.FileSystem.BytesRead += int64(bytesTransferred)
	case "write":
		metrics.FileSystem.WriteOperations++
		metrics.FileSystem.BytesWritten += int64(bytesTransferred)
	}
	metrics.FileSystem.Timestamp = time.Now()

	return nil
}

// GetModuleMetrics returns the current metrics for a module
func (rm *ResourceMonitor) GetModuleMetrics(sessionID, moduleName string) (*RuntimeMetrics, error) {
	return rm.getModuleMetrics(sessionID, moduleName)
}

// GetAllMetrics returns metrics for all active modules
func (rm *ResourceMonitor) GetAllMetrics() map[string]*RuntimeMetrics {
	rm.modulesMutex.RLock()
	defer rm.modulesMutex.RUnlock()

	result := make(map[string]*RuntimeMetrics)
	for key, metrics := range rm.activeModules {
		// Create a copy to avoid race conditions
		metricsCopy := *metrics
		result[key] = &metricsCopy
	}
	return result
}

// GetSessionMetrics returns metrics for all modules in a session
func (rm *ResourceMonitor) GetSessionMetrics(sessionID string) map[string]*RuntimeMetrics {
	rm.modulesMutex.RLock()
	defer rm.modulesMutex.RUnlock()

	result := make(map[string]*RuntimeMetrics)
	for key, metrics := range rm.activeModules {
		if metrics.SessionID == sessionID {
			metricsCopy := *metrics
			result[key] = &metricsCopy
		}
	}
	return result
}

// EnforceResourceLimits checks and enforces resource limits for all active modules
func (rm *ResourceMonitor) EnforceResourceLimits() []PolicyViolation {
	rm.modulesMutex.RLock()
	defer rm.modulesMutex.RUnlock()

	var allViolations []PolicyViolation

	for _, metrics := range rm.activeModules {
		if rm.validator != nil {
			violations := rm.validator.EnforceResourceLimits(metrics.SessionID, metrics)
			allViolations = append(allViolations, violations...)
		}
	}

	return allViolations
}

// GetResourceSummary returns a summary of resource usage across all modules
func (rm *ResourceMonitor) GetResourceSummary() *ResourceSummary {
	rm.modulesMutex.RLock()
	defer rm.modulesMutex.RUnlock()

	summary := &ResourceSummary{
		TotalModules:     len(rm.activeModules),
		TotalMemoryUsed:  0,
		TotalCPUTime:     0,
		NetworkRequests:  0,
		FileOperations:   0,
		PolicyViolations: 0,
		Timestamp:        time.Now(),
	}

	for _, metrics := range rm.activeModules {
		if metrics.Memory != nil {
			summary.TotalMemoryUsed += uint64(metrics.Memory.Used)
		}
		if metrics.CPU != nil {
			summary.TotalCPUTime += metrics.CPU.Used
		}
		if metrics.Network != nil {
			summary.NetworkRequests += metrics.Network.RequestCount
		}
		if metrics.FileSystem != nil {
			summary.FileOperations += metrics.FileSystem.ReadOperations + metrics.FileSystem.WriteOperations
		}
		summary.PolicyViolations += len(metrics.PolicyViolations)
	}

	return summary
}

// monitoringLoop runs the continuous monitoring process
func (rm *ResourceMonitor) monitoringLoop(ctx context.Context, interval time.Duration) {
	defer rm.monitoringWG.Done()

	ticker := time.NewTicker(interval)
	defer ticker.Stop()

	for {
		select {
		case <-ctx.Done():
			return
		case <-rm.monitoringStop:
			return
		case <-ticker.C:
			rm.performMonitoringCheck()
		}
	}
}

// performMonitoringCheck performs a monitoring check on all active modules
func (rm *ResourceMonitor) performMonitoringCheck() {
	violations := rm.EnforceResourceLimits()

	if len(violations) > 0 {
		rm.logger.Warn("resource limit violations detected", "count", len(violations))

		for _, violation := range violations {
			rm.recordViolation(violation)
		}
	}

	// Record monitoring metrics
	if rm.metrics != nil {
		summary := rm.GetResourceSummary()
		rm.metrics.RecordSecurityEvent("monitoring_check", map[string]interface{}{
			"total_modules":     summary.TotalModules,
			"total_memory_used": summary.TotalMemoryUsed,
			"total_cpu_time":    summary.TotalCPUTime.Milliseconds(),
			"policy_violations": summary.PolicyViolations,
		})
	}
}

// Helper methods

func (rm *ResourceMonitor) getModuleMetrics(sessionID, moduleName string) (*RuntimeMetrics, error) {
	rm.modulesMutex.RLock()
	defer rm.modulesMutex.RUnlock()

	moduleKey := fmt.Sprintf("%s:%s", sessionID, moduleName)
	metrics, exists := rm.activeModules[moduleKey]
	if !exists {
		return nil, fmt.Errorf("module %s not found for session %s", moduleName, sessionID)
	}
	return metrics, nil
}

func (rm *ResourceMonitor) recordViolation(violation PolicyViolation) {
	if rm.metrics != nil {
		rm.metrics.RecordSecurityEvent("policy_violation", map[string]interface{}{
			"type":        violation.Type,
			"session_id":  violation.SessionID,
			"module_name": violation.ModuleName,
			"severity":    string(violation.Severity),
		})
	}

	if rm.logger != nil {
		rm.logger.Error("policy violation recorded",
			"type", violation.Type,
			"session_id", violation.SessionID,
			"module_name", violation.ModuleName,
			"severity", string(violation.Severity),
			"description", violation.Description,
		)
	}
}

// ResourceSummary provides an overview of resource usage
type ResourceSummary struct {
	TotalModules     int           `json:"total_modules"`
	TotalMemoryUsed  uint64        `json:"total_memory_used"`
	TotalCPUTime     time.Duration `json:"total_cpu_time"`
	NetworkRequests  int           `json:"network_requests"`
	FileOperations   int           `json:"file_operations"`
	PolicyViolations int           `json:"policy_violations"`
	Timestamp        time.Time     `json:"timestamp"`
}
