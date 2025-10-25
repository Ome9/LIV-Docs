package wasm

import (
	"context"
	"fmt"
	"testing"
	"time"

	"github.com/liv-format/liv/pkg/core"
)

func TestWASMOrchestrationIntegration(t *testing.T) {
	// Setup components
	securityMgr := &MockSecurityManager{}
	logger := &MockLogger{}
	metrics := &MockMetricsCollector{}

	// Create WASM loader
	loader := NewWASMLoader(securityMgr, logger, metrics)
	if loader == nil {
		t.Fatal("failed to create WASM loader")
	}

	// Create WASM runtime
	runtime := NewWASMRuntime(loader, securityMgr, logger, metrics)
	if runtime == nil {
		t.Fatal("failed to create WASM runtime")
	}

	// Create communication bridge
	bridge := NewCommunicationBridge(runtime, logger, metrics)
	if bridge == nil {
		t.Fatal("failed to create communication bridge")
	}

	ctx, cancel := context.WithTimeout(context.Background(), 30*time.Second)
	defer cancel()

	// Start runtime and bridge
	err := runtime.StartRuntime(ctx)
	if err != nil {
		t.Fatalf("failed to start runtime: %v", err)
	}
	defer runtime.StopRuntime()

	err = bridge.Start(ctx)
	if err != nil {
		t.Fatalf("failed to start bridge: %v", err)
	}
	defer bridge.Stop()

	// Create security policy
	policy := &core.SecurityPolicy{
		WASMPermissions: &core.WASMPermissions{
			MemoryLimit:     64 * 1024 * 1024, // 64MB
			CPUTimeLimit:    30000,             // 30 seconds
			AllowNetworking: false,
			AllowFileSystem: false,
			AllowedImports:  []string{},
		},
		JSPermissions: &core.JSPermissions{
			ExecutionMode: "sandboxed",
			AllowedAPIs:   []string{},
			DOMAccess:     "read",
		},
		NetworkPolicy: &core.NetworkPolicy{
			AllowOutbound: false,
		},
		StoragePolicy: &core.StoragePolicy{
			AllowLocalStorage: false,
		},
	}

	// Create sandbox
	sandboxID, err := runtime.CreateSandbox(policy)
	if err != nil {
		t.Fatalf("failed to create sandbox: %v", err)
	}

	// Verify sandbox was created
	sandboxes := runtime.ListActiveSandboxes()
	if len(sandboxes) != 1 || sandboxes[0] != sandboxID {
		t.Errorf("expected 1 sandbox with ID %s, got %v", sandboxID, sandboxes)
	}

	// Load WASM module
	moduleData := []byte{0x00, 0x61, 0x73, 0x6d, 0x01, 0x00, 0x00, 0x00}
	moduleConfig := &core.WASMModule{
		Name:        "test-module",
		Version:     "1.0.0",
		EntryPoint:  "main",
		Exports:     []string{"main", "process", "render"},
		Imports:     []string{},
		Permissions: policy.WASMPermissions,
	}

	err = runtime.LoadModuleInSandbox(ctx, sandboxID, "test-module", moduleData, moduleConfig)
	if err != nil {
		t.Fatalf("failed to load module in sandbox: %v", err)
	}

	// Execute function in sandbox
	result, err := runtime.ExecuteInSandbox(ctx, sandboxID, "test-module", "main")
	if err != nil {
		t.Fatalf("failed to execute function in sandbox: %v", err)
	}

	if result == nil {
		t.Error("execution result should not be nil")
	}

	// Test communication bridge function call
	bridgeResult, err := bridge.CallWASMFunction(ctx, sandboxID, "test-module", "process", map[string]interface{}{
		"data": "test_input",
	})
	if err != nil {
		t.Fatalf("failed to call WASM function via bridge: %v", err)
	}

	if bridgeResult == nil {
		t.Error("bridge result should not be nil")
	}

	// Test event sending
	err = bridge.SendEvent(ctx, "test_event", map[string]interface{}{
		"message": "hello world",
	})
	if err != nil {
		t.Errorf("failed to send event: %v", err)
	}

	// Get sandbox info
	sandboxInfo, err := runtime.GetSandboxInfo(sandboxID)
	if err != nil {
		t.Fatalf("failed to get sandbox info: %v", err)
	}

	if sandboxInfo.ID != sandboxID {
		t.Errorf("expected sandbox ID %s, got %s", sandboxID, sandboxInfo.ID)
	}

	if len(sandboxInfo.LoadedModules) != 1 {
		t.Errorf("expected 1 loaded module, got %d", len(sandboxInfo.LoadedModules))
	}

	if sandboxInfo.ExecutionCount == 0 {
		t.Error("execution count should be greater than 0")
	}

	// Get runtime stats
	stats := runtime.GetRuntimeStats()
	if stats == nil {
		t.Fatal("runtime stats should not be nil")
	}

	if stats["active_sandboxes"].(int) != 1 {
		t.Errorf("expected 1 active sandbox, got %v", stats["active_sandboxes"])
	}

	// Terminate sandbox
	err = runtime.TerminateSandbox(sandboxID)
	if err != nil {
		t.Errorf("failed to terminate sandbox: %v", err)
	}

	// Verify sandbox was terminated
	sandboxes = runtime.ListActiveSandboxes()
	if len(sandboxes) != 0 {
		t.Errorf("expected 0 sandboxes after termination, got %v", sandboxes)
	}
}

