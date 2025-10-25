# LIV Professional Editor - Modern UI

## 🎨 Overview

A completely redesigned, professional document editor for the LIV format with modern UI components inspired by:
- **shadcn/ui** - Modern, accessible component design
- **Stirling-PDF** - Comprehensive PDF editing tools
- **embed-pdf viewer** - Clean viewing experience

## ✨ Features

### Modern UI Components
- **Drag & Drop Interface** - Intuitive component placement with Sortable.js
- **Anime.js Animations** - Smooth, professional animations throughout
- **shadcn-inspired Design** - Clean, accessible, beautiful components
- **Dark/Light Themes** - System-aware theme switching

### Document Editing
- Rich text formatting toolbar
- Multiple text elements (H1, H2, H3, paragraphs, quotes, lists)
- Content blocks (callouts, code blocks, tables)
- Layout components (dividers, spacers, columns, tags)
- Drag & drop element reordering
- Undo/Redo support
- Real-time word/element counting

### PDF Tools (Stirling-PDF Inspired)
- **Page Operations**: Add, delete, duplicate, rotate pages
- **Document Tools**: Merge, split, compress documents
- **Annotations**: Text, highlight, drawing, stamps
- **Security**: Digital signatures, encryption, redaction
- **Watermarking**: Add text/image watermarks

### Backend Integration
- Full integration with Go builder backend
- Document validation and integrity checking
- Digital signature support
- WASM module execution
- Asset management

## 🚀 Quick Start

### 1. Open the Modern Editor

The editor is available at `desktop/src/editor-modern.html`

```bash
cd desktop
npm start
```

### 2. Using the Editor

**Creating Content:**
- Drag components from the left sidebar onto the canvas
- Click elements to edit inline
- Use the formatting toolbar for text styling
- Access PDF tools from the tools tab

**Keyboard Shortcuts:**
- `Ctrl/Cmd + N` - New document
- `Ctrl/Cmd + O` - Open document
- `Ctrl/Cmd + S` - Save document
- `Ctrl/Cmd + Z` - Undo
- `Ctrl/Cmd + Y` - Redo
- `Ctrl/Cmd + B` - Bold
- `Ctrl/Cmd + I` - Italic
- `Ctrl/Cmd + U` - Underline

**Document Properties:**
- Set title, author, version in right sidebar
- Configure document type (static/interactive)
- Enable features (animations, interactivity, charts, forms)
- Adjust page settings (size, orientation, margins)

### 3. Saving & Exporting

**Save as LIV:**
- Click "Save" button or `Ctrl/Cmd + S`
- Choose location and filename
- Document saved with full metadata

**Export Formats:**
- HTML - Standalone HTML file
- PDF - Print-ready PDF (coming soon)
- EPUB - eBook format (coming soon)
- Markdown - Plain text format (coming soon)

## 📁 File Structure

```
desktop/src/
├── editor-modern.html    # Main editor HTML
├── styles.css           # Comprehensive CSS with design system
├── editor.js            # Main JavaScript with animations
├── preload.js           # Enhanced Electron API bridge
└── main.js              # Electron main process
```

## 🎨 Design System

### Color Palette (Dark Theme)
- **Background**: `rgb(10, 10, 10)`
- **Foreground**: `rgb(229, 229, 229)`
- **Primary**: `rgb(59, 130, 246)` - Blue
- **Accent**: `rgb(139, 92, 246)` - Purple
- **Success**: `rgb(16, 185, 129)` - Green
- **Warning**: `rgb(245, 158, 11)` - Orange
- **Destructive**: `rgb(239, 68, 68)` - Red

### Typography
- **Sans Serif**: Inter, -apple-system, BlinkMacSystemFont, Segoe UI
- **Monospace**: Fira Code, Monaco, Consolas

### Spacing Scale
- `xs`: 0.25rem (4px)
- `sm`: 0.5rem (8px)
- `md`: 1rem (16px)
- `lg`: 1.5rem (24px)
- `xl`: 2rem (32px)

### Border Radius
- Default: 0.5rem (8px)
- Small: 0.25rem (4px)
- Large: 1rem (16px)

## 🔧 Customization

### Changing Themes

Modify CSS variables in `styles.css`:

