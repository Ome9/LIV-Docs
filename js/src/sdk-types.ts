// Additional TypeScript definitions for the LIV SDK
// Extends the existing types.ts with SDK-specific interfaces

import {
    DocumentMetadata,
    DocumentContent,
    LegacySecurityPolicy,
    FeatureFlags,
    WASMConfiguration,
    ValidationResult,
    RendererOptions,
    LoaderOptions
} from './types';

/**
 * Extended document creation options for the SDK
 */
export interface SDKDocumentCreationOptions {
    metadata?: Partial<DocumentMetadata>;
    content?: Partial<DocumentContent>;
    security?: Partial<LegacySecurityPolicy>;
    features?: Partial<FeatureFlags>;
    wasmConfig?: Partial<WASMConfiguration>;
    template?: DocumentTemplate;
}

/**
 * Document templates for quick creation
 */
export type DocumentTemplate = 
    | 'blank'
    | 'text-document'
    | 'presentation'
    | 'interactive-chart'
    | 'data-visualization'
    | 'form-document'
    | 'multimedia-story';

/**
 * SDK rendering options with additional features
 */
export interface SDKRenderingOptions extends Partial<RendererOptions> {
    enableFallback?: boolean;
    strictSecurity?: boolean;
    maxRenderTime?: number;
    enablePerformanceMonitoring?: boolean;
    enableAccessibility?: boolean;
    theme?: 'light' | 'dark' | 'auto';
    responsive?: boolean;
    mobileOptimized?: boolean;
}

/**
 * SDK editing options
 */
export interface SDKEditingOptions {
    mode?: 'wysiwyg' | 'source' | 'split';
    enablePreview?: boolean;
    enableValidation?: boolean;
    autoSave?: boolean;
    autoSaveInterval?: number;
    enableCollaboration?: boolean;
    theme?: 'light' | 'dark' | 'auto';
    enableCodeCompletion?: boolean;
    enableSyntaxHighlighting?: boolean;
}

/**
 * Asset management options
 */
export interface AssetManagementOptions {
    type: 'image' | 'font' | 'data' | 'audio' | 'video';
    name: string;
    data: ArrayBuffer | string | Blob;
    mimeType?: string;
    compression?: 'none' | 'gzip' | 'brotli';
    optimization?: boolean;
    metadata?: Record<string, any>;
}

/**
 * WASM module configuration for SDK
 */
export interface SDKWASMModuleOptions {
    name: string;
    data: ArrayBuffer;
    version?: string;
    entryPoint?: string;
    permissions?: Partial<LegacySecurityPolicy['wasmPermissions']>;
    dependencies?: string[];
    exports?: string[];
    imports?: string[];
    metadata?: Record<string, any>;
}

/**
 * Document export options
 */
export interface DocumentExportOptions {
    format: 'pdf' | 'html' | 'markdown' | 'epub' | 'json' | 'zip';
    includeAssets?: boolean;
    includeWASM?: boolean;
    compression?: boolean;
    quality?: 'low' | 'medium' | 'high';
    metadata?: boolean;
    signatures?: boolean;
}

/**
 * Document import options
 */
export interface DocumentImportOptions {
    format: 'html' | 'markdown' | 'pdf' | 'epub' | 'json';
    preserveFormatting?: boolean;
    extractAssets?: boolean;
    generateFallback?: boolean;
    securityPolicy?: Partial<LegacySecurityPolicy>;
}

/**
 * SDK validation options
 */
export interface SDKValidationOptions {
    validateSignatures?: boolean;
    validateIntegrity?: boolean;
    validateSecurity?: boolean;
    validateAccessibility?: boolean;
    validatePerformance?: boolean;
    strictMode?: boolean;
}

/**
 * Performance metrics for SDK operations
 */
export interface SDKPerformanceMetrics {
    loadTime: number;
    renderTime: number;
    memoryUsage: number;
    cpuUsage?: number;
    networkRequests?: number;
    cacheHitRate?: number;
    errorCount: number;
    warningCount: number;
}

/**
 * SDK configuration options
 */
