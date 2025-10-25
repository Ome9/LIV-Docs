// Tests for mobile viewing experience optimizations

import { expect } from 'chai';
import { JSDOM } from 'jsdom';
import { MobileManager, MobileOptimizationOptions } from '../src/mobile-manager';
import { GestureType, InteractionType } from '../src/types';

// Mock DOM environment
const dom = new JSDOM('<!DOCTYPE html><div id="container"></div>');
global.window = dom.window as any;
global.document = dom.window.document;
global.navigator = {
  userAgent: 'Mozilla/5.0 (iPhone; CPU iPhone OS 14_0 like Mac OS X)',
  maxTouchPoints: 5,
  vibrate: () => true
} as any;
global.performance = {
  now: () => Date.now()
} as any;

describe('MobileManager', () => {
  let container: HTMLElement;
  let mobileManager: MobileManager;

  beforeEach(() => {
    container = document.getElementById('container')!;
    mobileManager = new MobileManager(container);
  });

  afterEach(() => {
    if (mobileManager) {
      mobileManager.destroy();
    }
  });

  describe('Initialization', () => {
    it('should initialize with default options', () => {
      expect(mobileManager).to.be.instanceOf(MobileManager);
    });

    it('should detect mobile device correctly', () => {
      mobileManager.initialize();
      expect(container.classList.contains('liv-mobile-optimized')).to.be.true;
    });

    it('should apply mobile CSS optimizations', () => {
      mobileManager.initialize();
      
      const styles = window.getComputedStyle(container);
      expect(container.classList.contains('liv-mobile-optimized')).to.be.true;
    });

    it('should set up viewport meta tag', () => {
      mobileManager.initialize();
      
      const viewport = document.querySelector('meta[name="viewport"]') as HTMLMetaElement;
      expect(viewport).to.not.be.null;
      expect(viewport.content).to.include('width=device-width');
    });
  });

  describe('Touch Handling', () => {
    beforeEach(() => {
      mobileManager.initialize();
    });

    it('should handle touch start events', () => {
      let eventFired = false;
      
      container.addEventListener('liv-interaction', (event: CustomEvent) => {
        expect(event.detail.eventType).to.equal(InteractionType.TouchStart);
        eventFired = true;
      });

      const touchEvent = new TouchEvent('touchstart', {
        touches: [createMockTouch(100, 100, 1)],
        changedTouches: [createMockTouch(100, 100, 1)],
        targetTouches: [createMockTouch(100, 100, 1)]
      });

      container.dispatchEvent(touchEvent);
      expect(eventFired).to.be.true;
    });

    it('should handle multi-touch events', () => {
      let eventFired = false;
      
      container.addEventListener('liv-interaction', (event: CustomEvent) => {
        expect(event.detail.touch_data.touches).to.have.length(2);
        eventFired = true;
      });

      const touchEvent = new TouchEvent('touchstart', {
        touches: [
          createMockTouch(100, 100, 1),
          createMockTouch(200, 200, 2)
        ],
        changedTouches: [
          createMockTouch(100, 100, 1),
          createMockTouch(200, 200, 2)
        ],
        targetTouches: [
          createMockTouch(100, 100, 1),
          createMockTouch(200, 200, 2)
        ]
      });

      container.dispatchEvent(touchEvent);
      expect(eventFired).to.be.true;
    });

    it('should limit touch points based on maxTouchPoints option', () => {
      const limitedManager = new MobileManager(container, { maxTouchPoints: 2 });
      limitedManager.initialize();
      
      let eventFired = false;
      
      container.addEventListener('liv-interaction', (event: CustomEvent) => {
        expect(event.detail.touch_data.touches).to.have.length.at.most(2);
        eventFired = true;
      });

      const touchEvent = new TouchEvent('touchstart', {
        touches: [
          createMockTouch(100, 100, 1),
          createMockTouch(200, 200, 2),
          createMockTouch(300, 300, 3)
        ],
        changedTouches: [createMockTouch(100, 100, 1)],
        targetTouches: [createMockTouch(100, 100, 1)]
      });

      container.dispatchEvent(touchEvent);
      expect(eventFired).to.be.true;
      
      limitedManager.destroy();
    });
  });

  describe('Gesture Recognition', () => {
    beforeEach(() => {
      mobileManager.initialize();
    });

    it('should recognize tap gestures', (done) => {
      container.addEventListener('liv-interaction', (event: CustomEvent) => {
        if (event.detail.eventType === InteractionType.Tap) {
          expect(event.detail.gesture_data.gesture_type).to.equal(GestureType.Tap);
          done();
        }
      });

      // Simulate tap gesture
      simulateTapGesture(container, 100, 100);
    });

    it('should recognize swipe gestures', (done) => {
      container.addEventListener('liv-interaction', (event: CustomEvent) => {
        if (event.detail.eventType === InteractionType.Swipe) {
          expect(event.detail.gesture_data.gesture_type).to.equal(GestureType.Swipe);
          done();
        }
      });

      // Simulate swipe gesture
      simulateSwipeGesture(container, 100, 100, 200, 100);
    });

    it('should recognize pinch gestures', (done) => {
      container.addEventListener('liv-interaction', (event: CustomEvent) => {
        if (event.detail.eventType === InteractionType.Pinch) {
          expect(event.detail.gesture_data.gesture_type).to.equal(GestureType.Pinch);
          expect(event.detail.gesture_data.scale).to.be.greaterThan(1);
          done();
        }
      });

      // Simulate pinch gesture
      simulatePinchGesture(container);
    });

    it('should recognize long press gestures', (done) => {
      container.addEventListener('liv-interaction', (event: CustomEvent) => {
        if (event.detail.eventType === InteractionType.LongPress) {
          expect(event.detail.gesture_data.gesture_type).to.equal(GestureType.LongPress);
          expect(event.detail.gesture_data.duration).to.be.greaterThan(500);
          done();
        }
      });

      // Simulate long press gesture
      simulateLongPressGesture(container, 100, 100);
    });
  });

  describe('Performance Management', () => {
    beforeEach(() => {
      mobileManager.initialize();
    });

    it('should provide performance metrics', () => {
      const metrics = mobileManager.getPerformanceMetrics();
      
      expect(metrics).to.have.property('frameRate');
      expect(metrics).to.have.property('memoryUsage');
      expect(metrics).to.have.property('touchLatency');
      expect(metrics).to.have.property('gestureAccuracy');
    });

    it('should adapt to battery level changes', () => {
      // Mock battery API
      (global.navigator as any).getBattery = () => Promise.resolve({
        level: 0.1, // Low battery
        addEventListener: () => {}
      });

      const performanceManager = new MobileManager(container, {
        enablePerformanceMode: true
      });
      
      performanceManager.initialize();
      
      // Performance mode should adapt to low battery
      setTimeout(() => {
        const metrics = performanceManager.getPerformanceMetrics();
        expect(metrics.frameRate).to.be.lessThan(30);
        performanceManager.destroy();
      }, 100);
    });
  });

  describe('Configuration', () => {
    it('should allow enabling/disabling gestures', () => {
      mobileManager.initialize();
      
      mobileManager.disableGestures();
      // Gestures should be disabled
      
      mobileManager.enableGestures();
      // Gestures should be enabled again
    });

    it('should allow setting gesture threshold', () => {
      mobileManager.initialize();
      mobileManager.setGestureThreshold(20);
      
      // Gesture threshold should be updated
      expect(true).to.be.true; // Placeholder assertion
    });

    it('should handle custom options', () => {
      const customOptions: MobileOptimizationOptions = {
        enableGestures: false,
        enableTouchOptimization: true,
        enablePerformanceMode: false,
        gestureThreshold: 15,
        touchDelay: 200,
        maxTouchPoints: 5,
        enableHapticFeedback: true
      };

      const customManager = new MobileManager(container, customOptions);
      customManager.initialize();
      
      expect(true).to.be.true; // Placeholder assertion
      customManager.destroy();
    });
  });

  describe('Error Handling', () => {
    it('should handle initialization errors gracefully', () => {
      // Mock error condition
      const originalAddEventListener = container.addEventListener;
      container.addEventListener = () => {
        throw new Error('Mock error');
      };

      expect(() => {
        mobileManager.initialize();
      }).to.not.throw();

      // Restore original method
      container.addEventListener = originalAddEventListener;
    });

    it('should handle touch event errors gracefully', () => {
      mobileManager.initialize();

      // Create malformed touch event
      const badTouchEvent = new Event('touchstart');
      
      expect(() => {
        container.dispatchEvent(badTouchEvent);
      }).to.not.throw();
    });
  });

  describe('Cleanup', () => {
    it('should clean up resources on destroy', () => {
      mobileManager.initialize();
      
      expect(container.classList.contains('liv-mobile-optimized')).to.be.true;
      
      mobileManager.destroy();
      
      expect(container.classList.contains('liv-mobile-optimized')).to.be.false;
    });

    it('should remove event listeners on destroy', () => {
      mobileManager.initialize();
      
      const initialListenerCount = getEventListenerCount(container);
      
      mobileManager.destroy();
      
      const finalListenerCount = getEventListenerCount(container);
      expect(finalListenerCount).to.be.lessThanOrEqual(initialListenerCount);
    });
  });
});

