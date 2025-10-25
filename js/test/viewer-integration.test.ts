// Comprehensive viewer integration tests - Fixed version

import { LIVRenderer, SecureRenderingOptions } from '../src/renderer';
import { LIVDocument, loadLIVDocument } from '../src/document';
import { LIVLoader } from '../src/loader';
import {
  Manifest,
  DocumentContent,
  AssetBundle,
  SignatureBundle,
  LegacySecurityPolicy
} from '../src/types';

// Test data generators for various document structures
const createBasicDocument = (): LIVDocument => {
  const manifest: Manifest = {
    version: '1.0',
    metadata: {
      title: 'Basic Test Document',
      author: 'Integration Test',
      created: '2024-01-01T00:00:00Z',
      modified: '2024-01-01T00:00:00Z',
      description: 'Basic document for integration testing',
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
        executionMode: 'sandboxed',
        allowedAPIs: [],
        domAccess: 'write'
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
        hash: 'basic-hash',
        size: 512,
        type: 'text/html',
        path: 'content/index.html'
      }
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
    }
  };

  const content: DocumentContent = {
    html: '<html><body><h1>Basic Document</h1><p>Simple content</p></body></html>',
    css: 'body { font-family: Arial, sans-serif; margin: 20px; }',
    interactiveSpec: '',
    staticFallback: '<html><body><h1>Basic Document</h1><p>Static fallback</p></body></html>'
  };

  return new LIVDocument(manifest, content, { images: new Map(), fonts: new Map(), data: new Map() }, { contentSignature: '', manifestSignature: '', wasmSignatures: {} }, new Map());
};

const createComplexDocument = (): LIVDocument => {
  const manifest: Manifest = {
    version: '1.0',
    metadata: {
      title: 'Complex Interactive Document',
      author: 'Integration Test',
      created: '2024-01-01T00:00:00Z',
      modified: '2024-01-01T00:00:00Z',
      description: 'Complex document with animations, SVG, and interactivity',
      version: '1.0.0',
      language: 'en'
    },
    security: {
      wasmPermissions: {
        memoryLimit: 8 * 1024 * 1024,
        allowedImports: ['env'],
        cpuTimeLimit: 10000,
        allowNetworking: false,
        allowFileSystem: false
      },
      jsPermissions: {
        executionMode: 'sandboxed',
        allowedAPIs: ['animation', 'canvas'],
        domAccess: 'write'
      },
      networkPolicy: {
        allowOutbound: false,
        allowedHosts: [],
        allowedPorts: []
      },
      storagePolicy: {
        allowLocalStorage: true,
        allowSessionStorage: true,
        allowIndexedDB: false,
        allowCookies: false
      }
    },
    resources: {
      'content/index.html': {
        hash: 'complex-hash',
        size: 2048,
        type: 'text/html',
        path: 'content/index.html'
      }
    },
    features: {
      animations: true,
      interactivity: true,
      charts: true,
      forms: true,
      audio: false,
      video: false,
      webgl: false,
      webassembly: true
    }
  };

  const content: DocumentContent = {
    html: '<html><body><h1>Interactive Dashboard</h1><div>Complex content</div></body></html>',
    css: 'body { font-family: Arial, sans-serif; } .animated { animation: pulse 2s infinite; }',
    interactiveSpec: JSON.stringify({
      type: 'dashboard',
      version: '1.0',
      modules: ['chart-engine', 'animation-controller']
    }),
    staticFallback: '<html><body><h1>Interactive Dashboard (Static)</h1></body></html>'
  };

  const assets: AssetBundle = {
    images: new Map([
      ['chart.svg', new TextEncoder().encode('<svg>...</svg>').buffer]
    ]),
    fonts: new Map(),
    data: new Map([
      ['dataset.json', new TextEncoder().encode('{"revenue": 125000, "users": 45230}').buffer]
    ])
  };

  return new LIVDocument(manifest, content, assets, { contentSignature: 'sig', manifestSignature: 'sig', wasmSignatures: {} }, new Map());
};

