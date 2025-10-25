package main

import (
	"context"
	"encoding/json"
	"fmt"
	"log"
	"time"

	"github.com/tetratelabs/wazero"
	"github.com/tetratelabs/wazero/api"
	"github.com/tetratelabs/wazero/imports/wasi_snapshot_preview1"
)

// WASMPermissions represents the security permissions for the WASM engine
type WASMPermissions struct {
	MemoryLimit         int      `json:"memory_limit"`
	AllowedImports      []string `json:"allowed_imports"`
	CPUTimeLimit        int      `json:"cpu_time_limit"`
	AllowNetworking     bool     `json:"allow_networking"`
	AllowFileSystem     bool     `json:"allow_file_system"`
	AllowedInteractions []string `json:"allowed_interactions"`
	MaxDataSize         int      `json:"max_data_size"`
	MaxElements         int      `json:"max_elements"`
}

// InteractionEvent represents a user interaction event
type InteractionEvent struct {
	EventType     string                 `json:"event_type"`
	TargetElement *string                `json:"target_element"`
	Position      *Position              `json:"position"`
	Data          map[string]interface{} `json:"data"`
	Timestamp     float64                `json:"timestamp"`
}

// Position represents a 2D coordinate
type Position struct {
	X float64 `json:"x"`
	Y float64 `json:"y"`
}

// RenderUpdate represents the output from the WASM engine
type RenderUpdate struct {
	DOMOperations    []DOMOperation    `json:"dom_operations"`
	StyleChanges     []StyleChange     `json:"style_changes"`
	AnimationUpdates []AnimationUpdate `json:"animation_updates"`
	Timestamp        float64           `json:"timestamp"`
}

// DOMOperation represents a DOM manipulation operation
type DOMOperation struct {
	Type       string            `json:"type"`
	ElementID  string            `json:"element_id"`
	Tag        string            `json:"tag,omitempty"`
	ParentID   *string           `json:"parent_id,omitempty"`
	Attributes map[string]string `json:"attributes,omitempty"`
}

// StyleChange represents a CSS style change
type StyleChange struct {
	ElementID string `json:"element_id"`
	Property  string `json:"property"`
	Value     string `json:"value"`
}

// AnimationUpdate represents an animation state update
type AnimationUpdate struct {
	AnimationID   string                 `json:"animation_id"`
	Progress      float64                `json:"progress"`
	CurrentValues map[string]interface{} `json:"current_values"`
}

// LIVInteractiveHost manages the WASM interactive engine
type LIVInteractiveHost struct {
	runtime     wazero.Runtime
	module      api.Module
	ctx         context.Context
	permissions WASMPermissions

	// Function exports from WASM
	initEngine          api.Function
	processInteraction  api.Function
	renderFrame         api.Function
	updateData          api.Function
	getPerformanceStats api.Function
	destroyEngine       api.Function

	// Memory management
	malloc api.Function
	free   api.Function
	memory api.Memory
}

// NewLIVInteractiveHost creates a new interactive host
func NewLIVInteractiveHost(wasmBytes []byte, permissions WASMPermissions) (*LIVInteractiveHost, error) {
	ctx := context.Background()

	// Create WASM runtime with security configuration
	config := wazero.NewRuntimeConfig().
		WithCoreFeatures(api.CoreFeaturesV2).
		WithMemoryLimitPages(uint32(permissions.MemoryLimit / 65536)) // 64KB pages

	runtime := wazero.NewRuntimeWithConfig(ctx, config)

	// Instantiate WASI for basic functionality
	if _, err := wasi_snapshot_preview1.Instantiate(ctx, runtime); err != nil {
		return nil, fmt.Errorf("failed to instantiate WASI: %w", err)
	}

	// Add host functions for secure communication
	hostModule, err := runtime.NewHostModuleBuilder("env").
		NewFunctionBuilder().
		WithFunc(consoleLog).
		Export("console_log").
		NewFunctionBuilder().
		WithFunc(performanceNow).
		Export("performance_now").
		Instantiate(ctx)
	if err != nil {
		return nil, fmt.Errorf("failed to instantiate host module: %w", err)
	}
	defer hostModule.Close(ctx)

	// Compile and instantiate the WASM module
	compiledModule, err := runtime.CompileModule(ctx, wasmBytes)
	if err != nil {
		return nil, fmt.Errorf("failed to compile WASM module: %w", err)
	}

	module, err := runtime.InstantiateModule(ctx, compiledModule, wazero.NewModuleConfig())
	if err != nil {
		return nil, fmt.Errorf("failed to instantiate WASM module: %w", err)
	}

	host := &LIVInteractiveHost{
		runtime:     runtime,
		module:      module,
		ctx:         ctx,
		permissions: permissions,
		memory:      module.Memory(),
	}

	// Get function exports
	if err := host.bindFunctions(); err != nil {
		return nil, fmt.Errorf("failed to bind functions: %w", err)
	}

	return host, nil
}

