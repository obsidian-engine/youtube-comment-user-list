import { useState, useRef, useEffect, ReactNode } from 'react'

interface TooltipProps {
  children: ReactNode
  content: string
  disabled?: boolean
  className?: string
}

export function Tooltip({ children, content, disabled = false, className = '' }: TooltipProps) {
  const [isVisible, setIsVisible] = useState(false)
  const [position, setPosition] = useState({ top: 0, left: 0 })
  const targetRef = useRef<HTMLDivElement>(null)
  const tooltipRef = useRef<HTMLDivElement>(null)

  useEffect(() => {
    if (isVisible && targetRef.current && tooltipRef.current) {
      const targetRect = targetRef.current.getBoundingClientRect()
      const tooltipRect = tooltipRef.current.getBoundingClientRect()
      const viewportWidth = window.innerWidth
      const viewportHeight = window.innerHeight

      let top = targetRect.bottom + 8
      let left = targetRect.left + targetRect.width / 2 - tooltipRect.width / 2

      if (left + tooltipRect.width > viewportWidth - 16) {
        left = viewportWidth - tooltipRect.width - 16
      }
      if (left < 16) left = 16
      if (top + tooltipRect.height > viewportHeight - 16) {
        top = targetRect.top - tooltipRect.height - 8
      }

      setPosition({ top, left })
    }
  }, [isVisible])

  const handleMouseEnter = () => {
    if (!disabled && content.trim()) setIsVisible(true)
  }

  const handleMouseLeave = () => {
    setIsVisible(false)
  }

  return (
    <>
      <div
        ref={targetRef}
        className={className}
        onMouseEnter={handleMouseEnter}
        onMouseLeave={handleMouseLeave}
      >
        {children}
      </div>

      {isVisible && !disabled && (
        <div
          ref={tooltipRef}
          style={{
            position: 'fixed',
            zIndex: 50,
            padding: '8px 12px',
            background: 'var(--c-ink)',
            color: '#fff',
            fontFamily: 'var(--f-mono)',
            fontSize: '12px',
            letterSpacing: '0.06em',
            boxShadow: '0 4px 16px rgba(0,0,0,0.2)',
            border: '1px solid rgba(255,255,255,0.1)',
            maxWidth: '300px',
            wordBreak: 'break-word',
            top: `${position.top}px`,
            left: `${position.left}px`,
          }}
        >
          {content}
          <div
            style={{
              position: 'absolute',
              top: '-5px',
              left: '50%',
              transform: 'translateX(-50%) rotate(45deg)',
              width: '8px',
              height: '8px',
              background: 'var(--c-ink)',
              borderLeft: '1px solid rgba(255,255,255,0.1)',
              borderTop: '1px solid rgba(255,255,255,0.1)',
            }}
          />
        </div>
      )}
    </>
  )
}
