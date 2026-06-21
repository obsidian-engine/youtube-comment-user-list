import { describe, it, expect } from 'vitest'
import { countVotes } from '../countVotes'
import type { Comment } from '../api'

const c = (
  channelId: string,
  message: string,
  publishedAt: string,
  id = `${channelId}-${publishedAt}`,
): Comment => ({
  id,
  channelId,
  displayName: channelId,
  message,
  publishedAt,
})

describe('countVotes - voters 出力', () => {
  it('exact マッチ時に message を trimmed で保存する', () => {
    const comments = [c('u1', '  hoge  ', '2024-01-01T00:00:01Z')]
    const { voters } = countVotes(comments, ['hoge'])
    expect(voters.hoge).toEqual([{ channelId: 'u1', displayName: 'u1', message: 'hoge' }])
  })

  it('partial マッチ時にコメント全文を message として保存する', () => {
    const comments = [c('u1', '賛成です', '2024-01-01T00:00:01Z')]
    const { voters } = countVotes(comments, ['賛成'], 'partial')
    expect(voters['賛成']).toEqual([{ channelId: 'u1', displayName: 'u1', message: '賛成です' }])
  })

  it('handle がある場合も message と一緒に保存する', () => {
    const comment: Comment = {
      id: '1',
      channelId: 'u1',
      displayName: 'taro',
      handle: '@taro',
      message: 'hoge',
      publishedAt: '2024-01-01T00:00:01Z',
    }
    const { voters } = countVotes([comment], ['hoge'])
    expect(voters.hoge).toEqual([
      { channelId: 'u1', displayName: 'taro', handle: '@taro', message: 'hoge' },
    ])
  })
})

describe('countVotes - 入力境界', () => {
  it('空コメント・空キーワードはすべて0', () => {
    expect(countVotes([], []).counts).toEqual({})
    expect(countVotes([], ['hoge']).counts).toEqual({ hoge: 0 })
  })

  it('キーワードが空配列なら集計しない（コメントあっても {} 返す）', () => {
    expect(countVotes([c('u1', 'hoge', '2024-01-01T00:00:00Z')], []).counts).toEqual({})
  })

  it('対象ワードを含むコメンターが居ないと全 0', () => {
    const comments = [
      c('u1', 'なし', '2024-01-01T00:00:01Z'),
      c('u2', '違う', '2024-01-01T00:00:02Z'),
    ]
    expect(countVotes(comments, ['hoge', 'fuga']).counts).toEqual({
      hoge: 0,
      fuga: 0,
    })
  })
})

describe('countVotes - 完全一致ルール', () => {
  it('完全一致でカウント', () => {
    const comments = [
      c('u1', 'hoge', '2024-01-01T00:00:01Z'),
      c('u2', 'fuga', '2024-01-01T00:00:02Z'),
    ]
    expect(countVotes(comments, ['hoge', 'fuga']).counts).toEqual({
      hoge: 1,
      fuga: 1,
    })
  })

  it('後方部分一致は NG（hogee）', () => {
    expect(countVotes([c('u1', 'hogee', '2024-01-01T00:00:01Z')], ['hoge']).counts).toEqual({
      hoge: 0,
    })
  })

  it('前方部分一致は NG（hhoge）', () => {
    expect(countVotes([c('u1', 'hhoge', '2024-01-01T00:00:01Z')], ['hoge']).counts).toEqual({
      hoge: 0,
    })
  })

  it('中間部分一致は NG（xhogex）', () => {
    expect(countVotes([c('u1', 'xhogex', '2024-01-01T00:00:01Z')], ['hoge']).counts).toEqual({
      hoge: 0,
    })
  })

  it('複合語は NG（hoge fuga）', () => {
    expect(
      countVotes([c('u1', 'hoge fuga', '2024-01-01T00:00:01Z')], ['hoge', 'fuga']).counts,
    ).toEqual({
      hoge: 0,
      fuga: 0,
    })
  })

  it('装飾は NG（hoge!!!）', () => {
    expect(countVotes([c('u1', 'hoge!!!', '2024-01-01T00:00:01Z')], ['hoge']).counts).toEqual({
      hoge: 0,
    })
  })

  it('読点付き NG（hoge、）', () => {
    expect(countVotes([c('u1', 'hoge、', '2024-01-01T00:00:01Z')], ['hoge']).counts).toEqual({
      hoge: 0,
    })
  })

  it('改行混入 NG', () => {
    expect(
      countVotes([c('u1', 'hoge\nfuga', '2024-01-01T00:00:01Z')], ['hoge', 'fuga']).counts,
    ).toEqual({
      hoge: 0,
      fuga: 0,
    })
  })
})

