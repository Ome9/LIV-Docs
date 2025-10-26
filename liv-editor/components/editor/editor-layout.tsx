"use client"

import { useState } from "react"
import { EditorNavbar } from "./editor-navbar"
import { EditorSidebar } from "./editor-sidebar"
import { PDFCanvas } from "./pdf-canvas"
import { PropertiesPanel } from "./properties-panel"
import { CodeEditor } from "./code-editor"
import { ComponentLibrary } from "./component-library"
import {
  ChevronLeft,
  ZoomIn,
  ZoomOut,
  Grid3x3,
  Maximize2,
  Type,
  Square,
  Circle,
  Code,
  Layers,
  RotateCw,
  Copy,
  Trash2,
} from "lucide-react"
import { IconButton } from "@/components/ui/icon-button"
import { Toolbar } from "@/components/ui/toolbar"
import { Button } from "@/components/ui/button"

const PROPERTIES_PANEL_MIN_WIDTH = 280
const PROPERTIES_PANEL_MAX_WIDTH = 600
const CODE_EDITOR_MIN_WIDTH = 200
const CODE_EDITOR_MAX_WIDTH = 500
const COMPONENT_LIBRARY_MIN_WIDTH = 200
const COMPONENT_LIBRARY_MAX_WIDTH = 500

const PAGE_SIZES = {
  A4: { width: 800, height: 1000, label: "A4" },
  Letter: { width: 850, height: 1100, label: "Letter" },
  A3: { width: 1100, height: 1550, label: "A3" },
  A5: { width: 550, height: 700, label: "A5" },
}

