import { describe, it, expect, vi, beforeEach } from 'vitest'
import { render, screen, act } from '@testing-library/react'
import userEvent from '@testing-library/user-event'
import { PollResults } from '../PollTab/PollResults'

const baseProps = {
  voters: {},
  totalVotes: 0,
  isLoading: false,
}

describe('PollResults', () => {
  describe('非表示条件', () => {
    it('keywords 空なら何もレンダリングしない', () => {
      const { container } = render(<PollResults keywords={[]} counts={{}} {...baseProps} />)
      expect(container.firstChild).toBeNull()
    })

    it('isLoading=true でも keywords 空なら非表示', () => {
      const { container } = render(
        <PollResults keywords={[]} counts={{}} {...baseProps} isLoading={true} />,
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
          voters={{ hoge: [], fuga: [] }}
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
          voters={{ a: [], b: [] }}
          totalVotes={0}
          isLoading={false}
        />,
      )
      const zeros = screen.getAllByText('0')
      expect(zeros.length).toBeGreaterThanOrEqual(3)
    })

    it('counts に欠落キーがあっても 0 として表示（堅牢性）', () => {
      render(
        <PollResults
          keywords={['hoge', 'missing']}
          counts={{ hoge: 5 }}
          voters={{ hoge: [], missing: [] }}
          totalVotes={5}
          isLoading={false}
        />,
      )
      expect(screen.getByText('hoge')).toBeInTheDocument()
      expect(screen.getByText('missing')).toBeInTheDocument()
      const missingRow = screen.getByText('missing').closest('tr')!
      expect(missingRow.textContent).toContain('0')
      const hogeRow = screen.getByText('hoge').closest('tr')!
      expect(hogeRow.textContent).toContain('5')
    })

    it('ヘッダー: 「キーワード」「票数」', () => {
      render(
        <PollResults
          keywords={['hoge']}
          counts={{ hoge: 1 }}
          voters={{ hoge: [] }}
          totalVotes={1}
          isLoading={false}
        />,
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
          voters={{ a: [], b: [] }}
          totalVotes={8}
          isLoading={false}
        />,
      )
      expect(screen.getByText('合計')).toBeInTheDocument()
      expect(screen.getByText('8')).toBeInTheDocument()
    })
  })

  describe('ローディング', () => {
    it('isLoading=true でスピナー表示', () => {
      const { container } = render(
        <PollResults
          keywords={['hoge']}
          counts={{ hoge: 0 }}
          voters={{ hoge: [] }}
          totalVotes={0}
          isLoading={true}
        />,
      )
      expect(container.querySelector('[data-testid="poll-loading-spinner"]')).toBeInTheDocument()
    })

    it('isLoading=false でスピナー非表示', () => {
      const { container } = render(
        <PollResults
          keywords={['hoge']}
          counts={{ hoge: 0 }}
          voters={{ hoge: [] }}
          totalVotes={0}
          isLoading={false}
        />,
      )
      expect(container.querySelector('[data-testid="poll-loading-spinner"]')).toBeNull()
    })
  })

  describe('投票ユーザー expand', () => {
    const voters = {
      hoge: [
        { channelId: 'UC1', displayName: 'taro' },
        { channelId: 'UC2', displayName: 'hanako' },
      ],
      fuga: [],
    }

    it('初期状態では voter list は非表示', () => {
      render(
        <PollResults
          keywords={['hoge']}
          counts={{ hoge: 2 }}
          voters={voters}
          totalVotes={2}
          isLoading={false}
        />,
      )
      expect(screen.queryByText('taro')).toBeNull()
    })

    it('キーワード行クリックで voter list を表示', async () => {
      const user = userEvent.setup()
      render(
        <PollResults
          keywords={['hoge']}
          counts={{ hoge: 2 }}
          voters={voters}
          totalVotes={2}
          isLoading={false}
        />,
      )
      await act(async () => {
        await user.click(screen.getByText('hoge'))
      })
      expect(screen.getByText('taro')).toBeInTheDocument()
      expect(screen.getByText('hanako')).toBeInTheDocument()
      expect(screen.queryByText('UC1')).toBeNull()
      expect(screen.queryByText('UC2')).toBeNull()
    })

    it('handle がある場合のみ表示する', async () => {
      const user = userEvent.setup()
      render(
        <PollResults
          keywords={['hoge']}
          counts={{ hoge: 1 }}
          voters={{
            hoge: [{ channelId: 'UC1', displayName: 'taro', handle: '@tarochannel' }],
          }}
          totalVotes={1}
          isLoading={false}
        />,
      )
      await act(async () => {
        await user.click(screen.getByText('hoge'))
      })
      expect(screen.getByText('@tarochannel')).toBeInTheDocument()
      expect(screen.queryByText('UC1')).toBeNull()
    })

    it('再クリックで折りたたまれる', async () => {
      const user = userEvent.setup()
      render(
        <PollResults
          keywords={['hoge']}
          counts={{ hoge: 2 }}
          voters={voters}
          totalVotes={2}
          isLoading={false}
        />,
      )
      await act(async () => {
        await user.click(screen.getByText('hoge'))
      })
      await act(async () => {
        await user.click(screen.getByText('hoge'))
      })
      expect(screen.queryByText('taro')).toBeNull()
    })

    it('voter 0 件のとき空メッセージ表示', async () => {
      const user = userEvent.setup()
      render(
        <PollResults
          keywords={['fuga']}
          counts={{ fuga: 0 }}
          voters={voters}
          totalVotes={0}
          isLoading={false}
        />,
      )
      await act(async () => {
        await user.click(screen.getByText('fuga'))
      })
      expect(screen.getByText('投票したユーザーはいません')).toBeInTheDocument()
    })
  })

  describe('コピー', () => {
    let writeTextSpy: ReturnType<typeof vi.spyOn>

    beforeEach(() => {
      writeTextSpy = vi.spyOn(navigator.clipboard, 'writeText').mockResolvedValue(undefined)
    })

    it('handle なしの場合は keyword + displayName + 末尾空タブでコピー', async () => {
      const user = userEvent.setup()
      render(
        <PollResults
          keywords={['hoge']}
          counts={{ hoge: 1 }}
          voters={{
            hoge: [{ channelId: 'UC1', displayName: 'taro' }],
          }}
          totalVotes={1}
          isLoading={false}
        />,
      )
      await act(async () => {
        await user.click(screen.getByText('hoge'))
      })
      await act(async () => {
        await user.click(screen.getByRole('button', { name: 'クリップボードにコピー' }))
      })
      expect(writeTextSpy).toHaveBeenCalledWith('\t\thoge\ttaro\t')
      expect(screen.getByRole('button', { name: 'コピー済' })).toBeInTheDocument()
    })

    it('handle ありの場合は keyword + displayName + handle を TSV でコピー', async () => {
      const user = userEvent.setup()
      render(
        <PollResults
          keywords={['hoge']}
          counts={{ hoge: 1 }}
          voters={{
            hoge: [{ channelId: 'UC1', displayName: 'taro', handle: '@tarochannel' }],
          }}
          totalVotes={1}
          isLoading={false}
        />,
      )
      await act(async () => {
        await user.click(screen.getByText('hoge'))
      })
      await act(async () => {
        await user.click(screen.getByRole('button', { name: 'クリップボードにコピー' }))
      })
      expect(writeTextSpy).toHaveBeenCalledWith('\t\thoge\ttaro\t@tarochannel')
    })

    it('handle あり/なし混在でも列位置が揃う', async () => {
      const user = userEvent.setup()
      render(
        <PollResults
          keywords={['hoge']}
          counts={{ hoge: 2 }}
          voters={{
            hoge: [
              { channelId: 'UC1', displayName: 'taro', handle: '@tarochannel' },
              { channelId: 'UC2', displayName: 'hanako' },
            ],
          }}
          totalVotes={2}
          isLoading={false}
        />,
      )
      await act(async () => {
        await user.click(screen.getByText('hoge'))
      })
      await act(async () => {
        await user.click(screen.getByRole('button', { name: 'クリップボードにコピー' }))
      })
      expect(writeTextSpy).toHaveBeenCalledWith('\t\thoge\ttaro\t@tarochannel\n\t\thoge\thanako\t')
    })

    it('videoId / savedAt があれば先頭2列に含める', async () => {
      const user = userEvent.setup()
      render(
        <PollResults
          keywords={['hoge']}
          counts={{ hoge: 1 }}
          voters={{
            hoge: [{ channelId: 'UC1', displayName: 'taro', handle: '@tarochannel' }],
          }}
          totalVotes={1}
          isLoading={false}
          videoId="VIDEO123"
          savedAt="2026-06-20T10:00:00Z"
        />,
      )
      await act(async () => {
        await user.click(screen.getByText('hoge'))
      })
      await act(async () => {
        await user.click(screen.getByRole('button', { name: 'クリップボードにコピー' }))
      })
      expect(writeTextSpy).toHaveBeenCalledWith(
        '2026-06-20T10:00:00Z\tVIDEO123\thoge\ttaro\t@tarochannel',
      )
    })
  })
})
