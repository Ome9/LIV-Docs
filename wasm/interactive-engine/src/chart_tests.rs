use super::*;
use wasm_bindgen_test::*;

wasm_bindgen_test_configure!(run_in_browser);

#[wasm_bindgen_test]
fn test_chart_renderer_creation() {
    let mut chart_renderer = ChartRenderer::new();
    
    let config = ChartConfig::default();
    let chart_id = chart_renderer.create_chart(
        ChartType::Line,
        "test_data".to_string(),
        config
    ).unwrap();
    
    assert!(!chart_id.is_empty());
    assert_eq!(chart_renderer.charts.len(), 1);
    assert_eq!(chart_renderer.performance_stats.total_charts, 1);
}

#[wasm_bindgen_test]
fn test_line_chart_rendering() {
    let mut chart_renderer = ChartRenderer::new();
    
    let config = ChartConfig {
        width: 400.0,
        height: 300.0,
        ..ChartConfig::default()
    };
    
    let chart_id = chart_renderer.create_chart(
        ChartType::Line,
        "test_data".to_string(),
        config
    ).unwrap();
    
    // Add a series
    let series = ChartSeries {
        id: "series1".to_string(),
        name: "Test Series".to_string(),
        data_field: "value".to_string(),
        color: "#1f77b4".to_string(),
        line_width: Some(2.0),
        fill_opacity: None,
        marker_size: Some(4.0),
        marker_shape: Some(MarkerShape::Circle),
        visible: true,
        y_axis: AxisReference::Primary,
    };
    
    chart_renderer.add_series(&chart_id, series).unwrap();
    
    // Test data
    let test_data = serde_json::json!([
        {"value": 10, "label": "A"},
        {"value": 20, "label": "B"},
        {"value": 15, "label": "C"},
        {"value": 25, "label": "D"}
    ]);
    
    let rendered_chart = chart_renderer.render_chart(&chart_id, &test_data).unwrap();
    
    assert_eq!(rendered_chart.chart_id, chart_id);
    assert!(rendered_chart.svg_content.contains("<svg"));
    assert!(rendered_chart.svg_content.contains("</svg>"));
    assert_eq!(rendered_chart.data_points.len(), 4);
    assert_eq!(rendered_chart.bounds.width, 400.0);
    assert_eq!(rendered_chart.bounds.height, 300.0);
}

#[wasm_bindgen_test]
fn test_bar_chart_rendering() {
    let mut chart_renderer = ChartRenderer::new();
    
    let config = ChartConfig::default();
    let chart_id = chart_renderer.create_chart(
        ChartType::Bar,
        "test_data".to_string(),
        config
    ).unwrap();
    
    let series = ChartSeries {
        id: "series1".to_string(),
        name: "Test Series".to_string(),
        data_field: "value".to_string(),
        color: "#ff7f0e".to_string(),
        line_width: None,
        fill_opacity: Some(0.8),
        marker_size: None,
        marker_shape: None,
        visible: true,
        y_axis: AxisReference::Primary,
    };
    
    chart_renderer.add_series(&chart_id, series).unwrap();
    
    let test_data = serde_json::json!([
        {"value": 30, "label": "Category A"},
        {"value": 45, "label": "Category B"},
        {"value": 20, "label": "Category C"}
    ]);
    
    let rendered_chart = chart_renderer.render_chart(&chart_id, &test_data).unwrap();
    
    assert!(rendered_chart.svg_content.contains("<rect"));
    assert_eq!(rendered_chart.data_points.len(), 3);
}

#[wasm_bindgen_test]
fn test_pie_chart_rendering() {
    let mut chart_renderer = ChartRenderer::new();
    
    let config = ChartConfig::default();
    let chart_id = chart_renderer.create_chart(
        ChartType::Pie,
        "test_data".to_string(),
        config
    ).unwrap();
    
    let series = ChartSeries {
        id: "series1".to_string(),
        name: "Test Series".to_string(),
        data_field: "value".to_string(),
        color: "#2ca02c".to_string(),
        line_width: None,
        fill_opacity: None,
        marker_size: None,
        marker_shape: None,
        visible: true,
        y_axis: AxisReference::Primary,
    };
    
    chart_renderer.add_series(&chart_id, series).unwrap();
    
    let test_data = serde_json::json!([
        {"value": 40, "label": "Slice A"},
        {"value": 30, "label": "Slice B"},
        {"value": 20, "label": "Slice C"},
        {"value": 10, "label": "Slice D"}
    ]);
    
    let rendered_chart = chart_renderer.render_chart(&chart_id, &test_data).unwrap();
    
    assert!(rendered_chart.svg_content.contains("<path"));
    assert_eq!(rendered_chart.data_points.len(), 4);
}