describe('countVotes - 空白の扱い', () => {
  it('前後 ASCII 空白は trim して一致扱い', () => {
    expect(countVotes([c('u1', '  hoge  ', '2024-01-01T00:00:01Z')], ['hoge']).counts).toEqual({
      hoge: 1,
    })
  })

  it('前後タブも trim 対象', () => {
    expect(countVotes([c('u1', '\thoge\t', '2024-01-01T00:00:01Z')], ['hoge']).counts).toEqual({
      hoge: 1,
    })
  })

  it('前後改行も trim 対象', () => {
    expect(countVotes([c('u1', '\nhoge\n', '2024-01-01T00:00:01Z')], ['hoge']).counts).toEqual({
      hoge: 1,
    })
  })

  it('前後の全角空白も trim される（String.prototype.trim 仕様）', () => {
    expect(countVotes([c('u1', '　hoge　', '2024-01-01T00:00:01Z')], ['hoge']).counts).toEqual({
      hoge: 1,
    })
  })

  it('中央の空白は除去されない（hoge fuga 全体は完全一致しない）', () => {
    expect(countVotes([c('u1', 'ho ge', '2024-01-01T00:00:01Z')], ['hoge']).counts).toEqual({
      hoge: 0,
    })
  })

  it('空白のみのメッセージはマッチしない', () => {
    expect(countVotes([c('u1', '   ', '2024-01-01T00:00:01Z')], ['hoge']).counts).toEqual({
      hoge: 0,
    })
  })
})

describe('countVotes - 大文字小文字（厳密区別）', () => {
  it('hoge ≠ HOGE', () => {
    const comments = [
      c('u1', 'hoge', '2024-01-01T00:00:01Z'),
      c('u2', 'HOGE', '2024-01-01T00:00:02Z'),
    ]
    expect(countVotes(comments, ['hoge']).counts).toEqual({ hoge: 1 })
    expect(countVotes(comments, ['HOGE']).counts).toEqual({ HOGE: 1 })
  })

  it('Hoge / hOgE などの混在も区別', () => {
    const comments = [
      c('u1', 'Hoge', '2024-01-01T00:00:01Z'),
      c('u2', 'hOgE', '2024-01-01T00:00:02Z'),
    ]
    expect(countVotes(comments, ['Hoge', 'hOgE']).counts).toEqual({
      Hoge: 1,
      hOgE: 1,
    })
  })

  it('両方 keyword 登録して別々に集計可能', () => {
    const comments = [
      c('u1', 'hoge', '2024-01-01T00:00:01Z'),
      c('u2', 'HOGE', '2024-01-01T00:00:02Z'),
      c('u3', 'hoge', '2024-01-01T00:00:03Z'),
    ]
    expect(countVotes(comments, ['hoge', 'HOGE']).counts).toEqual({
      hoge: 2,
      HOGE: 1,
    })
  })
})

