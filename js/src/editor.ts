import { LIVDocument } from './document';
import { LIVRenderer } from './renderer';
import { SandboxInterface } from './sandbox-interface';

import { LIVError, LIVErrorType } from './errors';

/**
 * Validation result interface
 */
interface ValidationResult {
    line: number;
    column: number;
    message: string;
    severity: 'error' | 'warning' | 'info';
    code?: string;
}

/**
 * Syntax highlighter for code editors
 */
class SyntaxHighlighter {
    private highlightRules: Map<string, RegExp[]> = new Map();

    constructor() {
        this.setupHighlightRules();
    }

    private setupHighlightRules(): void {
        // HTML highlighting rules
        this.highlightRules.set('html', [
            /(&lt;\/?)([a-zA-Z][a-zA-Z0-9]*)(.*?)(&gt;)/g, // HTML tags
            /(&quot;[^&quot;]*&quot;|'[^']*')/g, // Attributes
            /(&lt;!--.*?--&gt;)/g // Comments
        ]);

        // CSS highlighting rules
        this.highlightRules.set('css', [
            /([a-zA-Z-]+)(\s*:\s*)([^;]+)(;)/g, // Properties
            /(\/\*.*?\*\/)/g, // Comments
            /(\.[a-zA-Z][a-zA-Z0-9-]*|#[a-zA-Z][a-zA-Z0-9-]*)/g // Selectors
        ]);

        // JavaScript highlighting rules
        this.highlightRules.set('javascript', [
            /\b(function|var|let|const|if|else|for|while|return|class|extends)\b/g, // Keywords
            /(\/\/.*$|\/\*.*?\*\/)/gm, // Comments
            /(".*?"|'.*?'|`.*?`)/g // Strings
        ]);

        // JSON highlighting rules
        this.highlightRules.set('json', [
            /(".*?")(\s*:\s*)/g, // Keys
            /(".*?")/g, // String values
            /\b(true|false|null)\b/g, // Literals
            /\b\d+\.?\d*\b/g // Numbers
        ]);
    }

    highlight(code: string, language: string): string {
        const rules = this.highlightRules.get(language);
        if (!rules) return this.escapeHtml(code);

        let highlighted = this.escapeHtml(code);
        
        rules.forEach((rule, index) => {
            highlighted = highlighted.replace(rule, (match) => {
                return this.wrapWithSpan(match, `syntax-${language}-${index}`);
            });
        });

        return highlighted;
    }

    private escapeHtml(text: string): string {
        const div = document.createElement('div');
        div.textContent = text;
        return div.innerHTML;
    }

    private wrapWithSpan(text: string, className: string): string {
        return `<span class="${className}">${text}</span>`;
    }
}

/**
 * Code validator for different languages
 */
class CodeValidator {
    private document: LIVDocument | null = null;

    setDocument(document: LIVDocument | null): void {
        this.document = document;
    }

    async validateHTML(code: string): Promise<ValidationResult[]> {
        const results: ValidationResult[] = [];
        const lines = code.split('\n');

        // Basic HTML validation
        const tagStack: Array<{tag: string, line: number}> = [];
        const selfClosingTags = new Set(['img', 'br', 'hr', 'input', 'meta', 'link']);

        lines.forEach((line, lineIndex) => {
            // Check for unclosed tags
            const openTags = line.match(/<([a-zA-Z][a-zA-Z0-9]*)[^>]*>/g);
            const closeTags = line.match(/<\/([a-zA-Z][a-zA-Z0-9]*)[^>]*>/g);

            if (openTags) {
                openTags.forEach(tag => {
                    const tagName = tag.match(/<([a-zA-Z][a-zA-Z0-9]*)/)?.[1];
                    if (tagName && !selfClosingTags.has(tagName.toLowerCase())) {
                        tagStack.push({tag: tagName, line: lineIndex + 1});
                    }
                });
            }

            if (closeTags) {
                closeTags.forEach(tag => {
                    const tagName = tag.match(/<\/([a-zA-Z][a-zA-Z0-9]*)/)?.[1];
                    if (tagName) {
                        const lastOpen = tagStack.pop();
                        if (!lastOpen || lastOpen.tag !== tagName) {
                            results.push({
                                line: lineIndex + 1,
                                column: line.indexOf(tag) + 1,
                                message: `Mismatched closing tag: expected </${lastOpen?.tag || 'unknown'}>, found </${tagName}>`,
                                severity: 'error',
                                code: 'HTML001'
                            });
                        }
                    }
                });
            }

            // Check for invalid attributes
            if (line.includes('onclick') || line.includes('onload')) {
                results.push({
                    line: lineIndex + 1,
                    column: line.indexOf('on') + 1,
                    message: 'Inline event handlers are not allowed for security reasons',
                    severity: 'error',
                    code: 'HTML002'
                });
            }
        });

        // Check for unclosed tags
        tagStack.forEach(openTag => {
            results.push({
                line: openTag.line,
                column: 1,
                message: `Unclosed tag: <${openTag.tag}>`,
                severity: 'error',
                code: 'HTML003'
            });
        });

        return results;
    }

    async validateCSS(code: string): Promise<ValidationResult[]> {
        const results: ValidationResult[] = [];
        const lines = code.split('\n');

        let braceCount = 0;
        lines.forEach((line, lineIndex) => {
            // Count braces
            const openBraces = (line.match(/{/g) || []).length;
            const closeBraces = (line.match(/}/g) || []).length;
            braceCount += openBraces - closeBraces;

            // Check for invalid properties
            const propertyMatch = line.match(/([a-zA-Z-]+)\s*:\s*([^;]+);?/);
            if (propertyMatch) {
                const property = propertyMatch[1];
                const value = propertyMatch[2];

                // Check for potentially dangerous properties
                if (property === 'behavior' || property.startsWith('-moz-binding')) {
                    results.push({
                        line: lineIndex + 1,
                        column: line.indexOf(property) + 1,
                        message: `Property '${property}' is not allowed for security reasons`,
                        severity: 'error',
                        code: 'CSS001'
                    });
                }

                // Check for javascript: URLs
                if (value.includes('javascript:')) {
                    results.push({
                        line: lineIndex + 1,
                        column: line.indexOf('javascript:') + 1,
                        message: 'JavaScript URLs are not allowed in CSS',
                        severity: 'error',
                        code: 'CSS002'
                    });
                }
            }
        });

        if (braceCount !== 0) {
            results.push({
                line: lines.length,
                column: 1,
                message: `Mismatched braces: ${braceCount > 0 ? 'missing closing' : 'extra closing'} braces`,
                severity: 'error',
                code: 'CSS003'
            });
        }

        return results;
    }

    async validateJavaScript(code: string): Promise<ValidationResult[]> {
        const results: ValidationResult[] = [];
        const lines = code.split('\n');

        lines.forEach((line, lineIndex) => {
            // Check for dangerous functions
            const dangerousFunctions = ['eval', 'Function', 'setTimeout', 'setInterval'];
            dangerousFunctions.forEach(func => {
                if (line.includes(func + '(')) {
                    results.push({
                        line: lineIndex + 1,
                        column: line.indexOf(func) + 1,
                        message: `Function '${func}' is not allowed in sandboxed environment`,
                        severity: 'error',
                        code: 'JS001'
                    });
                }
            });

            // Check for DOM access that might be restricted
            if (line.includes('document.') || line.includes('window.')) {
                results.push({
                    line: lineIndex + 1,
                    column: Math.max(line.indexOf('document.'), line.indexOf('window.')) + 1,
                    message: 'Direct DOM/window access may be restricted in sandbox',
                    severity: 'warning',
                    code: 'JS002'
                });
            }
        });

        return results;
    }

    async validateJSON(code: string): Promise<ValidationResult[]> {
        const results: ValidationResult[] = [];

        try {
            JSON.parse(code);
        } catch (error) {
            const errorMessage = error instanceof Error ? error.message : 'Invalid JSON';
            const lineMatch = errorMessage.match(/line (\d+)/);
            const line = lineMatch ? parseInt(lineMatch[1]) : 1;

            results.push({
                line,
                column: 1,
                message: errorMessage,
                severity: 'error',
                code: 'JSON001'
            });
        }

        return results;
    }

    async validateManifest(code: string): Promise<ValidationResult[]> {
        const results: ValidationResult[] = [];

        try {
            const manifest = JSON.parse(code);
            
            // Validate required fields
            if (!manifest.version) {
                results.push({
                    line: 1,
                    column: 1,
                    message: 'Manifest must have a version field',
                    severity: 'error',
                    code: 'MANIFEST001'
                });
            }

            if (!manifest.metadata) {
                results.push({
                    line: 1,
                    column: 1,
                    message: 'Manifest must have metadata section',
                    severity: 'error',
                    code: 'MANIFEST002'
                });
            } else {
                if (!manifest.metadata.title) {
                    results.push({
                        line: 1,
                        column: 1,
                        message: 'Manifest metadata must have a title',
                        severity: 'error',
                        code: 'MANIFEST003'
                    });
                }
            }

            if (!manifest.security) {
                results.push({
                    line: 1,
                    column: 1,
                    message: 'Manifest must have security policy',
                    severity: 'error',
                    code: 'MANIFEST004'
                });
            }

        } catch (error) {
            results.push({
                line: 1,
                column: 1,
                message: 'Invalid JSON in manifest',
                severity: 'error',
                code: 'MANIFEST005'
            });
        }

        return results;
    }
}

/**
 * WYSIWYG Editor for LIV documents
 * Provides visual editing capabilities with live preview
 */
export class LIVEditor {
    private document: LIVDocument | null = null;
    private renderer: LIVRenderer;
    private sandbox: SandboxInterface;

    private editorContainer: HTMLElement;
    private previewContainer: HTMLElement;
    private toolbarContainer: HTMLElement;
    private propertiesContainer: HTMLElement;
    private isInitialized = false;
    private selectedElement: HTMLElement | null = null;
    private editMode: 'visual' | 'source' = 'visual';
    private sourceEditor: HTMLTextAreaElement | null = null;
    private isDirty = false;
    
    // Source editor enhancements
    private sourceEditors: Map<string, HTMLTextAreaElement> = new Map();
    private currentSourceTab: string = 'html';
    private validationResults: Map<string, ValidationResult[]> = new Map();
    private syntaxHighlighter: SyntaxHighlighter | null = null;
    private codeValidator: CodeValidator | null = null;
    
    // Visual editing enhancements
    private draggedElement: HTMLElement | null = null;
    private isResizing = false;
    private elementIdCounter = 0;
    private visualStylePanel: HTMLElement | null = null;

    constructor(
        editorContainer: HTMLElement, 
        previewContainer: HTMLElement,
        toolbarContainer: HTMLElement,
        propertiesContainer: HTMLElement
    ) {
        this.editorContainer = editorContainer;
        this.previewContainer = previewContainer;
        this.toolbarContainer = toolbarContainer;
        this.propertiesContainer = propertiesContainer;
        this.renderer = new LIVRenderer({ 
            container: previewContainer,
            permissions: {} as any // Will be set from document
        });
        this.sandbox = new SandboxInterface({
            securityPolicy: {
                wasmPermissions: {
                    memoryLimit: 64 * 1024 * 1024,
                    cpuTimeLimit: 5000,
                    allowNetworking: false,
                    allowFileSystem: false,
                    allowedImports: ['env']
                },
                jsPermissions: {
                    executionMode: 'sandboxed',
                    allowedAPIs: ['dom'],
                    domAccess: 'write'
                },
                networkPolicy: {
                    allowOutbound: false,
                    allowedHosts: [],
                    allowedPorts: []
                },
                storagePolicy: {
                    allowLocalStorage: true,
                    allowSessionStorage: true,
                    allowIndexedDB: false,
                    allowCookies: false
                }
            },
            enableLogging: true,
            enableMetrics: true,
            timeoutMs: 30000,
            maxMemoryMB: 64
        });

    }

