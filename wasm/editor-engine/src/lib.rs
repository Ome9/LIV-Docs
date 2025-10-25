use wasm_bindgen::prelude::*;
use serde::{Deserialize, Serialize};
use std::collections::HashMap;

// When the `wee_alloc` feature is enabled, use `wee_alloc` as the global allocator
#[cfg(feature = "wee_alloc")]
#[global_allocator]
static ALLOC: wee_alloc::WeeAlloc = wee_alloc::WeeAlloc::INIT;

// Set up panic hook for better error messages
#[wasm_bindgen(start)]
pub fn main() {
    console_error_panic_hook::set_once();
}

// Editor-specific data structures

#[derive(Serialize, Deserialize, Clone, Debug)]
pub struct EditorState {
    pub document: DocumentState,
    pub selection: Selection,
    pub history: EditHistory,
    pub validation_state: ValidationState,
    pub preview_mode: PreviewMode,
}

#[derive(Serialize, Deserialize, Clone, Debug)]
pub struct DocumentState {
    pub elements: Vec<EditableElement>,
    pub styles: HashMap<String, StyleRule>,
    pub scripts: HashMap<String, ScriptModule>,
    pub assets: HashMap<String, AssetReference>,
    pub metadata: DocumentMetadata,
}

#[derive(Serialize, Deserialize, Clone, Debug)]
pub struct EditableElement {
    pub id: String,
    pub element_type: ElementType,
    pub properties: HashMap<String, serde_json::Value>,
    pub children: Vec<String>,
    pub parent: Option<String>,
    pub locked: bool,
    pub visible: bool,
    pub bounds: BoundingBox,
}

#[derive(Serialize, Deserialize, Clone, Debug)]
pub enum ElementType {
    Text,
    Image,
    Chart,
    Animation,
    Container,
    Interactive,
    Vector,
    Embed,
}

#[derive(Serialize, Deserialize, Clone, Debug)]
pub struct StyleRule {
    pub selector: String,
    pub properties: HashMap<String, String>,
    pub media_queries: Vec<MediaQuery>,
    pub pseudo_classes: Vec<String>,
}

#[derive(Serialize, Deserialize, Clone, Debug)]
pub struct MediaQuery {
    pub condition: String,
    pub properties: HashMap<String, String>,
}

#[derive(Serialize, Deserialize, Clone, Debug)]
pub struct ScriptModule {
    pub name: String,
    pub content: String,
    pub module_type: ScriptType,
    pub dependencies: Vec<String>,
    pub exports: Vec<String>,
}

#[derive(Serialize, Deserialize, Clone, Debug)]
pub enum ScriptType {
    JavaScript,
    TypeScript,
    WASM,
}

#[derive(Serialize, Deserialize, Clone, Debug)]
pub struct AssetReference {
    pub id: String,
    pub name: String,
    pub asset_type: AssetType,
    pub size: u64,
    pub hash: String,
    pub url: Option<String>,
}

#[derive(Serialize, Deserialize, Clone, Debug)]
pub enum AssetType {
    Image,
    Font,
    Audio,
    Video,
    Data,
    Document,
}

#[derive(Serialize, Deserialize, Clone, Debug)]
pub struct DocumentMetadata {
    pub title: String,
    pub author: String,
    pub description: String,
    pub tags: Vec<String>,
    pub created: String, // ISO 8601 timestamp
    pub modified: String, // ISO 8601 timestamp
    pub version: String,
}

#[derive(Serialize, Deserialize, Clone, Debug)]
pub struct Selection {
    pub selected_elements: Vec<String>,
    pub selection_type: SelectionType,
    pub bounds: Option<BoundingBox>,
    pub anchor_point: Option<Position>,
}

#[derive(Serialize, Deserialize, Clone, Debug)]
pub enum SelectionType {
    None,
    Single,
    Multiple,
    Text,
    Area,
}

#[derive(Serialize, Deserialize, Clone, Debug)]
pub struct EditHistory {
    pub operations: Vec<EditOperation>,
    pub current_index: usize,
    pub max_operations: usize,
}

#[derive(Serialize, Deserialize, Clone, Debug)]
pub struct EditOperation {
    pub id: String,
    pub operation_type: OperationType,
    pub timestamp: f64,
    pub data: serde_json::Value,
    pub inverse_data: serde_json::Value,
}