describe('countVotes - 1コメンター1票ルール', () => {
  it('同 channelId は最初の対象ワードのみ', () => {
    const comments = [
      c('u1', 'hoge', '2024-01-01T00:00:01Z'),
      c('u1', 'fuga', '2024-01-01T00:00:02Z'),
    ]
    expect(countVotes(comments, ['hoge', 'fuga']).counts).toEqual({
      hoge: 1,
      fuga: 0,
    })
  })

  it('同コメンターが同じワードを連投しても1票', () => {
    const comments = [
      c('u1', 'hoge', '2024-01-01T00:00:01Z'),
      c('u1', 'hoge', '2024-01-01T00:00:02Z'),
      c('u1', 'hoge', '2024-01-01T00:00:03Z'),
    ]
    expect(countVotes(comments, ['hoge']).counts).toEqual({ hoge: 1 })
  })

  it('対象外コメント先行 → 後続の完全一致は採用', () => {
    const comments = [
      c('u1', 'こんにちは', '2024-01-01T00:00:01Z'),
      c('u1', 'hoge', '2024-01-01T00:00:02Z'),
    ]
    expect(countVotes(comments, ['hoge']).counts).toEqual({ hoge: 1 })
  })

  it('対象外コメント先行 → 後続も対象外なら未投票', () => {
    const comments = [
      c('u1', 'こんにちは', '2024-01-01T00:00:01Z'),
      c('u1', 'さよなら', '2024-01-01T00:00:02Z'),
    ]
    expect(countVotes(comments, ['hoge']).counts).toEqual({ hoge: 0 })
  })

  it('最初に hoge → 後で fuga: hoge 確定、fuga 無視', () => {
    const comments = [
      c('u1', 'hoge', '2024-01-01T00:00:01Z'),
      c('u1', 'なし', '2024-01-01T00:00:02Z'),
      c('u1', 'fuga', '2024-01-01T00:00:03Z'),
    ]
    expect(countVotes(comments, ['hoge', 'fuga']).counts).toEqual({
      hoge: 1,
      fuga: 0,
    })
  })
})

describe('countVotes - 順序とソート', () => {
  it('publishedAt 昇順で評価される（入力順序に依存しない）', () => {
    const comments = [
      c('u1', 'fuga', '2024-01-01T00:00:02Z'),
      c('u1', 'hoge', '2024-01-01T00:00:01Z'),
    ]
    expect(countVotes(comments, ['hoge', 'fuga']).counts).toEqual({
      hoge: 1,
      fuga: 0,
    })
  })

  it('publishedAt 同値は配列順に評価', () => {
    const comments = [
      c('u1', 'hoge', '2024-01-01T00:00:01Z'),
      c('u1', 'fuga', '2024-01-01T00:00:01Z'),
    ]
    expect(countVotes(comments, ['hoge', 'fuga']).counts).toEqual({
      hoge: 1,
      fuga: 0,
    })
  })

  it('入力配列を破壊しない（immutable）', () => {
    const comments = [
      c('u2', 'hoge', '2024-01-01T00:00:02Z'),
      c('u1', 'fuga', '2024-01-01T00:00:01Z'),
    ]
    const snapshot = JSON.stringify(comments)
    countVotes(comments, ['hoge', 'fuga'])
    expect(JSON.stringify(comments)).toBe(snapshot)
  })

  it('入力 keywords 配列を破壊しない', () => {
    const keywords = ['hoge', 'fuga']
    const snapshot = [...keywords]
    countVotes([c('u1', 'hoge', '2024-01-01T00:00:01Z')], keywords)
    expect(keywords).toEqual(snapshot)
  })
})