    /**
     * Initialize the editor with a document
     */
    async initialize(document?: LIVDocument): Promise<void> {
        try {
            await this.sandbox.initialize();
            
            if (document) {
                this.document = document;
                await this.loadDocument();
            } else {
                await this.createNewDocument();
            }
            
            this.setupEditorUI();
            this.setupEventHandlers();
            this.isInitialized = true;
        } catch (error) {
            throw new LIVError(
                LIVErrorType.VALIDATION,
                `Failed to initialize editor: ${error instanceof Error ? error.message : 'Unknown error'}`
            );
        }
    }

    /**
     * Load document into editor
     */
    private async loadDocument(): Promise<void> {
        if (!this.document) {
            throw new LIVError(LIVErrorType.INVALID_FILE, 'No document to load');
        }

        try {
            // Render document in preview pane
            await this.renderer.renderDocument(this.document);
            
            // Update source editors
            this.loadDocumentIntoSourceEditors();
            
            // Update properties panel
            this.updatePropertiesPanel();
        } catch (error) {
            throw new LIVError(
                LIVErrorType.PARSING,
                `Failed to load document: ${error instanceof Error ? error.message : 'Unknown error'}`
            );
        }
    }

    /**
     * Create a new empty document
     */
    private async createNewDocument(): Promise<void> {
        try {
            // Create basic document structure
            const htmlContent = `<!DOCTYPE html>
<html lang="en">
<head>
    <meta charset="UTF-8">
    <meta name="viewport" content="width=device-width, initial-scale=1.0">
    <title>New LIV Document</title>
    <style>
        body {
            font-family: -apple-system, BlinkMacSystemFont, 'Segoe UI', Roboto, sans-serif;
            margin: 0;
            padding: 20px;
            line-height: 1.6;
        }
        .container {
            max-width: 800px;
            margin: 0 auto;
        }
    </style>
</head>
<body>
    <div class="container">
        <h1>New LIV Document</h1>
        <p>Start editing your document here...</p>
    </div>
</body>
</html>`;
            
            this.document = await this.createDocumentFromHTML(htmlContent);
            if (this.document) {
                await this.renderer.renderDocument(this.document);
            }
            this.markDirty();
        } catch (error) {
            throw new LIVError(
                LIVErrorType.VALIDATION,
                `Failed to create new document: ${error instanceof Error ? error.message : 'Unknown error'}`
            );
        }
    }

    /**
     * Set up the editor UI
     */
    private setupEditorUI(): void {
        // Setup toolbar
        this.setupToolbar();
        
        // Setup editor content area
        this.setupEditorContent();
        
        // Setup properties panel
        this.setupPropertiesPanel();
        
        // Setup visual editing enhancements
        this.injectEditorStyles();
        this.setupDragAndDrop();
        this.setupVisualStylePanel();
        this.setupResizeHandles();
    }
    
    /**
     * Setup the toolbar with editing tools
     */
    private setupToolbar(): void {
        this.toolbarContainer.innerHTML = `
            <div class="editor-toolbar">
                <div class="toolbar-group">
                    <button id="mode-visual" class="toolbar-btn ${this.editMode === 'visual' ? 'active' : ''}">
                        <span class="icon">👁</span> Visual
                    </button>
                    <button id="mode-source" class="toolbar-btn ${this.editMode === 'source' ? 'active' : ''}">
                        <span class="icon">📝</span> Source
                    </button>
                </div>
                
                <div class="toolbar-separator"></div>
                
                <div class="toolbar-group">
                    <button id="format-bold" class="toolbar-btn" title="Bold">
                        <span class="icon">B</span>
                    </button>
                    <button id="format-italic" class="toolbar-btn" title="Italic">
                        <span class="icon">I</span>
                    </button>
                    <button id="format-underline" class="toolbar-btn" title="Underline">
                        <span class="icon">U</span>
                    </button>
                </div>
                
                <div class="toolbar-separator"></div>
                
                <div class="toolbar-group">
                    <button id="insert-heading" class="toolbar-btn" title="Insert Heading">
                        <span class="icon">H1</span>
                    </button>
                    <button id="insert-paragraph" class="toolbar-btn" title="Insert Paragraph">
                        <span class="icon">¶</span>
                    </button>
                    <button id="insert-image" class="toolbar-btn" title="Insert Image">
                        <span class="icon">🖼</span>
                    </button>
                    <button id="insert-link" class="toolbar-btn" title="Insert Link">
                        <span class="icon">🔗</span>
                    </button>
                </div>
                
                <div class="toolbar-separator"></div>
                
                <div class="toolbar-group">
                    <button id="insert-container" class="toolbar-btn" title="Insert Container">
                        <span class="icon">📦</span>
                    </button>
                    <button id="insert-interactive" class="toolbar-btn" title="Insert Interactive Element">
                        <span class="icon">⚡</span>
                    </button>
                    <button id="insert-chart" class="toolbar-btn" title="Insert Chart">
                        <span class="icon">📊</span>
                    </button>
                </div>
                
                <div class="toolbar-separator"></div>
                
                <div class="toolbar-group">
                    <button id="save-document" class="toolbar-btn primary" title="Save Document">
                        <span class="icon">💾</span> Save
                    </button>
                    <button id="preview-document" class="toolbar-btn" title="Preview Document">
                        <span class="icon">👁</span> Preview
                    </button>
                </div>
            </div>
        `;
    }
    
    /**
     * Setup the editor content area
     */
    private setupEditorContent(): void {
        this.editorContainer.innerHTML = `
            <div class="editor-content-wrapper">
                <div id="visual-editor" class="visual-editor ${this.editMode === 'visual' ? 'active' : 'hidden'}">
                    <!-- Visual editing area - will be populated with rendered content -->
                </div>
                <div id="source-editor-wrapper" class="source-editor-wrapper ${this.editMode === 'source' ? 'active' : 'hidden'}">
                    ${this.createSourceEditorTabs()}
                </div>
            </div>
        `;
        
        // Initialize source editors
        this.initializeSourceEditors();
        
        // Get reference to main source editor (HTML)
        this.sourceEditor = this.sourceEditors.get('html') || null;
    }

    /**
     * Create source editor tabs HTML
     */
    private createSourceEditorTabs(): string {
        return `
            <div class="source-tabs">
                <button class="source-tab active" data-tab="html">
                    <span class="tab-icon">🌐</span>
                    <span class="tab-label">HTML</span>
                    <span class="tab-status" id="html-status"></span>
                </button>
                <button class="source-tab" data-tab="css">
                    <span class="tab-icon">🎨</span>
                    <span class="tab-label">CSS</span>
                    <span class="tab-status" id="css-status"></span>
                </button>
                <button class="source-tab" data-tab="javascript">
                    <span class="tab-icon">⚡</span>
                    <span class="tab-label">JavaScript</span>
                    <span class="tab-status" id="js-status"></span>
                </button>
                <button class="source-tab" data-tab="manifest">
                    <span class="tab-icon">📋</span>
                    <span class="tab-label">Manifest</span>
                    <span class="tab-status" id="manifest-status"></span>
                </button>
                <div class="tab-actions">
                    <button class="tab-action-btn" id="format-code" title="Format Code">
                        <span class="icon">🎯</span>
                    </button>
                    <button class="tab-action-btn" id="validate-code" title="Validate Code">
                        <span class="icon">✓</span>
                    </button>
                    <button class="tab-action-btn" id="sync-visual" title="Sync with Visual Editor">
                        <span class="icon">🔄</span>
                    </button>
                </div>
            </div>
            <div class="source-editors">
                <div class="source-editor-container active" data-editor="html">
                    <div class="editor-header">
                        <div class="editor-info">
                            <span class="line-info">Line: 1, Column: 1</span>
                            <span class="char-count">0 characters</span>
                        </div>
                        <div class="editor-actions">
                            <button class="editor-btn" id="html-find" title="Find & Replace">Find</button>
                            <button class="editor-btn" id="html-goto" title="Go to Line">Go to</button>
                        </div>
                    </div>
                    <div class="editor-wrapper">
                        <textarea class="code-editor" id="html-editor" 
                                  placeholder="Enter HTML content..." 
                                  spellcheck="false"
                                  data-language="html"></textarea>
                        <div class="syntax-overlay" id="html-overlay"></div>
                        <div class="line-numbers" id="html-lines"></div>
                    </div>
                    <div class="validation-panel" id="html-validation"></div>
                </div>
                
                <div class="source-editor-container" data-editor="css">
                    <div class="editor-header">
                        <div class="editor-info">
                            <span class="line-info">Line: 1, Column: 1</span>
                            <span class="char-count">0 characters</span>
                        </div>
                        <div class="editor-actions">
                            <button class="editor-btn" id="css-find" title="Find & Replace">Find</button>
                            <button class="editor-btn" id="css-goto" title="Go to Line">Go to</button>
                        </div>
                    </div>
                    <div class="editor-wrapper">
                        <textarea class="code-editor" id="css-editor" 
                                  placeholder="Enter CSS styles..." 
                                  spellcheck="false"
                                  data-language="css"></textarea>
                        <div class="syntax-overlay" id="css-overlay"></div>
                        <div class="line-numbers" id="css-lines"></div>
                    </div>
                    <div class="validation-panel" id="css-validation"></div>
                </div>
                
                <div class="source-editor-container" data-editor="javascript">
                    <div class="editor-header">
                        <div class="editor-info">
                            <span class="line-info">Line: 1, Column: 1</span>
                            <span class="char-count">0 characters</span>
                        </div>
                        <div class="editor-actions">
                            <button class="editor-btn" id="js-find" title="Find & Replace">Find</button>
                            <button class="editor-btn" id="js-goto" title="Go to Line">Go to</button>
                        </div>
                    </div>
                    <div class="editor-wrapper">
                        <textarea class="code-editor" id="js-editor" 
                                  placeholder="Enter JavaScript code..." 
                                  spellcheck="false"
                                  data-language="javascript"></textarea>
                        <div class="syntax-overlay" id="js-overlay"></div>
                        <div class="line-numbers" id="js-lines"></div>
                    </div>
                    <div class="validation-panel" id="js-validation"></div>
                </div>
                
                <div class="source-editor-container" data-editor="manifest">
                    <div class="editor-header">
                        <div class="editor-info">
                            <span class="line-info">Line: 1, Column: 1</span>
                            <span class="char-count">0 characters</span>
                        </div>
                        <div class="editor-actions">
                            <button class="editor-btn" id="manifest-find" title="Find & Replace">Find</button>
                            <button class="editor-btn" id="manifest-goto" title="Go to Line">Go to</button>
                        </div>
                    </div>
                    <div class="editor-wrapper">
                        <textarea class="code-editor" id="manifest-editor" 
                                  placeholder="Enter manifest JSON..." 
                                  spellcheck="false"
                                  data-language="json"></textarea>
                        <div class="syntax-overlay" id="manifest-overlay"></div>
                        <div class="line-numbers" id="manifest-lines"></div>
                    </div>
                    <div class="validation-panel" id="manifest-validation"></div>
                </div>
            </div>
        `;
    }

    /**
     * Initialize source editors
     */
    private initializeSourceEditors(): void {
        // Initialize syntax highlighter and validator
        this.syntaxHighlighter = new SyntaxHighlighter();
        this.codeValidator = new CodeValidator();
        
        // Get editor elements
        const htmlEditor = this.editorContainer.querySelector('#html-editor') as HTMLTextAreaElement;
        const cssEditor = this.editorContainer.querySelector('#css-editor') as HTMLTextAreaElement;
        const jsEditor = this.editorContainer.querySelector('#js-editor') as HTMLTextAreaElement;
        const manifestEditor = this.editorContainer.querySelector('#manifest-editor') as HTMLTextAreaElement;

        // Store editor references
        this.sourceEditors.set('html', htmlEditor);
        this.sourceEditors.set('css', cssEditor);
        this.sourceEditors.set('javascript', jsEditor);
        this.sourceEditors.set('manifest', manifestEditor);

        // Setup event handlers for each editor
        this.setupSourceEditorEvents();
        
        // Setup tab switching
        this.setupSourceTabSwitching();
        
        // Initialize line numbers and syntax highlighting
        this.sourceEditors.forEach((editor, language) => {
            this.updateLineNumbers(editor, language);
            this.updateSyntaxHighlighting(editor, language);
        });
    }
    
