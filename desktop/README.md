# LIV Desktop Viewer

A cross-platform desktop application for viewing Live Interactive Visual (LIV) documents, built with Electron.

## Features

### Core Functionality
- **Native Desktop Experience**: Full desktop integration with native menus, keyboard shortcuts, and file associations
- **Secure Document Viewing**: Sandboxed execution environment for interactive content
- **File Association**: Automatic association with .liv files for double-click opening
- **Recent Files**: Track and quickly access recently opened documents
- **Cross-Platform**: Works on Windows, macOS, and Linux

### Desktop Integration
- **Native Menus**: Platform-specific menu bars with standard shortcuts
- **File Dialogs**: Native file open/save dialogs
- **System Notifications**: Integration with system notification systems
- **Auto-Updates**: Automatic application updates via electron-updater
- **Window Management**: Proper window state persistence and restoration

### Security Features
- **Sandboxed Execution**: All documents run in isolated environment
- **Content Security Policy**: Strict CSP for web content
- **Process Isolation**: Renderer process isolation with context bridge
- **Secure Defaults**: No Node.js integration in renderer process

## Installation

### Prerequisites
- Node.js 16 or later
- npm or yarn package manager

### Development Setup

```bash
# Install dependencies
cd desktop
npm install

# Start in development mode
npm run dev

# Build for production
npm run build

# Build for specific platforms
npm run build:win    # Windows
npm run build:mac    # macOS
npm run build:linux  # Linux
```

### Production Build

```bash
# Create distributable packages
npm run dist

# The built applications will be in the dist/ directory
```

## Architecture

### Process Structure
```
┌─────────────────────────────────────┐
│           Main Process              │
│  ┌─────────────────────────────┐    │
│  │     Electron Main           │    │
│  │   - Window Management       │    │
│  │   - File System Access     │    │
│  │   - Native Integration     │    │
│  │   - Auto Updates           │    │
│  └─────────────────────────────┘    │
└─────────────────────────────────────┘
                  │
                  │ IPC
                  ▼
┌─────────────────────────────────────┐
│         Renderer Process            │
│  ┌─────────────────────────────┐    │
│  │      Web Content            │    │
│  │   - LIV Viewer Interface    │    │
│  │   - Document Rendering      │    │
│  │   - User Interactions       │    │
│  └─────────────────────────────┘    │
└─────────────────────────────────────┘
                  │
                  │ HTTP
                  ▼
┌─────────────────────────────────────┐
│         Viewer Process              │
│  ┌─────────────────────────────┐    │
│  │    Go Web Server            │    │
│  │   - Document Processing     │    │
│  │   - WASM Execution          │    │
│  │   - Security Validation     │    │
│  └─────────────────────────────┘    │
└─────────────────────────────────────┘
```

### Component Integration
- **Electron Main Process**: Manages application lifecycle, native integration
- **Renderer Process**: Displays web-based viewer interface
- **Go Viewer Process**: Handles document processing and WASM execution
- **Context Bridge**: Secure communication between main and renderer processes

## Configuration

### Application Settings
Settings are stored using electron-store in platform-specific locations:

- **Windows**: `%APPDATA%/liv-viewer-desktop/config.json`
- **macOS**: `~/Library/Preferences/liv-viewer-desktop/config.json`
- **Linux**: `~/.config/liv-viewer-desktop/config.json`

### Default Settings
```json
{
  "windowBounds": { "width": 1200, "height": 800 },
  "recentFiles": [],
  "preferences": {
    "theme": "system",
    "autoUpdate": true,
    "fallbackMode": false,
    "debugMode": false
  }
}
```

### Environment Variables
- `NODE_ENV`: Set to 'development' for development mode
- `ELECTRON_IS_DEV`: Alternative development mode flag
- `LIV_VIEWER_PATH`: Custom path to LIV viewer executable

## File Associations

The application automatically registers file associations for `.liv` files:

### Windows
- Registry entries for file association
- Context menu integration
- Icon association

### macOS
- Info.plist configuration
- UTI (Uniform Type Identifier) registration
- Finder integration

