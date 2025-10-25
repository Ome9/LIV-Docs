// Mobile-specific optimizations and gesture handling for LIV documents

import { 
  InteractionEvent, 
  InteractionType, 
  Position, 
  TouchData, 
  GestureData, 
  GestureType,
  TouchPoint 
} from './types';
import { LIVError, LIVErrorType, ErrorHandler } from './errors';

export interface MobileOptimizationOptions {
  enableGestures?: boolean;
  enableTouchOptimization?: boolean;
  enablePerformanceMode?: boolean;
  gestureThreshold?: number;
  touchDelay?: number;
  maxTouchPoints?: number;
  enableHapticFeedback?: boolean;
}

export class MobileManager {
  private container: HTMLElement;
  private options: Required<MobileOptimizationOptions>;
  private gestureRecognizer: GestureRecognizer;
  private touchOptimizer: TouchOptimizer;
  private performanceManager: MobilePerformanceManager;
  private errorHandler: ErrorHandler;
  private isActive: boolean = false;
  private eventHandlers: Map<string, EventListener> = new Map();

  constructor(container: HTMLElement, options: MobileOptimizationOptions = {}) {
    this.container = container;
    this.options = {
      enableGestures: true,
      enableTouchOptimization: true,
      enablePerformanceMode: true,
      gestureThreshold: 10,
      touchDelay: 300,
      maxTouchPoints: 10,
      enableHapticFeedback: false,
      ...options
    };
    
    this.errorHandler = ErrorHandler.getInstance();
    this.gestureRecognizer = new GestureRecognizer(this.options);
    this.touchOptimizer = new TouchOptimizer(this.options);
    this.performanceManager = new MobilePerformanceManager(this.options);
  }

  initialize(): void {
    if (this.isActive) return;

    try {
      // Detect mobile environment
      if (!this.isMobileDevice()) {
        console.log('Mobile manager: Not a mobile device, skipping mobile optimizations');
        return;
      }

      // Apply mobile-specific CSS optimizations
      this.applyMobileCSS();

      // Set up touch event handling
      if (this.options.enableTouchOptimization) {
        this.setupTouchHandling();
      }

      // Initialize gesture recognition
      if (this.options.enableGestures) {
        this.gestureRecognizer.initialize(this.container);
      }

      // Enable performance optimizations
      if (this.options.enablePerformanceMode) {
        this.performanceManager.initialize(this.container);
      }

      // Set up viewport meta tag for mobile
      this.setupMobileViewport();

      // Prevent default touch behaviors that interfere with gestures
      this.preventDefaultTouchBehaviors();

      this.isActive = true;
      console.log('Mobile manager initialized successfully');

    } catch (error) {
      this.errorHandler.handleError(new LIVError(
        LIVErrorType.VALIDATION,
        'Failed to initialize mobile manager',
        error instanceof Error ? error : new Error(String(error))
      ));
    }
  }

  private isMobileDevice(): boolean {
    // Check for touch capability
    const hasTouch = 'ontouchstart' in window || navigator.maxTouchPoints > 0;
    
    // Check user agent for mobile indicators
    const userAgent = navigator.userAgent.toLowerCase();
    const mobileKeywords = ['mobile', 'android', 'iphone', 'ipad', 'ipod', 'blackberry', 'windows phone'];
    const hasMobileUA = mobileKeywords.some(keyword => userAgent.includes(keyword));
    
    // Check screen size (mobile-like dimensions)
    const hasSmallScreen = window.screen.width <= 768 || window.screen.height <= 768;
    
    return hasTouch || hasMobileUA || hasSmallScreen;
  }

