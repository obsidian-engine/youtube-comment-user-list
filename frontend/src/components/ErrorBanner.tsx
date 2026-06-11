interface Props {
  message: string
}

export function ErrorBanner({ message }: Props) {
  return (
    <div
      role="alert"
      aria-live="assertive"
      className="border-l-4 px-4 py-3 text-sm"
      style={{
        borderLeftColor: 'var(--c-error)',
        background: 'rgba(179,0,27,0.06)',
        color: 'var(--c-error)',
        borderTop: '1px solid rgba(179,0,27,0.18)',
        borderRight: '1px solid rgba(179,0,27,0.18)',
        borderBottom: '1px solid rgba(179,0,27,0.18)',
      }}
    >
      {message}
    </div>
  )
}
