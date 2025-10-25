/**
 * Keybinding Manager
 * Configurable keyboard shortcuts system
 */

const Mousetrap = require('mousetrap');

class KeybindingManager {
  constructor() {
    this.bindings = this.getDefaultBindings();
    this.customBindings = this.loadCustomBindings();
    this.callbacks = {};
    this.mousetrap = new Mousetrap();
  }

  /**
   * Default keybindings
   */
  getDefaultBindings() {
    return {
      // File operations
      'file.new': { keys: 'ctrl+n', description: 'New document' },
      'file.open': { keys: 'ctrl+o', description: 'Open document' },
      'file.save': { keys: 'ctrl+s', description: 'Save document' },
      'file.saveAs': { keys: 'ctrl+shift+s', description: 'Save as' },
      'file.export': { keys: 'ctrl+e', description: 'Export document' },
      'file.print': { keys: 'ctrl+p', description: 'Print document' },
      'file.close': { keys: 'ctrl+w', description: 'Close document' },
      
      // Edit operations
      'edit.undo': { keys: 'ctrl+z', description: 'Undo' },
      'edit.redo': { keys: 'ctrl+y', description: 'Redo' },
      'edit.cut': { keys: 'ctrl+x', description: 'Cut' },
      'edit.copy': { keys: 'ctrl+c', description: 'Copy' },
      'edit.paste': { keys: 'ctrl+v', description: 'Paste' },
      'edit.selectAll': { keys: 'ctrl+a', description: 'Select all' },
      'edit.delete': { keys: 'delete', description: 'Delete selection' },
      'edit.duplicate': { keys: 'ctrl+d', description: 'Duplicate' },
      
      // View operations
      'view.zoomIn': { keys: 'ctrl+plus', description: 'Zoom in' },
      'view.zoomOut': { keys: 'ctrl+-', description: 'Zoom out' },
      'view.zoomReset': { keys: 'ctrl+0', description: 'Reset zoom' },
      'view.fitWidth': { keys: 'ctrl+1', description: 'Fit to width' },
      'view.fitPage': { keys: 'ctrl+2', description: 'Fit to page' },
      'view.fullscreen': { keys: 'f11', description: 'Toggle fullscreen' },
      'view.toggleSidebar': { keys: 'ctrl+b', description: 'Toggle sidebar' },
      'view.toggleProperties': { keys: 'ctrl+i', description: 'Toggle properties' },
      
      // Navigation
      'nav.nextPage': { keys: 'pagedown', description: 'Next page' },
      'nav.prevPage': { keys: 'pageup', description: 'Previous page' },
      'nav.firstPage': { keys: 'home', description: 'First page' },
      'nav.lastPage': { keys: 'end', description: 'Last page' },
      'nav.goToPage': { keys: 'ctrl+g', description: 'Go to page' },
      
      // Tools
      'tool.text': { keys: 't', description: 'Text tool' },
      'tool.image': { keys: 'i', description: 'Image tool' },
      'tool.shape': { keys: 's', description: 'Shape tool' },
      'tool.line': { keys: 'l', description: 'Line tool' },
      'tool.arrow': { keys: 'a', description: 'Arrow tool' },
      'tool.pen': { keys: 'p', description: 'Pen tool' },
      'tool.highlighter': { keys: 'h', description: 'Highlighter' },
      'tool.eraser': { keys: 'e', description: 'Eraser' },
      'tool.select': { keys: 'v', description: 'Selection tool' },
      'tool.pan': { keys: 'space', description: 'Pan tool (hold)' },
      
      // Text formatting
      'format.bold': { keys: 'ctrl+b', description: 'Bold text' },
      'format.italic': { keys: 'ctrl+i', description: 'Italic text' },
      'format.underline': { keys: 'ctrl+u', description: 'Underline text' },
      'format.alignLeft': { keys: 'ctrl+l', description: 'Align left' },
      'format.alignCenter': { keys: 'ctrl+shift+c', description: 'Align center' },
      'format.alignRight': { keys: 'ctrl+r', description: 'Align right' },
      'format.increaseFontSize': { keys: 'ctrl+]', description: 'Increase font size' },
      'format.decreaseFontSize': { keys: 'ctrl+[', description: 'Decrease font size' },
      
      // Components (insert)
      'insert.text': { keys: 'ctrl+shift+t', description: 'Insert text box' },
      'insert.image': { keys: 'ctrl+shift+i', description: 'Insert image' },
      'insert.shape': { keys: 'ctrl+shift+s', description: 'Insert shape' },
      'insert.table': { keys: 'ctrl+shift+b', description: 'Insert table' },
      'insert.link': { keys: 'ctrl+k', description: 'Insert link' },
      'insert.qrcode': { keys: 'ctrl+shift+q', description: 'Insert QR code' },
      'insert.barcode': { keys: 'ctrl+shift+a', description: 'Insert barcode' },
      
      // Page operations
      'page.add': { keys: 'ctrl+shift+n', description: 'Add new page' },
      'page.delete': { keys: 'ctrl+shift+delete', description: 'Delete page' },
      'page.duplicate': { keys: 'ctrl+shift+d', description: 'Duplicate page' },
      'page.rotateRight': { keys: 'ctrl+right', description: 'Rotate page right' },
      'page.rotateLeft': { keys: 'ctrl+left', description: 'Rotate page left' },
      
      // Search
      'search.find': { keys: 'ctrl+f', description: 'Find text' },
      'search.findNext': { keys: 'f3', description: 'Find next' },
      'search.findPrev': { keys: 'shift+f3', description: 'Find previous' },
      'search.replace': { keys: 'ctrl+h', description: 'Find and replace' },
      
      // Quick actions
      'quick.save': { keys: 'ctrl+s', description: 'Quick save' },
      'quick.export': { keys: 'ctrl+e', description: 'Quick export' },
      'quick.share': { keys: 'ctrl+shift+e', description: 'Share document' },
      
      // Help
      'help.shortcuts': { keys: 'ctrl+/', description: 'Show shortcuts' },
      'help.docs': { keys: 'f1', description: 'Open documentation' }
    };
  }

