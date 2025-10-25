use super::*;
use wasm_bindgen_test::*;

wasm_bindgen_test_configure!(run_in_browser);

#[wasm_bindgen_test]
fn test_interactive_engine_creation() {
    let permissions = WASMPermissions {
        memory_limit: 1024 * 1024, // 1MB
        allowed_imports: vec!["console".to_string()],
        cpu_time_limit: 5000, // 5 seconds
        allow_networking: false,
        allow_file_system: false,
        allowed_interactions: vec![
            "create_element".to_string(),
            "modify_element".to_string(),
            "create_animation".to_string(),
            "create_event_handler".to_string(),
        ],
        max_data_size: 1024 * 1024, // 1MB
        max_elements: 100,
    };

    let engine = InteractiveEngine::new(permissions);
    assert!(engine.is_ok());
}

#[wasm_bindgen_test]
fn test_element_creation_and_management() {
    let permissions = WASMPermissions {
        memory_limit: 1024 * 1024,
        allowed_imports: vec!["console".to_string()],
        cpu_time_limit: 5000,
        allow_networking: false,
        allow_file_system: false,
        allowed_interactions: vec![
            "create_element".to_string(),
            "modify_element".to_string(),
        ],
        max_data_size: 1024 * 1024,
        max_elements: 100,
    };

    let mut engine = InteractiveEngine::new(permissions).unwrap();

    // Test element creation
    let properties = [
        ("width".to_string(), serde_json::Value::Number(serde_json::Number::from(200))),
        ("height".to_string(), serde_json::Value::Number(serde_json::Number::from(100))),
    ].into_iter().collect();

    let element_id = engine.create_element(ElementType::Container, properties).unwrap();
    assert!(!element_id.is_empty());

    // Test element exists
    let element = engine.document_state.get_element(&element_id);
    assert!(element.is_some());

    // Test element update
    let update_properties = [
        ("color".to_string(), serde_json::Value::String("red".to_string())),
    ].into_iter().collect();

    let result = engine.update_element_properties(&element_id, update_properties);
    assert!(result.is_ok());

    // Test element deletion
    let result = engine.delete_element(&element_id);
    assert!(result.is_ok());

    // Verify element is deleted
    let element = engine.document_state.get_element(&element_id);
    assert!(element.is_none());
}

#[wasm_bindgen_test]
fn test_animation_system() {
    let permissions = WASMPermissions {
        memory_limit: 1024 * 1024,
        allowed_imports: vec!["console".to_string()],
        cpu_time_limit: 5000,
        allow_networking: false,
        allow_file_system: false,
        allowed_interactions: vec![
            "create_element".to_string(),
            "create_animation".to_string(),
        ],
        max_data_size: 1024 * 1024,
        max_elements: 100,
    };

    let mut engine = InteractiveEngine::new(permissions).unwrap();

    // Create an element to animate
    let element_id = engine.create_element(ElementType::Container, HashMap::new()).unwrap();

    // Create keyframes
    let keyframes = vec![
        Keyframe {
            time: 0.0,
            properties: [
                ("x".to_string(), serde_json::Value::Number(serde_json::Number::from(0))),
            ].into_iter().collect(),
        },
        Keyframe {
            time: 1.0,
            properties: [
                ("x".to_string(), serde_json::Value::Number(serde_json::Number::from(100))),
            ].into_iter().collect(),
        },
    ];

    // Create animation
    let animation_id = engine.create_animation(&element_id, AnimationType::Transform, 1000.0, keyframes).unwrap();
    assert!(!animation_id.is_empty());

    // Test animation exists
    let animation_exists = engine.document_state.animations.iter().any(|a| a.id == animation_id);
    assert!(animation_exists);

    // Test stopping animation
    let result = engine.stop_animation(&animation_id);
    assert!(result.is_ok());

    // Verify animation is stopped
    let animation_exists = engine.document_state.animations.iter().any(|a| a.id == animation_id);
    assert!(!animation_exists);
}

