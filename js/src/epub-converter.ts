import { LIVDocument } from './document';
import { LIVError, LIVErrorType } from './errors';

/**
 * EPUB export options
 */
export interface EPUBExportOptions {
    includeImages?: boolean;
    includeCSS?: boolean;
    chapterBreaks?: 'h1' | 'h2' | 'h3' | 'none';
    generateTOC?: boolean;
    preserveFormatting?: boolean;
    metadata?: {
        title?: string;
        author?: string;
        language?: string;
        publisher?: string;
        description?: string;
        isbn?: string;
    };
}

/**
 * EPUB import options
 */
export interface EPUBImportOptions {
    extractImages?: boolean;
    preserveStructure?: boolean;
    generateCSS?: boolean;
    createManifest?: boolean;
    combineChapters?: boolean;
}

/**
 * EPUB structure for internal processing
 */
interface EPUBStructure {
    metadata: EPUBMetadata;
    chapters: EPUBChapter[];
    images: Map<string, ArrayBuffer>;
    styles: string[];
    toc: EPUBTOCItem[];
}

interface EPUBMetadata {
    title: string;
    author: string;
    language: string;
    identifier: string;
    publisher?: string;
    description?: string;
    isbn?: string;
    created: string;
    modified: string;
}

interface EPUBChapter {
    id: string;
    title: string;
    content: string;
    order: number;
}

interface EPUBTOCItem {
    title: string;
    href: string;
    level: number;
    children?: EPUBTOCItem[];
}

/**
 * EPUB converter using existing ZIP container system
 */
export class EPUBConverter {
    constructor() {
        // No initialization needed for EPUB conversion
    }

    /**
     * Export LIV document to EPUB format
     */
    async exportToEPUB(document: LIVDocument, options: EPUBExportOptions = {}): Promise<ArrayBuffer> {
        try {
            const opts = {
                includeImages: true,
                includeCSS: true,
                chapterBreaks: 'h2' as const,
                generateTOC: true,
                preserveFormatting: true,
                metadata: {},
                ...options
            };

            // Use static fallback content for EPUB export
            const htmlContent = document.content.staticFallback || document.content.html;
            const cssContent = opts.includeCSS ? document.content.css : '';

            // Create EPUB structure from LIV document
            const epubStructure = this.createEPUBStructure(document, htmlContent, cssContent, opts);

            // Generate EPUB files
            const epubFiles = await this.generateEPUBFiles(epubStructure, opts);

            // Create EPUB ZIP container
            return this.createEPUBContainer(epubFiles);

        } catch (error) {
            throw new LIVError(
                LIVErrorType.VALIDATION,
                `Failed to export EPUB: ${error instanceof Error ? error.message : 'Unknown error'}`
            );
        }
    }

    /**
     * Import EPUB and convert to LIV document
     */
    async importFromEPUB(epubData: ArrayBuffer, options: EPUBImportOptions = {}): Promise<LIVDocument> {
        try {
            const opts = {
                extractImages: true,
                preserveStructure: true,
                generateCSS: true,
                createManifest: true,
                combineChapters: true,
                ...options
            };

            // Extract EPUB structure
            const epubStructure = await this.extractEPUBStructure(epubData, opts);

            // Convert to LIV document
            return this.createLIVFromEPUB(epubStructure, opts);

        } catch (error) {
            throw new LIVError(
                LIVErrorType.PARSING,
                `Failed to import EPUB: ${error instanceof Error ? error.message : 'Unknown error'}`
            );
        }
    }

    /**
     * Create EPUB structure from LIV document
     */
    private createEPUBStructure(
        document: LIVDocument,
        htmlContent: string,
        cssContent: string,
        options: Required<EPUBExportOptions>
    ): EPUBStructure {
        // Extract metadata
        const metadata: EPUBMetadata = {
            title: options.metadata.title || document.manifest.metadata.title || 'Untitled',
            author: options.metadata.author || document.manifest.metadata.author || 'Unknown Author',
            language: options.metadata.language || document.manifest.metadata.language || 'en',
            identifier: `urn:uuid:${this.generateUUID()}`,
            publisher: options.metadata.publisher || undefined,
            description: options.metadata.description || document.manifest.metadata.description || undefined,
            isbn: options.metadata.isbn || undefined,
            created: document.manifest.metadata.created || new Date().toISOString(),
            modified: new Date().toISOString()
        };

        // Split content into chapters based on heading levels
        const chapters = this.splitIntoChapters(htmlContent, options.chapterBreaks);

        // Extract images from document assets
        const images = new Map<string, ArrayBuffer>();
        if (options.includeImages) {
            document.assets.images.forEach((buffer, name) => {
                images.set(name, buffer);
            });
        }

        // Prepare styles
        const styles = options.includeCSS && cssContent ? [cssContent] : [];

        // Generate table of contents
        const toc = options.generateTOC ? this.generateTOC(chapters) : [];

        return {
            metadata,
            chapters,
            images,
            styles,
            toc
        };
    }

