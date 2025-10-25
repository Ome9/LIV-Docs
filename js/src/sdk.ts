// LIV JavaScript SDK for Document Generation
// High-level API using existing document, loader, and renderer classes

import { LIVDocument, loadLIVDocument } from './document';
import { LIVLoader } from './loader';
import { LIVRenderer } from './renderer';
import { LIVEditor } from './editor';
import { 
    DocumentMetadata, 
    LegacySecurityPolicy,
    LegacyWASMPermissions, 
    LegacyJSPermissions, 
    LegacyNetworkPolicy, 
    LegacyStoragePolicy,
    Manifest,
    Resource,
    AssetBundle,
    ValidationResult,
    RendererOptions,
    LoaderOptions,
    DocumentContent,
    FeatureFlags,
    WASMConfiguration,
    WASMModule
} from './types';

// Import SDK-specific types
import {
    SDKDocumentCreationOptions,
    SDKRenderingOptions,
    SDKEditingOptions,
    AssetManagementOptions,
    SDKWASMModuleOptions,
    DocumentExportOptions,
    SDKValidationOptions,
    SDKPerformanceMetrics,
    SDKConfiguration,
    DocumentStatistics
} from './sdk-types';

// Re-export types for convenience
export type DocumentCreationOptions = SDKDocumentCreationOptions;
export type RenderingOptions = SDKRenderingOptions;
export type EditingOptions = SDKEditingOptions;
export type AssetOptions = AssetManagementOptions;
export type WASMModuleOptions = SDKWASMModuleOptions;

/**
 * LIV SDK - High-level API for creating and managing LIV documents
 */
export class LIVSDK {
    private static instance?: LIVSDK;

    /**
     * Get singleton instance of the SDK
     */
    static getInstance(): LIVSDK {
        if (!LIVSDK.instance) {
            LIVSDK.instance = new LIVSDK();
        }
        return LIVSDK.instance;
    }

    /**
     * Create a new LIV document with the specified content and metadata
     */
    async createDocument(options: SDKDocumentCreationOptions = {}): Promise<LIVDocumentBuilder> {
        return new LIVDocumentBuilder(options);
    }

    /**
     * Load an existing LIV document from various sources
     */
    async loadDocument(source: File | string | ArrayBuffer, options?: LoaderOptions): Promise<LIVDocument> {
        return loadLIVDocument(source, options);
    }

    /**
     * Create a renderer for displaying LIV documents
     */
    createRenderer(container: HTMLElement, options?: Partial<SDKRenderingOptions>): LIVRenderer {
        const defaultPermissions: LegacySecurityPolicy = {
            wasmPermissions: {
                memoryLimit: 64 * 1024 * 1024, // 64MB
                allowedImports: ['env'],
                cpuTimeLimit: 5000,
                allowNetworking: false,
                allowFileSystem: false
            },
            jsPermissions: {
                executionMode: 'sandboxed',
                allowedAPIs: ['dom'],
                domAccess: 'read'
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
            }
        };

        const rendererOptions: RendererOptions = {
            container,
            permissions: defaultPermissions,
            enableInteractivity: true,
            enableAnimations: true,
            fallbackMode: false,
            ...options
        };

        return new LIVRenderer(rendererOptions);
    }

    /**
     * Create an editor for modifying LIV documents
     */
    createEditor(
        editorContainer: HTMLElement,
        previewContainer: HTMLElement,
        toolbarContainer: HTMLElement,
        propertiesContainer: HTMLElement,
        options?: SDKEditingOptions
    ): LIVEditor {
        return new LIVEditor(
            editorContainer,
            previewContainer,
            toolbarContainer,
            propertiesContainer
        );
    }

    /**
     * Validate a LIV document
     */
    async validateDocument(document: LIVDocument): Promise<ValidationResult> {
        return document.validate();
    }

    /**
     * Convert between different document formats
     */
    async convertDocument(document: LIVDocument, targetFormat: 'pdf' | 'html' | 'markdown' | 'epub'): Promise<Blob> {
        // This would integrate with the conversion system from task 8
        // For now, return a placeholder
        throw new Error(`Conversion to ${targetFormat} not yet implemented`);
    }

