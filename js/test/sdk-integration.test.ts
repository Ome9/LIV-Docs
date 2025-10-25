/**
 * SDK Integration Tests
 * Tests the JavaScript SDK integration with existing infrastructure
 */

import { 
    LIVSDK, 
    LIVDocumentBuilder, 
    LIVHelpers,
    livSDK,
    LIVDocument,
    LIVLoader,
    LIVRenderer
} from '../src/sdk';

// Mock DOM environment for testing
const mockDOM = () => {
    const mockElement = {
        appendChild: jest.fn(),
        removeChild: jest.fn(),
        querySelector: jest.fn(),
        querySelectorAll: jest.fn(() => []),
        addEventListener: jest.fn(),
        removeEventListener: jest.fn(),
        getBoundingClientRect: jest.fn(() => ({ 
            width: 800, height: 600, top: 0, left: 0 
        })),
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
    
    return mockElement;
};

// Mock crypto for testing environments
if (typeof crypto === 'undefined') {
    (global as any).crypto = {
        subtle: {
            digest: jest.fn().mockResolvedValue(new ArrayBuffer(32))
        }
    };
}

describe('SDK Integration Tests', () => {
    let mockContainer: HTMLElement;

    beforeEach(() => {
        mockContainer = mockDOM();
    });

    describe('Core SDK Integration', () => {
        test('should integrate with existing loader infrastructure', async () => {
            const sdk = LIVSDK.getInstance();
            
            // Test that SDK can create a loader
            expect(sdk).toBeDefined();
            expect(typeof sdk.loadDocument).toBe('function');
            expect(typeof sdk.createRenderer).toBe('function');
        });

        test('should integrate with existing renderer infrastructure', () => {
            const sdk = LIVSDK.getInstance();
            const renderer = sdk.createRenderer(mockContainer);
            
            expect(renderer).toBeDefined();
            expect(renderer).toBeInstanceOf(LIVRenderer);
        });
    });    desc
ribe('Document Builder Integration', () => {
        test('should create documents using existing validation patterns', async () => {
            const sdk = LIVSDK.getInstance();
            const builder = await sdk.createDocument({
                metadata: {
                    title: 'Integration Test Document',
                    author: 'Test Suite'
                }
            });

            expect(builder).toBeInstanceOf(LIVDocumentBuilder);
            
            // Build document with content
            builder.setHTML('<h1>Integration Test</h1>');
            const document = await builder.build();
            
            expect(document).toBeInstanceOf(LIVDocument);
            expect(document.manifest.metadata.title).toBe('Integration Test Document');
        });

        test('should handle asset management with existing error handling', async () => {
            const sdk = LIVSDK.getInstance();
            const builder = await sdk.createDocument();
            
            // Add various asset types
            const imageData = new ArrayBuffer(1024);
            const fontData = new ArrayBuffer(2048);
            const jsonData = JSON.stringify({ test: true });
            
            builder
                .setMetadata({ title: 'Asset Test', author: 'Tester' })
                .setHTML('<h1>Asset Test</h1>')
                .addAsset({
                    type: 'image',
                    name: 'test.png',
                    data: imageData,
                    mimeType: 'image/png'
                })
                .addAsset({
                    type: 'font', 
                    name: 'font.woff2',
                    data: fontData,
                    mimeType: 'font/woff2'
                })
                .addAsset({
                    type: 'data',
                    name: 'config.json',
                    data: jsonData,
                    mimeType: 'application/json'
                });

            const document = await builder.build();
            
            // Verify assets were added correctly
            expect(document.assets.images.has('test.png')).toBe(true);
            expect(document.assets.fonts.has('font.woff2')).toBe(true);
            expect(document.assets.data.has('config.json')).toBe(true);
        });

        test('should integrate WASM modules with existing security context', async () => {
            const sdk = LIVSDK.getInstance();
            const builder = await sdk.createDocument();
            
            const wasmData = new ArrayBuffer(4096);
            
            builder
                .setMetadata({ title: 'WASM Test', author: 'Tester' })
                .setHTML('<h1>WASM Integration</h1>')
                .addWASMModule({
                    name: 'test-module',
                    data: wasmData,
                    version: '1.0.0',
                    entryPoint: 'init',
                    permissions: {
                        memoryLimit: 32 * 1024 * 1024,
                        allowNetworking: false,
                        allowFileSystem: false
                    }
                });

            const document = await builder.build();
            
            expect(document.wasmModules.has('test-module')).toBe(true);
            expect(document.manifest.features?.webassembly).toBe(true);
            expect(document.manifest.wasmConfig?.modules).toBeDefined();
        });
    });

    describe('Helper Functions Integration', () => {
        test('should create text documents using existing document patterns', async () => {
            const document = await LIVHelpers.createTextDocument(
                'Integration Text Test',
                'This tests the helper function integration with existing document infrastructure.',
                'Integration Tester'
            );

            expect(document).toBeInstanceOf(LIVDocument);
            expect(document.manifest.metadata.title).toBe('Integration Text Test');
            expect(document.content.html).toContain('Integration Text Test');
            expect(document.content.html).toContain('This tests the helper function');
        });

        test('should create chart documents with existing WASM interfaces', async () => {
            const chartData = {
                labels: ['Q1', 'Q2', 'Q3', 'Q4'],
                datasets: [{
                    label: 'Sales',
                    data: [100, 150, 200, 180]
                }]
            };

            const document = await LIVHelpers.createChartDocument(
                'Integration Chart Test',
                chartData,
                'bar'
            );

            expect(document).toBeInstanceOf(LIVDocument);
            expect(document.manifest.metadata.title).toBe('Integration Chart Test');
            expect(document.manifest.features?.charts).toBe(true);
            expect(document.manifest.features?.interactivity).toBe(true);
            expect(document.content.interactiveSpec).toContain('chartData');
        });

        test('should create presentations with existing animation systems', async () => {
            const slides = [
                { title: 'Slide 1', content: 'First slide content' },
                { title: 'Slide 2', content: 'Second slide content' },
                { title: 'Slide 3', content: 'Third slide content' }
            ];

            const document = await LIVHelpers.createPresentationDocument(
                'Integration Presentation Test',
                slides
            );

            expect(document).toBeInstanceOf(LIVDocument);
            expect(document.manifest.features?.animations).toBe(true);
            expect(document.manifest.features?.interactivity).toBe(true);
            expect(document.content.css).toContain('@keyframes');
            expect(document.content.interactiveSpec).toContain('showSlide');
        });
    }); 
   describe('Error Handling Integration', () => {
        test('should use existing error handling patterns', async () => {
            const sdk = LIVSDK.getInstance();
            
            // Test validation errors
            const builder = await sdk.createDocument();
            // Don't set required fields
            
            await expect(builder.build()).rejects.toThrow();
        });

        test('should handle renderer errors with existing error systems', () => {
            const sdk = LIVSDK.getInstance();
            
            // Test with invalid container
            expect(() => {
                sdk.createRenderer(null as any);
            }).toThrow();
        });

        test('should validate documents using existing validation systems', async () => {
            const sdk = LIVSDK.getInstance();
            const builder = await sdk.createDocument({
                metadata: {
                    title: 'Validation Test',
                    author: 'Tester'
                }
            });

            builder.setHTML('<h1>Valid Content</h1>');
            const document = await builder.build();
            
            const validation = await sdk.validateDocument(document);
            expect(validation).toHaveProperty('isValid');
            expect(validation).toHaveProperty('errors');
            expect(validation).toHaveProperty('warnings');
        });
    });

    describe('Performance Integration', () => {
        test('should handle large documents efficiently', async () => {
            const sdk = LIVSDK.getInstance();
            const builder = await sdk.createDocument();

            // Create large content to test performance
            const largeContent = Array(500).fill('<p>Large content paragraph with some text.</p>').join('\n');
            
            const startTime = performance.now();
            
            builder
                .setMetadata({ title: 'Large Document', author: 'Performance Test' })
                .setHTML(`<div class="content">${largeContent}</div>`)
                .setCSS('p { margin: 10px 0; font-size: 14px; }');

            const document = await builder.build();
            
            const endTime = performance.now();
            const buildTime = endTime - startTime;

            expect(document).toBeInstanceOf(LIVDocument);
            expect(buildTime).toBeLessThan(2000); // Should build in less than 2 seconds
            
            // Test document info
            const info = sdk.getDocumentInfo(document);
            expect(info.estimatedSize).toBeGreaterThan(0);
        });

        test('should handle multiple assets without performance degradation', async () => {
            const sdk = LIVSDK.getInstance();
            const builder = await sdk.createDocument();

            builder
                .setMetadata({ title: 'Multi-Asset Performance Test', author: 'Performance Test' })
                .setHTML('<h1>Performance Test</h1>');

            const startTime = performance.now();

            // Add many small assets
            for (let i = 0; i < 25; i++) {
                builder.addAsset({
                    type: 'data',
                    name: `asset-${i}.json`,
                    data: JSON.stringify({ index: i, data: 'test'.repeat(50) }),
                    mimeType: 'application/json'
                });
            }

            const document = await builder.build();
            
            const endTime = performance.now();
            const buildTime = endTime - startTime;

            expect(document).toBeInstanceOf(LIVDocument);
            expect(buildTime).toBeLessThan(1500); // Should build efficiently
            
            const info = sdk.getDocumentInfo(document);
            expect(info.resourceCount).toBeGreaterThanOrEqual(25);
        });
    });

    describe('Type System Integration', () => {
        test('should provide complete TypeScript definitions', () => {
            const sdk = LIVSDK.getInstance();
            
            // Test that all expected methods exist
            expect(typeof sdk.createDocument).toBe('function');
            expect(typeof sdk.loadDocument).toBe('function');
            expect(typeof sdk.createRenderer).toBe('function');
            expect(typeof sdk.createEditor).toBe('function');
            expect(typeof sdk.validateDocument).toBe('function');
            expect(typeof sdk.getDocumentInfo).toBe('function');
        });

        test('should support all documented interfaces', async () => {
            const sdk = LIVSDK.getInstance();
            
            // Test DocumentCreationOptions interface
            const options = {
                metadata: {
                    title: 'Type Test',
                    author: 'Type Tester',
                    description: 'Testing TypeScript interfaces'
                },
                features: {
                    animations: true,
                    interactivity: true,
                    charts: false
                }
            };

            const builder = await sdk.createDocument(options);
            expect(builder).toBeInstanceOf(LIVDocumentBuilder);
            
            const document = await builder.build();
            expect(document.manifest.metadata.title).toBe('Type Test');
            expect(document.manifest.features?.animations).toBe(true);
        });
    });
});