    /**
     * Setup the properties panel
     */
    private setupPropertiesPanel(): void {
        this.propertiesContainer.innerHTML = `
            <div class="properties-panel">
                <h3>Properties</h3>
                <div id="element-properties" class="element-properties">
                    <p class="no-selection">Select an element to edit its properties</p>
                </div>
                
                <div class="document-info">
                    <h4>Document Info</h4>
                    <div class="info-item">
                        <label>Title:</label>
                        <input type="text" id="doc-title" placeholder="Document title">
                    </div>
                    <div class="info-item">
                        <label>Author:</label>
                        <input type="text" id="doc-author" placeholder="Author name">
                    </div>
                    <div class="info-item">
                        <label>Description:</label>
                        <textarea id="doc-description" placeholder="Document description"></textarea>
                    </div>
                </div>
            </div>
        `;
    }

    /**
     * Setup event handlers for editor interactions
     */
    private setupEventHandlers(): void {
        // Toolbar button handlers
        this.setupToolbarHandlers();
        
        // Editor content handlers
        this.setupContentHandlers();
        
        // Properties panel handlers
        this.setupPropertiesHandlers();
    }
    
    /**
     * Setup toolbar button event handlers
     */
    private setupToolbarHandlers(): void {
        // Mode switching
        const visualBtn = this.toolbarContainer.querySelector('#mode-visual');
        const sourceBtn = this.toolbarContainer.querySelector('#mode-source');
        
        visualBtn?.addEventListener('click', () => this.switchMode('visual'));
        sourceBtn?.addEventListener('click', () => this.switchMode('source'));
        
        // Formatting buttons
        this.toolbarContainer.querySelector('#format-bold')?.addEventListener('click', () => this.applyFormat('bold'));
        this.toolbarContainer.querySelector('#format-italic')?.addEventListener('click', () => this.applyFormat('italic'));
        this.toolbarContainer.querySelector('#format-underline')?.addEventListener('click', () => this.applyFormat('underline'));
        
        // Insert buttons
        this.toolbarContainer.querySelector('#insert-heading')?.addEventListener('click', () => this.insertElement('heading'));
        this.toolbarContainer.querySelector('#insert-paragraph')?.addEventListener('click', () => this.insertElement('paragraph'));
        this.toolbarContainer.querySelector('#insert-image')?.addEventListener('click', () => this.insertElement('image'));
        this.toolbarContainer.querySelector('#insert-link')?.addEventListener('click', () => this.insertElement('link'));
        this.toolbarContainer.querySelector('#insert-container')?.addEventListener('click', () => this.insertElement('container'));
        this.toolbarContainer.querySelector('#insert-interactive')?.addEventListener('click', () => this.insertElement('interactive'));
        this.toolbarContainer.querySelector('#insert-chart')?.addEventListener('click', () => this.insertElement('chart'));
        
        // Action buttons
        this.toolbarContainer.querySelector('#save-document')?.addEventListener('click', () => this.saveDocument());
        this.toolbarContainer.querySelector('#preview-document')?.addEventListener('click', () => this.previewDocument());
    }
    
    /**
     * Setup content area event handlers
     */
    private setupContentHandlers(): void {
        // Visual editor click handling for element selection
        const visualEditor = this.editorContainer.querySelector('#visual-editor');
        visualEditor?.addEventListener('click', (e) => this.handleElementSelection(e));
        
        // Source editor change handling
        if (this.sourceEditor) {
            this.sourceEditor.addEventListener('input', () => {
                this.markDirty();
                this.debounceSourceUpdate();
            });
        }
    }
    
    /**
     * Setup properties panel event handlers
     */
    private setupPropertiesHandlers(): void {
        // Document info handlers
        const titleInput = this.propertiesContainer.querySelector('#doc-title') as HTMLInputElement;
        const authorInput = this.propertiesContainer.querySelector('#doc-author') as HTMLInputElement;
        const descInput = this.propertiesContainer.querySelector('#doc-description') as HTMLTextAreaElement;
        
        titleInput?.addEventListener('input', () => this.updateDocumentInfo());
        authorInput?.addEventListener('input', () => this.updateDocumentInfo());
        descInput?.addEventListener('input', () => this.updateDocumentInfo());
    }
    
    /**
     * Switch between visual and source editing modes
     */
    private switchMode(mode: 'visual' | 'source'): void {
        if (this.editMode === mode) return;
        
        this.editMode = mode;
        
        // Update toolbar buttons
        this.toolbarContainer.querySelector('#mode-visual')?.classList.toggle('active', mode === 'visual');
        this.toolbarContainer.querySelector('#mode-source')?.classList.toggle('active', mode === 'source');
        
        // Update editor visibility
        const visualEditor = this.editorContainer.querySelector('#visual-editor');
        const sourceWrapper = this.editorContainer.querySelector('#source-editor-wrapper');
        
        if (mode === 'visual') {
            visualEditor?.classList.remove('hidden');
            visualEditor?.classList.add('active');
            sourceWrapper?.classList.remove('active');
            sourceWrapper?.classList.add('hidden');
            
            // Update visual editor from source if needed
            if (this.sourceEditor && this.document) {
                this.updateDocumentFromSource();
            }
        } else {
            visualEditor?.classList.remove('active');
            visualEditor?.classList.add('hidden');
            sourceWrapper?.classList.remove('hidden');
            sourceWrapper?.classList.add('active');
            
            // Update source editor from document
            if (this.sourceEditor && this.document) {
                this.sourceEditor.value = this.document.content.html || '';
            }
        }
    }
    
    /**
     * Apply formatting to selected text
     */
    private applyFormat(format: string): void {
        if (this.editMode !== 'visual') return;
        
        try {
            document.execCommand(format, false);
            this.markDirty();
        } catch (error) {
            console.warn(`Failed to apply format ${format}:`, error);
        }
    }
    
    /**
     * Handle element selection in visual editor
     */
    private handleElementSelection(event: Event): void {
        const target = event.target as HTMLElement;
        
        // Use enhanced selection method
        this.selectElement(target);
    }
    
    /**
     * Update properties panel based on selected element
     */
    private updatePropertiesPanel(): void {
        const propertiesDiv = this.propertiesContainer.querySelector('#element-properties');
        if (!propertiesDiv) return;
        
        if (!this.selectedElement) {
            propertiesDiv.innerHTML = '<p class="no-selection">Select an element to edit its properties</p>';
            return;
        }
        
        const tagName = this.selectedElement.tagName.toLowerCase();
        
        propertiesDiv.innerHTML = `
            <h4>Element: &lt;${tagName}></h4>
            <div class="property-group">
                <label>Text Content:</label>
                <input type="text" id="prop-text" value="${this.selectedElement.textContent || ''}">
            </div>
            <div class="property-group">
                <label>CSS Class:</label>
                <input type="text" id="prop-class" value="${this.selectedElement.className || ''}">
            </div>
            <div class="property-group">
                <label>ID:</label>
                <input type="text" id="prop-id" value="${this.selectedElement.id || ''}">
            </div>
            ${this.getSpecificProperties(tagName)}
        `;
        
        // Add event listeners for property changes
        this.setupElementPropertyHandlers();
    }
    
    /**
     * Get element-specific properties based on tag type
     */
    private getSpecificProperties(tagName: string): string {
        switch (tagName) {
            case 'img':
                const img = this.selectedElement as HTMLImageElement;
                return `
                    <div class="property-group">
                        <label>Source URL:</label>
                        <input type="text" id="prop-src" value="${img.src || ''}">
                    </div>
                    <div class="property-group">
                        <label>Alt Text:</label>
                        <input type="text" id="prop-alt" value="${img.alt || ''}">
                    </div>
                `;
            case 'a':
                const link = this.selectedElement as HTMLAnchorElement;
                return `
                    <div class="property-group">
                        <label>Link URL:</label>
                        <input type="text" id="prop-href" value="${link.href || ''}">
                    </div>
                    <div class="property-group">
                        <label>Target:</label>
                        <select id="prop-target">
                            <option value="" ${!link.target ? 'selected' : ''}>Same window</option>
                            <option value="_blank" ${link.target === '_blank' ? 'selected' : ''}>New window</option>
                        </select>
                    </div>
                `;
            default:
                return '';
        }
    }
    
    /**
     * Setup event handlers for element property inputs
     */
    private setupElementPropertyHandlers(): void {
        const propertiesDiv = this.propertiesContainer.querySelector('#element-properties');
        if (!propertiesDiv || !this.selectedElement) return;
        
        // Text content
        const textInput = propertiesDiv.querySelector('#prop-text') as HTMLInputElement;
        textInput?.addEventListener('input', () => {
            if (this.selectedElement) {
                this.selectedElement.textContent = textInput.value;
                this.markDirty();
            }
        });
        
        // CSS class
        const classInput = propertiesDiv.querySelector('#prop-class') as HTMLInputElement;
        classInput?.addEventListener('input', () => {
            if (this.selectedElement) {
                this.selectedElement.className = classInput.value;
                this.markDirty();
            }
        });
        
        // ID
        const idInput = propertiesDiv.querySelector('#prop-id') as HTMLInputElement;
        idInput?.addEventListener('input', () => {
            if (this.selectedElement) {
                this.selectedElement.id = idInput.value;
                this.markDirty();
            }
        });
        
        // Element-specific properties
        this.setupSpecificPropertyHandlers();
    }
    
    /**
     * Setup handlers for element-specific properties
     */
    private setupSpecificPropertyHandlers(): void {
        if (!this.selectedElement) return;
        
        const tagName = this.selectedElement.tagName.toLowerCase();
        const propertiesDiv = this.propertiesContainer.querySelector('#element-properties');
        
        if (tagName === 'img') {
            const srcInput = propertiesDiv?.querySelector('#prop-src') as HTMLInputElement;
            const altInput = propertiesDiv?.querySelector('#prop-alt') as HTMLInputElement;
            
            srcInput?.addEventListener('input', () => {
                if (this.selectedElement) {
                    (this.selectedElement as HTMLImageElement).src = srcInput.value;
                    this.markDirty();
                }
            });
            
            altInput?.addEventListener('input', () => {
                if (this.selectedElement) {
                    (this.selectedElement as HTMLImageElement).alt = altInput.value;
                    this.markDirty();
                }
            });
        } else if (tagName === 'a') {
            const hrefInput = propertiesDiv?.querySelector('#prop-href') as HTMLInputElement;
            const targetSelect = propertiesDiv?.querySelector('#prop-target') as HTMLSelectElement;
            
            hrefInput?.addEventListener('input', () => {
                if (this.selectedElement) {
                    (this.selectedElement as HTMLAnchorElement).href = hrefInput.value;
                    this.markDirty();
                }
            });
            
            targetSelect?.addEventListener('change', () => {
                if (this.selectedElement) {
                    (this.selectedElement as HTMLAnchorElement).target = targetSelect.value;
                    this.markDirty();
                }
            });
        }
    }
    
    /**
     * Mark document as dirty (modified)
     */
    private markDirty(): void {
        this.isDirty = true;
        // Update save button state
        const saveBtn = this.toolbarContainer.querySelector('#save-document');
        saveBtn?.classList.add('dirty');
    }
    
