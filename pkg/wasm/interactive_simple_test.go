package wasm

import (
	"context"
	"testing"

	"github.com/liv-format/liv/pkg/core"
)

func TestInteractiveEngineBasicIntegration(t *testing.T) {
	logger := &MockLogger{}
	metrics := &MockMetricsCollector{}
	securityMgr := &MockSecurityManager{}

	// Create WASM loader
	loader := NewWASMLoader(securityMgr, logger, metrics)

	// Create WASM runtime
	runtime := NewWASMRuntime(loader, securityMgr, logger, metrics)

	// Create communication bridge
	bridge := NewCommunicationBridge(runtime, logger, metrics)

	ctx := context.Background()

	// Start the communication bridge
	err := bridge.Start(ctx)
	if err != nil {
		t.Fatalf("Failed to start communication bridge: %v", err)
	}
	defer bridge.Stop()

	// Test creating a sandbox
	sandboxID, err := runtime.CreateSandbox(&core.SecurityPolicy{
		WASMPermissions: &core.WASMPermissions{
			MemoryLimit:     1024 * 1024, // 1MB
			CPUTimeLimit:    5000,        // 5 seconds
			AllowNetworking: false,
			AllowFileSystem: false,
			AllowedImports:  []string{"console"},
		},
	})
	if err != nil {
		t.Fatalf("Failed to create sandbox: %v", err)
	}

	t.Logf("Created sandbox: %s", sandboxID)

	// Simulate loading the interactive engine WASM module
	wasmModuleData := []byte{0x00, 0x61, 0x73, 0x6d, 0x01, 0x00, 0x00, 0x00} // WASM header

	// Load the interactive engine module
	instance, err := loader.LoadModule(ctx, "interactive-engine", wasmModuleData)
	if err != nil {
		t.Fatalf("Failed to load WASM module: %v", err)
	}

	t.Logf("Loaded WASM module instance: %v", instance != nil)

	// Test basic communication
	message := &Message{
		Type: MessageTypeFunction,
		Payload: map[string]interface{}{
			"function": "test_function",
			"args":     map[string]interface{}{},
		},
	}

	response, err := bridge.SendMessage(ctx, sandboxID, "interactive-engine", message)
	if err != nil {
		t.Logf("Expected error for non-existent function: %v", err)
	} else {
		t.Logf("Unexpected success: %v", response)
	}

	// Clean up
	err = runtime.TerminateSandbox(sandboxID)
	if err != nil {
		t.Errorf("Failed to terminate sandbox: %v", err)
	}
}

func TestWASMLoaderBasicFunctionality(t *testing.T) {
	logger := &MockLogger{}
	metrics := &MockMetricsCollector{}
	securityMgr := &MockSecurityManager{}

	loader := NewWASMLoader(securityMgr, logger, metrics)

	ctx := context.Background()

	// Test loading a valid WASM module
	wasmModuleData := []byte{0x00, 0x61, 0x73, 0x6d, 0x01, 0x00, 0x00, 0x00} // WASM header

	instance, err := loader.LoadModule(ctx, "test-module", wasmModuleData)
	if err != nil {
		t.Fatalf("Failed to load WASM module: %v", err)
	}

	if instance == nil {
		t.Fatal("Expected non-nil instance")
	}

	// Test module info
	moduleInfo, err := loader.GetModuleInfo("test-module")
	if err != nil {
		t.Fatalf("Failed to get module info: %v", err)
	}

	if moduleInfo.Name != "test-module" {
		t.Errorf("Expected module name 'test-module', got '%s'", moduleInfo.Name)
	}

	// Test listing modules
	modules := loader.ListModules()
	if len(modules) != 1 {
		t.Errorf("Expected 1 module, got %d", len(modules))
	}

	if modules[0] != "test-module" {
		t.Errorf("Expected module 'test-module', got '%s'", modules[0])
	}

	// Test unloading module
	err = loader.UnloadModule("test-module")
	if err != nil {
		t.Fatalf("Failed to unload module: %v", err)
	}

	// Verify module is unloaded
	modules = loader.ListModules()
	if len(modules) != 0 {
		t.Errorf("Expected 0 modules after unload, got %d", len(modules))
	}
}

func TestCommunicationBridgeBasics(t *testing.T) {
	logger := &MockLogger{}
	metrics := &MockMetricsCollector{}
	securityMgr := &MockSecurityManager{}

	loader := NewWASMLoader(securityMgr, logger, metrics)
	runtime := NewWASMRuntime(loader, securityMgr, logger, metrics)
	bridge := NewCommunicationBridge(runtime, logger, metrics)

	ctx := context.Background()

	// Start the bridge
	err := bridge.Start(ctx)
	if err != nil {
		t.Fatalf("Failed to start communication bridge: %v", err)
	}
	defer bridge.Stop()

	// Test message handler registration
	testHandler := func(ctx context.Context, message *Message) (*Message, error) {
		return &Message{
			Type:   MessageTypeResponse,
			Source: "test_handler",
			Payload: map[string]interface{}{
				"result": "success",
			},
		}, nil
	}

	bridge.RegisterMessageHandler("test_message", testHandler)

	// Test event listener registration
	testListener := func(ctx context.Context, event *Message) {
		// Event received
	}

	bridge.RegisterEventListener("test_event", testListener)

	// Test sending an event
	err = bridge.SendEvent(ctx, "test_event", map[string]interface{}{
		"data": "test_data",
	})
	if err != nil {
		t.Fatalf("Failed to send event: %v", err)
	}

	// Note: In a real scenario, we'd need to wait for async processing
	// For this test, we just verify the methods don't error

	t.Logf("Communication bridge test completed successfully")
}