// PDF Editor - Main Application Logic
// This file handles all PDF editing functionality with full integration

class PDFEditor {
  constructor() {
    this.currentPDF = null;
    this.currentPage = 1;
    this.totalPages = 0;
    this.zoom = 1.0;
    this.activeTool = 'select';
    this.selectedElement = null;
    this.elements = [];
    this.history = { past: [], future: [] };
    this.keybindingManager = null;
    this.isLoading = false;

    this.init();
  }

  async init() {
    console.log('Initializing PDF Editor...');
    
    // Initialize keybinding manager
    this.keybindingManager = getKeybindingManager();
    this.setupKeyboardShortcuts();
    
    // Setup UI event listeners
    this.setupEventListeners();
    
    // Load shortcuts list
    this.loadShortcutsList();
    
    console.log('PDF Editor initialized successfully');
  }

  setupKeyboardShortcuts() {
    // File operations
    this.keybindingManager.register('file.new', () => this.createNewPDF());
    this.keybindingManager.register('file.open', () => this.openPDF());
    this.keybindingManager.register('file.save', () => this.savePDF());
    this.keybindingManager.register('file.export', () => this.exportPDF());
    this.keybindingManager.register('file.print', () => this.printPDF());

    // Edit operations
    this.keybindingManager.register('edit.undo', () => this.undo());
    this.keybindingManager.register('edit.redo', () => this.redo());
    this.keybindingManager.register('edit.cut', () => this.cut());
    this.keybindingManager.register('edit.copy', () => this.copy());
    this.keybindingManager.register('edit.paste', () => this.paste());
    this.keybindingManager.register('edit.delete', () => this.deleteSelected());
    this.keybindingManager.register('edit.selectAll', () => this.selectAll());
    this.keybindingManager.register('edit.duplicate', () => this.duplicate());

    // View operations
    this.keybindingManager.register('view.zoomIn', () => this.zoomIn());
    this.keybindingManager.register('view.zoomOut', () => this.zoomOut());
    this.keybindingManager.register('view.zoomReset', () => this.zoomReset());
    this.keybindingManager.register('view.fitWidth', () => this.fitWidth());
    this.keybindingManager.register('view.fitPage', () => this.fitPage());
    this.keybindingManager.register('view.fullscreen', () => this.toggleFullscreen());

    // Navigation
    this.keybindingManager.register('nav.nextPage', () => this.nextPage());
    this.keybindingManager.register('nav.prevPage', () => this.prevPage());
    this.keybindingManager.register('nav.firstPage', () => this.firstPage());
    this.keybindingManager.register('nav.lastPage', () => this.lastPage());

    // Tools
    this.keybindingManager.register('tool.select', () => this.setTool('select'));
    this.keybindingManager.register('tool.text', () => this.setTool('text'));
    this.keybindingManager.register('tool.image', () => this.setTool('image'));
    this.keybindingManager.register('tool.shape', () => this.setTool('rectangle'));
    this.keybindingManager.register('tool.pen', () => this.setTool('pen'));
    this.keybindingManager.register('tool.highlighter', () => this.setTool('highlighter'));
    this.keybindingManager.register('tool.eraser', () => this.setTool('eraser'));

    console.log('Keyboard shortcuts registered');
  }

