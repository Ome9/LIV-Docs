"use client"

interface CanvasRulersProps {
  zoom: number
}

export function CanvasRulers({ zoom: zoomLevel }: CanvasRulersProps) {
  const scale = zoomLevel / 100
  const majorInterval = 50
  const minorInterval = 10

  return (
    <>
      {/* Top Ruler */}
      <div className="absolute top-0 left-0 right-0 h-6 bg-muted border-b border-border flex items-end pointer-events-none">
        <svg width="100%" height="24" className="flex-1">
          {Array.from({ length: 1000 }).map((_, i) => {
            const pos = i * minorInterval * scale
            const isMajor = i % (majorInterval / minorInterval) === 0
            return (
              <g key={i}>
                <line
                  x1={pos}
                  y1={isMajor ? 12 : 18}
                  x2={pos}
                  y2="24"
                  stroke="currentColor"
                  className="text-muted-foreground"
                />
                {isMajor && (
                  <text x={pos + 2} y="12" fontSize="10" className="text-muted-foreground">
                    {i * minorInterval}
                  </text>
                )}
              </g>
            )
          })}
        </svg>
      </div>

      {/* Left Ruler */}
      <div className="absolute top-0 left-0 bottom-0 w-6 bg-muted border-r border-border flex flex-col items-end pointer-events-none">
        <svg width="24" height="100%" className="flex-1">
          {Array.from({ length: 1000 }).map((_, i) => {
            const pos = i * minorInterval * scale
            const isMajor = i % (majorInterval / minorInterval) === 0
            return (
              <g key={i}>
                <line
                  x1={isMajor ? 12 : 18}
                  y1={pos}
                  x2="24"
                  y2={pos}
                  stroke="currentColor"
                  className="text-muted-foreground"
                />
              </g>
            )
          })}
        </svg>
      </div>
    </>
  )
}