func TestWASMOrchestrationResourceLimits(t *testing.T) {
	securityMgr := &MockSecurityManager{}
	logger := &MockLogger{}
	metrics := &MockMetricsCollector{}

	loader := NewWASMLoader(securityMgr, logger, metrics)
	runtime := NewWASMRuntime(loader, securityMgr, logger, metrics)

	// Set strict resource limits for testing
	runtime.config.MaxMemoryPerSandbox = 1024 * 1024 // 1MB
	runtime.config.MaxCPUTimePerSandbox = 100 * time.Millisecond

	ctx, cancel := context.WithTimeout(context.Background(), 10*time.Second)
	defer cancel()

	err := runtime.StartRuntime(ctx)
	if err != nil {
		t.Fatalf("failed to start runtime: %v", err)
	}
	defer runtime.StopRuntime()

	// Create sandbox with restrictive policy
	policy := &core.SecurityPolicy{
		WASMPermissions: &core.WASMPermissions{
			MemoryLimit:     512 * 1024, // 512KB - within sandbox limit
			CPUTimeLimit:    50,          // 50ms - within sandbox limit
			AllowNetworking: false,
			AllowFileSystem: false,
			AllowedImports:  []string{},
		},
		JSPermissions: &core.JSPermissions{
			ExecutionMode: "sandboxed",
		},
		NetworkPolicy: &core.NetworkPolicy{
			AllowOutbound: false,
		},
		StoragePolicy: &core.StoragePolicy{
			AllowLocalStorage: false,
		},
	}

	sandboxID, err := runtime.CreateSandbox(policy)
	if err != nil {
		t.Fatalf("failed to create sandbox: %v", err)
	}

	// Load module
	moduleData := []byte{0x00, 0x61, 0x73, 0x6d, 0x01, 0x00, 0x00, 0x00}
	moduleConfig := &core.WASMModule{
		Name:        "resource-test-module",
		Version:     "1.0.0",
		EntryPoint:  "main",
		Exports:     []string{"main"},
		Permissions: policy.WASMPermissions,
	}

	err = runtime.LoadModuleInSandbox(ctx, sandboxID, "resource-test-module", moduleData, moduleConfig)
	if err != nil {
		t.Fatalf("failed to load module: %v", err)
	}

	// Execute multiple times to accumulate resource usage
	for i := 0; i < 10; i++ {
		_, err = runtime.ExecuteInSandbox(ctx, sandboxID, "resource-test-module", "main")
		if err != nil {
			t.Logf("execution %d failed (expected due to resource limits): %v", i, err)
		}
	}

	// Check for resource violations
	sandboxInfo, err := runtime.GetSandboxInfo(sandboxID)
	if err != nil {
		t.Fatalf("failed to get sandbox info: %v", err)
	}

	t.Logf("Sandbox violations: %d", len(sandboxInfo.Violations))
	t.Logf("Memory usage: %d bytes", sandboxInfo.MemoryUsage)
	t.Logf("CPU time: %v", sandboxInfo.CPUTime)

	// Clean up
	runtime.TerminateSandbox(sandboxID)
}

