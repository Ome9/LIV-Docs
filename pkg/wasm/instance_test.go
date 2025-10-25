package wasm

import (
	"context"
	"testing"
	"time"
)

func TestWASMInstance_Initialize(t *testing.T) {
	logger := &MockLogger{}
	metrics := &MockMetricsCollector{}

	instance := &WASMInstance{
		name:        "test-instance",
		data:        []byte{0x00, 0x61, 0x73, 0x6d, 0x01, 0x00, 0x00, 0x00},
		memoryLimit: 64 * 1024 * 1024,
		logger:      logger,
		metrics:     metrics,
		createdAt:   time.Now(),
	}

	ctx := context.Background()
	err := instance.initialize(ctx)

	if err != nil {
		t.Fatalf("initialize failed: %v", err)
	}

	if instance.memory == nil {
		t.Error("memory not initialized")
	}

	if instance.memoryUsage == 0 {
		t.Error("memory usage should be non-zero after initialization")
	}

	if len(instance.exports) == 0 {
		t.Error("exports should be populated after initialization")
	}

	if len(instance.functions) == 0 {
		t.Error("functions should be populated after initialization")
	}
}

func TestWASMInstance_Initialize_Terminated(t *testing.T) {
	logger := &MockLogger{}
	metrics := &MockMetricsCollector{}

	instance := &WASMInstance{
		name:       "test-instance",
		terminated: true,
		logger:     logger,
		metrics:    metrics,
	}

	ctx := context.Background()
	err := instance.initialize(ctx)

	if err == nil {
		t.Error("initialize should fail for terminated instance")
	}
}

func TestWASMInstance_Call_ValidFunction(t *testing.T) {
	logger := &MockLogger{}
	metrics := &MockMetricsCollector{}

	instance := &WASMInstance{
		name:        "test-instance",
		data:        []byte{0x00, 0x61, 0x73, 0x6d, 0x01, 0x00, 0x00, 0x00},
		memoryLimit: 64 * 1024 * 1024,
		logger:      logger,
		metrics:     metrics,
		createdAt:   time.Now(),
	}

	ctx := context.Background()
	err := instance.initialize(ctx)
	if err != nil {
		t.Fatalf("initialize failed: %v", err)
	}

	// Call a valid function
	result, err := instance.Call(ctx, "main")
	if err != nil {
		t.Errorf("Call failed: %v", err)
	}

	if result == nil {
		t.Error("Call should return a result")
	}

	// Check that call count increased
	if instance.callCount != 1 {
		t.Errorf("expected call count 1, got %d", instance.callCount)
	}

	// Check that last call time was updated
	if instance.lastCall.IsZero() {
		t.Error("last call time should be updated")
	}
}

func TestWASMInstance_Call_InvalidFunction(t *testing.T) {
	logger := &MockLogger{}
	metrics := &MockMetricsCollector{}

	instance := &WASMInstance{
		name:        "test-instance",
		data:        []byte{0x00, 0x61, 0x73, 0x6d, 0x01, 0x00, 0x00, 0x00},
		memoryLimit: 64 * 1024 * 1024,
		logger:      logger,
		metrics:     metrics,
		createdAt:   time.Now(),
	}

	ctx := context.Background()
	err := instance.initialize(ctx)
	if err != nil {
		t.Fatalf("initialize failed: %v", err)
	}

	// Call an invalid function
	_, err = instance.Call(ctx, "nonexistent_function")
	if err == nil {
		t.Error("Call should fail for nonexistent function")
	}
}

func TestWASMInstance_Call_Terminated(t *testing.T) {
	logger := &MockLogger{}
	metrics := &MockMetricsCollector{}

	instance := &WASMInstance{
		name:        "test-instance",
		terminated:  true,
		logger:      logger,
		metrics:     metrics,
	}

	ctx := context.Background()
	_, err := instance.Call(ctx, "main")

	if err == nil {
		t.Error("Call should fail for terminated instance")
	}
}