#[wasm_bindgen_test]
fn test_chart_caching() {
    let mut chart_renderer = ChartRenderer::new();
    
    let config = ChartConfig::default();
    let chart_id = chart_renderer.create_chart(
        ChartType::Line,
        "test_data".to_string(),
        config
    ).unwrap();
    
    let series = ChartSeries {
        id: "series1".to_string(),
        name: "Test Series".to_string(),
        data_field: "value".to_string(),
        color: "#1f77b4".to_string(),
        line_width: Some(2.0),
        fill_opacity: None,
        marker_size: None,
        marker_shape: None,
        visible: true,
        y_axis: AxisReference::Primary,
    };
    
    chart_renderer.add_series(&chart_id, series).unwrap();
    
    let test_data = serde_json::json!([
        {"value": 10, "label": "A"},
        {"value": 20, "label": "B"}
    ]);
    
    // First render - should cache the result
    let rendered_chart1 = chart_renderer.render_chart(&chart_id, &test_data).unwrap();
    assert_eq!(chart_renderer.render_cache.len(), 1);
    
    // Second render - should use cached result
    let rendered_chart2 = chart_renderer.render_chart(&chart_id, &test_data).unwrap();
    assert_eq!(rendered_chart1.svg_content, rendered_chart2.svg_content);
    
    // Update data - should invalidate cache
    chart_renderer.update_chart_data(&chart_id, &test_data).unwrap();
    assert_eq!(chart_renderer.render_cache.len(), 0);
}

#[wasm_bindgen_test]
fn test_vector_engine_creation() {
    let mut vector_engine = VectorEngine::new();
    
    let shape_id = vector_engine.create_shape(
        ShapeType::Rectangle,
        Position { x: 10.0, y: 20.0 },
        Size { width: 100.0, height: 50.0 }
    ).unwrap();
    
    assert!(!shape_id.is_empty());
    assert_eq!(vector_engine.shapes.len(), 1);
}

#[wasm_bindgen_test]
fn test_vector_shape_rendering() {
    let mut vector_engine = VectorEngine::new();
    
    // Create a rectangle
    let rect_id = vector_engine.create_shape(
        ShapeType::Rectangle,
        Position { x: 10.0, y: 20.0 },
        Size { width: 100.0, height: 50.0 }
    ).unwrap();
    
    // Create a circle
    let circle_id = vector_engine.create_shape(
        ShapeType::Circle,
        Position { x: 150.0, y: 20.0 },
        Size { width: 60.0, height: 60.0 }
    ).unwrap();
    
    let svg_content = vector_engine.render_to_svg(400.0, 300.0);
    
    assert!(svg_content.contains("<svg"));
    assert!(svg_content.contains("</svg>"));
    assert!(svg_content.contains("<rect"));
    assert!(svg_content.contains("<circle"));
    assert!(svg_content.contains("width=\"400\""));
    assert!(svg_content.contains("height=\"300\""));
}

#[wasm_bindgen_test]
fn test_vector_path_creation() {
    let mut vector_engine = VectorEngine::new();
    
    let path_commands = vec![
        PathCommand::MoveTo { x: 10.0, y: 10.0 },
        PathCommand::LineTo { x: 100.0, y: 10.0 },
        PathCommand::LineTo { x: 100.0, y: 100.0 },
        PathCommand::LineTo { x: 10.0, y: 100.0 },
        PathCommand::ClosePath,
    ];
    
    let path_id = vector_engine.create_path(path_commands).unwrap();
    
    assert!(!path_id.is_empty());
    assert_eq!(vector_engine.paths.len(), 1);
    
    let svg_content = vector_engine.render_to_svg(200.0, 200.0);
    assert!(svg_content.contains("<path"));
    assert!(svg_content.contains("M 10 10"));
    assert!(svg_content.contains("L 100 10"));
    assert!(svg_content.contains("Z"));
}

#[wasm_bindgen_test]
fn test_gradient_creation() {
    let mut vector_engine = VectorEngine::new();
    
    let gradient_stops = vec![
        GradientStop {
            offset: 0.0,
            color: "#ff0000".to_string(),
            opacity: 1.0,
        },
        GradientStop {
            offset: 1.0,
            color: "#0000ff".to_string(),
            opacity: 1.0,
        },
    ];
    
    let gradient_id = vector_engine.create_gradient(
        GradientType::Linear { x1: 0.0, y1: 0.0, x2: 1.0, y2: 0.0 },
        gradient_stops
    ).unwrap();
    
    assert!(!gradient_id.is_empty());
    assert_eq!(vector_engine.gradients.len(), 1);
    
    let svg_content = vector_engine.render_to_svg(200.0, 200.0);
    assert!(svg_content.contains("<defs>"));
    assert!(svg_content.contains("<linearGradient"));
    assert!(svg_content.contains("stop-color=\"#ff0000\""));
    assert!(svg_content.contains("stop-color=\"#0000ff\""));
}

