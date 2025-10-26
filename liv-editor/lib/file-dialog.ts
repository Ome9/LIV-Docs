// File dialog utilities for Electron integration
declare global {
  interface Window {
    electronAPI?: {
      showOpenDialog: (options: any) => Promise<string | null>
      showOpenDialogMultiple: (options: any) => Promise<string[] | null>
      showSaveDialog: (options: any) => Promise<string | null>
    }
  }
}

export class FileDialogClient {
  private isElectron(): boolean {
    return typeof window !== 'undefined' && window.electronAPI !== undefined
  }

  async openFile(options: {
    title?: string
    filters?: Array<{ name: string; extensions: string[] }>
    defaultPath?: string
  } = {}): Promise<string | null> {
    if (!this.isElectron()) {
      throw new Error('File dialogs only available in desktop app')
    }

    const defaultOptions = {
      title: 'Open File',
      filters: [
        { name: 'PDF Files', extensions: ['pdf'] },
        { name: 'All Files', extensions: ['*'] }
      ]
    }

    return window.electronAPI!.showOpenDialog({
      ...defaultOptions,
      ...options,
      properties: ['openFile']
    })
  }

  async openFiles(options: {
    title?: string
    filters?: Array<{ name: string; extensions: string[] }>
    defaultPath?: string
  } = {}): Promise<string[] | null> {
    if (!this.isElectron()) {
      throw new Error('File dialogs only available in desktop app')
    }

    const defaultOptions = {
      title: 'Select Files',
      filters: [
        { name: 'PDF Files', extensions: ['pdf'] },
        { name: 'All Files', extensions: ['*'] }
      ]
    }

    return window.electronAPI!.showOpenDialogMultiple({
      ...defaultOptions,
      ...options,
      properties: ['openFile']
    })
  }

  async saveFile(options: {
    title?: string
    defaultPath?: string
    filters?: Array<{ name: string; extensions: string[] }>
  } = {}): Promise<string | null> {
    if (!this.isElectron()) {
      throw new Error('File dialogs only available in desktop app')
    }

    const defaultOptions = {
      title: 'Save File',
      defaultPath: 'output.pdf',
      filters: [
        { name: 'PDF Files', extensions: ['pdf'] },
        { name: 'All Files', extensions: ['*'] }
      ]
    }

    return window.electronAPI!.showSaveDialog({
      ...defaultOptions,
      ...options
    })
  }

  async selectDirectory(options: {
    title?: string
    defaultPath?: string
  } = {}): Promise<string | null> {
    if (!this.isElectron()) {
      throw new Error('File dialogs only available in desktop app')
    }

    const defaultOptions = {
      title: 'Select Directory'
    }

    const result = await window.electronAPI!.showOpenDialog({
      ...defaultOptions,
      ...options,
      properties: ['openDirectory']
    })

    // Convert single result to string for directory selection
    return Array.isArray(result) ? result[0] || null : result
  }

  // Helper method to check if running in Electron
  get isDesktopApp(): boolean {
    return this.isElectron()
  }
}

// Create singleton instance
export const fileDialog = new FileDialogClient()

// React hook
export function useFileDialog() {
  return fileDialog
}