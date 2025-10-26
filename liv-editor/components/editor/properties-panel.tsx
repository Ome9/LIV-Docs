"use client"

import { CollapsibleCard } from "@/components/ui/collapsible-card"
import { ColorPicker } from "@/components/ui/color-picker"
import { SliderInput } from "@/components/ui/slider-input"
import { Tabs, TabsContent, TabsList, TabsTrigger } from "@/components/ui/tabs"
import { Palette, Zap, FileText, Layout } from "lucide-react"
import { useState } from "react"
import { Button } from "@/components/ui/button"
import { PdfToolsPanel } from "@/components/pdf-tools-panel"

interface PropertiesPanelProps {
  selectedElement: string | null
  panelWidth?: number
  autoAdjust?: boolean
  onAutoAdjustChange?: (value: boolean) => void
}

export function PropertiesPanel({ selectedElement, panelWidth = 320 }: PropertiesPanelProps) {
  const [properties, setProperties] = useState({
    fillColor: "#3b82f6",
    strokeColor: "#1f2937",
    opacity: 100,
    strokeWidth: 1,
    borderRadius: 4,
    animationType: "none",
    animationDuration: 0.5,
    animationDelay: 0,
  })

  const handlePreviewAnimation = () => {
    console.log("[v0] Previewing animation:", properties.animationType)
  }

  return (
    <div className="flex-1 border-l border-border bg-card overflow-hidden transition-all duration-300 flex flex-col min-w-0">
      <div className="sticky top-0 bg-card border-b border-border p-3 flex items-center justify-between gap-2 shrink-0">
        <div className="flex items-center gap-2 text-xs">
          <span className="text-muted-foreground">{Math.round(panelWidth)}px</span>
        </div>
      </div>
      
      <Tabs defaultValue="properties" className="flex-1 flex flex-col overflow-hidden">
        <TabsList className="grid w-full grid-cols-2 shrink-0">
          <TabsTrigger value="properties">Properties</TabsTrigger>
          <TabsTrigger value="pdf-tools">Tools</TabsTrigger>
        </TabsList>
        
        <TabsContent value="properties" className="flex-1 overflow-y-auto p-4 space-y-4 min-w-0 mt-0">
          {!selectedElement || selectedElement === "canvas" ? (
            <>
              {/* Document Settings when no element is selected */}
              <CollapsibleCard title="Document Settings" icon={<FileText size={18} />} defaultOpen>
                <div className="space-y-4 min-w-0">
                  <div>
                    <label className="text-sm font-medium">Page Size</label>
                    <select 
                      title="Page Size"
                      className="w-full mt-2 px-3 py-2 bg-background border border-border rounded-md text-sm outline-none focus:border-accent focus:ring-1 focus:ring-accent/50 transition-all"
                    >
                      <option value="A4">A4 (210 × 297 mm)</option>
                      <option value="Letter">Letter (8.5 × 11 in)</option>
                      <option value="A3">A3 (297 × 420 mm)</option>
                      <option value="A5">A5 (148 × 210 mm)</option>
                    </select>
                  </div>

                  <div className="grid grid-cols-2 gap-2">
                    <div>
                      <label className="text-xs font-medium text-muted-foreground">Top Margin</label>
                      <input
                        type="number"
                        placeholder="72"
                        min="0"
                        max="144"
                        className="w-full mt-1 px-2 py-1 bg-background border border-border rounded text-sm outline-none focus:border-accent focus:ring-1 focus:ring-accent/50 transition-all"
                      />
                    </div>
                    <div>
                      <label className="text-xs font-medium text-muted-foreground">Bottom Margin</label>
                      <input
                        type="number"
                        placeholder="72"
                        min="0"
                        max="144"
                        className="w-full mt-1 px-2 py-1 bg-background border border-border rounded text-sm outline-none focus:border-accent focus:ring-1 focus:ring-accent/50 transition-all"
                      />
                    </div>
                  </div>

                  <div className="grid grid-cols-2 gap-2">
                    <div>
                      <label className="text-xs font-medium text-muted-foreground">Left Margin</label>
                      <input
                        type="number"
                        placeholder="72"
                        min="0"
                        max="144"
                        className="w-full mt-1 px-2 py-1 bg-background border border-border rounded text-sm outline-none focus:border-accent focus:ring-1 focus:ring-accent/50 transition-all"
                      />
                    </div>
                    <div>
                      <label className="text-xs font-medium text-muted-foreground">Right Margin</label>
                      <input
                        type="number"
                        placeholder="72"
                        min="0"
                        max="144"
                        className="w-full mt-1 px-2 py-1 bg-background border border-border rounded text-sm outline-none focus:border-accent focus:ring-1 focus:ring-accent/50 transition-all"
                      />
                    </div>
                  </div>

                  <SliderInput
                    label="Line Height"
                    min="1"
                    max="3"
                    step="0.1"
                    defaultValue="1.5"
                  />

                  <div>
                    <label className="text-sm font-medium">Default Font</label>
                    <select 
                      title="Default Font"
                      className="w-full mt-2 px-3 py-2 bg-background border border-border rounded-md text-sm outline-none focus:border-accent focus:ring-1 focus:ring-accent/50 transition-all"
                    >
                      <option value="Times New Roman, serif">Times New Roman</option>
                      <option value="Arial, sans-serif">Arial</option>
                      <option value="Helvetica, sans-serif">Helvetica</option>
                      <option value="Calibri, sans-serif">Calibri</option>
                      <option value="Georgia, serif">Georgia</option>
                    </select>
                  </div>

                  <SliderInput
                    label="Default Font Size"
                    min="8"
                    max="72"
                    defaultValue="12"
                  />
                </div>
              </CollapsibleCard>

              <div className="text-center text-muted-foreground text-sm mt-8">
                <p>Click on the page to start typing</p>
                <p className="text-xs mt-1">Use Ctrl+B, Ctrl+I, Ctrl+U for formatting</p>
              </div>
            </>
          ) : (
            <>
              {/* Styling Section */}
              <CollapsibleCard title="Styling" icon={<Palette size={18} />} defaultOpen>
                <div className="space-y-4 min-w-0">
                  <ColorPicker
                    label="Fill Color"
                    value={properties.fillColor}
                    onChange={(e) => setProperties({ ...properties, fillColor: e.target.value })}
                  />

                  <ColorPicker
                    label="Stroke Color"
                    value={properties.strokeColor}
                    onChange={(e) => setProperties({ ...properties, strokeColor: e.target.value })}
                  />

                  <SliderInput
                    label="Opacity"
                    min="0"
                    max="100"
                    value={properties.opacity}
                    onChange={(e) => setProperties({ ...properties, opacity: Number.parseInt(e.target.value) })}
                  />

                  <SliderInput
                    label="Stroke Width"
                    min="0"
                    max="10"
                    value={properties.strokeWidth}
                    onChange={(e) => setProperties({ ...properties, strokeWidth: Number.parseInt(e.target.value) })}
                  />

                  <SliderInput
                    label="Border Radius"
                    min="0"
                    max="50"
                    value={properties.borderRadius}
                    onChange={(e) => setProperties({ ...properties, borderRadius: Number.parseInt(e.target.value) })}
                  />
                </div>
              </CollapsibleCard>

              {/* Animation Section */}
              <CollapsibleCard title="Animation" icon={<Zap size={18} />}>
                <div className="space-y-4 min-w-0">
                  <div>
                    <label className="text-sm font-medium">Animation Type</label>
                    <select
                      title="Animation Type"
                      value={properties.animationType}
                      onChange={(e) => setProperties({ ...properties, animationType: e.target.value })}
                      className="w-full mt-2 px-3 py-2 bg-background border border-border rounded-md text-sm outline-none focus:border-accent focus:ring-1 focus:ring-accent/50 transition-all"
                    >
                      <option value="none">None</option>
                      <option value="fadeIn">Fade In</option>
                      <option value="slideIn">Slide In</option>
                      <option value="scaleIn">Scale In</option>
                      <option value="rotateIn">Rotate In</option>
                      <option value="bounce">Bounce</option>
                    </select>
                  </div>

                  <SliderInput
                    label="Duration (s)"
                    min="0.1"
                    max="5"
                    step="0.1"
                    value={properties.animationDuration}
                    onChange={(e) => setProperties({ ...properties, animationDuration: Number.parseFloat(e.target.value) })}
                  />

                  <SliderInput
                    label="Delay (s)"
                    min="0"
                    max="2"
                    step="0.1"
                    value={properties.animationDelay}
                    onChange={(e) => setProperties({ ...properties, animationDelay: Number.parseFloat(e.target.value) })}
                  />

                  <Button
                    size="sm"
                    className="w-full gap-2 bg-primary/80 hover:bg-primary text-primary-foreground transition-colors"
                    onClick={handlePreviewAnimation}
                  >
                    <Zap size={16} />
                    Preview Animation
                  </Button>
                </div>
              </CollapsibleCard>

              {/* Typography Section */}
              <CollapsibleCard title="Typography" icon={<FileText size={18} />}>
                <div className="space-y-4 min-w-0">
                  <div>
                    <label className="text-sm font-medium">Font Family</label>
                    <select 
                      title="Font Family"
                      className="w-full mt-2 px-3 py-2 bg-background border border-border rounded-md text-sm outline-none focus:border-accent focus:ring-1 focus:ring-accent/50 transition-all"
                    >
                      <option>Inter</option>
                      <option>Geist</option>
                      <option>Mono</option>
                      <option>Serif</option>
                    </select>
                  </div>

                  <SliderInput label="Font Size" min="8" max="72" defaultValue="16" />

                  <SliderInput label="Line Height" min="1" max="3" step="0.1" defaultValue="1.5" />

                  <div>
                    <label className="text-sm font-medium">Font Weight</label>
                    <select 
                      title="Font Weight"
                      className="w-full mt-2 px-3 py-2 bg-background border border-border rounded-md text-sm outline-none focus:border-accent focus:ring-1 focus:ring-accent/50 transition-all"
                    >
                      <option>Light (300)</option>
                      <option>Regular (400)</option>
                      <option>Medium (500)</option>
                      <option>Semibold (600)</option>
                      <option>Bold (700)</option>
                    </select>
                  </div>
                </div>
              </CollapsibleCard>

              {/* Layout Section */}
              <CollapsibleCard title="Layout" icon={<Layout size={18} />}>
                <div className="space-y-4 min-w-0">
                  <div className="grid grid-cols-2 gap-2">
                    <div>
                      <label className="text-xs font-medium text-muted-foreground">X</label>
                      <input
                        type="number"
                        placeholder="0"
                        className="w-full mt-1 px-2 py-1 bg-background border border-border rounded text-sm outline-none focus:border-accent focus:ring-1 focus:ring-accent/50 transition-all"
                      />
                    </div>
                    <div>
                      <label className="text-xs font-medium text-muted-foreground">Y</label>
                      <input
                        type="number"
                        placeholder="0"
                        className="w-full mt-1 px-2 py-1 bg-background border border-border rounded text-sm outline-none focus:border-accent focus:ring-1 focus:ring-accent/50 transition-all"
                      />
                    </div>
                  </div>

                  <div className="grid grid-cols-2 gap-2">
                    <div>
                      <label className="text-xs font-medium text-muted-foreground">Width</label>
                      <input
                        type="number"
                        placeholder="100"
                        className="w-full mt-1 px-2 py-1 bg-background border border-border rounded text-sm outline-none focus:border-accent focus:ring-1 focus:ring-accent/50 transition-all"
                      />
                    </div>
                    <div>
                      <label className="text-xs font-medium text-muted-foreground">Height</label>
                      <input
                        type="number"
                        placeholder="100"
                        className="w-full mt-1 px-2 py-1 bg-background border border-border rounded text-sm outline-none focus:border-accent focus:ring-1 focus:ring-accent/50 transition-all"
                      />
                    </div>
                  </div>
                </div>
              </CollapsibleCard>
            </>
          )}
        </TabsContent>
        
        <TabsContent value="pdf-tools" className="flex-1 overflow-y-auto mt-0">
          <PdfToolsPanel />
        </TabsContent>
      </Tabs>
    </div>
  )
}