func TestWASMInstance_Call_WithArguments(t *testing.T) {
	logger := &MockLogger{}
	metrics := &MockMetricsCollector{}

	instance := &WASMInstance{
		name:        "test-instance",
		data:        []byte{0x00, 0x61, 0x73, 0x6d, 0x01, 0x00, 0x00, 0x00},
		memoryLimit: 64 * 1024 * 1024,
		logger:      logger,
		metrics:     metrics,
		createdAt:   time.Now(),
	}

	ctx := context.Background()
	err := instance.initialize(ctx)
	if err != nil {
		t.Fatalf("initialize failed: %v", err)
	}

	// Call function with arguments
	result, err := instance.Call(ctx, "process", "test_data", 42)
	if err != nil {
		t.Errorf("Call with arguments failed: %v", err)
	}

	if result == nil {
		t.Error("Call should return a result")
	}

	// Verify result contains processed data
	resultMap, ok := result.(map[string]interface{})
	if !ok {
		t.Error("result should be a map")
	}

	if resultMap["processed"] != "test_data" {
		t.Error("result should contain processed argument")
	}
}

func TestWASMInstance_GetExports(t *testing.T) {
	logger := &MockLogger{}
	metrics := &MockMetricsCollector{}

	instance := &WASMInstance{
		name:        "test-instance",
		data:        []byte{0x00, 0x61, 0x73, 0x6d, 0x01, 0x00, 0x00, 0x00},
		memoryLimit: 64 * 1024 * 1024,
		logger:      logger,
		metrics:     metrics,
		createdAt:   time.Now(),
	}

	ctx := context.Background()
	err := instance.initialize(ctx)
	if err != nil {
		t.Fatalf("initialize failed: %v", err)
	}

	exports := instance.GetExports()
	if len(exports) == 0 {
		t.Error("GetExports should return non-empty list")
	}

	// Check for expected default exports
	expectedExports := []string{"main", "init", "process", "render", "update", "cleanup"}
	for _, expected := range expectedExports {
		found := false
		for _, export := range exports {
			if export == expected {
				found = true
				break
			}
		}
		if !found {
			t.Errorf("expected export '%s' not found", expected)
		}
	}
}

func TestWASMInstance_GetExports_Terminated(t *testing.T) {
	logger := &MockLogger{}
	metrics := &MockMetricsCollector{}

	instance := &WASMInstance{
		name:       "test-instance",
		terminated: true,
		logger:     logger,
		metrics:    metrics,
	}

	exports := instance.GetExports()
	if len(exports) != 0 {
		t.Error("GetExports should return empty list for terminated instance")
	}
}

func TestWASMInstance_GetMemoryUsage(t *testing.T) {
	logger := &MockLogger{}
	metrics := &MockMetricsCollector{}

	instance := &WASMInstance{
		name:        "test-instance",
		data:        []byte{0x00, 0x61, 0x73, 0x6d, 0x01, 0x00, 0x00, 0x00},
		memoryLimit: 64 * 1024 * 1024,
		logger:      logger,
		metrics:     metrics,
		createdAt:   time.Now(),
	}

	ctx := context.Background()
	err := instance.initialize(ctx)
	if err != nil {
		t.Fatalf("initialize failed: %v", err)
	}

	usage := instance.GetMemoryUsage()
	if usage == 0 {
		t.Error("memory usage should be non-zero after initialization")
	}

	if usage > instance.memoryLimit {
		t.Errorf("memory usage %d exceeds limit %d", usage, instance.memoryLimit)
	}
}

func TestWASMInstance_SetMemoryLimit(t *testing.T) {
	logger := &MockLogger{}
	metrics := &MockMetricsCollector{}

	instance := &WASMInstance{
		name:        "test-instance",
		data:        []byte{0x00, 0x61, 0x73, 0x6d, 0x01, 0x00, 0x00, 0x00},
		memoryLimit: 64 * 1024 * 1024,
		logger:      logger,
		metrics:     metrics,
		createdAt:   time.Now(),
	}

	ctx := context.Background()
	err := instance.initialize(ctx)
	if err != nil {
		t.Fatalf("initialize failed: %v", err)
	}

	// Set a higher memory limit
	newLimit := uint64(128 * 1024 * 1024)
	err = instance.SetMemoryLimit(newLimit)
	if err != nil {
		t.Errorf("SetMemoryLimit failed: %v", err)
	}

	if instance.memoryLimit != newLimit {
		t.Errorf("memory limit not updated: expected %d, got %d", newLimit, instance.memoryLimit)
	}
}