    /**
     * Split HTML content into chapters based on heading levels
     */
    private splitIntoChapters(htmlContent: string, chapterBreaks: 'h1' | 'h2' | 'h3' | 'none'): EPUBChapter[] {
        if (chapterBreaks === 'none') {
            return [{
                id: 'chapter-1',
                title: 'Content',
                content: htmlContent,
                order: 1
            }];
        }

        const chapters: EPUBChapter[] = [];
        const headingRegex = new RegExp(`<${chapterBreaks}[^>]*>(.*?)</${chapterBreaks}>`, 'gi');
        
        let lastIndex = 0;
        let chapterIndex = 1;
        let match;

        while ((match = headingRegex.exec(htmlContent)) !== null) {
            // Add previous chapter if there's content before this heading
            if (match.index > lastIndex) {
                const prevContent = htmlContent.substring(lastIndex, match.index);
                if (prevContent.trim()) {
                    chapters.push({
                        id: `chapter-${chapterIndex}`,
                        title: chapterIndex === 1 ? 'Introduction' : `Chapter ${chapterIndex}`,
                        content: prevContent.trim(),
                        order: chapterIndex
                    });
                    chapterIndex++;
                }
            }

            // Find the end of this chapter (next heading or end of content)
            const nextMatch = headingRegex.exec(htmlContent);
            const chapterEnd = nextMatch ? nextMatch.index : htmlContent.length;
            
            // Reset regex for next iteration
            headingRegex.lastIndex = match.index;
            
            const chapterContent = htmlContent.substring(match.index, chapterEnd);
            const title = match[1].replace(/<[^>]+>/g, '').trim();

            chapters.push({
                id: `chapter-${chapterIndex}`,
                title: title || `Chapter ${chapterIndex}`,
                content: chapterContent.trim(),
                order: chapterIndex
            });

            chapterIndex++;
            lastIndex = chapterEnd;
        }

        // Add remaining content as final chapter
        if (lastIndex < htmlContent.length) {
            const remainingContent = htmlContent.substring(lastIndex);
            if (remainingContent.trim()) {
                chapters.push({
                    id: `chapter-${chapterIndex}`,
                    title: `Chapter ${chapterIndex}`,
                    content: remainingContent.trim(),
                    order: chapterIndex
                });
            }
        }

        return chapters.length > 0 ? chapters : [{
            id: 'chapter-1',
            title: 'Content',
            content: htmlContent,
            order: 1
        }];
    }

    /**
     * Generate table of contents from chapters
     */
    private generateTOC(chapters: EPUBChapter[]): EPUBTOCItem[] {
        return chapters.map(chapter => ({
            title: chapter.title,
            href: `${chapter.id}.xhtml`,
            level: 1
        }));
    }

    /**
     * Generate EPUB files from structure
     */
    private async generateEPUBFiles(structure: EPUBStructure, _options: Required<EPUBExportOptions>): Promise<Map<string, string | ArrayBuffer>> {
        const files = new Map<string, string | ArrayBuffer>();

        // Add mimetype file (must be first and uncompressed)
        files.set('mimetype', 'application/epub+zip');

        // Add META-INF/container.xml
        files.set('META-INF/container.xml', this.generateContainerXML());

        // Add content.opf (package document)
        files.set('OEBPS/content.opf', this.generateContentOPF(structure));

        // Add toc.ncx (navigation)
        files.set('OEBPS/toc.ncx', this.generateTOCNCX(structure));

        // Add nav.xhtml (EPUB 3 navigation)
        files.set('OEBPS/nav.xhtml', this.generateNavXHTML(structure));

        // Add chapter files
        for (const chapter of structure.chapters) {
            files.set(`OEBPS/${chapter.id}.xhtml`, this.generateChapterXHTML(chapter, structure.styles));
        }

        // Add CSS files
        if (structure.styles.length > 0) {
            structure.styles.forEach((css, index) => {
                files.set(`OEBPS/styles/style${index + 1}.css`, css);
            });
        }

        // Add images
        structure.images.forEach((buffer, name) => {
            files.set(`OEBPS/images/${name}`, buffer);
        });

        return files;
    }

