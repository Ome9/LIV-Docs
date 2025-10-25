/* ===================================================================
   LIV Professional Editor - Main JavaScript
   Features: Anime.js animations, Drag & Drop, Backend Integration
   =================================================================== */

// ===== State Management =====
const state = {
  elementIdCounter: 3,
  currentPage: 1,
  zoom: 100,
  history: [],
  historyIndex: -1,
  currentDocumentPath: null,
  isDirty: false,
  selectedElement: null,
};

// ===== Initialization =====
document.addEventListener('DOMContentLoaded', () => {
  initializeEditor();
  setupEventListeners();
  setupDragAndDrop();
  setupKeyboardShortcuts();
  animateInitialLoad();
  updateStatus();
});

function initializeEditor() {
  console.log('🚀 LIV Professional Editor initialized');
  
  // Check for backend connection
  if (window.electronAPI) {
    console.log('✅ Backend connection established');
    setupBackendListeners();
  } else {
    console.warn('⚠️ Running in browser mode - limited functionality');
  }
  
  // Load user preferences
  loadPreferences();
  
  // Initialize undo/redo
  saveState();
}

// ===== Event Listeners Setup =====
function setupEventListeners() {
  // Navbar actions
  document.getElementById('newDocBtn').addEventListener('click', handleNewDocument);
  document.getElementById('openDocBtn').addEventListener('click', handleOpenDocument);
  document.getElementById('saveDocBtn').addEventListener('click', handleSaveDocument);
  document.getElementById('exportBtn').addEventListener('click', handleExport);
  
  // Zoom controls
  document.getElementById('zoomIn').addEventListener('click', () => adjustZoom(10));
  document.getElementById('zoomOut').addEventListener('click', () => adjustZoom(-10));
  
  // Undo/Redo
  document.getElementById('undoBtn').addEventListener('click', undo);
  document.getElementById('redoBtn').addEventListener('click', redo);
  
  // Settings
  document.getElementById('settingsBtn').addEventListener('click', openSettings);
  
  // Sidebar toggles
  document.getElementById('toggleLeftSidebar').addEventListener('click', () => toggleSidebar('left'));
  document.getElementById('toggleRightSidebar').addEventListener('click', () => toggleSidebar('right'));
  
  // Tab switching
  document.querySelectorAll('.tab-btn').forEach(btn => {
    btn.addEventListener('click', () => switchTab(btn.dataset.tab));
  });
  
  // Formatting toolbar
  setupFormattingToolbar();
  
  // PDF tools
  setupPDFTools();
  
  // Element toolbar buttons (event delegation)
  document.getElementById('docPage').addEventListener('click', handleElementAction);
  
  // Content change detection
  document.getElementById('docPage').addEventListener('input', () => {
    state.isDirty = true;
    updateStatus();
  });
  
  // Properties panel
  setupPropertiesPanel();
  
  // Modal handlers
  document.getElementById('closeModal').addEventListener('click', closeModal);
  document.getElementById('cancelSettingsBtn').addEventListener('click', closeModal);
  document.getElementById('saveSettingsBtn').addEventListener('click', saveSettings);
}

// ===== Drag and Drop =====
function setupDragAndDrop() {
  const components = document.querySelectorAll('.component-card');
  const docPage = document.getElementById('docPage');
  
  components.forEach(component => {
    component.addEventListener('dragstart', handleDragStart);
  });
  
  docPage.addEventListener('dragover', handleDragOver);
  docPage.addEventListener('drop', handleDrop);
  docPage.addEventListener('dragleave', handleDragLeave);
  
  // Make elements reorderable with Sortable.js
  new Sortable(docPage, {
    animation: 200,
    handle: '.element-content',
    ghostClass: 'sortable-ghost',
    chosenClass: 'sortable-chosen',
    dragClass: 'sortable-drag',
    onEnd: (evt) => {
      saveState();
      updateStatus();
      showToast('Element moved', 'success');
    }
  });
}

let draggedType = null;

function handleDragStart(e) {
  draggedType = e.target.dataset.type || e.target.closest('.component-card').dataset.type;
  e.dataTransfer.effectAllowed = 'copy';
  e.dataTransfer.setData('text/html', e.target.innerHTML);
  
  // Animate drag start
  anime({
    targets: e.target,
    scale: 0.95,
    opacity: 0.7,
    duration: 150,
    easing: 'easeOutQuad'
  });
}

