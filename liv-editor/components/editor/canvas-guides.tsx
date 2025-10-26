"use client"

interface CanvasGuidesProps {
  elements: Array<{ id: string; x: number; y: number; width?: number; height?: number }>
  selectedElementId?: string | null
  canvasWidth: number
  canvasHeight: number
}

export function CanvasGuides({ elements, selectedElementId, canvasWidth, canvasHeight }: CanvasGuidesProps) {
  const guides: Array<{ type: "vertical" | "horizontal"; position: number }> = []

  if (selectedElementId) {
    const selected = elements.find((el) => el.id === selectedElementId)
    if (selected) {
      // Center guides
      const centerX = selected.x + (selected.width || 120) / 2
      const centerY = selected.y + (selected.height || 60) / 2

      // Check alignment with other elements
      elements.forEach((el) => {
        if (el.id === selectedElementId) return
        const elCenterX = el.x + (el.width || 120) / 2
        const elCenterY = el.y + (el.height || 60) / 2

        if (Math.abs(centerX - elCenterX) < 5) {
          guides.push({ type: "vertical", position: centerX })
        }
        if (Math.abs(centerY - elCenterY) < 5) {
          guides.push({ type: "horizontal", position: centerY })
        }
      })
    }
  }

  return (
    <>
      {guides.map((guide, idx) =>
        guide.type === "vertical" ? (
          <div
            key={idx}
            className="absolute top-0 bottom-0 w-px bg-pink-500 opacity-50 pointer-events-none"
            style={{ left: `${guide.position}px` }}
          />
        ) : (
          <div
            key={idx}
            className="absolute left-0 right-0 h-px bg-pink-500 opacity-50 pointer-events-none"
            style={{ top: `${guide.position}px` }}
          />
        ),
      )}
    </>
  )
}
