// LIV Secure Content Rendering Engine

import { 
  RenderUpdate, 
  DOMOperation, 
  StyleChange, 
  InteractionEvent, 
  RendererOptions,
  LegacySecurityPolicy,
  Position,
  LIVDocument,
  ValidationResult
} from './types';
import {
  LIVError,
  LIVErrorType,
  SecurityError,
  ValidationError,
  ErrorHandler
} from './errors';

export interface SecureRenderingOptions extends RendererOptions {
  enableFallback?: boolean;
  strictSecurity?: boolean;
  maxRenderTime?: number;
  errorHandler?: ErrorHandler;
  enableAnimations?: boolean;
  targetFPS?: number;
  enableSVG?: boolean;
  enableResponsiveDesign?: boolean;
}

export class LIVRenderer {
  private container: HTMLElement;
  private permissions: LegacySecurityPolicy;
  private wasmModule: any; // Will be loaded dynamically
  private sandbox: SandboxedDOM;
  private animationFrameId?: number;
  private document?: LIVDocument;
  private renderingState: RenderingState;
  private errorHandler: ErrorHandler;
  private options: SecureRenderingOptions;
  private animationEngine: AnimationEngine;
  private svgRenderer: SVGRenderer;
  private responsiveManager: ResponsiveManager;

  constructor(options: SecureRenderingOptions) {
    this.container = options.container;
    this.permissions = options.permissions;
    this.options = {
      enableFallback: true,
      strictSecurity: true,
      maxRenderTime: 5000, // 5 seconds
      enableAnimations: true,
      targetFPS: 60,
      enableSVG: true,
      enableResponsiveDesign: true,
      ...options
    };
    this.errorHandler = options.errorHandler || ErrorHandler.getInstance();
    this.renderingState = new RenderingState();
    this.sandbox = new SandboxedDOM(this.container, this.permissions, this.errorHandler);
    
    // Initialize animation and graphics engines with mobile optimization
    this.animationEngine = new AnimationEngine(this.options.targetFPS || 60);
    this.svgRenderer = new SVGRenderer(this.errorHandler);
    this.responsiveManager = new ResponsiveManager(this.container);
    
    // Initialize container with mobile optimizations
    this.initializeContainer();
    this.initializeMobileOptimizations();
  }

  private async initializeMobileOptimizations(): Promise<void> {
    try {
      // Inject mobile-specific styles
      const { injectMobileStyles } = await import('./mobile-styles');
      injectMobileStyles();
      
      // Set up mobile interaction handling
      this.setupMobileInteractionHandling();
      
    } catch (error) {
      console.warn('Failed to initialize mobile optimizations:', error);
    }
  }

  private setupMobileInteractionHandling(): void {
    // Listen for mobile interaction events from the responsive manager
    this.container.addEventListener('liv-interaction', (event: CustomEvent) => {
      const interactionEvent = event.detail;
      this.handleUserEvent(interactionEvent);
    });
    
    // Listen for orientation changes
    this.container.addEventListener('liv-orientation-change', (event: CustomEvent) => {
      this.handleOrientationChange(event.detail.orientation);
    });
    
    // Listen for performance changes
    this.container.addEventListener('liv-performance-change', (event: CustomEvent) => {
      this.handlePerformanceChange(event.detail);
    });
  }

  private handleOrientationChange(orientation: string): void {
    // Pause animations during orientation change to prevent glitches
    this.stopRenderLoop();
    
    // Wait for orientation change to complete
    setTimeout(() => {
      // Restart animations if they were enabled
      if (this.options.enableAnimations && this.document?.manifest.features?.animations) {
        this.startRenderLoop();
      }
      
      // Trigger re-layout
      this.responsiveManager.initializeClasses();
      
    }, 200);
  }

  private handlePerformanceChange(performanceData: any): void {
    // Adapt rendering based on performance metrics
    if (performanceData.frameRate < 20) {
      // Very poor performance - enable battery mode
      this.container.classList.add('liv-performance-battery');
      this.container.classList.remove('liv-performance-balanced');
      
      // Reduce animation quality
      this.animationEngine.setPerformanceMode('battery');
      
    } else if (performanceData.frameRate < 40) {
      // Moderate performance - enable balanced mode
      this.container.classList.add('liv-performance-balanced');
      this.container.classList.remove('liv-performance-battery');
      
      this.animationEngine.setPerformanceMode('balanced');
      
    } else {
      // Good performance - enable high quality
      this.container.classList.remove('liv-performance-battery', 'liv-performance-balanced');
      
      this.animationEngine.setPerformanceMode('high');
    }
  }

  private initializeContainer(): void {
    // Set up secure container with CSP
    this.container.style.position = 'relative';
    this.container.style.overflow = 'hidden';
    
    // Apply Content Security Policy
    if (this.permissions.contentSecurityPolicy) {
      const meta = document.createElement('meta');
      meta.httpEquiv = 'Content-Security-Policy';
      meta.content = this.permissions.contentSecurityPolicy;
      document.head.appendChild(meta);
    }
  }

  async loadWASMModule(wasmModule: any): Promise<void> {
    this.wasmModule = wasmModule;
    
    // Initialize WASM module with permissions
    if (this.wasmModule && this.wasmModule.init_interactive_engine) {
      this.wasmModule.init_interactive_engine();
    }
  }

  applyRenderUpdate(update: RenderUpdate): void {
    // Apply DOM operations
    for (const operation of update.domOperations) {
      this.applyDOMOperation(operation);
    }

    // Apply style changes
    for (const styleChange of update.styleChanges) {
      this.applyStyleChange(styleChange);
    }

    // Update animations
    for (const animationUpdate of update.animationUpdates) {
      this.updateAnimation(animationUpdate);
    }
  }

  private applyDOMOperation(operation: DOMOperation): void {
    switch (operation.type) {
      case 'Create':
        this.sandbox.createElement(operation.elementId, operation.tag, operation.parentId);
        break;
      case 'Update':
        this.sandbox.updateElement(operation.elementId, operation.attributes);
        break;
      case 'Remove':
        this.sandbox.removeElement(operation.elementId);
        break;
      case 'Move':
        this.sandbox.moveElement(operation.elementId, operation.newParentId, operation.index);
        break;
    }
  }

  private applyStyleChange(styleChange: StyleChange): void {
    this.sandbox.setStyle(styleChange.elementId, styleChange.property, styleChange.value);
  }

  private updateAnimation(animationUpdate: any): void {
    // Apply animation frame updates
    const element = this.sandbox.getElement(animationUpdate.animationId);
    if (element) {
      for (const [property, value] of Object.entries(animationUpdate.currentValues)) {
        this.sandbox.setStyle(animationUpdate.animationId, property, String(value));
      }
    }
  }

  handleUserEvent(event: Event): void {
    // Convert DOM event to InteractionEvent
    const interactionEvent = this.convertToInteractionEvent(event);
    
    // Pass event to WASM layer for processing
    if (this.wasmModule && this.wasmModule.process_interaction) {
      try {
        const updateData = this.wasmModule.process_interaction(JSON.stringify(interactionEvent));
        const renderUpdate: RenderUpdate = JSON.parse(updateData);
        this.applyRenderUpdate(renderUpdate);
      } catch (error) {
        console.error('Error processing interaction:', error);
      }
    }
  }

  private convertToInteractionEvent(event: Event): InteractionEvent {
    const rect = this.container.getBoundingClientRect();
    let position: Position | undefined;
    
    if (event instanceof MouseEvent) {
      position = {
        x: event.clientX - rect.left,
        y: event.clientY - rect.top
      };
    }

    return {
      eventType: this.getInteractionType(event),
      targetElement: (event.target as Element)?.id,
      position,
      data: this.extractEventData(event),
      timestamp: performance.now()
    };
  }