function handleDragOver(e) {
  e.preventDefault();
  e.dataTransfer.dropEffect = 'copy';
  
  const docPage = e.currentTarget;
  docPage.style.borderColor = 'rgb(59, 130, 246)';
  docPage.style.background = 'rgba(59, 130, 246, 0.05)';
}

function handleDragLeave(e) {
  const docPage = e.currentTarget;
  docPage.style.borderColor = '';
  docPage.style.background = '';
}

function handleDrop(e) {
  e.preventDefault();
  
  const docPage = e.currentTarget;
  docPage.style.borderColor = '';
  docPage.style.background = '';
  
  if (draggedType) {
    addElement(draggedType);
    draggedType = null;
    
    // Animate new element
    const newElement = docPage.lastElementChild;
    anime({
      targets: newElement,
      translateY: [-20, 0],
      opacity: [0, 1],
      duration: 400,
      easing: 'easeOutExpo'
    });
  }
}

// ===== Element Management =====
function addElement(type) {
  const docPage = document.getElementById('docPage');
  const elementId = state.elementIdCounter++;
  
  const element = document.createElement('div');
  element.className = 'doc-element';
  element.dataset.id = elementId;
  element.dataset.type = type;
  
  const toolbar = createElementToolbar();
  let content = '';
  
  switch(type) {
    case 'h1':
      content = '<h1 class="heading h1" contenteditable="true" spellcheck="true">Heading 1</h1>';
      break;
    case 'h2':
      content = '<h2 class="heading h2" contenteditable="true" spellcheck="true">Heading 2</h2>';
      break;
    case 'h3':
      content = '<h3 class="heading h3" contenteditable="true" spellcheck="true">Heading 3</h3>';
      break;
    case 'paragraph':
      content = '<p class="paragraph" contenteditable="true" spellcheck="true">Write your text here...</p>';
      break;
    case 'quote':
      content = '<blockquote class="quote" contenteditable="true" spellcheck="true">"Your inspiring quote goes here"</blockquote>';
      break;
    case 'list':
      content = '<ul class="list-block"><li contenteditable="true">List item 1</li><li contenteditable="true">List item 2</li><li contenteditable="true">List item 3</li></ul>';
      break;
    case 'callout-info':
      content = '<div class="callout info" contenteditable="true" spellcheck="true">ℹ️ <strong>Info:</strong> Add your information here</div>';
      break;
    case 'callout-warning':
      content = '<div class="callout warning" contenteditable="true" spellcheck="true">⚠️ <strong>Warning:</strong> Add your warning here</div>';
      break;
    case 'callout-success':
      content = '<div class="callout success" contenteditable="true" spellcheck="true">✓ <strong>Success:</strong> Operation completed successfully</div>';
      break;
    case 'code':
      content = '<pre class="code-block" contenteditable="true" spellcheck="false">// Write your code here\nfunction example() {\n  console.log("Hello World");\n}</pre>';
      break;
    case 'divider':
      content = '<hr class="divider">';
      break;
    case 'spacer':
      content = '<div style="height: 2rem;"></div>';
      break;
    case 'tag':
      content = '<div><span class="tag" contenteditable="true">Tag</span></div>';
      break;
    case 'table':
      content = '<table style="width:100%;border-collapse:collapse;margin:1rem 0;"><thead><tr><th style="border:1px solid #ddd;padding:0.5rem;background:#f5f5f5;">Header 1</th><th style="border:1px solid #ddd;padding:0.5rem;background:#f5f5f5;">Header 2</th></tr></thead><tbody><tr><td contenteditable="true" style="border:1px solid #ddd;padding:0.5rem;">Cell 1</td><td contenteditable="true" style="border:1px solid #ddd;padding:0.5rem;">Cell 2</td></tr></tbody></table>';
      break;
    case 'image':
      content = '<div style="text-align:center;padding:2rem;background:#f5f5f5;border:2px dashed #ddd;border-radius:8px;margin:1rem 0;"><p style="color:#888;">Click to upload image</p><button class="btn btn-primary" style="margin-top:1rem;" onclick="selectImage(this)">Choose Image</button></div>';
      break;
    default:
      content = '<p class="paragraph" contenteditable="true" spellcheck="true">New element</p>';
  }
  
  element.innerHTML = `<div class="element-content">${content}</div>${toolbar}`;
  docPage.appendChild(element);
  
  saveState();
  updateStatus();
  showToast('Element added', 'success');
  
  return element;
}

