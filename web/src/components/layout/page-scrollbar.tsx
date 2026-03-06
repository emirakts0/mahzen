import { useEffect, useRef, useState, useCallback } from "react"

/**
 * Overlay page scrollbar.
 *
 * The native html scrollbar is hidden (scrollbar-width: none) so the fixed
 * Grainient background stays seamless edge-to-edge. This component renders a
 * glass-themed thumb via position:fixed — it floats above all content and
 * never affects layout or the background.
 */
export function PageScrollbar() {
  const [thumbHeight, setThumbHeight] = useState(0)
  const [thumbTop, setThumbTop] = useState(0)
  const [visible, setVisible] = useState(false)
  const [hovered, setHovered] = useState(false)

  const isDragging = useRef(false)
  const dragStartY = useRef(0)
  const dragStartScroll = useRef(0)
  const thumbHeightRef = useRef(0)

  const update = useCallback(() => {
    const { scrollTop, scrollHeight, clientHeight } = document.documentElement
    if (scrollHeight <= clientHeight) {
      setVisible(false)
      return
    }
    setVisible(true)
    const ratio = clientHeight / scrollHeight
    const tHeight = Math.max(ratio * clientHeight, 32)
    const maxScroll = scrollHeight - clientHeight
    const tTop = maxScroll > 0
      ? (scrollTop / maxScroll) * (clientHeight - tHeight)
      : 0
    thumbHeightRef.current = tHeight
    setThumbHeight(tHeight)
    setThumbTop(tTop)
  }, [])

  // Listen to scroll + resize/content changes
  useEffect(() => {
    window.addEventListener("scroll", update, { passive: true })
    const ro = new ResizeObserver(update)
    ro.observe(document.documentElement)
    ro.observe(document.body)
    update()
    return () => {
      window.removeEventListener("scroll", update)
      ro.disconnect()
    }
  }, [update])

  // Drag handling
  const handleMouseDown = useCallback((e: React.MouseEvent) => {
    isDragging.current = true
    dragStartY.current = e.clientY
    dragStartScroll.current = document.documentElement.scrollTop
    e.preventDefault()
  }, [])

  useEffect(() => {
    const onMouseMove = (e: MouseEvent) => {
      if (!isDragging.current) return
      const { scrollHeight, clientHeight } = document.documentElement
      const tHeight = thumbHeightRef.current
      const trackRange = clientHeight - tHeight
      if (trackRange <= 0) return
      const dy = e.clientY - dragStartY.current
      const scrollRange = scrollHeight - clientHeight
      document.documentElement.scrollTop = Math.max(
        0,
        Math.min(dragStartScroll.current + (dy / trackRange) * scrollRange, scrollRange)
      )
    }
    const onMouseUp = () => {
      isDragging.current = false
    }
    window.addEventListener("mousemove", onMouseMove)
    window.addEventListener("mouseup", onMouseUp)
    return () => {
      window.removeEventListener("mousemove", onMouseMove)
      window.removeEventListener("mouseup", onMouseUp)
    }
  }, [])

  // Track click (jump to position)
  const handleTrackClick = useCallback((e: React.MouseEvent<HTMLDivElement>) => {
    // Ignore if click originated from the thumb
    if ((e.target as HTMLElement).dataset.thumb) return
    const { clientHeight } = document.documentElement
    const rect = e.currentTarget.getBoundingClientRect()
    const clickY = e.clientY - rect.top
    const tHeight = thumbHeightRef.current
    const { scrollHeight } = document.documentElement
    const scrollRange = scrollHeight - clientHeight
    const ratio = (clickY - tHeight / 2) / (clientHeight - tHeight)
    document.documentElement.scrollTop = Math.max(0, Math.min(ratio * scrollRange, scrollRange))
  }, [])

  if (!visible) return null

  return (
    <div
      onClick={handleTrackClick}
      style={{
        position: "fixed",
        top: 0,
        right: 0,
        width: "10px",
        height: "100vh",
        zIndex: 9999,
        pointerEvents: "auto",
        // Track is fully transparent — Grainient shows through
        background: "transparent",
      }}
    >
      <div
        data-thumb="true"
        onMouseDown={handleMouseDown}
        onMouseEnter={() => setHovered(true)}
        onMouseLeave={() => setHovered(false)}
        style={{
          position: "absolute",
          right: "3px",
          width: "4px",
          top: thumbTop,
          height: thumbHeight,
          borderRadius: "9999px",
          background: hovered ? "var(--glass-icon)" : "var(--glass-border)",
          transition: "background 0.15s, width 0.15s, right 0.15s",
          cursor: "pointer",
          // No backdrop-filter — keep it lightweight and truly transparent
        }}
      />
    </div>
  )
}