describe('countVotes - 出力形状', () => {
  it('counts のキー集合は keywords と一致（コメント由来のキーが混入しない）', () => {
    const comments = [c('u1', '違う', '2024-01-01T00:00:01Z')]
    const { counts } = countVotes(comments, ['hoge', 'fuga'])
    expect(Object.keys(counts).sort()).toEqual(['fuga', 'hoge'])
  })

  it('keywords 配列順序が counts キー順序に保たれる', () => {
    const comments = [c('u1', 'fuga', '2024-01-01T00:00:01Z')]
    const { counts } = countVotes(comments, ['zzz', 'aaa', 'fuga'])
    expect(Object.keys(counts)).toEqual(['zzz', 'aaa', 'fuga'])
  })

  it('keywords 重複時も全キーが含まれる', () => {
    const comments = [c('u1', 'hoge', '2024-01-01T00:00:01Z')]
    const { counts } = countVotes(comments, ['hoge', 'hoge'])
    expect(counts.hoge).toBe(1)
  })
})

describe('countVotes - 多コメンター大規模', () => {
  it('100 コメンターをそれぞれ別票として正しくカウント', () => {
    const comments = Array.from({ length: 100 }, (_, i) =>
      c(`u${i}`, 'hoge', `2024-01-01T00:${String(i).padStart(2, '0')}:00Z`),
    )
    expect(countVotes(comments, ['hoge']).counts).toEqual({ hoge: 100 })
  })

  it('複数ワード分散投票', () => {
    const comments: Comment[] = []
    for (let i = 0; i < 30; i++) {
      comments.push(c(`a${i}`, 'hoge', `2024-01-01T00:${String(i).padStart(2, '0')}:00Z`))
    }
    for (let i = 0; i < 20; i++) {
      comments.push(c(`b${i}`, 'fuga', `2024-01-01T01:${String(i).padStart(2, '0')}:00Z`))
    }
    for (let i = 0; i < 10; i++) {
      comments.push(c(`c${i}`, 'piyo', `2024-01-01T02:${String(i).padStart(2, '0')}:00Z`))
    }
    expect(countVotes(comments, ['hoge', 'fuga', 'piyo']).counts).toEqual({
      hoge: 30,
      fuga: 20,
      piyo: 10,
    })
  })
})

describe('countVotes - 国際化と特殊文字', () => {
  it('日本語キーワード', () => {
    const comments = [
      c('u1', '賛成', '2024-01-01T00:00:01Z'),
      c('u2', '反対', '2024-01-01T00:00:02Z'),
      c('u3', '賛成', '2024-01-01T00:00:03Z'),
    ]
    expect(countVotes(comments, ['賛成', '反対']).counts).toEqual({
      賛成: 2,
      反対: 1,
    })
  })

  it('絵文字キーワード', () => {
    const comments = [
      c('u1', '👍', '2024-01-01T00:00:01Z'),
      c('u2', '👎', '2024-01-01T00:00:02Z'),
      c('u3', '👍', '2024-01-01T00:00:03Z'),
    ]
    expect(countVotes(comments, ['👍', '👎']).counts).toEqual({ '👍': 2, '👎': 1 })
  })

  it('数字キーワード（投票番号）', () => {
    const comments = [
      c('u1', '1', '2024-01-01T00:00:01Z'),
      c('u2', '2', '2024-01-01T00:00:02Z'),
      c('u3', '1', '2024-01-01T00:00:03Z'),
    ]
    expect(countVotes(comments, ['1', '2', '3']).counts).toEqual({
      '1': 2,
      '2': 1,
      '3': 0,
    })
  })

  it('記号キーワード', () => {
    const comments = [c('u1', '○', '2024-01-01T00:00:01Z'), c('u2', '×', '2024-01-01T00:00:02Z')]
    expect(countVotes(comments, ['○', '×']).counts).toEqual({ '○': 1, '×': 1 })
  })

  it('長いキーワードでも完全一致', () => {
    const long = 'あ'.repeat(100)
    const comments = [c('u1', long, '2024-01-01T00:00:01Z')]
    expect(countVotes(comments, [long]).counts).toEqual({ [long]: 1 })
  })

  it('サロゲートペア絵文字（🇯🇵 = 国旗）も完全一致', () => {
    const comments = [c('u1', '🇯🇵', '2024-01-01T00:00:01Z')]
    expect(countVotes(comments, ['🇯🇵']).counts).toEqual({ '🇯🇵': 1 })
  })
})

