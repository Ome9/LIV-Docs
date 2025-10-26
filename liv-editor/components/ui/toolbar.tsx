"use client"

import * as React from "react"
import { cn } from "@/lib/utils"

interface ToolbarProps extends React.HTMLAttributes<HTMLDivElement> {
  variant?: "default" | "compact"
}

const Toolbar = React.forwardRef<HTMLDivElement, ToolbarProps>(({ className, variant = "default", ...props }, ref) => {
  const variantClasses = {
    default: "px-4 py-3 gap-2",
    compact: "px-2 py-2 gap-1",
  }

  return (
    <div
      ref={ref}
      className={cn(
        "flex items-center border-b border-border bg-muted/50 backdrop-blur-sm",
        variantClasses[variant],
        className,
      )}
      {...props}
    />
  )
})
Toolbar.displayName = "Toolbar"

export { Toolbar }
