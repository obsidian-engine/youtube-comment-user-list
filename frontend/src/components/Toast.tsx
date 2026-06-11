import { useState, useEffect } from 'react'

interface ToastProps {
  message: string
  type?: 'info' | 'success' | 'error'
  duration?: number
  onClose: () => void
}

export function Toast({ message, type = 'success', duration = 3000, onClose }: ToastProps) {
  const [isVisible, setIsVisible] = useState(false)

  useEffect(() => {
    setIsVisible(true)
    const timer = setTimeout(() => {
      setIsVisible(false)
      setTimeout(onClose, 300)
    }, duration)
    return () => clearTimeout(timer)
  }, [duration, onClose])

  const bgByType: Record<string, string> = {
    info: 'var(--c-accent-2)',
    success: 'var(--c-success)',
    error: 'var(--c-error)',
  }

  const iconByType: Record<string, string> = {
    info: '💡',
    success: '✅',
    error: '❌',
  }

  return (
    <div
      style={{
        position: 'fixed',
        top: '72px',
        right: '24px',
        zIndex: 100,
        minWidth: '300px',
        maxWidth: '400px',
        padding: '12px 16px',
        background: bgByType[type],
        color: '#fff',
        fontFamily: 'var(--f-mono)',
        fontSize: '12px',
        letterSpacing: '0.1em',
        border: `1px solid ${bgByType[type]}`,
        boxShadow: '0 4px 24px rgba(0,0,0,0.15)',
        opacity: isVisible ? 1 : 0,
        transform: isVisible ? 'translateY(0)' : 'translateY(-8px)',
        transition: 'opacity 0.3s, transform 0.3s',
        pointerEvents: isVisible ? 'auto' : 'none',
      }}
      role="status"
      aria-live="polite"
    >
      <div style={{ display: 'flex', alignItems: 'center', gap: '12px' }}>
        <span style={{ fontSize: '16px' }}>{iconByType[type]}</span>
        <span style={{ flex: 1 }}>{message}</span>
        <button
          onClick={() => {
            setIsVisible(false)
            setTimeout(onClose, 300)
          }}
          style={{
            background: 'none',
            border: 'none',
            color: '#fff',
            cursor: 'pointer',
            fontSize: '18px',
            opacity: 0.8,
            padding: '0 4px',
          }}
          aria-label="トーストを閉じる"
        >
          ×
        </button>
      </div>
    </div>
  )
}