function createElementToolbar() {
  return `
    <div class="element-toolbar">
      <button class="element-btn" data-action="moveUp" title="Move Up">
        <svg class="icon" width="14" height="14" viewBox="0 0 24 24" fill="none" stroke="currentColor" stroke-width="2">
          <polyline points="18 15 12 9 6 15"></polyline>
        </svg>
      </button>
      <button class="element-btn" data-action="moveDown" title="Move Down">
        <svg class="icon" width="14" height="14" viewBox="0 0 24 24" fill="none" stroke="currentColor" stroke-width="2">
          <polyline points="6 9 12 15 18 9"></polyline>
        </svg>
      </button>
      <button class="element-btn" data-action="duplicate" title="Duplicate">
        <svg class="icon" width="14" height="14" viewBox="0 0 24 24" fill="none" stroke="currentColor" stroke-width="2">
          <rect x="9" y="9" width="13" height="13" rx="2" ry="2"></rect>
          <path d="M5 15H4a2 2 0 0 1-2-2V4a2 2 0 0 1 2-2h9a2 2 0 0 1 2 2v1"></path>
        </svg>
      </button>
      <button class="element-btn danger" data-action="delete" title="Delete">
        <svg class="icon" width="14" height="14" viewBox="0 0 24 24" fill="none" stroke="currentColor" stroke-width="2">
          <line x1="18" y1="6" x2="6" y2="18"></line>
          <line x1="6" y1="6" x2="18" y2="18"></line>
        </svg>
      </button>
    </div>
  `;
}

function handleElementAction(e) {
  const btn = e.target.closest('.element-btn');
  if (!btn) return;
  
  const element = btn.closest('.doc-element');
  const action = btn.dataset.action;
  
  switch(action) {
    case 'moveUp':
      moveElementUp(element);
      break;
    case 'moveDown':
      moveElementDown(element);
      break;
    case 'duplicate':
      duplicateElement(element);
      break;
    case 'delete':
      deleteElement(element);
      break;
  }
}

function moveElementUp(element) {
  const prev = element.previousElementSibling;
  if (prev && prev.classList.contains('doc-element')) {
    element.parentNode.insertBefore(element, prev);
    animateElement(element, 'bounce');
    saveState();
    updateStatus();
  }
}

function moveElementDown(element) {
  const next = element.nextElementSibling;
  if (next && next.classList.contains('doc-element')) {
    element.parentNode.insertBefore(next, element);
    animateElement(element, 'bounce');
    saveState();
    updateStatus();
  }
}

function duplicateElement(element) {
  const clone = element.cloneNode(true);
  clone.dataset.id = state.elementIdCounter++;
  element.parentNode.insertBefore(clone, element.nextElementSibling);
  
  anime({
    targets: clone,
    scale: [0.8, 1],
    opacity: [0, 1],
    duration: 300,
    easing: 'easeOutBack'
  });
  
  saveState();
  updateStatus();
  showToast('Element duplicated', 'success');
}

function deleteElement(element) {
  if (!confirm('Delete this element?')) return;
  
  anime({
    targets: element,
    scale: 0.8,
    opacity: 0,
    duration: 200,
    easing: 'easeInQuad',
    complete: () => {
      element.remove();
      saveState();
      updateStatus();
      showToast('Element deleted', 'info');
    }
  });
}

// ===== Animations =====
function animateElement(element, type = 'bounce') {
  switch(type) {
    case 'bounce':
      anime({
        targets: element,
        translateY: [-10, 0],
        duration: 400,
        easing: 'easeOutBounce'
      });
      break;
    case 'fade':
      anime({
        targets: element,
        opacity: [0, 1],
        duration: 300,
        easing: 'easeOutQuad'
      });
      break;
    case 'scale':
      anime({
        targets: element,
        scale: [0.95, 1],
        duration: 200,
        easing: 'easeOutQuad'
      });
      break;
  }
}

function animateInitialLoad() {
  // Animate navbar
  anime({
    targets: '.navbar',
    translateY: [-20, 0],
    opacity: [0, 1],
    duration: 500,
    easing: 'easeOutExpo'
  });
  
  // Animate sidebars
  anime({
    targets: '.sidebar-left',
    translateX: [-50, 0],
    opacity: [0, 1],
    duration: 500,
    delay: 100,
    easing: 'easeOutExpo'
  });
  
  anime({
    targets: '.sidebar-right',
    translateX: [50, 0],
    opacity: [0, 1],
    duration: 500,
    delay: 100,
    easing: 'easeOutExpo'
  });
  
  // Animate document elements
  anime({
    targets: '.doc-element',
    translateY: [20, 0],
    opacity: [0, 1],
    duration: 600,
    delay: anime.stagger(100, {start: 300}),
    easing: 'easeOutExpo'
  });
}