  setupEventListeners() {
    // Add ripple effect to all buttons
    this.setupButtonRipples();

    // Navbar buttons
    document.getElementById('newBtn')?.addEventListener('click', () => this.createNewPDF());
    document.getElementById('openBtn')?.addEventListener('click', () => this.openPDF());
    document.getElementById('saveBtn')?.addEventListener('click', () => this.savePDF());
    document.getElementById('exportBtn')?.addEventListener('click', () => this.exportPDF());
    document.getElementById('undoBtn')?.addEventListener('click', () => this.undo());
    document.getElementById('redoBtn')?.addEventListener('click', () => this.redo());

    // Zoom controls
    document.getElementById('zoomInBtn')?.addEventListener('click', () => this.zoomIn());
    document.getElementById('zoomOutBtn')?.addEventListener('click', () => this.zoomOut());
    document.getElementById('fitWidthBtn')?.addEventListener('click', () => this.fitWidth());
    document.getElementById('fitPageBtn')?.addEventListener('click', () => this.fitPage());

    // Empty state buttons
    document.getElementById('emptyNewBtn')?.addEventListener('click', () => this.createNewPDF());
    document.getElementById('emptyOpenBtn')?.addEventListener('click', () => this.openPDF());

    // PDF Operations
    document.getElementById('mergePdfBtn')?.addEventListener('click', () => this.mergePDFs());
    document.getElementById('splitPdfBtn')?.addEventListener('click', () => this.splitPDF());
    document.getElementById('compressPdfBtn')?.addEventListener('click', () => this.compressPDF());
    document.getElementById('encryptPdfBtn')?.addEventListener('click', () => this.encryptPDF());
    document.getElementById('imagesToPdfBtn')?.addEventListener('click', () => this.imagesToPDF());

    // Tool buttons
    document.querySelectorAll('.tool-btn[data-tool]').forEach(btn => {
      btn.addEventListener('click', (e) => {
        const tool = e.currentTarget.getAttribute('data-tool');
        this.setTool(tool);
      });
    });

    // Sidebar tabs
    document.querySelectorAll('.tab-btn').forEach(btn => {
      btn.addEventListener('click', (e) => {
        const tab = e.currentTarget.getAttribute('data-tab');
        this.switchTab(e.currentTarget.closest('.sidebar'), tab);
      });
    });

    // Formatting toolbar
    document.getElementById('fontSelect')?.addEventListener('change', (e) => this.updateTextFormat('font', e.target.value));
    document.getElementById('fontSizeSelect')?.addEventListener('change', (e) => this.updateTextFormat('size', e.target.value));
    document.getElementById('boldBtn')?.addEventListener('click', () => this.updateTextFormat('bold'));
    document.getElementById('italicBtn')?.addEventListener('click', () => this.updateTextFormat('italic'));
    document.getElementById('underlineBtn')?.addEventListener('click', () => this.updateTextFormat('underline'));
    document.getElementById('textColorPicker')?.addEventListener('change', (e) => this.updateTextFormat('color', e.target.value));
    document.getElementById('alignLeftBtn')?.addEventListener('click', () => this.updateTextFormat('align', 'left'));
    document.getElementById('alignCenterBtn')?.addEventListener('click', () => this.updateTextFormat('align', 'center'));
    document.getElementById('alignRightBtn')?.addEventListener('click', () => this.updateTextFormat('align', 'right'));

    // Page operations
    document.getElementById('addPageBtn')?.addEventListener('click', () => this.addPage());
    document.getElementById('deletePageBtn')?.addEventListener('click', () => this.deletePage());

    // Component drag and drop
    this.setupDragAndDrop();

    // Color presets modal
    this.setupColorPresets();

    console.log('Event listeners setup complete');
  }

  setupButtonRipples() {
    // Add ripple effect to all buttons
    document.querySelectorAll('.btn, .tool-btn, .tool-btn-full').forEach(button => {
      button.addEventListener('click', (e) => {
        // Pulse animation on click
        anime({
          targets: button,
          scale: [1, 0.95, 1],
          duration: 200,
          easing: 'easeOutQuad'
        });
      });
    });
  }

  setupDragAndDrop() {
    const canvas = document.getElementById('pdfCanvas');
    if (!canvas) return;

    // Make component cards draggable
    document.querySelectorAll('.component-card').forEach(card => {
      card.addEventListener('dragstart', (e) => {
        e.dataTransfer.setData('componentType', card.getAttribute('data-type'));
        e.dataTransfer.effectAllowed = 'copy';
      });
    });

    // Setup drop zone on canvas
    canvas.addEventListener('dragover', (e) => {
      e.preventDefault();
      e.dataTransfer.dropEffect = 'copy';
    });

    canvas.addEventListener('drop', async (e) => {
      e.preventDefault();
      const componentType = e.dataTransfer.getData('componentType');
      if (componentType) {
        const rect = canvas.getBoundingClientRect();
        const x = e.clientX - rect.left;
        const y = e.clientY - rect.top;
        await this.addComponent(componentType, x, y);
      }
    });
  }

  setupColorPresets() {
    const colorPresetsBtn = document.getElementById('colorPresetsBtn');
    const modal = document.getElementById('colorPresetsModal');
    const closeBtn = document.getElementById('closeColorPresets');
    const textColorPicker = document.getElementById('textColorPicker');

    // Open modal
    colorPresetsBtn?.addEventListener('click', () => {
      modal.classList.remove('hidden');
      
      // Animate modal entrance
      anime({
        targets: modal,
        opacity: [0, 1],
        duration: 200,
        easing: 'easeOutQuad'
      });
      
      anime({
        targets: modal.querySelector('.modal'),
        scale: [0.9, 1],
        opacity: [0, 1],
        duration: 300,
        easing: 'easeOutBack'
      });
    });

    // Close modal
    const closeModal = () => {
      anime({
        targets: modal,
        opacity: 0,
        duration: 200,
        easing: 'easeInQuad',
        complete: () => {
          modal.classList.add('hidden');
        }
      });
    };

    closeBtn?.addEventListener('click', closeModal);
    modal?.addEventListener('click', (e) => {
      if (e.target === modal) {
        closeModal();
      }
    });

    // Color swatch selection
    document.querySelectorAll('.color-swatch').forEach(swatch => {
      swatch.addEventListener('click', () => {
        const color = swatch.getAttribute('data-color');
        if (textColorPicker) {
          textColorPicker.value = color;
          this.updateTextFormat('color', color);
        }
        
        // Animate selection
        anime({
          targets: swatch,
          scale: [1.15, 1.3, 1],
          duration: 300,
          easing: 'easeOutElastic(1, .5)'
        });
        
        closeModal();
        this.showToast('success', 'Color Applied', `Selected color: ${color}`);
      });
    });
  }

