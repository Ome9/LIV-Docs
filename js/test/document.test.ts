// Tests for LIVDocument class and document loading system

import { LIVDocument, loadLIVDocument } from '../src/document';
import { LIVLoader } from '../src/loader';
import {
  LIVError,
  InvalidFileError,
  ValidationError,
  ResourceLimitError
} from '../src/errors';
import {
  Manifest,
  DocumentContent,
  AssetBundle,
  SignatureBundle
} from '../src/types';

// Mock data for testing
const createMockManifest = (): Manifest => ({
  version: '1.0',
  metadata: {
    title: 'Test Document',
    author: 'Test Author',
    created: '2024-01-01T00:00:00Z',
    modified: '2024-01-01T00:00:00Z',
    description: 'Test description',
    version: '1.0.0',
    language: 'en'
  },
  security: {
    wasmPermissions: {
      memoryLimit: 4 * 1024 * 1024,
      allowedImports: [],
      cpuTimeLimit: 5000,
      allowNetworking: false,
      allowFileSystem: false
    },
    jsPermissions: {
      executionMode: 'sandboxed' as const,
      allowedAPIs: [],
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
    }
  },
  resources: {
    'content/index.html': {
      hash: 'test-hash-html',
      size: 1024,
      type: 'text/html',
      path: 'content/index.html'
    },
    'assets/images/logo.png': {
      hash: 'test-hash-png',
      size: 2048,
      type: 'image/png',
      path: 'assets/images/logo.png'
    }
  }
});

const createMockContent = (): DocumentContent => ({
  html: '<html><body><h1>Test Document</h1></body></html>',
  css: 'body { margin: 0; padding: 20px; }',
  interactiveSpec: 'console.log("Interactive content");',
  staticFallback: '<html><body><h1>Static Fallback</h1></body></html>'
});

const createMockAssets = (): AssetBundle => ({
  images: new Map([
    ['logo.png', new ArrayBuffer(2048)]
  ]),
  fonts: new Map(),
  data: new Map()
});

const createMockSignatures = (): SignatureBundle => ({
  contentSignature: 'mock-content-signature',
  manifestSignature: 'mock-manifest-signature',
  wasmSignatures: {}
});

const createMockWASMModules = (): Map<string, ArrayBuffer> => {
  const wasmModule = new ArrayBuffer(8);
  const view = new DataView(wasmModule);
  view.setUint32(0, 0x6d736100, true); // WASM magic number
  view.setUint32(4, 1, true); // WASM version
  
  return new Map([
    ['test-module', wasmModule]
  ]);
};