// ===== Document Operations =====
async function handleNewDocument() {
  if (state.isDirty) {
    const confirm = await showConfirmDialog('Create new document?', 'Unsaved changes will be lost.');
    if (!confirm) return;
  }
  
  // Clear document
  const docPage = document.getElementById('docPage');
  docPage.innerHTML = '';
  
  // Add default elements
  addElement('h1');
  addElement('paragraph');
  
  // Reset state
  state.currentDocumentPath = null;
  state.isDirty = false;
  state.elementIdCounter = 3;
  state.history = [];
  state.historyIndex = -1;
  
  // Clear properties
  document.getElementById('propTitle').value = '';
  document.getElementById('propAuthor').value = '';
  document.getElementById('propVersion').value = '1.0.0';
  document.getElementById('propDescription').value = '';
  
  updateStatus();
  showToast('New document created', 'success');
  
  // Animate
  animateInitialLoad();
}

async function handleOpenDocument() {
  if (!window.electronAPI) {
    showToast('Open function requires desktop app', 'warning');
    return;
  }
  
  try {
    const result = await window.electronAPI.showOpenDialog({
      title: 'Open LIV Document',
      filters: [
        { name: 'LIV Documents', extensions: ['liv'] },
        { name: 'All Files', extensions: ['*'] }
      ],
      properties: ['openFile']
    });
    
    if (result.canceled || !result.filePaths.length) return;
    
    const filePath = result.filePaths[0];
    const doc = await window.electronAPI.openDocument(filePath);
    
    loadDocumentData(doc);
    state.currentDocumentPath = filePath;
    state.isDirty = false;
    
    updateStatus();
    showToast('Document opened successfully', 'success');
    
  } catch (error) {
    console.error('Error opening document:', error);
    showToast(`Error: ${error.message}`, 'error');
  }
}

async function handleSaveDocument() {
  if (!window.electronAPI) {
    showToast('Save function requires desktop app', 'warning');
    return;
  }
  
  try {
    let filePath = state.currentDocumentPath;
    
    if (!filePath) {
      const result = await window.electronAPI.showSaveDialog({
        title: 'Save LIV Document',
        defaultPath: 'document.liv',
        filters: [
          { name: 'LIV Documents', extensions: ['liv'] }
        ]
      });
      
      if (result.canceled) return;
      filePath = result.filePath;
    }
    
    const docData = buildDocumentData(filePath);
    await window.electronAPI.saveDocument(docData);
    
    state.currentDocumentPath = filePath;
    state.isDirty = false;
    
    updateStatus();
    showToast('Document saved successfully', 'success');
    
  } catch (error) {
    console.error('Error saving document:', error);
    showToast(`Error: ${error.message}`, 'error');
  }
}

