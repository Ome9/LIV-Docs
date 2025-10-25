import { LIVDocument } from './document';
import { LIVError, LIVErrorType } from './errors';

/**
 * HTML export options
 */
export interface HTMLExportOptions {
    includeCSS?: boolean;
    includeInteractive?: boolean;
    standalone?: boolean;
    minify?: boolean;
    preserveStructure?: boolean;
}

/**
 * Markdown export options
 */
export interface MarkdownExportOptions {
    includeImages?: boolean;
    includeLinks?: boolean;
    preserveFormatting?: boolean;
    flavor?: 'github' | 'commonmark' | 'standard';
}

/**
 * HTML import options
 */
export interface HTMLImportOptions {
    extractCSS?: boolean;
    preserveStructure?: boolean;
    sanitize?: boolean;
    createManifest?: boolean;
}

/**
 * Markdown import options
 */
export interface MarkdownImportOptions {
    flavor?: 'github' | 'commonmark' | 'standard';
    preserveFormatting?: boolean;
    generateCSS?: boolean;
    createManifest?: boolean;
}

/**
 * HTML and Markdown converter using existing infrastructure
 */
export class HTMLMarkdownConverter {
    constructor() {
        // No initialization needed for HTML/Markdown conversion
    }

    /**
     * Export LIV document to HTML
     */
    async exportToHTML(document: LIVDocument, options: HTMLExportOptions = {}): Promise<string> {
        try {
            const opts = {
                includeCSS: true,
                includeInteractive: false,
                standalone: true,
                minify: false,
                preserveStructure: true,
                ...options
            };

            // Use static fallback mode for HTML export if interactive is disabled
            const htmlContent = opts.includeInteractive ? 
                document.content.html : 
                (document.content.staticFallback || document.content.html);

            const cssContent = opts.includeCSS ? document.content.css : '';

            if (opts.standalone) {
                return this.generateStandaloneHTML(htmlContent, cssContent, document, opts);
            } else {
                return this.processHTMLContent(htmlContent, opts);
            }
        } catch (error) {
            throw new LIVError(
                LIVErrorType.VALIDATION,
                `Failed to export HTML: ${error instanceof Error ? error.message : 'Unknown error'}`
            );
        }
    }

    /**
     * Export LIV document to Markdown
     */
    async exportToMarkdown(document: LIVDocument, options: MarkdownExportOptions = {}): Promise<string> {
        try {
            const opts = {
                includeImages: true,
                includeLinks: true,
                preserveFormatting: true,
                flavor: 'github' as const,
                ...options
            };

            // Use static fallback content for Markdown export
            const htmlContent = document.content.staticFallback || document.content.html;
            
            return this.convertHTMLToMarkdown(htmlContent, opts);
        } catch (error) {
            throw new LIVError(
                LIVErrorType.VALIDATION,
                `Failed to export Markdown: ${error instanceof Error ? error.message : 'Unknown error'}`
            );
        }
    }

    /**
     * Import HTML and convert to LIV document
     */
    async importFromHTML(htmlContent: string, options: HTMLImportOptions = {}): Promise<LIVDocument> {
        try {
            const opts = {
                extractCSS: true,
                preserveStructure: true,
                sanitize: true,
                createManifest: true,
                ...options
            };

            // Parse HTML content
            const parsedHTML = this.parseHTMLContent(htmlContent, opts);
            
            // Create LIV document from parsed HTML
            return this.createLIVFromHTML(parsedHTML, opts);
        } catch (error) {
            throw new LIVError(
                LIVErrorType.PARSING,
                `Failed to import HTML: ${error instanceof Error ? error.message : 'Unknown error'}`
            );
        }
    }

