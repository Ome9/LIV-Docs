import { LIVEditor } from '../src/editor';
import { LIVDocument } from '../src/document';

describe('Editor Workflow Tests', () => {
    let editorContainer: HTMLElement;
    let previewContainer: HTMLElement;
    let toolbarContainer: HTMLElement;
    let propertiesContainer: HTMLElement;
    let editor: LIVEditor;

    beforeEach(() => {
        // Create DOM elements for testing
        document.body.innerHTML = `
            <div id="editor-container"></div>
            <div id="preview-container"></div>
            <div id="toolbar-container"></div>
            <div id="properties-container"></div>
        `;

        editorContainer = document.getElementById('editor-container')!;
        previewContainer = document.getElementById('preview-container')!;
        toolbarContainer = document.getElementById('toolbar-container')!;
        propertiesContainer = document.getElementById('properties-container')!;

        editor = new LIVEditor(
            editorContainer,
            previewContainer,
            toolbarContainer,
            propertiesContainer
        );
    });

    afterEach(() => {
        if (editor) {
            editor.destroy();
        }
        document.body.innerHTML = '';
    });

    describe('Complete Document Creation Workflow', () => {
        it('should create a complete document from scratch', async () => {
            await editor.initialize();

            // Step 1: Create document structure
            editor.insertElement('heading', { text: 'My LIV Document' });
            editor.insertElement('paragraph', { text: 'Introduction paragraph with some content.' });
            
            // Step 2: Add container for layout
            editor.insertElement('container');
            const container = editorContainer.querySelector('.container') as HTMLElement;
            
            // Step 3: Add content to container
            (editor as any).selectElement(container);
            editor.insertElement('heading', { text: 'Section Heading' });
            editor.insertElement('paragraph', { text: 'Section content goes here.' });
            editor.insertElement('image', { 
                src: 'https://example.com/image.jpg', 
                alt: 'Example image' 
            });

            // Step 4: Add interactive elements
            editor.insertElement('interactive');
            editor.insertElement('chart');

            // Step 5: Verify document structure
            const document = editor.getDocument();
            expect(document).toBeTruthy();
            expect(document?.content.html).toContain('My LIV Document');
            expect(document?.content.html).toContain('interactive-element');
            expect(document?.content.html).toContain('chart-container');

            // Step 6: Test document metadata
            expect(document?.manifest.metadata.title).toBe('New LIV Document');
            expect(document?.manifest.version).toBe('1.0');
            expect(document?.manifest.security).toBeTruthy();
        });

        it('should handle complete editing workflow with styling', async () => {
            await editor.initialize();

            // Create elements
            editor.insertElement('heading', { text: 'Styled Heading' });
            editor.insertElement('paragraph', { text: 'Styled paragraph' });

            const heading = editorContainer.querySelector('h2') as HTMLElement;
            const paragraph = editorContainer.querySelector('p') as HTMLElement;

            // Apply styles through visual style panel
            (editor as any).selectElement(heading);
            
            // Simulate style panel interactions
            const stylePanel = document.querySelector('.visual-style-panel') as HTMLElement;
            expect(stylePanel).toBeTruthy();

            // Test background color change
            const bgColorInput = stylePanel.querySelector('#bg-color') as HTMLInputElement;
            bgColorInput.value = '#ff0000';
            bgColorInput.dispatchEvent(new Event('input'));
            expect(heading.style.backgroundColor).toBe('rgb(255, 0, 0)');

            // Test font size change
            const fontSizeInput = stylePanel.querySelector('#font-size') as HTMLInputElement;
            fontSizeInput.value = '24';
            fontSizeInput.dispatchEvent(new Event('input'));
            expect(heading.style.fontSize).toBe('24px');

            // Switch to paragraph and apply different styles
            (editor as any).selectElement(paragraph);
            
            const textColorInput = stylePanel.querySelector('#text-color') as HTMLInputElement;
            textColorInput.value = '#0000ff';
            textColorInput.dispatchEvent(new Event('input'));
            expect(paragraph.style.color).toBe('rgb(0, 0, 255)');

            // Verify document is marked as dirty
            expect(editor.isDirtyDocument()).toBeTruthy();
        });

        it('should handle drag and drop workflow', async () => {
            await editor.initialize();

            // Create source and target elements
            editor.insertElement('paragraph', { text: 'Draggable paragraph' });
            editor.insertElement('container');

            const paragraph = editorContainer.querySelector('p') as HTMLElement;
            const container = editorContainer.querySelector('.container') as HTMLElement;

            // Simulate drag and drop
            const dragStartEvent = new DragEvent('dragstart', {
                bubbles: true,
                dataTransfer: new DataTransfer()
            });
            
            paragraph.dispatchEvent(dragStartEvent);
            expect(paragraph.classList.contains('dragging')).toBeTruthy();

            // Simulate drop
            (editor as any).moveElement(paragraph, container);
            expect(paragraph.parentElement).toBe(container);

            // Verify WASM notification
            const sendEventSpy = jest.spyOn((editor as any).sandbox, 'sendEvent').mockResolvedValue(undefined);
            (editor as any).notifyWASMElementMoved(paragraph, container);
            
            expect(sendEventSpy).toHaveBeenCalledWith('element-moved', expect.objectContaining({
                elementId: paragraph.id,
                newParentId: container.id,
                operation: 'move'
            }));
        });
    });

    describe('Source Code Editing Workflow', () => {
        it('should handle complete source editing workflow', async () => {
            await editor.initialize();
            (editor as any).switchMode('source');

            // Step 1: Edit HTML
            const htmlEditor = editorContainer.querySelector('#html-editor') as HTMLTextAreaElement;
            htmlEditor.value = `
                <div class="document">
                    <h1>Source Edited Document</h1>
                    <p>This content was added via source editing.</p>
                    <div class="interactive-section">
                        <h2>Interactive Content</h2>
                        <div class="interactive-element">Interactive placeholder</div>
                    </div>
                </div>
            `;
            htmlEditor.dispatchEvent(new Event('input'));

            // Step 2: Edit CSS
            const cssTab = editorContainer.querySelector('[data-tab="css"]') as HTMLElement;
            cssTab.click();
            
            const cssEditor = editorContainer.querySelector('#css-editor') as HTMLTextAreaElement;
            cssEditor.value = `
                .document {
                    max-width: 800px;
                    margin: 0 auto;
                    padding: 20px;
                }
                
                .interactive-section {
                    background: #f0f0f0;
                    padding: 15px;
                    border-radius: 8px;
                }
                
                .interactive-element {
                    border: 2px dashed #007bff;
                    padding: 20px;
                    text-align: center;
                }
            `;
            cssEditor.dispatchEvent(new Event('input'));

            // Step 3: Edit JavaScript
            const jsTab = editorContainer.querySelector('[data-tab="javascript"]') as HTMLElement;
            jsTab.click();
            
            const jsEditor = editorContainer.querySelector('#js-editor') as HTMLTextAreaElement;
            jsEditor.value = `
                // Interactive functionality
                function initializeInteractive() {
                    const elements = document.querySelectorAll('.interactive-element');
                    elements.forEach(element => {
                        element.addEventListener('click', function() {
                            this.style.backgroundColor = '#e3f2fd';
                        });
                    });
                }
                
                // Initialize when DOM is ready
                document.addEventListener('DOMContentLoaded', initializeInteractive);
            `;
            jsEditor.dispatchEvent(new Event('input'));

            // Step 4: Update manifest
            const manifestTab = editorContainer.querySelector('[data-tab="manifest"]') as HTMLElement;
            manifestTab.click();
            
            const manifestEditor = editorContainer.querySelector('#manifest-editor') as HTMLTextAreaElement;
            const manifest = {
                version: "1.0",
                metadata: {
                    title: "Source Edited LIV Document",
                    author: "Test Author",
                    description: "A document created through source editing",
                    created: new Date().toISOString(),
                    modified: new Date().toISOString(),
                    version: "1.0.0",
                    language: "en"
                },
                security: {
                    wasmPermissions: {
                        memoryLimit: 67108864,
                        allowedImports: ["env"],
                        cpuTimeLimit: 5000,
                        allowNetworking: false,
                        allowFileSystem: false
                    },
                    jsPermissions: {
                        executionMode: "sandboxed",
                        allowedAPIs: ["dom", "canvas"],
                        domAccess: "write"
                    }
                },
                features: {
                    animations: true,
                    interactivity: true,
                    charts: false,
                    forms: false
                }
            };
            manifestEditor.value = JSON.stringify(manifest, null, 2);
            manifestEditor.dispatchEvent(new Event('input'));

            // Step 5: Validate all content
            await (editor as any).validateCode('html');
            await (editor as any).validateCode('css');
            await (editor as any).validateCode('javascript');
            await (editor as any).validateCode('manifest');

            // Step 6: Sync with visual editor
            await (editor as any).syncWithVisualEditor();

            // Verify document was updated
            const document = editor.getDocument();
            expect(document?.content.html).toContain('Source Edited Document');
            expect(document?.content.css).toContain('.document');
            expect(document?.content.interactiveSpec).toContain('initializeInteractive');
            expect(document?.manifest.metadata.title).toBe('Source Edited LIV Document');
        });

        it('should handle validation and error correction workflow', async () => {
            await editor.initialize();
            (editor as any).switchMode('source');

            // Step 1: Introduce HTML errors
            const htmlEditor = editorContainer.querySelector('#html-editor') as HTMLTextAreaElement;
            htmlEditor.value = '<div><h1>Unclosed heading<p>Missing closing tags';
            
            await (editor as any).validateCode('html');
            
            let validationPanel = editorContainer.querySelector('#html-validation');
            expect(validationPanel?.classList.contains('error')).toBeTruthy();
            
            // Step 2: Fix HTML errors
            htmlEditor.value = '<div><h1>Fixed Heading</h1><p>Properly closed tags</p></div>';
            await (editor as any).validateCode('html');
            
            validationPanel = editorContainer.querySelector('#html-validation');
            expect(validationPanel?.classList.contains('success')).toBeTruthy();

            // Step 3: Test CSS validation
            const cssTab = editorContainer.querySelector('[data-tab="css"]') as HTMLElement;
            cssTab.click();
            
            const cssEditor = editorContainer.querySelector('#css-editor') as HTMLTextAreaElement;
            cssEditor.value = 'body { behavior: url(malicious.htc); }'; // Security violation
            
            await (editor as any).validateCode('css');
            
            validationPanel = editorContainer.querySelector('#css-validation');
            expect(validationPanel?.classList.contains('error')).toBeTruthy();
            expect(validationPanel?.textContent).toContain('security');

            // Step 4: Fix CSS
            cssEditor.value = 'body { color: blue; font-family: Arial, sans-serif; }';
            await (editor as any).validateCode('css');
            
            validationPanel = editorContainer.querySelector('#css-validation');
            expect(validationPanel?.classList.contains('success')).toBeTruthy();
        });

        it('should handle code formatting workflow', async () => {
            await editor.initialize();
            (editor as any).switchMode('source');

            // Test HTML formatting
            const htmlEditor = editorContainer.querySelector('#html-editor') as HTMLTextAreaElement;
            htmlEditor.value = '<div><h1>Unformatted</h1><p>Content</p></div>';
            
            (editor as any).formatCurrentCode();
            expect(htmlEditor.value).toContain('\n');

            // Test CSS formatting
            const cssTab = editorContainer.querySelector('[data-tab="css"]') as HTMLElement;
            cssTab.click();
            
            const cssEditor = editorContainer.querySelector('#css-editor') as HTMLTextAreaElement;
            cssEditor.value = 'body{color:red;margin:0;}';
            
            (editor as any).formatCurrentCode();
            expect(cssEditor.value).toContain('{\n');
            expect(cssEditor.value).toContain(';\n');

            // Test JSON formatting
            const manifestTab = editorContainer.querySelector('[data-tab="manifest"]') as HTMLElement;
            manifestTab.click();
            
            const manifestEditor = editorContainer.querySelector('#manifest-editor') as HTMLTextAreaElement;
            manifestEditor.value = '{"version":"1.0","metadata":{"title":"test"}}';
            
            (editor as any).formatCurrentCode();
            expect(manifestEditor.value).toContain('  ');
            expect(manifestEditor.value).toContain('\n');
        });
    });

    describe('Document Persistence Workflow', () => {
        it('should handle save and load workflow', async () => {
            await editor.initialize();

            // Create document content
            editor.insertElement('heading', { text: 'Persistent Document' });
            editor.insertElement('paragraph', { text: 'This document will be saved and loaded.' });
            editor.insertElement('interactive');

            // Test save operation
            let savedDocument: any = null;
            editorContainer.addEventListener('document-save', (event: any) => {
                savedDocument = event.detail.document;
            });

            await editor.saveDocument();
            
            expect(savedDocument).toBeTruthy();
            expect(savedDocument.content.html).toContain('Persistent Document');
            expect(editor.isDirtyDocument()).toBeFalsy();

            // Test export
            const exportedHTML = editor.exportAsHTML();
            expect(exportedHTML).toContain('<!DOCTYPE html>');
            expect(exportedHTML).toContain('Persistent Document');

            // Test load from file
            const htmlContent = '<h1>Loaded Document</h1><p>From file</p>';
            const mockDocument = await (editor as any).createDocumentFromHTML(htmlContent);
            jest.spyOn(LIVDocument, 'fromFile').mockResolvedValue(mockDocument);

            const file = new File([htmlContent], 'test.html', { type: 'text/html' });
            await editor.loadFromFile(file);

            const loadedDocument = editor.getDocument();
            expect(loadedDocument?.content.html).toContain('Loaded Document');
        });

        it('should handle preview workflow', async () => {
            await editor.initialize();

            // Create content for preview
            editor.insertElement('heading', { text: 'Preview Document' });
            editor.insertElement('paragraph', { text: 'This will be previewed.' });

            // Test preview event
            let previewEventFired = false;
            let previewedDocument: any = null;

            editorContainer.addEventListener('document-preview', (event: any) => {
                previewEventFired = true;
                previewedDocument = event.detail.document;
            });

            await editor.previewDocument();

            expect(previewEventFired).toBeTruthy();
            expect(previewedDocument).toBeTruthy();
            expect(previewedDocument.content.html).toContain('Preview Document');
        });
    });

    describe('Mode Switching Workflow', () => {
        it('should handle seamless mode switching', async () => {
            await editor.initialize();

            // Start in visual mode
            expect((editor as any).editMode).toBe('visual');

            // Create content in visual mode
            editor.insertElement('heading', { text: 'Mode Switch Test' });
            editor.insertElement('paragraph', { text: 'Visual mode content' });

            // Switch to source mode
            (editor as any).switchMode('source');
            expect((editor as any).editMode).toBe('source');

            // Verify content is loaded in source editors
            const htmlEditor = editorContainer.querySelector('#html-editor') as HTMLTextAreaElement;
            expect(htmlEditor.value).toContain('Mode Switch Test');

            // Modify content in source mode
            htmlEditor.value = '<h1>Modified in Source</h1><p>Source mode changes</p>';
            htmlEditor.dispatchEvent(new Event('input'));

            // Switch back to visual mode
            (editor as any).switchMode('visual');
            expect((editor as any).editMode).toBe('visual');

            // Verify changes are reflected
            await (editor as any).syncWithVisualEditor();
            const document = editor.getDocument();
            expect(document?.content.html).toContain('Modified in Source');
        });

        it('should maintain state during mode switches', async () => {
            await editor.initialize();

            // Create and select element in visual mode
            editor.insertElement('paragraph', { text: 'State test' });
            const paragraph = editorContainer.querySelector('p') as HTMLElement;
            (editor as any).selectElement(paragraph);

            expect((editor as any).selectedElement).toBe(paragraph);
            expect(editor.isDirtyDocument()).toBeTruthy();

            // Switch to source mode
            (editor as any).switchMode('source');

            // Dirty state should be maintained
            expect(editor.isDirtyDocument()).toBeTruthy();

            // Switch back to visual mode
            (editor as any).switchMode('visual');

            // State should be preserved
            expect(editor.isDirtyDocument()).toBeTruthy();
        });
    });

    describe('Error Recovery Workflow', () => {
        it('should recover from validation errors', async () => {
            await editor.initialize();
            (editor as any).switchMode('source');

            // Introduce validation error
            const htmlEditor = editorContainer.querySelector('#html-editor') as HTMLTextAreaElement;
            htmlEditor.value = '<div><p>Broken HTML';
            
            await (editor as any).validateCode('html');
            
            let validationPanel = editorContainer.querySelector('#html-validation');
            expect(validationPanel?.classList.contains('error')).toBeTruthy();

            // Fix the error
            htmlEditor.value = '<div><p>Fixed HTML</p></div>';
            await (editor as any).validateCode('html');
            
            validationPanel = editorContainer.querySelector('#html-validation');
            expect(validationPanel?.classList.contains('success')).toBeTruthy();

            // Should be able to sync successfully
            await (editor as any).syncWithVisualEditor();
            
            const document = editor.getDocument();
            expect(document?.content.html).toContain('Fixed HTML');
        });

        it('should handle WASM communication failures gracefully', async () => {
            await editor.initialize();

            // Mock WASM communication failure
            const sendEventSpy = jest.spyOn((editor as any).sandbox, 'sendEvent').mockRejectedValue(new Error('WASM failed'));

            // Operations should continue despite WASM failures
            editor.insertElement('heading', { text: 'WASM Error Test' });
            
            const heading = editorContainer.querySelector('h2') as HTMLElement;
            expect(heading).toBeTruthy();
            expect(heading.textContent).toBe('WASM Error Test');

            // Editor should remain functional
            editor.insertElement('paragraph', { text: 'Still working' });
            
            const paragraph = editorContainer.querySelector('p') as HTMLElement;
            expect(paragraph).toBeTruthy();
            expect(paragraph.textContent).toBe('Still working');

            expect(sendEventSpy).toHaveBeenCalled();
        });
    });
});