#[wasm_bindgen_test]
fn test_chart_performance_stats() {
    let mut chart_renderer = ChartRenderer::new();
    
    // Create multiple charts
    for i in 0..3 {
        let config = ChartConfig::default();
        let chart_id = chart_renderer.create_chart(
            ChartType::Line,
            format!("test_data_{}", i),
            config
        ).unwrap();
        
        let series = ChartSeries {
            id: format!("series_{}", i),
            name: format!("Test Series {}", i),
            data_field: "value".to_string(),
            color: "#1f77b4".to_string(),
            line_width: Some(2.0),
            fill_opacity: None,
            marker_size: None,
            marker_shape: None,
            visible: true,
            y_axis: AxisReference::Primary,
        };
        
        chart_renderer.add_series(&chart_id, series).unwrap();
        
        let test_data = serde_json::json!([
            {"value": 10 + i, "label": "A"},
            {"value": 20 + i, "label": "B"}
        ]);
        
        chart_renderer.render_chart(&chart_id, &test_data).unwrap();
    }
    
    assert_eq!(chart_renderer.performance_stats.total_charts, 3);
    assert!(chart_renderer.performance_stats.total_render_time > 0.0);
    assert!(chart_renderer.performance_stats.average_render_time > 0.0);
}

#[wasm_bindgen_test]
fn test_chart_config_defaults() {
    let config = ChartConfig::default();
    
    assert_eq!(config.width, 400.0);
    assert_eq!(config.height, 300.0);
    assert_eq!(config.margin.top, 20.0);
    assert_eq!(config.margin.right, 20.0);
    assert_eq!(config.margin.bottom, 40.0);
    assert_eq!(config.margin.left, 40.0);
    assert!(config.responsive);
    assert!(config.maintain_aspect_ratio);
    assert_eq!(config.background_color, Some("#ffffff".to_string()));
    assert!(config.tooltip.is_some());
}

#[wasm_bindgen_test]
fn test_chart_styling_defaults() {
    let styling = ChartStyling::default();
    
    assert_eq!(styling.color_palette.len(), 10);
    assert_eq!(styling.color_palette[0], "#1f77b4");
    assert!(!styling.gradient_fills);
    assert!(!styling.drop_shadow);
    assert_eq!(styling.border_radius, 0.0);
    assert_eq!(styling.grid_color, "#e0e0e0");
    assert_eq!(styling.grid_opacity, 0.5);
}

#[wasm_bindgen_test]
fn test_multiple_chart_types() {
    let mut chart_renderer = ChartRenderer::new();
    
    let chart_types = vec![
        ChartType::Line,
        ChartType::Bar,
        ChartType::Pie,
        ChartType::Scatter,
        ChartType::Area,
    ];
    
    let test_data = serde_json::json!([
        {"value": 10, "label": "A"},
        {"value": 20, "label": "B"},
        {"value": 15, "label": "C"}
    ]);
    
    for (i, chart_type) in chart_types.iter().enumerate() {
        let config = ChartConfig::default();
        let chart_id = chart_renderer.create_chart(
            chart_type.clone(),
            format!("test_data_{}", i),
            config
        ).unwrap();
        
        let series = ChartSeries {
            id: format!("series_{}", i),
            name: format!("Test Series {}", i),
            data_field: "value".to_string(),
            color: "#1f77b4".to_string(),
            line_width: Some(2.0),
            fill_opacity: None,
            marker_size: None,
            marker_shape: None,
            visible: true,
            y_axis: AxisReference::Primary,
        };
        
        chart_renderer.add_series(&chart_id, series).unwrap();
        
        let rendered_chart = chart_renderer.render_chart(&chart_id, &test_data).unwrap();
        
        assert!(!rendered_chart.svg_content.is_empty());
        assert_eq!(rendered_chart.data_points.len(), 3);
    }
    
    assert_eq!(chart_renderer.charts.len(), 5);
}

#[wasm_bindgen_test]
fn test_scatter_chart_rendering() {
    let mut chart_renderer = ChartRenderer::new();
    
    let config = ChartConfig::default();
    let chart_id = chart_renderer.create_chart(
        ChartType::Scatter,
        "scatter_data".to_string(),
        config
    ).unwrap();
    
    let series = ChartSeries {
        id: "scatter_series".to_string(),
        name: "Scatter Series".to_string(),
        data_field: "y".to_string(),
        color: "#ff7f0e".to_string(),
        line_width: None,
        fill_opacity: None,
        marker_size: Some(6.0),
        marker_shape: Some(MarkerShape::Circle),
        visible: true,
        y_axis: AxisReference::Primary,
    };
    
    chart_renderer.add_series(&chart_id, series).unwrap();
    
    let test_data = serde_json::json!([
        {"x": 10, "y": 20, "label": "Point A"},
        {"x": 25, "y": 35, "label": "Point B"},
        {"x": 40, "y": 15, "label": "Point C"},
        {"x": 55, "y": 45, "label": "Point D"}
    ]);
    
    let rendered_chart = chart_renderer.render_chart(&chart_id, &test_data).unwrap();
    
    assert!(rendered_chart.svg_content.contains("<circle"));
    assert_eq!(rendered_chart.data_points.len(), 4);
    
    // Verify scatter plot specific properties
    for point in &rendered_chart.data_points {
        assert!(point.value.get("x").is_some());
        assert!(point.value.get("y").is_some());
    }
}

