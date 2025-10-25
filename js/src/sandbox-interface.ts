/**
 * LIV Document Format - JavaScript Sandbox Interface
 * Provides secure communication between JavaScript and Go WASM runtime
 */

export interface SecurityPolicy {
  wasmPermissions: WASMPermissions;
  jsPermissions: JSPermissions;
  networkPolicy: NetworkPolicy;
  storagePolicy: StoragePolicy;
  contentSecurityPolicy?: string;
  trustedDomains?: string[];
}

export interface WASMPermissions {
  memoryLimit: number;
  cpuTimeLimit: number;
  allowNetworking: boolean;
  allowFileSystem: boolean;
  allowedImports: string[];
}

export interface JSPermissions {
  executionMode: 'sandboxed' | 'trusted';
  allowedAPIs: string[];
  domAccess: 'none' | 'read' | 'write';
}

export interface NetworkPolicy {
  allowOutbound: boolean;
  allowedHosts: string[];
  allowedPorts: number[];
}

export interface StoragePolicy {
  allowLocalStorage: boolean;
  allowSessionStorage: boolean;
  allowIndexedDB: boolean;
  allowCookies: boolean;
}

export interface SandboxMessage {
  id: string;
  type: MessageType;
  source: string;
  target: string;
  payload: Record<string, any>;
  timestamp: number;
  response?: boolean;
}

export enum MessageType {
  FUNCTION_CALL = 'function_call',
  EVENT = 'event',
  DATA = 'data',
  CONTROL = 'control',
  RESPONSE = 'response',
  ERROR = 'error',
  HEARTBEAT = 'heartbeat'
}

export interface WASMModuleConfig {
  name: string;
  version: string;
  entryPoint: string;
  exports: string[];
  imports: string[];
  permissions: WASMPermissions;
  metadata?: Record<string, string>;
}

export interface SandboxConfig {
  securityPolicy: SecurityPolicy;
  enableLogging: boolean;
  enableMetrics: boolean;
  timeoutMs: number;
  maxMemoryMB: number;
}

export interface ExecutionResult {
  success: boolean;
  result?: any;
  error?: string;
  duration: number;
  memoryUsed: number;
}

export interface SandboxStats {
  sessionId: string;
  uptime: number;
  memoryUsage: number;
  cpuTime: number;
  functionCalls: number;
  networkRequests: number;
  errors: number;
  lastActivity: number;
}

/**
 * Main Sandbox Interface for secure JavaScript execution
 */
export class SandboxInterface {
  private sessionId: string;
  private config: SandboxConfig;
  private messageHandlers: Map<MessageType, (message: SandboxMessage) => void>;
  private responseCallbacks: Map<string, (response: SandboxMessage) => void>;
  private wasmModules: Map<string, WASMModuleConfig>;
  private stats: SandboxStats;
  private isInitialized: boolean = false;
  private isDestroyed: boolean = false;

  constructor(config: SandboxConfig) {
    this.sessionId = this.generateSessionId();
    this.config = config;
    this.messageHandlers = new Map();
    this.responseCallbacks = new Map();
    this.wasmModules = new Map();
    this.stats = {
      sessionId: this.sessionId,
      uptime: 0,
      memoryUsage: 0,
      cpuTime: 0,
      functionCalls: 0,
      networkRequests: 0,
      errors: 0,
      lastActivity: Date.now()
    };

    this.setupDefaultHandlers();
    this.startStatsTracking();
  }

  /**
   * Initialize the sandbox with the Go backend
   */
  async initialize(): Promise<void> {
    if (this.isInitialized) {
      throw new Error('Sandbox already initialized');
    }

    if (this.isDestroyed) {
      throw new Error('Cannot initialize destroyed sandbox');
    }

    try {
      // Send initialization message to Go backend
      const initMessage: SandboxMessage = {
        id: this.generateMessageId(),
        type: MessageType.CONTROL,
        source: 'js_sandbox',
        target: 'go_runtime',
        payload: {
          command: 'initialize',
          sessionId: this.sessionId,
          config: this.config
        },
        timestamp: Date.now()
      };

      const response = await this.sendMessage(initMessage);
      
      if (!response || response.type === MessageType.ERROR) {
        throw new Error(`Initialization failed: ${response?.payload?.error || 'Unknown error'}`);
      }

      this.isInitialized = true;
      this.log('Sandbox initialized successfully', { sessionId: this.sessionId });
    } catch (error) {
      this.handleError('Initialization failed', error);
      throw error;
    }
  }