    /**
     * Mark document as clean (saved)
     */
    private markClean(): void {
        this.isDirty = false;
        const saveBtn = this.toolbarContainer.querySelector('#save-document');
        saveBtn?.classList.remove('dirty');
    }
    
    /**
     * Debounced source update to prevent excessive re-rendering
     */
    private debounceSourceUpdate = this.debounce(() => {
        this.updateDocumentFromSource();
    }, 500);
    
    /**
     * Update document from source editor content
     */
    private async updateDocumentFromSource(): Promise<void> {
        if (!this.sourceEditor || !this.document) return;
        
        try {
            const htmlContent = this.sourceEditor.value;
            this.document = await this.createDocumentFromHTML(htmlContent);
            
            // Update visual editor if in visual mode
            if (this.editMode === 'visual') {
                await this.renderer.renderDocument(this.document);
            }
        } catch (error) {
            console.error('Failed to update document from source:', error);
        }
    }
    
    /**
     * Update document info from properties panel
     */
    private updateDocumentInfo(): void {
        if (!this.document) return;
        
        const titleInput = this.propertiesContainer.querySelector('#doc-title') as HTMLInputElement;
        const authorInput = this.propertiesContainer.querySelector('#doc-author') as HTMLInputElement;
        const descInput = this.propertiesContainer.querySelector('#doc-description') as HTMLTextAreaElement;
        
        // Update document metadata
        const metadata = this.document.getMetadata();
        if (titleInput?.value) metadata.title = titleInput.value;
        if (authorInput?.value) metadata.author = authorInput.value;
        if (descInput?.value) metadata.description = descInput.value;
        
        this.markDirty();
    }
    
    /**
     * Debounce utility function
     */
    private debounce(func: Function, wait: number): Function {
        let timeout: NodeJS.Timeout;
        return function executedFunction(...args: any[]) {
            const later = () => {
                clearTimeout(timeout);
                func(...args);
            };
            clearTimeout(timeout);
            timeout = setTimeout(later, wait);
        };
    }

    /**
     * Insert an element at the current cursor position
     */
    insertElement(elementType: string, properties?: Record<string, any>): void {
        if (!this.isInitialized) {
            throw new LIVError(LIVErrorType.VALIDATION, 'Editor not initialized');
        }

        if (this.editMode !== 'visual') {
            // In source mode, insert HTML directly
            this.insertHTMLInSource(elementType, properties);
            return;
        }

        const element = this.createElementWithWASM(elementType, properties);
        
        // Insert element at cursor or append to selected container
        this.insertElementIntoDOM(element);
        
        // Make the new element draggable
        element.draggable = true;
        element.addEventListener('dragstart', this.handleDragStart.bind(this));
        element.addEventListener('dragend', this.handleDragEnd.bind(this));
        
        // Select the newly created element
        this.selectElement(element);
        
        this.markDirty();
    }

    /**
     * Create element with WASM backend integration
     */
    private createElementWithWASM(elementType: string, properties?: Record<string, any>): HTMLElement {
        let element: HTMLElement;
        const elementId = this.generateElementId();
        
        switch (elementType) {
            case 'heading':
                element = document.createElement('h2');
                element.textContent = properties?.text || 'New Heading';
                element.className = 'editable-heading';
                break;
            case 'paragraph':
                element = document.createElement('p');
                element.textContent = properties?.text || 'New paragraph text...';
                element.className = 'editable-paragraph';
                break;
            case 'image':
                element = document.createElement('img');
                (element as HTMLImageElement).src = properties?.src || 'https://via.placeholder.com/300x200';
                (element as HTMLImageElement).alt = properties?.alt || 'Image';
                element.className = 'editable-image';
                element.style.maxWidth = '100%';
                element.style.height = 'auto';
                break;
            case 'link':
                element = document.createElement('a');
                (element as HTMLAnchorElement).href = properties?.href || '#';
                element.textContent = properties?.text || 'Link text';
                element.className = 'editable-link';
                break;
            case 'interactive':
                element = document.createElement('div');
                element.className = 'interactive-element editable-interactive';
                element.innerHTML = '<p>Interactive element - configure in properties panel</p>';
                element.style.border = '2px dashed #007bff';
                element.style.padding = '20px';
                element.style.textAlign = 'center';
                break;
            case 'chart':
                element = document.createElement('div');
                element.className = 'chart-container editable-chart';
                element.innerHTML = '<canvas width="400" height="300"></canvas>';
                element.style.border = '1px solid #ddd';
                element.style.padding = '10px';
                break;
            case 'container':
                element = document.createElement('div');
                element.className = 'container editable-container';
                element.style.minHeight = '100px';
                element.style.border = '1px dashed #ccc';
                element.style.padding = '20px';
                element.innerHTML = '<p style="color: #999; text-align: center;">Drop elements here</p>';
                break;
            default:
                throw new LIVError(LIVErrorType.VALIDATION, `Unknown element type: ${elementType}`);
        }
        
        // Set unique ID
        element.id = elementId;
        
        // Add common editing attributes
        element.setAttribute('data-element-type', elementType);
        element.setAttribute('data-editable', 'true');
        
        // Notify WASM backend of element creation
        this.notifyWASMElementCreated(element, elementType, properties);
        
        return element;
    }

    /**
     * Insert element into DOM at appropriate location
     */
    private insertElementIntoDOM(element: HTMLElement): void {
        const visualEditor = this.editorContainer.querySelector('#visual-editor');
        if (!visualEditor) return;

        // If there's a selected element, try to insert relative to it
        if (this.selectedElement) {
            const parent = this.selectedElement.parentElement;
            if (parent && this.isValidDropTarget(parent)) {
                parent.insertBefore(element, this.selectedElement.nextSibling);
                return;
            }
        }

        // Otherwise, append to the visual editor
        visualEditor.appendChild(element);
    }

    /**
     * Notify WASM backend of element creation
     */
    private notifyWASMElementCreated(element: HTMLElement, elementType: string, properties?: Record<string, any>): void {
        const creationData = {
            elementId: element.id,
            elementType,
            properties: properties || {},
            operation: 'create'
        };
        
        this.sandbox.sendEvent('element-created', creationData).catch(error => {
            console.warn('Failed to notify WASM of element creation:', error);
        });
    }
    
    /**
     * Insert HTML in source mode
     */
    private insertHTMLInSource(elementType: string, properties?: Record<string, any>): void {
        if (!this.sourceEditor) return;
        
        let html = '';
        
        switch (elementType) {
            case 'heading':
                html = '<h2>New Heading</h2>';
                break;
            case 'paragraph':
                html = '<p>New paragraph text...</p>';
                break;
            case 'image':
                html = `<img src="${properties?.src || 'https://via.placeholder.com/300x200'}" alt="${properties?.alt || 'Image'}">`;
                break;
            case 'link':
                html = `<a href="${properties?.href || '#'}">${properties?.text || 'Link text'}</a>`;
                break;
            case 'interactive':
                html = '<div class="interactive-element"><p>Interactive element</p></div>';
                break;
            case 'chart':
                html = '<div class="chart-container"><canvas width="400" height="300"></canvas></div>';
                break;
        }
        
        // Insert at cursor position
        const start = this.sourceEditor.selectionStart;
        const end = this.sourceEditor.selectionEnd;
        const value = this.sourceEditor.value;
        
        this.sourceEditor.value = value.substring(0, start) + html + value.substring(end);
        this.sourceEditor.selectionStart = this.sourceEditor.selectionEnd = start + html.length;
        
        this.markDirty();
    }

    /**
     * Update properties of the selected element
     */
    updateElementProperties(properties: Record<string, any>): void {
        if (!this.isInitialized || !this.selectedElement) {
            throw new LIVError(LIVErrorType.VALIDATION, 'Editor not initialized or no element selected');
        }

        // Apply properties to selected element
        Object.entries(properties).forEach(([key, value]) => {
            switch (key) {
                case 'textContent':
                    this.selectedElement!.textContent = value;
                    break;
                case 'className':
                    this.selectedElement!.className = value;
                    break;
                case 'id':
                    this.selectedElement!.id = value;
                    break;
                default:
                    this.selectedElement!.setAttribute(key, value);
            }
        });
        
        this.markDirty();
        this.updatePropertiesPanel();
    }

    /**
     * Get the current document
     */
    getDocument(): LIVDocument | null {
        return this.document;
    }
    
    /**
     * Check if document has unsaved changes
     */
    isDirtyDocument(): boolean {
        return this.isDirty;
    }

    /**
     * Save the current document
     */
    async saveDocument(): Promise<void> {
        if (!this.document) {
            throw new LIVError(LIVErrorType.INVALID_FILE, 'No document to save');
        }

        try {
            // Update document content from current editor state
            if (this.editMode === 'source' && this.sourceEditor) {
                await this.updateDocumentFromSource();
            }
            
            // Trigger save event for external handling
            const saveEvent = new CustomEvent('document-save', {
                detail: { document: this.document }
            });
            this.editorContainer.dispatchEvent(saveEvent);
            
            this.markClean();
        } catch (error) {
            throw new LIVError(
                LIVErrorType.VALIDATION,
                `Failed to save document: ${error instanceof Error ? error.message : 'Unknown error'}`
            );
        }
    }
    
    /**
     * Preview the current document
     */
    async previewDocument(): Promise<void> {
        if (!this.document) {
            throw new LIVError(LIVErrorType.INVALID_FILE, 'No document to preview');
        }
        
        try {
            // Update document from current editor state
            if (this.editMode === 'source' && this.sourceEditor) {
                await this.updateDocumentFromSource();
            }
            
            // Trigger preview event for external handling
            const previewEvent = new CustomEvent('document-preview', {
                detail: { document: this.document }
            });
            this.editorContainer.dispatchEvent(previewEvent);
        } catch (error) {
            throw new LIVError(
                LIVErrorType.VALIDATION,
                `Failed to preview document: ${error instanceof Error ? error.message : 'Unknown error'}`
            );
        }
    }
    
    /**
     * Load a document from file
     */
    async loadFromFile(file: File): Promise<void> {
        try {
            this.document = await LIVDocument.fromFile(file);
            await this.loadDocument();
            this.markClean();
        } catch (error) {
            throw new LIVError(
                LIVErrorType.PARSING,
                `Failed to load document from file: ${error instanceof Error ? error.message : 'Unknown error'}`
            );
        }
    }
    
    /**
     * Export document as HTML
     */
    exportAsHTML(): string {
        if (!this.document) {
            throw new LIVError(LIVErrorType.INVALID_FILE, 'No document to export');
        }
        
        return this.document.content.html || '';
    }
    
    /**
     * Create a document from HTML content
     */
    private async createDocumentFromHTML(htmlContent: string): Promise<LIVDocument> {
        // Create a basic manifest and document structure
        const manifest = {
            version: '1.0',
            metadata: {
                title: 'New LIV Document',
                author: 'LIV Editor',
                created: new Date().toISOString(),
                modified: new Date().toISOString(),
                description: '',
                version: '1.0.0',
                language: 'en'
            },
            security: {
                wasmPermissions: {
                    memoryLimit: 64 * 1024 * 1024,
                    allowedImports: ['env'],
                    cpuTimeLimit: 5000,
                    allowNetworking: false,
                    allowFileSystem: false
                },
                jsPermissions: {
                    executionMode: 'sandboxed' as const,
                    allowedAPIs: ['dom', 'canvas'],
                    domAccess: 'write' as const
                },
                networkPolicy: {
                    allowOutbound: false,
                    allowedHosts: [],
                    allowedPorts: []
                },
                storagePolicy: {
                    allowLocalStorage: true,
                    allowSessionStorage: true,
                    allowIndexedDB: false,
                    allowCookies: false
                },
                contentSecurityPolicy: "default-src 'self'; script-src 'self' 'unsafe-inline'; style-src 'self' 'unsafe-inline';",
                trustedDomains: []
            },
            features: {
                animations: true,
                interactivity: false,
                charts: false,
                forms: false,
                audio: false,
                video: false,
                webgl: false,
                webassembly: false
            },
            resources: {}
        };

        const content = {
            html: htmlContent,
            css: '',
            interactiveSpec: '',
            staticFallback: htmlContent
        };

        const assets = {
            images: new Map(),
            fonts: new Map(),
            data: new Map()
        };

        const signatures = {
            contentSignature: '',
            manifestSignature: '',
            wasmSignatures: {}
        };

        return new LIVDocument(manifest, content, assets, signatures, new Map());
    }

