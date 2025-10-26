"use client"

import { useState } from "react"
import { Button } from "@/components/ui/button"
import { IconButton } from "@/components/ui/icon-button"
import { Copy, Check, Play } from "lucide-react"

interface CodeEditorProps {
  onRun?: (code: string) => void
}

export function CodeEditor({ onRun }: CodeEditorProps) {
  const [code, setCode] = useState(`// Add your React component code here
import React from 'react'

export default function MyComponent() {
  return (
    <div className="p-4 bg-blue-50 rounded-lg">
      <h2 className="text-lg font-semibold">Hello World</h2>
      <p>Your component here</p>
    </div>
  )
}`)
  const [copied, setCopied] = useState(false)

  const handleCopy = () => {
    navigator.clipboard.writeText(code)
    setCopied(true)
    setTimeout(() => setCopied(false), 2000)
  }

  const handleRun = () => {
    onRun?.(code)
  }

  return (
    <div className="flex flex-col h-full bg-card border-l border-border">
      <div className="flex items-center justify-between p-3 border-b border-border/50 bg-muted/30">
        <h3 className="text-sm font-semibold">Code Editor</h3>
        <div className="flex items-center gap-2">
          <IconButton size="sm" onClick={handleCopy} title="Copy code" className="hover:bg-accent/20">
            {copied ? <Check size={16} className="text-green-600" /> : <Copy size={16} />}
          </IconButton>
          <Button
            size="sm"
            onClick={handleRun}
            className="gap-2 bg-primary/80 hover:bg-primary text-primary-foreground"
          >
            <Play size={14} />
            Run
          </Button>
        </div>
      </div>
      <textarea
        value={code}
        onChange={(e) => setCode(e.target.value)}
        className="flex-1 p-4 bg-background text-foreground font-mono text-sm resize-none outline-none border-none focus:ring-0"
        placeholder="Write your code here..."
        spellCheck="false"
      />
    </div>
  )
}
