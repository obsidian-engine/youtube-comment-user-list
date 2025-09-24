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
    // マウント時にアニメーション開始
    setIsVisible(true)

    // 指定時間後に自動で閉じる
    const timer = setTimeout(() => {
      setIsVisible(false)
      // フェードアウトアニメーション完了後にonCloseを呼ぶ
      setTimeout(onClose, 300)
    }, duration)

    return () => clearTimeout(timer)
  }, [duration, onClose])

  const typeStyles = {
    info: 'bg-sky-50 text-sky-800 ring-sky-300/60',
    success: 'bg-green-50 text-green-800 ring-green-300/60', 
    error: 'bg-rose-50 text-rose-800 ring-rose-300/60'
  }

  const iconByType = {
    info: '💡',
    success: '✅',
    error: '❌'
  }

  return (
    <div
      className={`fixed top-6 right-6 z-50 min-w-[300px] max-w-[400px] px-4 py-3 rounded-lg ring-1 shadow-lg transition-all duration-300 ${
        isVisible ? 'opacity-100 transform translate-y-0' : 'opacity-0 transform -translate-y-2'
      } ${typeStyles[type]}`}
      role="status"
      aria-live="polite"
    >
      <div className="flex items-center gap-3">
        <span className="text-lg">{iconByType[type]}</span>
        <span className="flex-1 font-medium">{message}</span>
        <button
          onClick={() => {
            setIsVisible(false)
            setTimeout(onClose, 300)
          }}
          className="text-current hover:opacity-70 transition-opacity"
          aria-label="トーストを閉じる"
        >
          ×
        </button>
      </div>
    </div>
  )
}