    /**
     * Inject CSS styles for visual editing
     */
    private injectEditorStyles(): void {
        const styleId = 'liv-editor-styles';
        if (document.getElementById(styleId)) return;

        const style = document.createElement('style');
        style.id = styleId;
        style.textContent = `
            /* Visual editing styles */
            .editor-selected {
                outline: 2px solid #007bff !important;
                outline-offset: 2px !important;
                position: relative !important;
            }

            .dragging {
                opacity: 0.5 !important;
                transform: rotate(5deg) !important;
            }

            .drop-zone-highlight {
                background-color: rgba(0, 123, 255, 0.1) !important;
                border: 2px dashed #007bff !important;
            }

            .drop-target {
                background-color: rgba(0, 123, 255, 0.2) !important;
                border: 2px solid #007bff !important;
            }

            .resize-handle {
                position: absolute;
                background: #007bff;
                border: 1px solid #fff;
                width: 8px;
                height: 8px;
                z-index: 1000;
            }

            .resize-nw { top: -4px; left: -4px; cursor: nw-resize; }
            .resize-n { top: -4px; left: 50%; transform: translateX(-50%); cursor: n-resize; }
            .resize-ne { top: -4px; right: -4px; cursor: ne-resize; }
            .resize-e { top: 50%; right: -4px; transform: translateY(-50%); cursor: e-resize; }
            .resize-se { bottom: -4px; right: -4px; cursor: se-resize; }
            .resize-s { bottom: -4px; left: 50%; transform: translateX(-50%); cursor: s-resize; }
            .resize-sw { bottom: -4px; left: -4px; cursor: sw-resize; }
            .resize-w { top: 50%; left: -4px; transform: translateY(-50%); cursor: w-resize; }

            .visual-style-panel {
                background: white;
                border: 1px solid #ddd;
                border-radius: 8px;
                box-shadow: 0 4px 12px rgba(0,0,0,0.15);
                padding: 16px;
                min-width: 250px;
                font-family: -apple-system, BlinkMacSystemFont, 'Segoe UI', Roboto, sans-serif;
            }

            .style-panel-header {
                display: flex;
                justify-content: space-between;
                align-items: center;
                margin-bottom: 16px;
                padding-bottom: 8px;
                border-bottom: 1px solid #eee;
            }

            .style-panel-header h4 {
                margin: 0;
                color: #333;
                font-size: 16px;
            }

            .close-btn {
                background: none;
                border: none;
                font-size: 18px;
                cursor: pointer;
                color: #999;
                padding: 0;
                width: 24px;
                height: 24px;
                display: flex;
                align-items: center;
                justify-content: center;
            }

            .close-btn:hover {
                color: #333;
            }

            .style-group {
                margin-bottom: 12px;
            }

            .style-group label {
                display: block;
                margin-bottom: 4px;
                font-size: 12px;
                font-weight: 500;
                color: #555;
            }

            .style-input {
                width: 100%;
                padding: 4px 8px;
                border: 1px solid #ddd;
                border-radius: 4px;
                font-size: 14px;
            }

            .style-input[type="range"] {
                padding: 0;
            }

            .style-input[type="color"] {
                height: 32px;
                padding: 2px;
            }

            .value-display {
                font-size: 11px;
                color: #666;
                margin-left: 8px;
            }

            .editable-container {
                min-height: 100px;
                position: relative;
            }

            .editable-container:empty::after {
                content: 'Drop elements here';
                position: absolute;
                top: 50%;
                left: 50%;
                transform: translate(-50%, -50%);
                color: #999;
                font-style: italic;
                pointer-events: none;
            }

            .editable-interactive {
                position: relative;
                cursor: pointer;
            }

            .editable-interactive::after {
                content: '⚡';
                position: absolute;
                top: 4px;
                right: 4px;
                background: #007bff;
                color: white;
                border-radius: 50%;
                width: 20px;
                height: 20px;
                display: flex;
                align-items: center;
                justify-content: center;
                font-size: 12px;
            }

            .editable-chart {
                position: relative;
            }

            .editable-chart::after {
                content: '📊';
                position: absolute;
                top: 4px;
                right: 4px;
                background: #28a745;
                color: white;
                border-radius: 4px;
                padding: 2px 6px;
                font-size: 12px;
            }

            /* Hover effects for editable elements */
            [data-editable="true"]:hover {
                outline: 1px dashed #007bff;
                outline-offset: 1px;
            }

            /* Visual feedback for drag operations */
            .visual-editor {
                position: relative;
            }

            .visual-editor.drag-active {
                background-color: rgba(0, 123, 255, 0.05);
            }

            /* Source editor styles */
            .source-tabs {
                display: flex;
                background: #f8f9fa;
                border-bottom: 1px solid #dee2e6;
                align-items: center;
            }

            .source-tab {
                background: none;
                border: none;
                padding: 12px 16px;
                cursor: pointer;
                display: flex;
                align-items: center;
                gap: 6px;
                font-size: 14px;
                color: #495057;
                border-bottom: 2px solid transparent;
                transition: all 0.2s;
            }

            .source-tab:hover {
                background: #e9ecef;
            }

            .source-tab.active {
                color: #007bff;
                border-bottom-color: #007bff;
                background: white;
            }

            .tab-icon {
                font-size: 16px;
            }

            .tab-status {
                background: #dc3545;
                color: white;
                border-radius: 10px;
                padding: 2px 6px;
                font-size: 11px;
                min-width: 16px;
                text-align: center;
            }

            .tab-status.warning {
                background: #ffc107;
                color: #212529;
            }

            .tab-status:empty {
                display: none;
            }

            .tab-actions {
                margin-left: auto;
                display: flex;
                gap: 4px;
                padding-right: 8px;
            }

            .tab-action-btn {
                background: none;
                border: 1px solid #dee2e6;
                border-radius: 4px;
                padding: 6px 8px;
                cursor: pointer;
                font-size: 12px;
                color: #495057;
                transition: all 0.2s;
            }

            .tab-action-btn:hover {
                background: #e9ecef;
                border-color: #adb5bd;
            }

            .source-editors {
                flex: 1;
                position: relative;
                overflow: hidden;
            }

            .source-editor-container {
                position: absolute;
                top: 0;
                left: 0;
                right: 0;
                bottom: 0;
                display: none;
                flex-direction: column;
            }

            .source-editor-container.active {
                display: flex;
            }

            .editor-header {
                background: #f8f9fa;
                border-bottom: 1px solid #dee2e6;
                padding: 8px 16px;
                display: flex;
                justify-content: space-between;
                align-items: center;
                font-size: 12px;
                color: #6c757d;
            }

            .editor-info {
                display: flex;
                gap: 16px;
            }

            .editor-actions {
                display: flex;
                gap: 8px;
            }

            .editor-btn {
                background: none;
                border: 1px solid #dee2e6;
                border-radius: 3px;
                padding: 4px 8px;
                cursor: pointer;
                font-size: 11px;
                color: #495057;
            }

            .editor-btn:hover {
                background: #e9ecef;
            }

            .editor-wrapper {
                flex: 1;
                position: relative;
                overflow: hidden;
                display: flex;
            }

            .line-numbers {
                background: #f8f9fa;
                border-right: 1px solid #dee2e6;
                padding: 8px 4px;
                font-family: 'Monaco', 'Menlo', 'Ubuntu Mono', monospace;
                font-size: 13px;
                line-height: 1.4;
                color: #6c757d;
                text-align: right;
                min-width: 40px;
                user-select: none;
                overflow: hidden;
            }

            .line-number {
                height: 18.2px;
                padding-right: 8px;
            }

            .code-editor {
                flex: 1;
                border: none;
                outline: none;
                padding: 8px 12px;
                font-family: 'Monaco', 'Menlo', 'Ubuntu Mono', monospace;
                font-size: 13px;
                line-height: 1.4;
                resize: none;
                background: transparent;
                color: #212529;
                white-space: pre;
                overflow-wrap: normal;
                overflow-x: auto;
                tab-size: 2;
            }

            .syntax-overlay {
                position: absolute;
                top: 0;
                left: 40px;
                right: 0;
                bottom: 0;
                padding: 8px 12px;
                font-family: 'Monaco', 'Menlo', 'Ubuntu Mono', monospace;
                font-size: 13px;
                line-height: 1.4;
                pointer-events: none;
                white-space: pre;
                overflow: hidden;
                color: transparent;
            }

            .validation-panel {
                max-height: 200px;
                overflow-y: auto;
                border-top: 1px solid #dee2e6;
                background: #f8f9fa;
            }

            .validation-panel.success {
                background: #d4edda;
                border-color: #c3e6cb;
            }

            .validation-panel.warning {
                background: #fff3cd;
                border-color: #ffeaa7;
            }

            .validation-panel.error {
                background: #f8d7da;
                border-color: #f5c6cb;
            }

            .validation-success {
                padding: 8px 16px;
                color: #155724;
                font-size: 14px;
            }

            .validation-results {
                padding: 8px 0;
            }

            .validation-section {
                margin-bottom: 12px;
            }

            .validation-section h4 {
                padding: 4px 16px;
                margin: 0 0 4px 0;
                font-size: 12px;
                font-weight: 600;
                text-transform: uppercase;
            }

            .validation-section.errors h4 {
                color: #721c24;
            }

            .validation-section.warnings h4 {
                color: #856404;
            }

            .validation-section ul {
                list-style: none;
                margin: 0;
                padding: 0;
            }

            .validation-item {
                padding: 4px 16px;
                cursor: pointer;
                font-size: 13px;
                display: flex;
                align-items: center;
                gap: 8px;
            }

            .validation-item:hover {
                background: rgba(0,0,0,0.05);
            }

            .validation-item.error {
                color: #721c24;
            }

            .validation-item.warning {
                color: #856404;
            }

            .validation-item .line-number {
                font-weight: 600;
                min-width: 60px;
            }

            .validation-item .message {
                flex: 1;
            }

            .validation-item .code {
                font-family: monospace;
                background: rgba(0,0,0,0.1);
                padding: 2px 4px;
                border-radius: 2px;
                font-size: 11px;
            }

            /* Syntax highlighting colors */
            .syntax-html-0 { color: #0066cc; } /* HTML tags */
            .syntax-html-1 { color: #cc6600; } /* Attributes */
            .syntax-html-2 { color: #999999; } /* Comments */

            .syntax-css-0 { color: #0066cc; } /* Properties */
            .syntax-css-1 { color: #999999; } /* Comments */
            .syntax-css-2 { color: #cc0066; } /* Selectors */

            .syntax-javascript-0 { color: #0066cc; font-weight: bold; } /* Keywords */
            .syntax-javascript-1 { color: #999999; } /* Comments */
            .syntax-javascript-2 { color: #cc6600; } /* Strings */

            .syntax-json-0 { color: #0066cc; } /* Keys */
            .syntax-json-1 { color: #cc6600; } /* String values */
            .syntax-json-2 { color: #cc0066; } /* Literals */
            .syntax-json-3 { color: #009900; } /* Numbers */
        `;

        document.head.appendChild(style);
    }

    /**
     * Setup drag and drop functionality
     */
    private setupDragAndDrop(): void {
        const visualEditor = this.editorContainer.querySelector('#visual-editor');
        if (!visualEditor) return;

        // Make elements draggable and set up drop zones
        this.makeDraggable(visualEditor as HTMLElement);
        this.setupDropZones(visualEditor as HTMLElement);
    }

