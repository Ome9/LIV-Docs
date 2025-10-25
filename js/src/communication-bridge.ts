/**
 * LIV Document Format - Communication Bridge
 * Handles secure communication between JavaScript and Go backend
 */

import { SandboxMessage, MessageType } from './sandbox-interface';

export interface BridgeConfig {
  enableEncryption: boolean;
  enableCompression: boolean;
  maxMessageSize: number;
  heartbeatInterval: number;
  reconnectAttempts: number;
  reconnectDelay: number;
}

export interface BridgeStats {
  messagesSent: number;
  messagesReceived: number;
  bytesTransferred: number;
  errors: number;
  lastHeartbeat: number;
  connectionStatus: ConnectionStatus;
}

export enum ConnectionStatus {
  DISCONNECTED = 'disconnected',
  CONNECTING = 'connecting',
  CONNECTED = 'connected',
  ERROR = 'error'
}

export type MessageHandler = (message: SandboxMessage) => void;
export type ErrorHandler = (error: Error) => void;
export type StatusHandler = (status: ConnectionStatus) => void;

/**
 * Communication bridge for secure JavaScript-Go communication
 */
export class CommunicationBridge {
  private config: BridgeConfig;
  private stats: BridgeStats;
  private messageHandlers: Set<MessageHandler>;
  private errorHandlers: Set<ErrorHandler>;
  private statusHandlers: Set<StatusHandler>;
  private heartbeatTimer?: number | undefined;
  private reconnectTimer?: number | undefined;
  private isDestroyed: boolean = false;

  constructor(config?: Partial<BridgeConfig>) {
    this.config = {
      enableEncryption: false,
      enableCompression: false,
      maxMessageSize: 1024 * 1024, // 1MB
      heartbeatInterval: 30000, // 30 seconds
      reconnectAttempts: 3,
      reconnectDelay: 1000, // 1 second
      ...config
    };

    this.stats = {
      messagesSent: 0,
      messagesReceived: 0,
      bytesTransferred: 0,
      errors: 0,
      lastHeartbeat: 0,
      connectionStatus: ConnectionStatus.DISCONNECTED
    };

    this.messageHandlers = new Set();
    this.errorHandlers = new Set();
    this.statusHandlers = new Set();

    this.setupGlobalHandlers();
  }

  /**
   * Initialize the communication bridge
   */
  async initialize(): Promise<void> {
    if (this.isDestroyed) {
      throw new Error('Bridge has been destroyed');
    }

    try {
      this.setStatus(ConnectionStatus.CONNECTING);
      
      // Check if Go bridge is available
      if (!this.isGoBridgeAvailable()) {
        throw new Error('Go communication bridge not available');
      }

      // Send initialization message
      const initMessage: SandboxMessage = {
        id: this.generateMessageId(),
        type: MessageType.CONTROL,
        source: 'js_bridge',
        target: 'go_bridge',
        payload: {
          command: 'initialize',
          config: this.config
        },
        timestamp: Date.now()
      };

      await this.sendMessage(initMessage);
      
      this.setStatus(ConnectionStatus.CONNECTED);
      this.startHeartbeat();
      
      console.log('[CommunicationBridge] Initialized successfully');
    } catch (error) {
      this.setStatus(ConnectionStatus.ERROR);
      this.handleError(error as Error);
      throw error;
    }
  }

  /**
   * Send a message to the Go backend
   */
  async sendMessage(message: SandboxMessage): Promise<void> {
    if (this.isDestroyed) {
      throw new Error('Bridge has been destroyed');
    }

    if (this.stats.connectionStatus !== ConnectionStatus.CONNECTED) {
      throw new Error('Bridge not connected');
    }

    try {
      // Validate message
      this.validateMessage(message);
      
      // Process message
      const processedMessage = await this.processOutgoingMessage(message);
      
      // Send to Go backend
      this.postToGo(processedMessage);
      
      // Update stats
      this.stats.messagesSent++;
      this.stats.bytesTransferred += JSON.stringify(processedMessage).length;
      
    } catch (error) {
      this.stats.errors++;
      this.handleError(error as Error);
      throw error;
    }
  }

  /**
   * Register a message handler
   */
  onMessage(handler: MessageHandler): void {
    this.messageHandlers.add(handler);
  }

  /**
   * Unregister a message handler
   */
  offMessage(handler: MessageHandler): void {
    this.messageHandlers.delete(handler);
  }

  /**
   * Register an error handler
   */
  onError(handler: ErrorHandler): void {
    this.errorHandlers.add(handler);
  }

  /**
   * Unregister an error handler
   */
  offError(handler: ErrorHandler): void {
    this.errorHandlers.delete(handler);
  }

