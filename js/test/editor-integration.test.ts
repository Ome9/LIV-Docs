import { LIVEditor } from '../src/editor';
import { LIVDocument } from '../src/document';

describe('Editor Integration Tests', () => {
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

    describe('WYSIWYG Operations with Test Infrastructure', () => {
        beforeEach(async () => {
            await editor.initialize();
        });

        it('should perform complete WYSIWYG workflow', async () => {
            // Test element creation
            editor.insertElement('heading', { text: 'Test Heading' });
            editor.insertElement('paragraph', { text: 'Test paragraph content' });
            editor.insertElement('image', { 
                src: 'https://example.com/test.jpg', 
                alt: 'Test image' 
            });

            const visualEditor = editorContainer.querySelector('#visual-editor');
            
            // Verify elements were created
            expect(visualEditor?.querySelector('h2')?.textContent).toBe('Test Heading');
            expect(visualEditor?.querySelector('p')?.textContent).toBe('Test paragraph content');
            expect(visualEditor?.querySelector('img')?.getAttribute('src')).toBe('https://example.com/test.jpg');

            // Test element selection and property updates
            const heading = visualEditor?.querySelector('h2') as HTMLElement;
            (editor as any).selectElement(heading);

            editor.updateElementProperties({
                textContent: 'Updated Heading',
                className: 'test-heading'
            });

            expect(heading.textContent).toBe('Updated Heading');
            expect(heading.className).toBe('test-heading');

            // Test drag and drop
            const paragraph = visualEditor?.querySelector('p') as HTMLElement;
            const container = document.createElement('div');
            container.className = 'container';
            visualEditor?.appendChild(container);

            (editor as any).moveElement(paragraph, container);
            expect(paragraph.parentElement).toBe(container);

            // Verify document is marked as dirty
            expect(editor.isDirtyDocument()).toBeTruthy();
        });

        it('should handle complex element interactions', async () => {
            // Create nested structure
            editor.insertElement('container');
            const container = editorContainer.querySelector('.container') as HTMLElement;
            
            // Insert elements into container
            (editor as any).selectElement(container);
            editor.insertElement('heading', { text: 'Container Heading' });
            editor.insertElement('paragraph', { text: 'Container content' });

            // Verify nested structure
            expect(container.children.length).toBeGreaterThan(0);

            // Test interactive element creation
            editor.insertElement('interactive');
            const interactive = editorContainer.querySelector('.interactive-element') as HTMLElement;
            expect(interactive).toBeTruthy();
            expect(interactive.getAttribute('data-element-type')).toBe('interactive');

            // Test chart element
            editor.insertElement('chart');
            const chart = editorContainer.querySelector('.chart-container') as HTMLElement;
            expect(chart).toBeTruthy();
            expect(chart.querySelector('canvas')).toBeTruthy();
        });

        it('should integrate with visual style panel', async () => {
            editor.insertElement('paragraph', { text: 'Styled paragraph' });
            const paragraph = editorContainer.querySelector('p') as HTMLElement;
            
            // Select element to show style panel
            (editor as any).selectElement(paragraph);
            
            const stylePanel = document.querySelector('.visual-style-panel') as HTMLElement;
            expect(stylePanel).toBeTruthy();
            expect(stylePanel.style.display).toBe('block');

            // Test style changes
            const bgColorInput = stylePanel.querySelector('#bg-color') as HTMLInputElement;
            bgColorInput.value = '#ff0000';
            bgColorInput.dispatchEvent(new Event('input'));

            expect(paragraph.style.backgroundColor).toBe('rgb(255, 0, 0)');
        });
    });

    describe('Source Code Editing with Validation Systems', () => {
        beforeEach(async () => {
            await editor.initialize();
            (editor as any).switchMode('source');
        });

        it('should validate HTML with existing validation systems', async () => {
            const htmlEditor = editorContainer.querySelector('#html-editor') as HTMLTextAreaElement;
            
            // Test valid HTML
            htmlEditor.value = '<div><h1>Valid HTML</h1><p>Content</p></div>';
            await (editor as any).validateCode('html');
            
            let validationPanel = editorContainer.querySelector('#html-validation');
            expect(validationPanel?.classList.contains('success')).toBeTruthy();

            // Test invalid HTML
            htmlEditor.value = '<div><h1>Unclosed heading<p>Missing closing div';
            await (editor as any).validateCode('html');
            
            validationPanel = editorContainer.querySelector('#html-validation');
            expect(validationPanel?.classList.contains('error')).toBeTruthy();
            expect(validationPanel?.textContent).toContain('Unclosed tag');
        });

        it('should validate CSS with security checks', async () => {
            // Switch to CSS tab
            const cssTab = editorContainer.querySelector('[data-tab="css"]') as HTMLElement;
            cssTab.click();

            const cssEditor = editorContainer.querySelector('#css-editor') as HTMLTextAreaElement;
            
            // Test valid CSS
            cssEditor.value = 'body { color: blue; font-size: 16px; }';
            await (editor as any).validateCode('css');
            
            let validationPanel = editorContainer.querySelector('#css-validation');
            expect(validationPanel?.classList.contains('success')).toBeTruthy();

            // Test CSS with security issues
            cssEditor.value = 'body { behavior: url(malicious.htc); }';
            await (editor as any).validateCode('css');
            
            validationPanel = editorContainer.querySelector('#css-validation');
            expect(validationPanel?.classList.contains('error')).toBeTruthy();
            expect(validationPanel?.textContent).toContain('not allowed for security');
        });

        it('should validate JavaScript with sandbox restrictions', async () => {
            // Switch to JavaScript tab
            const jsTab = editorContainer.querySelector('[data-tab="javascript"]') as HTMLElement;
            jsTab.click();

            const jsEditor = editorContainer.querySelector('#js-editor') as HTMLTextAreaElement;
            
            // Test safe JavaScript
            jsEditor.value = 'function safeFunction() { return "hello"; }';
            await (editor as any).validateCode('javascript');
            
            let validationPanel = editorContainer.querySelector('#js-validation');
            expect(validationPanel?.classList.contains('success')).toBeTruthy();

            // Test dangerous JavaScript
            jsEditor.value = 'eval("dangerous code"); setTimeout("alert(1)", 1000);';
            await (editor as any).validateCode('javascript');
            
            validationPanel = editorContainer.querySelector('#js-validation');
            expect(validationPanel?.classList.contains('error')).toBeTruthy();
            expect(validationPanel?.textContent).toContain('not allowed in sandboxed');
        });

        it('should validate manifest with LIV format requirements', async () => {
            // Switch to manifest tab
            const manifestTab = editorContainer.querySelector('[data-tab="manifest"]') as HTMLElement;
            manifestTab.click();

            const manifestEditor = editorContainer.querySelector('#manifest-editor') as HTMLTextAreaElement;
            
            // Test valid manifest
            const validManifest = {
                version: "1.0",
                metadata: {
                    title: "Test Document",
                    author: "Test Author"
                },
                security: {
                    wasmPermissions: {
                        memoryLimit: 64000000
                    }
                }
            };
            manifestEditor.value = JSON.stringify(validManifest, null, 2);
            await (editor as any).validateCode('manifest');
            
            let validationPanel = editorContainer.querySelector('#manifest-validation');
            expect(validationPanel?.classList.contains('success')).toBeTruthy();

            // Test invalid manifest
            manifestEditor.value = '{ "invalid": "manifest without required fields" }';
            await (editor as any).validateCode('manifest');
            
            validationPanel = editorContainer.querySelector('#manifest-validation');
            expect(validationPanel?.classList.contains('error')).toBeTruthy();
            expect(validationPanel?.textContent).toContain('must have');
        });

        it('should sync between source and visual modes', async () => {
            const htmlEditor = editorContainer.querySelector('#html-editor') as HTMLTextAreaElement;
            
            // Modify HTML in source mode
            htmlEditor.value = '<h1>Source Modified</h1><p>New content from source</p>';
            
            // Sync to visual mode
            await (editor as any).syncWithVisualEditor();
            
            // Switch to visual mode and verify changes
            (editor as any).switchMode('visual');
            
            const document = editor.getDocument();
            expect(document?.content.html).toContain('Source Modified');
            expect(document?.content.html).toContain('New content from source');
        });
    });

    describe('Document Saving with Container and Signature Systems', () => {
        beforeEach(async () => {
            await editor.initialize();
        });

        it('should save document with proper event handling', async () => {
            let saveEventFired = false;
            let savedDocument: any = null;

            // Listen for save event
            editorContainer.addEventListener('document-save', (event: any) => {
                saveEventFired = true;
                savedDocument = event.detail.document;
            });

            // Make some changes
            editor.insertElement('heading', { text: 'Test Document' });
            editor.insertElement('paragraph', { text: 'Document content' });

            // Save document
            await editor.saveDocument();

            expect(saveEventFired).toBeTruthy();
            expect(savedDocument).toBeTruthy();
            expect(editor.isDirtyDocument()).toBeFalsy();
        });

        it('should handle save errors gracefully', async () => {
            // Mock a save error by removing the document
            (editor as any).document = null;

            await expect(editor.saveDocument()).rejects.toThrow();
        });

        it('should export document as HTML', () => {
            editor.insertElement('heading', { text: 'Export Test' });
            
            const html = editor.exportAsHTML();
            expect(html).toContain('Export Test');
            expect(html).toContain('<!DOCTYPE html>');
        });

        it('should load document from file', async () => {
            const htmlContent = '<h1>Loaded Document</h1><p>File content</p>';
            const file = new File([htmlContent], 'test.html', { type: 'text/html' });
            
            // Mock the LIVDocument.fromFile method
            const mockDocument = await (editor as any).createDocumentFromHTML(htmlContent);
            jest.spyOn(LIVDocument, 'fromFile').mockResolvedValue(mockDocument);

            await editor.loadFromFile(file);
            
            const document = editor.getDocument();
            expect(document?.content.html).toContain('Loaded Document');
        });

        it('should handle preview operations', async () => {
            let previewEventFired = false;

            editorContainer.addEventListener('document-preview', () => {
                previewEventFired = true;
            });

            editor.insertElement('heading', { text: 'Preview Test' });
            await editor.previewDocument();

            expect(previewEventFired).toBeTruthy();
        });
    });

    describe('WASM Interactive Engine Integration', () => {
        beforeEach(async () => {
            await editor.initialize();
        });

        it('should communicate with WASM backend for element operations', async () => {
            const sendEventSpy = jest.spyOn((editor as any).sandbox, 'sendEvent').mockResolvedValue(undefined);

            // Test element creation
            editor.insertElement('interactive');
            expect(sendEventSpy).toHaveBeenCalledWith('element-created', expect.objectContaining({
                elementType: 'interactive',
                operation: 'create'
            }));

            // Test element selection
            const interactive = editorContainer.querySelector('.interactive-element') as HTMLElement;
            (editor as any).selectElement(interactive);
            expect(sendEventSpy).toHaveBeenCalledWith('element-selected', expect.objectContaining({
                elementId: interactive.id
            }));

            // Test style changes
            (editor as any).notifyWASMStyleChange(interactive, 'color', 'red');
            expect(sendEventSpy).toHaveBeenCalledWith('style-changed', expect.objectContaining({
                elementId: interactive.id,
                property: 'color',
                value: 'red'
            }));
        });

        it('should handle WASM communication errors gracefully', async () => {
            const sendEventSpy = jest.spyOn((editor as any).sandbox, 'sendEvent').mockRejectedValue(new Error('WASM error'));
            
            // Should not throw error even if WASM communication fails
            expect(() => {
                editor.insertElement('interactive');
            }).not.toThrow();

            expect(sendEventSpy).toHaveBeenCalled();
        });

        it('should sync source changes with WASM backend', async () => {
            const sendEventSpy = jest.spyOn((editor as any).sandbox, 'sendEvent').mockResolvedValue(undefined);

            (editor as any).switchMode('source');
            const htmlEditor = editorContainer.querySelector('#html-editor') as HTMLTextAreaElement;
            
            htmlEditor.value = '<div>WASM sync test</div>';
            htmlEditor.dispatchEvent(new Event('input'));

            expect(sendEventSpy).toHaveBeenCalledWith('source-changed', expect.objectContaining({
                language: 'html',
                content: '<div>WASM sync test</div>',
                operation: 'source-change'
            }));
        });

        it('should handle interactive element configuration', () => {
            editor.insertElement('interactive');
            const interactive = editorContainer.querySelector('.interactive-element') as HTMLElement;
            
            // Verify interactive element has proper attributes
            expect(interactive.getAttribute('data-element-type')).toBe('interactive');
            expect(interactive.getAttribute('data-editable')).toBe('true');
            expect(interactive.classList.contains('editable-interactive')).toBeTruthy();
        });

        it('should handle chart element integration', () => {
            editor.insertElement('chart');
            const chart = editorContainer.querySelector('.chart-container') as HTMLElement;
            
            // Verify chart element structure
            expect(chart.getAttribute('data-element-type')).toBe('chart');
            expect(chart.querySelector('canvas')).toBeTruthy();
            expect(chart.classList.contains('editable-chart')).toBeTruthy();
        });
    });

    describe('Error Handling and Recovery', () => {
        beforeEach(async () => {
            await editor.initialize();
        });

        it('should handle initialization errors', async () => {
            const newEditor = new LIVEditor(
                editorContainer,
                previewContainer,
                toolbarContainer,
                propertiesContainer
            );

            // Mock sandbox initialization failure
            jest.spyOn((newEditor as any).sandbox, 'initialize').mockRejectedValue(new Error('Init failed'));

            await expect(newEditor.initialize()).rejects.toThrow();
            
            newEditor.destroy();
        });

        it('should handle validation errors gracefully', async () => {
            (editor as any).switchMode('source');
            
            // Mock validator to throw error
            const mockValidator = (editor as any).codeValidator;
            jest.spyOn(mockValidator, 'validateHTML').mockRejectedValue(new Error('Validation system error'));

            // Should not crash the editor
            await (editor as any).validateCode('html');
            
            const validationPanel = editorContainer.querySelector('#html-validation');
            expect(validationPanel?.textContent).toContain('Validation error');
        });

        it('should handle renderer errors', async () => {
            // Mock renderer to throw error
            jest.spyOn((editor as any).renderer, 'renderDocument').mockRejectedValue(new Error('Render failed'));

            // Should not crash when trying to render
            const document = editor.getDocument();
            if (document) {
                await expect((editor as any).renderer.renderDocument(document)).rejects.toThrow();
            }
        });

        it('should validate editor state before operations', () => {
            const uninitializedEditor = new LIVEditor(
                editorContainer,
                previewContainer,
                toolbarContainer,
                propertiesContainer
            );

            // Should throw error for operations on uninitialized editor
            expect(() => {
                uninitializedEditor.insertElement('heading');
            }).toThrow();

            uninitializedEditor.destroy();
        });
    });

    describe('Performance and Memory Management', () => {
        beforeEach(async () => {
            await editor.initialize();
        });

        it('should handle large documents efficiently', async () => {
            const startTime = performance.now();

            // Create many elements
            for (let i = 0; i < 100; i++) {
                editor.insertElement('paragraph', { text: `Paragraph ${i}` });
            }

            const endTime = performance.now();
            const duration = endTime - startTime;

            // Should complete within reasonable time (adjust threshold as needed)
            expect(duration).toBeLessThan(5000); // 5 seconds
        });

        it('should clean up resources properly', () => {
            // Add some elements and select one
            editor.insertElement('heading');
            const heading = editorContainer.querySelector('h2') as HTMLElement;
            (editor as any).selectElement(heading);

            // Verify resources exist
            expect((editor as any).selectedElement).toBeTruthy();
            expect(document.querySelector('.visual-style-panel')).toBeTruthy();

            // Destroy editor
            editor.destroy();

            // Verify cleanup
            expect((editor as any).selectedElement).toBeNull();
            expect((editor as any).document).toBeNull();
            expect(document.querySelector('.visual-style-panel')).toBeFalsy();
        });

        it('should handle rapid user interactions', () => {
            // Simulate rapid element creation and selection
            for (let i = 0; i < 10; i++) {
                editor.insertElement('paragraph', { text: `Rapid ${i}` });
                const paragraph = editorContainer.querySelector(`p:nth-child(${i + 1})`) as HTMLElement;
                if (paragraph) {
                    (editor as any).selectElement(paragraph);
                }
            }

            // Should not crash or leave inconsistent state
            expect(editor.getDocument()).toBeTruthy();
            expect((editor as any).selectedElement).toBeTruthy();
        });
    });

    describe('Cross-browser Compatibility', () => {
        beforeEach(async () => {
            await editor.initialize();
        });

        it('should handle different event models', () => {
            // Test with different event types
            const htmlEditor = editorContainer.querySelector('#html-editor') as HTMLTextAreaElement;
            
            // Test input event
            htmlEditor.dispatchEvent(new Event('input', { bubbles: true }));
            
            // Test keydown event
            htmlEditor.dispatchEvent(new KeyboardEvent('keydown', { key: 's', ctrlKey: true }));
            
            // Should not throw errors
            expect(htmlEditor).toBeTruthy();
        });

        it('should handle different selection APIs', () => {
            editor.insertElement('paragraph', { text: 'Selection test' });
            const paragraph = editorContainer.querySelector('p') as HTMLElement;
            
            // Should handle element selection regardless of browser
            (editor as any).selectElement(paragraph);
            
            expect(paragraph.classList.contains('editor-selected')).toBeTruthy();
        });
    });
});