#[derive(Serialize, Deserialize, Clone, Debug)]
pub enum OperationType {
    Create,
    Update,
    Delete,
    Move,
    Style,
    Transform,
    Batch,
}

#[derive(Serialize, Deserialize, Clone, Debug)]
pub struct ValidationState {
    pub errors: Vec<ValidationError>,
    pub warnings: Vec<ValidationWarning>,
    pub is_valid: bool,
    pub last_validated: f64,
}

#[derive(Serialize, Deserialize, Clone, Debug)]
pub struct ValidationError {
    pub element_id: Option<String>,
    pub error_type: ErrorType,
    pub message: String,
    pub line: Option<u32>,
    pub column: Option<u32>,
}

#[derive(Serialize, Deserialize, Clone, Debug)]
pub struct ValidationWarning {
    pub element_id: Option<String>,
    pub warning_type: WarningType,
    pub message: String,
    pub suggestion: Option<String>,
}

#[derive(Serialize, Deserialize, Clone, Debug)]
pub enum ErrorType {
    Syntax,
    Semantic,
    Security,
    Performance,
    Accessibility,
}

#[derive(Serialize, Deserialize, Clone, Debug)]
pub enum WarningType {
    Performance,
    Accessibility,
    Compatibility,
    BestPractice,
}

#[derive(Serialize, Deserialize, Clone, Debug)]
pub enum PreviewMode {
    Design,
    Preview,
    Code,
    Split,
}

// Geometry types
#[derive(Serialize, Deserialize, Clone, Debug)]
pub struct Position {
    pub x: f64,
    pub y: f64,
}

#[derive(Serialize, Deserialize, Clone, Debug)]
pub struct BoundingBox {
    pub x: f64,
    pub y: f64,
    pub width: f64,
    pub height: f64,
}

// Editor operation results
#[derive(Serialize, Deserialize, Clone, Debug)]
pub struct EditorResult {
    pub success: bool,
    pub message: Option<String>,
    pub data: Option<serde_json::Value>,
    pub errors: Vec<String>,
}

#[derive(Serialize, Deserialize, Clone, Debug)]
pub struct ValidationReport {
    pub is_valid: bool,
    pub errors: Vec<ValidationError>,
    pub warnings: Vec<ValidationWarning>,
    pub performance_score: f64,
    pub accessibility_score: f64,
}

// Render update for preview
#[derive(Serialize, Deserialize, Clone, Debug)]
pub struct RenderUpdate {
    pub dom_operations: Vec<DOMOperation>,
    pub style_changes: Vec<StyleChange>,
    pub script_updates: Vec<ScriptUpdate>,
    pub asset_updates: Vec<AssetUpdate>,
}

#[derive(Serialize, Deserialize, Clone, Debug)]
pub enum DOMOperation {
    Create {
        element_id: String,
        tag: String,
        parent_id: Option<String>,
        attributes: HashMap<String, String>,
    },
    Update {
        element_id: String,
        attributes: HashMap<String, String>,
        content: Option<String>,
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
    pub selector: String,
    pub property: String,
    pub value: String,
    pub important: bool,
}

#[derive(Serialize, Deserialize, Clone, Debug)]
pub struct ScriptUpdate {
    pub module_name: String,
    pub content: String,
    pub action: ScriptAction,
}

#[derive(Serialize, Deserialize, Clone, Debug)]
pub enum ScriptAction {
    Add,
    Update,
    Remove,
    Reload,
}

#[derive(Serialize, Deserialize, Clone, Debug)]
pub struct AssetUpdate {
    pub asset_id: String,
    pub action: AssetAction,
    pub data: Option<Vec<u8>>,
}

#[derive(Serialize, Deserialize, Clone, Debug)]
pub enum AssetAction {
    Add,
    Update,
    Remove,
    Optimize,
}

// Default implementations
impl Default for EditorState {
    fn default() -> Self {
        Self {
            document: DocumentState::default(),
            selection: Selection::default(),
            history: EditHistory::default(),
            validation_state: ValidationState::default(),
            preview_mode: PreviewMode::Design,
        }
    }
}

impl Default for DocumentState {
    fn default() -> Self {
        Self {
            elements: Vec::new(),
            styles: HashMap::new(),
            scripts: HashMap::new(),
            assets: HashMap::new(),
            metadata: DocumentMetadata::default(),
        }
    }
}

impl Default for DocumentMetadata {
    fn default() -> Self {
        Self {
            title: "Untitled Document".to_string(),
            author: "Unknown".to_string(),
            description: String::new(),
            tags: Vec::new(),
            created: "1970-01-01T00:00:00Z".to_string(),
            modified: "1970-01-01T00:00:00Z".to_string(),
            version: "1.0.0".to_string(),
        }
    }
}

impl Default for Selection {
    fn default() -> Self {
        Self {
            selected_elements: Vec::new(),
            selection_type: SelectionType::None,
            bounds: None,
            anchor_point: None,
        }
    }
}

impl Default for EditHistory {
    fn default() -> Self {
        Self {
            operations: Vec::new(),
            current_index: 0,
            max_operations: 100,
        }
    }
}

impl Default for ValidationState {
    fn default() -> Self {
        Self {
            errors: Vec::new(),
            warnings: Vec::new(),
            is_valid: true,
            last_validated: 0.0,
        }
    }
}

// Export for JavaScript interop
#[wasm_bindgen]
extern "C" {
    #[wasm_bindgen(js_namespace = console)]
    fn log(s: &str);
}

// Global editor state
static mut EDITOR_STATE: Option<EditorState> = None;

// Core Editor Engine Implementation
pub struct EditorEngine {
    state: EditorState,
}

impl EditorEngine {
    pub fn new() -> Self {
        Self {
            state: EditorState::default(),
        }
    }

