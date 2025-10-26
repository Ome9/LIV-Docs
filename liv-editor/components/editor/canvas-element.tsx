"use client"

import type React from "react"

import { useState, useRef } from "react"
import { Copy, Trash2 } from "lucide-react"

interface CanvasElementProps {
  id: string
  type: string
  x: number
  y: number
  width?: number
  height?: number
  rotation?: number
  isSelected?: boolean
  onSelect: (id: string) => void
  onUpdate: (id: string, updates: any) => void
  onDelete: (id: string) => void
  onDuplicate: (id: string) => void
}

export function CanvasElement({
  id,
  type,
  x,
  y,
  width = 120,
  height = 60,
  rotation = 0,
  isSelected = false,
  onSelect,
  onUpdate,
  onDelete,
  onDuplicate,
}: CanvasElementProps) {
  const [isDragging, setIsDragging] = useState(false)
  const [dragStart, setDragStart] = useState({ x: 0, y: 0 })
  const [resizeHandle, setResizeHandle] = useState<string | null>(null)
  const elementRef = useRef<HTMLDivElement>(null)

  const handleMouseDown = (e: React.MouseEvent) => {
    if ((e.target as HTMLElement).closest("[data-resize]")) return
    e.stopPropagation()
    setIsDragging(true)
    setDragStart({ x: e.clientX - x, y: e.clientY - y })
    onSelect(id)
  }

  const handleMouseMove = (e: React.MouseEvent) => {
    if (!isDragging) return
    const newX = e.clientX - dragStart.x
    const newY = e.clientY - dragStart.y
    onUpdate(id, { x: newX, y: newY })
  }

  const handleMouseUp = () => {
    setIsDragging(false)
  }

  const handleResizeStart = (e: React.MouseEvent, handle: string) => {
    e.stopPropagation()
    setResizeHandle(handle)
    setDragStart({ x: e.clientX, y: e.clientY })
  }

  const handleResizeMove = (e: React.MouseEvent) => {
    if (!resizeHandle) return
    const deltaX = e.clientX - dragStart.x
    const deltaY = e.clientY - dragStart.y

    let newWidth = width
    let newHeight = height
    let newX = x
    let newY = y

    if (resizeHandle.includes("e")) newWidth = Math.max(40, width + deltaX)
    if (resizeHandle.includes("s")) newHeight = Math.max(40, height + deltaY)
    if (resizeHandle.includes("w")) {
      newWidth = Math.max(40, width - deltaX)
      newX = x + deltaX
    }
    if (resizeHandle.includes("n")) {
      newHeight = Math.max(40, height - deltaY)
      newY = y + deltaY
    }

    onUpdate(id, { x: newX, y: newY, width: newWidth, height: newHeight })
    setDragStart({ x: e.clientX, y: e.clientY })
  }

  const handleResizeEnd = () => {
    setResizeHandle(null)
  }

  return (
    <div
      ref={elementRef}
      className={`absolute transition-all ${isSelected ? "ring-2 ring-blue-500" : ""}`}
      style={{
        left: `${x}px`,
        top: `${y}px`,
        width: `${width}px`,
        height: `${height}px`,
        transform: `rotate(${rotation}deg)`,
      }}
      onMouseDown={handleMouseDown}
      onMouseMove={handleMouseMove}
      onMouseUp={handleMouseUp}
      onMouseLeave={handleMouseUp}
    >
      {/* Element Content */}
      <div className="w-full h-full bg-transparent border-2 border-black rounded-sm flex items-center justify-center cursor-move hover:bg-black/5 transition-colors">
        <span className="text-xs text-gray-600 pointer-events-none">{type}</span>
      </div>

      {/* Resize Handles */}
      {isSelected && (
        <>
          {["nw", "n", "ne", "e", "se", "s", "sw", "w"].map((handle) => (
            <div
              key={handle}
              data-resize={handle}
              className="absolute w-2 h-2 bg-blue-500 border border-white rounded-full cursor-pointer hover:w-3 hover:h-3 transition-all"
              style={{
                left: handle.includes("w") ? "-4px" : handle.includes("e") ? "calc(100% - 4px)" : "calc(50% - 4px)",
                top: handle.includes("n") ? "-4px" : handle.includes("s") ? "calc(100% - 4px)" : "calc(50% - 4px)",
              }}
              onMouseDown={(e) => handleResizeStart(e, handle)}
              onMouseMove={handleResizeMove}
              onMouseUp={handleResizeEnd}
            />
          ))}
        </>
      )}

      {/* Element Controls */}
      {isSelected && (
        <div className="absolute -top-8 left-0 flex gap-1 bg-card border border-border rounded-md p-1">
          <button
            onClick={() => onDuplicate(id)}
            className="p-1 hover:bg-muted rounded transition-colors"
            title="Duplicate"
          >
            <Copy size={14} />
          </button>
          <button
            onClick={() => onDelete(id)}
            className="p-1 hover:bg-red-100 text-red-600 rounded transition-colors"
            title="Delete"
          >
            <Trash2 size={14} />
          </button>
        </div>
      )}
    </div>
  )
}
