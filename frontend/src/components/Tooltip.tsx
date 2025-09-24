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
      let left = targetRect.left + (targetRect.width / 2) - (tooltipRect.width / 2)

      // 画面右端を超える場合の調整
      if (left + tooltipRect.width > viewportWidth - 16) {
        left = viewportWidth - tooltipRect.width - 16
      }

      // 画面左端を超える場合の調整
      if (left < 16) {
        left = 16
      }

      // 画面下端を超える場合は上側に表示
      if (top + tooltipRect.height > viewportHeight - 16) {
        top = targetRect.top - tooltipRect.height - 8
      }

      setPosition({ top, left })
    }
  }, [isVisible])

  const handleMouseEnter = () => {
    if (!disabled && content.trim()) {
      setIsVisible(true)
    }
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
          className="fixed z-50 px-3 py-2 text-sm text-white bg-slate-900 dark:bg-slate-700 rounded-lg shadow-lg border border-slate-700 dark:border-slate-500 max-w-xs break-words"
          style={{
            top: `${position.top}px`,
            left: `${position.left}px`,
          }}
        >
          {content}
          <div className="absolute -top-1 left-1/2 transform -translate-x-1/2 w-2 h-2 bg-slate-900 dark:bg-slate-700 border-l border-t border-slate-700 dark:border-slate-500 rotate-45" />
        </div>
      )}
    </>
  )
}