use wasm_bindgen::prelude::*;
use serde::{Deserialize, Serialize};
use std::collections::HashMap;
use std::sync::Mutex;
use std::time::{SystemTime, UNIX_EPOCH};

// When the `wee_alloc` feature is enabled, use `wee_alloc` as the global allocator
#[cfg(feature = "wee_alloc")]
#[global_allocator]
static ALLOC: wee_alloc::WeeAlloc = wee_alloc::WeeAlloc::INIT;

// Set up panic hook for better error messages
#[wasm_bindgen(start)]
pub fn main() {
    console_error_panic_hook::set_once();
}

// Global engine instance for memory-safe access
static ENGINE: Mutex<Option<InteractiveEngine>> = Mutex::new(None);

// Core data structures for interactive content

#[derive(Serialize, Deserialize, Clone, Debug)]
pub struct DocumentState {
    pub elements: Vec<InteractiveElement>,
    pub animations: Vec<Animation>,
    pub data_sources: HashMap<String, DataSource>,
    pub render_tree: RenderTree,
    pub viewport: Viewport,
}

#[derive(Serialize, Deserialize, Clone, Debug)]
pub struct InteractiveElement {
    pub id: String,
    pub element_type: ElementType,
    pub properties: HashMap<String, serde_json::Value>,
    pub children: Vec<String>,
    pub event_handlers: Vec<EventHandler>,
    pub transform: Transform,
    pub style: ElementStyle,
}

#[derive(Serialize, Deserialize, Clone, Debug)]
pub enum ElementType {
    Chart,
    Animation,
    Interactive,
    Vector,
    Text,
    Image,
    Container,
}

#[derive(Serialize, Deserialize, Clone, Debug)]
pub struct EventHandler {
    pub event_type: String,
    pub handler_id: String,
    pub parameters: HashMap<String, serde_json::Value>,
}

#[derive(Serialize, Deserialize, Clone, Debug)]
pub struct Transform {
    pub x: f64,
    pub y: f64,
    pub scale_x: f64,
    pub scale_y: f64,
    pub rotation: f64,
    pub opacity: f64,
}

#[derive(Serialize, Deserialize, Clone, Debug)]
pub struct ElementStyle {
    pub background_color: Option<String>,
    pub border_color: Option<String>,
    pub border_width: Option<f64>,
    pub border_radius: Option<f64>,
    pub shadow: Option<Shadow>,
}

#[derive(Serialize, Deserialize, Clone, Debug)]
pub struct Shadow {
    pub offset_x: f64,
    pub offset_y: f64,
    pub blur_radius: f64,
    pub color: String,
}

#[derive(Serialize, Deserialize, Clone, Debug)]
pub struct Animation {
    pub id: String,
    pub target_element: String,
    pub animation_type: AnimationType,
    pub duration: f64,
    pub easing: EasingFunction,
    pub keyframes: Vec<Keyframe>,
    pub loop_count: i32, // -1 for infinite
    pub direction: AnimationDirection,
}

#[derive(Serialize, Deserialize, Clone, Debug)]
pub enum AnimationType {
    Transform,
    Style,
    Path,
    Morph,
}

#[derive(Serialize, Deserialize, Clone, Debug)]
pub enum EasingFunction {
    Linear,
    EaseIn,
    EaseOut,
    EaseInOut,
    Cubic(f64, f64, f64, f64),
}

#[derive(Serialize, Deserialize, Clone, Debug)]
pub enum AnimationDirection {
    Normal,
    Reverse,
    Alternate,
    AlternateReverse,
}

#[derive(Serialize, Deserialize, Clone, Debug)]
pub struct Keyframe {
    pub time: f64, // 0.0 to 1.0
    pub properties: HashMap<String, serde_json::Value>,
}

#[derive(Serialize, Deserialize, Clone, Debug)]
pub struct DataSource {
    pub id: String,
    pub source_type: DataSourceType,
    pub data: serde_json::Value,
    pub update_frequency: Option<u32>, // milliseconds
    pub last_updated: f64,
}

#[derive(Serialize, Deserialize, Clone, Debug)]
pub enum DataSourceType {
    Static,
    Dynamic,
    Stream,
    Computed,
}

#[derive(Serialize, Deserialize, Clone, Debug)]
pub struct RenderTree {
    pub root: String,
    pub nodes: HashMap<String, RenderNode>,
    pub dirty_nodes: Vec<String>,
}

#[derive(Serialize, Deserialize, Clone, Debug)]
pub struct RenderNode {
    pub element_id: String,
    pub parent: Option<String>,
    pub children: Vec<String>,
    pub computed_style: ComputedStyle,
    pub bounds: BoundingBox,
    pub visible: bool,
}

#[derive(Serialize, Deserialize, Clone, Debug)]
pub struct ComputedStyle {
    pub position: Position,
    pub size: Size,
    pub color: String,
    pub background: String,
    pub transform: Transform,
}

#[derive(Serialize, Deserialize, Clone, Debug)]
pub struct Position {
    pub x: f64,
    pub y: f64,
}

#[derive(Serialize, Deserialize, Clone, Debug)]
pub struct Size {
    pub width: f64,
    pub height: f64,
}

#[derive(Serialize, Deserialize, Clone, Debug)]
pub struct BoundingBox {
    pub x: f64,
    pub y: f64,
    pub width: f64,
    pub height: f64,
}

#[derive(Serialize, Deserialize, Clone, Debug)]
pub struct Viewport {
    pub width: f64,
    pub height: f64,
    pub scale: f64,
    pub offset_x: f64,
    pub offset_y: f64,
}

// Render update structures for communication with JS layer

#[derive(Serialize, Deserialize, Clone, Debug)]
pub struct RenderUpdate {
    pub dom_operations: Vec<DOMOperation>,
    pub style_changes: Vec<StyleChange>,
    pub animation_updates: Vec<AnimationUpdate>,
    pub timestamp: f64,
}

#[derive(Serialize, Deserialize, Clone, Debug)]
pub enum DOMOperation {
    Create {
        element_id: String,
        tag: String,
        parent_id: Option<String>,
    },
    Update {
        element_id: String,
        attributes: HashMap<String, String>,
    },
    Remove {
        element_id: String,
    },
    Move {
        element_id: String,
        new_parent_id: String,
        index: usize,
    },
}

#[derive(Serialize, Deserialize, Clone, Debug)]
pub struct StyleChange {
    pub element_id: String,
    pub property: String,
    pub value: String,
}

#[derive(Serialize, Deserialize, Clone, Debug)]
pub struct AnimationUpdate {
    pub animation_id: String,
    pub progress: f64,
    pub current_values: HashMap<String, serde_json::Value>,
}

// Input event structures

#[derive(Serialize, Deserialize, Clone, Debug)]
pub struct InteractionEvent {
    pub event_type: InteractionType,
    pub target_element: Option<String>,
    pub position: Option<Position>,
    pub data: HashMap<String, serde_json::Value>,
    pub timestamp: f64,
    pub touch_data: Option<TouchData>,
    pub mouse_data: Option<MouseData>,
    pub keyboard_data: Option<KeyboardData>,
    pub gesture_data: Option<GestureData>,
    pub modifiers: EventModifiers,
}

#[derive(Serialize, Deserialize, Clone, Debug)]
pub struct TouchData {
    pub touches: Vec<TouchPoint>,
    pub changed_touches: Vec<TouchPoint>,
    pub target_touches: Vec<TouchPoint>,
    pub force: Option<f64>,
    pub rotation_angle: Option<f64>,
    pub scale: Option<f64>,
}

#[derive(Serialize, Deserialize, Clone, Debug)]
pub struct TouchPoint {
    pub identifier: u32,
    pub position: Position,
    pub radius: Option<f64>,
    pub rotation_angle: Option<f64>,
    pub force: Option<f64>,
}

#[derive(Serialize, Deserialize, Clone, Debug)]
pub struct MouseData {
    pub button: MouseButton,
    pub buttons: u16,
    pub position: Position,
    pub movement: Option<Position>,
    pub wheel_delta: Option<Position>,
}

#[derive(Serialize, Deserialize, Clone, Debug)]
pub enum MouseButton {
    None,
    Left,
    Middle,
    Right,
    Back,
    Forward,
}

#[derive(Serialize, Deserialize, Clone, Debug)]
pub struct KeyboardData {
    pub key: String,
    pub code: String,
    pub char_code: Option<u32>,
    pub key_code: Option<u32>,
    pub repeat: bool,
}

#[derive(Serialize, Deserialize, Clone, Debug)]
pub struct GestureData {
    pub gesture_type: GestureType,
    pub start_position: Position,
    pub current_position: Position,
    pub delta: Position,
    pub velocity: Option<Position>,
    pub scale: Option<f64>,
    pub rotation: Option<f64>,
    pub distance: Option<f64>,
    pub duration: f64,
}

#[derive(Serialize, Deserialize, Clone, Debug)]
pub enum GestureType {
    Tap,
    DoubleTap,
    LongPress,
    Pinch,
    Rotate,
    Swipe,
    Pan,
}

#[derive(Serialize, Deserialize, Clone, Debug)]
pub struct EventModifiers {
    pub ctrl: bool,
    pub shift: bool,
    pub alt: bool,
    pub meta: bool,
}

#[derive(Serialize, Deserialize, Clone, Debug)]
pub enum InteractionType {
    // Mouse events
    Click,
    DoubleClick,
    MouseDown,
    MouseUp,
    MouseMove,
    MouseEnter,
    MouseLeave,
    Hover,
    
    // Touch events
    TouchStart,
    TouchMove,
    TouchEnd,
    TouchCancel,
    
    // Gesture events
    Tap,
    DoubleTap,
    LongPress,
    Pinch,
    Rotate,
    Swipe,
    Pan,
    
    // Drag and drop
    DragStart,
    Drag,
    DragEnd,
    Drop,
    
    // Scroll and wheel
    Scroll,
    Wheel,
    
    // Keyboard events
    KeyDown,
    KeyUp,
    KeyPress,
    
    // System events
    Resize,
    Focus,
    Blur,
    
    // Custom events
    DataUpdate,
    StateChange,
}

// Error types

#[derive(Serialize, Deserialize, Clone, Debug)]
pub struct WASMError {
    pub code: String,
    pub message: String,
    pub details: Option<HashMap<String, serde_json::Value>>,
}

impl Default for Transform {
    fn default() -> Self {
        Self {
            x: 0.0,
            y: 0.0,
            scale_x: 1.0,
            scale_y: 1.0,
            rotation: 0.0,
            opacity: 1.0,
        }
    }
}

impl Default for Viewport {
    fn default() -> Self {
        Self {
            width: 1920.0,
            height: 1080.0,
            scale: 1.0,
            offset_x: 0.0,
            offset_y: 0.0,
        }
    }
}

// Core Interactive Engine Implementation
pub struct InteractiveEngine {
    document_state: DocumentState,
    security_context: SecurityContext,
    animation_controller: AnimationController,
    event_processor: EventProcessor,
    render_cache: RenderCache,
    performance_monitor: PerformanceMonitor,
    chart_renderer: ChartRenderer,
    vector_engine: VectorEngine,
    data_binding_manager: DataBindingManager,
    interaction_manager: InteractionManager,
    gesture_recognizer: GestureRecognizer,
    responsive_adapter: ResponsiveAdapter,
}

impl InteractiveEngine {
    pub fn new(permissions: WASMPermissions) -> Result<Self, WASMError> {
        let security_context = SecurityContext::new(permissions)?;
        
        Ok(Self {
            document_state: DocumentState::default(),
            security_context,
            animation_controller: AnimationController::new(),
            event_processor: EventProcessor::new(),
            render_cache: RenderCache::new(),
            performance_monitor: PerformanceMonitor::new(),
            chart_renderer: ChartRenderer::new(),
            vector_engine: VectorEngine::new(),
            data_binding_manager: DataBindingManager::new(),
            interaction_manager: InteractionManager::new(),
            gesture_recognizer: GestureRecognizer::new(),
            responsive_adapter: ResponsiveAdapter::new(),
        })
    }
    
    pub fn create_element(&mut self, element_type: ElementType, properties: HashMap<String, serde_json::Value>) -> Result<String, WASMError> {
        // Check permissions
        self.security_context.check_element_creation()?;
        
        // Generate unique ID
        let element_id = format!("element_{}", get_current_timestamp() as u64);
        
        // Create element
        let element = InteractiveElement {
            id: element_id.clone(),
            element_type,
            properties,
            children: Vec::new(),
            event_handlers: Vec::new(),
            transform: Transform::default(),
            style: ElementStyle {
                background_color: None,
                border_color: None,
                border_width: None,
                border_radius: None,
                shadow: None,
            },
        };
        
        // Add to document state
        self.document_state.add_element(element)?;
        
        Ok(element_id)
    }
    
    pub fn update_element_properties(&mut self, element_id: &str, properties: HashMap<String, serde_json::Value>) -> Result<(), WASMError> {
        self.security_context.check_element_modification(element_id)?;
        self.document_state.update_element(element_id, properties)
    }
    
    pub fn delete_element(&mut self, element_id: &str) -> Result<(), WASMError> {
        self.security_context.check_element_modification(element_id)?;
        self.document_state.remove_element(element_id)
    }
    
    pub fn create_animation(&mut self, target_element: &str, animation_type: AnimationType, duration: f64, keyframes: Vec<Keyframe>) -> Result<String, WASMError> {
        // Check permissions
        self.security_context.check_animation_creation()?;
        
        // Verify target element exists
        if self.document_state.get_element(target_element).is_none() {
            return Err(WASMError::new("TARGET_NOT_FOUND", "Target element not found"));
        }
        
        // Generate unique animation ID
        let animation_id = format!("anim_{}", get_current_timestamp() as u64);
        
        // Create animation
        let animation = Animation {
            id: animation_id.clone(),
            target_element: target_element.to_string(),
            animation_type,
            duration,
            easing: EasingFunction::EaseInOut,
            keyframes,
            loop_count: 1,
            direction: AnimationDirection::Normal,
        };
        
        // Add to document state
        self.document_state.animations.push(animation.clone());
        
        // Start animation
        self.animation_controller.start_animation(animation);
        
        Ok(animation_id)
    }
    
    pub fn stop_animation(&mut self, animation_id: &str) -> Result<(), WASMError> {
        self.animation_controller.stop_animation(animation_id);
        
        // Remove from document state
        self.document_state.animations.retain(|anim| anim.id != animation_id);
        
        Ok(())
    }
    
    pub fn add_event_handler(&mut self, element_id: &str, event_type: &str, handler_id: &str) -> Result<(), WASMError> {
        self.security_context.check_event_handler_creation()?;
        
        let element = self.document_state.get_element_mut(element_id)
            .ok_or_else(|| WASMError::new("ELEMENT_NOT_FOUND", "Element not found"))?;
        
        let event_handler = EventHandler {
            event_type: event_type.to_string(),
            handler_id: handler_id.to_string(),
            parameters: HashMap::new(),
        };
        
        element.event_handlers.push(event_handler);
        
        Ok(())
    }
    
    pub fn update_viewport(&mut self, width: f64, height: f64, scale: f64) -> Result<(), WASMError> {
        self.document_state.viewport.width = width;
        self.document_state.viewport.height = height;
        self.document_state.viewport.scale = scale;
        
        // Mark all elements as dirty for responsive recalculation
        let element_ids: Vec<String> = self.document_state.elements.iter().map(|e| e.id.clone()).collect();
        for element_id in element_ids {
            if !self.document_state.render_tree.dirty_nodes.contains(&element_id) {
                self.document_state.render_tree.dirty_nodes.push(element_id);
            }
        }
        
        Ok(())
    }
    
    pub fn get_element_bounds(&self, element_id: &str) -> Result<BoundingBox, WASMError> {
        let render_node = self.document_state.render_tree.nodes.get(element_id)
            .ok_or_else(|| WASMError::new("ELEMENT_NOT_FOUND", "Element not found in render tree"))?;
        
        Ok(render_node.bounds.clone())
    }
    
    pub fn query_elements_by_type(&self, element_type: ElementType) -> Vec<String> {
        self.document_state.elements.iter()
            .filter(|e| std::mem::discriminant(&e.element_type) == std::mem::discriminant(&element_type))
            .map(|e| e.id.clone())
            .collect()
    }

    pub fn process_interaction(&mut self, mut event: InteractionEvent) -> Result<RenderUpdate, WASMError> {
        // Check permissions for the interaction
        self.security_context.check_interaction_permission(&event)?;
        
        // Adapt event for responsive interaction
        self.responsive_adapter.adapt_event(&mut event)?;
        
        // Process touch input through gesture recognizer
        let mut gesture_events = Vec::new();
        if let Some(touch_data) = &event.touch_data {
            gesture_events = self.gesture_recognizer.process_touch_input(touch_data, event.timestamp);
        }
        
        // Process the event through interaction manager
        let interaction_responses = self.interaction_manager.process_event(&event)?;
        
        // Process the event through legacy event processor
        let legacy_changes = self.event_processor.process_event(&mut self.document_state, event)?;
        
        // Convert interaction responses to element changes
        let mut all_changes = legacy_changes;
        for response in interaction_responses {
            all_changes.extend(self.convert_interaction_response_to_changes(response)?);
        }
        
        // Process gesture events
        for gesture_event in gesture_events {
            all_changes.extend(self.process_gesture_event(gesture_event)?);
        }
        
        // Update performance metrics
        self.performance_monitor.record_interaction();
        
        // Generate render update
        let render_update = self.generate_render_update(all_changes)?;
        
        // Cache the update for optimization
        self.render_cache.cache_update(&render_update);
        
        // Clean up completed gesture recognitions
        self.gesture_recognizer.clear_completed_recognitions();
        
        Ok(render_update)
    }

    fn convert_interaction_response_to_changes(&self, response: InteractionResponse) -> Result<Vec<ElementChange>, WASMError> {
        let mut changes = Vec::new();
        
        match response.response_type {
            ResponseType::StateChanged => {
                if let Some(element_id) = response.target_element {
                    changes.push(ElementChange::Update {
                        element_id,
                        properties: response.data,
                    });
                }
            }
            ResponseType::Click | ResponseType::DoubleClick | ResponseType::Tap => {
                if let Some(element_id) = response.target_element {
                    // Trigger visual feedback for click/tap
                    changes.push(ElementChange::Update {
                        element_id: element_id.clone(),
                        properties: [
                            ("interaction_feedback".to_string(), serde_json::json!("active")),
                            ("last_interaction".to_string(), serde_json::json!(response.timestamp)),
                        ].into_iter().collect(),
                    });
                }
            }
            ResponseType::DragStart => {
                if let Some(element_id) = response.target_element {
                    changes.push(ElementChange::Update {
                        element_id,
                        properties: [
                            ("dragging".to_string(), serde_json::json!(true)),
                            ("drag_data".to_string(), serde_json::json!(response.data)),
                        ].into_iter().collect(),
                    });
                }
            }
            ResponseType::Drag => {
                if let Some(element_id) = response.target_element {
                    changes.push(ElementChange::Update {
                        element_id,
                        properties: response.data,
                    });
                }
            }
            ResponseType::DragEnd => {
                if let Some(element_id) = response.target_element {
                    changes.push(ElementChange::Update {
                        element_id,
                        properties: [
                            ("dragging".to_string(), serde_json::json!(false)),
                        ].into_iter().collect(),
                    });
                }
            }
            ResponseType::Gesture => {
                if let Some(element_id) = response.target_element {
                    changes.push(ElementChange::Update {
                        element_id,
                        properties: [
                            ("gesture_data".to_string(), serde_json::json!(response.data)),
                            ("gesture_timestamp".to_string(), serde_json::json!(response.timestamp)),
                        ].into_iter().collect(),
                    });
                }
            }
            ResponseType::Resize => {
                // Update viewport and trigger responsive recalculation
                if let Some(width) = response.data.get("width").and_then(|v| v.as_f64()) {
                    if let Some(height) = response.data.get("height").and_then(|v| v.as_f64()) {
                        self.responsive_adapter.initialize_device_detection(&Viewport {
                            width,
                            height,
                            scale: 1.0,
                            offset_x: 0.0,
                            offset_y: 0.0,
                        }).ok();
                    }
                }
            }
            _ => {
                // Handle other response types as needed
            }
        }
        
        Ok(changes)
    }

    fn process_gesture_event(&self, gesture_event: GestureEvent) -> Result<Vec<ElementChange>, WASMError> {
        let mut changes = Vec::new();
        
        // Create a synthetic interaction event for the gesture
        let gesture_interaction = InteractionEvent {
            event_type: match gesture_event.gesture_type {
                GestureType::Tap => InteractionType::Tap,
                GestureType::DoubleTap => InteractionType::DoubleTap,
                GestureType::LongPress => InteractionType::LongPress,
                GestureType::Pinch => InteractionType::Pinch,
                GestureType::Rotate => InteractionType::Rotate,
                GestureType::Swipe => InteractionType::Swipe,
                GestureType::Pan => InteractionType::Pan,
            },
            target_element: None, // Would need to determine from position
            position: Some(gesture_event.end_position),
            data: [
                ("gesture_confidence".to_string(), serde_json::json!(gesture_event.confidence)),
                ("gesture_duration".to_string(), serde_json::json!(gesture_event.duration)),
                ("gesture_velocity".to_string(), serde_json::json!(gesture_event.velocity)),
            ].into_iter().chain(
                gesture_event.properties.into_iter().map(|(k, v)| (k, serde_json::json!(v)))
            ).collect(),
            timestamp: gesture_event.timestamp,
            touch_data: None,
            mouse_data: None,
            keyboard_data: None,
            gesture_data: Some(GestureData {
                gesture_type: gesture_event.gesture_type,
                start_position: gesture_event.start_position,
                current_position: gesture_event.end_position,
                delta: Position {
                    x: gesture_event.end_position.x - gesture_event.start_position.x,
                    y: gesture_event.end_position.y - gesture_event.start_position.y,
                },
                velocity: Some(gesture_event.velocity),
                scale: None,
                rotation: None,
                distance: Some(((gesture_event.end_position.x - gesture_event.start_position.x).powi(2) + 
                              (gesture_event.end_position.y - gesture_event.start_position.y).powi(2)).sqrt()),
                duration: gesture_event.duration,
            }),
            modifiers: EventModifiers {
                ctrl: false,
                shift: false,
                alt: false,
                meta: false,
            },
        };
        
        // Process the gesture as a regular interaction
        // This would trigger any registered gesture handlers
        changes.push(ElementChange::Update {
            element_id: "gesture_system".to_string(),
            properties: [
                ("last_gesture".to_string(), serde_json::json!(gesture_interaction)),
            ].into_iter().collect(),
        });
        
        Ok(changes)
    }

    pub fn add_interaction_delegate(&mut self, target: &str, delegate: EventDelegate) -> Result<(), WASMError> {
        self.security_context.check_element_modification(target)?;
        self.interaction_manager.add_event_delegate(target, delegate);
        Ok(())
    }

    pub fn remove_interaction_delegate(&mut self, target: &str, handler_id: &str) -> Result<(), WASMError> {
        self.security_context.check_element_modification(target)?;
        self.interaction_manager.remove_event_delegate(target, handler_id);
        Ok(())
    }

    pub fn get_interaction_state(&self, element_id: &str) -> Option<&InteractionState> {
        self.interaction_manager.get_interaction_state(element_id)
    }

    pub fn get_interaction_metrics(&self) -> &InteractionMetrics {
        self.interaction_manager.get_performance_metrics()
    }

    pub fn update_device_capabilities(&mut self, device_info: DeviceInfo) -> Result<(), WASMError> {
        self.responsive_adapter.update_device_info(device_info);
        
        // Reinitialize with current viewport
        self.responsive_adapter.initialize_device_detection(&self.document_state.viewport)?;
        
        Ok(())
    }

    pub fn render_frame(&mut self, timestamp: f64) -> Result<RenderUpdate, WASMError> {
        // Check if we have permission to render
        self.security_context.check_render_permission()?;
        
        // Update animations
        let mut all_changes = self.animation_controller.update_animations(
            &mut self.document_state, 
            timestamp
        )?;
        
        // Update data bindings
        let binding_changes = self.data_binding_manager.update_bindings(
            &mut self.document_state,
            timestamp
        );
        all_changes.extend(binding_changes);
        
        // Generate render update if there are changes
        if !all_changes.is_empty() {
            let render_update = self.generate_render_update(all_changes)?;
            self.render_cache.cache_update(&render_update);
            Ok(render_update)
        } else {
            // Return empty update if no changes
            Ok(RenderUpdate::empty())
        }
    }

    pub fn update_data(&mut self, data_source_id: &str, data: &[u8]) -> Result<(), WASMError> {
        // Check permission to update data
        self.security_context.check_data_permission(data_source_id)?;
        
        // Validate data size
        if data.len() > self.security_context.max_data_size() {
            return Err(WASMError::new("DATA_SIZE_EXCEEDED", "Data size exceeds security limits"));
        }
        
        // Parse and validate data
        let parsed_data: serde_json::Value = serde_json::from_slice(data)
            .map_err(|e| WASMError::new("INVALID_DATA", &format!("Failed to parse data: {}", e)))?;
        
        // Update data source
        if let Some(data_source) = self.document_state.data_sources.get_mut(data_source_id) {
            data_source.data = parsed_data;
            data_source.last_updated = get_current_timestamp();
        }
        
        Ok(())
    }

