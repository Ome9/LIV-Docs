"use client"

import type React from "react"

import { useState } from "react"
import { Button } from "@/components/ui/button"
import { IconButton } from "@/components/ui/icon-button"
import { Input } from "@/components/ui/input"
import { Search, Plus, Zap, Box, Type, ImageIcon, ChevronRight } from "lucide-react"

interface Component {
  id: string
  name: string
  icon: React.ReactNode
  category: string
}

const components: Component[] = [
  { id: "button", name: "Button", icon: <Box size={16} />, category: "UI" },
  { id: "input", name: "Input", icon: <Type size={16} />, category: "Form" },
  { id: "card", name: "Card", icon: <Box size={16} />, category: "Layout" },
  { id: "badge", name: "Badge", icon: <Zap size={16} />, category: "UI" },
  { id: "image", name: "Image", icon: <ImageIcon size={16} />, category: "Media" },
  { id: "text", name: "Text", icon: <Type size={16} />, category: "Typography" },
  { id: "divider", name: "Divider", icon: <Box size={16} />, category: "Layout" },
  { id: "grid", name: "Grid", icon: <Box size={16} />, category: "Layout" },
]

interface ComponentLibraryProps {
  onAddComponent?: (componentId: string) => void
  onMinimize?: () => void
}

export function ComponentLibrary({ onAddComponent, onMinimize }: ComponentLibraryProps) {
  const [search, setSearch] = useState("")
  const [selectedCategory, setSelectedCategory] = useState<string | null>(null)

  const categories = Array.from(new Set(components.map((c) => c.category)))

  const filtered = components.filter((c) => {
    const matchesSearch = c.name.toLowerCase().includes(search.toLowerCase())
    const matchesCategory = !selectedCategory || c.category === selectedCategory
    return matchesSearch && matchesCategory
  })

  return (
    <div className="flex flex-col h-full bg-card">
      <div className="p-3 border-b border-border/50 bg-muted/30 space-y-3">
        <div className="flex items-center justify-between">
          <h3 className="text-sm font-semibold">Component Library</h3>
          <button onClick={onMinimize} className="p-1 hover:bg-muted rounded transition-colors" title="Minimize">
            <ChevronRight size={16} />
          </button>
        </div>
        <div className="relative">
          <Search size={16} className="absolute left-2 top-2.5 text-muted-foreground" />
          <Input
            placeholder="Search components..."
            value={search}
            onChange={(e) => setSearch(e.target.value)}
            className="pl-8 h-8 text-sm bg-background border-border/50"
          />
        </div>
      </div>

      <div className="flex gap-1 p-2 border-b border-border/50 overflow-x-auto">
        <Button
          size="sm"
          variant={selectedCategory === null ? "default" : "outline"}
          onClick={() => setSelectedCategory(null)}
          className="text-xs whitespace-nowrap"
        >
          All
        </Button>
        {categories.map((cat) => (
          <Button
            key={cat}
            size="sm"
            variant={selectedCategory === cat ? "default" : "outline"}
            onClick={() => setSelectedCategory(cat)}
            className="text-xs whitespace-nowrap"
          >
            {cat}
          </Button>
        ))}
      </div>

      <div className="flex-1 overflow-y-auto p-2 space-y-2">
        {filtered.map((comp) => (
          <div
            key={comp.id}
            className="panel-section-subtle flex items-center justify-between cursor-move hover:bg-muted/50 group"
            draggable
            onDragStart={(e) => {
              e.dataTransfer.effectAllowed = "copy"
              e.dataTransfer.setData("componentId", comp.id)
            }}
          >
            <div className="flex items-center gap-2 flex-1 min-w-0">
              <div className="text-muted-foreground">{comp.icon}</div>
              <div className="min-w-0">
                <p className="text-sm font-medium truncate">{comp.name}</p>
                <p className="text-xs text-muted-foreground">{comp.category}</p>
              </div>
            </div>
            <IconButton
              size="sm"
              variant="ghost"
              onClick={() => onAddComponent?.(comp.id)}
              className="opacity-0 group-hover:opacity-100 transition-opacity"
            >
              <Plus size={14} />
            </IconButton>
          </div>
        ))}
      </div>
    </div>
  )
}