  /**
   * Register a status change handler
   */
  onStatusChange(handler: StatusHandler): void {
    this.statusHandlers.add(handler);
  }

  /**
   * Unregister a status change handler
   */
  offStatusChange(handler: StatusHandler): void {
    this.statusHandlers.delete(handler);
  }

  /**
   * Get bridge statistics
   */
  getStats(): BridgeStats {
    return { ...this.stats };
  }

  /**
   * Get bridge configuration
   */
  getConfig(): BridgeConfig {
    return { ...this.config };
  }

  /**
   * Check if bridge is connected
   */
  isConnected(): boolean {
    return this.stats.connectionStatus === ConnectionStatus.CONNECTED;
  }

  /**
   * Reconnect to Go backend
   */
  async reconnect(): Promise<void> {
    if (this.isDestroyed) {
      throw new Error('Bridge has been destroyed');
    }

    this.stopHeartbeat();
    
    let attempts = 0;
    while (attempts < this.config.reconnectAttempts) {
      try {
        await this.initialize();
        return;
      } catch (error) {
        attempts++;
        if (attempts < this.config.reconnectAttempts) {
          await this.delay(this.config.reconnectDelay * attempts);
        }
      }
    }
    
    throw new Error(`Failed to reconnect after ${this.config.reconnectAttempts} attempts`);
  }

  /**
   * Destroy the bridge and clean up resources
   */
  destroy(): void {
    if (this.isDestroyed) {
      return;
    }

    try {
      // Send destruction message if connected
      if (this.stats.connectionStatus === ConnectionStatus.CONNECTED) {
        const destroyMessage: SandboxMessage = {
          id: this.generateMessageId(),
          type: MessageType.CONTROL,
          source: 'js_bridge',
          target: 'go_bridge',
          payload: {
            command: 'destroy'
          },
          timestamp: Date.now()
        };
        
        this.postToGo(destroyMessage);
      }
    } catch (error) {
      // Ignore errors during destruction
    }

    // Clean up
    this.stopHeartbeat();
    this.clearReconnectTimer();
    this.messageHandlers.clear();
    this.errorHandlers.clear();
    this.statusHandlers.clear();
    this.setStatus(ConnectionStatus.DISCONNECTED);
    this.isDestroyed = true;

    console.log('[CommunicationBridge] Destroyed');
  }

  // Private methods

  private setupGlobalHandlers(): void {
    // Listen for messages from Go
    if (typeof (globalThis as any).handleGoMessage === 'undefined') {
      (globalThis as any).handleGoMessage = (messageJson: string) => {
        try {
          const message: SandboxMessage = JSON.parse(messageJson);
          this.handleIncomingMessage(message);
        } catch (error) {
          this.handleError(new Error(`Failed to parse Go message: ${error}`));
        }
      };
    }

    // Listen for Go bridge status changes
    if (typeof (globalThis as any).handleGoBridgeStatus === 'undefined') {
      (globalThis as any).handleGoBridgeStatus = (status: string) => {
        switch (status) {
          case 'connected':
            this.setStatus(ConnectionStatus.CONNECTED);
            break;
          case 'disconnected':
            this.setStatus(ConnectionStatus.DISCONNECTED);
            break;
          case 'error':
            this.setStatus(ConnectionStatus.ERROR);
            break;
        }
      };
    }
  }

  private async handleIncomingMessage(message: SandboxMessage): Promise<void> {
    try {
      // Update stats
      this.stats.messagesReceived++;
      this.stats.bytesTransferred += JSON.stringify(message).length;
      
      // Process message
      const processedMessage = await this.processIncomingMessage(message);
      
      // Handle heartbeat messages
      if (processedMessage.type === MessageType.HEARTBEAT) {
        this.stats.lastHeartbeat = Date.now();
        return;
      }
      
      // Notify handlers
      for (const handler of this.messageHandlers) {
        try {
          handler(processedMessage);
        } catch (error) {
          this.handleError(new Error(`Message handler error: ${error}`));
        }
      }
    } catch (error) {
      this.stats.errors++;
      this.handleError(error as Error);
    }
  }

  private async processOutgoingMessage(message: SandboxMessage): Promise<SandboxMessage> {
    let processed = { ...message };
    
    // Compress if enabled
    if (this.config.enableCompression) {
      processed = await this.compressMessage(processed);
    }
    
    // Encrypt if enabled
    if (this.config.enableEncryption) {
      processed = await this.encryptMessage(processed);
    }
    
    return processed;
  }

