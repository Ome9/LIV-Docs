/**
 * Tests for LIV Document Format - Secure DOM API
 */

import {
  SecureDOMAPI,
  createSecureDOMAPI,
  DOMSecurityPolicy,
  ElementCreationOptions
} from '../src/dom-api';
import { JSPermissions } from '../src/sandbox-interface';

// Mock DOM environment
const mockElement = {
  tagName: 'DIV',
  setAttribute: jest.fn(),
  getAttribute: jest.fn(),
  appendChild: jest.fn(),
  removeChild: jest.fn(),
  addEventListener: jest.fn(),
  removeEventListener: jest.fn(),
  style: {
    setProperty: jest.fn()
  },
  parentNode: {
    removeChild: jest.fn()
  }
};

const mockDocument = {
  createElement: jest.fn(() => mockElement),
  querySelector: jest.fn(() => mockElement),
  querySelectorAll: jest.fn(() => [mockElement]),
  createDocumentFragment: jest.fn(() => ({
    childNodes: []
  }))
};

const mockWindow = {
  getComputedStyle: jest.fn(() => ({}))
};

// Setup global mocks
(global as any).document = mockDocument;
(global as any).window = mockWindow;

describe('SecureDOMAPI', () => {
  let domAPI: SecureDOMAPI;
  let permissions: JSPermissions;
  let policy: DOMSecurityPolicy;

  beforeEach(() => {
    permissions = {
      executionMode: 'sandboxed',
      allowedAPIs: ['console'],
      domAccess: 'write'
    };

    policy = {
      allowedElements: ['div', 'span', 'p'],
      allowedAttributes: ['id', 'class', 'style'],
      allowedEvents: ['click', 'mouseover'],
      allowedStyles: ['color', 'background-color'],
      maxElements: 100,
      allowScriptExecution: false,
      allowFormSubmission: false,
      allowNavigation: false
    };

    domAPI = createSecureDOMAPI(permissions, policy);

    // Reset mocks
    jest.clearAllMocks();
  });

  afterEach(() => {
    domAPI.cleanup();
  });

  describe('Initialization', () => {
    test('should create DOM API with permissions', () => {
      expect(domAPI).toBeInstanceOf(SecureDOMAPI);
    });

    test('should throw error with no DOM access', () => {
      const noAccessPermissions: JSPermissions = {
        executionMode: 'sandboxed',
        allowedAPIs: [],
        domAccess: 'none'
      };

      expect(() => createSecureDOMAPI(noAccessPermissions, policy))
        .toThrow('DOM access not permitted');
    });

    test('should use default policy when none provided', () => {
      const defaultAPI = createSecureDOMAPI(permissions);
      expect(defaultAPI).toBeInstanceOf(SecureDOMAPI);
    });
  });

  describe('Element Querying', () => {
    test('should query single element with read access', () => {
      const readPermissions: JSPermissions = {
        ...permissions,
        domAccess: 'read'
      };
      const readAPI = createSecureDOMAPI(readPermissions, policy);

      const result = readAPI.querySelector('div.test');
      expect(mockDocument.querySelector).toHaveBeenCalledWith('div.test');
      expect(result).toBe(mockElement);

      readAPI.cleanup();
    });

    test('should query multiple elements', () => {
      const readPermissions: JSPermissions = {
        ...permissions,
        domAccess: 'read'
      };
      const readAPI = createSecureDOMAPI(readPermissions, policy);

      const result = readAPI.querySelectorAll('div');
      expect(mockDocument.querySelectorAll).toHaveBeenCalledWith('div');
      expect(result).toEqual([mockElement]);

      readAPI.cleanup();
    });

    test('should sanitize selectors', () => {
      const readPermissions: JSPermissions = {
        ...permissions,
        domAccess: 'read'
      };
      const readAPI = createSecureDOMAPI(readPermissions, policy);

      readAPI.querySelector('div<script>alert("xss")</script>');
      expect(mockDocument.querySelector).toHaveBeenCalledWith('divscriptalert("xss")/script');

      readAPI.cleanup();
    });

    test('should handle query errors gracefully', () => {
      mockDocument.querySelector.mockImplementation(() => {
        throw new Error('Invalid selector');
      });

      const readPermissions: JSPermissions = {
        ...permissions,
        domAccess: 'read'
      };
      const readAPI = createSecureDOMAPI(readPermissions, policy);

      const result = readAPI.querySelector('invalid');
      expect(result).toBeNull();

      readAPI.cleanup();
    });
  });

  describe('Element Creation', () => {
    test('should create element with valid options', () => {
      const options: ElementCreationOptions = {
        tagName: 'div',
        attributes: { id: 'test', class: 'container' },
        textContent: 'Hello World'
      };

      const result = domAPI.createElement(options);
      
      expect(mockDocument.createElement).toHaveBeenCalledWith('div');
      expect(mockElement.setAttribute).toHaveBeenCalledWith('id', 'test');
      expect(mockElement.setAttribute).toHaveBeenCalledWith('class', 'container');
      expect(result).toBe(mockElement);
    });

    test('should reject disallowed elements', () => {
      const options: ElementCreationOptions = {
        tagName: 'script' // Not in allowed elements
      };

      const result = domAPI.createElement(options);
      expect(result).toBeNull();
      expect(mockDocument.createElement).not.toHaveBeenCalled();
    });

    test('should enforce element limit', () => {
      const limitedPolicy = { ...policy, maxElements: 1 };
      const limitedAPI = createSecureDOMAPI(permissions, limitedPolicy);

      // Create first element (should succeed)
      const options: ElementCreationOptions = { tagName: 'div' };
      const first = limitedAPI.createElement(options);
      expect(first).toBe(mockElement);

      // Create second element (should fail due to limit)
      const second = limitedAPI.createElement(options);
      expect(second).toBeNull();

      limitedAPI.cleanup();
    });

    test('should sanitize text content', () => {
      const options: ElementCreationOptions = {
        tagName: 'div',
        textContent: '<script>alert("xss")</script>'
      };

      domAPI.createElement(options);
      // Text content should be escaped
      expect(mockElement.textContent).toBeDefined();
    });

    test('should require write access for creation', () => {
      const readOnlyPermissions: JSPermissions = {
        ...permissions,
        domAccess: 'read'
      };
      const readOnlyAPI = createSecureDOMAPI(readOnlyPermissions, policy);

      const options: ElementCreationOptions = { tagName: 'div' };
      const result = readOnlyAPI.createElement(options);
      
      expect(result).toBeNull();
      expect(mockDocument.createElement).not.toHaveBeenCalled();

      readOnlyAPI.cleanup();
    });
  });

  describe('Element Manipulation', () => {
    test('should append child elements', () => {
      const parent = mockElement;
      const child = { ...mockElement, tagName: 'SPAN' };

      const result = domAPI.appendChild(parent, child);
      
      expect(result).toBe(true);
      expect(mockElement.appendChild).toHaveBeenCalledWith(child);
    });

    test('should remove elements', () => {
      const result = domAPI.removeElement(mockElement);
      
      expect(result).toBe(true);
      expect(mockElement.parentNode.removeChild).toHaveBeenCalledWith(mockElement);
    });

    test('should set attributes securely', () => {
      const result = domAPI.setAttribute(mockElement, 'id', 'test-id');
      
      expect(result).toBe(true);
      expect(mockElement.setAttribute).toHaveBeenCalledWith('id', 'test-id');
    });

    test('should reject dangerous attributes', () => {
      const result = domAPI.setAttribute(mockElement, 'onclick', 'alert("xss")');
      
      expect(result).toBe(false);
      expect(mockElement.setAttribute).not.toHaveBeenCalled();
    });

    test('should get attributes', () => {
      mockElement.getAttribute.mockReturnValue('test-value');
      
      const result = domAPI.getAttribute(mockElement, 'id');
      
      expect(result).toBe('test-value');
      expect(mockElement.getAttribute).toHaveBeenCalledWith('id');
    });
  });

  describe('Event Handling', () => {
    test('should add event listeners for allowed events', () => {
      const handler = jest.fn();
      const options = {
        element: mockElement,
        eventType: 'click',
        handler: handler
      };

      const result = domAPI.addEventListener(options);
      
      expect(result).toBe(true);
      expect(mockElement.addEventListener).toHaveBeenCalled();
    });

    test('should reject disallowed events', () => {
      const handler = jest.fn();
      const options = {
        element: mockElement,
        eventType: 'load', // Not in allowed events
        handler: handler
      };

      const result = domAPI.addEventListener(options);
      
      expect(result).toBe(false);
      expect(mockElement.addEventListener).not.toHaveBeenCalled();
    });

    test('should remove event listeners', () => {
      const handler = jest.fn();
      
      const result = domAPI.removeEventListener(mockElement, 'click', handler);
      
      expect(result).toBe(true);
      expect(mockElement.removeEventListener).toHaveBeenCalledWith('click', handler);
    });

    test('should wrap event handlers for security', () => {
      const handler = jest.fn();
      const options = {
        element: mockElement,
        eventType: 'click',
        handler: handler
      };

      domAPI.addEventListener(options);
      
      // Should have called addEventListener with wrapped handler
      expect(mockElement.addEventListener).toHaveBeenCalled();
      const [eventType, wrappedHandler] = mockElement.addEventListener.mock.calls[0];
      expect(eventType).toBe('click');
      expect(wrappedHandler).not.toBe(handler); // Should be wrapped
    });
  });

  describe('Style Management', () => {
    test('should set allowed styles', () => {
      const result = domAPI.setStyle(mockElement, 'color', 'red');
      
      expect(result).toBe(true);
      expect(mockElement.style.setProperty).toHaveBeenCalledWith('color', 'red');
    });

    test('should reject dangerous styles', () => {
      const result = domAPI.setStyle(mockElement, 'expression', 'alert("xss")');
      
      expect(result).toBe(false);
      expect(mockElement.style.setProperty).not.toHaveBeenCalled();
    });

    test('should sanitize style values', () => {
      domAPI.setStyle(mockElement, 'color', 'javascript:alert("xss")');
      
      // Should sanitize the value
      expect(mockElement.style.setProperty).toHaveBeenCalledWith('color', ':alert("xss")');
    });

    test('should get computed styles', () => {
      const mockStyles = { color: 'red' };
      mockWindow.getComputedStyle.mockReturnValue(mockStyles);
      
      const result = domAPI.getComputedStyle(mockElement);
      
      expect(result).toBe(mockStyles);
      expect(mockWindow.getComputedStyle).toHaveBeenCalledWith(mockElement);
    });
  });

  describe('Statistics and Monitoring', () => {
    test('should track created elements', () => {
      const options: ElementCreationOptions = { tagName: 'div' };
      domAPI.createElement(options);
      
      const stats = domAPI.getStats();
      expect(stats.createdElements).toBe(1);
    });

    test('should track event listeners', () => {
      const handler = jest.fn();
      const options = {
        element: mockElement,
        eventType: 'click',
        handler: handler
      };
      
      domAPI.addEventListener(options);
      
      const stats = domAPI.getStats();
      expect(stats.eventListeners).toBe(1);
    });

    test('should include permissions and policy in stats', () => {
      const stats = domAPI.getStats();
      
      expect(stats.permissions).toEqual(permissions);
      expect(stats.policy).toEqual(policy);
    });
  });

  describe('Cleanup', () => {
    test('should clean up all resources', () => {
      // Create element and add event listener
      const options: ElementCreationOptions = { tagName: 'div' };
      domAPI.createElement(options);
      
      const handler = jest.fn();
      domAPI.addEventListener({
        element: mockElement,
        eventType: 'click',
        handler: handler
      });

      domAPI.cleanup();
      
      // Should remove event listeners and elements
      expect(mockElement.removeEventListener).toHaveBeenCalled();
      expect(mockElement.parentNode.removeChild).toHaveBeenCalled();
      
      const stats = domAPI.getStats();
      expect(stats.createdElements).toBe(0);
      expect(stats.eventListeners).toBe(0);
    });

    test('should handle cleanup errors gracefully', () => {
      mockElement.removeEventListener.mockImplementation(() => {
        throw new Error('Cleanup error');
      });

      expect(() => domAPI.cleanup()).not.toThrow();
    });

    test('should prevent operations after cleanup', () => {
      domAPI.cleanup();
      
      const options: ElementCreationOptions = { tagName: 'div' };
      const result = domAPI.createElement(options);
      
      expect(result).toBeNull();
    });
  });

  describe('Security Validation', () => {
    test('should block script execution when disabled', () => {
      const options: ElementCreationOptions = {
        tagName: 'div',
        innerHTML: '<script>alert("xss")</script>'
      };

      domAPI.createElement(options);
      
      // innerHTML should not be set when script execution is disabled
      expect(mockElement.innerHTML).toBeUndefined();
    });

    test('should sanitize URLs in attributes', () => {
      domAPI.setAttribute(mockElement, 'href', 'javascript:alert("xss")');
      
      expect(mockElement.setAttribute).toHaveBeenCalledWith('href', '#');
    });

    test('should handle attribute sanitization', () => {
      domAPI.setAttribute(mockElement, 'title', '<script>alert("xss")</script>');
      
      expect(mockElement.setAttribute).toHaveBeenCalledWith('title', 'scriptalert("xss")/script');
    });
  });

  describe('Error Handling', () => {
    test('should handle DOM operation errors', () => {
      mockElement.appendChild.mockImplementation(() => {
        throw new Error('DOM error');
      });

      const result = domAPI.appendChild(mockElement, mockElement);
      expect(result).toBe(false);
    });

    test('should handle style setting errors', () => {
      mockElement.style.setProperty.mockImplementation(() => {
        throw new Error('Style error');
      });

      const result = domAPI.setStyle(mockElement, 'color', 'red');
      expect(result).toBe(false);
    });
  });
});

describe('Factory Function', () => {
  test('should create SecureDOMAPI instance', () => {
    const permissions: JSPermissions = {
      executionMode: 'sandboxed',
      allowedAPIs: [],
      domAccess: 'write'
    };

    const api = createSecureDOMAPI(permissions);
    expect(api).toBeInstanceOf(SecureDOMAPI);
    api.cleanup();
  });
});