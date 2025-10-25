# Desktop Application Integration Guide

This document describes how the LIV Desktop Application integrates with the existing LIV ecosystem and provides guidance for developers working with the desktop wrapper.

## Architecture Overview

The desktop application follows a multi-process architecture that integrates seamlessly with the existing LIV infrastructure:

```
┌─────────────────────────────────────────────────────────────┐
│                    Desktop Application                      │
├─────────────────────────────────────────────────────────────┤
│  ┌─────────────────┐    ┌─────────────────────────────────┐ │
│  │  Main Process   │    │      Renderer Process          │ │
│  │  (Electron)     │    │      (Web Interface)           │ │
│  │                 │    │                                 │ │
│  │ • Window Mgmt   │◄──►│ • LIV Viewer UI                │ │
│  │ • File System   │    │ • Document Display             │ │
│  │ • Native APIs   │    │ • User Interactions            │ │
│  │ • Auto Updates  │    │ • Settings Interface           │ │
│  └─────────────────┘    └─────────────────────────────────┘ │
└─────────────────────────────────────────────────────────────┘
                                   │
                                   │ HTTP/WebSocket
                                   ▼
┌─────────────────────────────────────────────────────────────┐
│                   LIV Viewer Process                        │
├─────────────────────────────────────────────────────────────┤
│  ┌─────────────────┐    ┌─────────────────────────────────┐ │
│  │   Go Server     │    │      WASM Runtime              │ │
│  │                 │    │                                 │ │
│  │ • HTTP Server   │◄──►│ • Interactive Engine           │ │
│  │ • Document API  │    │ • Chart Framework              │ │
│  │ • Security      │    │ • User Interaction             │ │
│  │ • Validation    │    │ • Memory Management            │ │
│  └─────────────────┘    └─────────────────────────────────┘ │
└─────────────────────────────────────────────────────────────┘
```

## Integration Points

### 1. Existing LIV Viewer Integration

The desktop application reuses the existing web-based LIV viewer:

```javascript
// Main process starts the Go viewer server
async function startViewerProcess() {
  const viewerPath = getViewerExecutablePath();
  const args = ['--web', '--port', port.toString()];
  
  viewerProcess = spawn(viewerPath, args);
  
  // Wait for server to start, then load in renderer
  setTimeout(() => {
    mainWindow.loadURL(`http://localhost:${port}`);
  }, 2000);
}
```

**Benefits:**
- No code duplication
- Consistent rendering across platforms
- Automatic security and validation
- Full WASM interactive engine support

### 2. File System Integration

The desktop app provides native file system access while maintaining security:

```javascript
// Secure file opening with validation
async function openFile() {
  const result = await dialog.showOpenDialog(mainWindow, {
    filters: [{ name: 'LIV Documents', extensions: ['liv'] }]
  });
  
  if (!result.canceled) {
    // Validate file before opening
    const filePath = result.filePaths[0];
    openFileByPath(filePath);
  }
}
```

**Features:**
- Native file dialogs
- File association support
- Recent files tracking
- Drag-and-drop support

### 3. Security Model Integration

The desktop application maintains the existing security model:

```javascript
// Secure renderer process configuration
webPreferences: {
  nodeIntegration: false,        // No Node.js in renderer
  contextIsolation: true,        // Isolated context
  enableRemoteModule: false,     // No remote module
  webSecurity: true,             // Web security enabled
  preload: path.join(__dirname, 'preload.js')
}
```

**Security Features:**
- Process isolation
- Context bridge for IPC
- Sandboxed document execution
- Content Security Policy

### 4. Settings and Preferences

Desktop-specific settings extend the existing configuration system:

```javascript
// Settings stored using electron-store
const store = new Store({
  defaults: {
    windowBounds: { width: 1200, height: 800 },
    recentFiles: [],
    preferences: {
      theme: 'system',
      autoUpdate: true,
      fallbackMode: false,  // Uses existing fallback system
      debugMode: false      // Enables existing debug features
    }
  }
});
```

## Development Integration

### Building with Existing Infrastructure

The desktop application integrates with the existing build system:

```bash
# Build the Go viewer first
go build -o bin/liv-viewer cmd/viewer/main.go

# Then build the desktop application
cd desktop
npm install
npm run build
```

### Testing Integration

Desktop tests work alongside existing tests:

```bash
# Run existing Go tests
go test ./...

# Run existing JavaScript tests
cd js && npm test

