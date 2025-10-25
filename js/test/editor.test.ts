import { LIVEditor } from '../src/editor';
import { LIVDocument } from '../src/document';
import { LIVError, LIVErrorType } from '../src/errors';

describe('LIVEditor', () => {
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

    describe('initialization', () => {
        it('should initialize editor with containers', () => {
            expect(editor).toBeDefined();
            expect(editor.getDocument()).toBeNull();
            expect(editor.isDirtyDocument()).toBeFalsy();
        });

        it('should initialize with new document', async () => {
            await editor.initialize();
            
            expect(editor.getDocument()).not.toBeNull();
            expect(toolbarContainer.innerHTML).toContain('editor-toolbar');
            expect(editorContainer.innerHTML).toContain('editor-content-wrapper');
            expect(propertiesContainer.innerHTML).toContain('properties-panel');
        });

        it('should initialize with existing document', async () => {
            const document = await (editor as any).createDocumentFromHTML('<h1>Test Document</h1>');
            
            await editor.initialize(document);
            
            expect(editor.getDocument()).toBe(document);
        });

        it('should throw error if initialization fails', async () => {
            // Mock sandbox initialization failure
            const mockSandbox = editor['sandbox'];
            jest.spyOn(mockSandbox, 'initialize').mockRejectedValue(new Error('Sandbox failed'));

            await expect(editor.initialize()).rejects.toThrow(LIVError);
        });
    });

    describe('document management', () => {
        beforeEach(async () => {
            await editor.initialize();
        });

        it('should create new document with default content', () => {
            const document = editor.getDocument();
            expect(document).not.toBeNull();
            expect(document!.getHTML()).toContain('New LIV Document');
        });

        it('should load document from file', async () => {
            const htmlContent = '<h1>Test Document</h1><p>Test content</p>';
            const file = new File([htmlContent], 'test.html', { type: 'text/html' });
            
            await editor.loadFromFile(file);
            
            const document = editor.getDocument();
            expect(document).not.toBeNull();
            expect(document!.getHTML()).toContain('Test Document');
        });

        it('should export document as HTML', () => {
            const html = editor.exportAsHTML();
            expect(html).toContain('<!DOCTYPE html>');
            expect(html).toContain('New LIV Document');
        });

        it('should mark document as dirty when modified', async () => {
            expect(editor.isDirtyDocument()).toBeFalsy();
            
            // Simulate content change
            editor['markDirty']();
            
            expect(editor.isDirtyDocument()).toBeTruthy();
        });
    });

    describe('editing modes', () => {
        beforeEach(async () => {
            await editor.initialize();
        });

        it('should start in visual mode', () => {
            const visualEditor = editorContainer.querySelector('#visual-editor');
            const sourceWrapper = editorContainer.querySelector('#source-editor-wrapper');
            
            expect(visualEditor?.classList.contains('active')).toBeTruthy();
            expect(sourceWrapper?.classList.contains('hidden')).toBeTruthy();
        });

        it('should switch to source mode', () => {
            editor['switchMode']('source');
            
            const visualEditor = editorContainer.querySelector('#visual-editor');
            const sourceWrapper = editorContainer.querySelector('#source-editor-wrapper');
            
            expect(visualEditor?.classList.contains('hidden')).toBeTruthy();
            expect(sourceWrapper?.classList.contains('active')).toBeTruthy();
        });

        it('should switch back to visual mode', () => {
            editor['switchMode']('source');
            editor['switchMode']('visual');
            
            const visualEditor = editorContainer.querySelector('#visual-editor');
            const sourceWrapper = editorContainer.querySelector('#source-editor-wrapper');
            
            expect(visualEditor?.classList.contains('active')).toBeTruthy();
            expect(sourceWrapper?.classList.contains('hidden')).toBeTruthy();
        });
    });

    describe('element insertion', () => {
        beforeEach(async () => {
            await editor.initialize();
        });

        it('should insert heading element', () => {
            editor.insertElement('heading');
            expect(editor.isDirtyDocument()).toBeTruthy();
        });

        it('should insert paragraph element', () => {
            editor.insertElement('paragraph');
            expect(editor.isDirtyDocument()).toBeTruthy();
        });

        it('should insert image element with properties', () => {
            editor.insertElement('image', {
                src: 'https://example.com/image.jpg',
                alt: 'Test image'
            });
            expect(editor.isDirtyDocument()).toBeTruthy();
        });

        it('should insert link element with properties', () => {
            editor.insertElement('link', {
                href: 'https://example.com',
                text: 'Example link'
            });
            expect(editor.isDirtyDocument()).toBeTruthy();
        });

        it('should throw error for invalid element type', () => {
            expect(() => {
                editor.insertElement('invalid-element');
            }).toThrow(LIVError);
        });
    });

    describe('element properties', () => {
        beforeEach(async () => {
            await editor.initialize();
        });

        it('should update element properties', () => {
            // Create a test element
            const element = document.createElement('p');
            element.textContent = 'Test paragraph';
            editor['selectedElement'] = element;

            editor.updateElementProperties({
                textContent: 'Updated paragraph',
                className: 'test-class'
            });

            expect(element.textContent).toBe('Updated paragraph');
            expect(element.className).toBe('test-class');
            expect(editor.isDirtyDocument()).toBeTruthy();
        });

        it('should throw error when no element selected', () => {
            editor['selectedElement'] = null;

            expect(() => {
                editor.updateElementProperties({ textContent: 'test' });
            }).toThrow(LIVError);
        });
    });

    describe('document operations', () => {
        beforeEach(async () => {
            await editor.initialize();
        });

        it('should save document', async () => {
            const saveEventSpy = jest.fn();
            editorContainer.addEventListener('document-save', saveEventSpy);

            await editor.saveDocument();

            expect(saveEventSpy).toHaveBeenCalled();
            expect(editor.isDirtyDocument()).toBeFalsy();
        });

        it('should preview document', async () => {
            const previewEventSpy = jest.fn();
            editorContainer.addEventListener('document-preview', previewEventSpy);

            await editor.previewDocument();

            expect(previewEventSpy).toHaveBeenCalled();
        });

        it('should throw error when saving without document', async () => {
            editor['document'] = null;

            await expect(editor.saveDocument()).rejects.toThrow(LIVError);
        });

        it('should throw error when previewing without document', async () => {
            editor['document'] = null;

            await expect(editor.previewDocument()).rejects.toThrow(LIVError);
        });
    });

    describe('error handling', () => {
        it('should throw error when not initialized', () => {
            expect(() => {
                editor.insertElement('heading');
            }).toThrow(LIVError);
        });

        it('should handle document loading errors', async () => {
            const invalidFile = new File(['invalid content'], 'test.txt', { type: 'text/plain' });
            
            await expect(editor.loadFromFile(invalidFile)).rejects.toThrow(LIVError);
        });
    });

    describe('cleanup', () => {
        it('should cleanup resources on destroy', async () => {
            await editor.initialize();
            
            editor.destroy();
            
            expect(editor.getDocument()).toBeNull();
            expect(editor['selectedElement']).toBeNull();
            expect(editor['isInitialized']).toBeFalsy();
        });
    });
});