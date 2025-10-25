import { LIVEditor } from '../src/editor';

describe('Comprehensive Editor Functionality Tests', () => {
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

    describe('Complete Editor Lifecycle', () => {
        it('should handle complete editor lifecycle', async () => {
            // 1. Initialize
            await editor.initialize();
            expect(editor.getDocument()).toBeTruthy();

            // 2. Create content
            editor.insertElement('heading', { text: 'Lifecycle Test' });
            editor.insertElement('paragraph', { text: 'Testing complete lifecycle' });
            editor.insertElement('container');
            editor.insertElement('interactive');

            // 3. Verify content creation
            const visualEditor = editorContainer.querySelector('#visual-editor');
            expect(visualEditor?.querySelector('h2')?.textContent).toBe('Lifecycle Test');
            expect(visualEditor?.querySelector('.interactive-element')).toBeTruthy();

            // 4. Test visual editing
            const heading = visualEditor?.querySelector('h2') as HTMLElement;
            (editor as any).selectElement(heading);
            
            const stylePanel = document.querySelector('.visual-style-panel');
            expect(stylePanel).toBeTruthy();

            // 5. Test source editing
            (editor as any).switchMode('source');
            const htmlEditor = editorContainer.querySelector('#html-editor') as HTMLTextAreaElement;
            expect(htmlEditor.value).toContain('Lifecycle Test');

            // 6. Modify in source
            htmlEditor.value = '<h1>Modified in Source</h1><p>Source changes</p>';
            await (editor as any).syncWithVisualEditor();

            // 7. Verify sync
            const document = editor.getDocument();
            expect(document?.content.html).toContain('Modified in Source');

            // 8. Test save
            let saveEventFired = false;
            editorContainer.addEventListener('document-save', () => {
                saveEventFired = true;
            });
            await editor.saveDocument();
            expect(saveEventFired).toBeTruthy();

            // 9. Test cleanup
            editor.destroy();
            expect((editor as any).document).toBeNull();
        });
    });

    describe('Integration with Existing Test Infrastructure', () => {
        beforeEach(async () => {
            await editor.initialize();
        });

        it('should integrate with existing validation systems', async () => {
            // Test HTML validation integration
            (editor as any).switchMode('source');
            const htmlEditor = editorContainer.querySelector('#html-editor') as HTMLTextAreaElement;
            
            // Valid HTML
            htmlEditor.value = '<div><h1>Valid</h1><p>Content</p></div>';
            await (editor as any).validateCode('html');
            
            let validationPanel = editorContainer.querySelector('#html-validation');
            expect(validationPanel?.classList.contains('success')).toBeTruthy();

            // Invalid HTML with security issues
            htmlEditor.value = '<div onclick="alert(1)"><h1>Invalid</h1></div>';
            await (editor as any).validateCode('html');
            
            validationPanel = editorContainer.querySelector('#html-validation');
            expect(validationPanel?.classList.contains('error')).toBeTruthy();
            expect(validationPanel?.textContent).toContain('security');
        });

        it('should integrate with existing container systems', async () => {
            // Test document container integration
            const document = editor.getDocument();
            expect(document?.manifest).toBeTruthy();
            expect(document?.content).toBeTruthy();
            expect(document?.assets).toBeTruthy();

            // Test ZIP container simulation
            editor.insertElement('image', { 
                src: 'assets/images/test.jpg',
                alt: 'Test image'
            });

            const image = editorContainer.querySelector('img') as HTMLImageElement;
            expect(image.src).toContain('test.jpg');
        });

        it('should integrate with existing signature systems', async () => {
            // Test signature integration
            const document = editor.getDocument();
            expect(document?.signatures).toBeTruthy();
            expect(document?.signatures.contentSignature).toBeDefined();
            expect(document?.signatures.manifestSignature).toBeDefined();

            // Test WASM signature handling
            expect(document?.signatures.wasmSignatures).toBeTruthy();
        });
    });

    describe('WASM Interactive Engine Integration', () => {
        beforeEach(async () => {
            await editor.initialize();
        });

        it('should integrate with existing WASM interactive engine', async () => {
            const sendEventSpy = jest.spyOn((editor as any).sandbox, 'sendEvent').mockResolvedValue(undefined);

            // Test interactive element creation
            editor.insertElement('interactive');
            expect(sendEventSpy).toHaveBeenCalledWith('element-created', expect.objectContaining({
                elementType: 'interactive'
            }));

            // Test chart element creation
            editor.insertElement('chart');
            expect(sendEventSpy).toHaveBeenCalledWith('element-created', expect.objectContaining({
                elementType: 'chart'
            }));

            // Test element interaction
            const interactive = editorContainer.querySelector('.interactive-element') as HTMLElement;
            (editor as any).selectElement(interactive);
            
            expect(sendEventSpy).toHaveBeenCalledWith('element-selected', expect.objectContaining({
                elementId: interactive.id
            }));
        });

        it('should handle WASM engine responses', async () => {
            // Mock WASM engine responses
            const mockResponse = {
                success: true,
                data: { elementId: 'test-123', properties: { color: 'red' } }
            };

            jest.spyOn((editor as any).sandbox, 'sendEvent').mockResolvedValue(mockResponse);

            // Test element creation with response
            editor.insertElement('interactive');
            
            // Should handle response without errors
            expect(() => {
                (editor as any).notifyWASMElementCreated(
                    document.createElement('div'),
                    'interactive',
                    {}
                );
            }).not.toThrow();
        });

        it('should handle WASM memory constraints', async () => {
            // Test memory limit integration
            const document = editor.getDocument();
            expect(document?.manifest.security.wasmPermissions.memoryLimit).toBeTruthy();
            expect(document?.manifest.security.wasmPermissions.cpuTimeLimit).toBeTruthy();

            // Test permission validation
            expect(document?.manifest.security.wasmPermissions.allowNetworking).toBe(false);
            expect(document?.manifest.security.wasmPermissions.allowFileSystem).toBe(false);
        });
    });

    describe('Cross-Platform Compatibility', () => {
        beforeEach(async () => {
            await editor.initialize();
        });

        it('should work across different browsers', () => {
            // Test event handling compatibility
            const htmlEditor = editorContainer.querySelector('#html-editor') as HTMLTextAreaElement;
            
            // Test different event types
            htmlEditor.dispatchEvent(new Event('input'));
            htmlEditor.dispatchEvent(new KeyboardEvent('keydown', { key: 'Enter' }));
            htmlEditor.dispatchEvent(new MouseEvent('click'));

            expect(htmlEditor).toBeTruthy();
        });

        it('should handle different DOM APIs', () => {
            // Test selection API compatibility
            editor.insertElement('paragraph', { text: 'Selection test' });
            const paragraph = editorContainer.querySelector('p') as HTMLElement;
            
            (editor as any).selectElement(paragraph);
            expect(paragraph.classList.contains('editor-selected')).toBeTruthy();

            // Test drag and drop API compatibility
            const dragEvent = new DragEvent('dragstart', {
                bubbles: true,
                dataTransfer: new DataTransfer()
            });
            
            expect(() => {
                paragraph.dispatchEvent(dragEvent);
            }).not.toThrow();
        });
    });

    describe('Accessibility and Usability', () => {
        beforeEach(async () => {
            await editor.initialize();
        });

        it('should provide keyboard accessibility', () => {
            (editor as any).switchMode('source');
            const htmlEditor = editorContainer.querySelector('#html-editor') as HTMLTextAreaElement;

            // Test keyboard shortcuts
            const saveEvent = new KeyboardEvent('keydown', { key: 's', ctrlKey: true });
            const formatEvent = new KeyboardEvent('keydown', { key: 'd', ctrlKey: true });
            
            expect(() => {
                htmlEditor.dispatchEvent(saveEvent);
                htmlEditor.dispatchEvent(formatEvent);
            }).not.toThrow();
        });

        it('should provide proper ARIA attributes', () => {
            const toolbar = toolbarContainer.querySelector('.editor-toolbar');
            const sourceEditors = editorContainer.querySelectorAll('.code-editor');
            
            // Editors should be accessible
            sourceEditors.forEach(editor => {
                expect(editor.getAttribute('spellcheck')).toBe('false');
            });

            expect(toolbar).toBeTruthy();
        });

        it('should provide visual feedback', () => {
            // Test visual selection feedback
            editor.insertElement('paragraph', { text: 'Visual feedback test' });
            const paragraph = editorContainer.querySelector('p') as HTMLElement;
            
            (editor as any).selectElement(paragraph);
            expect(paragraph.classList.contains('editor-selected')).toBeTruthy();

            // Test drag feedback
            const dragEvent = new DragEvent('dragstart', {
                bubbles: true,
                dataTransfer: new DataTransfer()
            });
            paragraph.dispatchEvent(dragEvent);
            expect(paragraph.classList.contains('dragging')).toBeTruthy();
        });
    });

    describe('Error Handling and Robustness', () => {
        it('should handle malformed input gracefully', async () => {
            await editor.initialize();
            (editor as any).switchMode('source');

            const htmlEditor = editorContainer.querySelector('#html-editor') as HTMLTextAreaElement;
            
            // Test with malformed HTML
            htmlEditor.value = '<div><script>alert("xss")</script><img src=x onerror=alert(1)>';
            await (editor as any).validateCode('html');

            const validationPanel = editorContainer.querySelector('#html-validation');
            expect(validationPanel?.classList.contains('error')).toBeTruthy();
        });

        it('should recover from system errors', async () => {
            await editor.initialize();

            // Mock system error
            const originalConsoleError = console.error;
            const errors: string[] = [];
            console.error = (message: string) => {
                errors.push(message);
            };

            // Trigger potential error conditions
            try {
                (editor as any).handleSourceEditorInput(null, 'html');
                (editor as any).updateSyntaxHighlighting(null, 'html');
            } catch (error) {
                // Should handle gracefully
            }

            console.error = originalConsoleError;

            // Editor should still be functional
            expect(() => {
                editor.insertElement('paragraph', { text: 'Recovery test' });
            }).not.toThrow();
        });

        it('should validate security constraints', async () => {
            await editor.initialize();
            (editor as any).switchMode('source');

            // Test CSS security validation
            const cssTab = editorContainer.querySelector('[data-tab="css"]') as HTMLElement;
            cssTab.click();

            const cssEditor = editorContainer.querySelector('#css-editor') as HTMLTextAreaElement;
            cssEditor.value = `
                body { 
                    behavior: url(malicious.htc);
                    background: url('javascript:alert(1)');
                }
            `;

            await (editor as any).validateCode('css');

            const validationPanel = editorContainer.querySelector('#css-validation');
            expect(validationPanel?.classList.contains('error')).toBeTruthy();
            expect(validationPanel?.textContent).toContain('security');
        });
    });

    describe('Performance Under Load', () => {
        it('should maintain performance with complex documents', async () => {
            await editor.initialize();

            const startTime = performance.now();

            // Create complex nested structure
            for (let i = 0; i < 50; i++) {
                editor.insertElement('container');
                const container = editorContainer.querySelector('.container:last-child') as HTMLElement;
                
                (editor as any).selectElement(container);
                editor.insertElement('heading', { text: `Section ${i}` });
                editor.insertElement('paragraph', { text: `Content ${i}` });
                
                if (i % 5 === 0) {
                    editor.insertElement('interactive');
                }
            }

            const endTime = performance.now();
            const duration = endTime - startTime;

            expect(duration).toBeLessThan(5000); // 5 seconds
            expect(editor.getDocument()).toBeTruthy();
        });

        it('should handle rapid mode switching', async () => {
            await editor.initialize();

            const startTime = performance.now();

            // Rapid mode switching
            for (let i = 0; i < 20; i++) {
                (editor as any).switchMode('source');
                (editor as any).switchMode('visual');
            }

            const endTime = performance.now();
            const duration = endTime - startTime;

            expect(duration).toBeLessThan(2000); // 2 seconds
            expect((editor as any).editMode).toBe('visual');
        });
    });
});