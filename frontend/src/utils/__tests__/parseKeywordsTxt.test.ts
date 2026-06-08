import { describe, it, expect } from 'vitest'
import { parseKeywordsTxt } from '../parseKeywordsTxt'

describe('parseKeywordsTxt - 入力境界', () => {
  it('空文字列は keywords/invalid 共に空', () => {
    expect(parseKeywordsTxt('')).toEqual({ keywords: [], invalid: [] })
  })

  it('単一行（改行なし）', () => {
    expect(parseKeywordsTxt('hoge').keywords).toEqual(['hoge'])
  })

  it('空行のみ', () => {
    expect(parseKeywordsTxt('\n\n\n')).toEqual({ keywords: [], invalid: [] })
  })

  it('空白のみの行のみ', () => {
    expect(parseKeywordsTxt('   \n\t\n')).toEqual({
      keywords: [],
      invalid: [],
    })
  })
})

describe('parseKeywordsTxt - 改行種別', () => {
  it('LF 区切り', () => {
    expect(parseKeywordsTxt('hoge\nfuga\npiyo').keywords).toEqual(['hoge', 'fuga', 'piyo'])
  })

  it('CRLF 区切り', () => {
    expect(parseKeywordsTxt('hoge\r\nfuga\r\npiyo').keywords).toEqual(['hoge', 'fuga', 'piyo'])
  })

  it('LF と CRLF 混在', () => {
    expect(parseKeywordsTxt('hoge\nfuga\r\npiyo').keywords).toEqual(['hoge', 'fuga', 'piyo'])
  })

  it('末尾改行があっても末尾に空要素が入らない', () => {
    expect(parseKeywordsTxt('hoge\nfuga\n').keywords).toEqual(['hoge', 'fuga'])
  })

  it('末尾 CRLF も同様', () => {
    expect(parseKeywordsTxt('hoge\r\nfuga\r\n').keywords).toEqual(['hoge', 'fuga'])
  })

  it('先頭空行は無視', () => {
    expect(parseKeywordsTxt('\n\nhoge\nfuga').keywords).toEqual(['hoge', 'fuga'])
  })
})

describe('parseKeywordsTxt - 空白の扱い', () => {
  it('行の前後 ASCII 空白を除去', () => {
    expect(parseKeywordsTxt('  hoge  \n\tfuga\t').keywords).toEqual(['hoge', 'fuga'])
  })

  it('前後タブのみも除去', () => {
    expect(parseKeywordsTxt('\thoge\t').keywords).toEqual(['hoge'])
  })

  it('空白のみの行は空行扱い', () => {
    expect(parseKeywordsTxt('hoge\n   \nfuga').keywords).toEqual(['hoge', 'fuga'])
  })

  it('単語中の空白は保持される（部分一致は countVotes 側で弾く）', () => {
    expect(parseKeywordsTxt('hoge fuga').keywords).toEqual(['hoge fuga'])
  })

  it('全角空白のみの行も空扱い（trim 仕様）', () => {
    expect(parseKeywordsTxt('　　　').keywords).toEqual([])
  })
})

describe('parseKeywordsTxt - 重複排除', () => {
  it('完全一致の重複は先勝ちで排除', () => {
    expect(parseKeywordsTxt('hoge\nfuga\nhoge').keywords).toEqual(['hoge', 'fuga'])
  })

  it('trim 後の同値も重複扱い', () => {
    expect(parseKeywordsTxt('hoge\n  hoge  ').keywords).toEqual(['hoge'])
  })

  it('大文字小文字違いは別物として保持（HOGE と hoge は両方残る）', () => {
    expect(parseKeywordsTxt('hoge\nHOGE').keywords).toEqual(['hoge', 'HOGE'])
  })

  it('同じワードが3回以上出現しても1つだけ残る', () => {
    expect(parseKeywordsTxt('hoge\nhoge\nhoge\nfuga\nhoge').keywords).toEqual(['hoge', 'fuga'])
  })
})

describe('parseKeywordsTxt - カンマ含む行の除外', () => {
  it('ASCII カンマ含む行は invalid に積む', () => {
    const result = parseKeywordsTxt('hoge\n1,000円\nfuga')
    expect(result.keywords).toEqual(['hoge', 'fuga'])
    expect(result.invalid).toEqual(['1,000円'])
  })

  it('先頭/中間/末尾どこに , があっても除外', () => {
    const result = parseKeywordsTxt(',abc\na,bc\nabc,\nok')
    expect(result.keywords).toEqual(['ok'])
    expect(result.invalid).toEqual([',abc', 'a,bc', 'abc,'])
  })

  it('複数カンマ含む行も1要素として invalid', () => {
    const result = parseKeywordsTxt('a,b,c,d')
    expect(result.keywords).toEqual([])
    expect(result.invalid).toEqual(['a,b,c,d'])
  })

  it('カンマ単体行も invalid', () => {
    const result = parseKeywordsTxt(',')
    expect(result.keywords).toEqual([])
    expect(result.invalid).toEqual([','])
  })

  it('全角読点「、」は除外対象外（ASCII カンマのみ判定）', () => {
    const result = parseKeywordsTxt('hoge、fuga')
    expect(result.keywords).toEqual(['hoge、fuga'])
    expect(result.invalid).toEqual([])
  })

  it('invalid に積まれる順序は出現順', () => {
    const result = parseKeywordsTxt('a,b\nok\nc,d')
    expect(result.invalid).toEqual(['a,b', 'c,d'])
  })
})

describe('parseKeywordsTxt - Unicode と特殊文字', () => {
  it('日本語キーワード', () => {
    expect(parseKeywordsTxt('賛成\n反対\n保留').keywords).toEqual(['賛成', '反対', '保留'])
  })

  it('絵文字キーワード', () => {
    expect(parseKeywordsTxt('👍\n👎').keywords).toEqual(['👍', '👎'])
  })

  it('混在ワード', () => {
    expect(parseKeywordsTxt('候補A\noption B\n3番').keywords).toEqual(['候補A', 'option B', '3番'])
  })

  it('数字のみキーワード', () => {
    expect(parseKeywordsTxt('1\n2\n3').keywords).toEqual(['1', '2', '3'])
  })

  it('記号キーワード', () => {
    expect(parseKeywordsTxt('○\n×\n△').keywords).toEqual(['○', '×', '△'])
  })

  it('長いワード（100 文字）も保持', () => {
    const long = 'あ'.repeat(100)
    expect(parseKeywordsTxt(long).keywords).toEqual([long])
  })
})

describe('parseKeywordsTxt - 出力契約', () => {
  it('invalid のみで keywords が空でも例外を投げない', () => {
    const result = parseKeywordsTxt('a,b\nc,d')
    expect(result.keywords).toEqual([])
    expect(result.invalid).toEqual(['a,b', 'c,d'])
  })

  it('keywords / invalid は両方常に array', () => {
    const result = parseKeywordsTxt('')
    expect(Array.isArray(result.keywords)).toBe(true)
    expect(Array.isArray(result.invalid)).toBe(true)
  })

  it('純関数（同入力で同出力）', () => {
    const r1 = parseKeywordsTxt('hoge\nfuga')
    const r2 = parseKeywordsTxt('hoge\nfuga')
    expect(r1).toEqual(r2)
  })
})