# Run desktop application tests
cd desktop && npm test
```

### Development Workflow

1. **Start Go viewer in development mode:**
   ```bash
   go run cmd/viewer/main.go --web --debug
   ```

2. **Start desktop application in development:**
   ```bash
   cd desktop
   npm run dev
   ```

3. **The desktop app will automatically connect to the Go viewer**

## API Integration

### Existing API Compatibility

The desktop application uses all existing LIV APIs:

- **Document API** (`/api/document`): Load and parse documents
- **Upload API** (`/api/upload`): Handle file uploads
- **Validation API** (`/api/validate`): Validate document integrity
- **Static Assets** (`/static/`): Serve web assets

### Desktop-Specific APIs

Additional APIs for desktop functionality:

```javascript
// Desktop-specific endpoints
app.use('/api/desktop/', (req, res) => {
  switch (req.path) {
    case '/system-info':
      // Return system information
      break;
    case '/recent-files':
      // Manage recent files
      break;
    case '/settings':
      // Desktop settings management
      break;
  }
});
```

## Packaging Integration

### Asset Bundling

The desktop application bundles existing web assets:

```json
{
  "build": {
    "files": [
      "src/**/*",
      "assets/**/*"
    ],
    "extraResources": [
      {
        "from": "../bin/",
        "to": "bin/",
        "filter": ["**/*"]
      }
    ]
  }
}
```

### Cross-Platform Distribution

Builds integrate with existing CI/CD:

```yaml
# GitHub Actions example
- name: Build Go Viewer
  run: |
    GOOS=windows GOARCH=amd64 go build -o bin/liv-viewer.exe cmd/viewer/main.go
    GOOS=darwin GOARCH=amd64 go build -o bin/liv-viewer-mac cmd/viewer/main.go
    GOOS=linux GOARCH=amd64 go build -o bin/liv-viewer-linux cmd/viewer/main.go

- name: Build Desktop App
  run: |
    cd desktop
    npm ci
    npm run build
```

## Configuration Integration

### Environment Variables

Desktop application respects existing environment variables:

- `LIV_DEBUG`: Enable debug mode
- `LIV_FALLBACK`: Force fallback mode
- `LIV_PORT`: Default port for viewer
- `LIV_SECURITY_STRICT`: Enhanced security mode

### Configuration Files

Desktop settings complement existing configuration:

```json
{
  "viewer": {
    "port": 8080,
    "fallback": false,
    "debug": false
  },
  "desktop": {
    "theme": "system",
    "autoUpdate": true,
    "windowBounds": { "width": 1200, "height": 800 }
  }
}
```

## Error Handling Integration

### Existing Error System

Desktop application uses existing error handling:

```javascript
// Reuse existing error types and handling
try {
  await loadDocument(filePath);
} catch (error) {
  if (error instanceof ValidationError) {
    showValidationError(error);
  } else if (error instanceof SecurityError) {
    showSecurityError(error);
  }
}
```

### Desktop-Specific Errors

Additional error handling for desktop features:

```javascript
// Desktop-specific error handling
viewerProcess.on('error', (error) => {
  console.error('Viewer process error:', error);
  showErrorPage(error);
});

viewerProcess.on('exit', (code) => {
  if (code !== 0) {
    dialog.showErrorBox('Viewer Error', 
      'The LIV viewer process has stopped unexpectedly.');
  }
});
```

## Performance Integration

### Resource Management

Desktop application optimizes existing resource usage:

```javascript
// Memory management
app.on('window-all-closed', () => {
  // Clean up viewer process
  stopViewerProcess();
  
  // Clean up resources
  if (process.platform !== 'darwin') {
    app.quit();
  }
});
```

### Caching Strategy

Leverages existing caching mechanisms:

- Document cache from Go viewer
- Asset cache from web interface
- Desktop-specific UI cache

## Future Integration Points

### Plugin System

Desktop application will integrate with planned plugin system:

```javascript
// Plugin loading (future)
const pluginManager = new PluginManager({
  pluginDir: path.join(app.getPath('userData'), 'plugins'),
  securityPolicy: 'strict'
});
```

### Cloud Integration

Desktop application will support cloud features:

```javascript
// Cloud sync (future)
const cloudSync = new CloudSync({
  provider: 'auto-detect',
  syncSettings: true,
  syncRecentFiles: true
});
```

## Migration Guide

### From Web Viewer

To migrate from web-only usage to desktop:

1. **Install desktop application**
2. **Import existing settings** (if any)
3. **Set up file associations**
4. **Configure preferences**

### From Other Viewers

To integrate with existing document viewers:

1. **Export documents to LIV format**
2. **Use conversion tools** (existing CLI)
3. **Import into desktop application**

## Troubleshooting Integration

### Common Integration Issues

1. **Viewer Process Not Starting**
   - Check Go viewer executable path
   - Verify system permissions
   - Check port availability

2. **File Association Problems**
   - Run installer as administrator
   - Check system file type settings
   - Verify MIME type registration

3. **Performance Issues**
   - Check existing performance metrics
   - Review WASM memory usage
   - Monitor system resources

### Debug Integration

Enable comprehensive debugging:

```bash
# Debug desktop application
DEBUG=* npm run dev

# Debug Go viewer
go run cmd/viewer/main.go --debug

# Debug WASM engine
LIV_WASM_DEBUG=1 npm run dev
```

## Best Practices

### Development

1. **Use existing APIs** whenever possible
2. **Follow security model** established by core system
3. **Maintain compatibility** with web viewer
4. **Test cross-platform** functionality

### Deployment

1. **Bundle all dependencies** including Go viewer
2. **Sign applications** for security
3. **Test file associations** on all platforms
4. **Verify auto-update** functionality

### Maintenance

1. **Keep Electron updated** for security
2. **Monitor Go viewer** compatibility
3. **Test with new LIV formats** as they're added
4. **Update documentation** as features change

This integration guide ensures the desktop application works seamlessly with the existing LIV ecosystem while providing enhanced desktop-specific functionality.