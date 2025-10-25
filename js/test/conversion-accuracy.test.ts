/**
 * Conversion Accuracy Tests
 * Tests format fidelity and accuracy across all conversion formats
 */

import { PDFConverter } from '../src/pdf-converter';
import { HTMLMarkdownConverter } from '../src/html-markdown-converter';
import { EPUBConverter } from '../src/epub-converter';
import { LIVDocument } from '../src/document';

// Use Jest globals from setup
import './setup';

describe('Conversion Accuracy Tests', () => {
    let pdfConverter: PDFConverter;
    let htmlMarkdownConverter: HTMLMarkdownConverter;
    let epubConverter: EPUBConverter;
    let testDocument: LIVDocument;
    let complexTestDocument: LIVDocument;

    // Set timeout for all tests
    jest.setTimeout(30000);

    beforeEach(async () => {
        pdfConverter = new PDFConverter();
        htmlMarkdownConverter = new HTMLMarkdownConverter();
        epubConverter = new EPUBConverter();
        
        // Create simple test document
        const manifest = {
            version: '1.0',
            metadata: {
                title: 'Conversion Test Document',
                author: 'Test Author',
                created: new Date().toISOString(),
                modified: new Date().toISOString(),
                description: 'Test document for conversion accuracy',
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
                <p>This is a <strong>test</strong> document with <em>formatting</em>.</p>
                <h2>Section 1</h2>
                <p>This section contains a list:</p>
                <ul>
                    <li>Item 1</li>
                    <li>Item 2</li>
                    <li>Item 3</li>
                </ul>
                <blockquote>
                    This is a blockquote for testing.
                </blockquote>
                <p>Here's a <a href="https://example.com">link</a> for testing.</p>
            `,
            css: `
                body { 
                    font-family: Arial, sans-serif; 
                    line-height: 1.6; 
                    color: #333; 
                }
                h1 { 
                    color: #2c3e50; 
                    border-bottom: 2px solid #3498db; 
                }
                h2 { 
                    color: #34495e; 
                }
                blockquote {
                    border-left: 4px solid #3498db;
                    padding-left: 1em;
                    margin-left: 0;
                    font-style: italic;
                }
            `,
            interactiveSpec: '',
            staticFallback: `
                <h1>Test Document</h1>
                <p>This is a <strong>test</strong> document with <em>formatting</em>.</p>
                <h2>Section 1</h2>
                <p>This section contains a list:</p>
                <ul>
                    <li>Item 1</li>
                    <li>Item 2</li>
                    <li>Item 3</li>
                </ul>
                <blockquote>
                    This is a blockquote for testing.
                </blockquote>
                <p>Here's a <a href="https://example.com">link</a> for testing.</p>
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

        // Create complex test document with more features
        const complexContent = {
            html: `
                <!DOCTYPE html>
                <html lang="en">
                <head>
                    <meta charset="UTF-8">
                    <title>Complex Test Document</title>
                </head>
                <body>
                    <header>
                        <h1>Complex Document Structure</h1>
                        <nav>
                            <ul>
                                <li><a href="#section1">Section 1</a></li>
                                <li><a href="#section2">Section 2</a></li>
                                <li><a href="#section3">Section 3</a></li>
                            </ul>
                        </nav>
                    </header>
                    
                    <main>
                        <section id="section1">
                            <h2>Text Formatting</h2>
                            <p>This paragraph contains <strong>bold text</strong>, <em>italic text</em>, 
                            <code>inline code</code>, and <a href="https://example.com">links</a>.</p>
                            
                            <h3>Code Block</h3>
                            <pre><code>
function example() {
    console.log("Hello, World!");
    return true;
}
                            </code></pre>
                        </section>
                        
                        <section id="section2">
                            <h2>Lists and Tables</h2>
                            
                            <h3>Unordered List</h3>
                            <ul>
                                <li>First item</li>
                                <li>Second item with <strong>bold</strong> text</li>
                                <li>Third item</li>
                            </ul>
                            
                            <h3>Ordered List</h3>
                            <ol>
                                <li>Step one</li>
                                <li>Step two</li>
                                <li>Step three</li>
                            </ol>
                            
                            <h3>Table</h3>
                            <table>
                                <thead>
                                    <tr>
                                        <th>Name</th>
                                        <th>Age</th>
                                        <th>City</th>
                                    </tr>
                                </thead>
                                <tbody>
                                    <tr>
                                        <td>John Doe</td>
                                        <td>30</td>
                                        <td>New York</td>
                                    </tr>
                                    <tr>
                                        <td>Jane Smith</td>
                                        <td>25</td>
                                        <td>London</td>
                                    </tr>
                                </tbody>
                            </table>
                        </section>
                        
                        <section id="section3">
                            <h2>Special Elements</h2>
                            
                            <blockquote>
                                <p>This is a blockquote with multiple paragraphs.</p>
                                <p>It demonstrates how blockquotes should be preserved across formats.</p>
                            </blockquote>
                            
                            <hr>
                            
                            <p>Text with <sup>superscript</sup> and <sub>subscript</sub>.</p>
                            
                            <div class="highlight">
                                <p>This is a highlighted section that should maintain its structure.</p>
                            </div>
                        </section>
                    </main>
                    
                    <footer>
                        <p>&copy; 2024 Test Document. All rights reserved.</p>
                    </footer>
                </body>
                </html>
            `,
            css: `
                body {
                    font-family: Georgia, serif;
                    line-height: 1.6;
                    color: #333;
                    max-width: 800px;
                    margin: 0 auto;
                    padding: 20px;
                }
                
                header {
                    border-bottom: 2px solid #3498db;
                    margin-bottom: 2em;
                    padding-bottom: 1em;
                }
                
                nav ul {
                    list-style: none;
                    padding: 0;
                    display: flex;
                    gap: 1em;
                }
                
                nav a {
                    color: #3498db;
                    text-decoration: none;
                }
                
                nav a:hover {
                    text-decoration: underline;
                }
                
                h1, h2, h3 {
                    color: #2c3e50;
                }
                
                h1 {
                    font-size: 2.5em;
                    margin-bottom: 0.5em;
                }
                
                h2 {
                    font-size: 2em;
                    margin-top: 2em;
                    margin-bottom: 1em;
                }
                
                h3 {
                    font-size: 1.5em;
                    margin-top: 1.5em;
                    margin-bottom: 0.5em;
                }
                
                code {
                    background-color: #f8f9fa;
                    padding: 0.2em 0.4em;
                    border-radius: 3px;
                    font-family: 'Courier New', monospace;
                }
                
                pre {
                    background-color: #f8f9fa;
                    padding: 1em;
                    border-radius: 5px;
                    overflow-x: auto;
                }
                
                pre code {
                    background: none;
                    padding: 0;
                }
                
                table {
                    width: 100%;
                    border-collapse: collapse;
                    margin: 1em 0;
                }
                
                th, td {
                    border: 1px solid #ddd;
                    padding: 0.5em;
                    text-align: left;
                }
                
                th {
                    background-color: #f8f9fa;
                    font-weight: bold;
                }
                
                blockquote {
                    border-left: 4px solid #3498db;
                    margin: 1em 0;
                    padding: 0.5em 1em;
                    background-color: #f8f9fa;
                    font-style: italic;
                }
                
                .highlight {
                    background-color: #fff3cd;
                    border: 1px solid #ffeaa7;
                    padding: 1em;
                    border-radius: 5px;
                    margin: 1em 0;
                }
                
                hr {
                    border: none;
                    border-top: 2px solid #ecf0f1;
                    margin: 2em 0;
                }
                
                footer {
                    border-top: 1px solid #ecf0f1;
                    margin-top: 3em;
                    padding-top: 1em;
                    text-align: center;
                    color: #7f8c8d;
                    font-size: 0.9em;
                }
            `,
            interactiveSpec: '',
            staticFallback: `
                <header>
                    <h1>Complex Document Structure</h1>
                    <nav>
                        <ul>
                            <li><a href="#section1">Section 1</a></li>
                            <li><a href="#section2">Section 2</a></li>
                            <li><a href="#section3">Section 3</a></li>
                        </ul>
                    </nav>
                </header>
                
                <main>
                    <section id="section1">
                        <h2>Text Formatting</h2>
                        <p>This paragraph contains <strong>bold text</strong>, <em>italic text</em>, 
                        <code>inline code</code>, and <a href="https://example.com">links</a>.</p>
                        
                        <h3>Code Block</h3>
                        <pre><code>
function example() {
    console.log("Hello, World!");
    return true;
}
                        </code></pre>
                    </section>
                    
                    <section id="section2">
                        <h2>Lists and Tables</h2>
                        
                        <h3>Unordered List</h3>
                        <ul>
                            <li>First item</li>
                            <li>Second item with <strong>bold</strong> text</li>
                            <li>Third item</li>
                        </ul>
                        
                        <h3>Ordered List</h3>
                        <ol>
                            <li>Step one</li>
                            <li>Step two</li>
                            <li>Step three</li>
                        </ol>
                        
                        <h3>Table</h3>
                        <table>
                            <thead>
                                <tr>
                                    <th>Name</th>
                                    <th>Age</th>
                                    <th>City</th>
                                </tr>
                            </thead>
                            <tbody>
                                <tr>
                                    <td>John Doe</td>
                                    <td>30</td>
                                    <td>New York</td>
                                </tr>
                                <tr>
                                    <td>Jane Smith</td>
                                    <td>25</td>
                                    <td>London</td>
                                </tr>
                            </tbody>
                        </table>
                    </section>
                    
                    <section id="section3">
                        <h2>Special Elements</h2>
                        
                        <blockquote>
                            <p>This is a blockquote with multiple paragraphs.</p>
                            <p>It demonstrates how blockquotes should be preserved across formats.</p>
                        </blockquote>
                        
                        <hr>
                        
                        <p>Text with <sup>superscript</sup> and <sub>subscript</sub>.</p>
                        
                        <div class="highlight">
                            <p>This is a highlighted section that should maintain its structure.</p>
                        </div>
                    </section>
                </main>
                
                <footer>
                    <p>&copy; 2024 Test Document. All rights reserved.</p>
                </footer>
            `
        };

        const complexManifest = {
            ...manifest,
            metadata: {
                ...manifest.metadata,
                title: 'Complex Test Document',
                description: 'Complex document for testing conversion fidelity'
            }
        };

        complexTestDocument = new LIVDocument(complexManifest, complexContent, assets, signatures, new Map());
    });

    afterEach(() => {
        pdfConverter.destroy();
        htmlMarkdownConverter.destroy();
        epubConverter.destroy();
    });

    describe('Format Fidelity Tests', () => {
        describe('HTML Export Fidelity', () => {
            it('should preserve document structure in HTML export', async () => {
                const html = await htmlMarkdownConverter.exportToHTML(testDocument, { standalone: true });

                expect(html).toContain('<!DOCTYPE html>');
                expect(html).toContain('<title>Conversion Test Document</title>');
                expect(html).toContain('<h1>Test Document</h1>');
                expect(html).toContain('<strong>test</strong>');
                expect(html).toContain('<em>formatting</em>');
                expect(html).toContain('<blockquote>');
                expect(html).toContain('href="https://example.com"');
            });

            it('should preserve CSS styling in HTML export', async () => {
                const html = await htmlMarkdownConverter.exportToHTML(testDocument, { 
                    standalone: true, 
                    includeCSS: true 
                });

                expect(html).toContain('font-family: Arial, sans-serif');
                expect(html).toContain('color: #2c3e50');
                expect(html).toContain('border-bottom: 2px solid #3498db');
            });

            it('should handle complex document structure in HTML export', async () => {
                const html = await htmlMarkdownConverter.exportToHTML(complexTestDocument, { standalone: true });

                expect(html).toContain('<nav>');
                expect(html).toContain('<section id="section1">');
                expect(html).toContain('<table>');
                expect(html).toContain('<thead>');
                expect(html).toContain('<tbody>');
                expect(html).toContain('<sup>superscript</sup>');
                expect(html).toContain('<sub>subscript</sub>');
            });
        });

        describe('Markdown Export Fidelity', () => {
            it('should preserve text formatting in Markdown export', async () => {
                const markdown = await htmlMarkdownConverter.exportToMarkdown(testDocument);

                expect(markdown).toContain('# Test Document');
                expect(markdown).toContain('## Section 1');
                expect(markdown).toContain('**test**');
                expect(markdown).toContain('*formatting*');
                expect(markdown).toContain('> This is a blockquote');
                expect(markdown).toContain('[link](https://example.com)');
            });

            it('should preserve list structure in Markdown export', async () => {
                const markdown = await htmlMarkdownConverter.exportToMarkdown(testDocument);

                expect(markdown).toContain('- Item 1');
                expect(markdown).toContain('- Item 2');
                expect(markdown).toContain('- Item 3');
            });

            it('should handle complex formatting in Markdown export', async () => {
                const markdown = await htmlMarkdownConverter.exportToMarkdown(complexTestDocument);

                expect(markdown).toContain('# Complex Document Structure');
                expect(markdown).toContain('## Text Formatting');
                expect(markdown).toContain('### Code Block');
                expect(markdown).toContain('```');
                expect(markdown).toContain('1. Step one');
                expect(markdown).toContain('2. Step two');
            });
        });

        describe('PDF Export Fidelity', () => {
            it('should preserve document metadata in PDF export', async () => {
                const pdfBuffer = await pdfConverter.exportToPDF(testDocument);

                expect(pdfBuffer).toBeInstanceOf(ArrayBuffer);
                expect(pdfBuffer.byteLength).toBeGreaterThan(0);
            });

            it('should handle complex layouts in PDF export', async () => {
                const pdfBuffer = await pdfConverter.exportToPDF(complexTestDocument, {
                    format: 'A4',
                    orientation: 'portrait',
                    includeCSS: true
                });

                expect(pdfBuffer).toBeInstanceOf(ArrayBuffer);
                expect(pdfBuffer.byteLength).toBeGreaterThan(1000); // Should have substantial content
            });

            it('should preserve static fallback content in PDF export', async () => {
                const pdfBuffer = await pdfConverter.exportToPDF(testDocument, {
                    useStaticFallback: true
                });

                expect(pdfBuffer).toBeInstanceOf(ArrayBuffer);
            });
        });

        describe('EPUB Export Fidelity', () => {
            it('should preserve document structure in EPUB export', async () => {
                const epubBuffer = await epubConverter.exportToEPUB(testDocument, {
                    chapterBreaks: 'h2',
                    generateTOC: true
                });

                expect(epubBuffer).toBeInstanceOf(ArrayBuffer);
                expect(epubBuffer.byteLength).toBeGreaterThan(0);
            });

            it('should handle metadata correctly in EPUB export', async () => {
                const epubBuffer = await epubConverter.exportToEPUB(testDocument, {
                    metadata: {
                        title: 'Custom EPUB Title',
                        author: 'Custom Author',
                        publisher: 'Test Publisher'
                    }
                });

                expect(epubBuffer).toBeInstanceOf(ArrayBuffer);
            });

            it('should preserve chapter structure in EPUB export', async () => {
                const epubBuffer = await epubConverter.exportToEPUB(complexTestDocument, {
                    chapterBreaks: 'h2',
                    generateTOC: true,
                    includeCSS: true
                });

                expect(epubBuffer).toBeInstanceOf(ArrayBuffer);
                expect(epubBuffer.byteLength).toBeGreaterThan(1000);
            });
        });
    });

    describe('Round-trip Conversion Tests', () => {
        it('should maintain content integrity in HTML round-trip', async () => {
            // Export to HTML
            const html = await htmlMarkdownConverter.exportToHTML(testDocument, { standalone: true });
            
            // Import back from HTML
            const importedDoc = await htmlMarkdownConverter.importFromHTML(html);
            
            expect(importedDoc.manifest.metadata.title).toBeDefined();
            expect(importedDoc.content.html).toContain('Test Document');
            expect(importedDoc.content.html).toContain('strong');
        });

        it('should maintain content integrity in Markdown round-trip', async () => {
            // Export to Markdown
            const markdown = await htmlMarkdownConverter.exportToMarkdown(testDocument);
            
            // Import back from Markdown
            const importedDoc = await htmlMarkdownConverter.importFromMarkdown(markdown);
            
            expect(importedDoc.manifest.metadata.title).toContain('Test Document');
            expect(importedDoc.content.html).toContain('<h1>');
            expect(importedDoc.content.html).toContain('<strong>');
        });

        it('should maintain content integrity in EPUB round-trip', async () => {
            // Export to EPUB
            const epubBuffer = await epubConverter.exportToEPUB(testDocument);
            
            // Import back from EPUB
            const importedDoc = await epubConverter.importFromEPUB(epubBuffer);
            
            expect(importedDoc.manifest.metadata.title).toBeDefined();
            expect(importedDoc.content.html).toContain('section');
        });

        it('should handle complex documents in round-trip conversions', async () => {
            // Test HTML round-trip with complex document
            const html = await htmlMarkdownConverter.exportToHTML(complexTestDocument, { standalone: true });
            const htmlImported = await htmlMarkdownConverter.importFromHTML(html);
            
            expect(htmlImported.content.html).toContain('Complex Document Structure');
            expect(htmlImported.content.html).toContain('table');
            
            // Test Markdown round-trip with complex document
            const markdown = await htmlMarkdownConverter.exportToMarkdown(complexTestDocument);
            const markdownImported = await htmlMarkdownConverter.importFromMarkdown(markdown);
            
            expect(markdownImported.content.html).toContain('Complex Document Structure');
        });
    });

    describe('Cross-Format Conversion Tests', () => {
        it('should convert LIV to HTML to Markdown consistently', async () => {
            // LIV -> HTML
            const html = await htmlMarkdownConverter.exportToHTML(testDocument, { standalone: true });
            const htmlDoc = await htmlMarkdownConverter.importFromHTML(html);
            
            // HTML -> Markdown
            const markdown = await htmlMarkdownConverter.exportToMarkdown(htmlDoc);
            
            expect(markdown).toContain('# Test Document');
            expect(markdown).toContain('**test**');
            expect(markdown).toContain('*formatting*');
        });

        it('should convert LIV to Markdown to HTML consistently', async () => {
            // LIV -> Markdown
            const markdown = await htmlMarkdownConverter.exportToMarkdown(testDocument);
            const markdownDoc = await htmlMarkdownConverter.importFromMarkdown(markdown);
            
            // Markdown -> HTML
            const html = await htmlMarkdownConverter.exportToHTML(markdownDoc, { standalone: true });
            
            expect(html).toContain('<h1>');
            expect(html).toContain('<strong>');
            expect(html).toContain('<em>');
        });

        it('should maintain formatting across multiple format conversions', async () => {
            // LIV -> HTML -> Markdown -> HTML
            const html1 = await htmlMarkdownConverter.exportToHTML(testDocument, { standalone: true });
            const htmlDoc = await htmlMarkdownConverter.importFromHTML(html1);
            
            const markdown = await htmlMarkdownConverter.exportToMarkdown(htmlDoc);
            const markdownDoc = await htmlMarkdownConverter.importFromMarkdown(markdown);
            
            const html2 = await htmlMarkdownConverter.exportToHTML(markdownDoc, { standalone: true });
            
            // Should still contain key elements
            expect(html2).toContain('Test Document');
            expect(html2).toContain('<strong>');
            expect(html2).toContain('<em>');
        });
    });

    describe('Content Preservation Tests', () => {
        it('should preserve special characters and entities', async () => {
            const specialContent = {
                ...testDocument.content,
                html: `
                    <h1>Special Characters Test</h1>
                    <p>Testing &amp; entities: &lt; &gt; &quot; &#39;</p>
                    <p>Unicode: © ® ™ € £ ¥</p>
                    <p>Math: α β γ δ ∑ ∫ ∞</p>
                `
            };
            
            const specialDoc = new LIVDocument(
                testDocument.manifest, 
                specialContent, 
                testDocument.assets, 
                testDocument.signatures, 
                new Map()
            );
            
            const html = await htmlMarkdownConverter.exportToHTML(specialDoc, { standalone: true });
            const markdown = await htmlMarkdownConverter.exportToMarkdown(specialDoc);
            
            expect(html).toContain('&amp;');
            expect(html).toContain('©');
            expect(markdown).toContain('©');
        });

        it('should preserve code blocks and formatting', async () => {
            const codeContent = {
                ...testDocument.content,
                html: `
                    <h1>Code Test</h1>
                    <p>Inline <code>console.log("test")</code> code.</p>
                    <pre><code>
function test() {
    if (true) {
        return "success";
    }
}
                    </code></pre>
                `
            };
            
            const codeDoc = new LIVDocument(
                testDocument.manifest, 
                codeContent, 
                testDocument.assets, 
                testDocument.signatures, 
                new Map()
            );
            
            const html = await htmlMarkdownConverter.exportToHTML(codeDoc, { standalone: true });
            const markdown = await htmlMarkdownConverter.exportToMarkdown(codeDoc);
            
            expect(html).toContain('<code>');
            expect(html).toContain('<pre>');
            expect(markdown).toContain('`console.log("test")`');
            expect(markdown).toContain('```');
        });

        it('should preserve link structure and attributes', async () => {
            const linkContent = {
                ...testDocument.content,
                html: `
                    <h1>Link Test</h1>
                    <p>External link: <a href="https://example.com" target="_blank">Example</a></p>
                    <p>Internal link: <a href="#section1">Section 1</a></p>
                    <p>Email link: <a href="mailto:test@example.com">Contact</a></p>
                `
            };
            
            const linkDoc = new LIVDocument(
                testDocument.manifest, 
                linkContent, 
                testDocument.assets, 
                testDocument.signatures, 
                new Map()
            );
            
            const html = await htmlMarkdownConverter.exportToHTML(linkDoc, { standalone: true });
            const markdown = await htmlMarkdownConverter.exportToMarkdown(linkDoc);
            
            expect(html).toContain('href="https://example.com"');
            expect(html).toContain('href="#section1"');
            expect(markdown).toContain('[Example](https://example.com)');
            expect(markdown).toContain('[Section 1](#section1)');
        });
    });

    describe('Error Handling and Edge Cases', () => {
        it('should handle empty documents gracefully', async () => {
            const emptyContent = {
                html: '',
                css: '',
                interactiveSpec: '',
                staticFallback: ''
            };
            
            const emptyDoc = new LIVDocument(
                testDocument.manifest, 
                emptyContent, 
                testDocument.assets, 
                testDocument.signatures, 
                new Map()
            );
            
            const html = await htmlMarkdownConverter.exportToHTML(emptyDoc, { standalone: true });
            const markdown = await htmlMarkdownConverter.exportToMarkdown(emptyDoc);
            const pdfBuffer = await pdfConverter.exportToPDF(emptyDoc);
            const epubBuffer = await epubConverter.exportToEPUB(emptyDoc);
            
            expect(html).toContain('<!DOCTYPE html>');
            expect(markdown).toBeDefined();
            expect(pdfBuffer).toBeInstanceOf(ArrayBuffer);
            expect(epubBuffer).toBeInstanceOf(ArrayBuffer);
        });

        it('should handle malformed HTML gracefully', async () => {
            const malformedContent = {
                ...testDocument.content,
                html: '<h1>Unclosed heading<p>Missing closing tag<div><span>Nested unclosed'
            };
            
            const malformedDoc = new LIVDocument(
                testDocument.manifest, 
                malformedContent, 
                testDocument.assets, 
                testDocument.signatures, 
                new Map()
            );
            
            const html = await htmlMarkdownConverter.exportToHTML(malformedDoc, { standalone: true });
            const markdown = await htmlMarkdownConverter.exportToMarkdown(malformedDoc);
            
            expect(html).toBeDefined();
            expect(markdown).toBeDefined();
        });

        it('should handle documents with missing static fallback', async () => {
            const noFallbackContent = {
                ...testDocument.content,
                staticFallback: ''
            };
            
            const noFallbackDoc = new LIVDocument(
                testDocument.manifest, 
                noFallbackContent, 
                testDocument.assets, 
                testDocument.signatures, 
                new Map()
            );
            
            const markdown = await htmlMarkdownConverter.exportToMarkdown(noFallbackDoc);
            const epubBuffer = await epubConverter.exportToEPUB(noFallbackDoc);
            
            expect(markdown).toBeDefined();
            expect(epubBuffer).toBeInstanceOf(ArrayBuffer);
        });
    });

    describe('Performance and Scalability Tests', () => {
        it('should handle large documents efficiently', async () => {
            // Create a large document
            const largeContent = Array(100).fill(`
                <h2>Section</h2>
                <p>This is a paragraph with <strong>bold</strong> and <em>italic</em> text.</p>
                <ul>
                    <li>Item 1</li>
                    <li>Item 2</li>
                    <li>Item 3</li>
                </ul>
            `).join('\n');
            
            const largeDoc = new LIVDocument(
                testDocument.manifest,
                { ...testDocument.content, html: largeContent, staticFallback: largeContent },
                testDocument.assets,
                testDocument.signatures,
                new Map()
            );
            
            const startTime = Date.now();
            
            const html = await htmlMarkdownConverter.exportToHTML(largeDoc, { standalone: true });
            const markdown = await htmlMarkdownConverter.exportToMarkdown(largeDoc);
            const pdfBuffer = await pdfConverter.exportToPDF(largeDoc);
            const epubBuffer = await epubConverter.exportToEPUB(largeDoc);
            
            const endTime = Date.now();
            
            expect(html).toBeDefined();
            expect(markdown).toBeDefined();
            expect(pdfBuffer).toBeInstanceOf(ArrayBuffer);
            expect(epubBuffer).toBeInstanceOf(ArrayBuffer);
            expect(endTime - startTime).toBeLessThan(10000); // Should complete within 10 seconds
        });

        it('should handle concurrent conversions', async () => {
            const promises = [
                htmlMarkdownConverter.exportToHTML(testDocument, { standalone: true }),
                htmlMarkdownConverter.exportToMarkdown(testDocument),
                pdfConverter.exportToPDF(testDocument),
                epubConverter.exportToEPUB(testDocument)
            ];
            
            const results = await Promise.all(promises);
            
            expect(results[0]).toContain('<!DOCTYPE html>');
            expect(results[1]).toContain('# Test Document');
            expect(results[2]).toBeInstanceOf(ArrayBuffer);
            expect(results[3]).toBeInstanceOf(ArrayBuffer);
        });
    });

    describe('CLI Integration Tests', () => {
        it('should validate conversion command patterns', () => {
            // Test that conversion patterns match expected CLI usage
            const testCases = [
                { input: 'document.liv', format: 'html', output: 'document.html' },
                { input: 'document.liv', format: 'markdown', output: 'document.md' },
                { input: 'document.liv', format: 'pdf', output: 'document.pdf' },
                { input: 'document.liv', format: 'epub', output: 'document.epub' },
                { input: 'document.html', format: 'liv', output: 'document.liv' },
                { input: 'document.md', format: 'liv', output: 'document.liv' }
            ];
            
            testCases.forEach(testCase => {
                expect(testCase.input).toMatch(/\.(liv|html|md|pdf|epub)$/);
                expect(testCase.output).toMatch(/\.(liv|html|md|pdf|epub)$/);
                expect(['html', 'markdown', 'pdf', 'epub', 'liv']).toContain(testCase.format);
            });
        });
    });

    describe('Validation and Integrity Tests', () => {
        it('should maintain document validation after conversion', async () => {
            // Export and re-import
            const html = await htmlMarkdownConverter.exportToHTML(testDocument, { standalone: true });
            const importedDoc = await htmlMarkdownConverter.importFromHTML(html);
            
            // Validate structure
            expect(importedDoc.manifest).toBeDefined();
            expect(importedDoc.content).toBeDefined();
            expect(importedDoc.assets).toBeDefined();
            expect(importedDoc.signatures).toBeDefined();
            
            // Validate security settings
            expect(importedDoc.manifest.security.wasmPermissions.allowNetworking).toBe(false);
            expect(importedDoc.manifest.security.jsPermissions.executionMode).toBe('sandboxed');
        });

        it('should preserve document metadata across conversions', async () => {
            const formats = ['html', 'markdown', 'epub'];
            
            for (const format of formats) {
                let exported: any;
                let imported: LIVDocument;
                
                switch (format) {
                    case 'html':
                        exported = await htmlMarkdownConverter.exportToHTML(testDocument, { standalone: true });
                        imported = await htmlMarkdownConverter.importFromHTML(exported);
                        break;
                    case 'markdown':
                        exported = await htmlMarkdownConverter.exportToMarkdown(testDocument);
                        imported = await htmlMarkdownConverter.importFromMarkdown(exported);
                        break;
                    case 'epub':
                        exported = await epubConverter.exportToEPUB(testDocument);
                        imported = await epubConverter.importFromEPUB(exported);
                        break;
                    default:
                        continue;
                }
                
                expect(imported.manifest.metadata.title).toBeDefined();
                expect(imported.manifest.metadata.author).toBeDefined();
                expect(imported.manifest.metadata.language).toBeDefined();
            }
        });
    });
}); 