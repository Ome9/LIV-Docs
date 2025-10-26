"use client"

import { IconButton } from "@/components/ui/icon-button"
import { Play, Pause, RotateCcw, Plus } from "lucide-react"
import { useState } from "react"

interface Keyframe {
  id: string
  time: number
  property: string
}

export function TimelinePanel() {
  const [isPlaying, setIsPlaying] = useState(false)
  const [currentTime, setCurrentTime] = useState(0)
  const [keyframes, setKeyframes] = useState<Keyframe[]>([
    { id: "1", time: 0, property: "opacity" },
    { id: "2", time: 1, property: "transform" },
  ])

  const duration = 5

  return (
    <div className="h-40 border-t border-border bg-muted/50 px-4 py-3 flex flex-col">
      <div className="flex items-center justify-between mb-3">
        <h3 className="font-semibold text-sm">Timeline</h3>
        <div className="flex items-center gap-2">
          <IconButton size="sm" variant={isPlaying ? "default" : "ghost"} onClick={() => setIsPlaying(!isPlaying)}>
            {isPlaying ? <Pause size={16} /> : <Play size={16} />}
          </IconButton>
          <IconButton size="sm" onClick={() => setCurrentTime(0)}>
            <RotateCcw size={16} />
          </IconButton>
          <IconButton size="sm">
            <Plus size={16} />
          </IconButton>
        </div>
      </div>

      {/* Timeline Ruler */}
      <div className="flex-1 bg-background rounded border border-border overflow-hidden flex flex-col">
        {/* Time Ruler */}
        <div className="h-6 border-b border-border bg-muted/50 flex items-center px-2 text-xs text-muted-foreground">
          {[...Array(Math.ceil(duration))].map((_, i) => (
            <div key={i} className="flex-1 text-center">
              {i}s
            </div>
          ))}
        </div>

        {/* Keyframes Track */}
        <div className="flex-1 relative overflow-x-auto">
          {/* Progress Line */}
          <div
            className="absolute top-0 bottom-0 w-0.5 bg-accent pointer-events-none"
            style={{ left: `${(currentTime / duration) * 100}%` }}
          />

          {/* Keyframe Items */}
          <div className="flex h-full">
            {[...Array(Math.ceil(duration * 10))].map((_, i) => {
              const time = i * 0.1
              const hasKeyframe = keyframes.some((k) => Math.abs(k.time - time) < 0.05)

              return (
                <div
                  key={i}
                  className="flex-1 border-r border-border/50 hover:bg-muted/50 transition-colors cursor-pointer relative group"
                  onClick={() => setCurrentTime(time)}
                >
                  {hasKeyframe && (
                    <div className="absolute left-1/2 top-1/2 -translate-x-1/2 -translate-y-1/2 w-3 h-3 bg-accent rounded-full" />
                  )}
                </div>
              )
            })}
          </div>
        </div>
      </div>

      {/* Time Display */}
      <div className="mt-2 flex items-center justify-between text-xs text-muted-foreground">
        <span>{currentTime.toFixed(2)}s</span>
        <span>{duration}s</span>
      </div>
    </div>
  )
}
