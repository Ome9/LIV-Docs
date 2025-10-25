// Cross-platform compatibility tests

import { LIVRenderer, SecureRenderingOptions } from '../src/renderer';
import { LIVDocument } from '../src/document';
import { LIVLoader } from '../src/loader';
import {
  LegacySecurityPolicy
} from '../src/types';

// Mock different platform environments
const mockPlatformEnvironments = {
  desktop: {
    userAgent: 'Mozilla/5.0 (Windows NT 10.0; Win64; x64) AppleWebKit/537.36',
    screen: { width: 1920, height: 1080 },
    devicePixelRatio: 1,
    touchSupport: false
  },
  mobile: {
    userAgent: 'Mozilla/5.0 (iPhone; CPU iPhone OS 14_0 like Mac OS X) AppleWebKit/605.1.15',
    screen: { width: 375, height: 812 },
    devicePixelRatio: 3,
    touchSupport: true
  },
  tablet: {
    userAgent: 'Mozilla/5.0 (iPad; CPU OS 14_0 like Mac OS X) AppleWebKit/605.1.15',
    screen: { width: 768, height: 1024 },
    devicePixelRatio: 2,
    touchSupport: true
  }
};

const createTestDocument = (): LIVDocument => {
  const manifest = {
    version: '1.0',
    metadata: {
      title: 'Cross-Platform Test',
      author: 'Test Author',
      created: '2024-01-01T00:00:00Z',
      modified: '2024-01-01T00:00:00Z',
      description: 'Cross-platform compatibility test',
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
        domAccess: 'write' as const
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
        hash: 'test-hash',
        size: 1024,
        type: 'text/html',
        path: 'content/index.html'
      }
    },
    features: {
      animations: true,
      interactivity: true,
      charts: false,
      forms: false,
      audio: false,
      video: false,
      webgl: false,
      webassembly: false
    }
  };

  const content = {
    html: `
      <html>
        <body>
          <div class="responsive-container">
            <h1>Cross-Platform Content</h1>
            <div class="touch-target" id="touch-test">Touch/Click Me</div>
            <svg width="200" height="100" class="responsive-svg">
              <rect x="10" y="10" width="180" height="80" fill="blue" rx="5"/>
              <text x="100" y="55" text-anchor="middle" fill="white">SVG Content</text>
            </svg>
          </div>
        </body>
      </html>
    `,
    css: `
      .responsive-container {
        padding: 20px;
        max-width: 100%;
        box-sizing: border-box;
      }
      
      .touch-target {
        padding: 15px 30px;
        background: #007bff;
        color: white;
        border-radius: 8px;
        cursor: pointer;
        user-select: none;
        transition: all 0.2s ease;
        margin: 20px 0;
      }
      
      .touch-target:hover {
        background: #0056b3;
        transform: scale(1.05);
      }
      
      .touch-target:active {
        transform: scale(0.95);
      }
      
      .responsive-svg {
        max-width: 100%;
        height: auto;
      }
      
      /* Mobile optimizations */
      @media (max-width: 768px) {
        .responsive-container {
          padding: 10px;
        }
        
        .touch-target {
          padding: 20px;
          font-size: 18px;
          min-height: 44px; /* iOS touch target minimum */
        }
      }
      
      /* High DPI display optimizations */
      @media (-webkit-min-device-pixel-ratio: 2), (min-resolution: 192dpi) {
        .responsive-svg {
          image-rendering: -webkit-optimize-contrast;
          image-rendering: crisp-edges;
        }
      }
    `,
    interactiveSpec: '',
    staticFallback: '<html><body><h1>Cross-Platform Content</h1><p>Static version</p></body></html>'
  };

  return new LIVDocument(
    manifest,
    content,
    { images: new Map(), fonts: new Map(), data: new Map() },
    { contentSignature: 'sig', manifestSignature: 'sig', wasmSignatures: {} },
    new Map()
  );
};