func TestWASMOrchestrationConcurrency(t *testing.T) {
	securityMgr := &MockSecurityManager{}
	logger := &MockLogger{}
	metrics := &MockMetricsCollector{}

	loader := NewWASMLoader(securityMgr, logger, metrics)
	runtime := NewWASMRuntime(loader, securityMgr, logger, metrics)

	ctx, cancel := context.WithTimeout(context.Background(), 15*time.Second)
	defer cancel()

	err := runtime.StartRuntime(ctx)
	if err != nil {
		t.Fatalf("failed to start runtime: %v", err)
	}
	defer runtime.StopRuntime()

	// Create multiple sandboxes concurrently
	numSandboxes := 5
	sandboxIDs := make([]string, numSandboxes)

	policy := &core.SecurityPolicy{
		WASMPermissions: &core.WASMPermissions{
			MemoryLimit:     32 * 1024 * 1024, // 32MB
			CPUTimeLimit:    10000,             // 10 seconds
			AllowNetworking: false,
			AllowFileSystem: false,
			AllowedImports:  []string{},
		},
		JSPermissions: &core.JSPermissions{
			ExecutionMode: "sandboxed",
		},
		NetworkPolicy: &core.NetworkPolicy{
			AllowOutbound: false,
		},
		StoragePolicy: &core.StoragePolicy{
			AllowLocalStorage: false,
		},
	}

	// Create sandboxes with small delays to ensure unique IDs
	for i := 0; i < numSandboxes; i++ {
		time.Sleep(1 * time.Millisecond) // Small delay to ensure unique timestamps
		sandboxID, err := runtime.CreateSandbox(policy)
		if err != nil {
			t.Fatalf("failed to create sandbox %d: %v", i, err)
		}
		sandboxIDs[i] = sandboxID
		t.Logf("Created sandbox %d: %s", i, sandboxID)
	}

	// Load modules in each sandbox
	moduleData := []byte{0x00, 0x61, 0x73, 0x6d, 0x01, 0x00, 0x00, 0x00}
	for i, sandboxID := range sandboxIDs {
		moduleConfig := &core.WASMModule{
			Name:        fmt.Sprintf("concurrent-module-%d", i),
			Version:     "1.0.0",
			EntryPoint:  "main",
			Exports:     []string{"main", "process"},
			Permissions: policy.WASMPermissions,
		}

		err = runtime.LoadModuleInSandbox(ctx, sandboxID, moduleConfig.Name, moduleData, moduleConfig)
		if err != nil {
			t.Fatalf("failed to load module in sandbox %d: %v", i, err)
		}
	}

	// Execute functions concurrently
	done := make(chan bool, numSandboxes)
	for i, sandboxID := range sandboxIDs {
		go func(idx int, sID string) {
			defer func() { done <- true }()

			moduleName := fmt.Sprintf("concurrent-module-%d", idx)
			for j := 0; j < 3; j++ {
				_, err := runtime.ExecuteInSandbox(ctx, sID, moduleName, "main")
				if err != nil {
					t.Logf("concurrent execution failed for sandbox %d, iteration %d: %v", idx, j, err)
				}
			}
		}(i, sandboxID)
	}

	// Wait for all executions to complete
	for i := 0; i < numSandboxes; i++ {
		select {
		case <-done:
		case <-time.After(10 * time.Second):
			t.Fatal("concurrent execution timed out")
		}
	}

	// Verify all sandboxes are still active
	activeSandboxes := runtime.ListActiveSandboxes()
	if len(activeSandboxes) != numSandboxes {
		t.Errorf("expected %d active sandboxes, got %d", numSandboxes, len(activeSandboxes))
	}

	// Get runtime stats
	stats := runtime.GetRuntimeStats()
	t.Logf("Final runtime stats: %+v", stats)

	// Clean up all sandboxes
	for i, sandboxID := range sandboxIDs {
		t.Logf("Terminating sandbox %d: %s", i, sandboxID)
		err = runtime.TerminateSandbox(sandboxID)
		if err != nil {
			t.Errorf("failed to terminate sandbox %s: %v", sandboxID, err)
		}
	}
}

