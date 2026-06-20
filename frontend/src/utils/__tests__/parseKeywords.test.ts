import { describe, test, expect } from 'vitest'
import { parseKeywords } from '../parseKeywords'

describe('parseKeywords', () => {
  test('改行区切りで分割する', () => {
    expect(parseKeywords('A\nB')).toEqual(['A', 'B'])
  })

  test('カンマ区切りで分割する', () => {
    expect(parseKeywords('A,B')).toEqual(['A', 'B'])
  })

  test('重複を排除する', () => {
    expect(parseKeywords('A, A')).toEqual(['A'])
  })

  test('前後の空白を除去し空エントリを除く', () => {
    expect(parseKeywords(' A , , B ')).toEqual(['A', 'B'])
  })

  test('空文字列は空配列を返す', () => {
    expect(parseKeywords('')).toEqual([])
  })
})