  private applyMobileCSS(): void {
    const style = document.createElement('style');
    style.textContent = `
      /* Mobile-specific optimizations */
      .liv-mobile-optimized {
        /* Improve touch target sizes */
        --min-touch-target: 44px;
        
        /* Optimize scrolling */
        -webkit-overflow-scrolling: touch;
        overflow-scrolling: touch;
        
        /* Prevent text selection during gestures */
        -webkit-user-select: none;
        -moz-user-select: none;
        -ms-user-select: none;
        user-select: none;
        
        /* Optimize rendering */
        -webkit-transform: translateZ(0);
        transform: translateZ(0);
        
        /* Prevent zoom on input focus */
        -webkit-text-size-adjust: 100%;
        text-size-adjust: 100%;
      }
      
      /* Touch-friendly interactive elements */
      .liv-mobile-optimized [data-interactive] {
        min-width: var(--min-touch-target);
        min-height: var(--min-touch-target);
        padding: 8px;
        margin: 4px;
      }
      
      /* Optimize animations for mobile */
      .liv-mobile-optimized .liv-animating {
        -webkit-transform: translateZ(0);
        transform: translateZ(0);
        will-change: transform, opacity;
      }
      
      /* Gesture feedback */
      .liv-gesture-active {
        -webkit-touch-callout: none;
        -webkit-user-select: none;
        user-select: none;
      }
      
      /* Mobile-specific responsive classes */
      @media (max-width: 480px) {
        .liv-mobile-optimized {
          font-size: 16px; /* Prevent zoom on iOS */
        }
        
        .liv-mobile-optimized .liv-chart {
          /* Optimize charts for small screens */
          max-width: 100%;
          height: auto;
        }
      }
      
      /* Orientation-specific optimizations */
      @media (orientation: portrait) {
        .liv-mobile-optimized .liv-landscape-only {
          display: none;
        }
      }
      
      @media (orientation: landscape) {
        .liv-mobile-optimized .liv-portrait-only {
          display: none;
        }
      }
      
      /* High DPI optimizations */
      @media (-webkit-min-device-pixel-ratio: 2), (min-resolution: 192dpi) {
        .liv-mobile-optimized .liv-vector-graphics {
          image-rendering: -webkit-optimize-contrast;
          image-rendering: crisp-edges;
        }
      }
    `;
    
    document.head.appendChild(style);
    this.container.classList.add('liv-mobile-optimized');
  }

  private setupMobileViewport(): void {
    // Check if viewport meta tag exists
    let viewport = document.querySelector('meta[name="viewport"]') as HTMLMetaElement;
    
    if (!viewport) {
      viewport = document.createElement('meta');
      viewport.name = 'viewport';
      document.head.appendChild(viewport);
    }
    
    // Set mobile-optimized viewport
    viewport.content = 'width=device-width, initial-scale=1.0, maximum-scale=5.0, user-scalable=yes, viewport-fit=cover';
  }

  private setupTouchHandling(): void {
    // Set up optimized touch event listeners
    const touchEvents = ['touchstart', 'touchmove', 'touchend', 'touchcancel'];
    
    touchEvents.forEach(eventType => {
      const handler = this.createTouchHandler(eventType);
      this.container.addEventListener(eventType, handler, { 
        passive: false, // Allow preventDefault for gesture handling
        capture: true 
      });
      this.eventHandlers.set(eventType, handler);
    });
  }

  private createTouchHandler(eventType: string): EventListener {
    return (event: Event) => {
      const touchEvent = event as TouchEvent;
      
      try {
        // Optimize touch handling
        const optimizedEvent = this.touchOptimizer.optimizeTouch(touchEvent);
        
        // Convert to LIV interaction event
        const interactionEvent = this.convertTouchToInteraction(optimizedEvent, eventType);
        
        // Process gesture recognition
        if (this.options.enableGestures) {
          const gesture = this.gestureRecognizer.processTouch(optimizedEvent);
          if (gesture) {
            this.handleGesture(gesture);
          }
        }
        
        // Dispatch interaction event
        this.dispatchInteractionEvent(interactionEvent);
        
      } catch (error) {
        this.errorHandler.handleError(new LIVError(
          LIVErrorType.VALIDATION,
          `Touch handling error for ${eventType}`,
          error instanceof Error ? error : new Error(String(error))
        ));
      }
    };
  }