  switchTab(sidebar, tabName) {
    // Update tab buttons
    sidebar.querySelectorAll('.tab-btn').forEach(btn => {
      btn.classList.toggle('active', btn.getAttribute('data-tab') === tabName);
    });

    // Update tab contents
    sidebar.querySelectorAll('.tab-content').forEach(content => {
      content.classList.toggle('active', content.id === `${tabName}-tab`);
    });
  }

  setTool(toolName) {
    this.activeTool = toolName;
    
    // Update UI with animations
    document.querySelectorAll('.tool-btn[data-tool]').forEach(btn => {
      const isActive = btn.getAttribute('data-tool') === toolName;
      btn.classList.toggle('active', isActive);
      
      // Bounce animation for selected tool
      if (isActive) {
        anime({
          targets: btn,
          scale: [1, 1.15, 1],
          duration: 400,
          easing: 'easeOutElastic(1, .6)'
        });
      }
    });

    // Show/hide formatting toolbar for text tool
    const formattingToolbar = document.getElementById('formattingToolbar');
    if (formattingToolbar) {
      if (toolName === 'text') {
        formattingToolbar.classList.remove('hidden');
        anime({
          targets: formattingToolbar,
          translateY: [-20, 0],
          opacity: [0, 1],
          duration: 300,
          easing: 'easeOutCubic'
        });
      } else {
        anime({
          targets: formattingToolbar,
          translateY: [0, -20],
          opacity: [1, 0],
          duration: 200,
          easing: 'easeInCubic',
          complete: () => {
            formattingToolbar.classList.add('hidden');
          }
        });
      }
    }

    this.showStatus(`Tool: ${toolName}`);
  }

  // ========== FILE OPERATIONS ==========

  async createNewPDF() {
    this.showLoading('Creating new PDF...');
    try {
      const result = await window.electronAPI.createNewPDF({ pageSize: 'Letter' });
      if (result.success) {
        this.currentPDF = result.data;
        await this.renderPDF();
        this.showToast('success', 'New PDF Created', 'A new blank PDF has been created');
      } else {
        throw new Error(result.error);
      }
    } catch (error) {
      console.error('Error creating PDF:', error);
      this.showToast('error', 'Error', 'Failed to create new PDF');
    } finally {
      this.hideLoading();
    }
  }

  async openPDF() {
    this.showLoading('Opening PDF...');
    try {
      const result = await window.electronAPI.openFile({
        filters: [{ name: 'PDF Files', extensions: ['pdf'] }]
      });

      if (result.success && result.data) {
        this.currentPDF = result.data;
        await this.renderPDF();
        this.showToast('success', 'PDF Opened', 'PDF loaded successfully');
      }
    } catch (error) {
      console.error('Error opening PDF:', error);
      this.showToast('error', 'Error', 'Failed to open PDF');
    } finally {
      this.hideLoading();
    }
  }

  async savePDF() {
    if (!this.currentPDF) {
      this.showToast('error', 'Error', 'No PDF loaded');
      return;
    }

    this.showLoading('Saving PDF...');
    try {
      const result = await window.electronAPI.saveDocument();
      if (result.success) {
        this.showToast('success', 'Saved', 'PDF saved successfully');
      } else {
        throw new Error(result.error);
      }
    } catch (error) {
      console.error('Error saving PDF:', error);
      this.showToast('error', 'Error', 'Failed to save PDF');
    } finally {
      this.hideLoading();
    }
  }

  async exportPDF() {
    if (!this.currentPDF) {
      this.showToast('error', 'Error', 'No PDF loaded');
      return;
    }

    this.showLoading('Exporting PDF...');
    try {
      const result = await window.electronAPI.exportDocument({
        format: 'pdf',
        quality: 'high'
      });
      if (result.success) {
        this.showToast('success', 'Exported', 'PDF exported successfully');
      }
    } catch (error) {
      console.error('Error exporting PDF:', error);
      this.showToast('error', 'Error', 'Failed to export PDF');
    } finally {
      this.hideLoading();
    }
  }

  async printPDF() {
    if (!this.currentPDF) {
      this.showToast('error', 'Error', 'No PDF loaded');
      return;
    }

    try {
      window.print();
    } catch (error) {
      console.error('Error printing:', error);
      this.showToast('error', 'Error', 'Failed to print PDF');
    }
  }

  // ========== PDF OPERATIONS ==========

  async mergePDFs() {
    this.showLoading('Merging PDFs...');
    try {
      const files = await window.electronAPI.openFile({
        filters: [{ name: 'PDF Files', extensions: ['pdf'] }],
        properties: ['openFile', 'multiSelections']
      });

      if (files.success && files.data && files.data.length > 1) {
        const result = await window.electronAPI.mergePDFs(files.data);
        if (result.success) {
          this.currentPDF = result.data;
          await this.renderPDF();
          this.showToast('success', 'Merged', `${files.data.length} PDFs merged successfully`);
        }
      }
    } catch (error) {
      console.error('Error merging PDFs:', error);
      this.showToast('error', 'Error', 'Failed to merge PDFs');
    } finally {
      this.hideLoading();
    }
  }

