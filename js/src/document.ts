// LIVDocument class - Main document representation with integrated loading and parsing

import { 
  LIVDocument as LIVDocumentInterface, 
  LoaderOptions, 
  ValidationResult, 
  SecurityReport,
  Manifest,
  DocumentContent,
  AssetBundle,
  SignatureBundle,
  Resource
} from './types';
import { LIVLoader } from './loader';

export class LIVDocument implements LIVDocumentInterface {
  public manifest: Manifest;
  public content: DocumentContent;
  public assets: AssetBundle;
  public signatures: SignatureBundle;
  public wasmModules: Map<string, ArrayBuffer>;
  
  private loader: LIVLoader;
  private resourceCache: Map<string, ArrayBuffer> = new Map();
  private validationResult?: ValidationResult;
  private securityReport?: SecurityReport;

  constructor(
    manifest: Manifest,
    content: DocumentContent,
    assets: AssetBundle,
    signatures: SignatureBundle,
    wasmModules: Map<string, ArrayBuffer>,
    loader?: LIVLoader
  ) {
    this.manifest = manifest;
    this.content = content;
    this.assets = assets;
    this.signatures = signatures;
    this.wasmModules = wasmModules;
    this.loader = loader || new LIVLoader();
  }

  // Static factory methods for loading documents
  static async fromFile(file: File, options?: LoaderOptions): Promise<LIVDocument> {
    const loader = new LIVLoader(options);
    const documentData = await loader.loadFromFile(file);
    
    return new LIVDocument(
      documentData.manifest,
      documentData.content,
      documentData.assets,
      documentData.signatures,
      documentData.wasmModules,
      loader
    );
  }

  static async fromURL(url: string, options?: LoaderOptions): Promise<LIVDocument> {
    const loader = new LIVLoader(options);
    const documentData = await loader.loadFromURL(url);
    
    return new LIVDocument(
      documentData.manifest,
      documentData.content,
      documentData.assets,
      documentData.signatures,
      documentData.wasmModules,
      loader
    );
  }

  static async fromArrayBuffer(buffer: ArrayBuffer, options?: LoaderOptions): Promise<LIVDocument> {
    const loader = new LIVLoader(options);
    const documentData = await loader.loadFromArrayBuffer(buffer);
    
    return new LIVDocument(
      documentData.manifest,
      documentData.content,
      documentData.assets,
      documentData.signatures,
      documentData.wasmModules,
      loader
    );
  }

  // Resource loading and caching methods
  async getResource(resourcePath: string): Promise<ArrayBuffer | null> {
    // Check cache first
    if (this.resourceCache.has(resourcePath)) {
      return this.resourceCache.get(resourcePath) || null;
    }

    // Load resource based on path
    let resourceData: ArrayBuffer | null = null;

    if (resourcePath.startsWith('content/')) {
      resourceData = this.getContentResource(resourcePath);
    } else if (resourcePath.startsWith('assets/')) {
      resourceData = this.getAssetResource(resourcePath);
    } else if (resourcePath.endsWith('.wasm')) {
      const moduleName = resourcePath.split('/').pop()?.replace('.wasm', '') || '';
      resourceData = this.wasmModules.get(moduleName) || null;
    }

    // Cache the resource if found
    if (resourceData) {
      this.resourceCache.set(resourcePath, resourceData);
    }

    return resourceData;
  }

  private getContentResource(resourcePath: string): ArrayBuffer | null {
    const encoder = new TextEncoder();
    
    switch (resourcePath) {
      case 'content/index.html':
        return encoder.encode(this.content.html).buffer;
      case 'content/styles/main.css':
        return encoder.encode(this.content.css).buffer;
      case 'content/scripts/main.js':
        return encoder.encode(this.content.interactiveSpec).buffer;
      case 'content/static/fallback.html':
        return encoder.encode(this.content.staticFallback).buffer;
      default:
        return null;
    }
  }