// bindFunctions binds the exported WASM functions
func (h *LIVInteractiveHost) bindFunctions() error {
	h.initEngine = h.module.ExportedFunction("init_interactive_engine")
	if h.initEngine == nil {
		return fmt.Errorf("init_interactive_engine function not exported")
	}

	h.processInteraction = h.module.ExportedFunction("process_interaction")
	if h.processInteraction == nil {
		return fmt.Errorf("process_interaction function not exported")
	}

	h.renderFrame = h.module.ExportedFunction("render_frame")
	if h.renderFrame == nil {
		return fmt.Errorf("render_frame function not exported")
	}

	h.updateData = h.module.ExportedFunction("update_data")
	if h.updateData == nil {
		return fmt.Errorf("update_data function not exported")
	}

	h.getPerformanceStats = h.module.ExportedFunction("get_performance_stats")
	if h.getPerformanceStats == nil {
		return fmt.Errorf("get_performance_stats function not exported")
	}

	h.destroyEngine = h.module.ExportedFunction("destroy_engine")
	if h.destroyEngine == nil {
		return fmt.Errorf("destroy_engine function not exported")
	}

	// Memory management functions (if available)
	h.malloc = h.module.ExportedFunction("malloc")
	h.free = h.module.ExportedFunction("free")

	return nil
}

// Initialize initializes the WASM interactive engine
func (h *LIVInteractiveHost) Initialize() error {
	// Serialize permissions to JSON
	permissionsJSON, err := json.Marshal(h.permissions)
	if err != nil {
		return fmt.Errorf("failed to serialize permissions: %w", err)
	}

	// Write permissions to WASM memory
	permissionsPtr, err := h.writeStringToMemory(string(permissionsJSON))
	if err != nil {
		return fmt.Errorf("failed to write permissions to memory: %w", err)
	}
	defer h.freeMemory(permissionsPtr)

	// Call init function
	results, err := h.initEngine.Call(h.ctx, uint64(permissionsPtr))
	if err != nil {
		return fmt.Errorf("failed to initialize engine: %w", err)
	}

	// Check result (0 = success, non-zero = error)
	if len(results) > 0 && results[0] != 0 {
		return fmt.Errorf("engine initialization failed with code: %d", results[0])
	}

	log.Println("LIV Interactive Engine initialized successfully")
	return nil
}

// ProcessInteraction processes a user interaction event
func (h *LIVInteractiveHost) ProcessInteraction(event InteractionEvent) (*RenderUpdate, error) {
	// Serialize event to JSON
	eventJSON, err := json.Marshal(event)
	if err != nil {
		return nil, fmt.Errorf("failed to serialize event: %w", err)
	}

	// Write event to WASM memory
	eventPtr, err := h.writeStringToMemory(string(eventJSON))
	if err != nil {
		return nil, fmt.Errorf("failed to write event to memory: %w", err)
	}
	defer h.freeMemory(eventPtr)

	// Call process_interaction function
	results, err := h.processInteraction.Call(h.ctx, uint64(eventPtr))
	if err != nil {
		return nil, fmt.Errorf("failed to process interaction: %w", err)
	}

	if len(results) == 0 {
		return nil, fmt.Errorf("no result from process_interaction")
	}

	// Read result from memory
	resultPtr := uint32(results[0])
	resultJSON, err := h.readStringFromMemory(resultPtr)
	if err != nil {
		return nil, fmt.Errorf("failed to read result from memory: %w", err)
	}

	// Parse result
	var renderUpdate RenderUpdate
	if err := json.Unmarshal([]byte(resultJSON), &renderUpdate); err != nil {
		return nil, fmt.Errorf("failed to parse render update: %w", err)
	}

	return &renderUpdate, nil
}