    pub fn load_document(&mut self, document_json: &str) -> EditorResult {
        match serde_json::from_str::<serde_json::Value>(document_json) {
            Ok(doc_value) => {
                // Convert LIV document to editor document state
                let document_state = self.convert_liv_to_editor_document(&doc_value);
                self.state.document = document_state;
                
                // Reset other state
                self.state.selection = Selection::default();
                self.state.history = EditHistory::default();
                self.state.validation_state = ValidationState::default();
                
                EditorResult {
                    success: true,
                    message: Some("Document loaded successfully".to_string()),
                    data: None,
                    errors: Vec::new(),
                }
            }
            Err(e) => EditorResult {
                success: false,
                message: Some(format!("Failed to parse document: {}", e)),
                data: None,
                errors: vec![e.to_string()],
            }
        }
    }

    pub fn save_document(&self) -> EditorResult {
        // Convert editor document state back to LIV document format
        let liv_document = self.convert_editor_to_liv_document();
        
        match serde_json::to_string(&liv_document) {
            Ok(json) => EditorResult {
                success: true,
                message: Some("Document saved successfully".to_string()),
                data: Some(serde_json::Value::String(json)),
                errors: Vec::new(),
            },
            Err(e) => EditorResult {
                success: false,
                message: Some(format!("Failed to serialize document: {}", e)),
                data: None,
                errors: vec![e.to_string()],
            }
        }
    }

    pub fn create_element(&mut self, element_type: ElementType, properties: HashMap<String, serde_json::Value>) -> EditorResult {
        let element_id = format!("element_{}", self.state.document.elements.len());
        
        let element = EditableElement {
            id: element_id.clone(),
            element_type,
            properties,
            children: Vec::new(),
            parent: None,
            locked: false,
            visible: true,
            bounds: BoundingBox { x: 0.0, y: 0.0, width: 100.0, height: 100.0 },
        };

        self.state.document.elements.push(element);
        
        // Add to history
        self.add_to_history(OperationType::Create, serde_json::json!({
            "element_id": element_id,
            "element": self.state.document.elements.last().unwrap()
        }));

        EditorResult {
            success: true,
            message: Some("Element created".to_string()),
            data: Some(serde_json::json!({"element_id": element_id})),
            errors: Vec::new(),
        }
    }

    pub fn update_element(&mut self, element_id: &str, properties: HashMap<String, serde_json::Value>) -> EditorResult {
        if let Some(element) = self.state.document.elements.iter_mut().find(|e| e.id == element_id) {
            let old_properties = element.properties.clone();
            
            for (key, value) in properties {
                element.properties.insert(key, value);
            }

            // Add to history
            self.add_to_history(OperationType::Update, serde_json::json!({
                "element_id": element_id,
                "old_properties": old_properties,
                "new_properties": element.properties
            }));

            EditorResult {
                success: true,
                message: Some("Element updated".to_string()),
                data: None,
                errors: Vec::new(),
            }
        } else {
            EditorResult {
                success: false,
                message: Some("Element not found".to_string()),
                data: None,
                errors: vec!["Element not found".to_string()],
            }
        }
    }