    /**
     * Import Markdown and convert to LIV document
     */
    async importFromMarkdown(markdownContent: string, options: MarkdownImportOptions = {}): Promise<LIVDocument> {
        try {
            const opts = {
                flavor: 'github' as const,
                preserveFormatting: true,
                generateCSS: true,
                createManifest: true,
                ...options
            };

            // Convert Markdown to HTML
            const htmlContent = this.convertMarkdownToHTML(markdownContent, opts);
            
            // Create LIV document from converted HTML
            return this.createLIVFromMarkdown(htmlContent, markdownContent, opts);
        } catch (error) {
            throw new LIVError(
                LIVErrorType.PARSING,
                `Failed to import Markdown: ${error instanceof Error ? error.message : 'Unknown error'}`
            );
        }
    }

    /**
     * Generate standalone HTML document
     */
    private generateStandaloneHTML(
        htmlContent: string, 
        cssContent: string, 
        document: LIVDocument, 
        options: Required<HTMLExportOptions>
    ): string {
        const title = document.manifest.metadata.title || 'LIV Document';
        const author = document.manifest.metadata.author || '';
        const description = document.manifest.metadata.description || '';

        let html = `<!DOCTYPE html>
<html lang="${document.manifest.metadata.language || 'en'}">
<head>
    <meta charset="UTF-8">
    <meta name="viewport" content="width=device-width, initial-scale=1.0">
    <title>${this.escapeHTML(title)}</title>`;

        if (author) {
            html += `\n    <meta name="author" content="${this.escapeHTML(author)}">`;
        }
        
        if (description) {
            html += `\n    <meta name="description" content="${this.escapeHTML(description)}">`;
        }

        if (cssContent && options.includeCSS) {
            html += `\n    <style>\n${cssContent}\n    </style>`;
        }

        html += `\n</head>
<body>
${htmlContent}
</body>
</html>`;

        return options.minify ? this.minifyHTML(html) : html;
    }

    /**
     * Process HTML content based on options
     */
    private processHTMLContent(htmlContent: string, options: Required<HTMLExportOptions>): string {
        let processedHTML = htmlContent;

        if (!options.preserveStructure) {
            // Extract body content only
            const bodyMatch = htmlContent.match(/<body[^>]*>([\s\S]*?)<\/body>/i);
            if (bodyMatch) {
                processedHTML = bodyMatch[1];
            }
        }

        return options.minify ? this.minifyHTML(processedHTML) : processedHTML;
    }

    /**
     * Convert HTML to Markdown
     */
    private convertHTMLToMarkdown(htmlContent: string, options: Required<MarkdownExportOptions>): string {
        // Simple HTML to Markdown conversion
        let markdown = htmlContent;

        // Convert headings
        markdown = markdown.replace(/<h1[^>]*>(.*?)<\/h1>/gi, '# $1\n\n');
        markdown = markdown.replace(/<h2[^>]*>(.*?)<\/h2>/gi, '## $1\n\n');
        markdown = markdown.replace(/<h3[^>]*>(.*?)<\/h3>/gi, '### $1\n\n');
        markdown = markdown.replace(/<h4[^>]*>(.*?)<\/h4>/gi, '#### $1\n\n');
        markdown = markdown.replace(/<h5[^>]*>(.*?)<\/h5>/gi, '##### $1\n\n');
        markdown = markdown.replace(/<h6[^>]*>(.*?)<\/h6>/gi, '###### $1\n\n');

        // Convert paragraphs
        markdown = markdown.replace(/<p[^>]*>(.*?)<\/p>/gi, '$1\n\n');

        // Convert emphasis
        if (options.preserveFormatting) {
            markdown = markdown.replace(/<strong[^>]*>(.*?)<\/strong>/gi, '**$1**');
            markdown = markdown.replace(/<b[^>]*>(.*?)<\/b>/gi, '**$1**');
            markdown = markdown.replace(/<em[^>]*>(.*?)<\/em>/gi, '*$1*');
            markdown = markdown.replace(/<i[^>]*>(.*?)<\/i>/gi, '*$1*');
        }

        // Convert links
        if (options.includeLinks) {
            markdown = markdown.replace(/<a[^>]*href="([^"]*)"[^>]*>(.*?)<\/a>/gi, '[$2]($1)');
        }