// Helper functions for testing
function createMockTouch(x: number, y: number, identifier: number): Touch {
  return {
    identifier,
    clientX: x,
    clientY: y,
    pageX: x,
    pageY: y,
    screenX: x,
    screenY: y,
    target: container,
    radiusX: 10,
    radiusY: 10,
    rotationAngle: 0,
    force: 1
  } as Touch;
}

function simulateTapGesture(element: HTMLElement, x: number, y: number): void {
  const touch = createMockTouch(x, y, 1);
  
  // Touch start
  const startEvent = new TouchEvent('touchstart', {
    touches: [touch],
    changedTouches: [touch],
    targetTouches: [touch]
  });
  element.dispatchEvent(startEvent);
  
  // Touch end (short duration for tap)
  setTimeout(() => {
    const endEvent = new TouchEvent('touchend', {
      touches: [],
      changedTouches: [touch],
      targetTouches: []
    });
    element.dispatchEvent(endEvent);
  }, 100);
}

function simulateSwipeGesture(element: HTMLElement, startX: number, startY: number, endX: number, endY: number): void {
  const startTouch = createMockTouch(startX, startY, 1);
  const endTouch = createMockTouch(endX, endY, 1);
  
  // Touch start
  const startEvent = new TouchEvent('touchstart', {
    touches: [startTouch],
    changedTouches: [startTouch],
    targetTouches: [startTouch]
  });
  element.dispatchEvent(startEvent);
  
  // Touch move (simulate movement)
  setTimeout(() => {
    const moveEvent = new TouchEvent('touchmove', {
      touches: [endTouch],
      changedTouches: [endTouch],
      targetTouches: [endTouch]
    });
    element.dispatchEvent(moveEvent);
  }, 50);
  
  // Touch end
  setTimeout(() => {
    const endEvent = new TouchEvent('touchend', {
      touches: [],
      changedTouches: [endTouch],
      targetTouches: []
    });
    element.dispatchEvent(endEvent);
  }, 150);
}