    pub fn delete_element(&mut self, element_id: &str) -> EditorResult {
        if let Some(pos) = self.state.document.elements.iter().position(|e| e.id == element_id) {
            let element = self.state.document.elements.remove(pos);
            
            // Add to history
            self.add_to_history(OperationType::Delete, serde_json::json!({
                "element_id": element_id,
                "element": element,
                "position": pos
            }));

            EditorResult {
                success: true,
                message: Some("Element deleted".to_string()),
                data: None,
                errors: Vec::new(),
            }
        } else {
            EditorResult {
                success: false,
                message: Some("Element not found".to_string()),
                data: None,
                errors: vec!["Element not found".to_string()],
            }
        }
    }

    pub fn select_element(&mut self, element_id: &str) -> EditorResult {
        if self.state.document.elements.iter().any(|e| e.id == element_id) {
            self.state.selection.selected_elements = vec![element_id.to_string()];
            self.state.selection.selection_type = SelectionType::Single;
            
            EditorResult {
                success: true,
                message: Some("Element selected".to_string()),
                data: None,
                errors: Vec::new(),
            }
        } else {
            EditorResult {
                success: false,
                message: Some("Element not found".to_string()),
                data: None,
                errors: vec!["Element not found".to_string()],
            }
        }
    }

    pub fn undo(&mut self) -> EditorResult {
        if self.state.history.current_index > 0 {
            self.state.history.current_index -= 1;
            let operation = &self.state.history.operations[self.state.history.current_index];
            
            // Apply inverse operation
            self.apply_inverse_operation(operation);
            
            EditorResult {
                success: true,
                message: Some("Undo successful".to_string()),
                data: None,
                errors: Vec::new(),
            }
        } else {
            EditorResult {
                success: false,
                message: Some("Nothing to undo".to_string()),
                data: None,
                errors: Vec::new(),
            }
        }
    }

    pub fn redo(&mut self) -> EditorResult {
        if self.state.history.current_index < self.state.history.operations.len() {
            let operation = &self.state.history.operations[self.state.history.current_index];
            
            // Apply operation
            self.apply_operation(operation);
            self.state.history.current_index += 1;
            
            EditorResult {
                success: true,
                message: Some("Redo successful".to_string()),
                data: None,
                errors: Vec::new(),
            }
        } else {
            EditorResult {
                success: false,
                message: Some("Nothing to redo".to_string()),
                data: None,
                errors: Vec::new(),
            }
        }
    }

    pub fn validate_document(&mut self) -> ValidationReport {
        let mut errors = Vec::new();
        let mut warnings = Vec::new();

        // Basic validation
        for element in &self.state.document.elements {
            // Check for required properties
            match element.element_type {
                ElementType::Text => {
                    if !element.properties.contains_key("content") {
                        errors.push(ValidationError {
                            element_id: Some(element.id.clone()),
                            error_type: ErrorType::Semantic,
                            message: "Text element missing content property".to_string(),
                            line: None,
                            column: None,
                        });
                    }
                }
                ElementType::Image => {
                    if !element.properties.contains_key("src") {
                        errors.push(ValidationError {
                            element_id: Some(element.id.clone()),
                            error_type: ErrorType::Semantic,
                            message: "Image element missing src property".to_string(),
                            line: None,
                            column: None,
                        });
                    }
                }
                _ => {}
            }

            // Check accessibility
            if element.element_type == ElementType::Image && !element.properties.contains_key("alt") {
                warnings.push(ValidationWarning {
                    element_id: Some(element.id.clone()),
                    warning_type: WarningType::Accessibility,
                    message: "Image element missing alt text".to_string(),
                    suggestion: Some("Add alt text for accessibility".to_string()),
                });
            }
        }

        let is_valid = errors.is_empty();
        
        // Update validation state
        self.state.validation_state = ValidationState {
            errors: errors.clone(),
            warnings: warnings.clone(),
            is_valid,
            last_validated: js_sys::Date::now(),
        };

        ValidationReport {
            is_valid,
            errors,
            warnings,
            performance_score: 85.0, // Placeholder
            accessibility_score: if warnings.is_empty() { 100.0 } else { 75.0 },
        }
    }

