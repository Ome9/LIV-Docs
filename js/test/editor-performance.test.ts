import { LIVEditor } from '../src/editor';

describe('Editor Performance Tests', () => {
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

    describe('Large Document Performance', () => {
        it('should handle large documents efficiently', async () => {
            await editor.initialize();

            const startTime = performance.now();

            // Create a large document with many elements
            for (let i = 0; i < 500; i++) {
                editor.insertElement('paragraph', { text: `Paragraph ${i} with some content to test performance.` });
                
                if (i % 10 === 0) {
                    editor.insertElement('heading', { text: `Section ${Math.floor(i / 10)}` });
                }
                
                if (i % 25 === 0) {
                    editor.insertElement('container');
                }
            }

            const endTime = performance.now();
            const duration = endTime - startTime;

            // Should complete within reasonable time
            expect(duration).toBeLessThan(10000); // 10 seconds

            // Verify document integrity
            const document = editor.getDocument();
            expect(document).toBeTruthy();
            expect(document?.content.html).toContain('Paragraph 499');
        });

        it('should handle rapid element creation and deletion', async () => {
            await editor.initialize();

            const startTime = performance.now();

            // Rapidly create and delete elements
            for (let i = 0; i < 100; i++) {
                // Create element
                editor.insertElement('paragraph', { text: `Rapid ${i}` });
                
                // Select and delete every other element
                if (i % 2 === 0) {
                    const paragraphs = editorContainer.querySelectorAll('p');
                    const lastParagraph = paragraphs[paragraphs.length - 1] as HTMLElement;
                    if (lastParagraph) {
                        (editor as any).selectElement(lastParagraph);
                        lastParagraph.remove();
                    }
                }
            }

            const endTime = performance.now();
            const duration = endTime - startTime;

            expect(duration).toBeLessThan(5000); // 5 seconds
        });

        it('should handle large source code efficiently', async () => {
            await editor.initialize();
            (editor as any).switchMode('source');

            const startTime = performance.now();

            // Generate large HTML content
            let largeHTML = '<div class="large-document">\n';
            for (let i = 0; i < 1000; i++) {
                largeHTML += `  <div class="section-${i}">\n`;
                largeHTML += `    <h2>Section ${i}</h2>\n`;
                largeHTML += `    <p>This is paragraph content for section ${i}. It contains some text to make the document larger.</p>\n`;
                largeHTML += `    <div class="subsection">\n`;
                largeHTML += `      <p>Subsection content with more text and <a href="#link-${i}">links</a>.</p>\n`;
                largeHTML += `    </div>\n`;
                largeHTML += `  </div>\n`;
            }
            largeHTML += '</div>';

            const htmlEditor = editorContainer.querySelector('#html-editor') as HTMLTextAreaElement;
            htmlEditor.value = largeHTML;
            htmlEditor.dispatchEvent(new Event('input'));

            const endTime = performance.now();
            const duration = endTime - startTime;

            expect(duration).toBeLessThan(3000); // 3 seconds

            // Verify syntax highlighting was applied
            const overlay = editorContainer.querySelector('#html-overlay');
            expect(overlay?.innerHTML).toContain('span');
        });
    });

    describe('Memory Management Performance', () => {
        it('should not leak memory with repeated operations', async () => {
            await editor.initialize();

            // Measure initial memory usage (approximate)
            const initialElementCount = document.querySelectorAll('*').length;

            // Perform many operations that could cause memory leaks
            for (let i = 0; i < 100; i++) {
                // Create elements
                editor.insertElement('paragraph', { text: `Memory test ${i}` });
                editor.insertElement('container');
                
                // Select elements
                const paragraphs = editorContainer.querySelectorAll('p');
                if (paragraphs.length > 0) {
                    (editor as any).selectElement(paragraphs[paragraphs.length - 1]);
                }
                
                // Switch modes
                (editor as any).switchMode('source');
                (editor as any).switchMode('visual');
                
                // Clean up some elements
                if (i % 10 === 0) {
                    const containers = editorContainer.querySelectorAll('.container');
                    containers.forEach((container, index) => {
                        if (index < containers.length / 2) {
                            container.remove();
                        }
                    });
                }
            }

            // Force garbage collection if available
            if ((global as any).gc) {
                (global as any).gc();
            }

            // Check that we haven't accumulated too many DOM elements
            const finalElementCount = document.querySelectorAll('*').length;
            const elementGrowth = finalElementCount - initialElementCount;
            
            // Should not have excessive element growth
            expect(elementGrowth).toBeLessThan(1000);
        });

        it('should clean up event listeners properly', async () => {
            await editor.initialize();

            // Track event listener additions (simplified)
            let eventListenerCount = 0;
            const originalAddEventListener = HTMLElement.prototype.addEventListener;
            HTMLElement.prototype.addEventListener = function(...args) {
                eventListenerCount++;
                return originalAddEventListener.apply(this, args);
            };

            // Perform operations that add event listeners
            for (let i = 0; i < 50; i++) {
                editor.insertElement('paragraph', { text: `Event test ${i}` });
                
                const paragraph = editorContainer.querySelector(`p:last-child`) as HTMLElement;
                if (paragraph) {
                    (editor as any).selectElement(paragraph);
                }
            }

            const listenersAfterOperations = eventListenerCount;

            // Destroy editor
            editor.destroy();

            // Restore original addEventListener
            HTMLElement.prototype.addEventListener = originalAddEventListener;

            // Should have added listeners but not excessively
            expect(listenersAfterOperations).toBeGreaterThan(0);
            expect(listenersAfterOperations).toBeLessThan(500);
        });
    });

    describe('Validation Performance', () => {
        it('should validate large documents quickly', async () => {
            await editor.initialize();
            (editor as any).switchMode('source');

            // Create large HTML content
            let largeHTML = '';
            for (let i = 0; i < 500; i++) {
                largeHTML += `<div class="item-${i}"><h3>Item ${i}</h3><p>Content for item ${i}</p></div>\n`;
            }

            const htmlEditor = editorContainer.querySelector('#html-editor') as HTMLTextAreaElement;
            htmlEditor.value = largeHTML;

            const startTime = performance.now();
            await (editor as any).validateCode('html');
            const endTime = performance.now();

            const duration = endTime - startTime;
            expect(duration).toBeLessThan(2000); // 2 seconds

            // Should complete validation
            const validationPanel = editorContainer.querySelector('#html-validation');
            expect(validationPanel?.classList.contains('success')).toBeTruthy();
        });

        it('should handle rapid validation requests', async () => {
            await editor.initialize();
            (editor as any).switchMode('source');

            const htmlEditor = editorContainer.querySelector('#html-editor') as HTMLTextAreaElement;

            const startTime = performance.now();

            // Simulate rapid typing with validation
            const promises = [];
            for (let i = 0; i < 20; i++) {
                htmlEditor.value = `<div>Content ${i}</div>`;
                promises.push((editor as any).validateCode('html'));
            }

            await Promise.all(promises);

            const endTime = performance.now();
            const duration = endTime - startTime;

            expect(duration).toBeLessThan(5000); // 5 seconds for all validations
        });
    });

    describe('Rendering Performance', () => {
        it('should render complex documents efficiently', async () => {
            await editor.initialize();

            // Create complex document structure
            for (let i = 0; i < 100; i++) {
                editor.insertElement('container');
                const container = editorContainer.querySelector('.container:last-child') as HTMLElement;
                
                (editor as any).selectElement(container);
                editor.insertElement('heading', { text: `Section ${i}` });
                editor.insertElement('paragraph', { text: `Content for section ${i}` });
                
                if (i % 5 === 0) {
                    editor.insertElement('image', { 
                        src: `https://example.com/image-${i}.jpg`,
                        alt: `Image ${i}`
                    });
                }
                
                if (i % 10 === 0) {
                    editor.insertElement('interactive');
                }
            }

            const startTime = performance.now();

            // Switch to source mode and back (triggers re-rendering)
            (editor as any).switchMode('source');
            await (editor as any).syncWithVisualEditor();
            (editor as any).switchMode('visual');

            const endTime = performance.now();
            const duration = endTime - startTime;

            expect(duration).toBeLessThan(3000); // 3 seconds
        });

        it('should handle rapid style changes efficiently', async () => {
            await editor.initialize();

            // Create elements to style
            for (let i = 0; i < 50; i++) {
                editor.insertElement('paragraph', { text: `Styled paragraph ${i}` });
            }

            const paragraphs = editorContainer.querySelectorAll('p');
            const startTime = performance.now();

            // Apply rapid style changes
            paragraphs.forEach((paragraph, index) => {
                (editor as any).selectElement(paragraph);
                
                // Simulate style panel changes
                (editor as any).notifyWASMStyleChange(paragraph, 'color', `hsl(${index * 7}, 70%, 50%)`);
                (editor as any).notifyWASMStyleChange(paragraph, 'fontSize', `${14 + (index % 10)}px`);
            });

            const endTime = performance.now();
            const duration = endTime - startTime;

            expect(duration).toBeLessThan(2000); // 2 seconds
        });
    });

    describe('WASM Communication Performance', () => {
        it('should handle high-frequency WASM communications', async () => {
            await editor.initialize();

            const sendEventSpy = jest.spyOn((editor as any).sandbox, 'sendEvent').mockResolvedValue(undefined);

            const startTime = performance.now();

            // Generate many WASM communications
            for (let i = 0; i < 200; i++) {
                editor.insertElement('paragraph', { text: `WASM test ${i}` });
                
                const paragraph = editorContainer.querySelector('p:last-child') as HTMLElement;
                if (paragraph) {
                    (editor as any).selectElement(paragraph);
                    (editor as any).notifyWASMStyleChange(paragraph, 'color', 'red');
                }
            }

            const endTime = performance.now();
            const duration = endTime - startTime;

            expect(duration).toBeLessThan(5000); // 5 seconds
            expect(sendEventSpy).toHaveBeenCalledTimes(600); // 3 calls per iteration (create, select, style)
        });

        it('should handle WASM communication failures gracefully', async () => {
            await editor.initialize();

            // Mock WASM failures
            const sendEventSpy = jest.spyOn((editor as any).sandbox, 'sendEvent').mockRejectedValue(new Error('WASM error'));

            const startTime = performance.now();

            // Operations should continue despite failures
            for (let i = 0; i < 100; i++) {
                editor.insertElement('paragraph', { text: `Failure test ${i}` });
            }

            const endTime = performance.now();
            const duration = endTime - startTime;

            // Should not be significantly slower due to error handling
            expect(duration).toBeLessThan(3000); // 3 seconds
            expect(sendEventSpy).toHaveBeenCalled();
        });
    });

    describe('Stress Testing', () => {
        it('should survive stress test with mixed operations', async () => {
            await editor.initialize();

            const operations = [
                () => editor.insertElement('paragraph', { text: 'Stress test paragraph' }),
                () => editor.insertElement('heading', { text: 'Stress test heading' }),
                () => editor.insertElement('container'),
                () => editor.insertElement('image', { src: 'https://example.com/stress.jpg' }),
                () => (editor as any).switchMode('source'),
                () => (editor as any).switchMode('visual'),
                () => {
                    const elements = editorContainer.querySelectorAll('[data-editable="true"]');
                    if (elements.length > 0) {
                        const randomElement = elements[Math.floor(Math.random() * elements.length)] as HTMLElement;
                        (editor as any).selectElement(randomElement);
                    }
                },
                () => {
                    const containers = editorContainer.querySelectorAll('.container');
                    if (containers.length > 1) {
                        containers[0].remove();
                    }
                }
            ];

            const startTime = performance.now();

            // Perform random operations
            for (let i = 0; i < 500; i++) {
                const randomOperation = operations[Math.floor(Math.random() * operations.length)];
                try {
                    randomOperation();
                } catch (error) {
                    // Some operations may fail, that's okay for stress testing
                }
            }

            const endTime = performance.now();
            const duration = endTime - startTime;

            expect(duration).toBeLessThan(15000); // 15 seconds

            // Editor should still be functional
            expect(editor.getDocument()).toBeTruthy();
            expect(() => editor.insertElement('paragraph', { text: 'Post-stress test' })).not.toThrow();
        });

        it('should handle concurrent operations', async () => {
            await editor.initialize();

            const startTime = performance.now();

            // Simulate concurrent operations
            const promises = [];
            
            for (let i = 0; i < 50; i++) {
                promises.push(new Promise<void>((resolve) => {
                    setTimeout(() => {
                        editor.insertElement('paragraph', { text: `Concurrent ${i}` });
                        resolve();
                    }, Math.random() * 100);
                }));
            }

            await Promise.all(promises);

            const endTime = performance.now();
            const duration = endTime - startTime;

            expect(duration).toBeLessThan(5000); // 5 seconds

            // Should have created all elements
            const paragraphs = editorContainer.querySelectorAll('p');
            expect(paragraphs.length).toBeGreaterThanOrEqual(50);
        });
    });
});