func TestWASMInstance_SetMemoryLimit_TooLow(t *testing.T) {
	logger := &MockLogger{}
	metrics := &MockMetricsCollector{}

	instance := &WASMInstance{
		name:        "test-instance",
		data:        []byte{0x00, 0x61, 0x73, 0x6d, 0x01, 0x00, 0x00, 0x00},
		memoryLimit: 64 * 1024 * 1024,
		logger:      logger,
		metrics:     metrics,
		createdAt:   time.Now(),
	}

	ctx := context.Background()
	err := instance.initialize(ctx)
	if err != nil {
		t.Fatalf("initialize failed: %v", err)
	}

	// Try to set limit lower than current usage
	err = instance.SetMemoryLimit(1024) // 1KB - too low
	if err == nil {
		t.Error("SetMemoryLimit should fail when limit is lower than current usage")
	}
}

func TestWASMInstance_SetMemoryLimit_Terminated(t *testing.T) {
	logger := &MockLogger{}
	metrics := &MockMetricsCollector{}

	instance := &WASMInstance{
		name:       "test-instance",
		terminated: true,
		logger:     logger,
		metrics:    metrics,
	}

	err := instance.SetMemoryLimit(128 * 1024 * 1024)
	if err == nil {
		t.Error("SetMemoryLimit should fail for terminated instance")
	}
}

func TestWASMInstance_Terminate(t *testing.T) {
	logger := &MockLogger{}
	metrics := &MockMetricsCollector{}

	instance := &WASMInstance{
		name:        "test-instance",
		data:        []byte{0x00, 0x61, 0x73, 0x6d, 0x01, 0x00, 0x00, 0x00},
		memoryLimit: 64 * 1024 * 1024,
		logger:      logger,
		metrics:     metrics,
		createdAt:   time.Now(),
	}

	ctx := context.Background()
	err := instance.initialize(ctx)
	if err != nil {
		t.Fatalf("initialize failed: %v", err)
	}

	// Terminate the instance
	err = instance.Terminate()
	if err != nil {
		t.Errorf("Terminate failed: %v", err)
	}

	if !instance.terminated {
		t.Error("instance should be marked as terminated")
	}

	if instance.memory != nil {
		t.Error("memory should be cleaned up after termination")
	}

	if instance.globals != nil {
		t.Error("globals should be cleaned up after termination")
	}

	if instance.functions != nil {
		t.Error("functions should be cleaned up after termination")
	}
}

func TestWASMInstance_Terminate_AlreadyTerminated(t *testing.T) {
	logger := &MockLogger{}
	metrics := &MockMetricsCollector{}

	instance := &WASMInstance{
		name:       "test-instance",
		terminated: true,
		logger:     logger,
		metrics:    metrics,
	}

	err := instance.Terminate()
	if err == nil {
		t.Error("Terminate should fail for already terminated instance")
	}
}

func TestWASMInstance_AllocateMemory(t *testing.T) {
	logger := &MockLogger{}
	metrics := &MockMetricsCollector{}

	instance := &WASMInstance{
		name:        "test-instance",
		data:        []byte{0x00, 0x61, 0x73, 0x6d, 0x01, 0x00, 0x00, 0x00},
		memoryLimit: 128 * 1024 * 1024,
		logger:      logger,
		metrics:     metrics,
		createdAt:   time.Now(),
	}

	ctx := context.Background()
	err := instance.initialize(ctx)
	if err != nil {
		t.Fatalf("initialize failed: %v", err)
	}

	initialUsage := instance.memoryUsage

	// Allocate additional memory
	allocSize := uint64(1024)
	err = instance.AllocateMemory(allocSize)
	if err != nil {
		t.Errorf("AllocateMemory failed: %v", err)
	}

	expectedUsage := initialUsage + allocSize
	if instance.memoryUsage != expectedUsage {
		t.Errorf("memory usage not updated correctly: expected %d, got %d", expectedUsage, instance.memoryUsage)
	}
}