// RenderFrame renders a frame at the given timestamp
func (h *LIVInteractiveHost) RenderFrame(timestamp float64) (*RenderUpdate, error) {
	// Call render_frame function
	results, err := h.renderFrame.Call(h.ctx, api.EncodeF64(timestamp))
	if err != nil {
		return nil, fmt.Errorf("failed to render frame: %w", err)
	}

	if len(results) == 0 {
		return nil, fmt.Errorf("no result from render_frame")
	}

	// Read result from memory
	resultPtr := uint32(results[0])
	resultJSON, err := h.readStringFromMemory(resultPtr)
	if err != nil {
		return nil, fmt.Errorf("failed to read result from memory: %w", err)
	}

	// Parse result
	var renderUpdate RenderUpdate
	if err := json.Unmarshal([]byte(resultJSON), &renderUpdate); err != nil {
		return nil, fmt.Errorf("failed to parse render update: %w", err)
	}

	return &renderUpdate, nil
}

// UpdateData updates a data source in the engine
func (h *LIVInteractiveHost) UpdateData(dataSourceID string, data []byte) error {
	// Write data source ID to memory
	idPtr, err := h.writeStringToMemory(dataSourceID)
	if err != nil {
		return fmt.Errorf("failed to write data source ID to memory: %w", err)
	}
	defer h.freeMemory(idPtr)

	// Write data to memory
	dataPtr, err := h.writeBytesToMemory(data)
	if err != nil {
		return fmt.Errorf("failed to write data to memory: %w", err)
	}
	defer h.freeMemory(dataPtr)

	// Call update_data function
	results, err := h.updateData.Call(h.ctx, uint64(idPtr), uint64(dataPtr), uint64(len(data)))
	if err != nil {
		return fmt.Errorf("failed to update data: %w", err)
	}

	// Check result (0 = success, non-zero = error)
	if len(results) > 0 && results[0] != 0 {
		return fmt.Errorf("data update failed with code: %d", results[0])
	}

	return nil
}

// GetPerformanceStats retrieves performance statistics from the engine
func (h *LIVInteractiveHost) GetPerformanceStats() (map[string]interface{}, error) {
	// Call get_performance_stats function
	results, err := h.getPerformanceStats.Call(h.ctx)
	if err != nil {
		return nil, fmt.Errorf("failed to get performance stats: %w", err)
	}

	if len(results) == 0 {
		return nil, fmt.Errorf("no result from get_performance_stats")
	}

	// Read result from memory
	resultPtr := uint32(results[0])
	resultJSON, err := h.readStringFromMemory(resultPtr)
	if err != nil {
		return nil, fmt.Errorf("failed to read result from memory: %w", err)
	}

	// Parse result
	var stats map[string]interface{}
	if err := json.Unmarshal([]byte(resultJSON), &stats); err != nil {
		return nil, fmt.Errorf("failed to parse performance stats: %w", err)
	}

	return stats, nil
}

// Destroy cleans up the engine and releases resources
func (h *LIVInteractiveHost) Destroy() error {
	// Call destroy_engine function
	if _, err := h.destroyEngine.Call(h.ctx); err != nil {
		log.Printf("Warning: failed to call destroy_engine: %v", err)
	}

	// Close the module and runtime
	if err := h.module.Close(h.ctx); err != nil {
		return fmt.Errorf("failed to close module: %w", err)
	}

	if err := h.runtime.Close(h.ctx); err != nil {
		return fmt.Errorf("failed to close runtime: %w", err)
	}

	log.Println("LIV Interactive Engine destroyed")
	return nil
}

// Memory management helpers

func (h *LIVInteractiveHost) writeStringToMemory(s string) (uint32, error) {
	data := []byte(s)
	return h.writeBytesToMemory(data)
}

func (h *LIVInteractiveHost) writeBytesToMemory(data []byte) (uint32, error) {
	// Allocate memory in WASM
	var ptr uint32
	if h.malloc != nil {
		results, err := h.malloc.Call(h.ctx, uint64(len(data)))
		if err != nil {
			return 0, fmt.Errorf("failed to allocate memory: %w", err)
		}
		ptr = uint32(results[0])
	} else {
		// Fallback: use a simple memory allocation strategy
		// In a real implementation, this would need proper memory management
		ptr = 1024 // Simple fixed offset for demo
	}

	// Write data to memory
	if !h.memory.Write(ptr, data) {
		return 0, fmt.Errorf("failed to write data to memory")
	}

	return ptr, nil
}