    fn generate_render_update(&self, changes: Vec<ElementChange>) -> Result<RenderUpdate, WASMError> {
        let mut dom_operations = Vec::new();
        let mut style_changes = Vec::new();
        let mut animation_updates = Vec::new();
        
        for change in changes {
            match change {
                ElementChange::Create { element_id, element_type, parent_id } => {
                    dom_operations.push(DOMOperation::Create {
                        element_id,
                        tag: element_type.to_tag(),
                        parent_id,
                    });
                }
                ElementChange::Update { element_id, properties } => {
                    for (property, value) in properties {
                        if property.starts_with("style.") {
                            style_changes.push(StyleChange {
                                element_id: element_id.clone(),
                                property: property.strip_prefix("style.").unwrap().to_string(),
                                value: value.to_string(),
                            });
                        } else {
                            dom_operations.push(DOMOperation::Update {
                                element_id: element_id.clone(),
                                attributes: [(property, value.to_string())].into_iter().collect(),
                            });
                        }
                    }
                }
                ElementChange::Remove { element_id } => {
                    dom_operations.push(DOMOperation::Remove { element_id });
                }
                ElementChange::AnimationUpdate { animation_id, progress, values } => {
                    animation_updates.push(AnimationUpdate {
                        animation_id,
                        progress,
                        current_values: values,
                    });
                }
            }
        }
        
        Ok(RenderUpdate {
            dom_operations,
            style_changes,
            animation_updates,
            timestamp: get_current_timestamp(),
        })
    }
}

impl Default for DocumentState {
    fn default() -> Self {
        Self {
            elements: Vec::new(),
            animations: Vec::new(),
            data_sources: HashMap::new(),
            render_tree: RenderTree::default(),
            viewport: Viewport::default(),
        }
    }
}

impl DocumentState {
    pub fn add_element(&mut self, element: InteractiveElement) -> Result<(), WASMError> {
        // Check if element already exists
        if self.elements.iter().any(|e| e.id == element.id) {
            return Err(WASMError::new("ELEMENT_EXISTS", "Element with this ID already exists"));
        }
        
        // Add to elements list
        self.elements.push(element.clone());
        
        // Add to render tree
        let render_node = RenderNode {
            element_id: element.id.clone(),
            parent: None, // Will be set when added to parent
            children: Vec::new(),
            computed_style: ComputedStyle::from_element(&element),
            bounds: BoundingBox { x: 0.0, y: 0.0, width: 0.0, height: 0.0 },
            visible: true,
        };
        
        self.render_tree.nodes.insert(element.id.clone(), render_node);
        self.render_tree.dirty_nodes.push(element.id);
        
        Ok(())
    }
    
    pub fn remove_element(&mut self, element_id: &str) -> Result<(), WASMError> {
        // Remove from elements list
        let element_index = self.elements.iter().position(|e| e.id == element_id)
            .ok_or_else(|| WASMError::new("ELEMENT_NOT_FOUND", "Element not found"))?;
        
        let element = self.elements.remove(element_index);
        
        // Remove from render tree
        self.render_tree.nodes.remove(&element.id);
        
        // Remove from dirty nodes if present
        self.render_tree.dirty_nodes.retain(|id| id != &element.id);
        
        // Remove any animations targeting this element
        self.animations.retain(|anim| anim.target_element != element.id);
        
        Ok(())
    }
    
    pub fn update_element(&mut self, element_id: &str, properties: HashMap<String, serde_json::Value>) -> Result<(), WASMError> {
        let element = self.elements.iter_mut()
            .find(|e| e.id == element_id)
            .ok_or_else(|| WASMError::new("ELEMENT_NOT_FOUND", "Element not found"))?;
        
        // Update element properties
        for (key, value) in properties {
            element.properties.insert(key, value);
        }
        
        // Mark as dirty for re-rendering
        if !self.render_tree.dirty_nodes.contains(&element.id) {
            self.render_tree.dirty_nodes.push(element.id);
        }
        
        // Update computed style in render tree
        if let Some(render_node) = self.render_tree.nodes.get_mut(&element.id) {
            render_node.computed_style = ComputedStyle::from_element(element);
        }
        
        Ok(())
    }
    
    pub fn get_element(&self, element_id: &str) -> Option<&InteractiveElement> {
        self.elements.iter().find(|e| e.id == element_id)
    }
    
    pub fn get_element_mut(&mut self, element_id: &str) -> Option<&mut InteractiveElement> {
        self.elements.iter_mut().find(|e| e.id == element_id)
    }
}

impl Default for RenderTree {
    fn default() -> Self {
        Self {
            root: "root".to_string(),
            nodes: HashMap::new(),
            dirty_nodes: Vec::new(),
        }
    }
}

impl ComputedStyle {
    pub fn from_element(element: &InteractiveElement) -> Self {
        // Extract position from transform
        let position = Position {
            x: element.transform.x,
            y: element.transform.y,
        };
        
        // Extract size from properties or use defaults
        let width = element.properties.get("width")
            .and_then(|v| v.as_f64())
            .unwrap_or(100.0);
        let height = element.properties.get("height")
            .and_then(|v| v.as_f64())
            .unwrap_or(100.0);
        
        let size = Size { width, height };
        
        // Extract colors from style
        let color = element.style.background_color.clone().unwrap_or_else(|| "#000000".to_string());
        let background = element.style.background_color.clone().unwrap_or_else(|| "transparent".to_string());
        
        Self {
            position,
            size,
            color,
            background,
            transform: element.transform.clone(),
        }
    }
}

// Security Context for permission checking and resource limits
#[derive(Clone, Debug)]
pub struct SecurityContext {
    permissions: WASMPermissions,
    resource_limits: ResourceLimits,
    allocated_memory: usize,
    interaction_count: u32,
    start_time: f64,
}

#[derive(Clone, Debug, Serialize, Deserialize)]
pub struct WASMPermissions {
    pub memory_limit: usize,
    pub allowed_imports: Vec<String>,
    pub cpu_time_limit: u32, // milliseconds
    pub allow_networking: bool,
    pub allow_file_system: bool,
    pub allowed_interactions: Vec<String>,
    pub max_data_size: usize,
    pub max_elements: u32,
}

#[derive(Clone, Debug)]
pub struct ResourceLimits {
    pub max_memory: usize,
    pub max_cpu_time: u32,
    pub max_interactions_per_second: u32,
    pub max_elements: u32,
}

impl SecurityContext {
    pub fn new(permissions: WASMPermissions) -> Result<Self, WASMError> {
        let resource_limits = ResourceLimits {
            max_memory: permissions.memory_limit,
            max_cpu_time: permissions.cpu_time_limit,
            max_interactions_per_second: 100, // Default limit
            max_elements: permissions.max_elements,
        };
        
        Ok(Self {
            permissions,
            resource_limits,
            allocated_memory: 0,
            interaction_count: 0,
            start_time: get_current_timestamp(),
        })
    }

    pub fn check_interaction_permission(&mut self, event: &InteractionEvent) -> Result<(), WASMError> {
        // Check if interaction type is allowed
        let interaction_type = format!("{:?}", event.event_type);
        if !self.permissions.allowed_interactions.contains(&interaction_type) {
            return Err(WASMError::new(
                "INTERACTION_NOT_ALLOWED",
                &format!("Interaction type '{}' is not permitted", interaction_type)
            ));
        }
        
        // Check interaction rate limiting
        self.interaction_count += 1;
        let elapsed = get_current_timestamp() - self.start_time;
        if elapsed > 0.0 {
            let rate = (self.interaction_count as f64) / (elapsed / 1000.0);
            if rate > self.resource_limits.max_interactions_per_second as f64 {
                return Err(WASMError::new(
                    "INTERACTION_RATE_EXCEEDED",
                    "Too many interactions per second"
                ));
            }
        }
        
        Ok(())
    }

    pub fn check_render_permission(&self) -> Result<(), WASMError> {
        // Check CPU time limit
        let elapsed = get_current_timestamp() - self.start_time;
        if elapsed > self.resource_limits.max_cpu_time as f64 {
            return Err(WASMError::new(
                "CPU_TIME_EXCEEDED",
                "CPU time limit exceeded"
            ));
        }
        
        Ok(())
    }

    pub fn check_data_permission(&self, _data_source_id: &str) -> Result<(), WASMError> {
        // For now, allow all data updates if we have general permissions
        // In the future, this could be more granular
        Ok(())
    }

    pub fn max_data_size(&self) -> usize {
        self.permissions.max_data_size
    }

    pub fn allocate_memory(&mut self, size: usize) -> Result<(), WASMError> {
        if self.allocated_memory + size > self.resource_limits.max_memory {
            return Err(WASMError::new(
                "MEMORY_LIMIT_EXCEEDED",
                "Memory allocation would exceed limit"
            ));
        }
        
        self.allocated_memory += size;
        Ok(())
    }

    pub fn deallocate_memory(&mut self, size: usize) {
        self.allocated_memory = self.allocated_memory.saturating_sub(size);
    }
    
    pub fn check_element_creation(&self) -> Result<(), WASMError> {
        // Check if we can create more elements
        if !self.permissions.allowed_interactions.contains(&"create_element".to_string()) {
            return Err(WASMError::new("ELEMENT_CREATION_NOT_ALLOWED", "Element creation is not permitted"));
        }
        Ok(())
    }
    
    pub fn check_element_modification(&self, _element_id: &str) -> Result<(), WASMError> {
        if !self.permissions.allowed_interactions.contains(&"modify_element".to_string()) {
            return Err(WASMError::new("ELEMENT_MODIFICATION_NOT_ALLOWED", "Element modification is not permitted"));
        }
        Ok(())
    }
    
    pub fn check_animation_creation(&self) -> Result<(), WASMError> {
        if !self.permissions.allowed_interactions.contains(&"create_animation".to_string()) {
            return Err(WASMError::new("ANIMATION_CREATION_NOT_ALLOWED", "Animation creation is not permitted"));
        }
        Ok(())
    }
    
    pub fn check_event_handler_creation(&self) -> Result<(), WASMError> {
        if !self.permissions.allowed_interactions.contains(&"create_event_handler".to_string()) {
            return Err(WASMError::new("EVENT_HANDLER_CREATION_NOT_ALLOWED", "Event handler creation is not permitted"));
        }
        Ok(())
    }
}

// Animation Controller for managing animations
pub struct AnimationController {
    active_animations: HashMap<String, ActiveAnimation>,
}

#[derive(Clone, Debug)]
pub struct ActiveAnimation {
    animation: Animation,
    start_time: f64,
    current_iteration: i32,
}

impl AnimationController {
    pub fn new() -> Self {
        Self {
            active_animations: HashMap::new(),
        }
    }

    pub fn start_animation(&mut self, animation: Animation) {
        let active_animation = ActiveAnimation {
            animation: animation.clone(),
            start_time: get_current_timestamp(),
            current_iteration: 0,
        };
        
        self.active_animations.insert(animation.id.clone(), active_animation);
    }

    pub fn stop_animation(&mut self, animation_id: &str) {
        self.active_animations.remove(animation_id);
    }

    pub fn update_animations(
        &mut self, 
        document_state: &mut DocumentState, 
        timestamp: f64
    ) -> Result<Vec<ElementChange>, WASMError> {
        let mut changes = Vec::new();
        let mut completed_animations = Vec::new();

        for (animation_id, active_animation) in &mut self.active_animations {
            let elapsed = timestamp - active_animation.start_time;
            let progress = (elapsed / active_animation.animation.duration).min(1.0);
            
            // Calculate current values based on progress and easing
            let eased_progress = apply_easing(progress, &active_animation.animation.easing);
            let current_values = interpolate_keyframes(&active_animation.animation.keyframes, eased_progress);
            
            // Create animation update
            changes.push(ElementChange::AnimationUpdate {
                animation_id: animation_id.clone(),
                progress: eased_progress,
                values: current_values,
            });
            
            // Check if animation is complete
            if progress >= 1.0 {
                active_animation.current_iteration += 1;
                
                if active_animation.animation.loop_count == -1 || 
                   active_animation.current_iteration < active_animation.animation.loop_count {
                    // Restart animation
                    active_animation.start_time = timestamp;
                } else {
                    // Animation completed
                    completed_animations.push(animation_id.clone());
                }
            }
        }

        // Remove completed animations
        for animation_id in completed_animations {
            self.active_animations.remove(&animation_id);
        }

        Ok(changes)
    }
}

// Interaction Manager for state management and event delegation
pub struct InteractionManager {
    interaction_states: HashMap<String, InteractionState>,
    active_gestures: HashMap<u32, ActiveGesture>,
    event_delegates: HashMap<String, Vec<EventDelegate>>,
    touch_tracking: HashMap<u32, TouchTracker>,
    mouse_state: MouseState,
    keyboard_state: KeyboardState,
    performance_metrics: InteractionMetrics,
}

#[derive(Serialize, Deserialize, Clone, Debug)]
pub struct InteractionState {
    pub element_id: String,
    pub state_type: InteractionStateType,
    pub start_time: f64,
    pub last_update: f64,
    pub properties: HashMap<String, serde_json::Value>,
}

#[derive(Serialize, Deserialize, Clone, Debug)]
pub enum InteractionStateType {
    Idle,
    Hover,
    Active,
    Pressed,
    Dragging,
    Focused,
    Disabled,
}

#[derive(Clone, Debug)]
pub struct ActiveGesture {
    pub gesture_type: GestureType,
    pub start_time: f64,
    pub start_position: Position,
    pub current_position: Position,
    pub touch_points: Vec<TouchPoint>,
    pub properties: HashMap<String, f64>,
}

#[derive(Clone, Debug)]
pub struct EventDelegate {
    pub element_id: String,
    pub event_types: Vec<InteractionType>,
    pub handler_id: String,
    pub capture: bool,
    pub priority: i32,
}

#[derive(Clone, Debug)]
pub struct TouchTracker {
    pub touch_id: u32,
    pub start_position: Position,
    pub current_position: Position,
    pub start_time: f64,
    pub last_update: f64,
    pub velocity: Position,
    pub target_element: Option<String>,
}

#[derive(Clone, Debug)]
pub struct MouseState {
    pub position: Position,
    pub buttons: u16,
    pub last_click_time: f64,
    pub click_count: u32,
    pub target_element: Option<String>,
    pub dragging: bool,
    pub drag_start_position: Option<Position>,
}

#[derive(Clone, Debug)]
pub struct KeyboardState {
    pub pressed_keys: HashMap<String, f64>,
    pub modifiers: EventModifiers,
    pub focused_element: Option<String>,
    pub composition_active: bool,
}

#[derive(Serialize, Deserialize, Clone, Debug, Default)]
pub struct InteractionMetrics {
    pub total_events: u32,
    pub events_per_second: f64,
    pub average_response_time: f64,
    pub gesture_recognition_time: f64,
    pub touch_points_processed: u32,
    pub mouse_events_processed: u32,
    pub keyboard_events_processed: u32,
}

impl InteractionManager {
    pub fn new() -> Self {
        Self {
            interaction_states: HashMap::new(),
            active_gestures: HashMap::new(),
            event_delegates: HashMap::new(),
            touch_tracking: HashMap::new(),
            mouse_state: MouseState {
                position: Position { x: 0.0, y: 0.0 },
                buttons: 0,
                last_click_time: 0.0,
                click_count: 0,
                target_element: None,
                dragging: false,
                drag_start_position: None,
            },
            keyboard_state: KeyboardState {
                pressed_keys: HashMap::new(),
                modifiers: EventModifiers {
                    ctrl: false,
                    shift: false,
                    alt: false,
                    meta: false,
                },
                focused_element: None,
                composition_active: false,
            },
            performance_metrics: InteractionMetrics::default(),
        }
    }

    pub fn process_event(&mut self, event: &InteractionEvent) -> Result<Vec<InteractionResponse>, WASMError> {
        let start_time = get_current_timestamp();
        let mut responses = Vec::new();

        // Update performance metrics
        self.performance_metrics.total_events += 1;

        // Process event based on type
        match event.event_type {
            // Mouse events
            InteractionType::MouseDown | InteractionType::MouseUp | InteractionType::MouseMove => {
                responses.extend(self.handle_mouse_event(event)?);
            }
            InteractionType::Click | InteractionType::DoubleClick => {
                responses.extend(self.handle_click_event(event)?);
            }
            
            // Touch events
            InteractionType::TouchStart | InteractionType::TouchMove | 
            InteractionType::TouchEnd | InteractionType::TouchCancel => {
                responses.extend(self.handle_touch_event(event)?);
            }
            
            // Keyboard events
            InteractionType::KeyDown | InteractionType::KeyUp | InteractionType::KeyPress => {
                responses.extend(self.handle_keyboard_event(event)?);
            }
            
            // Gesture events
            InteractionType::Tap | InteractionType::DoubleTap | InteractionType::LongPress |
            InteractionType::Pinch | InteractionType::Rotate | InteractionType::Swipe | InteractionType::Pan => {
                responses.extend(self.handle_gesture_event(event)?);
            }
            
            // Other events
            InteractionType::Scroll | InteractionType::Wheel => {
                responses.extend(self.handle_scroll_event(event)?);
            }
            InteractionType::Focus | InteractionType::Blur => {
                responses.extend(self.handle_focus_event(event)?);
            }
            InteractionType::Resize => {
                responses.extend(self.handle_resize_event(event)?);
            }
            
            _ => {
                // Handle other event types
                responses.push(InteractionResponse::new(
                    event.target_element.clone(),
                    ResponseType::EventProcessed,
                    HashMap::new(),
                ));
            }
        }

        // Update performance metrics
        let processing_time = get_current_timestamp() - start_time;
        self.update_performance_metrics(processing_time);

        // Process event delegation
        if let Some(target) = &event.target_element {
            responses.extend(self.delegate_event(target, event)?);
        }

        Ok(responses)
    }

    fn handle_mouse_event(&mut self, event: &InteractionEvent) -> Result<Vec<InteractionResponse>, WASMError> {
        let mut responses = Vec::new();
        
        if let Some(mouse_data) = &event.mouse_data {
            // Update mouse state
            self.mouse_state.position = mouse_data.position.clone();
            self.mouse_state.buttons = mouse_data.buttons;
            
            match event.event_type {
                InteractionType::MouseDown => {
                    self.mouse_state.target_element = event.target_element.clone();
                    if let Some(target) = &event.target_element {
                        self.set_interaction_state(target, InteractionStateType::Pressed, event.timestamp);
                        responses.push(InteractionResponse::new(
                            Some(target.clone()),
                            ResponseType::StateChanged,
                            [("state".to_string(), serde_json::json!("pressed"))].into_iter().collect(),
                        ));
                    }
                }
                InteractionType::MouseUp => {
                    if let Some(target) = &self.mouse_state.target_element {
                        self.set_interaction_state(target, InteractionStateType::Hover, event.timestamp);
                        responses.push(InteractionResponse::new(
                            Some(target.clone()),
                            ResponseType::StateChanged,
                            [("state".to_string(), serde_json::json!("hover"))].into_iter().collect(),
                        ));
                    }
                    self.mouse_state.target_element = None;
                }
                InteractionType::MouseMove => {
                    // Check for drag operations
                    if self.mouse_state.buttons > 0 && !self.mouse_state.dragging {
                        if let Some(start_pos) = &self.mouse_state.drag_start_position {
                            let distance = ((mouse_data.position.x - start_pos.x).powi(2) + 
                                          (mouse_data.position.y - start_pos.y).powi(2)).sqrt();
                            if distance > 5.0 { // Drag threshold
                                self.mouse_state.dragging = true;
                                responses.push(InteractionResponse::new(
                                    event.target_element.clone(),
                                    ResponseType::DragStart,
                                    [("start_position".to_string(), serde_json::json!(start_pos))].into_iter().collect(),
                                ));
                            }
                        }
                    }
                    
                    if self.mouse_state.dragging {
                        responses.push(InteractionResponse::new(
                            event.target_element.clone(),
                            ResponseType::Drag,
                            [
                                ("position".to_string(), serde_json::json!(mouse_data.position)),
                                ("movement".to_string(), serde_json::json!(mouse_data.movement)),
                            ].into_iter().collect(),
                        ));
                    }
                }
                _ => {}
            }
            
            self.performance_metrics.mouse_events_processed += 1;
        }
        
        Ok(responses)
    }

    fn handle_touch_event(&mut self, event: &InteractionEvent) -> Result<Vec<InteractionResponse>, WASMError> {
        let mut responses = Vec::new();
        
        if let Some(touch_data) = &event.touch_data {
            match event.event_type {
                InteractionType::TouchStart => {
                    for touch in &touch_data.changed_touches {
                        let tracker = TouchTracker {
                            touch_id: touch.identifier,
                            start_position: touch.position.clone(),
                            current_position: touch.position.clone(),
                            start_time: event.timestamp,
                            last_update: event.timestamp,
                            velocity: Position { x: 0.0, y: 0.0 },
                            target_element: event.target_element.clone(),
                        };
                        self.touch_tracking.insert(touch.identifier, tracker);
                        
                        if let Some(target) = &event.target_element {
                            self.set_interaction_state(target, InteractionStateType::Active, event.timestamp);
                        }
                    }
                    
                    responses.push(InteractionResponse::new(
                        event.target_element.clone(),
                        ResponseType::TouchStart,
                        [("touch_count".to_string(), serde_json::json!(touch_data.touches.len()))].into_iter().collect(),
                    ));
                }
                InteractionType::TouchMove => {
                    for touch in &touch_data.changed_touches {
                        if let Some(tracker) = self.touch_tracking.get_mut(&touch.identifier) {
                            // Calculate velocity
                            let time_delta = event.timestamp - tracker.last_update;
                            if time_delta > 0.0 {
                                tracker.velocity.x = (touch.position.x - tracker.current_position.x) / time_delta;
                                tracker.velocity.y = (touch.position.y - tracker.current_position.y) / time_delta;
                            }
                            
                            tracker.current_position = touch.position.clone();
                            tracker.last_update = event.timestamp;
                        }
                    }
                    
                    responses.push(InteractionResponse::new(
                        event.target_element.clone(),
                        ResponseType::TouchMove,
                        [
                            ("touches".to_string(), serde_json::json!(touch_data.touches)),
                            ("scale".to_string(), serde_json::json!(touch_data.scale)),
                            ("rotation".to_string(), serde_json::json!(touch_data.rotation_angle)),
                        ].into_iter().collect(),
                    ));
                }
                InteractionType::TouchEnd | InteractionType::TouchCancel => {
                    for touch in &touch_data.changed_touches {
                        if let Some(tracker) = self.touch_tracking.remove(&touch.identifier) {
                            let duration = event.timestamp - tracker.start_time;
                            
                            // Check for tap gesture
                            let distance = ((touch.position.x - tracker.start_position.x).powi(2) + 
                                          (touch.position.y - tracker.start_position.y).powi(2)).sqrt();
                            
                            if distance < 10.0 && duration < 300.0 {
                                responses.push(InteractionResponse::new(
                                    event.target_element.clone(),
                                    ResponseType::Tap,
                                    [
                                        ("position".to_string(), serde_json::json!(touch.position)),
                                        ("duration".to_string(), serde_json::json!(duration)),
                                    ].into_iter().collect(),
                                ));
                            }
                        }
                        
                        if let Some(target) = &event.target_element {
                            self.set_interaction_state(target, InteractionStateType::Idle, event.timestamp);
                        }
                    }
                    
                    responses.push(InteractionResponse::new(
                        event.target_element.clone(),
                        ResponseType::TouchEnd,
                        [("remaining_touches".to_string(), serde_json::json!(touch_data.touches.len()))].into_iter().collect(),
                    ));
                }
                _ => {}
            }
            
            self.performance_metrics.touch_points_processed += touch_data.changed_touches.len() as u32;
        }
        
        Ok(responses)
    }

    fn handle_keyboard_event(&mut self, event: &InteractionEvent) -> Result<Vec<InteractionResponse>, WASMError> {
        let mut responses = Vec::new();
        
        if let Some(keyboard_data) = &event.keyboard_data {
            match event.event_type {
                InteractionType::KeyDown => {
                    self.keyboard_state.pressed_keys.insert(keyboard_data.key.clone(), event.timestamp);
                    
                    // Update modifiers
                    match keyboard_data.key.as_str() {
                        "Control" => self.keyboard_state.modifiers.ctrl = true,
                        "Shift" => self.keyboard_state.modifiers.shift = true,
                        "Alt" => self.keyboard_state.modifiers.alt = true,
                        "Meta" => self.keyboard_state.modifiers.meta = true,
                        _ => {}
                    }
                    
                    responses.push(InteractionResponse::new(
                        event.target_element.clone(),
                        ResponseType::KeyDown,
                        [
                            ("key".to_string(), serde_json::json!(keyboard_data.key)),
                            ("modifiers".to_string(), serde_json::json!(self.keyboard_state.modifiers)),
                        ].into_iter().collect(),
                    ));
                }
                InteractionType::KeyUp => {
                    self.keyboard_state.pressed_keys.remove(&keyboard_data.key);
                    
                    // Update modifiers
                    match keyboard_data.key.as_str() {
                        "Control" => self.keyboard_state.modifiers.ctrl = false,
                        "Shift" => self.keyboard_state.modifiers.shift = false,
                        "Alt" => self.keyboard_state.modifiers.alt = false,
                        "Meta" => self.keyboard_state.modifiers.meta = false,
                        _ => {}
                    }
                    
                    responses.push(InteractionResponse::new(
                        event.target_element.clone(),
                        ResponseType::KeyUp,
                        [("key".to_string(), serde_json::json!(keyboard_data.key))].into_iter().collect(),
                    ));
                }
                InteractionType::KeyPress => {
                    responses.push(InteractionResponse::new(
                        event.target_element.clone(),
                        ResponseType::KeyPress,
                        [
                            ("key".to_string(), serde_json::json!(keyboard_data.key)),
                            ("char_code".to_string(), serde_json::json!(keyboard_data.char_code)),
                        ].into_iter().collect(),
                    ));
                }
                _ => {}
            }
            
            self.performance_metrics.keyboard_events_processed += 1;
        }
        
        Ok(responses)
    }