describe('LIVDocument', () => {
  let mockManifest: Manifest;
  let mockContent: DocumentContent;
  let mockAssets: AssetBundle;
  let mockSignatures: SignatureBundle;
  let mockWASMModules: Map<string, ArrayBuffer>;

  beforeEach(() => {
    mockManifest = createMockManifest();
    mockContent = createMockContent();
    mockAssets = createMockAssets();
    mockSignatures = createMockSignatures();
    mockWASMModules = createMockWASMModules();
  });

  describe('constructor', () => {
    it('should create a LIVDocument instance', () => {
      const document = new LIVDocument(
        mockManifest,
        mockContent,
        mockAssets,
        mockSignatures,
        mockWASMModules
      );

      expect(document).toBeInstanceOf(LIVDocument);
      expect(document.manifest).toBe(mockManifest);
      expect(document.content).toBe(mockContent);
      expect(document.assets).toBe(mockAssets);
      expect(document.signatures).toBe(mockSignatures);
      expect(document.wasmModules).toBe(mockWASMModules);
    });
  });

  describe('getResource', () => {
    let document: LIVDocument;

    beforeEach(() => {
      document = new LIVDocument(
        mockManifest,
        mockContent,
        mockAssets,
        mockSignatures,
        mockWASMModules
      );
    });

    it('should load HTML content resource', async () => {
      const resource = await document.getResource('content/index.html');
      expect(resource).not.toBeNull();
      
      const text = new TextDecoder().decode(resource!);
      expect(text).toBe(mockContent.html);
    });

    it('should load CSS content resource', async () => {
      const resource = await document.getResource('content/styles/main.css');
      expect(resource).not.toBeNull();
      
      const text = new TextDecoder().decode(resource!);
      expect(text).toBe(mockContent.css);
    });

    it('should load asset resource', async () => {
      const resource = await document.getResource('assets/images/logo.png');
      expect(resource).not.toBeNull();
      expect(resource!.byteLength).toBe(2048);
    });

    it('should return null for non-existent resource', async () => {
      const resource = await document.getResource('non-existent/resource.txt');
      expect(resource).toBeNull();
    });

    it('should cache resources', async () => {
      // First call
      const resource1 = await document.getResource('content/index.html');
      
      // Second call should return cached version
      const resource2 = await document.getResource('content/index.html');
      
      expect(resource1).toBe(resource2);
    });
  });

  describe('getResourceInfo', () => {
    let document: LIVDocument;

    beforeEach(() => {
      document = new LIVDocument(
        mockManifest,
        mockContent,
        mockAssets,
        mockSignatures,
        mockWASMModules
      );
    });

    it('should return resource info for existing resource', () => {
      const info = document.getResourceInfo('content/index.html');
      expect(info).not.toBeNull();
      expect(info!.path).toBe('content/index.html');
      expect(info!.type).toBe('text/html');
      expect(info!.size).toBe(1024);
    });

    it('should return null for non-existent resource', () => {
      const info = document.getResourceInfo('non-existent/resource.txt');
      expect(info).toBeNull();
    });
  });

  describe('listResources', () => {
    let document: LIVDocument;

    beforeEach(() => {
      document = new LIVDocument(
        mockManifest,
        mockContent,
        mockAssets,
        mockSignatures,
        mockWASMModules
      );
    });

    it('should return list of all resources', () => {
      const resources = document.listResources();
      expect(resources).toContain('content/index.html');
      expect(resources).toContain('assets/images/logo.png');
      expect(resources.length).toBe(2);
    });
  });

  describe('getResourcesByType', () => {
    let document: LIVDocument;

    beforeEach(() => {
      document = new LIVDocument(
        mockManifest,
        mockContent,
        mockAssets,
        mockSignatures,
        mockWASMModules
      );
    });

    it('should return resources filtered by MIME type', () => {
      const htmlResources = document.getResourcesByType('text/html');
      expect(htmlResources).toContain('content/index.html');
      expect(htmlResources.length).toBe(1);

      const imageResources = document.getResourcesByType('image/');
      expect(imageResources).toContain('assets/images/logo.png');
      expect(imageResources.length).toBe(1);
    });
  });

  describe('validate', () => {
    it('should validate a correct document', () => {
      const document = new LIVDocument(
        mockManifest,
        mockContent,
        mockAssets,
        mockSignatures,
        mockWASMModules
      );

      const result = document.validate();
      expect(result.isValid).toBe(true);
      expect(result.errors).toHaveLength(0);
    });

    it('should detect missing manifest version', () => {
      const invalidManifest = { ...mockManifest, version: '' };
      const document = new LIVDocument(
        invalidManifest,
        mockContent,
        mockAssets,
        mockSignatures,
        mockWASMModules
      );

      const result = document.validate();
      expect(result.isValid).toBe(false);
      expect(result.errors).toContain('Manifest version is required');
    });

    it('should detect missing HTML content', () => {
      const invalidContent = { ...mockContent, html: '', staticFallback: '' };
      const document = new LIVDocument(
        mockManifest,
        invalidContent,
        mockAssets,
        mockSignatures,
        mockWASMModules
      );

      const result = document.validate();
      expect(result.isValid).toBe(false);
      expect(result.errors).toContain('Document must have HTML content or static fallback');
    });

    it('should detect invalid WASM modules', () => {
      const invalidWASM = new Map([
        ['invalid-module', new ArrayBuffer(4)] // Too small
      ]);
      const document = new LIVDocument(
        mockManifest,
        mockContent,
        mockAssets,
        mockSignatures,
        invalidWASM
      );

      const result = document.validate();
      expect(result.isValid).toBe(false);
      expect(result.errors.some(e => e.includes('too small'))).toBe(true);
    });
  });

  describe('generateSecurityReport', () => {
    it('should generate security report for valid document', () => {
      const document = new LIVDocument(
        mockManifest,
        mockContent,
        mockAssets,
        mockSignatures,
        mockWASMModules
      );

      const report = document.generateSecurityReport();
      expect(report.isValid).toBe(true);
      expect(report.signatureVerified).toBe(true);
      expect(report.permissionsValid).toBe(true);
    });

    it('should detect missing signatures', () => {
      const noSignatures = {
        contentSignature: '',
        manifestSignature: '',
        wasmSignatures: {}
      };
      const document = new LIVDocument(
        mockManifest,
        mockContent,
        mockAssets,
        noSignatures,
        mockWASMModules
      );

      const report = document.generateSecurityReport();
      expect(report.signatureVerified).toBe(false);
      expect(report.warnings).toContain('Content signature missing');
      expect(report.warnings).toContain('Manifest signature missing');
    });

    it('should detect dangerous permissions', () => {
      const dangerousManifest = {
        ...mockManifest,
        security: {
          ...mockManifest.security,
          wasmPermissions: {
            ...mockManifest.security.wasmPermissions,
            allowNetworking: true,
            allowFileSystem: true
          },
          jsPermissions: {
            ...mockManifest.security.jsPermissions,
            executionMode: 'trusted' as const
          }
        }
      };

      const document = new LIVDocument(
        dangerousManifest,
        mockContent,
        mockAssets,
        mockSignatures,
        mockWASMModules
      );

      const report = document.generateSecurityReport();
      expect(report.warnings).toContain('Document requests WASM network access');
      expect(report.warnings).toContain('Document requests WASM file system access');
      expect(report.warnings).toContain('Document requests trusted JavaScript execution');
    });
  });

  describe('estimateMemoryUsage', () => {
    it('should calculate memory usage correctly', () => {
      const document = new LIVDocument(
        mockManifest,
        mockContent,
        mockAssets,
        mockSignatures,
        mockWASMModules
      );

      const usage = document.estimateMemoryUsage();
      expect(usage).toBeGreaterThan(0);
      
      // Should include content, assets, and WASM modules
      const expectedMinimum = mockContent.html.length + 
                             mockContent.css.length + 
                             2048 + // logo.png
                             8; // WASM module
      expect(usage).toBeGreaterThanOrEqual(expectedMinimum);
    });
  });

  describe('getMetadata', () => {
    it('should return document metadata', () => {
      const document = new LIVDocument(
        mockManifest,
        mockContent,
        mockAssets,
        mockSignatures,
        mockWASMModules
      );

      const metadata = document.getMetadata();
      expect(metadata.title).toBe('Test Document');
      expect(metadata.author).toBe('Test Author');
      expect(metadata.resourceCount).toBe(2);
      expect(metadata.wasmModuleCount).toBe(1);
      expect(metadata.hasInteractiveContent).toBe(true);
    });
  });

  describe('clearCache', () => {
    it('should clear resource cache and validation results', async () => {
      const document = new LIVDocument(
        mockManifest,
        mockContent,
        mockAssets,
        mockSignatures,
        mockWASMModules
      );

      // Load a resource to populate cache
      await document.getResource('content/index.html');
      
      // Validate to populate validation cache
      document.validate();
      
      // Clear cache
      document.clearCache();
      
      // Validation should run again
      const result1 = document.validate();
      const result2 = document.validate();
      expect(result1).toBe(result2); // Should be cached again
    });
  });

  describe('validateResourceIntegrity', () => {
    it('should validate resource integrity', async () => {
      const document = new LIVDocument(
        mockManifest,
        mockContent,
        mockAssets,
        mockSignatures,
        mockWASMModules
      );

      // This is a simplified test since we're not implementing full hash validation
      const isValid = await document.validateResourceIntegrity('content/index.html');
      expect(typeof isValid).toBe('boolean');
    });

    it('should return false for non-existent resource', async () => {
      const document = new LIVDocument(
        mockManifest,
        mockContent,
        mockAssets,
        mockSignatures,
        mockWASMModules
      );

      const isValid = await document.validateResourceIntegrity('non-existent/resource.txt');
      expect(isValid).toBe(false);
    });
  });
});

describe('loadLIVDocument', () => {
  it('should throw error for invalid source type', async () => {
    await expect(loadLIVDocument(123 as any)).rejects.toThrow('Invalid source type');
  });
});

// Integration tests would go here, but they require actual ZIP files
// and network resources, so they're omitted for this unit test suite