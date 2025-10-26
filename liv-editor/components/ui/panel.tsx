import * as React from "react"
import { cn } from "@/lib/utils"

interface PanelProps extends React.HTMLAttributes<HTMLDivElement> {
  variant?: "default" | "elevated"
}

const Panel = React.forwardRef<HTMLDivElement, PanelProps>(({ className, variant = "default", ...props }, ref) => {
  const variantClasses = {
    default: "bg-muted border border-border",
    elevated: "bg-muted shadow-lg",
  }

  return <div ref={ref} className={cn("rounded-lg p-4", variantClasses[variant], className)} {...props} />
})
Panel.displayName = "Panel"

export { Panel }