    fn handle_click_event(&mut self, event: &InteractionEvent) -> Result<Vec<InteractionResponse>, WASMError> {
        let mut responses = Vec::new();
        
        if let Some(mouse_data) = &event.mouse_data {
            match event.event_type {
                InteractionType::Click => {
                    // Check for double-click
                    let time_since_last_click = event.timestamp - self.mouse_state.last_click_time;
                    if time_since_last_click < 500.0 {
                        self.mouse_state.click_count += 1;
                    } else {
                        self.mouse_state.click_count = 1;
                    }
                    self.mouse_state.last_click_time = event.timestamp;
                    
                    responses.push(InteractionResponse::new(
                        event.target_element.clone(),
                        ResponseType::Click,
                        [
                            ("position".to_string(), serde_json::json!(mouse_data.position)),
                            ("button".to_string(), serde_json::json!(mouse_data.button)),
                            ("click_count".to_string(), serde_json::json!(self.mouse_state.click_count)),
                        ].into_iter().collect(),
                    ));
                }
                InteractionType::DoubleClick => {
                    responses.push(InteractionResponse::new(
                        event.target_element.clone(),
                        ResponseType::DoubleClick,
                        [("position".to_string(), serde_json::json!(mouse_data.position))].into_iter().collect(),
                    ));
                }
                _ => {}
            }
        }
        
        Ok(responses)
    }

    fn handle_gesture_event(&mut self, event: &InteractionEvent) -> Result<Vec<InteractionResponse>, WASMError> {
        let mut responses = Vec::new();
        
        if let Some(gesture_data) = &event.gesture_data {
            responses.push(InteractionResponse::new(
                event.target_element.clone(),
                ResponseType::Gesture,
                [
                    ("gesture_type".to_string(), serde_json::json!(gesture_data.gesture_type)),
                    ("delta".to_string(), serde_json::json!(gesture_data.delta)),
                    ("scale".to_string(), serde_json::json!(gesture_data.scale)),
                    ("rotation".to_string(), serde_json::json!(gesture_data.rotation)),
                    ("velocity".to_string(), serde_json::json!(gesture_data.velocity)),
                ].into_iter().collect(),
            ));
        }
        
        Ok(responses)
    }

    fn handle_scroll_event(&mut self, event: &InteractionEvent) -> Result<Vec<InteractionResponse>, WASMError> {
        let mut responses = Vec::new();
        
        responses.push(InteractionResponse::new(
            event.target_element.clone(),
            ResponseType::Scroll,
            event.data.clone(),
        ));
        
        Ok(responses)
    }

    fn handle_focus_event(&mut self, event: &InteractionEvent) -> Result<Vec<InteractionResponse>, WASMError> {
        let mut responses = Vec::new();
        
        match event.event_type {
            InteractionType::Focus => {
                self.keyboard_state.focused_element = event.target_element.clone();
                if let Some(target) = &event.target_element {
                    self.set_interaction_state(target, InteractionStateType::Focused, event.timestamp);
                }
            }
            InteractionType::Blur => {
                if let Some(target) = &self.keyboard_state.focused_element {
                    self.set_interaction_state(target, InteractionStateType::Idle, event.timestamp);
                }
                self.keyboard_state.focused_element = None;
            }
            _ => {}
        }
        
        responses.push(InteractionResponse::new(
            event.target_element.clone(),
            ResponseType::FocusChanged,
            [("focused".to_string(), serde_json::json!(event.event_type == InteractionType::Focus))].into_iter().collect(),
        ));
        
        Ok(responses)
    }

    fn handle_resize_event(&mut self, event: &InteractionEvent) -> Result<Vec<InteractionResponse>, WASMError> {
        let mut responses = Vec::new();
        
        responses.push(InteractionResponse::new(
            None,
            ResponseType::Resize,
            event.data.clone(),
        ));
        
        Ok(responses)
    }

    fn delegate_event(&self, target: &str, event: &InteractionEvent) -> Result<Vec<InteractionResponse>, WASMError> {
        let mut responses = Vec::new();
        
        if let Some(delegates) = self.event_delegates.get(target) {
            for delegate in delegates {
                if delegate.event_types.contains(&event.event_type) {
                    responses.push(InteractionResponse::new(
                        Some(delegate.element_id.clone()),
                        ResponseType::Delegated,
                        [
                            ("handler_id".to_string(), serde_json::json!(delegate.handler_id)),
                            ("original_target".to_string(), serde_json::json!(target)),
                        ].into_iter().collect(),
                    ));
                }
            }
        }
        
        Ok(responses)
    }

    fn set_interaction_state(&mut self, element_id: &str, state_type: InteractionStateType, timestamp: f64) {
        let state = InteractionState {
            element_id: element_id.to_string(),
            state_type,
            start_time: timestamp,
            last_update: timestamp,
            properties: HashMap::new(),
        };
        self.interaction_states.insert(element_id.to_string(), state);
    }

    fn update_performance_metrics(&mut self, processing_time: f64) {
        let current_time = get_current_timestamp();
        let time_window = 1000.0; // 1 second window
        
        // Calculate events per second
        self.performance_metrics.events_per_second = 
            self.performance_metrics.total_events as f64 / (current_time / 1000.0);
        
        // Update average response time
        let total_time = self.performance_metrics.average_response_time * (self.performance_metrics.total_events - 1) as f64;
        self.performance_metrics.average_response_time = 
            (total_time + processing_time) / self.performance_metrics.total_events as f64;
    }

    pub fn add_event_delegate(&mut self, target: &str, delegate: EventDelegate) {
        self.event_delegates.entry(target.to_string())
            .or_insert_with(Vec::new)
            .push(delegate);
    }

    pub fn remove_event_delegate(&mut self, target: &str, handler_id: &str) {
        if let Some(delegates) = self.event_delegates.get_mut(target) {
            delegates.retain(|d| d.handler_id != handler_id);
        }
    }

    pub fn get_interaction_state(&self, element_id: &str) -> Option<&InteractionState> {
        self.interaction_states.get(element_id)
    }

    pub fn get_performance_metrics(&self) -> &InteractionMetrics {
        &self.performance_metrics
    }
}

#[derive(Serialize, Deserialize, Clone, Debug)]
pub struct InteractionResponse {
    pub target_element: Option<String>,
    pub response_type: ResponseType,
    pub data: HashMap<String, serde_json::Value>,
    pub timestamp: f64,
}

#[derive(Serialize, Deserialize, Clone, Debug)]
pub enum ResponseType {
    EventProcessed,
    StateChanged,
    Click,
    DoubleClick,
    TouchStart,
    TouchMove,
    TouchEnd,
    Tap,
    DragStart,
    Drag,
    DragEnd,
    KeyDown,
    KeyUp,
    KeyPress,
    Gesture,
    Scroll,
    FocusChanged,
    Resize,
    Delegated,
}

impl InteractionResponse {
    pub fn new(target_element: Option<String>, response_type: ResponseType, data: HashMap<String, serde_json::Value>) -> Self {
        Self {
            target_element,
            response_type,
            data,
            timestamp: get_current_timestamp(),
        }
    }
}

// Gesture Recognizer for advanced gesture detection
pub struct GestureRecognizer {
    gesture_configs: HashMap<GestureType, GestureConfig>,
    active_recognizers: HashMap<String, GestureRecognition>,
    gesture_history: Vec<GestureEvent>,
}

#[derive(Clone, Debug)]
pub struct GestureConfig {
    pub min_distance: f64,
    pub max_distance: f64,
    pub min_duration: f64,
    pub max_duration: f64,
    pub min_velocity: f64,
    pub max_velocity: f64,
    pub angle_tolerance: f64,
    pub scale_threshold: f64,
    pub rotation_threshold: f64,
}

#[derive(Clone, Debug)]
pub struct GestureRecognition {
    pub gesture_type: GestureType,
    pub start_time: f64,
    pub touch_points: Vec<TouchPoint>,
    pub samples: Vec<GestureSample>,
    pub confidence: f64,
}

#[derive(Clone, Debug)]
pub struct GestureSample {
    pub timestamp: f64,
    pub position: Position,
    pub velocity: Position,
    pub pressure: Option<f64>,
}

#[derive(Serialize, Deserialize, Clone, Debug)]
pub struct GestureEvent {
    pub gesture_type: GestureType,
    pub confidence: f64,
    pub start_position: Position,
    pub end_position: Position,
    pub duration: f64,
    pub velocity: Position,
    pub properties: HashMap<String, f64>,
    pub timestamp: f64,
}

impl GestureRecognizer {
    pub fn new() -> Self {
        let mut recognizer = Self {
            gesture_configs: HashMap::new(),
            active_recognizers: HashMap::new(),
            gesture_history: Vec::new(),
        };
        
        // Initialize default gesture configurations
        recognizer.init_default_configs();
        recognizer
    }

    fn init_default_configs(&mut self) {
        // Tap gesture
        self.gesture_configs.insert(GestureType::Tap, GestureConfig {
            min_distance: 0.0,
            max_distance: 10.0,
            min_duration: 50.0,
            max_duration: 300.0,
            min_velocity: 0.0,
            max_velocity: 100.0,
            angle_tolerance: 0.0,
            scale_threshold: 0.0,
            rotation_threshold: 0.0,
        });

        // Double tap gesture
        self.gesture_configs.insert(GestureType::DoubleTap, GestureConfig {
            min_distance: 0.0,
            max_distance: 20.0,
            min_duration: 50.0,
            max_duration: 600.0,
            min_velocity: 0.0,
            max_velocity: 150.0,
            angle_tolerance: 0.0,
            scale_threshold: 0.0,
            rotation_threshold: 0.0,
        });

        // Long press gesture
        self.gesture_configs.insert(GestureType::LongPress, GestureConfig {
            min_distance: 0.0,
            max_distance: 15.0,
            min_duration: 500.0,
            max_duration: f64::INFINITY,
            min_velocity: 0.0,
            max_velocity: 50.0,
            angle_tolerance: 0.0,
            scale_threshold: 0.0,
            rotation_threshold: 0.0,
        });

        // Swipe gesture
        self.gesture_configs.insert(GestureType::Swipe, GestureConfig {
            min_distance: 50.0,
            max_distance: f64::INFINITY,
            min_duration: 50.0,
            max_duration: 500.0,
            min_velocity: 100.0,
            max_velocity: f64::INFINITY,
            angle_tolerance: 30.0,
            scale_threshold: 0.0,
            rotation_threshold: 0.0,
        });

        // Pan gesture
        self.gesture_configs.insert(GestureType::Pan, GestureConfig {
            min_distance: 10.0,
            max_distance: f64::INFINITY,
            min_duration: 100.0,
            max_duration: f64::INFINITY,
            min_velocity: 0.0,
            max_velocity: f64::INFINITY,
            angle_tolerance: 0.0,
            scale_threshold: 0.0,
            rotation_threshold: 0.0,
        });

        // Pinch gesture
        self.gesture_configs.insert(GestureType::Pinch, GestureConfig {
            min_distance: 0.0,
            max_distance: f64::INFINITY,
            min_duration: 100.0,
            max_duration: f64::INFINITY,
            min_velocity: 0.0,
            max_velocity: f64::INFINITY,
            angle_tolerance: 0.0,
            scale_threshold: 0.1,
            rotation_threshold: 0.0,
        });

        // Rotate gesture
        self.gesture_configs.insert(GestureType::Rotate, GestureConfig {
            min_distance: 0.0,
            max_distance: f64::INFINITY,
            min_duration: 100.0,
            max_duration: f64::INFINITY,
            min_velocity: 0.0,
            max_velocity: f64::INFINITY,
            angle_tolerance: 0.0,
            scale_threshold: 0.0,
            rotation_threshold: 5.0, // 5 degrees
        });
    }

    pub fn process_touch_input(&mut self, touch_data: &TouchData, timestamp: f64) -> Vec<GestureEvent> {
        let mut detected_gestures = Vec::new();

        // Process single-touch gestures
        if touch_data.touches.len() == 1 {
            detected_gestures.extend(self.process_single_touch(&touch_data.touches[0], timestamp));
        }
        // Process multi-touch gestures
        else if touch_data.touches.len() >= 2 {
            detected_gestures.extend(self.process_multi_touch(&touch_data.touches, timestamp));
        }

        // Update gesture history
        for gesture in &detected_gestures {
            self.gesture_history.push(gesture.clone());
            
            // Limit history size
            if self.gesture_history.len() > 100 {
                self.gesture_history.remove(0);
            }
        }

        detected_gestures
    }

    fn process_single_touch(&mut self, touch: &TouchPoint, timestamp: f64) -> Vec<GestureEvent> {
        let mut gestures = Vec::new();
        let recognition_id = format!("single_{}", touch.identifier);

        if let Some(recognition) = self.active_recognizers.get_mut(&recognition_id) {
            // Update existing recognition
            recognition.samples.push(GestureSample {
                timestamp,
                position: touch.position.clone(),
                velocity: self.calculate_velocity(&recognition.samples, &touch.position, timestamp),
                pressure: touch.force,
            });

            // Check for gesture completion
            if let Some(gesture) = self.check_gesture_completion(recognition, timestamp) {
                gestures.push(gesture);
                self.active_recognizers.remove(&recognition_id);
            }
        } else {
            // Start new recognition
            let recognition = GestureRecognition {
                gesture_type: GestureType::Tap, // Default, will be determined later
                start_time: timestamp,
                touch_points: vec![touch.clone()],
                samples: vec![GestureSample {
                    timestamp,
                    position: touch.position.clone(),
                    velocity: Position { x: 0.0, y: 0.0 },
                    pressure: touch.force,
                }],
                confidence: 0.0,
            };
            self.active_recognizers.insert(recognition_id, recognition);
        }

        gestures
    }

    fn process_multi_touch(&mut self, touches: &[TouchPoint], timestamp: f64) -> Vec<GestureEvent> {
        let mut gestures = Vec::new();
        let recognition_id = "multi_touch".to_string();

        if touches.len() == 2 {
            // Process pinch and rotate gestures
            let touch1 = &touches[0];
            let touch2 = &touches[1];
            
            let distance = self.calculate_distance(&touch1.position, &touch2.position);
            let angle = self.calculate_angle(&touch1.position, &touch2.position);
            let center = Position {
                x: (touch1.position.x + touch2.position.x) / 2.0,
                y: (touch1.position.y + touch2.position.y) / 2.0,
            };

            if let Some(recognition) = self.active_recognizers.get_mut(&recognition_id) {
                // Check for pinch gesture
                if let Some(last_sample) = recognition.samples.last() {
                    let last_distance = self.extract_distance_from_sample(last_sample);
                    let scale_change = distance / last_distance;
                    
                    if (scale_change - 1.0).abs() > 0.1 {
                        gestures.push(GestureEvent {
                            gesture_type: GestureType::Pinch,
                            confidence: 0.9,
                            start_position: recognition.samples[0].position.clone(),
                            end_position: center,
                            duration: timestamp - recognition.start_time,
                            velocity: Position { x: 0.0, y: 0.0 },
                            properties: [
                                ("scale".to_string(), scale_change),
                                ("distance".to_string(), distance),
                            ].into_iter().collect(),
                            timestamp,
                        });
                    }
                }

                recognition.samples.push(GestureSample {
                    timestamp,
                    position: center,
                    velocity: Position { x: 0.0, y: 0.0 },
                    pressure: None,
                });
            } else {
                // Start new multi-touch recognition
                let recognition = GestureRecognition {
                    gesture_type: GestureType::Pinch,
                    start_time: timestamp,
                    touch_points: touches.to_vec(),
                    samples: vec![GestureSample {
                        timestamp,
                        position: center,
                        velocity: Position { x: 0.0, y: 0.0 },
                        pressure: None,
                    }],
                    confidence: 0.0,
                };
                self.active_recognizers.insert(recognition_id, recognition);
            }
        }

        gestures
    }

    fn check_gesture_completion(&self, recognition: &GestureRecognition, timestamp: f64) -> Option<GestureEvent> {
        if recognition.samples.len() < 2 {
            return None;
        }

        let duration = timestamp - recognition.start_time;
        let start_pos = &recognition.samples[0].position;
        let end_pos = &recognition.samples.last().unwrap().position;
        let distance = self.calculate_distance(start_pos, end_pos);
        let velocity = self.calculate_average_velocity(&recognition.samples);

        // Check each gesture type
        for (gesture_type, config) in &self.gesture_configs {
            if self.matches_gesture_config(recognition, config, distance, duration, &velocity) {
                return Some(GestureEvent {
                    gesture_type: gesture_type.clone(),
                    confidence: self.calculate_confidence(recognition, config),
                    start_position: start_pos.clone(),
                    end_position: end_pos.clone(),
                    duration,
                    velocity,
                    properties: self.extract_gesture_properties(recognition, gesture_type),
                    timestamp,
                });
            }
        }

        None
    }

    fn matches_gesture_config(&self, recognition: &GestureRecognition, config: &GestureConfig, distance: f64, duration: f64, velocity: &Position) -> bool {
        let velocity_magnitude = (velocity.x.powi(2) + velocity.y.powi(2)).sqrt();
        
        distance >= config.min_distance &&
        distance <= config.max_distance &&
        duration >= config.min_duration &&
        duration <= config.max_duration &&
        velocity_magnitude >= config.min_velocity &&
        velocity_magnitude <= config.max_velocity
    }

    fn calculate_confidence(&self, recognition: &GestureRecognition, config: &GestureConfig) -> f64 {
        // Simple confidence calculation based on how well the gesture matches the config
        let mut confidence = 1.0;
        
        // Factor in sample consistency
        if recognition.samples.len() > 2 {
            let velocity_variance = self.calculate_velocity_variance(&recognition.samples);
            confidence *= (1.0 - velocity_variance.min(1.0));
        }
        
        confidence.max(0.0).min(1.0)
    }

    fn extract_gesture_properties(&self, recognition: &GestureRecognition, gesture_type: &GestureType) -> HashMap<String, f64> {
        let mut properties = HashMap::new();
        
        match gesture_type {
            GestureType::Swipe => {
                if let (Some(first), Some(last)) = (recognition.samples.first(), recognition.samples.last()) {
                    let angle = self.calculate_angle(&first.position, &last.position);
                    properties.insert("angle".to_string(), angle);
                    properties.insert("direction".to_string(), self.angle_to_direction(angle));
                }
            }
            GestureType::Pan => {
                let total_distance = self.calculate_total_path_distance(&recognition.samples);
                properties.insert("total_distance".to_string(), total_distance);
            }
            _ => {}
        }
        
        properties
    }

    fn calculate_distance(&self, pos1: &Position, pos2: &Position) -> f64 {
        ((pos2.x - pos1.x).powi(2) + (pos2.y - pos1.y).powi(2)).sqrt()
    }

    fn calculate_angle(&self, pos1: &Position, pos2: &Position) -> f64 {
        (pos2.y - pos1.y).atan2(pos2.x - pos1.x).to_degrees()
    }

    fn calculate_velocity(&self, samples: &[GestureSample], current_pos: &Position, timestamp: f64) -> Position {
        if let Some(last_sample) = samples.last() {
            let time_delta = timestamp - last_sample.timestamp;
            if time_delta > 0.0 {
                return Position {
                    x: (current_pos.x - last_sample.position.x) / time_delta,
                    y: (current_pos.y - last_sample.position.y) / time_delta,
                };
            }
        }
        Position { x: 0.0, y: 0.0 }
    }

    fn calculate_average_velocity(&self, samples: &[GestureSample]) -> Position {
        if samples.len() < 2 {
            return Position { x: 0.0, y: 0.0 };
        }

        let mut total_velocity = Position { x: 0.0, y: 0.0 };
        let mut count = 0;

        for i in 1..samples.len() {
            let time_delta = samples[i].timestamp - samples[i-1].timestamp;
            if time_delta > 0.0 {
                total_velocity.x += (samples[i].position.x - samples[i-1].position.x) / time_delta;
                total_velocity.y += (samples[i].position.y - samples[i-1].position.y) / time_delta;
                count += 1;
            }
        }

        if count > 0 {
            total_velocity.x /= count as f64;
            total_velocity.y /= count as f64;
        }

        total_velocity
    }

    fn calculate_velocity_variance(&self, samples: &[GestureSample]) -> f64 {
        if samples.len() < 3 {
            return 0.0;
        }

        let mut velocities = Vec::new();
        for i in 1..samples.len() {
            let time_delta = samples[i].timestamp - samples[i-1].timestamp;
            if time_delta > 0.0 {
                let velocity = Position {
                    x: (samples[i].position.x - samples[i-1].position.x) / time_delta,
                    y: (samples[i].position.y - samples[i-1].position.y) / time_delta,
                };
                velocities.push((velocity.x.powi(2) + velocity.y.powi(2)).sqrt());
            }
        }

        if velocities.is_empty() {
            return 0.0;
        }

        let mean = velocities.iter().sum::<f64>() / velocities.len() as f64;
        let variance = velocities.iter()
            .map(|v| (v - mean).powi(2))
            .sum::<f64>() / velocities.len() as f64;
        
        variance.sqrt() / mean.max(1.0)
    }

    fn calculate_total_path_distance(&self, samples: &[GestureSample]) -> f64 {
        let mut total_distance = 0.0;
        for i in 1..samples.len() {
            total_distance += self.calculate_distance(&samples[i-1].position, &samples[i].position);
        }
        total_distance
    }

    fn extract_distance_from_sample(&self, sample: &GestureSample) -> f64 {
        // This would extract distance from multi-touch sample data
        // For now, return a default value
        100.0
    }

    fn angle_to_direction(&self, angle: f64) -> f64 {
        // Convert angle to direction (0=right, 1=down, 2=left, 3=up)
        let normalized_angle = ((angle + 360.0) % 360.0) / 90.0;
        normalized_angle.round() % 4.0
    }

    pub fn clear_completed_recognitions(&mut self) {
        // Remove recognitions that have been inactive for too long
        let current_time = get_current_timestamp();
        let timeout = 1000.0; // 1 second timeout
        
        self.active_recognizers.retain(|_, recognition| {
            current_time - recognition.start_time < timeout
        });
    }

    pub fn get_gesture_history(&self) -> &[GestureEvent] {
        &self.gesture_history
    }
}

// Responsive Adapter for optimizing interactions based on device capabilities
pub struct ResponsiveAdapter {
    device_info: DeviceInfo,
    interaction_settings: InteractionSettings,
    performance_profile: PerformanceProfile,
    adaptive_thresholds: AdaptiveThresholds,
}

#[derive(Clone, Debug)]
pub struct DeviceInfo {
    pub device_type: DeviceType,
    pub screen_size: Size,
    pub pixel_density: f64,
    pub touch_support: bool,
    pub mouse_support: bool,
    pub keyboard_support: bool,
    pub max_touch_points: u32,
    pub has_force_touch: bool,
    pub has_hover_support: bool,
}

#[derive(Clone, Debug)]
pub enum DeviceType {
    Desktop,
    Tablet,
    Mobile,
    TV,
    Watch,
    Unknown,
}

#[derive(Clone, Debug)]
pub struct InteractionSettings {
    pub touch_target_size: f64,
    pub tap_timeout: f64,
    pub double_tap_timeout: f64,
    pub long_press_timeout: f64,
    pub drag_threshold: f64,
    pub scroll_sensitivity: f64,
    pub gesture_sensitivity: f64,
    pub hover_delay: f64,
}

#[derive(Clone, Debug)]
pub struct PerformanceProfile {
    pub target_fps: f64,
    pub max_event_frequency: f64,
    pub throttle_threshold: f64,
    pub batch_events: bool,
    pub use_passive_listeners: bool,
    pub debounce_scroll: bool,
    pub optimize_animations: bool,
}

#[derive(Clone, Debug)]
pub struct AdaptiveThresholds {
    pub min_touch_target: f64,
    pub max_touch_target: f64,
    pub min_drag_distance: f64,
    pub max_drag_distance: f64,
    pub velocity_scaling: f64,
    pub pressure_sensitivity: f64,
}

impl ResponsiveAdapter {
    pub fn new() -> Self {
        Self {
            device_info: DeviceInfo::default(),
            interaction_settings: InteractionSettings::default(),
            performance_profile: PerformanceProfile::default(),
            adaptive_thresholds: AdaptiveThresholds::default(),
        }
    }

    pub fn initialize_device_detection(&mut self, viewport: &Viewport) -> Result<(), WASMError> {
        // Detect device type based on screen size and capabilities
        self.device_info.screen_size = Size {
            width: viewport.width,
            height: viewport.height,
        };

        // Determine device type
        let screen_diagonal = (viewport.width.powi(2) + viewport.height.powi(2)).sqrt();
        self.device_info.device_type = if screen_diagonal < 600.0 {
            DeviceType::Mobile
        } else if screen_diagonal < 1200.0 {
            DeviceType::Tablet
        } else {
            DeviceType::Desktop
        };

        // Adapt interaction settings based on device type
        self.adapt_interaction_settings();
        self.adapt_performance_profile();
        self.adapt_thresholds();

        Ok(())
    }

    fn adapt_interaction_settings(&mut self) {
        match self.device_info.device_type {
            DeviceType::Mobile => {
                self.interaction_settings.touch_target_size = 44.0; // iOS HIG recommendation
                self.interaction_settings.tap_timeout = 300.0;
                self.interaction_settings.double_tap_timeout = 500.0;
                self.interaction_settings.long_press_timeout = 500.0;
                self.interaction_settings.drag_threshold = 10.0;
                self.interaction_settings.scroll_sensitivity = 1.0;
                self.interaction_settings.gesture_sensitivity = 1.0;
                self.interaction_settings.hover_delay = 0.0; // No hover on mobile
            }
            DeviceType::Tablet => {
                self.interaction_settings.touch_target_size = 48.0;
                self.interaction_settings.tap_timeout = 250.0;
                self.interaction_settings.double_tap_timeout = 400.0;
                self.interaction_settings.long_press_timeout = 600.0;
                self.interaction_settings.drag_threshold = 8.0;
                self.interaction_settings.scroll_sensitivity = 0.8;
                self.interaction_settings.gesture_sensitivity = 1.2;
                self.interaction_settings.hover_delay = 100.0;
            }
            DeviceType::Desktop => {
                self.interaction_settings.touch_target_size = 32.0;
                self.interaction_settings.tap_timeout = 200.0;
                self.interaction_settings.double_tap_timeout = 300.0;
                self.interaction_settings.long_press_timeout = 800.0;
                self.interaction_settings.drag_threshold = 5.0;
                self.interaction_settings.scroll_sensitivity = 0.6;
                self.interaction_settings.gesture_sensitivity = 0.8;
                self.interaction_settings.hover_delay = 200.0;
            }
            _ => {
                // Use default settings
            }
        }
    }

