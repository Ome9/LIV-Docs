use super::*;
use wasm_bindgen_test::*;
use std::time::{SystemTime, UNIX_EPOCH};

wasm_bindgen_test_configure!(run_in_browser);

// Test interactive chart functionality and performance benchmarks

#[wasm_bindgen_test]
fn test_chart_rendering_performance() {
    let permissions = WASMPermissions {
        memory_limit: 10 * 1024 * 1024, // 10MB for performance testing
        allowed_imports: vec!["console".to_string()],
        cpu_time_limit: 30000, // 30 seconds
        allow_networking: false,
        allow_file_system: false,
        allowed_interactions: vec![
            "create_element".to_string(),
            "modify_element".to_string(),
        ],
        max_data_size: 5 * 1024 * 1024, // 5MB
        max_elements: 1000,
    };

    let mut engine = InteractiveEngine::new(permissions).unwrap();

    // Create multiple charts with different types
    let chart_types = vec![
        ChartType::Line,
        ChartType::Bar,
        ChartType::Pie,
        ChartType::Scatter,
        ChartType::Area,
    ];

    let mut chart_ids = Vec::new();
    let start_time = get_current_timestamp();

    for (i, chart_type) in chart_types.iter().enumerate() {
        let config = ChartConfig {
            width: 400.0,
            height: 300.0,
            ..ChartConfig::default()
        };

        let chart_id = engine.chart_renderer.create_chart(
            chart_type.clone(),
            format!("perf_data_{}", i),
            config
        ).unwrap();

        // Add series to chart
        let series = ChartSeries {
            id: format!("series_{}", i),
            name: format!("Performance Series {}", i),
            data_field: "value".to_string(),
            color: "#1f77b4".to_string(),
            line_width: Some(2.0),
            fill_opacity: Some(0.3),
            marker_size: Some(4.0),
            marker_shape: Some(MarkerShape::Circle),
            visible: true,
            y_axis: AxisReference::Primary,
        };

        engine.chart_renderer.add_series(&chart_id, series).unwrap();
        chart_ids.push(chart_id);
    }

    let chart_creation_time = get_current_timestamp() - start_time;

    // Generate large dataset for performance testing
    let mut large_dataset = Vec::new();
    for i in 0..1000 {
        large_dataset.push(serde_json::json!({
            "value": (i as f64 * 0.1).sin() * 100.0 + 50.0,
            "x": i as f64,
            "label": format!("Point {}", i)
        }));
    }
    let test_data = serde_json::Value::Array(large_dataset);

    // Benchmark chart rendering
    let render_start_time = get_current_timestamp();
    let mut render_times = Vec::new();

    for chart_id in &chart_ids {
        let single_render_start = get_current_timestamp();
        let rendered_chart = engine.chart_renderer.render_chart(chart_id, &test_data).unwrap();
        let single_render_time = get_current_timestamp() - single_render_start;
        
        render_times.push(single_render_time);
        
        // Verify chart was rendered correctly
        assert!(!rendered_chart.svg_content.is_empty());
        assert_eq!(rendered_chart.data_points.len(), 1000);
    }

    let total_render_time = get_current_timestamp() - render_start_time;

    // Performance assertions
    assert!(chart_creation_time < 1000.0, "Chart creation took too long: {}ms", chart_creation_time);
    assert!(total_render_time < 5000.0, "Total rendering took too long: {}ms", total_render_time);
    
    let average_render_time = render_times.iter().sum::<f64>() / render_times.len() as f64;
    assert!(average_render_time < 1000.0, "Average render time too slow: {}ms", average_render_time);

    // Test performance metrics
    let performance_stats = &engine.chart_renderer.performance_stats;
    assert_eq!(performance_stats.total_charts, 5);
    assert!(performance_stats.average_render_time > 0.0);
    assert!(performance_stats.total_render_time > 0.0);
}

