"use client"

import { useState } from "react"
import Link from "next/link"
import { IconButton } from "@/components/ui/icon-button"
import { Button } from "@/components/ui/button"
import { Toolbar } from "@/components/ui/toolbar"
import {
  Menu,
  Save,
  Download,
  Undo2,
  Redo2,
  Eye,
  Settings,
  Share2,
  Check,
  FileText,
  Plus,
  MoreVertical,
} from "lucide-react"

export function EditorNavbar() {
  const [documentName, setDocumentName] = useState("Untitled PDF")
  const [isSaved, setIsSaved] = useState(true)
  const [isSaving, setIsSaving] = useState(false)
  const [pageCount, setPageCount] = useState(1)

  const handleSave = async () => {
    setIsSaving(true)
    await new Promise((resolve) => setTimeout(resolve, 800))
    setIsSaved(true)
    setIsSaving(false)
  }

  const handlePreview = () => {
    window.location.href = "/preview"
  }

  const handleShare = async () => {
    await navigator.clipboard.writeText(window.location.href)
    alert("Link copied to clipboard!")
  }

  const handleExport = () => {
    const element = document.createElement("a")
    element.setAttribute("href", "data:text/plain;charset=utf-8,")
    element.setAttribute("download", `${documentName}.pdf`)
    element.style.display = "none"
    document.body.appendChild(element)
    element.click()
    document.body.removeChild(element)
  }

  const handleUndo = () => {
    console.log("[v0] Undo triggered")
  }

  const handleRedo = () => {
    console.log("[v0] Redo triggered")
  }

  const handleAddPage = () => {
    setPageCount(pageCount + 1)
    console.log("[v0] Added new page")
  }

  const handlePrint = () => {
    window.print()
  }

  return (
    <Toolbar className="justify-between bg-card border-b border-border/50">
      <div className="flex items-center gap-4">
        <IconButton size="sm" variant="ghost" className="hover:bg-muted transition-colors" title="Menu">
          <Menu size={20} />
        </IconButton>
        <div className="w-8 h-8 rounded-lg bg-primary/20 flex items-center justify-center border border-primary/30">
          <span className="text-primary font-bold text-sm">L</span>
        </div>
        <input
          type="text"
          value={documentName}
          onChange={(e) => {
            setDocumentName(e.target.value)
            setIsSaved(false)
          }}
          placeholder="Untitled PDF"
          className="bg-transparent text-lg font-semibold outline-none hover:bg-muted/30 px-2 py-1 rounded transition-colors"
        />
        {!isSaved && <span className="text-xs text-muted-foreground animate-pulse">Unsaved</span>}

        <span className="text-xs text-muted-foreground border-l border-border/50 pl-4">
          <FileText size={14} className="inline mr-1" />
          {pageCount} page{pageCount !== 1 ? "s" : ""}
        </span>
      </div>

      <div className="flex items-center gap-2">
        <div className="flex items-center gap-1 border-r border-border/50 pr-4">
          <IconButton size="sm" title="Undo" onClick={handleUndo} className="hover:bg-muted transition-colors">
            <Undo2 size={18} />
          </IconButton>
          <IconButton size="sm" title="Redo" onClick={handleRedo} className="hover:bg-muted transition-colors">
            <Redo2 size={18} />
          </IconButton>
        </div>

        <div className="flex items-center gap-2">
          <Button
            variant="outline"
            size="sm"
            className="gap-2 bg-transparent hover:bg-muted transition-colors"
            onClick={handleAddPage}
            title="Add new page"
          >
            <Plus size={16} />
            Page
          </Button>
          <Button
            variant="outline"
            size="sm"
            className="gap-2 bg-transparent hover:bg-muted transition-colors"
            onClick={handlePreview}
          >
            <Eye size={16} />
            Preview
          </Button>
          <Button
            variant="outline"
            size="sm"
            className="gap-2 bg-transparent hover:bg-muted transition-colors"
            onClick={handleShare}
          >
            <Share2 size={16} />
            Share
          </Button>
          <Button
            size="sm"
            className="gap-2 bg-primary/80 hover:bg-primary text-primary-foreground transition-colors"
            onClick={handleSave}
            disabled={isSaving}
          >
            {isSaving ? (
              <>
                <div className="w-4 h-4 border-2 border-primary-foreground border-t-transparent rounded-full animate-spin" />
                Saving...
              </>
            ) : isSaved ? (
              <>
                <Check size={16} />
                Saved
              </>
            ) : (
              <>
                <Save size={16} />
                Save
              </>
            )}
          </Button>
          <Button
            variant="outline"
            size="sm"
            className="gap-2 bg-transparent hover:bg-muted transition-colors"
            onClick={handleExport}
          >
            <Download size={16} />
            Export
          </Button>
          <Button
            variant="outline"
            size="sm"
            className="gap-2 bg-transparent hover:bg-muted transition-colors"
            onClick={handlePrint}
            title="Print"
          >
            Print
          </Button>
        </div>

        <div className="flex items-center gap-1 border-l border-border/50 pl-4">
          <Link href="/settings">
            <IconButton size="sm" title="Settings" className="hover:bg-muted transition-colors">
              <Settings size={18} />
            </IconButton>
          </Link>
          <IconButton size="sm" title="More options" className="hover:bg-muted transition-colors">
            <MoreVertical size={18} />
          </IconButton>
        </div>
      </div>
    </Toolbar>
  )
}