  async splitPDF() {
    if (!this.currentPDF) {
      this.showToast('error', 'Error', 'No PDF loaded');
      return;
    }

    // Show modal to get page ranges
    const ranges = await this.showSplitDialog();
    if (!ranges) return;

    this.showLoading('Splitting PDF...');
    try {
      const result = await window.electronAPI.splitPDF(ranges);
      if (result.success) {
        this.showToast('success', 'Split Complete', 'PDF split into separate files');
      }
    } catch (error) {
      console.error('Error splitting PDF:', error);
      this.showToast('error', 'Error', 'Failed to split PDF');
    } finally {
      this.hideLoading();
    }
  }

  async compressPDF() {
    if (!this.currentPDF) {
      this.showToast('error', 'Error', 'No PDF loaded');
      return;
    }

    this.showLoading('Compressing PDF...');
    try {
      const result = await window.electronAPI.compressPDF({ quality: 0.8 });
      if (result.success) {
        this.currentPDF = result.data;
        await this.renderPDF();
        this.showToast('success', 'Compressed', 'PDF compressed successfully');
      }
    } catch (error) {
      console.error('Error compressing PDF:', error);
      this.showToast('error', 'Error', 'Failed to compress PDF');
    } finally {
      this.hideLoading();
    }
  }

  async encryptPDF() {
    if (!this.currentPDF) {
      this.showToast('error', 'Error', 'No PDF loaded');
      return;
    }

    const password = await this.showPasswordDialog();
    if (!password) return;

    this.showLoading('Encrypting PDF...');
    try {
      const result = await window.electronAPI.encryptPDF({
        userPassword: password,
        ownerPassword: password
      });
      if (result.success) {
        this.showToast('success', 'Encrypted', 'PDF encrypted successfully');
      }
    } catch (error) {
      console.error('Error encrypting PDF:', error);
      this.showToast('error', 'Error', 'Failed to encrypt PDF');
    } finally {
      this.hideLoading();
    }
  }

  async imagesToPDF() {
    this.showLoading('Converting images to PDF...');
    try {
      const files = await window.electronAPI.openFile({
        filters: [
          { name: 'Images', extensions: ['jpg', 'jpeg', 'png'] }
        ],
        properties: ['openFile', 'multiSelections']
      });

      if (files.success && files.data) {
        const result = await window.electronAPI.imagesToPDF(files.data);
        if (result.success) {
          this.currentPDF = result.data;
          await this.renderPDF();
          this.showToast('success', 'Converted', `${files.data.length} images converted to PDF`);
        }
      }
    } catch (error) {
      console.error('Error converting images:', error);
      this.showToast('error', 'Error', 'Failed to convert images to PDF');
    } finally {
      this.hideLoading();
    }
  }

  // ========== CONTENT OPERATIONS ==========

  async addComponent(type, x, y) {
    if (!this.currentPDF) {
      this.showToast('error', 'Error', 'No PDF loaded');
      return;
    }

    this.showLoading(`Adding ${type}...`);
    try {
      let result;
      
      switch (type) {
        case 'heading1':
          result = await window.electronAPI.addTextToPDF({
            text: 'Heading 1',
            x, y,
            size: 36,
            font: 'Helvetica-Bold',
            page: this.currentPage
          });
          break;

        case 'heading2':
          result = await window.electronAPI.addTextToPDF({
            text: 'Heading 2',
            x, y,
            size: 24,
            font: 'Helvetica-Bold',
            page: this.currentPage
          });
          break;

        case 'paragraph':
          result = await window.electronAPI.addTextToPDF({
            text: 'Paragraph text',
            x, y,
            size: 12,
            font: 'Helvetica',
            page: this.currentPage
          });
          break;

        case 'image':
          const imageFile = await window.electronAPI.openFile({
            filters: [{ name: 'Images', extensions: ['jpg', 'jpeg', 'png'] }]
          });
          if (imageFile.success && imageFile.data) {
            result = await window.electronAPI.addImageToPDF({
              imagePath: imageFile.data[0],
              x, y,
              width: 200,
              height: 150,
              page: this.currentPage
            });
          }
          break;

        case 'qrcode':
          const qrData = await this.showQRDialog();
          if (qrData) {
            result = await window.electronAPI.addQRCodeToPDF({
              data: qrData,
              x, y,
              size: 100,
              page: this.currentPage
            });
          }
          break;

        case 'barcode':
          const barcodeData = await this.showBarcodeDialog();
          if (barcodeData) {
            result = await window.electronAPI.addBarcodeToPDF({
              data: barcodeData.data,
              format: barcodeData.format || 'CODE128',
              x, y,
              width: 200,
              height: 60,
              page: this.currentPage
            });
          }
          break;

        default:
          console.log(`Component type ${type} not yet implemented`);
          return;
      }

      if (result && result.success) {
        await this.renderPDF();
        this.showToast('success', 'Added', `${type} added successfully`);
        
        // Celebrate with animation
        this.celebrateAction();
      }
    } catch (error) {
      console.error(`Error adding ${type}:`, error);
      this.showToast('error', 'Error', `Failed to add ${type}`);
    } finally {
      this.hideLoading();
    }
  }

