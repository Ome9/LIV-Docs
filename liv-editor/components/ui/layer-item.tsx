"use client"

import * as React from "react"
import { Eye, EyeOff, Lock, Unlock } from "lucide-react"
import { cn } from "@/lib/utils"

interface LayerItemProps extends React.HTMLAttributes<HTMLDivElement> {
  name: string
  isVisible?: boolean
  isLocked?: boolean
  isSelected?: boolean
  onVisibilityToggle?: () => void
  onLockToggle?: () => void
}

const LayerItem = React.forwardRef<HTMLDivElement, LayerItemProps>(
  (
    {
      name,
      isVisible = true,
      isLocked = false,
      isSelected = false,
      onVisibilityToggle,
      onLockToggle,
      className,
      ...props
    },
    ref,
  ) => {
    return (
      <div
        ref={ref}
        className={cn(
          "flex items-center justify-between p-2 rounded-md cursor-pointer transition-colors",
          isSelected ? "bg-accent text-accent-foreground" : "hover:bg-muted/50",
        )}
        {...props}
      >
        <span className="text-sm font-medium flex-1 truncate">{name}</span>
        <div className="flex items-center gap-1">
          <button
            onClick={(e) => {
              e.stopPropagation()
              onVisibilityToggle?.()
            }}
            className="p-1 hover:bg-muted/50 rounded transition-colors"
          >
            {isVisible ? <Eye size={16} /> : <EyeOff size={16} />}
          </button>
          <button
            onClick={(e) => {
              e.stopPropagation()
              onLockToggle?.()
            }}
            className="p-1 hover:bg-muted/50 rounded transition-colors"
          >
            {isLocked ? <Lock size={16} /> : <Unlock size={16} />}
          </button>
        </div>
      </div>
    )
  },
)
LayerItem.displayName = "LayerItem"

export { LayerItem }
