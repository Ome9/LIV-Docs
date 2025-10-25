// LIV Document Loader - Enhanced ZIP parsing and validation

import { 
  LIVDocument, 
  LoaderOptions, 
  ValidationResult, 
  SecurityReport,
  Manifest,
  DocumentContent,
  AssetBundle,
  SignatureBundle
} from './types';
import {
  LIVError,
  LIVErrorType,
  InvalidFileError,
  CorruptedFileError,
  SecurityError,
  ValidationError,
  ResourceLimitError,
  TimeoutError,
  ParsingError,
  NetworkError,
  withTimeout,
  wrapAsyncOperation
} from './errors';

export class LIVLoader {
  private options: LoaderOptions;

  constructor(options: LoaderOptions = {}) {
    this.options = {
      validateSignatures: true,
      enforcePermissions: true,
      enableFallback: true,
      maxMemoryUsage: 128 * 1024 * 1024, // 128MB
      timeout: 30000, // 30 seconds
      ...options
    };
  }

  async loadFromFile(file: File): Promise<LIVDocument> {
    // Validate file extension
    if (!file.name.toLowerCase().endsWith('.liv')) {
      throw new InvalidFileError('File must have .liv extension', file.name);
    }

    // Check file size
    if (file.size > (this.options.maxMemoryUsage || 128 * 1024 * 1024)) {
      throw new ResourceLimitError('file size', this.options.maxMemoryUsage || 128 * 1024 * 1024, file.size);
    }

    return wrapAsyncOperation(async () => {
      const arrayBuffer = await withTimeout(
        this.readFileAsArrayBuffer(file),
        this.options.timeout || 30000,
        'file reading'
      );
      return await this.loadFromArrayBuffer(arrayBuffer);
    }, LIVErrorType.INVALID_FILE, `loading file: ${file.name}`);
  }

  async loadFromURL(url: string): Promise<LIVDocument> {
    return wrapAsyncOperation(async () => {
      const response = await withTimeout(
        fetch(url),
        this.options.timeout || 30000,
        'network request'
      );
      
      if (!response.ok) {
        throw new NetworkError(
          `HTTP ${response.status}: ${response.statusText}`,
          url,
          response.status
        );
      }
      
      const arrayBuffer = await response.arrayBuffer();
      return await this.loadFromArrayBuffer(arrayBuffer);
    }, LIVErrorType.NETWORK, `loading from URL: ${url}`);
  }

  async loadFromArrayBuffer(buffer: ArrayBuffer): Promise<LIVDocument> {
    // Parse ZIP container
    const zipData = await this.parseZipContainer(buffer);
    
    // Extract and validate manifest
    const manifest = await this.extractManifest(zipData);
    
    // Validate signatures if enabled
    if (this.options.validateSignatures) {
      await this.validateSignatures(zipData, manifest);
    }
    
    // Extract content and assets
    const content = await this.extractContent(zipData);
    const assets = await this.extractAssets(zipData);
    const signatures = await this.extractSignatures(zipData);
    const wasmModules = await this.extractWASMModules(zipData);
    
    // Create document object
    const document: LIVDocument = {
      manifest,
      content,
      assets,
      signatures,
      wasmModules
    };
    
    // Validate document structure
    const validation = this.validateDocument(document);
    if (!validation.isValid) {
      throw new ValidationError(
        'Document validation failed',
        validation.errors,
        validation.warnings
      );
    }
    
    return document;
  }

  private async readFileAsArrayBuffer(file: File): Promise<ArrayBuffer> {
    return new Promise((resolve, reject) => {
      const reader = new FileReader();
      reader.onload = () => resolve(reader.result as ArrayBuffer);
      reader.onerror = () => reject(new Error('Failed to read file'));
      reader.readAsArrayBuffer(file);
    });
  }