### Linux
- Desktop entry file
- MIME type registration
- File manager integration

## Security Model

### Process Isolation
- Main process has full system access
- Renderer process is sandboxed
- No Node.js integration in renderer
- Context bridge for secure IPC

### Content Security
- Strict Content Security Policy
- No eval() or inline scripts
- Secure communication with viewer process
- Document validation before rendering

### Network Security
- Viewer process only binds to localhost
- Random port selection
- No external network access from documents
- CORS headers properly configured

## Development

### Project Structure
```
desktop/
├── src/
│   ├── main.js           # Main process entry point
│   ├── preload.js        # Context bridge script
│   ├── preferences.html  # Preferences dialog
│   └── error.html        # Error page
├── assets/
│   └── icons/           # Application icons
├── package.json         # Dependencies and build config
└── README.md           # This file
```

### Development Commands
```bash
# Start development server
npm run dev

# Enable debugging
DEBUG=* npm run dev

# Run with specific viewer path
LIV_VIEWER_PATH=/path/to/viewer npm run dev

# Build without packaging
npm run pack

# Clean build artifacts
rm -rf dist/ node_modules/.cache/
```

### Debugging
- Enable Developer Tools: `Ctrl+Shift+I` (Windows/Linux) or `Cmd+Alt+I` (macOS)
- Main process debugging: Use `--inspect` flag with Electron
- Renderer process debugging: Standard Chrome DevTools
- IPC debugging: Enable DEBUG environment variable

## Distribution

### Build Configuration
The application uses electron-builder for packaging:

```json
{
  "build": {
    "appId": "com.livformat.viewer",
    "productName": "LIV Viewer",
    "directories": {
      "output": "dist"
    },
    "files": ["src/**/*", "assets/**/*"],
    "extraResources": [
      {
        "from": "../bin/",
        "to": "bin/"
      }
    ]
  }
}
```

### Platform-Specific Builds
- **Windows**: NSIS installer (.exe)
- **macOS**: DMG disk image (.dmg)
- **Linux**: AppImage (.AppImage) and Debian package (.deb)

### Code Signing
For production releases:
- **Windows**: Authenticode signing
- **macOS**: Apple Developer ID signing
- **Linux**: GPG signing for packages

## Troubleshooting

### Common Issues

1. **Viewer Process Fails to Start**
   - Check if Go viewer executable exists
   - Verify executable permissions
   - Check system PATH
   - Review application logs

2. **File Association Not Working**
   - Run installer as administrator (Windows)
   - Check system file associations
   - Verify MIME type registration (Linux)

3. **Auto-Update Issues**
   - Check network connectivity
   - Verify update server configuration
   - Check application permissions

4. **Performance Issues**
   - Enable hardware acceleration
   - Check system resources
   - Review document complexity
   - Enable debug mode for profiling

### Log Files
- **Windows**: `%TEMP%/liv-viewer-desktop.log`
- **macOS**: `~/Library/Logs/liv-viewer-desktop.log`
- **Linux**: `~/.cache/liv-viewer-desktop/logs/main.log`

### Debug Mode
Enable debug mode in preferences or via command line:
```bash
# Enable debug logging
DEBUG=* npm start

# Enable Electron debug
npm start -- --enable-logging --log-level=0
```

## Contributing

### Development Workflow
1. Fork the repository
2. Create a feature branch
3. Make changes and test thoroughly
4. Submit a pull request

### Testing
```bash
# Run application tests
npm test

# Test specific platforms
npm run test:win
npm run test:mac
# npm run test:linux

# Integration tests
npm run test:integration
```

### Code Style
- Use ESLint for JavaScript linting
- Follow Electron security best practices
- Maintain consistent code formatting
- Add JSDoc comments for public APIs

## License

This project is licensed under the MIT License. See the main project LICENSE file for details.

## Support

- **Documentation**: See main project README
- **Issues**: Report bugs on GitHub Issues
- **Discussions**: Use GitHub Discussions for questions
- **Security**: Report security issues privately to maintainers