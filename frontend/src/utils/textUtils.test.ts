import { describe, it, expect } from 'vitest'
import { isTextTooLong, truncateText, isJapaneseTextTooLong, truncateJapaneseText } from './textUtils'

describe('textUtils', () => {
  describe('isTextTooLong', () => {
    // 日本語20文字制限のテストケース追加
    it('日本語文字がちょうど20文字（10個）の場合はfalse', () => {
      expect(isTextTooLong('あいうえおかきくけこ', 20)).toBe(false) // ちょうど20文字分（10個×2）
    })

    it('日本語文字が21文字分（11個）以上の場合はtrue', () => {
      expect(isTextTooLong('あいうえおかきくけこさ', 20)).toBe(true) // 22文字分（11個×2）
    })

    it('日本語10文字と半角1文字を超えるとtrue', () => {
      expect(isTextTooLong('あいうえおかきくけこA', 20)).toBe(true) // 21文字分（10個×2 + 1）
    })

    it('日本語9文字と半角2文字はfalse', () => {
      expect(isTextTooLong('あいうえおかきくけAB', 20)).toBe(false) // 20文字分（9個×2 + 2）
    })

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

  describe('isJapaneseTextTooLong', () => {
    it('日本語文字が20文字ちょうどの場合はfalse', () => {
      expect(isJapaneseTextTooLong('あいうえおかきくけこさしすせそたちつてと', 20)).toBe(false)
    })

    it('日本語文字が21文字の場合はtrue', () => {
      expect(isJapaneseTextTooLong('あいうえおかきくけこさしすせそたちつてとな', 20)).toBe(true)
    })

    it('英数字20文字はfalse', () => {
      expect(isJapaneseTextTooLong('12345678901234567890', 20)).toBe(false)
    })

    it('英数字21文字はtrue', () => {
      expect(isJapaneseTextTooLong('123456789012345678901', 20)).toBe(true)
    })

    it('日本語と英数字混在で合計20文字はfalse', () => {
      expect(isJapaneseTextTooLong('あいうえおかきくけこHello1234', 20)).toBe(false) // 合計20文字
    })

    it('日本語と英数字混在で合計21文字はtrue', () => {
      expect(isJapaneseTextTooLong('あいうえおかきくけこさHello12345', 20)).toBe(true) // 合計21文字
    })

    it('空文字はfalse', () => {
      expect(isJapaneseTextTooLong('', 20)).toBe(false)
    })

    it('nullまたはundefinedの場合はfalse', () => {
      expect(isJapaneseTextTooLong(null as any, 20)).toBe(false)
      expect(isJapaneseTextTooLong(undefined as any, 20)).toBe(false)
    })
  })

  describe('truncateJapaneseText', () => {
    it('日本語20文字ちょうどはそのまま返す', () => {
      const text = 'あいうえおかきくけこさしすせそたちつてと'
      expect(truncateJapaneseText(text, 20)).toBe(text)
    })

    it('日本語21文字は17文字+...になる', () => {
      const result = truncateJapaneseText('あいうえおかきくけこさしすせそたちつてとな', 20)
      expect(result).toBe('あいうえおかきくけこさしすせそたち...')
    })

    it('英数字21文字は17文字+...になる', () => {
      const result = truncateJapaneseText('123456789012345678901', 20)
      expect(result).toBe('12345678901234567...')
    })

    it('日本語と英数字混在で適切に省略される', () => {
      const result = truncateJapaneseText('あいうえおかきくけこさHello12345', 20)
      expect(result).toBe('あいうえおかきくけこさHello1...')
    })

    it('短いテキストはそのまま返す', () => {
      expect(truncateJapaneseText('短いテキスト', 20)).toBe('短いテキスト')
    })

    it('空文字はそのまま返す', () => {
      expect(truncateJapaneseText('', 20)).toBe('')
    })

    it('nullまたはundefinedの場合は空文字を返す', () => {
      expect(truncateJapaneseText(null as any, 20)).toBe('')
      expect(truncateJapaneseText(undefined as any, 20)).toBe('')
    })
  })

  describe('truncateText', () => {
    // 日本語20文字制限のテストケース追加
    it('日本語10文字はそのまま返す', () => {
      expect(truncateText('あいうえおかきくけこ', 20)).toBe('あいうえおかきくけこ')
    })

    it('日本語11文字は8文字+...になる', () => {
      const result = truncateText('あいうえおかきくけこさ', 20)
      expect(result).toBe('あいうえおかきく...')  // 8個×2 + 3(...)= 19文字分
    })

    it('日本語と英数字混在の場合も適切に省略される', () => {
      const result = truncateText('Hello世界これは長いテキストです', 20)
      expect(result).toMatch(/\.\.\.$/)
      // "Hello世界" で5+4=9文字分なので、もう少し入る
    })

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