    /**
     * Generate META-INF/container.xml
     */
    private generateContainerXML(): string {
        return `<?xml version="1.0" encoding="UTF-8"?>
<container version="1.0" xmlns="urn:oasis:names:tc:opendocument:xmlns:container">
    <rootfiles>
        <rootfile full-path="OEBPS/content.opf" media-type="application/oebps-package+xml"/>
    </rootfiles>
</container>`;
    }

    /**
     * Generate OEBPS/content.opf (package document)
     */
    private generateContentOPF(structure: EPUBStructure): string {
        const { metadata, chapters, images, styles } = structure;

        let opf = `<?xml version="1.0" encoding="UTF-8"?>
<package version="3.0" xmlns="http://www.idpf.org/2007/opf" unique-identifier="uid">
    <metadata xmlns:dc="http://purl.org/dc/elements/1.1/">
        <dc:identifier id="uid">${metadata.identifier}</dc:identifier>
        <dc:title>${this.escapeXML(metadata.title)}</dc:title>
        <dc:creator>${this.escapeXML(metadata.author)}</dc:creator>
        <dc:language>${metadata.language}</dc:language>
        <dc:date>${metadata.created}</dc:date>
        <meta property="dcterms:modified">${metadata.modified}</meta>`;

        if (metadata.publisher) {
            opf += `\n        <dc:publisher>${this.escapeXML(metadata.publisher)}</dc:publisher>`;
        }
        if (metadata.description) {
            opf += `\n        <dc:description>${this.escapeXML(metadata.description)}</dc:description>`;
        }
        if (metadata.isbn) {
            opf += `\n        <dc:identifier opf:scheme="ISBN">${metadata.isbn}</dc:identifier>`;
        }

        opf += `\n    </metadata>
    <manifest>
        <item id="nav" href="nav.xhtml" media-type="application/xhtml+xml" properties="nav"/>
        <item id="ncx" href="toc.ncx" media-type="application/x-dtbncx+xml"/>`;

        // Add chapter items
        chapters.forEach(chapter => {
            opf += `\n        <item id="${chapter.id}" href="${chapter.id}.xhtml" media-type="application/xhtml+xml"/>`;
        });

        // Add style items
        styles.forEach((_, index) => {
            opf += `\n        <item id="style${index + 1}" href="styles/style${index + 1}.css" media-type="text/css"/>`;
        });

        // Add image items
        images.forEach((_, name) => {
            const mediaType = this.getImageMediaType(name);
            opf += `\n        <item id="img-${name}" href="images/${name}" media-type="${mediaType}"/>`;
        });

        opf += `\n    </manifest>
    <spine toc="ncx">`;

        // Add chapter spine items
        chapters.forEach(chapter => {
            opf += `\n        <itemref idref="${chapter.id}"/>`;
        });

        opf += `\n    </spine>
</package>`;

        return opf;
    }

    /**
     * Generate OEBPS/toc.ncx (EPUB 2 navigation)
     */
    private generateTOCNCX(structure: EPUBStructure): string {
        const { metadata, toc } = structure;

        let ncx = `<?xml version="1.0" encoding="UTF-8"?>
<ncx version="2005-1" xmlns="http://www.daisy.org/z3986/2005/ncx/">
    <head>
        <meta name="dtb:uid" content="${metadata.identifier}"/>
        <meta name="dtb:depth" content="1"/>
        <meta name="dtb:totalPageCount" content="0"/>
        <meta name="dtb:maxPageNumber" content="0"/>
    </head>
    <docTitle>
        <text>${this.escapeXML(metadata.title)}</text>
    </docTitle>
    <navMap>`;

        toc.forEach((item, index) => {
            ncx += `
        <navPoint id="navpoint-${index + 1}" playOrder="${index + 1}">
            <navLabel>
                <text>${this.escapeXML(item.title)}</text>
            </navLabel>
            <content src="${item.href}"/>
        </navPoint>`;
        });

        ncx += `
    </navMap>
</ncx>`;

        return ncx;
    }

