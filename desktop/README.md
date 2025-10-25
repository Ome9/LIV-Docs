# LIV Desktop Viewer & PDF Editor

A cross-platform desktop application for viewing Live Interactive Visual (LIV) documents and editing PDFs, built with Electron. Features a comprehensive PDF editor with beautiful animations, Google Fonts, and modern UI.

## Features

### Core Functionality
- **Native Desktop Experience**: Full desktop integration with native menus, keyboard shortcuts, and file associations
- **Secure Document Viewing**: Sandboxed execution environment for interactive content
- **Comprehensive PDF Editor**: 25+ PDF operations with modern UI
- **Beautiful Animations**: 10+ animation types powered by Anime.js
- **Google Fonts**: 8 integrated fonts with 12 total options
- **Color Presets**: 42 curated colors from Material Design & Tailwind CSS
- **File Association**: Automatic association with .liv files for double-click opening
- **Recent Files**: Track and quickly access recently opened documents
- **Cross-Platform**: Works on Windows, macOS, and Linux

### PDF Editor Features

#### PDF Operations (25+ Methods)
- **File Operations**: Create, open, save, export PDFs
- **Page Management**: Merge, split, rotate, reorder pages
- **Compression**: Optimize PDF file size
- **Security**: Encrypt/decrypt, add passwords, digital signatures
- **Watermarks**: Add text or image watermarks
- **Stamps**: Custom stamps and seals
- **Forms**: Fill form fields, create interactive forms
- **Annotations**: Add comments, highlights, notes
- **Bookmarks**: Create and manage bookmarks
- **Attachments**: Attach files to PDFs
- **Redaction**: Remove sensitive information
- **QR Codes & Barcodes**: Generate and embed codes

#### Editing Tools
- **Text Tool**: Add and edit text with Google Fonts
  - 8 Google Fonts: Roboto, Open Sans, Lato, Montserrat, Playfair Display, Raleway, Poppins, Inter
  - 4 Standard Fonts: Helvetica, Times New Roman, Courier, Arial
  - 14 Font Sizes: 8px to 72px
  - Color presets with 42 curated colors
  
- **Image Tool**: Insert and manipulate images
  - Drag and drop support
  - Resize and position
  - Rotation and cropping
  
- **Shape Tools**: Draw rectangles, circles, lines
  - Custom colors and borders
  - Fill and stroke options
  
- **Signature Tool**: Add digital signatures
  - Draw or upload signatures
  - Timestamp support
  
- **Stamp Tool**: Quick stamps and seals
  - Custom stamp creation
  - Predefined stamps library

- **Component Library**: Drag & drop components
  - Pre-built elements
  - Custom components
  - Animation support

#### UI/UX Features
- **Modern Dark Theme**: Professional appearance
- **Keyboard Shortcuts**: 60+ customizable shortcuts
  - `Ctrl+Shift+N`: New PDF
  - `Ctrl+O`: Open PDF
  - `Ctrl+S`: Save PDF
  - `Ctrl+Z`: Undo
  - `Ctrl+Y`: Redo
  - `Ctrl++`: Zoom In
  - `Ctrl+-`: Zoom Out
  - `Ctrl+/`: Show shortcuts guide
  - And 50+ more!
  
- **Animations** (10+ types):
  - Tool selection bounce
  - Button pulse effects
  - Zoom stagger animation
  - Page navigation highlight
  - Confetti celebrations
  - Toast notifications
  - Loading animations
  - Modal transitions
  
- **Color Presets Modal**:
  - Material Design colors (16)
  - Tailwind CSS colors (16)
  - Grayscale palette (10)
  - One-click color application
  
- **Real-time Preview**: Instant visual feedback
- **Zoom Controls**: 25% to 200% zoom
- **Page Navigation**: Quick page jumping
- **Drag & Drop**: Elements and components
- **Toast Notifications**: User feedback
- **Loading Indicators**: Progress tracking

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

# Start in development mode (includes PDF editor)
npm run dev

# Or start directly
npm start

# Build for production
npm run build

# Build for specific platforms
npm run build:win    # Windows
npm run build:mac    # macOS
npm run build:linux  # Linux
```

### Using the PDF Editor

1. Launch the application: `npm start`
2. Open PDF Editor: 
   - File → New PDF (or `Ctrl+Shift+N`)
   - File → Open PDF (or `Ctrl+O`)
3. Use the toolbar to select tools:
   - Text tool (`T`) - Add text with Google Fonts
   - Image tool (`I`) - Insert images
   - Shape tools (`R`, `C`, `L`) - Draw shapes
   - Signature tool (`S`) - Add signatures
4. Access color presets:
   - Click the color wheel button in the text formatting toolbar
   - Choose from 42 curated colors
5. Use keyboard shortcuts:
   - Press `Ctrl+/` to see all shortcuts
   - Customize shortcuts in preferences
6. Drag & drop components from the library
7. Export your PDF: File → Save PDF (`Ctrl+S`)

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
│   ├── main.js                  # Main process entry point (22 IPC handlers)
│   ├── preload.js              # Context bridge script
│   ├── preferences.html        # Preferences dialog
│   ├── error.html              # Error page
│   ├── pdf-editor.html         # PDF editor UI (900+ lines)
│   ├── pdf-editor.css          # Editor styles (1,200+ lines)
│   ├── pdf-editor.js           # Editor logic (1,500+ lines)
│   ├── pdf-operations.js       # PDF operations module (25 methods)
│   └── keybindings-manager.js  # Keyboard shortcuts (60+ shortcuts)
├── assets/
│   └── icons/                  # Application icons
├── package.json                # Dependencies and build config
└── README.md                   # This file
```

### Key Files

#### `pdf-editor.html`
- Modern HTML5 structure
- Google Fonts integration (CDN)
- Component library sidebar
- Formatting toolbar
- Color presets modal
- Canvas rendering area
- Zero inline style warnings

#### `pdf-editor.css`
- 1,200+ lines of modern CSS
- Dark theme variables
- Responsive layout
- Animation keyframes
- Color swatch styles (42 colors)
- Smooth transitions
- Safari compatibility (webkit prefixes)

#### `pdf-editor.js`
- 1,500+ lines of JavaScript
- PDFEditor class with 50+ methods
- Tool management system
- Element manipulation
- Animation system (Anime.js)
- Event handlers
- File operations
- Export functionality

#### `pdf-operations.js`
- 25 PDF manipulation methods
- pdf-lib integration
- Merge, split, rotate operations
- Compression and optimization
- Watermark and stamp functions
- Form filling and annotations
- Encryption and signatures
- QR code and barcode generation

#### `keybindings-manager.js`
- 400+ lines of code
- 60+ keyboard shortcuts
- Customizable keybindings
- Conflict detection
- Visual shortcuts guide
- Save/load preferences
- Mousetrap.js integration

#### `main.js` (Updated)
- 22 IPC handlers for PDF operations
- File system integration
- Window management
- Error handling
- Progress tracking
- Security sandboxing

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