  /**
   * Load custom bindings from storage
   */
  loadCustomBindings() {
    try {
      const stored = localStorage.getItem('keybindings');
      return stored ? JSON.parse(stored) : {};
    } catch (e) {
      console.error('Failed to load custom keybindings:', e);
      return {};
    }
  }

  /**
   * Save custom bindings to storage
   */
  saveCustomBindings() {
    try {
      localStorage.setItem('keybindings', JSON.stringify(this.customBindings));
      return true;
    } catch (e) {
      console.error('Failed to save custom keybindings:', e);
      return false;
    }
  }

  /**
   * Get effective binding (custom overrides default)
   */
  getBinding(action) {
    return this.customBindings[action] || this.bindings[action];
  }

  /**
   * Get all bindings
   */
  getAllBindings() {
    const result = {};
    for (const action in this.bindings) {
      result[action] = this.getBinding(action);
    }
    return result;
  }

  /**
   * Set custom binding
   */
  setBinding(action, keys, description) {
    if (!this.bindings[action]) {
      console.warn(`Unknown action: ${action}`);
      return false;
    }

    // Unbind old keys
    const oldBinding = this.getBinding(action);
    if (oldBinding && this.callbacks[action]) {
      this.mousetrap.unbind(oldBinding.keys);
    }

    // Set new binding
    this.customBindings[action] = { keys, description: description || this.bindings[action].description };
    
    // Rebind if callback exists
    if (this.callbacks[action]) {
      this.mousetrap.bind(keys, this.callbacks[action]);
    }

    this.saveCustomBindings();
    return true;
  }

  /**
   * Reset binding to default
   */
  resetBinding(action) {
    if (!this.bindings[action]) return false;

    // Unbind custom
    const customBinding = this.customBindings[action];
    if (customBinding && this.callbacks[action]) {
      this.mousetrap.unbind(customBinding.keys);
    }

    // Remove custom binding
    delete this.customBindings[action];

    // Rebind default
    const defaultBinding = this.bindings[action];
    if (this.callbacks[action]) {
      this.mousetrap.bind(defaultBinding.keys, this.callbacks[action]);
    }

    this.saveCustomBindings();
    return true;
  }

  /**
   * Reset all bindings to default
   */
  resetAllBindings() {
    // Unbind all
    for (const action in this.customBindings) {
      const binding = this.customBindings[action];
      if (this.callbacks[action]) {
        this.mousetrap.unbind(binding.keys);
      }
    }

    // Clear custom bindings
    this.customBindings = {};

    // Rebind defaults
    for (const action in this.callbacks) {
      const binding = this.bindings[action];
      if (binding) {
        this.mousetrap.bind(binding.keys, this.callbacks[action]);
      }
    }

    this.saveCustomBindings();
    return true;
  }

  /**
   * Register callback for action
   */
  register(action, callback) {
    if (!this.bindings[action]) {
      console.warn(`Unknown action: ${action}`);
      return false;
    }

    this.callbacks[action] = callback;
    const binding = this.getBinding(action);
    
    this.mousetrap.bind(binding.keys, (e) => {
      e.preventDefault();
      callback(e);
      return false;
    });

    return true;
  }

  /**
   * Unregister callback
   */
  unregister(action) {
    if (!this.callbacks[action]) return false;

    const binding = this.getBinding(action);
    this.mousetrap.unbind(binding.keys);
    delete this.callbacks[action];

    return true;
  }

  /**
   * Register multiple callbacks
   */
  registerMultiple(actions) {
    for (const action in actions) {
      this.register(action, actions[action]);
    }
  }

  /**
   * Check if keys are already bound
   */
  isKeysBound(keys) {
    const allBindings = this.getAllBindings();
    for (const action in allBindings) {
      if (allBindings[action].keys === keys) {
        return action;
      }
    }
    return null;
  }

  /**
   * Get formatted key display (for UI)
   */
  formatKeys(keys) {
    return keys
      .replace(/ctrl/gi, '⌘')
      .replace(/shift/gi, '⇧')
      .replace(/alt/gi, '⌥')
      .replace(/\+/g, ' + ')
      .toUpperCase();
  }

  /**
   * Export keybindings as JSON
   */
  exportBindings() {
    return JSON.stringify(this.customBindings, null, 2);
  }

  /**
   * Import keybindings from JSON
   */
  importBindings(jsonString) {
    try {
      const imported = JSON.parse(jsonString);
      
      // Validate
      for (const action in imported) {
        if (!this.bindings[action]) {
          console.warn(`Unknown action in import: ${action}`);
          continue;
        }
        
        const { keys, description } = imported[action];
        this.setBinding(action, keys, description);
      }

      return true;
    } catch (e) {
      console.error('Failed to import keybindings:', e);
      return false;
    }
  }

  /**
   * Disable all keybindings
   */
  disable() {
    this.mousetrap.reset();
  }

  /**
   * Enable all keybindings
   */
  enable() {
    for (const action in this.callbacks) {
      const binding = this.getBinding(action);
      this.mousetrap.bind(binding.keys, this.callbacks[action]);
    }
  }
}

// Singleton instance
let keybindingManager = null;

function getKeybindingManager() {
  if (!keybindingManager) {
    keybindingManager = new KeybindingManager();
  }
  return keybindingManager;
}

module.exports = { KeybindingManager, getKeybindingManager };
