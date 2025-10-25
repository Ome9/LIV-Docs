/**
 * LIV JavaScript SDK Documentation
 * 
 * This file contains comprehensive documentation and examples for using the LIV JavaScript SDK.
 * The SDK provides a high-level API for creating, loading, rendering, and editing LIV documents.
 */

/**
 * @fileoverview LIV JavaScript SDK - Complete API Documentation
 * 
 * The LIV (Live Interactive Visual) JavaScript SDK provides a comprehensive set of tools
 * for working with LIV documents in web applications. It includes:
 * 
 * - Document creation and building
 * - Document loading from various sources
 * - Secure rendering with sandboxing
 * - WYSIWYG and source editing
 * - Asset management
 * - WASM module integration
 * - Format conversion
 * - Validation and security checking
 * 
 * @example Basic Usage
 * ```typescript
 * import { LIVSDK, LIVHelpers } from 'liv-document-format';
 * 
 * // Get SDK instance
 * const sdk = LIVSDK.getInstance();
 * 
 * // Create a simple document
 * const document = await LIVHelpers.createTextDocument(
 *   'My Document',
 *   'This is the content of my document.',
 *   'John Doe'
 * );
 * 
 * // Render the document
 * const container = document.getElementById('document-container');
 * const renderer = sdk.createRenderer(container);
 * await renderer.renderDocument(document);
 * ```
 * 
 * @example Advanced Document Creation
 * ```typescript
 * // Create a complex interactive document
 * const builder = await sdk.createDocument({
 *   metadata: {
 *     title: 'Interactive Dashboard',
 *     author: 'Data Team',
 *     description: 'Real-time analytics dashboard'
 *   },
 *   features: {
 *     interactivity: true,
 *     charts: true,
 *     animations: true
 *   }
 * });
 * 
 * // Add HTML content
 * builder.setHTML(`
 *   <div id="dashboard">
 *     <h1>Analytics Dashboard</h1>
 *     <div id="chart-container"></div>
 *     <div id="metrics"></div>
 *   </div>
 * `);
 * 
 * // Add CSS styling
 * builder.setCSS(`
 *   #dashboard {
 *     font-family: Arial, sans-serif;
 *     padding: 20px;
 *   }
 *   #chart-container {
 *     height: 400px;
 *     margin: 20px 0;
 *   }
 * `);
 * 
 * // Add interactive JavaScript
 * builder.setInteractiveSpec(`
 *   // Initialize dashboard when WASM is ready
 *   window.addEventListener('wasmReady', () => {
 *     initializeDashboard();
 *   });
 * `);
 * 
 * // Add assets
 * builder.addAsset({
 *   type: 'data',
 *   name: 'analytics.json',
 *   data: JSON.stringify(analyticsData)
 * });
 * 
 * // Build the document
 * const document = await builder.build();
 * ```
 * 
 * @example Document Loading and Validation
 * ```typescript
 * // Load document from file
 * const fileInput = document.getElementById('file-input') as HTMLInputElement;
 * const file = fileInput.files[0];
 * 
 * try {
 *   const document = await sdk.loadDocument(file, {
 *     validateSignatures: true,
 *     enforcePermissions: true
 *   });
 *   
 *   // Validate the document
 *   const validation = await sdk.validateDocument(document);
 *   if (!validation.isValid) {
 *     console.error('Document validation failed:', validation.errors);
 *   }
 *   
 *   // Get document information
 *   const info = sdk.getDocumentInfo(document);
 *   console.log('Document info:', info);
 *   
 * } catch (error) {
 *   console.error('Failed to load document:', error);
 * }
 * ```
 * 
 * @example Secure Rendering with Custom Permissions
 * ```typescript
 * // Create renderer with custom security policy
 * const renderer = sdk.createRenderer(container, {
 *   enableFallback: true,
 *   strictSecurity: true,
 *   permissions: {
 *     wasmPermissions: {
 *       memoryLimit: 32 * 1024 * 1024, // 32MB
 *       allowNetworking: false,
 *       allowFileSystem: false
 *     },
 *     jsPermissions: {
 *       executionMode: 'sandboxed',
 *       domAccess: 'read'
 *     }
 *   }
 * });
 * 
 * // Render with error handling
 * try {
 *   await renderer.renderDocument(document);
 *   console.log('Document rendered successfully');
 * } catch (error) {
 *   console.error('Rendering failed:', error);
 *   // Fallback content will be shown automatically if enabled
 * }
 * ```
 * 
 * @example Document Editing
 * ```typescript
 * // Create editor interface
 * const editor = sdk.createEditor(
 *   document.getElementById('editor'),
 *   document.getElementById('preview'),
 *   document.getElementById('toolbar'),
 *   document.getElementById('properties'),
 *   {
 *     mode: 'split',
 *     enablePreview: true,
 *     enableValidation: true,
 *     autoSave: true
 *   }
 * );
 * 
 * // Initialize with existing document
 * await editor.initialize(document);
 * 
 * // Listen for changes
 * editor.on('document:changed', (event) => {
 *   console.log('Document modified:', event.changes);
 * });
 * 
 * // Save changes
 * const updatedDocument = await editor.save();
 * ```
 * 
 * @example Asset Management
 * ```typescript
 * const builder = await sdk.createDocument();
 * 
 * // Add image asset
 * const imageFile = await fetch('/path/to/image.png');
 * const imageBuffer = await imageFile.arrayBuffer();
 * 
 * builder.addAsset({
 *   type: 'image',
 *   name: 'logo.png',
 *   data: imageBuffer,
 *   mimeType: 'image/png'
 * });
 * 
 * // Add font asset
 * const fontFile = await fetch('/path/to/font.woff2');
 * const fontBuffer = await fontFile.arrayBuffer();
 * 
 * builder.addAsset({
 *   type: 'font',
 *   name: 'custom-font.woff2',
 *   data: fontBuffer,
 *   mimeType: 'font/woff2'
 * });
 * 
 * // Add data asset
 * builder.addAsset({
 *   type: 'data',
 *   name: 'config.json',
 *   data: JSON.stringify({ theme: 'dark', version: '1.0' }),
 *   mimeType: 'application/json'
 * });
 * ```
 * 
 * @example WASM Module Integration
 * ```typescript
 * // Load WASM module
 * const wasmFile = await fetch('/path/to/module.wasm');
 * const wasmBuffer = await wasmFile.arrayBuffer();
 * 
 * // Add to document
 * builder.addWASMModule({
 *   name: 'chart-engine',
 *   data: wasmBuffer,
 *   version: '1.0.0',
 *   entryPoint: 'init_chart_engine',
 *   permissions: {
 *     memoryLimit: 16 * 1024 * 1024, // 16MB
 *     allowNetworking: false
 *   }
 * });
 * 
 * // Configure WASM in interactive spec
 * builder.setInteractiveSpec(`
 *   // Wait for WASM module to load
 *   window.addEventListener('wasmModuleLoaded', (event) => {
 *     if (event.detail.name === 'chart-engine') {
 *       const chartEngine = event.detail.module;
 *       // Use the WASM module
 *       chartEngine.createChart('myChart', chartData);
 *     }
 *   });
 * `);
 * ```
 * 
 * @example Helper Functions for Common Patterns
 * ```typescript
 * // Create a text document
 * const textDoc = await LIVHelpers.createTextDocument(
 *   'Meeting Notes',
 *   'Discussion points:\n1. Project timeline\n2. Resource allocation\n3. Next steps',
 *   'Meeting Organizer'
 * );
 * 
 * // Create a chart document
 * const chartDoc = await LIVHelpers.createChartDocument(
 *   'Sales Report',
 *   {
 *     labels: ['Q1', 'Q2', 'Q3', 'Q4'],
 *     datasets: [{
 *       label: 'Sales',
 *       data: [100, 150, 200, 180]
 *     }]
 *   },
 *   'bar'
 * );
 * 
 * // Create a presentation
 * const presentationDoc = await LIVHelpers.createPresentationDocument(
 *   'Product Launch',
 *   [
 *     { title: 'Introduction', content: 'Welcome to our product launch presentation.' },
 *     { title: 'Features', content: 'Our product includes amazing features...' },
 *     { title: 'Roadmap', content: 'Here is our development roadmap...' }
 *   ]
 * );
 * ```
 * 
 * @example Error Handling and Validation
 * ```typescript
 * import { LIVError, LIVErrorType } from 'liv-document-format';
 * 
 * try {
 *   const document = await sdk.loadDocument(file);
 *   const renderer = sdk.createRenderer(container);
 *   await renderer.renderDocument(document);
 * } catch (error) {
 *   if (error instanceof LIVError) {
 *     switch (error.type) {
 *       case LIVErrorType.INVALID_FILE:
 *         console.error('Invalid file format:', error.message);
 *         break;
 *       case LIVErrorType.SECURITY:
 *         console.error('Security violation:', error.message);
 *         break;
 *       case LIVErrorType.VALIDATION:
 *         console.error('Validation failed:', error.message);
 *         break;
 *       default:
 *         console.error('Unknown error:', error.message);
 *     }
 *   } else {
 *     console.error('Unexpected error:', error);
 *   }
 * }
 * ```
 * 
 * @example Performance Monitoring
 * ```typescript
 * // Enable performance monitoring
 * const renderer = sdk.createRenderer(container, {
 *   enablePerformanceMonitoring: true
 * });
 * 
 * // Listen for performance metrics
 * renderer.on('performance:metrics', (metrics) => {
 *   console.log('Render time:', metrics.renderTime);
 *   console.log('Memory usage:', metrics.memoryUsage);
 *   console.log('Frame rate:', metrics.averageFPS);
 * });
 * 
 * // Get current performance state
 * const performanceState = renderer.getPerformanceMetrics();
 * if (performanceState.averageFPS < 30) {
 *   console.warn('Low frame rate detected, consider reducing quality');
 * }
 * ```
 * 
 * @example Mobile Optimization
 * ```typescript
 * // Create mobile-optimized renderer
 * const renderer = sdk.createRenderer(container, {
 *   mobileOptimized: true,
 *   responsive: true,
 *   enableAnimations: true, // Will be automatically reduced on low-end devices
 *   targetFPS: 30 // Lower FPS for better battery life
 * });
 * 
 * // The renderer will automatically:
 * // - Detect mobile devices
 * // - Optimize touch interactions
 * // - Reduce animation quality on low-end devices
 * // - Adapt to orientation changes
 * // - Monitor battery level and adjust performance
 * ```
 * 
 * @example Accessibility Features
 * ```typescript
 * // Create accessible document
 * const builder = await sdk.createDocument({
 *   metadata: {
 *     title: 'Accessible Document',
 *     description: 'Document with accessibility features'
 *   }
 * });
 * 
 * // Add semantic HTML with proper ARIA labels
 * builder.setHTML(`
 *   <main role="main">
 *     <h1>Document Title</h1>
 *     <nav role="navigation" aria-label="Document sections">
 *       <ul>
 *         <li><a href="#section1">Section 1</a></li>
 *         <li><a href="#section2">Section 2</a></li>
 *       </ul>
 *     </nav>
 *     <section id="section1" aria-labelledby="heading1">
 *       <h2 id="heading1">Section 1</h2>
 *       <p>Content with proper semantic structure.</p>
 *     </section>
 *   </main>
 * `);
 * 
 * // Render with accessibility features
 * const renderer = sdk.createRenderer(container, {
 *   enableAccessibility: true,
 *   theme: 'auto' // Respects user's system preference
 * });
 * ```
 */