function buildDocumentData(filePath) {
  const elements = document.querySelectorAll('.doc-element');
  const title = document.getElementById('propTitle').value || 'Untitled Document';
  const docType = document.getElementById('propType').value || 'static';
  
  // Build clean HTML
  let html = `<!DOCTYPE html><html lang="en"><head>
<meta charset="UTF-8">
<meta name="viewport" content="width=device-width,initial-scale=1.0">
<title>${title}</title>
<style>
body{font-family:system-ui,sans-serif;padding:48px;max-width:900px;margin:0 auto;line-height:1.7;color:#1a1a1a}
.heading{font-weight:700;margin:32px 0 16px;line-height:1.2}
.h1{font-size:2.25rem}.h2{font-size:1.875rem}.h3{font-size:1.5rem}
.paragraph{font-size:1rem;margin-bottom:20px;line-height:1.8}
.callout{padding:20px;border-radius:12px;border-left:5px solid;margin:20px 0}
.info{background:#eff6ff;border-color:#3b82f6;color:#1e3a8a}
.warning{background:#fef3c7;border-color:#f59e0b;color:#78350f}
.success{background:#d1fae5;border-color:#10b981;color:#065f46}
.code-block{background:#1a1a1a;color:#e5e5e5;padding:20px;border-radius:12px;font-family:monospace;overflow-x:auto;margin:20px 0}
.tag{display:inline-block;padding:6px 16px;background:#3b82f6;color:#fff;border-radius:20px;font-size:13px;margin:4px}
.divider{height:2px;background:linear-gradient(90deg,transparent,#e5e7eb,transparent);margin:32px 0;border:none}
.quote{font-size:1.125rem;font-style:italic;padding-left:1.5rem;border-left:4px solid #3b82f6;margin:1.5rem 0}
.list-block{margin:1rem 0;padding-left:1.5rem}.list-block li{margin:0.5rem 0}
</style>
</head><body>`;
  
  elements.forEach(el => {
    const content = el.querySelector('.element-content');
    if (content) {
      let html = content.innerHTML;
      // Clean up: remove contenteditable attributes
      html = html.replace(/\s*contenteditable="[^"]*"/g, '');
      html += '\n';
      return html;
    }
  });
  
  html += '</body></html>';
  
  return {
    filePath,
    metadata: {
      title: title,
      author: document.getElementById('propAuthor').value || 'Unknown',
      version: document.getElementById('propVersion').value || '1.0.0',
      description: document.getElementById('propDescription').value || '',
      language: document.getElementById('propLanguage').value || 'en',
      created: new Date().toISOString(),
      modified: new Date().toISOString()
    },
    content: {
      html: html,
      css: '',
      static: html,
      js: ''
    },
    features: {
      animations: document.getElementById('propAnimations').checked,
      interactivity: document.getElementById('propInteractivity').checked,
      charts: document.getElementById('propCharts').checked,
      forms: document.getElementById('propForms').checked
    },
    documentType: docType
  };
}

function loadDocumentData(doc) {
  if (doc.metadata) {
    document.getElementById('propTitle').value = doc.metadata.title || '';
    document.getElementById('propAuthor').value = doc.metadata.author || '';
    document.getElementById('propVersion').value = doc.metadata.version || '1.0.0';
    document.getElementById('propDescription').value = doc.metadata.description || '';
    document.getElementById('propLanguage').value = doc.metadata.language || 'en';
  }
  
  if (doc.features) {
    document.getElementById('propAnimations').checked = doc.features.animations || false;
    document.getElementById('propInteractivity').checked = doc.features.interactivity || false;
    document.getElementById('propCharts').checked = doc.features.charts || false;
    document.getElementById('propForms').checked = doc.features.forms || false;
  }
  
  if (doc.documentType) {
    document.getElementById('propType').value = doc.documentType;
  }
  
  // Parse and load content
  if (doc.content && doc.content.html) {
    const parser = new DOMParser();
    const htmlDoc = parser.parseFromString(doc.content.html, 'text/html');
    const bodyContent = htmlDoc.body.innerHTML;
    
    const docPage = document.getElementById('docPage');
    docPage.innerHTML = bodyContent;
    
    // Re-add toolbars to elements
    docPage.querySelectorAll('.doc-element').forEach((el, idx) => {
      if (!el.querySelector('.element-toolbar')) {
        const content = el.innerHTML;
        el.innerHTML = `<div class="element-content">${content}</div>${createElementToolbar()}`;
        el.dataset.id = idx + 1;
      }
    });
    
    state.elementIdCounter = docPage.querySelectorAll('.doc-element').length + 1;
  }
}

// ===== Export Functions =====
async function handleExport() {
  const exportMenu = [
    { label: 'Export as HTML', action: () => exportAsHTML() },
    { label: 'Export as PDF', action: () => exportAsPDF() },
    { label: 'Export as EPUB', action: () => exportAsEPUB() },
    { label: 'Export as Markdown', action: () => exportAsMarkdown() }
  ];
  
  // Show export menu (simplified - in production use proper dropdown)
  const choice = prompt('Export as:\n1. HTML\n2. PDF\n3. EPUB\n4. Markdown\n\nEnter number:');
  
  switch(choice) {
    case '1': await exportAsHTML(); break;
    case '2': await exportAsPDF(); break;
    case '3': await exportAsEPUB(); break;
    case '4': await exportAsMarkdown(); break;
    default: showToast('Export cancelled', 'info');
  }
}

