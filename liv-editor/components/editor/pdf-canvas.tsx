"use client"
import { Type, Plus } from "lucide-react"
import type React from "react"

import { useState, useRef, useEffect, useCallback } from "react"

interface PDFCanvasProps {
  onSelectElement: (id: string | null) => void
  onAddTextBox?: (x: number, y: number) => void
  zoom: number
  showGrid: boolean
  pageSize: keyof typeof PAGE_SIZES
  rotation: number
  elements?: Array<{ id: string; type: string; x: number; y: number }>
}

const PAGE_SIZES = {
  A4: { width: 800, height: 1000, label: "A4" },
  Letter: { width: 850, height: 1100, label: "Letter" },
  A3: { width: 1100, height: 1550, label: "A3" },
  A5: { width: 550, height: 700, label: "A5" },
}

export function PDFCanvas({
  onSelectElement,
  onAddTextBox,
  zoom,
  showGrid,
  pageSize,
  rotation,
  elements = [],
}: PDFCanvasProps) {
  const currentPageSize = PAGE_SIZES[pageSize]
  const [contextMenu, setContextMenu] = useState<{ x: number; y: number; action?: string } | null>(null)
  const [pages, setPages] = useState([{ id: 'page-1', content: '' }])
  const [currentPage, setCurrentPage] = useState(0)
  const editorRefs = useRef<Array<HTMLDivElement | null>>([])
  const [isTyping, setIsTyping] = useState(false)
  const [savedSelection, setSavedSelection] = useState<{range: Range, pageIndex: number} | null>(null)
  const [activeFormats, setActiveFormats] = useState({ bold: false, italic: false, underline: false })
  const [toolbarFontSize, setToolbarFontSize] = useState(12)
  const [toolbarFontFamily, setToolbarFontFamily] = useState('Times New Roman, serif')
  const [documentSettings, setDocumentSettings] = useState({
    marginTop: 72,    // 1 inch = 72px
    marginBottom: 72,
    marginLeft: 72,
    marginRight: 72,
    lineHeight: 1.5,
    fontSize: 12,
    fontFamily: 'Times New Roman, serif'
  })

  const addNewPage = useCallback(() => {
    const newPage = { id: `page-${pages.length + 1}`, content: '' }
    setPages(prev => {
      const updated = [...prev, newPage]
      // Focus the new page after it's added
      setTimeout(() => {
        const newPageIndex = updated.length - 1
        const editor = editorRefs.current[newPageIndex]
        if (editor) {
          editor.focus()
          setCurrentPage(newPageIndex)
        }
      }, 100)
      return updated
    })
  }, [pages.length])

  const handleInput = useCallback((pageIndex: number, content: string) => {
    setPages(prev => prev.map((page, index) => 
      index === pageIndex ? { ...page, content } : page
    ))
    setIsTyping(true)
    
    // Clear typing indicator after a delay
    setTimeout(() => setIsTyping(false), 1000)
    
    // Auto-add new page if content overflows
    const editor = editorRefs.current[pageIndex]
    if (editor && editor.scrollHeight > editor.clientHeight * 0.9) {
      // Check if we're near the bottom and need a new page
      const selection = window.getSelection()
      if (selection && selection.rangeCount > 0) {
        const range = selection.getRangeAt(0)
        const rect = range.getBoundingClientRect()
        const editorRect = editor.getBoundingClientRect()
        
        if (rect.bottom > editorRect.bottom - 50) {
          setTimeout(addNewPage, 100)
        }
      }
    }
  }, [addNewPage])

  const focusEditor = useCallback((pageIndex: number) => {
    const editor = editorRefs.current[pageIndex]
    if (editor) {
      editor.focus()
      setCurrentPage(pageIndex)
    }
  }, [])

  const updatePageContentFromDOM = useCallback((pageIndex: number) => {
    const editor = editorRefs.current[pageIndex]
    if (editor) {
      const content = editor.innerHTML
      setPages(prev => prev.map((page, index) => 
        index === pageIndex ? { ...page, content } : page
      ))
    }
  }, [])

  // Check what formatting is active at cursor/selection
  const checkActiveFormats = useCallback(() => {
    const selection = window.getSelection()
    if (selection && selection.rangeCount > 0) {
      const range = selection.getRangeAt(0)
      const container = range.commonAncestorContainer
      const element = container.nodeType === Node.TEXT_NODE ? container.parentElement : container as Element
      
      if (element) {
        const computedStyle = window.getComputedStyle(element)
        
        // Check formatting
        const bold = !!element.closest('strong, b') || computedStyle.fontWeight === 'bold' || parseInt(computedStyle.fontWeight) >= 700
        const italic = !!element.closest('em, i') || computedStyle.fontStyle === 'italic'
        const underline = !!element.closest('u') || computedStyle.textDecoration.includes('underline')
        
        // Get font size (remove 'px' and convert to number)
        const fontSize = parseInt(computedStyle.fontSize) || documentSettings.fontSize
        
        // Get font family
        const fontFamily = computedStyle.fontFamily || documentSettings.fontFamily
        
        setActiveFormats({ bold, italic, underline })
        setToolbarFontSize(fontSize)
        setToolbarFontFamily(fontFamily)
      }
    } else {
      setActiveFormats({ bold: false, italic: false, underline: false })
      setToolbarFontSize(documentSettings.fontSize)
      setToolbarFontFamily(documentSettings.fontFamily)
    }
  }, [documentSettings.fontSize, documentSettings.fontFamily])

  // Save current selection
  const saveSelection = useCallback(() => {
    const selection = window.getSelection()
    if (selection && selection.rangeCount > 0 && !selection.isCollapsed) {
      const range = selection.getRangeAt(0).cloneRange()
      setSavedSelection({ range, pageIndex: currentPage })
      checkActiveFormats()
    }
  }, [currentPage, checkActiveFormats])

  // Restore saved selection
  const restoreSelection = useCallback(() => {
    if (savedSelection && savedSelection.pageIndex === currentPage) {
      const selection = window.getSelection()
      if (selection) {
        selection.removeAllRanges()
        selection.addRange(savedSelection.range)
      }
    }
  }, [savedSelection, currentPage])

  // Get current selection or use saved selection for formatting
  const getActiveSelection = useCallback(() => {
    const selection = window.getSelection()
    if (selection && selection.rangeCount > 0 && !selection.isCollapsed) {
      return selection
    } else if (savedSelection && savedSelection.pageIndex === currentPage) {
      const selection = window.getSelection()
      if (selection) {
        selection.removeAllRanges()
        selection.addRange(savedSelection.range)
        return selection
      }
    }
    return null
  }, [savedSelection, currentPage])

  // Helper function to apply formatting to selected text only
  const applyFormatToSelection = useCallback((formatType: 'bold' | 'italic' | 'underline') => {
    const selection = getActiveSelection()
    if (selection && selection.rangeCount > 0 && !selection.isCollapsed) {
      try {
        const range = selection.getRangeAt(0)
        
        // Check if the selection is already formatted
        const container = range.commonAncestorContainer
        const parentElement = container.nodeType === Node.TEXT_NODE ? container.parentElement : container as Element
        
        const tagName = formatType === 'bold' ? 'STRONG' : formatType === 'italic' ? 'EM' : 'U'
        let formattedAncestor = parentElement?.closest(tagName.toLowerCase()) as HTMLElement
        
        if (formattedAncestor && range.toString() === formattedAncestor.textContent) {
          // Remove formatting if already applied to the entire selection
          const parent = formattedAncestor.parentNode
          if (parent) {
            // Replace formatted element with its text content
            const textNode = document.createTextNode(formattedAncestor.textContent || '')
            parent.replaceChild(textNode, formattedAncestor)
            
            // Restore selection to the text
            const newRange = document.createRange()
            newRange.selectNodeContents(textNode)
            selection.removeAllRanges()
            selection.addRange(newRange)
          }
        } else {
          // Apply formatting
          const selectedContent = range.extractContents()
          
          let element: HTMLElement
          switch (formatType) {
            case 'bold':
              element = document.createElement('strong')
              break
            case 'italic':
              element = document.createElement('em')
              break
            case 'underline':
              element = document.createElement('u')
              break
          }
          
          element.appendChild(selectedContent)
          range.insertNode(element)
          
          // Restore selection
          selection.removeAllRanges()
          const newRange = document.createRange()
          newRange.selectNodeContents(element)
          selection.addRange(newRange)
        }

        // **** THIS IS THE FIX: Sync DOM back to React state ****
        updatePageContentFromDOM(currentPage)

        // Save the new selection
        saveSelection()

      } catch (error) {
        console.log(`${formatType} application failed:`, error)
      }
    }
  }, [currentPage, updatePageContentFromDOM, getActiveSelection, saveSelection])

  const handleKeyDown = useCallback((e: React.KeyboardEvent, pageIndex: number) => {
    // Handle common keyboard shortcuts
    if (e.ctrlKey || e.metaKey) {
      switch (e.key) {
        case 'b':
          e.preventDefault()
          applyFormatToSelection('bold')
          break
        case 'i':
          e.preventDefault()
          applyFormatToSelection('italic')
          break
        case 'u':
          e.preventDefault()
          applyFormatToSelection('underline')
          break
        case 'z':
          e.preventDefault()
          document.execCommand('undo')
          break
        case 'y':
          e.preventDefault()
          document.execCommand('redo')
          break
        case 'a':
          e.preventDefault()
          document.execCommand('selectAll')
          break
      }
    }

    // Handle Enter key for new page
    if (e.key === 'Enter' && e.ctrlKey) {
      e.preventDefault()
      addNewPage()
    }

    // Let the browser handle normal typing naturally
  }, [addNewPage, applyFormatToSelection])

  useEffect(() => {
    // Focus first page editor on mount
    if (editorRefs.current[0]) {
      editorRefs.current[0].focus()
    }
    
    // *** ADD THIS PART ***
    // Set initial content for all pages ONCE on mount
    pages.forEach((page, index) => {
      if (editorRefs.current[index] && !editorRefs.current[index].innerHTML) {
          editorRefs.current[index].innerHTML = page.content;
      }
    })
    // *** END OF ADDITION ***

  }, [])

  return (
    <div className="flex-1 flex flex-col bg-gray-100 relative overflow-hidden h-full">
      {/* Document Editor Area */}
      <div 
        className="flex-1 overflow-auto relative"
        onClick={(e) => {
          // Clear selection if clicking on background (not on text areas)
          if (e.target === e.currentTarget) {
            setSavedSelection(null)
            const selection = window.getSelection()
            if (selection) {
              selection.removeAllRanges()
            }
          }
        }}
      >
        {/* Grid Background */}
        {showGrid && (
          <div className="absolute inset-0 opacity-20 pointer-events-none grid-pattern" />
        )}

        {/* Document Container */}
        <div className="py-8 px-4 flex flex-col items-center min-h-full">
          {pages.map((page, pageIndex) => (
            <div key={page.id} className="mb-6 relative group">
              {/* Page Shadow */}
              <div className="absolute inset-0 bg-black/10 rounded-sm transform translate-x-1 translate-y-1" />
              
              {/* Page Content */}
              <div
                className="bg-white border border-gray-300 shadow-lg relative overflow-hidden"
                style={{
                  width: `${currentPageSize.width * (zoom / 100)}px`,
                  height: `${currentPageSize.height * (zoom / 100)}px`,
                  transform: `rotate(${rotation}deg)`,
                }}
                onClick={() => {
                  focusEditor(pageIndex)
                  setTimeout(checkActiveFormats, 10)
                }}
              >
                {/* Page Margins Guide */}
                <div 
                  className="absolute inset-0 border border-dashed border-blue-200 opacity-50 pointer-events-none"
                  style={{
                    margin: `${documentSettings.marginTop * (zoom / 100)}px ${documentSettings.marginRight * (zoom / 100)}px ${documentSettings.marginBottom * (zoom / 100)}px ${documentSettings.marginLeft * (zoom / 100)}px`
                  }}
                />

                {/* Editable Content Area */}
                <div
                  ref={(el) => { editorRefs.current[pageIndex] = el }}
                  contentEditable
                  suppressContentEditableWarning
                  className="w-full h-full outline-none cursor-text p-0 overflow-hidden"
                  style={{
                    padding: `${documentSettings.marginTop * (zoom / 100)}px ${documentSettings.marginRight * (zoom / 100)}px ${documentSettings.marginBottom * (zoom / 100)}px ${documentSettings.marginLeft * (zoom / 100)}px`,
                    fontSize: `${documentSettings.fontSize * (zoom / 100)}px`,
                    fontFamily: documentSettings.fontFamily,
                    lineHeight: documentSettings.lineHeight,
                    color: '#000',
                    minHeight: '100%',
                    wordWrap: 'break-word',
                    whiteSpace: 'pre-wrap',
                    direction: 'ltr',
                    textAlign: 'left',
                    unicodeBidi: 'plaintext',
                    writingMode: 'horizontal-tb'
                  }}
                  onInput={(e) => {
                    const content = e.currentTarget.innerHTML
                    handleInput(pageIndex, content)
                  }}
                  onKeyDown={(e) => handleKeyDown(e, pageIndex)}
                  onFocus={() => setCurrentPage(pageIndex)}
                  onMouseUp={saveSelection}
                  onKeyUp={saveSelection}
                  onPaste={(e) => {
                    // Handle paste events to maintain proper formatting
                    e.preventDefault()
                    const text = e.clipboardData.getData('text/plain')
                    document.execCommand('insertText', false, text)
                  }}
                  spellCheck="false"
                  dir="ltr"
                  lang="en"
                />

                {/* Page Number */}
                <div className="absolute bottom-2 right-4 text-xs text-gray-500 pointer-events-none">
                  Page {pageIndex + 1}
                </div>

                {/* Page Break Indicator */}
                {pageIndex < pages.length - 1 && (
                  <div className="absolute -bottom-3 left-1/2 transform -translate-x-1/2 text-xs text-gray-400 bg-white px-2 rounded">
                    Page Break
                  </div>
                )}
              </div>

              {/* Add Page Button */}
              {pageIndex === pages.length - 1 && (
                <button
                  onClick={addNewPage}
                  className="absolute -bottom-8 left-1/2 transform -translate-x-1/2 bg-blue-500 hover:bg-blue-600 text-white rounded-full p-2 shadow-lg transition-colors opacity-0 group-hover:opacity-100"
                  title="Add new page"
                >
                  <Plus size={16} />
                </button>
              )}
            </div>
          ))}

          {/* Document Stats */}
          <div className="mt-4 text-xs text-gray-500 text-center space-y-1">
            <div>Pages: {pages.length} | Current: {currentPage + 1}</div>
            <div>Zoom: {zoom}% | {pageSize} ({currentPageSize.width} × {currentPageSize.height}px)</div>
            {isTyping && <div className="text-green-600">● Typing...</div>}
          </div>
        </div>

        {/* Formatting Toolbar (Floating) */}
        <div 
          className="fixed bottom-4 left-1/2 transform -translate-x-1/2 bg-white border border-gray-300 rounded-lg shadow-lg px-4 py-2 flex items-center gap-2 z-50"
          onMouseDown={(e) => e.preventDefault()}
        >
          <div className="text-xs text-gray-500 mr-2">Selection Formatting:</div>
          {/* Font Size Input - applies to selection only */}
          <div className="flex items-center space-x-1">
            <label htmlFor="font-size-input" className="text-xs text-gray-600">Size:</label>
            <input
              id="font-size-input"
              type="number"
              min="8"
              max="72"
              value={toolbarFontSize}
              onChange={(e) => {
                const newSize = parseInt(e.target.value) || 12;
                setToolbarFontSize(newSize);
                
                const selection = getActiveSelection(); // Use your helper
                if (selection && selection.rangeCount > 0 && !selection.isCollapsed) {
                  // SELECTION EXISTS: Apply span to selection only
                  try {
                    const range = selection.getRangeAt(0);
                    const selectedContent = range.extractContents();
                    
                    const span = document.createElement('span');
                    span.style.fontSize = `${newSize}px`;
                    span.appendChild(selectedContent);
                    
                    range.insertNode(span);

                    // Restore selection to the new content
                    selection.removeAllRanges();
                    const newRange = document.createRange();
                    newRange.selectNodeContents(span);
                    selection.addRange(newRange);
                    
                    updatePageContentFromDOM(currentPage);
                    saveSelection(); // Use your helper
                  } catch (error) {
                    console.error("Failed to apply font size:", error);
                  }
                } else {
                  // NO SELECTION: Update the default document settings
                  setDocumentSettings(prev => ({ ...prev, fontSize: newSize }));
                }
              }}
              className="w-16 text-xs border border-gray-300 rounded px-2 py-1 text-center"
            />
            <span className="text-xs text-gray-500">px</span>
          </div>

          {/* Font Family Dropdown - applies to selection or sets default */}
          <select
            className="text-xs border border-gray-300 rounded px-2 py-1 min-w-[90px]"
            title="Font Family"
            value={toolbarFontFamily}
            onChange={(e) => {
              const fontFamily = e.target.value;
              setToolbarFontFamily(fontFamily);

              const selection = getActiveSelection(); // Use your helper
              if (selection && selection.rangeCount > 0 && !selection.isCollapsed) {
                // SELECTION EXISTS: Apply span to selection only
                try {
                  const range = selection.getRangeAt(0);
                  const selectedContent = range.extractContents();
                  
                  const span = document.createElement('span');
                  span.style.fontFamily = fontFamily;
                  span.appendChild(selectedContent);
                  
                  range.insertNode(span);

                  // Restore selection to the new content
                  selection.removeAllRanges();
                  const newRange = document.createRange();
                  newRange.selectNodeContents(span);
                  selection.addRange(newRange);

                  updatePageContentFromDOM(currentPage);
                  saveSelection(); // Use your helper
                } catch (error) {
                  console.error("Failed to apply font family:", error);
                }
              } else {
                // NO SELECTION: Update the default document settings
                setDocumentSettings(prev => ({ ...prev, fontFamily }));
              }
            }}
          >
            <option value="Times New Roman, serif">Times</option>
            <option value="Arial, sans-serif">Arial</option>
            <option value="Helvetica, sans-serif">Helvetica</option>
            <option value="Georgia, serif">Georgia</option>
            <option value="Courier New, monospace">Courier</option>
          </select>
          <div className="w-px h-4 bg-gray-300" />
          <button
            onClick={() => applyFormatToSelection('bold')}
            className={`px-3 py-1 text-sm font-bold rounded transition-colors ${
              activeFormats.bold 
                ? 'bg-blue-500 text-white hover:bg-blue-600' 
                : 'hover:bg-gray-100'
            }`}
            title="Bold (Ctrl+B)"
          >
            B
          </button>
          <button
            onClick={() => applyFormatToSelection('italic')}
            className={`px-3 py-1 text-sm italic rounded transition-colors ${
              activeFormats.italic 
                ? 'bg-blue-500 text-white hover:bg-blue-600' 
                : 'hover:bg-gray-100'
            }`}
            title="Italic (Ctrl+I)"
          >
            I
          </button>
          <button
            onClick={() => applyFormatToSelection('underline')}
            className={`px-3 py-1 text-sm underline rounded transition-colors ${
              activeFormats.underline 
                ? 'bg-blue-500 text-white hover:bg-blue-600' 
                : 'hover:bg-gray-100'
            }`}
            title="Underline (Ctrl+U)"
          >
            U
          </button>
          <div className="w-px h-4 bg-gray-300" />
          <button
            onClick={() => document.execCommand('justifyLeft')}
            className="px-2 py-1 text-sm hover:bg-gray-100 rounded transition-colors"
            title="Align Left"
          >
            ⟵
          </button>
          <button
            onClick={() => document.execCommand('justifyCenter')}
            className="px-2 py-1 text-sm hover:bg-gray-100 rounded transition-colors"
            title="Center"
          >
            ↔
          </button>
          <button
            onClick={() => document.execCommand('justifyRight')}
            className="px-2 py-1 text-sm hover:bg-gray-100 rounded transition-colors"
            title="Align Right"
          >
            ⟶
          </button>
          <div className="w-px h-4 bg-gray-300" />
          <button
            onClick={addNewPage}
            className="px-3 py-1 text-sm bg-blue-500 text-white hover:bg-blue-600 rounded transition-colors"
            title="New Page (Ctrl+Enter)"
          >
            + Page
          </button>
        </div>
      </div>

      <style jsx>{`
        .grid-pattern {
          background-image: 
            linear-gradient(rgba(0,0,0,.1) 1px, transparent 1px),
            linear-gradient(90deg, rgba(0,0,0,.1) 1px, transparent 1px);
          background-size: 20px 20px;
        }
        
        [contenteditable]:focus {
          outline: none;
        }
        
        [contenteditable] p {
          margin: 0;
          padding: 0;
          line-height: inherit;
        }
        
        [contenteditable] br {
          line-height: inherit;
        }
      `}</style>
    </div>
  )
}