  private convertTouchToInteraction(touchEvent: TouchEvent, eventType: string): InteractionEvent {
    const rect = this.container.getBoundingClientRect();
    
    // Convert touches to TouchPoint array
    const convertTouches = (touches: TouchList): TouchPoint[] => {
      return Array.from(touches).slice(0, this.options.maxTouchPoints).map(touch => ({
        identifier: touch.identifier,
        position: {
          x: touch.clientX - rect.left,
          y: touch.clientY - rect.top
        },
        radius: touch.radiusX ? Math.max(touch.radiusX, touch.radiusY) : undefined,
        rotation_angle: touch.rotationAngle || undefined,
        force: touch.force || undefined
      }));
    };

    const touchData: TouchData = {
      touches: convertTouches(touchEvent.touches),
      changed_touches: convertTouches(touchEvent.changedTouches),
      target_touches: convertTouches(touchEvent.targetTouches),
      force: touchEvent.touches.length > 0 ? touchEvent.touches[0].force : undefined,
      rotation_angle: touchEvent.touches.length > 0 ? touchEvent.touches[0].rotationAngle : undefined,
      scale: undefined // Will be calculated by gesture recognizer
    };

    // Determine interaction type
    let interactionType: InteractionType;
    switch (eventType) {
      case 'touchstart': interactionType = InteractionType.TouchStart; break;
      case 'touchmove': interactionType = InteractionType.TouchMove; break;
      case 'touchend': interactionType = InteractionType.TouchEnd; break;
      case 'touchcancel': interactionType = InteractionType.TouchCancel; break;
      default: interactionType = InteractionType.TouchStart;
    }

    // Get primary touch position
    const primaryTouch = touchEvent.touches.length > 0 ? touchEvent.touches[0] : touchEvent.changedTouches[0];
    const position: Position | undefined = primaryTouch ? {
      x: primaryTouch.clientX - rect.left,
      y: primaryTouch.clientY - rect.top
    } : undefined;

    return {
      eventType: interactionType,
      targetElement: (touchEvent.target as Element)?.id,
      position,
      data: {
        touchCount: touchEvent.touches.length,
        changedTouchCount: touchEvent.changedTouches.length
      },
      timestamp: performance.now(),
      touch_data: touchData,
      mouse_data: undefined,
      keyboard_data: undefined,
      gesture_data: undefined,
      modifiers: {
        ctrl: false,
        shift: false,
        alt: false,
        meta: false
      }
    };
  }

  private handleGesture(gesture: RecognizedGesture): void {
    try {
      // Provide haptic feedback if enabled
      if (this.options.enableHapticFeedback && 'vibrate' in navigator) {
        this.provideHapticFeedback(gesture.type);
      }

      // Create gesture interaction event
      const gestureEvent: InteractionEvent = {
        eventType: this.mapGestureToInteractionType(gesture.type),
        targetElement: gesture.targetElement,
        position: gesture.position,
        data: {
          gestureType: gesture.type,
          confidence: gesture.confidence,
          duration: gesture.duration
        },
        timestamp: performance.now(),
        touch_data: undefined,
        mouse_data: undefined,
        keyboard_data: undefined,
        gesture_data: {
          gesture_type: gesture.type,
          start_position: gesture.startPosition,
          current_position: gesture.position,
          delta: {
            x: gesture.position.x - gesture.startPosition.x,
            y: gesture.position.y - gesture.startPosition.y
          },
          velocity: gesture.velocity,
          scale: gesture.scale,
          rotation: gesture.rotation,
          distance: gesture.distance,
          duration: gesture.duration
        },
        modifiers: {
          ctrl: false,
          shift: false,
          alt: false,
          meta: false
        }
      };

      this.dispatchInteractionEvent(gestureEvent);

    } catch (error) {
      this.errorHandler.handleError(new LIVError(
        LIVErrorType.VALIDATION,
        'Gesture handling error',
        error instanceof Error ? error : new Error(String(error))
      ));
    }
  }

  private mapGestureToInteractionType(gestureType: GestureType): InteractionType {
    switch (gestureType) {
      case GestureType.Tap: return InteractionType.Tap;
      case GestureType.DoubleTap: return InteractionType.DoubleTap;
      case GestureType.LongPress: return InteractionType.LongPress;
      case GestureType.Pinch: return InteractionType.Pinch;
      case GestureType.Rotate: return InteractionType.Rotate;
      case GestureType.Swipe: return InteractionType.Swipe;
      case GestureType.Pan: return InteractionType.Pan;
      default: return InteractionType.Tap;
    }
  }

