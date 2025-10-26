"use client"

import Link from "next/link"
import { Button } from "@/components/ui/button"
import { Card } from "@/components/ui/card"
import { ArrowLeft, Play, Pause, RotateCcw, Download, Share2 } from "lucide-react"
import { useState } from "react"
import { IconButton } from "@/components/ui/icon-button"

export function PreviewContent() {
  const [isPlaying, setIsPlaying] = useState(false)

  return (
    <div className="min-h-screen bg-background">
      <header className="border-b border-border bg-muted/50 backdrop-blur-sm sticky top-0 z-50">
        <div className="max-w-7xl mx-auto px-6 py-4 flex items-center justify-between">
          <Link href="/editor">
            <Button variant="outline" size="sm" className="gap-2 bg-transparent">
              <ArrowLeft size={16} />
              Back to Editor
            </Button>
          </Link>
          <h1 className="text-xl font-semibold">Preview Mode</h1>
          <div className="flex items-center gap-2">
            <Button variant="outline" size="sm" className="gap-2 bg-transparent">
              <Download size={16} />
              Export
            </Button>
            <Button variant="outline" size="sm" className="gap-2 bg-transparent">
              <Share2 size={16} />
              Share
            </Button>
          </div>
        </div>
      </header>

      <main className="max-w-7xl mx-auto px-6 py-12">
        {/* Preview Controls */}
        <div className="mb-8 flex items-center justify-between">
          <div className="flex items-center gap-2">
            <IconButton size="sm" variant={isPlaying ? "default" : "ghost"} onClick={() => setIsPlaying(!isPlaying)}>
              {isPlaying ? <Pause size={18} /> : <Play size={18} />}
            </IconButton>
            <IconButton size="sm">
              <RotateCcw size={18} />
            </IconButton>
          </div>
          <div className="text-sm text-muted-foreground">
            <span>Fullscreen Preview</span>
          </div>
        </div>

        {/* Preview Canvas */}
        <Card className="bg-muted border-2 border-border rounded-lg p-12 text-center min-h-96 flex items-center justify-center">
          <div className="text-muted-foreground">
            <p className="text-lg font-medium mb-2">Document Preview</p>
            <p className="text-sm">Your document will be displayed here with all animations and interactions</p>
            <div className="mt-8 space-y-4">
              <div className="inline-block px-6 py-3 bg-primary text-primary-foreground rounded-lg font-medium">
                Interactive Element
              </div>
              <p className="text-xs text-muted-foreground">Click elements to interact with them</p>
            </div>
          </div>
        </Card>

        {/* Preview Info */}
        <div className="mt-8 grid grid-cols-1 md:grid-cols-3 gap-4">
          <Card className="p-4">
            <p className="text-xs font-medium text-muted-foreground mb-2">Document</p>
            <p className="font-semibold">Untitled Document</p>
          </Card>
          <Card className="p-4">
            <p className="text-xs font-medium text-muted-foreground mb-2">Pages</p>
            <p className="font-semibold">1 page</p>
          </Card>
          <Card className="p-4">
            <p className="text-xs font-medium text-muted-foreground mb-2">Animations</p>
            <p className="font-semibold">3 animations</p>
          </Card>
        </div>
      </main>
    </div>
  )
}
