import { LIVEditor } from '../src/editor';

describe('Visual Editing Capabilities', () => {
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

    describe('Drag and Drop', () => {
        beforeEach(async () => {
            await editor.initialize();
        });

        it('should make elements draggable', () => {
            // Insert a test element
            editor.insertElement('paragraph');
            
            const visualEditor = editorContainer.querySelector('#visual-editor');
            const paragraph = visualEditor?.querySelector('p');
            
            expect(paragraph?.draggable).toBeTruthy();
        });

        it('should handle drag start event', () => {
            editor.insertElement('paragraph');
            
            const visualEditor = editorContainer.querySelector('#visual-editor');
            const paragraph = visualEditor?.querySelector('p') as HTMLElement;
            
            // Simulate drag start
            const dragEvent = new DragEvent('dragstart', {
                bubbles: true,
                dataTransfer: new DataTransfer()
            });
            
            paragraph.dispatchEvent(dragEvent);
            
            expect(paragraph.classList.contains('dragging')).toBeTruthy();
        });

        it('should create drop zones', () => {
            editor.insertElement('container');
            
            const visualEditor = editorContainer.querySelector('#visual-editor');
            const container = visualEditor?.querySelector('.container');
            
            expect(container).toBeTruthy();
        });

        it('should handle element movement', () => {
            // Insert container and paragraph
            editor.insertElement('container');
            editor.insertElement('paragraph');
            
            const visualEditor = editorContainer.querySelector('#visual-editor');
            const container = visualEditor?.querySelector('.container') as HTMLElement;
            const paragraph = visualEditor?.querySelector('p') as HTMLElement;
            
            // Simulate moving paragraph into container
            const initialParent = paragraph.parentElement;
            
            // Use the private method via type assertion for testing
            (editor as any).moveElement(paragraph, container);
            
            expect(paragraph.parentElement).toBe(container);
            expect(paragraph.parentElement).not.toBe(initialParent);
        });
    });

    describe('Visual Style Panel', () => {
        beforeEach(async () => {
            await editor.initialize();
        });

        it('should create visual style panel', () => {
            const stylePanel = document.querySelector('.visual-style-panel');
            expect(stylePanel).toBeTruthy();
        });

        it('should show style panel when element is selected', () => {
            editor.insertElement('paragraph');
            
            const visualEditor = editorContainer.querySelector('#visual-editor');
            const paragraph = visualEditor?.querySelector('p') as HTMLElement;
            
            // Select the element
            (editor as any).selectElement(paragraph);
            
            const stylePanel = document.querySelector('.visual-style-panel') as HTMLElement;
            expect(stylePanel.style.display).toBe('block');
        });

        it('should populate style panel with current values', () => {
            editor.insertElement('paragraph');
            
            const visualEditor = editorContainer.querySelector('#visual-editor');
            const paragraph = visualEditor?.querySelector('p') as HTMLElement;
            
            // Set some styles
            paragraph.style.fontSize = '20px';
            paragraph.style.color = 'red';
            
            // Select the element
            (editor as any).selectElement(paragraph);
            
            const stylePanel = document.querySelector('.visual-style-panel');
            const fontSizeInput = stylePanel?.querySelector('#font-size') as HTMLInputElement;
            const textColorInput = stylePanel?.querySelector('#text-color') as HTMLInputElement;
            
            expect(fontSizeInput?.value).toBe('20');
            expect(textColorInput?.value).toBe('#ff0000');
        });

        it('should apply style changes to selected element', () => {
            editor.insertElement('paragraph');
            
            const visualEditor = editorContainer.querySelector('#visual-editor');
            const paragraph = visualEditor?.querySelector('p') as HTMLElement;
            
            // Select the element
            (editor as any).selectElement(paragraph);
            
            const stylePanel = document.querySelector('.visual-style-panel');
            const bgColorInput = stylePanel?.querySelector('#bg-color') as HTMLInputElement;
            
            // Change background color
            bgColorInput.value = '#ff0000';
            bgColorInput.dispatchEvent(new Event('input', { bubbles: true }));
            
            expect(paragraph.style.backgroundColor).toBe('rgb(255, 0, 0)');
        });
    });

    describe('Element Creation with WASM Integration', () => {
        beforeEach(async () => {
            await editor.initialize();
        });

        it('should create elements with unique IDs', () => {
            editor.insertElement('heading');
            editor.insertElement('paragraph');
            
            const visualEditor = editorContainer.querySelector('#visual-editor');
            const heading = visualEditor?.querySelector('h2');
            const paragraph = visualEditor?.querySelector('p');
            
            expect(heading?.id).toBeTruthy();
            expect(paragraph?.id).toBeTruthy();
            expect(heading?.id).not.toBe(paragraph?.id);
        });

        it('should add data attributes to created elements', () => {
            editor.insertElement('interactive');
            
            const visualEditor = editorContainer.querySelector('#visual-editor');
            const interactive = visualEditor?.querySelector('.interactive-element');
            
            expect(interactive?.getAttribute('data-element-type')).toBe('interactive');
            expect(interactive?.getAttribute('data-editable')).toBe('true');
        });

        it('should create different element types correctly', () => {
            const elementTypes = ['heading', 'paragraph', 'image', 'link', 'container', 'interactive', 'chart'];
            
            elementTypes.forEach(type => {
                editor.insertElement(type);
            });
            
            const visualEditor = editorContainer.querySelector('#visual-editor');
            
            expect(visualEditor?.querySelector('h2')).toBeTruthy(); // heading
            expect(visualEditor?.querySelector('p')).toBeTruthy(); // paragraph
            expect(visualEditor?.querySelector('img')).toBeTruthy(); // image
            expect(visualEditor?.querySelector('a')).toBeTruthy(); // link
            expect(visualEditor?.querySelector('.container')).toBeTruthy(); // container
            expect(visualEditor?.querySelector('.interactive-element')).toBeTruthy(); // interactive
            expect(visualEditor?.querySelector('.chart-container')).toBeTruthy(); // chart
        });
    });

    describe('Element Selection and Resize', () => {
        beforeEach(async () => {
            await editor.initialize();
        });

        it('should add selection styling to selected elements', () => {
            editor.insertElement('paragraph');
            
            const visualEditor = editorContainer.querySelector('#visual-editor');
            const paragraph = visualEditor?.querySelector('p') as HTMLElement;
            
            // Select the element
            (editor as any).selectElement(paragraph);
            
            expect(paragraph.classList.contains('editor-selected')).toBeTruthy();
        });

        it('should add resize handles to selected elements', () => {
            editor.insertElement('paragraph');
            
            const visualEditor = editorContainer.querySelector('#visual-editor');
            const paragraph = visualEditor?.querySelector('p') as HTMLElement;
            
            // Select the element
            (editor as any).selectElement(paragraph);
            
            const resizeHandles = paragraph.querySelectorAll('.resize-handle');
            expect(resizeHandles.length).toBe(8); // 8 resize handles (corners and sides)
        });

        it('should remove resize handles when selecting different element', () => {
            editor.insertElement('paragraph');
            editor.insertElement('heading');
            
            const visualEditor = editorContainer.querySelector('#visual-editor');
            const paragraph = visualEditor?.querySelector('p') as HTMLElement;
            const heading = visualEditor?.querySelector('h2') as HTMLElement;
            
            // Select paragraph first
            (editor as any).selectElement(paragraph);
            expect(paragraph.querySelectorAll('.resize-handle').length).toBe(8);
            
            // Select heading
            (editor as any).selectElement(heading);
            expect(paragraph.querySelectorAll('.resize-handle').length).toBe(0);
            expect(heading.querySelectorAll('.resize-handle').length).toBe(8);
        });
    });

    describe('WASM Backend Integration', () => {
        beforeEach(async () => {
            await editor.initialize();
        });

        it('should notify WASM backend of element creation', async () => {
            const sendEventSpy = jest.spyOn((editor as any).sandbox, 'sendEvent').mockResolvedValue(undefined);
            
            editor.insertElement('paragraph');
            
            expect(sendEventSpy).toHaveBeenCalledWith('element-created', expect.objectContaining({
                elementType: 'paragraph',
                operation: 'create'
            }));
        });

        it('should notify WASM backend of element selection', async () => {
            const sendEventSpy = jest.spyOn((editor as any).sandbox, 'sendEvent').mockResolvedValue(undefined);
            
            editor.insertElement('paragraph');
            
            const visualEditor = editorContainer.querySelector('#visual-editor');
            const paragraph = visualEditor?.querySelector('p') as HTMLElement;
            
            (editor as any).selectElement(paragraph);
            
            expect(sendEventSpy).toHaveBeenCalledWith('element-selected', expect.objectContaining({
                elementId: paragraph.id
            }));
        });

        it('should notify WASM backend of style changes', async () => {
            const sendEventSpy = jest.spyOn((editor as any).sandbox, 'sendEvent').mockResolvedValue(undefined);
            
            editor.insertElement('paragraph');
            
            const visualEditor = editorContainer.querySelector('#visual-editor');
            const paragraph = visualEditor?.querySelector('p') as HTMLElement;
            
            // Select and change style
            (editor as any).selectElement(paragraph);
            (editor as any).notifyWASMStyleChange(paragraph, 'color', 'red');
            
            expect(sendEventSpy).toHaveBeenCalledWith('style-changed', expect.objectContaining({
                elementId: paragraph.id,
                property: 'color',
                value: 'red',
                operation: 'style-change'
            }));
        });

        it('should notify WASM backend of element moves', async () => {
            const sendEventSpy = jest.spyOn((editor as any).sandbox, 'sendEvent').mockResolvedValue(undefined);
            
            editor.insertElement('container');
            editor.insertElement('paragraph');
            
            const visualEditor = editorContainer.querySelector('#visual-editor');
            const container = visualEditor?.querySelector('.container') as HTMLElement;
            const paragraph = visualEditor?.querySelector('p') as HTMLElement;
            
            (editor as any).notifyWASMElementMoved(paragraph, container);
            
            expect(sendEventSpy).toHaveBeenCalledWith('element-moved', expect.objectContaining({
                elementId: paragraph.id,
                newParentId: container.id,
                operation: 'move'
            }));
        });
    });

    describe('Style Injection', () => {
        beforeEach(async () => {
            await editor.initialize();
        });

        it('should inject editor styles into document head', () => {
            const styleElement = document.getElementById('liv-editor-styles');
            expect(styleElement).toBeTruthy();
            expect(styleElement?.tagName).toBe('STYLE');
        });

        it('should not inject styles multiple times', async () => {
            // Initialize another editor
            const editor2 = new LIVEditor(
                editorContainer,
                previewContainer,
                toolbarContainer,
                propertiesContainer
            );
            await editor2.initialize();
            
            const styleElements = document.querySelectorAll('#liv-editor-styles');
            expect(styleElements.length).toBe(1);
            
            editor2.destroy();
        });
    });

    describe('Cleanup', () => {
        beforeEach(async () => {
            await editor.initialize();
        });

        it('should clean up visual style panel on destroy', () => {
            editor.insertElement('paragraph');
            
            const visualEditor = editorContainer.querySelector('#visual-editor');
            const paragraph = visualEditor?.querySelector('p') as HTMLElement;
            (editor as any).selectElement(paragraph);
            
            // Verify style panel exists
            let stylePanel = document.querySelector('.visual-style-panel');
            expect(stylePanel).toBeTruthy();
            
            // Destroy editor
            editor.destroy();
            
            // Verify style panel is removed
            stylePanel = document.querySelector('.visual-style-panel');
            expect(stylePanel).toBeFalsy();
        });

        it('should remove resize handles on destroy', () => {
            editor.insertElement('paragraph');
            
            const visualEditor = editorContainer.querySelector('#visual-editor');
            const paragraph = visualEditor?.querySelector('p') as HTMLElement;
            (editor as any).selectElement(paragraph);
            
            // Verify resize handles exist
            let resizeHandles = document.querySelectorAll('.resize-handle');
            expect(resizeHandles.length).toBeGreaterThan(0);
            
            // Destroy editor
            editor.destroy();
            
            // Verify resize handles are removed
            resizeHandles = document.querySelectorAll('.resize-handle');
            expect(resizeHandles.length).toBe(0);
        });
    });
});