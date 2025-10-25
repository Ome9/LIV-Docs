/**
 * EPUB Converter Tests
 * Tests EPUB export/import functionality
 */

import { EPUBConverter, EPUBExportOptions, EPUBImportOptions } from '../src/epub-converter';
import { LIVDocument } from '../src/document';

// Use Jest globals from setup
import './setup';

describe('EPUB Converter Tests', () => {
    let converter: EPUBConverter;
    let testDocument: LIVDocument;

    // Set timeout for all tests
    jest.setTimeout(15000);

    beforeEach(async () => {
        converter = new EPUBConverter();
        
        // Create test document
        const manifest = {
            version: '1.0',
            metadata: {
                title: 'Test EPUB Document',
                author: 'Test Author',
                created: new Date().toISOString(),
                modified: new Date().toISOString(),
                description: 'Test document for EPUB conversion',
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
                    allowedAPIs: ['dom'],
                    domAccess: 'read' as const
                },
                networkPolicy: {
                    allowOutbound: false,
                    allowedHosts: [],
                    allowedPorts: []
                },
                storagePolicy: {
                    allowLocalStorage: false,
                    allowSessionStorage: false,
                    allowIndexedDB: false,
                    allowCookies: false
                },
                contentSecurityPolicy: "default-src 'self';",
                trustedDomains: []
            },
            features: {
                animations: false,
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
            html: `
                <h1>Chapter 1: Introduction</h1>
                <p>This is the first chapter with <strong>bold</strong> text.</p>
                
                <h2>Section 1.1</h2>
                <p>This is a subsection with <em>italic</em> text.</p>
                
                <h1>Chapter 2: Main Content</h1>
                <p>This is the second chapter.</p>
                <ul>
                    <li>List item 1</li>
                    <li>List item 2</li>
                </ul>
                
                <blockquote>
                    This is a blockquote example.
                </blockquote>
            `,
            css: `
                body { 
                    font-family: Georgia, serif; 
                    line-height: 1.6; 
                }
                h1 { 
                    color: #333; 
                    border-bottom: 2px solid #333; 
                }
                h2 { 
                    color: #666; 
                }
            `,
            interactiveSpec: '',
            staticFallback: `
                <h1>Chapter 1: Introduction</h1>
                <p>This is the first chapter with <strong>bold</strong> text.</p>
                
                <h2>Section 1.1</h2>
                <p>This is a subsection with <em>italic</em> text.</p>
                
                <h1>Chapter 2: Main Content</h1>
                <p>This is the second chapter.</p>
                <ul>
                    <li>List item 1</li>
                    <li>List item 2</li>
                </ul>
                
                <blockquote>
                    This is a blockquote example.
                </blockquote>
            `
        };

        const assets = {
            images: new Map<string, ArrayBuffer>(),
            fonts: new Map<string, ArrayBuffer>(),
            data: new Map<string, ArrayBuffer>()
        };

        // Add test image
        const testImageData = new ArrayBuffer(100);
        assets.images.set('test-image.jpg', testImageData);

        const signatures = {
            contentSignature: '',
            manifestSignature: '',
            wasmSignatures: {}
        };

        testDocument = new LIVDocument(manifest, content, assets, signatures, new Map());
    });

    afterEach(() => {
        converter.destroy();
    });

    describe('EPUB Export', () => {
        it('should export document to EPUB with default options', async () => {
            const epubBuffer = await converter.exportToEPUB(testDocument);

            expect(epubBuffer).toBeInstanceOf(ArrayBuffer);
            expect(epubBuffer.byteLength).toBeGreaterThan(0);
        });

        it('should export document to EPUB with custom metadata', async () => {
            const options: EPUBExportOptions = {
                metadata: {
                    title: 'Custom EPUB Title',
                    author: 'Custom Author',
                    publisher: 'Test Publisher',
                    description: 'Custom description for EPUB',
                    isbn: '978-0-123456-78-9'
                }
            };

            const epubBuffer = await converter.exportToEPUB(testDocument, options);

            expect(epubBuffer).toBeInstanceOf(ArrayBuffer);
            expect(epubBuffer.byteLength).toBeGreaterThan(0);
        });

        it('should export document with chapter breaks on H1', async () => {
            const options: EPUBExportOptions = {
                chapterBreaks: 'h1',
                generateTOC: true
            };

            const epubBuffer = await converter.exportToEPUB(testDocument, options);

            expect(epubBuffer).toBeInstanceOf(ArrayBuffer);
            expect(epubBuffer.byteLength).toBeGreaterThan(0);
        });

        it('should export document with chapter breaks on H2', async () => {
            const options: EPUBExportOptions = {
                chapterBreaks: 'h2',
                generateTOC: true
            };

            const epubBuffer = await converter.exportToEPUB(testDocument, options);

            expect(epubBuffer).toBeInstanceOf(ArrayBuffer);
        });

        it('should export document without chapter breaks', async () => {
            const options: EPUBExportOptions = {
                chapterBreaks: 'none',
                generateTOC: false
            };

            const epubBuffer = await converter.exportToEPUB(testDocument, options);

            expect(epubBuffer).toBeInstanceOf(ArrayBuffer);
        });

        it('should export document without CSS', async () => {
            const options: EPUBExportOptions = {
                includeCSS: false
            };

            const epubBuffer = await converter.exportToEPUB(testDocument, options);

            expect(epubBuffer).toBeInstanceOf(ArrayBuffer);
        });

        it('should export document without images', async () => {
            const options: EPUBExportOptions = {
                includeImages: false
            };

            const epubBuffer = await converter.exportToEPUB(testDocument, options);

            expect(epubBuffer).toBeInstanceOf(ArrayBuffer);
        });

        it('should handle documents with no content gracefully', async () => {
            const emptyDocument = { ...testDocument };
            emptyDocument.content.html = '';
            emptyDocument.content.staticFallback = '';

            const epubBuffer = await converter.exportToEPUB(emptyDocument);

            expect(epubBuffer).toBeInstanceOf(ArrayBuffer);
        });

        it('should handle export errors gracefully', async () => {
            const invalidDocument = { ...testDocument, content: null } as any;

            await expect(converter.exportToEPUB(invalidDocument)).rejects.toThrow();
        });
    });

    describe('EPUB Import', () => {
        it('should import EPUB and create LIV document', async () => {
            // Create a mock EPUB ArrayBuffer (simplified for testing)
            const mockEPUBData = new ArrayBuffer(1024);

            const options: EPUBImportOptions = {
                extractImages: true,
                generateCSS: true,
                combineChapters: true
            };

            const document = await converter.importFromEPUB(mockEPUBData, options);

            expect(document.manifest.metadata.title).toBe('Imported EPUB');
            expect(document.content.html).toContain('section');
            expect(document.content.css).toContain('font-family');
        });

        it('should import EPUB with custom options', async () => {
            const mockEPUBData = new ArrayBuffer(1024);

            const options: EPUBImportOptions = {
                extractImages: false,
                generateCSS: false,
                combineChapters: false,
                preserveStructure: true
            };

            const document = await converter.importFromEPUB(mockEPUBData, options);

            expect(document.manifest.metadata.title).toBe('Imported EPUB');
            expect(document.content.css).toBe('');
        });

        it('should handle import errors gracefully', async () => {
            const invalidEPUBData = new ArrayBuffer(0); // Empty buffer

            await expect(converter.importFromEPUB(invalidEPUBData)).rejects.toThrow();
        });

        it('should preserve EPUB structure in data assets', async () => {
            const mockEPUBData = new ArrayBuffer(1024);

            const document = await converter.importFromEPUB(mockEPUBData);

            expect(document.assets.data.has('original-epub-structure.json')).toBe(true);
        });
    });

    describe('Round-trip Conversion', () => {
        it('should maintain content integrity in round-trip conversion', async () => {
            // Export to EPUB
            const epubBuffer = await converter.exportToEPUB(testDocument);

            // Import back to LIV
            const importedDocument = await converter.importFromEPUB(epubBuffer);

            expect(importedDocument.manifest.metadata.title).toBeDefined();
            expect(importedDocument.content.html).toContain('Chapter');
        });

        it('should handle round-trip with different export options', async () => {
            const exportOptions: EPUBExportOptions = {
                chapterBreaks: 'h1',
                includeCSS: true,
                generateTOC: true
            };

            const importOptions: EPUBImportOptions = {
                combineChapters: true,
                generateCSS: true
            };

            // Export to EPUB
            const epubBuffer = await converter.exportToEPUB(testDocument, exportOptions);

            // Import back to LIV
            const importedDocument = await converter.importFromEPUB(epubBuffer, importOptions);

            expect(importedDocument.content.html).toBeDefined();
            expect(importedDocument.content.css).toBeDefined();
        });
    });

    describe('Performance and Memory Management', () => {
        it('should handle large documents efficiently', async () => {
            // Create a large document
            const largeContent = Array(1000).fill('<p>Large content paragraph.</p>').join('\n');
            const largeDocument = { ...testDocument };
            largeDocument.content.html = largeContent;
            largeDocument.content.staticFallback = largeContent;

            const startTime = Date.now();
            const epubBuffer = await converter.exportToEPUB(largeDocument);
            const endTime = Date.now();

            expect(epubBuffer).toBeInstanceOf(ArrayBuffer);
            expect(endTime - startTime).toBeLessThan(5000); // Should complete within 5 seconds
        });

        it('should clean up resources properly', () => {
            const converter1 = new EPUBConverter();
            const converter2 = new EPUBConverter();

            converter1.destroy();
            converter2.destroy();

            // Should not throw errors
            expect(() => converter1.destroy()).not.toThrow();
            expect(() => converter2.destroy()).not.toThrow();
        });

        it('should handle multiple concurrent conversions', async () => {
            const promises = Array(5).fill(null).map(() => 
                converter.exportToEPUB(testDocument)
            );

            const results = await Promise.all(promises);

            results.forEach(result => {
                expect(result).toBeInstanceOf(ArrayBuffer);
                expect(result.byteLength).toBeGreaterThan(0);
            });
        });
    });

    describe('Integration with Existing Infrastructure', () => {
        it('should use existing ZIP container system', async () => {
            const epubBuffer = await converter.exportToEPUB(testDocument);

            // EPUB is essentially a ZIP file
            expect(epubBuffer).toBeInstanceOf(ArrayBuffer);
            expect(epubBuffer.byteLength).toBeGreaterThan(0);
        });

        it('should respect document security settings', async () => {
            const document = await converter.importFromEPUB(new ArrayBuffer(1024));

            expect(document.manifest.security.wasmPermissions.allowNetworking).toBe(false);
            expect(document.manifest.security.jsPermissions.executionMode).toBe('sandboxed');
        });

        it('should handle documents with existing assets', async () => {
            // Add multiple assets to test document
            const imageData1 = new ArrayBuffer(200);
            const imageData2 = new ArrayBuffer(300);
            testDocument.assets.images.set('image1.png', imageData1);
            testDocument.assets.images.set('image2.jpg', imageData2);

            const epubBuffer = await converter.exportToEPUB(testDocument);

            expect(epubBuffer).toBeInstanceOf(ArrayBuffer);
        });

        it('should integrate with existing validation systems', async () => {
            const document = await converter.importFromEPUB(new ArrayBuffer(1024));

            // Should create valid LIV document structure
            expect(document.manifest).toBeDefined();
            expect(document.content).toBeDefined();
            expect(document.assets).toBeDefined();
            expect(document.signatures).toBeDefined();
        });
    });

    describe('Error Handling and Edge Cases', () => {
        it('should handle corrupted EPUB data', async () => {
            const corruptedData = new ArrayBuffer(10);
            // Fill with invalid data
            const view = new Uint8Array(corruptedData);
            view.fill(255);

            await expect(converter.importFromEPUB(corruptedData)).rejects.toThrow();
        });

        it('should handle documents with missing content', async () => {
            const emptyDocument = { ...testDocument };
            emptyDocument.content.html = '';
            emptyDocument.content.staticFallback = '';

            const epubBuffer = await converter.exportToEPUB(emptyDocument);

            expect(epubBuffer).toBeInstanceOf(ArrayBuffer);
        });

        it('should handle invalid export options gracefully', async () => {
            const invalidOptions = {
                chapterBreaks: 'invalid' as any,
                metadata: null as any
            };

            // Should use defaults for invalid options
            const epubBuffer = await converter.exportToEPUB(testDocument, invalidOptions);

            expect(epubBuffer).toBeInstanceOf(ArrayBuffer);
        });

        it('should handle documents with special characters in metadata', async () => {
            const specialDocument = { ...testDocument };
            specialDocument.manifest.metadata.title = 'Test & <Special> "Characters"';
            specialDocument.manifest.metadata.author = 'Author with \'quotes\'';

            const epubBuffer = await converter.exportToEPUB(specialDocument);

            expect(epubBuffer).toBeInstanceOf(ArrayBuffer);
        });
    });

    describe('EPUB Structure Validation', () => {
        it('should generate valid EPUB structure', async () => {
            const options: EPUBExportOptions = {
                generateTOC: true,
                chapterBreaks: 'h1'
            };

            const epubBuffer = await converter.exportToEPUB(testDocument, options);

            // Basic validation - EPUB should be a valid ZIP-like structure
            expect(epubBuffer).toBeInstanceOf(ArrayBuffer);
            expect(epubBuffer.byteLength).toBeGreaterThan(100); // Should have substantial content
        });

        it('should handle documents with complex HTML structure', async () => {
            const complexDocument = { ...testDocument };
            complexDocument.content.html = `
                <div class="container">
                    <header>
                        <h1>Complex Document</h1>
                        <nav>
                            <ul>
                                <li><a href="#section1">Section 1</a></li>
                                <li><a href="#section2">Section 2</a></li>
                            </ul>
                        </nav>
                    </header>
                    <main>
                        <section id="section1">
                            <h2>Section 1</h2>
                            <p>Content with <code>inline code</code> and <a href="http://example.com">links</a>.</p>
                            <pre><code>
                                function example() {
                                    return "code block";
                                }
                            </code></pre>
                        </section>
                        <section id="section2">
                            <h2>Section 2</h2>
                            <table>
                                <tr><th>Header 1</th><th>Header 2</th></tr>
                                <tr><td>Cell 1</td><td>Cell 2</td></tr>
                            </table>
                        </section>
                    </main>
                </div>
            `;

            const epubBuffer = await converter.exportToEPUB(complexDocument);

            expect(epubBuffer).toBeInstanceOf(ArrayBuffer);
        });
    });
});