    fn adapt_performance_profile(&mut self) {
        match self.device_info.device_type {
            DeviceType::Mobile => {
                self.performance_profile.target_fps = 60.0;
                self.performance_profile.max_event_frequency = 120.0; // Hz
                self.performance_profile.throttle_threshold = 100.0;
                self.performance_profile.batch_events = true;
                self.performance_profile.use_passive_listeners = true;
                self.performance_profile.debounce_scroll = true;
                self.performance_profile.optimize_animations = true;
            }
            DeviceType::Tablet => {
                self.performance_profile.target_fps = 60.0;
                self.performance_profile.max_event_frequency = 144.0;
                self.performance_profile.throttle_threshold = 80.0;
                self.performance_profile.batch_events = true;
                self.performance_profile.use_passive_listeners = true;
                self.performance_profile.debounce_scroll = false;
                self.performance_profile.optimize_animations = false;
            }
            DeviceType::Desktop => {
                self.performance_profile.target_fps = 120.0;
                self.performance_profile.max_event_frequency = 240.0;
                self.performance_profile.throttle_threshold = 60.0;
                self.performance_profile.batch_events = false;
                self.performance_profile.use_passive_listeners = false;
                self.performance_profile.debounce_scroll = false;
                self.performance_profile.optimize_animations = false;
            }
            _ => {}
        }
    }

    fn adapt_thresholds(&mut self) {
        let pixel_density = self.device_info.pixel_density;
        
        self.adaptive_thresholds.min_touch_target = 
            self.interaction_settings.touch_target_size * 0.8 * pixel_density;
        self.adaptive_thresholds.max_touch_target = 
            self.interaction_settings.touch_target_size * 1.5 * pixel_density;
        
        self.adaptive_thresholds.min_drag_distance = 
            self.interaction_settings.drag_threshold * pixel_density;
        self.adaptive_thresholds.max_drag_distance = 
            self.interaction_settings.drag_threshold * 10.0 * pixel_density;
        
        // Adjust velocity scaling based on device type
        self.adaptive_thresholds.velocity_scaling = match self.device_info.device_type {
            DeviceType::Mobile => 1.2,
            DeviceType::Tablet => 1.0,
            DeviceType::Desktop => 0.8,
            _ => 1.0,
        };
        
        self.adaptive_thresholds.pressure_sensitivity = if self.device_info.has_force_touch {
            1.0
        } else {
            0.0
        };
    }

    pub fn adapt_event(&self, event: &mut InteractionEvent) -> Result<(), WASMError> {
        // Adapt event based on device capabilities and settings
        match event.event_type {
            InteractionType::TouchStart | InteractionType::TouchMove | InteractionType::TouchEnd => {
                self.adapt_touch_event(event)?;
            }
            InteractionType::MouseMove | InteractionType::MouseDown | InteractionType::MouseUp => {
                self.adapt_mouse_event(event)?;
            }
            InteractionType::Scroll | InteractionType::Wheel => {
                self.adapt_scroll_event(event)?;
            }
            _ => {}
        }

        // Apply performance optimizations
        self.apply_performance_optimizations(event)?;

        Ok(())
    }

    fn adapt_touch_event(&self, event: &mut InteractionEvent) -> Result<(), WASMError> {
        if let Some(touch_data) = &mut event.touch_data {
            // Adjust touch target sizes
            for touch in &mut touch_data.touches {
                if let Some(radius) = &mut touch.radius {
                    *radius = (*radius).max(self.adaptive_thresholds.min_touch_target)
                                     .min(self.adaptive_thresholds.max_touch_target);
                }
                
                // Adjust force sensitivity
                if let Some(force) = &mut touch.force {
                    *force *= self.adaptive_thresholds.pressure_sensitivity;
                }
            }
            
            // Apply velocity scaling for gestures
            if let Some(scale) = &mut touch_data.scale {
                *scale = 1.0 + (*scale - 1.0) * self.adaptive_thresholds.velocity_scaling;
            }
        }
        
        Ok(())
    }

    fn adapt_mouse_event(&self, event: &mut InteractionEvent) -> Result<(), WASMError> {
        if let Some(mouse_data) = &mut event.mouse_data {
            // Adjust mouse movement sensitivity based on device
            if let Some(movement) = &mut mouse_data.movement {
                movement.x *= self.adaptive_thresholds.velocity_scaling;
                movement.y *= self.adaptive_thresholds.velocity_scaling;
            }
            
            // Adjust wheel sensitivity
            if let Some(wheel_delta) = &mut mouse_data.wheel_delta {
                wheel_delta.x *= self.interaction_settings.scroll_sensitivity;
                wheel_delta.y *= self.interaction_settings.scroll_sensitivity;
            }
        }
        
        Ok(())
    }

    fn adapt_scroll_event(&self, event: &mut InteractionEvent) -> Result<(), WASMError> {
        // Apply scroll sensitivity adjustments
        if let Some(delta_x) = event.data.get_mut("deltaX") {
            if let Some(delta_x_val) = delta_x.as_f64() {
                *delta_x = serde_json::json!(delta_x_val * self.interaction_settings.scroll_sensitivity);
            }
        }
        
        if let Some(delta_y) = event.data.get_mut("deltaY") {
            if let Some(delta_y_val) = delta_y.as_f64() {
                *delta_y = serde_json::json!(delta_y_val * self.interaction_settings.scroll_sensitivity);
            }
        }
        
        Ok(())
    }

    fn apply_performance_optimizations(&self, event: &mut InteractionEvent) -> Result<(), WASMError> {
        // Throttle high-frequency events
        let current_time = get_current_timestamp();
        let time_threshold = 1000.0 / self.performance_profile.max_event_frequency;
        
        // Add throttling information to event data
        event.data.insert("throttle_threshold".to_string(), serde_json::json!(time_threshold));
        event.data.insert("batch_events".to_string(), serde_json::json!(self.performance_profile.batch_events));
        event.data.insert("use_passive".to_string(), serde_json::json!(self.performance_profile.use_passive_listeners));
        
        Ok(())
    }

    pub fn should_throttle_event(&self, event_type: &InteractionType, last_event_time: f64) -> bool {
        let current_time = get_current_timestamp();
        let time_since_last = current_time - last_event_time;
        let min_interval = 1000.0 / self.performance_profile.max_event_frequency;
        
        match event_type {
            InteractionType::MouseMove | InteractionType::TouchMove | 
            InteractionType::Scroll | InteractionType::Wheel => {
                time_since_last < min_interval
            }
            _ => false,
        }
    }

    pub fn get_optimal_touch_target_size(&self, element_size: &Size) -> Size {
        let min_size = self.adaptive_thresholds.min_touch_target;
        Size {
            width: element_size.width.max(min_size),
            height: element_size.height.max(min_size),
        }
    }

    pub fn get_interaction_settings(&self) -> &InteractionSettings {
        &self.interaction_settings
    }

    pub fn get_performance_profile(&self) -> &PerformanceProfile {
        &self.performance_profile
    }

    pub fn update_device_info(&mut self, device_info: DeviceInfo) {
        self.device_info = device_info;
        self.adapt_interaction_settings();
        self.adapt_performance_profile();
        self.adapt_thresholds();
    }
}

impl Default for DeviceInfo {
    fn default() -> Self {
        Self {
            device_type: DeviceType::Desktop,
            screen_size: Size { width: 1920.0, height: 1080.0 },
            pixel_density: 1.0,
            touch_support: false,
            mouse_support: true,
            keyboard_support: true,
            max_touch_points: 0,
            has_force_touch: false,
            has_hover_support: true,
        }
    }
}

impl Default for InteractionSettings {
    fn default() -> Self {
        Self {
            touch_target_size: 44.0,
            tap_timeout: 300.0,
            double_tap_timeout: 500.0,
            long_press_timeout: 500.0,
            drag_threshold: 10.0,
            scroll_sensitivity: 1.0,
            gesture_sensitivity: 1.0,
            hover_delay: 200.0,
        }
    }
}

impl Default for PerformanceProfile {
    fn default() -> Self {
        Self {
            target_fps: 60.0,
            max_event_frequency: 120.0,
            throttle_threshold: 100.0,
            batch_events: false,
            use_passive_listeners: false,
            debounce_scroll: false,
            optimize_animations: false,
        }
    }
}

impl Default for AdaptiveThresholds {
    fn default() -> Self {
        Self {
            min_touch_target: 32.0,
            max_touch_target: 64.0,
            min_drag_distance: 5.0,
            max_drag_distance: 50.0,
            velocity_scaling: 1.0,
            pressure_sensitivity: 1.0,
        }
    }
}

// Event Processor for handling user interactions
pub struct EventProcessor;

impl EventProcessor {
    pub fn new() -> Self {
        Self
    }

    pub fn process_event(
        &self, 
        document_state: &mut DocumentState, 
        event: InteractionEvent
    ) -> Result<Vec<ElementChange>, WASMError> {
        let mut changes = Vec::new();
        
        match event.event_type {
            InteractionType::Click => {
                if let Some(target_element) = &event.target_element {
                    changes.extend(self.handle_click(document_state, target_element, &event)?);
                }
            }
            InteractionType::Hover => {
                if let Some(target_element) = &event.target_element {
                    changes.extend(self.handle_hover(document_state, target_element, &event)?);
                }
            }
            InteractionType::DataUpdate => {
                changes.extend(self.handle_data_update(document_state, &event)?);
            }
            _ => {
                // Handle other interaction types as needed
            }
        }
        
        Ok(changes)
    }

    fn handle_click(
        &self, 
        document_state: &mut DocumentState, 
        target_element: &str, 
        event: &InteractionEvent
    ) -> Result<Vec<ElementChange>, WASMError> {
        let mut changes = Vec::new();
        
        // Find the target element and its event handlers
        if let Some(element) = document_state.elements.iter().find(|e| e.id == target_element) {
            for handler in &element.event_handlers {
                if handler.event_type == "click" {
                    // Execute the event handler logic
                    changes.extend(self.execute_event_handler(document_state, handler, event)?);
                }
            }
            
            // Add visual feedback for click
            changes.push(ElementChange::Update {
                element_id: target_element.to_string(),
                properties: [
                    ("style.transform".to_string(), serde_json::Value::String("scale(0.95)".to_string())),
                ].into_iter().collect(),
            });
            
            // Reset visual feedback after short delay (simulated)
            changes.push(ElementChange::Update {
                element_id: target_element.to_string(),
                properties: [
                    ("style.transform".to_string(), serde_json::Value::String("scale(1.0)".to_string())),
                ].into_iter().collect(),
            });
        }
        
        Ok(changes)
    }

    fn handle_hover(
        &self, 
        document_state: &mut DocumentState, 
        target_element: &str, 
        event: &InteractionEvent
    ) -> Result<Vec<ElementChange>, WASMError> {
        let mut changes = Vec::new();
        
        // Check if this is hover enter or leave based on event data
        let is_hover_enter = event.data.get("hover_state")
            .and_then(|v| v.as_str())
            .map(|s| s == "enter")
            .unwrap_or(true);
        
        if let Some(element) = document_state.elements.iter().find(|e| e.id == target_element) {
            // Execute hover event handlers
            for handler in &element.event_handlers {
                if handler.event_type == "hover" {
                    changes.extend(self.execute_event_handler(document_state, handler, event)?);
                }
            }
            
            // Add visual hover effects
            if is_hover_enter {
                changes.push(ElementChange::Update {
                    element_id: target_element.to_string(),
                    properties: [
                        ("style.opacity".to_string(), serde_json::Value::String("0.8".to_string())),
                        ("style.cursor".to_string(), serde_json::Value::String("pointer".to_string())),
                    ].into_iter().collect(),
                });
            } else {
                changes.push(ElementChange::Update {
                    element_id: target_element.to_string(),
                    properties: [
                        ("style.opacity".to_string(), serde_json::Value::String("1.0".to_string())),
                        ("style.cursor".to_string(), serde_json::Value::String("default".to_string())),
                    ].into_iter().collect(),
                });
            }
        }
        
        Ok(changes)
    }

    fn handle_data_update(
        &self, 
        document_state: &mut DocumentState, 
        event: &InteractionEvent
    ) -> Result<Vec<ElementChange>, WASMError> {
        let mut changes = Vec::new();
        
        // Extract data source ID and new data from event
        let data_source_id = event.data.get("data_source_id")
            .and_then(|v| v.as_str())
            .ok_or_else(|| WASMError::new("INVALID_DATA_UPDATE", "Missing data source ID"))?;
        
        let new_data = event.data.get("data")
            .ok_or_else(|| WASMError::new("INVALID_DATA_UPDATE", "Missing data"))?;
        
        // Update data source
        if let Some(data_source) = document_state.data_sources.get_mut(data_source_id) {
            data_source.data = new_data.clone();
            
            // Find elements that depend on this data source
            let dependent_elements: Vec<String> = document_state.elements.iter()
                .filter(|e| {
                    e.properties.get("data_source")
                        .and_then(|v| v.as_str())
                        .map(|s| s == data_source_id)
                        .unwrap_or(false)
                })
                .map(|e| e.id.clone())
                .collect();
            
            // Update dependent elements
            for element_id in dependent_elements {
                changes.push(ElementChange::Update {
                    element_id: element_id.clone(),
                    properties: [
                        ("data".to_string(), new_data.clone()),
                        ("data_updated".to_string(), serde_json::Value::Bool(true)),
                    ].into_iter().collect(),
                });
                
                // Mark as dirty for re-rendering
                if !document_state.render_tree.dirty_nodes.contains(&element_id) {
                    document_state.render_tree.dirty_nodes.push(element_id);
                }
            }
        }
        
        Ok(changes)
    }

    fn execute_event_handler(
        &self, 
        document_state: &mut DocumentState, 
        handler: &EventHandler,
        event: &InteractionEvent
    ) -> Result<Vec<ElementChange>, WASMError> {
        let mut changes = Vec::new();
        
        // Execute handler based on handler ID
        match handler.handler_id.as_str() {
            "toggle_visibility" => {
                // Toggle element visibility
                if let Some(target_id) = handler.parameters.get("target")
                    .and_then(|v| v.as_str()) {
                    
                    if let Some(render_node) = document_state.render_tree.nodes.get_mut(target_id) {
                        render_node.visible = !render_node.visible;
                        changes.push(ElementChange::Update {
                            element_id: target_id.to_string(),
                            properties: [
                                ("style.display".to_string(), 
                                 serde_json::Value::String(if render_node.visible { "block" } else { "none" }.to_string())),
                            ].into_iter().collect(),
                        });
                    }
                }
            }
            "change_color" => {
                // Change element color
                if let Some(target_id) = handler.parameters.get("target")
                    .and_then(|v| v.as_str()) {
                    
                    let new_color = handler.parameters.get("color")
                        .and_then(|v| v.as_str())
                        .unwrap_or("#ff0000");
                    
                    changes.push(ElementChange::Update {
                        element_id: target_id.to_string(),
                        properties: [
                            ("style.backgroundColor".to_string(), serde_json::Value::String(new_color.to_string())),
                        ].into_iter().collect(),
                    });
                }
            }
            "update_text" => {
                // Update element text content
                if let Some(target_id) = handler.parameters.get("target")
                    .and_then(|v| v.as_str()) {
                    
                    let new_text = handler.parameters.get("text")
                        .and_then(|v| v.as_str())
                        .unwrap_or("Updated text");
                    
                    changes.push(ElementChange::Update {
                        element_id: target_id.to_string(),
                        properties: [
                            ("textContent".to_string(), serde_json::Value::String(new_text.to_string())),
                        ].into_iter().collect(),
                    });
                }
            }
            "trigger_animation" => {
                // Trigger an animation
                if let Some(animation_id) = handler.parameters.get("animation_id")
                    .and_then(|v| v.as_str()) {
                    
                    // Find and start the animation
                    if let Some(animation) = document_state.animations.iter().find(|a| a.id == animation_id) {
                        changes.push(ElementChange::AnimationUpdate {
                            animation_id: animation_id.to_string(),
                            progress: 0.0,
                            values: HashMap::new(),
                        });
                    }
                }
            }
            "navigate_to" => {
                // Handle navigation (would communicate with host)
                if let Some(url) = handler.parameters.get("url")
                    .and_then(|v| v.as_str()) {
                    
                    // This would typically send a message to the host application
                    changes.push(ElementChange::Update {
                        element_id: "navigation".to_string(),
                        properties: [
                            ("navigate_to".to_string(), serde_json::Value::String(url.to_string())),
                        ].into_iter().collect(),
                    });
                }
            }
            _ => {
                // Custom handler - could be extended
                return Err(WASMError::new("UNKNOWN_HANDLER", &format!("Unknown event handler: {}", handler.handler_id)));
            }
        }
        
        Ok(changes)
    }
}

// Render Cache for optimization
pub struct RenderCache {
    cached_updates: HashMap<String, RenderUpdate>,
    cache_size_limit: usize,
}

impl RenderCache {
    pub fn new() -> Self {
        Self {
            cached_updates: HashMap::new(),
            cache_size_limit: 100, // Limit cache size
        }
    }

    pub fn cache_update(&mut self, update: &RenderUpdate) {
        if self.cached_updates.len() >= self.cache_size_limit {
            // Simple LRU: remove oldest entry
            if let Some(oldest_key) = self.cached_updates.keys().next().cloned() {
                self.cached_updates.remove(&oldest_key);
            }
        }
        
        let cache_key = format!("update_{}", update.timestamp);
        self.cached_updates.insert(cache_key, update.clone());
    }
}

// Performance Monitor
pub struct PerformanceMonitor {
    interaction_count: u32,
    render_count: u32,
    start_time: f64,
}

impl PerformanceMonitor {
    pub fn new() -> Self {
        Self {
            interaction_count: 0,
            render_count: 0,
            start_time: get_current_timestamp(),
        }
    }

    pub fn record_interaction(&mut self) {
        self.interaction_count += 1;
    }

    pub fn record_render(&mut self) {
        self.render_count += 1;
    }

    pub fn get_stats(&self) -> PerformanceStats {
        let elapsed = get_current_timestamp() - self.start_time;
        PerformanceStats {
            interactions_per_second: if elapsed > 0.0 { 
                (self.interaction_count as f64) / (elapsed / 1000.0) 
            } else { 
                0.0 
            },
            renders_per_second: if elapsed > 0.0 { 
                (self.render_count as f64) / (elapsed / 1000.0) 
            } else { 
                0.0 
            },
            total_interactions: self.interaction_count,
            total_renders: self.render_count,
            uptime_ms: elapsed,
        }
    }
}

#[derive(Serialize, Deserialize, Clone, Debug)]
pub struct PerformanceStats {
    pub interactions_per_second: f64,
    pub renders_per_second: f64,
    pub total_interactions: u32,
    pub total_renders: u32,
    pub uptime_ms: f64,
}

// Element Change types for render updates
#[derive(Clone, Debug)]
pub enum ElementChange {
    Create {
        element_id: String,
        element_type: ElementType,
        parent_id: Option<String>,
    },
    Update {
        element_id: String,
        properties: HashMap<String, serde_json::Value>,
    },
    Remove {
        element_id: String,
    },
    AnimationUpdate {
        animation_id: String,
        progress: f64,
        values: HashMap<String, serde_json::Value>,
    },
}

impl ElementType {
    pub fn to_tag(&self) -> String {
        match self {
            ElementType::Chart => "div".to_string(),
            ElementType::Animation => "div".to_string(),
            ElementType::Interactive => "div".to_string(),
            ElementType::Vector => "svg".to_string(),
            ElementType::Text => "span".to_string(),
            ElementType::Image => "img".to_string(),
            ElementType::Container => "div".to_string(),
        }
    }
}

impl RenderUpdate {
    pub fn empty() -> Self {
        Self {
            dom_operations: Vec::new(),
            style_changes: Vec::new(),
            animation_updates: Vec::new(),
            timestamp: get_current_timestamp(),
        }
    }
}

// Helper functions
fn get_current_timestamp() -> f64 {
    SystemTime::now()
        .duration_since(UNIX_EPOCH)
        .unwrap_or_default()
        .as_millis() as f64
}

fn apply_easing(progress: f64, easing: &EasingFunction) -> f64 {
    match easing {
        EasingFunction::Linear => progress,
        EasingFunction::EaseIn => progress * progress,
        EasingFunction::EaseOut => 1.0 - (1.0 - progress) * (1.0 - progress),
        EasingFunction::EaseInOut => {
            if progress < 0.5 {
                2.0 * progress * progress
            } else {
                1.0 - 2.0 * (1.0 - progress) * (1.0 - progress)
            }
        }
        EasingFunction::Cubic(x1, y1, x2, y2) => {
            // Simplified cubic bezier approximation
            let t = progress;
            let t2 = t * t;
            let t3 = t2 * t;
            let mt = 1.0 - t;
            let mt2 = mt * mt;
            let mt3 = mt2 * mt;
            
            mt3 * 0.0 + 3.0 * mt2 * t * y1 + 3.0 * mt * t2 * y2 + t3 * 1.0
        }
    }
}

fn interpolate_keyframes(keyframes: &[Keyframe], progress: f64) -> HashMap<String, serde_json::Value> {
    let mut result = HashMap::new();
    
    if keyframes.is_empty() {
        return result;
    }
    
    // Find the two keyframes to interpolate between
    let mut prev_keyframe = &keyframes[0];
    let mut next_keyframe = &keyframes[keyframes.len() - 1];
    
    for i in 0..keyframes.len() - 1 {
        if keyframes[i].time <= progress && keyframes[i + 1].time >= progress {
            prev_keyframe = &keyframes[i];
            next_keyframe = &keyframes[i + 1];
            break;
        }
    }
    
    // Calculate interpolation factor
    let time_diff = next_keyframe.time - prev_keyframe.time;
    let local_progress = if time_diff > 0.0 {
        (progress - prev_keyframe.time) / time_diff
    } else {
        0.0
    };
    
    // Interpolate properties
    for (key, prev_value) in &prev_keyframe.properties {
        if let Some(next_value) = next_keyframe.properties.get(key) {
            let interpolated = interpolate_values(prev_value, next_value, local_progress);
            result.insert(key.clone(), interpolated);
        } else {
            result.insert(key.clone(), prev_value.clone());
        }
    }
    
    result
}

fn interpolate_values(
    prev: &serde_json::Value, 
    next: &serde_json::Value, 
    progress: f64
) -> serde_json::Value {
    match (prev, next) {
        (serde_json::Value::Number(p), serde_json::Value::Number(n)) => {
            let prev_f = p.as_f64().unwrap_or(0.0);
            let next_f = n.as_f64().unwrap_or(0.0);
            let interpolated = prev_f + (next_f - prev_f) * progress;
            serde_json::Value::Number(serde_json::Number::from_f64(interpolated).unwrap_or(serde_json::Number::from(0)))
        }
        _ => {
            // For non-numeric values, just use the next value when progress > 0.5
            if progress > 0.5 {
                next.clone()
            } else {
                prev.clone()
            }
        }
    }
}

impl WASMError {
    pub fn new(code: &str, message: &str) -> Self {
        Self {
            code: code.to_string(),
            message: message.to_string(),
            details: None,
        }
    }
}

// Chart Renderer Implementation
impl ChartRenderer {
    pub fn new() -> Self {
        Self {
            charts: HashMap::new(),
            render_cache: HashMap::new(),
            performance_stats: ChartPerformanceStats {
                total_charts: 0,
                total_render_time: 0.0,
                average_render_time: 0.0,
                cache_hit_rate: 0.0,
                memory_usage: 0,
            },
        }
    }

    pub fn create_chart(&mut self, chart_type: ChartType, data_source_id: String, config: ChartConfig) -> Result<String, WASMError> {
        let chart_id = format!("chart_{}", get_current_timestamp() as u64);
        
        let chart = Chart {
            id: chart_id.clone(),
            chart_type,
            data_source_id,
            config,
            axes: ChartAxes {
                x_axis: Some(ChartAxis::default()),
                y_axis: Some(ChartAxis::default()),
                secondary_y_axis: None,
            },
            series: Vec::new(),
            styling: ChartStyling::default(),
            interactions: ChartInteractions::default(),
            animations: ChartAnimations::default(),
        };

        self.charts.insert(chart_id.clone(), chart);
        self.performance_stats.total_charts += 1;

        Ok(chart_id)
    }

    pub fn add_series(&mut self, chart_id: &str, series: ChartSeries) -> Result<(), WASMError> {
        let chart = self.charts.get_mut(chart_id)
            .ok_or_else(|| WASMError::new("CHART_NOT_FOUND", "Chart not found"))?;
        
        chart.series.push(series);
        
        // Invalidate cache for this chart
        self.render_cache.remove(chart_id);
        
        Ok(())
    }

    pub fn update_chart_data(&mut self, chart_id: &str, data: &serde_json::Value) -> Result<(), WASMError> {
        let chart = self.charts.get(chart_id)
            .ok_or_else(|| WASMError::new("CHART_NOT_FOUND", "Chart not found"))?;
        
        // Invalidate cache for this chart
        self.render_cache.remove(chart_id);
        
        // Update performance stats
        self.performance_stats.cache_hit_rate = self.calculate_cache_hit_rate();
        
        Ok(())
    }

