"use client"

import * as React from "react"
import { cn } from "@/lib/utils"

interface ColorPickerProps extends React.InputHTMLAttributes<HTMLInputElement> {
  label?: string
  showHex?: boolean
}

const ColorPicker = React.forwardRef<HTMLInputElement, ColorPickerProps>(
  ({ label, showHex = true, className, ...props }, ref) => {
    const [hex, setHex] = React.useState(props.defaultValue || "#6366f1")

    return (
      <div className="space-y-2">
        {label && <label className="text-sm font-medium">{label}</label>}
        <div className="flex gap-2 min-w-0">
          <input
            ref={ref}
            type="color"
            className={cn("w-12 h-10 rounded cursor-pointer border border-border flex-shrink-0 min-w-12", className)}
            onChange={(e) => setHex(e.target.value)}
            {...props}
          />
          {showHex && (
            /* Added min-w-20 to prevent hex input from being cut off when resizing */
            <input
              type="text"
              value={hex}
              onChange={(e) => setHex(e.target.value)}
              placeholder="#000000"
              className="flex-1 px-3 py-2 bg-background border border-border rounded text-sm outline-none focus:border-accent transition-colors min-w-20"
            />
          )}
        </div>
      </div>
    )
  },
)
ColorPicker.displayName = "ColorPicker"

export { ColorPicker }