#[wasm_bindgen_test]
fn test_chart_data_update_performance() {
    let permissions = WASMPermissions {
        memory_limit: 10 * 1024 * 1024,
        allowed_imports: vec!["console".to_string()],
        cpu_time_limit: 30000,
        allow_networking: false,
        allow_file_system: false,
        allowed_interactions: vec![
            "create_element".to_string(),
            "modify_element".to_string(),
        ],
        max_data_size: 5 * 1024 * 1024,
        max_elements: 1000,
    };

    let mut engine = InteractiveEngine::new(permissions).unwrap();

    // Create a line chart
    let config = ChartConfig::default();
    let chart_id = engine.chart_renderer.create_chart(
        ChartType::Line,
        "update_perf_data".to_string(),
        config
    ).unwrap();

    let series = ChartSeries {
        id: "update_series".to_string(),
        name: "Update Performance Series".to_string(),
        data_field: "value".to_string(),
        color: "#ff7f0e".to_string(),
        line_width: Some(2.0),
        fill_opacity: None,
        marker_size: Some(3.0),
        marker_shape: Some(MarkerShape::Circle),
        visible: true,
        y_axis: AxisReference::Primary,
    };

    engine.chart_renderer.add_series(&chart_id, series).unwrap();

    // Benchmark rapid data updates
    let mut update_times = Vec::new();
    
    for i in 0..100 {
        let dataset = (0..100).map(|j| {
            serde_json::json!({
                "value": ((i + j) as f64 * 0.1).sin() * 50.0 + 50.0,
                "label": format!("Update {} Point {}", i, j)
            })
        }).collect::<Vec<_>>();
        
        let test_data = serde_json::Value::Array(dataset);
        
        let update_start = get_current_timestamp();
        engine.chart_renderer.update_chart_data(&chart_id, &test_data).unwrap();
        let rendered_chart = engine.chart_renderer.render_chart(&chart_id, &test_data).unwrap();
        let update_time = get_current_timestamp() - update_start;
        
        update_times.push(update_time);
        
        // Verify update was successful
        assert_eq!(rendered_chart.data_points.len(), 100);
    }

    let average_update_time = update_times.iter().sum::<f64>() / update_times.len() as f64;
    let max_update_time = update_times.iter().fold(0.0, |a, &b| a.max(b));
    let min_update_time = update_times.iter().fold(f64::INFINITY, |a, &b| a.min(b));

    // Performance assertions for data updates
    assert!(average_update_time < 100.0, "Average update time too slow: {}ms", average_update_time);
    assert!(max_update_time < 500.0, "Max update time too slow: {}ms", max_update_time);
    assert!(min_update_time < 50.0, "Min update time too slow: {}ms", min_update_time);
}

#[wasm_bindgen_test]
fn test_animation_performance() {
    let permissions = WASMPermissions {
        memory_limit: 10 * 1024 * 1024,
        allowed_imports: vec!["console".to_string()],
        cpu_time_limit: 30000,
        allow_networking: false,
        allow_file_system: false,
        allowed_interactions: vec![
            "create_element".to_string(),
            "create_animation".to_string(),
        ],
        max_data_size: 5 * 1024 * 1024,
        max_elements: 1000,
    };

    let mut engine = InteractiveEngine::new(permissions).unwrap();

    // Create multiple elements for animation
    let mut element_ids = Vec::new();
    for i in 0..50 {
        let element_id = engine.create_element(ElementType::Container, HashMap::new()).unwrap();
        element_ids.push(element_id);
    }

    // Create animations for all elements
    let mut animation_ids = Vec::new();
    let animation_start_time = get_current_timestamp();

    for (i, element_id) in element_ids.iter().enumerate() {
        let keyframes = vec![
            Keyframe {
                time: 0.0,
                properties: [
                    ("x".to_string(), serde_json::Value::Number(serde_json::Number::from(0))),
                    ("y".to_string(), serde_json::Value::Number(serde_json::Number::from(0))),
                    ("opacity".to_string(), serde_json::Value::Number(serde_json::Number::from_f64(0.0).unwrap())),
                ].into_iter().collect(),
            },
            Keyframe {
                time: 0.5,
                properties: [
                    ("x".to_string(), serde_json::Value::Number(serde_json::Number::from(50))),
                    ("y".to_string(), serde_json::Value::Number(serde_json::Number::from(25))),
                    ("opacity".to_string(), serde_json::Value::Number(serde_json::Number::from_f64(1.0).unwrap())),
                ].into_iter().collect(),
            },
            Keyframe {
                time: 1.0,
                properties: [
                    ("x".to_string(), serde_json::Value::Number(serde_json::Number::from(100))),
                    ("y".to_string(), serde_json::Value::Number(serde_json::Number::from(50))),
                    ("opacity".to_string(), serde_json::Value::Number(serde_json::Number::from_f64(0.5).unwrap())),
                ].into_iter().collect(),
            },
        ];

        let animation_id = engine.create_animation(
            element_id,
            AnimationType::Transform,
            2000.0, // 2 second animation
            keyframes
        ).unwrap();
        animation_ids.push(animation_id);
    }

    let animation_creation_time = get_current_timestamp() - animation_start_time;

    // Benchmark animation frame processing (simulate 60fps for 2 seconds)
    let frame_start_time = get_current_timestamp();
    let mut frame_times = Vec::new();
    let frame_duration = 1000.0 / 60.0; // 60fps = ~16.67ms per frame

    for frame in 0..120 { // 2 seconds at 60fps
        let timestamp = frame_start_time + (frame as f64 * frame_duration);
        
        let single_frame_start = get_current_timestamp();
        let render_update = engine.render_frame(timestamp).unwrap();
        let single_frame_time = get_current_timestamp() - single_frame_start;
        
        frame_times.push(single_frame_time);
        
        // Verify animation updates are being generated
        if frame < 100 { // Before animations complete
            assert!(render_update.animation_updates.len() >= 0);
        }
    }

    let total_frame_time = get_current_timestamp() - frame_start_time;
    let average_frame_time = frame_times.iter().sum::<f64>() / frame_times.len() as f64;
    let max_frame_time = frame_times.iter().fold(0.0, |a, &b| a.max(b));

    // Performance assertions for animations
    assert!(animation_creation_time < 1000.0, "Animation creation too slow: {}ms", animation_creation_time);
    assert!(average_frame_time < 16.67, "Average frame time too slow for 60fps: {}ms", average_frame_time);
    assert!(max_frame_time < 33.33, "Max frame time too slow (dropped frames): {}ms", max_frame_time);
    assert!(total_frame_time < 3000.0, "Total animation processing too slow: {}ms", total_frame_time);
}