    pub fn render_chart(&mut self, chart_id: &str, data: &serde_json::Value) -> Result<RenderedChart, WASMError> {
        let start_time = get_current_timestamp();
        
        // Check cache first
        if let Some(cached) = self.render_cache.get(chart_id) {
            return Ok(cached.clone());
        }

        let chart = self.charts.get(chart_id)
            .ok_or_else(|| WASMError::new("CHART_NOT_FOUND", "Chart not found"))?;

        let rendered_chart = match chart.chart_type {
            ChartType::Line => self.render_line_chart(chart, data)?,
            ChartType::Bar => self.render_bar_chart(chart, data)?,
            ChartType::Pie => self.render_pie_chart(chart, data)?,
            ChartType::Scatter => self.render_scatter_chart(chart, data)?,
            ChartType::Area => self.render_area_chart(chart, data)?,
            ChartType::Histogram => self.render_histogram_chart(chart, data)?,
            ChartType::Heatmap => self.render_heatmap_chart(chart, data)?,
            ChartType::Treemap => self.render_treemap_chart(chart, data)?,
            ChartType::Sankey => self.render_sankey_chart(chart, data)?,
            ChartType::Radar => self.render_radar_chart(chart, data)?,
            ChartType::Gauge => self.render_gauge_chart(chart, data)?,
            ChartType::Candlestick => self.render_candlestick_chart(chart, data)?,
        };

        let render_time = get_current_timestamp() - start_time;
        
        // Update performance stats
        self.performance_stats.total_render_time += render_time;
        self.performance_stats.average_render_time = 
            self.performance_stats.total_render_time / self.performance_stats.total_charts as f64;

        // Cache the result
        self.render_cache.insert(chart_id.to_string(), rendered_chart.clone());

        Ok(rendered_chart)
    }

    fn render_line_chart(&self, chart: &Chart, data: &serde_json::Value) -> Result<RenderedChart, WASMError> {
        let mut svg_content = String::new();
        let mut data_points = Vec::new();

        // Extract data points
        if let Some(data_array) = data.as_array() {
            for (i, item) in data_array.iter().enumerate() {
                for series in &chart.series {
                    if let Some(value) = item.get(&series.data_field) {
                        let x = i as f64;
                        let y = value.as_f64().unwrap_or(0.0);
                        
                        data_points.push(DataPoint {
                            x,
                            y,
                            value: value.clone(),
                            series_id: series.id.clone(),
                            label: item.get("label").and_then(|v| v.as_str()).map(|s| s.to_string()),
                            color: series.color.clone(),
                        });
                    }
                }
            }
        }

        // Generate SVG for line chart
        svg_content.push_str(&format!(
            r#"<svg width="{}" height="{}" viewBox="0 0 {} {}" xmlns="http://www.w3.org/2000/svg">"#,
            chart.config.width, chart.config.height, chart.config.width, chart.config.height
        ));

        // Add background
        if let Some(bg_color) = &chart.config.background_color {
            svg_content.push_str(&format!(
                r#"<rect width="100%" height="100%" fill="{}"/>"#,
                bg_color
            ));
        }

        // Draw axes
        self.draw_axes(&mut svg_content, chart);

        // Draw data series
        for series in &chart.series {
            if series.visible {
                self.draw_line_series(&mut svg_content, chart, series, &data_points);
            }
        }

        // Add title
        if let Some(title) = &chart.config.title {
            svg_content.push_str(&format!(
                r#"<text x="{}" y="30" text-anchor="middle" font-size="{}" font-family="{}" fill="{}">{}</text>"#,
                chart.config.width / 2.0,
                title.font_size,
                title.font_family,
                title.color,
                title.text
            ));
        }

        svg_content.push_str("</svg>");

        Ok(RenderedChart {
            chart_id: chart.id.clone(),
            svg_content,
            bounds: BoundingBox {
                x: 0.0,
                y: 0.0,
                width: chart.config.width,
                height: chart.config.height,
            },
            data_points,
            render_time: get_current_timestamp(),
            last_updated: get_current_timestamp(),
        })
    }

    fn render_bar_chart(&self, chart: &Chart, data: &serde_json::Value) -> Result<RenderedChart, WASMError> {
        let mut svg_content = String::new();
        let mut data_points = Vec::new();

        // Extract data points
        if let Some(data_array) = data.as_array() {
            for (i, item) in data_array.iter().enumerate() {
                for series in &chart.series {
                    if let Some(value) = item.get(&series.data_field) {
                        let x = i as f64;
                        let y = value.as_f64().unwrap_or(0.0);
                        
                        data_points.push(DataPoint {
                            x,
                            y,
                            value: value.clone(),
                            series_id: series.id.clone(),
                            label: item.get("label").and_then(|v| v.as_str()).map(|s| s.to_string()),
                            color: series.color.clone(),
                        });
                    }
                }
            }
        }

        // Generate SVG for bar chart
        svg_content.push_str(&format!(
            r#"<svg width="{}" height="{}" viewBox="0 0 {} {}" xmlns="http://www.w3.org/2000/svg">"#,
            chart.config.width, chart.config.height, chart.config.width, chart.config.height
        ));

        // Add background
        if let Some(bg_color) = &chart.config.background_color {
            svg_content.push_str(&format!(
                r#"<rect width="100%" height="100%" fill="{}"/>"#,
                bg_color
            ));
        }

        // Draw axes
        self.draw_axes(&mut svg_content, chart);

        // Draw bars
        let bar_width = (chart.config.width - chart.config.margin.left - chart.config.margin.right) / data_points.len() as f64 * 0.8;
        
        for (i, point) in data_points.iter().enumerate() {
            let x = chart.config.margin.left + (i as f64 * (chart.config.width - chart.config.margin.left - chart.config.margin.right) / data_points.len() as f64);
            let height = point.y * (chart.config.height - chart.config.margin.top - chart.config.margin.bottom) / 100.0; // Assuming max value of 100
            let y = chart.config.height - chart.config.margin.bottom - height;

            svg_content.push_str(&format!(
                r#"<rect x="{}" y="{}" width="{}" height="{}" fill="{}"/>"#,
                x, y, bar_width, height, point.color
            ));
        }

        svg_content.push_str("</svg>");

        Ok(RenderedChart {
            chart_id: chart.id.clone(),
            svg_content,
            bounds: BoundingBox {
                x: 0.0,
                y: 0.0,
                width: chart.config.width,
                height: chart.config.height,
            },
            data_points,
            render_time: get_current_timestamp(),
            last_updated: get_current_timestamp(),
        })
    }

    fn render_pie_chart(&self, chart: &Chart, data: &serde_json::Value) -> Result<RenderedChart, WASMError> {
        let mut svg_content = String::new();
        let mut data_points = Vec::new();

        // Extract data points and calculate total
        let mut total = 0.0;
        if let Some(data_array) = data.as_array() {
            for (i, item) in data_array.iter().enumerate() {
                for series in &chart.series {
                    if let Some(value) = item.get(&series.data_field) {
                        let y = value.as_f64().unwrap_or(0.0);
                        total += y;
                        
                        data_points.push(DataPoint {
                            x: i as f64,
                            y,
                            value: value.clone(),
                            series_id: series.id.clone(),
                            label: item.get("label").and_then(|v| v.as_str()).map(|s| s.to_string()),
                            color: series.color.clone(),
                        });
                    }
                }
            }
        }

        // Generate SVG for pie chart
        svg_content.push_str(&format!(
            r#"<svg width="{}" height="{}" viewBox="0 0 {} {}" xmlns="http://www.w3.org/2000/svg">"#,
            chart.config.width, chart.config.height, chart.config.width, chart.config.height
        ));

        let center_x = chart.config.width / 2.0;
        let center_y = chart.config.height / 2.0;
        let radius = (chart.config.width.min(chart.config.height) / 2.0) * 0.8;

        let mut current_angle = 0.0;
        
        for point in &data_points {
            let slice_angle = (point.y / total) * 2.0 * std::f64::consts::PI;
            let end_angle = current_angle + slice_angle;

            let x1 = center_x + radius * current_angle.cos();
            let y1 = center_y + radius * current_angle.sin();
            let x2 = center_x + radius * end_angle.cos();
            let y2 = center_y + radius * end_angle.sin();

            let large_arc = if slice_angle > std::f64::consts::PI { 1 } else { 0 };

            svg_content.push_str(&format!(
                r#"<path d="M {} {} L {} {} A {} {} 0 {} 1 {} {} Z" fill="{}"/>"#,
                center_x, center_y, x1, y1, radius, radius, large_arc, x2, y2, point.color
            ));

            current_angle = end_angle;
        }

        svg_content.push_str("</svg>");

        Ok(RenderedChart {
            chart_id: chart.id.clone(),
            svg_content,
            bounds: BoundingBox {
                x: 0.0,
                y: 0.0,
                width: chart.config.width,
                height: chart.config.height,
            },
            data_points,
            render_time: get_current_timestamp(),
            last_updated: get_current_timestamp(),
        })
    }

    // Placeholder implementations for other chart types
    fn render_scatter_chart(&self, chart: &Chart, data: &serde_json::Value) -> Result<RenderedChart, WASMError> {
        let mut svg_content = String::new();
        let mut data_points = Vec::new();

        // Extract data points
        if let Some(data_array) = data.as_array() {
            for (i, item) in data_array.iter().enumerate() {
                for series in &chart.series {
                    if let Some(x_value) = item.get("x").and_then(|v| v.as_f64()) {
                        if let Some(y_value) = item.get(&series.data_field).and_then(|v| v.as_f64()) {
                            data_points.push(DataPoint {
                                x: x_value,
                                y: y_value,
                                value: serde_json::json!({"x": x_value, "y": y_value}),
                                series_id: series.id.clone(),
                                label: item.get("label").and_then(|v| v.as_str()).map(|s| s.to_string()),
                                color: series.color.clone(),
                            });
                        }
                    }
                }
            }
        }

        // Generate SVG for scatter chart
        svg_content.push_str(&format!(
            r#"<svg width="{}" height="{}" viewBox="0 0 {} {}" xmlns="http://www.w3.org/2000/svg">"#,
            chart.config.width, chart.config.height, chart.config.width, chart.config.height
        ));

        // Add background
        if let Some(bg_color) = &chart.config.background_color {
            svg_content.push_str(&format!(
                r#"<rect width="100%" height="100%" fill="{}"/>"#,
                bg_color
            ));
        }

        // Draw axes
        self.draw_axes(&mut svg_content, chart);

        // Draw scatter points
        for point in &data_points {
            let x = chart.config.margin.left + (point.x * (chart.config.width - chart.config.margin.left - chart.config.margin.right) / 100.0);
            let y = chart.config.height - chart.config.margin.bottom - (point.y * (chart.config.height - chart.config.margin.top - chart.config.margin.bottom) / 100.0);
            
            let marker_size = chart.series.iter()
                .find(|s| s.id == point.series_id)
                .and_then(|s| s.marker_size)
                .unwrap_or(4.0);

            svg_content.push_str(&format!(
                r#"<circle cx="{}" cy="{}" r="{}" fill="{}" opacity="0.7"/>"#,
                x, y, marker_size, point.color
            ));
        }

        svg_content.push_str("</svg>");

        Ok(RenderedChart {
            chart_id: chart.id.clone(),
            svg_content,
            bounds: BoundingBox {
                x: 0.0,
                y: 0.0,
                width: chart.config.width,
                height: chart.config.height,
            },
            data_points,
            render_time: get_current_timestamp(),
            last_updated: get_current_timestamp(),
        })
    }

    fn render_area_chart(&self, chart: &Chart, data: &serde_json::Value) -> Result<RenderedChart, WASMError> {
        let mut svg_content = String::new();
        let mut data_points = Vec::new();

        // Extract data points
        if let Some(data_array) = data.as_array() {
            for (i, item) in data_array.iter().enumerate() {
                for series in &chart.series {
                    if let Some(value) = item.get(&series.data_field) {
                        let x = i as f64;
                        let y = value.as_f64().unwrap_or(0.0);
                        
                        data_points.push(DataPoint {
                            x,
                            y,
                            value: value.clone(),
                            series_id: series.id.clone(),
                            label: item.get("label").and_then(|v| v.as_str()).map(|s| s.to_string()),
                            color: series.color.clone(),
                        });
                    }
                }
            }
        }

        // Generate SVG for area chart
        svg_content.push_str(&format!(
            r#"<svg width="{}" height="{}" viewBox="0 0 {} {}" xmlns="http://www.w3.org/2000/svg">"#,
            chart.config.width, chart.config.height, chart.config.width, chart.config.height
        ));

        // Add background
        if let Some(bg_color) = &chart.config.background_color {
            svg_content.push_str(&format!(
                r#"<rect width="100%" height="100%" fill="{}"/>"#,
                bg_color
            ));
        }

        // Draw axes
        self.draw_axes(&mut svg_content, chart);

        // Draw area for each series
        for series in &chart.series {
            if series.visible {
                let series_points: Vec<&DataPoint> = data_points.iter()
                    .filter(|p| p.series_id == series.id)
                    .collect();

                if !series_points.is_empty() {
                    let mut path_data = String::new();
                    let baseline_y = chart.config.height - chart.config.margin.bottom;
                    
                    // Start from baseline
                    let first_x = chart.config.margin.left + (series_points[0].x * (chart.config.width - chart.config.margin.left - chart.config.margin.right) / series_points.len() as f64);
                    path_data.push_str(&format!("M {} {}", first_x, baseline_y));
                    
                    // Draw line to first point
                    let first_y = chart.config.height - chart.config.margin.bottom - (series_points[0].y * (chart.config.height - chart.config.margin.top - chart.config.margin.bottom) / 100.0);
                    path_data.push_str(&format!(" L {} {}", first_x, first_y));
                    
                    // Draw through all points
                    for (i, point) in series_points.iter().enumerate().skip(1) {
                        let x = chart.config.margin.left + (point.x * (chart.config.width - chart.config.margin.left - chart.config.margin.right) / series_points.len() as f64);
                        let y = chart.config.height - chart.config.margin.bottom - (point.y * (chart.config.height - chart.config.margin.top - chart.config.margin.bottom) / 100.0);
                        path_data.push_str(&format!(" L {} {}", x, y));
                    }
                    
                    // Close to baseline
                    let last_x = chart.config.margin.left + (series_points.last().unwrap().x * (chart.config.width - chart.config.margin.left - chart.config.margin.right) / series_points.len() as f64);
                    path_data.push_str(&format!(" L {} {} Z", last_x, baseline_y));

                    let fill_opacity = series.fill_opacity.unwrap_or(0.3);
                    svg_content.push_str(&format!(
                        r#"<path d="{}" fill="{}" fill-opacity="{}" stroke="{}" stroke-width="{}"/>"#,
                        path_data,
                        series.color,
                        fill_opacity,
                        series.color,
                        series.line_width.unwrap_or(2.0)
                    ));
                }
            }
        }

        svg_content.push_str("</svg>");

        Ok(RenderedChart {
            chart_id: chart.id.clone(),
            svg_content,
            bounds: BoundingBox {
                x: 0.0,
                y: 0.0,
                width: chart.config.width,
                height: chart.config.height,
            },
            data_points,
            render_time: get_current_timestamp(),
            last_updated: get_current_timestamp(),
        })
    }

    fn render_histogram_chart(&self, chart: &Chart, data: &serde_json::Value) -> Result<RenderedChart, WASMError> {
        let mut svg_content = String::new();
        let mut data_points = Vec::new();

        // Extract and bin data for histogram
        let mut values = Vec::new();
        if let Some(data_array) = data.as_array() {
            for item in data_array {
                for series in &chart.series {
                    if let Some(value) = item.get(&series.data_field).and_then(|v| v.as_f64()) {
                        values.push(value);
                    }
                }
            }
        }

        if values.is_empty() {
            return Err(WASMError::new("NO_DATA", "No data available for histogram"));
        }

        // Calculate bins
        let bin_count = 10; // Default bin count
        let min_val = values.iter().fold(f64::INFINITY, |a, &b| a.min(b));
        let max_val = values.iter().fold(f64::NEG_INFINITY, |a, &b| a.max(b));
        let bin_width = (max_val - min_val) / bin_count as f64;

        let mut bins = vec![0; bin_count];
        for value in &values {
            let bin_index = ((value - min_val) / bin_width).floor() as usize;
            let bin_index = bin_index.min(bin_count - 1);
            bins[bin_index] += 1;
        }

        // Convert bins to data points
        for (i, &count) in bins.iter().enumerate() {
            let x = i as f64;
            let y = count as f64;
            data_points.push(DataPoint {
                x,
                y,
                value: serde_json::json!({"bin_start": min_val + i as f64 * bin_width, "count": count}),
                series_id: "histogram".to_string(),
                label: Some(format!("{:.1}-{:.1}", min_val + i as f64 * bin_width, min_val + (i + 1) as f64 * bin_width)),
                color: chart.series.first().map(|s| s.color.clone()).unwrap_or_else(|| "#1f77b4".to_string()),
            });
        }

        // Generate SVG
        svg_content.push_str(&format!(
            r#"<svg width="{}" height="{}" viewBox="0 0 {} {}" xmlns="http://www.w3.org/2000/svg">"#,
            chart.config.width, chart.config.height, chart.config.width, chart.config.height
        ));

        // Add background
        if let Some(bg_color) = &chart.config.background_color {
            svg_content.push_str(&format!(
                r#"<rect width="100%" height="100%" fill="{}"/>"#,
                bg_color
            ));
        }

        // Draw axes
        self.draw_axes(&mut svg_content, chart);

        // Draw histogram bars
        let bar_width = (chart.config.width - chart.config.margin.left - chart.config.margin.right) / bin_count as f64 * 0.9;
        let max_count = bins.iter().max().unwrap_or(&1);
        
        for (i, point) in data_points.iter().enumerate() {
            let x = chart.config.margin.left + (i as f64 * (chart.config.width - chart.config.margin.left - chart.config.margin.right) / bin_count as f64);
            let height = point.y * (chart.config.height - chart.config.margin.top - chart.config.margin.bottom) / *max_count as f64;
            let y = chart.config.height - chart.config.margin.bottom - height;

            svg_content.push_str(&format!(
                r#"<rect x="{}" y="{}" width="{}" height="{}" fill="{}" stroke="#333" stroke-width="1"/>"#,
                x, y, bar_width, height, point.color
            ));
        }

        svg_content.push_str("</svg>");

        Ok(RenderedChart {
            chart_id: chart.id.clone(),
            svg_content,
            bounds: BoundingBox {
                x: 0.0,
                y: 0.0,
                width: chart.config.width,
                height: chart.config.height,
            },
            data_points,
            render_time: get_current_timestamp(),
            last_updated: get_current_timestamp(),
        })
    }

    fn render_heatmap_chart(&self, chart: &Chart, data: &serde_json::Value) -> Result<RenderedChart, WASMError> {
        let mut svg_content = String::new();
        let mut data_points = Vec::new();

        // Extract 2D grid data for heatmap
        if let Some(data_array) = data.as_array() {
            for (row, row_data) in data_array.iter().enumerate() {
                if let Some(row_array) = row_data.as_array() {
                    for (col, cell_value) in row_array.iter().enumerate() {
                        if let Some(value) = cell_value.as_f64() {
                            data_points.push(DataPoint {
                                x: col as f64,
                                y: row as f64,
                                value: serde_json::json!(value),
                                series_id: "heatmap".to_string(),
                                label: Some(format!("({}, {}): {}", col, row, value)),
                                color: self.value_to_color(value, 0.0, 100.0), // Assuming 0-100 range
                            });
                        }
                    }
                }
            }
        }

        if data_points.is_empty() {
            return Err(WASMError::new("NO_DATA", "No data available for heatmap"));
        }

        // Calculate grid dimensions
        let max_x = data_points.iter().map(|p| p.x as usize).max().unwrap_or(0) + 1;
        let max_y = data_points.iter().map(|p| p.y as usize).max().unwrap_or(0) + 1;

        // Generate SVG
        svg_content.push_str(&format!(
            r#"<svg width="{}" height="{}" viewBox="0 0 {} {}" xmlns="http://www.w3.org/2000/svg">"#,
            chart.config.width, chart.config.height, chart.config.width, chart.config.height
        ));

        // Add background
        if let Some(bg_color) = &chart.config.background_color {
            svg_content.push_str(&format!(
                r#"<rect width="100%" height="100%" fill="{}"/>"#,
                bg_color
            ));
        }

        // Draw heatmap cells
        let cell_width = (chart.config.width - chart.config.margin.left - chart.config.margin.right) / max_x as f64;
        let cell_height = (chart.config.height - chart.config.margin.top - chart.config.margin.bottom) / max_y as f64;

        for point in &data_points {
            let x = chart.config.margin.left + (point.x * cell_width);
            let y = chart.config.margin.top + (point.y * cell_height);

            svg_content.push_str(&format!(
                r#"<rect x="{}" y="{}" width="{}" height="{}" fill="{}" stroke="#fff" stroke-width="1"/>"#,
                x, y, cell_width, cell_height, point.color
            ));
        }

        svg_content.push_str("</svg>");

        Ok(RenderedChart {
            chart_id: chart.id.clone(),
            svg_content,
            bounds: BoundingBox {
                x: 0.0,
                y: 0.0,
                width: chart.config.width,
                height: chart.config.height,
            },
            data_points,
            render_time: get_current_timestamp(),
            last_updated: get_current_timestamp(),
        })
    }

    fn render_treemap_chart(&self, chart: &Chart, data: &serde_json::Value) -> Result<RenderedChart, WASMError> {
        // Simplified treemap implementation
        self.render_bar_chart(chart, data) // For now, use bar chart as fallback
    }

    fn render_sankey_chart(&self, chart: &Chart, data: &serde_json::Value) -> Result<RenderedChart, WASMError> {
        // Simplified sankey implementation
        self.render_line_chart(chart, data) // For now, use line chart as fallback
    }

    fn render_radar_chart(&self, chart: &Chart, data: &serde_json::Value) -> Result<RenderedChart, WASMError> {
        let mut svg_content = String::new();
        let mut data_points = Vec::new();

        // Extract data points for radar chart
        if let Some(data_array) = data.as_array() {
            for item in data_array {
                for series in &chart.series {
                    if let Some(value) = item.get(&series.data_field).and_then(|v| v.as_f64()) {
                        data_points.push(DataPoint {
                            x: 0.0, // Will be calculated based on angle
                            y: value,
                            value: serde_json::json!(value),
                            series_id: series.id.clone(),
                            label: item.get("label").and_then(|v| v.as_str()).map(|s| s.to_string()),
                            color: series.color.clone(),
                        });
                    }
                }
            }
        }

        // Generate SVG for radar chart
        svg_content.push_str(&format!(
            r#"<svg width="{}" height="{}" viewBox="0 0 {} {}" xmlns="http://www.w3.org/2000/svg">"#,
            chart.config.width, chart.config.height, chart.config.width, chart.config.height
        ));

        let center_x = chart.config.width / 2.0;
        let center_y = chart.config.height / 2.0;
        let radius = (chart.config.width.min(chart.config.height) / 2.0) * 0.8;

        // Draw radar grid (concentric circles and radial lines)
        let levels = 5;
        for level in 1..=levels {
            let level_radius = radius * (level as f64) / (levels as f64);
            svg_content.push_str(&format!(
                r#"<circle cx="{}" cy="{}" r="{}" fill="none" stroke="#e0e0e0" stroke-width="1"/>"#,
                center_x, center_y, level_radius
            ));
        }

        // Draw radial lines
        let num_axes = data_points.len().max(3);
        for i in 0..num_axes {
            let angle = (i as f64) * 2.0 * std::f64::consts::PI / (num_axes as f64) - std::f64::consts::PI / 2.0;
            let end_x = center_x + radius * angle.cos();
            let end_y = center_y + radius * angle.sin();
            
            svg_content.push_str(&format!(
                r#"<line x1="{}" y1="{}" x2="{}" y2="{}" stroke="#e0e0e0" stroke-width="1"/>"#,
                center_x, center_y, end_x, end_y
            ));
        }

        // Draw data polygon
        if !data_points.is_empty() {
            let mut path_data = String::new();
            
            for (i, point) in data_points.iter().enumerate() {
                let angle = (i as f64) * 2.0 * std::f64::consts::PI / (data_points.len() as f64) - std::f64::consts::PI / 2.0;
                let point_radius = (point.y / 100.0) * radius; // Assuming 0-100 scale
                let x = center_x + point_radius * angle.cos();
                let y = center_y + point_radius * angle.sin();
                
                if i == 0 {
                    path_data.push_str(&format!("M {} {}", x, y));
                } else {
                    path_data.push_str(&format!(" L {} {}", x, y));
                }
            }
            path_data.push_str(" Z");

            let series_color = data_points.first().map(|p| &p.color).unwrap_or(&"#1f77b4".to_string());
            svg_content.push_str(&format!(
                r#"<path d="{}" fill="{}" fill-opacity="0.3" stroke="{}" stroke-width="2"/>"#,
                path_data, series_color, series_color
            ));
        }

        svg_content.push_str("</svg>");

        Ok(RenderedChart {
            chart_id: chart.id.clone(),
            svg_content,
            bounds: BoundingBox {
                x: 0.0,
                y: 0.0,
                width: chart.config.width,
                height: chart.config.height,
            },
            data_points,
            render_time: get_current_timestamp(),
            last_updated: get_current_timestamp(),
        })
    }

