import { render, screen } from '@testing-library/react'
import App from '../App.jsx'

describe('Layout', () => {
  test('テーブル行に truncate-1 が含まれる', async () => {
    render(<App />)
    // 見出しがあること
    expect(await screen.findByText(/参加ユーザー/)).toBeInTheDocument()
    // 名前セルに truncate-1 クラスが適用されていること
    const nameCell = document.querySelector('td.truncate-1')
    expect(nameCell).not.toBeNull()
  })
})
