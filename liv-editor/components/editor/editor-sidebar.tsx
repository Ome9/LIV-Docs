"use client"

import { useState } from "react"
import { IconButton } from "@/components/ui/icon-button"
import { Panel } from "@/components/ui/panel"
import { LayerItem } from "@/components/ui/layer-item"
import { FileText, ImageIcon, Zap, SettingsIcon, ChevronDown, Plus, Upload } from "lucide-react"
import { Button } from "@/components/ui/button"
import { useButtonActions } from "@/hooks/use-button-actions"

interface EditorSidebarProps {
  activeTab: "pages" | "assets" | "animations" | "settings"
  onTabChange: (tab: "pages" | "assets" | "animations" | "settings") => void
}

const tabs = [
  { id: "pages", label: "Pages", icon: FileText },
  { id: "assets", label: "Assets", icon: ImageIcon },
  { id: "animations", label: "Animations", icon: Zap },
  { id: "settings", label: "Settings", icon: SettingsIcon },
]

const mockLayers = [
  { id: 1, name: "Background", visible: true, locked: false },
  { id: 2, name: "Header", visible: true, locked: false },
  { id: 3, name: "Title Text", visible: true, locked: false },
  { id: 4, name: "Button Group", visible: true, locked: false },
]

export function EditorSidebar({ activeTab, onTabChange }: EditorSidebarProps) {
  const [collapsed, setCollapsed] = useState(false)
  const [selectedLayer, setSelectedLayer] = useState<number | null>(null)
  const [layers, setLayers] = useState(mockLayers)
  const { executeAction } = useButtonActions()

  const handleAddPage = async () => {
    await executeAction(async () => {
      console.log("[v0] Adding new page")
    }, "add-page")
  }

  const handleUploadAsset = async () => {
    await executeAction(async () => {
      console.log("[v0] Opening file upload")
    }, "upload-asset")
  }

  if (collapsed) {
    return (
      <div className="w-16 border-r border-border bg-linear-to-b from-muted/50 to-background flex flex-col items-center py-4 gap-2 transition-all duration-300">
        {tabs.map((tab) => {
          const Icon = tab.icon
          return (
            <IconButton
              key={tab.id}
              size="md"
              variant={activeTab === tab.id ? "default" : "ghost"}
              onClick={() => onTabChange(tab.id as any)}
              title={tab.label}
              className="scale-in"
            >
              <Icon size={20} />
            </IconButton>
          )
        })}
        <IconButton size="md" variant="ghost" onClick={() => setCollapsed(false)} className="mt-auto hover:scale-110">
          <ChevronDown size={20} className="rotate-90" />
        </IconButton>
      </div>
    )
  }

  return (
    <div className="w-64 border-r border-border bg-linear-to-b from-muted/50 to-background flex flex-col transition-all duration-300">
      <div className="flex items-center justify-between p-4 border-b border-border/50">
        <h3 className="font-semibold text-sm">{tabs.find((t) => t.id === activeTab)?.label}</h3>
        <IconButton size="sm" variant="ghost" onClick={() => setCollapsed(true)}>
          <ChevronDown size={18} className="-rotate-90" />
        </IconButton>
      </div>

      <div className="flex-1 overflow-y-auto p-4">
        {activeTab === "pages" && (
          <div className="space-y-2">
            <Panel
              variant="elevated"
              className="cursor-pointer hover:border-accent hover:shadow-lg hover:shadow-accent/30 transition-all duration-300 scale-in"
            >
              <p className="font-medium text-sm">Page 1</p>
              <p className="text-xs text-muted-foreground">Active</p>
            </Panel>
            <Panel className="cursor-pointer hover:border-accent hover:shadow-lg hover:shadow-accent/30 transition-all duration-300 scale-in">
              <p className="font-medium text-sm">Page 2</p>
            </Panel>
            <Button
              variant="outline"
              size="sm"
              className="w-full gap-2 mt-4 bg-transparent hover:bg-accent/20 hover:text-accent hover:border-accent transition-all"
              onClick={handleAddPage}
            >
              <Plus size={16} />
              Add Page
            </Button>
          </div>
        )}

        {activeTab === "assets" && (
          <div className="space-y-2">
            <Panel className="text-center py-8 text-muted-foreground border-2 border-dashed cursor-pointer hover:border-accent hover:bg-accent/10 transition-all duration-300 scale-in">
              <Upload size={24} className="mx-auto mb-2 opacity-50" />
              <p className="text-sm">Drag files here</p>
              <p className="text-xs">or click to browse</p>
            </Panel>
            <div className="space-y-2 mt-4">
              <Panel className="p-3 flex items-center gap-3 cursor-pointer hover:bg-accent/20 hover:border-accent transition-all duration-300 scale-in">
                <div className="w-8 h-8 bg-linear-to-br from-primary to-accent rounded glow-effect" />
                <div className="flex-1 min-w-0">
                  <p className="text-sm font-medium truncate">logo.svg</p>
                  <p className="text-xs text-muted-foreground">24 KB</p>
                </div>
              </Panel>
            </div>
            <Button
              variant="outline"
              size="sm"
              className="w-full gap-2 mt-4 bg-transparent hover:bg-accent/20 hover:text-accent hover:border-accent transition-all"
              onClick={handleUploadAsset}
            >
              <Upload size={16} />
              Upload Asset
            </Button>
          </div>
        )}

        {activeTab === "animations" && (
          <div className="space-y-2">
            <Panel className="text-center py-8 text-muted-foreground scale-in">
              <Zap size={24} className="mx-auto mb-2 opacity-50" />
              <p className="text-sm">No animations</p>
              <p className="text-xs">Select an element to add animations</p>
            </Panel>
          </div>
        )}

        {activeTab === "settings" && (
          <div className="space-y-4 scale-in">
            <div>
              <label className="text-sm font-medium">Document Name</label>
              <input
                type="text"
                placeholder="Untitled"
                className="w-full mt-2 px-3 py-2 bg-background border border-border rounded-md text-sm outline-none focus:border-accent focus:shadow-lg focus:shadow-accent/30 transition-all"
              />
            </div>
            <div>
              <label className="text-sm font-medium">Permissions</label>
              <select 
                className="w-full mt-2 px-3 py-2 bg-background border border-border rounded-md text-sm outline-none focus:border-accent focus:shadow-lg focus:shadow-accent/30 transition-all"
                title="Document Permissions"
              >
                <option>Private</option>
                <option>Shared</option>
                <option>Public</option>
              </select>
            </div>
          </div>
        )}
      </div>

      {/* Layers Panel - Always visible at bottom */}
      {activeTab === "pages" && (
        <div className="border-t border-border/50 p-4 bg-background/50 slide-in-left">
          <h4 className="text-xs font-semibold text-muted-foreground mb-3 uppercase">Layers</h4>
          <div className="space-y-1">
            {layers.map((layer, index) => (
              <div 
                key={layer.id} 
                className={`scale-in animation-delay-${Math.min(index, 4)}`}
              >
                <LayerItem
                  name={layer.name}
                  isVisible={layer.visible}
                  isLocked={layer.locked}
                  isSelected={selectedLayer === layer.id}
                  onClick={() => setSelectedLayer(layer.id)}
                  onVisibilityToggle={() => {
                    setLayers(layers.map((l) => (l.id === layer.id ? { ...l, visible: !l.visible } : l)))
                  }}
                  onLockToggle={() => {
                    setLayers(layers.map((l) => (l.id === layer.id ? { ...l, locked: !l.locked } : l)))
                  }}
                />
              </div>
            ))}
          </div>
        </div>
      )}
      
      <style jsx>{`
        .animation-delay-0 { animation-delay: 0ms; }
        .animation-delay-1 { animation-delay: 50ms; }  
        .animation-delay-2 { animation-delay: 100ms; }
        .animation-delay-3 { animation-delay: 150ms; }
        .animation-delay-4 { animation-delay: 200ms; }
      `}</style>
    </div>
  )
}