  private provideHapticFeedback(gestureType: GestureType): void {
    if (!('vibrate' in navigator)) return;

    // Different vibration patterns for different gestures
    let pattern: number | number[];
    
    switch (gestureType) {
      case GestureType.Tap:
        pattern = 50; // Short vibration
        break;
      case GestureType.DoubleTap:
        pattern = [50, 50, 50]; // Double pulse
        break;
      case GestureType.LongPress:
        pattern = 200; // Longer vibration
        break;
      case GestureType.Pinch:
      case GestureType.Rotate:
        pattern = [25, 25, 25]; // Subtle feedback
        break;
      default:
        pattern = 30; // Default short vibration
    }

    navigator.vibrate(pattern);
  }

  private preventDefaultTouchBehaviors(): void {
    // Prevent default behaviors that interfere with gestures
    this.container.addEventListener('touchstart', (e) => {
      // Allow single touch for scrolling, prevent multi-touch defaults
      if (e.touches.length > 1) {
        e.preventDefault();
      }
    }, { passive: false });

    this.container.addEventListener('touchmove', (e) => {
      // Prevent default for multi-touch to enable gestures
      if (e.touches.length > 1) {
        e.preventDefault();
      }
    }, { passive: false });

    // Prevent context menu on long press
    this.container.addEventListener('contextmenu', (e) => {
      e.preventDefault();
    });

    // Prevent double-tap zoom
    this.container.addEventListener('dblclick', (e) => {
      e.preventDefault();
    });
  }

  private dispatchInteractionEvent(event: InteractionEvent): void {
    // Dispatch custom event that can be caught by the renderer
    const customEvent = new CustomEvent('liv-interaction', {
      detail: event,
      bubbles: true,
      cancelable: true
    });
    
    this.container.dispatchEvent(customEvent);
  }

  // Public methods for external control
  enableGestures(): void {
    this.options.enableGestures = true;
    if (this.isActive) {
      this.gestureRecognizer.initialize(this.container);
    }
  }

  disableGestures(): void {
    this.options.enableGestures = false;
    this.gestureRecognizer.destroy();
  }

  setGestureThreshold(threshold: number): void {
    this.options.gestureThreshold = threshold;
    this.gestureRecognizer.updateOptions(this.options);
  }

  getPerformanceMetrics(): MobilePerformanceMetrics {
    return this.performanceManager.getMetrics();
  }

  destroy(): void {
    if (!this.isActive) return;

    // Remove event listeners
    for (const [eventType, handler] of this.eventHandlers.entries()) {
      this.container.removeEventListener(eventType, handler);
    }
    this.eventHandlers.clear();

    // Destroy components
    this.gestureRecognizer.destroy();
    this.touchOptimizer.destroy();
    this.performanceManager.destroy();

    // Remove CSS class
    this.container.classList.remove('liv-mobile-optimized');

    this.isActive = false;
  }
}

// Gesture Recognition System
interface RecognizedGesture {
  type: GestureType;
  position: Position;
  startPosition: Position;
  targetElement?: string;
  confidence: number;
  duration: number;
  velocity?: Position;
  scale?: number;
  rotation?: number;
  distance?: number;
}

class GestureRecognizer {
  private options: Required<MobileOptimizationOptions>;
  private activeGestures: Map<string, GestureState> = new Map();
  private gestureHistory: GestureEvent[] = [];
  private isActive: boolean = false;

  constructor(options: Required<MobileOptimizationOptions>) {
    this.options = options;
  }

  initialize(container: HTMLElement): void {
    this.isActive = true;
    // Gesture recognition is handled through touch events processed by MobileManager
  }

  processTouch(touchEvent: TouchEvent): RecognizedGesture | null {
    if (!this.isActive) return null;

    try {
      const gestureId = this.getGestureId(touchEvent);
      const currentTime = performance.now();

      switch (touchEvent.type) {
        case 'touchstart':
          return this.handleTouchStart(touchEvent, gestureId, currentTime);
        case 'touchmove':
          return this.handleTouchMove(touchEvent, gestureId, currentTime);
        case 'touchend':
          return this.handleTouchEnd(touchEvent, gestureId, currentTime);
        case 'touchcancel':
          return this.handleTouchCancel(gestureId);
        default:
          return null;
      }
    } catch (error) {
      console.error('Gesture recognition error:', error);
      return null;
    }
  }

  private getGestureId(touchEvent: TouchEvent): string {
    // Create unique ID based on touch identifiers
    const identifiers = Array.from(touchEvent.touches).map(t => t.identifier).sort();
    return identifiers.join('-');
  }