  celebrateAction() {
    // Create confetti-like animation effect
    const canvas = document.getElementById('pdfCanvas');
    if (!canvas) return;

    for (let i = 0; i < 10; i++) {
      const particle = document.createElement('div');
      particle.style.position = 'fixed';
      particle.style.width = '10px';
      particle.style.height = '10px';
      particle.style.borderRadius = '50%';
      particle.style.backgroundColor = `hsl(${Math.random() * 360}, 70%, 60%)`;
      particle.style.pointerEvents = 'none';
      particle.style.left = '50%';
      particle.style.top = '50%';
      particle.style.zIndex = '10000';
      document.body.appendChild(particle);

      anime({
        targets: particle,
        translateX: () => anime.random(-200, 200),
        translateY: () => anime.random(-200, 200),
        scale: [1, 0],
        opacity: [1, 0],
        duration: 1000,
        easing: 'easeOutExpo',
        complete: () => {
          particle.remove();
        }
      });
    }
  }

  // ========== PAGE OPERATIONS ==========

  async addPage() {
    if (!this.currentPDF) {
      this.showToast('error', 'Error', 'No PDF loaded');
      return;
    }

    this.showLoading('Adding page...');
    try {
      const result = await window.electronAPI.addBlankPage({
        width: 612,
        height: 792,
        position: this.currentPage
      });

      if (result.success) {
        await this.renderPDF();
        this.showToast('success', 'Page Added', 'New blank page added');
      }
    } catch (error) {
      console.error('Error adding page:', error);
      this.showToast('error', 'Error', 'Failed to add page');
    } finally {
      this.hideLoading();
    }
  }

  async deletePage() {
    if (!this.currentPDF || this.totalPages === 0) {
      this.showToast('error', 'Error', 'No pages to delete');
      return;
    }

    const confirm = await this.showConfirmDialog('Delete this page?');
    if (!confirm) return;

    this.showLoading('Deleting page...');
    try {
      const result = await window.electronAPI.deletePDFPages([this.currentPage]);
      if (result.success) {
        if (this.currentPage > 1) {
          this.currentPage--;
        }
        await this.renderPDF();
        this.showToast('success', 'Page Deleted', 'Page removed successfully');
      }
    } catch (error) {
      console.error('Error deleting page:', error);
      this.showToast('error', 'Error', 'Failed to delete page');
    } finally {
      this.hideLoading();
    }
  }

  // ========== EDIT OPERATIONS ==========

  undo() {
    if (this.history.past.length === 0) {
      this.showToast('info', 'Nothing to Undo', 'No more actions to undo');
      return;
    }

    const previousState = this.history.past.pop();
    this.history.future.push(this.getCurrentState());
    this.restoreState(previousState);
    this.showStatus('Undo');
  }

  redo() {
    if (this.history.future.length === 0) {
      this.showToast('info', 'Nothing to Redo', 'No more actions to redo');
      return;
    }

    const nextState = this.history.future.pop();
    this.history.past.push(this.getCurrentState());
    this.restoreState(nextState);
    this.showStatus('Redo');
  }

  cut() {
    if (this.selectedElement) {
      this.copy();
      this.deleteSelected();
    }
  }

  copy() {
    if (this.selectedElement) {
      localStorage.setItem('clipboard', JSON.stringify(this.selectedElement));
      this.showStatus('Copied');
    }
  }

  paste() {
    const clipboardData = localStorage.getItem('clipboard');
    if (clipboardData) {
      const element = JSON.parse(clipboardData);
      element.x += 20;
      element.y += 20;
      this.elements.push(element);
      this.renderElements();
      this.showStatus('Pasted');
    }
  }

  deleteSelected() {
    if (this.selectedElement) {
      const index = this.elements.indexOf(this.selectedElement);
      if (index > -1) {
        this.elements.splice(index, 1);
        this.selectedElement = null;
        this.renderElements();
        this.showStatus('Deleted');
      }
    }
  }

  selectAll() {
    console.log('Select all elements');
    this.showStatus('All selected');
  }

  duplicate() {
    if (this.selectedElement) {
      const copy = { ...this.selectedElement, x: this.selectedElement.x + 20, y: this.selectedElement.y + 20 };
      this.elements.push(copy);
      this.renderElements();
      this.showStatus('Duplicated');
    }
  }

  // ========== VIEW OPERATIONS ==========

  zoomIn() {
    this.zoom = Math.min(this.zoom + 0.1, 3.0);
    this.updateZoom();
  }

  zoomOut() {
    this.zoom = Math.max(this.zoom - 0.1, 0.1);
    this.updateZoom();
  }

  zoomReset() {
    this.zoom = 1.0;
    this.updateZoom();
  }