  private async processIncomingMessage(message: SandboxMessage): Promise<SandboxMessage> {
    let processed = { ...message };
    
    // Decrypt if enabled
    if (this.config.enableEncryption) {
      processed = await this.decryptMessage(processed);
    }
    
    // Decompress if enabled
    if (this.config.enableCompression) {
      processed = await this.decompressMessage(processed);
    }
    
    return processed;
  }

  private validateMessage(message: SandboxMessage): void {
    if (!message.id || !message.type || !message.source || !message.target) {
      throw new Error('Invalid message format');
    }
    
    const messageSize = JSON.stringify(message).length;
    if (messageSize > this.config.maxMessageSize) {
      throw new Error(`Message size ${messageSize} exceeds limit ${this.config.maxMessageSize}`);
    }
  }

  private postToGo(message: SandboxMessage): void {
    const messageJson = JSON.stringify(message);
    
    if (typeof (globalThis as any).goSandboxBridge !== 'undefined') {
      // Use Go bridge if available
      (globalThis as any).goSandboxBridge.handleMessage(messageJson);
    } else if (typeof (globalThis as any).postMessage !== 'undefined') {
      // Use postMessage API
      (globalThis as any).postMessage({
        type: 'sandbox_message',
        data: messageJson
      }, '*');
    } else {
      throw new Error('No communication method available to Go backend');
    }
  }

  private isGoBridgeAvailable(): boolean {
    return typeof (globalThis as any).goSandboxBridge !== 'undefined' ||
           typeof (globalThis as any).postMessage !== 'undefined';
  }

  private startHeartbeat(): void {
    this.stopHeartbeat();
    
    this.heartbeatTimer = window.setInterval(() => {
      if (this.stats.connectionStatus === ConnectionStatus.CONNECTED) {
        const heartbeatMessage: SandboxMessage = {
          id: this.generateMessageId(),
          type: MessageType.HEARTBEAT,
          source: 'js_bridge',
          target: 'go_bridge',
          payload: {
            timestamp: Date.now(),
            stats: this.stats
          },
          timestamp: Date.now()
        };
        
        try {
          this.postToGo(heartbeatMessage);
        } catch (error) {
          this.handleError(new Error(`Heartbeat failed: ${error}`));
        }
      }
    }, this.config.heartbeatInterval);
  }

  private stopHeartbeat(): void {
    if (this.heartbeatTimer) {
      clearInterval(this.heartbeatTimer);
      this.heartbeatTimer = undefined;
    }
  }

  private clearReconnectTimer(): void {
    if (this.reconnectTimer) {
      clearTimeout(this.reconnectTimer);
      this.reconnectTimer = undefined;
    }
  }

  private setStatus(status: ConnectionStatus): void {
    if (this.stats.connectionStatus !== status) {
      this.stats.connectionStatus = status;
      
      for (const handler of this.statusHandlers) {
        try {
          handler(status);
        } catch (error) {
          console.error('[CommunicationBridge] Status handler error:', error);
        }
      }
    }
  }

  private handleError(error: Error): void {
    console.error('[CommunicationBridge] Error:', error);
    
    for (const handler of this.errorHandlers) {
      try {
        handler(error);
      } catch (handlerError) {
        console.error('[CommunicationBridge] Error handler error:', handlerError);
      }
    }
  }

  private generateMessageId(): string {
    return `msg_${Date.now()}_${Math.random().toString(36).substr(2, 9)}`;
  }

  private delay(ms: number): Promise<void> {
    return new Promise(resolve => setTimeout(resolve, ms));
  }

  // Placeholder methods for encryption/compression
  // In production, these would use proper libraries

  private async compressMessage(message: SandboxMessage): Promise<SandboxMessage> {
    // Placeholder for compression logic
    // Could use libraries like pako for gzip compression
    return message;
  }

  private async decompressMessage(message: SandboxMessage): Promise<SandboxMessage> {
    // Placeholder for decompression logic
    return message;
  }

  private async encryptMessage(message: SandboxMessage): Promise<SandboxMessage> {
    // Placeholder for encryption logic
    // Could use Web Crypto API for encryption
    return message;
  }

  private async decryptMessage(message: SandboxMessage): Promise<SandboxMessage> {
    // Placeholder for decryption logic
    return message;
  }
}

/**
 * Factory function to create a communication bridge
 */
export function createCommunicationBridge(config?: Partial<BridgeConfig>): CommunicationBridge {
  return new CommunicationBridge(config);
}

/**
 * Default bridge configuration
 */
export const DEFAULT_BRIDGE_CONFIG: BridgeConfig = {
  enableEncryption: false,
  enableCompression: false,
  maxMessageSize: 1024 * 1024, // 1MB
  heartbeatInterval: 30000, // 30 seconds
  reconnectAttempts: 3,
  reconnectDelay: 1000 // 1 second
};