    pub fn get_render_update(&self) -> RenderUpdate {
        let mut dom_operations = Vec::new();
        
        // Generate DOM operations for all elements
        for element in &self.state.document.elements {
            let tag = match element.element_type {
                ElementType::Text => "p",
                ElementType::Image => "img",
                ElementType::Container => "div",
                ElementType::Chart => "canvas",
                _ => "div",
            };

            let mut attributes = HashMap::new();
            attributes.insert("id".to_string(), element.id.clone());
            
            // Convert properties to attributes
            for (key, value) in &element.properties {
                if let Some(str_value) = value.as_str() {
                    attributes.insert(key.clone(), str_value.to_string());
                }
            }

            dom_operations.push(DOMOperation::Create {
                element_id: element.id.clone(),
                tag: tag.to_string(),
                parent_id: element.parent.clone(),
                attributes,
            });
        }

        RenderUpdate {
            dom_operations,
            style_changes: Vec::new(),
            script_updates: Vec::new(),
            asset_updates: Vec::new(),
        }
    }

    // Helper methods
    fn convert_liv_to_editor_document(&self, doc_value: &serde_json::Value) -> DocumentState {
        let mut document_state = DocumentState::default();
        
        // Extract metadata
        if let Some(metadata) = doc_value.get("metadata") {
            if let Some(title) = metadata.get("title").and_then(|v| v.as_str()) {
                document_state.metadata.title = title.to_string();
            }
            if let Some(author) = metadata.get("author").and_then(|v| v.as_str()) {
                document_state.metadata.author = author.to_string();
            }
        }

        // Parse HTML content into elements (simplified)
        if let Some(content) = doc_value.get("content") {
            if let Some(html) = content.get("html").and_then(|v| v.as_str()) {
                // Simple HTML parsing - in a real implementation, this would be more sophisticated
                let element = EditableElement {
                    id: "root".to_string(),
                    element_type: ElementType::Container,
                    properties: [("innerHTML".to_string(), serde_json::Value::String(html.to_string()))].into_iter().collect(),
                    children: Vec::new(),
                    parent: None,
                    locked: false,
                    visible: true,
                    bounds: BoundingBox { x: 0.0, y: 0.0, width: 800.0, height: 600.0 },
                };
                document_state.elements.push(element);
            }
        }

        document_state
    }

    fn convert_editor_to_liv_document(&self) -> serde_json::Value {
        let mut content_html = String::new();
        
        // Convert elements back to HTML (simplified)
        for element in &self.state.document.elements {
            if let Some(html) = element.properties.get("innerHTML").and_then(|v| v.as_str()) {
                content_html.push_str(html);
            }
        }

        serde_json::json!({
            "metadata": {
                "title": self.state.document.metadata.title,
                "author": self.state.document.metadata.author,
                "description": self.state.document.metadata.description,
                "version": self.state.document.metadata.version,
                "created": self.state.document.metadata.created,
                "modified": js_sys::Date::new_0().to_iso_string()
            },
            "content": {
                "html": content_html,
                "css": "",
                "interactiveSpec": "",
                "staticFallback": content_html
            }
        })
    }

    fn add_to_history(&mut self, operation_type: OperationType, data: serde_json::Value) {
        let operation = EditOperation {
            id: format!("op_{}", self.state.history.operations.len()),
            operation_type,
            timestamp: js_sys::Date::now(),
            data: data.clone(),
            inverse_data: data, // Simplified - should be actual inverse
        };

        // Remove operations after current index (for redo)
        self.state.history.operations.truncate(self.state.history.current_index);
        
        // Add new operation
        self.state.history.operations.push(operation);
        self.state.history.current_index = self.state.history.operations.len();

        // Limit history size
        if self.state.history.operations.len() > self.state.history.max_operations {
            self.state.history.operations.remove(0);
            self.state.history.current_index -= 1;
        }
    }

    fn apply_operation(&mut self, _operation: &EditOperation) {
        // Implementation would apply the operation
    }