describe('countVotes - 投票結果の独立性（regression）', () => {
  it('同じ comments を異なる keywords で集計しても干渉しない', () => {
    const comments = [
      c('u1', 'hoge', '2024-01-01T00:00:01Z'),
      c('u2', 'fuga', '2024-01-01T00:00:02Z'),
    ]
    const r1 = countVotes(comments, ['hoge'])
    const r2 = countVotes(comments, ['fuga'])
    expect(r1.counts).toEqual({ hoge: 1 })
    expect(r2.counts).toEqual({ fuga: 1 })
  })

  it('2回呼んでも同じ結果（純関数性）', () => {
    const comments = [
      c('u1', 'hoge', '2024-01-01T00:00:01Z'),
      c('u2', 'fuga', '2024-01-01T00:00:02Z'),
    ]
    const r1 = countVotes(comments, ['hoge', 'fuga'])
    const r2 = countVotes(comments, ['hoge', 'fuga'])
    expect(r1).toEqual(r2)
  })
})

describe('countVotes - partial モード', () => {
  it('キーワードを含むコメントをカウント', () => {
    const comments = [
      c('u1', 'hogeです', '2024-01-01T00:00:01Z'),
      c('u2', 'やっぱfuga', '2024-01-01T00:00:02Z'),
    ]
    expect(countVotes(comments, ['hoge', 'fuga'], 'partial').counts).toEqual({
      hoge: 1,
      fuga: 1,
    })
  })

  it('完全一致の場合も partial でカウント', () => {
    const comments = [c('u1', 'hoge', '2024-01-01T00:00:01Z')]
    expect(countVotes(comments, ['hoge'], 'partial').counts).toEqual({ hoge: 1 })
  })

  it('含まれない場合は 0', () => {
    const comments = [c('u1', 'piyo', '2024-01-01T00:00:01Z')]
    expect(countVotes(comments, ['hoge'], 'partial').counts).toEqual({ hoge: 0 })
  })

  it('1コメンター1票ルールは partial でも有効', () => {
    const comments = [
      c('u1', 'hoge投票', '2024-01-01T00:00:01Z'),
      c('u1', 'fuga投票', '2024-01-01T00:00:02Z'),
    ]
    expect(countVotes(comments, ['hoge', 'fuga'], 'partial').counts).toEqual({
      hoge: 1,
      fuga: 0,
    })
  })

  it('複数キーワードのうち最初にマッチしたものに投票', () => {
    const comments = [c('u1', 'hogehoge', '2024-01-01T00:00:01Z')]
    const { counts } = countVotes(comments, ['hoge', 'fuga'], 'partial')
    expect(counts['hoge']).toBe(1)
    expect(counts['fuga']).toBe(0)
  })

  it('matchMode 省略時は exact 動作', () => {
    const comments = [c('u1', 'hogeです', '2024-01-01T00:00:01Z')]
    expect(countVotes(comments, ['hoge']).counts).toEqual({ hoge: 0 })
  })

  it('包含関係のキーワードは最長一致を優先（配列順に依存しない）', () => {
    const comments = [c('u1', 'hogeだ', '2024-01-01T00:00:01Z')]
    expect(countVotes(comments, ['ho', 'hoge'], 'partial').counts).toEqual({
      ho: 0,
      hoge: 1,
    })
    expect(countVotes(comments, ['hoge', 'ho'], 'partial').counts).toEqual({
      hoge: 1,
      ho: 0,
    })
  })

  it('最長一致しても含まれなければ短いキーワードにフォールバック', () => {
    const comments = [c('u1', 'hoだ', '2024-01-01T00:00:01Z')]
    expect(countVotes(comments, ['ho', 'hoge'], 'partial').counts).toEqual({
      ho: 1,
      hoge: 0,
    })
  })
})
