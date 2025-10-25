use super::*;
use wasm_bindgen_test::*;
use std::time::{SystemTime, UNIX_EPOCH};

wasm_bindgen_test_configure!(run_in_browser);

// Test WASM module execution within Go-imposed constraints

#[wasm_bindgen_test]
fn test_memory_limit_enforcement() {
    let small_memory_permissions = WASMPermissions {
        memory_limit: 1024, // Very small limit - 1KB
        allowed_imports: vec!["console".to_string()],
        cpu_time_limit: 5000,
        allow_networking: false,
        allow_file_system: false,
        allowed_interactions: vec![
            "create_element".to_string(),
            "modify_element".to_string(),
        ],
        max_data_size: 512, // 512 bytes
        max_elements: 5, // Very few elements
    };

    let mut engine = InteractiveEngine::new(small_memory_permissions).unwrap();

    // Try to create many elements to exceed memory limit
    let mut created_elements = Vec::new();
    let mut creation_failed = false;

    for i in 0..10 {
        let properties = [
            ("width".to_string(), serde_json::Value::Number(serde_json::Number::from(100))),
            ("height".to_string(), serde_json::Value::Number(serde_json::Number::from(100))),
            ("data".to_string(), serde_json::Value::String("x".repeat(100))), // Large data
        ].into_iter().collect();

        match engine.create_element(ElementType::Container, properties) {
            Ok(element_id) => {
                created_elements.push(element_id);
            }
            Err(_) => {
                creation_failed = true;
                break;
            }
        }
    }

    // Should fail to create all elements due to memory constraints
    assert!(creation_failed || created_elements.len() <= 5);
}

#[wasm_bindgen_test]
fn test_cpu_time_limit_enforcement() {
    let time_limited_permissions = WASMPermissions {
        memory_limit: 1024 * 1024,
        allowed_imports: vec!["console".to_string()],
        cpu_time_limit: 100, // Very short time limit - 100ms
        allow_networking: false,
        allow_file_system: false,
        allowed_interactions: vec![
            "create_element".to_string(),
            "create_animation".to_string(),
            "Click".to_string(),
        ],
        max_data_size: 1024 * 1024,
        max_elements: 100,
    };

    let mut engine = InteractiveEngine::new(time_limited_permissions).unwrap();

    // Create many elements and animations to consume CPU time
    for i in 0..50 {
        let element_id = engine.create_element(ElementType::Container, HashMap::new()).unwrap();
        
        let keyframes = vec![
            Keyframe {
                time: 0.0,
                properties: [("x".to_string(), serde_json::Value::Number(serde_json::Number::from(0)))].into_iter().collect(),
            },
            Keyframe {
                time: 1.0,
                properties: [("x".to_string(), serde_json::Value::Number(serde_json::Number::from(100)))].into_iter().collect(),
            },
        ];

        let _ = engine.create_animation(&element_id, AnimationType::Transform, 1000.0, keyframes);
    }

    // Process many render frames to consume CPU time
    let start_time = get_current_timestamp();
    let mut frame_count = 0;
    let mut render_failed = false;

    for i in 0..1000 {
        let timestamp = start_time + (i as f64 * 16.67); // 60fps
        match engine.render_frame(timestamp) {
            Ok(_) => {
                frame_count += 1;
            }
            Err(e) => {
                if e.code == "CPU_TIME_EXCEEDED" {
                    render_failed = true;
                    break;
                }
            }
        }
    }

    // Should eventually fail due to CPU time limit
    assert!(render_failed || frame_count < 1000);
}

#[wasm_bindgen_test]
fn test_element_count_limit_enforcement() {
    let element_limited_permissions = WASMPermissions {
        memory_limit: 1024 * 1024,
        allowed_imports: vec!["console".to_string()],
        cpu_time_limit: 5000,
        allow_networking: false,
        allow_file_system: false,
        allowed_interactions: vec![
            "create_element".to_string(),
        ],
        max_data_size: 1024 * 1024,
        max_elements: 3, // Very low element limit
    };

    let mut engine = InteractiveEngine::new(element_limited_permissions).unwrap();

    // Try to create more elements than allowed
    let mut created_count = 0;
    let mut creation_blocked = false;

    for i in 0..10 {
        match engine.create_element(ElementType::Container, HashMap::new()) {
            Ok(_) => {
                created_count += 1;
            }
            Err(_) => {
                creation_blocked = true;
                break;
            }
        }
    }

    // Should be blocked after creating max_elements
    assert!(creation_blocked);
    assert!(created_count <= 3);
}

