import { useState, useCallback, useEffect } from 'react'

const HIDDEN_KEY = 'comment-search-hidden'

const loadHidden = (): Record<string, boolean> => {
  try {
    const data = localStorage.getItem(HIDDEN_KEY)
    return data ? JSON.parse(data) : {}
  } catch (e) {
    console.warn('Failed to load hidden state from localStorage', e)
    return {}
  }
}

const saveHidden = (hidden: Record<string, boolean>): void => {
  try {
    localStorage.setItem(HIDDEN_KEY, JSON.stringify(hidden))
  } catch (e) {
    console.error('Failed to save hidden state to localStorage', e)
  }
}

const clearHidden = (): void => {
  localStorage.removeItem(HIDDEN_KEY)
}

export function useHiddenState() {
  const [hidden, setHidden] = useState<Record<string, boolean>>(() => loadHidden())

  // LocalStorageへの保存（デバウンス付き）
  useEffect(() => {
    const timer = setTimeout(() => {
      saveHidden(hidden)
    }, 300)
    return () => clearTimeout(timer)
  }, [hidden])

  const hide = useCallback((commentId: string) => {
    setHidden((prev) => ({
      ...prev,
      [commentId]: true,
    }))
  }, [])

  const hideAll = useCallback((commentIds: string[]) => {
    setHidden((prev) => ({
      ...prev,
      ...Object.fromEntries(commentIds.map((id) => [id, true])),
    }))
  }, [])

  const isHidden = useCallback(
    (commentId: string) => {
      return !!hidden[commentId]
    },
    [hidden],
  )

  const clear = useCallback(() => {
    setHidden({})
    clearHidden()
  }, [])

  const hiddenCount = Object.values(hidden).filter(Boolean).length

  return { hidden, hide, hideAll, isHidden, clear, hiddenCount }
}