  fitWidth() {
    const canvas = document.getElementById('pdfCanvas');
    if (canvas && canvas.firstChild) {
      const page = canvas.firstChild;
      const containerWidth = canvas.offsetWidth - 80;
      const pageWidth = page.offsetWidth;
      this.zoom = containerWidth / pageWidth;
      this.updateZoom();
    }
  }

  fitPage() {
    const canvas = document.getElementById('pdfCanvas');
    if (canvas && canvas.firstChild) {
      const page = canvas.firstChild;
      const containerHeight = canvas.offsetHeight - 80;
      const pageHeight = page.offsetHeight;
      this.zoom = containerHeight / pageHeight;
      this.updateZoom();
    }
  }

  updateZoom() {
    document.getElementById('zoomLevel').textContent = `${Math.round(this.zoom * 100)}%`;
    
    const pages = document.querySelectorAll('.pdf-page');
    pages.forEach((page, index) => {
      // Smooth zoom animation
      anime({
        targets: page,
        scale: this.zoom,
        duration: 300,
        delay: index * 30,
        easing: 'easeOutCubic'
      });
    });
    
    this.showStatus(`Zoom: ${Math.round(this.zoom * 100)}%`);
  }

  toggleFullscreen() {
    if (!document.fullscreenElement) {
      document.documentElement.requestFullscreen();
    } else {
      document.exitFullscreen();
    }
  }

  // ========== NAVIGATION ==========

  nextPage() {
    if (this.currentPage < this.totalPages) {
      this.currentPage++;
      this.updatePageDisplay();
    }
  }

  prevPage() {
    if (this.currentPage > 1) {
      this.currentPage--;
      this.updatePageDisplay();
    }
  }

  firstPage() {
    this.currentPage = 1;
    this.updatePageDisplay();
  }

  lastPage() {
    this.currentPage = this.totalPages;
    this.updatePageDisplay();
  }

  updatePageDisplay() {
    document.getElementById('currentPageNum').textContent = this.currentPage;
    document.getElementById('totalPagesNum').textContent = this.totalPages;
    
    // Scroll to current page with smooth animation
    const pages = document.querySelectorAll('.pdf-page');
    if (pages[this.currentPage - 1]) {
      const targetPage = pages[this.currentPage - 1];
      
      // Highlight animation
      anime({
        targets: targetPage,
        boxShadow: [
          '0 10px 15px -3px rgba(59, 130, 246, 0)',
          '0 10px 15px -3px rgba(59, 130, 246, 0.5)',
          '0 10px 15px -3px rgba(0, 0, 0, 0.3)'
        ],
        duration: 600,
        easing: 'easeOutCubic'
      });
      
      targetPage.scrollIntoView({ behavior: 'smooth', block: 'center' });
    }
  }

  // ========== TEXT FORMATTING ==========

  updateTextFormat(property, value) {
    if (!this.selectedElement || this.selectedElement.type !== 'text') {
      this.showToast('info', 'No Text Selected', 'Select a text element first');
      return;
    }

    switch (property) {
      case 'font':
        this.selectedElement.font = value;
        break;
      case 'size':
        this.selectedElement.size = parseInt(value);
        break;
      case 'color':
        this.selectedElement.color = value;
        break;
      case 'bold':
        this.selectedElement.bold = !this.selectedElement.bold;
        break;
      case 'italic':
        this.selectedElement.italic = !this.selectedElement.italic;
        break;
      case 'underline':
        this.selectedElement.underline = !this.selectedElement.underline;
        break;
      case 'align':
        this.selectedElement.align = value;
        break;
    }

    this.renderElements();
    this.showStatus(`Format updated: ${property}`);
  }

  // ========== RENDERING ==========

  async renderPDF() {
    const canvas = document.getElementById('pdfCanvas');
    if (!canvas) return;

    // Clear canvas
    canvas.innerHTML = '';

    if (!this.currentPDF) {
      this.showEmptyState();
      return;
    }

    // Get PDF info
    try {
      const info = await window.electronAPI.getPDFInfo();
      if (info.success) {
        this.totalPages = info.data.pages;
        this.updatePageDisplay();
        this.updateDocumentInfo(info.data);
      }

      // Render pages (simplified - actual rendering would use PDF.js)
      for (let i = 1; i <= this.totalPages; i++) {
        const pageDiv = document.createElement('div');
        pageDiv.className = 'pdf-page';
        pageDiv.style.width = '612px';
        pageDiv.style.height = '792px';
        pageDiv.dataset.page = i;
        
        // Add page canvas
        const pageCanvas = document.createElement('canvas');
        pageCanvas.width = 612;
        pageCanvas.height = 792;
        pageDiv.appendChild(pageCanvas);
        
        canvas.appendChild(pageDiv);
        
        // Animate page appearance
        anime({
          targets: pageDiv,
          opacity: [0, 1],
          translateY: [20, 0],
          duration: 400,
          delay: i * 50,
          easing: 'easeOutCubic'
        });
      }

      // Update pages list
      this.renderPagesList();

    } catch (error) {
      console.error('Error rendering PDF:', error);
      this.showToast('error', 'Error', 'Failed to render PDF');
    }
  }

