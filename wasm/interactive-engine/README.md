# LIV Interactive Engine (Rust WASM)

A memory-safe, secure interactive logic engine implemented in Rust and compiled to WebAssembly for the LIV document format.

## Overview

The LIV Interactive Engine provides secure execution of interactive content within .liv documents. It implements:

- **Memory-safe execution** using Rust's ownership system
- **Runtime permission checking** with granular security controls
- **Resource limit enforcement** to prevent abuse
- **Secure Go-WASM communication** interface
- **Animation and interaction processing** with 60fps performance targets

## Architecture

```
┌─────────────────────────────────────────────────────────────┐
│                    JavaScript Layer                         │
│  ┌─────────────────┐    ┌─────────────────────────────────┐ │
│  │   DOM Updates   │    │     Event Handling              │ │
│  └─────────────────┘    └─────────────────────────────────┘ │
└─────────────────────────────────────────────────────────────┘
                              │
                              ▼
┌─────────────────────────────────────────────────────────────┐
│                    WASM Interface                           │
│  ┌─────────────────┐    ┌─────────────────────────────────┐ │
│  │ process_interaction │ │    render_frame                │ │
│  │ update_data     │    │    get_performance_stats       │ │
│  └─────────────────┘    └─────────────────────────────────┘ │
└─────────────────────────────────────────────────────────────┘
                              │
                              ▼
┌─────────────────────────────────────────────────────────────┐
│                 Rust Interactive Engine                     │
│  ┌─────────────────┐    ┌─────────────────────────────────┐ │
│  │ Security Context│    │   Animation Controller          │ │
│  │ Event Processor │    │   Performance Monitor           │ │
│  │ Render Cache    │    │   Document State Manager        │ │
│  └─────────────────┘    └─────────────────────────────────┘ │
└─────────────────────────────────────────────────────────────┘
                              │
                              ▼
┌─────────────────────────────────────────────────────────────┐
│                    Go Host Environment                      │
│  ┌─────────────────┐    ┌─────────────────────────────────┐ │
│  │ WASM Runtime    │    │   Security Orchestration       │ │
│  │ Resource Monitor│    │   Permission Validation         │ │
│  └─────────────────┘    └─────────────────────────────────┘ │
└─────────────────────────────────────────────────────────────┘
```

## Security Model

### Permission System

The engine enforces strict security through a comprehensive permission system:

```rust
pub struct WASMPermissions {
    pub memory_limit: usize,           // Maximum memory allocation
    pub allowed_imports: Vec<String>,  // Allowed WASM imports
    pub cpu_time_limit: u32,          // Maximum CPU time (ms)
    pub allow_networking: bool,        // Network access permission
    pub allow_file_system: bool,      // File system access permission
    pub allowed_interactions: Vec<String>, // Permitted interaction types
    pub max_data_size: usize,         // Maximum data payload size
    pub max_elements: u32,            // Maximum DOM elements
}
```

### Resource Limits

- **Memory**: Configurable memory limits with allocation tracking
- **CPU Time**: Execution time limits to prevent infinite loops
- **Interaction Rate**: Rate limiting to prevent spam attacks
- **Element Count**: Limits on DOM element creation
- **Data Size**: Limits on data payload sizes

### Sandboxing

- All WASM code runs in isolated memory space
- No direct access to system resources
- All external communication through controlled interfaces
- Automatic cleanup of resources on destruction

## API Reference

### Initialization

```javascript
const engine = new LIVInteractiveEngine();
await engine.initialize({
    memory_limit: 4 * 1024 * 1024,  // 4MB
    cpu_time_limit: 5000,           // 5 seconds
    allowed_interactions: ["Click", "Hover", "Touch"],
    max_elements: 1000
});
```

### Processing Interactions

```javascript
// Process user interaction
const renderUpdate = await engine.processInteraction(domEvent);

// Apply updates to DOM
engine.applyRenderUpdate(renderUpdate);
```

### Animation Loop

```javascript
// Start animation loop
engine.startAnimationLoop();

// Stop animation loop
engine.stopAnimationLoop();
```

### Data Updates

```javascript
// Update data source
await engine.updateData('chart_data', {
    values: [10, 20, 30, 40, 50]
});
```

### Performance Monitoring

```javascript
// Get performance statistics
const stats = await engine.getPerformanceStats();
console.log(`FPS: ${stats.renders_per_second}`);
console.log(`Interactions/sec: ${stats.interactions_per_second}`);
```

## Core Components

### InteractiveEngine

The main engine class that orchestrates all interactive functionality:

- **Document State Management**: Maintains the current state of all interactive elements
- **Security Context**: Enforces permissions and resource limits
- **Animation Controller**: Manages CSS and programmatic animations
- **Event Processor**: Handles user interactions and system events
- **Render Cache**: Optimizes rendering performance through caching