async function exportAsHTML() {
  if (!window.electronAPI) {
    showToast('Export requires desktop app', 'warning');
    return;
  }
  
  try {
    const result = await window.electronAPI.showSaveDialog({
      title: 'Export as HTML',
      defaultPath: 'document.html',
      filters: [{ name: 'HTML Files', extensions: ['html'] }]
    });
    
    if (result.canceled) return;
    
    const docData = buildDocumentData(result.filePath);
    await window.electronAPI.exportDocument({
      format: 'html',
      path: result.filePath,
      content: docData.content.html
    });
    
    showToast('Exported as HTML successfully', 'success');
    
  } catch (error) {
    console.error('Error exporting:', error);
    showToast(`Error: ${error.message}`, 'error');
  }
}

async function exportAsPDF() {
  showToast('PDF export coming soon', 'info');
}

async function exportAsEPUB() {
  showToast('EPUB export coming soon', 'info');
}

async function exportAsMarkdown() {
  showToast('Markdown export coming soon', 'info');
}

// ===== Formatting Toolbar =====
function setupFormattingToolbar() {
  document.getElementById('boldBtn').addEventListener('click', () => document.execCommand('bold'));
  document.getElementById('italicBtn').addEventListener('click', () => document.execCommand('italic'));
  document.getElementById('underlineBtn').addEventListener('click', () => document.execCommand('underline'));
  document.getElementById('strikeBtn').addEventListener('click', () => document.execCommand('strikethrough'));
  
  document.getElementById('alignLeftBtn').addEventListener('click', () => document.execCommand('justifyLeft'));
  document.getElementById('alignCenterBtn').addEventListener('click', () => document.execCommand('justifyCenter'));
  document.getElementById('alignRightBtn').addEventListener('click', () => document.execCommand('justifyRight'));
  document.getElementById('alignJustifyBtn').addEventListener('click', () => document.execCommand('justifyFull'));
  
  document.getElementById('textColorPicker').addEventListener('change', (e) => {
    document.execCommand('foreColor', false, e.target.value);
  });
  
  document.getElementById('bgColorPicker').addEventListener('change', (e) => {
    document.execCommand('hiliteColor', false, e.target.value);
  });
  
  document.getElementById('fontFamily').addEventListener('change', (e) => {
    document.execCommand('fontName', false, e.target.value);
  });
  
  document.getElementById('fontSize').addEventListener('change', (e) => {
    const size = e.target.value;
    document.execCommand('fontSize', false, '7');
    const fontElements = document.getElementsByTagName('font');
    for (let i = 0; i < fontElements.length; i++) {
      if (fontElements[i].size == '7') {
        fontElements[i].removeAttribute('size');
        fontElements[i].style.fontSize = size + 'px';
      }
    }
  });
}

// ===== PDF Tools Setup =====
function setupPDFTools() {
  document.getElementById('deletePageBtn').addEventListener('click', deletePage);
  document.getElementById('duplicatePageBtn').addEventListener('click', duplicatePage);
  document.getElementById('rotatePageBtn').addEventListener('click', rotatePage);
  document.getElementById('mergeBtn').addEventListener('click', () => showToast('Merge feature coming soon', 'info'));
  document.getElementById('splitBtn').addEventListener('click', () => showToast('Split feature coming soon', 'info'));
  document.getElementById('compressBtn').addEventListener('click', () => showToast('Compress feature coming soon', 'info'));
  document.getElementById('watermarkBtn').addEventListener('click', () => showToast('Watermark feature coming soon', 'info'));
  document.getElementById('signBtn').addEventListener('click', () => showToast('Digital signature coming soon', 'info'));
  document.getElementById('encryptBtn').addEventListener('click', () => showToast('Encryption coming soon', 'info'));
  document.getElementById('redactBtn').addEventListener('click', () => showToast('Redaction coming soon', 'info'));
}

function deletePage() {
  if (confirm('Delete current page?')) {
    showToast('Page deleted', 'success');
    // Implementation
  }
}

function duplicatePage() {
  showToast('Page duplicated', 'success');
  // Implementation
}

function rotatePage() {
  const docPage = document.getElementById('docPage');
  const currentRotation = parseInt(docPage.dataset.rotation || '0');
  const newRotation = (currentRotation + 90) % 360;
  docPage.dataset.rotation = newRotation;
  docPage.style.transform = `rotate(${newRotation}deg)`;
  showToast('Page rotated', 'success');
}

// ===== Utility Functions =====
function adjustZoom(delta) {
  state.zoom = Math.max(50, Math.min(200, state.zoom + delta));
  document.getElementById('zoomLevel').textContent = state.zoom + '%';
  document.getElementById('canvasInner').style.transform = `scale(${state.zoom / 100})`;
}