    /**
     * Make elements draggable
     */
    private makeDraggable(container: HTMLElement): void {
        // Add draggable attribute to all elements
        const elements = container.querySelectorAll('*');
        elements.forEach((element: Element) => {
            const htmlElement = element as HTMLElement;
            htmlElement.draggable = true;
            htmlElement.addEventListener('dragstart', this.handleDragStart.bind(this));
            htmlElement.addEventListener('dragend', this.handleDragEnd.bind(this));
        });
    }

    /**
     * Setup drop zones
     */
    private setupDropZones(container: HTMLElement): void {
        container.addEventListener('dragover', this.handleDragOver.bind(this));
        container.addEventListener('drop', this.handleDrop.bind(this));
        container.addEventListener('dragenter', this.handleDragEnter.bind(this));
        container.addEventListener('dragleave', this.handleDragLeave.bind(this));
    }

    /**
     * Handle drag start
     */
    private handleDragStart(event: DragEvent): void {
        this.draggedElement = event.target as HTMLElement;
        this.draggedElement.classList.add('dragging');
        
        // Store element data for WASM
        if (event.dataTransfer) {
            event.dataTransfer.setData('text/html', this.draggedElement.outerHTML);
            event.dataTransfer.setData('element-id', this.draggedElement.id || this.generateElementId());
            event.dataTransfer.effectAllowed = 'move';
        }

        // Highlight drop zones
        this.highlightDropZones();
    }

    /**
     * Handle drag end
     */
    private handleDragEnd(event: DragEvent): void {
        if (this.draggedElement) {
            this.draggedElement.classList.remove('dragging');
            this.draggedElement = null;
        }
        
        // Remove drop zone highlights
        this.removeDropZoneHighlights();
    }

    /**
     * Handle drag over
     */
    private handleDragOver(event: DragEvent): void {
        event.preventDefault();
        if (event.dataTransfer) {
            event.dataTransfer.dropEffect = 'move';
        }
    }

    /**
     * Handle drag enter
     */
    private handleDragEnter(event: DragEvent): void {
        event.preventDefault();
        const target = event.target as HTMLElement;
        if (this.isValidDropTarget(target)) {
            target.classList.add('drop-target');
        }
    }

    /**
     * Handle drag leave
     */
    private handleDragLeave(event: DragEvent): void {
        const target = event.target as HTMLElement;
        target.classList.remove('drop-target');
    }

    /**
     * Handle drop
     */
    private handleDrop(event: DragEvent): void {
        event.preventDefault();
        const target = event.target as HTMLElement;
        target.classList.remove('drop-target');

        if (!this.draggedElement || !this.isValidDropTarget(target)) {
            return;
        }

        // Perform the move operation
        this.moveElement(this.draggedElement, target);
        
        // Notify WASM backend of the change
        this.notifyWASMElementMoved(this.draggedElement, target);
        
        this.markDirty();
    }

    /**
     * Check if target is a valid drop target
     */
    private isValidDropTarget(target: HTMLElement): boolean {
        return target !== this.draggedElement && 
               !this.draggedElement?.contains(target) &&
               (target.tagName === 'DIV' || target.tagName === 'SECTION' || 
                target.tagName === 'ARTICLE' || target.classList.contains('container'));
    }

    /**
     * Move element to new parent
     */
    private moveElement(element: HTMLElement, newParent: HTMLElement): void {
        // Remove from current parent
        element.remove();
        
        // Add to new parent
        newParent.appendChild(element);
        
        // Update element selection
        this.selectElement(element);
    }

    /**
     * Notify WASM backend of element move
     */
    private notifyWASMElementMoved(element: HTMLElement, newParent: HTMLElement): void {
        // This would communicate with the WASM backend
        const moveData = {
            elementId: element.id || this.generateElementId(),
            newParentId: newParent.id || this.generateElementId(),
            operation: 'move'
        };
        
        // Send to WASM via sandbox
        this.sandbox.sendEvent('element-moved', moveData).catch(error => {
            console.warn('Failed to notify WASM of element move:', error);
        });
    }

    /**
     * Highlight drop zones
     */
    private highlightDropZones(): void {
        const visualEditor = this.editorContainer.querySelector('#visual-editor');
        if (!visualEditor) return;

        const containers = visualEditor.querySelectorAll('div, section, article, .container');
        containers.forEach((container: Element) => {
            const htmlContainer = container as HTMLElement;
            if (this.isValidDropTarget(htmlContainer)) {
                htmlContainer.classList.add('drop-zone-highlight');
            }
        });
    }

    /**
     * Remove drop zone highlights
     */
    private removeDropZoneHighlights(): void {
        const highlighted = this.editorContainer.querySelectorAll('.drop-zone-highlight, .drop-target');
        highlighted.forEach((element: Element) => {
            element.classList.remove('drop-zone-highlight', 'drop-target');
        });
    }

    /**
     * Generate unique element ID
     */
    private generateElementId(): string {
        return `element-${++this.elementIdCounter}`;
    }

    /**
     * Setup visual style panel
     */
    private setupVisualStylePanel(): void {
        this.visualStylePanel = document.createElement('div');
        this.visualStylePanel.className = 'visual-style-panel';
        this.visualStylePanel.style.display = 'none';
        this.visualStylePanel.innerHTML = `
            <div class="style-panel-header">
                <h4>Visual Styles</h4>
                <button class="close-btn" onclick="this.parentElement.parentElement.style.display='none'">×</button>
            </div>
            <div class="style-controls">
                <div class="style-group">
                    <label>Background Color:</label>
                    <input type="color" id="bg-color" class="style-input">
                </div>
                <div class="style-group">
                    <label>Text Color:</label>
                    <input type="color" id="text-color" class="style-input">
                </div>
                <div class="style-group">
                    <label>Font Size:</label>
                    <input type="range" id="font-size" min="8" max="72" class="style-input">
                    <span class="value-display">16px</span>
                </div>
                <div class="style-group">
                    <label>Padding:</label>
                    <input type="range" id="padding" min="0" max="50" class="style-input">
                    <span class="value-display">0px</span>
                </div>
                <div class="style-group">
                    <label>Margin:</label>
                    <input type="range" id="margin" min="0" max="50" class="style-input">
                    <span class="value-display">0px</span>
                </div>
                <div class="style-group">
                    <label>Border Radius:</label>
                    <input type="range" id="border-radius" min="0" max="25" class="style-input">
                    <span class="value-display">0px</span>
                </div>
                <div class="style-group">
                    <label>Opacity:</label>
                    <input type="range" id="opacity" min="0" max="100" value="100" class="style-input">
                    <span class="value-display">100%</span>
                </div>
            </div>
        `;

        document.body.appendChild(this.visualStylePanel);
        this.setupStylePanelEvents();
    }

    /**
     * Setup style panel event handlers
     */
    private setupStylePanelEvents(): void {
        if (!this.visualStylePanel) return;

        const styleInputs = this.visualStylePanel.querySelectorAll('.style-input');
        styleInputs.forEach((input: Element) => {
            const htmlInput = input as HTMLInputElement;
            htmlInput.addEventListener('input', (event) => {
                this.handleStyleChange(event.target as HTMLInputElement);
            });
        });
    }

    /**
     * Handle style changes from visual style panel
     */
    private handleStyleChange(input: HTMLInputElement): void {
        if (!this.selectedElement) return;

        const property = input.id;
        const value = input.value;
        let cssProperty = '';
        let cssValue = '';

        switch (property) {
            case 'bg-color':
                cssProperty = 'backgroundColor';
                cssValue = value;
                break;
            case 'text-color':
                cssProperty = 'color';
                cssValue = value;
                break;
            case 'font-size':
                cssProperty = 'fontSize';
                cssValue = `${value}px`;
                this.updateValueDisplay(input, `${value}px`);
                break;
            case 'padding':
                cssProperty = 'padding';
                cssValue = `${value}px`;
                this.updateValueDisplay(input, `${value}px`);
                break;
            case 'margin':
                cssProperty = 'margin';
                cssValue = `${value}px`;
                this.updateValueDisplay(input, `${value}px`);
                break;
            case 'border-radius':
                cssProperty = 'borderRadius';
                cssValue = `${value}px`;
                this.updateValueDisplay(input, `${value}px`);
                break;
            case 'opacity':
                cssProperty = 'opacity';
                cssValue = (parseInt(value) / 100).toString();
                this.updateValueDisplay(input, `${value}%`);
                break;
        }

        if (cssProperty && cssValue) {
            // Apply style to element
            (this.selectedElement.style as any)[cssProperty] = cssValue;
            
            // Notify WASM backend
            this.notifyWASMStyleChange(this.selectedElement, cssProperty, cssValue);
            
            this.markDirty();
        }
    }

    /**
     * Update value display next to range inputs
     */
    private updateValueDisplay(input: HTMLInputElement, displayValue: string): void {
        const valueDisplay = input.parentElement?.querySelector('.value-display');
        if (valueDisplay) {
            valueDisplay.textContent = displayValue;
        }
    }

    /**
     * Notify WASM backend of style changes
     */
    private notifyWASMStyleChange(element: HTMLElement, property: string, value: string): void {
        const styleData = {
            elementId: element.id || this.generateElementId(),
            property,
            value,
            operation: 'style-change'
        };
        
        this.sandbox.sendEvent('style-changed', styleData).catch(error => {
            console.warn('Failed to notify WASM of style change:', error);
        });
    }

    /**
     * Setup resize handles for elements
     */
    private setupResizeHandles(): void {
        // This will be called when an element is selected
        // to add resize handles around it
    }

    /**
     * Add resize handles to selected element
     */
    private addResizeHandles(element: HTMLElement): void {
        this.removeResizeHandles(); // Remove existing handles

        const handles = ['nw', 'n', 'ne', 'e', 'se', 's', 'sw', 'w'];
        handles.forEach(direction => {
            const handle = document.createElement('div');
            handle.className = `resize-handle resize-${direction}`;
            handle.addEventListener('mousedown', (e) => this.startResize(e, direction));
            element.appendChild(handle);
        });
    }

    /**
     * Remove resize handles
     */
    private removeResizeHandles(): void {
        const handles = document.querySelectorAll('.resize-handle');
        handles.forEach(handle => handle.remove());
    }

    /**
     * Start resize operation
     */
    private startResize(event: MouseEvent, direction: string): void {
        event.preventDefault();
        event.stopPropagation();
        
        this.isResizing = true;
        
        const startX = event.clientX;
        const startY = event.clientY;
        const element = this.selectedElement;
        
        if (!element) return;
        
        const startWidth = element.offsetWidth;
        const startHeight = element.offsetHeight;
        
        const handleMouseMove = (e: MouseEvent) => {
            if (!this.isResizing || !element) return;
            
            const deltaX = e.clientX - startX;
            const deltaY = e.clientY - startY;
            
            this.applyResize(element, direction, deltaX, deltaY, startWidth, startHeight);
        };
        
        const handleMouseUp = () => {
            this.isResizing = false;
            document.removeEventListener('mousemove', handleMouseMove);
            document.removeEventListener('mouseup', handleMouseUp);
            
            if (element) {
                this.notifyWASMElementResized(element);
            }
        };
        
        document.addEventListener('mousemove', handleMouseMove);
        document.addEventListener('mouseup', handleMouseUp);
    }