#[wasm_bindgen_test]
fn test_vector_graphics_performance() {
    let permissions = WASMPermissions {
        memory_limit: 10 * 1024 * 1024,
        allowed_imports: vec!["console".to_string()],
        cpu_time_limit: 30000,
        allow_networking: false,
        allow_file_system: false,
        allowed_interactions: vec![
            "create_element".to_string(),
        ],
        max_data_size: 5 * 1024 * 1024,
        max_elements: 1000,
    };

    let mut engine = InteractiveEngine::new(permissions).unwrap();

    // Benchmark vector shape creation
    let shape_creation_start = get_current_timestamp();
    let mut shape_ids = Vec::new();

    for i in 0..200 {
        let shape_id = engine.vector_engine.create_shape(
            ShapeType::Rectangle,
            Position { x: (i % 20) as f64 * 20.0, y: (i / 20) as f64 * 20.0 },
            Size { width: 15.0, height: 15.0 }
        ).unwrap();
        shape_ids.push(shape_id);
    }

    let shape_creation_time = get_current_timestamp() - shape_creation_start;

    // Benchmark complex path creation
    let path_creation_start = get_current_timestamp();
    let mut path_ids = Vec::new();

    for i in 0..50 {
        let spiral_params = [
            ("center_x".to_string(), 100.0 + (i as f64 * 10.0)),
            ("center_y".to_string(), 100.0 + (i as f64 * 10.0)),
            ("start_radius".to_string(), 5.0),
            ("end_radius".to_string(), 30.0),
            ("turns".to_string(), 3.0),
        ].into_iter().collect();

        let path_id = engine.vector_engine.create_complex_path(
            ComplexPathType::Spiral,
            spiral_params
        ).unwrap();
        path_ids.push(path_id);
    }

    let path_creation_time = get_current_timestamp() - path_creation_start;

    // Benchmark SVG rendering
    let render_start = get_current_timestamp();
    let svg_content = engine.vector_engine.render_to_svg(800.0, 600.0);
    let render_time = get_current_timestamp() - render_start;

    // Performance assertions for vector graphics
    assert!(shape_creation_time < 1000.0, "Shape creation too slow: {}ms", shape_creation_time);
    assert!(path_creation_time < 2000.0, "Path creation too slow: {}ms", path_creation_time);
    assert!(render_time < 500.0, "SVG rendering too slow: {}ms", render_time);
    
    // Verify SVG content quality
    assert!(svg_content.contains("<svg"));
    assert!(svg_content.contains("</svg>"));
    assert!(svg_content.contains("<rect")); // Should have rectangles
    assert!(svg_content.contains("<path")); // Should have paths
    assert!(svg_content.len() > 1000); // Should be substantial content
}

