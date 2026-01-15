import { useState, useCallback, useEffect } from 'react'

const HIDDEN_KEY = 'comment-search-hidden'

const loadHidden = (): Record<string, boolean> => {
  const data = localStorage.getItem(HIDDEN_KEY)
  return data ? JSON.parse(data) : {}
}

const saveHidden = (hidden: Record<string, boolean>): void => {
  localStorage.setItem(HIDDEN_KEY, JSON.stringify(hidden))
}

const clearHidden = (): void => {
  localStorage.removeItem(HIDDEN_KEY)
}

export function useHiddenState() {
  const [hidden, setHidden] = useState<Record<string, boolean>>(() => loadHidden())

  // LocalStorageへの保存
  useEffect(() => {
    saveHidden(hidden)
  }, [hidden])

  const hide = useCallback((commentId: string) => {
    setHidden((prev) => ({
      ...prev,
      [commentId]: true,
    }))
  }, [])

  const hideAll = useCallback((commentIds: string[]) => {
    setHidden((prev) => {
      const newHidden = { ...prev }
      commentIds.forEach((id) => {
        newHidden[id] = true
      })
      return newHidden
    })
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
