import { render, screen } from '@testing-library/react'
import App from '../App.jsx'

// Note: matchMedia and localStorage are mocked globally in setup.ts

describe('App UI', () => {
  test('ヘッダー/カウンタ/テーブルが描画される', async () => {
    render(<App />)
    // 実際に存在するテキストを確認
    expect(await screen.findByText('総ユーザー数')).toBeInTheDocument()
    expect(screen.getByText('名前')).toBeInTheDocument()
    // ユーザーがいない場合のメッセージも確認
    expect(screen.getByText('ユーザーがいません。')).toBeInTheDocument()
  })
})