  private handleTouchStart(touchEvent: TouchEvent, gestureId: string, timestamp: number): RecognizedGesture | null {
    const touches = Array.from(touchEvent.touches);
    const primaryTouch = touches[0];

    const gestureState: GestureState = {
      id: gestureId,
      startTime: timestamp,
      startPosition: {
        x: primaryTouch.clientX,
        y: primaryTouch.clientY
      },
      currentPosition: {
        x: primaryTouch.clientX,
        y: primaryTouch.clientY
      },
      touchCount: touches.length,
      lastMoveTime: timestamp,
      totalDistance: 0,
      velocity: { x: 0, y: 0 },
      scale: 1,
      rotation: 0
    };

    this.activeGestures.set(gestureId, gestureState);

    // Check for immediate gestures (like tap start)
    if (touches.length === 1) {
      // Could be start of tap, long press, or swipe
      return null; // Wait for more events
    } else if (touches.length === 2) {
      // Could be start of pinch or rotate
      gestureState.initialDistance = this.calculateDistance(touches[0], touches[1]);
      gestureState.initialAngle = this.calculateAngle(touches[0], touches[1]);
      return null; // Wait for movement
    }

    return null;
  }

  private handleTouchMove(touchEvent: TouchEvent, gestureId: string, timestamp: number): RecognizedGesture | null {
    const gestureState = this.activeGestures.get(gestureId);
    if (!gestureState) return null;

    const touches = Array.from(touchEvent.touches);
    const primaryTouch = touches[0];
    
    // Update gesture state
    const previousPosition = gestureState.currentPosition;
    gestureState.currentPosition = {
      x: primaryTouch.clientX,
      y: primaryTouch.clientY
    };

    // Calculate movement
    const deltaX = gestureState.currentPosition.x - previousPosition.x;
    const deltaY = gestureState.currentPosition.y - previousPosition.y;
    const deltaTime = timestamp - gestureState.lastMoveTime;
    
    gestureState.totalDistance += Math.sqrt(deltaX * deltaX + deltaY * deltaY);
    gestureState.lastMoveTime = timestamp;

    // Calculate velocity
    if (deltaTime > 0) {
      gestureState.velocity = {
        x: deltaX / deltaTime,
        y: deltaY / deltaTime
      };
    }

    // Detect gesture type based on movement
    if (touches.length === 1) {
      // Single touch - could be pan or swipe
      if (gestureState.totalDistance > this.options.gestureThreshold) {
        const duration = timestamp - gestureState.startTime;
        const totalDelta = {
          x: gestureState.currentPosition.x - gestureState.startPosition.x,
          y: gestureState.currentPosition.y - gestureState.startPosition.y
        };

        // Determine if it's a swipe (fast movement) or pan (slower movement)
        const speed = gestureState.totalDistance / duration;
        const isSwipe = speed > 0.5; // pixels per millisecond

        if (isSwipe && Math.abs(totalDelta.x) > Math.abs(totalDelta.y) * 2) {
          // Horizontal swipe
          return this.createGesture(GestureType.Swipe, gestureState, timestamp);
        } else if (isSwipe && Math.abs(totalDelta.y) > Math.abs(totalDelta.x) * 2) {
          // Vertical swipe
          return this.createGesture(GestureType.Swipe, gestureState, timestamp);
        } else {
          // Pan gesture
          return this.createGesture(GestureType.Pan, gestureState, timestamp);
        }
      }
    } else if (touches.length === 2 && gestureState.initialDistance) {
      // Two touches - pinch or rotate
      const currentDistance = this.calculateDistance(touches[0], touches[1]);
      const currentAngle = this.calculateAngle(touches[0], touches[1]);
      
      gestureState.scale = currentDistance / gestureState.initialDistance;
      gestureState.rotation = currentAngle - (gestureState.initialAngle || 0);

      // Determine if it's primarily a pinch or rotate
      const scaleChange = Math.abs(gestureState.scale - 1);
      const rotationChange = Math.abs(gestureState.rotation);

      if (scaleChange > 0.1) {
        return this.createGesture(GestureType.Pinch, gestureState, timestamp);
      } else if (rotationChange > 0.1) {
        return this.createGesture(GestureType.Rotate, gestureState, timestamp);
      }
    }

    return null;
  }

