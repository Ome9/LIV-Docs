// Mobile-specific CSS optimizations for LIV documents

export const MOBILE_CSS_OPTIMIZATIONS = `
/* Mobile Performance Optimizations */
.liv-mobile-optimized {
  /* Hardware acceleration for better performance */
  -webkit-transform: translateZ(0);
  transform: translateZ(0);
  -webkit-backface-visibility: hidden;
  backface-visibility: hidden;
  -webkit-perspective: 1000px;
  perspective: 1000px;
  
  /* Optimize scrolling */
  -webkit-overflow-scrolling: touch;
  overflow-scrolling: touch;
  
  /* Prevent text selection during gestures */
  -webkit-user-select: none;
  -moz-user-select: none;
  -ms-user-select: none;
  user-select: none;
  
  /* Prevent zoom on input focus */
  -webkit-text-size-adjust: 100%;
  text-size-adjust: 100%;
  
  /* Optimize font rendering */
  -webkit-font-smoothing: antialiased;
  -moz-osx-font-smoothing: grayscale;
  text-rendering: optimizeSpeed;
}

/* Touch-friendly interactive elements */
.liv-mobile-optimized [data-interactive],
.liv-mobile-optimized .liv-interactive {
  min-width: 44px;
  min-height: 44px;
  padding: 8px;
  margin: 4px;
  
  /* Touch feedback */
  -webkit-tap-highlight-color: rgba(0, 0, 0, 0.1);
  tap-highlight-color: rgba(0, 0, 0, 0.1);
  
  /* Prevent accidental selection */
  -webkit-touch-callout: none;
  -webkit-user-select: none;
  user-select: none;
}

/* Optimize animations for mobile */
.liv-mobile-optimized .liv-animating {
  -webkit-transform: translateZ(0);
  transform: translateZ(0);
  will-change: transform, opacity;
  
  /* Use GPU acceleration */
  -webkit-backface-visibility: hidden;
  backface-visibility: hidden;
}

/* Mobile-specific chart optimizations */
.liv-mobile-optimized .liv-chart {
  max-width: 100%;
  height: auto;
  
  /* Optimize SVG rendering */
  shape-rendering: optimizeSpeed;
  text-rendering: optimizeSpeed;
}

.liv-mobile-optimized .liv-chart svg {
  /* Prevent SVG scaling issues on mobile */
  max-width: 100%;
  height: auto;
  
  /* Optimize rendering */
  shape-rendering: geometricPrecision;
}

/* Gesture feedback styles */
.liv-gesture-active {
  -webkit-touch-callout: none;
  -webkit-user-select: none;
  user-select: none;
  
  /* Visual feedback for gestures */
  transition: transform 0.1s ease-out;
}

.liv-gesture-active.liv-pinching {
  transform: scale(var(--gesture-scale, 1));
}

.liv-gesture-active.liv-rotating {
  transform: rotate(var(--gesture-rotation, 0deg));
}

.liv-gesture-active.liv-panning {
  transform: translate(var(--gesture-x, 0px), var(--gesture-y, 0px));
}

/* Mobile breakpoint optimizations */
@media (max-width: 480px) {
  .liv-mobile-optimized {
    font-size: 16px; /* Prevent zoom on iOS */
  }
  
  .liv-mobile-optimized .liv-text {
    line-height: 1.4;
    word-wrap: break-word;
    hyphens: auto;
  }
  
  .liv-mobile-optimized .liv-container {
    padding: 8px;
    margin: 4px 0;
  }
  
  /* Optimize images for small screens */
  .liv-mobile-optimized .liv-image {
    max-width: 100%;
    height: auto;
    image-rendering: -webkit-optimize-contrast;
  }
}

/* Orientation-specific optimizations */
@media (orientation: portrait) {
  .liv-mobile-optimized .liv-landscape-only {
    display: none;
  }
  
  .liv-mobile-optimized .liv-chart {
    /* Adjust chart layout for portrait */
    aspect-ratio: 4/3;
  }
}

@media (orientation: landscape) {
  .liv-mobile-optimized .liv-portrait-only {
    display: none;
  }
  
  .liv-mobile-optimized .liv-chart {
    /* Adjust chart layout for landscape */
    aspect-ratio: 16/9;
  }
}

/* High DPI optimizations */
@media (-webkit-min-device-pixel-ratio: 2), (min-resolution: 192dpi) {
  .liv-mobile-optimized .liv-vector-graphics {
    image-rendering: -webkit-optimize-contrast;
    image-rendering: crisp-edges;
  }
  
  .liv-mobile-optimized .liv-text {
    -webkit-font-smoothing: subpixel-antialiased;
  }
}

/* Performance mode optimizations */
.liv-mobile-optimized.liv-performance-battery {
  /* Disable expensive effects in battery mode */
  animation-duration: 0s !important;
  transition-duration: 0.1s !important;
  
  /* Simplify shadows and effects */
  box-shadow: none !important;
  text-shadow: none !important;
  filter: none !important;
}

.liv-mobile-optimized.liv-performance-balanced {
  /* Reduce animation complexity */
  animation-duration: 0.2s;
  transition-duration: 0.2s;
}

/* Reduced motion support */
@media (prefers-reduced-motion: reduce) {
  .liv-mobile-optimized * {
    animation-duration: 0.01ms !important;
    animation-iteration-count: 1 !important;
    transition-duration: 0.01ms !important;
    scroll-behavior: auto !important;
  }
}

/* Dark mode optimizations for mobile */
@media (prefers-color-scheme: dark) {
  .liv-mobile-optimized {
    /* Optimize for OLED displays */
    background-color: #000000;
    color: #ffffff;
  }
  
  .liv-mobile-optimized .liv-chart {
    /* Invert chart colors for dark mode */
    filter: invert(1) hue-rotate(180deg);
  }
}

/* Bandwidth optimization */
.liv-mobile-optimized.liv-optimize-bandwidth {
  /* Reduce image quality for slow connections */
  image-rendering: pixelated;
  
  /* Disable non-essential animations */
  animation-play-state: paused;
}

.liv-mobile-optimized.liv-optimize-bandwidth .liv-image {
  /* Use lower quality images */
  image-rendering: -webkit-optimize-contrast;
  image-rendering: pixelated;
}

/* Touch-specific styles */
@media (pointer: coarse) {
  .liv-mobile-optimized .liv-button,
  .liv-mobile-optimized .liv-interactive {
    /* Larger touch targets */
    min-width: 48px;
    min-height: 48px;
    padding: 12px;
  }
  
  .liv-mobile-optimized .liv-text {
    /* Better readability on touch devices */
    font-size: 16px;
    line-height: 1.5;
  }
}

/* Hover effects only for non-touch devices */
@media (hover: hover) and (pointer: fine) {
  .liv-mobile-optimized .liv-interactive:hover {
    background-color: rgba(0, 0, 0, 0.05);
    transform: translateY(-1px);
  }
}

/* Focus styles for keyboard navigation */
.liv-mobile-optimized .liv-interactive:focus {
  outline: 2px solid #007bff;
  outline-offset: 2px;
}

/* Loading states optimized for mobile */
.liv-mobile-optimized .liv-loading {
  /* Simple loading animation for mobile */
  animation: liv-mobile-pulse 1.5s ease-in-out infinite;
}

@keyframes liv-mobile-pulse {
  0%, 100% { opacity: 0.6; }
  50% { opacity: 1; }
}

/* Error states */
.liv-mobile-optimized .liv-error {
  padding: 16px;
  margin: 8px;
  border-radius: 8px;
  background-color: #fff3cd;
  border: 1px solid #ffeaa7;
  color: #856404;
  
  /* Ensure readability on mobile */
  font-size: 14px;
  line-height: 1.4;
}

/* Accessibility improvements for mobile */
.liv-mobile-optimized .liv-sr-only {
  position: absolute;
  width: 1px;
  height: 1px;
  padding: 0;
  margin: -1px;
  overflow: hidden;
  clip: rect(0, 0, 0, 0);
  white-space: nowrap;
  border: 0;
}

/* Print styles for mobile browsers */
@media print {
  .liv-mobile-optimized {
    /* Optimize for mobile printing */
    -webkit-print-color-adjust: exact;
    color-adjust: exact;
  }
  
  .liv-mobile-optimized .liv-interactive {
    /* Show interactive elements in print */
    border: 1px solid #ccc;
  }
}

/* Safe area support for notched devices */
.liv-mobile-optimized {
  padding-left: env(safe-area-inset-left);
  padding-right: env(safe-area-inset-right);
  padding-top: env(safe-area-inset-top);
  padding-bottom: env(safe-area-inset-bottom);
}

/* Viewport units fix for mobile browsers */
.liv-mobile-optimized .liv-fullscreen {
  /* Use CSS custom properties for dynamic viewport */
  height: var(--vh, 1vh);
  width: var(--vw, 1vw);
}

/* Container queries support (when available) */
@container (max-width: 480px) {
  .liv-mobile-optimized .liv-responsive-content {
    font-size: 14px;
    padding: 8px;
  }
}

@container (min-width: 481px) {
  .liv-mobile-optimized .liv-responsive-content {
    font-size: 16px;
    padding: 16px;
  }
}
`;

export function injectMobileStyles(): void {
  // Check if styles are already injected
  if (document.getElementById('liv-mobile-styles')) {
    return;
  }

  const style = document.createElement('style');
  style.id = 'liv-mobile-styles';
  style.textContent = MOBILE_CSS_OPTIMIZATIONS;
  document.head.appendChild(style);
}

export function updateViewportUnits(): void {
  // Fix for mobile viewport units
  const vh = window.innerHeight * 0.01;
  const vw = window.innerWidth * 0.01;
  
  document.documentElement.style.setProperty('--vh', `${vh}px`);
  document.documentElement.style.setProperty('--vw', `${vw}px`);
}

// Initialize viewport units fix
if (typeof window !== 'undefined') {
  updateViewportUnits();
  
  window.addEventListener('resize', updateViewportUnits);
  window.addEventListener('orientationchange', () => {
    // Delay to ensure orientation change is complete
    setTimeout(updateViewportUnits, 100);
  });
}