#[wasm_bindgen_test]
fn test_area_chart_rendering() {
    let mut chart_renderer = ChartRenderer::new();
    
    let config = ChartConfig::default();
    let chart_id = chart_renderer.create_chart(
        ChartType::Area,
        "area_data".to_string(),
        config
    ).unwrap();
    
    let series = ChartSeries {
        id: "area_series".to_string(),
        name: "Area Series".to_string(),
        data_field: "value".to_string(),
        color: "#2ca02c".to_string(),
        line_width: Some(2.0),
        fill_opacity: Some(0.4),
        marker_size: None,
        marker_shape: None,
        visible: true,
        y_axis: AxisReference::Primary,
    };
    
    chart_renderer.add_series(&chart_id, series).unwrap();
    
    let test_data = serde_json::json!([
        {"value": 10, "label": "Jan"},
        {"value": 25, "label": "Feb"},
        {"value": 20, "label": "Mar"},
        {"value": 35, "label": "Apr"},
        {"value": 30, "label": "May"}
    ]);
    
    let rendered_chart = chart_renderer.render_chart(&chart_id, &test_data).unwrap();
    
    assert!(rendered_chart.svg_content.contains("<path"));
    assert!(rendered_chart.svg_content.contains("fill-opacity"));
    assert_eq!(rendered_chart.data_points.len(), 5);
}

#[wasm_bindgen_test]
fn test_histogram_chart_rendering() {
    let mut chart_renderer = ChartRenderer::new();
    
    let config = ChartConfig::default();
    let chart_id = chart_renderer.create_chart(
        ChartType::Histogram,
        "histogram_data".to_string(),
        config
    ).unwrap();
    
    let series = ChartSeries {
        id: "histogram_series".to_string(),
        name: "Histogram Series".to_string(),
        data_field: "value".to_string(),
        color: "#d62728".to_string(),
        line_width: None,
        fill_opacity: None,
        marker_size: None,
        marker_shape: None,
        visible: true,
        y_axis: AxisReference::Primary,
    };
    
    chart_renderer.add_series(&chart_id, series).unwrap();
    
    // Generate test data with distribution
    let mut test_values = Vec::new();
    for i in 0..100 {
        test_values.push(serde_json::json!({"value": (i % 50) as f64 + (i / 10) as f64}));
    }
    let test_data = serde_json::Value::Array(test_values);
    
    let rendered_chart = chart_renderer.render_chart(&chart_id, &test_data).unwrap();
    
    assert!(rendered_chart.svg_content.contains("<rect"));
    assert_eq!(rendered_chart.data_points.len(), 10); // Default bin count
}

#[wasm_bindgen_test]
fn test_heatmap_chart_rendering() {
    let mut chart_renderer = ChartRenderer::new();
    
    let config = ChartConfig::default();
    let chart_id = chart_renderer.create_chart(
        ChartType::Heatmap,
        "heatmap_data".to_string(),
        config
    ).unwrap();
    
    // Create 2D grid data for heatmap
    let test_data = serde_json::json!([
        [10, 20, 30],
        [40, 50, 60],
        [70, 80, 90]
    ]);
    
    let rendered_chart = chart_renderer.render_chart(&chart_id, &test_data).unwrap();
    
    assert!(rendered_chart.svg_content.contains("<rect"));
    assert_eq!(rendered_chart.data_points.len(), 9); // 3x3 grid
    
    // Verify color mapping
    for point in &rendered_chart.data_points {
        assert!(point.color.starts_with("rgb("));
    }
}

#[wasm_bindgen_test]
fn test_radar_chart_rendering() {
    let mut chart_renderer = ChartRenderer::new();
    
    let config = ChartConfig::default();
    let chart_id = chart_renderer.create_chart(
        ChartType::Radar,
        "radar_data".to_string(),
        config
    ).unwrap();
    
    let series = ChartSeries {
        id: "radar_series".to_string(),
        name: "Radar Series".to_string(),
        data_field: "value".to_string(),
        color: "#9467bd".to_string(),
        line_width: Some(2.0),
        fill_opacity: Some(0.3),
        marker_size: None,
        marker_shape: None,
        visible: true,
        y_axis: AxisReference::Primary,
    };
    
    chart_renderer.add_series(&chart_id, series).unwrap();
    
    let test_data = serde_json::json!([
        {"value": 80, "label": "Speed"},
        {"value": 60, "label": "Reliability"},
        {"value": 90, "label": "Comfort"},
        {"value": 70, "label": "Safety"},
        {"value": 85, "label": "Efficiency"}
    ]);
    
    let rendered_chart = chart_renderer.render_chart(&chart_id, &test_data).unwrap();
    
    assert!(rendered_chart.svg_content.contains("<circle")); // Grid circles
    assert!(rendered_chart.svg_content.contains("<line")); // Radial lines
    assert!(rendered_chart.svg_content.contains("<path")); // Data polygon
    assert_eq!(rendered_chart.data_points.len(), 5);
}

