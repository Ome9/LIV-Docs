/**
 * LIV JavaScript SDK Tests
 * Comprehensive test suite for the SDK functionality
 */

import { 
    LIVSDK, 
    LIVDocumentBuilder, 
    LIVHelpers,
    livSDK,
    DocumentCreationOptions,
    AssetManagementOptions,
    SDKWASMModuleOptions
} from '../src/sdk';
import { LIVDocument } from '../src/document';

// Mock DOM elements for testing
const createMockElement = (): HTMLElement => {
    const element = {
        appendChild: jest.fn(),
        removeChild: jest.fn(),
        querySelector: jest.fn(),
        querySelectorAll: jest.fn(() => []),
        addEventListener: jest.fn(),
        removeEventListener: jest.fn(),
        getBoundingClientRect: jest.fn(() => ({ width: 800, height: 600, top: 0, left: 0 })),
        style: {},
        classList: {
            add: jest.fn(),
            remove: jest.fn(),
            toggle: jest.fn(),
            contains: jest.fn()
        },
        attachShadow: jest.fn(() => ({
            appendChild: jest.fn(),
            children: [],
            querySelectorAll: jest.fn(() => [])
        }))
    } as any;
    return element;
};

// Mock crypto for testing environments
if (typeof crypto === 'undefined') {
    (global as any).crypto = {
        subtle: {
            digest: jest.fn().mockResolvedValue(new ArrayBuffer(32))
        }
    };
}

describe('LIVSDK', () => {
    let sdk: LIVSDK;
    let mockContainer: HTMLElement;

    beforeEach(() => {
        sdk = LIVSDK.getInstance();
        mockContainer = createMockElement();
    });

    describe('Singleton Pattern', () => {
        test('should return the same instance', () => {
            const sdk1 = LIVSDK.getInstance();
            const sdk2 = LIVSDK.getInstance();
            expect(sdk1).toBe(sdk2);
        });

        test('should return the same instance as livSDK export', () => {
            expect(sdk).toBe(livSDK);
        });
    });

    describe('Document Creation', () => {
        test('should create a document builder with default options', async () => {
            const builder = await sdk.createDocument();
            expect(builder).toBeInstanceOf(LIVDocumentBuilder);
        });

        test('should create a document builder with custom options', async () => {
            const options: DocumentCreationOptions = {
                metadata: {
                    title: 'Test Document',
                    author: 'Test Author'
                },
                features: {
                    animations: true,
                    interactivity: true
                }
            };

            const builder = await sdk.createDocument(options);
            expect(builder).toBeInstanceOf(LIVDocumentBuilder);
        });
    });

    describe('Renderer Creation', () => {
        test('should create a renderer with default options', () => {
            const renderer = sdk.createRenderer(mockContainer);
            expect(renderer).toBeDefined();
        });

        test('should create a renderer with custom options', () => {
            const renderer = sdk.createRenderer(mockContainer, {
                enableFallback: true,
                strictSecurity: true,
                maxRenderTime: 10000
            });
            expect(renderer).toBeDefined();
        });
    });

    describe('Editor Creation', () => {
        test('should create an editor with required containers', () => {
            const editor = sdk.createEditor(
                mockContainer,
                mockContainer,
                mockContainer,
                mockContainer
            );
            expect(editor).toBeDefined();
        });

        test('should create an editor with options', () => {
            const editor = sdk.createEditor(
                mockContainer,
                mockContainer,
                mockContainer,
                mockContainer,
                {
                    mode: 'split',
                    enablePreview: true,
                    autoSave: true
                }
            );
            expect(editor).toBeDefined();
        });
    });

    describe('Document Validation', () => {
        test('should validate a document', async () => {
            const builder = await sdk.createDocument({
                metadata: {
                    title: 'Test Document',
                    author: 'Test Author'
                }
            });

            builder.setHTML('<h1>Test Content</h1>');
            const document = await builder.build();
            
            const validation = await sdk.validateDocument(document);
            expect(validation).toHaveProperty('isValid');
            expect(validation).toHaveProperty('errors');
            expect(validation).toHaveProperty('warnings');
        });
    });

    describe('Document Information', () => {
        test('should get document information', async () => {
            const builder = await sdk.createDocument({
                metadata: {
                    title: 'Test Document',
                    author: 'Test Author'
                }
            });

            builder.setHTML('<h1>Test Content</h1>');
            const document = await builder.build();
            
            const info = sdk.getDocumentInfo(document);
            expect(info).toHaveProperty('title', 'Test Document');
            expect(info).toHaveProperty('author', 'Test Author');
            expect(info).toHaveProperty('estimatedSize');
        });
    });
});