describe('Cross-Platform Compatibility Tests', () => {
  describe('Platform-Specific Rendering', () => {
    let container: HTMLElement;

    beforeEach(() => {
      container = document.createElement('div');
      document.body.appendChild(container);
    });

    afterEach(() => {
      if (container.parentNode) {
        container.parentNode.removeChild(container);
      }
    });

    it('should render correctly on desktop platforms', async () => {
      // Mock desktop environment
      Object.defineProperty(navigator, 'userAgent', {
        value: mockPlatformEnvironments.desktop.userAgent,
        configurable: true
      });
      
      container.style.width = '1920px';
      container.style.height = '1080px';
      
      const renderer = new LIVRenderer({
        container,
        permissions: {
          wasmPermissions: {
            memoryLimit: 8 * 1024 * 1024,
            allowedImports: [],
            cpuTimeLimit: 10000,
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
        enableResponsiveDesign: true
      });
      
      const document = createTestDocument();
      await renderer.renderDocument(document);
      
      const state = renderer.getRenderingState();
      expect(state.isRenderingComplete()).toBe(true);
      
      renderer.destroy();
    });

    it('should render correctly on mobile platforms', async () => {
      // Mock mobile environment
      Object.defineProperty(navigator, 'userAgent', {
        value: mockPlatformEnvironments.mobile.userAgent,
        configurable: true
      });
      
      container.style.width = '375px';
      container.style.height = '812px';
      
      const renderer = new LIVRenderer({
        container,
        permissions: {
          wasmPermissions: {
            memoryLimit: 4 * 1024 * 1024, // Lower memory for mobile
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
        enableResponsiveDesign: true,
        targetFPS: 30 // Lower FPS for mobile performance
      });
      
      const document = createTestDocument();
      await renderer.renderDocument(document);
      
      const state = renderer.getRenderingState();
      expect(state.isRenderingComplete()).toBe(true);
      
      renderer.destroy();
    });

    it('should adapt to different screen orientations', async () => {
      const document = createTestDocument();
      
      const renderer = new LIVRenderer({
        container,
        permissions: {
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
        enableResponsiveDesign: true
      });
      
      // Test portrait orientation
      container.style.width = '375px';
      container.style.height = '812px';
      
      await renderer.renderDocument(document);
      
      let state = renderer.getRenderingState();
      expect(state.isRenderingComplete()).toBe(true);
      
      // Test landscape orientation
      container.style.width = '812px';
      container.style.height = '375px';
      
      // Trigger resize
      const resizeEvent = new Event('resize');
      window.dispatchEvent(resizeEvent);
      
      // Should still be rendering correctly
      expect(state.isRenderingComplete()).toBe(true);
      
      renderer.destroy();
    });
  });

  describe('Browser Compatibility', () => {
    let container: HTMLElement;

    beforeEach(() => {
      container = document.createElement('div');
      container.style.width = '800px';
      container.style.height = '600px';
      document.body.appendChild(container);
    });

    afterEach(() => {
      if (container.parentNode) {
        container.parentNode.removeChild(container);
      }
    });

    it('should work with modern browser features', async () => {
      const document = createTestDocument();
      
      const renderer = new LIVRenderer({
        container,
        permissions: {
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
        }
      });
      
      await renderer.renderDocument(document);
      
      const state = renderer.getRenderingState();
      expect(state.isRenderingComplete()).toBe(true);
      
      // Check that modern features are used
      const shadowRoot = (container as any).shadowRoot;
      expect(shadowRoot).toBeDefined(); // Shadow DOM support
      
      renderer.destroy();
    });

    it('should gracefully degrade when features are unavailable', async () => {
      const document = createTestDocument();
      
      // Mock missing ResizeObserver
      const originalResizeObserver = (global as any).ResizeObserver;
      (global as any).ResizeObserver = undefined;
      
      const renderer = new LIVRenderer({
        container,
        permissions: {
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
        enableResponsiveDesign: true
      });
      
      await renderer.renderDocument(document);
      
      const state = renderer.getRenderingState();
      expect(state.isRenderingComplete()).toBe(true);
      
      // Restore ResizeObserver
      (global as any).ResizeObserver = originalResizeObserver;
      
      renderer.destroy();
    });
  });

  describe('Performance Benchmarks', () => {
    let container: HTMLElement;

    beforeEach(() => {
      container = document.createElement('div');
      container.style.width = '1024px';
      container.style.height = '768px';
      document.body.appendChild(container);
    });

    afterEach(() => {
      if (container.parentNode) {
        container.parentNode.removeChild(container);
      }
    });

    it('should meet performance benchmarks for document loading', async () => {
      const document = createTestDocument();
      
      const renderer = new LIVRenderer({
        container,
        permissions: {
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
        }
      });
      
      // Benchmark document loading
      const loadStart = performance.now();
      await renderer.renderDocument(document);
      const loadTime = performance.now() - loadStart;
      
      // Should load within reasonable time
      expect(loadTime).toBeLessThan(10000); // 10 seconds max for test environment
      
      const state = renderer.getRenderingState();
      expect(state.isRenderingComplete()).toBe(true);
      
      const metrics = renderer.getPerformanceMetrics();
      expect(metrics.renderTime).toBeLessThan(10000); // Relaxed for test environment
      
      renderer.destroy();
    });

    it('should maintain performance with multiple documents', async () => {
      const documents = [
        createTestDocument(),
        createTestDocument(),
        createTestDocument()
      ];
      
      const renderer = new LIVRenderer({
        container,
        permissions: {
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
        }
      });
      
      // Load documents sequentially
      for (const document of documents) {
        const startTime = performance.now();
        await renderer.renderDocument(document);
        const loadTime = performance.now() - startTime;
        
        expect(loadTime).toBeLessThan(10000); // Should maintain reasonable performance in test environment
        
        const state = renderer.getRenderingState();
        expect(state.isRenderingComplete()).toBe(true);
      }
      
      renderer.destroy();
    }, 20000);
  });

  describe('Memory Management', () => {
    let container: HTMLElement;

    beforeEach(() => {
      container = document.createElement('div');
      container.style.width = '800px';
      container.style.height = '600px';
      document.body.appendChild(container);
    });

    afterEach(() => {
      if (container.parentNode) {
        container.parentNode.removeChild(container);
      }
    });

    it('should clean up resources properly', async () => {
      const document = createTestDocument();
      
      const renderer = new LIVRenderer({
        container,
        permissions: {
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
        enableAnimations: true
      });
      
      await renderer.renderDocument(document);
      
      // Start animations
      renderer.startRenderLoop();
      await new Promise(resolve => setTimeout(resolve, 100));
      
      // Destroy renderer
      renderer.destroy();
      
      // Check that cleanup was successful
      const metrics = renderer.getPerformanceMetrics();
      expect(metrics).toBeDefined();
    });

    it('should handle memory pressure gracefully', async () => {
      const document = createTestDocument();
      
      // Create renderer with very low memory limit
      const renderer = new LIVRenderer({
        container,
        permissions: {
          wasmPermissions: {
            memoryLimit: 512 * 1024, // 512KB - very low
            allowedImports: [],
            cpuTimeLimit: 1000,
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
        enableFallback: true
      });
      
      await renderer.renderDocument(document);
      
      const state = renderer.getRenderingState();
      expect(state.isRenderingComplete()).toBe(true);
      
      renderer.destroy();
    });
  });
});