    fn render_gauge_chart(&self, chart: &Chart, data: &serde_json::Value) -> Result<RenderedChart, WASMError> {
        let mut svg_content = String::new();
        let mut data_points = Vec::new();

        // Extract single value for gauge
        let mut gauge_value = 0.0;
        if let Some(data_array) = data.as_array() {
            if let Some(first_item) = data_array.first() {
                for series in &chart.series {
                    if let Some(value) = first_item.get(&series.data_field).and_then(|v| v.as_f64()) {
                        gauge_value = value;
                        data_points.push(DataPoint {
                            x: 0.0,
                            y: value,
                            value: serde_json::json!(value),
                            series_id: series.id.clone(),
                            label: Some(format!("{:.1}", value)),
                            color: series.color.clone(),
                        });
                        break;
                    }
                }
            }
        }

        // Generate SVG for gauge chart
        svg_content.push_str(&format!(
            r#"<svg width="{}" height="{}" viewBox="0 0 {} {}" xmlns="http://www.w3.org/2000/svg">"#,
            chart.config.width, chart.config.height, chart.config.width, chart.config.height
        ));

        let center_x = chart.config.width / 2.0;
        let center_y = chart.config.height * 0.8;
        let radius = (chart.config.width.min(chart.config.height) / 2.0) * 0.7;

        // Draw gauge background arc
        svg_content.push_str(&format!(
            r#"<path d="M {} {} A {} {} 0 0 1 {} {}" fill="none" stroke="#e0e0e0" stroke-width="20"/>"#,
            center_x - radius, center_y,
            radius, radius,
            center_x + radius, center_y
        ));

        // Draw gauge value arc
        let value_angle = (gauge_value / 100.0) * std::f64::consts::PI; // Assuming 0-100 scale
        let end_x = center_x + radius * (value_angle - std::f64::consts::PI).cos();
        let end_y = center_y + radius * (value_angle - std::f64::consts::PI).sin();
        
        let large_arc = if value_angle > std::f64::consts::PI / 2.0 { 1 } else { 0 };
        
        svg_content.push_str(&format!(
            r#"<path d="M {} {} A {} {} 0 {} 1 {} {}" fill="none" stroke="#4CAF50" stroke-width="20"/>"#,
            center_x - radius, center_y,
            radius, radius, large_arc,
            end_x, end_y
        ));

        // Draw gauge needle
        let needle_angle = value_angle - std::f64::consts::PI;
        let needle_end_x = center_x + (radius * 0.8) * needle_angle.cos();
        let needle_end_y = center_y + (radius * 0.8) * needle_angle.sin();
        
        svg_content.push_str(&format!(
            r#"<line x1="{}" y1="{}" x2="{}" y2="{}" stroke="#333" stroke-width="3"/>"#,
            center_x, center_y, needle_end_x, needle_end_y
        ));

        // Draw center circle
        svg_content.push_str(&format!(
            r#"<circle cx="{}" cy="{}" r="8" fill="#333"/>"#,
            center_x, center_y
        ));

        // Draw value text
        svg_content.push_str(&format!(
            r#"<text x="{}" y="{}" text-anchor="middle" font-size="24" font-family="Arial" fill="#333">{:.1}</text>"#,
            center_x, center_y + 40.0, gauge_value
        ));

        svg_content.push_str("</svg>");

        Ok(RenderedChart {
            chart_id: chart.id.clone(),
            svg_content,
            bounds: BoundingBox {
                x: 0.0,
                y: 0.0,
                width: chart.config.width,
                height: chart.config.height,
            },
            data_points,
            render_time: get_current_timestamp(),
            last_updated: get_current_timestamp(),
        })
    }

    fn render_candlestick_chart(&self, chart: &Chart, data: &serde_json::Value) -> Result<RenderedChart, WASMError> {
        let mut svg_content = String::new();
        let mut data_points = Vec::new();

        // Extract OHLC data for candlestick chart
        if let Some(data_array) = data.as_array() {
            for (i, item) in data_array.iter().enumerate() {
                let open = item.get("open").and_then(|v| v.as_f64()).unwrap_or(0.0);
                let high = item.get("high").and_then(|v| v.as_f64()).unwrap_or(0.0);
                let low = item.get("low").and_then(|v| v.as_f64()).unwrap_or(0.0);
                let close = item.get("close").and_then(|v| v.as_f64()).unwrap_or(0.0);
                
                data_points.push(DataPoint {
                    x: i as f64,
                    y: close,
                    value: serde_json::json!({"open": open, "high": high, "low": low, "close": close}),
                    series_id: "candlestick".to_string(),
                    label: item.get("label").and_then(|v| v.as_str()).map(|s| s.to_string()),
                    color: if close >= open { "#4CAF50".to_string() } else { "#F44336".to_string() },
                });
            }
        }

        // Generate SVG for candlestick chart
        svg_content.push_str(&format!(
            r#"<svg width="{}" height="{}" viewBox="0 0 {} {}" xmlns="http://www.w3.org/2000/svg">"#,
            chart.config.width, chart.config.height, chart.config.width, chart.config.height
        ));

        // Add background
        if let Some(bg_color) = &chart.config.background_color {
            svg_content.push_str(&format!(
                r#"<rect width="100%" height="100%" fill="{}"/>"#,
                bg_color
            ));
        }

        // Draw axes
        self.draw_axes(&mut svg_content, chart);

        // Draw candlesticks
        let candle_width = (chart.config.width - chart.config.margin.left - chart.config.margin.right) / data_points.len() as f64 * 0.6;
        
        for (i, point) in data_points.iter().enumerate() {
            if let Some(ohlc) = point.value.as_object() {
                let open = ohlc.get("open").and_then(|v| v.as_f64()).unwrap_or(0.0);
                let high = ohlc.get("high").and_then(|v| v.as_f64()).unwrap_or(0.0);
                let low = ohlc.get("low").and_then(|v| v.as_f64()).unwrap_or(0.0);
                let close = ohlc.get("close").and_then(|v| v.as_f64()).unwrap_or(0.0);
                
                let x = chart.config.margin.left + (i as f64 * (chart.config.width - chart.config.margin.left - chart.config.margin.right) / data_points.len() as f64);
                
                // Scale values to chart height (assuming reasonable price range)
                let scale_factor = (chart.config.height - chart.config.margin.top - chart.config.margin.bottom) / 100.0;
                let high_y = chart.config.height - chart.config.margin.bottom - (high * scale_factor);
                let low_y = chart.config.height - chart.config.margin.bottom - (low * scale_factor);
                let open_y = chart.config.height - chart.config.margin.bottom - (open * scale_factor);
                let close_y = chart.config.height - chart.config.margin.bottom - (close * scale_factor);
                
                // Draw high-low line
                svg_content.push_str(&format!(
                    r#"<line x1="{}" y1="{}" x2="{}" y2="{}" stroke="#333" stroke-width="1"/>"#,
                    x, high_y, x, low_y
                ));
                
                // Draw body rectangle
                let body_top = open_y.min(close_y);
                let body_height = (open_y - close_y).abs();
                
                svg_content.push_str(&format!(
                    r#"<rect x="{}" y="{}" width="{}" height="{}" fill="{}" stroke="#333" stroke-width="1"/>"#,
                    x - candle_width / 2.0, body_top, candle_width, body_height, point.color
                ));
            }
        }

        svg_content.push_str("</svg>");

        Ok(RenderedChart {
            chart_id: chart.id.clone(),
            svg_content,
            bounds: BoundingBox {
                x: 0.0,
                y: 0.0,
                width: chart.config.width,
                height: chart.config.height,
            },
            data_points,
            render_time: get_current_timestamp(),
            last_updated: get_current_timestamp(),
        })
    }

    fn draw_axes(&self, svg_content: &mut String, chart: &Chart) {
        // Draw X axis
        if let Some(x_axis) = &chart.axes.x_axis {
            let y = chart.config.height - chart.config.margin.bottom;
            svg_content.push_str(&format!(
                r#"<line x1="{}" y1="{}" x2="{}" y2="{}" stroke="{}" stroke-width="1"/>"#,
                chart.config.margin.left,
                y,
                chart.config.width - chart.config.margin.right,
                y,
                x_axis.color
            ));
        }

        // Draw Y axis
        if let Some(y_axis) = &chart.axes.y_axis {
            let x = chart.config.margin.left;
            svg_content.push_str(&format!(
                r#"<line x1="{}" y1="{}" x2="{}" y2="{}" stroke="{}" stroke-width="1"/>"#,
                x,
                chart.config.margin.top,
                x,
                chart.config.height - chart.config.margin.bottom,
                y_axis.color
            ));
        }
    }

    fn draw_line_series(&self, svg_content: &mut String, chart: &Chart, series: &ChartSeries, data_points: &[DataPoint]) {
        let series_points: Vec<&DataPoint> = data_points.iter()
            .filter(|p| p.series_id == series.id)
            .collect();

        if series_points.is_empty() {
            return;
        }

        let mut path_data = String::new();
        
        for (i, point) in series_points.iter().enumerate() {
            let x = chart.config.margin.left + (point.x * (chart.config.width - chart.config.margin.left - chart.config.margin.right) / series_points.len() as f64);
            let y = chart.config.height - chart.config.margin.bottom - (point.y * (chart.config.height - chart.config.margin.top - chart.config.margin.bottom) / 100.0);

            if i == 0 {
                path_data.push_str(&format!("M {} {}", x, y));
            } else {
                path_data.push_str(&format!(" L {} {}", x, y));
            }
        }

        svg_content.push_str(&format!(
            r#"<path d="{}" stroke="{}" stroke-width="{}" fill="none"/>"#,
            path_data,
            series.color,
            series.line_width.unwrap_or(2.0)
        ));
    }

    fn calculate_cache_hit_rate(&self) -> f64 {
        if self.charts.is_empty() {
            return 0.0;
        }
        
        let cached_count = self.render_cache.len();
        let total_count = self.charts.len();
        
        (cached_count as f64) / (total_count as f64) * 100.0
    }

    fn value_to_color(&self, value: f64, min_val: f64, max_val: f64) -> String {
        // Normalize value to 0-1 range
        let normalized = if max_val > min_val {
            ((value - min_val) / (max_val - min_val)).clamp(0.0, 1.0)
        } else {
            0.5
        };
        
        // Create color gradient from blue (cold) to red (hot)
        let red = (normalized * 255.0) as u8;
        let blue = ((1.0 - normalized) * 255.0) as u8;
        let green = 0u8;
        
        format!("rgb({}, {}, {})", red, green, blue)
    }

    pub fn update_chart_animation(&mut self, chart_id: &str, animation_progress: f64) -> Result<(), WASMError> {
        if let Some(chart) = self.charts.get_mut(chart_id) {
            // Update chart animation state
            // This could modify chart properties based on animation progress
            // For now, we'll just invalidate the cache to trigger re-render
            self.render_cache.remove(chart_id);
        }
        Ok(())
    }

    pub fn get_chart_data_bounds(&self, chart_id: &str) -> Result<(f64, f64, f64, f64), WASMError> {
        let chart = self.charts.get(chart_id)
            .ok_or_else(|| WASMError::new("CHART_NOT_FOUND", "Chart not found"))?;
        
        // Calculate data bounds (min_x, max_x, min_y, max_y)
        // This is useful for dynamic scaling and zoom operations
        Ok((0.0, 100.0, 0.0, 100.0)) // Default bounds
    }

    pub fn enable_chart_interactions(&mut self, chart_id: &str, interactions: ChartInteractions) -> Result<(), WASMError> {
        let chart = self.charts.get_mut(chart_id)
            .ok_or_else(|| WASMError::new("CHART_NOT_FOUND", "Chart not found"))?;
        
        chart.interactions = interactions;
        
        // Invalidate cache to reflect interaction changes
        self.render_cache.remove(chart_id);
        
        Ok(())
    }
}

impl DataSource {
    pub fn new(id: String, source_type: DataSourceType, data: serde_json::Value) -> Self {
        Self {
            id,
            source_type,
            data,
            update_frequency: None,
            last_updated: get_current_timestamp(),
        }
    }
    
    pub fn with_update_frequency(mut self, frequency: u32) -> Self {
        self.update_frequency = Some(frequency);
        self
    }

    pub fn update_data(&mut self, new_data: serde_json::Value) -> Result<(), WASMError> {
        // Validate data structure based on source type
        match self.source_type {
            DataSourceType::Static => {
                // Static data can be updated but won't auto-refresh
                self.data = new_data;
                self.last_updated = get_current_timestamp();
            }
            DataSourceType::Dynamic => {
                // Dynamic data supports real-time updates
                self.data = new_data;
                self.last_updated = get_current_timestamp();
            }
            DataSourceType::Stream => {
                // Stream data appends new values
                if let Some(existing_array) = self.data.as_array_mut() {
                    if let Some(new_array) = new_data.as_array() {
                        existing_array.extend(new_array.iter().cloned());
                        
                        // Limit stream size to prevent memory issues
                        const MAX_STREAM_SIZE: usize = 1000;
                        if existing_array.len() > MAX_STREAM_SIZE {
                            existing_array.drain(0..existing_array.len() - MAX_STREAM_SIZE);
                        }
                    }
                } else {
                    self.data = new_data;
                }
                self.last_updated = get_current_timestamp();
            }
            DataSourceType::Computed => {
                // Computed data is derived from other sources
                return Err(WASMError::new("COMPUTED_DATA_UPDATE", "Cannot directly update computed data source"));
            }
        }
        
        Ok(())
    }

    pub fn compute_from_sources(&mut self, sources: &HashMap<String, DataSource>, formula: &str) -> Result<(), WASMError> {
        if self.source_type != DataSourceType::Computed {
            return Err(WASMError::new("INVALID_OPERATION", "Can only compute data for computed data sources"));
        }
        
        // Simple computation examples - in a real implementation this would be more sophisticated
        match formula {
            "sum" => {
                let mut total = 0.0;
                for source in sources.values() {
                    if let Some(array) = source.data.as_array() {
                        for item in array {
                            if let Some(value) = item.as_f64() {
                                total += value;
                            }
                        }
                    }
                }
                self.data = serde_json::json!(total);
            }
            "average" => {
                let mut total = 0.0;
                let mut count = 0;
                for source in sources.values() {
                    if let Some(array) = source.data.as_array() {
                        for item in array {
                            if let Some(value) = item.as_f64() {
                                total += value;
                                count += 1;
                            }
                        }
                    }
                }
                let average = if count > 0 { total / count as f64 } else { 0.0 };
                self.data = serde_json::json!(average);
            }
            _ => {
                return Err(WASMError::new("UNKNOWN_FORMULA", "Unknown computation formula"));
            }
        }
        
        Ok(())
    }

    pub fn get_latest_values(&self, count: usize) -> Vec<serde_json::Value> {
        match &self.data {
            serde_json::Value::Array(arr) => {
                let start = if arr.len() > count { arr.len() - count } else { 0 };
                arr[start..].to_vec()
            }
            single_value => vec![single_value.clone()],
        }
    }

    pub fn get_data_statistics(&self) -> DataStatistics {
        let mut stats = DataStatistics::default();
        
        if let Some(array) = self.data.as_array() {
            let mut values = Vec::new();
            for item in array {
                if let Some(value) = item.as_f64() {
                    values.push(value);
                }
            }
            
            if !values.is_empty() {
                stats.count = values.len();
                stats.min = values.iter().fold(f64::INFINITY, |a, &b| a.min(b));
                stats.max = values.iter().fold(f64::NEG_INFINITY, |a, &b| a.max(b));
                stats.sum = values.iter().sum();
                stats.mean = stats.sum / values.len() as f64;
                
                // Calculate standard deviation
                let variance = values.iter()
                    .map(|&x| (x - stats.mean).powi(2))
                    .sum::<f64>() / values.len() as f64;
                stats.std_dev = variance.sqrt();
            }
        }
        
        stats
    }
}

#[derive(Serialize, Deserialize, Clone, Debug, Default)]
pub struct DataStatistics {
    pub count: usize,
    pub min: f64,
    pub max: f64,
    pub sum: f64,
    pub mean: f64,
    pub std_dev: f64,
}

// Enhanced data binding system
#[derive(Serialize, Deserialize, Clone, Debug)]
pub struct DataBinding {
    pub source_id: String,
    pub target_element: String,
    pub property_path: String,
    pub transform_function: Option<String>,
    pub update_trigger: UpdateTrigger,
}

#[derive(Serialize, Deserialize, Clone, Debug)]
pub enum UpdateTrigger {
    Immediate,
    Throttled(u32), // milliseconds
    OnChange,
    Manual,
}

pub struct DataBindingManager {
    bindings: HashMap<String, DataBinding>,
    last_update_times: HashMap<String, f64>,
}

impl DataBindingManager {
    pub fn new() -> Self {
        Self {
            bindings: HashMap::new(),
            last_update_times: HashMap::new(),
        }
    }

    pub fn add_binding(&mut self, binding: DataBinding) -> String {
        let binding_id = format!("binding_{}", get_current_timestamp() as u64);
        self.bindings.insert(binding_id.clone(), binding);
        binding_id
    }

    pub fn remove_binding(&mut self, binding_id: &str) {
        self.bindings.remove(binding_id);
        self.last_update_times.remove(binding_id);
    }

    pub fn update_bindings(&mut self, document_state: &mut DocumentState, current_time: f64) -> Vec<ElementChange> {
        let mut changes = Vec::new();
        
        for (binding_id, binding) in &self.bindings {
            let should_update = match binding.update_trigger {
                UpdateTrigger::Immediate => true,
                UpdateTrigger::Throttled(interval) => {
                    let last_update = self.last_update_times.get(binding_id).unwrap_or(&0.0);
                    current_time - last_update >= interval as f64
                }
                UpdateTrigger::OnChange => {
                    // Would need to track data changes - simplified for now
                    true
                }
                UpdateTrigger::Manual => false,
            };
            
            if should_update {
                if let Some(data_source) = document_state.data_sources.get(&binding.source_id) {
                    let new_value = self.extract_value_from_data(&data_source.data, &binding.property_path);
                    
                    let transformed_value = if let Some(transform) = &binding.transform_function {
                        self.apply_transform(new_value, transform)
                    } else {
                        new_value
                    };
                    
                    changes.push(ElementChange::Update {
                        element_id: binding.target_element.clone(),
                        properties: [(binding.property_path.clone(), transformed_value)].into_iter().collect(),
                    });
                    
                    self.last_update_times.insert(binding_id.clone(), current_time);
                }
            }
        }
        
        changes
    }

    fn extract_value_from_data(&self, data: &serde_json::Value, path: &str) -> serde_json::Value {
        // Simple path extraction - in a real implementation this would be more robust
        let parts: Vec<&str> = path.split('.').collect();
        let mut current = data;
        
        for part in parts {
            if let Some(obj) = current.as_object() {
                if let Some(value) = obj.get(part) {
                    current = value;
                } else {
                    return serde_json::Value::Null;
                }
            } else if let Some(array) = current.as_array() {
                if let Ok(index) = part.parse::<usize>() {
                    if let Some(value) = array.get(index) {
                        current = value;
                    } else {
                        return serde_json::Value::Null;
                    }
                } else {
                    return serde_json::Value::Null;
                }
            } else {
                return serde_json::Value::Null;
            }
        }
        
        current.clone()
    }

    fn apply_transform(&self, value: serde_json::Value, transform: &str) -> serde_json::Value {
        match transform {
            "uppercase" => {
                if let Some(s) = value.as_str() {
                    serde_json::Value::String(s.to_uppercase())
                } else {
                    value
                }
            }
            "lowercase" => {
                if let Some(s) = value.as_str() {
                    serde_json::Value::String(s.to_lowercase())
                } else {
                    value
                }
            }
            "round" => {
                if let Some(n) = value.as_f64() {
                    serde_json::Value::Number(serde_json::Number::from(n.round() as i64))
                } else {
                    value
                }
            }
            "percentage" => {
                if let Some(n) = value.as_f64() {
                    serde_json::Value::String(format!("{:.1}%", n * 100.0))
                } else {
                    value
                }
            }
            _ => value, // Unknown transform, return original value
        }
    }
}

// Chart and Visualization Framework

#[derive(Serialize, Deserialize, Clone, Debug)]
pub struct ChartRenderer {
    pub charts: HashMap<String, Chart>,
    pub render_cache: HashMap<String, RenderedChart>,
    pub performance_stats: ChartPerformanceStats,
}

#[derive(Serialize, Deserialize, Clone, Debug)]
pub struct Chart {
    pub id: String,
    pub chart_type: ChartType,
    pub data_source_id: String,
    pub config: ChartConfig,
    pub axes: ChartAxes,
    pub series: Vec<ChartSeries>,
    pub styling: ChartStyling,
    pub interactions: ChartInteractions,
    pub animations: ChartAnimations,
}

#[derive(Serialize, Deserialize, Clone, Debug)]
pub enum ChartType {
    Line,
    Bar,
    Pie,
    Scatter,
    Area,
    Histogram,
    Heatmap,
    Treemap,
    Sankey,
    Radar,
    Gauge,
    Candlestick,
}

#[derive(Serialize, Deserialize, Clone, Debug)]
pub struct ChartConfig {
    pub width: f64,
    pub height: f64,
    pub margin: ChartMargin,
    pub responsive: bool,
    pub maintain_aspect_ratio: bool,
    pub background_color: Option<String>,
    pub title: Option<ChartTitle>,
    pub legend: Option<ChartLegend>,
    pub tooltip: Option<ChartTooltip>,
}

#[derive(Serialize, Deserialize, Clone, Debug)]
pub struct ChartMargin {
    pub top: f64,
    pub right: f64,
    pub bottom: f64,
    pub left: f64,
}

#[derive(Serialize, Deserialize, Clone, Debug)]
pub struct ChartTitle {
    pub text: String,
    pub font_size: f64,
    pub font_family: String,
    pub color: String,
    pub alignment: TextAlignment,
}

#[derive(Serialize, Deserialize, Clone, Debug)]
pub struct ChartLegend {
    pub position: LegendPosition,
    pub show: bool,
    pub font_size: f64,
    pub color: String,
}

#[derive(Serialize, Deserialize, Clone, Debug)]
pub enum LegendPosition {
    Top,
    Bottom,
    Left,
    Right,
    TopLeft,
    TopRight,
    BottomLeft,
    BottomRight,
}

#[derive(Serialize, Deserialize, Clone, Debug)]
pub enum TextAlignment {
    Left,
    Center,
    Right,
}

#[derive(Serialize, Deserialize, Clone, Debug)]
pub struct ChartTooltip {
    pub enabled: bool,
    pub background_color: String,
    pub text_color: String,
    pub border_color: String,
    pub border_width: f64,
    pub font_size: f64,
    pub padding: f64,
}

#[derive(Serialize, Deserialize, Clone, Debug)]
pub struct ChartAxes {
    pub x_axis: Option<ChartAxis>,
    pub y_axis: Option<ChartAxis>,
    pub secondary_y_axis: Option<ChartAxis>,
}

#[derive(Serialize, Deserialize, Clone, Debug)]
pub struct ChartAxis {
    pub label: Option<String>,
    pub show_grid: bool,
    pub show_ticks: bool,
    pub tick_count: Option<u32>,
    pub min_value: Option<f64>,
    pub max_value: Option<f64>,
    pub scale_type: ScaleType,
    pub format: Option<String>,
    pub color: String,
    pub font_size: f64,
}

#[derive(Serialize, Deserialize, Clone, Debug)]
pub enum ScaleType {
    Linear,
    Logarithmic,
    Time,
    Category,
}

#[derive(Serialize, Deserialize, Clone, Debug)]
pub struct ChartSeries {
    pub id: String,
    pub name: String,
    pub data_field: String,
    pub color: String,
    pub line_width: Option<f64>,
    pub fill_opacity: Option<f64>,
    pub marker_size: Option<f64>,
    pub marker_shape: Option<MarkerShape>,
    pub visible: bool,
    pub y_axis: AxisReference,
}

#[derive(Serialize, Deserialize, Clone, Debug)]
pub enum MarkerShape {
    Circle,
    Square,
    Triangle,
    Diamond,
    Cross,
    Plus,
}

#[derive(Serialize, Deserialize, Clone, Debug)]
pub enum AxisReference {
    Primary,
    Secondary,
}

#[derive(Serialize, Deserialize, Clone, Debug)]
pub struct ChartStyling {
    pub color_palette: Vec<String>,
    pub gradient_fills: bool,
    pub drop_shadow: bool,
    pub border_radius: f64,
    pub grid_color: String,
    pub grid_opacity: f64,
}

#[derive(Serialize, Deserialize, Clone, Debug)]
pub struct ChartInteractions {
    pub zoom_enabled: bool,
    pub pan_enabled: bool,
    pub hover_effects: bool,
    pub click_events: bool,
    pub brush_selection: bool,
    pub crosshair: bool,
}

#[derive(Serialize, Deserialize, Clone, Debug)]
pub struct ChartAnimations {
    pub enabled: bool,
    pub duration: f64,
    pub easing: EasingFunction,
    pub stagger_delay: f64,
    pub entrance_animation: AnimationType,
    pub update_animation: AnimationType,
}

#[derive(Serialize, Deserialize, Clone, Debug)]
pub struct RenderedChart {
    pub chart_id: String,
    pub svg_content: String,
    pub bounds: BoundingBox,
    pub data_points: Vec<DataPoint>,
    pub render_time: f64,
    pub last_updated: f64,
}

#[derive(Serialize, Deserialize, Clone, Debug)]
pub struct DataPoint {
    pub x: f64,
    pub y: f64,
    pub value: serde_json::Value,
    pub series_id: String,
    pub label: Option<String>,
    pub color: String,
}

#[derive(Serialize, Deserialize, Clone, Debug)]
pub struct ChartPerformanceStats {
    pub total_charts: u32,
    pub total_render_time: f64,
    pub average_render_time: f64,
    pub cache_hit_rate: f64,
    pub memory_usage: u64,
}

// Vector Graphics Engine