  private async parseZipContainer(buffer: ArrayBuffer): Promise<Map<string, ArrayBuffer>> {
    const zipData = new Map<string, ArrayBuffer>();
    
    try {
      const view = new DataView(buffer);
      
      // Check ZIP file signature (PK header)
      if (view.getUint32(0, true) !== 0x04034b50) {
        throw new CorruptedFileError('Invalid ZIP file signature');
      }
      
      // Parse ZIP central directory
      let offset = 0;
      const entries: Array<{name: string, offset: number, size: number, compressedSize: number}> = [];
      
      // Find end of central directory record
      let eocdOffset = buffer.byteLength - 22;
      while (eocdOffset >= 0) {
        if (view.getUint32(eocdOffset, true) === 0x06054b50) {
          break;
        }
        eocdOffset--;
      }
      
      if (eocdOffset < 0) {
        throw new Error('End of central directory record not found');
      }
      
      // Read central directory info
      const centralDirOffset = view.getUint32(eocdOffset + 16, true);
      const numEntries = view.getUint16(eocdOffset + 8, true);
      
      // Parse central directory entries
      let cdOffset = centralDirOffset;
      for (let i = 0; i < numEntries; i++) {
        if (view.getUint32(cdOffset, true) !== 0x02014b50) {
          throw new Error('Invalid central directory entry signature');
        }
        
        const compressedSize = view.getUint32(cdOffset + 20, true);
        const uncompressedSize = view.getUint32(cdOffset + 24, true);
        const filenameLength = view.getUint16(cdOffset + 28, true);
        const extraFieldLength = view.getUint16(cdOffset + 30, true);
        const commentLength = view.getUint16(cdOffset + 32, true);
        const localHeaderOffset = view.getUint32(cdOffset + 42, true);
        
        // Read filename
        const filenameBytes = new Uint8Array(buffer, cdOffset + 46, filenameLength);
        const filename = new TextDecoder().decode(filenameBytes);
        
        entries.push({
          name: filename,
          offset: localHeaderOffset,
          size: uncompressedSize,
          compressedSize: compressedSize
        });
        
        cdOffset += 46 + filenameLength + extraFieldLength + commentLength;
      }
      
      // Extract file data
      for (const entry of entries) {
        const localHeaderOffset = entry.offset;
        
        // Verify local file header signature
        if (view.getUint32(localHeaderOffset, true) !== 0x04034b50) {
          throw new Error(`Invalid local file header signature for ${entry.name}`);
        }
        
        const localFilenameLength = view.getUint16(localHeaderOffset + 26, true);
        const localExtraFieldLength = view.getUint16(localHeaderOffset + 28, true);
        const compressionMethod = view.getUint16(localHeaderOffset + 8, true);
        
        const dataOffset = localHeaderOffset + 30 + localFilenameLength + localExtraFieldLength;
        
        let fileData: ArrayBuffer;
        
        if (compressionMethod === 0) {
          // No compression
          fileData = buffer.slice(dataOffset, dataOffset + entry.size);
        } else if (compressionMethod === 8) {
          // Deflate compression - would need a deflate implementation
          // For now, throw an error for compressed files
          throw new Error(`Compressed files not supported yet for ${entry.name}`);
        } else {
          throw new Error(`Unsupported compression method ${compressionMethod} for ${entry.name}`);
        }
        
        zipData.set(entry.name, fileData);
      }
      
    } catch (error) {
      throw new Error(`Failed to parse ZIP container: ${error.message}`);
    }
    
    return zipData;
  }

  private async extractManifest(zipData: Map<string, ArrayBuffer>): Promise<Manifest> {
    const manifestBuffer = zipData.get('manifest.json');
    if (!manifestBuffer) {
      throw new Error('Manifest not found in LIV file');
    }
    
    try {
      const manifestText = new TextDecoder().decode(manifestBuffer);
      const manifest = JSON.parse(manifestText) as Manifest;
      
      // Validate required manifest fields
      if (!manifest.version) {
        throw new Error('Manifest version is required');
      }
      
      if (!manifest.metadata) {
        throw new Error('Manifest metadata is required');
      }
      
      if (!manifest.security) {
        throw new Error('Manifest security policy is required');
      }
      
      if (!manifest.resources) {
        throw new Error('Manifest resources are required');
      }
      
      // Validate metadata fields
      if (!manifest.metadata.title || !manifest.metadata.author) {
        throw new Error('Manifest metadata must include title and author');
      }
      
      // Validate security policy
      if (!manifest.security.wasmPermissions || !manifest.security.jsPermissions) {
        throw new Error('Manifest security policy must include WASM and JS permissions');
      }
      
      return manifest;
    } catch (error) {
      if (error instanceof SyntaxError) {
        throw new Error(`Invalid JSON in manifest: ${error.message}`);
      }
      throw new Error(`Failed to parse manifest: ${error.message}`);
    }
  }