        // Convert images
        if (options.includeImages) {
            markdown = markdown.replace(/<img[^>]*src="([^"]*)"[^>]*alt="([^"]*)"[^>]*>/gi, '![$2]($1)');
            markdown = markdown.replace(/<img[^>]*src="([^"]*)"[^>]*>/gi, '![]($1)');
        }

        // Convert lists
        markdown = markdown.replace(/<ul[^>]*>([\s\S]*?)<\/ul>/gi, (_, content) => {
            return content.replace(/<li[^>]*>(.*?)<\/li>/gi, '- $1\n') + '\n';
        });

        markdown = markdown.replace(/<ol[^>]*>([\s\S]*?)<\/ol>/gi, (_, content) => {
            let counter = 1;
            return content.replace(/<li[^>]*>(.*?)<\/li>/gi, () => `${counter++}. $1\n`) + '\n';
        });

        // Convert code blocks
        markdown = markdown.replace(/<pre[^>]*><code[^>]*>([\s\S]*?)<\/code><\/pre>/gi, '```\n$1\n```\n\n');
        markdown = markdown.replace(/<code[^>]*>(.*?)<\/code>/gi, '`$1`');

        // Convert blockquotes
        markdown = markdown.replace(/<blockquote[^>]*>([\s\S]*?)<\/blockquote>/gi, (_, content) => {
            return content.split('\n').map((line: string) => `> ${line}`).join('\n') + '\n\n';
        });

        // Convert horizontal rules
        markdown = markdown.replace(/<hr[^>]*>/gi, '---\n\n');

        // Clean up HTML tags
        markdown = markdown.replace(/<[^>]+>/g, '');
        
        // Clean up extra whitespace
        markdown = markdown.replace(/\n{3,}/g, '\n\n');
        markdown = markdown.trim();

        return markdown;
    }

    /**
     * Parse HTML content
     */
    private parseHTMLContent(htmlContent: string, options: Required<HTMLImportOptions>): ParsedHTML {
        const parsed: ParsedHTML = {
            html: htmlContent,
            css: '',
            title: '',
            metadata: {}
        };

        // Extract title
        const titleMatch = htmlContent.match(/<title[^>]*>(.*?)<\/title>/i);
        if (titleMatch) {
            parsed.title = titleMatch[1];
        }

        // Extract CSS if requested
        if (options.extractCSS) {
            const styleMatches = htmlContent.match(/<style[^>]*>([\s\S]*?)<\/style>/gi);
            if (styleMatches) {
                parsed.css = styleMatches.map(match => {
                    const cssMatch = match.match(/<style[^>]*>([\s\S]*?)<\/style>/i);
                    return cssMatch ? cssMatch[1] : '';
                }).join('\n');
            }

            // Extract external CSS links
            const linkMatches = htmlContent.match(/<link[^>]*rel="stylesheet"[^>]*>/gi);
            if (linkMatches) {
                // Note: External CSS would need to be fetched separately
                parsed.metadata.externalCSS = linkMatches;
            }
        }

        // Extract metadata
        const metaMatches = htmlContent.match(/<meta[^>]*>/gi);
        if (metaMatches) {
            metaMatches.forEach(meta => {
                const nameMatch = meta.match(/name="([^"]*)"[^>]*content="([^"]*)"/i);
                const propertyMatch = meta.match(/property="([^"]*)"[^>]*content="([^"]*)"/i);
                
                if (nameMatch) {
                    parsed.metadata[nameMatch[1]] = nameMatch[2];
                } else if (propertyMatch) {
                    parsed.metadata[propertyMatch[1]] = propertyMatch[2];
                }
            });
        }

        // Sanitize HTML if requested
        if (options.sanitize) {
            parsed.html = this.sanitizeHTML(parsed.html);
        }

        return parsed;
    }

    /**
     * Convert Markdown to HTML
     */
    private convertMarkdownToHTML(markdownContent: string, options: Required<MarkdownImportOptions>): string {
        let html = markdownContent;

        // Convert headings
        html = html.replace(/^# (.*$)/gim, '<h1>$1</h1>');
        html = html.replace(/^## (.*$)/gim, '<h2>$1</h2>');
        html = html.replace(/^### (.*$)/gim, '<h3>$1</h3>');
        html = html.replace(/^#### (.*$)/gim, '<h4>$1</h4>');
        html = html.replace(/^##### (.*$)/gim, '<h5>$1</h5>');
        html = html.replace(/^###### (.*$)/gim, '<h6>$1</h6>');

        // Convert emphasis
        if (options.preserveFormatting) {
            html = html.replace(/\*\*(.*?)\*\*/g, '<strong>$1</strong>');
            html = html.replace(/\*(.*?)\*/g, '<em>$1</em>');
            html = html.replace(/__(.*?)__/g, '<strong>$1</strong>');
            html = html.replace(/_(.*?)_/g, '<em>$1</em>');
        }

        // Convert links
        html = html.replace(/\[([^\]]*)\]\(([^)]*)\)/g, '<a href="$2">$1</a>');

        // Convert images
        html = html.replace(/!\[([^\]]*)\]\(([^)]*)\)/g, '<img src="$2" alt="$1">');

        // Convert code blocks
        html = html.replace(/```([\s\S]*?)```/g, '<pre><code>$1</code></pre>');
        html = html.replace(/`([^`]*)`/g, '<code>$1</code>');

        // Convert lists
        html = html.replace(/^\* (.*)$/gm, '<li>$1</li>');
        html = html.replace(/^- (.*)$/gm, '<li>$1</li>');
        html = html.replace(/^(\d+)\. (.*)$/gm, '<li>$2</li>');

        // Wrap consecutive list items
        html = html.replace(/(<li>.*<\/li>\s*)+/g, (match) => {
            if (markdownContent.match(/^\d+\./m)) {
                return `<ol>${match}</ol>`;
            } else {
                return `<ul>${match}</ul>`;
            }
        });

        // Convert blockquotes
        html = html.replace(/^> (.*)$/gm, '<blockquote>$1</blockquote>');

        // Convert horizontal rules
        html = html.replace(/^---$/gm, '<hr>');

        // Convert paragraphs
        html = html.replace(/^(?!<[h|u|o|l|b|p])(.*$)/gm, '<p>$1</p>');

        // Clean up extra paragraph tags
        html = html.replace(/<p><\/p>/g, '');
        html = html.replace(/<p>(<[^>]+>)<\/p>/g, '$1');

        return html;
    }

    /**
     * Create LIV document from parsed HTML
     */
    private async createLIVFromHTML(parsedHTML: ParsedHTML, _options: Required<HTMLImportOptions>): Promise<LIVDocument> {
        const manifest = {
            version: '1.0',
            metadata: {
                title: parsedHTML.title || 'Imported HTML Document',
                author: parsedHTML.metadata.author || 'HTML Converter',
                created: new Date().toISOString(),
                modified: new Date().toISOString(),
                description: parsedHTML.metadata.description || 'Imported from HTML',
                version: '1.0.0',
                language: parsedHTML.metadata.language || 'en'
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
            html: parsedHTML.html,
            css: parsedHTML.css || this.generateDefaultCSS(),
            interactiveSpec: '',
            staticFallback: this.stripInteractiveElements(parsedHTML.html)
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

        const { LIVDocument } = await import('./document');
        return new LIVDocument(manifest, content, assets, signatures, new Map());
    }

    /**
     * Create LIV document from Markdown
     */
    private async createLIVFromMarkdown(
        htmlContent: string, 
        originalMarkdown: string, 
        _options: Required<MarkdownImportOptions>
    ): Promise<LIVDocument> {
        // Extract title from first heading
        const titleMatch = originalMarkdown.match(/^# (.*)$/m);
        const title = titleMatch ? titleMatch[1] : 'Imported Markdown Document';

        const manifest = {
            version: '1.0',
            metadata: {
                title,
                author: 'Markdown Converter',
                created: new Date().toISOString(),
                modified: new Date().toISOString(),
                description: 'Imported from Markdown',
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

        const cssContent = _options.generateCSS ? this.generateMarkdownCSS() : '';

        const content = {
            html: htmlContent,
            css: cssContent,
            interactiveSpec: '',
            staticFallback: htmlContent
        };

        const assets = {
            images: new Map<string, ArrayBuffer>(),
            fonts: new Map<string, ArrayBuffer>(),
            data: new Map<string, ArrayBuffer>()
        };

        // Store original markdown as data asset
        const markdownBuffer = new TextEncoder().encode(originalMarkdown);
        assets.data.set('original.md', markdownBuffer.buffer);

        const signatures = {
            contentSignature: '',
            manifestSignature: '',
            wasmSignatures: {}
        };

        const { LIVDocument } = await import('./document');
        return new LIVDocument(manifest, content, assets, signatures, new Map());
    }

    /**
     * Generate default CSS for HTML imports
     */
    private generateDefaultCSS(): string {
        return `
/* Default HTML Import Styles */
body {
    font-family: -apple-system, BlinkMacSystemFont, 'Segoe UI', Roboto, sans-serif;
    line-height: 1.6;
    color: #333;
    max-width: 800px;
    margin: 0 auto;
    padding: 20px;
}

h1, h2, h3, h4, h5, h6 {
    margin-top: 0;
    margin-bottom: 16px;
    font-weight: 600;
    line-height: 1.25;
}

h1 { font-size: 2em; }
h2 { font-size: 1.5em; }
h3 { font-size: 1.25em; }

p {
    margin-bottom: 16px;
}

a {
    color: #0366d6;
    text-decoration: none;
}

a:hover {
    text-decoration: underline;
}

img {
    max-width: 100%;
    height: auto;
}

code {
    background-color: #f6f8fa;
    border-radius: 3px;
    font-size: 85%;
    margin: 0;
    padding: 0.2em 0.4em;
}

pre {
    background-color: #f6f8fa;
    border-radius: 6px;
    font-size: 85%;
    line-height: 1.45;
    overflow: auto;
    padding: 16px;
}

blockquote {
    border-left: 4px solid #dfe2e5;
    margin: 0;
    padding: 0 16px;
    color: #6a737d;
}

ul, ol {
    margin-bottom: 16px;
    padding-left: 2em;
}

li {
    margin-bottom: 0.25em;
}

hr {
    border: none;
    border-top: 1px solid #e1e4e8;
    margin: 24px 0;
}
        `;
    }

    /**
     * Generate CSS for Markdown imports
     */
    private generateMarkdownCSS(): string {
        return `
/* Markdown Import Styles */
body {
    font-family: -apple-system, BlinkMacSystemFont, 'Segoe UI', Roboto, sans-serif;
    line-height: 1.6;
    color: #24292e;
    max-width: 980px;
    margin: 0 auto;
    padding: 45px;
}

h1, h2, h3, h4, h5, h6 {
    margin-top: 24px;
    margin-bottom: 16px;
    font-weight: 600;
    line-height: 1.25;
}

h1 {
    font-size: 2em;
    border-bottom: 1px solid #eaecef;
    padding-bottom: 0.3em;
}

h2 {
    font-size: 1.5em;
    border-bottom: 1px solid #eaecef;
    padding-bottom: 0.3em;
}

h3 { font-size: 1.25em; }
h4 { font-size: 1em; }
h5 { font-size: 0.875em; }
h6 { font-size: 0.85em; color: #6a737d; }

p {
    margin-top: 0;
    margin-bottom: 16px;
}

a {
    color: #0366d6;
    text-decoration: none;
}

a:hover {
    text-decoration: underline;
}

strong {
    font-weight: 600;
}

em {
    font-style: italic;
}

code {
    background-color: rgba(27,31,35,0.05);
    border-radius: 3px;
    font-size: 85%;
    margin: 0;
    padding: 0.2em 0.4em;
    font-family: 'SFMono-Regular', Consolas, 'Liberation Mono', Menlo, monospace;
}

pre {
    background-color: #f6f8fa;
    border-radius: 6px;
    font-size: 85%;
    line-height: 1.45;
    overflow: auto;
    padding: 16px;
}

pre code {
    background-color: transparent;
    border: 0;
    display: inline;
    line-height: inherit;
    margin: 0;
    max-width: auto;
    overflow: visible;
    padding: 0;
    word-wrap: normal;
}

blockquote {
    border-left: 0.25em solid #dfe2e5;
    color: #6a737d;
    margin: 0;
    padding: 0 1em;
}

ul, ol {
    margin-top: 0;
    margin-bottom: 16px;
    padding-left: 2em;
}

li {
    word-wrap: break-all;
}

li > p {
    margin-top: 16px;
}

li + li {
    margin-top: 0.25em;
}

hr {
    background-color: #e1e4e8;
    border: 0;
    height: 0.25em;
    margin: 24px 0;
    padding: 0;
}

img {
    max-width: 100%;
    box-sizing: content-box;
    background-color: #fff;
}

@media (max-width: 767px) {
    body {
        padding: 15px;
    }
}
        `;
    }

    /**
     * Strip interactive elements for static fallback
     */
    private stripInteractiveElements(html: string): string {
        // Remove script tags
        let staticHTML = html.replace(/<script[^>]*>[\s\S]*?<\/script>/gi, '');
        
        // Remove event handlers
        staticHTML = staticHTML.replace(/\s*on\w+="[^"]*"/gi, '');
        
        // Remove form elements (make them static)
        staticHTML = staticHTML.replace(/<input[^>]*>/gi, '');
        staticHTML = staticHTML.replace(/<button[^>]*>(.*?)<\/button>/gi, '<span class="static-button">$1</span>');
        staticHTML = staticHTML.replace(/<form[^>]*>([\s\S]*?)<\/form>/gi, '<div class="static-form">$1</div>');
        
        return staticHTML;
    }

    /**
     * Sanitize HTML content
     */
    private sanitizeHTML(html: string): string {
        // Remove potentially dangerous elements and attributes
        let sanitized = html;
        
        // Remove script tags
        sanitized = sanitized.replace(/<script[^>]*>[\s\S]*?<\/script>/gi, '');
        
        // Remove event handlers
        sanitized = sanitized.replace(/\s*on\w+="[^"]*"/gi, '');
        
        // Remove javascript: links
        sanitized = sanitized.replace(/href="javascript:[^"]*"/gi, 'href="#"');
        
        // Remove style attributes (optional, depending on security requirements)
        // sanitized = sanitized.replace(/\s*style="[^"]*"/gi, '');
        
        return sanitized;
    }

    /**
     * Minify HTML content
     */
    private minifyHTML(html: string): string {
        return html
            .replace(/\s+/g, ' ')
            .replace(/>\s+</g, '><')
            .replace(/\s+>/g, '>')
            .replace(/<\s+/g, '<')
            .trim();
    }

    /**
     * Escape HTML entities
     */
    private escapeHTML(text: string): string {
        // Check if we're in a browser environment
        if (typeof document !== 'undefined') {
            const div = document.createElement('div');
            div.textContent = text;
            return div.innerHTML;
        } else {
            // Fallback for test environment
            return text
                .replace(/&/g, '&amp;')
                .replace(/</g, '&lt;')
                .replace(/>/g, '&gt;')
                .replace(/"/g, '&quot;')
                .replace(/'/g, '&#39;');
        }
    }

    /**
     * Cleanup resources (no-op for HTML/Markdown converter)
     */
    destroy(): void {
        // No cleanup needed for HTML/Markdown conversion
    }
}

// Type definitions
interface ParsedHTML {
    html: string;
    css: string;
    title: string;
    metadata: Record<string, any>;
}