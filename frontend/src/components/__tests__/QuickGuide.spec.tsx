import { render, screen, fireEvent } from '@testing-library/react'
import { QuickGuide } from '../QuickGuide'

describe('QuickGuide コンポーネント', () => {
  test('初期状態では操作ガイドが折りたたまれている', () => {
    render(<QuickGuide />)

    const button = screen.getByRole('button', { name: /はじめての方へ - 操作ガイド/ })
    expect(button).toBeInTheDocument()
    expect(button).toHaveAttribute('aria-expanded', 'false')

    // 展開されたコンテンツは表示されていない
    expect(screen.queryByText('基本の使い方')).not.toBeInTheDocument()
  })

  test('ボタンクリックで操作ガイドが展開される', () => {
    render(<QuickGuide />)

    const button = screen.getByRole('button', { name: /はじめての方へ - 操作ガイド/ })
    fireEvent.click(button)

    expect(button).toHaveAttribute('aria-expanded', 'true')

    // 展開されたコンテンツが表示される
    expect(screen.getByText('基本の使い方')).toBeInTheDocument()
    expect(screen.getByText('YouTube動画のURLまたはvideoIdを下の入力欄に貼り付けて「切替」ボタンをクリックしてください。')).toBeInTheDocument()
  })

  test('展開状態でボタンクリックすると折りたたまれる', () => {
    render(<QuickGuide />)

    const button = screen.getByRole('button', { name: /はじめての方へ - 操作ガイド/ })

    // 展開
    fireEvent.click(button)
    expect(screen.getByText('基本の使い方')).toBeInTheDocument()

    // 折りたたみ
    fireEvent.click(button)
    expect(button).toHaveAttribute('aria-expanded', 'false')
    expect(screen.queryByText('基本の使い方')).not.toBeInTheDocument()
  })

  test('展開時に操作説明が正しく表示される', () => {
    render(<QuickGuide />)

    const button = screen.getByRole('button', { name: /はじめての方へ - 操作ガイド/ })
    fireEvent.click(button)

    // 各ボタンの説明
    expect(screen.getByText('切替:')).toBeInTheDocument()
    expect(screen.getByText('指定した動画の監視を開始')).toBeInTheDocument()
    expect(screen.getByText('今すぐ取得:')).toBeInTheDocument()
    expect(screen.getByText('手動でコメントを取得')).toBeInTheDocument()
    expect(screen.getByText('リセット:')).toBeInTheDocument()
    expect(screen.getByText('参加者リストをクリア')).toBeInTheDocument()

    // コツ
    expect(screen.getByText('💡 コツ:')).toBeInTheDocument()
    expect(screen.getByText('配信開始前にアプリを起動すると、より多くの参加者を取得できます')).toBeInTheDocument()

    // 配信終了後の説明
    expect(screen.getByText('🔄 配信終了後:')).toBeInTheDocument()
    expect(screen.getByText('自動検知')).toBeInTheDocument()
    expect(screen.getByText('自動的にクリア')).toBeInTheDocument()
    expect(screen.getByText('• 新しい配信を始める場合は、新しいvideoIdを入力して「切替」してください')).toBeInTheDocument()
    expect(screen.getByText('• 手動でリセットしたい場合は「リセット」ボタンをお使いください')).toBeInTheDocument()
  })

  test('アイコンの回転状態が正しく変化する', () => {
    render(<QuickGuide />)

    const button = screen.getByRole('button', { name: /はじめての方へ - 操作ガイド/ })
    const icon = button.querySelector('svg')

    // 初期状態（折りたたみ）
    expect(icon).not.toHaveClass('rotate-90')

    // 展開
    fireEvent.click(button)
    expect(icon).toHaveClass('rotate-90')

    // 再度折りたたみ
    fireEvent.click(button)
    expect(icon).not.toHaveClass('rotate-90')
  })

  test('aria-controls属性が正しく設定されている', () => {
    render(<QuickGuide />)

    const button = screen.getByRole('button', { name: /はじめての方へ - 操作ガイド/ })
    expect(button).toHaveAttribute('aria-controls', 'operation-guide')

    // 展開時にid付きの要素が存在する
    fireEvent.click(button)
    expect(document.getElementById('operation-guide')).toBeInTheDocument()
  })
})