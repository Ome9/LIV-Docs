"use client"

import * as React from "react"
import { cn } from "@/lib/utils"

interface DropdownMenuProps extends React.HTMLAttributes<HTMLDivElement> {
  trigger: React.ReactNode
  items: Array<{
    label: string
    onClick: () => void
    icon?: React.ReactNode
  }>
}

const DropdownMenu = React.forwardRef<HTMLDivElement, DropdownMenuProps>(
  ({ trigger, items, className, ...props }, ref) => {
    const [isOpen, setIsOpen] = React.useState(false)

    return (
      <div ref={ref} className={cn("relative inline-block", className)} {...props}>
        <button onClick={() => setIsOpen(!isOpen)}>{trigger}</button>

        {isOpen && (
          <div className="absolute top-full right-0 mt-2 bg-background border border-border rounded-lg shadow-lg overflow-hidden z-50">
            {items.map((item, index) => (
              <button
                key={index}
                onClick={() => {
                  item.onClick()
                  setIsOpen(false)
                }}
                className="w-full px-4 py-2 text-left text-sm hover:bg-muted transition-colors flex items-center gap-2"
              >
                {item.icon}
                {item.label}
              </button>
            ))}
          </div>
        )}
      </div>
    )
  },
)
DropdownMenu.displayName = "DropdownMenu"

export { DropdownMenu }