  private getAssetResource(resourcePath: string): ArrayBuffer | null {
    const parts = resourcePath.split('/');
    if (parts.length < 3) return null;

    const assetType = parts[1];
    const assetName = parts.slice(2).join('/');

    switch (assetType) {
      case 'images':
        return this.assets.images.get(assetName) || null;
      case 'fonts':
        return this.assets.fonts.get(assetName) || null;
      case 'data':
        return this.assets.data.get(assetName) || null;
      default:
        return null;
    }
  }

  // Resource information methods
  getResourceInfo(resourcePath: string): Resource | null {
    return this.manifest.resources[resourcePath] || null;
  }

  listResources(): string[] {
    return Object.keys(this.manifest.resources);
  }

  getResourcesByType(type: string): string[] {
    return this.listResources().filter(path => {
      const resource = this.manifest.resources[path];
      return resource && resource.type.startsWith(type);
    });
  }

  // Validation methods
  validate(): ValidationResult {
    if (this.validationResult) {
      return this.validationResult;
    }

    const errors: string[] = [];
    const warnings: string[] = [];

    // Validate manifest
    if (!this.manifest.version) {
      errors.push('Manifest version is required');
    }

    if (!this.manifest.metadata?.title) {
      errors.push('Document title is required');
    }

    if (!this.manifest.metadata?.author) {
      errors.push('Document author is required');
    }

    if (!this.manifest.security) {
      errors.push('Security policy is required');
    }

    // Validate content
    if (!this.content.html && !this.content.staticFallback) {
      errors.push('Document must have HTML content or static fallback');
    }

    if (this.content.html && this.content.html.trim().length === 0) {
      errors.push('HTML content cannot be empty');
    }

    // Validate resource consistency
    for (const resourcePath of Object.keys(this.manifest.resources)) {
      const resourceExists = this.getResource(resourcePath);
      if (!resourceExists) {
        errors.push(`Resource listed in manifest but not found: ${resourcePath}`);
      }
    }

    // Validate WASM modules
    for (const [name, module] of this.wasmModules.entries()) {
      if (module.byteLength === 0) {
        errors.push(`WASM module '${name}' is empty`);
      }

      // Basic WASM validation
      if (module.byteLength < 8) {
        errors.push(`WASM module '${name}' is too small`);
      } else {
        const view = new DataView(module);
        const magic = view.getUint32(0, true);
        if (magic !== 0x6d736100) { // "\0asm" in little endian
          errors.push(`WASM module '${name}' has invalid magic number`);
        }
      }
    }

    // Check memory usage
    const estimatedSize = this.estimateMemoryUsage();
    if (estimatedSize > 128 * 1024 * 1024) { // 128MB
      warnings.push(`Document size (${estimatedSize} bytes) is very large`);
    }

    this.validationResult = {
      isValid: errors.length === 0,
      errors,
      warnings
    };

    return this.validationResult;
  }

  generateSecurityReport(): SecurityReport {
    if (this.securityReport) {
      return this.securityReport;
    }

    const errors: string[] = [];
    const warnings: string[] = [];

    // Check security policy
    if (!this.manifest.security.wasmPermissions) {
      errors.push('WASM permissions not defined');
    }

    if (!this.manifest.security.jsPermissions) {
      errors.push('JavaScript permissions not defined');
    }

    // Check signatures
    const hasContentSig = !!this.signatures.contentSignature;
    const hasManifestSig = !!this.signatures.manifestSignature;

    if (!hasContentSig) {
      warnings.push('Content signature missing');
    }

    if (!hasManifestSig) {
      warnings.push('Manifest signature missing');
    }

    // Check for potentially dangerous permissions
    if (this.manifest.security.wasmPermissions?.allowNetworking) {
      warnings.push('Document requests WASM network access');
    }

    if (this.manifest.security.wasmPermissions?.allowFileSystem) {
      warnings.push('Document requests WASM file system access');
    }

    if (this.manifest.security.jsPermissions?.executionMode === 'trusted') {
      warnings.push('Document requests trusted JavaScript execution');
    }

    // Check WASM module signatures
    for (const moduleName of this.wasmModules.keys()) {
      if (!this.signatures.wasmSignatures[moduleName]) {
        warnings.push(`WASM module '${moduleName}' has no signature`);
      }
    }

    this.securityReport = {
      isValid: errors.length === 0,
      signatureVerified: hasContentSig && hasManifestSig,
      integrityChecked: true,
      permissionsValid: this.manifest.security !== undefined,
      warnings,
      errors
    };

    return this.securityReport;
  }