function toggleSidebar(side) {
  const sidebar = document.getElementById(side === 'left' ? 'leftSidebar' : 'rightSidebar');
  sidebar.classList.toggle('collapsed');
  
  anime({
    targets: sidebar,
    duration: 300,
    easing: 'easeOutExpo'
  });
}

function switchTab(tabName) {
  document.querySelectorAll('.tab-btn').forEach(btn => btn.classList.remove('active'));
  document.querySelectorAll('.tab-content').forEach(content => content.classList.remove('active'));
  
  document.querySelector(`[data-tab="${tabName}"]`).classList.add('active');
  document.getElementById(`${tabName}-tab`).classList.add('active');
}

function updateStatus() {
  const elements = document.querySelectorAll('.doc-element');
  const elementCount = elements.length;
  
  let wordCount = 0;
  elements.forEach(el => {
    const text = el.textContent || '';
    wordCount += text.trim().split(/\s+/).filter(w => w.length > 0).length;
  });
  
  document.getElementById('pageCount').textContent = '1 page';
  document.getElementById('elementCount').textContent = `${elementCount} element${elementCount !== 1 ? 's' : ''}`;
  document.getElementById('wordCount').textContent = `${wordCount} word${wordCount !== 1 ? 's' : ''}`;
  document.getElementById('statusText').textContent = state.isDirty ? 'Modified' : 'Ready';
  
  if (state.currentDocumentPath) {
    const fileName = state.currentDocumentPath.split(/[/\\]/).pop();
    document.title = `${fileName} - LIV Editor Pro`;
  } else {
    document.title = 'LIV Editor Pro';
  }
}

function setupPropertiesPanel() {
  const inputs = ['propTitle', 'propAuthor', 'propVersion', 'propDescription', 'propLanguage', 'propType'];
  inputs.forEach(id => {
    document.getElementById(id).addEventListener('change', () => {
      state.isDirty = true;
      updateStatus();
    });
  });
  
  const checkboxes = ['propAnimations', 'propInteractivity', 'propCharts', 'propForms'];
  checkboxes.forEach(id => {
    document.getElementById(id).addEventListener('change', () => {
      state.isDirty = true;
      updateStatus();
    });
  });
}

function setupKeyboardShortcuts() {
  document.addEventListener('keydown', (e) => {
    if (e.ctrlKey || e.metaKey) {
      switch(e.key.toLowerCase()) {
        case 'n':
          e.preventDefault();
          handleNewDocument();
          break;
        case 'o':
          e.preventDefault();
          handleOpenDocument();
          break;
        case 's':
          e.preventDefault();
          handleSaveDocument();
          break;
        case 'z':
          e.preventDefault();
          if (e.shiftKey) {
            redo();
          } else {
            undo();
          }
          break;
        case 'y':
          e.preventDefault();
          redo();
          break;
      }
    }
  });
}

// ===== Undo/Redo =====
function saveState() {
  const docPage = document.getElementById('docPage');
  const newState = {
    html: docPage.innerHTML,
    elementCounter: state.elementIdCounter
  };
  
  // Remove future states if we're not at the end
  state.history = state.history.slice(0, state.historyIndex + 1);
  state.history.push(newState);
  state.historyIndex++;
  
  // Limit history size
  if (state.history.length > 50) {
    state.history.shift();
    state.historyIndex--;
  }
}

function undo() {
  if (state.historyIndex > 0) {
    state.historyIndex--;
    restoreState(state.history[state.historyIndex]);
    showToast('Undo', 'info');
  }
}

function redo() {
  if (state.historyIndex < state.history.length - 1) {
    state.historyIndex++;
    restoreState(state.history[state.historyIndex]);
    showToast('Redo', 'info');
  }
}

function restoreState(savedState) {
  const docPage = document.getElementById('docPage');
  docPage.innerHTML = savedState.html;
  state.elementIdCounter = savedState.elementCounter;
  updateStatus();
}

