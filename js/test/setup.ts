/**
 * Jest test setup file
 */

import { TextEncoder, TextDecoder } from 'util';

// Add Node.js polyfills for browser APIs
(global as any).TextEncoder = TextEncoder;
(global as any).TextDecoder = TextDecoder;

// Mock performance API if not available
if (typeof performance === 'undefined') {
  (global as any).performance = {
    now: () => Date.now(),
    memory: {
      usedJSHeapSize: 1024 * 1024 // 1MB
    }
  };
}

// Mock console methods for cleaner test output
const originalConsoleError = console.error;
const originalConsoleWarn = console.warn;
const originalConsoleLog = console.log;

beforeEach(() => {
  // Suppress console output during tests unless explicitly needed
  console.error = jest.fn();
  console.warn = jest.fn();
  console.log = jest.fn();
});

afterEach(() => {
  // Restore console methods
  console.error = originalConsoleError;
  console.warn = originalConsoleWarn;
  console.log = originalConsoleLog;
  
  // Clean up any global state
  delete (globalThis as any).goSandboxBridge;
  delete (globalThis as any).handleGoMessage;
  delete (globalThis as any).handleGoBridgeStatus;
  delete (globalThis as any).postMessage;
});

// Mock DOM APIs that might not be available in jsdom
if (typeof window !== 'undefined') {
  // Mock requestAnimationFrame
  if (!window.requestAnimationFrame) {
    window.requestAnimationFrame = (callback: FrameRequestCallback) => {
      return setTimeout(callback, 16); // ~60fps
    };
  }

  // Mock cancelAnimationFrame
  if (!window.cancelAnimationFrame) {
    window.cancelAnimationFrame = (id: number) => {
      clearTimeout(id);
    };
  }

  // Mock ResizeObserver
  if (!window.ResizeObserver) {
    window.ResizeObserver = class ResizeObserver {
      observe() {}
      unobserve() {}
      disconnect() {}
    };
  }

  // Mock IntersectionObserver
  if (!window.IntersectionObserver) {
    window.IntersectionObserver = class IntersectionObserver {
      constructor() {}
      observe() {}
      unobserve() {}
      disconnect() {}
    };
  }
}

// Global test utilities
(global as any).createMockElement = (tagName: string = 'div') => ({
  tagName: tagName.toUpperCase(),
  setAttribute: jest.fn(),
  getAttribute: jest.fn(),
  appendChild: jest.fn(),
  removeChild: jest.fn(),
  addEventListener: jest.fn(),
  removeEventListener: jest.fn(),
  style: {
    setProperty: jest.fn(),
    getPropertyValue: jest.fn()
  },
  parentNode: {
    removeChild: jest.fn()
  },
  textContent: '',
  innerHTML: ''
});

(global as any).createMockDocument = () => ({
  createElement: jest.fn(() => (global as any).createMockElement()),
  querySelector: jest.fn(),
  querySelectorAll: jest.fn(() => []),
  createDocumentFragment: jest.fn(() => ({
    childNodes: []
  })),
  body: (global as any).createMockElement('body')
});

// Set up global mocks
if (typeof document === 'undefined') {
  (global as any).document = (global as any).createMockDocument();
}

if (typeof window === 'undefined') {
  (global as any).window = {
    getComputedStyle: jest.fn(() => ({})),
    addEventListener: jest.fn(),
    removeEventListener: jest.fn(),
    setTimeout: setTimeout,
    clearTimeout: clearTimeout,
    setInterval: setInterval,
    clearInterval: clearInterval
  };
}