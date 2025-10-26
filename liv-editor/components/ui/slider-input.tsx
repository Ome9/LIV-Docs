"use client"

import * as React from "react"
import { cn } from "@/lib/utils"

interface SliderInputProps extends React.InputHTMLAttributes<HTMLInputElement> {
  label?: string
  showValue?: boolean
}

const SliderInput = React.forwardRef<HTMLInputElement, SliderInputProps>(
  ({ label, showValue = true, className, ...props }, ref) => {
    const [value, setValue] = React.useState(props.defaultValue || props.value || 0)

    return (
      <div className="space-y-2">
        {label && (
          <div className="flex items-center justify-between">
            <label className="text-sm font-medium">{label}</label>
            {showValue && <span className="text-sm text-muted-foreground">{value}</span>}
          </div>
        )}
        <input
          ref={ref}
          type="range"
          className={cn("w-full h-2 bg-muted rounded-lg appearance-none cursor-pointer accent-primary", className)}
          onChange={(e) => setValue(e.target.value)}
          {...props}
        />
      </div>
    )
  },
)
SliderInput.displayName = "SliderInput"

export { SliderInput }