#[derive(Serialize, Deserialize, Clone, Debug)]
pub struct VectorEngine {
    pub shapes: HashMap<String, VectorShape>,
    pub paths: HashMap<String, VectorPath>,
    pub gradients: HashMap<String, Gradient>,
    pub patterns: HashMap<String, Pattern>,
    pub filters: HashMap<String, Filter>,
}

#[derive(Serialize, Deserialize, Clone, Debug)]
pub struct VectorShape {
    pub id: String,
    pub shape_type: ShapeType,
    pub position: Position,
    pub size: Size,
    pub fill: Fill,
    pub stroke: Stroke,
    pub transform: Transform,
    pub opacity: f64,
}

#[derive(Serialize, Deserialize, Clone, Debug)]
pub enum ShapeType {
    Rectangle,
    Circle,
    Ellipse,
    Line,
    Polygon,
    Path,
    Text,
}

#[derive(Serialize, Deserialize, Clone, Debug)]
pub struct Fill {
    pub color: Option<String>,
    pub gradient_id: Option<String>,
    pub pattern_id: Option<String>,
    pub opacity: f64,
}

#[derive(Serialize, Deserialize, Clone, Debug)]
pub struct Stroke {
    pub color: String,
    pub width: f64,
    pub dash_array: Option<Vec<f64>>,
    pub line_cap: LineCap,
    pub line_join: LineJoin,
    pub opacity: f64,
}

#[derive(Serialize, Deserialize, Clone, Debug)]
pub enum LineCap {
    Butt,
    Round,
    Square,
}

#[derive(Serialize, Deserialize, Clone, Debug)]
pub enum LineJoin {
    Miter,
    Round,
    Bevel,
}

#[derive(Serialize, Deserialize, Clone, Debug)]
pub struct VectorPath {
    pub id: String,
    pub commands: Vec<PathCommand>,
    pub fill: Fill,
    pub stroke: Stroke,
    pub transform: Transform,
}

#[derive(Serialize, Deserialize, Clone, Debug)]
pub enum PathCommand {
    MoveTo { x: f64, y: f64 },
    LineTo { x: f64, y: f64 },
    CurveTo { x1: f64, y1: f64, x2: f64, y2: f64, x: f64, y: f64 },
    QuadraticCurveTo { x1: f64, y1: f64, x: f64, y: f64 },
    Arc { rx: f64, ry: f64, rotation: f64, large_arc: bool, sweep: bool, x: f64, y: f64 },
    ClosePath,
}

#[derive(Serialize, Deserialize, Clone, Debug)]
pub struct Gradient {
    pub id: String,
    pub gradient_type: GradientType,
    pub stops: Vec<GradientStop>,
    pub transform: Option<Transform>,
}

#[derive(Serialize, Deserialize, Clone, Debug)]
pub enum GradientType {
    Linear { x1: f64, y1: f64, x2: f64, y2: f64 },
    Radial { cx: f64, cy: f64, r: f64, fx: Option<f64>, fy: Option<f64> },
}

#[derive(Serialize, Deserialize, Clone, Debug)]
pub struct GradientStop {
    pub offset: f64,
    pub color: String,
    pub opacity: f64,
}

#[derive(Serialize, Deserialize, Clone, Debug)]
pub struct Pattern {
    pub id: String,
    pub width: f64,
    pub height: f64,
    pub content: String, // SVG content
    pub transform: Option<Transform>,
}

#[derive(Serialize, Deserialize, Clone, Debug)]
pub struct Filter {
    pub id: String,
    pub filter_type: FilterType,
    pub parameters: HashMap<String, f64>,
}

#[derive(Serialize, Deserialize, Clone, Debug)]
pub enum FilterType {
    Blur,
    DropShadow,
    Glow,
    Emboss,
    ColorMatrix,
    Brightness,
    Contrast,
    Saturation,
}

#[derive(Serialize, Deserialize, Clone, Debug)]
pub enum ComplexPathType {
    Bezier,
    Spiral,
    Star,
    Wave,
}

// Export types for JavaScript interop
#[wasm_bindgen]
extern "C" {
    #[wasm_bindgen(js_namespace = console)]
    fn log(s: &str);
}

// WASM-bindgen interface functions
#[wasm_bindgen]
pub fn init_interactive_engine(permissions_json: &str) -> Result<(), JsValue> {
    let permissions: WASMPermissions = serde_json::from_str(permissions_json)
        .map_err(|e| JsValue::from_str(&format!("Failed to parse permissions: {}", e)))?;
    
    let engine = InteractiveEngine::new(permissions)
        .map_err(|e| JsValue::from_str(&format!("Failed to create engine: {}", e.message)))?;
    
    let mut global_engine = ENGINE.lock().unwrap();
    *global_engine = Some(engine);
    
    log("LIV Interactive Engine initialized with security context");
    Ok(())
}

#[wasm_bindgen]
pub fn process_interaction(event_json: &str) -> Result<String, JsValue> {
    let event: InteractionEvent = serde_json::from_str(event_json)
        .map_err(|e| JsValue::from_str(&format!("Failed to parse event: {}", e)))?;
    
    let mut global_engine = ENGINE.lock().unwrap();
    if let Some(engine) = global_engine.as_mut() {
        let render_update = engine.process_interaction(event)
            .map_err(|e| JsValue::from_str(&format!("Interaction failed: {}", e.message)))?;
        
        serde_json::to_string(&render_update)
            .map_err(|e| JsValue::from_str(&format!("Failed to serialize update: {}", e)))
    } else {
        Err(JsValue::from_str("Engine not initialized"))
    }
}

#[wasm_bindgen]
pub fn render_frame(timestamp: f64) -> Result<String, JsValue> {
    let mut global_engine = ENGINE.lock().unwrap();
    if let Some(engine) = global_engine.as_mut() {
        let render_update = engine.render_frame(timestamp)
            .map_err(|e| JsValue::from_str(&format!("Render failed: {}", e.message)))?;
        
        serde_json::to_string(&render_update)
            .map_err(|e| JsValue::from_str(&format!("Failed to serialize update: {}", e)))
    } else {
        Err(JsValue::from_str("Engine not initialized"))
    }
}

#[wasm_bindgen]
pub fn update_data(data_source_id: &str, data: &[u8]) -> Result<(), JsValue> {
    let mut global_engine = ENGINE.lock().unwrap();
    if let Some(engine) = global_engine.as_mut() {
        engine.update_data(data_source_id, data)
            .map_err(|e| JsValue::from_str(&format!("Data update failed: {}", e.message)))?;
        Ok(())
    } else {
        Err(JsValue::from_str("Engine not initialized"))
    }
}

#[wasm_bindgen]
pub fn get_performance_stats() -> Result<String, JsValue> {
    let global_engine = ENGINE.lock().unwrap();
    if let Some(engine) = global_engine.as_ref() {
        let stats = engine.performance_monitor.get_stats();
        serde_json::to_string(&stats)
            .map_err(|e| JsValue::from_str(&format!("Failed to serialize stats: {}", e)))
    } else {
        Err(JsValue::from_str("Engine not initialized"))
    }
}

#[wasm_bindgen]
pub fn create_element(element_type: &str, properties_json: &str) -> Result<String, JsValue> {
    let mut global_engine = ENGINE.lock().unwrap();
    if let Some(engine) = global_engine.as_mut() {
        let element_type = match element_type {
            "chart" => ElementType::Chart,
            "animation" => ElementType::Animation,
            "interactive" => ElementType::Interactive,
            "vector" => ElementType::Vector,
            "text" => ElementType::Text,
            "image" => ElementType::Image,
            "container" => ElementType::Container,
            _ => return Err(JsValue::from_str("Invalid element type")),
        };
        
        let properties: HashMap<String, serde_json::Value> = serde_json::from_str(properties_json)
            .map_err(|e| JsValue::from_str(&format!("Failed to parse properties: {}", e)))?;
        
        engine.create_element(element_type, properties)
            .map_err(|e| JsValue::from_str(&format!("Failed to create element: {}", e.message)))
    } else {
        Err(JsValue::from_str("Engine not initialized"))
    }
}

#[wasm_bindgen]
pub fn update_element(element_id: &str, properties_json: &str) -> Result<(), JsValue> {
    let mut global_engine = ENGINE.lock().unwrap();
    if let Some(engine) = global_engine.as_mut() {
        let properties: HashMap<String, serde_json::Value> = serde_json::from_str(properties_json)
            .map_err(|e| JsValue::from_str(&format!("Failed to parse properties: {}", e)))?;
        
        engine.update_element_properties(element_id, properties)
            .map_err(|e| JsValue::from_str(&format!("Failed to update element: {}", e.message)))
    } else {
        Err(JsValue::from_str("Engine not initialized"))
    }
}

#[wasm_bindgen]
pub fn delete_element(element_id: &str) -> Result<(), JsValue> {
    let mut global_engine = ENGINE.lock().unwrap();
    if let Some(engine) = global_engine.as_mut() {
        engine.delete_element(element_id)
            .map_err(|e| JsValue::from_str(&format!("Failed to delete element: {}", e.message)))
    } else {
        Err(JsValue::from_str("Engine not initialized"))
    }
}

#[wasm_bindgen]
pub fn create_animation(target_element: &str, animation_type: &str, duration: f64, keyframes_json: &str) -> Result<String, JsValue> {
    let mut global_engine = ENGINE.lock().unwrap();
    if let Some(engine) = global_engine.as_mut() {
        let animation_type = match animation_type {
            "transform" => AnimationType::Transform,
            "style" => AnimationType::Style,
            "path" => AnimationType::Path,
            "morph" => AnimationType::Morph,
            _ => return Err(JsValue::from_str("Invalid animation type")),
        };
        
        let keyframes: Vec<Keyframe> = serde_json::from_str(keyframes_json)
            .map_err(|e| JsValue::from_str(&format!("Failed to parse keyframes: {}", e)))?;
        
        engine.create_animation(target_element, animation_type, duration, keyframes)
            .map_err(|e| JsValue::from_str(&format!("Failed to create animation: {}", e.message)))
    } else {
        Err(JsValue::from_str("Engine not initialized"))
    }
}

#[wasm_bindgen]
pub fn stop_animation(animation_id: &str) -> Result<(), JsValue> {
    let mut global_engine = ENGINE.lock().unwrap();
    if let Some(engine) = global_engine.as_mut() {
        engine.stop_animation(animation_id)
            .map_err(|e| JsValue::from_str(&format!("Failed to stop animation: {}", e.message)))
    } else {
        Err(JsValue::from_str("Engine not initialized"))
    }
}

#[wasm_bindgen]
pub fn add_event_handler(element_id: &str, event_type: &str, handler_id: &str) -> Result<(), JsValue> {
    let mut global_engine = ENGINE.lock().unwrap();
    if let Some(engine) = global_engine.as_mut() {
        engine.add_event_handler(element_id, event_type, handler_id)
            .map_err(|e| JsValue::from_str(&format!("Failed to add event handler: {}", e.message)))
    } else {
        Err(JsValue::from_str("Engine not initialized"))
    }
}

#[wasm_bindgen]
pub fn update_viewport(width: f64, height: f64, scale: f64) -> Result<(), JsValue> {
    let mut global_engine = ENGINE.lock().unwrap();
    if let Some(engine) = global_engine.as_mut() {
        engine.update_viewport(width, height, scale)
            .map_err(|e| JsValue::from_str(&format!("Failed to update viewport: {}", e.message)))
    } else {
        Err(JsValue::from_str("Engine not initialized"))
    }
}

#[wasm_bindgen]
pub fn get_element_bounds(element_id: &str) -> Result<String, JsValue> {
    let global_engine = ENGINE.lock().unwrap();
    if let Some(engine) = global_engine.as_ref() {
        let bounds = engine.get_element_bounds(element_id)
            .map_err(|e| JsValue::from_str(&format!("Failed to get element bounds: {}", e.message)))?;
        
        serde_json::to_string(&bounds)
            .map_err(|e| JsValue::from_str(&format!("Failed to serialize bounds: {}", e)))
    } else {
        Err(JsValue::from_str("Engine not initialized"))
    }
}

#[wasm_bindgen]
pub fn query_elements_by_type(element_type: &str) -> Result<String, JsValue> {
    let global_engine = ENGINE.lock().unwrap();
    if let Some(engine) = global_engine.as_ref() {
        let element_type = match element_type {
            "chart" => ElementType::Chart,
            "animation" => ElementType::Animation,
            "interactive" => ElementType::Interactive,
            "vector" => ElementType::Vector,
            "text" => ElementType::Text,
            "image" => ElementType::Image,
            "container" => ElementType::Container,
            _ => return Err(JsValue::from_str("Invalid element type")),
        };
        
        let element_ids = engine.query_elements_by_type(element_type);
        serde_json::to_string(&element_ids)
            .map_err(|e| JsValue::from_str(&format!("Failed to serialize element IDs: {}", e)))
    } else {
        Err(JsValue::from_str("Engine not initialized"))
    }
}

#[wasm_bindgen]
pub fn destroy_engine() {
    let mut global_engine = ENGINE.lock().unwrap();
    *global_engine = None;
    log("LIV Interactive Engine destroyed");
}

// Include tests module
#[cfg(test)]
mod tests;

#[cfg(test)]
mod chart_tests;

#[cfg(test)]
mod constraint_tests;

#[cfg(test)]
mod performance_tests;

#[cfg(test)]
mod interaction_tests;

#[cfg(test)]
mod memory_safety_tests;

#[cfg(test)]
mod integration_tests;
// V
ector Engine Implementation
impl VectorEngine {
    pub fn new() -> Self {
        Self {
            shapes: HashMap::new(),
            paths: HashMap::new(),
            gradients: HashMap::new(),
            patterns: HashMap::new(),
            filters: HashMap::new(),
        }
    }

    pub fn animate_shape(&mut self, shape_id: &str, target_transform: Transform, duration: f64) -> Result<String, WASMError> {
        let shape = self.shapes.get(shape_id)
            .ok_or_else(|| WASMError::new("SHAPE_NOT_FOUND", "Shape not found"))?;
        
        // Create animation for the shape
        let animation_id = format!("anim_{}", get_current_timestamp() as u64);
        
        // In a real implementation, this would create an animation timeline
        // For now, we'll directly update the shape transform
        if let Some(shape) = self.shapes.get_mut(shape_id) {
            shape.transform = target_transform;
        }
        
        Ok(animation_id)
    }

    pub fn morph_path(&mut self, path_id: &str, target_commands: Vec<PathCommand>, duration: f64) -> Result<String, WASMError> {
        let path = self.paths.get(path_id)
            .ok_or_else(|| WASMError::new("PATH_NOT_FOUND", "Path not found"))?;
        
        // Create morphing animation for the path
        let animation_id = format!("morph_{}", get_current_timestamp() as u64);
        
        // In a real implementation, this would interpolate between path commands
        // For now, we'll directly update the path
        if let Some(path) = self.paths.get_mut(path_id) {
            path.commands = target_commands;
        }
        
        Ok(animation_id)
    }

    pub fn create_animated_gradient(&mut self, gradient_type: GradientType, stops: Vec<GradientStop>, animation_duration: f64) -> Result<String, WASMError> {
        let gradient_id = self.create_gradient(gradient_type, stops)?;
        
        // Add animation properties to the gradient
        // This could animate color stops, positions, etc.
        
        Ok(gradient_id)
    }

    pub fn apply_filter_to_shape(&mut self, shape_id: &str, filter_id: &str) -> Result<(), WASMError> {
        let shape = self.shapes.get_mut(shape_id)
            .ok_or_else(|| WASMError::new("SHAPE_NOT_FOUND", "Shape not found"))?;
        
        // Apply filter reference to shape
        // This would be stored in the shape's style properties
        
        Ok(())
    }

    pub fn create_complex_path(&mut self, path_type: ComplexPathType, parameters: HashMap<String, f64>) -> Result<String, WASMError> {
        let commands = match path_type {
            ComplexPathType::Bezier => self.generate_bezier_path(&parameters)?,
            ComplexPathType::Spiral => self.generate_spiral_path(&parameters)?,
            ComplexPathType::Star => self.generate_star_path(&parameters)?,
            ComplexPathType::Wave => self.generate_wave_path(&parameters)?,
        };
        
        self.create_path(commands)
    }

    fn generate_bezier_path(&self, parameters: &HashMap<String, f64>) -> Result<Vec<PathCommand>, WASMError> {
        let start_x = parameters.get("start_x").unwrap_or(&0.0);
        let start_y = parameters.get("start_y").unwrap_or(&0.0);
        let cp1_x = parameters.get("cp1_x").unwrap_or(&50.0);
        let cp1_y = parameters.get("cp1_y").unwrap_or(&0.0);
        let cp2_x = parameters.get("cp2_x").unwrap_or(&50.0);
        let cp2_y = parameters.get("cp2_y").unwrap_or(&100.0);
        let end_x = parameters.get("end_x").unwrap_or(&100.0);
        let end_y = parameters.get("end_y").unwrap_or(&100.0);
        
        Ok(vec![
            PathCommand::MoveTo { x: *start_x, y: *start_y },
            PathCommand::CurveTo { 
                x1: *cp1_x, y1: *cp1_y, 
                x2: *cp2_x, y2: *cp2_y, 
                x: *end_x, y: *end_y 
            },
        ])
    }

    fn generate_spiral_path(&self, parameters: &HashMap<String, f64>) -> Result<Vec<PathCommand>, WASMError> {
        let center_x = parameters.get("center_x").unwrap_or(&50.0);
        let center_y = parameters.get("center_y").unwrap_or(&50.0);
        let start_radius = parameters.get("start_radius").unwrap_or(&5.0);
        let end_radius = parameters.get("end_radius").unwrap_or(&40.0);
        let turns = parameters.get("turns").unwrap_or(&3.0);
        let steps = 100;
        
        let mut commands = Vec::new();
        
        for i in 0..=steps {
            let t = i as f64 / steps as f64;
            let angle = t * turns * 2.0 * std::f64::consts::PI;
            let radius = start_radius + (end_radius - start_radius) * t;
            let x = center_x + radius * angle.cos();
            let y = center_y + radius * angle.sin();
            
            if i == 0 {
                commands.push(PathCommand::MoveTo { x, y });
            } else {
                commands.push(PathCommand::LineTo { x, y });
            }
        }
        
        Ok(commands)
    }

    fn generate_star_path(&self, parameters: &HashMap<String, f64>) -> Result<Vec<PathCommand>, WASMError> {
        let center_x = parameters.get("center_x").unwrap_or(&50.0);
        let center_y = parameters.get("center_y").unwrap_or(&50.0);
        let outer_radius = parameters.get("outer_radius").unwrap_or(&40.0);
        let inner_radius = parameters.get("inner_radius").unwrap_or(&20.0);
        let points = parameters.get("points").unwrap_or(&5.0) as usize;
        
        let mut commands = Vec::new();
        
        for i in 0..(points * 2) {
            let angle = (i as f64) * std::f64::consts::PI / (points as f64) - std::f64::consts::PI / 2.0;
            let radius = if i % 2 == 0 { *outer_radius } else { *inner_radius };
            let x = center_x + radius * angle.cos();
            let y = center_y + radius * angle.sin();
            
            if i == 0 {
                commands.push(PathCommand::MoveTo { x, y });
            } else {
                commands.push(PathCommand::LineTo { x, y });
            }
        }
        
        commands.push(PathCommand::ClosePath);
        Ok(commands)
    }

    fn generate_wave_path(&self, parameters: &HashMap<String, f64>) -> Result<Vec<PathCommand>, WASMError> {
        let start_x = parameters.get("start_x").unwrap_or(&0.0);
        let start_y = parameters.get("start_y").unwrap_or(&50.0);
        let end_x = parameters.get("end_x").unwrap_or(&100.0);
        let amplitude = parameters.get("amplitude").unwrap_or(&20.0);
        let frequency = parameters.get("frequency").unwrap_or(&2.0);
        let steps = 50;
        
        let mut commands = Vec::new();
        
        for i in 0..=steps {
            let t = i as f64 / steps as f64;
            let x = start_x + (end_x - start_x) * t;
            let y = start_y + amplitude * (frequency * t * 2.0 * std::f64::consts::PI).sin();
            
            if i == 0 {
                commands.push(PathCommand::MoveTo { x, y });
            } else {
                commands.push(PathCommand::LineTo { x, y });
            }
        }
        
        Ok(commands)
    }

    pub fn create_shape(&mut self, shape_type: ShapeType, position: Position, size: Size) -> Result<String, WASMError> {
        let shape_id = format!("shape_{}", get_current_timestamp() as u64);
        
        let shape = VectorShape {
            id: shape_id.clone(),
            shape_type,
            position,
            size,
            fill: Fill::default(),
            stroke: Stroke::default(),
            transform: Transform::default(),
            opacity: 1.0,
        };

        self.shapes.insert(shape_id.clone(), shape);
        Ok(shape_id)
    }

    pub fn create_path(&mut self, commands: Vec<PathCommand>) -> Result<String, WASMError> {
        let path_id = format!("path_{}", get_current_timestamp() as u64);
        
        let path = VectorPath {
            id: path_id.clone(),
            commands,
            fill: Fill::default(),
            stroke: Stroke::default(),
            transform: Transform::default(),
        };

        self.paths.insert(path_id.clone(), path);
        Ok(path_id)
    }

    pub fn create_gradient(&mut self, gradient_type: GradientType, stops: Vec<GradientStop>) -> Result<String, WASMError> {
        let gradient_id = format!("gradient_{}", get_current_timestamp() as u64);
        
        let gradient = Gradient {
            id: gradient_id.clone(),
            gradient_type,
            stops,
            transform: None,
        };

        self.gradients.insert(gradient_id.clone(), gradient);
        Ok(gradient_id)
    }

    pub fn render_to_svg(&self, width: f64, height: f64) -> String {
        let mut svg_content = String::new();
        
        svg_content.push_str(&format!(
            r#"<svg width="{}" height="{}" viewBox="0 0 {} {}" xmlns="http://www.w3.org/2000/svg">"#,
            width, height, width, height
        ));

        // Add definitions for gradients, patterns, and filters
        svg_content.push_str("<defs>");
        
        for gradient in self.gradients.values() {
            self.render_gradient(&mut svg_content, gradient);
        }
        
        for pattern in self.patterns.values() {
            self.render_pattern(&mut svg_content, pattern);
        }
        
        for filter in self.filters.values() {
            self.render_filter(&mut svg_content, filter);
        }
        
        svg_content.push_str("</defs>");

        // Render shapes
        for shape in self.shapes.values() {
            self.render_shape(&mut svg_content, shape);
        }

        // Render paths
        for path in self.paths.values() {
            self.render_path(&mut svg_content, path);
        }

        svg_content.push_str("</svg>");
        svg_content
    }

    fn render_shape(&self, svg_content: &mut String, shape: &VectorShape) {
        let transform_str = self.transform_to_string(&shape.transform);
        let fill_str = self.fill_to_string(&shape.fill);
        let stroke_str = self.stroke_to_string(&shape.stroke);

        match shape.shape_type {
            ShapeType::Rectangle => {
                svg_content.push_str(&format!(
                    r#"<rect x="{}" y="{}" width="{}" height="{}" {} {} transform="{}" opacity="{}"/>"#,
                    shape.position.x, shape.position.y, shape.size.width, shape.size.height,
                    fill_str, stroke_str, transform_str, shape.opacity
                ));
            }
            ShapeType::Circle => {
                let radius = shape.size.width / 2.0;
                let cx = shape.position.x + radius;
                let cy = shape.position.y + radius;
                svg_content.push_str(&format!(
                    r#"<circle cx="{}" cy="{}" r="{}" {} {} transform="{}" opacity="{}"/>"#,
                    cx, cy, radius, fill_str, stroke_str, transform_str, shape.opacity
                ));
            }
            ShapeType::Ellipse => {
                let rx = shape.size.width / 2.0;
                let ry = shape.size.height / 2.0;
                let cx = shape.position.x + rx;
                let cy = shape.position.y + ry;
                svg_content.push_str(&format!(
                    r#"<ellipse cx="{}" cy="{}" rx="{}" ry="{}" {} {} transform="{}" opacity="{}"/>"#,
                    cx, cy, rx, ry, fill_str, stroke_str, transform_str, shape.opacity
                ));
            }
            ShapeType::Line => {
                svg_content.push_str(&format!(
                    r#"<line x1="{}" y1="{}" x2="{}" y2="{}" {} transform="{}" opacity="{}"/>"#,
                    shape.position.x, shape.position.y, 
                    shape.position.x + shape.size.width, shape.position.y + shape.size.height,
                    stroke_str, transform_str, shape.opacity
                ));
            }
            _ => {
                // Other shape types can be implemented as needed
            }
        }
    }

    fn render_path(&self, svg_content: &mut String, path: &VectorPath) {
        let path_data = self.path_commands_to_string(&path.commands);
        let transform_str = self.transform_to_string(&path.transform);
        let fill_str = self.fill_to_string(&path.fill);
        let stroke_str = self.stroke_to_string(&path.stroke);

        svg_content.push_str(&format!(
            r#"<path d="{}" {} {} transform="{}"/>"#,
            path_data, fill_str, stroke_str, transform_str
        ));
    }

