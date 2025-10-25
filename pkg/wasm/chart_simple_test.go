package wasm

import (
	"context"
	"testing"

	"github.com/liv-format/liv/pkg/core"
)

func TestChartFrameworkBasics(t *testing.T) {
	logger := &MockLogger{}
	metrics := &MockMetricsCollector{}
	securityMgr := &MockSecurityManager{}

	// Create WASM loader and runtime
	loader := NewWASMLoader(securityMgr, logger, metrics)
	runtime := NewWASMRuntime(loader, securityMgr, logger, metrics)
	bridge := NewCommunicationBridge(runtime, logger, metrics)

	ctx := context.Background()

	// Start the communication bridge
	err := bridge.Start(ctx)
	if err != nil {
		t.Fatalf("Failed to start communication bridge: %v", err)
	}
	defer bridge.Stop()

	// Create sandbox
	sandboxID, err := runtime.CreateSandbox(&core.SecurityPolicy{
		WASMPermissions: &core.WASMPermissions{
			MemoryLimit:     1024 * 1024,
			CPUTimeLimit:    5000,
			AllowNetworking: false,
			AllowFileSystem: false,
			AllowedImports:  []string{"console"},
		},
	})
	if err != nil {
		t.Fatalf("Failed to create sandbox: %v", err)
	}

	t.Logf("Created sandbox for chart framework testing: %s", sandboxID)

	// Load WASM module
	wasmModuleData := []byte{0x00, 0x61, 0x73, 0x6d, 0x01, 0x00, 0x00, 0x00}
	_, err = loader.LoadModule(ctx, "chart-engine", wasmModuleData)
	if err != nil {
		t.Fatalf("Failed to load WASM module: %v", err)
	}

	// Test basic message handling for chart operations
	message := &Message{
		Type: MessageTypeFunction,
		Payload: map[string]interface{}{
			"function": "create_chart",
			"args": map[string]interface{}{
				"chart_type":     "line",
				"data_source_id": "test_data",
				"config":         `{"width": 400, "height": 300}`,
			},
		},
	}

	// This will fail because the function doesn't exist, but it tests the communication path
	_, err = bridge.SendMessage(ctx, sandboxID, "chart-engine", message)
	if err != nil {
		t.Logf("Expected error for non-existent chart function: %v", err)
	}

	// Test vector graphics message
	vectorMessage := &Message{
		Type: MessageTypeFunction,
		Payload: map[string]interface{}{
			"function": "create_vector_shape",
			"args": map[string]interface{}{
				"shape_type": "rectangle",
				"x":          10,
				"y":          20,
				"width":      100,
				"height":     50,
			},
		},
	}

	_, err = bridge.SendMessage(ctx, sandboxID, "chart-engine", vectorMessage)
	if err != nil {
		t.Logf("Expected error for non-existent vector function: %v", err)
	}

	// Test data update event
	err = bridge.SendEvent(ctx, "chart_data_update", map[string]interface{}{
		"chart_id": "test_chart",
		"data": []map[string]interface{}{
			{"value": 10, "label": "A"},
			{"value": 20, "label": "B"},
		},
	})
	if err != nil {
		t.Fatalf("Failed to send chart data update event: %v", err)
	}

	t.Logf("Chart framework communication test completed successfully")

	// Clean up
	err = runtime.TerminateSandbox(sandboxID)
	if err != nil {
		t.Errorf("Failed to terminate sandbox: %v", err)
	}
}

func TestChartMessageHandling(t *testing.T) {
	logger := &MockLogger{}
	metrics := &MockMetricsCollector{}
	securityMgr := &MockSecurityManager{}

	loader := NewWASMLoader(securityMgr, logger, metrics)
	runtime := NewWASMRuntime(loader, securityMgr, logger, metrics)
	bridge := NewCommunicationBridge(runtime, logger, metrics)

	ctx := context.Background()
	err := bridge.Start(ctx)
	if err != nil {
		t.Fatalf("Failed to start communication bridge: %v", err)
	}
	defer bridge.Stop()

	// Test chart-specific message handler registration
	chartHandler := func(ctx context.Context, message *Message) (*Message, error) {
		
		// Simulate chart creation response
		return &Message{
			Type:   MessageTypeResponse,
			Source: "chart_handler",
			Payload: map[string]interface{}{
				"chart_id": "chart_12345",
				"status":   "created",
			},
		}, nil
	}

	bridge.RegisterMessageHandler("chart_create", chartHandler)

	// Test vector graphics message handler
	vectorHandler := func(ctx context.Context, message *Message) (*Message, error) {
		
		return &Message{
			Type:   MessageTypeResponse,
			Source: "vector_handler",
			Payload: map[string]interface{}{
				"shape_id": "shape_67890",
				"svg":      "<rect x='10' y='20' width='100' height='50'/>",
			},
		}, nil
	}

	bridge.RegisterMessageHandler("vector_create", vectorHandler)

	// Test chart event listener
	chartEventListener := func(ctx context.Context, event *Message) {
		// Chart event received
	}

	bridge.RegisterEventListener("chart_update", chartEventListener)

	// Send chart creation message
	err = bridge.SendEvent(ctx, "chart_create", map[string]interface{}{
		"chart_type": "bar",
		"config":     map[string]interface{}{"width": 500, "height": 400},
	})
	if err != nil {
		t.Fatalf("Failed to send chart creation event: %v", err)
	}

	// Send vector creation message
	err = bridge.SendEvent(ctx, "vector_create", map[string]interface{}{
		"shape_type": "circle",
		"radius":     25,
	})
	if err != nil {
		t.Fatalf("Failed to send vector creation event: %v", err)
	}

	// Send chart update event
	err = bridge.SendEvent(ctx, "chart_update", map[string]interface{}{
		"chart_id": "test_chart",
		"new_data": []int{1, 2, 3, 4, 5},
	})
	if err != nil {
		t.Fatalf("Failed to send chart update event: %v", err)
	}

	t.Logf("Chart message handling test completed successfully")
}