```css
:root {
  --primary: 59 130 246;  /* Change primary color */
  --accent: 139 92 246;   /* Change accent color */
  --radius: 0.5rem;       /* Change border radius */
}
```

### Adding New Components

1. Add component card to sidebar in `editor-modern.html`:
```html
<div class="component-card" draggable="true" data-type="your-type">
  <div class="component-icon">Icon</div>
  <span class="component-label">Your Component</span>
</div>
```

2. Add handler in `editor.js`:
```javascript
case 'your-type':
  content = '<div class="your-component">...</div>';
  break;
```

3. Add styles in `styles.css`:
```css
.your-component {
  /* Your styles */
}
```

## 🔌 Backend Integration

The editor integrates with your Go backend through Electron IPC:

### Document Operations
```javascript
// Save document
await window.electronAPI.saveDocument(docData);

// Open document
const doc = await window.electronAPI.openDocument(filePath);

// Build document
const result = await window.electronAPI.buildDocument(buildData);
```

### PDF Operations
```javascript
// Merge PDFs
await window.electronAPI.mergePDFs(filePaths, outputPath);

// Compress PDF
await window.electronAPI.compressPDF(inputPath, outputPath, quality);

// Sign PDF
await window.electronAPI.signPDF(filePath, signature, options);
```

### Asset Management
```javascript
// Upload image
const assetUrl = await window.electronAPI.uploadImage(filePath);

// Get asset URL
const url = await window.electronAPI.getAssetURL(assetPath);
```

## 📊 Status Indicators

The status bar displays real-time information:
- **Connection Status**: Shows backend connection state
- **Page Count**: Number of pages in document
- **Element Count**: Total elements on current page
- **Word Count**: Total words in document
- **Save Status**: Last save timestamp

## 🎬 Animations

All animations use Anime.js for smooth, professional effects:

- **Page Load**: Staggered fade-in of UI elements
- **Component Add**: Scale-up and fade-in
- **Element Move**: Bounce effect
- **Element Delete**: Scale-down and fade-out
- **Sidebar Toggle**: Slide animation
- **Modal Open**: Scale and fade-in with backdrop blur
- **Toast Notifications**: Slide-in from right

## 🔐 Security

- **Contextual isolation** enabled in Electron
- **Content Security Policy** enforced
- **No direct Node.js access** from renderer
- **IPC bridge** for controlled backend communication
- **Input sanitization** for all user data

## 🐛 Troubleshooting

### Editor not loading
- Check console for errors (`F12`)
- Ensure all dependencies installed (`npm install`)
- Verify Electron version compatibility

### Animations not working
- Ensure Anime.js CDN is accessible
- Check browser console for errors
- Try disabling browser extensions

### Backend integration issues
- Verify Electron main process is running
- Check IPC handler registration in `main.js`
- Enable debug mode: `localStorage.setItem('debugMode', 'true')`

### PDF tools not responding
- Ensure Go builder is compiled and in `bin/` directory
- Check file permissions
- Verify PDF library dependencies

## 📝 Development Tips

### Hot Reload
Development mode includes hot reload:
```bash
npm run dev
```

### Debug Mode
Enable console logging:
```javascript
localStorage.setItem('debugMode', 'true');
```

### Custom Components
Create reusable component templates in `editor.js`:
```javascript
const componentTemplates = {
  'custom-type': {
    html: '<div class="custom">...</div>',
    icon: '🎨',
    label: 'Custom Component'
  }
};
```

## 🤝 Contributing

To add new features:
1. Add UI components in `editor-modern.html`
2. Add styles in `styles.css`
3. Add logic in `editor.js`
4. Update `preload.js` for backend APIs
5. Test thoroughly with both static and interactive documents

## 📄 License

Part of the LIV Format project. See main project LICENSE for details.

## 🎯 Roadmap

- [ ] Advanced PDF editing (OCR, form filling)
- [ ] Collaborative editing
- [ ] Cloud storage integration
- [ ] Plugin system for custom components
- [ ] Mobile responsive design
- [ ] Accessibility improvements (ARIA labels)
- [ ] Keyboard navigation enhancements
- [ ] Export to more formats (Markdown, DOCX)

## 📞 Support

For issues or questions:
- Check GitHub Issues
- Review documentation in `docs/`
- Enable debug mode for detailed logs

---

**Built with ❤️ using modern web technologies and Go**
