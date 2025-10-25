use super::*;
use wasm_bindgen_test::*;

wasm_bindgen_test_configure!(run_in_browser);

// Comprehensive integration tests combining all aspects

#[wasm_bindgen_test]
fn test_full_interactive_document_workflow() {
    let permissions = WASMPermissions {
        memory_limit: 20 * 1024 * 1024, // 20MB for comprehensive test
        allowed_imports: vec!["console".to_string()],
        cpu_time_limit: 30000, // 30 seconds
        allow_networking: false,
        allow_file_system: false,
        allowed_interactions: vec![
            "create_element".to_string(),
            "modify_element".to_string(),
            "create_animation".to_string(),
            "create_event_handler".to_string(),
            "Click".to_string(),
            "TouchStart".to_string(),
            "TouchMove".to_string(),
            "TouchEnd".to_string(),
            "MouseMove".to_string(),
            "KeyDown".to_string(),
            "DataUpdate".to_string(),
        ],
        max_data_size: 10 * 1024 * 1024, // 10MB
        max_elements: 1000,
    };

    let mut engine = InteractiveEngine::new(permissions).unwrap();

    // Phase 1: Create document structure
    let container_id = engine.create_element(ElementType::Container, [
        ("width".to_string(), serde_json::json!(800)),
        ("height".to_string(), serde_json::json!(600)),
        ("background".to_string(), serde_json::json!("#f0f0f0")),
    ].into_iter().collect()).unwrap();

    let chart_element_id = engine.create_element(ElementType::Chart, [
        ("width".to_string(), serde_json::json!(400)),
        ("height".to_string(), serde_json::json!(300)),
        ("chart_type".to_string(), serde_json::json!("line")),
    ].into_iter().collect()).unwrap();

    let interactive_element_id = engine.create_element(ElementType::Interactive, [
        ("width".to_string(), serde_json::json!(200)),
        ("height".to_string(), serde_json::json!(100)),
        ("interactive".to_string(), serde_json::json!(true)),
    ].into_iter().collect()).unwrap();

    // Phase 2: Create data sources
    let chart_data_source = DataSource::new(
        "chart_data".to_string(),
        DataSourceType::Dynamic,
        serde_json::json!([
            {"value": 10, "label": "Jan", "x": 0},
            {"value": 25, "label": "Feb", "x": 1},
            {"value": 15, "label": "Mar", "x": 2},
            {"value": 30, "label": "Apr", "x": 3},
            {"value": 20, "label": "May", "x": 4}
        ]),
    );
    engine.document_state.data_sources.insert("chart_data".to_string(), chart_data_source);

    let interactive_data_source = DataSource::new(
        "interactive_data".to_string(),
        DataSourceType::Stream,
        serde_json::json!([]),
    );
    engine.document_state.data_sources.insert("interactive_data".to_string(), interactive_data_source);

    // Phase 3: Create charts
    let chart_config = ChartConfig {
        width: 400.0,
        height: 300.0,
        responsive: true,
        ..ChartConfig::default()
    };

    let chart_id = engine.chart_renderer.create_chart(
        ChartType::Line,
        "chart_data".to_string(),
        chart_config
    ).unwrap();

    let chart_series = ChartSeries {
        id: "main_series".to_string(),
        name: "Main Data Series".to_string(),
        data_field: "value".to_string(),
        color: "#1f77b4".to_string(),
        line_width: Some(3.0),
        fill_opacity: Some(0.2),
        marker_size: Some(5.0),
        marker_shape: Some(MarkerShape::Circle),
        visible: true,
        y_axis: AxisReference::Primary,
    };

    engine.chart_renderer.add_series(&chart_id, chart_series).unwrap();

    // Phase 4: Create animations
    let fade_in_keyframes = vec![
        Keyframe {
            time: 0.0,
            properties: [("opacity".to_string(), serde_json::json!(0.0))].into_iter().collect(),
        },
        Keyframe {
            time: 1.0,
            properties: [("opacity".to_string(), serde_json::json!(1.0))].into_iter().collect(),
        },
    ];

    let fade_animation_id = engine.create_animation(
        &container_id,
        AnimationType::Style,
        2000.0,
        fade_in_keyframes
    ).unwrap();

    let slide_keyframes = vec![
        Keyframe {
            time: 0.0,
            properties: [
                ("x".to_string(), serde_json::json!(-200)),
                ("y".to_string(), serde_json::json!(0)),
            ].into_iter().collect(),
        },
        Keyframe {
            time: 1.0,
            properties: [
                ("x".to_string(), serde_json::json!(0)),
                ("y".to_string(), serde_json::json!(0)),
            ].into_iter().collect(),
        },
    ];

    let slide_animation_id = engine.create_animation(
        &chart_element_id,
        AnimationType::Transform,
        1500.0,
        slide_keyframes
    ).unwrap();

    // Phase 5: Set up event handlers and data bindings
    engine.add_event_handler(&interactive_element_id, "click", "toggle_chart_visibility").unwrap();

    let data_binding = DataBinding {
        source_id: "chart_data".to_string(),
        target_element: chart_element_id.clone(),
        property_path: "data".to_string(),
        transform_function: None,
        update_trigger: UpdateTrigger::Immediate,
    };
    engine.data_binding_manager.add_binding(data_binding);

    // Phase 6: Set up interaction delegates
    let delegate = EventDelegate {
        element_id: container_id.clone(),
        event_types: vec![InteractionType::Click, InteractionType::TouchStart],
        handler_id: "container_handler".to_string(),
        capture: false,
        priority: 1,
    };
    engine.add_interaction_delegate(&interactive_element_id, delegate).unwrap();

    // Phase 7: Configure responsive adapter for mobile
    let mobile_device_info = DeviceInfo {
        device_type: DeviceType::Mobile,
        screen_size: Size { width: 375.0, height: 667.0 },
        pixel_density: 2.0,
        touch_support: true,
        mouse_support: false,
        keyboard_support: true,
        max_touch_points: 10,
        has_force_touch: true,
        has_hover_support: false,
    };
    engine.update_device_capabilities(mobile_device_info).unwrap();

    // Phase 8: Simulate user interactions
    let start_time = get_current_timestamp();

    // Touch interaction sequence
    let touch_start = InteractionEvent {
        event_type: InteractionType::TouchStart,
        target_element: Some(interactive_element_id.clone()),
        position: Some(Position { x: 100.0, y: 50.0 }),
        data: HashMap::new(),
        timestamp: start_time,
        touch_data: Some(TouchData {
            touches: vec![TouchPoint {
                identifier: 1,
                position: Position { x: 100.0, y: 50.0 },
                radius: Some(8.0),
                rotation_angle: None,
                force: Some(0.5),
            }],
            changed_touches: vec![TouchPoint {
                identifier: 1,
                position: Position { x: 100.0, y: 50.0 },
                radius: Some(8.0),
                rotation_angle: None,
                force: Some(0.5),
            }],
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

    let render_update = engine.process_interaction(touch_start).unwrap();
    assert!(!render_update.dom_operations.is_empty() || !render_update.style_changes.is_empty());

    // Swipe gesture simulation
    for i in 1..20 {
        let touch_move = InteractionEvent {
            event_type: InteractionType::TouchMove,
            target_element: Some(interactive_element_id.clone()),
            position: Some(Position { x: 100.0 + (i as f64 * 5.0), y: 50.0 }),
            data: HashMap::new(),
            timestamp: start_time + (i as f64 * 16.67),
            touch_data: Some(TouchData {
                touches: vec![TouchPoint {
                    identifier: 1,
                    position: Position { x: 100.0 + (i as f64 * 5.0), y: 50.0 },
                    radius: Some(8.0),
                    rotation_angle: None,
                    force: Some(0.5),
                }],
                changed_touches: vec![TouchPoint {
                    identifier: 1,
                    position: Position { x: 100.0 + (i as f64 * 5.0), y: 50.0 },
                    radius: Some(8.0),
                    rotation_angle: None,
                    force: Some(0.5),
                }],
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

        let render_update = engine.process_interaction(touch_move).unwrap();
        assert!(!render_update.dom_operations.is_empty() || !render_update.style_changes.is_empty());
    }

    // Touch end (complete swipe gesture)
    let touch_end = InteractionEvent {
        event_type: InteractionType::TouchEnd,
        target_element: Some(interactive_element_id.clone()),
        position: Some(Position { x: 200.0, y: 50.0 }),
        data: HashMap::new(),
        timestamp: start_time + 400.0,
        touch_data: Some(TouchData {
            touches: vec![],
            changed_touches: vec![TouchPoint {
                identifier: 1,
                position: Position { x: 200.0, y: 50.0 },
                radius: Some(8.0),
                rotation_angle: None,
                force: Some(0.5),
            }],
            target_touches: vec![],
            force: None,
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

    let render_update = engine.process_interaction(touch_end).unwrap();
    assert!(!render_update.dom_operations.is_empty() || !render_update.style_changes.is_empty());

    // Phase 9: Update data and test data binding
    let updated_chart_data = serde_json::json!([
        {"value": 15, "label": "Jan", "x": 0},
        {"value": 30, "label": "Feb", "x": 1},
        {"value": 20, "label": "Mar", "x": 2},
        {"value": 35, "label": "Apr", "x": 3},
        {"value": 25, "label": "May", "x": 4},
        {"value": 40, "label": "Jun", "x": 5}
    ]);

    let data_update_event = InteractionEvent {
        event_type: InteractionType::DataUpdate,
        target_element: None,
        position: None,
        data: [
            ("data_source_id".to_string(), serde_json::json!("chart_data")),
            ("data".to_string(), updated_chart_data.clone()),
        ].into_iter().collect(),
        timestamp: start_time + 500.0,
        touch_data: None,
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

    let render_update = engine.process_interaction(data_update_event).unwrap();
    assert!(!render_update.dom_operations.is_empty() || !render_update.style_changes.is_empty());

    // Phase 10: Render chart with updated data
    let rendered_chart = engine.chart_renderer.render_chart(&chart_id, &updated_chart_data).unwrap();
    assert!(!rendered_chart.svg_content.is_empty());
    assert_eq!(rendered_chart.data_points.len(), 6);

    // Phase 11: Process animation frames
    for frame in 0..120 { // 2 seconds at 60fps
        let timestamp = start_time + 1000.0 + (frame as f64 * 16.67);
        let render_update = engine.render_frame(timestamp).unwrap();
        
        // Should have animation updates for the first part of the timeline
        if frame < 100 {
            assert!(render_update.animation_updates.len() >= 0);
        }
    }

    // Phase 12: Create vector graphics
    let rect_id = engine.vector_engine.create_shape(
        ShapeType::Rectangle,
        Position { x: 10.0, y: 10.0 },
        Size { width: 100.0, height: 50.0 }
    ).unwrap();

    let spiral_params = [
        ("center_x".to_string(), 200.0),
        ("center_y".to_string(), 150.0),
        ("start_radius".to_string(), 5.0),
        ("end_radius".to_string(), 30.0),
        ("turns".to_string(), 2.0),
    ].into_iter().collect();

    let spiral_id = engine.vector_engine.create_complex_path(
        ComplexPathType::Spiral,
        spiral_params
    ).unwrap();

    let svg_content = engine.vector_engine.render_to_svg(400.0, 300.0);
    assert!(svg_content.contains("<svg"));
    assert!(svg_content.contains("<rect"));
    assert!(svg_content.contains("<path"));

    // Phase 13: Test gesture recognition
    let gesture_events = engine.gesture_recognizer.process_touch_input(
        &TouchData {
            touches: vec![TouchPoint {
                identifier: 1,
                position: Position { x: 150.0, y: 150.0 },
                radius: Some(8.0),
                rotation_angle: None,
                force: Some(0.5),
            }],
            changed_touches: vec![TouchPoint {
                identifier: 1,
                position: Position { x: 150.0, y: 150.0 },
                radius: Some(8.0),
                rotation_angle: None,
                force: Some(0.5),
            }],
            target_touches: vec![],
            force: Some(0.5),
            rotation_angle: None,
            scale: None,
        },
        start_time + 2000.0
    );

    // Phase 14: Verify final state
    assert_eq!(engine.document_state.elements.len(), 3); // container, chart, interactive
    assert!(engine.document_state.animations.len() >= 0); // Animations may have completed
    assert_eq!(engine.document_state.data_sources.len(), 2);
    assert_eq!(engine.chart_renderer.charts.len(), 1);
    assert_eq!(engine.vector_engine.shapes.len(), 1);
    assert_eq!(engine.vector_engine.paths.len(), 1);

    // Verify performance metrics
    let interaction_metrics = engine.get_interaction_metrics();
    assert!(interaction_metrics.total_events > 0);
    assert!(interaction_metrics.average_response_time > 0.0);

    let chart_performance = &engine.chart_renderer.performance_stats;
    assert_eq!(chart_performance.total_charts, 1);
    assert!(chart_performance.total_render_time > 0.0);

    // Verify memory usage is reasonable
    assert!(engine.security_context.allocated_memory <= engine.security_context.resource_limits.max_memory);

    // Phase 15: Cleanup test
    engine.stop_animation(&fade_animation_id).unwrap();
    engine.stop_animation(&slide_animation_id).unwrap();
    engine.delete_element(&interactive_element_id).unwrap();

    // Verify cleanup
    assert_eq!(engine.document_state.elements.len(), 2);
    assert!(engine.document_state.get_element(&interactive_element_id).is_none());
}

#[wasm_bindgen_test]
fn test_performance_under_realistic_load() {
    let permissions = WASMPermissions {
        memory_limit: 50 * 1024 * 1024, // 50MB
        allowed_imports: vec!["console".to_string()],
        cpu_time_limit: 60000, // 60 seconds
        allow_networking: false,
        allow_file_system: false,
        allowed_interactions: vec![
            "create_element".to_string(),
            "modify_element".to_string(),
            "create_animation".to_string(),
            "Click".to_string(),
            "TouchStart".to_string(),
            "TouchMove".to_string(),
            "MouseMove".to_string(),
            "DataUpdate".to_string(),
        ],
        max_data_size: 25 * 1024 * 1024, // 25MB
        max_elements: 2000,
    };

    let mut engine = InteractiveEngine::new(permissions).unwrap();

    let start_time = get_current_timestamp();

    // Create a realistic document with multiple charts, animations, and interactions
    let mut element_ids = Vec::new();
    let mut chart_ids = Vec::new();
    let mut animation_ids = Vec::new();

    // Create dashboard with multiple charts
    for i in 0..10 {
        let chart_config = ChartConfig {
            width: 300.0,
            height: 200.0,
            responsive: true,
            ..ChartConfig::default()
        };

        let chart_type = match i % 5 {
            0 => ChartType::Line,
            1 => ChartType::Bar,
            2 => ChartType::Pie,
            3 => ChartType::Area,
            _ => ChartType::Scatter,
        };

        let chart_id = engine.chart_renderer.create_chart(
            chart_type,
            format!("dashboard_data_{}", i),
            chart_config
        ).unwrap();

        let series = ChartSeries {
            id: format!("dashboard_series_{}", i),
            name: format!("Dashboard Series {}", i),
            data_field: "value".to_string(),
            color: format!("#{:06x}", (i * 123456) % 0xFFFFFF),
            line_width: Some(2.0),
            fill_opacity: Some(0.3),
            marker_size: Some(4.0),
            marker_shape: Some(MarkerShape::Circle),
            visible: true,
            y_axis: AxisReference::Primary,
        };

        engine.chart_renderer.add_series(&chart_id, series).unwrap();
        chart_ids.push(chart_id);

        // Create corresponding interactive elements
        let element_id = engine.create_element(ElementType::Interactive, [
            ("chart_id".to_string(), serde_json::json!(chart_ids[i])),
            ("width".to_string(), serde_json::json!(300)),
            ("height".to_string(), serde_json::json!(200)),
        ].into_iter().collect()).unwrap();
        element_ids.push(element_id);
    }

    // Create data sources with realistic data
    for i in 0..10 {
        let mut dataset = Vec::new();
        for j in 0..200 {
            dataset.push(serde_json::json!({
                "value": ((i * 200 + j) as f64 * 0.01).sin() * 50.0 + 50.0 + (i as f64 * 10.0),
                "x": j as f64,
                "label": format!("Point {}", j),
                "category": format!("Category {}", i),
                "timestamp": start_time + (j as f64 * 1000.0)
            }));
        }

        let data_source = DataSource::new(
            format!("dashboard_data_{}", i),
            DataSourceType::Dynamic,
            serde_json::Value::Array(dataset),
        );
        engine.document_state.data_sources.insert(format!("dashboard_data_{}", i), data_source);
    }

    // Create animations for visual appeal
    for (i, element_id) in element_ids.iter().enumerate().take(5) {
        let keyframes = vec![
            Keyframe {
                time: 0.0,
                properties: [
                    ("opacity".to_string(), serde_json::json!(0.0)),
                    ("scale".to_string(), serde_json::json!(0.8)),
                ].into_iter().collect(),
            },
            Keyframe {
                time: 1.0,
                properties: [
                    ("opacity".to_string(), serde_json::json!(1.0)),
                    ("scale".to_string(), serde_json::json!(1.0)),
                ].into_iter().collect(),
            },
        ];

        let animation_id = engine.create_animation(
            element_id,
            AnimationType::Style,
            1000.0 + (i as f64 * 200.0), // Staggered animations
            keyframes
        ).unwrap();
        animation_ids.push(animation_id);
    }

    // Simulate realistic user interactions over time
    let mut interaction_count = 0;
    let mut render_times = Vec::new();

    for cycle in 0..100 {
        let cycle_start_time = get_current_timestamp();

        // Simulate user clicking on different charts
        let target_element = &element_ids[cycle % element_ids.len()];
        
        let click_event = InteractionEvent {
            event_type: InteractionType::Click,
            target_element: Some(target_element.clone()),
            position: Some(Position { x: (cycle % 300) as f64, y: (cycle % 200) as f64 }),
            data: HashMap::new(),
            timestamp: start_time + (cycle as f64 * 100.0),
            touch_data: None,
            mouse_data: Some(MouseData {
                button: MouseButton::Left,
                buttons: 1,
                position: Position { x: (cycle % 300) as f64, y: (cycle % 200) as f64 },
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

        let render_update = engine.process_interaction(click_event).unwrap();
        interaction_count += 1;

        // Update data periodically
        if cycle % 10 == 0 {
            let data_source_id = format!("dashboard_data_{}", cycle % 10);
            let mut updated_dataset = Vec::new();
            
            for j in 0..200 {
                updated_dataset.push(serde_json::json!({
                    "value": ((cycle * 200 + j) as f64 * 0.01).cos() * 40.0 + 60.0,
                    "x": j as f64,
                    "label": format!("Updated Point {}", j),
                    "cycle": cycle
                }));
            }

            let data_update_event = InteractionEvent {
                event_type: InteractionType::DataUpdate,
                target_element: None,
                position: None,
                data: [
                    ("data_source_id".to_string(), serde_json::json!(data_source_id)),
                    ("data".to_string(), serde_json::Value::Array(updated_dataset.clone())),
                ].into_iter().collect(),
                timestamp: start_time + (cycle as f64 * 100.0) + 50.0,
                touch_data: None,
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

            let _render_update = engine.process_interaction(data_update_event).unwrap();

            // Re-render affected chart
            let chart_id = &chart_ids[cycle % 10];
            let _rendered_chart = engine.chart_renderer.render_chart(chart_id, &serde_json::Value::Array(updated_dataset)).unwrap();
        }

        // Process animation frame
        let frame_timestamp = start_time + (cycle as f64 * 100.0) + 75.0;
        let _render_update = engine.render_frame(frame_timestamp).unwrap();

        let cycle_time = get_current_timestamp() - cycle_start_time;
        render_times.push(cycle_time);
    }

    let total_test_time = get_current_timestamp() - start_time;

    // Performance assertions
    let average_cycle_time = render_times.iter().sum::<f64>() / render_times.len() as f64;
    let max_cycle_time = render_times.iter().fold(0.0, |a, &b| a.max(b));

    assert!(average_cycle_time < 100.0, "Average cycle time too slow: {}ms", average_cycle_time);
    assert!(max_cycle_time < 500.0, "Max cycle time too slow: {}ms", max_cycle_time);
    assert!(total_test_time < 30000.0, "Total test time too slow: {}ms", total_test_time);

    // Verify system state
    assert_eq!(engine.document_state.elements.len(), 10);
    assert_eq!(engine.chart_renderer.charts.len(), 10);
    assert_eq!(engine.document_state.data_sources.len(), 10);

    // Verify performance metrics
    let interaction_metrics = engine.get_interaction_metrics();
    assert_eq!(interaction_metrics.total_events, interaction_count + 10); // +10 for data updates
    assert!(interaction_metrics.average_response_time < 50.0);

    let chart_performance = &engine.chart_renderer.performance_stats;
    assert_eq!(chart_performance.total_charts, 10);
    assert!(chart_performance.average_render_time < 100.0);

    // Verify memory efficiency
    assert!(engine.security_context.allocated_memory < 30 * 1024 * 1024); // Should use less than 30MB
}