  private handleTouchEnd(touchEvent: TouchEvent, gestureId: string, timestamp: number): RecognizedGesture | null {
    const gestureState = this.activeGestures.get(gestureId);
    if (!gestureState) return null;

    const duration = timestamp - gestureState.startTime;
    
    // Determine final gesture type
    if (gestureState.totalDistance < this.options.gestureThreshold) {
      // Minimal movement - tap or long press
      if (duration > 500) {
        // Long press
        const gesture = this.createGesture(GestureType.LongPress, gestureState, timestamp);
        this.activeGestures.delete(gestureId);
        return gesture;
      } else {
        // Check for double tap
        const recentTap = this.findRecentTap(gestureState.startPosition, timestamp);
        if (recentTap && timestamp - recentTap.timestamp < 300) {
          // Double tap
          const gesture = this.createGesture(GestureType.DoubleTap, gestureState, timestamp);
          this.activeGestures.delete(gestureId);
          return gesture;
        } else {
          // Single tap
          const gesture = this.createGesture(GestureType.Tap, gestureState, timestamp);
          this.activeGestures.delete(gestureId);
          
          // Record tap for double-tap detection
          this.recordGestureEvent({
            type: GestureType.Tap,
            position: gestureState.currentPosition,
            timestamp: timestamp
          });
          
          return gesture;
        }
      }
    }

    // Movement-based gesture already detected in touchmove
    this.activeGestures.delete(gestureId);
    return null;
  }

  private handleTouchCancel(gestureId: string): RecognizedGesture | null {
    this.activeGestures.delete(gestureId);
    return null;
  }

  private calculateDistance(touch1: Touch, touch2: Touch): number {
    const dx = touch2.clientX - touch1.clientX;
    const dy = touch2.clientY - touch1.clientY;
    return Math.sqrt(dx * dx + dy * dy);
  }

  private calculateAngle(touch1: Touch, touch2: Touch): number {
    return Math.atan2(touch2.clientY - touch1.clientY, touch2.clientX - touch1.clientX);
  }

  private createGesture(type: GestureType, state: GestureState, timestamp: number): RecognizedGesture {
    return {
      type,
      position: state.currentPosition,
      startPosition: state.startPosition,
      confidence: this.calculateConfidence(type, state),
      duration: timestamp - state.startTime,
      velocity: state.velocity,
      scale: state.scale,
      rotation: state.rotation,
      distance: state.totalDistance
    };
  }

  private calculateConfidence(type: GestureType, state: GestureState): number {
    // Simple confidence calculation based on gesture characteristics
    switch (type) {
      case GestureType.Tap:
        return state.totalDistance < 5 ? 0.9 : 0.7;
      case GestureType.Swipe:
        return state.totalDistance > 50 ? 0.9 : 0.6;
      case GestureType.Pan:
        return 0.8;
      case GestureType.Pinch:
        return Math.abs(state.scale - 1) > 0.2 ? 0.9 : 0.6;
      case GestureType.Rotate:
        return Math.abs(state.rotation) > 0.2 ? 0.9 : 0.6;
      default:
        return 0.5;
    }
  }

  private findRecentTap(position: Position, timestamp: number): GestureEvent | null {
    const threshold = 50; // pixels
    const timeThreshold = 300; // milliseconds

    for (let i = this.gestureHistory.length - 1; i >= 0; i--) {
      const event = this.gestureHistory[i];
      if (timestamp - event.timestamp > timeThreshold) break;
      
      if (event.type === GestureType.Tap) {
        const distance = Math.sqrt(
          Math.pow(event.position.x - position.x, 2) + 
          Math.pow(event.position.y - position.y, 2)
        );
        
        if (distance < threshold) {
          return event;
        }
      }
    }
    
    return null;
  }

  private recordGestureEvent(event: GestureEvent): void {
    this.gestureHistory.push(event);
    
    // Keep only recent events
    const cutoff = performance.now() - 1000; // 1 second
    this.gestureHistory = this.gestureHistory.filter(e => e.timestamp > cutoff);
  }

  updateOptions(options: Required<MobileOptimizationOptions>): void {
    this.options = options;
  }