describe('LIVDocumentBuilder', () => {
    let builder: LIVDocumentBuilder;

    beforeEach(async () => {
        const sdk = LIVSDK.getInstance();
        builder = await sdk.createDocument();
    });

    describe('Metadata Management', () => {
        test('should set metadata', () => {
            const result = builder.setMetadata({
                title: 'Test Title',
                author: 'Test Author',
                description: 'Test Description'
            });
            expect(result).toBe(builder); // Should return builder for chaining
        });
    });

    describe('Content Management', () => {
        test('should set HTML content', () => {
            const html = '<h1>Test HTML</h1><p>Content</p>';
            const result = builder.setHTML(html);
            expect(result).toBe(builder);
        });

        test('should set CSS content', () => {
            const css = 'h1 { color: blue; } p { font-size: 16px; }';
            const result = builder.setCSS(css);
            expect(result).toBe(builder);
        });

        test('should set interactive specification', () => {
            const spec = 'console.log("Interactive content");';
            const result = builder.setInteractiveSpec(spec);
            expect(result).toBe(builder);
        });

        test('should set static fallback', () => {
            const fallback = '<h1>Static Fallback</h1>';
            const result = builder.setStaticFallback(fallback);
            expect(result).toBe(builder);
        });
    });

    describe('Security Management', () => {
        test('should set security policy', () => {
            const security = {
                wasmPermissions: {
                    memoryLimit: 32 * 1024 * 1024,
                    allowNetworking: false
                }
            };
            const result = builder.setSecurity(security);
            expect(result).toBe(builder);
        });
    });

    describe('Feature Management', () => {
        test('should set features', () => {
            const features = {
                animations: true,
                interactivity: true,
                charts: true
            };
            const result = builder.setFeatures(features);
            expect(result).toBe(builder);
        });

        test('should enable animations', () => {
            const result = builder.enableAnimations();
            expect(result).toBe(builder);
        });

        test('should enable charts', () => {
            const result = builder.enableCharts();
            expect(result).toBe(builder);
        });

        test('should enable forms', () => {
            const result = builder.enableForms();
            expect(result).toBe(builder);
        });
    });

    describe('Asset Management', () => {
        test('should add image asset', () => {
            const imageData = new ArrayBuffer(1024);
            const asset: AssetManagementOptions = {
                type: 'image',
                name: 'test.png',
                data: imageData,
                mimeType: 'image/png'
            };
            const result = builder.addAsset(asset);
            expect(result).toBe(builder);
        });

        test('should add font asset', () => {
            const fontData = new ArrayBuffer(2048);
            const asset: AssetManagementOptions = {
                type: 'font',
                name: 'font.woff2',
                data: fontData,
                mimeType: 'font/woff2'
            };
            const result = builder.addAsset(asset);
            expect(result).toBe(builder);
        });

        test('should add data asset from string', () => {
            const jsonData = JSON.stringify({ key: 'value' });
            const asset: AssetManagementOptions = {
                type: 'data',
                name: 'config.json',
                data: jsonData,
                mimeType: 'application/json'
            };
            const result = builder.addAsset(asset);
            expect(result).toBe(builder);
        });

        test('should throw error for unsupported asset type', () => {
            const asset = {
                type: 'unsupported' as any,
                name: 'test',
                data: new ArrayBuffer(100)
            };
            expect(() => builder.addAsset(asset)).toThrow('Unsupported asset type');
        });
    });

    describe('WASM Module Management', () => {
        test('should add WASM module', () => {
            const wasmData = new ArrayBuffer(4096);
            const module: SDKWASMModuleOptions = {
                name: 'test-module',
                data: wasmData,
                version: '1.0.0',
                entryPoint: 'main'
            };
            const result = builder.addWASMModule(module);
            expect(result).toBe(builder);
        });
    });

    describe('Document Building', () => {
        test('should build document with minimal content', async () => {
            builder
                .setMetadata({ title: 'Test', author: 'Author' })
                .setHTML('<h1>Test</h1>');

            const document = await builder.build();
            expect(document).toBeInstanceOf(LIVDocument);
        });

        test('should build document with all features', async () => {
            const imageData = new ArrayBuffer(1024);
            const wasmData = new ArrayBuffer(4096);

            builder
                .setMetadata({ title: 'Complex Document', author: 'Author' })
                .setHTML('<h1>Complex Document</h1>')
                .setCSS('h1 { color: red; }')
                .setInteractiveSpec('console.log("Interactive");')
                .addAsset({
                    type: 'image',
                    name: 'logo.png',
                    data: imageData,
                    mimeType: 'image/png'
                })
                .addWASMModule({
                    name: 'engine',
                    data: wasmData,
                    version: '1.0.0'
                })
                .enableAnimations()
                .enableCharts();

            const document = await builder.build();
            expect(document).toBeInstanceOf(LIVDocument);
        });

        test('should throw error when building without required fields', async () => {
            // No title or author set
            builder.setHTML('<h1>Test</h1>');
            
            await expect(builder.build()).rejects.toThrow('Document title is required');
        });

        test('should throw error when building without content', async () => {
            builder.setMetadata({ title: 'Test', author: 'Author' });
            // No HTML or fallback content
            
            await expect(builder.build()).rejects.toThrow('Document must have HTML content or static fallback');
        });
    });

    describe('Method Chaining', () => {
        test('should support fluent API chaining', async () => {
            const document = await builder
                .setMetadata({ title: 'Chained Document', author: 'Author' })
                .setHTML('<h1>Chained</h1>')
                .setCSS('h1 { color: green; }')
                .enableAnimations()
                .enableCharts()
                .build();

            expect(document).toBeInstanceOf(LIVDocument);
        });
    });
});