### SecurityContext

Provides runtime security enforcement:

```rust
impl SecurityContext {
    pub fn check_interaction_permission(&mut self, event: &InteractionEvent) -> Result<(), WASMError>;
    pub fn check_render_permission(&self) -> Result<(), WASMError>;
    pub fn allocate_memory(&mut self, size: usize) -> Result<(), WASMError>;
    pub fn deallocate_memory(&mut self, size: usize);
}
```

### AnimationController

Manages smooth 60fps animations:

- **Keyframe Interpolation**: Smooth transitions between animation states
- **Easing Functions**: Support for various easing curves (linear, ease-in, ease-out, cubic-bezier)
- **Loop Control**: Configurable loop counts and directions
- **Performance Optimization**: Efficient animation updates with minimal DOM manipulation

### EventProcessor

Handles user interactions securely:

- **Event Validation**: Ensures events are permitted by security policy
- **Handler Execution**: Executes event handlers within security constraints
- **State Updates**: Updates document state based on interactions

## Data Structures

### InteractionEvent

```rust
pub struct InteractionEvent {
    pub event_type: InteractionType,
    pub target_element: Option<String>,
    pub position: Option<Position>,
    pub data: HashMap<String, serde_json::Value>,
    pub timestamp: f64,
}
```

### RenderUpdate

```rust
pub struct RenderUpdate {
    pub dom_operations: Vec<DOMOperation>,
    pub style_changes: Vec<StyleChange>,
    pub animation_updates: Vec<AnimationUpdate>,
    pub timestamp: f64,
}
```

### Animation

```rust
pub struct Animation {
    pub id: String,
    pub target_element: String,
    pub animation_type: AnimationType,
    pub duration: f64,
    pub easing: EasingFunction,
    pub keyframes: Vec<Keyframe>,
    pub loop_count: i32,
    pub direction: AnimationDirection,
}
```

## Building

### Prerequisites

- Rust 1.70+ with `wasm32-unknown-unknown` target
- `wasm-pack` for building WASM modules
- Node.js for JavaScript integration testing

### Build Commands

```bash
# Build for web target
wasm-pack build --target web --out-dir pkg

# Build for Node.js target
wasm-pack build --target nodejs --out-dir pkg-node

# Run tests
cargo test

# Build optimized release
wasm-pack build --target web --out-dir pkg --release
```

### Integration

```javascript
// Import the WASM module
import init, { 
    init_interactive_engine,
    process_interaction,
    render_frame,
    update_data,
    get_performance_stats,
    destroy_engine
} from './pkg/liv_interactive_engine.js';

// Initialize WASM
await init();

// Use the JavaScript wrapper for easier integration
const engine = new LIVInteractiveEngine();
await engine.initialize(permissions);
```

## Performance Characteristics

### Memory Usage

- **Base Memory**: ~100KB for engine initialization
- **Per Element**: ~1KB average per interactive element
- **Animation Data**: ~500B per active animation
- **Render Cache**: Configurable, default 100 cached updates

### CPU Performance

- **Interaction Processing**: <1ms for simple interactions
- **Frame Rendering**: <16ms target (60fps)
- **Animation Updates**: <5ms for 10 concurrent animations
- **Memory Allocation**: <0.1ms for typical allocations

### Security Overhead

- **Permission Checking**: <0.1ms per operation
- **Resource Monitoring**: <0.05ms per allocation
- **Event Validation**: <0.2ms per interaction

## Error Handling

The engine provides comprehensive error handling with specific error codes:

- `INTERACTION_NOT_ALLOWED`: Interaction type not permitted
- `INTERACTION_RATE_EXCEEDED`: Too many interactions per second
- `CPU_TIME_EXCEEDED`: CPU time limit exceeded
- `MEMORY_LIMIT_EXCEEDED`: Memory allocation would exceed limit
- `DATA_SIZE_EXCEEDED`: Data size exceeds security limits
- `INVALID_DATA`: Data parsing or validation failed

## Testing

The engine includes comprehensive tests covering:

- Security permission enforcement
- Animation interpolation accuracy
- Performance characteristics
- Memory management
- Error handling scenarios

Run tests with:

```bash
cargo test
```

## Integration with Go Host

The engine is designed to work with a Go host environment that provides:

- WASM module loading and lifecycle management
- Security policy enforcement at the host level
- Resource monitoring and limits
- Communication bridge between WASM and JavaScript layers

See the main LIV documentation for details on Go integration.

## Future Enhancements

- **WebGL Support**: Hardware-accelerated graphics rendering
- **Audio Processing**: Real-time audio manipulation
- **Advanced Physics**: Physics simulation for interactive elements
- **Multi-threading**: Worker thread support for heavy computations
- **Streaming Data**: Real-time data streaming capabilities