  destroy(): void {
    this.activeGestures.clear();
    this.gestureHistory = [];
    this.isActive = false;
  }
}

// Supporting interfaces and classes
interface GestureState {
  id: string;
  startTime: number;
  startPosition: Position;
  currentPosition: Position;
  touchCount: number;
  lastMoveTime: number;
  totalDistance: number;
  velocity: Position;
  scale: number;
  rotation: number;
  initialDistance?: number;
  initialAngle?: number;
}

interface GestureEvent {
  type: GestureType;
  position: Position;
  timestamp: number;
}

// Touch Optimization System
class TouchOptimizer {
  private options: Required<MobileOptimizationOptions>;
  private touchBuffer: TouchEvent[] = [];
  private lastProcessedTime: number = 0;

  constructor(options: Required<MobileOptimizationOptions>) {
    this.options = options;
  }

  optimizeTouch(touchEvent: TouchEvent): TouchEvent {
    // Throttle touch events to improve performance
    const now = performance.now();
    if (now - this.lastProcessedTime < 16) { // ~60fps
      return touchEvent; // Skip processing for high-frequency events
    }
    
    this.lastProcessedTime = now;
    
    // Limit number of touch points
    if (touchEvent.touches.length > this.options.maxTouchPoints) {
      // Create a new TouchEvent with limited touches
      // Note: This is a simplified approach - in practice, you'd need to properly clone the event
      return touchEvent;
    }
    
    return touchEvent;
  }

  destroy(): void {
    this.touchBuffer = [];
  }
}

// Mobile Performance Management
export interface MobilePerformanceMetrics {
  frameRate: number;
  memoryUsage: number;
  touchLatency: number;
  gestureAccuracy: number;
  batteryLevel?: number;
  networkType?: string;
}

class MobilePerformanceManager {
  private options: Required<MobileOptimizationOptions>;
  private metrics: MobilePerformanceMetrics;
  private frameCount: number = 0;
  private lastFrameTime: number = 0;
  private performanceObserver?: PerformanceObserver;

  constructor(options: Required<MobileOptimizationOptions>) {
    this.options = options;
    this.metrics = {
      frameRate: 60,
      memoryUsage: 0,
      touchLatency: 0,
      gestureAccuracy: 0
    };
  }

  initialize(container: HTMLElement): void {
    // Monitor frame rate
    this.startFrameRateMonitoring();
    
    // Monitor memory usage if available
    this.startMemoryMonitoring();
    
    // Monitor network conditions
    this.startNetworkMonitoring();
    
    // Monitor battery if available
    this.startBatteryMonitoring();
  }

  private startFrameRateMonitoring(): void {
    const measureFrameRate = (timestamp: number) => {
      if (this.lastFrameTime > 0) {
        const delta = timestamp - this.lastFrameTime;
        this.frameCount++;
        
        if (this.frameCount % 60 === 0) { // Update every 60 frames
          this.metrics.frameRate = 1000 / (delta / 60);
        }
      }
      
      this.lastFrameTime = timestamp;
      requestAnimationFrame(measureFrameRate);
    };
    
    requestAnimationFrame(measureFrameRate);
  }

  private startMemoryMonitoring(): void {
    if ('memory' in performance) {
      setInterval(() => {
        const memory = (performance as any).memory;
        this.metrics.memoryUsage = memory.usedJSHeapSize / memory.jsHeapSizeLimit;
      }, 5000);
    }
  }

  private startNetworkMonitoring(): void {
    if ('connection' in navigator) {
      const connection = (navigator as any).connection;
      this.metrics.networkType = connection.effectiveType;
      
      connection.addEventListener('change', () => {
        this.metrics.networkType = connection.effectiveType;
      });
    }
  }

  private startBatteryMonitoring(): void {
    if ('getBattery' in navigator) {
      (navigator as any).getBattery().then((battery: any) => {
        this.metrics.batteryLevel = battery.level;
        
        battery.addEventListener('levelchange', () => {
          this.metrics.batteryLevel = battery.level;
        });
      });
    }
  }

  getMetrics(): MobilePerformanceMetrics {
    return { ...this.metrics };
  }

  destroy(): void {
    if (this.performanceObserver) {
      this.performanceObserver.disconnect();
    }
  }
}