  /**
   * Load a WASM module into the sandbox
   */
  async loadWASMModule(moduleData: ArrayBuffer, config: WASMModuleConfig): Promise<void> {
    this.ensureInitialized();

    try {
      // Validate module configuration
      this.validateWASMConfig(config);

      // Convert ArrayBuffer to base64 for transmission
      const moduleBase64 = this.arrayBufferToBase64(moduleData);

      const loadMessage: SandboxMessage = {
        id: this.generateMessageId(),
        type: MessageType.CONTROL,
        source: 'js_sandbox',
        target: 'go_runtime',
        payload: {
          command: 'load_wasm',
          sessionId: this.sessionId,
          moduleName: config.name,
          moduleData: moduleBase64,
          config: config
        },
        timestamp: Date.now()
      };

      const response = await this.sendMessage(loadMessage);
      
      if (!response || response.type === MessageType.ERROR) {
        throw new Error(`WASM module loading failed: ${response?.payload?.error || 'Unknown error'}`);
      }

      this.wasmModules.set(config.name, config);
      this.log('WASM module loaded successfully', { moduleName: config.name });
    } catch (error) {
      this.handleError('WASM module loading failed', error);
      throw error;
    }
  }

  /**
   * Execute a function in a WASM module
   */
  async executeWASMFunction(
    moduleName: string, 
    functionName: string, 
    args: any[] = []
  ): Promise<ExecutionResult> {
    this.ensureInitialized();

    const startTime = performance.now();
    
    try {
      // Validate module exists
      if (!this.wasmModules.has(moduleName)) {
        throw new Error(`WASM module '${moduleName}' not loaded`);
      }

      const module = this.wasmModules.get(moduleName)!;
      
      // Validate function is exported
      if (!module.exports.includes(functionName)) {
        throw new Error(`Function '${functionName}' not exported by module '${moduleName}'`);
      }

      const executeMessage: SandboxMessage = {
        id: this.generateMessageId(),
        type: MessageType.FUNCTION_CALL,
        source: 'js_sandbox',
        target: 'go_runtime',
        payload: {
          sessionId: this.sessionId,
          moduleName: moduleName,
          functionName: functionName,
          arguments: args
        },
        timestamp: Date.now()
      };

      const response = await this.sendMessage(executeMessage);
      const duration = performance.now() - startTime;
      
      this.stats.functionCalls++;
      this.stats.cpuTime += duration;
      this.stats.lastActivity = Date.now();

      if (!response) {
        throw new Error('No response received from WASM function execution');
      }

      if (response.type === MessageType.ERROR) {
        this.stats.errors++;
        return {
          success: false,
          error: response.payload?.error || 'Unknown execution error',
          duration: duration,
          memoryUsed: response.payload?.memoryUsed || 0
        };
      }

      return {
        success: true,
        result: response.payload?.result,
        duration: duration,
        memoryUsed: response.payload?.memoryUsed || 0
      };
    } catch (error) {
      const duration = performance.now() - startTime;
      this.stats.errors++;
      this.handleError('WASM function execution failed', error);
      
      return {
        success: false,
        error: error instanceof Error ? error.message : 'Unknown error',
        duration: duration,
        memoryUsed: 0
      };
    }
  }

  /**
   * Send an event to the Go backend
   */
  async sendEvent(eventType: string, eventData: any): Promise<void> {
    this.ensureInitialized();

    try {
      const eventMessage: SandboxMessage = {
        id: this.generateMessageId(),
        type: MessageType.EVENT,
        source: 'js_sandbox',
        target: 'go_runtime',
        payload: {
          eventType: eventType,
          data: eventData,
          sessionId: this.sessionId
        },
        timestamp: Date.now(),
        response: false
      };

      await this.sendMessage(eventMessage);
      this.log('Event sent', { eventType, sessionId: this.sessionId });
    } catch (error) {
      this.handleError('Event sending failed', error);
      throw error;
    }
  }