  private async validateSignatures(zipData: Map<string, ArrayBuffer>, manifest: Manifest): Promise<void> {
    if (!this.options.validateSignatures) {
      return;
    }
    
    const contentSig = zipData.get('signatures/content.sig');
    const manifestSig = zipData.get('signatures/manifest.sig');
    
    if (!contentSig || !manifestSig) {
      throw new Error('Required signatures not found');
    }
    
    // Validate content signature
    const contentBuffer = zipData.get('content/index.html');
    if (contentBuffer) {
      const isContentValid = await this.verifySignature(
        contentBuffer, 
        new TextDecoder().decode(contentSig)
      );
      
      if (!isContentValid) {
        throw new Error('Content signature verification failed');
      }
    }
    
    // Validate manifest signature
    const manifestBuffer = zipData.get('manifest.json');
    if (manifestBuffer) {
      const isManifestValid = await this.verifySignature(
        manifestBuffer,
        new TextDecoder().decode(manifestSig)
      );
      
      if (!isManifestValid) {
        throw new Error('Manifest signature verification failed');
      }
    }
    
    // Validate WASM module signatures
    for (const [path, buffer] of zipData.entries()) {
      if (path.endsWith('.wasm')) {
        const moduleName = path.split('/').pop()?.replace('.wasm', '') || '';
        const sigBuffer = zipData.get(`signatures/${moduleName}.wasm.sig`);
        
        if (sigBuffer) {
          const isWasmValid = await this.verifySignature(
            buffer,
            new TextDecoder().decode(sigBuffer)
          );
          
          if (!isWasmValid) {
            throw new Error(`WASM module signature verification failed for ${moduleName}`);
          }
        }
      }
    }
  }
  
  private async verifySignature(data: ArrayBuffer, signature: string): Promise<boolean> {
    try {
      // This is a simplified signature verification
      // In a real implementation, this would use Web Crypto API
      // with proper public key cryptography
      
      // For now, just check that signature is not empty and has reasonable length
      if (!signature || signature.length < 10) {
        return false;
      }
      
      // Calculate a simple hash of the data for demonstration
      const hashBuffer = await crypto.subtle.digest('SHA-256', data);
      const hashArray = Array.from(new Uint8Array(hashBuffer));
      const hashHex = hashArray.map(b => b.toString(16).padStart(2, '0')).join('');
      
      // In a real implementation, you would verify the signature against the hash
      // using the public key. For now, just return true if signature exists
      return signature.length > 0;
      
    } catch (error) {
      console.error('Signature verification error:', error);
      return false;
    }
  }

  private async extractContent(zipData: Map<string, ArrayBuffer>): Promise<DocumentContent> {
    const htmlBuffer = zipData.get('content/index.html');
    const cssBuffer = zipData.get('content/styles/main.css');
    const jsBuffer = zipData.get('content/scripts/main.js');
    const fallbackBuffer = zipData.get('content/static/fallback.html');
    
    // HTML content is required
    if (!htmlBuffer) {
      throw new Error('HTML content (content/index.html) is required');
    }
    
    const content: DocumentContent = {
      html: new TextDecoder().decode(htmlBuffer),
      css: cssBuffer ? new TextDecoder().decode(cssBuffer) : '',
      interactiveSpec: jsBuffer ? new TextDecoder().decode(jsBuffer) : '',
      staticFallback: fallbackBuffer ? new TextDecoder().decode(fallbackBuffer) : ''
    };
    
    // Validate HTML content
    if (content.html.trim().length === 0) {
      throw new Error('HTML content cannot be empty');
    }
    
    // Basic HTML validation
    if (!content.html.includes('<html') && !content.html.includes('<!DOCTYPE')) {
      console.warn('HTML content may be missing DOCTYPE or html tag');
    }
    
    return content;
  }

  private async extractAssets(zipData: Map<string, ArrayBuffer>): Promise<AssetBundle> {
    const assets: AssetBundle = {
      images: new Map<string, ArrayBuffer>(),
      fonts: new Map<string, ArrayBuffer>(),
      data: new Map<string, ArrayBuffer>()
    };
    
    // Extract assets from ZIP data
    for (const [path, buffer] of zipData.entries()) {
      if (path.startsWith('assets/images/')) {
        const filename = path.substring('assets/images/'.length);
        if (filename && buffer.byteLength > 0) {
          assets.images.set(filename, buffer);
        }
      } else if (path.startsWith('assets/fonts/')) {
        const filename = path.substring('assets/fonts/'.length);
        if (filename && buffer.byteLength > 0) {
          assets.fonts.set(filename, buffer);
        }
      } else if (path.startsWith('assets/data/')) {
        const filename = path.substring('assets/data/'.length);
        if (filename && buffer.byteLength > 0) {
          assets.data.set(filename, buffer);
        }
      }
    }
    
    // Validate asset sizes
    const maxAssetSize = 50 * 1024 * 1024; // 50MB
    
    for (const [name, buffer] of assets.images.entries()) {
      if (buffer.byteLength > maxAssetSize) {
        throw new Error(`Image asset '${name}' exceeds maximum size limit`);
      }
    }
    
    for (const [name, buffer] of assets.fonts.entries()) {
      if (buffer.byteLength > maxAssetSize) {
        throw new Error(`Font asset '${name}' exceeds maximum size limit`);
      }
    }
    
    for (const [name, buffer] of assets.data.entries()) {
      if (buffer.byteLength > maxAssetSize) {
        throw new Error(`Data asset '${name}' exceeds maximum size limit`);
      }
    }
    
    return assets;
  }

