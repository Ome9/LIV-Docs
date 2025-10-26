import * as React from "react"
import { cn } from "@/lib/utils"

interface IconButtonProps extends React.ButtonHTMLAttributes<HTMLButtonElement> {
  variant?: "default" | "ghost" | "outline"
  size?: "sm" | "md" | "lg"
}

const IconButton = React.forwardRef<HTMLButtonElement, IconButtonProps>(
  ({ className, variant = "ghost", size = "md", ...props }, ref) => {
    const sizeClasses = {
      sm: "h-8 w-8",
      md: "h-10 w-10",
      lg: "h-12 w-12",
    }

    const variantClasses = {
      default:
        "bg-gradient-to-r from-primary to-accent text-primary-foreground hover:shadow-lg hover:shadow-primary/50 hover:scale-105",
      ghost: "hover:bg-accent/20 text-foreground hover:text-accent",
      outline: "border border-border hover:bg-accent/20 hover:border-accent hover:text-accent",
    }

    return (
      <button
        ref={ref}
        className={cn(
          "inline-flex items-center justify-center rounded-md transition-all duration-200",
          sizeClasses[size],
          variantClasses[variant],
          className,
        )}
        {...props}
      />
    )
  },
)
IconButton.displayName = "IconButton"

export { IconButton }
