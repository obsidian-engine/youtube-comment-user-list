import { describe, it, expect } from 'vitest'
import { isTextTooLong, truncateText } from './textUtils'

describe('textUtils', () => {
  describe('isTextTooLong', () => {
    it('英数字のみの文字列で長さ判定ができる', () => {
      expect(isTextTooLong('', 20)).toBe(false)
      expect(isTextTooLong('short', 20)).toBe(false)
      expect(isTextTooLong('this is exactly twenty', 20)).toBe(true)
      expect(isTextTooLong('this is a very long string that exceeds the limit', 20)).toBe(true)
    })

    it('日本語文字で長さ判定ができる（日本語文字は2文字分）', () => {
      expect(isTextTooLong('短い', 20)).toBe(false) // 4文字分
      expect(isTextTooLong('これは日本語です', 20)).toBe(false) // 16文字分
      expect(isTextTooLong('これは非常に長い日本語の文字列です', 20)).toBe(true) // 34文字分
    })

    it('日本語と英数字が混在する文字列で長さ判定ができる', () => {
      expect(isTextTooLong('Hello世界', 20)).toBe(false) // 9文字分
      expect(isTextTooLong('Hello世界! This is very long text', 20)).toBe(true) // 40文字分超
    })
  })

  describe('truncateText', () => {
    it('短い文字列はそのまま返す', () => {
      expect(truncateText('short', 20)).toBe('short')
    })

    it('長い文字列を適切に省略する', () => {
      const result = truncateText('this is a very long string that exceeds limit', 20)
      expect(result).toMatch(/\.\.\.$/)
      expect(result.length).toBeLessThanOrEqual(20)
    })

    it('空文字に対して適切に処理する', () => {
      expect(truncateText('', 20)).toBe('')
    })
  })
})