/**
 * SDK API Reference
 * 
 * This section provides detailed API documentation for all SDK classes and methods.
 */

export const SDKDocumentation = {
  /**
   * Main SDK class documentation
   */
  LIVSDK: {
    description: 'Main SDK class providing high-level API for LIV document operations',
    methods: {
      getInstance: {
        description: 'Get singleton instance of the SDK',
        returns: 'LIVSDK',
        example: 'const sdk = LIVSDK.getInstance();'
      },
      createDocument: {
        description: 'Create a new LIV document builder',
        parameters: {
          options: 'DocumentCreationOptions - Optional creation options'
        },
        returns: 'Promise<LIVDocumentBuilder>',
        example: 'const builder = await sdk.createDocument({ metadata: { title: "My Doc" } });'
      },
      loadDocument: {
        description: 'Load an existing LIV document from various sources',
        parameters: {
          source: 'File | string | ArrayBuffer - Document source',
          options: 'LoaderOptions - Optional loading options'
        },
        returns: 'Promise<LIVDocument>',
        example: 'const doc = await sdk.loadDocument(file);'
      },
      createRenderer: {
        description: 'Create a renderer for displaying LIV documents',
        parameters: {
          container: 'HTMLElement - Container element',
          options: 'RenderingOptions - Optional rendering options'
        },
        returns: 'LIVRenderer',
        example: 'const renderer = sdk.createRenderer(container);'
      },
      createEditor: {
        description: 'Create an editor for modifying LIV documents',
        parameters: {
          editorContainer: 'HTMLElement - Editor container',
          previewContainer: 'HTMLElement - Preview container',
          toolbarContainer: 'HTMLElement - Toolbar container',
          propertiesContainer: 'HTMLElement - Properties container',
          options: 'EditingOptions - Optional editing options'
        },
        returns: 'LIVEditor',
        example: 'const editor = sdk.createEditor(editorEl, previewEl, toolbarEl, propsEl);'
      }
    }
  },

  /**
   * Document builder class documentation
   */
  LIVDocumentBuilder: {
    description: 'Builder class for creating new LIV documents with fluent API',
    methods: {
      setMetadata: {
        description: 'Set document metadata',
        parameters: { metadata: 'Partial<DocumentMetadata>' },
        returns: 'LIVDocumentBuilder',
        chainable: true
      },
      setHTML: {
        description: 'Set HTML content',
        parameters: { html: 'string' },
        returns: 'LIVDocumentBuilder',
        chainable: true
      },
      setCSS: {
        description: 'Set CSS styles',
        parameters: { css: 'string' },
        returns: 'LIVDocumentBuilder',
        chainable: true
      },
      setInteractiveSpec: {
        description: 'Set interactive JavaScript specification',
        parameters: { spec: 'string' },
        returns: 'LIVDocumentBuilder',
        chainable: true
      },
      addAsset: {
        description: 'Add an asset to the document',
        parameters: { options: 'AssetOptions' },
        returns: 'LIVDocumentBuilder',
        chainable: true
      },
      addWASMModule: {
        description: 'Add a WASM module to the document',
        parameters: { options: 'WASMModuleOptions' },
        returns: 'LIVDocumentBuilder',
        chainable: true
      },
      build: {
        description: 'Build the final LIV document',
        returns: 'Promise<LIVDocument>',
        example: 'const doc = await builder.build();'
      }
    }
  },

  /**
   * Helper functions documentation
   */
  LIVHelpers: {
    description: 'Static helper functions for common document creation patterns',
    methods: {
      createTextDocument: {
        description: 'Create a simple text document',
        parameters: {
          title: 'string - Document title',
          content: 'string - Text content',
          author: 'string - Optional author name'
        },
        returns: 'Promise<LIVDocument>',
        example: 'const doc = await LIVHelpers.createTextDocument("Title", "Content", "Author");'
      },
      createChartDocument: {
        description: 'Create an interactive chart document',
        parameters: {
          title: 'string - Chart title',
          chartData: 'any - Chart data object',
          chartType: '"bar" | "line" | "pie" - Chart type'
        },
        returns: 'Promise<LIVDocument>',
        example: 'const doc = await LIVHelpers.createChartDocument("Sales", data, "bar");'
      },
      createPresentationDocument: {
        description: 'Create an animated presentation document',
        parameters: {
          title: 'string - Presentation title',
          slides: 'Array<{title: string, content: string}> - Slide data'
        },
        returns: 'Promise<LIVDocument>',
        example: 'const doc = await LIVHelpers.createPresentationDocument("Title", slides);'
      }
    }
  }
};

