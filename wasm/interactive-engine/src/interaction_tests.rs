use super::*;
use wasm_bindgen_test::*;

wasm_bindgen_test_configure!(run_in_browser);

// Test user interaction handling and render update efficiency

#[wasm_bindgen_test]
fn test_comprehensive_mouse_interaction_handling() {
    let permissions = WASMPermissions {
        memory_limit: 5 * 1024 * 1024,
        allowed_imports: vec!["console".to_string()],
        cpu_time_limit: 10000,
        allow_networking: false,
        allow_file_system: false,
        allowed_interactions: vec![
            "create_element".to_string(),
            "MouseDown".to_string(),
            "MouseUp".to_string(),
            "MouseMove".to_string(),
            "Click".to_string(),
            "DoubleClick".to_string(),
        ],
        max_data_size: 2 * 1024 * 1024,
        max_elements: 200,
    };

    let mut engine = InteractiveEngine::new(permissions).unwrap();

    // Create interactive elements
    let element_id = engine.create_element(ElementType::Interactive, HashMap::new()).unwrap();

    // Test mouse down event
    let mouse_down_event = InteractionEvent {
        event_type: InteractionType::MouseDown,
        target_element: Some(element_id.clone()),
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

    let render_update = engine.process_interaction(mouse_down_event).unwrap();
    assert!(!render_update.dom_operations.is_empty() || !render_update.style_changes.is_empty());

    // Verify interaction state was updated
    if let Some(state) = engine.get_interaction_state(&element_id) {
        assert!(matches!(state.state_type, InteractionStateType::Pressed));
    }

    // Test mouse move (drag) events
    for i in 1..10 {
        let mouse_move_event = InteractionEvent {
            event_type: InteractionType::MouseMove,
            target_element: Some(element_id.clone()),
            position: Some(Position { x: 100.0 + (i as f64 * 5.0), y: 100.0 + (i as f64 * 2.0) }),
            data: HashMap::new(),
            timestamp: get_current_timestamp() + (i as f64 * 16.67), // 60fps intervals
            touch_data: None,
            mouse_data: Some(MouseData {
                button: MouseButton::Left,
                buttons: 1,
                position: Position { x: 100.0 + (i as f64 * 5.0), y: 100.0 + (i as f64 * 2.0) },
                movement: Some(Position { x: 5.0, y: 2.0 }),
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

        let render_update = engine.process_interaction(mouse_move_event).unwrap();
        
        // Should generate drag updates after threshold
        if i > 2 {
            assert!(!render_update.dom_operations.is_empty() || !render_update.style_changes.is_empty());
        }
    }

    // Test mouse up event
    let mouse_up_event = InteractionEvent {
        event_type: InteractionType::MouseUp,
        target_element: Some(element_id.clone()),
        position: Some(Position { x: 150.0, y: 120.0 }),
        data: HashMap::new(),
        timestamp: get_current_timestamp() + 200.0,
        touch_data: None,
        mouse_data: Some(MouseData {
            button: MouseButton::Left,
            buttons: 0,
            position: Position { x: 150.0, y: 120.0 },
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

    let render_update = engine.process_interaction(mouse_up_event).unwrap();
    assert!(!render_update.dom_operations.is_empty() || !render_update.style_changes.is_empty());

    // Verify interaction state was updated
    if let Some(state) = engine.get_interaction_state(&element_id) {
        assert!(matches!(state.state_type, InteractionStateType::Hover));
    }
}

#[wasm_bindgen_test]
fn test_comprehensive_touch_interaction_handling() {
    let permissions = WASMPermissions {
        memory_limit: 5 * 1024 * 1024,
        allowed_imports: vec!["console".to_string()],
        cpu_time_limit: 10000,
        allow_networking: false,
        allow_file_system: false,
        allowed_interactions: vec![
            "create_element".to_string(),
            "TouchStart".to_string(),
            "TouchMove".to_string(),
            "TouchEnd".to_string(),
            "Tap".to_string(),
            "Pinch".to_string(),
        ],
        max_data_size: 2 * 1024 * 1024,
        max_elements: 200,
    };

    let mut engine = InteractiveEngine::new(permissions).unwrap();

    // Create interactive element
    let element_id = engine.create_element(ElementType::Interactive, HashMap::new()).unwrap();

    // Test single touch start
    let touch_start_event = InteractionEvent {
        event_type: InteractionType::TouchStart,
        target_element: Some(element_id.clone()),
        position: Some(Position { x: 200.0, y: 200.0 }),
        data: HashMap::new(),
        timestamp: get_current_timestamp(),
        touch_data: Some(TouchData {
            touches: vec![TouchPoint {
                identifier: 1,
                position: Position { x: 200.0, y: 200.0 },
                radius: Some(8.0),
                rotation_angle: None,
                force: Some(0.5),
            }],
            changed_touches: vec![TouchPoint {
                identifier: 1,
                position: Position { x: 200.0, y: 200.0 },
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

    let render_update = engine.process_interaction(touch_start_event).unwrap();
    assert!(!render_update.dom_operations.is_empty() || !render_update.style_changes.is_empty());

    // Test touch move events
    for i in 1..20 {
        let touch_move_event = InteractionEvent {
            event_type: InteractionType::TouchMove,
            target_element: Some(element_id.clone()),
            position: Some(Position { x: 200.0 + (i as f64 * 2.0), y: 200.0 + (i as f64) }),
            data: HashMap::new(),
            timestamp: get_current_timestamp() + (i as f64 * 16.67),
            touch_data: Some(TouchData {
                touches: vec![TouchPoint {
                    identifier: 1,
                    position: Position { x: 200.0 + (i as f64 * 2.0), y: 200.0 + (i as f64) },
                    radius: Some(8.0),
                    rotation_angle: None,
                    force: Some(0.5 + (i as f64 * 0.01)),
                }],
                changed_touches: vec![TouchPoint {
                    identifier: 1,
                    position: Position { x: 200.0 + (i as f64 * 2.0), y: 200.0 + (i as f64) },
                    radius: Some(8.0),
                    rotation_angle: None,
                    force: Some(0.5 + (i as f64 * 0.01)),
                }],
                target_touches: vec![],
                force: Some(0.5 + (i as f64 * 0.01)),
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

        let render_update = engine.process_interaction(touch_move_event).unwrap();
        assert!(!render_update.dom_operations.is_empty() || !render_update.style_changes.is_empty());
    }

    // Test touch end (should trigger tap gesture)
    let touch_end_event = InteractionEvent {
        event_type: InteractionType::TouchEnd,
        target_element: Some(element_id.clone()),
        position: Some(Position { x: 240.0, y: 220.0 }),
        data: HashMap::new(),
        timestamp: get_current_timestamp() + 400.0,
        touch_data: Some(TouchData {
            touches: vec![],
            changed_touches: vec![TouchPoint {
                identifier: 1,
                position: Position { x: 240.0, y: 220.0 },
                radius: Some(8.0),
                rotation_angle: None,
                force: Some(0.7),
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

    let render_update = engine.process_interaction(touch_end_event).unwrap();
    assert!(!render_update.dom_operations.is_empty() || !render_update.style_changes.is_empty());
}

#[wasm_bindgen_test]
fn test_multi_touch_gesture_recognition() {
    let permissions = WASMPermissions {
        memory_limit: 5 * 1024 * 1024,
        allowed_imports: vec!["console".to_string()],
        cpu_time_limit: 10000,
        allow_networking: false,
        allow_file_system: false,
        allowed_interactions: vec![
            "create_element".to_string(),
            "TouchStart".to_string(),
            "TouchMove".to_string(),
            "TouchEnd".to_string(),
            "Pinch".to_string(),
            "Rotate".to_string(),
        ],
        max_data_size: 2 * 1024 * 1024,
        max_elements: 200,
    };

    let mut engine = InteractiveEngine::new(permissions).unwrap();

    // Create interactive element
    let element_id = engine.create_element(ElementType::Interactive, HashMap::new()).unwrap();

    // Test two-finger pinch gesture
    let pinch_start_event = InteractionEvent {
        event_type: InteractionType::TouchStart,
        target_element: Some(element_id.clone()),
        position: Some(Position { x: 150.0, y: 150.0 }),
        data: HashMap::new(),
        timestamp: get_current_timestamp(),
        touch_data: Some(TouchData {
            touches: vec![
                TouchPoint {
                    identifier: 1,
                    position: Position { x: 100.0, y: 150.0 },
                    radius: Some(8.0),
                    rotation_angle: None,
                    force: Some(0.5),
                },
                TouchPoint {
                    identifier: 2,
                    position: Position { x: 200.0, y: 150.0 },
                    radius: Some(8.0),
                    rotation_angle: None,
                    force: Some(0.5),
                },
            ],
            changed_touches: vec![
                TouchPoint {
                    identifier: 1,
                    position: Position { x: 100.0, y: 150.0 },
                    radius: Some(8.0),
                    rotation_angle: None,
                    force: Some(0.5),
                },
                TouchPoint {
                    identifier: 2,
                    position: Position { x: 200.0, y: 150.0 },
                    radius: Some(8.0),
                    rotation_angle: None,
                    force: Some(0.5),
                },
            ],
            target_touches: vec![],
            force: Some(0.5),
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

    let render_update = engine.process_interaction(pinch_start_event).unwrap();
    assert!(!render_update.dom_operations.is_empty() || !render_update.style_changes.is_empty());

    // Simulate pinch-in gesture (fingers moving closer)
    for i in 1..10 {
        let distance_reduction = i as f64 * 5.0;
        let scale = 1.0 - (distance_reduction / 100.0);
        
        let pinch_move_event = InteractionEvent {
            event_type: InteractionType::TouchMove,
            target_element: Some(element_id.clone()),
            position: Some(Position { x: 150.0, y: 150.0 }),
            data: HashMap::new(),
            timestamp: get_current_timestamp() + (i as f64 * 50.0),
            touch_data: Some(TouchData {
                touches: vec![
                    TouchPoint {
                        identifier: 1,
                        position: Position { x: 100.0 + distance_reduction, y: 150.0 },
                        radius: Some(8.0),
                        rotation_angle: None,
                        force: Some(0.5),
                    },
                    TouchPoint {
                        identifier: 2,
                        position: Position { x: 200.0 - distance_reduction, y: 150.0 },
                        radius: Some(8.0),
                        rotation_angle: None,
                        force: Some(0.5),
                    },
                ],
                changed_touches: vec![
                    TouchPoint {
                        identifier: 1,
                        position: Position { x: 100.0 + distance_reduction, y: 150.0 },
                        radius: Some(8.0),
                        rotation_angle: None,
                        force: Some(0.5),
                    },
                    TouchPoint {
                        identifier: 2,
                        position: Position { x: 200.0 - distance_reduction, y: 150.0 },
                        radius: Some(8.0),
                        rotation_angle: None,
                        force: Some(0.5),
                    },
                ],
                target_touches: vec![],
                force: Some(0.5),
                rotation_angle: None,
                scale: Some(scale),
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

        let render_update = engine.process_interaction(pinch_move_event).unwrap();
        
        // Should generate pinch gesture updates
        if i > 3 {
            assert!(!render_update.dom_operations.is_empty() || !render_update.style_changes.is_empty());
        }
    }
}

#[wasm_bindgen_test]
fn test_keyboard_interaction_handling() {
    let permissions = WASMPermissions {
        memory_limit: 5 * 1024 * 1024,
        allowed_imports: vec!["console".to_string()],
        cpu_time_limit: 10000,
        allow_networking: false,
        allow_file_system: false,
        allowed_interactions: vec![
            "create_element".to_string(),
            "KeyDown".to_string(),
            "KeyUp".to_string(),
            "KeyPress".to_string(),
        ],
        max_data_size: 2 * 1024 * 1024,
        max_elements: 200,
    };

    let mut engine = InteractiveEngine::new(permissions).unwrap();

    // Create interactive element
    let element_id = engine.create_element(ElementType::Interactive, HashMap::new()).unwrap();

    // Test key down events with modifiers
    let key_combinations = vec![
        ("Control", "ControlLeft", 17),
        ("Shift", "ShiftLeft", 16),
        ("a", "KeyA", 65),
        ("Enter", "Enter", 13),
        ("Escape", "Escape", 27),
    ];

    for (key, code, key_code) in key_combinations {
        let key_down_event = InteractionEvent {
            event_type: InteractionType::KeyDown,
            target_element: Some(element_id.clone()),
            position: None,
            data: HashMap::new(),
            timestamp: get_current_timestamp(),
            touch_data: None,
            mouse_data: None,
            keyboard_data: Some(KeyboardData {
                key: key.to_string(),
                code: code.to_string(),
                char_code: None,
                key_code: Some(key_code),
                repeat: false,
            }),
            gesture_data: None,
            modifiers: EventModifiers {
                ctrl: key == "Control",
                shift: key == "Shift",
                alt: false,
                meta: false,
            },
        };

        let render_update = engine.process_interaction(key_down_event).unwrap();
        assert!(!render_update.dom_operations.is_empty() || !render_update.style_changes.is_empty());

        // Test corresponding key up event
        let key_up_event = InteractionEvent {
            event_type: InteractionType::KeyUp,
            target_element: Some(element_id.clone()),
            position: None,
            data: HashMap::new(),
            timestamp: get_current_timestamp() + 100.0,
            touch_data: None,
            mouse_data: None,
            keyboard_data: Some(KeyboardData {
                key: key.to_string(),
                code: code.to_string(),
                char_code: None,
                key_code: Some(key_code),
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

        let render_update = engine.process_interaction(key_up_event).unwrap();
        assert!(!render_update.dom_operations.is_empty() || !render_update.style_changes.is_empty());
    }
}

#[wasm_bindgen_test]
fn test_event_delegation_system() {
    let permissions = WASMPermissions {
        memory_limit: 5 * 1024 * 1024,
        allowed_imports: vec!["console".to_string()],
        cpu_time_limit: 10000,
        allow_networking: false,
        allow_file_system: false,
        allowed_interactions: vec![
            "create_element".to_string(),
            "modify_element".to_string(),
            "Click".to_string(),
            "TouchStart".to_string(),
        ],
        max_data_size: 2 * 1024 * 1024,
        max_elements: 200,
    };

    let mut engine = InteractiveEngine::new(permissions).unwrap();

    // Create parent and child elements
    let parent_id = engine.create_element(ElementType::Container, HashMap::new()).unwrap();
    let child_id = engine.create_element(ElementType::Interactive, HashMap::new()).unwrap();

    // Set up event delegation
    let delegate = EventDelegate {
        element_id: parent_id.clone(),
        event_types: vec![InteractionType::Click, InteractionType::TouchStart],
        handler_id: "parent_handler".to_string(),
        capture: false,
        priority: 1,
    };

    engine.add_interaction_delegate(&child_id, delegate).unwrap();

    // Test click event on child (should delegate to parent)
    let click_event = InteractionEvent {
        event_type: InteractionType::Click,
        target_element: Some(child_id.clone()),
        position: Some(Position { x: 50.0, y: 50.0 }),
        data: HashMap::new(),
        timestamp: get_current_timestamp(),
        touch_data: None,
        mouse_data: Some(MouseData {
            button: MouseButton::Left,
            buttons: 1,
            position: Position { x: 50.0, y: 50.0 },
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
    assert!(!render_update.dom_operations.is_empty() || !render_update.style_changes.is_empty());

    // Test touch event on child (should also delegate to parent)
    let touch_event = InteractionEvent {
        event_type: InteractionType::TouchStart,
        target_element: Some(child_id.clone()),
        position: Some(Position { x: 50.0, y: 50.0 }),
        data: HashMap::new(),
        timestamp: get_current_timestamp() + 100.0,
        touch_data: Some(TouchData {
            touches: vec![TouchPoint {
                identifier: 1,
                position: Position { x: 50.0, y: 50.0 },
                radius: Some(8.0),
                rotation_angle: None,
                force: Some(0.5),
            }],
            changed_touches: vec![TouchPoint {
                identifier: 1,
                position: Position { x: 50.0, y: 50.0 },
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

    let render_update = engine.process_interaction(touch_event).unwrap();
    assert!(!render_update.dom_operations.is_empty() || !render_update.style_changes.is_empty());

    // Remove delegation and test
    engine.remove_interaction_delegate(&child_id, "parent_handler").unwrap();

    let click_event_after_removal = InteractionEvent {
        event_type: InteractionType::Click,
        target_element: Some(child_id.clone()),
        position: Some(Position { x: 50.0, y: 50.0 }),
        data: HashMap::new(),
        timestamp: get_current_timestamp() + 200.0,
        touch_data: None,
        mouse_data: Some(MouseData {
            button: MouseButton::Left,
            buttons: 1,
            position: Position { x: 50.0, y: 50.0 },
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

    let render_update = engine.process_interaction(click_event_after_removal).unwrap();
    // Should still work but without delegation
    assert!(!render_update.dom_operations.is_empty() || !render_update.style_changes.is_empty());
}

#[wasm_bindgen_test]
fn test_responsive_interaction_adaptation() {
    let permissions = WASMPermissions {
        memory_limit: 5 * 1024 * 1024,
        allowed_imports: vec!["console".to_string()],
        cpu_time_limit: 10000,
        allow_networking: false,
        allow_file_system: false,
        allowed_interactions: vec![
            "create_element".to_string(),
            "TouchStart".to_string(),
            "TouchMove".to_string(),
            "MouseMove".to_string(),
        ],
        max_data_size: 2 * 1024 * 1024,
        max_elements: 200,
    };

    let mut engine = InteractiveEngine::new(permissions).unwrap();

    // Test mobile device adaptation
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

    // Test touch event adaptation for mobile
    let touch_event = InteractionEvent {
        event_type: InteractionType::TouchStart,
        target_element: Some("test_element".to_string()),
        position: Some(Position { x: 100.0, y: 100.0 }),
        data: HashMap::new(),
        timestamp: get_current_timestamp(),
        touch_data: Some(TouchData {
            touches: vec![TouchPoint {
                identifier: 1,
                position: Position { x: 100.0, y: 100.0 },
                radius: Some(4.0), // Small radius that should be adapted
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

    let render_update = engine.process_interaction(touch_event).unwrap();
    assert!(!render_update.dom_operations.is_empty() || !render_update.style_changes.is_empty());

    // Test desktop device adaptation
    let desktop_device_info = DeviceInfo {
        device_type: DeviceType::Desktop,
        screen_size: Size { width: 1920.0, height: 1080.0 },
        pixel_density: 1.0,
        touch_support: false,
        mouse_support: true,
        keyboard_support: true,
        max_touch_points: 0,
        has_force_touch: false,
        has_hover_support: true,
    };

    engine.update_device_capabilities(desktop_device_info).unwrap();

    // Test mouse event adaptation for desktop
    let mouse_event = InteractionEvent {
        event_type: InteractionType::MouseMove,
        target_element: Some("test_element".to_string()),
        position: Some(Position { x: 200.0, y: 200.0 }),
        data: HashMap::new(),
        timestamp: get_current_timestamp() + 100.0,
        touch_data: None,
        mouse_data: Some(MouseData {
            button: MouseButton::None,
            buttons: 0,
            position: Position { x: 200.0, y: 200.0 },
            movement: Some(Position { x: 5.0, y: 3.0 }),
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

    let render_update = engine.process_interaction(mouse_event).unwrap();
    assert!(!render_update.dom_operations.is_empty() || !render_update.style_changes.is_empty());
}

#[wasm_bindgen_test]
fn test_interaction_performance_metrics() {
    let permissions = WASMPermissions {
        memory_limit: 5 * 1024 * 1024,
        allowed_imports: vec!["console".to_string()],
        cpu_time_limit: 10000,
        allow_networking: false,
        allow_file_system: false,
        allowed_interactions: vec![
            "create_element".to_string(),
            "Click".to_string(),
            "TouchStart".to_string(),
            "MouseMove".to_string(),
        ],
        max_data_size: 2 * 1024 * 1024,
        max_elements: 200,
    };

    let mut engine = InteractiveEngine::new(permissions).unwrap();

    // Create elements for interaction
    let element_id = engine.create_element(ElementType::Interactive, HashMap::new()).unwrap();

    // Generate various interactions to test metrics
    let interaction_types = vec![
        InteractionType::Click,
        InteractionType::TouchStart,
        InteractionType::MouseMove,
    ];

    for (i, interaction_type) in interaction_types.iter().cycle().take(100).enumerate() {
        let event = match interaction_type {
            InteractionType::Click => InteractionEvent {
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
            },
            InteractionType::TouchStart => InteractionEvent {
                event_type: InteractionType::TouchStart,
                target_element: Some(element_id.clone()),
                position: Some(Position { x: (i % 10) as f64 * 10.0, y: (i / 10) as f64 * 10.0 }),
                data: HashMap::new(),
                timestamp: get_current_timestamp() + (i as f64 * 10.0),
                touch_data: Some(TouchData {
                    touches: vec![TouchPoint {
                        identifier: 1,
                        position: Position { x: (i % 10) as f64 * 10.0, y: (i / 10) as f64 * 10.0 },
                        radius: Some(8.0),
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
            },
            InteractionType::MouseMove => InteractionEvent {
                event_type: InteractionType::MouseMove,
                target_element: Some(element_id.clone()),
                position: Some(Position { x: (i % 10) as f64 * 10.0, y: (i / 10) as f64 * 10.0 }),
                data: HashMap::new(),
                timestamp: get_current_timestamp() + (i as f64 * 10.0),
                touch_data: None,
                mouse_data: Some(MouseData {
                    button: MouseButton::None,
                    buttons: 0,
                    position: Position { x: (i % 10) as f64 * 10.0, y: (i / 10) as f64 * 10.0 },
                    movement: Some(Position { x: 1.0, y: 1.0 }),
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
            },
            _ => continue,
        };

        let _render_update = engine.process_interaction(event).unwrap();
    }

    // Check interaction metrics
    let metrics = engine.get_interaction_metrics();
    assert_eq!(metrics.total_events, 100);
    assert!(metrics.average_response_time > 0.0);
    assert!(metrics.events_per_second > 0.0);
    assert!(metrics.mouse_events_processed > 0);
    assert!(metrics.touch_points_processed > 0);
}