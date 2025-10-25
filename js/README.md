# LIV Document Format - JavaScript/TypeScript SDK

A secure, sandboxed JavaScript/TypeScript implementation for the LIV (Living Interactive Viewer) document format.

## Features

- **Secure Sandbox Interface**: Isolated execution environment for WASM modules with configurable permissions
- **Secure DOM API**: Controlled DOM access with security policies and sanitization
- **Communication Bridge**: Secure messaging between JavaScript and Go backend
- **TypeScript Support**: Full type definitions and IntelliSense support
- **Comprehensive Testing**: Unit tests with high coverage
- **Modern Architecture**: ES2020+ with modular design

## Installation

```bash
npm install liv-document-format-js
```

## Quick Start

```typescript
import { 
  initializeLIVSDK, 
  createSandbox, 
  createSecureDOMAPI, 
  createCommunicationBridge 
} from 'liv-document-format-js';

// Initialize the SDK
const sdk = initializeLIVSDK({
  enableSandbox: true,
  enableSecureDOM: true,
  enableCommunicationBridge: true
});

// Create a sandbox for WASM execution
const sandbox = createSandbox({
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
  timeoutMs: 30000
});

// Initialize and use the sandbox
async function useSandbox() {
  await sandbox.initialize();
  
  // Load a WASM module
  const moduleData = new ArrayBuffer(/* your WASM module data */);
  const moduleConfig = {
    name: 'my-module',
    version: '1.0.0',
    entryPoint: 'main',
    exports: ['calculate', 'render'],
    imports: ['env'],
    permissions: {
      memoryLimit: 8 * 1024 * 1024, // 8MB
      cpuTimeLimit: 2000, // 2 seconds
      allowNetworking: false,
      allowFileSystem: false,
      allowedImports: ['env']
    }
  };
  
  await sandbox.loadWASMModule(moduleData, moduleConfig);
  
  // Execute a function
  const result = await sandbox.executeWASMFunction('my-module', 'calculate', [10, 20]);
  console.log('Result:', result);
  
  // Clean up
  await sandbox.destroy();
}
```

## Secure DOM API

```typescript
import { createSecureDOMAPI } from 'liv-document-format-js';

const domAPI = createSecureDOMAPI({
  executionMode: 'sandboxed',
  allowedAPIs: ['console'],
  domAccess: 'write'
}, {
  allowedElements: ['div', 'span', 'p'],
  allowedAttributes: ['id', 'class', 'style'],
  allowedEvents: ['click', 'mouseover'],
  allowedStyles: ['color', 'background-color'],
  maxElements: 100,
  allowScriptExecution: false,
  allowFormSubmission: false,
  allowNavigation: false
});

// Create elements securely
const element = domAPI.createElement({
  tagName: 'div',
  attributes: { id: 'my-element', class: 'container' },
  textContent: 'Hello, secure world!'
});

// Add event listeners
domAPI.addEventListener({
  element: element!,
  eventType: 'click',
  handler: (event) => console.log('Clicked!'),
});

// Clean up when done
domAPI.cleanup();
```

## Communication Bridge

```typescript
import { createCommunicationBridge } from 'liv-document-format-js';

const bridge = createCommunicationBridge({
  enableEncryption: false,
  enableCompression: false,
  maxMessageSize: 1024 * 1024, // 1MB
  heartbeatInterval: 30000, // 30 seconds
  reconnectAttempts: 3,
  reconnectDelay: 1000 // 1 second
});

// Initialize the bridge
await bridge.initialize();

// Send messages
await bridge.sendMessage({
  id: 'msg-123',
  type: 'DATA',
  source: 'js_client',
  target: 'go_backend',
  payload: { action: 'process', data: [1, 2, 3] },
  timestamp: Date.now()
});

// Handle incoming messages
bridge.onMessage((message) => {
  console.log('Received message:', message);
});

// Handle errors
bridge.onError((error) => {
  console.error('Bridge error:', error);
});

// Clean up
bridge.destroy();
```

## API Reference

### SandboxInterface

The main interface for secure WASM execution:

- `initialize()`: Initialize the sandbox
- `loadWASMModule(data, config)`: Load a WASM module
- `executeWASMFunction(module, function, args)`: Execute a WASM function
- `sendEvent(type, data)`: Send events to the backend
- `getStats()`: Get sandbox statistics
- `destroy()`: Clean up resources

### SecureDOMAPI

Secure DOM manipulation interface:

- `createElement(options)`: Create DOM elements
- `querySelector(selector)`: Query elements
- `setAttribute(element, name, value)`: Set attributes
- `addEventListener(options)`: Add event listeners
- `setStyle(element, property, value)`: Set styles
- `cleanup()`: Clean up all created elements

### CommunicationBridge

Secure communication with Go backend:

- `initialize()`: Initialize the bridge
- `sendMessage(message)`: Send messages
- `onMessage(handler)`: Register message handler
- `onError(handler)`: Register error handler
- `reconnect()`: Reconnect to backend
- `destroy()`: Clean up resources

## Security Features

- **Sandboxed Execution**: WASM modules run in isolated environments
- **Permission System**: Granular control over resource access
- **DOM Sanitization**: All DOM operations are sanitized
- **Message Validation**: All communications are validated
- **Resource Limits**: Memory and CPU time limits enforced
- **Content Security Policy**: CSP integration for additional security

## Development

```bash
# Install dependencies
npm install

# Build the project
npm run build

# Run tests
npm test

# Run tests with coverage
npm run test:coverage

# Watch mode for development
npm run build:watch
npm run test:watch

# Lint code
npm run lint
npm run lint:fix
```

## Testing

The project includes comprehensive unit tests covering:

- Sandbox interface functionality
- DOM API security features
- Communication bridge reliability
- Error handling and edge cases
- Security policy enforcement

Run tests with:

```bash
npm test
```

## Browser Compatibility

- Chrome 80+
- Firefox 75+
- Safari 13+
- Edge 80+

## Node.js Compatibility

- Node.js 16+

## License

MIT License - see LICENSE file for details.

## Contributing

1. Fork the repository
2. Create a feature branch
3. Add tests for new functionality
4. Ensure all tests pass
5. Submit a pull request

## Security

If you discover a security vulnerability, please email security@liv-format.org instead of using the issue tracker.