  renderPagesList() {
    const pagesList = document.getElementById('pagesList');
    if (!pagesList) return;

    pagesList.innerHTML = '';

    for (let i = 1; i <= this.totalPages; i++) {
      const thumbnail = document.createElement('div');
      thumbnail.className = 'page-thumbnail';
      if (i === this.currentPage) {
        thumbnail.classList.add('active');
      }
      
      const pageNum = document.createElement('div');
      pageNum.className = 'page-number';
      pageNum.textContent = i;
      thumbnail.appendChild(pageNum);
      
      thumbnail.addEventListener('click', () => {
        this.currentPage = i;
        this.updatePageDisplay();
        this.renderPagesList();
      });
      
      pagesList.appendChild(thumbnail);
    }

    // Make pages sortable
    if (window.Sortable && pagesList) {
      Sortable.create(pagesList, {
        animation: 150,
        ghostClass: 'sortable-ghost',
        onEnd: (evt) => {
          this.reorderPages(evt.oldIndex, evt.newIndex);
        }
      });
    }
  }

  async reorderPages(oldIndex, newIndex) {
    const newOrder = [];
    for (let i = 1; i <= this.totalPages; i++) {
      newOrder.push(i);
    }
    
    const [moved] = newOrder.splice(oldIndex, 1);
    newOrder.splice(newIndex, 0, moved);

    this.showLoading('Reordering pages...');
    try {
      const result = await window.electronAPI.reorderPDFPages(newOrder);
      if (result.success) {
        await this.renderPDF();
        this.showToast('success', 'Reordered', 'Pages reordered successfully');
      }
    } catch (error) {
      console.error('Error reordering pages:', error);
      this.showToast('error', 'Error', 'Failed to reorder pages');
    } finally {
      this.hideLoading();
    }
  }

  renderElements() {
    // Render all PDF elements (text, images, shapes, etc.)
    this.elements.forEach(element => {
      // Element rendering logic
      console.log('Rendering element:', element);
    });
  }

  showEmptyState() {
    const canvas = document.getElementById('pdfCanvas');
    if (!canvas) return;

    canvas.innerHTML = `
      <div class="empty-state">
        <svg width="64" height="64" viewBox="0 0 24 24" fill="none" stroke="currentColor" stroke-width="1.5">
          <path d="M14 2H6a2 2 0 0 0-2 2v16a2 2 0 0 0 2 2h12a2 2 0 0 0 2-2V8z"></path>
          <polyline points="14 2 14 8 20 8"></polyline>
        </svg>
        <h3>No PDF Loaded</h3>
        <p>Open a PDF or create a new one to get started</p>
        <div class="empty-actions">
          <button class="btn btn-primary" onclick="editor.createNewPDF()">New PDF</button>
          <button class="btn btn-ghost" onclick="editor.openPDF()">Open PDF</button>
        </div>
      </div>
    `;
  }

  // ========== DOCUMENT INFO ==========

  updateDocumentInfo(info) {
    document.getElementById('docTitle').value = info.title || '';
    document.getElementById('docAuthor').value = info.author || '';
    document.getElementById('docSubject').value = info.subject || '';
    document.getElementById('docKeywords').value = info.keywords || '';
    
    document.getElementById('statPages').textContent = info.pages || 0;
    document.getElementById('statElements').textContent = this.elements.length;
    document.getElementById('statFileSize').textContent = this.formatFileSize(info.size || 0);
  }

  formatFileSize(bytes) {
    if (bytes < 1024) return bytes + ' B';
    if (bytes < 1024 * 1024) return (bytes / 1024).toFixed(1) + ' KB';
    return (bytes / (1024 * 1024)).toFixed(1) + ' MB';
  }

  // ========== SHORTCUTS UI ==========

  loadShortcutsList() {
    const shortcutsList = document.getElementById('shortcutsList');
    if (!shortcutsList || !this.keybindingManager) return;

    const bindings = this.keybindingManager.getAllBindings();
    shortcutsList.innerHTML = '';

    // Group by category
    const categories = {
      file: 'File',
      edit: 'Edit',
      view: 'View',
      nav: 'Navigation',
      tool: 'Tools',
      format: 'Format',
      insert: 'Insert',
      page: 'Page'
    };

    Object.entries(categories).forEach(([prefix, title]) => {
      const categoryBindings = Object.entries(bindings).filter(([action]) => action.startsWith(prefix));
      
      if (categoryBindings.length > 0) {
        const categoryDiv = document.createElement('div');
        categoryDiv.className = 'shortcut-category';
        categoryDiv.innerHTML = `<h4 class="section-title">${title}</h4>`;
        
        categoryBindings.forEach(([action, binding]) => {
          const item = document.createElement('div');
          item.className = 'shortcut-item';
          item.innerHTML = `
            <span class="shortcut-action">${binding.description}</span>
            <div class="shortcut-keys">
              ${this.formatShortcutKeys(binding.keys)}
            </div>
          `;
          categoryDiv.appendChild(item);
        });
        
        shortcutsList.appendChild(categoryDiv);
      }
    });
  }