    /**
     * Get document metadata and statistics
     */
    getDocumentInfo(document: LIVDocument) {
        return document.getMetadata();
    }
}

/**
 * Document builder class for creating new LIV documents
 */
export class LIVDocumentBuilder {
    private manifest: Partial<Manifest> = {};
    private content: Partial<DocumentContent> = {};
    private assets: AssetBundle = {
        images: new Map(),
        fonts: new Map(),
        data: new Map()
    };
    private wasmModules: Map<string, ArrayBuffer> = new Map();
    private resources: Record<string, Resource> = {};

    constructor(options: SDKDocumentCreationOptions = {}) {
        // Set default metadata
        this.manifest = {
            version: '1.0',
            metadata: {
                title: 'New LIV Document',
                author: 'Unknown',
                created: new Date().toISOString(),
                modified: new Date().toISOString(),
                description: '',
                version: '1.0',
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
                    executionMode: 'sandboxed',
                    allowedAPIs: ['dom'],
                    domAccess: 'read'
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
                }
            },
            resources: {},
            features: {
                animations: false,
                interactivity: false,
                charts: false,
                forms: false,
                audio: false,
                video: false,
                webgl: false,
                webassembly: false
            }
        };

        this.content = {
            html: '',
            css: '',
            interactiveSpec: '',
            staticFallback: ''
        };

