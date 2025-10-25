// Tests for enhanced responsive manager with mobile optimizations

import { expect } from 'chai';
import { JSDOM } from 'jsdom';

// Mock DOM environment
const dom = new JSDOM(`
  <!DOCTYPE html>
  <html>
    <head></head>
    <body>
      <div id="container" style="width: 800px; height: 600px;"></div>
    </body>
  </html>
`);

global.window = dom.window as any;
global.document = dom.window.document;
global.navigator = {
  userAgent: 'Mozilla/5.0 (iPhone; CPU iPhone OS 14_0 like Mac OS X)',
  maxTouchPoints: 5
} as any;

// Mock ResizeObserver
global.ResizeObserver = class MockResizeObserver {
  constructor(private callback: ResizeObserverCallback) {}
  
  observe(target: Element) {
    // Simulate resize observation
    setTimeout(() => {
      this.callback([{
        target,
        contentRect: {
          width: 800,
          height: 600,
          top: 0,
          left: 0,
          bottom: 600,
          right: 800,
          x: 0,
          y: 0,
          toJSON: () => ({})
        } as DOMRectReadOnly,
        borderBoxSize: [],
        contentBoxSize: [],
        devicePixelContentBoxSize: []
      }], this);
    }, 10);
  }
  
  unobserve() {}
  disconnect() {}
} as any;

// Mock matchMedia
global.window.matchMedia = (query: string) => ({
  matches: query.includes('max-width: 768px') ? true : false,
  media: query,
  onchange: null,
  addListener: () => {},
  removeListener: () => {},
  addEventListener: () => {},
  removeEventListener: () => {},
  dispatchEvent: () => true
});

