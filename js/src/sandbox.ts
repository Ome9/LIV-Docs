// Sandbox implementation for secure content execution

import { WASMPermissions } from './sandbox-interface';
import { WASMPermissions } from './sandbox-interface';
import { WASMPermissions } from './sandbox-interface';
import { SecurityPolicy } from './sandbox-interface';
import { WASMPermissions } from './sandbox-interface';
import { SecurityPolicy } from './sandbox-interface';
import { SecurityPolicy } from './sandbox-interface';
import { WASMPermissions } from './sandbox-interface';
import { SecurityPolicy } from './sandbox-interface';
import { SecurityPolicy } from './sandbox-interface';
import { LegacySecurityPolicy, LegacyWASMPermissions, LegacyJSPermissions } from './types';

export class ContentSandbox {
  private iframe: HTMLIFrameElement;
  private permissions: SecurityPolicy;
  private messageHandlers: Map<string, Function> = new Map();

  constructor(permissions: SecurityPolicy) {
    this.permissions = permissions;
    this.iframe = this.createSandboxedIframe();
    this.setupMessageHandling();
  }

  private createSandboxedIframe(): HTMLIFrameElement {
    const iframe = document.createElement('iframe');
    
    // Apply strict sandbox attributes
    const sandboxFlags = [
      'allow-scripts',
      'allow-same-origin'
    ];
    
    // Conditionally add permissions based on policy
    if (this.permissions.jsPermissions.executionMode === 'trusted') {
      sandboxFlags.push('allow-forms', 'allow-popups');
    }
    
    iframe.sandbox.add(...sandboxFlags);
    iframe.style.display = 'none';
    
    // Set CSP via srcdoc
    const csp = this.permissions.contentSecurityPolicy || 
      "default-src 'self'; script-src 'self' 'unsafe-inline'; style-src 'self' 'unsafe-inline';";
    
    iframe.srcdoc = `
      <!DOCTYPE html>
      <html>
      <head>
        <meta http-equiv="Content-Security-Policy" content="${csp}">
        <title>LIV Sandbox</title>
      </head>
      <body>
        <div id="sandbox-root"></div>
        <script>
          // Sandbox communication setup
          window.addEventListener('message', function(event) {
            if (event.source !== window.parent) return;
            
            try {
              const message = JSON.parse(event.data);
              handleSandboxMessage(message);
            } catch (error) {
              console.error('Sandbox message error:', error);
            }
          });
          
          function handleSandboxMessage(message) {
            switch (message.type) {
              case 'execute':
                executeCode(message.code, message.permissions);
                break;
              case 'loadWASM':
                loadWASMModule(message.module, message.config);
                break;
              case 'updateDOM':
                updateDOM(message.operations);
                break;
            }
          }
          
          function executeCode(code, permissions) {
            try {
              // Create restricted execution context
              const context = createRestrictedContext(permissions);
              const result = executeInContext(code, context);
              
              window.parent.postMessage(JSON.stringify({
                type: 'execution-result',
                success: true,
                result: result
              }), '*');
            } catch (error) {
              window.parent.postMessage(JSON.stringify({
                type: 'execution-result',
                success: false,
                error: error.message
              }), '*');
            }
          }
          
          function createRestrictedContext(permissions) {
            const context = {
              console: {
                log: (...args) => window.parent.postMessage(JSON.stringify({
                  type: 'console-log',
                  args: args
                }), '*')
              }
            };
            
            // Add DOM access based on permissions
            if (permissions.domAccess === 'read' || permissions.domAccess === 'write') {
              context.document = createRestrictedDocument(permissions.domAccess === 'write');
            }
            
            return context;
          }
          
          function createRestrictedDocument(allowWrite) {
            const root = document.getElementById('sandbox-root');
            
            return {
              getElementById: (id) => root.querySelector('#' + id),
              createElement: allowWrite ? (tag) => {
                if (isAllowedElement(tag)) {
                  return document.createElement(tag);
                }
                throw new Error('Element type not allowed: ' + tag);
              } : undefined,
              querySelector: (selector) => root.querySelector(selector),
              querySelectorAll: (selector) => root.querySelectorAll(selector)
            };
          }
          
          function isAllowedElement(tag) {
            const allowed = ['div', 'span', 'p', 'canvas', 'svg'];
            return allowed.includes(tag.toLowerCase());
          }
          
          function executeInContext(code, context) {
            // Create function with restricted context
            const func = new Function(...Object.keys(context), code);
            return func(...Object.values(context));
          }
          
          async function loadWASMModule(moduleData, config) {
            try {
              const wasmModule = await WebAssembly.instantiate(moduleData);
              
              window.parent.postMessage(JSON.stringify({
                type: 'wasm-loaded',
                success: true,
                exports: Object.keys(wasmModule.instance.exports)
              }), '*');
            } catch (error) {
              window.parent.postMessage(JSON.stringify({
                type: 'wasm-loaded',
                success: false,
                error: error.message
              }), '*');
            }
          }
          
          function updateDOM(operations) {
            const root = document.getElementById('sandbox-root');
            
            operations.forEach(op => {
              switch (op.type) {
                case 'create':
                  if (isAllowedElement(op.tag)) {
                    const element = document.createElement(op.tag);
                    element.id = op.id;
                    
                    const parent = op.parentId ? 
                      root.querySelector('#' + op.parentId) : root;
                    if (parent) {
                      parent.appendChild(element);
                    }
                  }
                  break;
                  
                case 'update':
                  const element = root.querySelector('#' + op.id);
                  if (element && op.attributes) {
                    Object.entries(op.attributes).forEach(([key, value]) => {
                      if (isAllowedAttribute(key)) {
                        element.setAttribute(key, value);
                      }
                    });
                  }
                  break;
                  
                case 'remove':
                  const toRemove = root.querySelector('#' + op.id);
                  if (toRemove && toRemove.parentNode) {
                    toRemove.parentNode.removeChild(toRemove);
                  }
                  break;
              }
            });
          }
          
          function isAllowedAttribute(attr) {
            const dangerous = ['onclick', 'onload', 'onerror'];
            return !dangerous.includes(attr.toLowerCase());
          }
        </script>
      </body>
      </html>
    `;
    
    document.body.appendChild(iframe);
    return iframe;
  }