    /**
     * Apply resize based on direction and delta
     */
    private applyResize(element: HTMLElement, direction: string, deltaX: number, deltaY: number, startWidth: number, startHeight: number): void {
        let newWidth = startWidth;
        let newHeight = startHeight;
        
        if (direction.includes('e')) {
            newWidth = Math.max(50, startWidth + deltaX);
        }
        if (direction.includes('w')) {
            newWidth = Math.max(50, startWidth - deltaX);
        }
        if (direction.includes('s')) {
            newHeight = Math.max(30, startHeight + deltaY);
        }
        if (direction.includes('n')) {
            newHeight = Math.max(30, startHeight - deltaY);
        }
        
        element.style.width = `${newWidth}px`;
        element.style.height = `${newHeight}px`;
        
        this.markDirty();
    }

    /**
     * Notify WASM backend of element resize
     */
    private notifyWASMElementResized(element: HTMLElement): void {
        const resizeData = {
            elementId: element.id || this.generateElementId(),
            width: element.style.width,
            height: element.style.height,
            operation: 'resize'
        };
        
        this.sandbox.sendEvent('element-resized', resizeData).catch(error => {
            console.warn('Failed to notify WASM of element resize:', error);
        });
    }

    /**
     * Enhanced element selection with visual editing features
     */
    private selectElement(element: HTMLElement): void {
        // Remove previous selection
        if (this.selectedElement) {
            this.selectedElement.classList.remove('editor-selected');
            this.removeResizeHandles();
        }
        
        // Set new selection
        this.selectedElement = element;
        element.classList.add('editor-selected');
        
        // Add resize handles
        this.addResizeHandles(element);
        
        // Update properties panel
        this.updatePropertiesPanel();
        
        // Show visual style panel
        this.showVisualStylePanel(element);
        
        // Notify WASM backend
        const elementId = element.id || this.generateElementId();
        if (!element.id) {
            element.id = elementId;
        }
        
        this.sandbox.sendEvent('element-selected', { elementId }).catch(error => {
            console.warn('Failed to notify WASM of element selection:', error);
        });
    }

    /**
     * Show visual style panel for selected element
     */
    private showVisualStylePanel(element: HTMLElement): void {
        if (!this.visualStylePanel) return;
        
        // Position panel near the selected element
        const rect = element.getBoundingClientRect();
        this.visualStylePanel.style.position = 'fixed';
        this.visualStylePanel.style.left = `${rect.right + 10}px`;
        this.visualStylePanel.style.top = `${rect.top}px`;
        this.visualStylePanel.style.display = 'block';
        this.visualStylePanel.style.zIndex = '1000';
        
        // Populate current values
        this.populateStylePanelValues(element);
    }

    /**
     * Populate style panel with current element values
     */
    private populateStylePanelValues(element: HTMLElement): void {
        if (!this.visualStylePanel) return;
        
        const computedStyle = window.getComputedStyle(element);
        
        // Background color
        const bgColorInput = this.visualStylePanel.querySelector('#bg-color') as HTMLInputElement;
        if (bgColorInput) {
            bgColorInput.value = this.rgbToHex(computedStyle.backgroundColor) || '#ffffff';
        }
        
        // Text color
        const textColorInput = this.visualStylePanel.querySelector('#text-color') as HTMLInputElement;
        if (textColorInput) {
            textColorInput.value = this.rgbToHex(computedStyle.color) || '#000000';
        }
        
        // Font size
        const fontSizeInput = this.visualStylePanel.querySelector('#font-size') as HTMLInputElement;
        if (fontSizeInput) {
            const fontSize = parseInt(computedStyle.fontSize) || 16;
            fontSizeInput.value = fontSize.toString();
            this.updateValueDisplay(fontSizeInput, `${fontSize}px`);
        }
        
        // Padding
        const paddingInput = this.visualStylePanel.querySelector('#padding') as HTMLInputElement;
        if (paddingInput) {
            const padding = parseInt(computedStyle.padding) || 0;
            paddingInput.value = padding.toString();
            this.updateValueDisplay(paddingInput, `${padding}px`);
        }
        
        // Margin
        const marginInput = this.visualStylePanel.querySelector('#margin') as HTMLInputElement;
        if (marginInput) {
            const margin = parseInt(computedStyle.margin) || 0;
            marginInput.value = margin.toString();
            this.updateValueDisplay(marginInput, `${margin}px`);
        }
        
        // Border radius
        const borderRadiusInput = this.visualStylePanel.querySelector('#border-radius') as HTMLInputElement;
        if (borderRadiusInput) {
            const borderRadius = parseInt(computedStyle.borderRadius) || 0;
            borderRadiusInput.value = borderRadius.toString();
            this.updateValueDisplay(borderRadiusInput, `${borderRadius}px`);
        }
        
        // Opacity
        const opacityInput = this.visualStylePanel.querySelector('#opacity') as HTMLInputElement;
        if (opacityInput) {
            const opacity = Math.round((parseFloat(computedStyle.opacity) || 1) * 100);
            opacityInput.value = opacity.toString();
            this.updateValueDisplay(opacityInput, `${opacity}%`);
        }
    }

    /**
     * Convert RGB color to hex
     */
    private rgbToHex(rgb: string): string {
        if (!rgb || rgb === 'rgba(0, 0, 0, 0)') return '#ffffff';
        
        const result = rgb.match(/\d+/g);
        if (!result || result.length < 3) return '#ffffff';
        
        const r = parseInt(result[0]);
        const g = parseInt(result[1]);
        const b = parseInt(result[2]);
        
        return '#' + ((1 << 24) + (r << 16) + (g << 8) + b).toString(16).slice(1);
    }

    /**
     * Setup source editor event handlers
     */
    private setupSourceEditorEvents(): void {
        this.sourceEditors.forEach((editor, language) => {
            // Input events for real-time updates
            editor.addEventListener('input', () => {
                this.handleSourceEditorInput(editor, language);
            });

            // Cursor position tracking
            editor.addEventListener('selectionchange', () => {
                this.updateCursorInfo(editor, language);
            });

            editor.addEventListener('keyup', () => {
                this.updateCursorInfo(editor, language);
            });

            editor.addEventListener('click', () => {
                this.updateCursorInfo(editor, language);
            });

            // Keyboard shortcuts
            editor.addEventListener('keydown', (event) => {
                this.handleSourceEditorKeydown(event, editor, language);
            });

            // Scroll synchronization for syntax highlighting
            editor.addEventListener('scroll', () => {
                this.syncSyntaxOverlay(editor, language);
            });
        });

        // Tab action buttons
        const formatBtn = this.editorContainer.querySelector('#format-code');
        const validateBtn = this.editorContainer.querySelector('#validate-code');
        const syncBtn = this.editorContainer.querySelector('#sync-visual');

        formatBtn?.addEventListener('click', () => this.formatCurrentCode());
        validateBtn?.addEventListener('click', () => this.validateCurrentCode());
        syncBtn?.addEventListener('click', () => this.syncWithVisualEditor());
    }

    /**
     * Setup source tab switching
     */
    private setupSourceTabSwitching(): void {
        const tabs = this.editorContainer.querySelectorAll('.source-tab');
        tabs.forEach(tab => {
            tab.addEventListener('click', () => {
                const tabName = (tab as HTMLElement).dataset.tab;
                if (tabName) {
                    this.switchSourceTab(tabName);
                }
            });
        });
    }

    /**
     * Switch source editor tab
     */
    private switchSourceTab(tabName: string): void {
        this.currentSourceTab = tabName;

        // Update tab appearance
        const tabs = this.editorContainer.querySelectorAll('.source-tab');
        tabs.forEach(tab => {
            tab.classList.toggle('active', (tab as HTMLElement).dataset.tab === tabName);
        });

        // Update editor visibility
        const editors = this.editorContainer.querySelectorAll('.source-editor-container');
        editors.forEach(editor => {
            editor.classList.toggle('active', (editor as HTMLElement).dataset.editor === tabName);
        });

        // Update main source editor reference
        this.sourceEditor = this.sourceEditors.get(tabName) || null;
    }

    /**
     * Handle source editor input
     */
    private handleSourceEditorInput(editor: HTMLTextAreaElement, language: string): void {
        // Update line numbers
        this.updateLineNumbers(editor, language);
        
        // Update syntax highlighting
        this.updateSyntaxHighlighting(editor, language);
        
        // Update character count
        this.updateCharacterCount(editor, language);
        
        // Mark as dirty
        this.markDirty();
        
        // Debounced validation
        this.debounceValidation(language);
        
        // Notify WASM backend of changes
        this.notifyWASMSourceChange(language, editor.value);
    }

    /**
     * Handle keyboard shortcuts in source editor
     */
    private handleSourceEditorKeydown(event: KeyboardEvent, editor: HTMLTextAreaElement, language: string): void {
        if (event.ctrlKey || event.metaKey) {
            switch (event.key) {
                case 's':
                    event.preventDefault();
                    this.saveDocument();
                    break;
                case 'f':
                    event.preventDefault();
                    this.showFindDialog(language);
                    break;
                case 'g':
                    event.preventDefault();
                    this.showGotoLineDialog(language);
                    break;
                case 'd':
                    event.preventDefault();
                    this.formatCurrentCode();
                    break;
                case 'Enter':
                    event.preventDefault();
                    this.validateCurrentCode();
                    break;
            }
        }

        // Auto-indentation
        if (event.key === 'Enter') {
            this.handleAutoIndentation(event, editor);
        }

        // Auto-closing brackets/quotes
        if (['(', '[', '{', '"', "'"].includes(event.key)) {
            this.handleAutoClosing(event, editor, event.key);
        }
    }

    /**
     * Update line numbers
     */
    private updateLineNumbers(editor: HTMLTextAreaElement, language: string): void {
        const lineNumbersElement = this.editorContainer.querySelector(`#${language}-lines`);
        if (!lineNumbersElement) return;

        const lines = editor.value.split('\n');
        const lineNumbers = lines.map((_, index) => `<div class="line-number">${index + 1}</div>`).join('');
        lineNumbersElement.innerHTML = lineNumbers;
    }

    /**
     * Update syntax highlighting
     */
    private updateSyntaxHighlighting(editor: HTMLTextAreaElement, language: string): void {
        if (!this.syntaxHighlighter) return;

        const overlayElement = this.editorContainer.querySelector(`#${language}-overlay`);
        if (!overlayElement) return;

        const highlighted = this.syntaxHighlighter.highlight(editor.value, language);
        overlayElement.innerHTML = highlighted;
        
        // Sync scroll position
        this.syncSyntaxOverlay(editor, language);
    }

    /**
     * Sync syntax overlay with editor scroll
     */
    private syncSyntaxOverlay(editor: HTMLTextAreaElement, language: string): void {
        const overlayElement = this.editorContainer.querySelector(`#${language}-overlay`) as HTMLElement;
        if (!overlayElement) return;

        overlayElement.scrollTop = editor.scrollTop;
        overlayElement.scrollLeft = editor.scrollLeft;
    }

    /**
     * Update cursor position info
     */
    private updateCursorInfo(editor: HTMLTextAreaElement, language: string): void {
        const container = editor.closest('.source-editor-container');
        const lineInfo = container?.querySelector('.line-info');
        if (!lineInfo) return;

        const cursorPos = editor.selectionStart;
        const textBeforeCursor = editor.value.substring(0, cursorPos);
        const lines = textBeforeCursor.split('\n');
        const line = lines.length;
        const column = lines[lines.length - 1].length + 1;

        lineInfo.textContent = `Line: ${line}, Column: ${column}`;
    }

    /**
     * Update character count
     */
    private updateCharacterCount(editor: HTMLTextAreaElement, language: string): void {
        const container = editor.closest('.source-editor-container');
        const charCount = container?.querySelector('.char-count');
        if (!charCount) return;

        const count = editor.value.length;
        const lines = editor.value.split('\n').length;
        charCount.textContent = `${count} characters, ${lines} lines`;
    }

    /**
     * Debounced validation
     */
    private debounceValidation = this.debounce((language: string) => {
        this.validateCode(language);
    }, 1000);