describe('LIVHelpers', () => {
    describe('Text Document Creation', () => {
        test('should create a simple text document', async () => {
            const document = await LIVHelpers.createTextDocument(
                'Test Document',
                'This is test content.',
                'Test Author'
            );

            expect(document).toBeInstanceOf(LIVDocument);
            
            const metadata = document.getMetadata();
            expect(metadata.title).toBe('Test Document');
            expect(metadata.author).toBe('Test Author');
        });

        test('should create text document without author', async () => {
            const document = await LIVHelpers.createTextDocument(
                'Test Document',
                'This is test content.'
            );

            expect(document).toBeInstanceOf(LIVDocument);
            
            const metadata = document.getMetadata();
            expect(metadata.author).toBe('Unknown');
        });
    });

    describe('Chart Document Creation', () => {
        test('should create a chart document', async () => {
            const chartData = {
                labels: ['A', 'B', 'C'],
                datasets: [{
                    label: 'Test Data',
                    data: [1, 2, 3]
                }]
            };

            const document = await LIVHelpers.createChartDocument(
                'Test Chart',
                chartData,
                'bar'
            );

            expect(document).toBeInstanceOf(LIVDocument);
            
            const metadata = document.getMetadata();
            expect(metadata.title).toBe('Test Chart');
            expect(metadata.hasCharts).toBe(true);
            expect(metadata.hasInteractiveContent).toBe(true);
        });
    });

    describe('Presentation Document Creation', () => {
        test('should create a presentation document', async () => {
            const slides = [
                { title: 'Slide 1', content: 'Content 1' },
                { title: 'Slide 2', content: 'Content 2' },
                { title: 'Slide 3', content: 'Content 3' }
            ];

            const document = await LIVHelpers.createPresentationDocument(
                'Test Presentation',
                slides
            );

            expect(document).toBeInstanceOf(LIVDocument);
            
            const metadata = document.getMetadata();
            expect(metadata.title).toBe('Test Presentation');
            expect(metadata.hasAnimations).toBe(true);
            expect(metadata.hasInteractiveContent).toBe(true);
        });
    });
});

describe('Error Handling', () => {
    let sdk: LIVSDK;

    beforeEach(() => {
        sdk = LIVSDK.getInstance();
    });

    test('should handle invalid document loading gracefully', async () => {
        const invalidData = new ArrayBuffer(10); // Too small to be valid
        
        await expect(sdk.loadDocument(invalidData)).rejects.toThrow();
    });

    test('should handle renderer creation with invalid container', () => {
        expect(() => {
            sdk.createRenderer(null as any);
        }).toThrow();
    });
});