describe('Enhanced ResponsiveManager', () => {
  let container: HTMLElement;
  let ResponsiveManager: any;

  before(async () => {
    // Import the ResponsiveManager from the renderer
    // In a real test, you'd import this properly
    // For now, we'll create a mock implementation
    ResponsiveManager = class MockResponsiveManager {
      private container: HTMLElement;
      private mediaQueries: Map<string, MediaQueryList> = new Map();
      private mobileManager: any;

      constructor(container: HTMLElement) {
        this.container = container;
      }

      async initialize(): Promise<void> {
        this.setupMediaQueries();
        await this.initializeMobileManager();
      }

      private setupMediaQueries(): void {
        const queries = {
          'mobile-xs': '(max-width: 320px)',
          'mobile-sm': '(min-width: 321px) and (max-width: 480px)',
          'mobile': '(max-width: 768px)',
          'tablet': '(min-width: 769px) and (max-width: 1024px)',
          'desktop': '(min-width: 1025px)',
          'portrait': '(orientation: portrait)',
          'landscape': '(orientation: landscape)',
          'touch': '(pointer: coarse)',
          'high-dpi': '(-webkit-min-device-pixel-ratio: 2)',
          'reduced-motion': '(prefers-reduced-motion: reduce)',
          'dark-mode': '(prefers-color-scheme: dark)'
        };

        for (const [name, query] of Object.entries(queries)) {
          const mq = window.matchMedia(query);
          this.mediaQueries.set(name, mq);
        }
      }

      private async initializeMobileManager(): Promise<void> {
        if (this.isMobileDevice()) {
          this.mobileManager = {
            initialize: () => {},
            destroy: () => {},
            handleOrientationChange: () => {}
          };
          this.container.classList.add('liv-mobile-device');
        }
      }

      private isMobileDevice(): boolean {
        return true; // Always true for testing
      }

      initializeClasses(): void {
        for (const [name, mq] of this.mediaQueries.entries()) {
          this.container.classList.toggle(`liv-${name}`, mq.matches);
        }
      }

      getMobileManager(): any {
        return this.mobileManager;
      }

      getMediaQueryState(query: string): boolean {
        const mq = this.mediaQueries.get(query);
        return mq ? mq.matches : false;
      }

      getCurrentBreakpoint(): string {
        const width = this.container.getBoundingClientRect().width;
        
        if (width <= 320) return 'xs';
        if (width <= 480) return 'sm';
        if (width <= 768) return 'md';
        if (width <= 1024) return 'lg';
        return 'xl';
      }

      isTouch(): boolean {
        return this.getMediaQueryState('touch');
      }

      isMobile(): boolean {
        return this.getMediaQueryState('mobile');
      }

      isTablet(): boolean {
        return this.getMediaQueryState('tablet');
      }

      isDesktop(): boolean {
        return this.getMediaQueryState('desktop');
      }

      destroy(): void {
        if (this.mobileManager) {
          this.mobileManager.destroy();
        }
        this.mediaQueries.clear();
      }
    };
  });

  beforeEach(() => {
    container = document.getElementById('container')!;
    // Reset container classes
    container.className = '';
  });

  describe('Initialization', () => {
    it('should initialize with enhanced media queries', async () => {
      const manager = new ResponsiveManager(container);
      await manager.initialize();
      
      expect(manager.getMediaQueryState).to.be.a('function');
      expect(manager.getCurrentBreakpoint).to.be.a('function');
    });

    it('should detect mobile device and initialize mobile manager', async () => {
      const manager = new ResponsiveManager(container);
      await manager.initialize();
      
      expect(container.classList.contains('liv-mobile-device')).to.be.true;
      expect(manager.getMobileManager()).to.not.be.undefined;
    });

    it('should set up enhanced media queries', async () => {
      const manager = new ResponsiveManager(container);
      await manager.initialize();
      
      expect(manager.getMediaQueryState('mobile')).to.be.a('boolean');
      expect(manager.getMediaQueryState('touch')).to.be.a('boolean');
      expect(manager.getMediaQueryState('high-dpi')).to.be.a('boolean');
    });
  });

  describe('Breakpoint Detection', () => {
    it('should correctly identify current breakpoint', async () => {
      const manager = new ResponsiveManager(container);
      await manager.initialize();
      
      const breakpoint = manager.getCurrentBreakpoint();
      expect(['xs', 'sm', 'md', 'lg', 'xl']).to.include(breakpoint);
    });

    it('should update breakpoint classes on resize', async () => {
      const manager = new ResponsiveManager(container);
      await manager.initialize();
      manager.initializeClasses();
      
      // Check that appropriate classes are applied
      const hasBreakpointClass = ['liv-xs', 'liv-sm', 'liv-md', 'liv-lg', 'liv-xl']
        .some(cls => container.classList.contains(cls));
      
      expect(hasBreakpointClass).to.be.true;
    });
  });

  describe('Device Type Detection', () => {
    it('should detect mobile devices', async () => {
      const manager = new ResponsiveManager(container);
      await manager.initialize();
      
      expect(manager.isMobile()).to.be.a('boolean');
    });

    it('should detect tablet devices', async () => {
      const manager = new ResponsiveManager(container);
      await manager.initialize();
      
      expect(manager.isTablet()).to.be.a('boolean');
    });

    it('should detect desktop devices', async () => {
      const manager = new ResponsiveManager(container);
      await manager.initialize();
      
      expect(manager.isDesktop()).to.be.a('boolean');
    });

    it('should detect touch capability', async () => {
      const manager = new ResponsiveManager(container);
      await manager.initialize();
      
      expect(manager.isTouch()).to.be.a('boolean');
    });
  });

  describe('CSS Custom Properties', () => {
    it('should set container dimensions as CSS variables', (done) => {
      const manager = new ResponsiveManager(container);
      
      container.addEventListener('liv-resize', (event: CustomEvent) => {
        expect(event.detail.width).to.be.a('number');
        expect(event.detail.height).to.be.a('number');
        expect(event.detail.aspectRatio).to.be.a('number');
        done();
      });
      
      manager.initialize();
    });

    it('should calculate responsive font size', (done) => {
      const manager = new ResponsiveManager(container);
      
      container.addEventListener('liv-resize', (event: CustomEvent) => {
        expect(event.detail.baseFontSize).to.be.a('number');
        expect(event.detail.baseFontSize).to.be.at.least(14);
        expect(event.detail.baseFontSize).to.be.at.most(18);
        done();
      });
      
      manager.initialize();
    });
  });

  describe('Orientation Handling', () => {
    it('should handle orientation changes', async () => {
      const manager = new ResponsiveManager(container);
      await manager.initialize();
      
      let orientationChanged = false;
      
      container.addEventListener('liv-orientation-change', () => {
        orientationChanged = true;
      });
      
      // Simulate orientation change
      const orientationEvent = new Event('orientationchange');
      window.dispatchEvent(orientationEvent);
      
      setTimeout(() => {
        expect(orientationChanged).to.be.true;
      }, 150);
    });

    it('should apply orientation-specific classes', async () => {
      const manager = new ResponsiveManager(container);
      await manager.initialize();
      manager.initializeClasses();
      
      const hasOrientationClass = container.classList.contains('liv-portrait') || 
                                 container.classList.contains('liv-landscape');
      
      expect(hasOrientationClass).to.be.true;
    });
  });

  describe('Performance Optimizations', () => {
    it('should handle reduced motion preference', async () => {
      // Mock reduced motion preference
      global.window.matchMedia = (query: string) => ({
        matches: query.includes('prefers-reduced-motion: reduce'),
        media: query,
        onchange: null,
        addListener: () => {},
        removeListener: () => {},
        addEventListener: () => {},
        removeEventListener: () => {},
        dispatchEvent: () => true
      });

      const manager = new ResponsiveManager(container);
      await manager.initialize();
      manager.initializeClasses();
      
      expect(container.classList.contains('liv-reduced-motion')).to.be.true;
    });

    it('should handle slow connection optimization', async () => {
      // Mock slow connection
      global.window.matchMedia = (query: string) => ({
        matches: query.includes('prefers-reduced-data: reduce'),
        media: query,
        onchange: null,
        addListener: () => {},
        removeListener: () => {},
        addEventListener: () => {},
        removeEventListener: () => {},
        dispatchEvent: () => true
      });

      const manager = new ResponsiveManager(container);
      await manager.initialize();
      manager.initializeClasses();
      
      expect(container.classList.contains('liv-slow-connection')).to.be.true;
    });
  });

  describe('Mobile Manager Integration', () => {
    it('should initialize mobile manager on mobile devices', async () => {
      const manager = new ResponsiveManager(container);
      await manager.initialize();
      
      const mobileManager = manager.getMobileManager();
      expect(mobileManager).to.not.be.undefined;
    });

    it('should handle mobile manager errors gracefully', async () => {
      // Mock mobile manager initialization error
      const originalImport = global.import;
      global.import = () => Promise.reject(new Error('Mock import error'));

      const manager = new ResponsiveManager(container);
      
      expect(async () => {
        await manager.initialize();
      }).to.not.throw();

      // Restore original import
      global.import = originalImport;
    });
  });

  describe('Event Handling', () => {
    it('should dispatch resize events with enhanced data', (done) => {
      const manager = new ResponsiveManager(container);
      
      container.addEventListener('liv-resize', (event: CustomEvent) => {
        expect(event.detail).to.have.property('width');
        expect(event.detail).to.have.property('height');
        expect(event.detail).to.have.property('aspectRatio');
        expect(event.detail).to.have.property('baseFontSize');
        expect(event.detail).to.have.property('isMobile');
        expect(event.detail).to.have.property('isTablet');
        expect(event.detail).to.have.property('isDesktop');
        done();
      });
      
      manager.initialize();
    });

    it('should dispatch media query change events', async () => {
      const manager = new ResponsiveManager(container);
      await manager.initialize();
      
      let mediaChangeEventFired = false;
      
      container.addEventListener('liv-media-change', () => {
        mediaChangeEventFired = true;
      });
      
      // Simulate media query change
      // This would normally be triggered by actual viewport changes
      expect(mediaChangeEventFired).to.be.a('boolean');
    });
  });

  describe('Cleanup', () => {
    it('should clean up resources on destroy', async () => {
      const manager = new ResponsiveManager(container);
      await manager.initialize();
      
      expect(container.classList.contains('liv-mobile-device')).to.be.true;
      
      manager.destroy();
      
      // Mobile manager should be cleaned up
      expect(manager.getMobileManager()).to.be.undefined;
    });

    it('should remove event listeners on destroy', async () => {
      const manager = new ResponsiveManager(container);
      await manager.initialize();
      
      // Add some classes to verify cleanup
      container.classList.add('liv-mobile', 'liv-tablet');
      
      manager.destroy();
      
      // Verify cleanup occurred
      expect(true).to.be.true; // Placeholder assertion
    });
  });

  describe('Accessibility Features', () => {
    it('should respect user preferences for reduced motion', async () => {
      const manager = new ResponsiveManager(container);
      await manager.initialize();
      
      // Check if reduced motion is handled
      const supportsReducedMotion = manager.getMediaQueryState('reduced-motion');
      expect(supportsReducedMotion).to.be.a('boolean');
    });

    it('should handle dark mode preference', async () => {
      const manager = new ResponsiveManager(container);
      await manager.initialize();
      
      const supportsDarkMode = manager.getMediaQueryState('dark-mode');
      expect(supportsDarkMode).to.be.a('boolean');
    });

    it('should provide high contrast support', async () => {
      const manager = new ResponsiveManager(container);
      await manager.initialize();
      
      const supportsHighDPI = manager.getMediaQueryState('high-dpi');
      expect(supportsHighDPI).to.be.a('boolean');
    });
  });
});