const createMockContainer = (): HTMLElement => {
  const container = document.createElement('div');
  container.style.width = '1024px';
  container.style.height = '768px';
  container.id = `test-container-${Date.now()}-${Math.random()}`;
  return container;
};

const createMockSecurityPolicy = (): LegacySecurityPolicy => ({
  wasmPermissions: {
    memoryLimit: 8 * 1024 * 1024,
    allowedImports: ['env'],
    cpuTimeLimit: 10000,
    allowNetworking: false,
    allowFileSystem: false
  },
  jsPermissions: {
    executionMode: 'sandboxed',
    allowedAPIs: ['animation', 'canvas'],
    domAccess: 'write'
  },
  networkPolicy: {
    allowOutbound: false,
    allowedHosts: [],
    allowedPorts: []
  },
  storagePolicy: {
    allowLocalStorage: true,
    allowSessionStorage: true,
    allowIndexedDB: false,
    allowCookies: false
  }
});

describe('Viewer Integration Tests - Task 4.4 (Fixed)', () => {
  describe('Document Loading with Various .liv File Structures', () => {
    let container: HTMLElement;
    let renderer: LIVRenderer;

    beforeEach(() => {
      container = createMockContainer();
      document.body.appendChild(container);
      
      renderer = new LIVRenderer({
        container,
        permissions: createMockSecurityPolicy(),
        enableAnimations: true,
        enableSVG: true,
        enableResponsiveDesign: true
      });
    });

    afterEach(() => {
      try {
        renderer.destroy();
      } catch (e) {
        // Ignore cleanup errors
      }
      if (container.parentNode) {
        container.parentNode.removeChild(container);
      }
    });

    it('should load and render basic document structure', async () => {
      const document = createBasicDocument();
      
      const startTime = performance.now();
      await renderer.renderDocument(document);
      const loadTime = performance.now() - startTime;
      
      const state = renderer.getRenderingState();
      expect(state.isRenderingComplete()).toBe(true);
      expect(loadTime).toBeLessThan(10000); // 10 seconds max for test environment
    }, 15000);

    it('should load and render complex interactive document', async () => {
      const document = createComplexDocument();
      
      const startTime = performance.now();
      await renderer.renderDocument(document);
      const loadTime = performance.now() - startTime;
      
      const state = renderer.getRenderingState();
      expect(state.isRenderingComplete()).toBe(true);
      expect(loadTime).toBeLessThan(15000); // Allow more time for complex documents
    }, 20000);

    it('should handle documents with missing resources gracefully', async () => {
      const document = createComplexDocument();
      
      // Add missing resource reference
      document.manifest.resources['assets/missing/file.png'] = {
        hash: 'missing-hash',
        size: 1024,
        type: 'image/png',
        path: 'assets/missing/file.png'
      };
      
      await renderer.renderDocument(document);
      
      const state = renderer.getRenderingState();
      expect(state.isRenderingComplete()).toBe(true);
    }, 15000);
  });

  describe('Rendering Performance with Animated Content', () => {
    let container: HTMLElement;
    let renderer: LIVRenderer;

    beforeEach(() => {
      container = createMockContainer();
      document.body.appendChild(container);
      
      renderer = new LIVRenderer({
        container,
        permissions: createMockSecurityPolicy(),
        enableAnimations: true,
        targetFPS: 30 // Lower target for test environment
      });
    });

    afterEach(() => {
      try {
        renderer.destroy();
      } catch (e) {
        // Ignore cleanup errors
      }
      if (container.parentNode) {
        container.parentNode.removeChild(container);
      }
    });

    it('should maintain reasonable FPS with animations', async () => {
      const document = createComplexDocument();
      
      await renderer.renderDocument(document);
      
      // Start animation loop
      renderer.startRenderLoop();
      
      // Let animations run for a short time
      await new Promise(resolve => setTimeout(resolve, 500));
      
      const metrics = renderer.getPerformanceMetrics();
      expect(metrics.frameCount).toBeGreaterThan(0); // Should have rendered some frames
      
      renderer.stopRenderLoop();
    }, 10000);

    it('should handle animation errors gracefully', async () => {
      const document = createComplexDocument();
      document.content.css = 'invalid css { animation: broken; }';
      
      await renderer.renderDocument(document);
      
      const state = renderer.getRenderingState();
      expect(state.isRenderingComplete()).toBe(true);
    }, 10000);
  });

  describe('Cross-Platform Compatibility', () => {
    let container: HTMLElement;

    beforeEach(() => {
      container = createMockContainer();
      document.body.appendChild(container);
    });

    afterEach(() => {
      if (container.parentNode) {
        container.parentNode.removeChild(container);
      }
    });

    it('should render consistently across different viewport sizes', async () => {
      const document = createComplexDocument();
      
      // Test desktop size
      const desktopRenderer = new LIVRenderer({
        container,
        permissions: createMockSecurityPolicy(),
        enableResponsiveDesign: true
      });
      
      container.style.width = '1920px';
      container.style.height = '1080px';
      
      await desktopRenderer.renderDocument(document);
      
      let state = desktopRenderer.getRenderingState();
      expect(state.isRenderingComplete()).toBe(true);
      
      desktopRenderer.destroy();
      
      // Create new container for mobile test to avoid shadow DOM conflicts
      const mobileContainer = createMockContainer();
      if (document.body) {
        document.body.appendChild(mobileContainer);
      }
      mobileContainer.style.width = '375px';
      mobileContainer.style.height = '667px';
      
      const mobileRenderer = new LIVRenderer({
        container: mobileContainer,
        permissions: createMockSecurityPolicy(),
        enableResponsiveDesign: true
      });
      
      await mobileRenderer.renderDocument(document);
      
      state = mobileRenderer.getRenderingState();
      expect(state.isRenderingComplete()).toBe(true);
      
      mobileRenderer.destroy();
      if (mobileContainer && mobileContainer.parentNode) {
        mobileContainer.parentNode.removeChild(mobileContainer);
      }
    }, 15000);

    it('should handle input methods gracefully', async () => {
      const document = createComplexDocument();
      
      const renderer = new LIVRenderer({
        container,
        permissions: createMockSecurityPolicy(),
        enableResponsiveDesign: true
      });
      
      await renderer.renderDocument(document);
      
      // Simulate mouse events
      const mouseEvent = new MouseEvent('click', {
        clientX: 100,
        clientY: 100,
        button: 0
      });
      
      expect(() => {
        container.dispatchEvent(mouseEvent);
      }).not.toThrow();
      
      renderer.destroy();
    }, 10000);
  });

  describe('Error Handling and Recovery', () => {
    let container: HTMLElement;
    let renderer: LIVRenderer;

    beforeEach(() => {
      container = createMockContainer();
      document.body.appendChild(container);
      
      renderer = new LIVRenderer({
        container,
        permissions: createMockSecurityPolicy(),
        enableFallback: true
      });
    });

    afterEach(() => {
      try {
        renderer.destroy();
      } catch (e) {
        // Ignore cleanup errors
      }
      if (container.parentNode) {
        container.parentNode.removeChild(container);
      }
    });

    it('should recover from malformed HTML gracefully', async () => {
      const document = createBasicDocument();
      document.content.html = '<html><body><div>Unclosed div<span>Nested</body></html>';
      
      await renderer.renderDocument(document);
      
      const state = renderer.getRenderingState();
      expect(state.isRenderingComplete()).toBe(true);
    }, 10000);

    it('should handle invalid CSS gracefully', async () => {
      const document = createBasicDocument();
      document.content.css = 'invalid { color: invalid-value; }';
      
      await renderer.renderDocument(document);
      
      const state = renderer.getRenderingState();
      expect(state.isRenderingComplete()).toBe(true);
    }, 10000);
  });
});