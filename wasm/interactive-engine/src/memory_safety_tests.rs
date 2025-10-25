use super::*;
use wasm_bindgen_test::*;

wasm_bindgen_test_configure!(run_in_browser);

// Test memory safety and resource limit compliance

#[wasm_bindgen_test]
fn test_memory_safety_with_concurrent_operations() {
    let permissions = WASMPermissions {
        memory_limit: 2 * 1024 * 1024, // 2MB
        allowed_imports: vec!["console".to_string()],
        cpu_time_limit: 10000,
        allow_networking: false,
        allow_file_system: false,
        allowed_interactions: vec![
            "create_element".to_string(),
            "modify_element".to_string(),
            "create_animation".to_string(),
            "Click".to_string(),
        ],
        max_data_size: 1024 * 1024, // 1MB
        max_elements: 100,
    };

    let mut engine = InteractiveEngine::new(permissions).unwrap();

    // Test concurrent element creation and modification
    let mut element_ids = Vec::new();
    
    // Create elements
    for i in 0..50 {
        let properties = [
            ("width".to_string(), serde_json::Value::Number(serde_json::Number::from(100))),
            ("height".to_string(), serde_json::Value::Number(serde_json::Number::from(100))),
            ("id".to_string(), serde_json::Value::String(format!("element_{}", i))),
        ].into_iter().collect();

        match engine.create_element(ElementType::Container, properties) {
            Ok(element_id) => {
                element_ids.push(element_id);
            }
            Err(_) => {
                // Expected when hitting limits
                break;
            }
        }
    }

    // Modify elements concurrently
    for (i, element_id) in element_ids.iter().enumerate() {
        let update_properties = [
            ("color".to_string(), serde_json::Value::String(format!("color_{}", i))),
            ("data".to_string(), serde_json::Value::String("x".repeat(100))),
        ].into_iter().collect();

        let result = engine.update_element_properties(element_id, update_properties);
        // Should succeed for valid elements
        assert!(result.is_ok());
    }

    // Create animations for elements
    for (i, element_id) in element_ids.iter().enumerate().take(20) {
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

        let result = engine.create_animation(element_id, AnimationType::Transform, 1000.0, keyframes);
        // Should succeed within memory limits
        if result.is_err() {
            // Expected when hitting memory limits
            break;
        }
    }

    // Process interactions on elements
    for (i, element_id) in element_ids.iter().enumerate().take(10) {
        let interaction_event = InteractionEvent {
            event_type: InteractionType::Click,
            target_element: Some(element_id.clone()),
            position: Some(Position { x: (i % 10) as f64 * 10.0, y: (i / 10) as f64 * 10.0 }),
            data: HashMap::new(),
            timestamp: get_current_timestamp() + (i as f64 * 10.0),
            touch_data: None,
            mouse_data: Some(MouseData {
                button: MouseButton::Left,
                buttons: 1,
                position: Position { x: (i % 10) as f64 * 10.0, y: (i / 10) as f64 * 10.0 },
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

        let result = engine.process_interaction(interaction_event);
        assert!(result.is_ok());
    }

    // Verify all elements are still accessible (no memory corruption)
    for element_id in &element_ids {
        let element = engine.document_state.get_element(element_id);
        assert!(element.is_some(), "Element {} should still exist", element_id);
    }

    // Verify memory usage is within bounds
    assert!(engine.security_context.allocated_memory <= engine.security_context.resource_limits.max_memory);
}

#[wasm_bindgen_test]
fn test_resource_cleanup_on_element_deletion() {
    let permissions = WASMPermissions {
        memory_limit: 5 * 1024 * 1024,
        allowed_imports: vec!["console".to_string()],
        cpu_time_limit: 10000,
        allow_networking: false,
        allow_file_system: false,
        allowed_interactions: vec![
            "create_element".to_string(),
            "modify_element".to_string(),
            "create_animation".to_string(),
        ],
        max_data_size: 2 * 1024 * 1024,
        max_elements: 200,
    };

    let mut engine = InteractiveEngine::new(permissions).unwrap();

    // Create elements with associated resources
    let mut element_ids = Vec::new();
    let mut animation_ids = Vec::new();

    for i in 0..20 {
        // Create element with data
        let properties = [
            ("width".to_string(), serde_json::Value::Number(serde_json::Number::from(100))),
            ("height".to_string(), serde_json::Value::Number(serde_json::Number::from(100))),
            ("data".to_string(), serde_json::Value::String("x".repeat(1000))), // 1KB per element
        ].into_iter().collect();

        let element_id = engine.create_element(ElementType::Container, properties).unwrap();
        element_ids.push(element_id.clone());

        // Create animation for element
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

        let animation_id = engine.create_animation(&element_id, AnimationType::Style, 2000.0, keyframes).unwrap();
        animation_ids.push(animation_id);
    }

    let initial_memory = engine.security_context.allocated_memory;
    let initial_element_count = engine.document_state.elements.len();
    let initial_animation_count = engine.document_state.animations.len();

    // Delete half of the elements
    for i in 0..10 {
        let element_id = &element_ids[i];
        let result = engine.delete_element(element_id);
        assert!(result.is_ok(), "Failed to delete element {}", element_id);
    }

    // Verify elements were deleted
    assert_eq!(engine.document_state.elements.len(), initial_element_count - 10);

    // Verify associated animations were cleaned up
    assert!(engine.document_state.animations.len() < initial_animation_count);

    // Verify memory was freed (should be less than initial)
    let final_memory = engine.security_context.allocated_memory;
    assert!(final_memory < initial_memory, "Memory should be freed after deletion");

    // Verify remaining elements are still valid
    for i in 10..20 {
        let element_id = &element_ids[i];
        let element = engine.document_state.get_element(element_id);
        assert!(element.is_some(), "Remaining element {} should still exist", element_id);
    }
}

#[wasm_bindgen_test]
fn test_memory_bounds_checking() {
    let strict_permissions = WASMPermissions {
        memory_limit: 512 * 1024, // 512KB - very small
        allowed_imports: vec!["console".to_string()],
        cpu_time_limit: 5000,
        allow_networking: false,
        allow_file_system: false,
        allowed_interactions: vec![
            "create_element".to_string(),
            "modify_element".to_string(),
        ],
        max_data_size: 256 * 1024, // 256KB
        max_elements: 20,
    };

    let mut engine = InteractiveEngine::new(strict_permissions).unwrap();

    // Try to allocate memory beyond limits
    let mut successful_allocations = 0;
    let mut allocation_failed = false;

    for i in 0..50 {
        let large_data = "x".repeat(10 * 1024); // 10KB per element
        let properties = [
            ("data".to_string(), serde_json::Value::String(large_data)),
            ("id".to_string(), serde_json::Value::String(format!("large_element_{}", i))),
        ].into_iter().collect();

        match engine.create_element(ElementType::Container, properties) {
            Ok(_) => {
                successful_allocations += 1;
            }
            Err(e) => {
                allocation_failed = true;
                // Should fail due to memory limits
                assert!(e.code == "ELEMENT_CREATION_NOT_ALLOWED" || 
                       e.message.contains("memory") || 
                       e.message.contains("limit"));
                break;
            }
        }
    }

    // Should have failed before allocating all 50 elements
    assert!(allocation_failed || successful_allocations < 50);
    assert!(successful_allocations <= 20); // Max elements limit

    // Verify memory usage is within bounds
    assert!(engine.security_context.allocated_memory <= engine.security_context.resource_limits.max_memory);
}

#[wasm_bindgen_test]
fn test_data_structure_integrity_under_stress() {
    let permissions = WASMPermissions {
        memory_limit: 10 * 1024 * 1024,
        allowed_imports: vec!["console".to_string()],
        cpu_time_limit: 15000,
        allow_networking: false,
        allow_file_system: false,
        allowed_interactions: vec![
            "create_element".to_string(),
            "modify_element".to_string(),
            "create_animation".to_string(),
            "Click".to_string(),
            "TouchStart".to_string(),
        ],
        max_data_size: 5 * 1024 * 1024,
        max_elements: 500,
    };

    let mut engine = InteractiveEngine::new(permissions).unwrap();

    // Create a complex document structure
    let mut element_ids = Vec::new();
    
    // Create containers
    for i in 0..50 {
        let element_id = engine.create_element(ElementType::Container, HashMap::new()).unwrap();
        element_ids.push(element_id);
    }

    // Create interactive elements
    for i in 0..50 {
        let element_id = engine.create_element(ElementType::Interactive, HashMap::new()).unwrap();
        element_ids.push(element_id);
    }

    // Create charts
    for i in 0..20 {
        let element_id = engine.create_element(ElementType::Chart, HashMap::new()).unwrap();
        element_ids.push(element_id);
    }

    // Perform rapid modifications
    for iteration in 0..100 {
        let element_id = &element_ids[iteration % element_ids.len()];
        
        let update_properties = [
            ("iteration".to_string(), serde_json::Value::Number(serde_json::Number::from(iteration))),
            ("timestamp".to_string(), serde_json::Value::Number(serde_json::Number::from_f64(get_current_timestamp()).unwrap())),
        ].into_iter().collect();

        let result = engine.update_element_properties(element_id, update_properties);
        assert!(result.is_ok(), "Element update failed at iteration {}", iteration);
    }

    // Perform rapid interactions
    for iteration in 0..200 {
        let element_id = &element_ids[iteration % element_ids.len()];
        
        let interaction_event = InteractionEvent {
            event_type: if iteration % 2 == 0 { InteractionType::Click } else { InteractionType::TouchStart },
            target_element: Some(element_id.clone()),
            position: Some(Position { x: (iteration % 100) as f64, y: (iteration / 100) as f64 }),
            data: HashMap::new(),
            timestamp: get_current_timestamp() + (iteration as f64),
            touch_data: if iteration % 2 == 1 {
                Some(TouchData {
                    touches: vec![TouchPoint {
                        identifier: 1,
                        position: Position { x: (iteration % 100) as f64, y: (iteration / 100) as f64 },
                        radius: Some(8.0),
                        rotation_angle: None,
                        force: Some(0.5),
                    }],
                    changed_touches: vec![],
                    target_touches: vec![],
                    force: Some(0.5),
                    rotation_angle: None,
                    scale: None,
                })
            } else {
                None
            },
            mouse_data: if iteration % 2 == 0 {
                Some(MouseData {
                    button: MouseButton::Left,
                    buttons: 1,
                    position: Position { x: (iteration % 100) as f64, y: (iteration / 100) as f64 },
                    movement: None,
                    wheel_delta: None,
                })
            } else {
                None
            },
            keyboard_data: None,
            gesture_data: None,
            modifiers: EventModifiers {
                ctrl: false,
                shift: false,
                alt: false,
                meta: false,
            },
        };

        let result = engine.process_interaction(interaction_event);
        assert!(result.is_ok(), "Interaction processing failed at iteration {}", iteration);
    }

    // Verify data structure integrity
    assert_eq!(engine.document_state.elements.len(), element_ids.len());
    
    // Verify all elements are still accessible and valid
    for element_id in &element_ids {
        let element = engine.document_state.get_element(element_id);
        assert!(element.is_some(), "Element {} should still exist after stress test", element_id);
        
        let element = element.unwrap();
        assert_eq!(element.id, *element_id);
        assert!(!element.properties.is_empty() || element.element_type != ElementType::Container);
    }

    // Verify render tree consistency
    for element_id in &element_ids {
        if let Some(render_node) = engine.document_state.render_tree.nodes.get(element_id) {
            assert_eq!(render_node.element_id, *element_id);
        }
    }
}

#[wasm_bindgen_test]
fn test_animation_memory_safety() {
    let permissions = WASMPermissions {
        memory_limit: 5 * 1024 * 1024,
        allowed_imports: vec!["console".to_string()],
        cpu_time_limit: 10000,
        allow_networking: false,
        allow_file_system: false,
        allowed_interactions: vec![
            "create_element".to_string(),
            "create_animation".to_string(),
        ],
        max_data_size: 2 * 1024 * 1024,
        max_elements: 200,
    };

    let mut engine = InteractiveEngine::new(permissions).unwrap();

    // Create elements for animation
    let mut element_ids = Vec::new();
    for i in 0..30 {
        let element_id = engine.create_element(ElementType::Container, HashMap::new()).unwrap();
        element_ids.push(element_id);
    }

    // Create many animations with complex keyframes
    let mut animation_ids = Vec::new();
    for (i, element_id) in element_ids.iter().enumerate() {
        let keyframes = vec![
            Keyframe {
                time: 0.0,
                properties: [
                    ("x".to_string(), serde_json::Value::Number(serde_json::Number::from(0))),
                    ("y".to_string(), serde_json::Value::Number(serde_json::Number::from(0))),
                    ("opacity".to_string(), serde_json::Value::Number(serde_json::Number::from_f64(0.0).unwrap())),
                    ("scale".to_string(), serde_json::Value::Number(serde_json::Number::from_f64(1.0).unwrap())),
                ].into_iter().collect(),
            },
            Keyframe {
                time: 0.25,
                properties: [
                    ("x".to_string(), serde_json::Value::Number(serde_json::Number::from(25))),
                    ("y".to_string(), serde_json::Value::Number(serde_json::Number::from(12))),
                    ("opacity".to_string(), serde_json::Value::Number(serde_json::Number::from_f64(0.5).unwrap())),
                    ("scale".to_string(), serde_json::Value::Number(serde_json::Number::from_f64(1.2).unwrap())),
                ].into_iter().collect(),
            },
            Keyframe {
                time: 0.75,
                properties: [
                    ("x".to_string(), serde_json::Value::Number(serde_json::Number::from(75))),
                    ("y".to_string(), serde_json::Value::Number(serde_json::Number::from(37))),
                    ("opacity".to_string(), serde_json::Value::Number(serde_json::Number::from_f64(0.8).unwrap())),
                    ("scale".to_string(), serde_json::Value::Number(serde_json::Number::from_f64(0.9).unwrap())),
                ].into_iter().collect(),
            },
            Keyframe {
                time: 1.0,
                properties: [
                    ("x".to_string(), serde_json::Value::Number(serde_json::Number::from(100))),
                    ("y".to_string(), serde_json::Value::Number(serde_json::Number::from(50))),
                    ("opacity".to_string(), serde_json::Value::Number(serde_json::Number::from_f64(1.0).unwrap())),
                    ("scale".to_string(), serde_json::Value::Number(serde_json::Number::from_f64(1.0).unwrap())),
                ].into_iter().collect(),
            },
        ];

        match engine.create_animation(element_id, AnimationType::Transform, 2000.0, keyframes) {
            Ok(animation_id) => {
                animation_ids.push(animation_id);
            }
            Err(_) => {
                // Expected when hitting memory limits
                break;
            }
        }
    }

    let initial_memory = engine.security_context.allocated_memory;

    // Process many animation frames
    let start_time = get_current_timestamp();
    for frame in 0..240 { // 4 seconds at 60fps
        let timestamp = start_time + (frame as f64 * 16.67);
        let result = engine.render_frame(timestamp);
        assert!(result.is_ok(), "Animation frame processing failed at frame {}", frame);
    }

    // Stop all animations
    for animation_id in &animation_ids {
        let result = engine.stop_animation(animation_id);
        assert!(result.is_ok(), "Failed to stop animation {}", animation_id);
    }

    // Verify animations were cleaned up
    assert_eq!(engine.document_state.animations.len(), 0);

    // Verify memory was freed
    let final_memory = engine.security_context.allocated_memory;
    assert!(final_memory <= initial_memory, "Memory should not increase after animation cleanup");

    // Verify elements are still intact
    for element_id in &element_ids {
        let element = engine.document_state.get_element(element_id);
        assert!(element.is_some(), "Element {} should still exist after animation cleanup", element_id);
    }
}

#[wasm_bindgen_test]
fn test_chart_memory_safety() {
    let permissions = WASMPermissions {
        memory_limit: 10 * 1024 * 1024,
        allowed_imports: vec!["console".to_string()],
        cpu_time_limit: 15000,
        allow_networking: false,
        allow_file_system: false,
        allowed_interactions: vec![
            "create_element".to_string(),
            "modify_element".to_string(),
        ],
        max_data_size: 5 * 1024 * 1024,
        max_elements: 300,
    };

    let mut engine = InteractiveEngine::new(permissions).unwrap();

    // Create multiple charts with large datasets
    let mut chart_ids = Vec::new();
    
    for i in 0..10 {
        let config = ChartConfig {
            width: 800.0,
            height: 600.0,
            ..ChartConfig::default()
        };

        match engine.chart_renderer.create_chart(
            ChartType::Line,
            format!("memory_test_data_{}", i),
            config
        ) {
            Ok(chart_id) => {
                chart_ids.push(chart_id);
            }
            Err(_) => {
                // Expected when hitting limits
                break;
            }
        }
    }

    // Add series to charts
    for (i, chart_id) in chart_ids.iter().enumerate() {
        let series = ChartSeries {
            id: format!("memory_series_{}", i),
            name: format!("Memory Test Series {}", i),
            data_field: "value".to_string(),
            color: format!("#{:06x}", i * 123456 % 0xFFFFFF),
            line_width: Some(2.0),
            fill_opacity: Some(0.3),
            marker_size: Some(4.0),
            marker_shape: Some(MarkerShape::Circle),
            visible: true,
            y_axis: AxisReference::Primary,
        };

        let result = engine.chart_renderer.add_series(chart_id, series);
        assert!(result.is_ok(), "Failed to add series to chart {}", chart_id);
    }

    let initial_memory = engine.security_context.allocated_memory;

    // Render charts with large datasets multiple times
    for iteration in 0..20 {
        // Generate large dataset
        let mut large_dataset = Vec::new();
        for j in 0..1000 {
            large_dataset.push(serde_json::json!({
                "value": ((iteration * 1000 + j) as f64 * 0.01).sin() * 100.0 + 50.0,
                "x": j as f64,
                "label": format!("Point {}_{}", iteration, j),
                "metadata": {
                    "iteration": iteration,
                    "index": j,
                    "timestamp": get_current_timestamp()
                }
            }));
        }
        let test_data = serde_json::Value::Array(large_dataset);

        // Render all charts
        for chart_id in &chart_ids {
            let result = engine.chart_renderer.render_chart(chart_id, &test_data);
            assert!(result.is_ok(), "Chart rendering failed for {} at iteration {}", chart_id, iteration);
            
            let rendered_chart = result.unwrap();
            assert_eq!(rendered_chart.data_points.len(), 1000);
            assert!(!rendered_chart.svg_content.is_empty());
        }

        // Update chart data
        for chart_id in &chart_ids {
            let result = engine.chart_renderer.update_chart_data(chart_id, &test_data);
            assert!(result.is_ok(), "Chart data update failed for {} at iteration {}", chart_id, iteration);
        }
    }

    // Verify memory usage is reasonable
    let final_memory = engine.security_context.allocated_memory;
    let memory_growth = final_memory - initial_memory;
    
    // Memory should not grow excessively (allow for some caching)
    assert!(memory_growth < 5 * 1024 * 1024, "Memory growth too high: {} bytes", memory_growth);
    assert!(final_memory <= engine.security_context.resource_limits.max_memory);

    // Verify chart performance stats
    let performance_stats = &engine.chart_renderer.performance_stats;
    assert!(performance_stats.total_charts > 0);
    assert!(performance_stats.total_render_time > 0.0);
    assert!(performance_stats.average_render_time > 0.0);
}

#[wasm_bindgen_test]
fn test_resource_limit_compliance_under_load() {
    let strict_permissions = WASMPermissions {
        memory_limit: 1024 * 1024, // 1MB
        allowed_imports: vec!["console".to_string()],
        cpu_time_limit: 2000, // 2 seconds
        allow_networking: false,
        allow_file_system: false,
        allowed_interactions: vec![
            "create_element".to_string(),
            "modify_element".to_string(),
            "create_animation".to_string(),
            "Click".to_string(),
        ],
        max_data_size: 512 * 1024, // 512KB
        max_elements: 50,
    };

    let mut engine = InteractiveEngine::new(strict_permissions).unwrap();

    // Continuously perform operations until limits are hit
    let mut operations_performed = 0;
    let mut limit_hit = false;

    // Create elements until limit
    for i in 0..100 {
        let properties = [
            ("data".to_string(), serde_json::Value::String("x".repeat(1000))), // 1KB per element
        ].into_iter().collect();

        match engine.create_element(ElementType::Container, properties) {
            Ok(_) => {
                operations_performed += 1;
            }
            Err(_) => {
                limit_hit = true;
                break;
            }
        }
    }

    assert!(limit_hit, "Should have hit element or memory limit");
    assert!(operations_performed <= 50, "Should not exceed max_elements limit");

    // Verify memory compliance
    assert!(engine.security_context.allocated_memory <= engine.security_context.resource_limits.max_memory);

    // Try to perform CPU-intensive operations
    let start_time = get_current_timestamp();
    let mut cpu_limit_hit = false;

    for frame in 0..1000 {
        let timestamp = start_time + (frame as f64 * 16.67);
        match engine.render_frame(timestamp) {
            Ok(_) => {
                // Continue
            }
            Err(e) => {
                if e.code == "CPU_TIME_EXCEEDED" {
                    cpu_limit_hit = true;
                    break;
                }
            }
        }
    }

    // Should eventually hit CPU time limit or complete within reasonable time
    let elapsed_time = get_current_timestamp() - start_time;
    assert!(cpu_limit_hit || elapsed_time < 5000.0, "Should hit CPU limit or complete quickly");
}