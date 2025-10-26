"use client"

import * as React from "react"
import { X } from "lucide-react"
import { cn } from "@/lib/utils"

interface ModalProps extends React.HTMLAttributes<HTMLDivElement> {
  isOpen: boolean
  onClose: () => void
  title?: string
  children: React.ReactNode
}

const Modal = React.forwardRef<HTMLDivElement, ModalProps>(
  ({ isOpen, onClose, title, children, className, ...props }, ref) => {
    if (!isOpen) return null

    return (
      <div className="fixed inset-0 z-50 flex items-center justify-center">
        <div className="absolute inset-0 bg-black/50 backdrop-blur-sm" onClick={onClose} />
        <div
          ref={ref}
          className={cn(
            "relative bg-background border border-border rounded-lg shadow-lg max-w-md w-full mx-4",
            className,
          )}
          {...props}
        >
          {title && (
            <div className="flex items-center justify-between p-6 border-b border-border">
              <h2 className="text-lg font-semibold">{title}</h2>
              <button onClick={onClose} className="p-1 hover:bg-muted rounded transition-colors">
                <X size={20} />
              </button>
            </div>
          )}
          <div className="p-6">{children}</div>
        </div>
      </div>
    )
  },
)
Modal.displayName = "Modal"

export { Modal }