    /**
     * Validate code in specific language
     */
    private async validateCode(language: string): Promise<void> {
        if (!this.codeValidator) return;

        const editor = this.sourceEditors.get(language);
        if (!editor) return;

        let results: ValidationResult[] = [];

        try {
            switch (language) {
                case 'html':
                    results = await this.codeValidator.validateHTML(editor.value);
                    break;
                case 'css':
                    results = await this.codeValidator.validateCSS(editor.value);
                    break;
                case 'javascript':
                    results = await this.codeValidator.validateJavaScript(editor.value);
                    break;
                case 'manifest':
                    results = await this.codeValidator.validateManifest(editor.value);
                    break;
            }
        } catch (error) {
            results = [{
                line: 1,
                column: 1,
                message: `Validation error: ${error instanceof Error ? error.message : 'Unknown error'}`,
                severity: 'error'
            }];
        }

        this.validationResults.set(language, results);
        this.displayValidationResults(language, results);
        this.updateTabStatus(language, results);
    }

    /**
     * Display validation results
     */
    private displayValidationResults(language: string, results: ValidationResult[]): void {
        const validationPanel = this.editorContainer.querySelector(`#${language}-validation`);
        if (!validationPanel) return;

        if (results.length === 0) {
            validationPanel.innerHTML = '<div class="validation-success">✓ No issues found</div>';
            validationPanel.className = 'validation-panel success';
            return;
        }

        const errors = results.filter(r => r.severity === 'error');
        const warnings = results.filter(r => r.severity === 'warning');

        let html = '<div class="validation-results">';
        
        if (errors.length > 0) {
            html += `<div class="validation-section errors">
                <h4>Errors (${errors.length})</h4>
                <ul>`;
            errors.forEach(error => {
                html += `<li class="validation-item error" data-line="${error.line}">
                    <span class="line-number">Line ${error.line}:</span>
                    <span class="message">${error.message}</span>
                    ${error.code ? `<span class="code">[${error.code}]</span>` : ''}
                </li>`;
            });
            html += '</ul></div>';
        }

        if (warnings.length > 0) {
            html += `<div class="validation-section warnings">
                <h4>Warnings (${warnings.length})</h4>
                <ul>`;
            warnings.forEach(warning => {
                html += `<li class="validation-item warning" data-line="${warning.line}">
                    <span class="line-number">Line ${warning.line}:</span>
                    <span class="message">${warning.message}</span>
                    ${warning.code ? `<span class="code">[${warning.code}]</span>` : ''}
                </li>`;
            });
            html += '</ul></div>';
        }

        html += '</div>';
        validationPanel.innerHTML = html;
        validationPanel.className = `validation-panel ${errors.length > 0 ? 'error' : 'warning'}`;

        // Add click handlers to jump to lines
        validationPanel.querySelectorAll('.validation-item').forEach(item => {
            item.addEventListener('click', () => {
                const line = parseInt((item as HTMLElement).dataset.line || '1');
                this.goToLine(language, line);
            });
        });
    }

    /**
     * Update tab status indicator
     */
    private updateTabStatus(language: string, results: ValidationResult[]): void {
        const statusElement = this.editorContainer.querySelector(`#${language}-status`);
        if (!statusElement) return;

        const errors = results.filter(r => r.severity === 'error').length;
        const warnings = results.filter(r => r.severity === 'warning').length;

        if (errors > 0) {
            statusElement.textContent = `${errors}`;
            statusElement.className = 'tab-status error';
        } else if (warnings > 0) {
            statusElement.textContent = `${warnings}`;
            statusElement.className = 'tab-status warning';
        } else {
            statusElement.textContent = '';
            statusElement.className = 'tab-status';
        }
    }

    /**
     * Format current code
     */
    private formatCurrentCode(): void {
        const editor = this.sourceEditors.get(this.currentSourceTab);
        if (!editor) return;

        let formatted = '';
        
        try {
            switch (this.currentSourceTab) {
                case 'html':
                    formatted = this.formatHTML(editor.value);
                    break;
                case 'css':
                    formatted = this.formatCSS(editor.value);
                    break;
                case 'javascript':
                    formatted = this.formatJavaScript(editor.value);
                    break;
                case 'manifest':
                    formatted = this.formatJSON(editor.value);
                    break;
                default:
                    return;
            }

            editor.value = formatted;
            this.handleSourceEditorInput(editor, this.currentSourceTab);
        } catch (error) {
            console.warn('Failed to format code:', error);
        }
    }

    /**
     * Format HTML code
     */
    private formatHTML(html: string): string {
        // Basic HTML formatting
        return html
            .replace(/></g, '>\n<')
            .replace(/^\s+|\s+$/gm, '')
            .split('\n')
            .map(line => line.trim())
            .filter(line => line.length > 0)
            .join('\n');
    }

    /**
     * Format CSS code
     */
    private formatCSS(css: string): string {
        return css
            .replace(/\s*{\s*/g, ' {\n  ')
            .replace(/;\s*/g, ';\n  ')
            .replace(/\s*}\s*/g, '\n}\n')
            .replace(/,\s*/g, ',\n')
            .trim();
    }

    /**
     * Format JavaScript code
     */
    private formatJavaScript(js: string): string {
        // Basic JS formatting
        return js
            .replace(/\s*{\s*/g, ' {\n  ')
            .replace(/;\s*/g, ';\n  ')
            .replace(/\s*}\s*/g, '\n}\n')
            .trim();
    }

    /**
     * Format JSON code
     */
    private formatJSON(json: string): string {
        try {
            const parsed = JSON.parse(json);
            return JSON.stringify(parsed, null, 2);
        } catch (error) {
            return json; // Return original if parsing fails
        }
    }

    /**
     * Validate current code
     */
    private async validateCurrentCode(): Promise<void> {
        await this.validateCode(this.currentSourceTab);
    }

    /**
     * Sync with visual editor
     */
    private async syncWithVisualEditor(): Promise<void> {
        if (!this.document) return;

        try {
            // Update document from all source editors
            const htmlEditor = this.sourceEditors.get('html');
            const cssEditor = this.sourceEditors.get('css');
            const jsEditor = this.sourceEditors.get('javascript');
            const manifestEditor = this.sourceEditors.get('manifest');

            if (htmlEditor) {
                this.document.content.html = htmlEditor.value;
            }
            if (cssEditor) {
                this.document.content.css = cssEditor.value;
            }
            if (jsEditor) {
                this.document.content.interactiveSpec = jsEditor.value;
            }
            if (manifestEditor) {
                try {
                    this.document.manifest = JSON.parse(manifestEditor.value);
                } catch (error) {
                    console.warn('Invalid manifest JSON, keeping existing manifest');
                }
            }

            // Re-render in visual editor
            if (this.editMode === 'visual') {
                await this.renderer.renderDocument(this.document);
            }

            // Notify WASM backend
            this.sandbox.sendEvent('document-synced', {
                operation: 'sync',
                timestamp: Date.now()
            }).catch(error => {
                console.warn('Failed to notify WASM of sync:', error);
            });

        } catch (error) {
            console.error('Failed to sync with visual editor:', error);
        }
    }

    /**
     * Show find dialog
     */
    private showFindDialog(language: string): void {
        // Implementation for find/replace dialog
        const findText = prompt('Find:');
        if (findText) {
            const editor = this.sourceEditors.get(language);
            if (editor) {
                const index = editor.value.indexOf(findText);
                if (index !== -1) {
                    editor.focus();
                    editor.setSelectionRange(index, index + findText.length);
                }
            }
        }
    }

    /**
     * Show goto line dialog
     */
    private showGotoLineDialog(language: string): void {
        const lineStr = prompt('Go to line:');
        if (lineStr) {
            const line = parseInt(lineStr);
            if (!isNaN(line)) {
                this.goToLine(language, line);
            }
        }
    }

    /**
     * Go to specific line
     */
    private goToLine(language: string, line: number): void {
        const editor = this.sourceEditors.get(language);
        if (!editor) return;

        const lines = editor.value.split('\n');
        if (line > 0 && line <= lines.length) {
            const position = lines.slice(0, line - 1).join('\n').length + (line > 1 ? 1 : 0);
            editor.focus();
            editor.setSelectionRange(position, position);
            editor.scrollTop = (line - 1) * 20; // Approximate line height
        }
    }

    /**
     * Handle auto-indentation
     */
    private handleAutoIndentation(_event: KeyboardEvent, editor: HTMLTextAreaElement): void {
        const cursorPos = editor.selectionStart;
        const textBeforeCursor = editor.value.substring(0, cursorPos);
        const lines = textBeforeCursor.split('\n');
        const currentLine = lines[lines.length - 1];
        
        // Get indentation of current line
        const indentMatch = currentLine.match(/^(\s*)/);
        const indent = indentMatch ? indentMatch[1] : '';
        
        // Add extra indentation for opening braces/tags
        let extraIndent = '';
        if (currentLine.includes('{') || currentLine.includes('<') && !currentLine.includes('</')) {
            extraIndent = '  ';
        }
        
        setTimeout(() => {
            const newCursorPos = editor.selectionStart;
            const newValue = editor.value.substring(0, newCursorPos) + indent + extraIndent + editor.value.substring(newCursorPos);
            editor.value = newValue;
            editor.setSelectionRange(newCursorPos + indent.length + extraIndent.length, newCursorPos + indent.length + extraIndent.length);
        }, 0);
    }

    /**
     * Handle auto-closing brackets/quotes
     */
    private handleAutoClosing(_event: KeyboardEvent, editor: HTMLTextAreaElement, char: string): void {
        const closingChars: Record<string, string> = {
            '(': ')',
            '[': ']',
            '{': '}',
            '"': '"',
            "'": "'"
        };

        const closingChar = closingChars[char];
        if (!closingChar) return;

        setTimeout(() => {
            const cursorPos = editor.selectionStart;
            const newValue = editor.value.substring(0, cursorPos) + closingChar + editor.value.substring(cursorPos);
            editor.value = newValue;
            editor.setSelectionRange(cursorPos, cursorPos);
        }, 0);
    }

    /**
     * Notify WASM backend of source changes
     */
    private notifyWASMSourceChange(language: string, content: string): void {
        const changeData = {
            language,
            content,
            operation: 'source-change',
            timestamp: Date.now()
        };
        
        this.sandbox.sendEvent('source-changed', changeData).catch(error => {
            console.warn('Failed to notify WASM of source change:', error);
        });
    }

    /**
     * Load document content into source editors
     */
    private loadDocumentIntoSourceEditors(): void {
        if (!this.document) return;

        const htmlEditor = this.sourceEditors.get('html');
        const cssEditor = this.sourceEditors.get('css');
        const jsEditor = this.sourceEditors.get('javascript');
        const manifestEditor = this.sourceEditors.get('manifest');

        if (htmlEditor) {
            htmlEditor.value = this.document.content.html || '';
            this.handleSourceEditorInput(htmlEditor, 'html');
        }
        if (cssEditor) {
            cssEditor.value = this.document.content.css || '';
            this.handleSourceEditorInput(cssEditor, 'css');
        }
        if (jsEditor) {
            jsEditor.value = this.document.content.interactiveSpec || '';
            this.handleSourceEditorInput(jsEditor, 'javascript');
        }
        if (manifestEditor) {
            manifestEditor.value = JSON.stringify(this.document.manifest, null, 2);
            this.handleSourceEditorInput(manifestEditor, 'manifest');
        }
    }

    /**
     * Cleanup editor resources
     */
    destroy(): void {
        // Remove event listeners and cleanup
        this.selectedElement = null;
        this.document = null;
        this.isInitialized = false;
        this.removeResizeHandles();
        
        if (this.visualStylePanel) {
            this.visualStylePanel.remove();
            this.visualStylePanel = null;
        }
        
        // Clear source editor references
        this.sourceEditors.clear();
        this.validationResults.clear();
        this.syntaxHighlighter = null;
        this.codeValidator = null;
    }
}