const KEYWORDS_KEY = 'comment-search-keywords'
const CHECKED_KEY = 'comment-search-checked'

export const loadKeywords = (): string[] => {
  const data = localStorage.getItem(KEYWORDS_KEY)
  return data ? JSON.parse(data) : ['メモ'] // デフォルト値
}

export const saveKeywords = (keywords: string[]): void => {
  localStorage.setItem(KEYWORDS_KEY, JSON.stringify(keywords))
}

export const loadChecked = (): Record<string, boolean> => {
  const data = localStorage.getItem(CHECKED_KEY)
  return data ? JSON.parse(data) : {}
}

export const saveChecked = (checked: Record<string, boolean>): void => {
  localStorage.setItem(CHECKED_KEY, JSON.stringify(checked))
}

export const clearChecked = (): void => {
  localStorage.removeItem(CHECKED_KEY)
}