export interface SDKConfiguration {
    enableCaching?: boolean;
    cacheSize?: number;
    enableLogging?: boolean;
    logLevel?: 'debug' | 'info' | 'warn' | 'error';
    enableMetrics?: boolean;
    enableSecurity?: boolean;
    enableAccessibility?: boolean;
    defaultSecurity?: Partial<LegacySecurityPolicy>;
    defaultFeatures?: Partial<FeatureFlags>;
}

/**
 * Document statistics and analytics
 */
export interface DocumentStatistics {
    size: {
        total: number;
        content: number;
        assets: number;
        wasm: number;
        metadata: number;
    };
    resources: {
        total: number;
        images: number;
        fonts: number;
        data: number;
        wasm: number;
    };
    features: {
        hasAnimations: boolean;
        hasInteractivity: boolean;
        hasCharts: boolean;
        hasForms: boolean;
        hasAudio: boolean;
        hasVideo: boolean;
        hasWASM: boolean;
    };
    security: {
        isSigned: boolean;
        hasPermissions: boolean;
        securityLevel: 'low' | 'medium' | 'high';
    };
    performance: {
        estimatedLoadTime: number;
        estimatedRenderTime: number;
        memoryRequirement: number;
        complexity: 'low' | 'medium' | 'high';
    };
}

/**
 * SDK event types for monitoring and callbacks
 */
export interface SDKEvents {
    'document:created': { document: any };
    'document:loaded': { document: any; loadTime: number };
    'document:rendered': { document: any; renderTime: number };
    'document:error': { error: Error; context: string };
    'validation:complete': { result: ValidationResult };
    'security:warning': { warning: string; severity: 'low' | 'medium' | 'high' };
    'performance:metrics': { metrics: SDKPerformanceMetrics };
    'asset:added': { asset: AssetManagementOptions };
    'wasm:loaded': { module: SDKWASMModuleOptions };
}

/**
 * SDK callback function types
 */
export type SDKEventCallback<T = any> = (data: T) => void;

/**
 * SDK error types specific to the high-level API
 */
export interface SDKError extends Error {
    code: string;
    category: 'validation' | 'security' | 'performance' | 'compatibility' | 'network';
    severity: 'low' | 'medium' | 'high' | 'critical';
    context?: Record<string, any>;
    suggestions?: string[];
}

/**
 * Document builder state for tracking changes
 */
export interface DocumentBuilderState {
    hasChanges: boolean;
    lastModified: Date;
    changeCount: number;
    validationState: 'valid' | 'invalid' | 'pending' | 'unknown';
    buildState: 'ready' | 'building' | 'error';
}

/**
 * Collaboration features (for future implementation)
 */
export interface CollaborationOptions {
    enabled: boolean;
    userId?: string;
    sessionId?: string;
    permissions?: 'read' | 'write' | 'admin';
    realTimeSync?: boolean;
    conflictResolution?: 'manual' | 'auto' | 'last-write-wins';
}

/**
 * Accessibility options for document creation and rendering
 */
export interface AccessibilityOptions {
    enableScreenReader?: boolean;
    enableKeyboardNavigation?: boolean;
    enableHighContrast?: boolean;
    enableLargeText?: boolean;
    enableReducedMotion?: boolean;
    altTextGeneration?: boolean;
    semanticStructure?: boolean;
    ariaLabels?: boolean;
}

/**
 * Mobile optimization options
 */
export interface MobileOptimizationOptions {
    enableTouchGestures?: boolean;
    enableSwipeNavigation?: boolean;
    optimizeForBattery?: boolean;
    adaptiveQuality?: boolean;
    offlineSupport?: boolean;
    progressiveLoading?: boolean;
    touchFriendlyUI?: boolean;
}

/**
 * Plugin system interfaces (for extensibility)
 */
export interface SDKPlugin {
    name: string;
    version: string;
    description?: string;
    initialize: (sdk: any) => Promise<void>;
    destroy?: () => Promise<void>;
    hooks?: Record<string, Function>;
}

export interface PluginManager {
    register(plugin: SDKPlugin): Promise<void>;
    unregister(pluginName: string): Promise<void>;
    getPlugin(name: string): SDKPlugin | undefined;
    listPlugins(): SDKPlugin[];
    executeHook(hookName: string, ...args: any[]): Promise<any[]>;
}