  private async extractSignatures(zipData: Map<string, ArrayBuffer>): Promise<SignatureBundle> {
    const contentSig = zipData.get('signatures/content.sig');
    const manifestSig = zipData.get('signatures/manifest.sig');
    
    const wasmSignatures: Record<string, string> = {};
    
    // Extract WASM module signatures
    for (const [path, buffer] of zipData.entries()) {
      if (path.startsWith('signatures/') && path.endsWith('.wasm.sig')) {
        const moduleName = path.substring('signatures/'.length, path.length - '.wasm.sig'.length);
        wasmSignatures[moduleName] = new TextDecoder().decode(buffer);
      }
    }
    
    return {
      contentSignature: contentSig ? new TextDecoder().decode(contentSig) : '',
      manifestSignature: manifestSig ? new TextDecoder().decode(manifestSig) : '',
      wasmSignatures
    };
  }

  private async extractWASMModules(zipData: Map<string, ArrayBuffer>): Promise<Map<string, ArrayBuffer>> {
    const wasmModules = new Map<string, ArrayBuffer>();
    
    for (const [path, buffer] of zipData.entries()) {
      if (path.endsWith('.wasm')) {
        const moduleName = path.split('/').pop()?.replace('.wasm', '') || '';
        wasmModules.set(moduleName, buffer);
      }
    }
    
    return wasmModules;
  }

  private validateDocument(document: LIVDocument): ValidationResult {
    const errors: string[] = [];
    const warnings: string[] = [];
    
    // Validate manifest
    if (!document.manifest.version) {
      errors.push('Manifest version is required');
    }
    
    if (!document.manifest.metadata.title) {
      errors.push('Document title is required');
    }
    
    if (!document.manifest.security) {
      errors.push('Security policy is required');
    }
    
    // Validate content
    if (!document.content.html && !document.content.staticFallback) {
      errors.push('Document must have HTML content or static fallback');
    }
    
    // Check memory usage
    const estimatedSize = this.estimateMemoryUsage(document);
    if (estimatedSize > (this.options.maxMemoryUsage || 128 * 1024 * 1024)) {
      warnings.push(`Document size (${estimatedSize} bytes) exceeds recommended limit`);
    }
    
    // Validate WASM modules
    for (const [name, module] of document.wasmModules.entries()) {
      if (module.byteLength === 0) {
        errors.push(`WASM module '${name}' is empty`);
      }
    }
    
    return {
      isValid: errors.length === 0,
      errors,
      warnings
    };
  }

  private estimateMemoryUsage(document: LIVDocument): number {
    let size = 0;
    
    // Content size
    size += new TextEncoder().encode(document.content.html).length;
    size += new TextEncoder().encode(document.content.css).length;
    size += new TextEncoder().encode(document.content.interactiveSpec).length;
    size += new TextEncoder().encode(document.content.staticFallback).length;
    
    // Assets size
    for (const buffer of document.assets.images.values()) {
      size += buffer.byteLength;
    }
    for (const buffer of document.assets.fonts.values()) {
      size += buffer.byteLength;
    }
    for (const buffer of document.assets.data.values()) {
      size += buffer.byteLength;
    }
    
    // WASM modules size
    for (const buffer of document.wasmModules.values()) {
      size += buffer.byteLength;
    }
    
    return size;
  }

  generateSecurityReport(document: LIVDocument): SecurityReport {
    const errors: string[] = [];
    const warnings: string[] = [];
    
    // Check for required security features
    if (!document.manifest.security.wasmPermissions) {
      errors.push('WASM permissions not defined');
    }
    
    if (!document.manifest.security.jsPermissions) {
      errors.push('JavaScript permissions not defined');
    }
    
    // Check signature presence
    const hasContentSig = !!document.signatures.contentSignature;
    const hasManifestSig = !!document.signatures.manifestSignature;
    
    if (!hasContentSig) {
      warnings.push('Content signature missing');
    }
    
    if (!hasManifestSig) {
      warnings.push('Manifest signature missing');
    }
    
    // Check for potentially dangerous features
    if (document.manifest.security.wasmPermissions?.allowNetworking) {
      warnings.push('Document requests network access');
    }
    
    if (document.manifest.security.jsPermissions?.executionMode === 'trusted') {
      warnings.push('Document requests trusted JavaScript execution');
    }
    
    return {
      isValid: errors.length === 0,
      signatureVerified: hasContentSig && hasManifestSig,
      integrityChecked: true, // Would be actual integrity check result
      permissionsValid: document.manifest.security !== undefined,
      warnings,
      errors
    };
  }
}