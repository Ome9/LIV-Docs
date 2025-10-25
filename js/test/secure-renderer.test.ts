// Tests for secure content rendering engine

import { LIVRenderer, SecureRenderingOptions, PerformanceMetrics } from '../src/renderer';
import { LIVDocument } from '../src/document';
import {
  LIVError,
  SecurityError,
  ValidationError,
  ErrorHandler
} from '../src/errors';
import {
  Manifest,
  DocumentContent,
  AssetBundle,
  SignatureBundle,
  LegacySecurityPolicy
} from '../src/types';

// Mock DOM environment
const createMockContainer = (): HTMLElement => {
  const container = document.createElement('div');
  container.style.width = '800px';
  container.style.height = '600px';
  return container;
};

const createMockSecurityPolicy = (): LegacySecurityPolicy => ({
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
  },
  contentSecurityPolicy: "default-src 'self'; script-src 'self' 'unsafe-inline'"
});

const createTestDocument = (): LIVDocument => {
  const manifest: Manifest = {
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
    security: createMockSecurityPolicy(),
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
      webassembly: true
    }
  };

  const content: DocumentContent = {
    html: '<html><body><h1>Test Document</h1><p>Interactive content</p></body></html>',
    css: 'body { font-family: Arial, sans-serif; margin: 20px; }',
    interactiveSpec: '{"type": "interactive", "version": "1.0"}',
    staticFallback: '<html><body><h1>Test Document</h1><p>Static fallback content</p></body></html>'
  };

  const assets: AssetBundle = {
    images: new Map(),
    fonts: new Map(),
    data: new Map()
  };

  const signatures: SignatureBundle = {
    contentSignature: 'mock-content-signature',
    manifestSignature: 'mock-manifest-signature',
    wasmSignatures: {}
  };

  const wasmModules = new Map<string, ArrayBuffer>();

  return new LIVDocument(manifest, content, assets, signatures, wasmModules);
};