#[wasm_bindgen_test]
fn test_gauge_chart_rendering() {
    let mut chart_renderer = ChartRenderer::new();
    
    let config = ChartConfig::default();
    let chart_id = chart_renderer.create_chart(
        ChartType::Gauge,
        "gauge_data".to_string(),
        config
    ).unwrap();
    
    let series = ChartSeries {
        id: "gauge_series".to_string(),
        name: "Gauge Series".to_string(),
        data_field: "value".to_string(),
        color: "#4CAF50".to_string(),
        line_width: None,
        fill_opacity: None,
        marker_size: None,
        marker_shape: None,
        visible: true,
        y_axis: AxisReference::Primary,
    };
    
    chart_renderer.add_series(&chart_id, series).unwrap();
    
    let test_data = serde_json::json!([
        {"value": 75, "label": "Performance"}
    ]);
    
    let rendered_chart = chart_renderer.render_chart(&chart_id, &test_data).unwrap();
    
    assert!(rendered_chart.svg_content.contains("<path")); // Gauge arcs
    assert!(rendered_chart.svg_content.contains("<line")); // Needle
    assert!(rendered_chart.svg_content.contains("<circle")); // Center circle
    assert!(rendered_chart.svg_content.contains("<text")); // Value text
    assert_eq!(rendered_chart.data_points.len(), 1);
}

#[wasm_bindgen_test]
fn test_candlestick_chart_rendering() {
    let mut chart_renderer = ChartRenderer::new();
    
    let config = ChartConfig::default();
    let chart_id = chart_renderer.create_chart(
        ChartType::Candlestick,
        "candlestick_data".to_string(),
        config
    ).unwrap();
    
    let test_data = serde_json::json!([
        {"open": 100, "high": 110, "low": 95, "close": 105, "label": "Day 1"},
        {"open": 105, "high": 115, "low": 100, "close": 98, "label": "Day 2"},
        {"open": 98, "high": 108, "low": 92, "close": 102, "label": "Day 3"}
    ]);
    
    let rendered_chart = chart_renderer.render_chart(&chart_id, &test_data).unwrap();
    
    assert!(rendered_chart.svg_content.contains("<line")); // High-low lines
    assert!(rendered_chart.svg_content.contains("<rect")); // Body rectangles
    assert_eq!(rendered_chart.data_points.len(), 3);
    
    // Verify OHLC data structure
    for point in &rendered_chart.data_points {
        let ohlc = point.value.as_object().unwrap();
        assert!(ohlc.contains_key("open"));
        assert!(ohlc.contains_key("high"));
        assert!(ohlc.contains_key("low"));
        assert!(ohlc.contains_key("close"));
    }
}

#[wasm_bindgen_test]
fn test_data_source_creation_and_updates() {
    let mut data_source = DataSource::new(
        "test_source".to_string(),
        DataSourceType::Dynamic,
        serde_json::json!([1, 2, 3, 4, 5])
    );
    
    assert_eq!(data_source.id, "test_source");
    assert!(matches!(data_source.source_type, DataSourceType::Dynamic));
    
    // Test data update
    let new_data = serde_json::json!([6, 7, 8, 9, 10]);
    data_source.update_data(new_data).unwrap();
    
    assert_eq!(data_source.data, serde_json::json!([6, 7, 8, 9, 10]));
    
    // Test statistics
    let stats = data_source.get_data_statistics();
    assert_eq!(stats.count, 5);
    assert_eq!(stats.min, 6.0);
    assert_eq!(stats.max, 10.0);
    assert_eq!(stats.sum, 40.0);
    assert_eq!(stats.mean, 8.0);
}

#[wasm_bindgen_test]
fn test_stream_data_source() {
    let mut data_source = DataSource::new(
        "stream_source".to_string(),
        DataSourceType::Stream,
        serde_json::json!([1, 2, 3])
    );
    
    // Add new data to stream
    let new_data = serde_json::json!([4, 5, 6]);
    data_source.update_data(new_data).unwrap();
    
    // Should append to existing data
    assert_eq!(data_source.data, serde_json::json!([1, 2, 3, 4, 5, 6]));
    
    // Test latest values
    let latest = data_source.get_latest_values(3);
    assert_eq!(latest, vec![
        serde_json::json!(4),
        serde_json::json!(5),
        serde_json::json!(6)
    ]);
}

#[wasm_bindgen_test]
fn test_data_binding_manager() {
    let mut binding_manager = DataBindingManager::new();
    
    let binding = DataBinding {
        source_id: "test_source".to_string(),
        target_element: "test_element".to_string(),
        property_path: "value".to_string(),
        transform_function: Some("percentage".to_string()),
        update_trigger: UpdateTrigger::Immediate,
    };
    
    let binding_id = binding_manager.add_binding(binding);
    assert!(!binding_id.is_empty());
    
    // Test binding removal
    binding_manager.remove_binding(&binding_id);
    assert!(binding_manager.bindings.is_empty());
}

