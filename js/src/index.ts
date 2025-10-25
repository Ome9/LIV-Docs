/**
 * LIV Document Format - JavaScript/TypeScript SDK
 * Main entry point for the LIV document format JavaScript implementation
 */

// Export all types and interfaces
export * from './types';

// Export sandbox interface
export {
  SandboxInterface,
  createSandbox,
  DEFAULT_SANDBOX_CONFIG,
  MessageType
} from './sandbox-interface';

// Export DOM API
export {
  SecureDOMAPI,
  createSecureDOMAPI
} from './dom-api';

// Export communication bridge
export {
  CommunicationBridge,
  createCommunicationBridge,
  DEFAULT_BRIDGE_CONFIG,
  ConnectionStatus
} from './communication-bridge';

// Export existing components
export * from './loader';
export * from './renderer';
export * from './sandbox';
export * from './editor';
export * from './document';

// Export SDK
export * from './sdk';

// Export converters
export * from './html-markdown-converter';
export * from './epub-converter';

// Version information
export const VERSION = '0.1.0';
export const SUPPORTED_FORMAT_VERSION = '1.0';

/**
 * Initialize the LIV JavaScript SDK with default configuration
 */
export function initializeLIVSDK(config?: {
  enableSandbox?: boolean;
  enableSecureDOM?: boolean;
  enableCommunicationBridge?: boolean;
  sandboxConfig?: any;
  domSecurityPolicy?: any;
  bridgeConfig?: any;
}) {
  const defaultConfig = {
    enableSandbox: true,
    enableSecureDOM: true,
    enableCommunicationBridge: true,
    ...config
  };

  console.log(`[LIV SDK] Initializing version ${VERSION}`);
  console.log(`[LIV SDK] Supported format version: ${SUPPORTED_FORMAT_VERSION}`);
  console.log(`[LIV SDK] Configuration:`, defaultConfig);

  return {
    version: VERSION,
    supportedFormatVersion: SUPPORTED_FORMAT_VERSION,
    config: defaultConfig,
    initialized: true
  };
}