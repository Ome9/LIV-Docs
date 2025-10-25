/**
 * Tests for LIV Document Format - Communication Bridge
 */

import {
  CommunicationBridge,
  createCommunicationBridge,
  DEFAULT_BRIDGE_CONFIG,
  ConnectionStatus,
  BridgeConfig,
  MessageHandler,
  ErrorHandler,
  StatusHandler
} from '../src/communication-bridge';
import { SandboxMessage, MessageType } from '../src/sandbox-interface';

// Mock global Go bridge
const mockGoBridge = {
  handleMessage: jest.fn()
};

describe('CommunicationBridge', () => {
  let bridge: CommunicationBridge;
  let config: BridgeConfig;

  beforeEach(() => {
    config = {
      ...DEFAULT_BRIDGE_CONFIG,
      heartbeatInterval: 100, // Shorter for tests
      reconnectDelay: 50
    };

    // Setup global mocks
    (globalThis as any).goSandboxBridge = mockGoBridge;
    
    bridge = createCommunicationBridge(config);
    
    // Reset mocks
    jest.clearAllMocks();
  });

  afterEach(() => {
    bridge.destroy();
    delete (globalThis as any).handleGoMessage;
    delete (globalThis as any).handleGoBridgeStatus;
  });

  describe('Initialization', () => {
    test('should create bridge with config', () => {
      expect(bridge).toBeInstanceOf(CommunicationBridge);
      expect(bridge.getConfig()).toEqual(config);
    });

    test('should initialize successfully', async () => {
      // Mock successful Go response
      mockGoBridge.handleMessage.mockImplementation((messageJson: string) => {
        const message = JSON.parse(messageJson);
        setTimeout(() => {
          const response: SandboxMessage = {
            id: message.id,
            type: MessageType.RESPONSE,
            source: 'go_bridge',
            target: 'js_bridge',
            payload: { success: true },
            timestamp: Date.now(),
            response: true
          };
          (globalThis as any).handleGoMessage(JSON.stringify(response));
        }, 10);
      });

      await expect(bridge.initialize()).resolves.toBeUndefined();
      expect(bridge.isConnected()).toBe(true);
    });

    test('should handle initialization failure', async () => {
      delete (globalThis as any).goSandboxBridge;
      
      await expect(bridge.initialize())
        .rejects.toThrow('Go communication bridge not available');
    });

    test('should setup global handlers', async () => {
      await bridge.initialize();
      
      expect(typeof (globalThis as any).handleGoMessage).toBe('function');
      expect(typeof (globalThis as any).handleGoBridgeStatus).toBe('function');
    });
  });

  describe('Message Handling', () => {
    beforeEach(async () => {
      // Mock successful initialization
      mockGoBridge.handleMessage.mockImplementation((messageJson: string) => {
        const message = JSON.parse(messageJson);
        setTimeout(() => {
          const response: SandboxMessage = {
            id: message.id,
            type: MessageType.RESPONSE,
            source: 'go_bridge',
            target: 'js_bridge',
            payload: { success: true },
            timestamp: Date.now(),
            response: true
          };
          (globalThis as any).handleGoMessage(JSON.stringify(response));
        }, 10);
      });

      await bridge.initialize();
    });

    test('should send messages successfully', async () => {
      const testMessage: SandboxMessage = {
        id: 'test-123',
        type: MessageType.DATA,
        source: 'js_bridge',
        target: 'go_bridge',
        payload: { data: 'test' },
        timestamp: Date.now()
      };

      await expect(bridge.sendMessage(testMessage)).resolves.toBeUndefined();
      expect(mockGoBridge.handleMessage).toHaveBeenCalled();
    });

    test('should validate messages before sending', async () => {
      const invalidMessage = {
        // Missing required fields
        type: MessageType.DATA,
        payload: {}
      } as SandboxMessage;

      await expect(bridge.sendMessage(invalidMessage))
        .rejects.toThrow('Invalid message format');
    });

    test('should enforce message size limits', async () => {
      const largePayload = 'x'.repeat(2 * 1024 * 1024); // 2MB
      const largeMessage: SandboxMessage = {
        id: 'large-123',
        type: MessageType.DATA,
        source: 'js_bridge',
        target: 'go_bridge',
        payload: { data: largePayload },
        timestamp: Date.now()
      };

      await expect(bridge.sendMessage(largeMessage))
        .rejects.toThrow('Message size');
    });

    test('should handle incoming messages', () => {
      const messageHandler = jest.fn();
      bridge.onMessage(messageHandler);

      const incomingMessage: SandboxMessage = {
        id: 'incoming-123',
        type: MessageType.EVENT,
        source: 'go_bridge',
        target: 'js_bridge',
        payload: { eventType: 'test' },
        timestamp: Date.now()
      };

      (globalThis as any).handleGoMessage(JSON.stringify(incomingMessage));
      
      expect(messageHandler).toHaveBeenCalledWith(incomingMessage);
    });

    test('should handle heartbeat messages', () => {
      const heartbeatMessage: SandboxMessage = {
        id: 'heartbeat-123',
        type: MessageType.HEARTBEAT,
        source: 'go_bridge',
        target: 'js_bridge',
        payload: { timestamp: Date.now() },
        timestamp: Date.now()
      };

      const messageHandler = jest.fn();
      bridge.onMessage(messageHandler);

      (globalThis as any).handleGoMessage(JSON.stringify(heartbeatMessage));
      
      // Heartbeat messages should not trigger regular handlers
      expect(messageHandler).not.toHaveBeenCalled();
      
      const stats = bridge.getStats();
      expect(stats.lastHeartbeat).toBeGreaterThan(0);
    });
  });

  describe('Event Handlers', () => {
    test('should register and unregister message handlers', () => {
      const handler1: MessageHandler = jest.fn();
      const handler2: MessageHandler = jest.fn();

      bridge.onMessage(handler1);
      bridge.onMessage(handler2);
      bridge.offMessage(handler1);

      const testMessage: SandboxMessage = {
        id: 'test-123',
        type: MessageType.EVENT,
        source: 'go_bridge',
        target: 'js_bridge',
        payload: {},
        timestamp: Date.now()
      };

      (globalThis as any).handleGoMessage(JSON.stringify(testMessage));

      expect(handler1).not.toHaveBeenCalled();
      expect(handler2).toHaveBeenCalledWith(testMessage);
    });

    test('should register and unregister error handlers', () => {
      const errorHandler: ErrorHandler = jest.fn();
      
      bridge.onError(errorHandler);
      
      // Trigger an error by sending invalid message
      const invalidMessage = {} as SandboxMessage;
      bridge.sendMessage(invalidMessage).catch(() => {});

      expect(errorHandler).toHaveBeenCalled();
      
      bridge.offError(errorHandler);
    });

    test('should register and unregister status handlers', () => {
      const statusHandler: StatusHandler = jest.fn();
      
      bridge.onStatusChange(statusHandler);
      
      (globalThis as any).handleGoBridgeStatus('connected');
      
      expect(statusHandler).toHaveBeenCalledWith(ConnectionStatus.CONNECTED);
      
      bridge.offStatusChange(statusHandler);
    });
  });

  describe('Connection Management', () => {
    test('should track connection status', async () => {
      expect(bridge.isConnected()).toBe(false);
      
      // Mock successful initialization
      mockGoBridge.handleMessage.mockImplementation((messageJson: string) => {
        const message = JSON.parse(messageJson);
        setTimeout(() => {
          const response: SandboxMessage = {
            id: message.id,
            type: MessageType.RESPONSE,
            source: 'go_bridge',
            target: 'js_bridge',
            payload: { success: true },
            timestamp: Date.now(),
            response: true
          };
          (globalThis as any).handleGoMessage(JSON.stringify(response));
        }, 10);
      });

      await bridge.initialize();
      expect(bridge.isConnected()).toBe(true);
    });

    test('should handle status changes from Go', () => {
      const statusHandler = jest.fn();
      bridge.onStatusChange(statusHandler);

      (globalThis as any).handleGoBridgeStatus('connected');
      expect(statusHandler).toHaveBeenCalledWith(ConnectionStatus.CONNECTED);

      (globalThis as any).handleGoBridgeStatus('disconnected');
      expect(statusHandler).toHaveBeenCalledWith(ConnectionStatus.DISCONNECTED);

      (globalThis as any).handleGoBridgeStatus('error');
      expect(statusHandler).toHaveBeenCalledWith(ConnectionStatus.ERROR);
    });

    test('should attempt reconnection', async () => {
      // Mock initial failure then success
      let callCount = 0;
      mockGoBridge.handleMessage.mockImplementation((messageJson: string) => {
        callCount++;
        const message = JSON.parse(messageJson);
        
        if (callCount === 1) {
          // First call fails
          setTimeout(() => {
            const response: SandboxMessage = {
              id: message.id,
              type: MessageType.ERROR,
              source: 'go_bridge',
              target: 'js_bridge',
              payload: { error: 'Connection failed' },
              timestamp: Date.now(),
              response: true
            };
            (globalThis as any).handleGoMessage(JSON.stringify(response));
          }, 10);
        } else {
          // Subsequent calls succeed
          setTimeout(() => {
            const response: SandboxMessage = {
              id: message.id,
              type: MessageType.RESPONSE,
              source: 'go_bridge',
              target: 'js_bridge',
              payload: { success: true },
              timestamp: Date.now(),
              response: true
            };
            (globalThis as any).handleGoMessage(JSON.stringify(response));
          }, 10);
        }
      });

      await expect(bridge.reconnect()).resolves.toBeUndefined();
      expect(callCount).toBeGreaterThan(1);
    });

    test('should fail after max reconnection attempts', async () => {
      const failingConfig = { ...config, reconnectAttempts: 2 };
      const failingBridge = createCommunicationBridge(failingConfig);

      // Mock all calls to fail
      mockGoBridge.handleMessage.mockImplementation((messageJson: string) => {
        const message = JSON.parse(messageJson);
        setTimeout(() => {
          const response: SandboxMessage = {
            id: message.id,
            type: MessageType.ERROR,
            source: 'go_bridge',
            target: 'js_bridge',
            payload: { error: 'Connection failed' },
            timestamp: Date.now(),
            response: true
          };
          (globalThis as any).handleGoMessage(JSON.stringify(response));
        }, 10);
      });

      await expect(failingBridge.reconnect())
        .rejects.toThrow('Failed to reconnect after 2 attempts');
      
      failingBridge.destroy();
    });
  });

  describe('Statistics', () => {
    beforeEach(async () => {
      mockGoBridge.handleMessage.mockImplementation((messageJson: string) => {
        const message = JSON.parse(messageJson);
        setTimeout(() => {
          const response: SandboxMessage = {
            id: message.id,
            type: MessageType.RESPONSE,
            source: 'go_bridge',
            target: 'js_bridge',
            payload: { success: true },
            timestamp: Date.now(),
            response: true
          };
          (globalThis as any).handleGoMessage(JSON.stringify(response));
        }, 10);
      });

      await bridge.initialize();
    });

    test('should track message statistics', async () => {
      const statsBefore = bridge.getStats();
      
      const testMessage: SandboxMessage = {
        id: 'test-123',
        type: MessageType.DATA,
        source: 'js_bridge',
        target: 'go_bridge',
        payload: { data: 'test' },
        timestamp: Date.now()
      };

      await bridge.sendMessage(testMessage);
      
      const statsAfter = bridge.getStats();
      expect(statsAfter.messagesSent).toBe(statsBefore.messagesSent + 1);
      expect(statsAfter.bytesTransferred).toBeGreaterThan(statsBefore.bytesTransferred);
    });

    test('should track incoming message statistics', () => {
      const statsBefore = bridge.getStats();
      
      const incomingMessage: SandboxMessage = {
        id: 'incoming-123',
        type: MessageType.EVENT,
        source: 'go_bridge',
        target: 'js_bridge',
        payload: { eventType: 'test' },
        timestamp: Date.now()
      };

      (globalThis as any).handleGoMessage(JSON.stringify(incomingMessage));
      
      const statsAfter = bridge.getStats();
      expect(statsAfter.messagesReceived).toBe(statsBefore.messagesReceived + 1);
    });

    test('should track error statistics', async () => {
      const statsBefore = bridge.getStats();
      
      const invalidMessage = {} as SandboxMessage;
      await bridge.sendMessage(invalidMessage).catch(() => {});
      
      const statsAfter = bridge.getStats();
      expect(statsAfter.errors).toBe(statsBefore.errors + 1);
    });
  });

  describe('Heartbeat', () => {
    beforeEach(async () => {
      mockGoBridge.handleMessage.mockImplementation((messageJson: string) => {
        const message = JSON.parse(messageJson);
        setTimeout(() => {
          const response: SandboxMessage = {
            id: message.id,
            type: MessageType.RESPONSE,
            source: 'go_bridge',
            target: 'js_bridge',
            payload: { success: true },
            timestamp: Date.now(),
            response: true
          };
          (globalThis as any).handleGoMessage(JSON.stringify(response));
        }, 10);
      });

      await bridge.initialize();
    });

    test('should send heartbeat messages', (done) => {
      // Wait for heartbeat interval
      setTimeout(() => {
        const calls = mockGoBridge.handleMessage.mock.calls;
        const heartbeatCall = calls.find(call => {
          const message = JSON.parse(call[0]);
          return message.type === MessageType.HEARTBEAT;
        });
        
        expect(heartbeatCall).toBeDefined();
        done();
      }, 150); // Wait longer than heartbeat interval
    });

    test('should include stats in heartbeat', (done) => {
      setTimeout(() => {
        const calls = mockGoBridge.handleMessage.mock.calls;
        const heartbeatCall = calls.find(call => {
          const message = JSON.parse(call[0]);
          return message.type === MessageType.HEARTBEAT;
        });
        
        if (heartbeatCall) {
          const message = JSON.parse(heartbeatCall[0]);
          expect(message.payload.stats).toBeDefined();
        }
        
        done();
      }, 150);
    });
  });

  describe('Cleanup and Destruction', () => {
    test('should destroy bridge cleanly', async () => {
      mockGoBridge.handleMessage.mockImplementation((messageJson: string) => {
        const message = JSON.parse(messageJson);
        setTimeout(() => {
          const response: SandboxMessage = {
            id: message.id,
            type: MessageType.RESPONSE,
            source: 'go_bridge',
            target: 'js_bridge',
            payload: { success: true },
            timestamp: Date.now(),
            response: true
          };
          (globalThis as any).handleGoMessage(JSON.stringify(response));
        }, 10);
      });

      await bridge.initialize();
      expect(bridge.isConnected()).toBe(true);
      
      bridge.destroy();
      expect(bridge.isConnected()).toBe(false);
    });

    test('should handle multiple destroy calls', () => {
      bridge.destroy();
      expect(() => bridge.destroy()).not.toThrow();
    });

    test('should prevent operations after destruction', async () => {
      bridge.destroy();
      
      await expect(bridge.initialize())
        .rejects.toThrow('Bridge has been destroyed');
    });

    test('should send destruction message when connected', async () => {
      mockGoBridge.handleMessage.mockImplementation((messageJson: string) => {
        const message = JSON.parse(messageJson);
        setTimeout(() => {
          const response: SandboxMessage = {
            id: message.id,
            type: MessageType.RESPONSE,
            source: 'go_bridge',
            target: 'js_bridge',
            payload: { success: true },
            timestamp: Date.now(),
            response: true
          };
          (globalThis as any).handleGoMessage(JSON.stringify(response));
        }, 10);
      });

      await bridge.initialize();
      bridge.destroy();
      
      // Should have sent destroy message
      const calls = mockGoBridge.handleMessage.mock.calls;
      const destroyCall = calls.find(call => {
        const message = JSON.parse(call[0]);
        return message.payload.command === 'destroy';
      });
      
      expect(destroyCall).toBeDefined();
    });
  });

  describe('Error Handling', () => {
    test('should handle malformed incoming messages', () => {
      const errorHandler = jest.fn();
      bridge.onError(errorHandler);

      // Send malformed JSON
      (globalThis as any).handleGoMessage('invalid json');
      
      expect(errorHandler).toHaveBeenCalled();
    });

    test('should handle message handler errors', () => {
      const faultyHandler: MessageHandler = jest.fn(() => {
        throw new Error('Handler error');
      });
      
      bridge.onMessage(faultyHandler);
      
      const testMessage: SandboxMessage = {
        id: 'test-123',
        type: MessageType.EVENT,
        source: 'go_bridge',
        target: 'js_bridge',
        payload: {},
        timestamp: Date.now()
      };

      expect(() => {
        (globalThis as any).handleGoMessage(JSON.stringify(testMessage));
      }).not.toThrow();
    });

    test('should handle postMessage fallback', async () => {
      delete (globalThis as any).goSandboxBridge;
      (globalThis as any).postMessage = jest.fn();

      const fallbackBridge = createCommunicationBridge(config);
      
      mockGoBridge.handleMessage.mockImplementation((messageJson: string) => {
        const message = JSON.parse(messageJson);
        setTimeout(() => {
          const response: SandboxMessage = {
            id: message.id,
            type: MessageType.RESPONSE,
            source: 'go_bridge',
            target: 'js_bridge',
            payload: { success: true },
            timestamp: Date.now(),
            response: true
          };
          (globalThis as any).handleGoMessage(JSON.stringify(response));
        }, 10);
      });

      await fallbackBridge.initialize();
      
      const testMessage: SandboxMessage = {
        id: 'test-123',
        type: MessageType.DATA,
        source: 'js_bridge',
        target: 'go_bridge',
        payload: { data: 'test' },
        timestamp: Date.now()
      };

      await fallbackBridge.sendMessage(testMessage);
      expect((globalThis as any).postMessage).toHaveBeenCalled();
      
      fallbackBridge.destroy();
      delete (globalThis as any).postMessage;
    });
  });
});

describe('Factory Function', () => {
  test('should create bridge with factory function', () => {
    const bridge = createCommunicationBridge();
    expect(bridge).toBeInstanceOf(CommunicationBridge);
    bridge.destroy();
  });

  test('should use custom config', () => {
    const customConfig = {
      ...DEFAULT_BRIDGE_CONFIG,
      maxMessageSize: 512 * 1024
    };
    
    const bridge = createCommunicationBridge(customConfig);
    expect(bridge.getConfig().maxMessageSize).toBe(512 * 1024);
    bridge.destroy();
  });
});

describe('Default Configuration', () => {
  test('should have reasonable defaults', () => {
    expect(DEFAULT_BRIDGE_CONFIG.maxMessageSize).toBe(1024 * 1024);
    expect(DEFAULT_BRIDGE_CONFIG.heartbeatInterval).toBe(30000);
    expect(DEFAULT_BRIDGE_CONFIG.reconnectAttempts).toBe(3);
    expect(DEFAULT_BRIDGE_CONFIG.enableEncryption).toBe(false);
    expect(DEFAULT_BRIDGE_CONFIG.enableCompression).toBe(false);
  });
});