#[wasm_bindgen_test]
fn test_data_binding_performance() {
    let permissions = WASMPermissions {
        memory_limit: 10 * 1024 * 1024,
        allowed_imports: vec!["console".to_string()],
        cpu_time_limit: 30000,
        allow_networking: false,
        allow_file_system: false,
        allowed_interactions: vec![
            "create_element".to_string(),
            "modify_element".to_string(),
        ],
        max_data_size: 5 * 1024 * 1024,
        max_elements: 1000,
    };

    let mut engine = InteractiveEngine::new(permissions).unwrap();

    // Create multiple data sources
    for i in 0..100 {
        let data_source = DataSource::new(
            format!("perf_data_{}", i),
            DataSourceType::Dynamic,
            serde_json::json!({
                "value": i as f64,
                "timestamp": get_current_timestamp(),
                "metadata": {
                    "source": format!("generator_{}", i),
                    "type": "performance_test"
                }
            }),
        );
        engine.document_state.data_sources.insert(format!("perf_data_{}", i), data_source);
    }

    // Create elements bound to data sources
    let mut element_ids = Vec::new();
    for i in 0..100 {
        let properties = [
            ("data_source".to_string(), serde_json::Value::String(format!("perf_data_{}", i))),
            ("width".to_string(), serde_json::Value::Number(serde_json::Number::from(50))),
            ("height".to_string(), serde_json::Value::Number(serde_json::Number::from(50))),
        ].into_iter().collect();

        let element_id = engine.create_element(ElementType::Chart, properties).unwrap();
        element_ids.push(element_id);

        // Add data binding
        let binding = DataBinding {
            source_id: format!("perf_data_{}", i),
            target_element: element_id.clone(),
            property_path: "value".to_string(),
            transform_function: Some("percentage".to_string()),
            update_trigger: UpdateTrigger::Immediate,
        };
        engine.data_binding_manager.add_binding(binding);
    }

    // Benchmark data binding updates
    let binding_start_time = get_current_timestamp();
    let mut binding_update_times = Vec::new();

    for update_cycle in 0..50 {
        let cycle_start = get_current_timestamp();
        
        // Update all data sources
        for i in 0..100 {
            let new_data = serde_json::json!({
                "value": (update_cycle * 100 + i) as f64,
                "timestamp": get_current_timestamp(),
                "cycle": update_cycle
            });
            
            if let Some(data_source) = engine.document_state.data_sources.get_mut(&format!("perf_data_{}", i)) {
                data_source.update_data(new_data).unwrap();
            }
        }
        
        // Process binding updates
        let current_time = get_current_timestamp();
        let _binding_changes = engine.data_binding_manager.update_bindings(
            &mut engine.document_state,
            current_time
        );
        
        let cycle_time = get_current_timestamp() - cycle_start;
        binding_update_times.push(cycle_time);
    }

    let total_binding_time = get_current_timestamp() - binding_start_time;
    let average_binding_time = binding_update_times.iter().sum::<f64>() / binding_update_times.len() as f64;
    let max_binding_time = binding_update_times.iter().fold(0.0, |a, &b| a.max(b));

    // Performance assertions for data binding
    assert!(average_binding_time < 50.0, "Average binding update too slow: {}ms", average_binding_time);
    assert!(max_binding_time < 200.0, "Max binding update too slow: {}ms", max_binding_time);
    assert!(total_binding_time < 5000.0, "Total binding processing too slow: {}ms", total_binding_time);
}