/**
 * Best Practices and Guidelines
 */
export const BestPractices = {
  security: [
    'Always validate documents before rendering',
    'Use strict security policies for untrusted content',
    'Enable signature validation for production use',
    'Limit WASM memory usage appropriately',
    'Sanitize user-provided HTML and CSS'
  ],
  
  performance: [
    'Use asset compression for large files',
    'Optimize images before adding to documents',
    'Limit concurrent animations on mobile devices',
    'Enable performance monitoring in production',
    'Use static fallback for complex interactive content'
  ],
  
  accessibility: [
    'Provide semantic HTML structure',
    'Include proper ARIA labels and roles',
    'Ensure keyboard navigation support',
    'Test with screen readers',
    'Respect user motion preferences'
  ],
  
  mobile: [
    'Enable mobile optimizations for touch devices',
    'Test on various screen sizes and orientations',
    'Consider battery usage for animations',
    'Optimize touch interaction areas',
    'Provide offline capabilities where possible'
  ]
};

/**
 * Common Patterns and Examples
 */
export const CommonPatterns = {
  documentCreation: `
    // Pattern: Create document with validation
    const builder = await sdk.createDocument();
    const document = await builder
      .setMetadata({ title: 'My Document', author: 'John Doe' })
      .setHTML('<h1>Hello World</h1>')
      .setCSS('h1 { color: blue; }')
      .build();
    
    const validation = document.validate();
    if (!validation.isValid) {
      throw new Error('Document validation failed');
    }
  `,
  
  secureRendering: `
    // Pattern: Secure rendering with fallback
    const renderer = sdk.createRenderer(container, {
      strictSecurity: true,
      enableFallback: true,
      permissions: {
        jsPermissions: { executionMode: 'sandboxed' }
      }
    });
    
    try {
      await renderer.renderDocument(document);
    } catch (error) {
      console.warn('Interactive rendering failed, using fallback');
    }
  `,
  
  assetManagement: `
    // Pattern: Efficient asset management
    const builder = await sdk.createDocument();
    
    // Add multiple assets efficiently
    const assets = [
      { type: 'image', name: 'logo.png', data: logoBuffer },
      { type: 'font', name: 'font.woff2', data: fontBuffer },
      { type: 'data', name: 'config.json', data: configString }
    ];
    
    assets.forEach(asset => builder.addAsset(asset));
  `,
  
  errorHandling: `
    // Pattern: Comprehensive error handling
    try {
      const document = await sdk.loadDocument(source);
      const renderer = sdk.createRenderer(container);
      await renderer.renderDocument(document);
    } catch (error) {
      if (error instanceof LIVError) {
        // Handle specific LIV errors
        handleLIVError(error);
      } else {
        // Handle unexpected errors
        console.error('Unexpected error:', error);
      }
    }
  `
};