/**
 * LIV Document Format - Secure DOM API Wrapper
 * Provides controlled access to DOM operations based on security policy
 */

import { JSPermissions } from './sandbox-interface';

export interface DOMSecurityPolicy {
  allowedElements: string[];
  allowedAttributes: string[];
  allowedEvents: string[];
  allowedStyles: string[];
  maxElements: number;
  allowScriptExecution: boolean;
  allowFormSubmission: boolean;
  allowNavigation: boolean;
}

export interface ElementCreationOptions {
  tagName: string;
  attributes?: Record<string, string>;
  textContent?: string;
  innerHTML?: string;
  styles?: Record<string, string>;
}

export interface EventListenerOptions {
  element: Element;
  eventType: string;
  handler: EventListener;
  options?: AddEventListenerOptions;
}

/**
 * Secure DOM API that enforces security policies
 */
export class SecureDOMAPI {
  private permissions: JSPermissions;
  private policy: DOMSecurityPolicy;
  private createdElements: Set<Element>;
  private eventListeners: Map<Element, EventListenerOptions[]>;
  private isDestroyed: boolean = false;

  constructor(permissions: JSPermissions, policy?: DOMSecurityPolicy) {
    this.permissions = permissions;
    this.policy = policy || this.getDefaultPolicy();
    this.createdElements = new Set();
    this.eventListeners = new Map();

    this.validatePermissions();
  }

  /**
   * Securely query DOM elements
   */
  querySelector(selector: string): Element | null {
    this.ensureReadAccess();
    
    try {
      // Sanitize selector to prevent XSS
      const sanitizedSelector = this.sanitizeSelector(selector);
      return document.querySelector(sanitizedSelector);
    } catch (error) {
      this.handleError('querySelector failed', error);
      return null;
    }
  }

  /**
   * Securely query multiple DOM elements
   */
  querySelectorAll(selector: string): NodeList {
    this.ensureReadAccess();
    
    try {
      const sanitizedSelector = this.sanitizeSelector(selector);
      return document.querySelectorAll(sanitizedSelector);
    } catch (error) {
      this.handleError('querySelectorAll failed', error);
      return document.createDocumentFragment().childNodes;
    }
  }

  /**
   * Securely create DOM elements
   */
  createElement(options: ElementCreationOptions): Element | null {
    this.ensureWriteAccess();
    
    try {
      // Validate element creation
      this.validateElementCreation(options);
      
      const element = document.createElement(options.tagName);
      
      // Set attributes securely
      if (options.attributes) {
        this.setAttributes(element, options.attributes);
      }
      
      // Set text content securely
      if (options.textContent) {
        element.textContent = this.sanitizeText(options.textContent);
      }
      
      // Set innerHTML securely (if allowed)
      if (options.innerHTML && this.policy.allowScriptExecution) {
        element.innerHTML = this.sanitizeHTML(options.innerHTML);
      }
      
      // Set styles securely
      if (options.styles) {
        this.setStyles(element, options.styles);
      }
      
      this.createdElements.add(element);
      return element;
    } catch (error) {
      this.handleError('createElement failed', error);
      return null;
    }
  }

  /**
   * Securely append child elements
   */
  appendChild(parent: Element, child: Element): boolean {
    this.ensureWriteAccess();
    
    try {
      // Validate parent and child elements
      if (!this.isElementSafe(parent) || !this.isElementSafe(child)) {
        throw new Error('Unsafe element detected');
      }
      
      parent.appendChild(child);
      return true;
    } catch (error) {
      this.handleError('appendChild failed', error);
      return false;
    }
  }

  /**
   * Securely remove elements
   */
  removeElement(element: Element): boolean {
    this.ensureWriteAccess();
    
    try {
      if (element.parentNode) {
        element.parentNode.removeChild(element);
      }
      
      // Clean up tracking
      this.createdElements.delete(element);
      this.removeEventListeners(element);
      
      return true;
    } catch (error) {
      this.handleError('removeElement failed', error);
      return false;
    }
  }

  /**
   * Securely set element attributes
   */
  setAttribute(element: Element, name: string, value: string): boolean {
    this.ensureWriteAccess();
    
    try {
      // Validate attribute
      if (!this.isAttributeAllowed(name)) {
        throw new Error(`Attribute '${name}' not allowed`);
      }
      
      const sanitizedValue = this.sanitizeAttributeValue(name, value);
      element.setAttribute(name, sanitizedValue);
      
      return true;
    } catch (error) {
      this.handleError('setAttribute failed', error);
      return false;
    }
  }

  /**
   * Securely get element attributes
   */
  getAttribute(element: Element, name: string): string | null {
    this.ensureReadAccess();
    
    try {
      return element.getAttribute(name);
    } catch (error) {
      this.handleError('getAttribute failed', error);
      return null;
    }
  }

