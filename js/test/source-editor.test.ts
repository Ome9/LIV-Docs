import { LIVEditor } from '../src/editor';

describe('Source Code Editor Integration', () => {
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

    describe('Multi-tab Source Editor', () => {
        beforeEach(async () => {
            await editor.initialize();
            // Switch to source mode
            (editor as any).switchMode('source');
        });

        it('should create multiple source editor tabs', () => {
            const tabs = editorContainer.querySelectorAll('.source-tab');
            expect(tabs.length).toBe(4); // HTML, CSS, JavaScript, Manifest

            const tabNames = Array.from(tabs).map(tab => (tab as HTMLElement).dataset.tab);
            expect(tabNames).toEqual(['html', 'css', 'javascript', 'manifest']);
        });

        it('should create corresponding editor containers', () => {
            const containers = editorContainer.querySelectorAll('.source-editor-container');
            expect(containers.length).toBe(4);

            const containerNames = Array.from(containers).map(container => (container as HTMLElement).dataset.editor);
            expect(containerNames).toEqual(['html', 'css', 'javascript', 'manifest']);
        });

        it('should have HTML tab active by default', () => {
            const activeTab = editorContainer.querySelector('.source-tab.active');
            expect((activeTab as HTMLElement)?.dataset.tab).toBe('html');

            const activeContainer = editorContainer.querySelector('.source-editor-container.active');
            expect((activeContainer as HTMLElement)?.dataset.editor).toBe('html');
        });

        it('should switch tabs correctly', () => {
            const cssTab = editorContainer.querySelector('[data-tab="css"]') as HTMLElement;
            cssTab.click();

            expect(cssTab.classList.contains('active')).toBeTruthy();
            
            const cssContainer = editorContainer.querySelector('[data-editor="css"]');
            expect(cssContainer?.classList.contains('active')).toBeTruthy();

            const htmlContainer = editorContainer.querySelector('[data-editor="html"]');
            expect(htmlContainer?.classList.contains('active')).toBeFalsy();
        });
    });

    describe('Syntax Highlighting', () => {
        beforeEach(async () => {
            await editor.initialize();
            (editor as any).switchMode('source');
        });

        it('should initialize syntax highlighter', () => {
            const syntaxHighlighter = (editor as any).syntaxHighlighter;
            expect(syntaxHighlighter).toBeTruthy();
        });

        it('should highlight HTML syntax', () => {
            const htmlEditor = editorContainer.querySelector('#html-editor') as HTMLTextAreaElement;
            htmlEditor.value = '<div class="test">Hello</div>';
            htmlEditor.dispatchEvent(new Event('input'));

            const overlay = editorContainer.querySelector('#html-overlay');
            expect(overlay?.innerHTML).toContain('span');
        });

        it('should update syntax highlighting on input', () => {
            const cssEditor = editorContainer.querySelector('#css-editor') as HTMLTextAreaElement;
            
            // Switch to CSS tab first
            const cssTab = editorContainer.querySelector('[data-tab="css"]') as HTMLElement;
            cssTab.click();

            cssEditor.value = 'body { color: red; }';
            cssEditor.dispatchEvent(new Event('input'));

            const overlay = editorContainer.querySelector('#css-overlay');
            expect(overlay?.innerHTML).toContain('span');
        });
    });

    describe('Code Validation', () => {
        beforeEach(async () => {
            await editor.initialize();
            (editor as any).switchMode('source');
        });

        it('should initialize code validator', () => {
            const codeValidator = (editor as any).codeValidator;
            expect(codeValidator).toBeTruthy();
        });

        it('should validate HTML and show errors', async () => {
            const htmlEditor = editorContainer.querySelector('#html-editor') as HTMLTextAreaElement;
            htmlEditor.value = '<div><p>Unclosed div';
            
            // Trigger validation
            await (editor as any).validateCode('html');
            
            const validationPanel = editorContainer.querySelector('#html-validation');
            expect(validationPanel?.textContent).toContain('Unclosed tag');
        });

        it('should validate CSS and show errors', async () => {
            const cssEditor = editorContainer.querySelector('#css-editor') as HTMLTextAreaElement;
            cssEditor.value = 'body { color: red; /* unclosed comment';
            
            await (editor as any).validateCode('css');
            
            const validationPanel = editorContainer.querySelector('#css-validation');
            expect(validationPanel?.classList.contains('error')).toBeTruthy();
        });

        it('should validate JavaScript and show security warnings', async () => {
            const jsEditor = editorContainer.querySelector('#js-editor') as HTMLTextAreaElement;
            jsEditor.value = 'eval("dangerous code");';
            
            await (editor as any).validateCode('javascript');
            
            const validationPanel = editorContainer.querySelector('#js-validation');
            expect(validationPanel?.textContent).toContain('eval');
        });

        it('should validate JSON manifest', async () => {
            const manifestEditor = editorContainer.querySelector('#manifest-editor') as HTMLTextAreaElement;
            manifestEditor.value = '{ invalid json }';
            
            await (editor as any).validateCode('manifest');
            
            const validationPanel = editorContainer.querySelector('#manifest-validation');
            expect(validationPanel?.classList.contains('error')).toBeTruthy();
        });

        it('should update tab status indicators', async () => {
            const htmlEditor = editorContainer.querySelector('#html-editor') as HTMLTextAreaElement;
            htmlEditor.value = '<div><p>Unclosed div';
            
            await (editor as any).validateCode('html');
            
            const statusIndicator = editorContainer.querySelector('#html-status');
            expect(statusIndicator?.classList.contains('error')).toBeTruthy();
            expect(statusIndicator?.textContent).toBeTruthy();
        });
    });

    describe('Line Numbers and Editor Features', () => {
        beforeEach(async () => {
            await editor.initialize();
            (editor as any).switchMode('source');
        });

        it('should display line numbers', () => {
            const htmlEditor = editorContainer.querySelector('#html-editor') as HTMLTextAreaElement;
            htmlEditor.value = 'line 1\nline 2\nline 3';
            htmlEditor.dispatchEvent(new Event('input'));

            const lineNumbers = editorContainer.querySelector('#html-lines');
            expect(lineNumbers?.children.length).toBe(3);
            expect(lineNumbers?.textContent).toContain('1');
            expect(lineNumbers?.textContent).toContain('2');
            expect(lineNumbers?.textContent).toContain('3');
        });

        it('should update cursor position info', () => {
            const htmlEditor = editorContainer.querySelector('#html-editor') as HTMLTextAreaElement;
            htmlEditor.value = 'test content';
            htmlEditor.setSelectionRange(5, 5);
            htmlEditor.dispatchEvent(new Event('keyup'));

            const lineInfo = editorContainer.querySelector('.line-info');
            expect(lineInfo?.textContent).toContain('Line: 1, Column: 6');
        });

        it('should update character count', () => {
            const htmlEditor = editorContainer.querySelector('#html-editor') as HTMLTextAreaElement;
            htmlEditor.value = 'test content with multiple lines\nline 2';
            htmlEditor.dispatchEvent(new Event('input'));

            const charCount = editorContainer.querySelector('.char-count');
            expect(charCount?.textContent).toContain('characters');
            expect(charCount?.textContent).toContain('lines');
        });
    });

    describe('Code Formatting', () => {
        beforeEach(async () => {
            await editor.initialize();
            (editor as any).switchMode('source');
        });

        it('should format HTML code', () => {
            const htmlEditor = editorContainer.querySelector('#html-editor') as HTMLTextAreaElement;
            htmlEditor.value = '<div><p>Test</p></div>';
            
            (editor as any).formatCurrentCode();
            
            expect(htmlEditor.value).toContain('\n');
        });

        it('should format CSS code', () => {
            // Switch to CSS tab
            const cssTab = editorContainer.querySelector('[data-tab="css"]') as HTMLElement;
            cssTab.click();

            const cssEditor = editorContainer.querySelector('#css-editor') as HTMLTextAreaElement;
            cssEditor.value = 'body{color:red;margin:0;}';
            
            (editor as any).formatCurrentCode();
            
            expect(cssEditor.value).toContain('{\n');
            expect(cssEditor.value).toContain(';\n');
        });

        it('should format JSON manifest', () => {
            // Switch to manifest tab
            const manifestTab = editorContainer.querySelector('[data-tab="manifest"]') as HTMLElement;
            manifestTab.click();

            const manifestEditor = editorContainer.querySelector('#manifest-editor') as HTMLTextAreaElement;
            manifestEditor.value = '{"version":"1.0","metadata":{"title":"test"}}';
            
            (editor as any).formatCurrentCode();
            
            expect(manifestEditor.value).toContain('  ');
            expect(manifestEditor.value).toContain('\n');
        });
    });

    describe('Bidirectional Sync', () => {
        beforeEach(async () => {
            await editor.initialize();
        });

        it('should load document content into source editors', () => {
            const htmlEditor = editorContainer.querySelector('#html-editor') as HTMLTextAreaElement;
            const cssEditor = editorContainer.querySelector('#css-editor') as HTMLTextAreaElement;
            
            expect(htmlEditor?.value).toContain('New LIV Document');
            expect(cssEditor?.value).toBeDefined();
        });

        it('should sync source changes to visual editor', async () => {
            (editor as any).switchMode('source');
            
            const htmlEditor = editorContainer.querySelector('#html-editor') as HTMLTextAreaElement;
            htmlEditor.value = '<h1>Updated Title</h1>';
            
            await (editor as any).syncWithVisualEditor();
            
            const document = editor.getDocument();
            expect(document?.content.html).toBe('<h1>Updated Title</h1>');
        });

        it('should notify WASM backend of source changes', () => {
            const sendEventSpy = jest.spyOn((editor as any).sandbox, 'sendEvent').mockResolvedValue(undefined);
            
            (editor as any).switchMode('source');
            const htmlEditor = editorContainer.querySelector('#html-editor') as HTMLTextAreaElement;
            htmlEditor.value = '<p>Test</p>';
            htmlEditor.dispatchEvent(new Event('input'));
            
            expect(sendEventSpy).toHaveBeenCalledWith('source-changed', expect.objectContaining({
                language: 'html',
                content: '<p>Test</p>',
                operation: 'source-change'
            }));
        });
    });

    describe('Keyboard Shortcuts', () => {
        beforeEach(async () => {
            await editor.initialize();
            (editor as any).switchMode('source');
        });

        it('should handle Ctrl+S for save', () => {
            const saveDocumentSpy = jest.spyOn(editor, 'saveDocument').mockResolvedValue(undefined);
            
            const htmlEditor = editorContainer.querySelector('#html-editor') as HTMLTextAreaElement;
            const event = new KeyboardEvent('keydown', { key: 's', ctrlKey: true });
            htmlEditor.dispatchEvent(event);
            
            expect(saveDocumentSpy).toHaveBeenCalled();
        });

        it('should handle Ctrl+D for format', () => {
            const formatSpy = jest.spyOn(editor as any, 'formatCurrentCode');
            
            const htmlEditor = editorContainer.querySelector('#html-editor') as HTMLTextAreaElement;
            const event = new KeyboardEvent('keydown', { key: 'd', ctrlKey: true });
            htmlEditor.dispatchEvent(event);
            
            expect(formatSpy).toHaveBeenCalled();
        });
    });

    describe('Error Handling and Reporting', () => {
        beforeEach(async () => {
            await editor.initialize();
            (editor as any).switchMode('source');
        });

        it('should handle validation errors gracefully', async () => {
            // Mock validator to throw error
            const mockValidator = (editor as any).codeValidator;
            jest.spyOn(mockValidator, 'validateHTML').mockRejectedValue(new Error('Validation failed'));
            
            await (editor as any).validateCode('html');
            
            const validationPanel = editorContainer.querySelector('#html-validation');
            expect(validationPanel?.textContent).toContain('Validation error');
        });

        it('should show clickable validation results', async () => {
            const htmlEditor = editorContainer.querySelector('#html-editor') as HTMLTextAreaElement;
            htmlEditor.value = '<div>\n<p>Unclosed div';
            
            await (editor as any).validateCode('html');
            
            const validationItems = editorContainer.querySelectorAll('.validation-item');
            expect(validationItems.length).toBeGreaterThan(0);
            
            // Should have click handlers
            const firstItem = validationItems[0] as HTMLElement;
            expect(firstItem.dataset.line).toBeTruthy();
        });
    });

    describe('Integration with Existing Systems', () => {
        beforeEach(async () => {
            await editor.initialize();
        });

        it('should integrate with existing validation systems', async () => {
            const document = editor.getDocument();
            expect(document).toBeTruthy();
            
            // Validator should have document reference
            const validator = (editor as any).codeValidator;
            expect(validator).toBeTruthy();
        });

        it('should integrate with existing error handling', () => {
            // Should use existing LIVError types
            expect(() => {
                (editor as any).validateCode('invalid-language');
            }).not.toThrow();
        });

        it('should maintain compatibility with visual editor', () => {
            // Should be able to switch between modes
            (editor as any).switchMode('source');
            expect((editor as any).editMode).toBe('source');
            
            (editor as any).switchMode('visual');
            expect((editor as any).editMode).toBe('visual');
        });
    });

    describe('Cleanup', () => {
        beforeEach(async () => {
            await editor.initialize();
        });

        it('should clean up source editor resources on destroy', () => {
            (editor as any).switchMode('source');
            
            // Verify resources exist
            expect((editor as any).sourceEditors.size).toBeGreaterThan(0);
            expect((editor as any).syntaxHighlighter).toBeTruthy();
            expect((editor as any).codeValidator).toBeTruthy();
            
            editor.destroy();
            
            // Verify cleanup
            expect((editor as any).sourceEditors.size).toBe(0);
            expect((editor as any).syntaxHighlighter).toBeNull();
            expect((editor as any).codeValidator).toBeNull();
        });
    });
});