  private setupMessageHandling(): void {
    window.addEventListener('message', (event) => {
      if (event.source !== this.iframe.contentWindow) return;
      
      try {
        const message = JSON.parse(event.data);
        const handler = this.messageHandlers.get(message.type);
        if (handler) {
          handler(message);
        }
      } catch (error) {
        console.error('Sandbox message handling error:', error);
      }
    });
  }

  async executeScript(code: string, permissions: WASMPermissions): Promise<any> {
    return new Promise((resolve, reject) => {
      const timeout = setTimeout(() => {
        reject(new Error('Script execution timeout'));
      }, permissions.cpuTimeLimit || 5000);

      this.messageHandlers.set('execution-result', (message) => {
        clearTimeout(timeout);
        this.messageHandlers.delete('execution-result');
        
        if (message.success) {
          resolve(message.result);
        } else {
          reject(new Error(message.error));
        }
      });

      this.postMessage({
        type: 'execute',
        code: code,
        permissions: this.permissions.jsPermissions
      });
    });
  }

  async loadWASM(moduleData: ArrayBuffer, config: any): Promise<any> {
    return new Promise((resolve, reject) => {
      const timeout = setTimeout(() => {
        reject(new Error('WASM loading timeout'));
      }, 10000);

      this.messageHandlers.set('wasm-loaded', (message) => {
        clearTimeout(timeout);
        this.messageHandlers.delete('wasm-loaded');
        
        if (message.success) {
          resolve(message.exports);
        } else {
          reject(new Error(message.error));
        }
      });

      this.postMessage({
        type: 'loadWASM',
        module: moduleData,
        config: config
      });
    });
  }

  updateDOM(operations: any[]): void {
    this.postMessage({
      type: 'updateDOM',
      operations: operations
    });
  }

  private postMessage(message: any): void {
    if (this.iframe.contentWindow) {
      this.iframe.contentWindow.postMessage(JSON.stringify(message), '*');
    }
  }

  getPermissions(): SecurityPolicy {
    return this.permissions;
  }

  updatePermissions(policy: SecurityPolicy): void {
    this.permissions = policy;
    // Note: Would need to recreate iframe for permission changes
  }

  destroy(): void {
    this.messageHandlers.clear();
    if (this.iframe.parentNode) {
      this.iframe.parentNode.removeChild(this.iframe);
    }
  }
}

export class PermissionManager {
  static validatePermissions(requested: WASMPermissions, policy: SecurityPolicy): boolean {
    const allowed = policy.wasmPermissions;
    
    // Check memory limit
    if (requested.memoryLimit > allowed.memoryLimit) {
      return false;
    }
    
    // Check CPU time limit
    if (requested.cpuTimeLimit > allowed.cpuTimeLimit) {
      return false;
    }
    
    // Check networking permission
    if (requested.allowNetworking && !allowed.allowNetworking) {
      return false;
    }
    
    // Check file system permission
    if (requested.allowFileSystem && !allowed.allowFileSystem) {
      return false;
    }
    
    // Check allowed imports
    for (const importName of requested.allowedImports) {
      if (!allowed.allowedImports.includes(importName)) {
        return false;
      }
    }
    
    return true;
  }

  static createRestrictedPermissions(base: WASMPermissions, restrictions: Partial<WASMPermissions>): WASMPermissions {
    return {
      memoryLimit: Math.min(base.memoryLimit, restrictions.memoryLimit || base.memoryLimit),
      cpuTimeLimit: Math.min(base.cpuTimeLimit, restrictions.cpuTimeLimit || base.cpuTimeLimit),
      allowNetworking: base.allowNetworking && (restrictions.allowNetworking !== false),
      allowFileSystem: base.allowFileSystem && (restrictions.allowFileSystem !== false),
      allowedImports: base.allowedImports.filter(imp => 
        !restrictions.allowedImports || restrictions.allowedImports.includes(imp)
      )
    };
  }
}