import { render, screen } from '@testing-library/react'
import { Header } from '../Header'

describe('Header コンポーネント', () => {
  test('タイトルが正しく表示される', () => {
    render(<Header active={false} userCount={0} />)

    expect(screen.getByText('YouTube Live —')).toBeInTheDocument()
    expect(screen.getByText('参加ユーザー')).toBeInTheDocument()
    expect(screen.getByText('配信中に参加したユーザーを収集し、終了時点で全員が表示されることを目指します。')).toBeInTheDocument()
  })

  test('ACTIVE状態が正しく表示される', () => {
    render(<Header active={true} userCount={5} />)

    expect(screen.getAllByText('ACTIVE')).toHaveLength(2) // ステータスバッジと状態表示
    // 最初のACTIVEが含まれるスパンの親要素（ボーダーバッジ）をチェック
    const statusBadge = screen.getAllByText('ACTIVE')[0].closest('.inline-flex')
    expect(statusBadge).toHaveClass('border-emerald-500/30', 'bg-emerald-500/10', 'text-emerald-700')
  })

  test('WAITING状態が正しく表示される', () => {
    render(<Header active={false} userCount={0} />)

    expect(screen.getAllByText('WAITING')).toHaveLength(2) // ステータスバッジと状態表示
    const statusBadge = screen.getAllByText('WAITING')[0].closest('.inline-flex')
    expect(statusBadge).toHaveClass('border-amber-500/30', 'bg-amber-400/10', 'text-amber-800')
  })

  test('参加者数が正しく表示される', () => {
    render(<Header active={true} userCount={42} />)

    expect(screen.getByTestId('counter')).toHaveTextContent('42')
    expect(screen.getByText('参加者')).toBeInTheDocument()
  })



  test('0人の場合も正しく表示される', () => {
    render(<Header active={false} userCount={0} />)

    expect(screen.getByTestId('counter')).toHaveTextContent('0')
  })
})