func (h *LIVInteractiveHost) readStringFromMemory(ptr uint32) (string, error) {
	// Read string length (assuming it's stored as a 32-bit integer before the string)
	lengthBytes, ok := h.memory.Read(ptr, 4)
	if !ok {
		return "", fmt.Errorf("failed to read string length")
	}

	length := uint32(lengthBytes[0]) | uint32(lengthBytes[1])<<8 | uint32(lengthBytes[2])<<16 | uint32(lengthBytes[3])<<24

	// Read string data
	stringBytes, ok := h.memory.Read(ptr+4, length)
	if !ok {
		return "", fmt.Errorf("failed to read string data")
	}

	return string(stringBytes), nil
}

func (h *LIVInteractiveHost) freeMemory(ptr uint32) {
	if h.free != nil {
		if _, err := h.free.Call(h.ctx, uint64(ptr)); err != nil {
			log.Printf("Warning: failed to free memory at %d: %v", ptr, err)
		}
	}
}

// Host functions for WASM

func consoleLog(ctx context.Context, m api.Module, offset, byteCount uint32) {
	buf, ok := m.Memory().Read(offset, byteCount)
	if !ok {
		log.Printf("Memory.Read(%d, %d) out of range", offset, byteCount)
		return
	}
	log.Printf("[WASM] %s", string(buf))
}

func performanceNow(ctx context.Context, m api.Module) float64 {
	return float64(time.Now().UnixNano()) / 1e6 // Convert to milliseconds
}

// Example usage
func main() {
	// Example WASM bytes (in reality, this would be loaded from a .wasm file)
	wasmBytes := []byte{} // Placeholder

	// Define security permissions
	permissions := WASMPermissions{
		MemoryLimit:         4 * 1024 * 1024, // 4MB
		AllowedImports:      []string{"env"},
		CPUTimeLimit:        5000, // 5 seconds
		AllowNetworking:     false,
		AllowFileSystem:     false,
		AllowedInteractions: []string{"Click", "Hover", "Touch"},
		MaxDataSize:         64 * 1024, // 64KB
		MaxElements:         1000,
	}

	// Create and initialize the host
	host, err := NewLIVInteractiveHost(wasmBytes, permissions)
	if err != nil {
		log.Fatalf("Failed to create host: %v", err)
	}
	defer host.Destroy()

	if err := host.Initialize(); err != nil {
		log.Fatalf("Failed to initialize engine: %v", err)
	}

	// Example interaction processing
	event := InteractionEvent{
		EventType:     "Click",
		TargetElement: stringPtr("button1"),
		Position:      &Position{X: 100, Y: 200},
		Data:          map[string]interface{}{"button": "left"},
		Timestamp:     float64(time.Now().UnixNano()) / 1e6,
	}

	renderUpdate, err := host.ProcessInteraction(event)
	if err != nil {
		log.Printf("Failed to process interaction: %v", err)
	} else {
		log.Printf("Render update: %+v", renderUpdate)
	}

	// Example frame rendering
	frameUpdate, err := host.RenderFrame(float64(time.Now().UnixNano()) / 1e6)
	if err != nil {
		log.Printf("Failed to render frame: %v", err)
	} else {
		log.Printf("Frame update: %+v", frameUpdate)
	}

	// Example data update
	chartData := map[string]interface{}{
		"values": []float64{10, 20, 30, 40, 50},
		"labels": []string{"A", "B", "C", "D", "E"},
	}

	chartDataJSON, _ := json.Marshal(chartData)
	if err := host.UpdateData("chart_data", chartDataJSON); err != nil {
		log.Printf("Failed to update data: %v", err)
	} else {
		log.Println("Data updated successfully")
	}

	// Get performance stats
	stats, err := host.GetPerformanceStats()
	if err != nil {
		log.Printf("Failed to get performance stats: %v", err)
	} else {
		log.Printf("Performance stats: %+v", stats)
	}
}

func stringPtr(s string) *string {
	return &s
}