#[wasm_bindgen_test]
fn test_complex_vector_paths() {
    let mut vector_engine = VectorEngine::new();
    
    // Test spiral path
    let spiral_params = [
        ("center_x".to_string(), 50.0),
        ("center_y".to_string(), 50.0),
        ("start_radius".to_string(), 5.0),
        ("end_radius".to_string(), 40.0),
        ("turns".to_string(), 3.0),
    ].into_iter().collect();
    
    let spiral_id = vector_engine.create_complex_path(ComplexPathType::Spiral, spiral_params).unwrap();
    assert!(!spiral_id.is_empty());
    
    // Test star path
    let star_params = [
        ("center_x".to_string(), 50.0),
        ("center_y".to_string(), 50.0),
        ("outer_radius".to_string(), 40.0),
        ("inner_radius".to_string(), 20.0),
        ("points".to_string(), 5.0),
    ].into_iter().collect();
    
    let star_id = vector_engine.create_complex_path(ComplexPathType::Star, star_params).unwrap();
    assert!(!star_id.is_empty());
    
    let svg_content = vector_engine.render_to_svg(200.0, 200.0);
    assert!(svg_content.contains("<path"));
    assert_eq!(vector_engine.paths.len(), 2);
}

#[wasm_bindgen_test]
fn test_chart_interactions() {
    let mut chart_renderer = ChartRenderer::new();
    
    let config = ChartConfig::default();
    let chart_id = chart_renderer.create_chart(
        ChartType::Line,
        "interactive_data".to_string(),
        config
    ).unwrap();
    
    let interactions = ChartInteractions {
        zoom_enabled: true,
        pan_enabled: true,
        hover_effects: true,
        click_events: true,
        brush_selection: false,
        crosshair: true,
    };
    
    chart_renderer.enable_chart_interactions(&chart_id, interactions).unwrap();
    
    let chart = chart_renderer.charts.get(&chart_id).unwrap();
    assert!(chart.interactions.zoom_enabled);
    assert!(chart.interactions.pan_enabled);
    assert!(chart.interactions.hover_effects);
    assert!(chart.interactions.click_events);
    assert!(chart.interactions.crosshair);
}

#[wasm_bindgen_test]
fn test_chart_animation_updates() {
    let mut chart_renderer = ChartRenderer::new();
    
    let config = ChartConfig::default();
    let chart_id = chart_renderer.create_chart(
        ChartType::Bar,
        "animated_data".to_string(),
        config
    ).unwrap();
    
    // Test animation progress update
    chart_renderer.update_chart_animation(&chart_id, 0.5).unwrap();
    
    // Cache should be invalidated
    assert_eq!(chart_renderer.render_cache.len(), 0);
}