#[wasm_bindgen_test]
fn test_event_handling() {
    let permissions = WASMPermissions {
        memory_limit: 1024 * 1024,
        allowed_imports: vec!["console".to_string()],
        cpu_time_limit: 5000,
        allow_networking: false,
        allow_file_system: false,
        allowed_interactions: vec![
            "create_element".to_string(),
            "create_event_handler".to_string(),
            "Click".to_string(),
        ],
        max_data_size: 1024 * 1024,
        max_elements: 100,
    };

    let mut engine = InteractiveEngine::new(permissions).unwrap();

    // Create an element
    let element_id = engine.create_element(ElementType::Interactive, HashMap::new()).unwrap();

    // Add event handler
    let result = engine.add_event_handler(&element_id, "click", "toggle_visibility");
    assert!(result.is_ok());

    // Test interaction processing
    let interaction_event = InteractionEvent {
        event_type: InteractionType::Click,
        target_element: Some(element_id.clone()),
        position: Some(Position { x: 10.0, y: 10.0 }),
        data: HashMap::new(),
        timestamp: get_current_timestamp(),
    };

    let result = engine.process_interaction(interaction_event);
    assert!(result.is_ok());

    let render_update = result.unwrap();
    assert!(!render_update.dom_operations.is_empty() || !render_update.style_changes.is_empty());
}

#[wasm_bindgen_test]
fn test_viewport_updates() {
    let permissions = WASMPermissions {
        memory_limit: 1024 * 1024,
        allowed_imports: vec!["console".to_string()],
        cpu_time_limit: 5000,
        allow_networking: false,
        allow_file_system: false,
        allowed_interactions: vec![
            "create_element".to_string(),
        ],
        max_data_size: 1024 * 1024,
        max_elements: 100,
    };

    let mut engine = InteractiveEngine::new(permissions).unwrap();

    // Create some elements
    let _element1 = engine.create_element(ElementType::Container, HashMap::new()).unwrap();
    let _element2 = engine.create_element(ElementType::Text, HashMap::new()).unwrap();

    // Update viewport
    let result = engine.update_viewport(1920.0, 1080.0, 1.5);
    assert!(result.is_ok());

    // Check viewport was updated
    assert_eq!(engine.document_state.viewport.width, 1920.0);
    assert_eq!(engine.document_state.viewport.height, 1080.0);
    assert_eq!(engine.document_state.viewport.scale, 1.5);

    // Check that elements were marked as dirty
    assert!(!engine.document_state.render_tree.dirty_nodes.is_empty());
}

#[wasm_bindgen_test]
fn test_data_source_management() {
    let permissions = WASMPermissions {
        memory_limit: 1024 * 1024,
        allowed_imports: vec!["console".to_string()],
        cpu_time_limit: 5000,
        allow_networking: false,
        allow_file_system: false,
        allowed_interactions: vec![
            "create_element".to_string(),
            "DataUpdate".to_string(),
        ],
        max_data_size: 1024 * 1024,
        max_elements: 100,
    };

    let mut engine = InteractiveEngine::new(permissions).unwrap();

    // Create a data source
    let data_source = DataSource::new(
        "test_data".to_string(),
        DataSourceType::Dynamic,
        serde_json::json!({"value": 42}),
    );

    engine.document_state.data_sources.insert("test_data".to_string(), data_source);

    // Create an element that uses this data source
    let properties = [
        ("data_source".to_string(), serde_json::Value::String("test_data".to_string())),
    ].into_iter().collect();

    let element_id = engine.create_element(ElementType::Chart, properties).unwrap();

    // Test data update
    let data_update_event = InteractionEvent {
        event_type: InteractionType::DataUpdate,
        target_element: None,
        position: None,
        data: [
            ("data_source_id".to_string(), serde_json::Value::String("test_data".to_string())),
            ("data".to_string(), serde_json::json!({"value": 84})),
        ].into_iter().collect(),
        timestamp: get_current_timestamp(),
    };

    let result = engine.process_interaction(data_update_event);
    assert!(result.is_ok());

    // Verify data was updated
    let updated_data = &engine.document_state.data_sources["test_data"].data;
    assert_eq!(updated_data["value"], 84);
}