  private getInteractionType(event: Event): any {
    switch (event.type) {
      case 'click': return 'Click';
      case 'mouseover': return 'Hover';
      case 'touchstart': return 'Touch';
      case 'scroll': return 'Scroll';
      case 'keydown': return 'Keyboard';
      case 'resize': return 'Resize';
      default: return 'Click';
    }
  }

  private extractEventData(event: Event): Record<string, any> {
    const data: Record<string, any> = {};
    
    if (event instanceof MouseEvent) {
      data.button = event.button;
      data.ctrlKey = event.ctrlKey;
      data.shiftKey = event.shiftKey;
      data.altKey = event.altKey;
    }
    
    if (event instanceof KeyboardEvent) {
      data.key = event.key;
      data.code = event.code;
      data.ctrlKey = event.ctrlKey;
      data.shiftKey = event.shiftKey;
      data.altKey = event.altKey;
    }
    
    return data;
  }

  startRenderLoop(): void {
    if (!this.options.enableAnimations) {
      return;
    }

    const renderFrame = (timestamp: number) => {
      try {
        // Update animation engine
        this.animationEngine.update(timestamp);
        
        // Process WASM render updates
        if (this.wasmModule && this.wasmModule.render_frame) {
          const updateData = this.wasmModule.render_frame(timestamp);
          if (updateData) {
            const renderUpdate: RenderUpdate = JSON.parse(updateData);
            this.applyRenderUpdate(renderUpdate);
          }
        }
        
        // Apply animation updates
        const animationUpdates = this.animationEngine.getUpdates();
        if (animationUpdates.length > 0) {
          this.applyAnimationUpdates(animationUpdates);
        }
        
        // Update performance metrics
        this.renderingState.updateFrameCount();
        
      } catch (error) {
        console.error('Error in render frame:', error);
        this.errorHandler.handleError(new LIVError(
          LIVErrorType.VALIDATION,
          'Animation render error',
          error instanceof Error ? error : new Error(String(error))
        ));
      }
      
      this.animationFrameId = requestAnimationFrame(renderFrame);
    };
    
    this.animationFrameId = requestAnimationFrame(renderFrame);
  }

  private applyAnimationUpdates(updates: AnimationUpdate[]): void {
    for (const update of updates) {
      this.sandbox.applyAnimationUpdate(update);
    }
  }

  stopRenderLoop(): void {
    if (this.animationFrameId !== undefined) {
      cancelAnimationFrame(this.animationFrameId);
      this.animationFrameId = undefined;
    }
  }

  // New method to render a LIV document
  async renderDocument(document: LIVDocument): Promise<void> {
    this.document = document;
    
    try {
      // Validate document before rendering
      const validation = document.validate();
      if (!validation.isValid && this.options.strictSecurity) {
        throw new ValidationError('Document validation failed', validation.errors, validation.warnings);
      }

      // Generate security report
      const securityReport = document.generateSecurityReport();
      if (!securityReport.isValid && this.options.strictSecurity) {
        throw new SecurityError('Security validation failed');
      }

      // Clear previous content
      this.sandbox.clear();
      this.renderingState.reset();

      // Try to render interactive content first
      const renderSuccess = await this.tryRenderInteractiveContent(document);
      
      if (!renderSuccess && this.options.enableFallback) {
        // Fall back to static content
        await this.renderStaticFallback(document);
      }

      this.renderingState.setRenderingComplete(true);
      
    } catch (error) {
      // Re-throw validation and security errors in strict mode
      if (error instanceof ValidationError || error instanceof SecurityError) {
        if (this.options.strictSecurity) {
          throw error;
        }
      }
      
      this.handleRenderingError(error);
      
      if (this.options.enableFallback) {
        await this.renderStaticFallback(document);
      }
    }
  }

  private async tryRenderInteractiveContent(document: LIVDocument): Promise<boolean> {
    try {
      // Check if interactive content is allowed by security policy
      if (this.permissions.jsPermissions.executionMode === 'none') {
        return false; // Force fallback mode
      }

      // Set rendering timeout
      const timeoutPromise = new Promise<never>((_, reject) => {
        setTimeout(() => reject(new Error('Rendering timeout')), this.options.maxRenderTime);
      });

      const renderPromise = this.performInteractiveRendering(document);
      
      await Promise.race([renderPromise, timeoutPromise]);
      return true;
      
    } catch (error) {
      this.errorHandler.handleError(new LIVError(
        LIVErrorType.VALIDATION,
        'Interactive rendering failed',
        error instanceof Error ? error : new Error(String(error))
      ));
      return false;
    }
  }

  private async performInteractiveRendering(document: LIVDocument): Promise<void> {
    // Inject HTML content into sandbox with SVG support
    await this.sandbox.setContent(document.content.html, { 
      interactive: true,
      enableSVG: this.options.enableSVG 
    });
    
    // Apply CSS styles with animation support
    await this.sandbox.applyStyles(document.content.css, {
      enableAnimations: this.options.enableAnimations
    });
    
    // Initialize SVG renderer if SVG content is detected
    if (this.options.enableSVG && this.containsSVG(document.content.html)) {
      await this.svgRenderer.initialize(this.sandbox);
    }
    
    // Set up responsive design if enabled
    if (this.options.enableResponsiveDesign) {
      this.responsiveManager.initialize();
      this.responsiveManager.initializeClasses();
    }
    
    // Load and initialize WASM module if available
    if (document.content.interactiveSpec && this.wasmModule) {
      await this.initializeInteractiveContent(document.content.interactiveSpec);
    }
    
    // Initialize animations if enabled
    if (this.options.enableAnimations && document.manifest.features?.animations) {
      await this.initializeAnimations(document.content.css);
    }
    
    // Set up event listeners
    this.setupEventListeners();
    
    // Start render loop for animations
    if (this.options.enableAnimations && document.manifest.features?.animations) {
      this.startRenderLoop();
    }
  }

  private containsSVG(html: string): boolean {
    return /<svg[\s\S]*?<\/svg>/i.test(html) || html.includes('<svg');
  }

  private async initializeAnimations(css: string): Promise<void> {
    try {
      // Parse CSS for animation definitions
      const animations = this.parseAnimationsFromCSS(css);
      
      // Register animations with the animation engine
      for (const animation of animations) {
        this.animationEngine.registerAnimation(animation);
      }
      
    } catch (error) {
      this.errorHandler.handleError(new LIVError(
        LIVErrorType.VALIDATION,
        'Failed to initialize animations',
        error instanceof Error ? error : new Error(String(error))
      ));
    }
  }

  private parseAnimationsFromCSS(css: string): CSSAnimation[] {
    const animations: CSSAnimation[] = [];
    
    // Parse @keyframes rules
    const keyframesRegex = /@keyframes\s+([^{]+)\s*{([^}]+)}/gi;
    let match;
    
    while ((match = keyframesRegex.exec(css)) !== null) {
      const name = match[1].trim();
      const keyframes = match[2];
      
      animations.push({
        name,
        keyframes: this.parseKeyframes(keyframes),
        duration: 1000, // Default 1 second
        easing: 'ease',
        iterations: 1
      });
    }
    
    // Parse animation properties
    const animationRegex = /animation(?:-name)?:\s*([^;]+);/gi;
    while ((match = animationRegex.exec(css)) !== null) {
      const animationProps = match[1].trim();
      // Parse animation shorthand properties
      // This is a simplified parser - a full implementation would be more comprehensive
    }
    
    return animations;
  }