// ===== Toast Notifications =====
function showToast(message, type = 'info') {
  const container = document.getElementById('toastContainer');
  const toast = document.createElement('div');
  toast.className = `toast ${type}`;
  
  const icons = {
    success: '✓',
    error: '✕',
    warning: '⚠',
    info: 'ℹ'
  };
  
  toast.innerHTML = `
    <span style="font-size:1.25rem;margin-right:0.5rem;">${icons[type] || 'ℹ'}</span>
    <span>${message}</span>
  `;
  
  container.appendChild(toast);
  
  anime({
    targets: toast,
    translateX: [100, 0],
    opacity: [0, 1],
    duration: 300,
    easing: 'easeOutExpo'
  });
  
  setTimeout(() => {
    anime({
      targets: toast,
      translateX: [0, 100],
      opacity: [1, 0],
      duration: 300,
      easing: 'easeInExpo',
      complete: () => toast.remove()
    });
  }, 3000);
}

// ===== Modal Functions =====
function openSettings() {
  document.getElementById('modalOverlay').classList.remove('hidden');
  document.getElementById('modalOverlay').style.display = 'flex';
  
  anime({
    targets: '#modalOverlay',
    opacity: [0, 1],
    duration: 200,
    easing: 'easeOutQuad'
  });
  
  anime({
    targets: '.modal',
    scale: [0.9, 1],
    opacity: [0, 1],
    duration: 300,
    easing: 'easeOutBack'
  });
}

function closeModal() {
  anime({
    targets: '#modalOverlay',
    opacity: [1, 0],
    duration: 200,
    easing: 'easeInQuad',
    complete: () => {
      document.getElementById('modalOverlay').style.display = 'none';
      document.getElementById('modalOverlay').classList.add('hidden');
    }
  });
}

function saveSettings() {
  const theme = document.getElementById('themeSetting').value;
  const autoSave = document.getElementById('autoSave').checked;
  const spellCheck = document.getElementById('spellCheck').checked;
  
  // Apply theme
  if (theme === 'light') {
    document.body.classList.remove('dark-theme');
    document.body.classList.add('light-theme');
  } else if (theme === 'dark') {
    document.body.classList.remove('light-theme');
    document.body.classList.add('dark-theme');
  }
  
  // Save to localStorage
  localStorage.setItem('theme', theme);
  localStorage.setItem('autoSave', autoSave);
  localStorage.setItem('spellCheck', spellCheck);
  
  closeModal();
  showToast('Settings saved', 'success');
}

function loadPreferences() {
  const theme = localStorage.getItem('theme') || 'dark';
  const autoSave = localStorage.getItem('autoSave') !== 'false';
  const spellCheck = localStorage.getItem('spellCheck') !== 'false';
  
  if (theme === 'light') {
    document.body.classList.add('light-theme');
  } else {
    document.body.classList.add('dark-theme');
  }
  
  document.getElementById('themeSetting').value = theme;
  document.getElementById('autoSave').checked = autoSave;
  document.getElementById('spellCheck').checked = spellCheck;
}

// ===== Backend Integration =====
function setupBackendListeners() {
  if (!window.electronAPI) return;
  
  // Listen for document load from main process
  window.electronAPI.onLoadDocument((doc) => {
    loadDocumentData(doc);
    state.currentDocumentPath = doc.filePath;
    state.isDirty = false;
    updateStatus();
    showToast('Document loaded', 'success');
  });
  
  // Auto-save if enabled
  setInterval(() => {
    if (state.isDirty && localStorage.getItem('autoSave') !== 'false' && state.currentDocumentPath) {
      handleSaveDocument();
    }
  }, 60000); // Every minute
}

function showConfirmDialog(title, message) {
  return new Promise((resolve) => {
    const result = confirm(`${title}\n\n${message}`);
    resolve(result);
  });
}

// ===== Image Upload =====
function selectImage(button) {
  if (!window.electronAPI) {
    showToast('Image upload requires desktop app', 'warning');
    return;
  }
  
  window.electronAPI.showOpenDialog({
    title: 'Select Image',
    filters: [
      { name: 'Images', extensions: ['jpg', 'jpeg', 'png', 'gif', 'svg', 'webp'] }
    ],
    properties: ['openFile']
  }).then(result => {
    if (!result.canceled && result.filePaths.length) {
      const imagePath = result.filePaths[0];
      const img = document.createElement('img');
      img.src = imagePath;
      img.style.maxWidth = '100%';
      img.style.borderRadius = '8px';
      
      const container = button.closest('div');
      container.innerHTML = '';
      container.appendChild(img);
      
      saveState();
      showToast('Image added', 'success');
    }
  });
}

// Export functions for global access
window.selectImage = selectImage;

console.log('✅ LIV Editor fully loaded and ready');