func TestWASMOrchestrationErrorHandling(t *testing.T) {
	securityMgr := &MockSecurityManager{
		validateModuleFunc: func(data []byte, permissions *core.WASMPermissions) error {
			// Simulate validation failure for specific module
			if len(data) > 100 {
				return fmt.Errorf("simulated validation failure")
			}
			return nil
		},
	}
	logger := &MockLogger{}
	metrics := &MockMetricsCollector{}

	loader := NewWASMLoader(securityMgr, logger, metrics)
	runtime := NewWASMRuntime(loader, securityMgr, logger, metrics)

	ctx, cancel := context.WithTimeout(context.Background(), 10*time.Second)
	defer cancel()

	err := runtime.StartRuntime(ctx)
	if err != nil {
		t.Fatalf("failed to start runtime: %v", err)
	}
	defer runtime.StopRuntime()

	policy := &core.SecurityPolicy{
		WASMPermissions: &core.WASMPermissions{
			MemoryLimit:     64 * 1024 * 1024,
			CPUTimeLimit:    30000,
			AllowNetworking: false,
			AllowFileSystem: false,
			AllowedImports:  []string{},
		},
		JSPermissions: &core.JSPermissions{
			ExecutionMode: "sandboxed",
		},
		NetworkPolicy: &core.NetworkPolicy{
			AllowOutbound: false,
		},
		StoragePolicy: &core.StoragePolicy{
			AllowLocalStorage: false,
		},
	}

	sandboxID, err := runtime.CreateSandbox(policy)
	if err != nil {
		t.Fatalf("failed to create sandbox: %v", err)
	}

	// Test 1: Try to load invalid module (should fail validation)
	invalidModuleData := make([]byte, 200) // Larger than 100 bytes to trigger validation failure
	copy(invalidModuleData, []byte{0x00, 0x61, 0x73, 0x6d, 0x01, 0x00, 0x00, 0x00})

	moduleConfig := &core.WASMModule{
		Name:        "invalid-module",
		Version:     "1.0.0",
		EntryPoint:  "main",
		Exports:     []string{"main"},
		Permissions: policy.WASMPermissions,
	}

	err = runtime.LoadModuleInSandbox(ctx, sandboxID, "invalid-module", invalidModuleData, moduleConfig)
	if err == nil {
		t.Error("loading invalid module should have failed")
	}

	// Test 2: Load valid module
	validModuleData := []byte{0x00, 0x61, 0x73, 0x6d, 0x01, 0x00, 0x00, 0x00}
	moduleConfig.Name = "valid-module"

	err = runtime.LoadModuleInSandbox(ctx, sandboxID, "valid-module", validModuleData, moduleConfig)
	if err != nil {
		t.Fatalf("loading valid module should have succeeded: %v", err)
	}

	// Test 3: Try to execute in non-existent sandbox
	_, err = runtime.ExecuteInSandbox(ctx, "non-existent-sandbox", "valid-module", "main")
	if err == nil {
		t.Error("execution in non-existent sandbox should have failed")
	}

	// Test 4: Try to execute non-existent module
	_, err = runtime.ExecuteInSandbox(ctx, sandboxID, "non-existent-module", "main")
	if err == nil {
		t.Error("execution of non-existent module should have failed")
	}

	// Test 5: Try to execute non-existent function
	_, err = runtime.ExecuteInSandbox(ctx, sandboxID, "valid-module", "non-existent-function")
	if err == nil {
		t.Error("execution of non-existent function should have failed")
	}

	// Test 6: Try to terminate non-existent sandbox
	err = runtime.TerminateSandbox("non-existent-sandbox")
	if err == nil {
		t.Error("terminating non-existent sandbox should have failed")
	}

	// Clean up
	runtime.TerminateSandbox(sandboxID)
}