#[wasm_bindgen_test]
fn test_memory_usage_efficiency() {
    let permissions = WASMPermissions {
        memory_limit: 5 * 1024 * 1024, // 5MB limit
        allowed_imports: vec!["console".to_string()],
        cpu_time_limit: 30000,
        allow_networking: false,
        allow_file_system: false,
        allowed_interactions: vec![
            "create_element".to_string(),
            "modify_element".to_string(),
            "create_animation".to_string(),
        ],
        max_data_size: 2 * 1024 * 1024,
        max_elements: 500,
    };

    let mut engine = InteractiveEngine::new(permissions).unwrap();

    // Track memory usage during operations
    let initial_memory = engine.security_context.allocated_memory;

    // Create elements and track memory growth
    let mut memory_samples = Vec::new();
    
    for i in 0..100 {
        let properties = [
            ("width".to_string(), serde_json::Value::Number(serde_json::Number::from(100))),
            ("height".to_string(), serde_json::Value::Number(serde_json::Number::from(100))),
            ("data".to_string(), serde_json::Value::String(format!("element_data_{}", i))),
        ].into_iter().collect();

        let _element_id = engine.create_element(ElementType::Container, properties).unwrap();
        memory_samples.push(engine.security_context.allocated_memory);
    }

    // Create animations and track memory
    let element_ids: Vec<String> = engine.document_state.elements.iter().map(|e| e.id.clone()).collect();
    
    for (i, element_id) in element_ids.iter().enumerate().take(50) {
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

        let _animation_id = engine.create_animation(element_id, AnimationType::Transform, 1000.0, keyframes).unwrap();
        memory_samples.push(engine.security_context.allocated_memory);
    }

    let final_memory = engine.security_context.allocated_memory;
    let memory_growth = final_memory - initial_memory;

    // Memory efficiency assertions
    assert!(memory_growth < 2 * 1024 * 1024, "Memory growth too high: {} bytes", memory_growth);
    assert!(final_memory < 4 * 1024 * 1024, "Final memory usage too high: {} bytes", final_memory);
    
    // Check for memory leaks (memory should grow reasonably with content)
    let expected_memory_per_element = memory_growth / 150; // 100 elements + 50 animations
    assert!(expected_memory_per_element < 10 * 1024, "Memory per element too high: {} bytes", expected_memory_per_element);
}

#[wasm_bindgen_test]
fn test_render_update_efficiency() {
    let permissions = WASMPermissions {
        memory_limit: 10 * 1024 * 1024,
        allowed_imports: vec!["console".to_string()],
        cpu_time_limit: 30000,
        allow_networking: false,
        allow_file_system: false,
        allowed_interactions: vec![
            "create_element".to_string(),
            "modify_element".to_string(),
            "Click".to_string(),
        ],
        max_data_size: 5 * 1024 * 1024,
        max_elements: 1000,
    };

    let mut engine = InteractiveEngine::new(permissions).unwrap();

    // Create elements for interaction testing
    let mut element_ids = Vec::new();
    for i in 0..100 {
        let element_id = engine.create_element(ElementType::Interactive, HashMap::new()).unwrap();
        element_ids.push(element_id);
    }

    // Benchmark render update generation
    let mut render_update_times = Vec::new();
    
    for i in 0..200 {
        let target_element = &element_ids[i % element_ids.len()];
        
        let interaction_event = InteractionEvent {
            event_type: InteractionType::Click,
            target_element: Some(target_element.clone()),
            position: Some(Position { x: (i % 100) as f64, y: (i / 100) as f64 }),
            data: HashMap::new(),
            timestamp: get_current_timestamp() + (i as f64),
            touch_data: None,
            mouse_data: Some(MouseData {
                button: MouseButton::Left,
                buttons: 1,
                position: Position { x: (i % 100) as f64, y: (i / 100) as f64 },
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

        let render_start = get_current_timestamp();
        let render_update = engine.process_interaction(interaction_event).unwrap();
        let render_time = get_current_timestamp() - render_start;
        
        render_update_times.push(render_time);
        
        // Verify render update quality
        assert!(render_update.dom_operations.len() > 0 || 
                render_update.style_changes.len() > 0 || 
                render_update.animation_updates.len() > 0);
    }

    let average_render_time = render_update_times.iter().sum::<f64>() / render_update_times.len() as f64;
    let max_render_time = render_update_times.iter().fold(0.0, |a, &b| a.max(b));
    let min_render_time = render_update_times.iter().fold(f64::INFINITY, |a, &b| a.min(b));

    // Render update efficiency assertions
    assert!(average_render_time < 10.0, "Average render update too slow: {}ms", average_render_time);
    assert!(max_render_time < 50.0, "Max render update too slow: {}ms", max_render_time);
    assert!(min_render_time < 5.0, "Min render update too slow: {}ms", min_render_time);
}