    fn apply_inverse_operation(&mut self, _operation: &EditOperation) {
        // Implementation would apply the inverse operation
    }
}

// WASM bindings
#[wasm_bindgen]
pub fn init_editor_engine() {
    log("LIV Editor Engine initialized");
    unsafe {
        EDITOR_STATE = Some(EditorState::default());
    }
}

#[wasm_bindgen]
pub fn load_document(document_json: &str) -> String {
    let mut engine = EditorEngine::new();
    let result = engine.load_document(document_json);
    
    unsafe {
        EDITOR_STATE = Some(engine.state);
    }
    
    serde_json::to_string(&result).unwrap_or_else(|_| "{}".to_string())
}

#[wasm_bindgen]
pub fn save_document() -> String {
    unsafe {
        if let Some(ref state) = EDITOR_STATE {
            let engine = EditorEngine { state: state.clone() };
            let result = engine.save_document();
            serde_json::to_string(&result).unwrap_or_else(|_| "{}".to_string())
        } else {
            serde_json::to_string(&EditorResult {
                success: false,
                message: Some("Editor not initialized".to_string()),
                data: None,
                errors: vec!["Editor not initialized".to_string()],
            }).unwrap_or_else(|_| "{}".to_string())
        }
    }
}

#[wasm_bindgen]
pub fn create_element(element_type: &str, properties_json: &str) -> String {
    unsafe {
        if let Some(ref mut state) = EDITOR_STATE {
            let mut engine = EditorEngine { state: state.clone() };
            
            let element_type = match element_type {
                "text" => ElementType::Text,
                "image" => ElementType::Image,
                "chart" => ElementType::Chart,
                "container" => ElementType::Container,
                _ => ElementType::Container,
            };

            let properties: HashMap<String, serde_json::Value> = 
                serde_json::from_str(properties_json).unwrap_or_default();

            let result = engine.create_element(element_type, properties);
            *state = engine.state;
            
            serde_json::to_string(&result).unwrap_or_else(|_| "{}".to_string())
        } else {
            "{}".to_string()
        }
    }
}

#[wasm_bindgen]
pub fn update_element(element_id: &str, properties_json: &str) -> String {
    unsafe {
        if let Some(ref mut state) = EDITOR_STATE {
            let mut engine = EditorEngine { state: state.clone() };
            
            let properties: HashMap<String, serde_json::Value> = 
                serde_json::from_str(properties_json).unwrap_or_default();

            let result = engine.update_element(element_id, properties);
            *state = engine.state;
            
            serde_json::to_string(&result).unwrap_or_else(|_| "{}".to_string())
        } else {
            "{}".to_string()
        }
    }
}

#[wasm_bindgen]
pub fn delete_element(element_id: &str) -> String {
    unsafe {
        if let Some(ref mut state) = EDITOR_STATE {
            let mut engine = EditorEngine { state: state.clone() };
            let result = engine.delete_element(element_id);
            *state = engine.state;
            
            serde_json::to_string(&result).unwrap_or_else(|_| "{}".to_string())
        } else {
            "{}".to_string()
        }
    }
}

#[wasm_bindgen]
pub fn select_element(element_id: &str) -> String {
    unsafe {
        if let Some(ref mut state) = EDITOR_STATE {
            let mut engine = EditorEngine { state: state.clone() };
            let result = engine.select_element(element_id);
            *state = engine.state;
            
            serde_json::to_string(&result).unwrap_or_else(|_| "{}".to_string())
        } else {
            "{}".to_string()
        }
    }
}

#[wasm_bindgen]
pub fn undo() -> String {
    unsafe {
        if let Some(ref mut state) = EDITOR_STATE {
            let mut engine = EditorEngine { state: state.clone() };
            let result = engine.undo();
            *state = engine.state;
            
            serde_json::to_string(&result).unwrap_or_else(|_| "{}".to_string())
        } else {
            "{}".to_string()
        }
    }
}

#[wasm_bindgen]
pub fn redo() -> String {
    unsafe {
        if let Some(ref mut state) = EDITOR_STATE {
            let mut engine = EditorEngine { state: state.clone() };
            let result = engine.redo();
            *state = engine.state;
            
            serde_json::to_string(&result).unwrap_or_else(|_| "{}".to_string())
        } else {
            "{}".to_string()
        }
    }
}

#[wasm_bindgen]
pub fn validate_document() -> String {
    unsafe {
        if let Some(ref mut state) = EDITOR_STATE {
            let mut engine = EditorEngine { state: state.clone() };
            let result = engine.validate_document();
            *state = engine.state;
            
            serde_json::to_string(&result).unwrap_or_else(|_| "{}".to_string())
        } else {
            "{}".to_string()
        }
    }
}

#[wasm_bindgen]
pub fn get_render_update() -> String {
    unsafe {
        if let Some(ref state) = EDITOR_STATE {
            let engine = EditorEngine { state: state.clone() };
            let result = engine.get_render_update();
            
            serde_json::to_string(&result).unwrap_or_else(|_| "{}".to_string())
        } else {
            "{}".to_string()
        }
    }
}