import { render, screen } from '@testing-library/react'
import App from '../App.jsx'

describe('操作方法表示機能', () => {
  test('操作方法セクションが表示される', async () => {
    render(<App />)

    expect(screen.getByText('操作方法')).toBeInTheDocument()
  })

  test('videoId入力の説明が表示される', async () => {
    render(<App />)

    expect(screen.getByText(/YouTube動画のURLまたはvideoIdを入力/)).toBeInTheDocument()
  })

  test('各ボタンの説明が表示される', async () => {
    render(<App />)

    expect(screen.getByText(/切替: 指定した動画の監視を開始/)).toBeInTheDocument()
    expect(screen.getByText(/今すぐ取得: 手動でコメントを取得/)).toBeInTheDocument()
    expect(screen.getByText(/リセット: 現在の参加者リストをクリア/)).toBeInTheDocument()
  })

  test('自動更新の説明が表示される', async () => {
    render(<App />)

    expect(screen.getByText(/自動間隔: 設定した間隔でデータを更新/)).toBeInTheDocument()
  })
})