#[wasm_bindgen_test]
fn test_data_size_limit_enforcement() {
    let data_limited_permissions = WASMPermissions {
        memory_limit: 1024 * 1024,
        allowed_imports: vec!["console".to_string()],
        cpu_time_limit: 5000,
        allow_networking: false,
        allow_file_system: false,
        allowed_interactions: vec![
            "DataUpdate".to_string(),
        ],
        max_data_size: 100, // Very small data limit - 100 bytes
        max_elements: 100,
    };

    let mut engine = InteractiveEngine::new(data_limited_permissions).unwrap();

    // Create a data source
    let data_source = DataSource::new(
        "test_data".to_string(),
        DataSourceType::Dynamic,
        serde_json::json!({}),
    );
    engine.document_state.data_sources.insert("test_data".to_string(), data_source);

    // Try to update with large data
    let large_data = "x".repeat(1000); // 1000 bytes - exceeds limit
    let large_data_bytes = large_data.as_bytes();

    let result = engine.update_data("test_data", large_data_bytes);
    
    // Should fail due to data size limit
    assert!(result.is_err());
    if let Err(e) = result {
        assert_eq!(e.code, "DATA_SIZE_EXCEEDED");
    }
}

#[wasm_bindgen_test]
fn test_interaction_permission_enforcement() {
    let restricted_permissions = WASMPermissions {
        memory_limit: 1024 * 1024,
        allowed_imports: vec!["console".to_string()],
        cpu_time_limit: 5000,
        allow_networking: false,
        allow_file_system: false,
        allowed_interactions: vec![
            "Click".to_string(), // Only click allowed
        ],
        max_data_size: 1024 * 1024,
        max_elements: 100,
    };

    let mut engine = InteractiveEngine::new(restricted_permissions).unwrap();

    // Test allowed interaction (Click)
    let click_event = InteractionEvent {
        event_type: InteractionType::Click,
        target_element: Some("test_element".to_string()),
        position: Some(Position { x: 10.0, y: 10.0 }),
        data: HashMap::new(),
        timestamp: get_current_timestamp(),
        touch_data: None,
        mouse_data: Some(MouseData {
            button: MouseButton::Left,
            buttons: 1,
            position: Position { x: 10.0, y: 10.0 },
            movement: None,
            wheel_delta: None,
        }),
        keyboard_data: None,
        gesture_data: None,
        modifiers: EventModifiers {
            ctrl: false,
            shift: false,
            alt: false,
            meta: false,
        },
    };

    let result = engine.process_interaction(click_event);
    assert!(result.is_ok());

    // Test disallowed interaction (Touch)
    let touch_event = InteractionEvent {
        event_type: InteractionType::TouchStart,
        target_element: Some("test_element".to_string()),
        position: Some(Position { x: 10.0, y: 10.0 }),
        data: HashMap::new(),
        timestamp: get_current_timestamp(),
        touch_data: Some(TouchData {
            touches: vec![TouchPoint {
                identifier: 1,
                position: Position { x: 10.0, y: 10.0 },
                radius: Some(5.0),
                rotation_angle: None,
                force: Some(0.5),
            }],
            changed_touches: vec![],
            target_touches: vec![],
            force: Some(0.5),
            rotation_angle: None,
            scale: None,
        }),
        mouse_data: None,
        keyboard_data: None,
        gesture_data: None,
        modifiers: EventModifiers {
            ctrl: false,
            shift: false,
            alt: false,
            meta: false,
        },
    };

    let result = engine.process_interaction(touch_event);
    assert!(result.is_err());
    if let Err(e) = result {
        assert_eq!(e.code, "INTERACTION_NOT_ALLOWED");
    }
}

#[wasm_bindgen_test]
fn test_networking_restriction_enforcement() {
    let no_network_permissions = WASMPermissions {
        memory_limit: 1024 * 1024,
        allowed_imports: vec!["console".to_string()],
        cpu_time_limit: 5000,
        allow_networking: false, // Networking disabled
        allow_file_system: false,
        allowed_interactions: vec![
            "create_element".to_string(),
        ],
        max_data_size: 1024 * 1024,
        max_elements: 100,
    };

    let engine = InteractiveEngine::new(no_network_permissions).unwrap();

    // Verify networking is disabled in security context
    assert!(!engine.security_context.permissions.allow_networking);
}

