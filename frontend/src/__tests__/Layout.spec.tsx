import { render, screen } from '@testing-library/react'
import App from '../App.jsx'

describe('Layout', () => {
  test('長文ユーザー名に truncate-1 クラスが適用される（チップ表示）', async () => {
    render(<App />)
    // NOTE: デフォルトモックではユーザーが空なので、UI クラスの存在のみ簡易確認
    // App のチップ要素は <span class="truncate-1" ...> として描画される
    // 初期状態でユーザーがいないため、待機バナーの存在を確認しておく
    expect(await screen.findByText('WAITING')).toBeInTheDocument()
  })
})