        // Apply provided options
        if (options.metadata) {
            this.setMetadata(options.metadata);
        }
        if (options.content) {
            if (options.content.html) this.setHTML(options.content.html);
            if (options.content.css) this.setCSS(options.content.css);
            if (options.content.interactiveSpec) this.setInteractiveSpec(options.content.interactiveSpec);
            if (options.content.staticFallback) this.setStaticFallback(options.content.staticFallback);
        }
        if (options.security) {
            this.setSecurity(options.security);
        }
        if (options.features) {
            this.setFeatures(options.features);
        }
        if (options.wasmConfig) {
            this.setWASMConfig(options.wasmConfig);
        }
    }

    /**
     * Set document metadata
     */
    setMetadata(metadata: Partial<DocumentMetadata>): LIVDocumentBuilder {
        this.manifest.metadata = {
            ...this.manifest.metadata!,
            ...metadata,
            modified: new Date().toISOString()
        };
        return this;
    }

    /**
     * Set HTML content
     */
    setHTML(html: string): LIVDocumentBuilder {
        this.content.html = html;
        this.updateResource('content/index.html', html, 'text/html');
        return this;
    }

    /**
     * Set CSS styles
     */
    setCSS(css: string): LIVDocumentBuilder {
        this.content.css = css;
        this.updateResource('content/styles/main.css', css, 'text/css');
        return this;
    }

    /**
     * Set interactive specification (JavaScript)
     */
    setInteractiveSpec(spec: string): LIVDocumentBuilder {
        this.content.interactiveSpec = spec;
        this.updateResource('content/scripts/main.js', spec, 'application/javascript');
        this.manifest.features!.interactivity = spec.length > 0;
        return this;
    }

    /**
     * Set static fallback content
     */
    setStaticFallback(fallback: string): LIVDocumentBuilder {
        this.content.staticFallback = fallback;
        this.updateResource('content/static/fallback.html', fallback, 'text/html');
        return this;
    }

    /**
     * Set security policy
     */
    setSecurity(security: Partial<LegacySecurityPolicy>): LIVDocumentBuilder {
        this.manifest.security = {
            ...this.manifest.security!,
            ...security
        };
        return this;
    }

    /**
     * Set feature flags
     */
    setFeatures(features: Partial<FeatureFlags>): LIVDocumentBuilder {
        this.manifest.features = {
            ...this.manifest.features!,
            ...features
        };
        return this;
    }

    /**
     * Set WASM configuration
     */
    setWASMConfig(config: Partial<WASMConfiguration>): LIVDocumentBuilder {
        if (!this.manifest.wasmConfig) {
            this.manifest.wasmConfig = {
                modules: {},
                permissions: this.manifest.security!.wasmPermissions!,
                memoryLimit: this.manifest.security!.wasmPermissions!.memoryLimit
            };
        }
        
        this.manifest.wasmConfig = {
            ...this.manifest.wasmConfig,
            ...config,
            modules: {
                ...this.manifest.wasmConfig.modules,
                ...(config.modules || {})
            }
        };
        
        if (config.modules && Object.keys(config.modules).length > 0) {
            this.manifest.features!.webassembly = true;
        }
        return this;
    }

    /**
     * Add an asset to the document
     */
    addAsset(options: AssetManagementOptions): LIVDocumentBuilder {
        const { type, name, data } = options;
        
        let buffer: ArrayBuffer;
        if (typeof data === 'string') {
            buffer = new TextEncoder().encode(data).buffer;
        } else if (data instanceof Blob) {
            // Handle Blob data - would need to be converted to ArrayBuffer
            throw new Error('Blob data not yet supported - please convert to ArrayBuffer first');
        } else {
            buffer = data;
        }

        // Store asset based on type
        switch (type) {
            case 'image':
                this.assets.images.set(name, buffer);
                break;
            case 'font':
                this.assets.fonts.set(name, buffer);
                break;
            case 'data':
                this.assets.data.set(name, buffer);
                break;
            case 'audio':
            case 'video':
                // For now, store audio/video as data assets
                this.assets.data.set(name, buffer);
                break;
            default:
                throw new Error(`Unsupported asset type: ${type}`);
        }

        // Update resource manifest
        const resourcePath = `assets/${type}s/${name}`;
        this.updateResourceFromBuffer(resourcePath, buffer, options.mimeType || 'application/octet-stream');

        return this;
    }

    /**
     * Add a WASM module to the document
     */
    addWASMModule(options: SDKWASMModuleOptions): LIVDocumentBuilder {
        const { name, data, version, entryPoint, permissions } = options;
        
        // Store WASM module
        this.wasmModules.set(name, data);

        // Update WASM configuration
        if (!this.manifest.wasmConfig) {
            this.manifest.wasmConfig = {
                modules: {},
                permissions: this.manifest.security!.wasmPermissions!,
                memoryLimit: this.manifest.security!.wasmPermissions!.memoryLimit
            };
        }

        if (!this.manifest.wasmConfig.modules) {
            this.manifest.wasmConfig.modules = {};
        }

        const wasmModule: WASMModule = {
            name,
            version: version || '1.0',
            entryPoint: entryPoint || 'main',
            exports: [], // Would be populated by WASM analysis
            imports: [], // Would be populated by WASM analysis
            permissions: permissions as LegacyWASMPermissions,
            metadata: {}
        };

        this.manifest.wasmConfig.modules![name] = wasmModule;
        this.manifest.features!.webassembly = true;

        // Update resource manifest
        const resourcePath = `${name}.wasm`;
        this.updateResourceFromBuffer(resourcePath, data, 'application/wasm');

        return this;
    }

    /**
     * Enable animations in the document
     */
    enableAnimations(): LIVDocumentBuilder {
        this.manifest.features!.animations = true;
        return this;
    }

    /**
     * Enable charts in the document
     */
    enableCharts(): LIVDocumentBuilder {
        this.manifest.features!.charts = true;
        return this;
    }

    /**
     * Enable forms in the document
     */
    enableForms(): LIVDocumentBuilder {
        this.manifest.features!.forms = true;
        return this;
    }

    /**
     * Build the final LIV document
     */
    async build(): Promise<LIVDocument> {
        // Validate required fields
        if (!this.content.html && !this.content.staticFallback) {
            throw new Error('Document must have HTML content or static fallback');
        }

        if (!this.manifest.metadata?.title) {
            throw new Error('Document title is required');
        }

        if (!this.manifest.metadata?.author) {
            throw new Error('Document author is required');
        }

        // Update resource manifest
        this.manifest.resources = this.resources;

        // Create signatures (placeholder - would be actual signing in production)
        const signatures = {
            contentSignature: await this.generateSignature(this.content.html || ''),
            manifestSignature: await this.generateSignature(JSON.stringify(this.manifest)),
            wasmSignatures: {} as Record<string, string>
        };

        // Sign WASM modules
        for (const [name, data] of this.wasmModules.entries()) {
            signatures.wasmSignatures[name] = await this.generateSignature(new TextDecoder().decode(data));
        }

        // Create the document
        return new LIVDocument(
            this.manifest as Manifest,
            this.content as DocumentContent,
            this.assets,
            signatures,
            this.wasmModules
        );
    }

    private updateResource(path: string, content: string, mimeType: string): void {
        const data = new TextEncoder().encode(content);
        this.updateResourceFromBuffer(path, data.buffer, mimeType);
    }

    private updateResourceFromBuffer(path: string, buffer: ArrayBuffer, mimeType: string): void {
        this.resources[path] = {
            hash: this.generateHash(buffer),
            size: buffer.byteLength,
            type: mimeType,
            path
        };
    }

    private generateHash(data: ArrayBuffer): string {
        // Simplified hash generation - in production would use crypto.subtle.digest
        const view = new Uint8Array(data);
        let hash = 0;
        for (let i = 0; i < view.length; i++) {
            hash = ((hash << 5) - hash + view[i]) & 0xffffffff;
        }
        return `sha256-${hash.toString(16)}`;
    }

    private async generateSignature(content: string): Promise<string> {
        // Simplified signature generation - in production would use proper cryptographic signing
        try {
            const encoder = new TextEncoder();
            const data = encoder.encode(content);
            const hashBuffer = await crypto.subtle.digest('SHA-256', data);
            const hashArray = Array.from(new Uint8Array(hashBuffer));
            return hashArray.map(b => b.toString(16).padStart(2, '0')).join('');
        } catch (error) {
            // Fallback for environments without crypto.subtle
            return `sig-${Date.now()}-${Math.random().toString(36)}`;
        }
    }
}