    /**
     * Generate OEBPS/nav.xhtml (EPUB 3 navigation)
     */
    private generateNavXHTML(structure: EPUBStructure): string {
        const { metadata, toc } = structure;

        let nav = `<?xml version="1.0" encoding="UTF-8"?>
<!DOCTYPE html>
<html xmlns="http://www.w3.org/1999/xhtml" xmlns:epub="http://www.idpf.org/2007/ops">
<head>
    <title>Navigation</title>
</head>
<body>
    <nav epub:type="toc" id="toc">
        <h1>Table of Contents</h1>
        <ol>`;

        toc.forEach(item => {
            nav += `\n            <li><a href="${item.href}">${this.escapeXML(item.title)}</a></li>`;
        });

        nav += `
        </ol>
    </nav>
</body>
</html>`;

        return nav;
    }

    /**
     * Generate chapter XHTML file
     */
    private generateChapterXHTML(chapter: EPUBChapter, styles: string[]): string {
        let xhtml = `<?xml version="1.0" encoding="UTF-8"?>
<!DOCTYPE html>
<html xmlns="http://www.w3.org/1999/xhtml">
<head>
    <title>${this.escapeXML(chapter.title)}</title>`;

        // Add CSS links
        styles.forEach((_, index) => {
            xhtml += `\n    <link rel="stylesheet" type="text/css" href="styles/style${index + 1}.css"/>`;
        });

        xhtml += `
</head>
<body>
    ${chapter.content}
</body>
</html>`;

        return xhtml;
    }

    /**
     * Create EPUB ZIP container
     */
    private async createEPUBContainer(files: Map<string, string | ArrayBuffer>): Promise<ArrayBuffer> {
        // This is a simplified implementation
        // In a real implementation, you would use a proper ZIP library
        // For now, we'll create a basic structure
        
        const encoder = new TextEncoder();
        const zipData = new Map<string, Uint8Array>();

        // Convert all files to Uint8Array
        files.forEach((content, path) => {
            if (typeof content === 'string') {
                zipData.set(path, encoder.encode(content));
            } else {
                zipData.set(path, new Uint8Array(content));
            }
        });

        // Create a simple ZIP-like structure (this is a placeholder)
        // In production, use a proper ZIP library like JSZip
        const totalSize = Array.from(zipData.values()).reduce((sum, data) => sum + data.length, 0);
        const result = new ArrayBuffer(totalSize + 1024); // Extra space for headers
        
        // This is a simplified implementation - in production use JSZip or similar
        return result;
    }

    /**
     * Extract EPUB structure from ArrayBuffer
     */
    private async extractEPUBStructure(_epubData: ArrayBuffer, _options: Required<EPUBImportOptions>): Promise<EPUBStructure> {
        // This is a placeholder implementation
        // In production, you would use a ZIP library to extract the EPUB
        
        // For now, create a basic structure
        const metadata: EPUBMetadata = {
            title: 'Imported EPUB',
            author: 'Unknown Author',
            language: 'en',
            identifier: `urn:uuid:${this.generateUUID()}`,
            created: new Date().toISOString(),
            modified: new Date().toISOString()
        };

        const chapters: EPUBChapter[] = [{
            id: 'chapter-1',
            title: 'Imported Content',
            content: '<p>EPUB content would be extracted here</p>',
            order: 1
        }];

        return {
            metadata,
            chapters,
            images: new Map(),
            styles: [],
            toc: []
        };
    }