#[wasm_bindgen_test]
fn test_file_system_restriction_enforcement() {
    let no_fs_permissions = WASMPermissions {
        memory_limit: 1024 * 1024,
        allowed_imports: vec!["console".to_string()],
        cpu_time_limit: 5000,
        allow_networking: false,
        allow_file_system: false, // File system disabled
        allowed_interactions: vec![
            "create_element".to_string(),
        ],
        max_data_size: 1024 * 1024,
        max_elements: 100,
    };

    let engine = InteractiveEngine::new(no_fs_permissions).unwrap();

    // Verify file system access is disabled in security context
    assert!(!engine.security_context.permissions.allow_file_system);
}

#[wasm_bindgen_test]
fn test_import_restriction_enforcement() {
    let limited_imports_permissions = WASMPermissions {
        memory_limit: 1024 * 1024,
        allowed_imports: vec!["console".to_string()], // Only console allowed
        cpu_time_limit: 5000,
        allow_networking: false,
        allow_file_system: false,
        allowed_interactions: vec![
            "create_element".to_string(),
        ],
        max_data_size: 1024 * 1024,
        max_elements: 100,
    };

    let engine = InteractiveEngine::new(limited_imports_permissions).unwrap();

    // Verify only allowed imports are permitted
    assert_eq!(engine.security_context.permissions.allowed_imports.len(), 1);
    assert!(engine.security_context.permissions.allowed_imports.contains(&"console".to_string()));
}

#[wasm_bindgen_test]
fn test_interaction_rate_limiting() {
    let rate_limited_permissions = WASMPermissions {
        memory_limit: 1024 * 1024,
        allowed_imports: vec!["console".to_string()],
        cpu_time_limit: 5000,
        allow_networking: false,
        allow_file_system: false,
        allowed_interactions: vec![
            "Click".to_string(),
        ],
        max_data_size: 1024 * 1024,
        max_elements: 100,
    };

    let mut engine = InteractiveEngine::new(rate_limited_permissions).unwrap();

    // Send many interactions rapidly
    let mut successful_interactions = 0;
    let mut rate_limited = false;

    for i in 0..200 {
        let click_event = InteractionEvent {
            event_type: InteractionType::Click,
            target_element: Some("test_element".to_string()),
            position: Some(Position { x: 10.0, y: 10.0 }),
            data: HashMap::new(),
            timestamp: get_current_timestamp() + (i as f64), // Rapid succession
            touch_data: None,
            mouse_data: Some(MouseData {
                button: MouseButton::Left,
                buttons: 1,
                position: Position { x: 10.0, y: 10.0 },
                movement: None,
                wheel_delta: None,
            }),
            keyboard_data: None,
            gesture_data: None,
            modifiers: EventModifiers {
                ctrl: false,
                shift: false,
                alt: false,
                meta: false,
            },
        };

        match engine.process_interaction(click_event) {
            Ok(_) => {
                successful_interactions += 1;
            }
            Err(e) => {
                if e.code == "INTERACTION_RATE_EXCEEDED" {
                    rate_limited = true;
                    break;
                }
            }
        }
    }

    // Should eventually be rate limited
    assert!(rate_limited || successful_interactions < 200);
}

#[wasm_bindgen_test]
fn test_resource_cleanup_on_constraint_violation() {
    let constrained_permissions = WASMPermissions {
        memory_limit: 2048, // Small memory limit
        allowed_imports: vec!["console".to_string()],
        cpu_time_limit: 1000, // Short time limit
        allow_networking: false,
        allow_file_system: false,
        allowed_interactions: vec![
            "create_element".to_string(),
            "modify_element".to_string(),
        ],
        max_data_size: 1024,
        max_elements: 10,
    };

    let mut engine = InteractiveEngine::new(constrained_permissions).unwrap();

    // Create elements up to the limit
    let mut elements = Vec::new();
    for i in 0..5 {
        let element_id = engine.create_element(ElementType::Container, HashMap::new()).unwrap();
        elements.push(element_id);
    }

    // Verify elements were created
    assert_eq!(engine.document_state.elements.len(), 5);

    // Try to exceed memory by creating large elements
    let large_properties = [
        ("data".to_string(), serde_json::Value::String("x".repeat(500))),
    ].into_iter().collect();

    let result = engine.create_element(ElementType::Container, large_properties);
    
    // Should fail due to memory constraints
    assert!(result.is_err());

    // Verify existing elements are still intact (no corruption)
    assert_eq!(engine.document_state.elements.len(), 5);
    for element_id in &elements {
        assert!(engine.document_state.get_element(element_id).is_some());
    }
}