export function EditorLayout() {
  const [selectedElement, setSelectedElement] = useState<string | null>(null)
  const [sidebarTab, setSidebarTab] = useState<"pages" | "assets" | "animations" | "settings">("pages")
  const [showCodeEditor, setShowCodeEditor] = useState(false)
  const [showComponentLibrary, setShowComponentLibrary] = useState(true)
  const [codeEditorWidth, setCodeEditorWidth] = useState(300)
  const [componentLibraryWidth, setComponentLibraryWidth] = useState(280)
  const [propertiesPanelWidth, setPropertiesPanelWidth] = useState(320)
  const [autoAdjustProperties, setAutoAdjustProperties] = useState(true)
  const [zoom, setZoom] = useState(100)
  const [showGrid, setShowGrid] = useState(true)
  const [pageSize, setPageSize] = useState<keyof typeof PAGE_SIZES>("A4")
  const [rotation, setRotation] = useState(0)
  const [elements, setElements] = useState<Array<{ id: string; type: string; x: number; y: number }>>([])

  const handleSelectElement = (elementId: string | null) => {
    if (elementId !== "canvas") {
      setSelectedElement(elementId)
      if (autoAdjustProperties && elementId) {
        const baseWidth = 320
        const adjustedWidth = Math.min(PROPERTIES_PANEL_MAX_WIDTH, baseWidth + Math.random() * 100)
        setPropertiesPanelWidth(adjustedWidth)
      }
    }
  }

  const handleAddTextBox = (x: number, y: number) => {
    const textElementId = `text-${Date.now()}`
    console.log("[v0] Creating text box at", x, y)
    setElements([...elements, { id: textElementId, type: "text", x, y }])
    setSelectedElement(textElementId)
  }

  const handleZoomIn = () => {
    setZoom((z) => Math.min(200, z + 10))
  }

  const handleZoomOut = () => {
    setZoom((z) => Math.max(25, z - 10))
  }

  const handleAddText = () => {
    console.log("[v0] Adding text element")
    const textElementId = `text-${Date.now()}`
    setElements([...elements, { id: textElementId, type: "text", x: 100, y: 100 }])
    setSelectedElement(textElementId)
  }

  const handleAddShape = () => {
    console.log("[v0] Adding shape element")
    const shapeElementId = `shape-${Date.now()}`
    setElements([...elements, { id: shapeElementId, type: "shape", x: 150, y: 150 }])
    setSelectedElement(shapeElementId)
  }

  const handleAddCircle = () => {
    console.log("[v0] Adding circle element")
    const circleElementId = `circle-${Date.now()}`
    setElements([...elements, { id: circleElementId, type: "circle", x: 200, y: 200 }])
    setSelectedElement(circleElementId)
  }

  const handleFullscreen = () => {
    const canvas = document.querySelector(".canvas-area")
    if (canvas?.requestFullscreen) {
      canvas.requestFullscreen()
    }
  }

  const handleRotate = () => {
    setRotation((r) => (r + 90) % 360)
  }

  const handleDuplicate = () => {
    console.log("[v0] Duplicating element")
    if (selectedElement) {
      const element = elements.find((el) => el.id === selectedElement)
      if (element) {
        const newElement = { ...element, id: `${element.type}-${Date.now()}` }
        setElements([...elements, newElement])
        setSelectedElement(newElement.id)
      }
    }
  }

  const handleDelete = () => {
    console.log("[v0] Deleting element")
    if (selectedElement) {
      setElements(elements.filter((el) => el.id !== selectedElement))
      setSelectedElement(null)
    }
  }

  const createResizeDivider = (
    currentWidth: number,
    setWidth: (width: number) => void,
    minWidth: number,
    maxWidth: number,
    isRightResize = false,
  ) => (
    <div
      className="w-1 bg-border hover:bg-accent/40 transition-colors cursor-col-resize shrink-0"
      onMouseDown={(e) => {
        const startX = e.clientX
        const startWidth = currentWidth
        const handleMouseMove = (moveEvent: MouseEvent) => {
          const delta = moveEvent.clientX - startX
          const newWidth = isRightResize ? startWidth + delta : startWidth - delta
          const constrainedWidth = Math.max(minWidth, Math.min(maxWidth, newWidth))
          setWidth(constrainedWidth)
        }
        const handleMouseUp = () => {
          document.removeEventListener("mousemove", handleMouseMove)
          document.removeEventListener("mouseup", handleMouseUp)
        }
        document.addEventListener("mousemove", handleMouseMove)
        document.addEventListener("mouseup", handleMouseUp)
      }}
    />
  )

  return (
    <div className="flex flex-col h-screen bg-background">
      <EditorNavbar />

      <Toolbar className="bg-card border-b border-border/50 justify-between shrink-0">
        <div className="flex items-center gap-2">
          <IconButton size="sm" onClick={handleZoomOut} className="hover:bg-muted transition-colors" title="Zoom out">
            <ZoomOut size={18} />
          </IconButton>
          <span className="text-sm font-medium w-12 text-center bg-muted/50 px-2 py-1 rounded transition-all hover:bg-muted">
            {zoom}%
          </span>
          <IconButton size="sm" onClick={handleZoomIn} className="hover:bg-muted transition-colors" title="Zoom in">
            <ZoomIn size={18} />
          </IconButton>

          <div className="border-l border-border/50 pl-4 ml-2">
            <select
              value={pageSize}
              onChange={(e) => setPageSize(e.target.value as keyof typeof PAGE_SIZES)}
              className="text-sm bg-muted/50 border border-border/50 rounded px-2 py-1 outline-none hover:bg-muted transition-colors"
              title="Page Size"
            >
              {Object.entries(PAGE_SIZES).map(([key, value]) => (
                <option key={key} value={key}>
                  {value.label}
                </option>
              ))}
            </select>
          </div>
        </div>

        <div className="flex items-center gap-1 border-r border-border/50 pr-4">
          <IconButton size="sm" title="Add Text" onClick={handleAddText} className="hover:bg-muted transition-colors">
            <Type size={18} />
          </IconButton>
          <IconButton
            size="sm"
            title="Add Rectangle"
            onClick={handleAddShape}
            className="hover:bg-muted transition-colors"
          >
            <Square size={18} />
          </IconButton>
          <IconButton
            size="sm"
            title="Add Circle"
            onClick={handleAddCircle}
            className="hover:bg-muted transition-colors"
          >
            <Circle size={18} />
          </IconButton>

          <IconButton size="sm" title="Rotate" onClick={handleRotate} className="hover:bg-muted transition-colors">
            <RotateCw size={18} />
          </IconButton>
          <IconButton
            size="sm"
            title="Duplicate"
            onClick={handleDuplicate}
            className="hover:bg-muted transition-colors"
          >
            <Copy size={18} />
          </IconButton>
          <IconButton size="sm" title="Delete" onClick={handleDelete} className="hover:bg-muted transition-colors">
            <Trash2 size={18} />
          </IconButton>
        </div>

        <div className="flex items-center gap-2">
          <Button
            size="sm"
            variant={showGrid ? "default" : "outline"}
            onClick={() => setShowGrid(!showGrid)}
            className="gap-2 text-xs"
          >
            <Grid3x3 size={16} />
            Grid
          </Button>
          <Button
            size="sm"
            variant={showCodeEditor ? "default" : "outline"}
            onClick={() => setShowCodeEditor(!showCodeEditor)}
            className="gap-2 text-xs"
          >
            <Code size={16} />
            Code
          </Button>
          <Button
            size="sm"
            variant={showComponentLibrary ? "default" : "outline"}
            onClick={() => setShowComponentLibrary(!showComponentLibrary)}
            className="gap-2 text-xs"
          >
            <Layers size={16} />
            Tools
          </Button>
          <IconButton
            size="sm"
            onClick={handleFullscreen}
            className="hover:bg-muted transition-colors"
            title="Fullscreen"
          >
            <Maximize2 size={18} />
          </IconButton>
        </div>
      </Toolbar>

      <div className="flex flex-1 overflow-hidden min-h-0">
        {/* Left Sidebar */}
        <EditorSidebar activeTab={sidebarTab} onTabChange={setSidebarTab} />

        {/* Main Canvas Area */}
        <div className="flex-1 flex flex-col overflow-hidden min-w-0">
          <div className="flex-1 overflow-hidden min-h-0">
            <PDFCanvas
              onSelectElement={handleSelectElement}
              onAddTextBox={handleAddTextBox}
              zoom={zoom}
              showGrid={showGrid}
              pageSize={pageSize}
              rotation={rotation}
              elements={elements}
            />
          </div>
        </div>

        {createResizeDivider(
          propertiesPanelWidth,
          setPropertiesPanelWidth,
          PROPERTIES_PANEL_MIN_WIDTH,
          PROPERTIES_PANEL_MAX_WIDTH,
          false,
        )}

        {/* Right Panels Container */}
        <div className="flex gap-0 border-l border-border overflow-hidden">
          {/* Properties Panel */}
          <div
            style={{ width: `${propertiesPanelWidth}px`, minWidth: `${PROPERTIES_PANEL_MIN_WIDTH}px` }}
            className="flex flex-col border-r border-border overflow-hidden shrink-0"
          >
            <PropertiesPanel
              selectedElement={selectedElement}
              panelWidth={propertiesPanelWidth}
              autoAdjust={autoAdjustProperties}
              onAutoAdjustChange={setAutoAdjustProperties}
            />
          </div>

          {/* Code Editor */}
          {showCodeEditor && (
            <>
              {createResizeDivider(
                codeEditorWidth,
                setCodeEditorWidth,
                CODE_EDITOR_MIN_WIDTH,
                CODE_EDITOR_MAX_WIDTH,
                true,
              )}
              <div
                style={{ width: `${codeEditorWidth}px`, minWidth: `${CODE_EDITOR_MIN_WIDTH}px` }}
                className="flex flex-col border-r border-border overflow-hidden shrink-0"
              >
                <CodeEditor onRun={(code) => console.log("[v0] Running code:", code)} />
              </div>
            </>
          )}

          {/* Component Library Dock */}
          {showComponentLibrary ? (
            <>
              {createResizeDivider(
                componentLibraryWidth,
                setComponentLibraryWidth,
                COMPONENT_LIBRARY_MIN_WIDTH,
                COMPONENT_LIBRARY_MAX_WIDTH,
                true,
              )}
              <div
                style={{ width: `${componentLibraryWidth}px`, minWidth: `${COMPONENT_LIBRARY_MIN_WIDTH}px` }}
                className="flex flex-col border-l border-border overflow-hidden shrink-0"
              >
                <ComponentLibrary
                  onAddComponent={(id) => console.log("[v0] Adding component:", id)}
                  onMinimize={() => setShowComponentLibrary(false)}
                />
              </div>
            </>
          ) : (
            <button
              onClick={() => setShowComponentLibrary(true)}
              className="w-12 flex items-center justify-center bg-muted hover:bg-muted/80 border-l border-border transition-colors shrink-0"
              title="Show component library"
            >
              <ChevronLeft size={18} />
            </button>
          )}
        </div>
      </div>
    </div>
  )
}