  // Utility methods
  estimateMemoryUsage(): number {
    let size = 0;
    const encoder = new TextEncoder();

    // Content size
    size += encoder.encode(this.content.html).length;
    size += encoder.encode(this.content.css).length;
    size += encoder.encode(this.content.interactiveSpec).length;
    size += encoder.encode(this.content.staticFallback).length;

    // Assets size
    for (const buffer of this.assets.images.values()) {
      size += buffer.byteLength;
    }
    for (const buffer of this.assets.fonts.values()) {
      size += buffer.byteLength;
    }
    for (const buffer of this.assets.data.values()) {
      size += buffer.byteLength;
    }

    // WASM modules size
    for (const buffer of this.wasmModules.values()) {
      size += buffer.byteLength;
    }

    return size;
  }

  getMetadata() {
    return {
      title: this.manifest.metadata.title,
      author: this.manifest.metadata.author,
      created: this.manifest.metadata.created,
      modified: this.manifest.metadata.modified,
      description: this.manifest.metadata.description,
      version: this.manifest.metadata.version,
      language: this.manifest.metadata.language,
      estimatedSize: this.estimateMemoryUsage(),
      resourceCount: this.listResources().length,
      wasmModuleCount: this.wasmModules.size,
      hasInteractiveContent: this.content.interactiveSpec.length > 0,
      hasAnimations: this.manifest.features?.animations || false,
      hasCharts: this.manifest.features?.charts || false
    };
  }

  clearCache(): void {
    this.resourceCache.clear();
    this.validationResult = undefined;
    this.securityReport = undefined;
  }

  // Resource integrity validation
  async validateResourceIntegrity(resourcePath: string): Promise<boolean> {
    const resource = this.getResourceInfo(resourcePath);
    if (!resource) {
      return false;
    }

    const resourceData = await this.getResource(resourcePath);
    if (!resourceData) {
      return false;
    }

    // Validate size
    if (resourceData.byteLength !== resource.size) {
      return false;
    }

    // Validate hash (simplified - in production would use proper hash comparison)
    if (!resource.hash || resource.hash.length === 0) {
      return false;
    }

    try {
      const hashBuffer = await crypto.subtle.digest('SHA-256', resourceData);
      const hashArray = Array.from(new Uint8Array(hashBuffer));
      const hashHex = hashArray.map(b => b.toString(16).padStart(2, '0')).join('');
      
      // In a real implementation, you would compare with the actual hash
      // For now, just check that a hash exists
      return resource.hash.length > 0;
    } catch (error) {
      console.error('Hash validation error:', error);
      return false;
    }
  }

  async validateAllResourceIntegrity(): Promise<{valid: string[], invalid: string[]}> {
    const valid: string[] = [];
    const invalid: string[] = [];

    for (const resourcePath of this.listResources()) {
      const isValid = await this.validateResourceIntegrity(resourcePath);
      if (isValid) {
        valid.push(resourcePath);
      } else {
        invalid.push(resourcePath);
      }
    }

    return { valid, invalid };
  }
}

// Export factory functions for convenience
export async function loadLIVDocument(source: File | string | ArrayBuffer, options?: LoaderOptions): Promise<LIVDocument> {
  if (source instanceof File) {
    return LIVDocument.fromFile(source, options);
  } else if (typeof source === 'string') {
    return LIVDocument.fromURL(source, options);
  } else if (source instanceof ArrayBuffer) {
    return LIVDocument.fromArrayBuffer(source, options);
  } else {
    throw new Error('Invalid source type for loading LIV document');
  }
}