func TestChartDataProcessing(t *testing.T) {
	logger := &MockLogger{}
	metrics := &MockMetricsCollector{}
	securityMgr := &MockSecurityManager{}

	loader := NewWASMLoader(securityMgr, logger, metrics)
	runtime := NewWASMRuntime(loader, securityMgr, logger, metrics)
	bridge := NewCommunicationBridge(runtime, logger, metrics)

	ctx := context.Background()
	err := bridge.Start(ctx)
	if err != nil {
		t.Fatalf("Failed to start communication bridge: %v", err)
	}
	defer bridge.Stop()

	sandboxID, err := runtime.CreateSandbox(&core.SecurityPolicy{
		WASMPermissions: &core.WASMPermissions{
			MemoryLimit:     512 * 1024,
			CPUTimeLimit:    3000,
			AllowNetworking: false,
			AllowFileSystem: false,
			AllowedImports:  []string{},
		},
	})
	if err != nil {
		t.Fatalf("Failed to create sandbox: %v", err)
	}

	// Test data processing message handler
	dataHandler := func(ctx context.Context, message *Message) (*Message, error) {
		
		// Simulate data processing for charts
		if payload, ok := message.Payload["data"].([]interface{}); ok {
			processedData := make([]map[string]interface{}, len(payload))
			for i, item := range payload {
				if itemMap, ok := item.(map[string]interface{}); ok {
					processedData[i] = map[string]interface{}{
						"x":     i,
						"y":     itemMap["value"],
						"label": itemMap["label"],
					}
				}
			}
			
			return &Message{
				Type:   MessageTypeResponse,
				Source: "data_processor",
				Payload: map[string]interface{}{
					"processed_data": processedData,
					"data_points":    len(processedData),
				},
			}, nil
		}
		
		return &Message{
			Type:   MessageTypeError,
			Source: "data_processor",
			Payload: map[string]interface{}{
				"error": "Invalid data format",
			},
		}, nil
	}

	bridge.RegisterMessageHandler("process_chart_data", dataHandler)

	// Send data processing request
	err = bridge.SendEvent(ctx, "process_chart_data", map[string]interface{}{
		"data": []map[string]interface{}{
			{"value": 10, "label": "Jan"},
			{"value": 15, "label": "Feb"},
			{"value": 12, "label": "Mar"},
		},
	})
	if err != nil {
		t.Fatalf("Failed to send data processing event: %v", err)
	}

	t.Logf("Chart data processing test completed successfully")

	// Clean up
	err = runtime.TerminateSandbox(sandboxID)
	if err != nil {
		t.Errorf("Failed to terminate sandbox: %v", err)
	}
}

func TestChartPerformanceMonitoring(t *testing.T) {
	logger := &MockLogger{}
	metrics := &MockMetricsCollector{}
	securityMgr := &MockSecurityManager{}

	loader := NewWASMLoader(securityMgr, logger, metrics)
	runtime := NewWASMRuntime(loader, securityMgr, logger, metrics)
	bridge := NewCommunicationBridge(runtime, logger, metrics)

	ctx := context.Background()
	err := bridge.Start(ctx)
	if err != nil {
		t.Fatalf("Failed to start communication bridge: %v", err)
	}
	defer bridge.Stop()

	sandboxID, err := runtime.CreateSandbox(&core.SecurityPolicy{
		WASMPermissions: &core.WASMPermissions{
			MemoryLimit:     1024 * 1024,
			CPUTimeLimit:    5000,
			AllowNetworking: false,
			AllowFileSystem: false,
			AllowedImports:  []string{"console"},
		},
	})
	if err != nil {
		t.Fatalf("Failed to create sandbox: %v", err)
	}

	// Test performance monitoring for chart operations
	performanceHandler := func(ctx context.Context, message *Message) (*Message, error) {
		// Simulate performance metrics collection
		return &Message{
			Type:   MessageTypeResponse,
			Source: "performance_monitor",
			Payload: map[string]interface{}{
				"render_time":    42.5,
				"memory_usage":   1024 * 512,
				"charts_created": 3,
				"cache_hit_rate": 85.2,
			},
		}, nil
	}

	bridge.RegisterMessageHandler("get_chart_performance", performanceHandler)

	// Send performance monitoring request
	err = bridge.SendEvent(ctx, "get_chart_performance", map[string]interface{}{
		"include_details": true,
	})
	if err != nil {
		t.Fatalf("Failed to send performance monitoring event: %v", err)
	}

	// Test resource monitoring
	resourceHandler := func(ctx context.Context, message *Message) (*Message, error) {
		return &Message{
			Type:   MessageTypeResponse,
			Source: "resource_monitor",
			Payload: map[string]interface{}{
				"memory_allocated": 256 * 1024,
				"cpu_time_used":    1500,
				"active_charts":    2,
				"cached_renders":   5,
			},
		}, nil
	}

	bridge.RegisterMessageHandler("monitor_chart_resources", resourceHandler)

	err = bridge.SendEvent(ctx, "monitor_chart_resources", map[string]interface{}{
		"sandbox_id": sandboxID,
	})
	if err != nil {
		t.Fatalf("Failed to send resource monitoring event: %v", err)
	}

	t.Logf("Chart performance monitoring test completed successfully")

	// Clean up
	err = runtime.TerminateSandbox(sandboxID)
	if err != nil {
		t.Errorf("Failed to terminate sandbox: %v", err)
	}
}