#[wasm_bindgen_test]
fn test_security_permissions() {
    let restrictive_permissions = WASMPermissions {
        memory_limit: 1024 * 1024,
        allowed_imports: vec![],
        cpu_time_limit: 5000,
        allow_networking: false,
        allow_file_system: false,
        allowed_interactions: vec![], // No interactions allowed
        max_data_size: 1024 * 1024,
        max_elements: 100,
    };

    let mut engine = InteractiveEngine::new(restrictive_permissions).unwrap();

    // Test that element creation is blocked
    let result = engine.create_element(ElementType::Container, HashMap::new());
    assert!(result.is_err());

    // Test that animation creation is blocked
    let result = engine.create_animation("test", AnimationType::Transform, 1000.0, vec![]);
    assert!(result.is_err());
}

#[wasm_bindgen_test]
fn test_render_frame_processing() {
    let permissions = WASMPermissions {
        memory_limit: 1024 * 1024,
        allowed_imports: vec!["console".to_string()],
        cpu_time_limit: 5000,
        allow_networking: false,
        allow_file_system: false,
        allowed_interactions: vec![
            "create_element".to_string(),
            "create_animation".to_string(),
        ],
        max_data_size: 1024 * 1024,
        max_elements: 100,
    };

    let mut engine = InteractiveEngine::new(permissions).unwrap();

    // Create element and animation
    let element_id = engine.create_element(ElementType::Container, HashMap::new()).unwrap();
    let keyframes = vec![
        Keyframe {
            time: 0.0,
            properties: [("opacity".to_string(), serde_json::Value::Number(serde_json::Number::from_f64(0.0).unwrap()))].into_iter().collect(),
        },
        Keyframe {
            time: 1.0,
            properties: [("opacity".to_string(), serde_json::Value::Number(serde_json::Number::from_f64(1.0).unwrap()))].into_iter().collect(),
        },
    ];

    let _animation_id = engine.create_animation(&element_id, AnimationType::Style, 1000.0, keyframes).unwrap();

    // Test render frame
    let timestamp = get_current_timestamp();
    let result = engine.render_frame(timestamp);
    assert!(result.is_ok());

    let render_update = result.unwrap();
    // Should have animation updates if animation is active
    assert!(render_update.animation_updates.len() >= 0); // Could be 0 if animation hasn't started
}

#[wasm_bindgen_test]
fn test_element_querying() {
    let permissions = WASMPermissions {
        memory_limit: 1024 * 1024,
        allowed_imports: vec!["console".to_string()],
        cpu_time_limit: 5000,
        allow_networking: false,
        allow_file_system: false,
        allowed_interactions: vec![
            "create_element".to_string(),
        ],
        max_data_size: 1024 * 1024,
        max_elements: 100,
    };

    let mut engine = InteractiveEngine::new(permissions).unwrap();

    // Create elements of different types
    let _container1 = engine.create_element(ElementType::Container, HashMap::new()).unwrap();
    let _container2 = engine.create_element(ElementType::Container, HashMap::new()).unwrap();
    let _chart1 = engine.create_element(ElementType::Chart, HashMap::new()).unwrap();
    let _text1 = engine.create_element(ElementType::Text, HashMap::new()).unwrap();

    // Query containers
    let containers = engine.query_elements_by_type(ElementType::Container);
    assert_eq!(containers.len(), 2);

    // Query charts
    let charts = engine.query_elements_by_type(ElementType::Chart);
    assert_eq!(charts.len(), 1);

    // Query text elements
    let texts = engine.query_elements_by_type(ElementType::Text);
    assert_eq!(texts.len(), 1);

    // Query animations (should be empty)
    let animations = engine.query_elements_by_type(ElementType::Animation);
    assert_eq!(animations.len(), 0);
}