    fn render_gradient(&self, svg_content: &mut String, gradient: &Gradient) {
        match &gradient.gradient_type {
            GradientType::Linear { x1, y1, x2, y2 } => {
                svg_content.push_str(&format!(
                    r#"<linearGradient id="{}" x1="{}%" y1="{}%" x2="{}%" y2="%">"#,
                    gradient.id, x1 * 100.0, y1 * 100.0, x2 * 100.0, y2 * 100.0
                ));
            }
            GradientType::Radial { cx, cy, r, fx, fy } => {
                let fx_str = fx.map(|f| format!(" fx=\"{}%\"", f * 100.0)).unwrap_or_default();
                let fy_str = fy.map(|f| format!(" fy=\"{}%\"", f * 100.0)).unwrap_or_default();
                svg_content.push_str(&format!(
                    r#"<radialGradient id="{}" cx="{}%" cy="{}%" r="{}%"{}{}>"#,
                    gradient.id, cx * 100.0, cy * 100.0, r * 100.0, fx_str, fy_str
                ));
            }
        }

        for stop in &gradient.stops {
            svg_content.push_str(&format!(
                r#"<stop offset="{}%" stop-color="{}" stop-opacity="{}"/>"#,
                stop.offset * 100.0, stop.color, stop.opacity
            ));
        }

        match gradient.gradient_type {
            GradientType::Linear { .. } => svg_content.push_str("</linearGradient>"),
            GradientType::Radial { .. } => svg_content.push_str("</radialGradient>"),
        }
    }

    fn render_pattern(&self, svg_content: &mut String, pattern: &Pattern) {
        svg_content.push_str(&format!(
            r#"<pattern id="{}" width="{}" height="{}" patternUnits="userSpaceOnUse">{}</pattern>"#,
            pattern.id, pattern.width, pattern.height, pattern.content
        ));
    }

    fn render_filter(&self, svg_content: &mut String, filter: &Filter) {
        svg_content.push_str(&format!(r#"<filter id="{}">"#, filter.id));
        
        match filter.filter_type {
            FilterType::Blur => {
                let std_deviation = filter.parameters.get("stdDeviation").unwrap_or(&2.0);
                svg_content.push_str(&format!(
                    r#"<feGaussianBlur stdDeviation="{}"/>"#, std_deviation
                ));
            }
            FilterType::DropShadow => {
                let dx = filter.parameters.get("dx").unwrap_or(&2.0);
                let dy = filter.parameters.get("dy").unwrap_or(&2.0);
                let std_deviation = filter.parameters.get("stdDeviation").unwrap_or(&1.0);
                svg_content.push_str(&format!(
                    r#"<feDropShadow dx="{}" dy="{}" stdDeviation="{}"/>"#,
                    dx, dy, std_deviation
                ));
            }
            _ => {
                // Other filter types can be implemented as needed
            }
        }
        
        svg_content.push_str("</filter>");
    }

    fn transform_to_string(&self, transform: &Transform) -> String {
        format!(
            "translate({},{}) scale({},{}) rotate({})",
            transform.x, transform.y, transform.scale_x, transform.scale_y, transform.rotation
        )
    }

    fn fill_to_string(&self, fill: &Fill) -> String {
        if let Some(color) = &fill.color {
            format!("fill=\"{}\" fill-opacity=\"{}\"", color, fill.opacity)
        } else if let Some(gradient_id) = &fill.gradient_id {
            format!("fill=\"url(#{})\" fill-opacity=\"{}\"", gradient_id, fill.opacity)
        } else if let Some(pattern_id) = &fill.pattern_id {
            format!("fill=\"url(#{})\" fill-opacity=\"{}\"", pattern_id, fill.opacity)
        } else {
            "fill=\"none\"".to_string()
        }
    }

    fn stroke_to_string(&self, stroke: &Stroke) -> String {
        let mut stroke_str = format!(
            "stroke=\"{}\" stroke-width=\"{}\" stroke-opacity=\"{}\"",
            stroke.color, stroke.width, stroke.opacity
        );

        if let Some(dash_array) = &stroke.dash_array {
            let dash_str: Vec<String> = dash_array.iter().map(|d| d.to_string()).collect();
            stroke_str.push_str(&format!(" stroke-dasharray=\"{}\"", dash_str.join(",")));
        }

        stroke_str.push_str(&format!(" stroke-linecap=\"{}\"", match stroke.line_cap {
            LineCap::Butt => "butt",
            LineCap::Round => "round",
            LineCap::Square => "square",
        }));

        stroke_str.push_str(&format!(" stroke-linejoin=\"{}\"", match stroke.line_join {
            LineJoin::Miter => "miter",
            LineJoin::Round => "round",
            LineJoin::Bevel => "bevel",
        }));

        stroke_str
    }

    fn path_commands_to_string(&self, commands: &[PathCommand]) -> String {
        let mut path_data = String::new();
        
        for command in commands {
            match command {
                PathCommand::MoveTo { x, y } => path_data.push_str(&format!("M {} {} ", x, y)),
                PathCommand::LineTo { x, y } => path_data.push_str(&format!("L {} {} ", x, y)),
                PathCommand::CurveTo { x1, y1, x2, y2, x, y } => {
                    path_data.push_str(&format!("C {} {} {} {} {} {} ", x1, y1, x2, y2, x, y));
                }
                PathCommand::QuadraticCurveTo { x1, y1, x, y } => {
                    path_data.push_str(&format!("Q {} {} {} {} ", x1, y1, x, y));
                }
                PathCommand::Arc { rx, ry, rotation, large_arc, sweep, x, y } => {
                    path_data.push_str(&format!(
                        "A {} {} {} {} {} {} {} ",
                        rx, ry, rotation,
                        if *large_arc { 1 } else { 0 },
                        if *sweep { 1 } else { 0 },
                        x, y
                    ));
                }
                PathCommand::ClosePath => path_data.push_str("Z "),
            }
        }
        
        path_data.trim().to_string()
    }
}

// Default implementations
impl Default for ChartConfig {
    fn default() -> Self {
        Self {
            width: 400.0,
            height: 300.0,
            margin: ChartMargin {
                top: 20.0,
                right: 20.0,
                bottom: 40.0,
                left: 40.0,
            },
            responsive: true,
            maintain_aspect_ratio: true,
            background_color: Some("#ffffff".to_string()),
            title: None,
            legend: None,
            tooltip: Some(ChartTooltip {
                enabled: true,
                background_color: "#000000".to_string(),
                text_color: "#ffffff".to_string(),
                border_color: "#cccccc".to_string(),
                border_width: 1.0,
                font_size: 12.0,
                padding: 8.0,
            }),
        }
    }
}

impl Default for ChartAxis {
    fn default() -> Self {
        Self {
            label: None,
            show_grid: true,
            show_ticks: true,
            tick_count: Some(5),
            min_value: None,
            max_value: None,
            scale_type: ScaleType::Linear,
            format: None,
            color: "#666666".to_string(),
            font_size: 12.0,
        }
    }
}

impl Default for ChartStyling {
    fn default() -> Self {
        Self {
            color_palette: vec![
                "#1f77b4".to_string(),
                "#ff7f0e".to_string(),
                "#2ca02c".to_string(),
                "#d62728".to_string(),
                "#9467bd".to_string(),
                "#8c564b".to_string(),
                "#e377c2".to_string(),
                "#7f7f7f".to_string(),
                "#bcbd22".to_string(),
                "#17becf".to_string(),
            ],
            gradient_fills: false,
            drop_shadow: false,
            border_radius: 0.0,
            grid_color: "#e0e0e0".to_string(),
            grid_opacity: 0.5,
        }
    }
}

impl Default for ChartInteractions {
    fn default() -> Self {
        Self {
            zoom_enabled: false,
            pan_enabled: false,
            hover_effects: true,
            click_events: true,
            brush_selection: false,
            crosshair: false,
        }
    }
}

impl Default for ChartAnimations {
    fn default() -> Self {
        Self {
            enabled: true,
            duration: 1000.0,
            easing: EasingFunction::EaseInOut,
            stagger_delay: 50.0,
            entrance_animation: AnimationType::Transform,
            update_animation: AnimationType::Transform,
        }
    }
}

impl Default for Fill {
    fn default() -> Self {
        Self {
            color: Some("#000000".to_string()),
            gradient_id: None,
            pattern_id: None,
            opacity: 1.0,
        }
    }
}

impl Default for Stroke {
    fn default() -> Self {
        Self {
            color: "#000000".to_string(),
            width: 1.0,
            dash_array: None,
            line_cap: LineCap::Butt,
            line_join: LineJoin::Miter,
            opacity: 1.0,
        }
    }
}

// WASM-bindgen interface functions for charts and visualization
#[wasm_bindgen]
pub fn create_chart(chart_type: &str, data_source_id: &str, config_json: &str) -> Result<String, JsValue> {
    let mut global_engine = ENGINE.lock().unwrap();
    if let Some(engine) = global_engine.as_mut() {
        let chart_type = match chart_type {
            "line" => ChartType::Line,
            "bar" => ChartType::Bar,
            "pie" => ChartType::Pie,
            "scatter" => ChartType::Scatter,
            "area" => ChartType::Area,
            "histogram" => ChartType::Histogram,
            "heatmap" => ChartType::Heatmap,
            "treemap" => ChartType::Treemap,
            "sankey" => ChartType::Sankey,
            "radar" => ChartType::Radar,
            "gauge" => ChartType::Gauge,
            "candlestick" => ChartType::Candlestick,
            _ => return Err(JsValue::from_str("Invalid chart type")),
        };
        
        let config: ChartConfig = serde_json::from_str(config_json)
            .map_err(|e| JsValue::from_str(&format!("Failed to parse config: {}", e)))?;
        
        engine.chart_renderer.create_chart(chart_type, data_source_id.to_string(), config)
            .map_err(|e| JsValue::from_str(&format!("Failed to create chart: {}", e.message)))
    } else {
        Err(JsValue::from_str("Engine not initialized"))
    }
}

#[wasm_bindgen]
pub fn render_chart(chart_id: &str, data_json: &str) -> Result<String, JsValue> {
    let mut global_engine = ENGINE.lock().unwrap();
    if let Some(engine) = global_engine.as_mut() {
        let data: serde_json::Value = serde_json::from_str(data_json)
            .map_err(|e| JsValue::from_str(&format!("Failed to parse data: {}", e)))?;
        
        let rendered_chart = engine.chart_renderer.render_chart(chart_id, &data)
            .map_err(|e| JsValue::from_str(&format!("Failed to render chart: {}", e.message)))?;
        
        serde_json::to_string(&rendered_chart)
            .map_err(|e| JsValue::from_str(&format!("Failed to serialize chart: {}", e)))
    } else {
        Err(JsValue::from_str("Engine not initialized"))
    }
}

#[wasm_bindgen]
pub fn add_chart_series(chart_id: &str, series_json: &str) -> Result<(), JsValue> {
    let mut global_engine = ENGINE.lock().unwrap();
    if let Some(engine) = global_engine.as_mut() {
        let series: ChartSeries = serde_json::from_str(series_json)
            .map_err(|e| JsValue::from_str(&format!("Failed to parse series: {}", e)))?;
        
        engine.chart_renderer.add_series(chart_id, series)
            .map_err(|e| JsValue::from_str(&format!("Failed to add series: {}", e.message)))
    } else {
        Err(JsValue::from_str("Engine not initialized"))
    }
}

#[wasm_bindgen]
pub fn create_vector_shape(shape_type: &str, x: f64, y: f64, width: f64, height: f64) -> Result<String, JsValue> {
    let mut global_engine = ENGINE.lock().unwrap();
    if let Some(engine) = global_engine.as_mut() {
        let shape_type = match shape_type {
            "rectangle" => ShapeType::Rectangle,
            "circle" => ShapeType::Circle,
            "ellipse" => ShapeType::Ellipse,
            "line" => ShapeType::Line,
            _ => return Err(JsValue::from_str("Invalid shape type")),
        };
        
        let position = Position { x, y };
        let size = Size { width, height };
        
        engine.vector_engine.create_shape(shape_type, position, size)
            .map_err(|e| JsValue::from_str(&format!("Failed to create shape: {}", e.message)))
    } else {
        Err(JsValue::from_str("Engine not initialized"))
    }
}

#[wasm_bindgen]
pub fn render_vector_graphics(width: f64, height: f64) -> Result<String, JsValue> {
    let global_engine = ENGINE.lock().unwrap();
    if let Some(engine) = global_engine.as_ref() {
        Ok(engine.vector_engine.render_to_svg(width, height))
    } else {
        Err(JsValue::from_str("Engine not initialized"))
    }
}

#[wasm_bindgen]
pub fn get_chart_performance_stats() -> Result<String, JsValue> {
    let global_engine = ENGINE.lock().unwrap();
    if let Some(engine) = global_engine.as_ref() {
        serde_json::to_string(&engine.chart_renderer.performance_stats)
            .map_err(|e| JsValue::from_str(&format!("Failed to serialize stats: {}", e)))
    } else {
        Err(JsValue::from_str("Engine not initialized"))
    }
}

#[wasm_bindgen]
pub fn create_data_source(source_id: &str, source_type: &str, data_json: &str) -> Result<(), JsValue> {
    let mut global_engine = ENGINE.lock().unwrap();
    if let Some(engine) = global_engine.as_mut() {
        let source_type = match source_type {
            "static" => DataSourceType::Static,
            "dynamic" => DataSourceType::Dynamic,
            "stream" => DataSourceType::Stream,
            "computed" => DataSourceType::Computed,
            _ => return Err(JsValue::from_str("Invalid data source type")),
        };
        
        let data: serde_json::Value = serde_json::from_str(data_json)
            .map_err(|e| JsValue::from_str(&format!("Failed to parse data: {}", e)))?;
        
        let data_source = DataSource::new(source_id.to_string(), source_type, data);
        engine.document_state.data_sources.insert(source_id.to_string(), data_source);
        
        Ok(())
    } else {
        Err(JsValue::from_str("Engine not initialized"))
    }
}

#[wasm_bindgen]
pub fn update_data_source(source_id: &str, data_json: &str) -> Result<(), JsValue> {
    let mut global_engine = ENGINE.lock().unwrap();
    if let Some(engine) = global_engine.as_mut() {
        let new_data: serde_json::Value = serde_json::from_str(data_json)
            .map_err(|e| JsValue::from_str(&format!("Failed to parse data: {}", e)))?;
        
        if let Some(data_source) = engine.document_state.data_sources.get_mut(source_id) {
            data_source.update_data(new_data)
                .map_err(|e| JsValue::from_str(&format!("Failed to update data source: {}", e.message)))?;
        } else {
            return Err(JsValue::from_str("Data source not found"));
        }
        
        Ok(())
    } else {
        Err(JsValue::from_str("Engine not initialized"))
    }
}

#[wasm_bindgen]
pub fn create_data_binding(source_id: &str, target_element: &str, property_path: &str, transform_function: Option<String>) -> Result<String, JsValue> {
    let mut global_engine = ENGINE.lock().unwrap();
    if let Some(engine) = global_engine.as_mut() {
        let binding = DataBinding {
            source_id: source_id.to_string(),
            target_element: target_element.to_string(),
            property_path: property_path.to_string(),
            transform_function,
            update_trigger: UpdateTrigger::Immediate,
        };
        
        let binding_id = engine.data_binding_manager.add_binding(binding);
        Ok(binding_id)
    } else {
        Err(JsValue::from_str("Engine not initialized"))
    }
}

#[wasm_bindgen]
pub fn remove_data_binding(binding_id: &str) -> Result<(), JsValue> {
    let mut global_engine = ENGINE.lock().unwrap();
    if let Some(engine) = global_engine.as_mut() {
        engine.data_binding_manager.remove_binding(binding_id);
        Ok(())
    } else {
        Err(JsValue::from_str("Engine not initialized"))
    }
}

#[wasm_bindgen]
pub fn get_data_statistics(source_id: &str) -> Result<String, JsValue> {
    let global_engine = ENGINE.lock().unwrap();
    if let Some(engine) = global_engine.as_ref() {
        if let Some(data_source) = engine.document_state.data_sources.get(source_id) {
            let stats = data_source.get_data_statistics();
            serde_json::to_string(&stats)
                .map_err(|e| JsValue::from_str(&format!("Failed to serialize statistics: {}", e)))
        } else {
            Err(JsValue::from_str("Data source not found"))
        }
    } else {
        Err(JsValue::from_str("Engine not initialized"))
    }
}

#[wasm_bindgen]
pub fn animate_vector_shape(shape_id: &str, target_transform_json: &str, duration: f64) -> Result<String, JsValue> {
    let mut global_engine = ENGINE.lock().unwrap();
    if let Some(engine) = global_engine.as_mut() {
        let target_transform: Transform = serde_json::from_str(target_transform_json)
            .map_err(|e| JsValue::from_str(&format!("Failed to parse transform: {}", e)))?;
        
        engine.vector_engine.animate_shape(shape_id, target_transform, duration)
            .map_err(|e| JsValue::from_str(&format!("Failed to animate shape: {}", e.message)))
    } else {
        Err(JsValue::from_str("Engine not initialized"))
    }
}

#[wasm_bindgen]
pub fn create_complex_path(path_type: &str, parameters_json: &str) -> Result<String, JsValue> {
    let mut global_engine = ENGINE.lock().unwrap();
    if let Some(engine) = global_engine.as_mut() {
        let path_type = match path_type {
            "bezier" => ComplexPathType::Bezier,
            "spiral" => ComplexPathType::Spiral,
            "star" => ComplexPathType::Star,
            "wave" => ComplexPathType::Wave,
            _ => return Err(JsValue::from_str("Invalid path type")),
        };
        
        let parameters: HashMap<String, f64> = serde_json::from_str(parameters_json)
            .map_err(|e| JsValue::from_str(&format!("Failed to parse parameters: {}", e)))?;
        
        engine.vector_engine.create_complex_path(path_type, parameters)
            .map_err(|e| JsValue::from_str(&format!("Failed to create complex path: {}", e.message)))
    } else {
        Err(JsValue::from_str("Engine not initialized"))
    }
}

#[wasm_bindgen]
pub fn enable_chart_interactions_wasm(chart_id: &str, interactions_json: &str) -> Result<(), JsValue> {
    let mut global_engine = ENGINE.lock().unwrap();
    if let Some(engine) = global_engine.as_mut() {
        let interactions: ChartInteractions = serde_json::from_str(interactions_json)
            .map_err(|e| JsValue::from_str(&format!("Failed to parse interactions: {}", e)))?;
        
        engine.chart_renderer.enable_chart_interactions(chart_id, interactions)
            .map_err(|e| JsValue::from_str(&format!("Failed to enable interactions: {}", e.message)))
    } else {
        Err(JsValue::from_str("Engine not initialized"))
    }
}

#[wasm_bindgen]
pub fn update_chart_data_wasm(chart_id: &str, data_json: &str) -> Result<(), JsValue> {
    let mut global_engine = ENGINE.lock().unwrap();
    if let Some(engine) = global_engine.as_mut() {
        let data: serde_json::Value = serde_json::from_str(data_json)
            .map_err(|e| JsValue::from_str(&format!("Failed to parse data: {}", e)))?;
        
        engine.chart_renderer.update_chart_data(chart_id, &data)
            .map_err(|e| JsValue::from_str(&format!("Failed to update chart data: {}", e.message)))
    } else {
        Err(JsValue::from_str("Engine not initialized"))
    }
}

#[wasm_bindgen]
pub fn add_interaction_delegate(target_element: &str, event_types_json: &str, handler_id: &str, capture: bool, priority: i32) -> Result<(), JsValue> {
    let mut global_engine = ENGINE.lock().unwrap();
    if let Some(engine) = global_engine.as_mut() {
        let event_types: Vec<String> = serde_json::from_str(event_types_json)
            .map_err(|e| JsValue::from_str(&format!("Failed to parse event types: {}", e)))?;
        
        let interaction_types: Vec<InteractionType> = event_types.iter()
            .filter_map(|event_type| match event_type.as_str() {
                "click" => Some(InteractionType::Click),
                "doubleclick" => Some(InteractionType::DoubleClick),
                "mousedown" => Some(InteractionType::MouseDown),
                "mouseup" => Some(InteractionType::MouseUp),
                "mousemove" => Some(InteractionType::MouseMove),
                "touchstart" => Some(InteractionType::TouchStart),
                "touchmove" => Some(InteractionType::TouchMove),
                "touchend" => Some(InteractionType::TouchEnd),
                "keydown" => Some(InteractionType::KeyDown),
                "keyup" => Some(InteractionType::KeyUp),
                "scroll" => Some(InteractionType::Scroll),
                "focus" => Some(InteractionType::Focus),
                "blur" => Some(InteractionType::Blur),
                _ => None,
            })
            .collect();
        
        let delegate = EventDelegate {
            element_id: target_element.to_string(),
            event_types: interaction_types,
            handler_id: handler_id.to_string(),
            capture,
            priority,
        };
        
        engine.add_interaction_delegate(target_element, delegate)
            .map_err(|e| JsValue::from_str(&format!("Failed to add interaction delegate: {}", e.message)))
    } else {
        Err(JsValue::from_str("Engine not initialized"))
    }
}

#[wasm_bindgen]
pub fn remove_interaction_delegate(target_element: &str, handler_id: &str) -> Result<(), JsValue> {
    let mut global_engine = ENGINE.lock().unwrap();
    if let Some(engine) = global_engine.as_mut() {
        engine.remove_interaction_delegate(target_element, handler_id)
            .map_err(|e| JsValue::from_str(&format!("Failed to remove interaction delegate: {}", e.message)))
    } else {
        Err(JsValue::from_str("Engine not initialized"))
    }
}

#[wasm_bindgen]
pub fn get_interaction_state(element_id: &str) -> Result<String, JsValue> {
    let global_engine = ENGINE.lock().unwrap();
    if let Some(engine) = global_engine.as_ref() {
        if let Some(state) = engine.get_interaction_state(element_id) {
            serde_json::to_string(state)
                .map_err(|e| JsValue::from_str(&format!("Failed to serialize interaction state: {}", e)))
        } else {
            Err(JsValue::from_str("Interaction state not found"))
        }
    } else {
        Err(JsValue::from_str("Engine not initialized"))
    }
}

#[wasm_bindgen]
pub fn get_interaction_metrics() -> Result<String, JsValue> {
    let global_engine = ENGINE.lock().unwrap();
    if let Some(engine) = global_engine.as_ref() {
        let metrics = engine.get_interaction_metrics();
        serde_json::to_string(metrics)
            .map_err(|e| JsValue::from_str(&format!("Failed to serialize interaction metrics: {}", e)))
    } else {
        Err(JsValue::from_str("Engine not initialized"))
    }
}

#[wasm_bindgen]
pub fn update_device_capabilities(device_info_json: &str) -> Result<(), JsValue> {
    let mut global_engine = ENGINE.lock().unwrap();
    if let Some(engine) = global_engine.as_mut() {
        let device_info: DeviceInfo = serde_json::from_str(device_info_json)
            .map_err(|e| JsValue::from_str(&format!("Failed to parse device info: {}", e)))?;
        
        engine.update_device_capabilities(device_info)
            .map_err(|e| JsValue::from_str(&format!("Failed to update device capabilities: {}", e.message)))
    } else {
        Err(JsValue::from_str("Engine not initialized"))
    }
}

#[wasm_bindgen]
pub fn process_touch_gesture(touch_data_json: &str, timestamp: f64) -> Result<String, JsValue> {
    let mut global_engine = ENGINE.lock().unwrap();
    if let Some(engine) = global_engine.as_mut() {
        let touch_data: TouchData = serde_json::from_str(touch_data_json)
            .map_err(|e| JsValue::from_str(&format!("Failed to parse touch data: {}", e)))?;
        
        let gesture_events = engine.gesture_recognizer.process_touch_input(&touch_data, timestamp);
        serde_json::to_string(&gesture_events)
            .map_err(|e| JsValue::from_str(&format!("Failed to serialize gesture events: {}", e)))
    } else {
        Err(JsValue::from_str("Engine not initialized"))
    }
}

#[wasm_bindgen]
pub fn get_gesture_history() -> Result<String, JsValue> {
    let global_engine = ENGINE.lock().unwrap();
    if let Some(engine) = global_engine.as_ref() {
        let history = engine.gesture_recognizer.get_gesture_history();
        serde_json::to_string(history)
            .map_err(|e| JsValue::from_str(&format!("Failed to serialize gesture history: {}", e)))
    } else {
        Err(JsValue::from_str("Engine not initialized"))
    }
}

#[wasm_bindgen]
pub fn get_optimal_touch_target_size(element_width: f64, element_height: f64) -> Result<String, JsValue> {
    let global_engine = ENGINE.lock().unwrap();
    if let Some(engine) = global_engine.as_ref() {
        let element_size = Size {
            width: element_width,
            height: element_height,
        };
        let optimal_size = engine.responsive_adapter.get_optimal_touch_target_size(&element_size);
        serde_json::to_string(&optimal_size)
            .map_err(|e| JsValue::from_str(&format!("Failed to serialize optimal size: {}", e)))
    } else {
        Err(JsValue::from_str("Engine not initialized"))
    }
}

#[wasm_bindgen]
pub fn get_interaction_settings() -> Result<String, JsValue> {
    let global_engine = ENGINE.lock().unwrap();
    if let Some(engine) = global_engine.as_ref() {
        let settings = engine.responsive_adapter.get_interaction_settings();
        serde_json::to_string(settings)
            .map_err(|e| JsValue::from_str(&format!("Failed to serialize interaction settings: {}", e)))
    } else {
        Err(JsValue::from_str("Engine not initialized"))
    }
}

#[wasm_bindgen]
pub fn should_throttle_event(event_type: &str, last_event_time: f64) -> Result<bool, JsValue> {
    let global_engine = ENGINE.lock().unwrap();
    if let Some(engine) = global_engine.as_ref() {
        let interaction_type = match event_type {
            "mousemove" => InteractionType::MouseMove,
            "touchmove" => InteractionType::TouchMove,
            "scroll" => InteractionType::Scroll,
            "wheel" => InteractionType::Wheel,
            _ => return Ok(false),
        };
        
        Ok(engine.responsive_adapter.should_throttle_event(&interaction_type, last_event_time))
    } else {
        Err(JsValue::from_str("Engine not initialized"))
    }
}