func TestWASMInstance_AllocateMemory_ExceedsLimit(t *testing.T) {
	logger := &MockLogger{}
	metrics := &MockMetricsCollector{}

	instance := &WASMInstance{
		name:        "test-instance",
		data:        []byte{0x00, 0x61, 0x73, 0x6d, 0x01, 0x00, 0x00, 0x00},
		memoryLimit: 64 * 1024, // Small limit for testing
		logger:      logger,
		metrics:     metrics,
		createdAt:   time.Now(),
	}

	ctx := context.Background()
	err := instance.initialize(ctx)
	if err != nil {
		t.Fatalf("initialize failed: %v", err)
	}

	// Try to allocate more than the limit allows
	allocSize := instance.memoryLimit
	err = instance.AllocateMemory(allocSize)
	if err == nil {
		t.Error("AllocateMemory should fail when exceeding limit")
	}
}

func TestWASMInstance_ReadWriteMemory(t *testing.T) {
	logger := &MockLogger{}
	metrics := &MockMetricsCollector{}

	instance := &WASMInstance{
		name:        "test-instance",
		data:        []byte{0x00, 0x61, 0x73, 0x6d, 0x01, 0x00, 0x00, 0x00},
		memoryLimit: 64 * 1024 * 1024,
		logger:      logger,
		metrics:     metrics,
		createdAt:   time.Now(),
	}

	ctx := context.Background()
	err := instance.initialize(ctx)
	if err != nil {
		t.Fatalf("initialize failed: %v", err)
	}

	// Write data to memory
	testData := []byte("Hello, WASM!")
	err = instance.WriteMemory(0, testData)
	if err != nil {
		t.Errorf("WriteMemory failed: %v", err)
	}

	// Read data back from memory
	readData, err := instance.ReadMemory(0, uint64(len(testData)))
	if err != nil {
		t.Errorf("ReadMemory failed: %v", err)
	}

	if string(readData) != string(testData) {
		t.Errorf("read data doesn't match written data: expected %s, got %s", testData, readData)
	}
}

func TestWASMInstance_ReadMemory_OutOfBounds(t *testing.T) {
	logger := &MockLogger{}
	metrics := &MockMetricsCollector{}

	instance := &WASMInstance{
		name:        "test-instance",
		data:        []byte{0x00, 0x61, 0x73, 0x6d, 0x01, 0x00, 0x00, 0x00},
		memoryLimit: 64 * 1024 * 1024,
		logger:      logger,
		metrics:     metrics,
		createdAt:   time.Now(),
	}

	ctx := context.Background()
	err := instance.initialize(ctx)
	if err != nil {
		t.Fatalf("initialize failed: %v", err)
	}

	// Try to read beyond memory bounds
	_, err = instance.ReadMemory(uint64(len(instance.memory)), 1024)
	if err == nil {
		t.Error("ReadMemory should fail for out of bounds access")
	}
}

func TestWASMInstance_WriteMemory_OutOfBounds(t *testing.T) {
	logger := &MockLogger{}
	metrics := &MockMetricsCollector{}

	instance := &WASMInstance{
		name:        "test-instance",
		data:        []byte{0x00, 0x61, 0x73, 0x6d, 0x01, 0x00, 0x00, 0x00},
		memoryLimit: 64 * 1024 * 1024,
		logger:      logger,
		metrics:     metrics,
		createdAt:   time.Now(),
	}

	ctx := context.Background()
	err := instance.initialize(ctx)
	if err != nil {
		t.Fatalf("initialize failed: %v", err)
	}

	// Try to write beyond memory bounds
	testData := []byte("test")
	err = instance.WriteMemory(uint64(len(instance.memory)), testData)
	if err == nil {
		t.Error("WriteMemory should fail for out of bounds access")
	}
}