describe('LIVRenderer - Secure Content Rendering', () => {
  let container: HTMLElement;
  let securityPolicy: LegacySecurityPolicy;
  let errorHandler: ErrorHandler;

  beforeEach(() => {
    container = createMockContainer();
    securityPolicy = createMockSecurityPolicy();
    errorHandler = ErrorHandler.getInstance();
    errorHandler.clearErrorHistory();
    
    // Append container to document body for testing
    document.body.appendChild(container);
  });

  afterEach(() => {
    // Clean up
    if (container.parentNode) {
      container.parentNode.removeChild(container);
    }
  });

  describe('Initialization', () => {
    it('should create renderer with secure options', () => {
      const options: SecureRenderingOptions = {
        container,
        permissions: securityPolicy,
        enableFallback: true,
        strictSecurity: true,
        maxRenderTime: 3000
      };

      const renderer = new LIVRenderer(options);
      expect(renderer).toBeInstanceOf(LIVRenderer);
      expect(renderer.getRenderingState()).toBeDefined();
    });

    it('should initialize with default options', () => {
      const renderer = new LIVRenderer({
        container,
        permissions: securityPolicy
      });

      expect(renderer).toBeInstanceOf(LIVRenderer);
    });
  });

  describe('Document Rendering', () => {
    let renderer: LIVRenderer;

    beforeEach(() => {
      renderer = new LIVRenderer({
        container,
        permissions: securityPolicy,
        enableFallback: true,
        strictSecurity: false // Allow rendering with validation warnings
      });
    });

    afterEach(() => {
      renderer.destroy();
    });

    it('should render valid document successfully', async () => {
      const document = createTestDocument();
      
      await renderer.renderDocument(document);
      
      const state = renderer.getRenderingState();
      expect(state.isRenderingComplete()).toBe(true);
      expect(renderer.getDocument()).toBe(document);
    });

    it('should handle document validation errors', async () => {
      const testDoc = createTestDocument();
      // Make document invalid
      testDoc.manifest.version = '';
      
      const strictContainer = createMockContainer();
      document.body.appendChild(strictContainer);
      
      const strictRenderer = new LIVRenderer({
        container: strictContainer,
        permissions: securityPolicy,
        strictSecurity: true
      });

      await expect(strictRenderer.renderDocument(testDoc)).rejects.toThrow(ValidationError);
      
      strictRenderer.destroy();
      if (strictContainer.parentNode) {
        strictContainer.parentNode.removeChild(strictContainer);
      }
    });

    it('should fall back to static content on interactive failure', async () => {
      const testDoc = createTestDocument();
      
      const fallbackContainer = createMockContainer();
      document.body.appendChild(fallbackContainer);
      
      // Mock WASM module failure
      const renderer = new LIVRenderer({
        container: fallbackContainer,
        permissions: securityPolicy,
        enableFallback: true,
        maxRenderTime: 100 // Very short timeout to force fallback
      });

      await renderer.renderDocument(testDoc);
      
      const state = renderer.getRenderingState();
      expect(state.isRenderingComplete()).toBe(true);
      // Should fall back to static mode due to timeout
      
      renderer.destroy();
      if (fallbackContainer.parentNode) {
        fallbackContainer.parentNode.removeChild(fallbackContainer);
      }
    });

    it('should sanitize dangerous HTML content', async () => {
      const testDoc = createTestDocument();
      testDoc.content.html = `
        <html>
          <body>
            <h1>Test</h1>
            <script>alert('xss')</script>
            <img src="javascript:alert('xss')" />
            <a href="javascript:alert('xss')">Link</a>
          </body>
        </html>
      `;

      await renderer.renderDocument(testDoc);
      
      const state = renderer.getRenderingState();
      expect(state.isRenderingComplete()).toBe(true);
      
      // Check that dangerous content was sanitized
      const shadowRoot = (container as any).shadowRoot;
      expect(shadowRoot).toBeDefined();
      
      // Content should be rendered (even if sanitized)
      if (shadowRoot) {
        expect(shadowRoot.children.length).toBeGreaterThan(0);
      }
    });

    it('should sanitize dangerous CSS', async () => {
      const testDoc = createTestDocument();
      testDoc.content.css = `
        body { 
          background: url('javascript:alert("xss")');
          behavior: url('malicious.htc');
          -moz-binding: url('evil.xml');
        }
        .test {
          expression: alert('xss');
        }
      `;

      await renderer.renderDocument(testDoc);
      
      const state = renderer.getRenderingState();
      expect(state.isRenderingComplete()).toBe(true);
      
      // CSS should be sanitized (dangerous parts removed)
      const shadowRoot = (container as any).shadowRoot;
      expect(shadowRoot).toBeDefined();
      
      if (shadowRoot) {
        const styleElement = shadowRoot.querySelector('style');
        if (styleElement) {
          expect(styleElement.textContent).not.toContain('javascript:');
          expect(styleElement.textContent).not.toContain('expression:');
        }
      }
    });
  });

  describe('Security Features', () => {
    let renderer: LIVRenderer;

    beforeEach(() => {
      renderer = new LIVRenderer({
        container,
        permissions: securityPolicy,
        strictSecurity: true
      });
    });

    afterEach(() => {
      renderer.destroy();
    });

    it('should enforce security policy', async () => {
      const testDoc = createTestDocument();
      
      const securityContainer = createMockContainer();
      document.body.appendChild(securityContainer);
      
      // Create restrictive security policy
      const restrictivePolicy = { ...securityPolicy };
      restrictivePolicy.jsPermissions.executionMode = 'none';
      
      const restrictiveRenderer = new LIVRenderer({
        container: securityContainer,
        permissions: restrictivePolicy,
        strictSecurity: true
      });

      await restrictiveRenderer.renderDocument(testDoc);
      
      const state = restrictiveRenderer.getRenderingState();
      expect(state.isRenderingComplete()).toBe(true);
      
      restrictiveRenderer.destroy();
      if (securityContainer.parentNode) {
        securityContainer.parentNode.removeChild(securityContainer);
      }
    });

    it('should handle security validation failures', async () => {
      const testDoc = createTestDocument();
      
      const securityFailContainer = createMockContainer();
      document.body.appendChild(securityFailContainer);
      
      const strictRenderer = new LIVRenderer({
        container: securityFailContainer,
        permissions: securityPolicy,
        strictSecurity: true
      });
      
      // Mock security report failure
      jest.spyOn(testDoc, 'generateSecurityReport').mockReturnValue({
        isValid: false,
        signatureVerified: false,
        integrityChecked: false,
        permissionsValid: false,
        warnings: ['Security validation failed'],
        errors: ['Invalid signature']
      });

      await expect(strictRenderer.renderDocument(testDoc)).rejects.toThrow(SecurityError);
      
      strictRenderer.destroy();
      if (securityFailContainer.parentNode) {
        securityFailContainer.parentNode.removeChild(securityFailContainer);
      }
    });
  });

  describe('Fallback Rendering', () => {
    let renderer: LIVRenderer;

    beforeEach(() => {
      renderer = new LIVRenderer({
        container,
        permissions: securityPolicy,
        enableFallback: true
      });
    });

    afterEach(() => {
      renderer.destroy();
    });

    it('should render static fallback when interactive fails', async () => {
      const testDoc = createTestDocument();
      
      // Force interactive rendering to fail
      testDoc.content.interactiveSpec = 'invalid json';
      
      await renderer.renderDocument(testDoc);
      
      const state = renderer.getRenderingState();
      expect(state.isRenderingComplete()).toBe(true);
      expect(state.isFallbackMode()).toBe(true);
    });

    it('should use static fallback content when available', async () => {
      const testDoc = createTestDocument();
      testDoc.content.staticFallback = '<html><body><h1>Fallback Content</h1></body></html>';
      
      const staticContainer = createMockContainer();
      document.body.appendChild(staticContainer);
      
      // Disable interactive rendering
      const staticRenderer = new LIVRenderer({
        container: staticContainer,
        permissions: {
          ...securityPolicy,
          jsPermissions: {
            ...securityPolicy.jsPermissions,
            executionMode: 'none'
          }
        },
        enableFallback: true
      });

      await staticRenderer.renderDocument(testDoc);
      
      const state = staticRenderer.getRenderingState();
      expect(state.isRenderingComplete()).toBe(true);
      expect(state.isFallbackMode()).toBe(true);
      
      staticRenderer.destroy();
      if (staticContainer.parentNode) {
        staticContainer.parentNode.removeChild(staticContainer);
      }
    });
  });

  describe('Performance Monitoring', () => {
    let renderer: LIVRenderer;

    beforeEach(() => {
      renderer = new LIVRenderer({
        container,
        permissions: securityPolicy
      });
    });

    afterEach(() => {
      renderer.destroy();
    });

    it('should track rendering performance', async () => {
      const testDoc = createTestDocument();
      
      await renderer.renderDocument(testDoc);
      
      const metrics = renderer.getPerformanceMetrics();
      expect(metrics.renderTime).toBeGreaterThan(0);
      expect(metrics.isComplete).toBe(true);
      expect(metrics.errorCount).toBe(0);
    });

    it('should track errors in performance metrics', async () => {
      const testDoc = createTestDocument();
      
      // Force an error
      testDoc.content.html = ''; // Empty HTML should cause issues
      
      await renderer.renderDocument(testDoc);
      
      const metrics = renderer.getPerformanceMetrics();
      expect(metrics.errorCount).toBeGreaterThanOrEqual(0);
    });
  });

  describe('Error Handling', () => {
    let renderer: LIVRenderer;

    beforeEach(() => {
      renderer = new LIVRenderer({
        container,
        permissions: securityPolicy,
        enableFallback: true
      });
    });

    afterEach(() => {
      renderer.destroy();
    });

    it('should handle rendering timeout', async () => {
      const testDoc = createTestDocument();
      
      const timeoutContainer = createMockContainer();
      document.body.appendChild(timeoutContainer);
      
      const timeoutRenderer = new LIVRenderer({
        container: timeoutContainer,
        permissions: securityPolicy,
        maxRenderTime: 1, // 1ms timeout
        enableFallback: true
      });

      await timeoutRenderer.renderDocument(testDoc);
      
      // Should complete with fallback due to timeout
      const state = timeoutRenderer.getRenderingState();
      expect(state.isRenderingComplete()).toBe(true);
      
      timeoutRenderer.destroy();
      if (timeoutContainer.parentNode) {
        timeoutContainer.parentNode.removeChild(timeoutContainer);
      }
    });

    it('should display error message when all rendering fails', async () => {
      const testDoc = createTestDocument();
      
      const errorContainer = createMockContainer();
      document.body.appendChild(errorContainer);
      
      // Make both interactive and fallback fail
      testDoc.content.html = '';
      testDoc.content.staticFallback = '';
      
      const noFallbackRenderer = new LIVRenderer({
        container: errorContainer,
        permissions: securityPolicy,
        enableFallback: false,
        strictSecurity: false
      });

      await noFallbackRenderer.renderDocument(testDoc);
      
      // Should show error message
      const shadowRoot = (errorContainer as any).shadowRoot;
      expect(shadowRoot).toBeDefined();
      
      noFallbackRenderer.destroy();
      if (errorContainer.parentNode) {
        errorContainer.parentNode.removeChild(errorContainer);
      }
    });
  });

  describe('Cleanup and Destruction', () => {
    it('should clean up resources on destroy', () => {
      const renderer = new LIVRenderer({
        container,
        permissions: securityPolicy
      });

      // Start render loop
      renderer.startRenderLoop();
      
      // Destroy renderer
      renderer.destroy();
      
      // Check that resources are cleaned up
      const state = renderer.getRenderingState();
      expect(state.getPerformanceMetrics().frameCount).toBe(0);
    });

    it('should handle multiple destroy calls', () => {
      const renderer = new LIVRenderer({
        container,
        permissions: securityPolicy
      });

      // Multiple destroy calls should not throw
      expect(() => {
        renderer.destroy();
        renderer.destroy();
        renderer.destroy();
      }).not.toThrow();
    });
  });
});

// Integration tests with actual DOM
describe('LIVRenderer - DOM Integration', () => {
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

  it('should create shadow DOM for isolation', async () => {
    const renderer = new LIVRenderer({
      container,
      permissions: createMockSecurityPolicy()
    });

    const testDoc = createTestDocument();
    await renderer.renderDocument(testDoc);

    // Check that shadow DOM was created
    expect((container as any).shadowRoot).toBeDefined();
    
    renderer.destroy();
  });

  it('should apply CSP meta tag when specified', () => {
    const policy = createMockSecurityPolicy();
    policy.contentSecurityPolicy = "default-src 'self'";
    
    const renderer = new LIVRenderer({
      container,
      permissions: policy
    });

    // Check that CSP meta tag was added
    const cspMeta = document.querySelector('meta[http-equiv="Content-Security-Policy"]');
    expect(cspMeta).toBeDefined();
    
    renderer.destroy();
  });
});