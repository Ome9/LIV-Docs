/**
 * HTML and Markdown Converter Tests
 * Tests HTML and Markdown export/import functionality
 */

import { HTMLMarkdownConverter, HTMLExportOptions, MarkdownExportOptions, HTMLImportOptions, MarkdownImportOptions } from '../src/html-markdown-converter';
import { LIVDocument } from '../src/document';

// Use Jest globals from setup
import './setup';

describe('HTML and Markdown Converter Tests', () => {
    let converter: HTMLMarkdownConverter;
    let testDocument: LIVDocument;

    // Set timeout for all tests
    jest.setTimeout(15000);

    beforeEach(async () => {
        converter = new HTMLMarkdownConverter();
        
        // Create test document
        const manifest = {
            version: '1.0',
            metadata: {
                title: 'Test Document',
                author: 'Test Author',
                created: new Date().toISOString(),
                modified: new Date().toISOString(),
                description: 'Test document for conversion',
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
            html: '<h1>Test Document</h1><p>This is a <strong>test</strong> document with <em>formatting</em>.</p>',
            css: 'body { font-family: Arial, sans-serif; }',
            interactiveSpec: '',
            staticFallback: '<h1>Test Document</h1><p>This is a <strong>test</strong> document with <em>formatting</em>.</p>'
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
        converter.destroy();
    });

    describe('HTML Export', () => {
        it('should export document to standalone HTML', async () => {
            const options: HTMLExportOptions = {
                standalone: true,
                includeCSS: true,
                includeInteractive: false
            };

            const html = await converter.exportToHTML(testDocument, options);

            expect(html).toContain('<!DOCTYPE html>');
            expect(html).toContain('<title>Test Document</title>');
            expect(html).toContain('<h1>Test Document</h1>');
            expect(html).toContain('font-family: Arial, sans-serif');
        });

        it('should export document to HTML without CSS', async () => {
            const options: HTMLExportOptions = {
                standalone: true,
                includeCSS: false
            };

            const html = await converter.exportToHTML(testDocument, options);

            expect(html).toContain('<h1>Test Document</h1>');
            expect(html).not.toContain('font-family: Arial, sans-serif');
        });

        it('should export document to minified HTML', async () => {
            const options: HTMLExportOptions = {
                standalone: true,
                minify: true
            };

            const html = await converter.exportToHTML(testDocument, options);

            expect(html).not.toContain('\n    ');
            expect(html).toContain('<h1>Test Document</h1>');
        });
    });

    describe('Markdown Export', () => {
        it('should export document to Markdown', async () => {
            const options: MarkdownExportOptions = {
                preserveFormatting: true,
                flavor: 'github'
            };

            const markdown = await converter.exportToMarkdown(testDocument, options);

            expect(markdown).toContain('# Test Document');
            expect(markdown).toContain('**test**');
            expect(markdown).toContain('*formatting*');
        });

        it('should export document to Markdown without formatting', async () => {
            const options: MarkdownExportOptions = {
                preserveFormatting: false
            };

            const markdown = await converter.exportToMarkdown(testDocument, options);

            expect(markdown).toContain('# Test Document');
            expect(markdown).toContain('test');
            expect(markdown).toContain('formatting');
        });
    });

    describe('HTML Import', () => {
        it('should import HTML and create LIV document', async () => {
            const htmlContent = `
                <!DOCTYPE html>
                <html>
                <head>
                    <title>Imported Document</title>
                    <meta name="author" content="Import Author">
                </head>
                <body>
                    <h1>Imported Document</h1>
                    <p>This is imported content.</p>
                </body>
                </html>
            `;

            const options: HTMLImportOptions = {
                extractCSS: true,
                sanitize: true,
                createManifest: true
            };

            const document = await converter.importFromHTML(htmlContent, options);

            expect(document.manifest.metadata.title).toBe('Imported Document');
            expect(document.content.html).toContain('<h1>Imported Document</h1>');
            expect(document.content.staticFallback).toContain('<h1>Imported Document</h1>');
        });
    });

    describe('Markdown Import', () => {
        it('should import Markdown and create LIV document', async () => {
            const markdownContent = `
# Imported Markdown

This is **imported** content from *Markdown*.

## Section 2

- List item 1
- List item 2

\`\`\`
code block
\`\`\`
            `;

            const options: MarkdownImportOptions = {
                generateCSS: true,
                preserveFormatting: true
            };

            const document = await converter.importFromMarkdown(markdownContent, options);

            expect(document.manifest.metadata.title).toBe('Imported Markdown');
            expect(document.content.html).toContain('<h1>Imported Markdown</h1>');
            expect(document.content.html).toContain('<strong>imported</strong>');
            expect(document.content.css).toContain('font-family');
        });
    });

    describe('Error Handling', () => {
        it('should handle export errors gracefully', async () => {
            const invalidDocument = { ...testDocument, content: null } as any;

            await expect(converter.exportToHTML(invalidDocument)).rejects.toThrow();
        });

        it('should handle import errors gracefully', async () => {
            const invalidHTML = '<html><head><title>Invalid</title></head><body><h1>Test</h1><script>alert("xss")</script></body></html>';

            const options: HTMLImportOptions = {
                sanitize: true
            };

            const document = await converter.importFromHTML(invalidHTML, options);
            
            // Should sanitize script tags
            expect(document.content.html).not.toContain('<script>');
        });
    });
});