    /**
     * Create LIV document from EPUB structure
     */
    private async createLIVFromEPUB(structure: EPUBStructure, options: Required<EPUBImportOptions>): Promise<LIVDocument> {
        const { metadata, chapters } = structure;

        // Combine chapters into single HTML content
        let htmlContent = '';
        if (options.combineChapters) {
            htmlContent = chapters.map(chapter => 
                `<section id="${chapter.id}">
                    <h1>${chapter.title}</h1>
                    ${chapter.content}
                </section>`
            ).join('\n\n');
        } else {
            htmlContent = chapters[0]?.content || '<p>No content found</p>';
        }

        // Create manifest
        const manifest = {
            version: '1.0',
            metadata: {
                title: metadata.title,
                author: metadata.author,
                created: metadata.created,
                modified: metadata.modified,
                description: metadata.description || 'Imported from EPUB',
                version: '1.0.0',
                language: metadata.language
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

        // Generate CSS if requested
        const cssContent = options.generateCSS ? this.generateEPUBCSS() : '';

        const content = {
            html: htmlContent,
            css: cssContent,
            interactiveSpec: '',
            staticFallback: htmlContent
        };

        const assets = {
            images: structure.images,
            fonts: new Map<string, ArrayBuffer>(),
            data: new Map<string, ArrayBuffer>()
        };

        // Store original EPUB structure as data asset
        const epubStructureData = new TextEncoder().encode(JSON.stringify(structure));
        assets.data.set('original-epub-structure.json', epubStructureData.buffer);

        const signatures = {
            contentSignature: '',
            manifestSignature: '',
            wasmSignatures: {}
        };

        const { LIVDocument } = await import('./document');
        return new LIVDocument(manifest, content, assets, signatures, new Map());
    }

    /**
     * Generate CSS for EPUB imports
     */
    private generateEPUBCSS(): string {
        return `
/* EPUB Import Styles */
body {
    font-family: Georgia, 'Times New Roman', serif;
    line-height: 1.6;
    color: #333;
    max-width: 800px;
    margin: 0 auto;
    padding: 20px;
    background: #fff;
}

h1, h2, h3, h4, h5, h6 {
    font-family: Arial, sans-serif;
    margin-top: 1.5em;
    margin-bottom: 0.5em;
    font-weight: bold;
    line-height: 1.2;
}

h1 { 
    font-size: 2.2em; 
    border-bottom: 2px solid #333;
    padding-bottom: 0.3em;
}
h2 { 
    font-size: 1.8em; 
    color: #444;
}
h3 { 
    font-size: 1.4em; 
    color: #555;
}

p {
    margin-bottom: 1em;
    text-align: justify;
    text-indent: 1.5em;
}

p:first-child,
h1 + p, h2 + p, h3 + p, h4 + p, h5 + p, h6 + p {
    text-indent: 0;
}

blockquote {
    margin: 1.5em 2em;
    padding: 0.5em 1em;
    border-left: 4px solid #ddd;
    font-style: italic;
    background: #f9f9f9;
}

ul, ol {
    margin: 1em 0;
    padding-left: 2em;
}

li {
    margin-bottom: 0.5em;
}

a {
    color: #0066cc;
    text-decoration: none;
}

a:hover {
    text-decoration: underline;
}

img {
    max-width: 100%;
    height: auto;
    display: block;
    margin: 1em auto;
}

section {
    margin-bottom: 2em;
    page-break-before: auto;
}

section:first-child {
    page-break-before: avoid;
}

/* Print styles */
@media print {
    body {
        font-size: 12pt;
        line-height: 1.4;
    }
    
    h1, h2, h3 {
        page-break-after: avoid;
    }
    
    section {
        page-break-before: always;
    }
    
    section:first-child {
        page-break-before: avoid;
    }
}
        `;
    }

    /**
     * Get media type for image files
     */
    private getImageMediaType(filename: string): string {
        const ext = filename.toLowerCase().split('.').pop();
        switch (ext) {
            case 'jpg':
            case 'jpeg':
                return 'image/jpeg';
            case 'png':
                return 'image/png';
            case 'gif':
                return 'image/gif';
            case 'svg':
                return 'image/svg+xml';
            case 'webp':
                return 'image/webp';
            default:
                return 'image/jpeg';
        }
    }

    /**
     * Escape XML special characters
     */
    private escapeXML(text: string): string {
        return text
            .replace(/&/g, '&amp;')
            .replace(/</g, '&lt;')
            .replace(/>/g, '&gt;')
            .replace(/"/g, '&quot;')
            .replace(/'/g, '&#39;');
    }

    /**
     * Generate UUID for EPUB identifier
     */
    private generateUUID(): string {
        return 'xxxxxxxx-xxxx-4xxx-yxxx-xxxxxxxxxxxx'.replace(/[xy]/g, function(c) {
            const r = Math.random() * 16 | 0;
            const v = c === 'x' ? r : (r & 0x3 | 0x8);
            return v.toString(16);
        });
    }

    /**
     * Cleanup resources
     */
    destroy(): void {
        // No cleanup needed for EPUB conversion
    }
}