  /**
   * Register a message handler for specific message types
   */
  onMessage(type: MessageType, handler: (message: SandboxMessage) => void): void {
    this.messageHandlers.set(type, handler);
  }

  /**
   * Get current sandbox statistics
   */
  getStats(): SandboxStats {
    this.stats.uptime = Date.now() - (this.stats.lastActivity - this.stats.uptime);
    return { ...this.stats };
  }

  /**
   * Get loaded WASM modules
   */
  getLoadedModules(): string[] {
    return Array.from(this.wasmModules.keys());
  }

  /**
   * Get WASM module configuration
   */
  getModuleConfig(moduleName: string): WASMModuleConfig | undefined {
    return this.wasmModules.get(moduleName);
  }

  /**
   * Destroy the sandbox and clean up resources
   */
  async destroy(): Promise<void> {
    if (this.isDestroyed) {
      return;
    }

    try {
      if (this.isInitialized) {
        const destroyMessage: SandboxMessage = {
          id: this.generateMessageId(),
          type: MessageType.CONTROL,
          source: 'js_sandbox',
          target: 'go_runtime',
          payload: {
            command: 'destroy',
            sessionId: this.sessionId
          },
          timestamp: Date.now()
        };

        await this.sendMessage(destroyMessage);
      }

      // Clean up resources
      this.messageHandlers.clear();
      this.responseCallbacks.clear();
      this.wasmModules.clear();
      this.isDestroyed = true;
      this.isInitialized = false;

      this.log('Sandbox destroyed', { sessionId: this.sessionId });
    } catch (error) {
      this.handleError('Sandbox destruction failed', error);
      throw error;
    }
  }

  /**
   * Handle incoming messages from Go backend
   */
  handleIncomingMessage(message: SandboxMessage): void {
    try {
      // Handle response messages
      if (message.response && this.responseCallbacks.has(message.id)) {
        const callback = this.responseCallbacks.get(message.id)!;
        callback(message);
        return;
      }

      // Handle regular messages
      const handler = this.messageHandlers.get(message.type);
      if (handler) {
        handler(message);
      } else {
        this.log('No handler for message type', { type: message.type });
      }
    } catch (error) {
      this.handleError('Message handling failed', error);
    }
  }

  // Private methods

  /**
   * Send a message to the Go backend and wait for response
   */
  private async sendMessage(message: SandboxMessage): Promise<SandboxMessage | null> {
    return new Promise((resolve, reject) => {
      const timeout = setTimeout(() => {
        this.responseCallbacks.delete(message.id);
        reject(new Error(`Message timeout: ${message.id}`));
      }, this.config.timeoutMs);

      // Store response callback
      this.responseCallbacks.set(message.id, (response: SandboxMessage) => {
        clearTimeout(timeout);
        this.responseCallbacks.delete(message.id);
        resolve(response);
      });

      try {
        // Send message to Go backend via postMessage or WebAssembly interface
        this.postMessageToGo(message);
      } catch (error) {
        clearTimeout(timeout);
        this.responseCallbacks.delete(message.id);
        reject(error);
      }
    });
  }

  /**
   * Post message to Go backend (implementation depends on integration method)
   */
  private postMessageToGo(message: SandboxMessage): void {
    if (typeof (globalThis as any).goSandboxBridge !== 'undefined') {
      // Use Go bridge if available
      (globalThis as any).goSandboxBridge.handleMessage(JSON.stringify(message));
    } else if (typeof (globalThis as any).postMessage !== 'undefined') {
      // Use postMessage API
      (globalThis as any).postMessage(message, '*');
    } else {
      throw new Error('No communication bridge available to Go backend');
    }
  }

  /**
   * Setup default message handlers
   */
  private setupDefaultHandlers(): void {
    this.onMessage(MessageType.HEARTBEAT, (message) => {
      this.log('Heartbeat received', { timestamp: message.timestamp });
    });

    this.onMessage(MessageType.EVENT, (message) => {
      this.log('Event received', { eventType: message.payload?.eventType });
    });

    this.onMessage(MessageType.ERROR, (message) => {
      this.handleError('Error message received', new Error(message.payload?.error));
    });
  }

