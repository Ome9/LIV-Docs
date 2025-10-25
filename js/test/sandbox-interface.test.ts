/**
 * Tests for LIV Document Format - JavaScript Sandbox Interface
 */

import {
  SandboxInterface,
  createSandbox,
  DEFAULT_SANDBOX_CONFIG,
  MessageType,
  SandboxConfig,
  WASMModuleConfig,
  SecurityPolicy
} from '../src/sandbox-interface';

// Mock global Go bridge
(globalThis as any).goSandboxBridge = {
  handleMessage: jest.fn((messageJson: string) => {
    const message = JSON.parse(messageJson);
    
    // Simulate Go backend responses
    setTimeout(() => {
      const response = {
        id: message.id,
        type: MessageType.RESPONSE,
        source: 'go_runtime',
        target: 'js_sandbox',
        payload: {
          success: true,
          result: message.type === MessageType.FUNCTION_CALL ? 
            { executed: message.payload.functionName } : 
            { acknowledged: true }
        },
        timestamp: Date.now(),
        response: true
      };
      
      if ((globalThis as any).handleGoMessage) {
        (globalThis as any).handleGoMessage(JSON.stringify(response));
      }
    }, 10);
  })
};

describe('SandboxInterface', () => {
  let sandbox: SandboxInterface;
  let config: SandboxConfig;

  beforeEach(() => {
    config = {
      ...DEFAULT_SANDBOX_CONFIG,
      timeoutMs: 1000 // Shorter timeout for tests
    };
    sandbox = createSandbox(config);
  });

  afterEach(async () => {
    if (sandbox) {
      await sandbox.destroy();
    }
  });

  describe('Initialization', () => {
    test('should create sandbox with default config', () => {
      expect(sandbox).toBeInstanceOf(SandboxInterface);
      expect(sandbox.getStats().sessionId).toBeDefined();
    });

    test('should initialize successfully', async () => {
      await expect(sandbox.initialize()).resolves.toBeUndefined();
      expect(sandbox.getStats().sessionId).toBeDefined();
    });

    test('should throw error when initializing twice', async () => {
      await sandbox.initialize();
      await expect(sandbox.initialize()).rejects.toThrow('Sandbox already initialized');
    });

    test('should throw error when initializing destroyed sandbox', async () => {
      await sandbox.destroy();
      await expect(sandbox.initialize()).rejects.toThrow('Cannot initialize destroyed sandbox');
    });
  });

  describe('WASM Module Management', () => {
    let moduleConfig: WASMModuleConfig;
    let moduleData: ArrayBuffer;

    beforeEach(async () => {
      await sandbox.initialize();
      
      moduleConfig = {
        name: 'test-module',
        version: '1.0.0',
        entryPoint: 'main',
        exports: ['add', 'multiply'],
        imports: ['env'],
        permissions: {
          memoryLimit: 1024 * 1024, // 1MB
          cpuTimeLimit: 1000,
          allowNetworking: false,
          allowFileSystem: false,
          allowedImports: ['env']
        }
      };

      // Create mock WASM module data
      moduleData = new ArrayBuffer(100);
    });

    test('should load WASM module successfully', async () => {
      await expect(sandbox.loadWASMModule(moduleData, moduleConfig)).resolves.toBeUndefined();
      expect(sandbox.getLoadedModules()).toContain('test-module');
    });

    test('should validate module configuration', async () => {
      const invalidConfig = { ...moduleConfig, name: '' };
      await expect(sandbox.loadWASMModule(moduleData, invalidConfig))
        .rejects.toThrow('WASM module name is required');
    });

    test('should enforce memory limits', async () => {
      const oversizedConfig = {
        ...moduleConfig,
        permissions: {
          ...moduleConfig.permissions,
          memoryLimit: 100 * 1024 * 1024 // 100MB, exceeds sandbox limit
        }
      };
      
      await expect(sandbox.loadWASMModule(moduleData, oversizedConfig))
        .rejects.toThrow('WASM module memory limit exceeds sandbox limit');
    });

    test('should get module configuration', async () => {
      await sandbox.loadWASMModule(moduleData, moduleConfig);
      const retrievedConfig = sandbox.getModuleConfig('test-module');
      expect(retrievedConfig).toEqual(moduleConfig);
    });
  });

  describe('Function Execution', () => {
    beforeEach(async () => {
      await sandbox.initialize();
      
      const moduleConfig: WASMModuleConfig = {
        name: 'math-module',
        version: '1.0.0',
        entryPoint: 'main',
        exports: ['add', 'multiply'],
        imports: ['env'],
        permissions: {
          memoryLimit: 1024 * 1024,
          cpuTimeLimit: 1000,
          allowNetworking: false,
          allowFileSystem: false,
          allowedImports: ['env']
        }
      };

      await sandbox.loadWASMModule(new ArrayBuffer(100), moduleConfig);
    });

    test('should execute WASM function successfully', async () => {
      const result = await sandbox.executeWASMFunction('math-module', 'add', [2, 3]);
      
      expect(result.success).toBe(true);
      expect(result.result).toBeDefined();
      expect(result.duration).toBeGreaterThan(0);
    });

    test('should handle non-existent module', async () => {
      const result = await sandbox.executeWASMFunction('non-existent', 'add', [2, 3]);
      
      expect(result.success).toBe(false);
      expect(result.error).toContain('not loaded');
    });

    test('should handle non-exported function', async () => {
      const result = await sandbox.executeWASMFunction('math-module', 'divide', [6, 2]);
      
      expect(result.success).toBe(false);
      expect(result.error).toContain('not exported');
    });

    test('should update statistics on execution', async () => {
      const statsBefore = sandbox.getStats();
      await sandbox.executeWASMFunction('math-module', 'add', [2, 3]);
      const statsAfter = sandbox.getStats();
      
      expect(statsAfter.functionCalls).toBe(statsBefore.functionCalls + 1);
      expect(statsAfter.cpuTime).toBeGreaterThan(statsBefore.cpuTime);
    });
  });

  describe('Event Handling', () => {
    beforeEach(async () => {
      await sandbox.initialize();
    });

    test('should send events successfully', async () => {
      await expect(sandbox.sendEvent('user_action', { action: 'click', x: 100, y: 200 }))
        .resolves.toBeUndefined();
    });

    test('should handle message types', () => {
      const messageHandler = jest.fn();
      sandbox.onMessage(MessageType.EVENT, messageHandler);
      
      const testMessage = {
        id: 'test-123',
        type: MessageType.EVENT,
        source: 'go_runtime',
        target: 'js_sandbox',
        payload: { eventType: 'test_event' },
        timestamp: Date.now()
      };
      
      sandbox.handleIncomingMessage(testMessage);
      expect(messageHandler).toHaveBeenCalledWith(testMessage);
    });
  });

  describe('Statistics and Monitoring', () => {
    beforeEach(async () => {
      await sandbox.initialize();
    });

    test('should track basic statistics', () => {
      const stats = sandbox.getStats();
      
      expect(stats.sessionId).toBeDefined();
      expect(stats.uptime).toBeGreaterThanOrEqual(0);
      expect(stats.functionCalls).toBe(0);
      expect(stats.errors).toBe(0);
    });

    test('should update error count on failures', async () => {
      const statsBefore = sandbox.getStats();
      
      // Trigger an error by calling non-existent function
      await sandbox.executeWASMFunction('non-existent', 'test', []);
      
      const statsAfter = sandbox.getStats();
      expect(statsAfter.errors).toBe(statsBefore.errors + 1);
    });
  });

  describe('Security and Permissions', () => {
    test('should enforce security policy', () => {
      const restrictiveConfig: SandboxConfig = {
        ...DEFAULT_SANDBOX_CONFIG,
        securityPolicy: {
          ...DEFAULT_SANDBOX_CONFIG.securityPolicy,
          wasmPermissions: {
            memoryLimit: 512 * 1024, // 512KB
            cpuTimeLimit: 500,
            allowNetworking: false,
            allowFileSystem: false,
            allowedImports: []
          }
        }
      };

      const restrictiveSandbox = createSandbox(restrictiveConfig);
      expect(restrictiveSandbox).toBeInstanceOf(SandboxInterface);
    });

    test('should validate permissions against policy', async () => {
      await sandbox.initialize();
      
      const moduleConfig: WASMModuleConfig = {
        name: 'network-module',
        version: '1.0.0',
        entryPoint: 'main',
        exports: ['fetch'],
        imports: ['env'],
        permissions: {
          memoryLimit: 1024 * 1024,
          cpuTimeLimit: 1000,
          allowNetworking: true, // This should be denied by default policy
          allowFileSystem: false,
          allowedImports: ['env']
        }
      };

      await expect(sandbox.loadWASMModule(new ArrayBuffer(100), moduleConfig))
        .rejects.toThrow('requests networking but sandbox policy denies it');
    });
  });

  describe('Cleanup and Destruction', () => {
    test('should destroy sandbox cleanly', async () => {
      await sandbox.initialize();
      await expect(sandbox.destroy()).resolves.toBeUndefined();
      
      // Should not be able to use sandbox after destruction
      await expect(sandbox.executeWASMFunction('test', 'func', []))
        .rejects.toThrow('Sandbox has been destroyed');
    });

    test('should handle multiple destroy calls', async () => {
      await sandbox.initialize();
      await sandbox.destroy();
      await expect(sandbox.destroy()).resolves.toBeUndefined();
    });

    test('should clean up resources on destroy', async () => {
      await sandbox.initialize();
      
      const moduleConfig: WASMModuleConfig = {
        name: 'cleanup-test',
        version: '1.0.0',
        entryPoint: 'main',
        exports: ['test'],
        imports: ['env'],
        permissions: {
          memoryLimit: 1024 * 1024,
          cpuTimeLimit: 1000,
          allowNetworking: false,
          allowFileSystem: false,
          allowedImports: ['env']
        }
      };

      await sandbox.loadWASMModule(new ArrayBuffer(100), moduleConfig);
      expect(sandbox.getLoadedModules()).toContain('cleanup-test');
      
      await sandbox.destroy();
      
      // Resources should be cleaned up
      expect(() => sandbox.getLoadedModules()).toThrow('Sandbox has been destroyed');
    });
  });

  describe('Error Handling', () => {
    test('should handle initialization errors gracefully', async () => {
      // Remove Go bridge to simulate error
      delete (globalThis as any).goSandboxBridge;
      
      await expect(sandbox.initialize())
        .rejects.toThrow('No communication bridge available');
      
      // Restore bridge for cleanup
      (globalThis as any).goSandboxBridge = { handleMessage: jest.fn() };
    });

    test('should handle message timeout', async () => {
      const timeoutConfig = { ...config, timeoutMs: 10 };
      const timeoutSandbox = createSandbox(timeoutConfig);
      
      // Mock Go bridge to not respond
      (globalThis as any).goSandboxBridge.handleMessage = jest.fn();
      
      await expect(timeoutSandbox.initialize())
        .rejects.toThrow('Message timeout');
      
      await timeoutSandbox.destroy();
    });
  });
});

