/**
 * 日本語テキストが指定した文字数を超えるかどうかを判定します（実際の文字数でカウント）
 */
export function isJapaneseTextTooLong(text: string, maxLength: number = 20): boolean {
  if (!text) return false
  return text.length > maxLength
}

/**
 * 日本語テキストを指定した文字数で省略します（実際の文字数でカウント）
 */
export function truncateJapaneseText(text: string, maxLength: number = 20): string {
  if (!text) return ''
  if (!isJapaneseTextTooLong(text, maxLength)) {
    return text
  }

  return text.slice(0, maxLength - 3) + '...'
}

/**
 * テキストが指定した文字数を超えるかどうかを判定します
 * 日本語文字（ひらがな、カタカナ、漢字）は2文字分として計算
 */
export function isTextTooLong(text: string, maxLength: number = 20): boolean {
  if (!text) return false
  
  let length = 0
  for (const char of text) {
    // 日本語文字（ひらがな、カタカナ、漢字、全角記号）は2文字分として計算
    if (/[\u3040-\u309F\u30A0-\u30FF\u4E00-\u9FAF\uFF00-\uFFEF]/.test(char)) {
      length += 2
    } else {
      length += 1
    }
  }
  
  return length > maxLength
}

/**
 * テキストを指定した長さで省略します
 */
export function truncateText(text: string, maxLength: number = 20): string {
  if (!text || !isTextTooLong(text, maxLength)) {
    return text
  }
  
  let length = 0
  let result = ''
  
  for (const char of text) {
    const charLength = /[\u3040-\u309F\u30A0-\u30FF\u4E00-\u9FAF\uFF00-\uFFEF]/.test(char) ? 2 : 1
    
    if (length + charLength > maxLength - 3) { // "..." の分を考慮
      break
    }
    
    result += char
    length += charLength
  }
  
  return result + '...'
}