  /**
   * Start tracking sandbox statistics
   */
  private startStatsTracking(): void {
    const startTime = Date.now();
    
    setInterval(() => {
      if (!this.isDestroyed) {
        this.stats.uptime = Date.now() - startTime;
        
        // Update memory usage if available
        if (typeof (performance as any).memory !== 'undefined') {
          this.stats.memoryUsage = (performance as any).memory.usedJSHeapSize;
        }
      }
    }, 1000); // Update every second
  }

  /**
   * Validate WASM module configuration
   */
  private validateWASMConfig(config: WASMModuleConfig): void {
    if (!config.name || config.name.trim() === '') {
      throw new Error('WASM module name is required');
    }

    if (!config.version || config.version.trim() === '') {
      throw new Error('WASM module version is required');
    }

    if (!config.entryPoint || config.entryPoint.trim() === '') {
      throw new Error('WASM module entry point is required');
    }

    if (!Array.isArray(config.exports) || config.exports.length === 0) {
      throw new Error('WASM module must have at least one export');
    }

    if (!config.permissions) {
      throw new Error('WASM module permissions are required');
    }

    // Validate permissions against sandbox policy
    const sandboxPerms = this.config.securityPolicy.wasmPermissions;
    
    if (config.permissions.memoryLimit > sandboxPerms.memoryLimit) {
      throw new Error(`WASM module memory limit exceeds sandbox limit: ${config.permissions.memoryLimit} > ${sandboxPerms.memoryLimit}`);
    }

    if (config.permissions.cpuTimeLimit > sandboxPerms.cpuTimeLimit) {
      throw new Error(`WASM module CPU time limit exceeds sandbox limit: ${config.permissions.cpuTimeLimit} > ${sandboxPerms.cpuTimeLimit}`);
    }

    if (config.permissions.allowNetworking && !sandboxPerms.allowNetworking) {
      throw new Error('WASM module requests networking but sandbox policy denies it');
    }

    if (config.permissions.allowFileSystem && !sandboxPerms.allowFileSystem) {
      throw new Error('WASM module requests file system access but sandbox policy denies it');
    }
  }

  /**
   * Ensure sandbox is initialized
   */
  private ensureInitialized(): void {
    if (!this.isInitialized) {
      throw new Error('Sandbox not initialized');
    }

    if (this.isDestroyed) {
      throw new Error('Sandbox has been destroyed');
    }
  }

  /**
   * Convert ArrayBuffer to base64 string
   */
  private arrayBufferToBase64(buffer: ArrayBuffer): string {
    const bytes = new Uint8Array(buffer);
    let binary = '';
    for (let i = 0; i < bytes.byteLength; i++) {
      binary += String.fromCharCode(bytes[i]);
    }
    return btoa(binary);
  }

  /**
   * Generate unique message ID
   */
  private generateMessageId(): string {
    return `msg_${Date.now()}_${Math.random().toString(36).substr(2, 9)}`;
  }

  /**
   * Generate unique session ID
   */
  private generateSessionId(): string {
    return `session_${Date.now()}_${Math.random().toString(36).substr(2, 9)}`;
  }

  /**
   * Log message with optional data
   */
  private log(message: string, data?: any): void {
    if (this.config.enableLogging) {
      console.log(`[Sandbox ${this.sessionId}] ${message}`, data || '');
    }
  }

  /**
   * Handle errors with logging and metrics
   */
  private handleError(message: string, error: any): void {
    this.stats.errors++;
    
    if (this.config.enableLogging) {
      console.error(`[Sandbox ${this.sessionId}] ${message}:`, error);
    }
  }
}

/**
 * Factory function to create a new sandbox instance
 */
export function createSandbox(config: SandboxConfig): SandboxInterface {
  return new SandboxInterface(config);
}

/**
 * Default sandbox configuration
 */
export const DEFAULT_SANDBOX_CONFIG: SandboxConfig = {
  securityPolicy: {
    wasmPermissions: {
      memoryLimit: 16 * 1024 * 1024, // 16MB
      cpuTimeLimit: 5000, // 5 seconds
      allowNetworking: false,
      allowFileSystem: false,
      allowedImports: ['env']
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
  },
  enableLogging: true,
  enableMetrics: true,
  timeoutMs: 30000, // 30 seconds
  maxMemoryMB: 64 // 64MB
};