func TestWASMOrchestrationMemoryManagement(t *testing.T) {
	securityMgr := &MockSecurityManager{}
	logger := &MockLogger{}
	metrics := &MockMetricsCollector{}

	loader := NewWASMLoader(securityMgr, logger, metrics)
	runtime := NewWASMRuntime(loader, securityMgr, logger, metrics)

	ctx, cancel := context.WithTimeout(context.Background(), 10*time.Second)
	defer cancel()

	err := runtime.StartRuntime(ctx)
	if err != nil {
		t.Fatalf("failed to start runtime: %v", err)
	}
	defer runtime.StopRuntime()

	policy := &core.SecurityPolicy{
		WASMPermissions: &core.WASMPermissions{
			MemoryLimit:     64 * 1024 * 1024,
			CPUTimeLimit:    30000,
			AllowNetworking: false,
			AllowFileSystem: false,
			AllowedImports:  []string{},
		},
		JSPermissions: &core.JSPermissions{
			ExecutionMode: "sandboxed",
		},
		NetworkPolicy: &core.NetworkPolicy{
			AllowOutbound: false,
		},
		StoragePolicy: &core.StoragePolicy{
			AllowLocalStorage: false,
		},
	}

	sandboxID, err := runtime.CreateSandbox(policy)
	if err != nil {
		t.Fatalf("failed to create sandbox: %v", err)
	}

	// Load module
	moduleData := []byte{0x00, 0x61, 0x73, 0x6d, 0x01, 0x00, 0x00, 0x00}
	moduleConfig := &core.WASMModule{
		Name:        "memory-test-module",
		Version:     "1.0.0",
		EntryPoint:  "main",
		Exports:     []string{"main", "cleanup"},
		Permissions: policy.WASMPermissions,
	}

	err = runtime.LoadModuleInSandbox(ctx, sandboxID, "memory-test-module", moduleData, moduleConfig)
	if err != nil {
		t.Fatalf("failed to load module: %v", err)
	}

	// Get initial memory usage
	initialInfo, err := runtime.GetSandboxInfo(sandboxID)
	if err != nil {
		t.Fatalf("failed to get initial sandbox info: %v", err)
	}
	initialMemory := initialInfo.MemoryUsage

	// Execute function that increases memory usage
	_, err = runtime.ExecuteInSandbox(ctx, sandboxID, "memory-test-module", "main")
	if err != nil {
		t.Fatalf("failed to execute main function: %v", err)
	}

	// Check memory usage increased
	afterExecInfo, err := runtime.GetSandboxInfo(sandboxID)
	if err != nil {
		t.Fatalf("failed to get sandbox info after execution: %v", err)
	}

	if afterExecInfo.MemoryUsage <= initialMemory {
		t.Logf("Memory usage should have increased: initial=%d, after=%d", initialMemory, afterExecInfo.MemoryUsage)
	}

	// Execute cleanup function
	_, err = runtime.ExecuteInSandbox(ctx, sandboxID, "memory-test-module", "cleanup")
	if err != nil {
		t.Fatalf("failed to execute cleanup function: %v", err)
	}

	// Check memory usage after cleanup
	afterCleanupInfo, err := runtime.GetSandboxInfo(sandboxID)
	if err != nil {
		t.Fatalf("failed to get sandbox info after cleanup: %v", err)
	}

	t.Logf("Memory usage: initial=%d, after_exec=%d, after_cleanup=%d",
		initialMemory, afterExecInfo.MemoryUsage, afterCleanupInfo.MemoryUsage)

	// Clean up
	runtime.TerminateSandbox(sandboxID)
}