  formatShortcutKeys(keys) {
    return keys.split('+').map(key => 
      `<span class="key">${key.toUpperCase()}</span>`
    ).join('+');
  }

  // ========== DIALOGS ==========

  async showSplitDialog() {
    // Simplified - would show a modal
    const ranges = prompt('Enter page ranges (e.g., 1-3, 4-6):');
    return ranges ? ranges.split(',').map(r => r.trim()) : null;
  }

  async showPasswordDialog() {
    const password = prompt('Enter password for encryption:');
    return password;
  }

  async showQRDialog() {
    const data = prompt('Enter data for QR code:');
    return data;
  }

  async showBarcodeDialog() {
    const data = prompt('Enter barcode data:');
    return data ? { data, format: 'CODE128' } : null;
  }

  async showConfirmDialog(message) {
    return confirm(message);
  }

  // ========== STATE MANAGEMENT ==========

  getCurrentState() {
    return {
      elements: JSON.parse(JSON.stringify(this.elements)),
      currentPage: this.currentPage,
      zoom: this.zoom
    };
  }

  restoreState(state) {
    this.elements = state.elements;
    this.currentPage = state.currentPage;
    this.zoom = state.zoom;
    this.renderElements();
    this.updateZoom();
  }

  // ========== UI FEEDBACK ==========

  showLoading(message) {
    this.isLoading = true;
    const overlay = document.createElement('div');
    overlay.className = 'loading-overlay';
    overlay.id = 'loadingOverlay';
    overlay.innerHTML = `
      <div class="loading-spinner"></div>
      <div style="color: var(--text-secondary); margin-top: 16px;">${message}</div>
    `;
    document.body.appendChild(overlay);

    // Fade in animation
    anime({
      targets: overlay,
      opacity: [0, 1],
      duration: 200,
      easing: 'easeOutQuad'
    });

    // Spinner rotation
    anime({
      targets: overlay.querySelector('.loading-spinner'),
      rotate: '360deg',
      duration: 1000,
      easing: 'linear',
      loop: true
    });
  }

  hideLoading() {
    this.isLoading = false;
    const overlay = document.getElementById('loadingOverlay');
    if (overlay) {
      anime({
        targets: overlay,
        opacity: [1, 0],
        duration: 200,
        easing: 'easeInQuad',
        complete: () => {
          overlay.remove();
        }
      });
    }
  }

  showToast(type, title, message) {
    const container = document.getElementById('toastContainer');
    if (!container) return;

    const toast = document.createElement('div');
    toast.className = `toast toast-${type}`;
    
    const icons = {
      success: '<svg width="20" height="20" viewBox="0 0 24 24" fill="none" stroke="currentColor" stroke-width="2"><polyline points="20 6 9 17 4 12"></polyline></svg>',
      error: '<svg width="20" height="20" viewBox="0 0 24 24" fill="none" stroke="currentColor" stroke-width="2"><circle cx="12" cy="12" r="10"></circle><line x1="15" y1="9" x2="9" y2="15"></line><line x1="9" y1="9" x2="15" y2="15"></line></svg>',
      info: '<svg width="20" height="20" viewBox="0 0 24 24" fill="none" stroke="currentColor" stroke-width="2"><circle cx="12" cy="12" r="10"></circle><line x1="12" y1="16" x2="12" y2="12"></line><line x1="12" y1="8" x2="12.01" y2="8"></line></svg>'
    };

    toast.innerHTML = `
      <div class="toast-icon">${icons[type]}</div>
      <div class="toast-content">
        <div class="toast-title">${title}</div>
        <div class="toast-message">${message}</div>
      </div>
      <button class="toast-close">×</button>
    `;

    toast.querySelector('.toast-close').addEventListener('click', () => {
      this.removeToast(toast);
    });

    container.appendChild(toast);

    // Entrance animation
    anime({
      targets: toast,
      translateX: [300, 0],
      opacity: [0, 1],
      duration: 400,
      easing: 'easeOutBack'
    });

    // Auto-remove after 5 seconds
    setTimeout(() => {
      this.removeToast(toast);
    }, 5000);
  }

  removeToast(toast) {
    anime({
      targets: toast,
      translateX: [0, 300],
      opacity: [1, 0],
      duration: 300,
      easing: 'easeInBack',
      complete: () => {
        toast.remove();
      }
    });
  }

  showStatus(message) {
    const statusMessage = document.getElementById('statusMessage');
    if (statusMessage) {
      statusMessage.textContent = message;
      
      // Reset after 3 seconds
      setTimeout(() => {
        statusMessage.textContent = 'Ready';
      }, 3000);
    }
  }
}

// Initialize editor when DOM is ready
let editor;
document.addEventListener('DOMContentLoaded', () => {
  editor = new PDFEditor();
  window.editor = editor; // Make globally accessible for debugging
});