  /**
   * Securely add event listeners
   */
  addEventListener(options: EventListenerOptions): boolean {
    this.ensureWriteAccess();
    
    try {
      // Validate event type
      if (!this.isEventAllowed(options.eventType)) {
        throw new Error(`Event type '${options.eventType}' not allowed`);
      }
      
      // Wrap handler for security
      const secureHandler = this.wrapEventHandler(options.handler);
      
      options.element.addEventListener(
        options.eventType, 
        secureHandler, 
        options.options
      );
      
      // Track event listener
      if (!this.eventListeners.has(options.element)) {
        this.eventListeners.set(options.element, []);
      }
      
      this.eventListeners.get(options.element)!.push({
        ...options,
        handler: secureHandler
      });
      
      return true;
    } catch (error) {
      this.handleError('addEventListener failed', error);
      return false;
    }
  }

  /**
   * Securely remove event listeners
   */
  removeEventListener(element: Element, eventType: string, handler: EventListener): boolean {
    this.ensureWriteAccess();
    
    try {
      element.removeEventListener(eventType, handler);
      
      // Clean up tracking
      const listeners = this.eventListeners.get(element);
      if (listeners) {
        const index = listeners.findIndex(l => 
          l.eventType === eventType && l.handler === handler
        );
        if (index >= 0) {
          listeners.splice(index, 1);
        }
      }
      
      return true;
    } catch (error) {
      this.handleError('removeEventListener failed', error);
      return false;
    }
  }

  /**
   * Securely set element styles
   */
  setStyle(element: Element, property: string, value: string): boolean {
    this.ensureWriteAccess();
    
    try {
      // Validate style property
      if (!this.isStyleAllowed(property)) {
        throw new Error(`Style property '${property}' not allowed`);
      }
      
      const sanitizedValue = this.sanitizeStyleValue(property, value);
      (element as HTMLElement).style.setProperty(property, sanitizedValue);
      
      return true;
    } catch (error) {
      this.handleError('setStyle failed', error);
      return false;
    }
  }

  /**
   * Securely get computed styles
   */
  getComputedStyle(element: Element): CSSStyleDeclaration | null {
    this.ensureReadAccess();
    
    try {
      return window.getComputedStyle(element);
    } catch (error) {
      this.handleError('getComputedStyle failed', error);
      return null;
    }
  }

  /**
   * Get DOM statistics
   */
  getStats(): Record<string, any> {
    return {
      createdElements: this.createdElements.size,
      eventListeners: Array.from(this.eventListeners.values())
        .reduce((sum, listeners) => sum + listeners.length, 0),
      permissions: this.permissions,
      policy: this.policy
    };
  }

  /**
   * Clean up all created elements and event listeners
   */
  cleanup(): void {
    try {
      // Remove all event listeners
      for (const [element, listeners] of this.eventListeners) {
        for (const listener of listeners) {
          element.removeEventListener(listener.eventType, listener.handler, listener.options);
        }
      }
      
      // Remove all created elements
      for (const element of this.createdElements) {
        if (element.parentNode) {
          element.parentNode.removeChild(element);
        }
      }
      
      // Clear tracking
      this.createdElements.clear();
      this.eventListeners.clear();
      this.isDestroyed = true;
    } catch (error) {
      this.handleError('cleanup failed', error);
    }
  }

  // Private helper methods

  private validatePermissions(): void {
    if (this.permissions.domAccess === 'none') {
      throw new Error('DOM access not permitted');
    }
  }

  private ensureReadAccess(): void {
    if (this.isDestroyed) {
      throw new Error('DOM API has been destroyed');
    }
    
    if (this.permissions.domAccess === 'none') {
      throw new Error('DOM read access not permitted');
    }
  }

  private ensureWriteAccess(): void {
    if (this.isDestroyed) {
      throw new Error('DOM API has been destroyed');
    }
    
    if (this.permissions.domAccess !== 'write') {
      throw new Error('DOM write access not permitted');
    }
  }

  private validateElementCreation(options: ElementCreationOptions): void {
    // Check element limit
    if (this.createdElements.size >= this.policy.maxElements) {
      throw new Error(`Maximum elements limit reached: ${this.policy.maxElements}`);
    }
    
    // Check allowed elements
    if (!this.policy.allowedElements.includes(options.tagName.toLowerCase())) {
      throw new Error(`Element type '${options.tagName}' not allowed`);
    }
  }

  private setAttributes(element: Element, attributes: Record<string, string>): void {
    for (const [name, value] of Object.entries(attributes)) {
      if (this.isAttributeAllowed(name)) {
        const sanitizedValue = this.sanitizeAttributeValue(name, value);
        element.setAttribute(name, sanitizedValue);
      }
    }
  }

