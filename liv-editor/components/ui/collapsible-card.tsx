"use client"

import * as React from "react"
import { ChevronDown } from "lucide-react"
import { cn } from "@/lib/utils"

interface CollapsibleCardProps extends React.HTMLAttributes<HTMLDivElement> {
  title: string
  defaultOpen?: boolean
  icon?: React.ReactNode
}

const CollapsibleCard = React.forwardRef<HTMLDivElement, CollapsibleCardProps>(
  ({ title, defaultOpen = true, icon, children, className, ...props }, ref) => {
    const [isOpen, setIsOpen] = React.useState(defaultOpen)

    return (
      <div ref={ref} className={cn("border border-border rounded-lg overflow-hidden", className)} {...props}>
        <button
          onClick={() => setIsOpen(!isOpen)}
          className="w-full flex items-center justify-between p-4 hover:bg-muted/50 transition-colors"
        >
          <div className="flex items-center gap-2">
            {icon}
            <span className="font-semibold">{title}</span>
          </div>
          <ChevronDown size={18} className={`transition-transform duration-200 ${isOpen ? "" : "-rotate-90"}`} />
        </button>
        {isOpen && <div className="px-4 pb-4 border-t border-border">{children}</div>}
      </div>
    )
  },
)
CollapsibleCard.displayName = "CollapsibleCard"

export { CollapsibleCard }