describe('Integration Tests', () => {
    test('should create, build, and validate a complete document', async () => {
        const sdk = LIVSDK.getInstance();
        
        // Create document
        const builder = await sdk.createDocument({
            metadata: {
                title: 'Integration Test Document',
                author: 'Test Suite',
                description: 'A document created for integration testing'
            },
            features: {
                animations: true,
                interactivity: true
            }
        });

        // Add content
        builder
            .setHTML(`
                <div class="container">
                    <h1>Integration Test</h1>
                    <p>This document tests the complete SDK workflow.</p>
                    <div id="interactive-area"></div>
                </div>
            `)
            .setCSS(`
                .container {
                    max-width: 800px;
                    margin: 0 auto;
                    padding: 20px;
                    font-family: Arial, sans-serif;
                }
                h1 {
                    color: #333;
                    animation: fadeIn 1s ease-in;
                }
                @keyframes fadeIn {
                    from { opacity: 0; }
                    to { opacity: 1; }
                }
            `)
            .setInteractiveSpec(`
                document.addEventListener('DOMContentLoaded', () => {
                    const area = document.getElementById('interactive-area');
                    if (area) {
                        area.innerHTML = '<p>Interactive content loaded!</p>';
                    }
                });
            `);

        // Add assets
        const testData = JSON.stringify({ test: true, timestamp: Date.now() });
        builder.addAsset({
            type: 'data',
            name: 'test-data.json',
            data: testData,
            mimeType: 'application/json'
        });

        // Build document
        const document = await builder.build();
        expect(document).toBeInstanceOf(LIVDocument);

        // Validate document
        const validation = await sdk.validateDocument(document);
        expect(validation.isValid).toBe(true);
        expect(validation.errors).toHaveLength(0);

        // Get document info
        const info = sdk.getDocumentInfo(document);
        expect(info.title).toBe('Integration Test Document');
        expect(info.hasAnimations).toBe(true);
        expect(info.hasInteractiveContent).toBe(true);
        expect(info.resourceCount).toBeGreaterThan(0);

        // Test renderer creation (without actual rendering)
        const mockContainer = createMockElement();
        const renderer = sdk.createRenderer(mockContainer, {
            enableFallback: true,
            strictSecurity: true
        });
        expect(renderer).toBeDefined();
    });
});

describe('Performance Tests', () => {
    test('should handle large documents efficiently', async () => {
        const sdk = LIVSDK.getInstance();
        const builder = await sdk.createDocument();

        // Create large content
        const largeContent = Array(1000).fill('<p>Large content paragraph</p>').join('\n');
        
        const startTime = performance.now();
        
        builder
            .setMetadata({ title: 'Large Document', author: 'Performance Test' })
            .setHTML(`<div>${largeContent}</div>`);

        const document = await builder.build();
        
        const endTime = performance.now();
        const buildTime = endTime - startTime;

        expect(document).toBeInstanceOf(LIVDocument);
        expect(buildTime).toBeLessThan(1000); // Should build in less than 1 second
    });

    test('should handle multiple assets efficiently', async () => {
        const sdk = LIVSDK.getInstance();
        const builder = await sdk.createDocument();

        builder.setMetadata({ title: 'Multi-Asset Document', author: 'Performance Test' })
               .setHTML('<h1>Multi-Asset Test</h1>');

        const startTime = performance.now();

        // Add multiple assets
        for (let i = 0; i < 50; i++) {
            builder.addAsset({
                type: 'data',
                name: `asset-${i}.json`,
                data: JSON.stringify({ index: i, data: 'test'.repeat(100) }),
                mimeType: 'application/json'
            });
        }

        const document = await builder.build();
        
        const endTime = performance.now();
        const buildTime = endTime - startTime;

        expect(document).toBeInstanceOf(LIVDocument);
        expect(buildTime).toBeLessThan(2000); // Should build in less than 2 seconds
        
        const info = sdk.getDocumentInfo(document);
        expect(info.resourceCount).toBeGreaterThanOrEqual(50);
    });
});