import React from 'react'
import { logger } from '../utils/logger'

type Variant = 'primary' | 'outline'
type Size = 'sm' | 'md'

type Props = {
  variant?: Variant
  size?: Size
  isLoading?: boolean
  disabled?: boolean
  type?: 'button' | 'submit'
  ariaLabel?: string
  loadingText?: string
  onClick?: () => void | Promise<void>
  children: React.ReactNode
  className?: string
}

const primaryStyle: React.CSSProperties = {
  display: 'inline-flex',
  alignItems: 'center',
  gap: '8px',
  padding: '9px 18px',
  background: 'var(--c-ink)',
  color: '#fff',
  fontFamily: 'var(--f-mono)',
  fontWeight: 700,
  fontSize: '12px',
  letterSpacing: '0.16em',
  textTransform: 'uppercase',
  border: '1px solid var(--c-ink)',
  cursor: 'pointer',
  transition: 'background 0.2s, border-color 0.2s',
}

const outlineStyle: React.CSSProperties = {
  display: 'inline-flex',
  alignItems: 'center',
  gap: '8px',
  padding: '9px 18px',
  background: 'transparent',
  color: 'var(--c-ink)',
  fontFamily: 'var(--f-mono)',
  fontWeight: 700,
  fontSize: '12px',
  letterSpacing: '0.16em',
  textTransform: 'uppercase',
  border: '2px solid var(--c-ink)',
  cursor: 'pointer',
  transition: 'background 0.2s, border-color 0.2s, color 0.2s',
}

export function LoadingButton({
  variant = 'primary',
  size = 'md',
  isLoading = false,
  disabled = false,
  type = 'button',
  ariaLabel,
  loadingText,
  onClick,
  children,
  className = '',
}: Props) {
  const isDisabled = disabled || isLoading
  const label = isLoading && loadingText ? loadingText : children
  const effectiveAriaLabel = isLoading && loadingText ? loadingText : ariaLabel
  const baseStyle = variant === 'primary' ? primaryStyle : outlineStyle
  const sizeAdj = size === 'sm' ? { fontSize: '11px' } : {}
  const disabledAdj = isDisabled ? { opacity: 0.5, cursor: 'not-allowed' } : {}

  return (
    <button
      type={type}
      aria-label={effectiveAriaLabel}
      aria-busy={isLoading}
      disabled={isDisabled}
      onClick={async () => {
        if (isDisabled) return
        try {
          await onClick?.()
        } catch (error) {
          logger.error('LoadingButton onClick error:', error)
          throw error
        }
      }}
      style={{ ...baseStyle, ...sizeAdj, ...disabledAdj }}
      className={className}
    >
      {isLoading && (
        <span
          aria-hidden="true"
          style={{
            display: 'inline-block',
            width: '14px',
            height: '14px',
            border: '2px solid currentColor',
            borderRightColor: 'transparent',
            borderRadius: '50%',
            animation: 'spin 0.7s linear infinite',
          }}
        />
      )}
      {label}
    </button>
  )
}
