const { contextBridge, ipcRenderer } = require('electron');

// Expose protected methods that allow the renderer process to use
// the ipcRenderer without exposing the entire object
contextBridge.exposeInMainWorld('electronAPI', {
    // Platform information
    platform: process.platform,
    versions: process.versions,

    // ===== Preferences Management =====
    getPreferences: () => ipcRenderer.invoke('get-preferences'),
    setPreferences: (preferences) => ipcRenderer.invoke('set-preferences', preferences),

    // ===== File Management =====
    getRecentFiles: () => ipcRenderer.invoke('get-recent-files'),
    openFileDialog: () => ipcRenderer.invoke('open-file-dialog'),

    // File dialogs with options
    showOpenDialog: (options) => ipcRenderer.invoke('show-open-dialog', options),
    showOpenDialogMultiple: (options) => ipcRenderer.invoke('show-open-dialog-multiple', options),
    showSaveDialog: (options) => ipcRenderer.invoke('show-save-dialog', options),

    // Simple file operations
    openFile: (options) => ipcRenderer.invoke('open-file', options),
    saveFile: (options) => ipcRenderer.invoke('save-file', options),
    readFile: (filePath) => ipcRenderer.invoke('read-file', filePath),
    writeFile: (filePath, content) => ipcRenderer.invoke('write-file', filePath, content),

    // ===== Document Operations =====
    // Open document
    openDocument: (filePath) => ipcRenderer.invoke('open-document', filePath),

    // Save document
    saveDocument: (data) => ipcRenderer.invoke('save-document', data),

    // Export document
    exportDocument: (options) => ipcRenderer.invoke('export-document', options),

    // Preview document
    previewDocument: (filePath) => ipcRenderer.invoke('preview-document', filePath),

    // Validate document
    validateDocument: (data) => ipcRenderer.invoke('validate-document', data),

    // Build document (using Go builder)
    buildDocument: (data) => ipcRenderer.invoke('build-document', data),

    // ===== PDF Operations =====
    // Merge PDFs
    mergePDFs: (filePaths, outputPath) => ipcRenderer.invoke('merge-pdfs', filePaths, outputPath),

    // Split PDF
    splitPDF: (filePath, options) => ipcRenderer.invoke('split-pdf', filePath, options),

    // Compress PDF
    compressPDF: (filePath, outputPath, quality) => ipcRenderer.invoke('compress-pdf', filePath, outputPath, quality),

    // Add watermark
    addWatermark: (filePath, watermarkText, options) => ipcRenderer.invoke('add-watermark', filePath, watermarkText, options),

    // Encrypt PDF
    encryptPDF: (filePath, password, options) => ipcRenderer.invoke('encrypt-pdf', filePath, password, options),

    // Sign PDF
    signPDF: (filePath, signature, options) => ipcRenderer.invoke('sign-pdf', filePath, signature, options),

    // Extract pages
    extractPages: (filePath, pageNumbers, outputPath) => ipcRenderer.invoke('extract-pages', filePath, pageNumbers, outputPath),

    // Rotate pages
    rotatePages: (filePath, rotation, pageNumbers) => ipcRenderer.invoke('rotate-pages', filePath, rotation, pageNumbers),

    // Additional PDF Operations
    addTextToPDF: (options) => ipcRenderer.invoke('pdf-add-text', options),
    addImageToPDF: (options) => ipcRenderer.invoke('pdf-add-image', options),
    addShapeToPDF: (options) => ipcRenderer.invoke('pdf-add-shape', options),
    addQRCodeToPDF: (options) => ipcRenderer.invoke('pdf-add-qrcode', options),
    addBarcodeToPDF: (options) => ipcRenderer.invoke('pdf-add-barcode', options),
    deletePDFPages: (pageNumbers) => ipcRenderer.invoke('pdf-delete-pages', pageNumbers),
    reorderPDFPages: (newOrder) => ipcRenderer.invoke('pdf-reorder-pages', newOrder),
    addBlankPage: (options) => ipcRenderer.invoke('pdf-add-blank-page', options),
    getPDFInfo: () => ipcRenderer.invoke('pdf-get-info'),
    setPDFInfo: (info) => ipcRenderer.invoke('pdf-set-info', info),
    imagesToPDF: (imagePaths) => ipcRenderer.invoke('images-to-pdf', imagePaths),

    // Select PDF file and return data
    selectPDF: () => ipcRenderer.invoke('select-pdf-file'),

    // ===== Go Backend PDF Operations =====
    goPDFExtractText: (filePath) => ipcRenderer.invoke('go-pdf-extract-text', filePath),
    goPDFMerge: (params) => ipcRenderer.invoke('go-pdf-merge', params),
    goPDFSplit: (params) => ipcRenderer.invoke('go-pdf-split', params),
    goPDFExtractPages: (params) => ipcRenderer.invoke('go-pdf-extract-pages', params),
    goPDFRotate: (params) => ipcRenderer.invoke('go-pdf-rotate', params),
    goPDFWatermark: (params) => ipcRenderer.invoke('go-pdf-watermark', params),
    goPDFCompress: (params) => ipcRenderer.invoke('go-pdf-compress', params),
    goPDFEncrypt: (params) => ipcRenderer.invoke('go-pdf-encrypt', params),
    goPDFInfo: (filePath) => ipcRenderer.invoke('go-pdf-info', filePath),
    goPDFSetInfo: (params) => ipcRenderer.invoke('go-pdf-set-info', params),
    goPDFToLIV: (params) => ipcRenderer.invoke('go-pdf-to-liv', params),

    // ===== File Operations for PDF Conversion =====
    saveTempFile: (file) => ipcRenderer.invoke('save-temp-file', file),
    readLIVFile: (filePath) => ipcRenderer.invoke('read-liv-file', filePath),

    // ===== Go Backend LIV Operations =====
    goLIVBuild: (params) => ipcRenderer.invoke('go-liv-build', params),
    goLIVValidate: (filePath) => ipcRenderer.invoke('go-liv-validate', filePath),

    // ===== Asset Management =====
    // Upload image
    uploadImage: (filePath) => ipcRenderer.invoke('upload-image', filePath),

    // Upload file
    uploadFile: (filePath) => ipcRenderer.invoke('upload-file', filePath),

    // Get asset URL
    getAssetURL: (assetPath) => ipcRenderer.invoke('get-asset-url', assetPath),

    // ===== Event Listeners =====
    // File operations
    onOpenFile: (callback) => {
        const listener = (event, filePath) => callback(filePath);
        ipcRenderer.on('open-file', listener);
        return () => ipcRenderer.removeListener('open-file', listener);
    },

    // Document load
    onLoadDocument: (callback) => {
        const listener = (event, data) => callback(data);
        ipcRenderer.on('load-document', listener);
        return () => ipcRenderer.removeListener('load-document', listener);
    },

    // Document save
    onDocumentSaved: (callback) => {
        const listener = (event, result) => callback(result);
        ipcRenderer.on('document-saved', listener);
        return () => ipcRenderer.removeListener('document-saved', listener);
    },

    // Build progress
    onBuildProgress: (callback) => {
        const listener = (event, progress) => callback(progress);
        ipcRenderer.on('build-progress', listener);
        return () => ipcRenderer.removeListener('build-progress', listener);
    },

    // Build complete
    onBuildComplete: (callback) => {
        const listener = (event, result) => callback(result);
        ipcRenderer.on('build-complete', listener);
        return () => ipcRenderer.removeListener('build-complete', listener);
    },

    // Error handling
    onError: (callback) => {
        const listener = (event, error) => callback(error);
        ipcRenderer.on('error', listener);
        return () => ipcRenderer.removeListener('error', listener);
    },

    // ===== Go Backend Integration =====
    // Call Go CLI commands
    callGoBuilder: (command, args) => ipcRenderer.invoke('call-go-builder', command, args),

    // Validate manifest
    validateManifest: (manifest) => ipcRenderer.invoke('validate-manifest', manifest),

    // Check integrity
    checkIntegrity: (filePath) => ipcRenderer.invoke('check-integrity', filePath),

    // Sign document
    signDocument: (filePath, keyPath) => ipcRenderer.invoke('sign-document', filePath, keyPath),

    // Verify signature
    verifySignature: (filePath) => ipcRenderer.invoke('verify-signature', filePath),

    // ===== WASM Operations =====
    // Load WASM module
    loadWASM: (wasmPath) => ipcRenderer.invoke('load-wasm', wasmPath),

    // Execute WASM function
    executeWASM: (moduleId, functionName, args) => ipcRenderer.invoke('execute-wasm', moduleId, functionName, args),

    // ===== System Integration =====
    openExternal: (url) => ipcRenderer.invoke('open-external', url),

    showInFolder: (filePath) => ipcRenderer.invoke('show-in-folder', filePath),

    copyToClipboard: (text) => ipcRenderer.invoke('copy-to-clipboard', text),

    // ===== Application Control =====
    quit: () => ipcRenderer.send('quit-app'),

    // Window control
    minimize: () => ipcRenderer.send('minimize-window'),
    maximize: () => ipcRenderer.send('maximize-window'),
    close: () => ipcRenderer.send('close-window'),

    isMaximized: () => ipcRenderer.invoke('is-maximized'),

    // ===== General IPC =====
    invoke: (channel, ...args) => ipcRenderer.invoke(channel, ...args),
    send: (channel, ...args) => ipcRenderer.send(channel, ...args),

    // ===== Development & Debugging =====
    log: (...args) => ipcRenderer.send('log', ...args),
    error: (...args) => ipcRenderer.send('error', ...args),
    warn: (...args) => ipcRenderer.send('warn', ...args),

    // Development helpers
    isDev: process.env.NODE_ENV === 'development',

    // Open DevTools
    toggleDevTools: () => ipcRenderer.send('toggle-devtools'),

    // ===== Security & Context =====
    isElectron: true,
    isDesktop: true,
    nodeVersion: process.versions.node,
    chromeVersion: process.versions.chrome,
    electronVersion: process.versions.electron
});

// Security: Remove Node.js globals from renderer process
delete window.require;
delete window.exports;
delete window.module;
