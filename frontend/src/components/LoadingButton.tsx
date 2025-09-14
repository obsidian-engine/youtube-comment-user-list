import React from 'react'

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

const clsByVariant: Record<Variant, string> = {
  primary:
    'px-3.5 py-2 rounded-md bg-neutral-900 text-white hover:bg-neutral-800 disabled:opacity-60 disabled:cursor-not-allowed dark:bg-white dark:text-neutral-900 dark:hover:bg-white/90 transition text-[14px]',
  outline:
    'px-3.5 py-2 rounded-md bg-white/90 dark:bg-white/5 border border-slate-300/80 dark:border-white/10 hover:bg-white dark:hover:bg-white/10 disabled:opacity-60 disabled:cursor-not-allowed transition text-[14px]',
}

const clsBySize: Record<Size, string> = {
  sm: 'text-[13px]',
  md: 'text-[14px]',
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

  return (
    <button
      type={type}
      aria-label={ariaLabel}
      aria-busy={isLoading}
      disabled={isDisabled}
      onClick={() => {
        if (isDisabled) return
        onClick?.()
      }}
      className={`${clsByVariant[variant]} ${clsBySize[size]} ${className}`}
    >
      {isLoading && (
        <span
          aria-hidden="true"
          className="mr-2 inline-block h-4 w-4 border-2 border-current border-r-transparent rounded-full animate-spin align-[-2px]"
        />
      )}
      {label}
    </button>
  )
}