/**
 * Helper functions for common document creation patterns
 */
export class LIVHelpers {
    /**
     * Create a simple text document
     */
    static async createTextDocument(title: string, content: string, author?: string): Promise<LIVDocument> {
        const sdk = LIVSDK.getInstance();
        const builder = await sdk.createDocument({
            metadata: {
                title,
                author: author || 'Unknown',
                description: `Simple text document: ${title}`
            }
        });

        const html = `
<!DOCTYPE html>
<html lang="en">
<head>
    <meta charset="UTF-8">
    <meta name="viewport" content="width=device-width, initial-scale=1.0">
    <title>${title}</title>
    <style>
        body {
            font-family: -apple-system, BlinkMacSystemFont, 'Segoe UI', Roboto, sans-serif;
            line-height: 1.6;
            max-width: 800px;
            margin: 0 auto;
            padding: 20px;
            color: #333;
        }
        h1 {
            color: #2c3e50;
            border-bottom: 2px solid #3498db;
            padding-bottom: 10px;
        }
    </style>
</head>
<body>
    <h1>${title}</h1>
    <div class="content">
        ${content.split('\n').map(p => `<p>${p}</p>`).join('\n')}
    </div>
</body>
</html>`;

        return builder.setHTML(html).build();
    }

    /**
     * Create an interactive chart document
     */
    static async createChartDocument(
        title: string, 
        chartData: any, 
        chartType: 'bar' | 'line' | 'pie' = 'bar'
    ): Promise<LIVDocument> {
        const sdk = LIVSDK.getInstance();
        const builder = await sdk.createDocument({
            metadata: {
                title,
                description: `Interactive ${chartType} chart: ${title}`
            },
            features: {
                charts: true,
                interactivity: true,
                webassembly: true
            }
        });

        const html = `
<!DOCTYPE html>
<html lang="en">
<head>
    <meta charset="UTF-8">
    <meta name="viewport" content="width=device-width, initial-scale=1.0">
    <title>${title}</title>
</head>
<body>
    <div id="chart-container">
        <h1>${title}</h1>
        <div id="chart" style="width: 100%; height: 400px;"></div>
    </div>
</body>
</html>`;

        const interactiveSpec = `
// Chart rendering logic would be implemented here
// This would interface with the WASM chart engine
const chartData = ${JSON.stringify(chartData)};
const chartType = "${chartType}";

// Initialize chart when WASM module is ready
if (window.wasmChartEngine) {
    window.wasmChartEngine.createChart('chart', chartType, chartData);
}`;

        return builder
            .setHTML(html)
            .setInteractiveSpec(interactiveSpec)
            .enableCharts()
            .build();
    }

