import { render, screen, fireEvent } from '@testing-library/react'
import App from '../App.jsx'

describe('操作ガイド表示機能', () => {
  test('操作ガイドボタンが表示される', async () => {
    render(<App />)

    expect(screen.getByText('はじめての方へ - 操作ガイド')).toBeInTheDocument()
  })

  test('ガイドを展開すると基本操作説明が表示される', async () => {
    render(<App />)

    // ガイドボタンをクリックして展開
    const guideButton = screen.getByText('はじめての方へ - 操作ガイド')
    fireEvent.click(guideButton)

    expect(screen.getByText(/YouTube動画のURLまたはvideoIdを下の入力欄に貼り付けて/)).toBeInTheDocument()
  })

  test('ガイド展開時に各ボタンの説明が表示される', async () => {
    render(<App />)

    // ガイドボタンをクリックして展開
    const guideButton = screen.getByText('はじめての方へ - 操作ガイド')
    fireEvent.click(guideButton)

    expect(screen.getByText('切替:')).toBeInTheDocument()
    expect(screen.getByText('指定した動画の監視を開始')).toBeInTheDocument()
    expect(screen.getByText('今すぐ取得:')).toBeInTheDocument()
    expect(screen.getByText('手動でコメントを取得')).toBeInTheDocument()
    expect(screen.getByText('リセット:')).toBeInTheDocument()
    expect(screen.getByText('参加者リストをクリア')).toBeInTheDocument()
  })

})