describe('Factory Functions', () => {
  test('should create sandbox with factory function', () => {
    const sandbox = createSandbox(DEFAULT_SANDBOX_CONFIG);
    expect(sandbox).toBeInstanceOf(SandboxInterface);
  });

  test('should use default config when none provided', () => {
    const sandbox = createSandbox(DEFAULT_SANDBOX_CONFIG);
    const stats = sandbox.getStats();
    expect(stats.sessionId).toBeDefined();
  });
});

describe('Default Configuration', () => {
  test('should have secure default configuration', () => {
    expect(DEFAULT_SANDBOX_CONFIG.securityPolicy.wasmPermissions.allowNetworking).toBe(false);
    expect(DEFAULT_SANDBOX_CONFIG.securityPolicy.wasmPermissions.allowFileSystem).toBe(false);
    expect(DEFAULT_SANDBOX_CONFIG.securityPolicy.jsPermissions.executionMode).toBe('sandboxed');
    expect(DEFAULT_SANDBOX_CONFIG.securityPolicy.storagePolicy.allowLocalStorage).toBe(false);
  });

  test('should have reasonable resource limits', () => {
    expect(DEFAULT_SANDBOX_CONFIG.securityPolicy.wasmPermissions.memoryLimit).toBe(16 * 1024 * 1024);
    expect(DEFAULT_SANDBOX_CONFIG.securityPolicy.wasmPermissions.cpuTimeLimit).toBe(5000);
    expect(DEFAULT_SANDBOX_CONFIG.maxMemoryMB).toBe(64);
    expect(DEFAULT_SANDBOX_CONFIG.timeoutMs).toBe(30000);
  });
});