  private parseKeyframes(keyframesStr: string): Keyframe[] {
    const keyframes: Keyframe[] = [];
    const frameRegex = /(\d+%|from|to)\s*{([^}]+)}/gi;
    let match;
    
    while ((match = frameRegex.exec(keyframesStr)) !== null) {
      const offset = this.parseOffset(match[1]);
      const styles = this.parseStyles(match[2]);
      
      keyframes.push({ offset, styles });
    }
    
    return keyframes.sort((a, b) => a.offset - b.offset);
  }

  private parseOffset(offsetStr: string): number {
    if (offsetStr === 'from') return 0;
    if (offsetStr === 'to') return 1;
    return parseInt(offsetStr) / 100;
  }

  private parseStyles(stylesStr: string): Record<string, string> {
    const styles: Record<string, string> = {};
    const declarations = stylesStr.split(';');
    
    for (const declaration of declarations) {
      const [property, value] = declaration.split(':').map(s => s.trim());
      if (property && value) {
        styles[property] = value;
      }
    }
    
    return styles;
  }

  private async renderStaticFallback(document: LIVDocument): Promise<void> {
    try {
      this.renderingState.setFallbackMode(true);
      
      // Use static fallback content if available
      const fallbackContent = document.content.staticFallback || document.content.html;
      
      // Render static content without interactive features
      await this.sandbox.setContent(fallbackContent, { interactive: false });
      
      // Apply basic CSS (filtered for security)
      const safeCss = this.sanitizeCSS(document.content.css);
      await this.sandbox.applyStyles(safeCss);
      
      this.renderingState.setRenderingComplete(true);
      
    } catch (error) {
      this.handleRenderingError(error);
      
      // Last resort: show error message
      this.renderErrorMessage('Failed to render document content');
    }
  }

  private async initializeInteractiveContent(interactiveSpec: string): Promise<void> {
    if (!this.wasmModule) {
      throw new Error('WASM module not loaded');
    }

    try {
      // Parse interactive specification
      const spec = JSON.parse(interactiveSpec);
      
      // Initialize WASM module with spec
      if (this.wasmModule.init_interactive_engine) {
        await this.wasmModule.init_interactive_engine(spec);
      }
      
      this.renderingState.setInteractiveMode(true);
      
    } catch (error) {
      throw new Error(`Failed to initialize interactive content: ${error.message}`);
    }
  }

  private setupEventListeners(): void {
    // Remove existing listeners
    this.removeEventListeners();
    
    // Add secure event listeners
    const events = ['click', 'mouseover', 'touchstart', 'scroll', 'keydown'];
    
    events.forEach(eventType => {
      this.container.addEventListener(eventType, this.handleUserEvent.bind(this), {
        passive: true,
        capture: false
      });
    });
  }

  private removeEventListeners(): void {
    const events = ['click', 'mouseover', 'touchstart', 'scroll', 'keydown'];
    
    events.forEach(eventType => {
      this.container.removeEventListener(eventType, this.handleUserEvent.bind(this));
    });
  }

  private sanitizeCSS(css: string): string {
    // Remove potentially dangerous CSS
    const dangerousPatterns = [
      /javascript:/gi,
      /expression\s*\(/gi,
      /behavior\s*:/gi,
      /binding\s*:/gi,
      /@import/gi,
      /url\s*\(\s*["']?javascript:/gi
    ];
    
    let sanitized = css;
    dangerousPatterns.forEach(pattern => {
      sanitized = sanitized.replace(pattern, '/* removed */');
    });
    
    return sanitized;
  }

  private handleRenderingError(error: any): void {
    const livError = error instanceof LIVError ? error : new LIVError(
      LIVErrorType.VALIDATION,
      'Rendering error occurred',
      error instanceof Error ? error : new Error(String(error))
    );
    
    this.errorHandler.handleError(livError);
    this.renderingState.addError(livError);
  }

  private renderErrorMessage(message: string): void {
    this.sandbox.clear();
    
    const errorHtml = `
      <div style="
        padding: 20px;
        background: #f8f9fa;
        border: 1px solid #dee2e6;
        border-radius: 4px;
        color: #495057;
        font-family: -apple-system, BlinkMacSystemFont, 'Segoe UI', Roboto, sans-serif;
        text-align: center;
      ">
        <h3 style="margin: 0 0 10px 0; color: #dc3545;">Rendering Error</h3>
        <p style="margin: 0;">${this.escapeHtml(message)}</p>
      </div>
    `;
    
    this.sandbox.setContent(errorHtml, { interactive: false });
  }

  private escapeHtml(text: string): string {
    const div = document.createElement('div');
    div.textContent = text;
    return div.innerHTML;
  }

  // Enhanced methods
  getRenderingState(): RenderingState {
    return this.renderingState;
  }

  getDocument(): LIVDocument | undefined {
    return this.document;
  }

  isInteractive(): boolean {
    return this.renderingState.isInteractive();
  }

  isFallbackMode(): boolean {
    return this.renderingState.isFallbackMode();
  }

  getPerformanceMetrics(): PerformanceMetrics {
    return this.renderingState.getPerformanceMetrics();
  }

  destroy(): void {
    this.stopRenderLoop();
    this.removeEventListeners();
    this.sandbox.destroy();
    this.renderingState.reset();
    
    // Clean up animation engine
    this.animationEngine.getUpdates(); // Clear any pending updates
    
    // Clean up responsive manager
    this.responsiveManager.destroy();
  }
}

// Rendering state management
class RenderingState {
  private isComplete: boolean = false;
  private isInteractiveMode: boolean = false;
  private isFallback: boolean = false;
  private errors: LIVError[] = [];
  private startTime: number = 0;
  private endTime: number = 0;
  private frameCount: number = 0;
  private lastFrameTime: number = 0;

  reset(): void {
    this.isComplete = false;
    this.isInteractiveMode = false;
    this.isFallback = false;
    this.errors = [];
    this.startTime = performance.now();
    this.endTime = 0;
    this.frameCount = 0;
    this.lastFrameTime = 0;
  }

  setRenderingComplete(complete: boolean): void {
    this.isComplete = complete;
    if (complete) {
      this.endTime = performance.now();
    }
  }

  setInteractiveMode(interactive: boolean): void {
    this.isInteractiveMode = interactive;
  }

  setFallbackMode(fallback: boolean): void {
    this.isFallback = fallback;
  }

  addError(error: LIVError): void {
    this.errors.push(error);
  }

  updateFrameCount(): void {
    this.frameCount++;
    this.lastFrameTime = performance.now();
  }

  isRenderingComplete(): boolean {
    return this.isComplete;
  }

  isInteractive(): boolean {
    return this.isInteractiveMode;
  }

  isFallbackMode(): boolean {
    return this.isFallback;
  }

  getErrors(): LIVError[] {
    return [...this.errors];
  }

  getRenderTime(): number {
    return this.endTime > 0 ? this.endTime - this.startTime : performance.now() - this.startTime;
  }

  getPerformanceMetrics(): PerformanceMetrics {
    const currentTime = performance.now();
    const totalTime = this.endTime > 0 ? this.endTime - this.startTime : currentTime - this.startTime;
    const timeInSeconds = totalTime / 1000;
    const fps = this.frameCount > 0 && timeInSeconds > 0 ? (this.frameCount / timeInSeconds) : 0;

    return {
      renderTime: totalTime,
      frameCount: this.frameCount,
      averageFPS: Math.min(fps, 60), // Cap at 60 FPS for realistic values
      isComplete: this.isComplete,
      isInteractive: this.isInteractiveMode,
      isFallback: this.isFallback,
      errorCount: this.errors.length
    };
  }
}

// Performance metrics interface
export interface PerformanceMetrics {
  renderTime: number;
  frameCount: number;
  averageFPS: number;
  isComplete: boolean;
  isInteractive: boolean;
  isFallback: boolean;
  errorCount: number;
}

// Animation interfaces
export interface CSSAnimation {
  name: string;
  keyframes: Keyframe[];
  duration: number;
  easing: string;
  iterations: number;
  delay?: number;
}

export interface Keyframe {
  offset: number;
  styles: Record<string, string>;
}

export interface AnimationUpdate {
  elementId: string;
  animationName: string;
  progress: number;
  styles: Record<string, string>;
}

// Enhanced Animation Engine with Mobile Optimizations
class AnimationEngine {
  private animations: Map<string, CSSAnimation> = new Map();
  private activeAnimations: Map<string, ActiveAnimation> = new Map();
  private targetFPS: number;
  private lastFrameTime: number = 0;
  private frameInterval: number;
  private isMobile: boolean = false;
  private performanceMode: 'high' | 'balanced' | 'battery' = 'balanced';
  private frameSkipCount: number = 0;
  private maxFrameSkip: number = 2;

  constructor(targetFPS: number = 60) {
    this.targetFPS = targetFPS;
    this.frameInterval = 1000 / targetFPS;
    this.detectMobileDevice();
    this.adaptToDevice();
  }

  private detectMobileDevice(): void {
    // Enhanced mobile detection for animation optimization
    const hasTouch = 'ontouchstart' in window || navigator.maxTouchPoints > 0;
    const userAgent = navigator.userAgent.toLowerCase();
    const mobileKeywords = ['mobile', 'android', 'iphone', 'ipad'];
    const hasMobileUA = mobileKeywords.some(keyword => userAgent.includes(keyword));
    
    // Check for battery API to detect mobile devices
    const hasBattery = 'getBattery' in navigator;
    
    // Check for reduced motion preference
    const prefersReducedMotion = window.matchMedia('(prefers-reduced-motion: reduce)').matches;
    
    this.isMobile = hasTouch || hasMobileUA || hasBattery;
    
    if (prefersReducedMotion) {
      this.performanceMode = 'battery';
    }
  }

  private adaptToDevice(): void {
    if (this.isMobile) {
      // Reduce target FPS on mobile for better battery life
      this.targetFPS = Math.min(this.targetFPS, 30);
      this.frameInterval = 1000 / this.targetFPS;
      this.maxFrameSkip = 3; // Allow more frame skipping on mobile
      
      // Monitor battery level if available
      this.monitorBatteryLevel();
      
      // Monitor performance and adapt
      this.monitorPerformance();
    }
  }

  private monitorBatteryLevel(): void {
    if ('getBattery' in navigator) {
      (navigator as any).getBattery().then((battery: any) => {
        const updatePerformanceMode = () => {
          if (battery.level < 0.2) {
            this.performanceMode = 'battery';
            this.targetFPS = 15; // Very low FPS for battery saving
          } else if (battery.level < 0.5) {
            this.performanceMode = 'balanced';
            this.targetFPS = 24; // Moderate FPS
          } else {
            this.performanceMode = 'high';
            this.targetFPS = this.isMobile ? 30 : 60;
          }
          
          this.frameInterval = 1000 / this.targetFPS;
        };
        
        updatePerformanceMode();
        battery.addEventListener('levelchange', updatePerformanceMode);
      });
    }
  }

  private monitorPerformance(): void {
    let frameCount = 0;
    let lastCheck = performance.now();
    
    const checkPerformance = () => {
      frameCount++;
      
      if (frameCount % 60 === 0) { // Check every 60 frames
        const now = performance.now();
        const actualFPS = 60000 / (now - lastCheck);
        lastCheck = now;
        
        // Adapt performance based on actual FPS
        if (actualFPS < this.targetFPS * 0.8) {
          // Performance is poor, reduce quality
          if (this.performanceMode === 'high') {
            this.performanceMode = 'balanced';
            this.targetFPS = Math.max(15, this.targetFPS * 0.75);
          } else if (this.performanceMode === 'balanced') {
            this.performanceMode = 'battery';
            this.targetFPS = Math.max(10, this.targetFPS * 0.5);
          }
          
          this.frameInterval = 1000 / this.targetFPS;
        }
      }
      
      requestAnimationFrame(checkPerformance);
    };
    
    requestAnimationFrame(checkPerformance);
  }

  registerAnimation(animation: CSSAnimation): void {
    this.animations.set(animation.name, animation);
  }

  startAnimation(elementId: string, animationName: string): void {
    const animation = this.animations.get(animationName);
    if (!animation) return;

    this.activeAnimations.set(`${elementId}-${animationName}`, {
      elementId,
      animation,
      startTime: performance.now(),
      progress: 0,
      isComplete: false
    });
  }

  stopAnimation(elementId: string, animationName: string): void {
    this.activeAnimations.delete(`${elementId}-${animationName}`);
  }

  update(timestamp: number): void {
    // Mobile-optimized frame rate control
    const timeSinceLastFrame = timestamp - this.lastFrameTime;
    
    if (timeSinceLastFrame < this.frameInterval) {
      // Skip frame to maintain target FPS
      this.frameSkipCount++;
      
      // On mobile, allow more aggressive frame skipping
      if (this.isMobile && this.frameSkipCount < this.maxFrameSkip) {
        return;
      }
      
      // For desktop, be more strict about frame timing
      if (!this.isMobile) {
        return;
      }
    }

    this.lastFrameTime = timestamp;
    this.frameSkipCount = 0;

    // Process animations with mobile optimizations
    const animationsToUpdate = Array.from(this.activeAnimations.entries());
    
    // Limit concurrent animations on mobile for performance
    const maxConcurrentAnimations = this.isMobile ? 5 : 20;
    const animationsSlice = animationsToUpdate.slice(0, maxConcurrentAnimations);

    for (const [key, activeAnimation] of animationsSlice) {
      const elapsed = timestamp - activeAnimation.startTime;
      const progress = Math.min(elapsed / activeAnimation.animation.duration, 1);

      activeAnimation.progress = progress;

      // Apply mobile-specific easing optimizations
      if (this.isMobile && this.performanceMode === 'battery') {
        // Use simpler easing functions on mobile in battery mode
        activeAnimation.animation.easing = this.simplifyEasing(activeAnimation.animation.easing);
      }

      if (progress >= 1) {
        if (activeAnimation.animation.iterations === Infinity || 
            activeAnimation.animation.iterations > 1) {
          // Restart animation
          activeAnimation.startTime = timestamp;
          activeAnimation.progress = 0;
          if (activeAnimation.animation.iterations !== Infinity) {
            activeAnimation.animation.iterations--;
          }
        } else {
          activeAnimation.isComplete = true;
          this.activeAnimations.delete(key);
        }
      }
    }
    
    // Clean up completed animations more aggressively on mobile
    if (this.isMobile && this.activeAnimations.size > maxConcurrentAnimations) {
      this.cleanupExcessAnimations(maxConcurrentAnimations);
    }
  }

  private simplifyEasing(easing: string): string {
    // Convert complex easing functions to simpler ones for mobile performance
    const simplifiedEasings: Record<string, string> = {
      'cubic-bezier(0.25, 0.46, 0.45, 0.94)': 'ease-out',
      'cubic-bezier(0.55, 0.055, 0.675, 0.19)': 'ease-in',
      'cubic-bezier(0.645, 0.045, 0.355, 1)': 'ease-in-out',
      'cubic-bezier(0.19, 1, 0.22, 1)': 'ease-out'
    };
    
    return simplifiedEasings[easing] || 'ease';
  }

  private cleanupExcessAnimations(maxAnimations: number): void {
    // Remove oldest animations that exceed the limit
    const animations = Array.from(this.activeAnimations.entries());
    const sortedByAge = animations.sort((a, b) => a[1].startTime - b[1].startTime);
    
    for (let i = maxAnimations; i < sortedByAge.length; i++) {
      this.activeAnimations.delete(sortedByAge[i][0]);
    }
  }

  getUpdates(): AnimationUpdate[] {
    const updates: AnimationUpdate[] = [];

    for (const activeAnimation of this.activeAnimations.values()) {
      const styles = this.interpolateStyles(
        activeAnimation.animation.keyframes,
        activeAnimation.progress
      );

      updates.push({
        elementId: activeAnimation.elementId,
        animationName: activeAnimation.animation.name,
        progress: activeAnimation.progress,
        styles
      });
    }

    return updates;
  }

  // Mobile-specific performance methods
  setPerformanceMode(mode: 'high' | 'balanced' | 'battery'): void {
    this.performanceMode = mode;
    
    switch (mode) {
      case 'battery':
        this.targetFPS = 15;
        this.maxFrameSkip = 5;
        break;
      case 'balanced':
        this.targetFPS = this.isMobile ? 24 : 30;
        this.maxFrameSkip = 3;
        break;
      case 'high':
        this.targetFPS = this.isMobile ? 30 : 60;
        this.maxFrameSkip = 1;
        break;
    }
    
    this.frameInterval = 1000 / this.targetFPS;
  }

  getPerformanceMode(): 'high' | 'balanced' | 'battery' {
    return this.performanceMode;
  }

  pauseAnimations(): void {
    // Pause all active animations
    for (const animation of this.activeAnimations.values()) {
      animation.isPaused = true;
      animation.pauseTime = performance.now();
    }
  }

  resumeAnimations(): void {
    const now = performance.now();
    
    // Resume all paused animations
    for (const animation of this.activeAnimations.values()) {
      if (animation.isPaused && animation.pauseTime) {
        const pauseDuration = now - animation.pauseTime;
        animation.startTime += pauseDuration;
        animation.isPaused = false;
        animation.pauseTime = undefined;
      }
    }
  }

  private interpolateStyles(keyframes: Keyframe[], progress: number): Record<string, string> {
    if (keyframes.length === 0) return {};
    if (keyframes.length === 1) return keyframes[0].styles;

    // Find the two keyframes to interpolate between
    let fromFrame = keyframes[0];
    let toFrame = keyframes[keyframes.length - 1];

    for (let i = 0; i < keyframes.length - 1; i++) {
      if (progress >= keyframes[i].offset && progress <= keyframes[i + 1].offset) {
        fromFrame = keyframes[i];
        toFrame = keyframes[i + 1];
        break;
      }
    }

    // Calculate local progress between the two keyframes
    const localProgress = (progress - fromFrame.offset) / (toFrame.offset - fromFrame.offset);

    // Interpolate styles
    const interpolatedStyles: Record<string, string> = {};
    const allProperties = new Set([...Object.keys(fromFrame.styles), ...Object.keys(toFrame.styles)]);

    for (const property of allProperties) {
      const fromValue = fromFrame.styles[property];
      const toValue = toFrame.styles[property];

      if (fromValue && toValue) {
        interpolatedStyles[property] = this.interpolateValue(fromValue, toValue, localProgress);
      } else {
        interpolatedStyles[property] = toValue || fromValue;
      }
    }

    return interpolatedStyles;
  }

  private interpolateValue(from: string, to: string, progress: number): string {
    // Simple numeric interpolation
    const fromNum = parseFloat(from);
    const toNum = parseFloat(to);

    if (!isNaN(fromNum) && !isNaN(toNum)) {
      const interpolated = fromNum + (toNum - fromNum) * progress;
      const unit = from.replace(/[\d.-]/g, '');
      return `${interpolated}${unit}`;
    }

    // For non-numeric values, use step interpolation
    return progress < 0.5 ? from : to;
  }
}

interface ActiveAnimation {
  elementId: string;
  animation: CSSAnimation;
  startTime: number;
  progress: number;
  isComplete: boolean;
  isPaused?: boolean;
  pauseTime?: number;
}

// SVG Renderer
class SVGRenderer {
  private errorHandler: ErrorHandler;
  private svgElements: Map<string, SVGElement> = new Map();

  constructor(errorHandler: ErrorHandler) {
    this.errorHandler = errorHandler;
  }

  async initialize(sandbox: SandboxedDOM): Promise<void> {
    try {
      // Find and process SVG elements
      const svgElements = sandbox.querySelectorAll('svg');
      
      for (const svg of svgElements) {
        await this.processSVGElement(svg as SVGElement);
      }
      
    } catch (error) {
      this.errorHandler.handleError(new LIVError(
        LIVErrorType.VALIDATION,
        'Failed to initialize SVG renderer',
        error instanceof Error ? error : new Error(String(error))
      ));
    }
  }

  private async processSVGElement(svg: SVGElement): Promise<void> {
    // Sanitize SVG content
    this.sanitizeSVG(svg);
    
    // Apply security restrictions
    this.applySVGSecurity(svg);
    
    // Enable SVG animations if supported
    this.enableSVGAnimations(svg);
    
    // Store reference
    if (svg.id) {
      this.svgElements.set(svg.id, svg);
    }
  }

  private sanitizeSVG(svg: SVGElement): void {
    // Remove dangerous SVG elements and attributes
    const dangerousElements = ['script', 'foreignObject', 'use'];
    const dangerousAttributes = ['onload', 'onclick', 'onmouseover'];

    // Remove dangerous child elements
    dangerousElements.forEach(tagName => {
      const elements = svg.querySelectorAll(tagName);
      elements.forEach(el => el.remove());
    });

    // Remove dangerous attributes from all elements
    const walker = document.createTreeWalker(
      svg,
      NodeFilter.SHOW_ELEMENT,
      null,
      false
    );

    let node;
    while (node = walker.nextNode()) {
      if (node instanceof Element) {
        dangerousAttributes.forEach(attr => {
          if (node.hasAttribute(attr)) {
            node.removeAttribute(attr);
          }
        });

        // Sanitize href attributes
        const href = node.getAttribute('href') || node.getAttribute('xlink:href');
        if (href && this.isDangerousURL(href)) {
          node.removeAttribute('href');
          node.removeAttribute('xlink:href');
        }
      }
    }
  }

  private applySVGSecurity(svg: SVGElement): void {
    // Set security attributes
    svg.setAttribute('xmlns', 'http://www.w3.org/2000/svg');
    
    // Prevent external resource loading
    svg.style.setProperty('--svg-external-resources', 'none');
  }

  private enableSVGAnimations(svg: SVGElement): void {
    // Enable SMIL animations if they exist
    const animateElements = svg.querySelectorAll('animate, animateTransform, animateMotion');
    
    for (const animate of animateElements) {
      // Validate animation attributes
      this.validateSVGAnimation(animate as SVGAnimationElement);
    }
  }

  private validateSVGAnimation(animate: SVGAnimationElement): void {
    // Ensure animation duration is reasonable
    const dur = animate.getAttribute('dur');
    if (dur) {
      const duration = parseFloat(dur);
      if (duration > 60) { // Max 60 seconds
        animate.setAttribute('dur', '60s');
      }
    }

    // Limit repeat count
    const repeatCount = animate.getAttribute('repeatCount');
    if (repeatCount === 'indefinite' || (repeatCount && parseFloat(repeatCount) > 100)) {
      animate.setAttribute('repeatCount', '100');
    }
  }

  private isDangerousURL(url: string): boolean {
    const dangerous = ['javascript:', 'data:', 'vbscript:', 'file:'];
    return dangerous.some(protocol => url.toLowerCase().startsWith(protocol));
  }
}

// Enhanced Responsive Manager with Mobile Optimizations
class ResponsiveManager {
  private container: HTMLElement;
  private resizeObserver?: ResizeObserver;
  private mediaQueries: Map<string, MediaQueryList> = new Map();
  private orientationHandler?: EventListener;
  private mobileManager?: any; // Will be imported dynamically

  constructor(container: HTMLElement) {
    this.container = container;
  }

  async initialize(): Promise<void> {
    // Set up resize observer
    if (typeof ResizeObserver !== 'undefined') {
      this.resizeObserver = new ResizeObserver(entries => {
        for (const entry of entries) {
          this.handleResize(entry.contentRect);
        }
      });
      
      this.resizeObserver.observe(this.container);
    }

    // Set up common media queries with enhanced mobile breakpoints
    this.setupMediaQueries();
    
    // Set up orientation change handling
    this.setupOrientationHandling();
    
    // Initialize mobile manager if on mobile device
    await this.initializeMobileManager();
  }

  private setupMediaQueries(): void {
    const queries = {
      // Enhanced mobile breakpoints
      'mobile-xs': '(max-width: 320px)',
      'mobile-sm': '(min-width: 321px) and (max-width: 480px)',
      'mobile': '(max-width: 768px)',
      'tablet': '(min-width: 769px) and (max-width: 1024px)',
      'desktop': '(min-width: 1025px)',
      'desktop-lg': '(min-width: 1440px)',
      
      // Orientation queries
      'portrait': '(orientation: portrait)',
      'landscape': '(orientation: landscape)',
      
      // Device-specific queries
      'touch': '(pointer: coarse)',
      'no-touch': '(pointer: fine)',
      'high-dpi': '(-webkit-min-device-pixel-ratio: 2), (min-resolution: 192dpi)',
      
      // Accessibility queries
      'reduced-motion': '(prefers-reduced-motion: reduce)',
      'dark-mode': '(prefers-color-scheme: dark)',
      'light-mode': '(prefers-color-scheme: light)',
      
      // Performance-based queries
      'slow-connection': '(prefers-reduced-data: reduce)'
    };

    for (const [name, query] of Object.entries(queries)) {
      const mq = window.matchMedia(query);
      this.mediaQueries.set(name, mq);
      
      // Use modern addEventListener instead of deprecated addListener
      const handler = () => this.handleMediaQueryChange(name, mq.matches);
      mq.addEventListener('change', handler);
    }
  }

  private setupOrientationHandling(): void {
    this.orientationHandler = () => {
      // Handle orientation change with debouncing
      setTimeout(() => {
        this.handleOrientationChange();
      }, 100);
    };
    
    window.addEventListener('orientationchange', this.orientationHandler);
    window.addEventListener('resize', this.orientationHandler);
  }

  private async initializeMobileManager(): Promise<void> {
    try {
      // Check if we're on a mobile device
      if (this.isMobileDevice()) {
        // Dynamically import mobile manager to avoid loading on desktop
        const { MobileManager } = await import('./mobile-manager');
        
        this.mobileManager = new MobileManager(this.container, {
          enableGestures: true,
          enableTouchOptimization: true,
          enablePerformanceMode: true,
          gestureThreshold: 10,
          touchDelay: 300,
          maxTouchPoints: 10,
          enableHapticFeedback: false
        });
        
        this.mobileManager.initialize();
        
        // Add mobile-specific CSS class
        this.container.classList.add('liv-mobile-device');
      }
    } catch (error) {
      console.warn('Failed to initialize mobile manager:', error);
    }
  }

  private isMobileDevice(): boolean {
    // Enhanced mobile detection
    const hasTouch = 'ontouchstart' in window || navigator.maxTouchPoints > 0;
    const userAgent = navigator.userAgent.toLowerCase();
    const mobileKeywords = [
      'mobile', 'android', 'iphone', 'ipad', 'ipod', 'blackberry', 
      'windows phone', 'webos', 'opera mini'
    ];
    const hasMobileUA = mobileKeywords.some(keyword => userAgent.includes(keyword));
    const hasSmallScreen = window.screen.width <= 768 || window.screen.height <= 768;
    
    // Check for coarse pointer (touch)
    const hasCoarsePointer = window.matchMedia('(pointer: coarse)').matches;
    
    return hasTouch || hasMobileUA || hasSmallScreen || hasCoarsePointer;
  }

  private handleResize(rect: DOMRectReadOnly): void {
    // Update CSS custom properties for responsive design
    this.container.style.setProperty('--container-width', `${rect.width}px`);
    this.container.style.setProperty('--container-height', `${rect.height}px`);
    
    // Calculate responsive breakpoints
    const aspectRatio = rect.width / rect.height;
    this.container.style.setProperty('--aspect-ratio', aspectRatio.toString());
    
    // Set responsive font size based on container width
    const baseFontSize = Math.max(14, Math.min(18, rect.width / 50));
    this.container.style.setProperty('--responsive-font-size', `${baseFontSize}px`);
    
    // Update responsive classes based on size
    this.updateSizeClasses(rect);
    
    // Dispatch custom resize event with enhanced data
    this.container.dispatchEvent(new CustomEvent('liv-resize', {
      detail: { 
        width: rect.width, 
        height: rect.height,
        aspectRatio,
        baseFontSize,
        isMobile: rect.width <= 768,
        isTablet: rect.width > 768 && rect.width <= 1024,
        isDesktop: rect.width > 1024
      }
    }));
  }

  private updateSizeClasses(rect: DOMRectReadOnly): void {
    // Remove existing size classes
    const sizeClasses = ['liv-xs', 'liv-sm', 'liv-md', 'liv-lg', 'liv-xl'];
    sizeClasses.forEach(cls => this.container.classList.remove(cls));
    
    // Add appropriate size class
    if (rect.width <= 320) {
      this.container.classList.add('liv-xs');
    } else if (rect.width <= 480) {
      this.container.classList.add('liv-sm');
    } else if (rect.width <= 768) {
      this.container.classList.add('liv-md');
    } else if (rect.width <= 1024) {
      this.container.classList.add('liv-lg');
    } else {
      this.container.classList.add('liv-xl');
    }
  }

  private handleMediaQueryChange(name: string, matches: boolean): void {
    // Update CSS classes based on media query matches
    this.container.classList.toggle(`liv-${name}`, matches);
    
    // Handle specific media query changes
    if (name === 'reduced-motion' && matches) {
      // Disable animations for users who prefer reduced motion
      this.container.style.setProperty('--animation-duration', '0s');
      this.container.style.setProperty('--transition-duration', '0s');
    } else if (name === 'reduced-motion' && !matches) {
      // Re-enable animations
      this.container.style.removeProperty('--animation-duration');
      this.container.style.removeProperty('--transition-duration');
    }
    
    if (name === 'slow-connection' && matches) {
      // Optimize for slow connections
      this.container.classList.add('liv-optimize-bandwidth');
    } else if (name === 'slow-connection' && !matches) {
      this.container.classList.remove('liv-optimize-bandwidth');
    }
    
    // Dispatch custom media query event
    this.container.dispatchEvent(new CustomEvent('liv-media-change', {
      detail: { query: name, matches }
    }));
  }

  private handleOrientationChange(): void {
    // Get current orientation
    const orientation = window.screen?.orientation?.type || 
                      (window.innerHeight > window.innerWidth ? 'portrait-primary' : 'landscape-primary');
    
    // Update orientation classes
    const orientationClasses = ['liv-portrait', 'liv-landscape'];
    orientationClasses.forEach(cls => this.container.classList.remove(cls));
    
    if (orientation.includes('portrait')) {
      this.container.classList.add('liv-portrait');
    } else {
      this.container.classList.add('liv-landscape');
    }
    
    // Update CSS custom property
    this.container.style.setProperty('--orientation', orientation);
    
    // Dispatch orientation change event
    this.container.dispatchEvent(new CustomEvent('liv-orientation-change', {
      detail: { orientation }
    }));
    
    // Notify mobile manager if available
    if (this.mobileManager && this.mobileManager.handleOrientationChange) {
      this.mobileManager.handleOrientationChange(orientation);
    }
  }

  // Initialize responsive classes immediately
  initializeClasses(): void {
    for (const [name, mq] of this.mediaQueries.entries()) {
      this.container.classList.toggle(`liv-${name}`, mq.matches);
    }
    
    // Initialize size classes
    const rect = this.container.getBoundingClientRect();
    this.updateSizeClasses(rect);
    
    // Initialize orientation
    this.handleOrientationChange();
  }

  // Public methods for external control
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
    if (this.resizeObserver) {
      this.resizeObserver.disconnect();
    }

    // Remove media query listeners
    for (const mq of this.mediaQueries.values()) {
      // Note: We can't easily remove the specific handler without storing references
      // In a production implementation, you'd want to store handler references
    }

    // Remove orientation handler
    if (this.orientationHandler) {
      window.removeEventListener('orientationchange', this.orientationHandler);
      window.removeEventListener('resize', this.orientationHandler);
    }

    // Destroy mobile manager
    if (this.mobileManager && this.mobileManager.destroy) {
      this.mobileManager.destroy();
    }

    this.mediaQueries.clear();
  }
}

interface ContentOptions {
  interactive?: boolean;
  sanitize?: boolean;
  allowScripts?: boolean;
  enableSVG?: boolean;
}

interface StyleOptions {
  enableAnimations?: boolean;
  sanitize?: boolean;
}

class SandboxedDOM {
  private container: HTMLElement;
  private permissions: LegacySecurityPolicy;
  private elements: Map<string, HTMLElement> = new Map();
  private shadowRoot: ShadowRoot;
  private errorHandler: ErrorHandler;
  private styleElement?: HTMLStyleElement;
  private contentFrame?: HTMLIFrameElement;

  constructor(container: HTMLElement, permissions: LegacySecurityPolicy, errorHandler: ErrorHandler) {
    this.container = container;
    this.permissions = permissions;
    this.errorHandler = errorHandler;
    
    // Create shadow DOM for isolation
    this.shadowRoot = container.attachShadow({ mode: 'closed' });
    this.initializeSandbox();
  }

  private initializeSandbox(): void {
    // Create style element for CSS
    this.styleElement = document.createElement('style');
    this.shadowRoot.appendChild(this.styleElement);
    
    // Set up basic sandbox styles
    this.styleElement.textContent = `
      :host {
        display: block;
        width: 100%;
        height: 100%;
        overflow: auto;
      }
      
      * {
        box-sizing: border-box;
      }
      
      /* Prevent potential security issues */
      script {
        display: none !important;
      }
      
      iframe {
        sandbox: allow-same-origin;
      }
    `;
  }

  async setContent(html: string, options: ContentOptions = {}): Promise<void> {
    const opts = {
      interactive: true,
      sanitize: true,
      allowScripts: false,
      enableSVG: true,
      ...options
    };

    try {
      // Clear existing content
      this.clear();
      
      // Sanitize HTML if required
      const sanitizedHtml = opts.sanitize ? this.sanitizeHTML(html) : html;
      
      if (opts.interactive && this.permissions.jsPermissions.executionMode !== 'none') {
        // Use iframe for interactive content
        await this.setInteractiveContent(sanitizedHtml, opts);
      } else {
        // Direct DOM insertion for static content
        await this.setStaticContent(sanitizedHtml);
      }
      
    } catch (error) {
      this.errorHandler.handleError(new LIVError(
        LIVErrorType.VALIDATION,
        'Failed to set content',
        error instanceof Error ? error : new Error(String(error))
      ));
      throw error;
    }
  }

  private async setInteractiveContent(html: string, options: ContentOptions): Promise<void> {
    // Create sandboxed iframe for interactive content
    this.contentFrame = document.createElement('iframe');
    this.contentFrame.style.width = '100%';
    this.contentFrame.style.height = '100%';
    this.contentFrame.style.border = 'none';
    
    // Set sandbox attributes based on permissions
    const sandboxFlags = this.getSandboxFlags();
    this.contentFrame.setAttribute('sandbox', sandboxFlags.join(' '));
    
    // Set CSP if available
    if (this.permissions.contentSecurityPolicy) {
      this.contentFrame.setAttribute('csp', this.permissions.contentSecurityPolicy);
    }
    
    this.shadowRoot.appendChild(this.contentFrame);
    
    // Wait for iframe to load
    await new Promise<void>((resolve, reject) => {
      const timeout = setTimeout(() => {
        reject(new Error('Content loading timeout'));
      }, 5000);
      
      this.contentFrame!.onload = () => {
        clearTimeout(timeout);
        resolve();
      };
      
      this.contentFrame!.onerror = () => {
        clearTimeout(timeout);
        reject(new Error('Content loading failed'));
      };
      
      // Set content
      const doc = this.contentFrame!.contentDocument;
      if (doc) {
        doc.open();
        doc.write(html);
        doc.close();
      }
    });
  }

  private async setStaticContent(html: string): Promise<void> {
    // Create container for static content
    const contentDiv = document.createElement('div');
    contentDiv.innerHTML = html;
    
    // Apply security restrictions
    this.applySecurityToElement(contentDiv);
    
    this.shadowRoot.appendChild(contentDiv);
  }

  async applyStyles(css: string, options: StyleOptions = {}): Promise<void> {
    try {
      if (!this.styleElement) {
        this.styleElement = document.createElement('style');
        this.shadowRoot.appendChild(this.styleElement);
      }
      
      const opts = {
        enableAnimations: true,
        sanitize: true,
        ...options
      };
      
      // Sanitize CSS
      let processedCSS = opts.sanitize ? this.sanitizeCSS(css) : css;
      
      // Process animations if enabled
      if (opts.enableAnimations) {
        processedCSS = this.enhanceAnimationCSS(processedCSS);
      }
      
      // Apply styles
      this.styleElement.textContent += '\n' + processedCSS;
      
    } catch (error) {
      this.errorHandler.handleError(new LIVError(
        LIVErrorType.VALIDATION,
        'Failed to apply styles',
        error instanceof Error ? error : new Error(String(error))
      ));
    }
  }

  private enhanceAnimationCSS(css: string): string {
    // Add performance optimizations for animations
    let enhanced = css;
    
    // Add will-change property for animated elements
    enhanced = enhanced.replace(
      /(animation(?:-name)?:\s*[^;]+;)/gi,
      '$1\n  will-change: transform, opacity;'
    );
    
    // Optimize transform animations
    enhanced = enhanced.replace(
      /transform:\s*translate\(/gi,
      'transform: translate3d('
    );
    
    return enhanced;
  }

  applyAnimationUpdate(update: AnimationUpdate): void {
    const element = this.elements.get(update.elementId);
    if (!element) return;

    // Apply animation styles
    for (const [property, value] of Object.entries(update.styles)) {
      if (this.isAllowedCSSProperty(property)) {
        (element.style as any)[property] = value;
      }
    }

    // Add animation class for CSS transitions
    element.classList.add('liv-animating');
    
    // Remove class after animation completes
    if (update.progress >= 1) {
      setTimeout(() => {
        element.classList.remove('liv-animating');
      }, 50);
    }
  }

  querySelectorAll(selector: string): NodeListOf<Element> {
    return this.shadowRoot.querySelectorAll(selector);
  }

  private getSandboxFlags(): string[] {
    const flags: string[] = [];
    
    // Always include basic safety
    flags.push('allow-same-origin');
    
    // Add flags based on permissions
    if (this.permissions.jsPermissions.executionMode === 'sandboxed') {
      flags.push('allow-scripts');
    }
    
    if (this.permissions.jsPermissions.domAccess === 'write') {
      flags.push('allow-forms');
    }
    
    if (this.permissions.networkPolicy?.allowOutbound) {
      flags.push('allow-same-origin');
    }
    
    return flags;
  }

  private sanitizeHTML(html: string): string {
    // Create a temporary DOM to parse and sanitize
    const tempDiv = document.createElement('div');
    tempDiv.innerHTML = html;
    
    // Remove dangerous elements and attributes
    this.sanitizeElement(tempDiv);
    
    return tempDiv.innerHTML;
  }

  private sanitizeElement(element: Element): void {
    // Remove dangerous elements (but preserve SVG elements)
    const dangerousElements = ['script', 'object', 'embed', 'applet', 'meta', 'link'];
    dangerousElements.forEach(tagName => {
      const elements = element.querySelectorAll(tagName);
      elements.forEach(el => {
        // Don't remove SVG script elements if they're part of SVG
        if (tagName === 'script' && el.closest('svg')) {
          return;
        }
        el.remove();
      });
    });
    
    // Sanitize attributes on all elements
    const walker = document.createTreeWalker(
      element,
      NodeFilter.SHOW_ELEMENT,
      null,
      false
    );
    
    let node;
    while (node = walker.nextNode()) {
      if (node instanceof Element) {
        this.sanitizeAttributes(node);
      }
    }
  }

  private sanitizeAttributes(element: Element): void {
    const dangerousAttributes = [
      'onclick', 'onload', 'onerror', 'onmouseover', 'onmouseout',
      'onfocus', 'onblur', 'onchange', 'onsubmit', 'onreset',
      'onselect', 'onkeydown', 'onkeypress', 'onkeyup'
    ];
    
    // Remove dangerous attributes
    dangerousAttributes.forEach(attr => {
      if (element.hasAttribute(attr)) {
        element.removeAttribute(attr);
      }
    });
    
    // Sanitize href and src attributes
    ['href', 'src'].forEach(attr => {
      const value = element.getAttribute(attr);
      if (value && this.isDangerousURL(value)) {
        element.setAttribute(attr, '#');
      }
    });
  }

  private isDangerousURL(url: string): boolean {
    const dangerous = ['javascript:', 'data:', 'vbscript:', 'file:', 'about:'];
    return dangerous.some(protocol => url.toLowerCase().startsWith(protocol));
  }

  private sanitizeCSS(css: string): string {
    // Remove dangerous CSS patterns
    const dangerousPatterns = [
      /javascript\s*:/gi,
      /expression\s*\(/gi,
      /behavior\s*:/gi,
      /binding\s*:/gi,
      /-moz-binding/gi,
      /url\s*\(\s*["']?javascript:/gi,
      /url\s*\(\s*["']?data:/gi,
      /@import/gi
    ];
    
    let sanitized = css;
    dangerousPatterns.forEach(pattern => {
      sanitized = sanitized.replace(pattern, '/* removed for security */');
    });
    
    return sanitized;
  }

  private applySecurityToElement(element: Element): void {
    // Apply security restrictions recursively
    this.sanitizeElement(element);
    
    // Add security event listeners if needed
    if (this.permissions.jsPermissions.executionMode === 'none') {
      // Prevent any script execution
      element.addEventListener('click', (e) => {
        const target = e.target as Element;
        if (target.tagName.toLowerCase() === 'a') {
          const href = target.getAttribute('href');
          if (href && this.isDangerousURL(href)) {
            e.preventDefault();
          }
        }
      });
    }
  }

  clear(): void {
    // Remove all content except style element
    Array.from(this.shadowRoot.children).forEach(child => {
      if (child !== this.styleElement) {
        child.remove();
      }
    });
    
    this.elements.clear();
    this.contentFrame = undefined;
  }

  createElement(id: string, tag: string, parentId?: string): HTMLElement {
    // Validate tag against allowed elements
    if (!this.isAllowedElement(tag)) {
      throw new Error(`Element type '${tag}' not allowed by security policy`);
    }

    const element = document.createElement(tag);
    element.id = id;
    
    // Apply security restrictions
    this.applySecurity(element);
    
    // Add to parent or shadow root
    const parent = parentId ? this.elements.get(parentId) : this.shadowRoot;
    if (parent) {
      parent.appendChild(element);
    }
    
    this.elements.set(id, element);
    return element;
  }

  updateElement(id: string, attributes: Record<string, string>): void {
    const element = this.elements.get(id);
    if (!element) return;

    for (const [key, value] of Object.entries(attributes)) {
      if (this.isAllowedAttribute(key)) {
        element.setAttribute(key, value);
      }
    }
  }

  removeElement(id: string): void {
    const element = this.elements.get(id);
    if (element && element.parentNode) {
      element.parentNode.removeChild(element);
      this.elements.delete(id);
    }
  }

  moveElement(id: string, newParentId: string, index: number): void {
    const element = this.elements.get(id);
    const newParent = this.elements.get(newParentId);
    
    if (element && newParent) {
      const children = Array.from(newParent.children);
      if (index >= children.length) {
        newParent.appendChild(element);
      } else {
        newParent.insertBefore(element, children[index]);
      }
    }
  }

  setStyle(id: string, property: string, value: string): void {
    const element = this.elements.get(id);
    if (!element) return;

    // Validate CSS property
    if (this.isAllowedCSSProperty(property)) {
      (element.style as any)[property] = value;
    }
  }

  getElement(id: string): HTMLElement | undefined {
    return this.elements.get(id);
  }

  private applySecurity(element: HTMLElement): void {
    // Remove potentially dangerous attributes
    const dangerousAttributes = ['onclick', 'onload', 'onerror', 'onmouseover'];
    dangerousAttributes.forEach(attr => {
      element.removeAttribute(attr);
    });

    // Prevent script execution
    if (element.tagName.toLowerCase() === 'script') {
      element.setAttribute('type', 'text/plain');
    }
  }

  private isAllowedElement(tag: string): boolean {
    const allowedElements = [
      'div', 'span', 'p', 'h1', 'h2', 'h3', 'h4', 'h5', 'h6',
      'img', 'svg', 'canvas', 'video', 'audio',
      'ul', 'ol', 'li', 'table', 'tr', 'td', 'th',
      'form', 'input', 'button', 'select', 'option'
    ];
    
    return allowedElements.includes(tag.toLowerCase());
  }

  private isAllowedAttribute(attr: string): boolean {
    const dangerousAttributes = [
      'onclick', 'onload', 'onerror', 'onmouseover', 'onmouseout',
      'onfocus', 'onblur', 'onchange', 'onsubmit'
    ];
    
    return !dangerousAttributes.includes(attr.toLowerCase());
  }

  private isAllowedCSSProperty(property: string): boolean {
    // Block dangerous CSS properties
    const dangerousProperties = [
      'behavior', 'binding', 'expression', 'javascript',
      'vbscript', 'mozbinding'
    ];
    
    return !dangerousProperties.some(dangerous => 
      property.toLowerCase().includes(dangerous)
    );
  }

  destroy(): void {
    this.elements.clear();
    this.shadowRoot.innerHTML = '';
  }
}