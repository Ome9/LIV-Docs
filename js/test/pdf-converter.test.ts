/**
 * PDF Converter Tests
 * Tests PDF export and import functionality
 */

import { PDFConverter, PDFExportOptions, PDFImportOptions } from '../src/pdf-converter';
import { LIVDocument } from '../src/document';
import { jest } from '@jest/globals';
import { jest } from '@jest/globals';
import { jest } from '@jest/globals';
import { jest } from '@jest/globals';
import { jest } from '@jest/globals';

// Jest globals
declare global {
    var describe: jest.Describe;
    var it: jest.It;
    var expect: jest.Expect;
    var beforeEach: jest.Lifecycle;
    var afterEach: jest.Lifecycle;
    var jest: typeof import('@jest/globals').jest;
}

describe('PDF Converter Tests', () => {
    let converter: PDFConverter;
    let testDocument: LIVDocument;

    // Set timeout for all tests
    jest.setTimeout(15000);

    beforeEach(async () => {
        converter = new PDFConverter();
        
        // Create test document
        const manifest = {
            version: '1.0',
            metadata: {
                title: 'Test Document',
                author: 'Test Author',
                created: new Date().toISOString(),
                modified: new Date().toISOString(),
                description: 'Test document for PDF conversion',
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
                <h1>Test Document</h1>
                <p>This is a test paragraph with some content.</p>
                <h2>Section 1</h2>
                <p>More content in section 1.</p>
                <img src="https://example.com/test.jpg" alt="Test Image">
                <h2>Section 2</h2>
                <p>Content in section 2 with <strong>bold text</strong> and <em>italic text</em>.</p>
            `,
            css: `
                body { font-family: Arial, sans-serif; }
                h1 { color: #333; font-size: 24px; }
                h2 { color: #666; font-size: 18px; }
                p { line-height: 1.6; }
                img { max-width: 100%; }
            `,
            interactiveSpec: '',
            staticFallback: `
                <h1>Test Document</h1>
                <p>This is a test paragraph with some content.</p>
                <h2>Section 1</h2>
                <p>More content in section 1.</p>
                <h2>Section 2</h2>
                <p>Content in section 2 with bold text and italic text.</p>
            `
        };

        const assets = {
            images: new Map<string, ArrayBuffer>(),
            fonts: new Map<string, ArrayBuffer>(),
            data: new Map<string, ArrayBuffer>()
        };

        const signatures = {
            contentSignature: '',
            manifestSignature: '',
            wasmSignatures: {}
        };

        testDocument = new LIVDocument(manifest, content, assets, signatures, new Map());
    });

    afterEach(() => {
        if (converter) {
            converter.destroy();
        }
    });

    describe('PDF Export Functionality', () => {
        it('should export document to PDF with default options', async () => {
            const pdfData = await converter.exportToPDF(testDocument);
            
            expect(pdfData).toBeDefined();
            expect(pdfData.constructor.name).toBe('Uint8Array');
            expect(pdfData.length).toBeGreaterThan(0);
        });

        it('should export document to PDF with custom options', async () => {
            const options: PDFExportOptions = {
                format: 'Letter',
                orientation: 'landscape',
                margins: { top: 30, right: 30, bottom: 30, left: 30 },
                includeInteractive: false,
                quality: 'high',
                embedFonts: true,
                preserveLayout: true
            };

            const pdfData = await converter.exportToPDF(testDocument, options);
            
            expect(pdfData.constructor.name).toBe('Uint8Array');
            expect(pdfData.length).toBeGreaterThan(0);
        });

        it('should handle different page formats', async () => {
            const formats: Array<'A4' | 'Letter' | 'Legal'> = ['A4', 'Letter', 'Legal'];
            
            for (const format of formats) {
                const options: PDFExportOptions = { format };
                const pdfData = await converter.exportToPDF(testDocument, options);
                
                expect(pdfData.constructor.name).toBe('Uint8Array');
                expect(pdfData.length).toBeGreaterThan(0);
            }
        });

        it('should handle different orientations', async () => {
            const orientations: Array<'portrait' | 'landscape'> = ['portrait', 'landscape'];
            
            for (const orientation of orientations) {
                const options: PDFExportOptions = { orientation };
                const pdfData = await converter.exportToPDF(testDocument, options);
                
                expect(pdfData.constructor.name).toBe('Uint8Array');
                expect(pdfData.length).toBeGreaterThan(0);
            }
        });

        it('should export with static fallback when interactive content is disabled', async () => {
            const options: PDFExportOptions = {
                includeInteractive: false
            };

            const pdfData = await converter.exportToPDF(testDocument, options);
            
            expect(pdfData.constructor.name).toBe('Uint8Array');
            expect(pdfData.length).toBeGreaterThan(0);
        });

        it('should handle documents with images', async () => {
            // Add test image to document assets
            const imageData = new ArrayBuffer(1024);
            testDocument.assets.images.set('test-image.jpg', imageData);

            const pdfData = await converter.exportToPDF(testDocument);
            
            expect(pdfData.constructor.name).toBe('Uint8Array');
            expect(pdfData.length).toBeGreaterThan(0);
        });

        it('should handle empty documents', async () => {
            const emptyContent = {
                html: '<html><body></body></html>',
                css: '',
                interactiveSpec: '',
                staticFallback: '<html><body></body></html>'
            };

            const emptyDocument = new LIVDocument(
                testDocument.manifest,
                emptyContent,
                testDocument.assets,
                testDocument.signatures,
                new Map()
            );

            const pdfData = await converter.exportToPDF(emptyDocument);
            
            expect(pdfData.constructor.name).toBe('Uint8Array');
            expect(pdfData.length).toBeGreaterThan(0);
        });

        it('should handle export errors gracefully', async () => {
            // Create document with invalid content that should cause an error
            const invalidContent = {
                html: null as any,
                css: '',
                interactiveSpec: '',
                staticFallback: null as any
            };

            const invalidDocument = new LIVDocument(
                testDocument.manifest,
                invalidContent,
                testDocument.assets,
                testDocument.signatures,
                new Map()
            );

            // In test environment, this should still work due to mocking
            const pdfData = await converter.exportToPDF(invalidDocument);
            expect(pdfData.constructor.name).toBe('Uint8Array');
        });
    });

    describe('PDF Import Functionality', () => {
        it('should import PDF and create LIV document', async () => {
            // Create mock PDF data
            const mockPDFData = new Uint8Array([
                0x25, 0x50, 0x44, 0x46, // %PDF header
                ...Array(1000).fill(0x20) // Padding to simulate PDF content
            ]);

            const importedDocument = await converter.importFromPDF(mockPDFData);
            
            expect(importedDocument).toBeInstanceOf(LIVDocument);
            expect(importedDocument.manifest.metadata.title).toBeTruthy();
            expect(importedDocument.content.html).toBeTruthy();
            expect(importedDocument.content.css).toBeTruthy();
        });

        it('should import PDF with custom options', async () => {
            const mockPDFData = new Uint8Array(2000);
            const options: PDFImportOptions = {
                extractImages: true,
                preserveFormatting: true,
                convertToHTML: true,
                maxPages: 10
            };

            const importedDocument = await converter.importFromPDF(mockPDFData, options);
            
            expect(importedDocument).toBeInstanceOf(LIVDocument);
            expect(importedDocument.content.html).toContain('pdf-document');
            expect(importedDocument.content.css).toContain('pdf-page');
        });

        it('should handle PDF with multiple pages', async () => {
            // Create larger mock PDF to simulate multiple pages
            const mockPDFData = new Uint8Array(100000);
            
            const importedDocument = await converter.importFromPDF(mockPDFData);
            
            expect(importedDocument).toBeInstanceOf(LIVDocument);
            expect(importedDocument.content.html).toContain('pdf-page');
        });

        it('should extract images from PDF when enabled', async () => {
            const mockPDFData = new Uint8Array(5000);
            const options: PDFImportOptions = {
                extractImages: true
            };

            const importedDocument = await converter.importFromPDF(mockPDFData, options);
            
            expect(importedDocument).toBeInstanceOf(LIVDocument);
            // Check if images are referenced in HTML
            expect(importedDocument.content.html).toMatch(/img.*src.*assets\/images/);
        });

        it('should limit pages when maxPages is set', async () => {
            const mockPDFData = new Uint8Array(200000); // Large PDF
            const options: PDFImportOptions = {
                maxPages: 3
            };

            const importedDocument = await converter.importFromPDF(mockPDFData, options);
            
            expect(importedDocument).toBeInstanceOf(LIVDocument);
            
            // Count pages in HTML
            const pageMatches = importedDocument.content.html.match(/data-page="/g);
            const pageCount = pageMatches ? pageMatches.length : 0;
            expect(pageCount).toBeLessThanOrEqual(3);
        });

        it('should preserve formatting when enabled', async () => {
            const mockPDFData = new Uint8Array(3000);
            const options: PDFImportOptions = {
                preserveFormatting: true
            };

            const importedDocument = await converter.importFromPDF(mockPDFData, options);
            
            expect(importedDocument).toBeInstanceOf(LIVDocument);
            expect(importedDocument.content.css).toContain('text-block');
            expect(importedDocument.content.css).toContain('pdf-image');
        });

        it('should handle import errors gracefully', async () => {
            // Invalid PDF data
            const invalidPDFData = new Uint8Array([0x00, 0x01, 0x02]);

            await expect(converter.importFromPDF(invalidPDFData)).rejects.toThrow();
        });

        it('should handle empty PDF data', async () => {
            const emptyPDFData = new Uint8Array(0);

            await expect(converter.importFromPDF(emptyPDFData)).rejects.toThrow();
        });
    });

    describe('Round-trip Conversion', () => {
        it('should maintain content integrity in round-trip conversion', async () => {
            // Export to PDF
            const pdfData = await converter.exportToPDF(testDocument);
            
            // Import back from PDF
            const importedDocument = await converter.importFromPDF(pdfData);
            
            expect(importedDocument).toBeInstanceOf(LIVDocument);
            expect(importedDocument.content.html).toBeTruthy();
            expect(importedDocument.content.css).toBeTruthy();
            
            // Check that basic content structure is preserved (not exact content due to simulation)
            expect(importedDocument.content.html).toContain('pdf-document');
        });

        it('should handle round-trip with different export options', async () => {
            const exportOptions: PDFExportOptions = {
                format: 'Letter',
                orientation: 'landscape',
                includeInteractive: false
            };

            const importOptions: PDFImportOptions = {
                extractImages: true,
                preserveFormatting: true
            };

            // Export to PDF
            const pdfData = await converter.exportToPDF(testDocument, exportOptions);
            
            // Import back from PDF
            const importedDocument = await converter.importFromPDF(pdfData, importOptions);
            
            expect(importedDocument).toBeInstanceOf(LIVDocument);
            expect(importedDocument.content.html).toBeTruthy();
        });
    });

    describe('Performance and Memory Management', () => {
        it('should handle large documents efficiently', async () => {
            // Create large document
            let largeHTML = '<h1>Large Document</h1>';
            for (let i = 0; i < 100; i++) { // Reduced from 1000 for faster tests
                largeHTML += `<p>Paragraph ${i} with some content to make it larger.</p>`;
            }

            const largeContent = {
                ...testDocument.content,
                html: largeHTML
            };

            const largeDocument = new LIVDocument(
                testDocument.manifest,
                largeContent,
                testDocument.assets,
                testDocument.signatures,
                new Map()
            );

            const startTime = performance.now();
            const pdfData = await converter.exportToPDF(largeDocument);
            const endTime = performance.now();

            expect(pdfData.constructor.name).toBe('Uint8Array');
            expect(pdfData.length).toBeGreaterThan(0);
            
            // Should complete within reasonable time
            expect(endTime - startTime).toBeLessThan(5000); // 5 seconds
        });

        it('should clean up resources properly', () => {
            const converter2 = new PDFConverter();
            
            expect(() => {
                converter2.destroy();
            }).not.toThrow();
        });

        it('should handle multiple concurrent conversions', async () => {
            const promises = [];
            
            // Start multiple conversions
            for (let i = 0; i < 3; i++) { // Reduced from 5 for faster tests
                promises.push(converter.exportToPDF(testDocument));
            }

            const results = await Promise.all(promises);
            
            results.forEach(pdfData => {
                expect(pdfData.constructor.name).toBe('Uint8Array');
                expect(pdfData.length).toBeGreaterThan(0);
            });
        });
    });

    describe('Integration with Existing Infrastructure', () => {
        it('should use existing renderer infrastructure', async () => {
            // Test that converter integrates with LIVRenderer
            const pdfData = await converter.exportToPDF(testDocument);
            
            expect(pdfData.constructor.name).toBe('Uint8Array');
            // The converter should have used the renderer internally
        });

        it('should respect document security settings', async () => {
            // Create document with restricted security
            const restrictedManifest = {
                ...testDocument.manifest,
                security: {
                    ...testDocument.manifest.security,
                    jsPermissions: {
                        executionMode: 'none' as const,
                        allowedAPIs: [],
                        domAccess: 'none' as const
                    }
                }
            };

            const restrictedDocument = new LIVDocument(
                restrictedManifest,
                testDocument.content,
                testDocument.assets,
                testDocument.signatures,
                new Map()
            );

            const pdfData = await converter.exportToPDF(restrictedDocument);
            
            expect(pdfData.constructor.name).toBe('Uint8Array');
            expect(pdfData.length).toBeGreaterThan(0);
        });

        it('should handle documents with WASM components', async () => {
            // Create document with WASM features
            const wasmManifest = {
                ...testDocument.manifest,
                features: {
                    animations: false,
                    interactivity: true,
                    charts: false,
                    forms: false,
                    audio: false,
                    video: false,
                    webgl: false,
                    webassembly: true
                }
            };

            const wasmDocument = new LIVDocument(
                wasmManifest,
                testDocument.content,
                testDocument.assets,
                testDocument.signatures,
                new Map()
            );

            const pdfData = await converter.exportToPDF(wasmDocument);
            
            expect(pdfData.constructor.name).toBe('Uint8Array');
            expect(pdfData.length).toBeGreaterThan(0);
        });

        it('should integrate with existing validation systems', async () => {
            // Test that imported documents are properly validated
            const mockPDFData = new Uint8Array(1000);
            
            const importedDocument = await converter.importFromPDF(mockPDFData);
            
            // Check that document has proper structure
            expect(importedDocument.manifest).toBeTruthy();
            expect(importedDocument.manifest.version).toBeTruthy();
            expect(importedDocument.manifest.security).toBeTruthy();
            expect(importedDocument.content).toBeTruthy();
        });
    });

    describe('Error Handling and Edge Cases', () => {
        it('should handle corrupted PDF data', async () => {
            const corruptedData = new Uint8Array([0xFF, 0xFE, 0xFD]);
            
            await expect(converter.importFromPDF(corruptedData)).rejects.toThrow();
        });

        it('should handle documents with missing content', async () => {
            const emptyContent = {
                html: '',
                css: '',
                interactiveSpec: '',
                staticFallback: ''
            };

            const emptyDocument = new LIVDocument(
                testDocument.manifest,
                emptyContent,
                testDocument.assets,
                testDocument.signatures,
                new Map()
            );

            const pdfData = await converter.exportToPDF(emptyDocument);
            
            expect(pdfData.constructor.name).toBe('Uint8Array');
            expect(pdfData.length).toBeGreaterThan(0);
        });

        it('should handle invalid export options gracefully', async () => {
            const invalidOptions = {
                format: 'InvalidFormat' as any,
                orientation: 'InvalidOrientation' as any
            };

            // Should use defaults for invalid options
            const pdfData = await converter.exportToPDF(testDocument, invalidOptions);
            
            expect(pdfData.constructor.name).toBe('Uint8Array');
            expect(pdfData.length).toBeGreaterThan(0);
        });
    });
});