  private setStyles(element: Element, styles: Record<string, string>): void {
    for (const [property, value] of Object.entries(styles)) {
      if (this.isStyleAllowed(property)) {
        const sanitizedValue = this.sanitizeStyleValue(property, value);
        (element as HTMLElement).style.setProperty(property, sanitizedValue);
      }
    }
  }

  private isElementSafe(element: Element): boolean {
    // Check if element is in allowed list
    return this.policy.allowedElements.includes(element.tagName.toLowerCase());
  }

  private isAttributeAllowed(name: string): boolean {
    const lowerName = name.toLowerCase();
    
    // Block dangerous attributes
    const dangerousAttributes = ['onclick', 'onload', 'onerror', 'javascript:', 'vbscript:'];
    if (dangerousAttributes.some(attr => lowerName.includes(attr))) {
      return false;
    }
    
    return this.policy.allowedAttributes.includes(lowerName) || 
           this.policy.allowedAttributes.includes('*');
  }

  private isEventAllowed(eventType: string): boolean {
    return this.policy.allowedEvents.includes(eventType.toLowerCase()) ||
           this.policy.allowedEvents.includes('*');
  }

  private isStyleAllowed(property: string): boolean {
    const lowerProperty = property.toLowerCase();
    
    // Block dangerous style properties
    const dangerousProperties = ['expression', 'javascript:', 'vbscript:', 'behavior'];
    if (dangerousProperties.some(prop => lowerProperty.includes(prop))) {
      return false;
    }
    
    return this.policy.allowedStyles.includes(lowerProperty) ||
           this.policy.allowedStyles.includes('*');
  }

  private sanitizeSelector(selector: string): string {
    // Remove potentially dangerous characters
    return selector.replace(/[<>\"']/g, '');
  }

  private sanitizeText(text: string): string {
    // Escape HTML entities
    const div = document.createElement('div');
    div.textContent = text;
    return div.innerHTML;
  }

  private sanitizeHTML(html: string): string {
    // Basic HTML sanitization - in production, use a proper library like DOMPurify
    return html
      .replace(/<script[^>]*>.*?<\/script>/gi, '')
      .replace(/javascript:/gi, '')
      .replace(/vbscript:/gi, '')
      .replace(/on\w+\s*=/gi, '');
  }

  private sanitizeAttributeValue(name: string, value: string): string {
    // Sanitize based on attribute type
    if (name.toLowerCase() === 'href' || name.toLowerCase() === 'src') {
      // URL sanitization
      if (value.match(/^(javascript:|vbscript:|data:)/i)) {
        return '#';
      }
    }
    
    return value.replace(/[<>\"']/g, '');
  }

  private sanitizeStyleValue(property: string, value: string): string {
    // Remove potentially dangerous style values
    return value
      .replace(/expression\s*\(/gi, '')
      .replace(/javascript:/gi, '')
      .replace(/vbscript:/gi, '')
      .replace(/url\s*\(\s*[\"']?javascript:/gi, 'url(#')
      .replace(/url\s*\(\s*[\"']?vbscript:/gi, 'url(#');
  }

  private wrapEventHandler(handler: EventListener): EventListener {
    return (event: Event) => {
      try {
        // Add security checks here if needed
        handler(event);
      } catch (error) {
        this.handleError('Event handler error', error);
      }
    };
  }

  private removeEventListeners(element: Element): void {
    const listeners = this.eventListeners.get(element);
    if (listeners) {
      for (const listener of listeners) {
        element.removeEventListener(listener.eventType, listener.handler, listener.options);
      }
      this.eventListeners.delete(element);
    }
  }

  private getDefaultPolicy(): DOMSecurityPolicy {
    return {
      allowedElements: [
        'div', 'span', 'p', 'h1', 'h2', 'h3', 'h4', 'h5', 'h6',
        'ul', 'ol', 'li', 'a', 'img', 'canvas', 'svg',
        'table', 'tr', 'td', 'th', 'thead', 'tbody'
      ],
      allowedAttributes: [
        'id', 'class', 'style', 'title', 'alt', 'src', 'href',
        'width', 'height', 'data-*'
      ],
      allowedEvents: [
        'click', 'mouseover', 'mouseout', 'focus', 'blur',
        'keydown', 'keyup', 'change', 'input'
      ],
      allowedStyles: [
        'color', 'background-color', 'font-size', 'font-family',
        'margin', 'padding', 'border', 'width', 'height',
        'display', 'position', 'top', 'left', 'right', 'bottom'
      ],
      maxElements: 1000,
      allowScriptExecution: false,
      allowFormSubmission: false,
      allowNavigation: false
    };
  }

  private handleError(message: string, error: any): void {
    console.error(`[SecureDOMAPI] ${message}:`, error);
  }
}

/**
 * Factory function to create a secure DOM API instance
 */
export function createSecureDOMAPI(permissions: JSPermissions, policy?: DOMSecurityPolicy): SecureDOMAPI {
  return new SecureDOMAPI(permissions, policy);
}