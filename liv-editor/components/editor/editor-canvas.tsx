"use client"

import { useState } from "react"
import { IconButton } from "@/components/ui/icon-button"
import { Toolbar } from "@/components/ui/toolbar"
import { ZoomIn, ZoomOut, Grid3x3, Maximize2, Type, Square, Circle } from "lucide-react"
import { useButtonActions } from "@/hooks/use-button-actions"

interface EditorCanvasProps {
  onSelectElement: (id: string | null) => void
}

export function EditorCanvas({ onSelectElement }: EditorCanvasProps) {
  const [zoom, setZoom] = useState(100)
  const [showGrid, setShowGrid] = useState(true)
  const { executeAction } = useButtonActions()

  const handleZoomIn = async () => {
    const newZoom = Math.min(200, zoom + 10)
    setZoom(newZoom)
  }

  const handleZoomOut = async () => {
    const newZoom = Math.max(25, zoom - 10)
    setZoom(newZoom)
  }

  const handleAddText = async () => {
    await executeAction(async () => {
      console.log("[v0] Adding text element")
      onSelectElement("text-element")
    }, "add-text")
  }

  const handleAddShape = async () => {
    await executeAction(async () => {
      console.log("[v0] Adding shape element")
      onSelectElement("shape-element")
    }, "add-shape")
  }

  const handleFullscreen = async () => {
    await executeAction(async () => {
      const canvas = document.querySelector(".canvas-area")
      if (canvas?.requestFullscreen) {
        canvas.requestFullscreen()
      }
    }, "fullscreen")
  }

  return (
    <div className="flex-1 flex flex-col bg-background relative overflow-hidden">
      {/* Canvas Toolbar */}
      <Toolbar className="bg-gradient-to-r from-background to-muted/50 border-b border-border/50">
        <div className="flex items-center gap-2">
          <IconButton size="sm" onClick={handleZoomOut} className="hover:bg-accent/20">
            <ZoomOut size={18} />
          </IconButton>
          <span className="text-sm font-medium w-12 text-center bg-muted/50 px-2 py-1 rounded transition-all hover:bg-accent/20">
            {zoom}%
          </span>
          <IconButton size="sm" onClick={handleZoomIn} className="hover:bg-accent/20">
            <ZoomIn size={18} />
          </IconButton>
        </div>

        <div className="flex-1" />

        <div className="flex items-center gap-1 border-r border-border/50 pr-4">
          <IconButton size="sm" title="Text" onClick={handleAddText} className="hover:bg-accent/20">
            <Type size={18} />
          </IconButton>
          <IconButton size="sm" title="Rectangle" onClick={handleAddShape} className="hover:bg-accent/20">
            <Square size={18} />
          </IconButton>
          <IconButton size="sm" title="Circle" onClick={handleAddShape} className="hover:bg-accent/20">
            <Circle size={18} />
          </IconButton>
        </div>

        <div className="flex items-center gap-2">
          <IconButton
            size="sm"
            variant={showGrid ? "default" : "ghost"}
            onClick={() => setShowGrid(!showGrid)}
            className="transition-all"
          >
            <Grid3x3 size={18} />
          </IconButton>
          <IconButton size="sm" onClick={handleFullscreen} className="hover:bg-accent/20">
            <Maximize2 size={18} />
          </IconButton>
        </div>
      </Toolbar>

      {/* Canvas Area */}
      <div className="flex-1 overflow-auto relative bg-gradient-to-br from-background to-muted/20 canvas-area">
        {/* Grid Background */}
        {showGrid && (
          <div
            className="absolute inset-0 opacity-5 transition-opacity duration-300"
            style={{
              backgroundImage:
                "linear-gradient(0deg, transparent 24%, rgba(255,255,255,.05) 25%, rgba(255,255,255,.05) 26%, transparent 27%, transparent 74%, rgba(255,255,255,.05) 75%, rgba(255,255,255,.05) 76%, transparent 77%, transparent), linear-gradient(90deg, transparent 24%, rgba(255,255,255,.05) 25%, rgba(255,255,255,.05) 26%, transparent 27%, transparent 74%, rgba(255,255,255,.05) 75%, rgba(255,255,255,.05) 76%, transparent 77%, transparent)",
              backgroundSize: "50px 50px",
            }}
          />
        )}

        {/* Canvas Content */}
        <div className="relative p-12 flex items-center justify-center min-h-full">
          <div
            className="bg-gradient-to-br from-muted to-muted/50 border-2 border-dashed border-border rounded-lg p-12 text-center cursor-pointer hover:border-accent hover:shadow-lg hover:shadow-accent/30 transition-all duration-300 scale-in"
            style={{ width: "800px", height: "600px" }}
            onClick={() => onSelectElement("canvas")}
          >
            <div className="text-muted-foreground">
              <p className="text-lg font-medium mb-2">Start editing</p>
              <p className="text-sm">Click to add text, shapes, or media</p>
            </div>
          </div>
        </div>
      </div>
    </div>
  )
}
