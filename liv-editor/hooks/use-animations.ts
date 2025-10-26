"use client"

import { useRef } from "react"

interface AnimationConfig {
  duration?: number
  delay?: number
  easing?: string
}

export function useAnimations() {
  const elementRef = useRef<HTMLElement>(null)

  const fadeIn = (config: AnimationConfig = {}) => {
    const { duration = 600, delay = 0, easing = "easeOutQuad" } = config
    if (!elementRef.current) return

    const element = elementRef.current
    element.style.opacity = "0"

    setTimeout(() => {
      element.style.transition = `opacity ${duration}ms ${easing}`
      element.style.opacity = "1"
    }, delay)
  }

  const slideIn = (direction: "left" | "right" | "up" | "down" = "left", config: AnimationConfig = {}) => {
    const { duration = 600, delay = 0, easing = "easeOutQuad" } = config
    if (!elementRef.current) return

    const element = elementRef.current
    const distance = 30

    const transforms = {
      left: `translateX(-${distance}px)`,
      right: `translateX(${distance}px)`,
      up: `translateY(-${distance}px)`,
      down: `translateY(${distance}px)`,
    }

    element.style.transform = transforms[direction]
    element.style.opacity = "0"

    setTimeout(() => {
      element.style.transition = `all ${duration}ms ${easing}`
      element.style.transform = "translate(0, 0)"
      element.style.opacity = "1"
    }, delay)
  }

  const scaleIn = (config: AnimationConfig = {}) => {
    const { duration = 600, delay = 0, easing = "easeOutQuad" } = config
    if (!elementRef.current) return

    const element = elementRef.current
    element.style.transform = "scale(0.95)"
    element.style.opacity = "0"

    setTimeout(() => {
      element.style.transition = `all ${duration}ms ${easing}`
      element.style.transform = "scale(1)"
      element.style.opacity = "1"
    }, delay)
  }

  const pulse = (config: AnimationConfig = {}) => {
    const { duration = 2000 } = config
    if (!elementRef.current) return

    const element = elementRef.current
    const keyframes = `
      @keyframes pulse {
        0%, 100% { opacity: 1; }
        50% { opacity: 0.5; }
      }
    `

    const style = document.createElement("style")
    style.textContent = keyframes
    document.head.appendChild(style)

    element.style.animation = `pulse ${duration}ms infinite`
  }

  return { elementRef, fadeIn, slideIn, scaleIn, pulse }
}