func TestWASMInstance_GetInstanceInfo(t *testing.T) {
	logger := &MockLogger{}
	metrics := &MockMetricsCollector{}

	instance := &WASMInstance{
		name:        "test-instance",
		data:        []byte{0x00, 0x61, 0x73, 0x6d, 0x01, 0x00, 0x00, 0x00},
		memoryLimit: 64 * 1024 * 1024,
		logger:      logger,
		metrics:     metrics,
		createdAt:   time.Now(),
	}

	ctx := context.Background()
	err := instance.initialize(ctx)
	if err != nil {
		t.Fatalf("initialize failed: %v", err)
	}

	info := instance.GetInstanceInfo()
	if info == nil {
		t.Fatal("GetInstanceInfo returned nil")
	}

	if info["name"] != instance.name {
		t.Errorf("expected name %s, got %v", instance.name, info["name"])
	}

	if info["memory_usage"] != instance.memoryUsage {
		t.Errorf("expected memory usage %d, got %v", instance.memoryUsage, info["memory_usage"])
	}

	if info["terminated"] != instance.terminated {
		t.Errorf("expected terminated %v, got %v", instance.terminated, info["terminated"])
	}
}

func TestWASMInstance_GetFunctionInfo(t *testing.T) {
	logger := &MockLogger{}
	metrics := &MockMetricsCollector{}

	instance := &WASMInstance{
		name:        "test-instance",
		data:        []byte{0x00, 0x61, 0x73, 0x6d, 0x01, 0x00, 0x00, 0x00},
		memoryLimit: 64 * 1024 * 1024,
		logger:      logger,
		metrics:     metrics,
		createdAt:   time.Now(),
	}

	ctx := context.Background()
	err := instance.initialize(ctx)
	if err != nil {
		t.Fatalf("initialize failed: %v", err)
	}

	// Get info for existing function
	funcInfo, err := instance.GetFunctionInfo("main")
	if err != nil {
		t.Errorf("GetFunctionInfo failed: %v", err)
	}

	if funcInfo.Name != "main" {
		t.Errorf("expected function name 'main', got %s", funcInfo.Name)
	}

	// Try to get info for nonexistent function
	_, err = instance.GetFunctionInfo("nonexistent")
	if err == nil {
		t.Error("GetFunctionInfo should fail for nonexistent function")
	}
}

func TestWASMInstance_IsTerminated(t *testing.T) {
	logger := &MockLogger{}
	metrics := &MockMetricsCollector{}

	instance := &WASMInstance{
		name:        "test-instance",
		data:        []byte{0x00, 0x61, 0x73, 0x6d, 0x01, 0x00, 0x00, 0x00},
		memoryLimit: 64 * 1024 * 1024,
		logger:      logger,
		metrics:     metrics,
		createdAt:   time.Now(),
	}

	// Initially not terminated
	if instance.IsTerminated() {
		t.Error("instance should not be terminated initially")
	}

	// After termination
	instance.terminated = true
	if !instance.IsTerminated() {
		t.Error("instance should be terminated after setting terminated flag")
	}
}

func TestWASMInstance_GetCallStats(t *testing.T) {
	logger := &MockLogger{}
	metrics := &MockMetricsCollector{}

	instance := &WASMInstance{
		name:        "test-instance",
		data:        []byte{0x00, 0x61, 0x73, 0x6d, 0x01, 0x00, 0x00, 0x00},
		memoryLimit: 64 * 1024 * 1024,
		logger:      logger,
		metrics:     metrics,
		createdAt:   time.Now(),
	}

	ctx := context.Background()
	err := instance.initialize(ctx)
	if err != nil {
		t.Fatalf("initialize failed: %v", err)
	}

	// Make some function calls
	_, err = instance.Call(ctx, "main")
	if err != nil {
		t.Errorf("Call failed: %v", err)
	}

	_, err = instance.Call(ctx, "init")
	if err != nil {
		t.Errorf("Call failed: %v", err)
	}

	// Get call stats
	stats := instance.GetCallStats()
	if stats == nil {
		t.Fatal("GetCallStats returned nil")
	}

	if stats["total_calls"] != int64(2) {
		t.Errorf("expected 2 total calls, got %v", stats["total_calls"])
	}

	if stats["last_call"].(time.Time).IsZero() {
		t.Error("last call time should not be zero")
	}
}