    /**
     * Create an animated presentation document
     */
    static async createPresentationDocument(
        title: string, 
        slides: Array<{title: string, content: string}>
    ): Promise<LIVDocument> {
        const sdk = LIVSDK.getInstance();
        const builder = await sdk.createDocument({
            metadata: {
                title,
                description: `Animated presentation: ${title}`
            },
            features: {
                animations: true,
                interactivity: true
            }
        });

        const html = `
<!DOCTYPE html>
<html lang="en">
<head>
    <meta charset="UTF-8">
    <meta name="viewport" content="width=device-width, initial-scale=1.0">
    <title>${title}</title>
</head>
<body>
    <div class="presentation">
        <div class="slide-container">
            ${slides.map((slide, index) => `
                <div class="slide" data-slide="${index}" ${index === 0 ? 'style="display: block;"' : 'style="display: none;"'}>
                    <h1>${slide.title}</h1>
                    <div class="content">${slide.content}</div>
                </div>
            `).join('')}
        </div>
        <div class="controls">
            <button id="prev">Previous</button>
            <span id="slide-counter">1 / ${slides.length}</span>
            <button id="next">Next</button>
        </div>
    </div>
</body>
</html>`;

        const css = `
.presentation {
    max-width: 900px;
    margin: 0 auto;
    padding: 20px;
    font-family: -apple-system, BlinkMacSystemFont, 'Segoe UI', Roboto, sans-serif;
}

.slide-container {
    min-height: 500px;
    border: 1px solid #ddd;
    border-radius: 8px;
    padding: 40px;
    background: white;
    box-shadow: 0 2px 10px rgba(0,0,0,0.1);
}

.slide {
    animation: slideIn 0.5s ease-in-out;
}

@keyframes slideIn {
    from { opacity: 0; transform: translateX(20px); }
    to { opacity: 1; transform: translateX(0); }
}

.controls {
    text-align: center;
    margin-top: 20px;
}

.controls button {
    padding: 10px 20px;
    margin: 0 10px;
    border: none;
    background: #3498db;
    color: white;
    border-radius: 4px;
    cursor: pointer;
}

.controls button:hover {
    background: #2980b9;
}

.controls button:disabled {
    background: #bdc3c7;
    cursor: not-allowed;
}`;

        const interactiveSpec = `
let currentSlide = 0;
const totalSlides = ${slides.length};

function showSlide(index) {
    document.querySelectorAll('.slide').forEach((slide, i) => {
        slide.style.display = i === index ? 'block' : 'none';
    });
    
    document.getElementById('slide-counter').textContent = \`\${index + 1} / \${totalSlides}\`;
    document.getElementById('prev').disabled = index === 0;
    document.getElementById('next').disabled = index === totalSlides - 1;
}

document.getElementById('prev').addEventListener('click', () => {
    if (currentSlide > 0) {
        currentSlide--;
        showSlide(currentSlide);
    }
});

document.getElementById('next').addEventListener('click', () => {
    if (currentSlide < totalSlides - 1) {
        currentSlide++;
        showSlide(currentSlide);
    }
});

// Initialize
showSlide(0);`;

        return builder
            .setHTML(html)
            .setCSS(css)
            .setInteractiveSpec(interactiveSpec)
            .enableAnimations()
            .build();
    }
}

// Export the main SDK instance and helper functions
export const livSDK = LIVSDK.getInstance();

// Re-export core classes for convenience
export { LIVDocument } from './document';
export { LIVLoader } from './loader';
export { LIVRenderer } from './renderer';
export { LIVEditor } from './editor';

// Export SDK-specific types
export * from './sdk-types';