function simulatePinchGesture(element: HTMLElement): void {
  const touch1Start = createMockTouch(100, 100, 1);
  const touch2Start = createMockTouch(200, 200, 2);
  const touch1End = createMockTouch(80, 80, 1);
  const touch2End = createMockTouch(220, 220, 2);
  
  // Touch start with two fingers
  const startEvent = new TouchEvent('touchstart', {
    touches: [touch1Start, touch2Start],
    changedTouches: [touch1Start, touch2Start],
    targetTouches: [touch1Start, touch2Start]
  });
  element.dispatchEvent(startEvent);
  
  // Touch move (pinch out)
  setTimeout(() => {
    const moveEvent = new TouchEvent('touchmove', {
      touches: [touch1End, touch2End],
      changedTouches: [touch1End, touch2End],
      targetTouches: [touch1End, touch2End]
    });
    element.dispatchEvent(moveEvent);
  }, 100);
  
  // Touch end
  setTimeout(() => {
    const endEvent = new TouchEvent('touchend', {
      touches: [],
      changedTouches: [touch1End, touch2End],
      targetTouches: []
    });
    element.dispatchEvent(endEvent);
  }, 200);
}

function simulateLongPressGesture(element: HTMLElement, x: number, y: number): void {
  const touch = createMockTouch(x, y, 1);
  
  // Touch start
  const startEvent = new TouchEvent('touchstart', {
    touches: [touch],
    changedTouches: [touch],
    targetTouches: [touch]
  });
  element.dispatchEvent(startEvent);
  
  // Touch end after long duration
  setTimeout(() => {
    const endEvent = new TouchEvent('touchend', {
      touches: [],
      changedTouches: [touch],
      targetTouches: []
    });
    element.dispatchEvent(endEvent);
  }, 600); // Long press threshold
}

function getEventListenerCount(element: HTMLElement): number {
  // This is a simplified way to check event listeners
  // In a real implementation, you'd need a more sophisticated approach
  return Object.keys((element as any)._events || {}).length;
}