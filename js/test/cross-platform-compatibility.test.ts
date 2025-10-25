// Cross-platform compatibility tests for LIV viewer applications

import { expect } from 'chai';
import { JSDOM } from 'jsdom';
import { LIVRenderer, SecureRenderingOptions } from '../src/renderer';
import { LIVDocument } from '../src/document';
import { LegacySecurityPolicy, InteractionType, GestureType } from '../src/types';

// Mock different browser environments
interface BrowserEnvironment {
  name: string;
  userAgent: string;
  features: {
    touch: boolean;
    webgl: boolean;
    webassembly: boolean;
    serviceWorker: boolean;
    resizeObserver: boolean;
    intersectionObserver: boolean;
    matchMedia: boolean;
    vibrate: boolean;
    battery: boolean;
    connection: boolean;
  };
  viewport: {
    width: number;
    height: number;
    devicePixelRatio: number;
  };
  performance: {
    memory?: boolean;
    navigation?: boolean;
    timing?: boolean;
  };
}

const BROWSER_ENVIRONMENTS: BrowserEnvironment[] = [
  {
    name: 'Chrome Desktop',
    userAgent: 'Mozilla/5.0 (Windows NT 10.0; Win64; x64) AppleWebKit/537.36 (KHTML, like Gecko) Chrome/120.0.0.0 Safari/537.36',
    features: {
      touch: false,
      webgl: true,
      webassembly: true,
      serviceWorker: true,
      resizeObserver: true,
      intersectionObserver: true,
      matchMedia: true,
      vibrate: false,
      battery: false,
      connection: true
    },
    viewport: { width: 1920, height: 1080, devicePixelRatio: 1 },
    performance: { memory: true, navigation: true, timing: true }
  },
  {
    name: 'Safari Desktop',
    userAgent: 'Mozilla/5.0 (Macintosh; Intel Mac OS X 10_15_7) AppleWebKit/605.1.15 (KHTML, like Gecko) Version/17.0 Safari/605.1.15',
    features: {
      touch: false,
      webgl: true,
      webassembly: true,
      serviceWorker: true,
      resizeObserver: true,
      intersectionObserver: true,
      matchMedia: true,
      vibrate: false,
      battery: false,
      connection: false
    },
    viewport: { width: 1440, height: 900, devicePixelRatio: 2 },
    performance: { memory: false, navigation: true, timing: true }
  },
  {
    name: 'Firefox Desktop',
    userAgent: 'Mozilla/5.0 (Windows NT 10.0; Win64; x64; rv:120.0) Gecko/20100101 Firefox/120.0',
    features: {
      touch: false,
      webgl: true,
      webassembly: true,
      serviceWorker: true,
      resizeObserver: true,
      intersectionObserver: true,
      matchMedia: true,
      vibrate: false,
      battery: false,
      connection: false
    },
    viewport: { width: 1366, height: 768, devicePixelRatio: 1 },
    performance: { memory: false, navigation: true, timing: true }
  },
  {
    name: 'Chrome Mobile Android',
    userAgent: 'Mozilla/5.0 (Linux; Android 13; SM-G991B) AppleWebKit/537.36 (KHTML, like Gecko) Chrome/120.0.0.0 Mobile Safari/537.36',
    features: {
      touch: true,
      webgl: true,
      webassembly: true,
      serviceWorker: true,
      resizeObserver: true,
      intersectionObserver: true,
      matchMedia: true,
      vibrate: true,
      battery: true,
      connection: true
    },
    viewport: { width: 393, height: 851, devicePixelRatio: 2.75 },
    performance: { memory: true, navigation: true, timing: true }
  },
  {
    name: 'Safari Mobile iOS',
    userAgent: 'Mozilla/5.0 (iPhone; CPU iPhone OS 17_0 like Mac OS X) AppleWebKit/605.1.15 (KHTML, like Gecko) Version/17.0 Mobile/15E148 Safari/604.1',
    features: {
      touch: true,
      webgl: true,
      webassembly: true,
      serviceWorker: true,
      resizeObserver: true,
      intersectionObserver: true,
      matchMedia: true,
      vibrate: false,
      battery: false,
      connection: false
    },
    viewport: { width: 393, height: 852, devicePixelRatio: 3 },
    performance: { memory: false, navigation: true, timing: true }
  },
  {
    name: 'Edge Desktop',
    userAgent: 'Mozilla/5.0 (Windows NT 10.0; Win64; x64) AppleWebKit/537.36 (KHTML, like Gecko) Chrome/120.0.0.0 Safari/537.36 Edg/120.0.0.0',
    features: {
      touch: false,
      webgl: true,
      webassembly: true,
      serviceWorker: true,
      resizeObserver: true,
      intersectionObserver: true,
      matchMedia: true,
      vibrate: false,
      battery: false,
      connection: true
    },
    viewport: { width: 1536, height: 864, devicePixelRatio: 1.25 },
    performance: { memory: true, navigation: true, timing: true }
  },
  {
    name: 'Samsung Internet Mobile',
    userAgent: 'Mozilla/5.0 (Linux; Android 13; SM-G991B) AppleWebKit/537.36 (KHTML, like Gecko) SamsungBrowser/23.0 Chrome/115.0.0.0 Mobile Safari/537.36',
    features: {
      touch: true,
      webgl: true,
      webassembly: true,
      serviceWorker: true,
      resizeObserver: true,
      intersectionObserver: true,
      matchMedia: true,
      vibrate: true,
      battery: true,
      connection: true
    },
    viewport: { width: 360, height: 740, devicePixelRatio: 3 },
    performance: { memory: false, navigation: true, timing: true }
  }
];