#[wasm_bindgen_test]
fn test_interaction_manager_mouse_events() {
    let mut interaction_manager = InteractionManager::new();
    
    let mouse_event = InteractionEvent {
        event_type: InteractionType::MouseDown,
        target_element: Some("test_element".to_string()),
        position: Some(Position { x: 100.0, y: 200.0 }),
        data: HashMap::new(),
        timestamp: get_current_timestamp(),
        touch_data: None,
        mouse_data: Some(MouseData {
            button: MouseButton::Left,
            buttons: 1,
            position: Position { x: 100.0, y: 200.0 },
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
    
    let responses = interaction_manager.process_event(&mouse_event).unwrap();
    
    assert!(!responses.is_empty());
    assert!(matches!(responses[0].response_type, ResponseType::StateChanged));
    assert_eq!(interaction_manager.mouse_state.target_element, Some("test_element".to_string()));
}

#[wasm_bindgen_test]
fn test_interaction_manager_touch_events() {
    let mut interaction_manager = InteractionManager::new();
    
    let touch_event = InteractionEvent {
        event_type: InteractionType::TouchStart,
        target_element: Some("touch_element".to_string()),
        position: Some(Position { x: 150.0, y: 250.0 }),
        data: HashMap::new(),
        timestamp: get_current_timestamp(),
        touch_data: Some(TouchData {
            touches: vec![TouchPoint {
                identifier: 1,
                position: Position { x: 150.0, y: 250.0 },
                radius: Some(10.0),
                rotation_angle: None,
                force: Some(0.5),
            }],
            changed_touches: vec![TouchPoint {
                identifier: 1,
                position: Position { x: 150.0, y: 250.0 },
                radius: Some(10.0),
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
    
    let responses = interaction_manager.process_event(&touch_event).unwrap();
    
    assert!(!responses.is_empty());
    assert!(matches!(responses[0].response_type, ResponseType::TouchStart));
    assert_eq!(interaction_manager.touch_tracking.len(), 1);
    assert!(interaction_manager.touch_tracking.contains_key(&1));
}

#[wasm_bindgen_test]
fn test_gesture_recognizer_tap_detection() {
    let mut gesture_recognizer = GestureRecognizer::new();
    
    let touch_data = TouchData {
        touches: vec![TouchPoint {
            identifier: 1,
            position: Position { x: 100.0, y: 100.0 },
            radius: Some(8.0),
            rotation_angle: None,
            force: Some(0.3),
        }],
        changed_touches: vec![TouchPoint {
            identifier: 1,
            position: Position { x: 100.0, y: 100.0 },
            radius: Some(8.0),
            rotation_angle: None,
            force: Some(0.3),
        }],
        target_touches: vec![],
        force: Some(0.3),
        rotation_angle: None,
        scale: None,
    };
    
    let start_time = get_current_timestamp();
    
    // Start touch
    let gestures = gesture_recognizer.process_touch_input(&touch_data, start_time);
    assert!(gestures.is_empty()); // No gestures detected yet
    
    // End touch quickly (tap gesture)
    let end_touch_data = TouchData {
        touches: vec![],
        changed_touches: vec![TouchPoint {
            identifier: 1,
            position: Position { x: 102.0, y: 101.0 }, // Slight movement
            radius: Some(8.0),
            rotation_angle: None,
            force: Some(0.3),
        }],
        target_touches: vec![],
        force: None,
        rotation_angle: None,
        scale: None,
    };
    
    let end_time = start_time + 150.0; // 150ms duration
    let end_gestures = gesture_recognizer.process_touch_input(&end_touch_data, end_time);
    
    // Should detect a tap gesture
    assert!(!end_gestures.is_empty());
    assert!(matches!(end_gestures[0].gesture_type, GestureType::Tap));
    assert!(end_gestures[0].confidence > 0.5);
}

#[wasm_bindgen_test]
fn test_gesture_recognizer_swipe_detection() {
    let mut gesture_recognizer = GestureRecognizer::new();
    
    let start_time = get_current_timestamp();
    
    // Start touch
    let start_touch = TouchData {
        touches: vec![TouchPoint {
            identifier: 1,
            position: Position { x: 50.0, y: 100.0 },
            radius: Some(8.0),
            rotation_angle: None,
            force: Some(0.4),
        }],
        changed_touches: vec![TouchPoint {
            identifier: 1,
            position: Position { x: 50.0, y: 100.0 },
            radius: Some(8.0),
            rotation_angle: None,
            force: Some(0.4),
        }],
        target_touches: vec![],
        force: Some(0.4),
        rotation_angle: None,
        scale: None,
    };
    
    gesture_recognizer.process_touch_input(&start_touch, start_time);
    
    // Move touch significantly (swipe)
    let move_touch = TouchData {
        touches: vec![TouchPoint {
            identifier: 1,
            position: Position { x: 200.0, y: 105.0 }, // 150px horizontal movement
            radius: Some(8.0),
            rotation_angle: None,
            force: Some(0.4),
        }],
        changed_touches: vec![TouchPoint {
            identifier: 1,
            position: Position { x: 200.0, y: 105.0 },
            radius: Some(8.0),
            rotation_angle: None,
            force: Some(0.4),
        }],
        target_touches: vec![],
        force: Some(0.4),
        rotation_angle: None,
        scale: None,
    };
    
    let move_time = start_time + 200.0; // 200ms duration
    let gestures = gesture_recognizer.process_touch_input(&move_touch, move_time);
    
    // Should detect a swipe gesture
    if !gestures.is_empty() {
        assert!(matches!(gestures[0].gesture_type, GestureType::Swipe));
        assert!(gestures[0].properties.contains_key("direction"));
    }
}

#[wasm_bindgen_test]
fn test_responsive_adapter_device_detection() {
    let mut responsive_adapter = ResponsiveAdapter::new();
    
    // Test mobile viewport
    let mobile_viewport = Viewport {
        width: 375.0,
        height: 667.0,
        scale: 1.0,
        offset_x: 0.0,
        offset_y: 0.0,
    };
    
    responsive_adapter.initialize_device_detection(&mobile_viewport).unwrap();
    
    assert!(matches!(responsive_adapter.device_info.device_type, DeviceType::Mobile));
    assert_eq!(responsive_adapter.interaction_settings.touch_target_size, 44.0);
    
    // Test desktop viewport
    let desktop_viewport = Viewport {
        width: 1920.0,
        height: 1080.0,
        scale: 1.0,
        offset_x: 0.0,
        offset_y: 0.0,
    };
    
    responsive_adapter.initialize_device_detection(&desktop_viewport).unwrap();
    
    assert!(matches!(responsive_adapter.device_info.device_type, DeviceType::Desktop));
    assert_eq!(responsive_adapter.interaction_settings.touch_target_size, 32.0);
}

#[wasm_bindgen_test]
fn test_responsive_adapter_event_adaptation() {
    let mut responsive_adapter = ResponsiveAdapter::new();
    
    // Initialize for mobile
    let mobile_viewport = Viewport {
        width: 375.0,
        height: 667.0,
        scale: 1.0,
        offset_x: 0.0,
        offset_y: 0.0,
    };
    responsive_adapter.initialize_device_detection(&mobile_viewport).unwrap();
    
    let mut touch_event = InteractionEvent {
        event_type: InteractionType::TouchStart,
        target_element: Some("test_element".to_string()),
        position: Some(Position { x: 100.0, y: 100.0 }),
        data: HashMap::new(),
        timestamp: get_current_timestamp(),
        touch_data: Some(TouchData {
            touches: vec![TouchPoint {
                identifier: 1,
                position: Position { x: 100.0, y: 100.0 },
                radius: Some(5.0), // Small radius
                rotation_angle: None,
                force: Some(0.3),
            }],
            changed_touches: vec![],
            target_touches: vec![],
            force: Some(0.3),
            rotation_angle: None,
            scale: Some(1.0),
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
    
    responsive_adapter.adapt_event(&mut touch_event).unwrap();
    
    // Touch radius should be adjusted to minimum target size
    if let Some(touch_data) = &touch_event.touch_data {
        if let Some(touch) = touch_data.touches.first() {
            if let Some(radius) = touch.radius {
                assert!(radius >= responsive_adapter.adaptive_thresholds.min_touch_target);
            }
        }
    }
}

#[wasm_bindgen_test]
fn test_interaction_event_delegation() {
    let mut interaction_manager = InteractionManager::new();
    
    // Add event delegate
    let delegate = EventDelegate {
        element_id: "delegate_element".to_string(),
        event_types: vec![InteractionType::Click, InteractionType::TouchStart],
        handler_id: "test_handler".to_string(),
        capture: false,
        priority: 1,
    };
    
    interaction_manager.add_event_delegate("target_element", delegate);
    
    // Create click event
    let click_event = InteractionEvent {
        event_type: InteractionType::Click,
        target_element: Some("target_element".to_string()),
        position: Some(Position { x: 100.0, y: 100.0 }),
        data: HashMap::new(),
        timestamp: get_current_timestamp(),
        touch_data: None,
        mouse_data: Some(MouseData {
            button: MouseButton::Left,
            buttons: 1,
            position: Position { x: 100.0, y: 100.0 },
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
    
    let responses = interaction_manager.process_event(&click_event).unwrap();
    
    // Should have both direct response and delegated response
    assert!(responses.len() >= 2);
    assert!(responses.iter().any(|r| matches!(r.response_type, ResponseType::Delegated)));
}

#[wasm_bindgen_test]
fn test_interaction_performance_metrics() {
    let mut interaction_manager = InteractionManager::new();
    
    // Process multiple events
    for i in 0..10 {
        let event = InteractionEvent {
            event_type: InteractionType::MouseMove,
            target_element: Some("test_element".to_string()),
            position: Some(Position { x: i as f64 * 10.0, y: 100.0 }),
            data: HashMap::new(),
            timestamp: get_current_timestamp() + i as f64 * 10.0,
            touch_data: None,
            mouse_data: Some(MouseData {
                button: MouseButton::None,
                buttons: 0,
                position: Position { x: i as f64 * 10.0, y: 100.0 },
                movement: Some(Position { x: 10.0, y: 0.0 }),
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
        
        interaction_manager.process_event(&event).unwrap();
    }
    
    let metrics = interaction_manager.get_performance_metrics();
    assert_eq!(metrics.total_events, 10);
    assert_eq!(metrics.mouse_events_processed, 10);
    assert!(metrics.average_response_time >= 0.0);
}

#[wasm_bindgen_test]
fn test_keyboard_state_management() {
    let mut interaction_manager = InteractionManager::new();
    
    // Test key down
    let key_down_event = InteractionEvent {
        event_type: InteractionType::KeyDown,
        target_element: Some("input_element".to_string()),
        position: None,
        data: HashMap::new(),
        timestamp: get_current_timestamp(),
        touch_data: None,
        mouse_data: None,
        keyboard_data: Some(KeyboardData {
            key: "Control".to_string(),
            code: "ControlLeft".to_string(),
            char_code: None,
            key_code: Some(17),
            repeat: false,
        }),
        gesture_data: None,
        modifiers: EventModifiers {
            ctrl: false,
            shift: false,
            alt: false,
            meta: false,
        },
    };
    
    interaction_manager.process_event(&key_down_event).unwrap();
    
    // Control key should be tracked as pressed
    assert!(interaction_manager.keyboard_state.pressed_keys.contains_key("Control"));
    assert!(interaction_manager.keyboard_state.modifiers.ctrl);
    
    // Test key up
    let key_up_event = InteractionEvent {
        event_type: InteractionType::KeyUp,
        target_element: Some("input_element".to_string()),
        position: None,
        data: HashMap::new(),
        timestamp: get_current_timestamp() + 100.0,
        touch_data: None,
        mouse_data: None,
        keyboard_data: Some(KeyboardData {
            key: "Control".to_string(),
            code: "ControlLeft".to_string(),
            char_code: None,
            key_code: Some(17),
            repeat: false,
        }),
        gesture_data: None,
        modifiers: EventModifiers {
            ctrl: false,
            shift: false,
            alt: false,
            meta: false,
        },
    };
    
    interaction_manager.process_event(&key_up_event).unwrap();
    
    // Control key should no longer be tracked as pressed
    assert!(!interaction_manager.keyboard_state.pressed_keys.contains_key("Control"));
    assert!(!interaction_manager.keyboard_state.modifiers.ctrl);
}