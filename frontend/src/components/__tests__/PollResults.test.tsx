import { describe, it, expect } from 'vitest'
import { render, screen } from '@testing-library/react'
import { PollResults } from '../PollTab/PollResults'

describe('PollResults', () => {
  describe('非表示条件', () => {
    it('keywords 空なら何もレンダリングしない', () => {
      const { container } = render(
        <PollResults keywords={[]} counts={{}} totalVotes={0} isLoading={false} />,
      )
      expect(container.firstChild).toBeNull()
    })

    it('isLoading=true でも keywords 空なら非表示', () => {
      const { container } = render(
        <PollResults keywords={[]} counts={{}} totalVotes={0} isLoading={true} />,
      )
      expect(container.firstChild).toBeNull()
    })
  })

  describe('テーブル表示', () => {
    it('keywords 各行を表示', () => {
      render(
        <PollResults
          keywords={['hoge', 'fuga']}
          counts={{ hoge: 3, fuga: 1 }}
          totalVotes={4}
          isLoading={false}
        />,
      )
      expect(screen.getByText('hoge')).toBeInTheDocument()
      expect(screen.getByText('fuga')).toBeInTheDocument()
      expect(screen.getByText('3')).toBeInTheDocument()
      expect(screen.getByText('1')).toBeInTheDocument()
    })

    it('票数 0 でも 0 と表示', () => {
      render(
        <PollResults
          keywords={['a', 'b']}
          counts={{ a: 0, b: 0 }}
          totalVotes={0}
          isLoading={false}
        />,
      )
      const zeros = screen.getAllByText('0')
      // a=0, b=0, 合計=0 → 3つ
      expect(zeros.length).toBeGreaterThanOrEqual(3)
    })

    it('counts に欠落キーがあっても 0 として表示（堅牢性）', () => {
      render(
        <PollResults
          keywords={['hoge', 'missing']}
          counts={{ hoge: 5 }}
          totalVotes={5}
          isLoading={false}
        />,
      )
      expect(screen.getByText('hoge')).toBeInTheDocument()
      expect(screen.getByText('missing')).toBeInTheDocument()
      // missing 行に 0 が描画される
      const missingRow = screen.getByText('missing').closest('tr')!
      expect(missingRow.textContent).toContain('0')
      // hoge 行に 5 が描画される
      const hogeRow = screen.getByText('hoge').closest('tr')!
      expect(hogeRow.textContent).toContain('5')
    })

    it('keywords 順序通りに表示される', () => {
      render(
        <PollResults
          keywords={['zzz', 'aaa', 'mmm']}
          counts={{ zzz: 1, aaa: 2, mmm: 3 }}
          totalVotes={6}
          isLoading={false}
        />,
      )
      const rows = screen.getAllByRole('row')
      // header + 3 data rows + 1 footer = 5
      expect(rows.length).toBe(5)
      const dataCells = rows[1].textContent
      expect(dataCells).toContain('zzz')
    })

    it('ヘッダー: 「キーワード」「票数」', () => {
      render(
        <PollResults keywords={['hoge']} counts={{ hoge: 1 }} totalVotes={1} isLoading={false} />,
      )
      expect(screen.getByText('キーワード')).toBeInTheDocument()
      expect(screen.getByText('票数')).toBeInTheDocument()
    })
  })

  describe('合計行', () => {
    it('合計票数を表示', () => {
      render(
        <PollResults
          keywords={['a', 'b']}
          counts={{ a: 3, b: 5 }}
          totalVotes={8}
          isLoading={false}
        />,
      )
      expect(screen.getByText('合計')).toBeInTheDocument()
      expect(screen.getByText('8')).toBeInTheDocument()
    })

    it('合計 0 でも表示', () => {
      render(<PollResults keywords={['a']} counts={{ a: 0 }} totalVotes={0} isLoading={false} />)
      expect(screen.getByText('合計')).toBeInTheDocument()
    })
  })

  describe('ローディング', () => {
    it('isLoading=true でスピナー表示', () => {
      const { container } = render(
        <PollResults keywords={['hoge']} counts={{ hoge: 0 }} totalVotes={0} isLoading={true} />,
      )
      expect(container.querySelector('.animate-spin')).toBeInTheDocument()
    })

    it('isLoading=false でスピナー非表示', () => {
      const { container } = render(
        <PollResults keywords={['hoge']} counts={{ hoge: 0 }} totalVotes={0} isLoading={false} />,
      )
      expect(container.querySelector('.animate-spin')).toBeNull()
    })
  })

  describe('特殊キーワード', () => {
    it('日本語キーワードを表示', () => {
      render(
        <PollResults
          keywords={['賛成', '反対']}
          counts={{ 賛成: 10, 反対: 3 }}
          totalVotes={13}
          isLoading={false}
        />,
      )
      expect(screen.getByText('賛成')).toBeInTheDocument()
      expect(screen.getByText('反対')).toBeInTheDocument()
      expect(screen.getByText('10')).toBeInTheDocument()
    })

    it('絵文字キーワードを表示', () => {
      render(
        <PollResults
          keywords={['👍', '👎']}
          counts={{ '👍': 5, '👎': 2 }}
          totalVotes={7}
          isLoading={false}
        />,
      )
      expect(screen.getByText('👍')).toBeInTheDocument()
      expect(screen.getByText('👎')).toBeInTheDocument()
    })

    it('大量キーワード（20件）でも全件描画', () => {
      const keywords = Array.from({ length: 20 }, (_, i) => `k${i}`)
      const counts = Object.fromEntries(keywords.map((k) => [k, 1]))
      render(<PollResults keywords={keywords} counts={counts} totalVotes={20} isLoading={false} />)
      const rows = screen.getAllByRole('row')
      // header + 20 data + 1 footer
      expect(rows.length).toBe(22)
    })
  })
})