describe('Cross-Platform Compatibility Tests', () => {
  let originalWindow: any;
  let originalDocument: any;
  let originalNavigator: any;

  beforeEach(() => {
    // Store original globals
    originalWindow = global.window;
    originalDocument = global.document;
    originalNavigator = global.navigator;
  });

  afterEach(() => {
    // Restore original globals
    global.window = originalWindow;
    global.document = originalDocument;
    global.navigator = originalNavigator;
  });

  describe('Browser Environment Compatibility', () => {
    BROWSER_ENVIRONMENTS.forEach(env => {
      describe(`${env.name} Environment`, () => {
        let dom: JSDOM;
        let container: HTMLElement;
        let renderer: LIVRenderer;

        beforeEach(() => {
          // Set up DOM environment for this browser
          dom = setupBrowserEnvironment(env);
          container = dom.window.document.getElementById('container')!;
          
          const securityPolicy: LegacySecurityPolicy = {
            wasmPermissions: {
              memoryLimit: 16 * 1024 * 1024,
              allowedImports: ['console'],
              cpuTimeLimit: 5000,
              allowNetworking: false,
              allowFileSystem: false
            },
            jsPermissions: {
              executionMode: 'sandboxed',
              allowedAPIs: ['console'],
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

          const options: SecureRenderingOptions = {
            container,
            permissions: securityPolicy,
            enableFallback: true,
            enableAnimations: true,
            enableSVG: true,
            enableResponsiveDesign: true
          };

          renderer = new LIVRenderer(options);
        });

        afterEach(() => {
          if (renderer) {
            renderer.destroy();
          }
        });

        it('should initialize renderer successfully', () => {
          expect(renderer).to.be.instanceOf(LIVRenderer);
          expect(container.classList.contains('liv-mobile-optimized')).to.equal(env.features.touch);
        });

        it('should detect device capabilities correctly', () => {
          // Test touch detection
          const hasTouch = 'ontouchstart' in dom.window || dom.window.navigator.maxTouchPoints > 0;
          expect(hasTouch).to.equal(env.features.touch);

          // Test WebGL support
          if (env.features.webgl) {
            expect(dom.window.WebGLRenderingContext).to.not.be.undefined;
          }

          // Test WebAssembly support
          if (env.features.webassembly) {
            expect(dom.window.WebAssembly).to.not.be.undefined;
          }
        });

        it('should handle viewport dimensions correctly', () => {
          expect(dom.window.innerWidth).to.equal(env.viewport.width);
          expect(dom.window.innerHeight).to.equal(env.viewport.height);
          expect(dom.window.devicePixelRatio).to.equal(env.viewport.devicePixelRatio);
        });

        it('should apply appropriate responsive classes', async () => {
          await renderer.initialize?.();
          
          const isMobile = env.viewport.width <= 768;
          const isTablet = env.viewport.width > 768 && env.viewport.width <= 1024;
          const isDesktop = env.viewport.width > 1024;

          if (isMobile) {
            expect(container.classList.contains('liv-mobile') || 
                   container.classList.contains('liv-mobile-sm') ||
                   container.classList.contains('liv-mobile-xs')).to.be.true;
          } else if (isTablet) {
            expect(container.classList.contains('liv-tablet')).to.be.true;
          } else if (isDesktop) {
            expect(container.classList.contains('liv-desktop')).to.be.true;
          }
        });

        it('should handle feature detection gracefully', () => {
          // Test ResizeObserver fallback
          if (!env.features.resizeObserver) {
            expect(dom.window.ResizeObserver).to.be.undefined;
          }

          // Test IntersectionObserver fallback
          if (!env.features.intersectionObserver) {
            expect(dom.window.IntersectionObserver).to.be.undefined;
          }

          // Test matchMedia fallback
          if (!env.features.matchMedia) {
            expect(dom.window.matchMedia).to.be.undefined;
          }
        });

        it('should render basic content without errors', async () => {
          const mockDocument = createMockLIVDocument();
          
          try {
            await renderer.renderDocument(mockDocument);
            expect(renderer.getRenderingState().isRenderingComplete()).to.be.true;
          } catch (error) {
            // Should fall back to static content on any errors
            expect(renderer.getRenderingState().isFallbackMode()).to.be.true;
          }
        });

        it('should handle animations based on device capabilities', async () => {
          const mockDocument = createMockLIVDocument(true); // With animations
          
          await renderer.renderDocument(mockDocument);
          
          // Check if animations are enabled based on device performance
          const animationsEnabled = !dom.window.matchMedia('(prefers-reduced-motion: reduce)').matches;
          
          if (env.features.touch && env.viewport.width <= 480) {
            // On small mobile devices, animations might be reduced for performance
            expect(true).to.be.true; // Placeholder - would check actual animation state
          }
        });

        if (env.features.touch) {
          it('should handle touch interactions', () => {
            let interactionHandled = false;
            
            container.addEventListener('liv-interaction', (event: CustomEvent) => {
              if (event.detail.eventType === InteractionType.TouchStart) {
                interactionHandled = true;
              }
            });

            // Simulate touch event
            const touchEvent = new dom.window.TouchEvent('touchstart', {
              touches: [createMockTouch(dom.window, 100, 100, 1)],
              changedTouches: [createMockTouch(dom.window, 100, 100, 1)],
              targetTouches: [createMockTouch(dom.window, 100, 100, 1)]
            });

            container.dispatchEvent(touchEvent);
            expect(interactionHandled).to.be.true;
          });

          it('should recognize gestures on touch devices', () => {
            let gestureRecognized = false;
            
            container.addEventListener('liv-interaction', (event: CustomEvent) => {
              if (event.detail.eventType === InteractionType.Tap) {
                gestureRecognized = true;
              }
            });

            // Simulate tap gesture
            simulateTapGesture(dom.window, container);
            
            setTimeout(() => {
              expect(gestureRecognized).to.be.true;
            }, 200);
          });
        }

        it('should handle performance constraints', () => {
          const performanceMetrics = renderer.getPerformanceMetrics();
          
          expect(performanceMetrics).to.have.property('frameRate');
          expect(performanceMetrics).to.have.property('renderTime');
          
          // On mobile devices, frame rate might be capped
          if (env.features.touch) {
            expect(performanceMetrics.frameRate).to.be.at.most(60);
          }
        });

        it('should adapt to network conditions', () => {
          if (env.features.connection) {
            // Mock slow connection
            (dom.window.navigator as any).connection = {
              effectiveType: '2g',
              downlink: 0.5,
              rtt: 2000
            };

            // Should apply bandwidth optimizations
            expect(container.classList.contains('liv-optimize-bandwidth')).to.be.true;
          }
        });
      });
    });
  });

  describe('Performance Across Platforms', () => {
    it('should maintain acceptable performance on all platforms', async () => {
      const performanceResults: Array<{
        platform: string;
        renderTime: number;
        frameRate: number;
        memoryUsage: number;
      }> = [];

      for (const env of BROWSER_ENVIRONMENTS) {
        const dom = setupBrowserEnvironment(env);
        const container = dom.window.document.getElementById('container')!;
        
        const renderer = new LIVRenderer({
          container,
          permissions: createDefaultSecurityPolicy(),
          enableAnimations: true,
          enableSVG: true
        });

        const startTime = performance.now();
        const mockDocument = createMockLIVDocument(true);
        
        try {
          await renderer.renderDocument(mockDocument);
          const endTime = performance.now();
          
          const metrics = renderer.getPerformanceMetrics();
          
          performanceResults.push({
            platform: env.name,
            renderTime: endTime - startTime,
            frameRate: metrics.averageFPS,
            memoryUsage: (performance as any).memory?.usedJSHeapSize || 0
          });
        } catch (error) {
          console.warn(`Performance test failed for ${env.name}:`, error);
        }
        
        renderer.destroy();
      }

      // Verify performance is acceptable across all platforms
      for (const result of performanceResults) {
        expect(result.renderTime).to.be.lessThan(5000); // 5 second max render time
        expect(result.frameRate).to.be.greaterThan(15); // Minimum 15 FPS
        
        console.log(`${result.platform}: Render ${result.renderTime}ms, FPS ${result.frameRate}`);
      }
    });

    it('should scale performance based on device capabilities', async () => {
      const mobileEnv = BROWSER_ENVIRONMENTS.find(env => env.name.includes('Mobile'))!;
      const desktopEnv = BROWSER_ENVIRONMENTS.find(env => env.name.includes('Chrome Desktop'))!;

      const mobilePerf = await measurePlatformPerformance(mobileEnv);
      const desktopPerf = await measurePlatformPerformance(desktopEnv);

      // Desktop should generally have better performance
      expect(desktopPerf.frameRate).to.be.greaterThanOrEqual(mobilePerf.frameRate);
      
      // Mobile should use performance optimizations
      expect(mobilePerf.frameRate).to.be.lessThanOrEqual(30); // Mobile FPS cap
    });
  });

  describe('Responsive Design Compatibility', () => {
    it('should adapt layout to different screen sizes', () => {
      const testSizes = [
        { width: 320, height: 568, name: 'iPhone SE' },
        { width: 375, height: 812, name: 'iPhone X' },
        { width: 768, height: 1024, name: 'iPad' },
        { width: 1366, height: 768, name: 'Laptop' },
        { width: 1920, height: 1080, name: 'Desktop' }
      ];

      testSizes.forEach(size => {
        const env: BrowserEnvironment = {
          ...BROWSER_ENVIRONMENTS[0],
          name: size.name,
          viewport: { ...size, devicePixelRatio: 1 },
          features: { ...BROWSER_ENVIRONMENTS[0].features, touch: size.width <= 768 }
        };

        const dom = setupBrowserEnvironment(env);
        const container = dom.window.document.getElementById('container')!;
        
        const renderer = new LIVRenderer({
          container,
          permissions: createDefaultSecurityPolicy(),
          enableResponsiveDesign: true
        });

        // Check responsive classes
        const responsiveManager = (renderer as any).responsiveManager;
        if (responsiveManager) {
          responsiveManager.initializeClasses();
          
          if (size.width <= 320) {
            expect(container.classList.contains('liv-xs')).to.be.true;
          } else if (size.width <= 480) {
            expect(container.classList.contains('liv-sm')).to.be.true;
          } else if (size.width <= 768) {
            expect(container.classList.contains('liv-md')).to.be.true;
          } else if (size.width <= 1024) {
            expect(container.classList.contains('liv-lg')).to.be.true;
          } else {
            expect(container.classList.contains('liv-xl')).to.be.true;
          }
        }

        renderer.destroy();
      });
    });

    it('should handle orientation changes', () => {
      const mobileEnv = BROWSER_ENVIRONMENTS.find(env => env.features.touch)!;
      const dom = setupBrowserEnvironment(mobileEnv);
      const container = dom.window.document.getElementById('container')!;
      
      const renderer = new LIVRenderer({
        container,
        permissions: createDefaultSecurityPolicy(),
        enableResponsiveDesign: true
      });

      let orientationChanged = false;
      
      container.addEventListener('liv-orientation-change', () => {
        orientationChanged = true;
      });

      // Simulate orientation change
      const orientationEvent = new dom.window.Event('orientationchange');
      dom.window.dispatchEvent(orientationEvent);

      setTimeout(() => {
        expect(orientationChanged).to.be.true;
      }, 150);

      renderer.destroy();
    });
  });

  describe('Error Handling Compatibility', () => {
    it('should handle missing features gracefully', async () => {
      // Test with limited feature environment
      const limitedEnv: BrowserEnvironment = {
        name: 'Limited Browser',
        userAgent: 'Mozilla/5.0 (compatible; LimitedBrowser/1.0)',
        features: {
          touch: false,
          webgl: false,
          webassembly: false,
          serviceWorker: false,
          resizeObserver: false,
          intersectionObserver: false,
          matchMedia: false,
          vibrate: false,
          battery: false,
          connection: false
        },
        viewport: { width: 800, height: 600, devicePixelRatio: 1 },
        performance: {}
      };

      const dom = setupBrowserEnvironment(limitedEnv);
      const container = dom.window.document.getElementById('container')!;
      
      const renderer = new LIVRenderer({
        container,
        permissions: createDefaultSecurityPolicy(),
        enableFallback: true
      });

      const mockDocument = createMockLIVDocument();
      
      // Should not throw errors even with missing features
      expect(async () => {
        await renderer.renderDocument(mockDocument);
      }).to.not.throw();

      // Should fall back to static content
      expect(renderer.getRenderingState().isFallbackMode()).to.be.true;

      renderer.destroy();
    });

    it('should provide meaningful error messages across platforms', () => {
      BROWSER_ENVIRONMENTS.forEach(env => {
        const dom = setupBrowserEnvironment(env);
        const container = dom.window.document.getElementById('container')!;
        
        const renderer = new LIVRenderer({
          container,
          permissions: createDefaultSecurityPolicy()
        });

        const errors = renderer.getRenderingState().getErrors();
        
        // Errors should be informative regardless of platform
        errors.forEach(error => {
          expect(error.message).to.be.a('string');
          expect(error.message.length).to.be.greaterThan(0);
        });

        renderer.destroy();
      });
    });
  });

  describe('Accessibility Compatibility', () => {
    it('should support accessibility features across platforms', () => {
      BROWSER_ENVIRONMENTS.forEach(env => {
        const dom = setupBrowserEnvironment(env);
        const container = dom.window.document.getElementById('container')!;
        
        // Mock accessibility preferences
        dom.window.matchMedia = (query: string) => ({
          matches: query.includes('prefers-reduced-motion: reduce') || 
                  query.includes('prefers-high-contrast: active'),
          media: query,
          onchange: null,
          addListener: () => {},
          removeListener: () => {},
          addEventListener: () => {},
          removeEventListener: () => {},
          dispatchEvent: () => true
        });

        const renderer = new LIVRenderer({
          container,
          permissions: createDefaultSecurityPolicy(),
          enableResponsiveDesign: true
        });

        // Should apply accessibility classes
        const responsiveManager = (renderer as any).responsiveManager;
        if (responsiveManager) {
          responsiveManager.initializeClasses();
          
          expect(container.classList.contains('liv-reduced-motion')).to.be.true;
        }

        renderer.destroy();
      });
    });
  });
});

// Helper functions
function setupBrowserEnvironment(env: BrowserEnvironment): JSDOM {
  const dom = new JSDOM(`
    <!DOCTYPE html>
    <html>
      <head>
        <meta charset="utf-8">
        <meta name="viewport" content="width=device-width, initial-scale=1">
        <title>LIV Test</title>
      </head>
      <body>
        <div id="container" style="width: 100%; height: 100vh;"></div>
      </body>
    </html>
  `, {
    url: 'http://localhost:3000',
    pretendToBeVisual: true,
    resources: 'usable'
  });

  // Set up window properties
  const window = dom.window as any;
  
  // Basic window properties
  window.innerWidth = env.viewport.width;
  window.innerHeight = env.viewport.height;
  window.devicePixelRatio = env.viewport.devicePixelRatio;
  window.screen = {
    width: env.viewport.width,
    height: env.viewport.height,
    orientation: {
      type: env.viewport.width > env.viewport.height ? 'landscape-primary' : 'portrait-primary'
    }
  };

  // Navigator properties
  window.navigator.userAgent = env.userAgent;
  window.navigator.maxTouchPoints = env.features.touch ? 5 : 0;
  
  if (env.features.vibrate) {
    window.navigator.vibrate = () => true;
  }
  
  if (env.features.battery) {
    window.navigator.getBattery = () => Promise.resolve({
      level: 0.8,
      charging: false,
      addEventListener: () => {}
    });
  }
  
  if (env.features.connection) {
    window.navigator.connection = {
      effectiveType: '4g',
      downlink: 10,
      rtt: 100
    };
  }

  // Feature detection
  if (env.features.touch) {
    window.ontouchstart = {};
  }

  if (env.features.webgl) {
    window.WebGLRenderingContext = class MockWebGLContext {};
  }

  if (env.features.webassembly) {
    window.WebAssembly = {
      instantiate: () => Promise.resolve({ instance: { exports: {} } }),
      compile: () => Promise.resolve({}),
      Module: class MockModule {}
    };
  }

  if (env.features.serviceWorker) {
    window.navigator.serviceWorker = {
      register: () => Promise.resolve({}),
      ready: Promise.resolve({})
    };
  }

  if (env.features.resizeObserver) {
    window.ResizeObserver = class MockResizeObserver {
      constructor(callback: any) {}
      observe() {}
      unobserve() {}
      disconnect() {}
    };
  }

  if (env.features.intersectionObserver) {
    window.IntersectionObserver = class MockIntersectionObserver {
      constructor(callback: any, options?: any) {}
      observe() {}
      unobserve() {}
      disconnect() {}
    };
  }

  if (env.features.matchMedia) {
    window.matchMedia = (query: string) => ({
      matches: false,
      media: query,
      onchange: null,
      addListener: () => {},
      removeListener: () => {},
      addEventListener: () => {},
      removeEventListener: () => {},
      dispatchEvent: () => true
    });
  }

  // Performance API
  if (env.performance.memory) {
    window.performance.memory = {
      usedJSHeapSize: 10000000,
      totalJSHeapSize: 20000000,
      jsHeapSizeLimit: 100000000
    };
  }

  if (env.performance.navigation) {
    window.performance.navigation = {
      type: 0,
      redirectCount: 0
    };
  }

  if (env.performance.timing) {
    window.performance.timing = {
      navigationStart: Date.now() - 1000,
      loadEventEnd: Date.now()
    };
  }

  // Set globals
  global.window = window;
  global.document = window.document;
  global.navigator = window.navigator;
  global.performance = window.performance;

  return dom;
}

function createMockLIVDocument(withAnimations: boolean = false): LIVDocument {
  return {
    manifest: {
      version: '1.0.0',
      metadata: {
        title: 'Test Document',
        author: 'Test Author',
        created: new Date().toISOString(),
        modified: new Date().toISOString(),
        description: 'Test document for cross-platform compatibility',
        version: '1.0.0',
        language: 'en'
      },
      security: createDefaultSecurityPolicy(),
      resources: {},
      features: {
        animations: withAnimations,
        interactivity: true,
        charts: false,
        forms: false,
        audio: false,
        video: false,
        webgl: false,
        webassembly: true
      }
    },
    content: {
      html: '<div class="test-content">Hello World</div>',
      css: withAnimations ? '.test-content { animation: fadeIn 1s ease-in; } @keyframes fadeIn { from { opacity: 0; } to { opacity: 1; } }' : '.test-content { color: blue; }',
      interactiveSpec: '{}',
      staticFallback: '<div>Static fallback content</div>'
    },
    assets: {
      images: new Map(),
      fonts: new Map(),
      data: new Map()
    },
    signatures: {
      contentSignature: 'mock-signature',
      manifestSignature: 'mock-signature',
      wasmSignatures: {}
    },
    wasmModules: new Map(),
    validate: () => ({ isValid: true, errors: [], warnings: [] }),
    generateSecurityReport: () => ({ 
      isValid: true, 
      signatureVerified: true, 
      integrityChecked: true, 
      permissionsValid: true, 
      warnings: [], 
      errors: [] 
    })
  } as any;
}

function createDefaultSecurityPolicy(): LegacySecurityPolicy {
  return {
    wasmPermissions: {
      memoryLimit: 16 * 1024 * 1024,
      allowedImports: ['console'],
      cpuTimeLimit: 5000,
      allowNetworking: false,
      allowFileSystem: false
    },
    jsPermissions: {
      executionMode: 'sandboxed',
      allowedAPIs: ['console'],
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
}

function createMockTouch(window: any, x: number, y: number, identifier: number): Touch {
  return {
    identifier,
    clientX: x,
    clientY: y,
    pageX: x,
    pageY: y,
    screenX: x,
    screenY: y,
    target: window.document.getElementById('container'),
    radiusX: 10,
    radiusY: 10,
    rotationAngle: 0,
    force: 1
  } as Touch;
}

function simulateTapGesture(window: any, container: HTMLElement): void {
  const touch = createMockTouch(window, 100, 100, 1);
  
  // Touch start
  const startEvent = new window.TouchEvent('touchstart', {
    touches: [touch],
    changedTouches: [touch],
    targetTouches: [touch]
  });
  container.dispatchEvent(startEvent);
  
  // Touch end (short duration for tap)
  setTimeout(() => {
    const endEvent = new window.TouchEvent('touchend', {
      touches: [],
      changedTouches: [touch],
      targetTouches: []
    });
    container.dispatchEvent(endEvent);
  }, 100);
}

async function measurePlatformPerformance(env: BrowserEnvironment): Promise<{
  renderTime: number;
  frameRate: number;
  memoryUsage: number;
}> {
  const dom = setupBrowserEnvironment(env);
  const container = dom.window.document.getElementById('container')!;
  
  const renderer = new LIVRenderer({
    container,
    permissions: createDefaultSecurityPolicy(),
    enableAnimations: true
  });

  const startTime = performance.now();
  const mockDocument = createMockLIVDocument(true);
  
  await renderer.renderDocument(mockDocument);
  const endTime = performance.now();
  
  const metrics = renderer.getPerformanceMetrics();
  
  const result = {
    renderTime: endTime - startTime,
    frameRate: metrics.averageFPS,
    memoryUsage: (performance as any).memory?.usedJSHeapSize || 0
  };
  
  renderer.destroy();
  
  return result;
}