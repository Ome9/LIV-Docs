// TypeScript type definitions for LIV Format

import { CommunicationBridge } from '.';

import { SecureDOMAPI } from '.';

import { SandboxInterface } from '.';

import { SandboxConfig } from './sandbox-interface';

// Core document types
export interface LIVDocument {
  manifest: Manifest;
  content: DocumentContent;
  assets: AssetBundle;
  signatures: SignatureBundle;
  wasmModules: Map<string, ArrayBuffer>;
}

export interface Manifest {
  version: string;
  metadata: DocumentMetadata;
  security: LegacySecurityPolicy;
  resources: Record<string, Resource>;
  wasmConfig?: WASMConfiguration;
  features?: FeatureFlags;
}

export interface DocumentMetadata {
  title: string;
  author: string;
  created: string; // ISO 8601
  modified: string; // ISO 8601
  description: string;
  version: string;
  language: string;
}

export interface DocumentContent {
  html: string;
  css: string;
  interactiveSpec: string;
  staticFallback: string;
}

export interface AssetBundle {
  images: Map<string, ArrayBuffer>;
  fonts: Map<string, ArrayBuffer>;
  data: Map<string, ArrayBuffer>;
}

export interface SignatureBundle {
  contentSignature: string;
  manifestSignature: string;
  wasmSignatures: Record<string, string>;
}

// Legacy Security types (kept for compatibility)
export interface LegacySecurityPolicy {
  wasmPermissions: LegacyWASMPermissions;
  jsPermissions: LegacyJSPermissions;
  networkPolicy: LegacyNetworkPolicy;
  storagePolicy: LegacyStoragePolicy;
  contentSecurityPolicy?: string;
  trustedDomains?: string[];
}

export interface LegacyWASMPermissions {
  memoryLimit: number;
  allowedImports: string[];
  cpuTimeLimit: number;
  allowNetworking: boolean;
  allowFileSystem: boolean;
}

export interface LegacyJSPermissions {
  executionMode: 'none' | 'sandboxed' | 'trusted';
  allowedAPIs: string[];
  domAccess: 'none' | 'read' | 'write';
}

export interface LegacyNetworkPolicy {
  allowOutbound: boolean;
  allowedHosts: string[];
  allowedPorts: number[];
}

export interface LegacyStoragePolicy {
  allowLocalStorage: boolean;
  allowSessionStorage: boolean;
  allowIndexedDB: boolean;
  allowCookies: boolean;
}

export interface Resource {
  hash: string;
  size: number;
  type: string;
  path: string;
}

export interface WASMConfiguration {
  modules: Record<string, WASMModule>;
  permissions: LegacyWASMPermissions;
  memoryLimit: number;
}

export interface WASMModule {
  name: string;
  version: string;
  entryPoint: string;
  exports: string[];
  imports: string[];
  permissions?: LegacyWASMPermissions;
  metadata?: Record<string, string>;
}

export interface FeatureFlags {
  animations: boolean;
  interactivity: boolean;
  charts: boolean;
  forms: boolean;
  audio: boolean;
  video: boolean;
  webgl: boolean;
  webassembly: boolean;
}

// Validation types
export interface ValidationResult {
  isValid: boolean;
  errors: string[];
  warnings: string[];
}

export interface SecurityReport {
  isValid: boolean;
  signatureVerified: boolean;
  integrityChecked: boolean;
  permissionsValid: boolean;
  warnings: string[];
  errors: string[];
}

// Interactive content types (from WASM)
export interface InteractiveElement {
  id: string;
  elementType: ElementType;
  properties: Record<string, any>;
  children: string[];
  eventHandlers: EventHandler[];
  transform: Transform;
  style: ElementStyle;
}

export type ElementType = 'Chart' | 'Animation' | 'Interactive' | 'Vector' | 'Text' | 'Image' | 'Container';

export interface EventHandler {
  eventType: string;
  handlerId: string;
  parameters: Record<string, any>;
}

export interface Transform {
  x: number;
  y: number;
  scaleX: number;
  scaleY: number;
  rotation: number;
  opacity: number;
}

export interface ElementStyle {
  backgroundColor?: string;
  borderColor?: string;
  borderWidth?: number;
  borderRadius?: number;
  shadow?: Shadow;
}

export interface Shadow {
  offsetX: number;
  offsetY: number;
  blurRadius: number;
  color: string;
}

// Render update types
export interface RenderUpdate {
  domOperations: DOMOperation[];
  styleChanges: StyleChange[];
  animationUpdates: AnimationUpdate[];
  timestamp: number;
}

export type DOMOperation = 
  | { type: 'Create'; elementId: string; tag: string; parentId?: string }
  | { type: 'Update'; elementId: string; attributes: Record<string, string> }
  | { type: 'Remove'; elementId: string }
  | { type: 'Move'; elementId: string; newParentId: string; index: number };

export interface StyleChange {
  elementId: string;
  property: string;
  value: string;
}

export interface AnimationUpdate {
  animationId: string;
  progress: number;
  currentValues: Record<string, any>;
}

// Interaction types
export interface InteractionEvent {
  eventType: InteractionType;
  targetElement?: string;
  position?: Position;
  data: Record<string, any>;
  timestamp: number;
}

export type InteractionType = 'Click' | 'Hover' | 'Touch' | 'Drag' | 'Scroll' | 'Keyboard' | 'Resize' | 'DataUpdate';

export interface Position {
  x: number;
  y: number;
}

// Error types
export interface WASMError {
  code: string;
  message: string;
  details?: Record<string, any>;
}

// Loader options
export interface LoaderOptions {
  validateSignatures?: boolean;
  enforcePermissions?: boolean;
  enableFallback?: boolean;
  maxMemoryUsage?: number;
  timeout?: number;
  sandboxConfig?: SandboxConfig;
}

// Renderer options
export interface RendererOptions {
  container: HTMLElement;
  permissions: LegacySecurityPolicy;
  enableInteractivity?: boolean;
  enableAnimations?: boolean;
  fallbackMode?: boolean;
  sandboxInterface?: SandboxInterface;
  domAPI?: SecureDOMAPI;
  communicationBridge?: CommunicationBridge;
}

// Editor types
export interface EditorOptions {
  mode: 'wysiwyg' | 'source' | 'split';
  enablePreview?: boolean;
  enableValidation?: boolean;
  autoSave?: boolean;
  theme?: string;
}

export interface EditorState {
  document: any; // Will be defined by WASM module
  selection: any;
  history: any;
  validationState: any;
  previewMode: string;
}

// Re-export error types and classes
export {
  LIVError,
  LIVErrorType,
  InvalidFileError,
  CorruptedFileError,
  UnsupportedFeatureError,
  SecurityError as LIVSecurityError,
  ValidationError,
  ResourceLimitError,
  TimeoutError,
  ParsingError,
  NetworkError,
  PermissionDeniedError,
  ErrorHandler,
  RecoveryStrategy,
  isLIVError,
  createErrorFromValidation,
  wrapAsyncOperation,
  withTimeout,
  setupGlobalErrorHandling
} from './errors';

// Re-export document class
export { LIVDocument, loadLIVDocument } from './document';